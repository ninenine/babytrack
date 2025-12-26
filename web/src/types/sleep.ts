export type SleepType = 'nap' | 'night'

export interface Sleep {
  id: string
  childId: string
  type: SleepType
  startTime: string
  endTime?: string
  quality?: number
  notes?: string
  createdAt: string
  updatedAt: string
}

export interface CreateSleepRequest {
  child_id: string
  type: SleepType
  start_time: string
  end_time?: string
  quality?: number
  notes?: string
}
