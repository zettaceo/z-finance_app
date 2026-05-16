import React from 'react'
import { useApp } from '../App.jsx'

export const PERSONA_CFG = {
  RETAIL: {
    label: 'Retail',
    color: '#00E599',
    bgColor: 'rgba(0,229,153,0.12)',
    desc: 'Conta pessoal simplificada',
  },
  BUSINESS: {
    label: 'Business',
    color: '#818CF8',
    bgColor: 'rgba(129,140,248,0.12)',
    desc: 'Conta empresarial completa',
  },
  INSTITUTIONAL: {
    label: 'Institutional',
    color: '#FBBF24',
    bgColor: 'rgba(251,191,36,0.12)',
    desc: 'Plataforma treasury global',
  },
}

export default function PersonaSwitcher() {
  const { store, dispatch } = useApp()
  const current = store.persona || 'BUSINESS'

  return (
    <div style={{ padding: '12px 16px 14px', borderBottom: '1px solid var(--border)' }}>
      <p style={{
        fontSize: 10, fontWeight: 700, color: 'var(--t3)',
        textTransform: 'uppercase', letterSpacing: '0.08em', margin: '0 0 8px',
      }}>
        Perfil de Acesso
      </p>
      <div style={{
        display: 'flex', background: 'var(--bg)', borderRadius: 10, padding: 3,
        border: '1px solid var(--border)',
      }}>
        {Object.entries(PERSONA_CFG).map(([id, p]) => {
          const active = current === id
          return (
            <button
              key={id}
              onClick={() => dispatch('persona_switch', { persona: id })}
              title={p.desc}
              style={{
                flex: 1, padding: '7px 2px', borderRadius: 8, border: 'none', cursor: 'pointer',
                background: active ? p.color : 'transparent',
                color: active ? '#040C1B' : 'var(--t3)',
                fontSize: 10, fontWeight: 700, letterSpacing: '0.03em',
                transition: 'all 0.18s ease',
                whiteSpace: 'nowrap',
                boxShadow: active ? `0 2px 8px ${p.color}40` : 'none',
              }}
            >
              {p.label.toUpperCase()}
            </button>
          )
        })}
      </div>
    </div>
  )
}
