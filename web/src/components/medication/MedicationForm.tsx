import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { DatePicker } from '@/components/shared/date-picker'
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
import { useCreateMedication } from '@/hooks'

const medicationFormSchema = z.object({
  name: z.string().min(1, 'Medication name is required'),
  dosage: z.string().min(1, 'Dosage is required'),
  unit: z.string(),
  frequency: z.string(),
  instructions: z.string().optional(),
  startDate: z.date({ error: 'Start date is required' }),
  endDate: z.date().optional(),
})

type MedicationFormValues = z.infer<typeof medicationFormSchema>

interface MedicationFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const frequencies = [
  { value: 'once_daily', label: 'Once daily' },
  { value: 'twice_daily', label: 'Twice daily' },
  { value: 'three_times_daily', label: 'Three times daily' },
  { value: 'four_times_daily', label: 'Four times daily' },
  { value: 'every_4_hours', label: 'Every 4 hours' },
  { value: 'every_6_hours', label: 'Every 6 hours' },
  { value: 'every_8_hours', label: 'Every 8 hours' },
  { value: 'as_needed', label: 'As needed' },
  { value: 'weekly', label: 'Weekly' },
]

const units = ['ml', 'mg', 'drops', 'tablets', 'tsp', 'tbsp']

function getDefaultValues(): MedicationFormValues {
  return {
    name: '',
    dosage: '',
    unit: 'ml',
    frequency: 'twice_daily',
    instructions: '',
    startDate: new Date(),
    endDate: undefined,
  }
}

export function MedicationForm({ open, onOpenChange }: MedicationFormProps) {
  const createMedication = useCreateMedication()

  const form = useForm<MedicationFormValues>({
    resolver: zodResolver(medicationFormSchema),
    defaultValues: getDefaultValues(),
  })

  useEffect(() => {
    if (open) {
      form.reset(getDefaultValues())
    }
  }, [open, form])

  const onSubmit = async (values: MedicationFormValues) => {
    try {
      await createMedication.mutateAsync({
        name: values.name,
        dosage: values.dosage,
        unit: values.unit,
        frequency: values.frequency,
        instructions: values.instructions || undefined,
        startDate: values.startDate,
        endDate: values.endDate,
      })
      toast.success('Medication added')
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to add medication')
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="bottom" className="h-[85vh] overflow-y-auto px-4 sm:px-6">
        <SheetHeader>
          <SheetTitle>Add Medication</SheetTitle>
        </SheetHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 py-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Medication Name</FormLabel>
                  <FormControl>
                    <Input placeholder="e.g., Amoxicillin" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="dosage"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Dosage</FormLabel>
                    <FormControl>
                      <Input placeholder="e.g., 5" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="unit"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Unit</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {units.map((u) => (
                          <SelectItem key={u} value={u}>
                            {u}
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
              name="frequency"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Frequency</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {frequencies.map((f) => (
                        <SelectItem key={f.value} value={f.value}>
                          {f.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="startDate"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Start Date</FormLabel>
                    <FormControl>
                      <DatePicker
                        date={field.value}
                        onDateChange={field.onChange}
                        placeholder="Select start date"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="endDate"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>End Date (optional)</FormLabel>
                    <FormControl>
                      <DatePicker
                        date={field.value}
                        onDateChange={field.onChange}
                        placeholder="Select end date"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="instructions"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Instructions</FormLabel>
                  <FormControl>
                    <Textarea placeholder="e.g., Take with food" rows={2} {...field} />
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
              <Button type="submit" className="flex-1" disabled={createMedication.isPending}>
                {createMedication.isPending ? 'Saving...' : 'Save'}
              </Button>
            </div>
          </form>
        </Form>
      </SheetContent>
    </Sheet>
  )
}
