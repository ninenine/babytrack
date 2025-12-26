import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useLiveQuery } from 'dexie-react-hooks'
import { db, type LocalVaccination } from '@/db/dexie'
import { addPendingEvent } from '@/db/events'
import { apiClient } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-client'
import { useFamilyStore } from '@/stores/family.store'

export interface CreateVaccinationInput {
  name: string
  dose: number
  scheduledAt: Date
  notes?: string
}

export interface RecordVaccinationInput {
  id: string
  administeredAt?: Date
  provider?: string
  location?: string
  lotNumber?: string
  notes?: string
}

export interface UpdateVaccinationInput {
  id: string
  name?: string
  dose?: number
  scheduledAt?: Date
  administeredAt?: Date | null
  provider?: string
  location?: string
  lotNumber?: string
  notes?: string
  completed?: boolean
}

interface VaccinationResponse {
  id: string
  child_id: string
  name: string
  dose: number
  scheduled_at: string
  administered_at?: string
  provider?: string
  location?: string
  lot_number?: string
  notes?: string
  completed: boolean
}

function mapResponseToLocal(response: VaccinationResponse): LocalVaccination {
  return {
    id: response.id,
    childId: response.child_id,
    name: response.name,
    dose: response.dose,
    scheduledAt: new Date(response.scheduled_at),
    administeredAt: response.administered_at ? new Date(response.administered_at) : undefined,
    provider: response.provider,
    location: response.location,
    lotNumber: response.lot_number,
    notes: response.notes,
    completed: response.completed,
    syncedAt: new Date(),
    pendingSync: false,
  }
}

export function useVaccinations() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  const vaccinations = useLiveQuery(
    () =>
      currentChild
        ? db.vaccinations.where('childId').equals(currentChild.id).sortBy('scheduledAt')
        : [],
    [currentChild?.id]
  )

  const syncQuery = useQuery({
    queryKey: queryKeys.vaccinations.byChild(currentChild?.id ?? ''),
    queryFn: async () => {
      if (!currentChild) return []

      const response = await apiClient.get<VaccinationResponse[]>('/api/vaccinations', {
        params: { child_id: currentChild.id },
      })

      for (const vax of response.data) {
        await db.vaccinations.put(mapResponseToLocal(vax))
      }

      return response.data
    },
    enabled: !!currentChild,
    staleTime: 30000,
  })

  const now = new Date()
  const upcoming = vaccinations?.filter((v) => !v.completed && new Date(v.scheduledAt) >= now) ?? []
  const overdue = vaccinations?.filter((v) => !v.completed && new Date(v.scheduledAt) < now) ?? []
  const completed = vaccinations?.filter((v) => v.completed) ?? []

  return {
    vaccinations: vaccinations ?? [],
    upcoming,
    overdue,
    completed,
    isLoading: vaccinations === undefined,
    isSyncing: syncQuery.isFetching,
    error: syncQuery.error,
    refetch: syncQuery.refetch,
  }
}

export function useCreateVaccination() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: CreateVaccinationInput) => {
      if (!currentChild) throw new Error('No child selected')

      const payload = {
        child_id: currentChild.id,
        name: input.name,
        dose: input.dose,
        scheduled_at: input.scheduledAt.toISOString(),
        notes: input.notes,
      }

      try {
        const response = await apiClient.post<VaccinationResponse>('/api/vaccinations', payload)
        await db.vaccinations.add(mapResponseToLocal(response.data))
        return response.data
      } catch {
        const localId = crypto.randomUUID()
        const localVax: LocalVaccination = {
          id: localId,
          childId: currentChild.id,
          name: input.name,
          dose: input.dose,
          scheduledAt: input.scheduledAt,
          notes: input.notes,
          completed: false,
          pendingSync: true,
        }
        await db.vaccinations.add(localVax)
        await addPendingEvent('vaccination', 'create', localId, payload)
        return localVax
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vaccinations.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useRecordVaccination() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: RecordVaccinationInput) => {
      const administeredAt = input.administeredAt ?? new Date()
      const payload = {
        administered_at: administeredAt.toISOString(),
        provider: input.provider,
        location: input.location,
        lot_number: input.lotNumber,
        notes: input.notes,
      }

      try {
        const response = await apiClient.post<VaccinationResponse>(
          `/api/vaccinations/${input.id}/record`,
          payload
        )
        await db.vaccinations.update(input.id, {
          administeredAt,
          provider: input.provider,
          location: input.location,
          lotNumber: input.lotNumber,
          notes: input.notes,
          completed: true,
          syncedAt: new Date(),
          pendingSync: false,
        })
        return response.data
      } catch {
        await db.vaccinations.update(input.id, {
          administeredAt,
          provider: input.provider,
          location: input.location,
          lotNumber: input.lotNumber,
          notes: input.notes,
          completed: true,
          pendingSync: true,
        })
        await addPendingEvent('vaccination', 'update', input.id, payload)
        return null
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vaccinations.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useUpdateVaccination() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: UpdateVaccinationInput) => {
      const payload: Record<string, unknown> = {}
      if (input.name !== undefined) payload.name = input.name
      if (input.dose !== undefined) payload.dose = input.dose
      if (input.scheduledAt !== undefined) payload.scheduled_at = input.scheduledAt.toISOString()
      if (input.administeredAt !== undefined) payload.administered_at = input.administeredAt?.toISOString() ?? null
      if (input.provider !== undefined) payload.provider = input.provider
      if (input.location !== undefined) payload.location = input.location
      if (input.lotNumber !== undefined) payload.lot_number = input.lotNumber
      if (input.notes !== undefined) payload.notes = input.notes
      if (input.completed !== undefined) payload.completed = input.completed

      const updateData: Partial<LocalVaccination> = {}
      if (input.name !== undefined) updateData.name = input.name
      if (input.dose !== undefined) updateData.dose = input.dose
      if (input.scheduledAt !== undefined) updateData.scheduledAt = input.scheduledAt
      if (input.administeredAt !== undefined) updateData.administeredAt = input.administeredAt ?? undefined
      if (input.provider !== undefined) updateData.provider = input.provider
      if (input.location !== undefined) updateData.location = input.location
      if (input.lotNumber !== undefined) updateData.lotNumber = input.lotNumber
      if (input.notes !== undefined) updateData.notes = input.notes
      if (input.completed !== undefined) updateData.completed = input.completed

      try {
        const response = await apiClient.put<VaccinationResponse>(
          `/api/vaccinations/${input.id}`,
          payload
        )
        await db.vaccinations.update(input.id, {
          ...updateData,
          syncedAt: new Date(),
          pendingSync: false,
        })
        return response.data
      } catch {
        await db.vaccinations.update(input.id, {
          ...updateData,
          pendingSync: true,
        })
        await addPendingEvent('vaccination', 'update', input.id, payload)
        return null
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vaccinations.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useDeleteVaccination() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await apiClient.delete(`/api/vaccinations/${id}`)
      } catch {
        await addPendingEvent('vaccination', 'delete', id)
      }
      await db.vaccinations.delete(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vaccinations.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useGenerateVaccinationSchedule() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async () => {
      if (!currentChild) throw new Error('No child selected')

      const response = await apiClient.post<VaccinationResponse[]>('/api/vaccinations/schedule', {
        child_id: currentChild.id,
      })

      for (const vax of response.data) {
        await db.vaccinations.put(mapResponseToLocal(vax))
      }

      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vaccinations.byChild(currentChild?.id ?? '') })
    },
  })
}
