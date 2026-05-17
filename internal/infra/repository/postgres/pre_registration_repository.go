package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PreRegistrationRepository struct {
	db dbExecutor
}

func NewPreRegistrationRepository(pool *pgxpool.Pool) *PreRegistrationRepository {
	return &PreRegistrationRepository{db: pool}
}

func NewPreRegistrationRepositoryWithTx(tx dbExecutor) *PreRegistrationRepository {
	return &PreRegistrationRepository{db: tx}
}

func (r *PreRegistrationRepository) Create(ctx context.Context, item *entity.PreRegistration) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO pre_registrations (
			id, full_name, email, phone, status, email_status, phone_status,
			email_token_hash, phone_code_hash, email_verified_at, phone_verified_at,
			expires_at, email_attempts, phone_attempts, email_blocked_until, phone_blocked_until,
			created_ip, user_agent, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19, $20
		)`,
		item.ID,
		item.FullName,
		item.Email,
		item.Phone,
		item.Status,
		item.EmailStatus,
		item.PhoneStatus,
		item.EmailTokenHash,
		item.PhoneCodeHash,
		item.EmailVerifiedAt,
		item.PhoneVerifiedAt,
		item.ExpiresAt,
		item.EmailAttempts,
		item.PhoneAttempts,
		item.EmailBlockedUntil,
		item.PhoneBlockedUntil,
		nullIfEmpty(item.CreatedIP),
		nullIfEmpty(item.UserAgent),
		item.CreatedAt,
		item.UpdatedAt,
	)
	return err
}

func (r *PreRegistrationRepository) Update(ctx context.Context, item *entity.PreRegistration) error {
	_, err := r.db.Exec(ctx, `
		UPDATE pre_registrations
		   SET full_name = $2,
		       email = $3,
		       phone = $4,
		       status = $5,
		       email_status = $6,
		       phone_status = $7,
		       email_token_hash = $8,
		       phone_code_hash = $9,
		       email_verified_at = $10,
		       phone_verified_at = $11,
		       expires_at = $12,
		       email_attempts = $13,
		       phone_attempts = $14,
		       email_blocked_until = $15,
		       phone_blocked_until = $16,
		       created_ip = $17,
		       user_agent = $18,
		       updated_at = $19
		 WHERE id = $1`,
		item.ID,
		item.FullName,
		item.Email,
		item.Phone,
		item.Status,
		item.EmailStatus,
		item.PhoneStatus,
		item.EmailTokenHash,
		item.PhoneCodeHash,
		item.EmailVerifiedAt,
		item.PhoneVerifiedAt,
		item.ExpiresAt,
		item.EmailAttempts,
		item.PhoneAttempts,
		item.EmailBlockedUntil,
		item.PhoneBlockedUntil,
		nullIfEmpty(item.CreatedIP),
		nullIfEmpty(item.UserAgent),
		item.UpdatedAt,
	)
	return err
}

func (r *PreRegistrationRepository) GetByID(ctx context.Context, id string) (*entity.PreRegistration, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, full_name, email, phone, status, email_status, phone_status,
		       email_token_hash, phone_code_hash, email_verified_at, phone_verified_at,
		       expires_at, email_attempts, phone_attempts, email_blocked_until, phone_blocked_until,
		       created_ip, user_agent, created_at, updated_at
		  FROM pre_registrations
		 WHERE id = $1`, id)
	return scanPreRegistration(row)
}

func (r *PreRegistrationRepository) GetByEmail(ctx context.Context, email string) (*entity.PreRegistration, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, full_name, email, phone, status, email_status, phone_status,
		       email_token_hash, phone_code_hash, email_verified_at, phone_verified_at,
		       expires_at, email_attempts, phone_attempts, email_blocked_until, phone_blocked_until,
		       created_ip, user_agent, created_at, updated_at
		  FROM pre_registrations
		 WHERE LOWER(email) = LOWER($1)`, email)
	return scanPreRegistration(row)
}

func (r *PreRegistrationRepository) GetByPhone(ctx context.Context, phone string) (*entity.PreRegistration, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, full_name, email, phone, status, email_status, phone_status,
		       email_token_hash, phone_code_hash, email_verified_at, phone_verified_at,
		       expires_at, email_attempts, phone_attempts, email_blocked_until, phone_blocked_until,
		       created_ip, user_agent, created_at, updated_at
		  FROM pre_registrations
		 WHERE phone = $1`, phone)
	return scanPreRegistration(row)
}

func scanPreRegistration(row pgx.Row) (*entity.PreRegistration, error) {
	var item entity.PreRegistration
	var emailVerifiedAt *time.Time
	var phoneVerifiedAt *time.Time
	var emailBlockedUntil *time.Time
	var phoneBlockedUntil *time.Time
	var createdIP *string
	var userAgent *string
	if err := row.Scan(
		&item.ID,
		&item.FullName,
		&item.Email,
		&item.Phone,
		&item.Status,
		&item.EmailStatus,
		&item.PhoneStatus,
		&item.EmailTokenHash,
		&item.PhoneCodeHash,
		&emailVerifiedAt,
		&phoneVerifiedAt,
		&item.ExpiresAt,
		&item.EmailAttempts,
		&item.PhoneAttempts,
		&emailBlockedUntil,
		&phoneBlockedUntil,
		&createdIP,
		&userAgent,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	item.EmailVerifiedAt = emailVerifiedAt
	item.PhoneVerifiedAt = phoneVerifiedAt
	item.EmailBlockedUntil = emailBlockedUntil
	item.PhoneBlockedUntil = phoneBlockedUntil
	if createdIP != nil {
		item.CreatedIP = *createdIP
	}
	if userAgent != nil {
		item.UserAgent = *userAgent
	}
	return &item, nil
}

var _ repository.PreRegistrationRepository = (*PreRegistrationRepository)(nil)
