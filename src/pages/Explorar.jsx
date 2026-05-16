import { useState } from 'react'
import { runSimulation, fmtBRL } from '../data/mock.js'
import {
  ShieldCheck, CreditCard, SlidersHorizontal, Activity,
  AlertTriangle, UserCheck, ToggleLeft, BarChart3, FileArchive,
  ChevronDown, ChevronUp
} from 'lucide-react'

const TABS = [
  { key:'card',        label:'Cartão JIT',   Icon: CreditCard },
  { key:'kyc',         label:'KYC & Limites', Icon: UserCheck },
  { key:'compliance',  label:'Compliance',    Icon: ShieldCheck },
  { key:'admin',       label:'Admin',         Icon: SlidersHorizontal },
  { key:'obs',         label:'Observabilidade', Icon: Activity },
]

export default function Explorar({ store, updateStore, showToast }) {
  const [activeTab, setActiveTab] = useState('card')

  const simulate = async (action, data = {}) => {
    await new Promise(r => setTimeout(r, 500))
    const { store: s, message } = runSimulation(store, action, data)
    updateStore(s)
    showToast(message, 'success')
  }

  return (
    <div style={{ display:'flex', flexDirection:'column', gap:20, animation:'slideUp 0.4s ease' }}>
      <div>
        <h2 style={{ fontFamily:'Syne', fontWeight:700, fontSize:22, marginBottom:4 }}>Explorar</h2>
        <p style={{ fontSize:13, color:'var(--t2)' }}>Compliance, Cartões JIT, Governança e Observabilidade</p>
      </div>

      {/* Tab navigation */}
      <div style={{ display:'flex', gap:4, background:'rgba(255,255,255,0.02)', borderRadius:12, padding:4, overflowX:'auto' }}>
        {TABS.map(({ key, label, Icon }) => (
          <button key={key} onClick={() => setActiveTab(key)}
            className={`tab-item ${activeTab === key ? 'active' : ''}`}
            style={{ display:'flex', alignItems:'center', gap:6, border:'none', cursor:'pointer', flexShrink:0 }}>
            <Icon size={13} /> {label}
          </button>
        ))}
      </div>

      {/* Content */}
      {activeTab === 'card'       && <CardJIT       store={store} simulate={simulate} />}
      {activeTab === 'kyc'        && <KYCPanel       store={store} simulate={simulate} />}
      {activeTab === 'compliance' && <CompliancePanel store={store} simulate={simulate} />}
      {activeTab === 'admin'      && <AdminPanel      store={store} simulate={simulate} updateStore={updateStore} showToast={showToast} />}
      {activeTab === 'obs'        && <ObsPanel        store={store} simulate={simulate} />}
    </div>
  )
}

/* ── Card JIT ── */
function CardJIT({ store, simulate }) {
  const [form, setForm] = useState({ merchant:'', mcc:'5411', amount:'', reason:'' })
  const set = (k,v) => setForm(p => ({ ...p, [k]: v }))
  const hasHold = (store?.account?.holds || 0) > 0

  return (
    <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:16 }}>
      {/* Virtual card visual */}
      <div style={{
        background:'linear-gradient(135deg, rgba(139,92,246,0.15), rgba(59,130,246,0.1))',
        border:'1px solid rgba(139,92,246,0.3)',
        borderRadius:24, padding:24, minHeight:180,
        display:'flex', flexDirection:'column', justifyContent:'space-between',
        position:'relative', overflow:'hidden',
      }}>
        <div style={{ position:'absolute', top:-40, right:-40, width:160, height:160, borderRadius:'50%', background:'rgba(139,92,246,0.07)' }} />
        <div style={{ display:'flex', justifyContent:'space-between', alignItems:'flex-start' }}>
          <div style={{ width:40, height:28, borderRadius:6, background:'linear-gradient(135deg, rgba(255,255,255,0.3), rgba(255,255,255,0.1))', border:'1px solid rgba(255,255,255,0.3)' }} />
          <span style={{ fontSize:11, fontWeight:700, letterSpacing:'0.1em', color:'rgba(255,255,255,0.5)' }}>JIT VIRTUAL</span>
        </div>
        <div>
          <p style={{ fontFamily:'Syne', letterSpacing:'0.2em', fontSize:16, marginBottom:10 }}>•••• •••• •••• 8421</p>
          <div style={{ display:'flex', justifyContent:'space-between' }}>
            <div>
              <p style={{ fontSize:9, color:'rgba(255,255,255,0.4)', marginBottom:2 }}>TITULAR</p>
              <p style={{ fontSize:12, fontWeight:600 }}>{store?.user?.name?.toUpperCase()}</p>
            </div>
            <div style={{ textAlign:'right' }}>
              <p style={{ fontSize:9, color:'rgba(255,255,255,0.4)', marginBottom:2 }}>HOLDS</p>
              <p style={{ fontSize:14, fontWeight:700, color: hasHold ? 'var(--gold)' : 'var(--green)', fontFamily:'Syne' }}>
                {fmtBRL(store?.account?.holds || 0)}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="card" style={{ padding:20, display:'flex', flexDirection:'column', gap:12 }}>
        <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, marginBottom:4 }}>Operações JIT</p>
        <input className="input" placeholder="Merchant (ex: iFood)" value={form.merchant} onChange={e=>set('merchant',e.target.value)} />
        <select className="input" value={form.mcc} onChange={e=>set('mcc',e.target.value)}>
          <option value="5411">5411 - Supermercado</option>
          <option value="5812">5812 - Restaurante</option>
          <option value="5999">5999 - Varejo Geral</option>
          <option value="4816">4816 - Tech / Software</option>
        </select>
        <input className="input" type="number" placeholder="Valor (R$)" step="0.01" value={form.amount} onChange={e=>set('amount',e.target.value)} />

        <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
          <button className="btn-primary" onClick={() => simulate('card_authorize', form)}>Autorizar Compra (Hold)</button>
          <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:8 }}>
            <button className="btn-ghost" disabled={!hasHold} onClick={() => simulate('card_confirm', {})}>Confirmar</button>
            <button className="btn-danger" disabled={!hasHold} onClick={() => simulate('card_reject', form)}>Rejeitar</button>
          </div>
        </div>

        {hasHold && (
          <div style={{ background:'rgba(245,158,11,0.08)', border:'1px solid rgba(245,158,11,0.2)', borderRadius:10, padding:'10px 14px', fontSize:12 }}>
            <span style={{ color:'var(--gold)', fontWeight:600 }}>Hold ativo: {fmtBRL(store?.account?.holds)}</span>
            <p style={{ color:'var(--t3)', marginTop:2 }}>Confirme ou rejeite para liberar</p>
          </div>
        )}
      </div>
    </div>
  )
}

/* ── KYC Panel ── */
function KYCPanel({ store, simulate }) {
  const kyc       = store?.kyc || {}
  const usedDaily = kyc.used?.daily || 0
  const limitDay  = kyc.limits?.daily || 10000000
  const pctDay    = Math.min(usedDaily / limitDay, 1)
  const usedMon   = kyc.used?.monthly || 0
  const limitMon  = kyc.limits?.monthly || 50000000
  const pctMon    = Math.min(usedMon / limitMon, 1)

  return (
    <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:16 }}>
      {/* Status card */}
      <div className="card card-accent" style={{ padding:22 }}>
        <div style={{ display:'flex', justifyContent:'space-between', alignItems:'flex-start', marginBottom:20 }}>
          <div>
            <p style={{ fontSize:11, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.08em', marginBottom:6 }}>KYC Level</p>
            <span className={`badge ${kyc.level === 'FULL' ? 'badge-green' : kyc.level === 'BASIC' ? 'badge-blue' : 'badge-gray'}`} style={{ fontSize:13, padding:'4px 12px' }}>
              {kyc.level || 'UNVERIFIED'}
            </span>
          </div>
          <UserCheck size={28} color='var(--green)' />
        </div>

        <div style={{ display:'flex', flexDirection:'column', gap:14 }}>
          <LimitBar label="Uso Diário" used={usedDaily} limit={limitDay} pct={pctDay} />
          <LimitBar label="Uso Mensal" used={usedMon} limit={limitMon} pct={pctMon} />
        </div>
      </div>

      {/* Actions */}
      <div className="card" style={{ padding:22, display:'flex', flexDirection:'column', gap:14 }}>
        <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14 }}>Gerenciar KYC</p>

        <InfoRow label="Limite Diário"   value={fmtBRL(limitDay)} />
        <InfoRow label="Limite Mensal"   value={fmtBRL(limitMon)} />
        <InfoRow label="Disponível Hoje" value={fmtBRL(limitDay - usedDaily)} green />

        <button className="btn-primary" onClick={() => simulate('kyc_limits')}>Ver Limites Detalhados</button>
        {kyc.level !== 'FULL' && (
          <button className="btn-ghost" onClick={() => simulate('kyc_upgrade')}>
            <UserCheck size={14} style={{marginRight:6}} /> Upgrade KYC → FULL
          </button>
        )}
        <button className="btn-ghost" onClick={() => simulate('kyc_limits')}>Ver Limites KYC</button>
      </div>
    </div>
  )
}

function LimitBar({ label, used, limit, pct }) {
  return (
    <div>
      <div style={{ display:'flex', justifyContent:'space-between', fontSize:11, marginBottom:6 }}>
        <span style={{ color:'var(--t3)' }}>{label}</span>
        <span style={{ color:'var(--t2)', fontWeight:600 }}>{fmtBRL(used)} / {fmtBRL(limit)}</span>
      </div>
      <div className="progress-track">
        <div className="progress-fill" style={{
          width: `${pct*100}%`,
          background: pct > 0.8 ? 'var(--red)' : pct > 0.5 ? 'var(--gold)' : undefined,
        }} />
      </div>
    </div>
  )
}

function InfoRow({ label, value, green }) {
  return (
    <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center', padding:'6px 0', borderBottom:'1px solid var(--border)' }}>
      <span style={{ fontSize:12, color:'var(--t3)' }}>{label}</span>
      <span style={{ fontSize:13, fontWeight:700, color: green ? 'var(--green)' : 'var(--t1)', fontFamily:'Syne' }}>{value}</span>
    </div>
  )
}

/* ── Compliance ── */
function CompliancePanel({ store, simulate }) {
  const [form, setForm] = useState({ type:'SUSPICIOUS_ACTIVITY', title:'', eventType:'MANUAL_REVIEW' })
  const set = (k,v) => setForm(p=>({...p,[k]:v}))
  const cases = store?.compliance || []

  return (
    <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:16 }}>
      <div className="card" style={{ padding:22 }}>
        <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, marginBottom:16 }}>Abrir Case</p>
        <div style={{ display:'flex', flexDirection:'column', gap:12 }}>
          <div>
            <label style={{ fontSize:11, color:'var(--t2)', marginBottom:5, display:'block' }}>Tipo</label>
            <select className="input" value={form.type} onChange={e=>set('type',e.target.value)}>
              <option value="SUSPICIOUS_ACTIVITY">Atividade Suspeita</option>
              <option value="AML_ALERT">Alerta AML</option>
              <option value="TRAVEL_RULE">Travel Rule</option>
              <option value="SANCTIONS_HIT">Sanção</option>
            </select>
          </div>
          <div>
            <label style={{ fontSize:11, color:'var(--t2)', marginBottom:5, display:'block' }}>Título</label>
            <input className="input" placeholder="Descreva o caso..." value={form.title} onChange={e=>set('title',e.target.value)} />
          </div>
          <button className="btn-primary" onClick={() => simulate('compliance_case', form)}>
            <AlertTriangle size={14} style={{marginRight:6}} /> Abrir Case
          </button>
          <button className="btn-ghost" onClick={() => simulate('compliance_event', form)}>Registrar Evento</button>
          <button className="btn-ghost" onClick={() => simulate('pre_registration', { email:'empresa@client.com' })}>Iniciar Pré-cadastro</button>
        </div>
      </div>

      <div className="card" style={{ padding:22 }}>
        <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, marginBottom:12 }}>
          Cases Ativos <span className="badge badge-red" style={{marginLeft:8}}>{cases.length}</span>
        </p>
        {cases.length === 0 ? (
          <div style={{ textAlign:'center', padding:'24px 0' }}>
            <ShieldCheck size={32} color='var(--t3)' style={{margin:'0 auto 10px'}} />
            <p style={{ fontSize:13, color:'var(--t3)' }}>Nenhum case aberto</p>
          </div>
        ) : (
          <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
            {cases.map(c => (
              <div key={c.id} style={{ background:'rgba(239,68,68,0.06)', border:'1px solid rgba(239,68,68,0.2)', borderRadius:10, padding:'10px 14px' }}>
                <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center' }}>
                  <p style={{ fontSize:12, fontWeight:600 }}>{c.title || c.type}</p>
                  <span className="badge badge-red" style={{fontSize:9}}>{c.status}</span>
                </div>
                <p style={{ fontSize:10, color:'var(--t3)', marginTop:4 }}>{c.type} · {c.id.slice(0,8)}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

/* ── Admin ── */
function AdminPanel({ store, simulate, updateStore, showToast }) {
  const [plan, setPlan] = useState(store?.user?.plan || 'BUSINESS')
  const [feature, setFeature] = useState('CRYPTO')
  const [daily, setDaily] = useState('50000')

  return (
    <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:16 }}>
      <div className="card" style={{ padding:22 }}>
        <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, marginBottom:16 }}>Gestão de Planos</p>
        <div style={{ display:'flex', flexDirection:'column', gap:12 }}>
          <div>
            <label style={{ fontSize:11, color:'var(--t2)', marginBottom:5, display:'block' }}>Plano do Usuário</label>
            <select className="input" value={plan} onChange={e=>setPlan(e.target.value)}>
              <option value="FREE">FREE</option>
              <option value="BASIC">BASIC</option>
              <option value="BUSINESS">BUSINESS</option>
            </select>
          </div>
          <button className="btn-primary" onClick={() => simulate('admin_plan_change', { plan })}>Trocar Plano</button>
        </div>

        <div className="divider" style={{margin:'16px 0'}} />

        <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, marginBottom:12 }}>Feature Toggles</p>
        <div style={{ display:'flex', flexDirection:'column', gap:10 }}>
          {['PIX','CRYPTO','CARD','COMPLIANCE','INVOICES'].map(f => (
            <div key={f} style={{ display:'flex', justifyContent:'space-between', alignItems:'center' }}>
              <span style={{ fontSize:13 }}>{f}</span>
              <button onClick={() => simulate('admin_feature_toggle', { feature:f, enabled: true })} style={{
                width:40, height:22, borderRadius:11, background:'rgba(0,232,122,0.15)', border:'1px solid rgba(0,232,122,0.3)',
                cursor:'pointer', position:'relative',
              }}>
                <div style={{ width:16, height:16, borderRadius:'50%', background:'var(--green)', position:'absolute', top:2, right:2, transition:'all 0.2s' }} />
              </button>
            </div>
          ))}
        </div>
      </div>

      <div className="card" style={{ padding:22 }}>
        <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, marginBottom:16 }}>Ajustar Limites</p>
        <div style={{ display:'flex', flexDirection:'column', gap:12 }}>
          <div>
            <label style={{ fontSize:11, color:'var(--t2)', marginBottom:5, display:'block' }}>Limite Diário (R$)</label>
            <input className="input" type="number" value={daily} onChange={e=>setDaily(e.target.value)} placeholder="50000" />
          </div>
          <button className="btn-primary" onClick={() => simulate('admin_limit_adjust', { daily })}>Aplicar Limite</button>
        </div>

        <div className="divider" style={{margin:'16px 0'}} />

        <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, marginBottom:12 }}>Pré-cadastros</p>
        <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
          <button className="btn-ghost" onClick={() => simulate('pre_registration', { email:'empresa@novo.com.br' })}>Iniciar Pré-cadastro</button>
          <button className="btn-ghost" onClick={() => simulate('pre_registration', { email:'pj@corporativo.com' })}>Simular Verificação</button>
        </div>

        {(store?.preRegistrations||[]).length > 0 && (
          <div style={{ marginTop:12, fontSize:12 }}>
            <p style={{ color:'var(--t3)', marginBottom:8 }}>{store.preRegistrations.length} pré-cadastro(s) pendente(s)</p>
            {store.preRegistrations.slice(0,3).map(pr => (
              <div key={pr.id} style={{ padding:'6px 0', borderBottom:'1px solid var(--border)', display:'flex', justifyContent:'space-between' }}>
                <span style={{ color:'var(--t2)' }}>{pr.email}</span>
                <span className="badge badge-gold" style={{fontSize:9}}>{pr.status}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

/* ── Observability ── */
function ObsPanel({ store, simulate }) {
  const obs = store?.observability || {}

  const metrics = [
    { label:'Uptime', value: obs.uptime || '99.97%', color:'var(--green)', badge:'badge-green' },
    { label:'Latência P99', value: `${obs.latencyMs || 42}ms`, color:'var(--blue)', badge:'badge-blue' },
    { label:'Webhooks Pendentes', value: obs.pendingWebhooks ?? 1, color:'var(--gold)', badge:'badge-gold' },
    { label:'Reconciliação', value: obs.reconcilePending ?? 0, color:'var(--green)', badge:'badge-green' },
    { label:'Retries Ativos', value: obs.retries ?? 2, color: obs.retries > 0 ? 'var(--gold)' : 'var(--green)', badge: obs.retries > 0 ? 'badge-gold' : 'badge-green' },
  ]

  return (
    <div style={{ display:'flex', flexDirection:'column', gap:16 }}>
      {/* Metrics grid */}
      <div style={{ display:'grid', gridTemplateColumns:'repeat(3, 1fr)', gap:12 }}>
        {metrics.map(m => (
          <div key={m.label} className="card" style={{ padding:18 }}>
            <p style={{ fontSize:11, color:'var(--t3)', marginBottom:8 }}>{m.label}</p>
            <p style={{ fontFamily:'Syne', fontWeight:800, fontSize:22, color:m.color, marginBottom:6 }}>{m.value}</p>
            <span className={`badge ${m.badge}`} style={{fontSize:9}}>
              {m.value === 0 || m.value === '99.97%' ? 'Normal' : 'Atenção'}
            </span>
          </div>
        ))}
      </div>

      {/* Actions */}
      <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:16 }}>
        <div className="card" style={{ padding:22 }}>
          <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, marginBottom:16 }}>Diagnóstico</p>
          <div style={{ display:'flex', flexDirection:'column', gap:10 }}>
            <button className="btn-ghost" onClick={() => simulate('observability_summary')} style={{display:'flex',alignItems:'center',gap:8}}>
              <BarChart3 size={14}/> Resumo Observabilidade
            </button>
            <button className="btn-ghost" onClick={() => simulate('audit_archive')} style={{display:'flex',alignItems:'center',gap:8}}>
              <FileArchive size={14}/> Exportar Auditoria
            </button>
          </div>
        </div>
        <div className="card" style={{ padding:22 }}>
          <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, marginBottom:12 }}>Status dos Serviços</p>
          {[
            ['Ledger / Transactions', true],
            ['PIX Partner Integration', true],
            ['Crypto Exchange Gateway', true],
            ['Webhook Worker', obs.pendingWebhooks === 0],
            ['Pricing Engine', true],
          ].map(([svc, ok]) => (
            <div key={svc} style={{ display:'flex', justifyContent:'space-between', padding:'6px 0', borderBottom:'1px solid var(--border)', fontSize:12 }}>
              <span style={{ color:'var(--t2)' }}>{svc}</span>
              <span style={{ color: ok ? 'var(--green)' : 'var(--gold)', fontWeight:600, fontSize:11 }}>
                {ok ? '● Operacional' : '● Degradado'}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
