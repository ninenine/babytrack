import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useLiveQuery } from 'dexie-react-hooks'
import { db, type LocalFeeding } from '@/db/dexie'
import { addPendingEvent } from '@/db/events'
import { apiClient } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-client'
import { useFamilyStore } from '@/stores/family.store'

export type FeedingType = 'breast' | 'bottle' | 'formula' | 'solid'

export interface CreateFeedingInput {
  type: FeedingType
  startTime: Date
  endTime?: Date
  amount?: number
  unit?: string
  side?: string
  notes?: string
}

export interface UpdateFeedingInput {
  id: string
  type?: FeedingType
  startTime?: Date
  endTime?: Date | null
  amount?: number | null
  unit?: string
  side?: string
  notes?: string
}

interface FeedingResponse {
  id: string
  child_id: string
  type: FeedingType
  start_time: string
  end_time?: string
  amount?: number
  unit?: string
  side?: string
  notes?: string
}

function mapResponseToLocal(response: FeedingResponse): LocalFeeding {
  return {
    id: response.id,
    childId: response.child_id,
    type: response.type,
    startTime: new Date(response.start_time),
    endTime: response.end_time ? new Date(response.end_time) : undefined,
    amount: response.amount,
    unit: response.unit,
    side: response.side,
    notes: response.notes,
    syncedAt: new Date(),
    pendingSync: false,
  }
}

export function useFeedings() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  // Live query from local DB
  const feedings = useLiveQuery(
    () =>
      currentChild
        ? db.feedings
            .where('childId')
            .equals(currentChild.id)
            .reverse()
            .sortBy('startTime')
        : [],
    [currentChild?.id]
  )

  // Sync from server query
  const syncQuery = useQuery({
    queryKey: queryKeys.feedings.byChild(currentChild?.id ?? ''),
    queryFn: async () => {
      if (!currentChild) return []

      const response = await apiClient.get<FeedingResponse[]>('/api/feeding', {
        params: { child_id: currentChild.id },
      })

      // Upsert into local DB
      for (const feeding of response.data) {
        await db.feedings.put(mapResponseToLocal(feeding))
      }

      return response.data
    },
    enabled: !!currentChild,
    staleTime: 30000,
  })

  return {
    feedings: feedings ?? [],
    isLoading: feedings === undefined,
    isSyncing: syncQuery.isFetching,
    error: syncQuery.error,
    refetch: syncQuery.refetch,
  }
}

export function useCreateFeeding() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: CreateFeedingInput) => {
      if (!currentChild) throw new Error('No child selected')

      const payload = {
        child_id: currentChild.id,
        type: input.type,
        start_time: input.startTime.toISOString(),
        end_time: input.endTime?.toISOString(),
        amount: input.amount,
        unit: input.unit,
        side: input.side,
        notes: input.notes,
      }

      try {
        const response = await apiClient.post<FeedingResponse>('/api/feeding', payload)
        await db.feedings.add(mapResponseToLocal(response.data))
        return response.data
      } catch {
        // Offline: save locally with pending sync
        const localId = crypto.randomUUID()
        const localFeeding: LocalFeeding = {
          id: localId,
          childId: currentChild.id,
          type: input.type,
          startTime: input.startTime,
          endTime: input.endTime,
          amount: input.amount,
          unit: input.unit,
          side: input.side,
          notes: input.notes,
          pendingSync: true,
        }
        await db.feedings.add(localFeeding)
        await addPendingEvent('feeding', 'create', localId, payload)
        return localFeeding
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.feedings.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useUpdateFeeding() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: UpdateFeedingInput) => {
      if (!currentChild) throw new Error('No child selected')

      const payload: Record<string, unknown> = {
        child_id: currentChild.id,
      }
      if (input.type !== undefined) payload.type = input.type
      if (input.startTime !== undefined) payload.start_time = input.startTime.toISOString()
      if (input.endTime !== undefined) payload.end_time = input.endTime?.toISOString() ?? null
      if (input.amount !== undefined) payload.amount = input.amount
      if (input.unit !== undefined) payload.unit = input.unit
      if (input.side !== undefined) payload.side = input.side
      if (input.notes !== undefined) payload.notes = input.notes

      const updateData: Partial<LocalFeeding> = {}
      if (input.type !== undefined) updateData.type = input.type
      if (input.startTime !== undefined) updateData.startTime = input.startTime
      if (input.endTime !== undefined) updateData.endTime = input.endTime ?? undefined
      if (input.amount !== undefined) updateData.amount = input.amount ?? undefined
      if (input.unit !== undefined) updateData.unit = input.unit
      if (input.side !== undefined) updateData.side = input.side
      if (input.notes !== undefined) updateData.notes = input.notes

      try {
        const response = await apiClient.put<FeedingResponse>(`/api/feeding/${input.id}`, payload)
        await db.feedings.update(input.id, {
          ...updateData,
          syncedAt: new Date(),
          pendingSync: false,
        })
        return response.data
      } catch {
        await db.feedings.update(input.id, {
          ...updateData,
          pendingSync: true,
        })
        await addPendingEvent('feeding', 'update', input.id, payload)
        return null
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.feedings.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useDeleteFeeding() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await apiClient.delete(`/api/feeding/${id}`)
      } catch {
        await addPendingEvent('feeding', 'delete', id)
      }
      await db.feedings.delete(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.feedings.byChild(currentChild?.id ?? '') })
    },
  })
}
