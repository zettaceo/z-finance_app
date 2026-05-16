import React, { useState, useCallback } from 'react'
import { Home, ArrowLeftRight, CreditCard, TrendingUp, MoreHorizontal, Zap, Bell, ChevronDown, LogOut, Menu, X } from 'lucide-react'
import { useApp } from '../App.jsx'

const NAV = [
  { id: 'home',    Icon: Home,           label: 'Início' },
  { id: 'mover',   Icon: ArrowLeftRight, label: 'Mover' },
  { id: 'cartoes', Icon: CreditCard,     label: 'Cartões' },
  { id: 'investir',Icon: TrendingUp,     label: 'Investir' },
  { id: 'mais',    Icon: MoreHorizontal, label: 'Mais' },
]

export default function Layout({ page, setPage, children }) {
  const { store, logout, setZionOpen } = useApp()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const u = store.user

  const nav = useCallback((id) => {
    setPage(id)
    setSidebarOpen(false)
  }, [setPage])

  const planColor = u.plan === 'INSTITUTIONAL' ? '#FBBF24' : u.plan === 'BUSINESS' ? '#818CF8' : 'var(--accent)'

  return (
    <div className="app-shell">
      {/* Sidebar overlay on mobile */}
      {sidebarOpen && (
        <div
          onClick={() => setSidebarOpen(false)}
          style={{ position: 'fixed', inset: 0, background: 'rgba(4,12,27,0.7)', backdropFilter: 'blur(4px)', zIndex: 99 }}
        />
      )}

      {/* Sidebar */}
      <aside className={`sidebar${sidebarOpen ? ' open' : ''}`}>
        <div className="sidebar-inner">
          {/* Logo */}
          <div style={{ padding: '24px 20px 20px', borderBottom: '1px solid var(--border)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{
                width: 40, height: 40, borderRadius: 12,
                background: 'linear-gradient(135deg, var(--accent), var(--accent-2))',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                flexShrink: 0,
              }}>
                <span style={{ fontFamily: 'Syne,sans-serif', fontWeight: 900, fontSize: 18, color: '#040C1B' }}>Z</span>
              </div>
              <div>
                <p style={{ fontFamily: 'Syne,sans-serif', fontWeight: 800, fontSize: 16, color: 'var(--t1)', margin: 0 }}>Z-Finance</p>
                <p style={{ fontSize: 11, color: 'var(--t3)', margin: 0 }}>Global Banking</p>
              </div>
            </div>
          </div>

          {/* User card */}
          <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{
                width: 44, height: 44, borderRadius: 14,
                background: 'linear-gradient(135deg, #1E3A5F, #0C1E38)',
                border: '2px solid var(--border)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                flexShrink: 0,
                fontSize: 16, fontWeight: 700, color: 'var(--accent)',
              }}>
                {u.name.charAt(0)}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{u.name}</p>
                <span style={{
                  fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em',
                  color: planColor, background: `${planColor}18`, borderRadius: 6, padding: '2px 6px',
                }}>
                  {u.plan}
                </span>
              </div>
            </div>
          </div>

          {/* Nav items */}
          <nav style={{ padding: '12px 12px', flex: 1 }}>
            {NAV.map(({ id, Icon, label }) => {
              const active = page === id
              return (
                <button key={id} onClick={() => nav(id)} style={{
                  width: '100%', display: 'flex', alignItems: 'center', gap: 12,
                  padding: '12px 14px', borderRadius: 12, border: 'none', cursor: 'pointer',
                  background: active ? 'rgba(0,229,153,0.1)' : 'transparent',
                  color: active ? 'var(--accent)' : 'var(--t2)',
                  fontSize: 14, fontWeight: active ? 700 : 500,
                  transition: 'all 0.15s ease',
                  marginBottom: 2,
                  textAlign: 'left',
                }}>
                  <Icon size={18} style={{ flexShrink: 0 }} />
                  {label}
                  {active && <div style={{ marginLeft: 'auto', width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)' }} />}
                </button>
              )
            })}
          </nav>

          {/* Zion AI button */}
          <div style={{ padding: '12px 12px' }}>
            <button onClick={() => setZionOpen(true)} style={{
              width: '100%', padding: '12px 14px', borderRadius: 14,
              background: 'linear-gradient(135deg, rgba(0,229,153,0.15), rgba(0,229,153,0.05))',
              border: '1px solid rgba(0,229,153,0.2)',
              cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 10,
            }}>
              <div style={{
                width: 28, height: 28, borderRadius: 8,
                background: 'linear-gradient(135deg, var(--accent), var(--accent-2))',
                display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
              }}>
                <Zap size={14} color="#040C1B" fill="#040C1B" />
              </div>
              <div style={{ textAlign: 'left' }}>
                <p style={{ fontSize: 13, fontWeight: 700, color: 'var(--accent)', margin: 0 }}>Zion AI</p>
                <p style={{ fontSize: 11, color: 'var(--t3)', margin: 0 }}>Assistente financeiro</p>
              </div>
            </button>
          </div>

          {/* Logout */}
          <div style={{ padding: '0 12px 12px' }}>
            <button onClick={logout} style={{
              width: '100%', padding: '12px 14px', borderRadius: 12, border: 'none',
              background: 'transparent', cursor: 'pointer',
              display: 'flex', alignItems: 'center', gap: 10,
              color: 'var(--t3)', fontSize: 14,
            }}>
              <LogOut size={16} />
              Sair
            </button>
          </div>
        </div>
      </aside>

      {/* Main */}
      <div className="main-area">
        {/* Topbar */}
        <header className="topbar">
          {/* Mobile: hamburger */}
          <button
            className="sidebar-toggle"
            onClick={() => setSidebarOpen(o => !o)}
            style={{
              width: 40, height: 40, borderRadius: 12, border: 'none',
              background: 'var(--surface-2)', cursor: 'pointer',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: 'var(--t1)',
            }}
          >
            {sidebarOpen ? <X size={18} /> : <Menu size={18} />}
          </button>

          {/* Center: Page title on mobile */}
          <div className="topbar-title">
            <span style={{ fontFamily: 'Syne,sans-serif', fontWeight: 800, fontSize: 17, color: 'var(--t1)' }}>
              {NAV.find(n => n.id === page)?.label || 'Z-Finance'}
            </span>
          </div>

          {/* Right actions */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <button style={{
              width: 40, height: 40, borderRadius: 12, border: 'none',
              background: 'var(--surface-2)', cursor: 'pointer',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: 'var(--t2)', position: 'relative',
            }}>
              <Bell size={17} />
              <span style={{
                position: 'absolute', top: 8, right: 8, width: 7, height: 7,
                borderRadius: '50%', background: 'var(--accent)',
                border: '1px solid var(--bg)',
              }} />
            </button>
            <button onClick={() => setZionOpen(true)} style={{
              width: 40, height: 40, borderRadius: 12, border: 'none',
              background: 'linear-gradient(135deg, var(--accent), var(--accent-2))',
              cursor: 'pointer',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: '#040C1B',
            }}>
              <Zap size={17} fill="#040C1B" />
            </button>
          </div>
        </header>

        {/* Page content */}
        <main className="content">
          {children}
        </main>
      </div>

      {/* Bottom nav (mobile) */}
      <nav className="bottom-nav">
        {NAV.map(({ id, Icon, label }) => {
          const active = page === id
          return (
            <button key={id} onClick={() => setPage(id)} style={{
              flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center',
              justifyContent: 'center', gap: 3, border: 'none', cursor: 'pointer',
              background: 'transparent', padding: '8px 0',
              color: active ? 'var(--accent)' : 'var(--t3)',
              transition: 'color 0.15s ease',
              minHeight: 44,
            }}>
              <Icon size={20} strokeWidth={active ? 2.5 : 1.8} />
              <span style={{ fontSize: 10, fontWeight: active ? 700 : 500 }}>{label}</span>
              {active && <div style={{ width: 4, height: 4, borderRadius: '50%', background: 'var(--accent)', marginTop: -1 }} />}
            </button>
          )
        })}
      </nav>
    </div>
  )
}
