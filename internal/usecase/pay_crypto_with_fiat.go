package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/crypto/address"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var (
	ErrCryptoInvalidAmount     = errors.New("valor fiat invalido")
	ErrCryptoIdempotencyNeeded = errors.New("idempotency key obrigatoria")
	ErrCryptoAccountNotFound   = errors.New("conta ativa nao encontrada")
)

type PayCryptoWithFiatInput struct {
	UserID         string
	FiatAmount     int64
	Destination    string
	IdempotencyKey string
	FeeOverride    int64
	SkipPricing    bool
}

type PayCryptoWithFiatUseCase struct {
	uow   ports.UnitOfWork
	clock Clock
	exchange ports.ExchangeGateway
	custody ports.CustodyGateway
	pricing *ResolvePricingUseCase
}

func NewPayCryptoWithFiatUseCase(uow ports.UnitOfWork, exchange ports.ExchangeGateway, custody ports.CustodyGateway, pricing *ResolvePricingUseCase) *PayCryptoWithFiatUseCase {
	return &PayCryptoWithFiatUseCase{
		uow:      uow,
		clock:    NewRealClock(),
		exchange: exchange,
		custody:  custody,
		pricing:  pricing,
	}
}

func (uc *PayCryptoWithFiatUseCase) Execute(ctx context.Context, input PayCryptoWithFiatInput) (*entity.CryptoTransfer, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "PayCryptoWithFiat")
	defer end()

	if input.FiatAmount <= 0 {
		return nil, ErrCryptoInvalidAmount
	}
	if input.IdempotencyKey == "" {
		return nil, ErrCryptoIdempotencyNeeded
	}

	network, asset, resolved, err := address.ResolveAddress(input.Destination)
	if err != nil {
		return nil, err
	}

	feeResult, rule, err := uc.resolveFee(ctx, input)
	if err != nil {
		return nil, err
	}

	quote, err := uc.exchange.Quote(ctx, string(asset))
	if err != nil {
		return nil, NewExternalDependencyError("exchange", "quote", err)
	}
	quoteAt := uc.clock.Now().UTC()
	cryptoAmount, ok := convertFiatToCrypto(feeResult.NetAmount, quote.PriceInBRLCents, 2)
	if !ok || cryptoAmount <= 0 {
		return nil, ErrCryptoInvalidAmount
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = uowTx.Rollback(ctx)
	}()

	accountID, err := findActiveAccountID(ctx, uowTx, input.UserID)
	if err != nil {
		return nil, err
	}

	existingTx, err := uowTx.TransactionRepository().GetByIdempotencyKey(ctx, accountID, input.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existingTx != nil {
		existingTransfer, err := uowTx.CryptoTransferRepository().GetByTransactionID(ctx, existingTx.ID)
		if err != nil {
			return nil, err
		}
		if existingTransfer != nil {
			return existingTransfer, nil
		}
	}

	ledgerBalance, err := uowTx.TransactionRepository().GetLedgerBalance(ctx, accountID)
	if err != nil {
		return nil, err
	}
	holdBalance, err := uowTx.TransactionRepository().GetHoldBalance(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if ledgerBalance-holdBalance < feeResult.NetAmount {
		return nil, repository.ErrInsufficientFunds
	}

	now := uc.clock.Now().UTC()
	withdrawal := &entity.Transaction{
		ID:             uuid.NewString(),
		AccountID:      accountID,
		UserID:         input.UserID,
		Type:           entity.TransactionTypeWithdrawal,
		Status:         entity.TransactionStatusHold,
		Amount:         input.FiatAmount,
		Fee:            feeResult.Fee,
		NetAmount:      feeResult.NetAmount,
		IdempotencyKey: input.IdempotencyKey,
		ExternalRef:    normalizeExternalRef("", entity.TransactionTypeWithdrawal, "CRYPTO_PAY"),
		OccurredAt:     now,
		CreatedAt:      now,
	}
	if err := uowTx.TransactionRepository().Create(ctx, withdrawal); err != nil {
		return nil, err
	}
	if err := appendConversionAudit(ctx, uowTx, uc.clock, conversionAuditInput{
		UserID:          input.UserID,
		OperationType:   entity.PricingOperationPixToCrypto,
		Asset:           string(asset),
		GrossAmount:     feeResult.Amount,
		Fee:             feeResult.Fee,
		NetAmount:       feeResult.NetAmount,
		QuotePrice:      quote.PriceInBRLCents,
		SpreadBps:       0,
		LiquiditySource: "EXCHANGE",
		RelatedType:     "transaction",
		RelatedID:       withdrawal.ID,
		QuotedAt:        &quoteAt,
	}); err != nil {
		return nil, err
	}

	transfer := &entity.CryptoTransfer{
		ID:            uuid.NewString(),
		UserID:        input.UserID,
		AccountID:     accountID,
		TransactionID: withdrawal.ID,
		Asset:         string(asset),
		Network:       string(network),
		Address:       resolved.Normalized,
		Amount:        cryptoAmount,
		Fee:           0,
		Status:        entity.CryptoTransferPendingExchange,
		Direction:     "BUY",
		CreatedAt:     now,
	}
	if err := uowTx.CryptoTransferRepository().Create(ctx, transfer); err != nil {
		return nil, err
	}
	_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "CRYPTO_PAY_CREATED", "crypto_transfer", transfer.ID, map[string]any{
		"asset":   transfer.Asset,
		"network": transfer.Network,
	})
	if feeResult.Fee > 0 {
		_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "PRICING_APPLIED", "transaction", withdrawal.ID, map[string]any{
			"operation":  string(entity.PricingOperationPixToCrypto),
			"amount":     feeResult.Amount,
			"fee":        feeResult.Fee,
			"net_amount": feeResult.NetAmount,
			"rule_id":    ruleID(rule),
		})
	}

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}

	if err := uc.confirmCryptoTransfer(ctx, transfer, withdrawal.ID); err != nil {
		_ = uc.rejectCryptoTransfer(ctx, transfer, withdrawal.ID)
		return transfer, err
	}

	return transfer, nil
}

func (uc *PayCryptoWithFiatUseCase) resolveFee(ctx context.Context, input PayCryptoWithFiatInput) (entity.FeeResult, *entity.PricingRule, error) {
	if input.SkipPricing || uc.pricing == nil {
		netAmount := input.FiatAmount - input.FeeOverride
		if netAmount < 0 {
			netAmount = 0
		}
		return entity.FeeResult{Amount: input.FiatAmount, Fee: input.FeeOverride, NetAmount: netAmount}, nil, nil
	}
	return uc.pricing.Execute(ctx, PricingInput{
		UserID:        input.UserID,
		OperationType: entity.PricingOperationPixToCrypto,
		Asset:         "BRL",
		GrossAmount:   input.FiatAmount,
		FeatureCode:   string(entity.PricingOperationPixToCrypto),
	})
}

func (uc *PayCryptoWithFiatUseCase) confirmCryptoTransfer(ctx context.Context, transfer *entity.CryptoTransfer, transactionID string) error {
	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = uowTx.Rollback(ctx)
	}()

	if transfer.Status == entity.CryptoTransferConfirmed {
		return nil
	}

	_, err = uc.exchange.Execute(ctx, transfer.Asset, transfer.Amount, "BUY")
	if err != nil {
		return NewExternalDependencyError("exchange", "execute", err)
	}
	if uc.custody == nil {
		return NewExternalDependencyError("custody", "send_transfer", errors.New("custodia nao configurada"))
	}
	custodyTransfer, err := uc.custody.SendTransfer(ctx, entity.CustodyTransfer{
		UserID:      transfer.UserID,
		Network:     transfer.Network,
		Asset:       transfer.Asset,
		Address:     transfer.Address,
		Amount:      transfer.Amount,
		ExternalRef: "CRYPTO_PAY",
	})
	if err != nil {
		return NewExternalDependencyError("custody", "send_transfer", err)
	}

	if _, err := uowTx.TransactionRepository().UpdateStatusIfCurrent(ctx, transactionID, entity.TransactionStatusHold, entity.TransactionStatusConfirmed); err != nil {
		return err
	}

	if err := uowTx.CryptoTransferRepository().UpdateStatus(ctx, transfer.ID, entity.CryptoTransferConfirmed, custodyTransfer.ProviderID); err != nil {
		return err
	}

	_ = appendAudit(ctx, uowTx, uc.clock, transfer.UserID, "CRYPTO_PAY_CONFIRMED", "crypto_transfer", transfer.ID, map[string]any{
		"transaction_id": transactionID,
	})

	if err := uowTx.Commit(ctx); err != nil {
		return err
	}

	transfer.Status = entity.CryptoTransferConfirmed
	return nil
}

func (uc *PayCryptoWithFiatUseCase) rejectCryptoTransfer(ctx context.Context, transfer *entity.CryptoTransfer, transactionID string) error {
	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = uowTx.Rollback(ctx)
	}()

	tx, err := uowTx.TransactionRepository().GetByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if tx.Status != entity.TransactionStatusConfirmed {
		_, _ = uowTx.TransactionRepository().UpdateStatusIfCurrent(ctx, transactionID, entity.TransactionStatusHold, entity.TransactionStatusConfirmed)
	}

	reversalKey := transactionID + ":reversal"
	existing, err := uowTx.TransactionRepository().GetByIdempotencyKey(ctx, tx.AccountID, reversalKey)
	if err != nil {
		return err
	}
	if existing == nil {
		now := uc.clock.Now().UTC()
		reversal := &entity.Transaction{
			ID:             uuid.NewString(),
			AccountID:      tx.AccountID,
			UserID:         tx.UserID,
			Type:           entity.TransactionTypeReversal,
			Status:         entity.TransactionStatusConfirmed,
			Amount:         tx.NetAmount,
			Fee:            0,
			NetAmount:      tx.NetAmount,
			IdempotencyKey: reversalKey,
			ExternalRef:    normalizeExternalRef("", entity.TransactionTypeReversal, "CRYPTO_PAY_REVERSAL"),
			ReversalOf:     transactionID,
			OccurredAt:     now,
			CreatedAt:      now,
		}
		if err := uowTx.TransactionRepository().Create(ctx, reversal); err != nil {
			return err
		}
	}

	_ = uowTx.CryptoTransferRepository().UpdateStatus(ctx, transfer.ID, entity.CryptoTransferRejected, "")
	_ = appendAudit(ctx, uowTx, uc.clock, transfer.UserID, "CRYPTO_PAY_REJECTED", "crypto_transfer", transfer.ID, map[string]any{
		"transaction_id": transactionID,
	})

	return uowTx.Commit(ctx)
}

func findActiveAccountID(ctx context.Context, uowTx ports.UnitOfWorkTx, userID string) (string, error) {
	accounts, err := uowTx.AccountRepository().ListByUser(ctx, userID)
	if err != nil {
		return "", err
	}
	for _, account := range accounts {
		if account.Status == entity.AccountStatusActive {
			return account.ID, nil
		}
	}
	return "", ErrCryptoAccountNotFound
}


