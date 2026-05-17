package entity

import "time"

type UserFeatureOverride struct {
	ID          string
	UserID      string
	FeatureCode string
	Enabled     bool
	Reason      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
