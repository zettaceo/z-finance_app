package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
)

func TestRepositoriesWithDatabase(t *testing.T) {
	databaseURL := os.Getenv("PG_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PG_TEST_DATABASE_URL nao configurada")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("falha ao abrir pool: %v", err)
	}
	defer pool.Close()

	if err := runMigrations(ctx, pool); err != nil {
		t.Fatalf("falha ao aplicar migracoes: %v", err)
	}

	limitRepo := NewKycLimitRepository(pool)
	limit := &entity.KycLimit{
		Level:        entity.KYCLevelBasic,
		DailyLimit:   123,
		MonthlyLimit: 456,
	}
	if err := limitRepo.Upsert(ctx, limit); err != nil {
		t.Fatalf("falha ao salvar limite: %v", err)
	}
	loaded, err := limitRepo.GetByLevel(ctx, entity.KYCLevelBasic)
	if err != nil {
		t.Fatalf("falha ao obter limite: %v", err)
	}
	if loaded.DailyLimit != 123 || loaded.MonthlyLimit != 456 {
		t.Fatalf("limites inesperados: %+v", loaded)
	}

	webhookRepo := NewWebhookRepository(pool)
	processed, err := webhookRepo.EnsureEvent(ctx, "payment_confirm", "ref-1")
	if err != nil {
		t.Fatalf("falha ao registrar evento: %v", err)
	}
	if processed {
		t.Fatalf("evento deveria ser novo")
	}
	processed, err = webhookRepo.EnsureEvent(ctx, "payment_confirm", "ref-1")
	if err != nil {
		t.Fatalf("falha ao registrar evento duplicado: %v", err)
	}
	if !processed {
		t.Fatalf("evento deveria ser idempotente")
	}
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	data, err := os.ReadFile("db/migrations/init.sql")
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, string(data))
	return err
}
