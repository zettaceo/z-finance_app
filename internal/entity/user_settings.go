package entity

import "time"

type UserSettings struct {
	UserID              string
	ConversionPriority  []string
	AllowCryptoToFiat   bool
	AutoConvertPixIn    bool
	PixInTargetAsset    string
	PixInPercentage     int32
	AutoConvertEnabled  bool
	AutoConvertAsset    string
	AutoConvertMinAmount int64
	FallbackAsset       string
	UXMode              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
