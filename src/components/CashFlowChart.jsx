import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { fmtBRL } from '../data/mock.js'

function CustomTooltip({ active, payload, label }) {
  if (!active || !payload?.length) return null
  return (
    <div style={{ background:'var(--card2)', border:'1px solid var(--border)', borderRadius:10, padding:'10px 14px', fontSize:12, boxShadow:'0 8px 24px rgba(0,0,0,0.4)' }}>
      <p style={{ fontWeight:700, marginBottom:6, color:'var(--t1)' }}>{label}</p>
      {payload.map(p => (
        <div key={p.name} style={{ display:'flex', alignItems:'center', gap:6, marginBottom:3 }}>
          <div style={{ width:8, height:8, borderRadius:2, background:p.color }} />
          <span style={{ color:'var(--t2)' }}>{p.name === 'income' ? 'Entradas' : 'Saídas'}:</span>
          <span style={{ fontWeight:600, color:'var(--t1)' }}>{fmtBRL(p.value)}</span>
        </div>
      ))}
    </div>
  )
}

export default function CashFlowChart({ cashFlow, period = '7d' }) {
  const data = (cashFlow || []).map(d => ({
    day:      d.day,
    income:   d.income / 100,
    expenses: d.expenses / 100,
    net:      (d.income - d.expenses) / 100,
  }))

  const totalIncome   = cashFlow?.reduce((a,c) => a + c.income, 0) || 0
  const totalExpenses = cashFlow?.reduce((a,c) => a + c.expenses, 0) || 0
  const netFlow       = totalIncome - totalExpenses

  return (
    <div className="card" style={{ padding:22 }}>
      {/* Header */}
      <div style={{ display:'flex', alignItems:'flex-start', justifyContent:'space-between', marginBottom:20 }}>
        <div>
          <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:15, marginBottom:4 }}>Fluxo Financeiro</p>
          <div style={{ display:'flex', alignItems:'center', gap:6 }}>
            <span style={{
              fontFamily:'Syne', fontWeight:800, fontSize:22,
              color: netFlow >= 0 ? 'var(--green)' : 'var(--red)',
            }}>
              {netFlow >= 0 ? '+' : ''}{fmtBRL(netFlow * 100)}
            </span>
            <span style={{ fontSize:12, color:'var(--t3)' }}>esta semana</span>
          </div>
        </div>
        <div style={{ display:'flex', gap:6 }}>
          {['7d','30d','90d'].map(p => (
            <button key={p} style={{
              padding:'4px 10px', borderRadius:8, border:'1px solid var(--border)',
              background: p === period ? 'rgba(0,232,122,0.1)' : 'transparent',
              color: p === period ? 'var(--green)' : 'var(--t3)',
              fontSize:11, fontWeight:600, cursor:'pointer',
            }}>{p}</button>
          ))}
        </div>
      </div>

      {/* Summary pills */}
      <div style={{ display:'flex', gap:10, marginBottom:20 }}>
        <div style={{ flex:1, background:'rgba(0,232,122,0.05)', border:'1px solid rgba(0,232,122,0.15)', borderRadius:10, padding:'10px 14px' }}>
          <p style={{ fontSize:10, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.08em', marginBottom:4 }}>Total Entradas</p>
          <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:15, color:'var(--green)' }}>+{fmtBRL(totalIncome)}</p>
        </div>
        <div style={{ flex:1, background:'rgba(239,68,68,0.05)', border:'1px solid rgba(239,68,68,0.15)', borderRadius:10, padding:'10px 14px' }}>
          <p style={{ fontSize:10, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.08em', marginBottom:4 }}>Total Saídas</p>
          <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:15, color:'var(--red)' }}>-{fmtBRL(totalExpenses)}</p>
        </div>
      </div>

      {/* Chart */}
      <div style={{ height:180 }}>
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top:5, right:0, left:0, bottom:0 }}>
            <defs>
              <linearGradient id="incGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%"  stopColor="#00e87a" stopOpacity={0.3}/>
                <stop offset="95%" stopColor="#00e87a" stopOpacity={0}/>
              </linearGradient>
              <linearGradient id="expGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%"  stopColor="#ef4444" stopOpacity={0.25}/>
                <stop offset="95%" stopColor="#ef4444" stopOpacity={0}/>
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" vertical={false} />
            <XAxis dataKey="day" tick={{ fill:'var(--t3)', fontSize:11 }} axisLine={false} tickLine={false} />
            <YAxis hide />
            <Tooltip content={<CustomTooltip />} />
            <Area type="monotone" dataKey="income"   stroke="#00e87a" strokeWidth={2} fill="url(#incGrad)" dot={false} />
            <Area type="monotone" dataKey="expenses" stroke="#ef4444" strokeWidth={2} fill="url(#expGrad)" dot={false} />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}
