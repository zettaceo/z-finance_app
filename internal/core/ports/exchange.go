package ports

import "context"

type ExchangeQuote struct {
	Asset        string
	PriceInBRLCents int64
}

type ExchangeGateway interface {
	Quote(ctx context.Context, asset string) (ExchangeQuote, error)
	Execute(ctx context.Context, asset string, amount int64, side string) (string, error)
}
