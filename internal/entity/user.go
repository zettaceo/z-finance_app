package entity

import "time"

type UserStatus string

const (
	UserStatusPending UserStatus = "PENDING"
	UserStatusActive  UserStatus = "ACTIVE"
	UserStatusBlocked UserStatus = "BLOCKED"
)

type UserType string

const (
	UserTypePF UserType = "PF"
	UserTypePJ UserType = "PJ"
)

type User struct {
	ID         string
	ExternalID string
	Email      string
	FullName   string
	Status     UserStatus
	UserType   UserType
	PasswordHash string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
