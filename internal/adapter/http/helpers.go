package httpadapter

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/bits"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type idempotencyContextKey string

const idempotencyKeyContextKey idempotencyContextKey = "idempotency_key"

func requireIdempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			writeError(r.Context(), w, http.StatusBadRequest, "Idempotency-Key obrigatoria", "IDEMPOTENCY_REQUIRED")
			return
		}
		ctx := context.WithValue(r.Context(), idempotencyKeyContextKey, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateWebhookSignature(secret string, repo repository.WebhookRepository, audit func(ctx context.Context, action string, data any), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			if audit != nil {
				audit(r.Context(), "WEBHOOK_SECRET_MISSING", map[string]any{
					"path": r.URL.Path,
					"ip":   clientIPFromRequest(r),
				})
			}
			writeError(r.Context(), w, http.StatusInternalServerError, "webhook secret nao configurado", "WEBHOOK_SECRET_MISSING")
			return
		}
		signature := r.Header.Get("X-Signature")
		if signature == "" {
			if audit != nil {
				audit(r.Context(), "WEBHOOK_SIGNATURE_MISSING", map[string]any{
					"path": r.URL.Path,
					"ip":   clientIPFromRequest(r),
				})
			}
			writeError(r.Context(), w, http.StatusUnauthorized, "assinatura ausente", "SIGNATURE_MISSING")
			return
		}
		timestamp, err := parseWebhookTimestamp(r.Header.Get("X-Timestamp"))
		if err != nil {
			if audit != nil {
				audit(r.Context(), "WEBHOOK_TIMESTAMP_INVALID", map[string]any{
					"path": r.URL.Path,
					"ip":   clientIPFromRequest(r),
				})
			}
			writeError(r.Context(), w, http.StatusUnauthorized, "timestamp invalido", "TIMESTAMP_INVALID")
			return
		}
		if !isWebhookTimestampFresh(timestamp, time.Now().UTC()) {
			if audit != nil {
				audit(r.Context(), "WEBHOOK_TIMESTAMP_EXPIRED", map[string]any{
					"path":      r.URL.Path,
					"ip":        clientIPFromRequest(r),
					"timestamp": timestamp.Format(time.RFC3339Nano),
				})
			}
			writeError(r.Context(), w, http.StatusUnauthorized, "timestamp expirado", "TIMESTAMP_EXPIRED")
			return
		}
		nonce := strings.TrimSpace(r.Header.Get("X-Nonce"))
		if nonce == "" {
			if audit != nil {
				audit(r.Context(), "WEBHOOK_NONCE_MISSING", map[string]any{
					"path": r.URL.Path,
					"ip":   clientIPFromRequest(r),
				})
			}
			writeError(r.Context(), w, http.StatusUnauthorized, "nonce ausente", "NONCE_MISSING")
			return
		}
		if repo != nil {
			alreadyProcessed, err := repo.EnsureEvent(r.Context(), "webhook_nonce", nonce)
			if err != nil {
				if audit != nil {
					audit(r.Context(), "WEBHOOK_NONCE_CHECK_FAILED", map[string]any{
						"path":  r.URL.Path,
						"ip":    clientIPFromRequest(r),
						"nonce": nonce,
					})
				}
				writeError(r.Context(), w, http.StatusInternalServerError, "falha ao validar nonce", "NONCE_CHECK_FAILED")
				return
			}
			if alreadyProcessed {
				if audit != nil {
					audit(r.Context(), "WEBHOOK_NONCE_REPLAY", map[string]any{
						"path":  r.URL.Path,
						"ip":    clientIPFromRequest(r),
						"nonce": nonce,
					})
				}
				writeError(r.Context(), w, http.StatusConflict, "nonce reutilizado", "NONCE_REPLAY")
				return
			}
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			if audit != nil {
				audit(r.Context(), "WEBHOOK_PAYLOAD_INVALID", map[string]any{
					"path": r.URL.Path,
					"ip":   clientIPFromRequest(r),
				})
			}
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		expected := hmacSHA256Hex([]byte(secret), body)
		if !hmac.Equal([]byte(expected), []byte(signature)) {
			if audit != nil {
				audit(r.Context(), "WEBHOOK_SIGNATURE_INVALID", map[string]any{
					"path": r.URL.Path,
					"ip":   clientIPFromRequest(r),
				})
			}
			writeError(r.Context(), w, http.StatusUnauthorized, "assinatura invalida", "SIGNATURE_INVALID")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func hmacSHA256Hex(secret, data []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

func parseWebhookTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("timestamp ausente")
	}
	parsedUnix, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return time.Unix(parsedUnix, 0).UTC(), nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func isWebhookTimestampFresh(value time.Time, now time.Time) bool {
	const maxSkew = 5 * time.Minute
	diff := now.Sub(value)
	if diff < 0 {
		diff = -diff
	}
	return diff <= maxSkew
}

func idempotencyKeyFromContext(ctx context.Context) string {
	key, _ := ctx.Value(idempotencyKeyContextKey).(string)
	return key
}

func withWebhookRetry(eventType string, path string, repo repository.WebhookRetryRepository, next http.Handler) http.Handler {
	if repo == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "payload invalido", "INVALID_PAYLOAD")
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		if recorder.status < http.StatusInternalServerError {
			return
		}
		if err := repo.Enqueue(r.Context(), &entity.WebhookRetryJob{
			ID:          uuid.NewString(),
			EventType:   eventType,
			Path:        path,
			Payload:     body,
			Headers:     map[string]string{},
			Attempts:    0,
			Status:      entity.WebhookRetryPending,
			NextRetryAt: time.Now().UTC().Add(1 * time.Minute),
			LastError:   fmt.Sprintf("status=%d", recorder.status),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}); err == nil {
			webhookRetryEnqueued.Add(1)
		}
	})
}

func transactionResponseFromEntity(tx *entity.Transaction) transactionResponse {
	return transactionResponse{
		ID:             tx.ID,
		AccountID:      tx.AccountID,
		UserID:         tx.UserID,
		Type:           string(tx.Type),
		Status:         string(tx.Status),
		Amount:         tx.Amount,
		Fee:            tx.Fee,
		NetAmount:      tx.NetAmount,
		IdempotencyKey: tx.IdempotencyKey,
		ExternalRef:    tx.ExternalRef,
		OccurredAt:     tx.OccurredAt.UTC().Format(time.RFC3339Nano),
		CreatedAt:      tx.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func adminUserSummaryFromEntity(user *entity.User) adminUserSummary {
	return adminUserSummary{
		ID:         user.ID,
		ExternalID: user.ExternalID,
		Email:      user.Email,
		FullName:   user.FullName,
		Status:     string(user.Status),
		UserType:   string(user.UserType),
		CreatedAt:  user.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:  user.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func adminAccountSummaryFromEntity(account *entity.Account, balance int64) adminAccountSummary {
	return adminAccountSummary{
		ID:        account.ID,
		Currency:  account.Currency,
		Scale:     account.Scale,
		Status:    string(account.Status),
		Balance:   balance,
		CreatedAt: account.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func nullableString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func pixResponseFromEntity(transfer *entity.PixTransfer) pixResponse {
	confirmedAt := ""
	if transfer.ConfirmedAt != nil {
		confirmedAt = transfer.ConfirmedAt.UTC().Format(time.RFC3339Nano)
	}
	return pixResponse{
		ID:            transfer.ID,
		TransactionID: transfer.TransactionID,
		UserID:        transfer.UserID,
		AccountID:     transfer.AccountID,
		Direction:     string(transfer.Direction),
		Status:        string(transfer.Status),
		Amount:        transfer.Amount,
		Fee:           transfer.Fee,
		NetAmount:     transfer.NetAmount,
		EndToEndID:    transfer.EndToEndID,
		ExternalRef:   transfer.ExternalRef,
		ConfirmedAt:   confirmedAt,
		OccurredAt:    transfer.OccurredAt.UTC().Format(time.RFC3339Nano),
		CreatedAt:     transfer.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:     transfer.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func paymentResponseFromEntity(payment *entity.Payment) paymentResponse {
	return paymentResponse{
		ID:        payment.ID,
		UserID:    payment.UserID,
		AccountID: payment.AccountID,
		Status:    string(payment.Status),
		Amount:    payment.Amount,
		Fee:       payment.Fee,
		NetAmount: payment.NetAmount,
		Barcode:   payment.Barcode,
	}
}

func ptrPreRegistrationResponse(value preRegistrationResponse) *preRegistrationResponse {
	return &value
}

func preRegistrationResponseFromEntity(item *entity.PreRegistration) preRegistrationResponse {
	emailVerified := item.EmailStatus == entity.VerificationVerified
	phoneVerified := item.PhoneStatus == entity.VerificationVerified
	return preRegistrationResponse{
		ID:            item.ID,
		Status:        string(item.Status),
		EmailStatus:   string(item.EmailStatus),
		PhoneStatus:   string(item.PhoneStatus),
		EmailVerified: emailVerified,
		PhoneVerified: phoneVerified,
		ExpiresAt:     item.ExpiresAt.UTC().Format(time.RFC3339Nano),
		CreatedAt:     item.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:     item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func swapResponseFromEntity(order *entity.TradeOrder) swapResponse {
	return swapResponse{
		ID:     order.ID,
		Status: string(order.Status),
		Side:   string(order.Side),
	}
}

func invoiceResponseFromEntity(invoice *entity.Invoice, qrPayload string) invoiceResponse {
	return invoiceResponse{
		ID:           invoice.ID,
		UserID:       invoice.UserID,
		AmountBRL:    invoice.AmountBRL,
		PixCopyPaste: invoice.PixCopyPaste,
		USDTAddress:  invoice.USDTAddress,
		CreatedAt:    invoice.CreatedAt.UTC().Format(time.RFC3339Nano),
		QRPayload:    qrPayload,
	}
}

func cryptoTransferResponseFromEntity(transfer *entity.CryptoTransfer) cryptoPayResponse {
	return cryptoPayResponse{
		ID:            transfer.ID,
		UserID:        transfer.UserID,
		AccountID:     transfer.AccountID,
		Asset:         transfer.Asset,
		Network:       transfer.Network,
		Address:       transfer.Address,
		Amount:        transfer.Amount,
		Fee:           transfer.Fee,
		Status:        string(transfer.Status),
		TransactionID: transfer.TransactionID,
		CreatedAt:     transfer.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func normalizeBarcode(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func isValidBarcode(value string) bool {
	length := len(value)
	if length != 44 && length != 47 {
		return false
	}
	for i := 0; i < length; i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

func safeMulInt64(a, b int64) (int64, bool) {
	if a < 0 || b < 0 {
		return 0, false
	}
	hi, lo := bits.Mul64(uint64(a), uint64(b))
	if hi != 0 {
		return 0, false
	}
	return int64(lo), true
}

func debitTradeType(side entity.TradeSide) entity.TransactionType {
	if side == entity.TradeSideSell {
		return entity.TransactionTypeWithdrawal
	}
	return entity.TransactionTypeTradeBuy
}

func buildNextCursor(items []*entity.Transaction) string {
	if len(items) == 0 {
		return ""
	}
	last := items[len(items)-1]
	return fmt.Sprintf("%s|%s", last.OccurredAt.UTC().Format(time.RFC3339Nano), last.ID)
}

func parseCursor(value string) (*time.Time, string, error) {
	if value == "" {
		return nil, "", nil
	}
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, "", fmt.Errorf("cursor invalido")
	}
	parsed, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, "", err
	}
	return &parsed, parts[1], nil
}

func parseTimeParam(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return &parsed, nil
	}
	parsed, err = time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseLimit(value string) (int, error) {
	if value == "" {
		return 50, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("limit invalido")
	}
	if parsed > 200 {
		return 200, nil
	}
	return parsed, nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func clientIPFromRequest(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func defaultIfEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
