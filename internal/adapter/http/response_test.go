package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"z-finance-api/internal/testutil"
)

func TestWriteJSONResponseEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), requestIDContextKey, "req-123")

	writeJSON(ctx, rr, http.StatusOK, map[string]string{"status": "ok"})

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type inesperado: %s", got)
	}

	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.RequestID != "req-123" {
		t.Fatalf("request_id inesperado: %s", response.RequestID)
	}
	if response.Error != nil {
		t.Fatalf("nao era esperado erro: %+v", response.Error)
	}

	data, ok := response.Data.(map[string]any)
	if !ok {
		t.Fatalf("data inesperado: %T", response.Data)
	}
	if data["status"] != "ok" {
		t.Fatalf("status inesperado: %v", data["status"])
	}
}

func TestWriteErrorResponseEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), requestIDContextKey, "req-456")

	writeError(ctx, rr, http.StatusBadRequest, "erro de teste", "TEST_ERROR")

	var response apiResponse
	testutil.DecodeJSONResponse(t, rr, &response)
	if response.RequestID != "req-456" {
		t.Fatalf("request_id inesperado: %s", response.RequestID)
	}
	if response.Error == nil || response.Error.Code != "TEST_ERROR" {
		t.Fatalf("erro inesperado: %+v", response.Error)
	}
}
