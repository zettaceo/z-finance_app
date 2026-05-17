# Analise Full Stack — Zetta Core Banking (Z-FINANCE)

Ultima atualizacao: 2026-01-29

Escopo: backend, frontend, dados, seguranca, compliance, observabilidade e operacao. A analise considera o estado real do codigo e a documentacao atual.

---

## 1) Resumo executivo

O projeto apresenta uma base solida para MVP regulatorio: ledger append-only, idempotencia, UoW com locks, auth com tokens rotativos, webhooks assinados e cobertura ampla de operacoes (Pix, Payments, Crypto, Card JIT, Invoice hibrido). O frontend cobre os endpoints e serve a demo.

Os principais riscos estao em: (1) governanca e compliance avancada (RBAC completo, AML/KYT, workflow de casos), (2) observabilidade e operacao (dashboards, runbooks) e (3) alinhamento rigoroso entre identidade do usuario (token) e payloads.

---

## 2) Cobertura do prompt vs estado atual

### Pricing Engine e monetizacao
- **Implementado**: `internal/pricing/engine/engine.go` + `internal/usecase/resolve_pricing.go`.
- **Tabelas**: `plans`, `user_plans`, `pricing_rules` estao em `db/migrations/init.sql`.
- **Entidades**: `Plan`, `UserPlan`, `PricingRule`, `FeeResult` estao em `internal/entity/pricing.go`.
- **Integracao**: fluxos com pricing em `send_pix.go`, `send_pix_from_crypto.go`, `authorize_card_jit.go`, `pay_crypto_with_fiat.go`, `invoice.go`, `ensure_fiat_coverage.go`, `auto_convert_pix.go`.
- **GAP**: regras de pricing nao sao seedadas por padrao; sem UI/admin para manutencao de regras.

### Auto-conversao
- **Implementado**: `auto_convert_pix.go`, `ensure_fiat_coverage.go`, `conversion/engine`.
- **GAP**: falta matriz de estados formal e testes de integridade de fluxo.

### Auditoria
- **Implementado**: `audit_logs` e uso em use cases criticos, webhooks e auth.
- **GAP**: consolidar runbooks e alertas baseados em audit trail.

---

## 3) Arquitetura (Clean Architecture)

### Pontos fortes
- Separacao de camadas clara: Adapters, UseCases, Entities, Repositories, Infra.
- Unit of Work via Postgres com locks pessimistas.
- Outbox + event bus (in-memory) para evolucao futura.

### Pontos a fortalecer
- Contratos HTTP ainda sem OpenAPI formal (documentacao manual).
- Falta pipeline de testes de concorrencia/ledger.
- Ausencia de DLQ para eventos/outbox (retry existe; falta dead-letter).

---

## 4) Ledger, consistencia e dados

### Pontos fortes
- Ledger append-only com constraints: `amount > 0`, `net_amount = amount - fee`.
- Idempotencia com indice unico `transactions (account_id, idempotency_key)`.

### Lacunas e riscos
- **`accounts.balance`**: coluna existe, mas o saldo e derivado do ledger. Recomendacao: tratar como cache controlado ou remover para evitar divergencia.
- **HOLD vs CONFIRMED**: falta matriz formal de estados e testes que validem reversals e reconciliacoes.

---

## 5) Autenticacao e seguranca

### Pontos fortes
- Access/refresh token rotativo.
- Rate limit de login configuravel.
- Middleware `requireAuth` aplica bearer token nas rotas sensiveis.

### Lacunas e riscos
- **User ID no payload**: varios endpoints exigem `user_id` no corpo e comparam com o token. Recomendado: derivar `user_id` exclusivamente do token e ignorar payload.
- **Rotacao de segredo**: nao ha politica de rotacao do `AUTH_SECRET` nem procedimento documentado.

---

## 6) Webhooks e integracoes

### Pontos fortes
- Assinatura HMAC (`WEBHOOK_SECRET`).
- Idempotencia de webhooks via tabela `webhook_events`.

### Lacunas e riscos
- **Anti-replay**: validacao de timestamp/nonce implementada.
- **Retry/reconciliacao**: retry/backoff e reconciliacao via endpoints admin implementados.

---

## 7) Compliance, KYC, risco

### Pontos fortes
- KYC levels e limites diarios/mensais.
- Velocity checker por minuto/hora.
- Compliance cases/events modelados.

### Lacunas e riscos
- **AML/KYT**: regras basicas, sem scoring e sem workflow de triagem.
- **RBAC**: roles existem e admin esta protegido; ainda falta matriz completa (operador/auditor/compliance).
- **Trilha regulatoria**: faltam evidencias de aprovacao, justificativas e retencao de documentos.

---

## 8) Observabilidade e operacao

### Pontos fortes
- `expvar` + request metrics.
- Trace id basico e headers padronizados.

### Lacunas e riscos
- **OpenTelemetry** integrado de forma basica; faltam dashboards.
- **Alertas e dashboards** precisam de evolucao.
- **Backup/restore e DR** nao documentados.

---

## 9) Frontend (demo)

### Pontos fortes
- UI multi-tela com login, rotas hash e cobertura ampla de endpoints.
- Configuracao white-label basica e logs de resposta.

### Lacunas e riscos
- Tokens em `sessionStorage` (ok para demo; nao para producao).
- Dependencia de CORS correto (env `CORS_ORIGINS`).
- UX de erro/estado (loading, retry, mensagens persistentes) ainda basico.

---

## 10) Riscos prioritarios

### P0 (nao negociar)
- Formalizar contratos (OpenAPI) e estados de Pix/Payments/Card/Invoice.
- Remover dependencia de `user_id` no payload para autorizacao.
- Implementar estrategia de retry/reconciliacao de webhooks.
- Padronizar auditoria em eventos sensiveis.

### P1 (alto impacto)
- Observabilidade avancada (OpenTelemetry, dashboards, alertas).
- Politicas de backup/restore e retencao de logs/outbox.
- Testes de integridade e concorrencia do ledger.

### P2 (regulatorio e escala)
- AML/KYT com risk scoring e workflow de compliance.
- RBAC e segregacao de funcoes.
- Integracoes reais com parceiros (BaaS/PSP/Exchange).

---

## 11) Conclusao

O sistema esta tecnicamente robusto para MVP demo-regulatorio, com base correta de ledger, idempotencia e pricing engine integrado. Para nivel investidor/regulatorio, o foco deve ser elevar governanca, resiliencia e observabilidade, sem quebrar a base atual.
