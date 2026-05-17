package entity

import "time"

type Plan struct {
	ID                string
	Code              string
	Description       string
	MonthlyPriceCents int64
	CreatedAt         time.Time
}

type UserPlan struct {
	UserID    string
	PlanID    string
	ValidFrom time.Time
	ValidUntil *time.Time
	CreatedAt time.Time
}

type PricingOperationType string

const (
	PricingOperationPixIn       PricingOperationType = "PIX_IN"
	PricingOperationPixOut      PricingOperationType = "PIX_OUT"
	PricingOperationPixToCrypto PricingOperationType = "PIX_TO_CRYPTO"
	PricingOperationCryptoToPix PricingOperationType = "CRYPTO_TO_PIX"
	PricingOperationSwap        PricingOperationType = "SWAP"
	PricingOperationCardCrypto  PricingOperationType = "CARD_CRYPTO"
	PricingOperationInvoice     PricingOperationType = "INVOICE"
)

type PricingFeeType string

const (
	PricingFeePercentage PricingFeeType = "PERCENTAGE"
	PricingFeeFixed      PricingFeeType = "FIXED"
)

const PricingAssetAny = "ANY"

type PricingRule struct {
	ID            string
	PlanID        string
	PricingVersionID string
	UserType      UserType
	OperationType PricingOperationType
	Asset         string
	FeeType       PricingFeeType
	FeeValue      int64
	MinFee        *int64
	MaxFee        *int64
	CreatedAt     time.Time
}

type FeeResult struct {
	Amount    int64
	Fee       int64
	NetAmount int64
}
