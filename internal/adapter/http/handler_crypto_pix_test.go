package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"z-finance-api/internal/testutil"
	"z-finance-api/internal/usecase"
)

func TestPixSendFromCryptoRequiresIdempotency(t *testing.T) {
	tokenService := usecase.NewTokenService("test-secret", time.Minute)
	accessToken, _, err := tokenService.GenerateAccessToken("user-1")
	if err != nil {
		t.Fatalf("falha ao gerar token: %v", err)
	}
	core := NewHandler(Dependencies{
		TokenService: tokenService,
	})
	handler := BuildHTTPHandler(core)

	req := httptest.NewRequest(http.MethodPost, "/pix/send/crypto", strings.NewReader(`{}`))
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
