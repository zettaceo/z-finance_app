package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type UserLimitOverrideRepository struct {
	db dbExecutor
}

func NewUserLimitOverrideRepository(pool *pgxpool.Pool) *UserLimitOverrideRepository {
	return &UserLimitOverrideRepository{db: pool}
}

func NewUserLimitOverrideRepositoryWithTx(tx dbExecutor) *UserLimitOverrideRepository {
	return &UserLimitOverrideRepository{db: tx}
}

func (r *UserLimitOverrideRepository) ListByUser(ctx context.Context, userID string) ([]*entity.UserLimitOverride, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, limit_code, limit_value, limit_window, reason, created_at, updated_at
		  FROM user_limit_overrides
		 WHERE user_id = $1
		 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entity.UserLimitOverride
	for rows.Next() {
		var item entity.UserLimitOverride
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.LimitCode,
			&item.LimitValue,
			&item.LimitWindow,
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

func (r *UserLimitOverrideRepository) Upsert(ctx context.Context, override *entity.UserLimitOverride) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_limit_overrides (
			id, user_id, limit_code, limit_value, limit_window, reason, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		ON CONFLICT (user_id, limit_code, limit_window)
		DO UPDATE SET limit_value = EXCLUDED.limit_value,
		              reason = EXCLUDED.reason,
		              updated_at = EXCLUDED.updated_at`,
		override.ID,
		override.UserID,
		override.LimitCode,
		override.LimitValue,
		override.LimitWindow,
		override.Reason,
		override.CreatedAt,
		override.UpdatedAt,
	)
	return err
}

var _ repository.UserLimitOverrideRepository = (*UserLimitOverrideRepository)(nil)
