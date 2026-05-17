Zetta Finance Platform (Z-FINANCE) — Status do Projeto

Ultima atualizacao: 2026-01-29

Resumo
- Backend em Go com PostgreSQL, Clean Architecture, Docker (opcional).
- Ledger append-only com int64, idempotencia e auditoria.
- Conversoes auditadas com `conversion_audits`.
- Unit of Work implementado para o fluxo de criacao de transacoes.
- Testes documentados em `TEST_REPORT.md`.
- Camada cripto: address intelligence (checksum), crypto pay, crypto->pix e auto conversao Pix.
- Card JIT com fallback cripto e timeout.
- Invoice hibrido (Pix + USDT).
- Auth: email/senha, access/refresh token rotativo, rate limit no login.
- Custodia white label: interfaces + mock realista (sem integracao externa).
- Compliance VASP: cases/events modelados + auditoria.
- Modulos: Core/Ledger, KYC/Limits, Pix, Payments, Card JIT, Swap/Crypto, Pricing, User Settings, Audit, Auth, Compliance.
- Responses HTTP padronizadas (envelope + headers) e handlers em adapters.
- Falhas externas retornam 502 com detalhes padronizados e audit log.
- Modulo de risco/velocity com limites diarios/mensais + janelas por minuto/hora.
- Frontend estatico multi-tela com login dedicado e rotas por hash (dashboard/operacoes/compliance/config/observabilidade).
- Marca e copy regulatorio ajustados para Z-FINANCE.

Como subir (Neon)
1) go run ./cmd/main.go

Como subir (Docker local - opcional)
1) docker compose down
2) docker compose up --build --force-recreate
3) docker compose ps

Env vars
- DB_URL (Neon ou Postgres local; DATABASE_URL ainda aceito)
- WEBHOOK_SECRET (no compose)
- WEBHOOK_ALLOWED_IPS (opcional, lista CSV de IPs permitidos para webhooks)
- WEBHOOK_RATE_LIMIT_PER_MINUTE (opcional, default 120)
- OTEL_SERVICE_NAME (opcional, default APP_NAME)
- OTEL_EXPORTER_OTLP_ENDPOINT (opcional, host:porta do collector)
- OTEL_EXPORTER_OTLP_INSECURE (opcional, default true)
- ALERT_PIX_PENDING_THRESHOLD (opcional, default 0)
- ALERT_PAYMENT_PENDING_THRESHOLD (opcional, default 0)
- ALERT_CARD_PENDING_THRESHOLD (opcional, default 0)
- ALERT_TRANSACTION_HOLD_THRESHOLD (opcional, default 0)
- ALERT_WEBHOOK_RETRY_DEAD_THRESHOLD (opcional, default 0)
- AUTH_SECRET
- CORS_ORIGINS (origens liberadas para o frontend)
- AUTH_ACCESS_TTL_MINUTES (default 15)
- AUTH_REFRESH_TTL_DAYS (default 30)
- AUTH_LOGIN_MAX_ATTEMPTS (default 5)
- AUTH_LOGIN_WINDOW_SECONDS (default 300)
- HTTP_PORT (opcional; PORT ainda aceito)
- APP_NAME (opcional)
- VELOCITY_MAX_TX_PER_MINUTE (opcional)
- VELOCITY_MAX_TX_PER_HOUR (opcional)
- VELOCITY_MAX_AMOUNT_PER_MINUTE (opcional)
- VELOCITY_MAX_AMOUNT_PER_HOUR (opcional)

Seed automatico
- Seed acontece se o admin nao existir no banco.
- User: Z-FINANCE Admin
- UserID: 00000000-0000-0000-0000-000000000001
- AccountID: 00000000-0000-0000-0000-000000000001
- Currency: BRL, Scale 2
- KYC Profile: FULL/VERIFIED
- KYC Limits: UNVERIFIED, BASIC, FULL (daily/monthly)

Principios aplicados
- Valores monetarios em int64 (atomic units).
- Ledger append-only; saldo calculado por transacoes CONFIRMED.
- Idempotency-Key obrigatoria para criacao de transacao.
- Locks e tx ACID via Unit of Work.
- HOLD para PAYMENT/TRADE_BUY/CARD_AUTH.
- Webhooks com HMAC.
- Responses HTTP com envelope {data|error|request_id} e headers de seguranca.
- User ID derivado do token em rotas autenticadas (payload opcional).

Tabelas principais (init.sql)
- users, kyc_profiles, kyc_limits
- roles, user_roles, role_separation_rules
- regulatory_profiles
- pricing_versions, pricing_rules
- plan_features
- pricing_campaigns, pricing_campaign_rules
- accounts
- transactions (reversal_of_transaction_id)
- audit_logs
- refresh_tokens
- login_audits
- compliance_cases
- compliance_events
- outbox_events
- webhook_events
- pix_transfers
- pix_keys
- payments
- card_authorizations
- trade_orders (idempotency_key, debit/credit account ids)
- user_settings
- crypto_transfers
- conversion_rules
- invoices

Endpoints principais
Health
- GET /health
- GET /health/db
- GET /debug/vars (metrics basicas)

Auth
- POST /auth/login
- POST /auth/refresh
- POST /auth/logout
- GET /auth/me

UX por niveis
- plan_code + ux_mode + features + limits + sugestoes em /auth/me.

Transactions
- POST /transactions (auth + Idempotency-Key)
- GET /transactions?account_id=...&limit=...&from=...&to=...&cursor=... (auth)
- POST /transactions/confirm (auth + Idempotency-Key)
- POST /transactions/reject (auth + Idempotency-Key)
- POST /transactions/reverse (auth + Idempotency-Key)

Ledger
- GET /accounts/{id}/balance (auth)

KYC
- GET /kyc/limits (auth)
- GET /kyc/limits?level=FULL (auth)
- PUT /kyc/limits (auth)
- POST /kyc/upgrade (auth)

PIX
- POST /pix/send (auth + Idempotency-Key)
- POST /pix/send/crypto (auth + Idempotency-Key)
- POST /webhooks/pix/receive (HMAC)

Payments/Boletos
- POST /payments/validate
- POST /payments/schedule (auth + Idempotency-Key)
- POST /webhooks/payments/confirm (HMAC)
- POST /webhooks/payments/reject (HMAC)

Card JIT
- POST /card/authorize (auth + Idempotency-Key)
- POST /webhooks/card/confirm (HMAC)
- POST /webhooks/card/reject (HMAC)

Swap/Crypto
- POST /crypto/swap (auth + Idempotency-Key, auto_execute opcional)
- POST /crypto/pay (auth + Idempotency-Key)

Pricing
- GET /pricing/quote?pair=BTC/BRL
- PUT /pricing/quote

Admin (pricing/plans)
- GET /admin/plans
- POST /admin/plans
- POST /admin/user-plans
- GET /admin/pricing/rules
- POST /admin/pricing/rules
- PUT /admin/pricing/rules/{id}
- GET /admin/pricing/versions
- POST /admin/pricing/versions
- GET /admin/pricing/versions/active
- PUT /admin/pricing/versions/{id}
- GET /admin/plan-features
- PUT /admin/plan-features
- GET /admin/plan-limits
- PUT /admin/plan-limits
- GET /admin/pricing/campaigns
- POST /admin/pricing/campaigns
- PUT /admin/pricing/campaigns/{id}
- GET /admin/pricing/campaigns/rules
- POST /admin/pricing/campaigns/rules
- GET /admin/reconcile/summary
- GET /admin/reconcile/pending
- GET /admin/roles
- POST /admin/roles
- GET /admin/user-roles
- POST /admin/user-roles
- DELETE /admin/user-roles
- GET /admin/roles/separation
- POST /admin/roles/separation
- DELETE /admin/roles/separation
- GET /admin/regulatory-profiles
- PUT /admin/regulatory-profiles
Observacao: todas as rotas /admin exigem role ADMIN.

Enforcement
- Bloqueio de roles conflitantes (separation of duties) no assign.

User Settings
- GET /user/settings?user_id=... (auth)
- PUT /user/settings (auth)

Invoices
- POST /invoices (auth)
- POST /invoices/pay (auth + Idempotency-Key)

Auditoria
- appendAudit grava em audit_logs e publica evento no event bus in-memory.
- appendAudit publica no outbox_events para processamento assinc.

Webhooks
- Headers: X-Signature (HMAC SHA256 do body), X-Timestamp (unix ou RFC3339), X-Nonce (unico)
- Endpoints: /webhooks/transactions/confirm | /reject | /webhooks/pix/receive | /webhooks/payments/confirm | /reject | /webhooks/card/confirm | /reject
Retry/backoff
- Fila de retry para falhas 5xx em webhooks (backoff exponencial, max 5 tentativas).
- Worker interno executa a cada 30s e marca SUCCEEDED/DEAD.
- Observabilidade via /debug/vars (expvar) e /admin/webhooks/retry.
Resumo de observabilidade
- GET /admin/observability/summary agrega pendencias (reconcile) + retries.
Auditoria (admin)
- GET /admin/audit/logs?user_id=&action=&entity_type=&from=&to=&limit=
Alertas (admin)
- GET /admin/alerts/check?older_than_minutes=
Tracing por fluxo
- Spans em pix, payments, card, swap e invoices com atributos basicos.
Operacoes
- Checklist de deploy e smoke tests em `OPS_CHECKLIST.md`.

Auditoria sensivel
- Auth: login success/fail, rate limit, refresh, logout.
- Webhooks: assinatura/timestamp/nonce invalido e replay detectado.

Arquivos chave
- cmd/main.go (bootstrap + migracao + seed + HTTP server)
- configs/config.go (configuracao centralizada)
- internal/usecase/auth.go
- internal/usecase/auth_tokens.go
- internal/infra/repository/postgres/refresh_token_repository.go
- internal/infra/repository/postgres/login_audit_repository.go
- internal/core/ports/unit_of_work.go
- internal/infra/repository/postgres/unit_of_work.go
- internal/crypto/address/*
- internal/infra/crypto/mock_exchange.go
- internal/infra/crypto/mock_transfer.go
- internal/infra/custody/mock_gateway.go
- internal/adapter/http/*
- internal/entity/*
- internal/infra/repository/postgres/*
- internal/infra/repository/postgres/velocity_repository.go
- internal/infra/repository/postgres/webhook_repository.go
- db/migrations/init.sql
- docker-compose.yml, Dockerfile
- internal/infra/pricing/cache.go
- internal/infra/events/bus.go
- internal/usecase/velocity_checker.go
- internal/testutil/http.go

Pendencias/Proximos passos
- Formalizar copy regulatorio (evitar termos “banco/bank”).
- Regras de compliance (AML/KYC/KYT internas).
- Integracao com parceiro custodial real (sem custodia propria).
- Observabilidade avancada (OpenTelemetry + dashboards).

Pricing (seed padrao)
- Regras de pricing por plano (FREE/PRO/BUSINESS) e perfil (PF/PJ) sem hardcode nos use cases.
- Percentuais em basis points (ex.: 30 = 0.30%).
- Plan features e plan limits seedados para FREE/PRO/BUSINESS.

IDs de to-dos (Cursor)
- todo-1: Criar docs/PROJECT_BIBLE.md com prompt integral
- todo-2: Escrever docs/ARCHITECT_REVIEW.md (analise full-stack)
- todo-3: Escrever docs/ROADMAP.md com fases e prioridades
