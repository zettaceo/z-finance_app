package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var ErrCryptoToPixInsufficient = errors.New("saldo cripto insuficiente")

type SendPixFromCryptoInput struct {
	UserID         string
	AccountID      string
	AmountBRL      int64
	Asset          string
	IdempotencyKey string
	ExternalRef    string
}

type SendPixFromCryptoUseCase struct {
	uow      ports.UnitOfWork
	exchange ports.ExchangeGateway
	clock    Clock
	pixRepo  repository.PixTransferRepository
	pricing  *ResolvePricingUseCase
}

func NewSendPixFromCryptoUseCase(uow ports.UnitOfWork, exchange ports.ExchangeGateway, pixRepo repository.PixTransferRepository, pricing *ResolvePricingUseCase) *SendPixFromCryptoUseCase {
	return &SendPixFromCryptoUseCase{uow: uow, exchange: exchange, clock: NewRealClock(), pixRepo: pixRepo, pricing: pricing}
}

func (uc *SendPixFromCryptoUseCase) Execute(ctx context.Context, input SendPixFromCryptoInput) (*entity.PixTransfer, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "SendPixFromCrypto")
	defer end()

	if input.AmountBRL <= 0 || input.UserID == "" || input.AccountID == "" || input.IdempotencyKey == "" {
		return nil, ErrCryptoInvalidAmount
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	existingTx, err := uowTx.TransactionRepository().GetByIdempotencyKey(ctx, input.AccountID, input.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existingTx != nil {
		existingTransfer, err := uowTx.PixTransferRepository().GetByIdempotencyKey(ctx, input.AccountID, input.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existingTransfer != nil {
			return existingTransfer, nil
		}
	}

	feeResult, rule, err := uc.resolveFee(ctx, input)
	if err != nil {
		return nil, err
	}
	externalRef := normalizeExternalRef(input.ExternalRef, entity.TransactionTypeWithdrawal, "PIX_OUT")
	quote, err := uc.exchange.Quote(ctx, input.Asset)
	if err != nil {
		return nil, NewExternalDependencyError("exchange", "quote", err)
	}
	quoteAt := uc.clock.Now().UTC()
	requiredCrypto, ok := convertFiatToCrypto(feeResult.NetAmount, quote.PriceInBRLCents, 2)
	if !ok || requiredCrypto <= 0 {
		return nil, ErrCryptoInvalidAmount
	}

	confirmed, err := uowTx.CryptoTransferRepository().SumConfirmedByUserAsset(ctx, input.UserID, input.Asset)
	if err != nil {
		return nil, err
	}
	if confirmed < requiredCrypto {
		return nil, ErrCryptoToPixInsufficient
	}

	now := uc.clock.Now().UTC()
	cryptoHold := &entity.CryptoTransfer{
		ID:            uuid.NewString(),
		UserID:        input.UserID,
		AccountID:     input.AccountID,
		TransactionID: uuid.NewString(),
		Asset:         input.Asset,
		Network:       "INTERNAL",
		Address:       "PIX_FROM_CRYPTO",
		Amount:        requiredCrypto,
		Fee:           0,
		Status:        entity.CryptoTransferPendingExchange,
		Direction:     "SELL",
		CreatedAt:     now,
	}
	if err := uowTx.CryptoTransferRepository().Create(ctx, cryptoHold); err != nil {
		return nil, err
	}

	txHash, err := uc.exchange.Execute(ctx, input.Asset, requiredCrypto, "SELL")
	if err != nil {
		return nil, NewExternalDependencyError("exchange", "execute", err)
	}
	if err := uowTx.CryptoTransferRepository().UpdateStatus(ctx, cryptoHold.ID, entity.CryptoTransferConfirmed, txHash); err != nil {
		return nil, err
	}

	tradeSell := &entity.Transaction{
		ID:             cryptoHold.TransactionID,
		AccountID:      input.AccountID,
		UserID:         input.UserID,
		Type:           entity.TransactionTypeTradeSell,
		Status:         entity.TransactionStatusConfirmed,
		Amount:         input.AmountBRL,
		Fee:            0,
		NetAmount:      input.AmountBRL,
		IdempotencyKey: input.IdempotencyKey + ":sell",
		ExternalRef:    normalizeExternalRef("", entity.TransactionTypeTradeSell, "CRYPTO_TO_PIX"),
		OccurredAt:     now,
		CreatedAt:      now,
	}
	if err := uowTx.TransactionRepository().Create(ctx, tradeSell); err != nil {
		return nil, err
	}
	if err := appendConversionAudit(ctx, uowTx, uc.clock, conversionAuditInput{
		UserID:          input.UserID,
		OperationType:   entity.PricingOperationCryptoToPix,
		Asset:           input.Asset,
		GrossAmount:     input.AmountBRL,
		Fee:             0,
		NetAmount:       input.AmountBRL,
		QuotePrice:      quote.PriceInBRLCents,
		SpreadBps:       0,
		LiquiditySource: "EXCHANGE",
		RelatedType:     "transaction",
		RelatedID:       tradeSell.ID,
		QuotedAt:        &quoteAt,
	}); err != nil {
		return nil, err
	}

	pixTx := &entity.Transaction{
		ID:             uuid.NewString(),
		AccountID:      input.AccountID,
		UserID:         input.UserID,
		Type:           entity.TransactionTypeWithdrawal,
		Status:         entity.TransactionStatusHold,
		Amount:         input.AmountBRL,
		Fee:            feeResult.Fee,
		NetAmount:      feeResult.NetAmount,
		IdempotencyKey: input.IdempotencyKey,
		ExternalRef:    externalRef,
		OccurredAt:     now,
		CreatedAt:      now,
	}
	if err := uowTx.TransactionRepository().Create(ctx, pixTx); err != nil {
		return nil, err
	}

	pixTransfer := &entity.PixTransfer{
		ID:             uuid.NewString(),
		TransactionID:  pixTx.ID,
		UserID:         input.UserID,
		AccountID:      input.AccountID,
		Direction:      entity.PixDirectionOut,
		Status:         entity.PixStatusCreated,
		Amount:         input.AmountBRL,
		Fee:            feeResult.Fee,
		NetAmount:      feeResult.NetAmount,
		IdempotencyKey: input.IdempotencyKey,
		ExternalRef:    externalRef,
		OccurredAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := uowTx.PixTransferRepository().Create(ctx, pixTransfer); err != nil {
		return nil, err
	}
	_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "CRYPTO_TO_PIX_CREATED", "pix_transfer", pixTransfer.ID, map[string]any{
		"crypto_transfer_id": cryptoHold.ID,
	})
	if feeResult.Fee > 0 {
		_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "PRICING_APPLIED", "transaction", pixTx.ID, map[string]any{
			"operation":  string(entity.PricingOperationCryptoToPix),
			"amount":     feeResult.Amount,
			"fee":        feeResult.Fee,
			"net_amount": feeResult.NetAmount,
			"rule_id":    ruleID(rule),
		})
	}

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}

	if uc.pixRepo != nil {
		if err := uc.pixRepo.UpdateStatus(ctx, pixTransfer.ID, entity.PixStatusPendingPartner); err != nil {
			return pixTransfer, err
		}
		pixTransfer.Status = entity.PixStatusPendingPartner
	}
	return pixTransfer, nil
}

func (uc *SendPixFromCryptoUseCase) resolveFee(ctx context.Context, input SendPixFromCryptoInput) (entity.FeeResult, *entity.PricingRule, error) {
	if uc.pricing == nil {
		return entity.FeeResult{Amount: input.AmountBRL, Fee: 0, NetAmount: input.AmountBRL}, nil, nil
	}
	return uc.pricing.Execute(ctx, PricingInput{
		UserID:        input.UserID,
		OperationType: entity.PricingOperationCryptoToPix,
		Asset:         "BRL",
		GrossAmount:   input.AmountBRL,
		FeatureCode:   string(entity.PricingOperationCryptoToPix),
	})
}
