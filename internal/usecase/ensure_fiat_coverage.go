package usecase

import (
	"context"
	"errors"

	"z-finance-api/internal/conversion/engine"
	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
)

var ErrCoverageInvalidInput = errors.New("dados invalidos")

type EnsureFiatCoverageInput struct {
	UserID         string
	AccountID      string
	RequiredAmount int64
	IdempotencyKey string
	ExternalRef    string
	Trigger        entity.ConversionTrigger
}

type EnsureFiatCoverageUseCase struct {
	uow      ports.UnitOfWork
	exchange ports.ExchangeGateway
	engine   *engine.Engine
	pricing  *ResolvePricingUseCase
	clock    Clock
}

func NewEnsureFiatCoverageUseCase(uow ports.UnitOfWork, exchange ports.ExchangeGateway, engine *engine.Engine, pricing *ResolvePricingUseCase) *EnsureFiatCoverageUseCase {
	return &EnsureFiatCoverageUseCase{
		uow:      uow,
		exchange: exchange,
		engine:   engine,
		pricing:  pricing,
		clock:    NewRealClock(),
	}
}

func (uc *EnsureFiatCoverageUseCase) Execute(ctx context.Context, input EnsureFiatCoverageInput) (bool, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "EnsureFiatCoverage")
	defer end()

	if input.UserID == "" || input.AccountID == "" || input.RequiredAmount <= 0 || input.IdempotencyKey == "" {
		return false, ErrCoverageInvalidInput
	}
	if uc.exchange == nil {
		return false, ErrCoverageInvalidInput
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	account, err := uowTx.AccountRepository().GetByIDForUpdate(ctx, input.AccountID)
	if err != nil {
		return false, err
	}
	if account.UserID != input.UserID || account.Status != entity.AccountStatusActive {
		return false, ErrAccountInactive
	}

	ledgerBalance, err := uowTx.TransactionRepository().GetLedgerBalance(ctx, input.AccountID)
	if err != nil {
		return false, err
	}
	holdBalance, err := uowTx.TransactionRepository().GetHoldBalance(ctx, input.AccountID)
	if err != nil {
		return false, err
	}
	available := ledgerBalance - holdBalance
	if available >= input.RequiredAmount {
		return true, nil
	}
	deficit := input.RequiredAmount - available

	assets := []string{"USDT", "ETH", "BTC", "MATIC"}
	if uc.engine != nil {
		resolved, err := uc.engine.ResolveAssets(ctx, input.UserID, input.Trigger)
		if err != nil {
			return false, err
		}
		if len(resolved) > 0 {
			assets = resolved
		}
	}

	remaining, err := liquidateCryptoForDeficit(ctx, uowTx, uc.exchange, uc.clock, input.UserID, input.AccountID, input.IdempotencyKey, input.ExternalRef, deficit, assets, uc.pricing)
	if err != nil {
		return false, err
	}

	if err := uowTx.Commit(ctx); err != nil {
		return false, err
	}
	return remaining <= 0, nil
}
