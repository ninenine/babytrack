import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useOnline } from './use-online'

describe('useOnline', () => {
  const originalNavigator = window.navigator

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    Object.defineProperty(window, 'navigator', {
      value: originalNavigator,
      writable: true,
    })
  })

  it('should return true when online', () => {
    Object.defineProperty(window.navigator, 'onLine', { value: true, writable: true })
    const { result } = renderHook(() => useOnline())
    expect(result.current).toBe(true)
  })

  it('should return false when offline', () => {
    Object.defineProperty(window.navigator, 'onLine', { value: false, writable: true })
    const { result } = renderHook(() => useOnline())
    expect(result.current).toBe(false)
  })

  it('should update when going offline', () => {
    Object.defineProperty(window.navigator, 'onLine', { value: true, writable: true })
    const { result } = renderHook(() => useOnline())

    expect(result.current).toBe(true)

    act(() => {
      window.dispatchEvent(new Event('offline'))
    })

    expect(result.current).toBe(false)
  })

  it('should update when coming online', () => {
    Object.defineProperty(window.navigator, 'onLine', { value: false, writable: true })
    const { result } = renderHook(() => useOnline())

    expect(result.current).toBe(false)

    act(() => {
      window.dispatchEvent(new Event('online'))
    })

    expect(result.current).toBe(true)
  })

  it('should clean up event listeners on unmount', () => {
    const addEventListenerSpy = vi.spyOn(window, 'addEventListener')
    const removeEventListenerSpy = vi.spyOn(window, 'removeEventListener')

    const { unmount } = renderHook(() => useOnline())

    expect(addEventListenerSpy).toHaveBeenCalledWith('online', expect.any(Function))
    expect(addEventListenerSpy).toHaveBeenCalledWith('offline', expect.any(Function))

    unmount()

    expect(removeEventListenerSpy).toHaveBeenCalledWith('online', expect.any(Function))
    expect(removeEventListenerSpy).toHaveBeenCalledWith('offline', expect.any(Function))
  })
})
