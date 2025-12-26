export interface Vaccination {
  id: string
  childId: string
  name: string
  dose: number
  scheduledAt: string
  administeredAt?: string
  provider?: string
  location?: string
  lotNumber?: string
  notes?: string
  completed: boolean
  createdAt: string
  updatedAt: string
}

export interface VaccinationSchedule {
  id: string
  name: string
  description: string
  ageMonths: number
  dose: number
}

export interface CreateVaccinationRequest {
  child_id: string
  name: string
  dose: number
  scheduled_at: string
}

export interface RecordVaccinationRequest {
  administered_at: string
  provider?: string
  location?: string
  lot_number?: string
  notes?: string
}
