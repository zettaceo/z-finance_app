/* ─────────────────────────────────────────────
   Z-FINANCE · Demo Mock Store
   All amounts in cents (BRL) unless stated
───────────────────────────────────────────── */

const uid = () => Math.random().toString(36).slice(2, 10)
const now = new Date()
const daysAgo = (d) => new Date(now - d * 86400000).toISOString()

export const DEFAULT_USER = {
  id:      '00000000-0000-0000-0000-000000000001',
  name:    'Rafael Mendonça',
  email:   'rafael@zetta.bank',
  type:    'PF',
  status:  'ACTIVE',
  plan:    'BUSINESS',
  uxMode:  'PRO',
  avatarInitials: 'RM',
}

export function buildMockStore() {
  return {
    user: { ...DEFAULT_USER },

    account: {
      id:       '00000000-0000-0000-0000-000000000002',
      balance:  2456710,   // R$ 24.567,10
      holds:    150000,    // R$ 1.500,00 bloqueado
    },

    investments: {
      cdb:   350000,   // R$ 3.500,00
      tesouro: 180000, // R$ 1.800,00
    },

    crypto: [
      {
        symbol:    'BTC',
        name:      'Bitcoin',
        amount:    0.5432,
        price:     345820.00,
        change24h: 3.2,
        history:   [310000,315000,308000,322000,335000,340000,345820],
      },
      {
        symbol:    'ETH',
        name:      'Ethereum',
        amount:    2.315,
        price:     18240.50,
        change24h: 1.5,
        history:   [16800,17200,16500,17800,18000,18100,18240],
      },
      {
        symbol:    'USDT',
        name:      'Tether',
        amount:    1250.00,
        price:     5.78,
        change24h: 0.1,
        history:   [5.75,5.76,5.77,5.76,5.78,5.78,5.78],
      },
    ],

    transactions: [
      { id: uid(), type:'PIX_OUT',     amount: -4590,   description: 'PicPay Tecnologia',    createdAt: daysAgo(0),   status:'CONFIRMED' },
      { id: uid(), type:'TRANSFER_IN', amount: 500000,  description: 'TED Recebida — Zetta', createdAt: daysAgo(1),   status:'CONFIRMED' },
      { id: uid(), type:'PAYMENT',     amount: -95000,  description: 'Conta de Luz — ENEL',  createdAt: daysAgo(1),   status:'CONFIRMED' },
      { id: uid(), type:'PIX_IN',      amount: 120000,  description: 'PIX — Ana Clara',      createdAt: daysAgo(2),   status:'CONFIRMED' },
      { id: uid(), type:'CARD_AUTH',   amount: -28500,  description: 'iFood Restaurante',    createdAt: daysAgo(2),   status:'CONFIRMED' },
      { id: uid(), type:'TRADE_BUY',   amount: -150000, description: 'Compra BTC',            createdAt: daysAgo(3),   status:'CONFIRMED' },
      { id: uid(), type:'PIX_OUT',     amount: -8200,   description: 'PIX — João Carvalho',  createdAt: daysAgo(4),   status:'CONFIRMED' },
      { id: uid(), type:'DEPOSIT',     amount: 1000000, description: 'Depósito Bancário',     createdAt: daysAgo(5),   status:'CONFIRMED' },
      { id: uid(), type:'PAYMENT',     amount: -42000,  description: 'Internet — Vivo Fibra', createdAt: daysAgo(6),  status:'CONFIRMED' },
      { id: uid(), type:'TRADE_SELL',  amount: 85000,   description: 'Venda ETH parcial',     createdAt: daysAgo(7),  status:'CONFIRMED' },
    ],

    // 7-day cash flow for chart
    cashFlow: (() => {
      const labels = ['Dom','Seg','Ter','Qua','Qui','Sex','Sáb']
      return labels.map((d, i) => ({
        day:      d,
        income:   [0, 500000, 120000, 0, 85000, 0, 1000000][i],
        expenses: [0, 95000,  28500,  150000, 8200, 42000, 4590][i],
      }))
    })(),

    kyc: {
      level:   'FULL',
      limits:  { daily: 10000000, monthly: 50000000 },
      used:    { daily: 1250000,  monthly: 8750000  },
    },

    pricingRate: { BTC: 345820, ETH: 18240, spread: 0.012 },

    invoices:        [],
    payments:        [],
    compliance:      [],
    preRegistrations:[],

    observability: {
      pendingWebhooks: 1,
      reconcilePending: 0,
      retries: 2,
      uptime: '99.97%',
      latencyMs: 42,
    },
  }
}

/* ─── Simulation Engine ─── */
export function runSimulation(store, action, formData = {}) {
  const s = JSON.parse(JSON.stringify(store)) // deep clone
  let message = ''
  let details = {}

  switch (action) {
    // ── PIX ──────────────────────────────────────────────
    case 'pix_send': {
      const amt = Math.round((Number(formData.amount) || 50) * 100)
      const fee = Math.round(amt * 0.002)
      s.account.balance -= amt + fee
      s.transactions.unshift({
        id: uid(), type: 'PIX_OUT',
        amount: -(amt + fee),
        description: `PIX → ${formData.receiver || 'Destinatário'}`,
        createdAt: new Date().toISOString(), status: 'CONFIRMED',
      })
      message = `PIX enviado: ${fmtBRL(amt)} (taxa: ${fmtBRL(fee)})`
      details = { id: `e2e_${uid()}`, status: 'CONFIRMED', net: fmtBRL(amt - fee) }
      break
    }
    case 'pix_receive': {
      const amt = Math.round((Number(formData.amount) || 200) * 100)
      s.account.balance += amt
      s.transactions.unshift({
        id: uid(), type: 'PIX_IN',
        amount: amt,
        description: `PIX ← ${formData.sender || 'Remetente'}`,
        createdAt: new Date().toISOString(), status: 'CONFIRMED',
      })
      message = `PIX recebido: ${fmtBRL(amt)}`
      details = { id: `e2e_${uid()}`, status: 'CONFIRMED' }
      break
    }
    case 'pix_crypto': {
      const usdtAmt = Number(formData.usdt || 100)
      const brlAmt  = Math.round(usdtAmt * s.pricingRate.ETH * 0.05)
      s.account.balance += brlAmt
      s.transactions.unshift({
        id: uid(), type: 'PIX_IN',
        amount: brlAmt,
        description: `PIX via Cripto (${usdtAmt} USDT)`,
        createdAt: new Date().toISOString(), status: 'CONFIRMED',
      })
      message = `PIX via cripto: recebido ${fmtBRL(brlAmt)} (${usdtAmt} USDT)`
      details = { usdt: usdtAmt, brl: fmtBRL(brlAmt), spread: '1.2%' }
      break
    }

    // ── Pagamentos ───────────────────────────────────────
    case 'payment_schedule': {
      const amt = Math.round((Number(formData.amount) || 120) * 100)
      s.payments.push({ id: uid(), amount: amt, barcode: formData.barcode || '00000.00000 00000.000000 00000.000000 0 00000000000000', status: 'PENDING', createdAt: new Date().toISOString() })
      message = `Pagamento agendado: ${fmtBRL(amt)}`
      details = { status: 'SCHEDULED', dueDate: formData.dueDate || 'Vencimento em 2 dias' }
      break
    }
    case 'payment_confirm': {
      const amt = 120000
      s.account.balance -= amt
      s.transactions.unshift({ id: uid(), type:'PAYMENT', amount: -amt, description:'Boleto confirmado', createdAt: new Date().toISOString(), status:'CONFIRMED' })
      message = 'Pagamento confirmado: R$ 1.200,00'
      details = { status: 'CONFIRMED', authCode: uid() }
      break
    }

    // ── Cobranças ────────────────────────────────────────
    case 'invoice_create': {
      const amt = Math.round((Number(formData.amount) || 500) * 100)
      const inv  = { id: uid(), amountBRL: amt, pixKey: `${uid()}@zetta.bank`, usdtAddr: `0x${uid()}${uid()}`, createdAt: new Date().toISOString(), status: 'PENDING' }
      s.invoices.push(inv)
      message = `Cobrança criada: ${fmtBRL(amt)}`
      details = { pixKey: inv.pixKey, usdt: `Endereço: ${inv.usdtAddr.slice(0,20)}...`, status: 'AGUARDANDO_PAGAMENTO' }
      break
    }
    case 'invoice_pay': {
      if (s.invoices.length === 0) { message = 'Nenhuma cobrança pendente'; break }
      const inv = s.invoices[s.invoices.length - 1]
      s.account.balance += inv.amountBRL
      inv.status = 'PAID'
      s.transactions.unshift({ id: uid(), type:'PIX_IN', amount: inv.amountBRL, description:`Pagamento de cobrança ${inv.id.slice(0,8)}`, createdAt: new Date().toISOString(), status:'CONFIRMED' })
      message = `Cobrança paga: ${fmtBRL(inv.amountBRL)}`
      details = { invoiceId: inv.id, method: 'PIX', status: 'PAID' }
      break
    }

    // ── Transferências ───────────────────────────────────
    case 'transfer_internal': {
      const amt = Math.round((Number(formData.amount) || 300) * 100)
      s.account.balance -= amt
      s.investments.cdb += amt
      s.transactions.unshift({ id: uid(), type:'TRANSFER_IN', amount: -amt, description:'Transferência → Conta Investimento', createdAt: new Date().toISOString(), status:'CONFIRMED' })
      message = `Transferência interna: ${fmtBRL(amt)} para Conta Investimento`
      details = { from: 'Conta Principal', to: 'Conta Investimento', amount: fmtBRL(amt) }
      break
    }
    case 'withdrawal': {
      const amt = Math.round((Number(formData.amount) || 100) * 100)
      s.account.balance -= amt
      s.transactions.unshift({ id: uid(), type:'WITHDRAWAL', amount: -amt, description:`Saque — ${formData.pix || 'Chave PIX'}`, createdAt: new Date().toISOString(), status:'CONFIRMED' })
      message = `Saque de ${fmtBRL(amt)} processado`
      details = { status: 'CONFIRMED', network: 'PIX', txId: uid() }
      break
    }

    // ── Cripto ───────────────────────────────────────────
    case 'crypto_swap': {
      const fromAmt = Number(formData.amount || 100)
      const fromCur = formData.from || 'BRL'
      const toCur   = formData.to   || 'BTC'
      if (fromCur === 'BRL') {
        const brlCents = Math.round(fromAmt * 100)
        const coin     = s.crypto.find(c => c.symbol === toCur)
        if (coin) {
          const received = fromAmt / coin.price
          coin.amount += received
          s.account.balance -= brlCents
          message = `Swap: ${fmtBRL(brlCents)} → ${received.toFixed(6)} ${toCur}`
          details = { sold: fmtBRL(brlCents), received: `${received.toFixed(6)} ${toCur}`, spread: '1.2%', rate: `R$ ${coin.price.toLocaleString('pt-BR')}` }
        }
      } else {
        const coin = s.crypto.find(c => c.symbol === fromCur)
        if (coin) {
          const brlReceived = fromAmt * coin.price * (1 - s.pricingRate.spread)
          coin.amount -= fromAmt
          s.account.balance += Math.round(brlReceived * 100)
          message = `Swap: ${fromAmt} ${fromCur} → ${fmtBRL(brlReceived * 100)}`
          details = { sold: `${fromAmt} ${fromCur}`, received: fmtBRL(brlReceived * 100), spread: '1.2%' }
        }
      }
      if (message) s.transactions.unshift({ id: uid(), type:'TRADE_BUY', amount: -Math.round(fromAmt * 100), description:`Swap ${fromCur} → ${toCur}`, createdAt: new Date().toISOString(), status:'CONFIRMED' })
      break
    }
    case 'pricing_quote': {
      const pair = formData.pair || 'BTC/BRL'
      const [base] = pair.split('/')
      const coin = s.crypto.find(c => c.symbol === base)
      const price = coin?.price || 345820
      message = `Cotação ${pair}: R$ ${price.toLocaleString('pt-BR')}`
      details = { pair, bid: fmtBRL(price * 100 * 0.999), ask: fmtBRL(price * 100 * 1.001), spread: '0.2%', ts: new Date().toISOString() }
      break
    }
    case 'crypto_liquidate': {
      const coin = s.crypto.find(c => c.symbol !== 'USDT')
      if (coin && coin.amount > 0) {
        const half    = coin.amount / 2
        const brl     = Math.round(half * coin.price * 100)
        coin.amount  -= half
        s.account.balance += brl
        message = `Liquidação: ${half.toFixed(4)} ${coin.symbol} → ${fmtBRL(brl)}`
        details = { sold: `${half.toFixed(4)} ${coin.symbol}`, received: fmtBRL(brl), status: 'CONFIRMED' }
        s.transactions.unshift({ id: uid(), type:'TRADE_SELL', amount: brl, description:`Liquidação ${coin.symbol}`, createdAt: new Date().toISOString(), status:'CONFIRMED' })
      } else {
        message = 'Sem posição disponível para liquidar'
      }
      break
    }

    // ── Card JIT ─────────────────────────────────────────
    case 'card_authorize': {
      const amt = Math.round((Number(formData.amount) || 250) * 100)
      s.account.holds = (s.account.holds || 0) + amt
      message = `Autorização JIT: ${fmtBRL(amt)} — ${formData.merchant || 'Merchant'}`
      details = { authCode: uid().toUpperCase(), status: 'HOLD', merchant: formData.merchant || 'Merchant', mcc: formData.mcc || '5411' }
      break
    }
    case 'card_confirm': {
      const holdAmt = s.account.holds || 0
      s.account.balance -= holdAmt
      s.account.holds    = 0
      s.transactions.unshift({ id: uid(), type:'CARD_AUTH', amount: -holdAmt, description:'Compra cartão JIT confirmada', createdAt: new Date().toISOString(), status:'CONFIRMED' })
      message = `Compra confirmada: ${fmtBRL(holdAmt)}`
      details = { status: 'SETTLED', amount: fmtBRL(holdAmt) }
      break
    }
    case 'card_reject': {
      s.account.holds = 0
      message = 'Autorização rejeitada — hold liberado'
      details = { status: 'REJECTED', reason: formData.reason || 'Fraude suspeita' }
      break
    }

    // ── KYC ─────────────────────────────────────────────
    case 'kyc_limits': {
      const k = s.kyc
      message = `KYC ${k.level}: diário ${fmtBRL(k.limits.daily)}, mensal ${fmtBRL(k.limits.monthly)}`
      details = { level: k.level, daily: fmtBRL(k.limits.daily), monthly: fmtBRL(k.limits.monthly), usedDaily: fmtBRL(k.used.daily) }
      break
    }
    case 'kyc_upgrade': {
      s.kyc.level = 'FULL'
      s.kyc.limits = { daily: 20000000, monthly: 100000000 }
      message = 'KYC atualizado para FULL — limites ampliados!'
      details = { newLevel: 'FULL', newDailyLimit: fmtBRL(20000000), providerRef: uid() }
      break
    }

    // ── Compliance ───────────────────────────────────────
    case 'compliance_case': {
      const c = { id: uid(), type: formData.type || 'SUSPICIOUS_ACTIVITY', status: 'OPEN', riskLevel: 'HIGH', title: formData.title || 'Atividade incomum detectada', createdAt: new Date().toISOString() }
      s.compliance.push(c)
      message = `Case aberto: ${c.title}`
      details = { caseId: c.id, riskLevel: 'HIGH', status: 'OPEN' }
      break
    }
    case 'compliance_event': {
      message = 'Evento de compliance registrado'
      details = { eventType: formData.eventType || 'MANUAL_REVIEW', payload: { analyst: 'Sistema Zetta', ts: new Date().toISOString() } }
      break
    }

    // ── Pré-cadastro ─────────────────────────────────────
    case 'pre_registration': {
      const pr = { id: uid(), email: formData.email || 'empresa@client.com', status: 'PENDING', createdAt: new Date().toISOString() }
      s.preRegistrations.push(pr)
      message = `Pré-cadastro iniciado para ${pr.email}`
      details = { id: pr.id, emailStatus: 'TOKEN_SENT', phoneStatus: 'PENDING', expiresAt: new Date(Date.now() + 3600000).toISOString() }
      break
    }

    // ── Admin ────────────────────────────────────────────
    case 'admin_plan_change': {
      const plan = formData.plan || 'BUSINESS'
      s.user.plan = plan
      message = `Plano alterado para ${plan}`
      details = { userId: s.user.id, newPlan: plan, effectiveAt: new Date().toISOString() }
      break
    }
    case 'admin_feature_toggle': {
      message = `Feature ${formData.feature || 'CRYPTO'} ${formData.enabled ? 'habilitada' : 'desabilitada'}`
      details = { feature: formData.feature || 'CRYPTO', enabled: formData.enabled ?? true, scope: 'USER_OVERRIDE' }
      break
    }
    case 'admin_limit_adjust': {
      s.kyc.limits.daily = Math.round(Number(formData.daily || 10000) * 100)
      message = `Limite diário ajustado para ${fmtBRL(s.kyc.limits.daily)}`
      details = { newDailyLimit: fmtBRL(s.kyc.limits.daily), approvedBy: 'Admin', auditId: uid() }
      break
    }

    // ── Observabilidade ───────────────────────────────────
    case 'observability_summary': {
      message = `Observabilidade: ${s.observability.pendingWebhooks} webhooks pendentes, latência ${s.observability.latencyMs}ms`
      details = { ...s.observability }
      break
    }
    case 'audit_archive': {
      message = 'Arquivo de auditoria gerado e exportado'
      details = { records: s.transactions.length, exportId: uid(), format: 'JSON', ts: new Date().toISOString() }
      break
    }

    default:
      message = `Simulação "${action}" executada`
      details = { action, ts: new Date().toISOString() }
  }

  return { store: s, message, details }
}

/* ─── Formatters ─── */
export function fmtBRL(cents) {
  return (Number(cents || 0) / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

export function fmtBRLFull(cents) {
  return (Number(cents || 0) / 100).toLocaleString('pt-BR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export function txTypeLabel(type) {
  const map = {
    PIX_OUT:     'PIX Enviado',
    PIX_IN:      'PIX Recebido',
    TRANSFER_IN: 'Transferência',
    PAYMENT:     'Pagamento',
    CARD_AUTH:   'Cartão JIT',
    TRADE_BUY:   'Compra Cripto',
    TRADE_SELL:  'Venda Cripto',
    DEPOSIT:     'Depósito',
    WITHDRAWAL:  'Saque',
    REVERSAL:    'Estorno',
  }
  return map[type] || type
}

export function txDirection(type) {
  const outs = ['PIX_OUT','PAYMENT','CARD_AUTH','TRADE_BUY','WITHDRAWAL']
  const neutral = ['REVERSAL']
  if (neutral.includes(type)) return 'neutral'
  return outs.includes(type) ? 'out' : 'in'
}

export function relativeTime(iso) {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1)   return 'agora'
  if (mins < 60)  return `${mins}min atrás`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24)   return `${hrs}h atrás`
  const days = Math.floor(hrs / 24)
  return `${days}d atrás`
}

export function cryptoBalanceBRL(cryptoList) {
  return cryptoList.reduce((acc, c) => acc + c.amount * c.price * 100, 0)
}

export function totalBalance(store) {
  return store.account.balance +
    store.investments.cdb + store.investments.tesouro +
    cryptoBalanceBRL(store.crypto)
}
