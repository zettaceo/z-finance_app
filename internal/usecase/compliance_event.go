package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var ErrComplianceEventInvalid = errors.New("evento de compliance invalido")

type AddComplianceEventInput struct {
	CaseID   string
	UserID   string
	Type     string
	Payload  map[string]any
}

type AddComplianceEventUseCase struct {
	uow   ports.UnitOfWork
	clock Clock
}

func NewAddComplianceEventUseCase(uow ports.UnitOfWork) *AddComplianceEventUseCase {
	return &AddComplianceEventUseCase{uow: uow, clock: NewRealClock()}
}

func (uc *AddComplianceEventUseCase) Execute(ctx context.Context, input AddComplianceEventInput) (*entity.ComplianceEvent, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "AddComplianceEvent")
	defer end()

	if input.CaseID == "" || input.Type == "" {
		return nil, ErrComplianceEventInvalid
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	if _, err := uowTx.ComplianceCaseRepository().GetByID(ctx, input.CaseID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		return nil, err
	}

	payload, err := json.Marshal(input.Payload)
	if err != nil {
		return nil, err
	}

	now := uc.clock.Now().UTC()
	item := &entity.ComplianceEvent{
		ID:        uuid.NewString(),
		CaseID:    input.CaseID,
		EventType: input.Type,
		Payload:   payload,
		CreatedAt: now,
	}
	if err := uowTx.ComplianceEventRepository().Append(ctx, item); err != nil {
		return nil, err
	}

	_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "COMPLIANCE_CASE_EVENT", "compliance_case", input.CaseID, map[string]any{
		"event_type": input.Type,
	})

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}
	return item, nil
}
