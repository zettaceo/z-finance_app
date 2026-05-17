package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PixRepository struct {
	db dbExecutor
}

func NewPixRepository(pool *pgxpool.Pool) *PixRepository {
	return &PixRepository{db: pool}
}

func NewPixRepositoryWithTx(tx dbExecutor) *PixRepository {
	return &PixRepository{db: tx}
}

func (r *PixRepository) Create(ctx context.Context, transfer *entity.PixTransfer) error {
	var metadata []byte
	if transfer.Metadata != nil {
		data, err := json.Marshal(transfer.Metadata)
		if err != nil {
			return err
		}
		metadata = data
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO pix_transfers (
			id, transaction_id, user_id, account_id, direction, status,
			amount, fee, net_amount, idempotency_key, end_to_end_id, external_ref, metadata,
			occurred_at, created_at, updated_at, confirmed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17
		)`,
		transfer.ID,
		nullIfEmpty(transfer.TransactionID),
		transfer.UserID,
		transfer.AccountID,
		transfer.Direction,
		transfer.Status,
		transfer.Amount,
		transfer.Fee,
		transfer.NetAmount,
		transfer.IdempotencyKey,
		nullIfEmpty(transfer.EndToEndID),
		nullIfEmpty(transfer.ExternalRef),
		metadata,
		transfer.OccurredAt,
		transfer.CreatedAt,
		transfer.UpdatedAt,
		transfer.ConfirmedAt,
	)
	return err
}

func (r *PixRepository) GetByID(ctx context.Context, id string) (*entity.PixTransfer, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, transaction_id, user_id, account_id, direction, status,
		       amount, fee, net_amount, idempotency_key, end_to_end_id, external_ref, metadata,
		       occurred_at, created_at, updated_at, confirmed_at
		  FROM pix_transfers
		 WHERE id = $1`, id)

	var transfer entity.PixTransfer
	var transactionID *string
	var endToEndID *string
	var externalRef *string
	var metadata []byte
	var confirmedAt *time.Time
	if err := row.Scan(
		&transfer.ID,
		&transactionID,
		&transfer.UserID,
		&transfer.AccountID,
		&transfer.Direction,
		&transfer.Status,
		&transfer.Amount,
		&transfer.Fee,
		&transfer.NetAmount,
		&transfer.IdempotencyKey,
		&endToEndID,
		&externalRef,
		&metadata,
		&transfer.OccurredAt,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
		&confirmedAt,
	); err != nil {
		return nil, err
	}

	if transactionID != nil {
		transfer.TransactionID = *transactionID
	}
	if endToEndID != nil {
		transfer.EndToEndID = *endToEndID
	}
	if externalRef != nil {
		transfer.ExternalRef = *externalRef
	}
	if len(metadata) > 0 {
		var decoded map[string]any
		if err := json.Unmarshal(metadata, &decoded); err == nil {
			transfer.Metadata = decoded
		}
	}
	transfer.ConfirmedAt = confirmedAt

	return &transfer, nil
}

func (r *PixRepository) GetByIdempotencyKey(ctx context.Context, accountID, key string) (*entity.PixTransfer, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, transaction_id, user_id, account_id, direction, status,
		       amount, fee, net_amount, idempotency_key, end_to_end_id, external_ref, metadata,
		       occurred_at, created_at, updated_at, confirmed_at
		  FROM pix_transfers
		 WHERE account_id = $1 AND idempotency_key = $2`, accountID, key)

	var transfer entity.PixTransfer
	var transactionID *string
	var endToEndID *string
	var externalRef *string
	var metadata []byte
	var confirmedAt *time.Time
	if err := row.Scan(
		&transfer.ID,
		&transactionID,
		&transfer.UserID,
		&transfer.AccountID,
		&transfer.Direction,
		&transfer.Status,
		&transfer.Amount,
		&transfer.Fee,
		&transfer.NetAmount,
		&transfer.IdempotencyKey,
		&endToEndID,
		&externalRef,
		&metadata,
		&transfer.OccurredAt,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
		&confirmedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if transactionID != nil {
		transfer.TransactionID = *transactionID
	}
	if endToEndID != nil {
		transfer.EndToEndID = *endToEndID
	}
	if externalRef != nil {
		transfer.ExternalRef = *externalRef
	}
	if len(metadata) > 0 {
		var decoded map[string]any
		if err := json.Unmarshal(metadata, &decoded); err == nil {
			transfer.Metadata = decoded
		}
	}
	transfer.ConfirmedAt = confirmedAt

	return &transfer, nil
}

func (r *PixRepository) GetByEndToEndID(ctx context.Context, endToEndID string) (*entity.PixTransfer, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, transaction_id, user_id, account_id, direction, status,
		       amount, fee, net_amount, idempotency_key, end_to_end_id, external_ref, metadata,
		       occurred_at, created_at, updated_at, confirmed_at
		  FROM pix_transfers
		 WHERE end_to_end_id = $1`, endToEndID)

	var transfer entity.PixTransfer
	var transactionID *string
	var endToEndIDPtr *string
	var externalRef *string
	var metadata []byte
	var confirmedAt *time.Time
	if err := row.Scan(
		&transfer.ID,
		&transactionID,
		&transfer.UserID,
		&transfer.AccountID,
		&transfer.Direction,
		&transfer.Status,
		&transfer.Amount,
		&transfer.Fee,
		&transfer.NetAmount,
		&transfer.IdempotencyKey,
		&endToEndIDPtr,
		&externalRef,
		&metadata,
		&transfer.OccurredAt,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
		&confirmedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if transactionID != nil {
		transfer.TransactionID = *transactionID
	}
	if endToEndIDPtr != nil {
		transfer.EndToEndID = *endToEndIDPtr
	}
	if externalRef != nil {
		transfer.ExternalRef = *externalRef
	}
	if len(metadata) > 0 {
		var decoded map[string]any
		if err := json.Unmarshal(metadata, &decoded); err == nil {
			transfer.Metadata = decoded
		}
	}
	transfer.ConfirmedAt = confirmedAt

	return &transfer, nil
}

func (r *PixRepository) UpdateStatus(ctx context.Context, id string, status entity.PixStatus) error {
	_, err := r.db.Exec(ctx, `
		UPDATE pix_transfers
		   SET status = $1, updated_at = NOW()
		 WHERE id = $2`, status, id)
	return err
}

func (r *PixRepository) UpdateStatusWithConfirmation(ctx context.Context, id string, status entity.PixStatus, confirmedAt *time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE pix_transfers
		   SET status = $1,
		       confirmed_at = $2,
		       updated_at = NOW()
		 WHERE id = $3`, status, confirmedAt, id)
	return err
}

func (r *PixRepository) UpdateTransactionID(ctx context.Context, id string, transactionID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE pix_transfers
		   SET transaction_id = $1,
		       updated_at = NOW()
		 WHERE id = $2`, transactionID, id)
	return err
}

var _ repository.PixTransferRepository = (*PixRepository)(nil)

func (r *PixRepository) ListPendingBefore(ctx context.Context, before time.Time, limit int) ([]*entity.PixTransfer, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, transaction_id, user_id, account_id, direction, status,
		       amount, fee, net_amount, idempotency_key, end_to_end_id, external_ref, metadata,
		       occurred_at, created_at, updated_at, confirmed_at
		  FROM pix_transfers
		 WHERE status IN ($1, $2)
		   AND updated_at <= $3
		 ORDER BY updated_at ASC
		 LIMIT $4`,
		entity.PixStatusCreated,
		entity.PixStatusPendingPartner,
		before,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.PixTransfer
	for rows.Next() {
		var transfer entity.PixTransfer
		var transactionID *string
		var endToEndID *string
		var externalRef *string
		var metadata []byte
		var confirmedAt *time.Time
		if err := rows.Scan(
			&transfer.ID,
			&transactionID,
			&transfer.UserID,
			&transfer.AccountID,
			&transfer.Direction,
			&transfer.Status,
			&transfer.Amount,
			&transfer.Fee,
			&transfer.NetAmount,
			&transfer.IdempotencyKey,
			&endToEndID,
			&externalRef,
			&metadata,
			&transfer.OccurredAt,
			&transfer.CreatedAt,
			&transfer.UpdatedAt,
			&confirmedAt,
		); err != nil {
			return nil, err
		}
		if transactionID != nil {
			transfer.TransactionID = *transactionID
		}
		if endToEndID != nil {
			transfer.EndToEndID = *endToEndID
		}
		if externalRef != nil {
			transfer.ExternalRef = *externalRef
		}
		if len(metadata) > 0 {
			var decoded map[string]any
			if err := json.Unmarshal(metadata, &decoded); err == nil {
				transfer.Metadata = decoded
			}
		}
		transfer.ConfirmedAt = confirmedAt
		items = append(items, &transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PixRepository) CountPendingBefore(ctx context.Context, before time.Time) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(1)
		  FROM pix_transfers
		 WHERE status IN ($1, $2)
		   AND updated_at <= $3`,
		entity.PixStatusCreated,
		entity.PixStatusPendingPartner,
		before,
	).Scan(&total)
	return total, err
}
