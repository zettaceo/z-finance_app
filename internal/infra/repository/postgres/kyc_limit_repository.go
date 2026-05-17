package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type KycLimitRepository struct {
	pool *pgxpool.Pool
}

func NewKycLimitRepository(pool *pgxpool.Pool) *KycLimitRepository {
	return &KycLimitRepository{pool: pool}
}

func (r *KycLimitRepository) Upsert(ctx context.Context, limit *entity.KycLimit) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kyc_limits (level, daily_limit, monthly_limit)
		VALUES ($1, $2, $3)
		ON CONFLICT (level) DO UPDATE
		SET daily_limit = EXCLUDED.daily_limit,
		    monthly_limit = EXCLUDED.monthly_limit`,
		limit.Level,
		limit.DailyLimit,
		limit.MonthlyLimit,
	)
	return err
}

func (r *KycLimitRepository) GetByLevel(ctx context.Context, level entity.KYCLevel) (*entity.KycLimit, error) {
	row := r.pool.QueryRow(ctx, "SELECT level, daily_limit, monthly_limit, created_at FROM kyc_limits WHERE level = $1", level)

	var limit entity.KycLimit
	if err := row.Scan(&limit.Level, &limit.DailyLimit, &limit.MonthlyLimit, &limit.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &limit, nil
}

func (r *KycLimitRepository) ListAll(ctx context.Context) ([]*entity.KycLimit, error) {
	rows, err := r.pool.Query(ctx, "SELECT level, daily_limit, monthly_limit, created_at FROM kyc_limits ORDER BY level")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.KycLimit
	for rows.Next() {
		var limit entity.KycLimit
		if err := rows.Scan(&limit.Level, &limit.DailyLimit, &limit.MonthlyLimit, &limit.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &limit)
	}

	return items, rows.Err()
}

var _ repository.KycLimitRepository = (*KycLimitRepository)(nil)
