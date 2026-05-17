package usecase

import (
	"context"
	"errors"
	"strings"

	"z-finance-api/internal/entity"
	pricingengine "z-finance-api/internal/pricing/engine"
	"z-finance-api/internal/repository"
)

var ErrPricingInvalidInput = errors.New("dados de pricing invalidos")
var ErrPricingFeatureDisabled = errors.New("feature desabilitada para o plano")

type PricingInput struct {
	UserID        string
	UserType      entity.UserType
	OperationType entity.PricingOperationType
	Asset         string
	GrossAmount   int64
	FeatureCode   string
}

type ResolvePricingUseCase struct {
	users        repository.UserRepository
	plans        repository.PlanRepository
	userPlans    repository.UserPlanRepository
	rules        repository.PricingRuleRepository
	versions     repository.PricingVersionRepository
	planFeatures repository.PlanFeatureRepository
	campaigns    repository.PricingCampaignRepository
	campaignRules repository.PricingCampaignRuleRepository
	engine       pricingengine.PricingEngine
	clock        Clock
}

func NewResolvePricingUseCase(
	users repository.UserRepository,
	plans repository.PlanRepository,
	userPlans repository.UserPlanRepository,
	rules repository.PricingRuleRepository,
	versions repository.PricingVersionRepository,
	planFeatures repository.PlanFeatureRepository,
	campaigns repository.PricingCampaignRepository,
	campaignRules repository.PricingCampaignRuleRepository,
	engine pricingengine.PricingEngine,
) *ResolvePricingUseCase {
	return &ResolvePricingUseCase{
		users:     users,
		plans:     plans,
		userPlans: userPlans,
		rules:     rules,
		versions:  versions,
		planFeatures: planFeatures,
		campaigns: campaigns,
		campaignRules: campaignRules,
		engine:    engine,
		clock:     NewRealClock(),
	}
}

func (uc *ResolvePricingUseCase) Execute(ctx context.Context, input PricingInput) (entity.FeeResult, *entity.PricingRule, error) {
	if input.UserID == "" || input.OperationType == "" || input.GrossAmount <= 0 {
		return entity.FeeResult{}, nil, ErrPricingInvalidInput
	}
	userType := uc.resolveUserType(ctx, input)
	plan := uc.resolvePlan(ctx, input.UserID)
	if !uc.isFeatureEnabled(ctx, plan, input.FeatureCode) {
		return entity.FeeResult{}, nil, ErrPricingFeatureDisabled
	}
	version := uc.resolvePricingVersion(ctx)
	rules := uc.resolveRules(ctx, plan, userType, input.OperationType, version)
	asset := strings.ToUpper(strings.TrimSpace(input.Asset))
	rule := uc.resolveCampaignRule(ctx, plan, userType, input.OperationType, asset)
	if rule == nil {
		rule = pricingengine.SelectRule(rules, input.OperationType, asset)
	}
	if uc.engine == nil {
		return entity.FeeResult{Amount: input.GrossAmount, Fee: 0, NetAmount: input.GrossAmount}, rule, nil
	}
	rulesToUse := rules
	if rule != nil {
		rulesToUse = append([]*entity.PricingRule{rule}, rules...)
	}
	result, err := uc.engine.ResolveFee(ctx, pricingengine.PricingInput{
		UserID:        input.UserID,
		UserType:      userType,
		Plan:          plan,
		OperationType: input.OperationType,
		Asset:         asset,
		GrossAmount:   input.GrossAmount,
		Rules:         rulesToUse,
	})
	if err != nil {
		return entity.FeeResult{}, rule, err
	}
	return result, rule, nil
}

func (uc *ResolvePricingUseCase) resolveUserType(ctx context.Context, input PricingInput) entity.UserType {
	if input.UserType != "" {
		return input.UserType
	}
	if uc.users == nil {
		return entity.UserTypePF
	}
	user, err := uc.users.GetByID(ctx, input.UserID)
	if err != nil || user == nil || user.UserType == "" {
		return entity.UserTypePF
	}
	return user.UserType
}

func (uc *ResolvePricingUseCase) resolvePlan(ctx context.Context, userID string) *entity.Plan {
	now := uc.clock.Now()
	if uc.userPlans != nil {
		plan, err := uc.userPlans.GetActiveByUser(ctx, userID, now)
		if err == nil && plan != nil && uc.plans != nil {
			item, planErr := uc.plans.GetByID(ctx, plan.PlanID)
			if planErr == nil && item != nil {
				return item
			}
		}
	}
	if uc.plans == nil {
		return &entity.Plan{Code: "FREE"}
	}
	free, err := uc.plans.GetByCode(ctx, "FREE")
	if err == nil && free != nil {
		return free
	}
	return &entity.Plan{Code: "FREE"}
}

func (uc *ResolvePricingUseCase) resolveRules(ctx context.Context, plan *entity.Plan, userType entity.UserType, operation entity.PricingOperationType, version *entity.PricingVersion) []*entity.PricingRule {
	if uc.rules == nil || plan == nil || plan.ID == "" {
		return nil
	}
	versionID := ""
	if version != nil {
		versionID = version.ID
	}
	items, err := uc.rules.ListByPlanAndUserType(ctx, plan.ID, userType, operation, versionID)
	if err != nil {
		return nil
	}
	return items
}

func (uc *ResolvePricingUseCase) resolvePricingVersion(ctx context.Context) *entity.PricingVersion {
	if uc.versions == nil {
		return nil
	}
	version, err := uc.versions.GetActive(ctx, uc.clock.Now())
	if err != nil {
		return nil
	}
	return version
}

func (uc *ResolvePricingUseCase) isFeatureEnabled(ctx context.Context, plan *entity.Plan, featureCode string) bool {
	if featureCode == "" || uc.planFeatures == nil || plan == nil {
		return true
	}
	features, err := uc.planFeatures.ListByPlan(ctx, plan.ID)
	if err != nil || len(features) == 0 {
		return true
	}
	code := strings.ToUpper(strings.TrimSpace(featureCode))
	for _, item := range features {
		if item != nil && strings.EqualFold(item.FeatureCode, code) {
			return item.Enabled
		}
	}
	return true
}

func (uc *ResolvePricingUseCase) resolveCampaignRule(ctx context.Context, plan *entity.Plan, userType entity.UserType, operation entity.PricingOperationType, asset string) *entity.PricingRule {
	if uc.campaigns == nil || uc.campaignRules == nil || plan == nil {
		return nil
	}
	campaigns, err := uc.campaigns.ListActiveByPlan(ctx, plan.ID, uc.clock.Now())
	if err != nil || len(campaigns) == 0 {
		return nil
	}
	for _, campaign := range campaigns {
		rules, err := uc.campaignRules.ListByCampaign(ctx, campaign.ID, plan.ID, userType, operation)
		if err != nil || len(rules) == 0 {
			continue
		}
		if rule := selectCampaignRule(rules, operation, asset); rule != nil {
			return rule
		}
	}
	return nil
}

func selectCampaignRule(rules []*entity.PricingCampaignRule, operation entity.PricingOperationType, asset string) *entity.PricingRule {
	normalized := strings.ToUpper(strings.TrimSpace(asset))
	var best *entity.PricingCampaignRule
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if rule.OperationType != operation {
			continue
		}
		if normalized != entity.PricingAssetAny && rule.Asset != normalized && rule.Asset != entity.PricingAssetAny {
			continue
		}
		if best == nil {
			best = rule
			continue
		}
		if best.Asset == entity.PricingAssetAny && rule.Asset != entity.PricingAssetAny {
			best = rule
		}
	}
	if best == nil {
		return nil
	}
	return &entity.PricingRule{
		ID:            best.ID,
		PlanID:        planIDFromCampaignRule(best),
		UserType:      best.UserType,
		OperationType: best.OperationType,
		Asset:         best.Asset,
		FeeType:       best.FeeType,
		FeeValue:      best.FeeValue,
		MinFee:        best.MinFee,
		MaxFee:        best.MaxFee,
		CreatedAt:     best.CreatedAt,
	}
}

func planIDFromCampaignRule(rule *entity.PricingCampaignRule) string {
	if rule.PlanID == nil {
		return ""
	}
	return *rule.PlanID
}
