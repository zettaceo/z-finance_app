package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
)

type TradeOrderRepository struct {
	pool *pgxpool.Pool
}

func NewTradeOrderRepository(pool *pgxpool.Pool) *TradeOrderRepository {
	return &TradeOrderRepository{pool: pool}
}

func (r *TradeOrderRepository) Create(ctx context.Context, order *entity.TradeOrder) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO trade_orders (
			id, user_id, status, side, base_currency, quote_currency,
			price, quantity, fee, idempotency_key, external_ref,
			debit_account_id, credit_account_id, debit_transaction_id, credit_transaction_id,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17
		)`,
		order.ID,
		order.UserID,
		order.Status,
		order.Side,
		order.BaseCurrency,
		order.QuoteCurrency,
		order.Price,
		order.Quantity,
		order.Fee,
		order.IdempotencyKey,
		nullIfEmpty(order.ExternalRef),
		nullIfEmpty(order.DebitAccountID),
		nullIfEmpty(order.CreditAccountID),
		nullIfEmpty(order.DebitTransaction),
		nullIfEmpty(order.CreditTransaction),
		order.CreatedAt,
		order.UpdatedAt,
	)
	return err
}

func (r *TradeOrderRepository) GetByID(ctx context.Context, id string) (*entity.TradeOrder, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, status, side, base_currency, quote_currency,
		       price, quantity, fee, idempotency_key, external_ref,
		       debit_account_id, credit_account_id, debit_transaction_id, credit_transaction_id,
		       created_at, updated_at
		  FROM trade_orders
		 WHERE id = $1`, id)

	var order entity.TradeOrder
	var externalRef *string
	var debitAccountID *string
	var creditAccountID *string
	var debitTx *string
	var creditTx *string
	if err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Side,
		&order.BaseCurrency,
		&order.QuoteCurrency,
		&order.Price,
		&order.Quantity,
		&order.Fee,
		&order.IdempotencyKey,
		&externalRef,
		&debitAccountID,
		&creditAccountID,
		&debitTx,
		&creditTx,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if externalRef != nil {
		order.ExternalRef = *externalRef
	}
	if debitAccountID != nil {
		order.DebitAccountID = *debitAccountID
	}
	if creditAccountID != nil {
		order.CreditAccountID = *creditAccountID
	}
	if debitTx != nil {
		order.DebitTransaction = *debitTx
	}
	if creditTx != nil {
		order.CreditTransaction = *creditTx
	}

	return &order, nil
}

func (r *TradeOrderRepository) UpdateStatus(ctx context.Context, id string, status entity.TradeStatus) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE trade_orders
		   SET status = $1, updated_at = NOW()
		 WHERE id = $2`, status, id)
	return err
}
