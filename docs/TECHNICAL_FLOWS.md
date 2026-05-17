# Fluxos Tecnicos — Core Financeiro

Este documento descreve, de forma direta, como o dinheiro entra, se move e e controlado pelo core.

## Como o dinheiro entra
- Entradas fiat (PIX/transferencias) geram uma transacao de ledger `DEPOSIT` com `amount`, `fee`, `net_amount`.
- Entradas cripto geram registros em `crypto_transfers` e conversoes sao auditadas em `conversion_audits`.
- Toda entrada passa por idempotencia via `Idempotency-Key` em operacoes de escrita.

## Como o dinheiro se move
- Toda mutacao de saldo e representada por eventos de ledger em `transactions`.
- O saldo e calculado somente a partir de `transactions` com `status = CONFIRMED`.
- Reversoes nao alteram o evento original; criam nova transacao `REVERSAL`.
- Conversoes (fiat <-> cripto) registram cotacao, spread e fonte de liquidez em `conversion_audits`.

## Como o risco e controlado
- Idempotencia obrigatoria evita duplicidade de operacoes.
- Velocidade e limites (KYC, plano e soft limits) bloqueiam volumes excessivos.
- Webhooks validam assinatura HMAC, timestamp e nonce (anti replay).
- Audit logs registram mudancas administrativas e eventos sensiveis.

## Como o sistema impede erro
- Ledger append-only protegido por constraints/triggers no banco.
- Unit of Work garante consistencia transacional e locks pessimistas.
- Falhas externas nao alteram saldo sem evento confirmado.
- Erros externos retornam 502 com detalhes de provider/operation.

## Como escalar sem quebrar
- Pricing, planos, limites e features sao configuraveis via dados (sem deploy).
- Outbox interno para eventos e retries de webhook.
- Interfaces desacopladas para futuras integracoes (Wallet/DEX/IA).

## Garantias principais
- Nenhuma mutacao de saldo ocorre sem evento de ledger.
- Nenhuma operacao financeira acontece sem idempotencia.
- Toda conversao e auditavel por transacao relacionada.
