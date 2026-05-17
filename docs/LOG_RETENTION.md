# Politica de Retencao de Logs e Evidencias

Objetivo: garantir rastreabilidade para auditoria e compliance VASP.

## Periodos minimos
- Audit logs: 7 anos.
- Logs de transacoes/ledger: 7 anos.
- Evidencias KYC/KYB: 7 anos apos encerramento da conta.
- Webhooks e retries: 2 anos (com evidencias criticas auditadas).

## Requisitos
- Logs devem ser imutaveis (append-only ou WORM quando aplicavel).
- Backup e arquivamento devem garantir integridade.
- Exclusao deve seguir politica formal e aprovacao.

## Operacao
- Arquivamento pode ser executado via `/admin/audit/archive` (ADMIN/COMPLIANCE).
- Sempre registrar motivo e manter evidencias do lote arquivado.
- Parametros: `older_than_days` (min 30), `limit` (padrao 1000, max 50000) e `dry_run`.
- `dry_run=true` retorna a estimativa sem mover dados.
- Execucao real usa transacao para mover e apagar com consistencia.
