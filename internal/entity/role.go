package entity

import "time"

type Role struct {
	Code        string
	Description string
	CreatedAt   time.Time
}

type UserRole struct {
	UserID    string
	RoleCode  string
	GrantedBy string
	GrantedAt time.Time
}

type RoleSeparationRule struct {
	ID        string
	RoleCodeA string
	RoleCodeB string
	Reason    string
	CreatedAt time.Time
}
