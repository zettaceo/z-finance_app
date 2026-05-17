package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PlanLimitRepository struct {
	db *pgxpool.Pool
}

func NewPlanLimitRepository(pool *pgxpool.Pool) *PlanLimitRepository {
	return &PlanLimitRepository{db: pool}
}

func (r *PlanLimitRepository) ListByPlan(ctx context.Context, planID string) ([]*entity.PlanLimit, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, plan_id, limit_code, limit_value, limit_window, created_at, updated_at
		  FROM plan_limits
		 WHERE plan_id = $1
		 ORDER BY limit_code ASC`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.PlanLimit
	for rows.Next() {
		var item entity.PlanLimit
		if err := rows.Scan(
			&item.ID,
			&item.PlanID,
			&item.LimitCode,
			&item.LimitValue,
			&item.LimitWindow,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *PlanLimitRepository) Upsert(ctx context.Context, limit *entity.PlanLimit) error {
	if limit == nil {
		return repository.ErrInvalidState
	}
	now := time.Now().UTC()
	if limit.CreatedAt.IsZero() {
		limit.CreatedAt = now
	}
	limit.UpdatedAt = now
	_, err := r.db.Exec(ctx, `
		INSERT INTO plan_limits (id, plan_id, limit_code, limit_value, limit_window, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (plan_id, limit_code, limit_window) DO UPDATE SET
			limit_value = EXCLUDED.limit_value,
			updated_at = EXCLUDED.updated_at`,
		limit.ID,
		limit.PlanID,
		limit.LimitCode,
		limit.LimitValue,
		limit.LimitWindow,
		limit.CreatedAt,
		limit.UpdatedAt,
	)
	return err
}

var _ repository.PlanLimitRepository = (*PlanLimitRepository)(nil)
