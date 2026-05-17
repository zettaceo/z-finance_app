package custody

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"

	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
)

type MockConfig struct {
	MinLatency  time.Duration
	MaxLatency  time.Duration
	ErrorRate   float64
	TimeoutRate float64
}

type MockGateway struct {
	cfg  MockConfig
	rand *rand.Rand
	mu   sync.Mutex
}

func DefaultMockConfig() MockConfig {
	return MockConfig{
		MinLatency:  40 * time.Millisecond,
		MaxLatency:  180 * time.Millisecond,
		ErrorRate:   0.05,
		TimeoutRate: 0.02,
	}
}

func NewMockGateway(cfg MockConfig) *MockGateway {
	if cfg.MinLatency <= 0 {
		cfg.MinLatency = DefaultMockConfig().MinLatency
	}
	if cfg.MaxLatency < cfg.MinLatency {
		cfg.MaxLatency = cfg.MinLatency
	}
	return &MockGateway{
		cfg:  cfg,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *MockGateway) CreateDepositAddress(ctx context.Context, userID, asset, network string) (entity.CustodyAddress, error) {
	if userID == "" || asset == "" || network == "" {
		return entity.CustodyAddress{}, errors.New("dados invalidos")
	}
	if err := m.simulate(ctx); err != nil {
		return entity.CustodyAddress{}, err
	}
	address := fmt.Sprintf("%s:%s", asset, uuid.NewString())
	return entity.CustodyAddress{
		ProviderID: uuid.NewString(),
		Network:    network,
		Asset:      asset,
		Address:    address,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

func (m *MockGateway) SendTransfer(ctx context.Context, transfer entity.CustodyTransfer) (entity.CustodyTransfer, error) {
	if transfer.UserID == "" || transfer.Asset == "" || transfer.Network == "" || transfer.Address == "" || transfer.Amount <= 0 {
		return entity.CustodyTransfer{}, errors.New("transferencia invalida")
	}
	if err := m.simulate(ctx); err != nil {
		return entity.CustodyTransfer{}, err
	}
	transfer.ProviderID = uuid.NewString()
	transfer.Status = entity.CustodyTransferConfirmed
	transfer.CreatedAt = time.Now().UTC()
	return transfer, nil
}

func (m *MockGateway) GetTransferStatus(ctx context.Context, providerID string) (entity.CustodyTransferStatus, error) {
	if providerID == "" {
		return "", errors.New("provider_id invalido")
	}
	if err := m.simulate(ctx); err != nil {
		return "", err
	}
	return entity.CustodyTransferConfirmed, nil
}

func (m *MockGateway) simulate(ctx context.Context) error {
	latency := m.randomLatency()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(latency):
	}
	if m.shouldTimeout() {
		return context.DeadlineExceeded
	}
	if m.shouldError() {
		return errors.New("custodia indisponivel")
	}
	return nil
}

func (m *MockGateway) randomLatency() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cfg.MaxLatency <= m.cfg.MinLatency {
		return m.cfg.MinLatency
	}
	diff := m.cfg.MaxLatency - m.cfg.MinLatency
	return m.cfg.MinLatency + time.Duration(m.rand.Int63n(int64(diff)))
}

func (m *MockGateway) shouldError() bool {
	if m.cfg.ErrorRate <= 0 {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.rand.Float64() < m.cfg.ErrorRate
}

func (m *MockGateway) shouldTimeout() bool {
	if m.cfg.TimeoutRate <= 0 {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.rand.Float64() < m.cfg.TimeoutRate
}

var _ ports.CustodyGateway = (*MockGateway)(nil)
