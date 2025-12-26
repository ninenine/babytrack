export type MedicationFrequency =
  | 'once_daily'
  | 'twice_daily'
  | 'three_times_daily'
  | 'four_times_daily'
  | 'every_4_hours'
  | 'every_6_hours'
  | 'every_8_hours'
  | 'as_needed'

export interface Medication {
  id: string
  childId: string
  name: string
  dosage: string
  unit: string
  frequency: MedicationFrequency
  instructions?: string
  startDate: string
  endDate?: string
  active: boolean
  createdAt: string
  updatedAt: string
}

export interface MedicationLog {
  id: string
  medicationId: string
  childId: string
  givenAt: string
  givenBy: string
  dosage: string
  notes?: string
  createdAt: string
}

export interface CreateMedicationRequest {
  child_id: string
  name: string
  dosage: string
  unit: string
  frequency: MedicationFrequency
  instructions?: string
  start_date: string
}

export interface LogDoseRequest {
  medication_id: string
  given_at: string
  dosage: string
  notes?: string
}
