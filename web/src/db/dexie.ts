import Dexie, { type Table } from 'dexie'

export interface LocalFeeding {
  id: string
  childId: string
  type: 'breast' | 'bottle' | 'formula' | 'solid'
  startTime: Date
  endTime?: Date
  amount?: number
  unit?: string
  side?: string
  notes?: string
  syncedAt?: Date
  pendingSync: boolean
}

export interface LocalSleep {
  id: string
  childId: string
  type: 'nap' | 'night'
  startTime: Date
  endTime?: Date
  quality?: number
  notes?: string
  syncedAt?: Date
  pendingSync: boolean
}

export interface LocalMedication {
  id: string
  childId: string
  name: string
  dosage: string
  unit: string
  frequency: string
  instructions?: string
  startDate: Date
  endDate?: Date
  active: boolean
  syncedAt?: Date
  pendingSync: boolean
}

export interface LocalMedicationLog {
  id: string
  medicationId: string
  childId: string
  givenAt: Date
  givenBy: string
  dosage: string
  notes?: string
  syncedAt?: Date
  pendingSync: boolean
}

export interface LocalNote {
  id: string
  childId: string
  authorId: string
  title?: string
  content: string
  tags?: string[]
  pinned: boolean
  syncedAt?: Date
  pendingSync: boolean
}

export interface LocalVaccination {
  id: string
  childId: string
  name: string
  dose: number
  scheduledAt: Date
  administeredAt?: Date
  provider?: string
  location?: string
  lotNumber?: string
  notes?: string
  completed: boolean
  syncedAt?: Date
  pendingSync: boolean
}

export interface LocalAppointment {
  id: string
  childId: string
  type: 'well_visit' | 'sick_visit' | 'specialist' | 'dental' | 'other'
  title: string
  provider?: string
  location?: string
  scheduledAt: Date
  duration: number
  notes?: string
  completed: boolean
  cancelled: boolean
  syncedAt?: Date
  pendingSync: boolean
}

export interface PendingEvent {
  id: string
  type: 'feeding' | 'sleep' | 'medication' | 'medication_log' | 'note' | 'vaccination' | 'appointment'
  action: 'create' | 'update' | 'delete'
  entityId: string
  data?: unknown
  timestamp: Date
  retryCount: number
}

class FamilyTrackerDB extends Dexie {
  feedings!: Table<LocalFeeding>
  sleep!: Table<LocalSleep>
  medications!: Table<LocalMedication>
  medicationLogs!: Table<LocalMedicationLog>
  notes!: Table<LocalNote>
  vaccinations!: Table<LocalVaccination>
  appointments!: Table<LocalAppointment>
  pendingEvents!: Table<PendingEvent>

  constructor() {
    super('FamilyTrackerDB')

    this.version(1).stores({
      feedings: 'id, childId, startTime, type, pendingSync',
      sleep: 'id, childId, startTime, type, pendingSync',
      medications: 'id, childId, name, active, pendingSync',
      medicationLogs: 'id, medicationId, childId, givenAt, pendingSync',
      notes: 'id, childId, authorId, pinned, pendingSync',
      pendingEvents: 'id, type, entityId, timestamp',
    })

    this.version(2).stores({
      feedings: 'id, childId, startTime, type, pendingSync',
      sleep: 'id, childId, startTime, type, pendingSync',
      medications: 'id, childId, name, active, pendingSync',
      medicationLogs: 'id, medicationId, childId, givenAt, pendingSync',
      notes: 'id, childId, authorId, pinned, pendingSync',
      vaccinations: 'id, childId, name, scheduledAt, completed, pendingSync',
      appointments: 'id, childId, type, scheduledAt, completed, cancelled, pendingSync',
      pendingEvents: 'id, type, entityId, timestamp',
    })
  }
}

export const db = new FamilyTrackerDB()
