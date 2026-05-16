import { useState } from 'react'
import {
  LayoutDashboard, ArrowLeftRight, TrendingUp, Compass,
  LogOut, Settings, Bell, Sparkles, ChevronDown,
  Menu, X, ChevronRight
} from 'lucide-react'
import Dashboard from '../pages/Dashboard.jsx'
import Movimentar from '../pages/Movimentar.jsx'
import Investir from '../pages/Investir.jsx'
import Explorar from '../pages/Explorar.jsx'
import ZionPanel from './ZionPanel.jsx'
import SimModal from './SimModal.jsx'

const NAV = [
  { key:'dashboard',  label:'Dashboard',   Icon: LayoutDashboard },
  { key:'movimentar', label:'Movimentar',  Icon: ArrowLeftRight  },
  { key:'investir',   label:'Investir',    Icon: TrendingUp       },
  { key:'explorar',   label:'Explorar',    Icon: Compass          },
]

export default function Layout({ page, setPage, store, updateStore, showZion, setShowZion, simModal, setSimModal, logout, showToast, isDemo }) {
  const [sideOpen, setSideOpen] = useState(false)
  const [userMenu, setUserMenu] = useState(false)

  const navigate = (key) => {
    setPage(key)
    setSideOpen(false)
  }

  const planColors = {
    FREE: 'badge-gray', BASIC: 'badge-blue', BUSINESS: 'badge-purple'
  }

  return (
    <div style={{ display:'flex', height:'100vh', overflow:'hidden', background:'var(--bg)' }}>

      {/* ── Mobile Sidebar Overlay ── */}
      {sideOpen && (
        <div className="sidebar-overlay hide-desktop" onClick={() => setSideOpen(false)} />
      )}

      {/* ── Sidebar ── */}
      <aside style={{
        width:'var(--sidebar-w)',
        background:'rgba(10,14,24,0.95)',
        backdropFilter:'blur(24px)',
        borderRight:'1px solid var(--border)',
        display:'flex', flexDirection:'column',
        position:'fixed', left:0, top:0, bottom:0,
        zIndex:250,
        transform: sideOpen ? 'translateX(0)' : 'translateX(-100%)',
        transition:'transform 0.25s ease',
      }}
        className="show-mobile"
      >
        <SidebarContent page={page} navigate={navigate} store={store} logout={logout} planColors={planColors} />
      </aside>

      {/* Desktop sidebar */}
      <aside className="hide-mobile" style={{
        width:'var(--sidebar-w)',
        flexShrink:0,
        background:'rgba(10,14,24,0.95)',
        backdropFilter:'blur(24px)',
        borderRight:'1px solid var(--border)',
        display:'flex', flexDirection:'column',
      }}>
        <SidebarContent page={page} navigate={navigate} store={store} logout={logout} planColors={planColors} />
      </aside>

      {/* ── Main Column ── */}
      <div style={{ flex:1, display:'flex', flexDirection:'column', minWidth:0, overflow:'hidden' }}>

        {/* TopBar */}
        <header style={{
          height:'var(--topbar-h)',
          background:'rgba(10,14,24,0.9)',
          backdropFilter:'blur(20px)',
          borderBottom:'1px solid var(--border)',
          display:'flex', alignItems:'center',
          padding:'0 20px', gap:12, flexShrink:0, position:'relative',
        }}>
          {/* Mobile menu */}
          <button className="btn-icon show-mobile hide-desktop" onClick={() => setSideOpen(true)}>
            <Menu size={16} />
          </button>

          {/* Breadcrumb */}
          <div className="hide-mobile" style={{ display:'flex', alignItems:'center', gap:6, fontSize:13, color:'var(--t3)' }}>
            <span>Z-Finance</span>
            <ChevronRight size={12} />
            <span style={{ color:'var(--t1)', fontWeight:600 }}>{NAV.find(n=>n.key===page)?.label}</span>
          </div>

          {/* Mobile brand */}
          <span className="show-mobile hide-desktop" style={{ fontFamily:'Syne', fontWeight:800, fontSize:16, letterSpacing:'0.06em' }}>
            Z-<span style={{ color:'var(--green)' }}>FINANCE</span>
          </span>

          <div style={{ flex:1 }} />

          {/* Demo badge */}
          {isDemo && (
            <div className="badge badge-gold hide-mobile">Demo</div>
          )}

          {/* Notifications */}
          <button className="btn-icon" style={{ position:'relative' }} onClick={() => showToast('2 alertas de compliance pendentes', 'info')}>
            <Bell size={15} />
            <span style={{ position:'absolute', top:-2, right:-2, width:8, height:8, borderRadius:'50%', background:'var(--red)', border:'1.5px solid var(--bg)' }} />
          </button>

          {/* Zion button */}
          <button onClick={() => setShowZion(!showZion)} style={{
            display:'flex', alignItems:'center', gap:8,
            background: showZion ? 'rgba(0,232,122,0.12)' : 'rgba(255,255,255,0.04)',
            border:`1px solid ${showZion ? 'rgba(0,232,122,0.35)' : 'var(--border)'}`,
            borderRadius:10, padding:'7px 12px', cursor:'pointer', transition:'all 0.2s',
          }}>
            <div style={{ width:8, height:8, borderRadius:'50%', background:'var(--green)', boxShadow:'0 0 8px rgba(0,232,122,0.6)' }} />
            <span style={{ fontSize:12, fontWeight:600, color: showZion ? 'var(--green)' : 'var(--t2)' }}>Zion AI</span>
          </button>

          {/* User avatar */}
          <div style={{ position:'relative' }}>
            <button onClick={() => setUserMenu(!userMenu)} style={{
              display:'flex', alignItems:'center', gap:8, background:'none', border:'none', cursor:'pointer', padding:'4px 6px', borderRadius:10,
            }}>
              <div style={{
                width:32, height:32, borderRadius:'50%',
                background:'linear-gradient(135deg, var(--green), var(--green2))',
                display:'flex', alignItems:'center', justifyContent:'center',
                fontSize:12, fontWeight:700, color:'#07090f',
              }}>
                {store?.user?.avatarInitials || 'ZF'}
              </div>
              <ChevronDown size={12} color='var(--t3)' />
            </button>

            {/* User dropdown */}
            {userMenu && (
              <>
                <div style={{ position:'fixed', inset:0, zIndex:300 }} onClick={() => setUserMenu(false)} />
                <div style={{
                  position:'absolute', right:0, top:'calc(100% + 8px)', width:220,
                  background:'var(--card2)', border:'1px solid rgba(255,255,255,0.1)',
                  borderRadius:14, padding:12, zIndex:310,
                  boxShadow:'0 16px 48px rgba(0,0,0,0.5)',
                }}>
                  <div style={{ paddingBottom:10, marginBottom:10, borderBottom:'1px solid var(--border)' }}>
                    <p style={{ fontSize:13, fontWeight:600, marginBottom:2 }}>{store?.user?.name}</p>
                    <p style={{ fontSize:11, color:'var(--t3)' }}>{store?.user?.email}</p>
                    <span className={`badge ${planColors[store?.user?.plan] || 'badge-gray'}`} style={{ marginTop:6 }}>{store?.user?.plan}</span>
                  </div>
                  <button className="btn-ghost" style={{ width:'100%', marginBottom:8, fontSize:12, padding:'8px 12px' }}
                    onClick={() => { setUserMenu(false); showToast('Configurações em breve', 'info') }}>
                    <Settings size={13} style={{ marginRight:6 }} /> Configurações
                  </button>
                  <button className="btn-danger" style={{ width:'100%', fontSize:12, padding:'8px 12px' }}
                    onClick={() => { setUserMenu(false); logout() }}>
                    <LogOut size={13} style={{ marginRight:6 }} /> Sair
                  </button>
                </div>
              </>
            )}
          </div>
        </header>

        {/* Content */}
        <main style={{ flex:1, overflowY:'auto', padding:'20px', paddingBottom:80 }}>
          <div style={{ maxWidth:1100, margin:'0 auto' }}>
            {page === 'dashboard'  && <Dashboard  store={store} updateStore={updateStore} setSimModal={setSimModal} showToast={showToast} />}
            {page === 'movimentar' && <Movimentar store={store} updateStore={updateStore} showToast={showToast} />}
            {page === 'investir'   && <Investir   store={store} updateStore={updateStore} showToast={showToast} />}
            {page === 'explorar'   && <Explorar   store={store} updateStore={updateStore} showToast={showToast} />}
          </div>
        </main>

        {/* Bottom Nav (mobile) */}
        <nav className="show-mobile hide-desktop" style={{
          position:'fixed', bottom:0, left:0, right:0,
          background:'rgba(10,14,24,0.95)',
          backdropFilter:'blur(20px)',
          borderTop:'1px solid var(--border)',
          display:'flex', padding:'8px 0 max(8px, env(safe-area-inset-bottom))',
          zIndex:100,
        }}>
          {NAV.map(({ key, label, Icon }) => (
            <button key={key} onClick={() => navigate(key)}
              className={`bnav-item ${page === key ? 'active' : ''}`}
              style={{ background:'none', border:'none', cursor:'pointer' }}>
              <Icon size={20} />
              <span>{label}</span>
            </button>
          ))}
        </nav>
      </div>

      {/* Zion Panel */}
      {showZion && (
        <ZionPanel store={store} onClose={() => setShowZion(false)} showToast={showToast} />
      )}

      {/* Simulation Modal */}
      {simModal && (
        <SimModal
          config={simModal}
          store={store}
          updateStore={updateStore}
          showToast={showToast}
          onClose={() => setSimModal(null)}
        />
      )}
    </div>
  )
}

function SidebarContent({ page, navigate, store, logout, planColors }) {
  return (
    <>
      {/* Logo */}
      <div style={{ padding:'20px 16px 16px', borderBottom:'1px solid var(--border)' }}>
        <div style={{ display:'flex', alignItems:'center', gap:10 }}>
          <div style={{
            width:36, height:36, borderRadius:10,
            background:'linear-gradient(135deg, rgba(0,232,122,0.15), rgba(0,232,122,0.05))',
            border:'1.5px solid rgba(0,232,122,0.4)',
            display:'flex', alignItems:'center', justifyContent:'center',
            boxShadow:'0 0 20px rgba(0,232,122,0.15)',
          }}>
            <span style={{ fontFamily:'Syne', fontWeight:900, fontSize:18, color:'var(--green)' }}>Z</span>
          </div>
          <div>
            <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, letterSpacing:'0.08em' }}>Z-FINANCE</p>
            <p style={{ fontSize:10, color:'var(--t3)', letterSpacing:'0.1em', textTransform:'uppercase' }}>Core Banking</p>
          </div>
        </div>
      </div>

      {/* Nav */}
      <nav style={{ padding:'12px 12px', flex:1 }}>
        <p style={{ fontSize:10, color:'var(--t3)', letterSpacing:'0.1em', textTransform:'uppercase', marginBottom:8, paddingLeft:12 }}>Menu</p>
        {NAV.map(({ key, label, Icon }) => (
          <button key={key} onClick={() => navigate(key)}
            className={`nav-item ${page === key ? 'active' : ''}`}
            style={{ width:'100%', border:'none', background:'none', cursor:'pointer', textAlign:'left', marginBottom:2 }}>
            <Icon size={16} />
            {label}
          </button>
        ))}
      </nav>

      {/* User info */}
      <div style={{ padding:'12px 16px', borderTop:'1px solid var(--border)' }}>
        <div style={{ display:'flex', alignItems:'center', gap:10 }}>
          <div style={{
            width:32, height:32, borderRadius:'50%', flexShrink:0,
            background:'linear-gradient(135deg, var(--green), var(--green2))',
            display:'flex', alignItems:'center', justifyContent:'center',
            fontSize:11, fontWeight:700, color:'#07090f',
          }}>
            {store?.user?.avatarInitials || 'ZF'}
          </div>
          <div style={{ minWidth:0, flex:1 }}>
            <p style={{ fontSize:12, fontWeight:600, whiteSpace:'nowrap', overflow:'hidden', textOverflow:'ellipsis' }}>
              {store?.user?.name || 'Usuário'}
            </p>
            <span className={`badge ${planColors[store?.user?.plan] || 'badge-gray'}`} style={{ fontSize:9 }}>
              {store?.user?.plan || 'FREE'}
            </span>
          </div>
        </div>
      </div>
    </>
  )
}
