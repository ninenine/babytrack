import { describe, it, expect, beforeEach, vi } from 'vitest'
import { db } from './dexie'
import { addPendingEvent, getPendingEvents } from './events'
import { syncPendingEvents, pullFromServer } from './sync'

// Mock the API client
vi.mock('@/lib/api-client', () => ({
  apiClient: {
    post: vi.fn(),
    get: vi.fn(),
  },
}))

import { apiClient } from '@/lib/api-client'

describe('Sync', () => {
  beforeEach(async () => {
    // Clear all tables before each test
    await db.pendingEvents.clear()
    await db.feedings.clear()
    await db.sleep.clear()
    await db.medications.clear()
    await db.notes.clear()
    await db.vaccinations.clear()
    await db.appointments.clear()

    // Reset mocks
    vi.clearAllMocks()
  })

  describe('syncPendingEvents', () => {
    it('should sync pending events to server', async () => {
      // Add a feeding to local DB
      await db.feedings.add({
        id: 'feeding-1',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date(),
        pendingSync: true,
      })

      // Add pending event
      await addPendingEvent('feeding', 'create', 'feeding-1', {
        child_id: 'child-1',
        type: 'bottle',
      })

      // Mock successful API response
      vi.mocked(apiClient.post).mockResolvedValueOnce({ data: {}, status: 200 })

      const result = await syncPendingEvents()

      expect(result.synced).toBe(1)
      expect(result.failed).toBe(0)
      expect(apiClient.post).toHaveBeenCalledWith('/api/sync/push', expect.objectContaining({
        events: expect.arrayContaining([
          expect.objectContaining({
            type: 'feeding',
            action: 'create',
            entity_id: 'feeding-1',
          }),
        ]),
      }))

      // Pending event should be removed
      const remaining = await getPendingEvents()
      expect(remaining).toHaveLength(0)

      // Feeding should be marked as synced
      const feeding = await db.feedings.get('feeding-1')
      expect(feeding?.pendingSync).toBe(false)
      expect(feeding?.syncedAt).toBeDefined()
    })

    it('should handle API failure and increment retry count', async () => {
      await addPendingEvent('feeding', 'create', 'feeding-1', {})

      // Mock API failure
      vi.mocked(apiClient.post).mockRejectedValueOnce(new Error('Network error'))

      const result = await syncPendingEvents()

      expect(result.synced).toBe(0)
      expect(result.failed).toBe(1)

      // Event should still be pending with incremented retry count
      const events = await getPendingEvents()
      expect(events).toHaveLength(1)
      expect(events[0]?.retryCount).toBe(1)
    })

    it('should skip events that exceeded max retries', async () => {
      // Manually add an event with high retry count
      await db.pendingEvents.add({
        id: 'event-1',
        type: 'feeding',
        action: 'create',
        entityId: 'feeding-1',
        timestamp: new Date(),
        retryCount: 3, // MAX_RETRIES
      })

      const result = await syncPendingEvents()

      expect(result.synced).toBe(0)
      expect(result.failed).toBe(1)

      // API should not be called
      expect(apiClient.post).not.toHaveBeenCalled()

      // Event should still be there
      const events = await getPendingEvents()
      expect(events).toHaveLength(1)
    })

    it('should sync multiple events in order', async () => {
      await addPendingEvent('feeding', 'create', 'feeding-1', { order: 1 })
      await addPendingEvent('sleep', 'create', 'sleep-1', { order: 2 })
      await addPendingEvent('note', 'create', 'note-1', { order: 3 })

      // Add entities to local DB
      await db.feedings.add({
        id: 'feeding-1',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date(),
        pendingSync: true,
      })
      await db.sleep.add({
        id: 'sleep-1',
        childId: 'child-1',
        type: 'nap',
        startTime: new Date(),
        pendingSync: true,
      })
      await db.notes.add({
        id: 'note-1',
        childId: 'child-1',
        authorId: 'user-1',
        content: 'Test note',
        pinned: false,
        pendingSync: true,
      })

      vi.mocked(apiClient.post).mockResolvedValue({ data: {}, status: 200 })

      const result = await syncPendingEvents()

      expect(result.synced).toBe(3)
      expect(result.failed).toBe(0)
      expect(apiClient.post).toHaveBeenCalledTimes(3)
    })

    it('should return zeros when no pending events', async () => {
      const result = await syncPendingEvents()

      expect(result.synced).toBe(0)
      expect(result.failed).toBe(0)
      expect(apiClient.post).not.toHaveBeenCalled()
    })

    it('should handle partial sync failure', async () => {
      await addPendingEvent('feeding', 'create', 'feeding-1', {})
      await addPendingEvent('sleep', 'create', 'sleep-1', {})

      await db.feedings.add({
        id: 'feeding-1',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date(),
        pendingSync: true,
      })
      await db.sleep.add({
        id: 'sleep-1',
        childId: 'child-1',
        type: 'nap',
        startTime: new Date(),
        pendingSync: true,
      })

      // First call succeeds, second fails
      vi.mocked(apiClient.post)
        .mockResolvedValueOnce({ data: {}, status: 200 })
        .mockRejectedValueOnce(new Error('Network error'))

      const result = await syncPendingEvents()

      expect(result.synced).toBe(1)
      expect(result.failed).toBe(1)

      // Only one event should remain (the failed one)
      const events = await getPendingEvents()
      expect(events).toHaveLength(1)
    })
  })

  describe('markEntityAsSynced', () => {
    it('should mark feeding as synced', async () => {
      await db.feedings.add({
        id: 'feeding-1',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date(),
        pendingSync: true,
      })

      await addPendingEvent('feeding', 'create', 'feeding-1', {})
      vi.mocked(apiClient.post).mockResolvedValueOnce({ data: {}, status: 200 })

      await syncPendingEvents()

      const feeding = await db.feedings.get('feeding-1')
      expect(feeding?.pendingSync).toBe(false)
      expect(feeding?.syncedAt).toBeDefined()
    })

    it('should mark sleep as synced', async () => {
      await db.sleep.add({
        id: 'sleep-1',
        childId: 'child-1',
        type: 'nap',
        startTime: new Date(),
        pendingSync: true,
      })

      await addPendingEvent('sleep', 'update', 'sleep-1', {})
      vi.mocked(apiClient.post).mockResolvedValueOnce({ data: {}, status: 200 })

      await syncPendingEvents()

      const sleep = await db.sleep.get('sleep-1')
      expect(sleep?.pendingSync).toBe(false)
      expect(sleep?.syncedAt).toBeDefined()
    })

    it('should mark medication as synced', async () => {
      await db.medications.add({
        id: 'med-1',
        childId: 'child-1',
        name: 'Vitamin D',
        dosage: '1',
        unit: 'ml',
        frequency: 'daily',
        startDate: new Date(),
        active: true,
        pendingSync: true,
      })

      await addPendingEvent('medication', 'create', 'med-1', {})
      vi.mocked(apiClient.post).mockResolvedValueOnce({ data: {}, status: 200 })

      await syncPendingEvents()

      const med = await db.medications.get('med-1')
      expect(med?.pendingSync).toBe(false)
    })

    it('should mark note as synced', async () => {
      await db.notes.add({
        id: 'note-1',
        childId: 'child-1',
        authorId: 'user-1',
        content: 'Test note',
        pinned: false,
        pendingSync: true,
      })

      await addPendingEvent('note', 'create', 'note-1', {})
      vi.mocked(apiClient.post).mockResolvedValueOnce({ data: {}, status: 200 })

      await syncPendingEvents()

      const note = await db.notes.get('note-1')
      expect(note?.pendingSync).toBe(false)
    })

    it('should mark vaccination as synced', async () => {
      await db.vaccinations.add({
        id: 'vax-1',
        childId: 'child-1',
        name: 'MMR',
        dose: 1,
        scheduledAt: new Date(),
        completed: false,
        pendingSync: true,
      })

      await addPendingEvent('vaccination', 'create', 'vax-1', {})
      vi.mocked(apiClient.post).mockResolvedValueOnce({ data: {}, status: 200 })

      await syncPendingEvents()

      const vax = await db.vaccinations.get('vax-1')
      expect(vax?.pendingSync).toBe(false)
    })

    it('should mark appointment as synced', async () => {
      await db.appointments.add({
        id: 'apt-1',
        childId: 'child-1',
        type: 'well_visit',
        title: 'Checkup',
        scheduledAt: new Date(),
        duration: 30,
        completed: false,
        cancelled: false,
        pendingSync: true,
      })

      await addPendingEvent('appointment', 'create', 'apt-1', {})
      vi.mocked(apiClient.post).mockResolvedValueOnce({ data: {}, status: 200 })

      await syncPendingEvents()

      const apt = await db.appointments.get('apt-1')
      expect(apt?.pendingSync).toBe(false)
    })

    it('should mark medication_log as synced', async () => {
      await db.medicationLogs.add({
        id: 'log-1',
        medicationId: 'med-1',
        childId: 'child-1',
        givenAt: new Date(),
        givenBy: 'user-1',
        dosage: '5ml',
        pendingSync: true,
      })

      await addPendingEvent('medication_log', 'create', 'log-1', {})
      vi.mocked(apiClient.post).mockResolvedValueOnce({ data: {}, status: 200 })

      await syncPendingEvents()

      const log = await db.medicationLogs.get('log-1')
      expect(log?.pendingSync).toBe(false)
      expect(log?.syncedAt).toBeDefined()
    })
  })

  describe('pullFromServer', () => {
    it('should pull events from server and apply them', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'feeding',
              action: 'create',
              entity_id: 'feeding-server-1',
              data: {
                childId: 'child-1',
                type: 'bottle',
                startTime: new Date().toISOString(),
              },
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      expect(apiClient.get).toHaveBeenCalledWith('/api/sync/pull', {
        params: { last_sync: undefined },
      })

      const feeding = await db.feedings.get('feeding-server-1')
      expect(feeding).toBeDefined()
      expect(feeding?.pendingSync).toBe(false)
    })

    it('should pass lastSync parameter when provided', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: { events: [] },
        status: 200,
      })

      await pullFromServer('2024-01-01T00:00:00Z')

      expect(apiClient.get).toHaveBeenCalledWith('/api/sync/pull', {
        params: { last_sync: '2024-01-01T00:00:00Z' },
      })
    })

    it('should handle API errors gracefully', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      vi.mocked(apiClient.get).mockRejectedValueOnce(new Error('Network error'))

      await pullFromServer()

      expect(consoleSpy).toHaveBeenCalledWith('Failed to pull from server:', expect.any(Error))
      consoleSpy.mockRestore()
    })

    it('should apply delete action for feeding', async () => {
      await db.feedings.add({
        id: 'feeding-to-delete',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date(),
        pendingSync: false,
      })

      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'feeding',
              action: 'delete',
              entity_id: 'feeding-to-delete',
              data: null,
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const feeding = await db.feedings.get('feeding-to-delete')
      expect(feeding).toBeUndefined()
    })

    it('should apply create/update action for sleep', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'sleep',
              action: 'create',
              entity_id: 'sleep-server-1',
              data: {
                childId: 'child-1',
                type: 'nap',
                startTime: new Date().toISOString(),
              },
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const sleep = await db.sleep.get('sleep-server-1')
      expect(sleep).toBeDefined()
      expect(sleep?.pendingSync).toBe(false)
    })

    it('should apply delete action for sleep', async () => {
      await db.sleep.add({
        id: 'sleep-to-delete',
        childId: 'child-1',
        type: 'nap',
        startTime: new Date(),
        pendingSync: false,
      })

      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'sleep',
              action: 'delete',
              entity_id: 'sleep-to-delete',
              data: null,
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const sleep = await db.sleep.get('sleep-to-delete')
      expect(sleep).toBeUndefined()
    })

    it('should apply create/update action for medication', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'medication',
              action: 'create',
              entity_id: 'med-server-1',
              data: {
                childId: 'child-1',
                name: 'Vitamin D',
                dosage: '1',
                unit: 'ml',
                frequency: 'daily',
                active: true,
              },
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const med = await db.medications.get('med-server-1')
      expect(med).toBeDefined()
      expect(med?.pendingSync).toBe(false)
    })

    it('should apply delete action for medication', async () => {
      await db.medications.add({
        id: 'med-to-delete',
        childId: 'child-1',
        name: 'Test',
        dosage: '1',
        unit: 'ml',
        frequency: 'daily',
        startDate: new Date(),
        active: true,
        pendingSync: false,
      })

      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'medication',
              action: 'delete',
              entity_id: 'med-to-delete',
              data: null,
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const med = await db.medications.get('med-to-delete')
      expect(med).toBeUndefined()
    })

    it('should apply create/update action for note', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'note',
              action: 'create',
              entity_id: 'note-server-1',
              data: {
                childId: 'child-1',
                authorId: 'user-1',
                content: 'Test note from server',
                pinned: false,
              },
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const note = await db.notes.get('note-server-1')
      expect(note).toBeDefined()
      expect(note?.pendingSync).toBe(false)
    })

    it('should apply delete action for note', async () => {
      await db.notes.add({
        id: 'note-to-delete',
        childId: 'child-1',
        authorId: 'user-1',
        content: 'Test',
        pinned: false,
        pendingSync: false,
      })

      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'note',
              action: 'delete',
              entity_id: 'note-to-delete',
              data: null,
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const note = await db.notes.get('note-to-delete')
      expect(note).toBeUndefined()
    })

    it('should apply create/update action for vaccination', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'vaccination',
              action: 'create',
              entity_id: 'vax-server-1',
              data: {
                childId: 'child-1',
                name: 'MMR',
                dose: 1,
                scheduledAt: new Date().toISOString(),
                completed: false,
              },
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const vax = await db.vaccinations.get('vax-server-1')
      expect(vax).toBeDefined()
      expect(vax?.pendingSync).toBe(false)
    })

    it('should apply delete action for vaccination', async () => {
      await db.vaccinations.add({
        id: 'vax-to-delete',
        childId: 'child-1',
        name: 'Test',
        dose: 1,
        scheduledAt: new Date(),
        completed: false,
        pendingSync: false,
      })

      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'vaccination',
              action: 'delete',
              entity_id: 'vax-to-delete',
              data: null,
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const vax = await db.vaccinations.get('vax-to-delete')
      expect(vax).toBeUndefined()
    })

    it('should apply create/update action for appointment', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'appointment',
              action: 'create',
              entity_id: 'apt-server-1',
              data: {
                childId: 'child-1',
                type: 'well_visit',
                title: 'Checkup',
                scheduledAt: new Date().toISOString(),
                duration: 30,
                completed: false,
                cancelled: false,
              },
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const apt = await db.appointments.get('apt-server-1')
      expect(apt).toBeDefined()
      expect(apt?.pendingSync).toBe(false)
    })

    it('should apply delete action for appointment', async () => {
      await db.appointments.add({
        id: 'apt-to-delete',
        childId: 'child-1',
        type: 'well_visit',
        title: 'Test',
        scheduledAt: new Date(),
        duration: 30,
        completed: false,
        cancelled: false,
        pendingSync: false,
      })

      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'appointment',
              action: 'delete',
              entity_id: 'apt-to-delete',
              data: null,
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const apt = await db.appointments.get('apt-to-delete')
      expect(apt).toBeUndefined()
    })

    it('should apply multiple events in sequence', async () => {
      vi.mocked(apiClient.get).mockResolvedValueOnce({
        data: {
          events: [
            {
              type: 'feeding',
              action: 'create',
              entity_id: 'feeding-1',
              data: { childId: 'child-1', type: 'bottle', startTime: new Date().toISOString() },
            },
            {
              type: 'sleep',
              action: 'create',
              entity_id: 'sleep-1',
              data: { childId: 'child-1', type: 'nap', startTime: new Date().toISOString() },
            },
          ],
        },
        status: 200,
      })

      await pullFromServer()

      const feeding = await db.feedings.get('feeding-1')
      const sleep = await db.sleep.get('sleep-1')
      expect(feeding).toBeDefined()
      expect(sleep).toBeDefined()
    })
  })
})
