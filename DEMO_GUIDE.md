Guia rapido de demo - Z-FINANCE

Objetivo
Demo rapida do frontend com backend ativo para investidores/reguladores.

Pre-requisitos
- Backend rodando (Go local ou Docker Compose).
- Frontend estatico servido em http://<HOST>:5174.
- CORS liberado no backend (env CORS_ORIGINS).

Passo a passo (5-8 min)
1) Suba o backend:
   - Local: go run ./cmd/main.go
   - Docker: docker compose up --build --force-recreate
2) Suba o frontend:
   - cd frontend
   - python3 -m http.server 5174 --bind 0.0.0.0
3) Abra o app: http://<HOST>:5174
4) Em Configuracao:
   - API Base URL = http://<HOST>:8080
   - User ID/Account ID conforme seed/demo
5) Login:
   - Email/senha do usuario seedado (ou o usuario criado pelo seeder).
   - Opcional: GET /auth/me para plano + ux_mode + features.
6) Dashboard:
   - Clique em "Atualizar saldo" e "Ultimas transacoes".
7) Operacoes:
   - Pix/Payments/Crypto: executar validacao ou envio com Idempotency-Key automatico.
8) Compliance:
   - KYC limits (GET/PUT) e KYC upgrade.

9) Pricing (admin):
   - Todas as rotas /admin exigem role ADMIN.
   - Listar planos: GET /admin/plans
   - Atribuir plano ao usuario: POST /admin/user-plans (plan_id ou plan_code)
   - Listar regras: GET /admin/pricing/rules?plan_id=...&user_type=PF
   - Ajustar regra: PUT /admin/pricing/rules/{id}
   - Versoes: GET /admin/pricing/versions | /active
   - Features: PUT /admin/plan-features
   - Limites: PUT /admin/plan-limits
   - Campanhas: GET/POST /admin/pricing/campaigns
   - Campanhas rules: GET/POST /admin/pricing/campaigns/rules

10) Reconciliacao (admin):
   - Resumo de pendencias: GET /admin/reconcile/summary?older_than_minutes=30
   - Listar pendencias: GET /admin/reconcile/pending?type=pix&older_than_minutes=30&limit=50

11) Observabilidade (admin):
   - Resumo consolidado: GET /admin/observability/summary?older_than_minutes=30
   - Retries webhook: GET /admin/webhooks/retry?status=PENDING&limit=50

12) Alertas (admin):
   - Checagem de thresholds: GET /admin/alerts/check?older_than_minutes=30

13) Auditoria (admin):
   - Logs com filtros: GET /admin/audit/logs?action=AUTH_LOGIN_SUCCESS&limit=50

Falhas externas
- Dependencias externas (exchange/custody) retornam 502 com detalhes padronizados.

14) Compliance base (admin):
   - Roles: GET/POST /admin/roles
   - User roles: GET/POST/DELETE /admin/user-roles
   - Separation rules: GET/POST/DELETE /admin/roles/separation
   - Regulatory profile: GET/PUT /admin/regulatory-profiles

Dicas de demo
- Mostrar o log de respostas no painel de Observabilidade.
- Ressaltar que o ledger e append-only e audita tudo.
- Evitar termos “banco/bank”; usar “plataforma financeira”.

Checklist antes da apresentacao
- Backend acessivel externamente (IP/porta corretos).
- CORS_ORIGINS inclui o host do frontend.
- Logo e textos alinhados com marca/regulatorio.
- Testes executados (ver `TEST_REPORT.md`).
