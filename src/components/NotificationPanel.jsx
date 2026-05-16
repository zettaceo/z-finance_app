import React from 'react'
import { X, Bell, CheckCheck, Trash2 } from 'lucide-react'
import { useApp } from '../App.jsx'

function relTime(iso) {
  if (!iso) return ''
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1)  return 'agora'
  if (m < 60) return `${m}min atrás`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h atrás`
  return `${Math.floor(h / 24)}d atrás`
}

const TYPE_COLORS = {
  tx:       { color: '#34D399', bg: 'rgba(52,211,153,0.1)' },
  market:   { color: '#60A5FA', bg: 'rgba(96,165,250,0.1)' },
  credit:   { color: '#FBBF24', bg: 'rgba(251,191,36,0.1)' },
  security: { color: '#F59E0B', bg: 'rgba(245,158,11,0.1)' },
  kyc:      { color: '#818CF8', bg: 'rgba(129,140,248,0.1)' },
}

export default function NotificationPanel({ onClose }) {
  const { store, dispatch } = useApp()
  const notifs = store.notifications || []
  const unreadCount = notifs.filter(n => !n.read).length

  return (
    <>
      {/* Backdrop */}
      <div
        onClick={onClose}
        style={{
          position: 'fixed', inset: 0, zIndex: 299,
          background: 'rgba(4,12,27,0.3)',
          backdropFilter: 'blur(2px)',
        }}
      />

      {/* Panel */}
      <div style={{
        position: 'fixed',
        top: 'var(--topbar-h)',
        right: 12,
        width: 'min(380px, calc(100vw - 24px))',
        background: 'var(--surface)',
        border: '1px solid var(--border)',
        borderRadius: 20,
        zIndex: 300,
        boxShadow: '0 24px 80px rgba(0,0,0,0.6)',
        animation: 'scaleIn 0.18s cubic-bezier(0.4,0,0.2,1)',
        overflow: 'hidden',
        maxHeight: 'calc(100dvh - var(--topbar-h) - 24px)',
        display: 'flex',
        flexDirection: 'column',
      }}>
        {/* Header */}
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          padding: '16px 20px', borderBottom: '1px solid var(--border)',
          flexShrink: 0,
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <Bell size={17} color="var(--t1)" />
            <span style={{ fontWeight: 700, fontSize: 15, color: 'var(--t1)' }}>Notificações</span>
            {unreadCount > 0 && (
              <span style={{
                fontSize: 11, fontWeight: 800, color: 'var(--accent)',
                background: 'rgba(0,229,153,0.15)', borderRadius: 20, padding: '2px 8px',
              }}>{unreadCount} novas</span>
            )}
          </div>
          <div style={{ display: 'flex', gap: 4 }}>
            {unreadCount > 0 && (
              <button
                onClick={() => dispatch('notif_read_all', {})}
                title="Marcar todas como lidas"
                style={{
                  width: 32, height: 32, borderRadius: 8, border: 'none', cursor: 'pointer',
                  background: 'var(--surface-2)',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  color: 'var(--accent)',
                }}
              >
                <CheckCheck size={15} />
              </button>
            )}
            <button
              onClick={() => dispatch('notif_clear', {})}
              title="Limpar lidas"
              style={{
                width: 32, height: 32, borderRadius: 8, border: 'none', cursor: 'pointer',
                background: 'var(--surface-2)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: 'var(--t3)',
              }}
            >
              <Trash2 size={14} />
            </button>
            <button
              onClick={onClose}
              style={{
                width: 32, height: 32, borderRadius: 8, border: 'none', cursor: 'pointer',
                background: 'var(--surface-2)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: 'var(--t2)',
              }}
            >
              <X size={15} />
            </button>
          </div>
        </div>

        {/* Notification list */}
        <div style={{ overflowY: 'auto', flex: 1 }} className="hide-scroll">
          {notifs.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '40px 20px', color: 'var(--t3)' }}>
              <Bell size={32} style={{ marginBottom: 12, opacity: 0.4 }} />
              <p style={{ margin: 0, fontSize: 14 }}>Nenhuma notificação</p>
            </div>
          ) : (
            notifs.map((n, i) => {
              const tc = TYPE_COLORS[n.type] || TYPE_COLORS.market
              return (
                <div
                  key={n.id}
                  onClick={() => !n.read && dispatch('notif_read', { id: n.id })}
                  style={{
                    display: 'flex', gap: 14, padding: '14px 20px',
                    borderBottom: i < notifs.length - 1 ? '1px solid var(--border)' : 'none',
                    cursor: n.read ? 'default' : 'pointer',
                    background: n.read ? 'transparent' : 'rgba(0,229,153,0.02)',
                    transition: 'background 0.15s',
                  }}
                  onMouseEnter={e => { if (!n.read) e.currentTarget.style.background = 'rgba(0,229,153,0.04)' }}
                  onMouseLeave={e => { e.currentTarget.style.background = n.read ? 'transparent' : 'rgba(0,229,153,0.02)' }}
                >
                  {/* Icon */}
                  <div style={{
                    width: 40, height: 40, borderRadius: 12, flexShrink: 0,
                    background: tc.bg,
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: 18, position: 'relative',
                  }}>
                    {n.icon}
                    {!n.read && (
                      <div style={{
                        position: 'absolute', top: 2, right: 2, width: 8, height: 8,
                        borderRadius: '50%', background: 'var(--accent)',
                        border: '1.5px solid var(--surface)',
                      }} />
                    )}
                  </div>

                  {/* Content */}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{
                      fontWeight: n.read ? 600 : 700, fontSize: 13,
                      color: n.read ? 'var(--t2)' : 'var(--t1)',
                      margin: '0 0 2px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis',
                    }}>{n.title}</p>
                    <p style={{ fontSize: 12, color: 'var(--t3)', margin: '0 0 4px', lineHeight: 1.4 }}>{n.body}</p>
                    <p style={{ fontSize: 10, color: 'var(--t3)', margin: 0 }}>{relTime(n.at)}</p>
                  </div>
                </div>
              )
            })
          )}
        </div>
      </div>
    </>
  )
}
