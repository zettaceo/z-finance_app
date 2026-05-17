package entity

import "time"

type PricingVersionStatus string

const (
	PricingVersionActive   PricingVersionStatus = "ACTIVE"
	PricingVersionInactive PricingVersionStatus = "INACTIVE"
)

type PricingVersion struct {
	ID          string
	Code        string
	Description string
	Status      PricingVersionStatus
	ValidFrom   time.Time
	ValidUntil  *time.Time
	CreatedAt   time.Time
}
