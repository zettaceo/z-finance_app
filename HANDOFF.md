# Handoff — Zetta Bank Core

Objetivo: permitir que novos agentes entendam o estado do backend sem reler o projeto inteiro.

## Resumo executivo
- Backend Go + PostgreSQL com ledger append-only e idempotencia.
- Fluxos: Pix (send/webhook), Payments/Boletos, Card JIT, Swap/Crypto, Invoice hibrido.
- Conversao automatica (engine) aplicada a Pix IN, Card JIT e Payments.
- Observabilidade basica (expvar, trace_id, spans por fluxo).
- Retry/backoff de webhooks com fila e worker interno.
- Endpoints admin de reconciliacao, observabilidade, auditoria e alertas.
- Rotas /admin exigem role ADMIN.
- Monetizacao: pricing versions, features por plano e campanhas.
- Conversoes auditadas em `conversion_audits`.
- Auth completo: email/senha, access/refresh token rotativo, rate limit de login, auditoria.
- Custodia: somente interfaces + mocks realistas (latencia/erro/timeout). Sem integracao real.
- Compliance VASP: modelagem de cases/events e auditoria, sem fornecedores pagos.
- Base regulatoria preparada: roles, separation rules e regulatory profiles por usuario.
- Enforcement inicial: bloqueio de roles conflitantes no assign.
- Testes e resultados documentados em `TEST_REPORT.md`.
- Worklog de continuidade em `WORKLOG.md`.
- Snapshot diario em `SNAPSHOT.md`.

## Decisoes principais
- Ledger por transacoes `CONFIRMED`. HOLD usado para pagamentos, card e compras cripto.
- Idempotency-Key obrigatoria nas rotas de escrita.
- Clean Architecture com UoW para consistencia e locks pessimistas.
- Falhas externas retornam 502 com detalhes padronizados e audit log.
- Tokens HS256 internos (sem OAuth). Refresh token armazenado por hash.
- Custodia via parceiros: `CustodyGateway` (mock), sem custodia propria.

## Como rodar
- Local: `go run ./cmd/main.go`
- Docker: `docker compose up --build --force-recreate`

## Variaveis de ambiente
- DB_URL / DATABASE_URL
- WEBHOOK_SECRET
- WEBHOOK_ALLOWED_IPS (opcional, CSV)
- WEBHOOK_RATE_LIMIT_PER_MINUTE (opcional)
- OTEL_SERVICE_NAME (opcional)
- OTEL_EXPORTER_OTLP_ENDPOINT (opcional)
- OTEL_EXPORTER_OTLP_INSECURE (opcional)
- ALERT_PIX_PENDING_THRESHOLD (opcional)
- ALERT_PAYMENT_PENDING_THRESHOLD (opcional)
- ALERT_CARD_PENDING_THRESHOLD (opcional)
- ALERT_TRANSACTION_HOLD_THRESHOLD (opcional)
- ALERT_WEBHOOK_RETRY_DEAD_THRESHOLD (opcional)
- HTTP_PORT / PORT
- APP_NAME
- AUTH_SECRET
- AUTH_ACCESS_TTL_MINUTES (default 15)
- AUTH_REFRESH_TTL_DAYS (default 30)
- AUTH_LOGIN_MAX_ATTEMPTS (default 5)
- AUTH_LOGIN_WINDOW_SECONDS (default 300)
- VELOCITY_MAX_TX_PER_MINUTE
- VELOCITY_MAX_TX_PER_HOUR
- VELOCITY_MAX_AMOUNT_PER_MINUTE
- VELOCITY_MAX_AMOUNT_PER_HOUR

## Endpoints principais (com auth)
- POST /auth/login
- POST /auth/refresh
- POST /auth/logout
- POST /transactions (auth)
- GET /transactions (auth)
- POST /pix/send (auth)
- POST /pix/send/crypto (auth)
- POST /payments/schedule (auth)
- POST /card/authorize (auth)
- POST /crypto/swap (auth)
- POST /crypto/pay (auth)
- POST /invoices (auth)
- POST /invoices/pay (auth)
- GET /accounts/{id}/balance (auth)
- GET/PUT /user/settings (auth)

Admin:
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
- GET /admin/observability/summary
- GET /admin/webhooks/retry
- GET /admin/alerts/check
- GET /admin/audit/logs
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

Publicos:
- GET /health, /health/db
- GET /debug/vars
- Webhooks com HMAC: /webhooks/*

## Ultimas entregas (P3–P5 + Auth + Observabilidade)
- P3: Crypto→Pix + Invoice hibrido (Pix + USDT).
- P4: Engine aplicada a Card JIT e Payments (auto venda cripto).
- P5: Observabilidade basica (expvar + spans + trace_id).
- Webhooks: anti-replay + retry/backoff + hardening (allowlist/rate limit).
- Admin: reconciliacao, observabilidade, auditoria, alertas.
- Auth: email/senha + access/refresh token rotativo, rate limit login, auditoria.
- Custodia: `CustodyGateway` com mock realista.
- Compliance: entidades e repos de cases/events + tabelas.

## Pontos de atencao
- Pix REJECTED corrigido: agora HOLD -> REJECTED + reversal e ledger seguro.
- Auth exige `AUTH_SECRET` e protege rotas sensiveis.
- Custodia ainda mock; pronto para plugar parceiro depois.
- Compliance esta apenas modelado (sem regras/fluxos automaticos).

## Proximos passos sugeridos
- Frontend seguro com fluxo de login/refresh.
- Regras de compliance (KYC/AML/KYT interno) e dashboards.
- Integração real com parceiro custodial e reconciliação.
- Observabilidade com OpenTelemetry (collector) + dashboards.
- Padronizar deploy usando `OPS_CHECKLIST.md`.
