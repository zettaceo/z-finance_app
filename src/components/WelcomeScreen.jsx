import React, { useState, useEffect } from 'react'
import { CheckCircle, Zap, X } from 'lucide-react'
import { useApp } from '../App.jsx'

const STEPS = [
  {
    id: 'zpass',
    icon: '🪪',
    title: 'Z-Pass criado',
    sub: 'Identidade financeira digital ativa',
    color: '#FBBF24',
    duration: 1200,
  },
  {
    id: 'accounts',
    icon: '🌍',
    title: 'Contas abertas',
    sub: 'BRL, USD e AED prontas para uso',
    color: '#00E599',
    duration: 1200,
  },
  {
    id: 'kyc',
    icon: '🛡️',
    title: 'KYC: FULL verificado',
    sub: 'Compliance FATF LOW · Não-PEP · Não-sancionado',
    color: '#34D399',
    duration: 1200,
  },
  {
    id: 'ready',
    icon: '✨',
    title: 'Plataforma global pronta',
    sub: '',
    color: '#818CF8',
    duration: 900,
  },
]

export default function WelcomeScreen({ onDone }) {
  const { store } = useApp()
  const [step, setStep] = useState(-1)
  const [done, setDone] = useState(false)
  const firstName = store.user.name.split(' ')[0]

  useEffect(() => {
    // Start after small delay
    const t0 = setTimeout(() => setStep(0), 400)
    return () => clearTimeout(t0)
  }, [])

  useEffect(() => {
    if (step < 0 || step >= STEPS.length) return
    const t = setTimeout(() => {
      if (step < STEPS.length - 1) {
        setStep(s => s + 1)
      } else {
        setDone(true)
      }
    }, STEPS[step].duration)
    return () => clearTimeout(t)
  }, [step])

  return (
    <div style={{
      position: 'fixed', inset: 0, zIndex: 500,
      background: '#040C1B',
      display: 'flex', flexDirection: 'column',
      alignItems: 'center', justifyContent: 'center',
      padding: '24px',
      animation: 'fadeIn 0.3s ease',
    }}>
      {/* Skip button */}
      <button
        onClick={onDone}
        style={{
          position: 'absolute', top: 20, right: 20,
          width: 36, height: 36, borderRadius: 10, border: 'none',
          background: 'rgba(255,255,255,0.06)', cursor: 'pointer',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          color: 'var(--t3)',
        }}
      >
        <X size={16} />
      </button>

      {/* Logo */}
      <div style={{
        width: 72, height: 72, borderRadius: 22, marginBottom: 40,
        background: 'linear-gradient(135deg, var(--accent), var(--accent-2))',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        boxShadow: '0 0 60px rgba(0,229,153,0.25)',
        animation: step >= 0 ? 'glow 2s ease-in-out infinite' : 'none',
      }}>
        <Zap size={36} color="#040C1B" fill="#040C1B" />
      </div>

      {/* Steps */}
      <div style={{ width: '100%', maxWidth: 340, marginBottom: 40 }}>
        {STEPS.map((s, i) => {
          const visible = step >= i
          const current = step === i && !done
          const completed = step > i || done

          return (
            <div key={s.id} style={{
              display: 'flex', alignItems: 'center', gap: 16,
              padding: '14px 16px', borderRadius: 16,
              marginBottom: 10,
              background: visible ? `${s.color}10` : 'rgba(255,255,255,0.03)',
              border: `1px solid ${visible ? `${s.color}30` : 'rgba(255,255,255,0.05)'}`,
              opacity: visible ? 1 : 0.3,
              transform: visible ? 'translateY(0) scale(1)' : 'translateY(8px) scale(0.98)',
              transition: 'all 0.4s cubic-bezier(0.4,0,0.2,1)',
            }}>
              <div style={{
                width: 44, height: 44, borderRadius: 14, flexShrink: 0,
                background: visible ? `${s.color}20` : 'rgba(255,255,255,0.04)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: 22,
              }}>
                {completed ? <CheckCircle size={22} color={s.color} /> : s.icon}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <p style={{
                  fontWeight: 700, fontSize: 14,
                  color: visible ? 'var(--t1)' : 'var(--t3)',
                  margin: '0 0 2px',
                }}>
                  {s.title}
                </p>
                {s.sub && (
                  <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0, lineHeight: 1.4 }}>
                    {s.sub}
                  </p>
                )}
              </div>
              {current && (
                <div style={{ display: 'flex', gap: 3, alignItems: 'center', flexShrink: 0 }}>
                  {[0, 1, 2].map(k => (
                    <div key={k} style={{
                      width: 5, height: 5, borderRadius: '50%', background: s.color,
                      animation: `dotBounce 0.9s ease-in-out ${k * 0.15}s infinite`,
                    }} />
                  ))}
                </div>
              )}
            </div>
          )
        })}
      </div>

      {/* Welcome message */}
      <div style={{
        textAlign: 'center',
        opacity: done ? 1 : 0,
        transform: done ? 'translateY(0)' : 'translateY(16px)',
        transition: 'all 0.6s cubic-bezier(0.4,0,0.2,1)',
      }}>
        <h2 style={{
          fontFamily: 'Syne, sans-serif',
          fontWeight: 800, fontSize: 'clamp(20px, 6vw, 28px)',
          color: 'var(--t1)', margin: '0 0 8px',
          lineHeight: 1.2,
        }}>
          Bem-vindo, {firstName}.
        </h2>
        <p style={{ fontSize: 15, color: 'var(--t3)', margin: '0 0 32px', lineHeight: 1.5 }}>
          Sua plataforma financeira global está pronta.
        </p>
        <button
          onClick={onDone}
          style={{
            padding: '16px 48px', borderRadius: 16, border: 'none', cursor: 'pointer',
            background: 'linear-gradient(135deg, var(--accent), var(--accent-2))',
            color: '#040C1B', fontWeight: 800, fontSize: 16,
            fontFamily: 'Syne, sans-serif',
            boxShadow: '0 8px 32px rgba(0,229,153,0.3)',
            transition: 'all 0.2s ease',
          }}
          onMouseEnter={e => e.currentTarget.style.transform = 'translateY(-2px)'}
          onMouseLeave={e => e.currentTarget.style.transform = 'translateY(0)'}
        >
          Começar →
        </button>
      </div>
    </div>
  )
}
