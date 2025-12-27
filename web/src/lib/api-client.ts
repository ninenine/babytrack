import { useSessionStore } from '@/stores/session.store'
import { API_ENDPOINTS } from './constants'

interface RequestConfig {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  headers?: Record<string, string>
  body?: unknown
  params?: Record<string, string | undefined>
  skipAuth?: boolean
}

interface ApiResponse<T = unknown> {
  data: T
  status: number
}

interface RefreshResponse {
  token: string
  user: {
    id: string
    email: string
    name: string
    avatar_url?: string
  }
}

class ApiClient {
  private baseUrl: string
  private refreshPromise: Promise<string | null> | null = null

  constructor(baseUrl: string = '') {
    this.baseUrl = baseUrl
  }

  private getAuthHeader(): Record<string, string> {
    const token = useSessionStore.getState().token
    return token ? { Authorization: `Bearer ${token}` } : {}
  }

  private buildUrl(path: string, params?: Record<string, string | undefined>): string {
    const url = new URL(path, window.location.origin)

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          url.searchParams.append(key, value)
        }
      })
    }

    return url.toString()
  }

  private async refreshToken(): Promise<string | null> {
    // If already refreshing, wait for that promise
    if (this.refreshPromise) {
      return this.refreshPromise
    }

    this.refreshPromise = this.doRefreshToken()

    try {
      return await this.refreshPromise
    } finally {
      this.refreshPromise = null
    }
  }

  private async doRefreshToken(): Promise<string | null> {
    const currentToken = useSessionStore.getState().token
    if (!currentToken) {
      return null
    }

    try {
      const response = await fetch(API_ENDPOINTS.AUTH.REFRESH, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${currentToken}`,
        },
      })

      if (!response.ok) {
        // Refresh failed, clear session
        useSessionStore.getState().clearSession()
        window.location.href = '/login'
        return null
      }

      const data: RefreshResponse = await response.json()

      // Update stored token and user
      useSessionStore.getState().setSession(
        {
          id: data.user.id,
          email: data.user.email,
          name: data.user.name,
          avatarUrl: data.user.avatar_url,
        },
        data.token
      )

      return data.token
    } catch {
      useSessionStore.getState().clearSession()
      window.location.href = '/login'
      return null
    }
  }

  async request<T>(path: string, config: RequestConfig = {}): Promise<ApiResponse<T>> {
    const { method = 'GET', headers = {}, body, params, skipAuth = false } = config

    const url = this.buildUrl(`${this.baseUrl}${path}`, params)
    const isRefreshEndpoint = path === API_ENDPOINTS.AUTH.REFRESH

    const doRequest = async (authHeaders: Record<string, string>) => {
      return fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...authHeaders,
          ...headers,
        },
        body: body ? JSON.stringify(body) : undefined,
      })
    }

    let response = await doRequest(skipAuth ? {} : this.getAuthHeader())

    // Handle 401 - try to refresh token (but not for the refresh endpoint itself)
    if (response.status === 401 && !isRefreshEndpoint && !skipAuth) {
      const newToken = await this.refreshToken()

      if (newToken) {
        // Retry the original request with the new token
        response = await doRequest({ Authorization: `Bearer ${newToken}` })
      } else {
        // Refresh failed, throw error
        throw new Error('Session expired. Please log in again.')
      }
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(error.error || `Request failed: ${response.status}`)
    }

    // Handle 204 No Content
    if (response.status === 204) {
      return { data: {} as T, status: response.status }
    }

    const data = await response.json()
    return { data, status: response.status }
  }

  async get<T>(path: string, config?: { params?: Record<string, string | undefined> }): Promise<ApiResponse<T>> {
    return this.request<T>(path, { method: 'GET', ...config })
  }

  async post<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(path, { method: 'POST', body })
  }

  async put<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(path, { method: 'PUT', body })
  }

  async patch<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(path, { method: 'PATCH', body })
  }

  async delete<T>(path: string): Promise<ApiResponse<T>> {
    return this.request<T>(path, { method: 'DELETE' })
  }
}

export const apiClient = new ApiClient()
