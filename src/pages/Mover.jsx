import React, { useState } from 'react'
import { Send, QrCode, FileText, ArrowLeftRight, Globe, Calendar, CheckCircle, XCircle, Plus } from 'lucide-react'
import { useApp } from '../App.jsx'
import SimModal from '../components/SimModal.jsx'

function fmt(cents) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100)
}

const CATEGORIES = [
  {
    id: 'pix', label: 'PIX', icon: Send, color: '#00E599',
    actions: [
      {
        label: 'Enviar PIX', sub: 'Transferência instantânea',
        modal: {
          title: 'Enviar PIX', action: 'pix_send', successMsg: 'PIX enviado com sucesso!',
          description: 'Pagamentos instantâneos 24h/7d. Sem tarifas para PF.',
          fields: [
            { key: 'key', label: 'Chave PIX', type: 'text', placeholder: 'CPF, e-mail, celular ou chave aleatória', default: 'contato@empresa.com' },
            { key: 'amount', label: 'Valor', type: 'number', placeholder: '10000', default: 10000, hint: 'Em centavos. R$ 100,00 = 10000' },
            { key: 'description', label: 'Descrição (opcional)', type: 'text', placeholder: 'Ex: Aluguel, Freelance...', default: '' },
          ],
          submitLabel: 'Enviar PIX',
        }
      },
      {
        label: 'Registrar Chave PIX', sub: 'Adicionar nova chave',
        modal: {
          title: 'Registrar Chave PIX', action: 'pix_key_register', successMsg: 'Chave PIX registrada!',
          fields: [
            { key: 'type', label: 'Tipo de chave', type: 'select', options: [{ value: 'cpf', label: 'CPF' }, { value: 'email', label: 'E-mail' }, { value: 'phone', label: 'Telefone' }, { value: 'random', label: 'Chave Aleatória' }] },
            { key: 'value', label: 'Valor da chave', type: 'text', placeholder: 'Informe o valor' },
          ],
          submitLabel: 'Registrar chave',
        }
      },
      {
        label: 'PIX via Crypto', sub: 'Pagar com USDT → BRL',
        modal: {
          title: 'PIX via Crypto', action: 'pix_send_crypto', successMsg: 'PIX enviado via crypto!',
          description: 'Converta USDT automaticamente e envie PIX em BRL.',
          fields: [
            { key: 'pixKey', label: 'Chave PIX destino', type: 'text', placeholder: 'CPF ou e-mail', default: 'destino@email.com' },
            { key: 'amountBRL', label: 'Valor em BRL (centavos)', type: 'number', default: 57800, hint: 'Será convertido de USDT automaticamente' },
          ],
          submitLabel: 'Enviar',
        }
      },
    ]
  },
  {
    id: 'pagamentos', label: 'Pagamentos', icon: FileText, color: '#818CF8',
    actions: [
      {
        label: 'Validar Boleto', sub: 'Verificar antes de pagar',
        modal: {
          title: 'Validar Boleto', action: 'payment_validate', successMsg: 'Boleto validado!',
          fields: [
            { key: 'barcode', label: 'Código de barras', type: 'text', placeholder: '00000.00000 00000.000000 00000.000000 0 00000000000000', default: '34191.79001 01043.510047 91020.150008 8 84210026000' },
          ],
          submitLabel: 'Validar',
        }
      },
      {
        label: 'Pagar Conta', sub: 'Boleto, concessionárias',
        modal: {
          title: 'Pagar Conta', action: 'payment_confirm', successMsg: 'Pagamento realizado!',
          fields: [
            { key: 'barcode', label: 'Código de barras', type: 'text', default: '34191.79001 01043.510047 91020.150008 8 84210026000' },
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 19900 },
            { key: 'scheduledDate', label: 'Data (YYYY-MM-DD)', type: 'text', default: new Date().toISOString().slice(0, 10) },
          ],
          submitLabel: 'Pagar',
        }
      },
      {
        label: 'Agendar Pagamento', sub: 'Programar para data futura',
        modal: {
          title: 'Agendar Pagamento', action: 'payment_schedule', successMsg: 'Pagamento agendado!',
          fields: [
            { key: 'barcode', label: 'Código de barras', type: 'text', default: '34191.79001 01043.510047 91020.150008 8 84210026000' },
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 29900 },
            { key: 'scheduledDate', label: 'Data agendada', type: 'text', placeholder: 'YYYY-MM-DD', default: '2026-06-01' },
          ],
          submitLabel: 'Agendar',
        }
      },
    ]
  },
  {
    id: 'cobrancas', label: 'Cobranças', icon: QrCode, color: '#34D399',
    actions: [
      {
        label: 'Criar Fatura', sub: 'PIX ou USDT',
        modal: {
          title: 'Nova Fatura', action: 'invoice_create', successMsg: 'Fatura criada!',
          description: 'Cobranças híbridas: cliente paga por PIX ou crypto.',
          fields: [
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 150000 },
            { key: 'currency', label: 'Moeda', type: 'select', options: [{ value: 'BRL', label: 'BRL (PIX)' }, { value: 'USDT', label: 'USDT' }] },
            { key: 'description', label: 'Descrição', type: 'text', default: 'Serviço prestado' },
            { key: 'clientEmail', label: 'E-mail do cliente', type: 'text', placeholder: 'cliente@empresa.com' },
          ],
          submitLabel: 'Gerar fatura',
        }
      },
      {
        label: 'Pagar Fatura', sub: 'Simular recebimento',
        modal: {
          title: 'Pagar Fatura', action: 'invoice_pay', successMsg: 'Fatura paga!',
          fields: [
            { key: 'invoiceId', label: 'ID da fatura', type: 'text', default: 'INV-001' },
            { key: 'method', label: 'Método de pagamento', type: 'select', options: [{ value: 'pix', label: 'PIX' }, { value: 'usdt', label: 'USDT' }] },
          ],
          submitLabel: 'Confirmar pagamento',
        }
      },
    ]
  },
  {
    id: 'transferencias', label: 'Transferências', icon: ArrowLeftRight, color: '#F59E0B',
    actions: [
      {
        label: 'Entre Contas', sub: 'BRL ↔ USD ↔ Investimentos',
        modal: {
          title: 'Transferência Interna', action: 'transfer_internal', successMsg: 'Transferência realizada!',
          description: 'Mova saldo entre suas contas instantaneamente.',
          fields: [
            { key: 'fromAccount', label: 'Conta de origem', type: 'select', options: [{ value: 'main', label: 'BRL Principal' }, { value: 'usd', label: 'USD Internacional' }, { value: 'invest', label: 'Investimentos' }] },
            { key: 'toAccount', label: 'Conta de destino', type: 'select', options: [{ value: 'usd', label: 'USD Internacional' }, { value: 'main', label: 'BRL Principal' }, { value: 'invest', label: 'Investimentos' }] },
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 50000 },
          ],
          submitLabel: 'Transferir',
        }
      },
      {
        label: 'Saque / Retirada', sub: 'TED para banco externo',
        modal: {
          title: 'Saque', action: 'withdrawal', successMsg: 'Saque solicitado!',
          fields: [
            { key: 'bankCode', label: 'Banco (ISPB)', type: 'text', default: '60746948', hint: 'Itaú: 60746948, Bradesco: 60746948' },
            { key: 'branch', label: 'Agência', type: 'text', default: '0001' },
            { key: 'account', label: 'Conta', type: 'text', default: '12345-6' },
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 100000 },
          ],
          submitLabel: 'Sacar',
        }
      },
    ]
  },
  {
    id: 'cambio', label: 'Câmbio', icon: Globe, color: '#60A5FA',
    actions: [
      {
        label: 'Cotação FX', sub: 'Obter taxa de câmbio',
        modal: {
          title: 'Cotação de Câmbio', action: 'pricing_quote', successMsg: 'Cotação obtida!',
          description: 'Taxas competitivas sem spread oculto.',
          fields: [
            { key: 'from', label: 'Moeda de origem', type: 'select', options: [{ value: 'BRL', label: 'BRL' }, { value: 'USD', label: 'USD' }, { value: 'USDT', label: 'USDT' }] },
            { key: 'to', label: 'Moeda de destino', type: 'select', options: [{ value: 'USD', label: 'USD' }, { value: 'BRL', label: 'BRL' }, { value: 'BTC', label: 'BTC' }] },
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 100000 },
          ],
          submitLabel: 'Cotar',
        }
      },
    ]
  },
]

export default function Mover() {
  const { modal, setModal } = useApp()
  const [activeCategory, setActiveCategory] = useState('pix')

  const cat = CATEGORIES.find(c => c.id === activeCategory)

  return (
    <div>
      <div style={{ marginBottom: 20 }}>
        <h2 style={{ fontSize: 22, fontWeight: 800, fontFamily: 'Syne,sans-serif', color: 'var(--t1)', margin: '0 0 4px' }}>Mover Dinheiro</h2>
        <p style={{ fontSize: 13, color: 'var(--t3)', margin: 0 }}>PIX, pagamentos, cobranças e transferências</p>
      </div>

      {/* Category tabs */}
      <div style={{ display: 'flex', gap: 8, overflowX: 'auto', paddingBottom: 4, marginBottom: 20 }} className="hide-scroll">
        {CATEGORIES.map(({ id, label, icon: Icon, color }) => {
          const active = activeCategory === id
          return (
            <button key={id} onClick={() => setActiveCategory(id)} style={{
              flexShrink: 0, display: 'flex', alignItems: 'center', gap: 6,
              padding: '8px 16px', borderRadius: 20, cursor: 'pointer',
              background: active ? `${color}20` : 'var(--surface)',
              border: `1px solid ${active ? `${color}50` : 'var(--border)'}`,
              color: active ? color : 'var(--t2)',
              fontSize: 13, fontWeight: 700,
              transition: 'all 0.15s',
            }}>
              <Icon size={14} />
              {label}
            </button>
          )
        })}
      </div>

      {/* Actions grid */}
      <div style={{ display: 'grid', gap: 12 }}>
        {cat?.actions.map(({ label, sub, modal: m }) => (
          <button key={label} onClick={() => setModal(m)} style={{
            display: 'flex', alignItems: 'center', gap: 16,
            padding: '20px', borderRadius: 18, border: '1px solid var(--border)',
            background: 'var(--surface)', cursor: 'pointer', textAlign: 'left',
            transition: 'all 0.15s',
          }}
          onMouseEnter={e => { e.currentTarget.style.borderColor = `${cat.color}50`; e.currentTarget.style.background = `${cat.color}08` }}
          onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--border)'; e.currentTarget.style.background = 'var(--surface)' }}
          >
            <div style={{
              width: 44, height: 44, borderRadius: 14, flexShrink: 0,
              background: `${cat.color}15`,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
            }}>
              {React.createElement(cat.icon, { size: 20, color: cat.color })}
            </div>
            <div style={{ flex: 1 }}>
              <p style={{ fontWeight: 700, fontSize: 15, color: 'var(--t1)', margin: '0 0 2px' }}>{label}</p>
              <p style={{ fontSize: 13, color: 'var(--t3)', margin: 0 }}>{sub}</p>
            </div>
            <div style={{
              width: 32, height: 32, borderRadius: 10,
              background: 'var(--surface-2)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: 'var(--t3)', flexShrink: 0,
            }}>
              →
            </div>
          </button>
        ))}
      </div>

      {/* PIX keys summary */}
      {activeCategory === 'pix' && (
        <div style={{
          marginTop: 20, background: 'var(--surface)',
          border: '1px solid var(--border)', borderRadius: 18, overflow: 'hidden',
        }}>
          <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--border)' }}>
            <p style={{ fontWeight: 700, fontSize: 14, color: 'var(--t1)', margin: 0 }}>Minhas Chaves PIX</p>
          </div>
          {[
            { type: 'CPF', value: '•••.456.789-••', status: 'active' },
            { type: 'E-mail', value: 'rafael@empresa.com.br', status: 'active' },
            { type: 'Telefone', value: '+55 (11) •••••-4321', status: 'active' },
            { type: 'Aleatória', value: 'a1b2c3d4-e5f6-...', status: 'active' },
          ].map((k, i) => (
            <div key={i} style={{
              display: 'flex', alignItems: 'center', padding: '14px 20px',
              borderBottom: i < 3 ? '1px solid var(--border)' : 'none',
            }}>
              <div style={{ flex: 1 }}>
                <p style={{ fontSize: 12, color: 'var(--t3)', margin: '0 0 2px', textTransform: 'uppercase', fontWeight: 600 }}>{k.type}</p>
                <p style={{ fontSize: 14, color: 'var(--t1)', margin: 0, fontFamily: 'DM Mono, monospace' }}>{k.value}</p>
              </div>
              <span style={{ fontSize: 11, color: 'var(--accent)', background: 'rgba(0,229,153,0.1)', borderRadius: 6, padding: '3px 8px', fontWeight: 700 }}>
                Ativa
              </span>
            </div>
          ))}
        </div>
      )}

      {modal && <SimModal config={modal} onClose={() => setModal(null)} />}
    </div>
  )
}
