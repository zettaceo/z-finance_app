# Z-Finance API

Motor de pagamentos Fiat da plataforma **Zetta**. Responsável por contas, ledger, PIX, pagamentos, cartões JIT, câmbio e compliance.

Faz parte do ecossistema modular Zetta — veja [ECOSYSTEM.md](./ECOSYSTEM.md) para a arquitetura completa.

---

## Stack

- **Go 1.22** · PostgreSQL 16 · Docker
- Clean Architecture (entity → usecase → adapter → infra)
- Ledger append-only · Unit of Work · Locks pessimistas
- OpenTelemetry OTLP · Webhooks HMAC

---

## Ecossistema Zetta

```
┌─────────────────────────────────────────────────────┐
│                  ZETTA PLATFORM                     │
├──────────┬──────────┬──────────┬──────────┬─────────┤
│z-finance │  z-swap  │  z-pay   │  z-pad   │obelisk-z│
│  (fiat)  │ (crypto) │(gateway) │  (card)  │(wallet) │
└────┬─────┴────┬─────┴────┬─────┴────┬─────┴────┬────┘
     │          │          │          │          │
     └──────────┴──────────┴──────────┴──────────┘
                      Zion AI (cross-cutting)
```

Z-Finance é o **núcleo fiat** — os outros produtos integram via as interfaces definidas em `internal/core/ports/`.

---

## Rodando

### Com Docker

```bash
docker compose up --build
# API em http://localhost:8080
```

### Com Neon (PostgreSQL na nuvem)

```bash
export DB_URL="postgresql://user:pass@host/dbname?sslmode=require"
go run ./cmd/main.go
```

---

## Variáveis de Ambiente

| Variável | Default | Descrição |
|---|---|---|
| `DB_URL` | — | PostgreSQL connection string |
| `HTTP_PORT` | `8080` | Porta HTTP |
| `WEBHOOK_SECRET` | — | HMAC secret para webhooks |
| `CORS_ORIGINS` | — | Origens permitidas (CSV) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | — | Collector OTLP |
| `Z_SWAP_URL` | — | URL do serviço z-swap |
| `OBELISK_Z_URL` | — | URL do serviço obelisk-z |
| `Z_PAY_URL` | — | URL do serviço z-pay |
| `ZION_AI_URL` | — | URL do serviço Zion AI |

---

## Seed Padrão

Ao iniciar sem dados, o sistema cria automaticamente:

- **Usuário admin**: `00000000-0000-0000-0000-000000000001`
- **Conta BRL**: `00000000-0000-0000-0000-000000000001`
- **KYC**: FULL/VERIFIED
- **Planos**: FREE / PRO / BUSINESS com pricing e features

---

## Endpoints Principais

### Core
| Método | Rota | Descrição |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/auth/me` | Capability snapshot (plan, features, limites) |
| POST | `/auth/login` | Login (email + senha) |
| POST | `/auth/refresh` | Refresh token |

### Contas e Ledger
| Método | Rota | Descrição |
|---|---|---|
| POST | `/transactions` | Criar transação *(Idempotency-Key)* |
| GET | `/transactions` | Listar transações |
| GET | `/accounts/{id}/balance` | Saldo calculado |

### PIX
| Método | Rota | Descrição |
|---|---|---|
| POST | `/pix/send` | Enviar PIX *(Idempotency-Key)* |
| POST | `/webhooks/pix/receive` | Receber PIX (HMAC) |

### Cripto / Swap
| Método | Rota | Descrição |
|---|---|---|
| POST | `/crypto/swap` | Swap fiat ↔ cripto |
| POST | `/crypto/pay` | Pagar com cripto |

### Cartão JIT
| Método | Rota | Descrição |
|---|---|---|
| POST | `/card/authorize` | Autorização JIT |

### Admin
| Método | Rota | Descrição |
|---|---|---|
| GET | `/admin/reconcile/summary` | Reconciliação |
| GET | `/admin/observability/summary` | Observabilidade |
| GET | `/admin/alerts/check` | Alertas de threshold |

> Rotas `/admin` exigem role `ADMIN`. Veja a lista completa em [ARCHITECTURE.md](./ARCHITECTURE.md).

---

## Documentação

| Arquivo | Conteúdo |
|---|---|
| [ECOSYSTEM.md](./ECOSYSTEM.md) | Arquitetura do ecossistema Zetta e contratos de integração |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Arquitetura técnica detalhada |
| [docs/PROJECT_BIBLE.md](./docs/PROJECT_BIBLE.md) | Visão e princípios do produto |
| [docs/ROADMAP.md](./docs/ROADMAP.md) | Roadmap |
| [docs/AML_POLICY.md](./docs/AML_POLICY.md) | Política AML |
| [docs/KYC_KYB_PROCEDURES.md](./docs/KYC_KYB_PROCEDURES.md) | Procedimentos KYC/KYB |
| [docs/RBAC_MATRIX.md](./docs/RBAC_MATRIX.md) | Matriz de permissões |
| [docs/TECHNICAL_FLOWS.md](./docs/TECHNICAL_FLOWS.md) | Fluxos técnicos detalhados |

---

## Princípios

- Valores monetários em `int64` (centavos) — sem `float`
- Ledger append-only — saldo calculado sobre eventos `CONFIRMED`
- `Idempotency-Key` obrigatória em todas as escritas
- ACID + locks pessimistas via Unit of Work
- Providers externos nunca alteram o ledger diretamente
- Falhas externas → `502` com `provider` + `operation` no body
