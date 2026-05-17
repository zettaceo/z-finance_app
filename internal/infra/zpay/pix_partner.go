// Package zpay implements PixPartnerClient via the z-pay service HTTP API.
// Z-Finance uses z-pay as its SPI partner for PIX settlement.
package zpay

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

type PixPartnerClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewPixPartnerClient(baseURL, apiKey string) *PixPartnerClient {
	return &PixPartnerClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *PixPartnerClient) Send(ctx context.Context, transfer *entity.PixTransfer) error {
	payload := map[string]any{
		"id":              transfer.ID,
		"user_id":         transfer.UserID,
		"account_id":      transfer.AccountID,
		"amount":          transfer.Amount,
		"fee":             transfer.Fee,
		"net_amount":      transfer.NetAmount,
		"end_to_end_id":   transfer.EndToEndID,
		"external_ref":    transfer.ExternalRef,
		"idempotency_key": transfer.IdempotencyKey,
		"metadata":        transfer.Metadata,
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/pix/send", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("zpay: build pix request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("zpay: pix send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("zpay: pix send returned %d", resp.StatusCode)
	}

	return nil
}

var _ ports.PixPartnerClient = (*PixPartnerClient)(nil)
