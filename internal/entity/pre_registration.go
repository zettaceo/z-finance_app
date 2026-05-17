package entity

import "time"

type PreRegistrationStatus string

const (
	PreRegistrationPending  PreRegistrationStatus = "PENDING"
	PreRegistrationVerified PreRegistrationStatus = "VERIFIED"
	PreRegistrationInvited  PreRegistrationStatus = "INVITED"
	PreRegistrationConverted PreRegistrationStatus = "CONVERTED"
	PreRegistrationExpired  PreRegistrationStatus = "EXPIRED"
)

type VerificationStatus string

const (
	VerificationPending VerificationStatus = "PENDING"
	VerificationVerified VerificationStatus = "VERIFIED"
	VerificationBlocked VerificationStatus = "BLOCKED"
	VerificationExpired VerificationStatus = "EXPIRED"
)

type PreRegistration struct {
	ID               string
	FullName         string
	Email            string
	Phone            string
	Status           PreRegistrationStatus
	EmailStatus       VerificationStatus
	PhoneStatus       VerificationStatus
	EmailTokenHash    string
	EmailTokenExpiresAt *time.Time
	PhoneCodeHash     string
	PhoneCodeExpiresAt *time.Time
	EmailVerifiedAt   *time.Time
	PhoneVerifiedAt   *time.Time
	ExpiresAt         time.Time
	EmailAttempts     int
	PhoneAttempts     int
	EmailBlockedUntil *time.Time
	PhoneBlockedUntil *time.Time
	CreatedIP         string
	UserAgent         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PreRegistrationAttempt struct {
	ID                string
	PreRegistrationID string
	Channel           string
	Success           bool
	Reason            string
	IP                string
	CreatedAt         time.Time
}
