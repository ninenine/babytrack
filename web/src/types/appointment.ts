export type AppointmentType = 'well_visit' | 'sick_visit' | 'specialist' | 'dental' | 'other'

export interface Appointment {
  id: string
  childId: string
  type: AppointmentType
  title: string
  provider?: string
  location?: string
  scheduledAt: string
  duration: number
  notes?: string
  completed: boolean
  cancelled: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateAppointmentRequest {
  child_id: string
  type: AppointmentType
  title: string
  provider?: string
  location?: string
  scheduled_at: string
  duration: number
  notes?: string
}
