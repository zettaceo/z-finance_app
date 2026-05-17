package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type ConversionAuditRepository struct {
	db dbExecutor
}

func NewConversionAuditRepository(pool *pgxpool.Pool) *ConversionAuditRepository {
	return &ConversionAuditRepository{db: pool}
}

func NewConversionAuditRepositoryWithTx(tx dbExecutor) *ConversionAuditRepository {
	return &ConversionAuditRepository{db: tx}
}

func (r *ConversionAuditRepository) Append(ctx context.Context, audit *entity.ConversionAudit) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO conversion_audits (
			id, user_id, operation_type, asset, gross_amount, fee, net_amount,
			quote_price, spread_bps, liquidity_source, quoted_at, related_type, related_id, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, NULLIF($13, '')::uuid, $14
		)`,
		audit.ID,
		audit.UserID,
		audit.OperationType,
		audit.Asset,
		audit.GrossAmount,
		audit.Fee,
		audit.NetAmount,
		audit.QuotePrice,
		audit.SpreadBps,
		audit.LiquiditySource,
		audit.QuotedAt,
		audit.RelatedType,
		audit.RelatedID,
		audit.CreatedAt,
	)
	return err
}

var _ repository.ConversionAuditRepository = (*ConversionAuditRepository)(nil)
