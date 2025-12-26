import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useLiveQuery } from 'dexie-react-hooks'
import { db, type LocalSleep } from '@/db/dexie'
import { addPendingEvent } from '@/db/events'
import { apiClient } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-client'
import { useFamilyStore } from '@/stores/family.store'

export type SleepType = 'nap' | 'night'

export interface StartSleepInput {
  type: SleepType
  startTime?: Date
  notes?: string
}

export interface EndSleepInput {
  id: string
  endTime?: Date
  quality?: number
  notes?: string
}

export interface UpdateSleepInput {
  id: string
  type?: SleepType
  startTime?: Date
  endTime?: Date | null
  quality?: number
  notes?: string
}

interface SleepResponse {
  id: string
  child_id: string
  type: SleepType
  start_time: string
  end_time?: string
  quality?: number
  notes?: string
}

function mapResponseToLocal(response: SleepResponse): LocalSleep {
  return {
    id: response.id,
    childId: response.child_id,
    type: response.type,
    startTime: new Date(response.start_time),
    endTime: response.end_time ? new Date(response.end_time) : undefined,
    quality: response.quality,
    notes: response.notes,
    syncedAt: new Date(),
    pendingSync: false,
  }
}

export function useSleep() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  const sleepRecords = useLiveQuery(
    () =>
      currentChild
        ? db.sleep
            .where('childId')
            .equals(currentChild.id)
            .reverse()
            .sortBy('startTime')
        : [],
    [currentChild?.id]
  )

  const syncQuery = useQuery({
    queryKey: queryKeys.sleep.byChild(currentChild?.id ?? ''),
    queryFn: async () => {
      if (!currentChild) return []

      const response = await apiClient.get<SleepResponse[]>('/api/sleep', {
        params: { child_id: currentChild.id },
      })

      for (const sleep of response.data) {
        await db.sleep.put(mapResponseToLocal(sleep))
      }

      return response.data
    },
    enabled: !!currentChild,
    staleTime: 30000,
  })

  const activeSleep = sleepRecords?.find((s) => !s.endTime)

  return {
    sleepRecords: sleepRecords ?? [],
    activeSleep,
    isLoading: sleepRecords === undefined,
    isSyncing: syncQuery.isFetching,
    error: syncQuery.error,
    refetch: syncQuery.refetch,
  }
}

export function useStartSleep() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: StartSleepInput) => {
      if (!currentChild) throw new Error('No child selected')

      const startTime = input.startTime ?? new Date()
      const payload = {
        child_id: currentChild.id,
        type: input.type,
        start_time: startTime.toISOString(),
        notes: input.notes,
      }

      try {
        const response = await apiClient.post<SleepResponse>('/api/sleep/start', payload)
        await db.sleep.add(mapResponseToLocal(response.data))
        return response.data
      } catch {
        const localId = crypto.randomUUID()
        const localSleep: LocalSleep = {
          id: localId,
          childId: currentChild.id,
          type: input.type,
          startTime,
          notes: input.notes,
          pendingSync: true,
        }
        await db.sleep.add(localSleep)
        await addPendingEvent('sleep', 'create', localId, payload)
        return localSleep
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sleep.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useEndSleep() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: EndSleepInput) => {
      const endTime = input.endTime ?? new Date()
      const payload = {
        end_time: endTime.toISOString(),
        quality: input.quality,
        notes: input.notes,
      }

      try {
        const response = await apiClient.post<SleepResponse>(`/api/sleep/${input.id}/end`, payload)
        await db.sleep.update(input.id, {
          endTime,
          quality: input.quality,
          notes: input.notes,
          syncedAt: new Date(),
          pendingSync: false,
        })
        return response.data
      } catch {
        await db.sleep.update(input.id, {
          endTime,
          quality: input.quality,
          notes: input.notes,
          pendingSync: true,
        })
        await addPendingEvent('sleep', 'update', input.id, payload)
        return null
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sleep.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useUpdateSleep() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: UpdateSleepInput) => {
      const payload: Record<string, unknown> = {}
      if (input.type !== undefined) payload.type = input.type
      if (input.startTime !== undefined) payload.start_time = input.startTime.toISOString()
      if (input.endTime !== undefined) payload.end_time = input.endTime?.toISOString() ?? null
      if (input.quality !== undefined) payload.quality = input.quality
      if (input.notes !== undefined) payload.notes = input.notes

      const updateData: Partial<LocalSleep> = {}
      if (input.type !== undefined) updateData.type = input.type
      if (input.startTime !== undefined) updateData.startTime = input.startTime
      if (input.endTime !== undefined) updateData.endTime = input.endTime ?? undefined
      if (input.quality !== undefined) updateData.quality = input.quality
      if (input.notes !== undefined) updateData.notes = input.notes

      try {
        const response = await apiClient.put<SleepResponse>(`/api/sleep/${input.id}`, payload)
        await db.sleep.update(input.id, {
          ...updateData,
          syncedAt: new Date(),
          pendingSync: false,
        })
        return response.data
      } catch {
        await db.sleep.update(input.id, {
          ...updateData,
          pendingSync: true,
        })
        await addPendingEvent('sleep', 'update', input.id, payload)
        return null
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sleep.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useDeleteSleep() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await apiClient.delete(`/api/sleep/${id}`)
      } catch {
        await addPendingEvent('sleep', 'delete', id)
      }
      await db.sleep.delete(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sleep.byChild(currentChild?.id ?? '') })
    },
  })
}
