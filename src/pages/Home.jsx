import React, { useState, useMemo } from 'react'
import { Eye, EyeOff, ArrowUpRight, ArrowDownLeft, RefreshCw, Send, QrCode, Wifi, Activity, Building2, Globe, Lock, Landmark, ChevronRight, ArrowUpCircle } from 'lucide-react'
import { AreaChart, Area, ResponsiveContainer, Tooltip, XAxis } from 'recharts'
import { useApp } from '../App.jsx'
import SimModal from '../components/SimModal.jsx'
import { PERSONA_CFG } from '../components/PersonaSwitcher.jsx'

function fmt(cents, cur = 'BRL') {
  const val = cents / 100
  if (cur === 'BRL') return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val)
  if (cur === 'USD') return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val)
  if (cur === 'AED') return `AED ${val.toLocaleString('en-AE', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
  return val.toFixed(2)
}

function fmtCrypto(amount, symbol) {
  if (symbol === 'BTC') return `${amount.toFixed(4)} BTC`
  if (symbol === 'ETH') return `${amount.toFixed(3)} ETH`
  return `${amount.toFixed(2)} ${symbol}`
}

const TX_ICONS = { pix: '⚡', payment: '📄', crypto: '₿', transfer: '↔', card: '💳', invoice: '🧾' }

const CRYPTO_COLORS = { BTC: '#F7931A', ETH: '#627EEA', SOL: '#9945FF', USDT: '#26A17B' }

export default function Home() {
  const { store, modal, setModal } = useApp()
  const [hidden, setHidden] = useState(false)
  const [activeAccount, setActiveAccount] = useState('main')

  const persona = store.persona || 'BUSINESS'
  const pc = PERSONA_CFG[persona]
  const isRetail = persona === 'RETAIL'
  const isInstitutional = persona === 'INSTITUTIONAL'

  const acc = store.accounts[activeAccount]
  const totalBRL = store.accounts.main.balance + store.accounts.invest.balance
  const cryptoTotal = store.crypto.reduce((s, c) => s + c.amount * c.price, 0)
  const cryptoBRL = store.crypto.reduce((s, c) => s + c.amount * c.price * 100, 0)
  const totalWealth = totalBRL + cryptoTotal

  const dailyLimitDisplay = isInstitutional
    ? 'Ilimitado'
    : isRetail
      ? fmt(500000)
      : fmt(store.kyc.limits.daily)
  const dailyLimitValue = isRetail ? 500000 : store.kyc.limits.daily
  const dailyUsed = store.kyc.used?.daily || store.kyc.limits?.used?.daily || 1250000

  const baseActions = [
    { label: 'PIX', icon: Send, locked: false, action: () => setModal({
      title: 'Enviar PIX', action: 'pix_send',
      successMsg: 'PIX enviado com sucesso!',
      description: 'Transferência instantânea via PIX',
      fields: [
        { key: 'key', label: 'Chave PIX', type: 'text', placeholder: 'CPF, e-mail, telefone ou chave aleatória' },
        { key: 'amount', label: 'Valor (centavos)', type: 'number', placeholder: '10000', default: 10000, hint: 'R$ 100,00 = 10000 centavos' },
        { key: 'description', label: 'Descrição', type: 'text', placeholder: 'Opcional', default: 'Transferência via PIX' },
      ],
      submitLabel: 'Enviar PIX',
    })},
    { label: 'Cobrar', icon: QrCode, locked: false, action: () => setModal({
      title: 'Gerar Cobrança', action: 'invoice_create',
      successMsg: 'Cobrança criada!',
      description: 'Crie uma cobrança por PIX ou USDT',
      fields: [
        { key: 'amount', label: 'Valor (centavos)', type: 'number', placeholder: '50000', default: 50000 },
        { key: 'currency', label: 'Moeda', type: 'select', options: [{ value: 'BRL', label: 'BRL (PIX)' }, { value: 'USDT', label: 'USDT (Crypto)' }] },
        { key: 'description', label: 'Descrição', type: 'text', placeholder: 'Ex: Serviço de consultoria', default: 'Cobrança Z-Finance' },
      ],
      submitLabel: 'Criar cobrança',
    })},
    { label: 'Câmbio', icon: RefreshCw, locked: isRetail, action: isRetail ? null : () => setModal({
      title: 'Câmbio de Moeda', action: 'transfer_internal',
      successMsg: 'Câmbio realizado!',
      description: `Taxas: USD ${(store.market.USD.rate / 100).toFixed(2)} | EUR ${(store.market.EUR.rate / 100).toFixed(2)}`,
      fields: [
        { key: 'fromAccount', label: 'De', type: 'select', options: [{ value: 'main', label: 'BRL' }, { value: 'usd', label: 'USD' }] },
        { key: 'toAccount', label: 'Para', type: 'select', options: [{ value: 'usd', label: 'USD' }, { value: 'main', label: 'BRL' }] },
        { key: 'amount', label: 'Valor (centavos origem)', type: 'number', default: 100000 },
      ],
      submitLabel: 'Converter',
    })},
    ...(isInstitutional
      ? [{ label: 'Treasury', icon: Landmark, locked: false, action: () => setModal({
          title: 'Operação Treasury', action: 'transfer_internal',
          successMsg: 'Operação executada!',
          description: 'Movimentação entre contas treasury',
          fields: [
            { key: 'fromAccount', label: 'Origem', type: 'select', options: [{ value: 'main', label: 'BRL' }, { value: 'usd', label: 'USD' }] },
            { key: 'toAccount', label: 'Destino', type: 'select', options: [{ value: 'usd', label: 'USD' }, { value: 'main', label: 'BRL' }] },
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 1000000 },
          ],
          submitLabel: 'Executar',
        })}]
      : [{ label: 'Depositar', icon: ArrowDownLeft, locked: false, action: () => setModal({
          title: 'Simular Depósito', action: 'pix_receive',
          successMsg: 'Depósito creditado!',
          fields: [
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 100000 },
            { key: 'payerName', label: 'Remetente', type: 'text', default: 'Cliente Externo' },
          ],
          submitLabel: 'Receber',
        })}]
    ),
  ]
  const quickActions = baseActions

  const recentTx = useMemo(() => store.transactions.slice(0, 8), [store.transactions])

  return (
    <div>
      {/* Balance hero card */}
      <div style={{
        background: 'linear-gradient(135deg, var(--surface) 0%, #0A1928 100%)',
        border: '1px solid var(--border)',
        borderRadius: 24,
        padding: '24px 20px',
        marginBottom: 20,
        position: 'relative',
        overflow: 'hidden',
      }}>
        {/* Decorative */}
        <div style={{
          position: 'absolute', top: -40, right: -40,
          width: 200, height: 200, borderRadius: '50%',
          background: `radial-gradient(circle, ${pc.color}10 0%, transparent 70%)`,
          pointerEvents: 'none',
        }} />

        {/* Persona badge */}
        <div style={{ position: 'absolute', top: 16, right: 16 }}>
          <span style={{
            fontSize: 10, fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.08em',
            color: pc.color, background: pc.bgColor,
            border: `1px solid ${pc.color}30`,
            borderRadius: 8, padding: '3px 8px',
          }}>
            {pc.label}
          </span>
        </div>

        {/* Account tabs */}
        <div style={{ display: 'flex', gap: 8, marginBottom: 20, overflowX: 'auto', paddingBottom: 2 }} className="hide-scroll">
          {[
            { id: 'main',   label: 'BRL',   color: 'var(--accent)', flag: '🇧🇷' },
            { id: 'usd',    label: 'USD',   color: '#60A5FA',       flag: '🇺🇸' },
            { id: 'aed',    label: 'AED',   color: '#FBBF24',       flag: '🇦🇪' },
            { id: 'invest', label: 'Invest', color: 'var(--gold)',  flag: '📈' },
          ].map(({ id, label, color, flag }) => (
            <button key={id} onClick={() => setActiveAccount(id)} style={{
              flexShrink: 0, padding: '6px 12px', borderRadius: 20, cursor: 'pointer',
              background: activeAccount === id ? `${color}20` : 'transparent',
              border: `1px solid ${activeAccount === id ? `${color}50` : 'var(--border)'}`,
              color: activeAccount === id ? color : 'var(--t3)',
              fontSize: 12, fontWeight: 700,
              transition: 'all 0.15s', display: 'flex', alignItems: 'center', gap: 4,
            }}>
              <span style={{ fontSize: 13 }}>{flag}</span>
              {label}
            </button>
          ))}
        </div>

        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16 }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <p style={{ fontSize: 12, color: 'var(--t3)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 6 }}>
              Saldo disponível
            </p>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, flexWrap: 'wrap' }}>
              <span style={{
                fontFamily: 'DM Mono, monospace',
                fontSize: 'clamp(24px, 7vw, 40px)',
                fontWeight: 700,
                color: 'var(--t1)',
                letterSpacing: '-1px',
              }}>
                {hidden ? '•••••••' : fmt(acc.balance, acc.currency)}
              </span>
            </div>
            <p style={{ fontSize: 13, color: 'var(--t3)', marginTop: 6 }}>
              Patrimônio total: {hidden ? '•••••' : fmt(totalBRL + cryptoTotal, 'BRL')}
            </p>
          </div>
          <button onClick={() => setHidden(h => !h)} style={{
            width: 40, height: 40, borderRadius: 12, border: 'none',
            background: 'var(--surface-2)', cursor: 'pointer',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: 'var(--t2)', flexShrink: 0,
          }}>
            {hidden ? <Eye size={17} /> : <EyeOff size={17} />}
          </button>
        </div>

        {/* KYC limit bar */}
        <div style={{ marginTop: 20 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
            <span style={{ fontSize: 11, color: 'var(--t3)' }}>Limite diário PIX</span>
            <span style={{ fontSize: 11, color: 'var(--t2)', fontFamily: 'DM Mono, monospace' }}>
              {isInstitutional ? (
                <span style={{ color: pc.color }}>Ilimitado</span>
              ) : (
                `${fmt(dailyUsed)} / ${dailyLimitDisplay}`
              )}
            </span>
          </div>
          {isInstitutional ? (
            <div style={{ height: 4, borderRadius: 2, background: `linear-gradient(90deg, ${pc.color}, ${pc.color}80)` }} />
          ) : (
            <div style={{ height: 4, borderRadius: 2, background: 'var(--surface-2)', overflow: 'hidden' }}>
              <div style={{
                height: '100%', borderRadius: 2,
                background: `linear-gradient(90deg, ${pc.color}, ${pc.color}90)`,
                width: `${Math.min((dailyUsed / dailyLimitValue) * 100, 100)}%`,
                transition: 'width 0.5s ease',
              }} />
            </div>
          )}
        </div>
      </div>

      {/* Quick actions */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(4, 1fr)',
        gap: 10,
        marginBottom: 24,
      }}>
        {quickActions.map(({ label, icon: Icon, action, locked }) => (
          <button key={label} onClick={locked ? undefined : action} style={{
            display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 8,
            padding: '14px 8px', borderRadius: 16,
            border: locked ? '1px dashed var(--border)' : '1px solid var(--border)',
            background: 'var(--surface)', cursor: locked ? 'not-allowed' : 'pointer',
            transition: 'all 0.15s ease',
            minHeight: 72, opacity: locked ? 0.5 : 1,
            position: 'relative',
          }}
          onMouseEnter={e => { if (!locked) { e.currentTarget.style.borderColor = `${pc.color}50`; e.currentTarget.style.background = `${pc.color}08` } }}
          onMouseLeave={e => { e.currentTarget.style.borderColor = locked ? 'var(--border)' : 'var(--border)'; e.currentTarget.style.background = 'var(--surface)' }}
          >
            <div style={{
              width: 36, height: 36, borderRadius: 10,
              background: locked ? 'rgba(255,255,255,0.04)' : `${pc.color}18`,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              position: 'relative',
            }}>
              {locked ? <Lock size={15} color="var(--t3)" /> : <Icon size={17} color={pc.color} />}
            </div>
            <span style={{ fontSize: 11, fontWeight: 600, color: locked ? 'var(--t3)' : 'var(--t2)', textAlign: 'center', lineHeight: 1.2 }}>{label}</span>
            {locked && (
              <span style={{
                position: 'absolute', top: 4, right: 4,
                fontSize: 8, fontWeight: 700, color: '#FBBF24',
                background: 'rgba(251,191,36,0.15)', borderRadius: 4, padding: '1px 4px',
              }}>PRO</span>
            )}
          </button>
        ))}
      </div>

      {/* Treasury Panel — Institutional only */}
      {isInstitutional && (
        <div style={{
          background: 'linear-gradient(135deg, #0D1F0D 0%, #0A1628 100%)',
          border: `1px solid ${pc.color}30`,
          borderRadius: 20, padding: '20px', marginBottom: 20,
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <div>
              <p style={{ fontWeight: 700, fontSize: 15, color: 'var(--t1)', margin: '0 0 2px' }}>Treasury Overview</p>
              <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>Posição consolidada global</p>
            </div>
            <span style={{
              fontSize: 11, fontWeight: 700, color: pc.color,
              background: pc.bgColor, borderRadius: 8, padding: '4px 10px',
            }}>AUM</span>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10, marginBottom: 16 }}>
            {[
              { label: 'BRL', value: fmt(store.accounts.main.balance + store.accounts.invest.balance), icon: '🇧🇷', color: 'var(--accent)' },
              { label: 'USD', value: `$${(store.accounts.usd.balance / 100).toLocaleString('en-US', { minimumFractionDigits: 2 })}`, icon: '🇺🇸', color: '#60A5FA' },
              { label: 'Crypto', value: fmt(cryptoBRL), icon: '₿', color: '#F7931A' },
            ].map(({ label, value, icon, color }) => (
              <div key={label} style={{
                background: 'rgba(255,255,255,0.04)', borderRadius: 14, padding: '12px 10px', textAlign: 'center',
              }}>
                <div style={{ fontSize: 16, marginBottom: 4 }}>{icon}</div>
                <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 12, fontWeight: 700, color, margin: '0 0 2px' }}>{hidden ? '•••' : value}</p>
                <p style={{ fontSize: 10, color: 'var(--t3)', margin: 0 }}>{label}</p>
              </div>
            ))}
          </div>
          <div style={{ borderTop: `1px solid ${pc.color}20`, paddingTop: 14 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: 12, color: 'var(--t3)' }}>AUM Total (BRL equiv.)</span>
              <span style={{ fontFamily: 'DM Mono, monospace', fontSize: 14, fontWeight: 700, color: pc.color }}>
                {hidden ? '•••••••' : fmt(totalWealth * 100)}
              </span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 8 }}>
              <span style={{ fontSize: 12, color: 'var(--t3)' }}>Crédito disponível</span>
              <span style={{ fontFamily: 'DM Mono, monospace', fontSize: 14, fontWeight: 700, color: '#60A5FA' }}>
                {hidden ? '•••••' : fmt(5000000)}
              </span>
            </div>
          </div>
        </div>
      )}

      {/* Cash flow chart */}
      <div style={{
        background: 'var(--surface)', border: '1px solid var(--border)',
        borderRadius: 20, padding: '20px', marginBottom: 20,
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <div>
            <p style={{ fontWeight: 700, fontSize: 15, color: 'var(--t1)', margin: '0 0 2px' }}>Fluxo de Caixa</p>
            <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>Últimos 7 dias</p>
          </div>
          <span style={{
            fontSize: 11, fontWeight: 700, color: 'var(--accent)',
            background: 'rgba(0,229,153,0.1)', borderRadius: 8, padding: '4px 10px',
          }}>7D</span>
        </div>
        <ResponsiveContainer width="100%" height={120}>
          <AreaChart data={store.cashFlow} margin={{ top: 4, right: 0, bottom: 0, left: 0 }}>
            <defs>
              <linearGradient id="cfGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="var(--accent)" stopOpacity={0.3} />
                <stop offset="100%" stopColor="var(--accent)" stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis dataKey="day" tick={{ fill: 'var(--t3)', fontSize: 10 }} axisLine={false} tickLine={false} />
            <Tooltip
              contentStyle={{ background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 8, color: 'var(--t1)', fontSize: 12 }}
              formatter={v => [fmt(v * 100), '']}
            />
            <Area type="monotone" dataKey="in" stroke="var(--accent)" strokeWidth={2} fill="url(#cfGrad)" dot={false} />
            <Area type="monotone" dataKey="out" stroke="#F87171" strokeWidth={1.5} fill="none" strokeDasharray="3 3" dot={false} />
          </AreaChart>
        </ResponsiveContainer>
        <div style={{ display: 'flex', gap: 16, marginTop: 12 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <div style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--accent)' }} />
            <span style={{ fontSize: 11, color: 'var(--t3)' }}>Entradas</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <div style={{ width: 8, height: 8, borderRadius: '50%', background: '#F87171' }} />
            <span style={{ fontSize: 11, color: 'var(--t3)' }}>Saídas</span>
          </div>
        </div>
      </div>

      {/* Crypto scroll row */}
      <div style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
          <p style={{ fontWeight: 700, fontSize: 15, color: 'var(--t1)', margin: 0 }}>Criptoativos</p>
          <span style={{ fontSize: 12, color: 'var(--accent)', fontWeight: 600 }}>
            {hidden ? '•••••' : fmt(cryptoTotal)}
          </span>
        </div>
        <div style={{ display: 'flex', gap: 12, overflowX: 'auto', paddingBottom: 8, scrollSnapType: 'x mandatory' }} className="hide-scroll">
          {store.crypto.map(c => {
            const color = CRYPTO_COLORS[c.symbol] || 'var(--accent)'
            const posValue = c.amount * c.price
            const upDown = c.change24h >= 0
            return (
              <div key={c.symbol} style={{
                flexShrink: 0, width: 160,
                background: 'var(--surface)', border: '1px solid var(--border)',
                borderRadius: 18, padding: '16px',
                scrollSnapAlign: 'start',
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <div style={{
                      width: 32, height: 32, borderRadius: 10,
                      background: `${color}20`,
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      fontSize: 14, fontWeight: 700, color,
                    }}>{c.symbol.charAt(0)}</div>
                    <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--t1)' }}>{c.symbol}</span>
                  </div>
                  <span style={{
                    fontSize: 10, fontWeight: 700,
                    color: upDown ? 'var(--accent)' : '#F87171',
                    background: upDown ? 'rgba(0,229,153,0.1)' : 'rgba(248,113,113,0.1)',
                    borderRadius: 6, padding: '2px 6px',
                  }}>
                    {upDown ? '+' : ''}{c.change24h.toFixed(1)}%
                  </span>
                </div>
                <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 15, fontWeight: 700, color: 'var(--t1)', margin: '0 0 2px' }}>
                  {hidden ? '•••' : fmt(posValue)}
                </p>
                <p style={{ fontSize: 11, color: 'var(--t3)', margin: 0 }}>{fmtCrypto(c.amount, c.symbol)}</p>
                {/* Mini sparkline */}
                <div style={{ height: 32, marginTop: 8 }}>
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={c.history.map((v, i) => ({ v, i }))}>
                      <defs>
                        <linearGradient id={`cg${c.symbol}`} x1="0" y1="0" x2="0" y2="1">
                          <stop offset="0%" stopColor={color} stopOpacity={0.3} />
                          <stop offset="100%" stopColor={color} stopOpacity={0} />
                        </linearGradient>
                      </defs>
                      <Area type="monotone" dataKey="v" stroke={color} strokeWidth={1.5} fill={`url(#cg${c.symbol})`} dot={false} />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Recent transactions */}
      <div style={{
        background: 'var(--surface)', border: '1px solid var(--border)',
        borderRadius: 20, overflow: 'hidden', marginBottom: 16,
      }}>
        <div style={{ padding: '16px 20px 12px', borderBottom: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <p style={{ fontWeight: 700, fontSize: 15, color: 'var(--t1)', margin: 0 }}>Transações Recentes</p>
          <span style={{ fontSize: 12, color: 'var(--t3)' }}>{store.transactions.length} total</span>
        </div>
        {recentTx.map((tx, i) => {
          const isCredit = tx.type === 'credit' || tx.amount > 0
          return (
            <div key={tx.id || i} style={{
              display: 'flex', alignItems: 'center', gap: 14,
              padding: '14px 20px',
              borderBottom: i < recentTx.length - 1 ? '1px solid var(--border)' : 'none',
            }}>
              <div style={{
                width: 40, height: 40, borderRadius: 12, flexShrink: 0,
                background: 'var(--surface-2)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: 17,
              }}>
                {TX_ICONS[tx.category] || '💰'}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <p style={{ fontWeight: 600, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                  {tx.description}
                </p>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{tx.date} • {tx.category}</p>
              </div>
              <span style={{
                fontFamily: 'DM Mono, monospace',
                fontSize: 14, fontWeight: 700,
                color: tx.amount > 0 ? 'var(--accent)' : 'var(--t1)',
                flexShrink: 0,
              }}>
                {tx.amount > 0 ? '+' : ''}{fmt(Math.abs(tx.amount))}
              </span>
            </div>
          )
        })}
      </div>

      {/* Observability status — Business & Institutional only */}
      {!isRetail && (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gap: 10,
          marginBottom: 8,
        }}>
          {[
            { label: 'Uptime', value: store.observability.uptime, color: 'var(--accent)', icon: Wifi },
            { label: 'Latência P99', value: `${store.observability.latencyP99}ms`, color: '#60A5FA', icon: Activity },
            { label: 'Webhooks', value: `${store.observability.pendingWebhooks} pend.`, color: 'var(--gold)', icon: RefreshCw },
          ].map(({ label, value, color, icon: Icon }) => (
            <div key={label} style={{
              background: 'var(--surface)', border: '1px solid var(--border)',
              borderRadius: 14, padding: '12px 10px', textAlign: 'center',
            }}>
              <Icon size={14} style={{ color, marginBottom: 6 }} />
              <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 13, fontWeight: 700, color: 'var(--t1)', margin: '0 0 2px' }}>{value}</p>
              <p style={{ fontSize: 10, color: 'var(--t3)', margin: 0 }}>{label}</p>
            </div>
          ))}
        </div>
      )}

      {/* Upgrade banner — Retail only */}
      {isRetail && (
        <div style={{
          background: 'linear-gradient(135deg, rgba(129,140,248,0.12), rgba(0,229,153,0.08))',
          border: '1px solid rgba(129,140,248,0.25)',
          borderRadius: 18, padding: '18px 20px', marginBottom: 8,
          display: 'flex', alignItems: 'center', gap: 14,
        }}>
          <div style={{
            width: 42, height: 42, borderRadius: 12, flexShrink: 0,
            background: 'rgba(129,140,248,0.15)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <ArrowUpCircle size={22} color="#818CF8" />
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 3px' }}>
              Desbloqueie o plano Business
            </p>
            <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0, lineHeight: 1.4 }}>
              Câmbio multi-moeda, PIX Internacional, observabilidade e mais
            </p>
          </div>
          <ChevronRight size={18} color="var(--t3)" style={{ flexShrink: 0 }} />
        </div>
      )}

      {modal && <SimModal config={modal} onClose={() => setModal(null)} />}
    </div>
  )
}
