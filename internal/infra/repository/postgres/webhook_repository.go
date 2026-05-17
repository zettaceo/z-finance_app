package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type WebhookRepository struct {
	pool *pgxpool.Pool
}

func NewWebhookRepository(pool *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{pool: pool}
}

func (r *WebhookRepository) ConfirmPayment(ctx context.Context, paymentID, transactionID string) error {
	return r.updatePaymentAndTransaction(ctx, paymentID, transactionID, entity.PaymentStatusConfirmed, entity.TransactionStatusConfirmed)
}

func (r *WebhookRepository) RejectPayment(ctx context.Context, paymentID, transactionID string) error {
	return r.updatePaymentAndTransaction(ctx, paymentID, transactionID, entity.PaymentStatusRejected, entity.TransactionStatusRejected)
}

func (r *WebhookRepository) ConfirmCardAuthorization(ctx context.Context, authorizationID, transactionID string) error {
	return r.updateCardAndTransaction(ctx, authorizationID, transactionID, entity.CardAuthStatusConfirmed, entity.TransactionStatusConfirmed)
}

func (r *WebhookRepository) RejectCardAuthorization(ctx context.Context, authorizationID, transactionID string) error {
	return r.updateCardAndTransaction(ctx, authorizationID, transactionID, entity.CardAuthStatusRejected, entity.TransactionStatusRejected)
}

func (r *WebhookRepository) EnsureEvent(ctx context.Context, eventType, referenceID string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO webhook_events (event_type, reference_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (event_type, reference_id) DO NOTHING`,
		eventType, referenceID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 0, nil
}

func (r *WebhookRepository) updatePaymentAndTransaction(ctx context.Context, paymentID, transactionID string, paymentStatus entity.PaymentStatus, txStatus entity.TransactionStatus) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	currentStatus, err := r.getPaymentStatus(ctx, tx, paymentID)
	if err != nil {
		return err
	}
	if currentStatus == entity.PaymentStatusRejected && paymentStatus == entity.PaymentStatusConfirmed {
		return repository.ErrInvalidState
	}
	if currentStatus == entity.PaymentStatusConfirmed && paymentStatus == entity.PaymentStatusRejected {
		return repository.ErrInvalidState
	}

	if currentStatus != paymentStatus {
		_, err = tx.Exec(ctx, `
			UPDATE payments
			   SET status = $1, updated_at = NOW()
			 WHERE id = $2`, paymentStatus, paymentID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE transactions
		   SET status = $1
		 WHERE id = $2 AND status = 'HOLD'`, txStatus, transactionID)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	return err
}

func (r *WebhookRepository) updateCardAndTransaction(ctx context.Context, authorizationID, transactionID string, authStatus entity.CardAuthStatus, txStatus entity.TransactionStatus) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	currentStatus, err := r.getCardStatus(ctx, tx, authorizationID)
	if err != nil {
		return err
	}
	if currentStatus == entity.CardAuthStatusRejected && authStatus == entity.CardAuthStatusConfirmed {
		return repository.ErrInvalidState
	}
	if currentStatus == entity.CardAuthStatusConfirmed && authStatus == entity.CardAuthStatusRejected {
		return repository.ErrInvalidState
	}

	if currentStatus != authStatus {
		_, err = tx.Exec(ctx, `
			UPDATE card_authorizations
			   SET status = $1, updated_at = NOW()
			 WHERE id = $2`, authStatus, authorizationID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE transactions
		   SET status = $1
		 WHERE id = $2 AND status = 'HOLD'`, txStatus, transactionID)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	return err
}

func (r *WebhookRepository) getPaymentStatus(ctx context.Context, tx pgx.Tx, paymentID string) (entity.PaymentStatus, error) {
	var status entity.PaymentStatus
	if err := tx.QueryRow(ctx, "SELECT status FROM payments WHERE id = $1", paymentID).Scan(&status); err != nil {
		if err == pgx.ErrNoRows {
			return "", repository.ErrNotFound
		}
		return "", err
	}
	return status, nil
}

func (r *WebhookRepository) getCardStatus(ctx context.Context, tx pgx.Tx, authorizationID string) (entity.CardAuthStatus, error) {
	var status entity.CardAuthStatus
	if err := tx.QueryRow(ctx, "SELECT status FROM card_authorizations WHERE id = $1", authorizationID).Scan(&status); err != nil {
		if err == pgx.ErrNoRows {
			return "", repository.ErrNotFound
		}
		return "", err
	}
	return status, nil
}

var _ repository.WebhookRepository = (*WebhookRepository)(nil)
