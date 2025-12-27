import { useEffect, useRef, useCallback, useState } from 'react'
import { toast } from 'sonner'
import { useSessionStore } from '@/stores/session.store'

interface NotificationEvent {
  id: string
  type: 'medication_due' | 'vaccination_due' | 'appointment_soon' | 'sleep_insight'
  title: string
  message: string
  childId?: string
  childName?: string
  timestamp: string
}

interface UseNotificationsOptions {
  enabled?: boolean
  onNotification?: (event: NotificationEvent) => void
}

export function useNotifications(options: UseNotificationsOptions = {}) {
  const { enabled = true, onNotification } = options
  const { token } = useSessionStore()
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const connectRef = useRef<() => void>(() => {})
  const [connected, setConnected] = useState(false)

  const connect = useCallback(() => {
    if (!token || !enabled) return

    // Close existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    // Create SSE connection with auth token as query param
    // (EventSource doesn't support custom headers)
    const url = `/api/notifications/stream?token=${encodeURIComponent(token)}`
    const eventSource = new EventSource(url)

    eventSource.onopen = () => {
      console.log('[SSE] Connected to notification stream')
      setConnected(true)
    }

    eventSource.onerror = (error) => {
      console.error('[SSE] Connection error:', error)
      setConnected(false)
      eventSource.close()

      // Reconnect after 5 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        console.log('[SSE] Attempting to reconnect...')
        connectRef.current()
      }, 5000)
    }

    eventSource.addEventListener('connected', () => {
      console.log('[SSE] Received connected event')
    })

    eventSource.addEventListener('notification', (event) => {
      try {
        const data: NotificationEvent = JSON.parse(event.data)
        console.log('[SSE] Received notification:', data)

        // Show toast based on notification type
        const toastOptions = {
          id: data.id,
          duration: 10000,
        }

        switch (data.type) {
          case 'medication_due':
            toast.warning(data.message, {
              ...toastOptions,
              description: data.childName ? `For ${data.childName}` : undefined,
            })
            break
          case 'vaccination_due':
            toast.info(data.message, {
              ...toastOptions,
              description: data.childName ? `For ${data.childName}` : undefined,
            })
            break
          case 'appointment_soon':
            toast.info(data.message, {
              ...toastOptions,
              description: data.childName ? `For ${data.childName}` : undefined,
            })
            break
          case 'sleep_insight':
            toast.success(data.message, {
              ...toastOptions,
              description: data.childName ? `For ${data.childName}` : undefined,
            })
            break
          default:
            toast(data.title, {
              ...toastOptions,
              description: data.message,
            })
        }

        // Call custom handler if provided
        onNotification?.(data)
      } catch (err) {
        console.error('[SSE] Failed to parse notification:', err)
      }
    })

    eventSourceRef.current = eventSource
  }, [token, enabled, onNotification])

  // Keep ref updated for reconnection
  useEffect(() => {
    connectRef.current = connect
  }, [connect])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    setConnected(false)
  }, [])

  useEffect(() => {
    // Check if notifications are enabled in localStorage
    const notificationsEnabled = localStorage.getItem('notifications') === 'true'
    if (!notificationsEnabled || !enabled) {
      disconnect()
      return
    }

    connect()

    return () => {
      disconnect()
    }
  }, [connect, disconnect, enabled])

  // Listen for notification preference changes
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'notifications') {
        if (e.newValue === 'true') {
          connect()
        } else {
          disconnect()
        }
      }
    }

    window.addEventListener('storage', handleStorageChange)
    return () => window.removeEventListener('storage', handleStorageChange)
  }, [connect, disconnect])

  return { connected, disconnect, reconnect: connect }
}
