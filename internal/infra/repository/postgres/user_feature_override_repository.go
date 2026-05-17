package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type UserFeatureOverrideRepository struct {
	db dbExecutor
}

func NewUserFeatureOverrideRepository(pool *pgxpool.Pool) *UserFeatureOverrideRepository {
	return &UserFeatureOverrideRepository{db: pool}
}

func NewUserFeatureOverrideRepositoryWithTx(tx dbExecutor) *UserFeatureOverrideRepository {
	return &UserFeatureOverrideRepository{db: tx}
}

func (r *UserFeatureOverrideRepository) ListByUser(ctx context.Context, userID string) ([]*entity.UserFeatureOverride, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, feature_code, enabled, reason, created_at, updated_at
		  FROM user_feature_overrides
		 WHERE user_id = $1
		 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entity.UserFeatureOverride
	for rows.Next() {
		var item entity.UserFeatureOverride
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.FeatureCode,
			&item.Enabled,
			&item.Reason,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &item)
	}
	return result, rows.Err()
}

func (r *UserFeatureOverrideRepository) Upsert(ctx context.Context, override *entity.UserFeatureOverride) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_feature_overrides (
			id, user_id, feature_code, enabled, reason, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (user_id, feature_code)
		DO UPDATE SET enabled = EXCLUDED.enabled,
		              reason = EXCLUDED.reason,
		              updated_at = EXCLUDED.updated_at`,
		override.ID,
		override.UserID,
		override.FeatureCode,
		override.Enabled,
		override.Reason,
		override.CreatedAt,
		override.UpdatedAt,
	)
	return err
}

var _ repository.UserFeatureOverrideRepository = (*UserFeatureOverrideRepository)(nil)
