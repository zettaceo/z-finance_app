package httpadapter

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"z-finance-api/internal/infra/observability"
)

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withStandardHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if traceID := observability.TraceIDFromContext(r.Context()); traceID != "" {
			w.Header().Set("X-Trace-Id", traceID)
		}
		if requestID := requestIDFromContext(r.Context()); requestID != "" {
			w.Header().Set("X-Request-Id", requestID)
		}
		next.ServeHTTP(w, r)
	})
}

var (
	requestMetrics = expvar.NewMap("http_requests")
	requestTotal   = expvar.NewInt("http_requests_total")
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		duration := time.Since(start)
		requestID := requestIDFromContext(r.Context())
		traceID := observability.TraceIDFromContext(r.Context())
		requestTotal.Add(1)
		requestMetrics.Add(fmt.Sprintf("status_%d", recorder.status), 1)
		observability.RecordHTTP(r.URL.Path, recorder.status, duration)
		log.Printf("request_id=%s trace_id=%s method=%s path=%s status=%d duration_ms=%d", requestID, traceID, r.Method, r.URL.Path, recorder.status, duration.Milliseconds())
	})
}

func withTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			traceID = requestIDFromContext(r.Context())
		}
		ctx := observability.ContextWithTrace(r.Context(), traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigins := allowedCORSOrigins()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,Idempotency-Key,X-Request-Id,X-Trace-Id")
			w.Header().Set("Access-Control-Max-Age", "600")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func allowedCORSOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ORIGINS"))
	if raw == "" {
		return []string{
			"http://95.111.247.134:5174",
			"http://localhost:5174",
			"http://localhost:5173",
		}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, value := range parts {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func isOriginAllowed(origin string, allowed []string) bool {
	for _, item := range allowed {
		if origin == item {
			return true
		}
	}
	return false
}

type authContextKey string

const userIDContextKey authContextKey = "user_id"

func userIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(userIDContextKey).(string)
	return value
}
