package httpadapter

import (
	"time"

	"z-finance-api/internal/entity"
)

type createTransactionRequest struct {
	AccountID   string `json:"account_id"`
	UserID      string `json:"user_id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Amount      int64  `json:"amount"`
	Fee         int64  `json:"fee"`
	NetAmount   int64  `json:"net_amount"`
	ExternalRef string `json:"external_ref"`
	OccurredAt  string `json:"occurred_at"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type preRegistrationRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type preRegistrationVerifyRequest struct {
	PreRegistrationID string `json:"pre_registration_id"`
	Token             string `json:"token"`
}

type preRegistrationResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	EmailStatus   string `json:"email_status"`
	PhoneStatus   string `json:"phone_status"`
	EmailVerified bool   `json:"email_verified"`
	PhoneVerified bool   `json:"phone_verified"`
	ExpiresAt     string `json:"expires_at"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type preRegistrationLookupResponse struct {
	State           string                  `json:"state"`
	PreRegistration *preRegistrationResponse `json:"pre_registration,omitempty"`
}

type auditEventRequest struct {
	Action          string         `json:"action"`
	EntityType      string         `json:"entity_type,omitempty"`
	EntityID        string         `json:"entity_id,omitempty"`
	Origin          string         `json:"origin,omitempty"`
	ClientTimestamp string         `json:"client_timestamp,omitempty"`
	Data            map[string]any `json:"data,omitempty"`
}

type loginResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	UserType     string `json:"user_type,omitempty"`
	PlanCode     string            `json:"plan_code,omitempty"`
	UXMode       string            `json:"ux_mode,omitempty"`
	AllowedModes []string          `json:"allowed_modes,omitempty"`
	Features     map[string]bool   `json:"features,omitempty"`
	Limits       map[string]int64  `json:"limits,omitempty"`
	Capabilities capabilitySnapshot `json:"capabilities,omitempty"`
}

type capabilitySnapshot struct {
	PlanCode     string           `json:"plan_code"`
	UXMode       string           `json:"ux_mode"`
	AllowedModes []string         `json:"allowed_modes"`
	Features     map[string]bool  `json:"features"`
	Limits       map[string]int64 `json:"limits"`
	AllowedFeatures  []string                 `json:"allowed_features,omitempty"`
	DisabledFeatures []capabilityDisabledFeature `json:"disabled_features,omitempty"`
	Suggestions      []string                 `json:"suggestions,omitempty"`
}

type capabilityDisabledFeature struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

type transactionResponse struct {
	ID             string `json:"id"`
	AccountID      string `json:"account_id"`
	UserID         string `json:"user_id"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	Amount         int64  `json:"amount"`
	Fee            int64  `json:"fee"`
	NetAmount      int64  `json:"net_amount"`
	IdempotencyKey string `json:"idempotency_key"`
	ExternalRef    string `json:"external_ref,omitempty"`
	OccurredAt     string `json:"occurred_at"`
	CreatedAt      string `json:"created_at"`
}

type transactionListResponse struct {
	Items      []transactionResponse `json:"items"`
	NextCursor string                `json:"next_cursor,omitempty"`
}

type withdrawalRequest struct {
	AccountID   string `json:"account_id"`
	UserID      string `json:"user_id"`
	Status      string `json:"status"`
	Amount      int64  `json:"amount"`
	Fee         int64  `json:"fee"`
	NetAmount   int64  `json:"net_amount"`
	ExternalRef string `json:"external_ref"`
	OccurredAt  string `json:"occurred_at"`
}

type reversalRequest struct {
	TransactionID string `json:"transaction_id"`
	Reason        string `json:"reason"`
}

type paymentValidateRequest struct {
	Barcode string `json:"barcode"`
}

type paymentValidateResponse struct {
	Barcode string `json:"barcode"`
	Valid   bool   `json:"valid"`
}

type paymentScheduleRequest struct {
	AccountID   string     `json:"account_id"`
	UserID      string     `json:"user_id"`
	Amount      int64      `json:"amount"`
	Fee         int64      `json:"fee"`
	NetAmount   int64      `json:"net_amount"`
	Barcode     string     `json:"barcode"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	DueDate     *time.Time `json:"due_date"`
	ExternalRef string     `json:"external_ref"`
}

type paymentWebhookRequest struct {
	PaymentID     string `json:"payment_id"`
	TransactionID string `json:"transaction_id"`
	UserID        string `json:"user_id"`
}

type paymentResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	AccountID string `json:"account_id"`
	Status    string `json:"status"`
	Amount    int64  `json:"amount"`
	Fee       int64  `json:"fee"`
	NetAmount int64  `json:"net_amount"`
	Barcode   string `json:"barcode,omitempty"`
}

type cardAuthorizeRequest struct {
	UserID       string `json:"user_id"`
	AccountID    string `json:"account_id"`
	Amount       int64  `json:"amount"`
	Fee          int64  `json:"fee"`
	NetAmount    int64  `json:"net_amount"`
	MerchantName string `json:"merchant_name"`
	MerchantMCC  string `json:"merchant_mcc"`
	AuthCode     string `json:"auth_code"`
	ExternalRef  string `json:"external_ref"`
}

type cardAuthorizeResponse struct {
	Approved        bool   `json:"approved"`
	AuthorizationID string `json:"authorization_id,omitempty"`
	Reason          string `json:"reason,omitempty"`
}

type cardWebhookRequest struct {
	AuthorizationID string `json:"authorization_id"`
	TransactionID   string `json:"transaction_id"`
	UserID          string `json:"user_id"`
}

type quoteRequest struct {
	Pair       string `json:"pair"`
	Price      int64  `json:"price"`
	TTLSeconds int64  `json:"ttl_seconds"`
	Source     string `json:"source"`
}

type quoteResponse struct {
	Pair      string `json:"pair"`
	Price     int64  `json:"price"`
	ExpiresAt string `json:"expires_at"`
	Source    string `json:"source"`
}

type swapRequest struct {
	UserID          string           `json:"user_id"`
	DebitAccountID  string           `json:"debit_account_id"`
	CreditAccountID string           `json:"credit_account_id"`
	Side            entity.TradeSide `json:"side"`
	BaseCurrency    string           `json:"base_currency"`
	QuoteCurrency   string           `json:"quote_currency"`
	Price           int64            `json:"price"`
	Quantity        int64            `json:"quantity"`
	Fee             int64            `json:"fee"`
	ExternalRef     string           `json:"external_ref"`
	AutoExecute     bool             `json:"auto_execute"`
}

type swapResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Side   string `json:"side"`
}

type cryptoPayRequest struct {
	UserID      string `json:"user_id"`
	FiatAmount  int64  `json:"fiat_amount"`
	Destination string `json:"destination"`
}

type cryptoPayResponse struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	AccountID     string `json:"account_id"`
	Asset         string `json:"asset"`
	Network       string `json:"network"`
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	Fee           int64  `json:"fee"`
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
	CreatedAt     string `json:"created_at"`
}

type sendPixFromCryptoRequest struct {
	UserID      string `json:"user_id"`
	AccountID   string `json:"account_id"`
	AmountBRL   int64  `json:"amount_brl"`
	Asset       string `json:"asset"`
	ExternalRef string `json:"external_ref"`
}

type userSettingsRequest struct {
	UserID             string   `json:"user_id"`
	ConversionPriority []string `json:"conversion_priority"`
	AllowCryptoToFiat  bool     `json:"allow_crypto_to_fiat"`
	AutoConvertEnabled bool     `json:"auto_convert_enabled"`
	AutoConvertAsset   string   `json:"auto_convert_asset"`
	AutoConvertMinAmount int64  `json:"auto_convert_min_amount"`
	UXMode             string   `json:"ux_mode"`
}

type createInvoiceRequest struct {
	UserID    string `json:"user_id"`
	AmountBRL int64  `json:"amount_brl"`
}

type invoiceResponse struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	AmountBRL    int64  `json:"amount_brl"`
	PixCopyPaste string `json:"pix_copy_paste"`
	USDTAddress  string `json:"usdt_address"`
	CreatedAt    string `json:"created_at"`
	QRPayload    string `json:"qr_payload,omitempty"`
}

type payInvoiceRequest struct {
	InvoiceID      string `json:"invoice_id"`
	PayerUserID    string `json:"payer_user_id"`
	PayerAccountID string `json:"payer_account_id"`
	Method         string `json:"method"`
}

type payInvoiceResponse struct {
	Invoice        invoiceResponse   `json:"invoice"`
	Method         string            `json:"method"`
	PixTransfer    *pixResponse      `json:"pix_transfer,omitempty"`
	CryptoTransfer *cryptoPayResponse `json:"crypto_transfer,omitempty"`
}

type confirmTransactionRequest struct {
	TransactionID string `json:"transaction_id"`
}

type kycLimitRequest struct {
	Level        string `json:"level"`
	DailyLimit   int64  `json:"daily_limit"`
	MonthlyLimit int64  `json:"monthly_limit"`
}

type kycLimitResponse struct {
	Level        string `json:"level"`
	DailyLimit   int64  `json:"daily_limit"`
	MonthlyLimit int64  `json:"monthly_limit"`
}

type kycUpgradeRequest struct {
	UserID      string `json:"user_id"`
	Level       string `json:"level"`
	Status      string `json:"status"`
	ProviderRef string `json:"provider_ref"`
}

type pixSendRequest struct {
	AccountID   string `json:"account_id"`
	UserID      string `json:"user_id"`
	Amount      int64  `json:"amount"`
	Fee         int64  `json:"fee"`
	NetAmount   int64  `json:"net_amount"`
	ExternalRef string `json:"external_ref"`
	Metadata    map[string]any `json:"metadata"`
}

type pixKeyRequest struct {
	UserID    string `json:"user_id"`
	AccountID string `json:"account_id"`
	Type      string `json:"type"`
	Key       string `json:"key"`
}

type pixWebhookRequest struct {
	TransferID string `json:"transfer_id"`
	Status     string `json:"status"`
}

type pixResponse struct {
	ID            string `json:"id"`
	TransactionID string `json:"transaction_id,omitempty"`
	UserID        string `json:"user_id"`
	AccountID     string `json:"account_id"`
	Direction     string `json:"direction"`
	Status        string `json:"status"`
	Amount        int64  `json:"amount"`
	Fee           int64  `json:"fee"`
	NetAmount     int64  `json:"net_amount"`
	EndToEndID    string `json:"end_to_end_id,omitempty"`
	ExternalRef   string `json:"external_ref,omitempty"`
	ConfirmedAt   string `json:"confirmed_at,omitempty"`
	OccurredAt    string `json:"occurred_at"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type pixKeyResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	AccountID string `json:"account_id"`
	Type      string `json:"type"`
	Key       string `json:"key"`
	CreatedAt string `json:"created_at"`
}

type accountBalanceResponse struct {
	AccountID string `json:"account_id"`
	Currency  string `json:"currency"`
	Scale     int32  `json:"scale"`
	Balance   int64  `json:"balance"`
}
