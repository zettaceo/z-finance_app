package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PreRegistrationAttemptRepository struct {
	db dbExecutor
}

func NewPreRegistrationAttemptRepository(pool *pgxpool.Pool) *PreRegistrationAttemptRepository {
	return &PreRegistrationAttemptRepository{db: pool}
}

func NewPreRegistrationAttemptRepositoryWithTx(tx dbExecutor) *PreRegistrationAttemptRepository {
	return &PreRegistrationAttemptRepository{db: tx}
}

func (r *PreRegistrationAttemptRepository) Append(ctx context.Context, attempt *entity.PreRegistrationAttempt) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO pre_registration_attempts (id, pre_registration_id, channel, success, reason, ip, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		attempt.ID,
		attempt.PreRegistrationID,
		attempt.Channel,
		attempt.Success,
		nullIfEmpty(attempt.Reason),
		nullIfEmpty(attempt.IP),
		attempt.CreatedAt,
	)
	return err
}

var _ repository.PreRegistrationAttemptRepository = (*PreRegistrationAttemptRepository)(nil)
