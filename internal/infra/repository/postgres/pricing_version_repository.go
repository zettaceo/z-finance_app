package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PricingVersionRepository struct {
	db *pgxpool.Pool
}

func NewPricingVersionRepository(pool *pgxpool.Pool) *PricingVersionRepository {
	return &PricingVersionRepository{db: pool}
}

func (r *PricingVersionRepository) ListAll(ctx context.Context) ([]*entity.PricingVersion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, description, status, valid_from, valid_until, created_at
		  FROM pricing_versions
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.PricingVersion
	for rows.Next() {
		var item entity.PricingVersion
		var validUntil *time.Time
		if err := rows.Scan(&item.ID, &item.Code, &item.Description, &item.Status, &item.ValidFrom, &validUntil, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.ValidUntil = validUntil
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *PricingVersionRepository) GetActive(ctx context.Context, now time.Time) (*entity.PricingVersion, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, code, description, status, valid_from, valid_until, created_at
		  FROM pricing_versions
		 WHERE status = 'ACTIVE'
		   AND valid_from <= $1
		   AND (valid_until IS NULL OR valid_until >= $1)
		 ORDER BY valid_from DESC
		 LIMIT 1`, now)
	var item entity.PricingVersion
	var validUntil *time.Time
	if err := row.Scan(&item.ID, &item.Code, &item.Description, &item.Status, &item.ValidFrom, &validUntil, &item.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	item.ValidUntil = validUntil
	return &item, nil
}

func (r *PricingVersionRepository) Create(ctx context.Context, version *entity.PricingVersion) error {
	if version == nil {
		return repository.ErrInvalidState
	}
	if version.CreatedAt.IsZero() {
		version.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO pricing_versions (id, code, description, status, valid_from, valid_until, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		version.ID,
		version.Code,
		nullIfEmpty(version.Description),
		version.Status,
		version.ValidFrom,
		version.ValidUntil,
		version.CreatedAt,
	)
	return err
}

func (r *PricingVersionRepository) UpdateStatus(ctx context.Context, id string, status entity.PricingVersionStatus) error {
	commandTag, err := r.db.Exec(ctx, `
		UPDATE pricing_versions
		   SET status = $1
		 WHERE id = $2`, status, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

var _ repository.PricingVersionRepository = (*PricingVersionRepository)(nil)
