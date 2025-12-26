export type FeedingType = 'breast' | 'bottle' | 'formula' | 'solid'

export interface Feeding {
  id: string
  childId: string
  type: FeedingType
  startTime: string
  endTime?: string
  amount?: number
  unit?: string
  side?: 'left' | 'right' | 'both'
  notes?: string
  createdAt: string
  updatedAt: string
}

export interface CreateFeedingRequest {
  child_id: string
  type: FeedingType
  start_time: string
  end_time?: string
  amount?: number
  unit?: string
  side?: string
  notes?: string
}
