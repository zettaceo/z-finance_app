import { useEffect, useState } from 'react'
import { fmtBRL, relativeTime, txTypeLabel } from '../data/mock.js'

export default function RiskGauge({ store }) {
  const [angle, setAngle] = useState(-90)

  const tx     = store?.transactions?.[0]
  const kyc    = store?.kyc
  const used   = kyc?.used?.daily || 0
  const limit  = kyc?.limits?.daily || 10000000
  const ratio  = Math.min(used / limit, 1)

  // 0 = -90deg (seguro), 1 = +90deg (crítico)
  const targetAngle = -90 + ratio * 180

  const riskLabel = ratio < 0.3 ? 'Seguro' : ratio < 0.6 ? 'Normal' : ratio < 0.8 ? 'Elevado' : 'Crítico'
  const riskColor = ratio < 0.3 ? 'var(--green)' : ratio < 0.6 ? 'var(--gold)' : 'var(--red)'

  useEffect(() => {
    const t = setTimeout(() => setAngle(targetAngle), 300)
    return () => clearTimeout(t)
  }, [targetAngle])

  return (
    <div className="card hover-glow" style={{ padding:22, display:'flex', flexDirection:'column', gap:14 }}>
      {/* Last activity */}
      <div>
        <p style={{ fontSize:10, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.1em', marginBottom:10 }}>Última Atividade</p>
        {tx ? (
          <>
            <p style={{ fontSize:13, fontWeight:600, marginBottom:2, overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>
              {tx.description}
            </p>
            <div style={{ display:'flex', justifyContent:'space-between', fontSize:11, color:'var(--t3)' }}>
              <span>{relativeTime(tx.createdAt)}</span>
              <span style={{ fontWeight:600, color: tx.amount >= 0 ? 'var(--green)' : 'var(--red)' }}>
                {tx.amount >= 0 ? '+' : ''}{fmtBRL(Math.abs(tx.amount))}
              </span>
            </div>
          </>
        ) : (
          <p style={{ fontSize:13, color:'var(--t3)' }}>Nenhuma atividade</p>
        )}
      </div>

      <div className="divider" />

      {/* Risk label */}
      <div>
        <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:12 }}>
          <p style={{ fontSize:10, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.1em' }}>Nível de Risco</p>
          <span style={{ fontFamily:'Syne', fontWeight:800, fontSize:18, color:riskColor }}>{riskLabel}</span>
        </div>

        {/* SVG Gauge */}
        <div style={{ display:'flex', justifyContent:'center' }}>
          <svg viewBox="0 0 160 90" style={{ width:140, height:80 }}>
            <defs>
              <linearGradient id="gaugeGrad" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%"   stopColor="#00e87a" />
                <stop offset="45%"  stopColor="#f59e0b" />
                <stop offset="100%" stopColor="#ef4444" />
              </linearGradient>
            </defs>
            {/* Track */}
            <path d="M 15 80 A 65 65 0 0 1 145 80" fill="none" stroke="rgba(255,255,255,0.06)" strokeWidth="10" strokeLinecap="round" />
            {/* Fill */}
            <path d="M 15 80 A 65 65 0 0 1 145 80" fill="none" stroke="url(#gaugeGrad)" strokeWidth="10" strokeLinecap="round"
              strokeDasharray="204" strokeDashoffset={204 * (1 - ratio)} opacity={0.9}
              style={{ transition:'stroke-dashoffset 1.2s cubic-bezier(0.34,1.56,0.64,1)' }}
            />
            {/* Needle */}
            <line
              x1="80" y1="80" x2="74" y2="20"
              stroke="white" strokeWidth="2" strokeLinecap="round"
              style={{ transformOrigin:'80px 80px', transform:`rotate(${angle}deg)`, transition:'transform 1.2s cubic-bezier(0.34,1.56,0.64,1)' }}
            />
            <circle cx="80" cy="80" r="5" fill="white" opacity={0.9} />
          </svg>
        </div>

        <div style={{ display:'flex', justifyContent:'space-between', fontSize:10, color:'var(--t3)' }}>
          <span>Seguro</span>
          <span>Crítico</span>
        </div>
      </div>

      {/* KYC Limit usage */}
      <div>
        <div style={{ display:'flex', justifyContent:'space-between', fontSize:11, marginBottom:6 }}>
          <span style={{ color:'var(--t3)' }}>Uso diário</span>
          <span style={{ color:'var(--t2)', fontWeight:600 }}>{Math.round(ratio * 100)}% de {fmtBRL(limit)}</span>
        </div>
        <div className="progress-track">
          <div className="progress-fill" style={{ width:`${ratio * 100}%`, background: ratio > 0.8 ? 'var(--red)' : ratio > 0.5 ? 'var(--gold)' : undefined }} />
        </div>
      </div>
    </div>
  )
}
