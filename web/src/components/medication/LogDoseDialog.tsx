import { useState } from 'react'
import { format } from 'date-fns'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useLogDose } from '@/hooks'
import { useSessionStore } from '@/stores/session.store'
import type { LocalMedication } from '@/db/dexie'

interface LogDoseDialogProps {
  medication: LocalMedication
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function LogDoseDialog({ medication, open, onOpenChange }: LogDoseDialogProps) {
  const logDose = useLogDose()
  const user = useSessionStore((state) => state.user)

  const [givenAt, setGivenAt] = useState(format(new Date(), "yyyy-MM-dd'T'HH:mm"))
  const [dosage, setDosage] = useState(medication.dosage)
  const [notes, setNotes] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    await logDose.mutateAsync({
      medicationId: medication.id,
      givenAt: new Date(givenAt),
      givenBy: user?.name ?? 'Unknown',
      dosage: `${dosage} ${medication.unit}`,
      notes: notes || undefined,
    })

    setGivenAt(format(new Date(), "yyyy-MM-dd'T'HH:mm"))
    setDosage(medication.dosage)
    setNotes('')
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Log Dose - {medication.name}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="givenAt">Time Given</Label>
            <Input
              id="givenAt"
              type="datetime-local"
              value={givenAt}
              onChange={(e) => setGivenAt(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="dosage">Dosage ({medication.unit})</Label>
            <Input
              id="dosage"
              value={dosage}
              onChange={(e) => setDosage(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">Notes (optional)</Label>
            <Textarea
              id="notes"
              placeholder="Any notes about this dose..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={2}
            />
          </div>

          <div className="flex gap-3 pt-2">
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
              disabled={logDose.isPending}
            >
              {logDose.isPending ? 'Logging...' : 'Log Dose'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
