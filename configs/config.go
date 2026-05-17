package configs

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL            string
	HTTPPort               string
	AppName                string
	WebhookSecret          string
	WebhookAllowedIPs      []string
	WebhookRateLimitPerMin int64
	OtelServiceName        string
	OtelEndpoint           string
	OtelInsecure           bool

	// Zetta ecosystem service URLs (optional — mocks used when empty)
	ZSwapURL    string
	ZSwapAPIKey string
	ObeliskZURL    string
	ObeliskZAPIKey string
	ZPayURL     string
	ZPayAPIKey  string
	ZionAIURL   string
	ZionAIAPIKey string
	AlertPixPendingThreshold        int64
	AlertPaymentPendingThreshold    int64
	AlertCardPendingThreshold       int64
	AlertTransactionHoldThreshold   int64
	AlertWebhookRetryDeadThreshold  int64
	AuthSecret             string
	AccessTokenTTLMinutes  int64
	RefreshTokenTTLDays    int64
	LoginMaxAttempts       int64
	LoginWindowSeconds     int64
	VelocityMaxTxPerMinute int64
	VelocityMaxTxPerHour   int64
	VelocityMaxAmtPerMin   int64
	VelocityMaxAmtPerHour  int64
	PreRegistrationExpiryHours   int64
	PreRegistrationMaxEmailAttempts int64
	PreRegistrationMaxPhoneAttempts int64
	PreRegistrationBlockMinutes  int64
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:   firstNonEmpty(os.Getenv("DB_URL"), os.Getenv("DATABASE_URL")),
		HTTPPort:      defaultIfEmpty(firstNonEmpty(os.Getenv("HTTP_PORT"), os.Getenv("PORT")), "8080"),
		AppName:       defaultIfEmpty(os.Getenv("APP_NAME"), "Z-BANK"),
		WebhookSecret: os.Getenv("WEBHOOK_SECRET"),
		WebhookAllowedIPs: parseCSVEnv("WEBHOOK_ALLOWED_IPS"),
		WebhookRateLimitPerMin: parseInt64Env("WEBHOOK_RATE_LIMIT_PER_MINUTE", 120),
		OtelServiceName: defaultIfEmpty(os.Getenv("OTEL_SERVICE_NAME"), defaultIfEmpty(os.Getenv("APP_NAME"), "Z-BANK")),
		OtelEndpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OtelInsecure:    parseBoolEnv("OTEL_EXPORTER_OTLP_INSECURE", true),
		AlertPixPendingThreshold:       parseInt64Env("ALERT_PIX_PENDING_THRESHOLD", 0),
		AlertPaymentPendingThreshold:   parseInt64Env("ALERT_PAYMENT_PENDING_THRESHOLD", 0),
		AlertCardPendingThreshold:      parseInt64Env("ALERT_CARD_PENDING_THRESHOLD", 0),
		AlertTransactionHoldThreshold:  parseInt64Env("ALERT_TRANSACTION_HOLD_THRESHOLD", 0),
		AlertWebhookRetryDeadThreshold: parseInt64Env("ALERT_WEBHOOK_RETRY_DEAD_THRESHOLD", 0),
		ZSwapURL:     os.Getenv("Z_SWAP_URL"),
		ZSwapAPIKey:  os.Getenv("Z_SWAP_API_KEY"),
		ObeliskZURL:     os.Getenv("OBELISK_Z_URL"),
		ObeliskZAPIKey:  os.Getenv("OBELISK_Z_API_KEY"),
		ZPayURL:      os.Getenv("Z_PAY_URL"),
		ZPayAPIKey:   os.Getenv("Z_PAY_API_KEY"),
		ZionAIURL:    os.Getenv("ZION_AI_URL"),
		ZionAIAPIKey: os.Getenv("ZION_AI_API_KEY"),
		AuthSecret:    os.Getenv("AUTH_SECRET"),
		AccessTokenTTLMinutes: parseInt64Env("AUTH_ACCESS_TTL_MINUTES", 15),
		RefreshTokenTTLDays:   parseInt64Env("AUTH_REFRESH_TTL_DAYS", 30),
		LoginMaxAttempts:      parseInt64Env("AUTH_LOGIN_MAX_ATTEMPTS", 5),
		LoginWindowSeconds:    parseInt64Env("AUTH_LOGIN_WINDOW_SECONDS", 300),
		VelocityMaxTxPerMinute: parseInt64Env(
			"VELOCITY_MAX_TX_PER_MINUTE",
			0,
		),
		VelocityMaxTxPerHour: parseInt64Env(
			"VELOCITY_MAX_TX_PER_HOUR",
			0,
		),
		VelocityMaxAmtPerMin: parseInt64Env(
			"VELOCITY_MAX_AMOUNT_PER_MINUTE",
			0,
		),
		VelocityMaxAmtPerHour: parseInt64Env(
			"VELOCITY_MAX_AMOUNT_PER_HOUR",
			0,
		),
		PreRegistrationExpiryHours: parseInt64Env("PRE_REG_EXPIRY_HOURS", 48),
		PreRegistrationMaxEmailAttempts: parseInt64Env("PRE_REG_MAX_EMAIL_ATTEMPTS", 5),
		PreRegistrationMaxPhoneAttempts: parseInt64Env("PRE_REG_MAX_PHONE_ATTEMPTS", 5),
		PreRegistrationBlockMinutes: parseInt64Env("PRE_REG_BLOCK_MINUTES", 30),
	}
	if cfg.AuthSecret == "" {
		return Config{}, errors.New("AUTH_SECRET nao configurado")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DB_URL/DATABASE_URL nao configurada")
	}
	return cfg, nil
}

func parseInt64Env(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseBoolEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "true", "1", "yes", "y":
		return true
	case "false", "0", "no", "n":
		return false
	default:
		return fallback
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func defaultIfEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func parseCSVEnv(key string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
