package entity

import "time"

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "ACTIVE"
	AccountStatusBlocked AccountStatus = "BLOCKED"
	AccountStatusClosed  AccountStatus = "CLOSED"
)

type Account struct {
	ID        string
	UserID    string
	Currency  string
	Scale     int32
	Status    AccountStatus
	CreatedAt time.Time
}
