package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type CardAuthorizationRepository struct {
	db dbExecutor
}

func NewCardAuthorizationRepository(pool *pgxpool.Pool) *CardAuthorizationRepository {
	return &CardAuthorizationRepository{db: pool}
}

func NewCardAuthorizationRepositoryWithTx(tx dbExecutor) *CardAuthorizationRepository {
	return &CardAuthorizationRepository{db: tx}
}

func (r *CardAuthorizationRepository) Create(ctx context.Context, auth *entity.CardAuthorization) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO card_authorizations (
			id, user_id, account_id, status, amount, fee, net_amount,
			merchant_name, merchant_mcc, auth_code, external_ref, transaction_id,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14
		)`,
		auth.ID,
		auth.UserID,
		auth.AccountID,
		auth.Status,
		auth.Amount,
		auth.Fee,
		auth.NetAmount,
		nullIfEmpty(auth.MerchantName),
		nullIfEmpty(auth.MerchantMCC),
		nullIfEmpty(auth.AuthCode),
		nullIfEmpty(auth.ExternalRef),
		nullIfEmpty(auth.TransactionID),
		auth.CreatedAt,
		auth.UpdatedAt,
	)
	return err
}

func (r *CardAuthorizationRepository) GetByID(ctx context.Context, id string) (*entity.CardAuthorization, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, account_id, status, amount, fee, net_amount,
		       merchant_name, merchant_mcc, auth_code, external_ref, transaction_id,
		       created_at, updated_at
		  FROM card_authorizations
		 WHERE id = $1`, id)

	var auth entity.CardAuthorization
	var merchantName *string
	var merchantMCC *string
	var authCode *string
	var externalRef *string
	var transactionID *string
	if err := row.Scan(
		&auth.ID,
		&auth.UserID,
		&auth.AccountID,
		&auth.Status,
		&auth.Amount,
		&auth.Fee,
		&auth.NetAmount,
		&merchantName,
		&merchantMCC,
		&authCode,
		&externalRef,
		&transactionID,
		&auth.CreatedAt,
		&auth.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if merchantName != nil {
		auth.MerchantName = *merchantName
	}
	if merchantMCC != nil {
		auth.MerchantMCC = *merchantMCC
	}
	if authCode != nil {
		auth.AuthCode = *authCode
	}
	if externalRef != nil {
		auth.ExternalRef = *externalRef
	}
	if transactionID != nil {
		auth.TransactionID = *transactionID
	}

	return &auth, nil
}

func (r *CardAuthorizationRepository) UpdateStatus(ctx context.Context, id string, status entity.CardAuthStatus) error {
	_, err := r.db.Exec(ctx, `
		UPDATE card_authorizations
		   SET status = $1, updated_at = NOW()
		 WHERE id = $2`, status, id)
	return err
}

var _ repository.CardAuthorizationRepository = (*CardAuthorizationRepository)(nil)

func (r *CardAuthorizationRepository) ListPendingBefore(ctx context.Context, before time.Time, limit int) ([]*entity.CardAuthorization, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, account_id, status, amount, fee, net_amount,
		       merchant_name, merchant_mcc, auth_code, external_ref, transaction_id,
		       created_at, updated_at
		  FROM card_authorizations
		 WHERE status = $1
		   AND updated_at <= $2
		 ORDER BY updated_at ASC
		 LIMIT $3`,
		entity.CardAuthStatusHold,
		before,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.CardAuthorization
	for rows.Next() {
		var auth entity.CardAuthorization
		var merchantName *string
		var merchantMCC *string
		var authCode *string
		var externalRef *string
		var transactionID *string
		if err := rows.Scan(
			&auth.ID,
			&auth.UserID,
			&auth.AccountID,
			&auth.Status,
			&auth.Amount,
			&auth.Fee,
			&auth.NetAmount,
			&merchantName,
			&merchantMCC,
			&authCode,
			&externalRef,
			&transactionID,
			&auth.CreatedAt,
			&auth.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if merchantName != nil {
			auth.MerchantName = *merchantName
		}
		if merchantMCC != nil {
			auth.MerchantMCC = *merchantMCC
		}
		if authCode != nil {
			auth.AuthCode = *authCode
		}
		if externalRef != nil {
			auth.ExternalRef = *externalRef
		}
		if transactionID != nil {
			auth.TransactionID = *transactionID
		}
		items = append(items, &auth)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CardAuthorizationRepository) CountPendingBefore(ctx context.Context, before time.Time) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(1)
		  FROM card_authorizations
		 WHERE status = $1
		   AND updated_at <= $2`,
		entity.CardAuthStatusHold,
		before,
	).Scan(&total)
	return total, err
}
