package httpadapter

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type webhookRateLimiter struct {
	mu     sync.Mutex
	window time.Time
	counts map[string]int
}

func (l *webhookRateLimiter) allow(ip string, limit int) bool {
	if limit <= 0 {
		return true
	}
	now := time.Now().UTC().Truncate(time.Minute)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.window.IsZero() || !l.window.Equal(now) {
		l.window = now
		l.counts = map[string]int{}
	}
	count := l.counts[ip] + 1
	l.counts[ip] = count
	return count <= limit
}

func withWebhookSecurity(handler *Handler, next http.Handler) http.Handler {
	if handler == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(handler.webhookAllowedIPs) > 0 {
			clientIP := clientIPFromRequest(r)
			if !isIPAllowed(clientIP, handler.webhookAllowedIPs) {
				writeError(r.Context(), w, http.StatusForbidden, "ip nao permitido", "WEBHOOK_IP_DENIED")
				return
			}
		}
		if handler.webhookRateLimiter != nil {
			clientIP := clientIPFromRequest(r)
			if !handler.webhookRateLimiter.allow(clientIP, handler.webhookRateLimitPerMinute) {
				writeError(r.Context(), w, http.StatusTooManyRequests, "rate limit excedido", "WEBHOOK_RATE_LIMIT")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func isIPAllowed(ip string, allowlist []string) bool {
	for _, allowed := range allowlist {
		if ip == allowed {
			return true
		}
		if strings.HasPrefix(allowed, "ip:") && strings.TrimPrefix(allowed, "ip:") == ip {
			return true
		}
	}
	return false
}
