package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type AuditLogRepository struct {
	db dbExecutor
}

func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{db: pool}
}

func NewAuditLogRepositoryWithTx(tx dbExecutor) *AuditLogRepository {
	return &AuditLogRepository{db: tx}
}

func (r *AuditLogRepository) Append(ctx context.Context, logEntry *entity.AuditLog) error {
	payload := []byte(nil)
	if len(logEntry.Data) > 0 {
		payload = logEntry.Data
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, data, created_at)
		VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6)`,
		nullIfEmpty(logEntry.UserID),
		logEntry.Action,
		logEntry.EntityType,
		logEntry.EntityID,
		payload,
		logEntry.CreatedAt,
	)
	return err
}

func (r *AuditLogRepository) ListByUser(ctx context.Context, userID string, from, to *time.Time, limit int) ([]*entity.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, action, entity_type, entity_id, data, created_at
		  FROM audit_logs
		 WHERE user_id = $1
		   AND ($2::timestamptz IS NULL OR created_at >= $2)
		   AND ($3::timestamptz IS NULL OR created_at <= $3)
		 ORDER BY created_at DESC
		 LIMIT $4`, userID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*entity.AuditLog
	for rows.Next() {
		var item entity.AuditLog
		var data []byte
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Action,
			&item.EntityType,
			&item.EntityID,
			&data,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if len(data) > 0 {
			if json.Valid(data) {
				item.Data = data
			}
		}
		results = append(results, &item)
	}
	return results, rows.Err()
}

var _ repository.AuditLogRepository = (*AuditLogRepository)(nil)
