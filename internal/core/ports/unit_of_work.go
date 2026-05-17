package ports

import (
	"context"

	"z-finance-api/internal/repository"
)

type UnitOfWork interface {
	Begin(ctx context.Context) (UnitOfWorkTx, error)
}

type UnitOfWorkTx interface {
	TransactionRepository() repository.TransactionRepository
	AccountRepository() repository.AccountRepository
	PixTransferRepository() repository.PixTransferRepository
	PixKeyRepository() repository.PixKeyRepository
	AuditLogRepository() repository.AuditLogRepository
	UserRepository() repository.UserRepository
	CryptoTransferRepository() repository.CryptoTransferRepository
	ConversionRuleRepository() repository.ConversionRuleRepository
	InvoiceRepository() repository.InvoiceRepository
	CardAuthorizationRepository() repository.CardAuthorizationRepository
	RefreshTokenRepository() repository.RefreshTokenRepository
	LoginAuditRepository() repository.LoginAuditRepository
	PreRegistrationRepository() repository.PreRegistrationRepository
	PreRegistrationAttemptRepository() repository.PreRegistrationAttemptRepository
	OutboxRepository() repository.OutboxRepository
	ComplianceCaseRepository() repository.ComplianceCaseRepository
	ComplianceEventRepository() repository.ComplianceEventRepository
	ConversionAuditRepository() repository.ConversionAuditRepository
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
