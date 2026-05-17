package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"z-finance-api/internal/testutil"
	"z-finance-api/internal/usecase"
)

type fakeWebhookRepo struct {
	replay bool
}

func (f fakeWebhookRepo) ConfirmPayment(context.Context, string, string) error { return nil }
func (f fakeWebhookRepo) RejectPayment(context.Context, string, string) error  { return nil }
func (f fakeWebhookRepo) ConfirmCardAuthorization(context.Context, string, string) error {
	return nil
}
func (f fakeWebhookRepo) RejectCardAuthorization(context.Context, string, string) error { return nil }
func (f fakeWebhookRepo) EnsureEvent(context.Context, string, string) (bool, error) {
	return f.replay, nil
}

func webhookTestHandler(secret string, repo fakeWebhookRepo, allowedIPs []string, rateLimit int) http.Handler {
	core := NewHandler(Dependencies{
		WebhookSecret:            secret,
		WebhookRepo:              repo,
		WebhookAllowedIPs:        allowedIPs,
		WebhookRateLimitPerMinute: rateLimit,
	})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(r.Context(), w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return withWebhookSecurity(core, validateWebhookSignature(secret, repo, nil, next))
}

func TestRequireAuthMissingToken(t *testing.T) {
	tokenService := usecase.NewTokenService("test-secret", time.Minute)
	core := NewHandler(Dependencies{TokenService: tokenService})
	handler := BuildHTTPHandler(core)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "TOKEN_MISSING" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	tokenService := usecase.NewTokenService("test-secret", time.Minute)
	core := NewHandler(Dependencies{TokenService: tokenService})
	handler := BuildHTTPHandler(core)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "TOKEN_INVALID" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestRequireIdempotencyOnTransactionReverse(t *testing.T) {
	tokenService := usecase.NewTokenService("test-secret", time.Minute)
	accessToken, _, err := tokenService.GenerateAccessToken("user-1")
	if err != nil {
		t.Fatalf("falha ao gerar token: %v", err)
	}
	core := NewHandler(Dependencies{TokenService: tokenService})
	handler := BuildHTTPHandler(core)

	req := httptest.NewRequest(http.MethodPost, "/transactions/reverse", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "IDEMPOTENCY_REQUIRED" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestRequireIdempotencyOnInvoiceCreate(t *testing.T) {
	tokenService := usecase.NewTokenService("test-secret", time.Minute)
	accessToken, _, err := tokenService.GenerateAccessToken("user-1")
	if err != nil {
		t.Fatalf("falha ao gerar token: %v", err)
	}
	core := NewHandler(Dependencies{TokenService: tokenService})
	handler := BuildHTTPHandler(core)

	req := httptest.NewRequest(http.MethodPost, "/invoices", strings.NewReader(`{"user_id":"user-1","amount_brl":100}`))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "IDEMPOTENCY_REQUIRED" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestWebhookSignatureMissing(t *testing.T) {
	handler := webhookTestHandler("secret", fakeWebhookRepo{replay: false}, nil, 0)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/transactions/confirm", strings.NewReader(`{}`))
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("X-Nonce", "nonce-1")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "SIGNATURE_MISSING" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestWebhookTimestampInvalid(t *testing.T) {
	handler := webhookTestHandler("secret", fakeWebhookRepo{replay: false}, nil, 0)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/transactions/confirm", strings.NewReader(`{}`))
	req.Header.Set("X-Signature", "deadbeef")
	req.Header.Set("X-Timestamp", "not-a-time")
	req.Header.Set("X-Nonce", "nonce-2")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "TIMESTAMP_INVALID" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestWebhookTimestampExpired(t *testing.T) {
	handler := webhookTestHandler("secret", fakeWebhookRepo{replay: false}, nil, 0)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/transactions/confirm", strings.NewReader(`{}`))
	req.Header.Set("X-Signature", "deadbeef")
	req.Header.Set("X-Timestamp", time.Now().Add(-10*time.Minute).UTC().Format(time.RFC3339))
	req.Header.Set("X-Nonce", "nonce-3")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "TIMESTAMP_EXPIRED" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestWebhookNonceMissing(t *testing.T) {
	handler := webhookTestHandler("secret", fakeWebhookRepo{replay: false}, nil, 0)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/transactions/confirm", strings.NewReader(`{}`))
	req.Header.Set("X-Signature", "deadbeef")
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "NONCE_MISSING" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestWebhookNonceReplay(t *testing.T) {
	secret := "secret"
	body := `{"event":"ok"}`
	signature := hmacSHA256Hex([]byte(secret), []byte(body))
	handler := webhookTestHandler(secret, fakeWebhookRepo{replay: true}, nil, 0)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/transactions/confirm", strings.NewReader(body))
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("X-Nonce", "nonce-4")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "NONCE_REPLAY" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestWebhookSignatureInvalid(t *testing.T) {
	handler := webhookTestHandler("secret", fakeWebhookRepo{replay: false}, nil, 0)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/transactions/confirm", strings.NewReader(`{}`))
	req.Header.Set("X-Signature", "invalid")
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("X-Nonce", "nonce-5")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "SIGNATURE_INVALID" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestWebhookIPDenied(t *testing.T) {
	handler := webhookTestHandler("secret", fakeWebhookRepo{replay: false}, []string{"1.1.1.1"}, 0)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/transactions/confirm", strings.NewReader(`{}`))
	req.Header.Set("X-Forwarded-For", "2.2.2.2")
	req.Header.Set("X-Signature", "deadbeef")
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("X-Nonce", "nonce-6")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status inesperado: %d", rr.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.Error == nil || response.Error.Code != "WEBHOOK_IP_DENIED" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}

func TestWebhookRateLimit(t *testing.T) {
	secret := "secret"
	body := `{"event":"ok"}`
	signature := hmacSHA256Hex([]byte(secret), []byte(body))
	handler := webhookTestHandler(secret, fakeWebhookRepo{replay: false}, nil, 1)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/transactions/confirm", strings.NewReader(body))
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("X-Nonce", "nonce-7")
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	req2 := httptest.NewRequest(http.MethodPost, "/webhooks/transactions/confirm", strings.NewReader(body))
	req2.Header.Set("X-Signature", signature)
	req2.Header.Set("X-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req2.Header.Set("X-Nonce", "nonce-8")
	req2.Header.Set("X-Forwarded-For", "10.0.0.1")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("status inesperado: %d", rr2.Code)
	}
	var response apiResponse
	testutil.DecodeJSONResponse(t, rr2, &response)
	if response.Error == nil || response.Error.Code != "WEBHOOK_RATE_LIMIT" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}
