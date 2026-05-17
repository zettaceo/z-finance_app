package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type AccountRepository struct {
	db dbExecutor
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: pool}
}

func NewAccountRepositoryWithTx(tx pgx.Tx) *AccountRepository {
	return &AccountRepository{db: tx}
}

func (r *AccountRepository) Create(ctx context.Context, account *entity.Account) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO accounts (id, user_id, currency, scale, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		account.ID,
		account.UserID,
		account.Currency,
		account.Scale,
		account.Status,
		account.CreatedAt,
	)
	return err
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*entity.Account, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, currency, scale, status, created_at
		  FROM accounts
		 WHERE id = $1`, id)

	var account entity.Account
	if err := row.Scan(
		&account.ID,
		&account.UserID,
		&account.Currency,
		&account.Scale,
		&account.Status,
		&account.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepository) GetByIDForUpdate(ctx context.Context, id string) (*entity.Account, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, currency, scale, status, created_at
		  FROM accounts
		 WHERE id = $1 FOR UPDATE`, id)

	var account entity.Account
	if err := row.Scan(
		&account.ID,
		&account.UserID,
		&account.Currency,
		&account.Scale,
		&account.Status,
		&account.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepository) ListByUser(ctx context.Context, userID string) ([]*entity.Account, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, currency, scale, status, created_at
		  FROM accounts
		 WHERE user_id = $1
		 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*entity.Account
	for rows.Next() {
		var account entity.Account
		if err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Currency,
			&account.Scale,
			&account.Status,
			&account.CreatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, &account)
	}

	return accounts, rows.Err()
}

func (r *AccountRepository) UpdateStatus(ctx context.Context, id string, status entity.AccountStatus) error {
	_, err := r.db.Exec(ctx, `
		UPDATE accounts
		   SET status = $1
		 WHERE id = $2`, status, id)
	return err
}

var _ repository.AccountRepository = (*AccountRepository)(nil)
