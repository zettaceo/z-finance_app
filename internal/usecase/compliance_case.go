package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
)

var ErrComplianceInvalidInput = errors.New("dados de compliance invalidos")

type OpenComplianceCaseInput struct {
	UserID    string
	Type      entity.ComplianceCaseType
	RiskLevel entity.ComplianceRiskLevel
	Title     string
	Summary   string
	Metadata  map[string]any
}

type OpenComplianceCaseUseCase struct {
	uow   ports.UnitOfWork
	clock Clock
}

func NewOpenComplianceCaseUseCase(uow ports.UnitOfWork) *OpenComplianceCaseUseCase {
	return &OpenComplianceCaseUseCase{uow: uow, clock: NewRealClock()}
}

func (uc *OpenComplianceCaseUseCase) Execute(ctx context.Context, input OpenComplianceCaseInput) (*entity.ComplianceCase, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "OpenComplianceCase")
	defer end()

	if input.UserID == "" || input.Type == "" || input.RiskLevel == "" {
		return nil, ErrComplianceInvalidInput
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	payload, err := json.Marshal(input.Metadata)
	if err != nil {
		return nil, err
	}

	now := uc.clock.Now().UTC()
	item := &entity.ComplianceCase{
		ID:        uuid.NewString(),
		UserID:    input.UserID,
		Type:      input.Type,
		Status:    entity.ComplianceCaseOpen,
		RiskLevel: input.RiskLevel,
		Title:     input.Title,
		Summary:   input.Summary,
		Metadata:  payload,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uowTx.ComplianceCaseRepository().Create(ctx, item); err != nil {
		return nil, err
	}

	_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "COMPLIANCE_CASE_OPENED", "compliance_case", item.ID, map[string]any{
		"type":       item.Type,
		"risk_level": item.RiskLevel,
	})

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}

	return item, nil
}
