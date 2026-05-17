// Package zswap implements ExchangeGateway via the z-swap service HTTP API.
// Replace the mock in cmd/main.go with this when Z_SWAP_URL is set.
package zswap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"z-finance-api/internal/core/ports"
)

type ExchangeGateway struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewExchangeGateway(baseURL, apiKey string) *ExchangeGateway {
	return &ExchangeGateway{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (g *ExchangeGateway) Quote(ctx context.Context, asset string) (ports.ExchangeQuote, error) {
	url := fmt.Sprintf("%s/api/v1/quote?asset=%s&currency=BRL", g.baseURL, asset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ports.ExchangeQuote{}, fmt.Errorf("zswap: build request: %w", err)
	}
	req.Header.Set("X-Service-Key", g.apiKey)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return ports.ExchangeQuote{}, fmt.Errorf("zswap: quote request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ports.ExchangeQuote{}, fmt.Errorf("zswap: quote returned %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Asset           string `json:"asset"`
			PriceInBRLCents int64  `json:"price_in_brl_cents"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return ports.ExchangeQuote{}, fmt.Errorf("zswap: decode quote: %w", err)
	}

	return ports.ExchangeQuote{
		Asset:           body.Data.Asset,
		PriceInBRLCents: body.Data.PriceInBRLCents,
	}, nil
}

func (g *ExchangeGateway) Execute(ctx context.Context, asset string, amount int64, side string) (string, error) {
	payload := map[string]any{
		"asset":  asset,
		"amount": amount,
		"side":   side,
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/v1/execute", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("zswap: build execute request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", g.apiKey)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("zswap: execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("zswap: execute returned %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			TxHash string `json:"tx_hash"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("zswap: decode execute: %w", err)
	}

	return body.Data.TxHash, nil
}

var _ ports.ExchangeGateway = (*ExchangeGateway)(nil)
