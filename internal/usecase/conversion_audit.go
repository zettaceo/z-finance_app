package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
)

type conversionAuditInput struct {
	UserID          string
	OperationType   entity.PricingOperationType
	Asset           string
	GrossAmount     int64
	Fee             int64
	NetAmount       int64
	QuotePrice      int64
	SpreadBps       int64
	LiquiditySource string
	RelatedType     string
	RelatedID       string
	QuotedAt        *time.Time
}

func appendConversionAudit(ctx context.Context, uowTx ports.UnitOfWorkTx, clock Clock, input conversionAuditInput) error {
	if uowTx == nil {
		return nil
	}
	if input.UserID == "" || input.Asset == "" || input.OperationType == "" {
		return nil
	}
	quotedAt := input.QuotedAt
	now := clock.Now().UTC()
	if quotedAt == nil {
		quotedAt = &now
	}
	liquiditySource := input.LiquiditySource
	if liquiditySource == "" {
		liquiditySource = "UNKNOWN"
	}
	audit := &entity.ConversionAudit{
		ID:              uuid.NewString(),
		UserID:          input.UserID,
		OperationType:   input.OperationType,
		Asset:           input.Asset,
		GrossAmount:     input.GrossAmount,
		Fee:             input.Fee,
		NetAmount:       input.NetAmount,
		QuotePrice:      input.QuotePrice,
		SpreadBps:       input.SpreadBps,
		LiquiditySource: liquiditySource,
		QuotedAt:        quotedAt,
		RelatedType:     input.RelatedType,
		RelatedID:       input.RelatedID,
		CreatedAt:       now,
	}
	return uowTx.ConversionAuditRepository().Append(ctx, audit)
}
