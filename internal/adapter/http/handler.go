package httpadapter

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/crypto/address"
	"z-finance-api/internal/infra/events"
	"z-finance-api/internal/infra/pricing"
	"z-finance-api/internal/repository"
	"z-finance-api/internal/usecase"
)

type Dependencies struct {
	Pool            *pgxpool.Pool
	TransactionRepo repository.TransactionRepository
	WebhookRepo     repository.WebhookRepository
	WebhookRetryRepo repository.WebhookRetryRepository
	WebhookAllowedIPs []string
	WebhookRateLimitPerMinute int
	AlertPixPendingThreshold       int
	AlertPaymentPendingThreshold   int
	AlertCardPendingThreshold      int
	AlertTransactionHoldThreshold  int
	AlertWebhookRetryDeadThreshold int
	AccountRepo     repository.AccountRepository
	UserRepo        repository.UserRepository
	PixRepo         repository.PixTransferRepository
	PlanRepo        repository.PlanRepository
	UserPlanRepo    repository.UserPlanRepository
	PricingRuleRepo repository.PricingRuleRepository
	PricingVersionRepo repository.PricingVersionRepository
	PlanFeatureRepo repository.PlanFeatureRepository
	PlanLimitRepo repository.PlanLimitRepository
	PricingCampaignRepo repository.PricingCampaignRepository
	PricingCampaignRuleRepo repository.PricingCampaignRuleRepository
	RoleRepo        repository.RoleRepository
	UserRoleRepo    repository.UserRoleRepository
	RoleSeparationRepo repository.RoleSeparationRepository
	RegulatoryProfileRepo repository.RegulatoryProfileRepository
	KycRepo         repository.KYCRepository
	KycLimitRepo    repository.KycLimitRepository
	PaymentRepo     repository.PaymentRepository
	CardRepo        repository.CardAuthorizationRepository
	TradeRepo       repository.TradeOrderRepository
	SettingsRepo    repository.UserSettingsRepository
	UserFeatureOverrideRepo repository.UserFeatureOverrideRepository
	UserLimitOverrideRepo repository.UserLimitOverrideRepository
	RefreshTokenRepo repository.RefreshTokenRepository
	PricingCache    *pricing.Cache
	CreateTxUC      *usecase.CreateTransactionUseCase
	RegisterPixKeyUC *usecase.RegisterPixKeyUseCase
	SendPixUC        *usecase.SendPixUseCase
	PixWebhookUC     *usecase.HandlePixWebhookUseCase
	PayCryptoWithFiatUC *usecase.PayCryptoWithFiatUseCase
	AuthorizeCardJITUC *usecase.AuthorizeCardJITUseCase
	SendPixFromCryptoUC *usecase.SendPixFromCryptoUseCase
	CreateInvoiceUC     *usecase.CreateInvoiceUseCase
	PayInvoiceUC        *usecase.PayInvoiceUseCase
	EnsureFiatCoverageUC *usecase.EnsureFiatCoverageUseCase
	TokenService   *usecase.TokenService
	LoginUC        *usecase.LoginUseCase
	RefreshTokenUC *usecase.RefreshTokenUseCase
	LogoutUC       *usecase.LogoutUseCase
	PricingUC      *usecase.ResolvePricingUseCase
	PreRegistrationUC *usecase.PreRegistrationUseCase
	EventBus        *events.Bus
	VelocityChecker *usecase.VelocityChecker
	WebhookSecret   string
}

type Handler struct {
	pool            *pgxpool.Pool
	txRepo          repository.TransactionRepository
	webhookRepo     repository.WebhookRepository
	webhookRetryRepo repository.WebhookRetryRepository
	webhookAllowedIPs []string
	webhookRateLimitPerMinute int
	webhookRateLimiter *webhookRateLimiter
	alertPixPendingThreshold       int
	alertPaymentPendingThreshold   int
	alertCardPendingThreshold      int
	alertTransactionHoldThreshold  int
	alertWebhookRetryDeadThreshold int
	accountRepo     repository.AccountRepository
	userRepo        repository.UserRepository
	pixRepo         repository.PixTransferRepository
	planRepo        repository.PlanRepository
	userPlanRepo    repository.UserPlanRepository
	pricingRuleRepo repository.PricingRuleRepository
	pricingVersionRepo repository.PricingVersionRepository
	planFeatureRepo repository.PlanFeatureRepository
	planLimitRepo repository.PlanLimitRepository
	pricingCampaignRepo repository.PricingCampaignRepository
	pricingCampaignRuleRepo repository.PricingCampaignRuleRepository
	roleRepo        repository.RoleRepository
	userRoleRepo    repository.UserRoleRepository
	roleSeparationRepo repository.RoleSeparationRepository
	regulatoryProfileRepo repository.RegulatoryProfileRepository
	kycRepo         repository.KYCRepository
	kycLimitRepo    repository.KycLimitRepository
	paymentRepo     repository.PaymentRepository
	cardRepo        repository.CardAuthorizationRepository
	tradeRepo       repository.TradeOrderRepository
	settingsRepo    repository.UserSettingsRepository
	userFeatureOverrideRepo repository.UserFeatureOverrideRepository
	userLimitOverrideRepo repository.UserLimitOverrideRepository
	refreshTokenRepo repository.RefreshTokenRepository
	pricingCache    *pricing.Cache
	createTxUC      *usecase.CreateTransactionUseCase
	registerPixKeyUC *usecase.RegisterPixKeyUseCase
	sendPixUC        *usecase.SendPixUseCase
	pixWebhookUC     *usecase.HandlePixWebhookUseCase
	payCryptoWithFiatUC *usecase.PayCryptoWithFiatUseCase
	authorizeCardJITUC *usecase.AuthorizeCardJITUseCase
	sendPixFromCryptoUC *usecase.SendPixFromCryptoUseCase
	createInvoiceUC     *usecase.CreateInvoiceUseCase
	payInvoiceUC        *usecase.PayInvoiceUseCase
	ensureFiatCoverageUC *usecase.EnsureFiatCoverageUseCase
	tokenService   *usecase.TokenService
	loginUC        *usecase.LoginUseCase
	refreshTokenUC *usecase.RefreshTokenUseCase
	logoutUC       *usecase.LogoutUseCase
	pricingUC      *usecase.ResolvePricingUseCase
	preRegistrationUC *usecase.PreRegistrationUseCase
	eventBus        *events.Bus
	velocityChecker *usecase.VelocityChecker
	webhookSecret   string
}

func NewHandler(deps Dependencies) *Handler {
	handler := &Handler{
		pool:            deps.Pool,
		txRepo:          deps.TransactionRepo,
		webhookRepo:     deps.WebhookRepo,
		webhookRetryRepo: deps.WebhookRetryRepo,
		webhookAllowedIPs: deps.WebhookAllowedIPs,
		webhookRateLimitPerMinute: deps.WebhookRateLimitPerMinute,
		webhookRateLimiter: &webhookRateLimiter{},
		alertPixPendingThreshold: deps.AlertPixPendingThreshold,
		alertPaymentPendingThreshold: deps.AlertPaymentPendingThreshold,
		alertCardPendingThreshold: deps.AlertCardPendingThreshold,
		alertTransactionHoldThreshold: deps.AlertTransactionHoldThreshold,
		alertWebhookRetryDeadThreshold: deps.AlertWebhookRetryDeadThreshold,
		accountRepo:     deps.AccountRepo,
		userRepo:        deps.UserRepo,
		pixRepo:         deps.PixRepo,
		planRepo:        deps.PlanRepo,
		userPlanRepo:    deps.UserPlanRepo,
		pricingRuleRepo: deps.PricingRuleRepo,
		pricingVersionRepo: deps.PricingVersionRepo,
		planFeatureRepo: deps.PlanFeatureRepo,
		planLimitRepo: deps.PlanLimitRepo,
		pricingCampaignRepo: deps.PricingCampaignRepo,
		pricingCampaignRuleRepo: deps.PricingCampaignRuleRepo,
		roleRepo:        deps.RoleRepo,
		userRoleRepo:    deps.UserRoleRepo,
		roleSeparationRepo: deps.RoleSeparationRepo,
		regulatoryProfileRepo: deps.RegulatoryProfileRepo,
		kycRepo:         deps.KycRepo,
		kycLimitRepo:    deps.KycLimitRepo,
		paymentRepo:     deps.PaymentRepo,
		cardRepo:        deps.CardRepo,
		tradeRepo:       deps.TradeRepo,
		settingsRepo:    deps.SettingsRepo,
		userFeatureOverrideRepo: deps.UserFeatureOverrideRepo,
		userLimitOverrideRepo: deps.UserLimitOverrideRepo,
		refreshTokenRepo: deps.RefreshTokenRepo,
		pricingCache:    deps.PricingCache,
		createTxUC:      deps.CreateTxUC,
		registerPixKeyUC: deps.RegisterPixKeyUC,
		sendPixUC:        deps.SendPixUC,
		pixWebhookUC:     deps.PixWebhookUC,
		payCryptoWithFiatUC: deps.PayCryptoWithFiatUC,
		authorizeCardJITUC: deps.AuthorizeCardJITUC,
		sendPixFromCryptoUC: deps.SendPixFromCryptoUC,
		createInvoiceUC:     deps.CreateInvoiceUC,
		payInvoiceUC:        deps.PayInvoiceUC,
		ensureFiatCoverageUC: deps.EnsureFiatCoverageUC,
		tokenService:   deps.TokenService,
		loginUC:        deps.LoginUC,
		refreshTokenUC: deps.RefreshTokenUC,
		logoutUC:       deps.LogoutUC,
		pricingUC:      deps.PricingUC,
		preRegistrationUC: deps.PreRegistrationUC,
		eventBus:        deps.EventBus,
		velocityChecker: deps.VelocityChecker,
		webhookSecret:   deps.WebhookSecret,
	}
	return handler
}

func BuildHTTPHandler(handler *Handler) http.Handler {
	mux := http.NewServeMux()
	handler.registerRoutes(mux)
	return withRequestID(withTrace(withCORS(withStandardHeaders(withRequestLogging(mux)))))
}

func (h *Handler) registerRoutes(mux *http.ServeMux) {
	admin := func(roles ...string) func(http.Handler) http.Handler {
		allowed := make([]string, 0, len(roles)+1)
		allowed = append(allowed, "ADMIN")
		allowed = append(allowed, roles...)
		return func(next http.Handler) http.Handler {
			return h.requireAuth(h.requireRoles(allowed...)(next))
		}
	}

	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/health/db", h.handleHealthDB)
	mux.Handle("/debug/vars", expvar.Handler())
	mux.Handle("/pre-registrations", http.HandlerFunc(h.handlePreRegistrations))
	mux.Handle("/pre-registrations/verify/email", http.HandlerFunc(h.handlePreRegistrationVerifyEmail))
	mux.Handle("/pre-registrations/verify/phone", http.HandlerFunc(h.handlePreRegistrationVerifyPhone))
	mux.Handle("/pre-registrations/lookup", http.HandlerFunc(h.handlePreRegistrationLookup))
	mux.Handle("/auth/login", http.HandlerFunc(h.handleLogin))
	mux.Handle("/auth/refresh", http.HandlerFunc(h.handleRefresh))
	mux.Handle("/auth/logout", h.requireAuth(http.HandlerFunc(h.handleLogout)))
	mux.Handle("/auth/me", h.requireAuth(http.HandlerFunc(h.handleAuthMe)))
	mux.Handle("/transactions", h.requireAuth(http.HandlerFunc(h.handleTransactions)))
	mux.Handle("/transactions/reverse", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handleTransactionReverse))))
	mux.Handle("/transactions/confirm", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handleTransactionConfirm))))
	mux.Handle("/transactions/reject", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handleTransactionReject))))
	mux.Handle("/pix/keys", h.requireAuth(http.HandlerFunc(h.handlePixKeys)))
	mux.Handle("/pix/send", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handlePixSend))))
	mux.Handle("/pix/send/crypto", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handlePixSendFromCrypto))))
	mux.Handle("/pix/webhook", withWebhookSecurity(h, withWebhookRetry("pix_receive", "/pix/webhook", h.webhookRetryRepo,
		validateWebhookSignature(h.webhookSecret, h.webhookRepo, h.auditWebhook, http.HandlerFunc(h.handlePixWebhook)))))
	mux.Handle("/pricing/quote", http.HandlerFunc(h.handlePricingQuote))
	mux.Handle("/admin/plans", admin("AUDIT")(http.HandlerFunc(h.handleAdminPlans)))
	mux.Handle("/admin/plan-history", admin("COMPLIANCE", "AUDIT", "OPS", "VIEWER")(http.HandlerFunc(h.handleAdminPlanHistory)))
	mux.Handle("/admin/plan-feature-history", admin("COMPLIANCE", "AUDIT", "OPS", "VIEWER")(http.HandlerFunc(h.handleAdminPlanFeatureHistory)))
	mux.Handle("/admin/plan-limit-history", admin("COMPLIANCE", "AUDIT", "OPS", "VIEWER")(http.HandlerFunc(h.handleAdminPlanLimitHistory)))
	mux.Handle("/admin/user-plans", admin()(http.HandlerFunc(h.handleAdminUserPlans)))
	mux.Handle("/admin/pricing/rules", admin("AUDIT")(http.HandlerFunc(h.handleAdminPricingRules)))
	mux.Handle("/admin/pricing/rules/", admin()(http.HandlerFunc(h.handleAdminPricingRuleByID)))
	mux.Handle("/admin/pricing/versions", admin("AUDIT")(http.HandlerFunc(h.handleAdminPricingVersions)))
	mux.Handle("/admin/pricing/versions/active", admin("COMPLIANCE", "AUDIT", "OPS", "VIEWER")(http.HandlerFunc(h.handleAdminPricingVersionActive)))
	mux.Handle("/admin/pricing/versions/", admin()(http.HandlerFunc(h.handleAdminPricingVersionStatus)))
	mux.Handle("/admin/plan-features", admin("COMPLIANCE", "AUDIT")(http.HandlerFunc(h.handleAdminPlanFeatures)))
	mux.Handle("/admin/plan-features/rollback", admin("COMPLIANCE")(http.HandlerFunc(h.handleAdminPlanFeaturesRollback)))
	mux.Handle("/admin/plan-limits", admin("COMPLIANCE", "AUDIT")(http.HandlerFunc(h.handleAdminPlanLimits)))
	mux.Handle("/admin/plan-limits/rollback", admin("COMPLIANCE")(http.HandlerFunc(h.handleAdminPlanLimitsRollback)))
	mux.Handle("/admin/plans/rollback", admin()(http.HandlerFunc(h.handleAdminPlansRollback)))
	mux.Handle("/admin/pricing/campaigns", admin("AUDIT")(http.HandlerFunc(h.handleAdminPricingCampaigns)))
	mux.Handle("/admin/pricing/campaigns/", admin()(http.HandlerFunc(h.handleAdminPricingCampaignStatus)))
	mux.Handle("/admin/pricing/campaigns/rules", admin("AUDIT")(http.HandlerFunc(h.handleAdminPricingCampaignRules)))
	mux.Handle("/admin/roles", admin()(http.HandlerFunc(h.handleAdminRoles)))
	mux.Handle("/admin/user-roles", admin()(http.HandlerFunc(h.handleAdminUserRoles)))
	mux.Handle("/admin/roles/separation", admin()(http.HandlerFunc(h.handleAdminRoleSeparation)))
	mux.Handle("/admin/regulatory-profiles", admin("COMPLIANCE", "AUDIT")(http.HandlerFunc(h.handleAdminRegulatoryProfiles)))
	mux.Handle("/admin/reconcile/summary", admin("COMPLIANCE", "OPS", "AUDIT", "VIEWER")(http.HandlerFunc(h.handleAdminReconcileSummary)))
	mux.Handle("/admin/reconcile/pending", admin("COMPLIANCE", "OPS", "AUDIT", "VIEWER")(http.HandlerFunc(h.handleAdminReconcilePending)))
	mux.Handle("/admin/webhooks/retry", admin("COMPLIANCE", "OPS", "AUDIT", "VIEWER")(http.HandlerFunc(h.handleAdminWebhookRetryList)))
	mux.Handle("/admin/observability/summary", admin("COMPLIANCE", "OPS", "AUDIT", "VIEWER")(http.HandlerFunc(h.handleAdminObservabilitySummary)))
	mux.Handle("/admin/audit/logs", admin("COMPLIANCE", "AUDIT", "OPS", "VIEWER")(http.HandlerFunc(h.handleAdminAuditLogs)))
	mux.Handle("/admin/audit/archive", admin("COMPLIANCE")(http.HandlerFunc(h.handleAdminAuditArchive)))
	mux.Handle("/admin/alerts/check", admin("COMPLIANCE", "OPS", "AUDIT", "VIEWER")(http.HandlerFunc(h.handleAdminAlertsCheck)))
	mux.Handle("/admin/users", admin("COMPLIANCE", "AUDIT", "OPS")(http.HandlerFunc(h.handleAdminUsers)))
	mux.Handle("/admin/users/", admin("COMPLIANCE", "AUDIT", "OPS")(http.HandlerFunc(h.handleAdminUserByID)))
	mux.Handle("/admin/accounts/", admin("COMPLIANCE", "AUDIT", "OPS", "VIEWER")(http.HandlerFunc(h.handleAdminAccounts)))
	mux.Handle("/payments/validate", http.HandlerFunc(h.handlePaymentValidate))
	mux.Handle("/payments/schedule", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handlePaymentSchedule))))
	mux.Handle("/webhooks/payments/confirm", withWebhookSecurity(h, withWebhookRetry("payment_confirm", "/webhooks/payments/confirm", h.webhookRetryRepo,
		validateWebhookSignature(h.webhookSecret, h.webhookRepo, h.auditWebhook, http.HandlerFunc(h.handlePaymentConfirmWebhook)))))
	mux.Handle("/webhooks/payments/reject", withWebhookSecurity(h, withWebhookRetry("payment_reject", "/webhooks/payments/reject", h.webhookRetryRepo,
		validateWebhookSignature(h.webhookSecret, h.webhookRepo, h.auditWebhook, http.HandlerFunc(h.handlePaymentRejectWebhook)))))
	mux.Handle("/card/authorize", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handleCardAuthorize))))
	mux.Handle("/webhooks/card/confirm", withWebhookSecurity(h, withWebhookRetry("card_confirm", "/webhooks/card/confirm", h.webhookRetryRepo,
		validateWebhookSignature(h.webhookSecret, h.webhookRepo, h.auditWebhook, http.HandlerFunc(h.handleCardConfirmWebhook)))))
	mux.Handle("/webhooks/card/reject", withWebhookSecurity(h, withWebhookRetry("card_reject", "/webhooks/card/reject", h.webhookRetryRepo,
		validateWebhookSignature(h.webhookSecret, h.webhookRepo, h.auditWebhook, http.HandlerFunc(h.handleCardRejectWebhook)))))
	mux.Handle("/crypto/swap", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handleCryptoSwap))))
	mux.Handle("/crypto/pay", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handleCryptoPay))))
	mux.Handle("/invoices", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handleInvoices))))
	mux.Handle("/invoices/pay", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handleInvoicePay))))
	mux.Handle("/user/settings", h.requireAuth(http.HandlerFunc(h.handleUserSettings)))
	mux.Handle("/webhooks/pix/receive", withWebhookSecurity(h, withWebhookRetry("pix_receive", "/webhooks/pix/receive", h.webhookRetryRepo,
		validateWebhookSignature(h.webhookSecret, h.webhookRepo, h.auditWebhook, http.HandlerFunc(h.handlePixWebhook)))))
	mux.Handle("/webhooks/transactions/confirm", withWebhookSecurity(h, withWebhookRetry("transaction_confirm", "/webhooks/transactions/confirm", h.webhookRetryRepo,
		validateWebhookSignature(h.webhookSecret, h.webhookRepo, h.auditWebhook, http.HandlerFunc(h.handleTransactionConfirmWebhook)))))
	mux.Handle("/webhooks/transactions/reject", withWebhookSecurity(h, withWebhookRetry("transaction_reject", "/webhooks/transactions/reject", h.webhookRetryRepo,
		validateWebhookSignature(h.webhookSecret, h.webhookRepo, h.auditWebhook, http.HandlerFunc(h.handleTransactionRejectWebhook)))))
	mux.Handle("/withdrawals", h.requireAuth(requireIdempotency(http.HandlerFunc(h.handleWithdrawal))))
	mux.Handle("/accounts/", h.requireAuth(http.HandlerFunc(h.handleAccountBalance)))
	mux.Handle("/kyc/limits", h.requireAuth(http.HandlerFunc(h.handleKycLimits)))
	mux.Handle("/kyc/upgrade", h.requireAuth(http.HandlerFunc(h.handleKycUpgrade)))
	mux.Handle("/", http.HandlerFunc(h.handleNotFound))
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleHealthDB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		writeError(r.Context(), w, http.StatusServiceUnavailable, "banco indisponivel", "DB_UNAVAILABLE")
		return
	}
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) auditWebhook(ctx context.Context, action string, data any) {
	if h == nil {
		return
	}
	_ = h.appendAudit(ctx, "", action, "webhook", "", data)
}

func resolveAuthUserID(ctx context.Context, payloadUserID string) (string, error) {
	authUserID := userIDFromContext(ctx)
	if authUserID == "" {
		return "", errors.New("usuario nao autenticado")
	}
	if payloadUserID != "" && payloadUserID != authUserID {
		return "", errors.New("user_id divergente")
	}
	return authUserID, nil
}

type adminPlanRequest struct {
	Code              string `json:"code"`
	Description       string `json:"description"`
	MonthlyPriceCents int64  `json:"monthly_price_cents"`
}

type adminUserPlanRequest struct {
	UserID    string `json:"user_id"`
	PlanID    string `json:"plan_id"`
	PlanCode  string `json:"plan_code"`
	ValidFrom string `json:"valid_from"`
	ValidUntil string `json:"valid_until"`
}

type adminPricingRuleRequest struct {
	PlanID        string `json:"plan_id"`
	PricingVersionID string `json:"pricing_version_id"`
	UserType      string `json:"user_type"`
	OperationType string `json:"operation_type"`
	Asset         string `json:"asset"`
	FeeType       string `json:"fee_type"`
	FeeValue      int64  `json:"fee_value"`
	MinFee        *int64 `json:"min_fee"`
	MaxFee        *int64 `json:"max_fee"`
}

type adminPricingVersionRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ValidFrom   string `json:"valid_from"`
	ValidUntil  string `json:"valid_until"`
}

type adminPlanFeatureRequest struct {
	PlanID      string `json:"plan_id"`
	FeatureCode string `json:"feature_code"`
	Enabled     bool   `json:"enabled"`
	Metadata    json.RawMessage `json:"metadata"`
}

type adminPlanLimitRequest struct {
	PlanID     string `json:"plan_id"`
	LimitCode  string `json:"limit_code"`
	LimitValue int64  `json:"limit_value"`
	LimitWindow string `json:"limit_window"`
}

type adminPricingCampaignRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	ValidFrom   string `json:"valid_from"`
	ValidUntil  string `json:"valid_until"`
}

type adminPricingCampaignRuleRequest struct {
	CampaignID    string `json:"campaign_id"`
	PlanID        string `json:"plan_id"`
	UserType      string `json:"user_type"`
	OperationType string `json:"operation_type"`
	Asset         string `json:"asset"`
	FeeType       string `json:"fee_type"`
	FeeValue      int64  `json:"fee_value"`
	MinFee        *int64 `json:"min_fee"`
	MaxFee        *int64 `json:"max_fee"`
}

type adminStatusRequest struct {
	Status string `json:"status"`
}

type adminRoleRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type adminUserRoleRequest struct {
	UserID   string `json:"user_id"`
	RoleCode string `json:"role_code"`
}

type adminRoleSeparationRequest struct {
	RoleCodeA string `json:"role_code_a"`
	RoleCodeB string `json:"role_code_b"`
	Reason    string `json:"reason"`
}

type adminRegulatoryProfileRequest struct {
	UserID                     string `json:"user_id"`
	JurisdictionCode           string `json:"jurisdiction_code"`
	JurisdictionRisk           string `json:"jurisdiction_risk"`
	AMLTier                    string `json:"aml_tier"`
	TravelRuleRequired         bool   `json:"travel_rule_required"`
	SanctionsScreeningRequired bool   `json:"sanctions_screening_required"`
}

type adminUserFeatureRequest struct {
	FeatureCode string `json:"feature_code"`
	Enabled     bool   `json:"enabled"`
	Reason      string `json:"reason,omitempty"`
}

type adminUserLimitRequest struct {
	LimitCode   string `json:"limit_code"`
	LimitValue  int64  `json:"limit_value"`
	LimitWindow string `json:"limit_window,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type adminAccountFreezeRequest struct {
	Frozen bool   `json:"frozen"`
	Reason string `json:"reason,omitempty"`
}

type adminRollbackRequest struct {
	HistoryID string `json:"history_id"`
	Reason    string `json:"reason,omitempty"`
}

type adminAuditArchiveRequest struct {
	OlderThanDays int  `json:"older_than_days"`
	Limit         int  `json:"limit"`
	DryRun        bool `json:"dry_run"`
}

type adminUserSummary struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id,omitempty"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Status     string `json:"status"`
	UserType   string `json:"user_type"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type adminAccountSummary struct {
	ID        string `json:"id"`
	Currency  string `json:"currency"`
	Scale     int32  `json:"scale"`
	Status    string `json:"status"`
	Balance   int64  `json:"balance"`
	CreatedAt string `json:"created_at"`
}

type adminUserFeatureOverrideResponse struct {
	ID          string `json:"id"`
	FeatureCode string `json:"feature_code"`
	Enabled     bool   `json:"enabled"`
	Reason      string `json:"reason,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

type adminUserLimitOverrideResponse struct {
	ID          string `json:"id"`
	LimitCode   string `json:"limit_code"`
	LimitValue  int64  `json:"limit_value"`
	LimitWindow string `json:"limit_window"`
	Reason      string `json:"reason,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

type adminUserDetailResponse struct {
	User             adminUserSummary                   `json:"user"`
	PlanCode         string                             `json:"plan_code,omitempty"`
	UXMode           string                             `json:"ux_mode,omitempty"`
	AllowedModes     []string                           `json:"allowed_modes,omitempty"`
	Features         map[string]bool                    `json:"features,omitempty"`
	Limits           map[string]int64                   `json:"limits,omitempty"`
	FeatureOverrides []adminUserFeatureOverrideResponse `json:"feature_overrides,omitempty"`
	LimitOverrides   []adminUserLimitOverrideResponse   `json:"limit_overrides,omitempty"`
	Accounts         []adminAccountSummary              `json:"accounts"`
	TotalBalance     int64                              `json:"total_balance"`
}

type adminTimelineEntry struct {
	Type      string          `json:"type"`
	EventTime string          `json:"event_time"`
	Payload   json.RawMessage `json:"payload"`
}

func (h *Handler) handleAdminPlans(w http.ResponseWriter, r *http.Request) {
	if h.planRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "planos nao configurados", "PLANS_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		if !h.authorizeRoles(w, r, "ADMIN", "AUDIT") {
			return
		}
		items, err := h.planRepo.ListAll(r.Context())
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar planos", "PLANS_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		if !h.authorizeRoles(w, r, "ADMIN") {
			return
		}
		var payload adminPlanRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		code := strings.ToUpper(strings.TrimSpace(payload.Code))
		if code == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "code obrigatorio", "PLAN_CODE_REQUIRED")
			return
		}
		plan := &entity.Plan{
			ID:                uuid.NewString(),
			Code:              code,
			Description:       strings.TrimSpace(payload.Description),
			MonthlyPriceCents: payload.MonthlyPriceCents,
			CreatedAt:         time.Now().UTC(),
		}
		if err := h.planRepo.Create(r.Context(), plan); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao criar plano", "PLAN_CREATE_FAILED")
			return
		}
		reason := strings.TrimSpace(r.Header.Get("X-Reason"))
		var status string
		var validFrom time.Time
		var validUntil *time.Time
		if err := h.pool.QueryRow(r.Context(), `
			SELECT status, valid_from, valid_until
			  FROM plans
			 WHERE id = $1`, plan.ID).Scan(&status, &validFrom, &validUntil); err == nil {
			_, _ = h.pool.Exec(r.Context(), `
				INSERT INTO plan_history (
					plan_id, code, description, monthly_price_cents, status, valid_from, valid_until,
					change_type, change_reason, changed_by, changed_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, 'CREATE', NULLIF($8, ''), NULLIF($9, '')::uuid, NOW())`,
				plan.ID,
				plan.Code,
				nullableString(plan.Description),
				plan.MonthlyPriceCents,
				status,
				validFrom,
				validUntil,
				reason,
				userIDFromContext(r.Context()),
			)
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_CREATE", "plan", plan.ID, map[string]any{
			"code":         plan.Code,
			"monthly_price": plan.MonthlyPriceCents,
			"status":       status,
			"valid_from":   validFrom.UTC().Format(time.RFC3339Nano),
			"valid_until":  formatTimeOrEmpty(validUntil),
			"reason":       reason,
			"origin":       defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
		})
		writeJSON(r.Context(), w, http.StatusCreated, plan)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminPlanFeaturesRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.planFeatureRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "plan features nao configuradas", "PLAN_FEATURES_MISSING")
		return
	}
	var payload adminRollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	historyID := strings.TrimSpace(payload.HistoryID)
	if historyID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "history_id obrigatorio", "HISTORY_ID_REQUIRED")
		return
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		reason = historyID
	}
	var planID, featureCode string
	var enabled bool
	if err := h.pool.QueryRow(r.Context(), `
		SELECT plan_id, feature_code, enabled
		  FROM plan_feature_history
		 WHERE id = $1`, historyID).Scan(&planID, &featureCode, &enabled); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(r.Context(), w, http.StatusNotFound, "historico nao encontrado", "HISTORY_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao carregar historico", "HISTORY_LOOKUP_FAILED")
		return
	}
	item := &entity.PlanFeature{
		ID:          uuid.NewString(),
		PlanID:      planID,
		FeatureCode: featureCode,
		Enabled:     enabled,
	}
	if err := h.planFeatureRepo.Upsert(r.Context(), item); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao restaurar feature", "PLAN_FEATURE_ROLLBACK_FAILED")
		return
	}
	_, _ = h.pool.Exec(r.Context(), `
		INSERT INTO plan_feature_history (
			plan_id, feature_code, enabled, change_type, change_reason, changed_by, changed_at
		) VALUES ($1, $2, $3, 'ROLLBACK', $4, NULLIF($5, '')::uuid, NOW())`,
		planID,
		featureCode,
		enabled,
		reason,
		userIDFromContext(r.Context()),
	)
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_FEATURE_ROLLBACK", "plan_feature", planID, map[string]any{
		"history_id":  historyID,
		"feature_code": featureCode,
		"enabled":     enabled,
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *Handler) handleAdminPlanLimitsRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.planLimitRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "plan limits nao configurados", "PLAN_LIMITS_MISSING")
		return
	}
	var payload adminRollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	historyID := strings.TrimSpace(payload.HistoryID)
	if historyID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "history_id obrigatorio", "HISTORY_ID_REQUIRED")
		return
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		reason = historyID
	}
	var planID, limitCode, limitWindow string
	var limitValue int64
	if err := h.pool.QueryRow(r.Context(), `
		SELECT plan_id, limit_code, limit_value, limit_window
		  FROM plan_limit_history
		 WHERE id = $1`, historyID).Scan(&planID, &limitCode, &limitValue, &limitWindow); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(r.Context(), w, http.StatusNotFound, "historico nao encontrado", "HISTORY_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao carregar historico", "HISTORY_LOOKUP_FAILED")
		return
	}
	item := &entity.PlanLimit{
		ID:          uuid.NewString(),
		PlanID:      planID,
		LimitCode:   limitCode,
		LimitValue:  limitValue,
		LimitWindow: limitWindow,
	}
	if err := h.planLimitRepo.Upsert(r.Context(), item); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao restaurar limite", "PLAN_LIMIT_ROLLBACK_FAILED")
		return
	}
	_, _ = h.pool.Exec(r.Context(), `
		INSERT INTO plan_limit_history (
			plan_id, limit_code, limit_value, limit_window, change_type, change_reason, changed_by, changed_at
		) VALUES ($1, $2, $3, $4, 'ROLLBACK', $5, NULLIF($6, '')::uuid, NOW())`,
		planID,
		limitCode,
		limitValue,
		limitWindow,
		reason,
		userIDFromContext(r.Context()),
	)
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_LIMIT_ROLLBACK", "plan_limit", planID, map[string]any{
		"history_id":  historyID,
		"limit_code":  limitCode,
		"limit_window": limitWindow,
		"limit_value": limitValue,
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *Handler) handleAdminPlansRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	var payload adminRollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	historyID := strings.TrimSpace(payload.HistoryID)
	if historyID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "history_id obrigatorio", "HISTORY_ID_REQUIRED")
		return
	}
	reason := strings.TrimSpace(payload.Reason)
	if reason == "" {
		reason = historyID
	}
	var planID, code, status string
	var description *string
	var monthlyPrice int64
	var validFrom time.Time
	var validUntil *time.Time
	if err := h.pool.QueryRow(r.Context(), `
		SELECT plan_id, code, description, monthly_price_cents, status, valid_from, valid_until
		  FROM plan_history
		 WHERE id = $1`, historyID).Scan(&planID, &code, &description, &monthlyPrice, &status, &validFrom, &validUntil); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(r.Context(), w, http.StatusNotFound, "historico nao encontrado", "HISTORY_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao carregar historico", "HISTORY_LOOKUP_FAILED")
		return
	}
	_, err := h.pool.Exec(r.Context(), `
		UPDATE plans
		   SET code = $1,
		       description = $2,
		       monthly_price_cents = $3,
		       status = $4,
		       valid_from = $5,
		       valid_until = $6
		 WHERE id = $7`,
		code,
		description,
		monthlyPrice,
		status,
		validFrom,
		validUntil,
		planID,
	)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao restaurar plano", "PLAN_ROLLBACK_FAILED")
		return
	}
	_, _ = h.pool.Exec(r.Context(), `
		INSERT INTO plan_history (
			plan_id, code, description, monthly_price_cents, status, valid_from, valid_until,
			change_type, change_reason, changed_by, changed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'ROLLBACK', $8, NULLIF($9, '')::uuid, NOW())`,
		planID,
		code,
		description,
		monthlyPrice,
		status,
		validFrom,
		validUntil,
		reason,
		userIDFromContext(r.Context()),
	)
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_ROLLBACK", "plan", planID, map[string]any{
		"history_id": historyID,
		"code":       code,
		"status":     status,
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *Handler) handleAdminPlanHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	planID := strings.TrimSpace(r.URL.Query().Get("plan_id"))
	if planID != "" {
		if _, err := uuid.Parse(planID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "plan_id invalido", "INVALID_PLAN_ID")
			return
		}
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	if limit > 500 {
		limit = 500
	}
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, plan_id, code, description, monthly_price_cents, status, valid_from, valid_until,
		       change_type, change_reason, changed_by, changed_at
		  FROM plan_history
		 WHERE ($1::uuid IS NULL OR plan_id = $1)
		 ORDER BY changed_at DESC
		 LIMIT $2`, nullIfEmpty(planID), limit)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar historico", "PLAN_HISTORY_LIST_FAILED")
		return
	}
	defer rows.Close()

	items := make([]map[string]any, 0, limit)
	for rows.Next() {
		var id, planIDValue, code, status, changeType string
		var description *string
		var monthlyPrice int64
		var validFrom time.Time
		var validUntil *time.Time
		var changeReason *string
		var changedBy *string
		var changedAt time.Time
		if err := rows.Scan(
			&id,
			&planIDValue,
			&code,
			&description,
			&monthlyPrice,
			&status,
			&validFrom,
			&validUntil,
			&changeType,
			&changeReason,
			&changedBy,
			&changedAt,
		); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao ler historico", "PLAN_HISTORY_SCAN_FAILED")
			return
		}
		entry := map[string]any{
			"id":                  id,
			"plan_id":             planIDValue,
			"code":                code,
			"monthly_price_cents": monthlyPrice,
			"status":              status,
			"valid_from":          validFrom.UTC().Format(time.RFC3339Nano),
			"change_type":         changeType,
			"changed_at":          changedAt.UTC().Format(time.RFC3339Nano),
		}
		if description != nil {
			entry["description"] = *description
		}
		if validUntil != nil {
			entry["valid_until"] = validUntil.UTC().Format(time.RFC3339Nano)
		}
		if changeReason != nil {
			entry["change_reason"] = *changeReason
		}
		if changedBy != nil {
			entry["changed_by"] = *changedBy
		}
		items = append(items, entry)
	}
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar historico", "PLAN_HISTORY_LIST_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_HISTORY_LIST", "plan_history", "", map[string]any{
		"plan_id": defaultIfEmpty(planID, "all"),
		"limit":   limit,
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) handleAdminPlanFeatureHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	planID := strings.TrimSpace(r.URL.Query().Get("plan_id"))
	if planID != "" {
		if _, err := uuid.Parse(planID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "plan_id invalido", "INVALID_PLAN_ID")
			return
		}
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	if limit > 500 {
		limit = 500
	}
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, plan_id, feature_code, enabled, change_type, change_reason, changed_by, changed_at
		  FROM plan_feature_history
		 WHERE ($1::uuid IS NULL OR plan_id = $1)
		 ORDER BY changed_at DESC
		 LIMIT $2`, nullIfEmpty(planID), limit)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar historico", "PLAN_FEATURE_HISTORY_LIST_FAILED")
		return
	}
	defer rows.Close()

	items := make([]map[string]any, 0, limit)
	for rows.Next() {
		var id, planIDValue, featureCode, changeType string
		var enabled bool
		var changeReason *string
		var changedBy *string
		var changedAt time.Time
		if err := rows.Scan(
			&id,
			&planIDValue,
			&featureCode,
			&enabled,
			&changeType,
			&changeReason,
			&changedBy,
			&changedAt,
		); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao ler historico", "PLAN_FEATURE_HISTORY_SCAN_FAILED")
			return
		}
		entry := map[string]any{
			"id":           id,
			"plan_id":      planIDValue,
			"feature_code": featureCode,
			"enabled":      enabled,
			"change_type":  changeType,
			"changed_at":   changedAt.UTC().Format(time.RFC3339Nano),
		}
		if changeReason != nil {
			entry["change_reason"] = *changeReason
		}
		if changedBy != nil {
			entry["changed_by"] = *changedBy
		}
		items = append(items, entry)
	}
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar historico", "PLAN_FEATURE_HISTORY_LIST_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_FEATURE_HISTORY_LIST", "plan_feature_history", "", map[string]any{
		"plan_id": defaultIfEmpty(planID, "all"),
		"limit":   limit,
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) handleAdminPlanLimitHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	planID := strings.TrimSpace(r.URL.Query().Get("plan_id"))
	if planID != "" {
		if _, err := uuid.Parse(planID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "plan_id invalido", "INVALID_PLAN_ID")
			return
		}
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	if limit > 500 {
		limit = 500
	}
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, plan_id, limit_code, limit_value, limit_window, change_type, change_reason, changed_by, changed_at
		  FROM plan_limit_history
		 WHERE ($1::uuid IS NULL OR plan_id = $1)
		 ORDER BY changed_at DESC
		 LIMIT $2`, nullIfEmpty(planID), limit)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar historico", "PLAN_LIMIT_HISTORY_LIST_FAILED")
		return
	}
	defer rows.Close()

	items := make([]map[string]any, 0, limit)
	for rows.Next() {
		var id, planIDValue, limitCode, limitWindow, changeType string
		var limitValue int64
		var changeReason *string
		var changedBy *string
		var changedAt time.Time
		if err := rows.Scan(
			&id,
			&planIDValue,
			&limitCode,
			&limitValue,
			&limitWindow,
			&changeType,
			&changeReason,
			&changedBy,
			&changedAt,
		); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao ler historico", "PLAN_LIMIT_HISTORY_SCAN_FAILED")
			return
		}
		entry := map[string]any{
			"id":           id,
			"plan_id":      planIDValue,
			"limit_code":   limitCode,
			"limit_value":  limitValue,
			"limit_window": limitWindow,
			"change_type":  changeType,
			"changed_at":   changedAt.UTC().Format(time.RFC3339Nano),
		}
		if changeReason != nil {
			entry["change_reason"] = *changeReason
		}
		if changedBy != nil {
			entry["changed_by"] = *changedBy
		}
		items = append(items, entry)
	}
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar historico", "PLAN_LIMIT_HISTORY_LIST_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_LIMIT_HISTORY_LIST", "plan_limit_history", "", map[string]any{
		"plan_id": defaultIfEmpty(planID, "all"),
		"limit":   limit,
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) handleAdminUserPlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.userPlanRepo == nil || h.planRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "planos nao configurados", "PLAN_ASSIGN_MISSING")
		return
	}
	var payload adminUserPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	userID := strings.TrimSpace(payload.UserID)
	if userID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "user_id obrigatorio", "USER_ID_REQUIRED")
		return
	}
	var previousPlanCode string
	if h.userPlanRepo != nil && h.planRepo != nil {
		if active, err := h.userPlanRepo.GetActiveByUser(r.Context(), userID, time.Now().UTC()); err == nil && active != nil {
			if plan, err := h.planRepo.GetByID(r.Context(), active.PlanID); err == nil && plan != nil {
				previousPlanCode = plan.Code
			}
		}
	}
	planID := strings.TrimSpace(payload.PlanID)
	if planID == "" && payload.PlanCode != "" {
		plan, err := h.planRepo.GetByCode(r.Context(), strings.ToUpper(strings.TrimSpace(payload.PlanCode)))
		if err != nil || plan == nil {
			writeError(r.Context(), w, http.StatusBadRequest, "plan_code invalido", "PLAN_NOT_FOUND")
			return
		}
		planID = plan.ID
	}
	if planID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "plan_id ou plan_code obrigatorio", "PLAN_REQUIRED")
		return
	}
	validFrom := time.Now().UTC()
	if payload.ValidFrom != "" {
		parsed, err := time.Parse(time.RFC3339, payload.ValidFrom)
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "valid_from invalido", "INVALID_VALID_FROM")
			return
		}
		validFrom = parsed
	}
	var validUntil *time.Time
	if payload.ValidUntil != "" {
		parsed, err := time.Parse(time.RFC3339, payload.ValidUntil)
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "valid_until invalido", "INVALID_VALID_UNTIL")
			return
		}
		validUntil = &parsed
	}
	item := &entity.UserPlan{
		UserID:     userID,
		PlanID:     planID,
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
		CreatedAt:  time.Now().UTC(),
	}
	if err := h.userPlanRepo.Create(r.Context(), item); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atribuir plano", "PLAN_ASSIGN_FAILED")
		return
	}
	if h.planRepo != nil {
		if plan, err := h.planRepo.GetByID(r.Context(), planID); err == nil && plan != nil {
			_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_CHANGE", "user_plan", userID, map[string]any{
				"from":   previousPlanCode,
				"to":     plan.Code,
				"origin": defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
			})
		}
	}
	writeJSON(r.Context(), w, http.StatusCreated, item)
}

func (h *Handler) handleAdminPricingRules(w http.ResponseWriter, r *http.Request) {
	if h.pricingRuleRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pricing nao configurado", "PRICING_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		if !h.authorizeRoles(w, r, "ADMIN", "AUDIT") {
			return
		}
		filter := repository.PricingRuleFilter{
			PlanID:        strings.TrimSpace(r.URL.Query().Get("plan_id")),
			UserType:      entity.UserType(strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("user_type")))),
			OperationType: entity.PricingOperationType(strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("operation_type")))),
			Asset:         strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("asset"))),
			PricingVersionID: strings.TrimSpace(r.URL.Query().Get("pricing_version_id")),
		}
		items, err := h.pricingRuleRepo.List(r.Context(), filter)
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar pricing rules", "PRICING_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		if !h.authorizeRoles(w, r, "ADMIN") {
			return
		}
		var payload adminPricingRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		rule, err := pricingRuleFromPayload(payload, "")
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "PRICING_RULE_INVALID")
			return
		}
		rule.ID = uuid.NewString()
		rule.CreatedAt = time.Now().UTC()
		if err := h.pricingRuleRepo.Create(r.Context(), rule); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao criar pricing rule", "PRICING_CREATE_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PRICING_RULE_CREATE", "pricing_rule", rule.ID, map[string]any{
			"plan_id":        rule.PlanID,
			"user_type":      string(rule.UserType),
			"operation_type": string(rule.OperationType),
			"asset":          rule.Asset,
			"origin":         defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
		})
		writeJSON(r.Context(), w, http.StatusCreated, rule)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminPricingRuleByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if !h.authorizeRoles(w, r, "ADMIN") {
		return
	}
	if h.pricingRuleRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pricing nao configurado", "PRICING_MISSING")
		return
	}
	ruleID := strings.TrimPrefix(r.URL.Path, "/admin/pricing/rules/")
	ruleID = strings.TrimSpace(ruleID)
	if ruleID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "id obrigatorio", "RULE_ID_REQUIRED")
		return
	}
	var payload adminPricingRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	rule, err := pricingRuleFromPayload(payload, ruleID)
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "PRICING_RULE_INVALID")
		return
	}
	if err := h.pricingRuleRepo.Update(r.Context(), rule); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "pricing rule nao encontrada", "PRICING_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar pricing rule", "PRICING_UPDATE_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PRICING_RULE_UPDATE", "pricing_rule", rule.ID, map[string]any{
		"plan_id":        rule.PlanID,
		"user_type":      string(rule.UserType),
		"operation_type": string(rule.OperationType),
		"asset":          rule.Asset,
		"origin":         defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
	})
	writeJSON(r.Context(), w, http.StatusOK, rule)
}

func (h *Handler) handleAdminRoles(w http.ResponseWriter, r *http.Request) {
	if h.roleRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLES_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := h.roleRepo.ListAll(r.Context())
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar roles", "ROLES_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var payload adminRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		code := strings.ToUpper(strings.TrimSpace(payload.Code))
		if code == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "code obrigatorio", "ROLE_CODE_REQUIRED")
			return
		}
		role := &entity.Role{
			Code:        code,
			Description: strings.TrimSpace(payload.Description),
			CreatedAt:   time.Now().UTC(),
		}
		if err := h.roleRepo.Create(r.Context(), role); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao criar role", "ROLE_CREATE_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "ROLE_CREATE", "role", role.Code, role)
		writeJSON(r.Context(), w, http.StatusCreated, role)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminUserRoles(w http.ResponseWriter, r *http.Request) {
	if h.userRoleRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "user roles nao configuradas", "USER_ROLES_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
		if userID == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id obrigatorio", "USER_ID_REQUIRED")
			return
		}
		if _, err := uuid.Parse(userID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id invalido", "INVALID_USER_ID")
			return
		}
		items, err := h.userRoleRepo.ListByUser(r.Context(), userID)
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar user roles", "USER_ROLES_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var payload adminUserRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		userID := strings.TrimSpace(payload.UserID)
		roleCode := strings.ToUpper(strings.TrimSpace(payload.RoleCode))
		if userID == "" || roleCode == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id e role_code obrigatorios", "USER_ROLE_REQUIRED")
			return
		}
		if _, err := uuid.Parse(userID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id invalido", "INVALID_USER_ID")
			return
		}
		if h.roleSeparationRepo != nil {
			conflict, err := h.roleSeparationRepo.HasConflict(r.Context(), userID, roleCode)
			if err != nil {
				writeError(r.Context(), w, http.StatusInternalServerError, "falha ao validar separacao", "ROLE_SEPARATION_CHECK_FAILED")
				return
			}
			if conflict {
				writeError(r.Context(), w, http.StatusConflict, "separation of duties violada", "ROLE_SEPARATION_VIOLATION")
				return
			}
		}
		item := &entity.UserRole{
			UserID:    userID,
			RoleCode:  roleCode,
			GrantedBy: userIDFromContext(r.Context()),
			GrantedAt: time.Now().UTC(),
		}
		if err := h.userRoleRepo.Assign(r.Context(), item); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atribuir role", "USER_ROLE_ASSIGN_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "USER_ROLE_ASSIGN", "user_role", userID, item)
		writeJSON(r.Context(), w, http.StatusCreated, item)
	case http.MethodDelete:
		userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
		roleCode := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("role_code")))
		if userID == "" || roleCode == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id e role_code obrigatorios", "USER_ROLE_REQUIRED")
			return
		}
		if _, err := uuid.Parse(userID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id invalido", "INVALID_USER_ID")
			return
		}
		if err := h.userRoleRepo.Remove(r.Context(), userID, roleCode); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao remover role", "USER_ROLE_REMOVE_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "USER_ROLE_REMOVE", "user_role", userID, map[string]any{
			"user_id":   userID,
			"role_code": roleCode,
		})
		writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "REMOVED"})
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminRoleSeparation(w http.ResponseWriter, r *http.Request) {
	if h.roleSeparationRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "separation rules nao configuradas", "ROLE_SEPARATION_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := h.roleSeparationRepo.ListAll(r.Context())
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar regras", "ROLE_SEPARATION_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var payload adminRoleSeparationRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		roleA := strings.ToUpper(strings.TrimSpace(payload.RoleCodeA))
		roleB := strings.ToUpper(strings.TrimSpace(payload.RoleCodeB))
		if roleA == "" || roleB == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "role_code_a e role_code_b obrigatorios", "ROLE_SEPARATION_REQUIRED")
			return
		}
		if roleA == roleB {
			writeError(r.Context(), w, http.StatusBadRequest, "roles devem ser diferentes", "ROLE_SEPARATION_INVALID")
			return
		}
		rule := &entity.RoleSeparationRule{
			ID:        uuid.NewString(),
			RoleCodeA: roleA,
			RoleCodeB: roleB,
			Reason:    strings.TrimSpace(payload.Reason),
			CreatedAt: time.Now().UTC(),
		}
		if err := h.roleSeparationRepo.Create(r.Context(), rule); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao criar regra", "ROLE_SEPARATION_CREATE_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "ROLE_SEPARATION_CREATE", "role_separation", rule.ID, rule)
		writeJSON(r.Context(), w, http.StatusCreated, rule)
	case http.MethodDelete:
		roleA := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("role_code_a")))
		roleB := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("role_code_b")))
		if roleA == "" || roleB == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "role_code_a e role_code_b obrigatorios", "ROLE_SEPARATION_REQUIRED")
			return
		}
		if err := h.roleSeparationRepo.Remove(r.Context(), roleA, roleB); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao remover regra", "ROLE_SEPARATION_REMOVE_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "ROLE_SEPARATION_REMOVE", "role_separation", "", map[string]any{
			"role_code_a": roleA,
			"role_code_b": roleB,
		})
		writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "REMOVED"})
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminRegulatoryProfiles(w http.ResponseWriter, r *http.Request) {
	if h.regulatoryProfileRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "regulatory profiles nao configurados", "REGULATORY_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
		if userID == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id obrigatorio", "USER_ID_REQUIRED")
			return
		}
		if _, err := uuid.Parse(userID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id invalido", "INVALID_USER_ID")
			return
		}
		profile, err := h.regulatoryProfileRepo.GetByUserID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				writeError(r.Context(), w, http.StatusNotFound, "perfil nao encontrado", "REGULATORY_NOT_FOUND")
				return
			}
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar perfil", "REGULATORY_LOOKUP_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, profile)
	case http.MethodPut:
		var payload adminRegulatoryProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		userID := strings.TrimSpace(payload.UserID)
		if userID == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id obrigatorio", "USER_ID_REQUIRED")
			return
		}
		if _, err := uuid.Parse(userID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id invalido", "INVALID_USER_ID")
			return
		}
		jurisdiction := strings.ToUpper(strings.TrimSpace(payload.JurisdictionCode))
		if jurisdiction == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "jurisdiction_code obrigatorio", "JURISDICTION_REQUIRED")
			return
		}
		risk := strings.ToUpper(strings.TrimSpace(payload.JurisdictionRisk))
		if risk == "" {
			risk = string(entity.JurisdictionRiskLow)
		}
		amlTier := strings.ToUpper(strings.TrimSpace(payload.AMLTier))
		if amlTier == "" {
			amlTier = string(entity.AMLTierBasic)
		}
		profile := &entity.RegulatoryProfile{
			ID:                         uuid.NewString(),
			UserID:                     userID,
			JurisdictionCode:           jurisdiction,
			JurisdictionRisk:           entity.JurisdictionRisk(risk),
			AMLTier:                    entity.AMLTier(amlTier),
			TravelRuleRequired:         payload.TravelRuleRequired,
			SanctionsScreeningRequired: payload.SanctionsScreeningRequired,
		}
		if err := h.regulatoryProfileRepo.Upsert(r.Context(), profile); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao salvar perfil", "REGULATORY_UPSERT_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "REGULATORY_PROFILE_UPSERT", "regulatory_profile", userID, profile)
		writeJSON(r.Context(), w, http.StatusOK, profile)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminReconcileSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	cutoff := reconcileCutoff(r)
	summary := map[string]any{
		"cutoff": cutoff.Format(time.RFC3339Nano),
	}
	if h.pixRepo != nil {
		count, err := h.pixRepo.CountPendingBefore(r.Context(), cutoff)
		if err == nil {
			summary["pix_pending"] = count
		}
	}
	if h.paymentRepo != nil {
		count, err := h.paymentRepo.CountPendingBefore(r.Context(), cutoff)
		if err == nil {
			summary["payments_pending"] = count
		}
	}
	if h.cardRepo != nil {
		count, err := h.cardRepo.CountPendingBefore(r.Context(), cutoff)
		if err == nil {
			summary["card_pending"] = count
		}
	}
	if h.txRepo != nil {
		count, err := h.txRepo.CountHoldBefore(r.Context(), cutoff)
		if err == nil {
			summary["transactions_hold"] = count
		}
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "RECONCILE_SUMMARY", "reconcile", "", summary)
	writeJSON(r.Context(), w, http.StatusOK, summary)
}

func (h *Handler) handleAdminReconcilePending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	kind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	cutoff := reconcileCutoff(r)

	response := map[string]any{
		"type":   defaultIfEmpty(kind, "all"),
		"limit":  limit,
		"cutoff": cutoff.Format(time.RFC3339Nano),
	}
	switch kind {
	case "pix":
		response["items"] = h.listPendingPix(r, cutoff, limit)
	case "payment", "payments":
		response["items"] = h.listPendingPayments(r, cutoff, limit)
	case "card":
		response["items"] = h.listPendingCards(r, cutoff, limit)
	case "transaction", "transactions":
		response["items"] = h.listPendingTransactions(r, cutoff, limit)
	default:
		response["pix"] = h.listPendingPix(r, cutoff, limit)
		response["payments"] = h.listPendingPayments(r, cutoff, limit)
		response["card"] = h.listPendingCards(r, cutoff, limit)
		response["transactions"] = h.listPendingTransactions(r, cutoff, limit)
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "RECONCILE_LIST", "reconcile", "", response)
	writeJSON(r.Context(), w, http.StatusOK, response)
}

func (h *Handler) handleAdminWebhookRetryList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.webhookRetryRepo == nil {
		writeError(r.Context(), w, http.StatusNotFound, "retry nao habilitado", "RETRY_NOT_AVAILABLE")
		return
	}
	statusRaw := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("status")))
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	status := entity.WebhookRetryStatus(statusRaw)
	if status == "" {
		status = entity.WebhookRetryPending
	}
	switch status {
	case entity.WebhookRetryPending, entity.WebhookRetryProcessing, entity.WebhookRetrySucceeded, entity.WebhookRetryDead:
	default:
		writeError(r.Context(), w, http.StatusBadRequest, "status invalido", "INVALID_STATUS")
		return
	}
	items, err := h.webhookRetryRepo.ListByStatus(r.Context(), status, limit)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar retries", "RETRY_LIST_FAILED")
		return
	}
	response := map[string]any{
		"status": status,
		"limit":  limit,
		"items":  mapRetryJobs(items),
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "WEBHOOK_RETRY_LIST", "webhook_retry", "", response)
	writeJSON(r.Context(), w, http.StatusOK, response)
}

func mapRetryJobs(items []*entity.WebhookRetryJob) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, map[string]any{
			"id":           item.ID,
			"event_type":   item.EventType,
			"path":         item.Path,
			"attempts":     item.Attempts,
			"status":       item.Status,
			"next_retry_at": item.NextRetryAt.UTC().Format(time.RFC3339Nano),
			"last_error":   item.LastError,
			"updated_at":   item.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return result
}

func (h *Handler) handleAdminObservabilitySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	cutoff := reconcileCutoff(r)
	summary := map[string]any{
		"cutoff": cutoff.Format(time.RFC3339Nano),
	}

	if h.pixRepo != nil {
		if count, err := h.pixRepo.CountPendingBefore(r.Context(), cutoff); err == nil {
			summary["pix_pending"] = count
		}
	}
	if h.paymentRepo != nil {
		if count, err := h.paymentRepo.CountPendingBefore(r.Context(), cutoff); err == nil {
			summary["payments_pending"] = count
		}
	}
	if h.cardRepo != nil {
		if count, err := h.cardRepo.CountPendingBefore(r.Context(), cutoff); err == nil {
			summary["card_pending"] = count
		}
	}
	if h.txRepo != nil {
		if count, err := h.txRepo.CountHoldBefore(r.Context(), cutoff); err == nil {
			summary["transactions_hold"] = count
		}
	}

	if h.webhookRetryRepo != nil {
		if count, err := h.webhookRetryRepo.CountByStatus(r.Context(), entity.WebhookRetryPending); err == nil {
			summary["webhook_retry_pending"] = count
		}
		if count, err := h.webhookRetryRepo.CountByStatus(r.Context(), entity.WebhookRetrySucceeded); err == nil {
			summary["webhook_retry_succeeded"] = count
		}
		if count, err := h.webhookRetryRepo.CountByStatus(r.Context(), entity.WebhookRetryDead); err == nil {
			summary["webhook_retry_dead"] = count
		}
	}

	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "OBSERVABILITY_SUMMARY", "observability", "", summary)
	writeJSON(r.Context(), w, http.StatusOK, summary)
}

func (h *Handler) handleAdminAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	action := strings.TrimSpace(r.URL.Query().Get("action"))
	entityType := strings.TrimSpace(r.URL.Query().Get("entity_type"))
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	from, err := parseOptionalTime(r.URL.Query().Get("from"))
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "from invalido", "INVALID_FROM")
		return
	}
	to, err := parseOptionalTime(r.URL.Query().Get("to"))
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "to invalido", "INVALID_TO")
		return
	}
	if userID != "" {
		if _, err := uuid.Parse(userID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "user_id invalido", "INVALID_USER_ID")
			return
		}
	}
	if limit > 500 {
		limit = 500
	}
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, user_id, action, entity_type, entity_id, data, created_at
		  FROM audit_logs
		 WHERE ($1::uuid IS NULL OR user_id = $1)
		   AND ($2 = '' OR action = $2)
		   AND ($3 = '' OR entity_type = $3)
		   AND ($4::timestamptz IS NULL OR created_at >= $4)
		   AND ($5::timestamptz IS NULL OR created_at <= $5)
		 ORDER BY created_at DESC
		 LIMIT $6`,
		nullIfEmpty(userID),
		action,
		entityType,
		from,
		to,
		limit,
	)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar auditoria", "AUDIT_LIST_FAILED")
		return
	}
	defer rows.Close()

	items := make([]map[string]any, 0, limit)
	for rows.Next() {
		var id, uid, act, eType string
		var entityID *string
		var data []byte
		var createdAt time.Time
		if err := rows.Scan(&id, &uid, &act, &eType, &entityID, &data, &createdAt); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao ler auditoria", "AUDIT_SCAN_FAILED")
			return
		}
		entry := map[string]any{
			"id":          id,
			"user_id":     uid,
			"action":      act,
			"entity_type": eType,
			"created_at":  createdAt.UTC().Format(time.RFC3339Nano),
		}
		if entityID != nil {
			entry["entity_id"] = *entityID
		}
		if len(data) > 0 && json.Valid(data) {
			entry["data"] = json.RawMessage(data)
		}
		items = append(items, entry)
	}
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar auditoria", "AUDIT_LIST_FAILED")
		return
	}

	response := map[string]any{
		"user_id":     defaultIfEmpty(userID, "all"),
		"action":      defaultIfEmpty(action, "all"),
		"entity_type": defaultIfEmpty(entityType, "all"),
		"from":        formatTimeOrEmpty(from),
		"to":          formatTimeOrEmpty(to),
		"limit":       limit,
		"items":       items,
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "AUDIT_LIST", "audit", "", response)
	writeJSON(r.Context(), w, http.StatusOK, response)
}

func (h *Handler) handleAdminAuditArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if !h.authorizeRoles(w, r, "ADMIN", "COMPLIANCE") {
		return
	}
	var payload adminAuditArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	olderThanDays := payload.OlderThanDays
	if olderThanDays < 30 {
		writeError(r.Context(), w, http.StatusBadRequest, "older_than_days minimo 30", "INVALID_OLDER_THAN_DAYS")
		return
	}
	limit := payload.Limit
	if limit <= 0 {
		limit = 1000
	}
	if limit > 50000 {
		limit = 50000
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -olderThanDays)

	if payload.DryRun {
		var count int64
		if err := h.pool.QueryRow(r.Context(), `
			SELECT COUNT(*) FROM audit_logs WHERE created_at < $1
		`, cutoff).Scan(&count); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao estimar arquivo", "AUDIT_ARCHIVE_DRY_RUN_FAILED")
			return
		}
		response := map[string]any{
			"cutoff":   cutoff.Format(time.RFC3339Nano),
			"limit":    limit,
			"dry_run":  true,
			"eligible": count,
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "AUDIT_ARCHIVE_DRY_RUN", "audit", "", response)
		writeJSON(r.Context(), w, http.StatusOK, response)
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao iniciar transacao", "AUDIT_ARCHIVE_BEGIN_FAILED")
		return
	}
	defer func() { _ = tx.Rollback(r.Context()) }()

	var inserted int64
	var deleted int64
	if err := tx.QueryRow(r.Context(), `
		WITH moved AS (
			SELECT id, user_id, action, entity_type, entity_id, data, created_at
			  FROM audit_logs
			 WHERE created_at < $1
			 ORDER BY created_at ASC
			 LIMIT $2
		),
		inserted AS (
			INSERT INTO audit_logs_archive (id, user_id, action, entity_type, entity_id, data, created_at)
			SELECT id, user_id, action, entity_type, entity_id, data, created_at FROM moved
			ON CONFLICT (id) DO NOTHING
			RETURNING id
		),
		deleted AS (
			DELETE FROM audit_logs WHERE id IN (SELECT id FROM inserted)
			RETURNING id
		)
		SELECT (SELECT COUNT(*) FROM inserted), (SELECT COUNT(*) FROM deleted)
	`, cutoff, limit).Scan(&inserted, &deleted); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao arquivar auditoria", "AUDIT_ARCHIVE_FAILED")
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao confirmar arquivo", "AUDIT_ARCHIVE_COMMIT_FAILED")
		return
	}

	response := map[string]any{
		"cutoff":   cutoff.Format(time.RFC3339Nano),
		"limit":    limit,
		"archived": inserted,
		"deleted":  deleted,
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "AUDIT_ARCHIVE_RUN", "audit", "", response)
	writeJSON(r.Context(), w, http.StatusOK, response)
}


func (h *Handler) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.userRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "usuarios nao configurados", "USERS_MISSING")
		return
	}
	userID := strings.TrimSpace(r.URL.Query().Get("id"))
	if userID == "" {
		userID = strings.TrimSpace(r.URL.Query().Get("user_id"))
	}
	email := strings.TrimSpace(r.URL.Query().Get("email"))
	document := strings.TrimSpace(r.URL.Query().Get("document"))
	if document == "" {
		document = strings.TrimSpace(r.URL.Query().Get("external_id"))
	}
	if userID == "" && email == "" && document == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "id, email ou documento obrigatorio", "FILTER_REQUIRED")
		return
	}
	if userID != "" {
		if _, err := uuid.Parse(userID); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "id invalido", "INVALID_USER_ID")
			return
		}
	}
	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "limit invalido", "INVALID_LIMIT")
		return
	}
	items, err := h.userRepo.Search(r.Context(), repository.UserSearchFilter{
		ID:         userID,
		Email:      email,
		ExternalID: document,
		Limit:      limit,
	})
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao buscar usuarios", "USER_SEARCH_FAILED")
		return
	}
	response := make([]adminUserSummary, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		response = append(response, adminUserSummaryFromEntity(item))
	}
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": response})
}

func (h *Handler) handleAdminUserByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/users/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(r.Context(), w, http.StatusNotFound, "endpoint nao encontrado", "NOT_FOUND")
		return
	}
	userID := parts[0]
	if _, err := uuid.Parse(userID); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "user_id invalido", "INVALID_USER_ID")
		return
	}
	if len(parts) == 1 {
		h.handleAdminUserDetail(w, r, userID)
		return
	}
	switch parts[1] {
	case "features":
		h.handleAdminUserFeatures(w, r, userID)
	case "limits":
		h.handleAdminUserLimits(w, r, userID)
	case "sessions":
		if len(parts) >= 3 && parts[2] == "reset" {
			h.handleAdminUserSessionsReset(w, r, userID)
			return
		}
		writeError(r.Context(), w, http.StatusNotFound, "endpoint nao encontrado", "NOT_FOUND")
	case "timeline":
		h.handleAdminUserTimeline(w, r, userID)
	default:
		writeError(r.Context(), w, http.StatusNotFound, "endpoint nao encontrado", "NOT_FOUND")
	}
}

func (h *Handler) handleAdminAccounts(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/accounts/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(r.Context(), w, http.StatusNotFound, "endpoint nao encontrado", "NOT_FOUND")
		return
	}
	accountID := parts[0]
	if _, err := uuid.Parse(accountID); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "account_id invalido", "INVALID_ACCOUNT_ID")
		return
	}
	if len(parts) == 1 {
		h.handleAdminAccountDetail(w, r, accountID)
		return
	}
	switch parts[1] {
	case "ledger":
		h.handleAdminAccountLedger(w, r, accountID)
	case "freeze":
		h.handleAdminAccountFreeze(w, r, accountID)
	default:
		writeError(r.Context(), w, http.StatusNotFound, "endpoint nao encontrado", "NOT_FOUND")
	}
}

func (h *Handler) handleAdminUserDetail(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.userRepo == nil || h.accountRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "repositorios nao configurados", "ADMIN_USER_MISSING")
		return
	}
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "usuario nao encontrado", "USER_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar usuario", "USER_LOOKUP_FAILED")
		return
	}
	accounts, err := h.accountRepo.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar contas", "ACCOUNT_LIST_FAILED")
		return
	}
	accountSummaries := make([]adminAccountSummary, 0, len(accounts))
	var totalBalance int64
	for _, account := range accounts {
		if account == nil {
			continue
		}
		balance := int64(0)
		if h.txRepo != nil {
			if value, err := h.txRepo.GetLedgerBalance(r.Context(), account.ID); err == nil {
				balance = value
			}
		}
		accountSummaries = append(accountSummaries, adminAccountSummaryFromEntity(account, balance))
		totalBalance += balance
	}
	snapshot := h.resolveCapabilitySnapshot(r.Context(), userID)

	featureOverrides := []adminUserFeatureOverrideResponse{}
	if h.userFeatureOverrideRepo != nil {
		if overrides, err := h.userFeatureOverrideRepo.ListByUser(r.Context(), userID); err == nil {
			for _, item := range overrides {
				if item == nil {
					continue
				}
				featureOverrides = append(featureOverrides, adminUserFeatureOverrideResponse{
					ID:          item.ID,
					FeatureCode: item.FeatureCode,
					Enabled:     item.Enabled,
					Reason:      item.Reason,
					UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339Nano),
				})
			}
		}
	}
	limitOverrides := []adminUserLimitOverrideResponse{}
	if h.userLimitOverrideRepo != nil {
		if overrides, err := h.userLimitOverrideRepo.ListByUser(r.Context(), userID); err == nil {
			for _, item := range overrides {
				if item == nil {
					continue
				}
				limitOverrides = append(limitOverrides, adminUserLimitOverrideResponse{
					ID:          item.ID,
					LimitCode:   item.LimitCode,
					LimitValue:  item.LimitValue,
					LimitWindow: item.LimitWindow,
					Reason:      item.Reason,
					UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339Nano),
				})
			}
		}
	}

	response := adminUserDetailResponse{
		User:             adminUserSummaryFromEntity(user),
		PlanCode:         snapshot.PlanCode,
		UXMode:           snapshot.UXMode,
		AllowedModes:     snapshot.AllowedModes,
		Features:         snapshot.Features,
		Limits:           snapshot.Limits,
		FeatureOverrides: featureOverrides,
		LimitOverrides:   limitOverrides,
		Accounts:         accountSummaries,
		TotalBalance:     totalBalance,
	}
	writeJSON(r.Context(), w, http.StatusOK, response)
}

func (h *Handler) handleAdminUserFeatures(w http.ResponseWriter, r *http.Request, userID string) {
	if h.userFeatureOverrideRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "override nao configurado", "USER_FEATURE_OVERRIDE_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		if err := h.ensureRole(r.Context(), "ADMIN", "COMPLIANCE", "AUDIT"); err != nil {
			switch {
			case errors.Is(err, errRoleRepoMissing):
				writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLE_REPO_MISSING")
			case errors.Is(err, errRoleForbidden):
				writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "ROLE_FORBIDDEN")
			default:
				writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar roles", "ROLE_LOOKUP_FAILED")
			}
			return
		}
		items, err := h.userFeatureOverrideRepo.ListByUser(r.Context(), userID)
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar overrides", "USER_FEATURE_OVERRIDE_LIST_FAILED")
			return
		}
		response := make([]adminUserFeatureOverrideResponse, 0, len(items))
		for _, item := range items {
			if item == nil {
				continue
			}
			response = append(response, adminUserFeatureOverrideResponse{
				ID:          item.ID,
				FeatureCode: item.FeatureCode,
				Enabled:     item.Enabled,
				Reason:      item.Reason,
				UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339Nano),
			})
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": response})
	case http.MethodPut:
		if err := h.ensureRole(r.Context(), "ADMIN", "COMPLIANCE"); err != nil {
			switch {
			case errors.Is(err, errRoleRepoMissing):
				writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLE_REPO_MISSING")
			case errors.Is(err, errRoleForbidden):
				writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "ROLE_FORBIDDEN")
			default:
				writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar roles", "ROLE_LOOKUP_FAILED")
			}
			return
		}
		var payload adminUserFeatureRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		code := strings.ToUpper(strings.TrimSpace(payload.FeatureCode))
		if code == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "feature_code obrigatorio", "FEATURE_CODE_REQUIRED")
			return
		}
		now := time.Now().UTC()
		override := &entity.UserFeatureOverride{
			ID:          uuid.NewString(),
			UserID:      userID,
			FeatureCode: code,
			Enabled:     payload.Enabled,
			Reason:      strings.TrimSpace(payload.Reason),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := h.userFeatureOverrideRepo.Upsert(r.Context(), override); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao salvar override", "USER_FEATURE_OVERRIDE_UPSERT_FAILED")
			return
		}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "USER_FEATURE_OVERRIDE_UPSERT", "user_feature_override", override.ID, map[string]any{
		"user_id":      userID,
		"feature_code": code,
		"enabled":      override.Enabled,
		"reason":       override.Reason,
		"origin":       defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
	})
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"status": "ok"})
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminUserLimits(w http.ResponseWriter, r *http.Request, userID string) {
	if h.userLimitOverrideRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "override nao configurado", "USER_LIMIT_OVERRIDE_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		if err := h.ensureRole(r.Context(), "ADMIN", "COMPLIANCE", "AUDIT"); err != nil {
			switch {
			case errors.Is(err, errRoleRepoMissing):
				writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLE_REPO_MISSING")
			case errors.Is(err, errRoleForbidden):
				writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "ROLE_FORBIDDEN")
			default:
				writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar roles", "ROLE_LOOKUP_FAILED")
			}
			return
		}
		items, err := h.userLimitOverrideRepo.ListByUser(r.Context(), userID)
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar overrides", "USER_LIMIT_OVERRIDE_LIST_FAILED")
			return
		}
		response := make([]adminUserLimitOverrideResponse, 0, len(items))
		for _, item := range items {
			if item == nil {
				continue
			}
			response = append(response, adminUserLimitOverrideResponse{
				ID:          item.ID,
				LimitCode:   item.LimitCode,
				LimitValue:  item.LimitValue,
				LimitWindow: item.LimitWindow,
				Reason:      item.Reason,
				UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339Nano),
			})
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": response})
	case http.MethodPut:
		var payload adminUserLimitRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		code := strings.ToUpper(strings.TrimSpace(payload.LimitCode))
		if code == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "limit_code obrigatorio", "LIMIT_CODE_REQUIRED")
			return
		}
		if payload.LimitValue < 0 {
			writeError(r.Context(), w, http.StatusBadRequest, "limit_value invalido", "INVALID_LIMIT_VALUE")
			return
		}
		window := strings.ToUpper(strings.TrimSpace(payload.LimitWindow))
		if window == "" {
			window = "MONTHLY"
		}
		now := time.Now().UTC()
		override := &entity.UserLimitOverride{
			ID:          uuid.NewString(),
			UserID:      userID,
			LimitCode:   code,
			LimitValue:  payload.LimitValue,
			LimitWindow: window,
			Reason:      strings.TrimSpace(payload.Reason),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := h.userLimitOverrideRepo.Upsert(r.Context(), override); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao salvar override", "USER_LIMIT_OVERRIDE_UPSERT_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "USER_LIMIT_OVERRIDE_UPSERT", "user_limit_override", override.ID, map[string]any{
			"user_id":      userID,
			"limit_code":   code,
			"limit_value":  override.LimitValue,
			"limit_window": window,
			"reason":       override.Reason,
		"origin":       defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
		})
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"status": "ok"})
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminUserSessionsReset(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if err := h.ensureRole(r.Context(), "ADMIN", "COMPLIANCE", "OPS"); err != nil {
		switch {
		case errors.Is(err, errRoleRepoMissing):
			writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLE_REPO_MISSING")
		case errors.Is(err, errRoleForbidden):
			writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "ROLE_FORBIDDEN")
		default:
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar roles", "ROLE_LOOKUP_FAILED")
		}
		return
	}
	if h.refreshTokenRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "tokens nao configurados", "REFRESH_TOKEN_MISSING")
		return
	}
	count, err := h.refreshTokenRepo.RevokeAllByUser(r.Context(), userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao resetar sessoes", "SESSION_RESET_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "USER_SESSIONS_RESET", "user", userID, map[string]any{
		"revoked": count,
		"origin":  defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{"revoked": count})
}

func (h *Handler) handleAdminUserTimeline(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if err := h.ensureRole(r.Context(), "ADMIN", "COMPLIANCE", "AUDIT", "OPS"); err != nil {
		switch {
		case errors.Is(err, errRoleRepoMissing):
			writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLE_REPO_MISSING")
		case errors.Is(err, errRoleForbidden):
			writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "ROLE_FORBIDDEN")
		default:
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar roles", "ROLE_LOOKUP_FAILED")
		}
		return
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	if limit > 500 {
		limit = 500
	}
	rows, err := h.pool.Query(r.Context(), `
		SELECT entry_type, event_time, payload
		  FROM (
		    SELECT 'audit'::text AS entry_type,
		           created_at AS event_time,
		           jsonb_build_object(
		             'id', id,
		             'action', action,
		             'entity_type', entity_type,
		             'entity_id', entity_id,
		             'data', data
		           ) AS payload
		      FROM audit_logs
		     WHERE user_id = $1
		    UNION ALL
		    SELECT 'transaction'::text AS entry_type,
		           occurred_at AS event_time,
		           jsonb_build_object(
		             'id', id,
		             'account_id', account_id,
		             'type', type,
		             'status', status,
		             'amount', amount,
		             'fee', fee,
		             'net_amount', net_amount
		           ) AS payload
		      FROM transactions
		     WHERE user_id = $1
		  ) timeline
		 ORDER BY event_time DESC
		 LIMIT $2`, userID, limit)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar timeline", "TIMELINE_LIST_FAILED")
		return
	}
	defer rows.Close()

	items := make([]adminTimelineEntry, 0, limit)
	for rows.Next() {
		var entryType string
		var eventTime time.Time
		var payload []byte
		if err := rows.Scan(&entryType, &eventTime, &payload); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao ler timeline", "TIMELINE_SCAN_FAILED")
			return
		}
		if !json.Valid(payload) {
			payload = []byte(`{}`)
		}
		items = append(items, adminTimelineEntry{
			Type:      entryType,
			EventTime: eventTime.UTC().Format(time.RFC3339Nano),
			Payload:   json.RawMessage(payload),
		})
	}
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar timeline", "TIMELINE_LIST_FAILED")
		return
	}
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) handleAdminAccountDetail(w http.ResponseWriter, r *http.Request, accountID string) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if err := h.ensureRole(r.Context(), "ADMIN", "COMPLIANCE", "AUDIT", "OPS", "VIEWER"); err != nil {
		switch {
		case errors.Is(err, errRoleRepoMissing):
			writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLE_REPO_MISSING")
		case errors.Is(err, errRoleForbidden):
			writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "ROLE_FORBIDDEN")
		default:
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar roles", "ROLE_LOOKUP_FAILED")
		}
		return
	}
	if h.accountRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "contas nao configuradas", "ACCOUNT_MISSING")
		return
	}
	account, err := h.accountRepo.GetByID(r.Context(), accountID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "conta nao encontrada", "ACCOUNT_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar conta", "ACCOUNT_LOOKUP_FAILED")
		return
	}
	balance := int64(0)
	if h.txRepo != nil {
		if value, err := h.txRepo.GetLedgerBalance(r.Context(), accountID); err == nil {
			balance = value
		}
	}
	writeJSON(r.Context(), w, http.StatusOK, adminAccountSummaryFromEntity(account, balance))
}

func (h *Handler) handleAdminAccountLedger(w http.ResponseWriter, r *http.Request, accountID string) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if err := h.ensureRole(r.Context(), "ADMIN", "COMPLIANCE", "AUDIT", "OPS", "VIEWER"); err != nil {
		switch {
		case errors.Is(err, errRoleRepoMissing):
			writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLE_REPO_MISSING")
		case errors.Is(err, errRoleForbidden):
			writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "ROLE_FORBIDDEN")
		default:
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar roles", "ROLE_LOOKUP_FAILED")
		}
		return
	}
	if h.txRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "ledger nao configurado", "LEDGER_MISSING")
		return
	}
	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "limit invalido", "INVALID_LIMIT")
		return
	}
	from, err := parseTimeParam(r.URL.Query().Get("from"))
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "from invalido", "INVALID_FROM")
		return
	}
	to, err := parseTimeParam(r.URL.Query().Get("to"))
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "to invalido", "INVALID_TO")
		return
	}
	cursorAt, cursorID, err := parseCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "cursor invalido", "INVALID_CURSOR")
		return
	}
	filter := repository.TransactionFilter{
		AccountID: accountID,
		From:      from,
		To:        to,
		CursorAt:  cursorAt,
		CursorID:  cursorID,
		Limit:     limit,
	}
	items, err := h.txRepo.ListByAccount(r.Context(), filter)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar transacoes", "TRANSACTION_LIST_FAILED")
		return
	}
	response := transactionListResponse{
		Items:      make([]transactionResponse, 0, len(items)),
		NextCursor: buildNextCursor(items),
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		response.Items = append(response.Items, transactionResponseFromEntity(item))
	}
	writeJSON(r.Context(), w, http.StatusOK, response)
}

func (h *Handler) handleAdminAccountFreeze(w http.ResponseWriter, r *http.Request, accountID string) {
	if r.Method != http.MethodPut {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if err := h.ensureRole(r.Context(), "ADMIN", "COMPLIANCE", "OPS"); err != nil {
		switch {
		case errors.Is(err, errRoleRepoMissing):
			writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLE_REPO_MISSING")
		case errors.Is(err, errRoleForbidden):
			writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "ROLE_FORBIDDEN")
		default:
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar roles", "ROLE_LOOKUP_FAILED")
		}
		return
	}
	if h.accountRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "contas nao configuradas", "ACCOUNT_MISSING")
		return
	}
	var payload adminAccountFreezeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	account, err := h.accountRepo.GetByID(r.Context(), accountID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "conta nao encontrada", "ACCOUNT_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar conta", "ACCOUNT_LOOKUP_FAILED")
		return
	}
	status := entity.AccountStatusActive
	if payload.Frozen {
		status = entity.AccountStatusBlocked
	}
	if err := h.accountRepo.UpdateStatus(r.Context(), accountID, status); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar conta", "ACCOUNT_UPDATE_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "ACCOUNT_STATUS_UPDATE", "account", accountID, map[string]any{
		"user_id": account.UserID,
		"status":  string(status),
		"reason":  strings.TrimSpace(payload.Reason),
		"origin":  defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{
		"account_id": accountID,
		"status":     string(status),
	})
}

func formatTimeOrEmpty(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func (h *Handler) handleAdminAlertsCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	cutoff := reconcileCutoff(r)
	alerts := h.evaluateAlerts(r.Context(), cutoff)
	response := map[string]any{
		"cutoff": cutoff.Format(time.RFC3339Nano),
		"alerts": alerts,
		"alert_count": len(alerts),
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "ALERTS_CHECK", "alerts", "", response)
	writeJSON(r.Context(), w, http.StatusOK, response)
}

func (h *Handler) evaluateAlerts(ctx context.Context, cutoff time.Time) []map[string]any {
	var alerts []map[string]any

	if h.pixRepo != nil && h.alertPixPendingThreshold > 0 {
		if count, err := h.pixRepo.CountPendingBefore(ctx, cutoff); err == nil && count >= int64(h.alertPixPendingThreshold) {
			alerts = append(alerts, map[string]any{
				"type": "PIX_PENDING",
				"count": count,
				"threshold": h.alertPixPendingThreshold,
			})
		}
	}
	if h.paymentRepo != nil && h.alertPaymentPendingThreshold > 0 {
		if count, err := h.paymentRepo.CountPendingBefore(ctx, cutoff); err == nil && count >= int64(h.alertPaymentPendingThreshold) {
			alerts = append(alerts, map[string]any{
				"type": "PAYMENT_PENDING",
				"count": count,
				"threshold": h.alertPaymentPendingThreshold,
			})
		}
	}
	if h.cardRepo != nil && h.alertCardPendingThreshold > 0 {
		if count, err := h.cardRepo.CountPendingBefore(ctx, cutoff); err == nil && count >= int64(h.alertCardPendingThreshold) {
			alerts = append(alerts, map[string]any{
				"type": "CARD_PENDING",
				"count": count,
				"threshold": h.alertCardPendingThreshold,
			})
		}
	}
	if h.txRepo != nil && h.alertTransactionHoldThreshold > 0 {
		if count, err := h.txRepo.CountHoldBefore(ctx, cutoff); err == nil && count >= int64(h.alertTransactionHoldThreshold) {
			alerts = append(alerts, map[string]any{
				"type": "TRANSACTION_HOLD",
				"count": count,
				"threshold": h.alertTransactionHoldThreshold,
			})
		}
	}
	if h.webhookRetryRepo != nil && h.alertWebhookRetryDeadThreshold > 0 {
		if count, err := h.webhookRetryRepo.CountByStatus(ctx, entity.WebhookRetryDead); err == nil && count >= h.alertWebhookRetryDeadThreshold {
			alerts = append(alerts, map[string]any{
				"type": "WEBHOOK_RETRY_DEAD",
				"count": count,
				"threshold": h.alertWebhookRetryDeadThreshold,
			})
		}
	}
	return alerts
}

func reconcileCutoff(r *http.Request) time.Time {
	minutes := parseIntDefault(r.URL.Query().Get("older_than_minutes"), 30)
	if minutes < 1 {
		minutes = 30
	}
	if minutes > 1440 {
		minutes = 1440
	}
	return time.Now().UTC().Add(-time.Duration(minutes) * time.Minute)
}

func parseIntDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func (h *Handler) listPendingPix(r *http.Request, cutoff time.Time, limit int) []map[string]any {
	if h.pixRepo == nil {
		return nil
	}
	items, err := h.pixRepo.ListPendingBefore(r.Context(), cutoff, limit)
	if err != nil {
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"id":         item.ID,
			"user_id":    item.UserID,
			"account_id": item.AccountID,
			"status":     item.Status,
			"direction":  item.Direction,
			"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return result
}

func (h *Handler) listPendingPayments(r *http.Request, cutoff time.Time, limit int) []map[string]any {
	if h.paymentRepo == nil {
		return nil
	}
	items, err := h.paymentRepo.ListPendingBefore(r.Context(), cutoff, limit)
	if err != nil {
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"id":         item.ID,
			"user_id":    item.UserID,
			"account_id": item.AccountID,
			"status":     item.Status,
			"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return result
}

func (h *Handler) listPendingCards(r *http.Request, cutoff time.Time, limit int) []map[string]any {
	if h.cardRepo == nil {
		return nil
	}
	items, err := h.cardRepo.ListPendingBefore(r.Context(), cutoff, limit)
	if err != nil {
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"id":         item.ID,
			"user_id":    item.UserID,
			"account_id": item.AccountID,
			"status":     item.Status,
			"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return result
}

func (h *Handler) listPendingTransactions(r *http.Request, cutoff time.Time, limit int) []map[string]any {
	if h.txRepo == nil {
		return nil
	}
	items, err := h.txRepo.ListHoldBefore(r.Context(), cutoff, limit)
	if err != nil {
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"id":         item.ID,
			"user_id":    item.UserID,
			"account_id": item.AccountID,
			"type":       item.Type,
			"status":     item.Status,
			"occurred_at": item.OccurredAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return result
}

func pricingRuleFromPayload(payload adminPricingRuleRequest, ruleID string) (*entity.PricingRule, error) {
	planID := strings.TrimSpace(payload.PlanID)
	if planID == "" {
		return nil, errors.New("plan_id obrigatorio")
	}
	userType := entity.UserType(strings.ToUpper(strings.TrimSpace(payload.UserType)))
	if userType == "" {
		return nil, errors.New("user_type obrigatorio")
	}
	operation := entity.PricingOperationType(strings.ToUpper(strings.TrimSpace(payload.OperationType)))
	if operation == "" {
		return nil, errors.New("operation_type obrigatorio")
	}
	asset := strings.ToUpper(strings.TrimSpace(payload.Asset))
	if asset == "" {
		asset = entity.PricingAssetAny
	}
	feeType := entity.PricingFeeType(strings.ToUpper(strings.TrimSpace(payload.FeeType)))
	if feeType == "" {
		return nil, errors.New("fee_type obrigatorio")
	}
	versionID := strings.TrimSpace(payload.PricingVersionID)
	return &entity.PricingRule{
		ID:            ruleID,
		PlanID:        planID,
		PricingVersionID: versionID,
		UserType:      userType,
		OperationType: operation,
		Asset:         asset,
		FeeType:       feeType,
		FeeValue:      payload.FeeValue,
		MinFee:        payload.MinFee,
		MaxFee:        payload.MaxFee,
	}, nil
}

func (h *Handler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.tokenService == nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "auth nao configurada", "AUTH_MISSING")
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(r.Context(), w, http.StatusUnauthorized, "token ausente", "TOKEN_MISSING")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		userID, err := h.tokenService.ParseAccessToken(token)
		if err != nil {
			writeError(r.Context(), w, http.StatusUnauthorized, "token invalido", "TOKEN_INVALID")
			return
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) requireRoles(allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !h.authorizeRoles(w, r, allowed...) {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (h *Handler) authorizeRoles(w http.ResponseWriter, r *http.Request, allowed ...string) bool {
	if h.userRoleRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "roles nao configuradas", "ROLE_REPO_MISSING")
		return false
	}
	userID := userIDFromContext(r.Context())
	if userID == "" {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return false
	}
	roles, err := h.userRoleRepo.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar roles", "ROLE_LOOKUP_FAILED")
		return false
	}
	allowedSet := roleSet(allowed...)
	if !roleAllowed(roles, allowedSet) {
		_ = h.appendAudit(r.Context(), userID, "ADMIN_ACCESS_DENIED", "role", "", map[string]any{
			"required_roles": rolesFromSet(allowedSet),
		})
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "ROLE_FORBIDDEN")
		return false
	}
	return true
}

func roleSet(allowed ...string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, role := range allowed {
		code := strings.ToUpper(strings.TrimSpace(role))
		if code == "" {
			continue
		}
		result[code] = struct{}{}
	}
	return result
}

func roleAllowed(userRoles []*entity.UserRole, allowed map[string]struct{}) bool {
	for _, role := range userRoles {
		if role == nil {
			continue
		}
		code := strings.ToUpper(strings.TrimSpace(role.RoleCode))
		if code == "" {
			continue
		}
		if _, ok := allowed[code]; ok {
			return true
		}
	}
	return false
}

func rolesFromSet(allowed map[string]struct{}) []string {
	result := make([]string, 0, len(allowed))
	for code := range allowed {
		result = append(result, code)
	}
	return result
}

var errRoleRepoMissing = errors.New("role_repo_missing")
var errRoleForbidden = errors.New("role_forbidden")

func (h *Handler) ensureRole(ctx context.Context, allowed ...string) error {
	if h.userRoleRepo == nil {
		return errRoleRepoMissing
	}
	userID := userIDFromContext(ctx)
	if userID == "" {
		return errRoleForbidden
	}
	roles, err := h.userRoleRepo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	allowedSet := map[string]struct{}{}
	for _, role := range allowed {
		code := strings.ToUpper(strings.TrimSpace(role))
		if code == "" {
			continue
		}
		allowedSet[code] = struct{}{}
	}
	if !roleAllowed(roles, allowedSet) {
		return errRoleForbidden
	}
	return nil
}

func (h *Handler) handlePreRegistrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.preRegistrationUC == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pre-cadastro nao configurado", "PRE_REGISTRATION_MISSING")
		return
	}
	var payload preRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	item, err := h.preRegistrationUC.Start(r.Context(), usecase.PreRegistrationInput{
		FullName:  payload.FullName,
		Email:     payload.Email,
		Phone:     payload.Phone,
		IP:        clientIPFromRequest(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrPreRegistrationInvalid):
			writeError(r.Context(), w, http.StatusBadRequest, "pre-cadastro invalido", "PRE_REGISTRATION_INVALID")
		case errors.Is(err, usecase.ErrPreRegistrationUserExists):
			writeError(r.Context(), w, http.StatusConflict, "usuario ja cadastrado", "PRE_REGISTRATION_USER_EXISTS")
		case errors.Is(err, usecase.ErrPreRegistrationConflict):
			writeError(r.Context(), w, http.StatusConflict, "pre-cadastro conflitante", "PRE_REGISTRATION_CONFLICT")
		case errors.Is(err, usecase.ErrPreRegistrationBlocked):
			writeError(r.Context(), w, http.StatusTooManyRequests, "verificacao bloqueada", "PRE_REGISTRATION_BLOCKED")
		case errors.Is(err, usecase.ErrPreRegistrationAlreadyVerified):
			writeError(r.Context(), w, http.StatusConflict, "pre-cadastro indisponivel", "PRE_REGISTRATION_UNAVAILABLE")
		default:
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao iniciar pre-cadastro", "PRE_REGISTRATION_START_FAILED")
		}
		return
	}
	_ = h.appendAudit(r.Context(), "", "PRE_REGISTRATION_START", "pre_registration", item.ID, map[string]any{
		"email":  item.Email,
		"phone":  item.Phone,
		"status": item.Status,
		"ip":     clientIPFromRequest(r),
	})
	writeJSON(r.Context(), w, http.StatusCreated, preRegistrationResponseFromEntity(item))
}

func (h *Handler) handlePreRegistrationVerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.preRegistrationUC == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pre-cadastro nao configurado", "PRE_REGISTRATION_MISSING")
		return
	}
	var preID string
	var token string
	if r.Method == http.MethodGet {
		preID = strings.TrimSpace(r.URL.Query().Get("pre_registration_id"))
		if preID == "" {
			preID = strings.TrimSpace(r.URL.Query().Get("id"))
		}
		token = strings.TrimSpace(r.URL.Query().Get("token"))
	} else {
		var payload preRegistrationVerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		preID = payload.PreRegistrationID
		token = payload.Token
	}
	item, err := h.preRegistrationUC.VerifyEmail(r.Context(), usecase.PreRegistrationVerifyInput{
		PreRegistrationID: preID,
		Token:             token,
		IP:                clientIPFromRequest(r),
		Channel:           "EMAIL",
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrPreRegistrationInvalid):
			writeError(r.Context(), w, http.StatusBadRequest, "pre-cadastro invalido", "PRE_REGISTRATION_INVALID")
		case errors.Is(err, usecase.ErrPreRegistrationExpired):
			writeError(r.Context(), w, http.StatusGone, "pre-cadastro expirado", "PRE_REGISTRATION_EXPIRED")
		case errors.Is(err, usecase.ErrPreRegistrationBlocked):
			writeError(r.Context(), w, http.StatusTooManyRequests, "verificacao bloqueada", "PRE_REGISTRATION_BLOCKED")
		case errors.Is(err, usecase.ErrPreRegistrationTokenInvalid):
			writeError(r.Context(), w, http.StatusBadRequest, "token invalido", "PRE_REGISTRATION_TOKEN_INVALID")
		case errors.Is(err, usecase.ErrPreRegistrationAlreadyVerified):
			writeError(r.Context(), w, http.StatusConflict, "email ja verificado", "PRE_REGISTRATION_ALREADY_VERIFIED")
		case errors.Is(err, repository.ErrNotFound):
			writeError(r.Context(), w, http.StatusNotFound, "pre-cadastro nao encontrado", "PRE_REGISTRATION_NOT_FOUND")
		default:
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar email", "PRE_REGISTRATION_EMAIL_VERIFY_FAILED")
		}
		return
	}
	_ = h.appendAudit(r.Context(), "", "PRE_REGISTRATION_EMAIL_VERIFIED", "pre_registration", item.ID, map[string]any{
		"email": item.Email,
		"ip":    clientIPFromRequest(r),
	})
	writeJSON(r.Context(), w, http.StatusOK, preRegistrationResponseFromEntity(item))
}

func (h *Handler) handlePreRegistrationVerifyPhone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.preRegistrationUC == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pre-cadastro nao configurado", "PRE_REGISTRATION_MISSING")
		return
	}
	var payload preRegistrationVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	item, err := h.preRegistrationUC.VerifyPhone(r.Context(), usecase.PreRegistrationVerifyInput{
		PreRegistrationID: payload.PreRegistrationID,
		Token:             payload.Token,
		IP:                clientIPFromRequest(r),
		Channel:           "PHONE",
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrPreRegistrationInvalid):
			writeError(r.Context(), w, http.StatusBadRequest, "pre-cadastro invalido", "PRE_REGISTRATION_INVALID")
		case errors.Is(err, usecase.ErrPreRegistrationExpired):
			writeError(r.Context(), w, http.StatusGone, "pre-cadastro expirado", "PRE_REGISTRATION_EXPIRED")
		case errors.Is(err, usecase.ErrPreRegistrationBlocked):
			writeError(r.Context(), w, http.StatusTooManyRequests, "verificacao bloqueada", "PRE_REGISTRATION_BLOCKED")
		case errors.Is(err, usecase.ErrPreRegistrationTokenInvalid):
			writeError(r.Context(), w, http.StatusBadRequest, "codigo invalido", "PRE_REGISTRATION_TOKEN_INVALID")
		case errors.Is(err, usecase.ErrPreRegistrationAlreadyVerified):
			writeError(r.Context(), w, http.StatusConflict, "telefone ja verificado", "PRE_REGISTRATION_ALREADY_VERIFIED")
		case errors.Is(err, repository.ErrNotFound):
			writeError(r.Context(), w, http.StatusNotFound, "pre-cadastro nao encontrado", "PRE_REGISTRATION_NOT_FOUND")
		default:
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao verificar telefone", "PRE_REGISTRATION_PHONE_VERIFY_FAILED")
		}
		return
	}
	_ = h.appendAudit(r.Context(), "", "PRE_REGISTRATION_PHONE_VERIFIED", "pre_registration", item.ID, map[string]any{
		"phone": item.Phone,
		"ip":    clientIPFromRequest(r),
	})
	writeJSON(r.Context(), w, http.StatusOK, preRegistrationResponseFromEntity(item))
}

func (h *Handler) handlePreRegistrationLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.preRegistrationUC == nil || h.userRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pre-cadastro nao configurado", "PRE_REGISTRATION_MISSING")
		return
	}
	email := strings.TrimSpace(r.URL.Query().Get("email"))
	phone := strings.TrimSpace(r.URL.Query().Get("phone"))
	if email == "" && phone == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "email ou telefone obrigatorio", "PRE_REGISTRATION_FILTER_REQUIRED")
		return
	}
	if email != "" {
		existing, err := h.userRepo.GetByEmail(r.Context(), strings.ToLower(email))
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar usuario", "PRE_REGISTRATION_LOOKUP_FAILED")
			return
		}
		if existing != nil {
			writeJSON(r.Context(), w, http.StatusOK, preRegistrationLookupResponse{
				State: "user_exists",
			})
			return
		}
	}
	item, err := h.preRegistrationUC.Lookup(r.Context(), email, phone)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(r.Context(), w, http.StatusOK, preRegistrationLookupResponse{
				State: "not_found",
			})
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar pre-cadastro", "PRE_REGISTRATION_LOOKUP_FAILED")
		return
	}
	state := "pre_registration"
	if item.Status == entity.PreRegistrationExpired {
		state = "expired"
	}
	_ = h.appendAudit(r.Context(), "", "PRE_REGISTRATION_LOOKUP", "pre_registration", item.ID, map[string]any{
		"email": item.Email,
		"phone": item.Phone,
		"ip":    clientIPFromRequest(r),
	})
	writeJSON(r.Context(), w, http.StatusOK, preRegistrationLookupResponse{
		State:           state,
		PreRegistration: ptrPreRegistrationResponse(preRegistrationResponseFromEntity(item)),
	})
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.loginUC == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "auth nao configurada", "AUTH_MISSING")
		return
	}
	var payload loginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	result, err := h.loginUC.Execute(r.Context(), usecase.LoginInput{
		Email:    payload.Email,
		Password: payload.Password,
		IP:       clientIPFromRequest(r),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrAuthRateLimited):
			_ = h.appendAudit(r.Context(), "", "AUTH_LOGIN_RATE_LIMIT", "auth", payload.Email, map[string]any{
				"ip": clientIPFromRequest(r),
			})
			writeError(r.Context(), w, http.StatusTooManyRequests, "muitas tentativas", "RATE_LIMITED")
		default:
			_ = h.appendAudit(r.Context(), "", "AUTH_LOGIN_FAILED", "auth", payload.Email, map[string]any{
				"ip": clientIPFromRequest(r),
			})
			writeError(r.Context(), w, http.StatusUnauthorized, "credenciais invalidas", "INVALID_CREDENTIALS")
		}
		return
	}
	_ = h.appendAudit(r.Context(), result.UserID, "AUTH_LOGIN_SUCCESS", "auth", result.UserID, map[string]any{
		"ip": clientIPFromRequest(r),
	})
	planCode, uxMode, allowedModes, features, limits := h.resolveUXContext(r.Context(), result.UserID)
	writeJSON(r.Context(), w, http.StatusOK, loginResponse{
		UserID:       result.UserID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.UTC().Format(time.RFC3339Nano),
		PlanCode:     planCode,
		UXMode:       uxMode,
		AllowedModes: allowedModes,
		Features:     features,
		Capabilities: capabilitySnapshot{
			PlanCode:     planCode,
			UXMode:       uxMode,
			AllowedModes: allowedModes,
			Features:     features,
			Limits:       limits,
		},
	})
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.refreshTokenUC == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "auth nao configurada", "AUTH_MISSING")
		return
	}
	var payload refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	token := strings.TrimSpace(payload.RefreshToken)
	result, err := h.refreshTokenUC.Execute(r.Context(), usecase.RefreshInput{
		RefreshToken: token,
		IP:           clientIPFromRequest(r),
	})
	if err != nil {
		_ = h.appendAudit(r.Context(), "", "AUTH_REFRESH_FAILED", "auth", "", map[string]any{
			"ip":         clientIPFromRequest(r),
			"token_len":  len(token),
			"token_hash": hashToken(token),
			"error":      err.Error(),
		})
		writeError(r.Context(), w, http.StatusUnauthorized, "refresh token invalido", "REFRESH_INVALID")
		return
	}
	_ = h.appendAudit(r.Context(), result.UserID, "AUTH_REFRESH_SUCCESS", "auth", result.UserID, map[string]any{
		"ip": clientIPFromRequest(r),
	})
	planCode, uxMode, allowedModes, features, limits := h.resolveUXContext(r.Context(), result.UserID)
	writeJSON(r.Context(), w, http.StatusOK, loginResponse{
		UserID:       result.UserID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.UTC().Format(time.RFC3339Nano),
		PlanCode:     planCode,
		UXMode:       uxMode,
		AllowedModes: allowedModes,
		Features:     features,
		Capabilities: capabilitySnapshot{
			PlanCode:     planCode,
			UXMode:       uxMode,
			AllowedModes: allowedModes,
			Features:     features,
			Limits:       limits,
		},
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.logoutUC == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "auth nao configurada", "AUTH_MISSING")
		return
	}
	var payload refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	token := strings.TrimSpace(payload.RefreshToken)
	if err := h.logoutUC.Execute(r.Context(), usecase.LogoutInput{RefreshToken: token}); err != nil {
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "AUTH_LOGOUT_FAILED", "auth", userIDFromContext(r.Context()), map[string]any{
			"ip":         clientIPFromRequest(r),
			"token_len":  len(token),
			"token_hash": hashToken(token),
			"error":      err.Error(),
		})
		writeError(r.Context(), w, http.StatusBadRequest, "logout invalido", "LOGOUT_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "AUTH_LOGOUT_SUCCESS", "auth", userIDFromContext(r.Context()), map[string]any{
		"ip": clientIPFromRequest(r),
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	userID := userIDFromContext(r.Context())
	if userID == "" {
		writeError(r.Context(), w, http.StatusUnauthorized, "token invalido", "TOKEN_INVALID")
		return
	}
	planCode, uxMode, allowedModes, features, limits := h.resolveUXContext(r.Context(), userID)
	writeJSON(r.Context(), w, http.StatusOK, map[string]any{
		"user_id":       userID,
		"plan_code":     planCode,
		"ux_mode":       uxMode,
		"allowed_modes": allowedModes,
		"features":      features,
		"limits":        limits,
		"capabilities": capabilitySnapshot{
			PlanCode:     planCode,
			UXMode:       uxMode,
			AllowedModes: allowedModes,
			Features:     features,
			Limits:       limits,
		},
	})
}
func (h *Handler) handleTransactions(w http.ResponseWriter, r *http.Request) {
	authUserID := userIDFromContext(r.Context())
	switch r.Method {
	case http.MethodPost:
		requireIdempotency(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload createTransactionRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
				return
			}

			idempotencyKey := idempotencyKeyFromContext(r.Context())
			if idempotencyKey == "" {
				writeError(r.Context(), w, http.StatusBadRequest, "idempotency key obrigatoria", "IDEMPOTENCY_REQUIRED")
				return
			}
			resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
			if err != nil {
				writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
				return
			}
			if err := h.ensureAccountOwnership(r.Context(), authUserID, payload.AccountID); err != nil {
				writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
				return
			}

			occurredAt, err := parseOptionalTime(payload.OccurredAt)
			if err != nil {
				writeError(r.Context(), w, http.StatusBadRequest, "occurred_at invalido", "INVALID_OCCURRED_AT")
				return
			}

			netAmount := payload.NetAmount
			if netAmount == 0 && payload.Amount > 0 {
				netAmount = payload.Amount - payload.Fee
			}

			txType := entity.TransactionType(payload.Type)
			status := entity.TransactionStatus(defaultIfEmpty(payload.Status, string(entity.TransactionStatusCreated)))
			if txType == entity.TransactionTypeDeposit {
				status = entity.TransactionStatusConfirmed
			}
			if txType == entity.TransactionTypePayment || txType == entity.TransactionTypeTradeBuy || txType == entity.TransactionTypeCardAuth {
				if payload.Status == "" {
					status = entity.TransactionStatusHold
				}
			}

			input := usecase.CreateTransactionInput{
				ID:             uuid.NewString(),
				AccountID:      payload.AccountID,
				UserID:         resolvedUserID,
				Type:           txType,
				Status:         status,
				Amount:         payload.Amount,
				Fee:            payload.Fee,
				NetAmount:      netAmount,
				IdempotencyKey: idempotencyKey,
				ExternalRef:    payload.ExternalRef,
				OccurredAt:     occurredAt,
			}

			tx, err := h.createTxUC.Execute(r.Context(), input)
			if err != nil {
				if errors.Is(err, repository.ErrInsufficientFunds) {
					writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
					return
				}
				if errors.Is(err, repository.ErrNotFound) {
					writeError(r.Context(), w, http.StatusNotFound, "conta nao encontrada", "ACCOUNT_NOT_FOUND")
					return
				}
				if errors.Is(err, usecase.ErrAccountInactive) {
					writeError(r.Context(), w, http.StatusConflict, "conta bloqueada ou encerrada", "ACCOUNT_INACTIVE")
					return
				}
				writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "TRANSACTION_CREATE_FAILED")
				return
			}

			writeJSON(r.Context(), w, http.StatusCreated, transactionResponseFromEntity(tx))
		})).ServeHTTP(w, r)
	case http.MethodGet:
		accountID := r.URL.Query().Get("account_id")
		if accountID == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "account_id obrigatorio", "ACCOUNT_ID_REQUIRED")
			return
		}
		if err := h.ensureAccountOwnership(r.Context(), authUserID, accountID); err != nil {
			writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
			return
		}

		limit, err := parseLimit(r.URL.Query().Get("limit"))
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "limit invalido", "INVALID_LIMIT")
			return
		}

		from, err := parseTimeParam(r.URL.Query().Get("from"))
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "from invalido", "INVALID_FROM")
			return
		}
		to, err := parseTimeParam(r.URL.Query().Get("to"))
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "to invalido", "INVALID_TO")
			return
		}

		_, _, _, _, limits := h.resolveUXContext(r.Context(), authUserID)
		if maxItems, ok := limits["HISTORY_ITEMS"]; ok && maxItems > 0 && int64(limit) > maxItems {
			limit = int(maxItems)
		}
		if maxDays, ok := limits["HISTORY_DAYS"]; ok && maxDays > 0 {
			cutoff := time.Now().UTC().Add(-time.Duration(maxDays) * 24 * time.Hour)
			if from == nil || from.Before(cutoff) {
				from = &cutoff
			}
		}

		cursorAt, cursorID, err := parseCursor(r.URL.Query().Get("cursor"))
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "cursor invalido", "INVALID_CURSOR")
			return
		}

		items, err := h.txRepo.ListByAccount(r.Context(), repository.TransactionFilter{
			AccountID: accountID,
			From:      from,
			To:        to,
			CursorAt:  cursorAt,
			CursorID:  cursorID,
			Limit:     limit,
		})
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar transacoes", "TRANSACTION_LIST_FAILED")
			return
		}

		response := make([]transactionResponse, 0, len(items))
		for _, item := range items {
			response = append(response, transactionResponseFromEntity(item))
		}

		nextCursor := buildNextCursor(items)
		writeJSON(r.Context(), w, http.StatusOK, transactionListResponse{
			Items:      response,
			NextCursor: nextCursor,
		})
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleTransactionReverse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	var payload reversalRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.TransactionID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "transaction_id obrigatorio", "TRANSACTION_ID_REQUIRED")
		return
	}

	original, err := h.txRepo.GetByID(r.Context(), payload.TransactionID)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "transacao nao encontrada", "TRANSACTION_NOT_FOUND")
		return
	}
	if original.UserID != userIDFromContext(r.Context()) {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	if original.Status != entity.TransactionStatusConfirmed {
		writeError(r.Context(), w, http.StatusConflict, "transacao nao confirmada", "TRANSACTION_NOT_CONFIRMED")
		return
	}

	var existing string
	if err := h.pool.QueryRow(r.Context(), "SELECT id FROM transactions WHERE reversal_of_transaction_id = $1", payload.TransactionID).Scan(&existing); err == nil {
		writeError(r.Context(), w, http.StatusConflict, "transacao ja revertida", "TRANSACTION_ALREADY_REVERSED")
		return
	}

	reversalInput := usecase.CreateTransactionInput{
		ID:             uuid.NewString(),
		AccountID:      original.AccountID,
		UserID:         original.UserID,
		Type:           entity.TransactionTypeReversal,
		Status:         entity.TransactionStatusConfirmed,
		Amount:         original.Amount,
		Fee:            0,
		NetAmount:      original.NetAmount,
		IdempotencyKey: idempotencyKeyFromContext(r.Context()),
		ExternalRef:    "",
		ReversalOf:     payload.TransactionID,
		OccurredAt:     nil,
	}

	tx, err := h.createTxUC.Execute(r.Context(), reversalInput)
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "TRANSACTION_REVERSAL_FAILED")
		return
	}

	_ = h.appendAudit(r.Context(), original.UserID, "TRANSACTION_REVERSAL", "transaction", tx.ID, map[string]any{
		"reversal_of": payload.TransactionID,
		"reason":      strings.TrimSpace(payload.Reason),
	})

	writeJSON(r.Context(), w, http.StatusCreated, transactionResponseFromEntity(tx))
}

func (h *Handler) handleTransactionConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	var payload confirmTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.TransactionID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "transaction_id obrigatorio", "TRANSACTION_ID_REQUIRED")
		return
	}
	existing, err := h.txRepo.GetByID(r.Context(), payload.TransactionID)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "transacao nao encontrada", "TRANSACTION_NOT_FOUND")
		return
	}
	if existing.UserID != userIDFromContext(r.Context()) {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}

	tx, err := h.txRepo.UpdateStatusIfCurrent(r.Context(), payload.TransactionID, entity.TransactionStatusHold, entity.TransactionStatusConfirmed)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar transacao", "TRANSACTION_UPDATE_FAILED")
		return
	}
	if tx == nil {
		if existing.Status == entity.TransactionStatusConfirmed {
			writeJSON(r.Context(), w, http.StatusOK, transactionResponseFromEntity(existing))
			return
		}
		writeError(r.Context(), w, http.StatusConflict, "transacao nao esta em HOLD", "TRANSACTION_NOT_HOLD")
		return
	}

	writeJSON(r.Context(), w, http.StatusOK, transactionResponseFromEntity(tx))
}

func (h *Handler) handleTransactionReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	var payload confirmTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.TransactionID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "transaction_id obrigatorio", "TRANSACTION_ID_REQUIRED")
		return
	}
	existing, err := h.txRepo.GetByID(r.Context(), payload.TransactionID)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "transacao nao encontrada", "TRANSACTION_NOT_FOUND")
		return
	}
	if existing.UserID != userIDFromContext(r.Context()) {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}

	tx, err := h.txRepo.UpdateStatusIfCurrent(r.Context(), payload.TransactionID, entity.TransactionStatusHold, entity.TransactionStatusRejected)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar transacao", "TRANSACTION_UPDATE_FAILED")
		return
	}
	if tx == nil {
		if existing.Status == entity.TransactionStatusRejected {
			writeJSON(r.Context(), w, http.StatusOK, transactionResponseFromEntity(existing))
			return
		}
		writeError(r.Context(), w, http.StatusConflict, "transacao nao esta em HOLD", "TRANSACTION_NOT_HOLD")
		return
	}

	_ = h.appendAudit(r.Context(), tx.UserID, "TRANSACTION_REJECT", "transaction", tx.ID, map[string]any{
		"status": "REJECTED",
	})

	writeJSON(r.Context(), w, http.StatusOK, transactionResponseFromEntity(tx))
}

func (h *Handler) handlePixKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	var payload pixKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	if err := h.ensureAccountOwnership(r.Context(), resolvedUserID, payload.AccountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}

	key, err := h.registerPixKeyUC.Execute(r.Context(), usecase.RegisterPixKeyInput{
		UserID:    resolvedUserID,
		AccountID: payload.AccountID,
		Type:      payload.Type,
		Key:       payload.Key,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrPixKeyDuplicate) {
			writeError(r.Context(), w, http.StatusConflict, "chave ja cadastrada", "PIX_KEY_DUPLICATE")
			return
		}
		if errors.Is(err, usecase.ErrPixKeyInvalidType) || errors.Is(err, usecase.ErrPixKeyInvalidFormat) {
			writeError(r.Context(), w, http.StatusBadRequest, "chave invalida", "PIX_KEY_INVALID")
			return
		}
		if errors.Is(err, usecase.ErrAccountInactive) {
			writeError(r.Context(), w, http.StatusConflict, "conta bloqueada ou encerrada", "ACCOUNT_INACTIVE")
			return
		}
		writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "PIX_KEY_CREATE_FAILED")
		return
	}

	writeJSON(r.Context(), w, http.StatusCreated, pixKeyResponse{
		ID:        key.ID,
		UserID:    key.UserID,
		AccountID: key.AccountID,
		Type:      string(key.Type),
		Key:       key.Key,
		CreatedAt: key.CreatedAt.UTC().Format(time.RFC3339Nano),
	})
}

func (h *Handler) handlePixSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	ctx, span := otel.Tracer("zetta").Start(r.Context(), "pix.send")
	defer span.End()
	r = r.WithContext(ctx)

	var payload pixSendRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	span.SetAttributes(
		attribute.String("user_id", resolvedUserID),
		attribute.String("account_id", payload.AccountID),
		attribute.Int64("amount", payload.Amount),
	)
	if err := h.ensureAccountOwnership(r.Context(), resolvedUserID, payload.AccountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}

	idempotencyKey := idempotencyKeyFromContext(r.Context())
	if idempotencyKey == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "idempotency key obrigatoria", "IDEMPOTENCY_REQUIRED")
		return
	}

	if err := h.velocityChecker.Check(r.Context(), resolvedUserID, payload.Amount); err != nil {
		if h.handleVelocityError(r.Context(), w, err) {
			return
		}
		writeError(r.Context(), w, http.StatusConflict, err.Error(), "RISK_CHECK_FAILED")
		return
	}

	transfer, err := h.sendPixUC.Execute(r.Context(), usecase.SendPixInput{
		AccountID:      payload.AccountID,
		UserID:         resolvedUserID,
		Amount:         payload.Amount,
		Fee:            payload.Fee,
		NetAmount:      payload.NetAmount,
		IdempotencyKey: idempotencyKey,
		ExternalRef:    payload.ExternalRef,
		Metadata:       payload.Metadata,
	})
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientFunds) {
			writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
			return
		}
		if errors.Is(err, usecase.ErrAccountInactive) {
			writeError(r.Context(), w, http.StatusConflict, "conta bloqueada ou encerrada", "ACCOUNT_INACTIVE")
			return
		}
		writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "PIX_SEND_FAILED")
		return
	}

	writeJSON(r.Context(), w, http.StatusCreated, pixResponseFromEntity(transfer))
}

func (h *Handler) handlePixSendFromCrypto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	ctx, span := otel.Tracer("zetta").Start(r.Context(), "pix.send_from_crypto")
	defer span.End()
	r = r.WithContext(ctx)

	var payload sendPixFromCryptoRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}

	idempotencyKey := idempotencyKeyFromContext(r.Context())
	if idempotencyKey == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "idempotency key obrigatoria", "IDEMPOTENCY_REQUIRED")
		return
	}

	if payload.AccountID == "" || payload.AmountBRL <= 0 || payload.Asset == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	span.SetAttributes(
		attribute.String("user_id", resolvedUserID),
		attribute.String("account_id", payload.AccountID),
		attribute.String("asset", payload.Asset),
		attribute.Int64("amount_brl", payload.AmountBRL),
	)
	if err := h.ensureAccountOwnership(r.Context(), resolvedUserID, payload.AccountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}

	if err := h.velocityChecker.Check(r.Context(), resolvedUserID, payload.AmountBRL); err != nil {
		if h.handleVelocityError(r.Context(), w, err) {
			return
		}
		writeError(r.Context(), w, http.StatusConflict, err.Error(), "RISK_CHECK_FAILED")
		return
	}

	transfer, err := h.sendPixFromCryptoUC.Execute(r.Context(), usecase.SendPixFromCryptoInput{
		UserID:         resolvedUserID,
		AccountID:      payload.AccountID,
		AmountBRL:      payload.AmountBRL,
		Asset:          payload.Asset,
		IdempotencyKey: idempotencyKey,
		ExternalRef:    payload.ExternalRef,
	})
	if err != nil {
		if h.handleExternalDependencyError(r.Context(), w, resolvedUserID, err) {
			return
		}
		switch {
		case errors.Is(err, usecase.ErrCryptoInvalidAmount):
			writeError(r.Context(), w, http.StatusBadRequest, "valor invalido", "INVALID_AMOUNT")
		case errors.Is(err, usecase.ErrCryptoToPixInsufficient), errors.Is(err, repository.ErrInsufficientFunds):
			writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
		default:
			writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "PIX_FROM_CRYPTO_FAILED")
		}
		return
	}

	writeJSON(r.Context(), w, http.StatusCreated, pixResponseFromEntity(transfer))
}

func (h *Handler) handlePricingQuote(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pair := r.URL.Query().Get("pair")
		if pair == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "pair obrigatorio", "PAIR_REQUIRED")
			return
		}
		quote, ok := h.pricingCache.Get(pair)
		if !ok {
			writeError(r.Context(), w, http.StatusNotFound, "quote nao encontrado", "QUOTE_NOT_FOUND")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, quoteResponse{
			Pair:      quote.Pair,
			Price:     quote.Price,
			ExpiresAt: quote.ExpiresAt.UTC().Format(time.RFC3339Nano),
			Source:    quote.Source,
		})
	case http.MethodPut:
		var payload quoteRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		if payload.Pair == "" || payload.Price <= 0 || payload.TTLSeconds <= 0 {
			writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
			return
		}
		quote := h.pricingCache.Set(payload.Pair, payload.Price, time.Duration(payload.TTLSeconds)*time.Second, defaultIfEmpty(payload.Source, "manual"))
		writeJSON(r.Context(), w, http.StatusOK, quoteResponse{
			Pair:      quote.Pair,
			Price:     quote.Price,
			ExpiresAt: quote.ExpiresAt.UTC().Format(time.RFC3339Nano),
			Source:    quote.Source,
		})
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handlePaymentValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	var payload paymentValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	barcode := normalizeBarcode(payload.Barcode)
	valid := isValidBarcode(barcode)
	writeJSON(r.Context(), w, http.StatusOK, paymentValidateResponse{
		Barcode: barcode,
		Valid:   valid,
	})
}

func (h *Handler) handlePaymentSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	ctx, span := otel.Tracer("zetta").Start(r.Context(), "payment.schedule")
	defer span.End()
	r = r.WithContext(ctx)
	var payload paymentScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.AccountID == "" || payload.Amount <= 0 {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	span.SetAttributes(
		attribute.String("user_id", resolvedUserID),
		attribute.String("account_id", payload.AccountID),
		attribute.Int64("amount", payload.Amount),
	)
	if err := h.ensureAccountOwnership(r.Context(), resolvedUserID, payload.AccountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	barcode := normalizeBarcode(payload.Barcode)
	if barcode != "" && !isValidBarcode(barcode) {
		writeError(r.Context(), w, http.StatusBadRequest, "barcode invalido", "INVALID_BARCODE")
		return
	}

	idempotencyKey := idempotencyKeyFromContext(r.Context())
	existing, err := h.paymentRepo.GetByIdempotencyKey(r.Context(), payload.AccountID, idempotencyKey)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar pagamento", "PAYMENT_LOOKUP_FAILED")
		return
	}
	if existing != nil {
		writeJSON(r.Context(), w, http.StatusOK, paymentResponseFromEntity(existing))
		return
	}

	if err := h.velocityChecker.Check(r.Context(), resolvedUserID, payload.Amount); err != nil {
		if h.handleVelocityError(r.Context(), w, err) {
			return
		}
		writeError(r.Context(), w, http.StatusConflict, err.Error(), "RISK_CHECK_FAILED")
		return
	}

	netAmount := payload.NetAmount
	if netAmount == 0 && payload.Amount > 0 {
		netAmount = payload.Amount - payload.Fee
	}

	txInput := usecase.CreateTransactionInput{
		ID:             uuid.NewString(),
		AccountID:      payload.AccountID,
		UserID:         resolvedUserID,
		Type:           entity.TransactionTypePayment,
		Status:         entity.TransactionStatusHold,
		Amount:         payload.Amount,
		Fee:            payload.Fee,
		NetAmount:      netAmount,
		IdempotencyKey: idempotencyKey,
		ExternalRef:    payload.ExternalRef,
		OccurredAt:     nil,
	}

	txEntity, err := h.createTxUC.Execute(r.Context(), txInput)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientFunds) {
			if h.ensureFiatCoverageUC == nil {
				writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
				return
			}
			settings, _ := h.settingsRepo.GetByUserID(r.Context(), resolvedUserID)
			if settings != nil && !settings.AllowCryptoToFiat {
				writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
				return
			}
			ok, coverErr := h.ensureFiatCoverageUC.Execute(r.Context(), usecase.EnsureFiatCoverageInput{
				UserID:         resolvedUserID,
				AccountID:      payload.AccountID,
				RequiredAmount: netAmount,
				IdempotencyKey: "payment-sell-" + idempotencyKey,
				ExternalRef:    "PAYMENT_SELL",
				Trigger:        entity.ConversionTriggerPayment,
			})
			if coverErr != nil {
				writeError(r.Context(), w, http.StatusBadRequest, coverErr.Error(), "AUTO_CONVERT_FAILED")
				return
			}
			if !ok {
				writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
				return
			}
			txEntity, err = h.createTxUC.Execute(r.Context(), txInput)
			if err != nil {
				if errors.Is(err, repository.ErrInsufficientFunds) {
					writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
					return
				}
				writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "TRANSACTION_CREATE_FAILED")
				return
			}
		} else {
			writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "TRANSACTION_CREATE_FAILED")
			return
		}
	}

	now := time.Now().UTC()
	payment := &entity.Payment{
		ID:             uuid.NewString(),
		UserID:         resolvedUserID,
		AccountID:      payload.AccountID,
		Status:         entity.PaymentStatusPendingPartner,
		Amount:         payload.Amount,
		Fee:            payload.Fee,
		NetAmount:      netAmount,
		IdempotencyKey: idempotencyKey,
		Barcode:        barcode,
		ScheduledAt:    payload.ScheduledAt,
		DueDate:        payload.DueDate,
		ExternalRef:    payload.ExternalRef,
		TransactionID:  txEntity.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := h.paymentRepo.Create(r.Context(), payment); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao criar pagamento", "PAYMENT_CREATE_FAILED")
		return
	}

	_ = h.appendAudit(r.Context(), resolvedUserID, "PAYMENT_SCHEDULE", "payment", payment.ID, map[string]any{
		"status": "PENDING_PARTNER",
	})

	writeJSON(r.Context(), w, http.StatusCreated, paymentResponseFromEntity(payment))
}

func (h *Handler) handlePaymentConfirmWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	var payload paymentWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.PaymentID == "" || payload.TransactionID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	ref := webhookReference(payload.PaymentID, payload.TransactionID)
	alreadyProcessed, err := h.webhookRepo.EnsureEvent(r.Context(), "payment_confirm", ref)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao registrar webhook", "WEBHOOK_EVENT_FAILED")
		return
	}
	if alreadyProcessed {
		h.respondPaymentStatus(r.Context(), w, payload.PaymentID, "CONFIRMED")
		return
	}
	if err := h.webhookRepo.ConfirmPayment(r.Context(), payload.PaymentID, payload.TransactionID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "pagamento nao encontrado", "PAYMENT_NOT_FOUND")
			return
		}
		if errors.Is(err, repository.ErrInvalidState) {
			writeError(r.Context(), w, http.StatusConflict, "pagamento em estado invalido", "PAYMENT_INVALID_STATE")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar pagamento", "PAYMENT_UPDATE_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), payload.UserID, "PAYMENT_CONFIRM", "payment", payload.PaymentID, map[string]any{
		"status": "CONFIRMED",
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "CONFIRMED"})
}

func (h *Handler) handlePaymentRejectWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	var payload paymentWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.PaymentID == "" || payload.TransactionID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	ref := webhookReference(payload.PaymentID, payload.TransactionID)
	alreadyProcessed, err := h.webhookRepo.EnsureEvent(r.Context(), "payment_reject", ref)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao registrar webhook", "WEBHOOK_EVENT_FAILED")
		return
	}
	if alreadyProcessed {
		h.respondPaymentStatus(r.Context(), w, payload.PaymentID, "REJECTED")
		return
	}
	if err := h.webhookRepo.RejectPayment(r.Context(), payload.PaymentID, payload.TransactionID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "pagamento nao encontrado", "PAYMENT_NOT_FOUND")
			return
		}
		if errors.Is(err, repository.ErrInvalidState) {
			writeError(r.Context(), w, http.StatusConflict, "pagamento em estado invalido", "PAYMENT_INVALID_STATE")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar pagamento", "PAYMENT_UPDATE_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), payload.UserID, "PAYMENT_REJECT", "payment", payload.PaymentID, map[string]any{
		"status": "REJECTED",
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "REJECTED"})
}

func (h *Handler) handleCardAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	ctx, span := otel.Tracer("zetta").Start(r.Context(), "card.authorize")
	defer span.End()
	r = r.WithContext(ctx)
	var payload cardAuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.AccountID == "" || payload.Amount <= 0 {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	span.SetAttributes(
		attribute.String("user_id", resolvedUserID),
		attribute.String("account_id", payload.AccountID),
		attribute.Int64("amount", payload.Amount),
		attribute.String("merchant_name", payload.MerchantName),
		attribute.String("merchant_mcc", payload.MerchantMCC),
	)
	if err := h.ensureAccountOwnership(r.Context(), resolvedUserID, payload.AccountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}

	settings, _ := h.settingsRepo.GetByUserID(r.Context(), resolvedUserID)
	if settings != nil && !settings.AllowCryptoToFiat {
		writeJSON(r.Context(), w, http.StatusOK, cardAuthorizeResponse{Approved: false, Reason: "crypto_to_fiat_desabilitado"})
		return
	}

	result, err := h.authorizeCardJITUC.Execute(r.Context(), usecase.AuthorizeCardInput{
		UserID:         resolvedUserID,
		AccountID:      payload.AccountID,
		Amount:         payload.Amount,
		Fee:            payload.Fee,
		NetAmount:      payload.NetAmount,
		MerchantName:   payload.MerchantName,
		MerchantMCC:    payload.MerchantMCC,
		AuthCode:       payload.AuthCode,
		ExternalRef:    payload.ExternalRef,
		IdempotencyKey: idempotencyKeyFromContext(r.Context()),
	})
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao autorizar cartao", "CARD_AUTH_FAILED")
		return
	}
	writeJSON(r.Context(), w, http.StatusOK, cardAuthorizeResponse{
		Approved:        result.Approved,
		AuthorizationID: result.AuthorizationID,
		Reason:          result.Reason,
	})
}

func (h *Handler) handleCardConfirmWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	var payload cardWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.AuthorizationID == "" || payload.TransactionID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	ref := webhookReference(payload.AuthorizationID, payload.TransactionID)
	alreadyProcessed, err := h.webhookRepo.EnsureEvent(r.Context(), "card_confirm", ref)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao registrar webhook", "WEBHOOK_EVENT_FAILED")
		return
	}
	if alreadyProcessed {
		h.respondCardStatus(r.Context(), w, payload.AuthorizationID, "CONFIRMED")
		return
	}
	if err := h.webhookRepo.ConfirmCardAuthorization(r.Context(), payload.AuthorizationID, payload.TransactionID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "autorizacao nao encontrada", "CARD_AUTH_NOT_FOUND")
			return
		}
		if errors.Is(err, repository.ErrInvalidState) {
			writeError(r.Context(), w, http.StatusConflict, "autorizacao em estado invalido", "CARD_AUTH_INVALID_STATE")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar autorizacao", "CARD_AUTH_UPDATE_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), payload.UserID, "CARD_AUTH_CONFIRM", "card_authorization", payload.AuthorizationID, map[string]any{
		"status": "CONFIRMED",
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "CONFIRMED"})
}

func (h *Handler) handleCardRejectWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	var payload cardWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.AuthorizationID == "" || payload.TransactionID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	ref := webhookReference(payload.AuthorizationID, payload.TransactionID)
	alreadyProcessed, err := h.webhookRepo.EnsureEvent(r.Context(), "card_reject", ref)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao registrar webhook", "WEBHOOK_EVENT_FAILED")
		return
	}
	if alreadyProcessed {
		h.respondCardStatus(r.Context(), w, payload.AuthorizationID, "REJECTED")
		return
	}
	if err := h.webhookRepo.RejectCardAuthorization(r.Context(), payload.AuthorizationID, payload.TransactionID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "autorizacao nao encontrada", "CARD_AUTH_NOT_FOUND")
			return
		}
		if errors.Is(err, repository.ErrInvalidState) {
			writeError(r.Context(), w, http.StatusConflict, "autorizacao em estado invalido", "CARD_AUTH_INVALID_STATE")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar autorizacao", "CARD_AUTH_UPDATE_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), payload.UserID, "CARD_AUTH_REJECT", "card_authorization", payload.AuthorizationID, map[string]any{
		"status": "REJECTED",
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "REJECTED"})
}

func (h *Handler) handleCryptoSwap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	ctx, span := otel.Tracer("zetta").Start(r.Context(), "crypto.swap")
	defer span.End()
	r = r.WithContext(ctx)
	var payload swapRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.DebitAccountID == "" || payload.CreditAccountID == "" || payload.Price <= 0 || payload.Quantity <= 0 {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	span.SetAttributes(
		attribute.String("user_id", resolvedUserID),
		attribute.String("debit_account_id", payload.DebitAccountID),
		attribute.String("credit_account_id", payload.CreditAccountID),
		attribute.String("base_currency", payload.BaseCurrency),
		attribute.String("quote_currency", payload.QuoteCurrency),
		attribute.Int64("price", payload.Price),
		attribute.Int64("quantity", payload.Quantity),
		attribute.Bool("auto_execute", payload.AutoExecute),
	)
	if err := h.ensureAccountOwnership(r.Context(), resolvedUserID, payload.DebitAccountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	if err := h.ensureAccountOwnership(r.Context(), resolvedUserID, payload.CreditAccountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	total, ok := safeMulInt64(payload.Price, payload.Quantity)
	if !ok {
		writeError(r.Context(), w, http.StatusBadRequest, "overflow", "OVERFLOW")
		return
	}

	idempotencyKey := idempotencyKeyFromContext(r.Context())
	existingOrder, err := h.getTradeByIdempotency(r.Context(), resolvedUserID, idempotencyKey)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar swap", "SWAP_LOOKUP_FAILED")
		return
	}
	if existingOrder != nil {
		writeJSON(r.Context(), w, http.StatusOK, swapResponseFromEntity(existingOrder))
		return
	}

	if err := h.velocityChecker.Check(r.Context(), resolvedUserID, total); err != nil {
		if h.handleVelocityError(r.Context(), w, err) {
			return
		}
		writeError(r.Context(), w, http.StatusConflict, err.Error(), "RISK_CHECK_FAILED")
		return
	}

	now := time.Now().UTC()
	order := &entity.TradeOrder{
		ID:              uuid.NewString(),
		UserID:          resolvedUserID,
		IdempotencyKey:  idempotencyKey,
		Status:          entity.TradeStatusCreated,
		Side:            payload.Side,
		BaseCurrency:    payload.BaseCurrency,
		QuoteCurrency:   payload.QuoteCurrency,
		Price:           payload.Price,
		Quantity:        payload.Quantity,
		Fee:             payload.Fee,
		ExternalRef:     payload.ExternalRef,
		DebitAccountID:  payload.DebitAccountID,
		CreditAccountID: payload.CreditAccountID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := h.tradeRepo.Create(r.Context(), order); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao criar swap", "SWAP_CREATE_FAILED")
		return
	}

	if payload.AutoExecute {
		if err := h.executeTrade(r.Context(), order, total); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "SWAP_EXECUTE_FAILED")
			return
		}
	}

	writeJSON(r.Context(), w, http.StatusCreated, swapResponseFromEntity(order))
}

func (h *Handler) handleCryptoPay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	ctx, span := otel.Tracer("zetta").Start(r.Context(), "crypto.pay")
	defer span.End()
	r = r.WithContext(ctx)
	var payload cryptoPayRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	idempotencyKey := idempotencyKeyFromContext(r.Context())
	if idempotencyKey == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "idempotency key obrigatoria", "IDEMPOTENCY_REQUIRED")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	span.SetAttributes(
		attribute.String("user_id", resolvedUserID),
		attribute.Int64("fiat_amount", payload.FiatAmount),
		attribute.String("destination", payload.Destination),
	)

	transfer, err := h.payCryptoWithFiatUC.Execute(r.Context(), usecase.PayCryptoWithFiatInput{
		UserID:         resolvedUserID,
		FiatAmount:     payload.FiatAmount,
		Destination:    payload.Destination,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		if h.handleExternalDependencyError(r.Context(), w, resolvedUserID, err) {
			return
		}
		switch {
		case errors.Is(err, repository.ErrInsufficientFunds):
			writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
		case errors.Is(err, address.ErrInvalidAddress):
			writeError(r.Context(), w, http.StatusBadRequest, "endereco invalido", "INVALID_ADDRESS")
		case errors.Is(err, usecase.ErrCryptoAccountNotFound):
			writeError(r.Context(), w, http.StatusNotFound, "conta nao encontrada", "ACCOUNT_NOT_FOUND")
		default:
			writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "CRYPTO_PAY_FAILED")
		}
		return
	}

	writeJSON(r.Context(), w, http.StatusCreated, cryptoPayResponse{
		ID:            transfer.ID,
		UserID:        transfer.UserID,
		AccountID:     transfer.AccountID,
		Asset:         transfer.Asset,
		Network:       transfer.Network,
		Address:       transfer.Address,
		Amount:        transfer.Amount,
		Fee:           transfer.Fee,
		Status:        string(transfer.Status),
		TransactionID: transfer.TransactionID,
		CreatedAt:     transfer.CreatedAt.UTC().Format(time.RFC3339Nano),
	})
}

func (h *Handler) handleInvoices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	ctx, span := otel.Tracer("zetta").Start(r.Context(), "invoice.create")
	defer span.End()
	r = r.WithContext(ctx)

	var payload createInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	idempotencyKey := idempotencyKeyFromContext(r.Context())
	if idempotencyKey == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "idempotency key obrigatoria", "IDEMPOTENCY_REQUIRED")
		return
	}
	if payload.AmountBRL <= 0 {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	span.SetAttributes(
		attribute.String("user_id", resolvedUserID),
		attribute.Int64("amount_brl", payload.AmountBRL),
	)

	result, err := h.createInvoiceUC.Execute(r.Context(), usecase.CreateInvoiceInput{
		UserID:         resolvedUserID,
		AmountBRL:      payload.AmountBRL,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "INVOICE_CREATE_FAILED")
		return
	}

	writeJSON(r.Context(), w, http.StatusCreated, invoiceResponseFromEntity(result.Invoice, result.QRPayload))
}

func (h *Handler) handleInvoicePay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	ctx, span := otel.Tracer("zetta").Start(r.Context(), "invoice.pay")
	defer span.End()
	r = r.WithContext(ctx)

	var payload payInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}

	idempotencyKey := idempotencyKeyFromContext(r.Context())
	if idempotencyKey == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "idempotency key obrigatoria", "IDEMPOTENCY_REQUIRED")
		return
	}
	resolvedPayerID, err := resolveAuthUserID(r.Context(), payload.PayerUserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	span.SetAttributes(
		attribute.String("payer_user_id", resolvedPayerID),
		attribute.String("payer_account_id", payload.PayerAccountID),
		attribute.String("invoice_id", payload.InvoiceID),
		attribute.String("method", payload.Method),
	)
	if err := h.ensureAccountOwnership(r.Context(), resolvedPayerID, payload.PayerAccountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}

	result, err := h.payInvoiceUC.Execute(r.Context(), usecase.PayInvoiceInput{
		InvoiceID:      payload.InvoiceID,
		PayerUserID:    resolvedPayerID,
		PayerAccountID: payload.PayerAccountID,
		Method:         payload.Method,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		if h.handleExternalDependencyError(r.Context(), w, resolvedPayerID, err) {
			return
		}
		switch {
		case errors.Is(err, repository.ErrNotFound):
			writeError(r.Context(), w, http.StatusNotFound, "invoice nao encontrada", "INVOICE_NOT_FOUND")
		case errors.Is(err, repository.ErrInsufficientFunds):
			writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
		case errors.Is(err, usecase.ErrInvoiceInvalidMethod):
			writeError(r.Context(), w, http.StatusBadRequest, "metodo invalido", "INVALID_METHOD")
		default:
			writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "INVOICE_PAY_FAILED")
		}
		return
	}

	response := payInvoiceResponse{
		Invoice: invoiceResponseFromEntity(result.Invoice, ""),
		Method:  result.Method,
	}
	if result.PixTransfer != nil {
		transfer := pixResponseFromEntity(result.PixTransfer)
		response.PixTransfer = &transfer
	}
	if result.CryptoTransfer != nil {
		crypto := cryptoTransferResponseFromEntity(result.CryptoTransfer)
		response.CryptoTransfer = &crypto
	}

	writeJSON(r.Context(), w, http.StatusCreated, response)
}

func (h *Handler) handleUserSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userID := r.URL.Query().Get("user_id")
		resolvedUserID, err := resolveAuthUserID(r.Context(), userID)
		if err != nil {
			writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
			return
		}
		settings, err := h.settingsRepo.GetByUserID(r.Context(), resolvedUserID)
		if err != nil {
			writeError(r.Context(), w, http.StatusNotFound, "settings nao encontrado", "SETTINGS_NOT_FOUND")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, settings)
	case http.MethodPut:
		var payload userSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
		if err != nil {
			writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
			return
		}
		var previous *entity.UserSettings
		if h.settingsRepo != nil {
			if existing, err := h.settingsRepo.GetByUserID(r.Context(), resolvedUserID); err == nil {
				previous = existing
			}
		}
		_, _, allowedModes, _, _ := h.resolveUXContext(r.Context(), resolvedUserID)
		requestedMode := strings.ToUpper(strings.TrimSpace(payload.UXMode))
		if requestedMode == "" && previous != nil {
			requestedMode = strings.ToUpper(strings.TrimSpace(previous.UXMode))
		}
		if requestedMode == "" {
			requestedMode = defaultUXMode(allowedModes)
		}
		if !modeAllowed(requestedMode, allowedModes) {
			writeError(r.Context(), w, http.StatusConflict, "ux_mode nao permitido", "UX_MODE_NOT_ALLOWED")
			return
		}
		settings := &entity.UserSettings{
			UserID:             resolvedUserID,
			ConversionPriority: payload.ConversionPriority,
			AllowCryptoToFiat:  payload.AllowCryptoToFiat,
			AutoConvertEnabled: payload.AutoConvertEnabled,
			AutoConvertAsset:   payload.AutoConvertAsset,
			AutoConvertMinAmount: payload.AutoConvertMinAmount,
			UXMode:             requestedMode,
		}
		if err := h.settingsRepo.Upsert(r.Context(), settings); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao salvar settings", "SETTINGS_UPDATE_FAILED")
			return
		}
		if previous != nil && strings.ToUpper(strings.TrimSpace(previous.UXMode)) != requestedMode {
			_ = h.appendAudit(r.Context(), resolvedUserID, "UX_MODE_CHANGE", "user_settings", resolvedUserID, map[string]any{
				"from":   previous.UXMode,
				"to":     requestedMode,
				"origin": defaultIfEmpty(r.Header.Get("X-Origin"), "API"),
			})
		}
		writeJSON(r.Context(), w, http.StatusOK, settings)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handlePixWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	var payload pixWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}

	transfer, err := h.pixWebhookUC.Execute(r.Context(), usecase.PixWebhookInput{
		TransferID: payload.TransferID,
		Status:     payload.Status,
	})
	if err != nil {
		if h.handleExternalDependencyError(r.Context(), w, "", err) {
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "pix nao encontrado", "PIX_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "PIX_WEBHOOK_FAILED")
		return
	}

	writeJSON(r.Context(), w, http.StatusOK, pixResponseFromEntity(transfer))
}

func (h *Handler) handleTransactionConfirmWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}

	var payload confirmTransactionRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.TransactionID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "transaction_id obrigatorio", "TRANSACTION_ID_REQUIRED")
		return
	}
	ref := webhookReference(payload.TransactionID)
	alreadyProcessed, err := h.webhookRepo.EnsureEvent(r.Context(), "transaction_confirm", ref)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao registrar webhook", "WEBHOOK_EVENT_FAILED")
		return
	}
	if alreadyProcessed {
		h.respondTransactionStatus(r.Context(), w, payload.TransactionID, "CONFIRMED")
		return
	}

	tx, err := h.txRepo.UpdateStatusIfCurrent(r.Context(), payload.TransactionID, entity.TransactionStatusHold, entity.TransactionStatusConfirmed)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar transacao", "TRANSACTION_UPDATE_FAILED")
		return
	}
	if tx == nil {
		writeError(r.Context(), w, http.StatusNotFound, "transacao nao encontrada ou nao esta em HOLD", "TRANSACTION_NOT_HOLD")
		return
	}

	_ = h.appendAudit(r.Context(), tx.UserID, "TRANSACTION_CONFIRM", "transaction", tx.ID, map[string]any{
		"status": "CONFIRMED",
	})

	writeJSON(r.Context(), w, http.StatusOK, transactionResponseFromEntity(tx))
}

func (h *Handler) handleTransactionRejectWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}

	var payload confirmTransactionRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.TransactionID == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "transaction_id obrigatorio", "TRANSACTION_ID_REQUIRED")
		return
	}
	ref := webhookReference(payload.TransactionID)
	alreadyProcessed, err := h.webhookRepo.EnsureEvent(r.Context(), "transaction_reject", ref)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao registrar webhook", "WEBHOOK_EVENT_FAILED")
		return
	}
	if alreadyProcessed {
		h.respondTransactionStatus(r.Context(), w, payload.TransactionID, "REJECTED")
		return
	}

	tx, err := h.txRepo.UpdateStatusIfCurrent(r.Context(), payload.TransactionID, entity.TransactionStatusHold, entity.TransactionStatusRejected)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar transacao", "TRANSACTION_UPDATE_FAILED")
		return
	}
	if tx == nil {
		writeError(r.Context(), w, http.StatusNotFound, "transacao nao encontrada ou nao esta em HOLD", "TRANSACTION_NOT_HOLD")
		return
	}

	_ = h.appendAudit(r.Context(), tx.UserID, "TRANSACTION_REJECT", "transaction", tx.ID, map[string]any{
		"status": "REJECTED",
	})

	writeJSON(r.Context(), w, http.StatusOK, transactionResponseFromEntity(tx))
}

func (h *Handler) handleWithdrawal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	var payload withdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}

	idempotencyKey := idempotencyKeyFromContext(r.Context())
	if idempotencyKey == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "idempotency key obrigatoria", "IDEMPOTENCY_REQUIRED")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	if err := h.ensureAccountOwnership(r.Context(), resolvedUserID, payload.AccountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}

	occurredAt, err := parseOptionalTime(payload.OccurredAt)
	if err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "occurred_at invalido", "INVALID_OCCURRED_AT")
		return
	}

	netAmount := payload.NetAmount
	if netAmount == 0 && payload.Amount > 0 {
		netAmount = payload.Amount - payload.Fee
	}

	if err := h.velocityChecker.Check(r.Context(), resolvedUserID, payload.Amount); err != nil {
		if h.handleVelocityError(r.Context(), w, err) {
			return
		}
		writeError(r.Context(), w, http.StatusConflict, err.Error(), "RISK_CHECK_FAILED")
		return
	}

	status := entity.TransactionStatus(defaultIfEmpty(payload.Status, string(entity.TransactionStatusConfirmed)))
	input := usecase.CreateTransactionInput{
		ID:             uuid.NewString(),
		AccountID:      payload.AccountID,
		UserID:         resolvedUserID,
		Type:           entity.TransactionTypeWithdrawal,
		Status:         status,
		Amount:         payload.Amount,
		Fee:            payload.Fee,
		NetAmount:      netAmount,
		IdempotencyKey: idempotencyKey,
		ExternalRef:    payload.ExternalRef,
		OccurredAt:     occurredAt,
	}

	tx, err := h.createTxUC.Execute(r.Context(), input)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientFunds) {
			writeError(r.Context(), w, http.StatusConflict, "saldo insuficiente", "INSUFFICIENT_FUNDS")
			return
		}
		writeError(r.Context(), w, http.StatusBadRequest, err.Error(), "TRANSACTION_CREATE_FAILED")
		return
	}

	writeJSON(r.Context(), w, http.StatusCreated, transactionResponseFromEntity(tx))
}

func (h *Handler) handleAccountBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/accounts/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "balance" || parts[0] == "" {
		writeError(r.Context(), w, http.StatusNotFound, "endpoint nao encontrado", "NOT_FOUND")
		return
	}

	accountID := parts[0]
	if err := h.ensureAccountOwnership(r.Context(), userIDFromContext(r.Context()), accountID); err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}
	account, err := h.accountRepo.GetByID(r.Context(), accountID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "conta nao encontrada", "ACCOUNT_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar conta", "ACCOUNT_LOOKUP_FAILED")
		return
	}

	balance, err := h.txRepo.GetLedgerBalance(r.Context(), accountID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao calcular saldo", "BALANCE_CALC_FAILED")
		return
	}

	writeJSON(r.Context(), w, http.StatusOK, accountBalanceResponse{
		AccountID: accountID,
		Currency:  account.Currency,
		Scale:     account.Scale,
		Balance:   balance,
	})
}

func (h *Handler) handleKycLimits(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		var payload kycLimitRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		if payload.Level == "" || payload.DailyLimit < 0 || payload.MonthlyLimit < 0 {
			writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
			return
		}

		limit := &entity.KycLimit{
			Level:        entity.KYCLevel(payload.Level),
			DailyLimit:   payload.DailyLimit,
			MonthlyLimit: payload.MonthlyLimit,
		}
		if err := h.kycLimitRepo.Upsert(r.Context(), limit); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar limites", "KYC_LIMIT_UPDATE_FAILED")
			return
		}

		_ = h.appendAudit(r.Context(), "", "KYC_LIMIT_UPSERT", "kyc_limits", payload.Level, payload)
		writeJSON(r.Context(), w, http.StatusOK, payload)
	case http.MethodGet:
		level := r.URL.Query().Get("level")
		if level != "" {
			item, err := h.kycLimitRepo.GetByLevel(r.Context(), entity.KYCLevel(level))
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					writeError(r.Context(), w, http.StatusNotFound, "limite nao encontrado", "KYC_LIMIT_NOT_FOUND")
					return
				}
				writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar limites", "KYC_LIMIT_LOOKUP_FAILED")
				return
			}
			writeJSON(r.Context(), w, http.StatusOK, kycLimitResponse{
				Level:        string(item.Level),
				DailyLimit:   item.DailyLimit,
				MonthlyLimit: item.MonthlyLimit,
			})
			return
		}

		items, err := h.kycLimitRepo.ListAll(r.Context())
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar limites", "KYC_LIMIT_LIST_FAILED")
			return
		}

		response := make([]kycLimitResponse, 0, len(items))
		for _, item := range items {
			response = append(response, kycLimitResponse{
				Level:        string(item.Level),
				DailyLimit:   item.DailyLimit,
				MonthlyLimit: item.MonthlyLimit,
			})
		}

		writeJSON(r.Context(), w, http.StatusOK, response)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleKycUpgrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}

	var payload kycUpgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	if payload.Level == "" || payload.Status == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "dados invalidos", "INVALID_DATA")
		return
	}
	resolvedUserID, err := resolveAuthUserID(r.Context(), payload.UserID)
	if err != nil {
		writeError(r.Context(), w, http.StatusForbidden, "acesso negado", "FORBIDDEN")
		return
	}

	profile := &entity.KYCProfile{
		UserID:      resolvedUserID,
		Level:       entity.KYCLevel(payload.Level),
		Status:      entity.KYCStatus(payload.Status),
		ProviderRef: payload.ProviderRef,
	}
	if err := h.kycRepo.Upsert(r.Context(), profile); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar kyc", "KYC_UPDATE_FAILED")
		return
	}

	_ = h.appendAudit(r.Context(), resolvedUserID, "KYC_UPGRADE", "kyc_profile", resolvedUserID, payload)
	writeJSON(r.Context(), w, http.StatusOK, payload)
}

func (h *Handler) handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeError(r.Context(), w, http.StatusNotFound, "endpoint nao encontrado", "NOT_FOUND")
}

func (h *Handler) getTradeByIdempotency(ctx context.Context, userID, key string) (*entity.TradeOrder, error) {
	row := h.pool.QueryRow(ctx, `
		SELECT id, user_id, status, side, base_currency, quote_currency,
		       price, quantity, fee, idempotency_key, external_ref,
		       debit_account_id, credit_account_id, debit_transaction_id, credit_transaction_id,
		       created_at, updated_at
		  FROM trade_orders
		 WHERE user_id = $1 AND idempotency_key = $2`, userID, key)

	var order entity.TradeOrder
	var externalRef *string
	var debitAccountID *string
	var creditAccountID *string
	var debitTx *string
	var creditTx *string
	if err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Side,
		&order.BaseCurrency,
		&order.QuoteCurrency,
		&order.Price,
		&order.Quantity,
		&order.Fee,
		&order.IdempotencyKey,
		&externalRef,
		&debitAccountID,
		&creditAccountID,
		&debitTx,
		&creditTx,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if externalRef != nil {
		order.ExternalRef = *externalRef
	}
	if debitAccountID != nil {
		order.DebitAccountID = *debitAccountID
	}
	if creditAccountID != nil {
		order.CreditAccountID = *creditAccountID
	}
	if debitTx != nil {
		order.DebitTransaction = *debitTx
	}
	if creditTx != nil {
		order.CreditTransaction = *creditTx
	}
	return &order, nil
}

func (h *Handler) executeTrade(ctx context.Context, order *entity.TradeOrder, total int64) error {
	if order.Status == entity.TradeStatusExecuted {
		return nil
	}

	debitAmount := total
	if order.Side == entity.TradeSideSell {
		debitAmount = order.Quantity
	}

	debitTx := usecase.CreateTransactionInput{
		ID:             uuid.NewString(),
		AccountID:      order.DebitAccountID,
		UserID:         order.UserID,
		Type:           debitTradeType(order.Side),
		Status:         entity.TransactionStatusConfirmed,
		Amount:         debitAmount,
		Fee:            0,
		NetAmount:      debitAmount,
		IdempotencyKey: order.ID + ":debit",
		ExternalRef:    order.ExternalRef,
		OccurredAt:     nil,
	}

	debitEntity, err := h.createTxUC.Execute(ctx, debitTx)
	if err != nil {
		return err
	}

	creditAmount := order.Quantity
	creditType := entity.TransactionTypeDeposit
	if order.Side == entity.TradeSideSell {
		creditAmount = total
		creditType = entity.TransactionTypeTradeSell
	}
	if order.Fee > creditAmount {
		return fmt.Errorf("fee invalida")
	}

	feeResult := entity.FeeResult{Amount: creditAmount, Fee: 0, NetAmount: creditAmount}
	var pricingRule *entity.PricingRule
	if h.pricingUC != nil {
		asset := defaultIfEmpty(order.QuoteCurrency, "BRL")
		resolved, rule, err := h.pricingUC.Execute(ctx, usecase.PricingInput{
			UserID:        order.UserID,
			OperationType: entity.PricingOperationSwap,
			Asset:         asset,
			GrossAmount:   creditAmount,
		})
		if err != nil {
			return err
		}
		feeResult = resolved
		pricingRule = rule
	}

	creditTx := usecase.CreateTransactionInput{
		ID:             uuid.NewString(),
		AccountID:      order.CreditAccountID,
		UserID:         order.UserID,
		Type:           creditType,
		Status:         entity.TransactionStatusConfirmed,
		Amount:         creditAmount,
		Fee:            feeResult.Fee,
		NetAmount:      feeResult.NetAmount,
		IdempotencyKey: order.ID + ":credit",
		ExternalRef:    order.ExternalRef,
		OccurredAt:     nil,
	}

	creditEntity, err := h.createTxUC.Execute(ctx, creditTx)
	if err != nil {
		return err
	}

	if _, err := h.pool.Exec(ctx, `
		UPDATE trade_orders
		   SET status = 'EXECUTED',
		       debit_transaction_id = $1,
		       credit_transaction_id = $2,
		       updated_at = NOW()
		 WHERE id = $3`, debitEntity.ID, creditEntity.ID, order.ID); err != nil {
		return err
	}

	order.Status = entity.TradeStatusExecuted
	order.DebitTransaction = debitEntity.ID
	order.CreditTransaction = creditEntity.ID

	if feeResult.Fee > 0 {
		ruleID := ""
		if pricingRule != nil {
			ruleID = pricingRule.ID
		}
		_ = h.appendAudit(ctx, order.UserID, "PRICING_APPLIED", "transaction", creditEntity.ID, map[string]any{
			"operation":  string(entity.PricingOperationSwap),
			"amount":     feeResult.Amount,
			"fee":        feeResult.Fee,
			"net_amount": feeResult.NetAmount,
			"rule_id":    ruleID,
		})
	}

	_ = h.appendAudit(ctx, order.UserID, "TRADE_EXECUTE", "trade_order", order.ID, map[string]any{
		"status": "EXECUTED",
	})

	return nil
}

func (h *Handler) appendAudit(ctx context.Context, userID, action, entityType, entityID string, data any) error {
	var payload []byte
	if data != nil {
		var err error
		payload, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	_, err := h.pool.Exec(ctx, `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, data, created_at)
		VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, NOW())`,
		nullIfEmpty(userID),
		action,
		entityType,
		entityID,
		payload,
	)
	if err == nil {
		outboxPayload, marshalErr := json.Marshal(map[string]any{
			"action":      action,
			"user_id":     userID,
			"entity_type": entityType,
			"entity_id":   entityID,
			"data":        data,
		})
		if marshalErr == nil {
			_, err = h.pool.Exec(ctx, `
				INSERT INTO outbox_events (event_type, payload, status, created_at)
				VALUES ($1, $2, 'PENDING', NOW())`, action, outboxPayload)
		}
	}
	if err == nil && h.eventBus != nil {
		h.eventBus.Publish(action, map[string]any{
			"user_id":    userID,
			"entityType": entityType,
			"entityID":   entityID,
			"data":       data,
		})
	}
	return err
}

func (h *Handler) handleVelocityError(ctx context.Context, w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, usecase.ErrDailyLimitExceeded):
		writeError(ctx, w, http.StatusConflict, err.Error(), "DAILY_LIMIT_EXCEEDED")
		return true
	case errors.Is(err, usecase.ErrMonthlyLimitExceeded):
		writeError(ctx, w, http.StatusConflict, err.Error(), "MONTHLY_LIMIT_EXCEEDED")
		return true
	case errors.Is(err, usecase.ErrVelocityCountExceeded):
		writeError(ctx, w, http.StatusConflict, err.Error(), "VELOCITY_TX_LIMIT")
		return true
	case errors.Is(err, usecase.ErrVelocityAmountExceeded):
		writeError(ctx, w, http.StatusConflict, err.Error(), "VELOCITY_AMOUNT_LIMIT")
		return true
	default:
		return false
	}
}

func (h *Handler) handleExternalDependencyError(ctx context.Context, w http.ResponseWriter, userID string, err error) bool {
	externalErr, ok := usecase.AsExternalDependencyError(err)
	if !ok {
		return false
	}
	details := map[string]any{
		"provider":  externalErr.Provider,
		"operation": externalErr.Operation,
	}
	_ = h.appendAudit(ctx, userID, "EXTERNAL_DEPENDENCY_FAILURE", "dependency", "", details)
	writeErrorWithDetails(ctx, w, http.StatusBadGateway, "falha em dependencia externa", "EXTERNAL_DEPENDENCY_FAILED", details)
	return true
}

func hashToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(trimmed))
	return fmt.Sprintf("%x", sum[:])
}

func webhookReference(parts ...string) string {
	return strings.Join(parts, ":")
}

func (h *Handler) respondPaymentStatus(ctx context.Context, w http.ResponseWriter, paymentID, fallback string) {
	payment, err := h.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(ctx, w, http.StatusNotFound, "pagamento nao encontrado", "PAYMENT_NOT_FOUND")
			return
		}
		writeError(ctx, w, http.StatusInternalServerError, "falha ao consultar pagamento", "PAYMENT_LOOKUP_FAILED")
		return
	}
	status := fallback
	if payment != nil {
		status = string(payment.Status)
	}
	writeJSON(ctx, w, http.StatusOK, map[string]string{"status": status})
}

func (h *Handler) resolveUXContext(ctx context.Context, userID string) (string, string, []string, map[string]bool, map[string]int64) {
	planCode := "FREE"
	planID := ""
	if h.userPlanRepo != nil && h.planRepo != nil {
		if planRef, err := h.userPlanRepo.GetActiveByUser(ctx, userID, time.Now().UTC()); err == nil && planRef != nil {
			if plan, err := h.planRepo.GetByID(ctx, planRef.PlanID); err == nil && plan != nil && plan.Code != "" {
				planCode = plan.Code
				planID = plan.ID
			}
		}
	}
	planCode = strings.ToUpper(strings.TrimSpace(planCode))

	features := map[string]bool{}
	if h.planFeatureRepo != nil && planID != "" {
		if items, err := h.planFeatureRepo.ListByPlan(ctx, planID); err == nil {
			for _, item := range items {
				if item == nil {
					continue
				}
				code := strings.ToUpper(strings.TrimSpace(item.FeatureCode))
				if code != "" {
					features[code] = item.Enabled
				}
			}
		}
	}

	allowedModes := allowedUXModes(features)
	uxMode := defaultUXMode(allowedModes)
	if h.settingsRepo != nil {
		if settings, err := h.settingsRepo.GetByUserID(ctx, userID); err == nil && settings != nil {
			requested := strings.ToUpper(strings.TrimSpace(settings.UXMode))
			if requested != "" && modeAllowed(requested, allowedModes) {
				uxMode = requested
			}
		}
	}

	limits := map[string]int64{}
	if h.planLimitRepo != nil && planID != "" {
		if items, err := h.planLimitRepo.ListByPlan(ctx, planID); err == nil {
			for _, item := range items {
				if item == nil {
					continue
				}
				code := strings.ToUpper(strings.TrimSpace(item.LimitCode))
				if code != "" {
					limits[code] = item.LimitValue
				}
			}
		}
	}
	return planCode, uxMode, allowedModes, features, limits
}

func (h *Handler) resolveCapabilitySnapshot(ctx context.Context, userID string) capabilitySnapshot {
	planCode, uxMode, allowedModes, features, limits := h.resolveUXContext(ctx, userID)

	if h.userFeatureOverrideRepo != nil {
		if overrides, err := h.userFeatureOverrideRepo.ListByUser(ctx, userID); err == nil {
			for _, override := range overrides {
				if override == nil {
					continue
				}
				code := strings.ToUpper(strings.TrimSpace(override.FeatureCode))
				if code == "" {
					continue
				}
				features[code] = override.Enabled
			}
		}
	}
	if h.userLimitOverrideRepo != nil {
		if overrides, err := h.userLimitOverrideRepo.ListByUser(ctx, userID); err == nil {
			for _, override := range overrides {
				if override == nil {
					continue
				}
				code := strings.ToUpper(strings.TrimSpace(override.LimitCode))
				if code == "" {
					continue
				}
				limits[code] = override.LimitValue
			}
		}
	}

	allowedFeatures := []string{}
	disabled := []capabilityDisabledFeature{}
	for code, enabled := range features {
		if enabled {
			allowedFeatures = append(allowedFeatures, code)
			continue
		}
		disabled = append(disabled, capabilityDisabledFeature{
			Code:   code,
			Reason: "feature_disabled",
		})
	}
	suggestions := []string{}
	if strings.EqualFold(planCode, "FREE") && len(disabled) > 0 {
		suggestions = append(suggestions, "Pelo seu uso, o modo PRO pode te dar mais controle.")
	}
	return capabilitySnapshot{
		PlanCode:         planCode,
		UXMode:           uxMode,
		AllowedModes:     allowedModes,
		Features:         features,
		Limits:           limits,
		AllowedFeatures:  allowedFeatures,
		DisabledFeatures: disabled,
		Suggestions:      suggestions,
	}
}

func allowedUXModes(features map[string]bool) []string {
	allowed := []string{"LITE", "PRO"}
	if enabled, ok := features["UX_ADVANCED"]; ok && enabled {
		allowed = append(allowed, "ADVANCED")
	}
	return allowed
}

func defaultUXMode(allowed []string) string {
	if len(allowed) == 0 {
		return "LITE"
	}
	return allowed[0]
}

func modeAllowed(mode string, allowed []string) bool {
	for _, item := range allowed {
		if item == mode {
			return true
		}
	}
	return false
}

func (h *Handler) respondCardStatus(ctx context.Context, w http.ResponseWriter, authorizationID, fallback string) {
	auth, err := h.cardRepo.GetByID(ctx, authorizationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(ctx, w, http.StatusNotFound, "autorizacao nao encontrada", "CARD_AUTH_NOT_FOUND")
			return
		}
		writeError(ctx, w, http.StatusInternalServerError, "falha ao consultar autorizacao", "CARD_AUTH_LOOKUP_FAILED")
		return
	}
	status := fallback
	if auth != nil {
		status = string(auth.Status)
	}
	writeJSON(ctx, w, http.StatusOK, map[string]string{"status": status})
}

func (h *Handler) respondTransactionStatus(ctx context.Context, w http.ResponseWriter, transactionID, fallback string) {
	tx, err := h.txRepo.GetByID(ctx, transactionID)
	if err != nil {
		writeError(ctx, w, http.StatusNotFound, "transacao nao encontrada", "TRANSACTION_NOT_FOUND")
		return
	}
	status := fallback
	if tx != nil {
		status = string(tx.Status)
	}
	writeJSON(ctx, w, http.StatusOK, map[string]string{"status": status})
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (h *Handler) ensureAccountOwnership(ctx context.Context, userID, accountID string) error {
	if userID == "" || accountID == "" {
		return repository.ErrNotFound
	}
	account, err := h.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account.UserID != userID {
		return repository.ErrNotFound
	}
	return nil
}

func (h *Handler) handleAdminPricingVersions(w http.ResponseWriter, r *http.Request) {
	if h.pricingVersionRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pricing versions nao configuradas", "PRICING_VERSION_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		if !h.authorizeRoles(w, r, "ADMIN", "AUDIT") {
			return
		}
		items, err := h.pricingVersionRepo.ListAll(r.Context())
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar versions", "PRICING_VERSION_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		if !h.authorizeRoles(w, r, "ADMIN") {
			return
		}
		var payload adminPricingVersionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		code := strings.ToUpper(strings.TrimSpace(payload.Code))
		if code == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "code obrigatorio", "PRICING_VERSION_CODE_REQUIRED")
			return
		}
		status := strings.ToUpper(strings.TrimSpace(payload.Status))
		if status == "" {
			status = string(entity.PricingVersionActive)
		}
		validFrom := time.Now().UTC()
		if payload.ValidFrom != "" {
			parsed, err := time.Parse(time.RFC3339, payload.ValidFrom)
			if err != nil {
				writeError(r.Context(), w, http.StatusBadRequest, "valid_from invalido", "INVALID_VALID_FROM")
				return
			}
			validFrom = parsed
		}
		validUntil, err := parseOptionalTime(payload.ValidUntil)
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "valid_until invalido", "INVALID_VALID_UNTIL")
			return
		}
		version := &entity.PricingVersion{
			ID:          uuid.NewString(),
			Code:        code,
			Description: strings.TrimSpace(payload.Description),
			Status:      entity.PricingVersionStatus(status),
			ValidFrom:   validFrom,
			ValidUntil:  validUntil,
			CreatedAt:   time.Now().UTC(),
		}
		if err := h.pricingVersionRepo.Create(r.Context(), version); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao criar version", "PRICING_VERSION_CREATE_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PRICING_VERSION_CREATE", "pricing_version", version.ID, version)
		writeJSON(r.Context(), w, http.StatusCreated, version)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminPricingVersionActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if h.pricingVersionRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pricing versions nao configuradas", "PRICING_VERSION_MISSING")
		return
	}
	version, err := h.pricingVersionRepo.GetActive(r.Context(), time.Now().UTC())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "version nao encontrada", "PRICING_VERSION_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao consultar version", "PRICING_VERSION_LOOKUP_FAILED")
		return
	}
	writeJSON(r.Context(), w, http.StatusOK, version)
}

func (h *Handler) handleAdminPricingVersionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if !h.authorizeRoles(w, r, "ADMIN") {
		return
	}
	if h.pricingVersionRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pricing versions nao configuradas", "PRICING_VERSION_MISSING")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/admin/pricing/versions/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "id obrigatorio", "PRICING_VERSION_ID_REQUIRED")
		return
	}
	var payload adminStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	status := strings.ToUpper(strings.TrimSpace(payload.Status))
	if status == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "status obrigatorio", "STATUS_REQUIRED")
		return
	}
	if status != string(entity.PricingVersionActive) && status != string(entity.PricingVersionInactive) {
		writeError(r.Context(), w, http.StatusBadRequest, "status invalido", "STATUS_INVALID")
		return
	}
	if err := h.pricingVersionRepo.UpdateStatus(r.Context(), id, entity.PricingVersionStatus(status)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "version nao encontrada", "PRICING_VERSION_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar version", "PRICING_VERSION_UPDATE_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PRICING_VERSION_STATUS", "pricing_version", id, map[string]any{
		"status": status,
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": status})
}

func (h *Handler) handleAdminPlanFeatures(w http.ResponseWriter, r *http.Request) {
	if h.planFeatureRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "plan features nao configuradas", "PLAN_FEATURES_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		if !h.authorizeRoles(w, r, "ADMIN", "COMPLIANCE", "AUDIT") {
			return
		}
		planID := strings.TrimSpace(r.URL.Query().Get("plan_id"))
		if planID == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "plan_id obrigatorio", "PLAN_ID_REQUIRED")
			return
		}
		items, err := h.planFeatureRepo.ListByPlan(r.Context(), planID)
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar features", "PLAN_FEATURE_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPut:
		if !h.authorizeRoles(w, r, "ADMIN") {
			return
		}
		var payload adminPlanFeatureRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		planID := strings.TrimSpace(payload.PlanID)
		featureCode := strings.ToUpper(strings.TrimSpace(payload.FeatureCode))
		if planID == "" || featureCode == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "plan_id e feature_code obrigatorios", "PLAN_FEATURE_REQUIRED")
			return
		}
		var previousEnabled *bool
		if items, err := h.planFeatureRepo.ListByPlan(r.Context(), planID); err == nil {
			for _, item := range items {
				if item != nil && strings.EqualFold(item.FeatureCode, featureCode) {
					value := item.Enabled
					previousEnabled = &value
					break
				}
			}
		}
		item := &entity.PlanFeature{
			ID:          uuid.NewString(),
			PlanID:      planID,
			FeatureCode: featureCode,
			Enabled:     payload.Enabled,
			Metadata:    payload.Metadata,
		}
		if err := h.planFeatureRepo.Upsert(r.Context(), item); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao salvar feature", "PLAN_FEATURE_UPSERT_FAILED")
			return
		}
		reason := strings.TrimSpace(r.Header.Get("X-Reason"))
		_, _ = h.pool.Exec(r.Context(), `
			INSERT INTO plan_feature_history (
				plan_id, feature_code, enabled, change_type, change_reason, changed_by, changed_at
			) VALUES ($1, $2, $3, 'UPSERT', NULLIF($4, ''), NULLIF($5, '')::uuid, NOW())`,
			planID,
			featureCode,
			payload.Enabled,
			reason,
			userIDFromContext(r.Context()),
		)
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_FEATURE_UPSERT", "plan_feature", planID, map[string]any{
			"feature_code": featureCode,
			"from":         previousEnabled,
			"to":           payload.Enabled,
			"origin":       defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
		})
		writeJSON(r.Context(), w, http.StatusOK, item)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminPlanLimits(w http.ResponseWriter, r *http.Request) {
	if h.planLimitRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "plan limits nao configurados", "PLAN_LIMITS_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		if !h.authorizeRoles(w, r, "ADMIN", "COMPLIANCE", "AUDIT") {
			return
		}
		planID := strings.TrimSpace(r.URL.Query().Get("plan_id"))
		if planID == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "plan_id obrigatorio", "PLAN_ID_REQUIRED")
			return
		}
		items, err := h.planLimitRepo.ListByPlan(r.Context(), planID)
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar limites", "PLAN_LIMIT_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPut:
		if !h.authorizeRoles(w, r, "ADMIN") {
			return
		}
		var payload adminPlanLimitRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		planID := strings.TrimSpace(payload.PlanID)
		limitCode := strings.ToUpper(strings.TrimSpace(payload.LimitCode))
		limitWindow := strings.ToUpper(strings.TrimSpace(payload.LimitWindow))
		if planID == "" || limitCode == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "plan_id e limit_code obrigatorios", "PLAN_LIMIT_REQUIRED")
			return
		}
		if limitWindow == "" {
			limitWindow = "MONTHLY"
		}
		var previousValue *int64
		if items, err := h.planLimitRepo.ListByPlan(r.Context(), planID); err == nil {
			for _, item := range items {
				if item != nil && strings.EqualFold(item.LimitCode, limitCode) && strings.EqualFold(item.LimitWindow, limitWindow) {
					value := item.LimitValue
					previousValue = &value
					break
				}
			}
		}
		item := &entity.PlanLimit{
			ID:          uuid.NewString(),
			PlanID:      planID,
			LimitCode:   limitCode,
			LimitValue:  payload.LimitValue,
			LimitWindow: limitWindow,
		}
		if err := h.planLimitRepo.Upsert(r.Context(), item); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao salvar limite", "PLAN_LIMIT_UPSERT_FAILED")
			return
		}
		reason := strings.TrimSpace(r.Header.Get("X-Reason"))
		_, _ = h.pool.Exec(r.Context(), `
			INSERT INTO plan_limit_history (
				plan_id, limit_code, limit_value, limit_window, change_type, change_reason, changed_by, changed_at
			) VALUES ($1, $2, $3, $4, 'UPSERT', NULLIF($5, ''), NULLIF($6, '')::uuid, NOW())`,
			planID,
			limitCode,
			payload.LimitValue,
			limitWindow,
			reason,
			userIDFromContext(r.Context()),
		)
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PLAN_LIMIT_UPSERT", "plan_limit", planID, map[string]any{
			"limit_code":   limitCode,
			"limit_window": limitWindow,
			"from":         previousValue,
			"to":           payload.LimitValue,
			"origin":       defaultIfEmpty(r.Header.Get("X-Origin"), "ADMIN"),
		})
		writeJSON(r.Context(), w, http.StatusOK, item)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminPricingCampaigns(w http.ResponseWriter, r *http.Request) {
	if h.pricingCampaignRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pricing campaigns nao configuradas", "PRICING_CAMPAIGN_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		if !h.authorizeRoles(w, r, "ADMIN", "AUDIT") {
			return
		}
		items, err := h.pricingCampaignRepo.ListAll(r.Context())
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar campaigns", "PRICING_CAMPAIGN_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		if !h.authorizeRoles(w, r, "ADMIN") {
			return
		}
		var payload adminPricingCampaignRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		code := strings.ToUpper(strings.TrimSpace(payload.Code))
		if code == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "code obrigatorio", "PRICING_CAMPAIGN_CODE_REQUIRED")
			return
		}
		status := strings.ToUpper(strings.TrimSpace(payload.Status))
		if status == "" {
			status = string(entity.PricingCampaignActive)
		}
		validFrom := time.Now().UTC()
		if payload.ValidFrom != "" {
			parsed, err := time.Parse(time.RFC3339, payload.ValidFrom)
			if err != nil {
				writeError(r.Context(), w, http.StatusBadRequest, "valid_from invalido", "INVALID_VALID_FROM")
				return
			}
			validFrom = parsed
		}
		validUntil, err := parseOptionalTime(payload.ValidUntil)
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "valid_until invalido", "INVALID_VALID_UNTIL")
			return
		}
		campaign := &entity.PricingCampaign{
			ID:          uuid.NewString(),
			Code:        code,
			Description: strings.TrimSpace(payload.Description),
			Status:      entity.PricingCampaignStatus(status),
			Priority:    payload.Priority,
			ValidFrom:   validFrom,
			ValidUntil:  validUntil,
			CreatedAt:   time.Now().UTC(),
		}
		if err := h.pricingCampaignRepo.Create(r.Context(), campaign); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao criar campaign", "PRICING_CAMPAIGN_CREATE_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PRICING_CAMPAIGN_CREATE", "pricing_campaign", campaign.ID, campaign)
		writeJSON(r.Context(), w, http.StatusCreated, campaign)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}

func (h *Handler) handleAdminPricingCampaignStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
		return
	}
	if !h.authorizeRoles(w, r, "ADMIN") {
		return
	}
	if h.pricingCampaignRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pricing campaigns nao configuradas", "PRICING_CAMPAIGN_MISSING")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/admin/pricing/campaigns/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "id obrigatorio", "PRICING_CAMPAIGN_ID_REQUIRED")
		return
	}
	var payload adminStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
		return
	}
	status := strings.ToUpper(strings.TrimSpace(payload.Status))
	if status == "" {
		writeError(r.Context(), w, http.StatusBadRequest, "status obrigatorio", "STATUS_REQUIRED")
		return
	}
	if status != string(entity.PricingCampaignActive) && status != string(entity.PricingCampaignInactive) {
		writeError(r.Context(), w, http.StatusBadRequest, "status invalido", "STATUS_INVALID")
		return
	}
	if err := h.pricingCampaignRepo.UpdateStatus(r.Context(), id, entity.PricingCampaignStatus(status)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(r.Context(), w, http.StatusNotFound, "campaign nao encontrada", "PRICING_CAMPAIGN_NOT_FOUND")
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "falha ao atualizar campaign", "PRICING_CAMPAIGN_UPDATE_FAILED")
		return
	}
	_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PRICING_CAMPAIGN_STATUS", "pricing_campaign", id, map[string]any{
		"status": status,
	})
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": status})
}

func (h *Handler) handleAdminPricingCampaignRules(w http.ResponseWriter, r *http.Request) {
	if h.pricingCampaignRuleRepo == nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "pricing campaign rules nao configuradas", "PRICING_CAMPAIGN_RULES_MISSING")
		return
	}
	switch r.Method {
	case http.MethodGet:
		if !h.authorizeRoles(w, r, "ADMIN", "AUDIT") {
			return
		}
		campaignID := strings.TrimSpace(r.URL.Query().Get("campaign_id"))
		planID := strings.TrimSpace(r.URL.Query().Get("plan_id"))
		userType := entity.UserType(strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("user_type"))))
		operation := entity.PricingOperationType(strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("operation_type"))))
		if campaignID == "" || userType == "" || operation == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "campaign_id, user_type e operation_type obrigatorios", "CAMPAIGN_RULES_REQUIRED")
			return
		}
		items, err := h.pricingCampaignRuleRepo.ListByCampaign(r.Context(), campaignID, planID, userType, operation)
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao listar rules", "PRICING_CAMPAIGN_RULES_LIST_FAILED")
			return
		}
		writeJSON(r.Context(), w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		if !h.authorizeRoles(w, r, "ADMIN") {
			return
		}
		var payload adminPricingCampaignRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		campaignID := strings.TrimSpace(payload.CampaignID)
		if campaignID == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "campaign_id obrigatorio", "CAMPAIGN_ID_REQUIRED")
			return
		}
		userType := entity.UserType(strings.ToUpper(strings.TrimSpace(payload.UserType)))
		operation := entity.PricingOperationType(strings.ToUpper(strings.TrimSpace(payload.OperationType)))
		asset := strings.ToUpper(strings.TrimSpace(payload.Asset))
		feeType := entity.PricingFeeType(strings.ToUpper(strings.TrimSpace(payload.FeeType)))
		if userType == "" || operation == "" || asset == "" || feeType == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "dados obrigatorios ausentes", "CAMPAIGN_RULES_INVALID")
			return
		}
		var planID *string
		if strings.TrimSpace(payload.PlanID) != "" {
			value := strings.TrimSpace(payload.PlanID)
			planID = &value
		}
		item := &entity.PricingCampaignRule{
			ID:            uuid.NewString(),
			CampaignID:    campaignID,
			PlanID:        planID,
			UserType:      userType,
			OperationType: operation,
			Asset:         asset,
			FeeType:       feeType,
			FeeValue:      payload.FeeValue,
			MinFee:        payload.MinFee,
			MaxFee:        payload.MaxFee,
			CreatedAt:     time.Now().UTC(),
		}
		if err := h.pricingCampaignRuleRepo.Create(r.Context(), item); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "falha ao criar rule", "PRICING_CAMPAIGN_RULE_CREATE_FAILED")
			return
		}
		_ = h.appendAudit(r.Context(), userIDFromContext(r.Context()), "PRICING_CAMPAIGN_RULE_CREATE", "pricing_campaign_rule", item.ID, item)
		writeJSON(r.Context(), w, http.StatusCreated, item)
	default:
		writeError(r.Context(), w, http.StatusMethodNotAllowed, "metodo nao permitido", "METHOD_NOT_ALLOWED")
	}
}
