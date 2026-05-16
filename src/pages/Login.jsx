import React, { useState, useCallback } from 'react'
import { Delete } from 'lucide-react'

export default function Login({ onLogin }) {
  const [pin, setPin] = useState('')
  const [shake, setShake] = useState(false)

  const press = useCallback((d) => {
    if (pin.length >= 6) return
    setPin(p => p + d)
  }, [pin])

  const del = useCallback(() => setPin(p => p.slice(0, -1)), [])

  const submit = useCallback(() => {
    if (pin.length < 4) { triggerShake(); return }
    onLogin(pin)
    setPin('')
  }, [pin, onLogin])

  const triggerShake = () => {
    setShake(true)
    setTimeout(() => { setShake(false); setPin('') }, 600)
  }

  const handleKey = useCallback((e) => {
    if (e.key >= '0' && e.key <= '9') press(e.key)
    else if (e.key === 'Backspace') del()
    else if (e.key === 'Enter') submit()
  }, [press, del, submit])

  React.useEffect(() => {
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [handleKey])

  const pad = ['1','2','3','4','5','6','7','8','9','','0','⌫']

  return (
    <div style={{
      minHeight: '100dvh',
      background: 'var(--bg)',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      position: 'relative',
      overflow: 'hidden',
      padding: '24px 24px calc(24px + var(--safe-bottom))',
    }}>
      {/* Background orbs */}
      <div style={{ position: 'absolute', inset: 0, overflow: 'hidden', pointerEvents: 'none' }}>
        <div style={{
          position: 'absolute', width: 500, height: 500,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(0,229,153,0.12) 0%, transparent 70%)',
          top: '-15%', left: '-10%',
          animation: 'float 8s ease-in-out infinite',
        }} />
        <div style={{
          position: 'absolute', width: 400, height: 400,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(99,102,241,0.1) 0%, transparent 70%)',
          bottom: '-10%', right: '-5%',
          animation: 'float 10s ease-in-out infinite reverse',
        }} />
        <div style={{
          position: 'absolute', width: 200, height: 200,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(251,191,36,0.07) 0%, transparent 70%)',
          top: '40%', right: '15%',
          animation: 'float 6s ease-in-out infinite',
        }} />
      </div>

      {/* Grid overlay */}
      <div style={{
        position: 'absolute', inset: 0,
        backgroundImage: 'linear-gradient(rgba(0,229,153,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(0,229,153,0.03) 1px, transparent 1px)',
        backgroundSize: '40px 40px',
        pointerEvents: 'none',
      }} />

      <div style={{ position: 'relative', width: '100%', maxWidth: 360, textAlign: 'center' }}>
        {/* Logo */}
        <div style={{ marginBottom: 40 }}>
          <div style={{
            display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
            width: 72, height: 72,
            background: 'linear-gradient(135deg, var(--accent), var(--accent-2))',
            borderRadius: 20,
            marginBottom: 20,
            boxShadow: '0 0 40px rgba(0,229,153,0.3)',
          }}>
            <span style={{ fontSize: 28, fontFamily: 'Syne, sans-serif', fontWeight: 900, color: '#040C1B', letterSpacing: '-1px' }}>Z</span>
          </div>
          <h1 style={{
            fontFamily: 'Syne, sans-serif',
            fontWeight: 900,
            fontSize: 28,
            color: 'var(--t1)',
            margin: '0 0 4px',
            letterSpacing: '-0.5px',
          }}>Z-Finance</h1>
          <p style={{ color: 'var(--t2)', fontSize: 14, margin: 0 }}>Global Banking Platform</p>
        </div>

        {/* PIN dots */}
        <div style={{
          display: 'flex', justifyContent: 'center', gap: 16,
          marginBottom: 40,
          animation: shake ? 'shake 0.5s ease' : 'none',
        }}>
          {[...Array(6)].map((_, i) => (
            <div key={i} style={{
              width: 14, height: 14, borderRadius: '50%',
              background: i < pin.length ? 'var(--accent)' : 'var(--surface-2)',
              border: i < pin.length ? 'none' : '2px solid var(--border)',
              transition: 'all 0.15s ease',
              boxShadow: i < pin.length ? '0 0 10px rgba(0,229,153,0.5)' : 'none',
            }} />
          ))}
        </div>

        <p style={{ color: 'var(--t3)', fontSize: 13, marginBottom: 24 }}>
          Digite seu PIN de acesso
        </p>

        {/* Numpad */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gap: 12,
          marginBottom: 24,
        }}>
          {pad.map((k, i) => {
            if (k === '') return <div key={i} />
            if (k === '⌫') return (
              <button key={i} onClick={del} style={{
                height: 64, borderRadius: 16,
                background: 'var(--surface-2)',
                border: '1px solid var(--border)',
                color: 'var(--t2)', cursor: 'pointer',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                transition: 'all 0.15s ease',
                fontSize: 18,
              }}>
                <Delete size={20} />
              </button>
            )
            return (
              <button key={i} onClick={() => press(k)} style={{
                height: 64, borderRadius: 16,
                background: 'var(--surface-2)',
                border: '1px solid var(--border)',
                color: 'var(--t1)', cursor: 'pointer',
                fontSize: 22, fontWeight: 600,
                fontFamily: 'DM Mono, monospace',
                transition: 'all 0.12s ease',
              }}
              onMouseDown={e => e.currentTarget.style.transform = 'scale(0.95)'}
              onMouseUp={e => e.currentTarget.style.transform = 'scale(1)'}
              >
                {k}
              </button>
            )
          })}
        </div>

        {/* Enter */}
        <button
          onClick={submit}
          disabled={pin.length < 4}
          style={{
            width: '100%', height: 56,
            background: pin.length >= 4
              ? 'linear-gradient(135deg, var(--accent), var(--accent-2))'
              : 'var(--surface-2)',
            border: 'none', borderRadius: 16,
            color: pin.length >= 4 ? '#040C1B' : 'var(--t3)',
            fontSize: 15, fontWeight: 700,
            cursor: pin.length >= 4 ? 'pointer' : 'default',
            transition: 'all 0.2s ease',
            boxShadow: pin.length >= 4 ? '0 4px 20px rgba(0,229,153,0.3)' : 'none',
            letterSpacing: '0.03em',
          }}
        >
          Acessar conta
        </button>

        <p style={{ marginTop: 20, fontSize: 12, color: 'var(--t3)' }}>
          Demo: qualquer PIN com 4+ dígitos
        </p>
      </div>
    </div>
  )
}
