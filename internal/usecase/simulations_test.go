package usecase

import (
	"context"
	"errors"
	"testing"

	"z-finance-api/internal/core/ports"
)

type failingExchange struct {
	quoteErr error
	execErr  error
}

func (f failingExchange) Quote(_ context.Context, asset string) (ports.ExchangeQuote, error) {
	if f.quoteErr != nil {
		return ports.ExchangeQuote{}, f.quoteErr
	}
	return ports.ExchangeQuote{Asset: asset, PriceInBRLCents: 100}, nil
}

func (f failingExchange) Execute(_ context.Context, _ string, _ int64, _ string) (string, error) {
	if f.execErr != nil {
		return "", f.execErr
	}
	return "ok", nil
}

func TestSimulateExchangeFailure(t *testing.T) {
	result := SimulateExchangeFailure(context.Background(), failingExchange{quoteErr: errors.New("falha quote")}, "BTC")
	if result.Status != "detected" {
		t.Fatalf("esperado status detected, obtido %s", result.Status)
	}
	if result.Error == "" {
		t.Fatalf("esperado erro preenchido")
	}
}
