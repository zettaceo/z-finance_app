import React from 'react'
import { CheckCircle, XCircle, AlertTriangle, Info } from 'lucide-react'

const icons = {
  success: CheckCircle,
  error: XCircle,
  warning: AlertTriangle,
  info: Info,
}

const colors = {
  success: 'var(--accent)',
  error: '#F87171',
  warning: 'var(--gold)',
  info: '#60A5FA',
}

export default function Toast({ toasts }) {
  return (
    <div style={{
      position: 'fixed',
      top: 'calc(var(--topbar-h, 60px) + 8px)',
      right: 16,
      zIndex: 9999,
      display: 'flex',
      flexDirection: 'column',
      gap: 8,
      pointerEvents: 'none',
      maxWidth: 'calc(100vw - 32px)',
    }}>
      {toasts.map(t => {
        const Icon = icons[t.type] || Info
        const color = colors[t.type] || colors.info
        return (
          <div key={t.id} style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            background: 'var(--surface)',
            border: `1px solid ${color}40`,
            borderLeft: `3px solid ${color}`,
            borderRadius: 12,
            padding: '12px 16px',
            color: 'var(--t1)',
            fontSize: 14,
            fontWeight: 500,
            backdropFilter: 'blur(20px)',
            boxShadow: '0 8px 32px rgba(0,0,0,0.4)',
            animation: 'toastIn 0.3s ease',
            pointerEvents: 'all',
            maxWidth: 340,
          }}>
            <Icon size={18} style={{ color, flexShrink: 0 }} />
            <span>{t.msg}</span>
          </div>
        )
      })}
    </div>
  )
}
