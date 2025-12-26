import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useLiveQuery } from 'dexie-react-hooks'
import { db, type LocalAppointment } from '@/db/dexie'
import { addPendingEvent } from '@/db/events'
import { apiClient } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-client'
import { useFamilyStore } from '@/stores/family.store'

export type AppointmentType = 'well_visit' | 'sick_visit' | 'specialist' | 'dental' | 'other'

export interface CreateAppointmentInput {
  type: AppointmentType
  title: string
  provider?: string
  location?: string
  scheduledAt: Date
  duration: number
  notes?: string
}

export interface UpdateAppointmentInput {
  id: string
  type?: AppointmentType
  title?: string
  provider?: string
  location?: string
  scheduledAt?: Date
  duration?: number
  notes?: string
  completed?: boolean
  cancelled?: boolean
}

interface AppointmentResponse {
  id: string
  child_id: string
  type: AppointmentType
  title: string
  provider?: string
  location?: string
  scheduled_at: string
  duration: number
  notes?: string
  completed: boolean
  cancelled: boolean
}

function mapResponseToLocal(response: AppointmentResponse): LocalAppointment {
  return {
    id: response.id,
    childId: response.child_id,
    type: response.type,
    title: response.title,
    provider: response.provider,
    location: response.location,
    scheduledAt: new Date(response.scheduled_at),
    duration: response.duration,
    notes: response.notes,
    completed: response.completed,
    cancelled: response.cancelled,
    syncedAt: new Date(),
    pendingSync: false,
  }
}

export function useAppointments() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  const appointments = useLiveQuery(
    () =>
      currentChild
        ? db.appointments.where('childId').equals(currentChild.id).sortBy('scheduledAt')
        : [],
    [currentChild?.id]
  )

  const syncQuery = useQuery({
    queryKey: queryKeys.appointments.byChild(currentChild?.id ?? ''),
    queryFn: async () => {
      if (!currentChild) return []

      const response = await apiClient.get<AppointmentResponse[]>('/api/appointments', {
        params: { child_id: currentChild.id },
      })

      for (const apt of response.data) {
        await db.appointments.put(mapResponseToLocal(apt))
      }

      return response.data
    },
    enabled: !!currentChild,
    staleTime: 30000,
  })

  const now = new Date()
  const upcoming = appointments?.filter(
    (a) => !a.completed && !a.cancelled && new Date(a.scheduledAt) >= now
  ) ?? []
  const past = appointments?.filter(
    (a) => a.completed || new Date(a.scheduledAt) < now
  ) ?? []
  const cancelled = appointments?.filter((a) => a.cancelled) ?? []

  return {
    appointments: appointments ?? [],
    upcoming,
    past,
    cancelled,
    isLoading: appointments === undefined,
    isSyncing: syncQuery.isFetching,
    error: syncQuery.error,
    refetch: syncQuery.refetch,
  }
}

export function useCreateAppointment() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: CreateAppointmentInput) => {
      if (!currentChild) throw new Error('No child selected')

      const payload = {
        child_id: currentChild.id,
        type: input.type,
        title: input.title,
        provider: input.provider,
        location: input.location,
        scheduled_at: input.scheduledAt.toISOString(),
        duration: input.duration,
        notes: input.notes,
      }

      try {
        const response = await apiClient.post<AppointmentResponse>('/api/appointments', payload)
        await db.appointments.add(mapResponseToLocal(response.data))
        return response.data
      } catch {
        const localId = crypto.randomUUID()
        const localApt: LocalAppointment = {
          id: localId,
          childId: currentChild.id,
          type: input.type,
          title: input.title,
          provider: input.provider,
          location: input.location,
          scheduledAt: input.scheduledAt,
          duration: input.duration,
          notes: input.notes,
          completed: false,
          cancelled: false,
          pendingSync: true,
        }
        await db.appointments.add(localApt)
        await addPendingEvent('appointment', 'create', localId, payload)
        return localApt
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.appointments.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useUpdateAppointment() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: UpdateAppointmentInput) => {
      const payload: Record<string, unknown> = {}
      if (input.type !== undefined) payload.type = input.type
      if (input.title !== undefined) payload.title = input.title
      if (input.provider !== undefined) payload.provider = input.provider
      if (input.location !== undefined) payload.location = input.location
      if (input.scheduledAt !== undefined) payload.scheduled_at = input.scheduledAt.toISOString()
      if (input.duration !== undefined) payload.duration = input.duration
      if (input.notes !== undefined) payload.notes = input.notes
      if (input.completed !== undefined) payload.completed = input.completed
      if (input.cancelled !== undefined) payload.cancelled = input.cancelled

      try {
        const response = await apiClient.patch<AppointmentResponse>(
          `/api/appointments/${input.id}`,
          payload
        )
        await db.appointments.update(input.id, {
          ...input,
          syncedAt: new Date(),
          pendingSync: false,
        })
        return response.data
      } catch {
        await db.appointments.update(input.id, {
          ...input,
          pendingSync: true,
        })
        await addPendingEvent('appointment', 'update', input.id, payload)
        return null
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.appointments.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useDeleteAppointment() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await apiClient.delete(`/api/appointments/${id}`)
      } catch {
        await addPendingEvent('appointment', 'delete', id)
      }
      await db.appointments.delete(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.appointments.byChild(currentChild?.id ?? '') })
    },
  })
}
