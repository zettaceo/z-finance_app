import { useState } from 'react'
import {
  Mail, Lock, Zap, ArrowRight, Globe, Eye, EyeOff,
  Shield, TrendingUp, Cpu
} from 'lucide-react'

const FEATURES = [
  { icon: Shield, label: 'KYC & Compliance VASP-ready' },
  { icon: TrendingUp, label: 'Pricing Engine desacoplado' },
  { icon: Cpu, label: 'Ledger append-only + PIX + Cripto' },
]

export default function Login({ onDemo, onLive }) {
  const [tab,      setTab]      = useState('demo')
  const [email,    setEmail]    = useState('')
  const [password, setPassword] = useState('')
  const [apiBase,  setApiBase]  = useState('http://95.111.247.134:8080')
  const [showPw,   setShowPw]   = useState(false)
  const [loading,  setLoading]  = useState(false)
  const [error,    setError]    = useState('')

  async function handleLive(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await onLive(email, password, apiBase)
    } catch (err) {
      setError(err.message || 'Falha na conexão')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center relative overflow-hidden" style={{ background: 'var(--bg)' }}>
      {/* Orbs */}
      <div className="orb" style={{ width:500, height:500, top:'-10%', left:'-10%', background:'radial-gradient(circle, rgba(0,232,122,0.08) 0%, transparent 70%)' }} />
      <div className="orb" style={{ width:400, height:400, bottom:'-10%', right:'-5%', background:'radial-gradient(circle, rgba(59,130,246,0.07) 0%, transparent 70%)', animationDelay:'4s' }} />
      <div className="orb" style={{ width:300, height:300, top:'40%', left:'60%', background:'radial-gradient(circle, rgba(139,92,246,0.05) 0%, transparent 70%)', animationDelay:'2s' }} />

      {/* Grid overlay */}
      <div className="absolute inset-0" style={{ backgroundImage:'linear-gradient(rgba(255,255,255,0.015) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.015) 1px, transparent 1px)', backgroundSize:'60px 60px' }} />

      <div className="relative w-full max-w-md px-6 animate-fade-in">
        {/* Logo */}
        <div className="flex items-center gap-4 mb-10">
          <div style={{
            width:52, height:52, borderRadius:14,
            background:'linear-gradient(135deg, rgba(0,232,122,0.15), rgba(0,232,122,0.05))',
            border:'1.5px solid rgba(0,232,122,0.4)',
            display:'flex', alignItems:'center', justifyContent:'center',
            boxShadow:'0 0 30px rgba(0,232,122,0.2)',
          }}>
            <span style={{ fontFamily:'Syne', fontWeight:900, fontSize:22, color:'var(--green)' }}>Z</span>
          </div>
          <div>
            <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:20, letterSpacing:'0.1em', color:'var(--t1)' }}>Z-FINANCE</p>
            <p style={{ fontSize:12, color:'var(--t3)', letterSpacing:'0.08em', textTransform:'uppercase' }}>Core Banking Platform</p>
          </div>
        </div>

        {/* Card */}
        <div style={{
          background:'rgba(13,18,32,0.9)',
          backdropFilter:'blur(32px)',
          border:'1px solid rgba(255,255,255,0.08)',
          borderRadius:24,
          padding:28,
          boxShadow:'0 32px 80px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.05)',
        }}>
          <h1 style={{ fontFamily:'Syne', fontWeight:700, fontSize:22, marginBottom:6 }}>Acesso Seguro</h1>
          <p style={{ fontSize:13, color:'var(--t2)', marginBottom:24 }}>Plataforma institucional com segurança reforçada.</p>

          {/* Tabs */}
          <div style={{ display:'flex', gap:6, marginBottom:24, background:'rgba(255,255,255,0.03)', borderRadius:12, padding:4 }}>
            {[
              { key:'demo', label:'Modo Demo' },
              { key:'live', label:'API Live' },
            ].map(t => (
              <button key={t.key} onClick={() => { setTab(t.key); setError('') }}
                style={{
                  flex:1, padding:'9px 0', borderRadius:9, border:'none', cursor:'pointer',
                  background: tab === t.key ? 'rgba(0,232,122,0.12)' : 'transparent',
                  color: tab === t.key ? 'var(--green)' : 'var(--t3)',
                  fontWeight: tab === t.key ? 700 : 500,
                  fontSize:13, transition:'all 0.2s',
                }}>{t.label}</button>
            ))}
          </div>

          {tab === 'demo' ? (
            <div>
              {/* Demo info */}
              <div style={{ background:'rgba(0,232,122,0.04)', border:'1px solid rgba(0,232,122,0.15)', borderRadius:12, padding:16, marginBottom:20 }}>
                <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:10 }}>
                  <Zap size={14} color='var(--green)' />
                  <span style={{ fontSize:12, fontWeight:700, color:'var(--green)' }}>Demo com dados realistas</span>
                </div>
                <p style={{ fontSize:12, color:'var(--t2)', lineHeight:1.6, marginBottom:12 }}>
                  Explore todas as funcionalidades: PIX, Cripto, Card JIT, Pricing Engine e Compliance — sem conexão com backend.
                </p>
                <div style={{ display:'flex', flexDirection:'column', gap:6 }}>
                  {FEATURES.map(({ icon: Icon, label }) => (
                    <div key={label} style={{ display:'flex', alignItems:'center', gap:8 }}>
                      <Icon size={12} color='var(--t3)' />
                      <span style={{ fontSize:11, color:'var(--t3)' }}>{label}</span>
                    </div>
                  ))}
                </div>
              </div>
              <button className="btn-primary" style={{ width:'100%', fontSize:14, padding:'13px 0', display:'flex', alignItems:'center', justifyContent:'center', gap:8 }}
                onClick={onDemo}>
                Entrar no modo demo <ArrowRight size={16} />
              </button>
            </div>
          ) : (
            <form onSubmit={handleLive} style={{ display:'flex', flexDirection:'column', gap:14 }}>
              {/* API Base */}
              <div>
                <label style={{ fontSize:12, color:'var(--t2)', marginBottom:6, display:'block' }}>API Base URL</label>
                <div style={{ position:'relative' }}>
                  <Globe size={14} style={{ position:'absolute', left:12, top:'50%', transform:'translateY(-50%)', color:'var(--t3)' }} />
                  <input className="input" style={{ paddingLeft:36 }} type="url" placeholder="http://..." value={apiBase} onChange={e=>setApiBase(e.target.value)} />
                </div>
              </div>
              {/* Email */}
              <div>
                <label style={{ fontSize:12, color:'var(--t2)', marginBottom:6, display:'block' }}>Email</label>
                <div style={{ position:'relative' }}>
                  <Mail size={14} style={{ position:'absolute', left:12, top:'50%', transform:'translateY(-50%)', color:'var(--t3)' }} />
                  <input className="input" style={{ paddingLeft:36 }} type="email" placeholder="investidor@zetta.bank" required value={email} onChange={e=>setEmail(e.target.value)} />
                </div>
              </div>
              {/* Password */}
              <div>
                <label style={{ fontSize:12, color:'var(--t2)', marginBottom:6, display:'block' }}>Senha</label>
                <div style={{ position:'relative' }}>
                  <Lock size={14} style={{ position:'absolute', left:12, top:'50%', transform:'translateY(-50%)', color:'var(--t3)' }} />
                  <input className="input" style={{ paddingLeft:36, paddingRight:40 }} type={showPw ? 'text' : 'password'} placeholder="••••••••" required value={password} onChange={e=>setPassword(e.target.value)} />
                  <button type="button" onClick={() => setShowPw(!showPw)} style={{ position:'absolute', right:12, top:'50%', transform:'translateY(-50%)', background:'none', border:'none', color:'var(--t3)', cursor:'pointer' }}>
                    {showPw ? <EyeOff size={14}/> : <Eye size={14}/>}
                  </button>
                </div>
              </div>
              {error && (
                <p style={{ fontSize:12, color:'var(--red)', background:'rgba(239,68,68,0.08)', padding:'8px 12px', borderRadius:8, border:'1px solid rgba(239,68,68,0.2)' }}>
                  {error}
                </p>
              )}
              <button className="btn-primary" type="submit" disabled={loading}
                style={{ marginTop:4, display:'flex', alignItems:'center', justifyContent:'center', gap:8, opacity: loading ? 0.7 : 1 }}>
                {loading ? 'Conectando...' : <><span>Entrar</span><ArrowRight size={16}/></>}
              </button>
            </form>
          )}
        </div>

        {/* Footer */}
        <p style={{ textAlign:'center', fontSize:11, color:'var(--t3)', marginTop:24 }}>
          Z-FINANCE © {new Date().getFullYear()} · Zetta Core Banking ·{' '}
          <span style={{ color:'var(--green)', fontWeight:600 }}>VASP-ready</span>
        </p>
      </div>
    </div>
  )
}
