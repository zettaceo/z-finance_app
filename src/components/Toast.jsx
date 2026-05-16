import { CheckCircle, AlertCircle, Info, X } from 'lucide-react'

const icons = {
  success: <CheckCircle size={16} color='var(--green)' />,
  error:   <AlertCircle size={16} color='var(--red)' />,
  info:    <Info size={16} color='var(--blue)' />,
}

const colors = {
  success: 'rgba(0,232,122,0.15)',
  error:   'rgba(239,68,68,0.15)',
  info:    'rgba(59,130,246,0.15)',
}

export default function Toast({ message, type = 'success' }) {
  return (
    <div className="toast" style={{ borderLeft: `3px solid ${type === 'success' ? 'var(--green)' : type === 'error' ? 'var(--red)' : 'var(--blue)'}` }}>
      <div style={{ display:'flex', alignItems:'flex-start', gap:10 }}>
        <div style={{ marginTop:1, flexShrink:0 }}>{icons[type]}</div>
        <span style={{ fontSize:13, color:'var(--t1)', lineHeight:1.4 }}>{message}</span>
      </div>
    </div>
  )
}
