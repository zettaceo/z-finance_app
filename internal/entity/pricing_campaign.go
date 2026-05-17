package entity

import "time"

type PricingCampaignStatus string

const (
	PricingCampaignActive   PricingCampaignStatus = "ACTIVE"
	PricingCampaignInactive PricingCampaignStatus = "INACTIVE"
)

type PricingCampaign struct {
	ID          string
	Code        string
	Description string
	Status      PricingCampaignStatus
	Priority    int
	ValidFrom   time.Time
	ValidUntil  *time.Time
	CreatedAt   time.Time
}

type PricingCampaignRule struct {
	ID            string
	CampaignID    string
	PlanID        *string
	UserType      UserType
	OperationType PricingOperationType
	Asset         string
	FeeType       PricingFeeType
	FeeValue      int64
	MinFee        *int64
	MaxFee        *int64
	CreatedAt     time.Time
}
