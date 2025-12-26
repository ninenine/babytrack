import { useState, useEffect } from 'react'
import { format } from 'date-fns'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useUpdateSleep, type SleepType, type UpdateSleepInput } from '@/hooks'
import type { LocalSleep } from '@/db/dexie'

interface SleepFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sleep: LocalSleep | null
}

export function SleepFormDialog({ open, onOpenChange, sleep }: SleepFormDialogProps) {
  const updateSleep = useUpdateSleep()

  const [type, setType] = useState<SleepType>('nap')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const [quality, setQuality] = useState('')
  const [notes, setNotes] = useState('')

  useEffect(() => {
    if (sleep) {
      setType(sleep.type)
      setStartTime(format(new Date(sleep.startTime), "yyyy-MM-dd'T'HH:mm"))
      setEndTime(sleep.endTime ? format(new Date(sleep.endTime), "yyyy-MM-dd'T'HH:mm") : '')
      setQuality(sleep.quality?.toString() || '')
      setNotes(sleep.notes || '')
    }
  }, [sleep, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!sleep) return

    const input: UpdateSleepInput = {
      id: sleep.id,
      type,
      startTime: new Date(startTime),
      endTime: endTime ? new Date(endTime) : null,
      quality: quality ? parseInt(quality) : undefined,
      notes: notes || undefined,
    }

    await updateSleep.mutateAsync(input)
    onOpenChange(false)
  }

  if (!sleep) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit Sleep Record</DialogTitle>
          <DialogDescription>
            Update the sleep record details.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="type">Type</Label>
            <Select value={type} onValueChange={(v) => setType(v as SleepType)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="nap">Nap</SelectItem>
                <SelectItem value="night">Night Sleep</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="startTime">Start Time</Label>
            <Input
              id="startTime"
              type="datetime-local"
              value={startTime}
              onChange={(e) => setStartTime(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="endTime">End Time</Label>
            <Input
              id="endTime"
              type="datetime-local"
              value={endTime}
              onChange={(e) => setEndTime(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Leave empty if still sleeping
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="quality">Quality (1-5)</Label>
            <Select value={quality} onValueChange={setQuality}>
              <SelectTrigger>
                <SelectValue placeholder="Rate sleep quality" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">Not rated</SelectItem>
                <SelectItem value="1">1 - Poor</SelectItem>
                <SelectItem value="2">2 - Fair</SelectItem>
                <SelectItem value="3">3 - Good</SelectItem>
                <SelectItem value="4">4 - Very Good</SelectItem>
                <SelectItem value="5">5 - Excellent</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">Notes</Label>
            <Textarea
              id="notes"
              placeholder="Any notes..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={2}
            />
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={updateSleep.isPending || !startTime}>
              {updateSleep.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
