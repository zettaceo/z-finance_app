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

var (
	ErrPixIdempotencyRequired = errors.New("idempotency key obrigatoria")
	ErrPixInvalidAmount       = errors.New("valor pix invalido")
)

type SendPixInput struct {
	AccountID      string
	UserID         string
	Amount         int64
	Fee            int64
	NetAmount      int64
	SkipPricing    bool
	IdempotencyKey string
	ExternalRef    string
	Metadata       map[string]any
}

type SendPixUseCase struct {
	uow         ports.UnitOfWork
	transfers   repository.PixTransferRepository
	partner     ports.PixPartnerClient
	auditLogger repository.AuditLogRepository
	pricing     *ResolvePricingUseCase
	clock       Clock
}

func NewSendPixUseCase(uow ports.UnitOfWork, transfers repository.PixTransferRepository, partner ports.PixPartnerClient, audit repository.AuditLogRepository, pricing *ResolvePricingUseCase) *SendPixUseCase {
	return &SendPixUseCase{
		uow:         uow,
		transfers:   transfers,
		partner:     partner,
		auditLogger: audit,
		pricing:     pricing,
		clock:       NewRealClock(),
	}
}

func NewSendPixUseCaseWithClock(uow ports.UnitOfWork, transfers repository.PixTransferRepository, partner ports.PixPartnerClient, audit repository.AuditLogRepository, pricing *ResolvePricingUseCase, clock Clock) *SendPixUseCase {
	return &SendPixUseCase{
		uow:         uow,
		transfers:   transfers,
		partner:     partner,
		auditLogger: audit,
		pricing:     pricing,
		clock:       clock,
	}
}

func (uc *SendPixUseCase) Execute(ctx context.Context, input SendPixInput) (*entity.PixTransfer, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "SendPix")
	defer end()

	if input.IdempotencyKey == "" {
		return nil, ErrPixIdempotencyRequired
	}
	if input.Amount <= 0 {
		return nil, ErrPixInvalidAmount
	}
	feeResult, rule, err := uc.resolveFee(ctx, input)
	if err != nil {
		return nil, err
	}
	if feeResult.Fee < 0 || feeResult.Fee > input.Amount {
		return nil, ErrInvalidFee
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = uowTx.Rollback(ctx)
	}()

	account, err := uowTx.AccountRepository().GetByIDForUpdate(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}
	if account.UserID != input.UserID || account.Status != entity.AccountStatusActive {
		return nil, ErrAccountInactive
	}

	existing, err := uowTx.PixTransferRepository().GetByIdempotencyKey(ctx, input.AccountID, input.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	ledgerBalance, err := uowTx.TransactionRepository().GetLedgerBalance(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}
	holdBalance, err := uowTx.TransactionRepository().GetHoldBalance(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}
	available := ledgerBalance - holdBalance
	if available < feeResult.NetAmount {
		return nil, repository.ErrInsufficientFunds
	}

	now := uc.clock.Now().UTC()
	externalRef := normalizeExternalRef(input.ExternalRef, entity.TransactionTypeWithdrawal, "PIX_OUT")
	tx := &entity.Transaction{
		ID:             uuid.NewString(),
		AccountID:      input.AccountID,
		UserID:         input.UserID,
		Type:           entity.TransactionTypeWithdrawal,
		Status:         entity.TransactionStatusHold,
		Amount:         input.Amount,
		Fee:            feeResult.Fee,
		NetAmount:      feeResult.NetAmount,
		IdempotencyKey: input.IdempotencyKey,
		ExternalRef:    externalRef,
		OccurredAt:     now,
		CreatedAt:      now,
	}
	if err := uowTx.TransactionRepository().Create(ctx, tx); err != nil {
		return nil, err
	}

	transfer := &entity.PixTransfer{
		ID:             uuid.NewString(),
		TransactionID:  tx.ID,
		UserID:         input.UserID,
		AccountID:      input.AccountID,
		Direction:      entity.PixDirectionOut,
		Status:         entity.PixStatusCreated,
		Amount:         input.Amount,
		Fee:            feeResult.Fee,
		NetAmount:      feeResult.NetAmount,
		IdempotencyKey: input.IdempotencyKey,
		ExternalRef:    externalRef,
		Metadata:       input.Metadata,
		OccurredAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := uowTx.PixTransferRepository().Create(ctx, transfer); err != nil {
		return nil, err
	}

	if feeResult.Fee > 0 {
		_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "PRICING_APPLIED", "transaction", tx.ID, map[string]any{
			"operation":  string(entity.PricingOperationPixOut),
			"amount":     feeResult.Amount,
			"fee":        feeResult.Fee,
			"net_amount": feeResult.NetAmount,
			"rule_id":    ruleID(rule),
		})
	}

	_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "PIX_SEND_CREATED", "pix_transfer", transfer.ID, map[string]any{
		"status": "CREATED",
	})

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}

	if uc.partner != nil {
		if err := uc.partner.Send(ctx, transfer); err != nil {
			return transfer, err
		}
	}

	if err := uc.transfers.UpdateStatus(ctx, transfer.ID, entity.PixStatusPendingPartner); err != nil {
		return transfer, err
	}
	transfer.Status = entity.PixStatusPendingPartner

	if uc.auditLogger != nil {
		_ = uc.auditLogger.Append(ctx, &entity.AuditLog{
			ID:         uuid.NewString(),
			UserID:     input.UserID,
			Action:     "PIX_SEND_PENDING",
			EntityType: "pix_transfer",
			EntityID:   transfer.ID,
			CreatedAt:  uc.clock.Now().UTC(),
		})
	}

	return transfer, nil
}

func (uc *SendPixUseCase) resolveFee(ctx context.Context, input SendPixInput) (entity.FeeResult, *entity.PricingRule, error) {
	if input.SkipPricing || uc.pricing == nil {
		netAmount := input.NetAmount
		if netAmount == 0 && input.Amount > 0 {
			netAmount = input.Amount - input.Fee
		}
		return entity.FeeResult{Amount: input.Amount, Fee: input.Fee, NetAmount: netAmount}, nil, nil
	}
	return uc.pricing.Execute(ctx, PricingInput{
		UserID:        input.UserID,
		OperationType: entity.PricingOperationPixOut,
		Asset:         "BRL",
		GrossAmount:   input.Amount,
		FeatureCode:   string(entity.PricingOperationPixOut),
	})
}
