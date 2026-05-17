package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"z-finance-api/internal/conversion/engine"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/repository"
)

var ErrConversionRuleNotFound = errors.New("regra de conversao nao encontrada")

type AutoConvertPixUseCase struct {
	uow      ports.UnitOfWork
	exchange ports.ExchangeGateway
	clock    Clock
	engine   *engine.Engine
	settings repository.UserSettingsRepository
	pricing  *ResolvePricingUseCase
}

func NewAutoConvertPixUseCase(uow ports.UnitOfWork, exchange ports.ExchangeGateway, engine *engine.Engine, settings repository.UserSettingsRepository, pricing *ResolvePricingUseCase) *AutoConvertPixUseCase {
	return &AutoConvertPixUseCase{uow: uow, exchange: exchange, clock: NewRealClock(), engine: engine, settings: settings, pricing: pricing}
}

func (uc *AutoConvertPixUseCase) Execute(ctx context.Context, transfer *entity.PixTransfer) (*entity.CryptoTransfer, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "AutoConvertPix")
	defer end()

	if transfer == nil || transfer.Direction != entity.PixDirectionIn {
		return nil, ErrConversionRuleNotFound
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	if uc.engine == nil {
		return nil, ErrConversionRuleNotFound
	}
	result, err := uc.engine.ResolvePixIn(ctx, transfer.UserID)
	if err != nil {
		return nil, err
	}
	if result == nil || result.TargetAsset == "" {
		return nil, ErrConversionRuleNotFound
	}

	if uc.settings != nil {
		settings, err := uc.settings.GetByUserID(ctx, transfer.UserID)
		if err != nil {
			return nil, err
		}
		if settings != nil {
			if !settings.AutoConvertEnabled {
				return nil, ErrConversionRuleNotFound
			}
			if settings.AutoConvertMinAmount > 0 && transfer.NetAmount < settings.AutoConvertMinAmount {
				return nil, ErrConversionRuleNotFound
			}
		}
	}

	idempotencyKey := "auto-pix-" + transfer.ID
	existingTx, err := uowTx.TransactionRepository().GetByIdempotencyKey(ctx, transfer.AccountID, idempotencyKey)
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

	feeResult, rule, err := uc.resolveFee(ctx, transfer)
	if err != nil {
		return nil, err
	}
	quote, err := uc.exchange.Quote(ctx, result.TargetAsset)
	if err != nil {
		return nil, NewExternalDependencyError("exchange", "quote", err)
	}
	quoteAt := uc.clock.Now().UTC()
	cryptoAmount, ok := convertFiatToCrypto(feeResult.NetAmount, quote.PriceInBRLCents, 2)
	if !ok || cryptoAmount <= 0 {
		return nil, ErrCryptoInvalidAmount
	}

	now := uc.clock.Now().UTC()
	withdrawal := &entity.Transaction{
		ID:             uuid.NewString(),
		AccountID:      transfer.AccountID,
		UserID:         transfer.UserID,
		Type:           entity.TransactionTypeTradeBuy,
		Status:         entity.TransactionStatusHold,
		Amount:         feeResult.Amount,
		Fee:            feeResult.Fee,
		NetAmount:      feeResult.NetAmount,
		IdempotencyKey: idempotencyKey,
		ExternalRef:    normalizeExternalRef("", entity.TransactionTypeTradeBuy, "AUTO_PIX_IN"),
		OccurredAt:     now,
		CreatedAt:      now,
	}
	if err := uowTx.TransactionRepository().Create(ctx, withdrawal); err != nil {
		return nil, err
	}
	if err := appendConversionAudit(ctx, uowTx, uc.clock, conversionAuditInput{
		UserID:          transfer.UserID,
		OperationType:   entity.PricingOperationPixToCrypto,
		Asset:           result.TargetAsset,
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

	cryptoTransfer := &entity.CryptoTransfer{
		ID:            uuid.NewString(),
		UserID:        transfer.UserID,
		AccountID:     transfer.AccountID,
		TransactionID: withdrawal.ID,
		Asset:         result.TargetAsset,
		Network:       "INTERNAL",
		Address:       "AUTO_CONVERSION",
		Amount:        cryptoAmount,
		Fee:           0,
		Status:        entity.CryptoTransferConfirmed,
		Direction:     "BUY",
		CreatedAt:     now,
	}
	if err := uowTx.CryptoTransferRepository().Create(ctx, cryptoTransfer); err != nil {
		return nil, err
	}
	if _, err := uowTx.TransactionRepository().UpdateStatusIfCurrent(ctx, withdrawal.ID, entity.TransactionStatusHold, entity.TransactionStatusConfirmed); err != nil {
		return nil, err
	}

	txHash, err := uc.exchange.Execute(ctx, result.TargetAsset, cryptoAmount, "BUY")
	if err != nil {
		return nil, NewExternalDependencyError("exchange", "execute", err)
	}
	if err := uowTx.CryptoTransferRepository().UpdateStatus(ctx, cryptoTransfer.ID, entity.CryptoTransferConfirmed, txHash); err != nil {
		return nil, err
	}

	_ = appendAudit(ctx, uowTx, uc.clock, transfer.UserID, "AUTO_PIX_CONVERT", "crypto_transfer", cryptoTransfer.ID, map[string]any{
		"pix_transfer_id": transfer.ID,
	})
	if feeResult.Fee > 0 {
		_ = appendAudit(ctx, uowTx, uc.clock, transfer.UserID, "PRICING_APPLIED", "transaction", withdrawal.ID, map[string]any{
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
	return cryptoTransfer, nil
}

func (uc *AutoConvertPixUseCase) resolveFee(ctx context.Context, transfer *entity.PixTransfer) (entity.FeeResult, *entity.PricingRule, error) {
	if uc.pricing == nil {
		return entity.FeeResult{Amount: transfer.NetAmount, Fee: 0, NetAmount: transfer.NetAmount}, nil, nil
	}
	return uc.pricing.Execute(ctx, PricingInput{
		UserID:        transfer.UserID,
		OperationType: entity.PricingOperationPixToCrypto,
		Asset:         "BRL",
		GrossAmount:   transfer.NetAmount,
		FeatureCode:   string(entity.PricingOperationPixToCrypto),
	})
}
