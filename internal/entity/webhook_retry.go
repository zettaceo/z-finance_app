package entity

import "time"

type WebhookRetryStatus string

const (
	WebhookRetryPending    WebhookRetryStatus = "PENDING"
	WebhookRetryProcessing WebhookRetryStatus = "PROCESSING"
	WebhookRetrySucceeded  WebhookRetryStatus = "SUCCEEDED"
	WebhookRetryDead       WebhookRetryStatus = "DEAD"
)

type WebhookRetryJob struct {
	ID          string
	EventType   string
	Path        string
	Payload     []byte
	Headers     map[string]string
	Attempts    int
	Status      WebhookRetryStatus
	NextRetryAt time.Time
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
