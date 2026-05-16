import { useState } from 'react'
import { runSimulation, fmtBRL } from '../data/mock.js'
import { TrendingUp, ArrowLeftRight, Zap, BarChart2, RefreshCw } from 'lucide-react'

const COIN_COLORS = {
  BTC:  { color:'#f7931a', bg:'rgba(247,147,26,0.12)',  border:'rgba(247,147,26,0.2)', letter:'₿' },
  ETH:  { color:'#627eea', bg:'rgba(98,126,234,0.12)',  border:'rgba(98,126,234,0.2)', letter:'Ξ' },
  USDT: { color:'#26a17b', bg:'rgba(38,161,123,0.12)',  border:'rgba(38,161,123,0.2)', letter:'₮' },
}

export default function Investir({ store, updateStore, showToast }) {
  const [swapFrom, setSwapFrom] = useState('BRL')
  const [swapTo,   setSwapTo]   = useState('BTC')
  const [swapAmt,  setSwapAmt]  = useState('')
  const [loading,  setLoading]  = useState(false)
  const [quotePair, setQuotePair] = useState('BTC/BRL')
  const [quoteResult, setQuoteResult] = useState(null)

  const totalCryptoBRL = (store?.crypto || []).reduce((a, c) => a + c.amount * c.price * 100, 0)

  const swapPreview = () => {
    if (!swapAmt) return null
    const amt = Number(swapAmt)
    if (swapFrom === 'BRL') {
      const coin = store?.crypto?.find(c => c.symbol === swapTo)
      if (!coin) return null
      const spread = store?.pricingRate?.spread || 0.012
      return { receive: (amt / coin.price * (1 - spread)).toFixed(6), unit: swapTo, fee: fmtBRL(amt * spread * 100) }
    } else {
      const coin = store?.crypto?.find(c => c.symbol === swapFrom)
      if (!coin) return null
      const spread = store?.pricingRate?.spread || 0.012
      return { receive: fmtBRL(amt * coin.price * (1 - spread) * 100), unit: 'BRL', fee: fmtBRL(amt * coin.price * spread * 100) }
    }
  }

  const preview = swapPreview()

  const doSwap = async (e) => {
    e.preventDefault()
    setLoading(true)
    await new Promise(r => setTimeout(r, 700))
    const { store: s, message } = runSimulation(store, 'crypto_swap', { from: swapFrom, to: swapTo, amount: swapAmt })
    updateStore(s)
    showToast(message, 'success')
    setSwapAmt('')
    setLoading(false)
  }

  const doQuote = async () => {
    await new Promise(r => setTimeout(r, 400))
    const { message, details } = runSimulation(store, 'pricing_quote', { pair: quotePair })
    setQuoteResult(details)
    showToast(message, 'success')
  }

  const doLiquidate = async () => {
    setLoading(true)
    await new Promise(r => setTimeout(r, 700))
    const { store: s, message } = runSimulation(store, 'crypto_liquidate', {})
    updateStore(s)
    showToast(message, 'success')
    setLoading(false)
  }

  return (
    <div style={{ display:'flex', flexDirection:'column', gap:20, animation:'slideUp 0.4s ease' }}>
      <div>
        <h2 style={{ fontFamily:'Syne', fontWeight:700, fontSize:22, marginBottom:4 }}>Investir</h2>
        <p style={{ fontSize:13, color:'var(--t2)' }}>Cripto, Swap e Pricing Engine</p>
      </div>

      {/* Crypto Cards */}
      <div>
        <p style={{ fontSize:11, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.1em', marginBottom:12 }}>Portfólio Cripto</p>
        <div style={{ display:'grid', gridTemplateColumns:'repeat(3, 1fr)', gap:14 }}>
          {(store?.crypto || []).map(coin => {
            const cfg   = COIN_COLORS[coin.symbol] || { color:'var(--t2)', bg:'rgba(255,255,255,0.06)', border:'var(--border)', letter:'?' }
            const brl   = coin.amount * coin.price * 100
            const pct   = totalCryptoBRL ? brl / totalCryptoBRL * 100 : 0
            const isPos = coin.change24h >= 0

            return (
              <div key={coin.symbol} className="card hover-glow" style={{ padding:20, background:`linear-gradient(135deg, ${cfg.bg}, var(--card))`, border:`1px solid ${cfg.border}` }}>
                <div style={{ display:'flex', justifyContent:'space-between', alignItems:'flex-start', marginBottom:14 }}>
                  <div style={{ width:42, height:42, borderRadius:'50%', background:cfg.bg, border:`2px solid ${cfg.color}33`, display:'flex', alignItems:'center', justifyContent:'center', fontSize:18, color:cfg.color, fontWeight:900 }}>
                    {cfg.letter}
                  </div>
                  <span className={`badge ${isPos ? 'badge-green' : 'badge-red'}`}>
                    {isPos ? '+' : ''}{coin.change24h.toFixed(1)}% 24h
                  </span>
                </div>
                <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:15, marginBottom:2 }}>{coin.name}</p>
                <p style={{ fontSize:12, color:'var(--t3)', marginBottom:12 }}>{coin.symbol}</p>
                <p style={{ fontFamily:'Syne', fontWeight:800, fontSize:20, color:cfg.color, marginBottom:4 }}>
                  {coin.amount.toLocaleString('pt-BR', { minimumFractionDigits:4, maximumFractionDigits:4 })}
                </p>
                <p style={{ fontSize:13, color:'var(--t2)', marginBottom:12 }}>≈ {fmtBRL(brl)}</p>
                <div className="progress-track">
                  <div className="progress-fill" style={{ width:`${pct}%`, background:cfg.color }} />
                </div>
                <p style={{ fontSize:10, color:'var(--t3)', marginTop:4 }}>{pct.toFixed(1)}% do portfólio cripto</p>
                <p style={{ fontSize:11, color:'var(--t3)', marginTop:6 }}>R$ {coin.price.toLocaleString('pt-BR', {minimumFractionDigits:2})} / {coin.symbol}</p>
              </div>
            )
          })}
        </div>
      </div>

      {/* Swap + Quote */}
      <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:16 }}>

        {/* Swap Widget */}
        <div className="card" style={{ padding:22 }}>
          <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:20 }}>
            <div style={{ width:32, height:32, borderRadius:8, background:'rgba(0,232,122,0.1)', border:'1px solid rgba(0,232,122,0.25)', display:'flex', alignItems:'center', justifyContent:'center' }}>
              <ArrowLeftRight size={15} color='var(--green)' />
            </div>
            <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:15 }}>Swap Fiat ↔ Cripto</p>
          </div>

          <form onSubmit={doSwap} style={{ display:'flex', flexDirection:'column', gap:12 }}>
            {/* From/To row */}
            <div style={{ display:'grid', gridTemplateColumns:'1fr auto 1fr', gap:8, alignItems:'center' }}>
              <div>
                <label style={{ fontSize:11, color:'var(--t2)', marginBottom:5, display:'block' }}>De</label>
                <select className="input" value={swapFrom} onChange={e => { setSwapFrom(e.target.value); setSwapTo(e.target.value === 'BRL' ? 'BTC' : 'BRL') }}>
                  <option value="BRL">BRL (Real)</option>
                  {(store?.crypto||[]).map(c => <option key={c.symbol} value={c.symbol}>{c.symbol}</option>)}
                </select>
              </div>
              <button type="button" onClick={() => { const t = swapFrom; setSwapFrom(swapTo === 'BRL' ? 'BRL' : swapTo); setSwapTo(t === 'BRL' ? store?.crypto?.[0]?.symbol || 'BTC' : 'BRL') }}
                style={{ marginTop:18, width:32, height:32, borderRadius:'50%', background:'rgba(255,255,255,0.04)', border:'1px solid var(--border)', display:'flex', alignItems:'center', justifyContent:'center', cursor:'pointer' }}>
                <RefreshCw size={13} color='var(--t2)' />
              </button>
              <div>
                <label style={{ fontSize:11, color:'var(--t2)', marginBottom:5, display:'block' }}>Para</label>
                <select className="input" value={swapTo} onChange={e => setSwapTo(e.target.value)}>
                  {swapFrom === 'BRL'
                    ? (store?.crypto||[]).map(c => <option key={c.symbol} value={c.symbol}>{c.symbol}</option>)
                    : <option value="BRL">BRL (Real)</option>
                  }
                </select>
              </div>
            </div>

            {/* Amount */}
            <div>
              <label style={{ fontSize:11, color:'var(--t2)', marginBottom:5, display:'block' }}>Valor ({swapFrom})</label>
              <input className="input" type="number" placeholder="100" step="0.01" min="0.01" required
                value={swapAmt} onChange={e => setSwapAmt(e.target.value)} />
            </div>

            {/* Preview */}
            {preview && (
              <div style={{ background:'rgba(0,232,122,0.04)', border:'1px solid rgba(0,232,122,0.15)', borderRadius:10, padding:'12px 14px' }}>
                <div style={{ display:'flex', justifyContent:'space-between', fontSize:12, marginBottom:4 }}>
                  <span style={{ color:'var(--t3)' }}>Você receberá:</span>
                  <span style={{ fontWeight:700, color:'var(--green)', fontFamily:'Syne' }}>{preview.receive} {preview.unit}</span>
                </div>
                <div style={{ display:'flex', justifyContent:'space-between', fontSize:11 }}>
                  <span style={{ color:'var(--t3)' }}>Taxa (spread 1.2%):</span>
                  <span style={{ color:'var(--t3)' }}>{preview.fee}</span>
                </div>
              </div>
            )}

            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Executando...' : `Swap ${swapFrom} → ${swapTo}`}
            </button>
          </form>
        </div>

        {/* Pricing Engine */}
        <div className="card" style={{ padding:22 }}>
          <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:20 }}>
            <div style={{ width:32, height:32, borderRadius:8, background:'rgba(59,130,246,0.1)', border:'1px solid rgba(59,130,246,0.25)', display:'flex', alignItems:'center', justifyContent:'center' }}>
              <BarChart2 size={15} color='var(--blue)' />
            </div>
            <p style={{ fontFamily:'Syne', fontWeight:700, fontSize:15 }}>Pricing Engine</p>
          </div>

          <div style={{ display:'flex', flexDirection:'column', gap:12 }}>
            <div>
              <label style={{ fontSize:11, color:'var(--t2)', marginBottom:5, display:'block' }}>Par de cotação</label>
              <select className="input" value={quotePair} onChange={e => setQuotePair(e.target.value)}>
                <option value="BTC/BRL">BTC / BRL</option>
                <option value="ETH/BRL">ETH / BRL</option>
                <option value="USDT/BRL">USDT / BRL</option>
              </select>
            </div>
            <button onClick={doQuote} className="btn-ghost" style={{ display:'flex', alignItems:'center', justifyContent:'center', gap:8 }}>
              <Zap size={14} /> Obter Cotação
            </button>

            {quoteResult && (
              <div style={{ background:'var(--card2)', border:'1px solid var(--border)', borderRadius:10, padding:14 }}>
                <p style={{ fontSize:11, color:'var(--t3)', marginBottom:8, fontWeight:700 }}>RESULTADO — {quotePair}</p>
                {Object.entries(quoteResult).map(([k,v]) => (
                  <div key={k} style={{ display:'flex', justifyContent:'space-between', fontSize:12, marginBottom:5 }}>
                    <span style={{ color:'var(--t3)', textTransform:'capitalize' }}>{k}:</span>
                    <span style={{ fontWeight:600, color:'var(--t1)' }}>{String(v)}</span>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="divider" style={{ margin:'18px 0' }} />

          {/* Liquidate */}
          <div>
            <p style={{ fontSize:12, color:'var(--t2)', marginBottom:10, fontWeight:600 }}>Liquidação Parcial</p>
            <p style={{ fontSize:11, color:'var(--t3)', marginBottom:12, lineHeight:1.5 }}>
              Converte 50% da maior posição cripto em BRL.
            </p>
            <button onClick={doLiquidate} disabled={loading} className="btn-danger" style={{ width:'100%' }}>
              {loading ? 'Processando...' : 'Liquidar 50%'}
            </button>
          </div>
        </div>
      </div>

      {/* Total Summary */}
      <div className="card card-accent" style={{ padding:20 }}>
        <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between' }}>
          <div>
            <p style={{ fontSize:11, color:'var(--t3)', textTransform:'uppercase', letterSpacing:'0.08em', marginBottom:6 }}>Total em Cripto</p>
            <p style={{ fontFamily:'Syne', fontWeight:800, fontSize:24, color:'var(--green)' }}>{fmtBRL(totalCryptoBRL)}</p>
          </div>
          <div style={{ textAlign:'right' }}>
            <p style={{ fontSize:11, color:'var(--t3)', marginBottom:6 }}>Performance 24h</p>
            <div style={{ display:'flex', alignItems:'center', gap:6 }}>
              <TrendingUp size={14} color='var(--green)' />
              <span style={{ fontFamily:'Syne', fontWeight:700, fontSize:16, color:'var(--green)' }}>+{((store?.crypto?.[0]?.change24h || 3.2) * 1.1).toFixed(1)}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
