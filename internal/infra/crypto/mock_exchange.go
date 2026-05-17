package crypto

import (
	"context"
	"fmt"

	"z-finance-api/internal/core/ports"
)

type MockExchangeGateway struct{}

func NewMockExchangeGateway() *MockExchangeGateway {
	return &MockExchangeGateway{}
}

func (m *MockExchangeGateway) Quote(ctx context.Context, asset string) (ports.ExchangeQuote, error) {
	switch asset {
	case "USDT":
		return ports.ExchangeQuote{Asset: asset, PriceInBRLCents: 500}, nil
	case "BTC":
		return ports.ExchangeQuote{Asset: asset, PriceInBRLCents: 25000000}, nil
	case "ETH":
		return ports.ExchangeQuote{Asset: asset, PriceInBRLCents: 1500000}, nil
	case "MATIC":
		return ports.ExchangeQuote{Asset: asset, PriceInBRLCents: 500}, nil
	default:
		return ports.ExchangeQuote{}, fmt.Errorf("asset nao suportado")
	}
}

func (m *MockExchangeGateway) Execute(ctx context.Context, asset string, amount int64, side string) (string, error) {
	return "mock-tx-hash", nil
}

var _ ports.ExchangeGateway = (*MockExchangeGateway)(nil)
