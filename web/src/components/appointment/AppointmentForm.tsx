import { useState, useEffect } from 'react'
import { format, addDays } from 'date-fns'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { useCreateAppointment, useUpdateAppointment, type AppointmentType } from '@/hooks'
import type { LocalAppointment } from '@/db/dexie'

interface AppointmentFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  appointment?: LocalAppointment | null
}

const appointmentTypes: { value: AppointmentType; label: string }[] = [
  { value: 'well_visit', label: 'Well Visit' },
  { value: 'sick_visit', label: 'Sick Visit' },
  { value: 'specialist', label: 'Specialist' },
  { value: 'dental', label: 'Dental' },
  { value: 'other', label: 'Other' },
]

const durations = [
  { value: 15, label: '15 minutes' },
  { value: 30, label: '30 minutes' },
  { value: 45, label: '45 minutes' },
  { value: 60, label: '1 hour' },
  { value: 90, label: '1.5 hours' },
  { value: 120, label: '2 hours' },
]

export function AppointmentForm({ open, onOpenChange, appointment }: AppointmentFormProps) {
  const createAppointment = useCreateAppointment()
  const updateAppointment = useUpdateAppointment()
  const isEditing = !!appointment

  const [type, setType] = useState<AppointmentType>('well_visit')
  const [title, setTitle] = useState('')
  const [provider, setProvider] = useState('')
  const [location, setLocation] = useState('')
  const [scheduledAt, setScheduledAt] = useState(() => format(addDays(new Date(), 7), "yyyy-MM-dd'T'10:00"))
  const [duration, setDuration] = useState('30')
  const [notes, setNotes] = useState('')

  useEffect(() => {
    if (appointment) {
      setType(appointment.type as AppointmentType)
      setTitle(appointment.title)
      setProvider(appointment.provider || '')
      setLocation(appointment.location || '')
      setScheduledAt(format(new Date(appointment.scheduledAt), "yyyy-MM-dd'T'HH:mm"))
      setDuration(appointment.duration.toString())
      setNotes(appointment.notes || '')
    } else {
      setType('well_visit')
      setTitle('')
      setProvider('')
      setLocation('')
      setScheduledAt(format(addDays(new Date(), 7), "yyyy-MM-dd'T'10:00"))
      setDuration('30')
      setNotes('')
    }
  }, [appointment, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (isEditing && appointment) {
      await updateAppointment.mutateAsync({
        id: appointment.id,
        type,
        title,
        provider: provider || undefined,
        location: location || undefined,
        scheduledAt: new Date(scheduledAt),
        duration: parseInt(duration),
        notes: notes || undefined,
      })
    } else {
      await createAppointment.mutateAsync({
        type,
        title,
        provider: provider || undefined,
        location: location || undefined,
        scheduledAt: new Date(scheduledAt),
        duration: parseInt(duration),
        notes: notes || undefined,
      })
    }

    onOpenChange(false)
  }

  const isPending = createAppointment.isPending || updateAppointment.isPending

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="bottom" className="h-[85vh] overflow-y-auto px-4 sm:px-6">
        <SheetHeader>
          <SheetTitle>{isEditing ? 'Edit Appointment' : 'Add Appointment'}</SheetTitle>
        </SheetHeader>

        <form onSubmit={handleSubmit} className="space-y-6 py-4">
          <div className="space-y-2">
            <Label htmlFor="type">Type</Label>
            <Select value={type} onValueChange={(v) => setType(v as AppointmentType)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {appointmentTypes.map((t) => (
                  <SelectItem key={t.value} value={t.value}>
                    {t.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="title">Title</Label>
            <Input
              id="title"
              placeholder="e.g., 6 Month Checkup"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
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
            <Label htmlFor="location">Location</Label>
            <Input
              id="location"
              placeholder="Pediatric Clinic"
              value={location}
              onChange={(e) => setLocation(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="scheduledAt">Date & Time</Label>
              <Input
                id="scheduledAt"
                type="datetime-local"
                value={scheduledAt}
                onChange={(e) => setScheduledAt(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="duration">Duration</Label>
              <Select value={duration} onValueChange={setDuration}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {durations.map((d) => (
                    <SelectItem key={d.value} value={d.value.toString()}>
                      {d.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
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
              disabled={isPending}
            >
              {isPending ? 'Saving...' : isEditing ? 'Save Changes' : 'Save'}
            </Button>
          </div>
        </form>
      </SheetContent>
    </Sheet>
  )
}
