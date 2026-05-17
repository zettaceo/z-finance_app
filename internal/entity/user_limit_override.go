package entity

import "time"

type UserLimitOverride struct {
	ID          string
	UserID      string
	LimitCode   string
	LimitValue  int64
	LimitWindow string
	Reason      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
