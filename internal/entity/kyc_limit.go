package entity

import "time"

type KycLimit struct {
	Level        KYCLevel
	DailyLimit   int64
	MonthlyLimit int64
	CreatedAt    time.Time
}
