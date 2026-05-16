import { fmtBRL, txTypeLabel, txDirection, relativeTime } from '../data/mock.js'
import {
  ArrowUpRight, ArrowDownLeft, RefreshCw, CreditCard,
  TrendingUp, TrendingDown, Banknote, RotateCcw
} from 'lucide-react'

const typeConfig = {
  PIX_OUT:     { Icon: ArrowUpRight,   color:'var(--red)',    bg:'rgba(239,68,68,0.1)',  dir:'out' },
  PIX_IN:      { Icon: ArrowDownLeft,  color:'var(--green)',  bg:'rgba(0,232,122,0.1)',  dir:'in' },
  TRANSFER_IN: { Icon: RefreshCw,      color:'var(--blue)',   bg:'rgba(59,130,246,0.1)', dir:'in' },
  PAYMENT:     { Icon: Banknote,       color:'var(--red)',    bg:'rgba(239,68,68,0.1)',  dir:'out' },
  CARD_AUTH:   { Icon: CreditCard,     color:'var(--purple)', bg:'rgba(139,92,246,0.1)', dir:'out' },
  TRADE_BUY:   { Icon: TrendingUp,     color:'var(--gold)',   bg:'rgba(245,158,11,0.1)', dir:'out' },
  TRADE_SELL:  { Icon: TrendingDown,   color:'var(--blue)',   bg:'rgba(59,130,246,0.1)', dir:'in' },
  DEPOSIT:     { Icon: ArrowDownLeft,  color:'var(--green)',  bg:'rgba(0,232,122,0.1)',  dir:'in' },
  WITHDRAWAL:  { Icon: ArrowUpRight,   color:'var(--red)',    bg:'rgba(239,68,68,0.1)',  dir:'out' },
  REVERSAL:    { Icon: RotateCcw,      color:'var(--gold)',   bg:'rgba(245,158,11,0.1)', dir:'neu' },
}

export default function ActivityFeed({ transactions, limit = 6, onViewAll }) {
  const items = (transactions || []).slice(0, limit)

  return (
    <div className="card" style={{ padding:22 }}>
      {/* Header */}
      <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between', marginBottom:16 }}>
        <div style={{ display:'flex', alignItems:'center', gap:10 }}>
          <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:15 }}>Movimentações</p>
          <div className="pulse-dot" style={{ width:6, height:6 }} />
        </div>
        <button onClick={onViewAll} style={{ fontSize:12, color:'var(--green)', fontWeight:600, background:'none', border:'none', cursor:'pointer' }}>
          Ver extrato →
        </button>
      </div>

      {/* Items */}
      {items.length === 0 ? (
        <EmptyState />
      ) : (
        <div>
          {items.map(tx => (
            <TxRow key={tx.id} tx={tx} />
          ))}
        </div>
      )}
    </div>
  )
}

function TxRow({ tx }) {
  const cfg = typeConfig[tx.type] || { Icon: RefreshCw, color:'var(--t2)', bg:'rgba(255,255,255,0.06)', dir:'neu' }
  const { Icon, color, bg } = cfg
  const isNeg = tx.amount < 0

  return (
    <div className="tx-item">
      <div style={{ width:36, height:36, borderRadius:10, background:bg, display:'flex', alignItems:'center', justifyContent:'center', flexShrink:0 }}>
        <Icon size={15} color={color} />
      </div>
      <div style={{ flex:1, minWidth:0 }}>
        <p style={{ fontSize:13, fontWeight:500, color:'var(--t1)', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap', marginBottom:2 }}>
          {tx.description}
        </p>
        <p style={{ fontSize:11, color:'var(--t3)' }}>{txTypeLabel(tx.type)} · {relativeTime(tx.createdAt)}</p>
      </div>
      <div style={{ textAlign:'right', flexShrink:0 }}>
        <p style={{ fontSize:13, fontWeight:700, color: isNeg ? 'var(--red)' : 'var(--green)', fontFamily:'Syne' }}>
          {isNeg ? '' : '+'}{fmtBRL(Math.abs(tx.amount))}
        </p>
        <span className={`badge ${tx.status === 'CONFIRMED' ? 'badge-green' : 'badge-gold'}`} style={{ fontSize:9 }}>
          {tx.status === 'CONFIRMED' ? 'Confirmado' : 'Pendente'}
        </span>
      </div>
    </div>
  )
}

function EmptyState() {
  return (
    <div style={{ padding:'24px 0', textAlign:'center' }}>
      <p style={{ fontSize:13, color:'var(--t3)' }}>Nenhuma movimentação recente</p>
      <p style={{ fontSize:11, color:'var(--t3)', marginTop:4 }}>Execute uma simulação para ver transações aqui</p>
    </div>
  )
}
