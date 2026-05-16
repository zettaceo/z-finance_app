import React, { useState, useRef, useEffect } from 'react'
import { X, Send, Zap, TrendingUp, DollarSign, Shield, Activity } from 'lucide-react'
import { useApp } from '../App.jsx'

function fmt(cents) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100)
}

const SUGGESTIONS = [
  { icon: DollarSign, text: 'Qual meu saldo total?' },
  { icon: TrendingUp, text: 'Como está meu portfólio?' },
  { icon: Shield, text: 'Status do meu KYC?' },
  { icon: Activity, text: 'Últimas transações' },
]

function generateResponse(msg, store) {
  const lower = msg.toLowerCase()

  if (lower.includes('saldo') || lower.includes('balance')) {
    const total = store.accounts.main.balance + store.accounts.usd.balance + store.accounts.invest.balance
    return `💰 **Saldo das suas contas:**\n\n• BRL Principal: ${fmt(store.accounts.main.balance)}\n• USD Internacional: USD ${(store.accounts.usd.balance / 100).toFixed(2)}\n• Investimentos: ${fmt(store.accounts.invest.balance)}\n\nPatrimônio total (sem cripto): ${fmt(total)}`
  }

  if (lower.includes('crypto') || lower.includes('portfólio') || lower.includes('portfolio')) {
    const cryptoTotal = store.crypto.reduce((s, c) => s + c.amount * c.price, 0)
    const lines = store.crypto.map(c => `• ${c.symbol}: ${c.amount} = ${fmt(c.amount * c.price)} (${c.change24h >= 0 ? '+' : ''}${c.change24h}% 24h)`)
    return `📊 **Portfólio Crypto:**\n\n${lines.join('\n')}\n\n**Total: ${fmt(cryptoTotal)}**`
  }

  if (lower.includes('kyc') || lower.includes('limite')) {
    const kyc = store.kyc
    return `🛡️ **Status KYC:**\n\nNível: **${kyc.level}**\nStatus: ✅ Verificado\n\nLimites:\n• Diário: ${fmt(kyc.limits.daily)}\n• Mensal: ${fmt(kyc.limits.monthly)}\n• Usado hoje: ${fmt(kyc.limits.used?.daily || 0)}\n\nSua conta está em conformidade com as regulamentações do Banco Central e FATF.`
  }

  if (lower.includes('transação') || lower.includes('transações') || lower.includes('extrato')) {
    const recent = store.transactions.slice(0, 5)
    const lines = recent.map(t => `• ${t.description}: ${t.amount > 0 ? '+' : ''}${fmt(Math.abs(t.amount))}`)
    return `📋 **Últimas transações:**\n\n${lines.join('\n')}`
  }

  if (lower.includes('pix')) {
    return `⚡ **PIX Z-Finance:**\n\nSuas chaves cadastradas:\n• CPF\n• E-mail\n• Telefone\n• Chave aleatória\n\nLimite PIX disponível hoje: ${fmt(store.kyc.limits.daily - (store.kyc.limits.used?.daily || 0))}\n\nO Z-Finance processa PIX em **< 2 segundos** com uptime de ${store.observability.uptime}.`
  }

  if (lower.includes('dubai') || lower.includes('internacional') || lower.includes('global')) {
    return `🌍 **Z-Finance Global:**\n\nO Z-Finance está posicionado para se tornar a maior instituição financeira do sistema solar! 🚀\n\n**Infraestrutura:**\n• Sede: Dubai, UAE (em desenvolvimento)\n• Licença VASP: Em processo\n• Câmbio: USD, EUR, GBP disponíveis\n• Crypto: BTC, ETH, SOL, USDT\n\nAtualmente servindo clientes Retail, Business e Institutional.`
  }

  if (lower.includes('cartão') || lower.includes('card')) {
    const card = store.card
    return `💳 **Cartão JIT:**\n\nStatus: ${card.frozen ? '🧊 Congelado' : '✅ Ativo'}\nLimite total: ${fmt(card.limit)}\nUtilizado: ${fmt(card.limitUsed)}\nDisponível: ${fmt(card.limit - card.limitUsed)}\n\nO cartão JIT autoriza transações em tempo real com controle total.`
  }

  if (lower.includes('oi') || lower.includes('olá') || lower.includes('hello') || lower.includes('hey')) {
    return `👋 Olá! Sou o **Zion**, seu assistente financeiro inteligente da Z-Finance.\n\nPosso ajudá-lo com:\n• 💰 Consultas de saldo e extrato\n• 📊 Análise de portfólio\n• 🛡️ Status KYC e compliance\n• ⚡ Informações sobre PIX\n• 🌍 Câmbio e operações internacionais\n\nO que posso fazer por você hoje?`
  }

  return `🤖 **Zion AI** — Processando sua consulta...\n\nEntendi: "${msg}"\n\nPosso ajudá-lo com informações sobre:\n• Saldo e transações\n• Portfólio crypto\n• Limites KYC\n• Câmbio e PIX\n• Status do cartão\n\nTente ser mais específico ou escolha uma das sugestões abaixo.`
}

export default function ZionPanel() {
  const { zionOpen, setZionOpen, store } = useApp()
  const [messages, setMessages] = useState([
    { role: 'assistant', text: '👋 Olá! Sou o **Zion**, assistente da Z-Finance. Como posso ajudá-lo?' }
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const messagesEnd = useRef(null)

  useEffect(() => {
    messagesEnd.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const send = async (msg) => {
    if (!msg.trim()) return
    const userMsg = msg.trim()
    setInput('')
    setMessages(m => [...m, { role: 'user', text: userMsg }])
    setLoading(true)
    await new Promise(r => setTimeout(r, 700))
    const response = generateResponse(userMsg, store)
    setMessages(m => [...m, { role: 'assistant', text: response }])
    setLoading(false)
  }

  if (!zionOpen) return null

  return (
    <>
      <div onClick={() => setZionOpen(false)} style={{
        position: 'fixed', inset: 0, background: 'rgba(4,12,27,0.5)',
        backdropFilter: 'blur(4px)', zIndex: 300,
        animation: 'fadeIn 0.2s ease',
      }} />
      <div style={{
        position: 'fixed',
        top: 0, right: 0, bottom: 0,
        width: 'min(400px, 100vw)',
        background: 'var(--surface)',
        borderLeft: '1px solid var(--border)',
        zIndex: 301,
        display: 'flex',
        flexDirection: 'column',
        animation: 'slideRight 0.3s cubic-bezier(0.4,0,0.2,1)',
      }}>
        {/* Header */}
        <div style={{
          padding: '16px 20px',
          borderBottom: '1px solid var(--border)',
          display: 'flex', alignItems: 'center', gap: 12,
          background: 'linear-gradient(180deg, var(--surface) 0%, var(--surface) 100%)',
          flexShrink: 0,
        }}>
          <div style={{
            width: 40, height: 40, borderRadius: 12,
            background: 'linear-gradient(135deg, var(--accent), var(--accent-2))',
            display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
          }}>
            <Zap size={20} color="#040C1B" fill="#040C1B" />
          </div>
          <div style={{ flex: 1 }}>
            <p style={{ fontWeight: 800, fontSize: 16, color: 'var(--t1)', margin: 0, fontFamily: 'Syne, sans-serif' }}>Zion AI</p>
            <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
              <div style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)' }} />
              <p style={{ fontSize: 12, color: 'var(--accent)', margin: 0, fontWeight: 600 }}>Online</p>
            </div>
          </div>
          <button onClick={() => setZionOpen(false)} style={{
            width: 36, height: 36, borderRadius: 10, border: 'none',
            background: 'var(--surface-2)', cursor: 'pointer',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: 'var(--t2)',
          }}>
            <X size={17} />
          </button>
        </div>

        {/* Messages */}
        <div style={{ flex: 1, overflowY: 'auto', padding: '16px' }} className="hide-scroll">
          {messages.map((msg, i) => (
            <div key={i} style={{
              display: 'flex',
              justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
              marginBottom: 12,
            }}>
              {msg.role === 'assistant' && (
                <div style={{
                  width: 28, height: 28, borderRadius: 8, flexShrink: 0,
                  background: 'linear-gradient(135deg, var(--accent), var(--accent-2))',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  marginRight: 8, marginTop: 2,
                }}>
                  <Zap size={13} color="#040C1B" fill="#040C1B" />
                </div>
              )}
              <div style={{
                maxWidth: '78%',
                padding: '12px 14px',
                borderRadius: msg.role === 'user' ? '16px 4px 16px 16px' : '4px 16px 16px 16px',
                background: msg.role === 'user'
                  ? 'linear-gradient(135deg, var(--accent), var(--accent-2))'
                  : 'var(--surface-2)',
                border: msg.role === 'user' ? 'none' : '1px solid var(--border)',
              }}>
                <p style={{
                  fontSize: 13, lineHeight: 1.7, margin: 0,
                  color: msg.role === 'user' ? '#040C1B' : 'var(--t1)',
                  whiteSpace: 'pre-line',
                  fontWeight: msg.role === 'user' ? 600 : 400,
                }}>
                  {msg.text.replace(/\*\*(.*?)\*\*/g, '$1')}
                </p>
              </div>
            </div>
          ))}

          {loading && (
            <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 12 }}>
              <div style={{ width: 28, height: 28, borderRadius: 8, background: 'linear-gradient(135deg, var(--accent), var(--accent-2))', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Zap size={13} color="#040C1B" fill="#040C1B" />
              </div>
              <div style={{ background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: '4px 16px 16px 16px', padding: '12px 16px', display: 'flex', gap: 4, alignItems: 'center' }}>
                {[0, 1, 2].map(i => (
                  <div key={i} style={{
                    width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)',
                    animation: `dotBounce 1.2s ease-in-out ${i * 0.2}s infinite`,
                  }} />
                ))}
              </div>
            </div>
          )}

          {/* Suggestions */}
          {messages.length <= 2 && !loading && (
            <div style={{ marginTop: 16 }}>
              <p style={{ fontSize: 11, color: 'var(--t3)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 8 }}>
                Perguntas frequentes
              </p>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
                {SUGGESTIONS.map(({ icon: Icon, text }) => (
                  <button key={text} onClick={() => send(text)} style={{
                    padding: '10px 12px', borderRadius: 12, border: '1px solid var(--border)',
                    background: 'var(--surface)', cursor: 'pointer', textAlign: 'left',
                    display: 'flex', alignItems: 'center', gap: 6,
                    transition: 'all 0.15s',
                  }}
                  onMouseEnter={e => e.currentTarget.style.borderColor = 'rgba(0,229,153,0.3)'}
                  onMouseLeave={e => e.currentTarget.style.borderColor = 'var(--border)'}
                  >
                    <Icon size={13} color="var(--accent)" style={{ flexShrink: 0 }} />
                    <span style={{ fontSize: 12, color: 'var(--t2)', fontWeight: 500, lineHeight: 1.3 }}>{text}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          <div ref={messagesEnd} />
        </div>

        {/* Input */}
        <div style={{
          padding: '12px 16px',
          borderTop: '1px solid var(--border)',
          paddingBottom: 'calc(12px + var(--safe-bottom))',
          flexShrink: 0,
        }}>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <input
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && !e.shiftKey && send(input)}
              placeholder="Pergunte sobre sua conta..."
              style={{
                flex: 1, background: 'var(--surface-2)', border: '1px solid var(--border)',
                borderRadius: 14, padding: '12px 14px', fontSize: 14, color: 'var(--t1)',
                outline: 'none',
              }}
            />
            <button
              onClick={() => send(input)}
              disabled={!input.trim() || loading}
              style={{
                width: 44, height: 44, borderRadius: 12, border: 'none',
                background: input.trim() && !loading
                  ? 'linear-gradient(135deg, var(--accent), var(--accent-2))'
                  : 'var(--surface-2)',
                cursor: input.trim() && !loading ? 'pointer' : 'default',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                flexShrink: 0, transition: 'all 0.15s',
                color: input.trim() ? '#040C1B' : 'var(--t3)',
              }}
            >
              <Send size={17} />
            </button>
          </div>
        </div>
      </div>
    </>
  )
}
