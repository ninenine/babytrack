import { useEffect } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { addMonths } from 'date-fns'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { DatePicker } from '@/components/ui/date-picker'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
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
import { useCreateVaccination, useUpdateVaccination } from '@/hooks'
import type { LocalVaccination } from '@/db/dexie'

const vaccinationFormSchema = z.object({
  name: z.string().min(1, 'Vaccine name is required'),
  dose: z.string().min(1, 'Dose number is required'),
  scheduledAt: z.date({ error: 'Scheduled date is required' }),
  provider: z.string().optional(),
  location: z.string().optional(),
  lotNumber: z.string().optional(),
  notes: z.string().optional(),
})

type VaccinationFormValues = z.infer<typeof vaccinationFormSchema>

interface VaccinationFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  vaccination?: LocalVaccination | null
}

function getDefaultValues(): VaccinationFormValues {
  return {
    name: '',
    dose: '1',
    scheduledAt: addMonths(new Date(), 1),
    provider: '',
    location: '',
    lotNumber: '',
    notes: '',
  }
}

function vaccinationToFormValues(vaccination: LocalVaccination): VaccinationFormValues {
  return {
    name: vaccination.name,
    dose: vaccination.dose.toString(),
    scheduledAt: new Date(vaccination.scheduledAt),
    provider: vaccination.provider || '',
    location: vaccination.location || '',
    lotNumber: vaccination.lotNumber || '',
    notes: vaccination.notes || '',
  }
}

export function VaccinationForm({ open, onOpenChange, vaccination }: VaccinationFormProps) {
  const createVaccination = useCreateVaccination()
  const updateVaccination = useUpdateVaccination()
  const isEditing = !!vaccination

  const form = useForm<VaccinationFormValues>({
    resolver: zodResolver(vaccinationFormSchema),
    defaultValues: getDefaultValues(),
  })

  // Watch for isEditing to conditionally render fields
  const watchName = useWatch({ control: form.control, name: 'name' })

  useEffect(() => {
    if (open) {
      if (vaccination) {
        form.reset(vaccinationToFormValues(vaccination))
      } else {
        form.reset(getDefaultValues())
      }
    }
  }, [vaccination, open, form])

  const onSubmit = async (values: VaccinationFormValues) => {
    try {
      if (isEditing && vaccination) {
        await updateVaccination.mutateAsync({
          id: vaccination.id,
          name: values.name,
          dose: parseInt(values.dose),
          scheduledAt: values.scheduledAt,
          provider: values.provider || undefined,
          location: values.location || undefined,
          lotNumber: values.lotNumber || undefined,
          notes: values.notes || undefined,
        })
        toast.success('Vaccination updated')
      } else {
        await createVaccination.mutateAsync({
          name: values.name,
          dose: parseInt(values.dose),
          scheduledAt: values.scheduledAt,
          notes: values.notes || undefined,
        })
        toast.success('Vaccination scheduled')
      }
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to save vaccination')
    }
  }

  const isPending = createVaccination.isPending || updateVaccination.isPending

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="bottom" className="h-[80vh] overflow-y-auto px-4 sm:px-6">
        <SheetHeader>
          <SheetTitle>{isEditing ? 'Edit Vaccination' : 'Add Vaccination'}</SheetTitle>
        </SheetHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 py-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Vaccine Name</FormLabel>
                  <FormControl>
                    <Input placeholder="e.g., DTaP, MMR, Hepatitis B" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="dose"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Dose Number</FormLabel>
                    <FormControl>
                      <Input type="number" min="1" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="scheduledAt"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Scheduled Date</FormLabel>
                    <FormControl>
                      <DatePicker
                        date={field.value}
                        onDateChange={field.onChange}
                        placeholder="Select date"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {isEditing && (
              <>
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
                        <Input placeholder="Clinic name" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="lotNumber"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Lot Number</FormLabel>
                      <FormControl>
                        <Input placeholder="Vaccine lot number" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </>
            )}

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
              <Button
                type="submit"
                className="flex-1"
                disabled={isPending || !watchName}
              >
                {isPending ? 'Saving...' : isEditing ? 'Save Changes' : 'Save'}
              </Button>
            </div>
          </form>
        </Form>
      </SheetContent>
    </Sheet>
  )
}
