import React, { useState, useEffect, useCallback } from 'react'
import { X, ChevronRight, CheckCircle, AlertCircle, Loader } from 'lucide-react'
import { useApp } from '../App.jsx'

export default function SimModal({ config, onClose }) {
  const { dispatch, toast } = useApp()
  const [fields, setFields] = useState({})
  const [status, setStatus] = useState('idle') // idle | loading | success | error
  const [result, setResult] = useState(null)

  useEffect(() => {
    if (config?.defaults) setFields(config.defaults)
    else setFields({})
    setStatus('idle')
    setResult(null)
  }, [config])

  const handleSubmit = useCallback(async () => {
    setStatus('loading')
    await new Promise(r => setTimeout(r, 800))
    try {
      dispatch(config.action, { ...fields, ...config.extraData })
      setResult({ ok: true, msg: config.successMsg || 'Operação realizada com sucesso!' })
      setStatus('success')
      toast(config.successMsg || 'Operação realizada!', 'success')
    } catch (e) {
      setResult({ ok: false, msg: e.message })
      setStatus('error')
    }
  }, [config, fields, dispatch, toast])

  if (!config) return null

  return (
    <>
      <div
        onClick={onClose}
        style={{
          position: 'fixed', inset: 0, background: 'rgba(4,12,27,0.7)',
          backdropFilter: 'blur(8px)', zIndex: 200,
          animation: 'fadeIn 0.2s ease',
        }}
      />
      <div style={{
        position: 'fixed',
        bottom: 0, left: 0, right: 0,
        zIndex: 201,
        background: 'var(--surface)',
        borderRadius: '24px 24px 0 0',
        padding: '0 0 calc(var(--bnav-h) + var(--safe-bottom))',
        animation: 'slideUp 0.3s cubic-bezier(0.4,0,0.2,1)',
        maxHeight: '90dvh',
        overflowY: 'auto',
        border: '1px solid var(--border)',
        borderBottom: 'none',
      }} className="hide-scroll">
        {/* Handle */}
        <div style={{ display: 'flex', justifyContent: 'center', padding: '12px 0 4px' }}>
          <div style={{ width: 40, height: 4, borderRadius: 2, background: 'var(--border)' }} />
        </div>

        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 20px 20px' }}>
          <div>
            <p style={{ fontSize: 11, color: 'var(--accent)', fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', marginBottom: 4 }}>
              Simulação
            </p>
            <h3 style={{ fontSize: 20, fontWeight: 700, color: 'var(--t1)', margin: 0 }}>{config.title}</h3>
          </div>
          <button onClick={onClose} style={{
            width: 36, height: 36, borderRadius: '50%', background: 'var(--surface-2)',
            border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: 'var(--t2)',
          }}>
            <X size={18} />
          </button>
        </div>

        <div style={{ padding: '0 20px 24px' }}>
          {status === 'success' || status === 'error' ? (
            <div style={{ textAlign: 'center', padding: '32px 0' }}>
              {status === 'success'
                ? <CheckCircle size={56} style={{ color: 'var(--accent)', marginBottom: 16 }} />
                : <AlertCircle size={56} style={{ color: '#F87171', marginBottom: 16 }} />}
              <p style={{ fontSize: 18, fontWeight: 700, color: 'var(--t1)', marginBottom: 8 }}>
                {status === 'success' ? 'Sucesso!' : 'Erro'}
              </p>
              <p style={{ color: 'var(--t2)', fontSize: 14, marginBottom: 32 }}>{result?.msg}</p>
              <button onClick={onClose} style={{
                background: 'var(--accent)', color: '#040C1B', border: 'none',
                borderRadius: 14, padding: '14px 32px', fontSize: 15, fontWeight: 700,
                cursor: 'pointer', width: '100%',
              }}>
                Fechar
              </button>
            </div>
          ) : (
            <>
              {config.description && (
                <p style={{ color: 'var(--t2)', fontSize: 14, marginBottom: 24, lineHeight: 1.6 }}>{config.description}</p>
              )}

              {config.fields?.map(f => (
                <div key={f.key} style={{ marginBottom: 16 }}>
                  <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: 'var(--t2)', marginBottom: 6, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                    {f.label}
                  </label>
                  {f.type === 'select' ? (
                    <select
                      value={fields[f.key] ?? f.default ?? ''}
                      onChange={e => setFields(p => ({ ...p, [f.key]: e.target.value }))}
                      style={{ width: '100%', background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 12, padding: '12px 14px', fontSize: 15, color: 'var(--t1)', appearance: 'none' }}
                    >
                      {f.options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
                    </select>
                  ) : f.type === 'textarea' ? (
                    <textarea
                      value={fields[f.key] ?? f.default ?? ''}
                      onChange={e => setFields(p => ({ ...p, [f.key]: e.target.value }))}
                      rows={3}
                      placeholder={f.placeholder}
                      style={{ width: '100%', background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 12, padding: '12px 14px', fontSize: 15, color: 'var(--t1)', resize: 'none', fontFamily: 'inherit', boxSizing: 'border-box' }}
                    />
                  ) : (
                    <input
                      type={f.type || 'text'}
                      value={fields[f.key] ?? f.default ?? ''}
                      onChange={e => setFields(p => ({ ...p, [f.key]: f.type === 'number' ? Number(e.target.value) : e.target.value }))}
                      placeholder={f.placeholder}
                      step={f.step}
                      min={f.min}
                      max={f.max}
                      style={{ width: '100%', background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: 12, padding: '12px 14px', fontSize: 15, color: 'var(--t1)', boxSizing: 'border-box' }}
                    />
                  )}
                  {f.hint && <p style={{ fontSize: 11, color: 'var(--t3)', marginTop: 4 }}>{f.hint}</p>}
                </div>
              ))}

              <button
                onClick={handleSubmit}
                disabled={status === 'loading'}
                style={{
                  width: '100%', background: 'var(--accent)', color: '#040C1B',
                  border: 'none', borderRadius: 14, padding: '16px', fontSize: 15,
                  fontWeight: 700, cursor: status === 'loading' ? 'wait' : 'pointer',
                  display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
                  marginTop: 8,
                  opacity: status === 'loading' ? 0.7 : 1,
                  transition: 'opacity 0.2s',
                }}
              >
                {status === 'loading' ? (
                  <><Loader size={18} style={{ animation: 'spin 1s linear infinite' }} /> Processando...</>
                ) : (
                  <>{config.submitLabel || 'Confirmar'} <ChevronRight size={18} /></>
                )}
              </button>
            </>
          )}
        </div>
      </div>
    </>
  )
}
