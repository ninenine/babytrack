import { Pin, Trash2 } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { useDeleteNote, useToggleNotePin } from '@/hooks'
import type { LocalNote } from '@/db/dexie'

interface NoteCardProps {
  note: LocalNote
}

export function NoteCard({ note }: NoteCardProps) {
  const deleteNote = useDeleteNote()
  const togglePin = useToggleNotePin()

  const handleDelete = () => {
    deleteNote.mutate(note.id)
  }

  const handleTogglePin = async () => {
    await togglePin.mutateAsync({ id: note.id, pinned: !note.pinned })
  }

  return (
    <Card className={note.pinned ? 'border-primary' : ''}>
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            {note.title && (
              <div className="flex items-center gap-2">
                {note.pinned && <Pin className="h-4 w-4 text-primary" />}
                <h3 className="font-medium truncate">{note.title}</h3>
                {note.pendingSync && (
                  <Badge variant="outline" className="text-xs text-yellow-600">
                    Pending
                  </Badge>
                )}
              </div>
            )}
            <p className="text-sm text-muted-foreground mt-1 whitespace-pre-wrap line-clamp-3">
              {note.content}
            </p>
            {note.tags && note.tags.length > 0 && (
              <div className="flex flex-wrap gap-1 mt-2">
                {note.tags.map((tag) => (
                  <Badge key={tag} variant="secondary" className="text-xs">
                    {tag}
                  </Badge>
                ))}
              </div>
            )}
          </div>

          <div className="flex flex-col gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={handleTogglePin}
              disabled={togglePin.isPending}
            >
              <Pin className={`h-4 w-4 ${note.pinned ? 'fill-primary text-primary' : 'text-muted-foreground'}`} />
            </Button>

            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8">
                  <Trash2 className="h-4 w-4 text-muted-foreground" />
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete Note</AlertDialogTitle>
                  <AlertDialogDescription>
                    Are you sure you want to delete this note? This action cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                    Delete
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
