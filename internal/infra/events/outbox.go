package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxProcessor struct {
	pool     *pgxpool.Pool
	bus      *Bus
	interval time.Duration
}

func NewOutboxProcessor(pool *pgxpool.Pool, bus *Bus, interval time.Duration) *OutboxProcessor {
	return &OutboxProcessor{
		pool:     pool,
		bus:      bus,
		interval: interval,
	}
}

func (p *OutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		p.processBatch(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *OutboxProcessor) processBatch(ctx context.Context) {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("outbox: falha ao iniciar tx: %v", err)
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, `
		SELECT id, event_type, payload
		  FROM outbox_events
		 WHERE status = 'PENDING'
		 ORDER BY created_at
		 LIMIT 50
		 FOR UPDATE SKIP LOCKED`)
	if err != nil {
		log.Printf("outbox: falha ao buscar eventos: %v", err)
		return
	}
	defer rows.Close()

	type outboxRow struct {
		id        string
		eventType string
		payload   []byte
	}
	var items []outboxRow
	for rows.Next() {
		var item outboxRow
		if err := rows.Scan(&item.id, &item.eventType, &item.payload); err != nil {
			log.Printf("outbox: falha ao ler evento: %v", err)
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		log.Printf("outbox: falha ao iterar eventos: %v", err)
		return
	}

	for _, item := range items {
		status := "PROCESSED"
		var data map[string]any
		if err := json.Unmarshal(item.payload, &data); err != nil {
			status = "FAILED"
			log.Printf("outbox: payload invalido %s: %v", item.id, err)
		} else if p.bus != nil {
			p.bus.Publish(item.eventType, data)
		}

		if _, err := tx.Exec(ctx, `
			UPDATE outbox_events
			   SET status = $1, processed_at = NOW()
			 WHERE id = $2`, status, item.id); err != nil {
			log.Printf("outbox: falha ao atualizar evento %s: %v", item.id, err)
			return
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("outbox: falha ao commit: %v", err)
	}
}
