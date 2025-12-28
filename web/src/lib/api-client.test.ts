import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/test/mocks/server'
import { apiClient } from './api-client'
import { useSessionStore } from '@/stores/session.store'

// Store original location
const originalLocation = window.location

describe('ApiClient', () => {
  beforeEach(() => {
    // Mock window.location with proper origin
    Object.defineProperty(window, 'location', {
      value: {
        origin: 'http://localhost:3000',
        href: '',
      },
      writable: true,
    })
    useSessionStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
    })
  })

  afterEach(() => {
    // Restore original location
    Object.defineProperty(window, 'location', {
      value: originalLocation,
      writable: true,
    })
  })

  describe('get', () => {
    it('should make GET request', async () => {
      server.use(
        http.get('http://localhost:3000/api/test', () => {
          return HttpResponse.json({ data: 'test' })
        })
      )

      const result = await apiClient.get('/api/test')

      expect(result.data).toEqual({ data: 'test' })
      expect(result.status).toBe(200)
    })

    it('should include auth header when token exists', async () => {
      useSessionStore.setState({
        user: { id: '1', email: 'test@test.com', name: 'Test' },
        token: 'test-token',
        isAuthenticated: true,
      })

      let capturedAuthHeader: string | null = null

      server.use(
        http.get('http://localhost:3000/api/test', ({ request }) => {
          capturedAuthHeader = request.headers.get('Authorization')
          return HttpResponse.json({})
        })
      )

      await apiClient.get('/api/test')

      expect(capturedAuthHeader).toBe('Bearer test-token')
    })

    it('should append query params', async () => {
      let capturedUrl = ''

      server.use(
        http.get('http://localhost:3000/api/test', ({ request }) => {
          capturedUrl = request.url
          return HttpResponse.json({})
        })
      )

      await apiClient.get('/api/test', { params: { foo: 'bar', baz: 'qux' } })

      expect(capturedUrl).toContain('foo=bar')
      expect(capturedUrl).toContain('baz=qux')
    })

    it('should skip undefined params', async () => {
      let capturedUrl = ''

      server.use(
        http.get('http://localhost:3000/api/test', ({ request }) => {
          capturedUrl = request.url
          return HttpResponse.json({})
        })
      )

      await apiClient.get('/api/test', { params: { foo: 'bar', baz: undefined } })

      expect(capturedUrl).toContain('foo=bar')
      expect(capturedUrl).not.toContain('baz')
    })
  })

  describe('post', () => {
    it('should make POST request with body', async () => {
      let capturedBody: unknown = null

      server.use(
        http.post('http://localhost:3000/api/test', async ({ request }) => {
          capturedBody = await request.json()
          return HttpResponse.json({ id: '123' }, { status: 201 })
        })
      )

      const body = { name: 'Test', value: 42 }
      const result = await apiClient.post('/api/test', body)

      expect(capturedBody).toEqual(body)
      expect(result.status).toBe(201)
    })
  })

  describe('put', () => {
    it('should make PUT request', async () => {
      server.use(
        http.put('http://localhost:3000/api/test/1', () => {
          return HttpResponse.json({ updated: true })
        })
      )

      const result = await apiClient.put('/api/test/1', { name: 'Updated' })

      expect(result.data).toEqual({ updated: true })
    })
  })

  describe('patch', () => {
    it('should make PATCH request', async () => {
      server.use(
        http.patch('http://localhost:3000/api/test/1', () => {
          return HttpResponse.json({ patched: true })
        })
      )

      const result = await apiClient.patch('/api/test/1', { field: 'value' })

      expect(result.data).toEqual({ patched: true })
    })
  })

  describe('delete', () => {
    it('should make DELETE request', async () => {
      server.use(
        http.delete('http://localhost:3000/api/test/1', () => {
          return new HttpResponse(null, { status: 204 })
        })
      )

      const result = await apiClient.delete('/api/test/1')

      expect(result.status).toBe(204)
    })
  })

  describe('error handling', () => {
    it('should throw error on non-ok response', async () => {
      server.use(
        http.get('http://localhost:3000/api/test', () => {
          return HttpResponse.json({ error: 'Bad request' }, { status: 400 })
        })
      )

      await expect(apiClient.get('/api/test')).rejects.toThrow('Bad request')
    })

    it('should handle error without message', async () => {
      server.use(
        http.get('http://localhost:3000/api/test', () => {
          return HttpResponse.json({}, { status: 500 })
        })
      )

      await expect(apiClient.get('/api/test')).rejects.toThrow('Request failed: 500')
    })
  })

  describe('token refresh', () => {
    it('should attempt refresh on 401 and update session', async () => {
      useSessionStore.setState({
        user: { id: '1', email: 'test@test.com', name: 'Test' },
        token: 'old-token',
        isAuthenticated: true,
      })

      let refreshCalled = false
      let retryAuthHeader: string | null = null

      server.use(
        http.get('http://localhost:3000/api/test', ({ request }) => {
          const auth = request.headers.get('Authorization')
          if (auth === 'Bearer old-token') {
            return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 })
          }
          retryAuthHeader = auth
          return HttpResponse.json({ data: 'success' })
        }),
        http.post('http://localhost:3000/api/auth/refresh', () => {
          refreshCalled = true
          return HttpResponse.json({
            token: 'new-token',
            user: { id: '1', email: 'test@test.com', name: 'Test' },
          })
        })
      )

      const result = await apiClient.get('/api/test')

      expect(refreshCalled).toBe(true)
      expect(retryAuthHeader).toBe('Bearer new-token')
      expect(result.data).toEqual({ data: 'success' })
    })

    it('should redirect to login on refresh failure', async () => {
      useSessionStore.setState({
        user: { id: '1', email: 'test@test.com', name: 'Test' },
        token: 'old-token',
        isAuthenticated: true,
      })

      server.use(
        http.get('http://localhost:3000/api/test', () => {
          return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 })
        }),
        http.post('http://localhost:3000/api/auth/refresh', () => {
          return HttpResponse.json({ error: 'Invalid token' }, { status: 401 })
        })
      )

      await expect(apiClient.get('/api/test')).rejects.toThrow(
        'Session expired. Please log in again.'
      )
      expect(window.location.href).toBe('/login')
    })

    it('should clear session on refresh network error', async () => {
      useSessionStore.setState({
        user: { id: '1', email: 'test@test.com', name: 'Test' },
        token: 'old-token',
        isAuthenticated: true,
      })

      server.use(
        http.get('http://localhost:3000/api/test', () => {
          return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 })
        }),
        http.post('http://localhost:3000/api/auth/refresh', () => {
          return HttpResponse.error()
        })
      )

      await expect(apiClient.get('/api/test')).rejects.toThrow(
        'Session expired. Please log in again.'
      )
      expect(window.location.href).toBe('/login')
      expect(useSessionStore.getState().isAuthenticated).toBe(false)
    })
  })

  describe('204 No Content', () => {
    it('should handle 204 response', async () => {
      server.use(
        http.delete('http://localhost:3000/api/test/1', () => {
          return new HttpResponse(null, { status: 204 })
        })
      )

      const result = await apiClient.delete('/api/test/1')

      expect(result.status).toBe(204)
      expect(result.data).toEqual({})
    })
  })
})
