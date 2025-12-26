import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useTimer } from './use-timer'

describe('useTimer', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns 0 elapsed when startTime is null', () => {
    const { result } = renderHook(() => useTimer(null))

    expect(result.current.elapsed).toBe(0)
    expect(result.current.formattedElapsed).toBe('0:00')
    expect(result.current.hours).toBe(0)
    expect(result.current.minutes).toBe(0)
    expect(result.current.seconds).toBe(0)
  })

  it('calculates elapsed time correctly', () => {
    const now = Date.now()
    vi.setSystemTime(now)

    // Start time was 5 minutes ago
    const startTime = new Date(now - 5 * 60 * 1000)
    const { result } = renderHook(() => useTimer(startTime))

    expect(result.current.minutes).toBe(5)
    expect(result.current.seconds).toBe(0)
    expect(result.current.formattedElapsed).toBe('5:00')
  })

  it('formats time as MM:SS for times under an hour', () => {
    const now = Date.now()
    vi.setSystemTime(now)

    // 45 minutes and 30 seconds ago
    const startTime = new Date(now - (45 * 60 + 30) * 1000)
    const { result } = renderHook(() => useTimer(startTime))

    expect(result.current.formattedElapsed).toBe('45:30')
  })

  it('formats time as H:MM:SS for times over an hour', () => {
    const now = Date.now()
    vi.setSystemTime(now)

    // 1 hour, 5 minutes, 30 seconds ago
    const startTime = new Date(now - (1 * 3600 + 5 * 60 + 30) * 1000)
    const { result } = renderHook(() => useTimer(startTime))

    expect(result.current.formattedElapsed).toBe('1:05:30')
    expect(result.current.hours).toBe(1)
    expect(result.current.minutes).toBe(5)
    expect(result.current.seconds).toBe(30)
  })

  it('updates every second', () => {
    const now = Date.now()
    vi.setSystemTime(now)

    const startTime = new Date(now)
    const { result } = renderHook(() => useTimer(startTime))

    expect(result.current.seconds).toBe(0)

    // Advance 3 seconds
    act(() => {
      vi.advanceTimersByTime(3000)
    })

    expect(result.current.seconds).toBe(3)
    expect(result.current.formattedElapsed).toBe('0:03')
  })

  it('resets when startTime changes to null', () => {
    const now = Date.now()
    vi.setSystemTime(now)

    const startTime = new Date(now - 60000)
    const { result, rerender } = renderHook(
      ({ start }) => useTimer(start),
      { initialProps: { start: startTime as Date | null } }
    )

    expect(result.current.minutes).toBe(1)

    // Set startTime to null
    rerender({ start: null })

    expect(result.current.elapsed).toBe(0)
    expect(result.current.formattedElapsed).toBe('0:00')
  })

  it('handles multi-hour durations', () => {
    const now = Date.now()
    vi.setSystemTime(now)

    // 2 hours, 30 minutes, 45 seconds ago
    const startTime = new Date(now - (2 * 3600 + 30 * 60 + 45) * 1000)
    const { result } = renderHook(() => useTimer(startTime))

    expect(result.current.formattedElapsed).toBe('2:30:45')
    expect(result.current.hours).toBe(2)
    expect(result.current.minutes).toBe(30)
    expect(result.current.seconds).toBe(45)
  })
})
