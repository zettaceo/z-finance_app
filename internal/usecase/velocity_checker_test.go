package usecase

import (
	"context"
	"testing"
	"time"

	"z-finance-api/internal/entity"
)

type fakeVelocityRepo struct {
	kycLevels       map[string]entity.KYCLevel
	dailyLimits     map[entity.KYCLevel]int64
	monthlyLimits   map[entity.KYCLevel]int64
	dailySpent      map[string]int64
	monthlySpent    map[string]int64
	recentTxCount   map[string]int64
	recentNetAmount map[string]int64
}

func (f *fakeVelocityRepo) GetKYCLevel(_ context.Context, userID string) (entity.KYCLevel, error) {
	if level, ok := f.kycLevels[userID]; ok {
		return level, nil
	}
	return entity.KYCLevelBasic, nil
}

func (f *fakeVelocityRepo) GetLimitsForLevel(_ context.Context, level entity.KYCLevel) (int64, int64, bool, error) {
	daily, dOk := f.dailyLimits[level]
	monthly, mOk := f.monthlyLimits[level]
	return daily, monthly, dOk || mOk, nil
}

func (f *fakeVelocityRepo) SumConfirmedSpentSince(_ context.Context, userID string, _ time.Time) (int64, error) {
	return f.dailySpent[userID], nil
}

func (f *fakeVelocityRepo) CountRecentTransactionsSince(_ context.Context, userID string, _ time.Time) (int64, error) {
	return f.recentTxCount[userID], nil
}

func (f *fakeVelocityRepo) SumRecentNetAmountSince(_ context.Context, userID string, _ time.Time) (int64, error) {
	return f.recentNetAmount[userID], nil
}

func TestVelocityChecker_DailyLimitExceeded(t *testing.T) {
	repo := &fakeVelocityRepo{
		kycLevels:     map[string]entity.KYCLevel{"user-1": entity.KYCLevelBasic},
		dailyLimits:   map[entity.KYCLevel]int64{entity.KYCLevelBasic: 100},
		monthlyLimits: map[entity.KYCLevel]int64{entity.KYCLevelBasic: 1000},
		dailySpent:    map[string]int64{"user-1": 90},
		monthlySpent:  map[string]int64{"user-1": 200},
	}
	checker := NewVelocityChecker(repo, VelocityPolicy{})
	if err := checker.Check(context.Background(), "user-1", 20); err != ErrDailyLimitExceeded {
		t.Fatalf("esperado ErrDailyLimitExceeded, obtido %v", err)
	}
}

func TestVelocityChecker_VelocityCountExceeded(t *testing.T) {
	repo := &fakeVelocityRepo{
		kycLevels:     map[string]entity.KYCLevel{"user-2": entity.KYCLevelBasic},
		dailyLimits:   map[entity.KYCLevel]int64{entity.KYCLevelBasic: 0},
		monthlyLimits: map[entity.KYCLevel]int64{entity.KYCLevelBasic: 0},
		recentTxCount: map[string]int64{"user-2": 2},
	}
	policy := VelocityPolicy{MaxTxPerMinute: 2}
	checker := NewVelocityChecker(repo, policy)
	if err := checker.Check(context.Background(), "user-2", 10); err != ErrVelocityCountExceeded {
		t.Fatalf("esperado ErrVelocityCountExceeded, obtido %v", err)
	}
}

func TestVelocityChecker_MultiUserIsolation(t *testing.T) {
	repo := &fakeVelocityRepo{
		kycLevels:     map[string]entity.KYCLevel{"user-a": entity.KYCLevelBasic, "user-b": entity.KYCLevelBasic},
		dailyLimits:   map[entity.KYCLevel]int64{entity.KYCLevelBasic: 100},
		monthlyLimits: map[entity.KYCLevel]int64{entity.KYCLevelBasic: 0},
		dailySpent:    map[string]int64{"user-a": 10, "user-b": 99},
	}
	checker := NewVelocityChecker(repo, VelocityPolicy{})
	if err := checker.Check(context.Background(), "user-a", 20); err != nil {
		t.Fatalf("nao esperava bloqueio para user-a: %v", err)
	}
	if err := checker.Check(context.Background(), "user-b", 10); err != ErrDailyLimitExceeded {
		t.Fatalf("esperado ErrDailyLimitExceeded para user-b, obtido %v", err)
	}
}
