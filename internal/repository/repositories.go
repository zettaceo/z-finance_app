package repository

import (
	"context"
	"errors"
	"time"

	"z-finance-api/internal/entity"
)

var ErrInsufficientFunds = errors.New("saldo insuficiente")
var ErrInvalidState = errors.New("estado invalido")
var ErrNotFound = errors.New("nao encontrado")

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateStatus(ctx context.Context, id string, status entity.UserStatus) error
	Search(ctx context.Context, filter UserSearchFilter) ([]*entity.User, error)
}

type UserSearchFilter struct {
	ID         string
	Email      string
	ExternalID string
	Limit      int
}

type PlanRepository interface {
	GetByID(ctx context.Context, id string) (*entity.Plan, error)
	GetByCode(ctx context.Context, code string) (*entity.Plan, error)
	ListAll(ctx context.Context) ([]*entity.Plan, error)
	Create(ctx context.Context, plan *entity.Plan) error
}

type UserPlanRepository interface {
	GetActiveByUser(ctx context.Context, userID string, at time.Time) (*entity.UserPlan, error)
	Create(ctx context.Context, plan *entity.UserPlan) error
}

type PricingRuleRepository interface {
	ListByPlanAndUserType(ctx context.Context, planID string, userType entity.UserType, operation entity.PricingOperationType, pricingVersionID string) ([]*entity.PricingRule, error)
	List(ctx context.Context, filter PricingRuleFilter) ([]*entity.PricingRule, error)
	Create(ctx context.Context, rule *entity.PricingRule) error
	Update(ctx context.Context, rule *entity.PricingRule) error
}

type PricingRuleFilter struct {
	PlanID        string
	UserType      entity.UserType
	OperationType entity.PricingOperationType
	Asset         string
	PricingVersionID string
}

type PricingVersionRepository interface {
	ListAll(ctx context.Context) ([]*entity.PricingVersion, error)
	GetActive(ctx context.Context, now time.Time) (*entity.PricingVersion, error)
	Create(ctx context.Context, version *entity.PricingVersion) error
	UpdateStatus(ctx context.Context, id string, status entity.PricingVersionStatus) error
}

type PlanFeatureRepository interface {
	ListByPlan(ctx context.Context, planID string) ([]*entity.PlanFeature, error)
	Upsert(ctx context.Context, feature *entity.PlanFeature) error
}

type PlanLimitRepository interface {
	ListByPlan(ctx context.Context, planID string) ([]*entity.PlanLimit, error)
	Upsert(ctx context.Context, limit *entity.PlanLimit) error
}

type PricingCampaignRepository interface {
	ListAll(ctx context.Context) ([]*entity.PricingCampaign, error)
	ListActiveByPlan(ctx context.Context, planID string, now time.Time) ([]*entity.PricingCampaign, error)
	Create(ctx context.Context, campaign *entity.PricingCampaign) error
	UpdateStatus(ctx context.Context, id string, status entity.PricingCampaignStatus) error
}

type PricingCampaignRuleRepository interface {
	ListByCampaign(ctx context.Context, campaignID string, planID string, userType entity.UserType, operation entity.PricingOperationType) ([]*entity.PricingCampaignRule, error)
	Create(ctx context.Context, rule *entity.PricingCampaignRule) error
}

type AccountRepository interface {
	Create(ctx context.Context, account *entity.Account) error
	GetByID(ctx context.Context, id string) (*entity.Account, error)
	GetByIDForUpdate(ctx context.Context, id string) (*entity.Account, error)
	ListByUser(ctx context.Context, userID string) ([]*entity.Account, error)
	UpdateStatus(ctx context.Context, id string, status entity.AccountStatus) error
}

type TransactionFilter struct {
	AccountID string
	From      *time.Time
	To        *time.Time
	Type      *entity.TransactionType
	Status    *entity.TransactionStatus
	CursorAt  *time.Time
	CursorID  string
	Limit     int
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	GetByID(ctx context.Context, id string) (*entity.Transaction, error)
	GetByIdempotencyKey(ctx context.Context, accountID, key string) (*entity.Transaction, error)
	ListByAccount(ctx context.Context, filter TransactionFilter) ([]*entity.Transaction, error)
	GetLedgerBalance(ctx context.Context, accountID string) (int64, error)
	GetHoldBalance(ctx context.Context, accountID string) (int64, error)
	UpdateStatusIfCurrent(ctx context.Context, id string, from, to entity.TransactionStatus) (*entity.Transaction, error)
	ListHoldBefore(ctx context.Context, before time.Time, limit int) ([]*entity.Transaction, error)
	CountHoldBefore(ctx context.Context, before time.Time) (int64, error)
	SumOutgoingNetAmountByUserBetween(ctx context.Context, userID string, from, to time.Time) (int64, error)
}

type VelocityRepository interface {
	GetKYCLevel(ctx context.Context, userID string) (entity.KYCLevel, error)
	GetLimitsForLevel(ctx context.Context, level entity.KYCLevel) (int64, int64, bool, error)
	SumConfirmedSpentSince(ctx context.Context, userID string, from time.Time) (int64, error)
	CountRecentTransactionsSince(ctx context.Context, userID string, from time.Time) (int64, error)
	SumRecentNetAmountSince(ctx context.Context, userID string, from time.Time) (int64, error)
}

type PixTransferRepository interface {
	Create(ctx context.Context, transfer *entity.PixTransfer) error
	GetByID(ctx context.Context, id string) (*entity.PixTransfer, error)
	GetByIdempotencyKey(ctx context.Context, accountID, key string) (*entity.PixTransfer, error)
	GetByEndToEndID(ctx context.Context, endToEndID string) (*entity.PixTransfer, error)
	UpdateStatus(ctx context.Context, id string, status entity.PixStatus) error
	UpdateStatusWithConfirmation(ctx context.Context, id string, status entity.PixStatus, confirmedAt *time.Time) error
	UpdateTransactionID(ctx context.Context, id string, transactionID string) error
	ListPendingBefore(ctx context.Context, before time.Time, limit int) ([]*entity.PixTransfer, error)
	CountPendingBefore(ctx context.Context, before time.Time) (int64, error)
}

type PixKeyRepository interface {
	Create(ctx context.Context, key *entity.PixKey) error
	GetByKey(ctx context.Context, value string) (*entity.PixKey, error)
	GetByID(ctx context.Context, id string) (*entity.PixKey, error)
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	GetByID(ctx context.Context, id string) (*entity.Payment, error)
	GetByIdempotencyKey(ctx context.Context, accountID, key string) (*entity.Payment, error)
	UpdateStatus(ctx context.Context, id string, status entity.PaymentStatus) error
	ListPendingBefore(ctx context.Context, before time.Time, limit int) ([]*entity.Payment, error)
	CountPendingBefore(ctx context.Context, before time.Time) (int64, error)
	CountScheduledByUserBetween(ctx context.Context, userID string, from, to time.Time) (int64, error)
}

type CardAuthorizationRepository interface {
	Create(ctx context.Context, auth *entity.CardAuthorization) error
	GetByID(ctx context.Context, id string) (*entity.CardAuthorization, error)
	UpdateStatus(ctx context.Context, id string, status entity.CardAuthStatus) error
	ListPendingBefore(ctx context.Context, before time.Time, limit int) ([]*entity.CardAuthorization, error)
	CountPendingBefore(ctx context.Context, before time.Time) (int64, error)
}

type WebhookRepository interface {
	ConfirmPayment(ctx context.Context, paymentID, transactionID string) error
	RejectPayment(ctx context.Context, paymentID, transactionID string) error
	ConfirmCardAuthorization(ctx context.Context, authorizationID, transactionID string) error
	RejectCardAuthorization(ctx context.Context, authorizationID, transactionID string) error
	EnsureEvent(ctx context.Context, eventType, referenceID string) (bool, error)
}

type WebhookRetryRepository interface {
	Enqueue(ctx context.Context, job *entity.WebhookRetryJob) error
	ListDue(ctx context.Context, now time.Time, limit int) ([]*entity.WebhookRetryJob, error)
	ListByStatus(ctx context.Context, status entity.WebhookRetryStatus, limit int) ([]*entity.WebhookRetryJob, error)
	CountByStatus(ctx context.Context, status entity.WebhookRetryStatus) (int, error)
	MarkSuccess(ctx context.Context, id string) error
	MarkFailure(ctx context.Context, id string, attempts int, nextRetryAt time.Time, lastError string, status entity.WebhookRetryStatus) error
}

type TradeOrderRepository interface {
	Create(ctx context.Context, order *entity.TradeOrder) error
	GetByID(ctx context.Context, id string) (*entity.TradeOrder, error)
	UpdateStatus(ctx context.Context, id string, status entity.TradeStatus) error
}

type UserSettingsRepository interface {
	Upsert(ctx context.Context, settings *entity.UserSettings) error
	GetByUserID(ctx context.Context, userID string) (*entity.UserSettings, error)
}

type KYCRepository interface {
	Create(ctx context.Context, profile *entity.KYCProfile) error
	GetByUserID(ctx context.Context, userID string) (*entity.KYCProfile, error)
	UpdateLevel(ctx context.Context, userID string, level entity.KYCLevel) error
	UpdateStatus(ctx context.Context, userID string, status entity.KYCStatus) error
	Upsert(ctx context.Context, profile *entity.KYCProfile) error
}

type KycLimitRepository interface {
	Upsert(ctx context.Context, limit *entity.KycLimit) error
	GetByLevel(ctx context.Context, level entity.KYCLevel) (*entity.KycLimit, error)
	ListAll(ctx context.Context) ([]*entity.KycLimit, error)
}

type AuditLogRepository interface {
	Append(ctx context.Context, log *entity.AuditLog) error
	ListByUser(ctx context.Context, userID string, from, to *time.Time, limit int) ([]*entity.AuditLog, error)
}

type ConversionAuditRepository interface {
	Append(ctx context.Context, audit *entity.ConversionAudit) error
}

type RoleRepository interface {
	ListAll(ctx context.Context) ([]*entity.Role, error)
	Create(ctx context.Context, role *entity.Role) error
}

type UserRoleRepository interface {
	ListByUser(ctx context.Context, userID string) ([]*entity.UserRole, error)
	Assign(ctx context.Context, userRole *entity.UserRole) error
	Remove(ctx context.Context, userID, roleCode string) error
}

type RoleSeparationRepository interface {
	ListAll(ctx context.Context) ([]*entity.RoleSeparationRule, error)
	Create(ctx context.Context, rule *entity.RoleSeparationRule) error
	Remove(ctx context.Context, roleCodeA, roleCodeB string) error
	HasConflict(ctx context.Context, userID, roleCode string) (bool, error)
}

type RegulatoryProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*entity.RegulatoryProfile, error)
	Upsert(ctx context.Context, profile *entity.RegulatoryProfile) error
}

type CryptoTransferRepository interface {
	Create(ctx context.Context, transfer *entity.CryptoTransfer) error
	GetByTransactionID(ctx context.Context, transactionID string) (*entity.CryptoTransfer, error)
	UpdateStatus(ctx context.Context, id string, status entity.CryptoTransferStatus, txHash string) error
	SumConfirmedByUserAsset(ctx context.Context, userID, asset string) (int64, error)
}

type ConversionRuleRepository interface {
	GetActiveByUserAndTrigger(ctx context.Context, userID string, trigger entity.ConversionTrigger) (*entity.ConversionRule, error)
}

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *entity.Invoice) error
	GetByID(ctx context.Context, id string) (*entity.Invoice, error)
	GetByIdempotencyKey(ctx context.Context, userID, key string) (*entity.Invoice, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, id string, replacedBy string) error
	RevokeAllByUser(ctx context.Context, userID string) (int64, error)
}

type UserFeatureOverrideRepository interface {
	ListByUser(ctx context.Context, userID string) ([]*entity.UserFeatureOverride, error)
	Upsert(ctx context.Context, override *entity.UserFeatureOverride) error
}

type UserLimitOverrideRepository interface {
	ListByUser(ctx context.Context, userID string) ([]*entity.UserLimitOverride, error)
	Upsert(ctx context.Context, override *entity.UserLimitOverride) error
}

type LoginAuditRepository interface {
	Append(ctx context.Context, audit *entity.LoginAudit) error
	CountRecentFailures(ctx context.Context, email, ip string, since time.Time) (int64, error)
}

type PreRegistrationRepository interface {
	Create(ctx context.Context, item *entity.PreRegistration) error
	Update(ctx context.Context, item *entity.PreRegistration) error
	GetByID(ctx context.Context, id string) (*entity.PreRegistration, error)
	GetByEmail(ctx context.Context, email string) (*entity.PreRegistration, error)
	GetByPhone(ctx context.Context, phone string) (*entity.PreRegistration, error)
}

type PreRegistrationAttemptRepository interface {
	Append(ctx context.Context, attempt *entity.PreRegistrationAttempt) error
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, eventType string, payload map[string]any) error
}

type ComplianceCaseRepository interface {
	Create(ctx context.Context, item *entity.ComplianceCase) error
	GetByID(ctx context.Context, id string) (*entity.ComplianceCase, error)
	ListByUser(ctx context.Context, userID string, limit int) ([]*entity.ComplianceCase, error)
	UpdateStatus(ctx context.Context, id string, status entity.ComplianceCaseStatus) error
}

type ComplianceEventRepository interface {
	Append(ctx context.Context, item *entity.ComplianceEvent) error
	ListByCase(ctx context.Context, caseID string, limit int) ([]*entity.ComplianceEvent, error)
}
