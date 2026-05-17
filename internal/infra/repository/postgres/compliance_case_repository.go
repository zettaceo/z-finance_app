package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type ComplianceCaseRepository struct {
	db dbExecutor
}

func NewComplianceCaseRepository(pool *pgxpool.Pool) *ComplianceCaseRepository {
	return &ComplianceCaseRepository{db: pool}
}

func NewComplianceCaseRepositoryWithTx(tx dbExecutor) *ComplianceCaseRepository {
	return &ComplianceCaseRepository{db: tx}
}

func (r *ComplianceCaseRepository) Create(ctx context.Context, item *entity.ComplianceCase) error {
	payload := []byte(nil)
	if len(item.Metadata) > 0 {
		payload = item.Metadata
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO compliance_cases (
			id, user_id, type, status, risk_level, title, summary, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`,
		item.ID,
		item.UserID,
		item.Type,
		item.Status,
		item.RiskLevel,
		item.Title,
		item.Summary,
		payload,
		item.CreatedAt,
		item.UpdatedAt,
	)
	return err
}

func (r *ComplianceCaseRepository) GetByID(ctx context.Context, id string) (*entity.ComplianceCase, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, type, status, risk_level, title, summary, metadata, created_at, updated_at
		  FROM compliance_cases
		 WHERE id = $1`, id)
	var item entity.ComplianceCase
	var metadata []byte
	if err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Type,
		&item.Status,
		&item.RiskLevel,
		&item.Title,
		&item.Summary,
		&metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	if len(metadata) > 0 && json.Valid(metadata) {
		item.Metadata = metadata
	}
	return &item, nil
}

func (r *ComplianceCaseRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*entity.ComplianceCase, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, type, status, risk_level, title, summary, metadata, created_at, updated_at
		  FROM compliance_cases
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.ComplianceCase
	for rows.Next() {
		var item entity.ComplianceCase
		var metadata []byte
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Type,
			&item.Status,
			&item.RiskLevel,
			&item.Title,
			&item.Summary,
			&metadata,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if len(metadata) > 0 && json.Valid(metadata) {
			item.Metadata = metadata
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *ComplianceCaseRepository) UpdateStatus(ctx context.Context, id string, status entity.ComplianceCaseStatus) error {
	_, err := r.db.Exec(ctx, `
		UPDATE compliance_cases
		   SET status = $1, updated_at = NOW()
		 WHERE id = $2`, status, id)
	return err
}

var _ repository.ComplianceCaseRepository = (*ComplianceCaseRepository)(nil)
