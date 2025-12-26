import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useLiveQuery } from 'dexie-react-hooks'
import { db, type LocalNote } from '@/db/dexie'
import { addPendingEvent } from '@/db/events'
import { apiClient } from '@/lib/api-client'
import { queryKeys } from '@/lib/query-client'
import { useFamilyStore } from '@/stores/family.store'
import { useSessionStore } from '@/stores/session.store'

export interface CreateNoteInput {
  title?: string
  content: string
  tags?: string[]
  pinned?: boolean
}

export interface UpdateNoteInput {
  id: string
  title?: string
  content?: string
  tags?: string[]
  pinned?: boolean
}

interface NoteResponse {
  id: string
  child_id: string
  author_id: string
  title?: string
  content: string
  tags?: string[]
  pinned: boolean
  created_at: string
  updated_at: string
}

function mapResponseToLocal(response: NoteResponse): LocalNote {
  return {
    id: response.id,
    childId: response.child_id,
    authorId: response.author_id,
    title: response.title,
    content: response.content,
    tags: response.tags,
    pinned: response.pinned,
    syncedAt: new Date(),
    pendingSync: false,
  }
}

export function useNotes(searchQuery?: string) {
  const currentChild = useFamilyStore((state) => state.currentChild)

  const notes = useLiveQuery(
    () =>
      currentChild
        ? db.notes.where('childId').equals(currentChild.id).reverse().toArray()
        : [],
    [currentChild?.id]
  )

  const syncQuery = useQuery({
    queryKey: queryKeys.notes.byChild(currentChild?.id ?? ''),
    queryFn: async () => {
      if (!currentChild) return []

      const response = await apiClient.get<NoteResponse[]>('/api/notes', {
        params: { child_id: currentChild.id },
      })

      for (const note of response.data) {
        await db.notes.put(mapResponseToLocal(note))
      }

      return response.data
    },
    enabled: !!currentChild,
    staleTime: 30000,
  })

  // Filter notes based on search query
  const filteredNotes = notes?.filter((note) => {
    if (!searchQuery) return true
    const query = searchQuery.toLowerCase()
    return (
      note.title?.toLowerCase().includes(query) ||
      note.content.toLowerCase().includes(query) ||
      note.tags?.some((tag) => tag.toLowerCase().includes(query))
    )
  })

  // Sort: pinned first, then by most recent
  const sortedNotes = filteredNotes?.sort((a, b) => {
    if (a.pinned && !b.pinned) return -1
    if (!a.pinned && b.pinned) return 1
    return 0
  })

  const pinnedNotes = sortedNotes?.filter((n) => n.pinned) ?? []
  const unpinnedNotes = sortedNotes?.filter((n) => !n.pinned) ?? []

  return {
    notes: sortedNotes ?? [],
    pinnedNotes,
    unpinnedNotes,
    isLoading: notes === undefined,
    isSyncing: syncQuery.isFetching,
    error: syncQuery.error,
    refetch: syncQuery.refetch,
  }
}

export function useCreateNote() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)
  const user = useSessionStore((state) => state.user)

  return useMutation({
    mutationFn: async (input: CreateNoteInput) => {
      if (!currentChild) throw new Error('No child selected')
      if (!user) throw new Error('Not authenticated')

      const payload = {
        child_id: currentChild.id,
        title: input.title,
        content: input.content,
        tags: input.tags,
        pinned: input.pinned ?? false,
      }

      try {
        const response = await apiClient.post<NoteResponse>('/api/notes', payload)
        await db.notes.add(mapResponseToLocal(response.data))
        return response.data
      } catch {
        const localId = crypto.randomUUID()
        const localNote: LocalNote = {
          id: localId,
          childId: currentChild.id,
          authorId: user.id,
          title: input.title,
          content: input.content,
          tags: input.tags,
          pinned: input.pinned ?? false,
          pendingSync: true,
        }
        await db.notes.add(localNote)
        await addPendingEvent('note', 'create', localId, payload)
        return localNote
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notes.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useUpdateNote() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (input: UpdateNoteInput) => {
      const payload: Record<string, unknown> = {}
      if (input.title !== undefined) payload.title = input.title
      if (input.content !== undefined) payload.content = input.content
      if (input.tags !== undefined) payload.tags = input.tags
      if (input.pinned !== undefined) payload.pinned = input.pinned

      try {
        const response = await apiClient.patch<NoteResponse>(`/api/notes/${input.id}`, payload)
        await db.notes.update(input.id, {
          title: input.title,
          content: input.content,
          tags: input.tags,
          pinned: input.pinned,
          syncedAt: new Date(),
          pendingSync: false,
        })
        return response.data
      } catch {
        await db.notes.update(input.id, {
          title: input.title,
          content: input.content,
          tags: input.tags,
          pinned: input.pinned,
          pendingSync: true,
        })
        await addPendingEvent('note', 'update', input.id, payload)
        return null
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notes.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useDeleteNote() {
  const queryClient = useQueryClient()
  const currentChild = useFamilyStore((state) => state.currentChild)

  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await apiClient.delete(`/api/notes/${id}`)
      } catch {
        await addPendingEvent('note', 'delete', id)
      }
      await db.notes.delete(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notes.byChild(currentChild?.id ?? '') })
    },
  })
}

export function useToggleNotePin() {
  const updateNote = useUpdateNote()

  return useMutation({
    mutationFn: async ({ id, pinned }: { id: string; pinned: boolean }) => {
      return updateNote.mutateAsync({ id, pinned })
    },
  })
}
