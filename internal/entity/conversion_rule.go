package entity

import "time"

type ConversionTrigger string

const (
	ConversionTriggerPixIn   ConversionTrigger = "PIX_IN"
	ConversionTriggerCardJIT ConversionTrigger = "CARD_AUTH"
	ConversionTriggerPayment ConversionTrigger = "PAYMENT"
)

type ConversionRule struct {
	ID          string
	UserID      string
	Trigger     ConversionTrigger
	SourceAsset string
	TargetAsset string
	Enabled     bool
	CreatedAt   time.Time
}
