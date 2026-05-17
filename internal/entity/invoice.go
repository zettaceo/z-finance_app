package entity

import "time"

type Invoice struct {
	ID          string
	UserID      string
	AmountBRL   int64
	PixCopyPaste string
	USDTAddress string
	IdempotencyKey string
	CreatedAt   time.Time
}
