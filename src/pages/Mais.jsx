import React, { useState } from 'react'
import { Shield, Activity, Users, Settings, BarChart3, FileText, ChevronRight, AlertTriangle, CheckCircle, Server, Archive, ToggleLeft, ToggleRight, Crown, UserX, RefreshCw } from 'lucide-react'
import { useApp } from '../App.jsx'
import SimModal from '../components/SimModal.jsx'

function fmt(cents) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100)
}

const SECTIONS = [
  { id: 'kyc', label: 'KYC & Limites', icon: Shield, color: '#34D399' },
  { id: 'compliance', label: 'Compliance', icon: AlertTriangle, color: '#F59E0B' },
  { id: 'pre_reg', label: 'Pré-cadastro', icon: Users, color: '#60A5FA' },
  { id: 'admin', label: 'Admin', icon: Crown, color: '#FBBF24' },
  { id: 'obs', label: 'Observabilidade', icon: Activity, color: '#A78BFA' },
  { id: 'settings', label: 'Configurações', icon: Settings, color: '#94A3B8' },
]

export default function Mais() {
  const { store, dispatch, toast, modal, setModal } = useApp()
  const [activeSection, setActiveSection] = useState('kyc')

  const kyc = store.kyc
  const obs = store.observability
  const compliance = store.compliance
  const preRegs = store.preRegistrations

  return (
    <div>
      <div style={{ marginBottom: 20 }}>
        <h2 style={{ fontSize: 22, fontWeight: 800, fontFamily: 'Syne,sans-serif', color: 'var(--t1)', margin: '0 0 4px' }}>Mais</h2>
        <p style={{ fontSize: 13, color: 'var(--t3)', margin: 0 }}>KYC, compliance, admin e observabilidade</p>
      </div>

      {/* Section tabs */}
      <div style={{ display: 'flex', gap: 8, overflowX: 'auto', paddingBottom: 8, marginBottom: 20 }} className="hide-scroll">
        {SECTIONS.map(({ id, label, icon: Icon, color }) => {
          const active = activeSection === id
          return (
            <button key={id} onClick={() => setActiveSection(id)} style={{
              flexShrink: 0, display: 'flex', alignItems: 'center', gap: 6,
              padding: '8px 14px', borderRadius: 20, cursor: 'pointer',
              background: active ? `${color}20` : 'var(--surface)',
              border: `1px solid ${active ? `${color}50` : 'var(--border)'}`,
              color: active ? color : 'var(--t2)',
              fontSize: 12, fontWeight: 700, transition: 'all 0.15s',
            }}>
              <Icon size={13} />
              {label}
            </button>
          )
        })}
      </div>

      {/* KYC */}
      {activeSection === 'kyc' && (
        <div style={{ display: 'grid', gap: 12 }}>
          {/* KYC status card */}
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 18, padding: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 16 }}>
              <div>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: '0 0 4px', textTransform: 'uppercase', fontWeight: 600 }}>Nível KYC</p>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={{ fontSize: 22, fontWeight: 800, color: 'var(--accent)', fontFamily: 'Syne,sans-serif' }}>{kyc.level}</span>
                  <CheckCircle size={18} color="var(--accent)" />
                </div>
              </div>
              <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--accent)', background: 'rgba(0,229,153,0.1)', borderRadius: 8, padding: '4px 10px' }}>
                VASP Compliant
              </span>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 16 }}>
              {[
                { label: 'Limite Diário', value: fmt(kyc.limits.daily) },
                { label: 'Limite Mensal', value: fmt(kyc.limits.monthly) },
                { label: 'Usado hoje', value: fmt(kyc.limits.used?.daily || 0) },
                { label: 'Disponível hoje', value: fmt(kyc.limits.daily - (kyc.limits.used?.daily || 0)) },
              ].map(({ label, value }) => (
                <div key={label} style={{ background: 'var(--surface-2)', borderRadius: 12, padding: '12px' }}>
                  <p style={{ fontSize: 11, color: 'var(--t3)', margin: '0 0 4px', textTransform: 'uppercase' }}>{label}</p>
                  <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 14, fontWeight: 700, color: 'var(--t1)', margin: 0 }}>{value}</p>
                </div>
              ))}
            </div>
          </div>

          {[
            {
              label: 'Consultar Limites', sub: 'Ver limites detalhados por nível',
              action: 'kyc_limits', successMsg: 'Limites consultados!',
              fields: [{ key: 'level', label: 'Nível KYC', type: 'select', options: [{ value: 'BASIC', label: 'BASIC' }, { value: 'ENHANCED', label: 'ENHANCED' }, { value: 'FULL', label: 'FULL' }] }],
            },
            {
              label: 'Upgrade de KYC', sub: 'Aumentar nível de verificação',
              action: 'kyc_upgrade', successMsg: 'Upgrade solicitado!',
              fields: [
                { key: 'targetLevel', label: 'Nível desejado', type: 'select', options: [{ value: 'ENHANCED', label: 'ENHANCED' }, { value: 'FULL', label: 'FULL' }] },
                { key: 'documents', label: 'Documentos enviados', type: 'textarea', placeholder: 'Lista de documentos...', default: 'RG, CPF, Comprovante de Residência, Selfie' },
              ],
            },
          ].map(({ label, sub, action, successMsg, fields }) => (
            <button key={label} onClick={() => setModal({ title: label, action, successMsg: successMsg || `${label} realizado!`, fields, submitLabel: 'Confirmar' })} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px',
              borderRadius: 16, border: '1px solid var(--border)', background: 'var(--surface)',
              cursor: 'pointer', textAlign: 'left',
            }}>
              <Shield size={20} color="#34D399" style={{ flexShrink: 0 }} />
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{sub}</p>
              </div>
              <ChevronRight size={16} color="var(--t3)" />
            </button>
          ))}
        </div>
      )}

      {/* Compliance */}
      {activeSection === 'compliance' && (
        <div style={{ display: 'grid', gap: 12 }}>
          {/* Cases summary */}
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 18, padding: '20px' }}>
            <p style={{ fontWeight: 700, fontSize: 15, color: 'var(--t1)', margin: '0 0 12px' }}>Casos de Compliance</p>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10 }}>
              {[
                { label: 'Em análise', value: compliance?.casesOpen || 2, color: '#F59E0B' },
                { label: 'Resolvidos', value: compliance?.casesClosed || 18, color: 'var(--accent)' },
                { label: 'Escalados', value: compliance?.casesEscalated || 1, color: '#F87171' },
              ].map(({ label, value, color }) => (
                <div key={label} style={{ background: 'var(--surface-2)', borderRadius: 12, padding: '12px 10px', textAlign: 'center' }}>
                  <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 20, fontWeight: 800, color, margin: '0 0 4px' }}>{value}</p>
                  <p style={{ fontSize: 10, color: 'var(--t3)', margin: 0 }}>{label}</p>
                </div>
              ))}
            </div>
          </div>

          {[
            {
              label: 'Abrir Caso', sub: 'Iniciar investigação de compliance',
              fields: [
                { key: 'userId', label: 'ID do usuário', type: 'text', default: 'USR-001' },
                { key: 'reason', label: 'Motivo', type: 'select', options: [{ value: 'aml', label: 'AML (Anti-Lavagem)' }, { value: 'fraud', label: 'Fraude' }, { value: 'kyc', label: 'KYC Incompleto' }, { value: 'unusual', label: 'Atividade Incomum' }] },
                { key: 'description', label: 'Descrição detalhada', type: 'textarea', placeholder: 'Descreva o caso...', default: 'Transações acima do padrão histórico' },
              ],
              action: 'compliance_case', successMsg: 'Caso aberto!',
            },
            {
              label: 'Registrar Evento', sub: 'Log de evento regulatório',
              fields: [
                { key: 'eventType', label: 'Tipo de evento', type: 'select', options: [{ value: 'suspicious_tx', label: 'Transação Suspeita' }, { value: 'high_value', label: 'Alto Valor' }, { value: 'pep_match', label: 'Match PEP' }, { value: 'sanction_match', label: 'Match Sanção' }] },
                { key: 'details', label: 'Detalhes', type: 'textarea', default: 'Transação de alto risco identificada pelo sistema' },
              ],
              action: 'compliance_event', successMsg: 'Evento registrado!',
            },
          ].map(({ label, sub, action, successMsg, fields }) => (
            <button key={label} onClick={() => setModal({ title: label, action, successMsg, fields, submitLabel: 'Registrar' })} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px',
              borderRadius: 16, border: '1px solid var(--border)', background: 'var(--surface)',
              cursor: 'pointer', textAlign: 'left',
            }}>
              <AlertTriangle size={20} color="#F59E0B" style={{ flexShrink: 0 }} />
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{sub}</p>
              </div>
              <ChevronRight size={16} color="var(--t3)" />
            </button>
          ))}
        </div>
      )}

      {/* Pre-registration */}
      {activeSection === 'pre_reg' && (
        <div style={{ display: 'grid', gap: 12 }}>
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 18, overflow: 'hidden' }}>
            <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between' }}>
              <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: 0 }}>Lista de espera</p>
              <span style={{ fontSize: 12, color: '#60A5FA', fontWeight: 700 }}>{preRegs?.length || 3} cadastros</span>
            </div>
            {(preRegs || [
              { name: 'Ana Costa', email: 'ana@startup.com', plan: 'BUSINESS', status: 'pending' },
              { name: 'Dubai Corp LLC', email: 'ops@dubaicorp.ae', plan: 'INSTITUTIONAL', status: 'verified' },
              { name: 'João Freitas', email: 'joao@gmail.com', plan: 'RETAIL', status: 'pending' },
            ]).map((p, i, arr) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '14px 20px', borderBottom: i < arr.length - 1 ? '1px solid var(--border)' : 'none' }}>
                <div style={{ width: 36, height: 36, borderRadius: 10, background: 'var(--surface-2)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 14, fontWeight: 700, color: 'var(--accent)', flexShrink: 0 }}>
                  {p.name.charAt(0)}
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <p style={{ fontWeight: 600, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{p.name}</p>
                  <p style={{ fontSize: 11, color: 'var(--t3)', margin: 0 }}>{p.email} · {p.plan}</p>
                </div>
                <span style={{
                  fontSize: 10, fontWeight: 700, textTransform: 'uppercase',
                  color: p.status === 'verified' ? 'var(--accent)' : '#F59E0B',
                  background: p.status === 'verified' ? 'rgba(0,229,153,0.1)' : 'rgba(245,158,11,0.1)',
                  borderRadius: 6, padding: '3px 8px',
                }}>
                  {p.status}
                </span>
              </div>
            ))}
          </div>

          {[
            {
              label: 'Novo Pré-cadastro', sub: 'Registrar interesse de cliente',
              action: 'pre_registration', successMsg: 'Pré-cadastro criado!',
              fields: [
                { key: 'name', label: 'Nome completo', type: 'text', placeholder: 'Nome da empresa ou pessoa', default: 'Abu Dhabi Holdings' },
                { key: 'email', label: 'E-mail', type: 'text', placeholder: 'email@empresa.com', default: 'cfo@abudhabi.ae' },
                { key: 'plan', label: 'Plano de interesse', type: 'select', options: [{ value: 'RETAIL', label: 'Retail' }, { value: 'BUSINESS', label: 'Business' }, { value: 'INSTITUTIONAL', label: 'Institutional' }] },
                { key: 'phone', label: 'Telefone (opcional)', type: 'text', placeholder: '+971 XX XXXX XXXX' },
              ],
            },
            {
              label: 'Verificar Pré-cadastro', sub: 'Aprovar cliente da lista de espera',
              action: 'pre_registration_verify', successMsg: 'Pré-cadastro verificado!',
              fields: [
                { key: 'preRegId', label: 'ID do pré-cadastro', type: 'text', default: 'PRE-001' },
                { key: 'status', label: 'Status', type: 'select', options: [{ value: 'approved', label: 'Aprovado' }, { value: 'rejected', label: 'Rejeitado' }, { value: 'waitlist', label: 'Manter na espera' }] },
              ],
            },
          ].map(({ label, sub, action, successMsg, fields }) => (
            <button key={label} onClick={() => setModal({ title: label, action, successMsg: successMsg || `${label}!`, fields, submitLabel: 'Salvar' })} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px',
              borderRadius: 16, border: '1px solid var(--border)', background: 'var(--surface)',
              cursor: 'pointer', textAlign: 'left',
            }}>
              <Users size={20} color="#60A5FA" style={{ flexShrink: 0 }} />
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{sub}</p>
              </div>
              <ChevronRight size={16} color="var(--t3)" />
            </button>
          ))}
        </div>
      )}

      {/* Admin */}
      {activeSection === 'admin' && (
        <div style={{ display: 'grid', gap: 12 }}>
          <div style={{
            background: 'linear-gradient(135deg, rgba(251,191,36,0.1), rgba(251,191,36,0.05))',
            border: '1px solid rgba(251,191,36,0.3)',
            borderRadius: 16, padding: '16px 20px',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
              <Crown size={16} color="var(--gold)" />
              <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--gold)', margin: 0 }}>Painel Administrativo</p>
            </div>
            <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>
              Acesso privilegiado: RBAC role ADMIN. Todas ações são auditadas.
            </p>
          </div>

          {[
            {
              label: 'Alterar Plano do Usuário', icon: Crown, color: 'var(--gold)', sub: 'Upgrade/downgrade de plano',
              action: 'admin_plan_change', successMsg: 'Plano alterado!',
              fields: [
                { key: 'userId', label: 'ID do usuário', type: 'text', default: 'USR-001' },
                { key: 'plan', label: 'Novo plano', type: 'select', options: [{ value: 'RETAIL', label: 'Retail' }, { value: 'BUSINESS', label: 'Business' }, { value: 'INSTITUTIONAL', label: 'Institutional' }] },
              ],
            },
            {
              label: 'Toggle de Feature', icon: ToggleRight, color: '#818CF8', sub: 'Ativar/desativar funcionalidade',
              action: 'admin_feature_toggle', successMsg: 'Feature atualizada!',
              fields: [
                { key: 'feature', label: 'Feature', type: 'select', options: [{ value: 'crypto_swap', label: 'Crypto Swap' }, { value: 'pix_international', label: 'PIX Internacional' }, { value: 'jit_card', label: 'Cartão JIT' }, { value: 'auto_invest', label: 'Auto Invest' }] },
                { key: 'enabled', label: 'Estado', type: 'select', options: [{ value: 'true', label: 'Ativado' }, { value: 'false', label: 'Desativado' }] },
              ],
            },
            {
              label: 'Ajustar Limite', icon: BarChart3, color: '#34D399', sub: 'Modificar limites por usuário',
              action: 'admin_limit_adjust', successMsg: 'Limite ajustado!',
              fields: [
                { key: 'userId', label: 'ID do usuário', type: 'text', default: 'USR-001' },
                { key: 'limitType', label: 'Tipo de limite', type: 'select', options: [{ value: 'daily_pix', label: 'PIX Diário' }, { value: 'monthly_pix', label: 'PIX Mensal' }, { value: 'card_limit', label: 'Limite Cartão' }] },
                { key: 'newLimit', label: 'Novo limite (centavos)', type: 'number', default: 50000000 },
              ],
            },
            {
              label: 'Bloquear Usuário', icon: UserX, color: '#F87171', sub: 'Suspender conta de usuário',
              action: 'admin_user_block', successMsg: 'Usuário bloqueado!',
              fields: [
                { key: 'userId', label: 'ID do usuário', type: 'text', default: 'USR-999' },
                { key: 'reason', label: 'Motivo', type: 'select', options: [{ value: 'fraud', label: 'Fraude confirmada' }, { value: 'kyc_fail', label: 'Falha no KYC' }, { value: 'compliance', label: 'Ordem de compliance' }, { value: 'request', label: 'Solicitação do cliente' }] },
              ],
            },
          ].map(({ label, icon: Icon, color, sub, action, successMsg, fields }) => (
            <button key={label} onClick={() => setModal({ title: label, action, successMsg, fields, submitLabel: 'Executar' })} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px',
              borderRadius: 16, border: '1px solid var(--border)', background: 'var(--surface)',
              cursor: 'pointer', textAlign: 'left',
            }}>
              <div style={{ width: 40, height: 40, borderRadius: 12, background: `${color}15`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Icon size={18} color={color} />
              </div>
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{sub}</p>
              </div>
              <ChevronRight size={16} color="var(--t3)" />
            </button>
          ))}
        </div>
      )}

      {/* Observability */}
      {activeSection === 'obs' && (
        <div style={{ display: 'grid', gap: 12 }}>
          {/* Status cards */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 10 }}>
            {[
              { label: 'Uptime', value: obs.uptime, sub: '30 dias', color: 'var(--accent)', good: true },
              { label: 'Latência P99', value: `${obs.latencyP99}ms`, sub: 'Tempo resposta', color: obs.latencyP99 < 100 ? 'var(--accent)' : '#F59E0B', good: obs.latencyP99 < 100 },
              { label: 'Webhooks', value: `${obs.pendingWebhooks}`, sub: 'Pendentes', color: obs.pendingWebhooks > 5 ? '#F87171' : '#60A5FA', good: obs.pendingWebhooks < 5 },
              { label: 'Status', value: 'Operacional', sub: 'Sistema', color: 'var(--accent)', good: true },
            ].map(({ label, value, sub, color, good }) => (
              <div key={label} style={{ background: 'var(--surface)', border: `1px solid ${good ? 'var(--border)' : `${color}40`}`, borderRadius: 16, padding: '16px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ fontSize: 11, color: 'var(--t3)', textTransform: 'uppercase', fontWeight: 600 }}>{label}</span>
                  <div style={{ width: 8, height: 8, borderRadius: '50%', background: color, marginTop: 2 }} />
                </div>
                <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 18, fontWeight: 700, color, margin: '0 0 2px' }}>{value}</p>
                <p style={{ fontSize: 11, color: 'var(--t3)', margin: 0 }}>{sub}</p>
              </div>
            ))}
          </div>

          {[
            {
              label: 'Dashboard de Saúde', icon: Activity, color: '#A78BFA', sub: 'Ver métricas em tempo real',
              action: 'obs_summary', successMsg: 'Sumário obtido!',
              fields: [{ key: 'period', label: 'Período', type: 'select', options: [{ value: '1h', label: '1 hora' }, { value: '24h', label: '24 horas' }, { value: '7d', label: '7 dias' }] }],
            },
            {
              label: 'Reprocessar Webhook', icon: Server, color: '#60A5FA', sub: 'Reenviar webhook com falha',
              action: 'webhook_retry', successMsg: 'Webhook reenviado!',
              fields: [
                { key: 'webhookId', label: 'ID do webhook', type: 'text', default: 'WH-' + Date.now().toString().slice(-6) },
                { key: 'maxRetries', label: 'Máximo de tentativas', type: 'number', default: 3 },
              ],
            },
            {
              label: 'Arquivar Audit Log', icon: Archive, color: '#94A3B8', sub: 'Arquivar logs de auditoria',
              action: 'audit_archive', successMsg: 'Logs arquivados!',
              fields: [
                { key: 'before', label: 'Arquivar antes de', type: 'text', default: '2026-01-01', placeholder: 'YYYY-MM-DD' },
                { key: 'compress', label: 'Comprimir', type: 'select', options: [{ value: 'true', label: 'Sim (gzip)' }, { value: 'false', label: 'Não' }] },
              ],
            },
          ].map(({ label, icon: Icon, color, sub, action, successMsg, fields }) => (
            <button key={label} onClick={() => setModal({ title: label, action, successMsg, fields, submitLabel: 'Executar' })} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px',
              borderRadius: 16, border: '1px solid var(--border)', background: 'var(--surface)',
              cursor: 'pointer', textAlign: 'left',
            }}>
              <div style={{ width: 40, height: 40, borderRadius: 12, background: `${color}15`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Icon size={18} color={color} />
              </div>
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{sub}</p>
              </div>
              <ChevronRight size={16} color="var(--t3)" />
            </button>
          ))}

          {/* Audit log preview */}
          {store.auditLogs?.length > 0 && (
            <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 18, overflow: 'hidden' }}>
              <div style={{ padding: '14px 20px', borderBottom: '1px solid var(--border)' }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: 0 }}>Audit Trail recente</p>
              </div>
              {store.auditLogs.slice(0, 5).map((log, i) => (
                <div key={i} style={{ padding: '12px 20px', borderBottom: i < 4 ? '1px solid var(--border)' : 'none' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8 }}>
                    <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 12, color: 'var(--accent)', margin: 0 }}>{log.action}</p>
                    <span style={{ fontSize: 10, color: 'var(--t3)', flexShrink: 0 }}>{log.ts || log.timestamp}</span>
                  </div>
                  {log.details && <p style={{ fontSize: 11, color: 'var(--t3)', margin: '2px 0 0' }}>{JSON.stringify(log.details).slice(0, 80)}...</p>}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Settings */}
      {activeSection === 'settings' && (
        <div style={{ display: 'grid', gap: 12 }}>
          {/* User info */}
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 18, padding: '20px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 14, marginBottom: 16 }}>
              <div style={{ width: 56, height: 56, borderRadius: 18, background: 'linear-gradient(135deg, #1E3A5F, #0C1E38)', border: '2px solid var(--border)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 22, fontWeight: 700, color: 'var(--accent)' }}>
                {store.user.name.charAt(0)}
              </div>
              <div>
                <p style={{ fontWeight: 700, fontSize: 16, color: 'var(--t1)', margin: '0 0 4px' }}>{store.user.name}</p>
                <p style={{ fontSize: 13, color: 'var(--t3)', margin: '0 0 4px' }}>{store.user.email}</p>
                <div style={{ display: 'flex', gap: 6 }}>
                  <span style={{ fontSize: 10, fontWeight: 700, color: 'var(--accent)', background: 'rgba(0,229,153,0.1)', borderRadius: 6, padding: '2px 8px' }}>{store.user.plan}</span>
                  <span style={{ fontSize: 10, fontWeight: 700, color: '#818CF8', background: 'rgba(129,140,248,0.1)', borderRadius: 6, padding: '2px 8px' }}>{store.user.type}</span>
                </div>
              </div>
            </div>
          </div>

          {[
            {
              label: 'Modo UX', sub: `Atual: ${store.user.uxMode}`, icon: Settings, color: '#94A3B8',
              action: 'settings_ux_mode', successMsg: 'Modo UX alterado!',
              fields: [{ key: 'mode', label: 'Modo', type: 'select', options: [{ value: 'PRO', label: 'PRO (Avançado)' }, { value: 'SIMPLE', label: 'SIMPLE (Simplificado)' }] }],
            },
            {
              label: 'Auto-conversão Crypto', sub: 'Converter cripto automaticamente', icon: RefreshCw, color: '#34D399',
              action: 'settings_auto_convert', successMsg: 'Configuração salva!',
              fields: [
                { key: 'enabled', label: 'Ativado', type: 'select', options: [{ value: 'true', label: 'Sim' }, { value: 'false', label: 'Não' }] },
                { key: 'targetCurrency', label: 'Moeda destino', type: 'select', options: [{ value: 'BRL', label: 'BRL' }, { value: 'USDT', label: 'USDT' }, { value: 'USD', label: 'USD' }] },
                { key: 'threshold', label: 'Gatilho (centavos)', type: 'number', default: 100000 },
              ],
            },
          ].map(({ label, sub, icon: Icon, color, action, successMsg, fields }) => (
            <button key={label} onClick={() => setModal({ title: label, action, successMsg, fields, submitLabel: 'Salvar' })} style={{
              display: 'flex', alignItems: 'center', gap: 14, padding: '18px 20px',
              borderRadius: 16, border: '1px solid var(--border)', background: 'var(--surface)',
              cursor: 'pointer', textAlign: 'left',
            }}>
              <div style={{ width: 40, height: 40, borderRadius: 12, background: `${color}15`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Icon size={18} color={color} />
              </div>
              <div style={{ flex: 1 }}>
                <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: '0 0 2px' }}>{label}</p>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: 0 }}>{sub}</p>
              </div>
              <ChevronRight size={16} color="var(--t3)" />
            </button>
          ))}
        </div>
      )}

      {modal && <SimModal config={modal} onClose={() => setModal(null)} />}
    </div>
  )
}

