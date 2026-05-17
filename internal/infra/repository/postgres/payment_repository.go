package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO payments (
			id, user_id, account_id, status, amount, fee, net_amount, idempotency_key,
			barcode, scheduled_at, due_date, external_ref, transaction_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14, $15
		)`,
		payment.ID,
		payment.UserID,
		payment.AccountID,
		payment.Status,
		payment.Amount,
		payment.Fee,
		payment.NetAmount,
		payment.IdempotencyKey,
		nullIfEmpty(payment.Barcode),
		payment.ScheduledAt,
		payment.DueDate,
		nullIfEmpty(payment.ExternalRef),
		nullIfEmpty(payment.TransactionID),
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	return err
}

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*entity.Payment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, account_id, status, amount, fee, net_amount, idempotency_key,
		       barcode, scheduled_at, due_date, external_ref, transaction_id, created_at, updated_at
		  FROM payments
		 WHERE id = $1`, id)

	var payment entity.Payment
	var barcode *string
	var externalRef *string
	var transactionID *string
	if err := row.Scan(
		&payment.ID,
		&payment.UserID,
		&payment.AccountID,
		&payment.Status,
		&payment.Amount,
		&payment.Fee,
		&payment.NetAmount,
		&payment.IdempotencyKey,
		&barcode,
		&payment.ScheduledAt,
		&payment.DueDate,
		&externalRef,
		&transactionID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if barcode != nil {
		payment.Barcode = *barcode
	}
	if externalRef != nil {
		payment.ExternalRef = *externalRef
	}
	if transactionID != nil {
		payment.TransactionID = *transactionID
	}

	return &payment, nil
}

func (r *PaymentRepository) GetByIdempotencyKey(ctx context.Context, accountID, key string) (*entity.Payment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, account_id, status, amount, fee, net_amount, idempotency_key,
		       barcode, scheduled_at, due_date, external_ref, transaction_id, created_at, updated_at
		  FROM payments
		 WHERE account_id = $1 AND idempotency_key = $2`, accountID, key)

	var payment entity.Payment
	var barcode *string
	var externalRef *string
	var transactionID *string
	if err := row.Scan(
		&payment.ID,
		&payment.UserID,
		&payment.AccountID,
		&payment.Status,
		&payment.Amount,
		&payment.Fee,
		&payment.NetAmount,
		&payment.IdempotencyKey,
		&barcode,
		&payment.ScheduledAt,
		&payment.DueDate,
		&externalRef,
		&transactionID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if barcode != nil {
		payment.Barcode = *barcode
	}
	if externalRef != nil {
		payment.ExternalRef = *externalRef
	}
	if transactionID != nil {
		payment.TransactionID = *transactionID
	}

	return &payment, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id string, status entity.PaymentStatus) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE payments
		   SET status = $1, updated_at = NOW()
		 WHERE id = $2`, status, id)
	return err
}

func (r *PaymentRepository) ListPendingBefore(ctx context.Context, before time.Time, limit int) ([]*entity.Payment, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, account_id, status, amount, fee, net_amount, idempotency_key,
		       barcode, scheduled_at, due_date, external_ref, transaction_id, created_at, updated_at
		  FROM payments
		 WHERE status IN ($1, $2)
		   AND updated_at <= $3
		 ORDER BY updated_at ASC
		 LIMIT $4`,
		entity.PaymentStatusCreated,
		entity.PaymentStatusPendingPartner,
		before,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.Payment
	for rows.Next() {
		var payment entity.Payment
		var barcode *string
		var externalRef *string
		var transactionID *string
		if err := rows.Scan(
			&payment.ID,
			&payment.UserID,
			&payment.AccountID,
			&payment.Status,
			&payment.Amount,
			&payment.Fee,
			&payment.NetAmount,
			&payment.IdempotencyKey,
			&barcode,
			&payment.ScheduledAt,
			&payment.DueDate,
			&externalRef,
			&transactionID,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if barcode != nil {
			payment.Barcode = *barcode
		}
		if externalRef != nil {
			payment.ExternalRef = *externalRef
		}
		if transactionID != nil {
			payment.TransactionID = *transactionID
		}
		items = append(items, &payment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PaymentRepository) CountPendingBefore(ctx context.Context, before time.Time) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		  FROM payments
		 WHERE status IN ($1, $2)
		   AND updated_at <= $3`,
		entity.PaymentStatusCreated,
		entity.PaymentStatusPendingPartner,
		before,
	).Scan(&total)
	return total, err
}

func (r *PaymentRepository) CountScheduledByUserBetween(ctx context.Context, userID string, from, to time.Time) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		  FROM payments
		 WHERE user_id = $1
		   AND scheduled_at IS NOT NULL
		   AND created_at >= $2
		   AND created_at < $3`,
		userID,
		from,
		to,
	).Scan(&total)
	return total, err
}
