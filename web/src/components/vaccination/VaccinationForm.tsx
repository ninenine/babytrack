import { useState, useEffect } from 'react'
import { addMonths } from 'date-fns'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { DatePicker } from '@/components/ui/date-picker'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { useCreateVaccination, useUpdateVaccination } from '@/hooks'
import type { LocalVaccination } from '@/db/dexie'

interface VaccinationFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  vaccination?: LocalVaccination | null
}

export function VaccinationForm({ open, onOpenChange, vaccination }: VaccinationFormProps) {
  const createVaccination = useCreateVaccination()
  const updateVaccination = useUpdateVaccination()
  const isEditing = !!vaccination

  const [name, setName] = useState('')
  const [dose, setDose] = useState('1')
  const [scheduledAt, setScheduledAt] = useState<Date | undefined>(addMonths(new Date(), 1))
  const [provider, setProvider] = useState('')
  const [location, setLocation] = useState('')
  const [lotNumber, setLotNumber] = useState('')
  const [notes, setNotes] = useState('')

  const resetForm = () => {
    setName('')
    setDose('1')
    setScheduledAt(addMonths(new Date(), 1))
    setProvider('')
    setLocation('')
    setLotNumber('')
    setNotes('')
  }

  useEffect(() => {
    if (vaccination) {
      setName(vaccination.name)
      setDose(vaccination.dose.toString())
      setScheduledAt(new Date(vaccination.scheduledAt))
      setProvider(vaccination.provider || '')
      setLocation(vaccination.location || '')
      setLotNumber(vaccination.lotNumber || '')
      setNotes(vaccination.notes || '')
    } else {
      resetForm()
    }
  }, [vaccination, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!scheduledAt) return

    if (isEditing && vaccination) {
      await updateVaccination.mutateAsync({
        id: vaccination.id,
        name,
        dose: parseInt(dose),
        scheduledAt,
        provider: provider || undefined,
        location: location || undefined,
        lotNumber: lotNumber || undefined,
        notes: notes || undefined,
      })
    } else {
      await createVaccination.mutateAsync({
        name,
        dose: parseInt(dose),
        scheduledAt,
        notes: notes || undefined,
      })
    }

    resetForm()
    onOpenChange(false)
  }

  const isPending = createVaccination.isPending || updateVaccination.isPending

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="bottom" className="h-[80vh] overflow-y-auto px-4 sm:px-6">
        <SheetHeader>
          <SheetTitle>{isEditing ? 'Edit Vaccination' : 'Add Vaccination'}</SheetTitle>
        </SheetHeader>

        <form onSubmit={handleSubmit} className="space-y-6 py-4">
          <div className="space-y-2">
            <Label htmlFor="name">Vaccine Name</Label>
            <Input
              id="name"
              placeholder="e.g., DTaP, MMR, Hepatitis B"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="dose">Dose Number</Label>
              <Input
                id="dose"
                type="number"
                min="1"
                value={dose}
                onChange={(e) => setDose(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label>Scheduled Date</Label>
              <DatePicker
                date={scheduledAt}
                onDateChange={setScheduledAt}
                placeholder="Select date"
              />
            </div>
          </div>

          {isEditing && (
            <>
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
                <Label htmlFor="location">Location</Label>
                <Input
                  id="location"
                  placeholder="Clinic name"
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="lotNumber">Lot Number</Label>
                <Input
                  id="lotNumber"
                  placeholder="Vaccine lot number"
                  value={lotNumber}
                  onChange={(e) => setLotNumber(e.target.value)}
                />
              </div>
            </>
          )}

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
              disabled={isPending || !scheduledAt || !name}
            >
              {isPending ? 'Saving...' : isEditing ? 'Save Changes' : 'Save'}
            </Button>
          </div>
        </form>
      </SheetContent>
    </Sheet>
  )
}
