package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type CryptoTransferRepository struct {
	db dbExecutor
}

func NewCryptoTransferRepository(pool *pgxpool.Pool) *CryptoTransferRepository {
	return &CryptoTransferRepository{db: pool}
}

func NewCryptoTransferRepositoryWithTx(tx dbExecutor) *CryptoTransferRepository {
	return &CryptoTransferRepository{db: tx}
}

func (r *CryptoTransferRepository) Create(ctx context.Context, transfer *entity.CryptoTransfer) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO crypto_transfers (
			id, user_id, account_id, transaction_id, asset, network, address,
			amount, fee, status, direction, tx_hash, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13
		)`,
		transfer.ID,
		transfer.UserID,
		transfer.AccountID,
		transfer.TransactionID,
		transfer.Asset,
		transfer.Network,
		transfer.Address,
		transfer.Amount,
		transfer.Fee,
		transfer.Status,
		transfer.Direction,
		nullIfEmpty(transfer.TxHash),
		transfer.CreatedAt,
	)
	return err
}

func (r *CryptoTransferRepository) GetByTransactionID(ctx context.Context, transactionID string) (*entity.CryptoTransfer, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, account_id, transaction_id, asset, network, address,
		       amount, fee, status, direction, tx_hash, created_at
		  FROM crypto_transfers
		 WHERE transaction_id = $1`, transactionID)
	var transfer entity.CryptoTransfer
	var txHash *string
	if err := row.Scan(
		&transfer.ID,
		&transfer.UserID,
		&transfer.AccountID,
		&transfer.TransactionID,
		&transfer.Asset,
		&transfer.Network,
		&transfer.Address,
		&transfer.Amount,
		&transfer.Fee,
		&transfer.Status,
		&transfer.Direction,
		&txHash,
		&transfer.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if txHash != nil {
		transfer.TxHash = *txHash
	}
	return &transfer, nil
}

func (r *CryptoTransferRepository) UpdateStatus(ctx context.Context, id string, status entity.CryptoTransferStatus, txHash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE crypto_transfers
		   SET status = $1,
		       tx_hash = NULLIF($2, '')
		 WHERE id = $3`, status, txHash, id)
	return err
}

func (r *CryptoTransferRepository) SumConfirmedByUserAsset(ctx context.Context, userID, asset string) (int64, error) {
	row := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(CASE direction WHEN 'BUY' THEN amount WHEN 'SELL' THEN -amount ELSE 0 END), 0)
		  FROM crypto_transfers
		 WHERE user_id = $1 AND asset = $2 AND status = 'CONFIRMED'`, userID, asset)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

var _ repository.CryptoTransferRepository = (*CryptoTransferRepository)(nil)
