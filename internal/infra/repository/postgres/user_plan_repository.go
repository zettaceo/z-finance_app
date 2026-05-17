package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type UserPlanRepository struct {
	db dbExecutor
}

func NewUserPlanRepository(pool *pgxpool.Pool) *UserPlanRepository {
	return &UserPlanRepository{db: pool}
}

func NewUserPlanRepositoryWithTx(tx dbExecutor) *UserPlanRepository {
	return &UserPlanRepository{db: tx}
}

func (r *UserPlanRepository) GetActiveByUser(ctx context.Context, userID string, at time.Time) (*entity.UserPlan, error) {
	row := r.db.QueryRow(ctx, `
		SELECT user_id, plan_id, valid_from, valid_until, created_at
		  FROM user_plans
		 WHERE user_id = $1
		   AND valid_from <= $2
		   AND (valid_until IS NULL OR valid_until >= $2)
		 ORDER BY valid_from DESC
		 LIMIT 1`, userID, at)
	var plan entity.UserPlan
	var validUntil *time.Time
	if err := row.Scan(
		&plan.UserID,
		&plan.PlanID,
		&plan.ValidFrom,
		&validUntil,
		&plan.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	plan.ValidUntil = validUntil
	return &plan, nil
}

var _ repository.UserPlanRepository = (*UserPlanRepository)(nil)

func (r *UserPlanRepository) Create(ctx context.Context, plan *entity.UserPlan) error {
	if plan == nil {
		return repository.ErrInvalidState
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_plans (user_id, plan_id, valid_from, valid_until, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		plan.UserID,
		plan.PlanID,
		plan.ValidFrom,
		plan.ValidUntil,
		plan.CreatedAt,
	)
	return err
}
