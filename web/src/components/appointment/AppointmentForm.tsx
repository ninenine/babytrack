import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { addDays } from 'date-fns'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { DateTimePicker } from '@/components/ui/datetime-picker'
import { Input } from '@/components/ui/input'
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
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { useCreateAppointment, useUpdateAppointment, type AppointmentType } from '@/hooks'
import type { LocalAppointment } from '@/db/dexie'

const appointmentFormSchema = z.object({
  type: z.enum(['well_visit', 'sick_visit', 'specialist', 'dental', 'other']),
  title: z.string().min(1, 'Title is required'),
  provider: z.string().optional(),
  location: z.string().optional(),
  scheduledAt: z.date({ error: 'Date & time is required' }),
  duration: z.string(),
  notes: z.string().optional(),
})

type AppointmentFormValues = z.infer<typeof appointmentFormSchema>

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
  { value: '15', label: '15 minutes' },
  { value: '30', label: '30 minutes' },
  { value: '45', label: '45 minutes' },
  { value: '60', label: '1 hour' },
  { value: '90', label: '1.5 hours' },
  { value: '120', label: '2 hours' },
]

function getDefaultValues(): AppointmentFormValues {
  const defaultDate = addDays(new Date(), 7)
  defaultDate.setHours(10, 0, 0, 0)
  return {
    type: 'well_visit',
    title: '',
    provider: '',
    location: '',
    scheduledAt: defaultDate,
    duration: '30',
    notes: '',
  }
}

function appointmentToFormValues(appointment: LocalAppointment): AppointmentFormValues {
  return {
    type: appointment.type as AppointmentType,
    title: appointment.title,
    provider: appointment.provider || '',
    location: appointment.location || '',
    scheduledAt: new Date(appointment.scheduledAt),
    duration: appointment.duration.toString(),
    notes: appointment.notes || '',
  }
}

export function AppointmentForm({ open, onOpenChange, appointment }: AppointmentFormProps) {
  const createAppointment = useCreateAppointment()
  const updateAppointment = useUpdateAppointment()
  const isEditing = !!appointment

  const form = useForm<AppointmentFormValues>({
    resolver: zodResolver(appointmentFormSchema),
    defaultValues: getDefaultValues(),
  })

  useEffect(() => {
    if (open) {
      if (appointment) {
        form.reset(appointmentToFormValues(appointment))
      } else {
        form.reset(getDefaultValues())
      }
    }
  }, [appointment, open, form])

  const onSubmit = async (values: AppointmentFormValues) => {
    const payload = {
      type: values.type,
      title: values.title,
      provider: values.provider || undefined,
      location: values.location || undefined,
      scheduledAt: values.scheduledAt,
      duration: parseInt(values.duration),
      notes: values.notes || undefined,
    }

    try {
      if (isEditing && appointment) {
        await updateAppointment.mutateAsync({
          id: appointment.id,
          ...payload,
        })
        toast.success('Appointment updated')
      } else {
        await createAppointment.mutateAsync(payload)
        toast.success('Appointment scheduled')
      }
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to save appointment')
    }
  }

  const isPending = createAppointment.isPending || updateAppointment.isPending

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="bottom" className="h-[85vh] overflow-y-auto px-4 sm:px-6">
        <SheetHeader>
          <SheetTitle>{isEditing ? 'Edit Appointment' : 'Add Appointment'}</SheetTitle>
        </SheetHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 py-4">
            <FormField
              control={form.control}
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {appointmentTypes.map((t) => (
                        <SelectItem key={t.value} value={t.value}>
                          {t.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="title"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Title</FormLabel>
                  <FormControl>
                    <Input placeholder="e.g., 6 Month Checkup" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="provider"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Provider/Doctor</FormLabel>
                  <FormControl>
                    <Input placeholder="Dr. Smith" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="location"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Location</FormLabel>
                  <FormControl>
                    <Input placeholder="Pediatric Clinic" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="scheduledAt"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Date & Time</FormLabel>
                    <FormControl>
                      <DateTimePicker
                        date={field.value}
                        onDateChange={field.onChange}
                        placeholder="Select date & time"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="duration"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Duration</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {durations.map((d) => (
                          <SelectItem key={d.value} value={d.value}>
                            {d.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="notes"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Notes</FormLabel>
                  <FormControl>
                    <Textarea placeholder="Any notes..." rows={2} {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="flex gap-3 pt-4">
              <Button
                type="button"
                variant="outline"
                className="flex-1"
                onClick={() => onOpenChange(false)}
              >
                Cancel
              </Button>
              <Button type="submit" className="flex-1" disabled={isPending}>
                {isPending ? 'Saving...' : isEditing ? 'Save Changes' : 'Save'}
              </Button>
            </div>
          </form>
        </Form>
      </SheetContent>
    </Sheet>
  )
}
