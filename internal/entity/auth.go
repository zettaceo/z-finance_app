package entity

import "time"

type RefreshToken struct {
	ID         string
	UserID     string
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy string
	CreatedAt  time.Time
}

type LoginAudit struct {
	ID        string
	UserID    string
	Email     string
	IP        string
	Success   bool
	Reason    string
	CreatedAt time.Time
}
