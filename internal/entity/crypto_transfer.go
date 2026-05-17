package entity

import "time"

type CryptoTransferStatus string

const (
	CryptoTransferPendingExchange CryptoTransferStatus = "PENDING_EXCHANGE"
	CryptoTransferConfirmed       CryptoTransferStatus = "CONFIRMED"
	CryptoTransferRejected        CryptoTransferStatus = "REJECTED"
)

type CryptoTransfer struct {
	ID        string
	UserID    string
	AccountID string
	TransactionID string
	Asset     string
	Network   string
	Address   string
	Amount    int64
	Fee       int64
	Status    CryptoTransferStatus
	Direction string
	TxHash    string
	CreatedAt time.Time
}
