package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PixKeyRepository struct {
	db dbExecutor
}

func NewPixKeyRepository(pool *pgxpool.Pool) *PixKeyRepository {
	return &PixKeyRepository{db: pool}
}

func NewPixKeyRepositoryWithTx(tx dbExecutor) *PixKeyRepository {
	return &PixKeyRepository{db: tx}
}

func (r *PixKeyRepository) Create(ctx context.Context, key *entity.PixKey) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO pix_keys (id, user_id, account_id, type, key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		key.ID,
		key.UserID,
		key.AccountID,
		key.Type,
		key.Key,
		key.CreatedAt,
	)
	return err
}

func (r *PixKeyRepository) GetByKey(ctx context.Context, value string) (*entity.PixKey, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, account_id, type, key, created_at
		  FROM pix_keys
		 WHERE key = $1`, value)
	var pixKey entity.PixKey
	if err := row.Scan(
		&pixKey.ID,
		&pixKey.UserID,
		&pixKey.AccountID,
		&pixKey.Type,
		&pixKey.Key,
		&pixKey.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &pixKey, nil
}

func (r *PixKeyRepository) GetByID(ctx context.Context, id string) (*entity.PixKey, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, account_id, type, key, created_at
		  FROM pix_keys
		 WHERE id = $1`, id)
	var pixKey entity.PixKey
	if err := row.Scan(
		&pixKey.ID,
		&pixKey.UserID,
		&pixKey.AccountID,
		&pixKey.Type,
		&pixKey.Key,
		&pixKey.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &pixKey, nil
}

var _ repository.PixKeyRepository = (*PixKeyRepository)(nil)
