package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PlanRepository struct {
	db dbExecutor
}

func NewPlanRepository(pool *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{db: pool}
}

func NewPlanRepositoryWithTx(tx dbExecutor) *PlanRepository {
	return &PlanRepository{db: tx}
}

func (r *PlanRepository) GetByID(ctx context.Context, id string) (*entity.Plan, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, code, description, monthly_price_cents, created_at
		  FROM plans
		 WHERE id = $1`, id)
	return scanPlan(row)
}

func (r *PlanRepository) GetByCode(ctx context.Context, code string) (*entity.Plan, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, code, description, monthly_price_cents, created_at
		  FROM plans
		 WHERE code = $1`, code)
	return scanPlan(row)
}

func (r *PlanRepository) ListAll(ctx context.Context) ([]*entity.Plan, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, description, monthly_price_cents, created_at
		  FROM plans
		 ORDER BY code ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.Plan
	for rows.Next() {
		var plan entity.Plan
		var description *string
		if err := rows.Scan(
			&plan.ID,
			&plan.Code,
			&description,
			&plan.MonthlyPriceCents,
			&plan.CreatedAt,
		); err != nil {
			return nil, err
		}
		if description != nil {
			plan.Description = *description
		}
		items = append(items, &plan)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PlanRepository) Create(ctx context.Context, plan *entity.Plan) error {
	if plan == nil {
		return repository.ErrInvalidState
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO plans (id, code, description, monthly_price_cents, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		plan.ID,
		plan.Code,
		nullableString(plan.Description),
		plan.MonthlyPriceCents,
		plan.CreatedAt,
	)
	return err
}

func scanPlan(row pgx.Row) (*entity.Plan, error) {
	var plan entity.Plan
	var description *string
	if err := row.Scan(
		&plan.ID,
		&plan.Code,
		&description,
		&plan.MonthlyPriceCents,
		&plan.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	if description != nil {
		plan.Description = *description
	}
	return &plan, nil
}

var _ repository.PlanRepository = (*PlanRepository)(nil)

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
