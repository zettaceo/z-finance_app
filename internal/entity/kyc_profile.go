package entity

import "time"

type KYCLevel string

const (
	KYCLevelUnverified KYCLevel = "UNVERIFIED"
	KYCLevelBasic      KYCLevel = "BASIC"
	KYCLevelFull       KYCLevel = "FULL"
)

type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "PENDING"
	KYCStatusVerified KYCStatus = "VERIFIED"
	KYCStatusRejected KYCStatus = "REJECTED"
)

type KYCProfile struct {
	ID          string
	UserID      string
	Level       KYCLevel
	Status      KYCStatus
	ProviderRef string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
