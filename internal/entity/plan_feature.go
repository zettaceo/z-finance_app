package entity

import "time"

type PlanFeature struct {
	ID          string
	PlanID      string
	FeatureCode string
	Enabled     bool
	Metadata    []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
