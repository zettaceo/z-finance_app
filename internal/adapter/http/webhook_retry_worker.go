package httpadapter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

const (
	webhookRetryMaxAttempts = 5
)

func StartWebhookRetryWorker(ctx context.Context, handler *Handler, repo repository.WebhookRetryRepository, interval time.Duration) {
	if handler == nil || repo == nil {
		return
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			webhookRetryLastRun.Set(time.Now().UTC().Format(time.RFC3339))
			processWebhookRetries(ctx, handler, repo)
		}
	}
}

func processWebhookRetries(ctx context.Context, handler *Handler, repo repository.WebhookRetryRepository) {
	now := time.Now().UTC()
	jobs, err := repo.ListDue(ctx, now, 100)
	if err != nil {
		webhookRetryFailed.Add(1)
		_ = handler.appendAudit(ctx, "", "WEBHOOK_RETRY_LIST_FAILED", "webhook_retry", "", map[string]any{
			"error": err.Error(),
		})
		return
	}
	for _, job := range jobs {
		if job == nil {
			continue
		}
		shouldStop := processWebhookRetryJob(ctx, handler, repo, job)
		if shouldStop {
			break
		}
	}
}

func processWebhookRetryJob(ctx context.Context, handler *Handler, repo repository.WebhookRetryRepository, job *entity.WebhookRetryJob) bool {
	if job.Attempts >= webhookRetryMaxAttempts {
		_ = repo.MarkFailure(ctx, job.ID, job.Attempts, time.Now().UTC(), "max attempts reached", entity.WebhookRetryDead)
		return false
	}

	status, err := handler.executeWebhookRetry(ctx, job)
	if err == nil && status >= 200 && status < 300 {
		_ = repo.MarkSuccess(ctx, job.ID)
		webhookRetrySucceeded.Add(1)
		_ = handler.appendAudit(ctx, "", "WEBHOOK_RETRY_SUCCESS", "webhook_retry", job.ID, map[string]any{
			"event_type": job.EventType,
			"status":     status,
		})
		return false
	}

	nextAttempts := job.Attempts + 1
	nextRetryAt := time.Now().UTC().Add(backoffDuration(nextAttempts))
	statusValue := entity.WebhookRetryPending
	if nextAttempts >= webhookRetryMaxAttempts {
		statusValue = entity.WebhookRetryDead
		webhookRetryDead.Add(1)
	}
	lastError := fmt.Sprintf("status=%d", status)
	if err != nil {
		lastError = err.Error()
	}
	_ = repo.MarkFailure(ctx, job.ID, nextAttempts, nextRetryAt, lastError, statusValue)
	webhookRetryFailed.Add(1)
	_ = handler.appendAudit(ctx, "", "WEBHOOK_RETRY_FAILED", "webhook_retry", job.ID, map[string]any{
		"event_type": job.EventType,
		"attempts":   nextAttempts,
		"status":     status,
		"error":      lastError,
	})
	return false
}

func (h *Handler) executeWebhookRetry(ctx context.Context, job *entity.WebhookRetryJob) (int, error) {
	if h == nil || job == nil {
		return http.StatusInternalServerError, fmt.Errorf("handler ou job nulo")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, job.Path, bytes.NewReader(job.Payload))
	if err != nil {
		return http.StatusInternalServerError, err
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	switch job.EventType {
	case "payment_confirm":
		h.handlePaymentConfirmWebhook(rec, req)
	case "payment_reject":
		h.handlePaymentRejectWebhook(rec, req)
	case "card_confirm":
		h.handleCardConfirmWebhook(rec, req)
	case "card_reject":
		h.handleCardRejectWebhook(rec, req)
	case "transaction_confirm":
		h.handleTransactionConfirmWebhook(rec, req)
	case "transaction_reject":
		h.handleTransactionRejectWebhook(rec, req)
	case "pix_receive":
		h.handlePixWebhook(rec, req)
	default:
		return http.StatusBadRequest, fmt.Errorf("event_type nao suportado")
	}
	return rec.Code, nil
}

func backoffDuration(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := time.Duration(1<<uint(attempt-1)) * time.Minute
	if delay > time.Hour {
		return time.Hour
	}
	return delay
}
