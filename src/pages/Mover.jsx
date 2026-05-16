import React, { useState } from 'react'
import { Send, QrCode, FileText, ArrowLeftRight, Globe, Calendar, CheckCircle, XCircle, Plus, Plane } from 'lucide-react'
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
        label: 'Entre Contas', sub: 'BRL ↔ USD ↔ AED ↔ Invest',
        modal: {
          title: 'Transferência Interna', action: 'transfer_internal', successMsg: 'Transferência realizada!',
          description: 'Câmbio automático entre contas multi-moeda. Taxa em tempo real.',
          fields: [
            { key: 'fromAccount', label: 'Conta de origem', type: 'select', options: [{ value: 'main', label: '🇧🇷 BRL Principal' }, { value: 'usd', label: '🇺🇸 USD Internacional' }, { value: 'aed', label: '🇦🇪 AED Dubai' }, { value: 'invest', label: '📈 Investimentos' }] },
            { key: 'toAccount', label: 'Conta de destino', type: 'select', options: [{ value: 'usd', label: '🇺🇸 USD Internacional' }, { value: 'main', label: '🇧🇷 BRL Principal' }, { value: 'aed', label: '🇦🇪 AED Dubai' }, { value: 'invest', label: '📈 Investimentos' }] },
            { key: 'amount', label: 'Valor (centavos da moeda de origem)', type: 'number', default: 50000, hint: 'Conversão aplicada automaticamente' },
          ],
          submitLabel: 'Converter e Transferir',
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
        label: 'Cotação FX', sub: 'USD, AED, EUR, GBP',
        modal: {
          title: 'Cotação de Câmbio', action: 'pricing_quote', successMsg: 'Cotação obtida!',
          description: 'Taxas competitivas sem spread oculto. Acesso a múltiplas moedas globais.',
          fields: [
            { key: 'from', label: 'Moeda de origem', type: 'select', options: [{ value: 'BRL', label: '🇧🇷 BRL' }, { value: 'USD', label: '🇺🇸 USD' }, { value: 'AED', label: '🇦🇪 AED' }, { value: 'USDT', label: '₮ USDT' }] },
            { key: 'to', label: 'Moeda de destino', type: 'select', options: [{ value: 'USD', label: '🇺🇸 USD' }, { value: 'AED', label: '🇦🇪 AED' }, { value: 'BRL', label: '🇧🇷 BRL' }, { value: 'BTC', label: '₿ BTC' }] },
            { key: 'amount', label: 'Valor (centavos)', type: 'number', default: 100000 },
          ],
          submitLabel: 'Cotar',
        }
      },
      {
        label: 'PIX → AED', sub: 'Enviar para Dubai instantâneo',
        modal: {
          title: 'PIX Internacional → AED', action: 'transfer_internal', successMsg: 'Conversão realizada!',
          description: 'Converta BRL da sua conta PIX para Dirhams dos Emirados. Taxa: 1 AED ≈ R$ 1,58.',
          fields: [
            { key: 'fromAccount', label: 'Origem', type: 'select', options: [{ value: 'main', label: '🇧🇷 Conta BRL' }] },
            { key: 'toAccount', label: 'Destino', type: 'select', options: [{ value: 'aed', label: '🇦🇪 Conta AED Dubai' }] },
            { key: 'amount', label: 'Valor em BRL (centavos)', type: 'number', default: 158000, hint: 'R$ 1.580,00 → ≈ AED 1.000,00' },
          ],
          submitLabel: 'Converter para AED',
        }
      },
    ]
  },
  {
    id: 'pixintl', label: 'PIX Intl', icon: Plane, color: '#F59E0B',
    actions: [
      {
        label: 'PIX → USD (EUA)', sub: 'Conta americana em dólares',
        modal: {
          title: 'PIX Internacional → USD', action: 'pix_international', successMsg: 'PIX Internacional enviado!',
          description: '🇧🇷→🇺🇸  PIX em BRL convertido e entregue em USD. Taxa 1.5% · Liquidação em até 1h.',
          fields: [
            { key: 'amountBRL', label: 'Valor em BRL (centavos)', type: 'number', default: 578000, hint: 'R$ 5.780,00 → aprox. USD 1.000,00 (descontando taxa 1.5%)' },
            { key: 'destination', label: 'Moeda destino', type: 'select', options: [{ value: 'USD', label: '🇺🇸 USD — Dólar Americano' }] },
            { key: 'recipient', label: 'Beneficiário', type: 'text', default: 'John Smith — Chase Bank' },
            { key: 'swiftOrAccount', label: 'SWIFT / Account', type: 'text', default: 'CHASUS33 — 4532xxxx' },
          ],
          submitLabel: 'Enviar PIX Internacional',
        }
      },
      {
        label: 'PIX → AED (Dubai)', sub: 'Conta Dubai em Dirhams',
        modal: {
          title: 'PIX Internacional → AED', action: 'pix_international', successMsg: 'PIX Internacional enviado!',
          description: '🇧🇷→🇦🇪  PIX em BRL convertido e entregue em AED. Taxa 1.5% · Liquidação em até 30min.',
          fields: [
            { key: 'amountBRL', label: 'Valor em BRL (centavos)', type: 'number', default: 158000, hint: 'R$ 1.580,00 → aprox. AED 1.000,00 (descontando taxa 1.5%)' },
            { key: 'destination', label: 'Moeda destino', type: 'select', options: [{ value: 'AED', label: '🇦🇪 AED — Dirham dos Emirados' }] },
            { key: 'recipient', label: 'Beneficiário', type: 'text', default: 'Dubai Corp LLC' },
            { key: 'iban', label: 'IBAN Dubai', type: 'text', default: 'AE070331234567890123456' },
          ],
          submitLabel: 'Enviar para Dubai',
        }
      },
      {
        label: 'PIX → EUR (Europa)', sub: 'Conta europeia em Euros',
        modal: {
          title: 'PIX Internacional → EUR', action: 'pix_international', successMsg: 'PIX Internacional enviado!',
          description: '🇧🇷→🇪🇺  PIX em BRL convertido e entregue em EUR via SEPA. Taxa 1.5% · Liquidação D+1.',
          fields: [
            { key: 'amountBRL', label: 'Valor em BRL (centavos)', type: 'number', default: 624000, hint: 'R$ 6.240,00 → aprox. EUR 1.000,00 (descontando taxa 1.5%)' },
            { key: 'destination', label: 'Moeda destino', type: 'select', options: [{ value: 'EUR', label: '🇪🇺 EUR — Euro' }] },
            { key: 'recipient', label: 'Beneficiário', type: 'text', default: 'Maria García' },
            { key: 'iban', label: 'IBAN Europa', type: 'text', default: 'ES91 2100 0418 4502 0005 1332' },
          ],
          submitLabel: 'Enviar para Europa',
        }
      },
    ]
  },
]

export default function Mover() {
  const { store, modal, setModal } = useApp()
  const [activeCategory, setActiveCategory] = useState('pix')

  const cat = CATEGORIES.find(c => c.id === activeCategory)

  return (
    <div>
      <div style={{ marginBottom: 20 }}>
        <h2 style={{ fontSize: 22, fontWeight: 800, fontFamily: 'Syne,sans-serif', color: 'var(--t1)', margin: '0 0 4px' }}>Mover Dinheiro</h2>
        <p style={{ fontSize: 13, color: 'var(--t3)', margin: 0 }}>PIX, pagamentos, cobranças e transferências</p>
      </div>

      {/* FX Rate strip */}
      <div style={{
        display: 'flex', gap: 8, overflowX: 'auto', paddingBottom: 4, marginBottom: 16,
      }} className="hide-scroll">
        {[
          { cur: 'USD', flag: '🇺🇸', rate: store.market.USD.rate, change: store.market.USD.change },
          { cur: 'AED', flag: '🇦🇪', rate: store.market.AED.rate, change: store.market.AED.change },
          { cur: 'EUR', flag: '🇪🇺', rate: store.market.EUR.rate, change: store.market.EUR.change },
          { cur: 'GBP', flag: '🇬🇧', rate: store.market.GBP.rate, change: store.market.GBP.change },
        ].map(({ cur, flag, rate, change }) => (
          <div key={cur} style={{
            flexShrink: 0, background: 'var(--surface)', border: '1px solid var(--border)',
            borderRadius: 12, padding: '8px 14px',
            display: 'flex', alignItems: 'center', gap: 8,
          }}>
            <span style={{ fontSize: 16 }}>{flag}</span>
            <div>
              <p style={{ fontSize: 10, color: 'var(--t3)', margin: '0 0 1px', fontWeight: 600 }}>{cur}/BRL</p>
              <p style={{ fontFamily: 'DM Mono, monospace', fontSize: 13, fontWeight: 700, color: 'var(--t1)', margin: 0 }}>
                R$ {(rate / 100).toFixed(2)}
                <span style={{ fontSize: 10, marginLeft: 4, color: change >= 0 ? 'var(--accent)' : '#F87171' }}>
                  {change >= 0 ? '+' : ''}{change.toFixed(1)}%
                </span>
              </p>
            </div>
          </div>
        ))}
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

      {/* PIX Internacional info */}
      {activeCategory === 'pixintl' && (
        <div style={{ marginTop: 16 }}>
          {/* Breakdown explainer */}
          <div style={{
            background: 'linear-gradient(135deg, rgba(245,158,11,0.08), rgba(245,158,11,0.03))',
            border: '1px solid rgba(245,158,11,0.2)',
            borderRadius: 16, padding: '16px',
          }}>
            <p style={{ fontWeight: 700, fontSize: 13, color: '#F59E0B', margin: '0 0 12px' }}>
              Como funciona o PIX Internacional
            </p>
            {[
              { icon: '⚡', text: 'Você envia PIX em BRL da sua conta Z-Finance' },
              { icon: '🔄', text: 'Conversão automática ao câmbio do momento (spread 0%)' },
              { icon: '💸', text: 'Taxa de 1.5% sobre o valor convertido' },
              { icon: '🌍', text: 'Destinatário recebe em USD, AED ou EUR em até 1h' },
            ].map(({ icon, text }) => (
              <div key={text} style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
                <span style={{ fontSize: 16, flexShrink: 0 }}>{icon}</span>
                <span style={{ fontSize: 12, color: 'var(--t2)', lineHeight: 1.4 }}>{text}</span>
              </div>
            ))}
          </div>

          {/* Supported countries */}
          <div style={{ marginTop: 12, background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 16, overflow: 'hidden' }}>
            <div style={{ padding: '12px 16px', borderBottom: '1px solid var(--border)' }}>
              <p style={{ fontWeight: 700, fontSize: 13, color: 'var(--t1)', margin: 0 }}>Países suportados</p>
            </div>
            {[
              { flag: '🇺🇸', name: 'Estados Unidos', cur: 'USD', time: '~1h', status: 'live' },
              { flag: '🇦🇪', name: 'Emirados Árabes', cur: 'AED', time: '~30min', status: 'live' },
              { flag: '🇪🇺', name: 'União Europeia', cur: 'EUR', time: 'D+1', status: 'live' },
              { flag: '🇬🇧', name: 'Reino Unido', cur: 'GBP', time: 'D+1', status: 'soon' },
              { flag: '🇨🇳', name: 'China', cur: 'CNY', time: 'D+2', status: 'soon' },
            ].map(({ flag, name, cur, time, status }) => (
              <div key={cur} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 16px', borderBottom: '1px solid var(--border)' }}>
                <span style={{ fontSize: 20, flexShrink: 0 }}>{flag}</span>
                <div style={{ flex: 1 }}>
                  <p style={{ fontWeight: 600, fontSize: 13, color: 'var(--t1)', margin: '0 0 1px' }}>{name}</p>
                  <p style={{ fontSize: 11, color: 'var(--t3)', margin: 0 }}>{cur} · {time}</p>
                </div>
                <span style={{
                  fontSize: 9, fontWeight: 700, padding: '2px 6px', borderRadius: 6,
                  color: status === 'live' ? '#34D399' : 'var(--t3)',
                  background: status === 'live' ? 'rgba(52,211,153,0.15)' : 'var(--surface-2)',
                }}>
                  {status === 'live' ? '● LIVE' : '⏱ EM BREVE'}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

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
