import { create } from 'zustand'

interface UIState {
  isSidebarOpen: boolean
  isLoading: boolean
  error: string | null
  theme: 'light' | 'dark' | 'system'
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
  setTheme: (theme: 'light' | 'dark' | 'system') => void
}

export const useUIStore = create<UIState>((set) => ({
  isSidebarOpen: false,
  isLoading: false,
  error: null,
  theme: 'system',

  toggleSidebar: () =>
    set((state) => ({ isSidebarOpen: !state.isSidebarOpen })),

  setSidebarOpen: (open) => set({ isSidebarOpen: open }),

  setLoading: (loading) => set({ isLoading: loading }),

  setError: (error) => set({ error }),

  clearError: () => set({ error: null }),

  setTheme: (theme) => set({ theme }),
}))
