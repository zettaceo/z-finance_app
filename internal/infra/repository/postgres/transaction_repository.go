package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type TransactionRepository struct {
	db dbExecutor
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: pool}
}

func NewTransactionRepositoryWithTx(tx pgx.Tx) *TransactionRepository {
	return &TransactionRepository{db: tx}
}

func (r *TransactionRepository) Create(ctx context.Context, txEntity *entity.Transaction) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO transactions (
			id, account_id, user_id, type, status, amount, fee, net_amount,
			idempotency_key, external_ref, reversal_of_transaction_id, occurred_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`,
		txEntity.ID,
		txEntity.AccountID,
		txEntity.UserID,
		txEntity.Type,
		txEntity.Status,
		txEntity.Amount,
		txEntity.Fee,
		txEntity.NetAmount,
		txEntity.IdempotencyKey,
		txEntity.ExternalRef,
		nullIfEmpty(txEntity.ReversalOf),
		txEntity.OccurredAt,
		txEntity.CreatedAt,
	)
	return err
}

func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*entity.Transaction, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, account_id, user_id, type, status, amount, fee, net_amount,
		       idempotency_key, external_ref, reversal_of_transaction_id, occurred_at, created_at
		  FROM transactions
		 WHERE id = $1`, id)

	var txEntity entity.Transaction
	var reversalOf *string
	err := row.Scan(
		&txEntity.ID,
		&txEntity.AccountID,
		&txEntity.UserID,
		&txEntity.Type,
		&txEntity.Status,
		&txEntity.Amount,
		&txEntity.Fee,
		&txEntity.NetAmount,
		&txEntity.IdempotencyKey,
		&txEntity.ExternalRef,
		&reversalOf,
		&txEntity.OccurredAt,
		&txEntity.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if reversalOf != nil {
		txEntity.ReversalOf = *reversalOf
	}

	return &txEntity, nil
}

func (r *TransactionRepository) GetByIdempotencyKey(ctx context.Context, accountID, key string) (*entity.Transaction, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, account_id, user_id, type, status, amount, fee, net_amount,
		       idempotency_key, external_ref, reversal_of_transaction_id, occurred_at, created_at
		  FROM transactions
		 WHERE account_id = $1 AND idempotency_key = $2`, accountID, key)

	var txEntity entity.Transaction
	var reversalOf *string
	err := row.Scan(
		&txEntity.ID,
		&txEntity.AccountID,
		&txEntity.UserID,
		&txEntity.Type,
		&txEntity.Status,
		&txEntity.Amount,
		&txEntity.Fee,
		&txEntity.NetAmount,
		&txEntity.IdempotencyKey,
		&txEntity.ExternalRef,
		&reversalOf,
		&txEntity.OccurredAt,
		&txEntity.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if reversalOf != nil {
		txEntity.ReversalOf = *reversalOf
	}

	return &txEntity, nil
}

func (r *TransactionRepository) ListByAccount(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	var builder strings.Builder
	builder.WriteString(`
		SELECT id, account_id, user_id, type, status, amount, fee, net_amount,
		       idempotency_key, external_ref, reversal_of_transaction_id, occurred_at, created_at
		  FROM transactions
		 WHERE account_id = $1`)

	args := []any{filter.AccountID}
	argIndex := 2

	if filter.From != nil {
		builder.WriteString(fmt.Sprintf(" AND occurred_at >= $%d", argIndex))
		args = append(args, *filter.From)
		argIndex++
	}
	if filter.To != nil {
		builder.WriteString(fmt.Sprintf(" AND occurred_at <= $%d", argIndex))
		args = append(args, *filter.To)
		argIndex++
	}
	if filter.Type != nil {
		builder.WriteString(fmt.Sprintf(" AND type = $%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}
	if filter.Status != nil {
		builder.WriteString(fmt.Sprintf(" AND status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}
	if filter.CursorAt != nil && filter.CursorID != "" {
		builder.WriteString(fmt.Sprintf(" AND (occurred_at, id) < ($%d, $%d)", argIndex, argIndex+1))
		args = append(args, *filter.CursorAt, filter.CursorID)
		argIndex += 2
	}

	builder.WriteString(fmt.Sprintf(" ORDER BY occurred_at DESC, id DESC LIMIT $%d", argIndex))
	args = append(args, limit)

	rows, err := r.db.Query(ctx, builder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entity.Transaction
	for rows.Next() {
		var txEntity entity.Transaction
		var reversalOf *string
		if err := rows.Scan(
			&txEntity.ID,
			&txEntity.AccountID,
			&txEntity.UserID,
			&txEntity.Type,
			&txEntity.Status,
			&txEntity.Amount,
			&txEntity.Fee,
			&txEntity.NetAmount,
			&txEntity.IdempotencyKey,
			&txEntity.ExternalRef,
			&reversalOf,
			&txEntity.OccurredAt,
			&txEntity.CreatedAt,
		); err != nil {
			return nil, err
		}
		if reversalOf != nil {
			txEntity.ReversalOf = *reversalOf
		}
		result = append(result, &txEntity)
	}

	return result, rows.Err()
}

func (r *TransactionRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*entity.Transaction, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, account_id, user_id, type, status, amount, fee, net_amount,
		       idempotency_key, external_ref, reversal_of_transaction_id, occurred_at, created_at
		  FROM transactions
		 WHERE user_id = $1
		 ORDER BY occurred_at DESC, id DESC
		 LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entity.Transaction
	for rows.Next() {
		var txEntity entity.Transaction
		var reversalOf *string
		if err := rows.Scan(
			&txEntity.ID,
			&txEntity.AccountID,
			&txEntity.UserID,
			&txEntity.Type,
			&txEntity.Status,
			&txEntity.Amount,
			&txEntity.Fee,
			&txEntity.NetAmount,
			&txEntity.IdempotencyKey,
			&txEntity.ExternalRef,
			&reversalOf,
			&txEntity.OccurredAt,
			&txEntity.CreatedAt,
		); err != nil {
			return nil, err
		}
		if reversalOf != nil {
			txEntity.ReversalOf = *reversalOf
		}
		result = append(result, &txEntity)
	}
	return result, rows.Err()
}

func (r *TransactionRepository) ListHoldBefore(ctx context.Context, before time.Time, limit int) ([]*entity.Transaction, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, account_id, user_id, type, status, amount, fee, net_amount,
		       idempotency_key, external_ref, reversal_of_transaction_id, occurred_at, created_at
		  FROM transactions
		 WHERE status IN ($1, $2)
		   AND occurred_at <= $3
		 ORDER BY occurred_at ASC
		 LIMIT $4`,
		entity.TransactionStatusHold,
		entity.TransactionStatusPendingPartner,
		before,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entity.Transaction
	for rows.Next() {
		var txEntity entity.Transaction
		var reversalOf *string
		if err := rows.Scan(
			&txEntity.ID,
			&txEntity.AccountID,
			&txEntity.UserID,
			&txEntity.Type,
			&txEntity.Status,
			&txEntity.Amount,
			&txEntity.Fee,
			&txEntity.NetAmount,
			&txEntity.IdempotencyKey,
			&txEntity.ExternalRef,
			&reversalOf,
			&txEntity.OccurredAt,
			&txEntity.CreatedAt,
		); err != nil {
			return nil, err
		}
		if reversalOf != nil {
			txEntity.ReversalOf = *reversalOf
		}
		result = append(result, &txEntity)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *TransactionRepository) SumOutgoingNetAmountByUserBetween(ctx context.Context, userID string, from, to time.Time) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(net_amount), 0)
		  FROM transactions
		 WHERE user_id = $1
		   AND status IN ($2, $3)
		   AND type IN ($4, $5, $6, $7)
		   AND occurred_at >= $8
		   AND occurred_at < $9`,
		userID,
		entity.TransactionStatusConfirmed,
		entity.TransactionStatusHold,
		entity.TransactionTypeWithdrawal,
		entity.TransactionTypePayment,
		entity.TransactionTypeTradeBuy,
		entity.TransactionTypeCardAuth,
		from,
		to,
	).Scan(&total)
	return total, err
}

func (r *TransactionRepository) CountHoldBefore(ctx context.Context, before time.Time) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(1)
		  FROM transactions
		 WHERE status IN ($1, $2)
		   AND occurred_at <= $3`,
		entity.TransactionStatusHold,
		entity.TransactionStatusPendingPartner,
		before,
	).Scan(&total)
	return total, err
}

func (r *TransactionRepository) GetLedgerBalance(ctx context.Context, accountID string) (int64, error) {
	row := r.db.QueryRow(ctx, ledgerBalanceSQL, accountID)

	var balance int64
	if err := row.Scan(&balance); err != nil {
		return 0, err
	}

	return balance, nil
}

func (r *TransactionRepository) GetHoldBalance(ctx context.Context, accountID string) (int64, error) {
	row := r.db.QueryRow(ctx, holdBalanceSQL, accountID)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *TransactionRepository) UpdateStatusIfCurrent(ctx context.Context, id string, from, to entity.TransactionStatus) (*entity.Transaction, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE transactions
		   SET status = $1
		 WHERE id = $2 AND status = $3
		RETURNING id, account_id, user_id, type, status, amount, fee, net_amount,
		          idempotency_key, external_ref, reversal_of_transaction_id, occurred_at, created_at`,
		to, id, from)

	var txEntity entity.Transaction
	var reversalOf *string
	if err := row.Scan(
		&txEntity.ID,
		&txEntity.AccountID,
		&txEntity.UserID,
		&txEntity.Type,
		&txEntity.Status,
		&txEntity.Amount,
		&txEntity.Fee,
		&txEntity.NetAmount,
		&txEntity.IdempotencyKey,
		&txEntity.ExternalRef,
		&reversalOf,
		&txEntity.OccurredAt,
		&txEntity.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if reversalOf != nil {
		txEntity.ReversalOf = *reversalOf
	}
	return &txEntity, nil
}

const holdBalanceSQL = `
	SELECT COALESCE(SUM(net_amount), 0)
	FROM transactions
	WHERE account_id = $1
	  AND status IN ('HOLD', 'PENDING_PARTNER')
	  AND type IN ('WITHDRAWAL', 'PAYMENT', 'TRADE_BUY', 'CARD_AUTH')`

const ledgerBalanceSQL = `
	SELECT COALESCE(SUM(
		CASE t.type
			WHEN 'DEPOSIT' THEN t.net_amount
			WHEN 'WITHDRAWAL' THEN -t.net_amount
			WHEN 'PAYMENT' THEN -t.net_amount
			WHEN 'TRADE_BUY' THEN -t.net_amount
			WHEN 'TRADE_SELL' THEN t.net_amount
			WHEN 'CARD_AUTH' THEN -t.net_amount
			WHEN 'REVERSAL' THEN CASE WHEN rev.status = 'CONFIRMED' THEN t.net_amount ELSE 0 END
			ELSE 0
		END
	), 0)
	FROM transactions t
	LEFT JOIN transactions rev
	       ON t.reversal_of_transaction_id = rev.id
	WHERE t.account_id = $1 AND t.status = 'CONFIRMED'`
