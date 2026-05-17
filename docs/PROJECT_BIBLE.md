# Biblia do Projeto — Zetta Core Banking (Z-BANCK)

Este documento e a fonte primaria de visao e requisitos. Ele preserva o prompt original integral e apresenta requisitos derivados para orientar evolucao tecnica e produto.

---

## Estado atual (resumo rapido)
- Backend hardenizado com idempotencia, ledger append-only e auditoria.
- Conversoes auditadas em `conversion_audits`.
- Falhas externas retornam 502 com detalhes e audit log.
- Rotas /admin exigem role ADMIN.
- Documentacao tecnica atualizada (TECHNICAL_FLOWS, ARCHITECTURE, OPS_CHECKLIST).

---

## Prompt original (integral)

```
ATUE COMO:
Senior Backend Architect especializado em Core Banking, Fintechs reguladas, Pricing Engines e Ledger Financeiro Append-Only.

Você está trabalhando no projeto Zetta Core Banking (Z-BANCK), já em produção em VPS, com:
- Ledger append-only
- Unit of Work
- Idempotência
- Pix, Crypto, Card JIT, Invoice híbrido
- Clean Architecture (Entities / UseCases / Ports / Infra)

⚠️ REGRA ABSOLUTA:
NÃO QUEBRE nada que já esteja funcionando.
Toda alteração deve ser incremental, compatível e retrocompatível.

---

CONTEXTO DA MISSÃO

O banco já funciona tecnicamente.
Agora precisamos implementar a **LÓGICA DE MONETIZAÇÃO OFICIAL DO Z-BANCK**, de forma:

- Centralizada
- Configurável
- Auditável
- Sem hardcode de taxas
- Compatível com PF e PJ
- Compatível com planos FREE / PRO / BUSINESS

Nenhuma taxa deve ser calculada “na mão” nos use cases.
Toda taxa deve passar pelo Pricing Engine.

---

OBJETIVO PRINCIPAL

Criar um **PRICING ENGINE** desacoplado que:
- Calcule fees por operação
- Suporte planos
- Suporte PF / PJ
- Suporte diferentes ativos (fiat/crypto)
- Gere taxas sempre menores ou iguais ao mercado (Binance/Bancos/DEX)

E integrar esse engine aos fluxos já existentes:
- Pix
- Crypto swap / pay
- Card JIT
- Invoice / QR híbrido
- Auto-conversão fiat → crypto

---

PARTE 1 — MODELAGEM DE DADOS (MIGRATIONS)

Criar as seguintes tabelas (sem remover nada existente):

1️⃣ plans
Campos:
- id (uuid)
- code (FREE, PRO, BUSINESS)
- description
- monthly_price_cents
- created_at

2️⃣ user_plans
- user_id
- plan_id
- valid_from
- valid_until
- created_at

3️⃣ pricing_rules
- id
- plan_id
- user_type (PF | PJ)
- operation_type (PIX_IN, PIX_OUT, PIX_TO_CRYPTO, CRYPTO_TO_PIX, SWAP, CARD_CRYPTO, INVOICE)
- asset (BRL, BTC, ETH, USDT, ANY)
- fee_type (PERCENTAGE | FIXED)
- fee_value (int64)
- min_fee (nullable)
- max_fee (nullable)
- created_at

⚠️ REGRAS:
- Nunca usar ENUM no banco
- Usar CHECK CONSTRAINT quando aplicável
- Migrations idempotentes

Forneça o SQL final.

---

PARTE 2 — ENTIDADES DE DOMÍNIO

Criar entidades em `internal/entity`:

- Plan
- UserPlan
- PricingRule

Criar Value Object:
- FeeResult { Amount, Fee, NetAmount }

---

PARTE 3 — PRICING ENGINE (CORE)

Criar módulo:
`internal/pricing/engine`

Responsabilidade:
Calcular taxas SEM acessar banco diretamente.

Interface:
```go
type PricingEngine interface {
  ResolveFee(ctx context.Context, input PricingInput) (FeeResult, error)
}
```
PricingInput deve conter:
UserID
UserType (PF/PJ)
Plan
OperationType
Asset
GrossAmount (int64)
⚠️ REGRAS:
Nunca calcular taxa nos handlers
Nunca usar valores mágicos
Sempre retornar Amount, Fee, NetAmount
PARTE 4 — USE CASE DE ORQUESTRAÇÃO
Criar UseCase: ResolvePricingUseCase
Responsável por:
Buscar plano do usuário
Buscar regras de pricing
Aplicar melhor regra (mais específica vence)
Retornar FeeResult
PARTE 5 — INTEGRAÇÃO COM FLUXOS EXISTENTES
Atualizar SEM QUEBRAR:
SendPixUseCase
CryptoPay / Swap
Card JIT Authorize
Invoice / QR
Auto-convert fiat → crypto
⚠️ REGRA CRÍTICA: Toda Transaction criada deve:
Registrar Amount (bruto)
Registrar Fee
Registrar NetAmount
Ledger continua append-only.
PARTE 6 — AUTO-CONVERSÃO (FEATURE CHECKMATE)
Implementar lógica configurável:
UserSetting:
auto_convert_enabled
auto_convert_asset (BTC, ETH, USDT)
auto_convert_min_amount
Fluxo:
Pix recebido
PricingEngine calcula taxa
ConversionEngine converte automaticamente
Ledger registra tudo auditável
PARTE 7 — AUDITORIA E TRANSPARÊNCIA
Toda taxa aplicada deve:
Gerar AuditLog
Ser rastreável por transaction_id
Nunca ser ocultada do backend (mesmo que UX oculte)
PARTE 8 — TESTES
Criar testes para:
PricingEngine
ResolvePricingUseCase
Comparação FREE vs PRO
PF vs PJ
```

---

## Requisitos derivados (criterios de implementacao)

### Requisitos tecnicos invariaveis
- Ledger append-only com `int64` em todas as operacoes monetarias.
- Idempotency-Key obrigatoria em rotas de escrita e controle por indices unicos.
- Unit of Work com locks pessimistas para evitar race conditions.
- Pricing Engine desacoplado, sem acesso direto a banco e sem valores magicos.
- FeeResult sempre retornando `Amount`, `Fee` e `NetAmount`.

### Requisitos de dados
- Tabelas `plans`, `user_plans`, `pricing_rules` sem ENUM, com CHECK constraints.
- Regras de pricing com seletor por plano, user_type e asset (ANY como fallback).
- Migrations idempotentes e compatibilidade retroativa.

### Requisitos de produto e monetizacao
- Planos FREE/PRO/BUSINESS e perfis PF/PJ.
- Fees centralizados e auditaveis, sempre via Pricing Engine.
- Taxas competitivas (nao acima do mercado).
- Integracao em Pix, Crypto, Card JIT, Invoice, Auto-conversao.

### Requisitos de auditoria e compliance
- AuditLog para toda taxa aplicada (rastreavel por `transaction_id`).
- Transparencia interna do backend, mesmo quando UX ocultar detalhes.
- Nao quebrar fluxos existentes e manter retrocompatibilidade.

### Requisitos de testes
- Testes unitarios para engine e use case de pricing.
- Comparativos FREE vs PRO e PF vs PJ.
