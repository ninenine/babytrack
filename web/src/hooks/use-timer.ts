import { useState, useEffect, useCallback } from 'react'

export function useTimer(startTime: Date | null) {
  const [elapsed, setElapsed] = useState(0)

  useEffect(() => {
    if (!startTime) {
      setElapsed(0)
      return
    }

    const updateElapsed = () => {
      setElapsed(Date.now() - startTime.getTime())
    }

    updateElapsed()
    const interval = setInterval(updateElapsed, 1000)

    return () => clearInterval(interval)
  }, [startTime])

  const formatElapsed = useCallback((ms: number) => {
    const totalSeconds = Math.floor(ms / 1000)
    const hours = Math.floor(totalSeconds / 3600)
    const minutes = Math.floor((totalSeconds % 3600) / 60)
    const seconds = totalSeconds % 60

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
    }
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }, [])

  return {
    elapsed,
    formattedElapsed: formatElapsed(elapsed),
    hours: Math.floor(elapsed / 3600000),
    minutes: Math.floor((elapsed % 3600000) / 60000),
    seconds: Math.floor((elapsed % 60000) / 1000),
  }
}
