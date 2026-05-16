import { useState, useEffect, useRef } from 'react'
import { X, Sparkles, Send, TrendingUp, AlertTriangle, CheckCircle } from 'lucide-react'
import { fmtBRL, totalBalance } from '../data/mock.js'

const ZION_MESSAGES = (store) => {
  const balance = store?.account?.balance || 0
  const crypto  = store?.crypto || []
  const btc     = crypto.find(c => c.symbol === 'BTC')
  const kyc     = store?.kyc

  return [
    {
      type: 'insight',
      icon: TrendingUp,
      color: 'var(--green)',
      text: `Seu saldo total cresceu ${btc?.change24h || 3.2}% nas últimas 24h. O Bitcoin está impulsionando o portfólio — considere rebalancear se a exposição cripto superar 20%.`,
    },
    {
      type: 'alert',
      icon: AlertTriangle,
      color: 'var(--gold)',
      text: `Detectei 2 pagamentos recorrentes próximos ao vencimento. Certifique-se de que o saldo disponível (${fmtBRL(balance)}) cubra as saídas.`,
    },
    {
      type: 'positive',
      icon: CheckCircle,
      color: 'var(--green)',
      text: `KYC ${kyc?.level} ativo — limites ampliados. Você ainda tem ${fmtBRL((kyc?.limits?.daily || 10000000) - (kyc?.used?.daily || 0))} disponíveis no limite diário.`,
    },
  ]
}

const QUICK_PROMPTS = [
  'Qual minha posição em BTC?',
  'Resumo da semana',
  'Alertas de compliance',
  'Sugestão de rebalanceamento',
]

export default function ZionPanel({ store, onClose, showToast }) {
  const [messages, setMessages] = useState([])
  const [input,    setInput]    = useState('')
  const [typing,   setTyping]   = useState(false)
  const bottomRef = useRef(null)

  useEffect(() => {
    const insights = ZION_MESSAGES(store)
    setTimeout(() => {
      setMessages([{
        role: 'zion',
        text: `Olá, ${store?.user?.name?.split(' ')[0] || 'Investidor'}! Sou o Zion, sua IA financeira. Aqui estão os insights do momento:`,
        id: Date.now(),
      }])
      insights.forEach((ins, i) => {
        setTimeout(() => {
          setMessages(prev => [...prev, { role:'zion', ...ins, id: Date.now()+i }])
        }, 600 + i * 400)
      })
    }, 200)
  }, [])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior:'smooth' })
  }, [messages, typing])

  const sendMessage = (text) => {
    if (!text.trim()) return
    const userMsg = { role:'user', text: text.trim(), id: Date.now() }
    setMessages(prev => [...prev, userMsg])
    setInput('')
    setTyping(true)

    setTimeout(() => {
      setTyping(false)
      const resp = getZionResponse(text, store)
      setMessages(prev => [...prev, { role:'zion', text: resp, id: Date.now(), type:'insight', icon: Sparkles, color:'var(--green)' }])
    }, 1400)
  }

  return (
    <div className="zion-panel">
      {/* Header */}
      <div style={{ padding:'16px 20px', borderBottom:'1px solid var(--border)', display:'flex', alignItems:'center', gap:10 }}>
        <div style={{ width:36, height:36, borderRadius:'50%', background:'radial-gradient(circle at 40% 35%, #1aff8a, #0a8a4a 40%, #041a0e)', boxShadow:'0 0 16px rgba(0,232,122,0.4)', display:'flex', alignItems:'center', justifyContent:'center', flexShrink:0 }}>
          <Sparkles size={14} color='white' />
        </div>
        <div style={{ flex:1 }}>
          <p style={{ fontWeight:700, fontSize:14, fontFamily:'Syne' }}>Zion AI</p>
          <div style={{ display:'flex', alignItems:'center', gap:6 }}>
            <div className="pulse-dot" />
            <span style={{ fontSize:11, color:'var(--green)' }}>Online · Analisando portfólio</span>
          </div>
        </div>
        <button className="btn-icon" onClick={onClose}><X size={15}/></button>
      </div>

      {/* Messages */}
      <div style={{ flex:1, overflowY:'auto', padding:'16px 20px', display:'flex', flexDirection:'column', gap:12 }}>
        {messages.map(msg => (
          <MessageBubble key={msg.id} msg={msg} />
        ))}
        {typing && (
          <div style={{ display:'flex', gap:8, alignItems:'flex-start' }}>
            <div style={{ width:28, height:28, borderRadius:'50%', background:'radial-gradient(circle, #1aff8a, #0a8a4a)', flexShrink:0 }} />
            <div style={{ background:'var(--card)', border:'1px solid var(--border)', borderRadius:'12px 12px 12px 4px', padding:'10px 14px', display:'flex', gap:4 }}>
              <div className="typing-dot" />
              <div className="typing-dot" />
              <div className="typing-dot" />
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      {/* Quick prompts */}
      <div style={{ padding:'8px 20px', borderTop:'1px solid var(--border)', display:'flex', gap:6, flexWrap:'wrap' }}>
        {QUICK_PROMPTS.map(q => (
          <button key={q} onClick={() => sendMessage(q)} style={{
            padding:'5px 10px', borderRadius:999, border:'1px solid var(--border)',
            background:'rgba(255,255,255,0.03)', color:'var(--t2)', fontSize:11,
            cursor:'pointer', transition:'all 0.15s', whiteSpace:'nowrap',
          }}
            onMouseEnter={e => { e.target.style.borderColor='var(--border-h)'; e.target.style.color='var(--green)' }}
            onMouseLeave={e => { e.target.style.borderColor='var(--border)'; e.target.style.color='var(--t2)' }}
          >{q}</button>
        ))}
      </div>

      {/* Input */}
      <div style={{ padding:'12px 20px 20px', borderTop:'1px solid var(--border)', display:'flex', gap:10 }}>
        <input className="input" placeholder="Pergunte ao Zion..." value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && !e.shiftKey && sendMessage(input)}
          style={{ flex:1 }}
        />
        <button className="btn-primary" style={{ padding:'10px 14px', flexShrink:0 }} onClick={() => sendMessage(input)}>
          <Send size={14} />
        </button>
      </div>
    </div>
  )
}

function MessageBubble({ msg }) {
  const isUser = msg.role === 'user'
  const Icon   = msg.icon

  if (isUser) {
    return (
      <div style={{ display:'flex', justifyContent:'flex-end' }}>
        <div style={{ background:'rgba(0,232,122,0.12)', border:'1px solid rgba(0,232,122,0.2)', borderRadius:'12px 12px 4px 12px', padding:'10px 14px', maxWidth:'80%' }}>
          <p style={{ fontSize:13, color:'var(--t1)', lineHeight:1.5 }}>{msg.text}</p>
        </div>
      </div>
    )
  }

  return (
    <div style={{ display:'flex', gap:8, alignItems:'flex-start', animation:'slideUp 0.3s ease' }}>
      <div style={{ width:28, height:28, borderRadius:'50%', background:'radial-gradient(circle, #1aff8a, #0a8a4a)', flexShrink:0, marginTop:2 }} />
      <div>
        {Icon && (
          <div style={{ display:'flex', alignItems:'center', gap:6, marginBottom:6 }}>
            <Icon size={12} color={msg.color || 'var(--green)'} />
            <span style={{ fontSize:10, fontWeight:700, color: msg.color || 'var(--green)', textTransform:'uppercase', letterSpacing:'0.08em' }}>
              {msg.type === 'alert' ? 'ALERTA' : msg.type === 'positive' ? 'POSITIVO' : 'INSIGHT'}
            </span>
          </div>
        )}
        <div style={{ background:'var(--card)', border:'1px solid var(--border)', borderRadius:'12px 12px 12px 4px', padding:'10px 14px', maxWidth:'85%' }}>
          <p style={{ fontSize:13, color:'var(--t2)', lineHeight:1.6 }}>{msg.text}</p>
        </div>
      </div>
    </div>
  )
}

function getZionResponse(input, store) {
  const lower = input.toLowerCase()
  const btc   = store?.crypto?.find(c => c.symbol === 'BTC')
  const bal   = store?.account?.balance || 0

  if (lower.includes('btc') || lower.includes('bitcoin')) {
    return `Sua posição em Bitcoin: ${btc?.amount?.toFixed(4) || 0} BTC ≈ ${fmtBRL((btc?.amount || 0) * (btc?.price || 0) * 100)}. Variação 24h: +${btc?.change24h || 3.2}%. O nível de suporte técnico está em R$ ${((btc?.price || 345820) * 0.95).toLocaleString('pt-BR', {maximumFractionDigits:0})}.`
  }
  if (lower.includes('semana') || lower.includes('resumo')) {
    return `Resumo da semana: saldo disponível ${fmtBRL(bal)}, portfólio cripto +${btc?.change24h || 3.2}%, 2 pagamentos agendados. Fluxo líquido: +${fmtBRL(Math.abs(bal * 0.05))}. Performance dentro do esperado.`
  }
  if (lower.includes('compliance') || lower.includes('alerta')) {
    return `2 alertas de compliance: (1) verificação de endereço pendente para pré-cadastro ID #a3f2b1c0, (2) movimentação de R$ 5.000 para KYC BASIC — dentro do limite mas próximo ao teto diário.`
  }
  if (lower.includes('rebalanceamento') || lower.includes('sugestão')) {
    return `Sugestão: sua exposição cripto representa ~${Math.round(((btc?.amount || 0.54) * (btc?.price || 345820) * 100 / bal) * 100) || 7}% do portfólio total. Para perfil moderado, recomendo máximo 15%. Considere converter parte do BTC para CDB com liquidez diária.`
  }
  return `Analisei sua solicitação. Minha recomendação é monitorar o saldo disponível (${fmtBRL(bal)}) e manter pelo menos 30% em ativos líquidos. Posso ajudar com simulações específicas.`
}
