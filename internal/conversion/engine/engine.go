package engine

import (
	"context"
	"strings"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type Result struct {
	TargetAsset string
	Percentage  int32
}

type Engine struct {
	rules    repository.ConversionRuleRepository
	settings repository.UserSettingsRepository
}

func NewEngine(rules repository.ConversionRuleRepository, settings repository.UserSettingsRepository) *Engine {
	return &Engine{rules: rules, settings: settings}
}

func (e *Engine) ResolvePixIn(ctx context.Context, userID string) (*Result, error) {
	if e.rules != nil {
		rule, err := e.rules.GetActiveByUserAndTrigger(ctx, userID, entity.ConversionTriggerPixIn)
		if err != nil {
			return nil, err
		}
		if rule != nil && rule.Enabled {
			return &Result{TargetAsset: rule.TargetAsset, Percentage: 100}, nil
		}
	}

	if e.settings != nil {
		settings, err := e.settings.GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if settings.AutoConvertEnabled && settings.AutoConvertAsset != "" {
			return &Result{TargetAsset: settings.AutoConvertAsset, Percentage: 100}, nil
		}
		if settings.AutoConvertPixIn && settings.PixInTargetAsset != "" {
			return &Result{TargetAsset: settings.PixInTargetAsset, Percentage: settings.PixInPercentage}, nil
		}
	}

	return nil, nil
}

func (e *Engine) ResolveAssets(ctx context.Context, userID string, trigger entity.ConversionTrigger) ([]string, error) {
	assets := defaultAssets()
	if e == nil {
		return assets, nil
	}

	var preferred []string
	if e.settings != nil {
		settings, err := e.settings.GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if settings != nil {
			preferred = normalizeAssets(settings.ConversionPriority)
			fallback := normalizeAsset(settings.FallbackAsset)
			if fallback != "" {
				preferred = prependIfMissing(preferred, fallback)
			}
		}
	}
	if len(preferred) == 0 {
		preferred = assets
	}

	if e.rules != nil {
		rule, err := e.rules.GetActiveByUserAndTrigger(ctx, userID, trigger)
		if err != nil {
			return nil, err
		}
		if rule != nil && rule.Enabled {
			target := normalizeAsset(rule.TargetAsset)
			if target != "" {
				preferred = prependIfMissing(preferred, target)
			}
		}
	}
	return preferred, nil
}

func defaultAssets() []string {
	return []string{"USDT", "ETH", "BTC", "MATIC"}
}

func normalizeAssets(values []string) []string {
	cleaned := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		normalized := normalizeAsset(value)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		cleaned = append(cleaned, normalized)
	}
	return cleaned
}

func normalizeAsset(value string) string {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	return normalized
}

func prependIfMissing(values []string, value string) []string {
	for _, item := range values {
		if item == value {
			return values
		}
	}
	return append([]string{value}, values...)
}
