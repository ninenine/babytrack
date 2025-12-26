import { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSessionStore } from '@/stores/session.store'
import { apiClient } from '@/api/client'

export function useAuth() {
  const navigate = useNavigate()
  const { user, token, isAuthenticated, setSession, clearSession } = useSessionStore()

  const login = useCallback(() => {
    window.location.href = '/api/auth/google'
  }, [])

  const logout = useCallback(() => {
    clearSession()
    navigate('/login')
  }, [clearSession, navigate])

  const refreshToken = useCallback(async () => {
    if (!token) return false

    try {
      const response = await apiClient.post<{
        user: {
          id: string
          email: string
          name: string
          avatar_url?: string
        }
        token: string
      }>('/api/auth/refresh')

      setSession(
        {
          id: response.data.user.id,
          email: response.data.user.email,
          name: response.data.user.name,
          avatarUrl: response.data.user.avatar_url,
        },
        response.data.token
      )

      return true
    } catch {
      clearSession()
      return false
    }
  }, [token, setSession, clearSession])

  const checkAuth = useCallback(async () => {
    if (!token) {
      return false
    }

    try {
      const response = await apiClient.get<{
        id: string
        email: string
        name: string
        avatar_url?: string
      }>('/api/auth/me')

      setSession(
        {
          id: response.data.id,
          email: response.data.email,
          name: response.data.name,
          avatarUrl: response.data.avatar_url,
        },
        token
      )

      return true
    } catch {
      clearSession()
      return false
    }
  }, [token, setSession, clearSession])

  return {
    user,
    token,
    isAuthenticated,
    login,
    logout,
    refreshToken,
    checkAuth,
  }
}
