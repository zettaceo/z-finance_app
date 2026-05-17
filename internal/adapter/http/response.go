package httpadapter

import (
	"context"
	"encoding/json"
	"net/http"
)

type apiError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type apiResponse struct {
	Data      any      `json:"data,omitempty"`
	Error     *apiError `json:"error,omitempty"`
	RequestID string   `json:"request_id,omitempty"`
}

type contextKey string

const requestIDContextKey contextKey = "request_id"

func writeJSON(ctx context.Context, w http.ResponseWriter, status int, payload any) {
	response := apiResponse{
		Data:      payload,
		RequestID: requestIDFromContext(ctx),
	}
	writeResponse(w, status, response)
}

func writeError(ctx context.Context, w http.ResponseWriter, status int, message, code string) {
	response := apiResponse{
		Error: &apiError{
			Code:    code,
			Message: message,
		},
		RequestID: requestIDFromContext(ctx),
	}
	writeResponse(w, status, response)
}

func writeErrorWithDetails(ctx context.Context, w http.ResponseWriter, status int, message, code string, details map[string]any) {
	response := apiResponse{
		Error: &apiError{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: requestIDFromContext(ctx),
	}
	writeResponse(w, status, response)
}

func writeResponse(w http.ResponseWriter, status int, payload apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func requestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDContextKey).(string)
	return value
}
