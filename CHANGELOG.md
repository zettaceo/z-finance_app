CHANGELOG - Z-FINANCE

2026-01-29
- Hardening final do backend (admin por role, auditoria reforcada).
- Tratamento padronizado de falhas externas com erro 502 e audit log.
- Documentacao tecnica atualizada (fluxos, resiliencia, operacao).
- Conversoes auditadas em `conversion_audits` nos fluxos criticos.

2026-01-22
- Documentacao de handoff atualizada (README/STATUS/ARCHITECTURE).
- Frontend estatico multi-tela com login dedicado e rotas por hash.
- Branding e copy regulatorio ajustados para evitar termos de "banco/bank".
- CORS configuravel por env (CORS_ORIGINS) para demos externas.

2026-01-25
- Seed de pricing rules por plano e perfil (PF/PJ) via migracao.
- Endpoints admin para planos e pricing rules (CRUD basico).
- OpenAPI atualizado com endpoints admin + demo guide com passos de pricing.
- Webhooks com protecao anti-replay (timestamp + nonce).
- Auditoria sensivel adicionada para auth e validacoes de webhook.
- Endpoints de reconciliacao (summary + pending) para pendencias.
- Retry/backoff automatico para falhas 5xx em webhooks (fila + worker).
- Endpoint admin para listar retries e metricas expvar.
- Endpoint admin de observabilidade (summary).
- Hardening de webhooks: allowlist IP + rate limit por minuto.
- Endpoint admin para consulta de auditoria com filtros.
- OpenTelemetry basico (OTLP HTTP + tracing HTTP).
- Alertas operacionais via thresholds e endpoint admin.
- Tracing por fluxo (pix/payments/card/swap/invoices).
- Demo guide atualizado com observabilidade/alertas/auditoria.
- Ops checklist para deploy e smoke tests.
- Base regulatoria: roles, separation rules e regulatory profiles.
- Admin endpoints: roles, user-roles, separation rules, regulatory profiles.
- Pricing engine consolidado: versoes, features por plano, campanhas e overrides.
- Enforcement inicial: bloqueio de roles conflitantes (separation of duties).
- UX por niveis: snapshot de capacidades no login/refresh e /auth/me.
- Soft limits por plano (history items/days) + plan limits admin.
# Changelog

## 2026-01-22
- P3: Crypto→Pix, Invoice hibrido (Pix + USDT).
- P4: Engine aplicada a Card JIT e Payments (auto venda cripto).
- P5: Observabilidade basica (expvar + trace_id + spans).
- Auth: login email/senha, access/refresh token rotativo, rate limit login, auditoria.
- Custodia: interface `CustodyGateway` + mock realista.
- Compliance: cases/events (modelagem + repos + tabelas).
- Correcao Pix REJECTED (HOLD -> REJECTED + reversal).
