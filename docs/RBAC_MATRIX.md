# Matriz RBAC (Baseline)

Roles suportadas: ADMIN, COMPLIANCE, AUDIT, OPS, VIEWER.

| Capacidade | ADMIN | COMPLIANCE | AUDIT | OPS | VIEWER |
| --- | --- | --- | --- | --- | --- |
| Gerenciar planos e pricing | sim | nao | leitura | nao | nao |
| Ajustar limites e features de usuario | sim | sim | leitura | nao | nao |
| Congelar conta / resetar sessao | sim | sim | leitura | sim | nao |
| Ver ledger detalhado | sim | sim | sim | sim | leitura |
| Ver audit logs | sim | sim | sim | leitura | leitura |
| Gerenciar roles e separacao | sim | nao | leitura | nao | nao |
| Gerenciar perfis regulatorios | sim | sim | leitura | nao | nao |
| Reconciliacao e alertas | sim | sim | leitura | sim | leitura |

Observacoes:
- ADMIN e o unico role com poder de administracao global.
- COMPLIANCE opera casos e acessa dados sensiveis.
- AUDIT apenas leitura, sem alteracoes.
- OPS executa operacoes operacionais aprovadas.
