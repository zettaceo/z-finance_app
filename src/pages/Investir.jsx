import React, { useState } from 'react'
import { TrendingUp, TrendingDown, ArrowUpDown, Zap, DollarSign, RefreshCw } from 'lucide-react'
import { AreaChart, Area, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { useApp } from '../App.jsx'
import SimModal from '../components/SimModal.jsx'

function fmt(cents, cur = 'BRL') {
  const val = cents / 100
  if (cur === 'BRL') return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val)
  if (cur === 'USD') return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val)
  return val.toFixed(2)
}

const CRYPTO_COLORS = { BTC: '#F7931A', ETH: '#627EEA', SOL: '#9945FF', USDT: '#26A17B' }
const CRYPTO_NAMES = { BTC: 'Bitcoin', ETH: 'Ethereum', SOL: 'Solana', USDT: 'Tether' }

export default function Investir() {
  const { store, modal, setModal } = useApp()
  const [selectedCrypto, setSelectedCrypto] = useState('BTC')
  const [tab, setTab] = useState('portfolio') // portfolio | swap | pricing

  const selected = store.crypto.find(c => c.symbol === selectedCrypto)
  const totalCrypto = store.crypto.reduce((s, c) => s + c.amount * c.price, 0)
  const totalInvest = store.accounts.invest.balance + totalCrypto

  return (
    <div>
      <div style={{ marginBottom: 20 }}>
        <h2 style={{ fontSize: 22, fontWeight: 800, fontFamily: 'Syne,sans-serif', color: 'var(--t1)', margin: '0 0 4px' }}>Investimentos</h2>
        <p style={{ fontSize: 13, color: 'var(--t3)', margin: 0 }}>Criptoativos, swap e engine de precificação</p>
      </div>

      {/* Portfolio summary */}
      <div style={{
        background: 'linear-gradient(135deg, #0A1928, #081628)',
        border: '1px solid var(--border)',
        borderRadius: 20, padding: '20px',
        marginBottom: 20,
        position: 'relative', overflow: 'hidden',
      }}>
        <div style={{ position: 'absolute', top: -30, right: -30, width: 150, height: 150, borderRadius: '50%', background: 'radial-gradient(circle, rgba(251,191,36,0.1) 0%, transparent 70%)', pointerEvents: 'none' }} />
        <p style={{ fontSize: 12, color: 'var(--t3)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 4 }}>Portfólio Total</p>
        <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 32, fontWeight: 700, color: 'var(--t1)', margin: '0 0 4px' }}>
          {fmt(totalInvest)}
        </p>
        <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
          <div>
            <p style={{ fontSize: 11, color: 'var(--t3)', margin: '0 0 2px' }}>Renda variável</p>
            <p style={{ fontSize: 14, fontWeight: 700, color: 'var(--gold)', margin: 0, fontFamily: 'DM Mono, monospace' }}>{fmt(store.accounts.invest.balance)}</p>
          </div>
          <div>
            <p style={{ fontSize: 11, color: 'var(--t3)', margin: '0 0 2px' }}>Criptoativos</p>
            <p style={{ fontSize: 14, fontWeight: 700, color: '#818CF8', margin: 0, fontFamily: 'DM Mono, monospace' }}>{fmt(totalCrypto)}</p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 8, marginBottom: 20 }}>
        {[
          { id: 'portfolio', label: 'Portfólio' },
          { id: 'swap', label: 'Swap' },
          { id: 'pricing', label: 'Precificação' },
        ].map(t => (
          <button key={t.id} onClick={() => setTab(t.id)} style={{
            padding: '8px 18px', borderRadius: 20, cursor: 'pointer',
            background: tab === t.id ? 'rgba(129,140,248,0.2)' : 'var(--surface)',
            border: `1px solid ${tab === t.id ? 'rgba(129,140,248,0.5)' : 'var(--border)'}`,
            color: tab === t.id ? '#818CF8' : 'var(--t2)',
            fontSize: 13, fontWeight: 700, transition: 'all 0.15s',
          }}>{t.label}</button>
        ))}
      </div>

      {tab === 'portfolio' && (
        <>
          {/* Crypto selector */}
          <div style={{ display: 'flex', gap: 8, overflowX: 'auto', paddingBottom: 4, marginBottom: 16 }} className="hide-scroll">
            {store.crypto.map(c => {
              const color = CRYPTO_COLORS[c.symbol]
              const active = selectedCrypto === c.symbol
              return (
                <button key={c.symbol} onClick={() => setSelectedCrypto(c.symbol)} style={{
                  flexShrink: 0, display: 'flex', alignItems: 'center', gap: 8,
                  padding: '8px 14px', borderRadius: 14, cursor: 'pointer',
                  background: active ? `${color}20` : 'var(--surface)',
                  border: `1px solid ${active ? `${color}60` : 'var(--border)'}`,
                  transition: 'all 0.15s',
                }}>
                  <div style={{ width: 24, height: 24, borderRadius: 8, background: `${color}20`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 11, fontWeight: 700, color }}>
                    {c.symbol.charAt(0)}
                  </div>
                  <span style={{ fontSize: 13, fontWeight: 700, color: active ? color : 'var(--t2)' }}>{c.symbol}</span>
                </button>
              )
            })}
          </div>

          {/* Selected crypto detail */}
          {selected && (
            <div style={{
              background: 'var(--surface)', border: '1px solid var(--border)',
              borderRadius: 20, padding: '20px', marginBottom: 20,
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 16 }}>
                <div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                    <div style={{
                      width: 36, height: 36, borderRadius: 10,
                      background: `${CRYPTO_COLORS[selected.symbol]}20`,
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      fontSize: 14, fontWeight: 700, color: CRYPTO_COLORS[selected.symbol],
                    }}>
                      {selected.symbol.charAt(0)}
                    </div>
                    <div>
                      <p style={{ fontWeight: 700, fontSize: 16, color: 'var(--t1)', margin: 0 }}>{CRYPTO_NAMES[selected.symbol]}</p>
                      <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{selected.symbol}</p>
                    </div>
                  </div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 20, fontWeight: 700, color: 'var(--t1)', margin: '0 0 4px' }}>
                    {fmt(selected.amount * selected.price)}
                  </p>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 4, justifyContent: 'flex-end' }}>
                    {selected.change24h >= 0
                      ? <TrendingUp size={14} color="var(--accent)" />
                      : <TrendingDown size={14} color="#F87171" />}
                    <span style={{
                      fontSize: 13, fontWeight: 700,
                      color: selected.change24h >= 0 ? 'var(--accent)' : '#F87171',
                    }}>
                      {selected.change24h >= 0 ? '+' : ''}{selected.change24h.toFixed(2)}%
                    </span>
                  </div>
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 16 }}>
                <div style={{ background: 'var(--surface-2)', borderRadius: 12, padding: '12px' }}>
                  <p style={{ fontSize: 11, color: 'var(--t3)', margin: '0 0 4px', textTransform: 'uppercase' }}>Saldo</p>
                  <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 14, fontWeight: 700, color: 'var(--t1)', margin: 0 }}>
                    {selected.amount} {selected.symbol}
                  </p>
                </div>
                <div style={{ background: 'var(--surface-2)', borderRadius: 12, padding: '12px' }}>
                  <p style={{ fontSize: 11, color: 'var(--t3)', margin: '0 0 4px', textTransform: 'uppercase' }}>Cotação</p>
                  <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 14, fontWeight: 700, color: 'var(--t1)', margin: 0 }}>
                    {fmt(selected.price)}
                  </p>
                </div>
              </div>

              {/* Sparkline full */}
              <div style={{ height: 80 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={selected.history.map((v, i) => ({ v, i }))}>
                    <defs>
                      <linearGradient id={`grad${selected.symbol}`} x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor={CRYPTO_COLORS[selected.symbol]} stopOpacity={0.3} />
                        <stop offset="100%" stopColor={CRYPTO_COLORS[selected.symbol]} stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <Tooltip
                      contentStyle={{ background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 8, fontSize: 12, color: 'var(--t1)' }}
                      formatter={v => [fmt(v * 100), 'Preço']}
                    />
                    <Area type="monotone" dataKey="v" stroke={CRYPTO_COLORS[selected.symbol]} strokeWidth={2} fill={`url(#grad${selected.symbol})`} dot={false} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>

              {/* Actions */}
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginTop: 16 }}>
                <button onClick={() => setModal({
                  title: `Liquidar ${selected.symbol}`, action: 'crypto_liquidate', successMsg: 'Liquidação realizada!',
                  description: `Converta ${selected.symbol} em BRL automaticamente.`,
                  fields: [
                    { key: 'symbol', label: 'Ativo', type: 'text', default: selected.symbol },
                    { key: 'amount', label: 'Quantidade', type: 'number', default: selected.amount / 2, step: 0.001 },
                  ],
                  submitLabel: 'Liquidar',
                })} style={{
                  padding: '12px', borderRadius: 12, border: '1px solid #F87171',
                  background: 'rgba(248,113,113,0.1)', cursor: 'pointer',
                  color: '#F87171', fontSize: 13, fontWeight: 700,
                }}>
                  Liquidar
                </button>
                <button onClick={() => setModal({
                  title: `Pagar com ${selected.symbol}`, action: 'crypto_pay', successMsg: 'Pagamento crypto enviado!',
                  fields: [
                    { key: 'symbol', label: 'Ativo', type: 'text', default: selected.symbol },
                    { key: 'amount', label: 'Quantidade', type: 'number', default: 0.01, step: 0.001 },
                    { key: 'recipient', label: 'Endereço destino', type: 'text', default: '0x742d35Cc6634C0532925a3b844Bc454e4438f44e' },
                  ],
                  submitLabel: 'Enviar',
                })} style={{
                  padding: '12px', borderRadius: 12, border: '1px solid var(--accent)',
                  background: 'rgba(0,229,153,0.1)', cursor: 'pointer',
                  color: 'var(--accent)', fontSize: 13, fontWeight: 700,
                }}>
                  Enviar
                </button>
              </div>
            </div>
          )}
        </>
      )}

      {tab === 'swap' && (
        <div style={{
          background: 'var(--surface)', border: '1px solid var(--border)',
          borderRadius: 20, padding: '20px',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 20 }}>
            <ArrowUpDown size={20} color="#818CF8" />
            <p style={{ fontWeight: 700, fontSize: 16, color: 'var(--t1)', margin: 0 }}>Crypto Swap</p>
          </div>

          {[
            {
              label: 'BTC → ETH', sub: 'Bitcoin para Ethereum',
              action: () => setModal({
                title: 'Swap BTC → ETH', action: 'crypto_swap', successMsg: 'Swap realizado!',
                description: 'Troca instantânea com melhor taxa do mercado.',
                fields: [
                  { key: 'fromSymbol', label: 'De', type: 'text', default: 'BTC' },
                  { key: 'toSymbol', label: 'Para', type: 'text', default: 'ETH' },
                  { key: 'fromAmount', label: 'Quantidade BTC', type: 'number', default: 0.01, step: 0.0001 },
                ],
                submitLabel: 'Executar swap',
              })
            },
            {
              label: 'ETH → USDT', sub: 'Ethereum para Tether',
              action: () => setModal({
                title: 'Swap ETH → USDT', action: 'crypto_swap', successMsg: 'Swap realizado!',
                fields: [
                  { key: 'fromSymbol', label: 'De', type: 'text', default: 'ETH' },
                  { key: 'toSymbol', label: 'Para', type: 'text', default: 'USDT' },
                  { key: 'fromAmount', label: 'Quantidade ETH', type: 'number', default: 0.1, step: 0.01 },
                ],
                submitLabel: 'Executar swap',
              })
            },
            {
              label: 'USDT → BRL', sub: 'Tether para Reais',
              action: () => setModal({
                title: 'USDT → BRL', action: 'crypto_swap', successMsg: 'Conversão realizada!',
                fields: [
                  { key: 'fromSymbol', label: 'De', type: 'text', default: 'USDT' },
                  { key: 'toSymbol', label: 'Para', type: 'text', default: 'BRL' },
                  { key: 'fromAmount', label: 'Quantidade USDT', type: 'number', default: 100, step: 1 },
                ],
                submitLabel: 'Converter',
              })
            },
          ].map(({ label, sub, action }) => (
            <button key={label} onClick={action} style={{
              width: '100%', display: 'flex', alignItems: 'center', gap: 14, padding: '16px',
              borderRadius: 14, border: '1px solid var(--border)', background: 'var(--surface-2)',
              cursor: 'pointer', textAlign: 'left', marginBottom: 10,
              transition: 'all 0.15s',
            }}
            onMouseEnter={e => { e.currentTarget.style.borderColor = 'rgba(129,140,248,0.4)' }}
            onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--border)' }}
            >
              <div style={{ width: 40, height: 40, borderRadius: 12, background: 'rgba(129,140,248,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <ArrowUpDown size={18} color="#818CF8" />
              </div>
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{sub}</p>
              </div>
              <span style={{ color: 'var(--t3)' }}>→</span>
            </button>
          ))}
        </div>
      )}

      {tab === 'pricing' && (
        <div style={{ display: 'grid', gap: 12 }}>
          {[
            { label: 'Obter Cotação', sub: 'Quote de qualquer par', icon: DollarSign, action: 'pricing_quote',
              fields: [
                { key: 'from', label: 'De', type: 'select', options: [{ value: 'BRL', label: 'BRL' }, { value: 'USD', label: 'USD' }, { value: 'USDT', label: 'USDT' }, { value: 'BTC', label: 'BTC' }] },
                { key: 'to', label: 'Para', type: 'select', options: [{ value: 'USD', label: 'USD' }, { value: 'BRL', label: 'BRL' }, { value: 'BTC', label: 'BTC' }, { value: 'ETH', label: 'ETH' }] },
                { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 100000 },
              ]
            },
            { label: 'Atualizar Preço', sub: 'Update de cotação de ativo', icon: RefreshCw, action: 'pricing_update',
              fields: [
                { key: 'symbol', label: 'Ativo', type: 'select', options: [{ value: 'BTC', label: 'BTC' }, { value: 'ETH', label: 'ETH' }, { value: 'SOL', label: 'SOL' }, { value: 'USDT', label: 'USDT' }] },
                { key: 'price', label: 'Novo preço (centavos)', type: 'number', default: 350000 },
              ]
            },
          ].map(({ label, sub, icon: Icon, action, fields }) => (
            <button key={label} onClick={() => setModal({ title: label, action, successMsg: `${label} executada!`, fields, submitLabel: 'Executar' })} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '20px',
              borderRadius: 18, border: '1px solid var(--border)', background: 'var(--surface)',
              cursor: 'pointer', textAlign: 'left', transition: 'all 0.15s',
            }}>
              <div style={{ width: 44, height: 44, borderRadius: 14, background: 'rgba(251,191,36,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Icon size={20} color="var(--gold)" />
              </div>
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 15, color: 'var(--t1)', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 13, color: 'var(--t3)', margin: 0 }}>{sub}</p>
              </div>
              <span style={{ color: 'var(--t3)' }}>→</span>
            </button>
          ))}

          {/* Market rates */}
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: 0 }}>Câmbio ao vivo</p>
            </div>
            {Object.entries(store.market).map(([cur, { rate }], i, arr) => (
              <div key={cur} style={{
                display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                padding: '14px 20px', borderBottom: i < arr.length - 1 ? '1px solid var(--border)' : 'none',
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <span style={{ fontSize: 18 }}>{cur === 'USD' ? '🇺🇸' : cur === 'EUR' ? '🇪🇺' : '🇬🇧'}</span>
                  <div>
                    <p style={{ fontWeight: 600, fontSize: 14, color: 'var(--t1)', margin: 0 }}>BRL/{cur}</p>
                    <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>Compra</p>
                  </div>
                </div>
                <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 16, fontWeight: 700, color: 'var(--gold)', margin: 0 }}>
                  R$ {(rate / 100).toFixed(2)}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}

      {modal && <SimModal config={modal} onClose={() => setModal(null)} />}
    </div>
  )
}
