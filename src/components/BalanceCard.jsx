import { useState, useEffect, useRef } from 'react'
import { Eye, EyeOff, TrendingUp } from 'lucide-react'
import { fmtBRL, cryptoBalanceBRL } from '../data/mock.js'

function useCountUp(target, duration = 1000) {
  const [value, setValue] = useState(0)
  const startRef = useRef(null)

  useEffect(() => {
    let raf
    const start = performance.now()
    const animate = (now) => {
      const progress = Math.min((now - start) / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      setValue(Math.round(target * eased))
      if (progress < 1) raf = requestAnimationFrame(animate)
    }
    raf = requestAnimationFrame(animate)
    return () => cancelAnimationFrame(raf)
  }, [target, duration])

  return value
}

export default function BalanceCard({ store }) {
  const [hidden, setHidden] = useState(false)

  const available = store?.account?.balance || 0
  const holds     = store?.account?.holds   || 0
  const cdb       = store?.investments?.cdb || 0
  const tesouro   = store?.investments?.tesouro || 0
  const cryptoBRL = cryptoBalanceBRL(store?.crypto || [])
  const total     = available + cdb + tesouro + Math.round(cryptoBRL)

  const animTotal = useCountUp(total / 100, 1200)

  const bars = [
    { label:'Disponível',   value: available,            color:'var(--green)',  pct: total ? (available / total * 100) : 0 },
    { label:'Investimentos', value: cdb + tesouro,       color:'var(--blue)',   pct: total ? ((cdb+tesouro) / total * 100) : 0 },
    { label:'Cripto',        value: Math.round(cryptoBRL), color:'var(--gold)', pct: total ? (cryptoBRL / total * 100) : 0 },
    { label:'Bloqueado',     value: holds,               color:'var(--t3)',     pct: total ? (holds / total * 100) : 0 },
  ]

  return (
    <div className="card hover-glow" style={{ padding:24, flex:1 }}>
      {/* Header */}
      <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between', marginBottom:20 }}>
        <div>
          <p style={{ fontSize:11, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.1em', marginBottom:4 }}>Patrimônio Total</p>
          <div style={{ display:'flex', alignItems:'center', gap:6 }}>
            <div className="pulse-dot" />
            <span style={{ fontSize:11, color:'var(--green)' }}>Ao vivo</span>
          </div>
        </div>
        <button className="btn-icon" onClick={() => setHidden(!hidden)}>
          {hidden ? <EyeOff size={14}/> : <Eye size={14}/>}
        </button>
      </div>

      {/* Big Number */}
      <div style={{ marginBottom:24 }}>
        {hidden ? (
          <div style={{ fontSize:36, fontFamily:'Syne', fontWeight:800, letterSpacing:'0.04em', color:'var(--t3)' }}>••••••</div>
        ) : (
          <div className="number-animate" style={{ display:'flex', alignItems:'baseline', gap:6 }}>
            <span style={{ fontSize:14, fontWeight:600, color:'var(--t2)', fontFamily:'Syne' }}>R$</span>
            <span style={{
              fontSize:38, fontFamily:'Syne', fontWeight:800,
              background:'linear-gradient(135deg, #fff 20%, var(--green) 80%)',
              WebkitBackgroundClip:'text', WebkitTextFillColor:'transparent',
              backgroundClip:'text',
              filter:'drop-shadow(0 0 20px rgba(0,232,122,0.3))',
            }}>
              {animTotal.toLocaleString('pt-BR', { minimumFractionDigits:2, maximumFractionDigits:2 })}
            </span>
          </div>
        )}
        <div style={{ display:'flex', alignItems:'center', gap:6, marginTop:6 }}>
          <TrendingUp size={12} color='var(--green)' />
          <span style={{ fontSize:12, color:'var(--green)', fontWeight:600 }}>+3.2% esta semana</span>
          <span style={{ fontSize:12, color:'var(--t3)' }}>vs semana anterior</span>
        </div>
      </div>

      {/* Stacked bar */}
      <div style={{ display:'flex', height:6, borderRadius:999, overflow:'hidden', gap:1, marginBottom:18 }}>
        {bars.map(b => b.pct > 0 && (
          <div key={b.label} style={{ width:`${b.pct}%`, background:b.color, transition:'width 1s ease', borderRadius:999 }} />
        ))}
      </div>

      {/* Breakdown */}
      <div style={{ display:'flex', flexDirection:'column', gap:10 }}>
        {bars.map(b => (
          <div key={b.label} style={{ display:'flex', alignItems:'center', justifyContent:'space-between' }}>
            <div style={{ display:'flex', alignItems:'center', gap:8 }}>
              <div style={{ width:8, height:8, borderRadius:2, background:b.color, flexShrink:0 }} />
              <span style={{ fontSize:12, color:'var(--t2)' }}>{b.label}</span>
            </div>
            <span style={{ fontSize:12, fontWeight:600, color: hidden ? 'var(--t3)' : 'var(--t1)', fontFamily: hidden ? 'sans-serif' : 'Syne' }}>
              {hidden ? '••••' : fmtBRL(b.value)}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
