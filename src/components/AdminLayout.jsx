import React from 'react'
import { Crown, LogOut, ShieldAlert } from 'lucide-react'
import { useApp } from '../App.jsx'

export default function AdminLayout({ children }) {
  const { logout, store } = useApp()

  return (
    <div style={{ minHeight: '100dvh', background: '#0B0F1A', color: '#F1F5F9' }}>
      {/* Admin topbar */}
      <header style={{
        background: '#0D1321',
        borderBottom: '1px solid rgba(251,191,36,0.15)',
        padding: '0 20px',
        height: 60,
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        position: 'sticky',
        top: 0,
        zIndex: 50,
      }}>
        {/* Crown badge */}
        <div style={{
          width: 36, height: 36, borderRadius: 10,
          background: 'rgba(251,191,36,0.12)',
          border: '1px solid rgba(251,191,36,0.25)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          flexShrink: 0,
        }}>
          <Crown size={17} color="#FBBF24" />
        </div>

        <div>
          <p style={{ fontFamily: 'Syne, sans-serif', fontWeight: 800, fontSize: 15, color: '#F1F5F9', margin: 0, lineHeight: 1.2 }}>Admin Console</p>
          <p style={{ fontSize: 10, color: '#475569', margin: 0, letterSpacing: '0.04em', textTransform: 'uppercase', fontWeight: 600 }}>Z-Finance · Backoffice</p>
        </div>

        {/* Security badge */}
        <div style={{
          display: 'flex', alignItems: 'center', gap: 5,
          background: 'rgba(248,113,113,0.08)',
          border: '1px solid rgba(248,113,113,0.2)',
          borderRadius: 8, padding: '4px 10px',
          marginLeft: 8,
        }}>
          <ShieldAlert size={12} color="#F87171" />
          <span style={{ fontSize: 10, fontWeight: 700, color: '#F87171', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Acesso restrito</span>
        </div>

        {/* Right side */}
        <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ fontSize: 12, color: '#64748B', fontFamily: 'DM Mono, monospace' }}>
            {store.user.name}
          </span>
          <button
            onClick={logout}
            style={{
              display: 'flex', alignItems: 'center', gap: 6,
              padding: '7px 14px', borderRadius: 8, border: 'none',
              background: 'rgba(248,113,113,0.08)',
              color: '#F87171', fontSize: 13, fontWeight: 600, cursor: 'pointer',
              transition: 'all 0.15s ease',
            }}
          >
            <LogOut size={14} />
            Sair
          </button>
        </div>
      </header>

      {/* Content */}
      <main style={{ maxWidth: 900, margin: '0 auto', padding: '24px 20px 48px' }}>
        {children}
      </main>
    </div>
  )
}
