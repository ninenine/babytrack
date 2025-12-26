import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { DateTimePicker } from '@/components/ui/datetime-picker'
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
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { useUpdateSleep, type SleepType } from '@/hooks'
import type { LocalSleep } from '@/db/dexie'

const sleepFormSchema = z.object({
  type: z.enum(['nap', 'night']),
  startTime: z.date({ error: 'Start time is required' }),
  endTime: z.date().optional(),
  quality: z.string().optional(),
  notes: z.string().optional(),
})

type SleepFormValues = z.infer<typeof sleepFormSchema>

interface SleepFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sleep: LocalSleep | null
}

function sleepToFormValues(sleep: LocalSleep): SleepFormValues {
  return {
    type: sleep.type as SleepType,
    startTime: new Date(sleep.startTime),
    endTime: sleep.endTime ? new Date(sleep.endTime) : undefined,
    quality: sleep.quality?.toString(),
    notes: sleep.notes || '',
  }
}

export function SleepFormDialog({ open, onOpenChange, sleep }: SleepFormDialogProps) {
  const updateSleep = useUpdateSleep()

  const form = useForm<SleepFormValues>({
    resolver: zodResolver(sleepFormSchema),
    defaultValues: {
      type: 'nap',
      startTime: new Date(),
      endTime: undefined,
      quality: undefined,
      notes: '',
    },
  })

  useEffect(() => {
    if (sleep && open) {
      form.reset(sleepToFormValues(sleep))
    }
  }, [sleep, open, form])

  const onSubmit = async (values: SleepFormValues) => {
    if (!sleep) return

    try {
      await updateSleep.mutateAsync({
        id: sleep.id,
        type: values.type,
        startTime: values.startTime,
        endTime: values.endTime ?? null,
        quality: values.quality ? parseInt(values.quality) : undefined,
        notes: values.notes || undefined,
      })
      toast.success('Sleep record updated')
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to update sleep record')
    }
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

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
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
                      <SelectItem value="nap">Nap</SelectItem>
                      <SelectItem value="night">Night Sleep</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="startTime"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Start Time</FormLabel>
                  <FormControl>
                    <DateTimePicker
                      date={field.value}
                      onDateChange={field.onChange}
                      placeholder="Select start time"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="endTime"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>End Time</FormLabel>
                  <FormControl>
                    <DateTimePicker
                      date={field.value}
                      onDateChange={field.onChange}
                      placeholder="Select end time"
                    />
                  </FormControl>
                  <FormDescription>
                    Leave empty if still sleeping
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="quality"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Quality (1-5)</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Rate sleep quality" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="1">1 - Poor</SelectItem>
                      <SelectItem value="2">2 - Fair</SelectItem>
                      <SelectItem value="3">3 - Good</SelectItem>
                      <SelectItem value="4">4 - Very Good</SelectItem>
                      <SelectItem value="5">5 - Excellent</SelectItem>
                    </SelectContent>
                  </Select>
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
                    <Textarea
                      placeholder="Any notes..."
                      rows={2}
                      {...field}
                    />
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
              <Button type="submit" disabled={updateSleep.isPending}>
                {updateSleep.isPending ? 'Saving...' : 'Save Changes'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
