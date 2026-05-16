import { useState, useCallback, useEffect } from 'react'
import { buildMockStore } from './data/mock.js'
import Login from './pages/Login.jsx'
import Layout from './components/Layout.jsx'
import Toast from './components/Toast.jsx'

const SESSION_KEY = 'zf_session'

export default function App() {
  const [auth,     setAuth]     = useState(null)   // null | { token, demo }
  const [page,     setPage]     = useState('dashboard')
  const [store,    setStore]    = useState(null)
  const [toast,    setToast]    = useState(null)
  const [showZion, setShowZion] = useState(false)
  const [simModal, setSimModal] = useState(null)   // { action, title, fields[] }

  // restore session
  useEffect(() => {
    try {
      const saved = JSON.parse(sessionStorage.getItem(SESSION_KEY) || 'null')
      if (saved?.auth) {
        setAuth(saved.auth)
        setStore(saved.store || buildMockStore())
      }
    } catch {}
  }, [])

  const showToast = useCallback((message, type = 'success') => {
    const id = Date.now()
    setToast({ id, message, type })
    setTimeout(() => setToast(t => t?.id === id ? null : t), 3800)
  }, [])

  const loginDemo = useCallback(() => {
    const s = buildMockStore()
    setStore(s)
    setAuth({ demo: true, token: 'demo' })
    sessionStorage.setItem(SESSION_KEY, JSON.stringify({ auth: { demo: true, token: 'demo' }, store: s }))
    showToast('Modo demo ativado — bem-vindo, Rafael!', 'success')
  }, [showToast])

  const loginLive = useCallback(async (email, password, apiBase) => {
    try {
      const resp = await fetch(`${apiBase.replace(/\/$/, '')}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })
      if (!resp.ok) throw new Error((await resp.json()).error || 'Credenciais inválidas')
      const data = await resp.json()
      const s = buildMockStore()
      s.user.email = email
      setStore(s)
      setAuth({ demo: false, token: data.access_token, apiBase })
      sessionStorage.setItem(SESSION_KEY, JSON.stringify({ auth: { demo: false, token: data.access_token, apiBase }, store: s }))
      showToast('Conectado com sucesso!', 'success')
    } catch (e) {
      throw e
    }
  }, [showToast])

  const logout = useCallback(() => {
    setAuth(null)
    setStore(null)
    setPage('dashboard')
    setShowZion(false)
    sessionStorage.removeItem(SESSION_KEY)
    showToast('Sessão encerrada', 'info')
  }, [showToast])

  const updateStore = useCallback((newStore) => {
    setStore(newStore)
    sessionStorage.setItem(SESSION_KEY, JSON.stringify({
      auth,
      store: newStore,
    }))
  }, [auth])

  if (!auth) {
    return (
      <>
        <Login onDemo={loginDemo} onLive={loginLive} />
        {toast && <Toast {...toast} />}
      </>
    )
  }

  return (
    <>
      <Layout
        page={page}
        setPage={setPage}
        store={store}
        updateStore={updateStore}
        showZion={showZion}
        setShowZion={setShowZion}
        simModal={simModal}
        setSimModal={setSimModal}
        logout={logout}
        showToast={showToast}
        isDemo={auth.demo}
      />
      {toast && <Toast {...toast} />}
    </>
  )
}
