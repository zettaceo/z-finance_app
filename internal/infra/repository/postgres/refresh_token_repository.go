package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type RefreshTokenRepository struct {
	db dbExecutor
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: pool}
}

func NewRefreshTokenRepositoryWithTx(tx dbExecutor) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: tx}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, expires_at, revoked_at, replaced_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)`,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.RevokedAt,
		nullIfEmpty(token.ReplacedBy),
		token.CreatedAt,
	)
	return err
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*entity.RefreshToken, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, replaced_by, created_at
		  FROM refresh_tokens
		 WHERE token_hash = $1`, hash)
	var token entity.RefreshToken
	var revokedAt *time.Time
	var replacedBy *string
	if err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&revokedAt,
		&replacedBy,
		&token.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	token.RevokedAt = revokedAt
	if replacedBy != nil {
		token.ReplacedBy = *replacedBy
	}
	return &token, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string, replacedBy string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE refresh_tokens
		   SET revoked_at = NOW(),
		       replaced_by = NULLIF($2, '')::uuid
		 WHERE id = $1`, id, replacedBy)
	return err
}

func (r *RefreshTokenRepository) RevokeAllByUser(ctx context.Context, userID string) (int64, error) {
	cmd, err := r.db.Exec(ctx, `
		UPDATE refresh_tokens
		   SET revoked_at = NOW()
		 WHERE user_id = $1
		   AND revoked_at IS NULL`, userID)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

var _ repository.RefreshTokenRepository = (*RefreshTokenRepository)(nil)
