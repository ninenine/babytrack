import { db, PendingEvent } from './dexie'

export async function addPendingEvent(
  type: PendingEvent['type'],
  action: PendingEvent['action'],
  entityId: string,
  data?: unknown
): Promise<void> {
  const event: PendingEvent = {
    id: crypto.randomUUID(),
    type,
    action,
    entityId,
    data,
    timestamp: new Date(),
    retryCount: 0,
  }

  await db.pendingEvents.add(event)
}

export async function getPendingEvents(): Promise<PendingEvent[]> {
  return db.pendingEvents.orderBy('timestamp').toArray()
}

export async function removePendingEvent(id: string): Promise<void> {
  await db.pendingEvents.delete(id)
}

export async function incrementRetryCount(id: string): Promise<void> {
  await db.pendingEvents.update(id, {
    retryCount: (await db.pendingEvents.get(id))?.retryCount ?? 0 + 1,
  })
}

export async function clearPendingEvents(): Promise<void> {
  await db.pendingEvents.clear()
}

export async function getPendingEventCount(): Promise<number> {
  return db.pendingEvents.count()
}
