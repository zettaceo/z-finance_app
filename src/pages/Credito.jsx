import React, { useState, useMemo } from 'react'
import { Shield, TrendingUp, Zap, CreditCard, CheckCircle, Clock, AlertCircle, ChevronRight, DollarSign } from 'lucide-react'
import { AreaChart, Area, ResponsiveContainer } from 'recharts'
import { useApp } from '../App.jsx'
import SimModal from '../components/SimModal.jsx'

function fmt(cents) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100)
}

const SCORE_COLORS = [
  { min: 0,   max: 300,  label: 'Muito Baixo', color: '#EF4444' },
  { min: 300, max: 500,  label: 'Baixo',       color: '#F97316' },
  { min: 500, max: 700,  label: 'Regular',     color: '#FBBF24' },
  { min: 700, max: 850,  label: 'Bom',         color: '#34D399' },
  { min: 850, max: 1000, label: 'Excelente',   color: '#00E599' },
]

function scoreConfig(score) {
  return SCORE_COLORS.find(c => score >= c.min && score < c.max) || SCORE_COLORS[SCORE_COLORS.length - 1]
}

function ScoreGauge({ score }) {
  const cfg = scoreConfig(score)
  const pct = score / 1000
  const r = 70
  const circ = 2 * Math.PI * r
  const arc = circ * 0.75 // 270° arc
  const offset = arc * (1 - pct)

  return (
    <div style={{ position: 'relative', width: 180, height: 130, margin: '0 auto' }}>
      <svg width="180" height="130" viewBox="0 0 180 130" style={{ overflow: 'visible' }}>
        <defs>
          <linearGradient id="scoreGrad" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#EF4444" />
            <stop offset="33%" stopColor="#FBBF24" />
            <stop offset="66%" stopColor="#34D399" />
            <stop offset="100%" stopColor="#00E599" />
          </linearGradient>
        </defs>
        {/* Background arc */}
        <circle
          cx="90" cy="95" r={r}
          fill="none" stroke="rgba(255,255,255,0.06)" strokeWidth="12"
          strokeDasharray={`${arc} ${circ - arc}`}
          strokeDashoffset={circ * 0.125}
          strokeLinecap="round"
          transform="rotate(135 90 95)"
        />
        {/* Score arc */}
        <circle
          cx="90" cy="95" r={r}
          fill="none" stroke="url(#scoreGrad)" strokeWidth="12"
          strokeDasharray={`${arc - offset} ${circ - (arc - offset)}`}
          strokeDashoffset={circ * 0.125}
          strokeLinecap="round"
          transform="rotate(135 90 95)"
          style={{ transition: 'stroke-dasharray 1s ease' }}
        />
      </svg>
      <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, textAlign: 'center' }}>
        <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 36, fontWeight: 800, color: cfg.color, margin: 0, lineHeight: 1 }}>{score}</p>
        <p style={{ fontSize: 12, fontWeight: 700, color: cfg.color, margin: '4px 0 0' }}>{cfg.label}</p>
      </div>
    </div>
  )
}

const OFFER_LABELS = { PERSONAL: 'Crédito Pessoal', CRYPTO_COL: 'Colateral Crypto', BUSINESS: 'Empresarial' }
const OFFER_ICONS = { PERSONAL: DollarSign, CRYPTO_COL: Zap, BUSINESS: CreditCard }

export default function Credito() {
  const { store, modal, setModal, dispatch } = useApp()
  const [tab, setTab] = useState('score')
  const [simAmt, setSimAmt] = useState(500000)
  const [simMonths, setSimMonths] = useState(12)
  const [simRate, setSimRate] = useState(1.8)

  const c = store.credit
  const availableLimit = c.limit - c.limitUsed
  const usedPct = (c.limitUsed / c.limit) * 100

  const monthly = useMemo(() => {
    const r = simRate / 100
    if (r === 0) return simAmt / simMonths
    return Math.round(simAmt * r / (1 - Math.pow(1 + r, -simMonths)))
  }, [simAmt, simMonths, simRate])

  const totalPay = monthly * simMonths
  const totalInterest = totalPay - simAmt

  const histData = c.scoreHistory.map((v, i) => ({ v, i }))
  const scoreCfg = scoreConfig(c.score)

  const TABS = [
    { id: 'score',   label: 'Z-Score' },
    { id: 'linhas',  label: 'Linhas' },
    { id: 'simular', label: 'Simulador' },
    { id: 'hist',    label: 'Histórico' },
  ]

  return (
    <div>
      <div style={{ marginBottom: 20 }}>
        <h2 style={{ fontSize: 22, fontWeight: 800, fontFamily: 'Syne,sans-serif', color: 'var(--t1)', margin: '0 0 4px' }}>Crédito</h2>
        <p style={{ fontSize: 13, color: 'var(--t3)', margin: 0 }}>Z-Score, linhas de crédito e simulador</p>
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 6, marginBottom: 20, background: 'var(--surface)', borderRadius: 14, padding: 4 }}>
        {TABS.map(t => (
          <button key={t.id} onClick={() => setTab(t.id)} style={{
            flex: 1, padding: '8px 4px', borderRadius: 10, border: 'none', cursor: 'pointer',
            background: tab === t.id ? 'var(--accent)' : 'transparent',
            color: tab === t.id ? '#040C1B' : 'var(--t3)',
            fontSize: 12, fontWeight: 700, transition: 'all 0.15s',
          }}>{t.label}</button>
        ))}
      </div>

      {/* ── Z-Score tab ── */}
      {tab === 'score' && (
        <>
          {/* Score card */}
          <div style={{
            background: 'linear-gradient(135deg, var(--surface) 0%, #0A1928 100%)',
            border: `1px solid ${scoreCfg.color}30`,
            borderRadius: 24, padding: '28px 20px 20px',
            marginBottom: 16, textAlign: 'center',
          }}>
            <p style={{ fontSize: 11, fontWeight: 700, color: 'var(--t3)', textTransform: 'uppercase', letterSpacing: '0.1em', marginBottom: 16 }}>
              Z-Score Global
            </p>
            <ScoreGauge score={c.score} />
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 20, paddingTop: 16, borderTop: '1px solid var(--border)' }}>
              <span style={{ fontSize: 12, color: 'var(--t3)' }}>0 — Muito Baixo</span>
              <span style={{ fontSize: 12, color: 'var(--t3)' }}>1000 — Excelente</span>
            </div>
          </div>

          {/* Score breakdown */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginBottom: 16 }}>
            {[
              { label: 'Histórico de Pagamentos', value: '98%', color: '#34D399', icon: CheckCircle },
              { label: 'Utilização de Crédito', value: `${usedPct.toFixed(0)}%`, color: usedPct > 70 ? '#F87171' : '#FBBF24', icon: TrendingUp },
              { label: 'Colateral Disponível', value: fmt(c.collateral.totalBRL), color: '#818CF8', icon: Shield },
              { label: 'Tempo de Relacionamento', value: '14 meses', color: '#60A5FA', icon: Clock },
            ].map(({ label, value, color, icon: Icon }) => (
              <div key={label} style={{
                background: 'var(--surface)', border: '1px solid var(--border)',
                borderRadius: 16, padding: '14px',
              }}>
                <Icon size={16} style={{ color, marginBottom: 8 }} />
                <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 15, fontWeight: 700, color, margin: '0 0 4px' }}>{value}</p>
                <p style={{ fontSize: 11, color: 'var(--t3)', margin: 0, lineHeight: 1.3 }}>{label}</p>
              </div>
            ))}
          </div>

          {/* Score history sparkline */}
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 18, padding: '16px 20px' }}>
            <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 12px' }}>Evolução do Z-Score</p>
            <ResponsiveContainer width="100%" height={80}>
              <AreaChart data={histData}>
                <defs>
                  <linearGradient id="scoreHistGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor={scoreCfg.color} stopOpacity={0.3} />
                    <stop offset="100%" stopColor={scoreCfg.color} stopOpacity={0} />
                  </linearGradient>
                </defs>
                <Area type="monotone" dataKey="v" stroke={scoreCfg.color} strokeWidth={2} fill="url(#scoreHistGrad)" dot={false} />
              </AreaChart>
            </ResponsiveContainer>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 4 }}>
              <span style={{ fontSize: 11, color: 'var(--t3)' }}>6 meses atrás</span>
              <span style={{ fontSize: 11, color: scoreCfg.color, fontWeight: 700 }}>+{c.score - c.scoreHistory[0]} pts</span>
              <span style={{ fontSize: 11, color: 'var(--t3)' }}>Hoje</span>
            </div>
          </div>
        </>
      )}

      {/* ── Linhas de crédito tab ── */}
      {tab === 'linhas' && (
        <>
          {/* Available limit */}
          <div style={{
            background: 'linear-gradient(135deg, rgba(0,229,153,0.08), rgba(0,229,153,0.03))',
            border: '1px solid rgba(0,229,153,0.2)',
            borderRadius: 20, padding: '20px', marginBottom: 16,
          }}>
            <p style={{ fontSize: 12, color: 'var(--t3)', fontWeight: 600, margin: '0 0 8px', textTransform: 'uppercase', letterSpacing: '0.08em' }}>
              Limite Disponível
            </p>
            <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 32, fontWeight: 800, color: 'var(--accent)', margin: '0 0 4px' }}>
              {fmt(availableLimit)}
            </p>
            <p style={{ fontSize: 13, color: 'var(--t3)', margin: '0 0 16px' }}>de {fmt(c.limit)} total</p>
            <div style={{ height: 6, borderRadius: 3, background: 'rgba(255,255,255,0.06)', overflow: 'hidden' }}>
              <div style={{ height: '100%', borderRadius: 3, background: 'linear-gradient(90deg, var(--accent), var(--accent-2))', width: `${usedPct}%`, transition: 'width 0.5s ease' }} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 6 }}>
              <span style={{ fontSize: 11, color: 'var(--t3)' }}>Utilizado: {fmt(c.limitUsed)}</span>
              <span style={{ fontSize: 11, color: 'var(--t3)' }}>{usedPct.toFixed(0)}% em uso</span>
            </div>
          </div>

          {/* Active loans */}
          {c.loans.filter(l => l.status === 'ACTIVE').length > 0 && (
            <div style={{ marginBottom: 16 }}>
              <p style={{ fontSize: 13, fontWeight: 700, color: 'var(--t2)', margin: '0 0 10px' }}>Empréstimos Ativos</p>
              {c.loans.filter(l => l.status === 'ACTIVE').map(loan => (
                <div key={loan.id} style={{
                  background: 'var(--surface)', border: '1px solid var(--border)',
                  borderRadius: 16, padding: '16px', marginBottom: 10,
                }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                    <div>
                      <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{OFFER_LABELS[loan.type] || loan.type}</p>
                      <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>Taxa {loan.rate}% a.m. · {loan.months} meses</p>
                    </div>
                    <span style={{
                      fontSize: 10, fontWeight: 700, color: '#34D399',
                      background: 'rgba(52,211,153,0.15)', borderRadius: 8, padding: '3px 8px',
                    }}>ATIVO</span>
                  </div>
                  <div style={{ height: 4, borderRadius: 2, background: 'var(--surface-2)', overflow: 'hidden', marginBottom: 8 }}>
                    <div style={{ height: '100%', borderRadius: 2, background: '#818CF8', width: `${(loan.paid / loan.months) * 100}%` }} />
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ fontSize: 11, color: 'var(--t3)' }}>Parcela {loan.paid}/{loan.months}</span>
                    <span style={{ fontSize: 11, color: 'var(--t3)' }}>Próximo venc.: {new Date(loan.nextDue).toLocaleDateString('pt-BR')}</span>
                  </div>
                  <button
                    onClick={() => setModal({ title: 'Pagar Parcela', action: 'credit_pay', successMsg: 'Parcela paga com sucesso!', fields: [], submitLabel: 'Confirmar pagamento', description: `Parcela ${loan.paid + 1}/${loan.months} — ${fmt(Math.round(loan.amount / loan.months))}` })}
                    style={{ marginTop: 12, width: '100%', padding: '10px', borderRadius: 10, border: 'none', cursor: 'pointer', background: 'rgba(129,140,248,0.15)', color: '#818CF8', fontWeight: 700, fontSize: 13 }}
                  >
                    Pagar Parcela
                  </button>
                </div>
              ))}
            </div>
          )}

          {/* Offers */}
          <p style={{ fontSize: 13, fontWeight: 700, color: 'var(--t2)', margin: '0 0 10px' }}>Ofertas Disponíveis</p>
          {c.offers.map((offer, i) => {
            const Icon = OFFER_ICONS[offer.type] || DollarSign
            return (
              <div key={offer.id || i} style={{
                background: 'var(--surface)', border: '1px solid var(--border)',
                borderRadius: 16, padding: '16px', marginBottom: 10,
                display: 'flex', alignItems: 'center', gap: 14,
              }}>
                <div style={{
                  width: 44, height: 44, borderRadius: 14, flexShrink: 0,
                  background: 'rgba(0,229,153,0.1)',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                }}>
                  <Icon size={20} color="var(--accent)" />
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{OFFER_LABELS[offer.type] || offer.type}</p>
                  <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>
                    Até {fmt(offer.limit)} · {offer.rate}% a.m. · {offer.months}m
                    {offer.collateral && ` · Colateral ${offer.collateral}`}
                  </p>
                </div>
                <button
                  onClick={() => setModal({
                    title: 'Solicitar Crédito', action: 'credit_request', successMsg: 'Crédito aprovado e creditado!',
                    description: `${OFFER_LABELS[offer.type]} — Taxa ${offer.rate}% a.m.`,
                    fields: [
                      { key: 'amount', label: 'Valor (centavos)', type: 'number', default: Math.round(offer.limit / 2) },
                      { key: 'months', label: 'Prazo (meses)', type: 'number', default: offer.months, min: 1, max: offer.months },
                      { key: 'type', label: 'Tipo', type: 'select', options: [{ value: offer.type, label: OFFER_LABELS[offer.type] }] },
                      { key: 'rate', label: 'Taxa mensal (%)', type: 'number', default: offer.rate, step: 0.1 },
                    ],
                    submitLabel: 'Solicitar',
                  })}
                  style={{ flexShrink: 0, padding: '8px 14px', borderRadius: 10, border: 'none', cursor: 'pointer', background: 'rgba(0,229,153,0.15)', color: 'var(--accent)', fontWeight: 700, fontSize: 12 }}
                >
                  Contratar
                </button>
              </div>
            )
          })}
        </>
      )}

      {/* ── Simulador tab ── */}
      {tab === 'simular' && (
        <div>
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 20, padding: '20px', marginBottom: 16 }}>
            <p style={{ fontWeight: 700, fontSize: 15, color: 'var(--t1)', margin: '0 0 20px' }}>Simulador de Empréstimo</p>

            {[
              { label: 'Valor do empréstimo', value: simAmt, setter: setSimAmt, min: 100000, max: 10000000, step: 50000, fmt: (v) => fmt(v), type: 'range' },
              { label: 'Prazo (meses)', value: simMonths, setter: setSimMonths, min: 1, max: 60, step: 1, fmt: (v) => `${v} meses`, type: 'range' },
              { label: 'Taxa mensal (%)', value: simRate, setter: setSimRate, min: 0.5, max: 5, step: 0.1, fmt: (v) => `${v.toFixed(1)}% a.m.`, type: 'range' },
            ].map(({ label, value, setter, min, max, step, fmt: fmtFn }) => (
              <div key={label} style={{ marginBottom: 20 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ fontSize: 12, color: 'var(--t3)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</span>
                  <span style={{ fontFamily: 'DM Mono, monospace', fontSize: 13, fontWeight: 700, color: 'var(--accent)' }}>{fmtFn(value)}</span>
                </div>
                <input type="range" min={min} max={max} step={step} value={value}
                  onChange={e => setter(Number(e.target.value))}
                  style={{ width: '100%', accentColor: 'var(--accent)', cursor: 'pointer', height: 4 }}
                />
              </div>
            ))}
          </div>

          {/* Result card */}
          <div style={{
            background: 'linear-gradient(135deg, rgba(0,229,153,0.08), rgba(129,140,248,0.05))',
            border: '1px solid rgba(0,229,153,0.2)',
            borderRadius: 20, padding: '20px',
          }}>
            <p style={{ fontSize: 12, color: 'var(--t3)', margin: '0 0 16px', fontWeight: 600, textTransform: 'uppercase' }}>Resultado da Simulação</p>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              {[
                { label: 'Parcela mensal', value: fmt(monthly), color: 'var(--accent)', big: true },
                { label: 'Total pago', value: fmt(totalPay), color: 'var(--t1)', big: true },
                { label: 'Valor solicitado', value: fmt(simAmt), color: '#60A5FA', big: false },
                { label: 'Total de juros', value: fmt(totalInterest), color: '#F87171', big: false },
              ].map(({ label, value, color, big }) => (
                <div key={label} style={{ textAlign: 'center', padding: 12, background: 'rgba(255,255,255,0.03)', borderRadius: 12 }}>
                  <p style={{ fontFamily: 'DM Mono, monospace', fontSize: big ? 18 : 14, fontWeight: 700, color, margin: '0 0 4px' }}>{value}</p>
                  <p style={{ fontSize: 11, color: 'var(--t3)', margin: 0 }}>{label}</p>
                </div>
              ))}
            </div>
            <button
              onClick={() => dispatch('credit_simulate', { amount: simAmt / 100, rate: simRate, months: simMonths })}
              style={{ marginTop: 16, width: '100%', padding: '14px', borderRadius: 14, border: 'none', cursor: 'pointer', background: 'var(--accent)', color: '#040C1B', fontWeight: 700, fontSize: 15 }}
            >
              Solicitar Este Crédito
            </button>
          </div>
        </div>
      )}

      {/* ── Histórico tab ── */}
      {tab === 'hist' && (
        <div>
          {c.loans.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '48px 20px' }}>
              <AlertCircle size={40} style={{ color: 'var(--t3)', marginBottom: 12 }} />
              <p style={{ color: 'var(--t3)' }}>Nenhum empréstimo no histórico</p>
            </div>
          ) : (
            c.loans.map(loan => (
              <div key={loan.id} style={{
                background: 'var(--surface)', border: '1px solid var(--border)',
                borderRadius: 16, padding: '16px', marginBottom: 12,
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 10 }}>
                  <div>
                    <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 4px' }}>{OFFER_LABELS[loan.type] || loan.type}</p>
                    <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{fmt(loan.amount)} · {loan.rate}% a.m.</p>
                  </div>
                  <span style={{
                    fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 8,
                    color: loan.status === 'ACTIVE' ? '#34D399' : loan.status === 'PAID' ? '#818CF8' : '#F87171',
                    background: loan.status === 'ACTIVE' ? 'rgba(52,211,153,0.15)' : loan.status === 'PAID' ? 'rgba(129,140,248,0.15)' : 'rgba(248,113,113,0.15)',
                  }}>
                    {loan.status === 'ACTIVE' ? 'ATIVO' : loan.status === 'PAID' ? 'QUITADO' : loan.status}
                  </span>
                </div>
                <div style={{ height: 4, borderRadius: 2, background: 'var(--surface-2)', overflow: 'hidden' }}>
                  <div style={{ height: '100%', borderRadius: 2, background: '#818CF8', width: `${Math.min((loan.paid / loan.months) * 100, 100)}%` }} />
                </div>
                <p style={{ fontSize: 11, color: 'var(--t3)', margin: '6px 0 0' }}>
                  {loan.paid}/{loan.months} parcelas pagas · Início {new Date(loan.startDate).toLocaleDateString('pt-BR')}
                </p>
              </div>
            ))
          )}
        </div>
      )}

      {modal && <SimModal config={modal} onClose={() => setModal(null)} />}
    </div>
  )
}
