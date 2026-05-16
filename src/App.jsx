import React, { useState, useCallback, useEffect, createContext, useContext } from 'react'
import { buildStore, simulate } from './data/mock.js'
import Login from './pages/Login.jsx'
import Layout from './components/Layout.jsx'
import Home from './pages/Home.jsx'
import Mover from './pages/Mover.jsx'
import Cartoes from './pages/Cartoes.jsx'
import Investir from './pages/Investir.jsx'
import Credito from './pages/Credito.jsx'
import ZPass from './pages/ZPass.jsx'
import Mais from './pages/Mais.jsx'
import Toast from './components/Toast.jsx'
import ZionPanel from './components/ZionPanel.jsx'
import WelcomeScreen from './components/WelcomeScreen.jsx'

export const AppCtx = createContext(null)
export const useApp = () => useContext(AppCtx)

const PAGES = { home: Home, mover: Mover, cartoes: Cartoes, investir: Investir, credito: Credito, zpass: ZPass, mais: Mais }

export default function App() {
  const [authed, setAuthed] = useState(() => sessionStorage.getItem('zf_auth') === '1')
  const [store, setStore] = useState(() => buildStore())
  const [page, setPage] = useState('home')
  const [toasts, setToasts] = useState([])
  const [zionOpen, setZionOpen] = useState(false)
  const [modal, setModal] = useState(null)
  const [showWelcome, setShowWelcome] = useState(false)

  const toast = useCallback((msg, type = 'success') => {
    const id = Date.now()
    setToasts(t => [...t, { id, msg, type }])
    setTimeout(() => setToasts(t => t.filter(x => x.id !== id)), 4000)
  }, [])

  const dispatch = useCallback((action, data) => {
    setStore(prev => {
      try {
        return simulate(prev, action, data)
      } catch (e) {
        toast(e.message, 'error')
        return prev
      }
    })
  }, [toast])

  const login = useCallback((pin) => {
    if (pin === '1234' || pin.length >= 4) {
      sessionStorage.setItem('zf_auth', '1')
      setAuthed(true)
      setStore(buildStore())
      if (!sessionStorage.getItem('zf_welcomed')) {
        sessionStorage.setItem('zf_welcomed', '1')
        setShowWelcome(true)
      }
    } else {
      toast('PIN incorreto', 'error')
    }
  }, [toast])

  const logout = useCallback(() => {
    sessionStorage.removeItem('zf_auth')
    setAuthed(false)
    setPage('home')
  }, [])

  const PageComponent = PAGES[page] || Home

  if (!authed) return (
    <AppCtx.Provider value={{ store, dispatch, toast, modal, setModal, zionOpen, setZionOpen }}>
      <Login onLogin={login} />
      <Toast toasts={toasts} />
    </AppCtx.Provider>
  )

  return (
    <AppCtx.Provider value={{ store, dispatch, toast, modal, setModal, zionOpen, setZionOpen, logout }}>
      {showWelcome && <WelcomeScreen onDone={() => setShowWelcome(false)} />}
      <Layout page={page} setPage={setPage}>
        <PageComponent />
      </Layout>
      <ZionPanel />
      <Toast toasts={toasts} />
    </AppCtx.Provider>
  )
}
