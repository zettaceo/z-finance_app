package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PricingCampaignRuleRepository struct {
	db *pgxpool.Pool
}

func NewPricingCampaignRuleRepository(pool *pgxpool.Pool) *PricingCampaignRuleRepository {
	return &PricingCampaignRuleRepository{db: pool}
}

func (r *PricingCampaignRuleRepository) ListByCampaign(ctx context.Context, campaignID string, planID string, userType entity.UserType, operation entity.PricingOperationType) ([]*entity.PricingCampaignRule, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, campaign_id, plan_id, user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee, created_at
		  FROM pricing_campaign_rules
		 WHERE campaign_id = $1
		   AND ($2::uuid IS NULL OR plan_id = $2 OR plan_id IS NULL)
		   AND user_type = $3
		   AND operation_type = $4
		 ORDER BY created_at DESC`,
		campaignID,
		nullIfEmpty(planID),
		userType,
		operation,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.PricingCampaignRule
	for rows.Next() {
		var item entity.PricingCampaignRule
		var planRef *string
		var minFee *int64
		var maxFee *int64
		if err := rows.Scan(
			&item.ID,
			&item.CampaignID,
			&planRef,
			&item.UserType,
			&item.OperationType,
			&item.Asset,
			&item.FeeType,
			&item.FeeValue,
			&minFee,
			&maxFee,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.PlanID = planRef
		item.MinFee = minFee
		item.MaxFee = maxFee
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *PricingCampaignRuleRepository) Create(ctx context.Context, rule *entity.PricingCampaignRule) error {
	if rule == nil {
		return repository.ErrInvalidState
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO pricing_campaign_rules (
			id, campaign_id, plan_id, user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`,
		rule.ID,
		rule.CampaignID,
		nullIfEmptyPtr(rule.PlanID),
		rule.UserType,
		rule.OperationType,
		rule.Asset,
		rule.FeeType,
		rule.FeeValue,
		rule.MinFee,
		rule.MaxFee,
		rule.CreatedAt,
	)
	return err
}

var _ repository.PricingCampaignRuleRepository = (*PricingCampaignRuleRepository)(nil)
