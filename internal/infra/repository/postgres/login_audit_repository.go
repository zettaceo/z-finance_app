package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type LoginAuditRepository struct {
	db dbExecutor
}

func NewLoginAuditRepository(pool *pgxpool.Pool) *LoginAuditRepository {
	return &LoginAuditRepository{db: pool}
}

func NewLoginAuditRepositoryWithTx(tx dbExecutor) *LoginAuditRepository {
	return &LoginAuditRepository{db: tx}
}

func (r *LoginAuditRepository) Append(ctx context.Context, audit *entity.LoginAudit) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO login_audits (id, user_id, email, ip, success, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		audit.ID,
		nullIfEmpty(audit.UserID),
		nullIfEmpty(audit.Email),
		nullIfEmpty(audit.IP),
		audit.Success,
		nullIfEmpty(audit.Reason),
		audit.CreatedAt,
	)
	return err
}

func (r *LoginAuditRepository) CountRecentFailures(ctx context.Context, email, ip string, since time.Time) (int64, error) {
	row := r.db.QueryRow(ctx, `
		SELECT COALESCE(COUNT(1), 0)
		  FROM login_audits
		 WHERE success = FALSE
		   AND created_at >= $1
		   AND (email = $2 OR ip = $3)`, since, email, ip)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

var _ repository.LoginAuditRepository = (*LoginAuditRepository)(nil)
