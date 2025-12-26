export interface Note {
  id: string
  childId: string
  authorId: string
  title?: string
  content: string
  tags?: string[]
  pinned: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateNoteRequest {
  child_id: string
  title?: string
  content: string
  tags?: string[]
}

export interface UpdateNoteRequest {
  title?: string
  content?: string
  tags?: string[]
  pinned?: boolean
}
