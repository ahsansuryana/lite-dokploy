import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'

interface AuthContextType {
  token: string | null
  username: string | null
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  isAuthenticated: boolean
  isLoading: boolean
}

const AuthContext = createContext<AuthContextType>({
  token: null,
  username: null,
  login: async () => {},
  logout: () => {},
  isAuthenticated: false,
  isLoading: true,
})

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [username, setUsername] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    if (token) {
      fetch('/api/auth/me', {
        headers: { Authorization: `Bearer ${token}` },
      })
        .then((res) => {
          if (res.ok) return res.json()
          throw new Error('invalid token')
        })
        .then((data) => setUsername(data.username))
        .catch(() => {
          setToken(null)
          setUsername(null)
          localStorage.removeItem('token')
        })
        .finally(() => setIsLoading(false))
    } else {
      setIsLoading(false)
    }
  }, [token])

  const login = async (username: string, password: string) => {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
    if (!res.ok) {
      const text = await res.text()
      throw new Error(text || 'login failed')
    }
    const data = await res.json()
    localStorage.setItem('token', data.token)
    setToken(data.token)
    setUsername(username)
  }

  const logout = () => {
    localStorage.removeItem('token')
    setToken(null)
    setUsername(null)
  }

  return (
    <AuthContext.Provider
      value={{ token, username, login, logout, isAuthenticated: !!token && !!username, isLoading }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
