package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type KYCRepository struct {
	pool *pgxpool.Pool
}

func NewKYCRepository(pool *pgxpool.Pool) *KYCRepository {
	return &KYCRepository{pool: pool}
}

func (r *KYCRepository) Create(ctx context.Context, profile *entity.KYCProfile) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kyc_profiles (user_id, level, status, provider_ref, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		profile.UserID,
		profile.Level,
		profile.Status,
		nullIfEmpty(profile.ProviderRef),
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	return err
}

func (r *KYCRepository) GetByUserID(ctx context.Context, userID string) (*entity.KYCProfile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, level, status, provider_ref, created_at, updated_at
		  FROM kyc_profiles
		 WHERE user_id = $1`, userID)

	var profile entity.KYCProfile
	var providerRef *string
	if err := row.Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Level,
		&profile.Status,
		&providerRef,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	if providerRef != nil {
		profile.ProviderRef = *providerRef
	}
	return &profile, nil
}

func (r *KYCRepository) UpdateLevel(ctx context.Context, userID string, level entity.KYCLevel) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE kyc_profiles
		   SET level = $1, updated_at = NOW()
		 WHERE user_id = $2`, level, userID)
	return err
}

func (r *KYCRepository) UpdateStatus(ctx context.Context, userID string, status entity.KYCStatus) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE kyc_profiles
		   SET status = $1, updated_at = NOW()
		 WHERE user_id = $2`, status, userID)
	return err
}

func (r *KYCRepository) Upsert(ctx context.Context, profile *entity.KYCProfile) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kyc_profiles (user_id, level, status, provider_ref, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET level = EXCLUDED.level,
		    status = EXCLUDED.status,
		    provider_ref = EXCLUDED.provider_ref,
		    updated_at = NOW()`,
		profile.UserID,
		profile.Level,
		profile.Status,
		nullIfEmpty(profile.ProviderRef),
	)
	return err
}

var _ repository.KYCRepository = (*KYCRepository)(nil)
