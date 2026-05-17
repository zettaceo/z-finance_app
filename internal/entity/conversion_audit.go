package entity

import "time"

type ConversionAudit struct {
	ID             string
	UserID         string
	OperationType  PricingOperationType
	Asset          string
	GrossAmount    int64
	Fee            int64
	NetAmount      int64
	QuotePrice     int64
	SpreadBps      int64
	LiquiditySource string
	QuotedAt       *time.Time
	RelatedType    string
	RelatedID      string
	CreatedAt      time.Time
}
