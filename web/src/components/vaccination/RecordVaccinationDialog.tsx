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
import { useRecordVaccination } from '@/hooks'
import type { LocalVaccination } from '@/db/dexie'

interface RecordVaccinationDialogProps {
  vaccination: LocalVaccination
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RecordVaccinationDialog({ vaccination, open, onOpenChange }: RecordVaccinationDialogProps) {
  const recordVaccination = useRecordVaccination()

  const [administeredAt, setAdministeredAt] = useState(format(new Date(), "yyyy-MM-dd'T'HH:mm"))
  const [provider, setProvider] = useState('')
  const [location, setLocation] = useState('')
  const [lotNumber, setLotNumber] = useState('')
  const [notes, setNotes] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    await recordVaccination.mutateAsync({
      id: vaccination.id,
      administeredAt: new Date(administeredAt),
      provider: provider || undefined,
      location: location || undefined,
      lotNumber: lotNumber || undefined,
      notes: notes || undefined,
    })

    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Record {vaccination.name} (Dose {vaccination.dose})</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="administeredAt">Date Administered</Label>
            <Input
              id="administeredAt"
              type="datetime-local"
              value={administeredAt}
              onChange={(e) => setAdministeredAt(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="provider">Provider/Doctor</Label>
            <Input
              id="provider"
              placeholder="Dr. Smith"
              value={provider}
              onChange={(e) => setProvider(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="location">Location/Clinic</Label>
            <Input
              id="location"
              placeholder="Pediatric Clinic"
              value={location}
              onChange={(e) => setLocation(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="lotNumber">Lot Number</Label>
            <Input
              id="lotNumber"
              placeholder="ABC123"
              value={lotNumber}
              onChange={(e) => setLotNumber(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">Notes</Label>
            <Textarea
              id="notes"
              placeholder="Any reactions or notes..."
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
              disabled={recordVaccination.isPending}
            >
              {recordVaccination.isPending ? 'Recording...' : 'Record'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
