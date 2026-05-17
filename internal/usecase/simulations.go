package usecase

import (
	"context"
	"time"

	"z-finance-api/internal/core/ports"
)

type SimulationResult struct {
	Name    string
	Status  string
	Error   string
	Details map[string]any
}

func SimulateHighVolume(ctx context.Context, checker *VelocityChecker, userID string, amount int64, iterations int) SimulationResult {
	result := SimulationResult{
		Name:   "high_volume",
		Status: "ok",
		Details: map[string]any{
			"user_id":    userID,
			"amount":     amount,
			"iterations": iterations,
		},
	}
	if checker == nil || iterations <= 0 || amount <= 0 {
		result.Status = "skipped"
		return result
	}
	for i := 0; i < iterations; i++ {
		if err := checker.Check(ctx, userID, amount); err != nil {
			result.Status = "blocked"
			result.Error = err.Error()
			result.Details["blocked_at"] = i + 1
			return result
		}
	}
	return result
}

func SimulateConcurrentUsers(ctx context.Context, checker *VelocityChecker, userIDs []string, amount int64) SimulationResult {
	result := SimulationResult{
		Name:   "concurrent_users",
		Status: "ok",
		Details: map[string]any{
			"user_count": len(userIDs),
			"amount":     amount,
		},
	}
	if checker == nil || len(userIDs) == 0 || amount <= 0 {
		result.Status = "skipped"
		return result
	}
	failures := 0
	for _, userID := range userIDs {
		if err := checker.Check(ctx, userID, amount); err != nil {
			failures++
		}
	}
	if failures > 0 {
		result.Status = "blocked"
		result.Details["failures"] = failures
	}
	return result
}

func SimulateExchangeFailure(ctx context.Context, exchange ports.ExchangeGateway, asset string) SimulationResult {
	result := SimulationResult{
		Name:   "exchange_failure",
		Status: "ok",
		Details: map[string]any{
			"asset": asset,
		},
	}
	if exchange == nil || asset == "" {
		result.Status = "skipped"
		return result
	}
	if _, err := exchange.Quote(ctx, asset); err != nil {
		result.Status = "detected"
		result.Error = err.Error()
		return result
	}
	if _, err := exchange.Execute(ctx, asset, 1, "BUY"); err != nil {
		result.Status = "detected"
		result.Error = err.Error()
		return result
	}
	result.Status = "no_error"
	return result
}

func SimulateConfirmationDelay(delay time.Duration) SimulationResult {
	return SimulationResult{
		Name:   "confirmation_delay",
		Status: "simulated",
		Details: map[string]any{
			"delay_ms": delay.Milliseconds(),
		},
	}
}

func SimulatePartialConversion(stepCount int, failAt int) SimulationResult {
	result := SimulationResult{
		Name:   "partial_conversion",
		Status: "ok",
		Details: map[string]any{
			"steps":   stepCount,
			"fail_at": failAt,
		},
	}
	if stepCount <= 0 {
		result.Status = "skipped"
		return result
	}
	if failAt > 0 && failAt <= stepCount {
		result.Status = "partial"
		return result
	}
	return result
}
