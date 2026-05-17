package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PricingCampaignRepository struct {
	db *pgxpool.Pool
}

func NewPricingCampaignRepository(pool *pgxpool.Pool) *PricingCampaignRepository {
	return &PricingCampaignRepository{db: pool}
}

func (r *PricingCampaignRepository) ListAll(ctx context.Context) ([]*entity.PricingCampaign, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, description, status, priority, valid_from, valid_until, created_at
		  FROM pricing_campaigns
		 ORDER BY priority DESC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.PricingCampaign
	for rows.Next() {
		var item entity.PricingCampaign
		var validUntil *time.Time
		if err := rows.Scan(&item.ID, &item.Code, &item.Description, &item.Status, &item.Priority, &item.ValidFrom, &validUntil, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.ValidUntil = validUntil
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *PricingCampaignRepository) ListActiveByPlan(ctx context.Context, planID string, now time.Time) ([]*entity.PricingCampaign, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, description, status, priority, valid_from, valid_until, created_at
		  FROM pricing_campaigns
		 WHERE status = 'ACTIVE'
		   AND valid_from <= $1
		   AND (valid_until IS NULL OR valid_until >= $1)
		 ORDER BY priority DESC, created_at DESC`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.PricingCampaign
	for rows.Next() {
		var item entity.PricingCampaign
		var validUntil *time.Time
		if err := rows.Scan(&item.ID, &item.Code, &item.Description, &item.Status, &item.Priority, &item.ValidFrom, &validUntil, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.ValidUntil = validUntil
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *PricingCampaignRepository) Create(ctx context.Context, campaign *entity.PricingCampaign) error {
	if campaign == nil {
		return repository.ErrInvalidState
	}
	if campaign.CreatedAt.IsZero() {
		campaign.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO pricing_campaigns (id, code, description, status, priority, valid_from, valid_until, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		campaign.ID,
		campaign.Code,
		nullIfEmpty(campaign.Description),
		campaign.Status,
		campaign.Priority,
		campaign.ValidFrom,
		campaign.ValidUntil,
		campaign.CreatedAt,
	)
	return err
}

func (r *PricingCampaignRepository) UpdateStatus(ctx context.Context, id string, status entity.PricingCampaignStatus) error {
	commandTag, err := r.db.Exec(ctx, `
		UPDATE pricing_campaigns
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

var _ repository.PricingCampaignRepository = (*PricingCampaignRepository)(nil)
