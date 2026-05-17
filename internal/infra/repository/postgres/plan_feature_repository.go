package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PlanFeatureRepository struct {
	db *pgxpool.Pool
}

func NewPlanFeatureRepository(pool *pgxpool.Pool) *PlanFeatureRepository {
	return &PlanFeatureRepository{db: pool}
}

func (r *PlanFeatureRepository) ListByPlan(ctx context.Context, planID string) ([]*entity.PlanFeature, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, plan_id, feature_code, enabled, metadata, created_at, updated_at
		  FROM plan_features
		 WHERE plan_id = $1
		 ORDER BY feature_code ASC`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.PlanFeature
	for rows.Next() {
		var item entity.PlanFeature
		var metadata []byte
		if err := rows.Scan(&item.ID, &item.PlanID, &item.FeatureCode, &item.Enabled, &metadata, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Metadata = metadata
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *PlanFeatureRepository) Upsert(ctx context.Context, feature *entity.PlanFeature) error {
	if feature == nil {
		return repository.ErrInvalidState
	}
	now := time.Now().UTC()
	if feature.CreatedAt.IsZero() {
		feature.CreatedAt = now
	}
	feature.UpdatedAt = now
	_, err := r.db.Exec(ctx, `
		INSERT INTO plan_features (id, plan_id, feature_code, enabled, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (plan_id, feature_code) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at`,
		feature.ID,
		feature.PlanID,
		feature.FeatureCode,
		feature.Enabled,
		feature.Metadata,
		feature.CreatedAt,
		feature.UpdatedAt,
	)
	return err
}

var _ repository.PlanFeatureRepository = (*PlanFeatureRepository)(nil)
