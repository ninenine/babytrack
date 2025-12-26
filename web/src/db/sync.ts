import { db } from './dexie'
import { getPendingEvents, removePendingEvent, incrementRetryCount } from './events'
import { apiClient } from '@/lib/api-client'

const MAX_RETRIES = 3

export async function syncPendingEvents(): Promise<{
  synced: number
  failed: number
}> {
  const events = await getPendingEvents()
  let synced = 0
  let failed = 0

  for (const event of events) {
    if (event.retryCount >= MAX_RETRIES) {
      // Too many retries, skip but don't remove
      failed++
      continue
    }

    try {
      await apiClient.post('/api/sync/push', {
        events: [
          {
            id: event.id,
            type: event.type,
            action: event.action,
            entity_id: event.entityId,
            data: event.data,
            timestamp: event.timestamp.toISOString(),
          },
        ],
      })

      // Mark entity as synced
      await markEntityAsSynced(event.type, event.entityId)

      // Remove from pending
      await removePendingEvent(event.id)
      synced++
    } catch (error) {
      console.error('Failed to sync event:', event.id, error)
      await incrementRetryCount(event.id)
      failed++
    }
  }

  return { synced, failed }
}

async function markEntityAsSynced(
  type: string,
  entityId: string
): Promise<void> {
  const now = new Date()

  switch (type) {
    case 'feeding':
      await db.feedings.update(entityId, { syncedAt: now, pendingSync: false })
      break
    case 'sleep':
      await db.sleep.update(entityId, { syncedAt: now, pendingSync: false })
      break
    case 'medication':
      await db.medications.update(entityId, { syncedAt: now, pendingSync: false })
      break
    case 'medication_log':
      await db.medicationLogs.update(entityId, { syncedAt: now, pendingSync: false })
      break
    case 'note':
      await db.notes.update(entityId, { syncedAt: now, pendingSync: false })
      break
    case 'vaccination':
      await db.vaccinations.update(entityId, { syncedAt: now, pendingSync: false })
      break
    case 'appointment':
      await db.appointments.update(entityId, { syncedAt: now, pendingSync: false })
      break
  }
}

interface SyncEvent {
  type: string
  action: string
  entity_id: string
  data: unknown
}

export async function pullFromServer(lastSync?: string): Promise<void> {
  try {
    const response = await apiClient.get<{ events: SyncEvent[] }>('/api/sync/pull', {
      params: { last_sync: lastSync },
    })

    const { events } = response.data

    for (const event of events) {
      await applyServerEvent(event)
    }
  } catch (error) {
    console.error('Failed to pull from server:', error)
  }
}

async function applyServerEvent(event: SyncEvent): Promise<void> {
  const { type, action, entity_id, data } = event

  switch (type) {
    case 'feeding':
      if (action === 'delete') {
        await db.feedings.delete(entity_id)
      } else {
        await db.feedings.put({ ...(data as object), id: entity_id, pendingSync: false } as never)
      }
      break
    case 'sleep':
      if (action === 'delete') {
        await db.sleep.delete(entity_id)
      } else {
        await db.sleep.put({ ...(data as object), id: entity_id, pendingSync: false } as never)
      }
      break
    case 'medication':
      if (action === 'delete') {
        await db.medications.delete(entity_id)
      } else {
        await db.medications.put({ ...(data as object), id: entity_id, pendingSync: false } as never)
      }
      break
    case 'note':
      if (action === 'delete') {
        await db.notes.delete(entity_id)
      } else {
        await db.notes.put({ ...(data as object), id: entity_id, pendingSync: false } as never)
      }
      break
    case 'vaccination':
      if (action === 'delete') {
        await db.vaccinations.delete(entity_id)
      } else {
        await db.vaccinations.put({ ...(data as object), id: entity_id, pendingSync: false } as never)
      }
      break
    case 'appointment':
      if (action === 'delete') {
        await db.appointments.delete(entity_id)
      } else {
        await db.appointments.put({ ...(data as object), id: entity_id, pendingSync: false } as never)
      }
      break
  }
}
