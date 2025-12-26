import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useLiveQuery } from 'dexie-react-hooks'
import { db, type LocalMedication, type LocalMedicationLog } from '@/db/dexie'
import { addPendingEvent } from '@/db/events'
import { apiClient } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-client'
import { useFamilyStore } from '@/stores/family.store'

export interface CreateMedicationInput {
  name: string
  dosage: string
  unit: string
  frequency: string
  instructions?: string
  startDate: Date
  endDate?: Date
}

export interface LogDoseInput {
  medicationId: string
  givenAt?: Date
  givenBy: string
  dosage: string
  notes?: string
}

export interface UpdateMedicationLogInput {
  id: string
  givenAt?: Date
  givenBy?: string
  dosage?: string
  notes?: string
}

interface MedicationResponse {
  id: string
  child_id: string
  name: string
  dosage: string
  unit: string
  frequency: string
  instructions?: string
  start_date: string
  end_date?: string
  active: boolean
}

interface MedicationLogResponse {
  id: string
  medication_id: string
  child_id: string
  given_at: string
  given_by: string
  dosage: string
  notes?: string
}

function mapMedicationToLocal(response: MedicationResponse): LocalMedication {
  return {
    id: response.id,
    childId: response.child_id,
    name: response.name,
    dosage: response.dosage,
    unit: response.unit,
    frequency: response.frequency,
    instructions: response.instructions,
    startDate: new Date(response.start_date),
    endDate: response.end_date ? new Date(response.end_date) : undefined,
    active: response.active,
    syncedAt: new Date(),
    pendingSync: false,
  }
}

function mapLogToLocal(response: MedicationLogResponse): LocalMedicationLog {
  return {
    id: response.id,
    medicationId: response.medication_id,
    childId: response.child_id,
    givenAt: new Date(response.given_at),
    givenBy: response.given_by,
    dosage: response.dosage,
    notes: response.notes,
    syncedAt: new Date(),
    pendingSync: false,
  }
}

export function useMedications() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  const medications = useLiveQuery(
    () =>
      currentChild
        ? db.medications.where('childId').equals(currentChild.id).toArray()
        : [],
    [currentChild?.id]
  )

  const syncQuery = useQuery({
    queryKey: queryKeys.medications.byChild(currentChild?.id ?? ''),
    queryFn: async () => {
      if (!currentChild) return []

      const response = await apiClient.get<MedicationResponse[]>('/api/medications', {
        params: { child_id: currentChild.id },
      })

      for (const med of response.data) {
        await db.medications.put(mapMedicationToLocal(med))
      }

      return response.data
    },
    enabled: !!currentChild,
    staleTime: 30000,
  })

  const activeMedications = medications?.filter((m) => m.active) ?? []
  const inactiveMedications = medications?.filter((m) => !m.active) ?? []

  return {
    medications: medications ?? [],
    activeMedications,
    inactiveMedications,
    isLoading: medications === undefined,
    isSyncing: syncQuery.isFetching,
    error: syncQuery.error,
    refetch: syncQuery.refetch,
  }
}

export function useMedicationLogs(medicationId: string) {
  const logs = useLiveQuery(
    () => db.medicationLogs.where('medicationId').equals(medicationId).reverse().sortBy('givenAt'),
    [medicationId]
  )

  const syncQuery = useQuery({
    queryKey: [...queryKeys.medications.byId(medicationId), 'logs'],
    queryFn: async () => {
      const response = await apiClient.get<MedicationLogResponse[]>(
        `/api/medications/${medicationId}/logs`
      )

      for (const log of response.data) {
        await db.medicationLogs.put(mapLogToLocal(log))
      }

      return response.data
    },
    staleTime: 30000,
  })

  return {
    logs: logs ?? [],
    isLoading: logs === undefined,
    isSyncing: syncQuery.isFetching,
  }
}

export function useAllMedicationLogs() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  const logs = useLiveQuery(
    () =>
      currentChild
        ? db.medicationLogs
            .where('childId')
            .equals(currentChild.id)
            .reverse()
            .sortBy('givenAt')
        : [],
    [currentChild?.id]
  )

  const medications = useLiveQuery(
    () =>
      currentChild
        ? db.medications.where('childId').equals(currentChild.id).toArray()
        : [],
    [currentChild?.id]
  )

  // Enrich logs with medication names
  const enrichedLogs = (logs ?? []).map((log) => {
    const medication = (medications ?? []).find((m) => m.id === log.medicationId)
    return {
      ...log,
      medicationName: medication?.name ?? 'Unknown',
    }
  })

  return {
    logs: enrichedLogs,
    isLoading: logs === undefined,
  }
}

export function useCreateMedication() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: CreateMedicationInput) => {
      if (!currentChild) throw new Error('No child selected')

      const payload = {
        child_id: currentChild.id,
        name: input.name,
        dosage: input.dosage,
        unit: input.unit,
        frequency: input.frequency,
        instructions: input.instructions,
        start_date: input.startDate.toISOString(),
        end_date: input.endDate?.toISOString(),
      }

      try {
        const response = await apiClient.post<MedicationResponse>('/api/medications', payload)
        await db.medications.add(mapMedicationToLocal(response.data))
        return response.data
      } catch {
        const localId = crypto.randomUUID()
        const localMed: LocalMedication = {
          id: localId,
          childId: currentChild.id,
          name: input.name,
          dosage: input.dosage,
          unit: input.unit,
          frequency: input.frequency,
          instructions: input.instructions,
          startDate: input.startDate,
          endDate: input.endDate,
          active: true,
          pendingSync: true,
        }
        await db.medications.add(localMed)
        await addPendingEvent('medication', 'create', localId, payload)
        return localMed
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.medications.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useLogDose() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: LogDoseInput) => {
      if (!currentChild) throw new Error('No child selected')

      const givenAt = input.givenAt ?? new Date()
      const payload = {
        medication_id: input.medicationId,
        child_id: currentChild.id,
        given_at: givenAt.toISOString(),
        given_by: input.givenBy,
        dosage: input.dosage,
        notes: input.notes,
      }

      try {
        const response = await apiClient.post<MedicationLogResponse>('/api/medications/log', payload)
        await db.medicationLogs.add(mapLogToLocal(response.data))
        return response.data
      } catch {
        const localId = crypto.randomUUID()
        const localLog: LocalMedicationLog = {
          id: localId,
          medicationId: input.medicationId,
          childId: currentChild.id,
          givenAt,
          givenBy: input.givenBy,
          dosage: input.dosage,
          notes: input.notes,
          pendingSync: true,
        }
        await db.medicationLogs.add(localLog)
        await addPendingEvent('medication_log', 'create', localId, payload)
        return localLog
      }
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: [...queryKeys.medications.byId(variables.medicationId), 'logs'],
      })
    },
  })
}

export function useUpdateMedicationLog() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: UpdateMedicationLogInput) => {
      const payload: Record<string, unknown> = {}
      if (input.givenAt !== undefined) payload.given_at = input.givenAt.toISOString()
      if (input.givenBy !== undefined) payload.given_by = input.givenBy
      if (input.dosage !== undefined) payload.dosage = input.dosage
      if (input.notes !== undefined) payload.notes = input.notes

      const updateData: Partial<LocalMedicationLog> = {}
      if (input.givenAt !== undefined) updateData.givenAt = input.givenAt
      if (input.givenBy !== undefined) updateData.givenBy = input.givenBy
      if (input.dosage !== undefined) updateData.dosage = input.dosage
      if (input.notes !== undefined) updateData.notes = input.notes

      // Get the log to find medicationId for cache invalidation
      const existingLog = await db.medicationLogs.get(input.id)

      try {
        await apiClient.put(`/api/medications/log/${input.id}`, payload)
        await db.medicationLogs.update(input.id, {
          ...updateData,
          syncedAt: new Date(),
          pendingSync: false,
        })
      } catch {
        await db.medicationLogs.update(input.id, {
          ...updateData,
          pendingSync: true,
        })
        await addPendingEvent('medication_log', 'update', input.id, payload)
      }

      return existingLog?.medicationId
    },
    onSuccess: (medicationId) => {
      if (medicationId) {
        queryClient.invalidateQueries({
          queryKey: [...queryKeys.medications.byId(medicationId), 'logs'],
        })
      }
    },
  })
}

export function useDeleteMedicationLog() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      const existingLog = await db.medicationLogs.get(id)

      try {
        await apiClient.delete(`/api/medications/log/${id}`)
      } catch {
        await addPendingEvent('medication_log', 'delete', id)
      }
      await db.medicationLogs.delete(id)

      return existingLog?.medicationId
    },
    onSuccess: (medicationId) => {
      if (medicationId) {
        queryClient.invalidateQueries({
          queryKey: [...queryKeys.medications.byId(medicationId), 'logs'],
        })
      }
    },
  })
}

export function useDeleteMedication() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await apiClient.delete(`/api/medications/${id}`)
      } catch {
        await addPendingEvent('medication', 'delete', id)
      }
      await db.medications.delete(id)
      await db.medicationLogs.where('medicationId').equals(id).delete()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.medications.byChild(currentChild?.id ?? '') })
    },
  })
}
