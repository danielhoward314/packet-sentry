import { createContext, useContext, useEffect, useState } from 'react'
import { validateJwt } from '@/lib/validateJwt'

interface AuthContextType {
  isAuthenticated: boolean
  loading: boolean
  setIsAuthenticated: (val: boolean) => void
  recheckAuth: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [loading, setLoading] = useState(true)

  const checkAuth = async () => {
    const token = localStorage.getItem('jwt')
    if (!token) {
      setIsAuthenticated(false)
      setLoading(false)
      return
    }

    try {
      const valid = await validateJwt(token)
      setIsAuthenticated(valid)
    } catch (e) {
      console.error('JWT validation failed:', e)
      setIsAuthenticated(false)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    checkAuth()
  }, [])

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        loading,
        setIsAuthenticated,
        recheckAuth: checkAuth,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used within AuthProvider')
  return context
}
