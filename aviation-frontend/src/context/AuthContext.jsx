import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { getMe } from '../api/client.js'

const AuthContext = createContext(null)

function parseJWT(token) {
  try {
    const base64Url = token.split('.')[1]
    if (!base64Url) return null
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(base64))
  } catch {
    return null
  }
}

function extractUser(payload) {
  if (!payload) return null
  return {
    id: payload.user_id ?? payload.id ?? payload.sub ?? null,
    username: payload.username ?? payload.name ?? payload.email ?? 'Пользователь',
    role: payload.role ?? null,
  }
}

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => localStorage.getItem('sky_token'))
  const [user, setUser] = useState(() => {
    const raw = localStorage.getItem('sky_user')
    try { return raw ? JSON.parse(raw) : null } catch { return null }
  })
  const [passenger, setPassenger] = useState(null)

  const fetchMe = useCallback(async () => {
    try {
      const res = await getMe()
      setPassenger(res.data?.passenger ?? null)
      return res.data
    } catch {
      setPassenger(null)
      return null
    }
  }, [])

  useEffect(() => {
    if (token) {
      fetchMe()
    } else {
      setPassenger(null)
    }
  }, [token, fetchMe])

  const login = (newToken) => {
    const payload = parseJWT(newToken)
    const userInfo = extractUser(payload)
    localStorage.setItem('sky_token', newToken)
    if (userInfo) {
      localStorage.setItem('sky_user', JSON.stringify(userInfo))
    }
    setToken(newToken)
    setUser(userInfo)
  }

  const logout = () => {
    localStorage.removeItem('sky_token')
    localStorage.removeItem('sky_user')
    setToken(null)
    setUser(null)
    setPassenger(null)
  }

  const hasPassenger = !!passenger

  const value = {
    token,
    user,
    passenger,
    hasPassenger,
    fetchMe,
    setPassenger,
    login,
    logout,
    isAuthenticated: !!token,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
