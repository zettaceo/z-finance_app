/* ═══════════════════════════════════════════════════════
   Z-FINANCE · Mock Store & Simulation Engine v3
   Cobre 100% dos endpoints do backend
═══════════════════════════════════════════════════════ */

const uid = () => Math.random().toString(36).slice(2,10).toUpperCase()
const now = new Date()
const daysAgo = (d,h=0) => new Date(now - d*86400000 - h*3600000).toISOString()

/* ── Formatters (export first, used everywhere) ─────── */
export const fmtBRL = (cents) =>
  (Number(cents||0)/100).toLocaleString('pt-BR',{style:'currency',currency:'BRL'})

export const fmtNum = (n, dec=2) =>
  Number(n||0).toLocaleString('pt-BR',{minimumFractionDigits:dec,maximumFractionDigits:dec})

export const relTime = (iso) => {
  if (!iso) return '—'
  const d = Date.now() - new Date(iso).getTime()
  const m = Math.floor(d/60000)
  if (m < 1)  return 'agora'
  if (m < 60) return `${m}min`
  const h = Math.floor(m/60)
  if (h < 24) return `${h}h`
  return `${Math.floor(h/24)}d`
}

export const txLabel = (type) => ({
  PIX_OUT:'PIX Enviado', PIX_IN:'PIX Recebido',
  TRANSFER_IN:'Transferência', TRANSFER_OUT:'Transferência',
  PAYMENT:'Pagamento', CARD_AUTH:'Cartão JIT',
  TRADE_BUY:'Compra Cripto', TRADE_SELL:'Venda Cripto',
  DEPOSIT:'Depósito', WITHDRAWAL:'Saque', REVERSAL:'Estorno',
}[type] || type)

export const txDir = (type) => {
  if (['PIX_IN','TRANSFER_IN','DEPOSIT','TRADE_SELL'].includes(type)) return 'in'
  if (['PIX_OUT','PAYMENT','CARD_AUTH','TRADE_BUY','WITHDRAWAL','TRANSFER_OUT'].includes(type)) return 'out'
  return 'neu'
}

/* ── Default Store ──────────────────────────────────── */
export function buildStore() {
  return {
    user: {
      id:       '00000000-0000-0000-0000-000000000001',
      name:     'Rafael Mendonça',
      email:    'rafael@zetta.bank',
      type:     'PJ',
      status:   'ACTIVE',
      plan:     'BUSINESS',
      uxMode:   'PRO',
      initials: 'RM',
      kycLevel: 'FULL',
    },

    accounts: {
      main: {
        id:       '00000000-0000-0000-0000-000000000010',
        label:    'Conta Principal',
        currency: 'BRL',
        balance:  2456710,  // R$ 24.567,10
        holds:    0,
        iban:     'ZF •••• •••• 8421',
      },
      usd: {
        id:       '00000000-0000-0000-0000-000000000011',
        label:    'Conta USD',
        currency: 'USD',
        balance:  150000,   // US$ 1.500,00
        holds:    0,
        iban:     'ZF •••• •••• 2204',
      },
      invest: {
        id:       '00000000-0000-0000-0000-000000000012',
        label:    'Investimentos',
        currency: 'BRL',
        balance:  530000,   // R$ 5.300,00
        holds:    0,
        iban:     'ZF •••• •••• 3309',
      },
    },

    crypto: [
      { symbol:'BTC', name:'Bitcoin',  amount:0.5432, price:345820, change24h:3.2,  change7d:8.1,  history:[310000,315000,308000,322000,335000,340000,345820], color:'#F7931A', letter:'₿' },
      { symbol:'ETH', name:'Ethereum', amount:2.315,  price:18240,  change24h:1.5,  change7d:4.2,  history:[16800,17200,16500,17800,18000,18100,18240],         color:'#627EEA', letter:'Ξ' },
      { symbol:'SOL', name:'Solana',   amount:12.5,   price:890,    change24h:-0.8, change7d:12.4, history:[810,840,795,870,885,878,890],                        color:'#9945FF', letter:'◎' },
      { symbol:'USDT',name:'Tether',   amount:850,    price:578,    change24h:0.1,  change7d:0.0,  history:[576,577,577,577,578,578,578],                        color:'#26A17B', letter:'₮' },
    ],

    investments: {
      cdb:     350000,  // CDB Banco Zetta
      tesouro: 180000,  // Tesouro SELIC
      lci:     0,
    },

    market: {
      USD: { rate:578, change:0.3 },
      EUR: { rate:624, change:-0.2 },
      GBP: { rate:731, change:0.1 },
    },

    transactions: [
      { id:uid(), type:'PIX_OUT',     amount:-4590,   desc:'PicPay Tecnologia',      at:daysAgo(0,2),  status:'CONFIRMED', category:'Serviços' },
      { id:uid(), type:'TRANSFER_IN', amount:500000,  desc:'TED Recebida — ZettaCo', at:daysAgo(1,1),  status:'CONFIRMED', category:'Transferência' },
      { id:uid(), type:'PAYMENT',     amount:-95000,  desc:'ENEL — Conta de Luz',    at:daysAgo(1,4),  status:'CONFIRMED', category:'Utilidades' },
      { id:uid(), type:'PIX_IN',      amount:120000,  desc:'PIX — Ana Clara Lima',   at:daysAgo(2,3),  status:'CONFIRMED', category:'PIX' },
      { id:uid(), type:'CARD_AUTH',   amount:-28500,  desc:'iFood Restaurantes',     at:daysAgo(2,6),  status:'CONFIRMED', category:'Alimentação' },
      { id:uid(), type:'TRADE_BUY',   amount:-150000, desc:'Compra 0.43 BTC',        at:daysAgo(3,0),  status:'CONFIRMED', category:'Cripto' },
      { id:uid(), type:'PIX_OUT',     amount:-8200,   desc:'PIX — João Carvalho',    at:daysAgo(4,5),  status:'CONFIRMED', category:'PIX' },
      { id:uid(), type:'DEPOSIT',     amount:1000000, desc:'Depósito TED Externo',   at:daysAgo(5,8),  status:'CONFIRMED', category:'Depósito' },
      { id:uid(), type:'PAYMENT',     amount:-42000,  desc:'Vivo Fibra — Internet',  at:daysAgo(6,2),  status:'CONFIRMED', category:'Utilidades' },
      { id:uid(), type:'TRADE_SELL',  amount:85000,   desc:'Venda 0.12 ETH',         at:daysAgo(7,9),  status:'CONFIRMED', category:'Cripto' },
      { id:uid(), type:'PIX_IN',      amount:250000,  desc:'PIX — Mariana Costa',    at:daysAgo(8,3),  status:'CONFIRMED', category:'PIX' },
      { id:uid(), type:'WITHDRAWAL',  amount:-50000,  desc:'Saque PIX — Chave CPF',  at:daysAgo(9,1),  status:'CONFIRMED', category:'Saque' },
    ],

    cashFlow: [
      { day:'Dom', income:0,       expenses:0 },
      { day:'Seg', income:500000,  expenses:95000 },
      { day:'Ter', income:120000,  expenses:28500 },
      { day:'Qua', income:0,       expenses:150000 },
      { day:'Qui', income:85000,   expenses:8200 },
      { day:'Sex', income:0,       expenses:42000 },
      { day:'Sáb', income:1000000, expenses:4590 },
    ],

    kyc: {
      level:   'FULL',
      status:  'VERIFIED',
      limits:  { daily:10000000, monthly:50000000 },
      used:    { daily:1250000,  monthly:8750000 },
      providerRef: `KYC-${uid()}`,
    },

    pricingRate: { BTC:345820, ETH:18240, SOL:890, USDT:578, spread:0.012 },

    card: {
      number:    '•••• •••• •••• 8421',
      holder:    'RAFAEL MENDONCA',
      exp:       '12/27',
      cvv:       '•••',
      frozen:    false,
      limit:     500000,
      limitUsed: 28500,
    },

    invoices:         [],
    payments:         [],
    pixKeys:          [{ id:uid(), type:'EMAIL', key:'rafael@zetta.bank', createdAt:daysAgo(30) }],
    compliance:       { casesOpen:2, casesClosed:18, casesEscalated:1, events:[] },
    preRegistrations: [
      { id:uid(), name:'Ana Costa',       email:'ana@startup.com',    plan:'BUSINESS',     status:'pending',  createdAt:daysAgo(3) },
      { id:uid(), name:'Dubai Corp LLC',  email:'ops@dubaicorp.ae',   plan:'INSTITUTIONAL',status:'verified', createdAt:daysAgo(7) },
      { id:uid(), name:'João Freitas',    email:'joao@gmail.com',     plan:'RETAIL',       status:'pending',  createdAt:daysAgo(1) },
    ],
    auditLogs: [
      { id:uid(), action:'AUTH_LOGIN_SUCCESS', entityType:'USER', details:{ ip:'177.10.20.1' }, ts:daysAgo(0,1) },
      { id:uid(), action:'PIX_SEND',           entityType:'TRANSACTION', details:{ amount:4590 },  ts:daysAgo(0,2) },
      { id:uid(), action:'CARD_AUTHORIZE',     entityType:'CARD',        details:{ merchant:'iFood' }, ts:daysAgo(0,3) },
      { id:uid(), action:'KYC_UPGRADE',        entityType:'USER',        details:{ level:'FULL' }, ts:daysAgo(1) },
      { id:uid(), action:'COMPLIANCE_CASE',    entityType:'COMPLIANCE',  details:{ reason:'aml' }, ts:daysAgo(2) },
    ],

    observability: {
      uptime:          '99.98%',
      latencyP99:      42,
      pendingWebhooks: 1,
      reconcilePending:3,
      retries:         2,
      traceId:         uid(),
    },

    /* ── Reconciliação ──────────────────────────────── */
    reconcile: {
      summary: { pix:2, payment:1, card:0, webhook:1 },
      pending: [
        { id:uid(), type:'pix',     amount:45000,  desc:'PIX pendente (45min)',  createdAt:daysAgo(0,1),  status:'PENDING' },
        { id:uid(), type:'payment', amount:120000, desc:'Boleto pendente (2h)',  createdAt:daysAgo(0,2),  status:'PENDING' },
        { id:uid(), type:'pix',     amount:8800,   desc:'PIX pendente (3h)',     createdAt:daysAgo(0,3),  status:'PENDING' },
        { id:uid(), type:'webhook', amount:0,      desc:'Webhook sem entrega',   createdAt:daysAgo(0,5),  status:'DEAD' },
      ],
    },

    /* ── Alertas ────────────────────────────────────── */
    alerts: {
      thresholds: { pixPending:10, paymentPending:5, cardPending:10, txHold:30, webhookDead:5 },
      last: [
        { type:'PIX_PENDING',        value:2, threshold:10, status:'OK',   checkedAt:daysAgo(0,0.5) },
        { type:'PAYMENT_PENDING',    value:1, threshold:5,  status:'OK',   checkedAt:daysAgo(0,0.5) },
        { type:'CARD_PENDING',       value:0, threshold:10, status:'OK',   checkedAt:daysAgo(0,0.5) },
        { type:'WEBHOOK_RETRY_DEAD', value:1, threshold:5,  status:'OK',   checkedAt:daysAgo(0,0.5) },
        { type:'TX_HOLD',            value:0, threshold:30, status:'OK',   checkedAt:daysAgo(0,0.5) },
      ],
    },

    /* ── RBAC / Roles ───────────────────────────────── */
    roles: [
      { id:uid(), name:'ADMIN',       description:'Acesso total ao sistema',             usersCount:2, createdAt:daysAgo(180) },
      { id:uid(), name:'COMPLIANCE',  description:'Análise de compliance e AML/KYT',     usersCount:3, createdAt:daysAgo(180) },
      { id:uid(), name:'TRADER',      description:'Operações de trading e cripto',        usersCount:5, createdAt:daysAgo(90) },
      { id:uid(), name:'SUPPORT',     description:'Suporte ao cliente — leitura',         usersCount:8, createdAt:daysAgo(60) },
      { id:uid(), name:'RISK',        description:'Gestão de risco e limites',            usersCount:2, createdAt:daysAgo(45) },
    ],
    userRoles: [
      { userId:'USR-001', userName:'Rafael Mendonça',  roleName:'ADMIN',      assignedAt:daysAgo(90),  assignedBy:'SYSTEM' },
      { userId:'USR-002', userName:'Ana Costa',         roleName:'COMPLIANCE', assignedAt:daysAgo(60),  assignedBy:'rafael@zetta.bank' },
      { userId:'USR-003', userName:'Lucas Almeida',     roleName:'TRADER',     assignedAt:daysAgo(30),  assignedBy:'rafael@zetta.bank' },
    ],
    separationRules: [
      { id:uid(), roleA:'ADMIN',      roleB:'TRADER',     reason:'Segregação de funções regulatória', createdAt:daysAgo(180) },
      { id:uid(), roleA:'COMPLIANCE', roleB:'TRADER',     reason:'Independência de compliance',       createdAt:daysAgo(90) },
    ],

    /* ── Precificação avançada ──────────────────────── */
    pricingVersions: [
      { id:uid(), version:'v3.0', status:'ACTIVE',   notes:'Tarifas Dubai-ready',     createdAt:daysAgo(30),  activatedAt:daysAgo(30) },
      { id:uid(), version:'v2.5', status:'ARCHIVED',  notes:'Ajuste de spread cripto', createdAt:daysAgo(90),  activatedAt:daysAgo(90) },
      { id:uid(), version:'v2.0', status:'ARCHIVED',  notes:'Versão de lançamento',   createdAt:daysAgo(180), activatedAt:daysAgo(180) },
    ],
    pricingCampaigns: [
      { id:uid(), name:'Lançamento Dubai Q1',  status:'ACTIVE', discount:15, startDate:'2026-01-01', endDate:'2026-03-31', planCode:'INSTITUTIONAL', usageCount:4  },
      { id:uid(), name:'Black Friday 2026',    status:'DRAFT',  discount:20, startDate:'2026-11-25', endDate:'2026-11-30', planCode:'BUSINESS',     usageCount:0  },
      { id:uid(), name:'Early Adopter 2025',   status:'EXPIRED',discount:30, startDate:'2025-01-01', endDate:'2025-12-31', planCode:'RETAIL',       usageCount:42 },
    ],
    pricingRules: [
      { id:uid(), planCode:'RETAIL',       userType:'PF', featureCode:'PIX_SEND',    feeType:'PERCENT', feeValue:0 },
      { id:uid(), planCode:'BUSINESS',     userType:'PJ', featureCode:'PIX_SEND',    feeType:'FLAT',    feeValue:0 },
      { id:uid(), planCode:'BUSINESS',     userType:'PJ', featureCode:'CRYPTO_SWAP', feeType:'PERCENT', feeValue:0.8 },
      { id:uid(), planCode:'INSTITUTIONAL',userType:'PJ', featureCode:'CRYPTO_SWAP', feeType:'PERCENT', feeValue:0.5 },
      { id:uid(), planCode:'INSTITUTIONAL',userType:'PJ', featureCode:'PIX_SEND',    feeType:'FLAT',    feeValue:0 },
      { id:uid(), planCode:'RETAIL',       userType:'PF', featureCode:'CARD_JIT',    feeType:'PERCENT', feeValue:1.5 },
    ],

    /* ── Perfil regulatório ─────────────────────────── */
    regulatoryProfile: {
      userId:       '00000000-0000-0000-0000-000000000001',
      vasp:         true,
      fatfRisk:     'LOW',
      pep:          false,
      sanctions:    false,
      jurisdiction: 'BRA',
      licenseType:  'EMI',
      lastReviewAt: daysAgo(30),
      nextReviewAt: new Date(Date.now() + 150*86400000).toISOString(),
    },

    /* ── Trade Orders ───────────────────────────────── */
    tradeOrders: [
      { id:uid(), symbol:'BTC', side:'BUY',  amount:0.0543, price:345000, total:1873350, status:'FILLED',    filledAt:daysAgo(3) },
      { id:uid(), symbol:'ETH', side:'SELL', amount:0.5,    price:18100,  total:905000,  status:'FILLED',    filledAt:daysAgo(7) },
      { id:uid(), symbol:'SOL', side:'BUY',  amount:5,      price:850,    total:4250,    status:'CANCELLED', createdAt:daysAgo(5) },
      { id:uid(), symbol:'BTC', side:'SELL', amount:0.02,   price:340000, total:680000,  status:'FILLED',    filledAt:daysAgo(12) },
    ],

    /* ── Conversion Audits ──────────────────────────── */
    conversionAudits: [
      { id:uid(), trigger:'PIX_RECEIVE', from:'USDT', to:'BRL',  amount:100,   converted:57224, rate:572.24,   createdAt:daysAgo(2) },
      { id:uid(), trigger:'CARD_JIT',    from:'BTC',  to:'BRL',  amount:0.001, converted:34582, rate:34582000, createdAt:daysAgo(4) },
      { id:uid(), trigger:'PIX_RECEIVE', from:'USDT', to:'BRL',  amount:500,   converted:286120,rate:572.24,   createdAt:daysAgo(6) },
    ],
  }
}

/* ── Simulation Engine ──────────────────────────────── */
export function simulate(store, action, data = {}) {
  const s = JSON.parse(JSON.stringify(store))
  let msg = '', details = {}

  switch (action) {

    /* ── PIX ───────────────────────────── */
    case 'pix_send': {
      const amt = cents(data.amount || 100)
      const fee = Math.round(amt * 0.002)
      s.accounts.main.balance -= amt + fee
      s.kyc.used.daily = (s.kyc.used.daily || 0) + amt
      s.transactions.unshift({ id:uid(), type:'PIX_OUT', amount:-(amt+fee), desc:`PIX → ${data.receiver||'Destinatário'}`, at:now.toISOString(), status:'CONFIRMED', category:'PIX' })
      msg = `PIX enviado: ${fmtBRL(amt)} (taxa ${fmtBRL(fee)})`
      details = { endToEndId:`E${uid()}`, status:'CONFIRMED', fee:fmtBRL(fee), net:fmtBRL(amt-fee) }
      break
    }
    case 'pix_receive': {
      const amt = cents(data.amount || 200)
      s.accounts.main.balance += amt
      s.transactions.unshift({ id:uid(), type:'PIX_IN', amount:amt, desc:`PIX ← ${data.sender||'Remetente'}`, at:now.toISOString(), status:'CONFIRMED', category:'PIX' })
      msg = `PIX recebido: ${fmtBRL(amt)}`
      details = { endToEndId:`E${uid()}`, status:'CONFIRMED' }
      break
    }
    case 'pix_key_register': {
      s.pixKeys.push({ id:uid(), type:data.type||'EMAIL', key:data.key||'nova@chave.com', createdAt:now.toISOString() })
      msg = `Chave PIX ${data.type||'EMAIL'} registrada`
      details = { type:data.type, key:data.key, status:'ACTIVE' }
      break
    }
    case 'pix_send_crypto': {
      const usdt = Number(data.usdt || 100)
      const brl  = Math.round(usdt * s.pricingRate.USDT * 0.988)
      s.accounts.main.balance += brl
      s.transactions.unshift({ id:uid(), type:'PIX_IN', amount:brl, desc:`PIX via USDT (${usdt} USDT)`, at:now.toISOString(), status:'CONFIRMED', category:'Cripto' })
      msg = `PIX via cripto: recebeu ${fmtBRL(brl)} de ${usdt} USDT`
      details = { usdt, brl:fmtBRL(brl), spread:'1.2%', txHash:`0x${uid()}${uid()}` }
      break
    }

    /* ── Pagamentos ─────────────────────── */
    case 'payment_validate': {
      msg = `Boleto validado: ${data.barcode?.slice(0,20)||'00000000000'}...`
      details = { status:'VALID', amount:fmtBRL(cents(data.amount||120)), beneficiary:'Empresa Exemplo LTDA', dueDate:data.dueDate||'2026-06-01' }
      break
    }
    case 'payment_schedule': {
      const amt = cents(data.amount || 120)
      s.payments.push({ id:uid(), amount:amt, barcode:data.barcode||'00000.00000 00000.000000 0', status:'SCHEDULED', scheduledAt:data.dueDate||now.toISOString(), createdAt:now.toISOString() })
      msg = `Pagamento agendado: ${fmtBRL(amt)}`
      details = { status:'SCHEDULED', dueDate:data.dueDate||'2026-06-01', authId:uid() }
      break
    }
    case 'payment_confirm': {
      const amt = s.payments.length ? s.payments[0].amount : cents(data.amount||120)
      s.accounts.main.balance -= amt
      s.transactions.unshift({ id:uid(), type:'PAYMENT', amount:-amt, desc:'Boleto Confirmado', at:now.toISOString(), status:'CONFIRMED', category:'Pagamento' })
      if (s.payments.length) { s.payments[0].status = 'CONFIRMED' }
      msg = `Pagamento confirmado: ${fmtBRL(amt)}`
      details = { status:'CONFIRMED', authCode:uid() }
      break
    }
    case 'payment_reject': {
      if (s.payments.length) s.payments[0].status = 'REJECTED'
      msg = 'Pagamento rejeitado'
      details = { status:'REJECTED', reason:data.reason||'FRAUD_SUSPICION' }
      break
    }

    /* ── Cobranças (Invoice) ─────────────── */
    case 'invoice_create': {
      const amt = cents(data.amount || 500)
      const inv = { id:uid(), amount:amt, pixKey:`${uid()}@zetta.bank`, usdtAddr:`0x${uid()}${uid()}`, createdAt:now.toISOString(), status:'PENDING', desc:data.desc||'Cobrança Z-Finance' }
      s.invoices.push(inv)
      msg = `Cobrança criada: ${fmtBRL(amt)}`
      details = { id:inv.id, pixKey:inv.pixKey, usdtAddr:`${inv.usdtAddr.slice(0,18)}...`, status:'PENDING' }
      break
    }
    case 'invoice_pay': {
      const inv = s.invoices.find(i=>i.status==='PENDING')
      if (!inv) { msg = 'Nenhuma cobrança pendente'; break }
      s.accounts.main.balance += inv.amount
      inv.status = 'PAID'
      s.transactions.unshift({ id:uid(), type:'PIX_IN', amount:inv.amount, desc:`Cobrança paga — ${inv.id.slice(0,8)}`, at:now.toISOString(), status:'CONFIRMED', category:'Cobrança' })
      msg = `Cobrança paga: ${fmtBRL(inv.amount)}`
      details = { invoiceId:inv.id, method:data.method||'PIX', status:'PAID' }
      break
    }

    /* ── Transferências ─────────────────── */
    case 'transfer_internal': {
      const amt = cents(data.amount || 300)
      const from = data.from || 'main'
      const to   = data.to   || 'invest'
      if (s.accounts[from]) s.accounts[from].balance -= amt
      if (s.accounts[to])   s.accounts[to].balance   += amt
      s.transactions.unshift({ id:uid(), type:'TRANSFER_OUT', amount:-amt, desc:`Transferência → ${s.accounts[to]?.label||to}`, at:now.toISOString(), status:'CONFIRMED', category:'Transferência' })
      msg = `${fmtBRL(amt)} transferidos para ${s.accounts[to]?.label||to}`
      details = { from, to, amount:fmtBRL(amt), status:'CONFIRMED' }
      break
    }
    case 'withdrawal': {
      const amt = cents(data.amount || 100)
      s.accounts.main.balance -= amt
      s.transactions.unshift({ id:uid(), type:'WITHDRAWAL', amount:-amt, desc:`Saque — ${data.pix||'Chave PIX'}`, at:now.toISOString(), status:'CONFIRMED', category:'Saque' })
      msg = `Saque de ${fmtBRL(amt)} processado`
      details = { status:'CONFIRMED', pixKey:data.pix, txId:uid() }
      break
    }

    /* ── Card JIT ───────────────────────── */
    case 'card_authorize': {
      const amt = cents(data.amount || 250)
      s.accounts.main.holds = (s.accounts.main.holds||0) + amt
      msg = `Autorização JIT: ${fmtBRL(amt)} — ${data.merchant||'Merchant'}`
      details = { authCode:uid(), status:'HOLD', merchant:data.merchant||'Merchant', mcc:data.mcc||'5411', amount:fmtBRL(amt) }
      break
    }
    case 'card_confirm': {
      const hold = s.accounts.main.holds || 0
      if (!hold) { msg='Nenhum hold ativo'; break }
      s.accounts.main.balance -= hold
      s.accounts.main.holds   = 0
      s.card.limitUsed += hold
      s.transactions.unshift({ id:uid(), type:'CARD_AUTH', amount:-hold, desc:'Compra JIT confirmada', at:now.toISOString(), status:'CONFIRMED', category:'Cartão' })
      msg = `Compra JIT confirmada: ${fmtBRL(hold)}`
      details = { status:'SETTLED', amount:fmtBRL(hold) }
      break
    }
    case 'card_reject': {
      const hold = s.accounts.main.holds || 0
      s.accounts.main.holds = 0
      msg = `Compra rejeitada — hold de ${fmtBRL(hold)} liberado`
      details = { status:'REJECTED', reason:data.reason||'FRAUD_SUSPICION', holdReleased:fmtBRL(hold) }
      break
    }
    case 'card_freeze': {
      s.card.frozen = true
      msg = 'Cartão congelado com sucesso'
      details = { frozen:true, updatedAt:now.toISOString() }
      break
    }
    case 'card_unfreeze': {
      s.card.frozen = false
      msg = 'Cartão desbloqueado com sucesso'
      details = { frozen:false, updatedAt:now.toISOString() }
      break
    }

    /* ── Cripto / Swap ──────────────────── */
    case 'crypto_swap': {
      const from = data.from || 'BRL'
      const to   = data.to   || 'BTC'
      const amt  = Number(data.amount || 100)
      const spread = s.pricingRate.spread
      if (from === 'BRL') {
        const brlCents = cents(amt)
        const coin = s.crypto.find(c=>c.symbol===to)
        if (coin) {
          const received = (amt / coin.price) * (1 - spread)
          coin.amount += received
          s.accounts.main.balance -= brlCents
          const fee = brlCents * spread
          s.transactions.unshift({ id:uid(), type:'TRADE_BUY', amount:-brlCents, desc:`Swap BRL → ${received.toFixed(6)} ${to}`, at:now.toISOString(), status:'CONFIRMED', category:'Cripto' })
          msg = `Swap: ${fmtBRL(brlCents)} → ${received.toFixed(6)} ${to}`
          details = { sold:fmtBRL(brlCents), received:`${received.toFixed(6)} ${to}`, fee:fmtBRL(fee), rate:`R$ ${coin.price.toLocaleString('pt-BR')}/${to}`, spread:'1.2%' }
        }
      } else {
        const coin = s.crypto.find(c=>c.symbol===from)
        if (coin && coin.amount >= amt) {
          const brl = Math.round(amt * coin.price * (1 - spread) * 100)
          coin.amount -= amt
          s.accounts.main.balance += brl
          s.transactions.unshift({ id:uid(), type:'TRADE_SELL', amount:brl, desc:`Swap ${amt} ${from} → BRL`, at:now.toISOString(), status:'CONFIRMED', category:'Cripto' })
          msg = `Swap: ${amt} ${from} → ${fmtBRL(brl)}`
          details = { sold:`${amt} ${from}`, received:fmtBRL(brl), spread:'1.2%', rate:`R$ ${coin.price.toLocaleString('pt-BR')}/${from}` }
        } else {
          msg = `Saldo insuficiente de ${from}`
          details = { error:'INSUFFICIENT_BALANCE', available:coin?.amount||0 }
        }
      }
      break
    }
    case 'crypto_pay': {
      const coin = s.crypto.find(c=>c.symbol===(data.asset||'ETH'))
      if (!coin) { msg='Ativo não encontrado'; break }
      const usdAmt = Number(data.amount||50)
      const coinAmt = usdAmt / (coin.price / 578)
      coin.amount -= coinAmt
      msg = `Pago com ${coinAmt.toFixed(6)} ${coin.symbol} ≈ US$ ${usdAmt}`
      details = { asset:coin.symbol, spent:coinAmt.toFixed(6), usdValue:`$${usdAmt}`, txHash:`0x${uid()}${uid()}`, status:'CONFIRMED' }
      break
    }
    case 'pricing_quote': {
      const pair = data.pair || 'BTC/BRL'
      const [base] = pair.split('/')
      const coin = s.crypto.find(c=>c.symbol===base)
      const price = coin?.price || 345820
      const bid = price * 0.999, ask = price * 1.001
      msg = `Cotação ${pair}: R$ ${price.toLocaleString('pt-BR')}`
      details = { pair, bid:fmtBRL(bid*100), ask:fmtBRL(ask*100), spread:'0.2%', ts:now.toISOString(), source:'PricingEngine/v2' }
      break
    }
    case 'pricing_update': {
      s.pricingRate.BTC = s.pricingRate.BTC * (1 + (Math.random()-0.5)*0.02)
      msg = 'Pricing atualizado com snapshot de mercado'
      details = { BTC:fmtBRL(s.pricingRate.BTC*100), ETH:fmtBRL(s.pricingRate.ETH*100), updatedAt:now.toISOString() }
      break
    }
    case 'crypto_liquidate': {
      const coin = s.crypto.find(c=>c.amount > 0 && c.symbol !== 'USDT')
      if (!coin) { msg='Nenhuma posição para liquidar'; break }
      const half = coin.amount / 2
      const brl  = Math.round(half * coin.price * 100)
      coin.amount -= half
      s.accounts.main.balance += brl
      s.transactions.unshift({ id:uid(), type:'TRADE_SELL', amount:brl, desc:`Liquidação ${half.toFixed(4)} ${coin.symbol}`, at:now.toISOString(), status:'CONFIRMED', category:'Cripto' })
      msg = `Liquidado ${half.toFixed(4)} ${coin.symbol} → ${fmtBRL(brl)}`
      details = { sold:`${half.toFixed(4)} ${coin.symbol}`, received:fmtBRL(brl), status:'CONFIRMED' }
      break
    }

    /* ── KYC ────────────────────────────── */
    case 'kyc_limits': {
      const k = s.kyc
      msg = `KYC ${k.level}: diário ${fmtBRL(k.limits.daily)}, mensal ${fmtBRL(k.limits.monthly)}`
      details = { level:k.level, status:k.status, daily:fmtBRL(k.limits.daily), monthly:fmtBRL(k.limits.monthly), usedDaily:fmtBRL(k.used.daily), remainingDaily:fmtBRL(k.limits.daily - k.used.daily) }
      break
    }
    case 'kyc_upgrade': {
      s.kyc.level = 'FULL'; s.kyc.status = 'VERIFIED'
      s.kyc.limits = { daily:20000000, monthly:100000000 }
      msg = 'KYC FULL verificado — limites máximos liberados'
      details = { newLevel:'FULL', daily:fmtBRL(20000000), monthly:fmtBRL(100000000), providerRef:`KYC-${uid()}` }
      break
    }

    /* ── Compliance ─────────────────────── */
    case 'compliance_case': {
      const c = { id:uid(), type:data.type||'SUSPICIOUS_ACTIVITY', status:'OPEN', riskLevel:data.risk||'HIGH', title:data.title||'Atividade suspeita detectada', createdAt:now.toISOString() }
      s.compliance.push(c)
      msg = `Case aberto: ${c.title} [${c.riskLevel}]`
      details = { caseId:c.id, type:c.type, riskLevel:c.riskLevel, status:'OPEN' }
      break
    }
    case 'compliance_event': {
      msg = 'Evento de compliance registrado'
      details = { eventType:data.eventType||'MANUAL_REVIEW', caseId:data.caseId||'—', analyst:'Sistema Zetta', ts:now.toISOString() }
      break
    }

    /* ── Pre-registration ───────────────── */
    case 'pre_registration': {
      const pr = { id:uid(), fullName:data.fullName||'Empresa Exemplo', email:data.email||'empresa@example.com', phone:data.phone||'+5511999999999', status:'PENDING', emailStatus:'TOKEN_SENT', phoneStatus:'PENDING', createdAt:now.toISOString(), expiresAt:new Date(Date.now()+3600000).toISOString() }
      s.preRegistrations.push(pr)
      msg = `Pré-cadastro iniciado: ${pr.email}`
      details = { id:pr.id, emailStatus:'TOKEN_SENT', phoneStatus:'PENDING', expiresAt:pr.expiresAt }
      break
    }
    case 'pre_registration_verify': {
      const pr = s.preRegistrations.find(p=>p.status==='PENDING')
      if (pr) { pr.emailStatus='VERIFIED'; pr.phoneStatus='VERIFIED'; pr.status='VERIFIED' }
      msg = 'Contato verificado — pré-cadastro aprovado'
      details = { status:'VERIFIED', emailStatus:'VERIFIED', phoneStatus:'VERIFIED' }
      break
    }

    /* ── Admin ──────────────────────────── */
    case 'admin_plan_change': {
      s.user.plan = data.plan || 'BUSINESS'
      msg = `Plano alterado para ${s.user.plan}`
      details = { userId:s.user.id, newPlan:s.user.plan, effectiveAt:now.toISOString(), auditId:uid() }
      break
    }
    case 'admin_feature_toggle': {
      msg = `Feature ${data.feature||'CRYPTO'} ${data.enabled?'habilitada':'desabilitada'}`
      details = { feature:data.feature, enabled:data.enabled??true, scope:'USER_OVERRIDE', auditId:uid() }
      break
    }
    case 'admin_limit_adjust': {
      s.kyc.limits.daily = cents(data.daily || 50000)
      s.kyc.limits.monthly = cents(data.monthly || 200000)
      msg = `Limites ajustados: diário ${fmtBRL(s.kyc.limits.daily)}`
      details = { daily:fmtBRL(s.kyc.limits.daily), monthly:fmtBRL(s.kyc.limits.monthly), approvedBy:'ADMIN', auditId:uid() }
      break
    }
    case 'admin_user_block': {
      s.user.status = 'BLOCKED'
      msg = 'Usuário bloqueado — acesso suspenso'
      details = { userId:s.user.id, status:'BLOCKED', reason:data.reason||'COMPLIANCE_REVIEW', ts:now.toISOString() }
      break
    }

    /* ── Observability ──────────────────── */
    case 'obs_summary': {
      msg = `Sistema: ${s.observability.uptime} uptime, P99 ${s.observability.latencyP99}ms`
      details = { ...s.observability, traceId:uid(), checkedAt:now.toISOString() }
      break
    }
    case 'audit_archive': {
      const log = { id:uid(), action:'AUDIT_EXPORT', entityType:'SYSTEM', data:{ records:s.transactions.length, exportedAt:now.toISOString() }, createdAt:now.toISOString() }
      s.auditLogs.push(log)
      msg = `Auditoria exportada: ${s.transactions.length} registros`
      details = { exportId:log.id, records:s.transactions.length, format:'JSON+NDJSON', ts:now.toISOString() }
      break
    }
    case 'webhook_retry': {
      s.observability.pendingWebhooks = Math.max(0, s.observability.pendingWebhooks - 1)
      msg = `Webhook reprocessado — ${s.observability.pendingWebhooks} pendentes`
      details = { processed:1, remaining:s.observability.pendingWebhooks, deliveredAt:now.toISOString() }
      break
    }

    /* ── User Settings ──────────────────── */
    case 'settings_ux_mode': {
      s.user.uxMode = data.mode || 'PRO'
      msg = `Modo UX alterado para ${s.user.uxMode}`
      details = { uxMode:s.user.uxMode, userId:s.user.id }
      break
    }
    case 'settings_auto_convert': {
      msg = `Auto-conversão ${data.enabled?'habilitada':'desabilitada'}: ${data.asset||'BTC'}`
      details = { autoConvertEnabled:data.enabled, asset:data.asset||'BTC', minAmount:fmtBRL(cents(data.min||100)) }
      break
    }

    /* ── Reconciliação ──────────────────── */
    case 'reconcile_summary': {
      msg = `Reconciliação: ${s.reconcile.summary.pix} PIX, ${s.reconcile.summary.payment} boletos pendentes`
      break
    }
    case 'reconcile_resolve': {
      const idx = s.reconcile.pending.findIndex(p => p.id === data.pendingId)
      if (idx >= 0) {
        s.reconcile.pending.splice(idx, 1)
        const type = data.pendingId ? s.reconcile.summary : null
        if (s.reconcile.summary.pix   > 0) s.reconcile.summary.pix--
      } else {
        s.reconcile.pending.shift()
      }
      s.observability.reconcilePending = s.reconcile.pending.length
      msg = `Pendência resolvida manualmente`
      break
    }

    /* ── Alertas ────────────────────────── */
    case 'alerts_check': {
      s.alerts.last = s.alerts.last.map(a => ({ ...a, checkedAt: now.toISOString() }))
      msg = `Check de alertas: todos os ${s.alerts.last.length} thresholds OK`
      break
    }
    case 'alerts_update_threshold': {
      const key = data.type || 'pixPending'
      s.alerts.thresholds[key] = Number(data.value || 10)
      msg = `Threshold ${key} atualizado para ${data.value}`
      break
    }

    /* ── RBAC / Roles ───────────────────── */
    case 'role_create': {
      const newRole = { id:uid(), name:(data.name||'NOVO_ROLE').toUpperCase(), description:data.description||'', usersCount:0, createdAt:now.toISOString() }
      s.roles.push(newRole)
      msg = `Role ${newRole.name} criada`
      break
    }
    case 'role_assign': {
      s.userRoles.push({ userId:data.userId||uid(), userName:data.userName||'Usuário', roleName:data.roleName||'SUPPORT', assignedAt:now.toISOString(), assignedBy:s.user.email })
      const r = s.roles.find(r => r.name === (data.roleName||'SUPPORT'))
      if (r) r.usersCount++
      msg = `Role ${data.roleName} atribuída ao usuário ${data.userId}`
      break
    }
    case 'role_remove': {
      const bi = s.userRoles.findIndex(ur => ur.userId === data.userId && ur.roleName === data.roleName)
      if (bi >= 0) {
        s.userRoles.splice(bi, 1)
        const r = s.roles.find(r => r.name === data.roleName)
        if (r && r.usersCount > 0) r.usersCount--
      }
      msg = `Role ${data.roleName} removida do usuário ${data.userId}`
      break
    }
    case 'role_separation_add': {
      s.separationRules.push({ id:uid(), roleA:data.roleA||'ADMIN', roleB:data.roleB||'TRADER', reason:data.reason||'Segregação de funções', createdAt:now.toISOString() })
      msg = `Regra de separação ${data.roleA}↔${data.roleB} criada`
      break
    }
    case 'role_separation_remove': {
      const ri = s.separationRules.findIndex(r => r.roleA === data.roleA && r.roleB === data.roleB)
      if (ri >= 0) s.separationRules.splice(ri, 1)
      msg = `Regra de separação removida`
      break
    }

    /* ── Perfil Regulatório ──────────────── */
    case 'regulatory_profile_update': {
      Object.assign(s.regulatoryProfile, {
        fatfRisk:     data.fatfRisk     || s.regulatoryProfile.fatfRisk,
        pep:          data.pep === 'true',
        sanctions:    data.sanctions === 'true',
        jurisdiction: data.jurisdiction || s.regulatoryProfile.jurisdiction,
        licenseType:  data.licenseType  || s.regulatoryProfile.licenseType,
        vasp:         data.vasp !== 'false',
        lastReviewAt: now.toISOString(),
        nextReviewAt: new Date(Date.now() + 180*86400000).toISOString(),
      })
      msg = `Perfil regulatório atualizado — risco ${s.regulatoryProfile.fatfRisk}`
      break
    }

    /* ── Pricing avançado ───────────────── */
    case 'pricing_version_create': {
      const ver = { id:uid(), version:data.version||`v${Date.now().toString().slice(-4)}`, status:'DRAFT', notes:data.notes||'', createdAt:now.toISOString() }
      s.pricingVersions.unshift(ver)
      msg = `Versão de pricing ${ver.version} criada (DRAFT)`
      break
    }
    case 'pricing_version_activate': {
      s.pricingVersions.forEach(v => { if (v.status === 'ACTIVE') v.status = 'ARCHIVED' })
      const v = s.pricingVersions.find(v => v.version === data.version) || s.pricingVersions[0]
      if (v) { v.status = 'ACTIVE'; v.activatedAt = now.toISOString() }
      msg = `Versão ${v?.version} ativada`
      break
    }
    case 'pricing_rule_create': {
      s.pricingRules.push({ id:uid(), planCode:data.planCode||'BUSINESS', userType:data.userType||'PJ', featureCode:data.featureCode||'PIX_SEND', feeType:data.feeType||'FLAT', feeValue:Number(data.feeValue||0) })
      msg = `Regra de pricing criada para ${data.planCode}/${data.featureCode}`
      break
    }
    case 'pricing_campaign_create': {
      s.pricingCampaigns.push({ id:uid(), name:data.name||'Nova Campanha', status:'DRAFT', discount:Number(data.discount||10), startDate:data.startDate||'2026-01-01', endDate:data.endDate||'2026-12-31', planCode:data.planCode||'BUSINESS', usageCount:0 })
      msg = `Campanha "${data.name}" criada`
      break
    }
    case 'pricing_campaign_update': {
      const camp = s.pricingCampaigns.find(c => c.id === data.campaignId) || s.pricingCampaigns[0]
      if (camp) { camp.status = data.status || camp.status; camp.discount = Number(data.discount || camp.discount) }
      msg = `Campanha atualizada: status ${camp?.status}`
      break
    }

    /* ── Trade Orders ───────────────────── */
    case 'trade_order_create': {
      const sym = data.symbol || 'BTC'
      const coinPrice = s.pricingRate[sym] || 100
      const amt = Number(data.amount || 0.001)
      const total = Math.round(amt * coinPrice)
      const order = { id:uid(), symbol:sym, side:data.side||'BUY', amount:amt, price:coinPrice, total, status:'FILLED', filledAt:now.toISOString() }
      s.tradeOrders.unshift(order)
      if (data.side === 'SELL') {
        const coin = s.crypto.find(c => c.symbol === sym)
        if (coin && coin.amount >= amt) { coin.amount -= amt; s.accounts.main.balance += total }
      } else {
        const coin = s.crypto.find(c => c.symbol === sym)
        if (coin) coin.amount += amt
        s.accounts.main.balance -= total
      }
      msg = `Ordem ${data.side} ${amt} ${sym} @ ${fmtBRL(coinPrice)} executada`
      break
    }
    case 'trade_order_cancel': {
      const o = s.tradeOrders.find(o => o.status !== 'FILLED' && o.status !== 'CANCELLED')
      if (o) o.status = 'CANCELLED'
      msg = `Ordem cancelada`
      break
    }

    default:
      msg = `Ação "${action}" executada`
  }

  /* Audit trail automático */
  if (msg && action !== 'obs_summary' && action !== 'reconcile_summary' && action !== 'alerts_check') {
    s.auditLogs.unshift({ id:uid(), action:action.toUpperCase(), entityType:'SYSTEM', details:data, ts:now.toISOString() })
    if (s.auditLogs.length > 50) s.auditLogs.pop()
  }

  return s
}

/* ── Helpers ────────────────────────────────────────── */
export const cents = (v) => Math.round(Number(v||0) * 100)

export function cryptoBRL(crypto) {
  return (crypto||[]).reduce((a,c) => a + c.amount * c.price * 100, 0)
}

export function totalWealthBRL(store) {
  const accounts = Object.values(store.accounts).reduce((a,acc) => {
    if (acc.currency === 'BRL') return a + acc.balance
    if (acc.currency === 'USD') return a + acc.balance * store.market.USD.rate
    return a
  }, 0)
  const invests = (store.investments.cdb||0) + (store.investments.tesouro||0) + (store.investments.lci||0)
  return accounts + invests + cryptoBRL(store.crypto)
}
