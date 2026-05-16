import React, { useState } from 'react'
import {
  Crown, Activity, RefreshCw, Bell, GitBranch, Tag, Globe,
  CheckCircle, AlertTriangle, ChevronRight, Server, Archive,
  ToggleRight, BarChart3, UserX, Plus, Trash2, GitCommit,
} from 'lucide-react'
import { useApp } from '../App.jsx'
import SimModal from '../components/SimModal.jsx'

function fmt(cents) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100)
}
function fmtDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('pt-BR', { day: '2-digit', month: 'short', year: '2-digit' })
}

const SECTIONS = [
  { id: 'admin',       label: 'Admin',           Icon: Crown,      color: '#FBBF24' },
  { id: 'obs',         label: 'Observabilidade', Icon: Activity,   color: '#A78BFA' },
  { id: 'reconcile',   label: 'Reconciliação',   Icon: RefreshCw,  color: '#34D399' },
  { id: 'alerts',      label: 'Alertas',         Icon: Bell,       color: '#F87171' },
  { id: 'roles',       label: 'RBAC / Roles',    Icon: GitBranch,  color: '#818CF8' },
  { id: 'pricing_adv', label: 'Precificação',    Icon: Tag,        color: '#F59E0B' },
  { id: 'regulatory',  label: 'Regulatório',     Icon: Globe,      color: '#60A5FA' },
]

export default function AdminConsole() {
  const { store, dispatch, toast, modal, setModal } = useApp()
  const [activeSection, setActiveSection] = useState('admin')

  const obs = store.observability

  return (
    <div>
      <div style={{ marginBottom: 20 }}>
        <h2 style={{ fontSize: 22, fontWeight: 800, fontFamily: 'Syne,sans-serif', color: '#F1F5F9', margin: '0 0 4px' }}>Backoffice</h2>
        <p style={{ fontSize: 13, color: '#475569', margin: 0 }}>Operações internas · todas as ações são auditadas</p>
      </div>

      {/* Section tabs */}
      <div style={{ display: 'flex', gap: 8, overflowX: 'auto', paddingBottom: 8, marginBottom: 20 }} className="hide-scroll">
        {SECTIONS.map(({ id, label, Icon, color }) => {
          const active = activeSection === id
          return (
            <button key={id} onClick={() => setActiveSection(id)} style={{
              flexShrink: 0, display: 'flex', alignItems: 'center', gap: 6,
              padding: '8px 14px', borderRadius: 20, cursor: 'pointer',
              background: active ? `${color}20` : 'rgba(255,255,255,0.04)',
              border: `1px solid ${active ? `${color}50` : 'rgba(255,255,255,0.08)'}`,
              color: active ? color : '#64748B',
              fontSize: 12, fontWeight: 700, transition: 'all 0.15s',
            }}>
              <Icon size={13} />
              {label}
            </button>
          )
        })}
      </div>

      {/* ── Admin ── */}
      {activeSection === 'admin' && (
        <div style={{ display: 'grid', gap: 12 }}>
          <div style={{
            background: 'rgba(251,191,36,0.07)',
            border: '1px solid rgba(251,191,36,0.25)',
            borderRadius: 16, padding: '16px 20px',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
              <Crown size={16} color="#FBBF24" />
              <p style={{ fontWeight: 700, fontSize: 14, color: '#FBBF24', margin: 0 }}>Painel Administrativo</p>
            </div>
            <p style={{ fontSize: 12, color: '#64748B', margin: 0 }}>
              Acesso privilegiado: RBAC role ADMIN. Todas as ações são auditadas.
            </p>
          </div>

          {[
            {
              label: 'Alterar Plano do Usuário', Icon: Crown, color: '#FBBF24', sub: 'Upgrade/downgrade de plano',
              action: 'admin_plan_change', successMsg: 'Plano alterado!',
              fields: [
                { key: 'userId', label: 'ID do usuário', type: 'text', default: 'USR-001' },
                { key: 'plan', label: 'Novo plano', type: 'select', options: [{ value: 'RETAIL', label: 'Retail' }, { value: 'BUSINESS', label: 'Business' }, { value: 'INSTITUTIONAL', label: 'Institutional' }] },
              ],
            },
            {
              label: 'Toggle de Feature', Icon: ToggleRight, color: '#818CF8', sub: 'Ativar/desativar funcionalidade',
              action: 'admin_feature_toggle', successMsg: 'Feature atualizada!',
              fields: [
                { key: 'feature', label: 'Feature', type: 'select', options: [{ value: 'crypto_swap', label: 'Crypto Swap' }, { value: 'pix_international', label: 'PIX Internacional' }, { value: 'jit_card', label: 'Cartão JIT' }, { value: 'auto_invest', label: 'Auto Invest' }] },
                { key: 'enabled', label: 'Estado', type: 'select', options: [{ value: 'true', label: 'Ativado' }, { value: 'false', label: 'Desativado' }] },
              ],
            },
            {
              label: 'Ajustar Limite', Icon: BarChart3, color: '#34D399', sub: 'Modificar limites por usuário',
              action: 'admin_limit_adjust', successMsg: 'Limite ajustado!',
              fields: [
                { key: 'userId', label: 'ID do usuário', type: 'text', default: 'USR-001' },
                { key: 'limitType', label: 'Tipo de limite', type: 'select', options: [{ value: 'daily_pix', label: 'PIX Diário' }, { value: 'monthly_pix', label: 'PIX Mensal' }, { value: 'card_limit', label: 'Limite Cartão' }] },
                { key: 'newLimit', label: 'Novo limite (centavos)', type: 'number', default: 50000000 },
              ],
            },
            {
              label: 'Bloquear Usuário', Icon: UserX, color: '#F87171', sub: 'Suspender conta de usuário',
              action: 'admin_user_block', successMsg: 'Usuário bloqueado!',
              fields: [
                { key: 'userId', label: 'ID do usuário', type: 'text', default: 'USR-999' },
                { key: 'reason', label: 'Motivo', type: 'select', options: [{ value: 'fraud', label: 'Fraude confirmada' }, { value: 'kyc_fail', label: 'Falha no KYC' }, { value: 'compliance', label: 'Ordem de compliance' }, { value: 'request', label: 'Solicitação do cliente' }] },
              ],
            },
          ].map(({ label, Icon, color, sub, action, successMsg, fields }) => (
            <button key={label} onClick={() => setModal({ title: label, action, successMsg, fields, submitLabel: 'Executar' })} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px',
              borderRadius: 16, border: '1px solid rgba(255,255,255,0.07)', background: 'rgba(255,255,255,0.03)',
              cursor: 'pointer', textAlign: 'left',
            }}>
              <div style={{ width: 40, height: 40, borderRadius: 12, background: `${color}15`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Icon size={18} color={color} />
              </div>
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 12, color: '#64748B', margin: 0 }}>{sub}</p>
              </div>
              <ChevronRight size={16} color="#475569" />
            </button>
          ))}
        </div>
      )}

      {/* ── Observabilidade ── */}
      {activeSection === 'obs' && (
        <div style={{ display: 'grid', gap: 12 }}>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 10 }}>
            {[
              { label: 'Uptime', value: obs.uptime, sub: '30 dias', color: '#34D399', good: true },
              { label: 'Latência P99', value: `${obs.latencyP99}ms`, sub: 'Tempo resposta', color: obs.latencyP99 < 100 ? '#34D399' : '#F59E0B', good: obs.latencyP99 < 100 },
              { label: 'Webhooks', value: `${obs.pendingWebhooks}`, sub: 'Pendentes', color: obs.pendingWebhooks > 5 ? '#F87171' : '#60A5FA', good: obs.pendingWebhooks < 5 },
              { label: 'Status', value: 'Operacional', sub: 'Sistema', color: '#34D399', good: true },
            ].map(({ label, value, sub, color, good }) => (
              <div key={label} style={{ background: 'rgba(255,255,255,0.03)', border: `1px solid ${good ? 'rgba(255,255,255,0.07)' : `${color}40`}`, borderRadius: 16, padding: '16px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ fontSize: 11, color: '#475569', textTransform: 'uppercase', fontWeight: 600 }}>{label}</span>
                  <div style={{ width: 8, height: 8, borderRadius: '50%', background: color, marginTop: 2 }} />
                </div>
                <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 18, fontWeight: 700, color, margin: '0 0 2px' }}>{value}</p>
                <p style={{ fontSize: 11, color: '#475569', margin: 0 }}>{sub}</p>
              </div>
            ))}
          </div>

          {[
            {
              label: 'Dashboard de Saúde', Icon: Activity, color: '#A78BFA', sub: 'Ver métricas em tempo real',
              action: 'obs_summary', successMsg: 'Sumário obtido!',
              fields: [{ key: 'period', label: 'Período', type: 'select', options: [{ value: '1h', label: '1 hora' }, { value: '24h', label: '24 horas' }, { value: '7d', label: '7 dias' }] }],
            },
            {
              label: 'Reprocessar Webhook', Icon: Server, color: '#60A5FA', sub: 'Reenviar webhook com falha',
              action: 'webhook_retry', successMsg: 'Webhook reenviado!',
              fields: [
                { key: 'webhookId', label: 'ID do webhook', type: 'text', default: 'WH-' + Date.now().toString().slice(-6) },
                { key: 'maxRetries', label: 'Máximo de tentativas', type: 'number', default: 3 },
              ],
            },
            {
              label: 'Arquivar Audit Log', Icon: Archive, color: '#94A3B8', sub: 'Arquivar logs de auditoria',
              action: 'audit_archive', successMsg: 'Logs arquivados!',
              fields: [
                { key: 'before', label: 'Arquivar antes de', type: 'text', default: '2026-01-01', placeholder: 'YYYY-MM-DD' },
                { key: 'compress', label: 'Comprimir', type: 'select', options: [{ value: 'true', label: 'Sim (gzip)' }, { value: 'false', label: 'Não' }] },
              ],
            },
          ].map(({ label, Icon, color, sub, action, successMsg, fields }) => (
            <button key={label} onClick={() => setModal({ title: label, action, successMsg, fields, submitLabel: 'Executar' })} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px',
              borderRadius: 16, border: '1px solid rgba(255,255,255,0.07)', background: 'rgba(255,255,255,0.03)',
              cursor: 'pointer', textAlign: 'left',
            }}>
              <div style={{ width: 40, height: 40, borderRadius: 12, background: `${color}15`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Icon size={18} color={color} />
              </div>
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 12, color: '#64748B', margin: 0 }}>{sub}</p>
              </div>
              <ChevronRight size={16} color="#475569" />
            </button>
          ))}

          {store.auditLogs?.length > 0 && (
            <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
              <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)' }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Audit Trail recente</p>
              </div>
              {store.auditLogs.slice(0, 5).map((log, i) => (
                <div key={i} style={{ padding: '12px 20px', borderBottom: i < 4 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8 }}>
                    <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 12, color: '#34D399', margin: 0 }}>{log.action}</p>
                    <span style={{ fontSize: 10, color: '#475569', flexShrink: 0 }}>{log.ts || log.timestamp}</span>
                  </div>
                  {log.details && <p style={{ fontSize: 11, color: '#475569', margin: '2px 0 0' }}>{JSON.stringify(log.details).slice(0, 80)}...</p>}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* ── Reconciliação ── */}
      {activeSection === 'reconcile' && (
        <div style={{ display: 'grid', gap: 12 }}>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 10 }}>
            {[
              { label: 'PIX pendentes',     value: store.reconcile.summary.pix,     color: '#F59E0B' },
              { label: 'Boletos pendentes', value: store.reconcile.summary.payment, color: '#F87171' },
              { label: 'Card pendentes',    value: store.reconcile.summary.card,    color: '#34D399' },
              { label: 'Webhook dead',      value: store.reconcile.summary.webhook, color: '#818CF8' },
            ].map(({ label, value, color }) => (
              <div key={label} style={{ background: 'rgba(255,255,255,0.03)', border: `1px solid ${value > 0 ? color + '40' : 'rgba(255,255,255,0.07)'}`, borderRadius: 14, padding: '14px' }}>
                <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 24, fontWeight: 800, color: value > 0 ? color : '#F1F5F9', margin: '0 0 4px' }}>{value}</p>
                <p style={{ fontSize: 11, color: '#475569', margin: 0 }}>{label}</p>
              </div>
            ))}
          </div>

          <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Pendências</p>
              <button onClick={() => setModal({ title: 'Verificar Reconciliação', action: 'reconcile_summary', successMsg: 'Reconciliação verificada!', fields: [{ key: 'olderThanMinutes', label: 'Mais velhos que (min)', type: 'number', default: 30 }], submitLabel: 'Verificar' })} style={{ fontSize: 12, color: '#34D399', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 700 }}>
                Atualizar
              </button>
            </div>
            {store.reconcile.pending.map((p, i) => (
              <div key={p.id} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '14px 20px', borderBottom: i < store.reconcile.pending.length - 1 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                <div style={{ width: 8, height: 8, borderRadius: '50%', background: p.status === 'DEAD' ? '#F87171' : '#F59E0B', flexShrink: 0 }} />
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 13, fontWeight: 600, color: '#F1F5F9', margin: '0 0 2px' }}>{p.desc}</p>
                  <p style={{ fontSize: 11, color: '#475569', margin: 0 }}>{p.type.toUpperCase()} · {fmtDate(p.createdAt)}</p>
                </div>
                {p.amount > 0 && <span style={{ fontFamily: 'DM Mono, monospace', fontSize: 13, color: '#94A3B8' }}>{fmt(p.amount)}</span>}
                <button onClick={() => { dispatch('reconcile_resolve', { pendingId: p.id }); toast('Pendência resolvida!', 'success') }} style={{ fontSize: 11, color: '#34D399', background: 'rgba(52,211,153,0.1)', border: '1px solid rgba(52,211,153,0.2)', borderRadius: 8, padding: '4px 10px', cursor: 'pointer', fontWeight: 700 }}>
                  Resolver
                </button>
              </div>
            ))}
            {store.reconcile.pending.length === 0 && (
              <div style={{ padding: '32px', textAlign: 'center' }}>
                <CheckCircle size={32} color="#34D399" style={{ margin: '0 auto 8px', display: 'block' }} />
                <p style={{ color: '#94A3B8', fontSize: 14, fontWeight: 600, margin: 0 }}>Nenhuma pendência</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── Alertas ── */}
      {activeSection === 'alerts' && (
        <div style={{ display: 'grid', gap: 12 }}>
          <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Status dos Alertas</p>
              <button onClick={() => { dispatch('alerts_check', {}); toast('Alertas verificados!', 'success') }} style={{ fontSize: 12, color: '#34D399', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 700 }}>
                Checar agora
              </button>
            </div>
            {store.alerts.last.map((a, i) => (
              <div key={a.type} style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '14px 20px', borderBottom: i < store.alerts.last.length - 1 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                {a.status === 'OK'
                  ? <CheckCircle size={18} color="#34D399" style={{ flexShrink: 0 }} />
                  : <AlertTriangle size={18} color="#F87171" style={{ flexShrink: 0 }} />}
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 13, fontWeight: 600, color: '#F1F5F9', margin: '0 0 2px', fontFamily: 'DM Mono, monospace' }}>{a.type}</p>
                  <p style={{ fontSize: 11, color: '#475569', margin: 0 }}>Atual: {a.value} / Limite: {a.threshold}</p>
                </div>
                <div style={{ height: 4, width: 60, background: 'rgba(255,255,255,0.07)', borderRadius: 2, overflow: 'hidden' }}>
                  <div style={{ height: '100%', width: `${Math.min((a.value / a.threshold) * 100, 100)}%`, background: a.status === 'OK' ? '#34D399' : '#F87171', borderRadius: 2 }} />
                </div>
                <span style={{ fontSize: 10, fontWeight: 700, color: a.status === 'OK' ? '#34D399' : '#F87171', background: a.status === 'OK' ? 'rgba(52,211,153,0.1)' : 'rgba(248,113,113,0.1)', borderRadius: 6, padding: '2px 8px' }}>{a.status}</span>
              </div>
            ))}
          </div>

          <button onClick={() => setModal({
            title: 'Atualizar Threshold', action: 'alerts_update_threshold', successMsg: 'Threshold atualizado!',
            fields: [
              { key: 'type', label: 'Tipo de alerta', type: 'select', options: [{ value: 'pixPending', label: 'PIX Pendentes' }, { value: 'paymentPending', label: 'Boletos Pendentes' }, { value: 'cardPending', label: 'Card Pendentes' }, { value: 'webhookDead', label: 'Webhook Dead' }, { value: 'txHold', label: 'TX em Hold' }] },
              { key: 'value', label: 'Novo threshold', type: 'number', default: 10 },
            ],
            submitLabel: 'Atualizar',
          })} style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px', borderRadius: 16, border: '1px solid rgba(255,255,255,0.07)', background: 'rgba(255,255,255,0.03)', cursor: 'pointer', textAlign: 'left' }}>
            <div style={{ width: 40, height: 40, borderRadius: 12, background: 'rgba(248,113,113,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
              <Bell size={18} color="#F87171" />
            </div>
            <div style={{ flex: 1 }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: '0 0 2px' }}>Configurar Thresholds</p>
              <p style={{ fontSize: 12, color: '#64748B', margin: 0 }}>Ajustar limites de alerta por tipo</p>
            </div>
            <ChevronRight size={16} color="#475569" />
          </button>
        </div>
      )}

      {/* ── RBAC / Roles ── */}
      {activeSection === 'roles' && (
        <div style={{ display: 'grid', gap: 12 }}>
          <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Roles do sistema</p>
              <button onClick={() => setModal({ title: 'Criar Role', action: 'role_create', successMsg: 'Role criada!', fields: [{ key: 'name', label: 'Nome da role', type: 'text', placeholder: 'Ex: RISK_MANAGER' }, { key: 'description', label: 'Descrição', type: 'text', placeholder: 'O que esta role pode fazer?' }], submitLabel: 'Criar' })} style={{ width: 28, height: 28, borderRadius: 8, background: 'rgba(52,211,153,0.1)', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Plus size={14} color="#34D399" />
              </button>
            </div>
            {store.roles.map((r, i) => (
              <div key={r.id} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '14px 20px', borderBottom: i < store.roles.length - 1 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                <div style={{ width: 36, height: 36, borderRadius: 10, background: 'rgba(129,140,248,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                  <GitBranch size={16} color="#818CF8" />
                </div>
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 13, fontWeight: 700, color: '#F1F5F9', margin: '0 0 2px', fontFamily: 'DM Mono, monospace' }}>{r.name}</p>
                  <p style={{ fontSize: 11, color: '#475569', margin: 0 }}>{r.description}</p>
                </div>
                <span style={{ fontSize: 12, fontFamily: 'DM Mono, monospace', color: '#818CF8', background: 'rgba(129,140,248,0.1)', borderRadius: 8, padding: '3px 10px', fontWeight: 700 }}>{r.usersCount} users</span>
              </div>
            ))}
          </div>

          <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Atribuições de roles</p>
              <button onClick={() => setModal({ title: 'Atribuir Role', action: 'role_assign', successMsg: 'Role atribuída!', fields: [{ key: 'userId', label: 'ID do usuário', type: 'text', placeholder: 'USR-001', default: 'USR-004' }, { key: 'userName', label: 'Nome do usuário', type: 'text', default: 'Novo Operador' }, { key: 'roleName', label: 'Role', type: 'select', options: store.roles.map(r => ({ value: r.name, label: r.name })) }], submitLabel: 'Atribuir' })} style={{ fontSize: 12, color: '#34D399', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 700 }}>
                + Atribuir
              </button>
            </div>
            {store.userRoles.map((ur, i) => (
              <div key={`${ur.userId}-${ur.roleName}`} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '14px 20px', borderBottom: i < store.userRoles.length - 1 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                <div style={{ width: 32, height: 32, borderRadius: 8, background: 'rgba(255,255,255,0.06)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, fontWeight: 700, color: '#34D399', flexShrink: 0 }}>{ur.userName.charAt(0)}</div>
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 13, fontWeight: 600, color: '#F1F5F9', margin: '0 0 2px' }}>{ur.userName}</p>
                  <p style={{ fontSize: 11, color: '#475569', margin: 0 }}>{ur.userId} · atribuído por {ur.assignedBy}</p>
                </div>
                <span style={{ fontSize: 11, fontWeight: 700, color: '#818CF8', background: 'rgba(129,140,248,0.1)', borderRadius: 6, padding: '2px 8px', fontFamily: 'DM Mono, monospace' }}>{ur.roleName}</span>
                <button onClick={() => { dispatch('role_remove', { userId: ur.userId, roleName: ur.roleName }); toast(`Role ${ur.roleName} removida`, 'warning') }} style={{ width: 28, height: 28, borderRadius: 8, background: 'rgba(248,113,113,0.1)', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Trash2 size={13} color="#F87171" />
                </button>
              </div>
            ))}
          </div>

          <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Regras de Separação</p>
              <button onClick={() => setModal({ title: 'Nova Regra de Separação', action: 'role_separation_add', successMsg: 'Regra criada!', fields: [{ key: 'roleA', label: 'Role A', type: 'select', options: store.roles.map(r => ({ value: r.name, label: r.name })) }, { key: 'roleB', label: 'Role B (incompatível)', type: 'select', options: store.roles.map(r => ({ value: r.name, label: r.name })) }, { key: 'reason', label: 'Motivo regulatório', type: 'text', default: 'Segregação de funções' }], submitLabel: 'Criar regra' })} style={{ fontSize: 12, color: '#818CF8', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 700 }}>
                + Nova regra
              </button>
            </div>
            {store.separationRules.map((r, i) => (
              <div key={r.id} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '14px 20px', borderBottom: i < store.separationRules.length - 1 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 13, fontWeight: 700, color: '#F1F5F9', margin: '0 0 2px', fontFamily: 'DM Mono, monospace' }}>
                    {r.roleA} <span style={{ color: '#F87171' }}>⊥</span> {r.roleB}
                  </p>
                  <p style={{ fontSize: 11, color: '#475569', margin: 0 }}>{r.reason}</p>
                </div>
                <button onClick={() => { dispatch('role_separation_remove', { roleA: r.roleA, roleB: r.roleB }); toast('Regra removida', 'warning') }} style={{ width: 28, height: 28, borderRadius: 8, background: 'rgba(248,113,113,0.1)', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Trash2 size={13} color="#F87171" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── Precificação Avançada ── */}
      {activeSection === 'pricing_adv' && (
        <div style={{ display: 'grid', gap: 12 }}>
          <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Versões de Pricing</p>
              <button onClick={() => setModal({ title: 'Nova Versão', action: 'pricing_version_create', successMsg: 'Versão criada!', fields: [{ key: 'version', label: 'Versão (ex: v4.0)', type: 'text', default: 'v4.0' }, { key: 'notes', label: 'Notas de versão', type: 'text', placeholder: 'O que mudou?' }], submitLabel: 'Criar versão' })} style={{ fontSize: 12, color: '#FBBF24', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 700 }}>
                + Nova versão
              </button>
            </div>
            {store.pricingVersions.map((v, i) => (
              <div key={v.id} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '14px 20px', borderBottom: i < store.pricingVersions.length - 1 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                <GitCommit size={16} color={v.status === 'ACTIVE' ? '#34D399' : '#475569'} style={{ flexShrink: 0 }} />
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 13, fontWeight: 700, color: '#F1F5F9', margin: '0 0 2px', fontFamily: 'DM Mono, monospace' }}>{v.version}</p>
                  <p style={{ fontSize: 11, color: '#475569', margin: 0 }}>{v.notes} · {fmtDate(v.createdAt)}</p>
                </div>
                <span style={{ fontSize: 10, fontWeight: 700, color: v.status === 'ACTIVE' ? '#34D399' : '#475569', background: v.status === 'ACTIVE' ? 'rgba(52,211,153,0.1)' : 'rgba(255,255,255,0.05)', borderRadius: 6, padding: '2px 8px' }}>{v.status}</span>
                {v.status !== 'ACTIVE' && (
                  <button onClick={() => { dispatch('pricing_version_activate', { version: v.version }); toast(`Versão ${v.version} ativada!`, 'success') }} style={{ fontSize: 11, color: '#818CF8', background: 'rgba(129,140,248,0.1)', border: 'none', borderRadius: 8, padding: '4px 10px', cursor: 'pointer', fontWeight: 700 }}>
                    Ativar
                  </button>
                )}
              </div>
            ))}
          </div>

          <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Campanhas</p>
              <button onClick={() => setModal({ title: 'Nova Campanha', action: 'pricing_campaign_create', successMsg: 'Campanha criada!', fields: [{ key: 'name', label: 'Nome da campanha', type: 'text', default: 'Promo Q3 2026' }, { key: 'discount', label: 'Desconto (%)', type: 'number', default: 10 }, { key: 'planCode', label: 'Plano', type: 'select', options: [{ value: 'RETAIL', label: 'Retail' }, { value: 'BUSINESS', label: 'Business' }, { value: 'INSTITUTIONAL', label: 'Institutional' }] }, { key: 'startDate', label: 'Início (YYYY-MM-DD)', type: 'text', default: '2026-07-01' }, { key: 'endDate', label: 'Fim (YYYY-MM-DD)', type: 'text', default: '2026-09-30' }], submitLabel: 'Criar campanha' })} style={{ fontSize: 12, color: '#FBBF24', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 700 }}>
                + Nova
              </button>
            </div>
            {store.pricingCampaigns.map((c, i) => (
              <div key={c.id} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '14px 20px', borderBottom: i < store.pricingCampaigns.length - 1 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 13, fontWeight: 700, color: '#F1F5F9', margin: '0 0 2px' }}>{c.name}</p>
                  <p style={{ fontSize: 11, color: '#475569', margin: 0 }}>{c.planCode} · -{c.discount}% · {fmtDate(c.startDate)} → {fmtDate(c.endDate)} · {c.usageCount} usos</p>
                </div>
                <span style={{ fontSize: 10, fontWeight: 700, color: c.status === 'ACTIVE' ? '#34D399' : c.status === 'DRAFT' ? '#F59E0B' : '#475569', background: c.status === 'ACTIVE' ? 'rgba(52,211,153,0.1)' : c.status === 'DRAFT' ? 'rgba(245,158,11,0.1)' : 'rgba(255,255,255,0.05)', borderRadius: 6, padding: '2px 8px' }}>{c.status}</span>
              </div>
            ))}
          </div>

          <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Regras de Pricing</p>
              <button onClick={() => setModal({ title: 'Nova Regra', action: 'pricing_rule_create', successMsg: 'Regra criada!', fields: [{ key: 'planCode', label: 'Plano', type: 'select', options: [{ value: 'RETAIL', label: 'Retail' }, { value: 'BUSINESS', label: 'Business' }, { value: 'INSTITUTIONAL', label: 'Institutional' }] }, { key: 'userType', label: 'Tipo usuário', type: 'select', options: [{ value: 'PF', label: 'PF' }, { value: 'PJ', label: 'PJ' }] }, { key: 'featureCode', label: 'Feature', type: 'select', options: [{ value: 'PIX_SEND', label: 'PIX Envio' }, { value: 'CRYPTO_SWAP', label: 'Crypto Swap' }, { value: 'CARD_JIT', label: 'Cartão JIT' }, { value: 'INVOICE', label: 'Fatura' }] }, { key: 'feeType', label: 'Tipo de taxa', type: 'select', options: [{ value: 'FLAT', label: 'FLAT (valor fixo)' }, { value: 'PERCENT', label: 'PERCENT (%)' }] }, { key: 'feeValue', label: 'Valor da taxa', type: 'number', default: 0, step: 0.1 }], submitLabel: 'Criar regra' })} style={{ fontSize: 12, color: '#FBBF24', background: 'none', border: 'none', cursor: 'pointer', fontWeight: 700 }}>
                + Nova
              </button>
            </div>
            {store.pricingRules.map((r, i) => (
              <div key={r.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '12px 20px', borderBottom: i < store.pricingRules.length - 1 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                <span style={{ fontSize: 10, fontWeight: 700, color: '#818CF8', background: 'rgba(129,140,248,0.1)', borderRadius: 6, padding: '2px 8px', flexShrink: 0 }}>{r.planCode}</span>
                <span style={{ fontSize: 11, color: '#94A3B8', fontFamily: 'DM Mono, monospace', flex: 1 }}>{r.featureCode}</span>
                <span style={{ fontSize: 12, color: '#64748B' }}>{r.userType}</span>
                <span style={{ fontSize: 12, fontWeight: 700, color: r.feeValue === 0 ? '#34D399' : '#FBBF24', fontFamily: 'DM Mono, monospace' }}>
                  {r.feeType === 'PERCENT' ? `${r.feeValue}%` : r.feeValue === 0 ? 'Grátis' : `R$ ${r.feeValue}`}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── Regulatório ── */}
      {activeSection === 'regulatory' && (
        <div style={{ display: 'grid', gap: 12 }}>
          <div style={{ background: 'rgba(96,165,250,0.05)', border: '1px solid rgba(96,165,250,0.2)', borderRadius: 18, padding: '20px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 }}>
              <Globe size={18} color="#60A5FA" />
              <p style={{ fontWeight: 700, fontSize: 15, color: '#F1F5F9', margin: 0 }}>Perfil Regulatório</p>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginBottom: 16 }}>
              {[
                { label: 'Jurisdição', value: store.regulatoryProfile.jurisdiction, color: '#60A5FA' },
                { label: 'Licença',    value: store.regulatoryProfile.licenseType,  color: '#818CF8' },
                { label: 'Risco FATF', value: store.regulatoryProfile.fatfRisk,     color: store.regulatoryProfile.fatfRisk === 'LOW' ? '#34D399' : '#F59E0B' },
                { label: 'VASP',       value: store.regulatoryProfile.vasp ? 'Certificado' : 'Não', color: store.regulatoryProfile.vasp ? '#34D399' : '#475569' },
              ].map(({ label, value, color }) => (
                <div key={label} style={{ background: 'rgba(255,255,255,0.04)', borderRadius: 12, padding: '12px' }}>
                  <p style={{ fontSize: 10, color: '#475569', margin: '0 0 4px', textTransform: 'uppercase', fontWeight: 600 }}>{label}</p>
                  <p style={{ fontSize: 14, fontWeight: 700, color, margin: 0, fontFamily: 'DM Mono, monospace' }}>{value}</p>
                </div>
              ))}
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
              {[
                { label: 'PEP',     value: store.regulatoryProfile.pep,       ok: !store.regulatoryProfile.pep },
                { label: 'Sanções', value: store.regulatoryProfile.sanctions, ok: !store.regulatoryProfile.sanctions },
              ].map(({ label, value, ok }) => (
                <div key={label} style={{ background: 'rgba(255,255,255,0.04)', borderRadius: 12, padding: '12px', display: 'flex', alignItems: 'center', gap: 8 }}>
                  {ok ? <CheckCircle size={16} color="#34D399" /> : <AlertTriangle size={16} color="#F87171" />}
                  <div>
                    <p style={{ fontSize: 10, color: '#475569', margin: '0 0 2px', textTransform: 'uppercase', fontWeight: 600 }}>{label}</p>
                    <p style={{ fontSize: 13, fontWeight: 700, color: ok ? '#34D399' : '#F87171', margin: 0 }}>{value ? 'Sim' : 'Não'}</p>
                  </div>
                </div>
              ))}
            </div>
            <div style={{ marginTop: 14, paddingTop: 14, borderTop: '1px solid rgba(255,255,255,0.07)', display: 'flex', justifyContent: 'space-between' }}>
              <div>
                <p style={{ fontSize: 10, color: '#475569', margin: '0 0 2px', textTransform: 'uppercase' }}>Última revisão</p>
                <p style={{ fontSize: 12, color: '#94A3B8', margin: 0 }}>{fmtDate(store.regulatoryProfile.lastReviewAt)}</p>
              </div>
              <div style={{ textAlign: 'right' }}>
                <p style={{ fontSize: 10, color: '#475569', margin: '0 0 2px', textTransform: 'uppercase' }}>Próxima revisão</p>
                <p style={{ fontSize: 12, color: '#FBBF24', margin: 0, fontWeight: 700 }}>{fmtDate(store.regulatoryProfile.nextReviewAt)}</p>
              </div>
            </div>
          </div>

          <button onClick={() => setModal({
            title: 'Atualizar Perfil Regulatório', action: 'regulatory_profile_update', successMsg: 'Perfil regulatório atualizado!',
            description: 'Mantenha o perfil atualizado para conformidade com FATF, BACEN e VASP.',
            fields: [
              { key: 'fatfRisk',     label: 'Classificação de risco FATF', type: 'select', options: [{ value: 'LOW', label: 'LOW' }, { value: 'MEDIUM', label: 'MEDIUM' }, { value: 'HIGH', label: 'HIGH' }] },
              { key: 'jurisdiction', label: 'Jurisdição principal', type: 'select', options: [{ value: 'BRA', label: 'Brasil (BRA)' }, { value: 'ARE', label: 'Dubai — UAE (ARE)' }, { value: 'GBR', label: 'Reino Unido (GBR)' }, { value: 'USA', label: 'EUA (USA)' }] },
              { key: 'licenseType',  label: 'Tipo de licença', type: 'select', options: [{ value: 'EMI', label: 'EMI (E-Money Institution)' }, { value: 'PI', label: 'PI (Payment Institution)' }, { value: 'VASP', label: 'VASP (Virtual Asset)' }, { value: 'BANK', label: 'Banco completo' }] },
              { key: 'vasp',        label: 'Certificação VASP', type: 'select', options: [{ value: 'true', label: 'Certificado' }, { value: 'false', label: 'Não certificado' }] },
              { key: 'pep',         label: 'Pessoa Politicamente Exposta (PEP)', type: 'select', options: [{ value: 'false', label: 'Não' }, { value: 'true', label: 'Sim' }] },
              { key: 'sanctions',   label: 'Lista de Sanções', type: 'select', options: [{ value: 'false', label: 'Limpo' }, { value: 'true', label: 'Match encontrado' }] },
            ],
            submitLabel: 'Salvar perfil',
          })} style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px', borderRadius: 16, border: '1px solid rgba(255,255,255,0.07)', background: 'rgba(255,255,255,0.03)', cursor: 'pointer', textAlign: 'left' }}>
            <div style={{ width: 40, height: 40, borderRadius: 12, background: 'rgba(96,165,250,0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
              <Globe size={18} color="#60A5FA" />
            </div>
            <div style={{ flex: 1 }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: '0 0 2px' }}>Editar Perfil Regulatório</p>
              <p style={{ fontSize: 12, color: '#64748B', margin: 0 }}>Jurisdição, licença, risco FATF, VASP, PEP, sanções</p>
            </div>
            <ChevronRight size={16} color="#475569" />
          </button>

          {store.conversionAudits?.length > 0 && (
            <div style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden' }}>
              <div style={{ padding: '14px 20px', borderBottom: '1px solid rgba(255,255,255,0.07)' }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: '#F1F5F9', margin: 0 }}>Audit de Conversões</p>
                <p style={{ fontSize: 12, color: '#475569', margin: '2px 0 0' }}>Log regulatório de conversões automáticas</p>
              </div>
              {store.conversionAudits.map((a, i) => (
                <div key={a.id} style={{ padding: '12px 20px', borderBottom: i < store.conversionAudits.length - 1 ? '1px solid rgba(255,255,255,0.07)' : 'none' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
                    <span style={{ fontSize: 11, fontWeight: 700, color: '#34D399', fontFamily: 'DM Mono, monospace' }}>{a.trigger}</span>
                    <span style={{ fontSize: 10, color: '#475569' }}>{fmtDate(a.createdAt)}</span>
                  </div>
                  <p style={{ fontSize: 12, color: '#94A3B8', margin: 0, fontFamily: 'DM Mono, monospace' }}>
                    {a.amount} {a.from} → {fmt(a.converted)} {a.to} @ {a.rate.toLocaleString('pt-BR')}
                  </p>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {modal && <SimModal config={modal} onClose={() => setModal(null)} />}
    </div>
  )
}
