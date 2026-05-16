import React, { useState } from 'react'
import { Snowflake, Flame, Eye, EyeOff, CreditCard, Lock, Unlock, CheckCircle, XCircle, AlertCircle, Activity } from 'lucide-react'
import { useApp } from '../App.jsx'
import SimModal from '../components/SimModal.jsx'

function fmt(cents) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100)
}

const CARD_TIERS = {
  RETAIL: {
    label: 'Retail',
    gradient: 'linear-gradient(135deg, #1E293B 0%, #334155 50%, #475569 100%)',
    border: 'rgba(148,163,184,0.25)',
    shadow: 'none',
    textColor: '#94A3B8',
    accentColor: '#CBD5E1',
    chip: 'linear-gradient(135deg, #64748B, #475569)',
    shimmer: false,
  },
  BUSINESS: {
    label: 'Business',
    gradient: 'linear-gradient(135deg, #040C1B 0%, #081628 40%, #0C1E38 70%, #818CF820 100%)',
    border: 'rgba(129,140,248,0.3)',
    shadow: '0 20px 60px rgba(129,140,248,0.12)',
    textColor: '#E2E8F0',
    accentColor: '#818CF8',
    chip: 'linear-gradient(135deg, #D4A574, #C8954A)',
    shimmer: false,
  },
  INSTITUTIONAL: {
    label: 'Institutional',
    gradient: 'linear-gradient(135deg, #0A0A0A 0%, #141414 40%, #0D0D0D 70%, #FBBF2415 100%)',
    border: 'rgba(251,191,36,0.4)',
    shadow: '0 24px 80px rgba(251,191,36,0.15)',
    textColor: '#FBBF24',
    accentColor: '#FBBF24',
    chip: 'linear-gradient(135deg, #F59E0B, #D97706)',
    shimmer: true,
  },
}

export default function Cartoes() {
  const { store, dispatch, toast, modal, setModal } = useApp()
  const [cardVisible, setCardVisible] = useState(false)
  const card = store.card
  const persona = store.persona || 'BUSINESS'
  const tier = CARD_TIERS[persona] || CARD_TIERS.BUSINESS

  const usedPct = (card.limitUsed / card.limit) * 100

  const handleFreeze = () => {
    dispatch(card.frozen ? 'card_unfreeze' : 'card_freeze', { cardId: 'CARD-001' })
    toast(card.frozen ? 'Cartão desbloqueado!' : 'Cartão congelado!', card.frozen ? 'success' : 'warning')
  }

  const transactions = [
    { desc: 'Mercado Extra', amount: -8750, date: '15/05', category: 'Supermercado', status: 'confirmed' },
    { desc: 'Netflix', amount: -5500, date: '14/05', category: 'Streaming', status: 'confirmed' },
    { desc: 'Posto Ipiranga', amount: -24000, date: '13/05', category: 'Combustível', status: 'confirmed' },
    { desc: 'Amazon.com.br', amount: -18900, date: '12/05', category: 'E-commerce', status: 'confirmed' },
    { desc: 'iFood', amount: -7200, date: '11/05', category: 'Alimentação', status: 'pending' },
  ]

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <h2 style={{ fontSize: 22, fontWeight: 800, fontFamily: 'Syne,sans-serif', color: 'var(--t1)', margin: '0 0 4px' }}>Cartões</h2>
        <p style={{ fontSize: 13, color: 'var(--t3)', margin: 0 }}>Gestão de cartão JIT (Just-In-Time)</p>
      </div>

      {/* Card visual */}
      <div style={{
        position: 'relative', borderRadius: 24, padding: '28px 24px',
        background: card.frozen
          ? 'linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)'
          : tier.gradient,
        border: card.frozen ? '1px solid #4B5563' : `1px solid ${tier.border}`,
        marginBottom: 20,
        overflow: 'hidden',
        minHeight: 195,
        boxShadow: card.frozen ? 'none' : tier.shadow,
        transition: 'all 0.5s ease',
      }}>
        {/* Shimmer for Institutional */}
        {tier.shimmer && !card.frozen && (
          <div style={{
            position: 'absolute', top: 0, left: 0, right: 0, bottom: 0,
            background: 'linear-gradient(105deg, transparent 30%, rgba(251,191,36,0.18) 50%, transparent 70%)',
            animation: 'goldShimmer 3.5s ease-in-out infinite',
            pointerEvents: 'none', zIndex: 1,
          }} />
        )}

        {/* BG orbs */}
        <div style={{ position: 'absolute', top: -30, right: -30, width: 180, height: 180, borderRadius: '50%', background: card.frozen ? 'rgba(75,85,99,0.15)' : `${tier.accentColor}10`, pointerEvents: 'none' }} />
        <div style={{ position: 'absolute', bottom: -50, left: '30%', width: 150, height: 150, borderRadius: '50%', background: card.frozen ? 'rgba(75,85,99,0.08)' : `${tier.accentColor}06`, pointerEvents: 'none' }} />

        {/* Chip + brand */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 24, position: 'relative', zIndex: 2 }}>
          <div style={{
            width: 44, height: 34, borderRadius: 8,
            background: tier.chip,
            boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.2)',
          }}>
            <div style={{ height: '50%', borderBottom: '1px solid rgba(0,0,0,0.15)' }} />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 3 }}>
            {card.frozen && <Snowflake size={14} color="#60A5FA" />}
            <span style={{ fontFamily: 'Syne,sans-serif', fontWeight: 900, fontSize: 16, color: card.frozen ? '#9CA3AF' : tier.textColor, letterSpacing: '0.02em' }}>
              Z-Finance
            </span>
            <span style={{
              fontSize: 9, fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.12em',
              color: card.frozen ? '#6B7280' : tier.accentColor,
              background: card.frozen ? 'rgba(107,114,128,0.15)' : `${tier.accentColor}20`,
              borderRadius: 5, padding: '2px 6px',
            }}>
              {tier.label}
            </span>
          </div>
        </div>

        {/* Number */}
        <p style={{
          fontFamily: 'DM Mono, monospace', fontSize: 19, letterSpacing: '0.18em',
          color: card.frozen ? '#6B7280' : tier.textColor, margin: '0 0 20px',
          position: 'relative', zIndex: 2,
        }}>
          {cardVisible ? '5412 7539 4821 8421' : card.number}
        </p>

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', position: 'relative', zIndex: 2 }}>
          <div>
            <p style={{ fontSize: 10, color: 'rgba(255,255,255,0.4)', textTransform: 'uppercase', letterSpacing: '0.06em', margin: '0 0 2px' }}>Titular</p>
            <p style={{ fontSize: 13, fontWeight: 700, color: card.frozen ? '#9CA3AF' : tier.textColor, margin: 0, letterSpacing: '0.05em' }}>
              {store.user.name.toUpperCase()}
            </p>
          </div>
          <div style={{ textAlign: 'right' }}>
            <p style={{ fontSize: 10, color: 'rgba(255,255,255,0.4)', textTransform: 'uppercase', letterSpacing: '0.06em', margin: '0 0 2px' }}>Validade</p>
            <p style={{ fontSize: 13, fontFamily: 'DM Mono, monospace', color: card.frozen ? '#9CA3AF' : tier.textColor, margin: 0 }}>
              {cardVisible ? card.exp : '••/••'}
            </p>
          </div>
        </div>

        {card.frozen && (
          <div style={{
            position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center',
            background: 'rgba(4,12,27,0.5)', backdropFilter: 'blur(3px)', borderRadius: 24, zIndex: 3,
          }}>
            <div style={{ textAlign: 'center' }}>
              <Snowflake size={36} color="#60A5FA" style={{ marginBottom: 8 }} />
              <p style={{ color: '#60A5FA', fontWeight: 700, fontSize: 16, margin: 0 }}>Cartão Congelado</p>
            </div>
          </div>
        )}
      </div>

      {/* Card actions */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10, marginBottom: 20 }}>
        <button onClick={() => setCardVisible(v => !v)} style={{
          padding: '14px 8px', borderRadius: 14, border: '1px solid var(--border)',
          background: 'var(--surface)', cursor: 'pointer',
          display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6,
        }}>
          {cardVisible ? <EyeOff size={20} color="var(--t2)" /> : <Eye size={20} color="var(--t2)" />}
          <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--t2)' }}>{cardVisible ? 'Ocultar' : 'Ver dados'}</span>
        </button>
        <button onClick={handleFreeze} style={{
          padding: '14px 8px', borderRadius: 14, border: `1px solid ${card.frozen ? 'rgba(96,165,250,0.4)' : 'var(--border)'}`,
          background: card.frozen ? 'rgba(96,165,250,0.1)' : 'var(--surface)', cursor: 'pointer',
          display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6,
        }}>
          {card.frozen
            ? <Flame size={20} color="#F59E0B" />
            : <Snowflake size={20} color="#60A5FA" />}
          <span style={{ fontSize: 11, fontWeight: 600, color: card.frozen ? '#F59E0B' : '#60A5FA' }}>
            {card.frozen ? 'Desbloquear' : 'Congelar'}
          </span>
        </button>
        <button onClick={() => setModal({
          title: 'Autorizar JIT', action: 'card_authorize', successMsg: 'Transação autorizada!',
          description: 'Just-In-Time: autorize transações sob demanda.',
          fields: [
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 15000 },
            { key: 'merchant', label: 'Estabelecimento', type: 'text', default: 'Amazon Brasil' },
            { key: 'mcc', label: 'MCC', type: 'text', default: '5411', hint: '5411 = Supermercado' },
          ],
          submitLabel: 'Autorizar',
        })} style={{
          padding: '14px 8px', borderRadius: 14, border: '1px solid var(--border)',
          background: 'var(--surface)', cursor: 'pointer',
          display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6,
        }}>
          <Activity size={20} color="var(--accent)" />
          <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--t2)' }}>Autorizar JIT</span>
        </button>
      </div>

      {/* Limit bar */}
      <div style={{
        background: 'var(--surface)', border: '1px solid var(--border)',
        borderRadius: 18, padding: '20px', marginBottom: 20,
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
          <div>
            <p style={{ fontSize: 12, color: 'var(--t3)', margin: '0 0 2px', fontWeight: 600, textTransform: 'uppercase' }}>Limite utilizado</p>
            <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 18, fontWeight: 700, color: 'var(--t1)', margin: 0 }}>
              {fmt(card.limitUsed)}
            </p>
          </div>
          <div style={{ textAlign: 'right' }}>
            <p style={{ fontSize: 12, color: 'var(--t3)', margin: '0 0 2px', fontWeight: 600, textTransform: 'uppercase' }}>Limite total</p>
            <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 18, fontWeight: 700, color: 'var(--t1)', margin: 0 }}>
              {fmt(card.limit)}
            </p>
          </div>
        </div>
        <div style={{ height: 8, borderRadius: 4, background: 'var(--surface-2)', overflow: 'hidden' }}>
          <div style={{
            height: '100%', borderRadius: 4,
            background: usedPct > 80
              ? 'linear-gradient(90deg, #F59E0B, #EF4444)'
              : 'linear-gradient(90deg, var(--accent), var(--accent-2))',
            width: `${usedPct}%`,
            transition: 'width 0.5s ease',
          }} />
        </div>
        <p style={{ fontSize: 12, color: 'var(--t3)', marginTop: 8 }}>
          {fmt(card.limit - card.limitUsed)} disponível · {usedPct.toFixed(0)}% utilizado
        </p>
      </div>

      {/* JIT Controls */}
      <div style={{
        background: 'var(--surface)', border: '1px solid var(--border)',
        borderRadius: 18, overflow: 'hidden', marginBottom: 20,
      }}>
        <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)' }}>
          <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: 0 }}>Controles JIT</p>
          <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0, marginTop: 2 }}>Autorização Just-In-Time de transações</p>
        </div>
        {[
          { label: 'Confirmar Transação', sub: 'Aprovar compra pendente', action: 'card_confirm', color: 'var(--accent)', icon: CheckCircle },
          { label: 'Rejeitar Transação', sub: 'Negar compra pendente', action: 'card_reject', color: '#F87171', icon: XCircle },
        ].map(({ label, sub, action, color, icon: Icon }) => (
          <button key={action} onClick={() => setModal({
            title: label, action, successMsg: `${label} realizada!`,
            fields: [
              { key: 'transactionId', label: 'ID da transação', type: 'text', default: 'TXN-' + Date.now().toString().slice(-6) },
              { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 15000 },
            ],
            submitLabel: label,
          })} style={{
            width: '100%', display: 'flex', alignItems: 'center', gap: 14, padding: '14px 20px',
            border: 'none', background: 'transparent', cursor: 'pointer', textAlign: 'left',
            borderBottom: action === 'card_confirm' ? '1px solid var(--border)' : 'none',
          }}>
            <Icon size={20} style={{ color, flexShrink: 0 }} />
            <div style={{ flex: 1 }}>
              <p style={{ fontWeight: 600, fontSize: 14, color: 'var(--t1)', margin: 0 }}>{label}</p>
              <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{sub}</p>
            </div>
            <span style={{ color: 'var(--t3)', fontSize: 18 }}>→</span>
          </button>
        ))}
      </div>

      {/* Recent card transactions */}
      <div style={{
        background: 'var(--surface)', border: '1px solid var(--border)',
        borderRadius: 18, overflow: 'hidden',
      }}>
        <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)' }}>
          <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: 0 }}>Lançamentos recentes</p>
        </div>
        {transactions.map((tx, i) => (
          <div key={i} style={{
            display: 'flex', alignItems: 'center', gap: 14, padding: '14px 20px',
            borderBottom: i < transactions.length - 1 ? '1px solid var(--border)' : 'none',
          }}>
            <div style={{ flex: 1 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 2 }}>
                <p style={{ fontWeight: 600, fontSize: 14, color: 'var(--t1)', margin: 0 }}>{tx.desc}</p>
                {tx.status === 'pending' && (
                  <span style={{ fontSize: 10, color: '#F59E0B', background: 'rgba(245,158,11,0.1)', borderRadius: 4, padding: '1px 6px', fontWeight: 700 }}>Pendente</span>
                )}
              </div>
              <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{tx.date} · {tx.category}</p>
            </div>
            <span style={{ fontFamily: 'DM Mono, monospace', fontSize: 14, fontWeight: 700, color: 'var(--t1)' }}>
              {fmt(Math.abs(tx.amount))}
            </span>
          </div>
        ))}
      </div>

      {modal && <SimModal config={modal} onClose={() => setModal(null)} />}
    </div>
  )
}
