Relatorio de testes — Z-FINANCE

Data: 2026-01-31

## 2026-01-29
1) Testes automatizados (Go)
- go test ./...

Resultado: OK.
Observacao: corrigida duplicacao do handler `handleAdminAuditArchive` antes da execucao.

Resumo
- Testes unitarios e de integracao (Go) executados com sucesso.
- E2E manual executado com backend em execucao.
- Falha encontrada em auth (refresh/logout) corrigida e validada.

1) Testes automatizados (Go)
- go test ./...
- go test ./... -race
- go test ./... -count=1
- go test ./... -race -count=1

Resultado: todos passaram.

2) Smoke tests (sem auth)
- GET /health -> 200
- GET /health/db -> 200
- GET /debug/vars -> 200

3) E2E manual (auth)
Fluxo:
- POST /auth/login -> 200
- GET /auth/me -> 200
- POST /auth/refresh -> 200 (com refresh_token valido)
- POST /auth/logout -> 200 (com refresh_token apos rotacao)

Resultado: OK.
Observacao: logout com refresh_token antigo falha (esperado, pois o refresh faz rotacao).

4) E2E manual (fluxos principais)
Fluxo:
- POST /auth/login -> 200
- GET /auth/me -> 200
- GET /user/settings -> 200
- POST /payments/validate -> 200
- POST /invoices -> 201
- GET /accounts/{account_id}/balance -> 200

Resultado: OK.
Observacao: o endpoint de balance exige account_id (user_id retorna 403).

5) Incidente corrigido (refresh/logout)
Sintoma:
- /auth/refresh -> 401 REFRESH_INVALID
- /auth/logout -> 400 LOGOUT_FAILED

Causa:
- Atualizacao do refresh_tokens falhava ao gravar replaced_by (campo UUID) com texto.

Correcao:
- Cast explicitado para UUID no UPDATE de refresh_tokens:
  replaced_by = NULLIF($2, '')::uuid

Status: resolvido e validado no E2E.

---

Data: 2026-01-29

Resumo
- go test ./... executado com sucesso.

1) Testes automatizados (Go)
- go test ./...

Resultado: OK.

2) Testes automatizados (Go)
- go test ./... -race

Resultado: OK.

3) Testes automatizados (Go)
- go test ./... -count=1

Resultado: OK.

4) Testes automatizados (Go)
- go test ./... -race -count=1

Resultado: OK.

---

Data: 2026-01-29

Resumo
- go test ./... executado com sucesso apos pre-cadastro.

1) Testes automatizados (Go)
- go test ./...

Resultado: OK.
