package engine

import (
	"context"
	"strings"
	"time"

	"z-finance-api/internal/entity"
)

type PricingInput struct {
	UserID        string
	UserType      entity.UserType
	Plan          *entity.Plan
	OperationType entity.PricingOperationType
	Asset         string
	GrossAmount   int64
	Rules         []*entity.PricingRule
}

type PricingEngine interface {
	ResolveFee(ctx context.Context, input PricingInput) (entity.FeeResult, error)
}

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) ResolveFee(_ context.Context, input PricingInput) (entity.FeeResult, error) {
	amount := input.GrossAmount
	if amount <= 0 {
		return entity.FeeResult{Amount: amount, Fee: 0, NetAmount: 0}, nil
	}
	rule := SelectRule(input.Rules, input.OperationType, input.Asset)
	fee := int64(0)
	if rule != nil {
		fee = calculateFee(rule, amount)
	}
	if fee < 0 {
		fee = 0
	}
	if fee > amount {
		fee = amount
	}
	return entity.FeeResult{
		Amount:    amount,
		Fee:       fee,
		NetAmount: amount - fee,
	}, nil
}

func SelectRule(rules []*entity.PricingRule, operation entity.PricingOperationType, asset string) *entity.PricingRule {
	if len(rules) == 0 {
		return nil
	}
	normalizedAsset := normalizeAsset(asset)
	var best *entity.PricingRule
	bestScore := 0
	bestCreated := time.Time{}
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if rule.OperationType != operation {
			continue
		}
		score := assetScore(rule.Asset, normalizedAsset)
		if score == 0 {
			continue
		}
		if best == nil || score > bestScore || (score == bestScore && rule.CreatedAt.After(bestCreated)) {
			best = rule
			bestScore = score
			bestCreated = rule.CreatedAt
		}
	}
	return best
}

func assetScore(ruleAsset, requested string) int {
	normalized := normalizeAsset(ruleAsset)
	if normalized == "" {
		return 0
	}
	if normalized == entity.PricingAssetAny {
		return 1
	}
	if requested != "" && normalized == requested {
		return 2
	}
	return 0
}

func normalizeAsset(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func calculateFee(rule *entity.PricingRule, amount int64) int64 {
	if rule == nil {
		return 0
	}
	fee := int64(0)
	switch rule.FeeType {
	case entity.PricingFeePercentage:
		fee = (amount * rule.FeeValue) / 10000
	case entity.PricingFeeFixed:
		fee = rule.FeeValue
	default:
		fee = 0
	}
	if rule.MinFee != nil && fee < *rule.MinFee {
		fee = *rule.MinFee
	}
	if rule.MaxFee != nil && fee > *rule.MaxFee {
		fee = *rule.MaxFee
	}
	return fee
}
