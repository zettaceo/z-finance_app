package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/repository"
)

type OutboxRepository struct {
	db dbExecutor
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{db: pool}
}

func NewOutboxRepositoryWithTx(tx dbExecutor) *OutboxRepository {
	return &OutboxRepository{db: tx}
}

func (r *OutboxRepository) Enqueue(ctx context.Context, eventType string, payload map[string]any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO outbox_events (event_type, payload)
		VALUES ($1, $2)`, eventType, data)
	return err
}

var _ repository.OutboxRepository = (*OutboxRepository)(nil)
