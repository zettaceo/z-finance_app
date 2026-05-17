# Ecossistema Zetta — Arquitetura de Integração

## Visão Geral

O ecossistema Zetta é composto por serviços independentes que se integram via contratos bem definidos (interfaces Go em `internal/core/ports/`). Nenhum serviço acessa o banco de dados de outro — toda comunicação é via API HTTP autenticada.

```
                        ┌─────────────────────────────────────┐
                        │           ZETTA PLATFORM            │
                        └─────────────────────────────────────┘
                                          │
               ┌──────────────────────────┼──────────────────────────┐
               │                          │                          │
        ┌──────▼──────┐           ┌───────▼──────┐          ┌───────▼──────┐
        │  z-finance  │◄─────────►│   z-swap     │          │    z-pay     │
        │  (fiat API) │           │ (crypto DEX) │          │  (gateway)   │
        │  :8080      │           │  :8081       │          │  :8090       │
        └──────┬──────┘           └──────────────┘          └──────────────┘
               │
        ┌──────┴───────────────────────────────────┐
        │                                          │
 ┌──────▼──────┐                          ┌────────▼─────┐
 │  obelisk-z  │                          │    z-pad     │
 │  (wallet)   │                          │   (cards)    │
 │  :8092      │                          │   :8093      │
 └─────────────┘                          └──────────────┘

                    ┌──────────────────────────┐
                    │        Zion AI           │
                    │  (assistente financeiro) │
                    │        :8094             │
                    └──────────────────────────┘
```

---

## Serviços

### Z-Finance (este repositório)
- **Papel**: motor fiat — ledger, contas, PIX, pagamentos, KYC, compliance, cartão JIT
- **Porta**: 8080
- **Banco**: `zfinance_core` (PostgreSQL)
- **Integra com**: z-swap (câmbio), obelisk-z (custodia), z-pay (PIX partner), Zion AI

### Z-Swap
- **Papel**: motor cripto — cotações, swaps, DEX, liquidez
- **Porta**: 8081
- **Banco**: `zswap_core` (PostgreSQL)
- **Usa Z-Finance para**: operações fiat nos fluxos cripto→fiat

### Z-Pay
- **Papel**: gateway de pagamentos para merchants
- **Porta**: 8090
- **Banco**: `zpay_core` (PostgreSQL)
- **Usa Z-Finance para**: liquidação fiat, PIX, boleto

### Z-Pad
- **Papel**: gestão de cartões (emissão, limites, controles)
- **Porta**: 8093
- **Banco**: `zpad_core` (PostgreSQL)
- **Usa Z-Finance para**: autorização JIT Card

### Obelisk-Z
- **Papel**: carteira self-custodial (non-custodial wallet)
- **Porta**: 8092
- **Banco**: nenhum (chaves locais / HSM)
- **Usa Z-Finance para**: operações fiat on/off-ramp

### Zion AI
- **Papel**: assistente financeiro cross-cutting
- **Porta**: 8094
- **Usa todos os serviços para**: análise e propostas de intents

---

## Contratos de Integração

Todos os contratos estão definidos em `internal/core/ports/`. Z-Finance chama os outros serviços através dessas interfaces — e os outros serviços chamam Z-Finance via HTTP diretamente.

---

### 1. ExchangeGateway → Z-Swap

**Arquivo**: `internal/core/ports/exchange.go`

```go
type ExchangeGateway interface {
    Quote(ctx context.Context, asset string) (ExchangeQuote, error)
    Execute(ctx context.Context, asset string, amount int64, side string) (string, error)
}

type ExchangeQuote struct {
    Asset           string
    PriceInBRLCents int64
}
```

**Implementação atual**: `internal/infra/crypto/mock_exchange.go` (mock)

**Implementação Z-Swap**: criar `internal/infra/zswap/exchange_gateway.go` que chama `GET /api/v1/quote?asset=BTC` e `POST /api/v1/execute` no serviço z-swap.

**Variável de ambiente**: `Z_SWAP_URL`

---

### 2. CustodyGateway → Obelisk-Z

**Arquivo**: `internal/core/ports/custody.go`

```go
type CustodyGateway interface {
    CreateDepositAddress(ctx context.Context, userID, asset, network string) (entity.CustodyAddress, error)
    SendTransfer(ctx context.Context, transfer entity.CustodyTransfer) (entity.CustodyTransfer, error)
    GetTransferStatus(ctx context.Context, providerID string) (entity.CustodyTransferStatus, error)
}
```

**Implementação atual**: `internal/infra/custody/mock_gateway.go` (mock)

**Implementação Obelisk-Z**: criar `internal/infra/obeliskz/custody_gateway.go` que chama os endpoints de custódia do Obelisk-Z.

**Variável de ambiente**: `OBELISK_Z_URL`

---

### 3. PixPartnerClient → Z-Pay

**Arquivo**: `internal/core/ports/pix_partner.go`

```go
type PixPartnerClient interface {
    Send(ctx context.Context, transfer *entity.PixTransfer) error
}
```

**Propósito**: Z-Finance envia PIX via Z-Pay como parceiro SPI (Sistema de Pagamentos Instantâneos).

**Implementação Z-Pay**: criar `internal/infra/zpay/pix_partner.go` que chama `POST /internal/pix/send` no z-pay com autenticação service-to-service.

**Variável de ambiente**: `Z_PAY_URL`

---

### 4. AIProvider → Zion AI

**Arquivo**: `internal/core/ports/providers.go`

```go
type AIProvider interface {
    Analyze(ctx context.Context, userID string, input map[string]any) (map[string]any, error)
    ProposeIntent(ctx context.Context, userID string, input map[string]any) (ExecutionIntent, error)
}

// Invariantes de segurança:
// - Zion AI apenas propõe ExecutionIntents — NUNCA executa saldo
// - Todo intent requer confirmação explícita do usuário
// - Z-Finance executa o intent via use case normal após confirmação
type ExecutionIntent interface {
    ID() string
    UserID() string
    Kind() string
    Payload() map[string]any
}
```

**Variável de ambiente**: `ZION_AI_URL`

---

### 5. TransferGateway → Obelisk-Z / Rede externa

**Arquivo**: `internal/core/ports/transfer.go`

```go
type TransferGateway interface {
    Send(ctx context.Context, network, asset, address string, amount int64) (string, error)
}
```

**Propósito**: envio de cripto para endereços externos (on-chain withdrawals).

---

## Como Conectar um Novo Serviço

Para substituir um mock por uma implementação real:

1. Criar arquivo em `internal/infra/<servico>/` implementando a interface do port
2. Adicionar variável de ambiente `<SERVICO>_URL` ao `configs/config.go`
3. Em `cmd/main.go`, instanciar a implementação real e injetar via construtor do handler
4. Adicionar health check do serviço em `GET /health`

**Exemplo** — conectar Z-Swap ao ExchangeGateway:

```go
// cmd/main.go (trecho)
var exchangeGateway ports.ExchangeGateway
if cfg.ZSwapURL != "" {
    exchangeGateway = zswap.NewExchangeGateway(cfg.ZSwapURL, cfg.ZSwapAPIKey)
} else {
    exchangeGateway = crypto.NewMockExchangeGateway()
    log.Println("WARN: usando mock exchange — defina Z_SWAP_URL para produção")
}
```

---

## Comunicação Entre Serviços

| Direção | Protocolo | Auth |
|---|---|---|
| Z-Finance → Z-Swap | HTTP REST | API Key (`X-Service-Key`) |
| Z-Finance → Obelisk-Z | HTTP REST | API Key + mTLS (produção) |
| Z-Finance → Z-Pay | HTTP REST | API Key (`X-Service-Key`) |
| Z-Finance → Zion AI | HTTP REST | API Key |
| Z-Pay → Z-Finance | HTTP REST | API Key |
| Z-Swap → Z-Finance | HTTP REST | API Key |
| Z-Pad → Z-Finance | HTTP Webhook | HMAC SHA256 |

**Regra de ouro**: nenhum serviço acessa o banco de dados de outro. Todo acesso é via API.

---

## Fluxos Cross-Service

### Swap Fiat → Cripto (usuário no Z-Finance)
```
User → Z-Finance POST /crypto/swap
  Z-Finance → ExchangeGateway.Quote("BTC")    [z-swap]
  Z-Finance → ExchangeGateway.Execute(...)     [z-swap]
  Z-Finance → ledger: debita BRL, credita BTC  [interno]
  Z-Finance → CustodyGateway.CreateDepositAddress [obelisk-z]
```

### PIX Enviado via Z-Pay
```
User → Z-Finance POST /pix/send
  Z-Finance → pricing engine [interno]
  Z-Finance → velocity check [interno]
  Z-Finance → PixPartnerClient.Send(...) [z-pay]
  Z-Pay → SPI / banco destino [externo]
  Z-Pay → webhook Z-Finance /webhooks/pix/receive [HMAC]
  Z-Finance → ledger confirma [interno]
```

### Autorização de Cartão JIT via Z-Pad
```
Rede bandeira → Z-Pad POST /authorize
  Z-Pad → Z-Finance POST /card/authorize [Idempotency-Key]
  Z-Finance → pricing engine [interno]
  Z-Finance → saldo disponível? [interno]
  Z-Finance → ledger: CARD_AUTH PENDING [interno]
  Z-Finance → 200 OK → Z-Pad → bandeira
  [liquidação] → webhook Z-Finance /webhooks/card/confirm
```

### Zion AI propondo intent
```
User → Zion AI (frontend)
  Zion → AIProvider.Analyze(userID, input) [z-finance]
  Z-Finance → analisa contexto financeiro [interno]
  Zion → AIProvider.ProposeIntent(...) [z-finance]
  Z-Finance → retorna ExecutionIntent (sem executar)
  User confirma intent → Z-Finance executa use case normal
```

---

## Rede Docker (desenvolvimento)

```yaml
# docker-compose.yml (raiz do monorepo ou por serviço)
networks:
  zetta-platform:
    driver: bridge

# Cada serviço referencia:
services:
  z-finance:
    environment:
      Z_SWAP_URL: http://z-swap:8081
      OBELISK_Z_URL: http://obelisk-z:8092
      Z_PAY_URL: http://z-pay:8090
      ZION_AI_URL: http://zion-ai:8094
    networks:
      - zetta-platform
```

---

## Estado Atual dos Integrações

| Interface | Status | Implementação |
|---|---|---|
| ExchangeGateway → Z-Swap | Mock | `internal/infra/crypto/mock_exchange.go` |
| CustodyGateway → Obelisk-Z | Mock | `internal/infra/custody/mock_gateway.go` |
| PixPartnerClient → Z-Pay | Mock (inline) | `internal/usecase/send_pix.go` |
| AIProvider → Zion AI | Não implementado | — |
| TransferGateway | Mock | `internal/infra/crypto/mock_transfer.go` |

**Para produção**: substituir cada mock pela implementação real do serviço correspondente, seguindo o padrão da seção "Como Conectar um Novo Serviço".
