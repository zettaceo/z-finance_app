package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"expvar"
	"fmt"
	"log"
	"strings"
	"time"
)

type contextKey string

const (
	traceIDKey contextKey = "trace_id"
	spanIDKey  contextKey = "span_id"
)

var (
	httpDurationTotal      = expvar.NewInt("http_duration_ms_total")
	httpDurationCount      = expvar.NewInt("http_duration_ms_count")
	httpDurationByPath     = expvar.NewMap("http_duration_ms_by_path")
	httpDurationCountPath  = expvar.NewMap("http_duration_ms_count_by_path")
	usecaseDurationTotal   = expvar.NewMap("usecase_duration_ms_total")
	usecaseDurationCount   = expvar.NewMap("usecase_duration_ms_count")
	spanDurationTotal      = expvar.NewMap("span_duration_ms_total")
	spanDurationCount      = expvar.NewMap("span_duration_ms_count")
)

func ContextWithTrace(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		traceID = newID()
	}
	return context.WithValue(ctx, traceIDKey, traceID)
}

func ContextWithSpan(ctx context.Context, spanID string) context.Context {
	if spanID == "" {
		spanID = newID()
	}
	return context.WithValue(ctx, spanIDKey, spanID)
}

func TraceIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(traceIDKey).(string)
	return value
}

func SpanIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(spanIDKey).(string)
	return value
}

func RecordHTTP(path string, status int, duration time.Duration) {
	ms := duration.Milliseconds()
	httpDurationTotal.Add(ms)
	httpDurationCount.Add(1)
	normalized := normalizePath(path)
	if normalized != "" {
		httpDurationByPath.Add(normalized, ms)
		httpDurationCountPath.Add(normalized, 1)
	}
	_ = status
}

func StartUseCaseSpan(ctx context.Context, name string) (context.Context, func()) {
	return startSpan(ctx, "usecase", name)
}

func StartSpan(ctx context.Context, name string) (context.Context, func()) {
	return startSpan(ctx, "span", name)
}

func startSpan(ctx context.Context, kind, name string) (context.Context, func()) {
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		traceID = newID()
		ctx = ContextWithTrace(ctx, traceID)
	}
	parentSpan := SpanIDFromContext(ctx)
	spanID := newID()
	ctx = ContextWithSpan(ctx, spanID)
	start := time.Now()

	return ctx, func() {
		duration := time.Since(start)
		ms := duration.Milliseconds()
		spanKey := fmt.Sprintf("%s.%s", kind, name)
		spanDurationTotal.Add(spanKey, ms)
		spanDurationCount.Add(spanKey, 1)
		if kind == "usecase" {
			usecaseDurationTotal.Add(name, ms)
			usecaseDurationCount.Add(name, 1)
		}
		log.Printf("trace_id=%s span_id=%s parent_span=%s span=%s duration_ms=%d", traceID, spanID, parentSpan, spanKey, ms)
	}
}

func normalizePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	return trimmed
}

func newID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fmt.Sprintf("trace-%d", time.Now().UnixNano())
}
