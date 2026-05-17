package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type PricingRuleRepository struct {
	db dbExecutor
}

func NewPricingRuleRepository(pool *pgxpool.Pool) *PricingRuleRepository {
	return &PricingRuleRepository{db: pool}
}

func NewPricingRuleRepositoryWithTx(tx dbExecutor) *PricingRuleRepository {
	return &PricingRuleRepository{db: tx}
}

func (r *PricingRuleRepository) ListByPlanAndUserType(ctx context.Context, planID string, userType entity.UserType, operation entity.PricingOperationType, pricingVersionID string) ([]*entity.PricingRule, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, plan_id, pricing_version_id, user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee, created_at
		  FROM pricing_rules
		 WHERE plan_id = $1
		   AND user_type = $2
		   AND operation_type = $3
		   AND ($4::uuid IS NULL OR pricing_version_id = $4 OR pricing_version_id IS NULL)
		 ORDER BY created_at DESC`, planID, userType, operation, nullIfEmpty(pricingVersionID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.PricingRule
	for rows.Next() {
		var rule entity.PricingRule
		var minFee *int64
		var maxFee *int64
		if err := rows.Scan(
			&rule.ID,
			&rule.PlanID,
			&rule.PricingVersionID,
			&rule.UserType,
			&rule.OperationType,
			&rule.Asset,
			&rule.FeeType,
			&rule.FeeValue,
			&minFee,
			&maxFee,
			&rule.CreatedAt,
		); err != nil {
			return nil, err
		}
		rule.MinFee = minFee
		rule.MaxFee = maxFee
		items = append(items, &rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

var _ repository.PricingRuleRepository = (*PricingRuleRepository)(nil)

func (r *PricingRuleRepository) List(ctx context.Context, filter repository.PricingRuleFilter) ([]*entity.PricingRule, error) {
	query := `
		SELECT id, plan_id, pricing_version_id, user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee, created_at
		  FROM pricing_rules
		 WHERE ($1::uuid IS NULL OR plan_id = $1)
		   AND ($2 = '' OR user_type = $2)
		   AND ($3 = '' OR operation_type = $3)
		   AND ($4 = '' OR asset = $4)
		   AND ($5::uuid IS NULL OR pricing_version_id = $5)
		 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query,
		nullIfEmpty(filter.PlanID),
		string(filter.UserType),
		string(filter.OperationType),
		filter.Asset,
		nullIfEmpty(filter.PricingVersionID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.PricingRule
	for rows.Next() {
		var rule entity.PricingRule
		var minFee *int64
		var maxFee *int64
		if err := rows.Scan(
			&rule.ID,
			&rule.PlanID,
			&rule.PricingVersionID,
			&rule.UserType,
			&rule.OperationType,
			&rule.Asset,
			&rule.FeeType,
			&rule.FeeValue,
			&minFee,
			&maxFee,
			&rule.CreatedAt,
		); err != nil {
			return nil, err
		}
		rule.MinFee = minFee
		rule.MaxFee = maxFee
		items = append(items, &rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PricingRuleRepository) Create(ctx context.Context, rule *entity.PricingRule) error {
	if rule == nil {
		return repository.ErrInvalidState
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO pricing_rules (id, plan_id, pricing_version_id, user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		rule.ID,
		rule.PlanID,
		nullIfEmpty(rule.PricingVersionID),
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

func (r *PricingRuleRepository) Update(ctx context.Context, rule *entity.PricingRule) error {
	if rule == nil {
		return repository.ErrInvalidState
	}
	commandTag, err := r.db.Exec(ctx, `
		UPDATE pricing_rules
		   SET plan_id = $1,
		       pricing_version_id = $2,
		       user_type = $3,
		       operation_type = $4,
		       asset = $5,
		       fee_type = $6,
		       fee_value = $7,
		       min_fee = $8,
		       max_fee = $9
		 WHERE id = $10`,
		rule.PlanID,
		nullIfEmpty(rule.PricingVersionID),
		rule.UserType,
		rule.OperationType,
		rule.Asset,
		rule.FeeType,
		rule.FeeValue,
		rule.MinFee,
		rule.MaxFee,
		rule.ID,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
