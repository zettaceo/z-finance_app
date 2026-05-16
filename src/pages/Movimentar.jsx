import { useState } from 'react'
import { runSimulation, fmtBRL, relativeTime, txTypeLabel } from '../data/mock.js'
import {
  Send, Download, QrCode, Receipt, RefreshCw,
  ArrowUpRight, CheckCircle2, Clock, XCircle,
  ChevronRight
} from 'lucide-react'

const TABS = [
  { key:'pix',      label:'PIX',         Icon: Send },
  { key:'payments', label:'Pagamentos',   Icon: Receipt },
  { key:'invoices', label:'Cobranças',    Icon: QrCode },
  { key:'transfers',label:'Transferências',Icon: RefreshCw },
]

export default function Movimentar({ store, updateStore, showToast }) {
  const [activeTab, setActiveTab] = useState('pix')
  const [loading,   setLoading]   = useState(false)
  const [lastResult, setLastResult] = useState(null)

  const simulate = async (action, data) => {
    setLoading(true)
    await new Promise(r => setTimeout(r, 700))
    const { store: s, message, details } = runSimulation(store, action, data)
    updateStore(s)
    setLastResult({ action, message, details, ts: new Date().toISOString() })
    showToast(message, 'success')
    setLoading(false)
  }

  return (
    <div style={{ display:'flex', flexDirection:'column', gap:20, animation:'slideUp 0.4s ease' }}>
      <div>
        <h2 style={{ fontFamily:'Syne', fontWeight:700, fontSize:22, marginBottom:4 }}>Movimentar</h2>
        <p style={{ fontSize:13, color:'var(--t2)' }}>Operações financeiras com Pricing Engine ativo</p>
      </div>

      {/* Balance pill */}
      <div style={{ background:'rgba(0,232,122,0.04)', border:'1px solid rgba(0,232,122,0.15)', borderRadius:12, padding:'12px 18px', display:'flex', justifyContent:'space-between', alignItems:'center' }}>
        <div>
          <p style={{ fontSize:11, color:'var(--t3)', marginBottom:2 }}>SALDO DISPONÍVEL</p>
          <p style={{ fontFamily:'Syne', fontWeight:800, fontSize:20, color:'var(--green)' }}>{fmtBRL(store?.account?.balance)}</p>
        </div>
        <div style={{ textAlign:'right' }}>
          <p style={{ fontSize:11, color:'var(--t3)', marginBottom:2 }}>BLOQUEADO</p>
          <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:14, color:'var(--t2)' }}>{fmtBRL(store?.account?.holds || 0)}</p>
        </div>
      </div>

      {/* Tabs */}
      <div style={{ display:'flex', gap:4, background:'rgba(255,255,255,0.02)', borderRadius:12, padding:4, overflow:'auto' }}>
        {TABS.map(({ key, label, Icon }) => (
          <button key={key} onClick={() => setActiveTab(key)}
            className={`tab-item ${activeTab === key ? 'active' : ''}`}
            style={{ display:'flex', alignItems:'center', gap:7, border:'none', cursor:'pointer', flexShrink:0 }}>
            <Icon size={13} />
            {label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:16 }}>
        {activeTab === 'pix' && (
          <>
            <ActionForm
              title="Enviar PIX"
              action="pix_send"
              simulate={simulate}
              loading={loading}
              icon={<Send size={16} color='var(--green)' />}
              fields={[
                { key:'receiver', label:'Chave PIX / Conta', placeholder:'CPF, email, telefone, EVP', required:true },
                { key:'amount',   label:'Valor (R$)', type:'number', placeholder:'250,00', step:'0.01', min:'0.01', required:true },
                { key:'description', label:'Descrição', placeholder:'PIX transferência', default:'PIX transferência' },
              ]}
            />
            <ActionForm
              title="Receber PIX"
              action="pix_receive"
              simulate={simulate}
              loading={loading}
              icon={<Download size={16} color='var(--blue)' />}
              fields={[
                { key:'sender', label:'Remetente', placeholder:'Nome do remetente', required:true },
                { key:'amount', label:'Valor esperado (R$)', type:'number', placeholder:'500,00', step:'0.01', min:'0.01', required:true },
              ]}
            />
            <ActionForm
              title="PIX via Cripto"
              action="pix_crypto"
              simulate={simulate}
              loading={loading}
              icon={<ArrowUpRight size={16} color='var(--gold)' />}
              fields={[
                { key:'usdt', label:'USDT a converter', type:'number', placeholder:'100', step:'1', min:'1', required:true, hint:'Taxa de spread: 1.2%' },
              ]}
            />
          </>
        )}

        {activeTab === 'payments' && (
          <>
            <ActionForm
              title="Agendar Pagamento"
              action="payment_schedule"
              simulate={simulate}
              loading={loading}
              icon={<Clock size={16} color='var(--blue)' />}
              fields={[
                { key:'barcode', label:'Código de Barras', placeholder:'00000.00000 00000...', required:true },
                { key:'amount',  label:'Valor (R$)', type:'number', placeholder:'120,00', step:'0.01', min:'0.01', required:true },
                { key:'dueDate', label:'Data de Vencimento', type:'date' },
              ]}
            />
            <ActionForm
              title="Confirmar Pagamento"
              action="payment_confirm"
              simulate={simulate}
              loading={loading}
              icon={<CheckCircle2 size={16} color='var(--green)' />}
              fields={[]}
            />
          </>
        )}

        {activeTab === 'invoices' && (
          <>
            <ActionForm
              title="Criar Cobrança (PIX + USDT)"
              action="invoice_create"
              simulate={simulate}
              loading={loading}
              icon={<QrCode size={16} color='var(--purple)' />}
              fields={[
                { key:'amount',      label:'Valor (R$)', type:'number', placeholder:'500,00', step:'0.01', min:'0.01', required:true },
                { key:'description', label:'Descrição', placeholder:'Serviço de consultoria' },
              ]}
            />
            <ActionForm
              title="Pagar Última Cobrança"
              action="invoice_pay"
              simulate={simulate}
              loading={loading}
              icon={<CheckCircle2 size={16} color='var(--green)' />}
              fields={[]}
              hint={store?.invoices?.length > 0
                ? `${store.invoices.length} cobrança(s) pendente(s)`
                : 'Crie uma cobrança primeiro'}
            />
          </>
        )}

        {activeTab === 'transfers' && (
          <>
            <ActionForm
              title="Transferência Interna"
              action="transfer_internal"
              simulate={simulate}
              loading={loading}
              icon={<RefreshCw size={16} color='var(--blue)' />}
              fields={[
                { key:'amount', label:'Valor (R$)', type:'number', placeholder:'1.000,00', step:'0.01', min:'0.01', required:true },
              ]}
            />
            <ActionForm
              title="Saque"
              action="withdrawal"
              simulate={simulate}
              loading={loading}
              icon={<ArrowUpRight size={16} color='var(--red)' />}
              fields={[
                { key:'pix',    label:'Chave PIX Destino', placeholder:'CPF ou email', required:true },
                { key:'amount', label:'Valor (R$)', type:'number', placeholder:'200,00', step:'0.01', min:'0.01', required:true },
              ]}
            />
          </>
        )}
      </div>

      {/* Recent Simulations */}
      <div className="card" style={{ padding:22 }}>
        <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:15, marginBottom:16 }}>Simulações Recentes</p>
        {store?.transactions?.length > 0 ? (
          <div>
            {store.transactions.slice(0,5).map(tx => (
              <div key={tx.id} className="tx-item">
                <div style={{ width:8, height:8, borderRadius:'50%', background: tx.amount >= 0 ? 'var(--green)' : 'var(--red)', flexShrink:0, marginLeft:6 }} />
                <div style={{ flex:1 }}>
                  <p style={{ fontSize:13, fontWeight:500 }}>{tx.description}</p>
                  <p style={{ fontSize:11, color:'var(--t3)' }}>{txTypeLabel(tx.type)} · {relativeTime(tx.createdAt)}</p>
                </div>
                <span style={{ fontFamily:'Syne', fontWeight:700, fontSize:13, color: tx.amount >= 0 ? 'var(--green)' : 'var(--red)' }}>
                  {tx.amount >= 0 ? '+' : ''}{fmtBRL(Math.abs(tx.amount))}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p style={{ fontSize:13, color:'var(--t3)', textAlign:'center', padding:'16px 0' }}>Execute simulações acima</p>
        )}
      </div>
    </div>
  )
}

function ActionForm({ title, action, simulate, loading, icon, fields, hint }) {
  const [form, setForm] = useState({})

  const set = (k, v) => setForm(prev => ({ ...prev, [k]: v }))

  const handleSubmit = (e) => {
    e.preventDefault()
    simulate(action, form)
    setForm({})
    e.target.reset()
  }

  return (
    <div className="card" style={{ padding:20 }}>
      <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:16 }}>
        <div style={{ width:32, height:32, borderRadius:8, background:'rgba(255,255,255,0.04)', border:'1px solid var(--border)', display:'flex', alignItems:'center', justifyContent:'center' }}>
          {icon}
        </div>
        <p style={{ fontWeight:700, fontSize:14, fontFamily:'Syne' }}>{title}</p>
      </div>

      {hint && (
        <p style={{ fontSize:11, color:'var(--t3)', marginBottom:12, padding:'6px 10px', background:'rgba(255,255,255,0.03)', borderRadius:8 }}>{hint}</p>
      )}

      <form onSubmit={handleSubmit} style={{ display:'flex', flexDirection:'column', gap:12 }}>
        {fields.map(f => (
          <div key={f.key}>
            <label style={{ fontSize:11, color:'var(--t2)', marginBottom:5, display:'block' }}>{f.label}</label>
            {f.type === 'select' ? (
              <select className="input" onChange={e => set(f.key, e.target.value)} defaultValue={f.default}>
                {(f.options||[]).map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            ) : (
              <input className="input" type={f.type||'text'} placeholder={f.placeholder} step={f.step} min={f.min} required={f.required} defaultValue={f.default}
                onChange={e => set(f.key, e.target.value)} />
            )}
            {f.hint && <p style={{ fontSize:10, color:'var(--t3)', marginTop:3 }}>{f.hint}</p>}
          </div>
        ))}
        {fields.length === 0 && (
          <p style={{ fontSize:12, color:'var(--t3)', padding:'8px 0' }}>Clique em executar para simular</p>
        )}
        <button type="submit" className="btn-primary" disabled={loading} style={{ marginTop:4 }}>
          {loading ? 'Processando...' : 'Executar'}
        </button>
      </form>
    </div>
  )
}
