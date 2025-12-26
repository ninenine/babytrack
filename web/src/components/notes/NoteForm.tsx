import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { useCreateNote, type CreateNoteInput } from '@/hooks'

interface NoteFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function NoteForm({ open, onOpenChange }: NoteFormProps) {
  const createNote = useCreateNote()

  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tags, setTags] = useState('')
  const [pinned, setPinned] = useState(false)

  const resetForm = () => {
    setTitle('')
    setContent('')
    setTags('')
    setPinned(false)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const input: CreateNoteInput = {
      title: title || undefined,
      content,
      tags: tags ? tags.split(',').map((t) => t.trim()).filter(Boolean) : undefined,
      pinned,
    }

    await createNote.mutateAsync(input)
    resetForm()
    onOpenChange(false)
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="bottom" className="h-[75vh] overflow-y-auto px-4 sm:px-6">
        <SheetHeader>
          <SheetTitle>Add Note</SheetTitle>
        </SheetHeader>

        <form onSubmit={handleSubmit} className="space-y-6 py-4">
          <div className="space-y-2">
            <Label htmlFor="title">Title (optional)</Label>
            <Input
              id="title"
              placeholder="Note title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="content">Content</Label>
            <Textarea
              id="content"
              placeholder="Write your note..."
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={6}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="tags">Tags (comma separated)</Label>
            <Input
              id="tags"
              placeholder="e.g., health, milestone, feeding"
              value={tags}
              onChange={(e) => setTags(e.target.value)}
            />
          </div>

          <div className="flex items-center justify-between">
            <Label htmlFor="pinned">Pin this note</Label>
            <Switch
              id="pinned"
              checked={pinned}
              onCheckedChange={setPinned}
            />
          </div>

          <div className="flex gap-3 pt-4">
            <Button
              type="button"
              variant="outline"
              className="flex-1"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              className="flex-1"
              disabled={createNote.isPending}
            >
              {createNote.isPending ? 'Saving...' : 'Save'}
            </Button>
          </div>
        </form>
      </SheetContent>
    </Sheet>
  )
}
