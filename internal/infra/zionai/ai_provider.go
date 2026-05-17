// Package zionai implements AIProvider via the Zion AI service HTTP API.
// Zion AI only proposes ExecutionIntents — it NEVER executes balance changes directly.
// All intent execution goes through Z-Finance use cases after explicit user confirmation.
package zionai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"z-finance-api/internal/core/ports"
)

type AIProvider struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewAIProvider(baseURL, apiKey string) *AIProvider {
	return &AIProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *AIProvider) Analyze(ctx context.Context, userID string, input map[string]any) (map[string]any, error) {
	payload := map[string]any{
		"user_id": userID,
		"input":   input,
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/analyze", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("zion-ai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zion-ai: analyze request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zion-ai: analyze returned %d", resp.StatusCode)
	}

	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("zion-ai: decode analyze: %w", err)
	}
	return body.Data, nil
}

func (p *AIProvider) ProposeIntent(ctx context.Context, userID string, input map[string]any) (ports.ExecutionIntent, error) {
	payload := map[string]any{
		"user_id": userID,
		"input":   input,
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/propose-intent", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("zion-ai: build propose request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zion-ai: propose intent request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zion-ai: propose intent returned %d", resp.StatusCode)
	}

	var body struct {
		Data intentDTO `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("zion-ai: decode propose intent: %w", err)
	}
	return &body.Data, nil
}

// intentDTO is the wire format for an ExecutionIntent from Zion AI.
type intentDTO struct {
	IntentID   string         `json:"id"`
	IntentUser string         `json:"user_id"`
	IntentKind string         `json:"kind"`
	IntentData map[string]any `json:"payload"`
}

func (i *intentDTO) ID() string              { return i.IntentID }
func (i *intentDTO) UserID() string          { return i.IntentUser }
func (i *intentDTO) Kind() string            { return i.IntentKind }
func (i *intentDTO) Payload() map[string]any { return i.IntentData }

var _ ports.AIProvider = (*AIProvider)(nil)
var _ ports.ExecutionIntent = (*intentDTO)(nil)
