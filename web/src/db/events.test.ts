import { describe, it, expect, beforeEach } from 'vitest'
import { db } from './dexie'
import {
  addPendingEvent,
  getPendingEvents,
  removePendingEvent,
  incrementRetryCount,
  clearPendingEvents,
  getPendingEventCount,
} from './events'

describe('Pending Events', () => {
  beforeEach(async () => {
    // Clear the database before each test
    await db.pendingEvents.clear()
  })

  describe('addPendingEvent', () => {
    it('should add a pending event to the database', async () => {
      await addPendingEvent('feeding', 'create', 'entity-123', { foo: 'bar' })

      const events = await getPendingEvents()
      expect(events).toHaveLength(1)
      expect(events[0]?.type).toBe('feeding')
      expect(events[0]?.action).toBe('create')
      expect(events[0]?.entityId).toBe('entity-123')
      expect(events[0]?.data).toEqual({ foo: 'bar' })
      expect(events[0]?.retryCount).toBe(0)
    })

    it('should generate unique ids for each event', async () => {
      await addPendingEvent('feeding', 'create', 'entity-1')
      await addPendingEvent('feeding', 'create', 'entity-2')

      const events = await getPendingEvents()
      expect(events).toHaveLength(2)
      expect(events[0]?.id).not.toBe(events[1]?.id)
    })

    it('should set timestamp to current time', async () => {
      const before = new Date()
      await addPendingEvent('sleep', 'update', 'entity-123')
      const after = new Date()

      const events = await getPendingEvents()
      expect(events[0]?.timestamp.getTime()).toBeGreaterThanOrEqual(before.getTime())
      expect(events[0]?.timestamp.getTime()).toBeLessThanOrEqual(after.getTime())
    })
  })

  describe('getPendingEvents', () => {
    it('should return all added events', async () => {
      await addPendingEvent('feeding', 'create', 'entity-1')
      await addPendingEvent('sleep', 'create', 'entity-2')
      await addPendingEvent('note', 'create', 'entity-3')

      const events = await getPendingEvents()
      expect(events).toHaveLength(3)
      // All events should be present (order may vary due to timestamp precision)
      const entityIds = events.map((e) => e.entityId)
      expect(entityIds).toContain('entity-1')
      expect(entityIds).toContain('entity-2')
      expect(entityIds).toContain('entity-3')
    })

    it('should return empty array when no events', async () => {
      const events = await getPendingEvents()
      expect(events).toHaveLength(0)
    })
  })

  describe('removePendingEvent', () => {
    it('should remove a pending event by id', async () => {
      await addPendingEvent('feeding', 'create', 'entity-1')
      await addPendingEvent('sleep', 'create', 'entity-2')

      const events = await getPendingEvents()
      const eventToRemove = events.find((e) => e.entityId === 'entity-1')
      await removePendingEvent(eventToRemove!.id)

      const remaining = await getPendingEvents()
      expect(remaining).toHaveLength(1)
      expect(remaining[0]?.entityId).toBe('entity-2')
    })

    it('should not throw when removing non-existent event', async () => {
      await expect(removePendingEvent('non-existent-id')).resolves.not.toThrow()
    })
  })

  describe('incrementRetryCount', () => {
    it('should increment the retry count', async () => {
      await addPendingEvent('feeding', 'create', 'entity-1')

      const events = await getPendingEvents()
      expect(events[0]?.retryCount).toBe(0)

      await incrementRetryCount(events[0]!.id)

      const updated = await getPendingEvents()
      expect(updated[0]?.retryCount).toBe(1)
    })

    it('should not throw when event does not exist', async () => {
      await expect(incrementRetryCount('non-existent-id')).resolves.not.toThrow()
    })
  })

  describe('clearPendingEvents', () => {
    it('should remove all pending events', async () => {
      await addPendingEvent('feeding', 'create', 'entity-1')
      await addPendingEvent('sleep', 'create', 'entity-2')
      await addPendingEvent('note', 'create', 'entity-3')

      expect(await getPendingEventCount()).toBe(3)

      await clearPendingEvents()

      expect(await getPendingEventCount()).toBe(0)
    })
  })

  describe('getPendingEventCount', () => {
    it('should return correct count', async () => {
      expect(await getPendingEventCount()).toBe(0)

      await addPendingEvent('feeding', 'create', 'entity-1')
      expect(await getPendingEventCount()).toBe(1)

      await addPendingEvent('sleep', 'create', 'entity-2')
      expect(await getPendingEventCount()).toBe(2)

      await clearPendingEvents()
      expect(await getPendingEventCount()).toBe(0)
    })
  })

  describe('event types', () => {
    it('should support all entity types', async () => {
      const types = ['feeding', 'sleep', 'medication', 'medication_log', 'note', 'vaccination', 'appointment'] as const

      for (const type of types) {
        await addPendingEvent(type, 'create', `${type}-entity`)
      }

      const events = await getPendingEvents()
      expect(events).toHaveLength(7)

      const eventTypes = events.map((e) => e.type)
      for (const type of types) {
        expect(eventTypes).toContain(type)
      }
    })

    it('should support all action types', async () => {
      const actions = ['create', 'update', 'delete'] as const

      for (const action of actions) {
        await addPendingEvent('feeding', action, `${action}-entity`)
      }

      const events = await getPendingEvents()
      expect(events).toHaveLength(3)

      const eventActions = events.map((e) => e.action)
      for (const action of actions) {
        expect(eventActions).toContain(action)
      }
    })
  })
})
