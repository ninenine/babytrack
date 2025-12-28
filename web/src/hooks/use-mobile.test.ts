import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useMobile } from './use-mobile'

describe('useMobile', () => {
  const originalInnerWidth = window.innerWidth
  let mockMatchMedia: ReturnType<typeof vi.fn>
  let changeHandler: (() => void) | null = null

  beforeEach(() => {
    vi.clearAllMocks()
    changeHandler = null

    mockMatchMedia = vi.fn().mockImplementation((query: string) => ({
      matches: query.includes('767'),
      media: query,
      onchange: null,
      addEventListener: (event: string, handler: () => void) => {
        if (event === 'change') {
          changeHandler = handler
        }
      },
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }))

    Object.defineProperty(window, 'matchMedia', {
      value: mockMatchMedia,
      writable: true,
    })
  })

  afterEach(() => {
    Object.defineProperty(window, 'innerWidth', {
      value: originalInnerWidth,
      writable: true,
    })
  })

  it('should return true on mobile screen', () => {
    Object.defineProperty(window, 'innerWidth', { value: 500, writable: true })
    const { result } = renderHook(() => useMobile())
    expect(result.current).toBe(true)
  })

  it('should return false on desktop screen', () => {
    Object.defineProperty(window, 'innerWidth', { value: 1024, writable: true })
    const { result } = renderHook(() => useMobile())
    expect(result.current).toBe(false)
  })

  it('should return false at exactly 768px (breakpoint)', () => {
    Object.defineProperty(window, 'innerWidth', { value: 768, writable: true })
    const { result } = renderHook(() => useMobile())
    expect(result.current).toBe(false)
  })

  it('should return true at 767px (just below breakpoint)', () => {
    Object.defineProperty(window, 'innerWidth', { value: 767, writable: true })
    const { result } = renderHook(() => useMobile())
    expect(result.current).toBe(true)
  })

  it('should update when screen size changes', () => {
    Object.defineProperty(window, 'innerWidth', { value: 1024, writable: true })
    const { result } = renderHook(() => useMobile())

    expect(result.current).toBe(false)

    // Simulate resize to mobile
    act(() => {
      Object.defineProperty(window, 'innerWidth', { value: 500, writable: true })
      if (changeHandler) {
        changeHandler()
      }
    })

    expect(result.current).toBe(true)
  })
})
