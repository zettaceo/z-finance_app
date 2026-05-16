import BalanceCard    from '../components/BalanceCard.jsx'
import RiskGauge     from '../components/RiskGauge.jsx'
import ActivityFeed  from '../components/ActivityFeed.jsx'
import CryptoPortfolio from '../components/CryptoPortfolio.jsx'
import CashFlowChart from '../components/CashFlowChart.jsx'
import { fmtBRL }   from '../data/mock.js'
import { ArrowDownLeft, CreditCard, ArrowUpRight, Receipt, RefreshCw, Wallet, ChevronRight } from 'lucide-react'

const QUICK_ACTIONS = [
  {
    key:   'invoice_create',
    label: 'Receber',
    sub:   'Criar cobrança',
    Icon:  ArrowDownLeft,
    color: 'var(--green)',
    bg:    'rgba(0,232,122,0.12)',
    border:'rgba(0,232,122,0.25)',
    fields:[
      { key:'amount', label:'Valor (R$)', type:'number', placeholder:'500,00', step:'0.01', min:'0.01', required:true },
      { key:'description', label:'Descrição (opcional)', type:'text', placeholder:'Cobrança de serviço' },
    ],
    title: 'Criar Cobrança',
  },
  {
    key:   'payment_schedule',
    label: 'Pagar',
    sub:   'Boleto / PIX',
    Icon:  Receipt,
    color: 'var(--blue)',
    bg:    'rgba(59,130,246,0.12)',
    border:'rgba(59,130,246,0.25)',
    fields:[
      { key:'barcode', label:'Código de Barras', type:'text', placeholder:'000 0000 00000 0...', required:true },
      { key:'amount',  label:'Valor (R$)', type:'number', placeholder:'120,00', step:'0.01', min:'0.01', required:true },
      { key:'dueDate', label:'Vencimento', type:'date' },
    ],
    title: 'Agendar Pagamento',
  },
  {
    key:   'pix_send',
    label: 'Enviar',
    sub:   'PIX ou TED',
    Icon:  ArrowUpRight,
    color: 'var(--purple)',
    bg:    'rgba(139,92,246,0.12)',
    border:'rgba(139,92,246,0.25)',
    fields:[
      { key:'receiver', label:'Destinatário / Chave PIX', type:'text', placeholder:'email@bank.com', required:true },
      { key:'amount',   label:'Valor (R$)', type:'number', placeholder:'250,00', step:'0.01', min:'0.01', required:true },
      { key:'description', label:'Descrição', type:'text', placeholder:'PIX transferência', default:'PIX transferência' },
    ],
    title: 'Enviar PIX',
  },
  {
    key:   'transfer_internal',
    label: 'Transferir',
    sub:   'Entre contas',
    Icon:  RefreshCw,
    color: 'var(--gold)',
    bg:    'rgba(245,158,11,0.12)',
    border:'rgba(245,158,11,0.25)',
    fields:[
      { key:'amount', label:'Valor (R$)', type:'number', placeholder:'1.000,00', step:'0.01', min:'0.01', required:true },
      { key:'from',   label:'Conta Origem', type:'select', options:[{ value:'main', label:'Conta Principal' }], default:'main' },
      { key:'to',     label:'Conta Destino', type:'select', options:[{ value:'investments', label:'Conta Investimento' }], default:'investments' },
    ],
    title: 'Transferência Interna',
  },
]

export default function Dashboard({ store, updateStore, setSimModal, showToast }) {
  if (!store) return null

  const greeting = () => {
    const h = new Date().getHours()
    if (h < 12) return 'Bom dia'
    if (h < 18) return 'Boa tarde'
    return 'Boa noite'
  }

  return (
    <div style={{ display:'flex', flexDirection:'column', gap:20, animation:'slideUp 0.4s ease' }}>

      {/* Greeting */}
      <div>
        <p style={{ fontSize:13, color:'var(--t3)', marginBottom:4 }}>
          {greeting()}, <span style={{ color:'var(--t2)', fontWeight:600 }}>{store.user.name.split(' ')[0]}</span>
        </p>
        <h2 style={{ fontFamily:'Syne', fontWeight:700, fontSize:22 }}>
          Visão Geral do Portfólio
        </h2>
      </div>

      {/* Row 1: Balance + Risk */}
      <div style={{ display:'grid', gridTemplateColumns:'1fr 320px', gap:16 }}>
        <BalanceCard store={store} />
        <RiskGauge   store={store} />
      </div>

      {/* Quick Actions */}
      <div>
        <p style={{ fontSize:11, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.1em', marginBottom:12 }}>Ações Rápidas</p>
        <div style={{ display:'grid', gridTemplateColumns:'repeat(4, 1fr)', gap:10 }}>
          {QUICK_ACTIONS.map(action => (
            <button key={action.key}
              onClick={() => setSimModal({ action: action.key, title: action.title, fields: action.fields })}
              style={{
                background:`linear-gradient(135deg, ${action.bg}, var(--card))`,
                border:`1px solid ${action.border}`,
                borderRadius:16, padding:'16px 12px',
                cursor:'pointer', transition:'all 0.2s ease',
                display:'flex', flexDirection:'column', alignItems:'center', gap:10,
                textAlign:'center',
              }}
              onMouseEnter={e => { e.currentTarget.style.transform='translateY(-3px)'; e.currentTarget.style.boxShadow='0 10px 28px rgba(0,0,0,0.3)' }}
              onMouseLeave={e => { e.currentTarget.style.transform='translateY(0)'; e.currentTarget.style.boxShadow='none' }}>
              <div style={{ width:42, height:42, borderRadius:12, background: action.bg, border:`1px solid ${action.border}`, display:'flex', alignItems:'center', justifyContent:'center' }}>
                <action.Icon size={18} color={action.color} />
              </div>
              <div>
                <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:13, color:'var(--t1)', marginBottom:2 }}>{action.label}</p>
                <p style={{ fontSize:10, color:'var(--t3)' }}>{action.sub}</p>
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Row 2: Activity + Crypto */}
      <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:16 }}>
        <ActivityFeed
          transactions={store.transactions}
          onViewAll={() => showToast('Extrato completo em breve', 'info')}
        />
        <CryptoPortfolio
          crypto={store.crypto}
          onViewAll={() => showToast('Navegue para Investir para ver a carteira completa', 'info')}
        />
      </div>

      {/* Row 3: Cash Flow Chart */}
      <CashFlowChart cashFlow={store.cashFlow} />

      {/* Row 4: Quick stats */}
      <div style={{ display:'grid', gridTemplateColumns:'repeat(3, 1fr)', gap:16 }}>
        <StatCard label="KYC Level" value={store.kyc.level} sub="Limites ampliados" color="var(--green)" badge="badge-green" />
        <StatCard label="Plano Ativo" value={store.user.plan} sub="Pricing Engine ativo" color="var(--purple)" badge="badge-purple" />
        <StatCard label="Webhooks Pendentes" value={store.observability.pendingWebhooks} sub={`Latência ${store.observability.latencyMs}ms`} color="var(--blue)" badge="badge-blue" />
      </div>
    </div>
  )
}

function StatCard({ label, value, sub, color, badge }) {
  return (
    <div className="card hover-glow" style={{ padding:18, display:'flex', flexDirection:'column', gap:8 }}>
      <p style={{ fontSize:11, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.08em' }}>{label}</p>
      <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between' }}>
        <span style={{ fontFamily:'Syne', fontWeight:800, fontSize:22, color }}>{value}</span>
        <ChevronRight size={14} color='var(--t3)' />
      </div>
      <p style={{ fontSize:11, color:'var(--t3)' }}>{sub}</p>
    </div>
  )
}
