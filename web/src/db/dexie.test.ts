import { describe, it, expect, beforeEach } from 'vitest'
import { db, type LocalFeeding, type LocalSleep, type LocalNote } from './dexie'

describe('Dexie Database', () => {
  beforeEach(async () => {
    // Clear all tables before each test
    await db.feedings.clear()
    await db.sleep.clear()
    await db.medications.clear()
    await db.medicationLogs.clear()
    await db.notes.clear()
    await db.vaccinations.clear()
    await db.appointments.clear()
    await db.pendingEvents.clear()
  })

  describe('Feedings', () => {
    it('should add and retrieve a feeding', async () => {
      const feeding: LocalFeeding = {
        id: 'feeding-1',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date('2024-01-01T10:00:00'),
        endTime: new Date('2024-01-01T10:30:00'),
        amount: 120,
        unit: 'ml',
        pendingSync: false,
      }

      await db.feedings.add(feeding)

      const retrieved = await db.feedings.get('feeding-1')
      expect(retrieved).toEqual(feeding)
    })

    it('should query feedings by childId', async () => {
      await db.feedings.add({
        id: 'feeding-1',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date(),
        pendingSync: false,
      })
      await db.feedings.add({
        id: 'feeding-2',
        childId: 'child-2',
        type: 'breast',
        startTime: new Date(),
        pendingSync: false,
      })
      await db.feedings.add({
        id: 'feeding-3',
        childId: 'child-1',
        type: 'formula',
        startTime: new Date(),
        pendingSync: false,
      })

      const child1Feedings = await db.feedings.where('childId').equals('child-1').toArray()
      expect(child1Feedings).toHaveLength(2)
    })

    it('should query pending sync feedings', async () => {
      await db.feedings.add({
        id: 'feeding-1',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date(),
        pendingSync: true,
      })
      await db.feedings.add({
        id: 'feeding-2',
        childId: 'child-1',
        type: 'breast',
        startTime: new Date(),
        pendingSync: false,
      })

      const pending = await db.feedings.filter((f) => f.pendingSync === true).toArray()
      expect(pending).toHaveLength(1)
      expect(pending[0]?.id).toBe('feeding-1')
    })

    it('should update a feeding', async () => {
      await db.feedings.add({
        id: 'feeding-1',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date(),
        amount: 100,
        pendingSync: true,
      })

      await db.feedings.update('feeding-1', {
        amount: 150,
        pendingSync: false,
        syncedAt: new Date(),
      })

      const updated = await db.feedings.get('feeding-1')
      expect(updated?.amount).toBe(150)
      expect(updated?.pendingSync).toBe(false)
      expect(updated?.syncedAt).toBeDefined()
    })

    it('should delete a feeding', async () => {
      await db.feedings.add({
        id: 'feeding-1',
        childId: 'child-1',
        type: 'bottle',
        startTime: new Date(),
        pendingSync: false,
      })

      await db.feedings.delete('feeding-1')

      const deleted = await db.feedings.get('feeding-1')
      expect(deleted).toBeUndefined()
    })

    it('should support all feeding types', async () => {
      const types: LocalFeeding['type'][] = ['breast', 'bottle', 'formula', 'solid']

      for (const type of types) {
        await db.feedings.add({
          id: `feeding-${type}`,
          childId: 'child-1',
          type,
          startTime: new Date(),
          pendingSync: false,
        })
      }

      const all = await db.feedings.toArray()
      expect(all).toHaveLength(4)
    })
  })

  describe('Sleep', () => {
    it('should add and retrieve sleep records', async () => {
      const sleep: LocalSleep = {
        id: 'sleep-1',
        childId: 'child-1',
        type: 'nap',
        startTime: new Date('2024-01-01T14:00:00'),
        endTime: new Date('2024-01-01T15:30:00'),
        quality: 4,
        pendingSync: false,
      }

      await db.sleep.add(sleep)

      const retrieved = await db.sleep.get('sleep-1')
      expect(retrieved).toEqual(sleep)
    })

    it('should support nap and night sleep types', async () => {
      await db.sleep.add({
        id: 'sleep-nap',
        childId: 'child-1',
        type: 'nap',
        startTime: new Date(),
        pendingSync: false,
      })
      await db.sleep.add({
        id: 'sleep-night',
        childId: 'child-1',
        type: 'night',
        startTime: new Date(),
        pendingSync: false,
      })

      const naps = await db.sleep.where('type').equals('nap').toArray()
      const nights = await db.sleep.where('type').equals('night').toArray()

      expect(naps).toHaveLength(1)
      expect(nights).toHaveLength(1)
    })
  })

  describe('Notes', () => {
    it('should add and retrieve notes', async () => {
      const note: LocalNote = {
        id: 'note-1',
        childId: 'child-1',
        authorId: 'user-1',
        title: 'First Steps',
        content: 'Baby took their first steps today!',
        tags: ['milestone', 'walking'],
        pinned: true,
        pendingSync: false,
      }

      await db.notes.add(note)

      const retrieved = await db.notes.get('note-1')
      expect(retrieved).toEqual(note)
    })

    it('should query pinned notes', async () => {
      await db.notes.add({
        id: 'note-1',
        childId: 'child-1',
        authorId: 'user-1',
        content: 'Pinned note',
        pinned: true,
        pendingSync: false,
      })
      await db.notes.add({
        id: 'note-2',
        childId: 'child-1',
        authorId: 'user-1',
        content: 'Regular note',
        pinned: false,
        pendingSync: false,
      })

      const pinned = await db.notes.filter((n) => n.pinned === true).toArray()
      expect(pinned).toHaveLength(1)
      expect(pinned[0]?.id).toBe('note-1')
    })
  })

  describe('Medications', () => {
    it('should add and retrieve medications', async () => {
      await db.medications.add({
        id: 'med-1',
        childId: 'child-1',
        name: 'Vitamin D',
        dosage: '400',
        unit: 'IU',
        frequency: 'daily',
        startDate: new Date(),
        active: true,
        pendingSync: false,
      })

      const med = await db.medications.get('med-1')
      expect(med?.name).toBe('Vitamin D')
    })

    it('should query active medications', async () => {
      await db.medications.add({
        id: 'med-1',
        childId: 'child-1',
        name: 'Active Med',
        dosage: '1',
        unit: 'ml',
        frequency: 'daily',
        startDate: new Date(),
        active: true,
        pendingSync: false,
      })
      await db.medications.add({
        id: 'med-2',
        childId: 'child-1',
        name: 'Inactive Med',
        dosage: '1',
        unit: 'ml',
        frequency: 'daily',
        startDate: new Date(),
        active: false,
        pendingSync: false,
      })

      const active = await db.medications.filter((m) => m.active === true).toArray()
      expect(active).toHaveLength(1)
      expect(active[0]?.name).toBe('Active Med')
    })
  })

  describe('Vaccinations', () => {
    it('should add and retrieve vaccinations', async () => {
      await db.vaccinations.add({
        id: 'vax-1',
        childId: 'child-1',
        name: 'MMR',
        dose: 1,
        scheduledAt: new Date('2024-06-01'),
        completed: false,
        pendingSync: false,
      })

      const vax = await db.vaccinations.get('vax-1')
      expect(vax?.name).toBe('MMR')
      expect(vax?.dose).toBe(1)
    })

    it('should track completed vaccinations', async () => {
      await db.vaccinations.add({
        id: 'vax-1',
        childId: 'child-1',
        name: 'Completed Vax',
        dose: 1,
        scheduledAt: new Date(),
        administeredAt: new Date(),
        completed: true,
        pendingSync: false,
      })
      await db.vaccinations.add({
        id: 'vax-2',
        childId: 'child-1',
        name: 'Pending Vax',
        dose: 1,
        scheduledAt: new Date(),
        completed: false,
        pendingSync: false,
      })

      const completed = await db.vaccinations.filter((v) => v.completed === true).toArray()
      expect(completed).toHaveLength(1)
      expect(completed[0]?.name).toBe('Completed Vax')
    })
  })

  describe('Appointments', () => {
    it('should add and retrieve appointments', async () => {
      await db.appointments.add({
        id: 'apt-1',
        childId: 'child-1',
        type: 'well_visit',
        title: '6 Month Checkup',
        provider: 'Dr. Smith',
        location: 'Pediatric Clinic',
        scheduledAt: new Date('2024-06-15T10:00:00'),
        duration: 30,
        completed: false,
        cancelled: false,
        pendingSync: false,
      })

      const apt = await db.appointments.get('apt-1')
      expect(apt?.title).toBe('6 Month Checkup')
      expect(apt?.type).toBe('well_visit')
    })

    it('should support all appointment types', async () => {
      const types: Array<'well_visit' | 'sick_visit' | 'specialist' | 'dental' | 'other'> = [
        'well_visit',
        'sick_visit',
        'specialist',
        'dental',
        'other',
      ]

      for (const type of types) {
        await db.appointments.add({
          id: `apt-${type}`,
          childId: 'child-1',
          type,
          title: `${type} appointment`,
          scheduledAt: new Date(),
          duration: 30,
          completed: false,
          cancelled: false,
          pendingSync: false,
        })
      }

      const all = await db.appointments.toArray()
      expect(all).toHaveLength(5)
    })
  })

  describe('Bulk Operations', () => {
    it('should handle bulk add', async () => {
      const feedings: LocalFeeding[] = Array.from({ length: 100 }, (_, i) => ({
        id: `feeding-${i}`,
        childId: 'child-1',
        type: 'bottle' as const,
        startTime: new Date(),
        pendingSync: false,
      }))

      await db.feedings.bulkAdd(feedings)

      const count = await db.feedings.count()
      expect(count).toBe(100)
    })

    it('should handle bulk delete', async () => {
      const ids = ['feeding-1', 'feeding-2', 'feeding-3']

      for (const id of ids) {
        await db.feedings.add({
          id,
          childId: 'child-1',
          type: 'bottle',
          startTime: new Date(),
          pendingSync: false,
        })
      }

      await db.feedings.bulkDelete(ids)

      const count = await db.feedings.count()
      expect(count).toBe(0)
    })
  })
})
