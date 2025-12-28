import { describe, it, expect, beforeEach } from 'vitest'
import { useSessionStore } from './session.store'

describe('useSessionStore', () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    useSessionStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
    })
  })

  describe('initial state', () => {
    it('should start with no user and not authenticated', () => {
      const state = useSessionStore.getState()
      expect(state.user).toBeNull()
      expect(state.token).toBeNull()
      expect(state.isAuthenticated).toBe(false)
    })
  })

  describe('setSession', () => {
    it('should set user and token', () => {
      const user = {
        id: 'user-1',
        email: 'test@example.com',
        name: 'Test User',
        avatarUrl: 'https://example.com/avatar.jpg',
      }
      const token = 'test-token-123'

      useSessionStore.getState().setSession(user, token)

      const state = useSessionStore.getState()
      expect(state.user).toEqual(user)
      expect(state.token).toBe(token)
      expect(state.isAuthenticated).toBe(true)
    })

    it('should set user without avatar', () => {
      const user = {
        id: 'user-2',
        email: 'noavatar@example.com',
        name: 'No Avatar User',
      }
      const token = 'token-456'

      useSessionStore.getState().setSession(user, token)

      const state = useSessionStore.getState()
      expect(state.user).toEqual(user)
      expect(state.user?.avatarUrl).toBeUndefined()
    })
  })

  describe('clearSession', () => {
    it('should clear user and token', () => {
      // First set a session
      const user = { id: 'user-1', email: 'test@example.com', name: 'Test' }
      useSessionStore.getState().setSession(user, 'token')
      expect(useSessionStore.getState().isAuthenticated).toBe(true)

      // Then clear it
      useSessionStore.getState().clearSession()

      const state = useSessionStore.getState()
      expect(state.user).toBeNull()
      expect(state.token).toBeNull()
      expect(state.isAuthenticated).toBe(false)
    })
  })
})
