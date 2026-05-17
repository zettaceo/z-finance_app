package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type InvoiceRepository struct {
	db dbExecutor
}

func NewInvoiceRepository(pool *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{db: pool}
}

func NewInvoiceRepositoryWithTx(tx dbExecutor) *InvoiceRepository {
	return &InvoiceRepository{db: tx}
}

func (r *InvoiceRepository) Create(ctx context.Context, invoice *entity.Invoice) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO invoices (id, user_id, amount_brl, pix_copy_paste, usdt_address, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		invoice.ID,
		invoice.UserID,
		invoice.AmountBRL,
		invoice.PixCopyPaste,
		invoice.USDTAddress,
		nullIfEmpty(invoice.IdempotencyKey),
		invoice.CreatedAt,
	)
	return err
}

func (r *InvoiceRepository) GetByID(ctx context.Context, id string) (*entity.Invoice, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, amount_brl, pix_copy_paste, usdt_address, idempotency_key, created_at
		  FROM invoices
		 WHERE id = $1`, id)
	var invoice entity.Invoice
	if err := row.Scan(
		&invoice.ID,
		&invoice.UserID,
		&invoice.AmountBRL,
		&invoice.PixCopyPaste,
		&invoice.USDTAddress,
		&invoice.IdempotencyKey,
		&invoice.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &invoice, nil
}

func (r *InvoiceRepository) GetByIdempotencyKey(ctx context.Context, userID, key string) (*entity.Invoice, error) {
	if key == "" {
		return nil, repository.ErrNotFound
	}
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, amount_brl, pix_copy_paste, usdt_address, idempotency_key, created_at
		  FROM invoices
		 WHERE user_id = $1
		   AND idempotency_key = $2`, userID, key)
	var invoice entity.Invoice
	if err := row.Scan(
		&invoice.ID,
		&invoice.UserID,
		&invoice.AmountBRL,
		&invoice.PixCopyPaste,
		&invoice.USDTAddress,
		&invoice.IdempotencyKey,
		&invoice.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &invoice, nil
}

var _ repository.InvoiceRepository = (*InvoiceRepository)(nil)
