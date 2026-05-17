Frontend Overview - Z-FINANCE

Escopo
- UI estatica (HTML/CSS/JS) para demo de investidores.
- Login/refresh/logout com tokens rotativos.
- Navegacao por hash: #/dashboard, #/operacoes, #/compliance, #/config, #/observabilidade.

Arquivos principais
- frontend/index.html: estrutura das telas e formularios.
- frontend/styles.css: tema escuro premium + responsivo mobile-first.
- frontend/app.js: estado, auth, chamadas API, roteamento e logs.

Configuracao local
- API Base URL configuravel no painel (persistido em localStorage).
- Tokens e expiracao em sessionStorage.
- CORS liberado no backend via CORS_ORIGINS.

Fluxo de login
1) Preencha email/senha.
2) POST /auth/login retorna access/refresh tokens.
3) UI libera acesso ao app e oculta a tela de login.
4) Refresh token rotativo em /auth/refresh.
5) /auth/me retorna plano, ux_mode e features (contexto de UX).
6) Capability snapshot no login/refresh para renderizacao por niveis.

Cobertura de endpoints (frontend)
- Auth: /auth/login, /auth/refresh, /auth/logout
- Session: /auth/me
- Ledger: /accounts/{id}/balance
- Transactions: /transactions, /transactions/confirm, /transactions/reject, /transactions/reverse
- Pix: /pix/send, /pix/keys
- Payments: /payments/validate, /payments/schedule
- Crypto: /crypto/pay, /crypto/swap
- Pricing: /pricing/quote (GET/PUT)
- KYC: /kyc/limits (GET/PUT), /kyc/upgrade
- User settings: /user/settings
- Invoices: /invoices, /invoices/pay
- Observabilidade: /health, /health/db, /debug/vars

Branding e copy
- Nome padrao: Z-FINANCE (white-label via inputs de marca).
- Evitar linguagem regulatoria sensivel (sem "banco/bank").
- Disclaimers no footer.

Riscos/pendencias
- Revisar todos os textos com enfoque regulatorio local.
- Testes E2E ainda nao implementados.
- Frontend esta em pausa; evolucao de UX sera retomada depois do hardening do backend.
