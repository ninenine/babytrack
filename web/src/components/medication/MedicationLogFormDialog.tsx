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
import { useUpdateMedicationLog } from '@/hooks'
import type { LocalMedicationLog } from '@/db/dexie'

interface EnrichedMedicationLog extends LocalMedicationLog {
  medicationName?: string
}

interface MedicationLogFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  log: EnrichedMedicationLog | null
}

export function MedicationLogFormDialog({ open, onOpenChange, log }: MedicationLogFormDialogProps) {
  const updateLog = useUpdateMedicationLog()

  const [givenAt, setGivenAt] = useState('')
  const [givenBy, setGivenBy] = useState('')
  const [dosage, setDosage] = useState('')
  const [notes, setNotes] = useState('')

  useEffect(() => {
    if (log) {
      setGivenAt(format(new Date(log.givenAt), "yyyy-MM-dd'T'HH:mm"))
      setGivenBy(log.givenBy)
      setDosage(log.dosage)
      setNotes(log.notes || '')
    }
  }, [log, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!log) return

    await updateLog.mutateAsync({
      id: log.id,
      givenAt: new Date(givenAt),
      givenBy,
      dosage,
      notes: notes || undefined,
    })

    onOpenChange(false)
  }

  if (!log) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit Medication Log</DialogTitle>
          <DialogDescription>
            Update the dose record for {log.medicationName || 'medication'}.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="givenAt">Given At</Label>
            <Input
              id="givenAt"
              type="datetime-local"
              value={givenAt}
              onChange={(e) => setGivenAt(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="givenBy">Given By</Label>
            <Input
              id="givenBy"
              placeholder="Name of person"
              value={givenBy}
              onChange={(e) => setGivenBy(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="dosage">Dosage</Label>
            <Input
              id="dosage"
              placeholder="e.g., 5ml, 1 tablet"
              value={dosage}
              onChange={(e) => setDosage(e.target.value)}
              required
            />
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
            <Button type="submit" disabled={updateLog.isPending || !givenBy || !dosage}>
              {updateLog.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
