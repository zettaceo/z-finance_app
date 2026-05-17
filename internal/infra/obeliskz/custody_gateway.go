// Package obeliskz implements CustodyGateway via the obelisk-z service HTTP API.
// Replace the mock in cmd/main.go with this when OBELISK_Z_URL is set.
package obeliskz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
)

type CustodyGateway struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewCustodyGateway(baseURL, apiKey string) *CustodyGateway {
	return &CustodyGateway{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (g *CustodyGateway) do(ctx context.Context, method, path string, payload any) (*http.Response, error) {
	var b []byte
	if payload != nil {
		var err error
		b, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, g.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", g.apiKey)
	return g.httpClient.Do(req)
}

func (g *CustodyGateway) CreateDepositAddress(ctx context.Context, userID, asset, network string) (entity.CustodyAddress, error) {
	resp, err := g.do(ctx, http.MethodPost, "/api/v1/addresses", map[string]any{
		"user_id": userID,
		"asset":   asset,
		"network": network,
	})
	if err != nil {
		return entity.CustodyAddress{}, fmt.Errorf("obelisk-z: create address: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return entity.CustodyAddress{}, fmt.Errorf("obelisk-z: create address returned %d", resp.StatusCode)
	}

	var body struct {
		Data entity.CustodyAddress `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return entity.CustodyAddress{}, fmt.Errorf("obelisk-z: decode address: %w", err)
	}
	return body.Data, nil
}

func (g *CustodyGateway) SendTransfer(ctx context.Context, transfer entity.CustodyTransfer) (entity.CustodyTransfer, error) {
	resp, err := g.do(ctx, http.MethodPost, "/api/v1/transfers", transfer)
	if err != nil {
		return entity.CustodyTransfer{}, fmt.Errorf("obelisk-z: send transfer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return entity.CustodyTransfer{}, fmt.Errorf("obelisk-z: send transfer returned %d", resp.StatusCode)
	}

	var body struct {
		Data entity.CustodyTransfer `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return entity.CustodyTransfer{}, fmt.Errorf("obelisk-z: decode transfer: %w", err)
	}
	return body.Data, nil
}

func (g *CustodyGateway) GetTransferStatus(ctx context.Context, providerID string) (entity.CustodyTransferStatus, error) {
	resp, err := g.do(ctx, http.MethodGet, "/api/v1/transfers/"+providerID+"/status", nil)
	if err != nil {
		return "", fmt.Errorf("obelisk-z: get status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("obelisk-z: get status returned %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Status entity.CustodyTransferStatus `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("obelisk-z: decode status: %w", err)
	}
	return body.Data.Status, nil
}

var _ ports.CustodyGateway = (*CustodyGateway)(nil)
