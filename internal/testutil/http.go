package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func DecodeJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, target any) {
	t.Helper()
	decoder := json.NewDecoder(rr.Body)
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("falha ao decodificar resposta: %v", err)
	}
}
