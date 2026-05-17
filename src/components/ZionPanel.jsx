import React, { useState, useRef, useEffect } from 'react'
import { X, Send, Zap, TrendingUp, DollarSign, Shield, Activity, TrendingDown, Sparkles, Clock, ArrowRight } from 'lucide-react'
import { useApp } from '../App.jsx'

function fmt(cents) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100)
}

function generateResponse(msg, store) {
  const lower = msg.toLowerCase()
  const c = store.accounts
  const btc = store.crypto.find(x => x.symbol === 'BTC')

  if (lower.includes('saldo') || lower.includes('balance') || lower.includes('ver saldo')) {
    const aed = c.aed ? `\n• AED Dubai: AED ${(c.aed.balance/100).toFixed(2)}` : ''
    const total = c.main.balance + c.invest.balance
    return `💰 **Saldo das suas contas:**\n\n• BRL Principal: ${fmt(c.main.balance)}\n• USD Internacional: USD ${(c.usd.balance/100).toFixed(2)}${aed}\n• Investimentos: ${fmt(c.invest.balance)}\n\nPatrimônio total (contas): ${fmt(total)}`
  }

  if (lower.includes('btc') || lower.includes('bitcoin')) {
    if (btc) {
      const val = Math.round(btc.amount * btc.price * 100)
      const gain = Math.round(btc.amount * btc.price * btc.change24h / 100 * 100)
      return `₿ **Bitcoin (BTC):**\n\nSaldo: ${btc.amount.toFixed(4)} BTC\nValor: ${fmt(val)}\nVariação 24h: ${btc.change24h >= 0 ? '+' : ''}${btc.change24h}% (${btc.change24h >= 0 ? '+' : ''}${fmt(Math.abs(gain))})\n\nPreço atual: ${fmt(btc.price * 100)}\n\n💡 Dica: Use o simulador de crédito com BTC como colateral para obter taxas menores.`
    }
  }

  if (lower.includes('crypto') || lower.includes('portfólio') || lower.includes('portfolio') || lower.includes('cripto')) {
    const total = store.crypto.reduce((s, c) => s + c.amount * c.price * 100, 0)
    const lines = store.crypto.map(c => `• ${c.symbol}: ${fmt(Math.round(c.amount * c.price * 100))} (${c.change24h >= 0 ? '+' : ''}${c.change24h}% 24h)`)
    return `📊 **Portfólio Cripto:**\n\n${lines.join('\n')}\n\n**Total: ${fmt(total)}**\n\n💡 Acesse a aba Investir para detalhes e operações.`
  }

  if (lower.includes('kyc') || lower.includes('limite') || lower.includes('limite pix')) {
    const kyc = store.kyc
    const persona = store.persona || 'BUSINESS'
    const limit = persona === 'INSTITUTIONAL' ? 'Ilimitado' : fmt(kyc.limits.daily)
    return `🛡️ **Status KYC e Limites:**\n\nNível: **${kyc.level}** ✅\nPersona: **${persona}**\n\nLimite PIX diário: ${limit}\nLimite mensal: ${fmt(kyc.limits.monthly)}\nUsado hoje: ${fmt(kyc.limits.used?.daily || 1250000)}\n\nCompliance: FATF LOW · Não-PEP · Sem Sanções\n\nSua conta está em plena conformidade.`
  }

  if (lower.includes('transação') || lower.includes('transações') || lower.includes('extrato') || lower.includes('últimas')) {
    const recent = store.transactions.slice(0, 5)
    const lines = recent.map(t => `• ${t.desc || t.description}: ${t.amount > 0 ? '+' : ''}${fmt(Math.abs(t.amount))}`)
    return `📋 **Últimas 5 transações:**\n\n${lines.join('\n')}\n\nAcesse a Home para ver o extrato completo.`
  }

  if (lower.includes('pix') || lower.includes('enviar pix')) {
    return `⚡ **PIX Z-Finance:**\n\nSuas chaves cadastradas:\n• CPF · E-mail · Telefone · Aleatória\n\nLimite disponível hoje: ${fmt(store.kyc.limits.daily - (store.kyc.limits.used?.daily || 1250000))}\n\nO Z-Finance processa PIX em **< 2 segundos** com ${store.observability.uptime} de uptime.\n\n💡 Use "PIX Internacional" no Mover para enviar para USD, AED ou EUR.`
  }

  if (lower.includes('dirham') || (lower.includes('aed') && (lower.includes('saldo') || lower.includes('conta') || lower.includes('dubai')))) {
    const aed = c.aed
    const aedRate = store.market.AED.rate
    if (aed) {
      const brlEquiv = Math.round(aed.balance * aedRate / 100)
      return `🇦🇪 **Conta AED (Dubai):**\n\nSaldo: AED ${(aed.balance/100).toFixed(2)}\nEquivalente: ${fmt(brlEquiv)}\n\nCotação atual: R$ ${(aedRate/100).toFixed(2)} por AED 1,00\nJurisdição: 🇦🇪 ARE — VASP ativo\n\n💡 Você pode enviar BRL via PIX e converter automaticamente para AED na ação "PIX → AED" da aba Mover.`
    }
  }

  if (lower.includes('pix internacional') || lower.includes('pix intl') || lower.includes('enviar dubai') || lower.includes('enviar exterior')) {
    return `🌍 **PIX Internacional:**\n\nEnvie BRL via PIX e o destinatário recebe na moeda local em até D+1.\n\n**Países suportados:**\n• 🇺🇸 Estados Unidos (USD) — liquidação D+0\n• 🇦🇪 Emirados Árabes (AED) — liquidação D+0\n• 🇪🇺 Zona do Euro (EUR) — liquidação D+1 (SEPA)\n• 🇲🇽 México (MXN) — em breve\n• 🇦🇷 Argentina (ARS) — em breve\n\nTaxa: 1.5% sobre o valor + spread de câmbio transparente.\n\n💡 Acesse Mover → "PIX Intl" para iniciar o envio.`
  }

  if (lower.includes('câmbio') || lower.includes('cambio') || lower.includes('cotar') || lower.includes('dólar') || lower.includes('aed')) {
    const m = store.market
    return `💱 **Câmbio ao vivo:**\n\n🇺🇸 USD: R$ ${(m.USD.rate/100).toFixed(2)} (${m.USD.change >= 0 ? '+' : ''}${m.USD.change}%)\n🇦🇪 AED: R$ ${(m.AED.rate/100).toFixed(2)} (${m.AED.change >= 0 ? '+' : ''}${m.AED.change}%)\n🇪🇺 EUR: R$ ${(m.EUR.rate/100).toFixed(2)} (${m.EUR.change >= 0 ? '+' : ''}${m.EUR.change}%)\n🇬🇧 GBP: R$ ${(m.GBP.rate/100).toFixed(2)} (${m.GBP.change >= 0 ? '+' : ''}${m.GBP.change}%)\n\n💡 Use o PIX Internacional para enviar BRL e receber em qualquer moeda.`
  }

  if (lower.includes('cartão') || lower.includes('card') || lower.includes('jit')) {
    const card = store.card
    return `💳 **Cartão JIT:**\n\nStatus: ${card.frozen ? '🧊 Congelado' : '✅ Ativo'}\nLimite: ${fmt(card.limit)}\nUtilizado: ${fmt(card.limitUsed)}\nDisponível: ${fmt(card.limit - card.limitUsed)}\n\nO cartão JIT não tem saldo pré-carregado — cada transação é autorizada em tempo real.`
  }

  if (lower.includes('crédito') || lower.includes('credito') || lower.includes('empréstimo') || lower.includes('score') || lower.includes('z-score')) {
    const cr = store.credit
    const active = (cr.loans || []).filter(l => l.status === 'ACTIVE')
    const activeBlock = active.length > 0
      ? `\n\n**Empréstimos ativos (${active.length}):**\n${active.map(l => `• ${l.type}: ${fmt(l.balance)} restantes · próxima parcela em ${Math.max(0, Math.ceil((new Date(l.nextDue) - Date.now()) / 86400000))} dias`).join('\n')}`
      : ''
    return `🏦 **Z-Score e Crédito:**\n\nZ-Score: **${cr.score}** (${cr.scoreLabel})\nLimite total: ${fmt(cr.limit)}\nUtilizado: ${fmt(cr.limitUsed)}\nDisponível: ${fmt(cr.limit - cr.limitUsed)}${activeBlock}\n\n**Ofertas disponíveis:**\n• Pessoal: até ${fmt(2000000)} a 1.8% a.m.\n• Crypto colateral: até ${fmt(3000000)} a 1.2% a.m.\n• Empresarial: até ${fmt(10000000)} a 2.1% a.m.\n\n💡 Acesse a aba Crédito para simular e contratar.`
  }

  if (lower.includes('rentabilizar') || lower.includes('investir') || lower.includes('investimento') || lower.includes('rendimento')) {
    return `📈 **Investimentos:**\n\nCDB Banco Z: ${fmt(store.investments.cdb)} ✅\nTesouro SELIC: ${fmt(store.investments.tesouro)} ✅\n\nSaldo em conta: ${fmt(c.main.balance)}\n\n💡 Você tem ${fmt(c.main.balance)} na conta corrente. Rentabilizando em CDB a 14% a.a., geraria aprox. ${fmt(Math.round(c.main.balance * 0.14 / 12))} por mês.\n\nAcesse Investir para alocar.`
  }

  if (lower.includes('z-pass') || lower.includes('passaporte') || lower.includes('zpass')) {
    return `🪪 **Z-Pass (Identidade Financeira):**\n\nSeu passaporte financeiro digital está ativo.\n\nJurisdições: 🇧🇷 BRA ✅ · 🇦🇪 ARE ✅\nKYC: FULL verificado\nVASP: Ativo\nFATF: Baixo risco\n\nAcesse "Z-Pass" no menu lateral para ver e exportar seu passaporte.`
  }

  if (lower.includes('oi') || lower.includes('olá') || lower.includes('hello') || lower.includes('hey') || lower.includes('bom dia') || lower.includes('boa tarde')) {
    const persona = store.persona || 'BUSINESS'
    return `👋 Olá, ${store.user.name.split(' ')[0]}! Sou o **Zion**, seu assistente financeiro.\n\nPerfil ativo: **${persona}**\n\nPosso ajudá-lo com:\n• 💰 Saldo e extrato\n• 📊 Portfólio cripto\n• 💱 Câmbio ao vivo\n• ⚡ PIX e transferências\n• 🏦 Crédito e Z-Score\n• 🪪 Z-Pass e compliance\n\nO que posso fazer por você?`
  }

  if (lower.includes('dubai') || lower.includes('internacional') || lower.includes('global') || lower.includes('são paulo')) {
    return `🌍 **Z-Finance Global:**\n\nA plataforma financeira multi-jurisdicional.\n\n🇧🇷 Brasil: Conta BRL + PIX\n🇦🇪 Dubai: Conta AED, jurisdição ARE ativa\n🇺🇸 USA: PIX Internacional disponível\n🇪🇺 Europa: SEPA D+1\n\nNão sou apenas um banco — sou uma infraestrutura financeira global acessível via PIX.`
  }

  return `🤖 **Zion AI** — Processando: "${msg}"\n\nPosso ajudá-lo com:\n• Ver saldo · Portfólio cripto\n• Câmbio · PIX Internacional\n• Crédito · Z-Score\n• Z-Pass · Cartão JIT\n\nTente uma das sugestões abaixo ↓`
}

function InsightCards({ store, onSend }) {
  const btc = store.crypto.find(x => x.symbol === 'BTC')
  const btcGain = btc ? Math.round(btc.amount * btc.price * btc.change24h / 100 * 100) : 0
  const idleCash = store.accounts.main.balance
  const loan = store.credit?.loans?.find(l => l.status === 'ACTIVE')
  const daysToNextDue = loan ? Math.ceil((new Date(loan.nextDue) - Date.now()) / 86400000) : null

  const insights = [
    btc && btcGain !== 0 && {
      icon: btcGain > 0 ? TrendingUp : TrendingDown,
      color: btcGain > 0 ? '#34D399' : '#F87171',
      title: `BTC ${btcGain > 0 ? 'valorizou' : 'caiu'} ${fmt(Math.abs(btcGain))} hoje`,
      sub: `${btc.change24h >= 0 ? '+' : ''}${btc.change24h}% nas últimas 24h`,
      action: 'Como está meu portfólio?',
    },
    idleCash > 500000 && {
      icon: DollarSign,
      color: '#818CF8',
      title: `${fmt(idleCash)} na conta corrente`,
      sub: 'Rentabilize com CDB ou Tesouro SELIC',
      action: 'Quero rentabilizar meu saldo',
    },
    loan && daysToNextDue !== null && daysToNextDue <= 10 && {
      icon: Clock,
      color: '#F59E0B',
      title: `Parcela de crédito em ${daysToNextDue} dias`,
      sub: `${fmt(Math.round(loan.amount / loan.months))} — vence em ${new Date(loan.nextDue).toLocaleDateString('pt-BR')}`,
      action: 'Meu crédito ativo',
    },
  ].filter(Boolean).slice(0, 3)

  if (insights.length === 0) return null

  return (
    <div style={{ marginBottom: 16 }}>
      <p style={{ fontSize: 11, color: 'var(--t3)', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em', margin: '0 0 8px' }}>
        <Sparkles size={10} style={{ verticalAlign: 'middle', marginRight: 4 }} />
        Insights proativos
      </p>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {insights.map(({ icon: Icon, color, title, sub, action }) => (
          <button key={title} onClick={() => onSend(action)} style={{
            display: 'flex', alignItems: 'center', gap: 12,
            padding: '12px 14px', borderRadius: 14,
            background: `${color}10`, border: `1px solid ${color}25`,
            cursor: 'pointer', textAlign: 'left', width: '100%',
            transition: 'all 0.15s',
          }}>
            <div style={{
              width: 36, height: 36, borderRadius: 10, flexShrink: 0,
              background: `${color}20`,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
            }}>
              <Icon size={17} color={color} />
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <p style={{ fontSize: 13, fontWeight: 700, color: 'var(--t1)', margin: '0 0 1px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{title}</p>
              <p style={{ fontSize: 11, color: 'var(--t3)', margin: 0 }}>{sub}</p>
            </div>
            <ArrowRight size={14} color={color} style={{ flexShrink: 0, opacity: 0.6 }} />
          </button>
        ))}
      </div>
    </div>
  )
}

const QUICK_CHIPS = [
  { label: 'Ver saldo', msg: 'Ver saldo' },
  { label: 'Portfólio', msg: 'Como está meu portfólio?' },
  { label: 'Câmbio', msg: 'Cotar câmbio' },
  { label: 'PIX', msg: 'Enviar PIX' },
  { label: 'Crédito', msg: 'Meu crédito ativo' },
  { label: 'Z-Pass', msg: 'Status Z-Pass' },
]

export default function ZionPanel() {
  const { zionOpen, setZionOpen, store } = useApp()
  const [messages, setMessages] = useState(() => {
    try {
      const saved = sessionStorage.getItem('zion_history')
      return saved ? JSON.parse(saved) : [{ role: 'assistant', text: `👋 Olá, ${store.user.name.split(' ')[0]}! Sou o **Zion**, seu assistente financeiro inteligente da Z-Finance.\n\nTenho 3 insights proativos para você hoje. Clique em qualquer um ou faça uma pergunta.` }]
    } catch { return [{ role: 'assistant', text: '👋 Olá! Sou o Zion, assistente da Z-Finance.' }] }
  })
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const messagesEnd = useRef(null)
  const isFirst = messages.length <= 1

  useEffect(() => {
    messagesEnd.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, loading])

  useEffect(() => {
    try { sessionStorage.setItem('zion_history', JSON.stringify(messages.slice(-20))) } catch {}
  }, [messages])

  const send = async (msg) => {
    if (!msg.trim()) return
    const userMsg = msg.trim()
    setInput('')
    setMessages(m => [...m, { role: 'user', text: userMsg }])
    setLoading(true)
    await new Promise(r => setTimeout(r, 600 + Math.random() * 400))
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
        width: 'min(420px, 100vw)',
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
          background: 'linear-gradient(180deg, rgba(0,229,153,0.04) 0%, var(--surface) 100%)',
          flexShrink: 0,
        }}>
          <div style={{
            width: 40, height: 40, borderRadius: 12,
            background: 'linear-gradient(135deg, var(--accent), var(--accent-2))',
            display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
            boxShadow: '0 4px 16px rgba(0,229,153,0.3)',
          }}>
            <Zap size={20} color="#040C1B" fill="#040C1B" />
          </div>
          <div style={{ flex: 1 }}>
            <p style={{ fontWeight: 800, fontSize: 16, color: 'var(--t1)', margin: 0, fontFamily: 'Syne, sans-serif' }}>Zion AI</p>
            <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
              <div style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--accent)', boxShadow: '0 0 6px var(--accent)' }} />
              <p style={{ fontSize: 11, color: 'var(--accent)', margin: 0, fontWeight: 600 }}>Assistente ativo · Dados em tempo real</p>
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
                maxWidth: '80%',
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

          {/* Thinking animation */}
          {loading && (
            <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 12 }}>
              <div style={{ width: 28, height: 28, borderRadius: 8, background: 'linear-gradient(135deg, var(--accent), var(--accent-2))', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Zap size={13} color="#040C1B" fill="#040C1B" />
              </div>
              <div style={{ background: 'var(--surface-2)', border: '1px solid var(--border)', borderRadius: '4px 16px 16px 16px', padding: '12px 16px' }}>
                <div style={{ display: 'flex', gap: 5, alignItems: 'center' }}>
                  {[0, 1, 2].map(i => (
                    <div key={i} style={{
                      width: 7, height: 7, borderRadius: '50%', background: 'var(--accent)',
                      animation: `dotBounce 1s ease-in-out ${i * 0.18}s infinite`,
                      opacity: 0.8,
                    }} />
                  ))}
                  <span style={{ fontSize: 11, color: 'var(--t3)', marginLeft: 4 }}>analisando dados...</span>
                </div>
              </div>
            </div>
          )}

          {/* Proactive insights (first load) */}
          {isFirst && !loading && (
            <InsightCards store={store} onSend={send} />
          )}

          {/* Suggestions */}
          {isFirst && !loading && (
            <div style={{ marginTop: 4 }}>
              <p style={{ fontSize: 11, color: 'var(--t3)', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 8 }}>
                Perguntas frequentes
              </p>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
                {[
                  { icon: DollarSign, text: 'Qual meu saldo total?' },
                  { icon: TrendingUp, text: 'Como está meu portfólio?' },
                  { icon: Shield, text: 'Status do meu KYC?' },
                  { icon: Activity, text: 'Últimas transações' },
                ].map(({ icon: Icon, text }) => (
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

        {/* Quick chips */}
        <div style={{
          padding: '8px 12px 4px',
          borderTop: '1px solid var(--border)',
          display: 'flex', gap: 6, overflowX: 'auto', flexShrink: 0,
        }} className="hide-scroll">
          {QUICK_CHIPS.map(({ label, msg }) => (
            <button key={label} onClick={() => send(msg)} style={{
              flexShrink: 0, padding: '6px 12px', borderRadius: 20,
              border: '1px solid var(--border)', background: 'var(--surface-2)',
              color: 'var(--t2)', fontSize: 12, fontWeight: 600, cursor: 'pointer',
              whiteSpace: 'nowrap', transition: 'all 0.15s',
            }}
            onMouseEnter={e => { e.currentTarget.style.borderColor = 'rgba(0,229,153,0.4)'; e.currentTarget.style.color = 'var(--accent)' }}
            onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--border)'; e.currentTarget.style.color = 'var(--t2)' }}
            >
              {label}
            </button>
          ))}
        </div>

        {/* Input */}
        <div style={{
          padding: '8px 16px',
          paddingBottom: 'calc(8px + var(--safe-bottom))',
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
