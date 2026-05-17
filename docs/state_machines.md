# State Machines — Z-FINANCE

Diagramas de estados para fluxos criticos. Use como referencia de negocio e para testes.

---

## Transacoes (ledger)

```mermaid
stateDiagram-v2
  [*] --> Created
  Created --> Confirmed: deposit_or_manual_confirm
  Created --> Hold: payment_trade_card
  Hold --> Confirmed: confirm
  Hold --> Rejected: reject
  Confirmed --> Reversed: reversal_transaction
  Rejected --> [*]
  Reversed --> [*]
```

Nota: reversao cria uma nova transacao no ledger; nao altera o evento original.

---

## Pix (send)

```mermaid
stateDiagram-v2
  [*] --> Created
  Created --> PendingPartner
  PendingPartner --> Confirmed
  PendingPartner --> Rejected
  Confirmed --> [*]
  Rejected --> [*]
```

---

## Payments (boletos)

```mermaid
stateDiagram-v2
  [*] --> Created
  Created --> PendingPartner
  PendingPartner --> Confirmed
  PendingPartner --> Rejected
  Confirmed --> [*]
  Rejected --> [*]
```

---

## Card JIT

```mermaid
stateDiagram-v2
  [*] --> Hold
  Hold --> Confirmed
  Hold --> Rejected
  Confirmed --> [*]
  Rejected --> [*]
```

---

## Invoices (hibrido Pix + USDT)

```mermaid
stateDiagram-v2
  [*] --> Created
  Created --> Paid: pay_pix_or_usdt
  Paid --> [*]
```

