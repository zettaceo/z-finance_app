package entity

import "time"

type CardAuthStatus string

const (
	CardAuthStatusHold      CardAuthStatus = "HOLD"
	CardAuthStatusConfirmed CardAuthStatus = "CONFIRMED"
	CardAuthStatusRejected  CardAuthStatus = "REJECTED"
)

type CardAuthorization struct {
	ID            string
	UserID        string
	AccountID     string
	Status        CardAuthStatus
	Amount        int64
	Fee           int64
	NetAmount     int64
	MerchantName  string
	MerchantMCC   string
	AuthCode      string
	ExternalRef   string
	TransactionID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
