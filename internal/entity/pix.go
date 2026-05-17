package entity

import "time"

type PixDirection string

const (
	PixDirectionIn  PixDirection = "IN"
	PixDirectionOut PixDirection = "OUT"
)

type PixStatus string

const (
	PixStatusCreated        PixStatus = "CREATED"
	PixStatusPendingPartner PixStatus = "PENDING_PARTNER"
	PixStatusConfirmed      PixStatus = "CONFIRMED"
	PixStatusRejected       PixStatus = "REJECTED"
)

type PixTransfer struct {
	ID             string
	TransactionID  string
	UserID         string
	AccountID      string
	Direction      PixDirection
	Status         PixStatus
	Amount         int64
	Fee            int64
	NetAmount      int64
	IdempotencyKey string
	EndToEndID     string
	ExternalRef    string
	Metadata       map[string]any
	OccurredAt     time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ConfirmedAt    *time.Time
}

type PixKeyType string

const (
	PixKeyTypeCPF   PixKeyType = "CPF"
	PixKeyTypeEmail PixKeyType = "EMAIL"
	PixKeyTypePhone PixKeyType = "PHONE"
	PixKeyTypeEVP   PixKeyType = "EVP"
)

type PixKey struct {
	ID        string
	UserID    string
	AccountID string
	Type      PixKeyType
	Key       string
	CreatedAt time.Time
}
