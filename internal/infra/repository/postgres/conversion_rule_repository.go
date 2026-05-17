package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type ConversionRuleRepository struct {
	db dbExecutor
}

func NewConversionRuleRepository(pool *pgxpool.Pool) *ConversionRuleRepository {
	return &ConversionRuleRepository{db: pool}
}

func NewConversionRuleRepositoryWithTx(tx dbExecutor) *ConversionRuleRepository {
	return &ConversionRuleRepository{db: tx}
}

func (r *ConversionRuleRepository) GetActiveByUserAndTrigger(ctx context.Context, userID string, trigger entity.ConversionTrigger) (*entity.ConversionRule, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, trigger_event, source_asset, target_asset, enabled, created_at
		  FROM conversion_rules
		 WHERE user_id = $1 AND trigger_event = $2 AND enabled = TRUE
		 ORDER BY created_at DESC
		 LIMIT 1`, userID, trigger)
	var rule entity.ConversionRule
	if err := row.Scan(
		&rule.ID,
		&rule.UserID,
		&rule.Trigger,
		&rule.SourceAsset,
		&rule.TargetAsset,
		&rule.Enabled,
		&rule.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rule, nil
}

var _ repository.ConversionRuleRepository = (*ConversionRuleRepository)(nil)
