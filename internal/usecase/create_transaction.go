package usecase

import (
	"context"
	"errors"
	"time"

	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var (
	ErrInvalidAmount      = errors.New("valor invalido")
	ErrInvalidFee         = errors.New("taxa invalida")
	ErrInvalidNetAmount   = errors.New("net amount invalido")
	ErrMissingIdempotency = errors.New("idempotency key obrigatoria")
	ErrAccountInactive    = errors.New("conta bloqueada ou encerrada")
)

type CreateTransactionInput struct {
	ID             string
	AccountID      string
	UserID         string
	Type           entity.TransactionType
	Status         entity.TransactionStatus
	Amount         int64
	Fee            int64
	NetAmount      int64
	IdempotencyKey string
	ExternalRef    string
	ReversalOf     string
	OccurredAt     *time.Time
}

type CreateTransactionUseCase struct {
	uow ports.UnitOfWork
	clock Clock
}

func NewCreateTransactionUseCase(uow ports.UnitOfWork) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{uow: uow, clock: NewRealClock()}
}

func NewCreateTransactionUseCaseWithClock(uow ports.UnitOfWork, clock Clock) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{uow: uow, clock: clock}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateTransactionInput) (*entity.Transaction, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "CreateTransaction")
	defer end()

	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if input.Fee < 0 || input.Fee > input.Amount {
		return nil, ErrInvalidFee
	}
	if input.NetAmount != input.Amount-input.Fee || input.NetAmount < 0 {
		return nil, ErrInvalidNetAmount
	}
	if input.IdempotencyKey == "" {
		return nil, ErrMissingIdempotency
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = uowTx.Rollback(ctx)
	}()

	txRepo := uowTx.TransactionRepository()
	accountRepo := uowTx.AccountRepository()

	account, err := accountRepo.GetByIDForUpdate(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}
	if account.Status != entity.AccountStatusActive {
		return nil, ErrAccountInactive
	}

	existing, err := txRepo.GetByIdempotencyKey(ctx, input.AccountID, input.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	occurredAt := uc.clock.Now().UTC()
	if input.OccurredAt != nil {
		occurredAt = input.OccurredAt.UTC()
	}

	tx := &entity.Transaction{
		ID:             input.ID,
		AccountID:      input.AccountID,
		UserID:         input.UserID,
		Type:           input.Type,
		Status:         input.Status,
		Amount:         input.Amount,
		Fee:            input.Fee,
		NetAmount:      input.NetAmount,
		IdempotencyKey: input.IdempotencyKey,
		ExternalRef:    normalizeExternalRef(input.ExternalRef, input.Type, ""),
		ReversalOf:     input.ReversalOf,
		OccurredAt:     occurredAt,
		CreatedAt:      uc.clock.Now().UTC(),
	}

	if entity.IsDebitType(tx.Type) {
		ledgerBalance, err := txRepo.GetLedgerBalance(ctx, tx.AccountID)
		if err != nil {
			return nil, err
		}
		available := ledgerBalance
		if tx.Status == entity.TransactionStatusHold || tx.Status == entity.TransactionStatusPendingPartner {
			holdBalance, err := txRepo.GetHoldBalance(ctx, tx.AccountID)
			if err != nil {
				return nil, err
			}
			available = ledgerBalance - holdBalance
		}
		if available < tx.NetAmount {
			return nil, repository.ErrInsufficientFunds
		}
	}

	if err := txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}

	return tx, nil
}

func defaultExternalRef(value string, txType entity.TransactionType) string {
	return normalizeExternalRef(value, txType, string(txType))
}
