Worklog — Z-FINANCE


2026-01-29
- Corrigida duplicacao de handleAdminAuditArchive em handler.go.
- Politica de retencao atualizada com parametros e dry_run.
- Testes go test ./... executados com sucesso.

2026-01-29
- Pre-cadastro institucional implementado (schema, repos, usecase e handlers).
- Endpoints publicos e verificacao dupla com antifraude adicionados.
- OpenAPI atualizado com pre-cadastro.

2026-01-31
- Criadas regras de continuidade e anti-perda em `.cursor/rules/context-continuity.mdc`.
- Documentado relatorio de testes em `TEST_REPORT.md` e referencias em docs.
- Corrigido refresh/logout (replaced_by UUID) e validado E2E auth.

## 2026-01-29
- Removida duplicacao do handler de arquivamento de auditoria.
- `go test ./...` executado com sucesso.

## 2026-01-31
- Adicionadas politicas de compliance (AML/KYC/KYB, RBAC, retencao, segredos).
- Atualizado README e OPS_CHECKLIST com referencias.

## 2026-01-31
- RBAC: restricao granular em rotas admin e sub-acoes (usuarios/contas).
- Admin: roles aplicadas para leitura e acoes sensiveis.

## 2026-01-31
- RBAC: restricao granular em rotas admin e sub-acoes (usuarios/contas).
- Admin: roles aplicadas para leitura e acoes sensiveis.
