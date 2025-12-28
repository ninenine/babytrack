import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { DateTimePicker } from '@/components/shared/datetime-picker'
import { Input } from '@/components/ui/input'
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
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { useUpdateMedicationLog } from '@/hooks'
import type { LocalMedicationLog } from '@/db/dexie'

const medicationLogFormSchema = z.object({
  givenAt: z.date({ error: 'Given at is required' }),
  givenBy: z.string().min(1, 'Given by is required'),
  dosage: z.string().min(1, 'Dosage is required'),
  notes: z.string().optional(),
})

type MedicationLogFormValues = z.infer<typeof medicationLogFormSchema>

interface EnrichedMedicationLog extends LocalMedicationLog {
  medicationName?: string
}

interface MedicationLogFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  log: EnrichedMedicationLog | null
}

function logToFormValues(log: LocalMedicationLog): MedicationLogFormValues {
  return {
    givenAt: new Date(log.givenAt),
    givenBy: log.givenBy,
    dosage: log.dosage,
    notes: log.notes || '',
  }
}

export function MedicationLogFormDialog({ open, onOpenChange, log }: MedicationLogFormDialogProps) {
  const updateLog = useUpdateMedicationLog()

  const form = useForm<MedicationLogFormValues>({
    resolver: zodResolver(medicationLogFormSchema),
    defaultValues: {
      givenAt: new Date(),
      givenBy: '',
      dosage: '',
      notes: '',
    },
  })

  useEffect(() => {
    if (log && open) {
      form.reset(logToFormValues(log))
    }
  }, [log, open, form])

  const onSubmit = async (values: MedicationLogFormValues) => {
    if (!log) return

    try {
      await updateLog.mutateAsync({
        id: log.id,
        givenAt: values.givenAt,
        givenBy: values.givenBy,
        dosage: values.dosage,
        notes: values.notes || undefined,
      })
      toast.success('Medication log updated')
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to update medication log')
    }
  }

  if (!log) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Edit Medication Log</DialogTitle>
          <DialogDescription>
            Update the dose record for {log.medicationName || 'medication'}.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="givenAt"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Given At</FormLabel>
                  <FormControl>
                    <DateTimePicker
                      date={field.value}
                      onDateChange={field.onChange}
                      placeholder="Select date and time"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="givenBy"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Given By</FormLabel>
                  <FormControl>
                    <Input placeholder="Name of person" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="dosage"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Dosage</FormLabel>
                  <FormControl>
                    <Input placeholder="e.g., 5ml, 1 tablet" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

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

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={updateLog.isPending}>
                {updateLog.isPending ? 'Saving...' : 'Save Changes'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
