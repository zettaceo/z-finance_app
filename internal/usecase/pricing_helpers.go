package usecase

import "z-finance-api/internal/entity"

func ruleID(rule *entity.PricingRule) string {
	if rule == nil {
		return ""
	}
	return rule.ID
}
