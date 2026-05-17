package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type ComplianceEventRepository struct {
	db dbExecutor
}

func NewComplianceEventRepository(pool *pgxpool.Pool) *ComplianceEventRepository {
	return &ComplianceEventRepository{db: pool}
}

func NewComplianceEventRepositoryWithTx(tx dbExecutor) *ComplianceEventRepository {
	return &ComplianceEventRepository{db: tx}
}

func (r *ComplianceEventRepository) Append(ctx context.Context, item *entity.ComplianceEvent) error {
	payload := []byte(nil)
	if len(item.Payload) > 0 {
		payload = item.Payload
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO compliance_events (id, case_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		item.ID,
		item.CaseID,
		item.EventType,
		payload,
		item.CreatedAt,
	)
	return err
}

func (r *ComplianceEventRepository) ListByCase(ctx context.Context, caseID string, limit int) ([]*entity.ComplianceEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, case_id, event_type, payload, created_at
		  FROM compliance_events
		 WHERE case_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`, caseID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.ComplianceEvent
	for rows.Next() {
		var item entity.ComplianceEvent
		var payload []byte
		if err := rows.Scan(
			&item.ID,
			&item.CaseID,
			&item.EventType,
			&payload,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if len(payload) > 0 && json.Valid(payload) {
			item.Payload = payload
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, repository.ErrNotFound
	}
	return items, nil
}

var _ repository.ComplianceEventRepository = (*ComplianceEventRepository)(nil)
