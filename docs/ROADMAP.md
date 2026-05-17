# Roadmap — Z-FINANCE (Full Stack)

Objetivo: elevar o MVP para um produto demo-regulatorio forte, mantendo compatibilidade e sem quebrar fluxos existentes.

---

## P0 — Contratos e seguranca de fluxo (2-4 semanas)

### Metas
- Formalizar contratos HTTP (OpenAPI) e estados de operacao.
- Consolidar regras de integridade e auditoria.
- Endurecer seguranca de payload vs token.

### Entregaveis
- `docs/api/openapi.yaml` cobrindo todas as rotas atuais.
- `docs/state_machines.md` com estados de Pix, Payments, Card JIT, Transactions, Invoices.
- Ajustes de handlers: derivar `user_id` do token e validar ownership sem depender de payload.
- Auditoria de eventos sensiveis (falha de assinatura, refresh revogado, tentativas repetidas).

Estado atual:
- Anti-replay e auditoria sensivel implementados.

### Criterios de pronto
- OpenAPI valida requests/responses em 100% das rotas existentes.
- Estados documentados e alinhados com DB.
- Nenhuma rota critica aceita `user_id` divergente do token.

---

## P1 — Resiliencia operacional (4-6 semanas)

### Metas
- Resiliencia em webhooks e conciliacao de estados.
- Observabilidade end-to-end.
- Testes de integridade e concorrencia.

### Entregaveis
- Retry/backoff e reconciliacao para webhooks pendentes.
- OpenTelemetry + dashboards basicos (SLOs, falhas, latencias).
- Testes de ledger (double spend, idempotencia, HOLD/reversal).
- Politicas de backup/restore e retencao documentadas.

Estado atual:
- Retry/backoff e reconciliacao via endpoints admin implementados.

### Criterios de pronto
- Webhooks com retry e reconciliacao automatica.
- Tracing distribuido visivel em pelo menos 3 fluxos criticos.
- Suite de testes de integridade executando em CI.

---

## P2 — Compliance e governanca (6-8 semanas)

### Metas
- Aumentar maturidade regulatoria.
- Adicionar segregacao de funcoes e trilhas de aprovacao.

### Entregaveis
- RBAC (operador, auditor, compliance, admin).
- AML/KYT com regras configuraveis e risk scoring.
- Workflow de compliance com evidencias e justificativas.

### Criterios de pronto
- Roles aplicadas a rotas sensiveis e auditadas.
- Casos de compliance com historico de aprovacao.

---

## P3 — Integracoes reais e escala (8+ semanas)

### Metas
- Conectar parceiros reais.
- Elevar performance e escala.

### Entregaveis
- Adaptadores reais para BaaS/PSP/Exchange.
- Benchmarks p95/p99 para Card JIT (< 200ms).
- Particionamento/estrategia de escala para ledger/audit.

### Criterios de pronto
- Sandbox/staging com contratos reais.
- Capacidade validada para volume de transacoes alto.

---

## Dependencias

- P0 bloqueia P1 (contratos e seguranca).
- P1 bloqueia P2 (observabilidade e resiliencia).
- P2 bloqueia P3 (governanca e compliance).
