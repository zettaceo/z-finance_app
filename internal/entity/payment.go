package entity

import "time"

type PaymentStatus string

const (
	PaymentStatusCreated        PaymentStatus = "CREATED"
	PaymentStatusPendingPartner PaymentStatus = "PENDING_PARTNER"
	PaymentStatusConfirmed      PaymentStatus = "CONFIRMED"
	PaymentStatusRejected       PaymentStatus = "REJECTED"
)

type Payment struct {
	ID             string
	UserID         string
	AccountID      string
	Status         PaymentStatus
	Amount         int64
	Fee            int64
	NetAmount      int64
	IdempotencyKey string
	Barcode        string
	ScheduledAt    *time.Time
	DueDate        *time.Time
	ExternalRef    string
	TransactionID  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
