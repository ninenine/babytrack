import { useSessionStore } from '@/stores/session.store'

interface RequestConfig {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  headers?: Record<string, string>
  body?: unknown
  params?: Record<string, string | undefined>
}

interface ApiResponse<T = unknown> {
  data: T
  status: number
}

class ApiClient {
  private baseUrl: string

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

  async request<T>(path: string, config: RequestConfig = {}): Promise<ApiResponse<T>> {
    const { method = 'GET', headers = {}, body, params } = config

    const url = this.buildUrl(`${this.baseUrl}${path}`, params)

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...this.getAuthHeader(),
        ...headers,
      },
      body: body ? JSON.stringify(body) : undefined,
    })

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
