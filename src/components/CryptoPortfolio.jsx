import { fmtBRL } from '../data/mock.js'

const COIN_COLORS = {
  BTC:  { color:'#f7931a', bg:'rgba(247,147,26,0.12)',  letter:'₿' },
  ETH:  { color:'#627eea', bg:'rgba(98,126,234,0.12)',  letter:'Ξ' },
  USDT: { color:'#26a17b', bg:'rgba(38,161,123,0.12)',  letter:'₮' },
  SOL:  { color:'#9945ff', bg:'rgba(153,69,255,0.12)',  letter:'◎' },
}

function Sparkline({ history, positive }) {
  if (!history || history.length < 2) return null
  const min = Math.min(...history)
  const max = Math.max(...history)
  const range = max - min || 1
  const w = 72, h = 28
  const pts = history.map((v, i) => ({
    x: (i / (history.length - 1)) * w,
    y: h - ((v - min) / range) * (h - 4) - 2,
  }))
  const path = pts.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ')
  const area = `${path} L ${w} ${h} L 0 ${h} Z`
  const color = positive ? '#00e87a' : '#ef4444'

  return (
    <svg width={w} height={h} viewBox={`0 0 ${w} ${h}`} className="sparkline">
      <defs>
        <linearGradient id={`grad-${history[0]}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%"   stopColor={color} stopOpacity={0.25} />
          <stop offset="100%" stopColor={color} stopOpacity={0} />
        </linearGradient>
      </defs>
      <path d={area} fill={`url(#grad-${history[0]})`} />
      <path d={path}  fill="none" stroke={color} strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

export default function CryptoPortfolio({ crypto, onViewAll }) {
  const totalBRL = (crypto || []).reduce((a, c) => a + c.amount * c.price * 100, 0)

  return (
    <div className="card" style={{ padding:22, display:'flex', flexDirection:'column' }}>
      {/* Header */}
      <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between', marginBottom:16 }}>
        <div>
          <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:15 }}>Carteira Cripto</p>
          <p style={{ fontSize:12, color:'var(--t2)', marginTop:2 }}>≈ {fmtBRL(totalBRL)}</p>
        </div>
        <button onClick={onViewAll} style={{ fontSize:12, color:'var(--green)', fontWeight:600, background:'none', border:'none', cursor:'pointer' }}>
          Ver tudo →
        </button>
      </div>

      {/* Coins */}
      <div style={{ flex:1 }}>
        {(crypto || []).map(coin => {
          const cfg      = COIN_COLORS[coin.symbol] || { color:'var(--t2)', bg:'rgba(255,255,255,0.06)', letter:'?' }
          const brlValue = coin.amount * coin.price * 100
          const isPos    = coin.change24h >= 0

          return (
            <div key={coin.symbol} className="coin-item hover-glow" style={{ borderRadius:8, cursor:'default', transition:'all 0.2s', padding:'10px 0' }}
              onMouseEnter={e => e.currentTarget.style.background='rgba(255,255,255,0.02)'}
              onMouseLeave={e => e.currentTarget.style.background='transparent'}>
              {/* Icon */}
              <div style={{ width:38, height:38, borderRadius:'50%', background:cfg.bg, border:`1px solid ${cfg.color}33`, display:'flex', alignItems:'center', justifyContent:'center', flexShrink:0, fontSize:16, color:cfg.color, fontWeight:900 }}>
                {cfg.letter}
              </div>

              {/* Info */}
              <div style={{ flex:1, minWidth:0 }}>
                <p style={{ fontSize:13, fontWeight:600, fontFamily:'Syne', marginBottom:2 }}>{coin.name}</p>
                <p style={{ fontSize:11, color:'var(--t3)' }}>
                  {coin.amount.toLocaleString('pt-BR', { minimumFractionDigits:4, maximumFractionDigits:4 })} {coin.symbol}
                </p>
              </div>

              {/* Sparkline */}
              <div style={{ flexShrink:0 }}>
                <Sparkline history={coin.history} positive={isPos} />
              </div>

              {/* Values */}
              <div style={{ textAlign:'right', flexShrink:0, minWidth:70 }}>
                <p style={{ fontSize:12, fontWeight:700, fontFamily:'Syne', marginBottom:2 }}>{fmtBRL(brlValue)}</p>
                <span className={isPos ? 'badge badge-green' : 'badge badge-red'} style={{ fontSize:9 }}>
                  {isPos ? '+' : ''}{coin.change24h.toFixed(1)}%
                </span>
              </div>
            </div>
          )
        })}
      </div>

      {/* View button */}
      <button onClick={onViewAll} className="btn-ghost" style={{ width:'100%', marginTop:14, fontSize:12 }}>
        Ver Carteira Completa
      </button>
    </div>
  )
}
