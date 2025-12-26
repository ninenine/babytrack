import { useState } from 'react'
import { Plus, Search, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { NoteForm, NoteCard } from '@/components/notes'
import { useNotes } from '@/hooks'
import { useFamilyStore } from '@/stores/family.store'

export function NotesPage() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const [searchQuery, setSearchQuery] = useState('')
  const { notes, isLoading, isSyncing } = useNotes(searchQuery)
  const [showForm, setShowForm] = useState(false)

  if (!currentChild) {
    return (
      <div className="flex items-center justify-center h-[50vh] text-muted-foreground">
        No child selected
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h1 className="text-2xl font-bold">Notes</h1>
          {isSyncing && (
            <RefreshCw className="h-4 w-4 animate-spin text-muted-foreground" />
          )}
        </div>
        <Button size="sm" onClick={() => setShowForm(true)}>
          <Plus className="h-4 w-4 mr-1" />
          Add
        </Button>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search notes..."
          className="pl-9"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
      </div>

      {isLoading ? (
        <div className="space-y-3">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
      ) : notes.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          {searchQuery ? (
            <p>No notes matching "{searchQuery}"</p>
          ) : (
            <>
              <p>No notes yet</p>
              <p className="text-sm mt-1">Tap + Add to create a note</p>
            </>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {notes.map((note) => (
            <NoteCard key={note.id} note={note} />
          ))}
        </div>
      )}

      <NoteForm open={showForm} onOpenChange={setShowForm} />
    </div>
  )
}
