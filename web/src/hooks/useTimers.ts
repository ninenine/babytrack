import { useState, useEffect, useCallback, useRef } from 'react'

interface TimerState {
  elapsed: number
  isRunning: boolean
}

export function useTimer(startTime?: Date) {
  const [state, setState] = useState<TimerState>({
    elapsed: startTime ? Date.now() - startTime.getTime() : 0,
    isRunning: !!startTime,
  })

  const intervalRef = useRef<number | null>(null)

  const start = useCallback(() => {
    if (intervalRef.current) return

    const startedAt = Date.now()
    setState((prev) => ({ ...prev, isRunning: true }))

    intervalRef.current = window.setInterval(() => {
      setState((prev) => ({
        ...prev,
        elapsed: Date.now() - startedAt + prev.elapsed,
      }))
    }, 1000)
  }, [])

  const stop = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
    setState((prev) => ({ ...prev, isRunning: false }))
  }, [])

  const reset = useCallback(() => {
    stop()
    setState({ elapsed: 0, isRunning: false })
  }, [stop])

  useEffect(() => {
    if (startTime) {
      start()
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [startTime, start])

  const formatElapsed = useCallback((ms: number) => {
    const seconds = Math.floor(ms / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)

    const pad = (n: number) => n.toString().padStart(2, '0')

    if (hours > 0) {
      return `${hours}:${pad(minutes % 60)}:${pad(seconds % 60)}`
    }
    return `${pad(minutes)}:${pad(seconds % 60)}`
  }, [])

  return {
    elapsed: state.elapsed,
    isRunning: state.isRunning,
    formattedElapsed: formatElapsed(state.elapsed),
    start,
    stop,
    reset,
  }
}

export function useCountdown(targetDate: Date) {
  const [remaining, setRemaining] = useState(
    Math.max(0, targetDate.getTime() - Date.now())
  )

  useEffect(() => {
    const interval = setInterval(() => {
      const newRemaining = Math.max(0, targetDate.getTime() - Date.now())
      setRemaining(newRemaining)

      if (newRemaining === 0) {
        clearInterval(interval)
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [targetDate])

  const formatRemaining = useCallback((ms: number) => {
    const seconds = Math.floor(ms / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)

    if (days > 0) {
      return `${days}d ${hours % 24}h`
    }
    if (hours > 0) {
      return `${hours}h ${minutes % 60}m`
    }
    return `${minutes}m ${seconds % 60}s`
  }, [])

  return {
    remaining,
    formattedRemaining: formatRemaining(remaining),
    isExpired: remaining === 0,
  }
}
