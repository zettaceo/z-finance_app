package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var (
	ErrInvoiceInvalidAmount = errors.New("valor de invoice invalido")
	ErrInvoiceInvalidInput  = errors.New("dados de invoice invalidos")
	ErrInvoiceInvalidMethod = errors.New("metodo de pagamento invalido")
	ErrInvoiceMissingIdempotency = errors.New("idempotency key obrigatoria")
)

type CreateInvoiceInput struct {
	UserID    string
	AmountBRL int64
	IdempotencyKey string
}

type CreateInvoiceResult struct {
	Invoice   *entity.Invoice
	QRPayload string
}

type CreateInvoiceUseCase struct {
	uow   ports.UnitOfWork
	custody ports.CustodyGateway
	clock Clock
}

func NewCreateInvoiceUseCase(uow ports.UnitOfWork, custody ports.CustodyGateway) *CreateInvoiceUseCase {
	return &CreateInvoiceUseCase{uow: uow, custody: custody, clock: NewRealClock()}
}

func (uc *CreateInvoiceUseCase) Execute(ctx context.Context, input CreateInvoiceInput) (*CreateInvoiceResult, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "CreateInvoice")
	defer end()

	if input.UserID == "" {
		return nil, ErrInvoiceInvalidInput
	}
	if input.AmountBRL <= 0 {
		return nil, ErrInvoiceInvalidAmount
	}
	if input.IdempotencyKey == "" {
		return nil, ErrInvoiceMissingIdempotency
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	existing, err := uowTx.InvoiceRepository().GetByIdempotencyKey(ctx, input.UserID, input.IdempotencyKey)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return &CreateInvoiceResult{
			Invoice:   existing,
			QRPayload: buildHybridQR(existing),
		}, nil
	}

	address, err := uc.generateUSDTAddress(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	now := uc.clock.Now().UTC()
	invoiceID := uuid.NewString()
	invoice := &entity.Invoice{
		ID:           invoiceID,
		UserID:       input.UserID,
		AmountBRL:    input.AmountBRL,
		PixCopyPaste: buildPixCopyPaste(invoiceID, input.UserID, input.AmountBRL),
		USDTAddress:  address,
		IdempotencyKey: input.IdempotencyKey,
		CreatedAt:    now,
	}

	if err := uowTx.InvoiceRepository().Create(ctx, invoice); err != nil {
		return nil, err
	}

	_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "INVOICE_CREATE", "invoice", invoice.ID, map[string]any{
		"amount_brl": invoice.AmountBRL,
	})

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}

	return &CreateInvoiceResult{
		Invoice:   invoice,
		QRPayload: buildHybridQR(invoice),
	}, nil
}

type PayInvoiceInput struct {
	InvoiceID      string
	PayerUserID    string
	PayerAccountID string
	Method         string
	IdempotencyKey string
}

type PayInvoiceResult struct {
	Invoice        *entity.Invoice
	Method         string
	PixTransfer    *entity.PixTransfer
	CryptoTransfer *entity.CryptoTransfer
}

type PayInvoiceUseCase struct {
	invoiceRepo repository.InvoiceRepository
	sendPixUC   *SendPixUseCase
	payCryptoUC *PayCryptoWithFiatUseCase
	pricing     *ResolvePricingUseCase
}

func NewPayInvoiceUseCase(invoiceRepo repository.InvoiceRepository, sendPixUC *SendPixUseCase, payCryptoUC *PayCryptoWithFiatUseCase, pricing *ResolvePricingUseCase) *PayInvoiceUseCase {
	return &PayInvoiceUseCase{
		invoiceRepo: invoiceRepo,
		sendPixUC:   sendPixUC,
		payCryptoUC: payCryptoUC,
		pricing:     pricing,
	}
}

func (uc *PayInvoiceUseCase) Execute(ctx context.Context, input PayInvoiceInput) (*PayInvoiceResult, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "PayInvoice")
	defer end()

	if input.InvoiceID == "" || input.PayerUserID == "" || input.PayerAccountID == "" || input.IdempotencyKey == "" {
		return nil, ErrInvoiceInvalidInput
	}

	method := strings.ToUpper(strings.TrimSpace(input.Method))
	if method == "" {
		method = "PIX"
	}
	if method != "PIX" && method != "USDT" {
		return nil, ErrInvoiceInvalidMethod
	}

	invoice, err := uc.invoiceRepo.GetByID(ctx, input.InvoiceID)
	if err != nil {
		return nil, err
	}

	result := &PayInvoiceResult{
		Invoice: invoice,
		Method:  method,
	}

	feeResult := entity.FeeResult{Amount: invoice.AmountBRL, Fee: 0, NetAmount: invoice.AmountBRL}
	if uc.pricing != nil {
		resolved, _, err := uc.pricing.Execute(ctx, PricingInput{
			UserID:        input.PayerUserID,
			OperationType: entity.PricingOperationInvoice,
			Asset:         "BRL",
			GrossAmount:   invoice.AmountBRL,
			FeatureCode:   string(entity.PricingOperationInvoice),
		})
		if err != nil {
			return nil, err
		}
		feeResult = resolved
	}

	switch method {
	case "PIX":
		if uc.sendPixUC == nil {
			return nil, ErrInvoiceInvalidInput
		}
		pixTransfer, err := uc.sendPixUC.Execute(ctx, SendPixInput{
			AccountID:      input.PayerAccountID,
			UserID:         input.PayerUserID,
			Amount:         invoice.AmountBRL,
			Fee:            feeResult.Fee,
			NetAmount:      feeResult.NetAmount,
			SkipPricing:    true,
			IdempotencyKey: input.IdempotencyKey,
			ExternalRef:    "INVOICE:" + invoice.ID,
			Metadata: map[string]any{
				"invoice_id":      invoice.ID,
				"payee_user_id":   invoice.UserID,
				"pix_copy_paste":  invoice.PixCopyPaste,
				"usdt_address":    invoice.USDTAddress,
				"payment_method":  "PIX",
			},
		})
		if err != nil {
			return nil, err
		}
		result.PixTransfer = pixTransfer
	case "USDT":
		if uc.payCryptoUC == nil {
			return nil, ErrInvoiceInvalidInput
		}
		cryptoTransfer, err := uc.payCryptoUC.Execute(ctx, PayCryptoWithFiatInput{
			UserID:         input.PayerUserID,
			FiatAmount:     invoice.AmountBRL,
			Destination:    invoice.USDTAddress,
			IdempotencyKey: input.IdempotencyKey,
			FeeOverride:    feeResult.Fee,
			SkipPricing:    true,
		})
		if err != nil {
			return nil, err
		}
		result.CryptoTransfer = cryptoTransfer
	}

	return result, nil
}

func buildPixCopyPaste(invoiceID, userID string, amountBRL int64) string {
	return fmt.Sprintf("ZETTA|PIX|%s|%d|%s", invoiceID, amountBRL, userID)
}

func buildHybridQR(invoice *entity.Invoice) string {
	return fmt.Sprintf("PIX:%s|USDT:%s", invoice.PixCopyPaste, invoice.USDTAddress)
}

func generateUSDTAddress() (string, error) {
	raw, err := generateEVMAddress()
	if err != nil {
		return "", err
	}
	return "bsc:" + raw, nil
}

func (uc *CreateInvoiceUseCase) generateUSDTAddress(ctx context.Context, userID string) (string, error) {
	if uc.custody == nil {
		return generateUSDTAddress()
	}
	address, err := uc.custody.CreateDepositAddress(ctx, userID, "USDT", "BSC")
	if err != nil {
		return "", err
	}
	return address.Address, nil
}

func generateEVMAddress() (string, error) {
	buffer := make([]byte, 20)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(buffer), nil
}
