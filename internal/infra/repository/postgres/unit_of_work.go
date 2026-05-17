package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/repository"
)

type PostgresUnitOfWork struct {
	pool *pgxpool.Pool
}

func NewPostgresUnitOfWork(pool *pgxpool.Pool) *PostgresUnitOfWork {
	return &PostgresUnitOfWork{pool: pool}
}

func (u *PostgresUnitOfWork) Begin(ctx context.Context) (ports.UnitOfWorkTx, error) {
	tx, err := u.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &postgresUnitOfWorkTx{tx: tx}, nil
}

type postgresUnitOfWorkTx struct {
	tx           pgx.Tx
	txRepo       repository.TransactionRepository
	accountRepo  repository.AccountRepository
	pixTransfer  repository.PixTransferRepository
	pixKey       repository.PixKeyRepository
	auditRepo    repository.AuditLogRepository
	userRepo     repository.UserRepository
	cryptoRepo   repository.CryptoTransferRepository
	conversionRepo repository.ConversionRuleRepository
	invoiceRepo    repository.InvoiceRepository
	cardRepo      repository.CardAuthorizationRepository
	refreshTokenRepo repository.RefreshTokenRepository
	loginAuditRepo repository.LoginAuditRepository
	preRegistrationRepo repository.PreRegistrationRepository
	preRegistrationAttemptRepo repository.PreRegistrationAttemptRepository
	outboxRepo repository.OutboxRepository
	complianceCaseRepo repository.ComplianceCaseRepository
	complianceEventRepo repository.ComplianceEventRepository
	conversionAuditRepo repository.ConversionAuditRepository
}

func (u *postgresUnitOfWorkTx) TransactionRepository() repository.TransactionRepository {
	if u.txRepo == nil {
		u.txRepo = NewTransactionRepositoryWithTx(u.tx)
	}
	return u.txRepo
}

func (u *postgresUnitOfWorkTx) AccountRepository() repository.AccountRepository {
	if u.accountRepo == nil {
		u.accountRepo = NewAccountRepositoryWithTx(u.tx)
	}
	return u.accountRepo
}

func (u *postgresUnitOfWorkTx) PixTransferRepository() repository.PixTransferRepository {
	if u.pixTransfer == nil {
		u.pixTransfer = NewPixRepositoryWithTx(u.tx)
	}
	return u.pixTransfer
}

func (u *postgresUnitOfWorkTx) PixKeyRepository() repository.PixKeyRepository {
	if u.pixKey == nil {
		u.pixKey = NewPixKeyRepositoryWithTx(u.tx)
	}
	return u.pixKey
}

func (u *postgresUnitOfWorkTx) AuditLogRepository() repository.AuditLogRepository {
	if u.auditRepo == nil {
		u.auditRepo = NewAuditLogRepositoryWithTx(u.tx)
	}
	return u.auditRepo
}

func (u *postgresUnitOfWorkTx) UserRepository() repository.UserRepository {
	if u.userRepo == nil {
		u.userRepo = NewUserRepositoryWithTx(u.tx)
	}
	return u.userRepo
}

func (u *postgresUnitOfWorkTx) CryptoTransferRepository() repository.CryptoTransferRepository {
	if u.cryptoRepo == nil {
		u.cryptoRepo = NewCryptoTransferRepositoryWithTx(u.tx)
	}
	return u.cryptoRepo
}

func (u *postgresUnitOfWorkTx) ConversionRuleRepository() repository.ConversionRuleRepository {
	if u.conversionRepo == nil {
		u.conversionRepo = NewConversionRuleRepositoryWithTx(u.tx)
	}
	return u.conversionRepo
}

func (u *postgresUnitOfWorkTx) InvoiceRepository() repository.InvoiceRepository {
	if u.invoiceRepo == nil {
		u.invoiceRepo = NewInvoiceRepositoryWithTx(u.tx)
	}
	return u.invoiceRepo
}

func (u *postgresUnitOfWorkTx) CardAuthorizationRepository() repository.CardAuthorizationRepository {
	if u.cardRepo == nil {
		u.cardRepo = NewCardAuthorizationRepositoryWithTx(u.tx)
	}
	return u.cardRepo
}

func (u *postgresUnitOfWorkTx) RefreshTokenRepository() repository.RefreshTokenRepository {
	if u.refreshTokenRepo == nil {
		u.refreshTokenRepo = NewRefreshTokenRepositoryWithTx(u.tx)
	}
	return u.refreshTokenRepo
}

func (u *postgresUnitOfWorkTx) LoginAuditRepository() repository.LoginAuditRepository {
	if u.loginAuditRepo == nil {
		u.loginAuditRepo = NewLoginAuditRepositoryWithTx(u.tx)
	}
	return u.loginAuditRepo
}

func (u *postgresUnitOfWorkTx) PreRegistrationRepository() repository.PreRegistrationRepository {
	if u.preRegistrationRepo == nil {
		u.preRegistrationRepo = NewPreRegistrationRepositoryWithTx(u.tx)
	}
	return u.preRegistrationRepo
}

func (u *postgresUnitOfWorkTx) PreRegistrationAttemptRepository() repository.PreRegistrationAttemptRepository {
	if u.preRegistrationAttemptRepo == nil {
		u.preRegistrationAttemptRepo = NewPreRegistrationAttemptRepositoryWithTx(u.tx)
	}
	return u.preRegistrationAttemptRepo
}

func (u *postgresUnitOfWorkTx) OutboxRepository() repository.OutboxRepository {
	if u.outboxRepo == nil {
		u.outboxRepo = NewOutboxRepositoryWithTx(u.tx)
	}
	return u.outboxRepo
}

func (u *postgresUnitOfWorkTx) ComplianceCaseRepository() repository.ComplianceCaseRepository {
	if u.complianceCaseRepo == nil {
		u.complianceCaseRepo = NewComplianceCaseRepositoryWithTx(u.tx)
	}
	return u.complianceCaseRepo
}

func (u *postgresUnitOfWorkTx) ComplianceEventRepository() repository.ComplianceEventRepository {
	if u.complianceEventRepo == nil {
		u.complianceEventRepo = NewComplianceEventRepositoryWithTx(u.tx)
	}
	return u.complianceEventRepo
}

func (u *postgresUnitOfWorkTx) ConversionAuditRepository() repository.ConversionAuditRepository {
	if u.conversionAuditRepo == nil {
		u.conversionAuditRepo = NewConversionAuditRepositoryWithTx(u.tx)
	}
	return u.conversionAuditRepo
}

func (u *postgresUnitOfWorkTx) Commit(ctx context.Context) error {
	return u.tx.Commit(ctx)
}

func (u *postgresUnitOfWorkTx) Rollback(ctx context.Context) error {
	return u.tx.Rollback(ctx)
}

var _ ports.UnitOfWork = (*PostgresUnitOfWork)(nil)
