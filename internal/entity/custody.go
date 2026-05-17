package entity

import "time"

type CustodyTransferStatus string

const (
	CustodyTransferPending   CustodyTransferStatus = "PENDING"
	CustodyTransferConfirmed CustodyTransferStatus = "CONFIRMED"
	CustodyTransferFailed    CustodyTransferStatus = "FAILED"
)

type CustodyAddress struct {
	ProviderID string
	Network    string
	Asset      string
	Address    string
	Tag        string
	CreatedAt  time.Time
}

type CustodyTransfer struct {
	ProviderID  string
	UserID      string
	Network     string
	Asset       string
	Address     string
	Amount      int64
	Status      CustodyTransferStatus
	ExternalRef string
	CreatedAt   time.Time
}
