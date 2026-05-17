package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/configs"
	httpadapter "z-finance-api/internal/adapter/http"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/infra/events"
	"z-finance-api/internal/infra/crypto"
	"z-finance-api/internal/infra/custody"
	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/infra/zswap"
	"z-finance-api/internal/infra/obeliskz"
	"z-finance-api/internal/infra/zpay"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/conversion/engine"
	"z-finance-api/internal/infra/pricing"
	"z-finance-api/internal/infra/repository/postgres"
	pricingengine "z-finance-api/internal/pricing/engine"
	"z-finance-api/internal/usecase"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("falha ao carregar configuracoes: %v", err)
	}

	log.SetPrefix(cfg.AppName + " ")
	log.SetFlags(log.LstdFlags | log.LUTC)

	shutdownTracing, err := observability.InitTracing(context.Background(), observability.TracingConfig{
		ServiceName: cfg.OtelServiceName,
		Endpoint:    cfg.OtelEndpoint,
		Insecure:    cfg.OtelInsecure,
	})
	if err != nil {
		log.Fatalf("falha ao iniciar tracing: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdownTracing(ctx)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("falha ao criar pool: %v", err)
	}
	defer pool.Close()

	if err := waitForDatabase(pool); err != nil {
		log.Fatalf("falha ao pingar o banco: %v", err)
	}

	if err := runMigrations(ctx, pool); err != nil {
		log.Fatalf("falha ao executar migracoes: %v", err)
	}

	if err := seedAdminIfMissing(ctx, pool); err != nil {
		log.Fatalf("falha ao executar seed: %v", err)
	}
	if err := ensureKycLimits(ctx, pool); err != nil {
		log.Fatalf("falha ao garantir limites kyc: %v", err)
	}

	log.Println("conexao com o banco estabelecida e migracoes aplicadas")

	txRepo := postgres.NewTransactionRepository(pool)
	uow := postgres.NewPostgresUnitOfWork(pool)
	createTxUC := usecase.NewCreateTransactionUseCase(uow)
	accountRepo := postgres.NewAccountRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	planRepo := postgres.NewPlanRepository(pool)
	userPlanRepo := postgres.NewUserPlanRepository(pool)
	pricingRuleRepo := postgres.NewPricingRuleRepository(pool)
	pricingVersionRepo := postgres.NewPricingVersionRepository(pool)
	planFeatureRepo := postgres.NewPlanFeatureRepository(pool)
	planLimitRepo := postgres.NewPlanLimitRepository(pool)
	pricingCampaignRepo := postgres.NewPricingCampaignRepository(pool)
	pricingCampaignRuleRepo := postgres.NewPricingCampaignRuleRepository(pool)
	roleRepo := postgres.NewRoleRepository(pool)
	userRoleRepo := postgres.NewUserRoleRepository(pool)
	roleSeparationRepo := postgres.NewRoleSeparationRepository(pool)
	regulatoryProfileRepo := postgres.NewRegulatoryProfileRepository(pool)
	pixRepo := postgres.NewPixRepository(pool)
	auditRepo := postgres.NewAuditLogRepository(pool)
	paymentRepo := postgres.NewPaymentRepository(pool)
	cardRepo := postgres.NewCardAuthorizationRepository(pool)
	tradeRepo := postgres.NewTradeOrderRepository(pool)
	settingsRepo := postgres.NewUserSettingsRepository(pool)
	conversionRepo := postgres.NewConversionRuleRepository(pool)
	userFeatureOverrideRepo := postgres.NewUserFeatureOverrideRepository(pool)
	userLimitOverrideRepo := postgres.NewUserLimitOverrideRepository(pool)
	kycRepo := postgres.NewKYCRepository(pool)
	kycLimitRepo := postgres.NewKycLimitRepository(pool)
	invoiceRepo := postgres.NewInvoiceRepository(pool)
	pricingCache := pricing.NewCache()
	eventBus := events.NewBus()
	velocityRepo := postgres.NewVelocityRepository(pool)
	velocityChecker := usecase.NewVelocityChecker(velocityRepo, velocityPolicyFromConfig(cfg))
	webhookRepo := postgres.NewWebhookRepository(pool)
	webhookRetryRepo := postgres.NewWebhookRetryRepository(pool)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(pool)
	outboxProcessor := events.NewOutboxProcessor(pool, eventBus, 2*time.Second)
	go outboxProcessor.Start(context.Background())

	// Wire ecosystem integrations — use real service when URL is configured, mock otherwise.
	var exchangeGateway ports.ExchangeGateway
	if cfg.ZSwapURL != "" {
		log.Printf("ExchangeGateway: z-swap @ %s", cfg.ZSwapURL)
		exchangeGateway = zswap.NewExchangeGateway(cfg.ZSwapURL, cfg.ZSwapAPIKey)
	} else {
		log.Println("ExchangeGateway: mock (defina Z_SWAP_URL para usar z-swap)")
		exchangeGateway = crypto.NewMockExchangeGateway()
	}

	var custodyGateway ports.CustodyGateway
	if cfg.ObeliskZURL != "" {
		log.Printf("CustodyGateway: obelisk-z @ %s", cfg.ObeliskZURL)
		custodyGateway = obeliskz.NewCustodyGateway(cfg.ObeliskZURL, cfg.ObeliskZAPIKey)
	} else {
		log.Println("CustodyGateway: mock (defina OBELISK_Z_URL para usar obelisk-z)")
		custodyGateway = custody.NewMockGateway(custody.DefaultMockConfig())
	}

	var pixPartnerClient ports.PixPartnerClient
	if cfg.ZPayURL != "" {
		log.Printf("PixPartnerClient: z-pay @ %s", cfg.ZPayURL)
		pixPartnerClient = zpay.NewPixPartnerClient(cfg.ZPayURL, cfg.ZPayAPIKey)
	} else {
		log.Println("PixPartnerClient: nil (PIX enviado sem parceiro SPI externo)")
		pixPartnerClient = nil
	}
	conversionEngine := engine.NewEngine(conversionRepo, settingsRepo)
	pricingEngine := pricingengine.New()
	resolvePricingUC := usecase.NewResolvePricingUseCase(
		userRepo,
		planRepo,
		userPlanRepo,
		pricingRuleRepo,
		pricingVersionRepo,
		planFeatureRepo,
		pricingCampaignRepo,
		pricingCampaignRuleRepo,
		pricingEngine,
	)
	autoConvertPixUC := usecase.NewAutoConvertPixUseCase(uow, exchangeGateway, conversionEngine, settingsRepo, resolvePricingUC)
	registerPixKeyUC := usecase.NewRegisterPixKeyUseCase(uow)
	sendPixUC := usecase.NewSendPixUseCase(uow, pixRepo, pixPartnerClient, auditRepo, resolvePricingUC)
	pixWebhookUC := usecase.NewHandlePixWebhookUseCase(uow, autoConvertPixUC)
	payCryptoWithFiatUC := usecase.NewPayCryptoWithFiatUseCase(uow, exchangeGateway, custodyGateway, resolvePricingUC)
	authorizeCardJITUC := usecase.NewAuthorizeCardJITUseCase(uow, exchangeGateway, settingsRepo, conversionEngine, resolvePricingUC)
	sendPixFromCryptoUC := usecase.NewSendPixFromCryptoUseCase(uow, exchangeGateway, pixRepo, resolvePricingUC)
	createInvoiceUC := usecase.NewCreateInvoiceUseCase(uow, custodyGateway)
	payInvoiceUC := usecase.NewPayInvoiceUseCase(invoiceRepo, sendPixUC, payCryptoWithFiatUC, resolvePricingUC)
	ensureFiatCoverageUC := usecase.NewEnsureFiatCoverageUseCase(uow, exchangeGateway, conversionEngine, resolvePricingUC)
	tokenService := usecase.NewTokenService(cfg.AuthSecret, time.Duration(cfg.AccessTokenTTLMinutes)*time.Minute)
	loginUC := usecase.NewLoginUseCase(uow, userRepo, tokenService, time.Duration(cfg.RefreshTokenTTLDays)*24*time.Hour, cfg.LoginMaxAttempts, time.Duration(cfg.LoginWindowSeconds)*time.Second)
	refreshUC := usecase.NewRefreshTokenUseCase(uow, tokenService, time.Duration(cfg.RefreshTokenTTLDays)*24*time.Hour)
	logoutUC := usecase.NewLogoutUseCase(uow, tokenService)
	preRegistrationPolicy := usecase.PreRegistrationPolicy{
		Expiry:           time.Duration(cfg.PreRegistrationExpiryHours) * time.Hour,
		MaxEmailAttempts: int(cfg.PreRegistrationMaxEmailAttempts),
		MaxPhoneAttempts: int(cfg.PreRegistrationMaxPhoneAttempts),
		BlockDuration:    time.Duration(cfg.PreRegistrationBlockMinutes) * time.Minute,
	}
	preRegistrationUC := usecase.NewPreRegistrationUseCase(uow, userRepo, preRegistrationPolicy)

	handler := httpadapter.NewHandler(httpadapter.Dependencies{
		Pool:            pool,
		TransactionRepo: txRepo,
		WebhookRepo:     webhookRepo,
		WebhookRetryRepo: webhookRetryRepo,
		WebhookAllowedIPs: cfg.WebhookAllowedIPs,
		WebhookRateLimitPerMinute: int(cfg.WebhookRateLimitPerMin),
		AlertPixPendingThreshold: int(cfg.AlertPixPendingThreshold),
		AlertPaymentPendingThreshold: int(cfg.AlertPaymentPendingThreshold),
		AlertCardPendingThreshold: int(cfg.AlertCardPendingThreshold),
		AlertTransactionHoldThreshold: int(cfg.AlertTransactionHoldThreshold),
		AlertWebhookRetryDeadThreshold: int(cfg.AlertWebhookRetryDeadThreshold),
		AccountRepo:     accountRepo,
		UserRepo:        userRepo,
		PixRepo:         pixRepo,
		PlanRepo:        planRepo,
		UserPlanRepo:    userPlanRepo,
		PricingRuleRepo: pricingRuleRepo,
		PricingVersionRepo: pricingVersionRepo,
		PlanFeatureRepo: planFeatureRepo,
		PlanLimitRepo: planLimitRepo,
		PricingCampaignRepo: pricingCampaignRepo,
		PricingCampaignRuleRepo: pricingCampaignRuleRepo,
		RoleRepo:        roleRepo,
		UserRoleRepo:    userRoleRepo,
		RoleSeparationRepo: roleSeparationRepo,
		RegulatoryProfileRepo: regulatoryProfileRepo,
		KycRepo:         kycRepo,
		KycLimitRepo:    kycLimitRepo,
		PaymentRepo:     paymentRepo,
		CardRepo:        cardRepo,
		TradeRepo:       tradeRepo,
		SettingsRepo:    settingsRepo,
		UserFeatureOverrideRepo: userFeatureOverrideRepo,
		UserLimitOverrideRepo: userLimitOverrideRepo,
		RefreshTokenRepo: refreshTokenRepo,
		PricingCache:    pricingCache,
		CreateTxUC:      createTxUC,
		RegisterPixKeyUC: registerPixKeyUC,
		SendPixUC:        sendPixUC,
		PixWebhookUC:     pixWebhookUC,
		PayCryptoWithFiatUC: payCryptoWithFiatUC,
		AuthorizeCardJITUC: authorizeCardJITUC,
		SendPixFromCryptoUC: sendPixFromCryptoUC,
		CreateInvoiceUC:     createInvoiceUC,
		PayInvoiceUC:        payInvoiceUC,
		EnsureFiatCoverageUC: ensureFiatCoverageUC,
		TokenService:   tokenService,
		LoginUC:        loginUC,
		RefreshTokenUC: refreshUC,
		LogoutUC:       logoutUC,
		PricingUC:      resolvePricingUC,
		PreRegistrationUC: preRegistrationUC,
		EventBus:        eventBus,
		VelocityChecker: velocityChecker,
		WebhookSecret:   cfg.WebhookSecret,
	})

	eventBus.Subscribe("pre_registration_email_verification", func(event events.Event) {
		log.Printf("pre-registration email verification: %v", event.Payload)
	})
	eventBus.Subscribe("pre_registration_phone_verification", func(event events.Event) {
		log.Printf("pre-registration phone verification: %v", event.Payload)
	})

	router := httpadapter.BuildHTTPHandler(handler)
	go httpadapter.StartWebhookRetryWorker(context.Background(), handler, webhookRetryRepo, 30*time.Second)

	log.Printf("servidor iniciado na porta %s", cfg.HTTPPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.HTTPPort), router); err != nil {
		log.Fatalf("falha ao iniciar servidor: %v", err)
	}
}

func waitForDatabase(pool *pgxpool.Pool) error {
	timeout := time.After(45 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := pool.Ping(ctx)
		cancel()
		if err == nil {
			return nil
		}

		log.Printf("banco indisponivel, tentando novamente: %v", err)

		select {
		case <-timeout:
			return err
		case <-ticker.C:
		}
	}
}

func velocityPolicyFromConfig(cfg configs.Config) usecase.VelocityPolicy {
	defaults := usecase.DefaultVelocityPolicy()
	return usecase.VelocityPolicy{
		MaxTxPerMinute:     defaultIfZero(cfg.VelocityMaxTxPerMinute, defaults.MaxTxPerMinute),
		MaxTxPerHour:       defaultIfZero(cfg.VelocityMaxTxPerHour, defaults.MaxTxPerHour),
		MaxAmountPerMinute: defaultIfZero(cfg.VelocityMaxAmtPerMin, defaults.MaxAmountPerMinute),
		MaxAmountPerHour:   defaultIfZero(cfg.VelocityMaxAmtPerHour, defaults.MaxAmountPerHour),
	}
}

func defaultIfZero(value, fallback int64) int64 {
	if value == 0 {
		return fallback
	}
	return value
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	data, err := os.ReadFile("db/migrations/init.sql")
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, string(data))
	return err
}

func seedAdminIfMissing(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	adminID := "00000000-0000-0000-0000-000000000001"
	if err := pool.QueryRow(ctx, "SELECT COUNT(1) FROM users WHERE id = $1", adminID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	now := time.Now().UTC()
	accountID := "00000000-0000-0000-0000-000000000001"

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, external_id, email, full_name, status, user_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		adminID,
		"zetta-admin",
		"admin@zetta.local",
		"Z-BANK Admin",
		string(entity.UserStatusActive),
		string(entity.UserTypePF),
		now,
		now,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO accounts (id, user_id, currency, scale, balance, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		accountID,
		adminID,
		"BRL",
		2,
		0,
		string(entity.AccountStatusActive),
		now,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO kyc_profiles (user_id, level, status, provider_ref, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		adminID,
		"FULL",
		"VERIFIED",
		"seed",
		now,
		now,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO kyc_limits (level, daily_limit, monthly_limit)
		VALUES ($1, $2, $3), ($4, $5, $6), ($7, $8, $9)
		ON CONFLICT (level) DO NOTHING`,
		"UNVERIFIED", int64(100_000_00), int64(500_000_00),
		"BASIC", int64(500_000_00), int64(2_000_000_00),
		"FULL", int64(5_000_000_00), int64(20_000_000_00),
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func ensureKycLimits(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO kyc_limits (level, daily_limit, monthly_limit)
		VALUES ($1, $2, $3), ($4, $5, $6), ($7, $8, $9)
		ON CONFLICT (level) DO NOTHING`,
		"UNVERIFIED", int64(100_000_00), int64(500_000_00),
		"BASIC", int64(500_000_00), int64(2_000_000_00),
		"FULL", int64(5_000_000_00), int64(20_000_000_00),
	)
	return err
}
