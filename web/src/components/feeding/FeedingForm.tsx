import { useEffect } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { format } from 'date-fns'
import { Button } from '@/components/ui/button'
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
import { useCreateFeeding, useUpdateFeeding, type FeedingType } from '@/hooks'
import { cn } from '@/lib/utils'
import type { LocalFeeding } from '@/db/dexie'

// Zod schema for form validation
const feedingFormSchema = z.object({
  type: z.enum(['breast', 'bottle', 'formula', 'solid']),
  startTime: z.string().min(1, 'Start time is required'),
  endTime: z.string().optional(),
  amount: z.string().optional(),
  unit: z.string().optional(),
  side: z.string().optional(),
  notes: z.string().optional(),
})

type FeedingFormValues = z.infer<typeof feedingFormSchema>

interface FeedingFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  feeding?: LocalFeeding | null
}

const feedingTypes: { value: FeedingType; label: string; icon: string }[] = [
  { value: 'breast', label: 'Breast', icon: '🤱' },
  { value: 'bottle', label: 'Bottle', icon: '🍼' },
  { value: 'formula', label: 'Formula', icon: '🥛' },
  { value: 'solid', label: 'Solid', icon: '🥣' },
]

function getDefaultValues(): FeedingFormValues {
  return {
    type: 'bottle',
    startTime: format(new Date(), "yyyy-MM-dd'T'HH:mm"),
    endTime: '',
    amount: '',
    unit: 'ml',
    side: '',
    notes: '',
  }
}

function feedingToFormValues(feeding: LocalFeeding): FeedingFormValues {
  return {
    type: feeding.type as FeedingType,
    startTime: format(new Date(feeding.startTime), "yyyy-MM-dd'T'HH:mm"),
    endTime: feeding.endTime ? format(new Date(feeding.endTime), "yyyy-MM-dd'T'HH:mm") : '',
    amount: feeding.amount?.toString() || '',
    unit: feeding.unit || 'ml',
    side: feeding.side || '',
    notes: feeding.notes || '',
  }
}

export function FeedingForm({ open, onOpenChange, feeding }: FeedingFormProps) {
  const createFeeding = useCreateFeeding()
  const updateFeeding = useUpdateFeeding()
  const isEditing = !!feeding

  const form = useForm<FeedingFormValues>({
    resolver: zodResolver(feedingFormSchema),
    defaultValues: getDefaultValues(),
  })

  const watchType = useWatch({ control: form.control, name: 'type' })

  // Reset form when dialog opens/closes or feeding changes
  useEffect(() => {
    if (open) {
      if (feeding) {
        form.reset(feedingToFormValues(feeding))
      } else {
        form.reset(getDefaultValues())
      }
    }
  }, [feeding, open, form])

  const onSubmit = async (values: FeedingFormValues) => {
    const payload = {
      type: values.type,
      startTime: new Date(values.startTime),
      endTime: values.endTime ? new Date(values.endTime) : undefined,
      amount: values.amount ? parseFloat(values.amount) : undefined,
      unit: values.unit || undefined,
      side: values.side || undefined,
      notes: values.notes || undefined,
    }

    if (isEditing && feeding) {
      await updateFeeding.mutateAsync({
        id: feeding.id,
        ...payload,
        endTime: payload.endTime ?? null,
        amount: payload.amount ?? null,
      })
    } else {
      await createFeeding.mutateAsync(payload)
    }

    onOpenChange(false)
  }

  const isPending = createFeeding.isPending || updateFeeding.isPending

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="bottom" className="h-[85vh] overflow-y-auto px-4 sm:px-6">
        <SheetHeader>
          <SheetTitle>{isEditing ? 'Edit Feeding' : 'Log Feeding'}</SheetTitle>
        </SheetHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 py-4">
            {/* Feeding Type Selection */}
            <FormField
              control={form.control}
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <FormControl>
                    <div className="grid grid-cols-4 gap-2">
                      {feedingTypes.map((ft) => (
                        <button
                          key={ft.value}
                          type="button"
                          onClick={() => field.onChange(ft.value)}
                          className={cn(
                            'flex flex-col items-center gap-1 p-3 rounded-lg border transition-colors',
                            field.value === ft.value
                              ? 'border-primary bg-primary/10 text-primary'
                              : 'border-border hover:border-primary/50'
                          )}
                        >
                          <span className="text-2xl">{ft.icon}</span>
                          <span className="text-xs font-medium">{ft.label}</span>
                        </button>
                      ))}
                    </div>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Time Inputs */}
            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="startTime"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Start Time</FormLabel>
                    <FormControl>
                      <Input type="datetime-local" {...field} />
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
                      <Input type="datetime-local" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {/* Side Selection (for breastfeeding) */}
            {watchType === 'breast' && (
              <FormField
                control={form.control}
                name="side"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Side</FormLabel>
                    <FormControl>
                      <div className="flex gap-2">
                        {['left', 'right', 'both'].map((s) => (
                          <Button
                            key={s}
                            type="button"
                            variant={field.value === s ? 'default' : 'outline'}
                            className="flex-1"
                            onClick={() => field.onChange(s)}
                          >
                            {s.charAt(0).toUpperCase() + s.slice(1)}
                          </Button>
                        ))}
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}

            {/* Amount/Unit (for bottle/formula) */}
            {(watchType === 'bottle' || watchType === 'formula') && (
              <div className="grid grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="amount"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Amount</FormLabel>
                      <FormControl>
                        <Input
                          type="number"
                          placeholder="0"
                          step="5"
                          min="0"
                          {...field}
                        />
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
                          <SelectItem value="ml">ml</SelectItem>
                          <SelectItem value="oz">oz</SelectItem>
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            )}

            {/* Notes */}
            <FormField
              control={form.control}
              name="notes"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Notes</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Optional notes..."
                      rows={3}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Actions */}
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
        </Form>
      </SheetContent>
    </Sheet>
  )
}
