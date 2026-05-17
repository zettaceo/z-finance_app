package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type VelocityRepository struct {
	pool *pgxpool.Pool
}

func NewVelocityRepository(pool *pgxpool.Pool) *VelocityRepository {
	return &VelocityRepository{pool: pool}
}

func (r *VelocityRepository) GetKYCLevel(ctx context.Context, userID string) (entity.KYCLevel, error) {
	var level entity.KYCLevel
	err := r.pool.QueryRow(ctx, "SELECT level FROM kyc_profiles WHERE user_id = $1", userID).Scan(&level)
	if err != nil {
		if err == pgx.ErrNoRows {
			return entity.KYCLevelUnverified, nil
		}
		return "", err
	}
	if level == "" {
		return entity.KYCLevelUnverified, nil
	}
	return level, nil
}

func (r *VelocityRepository) GetLimitsForLevel(ctx context.Context, level entity.KYCLevel) (int64, int64, bool, error) {
	var daily int64
	var monthly int64
	err := r.pool.QueryRow(ctx, "SELECT daily_limit, monthly_limit FROM kyc_limits WHERE level = $1", level).Scan(&daily, &monthly)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}
	return daily, monthly, true, nil
}

func (r *VelocityRepository) SumConfirmedSpentSince(ctx context.Context, userID string, from time.Time) (int64, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(net_amount), 0)
		FROM transactions
		WHERE user_id = $1
		  AND status = 'CONFIRMED'
		  AND type IN ('WITHDRAWAL', 'PAYMENT', 'TRADE_BUY', 'CARD_AUTH')
		  AND occurred_at >= $2`, userID, from)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *VelocityRepository) CountRecentTransactionsSince(ctx context.Context, userID string, from time.Time) (int64, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(COUNT(1), 0)
		FROM transactions
		WHERE user_id = $1
		  AND status IN ('CONFIRMED', 'HOLD', 'PENDING_PARTNER')
		  AND type IN ('WITHDRAWAL', 'PAYMENT', 'TRADE_BUY', 'CARD_AUTH')
		  AND occurred_at >= $2`, userID, from)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *VelocityRepository) SumRecentNetAmountSince(ctx context.Context, userID string, from time.Time) (int64, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(net_amount), 0)
		FROM transactions
		WHERE user_id = $1
		  AND status IN ('CONFIRMED', 'HOLD', 'PENDING_PARTNER')
		  AND type IN ('WITHDRAWAL', 'PAYMENT', 'TRADE_BUY', 'CARD_AUTH')
		  AND occurred_at >= $2`, userID, from)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

var _ repository.VelocityRepository = (*VelocityRepository)(nil)
