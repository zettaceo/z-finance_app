package usecase

import (
	"context"
	"errors"
	"time"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var (
	ErrDailyLimitExceeded    = errors.New("limite diario excedido")
	ErrMonthlyLimitExceeded  = errors.New("limite mensal excedido")
	ErrVelocityCountExceeded = errors.New("limite de transacoes excedido")
	ErrVelocityAmountExceeded = errors.New("limite de valor excedido")
)

type VelocityPolicy struct {
	MaxTxPerMinute     int64
	MaxTxPerHour       int64
	MaxAmountPerMinute int64
	MaxAmountPerHour   int64
}

func DefaultVelocityPolicy() VelocityPolicy {
	return VelocityPolicy{
		MaxTxPerMinute:     5,
		MaxTxPerHour:       50,
		MaxAmountPerMinute: 50_000_00,
		MaxAmountPerHour:   500_000_00,
	}
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

func NewRealClock() Clock {
	return realClock{}
}

type VelocityChecker struct {
	repo   repository.VelocityRepository
	policy VelocityPolicy
	clock  Clock
}

func NewVelocityChecker(repo repository.VelocityRepository, policy VelocityPolicy) *VelocityChecker {
	return &VelocityChecker{
		repo:   repo,
		policy: policy,
		clock:  realClock{},
	}
}

func (v *VelocityChecker) Check(ctx context.Context, userID string, amount int64) error {
	if amount <= 0 {
		return nil
	}

	level, err := v.repo.GetKYCLevel(ctx, userID)
	if err != nil {
		return err
	}

	dailyLimit, monthlyLimit, found, err := v.repo.GetLimitsForLevel(ctx, level)
	if err != nil {
		return err
	}
	if !found {
		dailyLimit, monthlyLimit = defaultLimits(level)
	}

	now := v.clock.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	if dailyLimit > 0 {
		dailySpent, err := v.repo.SumConfirmedSpentSince(ctx, userID, dayStart)
		if err != nil {
			return err
		}
		if dailySpent+amount > dailyLimit {
			return ErrDailyLimitExceeded
		}
	}

	if monthlyLimit > 0 {
		monthlySpent, err := v.repo.SumConfirmedSpentSince(ctx, userID, monthStart)
		if err != nil {
			return err
		}
		if monthlySpent+amount > monthlyLimit {
			return ErrMonthlyLimitExceeded
		}
	}

	if v.policy.MaxTxPerMinute > 0 || v.policy.MaxAmountPerMinute > 0 {
		oneMinuteAgo := now.Add(-1 * time.Minute)
		if v.policy.MaxTxPerMinute > 0 {
			count, err := v.repo.CountRecentTransactionsSince(ctx, userID, oneMinuteAgo)
			if err != nil {
				return err
			}
			if count+1 > v.policy.MaxTxPerMinute {
				return ErrVelocityCountExceeded
			}
		}
		if v.policy.MaxAmountPerMinute > 0 {
			total, err := v.repo.SumRecentNetAmountSince(ctx, userID, oneMinuteAgo)
			if err != nil {
				return err
			}
			if total+amount > v.policy.MaxAmountPerMinute {
				return ErrVelocityAmountExceeded
			}
		}
	}

	if v.policy.MaxTxPerHour > 0 || v.policy.MaxAmountPerHour > 0 {
		oneHourAgo := now.Add(-1 * time.Hour)
		if v.policy.MaxTxPerHour > 0 {
			count, err := v.repo.CountRecentTransactionsSince(ctx, userID, oneHourAgo)
			if err != nil {
				return err
			}
			if count+1 > v.policy.MaxTxPerHour {
				return ErrVelocityCountExceeded
			}
		}
		if v.policy.MaxAmountPerHour > 0 {
			total, err := v.repo.SumRecentNetAmountSince(ctx, userID, oneHourAgo)
			if err != nil {
				return err
			}
			if total+amount > v.policy.MaxAmountPerHour {
				return ErrVelocityAmountExceeded
			}
		}
	}

	return nil
}

func defaultLimits(level entity.KYCLevel) (int64, int64) {
	switch level {
	case entity.KYCLevelFull:
		return 5_000_000_00, 20_000_000_00
	case entity.KYCLevelBasic:
		return 500_000_00, 2_000_000_00
	default:
		return 100_000_00, 500_000_00
	}
}
