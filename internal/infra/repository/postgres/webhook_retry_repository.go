package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type WebhookRetryRepository struct {
	db *pgxpool.Pool
}

func NewWebhookRetryRepository(pool *pgxpool.Pool) *WebhookRetryRepository {
	return &WebhookRetryRepository{db: pool}
}

func (r *WebhookRetryRepository) Enqueue(ctx context.Context, job *entity.WebhookRetryJob) error {
	if job == nil {
		return repository.ErrInvalidState
	}
	headers, err := json.Marshal(job.Headers)
	if err != nil {
		return err
	}
	payload := json.RawMessage(job.Payload)
	_, err = r.db.Exec(ctx, `
		INSERT INTO webhook_retry_jobs (
			id, event_type, path, payload, headers, attempts, status, next_retry_at, last_error, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		)`,
		job.ID,
		job.EventType,
		job.Path,
		payload,
		headers,
		job.Attempts,
		job.Status,
		job.NextRetryAt,
		nullIfEmpty(job.LastError),
	)
	return err
}

func (r *WebhookRetryRepository) ListDue(ctx context.Context, now time.Time, limit int) ([]*entity.WebhookRetryJob, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, event_type, path, payload, headers, attempts, status, next_retry_at, last_error, created_at, updated_at
		  FROM webhook_retry_jobs
		 WHERE status = $1
		   AND next_retry_at <= $2
		 ORDER BY next_retry_at ASC
		 LIMIT $3`,
		entity.WebhookRetryPending,
		now,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.WebhookRetryJob
	for rows.Next() {
		var job entity.WebhookRetryJob
		var payload json.RawMessage
		var headers []byte
		if err := rows.Scan(
			&job.ID,
			&job.EventType,
			&job.Path,
			&payload,
			&headers,
			&job.Attempts,
			&job.Status,
			&job.NextRetryAt,
			&job.LastError,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, err
		}
		job.Payload = payload
		if len(headers) > 0 {
			_ = json.Unmarshal(headers, &job.Headers)
		}
		items = append(items, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *WebhookRetryRepository) ListByStatus(ctx context.Context, status entity.WebhookRetryStatus, limit int) ([]*entity.WebhookRetryJob, error) {
	if status == "" {
		status = entity.WebhookRetryPending
	}
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, event_type, path, payload, headers, attempts, status, next_retry_at, last_error, created_at, updated_at
		  FROM webhook_retry_jobs
		 WHERE status = $1
		 ORDER BY updated_at DESC
		 LIMIT $2`,
		status,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.WebhookRetryJob
	for rows.Next() {
		var job entity.WebhookRetryJob
		var payload json.RawMessage
		var headers []byte
		if err := rows.Scan(
			&job.ID,
			&job.EventType,
			&job.Path,
			&payload,
			&headers,
			&job.Attempts,
			&job.Status,
			&job.NextRetryAt,
			&job.LastError,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, err
		}
		job.Payload = payload
		if len(headers) > 0 {
			_ = json.Unmarshal(headers, &job.Headers)
		}
		items = append(items, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *WebhookRetryRepository) CountByStatus(ctx context.Context, status entity.WebhookRetryStatus) (int, error) {
	if status == "" {
		status = entity.WebhookRetryPending
	}
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(1)
		  FROM webhook_retry_jobs
		 WHERE status = $1`,
		status,
	).Scan(&count)
	return count, err
}

func (r *WebhookRetryRepository) MarkSuccess(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE webhook_retry_jobs
		   SET status = $1,
		       updated_at = NOW()
		 WHERE id = $2`,
		entity.WebhookRetrySucceeded,
		id,
	)
	return err
}

func (r *WebhookRetryRepository) MarkFailure(ctx context.Context, id string, attempts int, nextRetryAt time.Time, lastError string, status entity.WebhookRetryStatus) error {
	_, err := r.db.Exec(ctx, `
		UPDATE webhook_retry_jobs
		   SET attempts = $1,
		       status = $2,
		       next_retry_at = $3,
		       last_error = $4,
		       updated_at = NOW()
		 WHERE id = $5`,
		attempts,
		status,
		nextRetryAt,
		nullIfEmpty(lastError),
		id,
	)
	return err
}

var _ repository.WebhookRetryRepository = (*WebhookRetryRepository)(nil)
