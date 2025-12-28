import { describe, it, expect, beforeEach } from 'vitest'
import { useUIStore } from './ui.store'

describe('useUIStore', () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    useUIStore.setState({
      isSidebarOpen: false,
      isLoading: false,
      error: null,
      theme: 'system',
    })
  })

  describe('sidebar', () => {
    it('should toggle sidebar', () => {
      expect(useUIStore.getState().isSidebarOpen).toBe(false)
      useUIStore.getState().toggleSidebar()
      expect(useUIStore.getState().isSidebarOpen).toBe(true)
      useUIStore.getState().toggleSidebar()
      expect(useUIStore.getState().isSidebarOpen).toBe(false)
    })

    it('should set sidebar open state', () => {
      useUIStore.getState().setSidebarOpen(true)
      expect(useUIStore.getState().isSidebarOpen).toBe(true)
      useUIStore.getState().setSidebarOpen(false)
      expect(useUIStore.getState().isSidebarOpen).toBe(false)
    })
  })

  describe('loading', () => {
    it('should set loading state', () => {
      expect(useUIStore.getState().isLoading).toBe(false)
      useUIStore.getState().setLoading(true)
      expect(useUIStore.getState().isLoading).toBe(true)
      useUIStore.getState().setLoading(false)
      expect(useUIStore.getState().isLoading).toBe(false)
    })
  })

  describe('error', () => {
    it('should set error', () => {
      expect(useUIStore.getState().error).toBeNull()
      useUIStore.getState().setError('Something went wrong')
      expect(useUIStore.getState().error).toBe('Something went wrong')
    })

    it('should clear error', () => {
      useUIStore.getState().setError('An error')
      expect(useUIStore.getState().error).toBe('An error')
      useUIStore.getState().clearError()
      expect(useUIStore.getState().error).toBeNull()
    })
  })

  describe('theme', () => {
    it('should set theme', () => {
      expect(useUIStore.getState().theme).toBe('system')
      useUIStore.getState().setTheme('dark')
      expect(useUIStore.getState().theme).toBe('dark')
      useUIStore.getState().setTheme('light')
      expect(useUIStore.getState().theme).toBe('light')
    })
  })
})
