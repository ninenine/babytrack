import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Keep data fresh for 30 seconds
      staleTime: 30 * 1000,
      // Cache data for 5 minutes
      gcTime: 5 * 60 * 1000,
      // Retry failed requests up to 2 times
      retry: 2,
      // Don't retry on 4xx errors
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      // Refetch on window focus for real-time data
      refetchOnWindowFocus: true,
      // Don't refetch on mount if data is fresh
      refetchOnMount: true,
      // Keep previous data while fetching new data
      placeholderData: (previousData: unknown) => previousData,
    },
    mutations: {
      // Retry mutations once on failure
      retry: 1,
    },
  },
})

// Query keys for type-safe query invalidation
export const queryKeys = {
  // Auth
  auth: {
    me: ['auth', 'me'] as const,
  },
  // Families
  families: {
    all: ['families'] as const,
    byId: (id: string) => ['families', id] as const,
  },
  // Feedings
  feedings: {
    all: ['feedings'] as const,
    byChild: (childId: string) => ['feedings', 'child', childId] as const,
    byId: (id: string) => ['feedings', id] as const,
  },
  // Sleep
  sleep: {
    all: ['sleep'] as const,
    byChild: (childId: string) => ['sleep', 'child', childId] as const,
    active: (childId: string) => ['sleep', 'active', childId] as const,
    byId: (id: string) => ['sleep', id] as const,
  },
  // Medications
  medications: {
    all: ['medications'] as const,
    byChild: (childId: string) => ['medications', 'child', childId] as const,
    active: (childId: string) => ['medications', 'active', childId] as const,
    byId: (id: string) => ['medications', id] as const,
    logs: (medicationId: string) => ['medications', medicationId, 'logs'] as const,
  },
  // Vaccinations
  vaccinations: {
    all: ['vaccinations'] as const,
    byChild: (childId: string) => ['vaccinations', 'child', childId] as const,
    upcoming: (childId: string) => ['vaccinations', 'upcoming', childId] as const,
    schedule: ['vaccinations', 'schedule'] as const,
    byId: (id: string) => ['vaccinations', id] as const,
  },
  // Appointments
  appointments: {
    all: ['appointments'] as const,
    byChild: (childId: string) => ['appointments', 'child', childId] as const,
    upcoming: (childId: string) => ['appointments', 'upcoming', childId] as const,
    byId: (id: string) => ['appointments', id] as const,
  },
  // Notes
  notes: {
    all: ['notes'] as const,
    byChild: (childId: string) => ['notes', 'child', childId] as const,
    byId: (id: string) => ['notes', id] as const,
  },
  // Sync
  sync: {
    status: ['sync', 'status'] as const,
  },
} as const
