import { useState } from 'react'
import { X, CheckCircle } from 'lucide-react'
import { runSimulation } from '../data/mock.js'

export default function SimModal({ config, store, updateStore, showToast, onClose }) {
  const [formData, setFormData] = useState({})
  const [result,   setResult]   = useState(null)
  const [loading,  setLoading]  = useState(false)

  const set = (k, v) => setFormData(prev => ({ ...prev, [k]: v }))

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    await new Promise(r => setTimeout(r, 600)) // simulate latency
    const { store: newStore, message, details } = runSimulation(store, config.action, formData)
    updateStore(newStore)
    setResult({ message, details })
    setLoading(false)
    showToast(message, 'success')
  }

  return (
    <div className="modal-overlay" onClick={e => e.target === e.currentTarget && !result && onClose()}>
      <div className="modal-box">
        {/* Header */}
        <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between', marginBottom:20 }}>
          <h3 style={{ fontFamily:'Syne', fontWeight:700, fontSize:17 }}>{config.title}</h3>
          <button className="btn-icon" onClick={onClose}><X size={15}/></button>
        </div>

        {result ? (
          <ResultView result={result} onClose={onClose} />
        ) : (
          <form onSubmit={handleSubmit} style={{ display:'flex', flexDirection:'column', gap:14 }}>
            {(config.fields || []).map(field => (
              <div key={field.key}>
                <label style={{ fontSize:12, color:'var(--t2)', marginBottom:6, display:'block' }}>{field.label}</label>
                {field.type === 'select' ? (
                  <select className="input" value={formData[field.key] || field.default || ''} onChange={e => set(field.key, e.target.value)}>
                    {(field.options || []).map(opt => (
                      <option key={opt.value} value={opt.value}>{opt.label}</option>
                    ))}
                  </select>
                ) : (
                  <input className="input" type={field.type || 'text'}
                    placeholder={field.placeholder || ''}
                    defaultValue={field.default || ''}
                    step={field.step}
                    min={field.min}
                    required={field.required}
                    onChange={e => set(field.key, e.target.value)}
                  />
                )}
                {field.hint && <p style={{ fontSize:11, color:'var(--t3)', marginTop:4 }}>{field.hint}</p>}
              </div>
            ))}

            <div style={{ display:'flex', gap:10, marginTop:8 }}>
              <button type="button" className="btn-ghost" style={{ flex:1 }} onClick={onClose}>Cancelar</button>
              <button type="submit" className="btn-primary" style={{ flex:2 }} disabled={loading}>
                {loading ? <LoadingDots /> : 'Executar Simulação'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}

function ResultView({ result, onClose }) {
  return (
    <div style={{ animation:'slideUp 0.25s ease' }}>
      <div style={{ display:'flex', flexDirection:'column', alignItems:'center', gap:12, marginBottom:20, padding:'16px 0' }}>
        <div style={{ width:48, height:48, borderRadius:'50%', background:'rgba(0,232,122,0.12)', border:'2px solid rgba(0,232,122,0.3)', display:'flex', alignItems:'center', justifyContent:'center' }}>
          <CheckCircle size={22} color='var(--green)' />
        </div>
        <p style={{ fontSize:14, fontWeight:600, textAlign:'center' }}>{result.message}</p>
      </div>

      {result.details && Object.keys(result.details).length > 0 && (
        <div className="result-box">
          {JSON.stringify(result.details, null, 2)}
        </div>
      )}

      <button className="btn-primary" style={{ width:'100%', marginTop:16 }} onClick={onClose}>
        Fechar
      </button>
    </div>
  )
}

function LoadingDots() {
  return (
    <span style={{ display:'flex', alignItems:'center', gap:4 }}>
      Processando
      <span className="typing-dot" style={{ background:'#07090f' }} />
      <span className="typing-dot" style={{ background:'#07090f', animationDelay:'0.2s' }} />
      <span className="typing-dot" style={{ background:'#07090f', animationDelay:'0.4s' }} />
    </span>
  )
}
