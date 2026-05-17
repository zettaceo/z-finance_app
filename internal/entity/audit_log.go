package entity

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         string
	UserID     string
	Action     string
	EntityType string
	EntityID   string
	Data       json.RawMessage
	CreatedAt  time.Time
}
