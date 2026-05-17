import React, { useState } from 'react'
import { Shield, CheckCircle, Globe, Star, Download, QrCode, AlertTriangle, Lock, Zap } from 'lucide-react'
import { useApp } from '../App.jsx'

/* Decorative QR – fixed pattern for demo */
const QR_CELLS = (() => {
  const pattern = [
    [1,1,1,1,1,1,1,0,1,0,1,0,0,0,1,1,1,1,1,1,1],
    [1,0,0,0,0,0,1,0,0,1,0,1,0,0,1,0,0,0,0,0,1],
    [1,0,1,1,1,0,1,0,1,0,1,0,1,0,1,0,1,1,1,0,1],
    [1,0,1,1,1,0,1,0,0,1,0,0,1,0,1,0,1,1,1,0,1],
    [1,0,1,1,1,0,1,0,1,1,1,0,1,0,1,0,1,1,1,0,1],
    [1,0,0,0,0,0,1,0,1,0,1,1,1,0,1,0,0,0,0,0,1],
    [1,1,1,1,1,1,1,0,1,0,1,0,1,0,1,1,1,1,1,1,1],
    [0,0,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0],
    [1,0,1,1,0,1,1,1,0,0,1,0,1,1,0,1,1,0,1,0,1],
    [0,1,0,0,1,0,0,0,1,0,0,1,0,0,1,0,0,1,0,1,0],
    [1,1,1,0,1,1,1,0,1,1,0,0,1,0,1,1,1,0,1,1,1],
    [0,0,0,1,0,0,0,1,0,0,1,0,0,1,0,0,0,1,0,0,0],
    [1,0,1,0,1,0,1,1,0,1,1,1,0,0,1,0,1,0,1,0,1],
    [0,0,0,0,0,0,0,0,1,0,0,1,0,1,0,0,0,0,0,0,0],
    [1,1,1,1,1,1,1,0,1,1,0,0,1,0,1,0,1,1,1,1,1],
    [1,0,0,0,0,0,1,0,0,0,1,0,0,0,0,1,0,0,0,0,1],
    [1,0,1,1,1,0,1,0,1,0,1,1,0,0,1,0,1,1,1,0,1],
    [1,0,1,1,1,0,1,0,0,1,0,0,1,0,0,0,0,0,0,1,1],
    [1,0,1,1,1,0,1,0,1,0,1,0,0,1,1,1,0,1,0,1,0],
    [1,0,0,0,0,0,1,0,0,1,0,1,0,0,0,0,1,0,1,0,1],
    [1,1,1,1,1,1,1,0,1,0,1,0,1,0,1,1,0,1,1,1,0],
  ]
  return pattern
})()

function QRCode({ size = 120, color = '#00E599' }) {
  const cells = QR_CELLS.length
  const cell = size / cells
  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
      <rect width={size} height={size} fill="rgba(0,0,0,0.3)" rx="8" />
      {QR_CELLS.map((row, r) =>
        row.map((v, c) => v
          ? <rect key={`${r}-${c}`} x={c * cell + 1} y={r * cell + 1} width={cell - 1} height={cell - 1} rx={cell * 0.15} fill={color} opacity={0.9} />
          : null
        )
      )}
    </svg>
  )
}

const JURISDICTIONS = [
  { code: 'BRA', name: 'Brasil', flag: '🇧🇷', status: 'ACTIVE', label: 'Habilitado' },
  { code: 'ARE', name: 'Emirados Árabes', flag: '🇦🇪', status: 'ACTIVE', label: 'Habilitado' },
  { code: 'USA', name: 'Estados Unidos', flag: '🇺🇸', status: 'PENDING', label: 'Em revisão' },
  { code: 'EUR', name: 'União Europeia', flag: '🇪🇺', status: 'PENDING', label: 'Em revisão' },
]

const PLAN_COLORS = {
  RETAIL:        { color: '#00E599', bg: 'rgba(0,229,153,0.12)' },
  BUSINESS:      { color: '#818CF8', bg: 'rgba(129,140,248,0.12)' },
  INSTITUTIONAL: { color: '#FBBF24', bg: 'rgba(251,191,36,0.12)' },
}

export default function ZPass() {
  const { store } = useApp()
  const [exported, setExported] = useState(false)
  const u = store.user
  const reg = store.regulatoryProfile
  const planCfg = PLAN_COLORS[store.persona] || PLAN_COLORS.BUSINESS
  const idHash = (u.id || '').replace(/-/g, '').slice(-8).toUpperCase().padStart(8, 'F')
  const zpId = `ZP-${idHash.slice(0, 4)}-${idHash.slice(4, 8)}`

  const verifiedAt = new Date(Date.now() - 60 * 86400000)
  const nextReview = new Date(Date.now() + 150 * 86400000)

  function handleExport() {
    const payload = {
      zpass_id: zpId,
      issued_at: new Date().toISOString(),
      issuer: 'Z-Finance Identity Service',
      version: '1.0',
      holder: {
        name: u.name,
        email: u.email,
        type: u.type,
        plan: store.persona,
      },
      verification: {
        kyc_level: u.kycLevel || 'FULL',
        kyc_status: 'VERIFIED',
        verified_at: verifiedAt.toISOString(),
        next_review: nextReview.toISOString(),
        vasp_active: true,
      },
      jurisdictions: JURISDICTIONS.map(j => ({ code: j.code, name: j.name, status: j.status })),
      regulatory_profile: {
        fatf_risk: reg?.fatfRisk || 'LOW',
        pep: reg?.pep ?? false,
        sanctions: reg?.sanctions ?? false,
        travel_rule: reg?.travelRule ?? true,
        license_emi: reg?.licenseEmi ?? true,
      },
      signature: `sha256:${idHash}-${Date.now().toString(36)}`,
    }
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `z-pass-${zpId}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    setExported(true)
    setTimeout(() => setExported(false), 2500)
  }

  return (
    <div>
      <div style={{ marginBottom: 20 }}>
        <h2 style={{ fontSize: 22, fontWeight: 800, fontFamily: 'Syne,sans-serif', color: 'var(--t1)', margin: '0 0 4px' }}>Z-Pass</h2>
        <p style={{ fontSize: 13, color: 'var(--t3)', margin: 0 }}>Identidade Financeira Digital</p>
      </div>

      {/* Passport card */}
      <div style={{
        background: 'linear-gradient(135deg, #0A1628 0%, #081020 50%, #0D1F14 100%)',
        border: `1px solid ${planCfg.color}35`,
        borderRadius: 24, padding: '28px 24px',
        marginBottom: 16, position: 'relative', overflow: 'hidden',
      }}>
        {/* Background shimmer */}
        <div style={{
          position: 'absolute', top: -60, right: -60, width: 250, height: 250, borderRadius: '50%',
          background: `radial-gradient(circle, ${planCfg.color}12 0%, transparent 70%)`,
          pointerEvents: 'none',
        }} />
        <div style={{
          position: 'absolute', bottom: -40, left: -40, width: 160, height: 160, borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(129,140,248,0.08) 0%, transparent 70%)',
          pointerEvents: 'none',
        }} />

        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 24, position: 'relative' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
              <div style={{ width: 8, height: 8, borderRadius: '50%', background: planCfg.color }} />
              <span style={{ fontSize: 10, fontWeight: 800, color: planCfg.color, textTransform: 'uppercase', letterSpacing: '0.12em' }}>
                Z-Finance · Passaporte Financeiro
              </span>
            </div>
            <p style={{ fontFamily: 'Syne,sans-serif', fontSize: 11, color: 'var(--t3)', margin: 0, letterSpacing: '0.04em' }}>
              IDENTIDADE FINANCEIRA DIGITAL
            </p>
          </div>
          <div style={{
            width: 40, height: 40, borderRadius: 12,
            background: `linear-gradient(135deg, ${planCfg.color}, ${planCfg.color}80)`,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <span style={{ fontFamily: 'Syne,sans-serif', fontWeight: 900, fontSize: 20, color: '#040C1B' }}>Z</span>
          </div>
        </div>

        {/* User info */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 24, position: 'relative' }}>
          <div style={{
            width: 64, height: 64, borderRadius: 18,
            background: 'linear-gradient(135deg, #1E3A5F, #0C1E38)',
            border: `2px solid ${planCfg.color}50`,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 24, fontWeight: 800, color: planCfg.color, flexShrink: 0,
          }}>
            {u.name.charAt(0)}
          </div>
          <div>
            <p style={{ fontFamily: 'Syne,sans-serif', fontSize: 20, fontWeight: 800, color: 'var(--t1)', margin: '0 0 4px' }}>{u.name}</p>
            <p style={{ fontSize: 12, color: 'var(--t3)', margin: '0 0 6px', fontFamily: 'DM Mono, monospace' }}>{u.email}</p>
            <span style={{
              fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em',
              color: planCfg.color, background: planCfg.bg,
              border: `1px solid ${planCfg.color}30`,
              borderRadius: 8, padding: '3px 10px',
            }}>
              {store.persona || 'BUSINESS'} Plan
            </span>
          </div>
        </div>

        {/* Z-Pass ID + QR */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', position: 'relative' }}>
          <div>
            <p style={{ fontSize: 10, color: 'var(--t3)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.1em', margin: '0 0 4px' }}>Z-Pass ID</p>
            <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 20, fontWeight: 800, color: planCfg.color, margin: '0 0 12px', letterSpacing: '0.04em' }}>
              {zpId}
            </p>
            <div style={{ display: 'flex', gap: 8 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                <Shield size={12} color="#34D399" />
                <span style={{ fontSize: 11, color: '#34D399', fontWeight: 700 }}>KYC {u.kycLevel}</span>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                <Zap size={12} color="#FBBF24" />
                <span style={{ fontSize: 11, color: '#FBBF24', fontWeight: 700 }}>VASP</span>
              </div>
            </div>
          </div>
          <QRCode size={90} color={planCfg.color} />
        </div>

        {/* Bottom stripe */}
        <div style={{ marginTop: 20, paddingTop: 16, borderTop: `1px solid ${planCfg.color}15` }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <p style={{ fontSize: 10, color: 'var(--t3)', margin: '0 0 2px', fontWeight: 600, textTransform: 'uppercase' }}>Verificado em</p>
              <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 12, color: 'var(--t2)', margin: 0 }}>
                {verifiedAt.toLocaleDateString('pt-BR')}
              </p>
            </div>
            <div style={{ textAlign: 'center' }}>
              <p style={{ fontSize: 10, color: 'var(--t3)', margin: '0 0 2px', fontWeight: 600, textTransform: 'uppercase' }}>Tipo</p>
              <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 12, color: 'var(--t2)', margin: 0 }}>{u.type}</p>
            </div>
            <div style={{ textAlign: 'right' }}>
              <p style={{ fontSize: 10, color: 'var(--t3)', margin: '0 0 2px', fontWeight: 600, textTransform: 'uppercase' }}>Próxima revisão</p>
              <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 12, color: 'var(--t2)', margin: 0 }}>
                {nextReview.toLocaleDateString('pt-BR')}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Compliance flags */}
      <div style={{
        background: 'var(--surface)', border: '1px solid var(--border)',
        borderRadius: 18, padding: '16px 20px', marginBottom: 16,
      }}>
        <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 14px' }}>Status de Compliance</p>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10 }}>
          {[
            { label: 'FATF Risk', value: reg?.fatfRisk || 'LOW', color: '#34D399', ok: true },
            { label: 'PEP', value: reg?.pep ? 'SIM' : 'NÃO', color: reg?.pep ? '#F87171' : '#34D399', ok: !reg?.pep },
            { label: 'Sanções', value: reg?.sanctions ? 'SIM' : 'NÃO', color: reg?.sanctions ? '#F87171' : '#34D399', ok: !reg?.sanctions },
            { label: 'VASP', value: reg?.vasp ? 'ATIVO' : 'NÃO', color: '#FBBF24', ok: true },
            { label: 'Licença', value: reg?.licenseType || 'EMI', color: '#60A5FA', ok: true },
            { label: 'KYC Level', value: u.kycLevel, color: '#818CF8', ok: true },
          ].map(({ label, value, color, ok }) => (
            <div key={label} style={{
              background: 'var(--bg)', borderRadius: 12, padding: '12px 10px', textAlign: 'center',
              border: `1px solid ${ok ? 'var(--border)' : 'rgba(248,113,113,0.3)'}`,
            }}>
              <div style={{ marginBottom: 4 }}>
                {ok
                  ? <CheckCircle size={14} style={{ color }} />
                  : <AlertTriangle size={14} style={{ color: '#F87171' }} />}
              </div>
              <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 12, fontWeight: 700, color, margin: '0 0 2px' }}>{value}</p>
              <p style={{ fontSize: 10, color: 'var(--t3)', margin: 0 }}>{label}</p>
            </div>
          ))}
        </div>
      </div>

      {/* Jurisdictions */}
      <div style={{
        background: 'var(--surface)', border: '1px solid var(--border)',
        borderRadius: 18, overflow: 'hidden', marginBottom: 16,
      }}>
        <div style={{ padding: '16px 20px 12px', borderBottom: '1px solid var(--border)' }}>
          <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: 0 }}>Jurisdições Habilitadas</p>
        </div>
        {JURISDICTIONS.map((j, i) => (
          <div key={j.code} style={{
            display: 'flex', alignItems: 'center', gap: 14, padding: '14px 20px',
            borderBottom: i < JURISDICTIONS.length - 1 ? '1px solid var(--border)' : 'none',
          }}>
            <span style={{ fontSize: 22, flexShrink: 0 }}>{j.flag}</span>
            <div style={{ flex: 1 }}>
              <p style={{ fontWeight: 600, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{j.name}</p>
              <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{j.code}</p>
            </div>
            <span style={{
              fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 8,
              color: j.status === 'ACTIVE' ? '#34D399' : '#FBBF24',
              background: j.status === 'ACTIVE' ? 'rgba(52,211,153,0.15)' : 'rgba(251,191,36,0.15)',
            }}>
              {j.label.toUpperCase()}
            </span>
          </div>
        ))}
      </div>

      {/* Export button */}
      <button
        onClick={handleExport}
        style={{
          width: '100%', padding: '16px', borderRadius: 16, cursor: 'pointer',
          background: exported
            ? 'rgba(52,211,153,0.15)'
            : 'linear-gradient(135deg, rgba(0,229,153,0.15), rgba(129,140,248,0.1))',
          border: `1px solid ${exported ? 'rgba(52,211,153,0.4)' : 'rgba(0,229,153,0.25)'}`,
          display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 10,
          color: exported ? '#34D399' : 'var(--accent)',
          fontWeight: 700, fontSize: 15,
          transition: 'all 0.3s ease',
        }}
      >
        {exported
          ? <><CheckCircle size={18} /> Passaporte Exportado!</>
          : <><Download size={18} /> Exportar Passaporte Financeiro Digital</>}
      </button>
    </div>
  )
}
