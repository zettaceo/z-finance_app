package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/configs"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/infra/repository/postgres"
	"z-finance-api/internal/usecase"
)

type fixedClock struct {
	value time.Time
}

func (f fixedClock) Now() time.Time {
	return f.value
}

func main() {
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("falha ao carregar configuracoes: %v", err)
	}
	log.SetPrefix("Z-BANK Seeder ")
	log.SetFlags(log.LstdFlags | log.LUTC)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("falha ao criar pool: %v", err)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	accountRepo := postgres.NewAccountRepository(pool)
	txRepo := postgres.NewTransactionRepository(pool)
	pixRepo := postgres.NewPixRepository(pool)
	auditRepo := postgres.NewAuditLogRepository(pool)
	planRepo := postgres.NewPlanRepository(pool)
	userPlanRepo := postgres.NewUserPlanRepository(pool)
	uow := postgres.NewPostgresUnitOfWork(pool)

	userUC := usecase.NewCreateUserUseCase(userRepo)
	accountUC := usecase.NewCreateAccountUseCase(accountRepo)

	userID := "11111111-1111-1111-1111-111111111111"
	accountID := "22222222-2222-2222-2222-222222222222"
	email := "investidor@zetta.bank"
	password := "123456"
	fullName := "Investidor Zetta"

	user, err := userUC.Execute(ctx, usecase.CreateUserInput{
		ID:       userID,
		Email:    email,
		FullName: fullName,
		Password: password,
	})
	if err != nil {
		log.Fatalf("falha ao criar usuario: %v", err)
	}

	account, err := accountUC.Execute(ctx, usecase.CreateAccountInput{
		ID:       accountID,
		UserID:   user.ID,
		Currency: "BRL",
		Scale:    2,
	})
	if err != nil {
		log.Fatalf("falha ao criar conta: %v", err)
	}

	if plan, err := planRepo.GetByCode(ctx, "PRO"); err == nil && plan != nil {
		if active, err := userPlanRepo.GetActiveByUser(ctx, user.ID, time.Now().UTC()); err == nil && active == nil {
			_ = userPlanRepo.Create(ctx, &entity.UserPlan{
				UserID:    user.ID,
				PlanID:    plan.ID,
				ValidFrom: time.Now().UTC(),
				CreatedAt: time.Now().UTC(),
			})
		}
	}

	now := time.Now().UTC()
	salaryDate := now.AddDate(0, 0, -15)
	rentDate := now.AddDate(0, 0, -12)
	pixDate := now.AddDate(0, 0, -10)
	uberDate := now.AddDate(0, 0, -5)

	createTxWithDate := func(at time.Time) *usecase.CreateTransactionUseCase {
		return usecase.NewCreateTransactionUseCaseWithClock(uow, fixedClock{value: at})
	}

	if _, err := createTxWithDate(salaryDate).Execute(ctx, usecase.CreateTransactionInput{
		ID:             uuid.NewString(),
		AccountID:      account.ID,
		UserID:         user.ID,
		Type:           entity.TransactionTypeDeposit,
		Status:         entity.TransactionStatusConfirmed,
		Amount:         2_500_000,
		Fee:            0,
		NetAmount:      2_500_000,
		IdempotencyKey: "seed-salary-25000",
		ExternalRef:    "SALARIO",
		OccurredAt:     &salaryDate,
	}); err != nil {
		log.Fatalf("falha ao criar deposito salario: %v", err)
	}

	if _, err := createTxWithDate(rentDate).Execute(ctx, usecase.CreateTransactionInput{
		ID:             uuid.NewString(),
		AccountID:      account.ID,
		UserID:         user.ID,
		Type:           entity.TransactionTypePayment,
		Status:         entity.TransactionStatusConfirmed,
		Amount:         450_000,
		Fee:            0,
		NetAmount:      450_000,
		IdempotencyKey: "seed-rent-4500",
		ExternalRef:    "ALUGUEL",
		OccurredAt:     &rentDate,
	}); err != nil {
		log.Fatalf("falha ao criar pagamento aluguel: %v", err)
	}

	sendPixUC := usecase.NewSendPixUseCaseWithClock(uow, pixRepo, nil, auditRepo, nil, fixedClock{value: pixDate})
	transfer, err := sendPixUC.Execute(ctx, usecase.SendPixInput{
		AccountID:      account.ID,
		UserID:         user.ID,
		Amount:         35_000,
		Fee:            0,
		NetAmount:      35_000,
		IdempotencyKey: "seed-pix-350",
		ExternalRef:    "PIX_ENVIADO",
	})
	if err != nil {
		log.Fatalf("falha ao enviar pix: %v", err)
	}

	webhookUC := usecase.NewHandlePixWebhookUseCaseWithClock(uow, nil, fixedClock{value: pixDate.Add(2 * time.Minute)})
	if _, err := webhookUC.Execute(ctx, usecase.PixWebhookInput{
		TransferID: transfer.ID,
		Status:     "CONFIRMED",
	}); err != nil {
		log.Fatalf("falha ao confirmar pix: %v", err)
	}

	if _, err := createTxWithDate(uberDate).Execute(ctx, usecase.CreateTransactionInput{
		ID:             uuid.NewString(),
		AccountID:      account.ID,
		UserID:         user.ID,
		Type:           entity.TransactionTypePayment,
		Status:         entity.TransactionStatusConfirmed,
		Amount:         4_290,
		Fee:            0,
		NetAmount:      4_290,
		IdempotencyKey: "seed-uber-42-90",
		ExternalRef:    "UBER",
		OccurredAt:     &uberDate,
	}); err != nil {
		log.Fatalf("falha ao criar pagamento uber: %v", err)
	}

	evp := os.Getenv("DEMO_PIX_KEY")
	if evp == "" {
		evp = uuid.NewString()
	}
	registerPixKeyUC := usecase.NewRegisterPixKeyUseCaseWithClock(uow, fixedClock{value: now})
	if _, err := registerPixKeyUC.Execute(ctx, usecase.RegisterPixKeyInput{
		UserID:    user.ID,
		AccountID: account.ID,
		Type:      string(entity.PixKeyTypeEVP),
		Key:       evp,
	}); err != nil {
		if !errors.Is(err, usecase.ErrPixKeyDuplicate) {
			log.Fatalf("falha ao registrar chave pix: %v", err)
		}
	}

	ledgerBalance, err := txRepo.GetLedgerBalance(ctx, account.ID)
	if err != nil {
		log.Fatalf("falha ao calcular saldo: %v", err)
	}

	log.Printf("Seeder concluido. Usuario=%s Conta=%s Saldo=%d", user.ID, account.ID, ledgerBalance)
}
