package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var (
	ErrPixWebhookInvalid = errors.New("payload pix webhook invalido")
)

type PixWebhookInput struct {
	TransferID string
	Status     string
}

type HandlePixWebhookUseCase struct {
	uow ports.UnitOfWork
	clock Clock
	autoConvert *AutoConvertPixUseCase
}

func NewHandlePixWebhookUseCase(uow ports.UnitOfWork, autoConvert *AutoConvertPixUseCase) *HandlePixWebhookUseCase {
	return &HandlePixWebhookUseCase{uow: uow, clock: NewRealClock(), autoConvert: autoConvert}
}

func NewHandlePixWebhookUseCaseWithClock(uow ports.UnitOfWork, autoConvert *AutoConvertPixUseCase, clock Clock) *HandlePixWebhookUseCase {
	return &HandlePixWebhookUseCase{uow: uow, clock: clock, autoConvert: autoConvert}
}

func (uc *HandlePixWebhookUseCase) Execute(ctx context.Context, input PixWebhookInput) (*entity.PixTransfer, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "HandlePixWebhook")
	defer end()

	if input.TransferID == "" || input.Status == "" {
		return nil, ErrPixWebhookInvalid
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = uowTx.Rollback(ctx)
	}()

	transfer, err := uowTx.PixTransferRepository().GetByID(ctx, input.TransferID)
	if err != nil {
		return nil, err
	}
	if transfer == nil {
		return nil, repository.ErrNotFound
	}

	if transfer.Status == entity.PixStatusConfirmed || transfer.Status == entity.PixStatusRejected {
		if transfer.Status == entity.PixStatusConfirmed && transfer.Direction == entity.PixDirectionIn && uc.autoConvert != nil {
			if _, err := uc.autoConvert.Execute(ctx, transfer); err != nil && !errors.Is(err, ErrConversionRuleNotFound) {
				return nil, err
			}
		}
		return transfer, nil
	}

	now := uc.clock.Now().UTC()
	switch strings.ToUpper(input.Status) {
	case "CONFIRMED":
		if err := uowTx.PixTransferRepository().UpdateStatusWithConfirmation(ctx, transfer.ID, entity.PixStatusConfirmed, &now); err != nil {
			return nil, err
		}
		if transfer.Direction == entity.PixDirectionIn {
			if transfer.TransactionID == "" {
				deposit := &entity.Transaction{
					ID:             uuid.NewString(),
					AccountID:      transfer.AccountID,
					UserID:         transfer.UserID,
					Type:           entity.TransactionTypeDeposit,
					Status:         entity.TransactionStatusConfirmed,
					Amount:         transfer.Amount,
					Fee:            transfer.Fee,
					NetAmount:      transfer.NetAmount,
					IdempotencyKey: transfer.IdempotencyKey + ":deposit",
					ExternalRef:    normalizeExternalRef("", entity.TransactionTypeDeposit, "PIX_IN"),
					OccurredAt:     now,
					CreatedAt:      now,
				}
				if err := uowTx.TransactionRepository().Create(ctx, deposit); err != nil {
					return nil, err
				}
				if err := uowTx.PixTransferRepository().UpdateTransactionID(ctx, transfer.ID, deposit.ID); err != nil {
					return nil, err
				}
				transfer.TransactionID = deposit.ID
			}
		} else {
			tx, err := uowTx.TransactionRepository().UpdateStatusIfCurrent(ctx, transfer.TransactionID, entity.TransactionStatusHold, entity.TransactionStatusConfirmed)
			if err != nil {
				return nil, err
			}
			if tx == nil {
				return nil, repository.ErrInvalidState
			}
		}
		_ = appendAudit(ctx, uowTx, uc.clock, transfer.UserID, "PIX_CONFIRMED", "pix_transfer", transfer.ID, map[string]any{
			"transaction_id": transfer.TransactionID,
		})
		transfer.Status = entity.PixStatusConfirmed
		transfer.ConfirmedAt = &now
	case "REJECTED":
		if err := uowTx.PixTransferRepository().UpdateStatusWithConfirmation(ctx, transfer.ID, entity.PixStatusRejected, &now); err != nil {
			return nil, err
		}
		tx, err := uowTx.TransactionRepository().UpdateStatusIfCurrent(ctx, transfer.TransactionID, entity.TransactionStatusHold, entity.TransactionStatusRejected)
		if err != nil {
			return nil, err
		}
		if tx == nil {
			existing, err := uowTx.TransactionRepository().GetByID(ctx, transfer.TransactionID)
			if err != nil {
				return nil, err
			}
			if existing.Status != entity.TransactionStatusConfirmed {
				tx = existing
			} else {
				tx = existing
			}
		}

		reversalKey := transfer.IdempotencyKey + ":reversal"
		existingReversal, err := uowTx.TransactionRepository().GetByIdempotencyKey(ctx, transfer.AccountID, reversalKey)
		if err != nil {
			return nil, err
		}
		if existingReversal == nil {
			reversal := &entity.Transaction{
				ID:             uuid.NewString(),
				AccountID:      transfer.AccountID,
				UserID:         transfer.UserID,
				Type:           entity.TransactionTypeReversal,
				Status:         entity.TransactionStatusConfirmed,
				Amount:         tx.NetAmount,
				Fee:            0,
				NetAmount:      tx.NetAmount,
				IdempotencyKey: reversalKey,
				ExternalRef:    normalizeExternalRef("", entity.TransactionTypeReversal, "PIX_REJECT_REVERSAL"),
				ReversalOf:     transfer.TransactionID,
				OccurredAt:     now,
				CreatedAt:      now,
			}
			if err := uowTx.TransactionRepository().Create(ctx, reversal); err != nil {
				return nil, err
			}
		}
		_ = appendAudit(ctx, uowTx, uc.clock, transfer.UserID, "PIX_REJECTED", "pix_transfer", transfer.ID, map[string]any{
			"transaction_id": transfer.TransactionID,
			"hold_released":  true,
		})
		transfer.Status = entity.PixStatusRejected
		transfer.ConfirmedAt = &now
	default:
		return nil, ErrPixWebhookInvalid
	}

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}

	if transfer.Direction == entity.PixDirectionIn && transfer.Status == entity.PixStatusConfirmed && uc.autoConvert != nil {
		if _, err := uc.autoConvert.Execute(ctx, transfer); err != nil && !errors.Is(err, ErrConversionRuleNotFound) {
			return nil, err
		}
	}

	return transfer, nil
}
