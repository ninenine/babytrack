import { useState, useEffect } from 'react'
import { format } from 'date-fns'
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
import { useCreateFeeding, useUpdateFeeding, type FeedingType } from '@/hooks'
import { cn } from '@/lib/utils'
import type { LocalFeeding } from '@/db/dexie'

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

export function FeedingForm({ open, onOpenChange, feeding }: FeedingFormProps) {
  const createFeeding = useCreateFeeding()
  const updateFeeding = useUpdateFeeding()
  const isEditing = !!feeding

  const [type, setType] = useState<FeedingType>('bottle')
  const [startTime, setStartTime] = useState(format(new Date(), "yyyy-MM-dd'T'HH:mm"))
  const [endTime, setEndTime] = useState('')
  const [amount, setAmount] = useState('')
  const [unit, setUnit] = useState('ml')
  const [side, setSide] = useState('')
  const [notes, setNotes] = useState('')

  const resetForm = () => {
    setType('bottle')
    setStartTime(format(new Date(), "yyyy-MM-dd'T'HH:mm"))
    setEndTime('')
    setAmount('')
    setUnit('ml')
    setSide('')
    setNotes('')
  }

  useEffect(() => {
    if (feeding) {
      setType(feeding.type as FeedingType)
      setStartTime(format(new Date(feeding.startTime), "yyyy-MM-dd'T'HH:mm"))
      setEndTime(feeding.endTime ? format(new Date(feeding.endTime), "yyyy-MM-dd'T'HH:mm") : '')
      setAmount(feeding.amount?.toString() || '')
      setUnit(feeding.unit || 'ml')
      setSide(feeding.side || '')
      setNotes(feeding.notes || '')
    } else {
      resetForm()
    }
  }, [feeding, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (isEditing && feeding) {
      await updateFeeding.mutateAsync({
        id: feeding.id,
        type,
        startTime: new Date(startTime),
        endTime: endTime ? new Date(endTime) : null,
        amount: amount ? parseFloat(amount) : null,
        unit: unit || undefined,
        side: side || undefined,
        notes: notes || undefined,
      })
    } else {
      await createFeeding.mutateAsync({
        type,
        startTime: new Date(startTime),
        endTime: endTime ? new Date(endTime) : undefined,
        amount: amount ? parseFloat(amount) : undefined,
        unit: unit || undefined,
        side: side || undefined,
        notes: notes || undefined,
      })
    }

    resetForm()
    onOpenChange(false)
  }

  const isPending = createFeeding.isPending || updateFeeding.isPending

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="bottom" className="h-[85vh] overflow-y-auto px-4 sm:px-6">
        <SheetHeader>
          <SheetTitle>{isEditing ? 'Edit Feeding' : 'Log Feeding'}</SheetTitle>
        </SheetHeader>

        <form onSubmit={handleSubmit} className="space-y-6 py-4">
          {/* Feeding Type Selection */}
          <div className="space-y-2">
            <Label>Type</Label>
            <div className="grid grid-cols-4 gap-2">
              {feedingTypes.map((ft) => (
                <button
                  key={ft.value}
                  type="button"
                  onClick={() => setType(ft.value)}
                  className={cn(
                    'flex flex-col items-center gap-1 p-3 rounded-lg border transition-colors',
                    type === ft.value
                      ? 'border-primary bg-primary/10 text-primary'
                      : 'border-border hover:border-primary/50'
                  )}
                >
                  <span className="text-2xl">{ft.icon}</span>
                  <span className="text-xs font-medium">{ft.label}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Time Inputs */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="startTime">Start Time</Label>
              <Input
                id="startTime"
                type="datetime-local"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="endTime">End Time</Label>
              <Input
                id="endTime"
                type="datetime-local"
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
              />
            </div>
          </div>

          {/* Side Selection (for breastfeeding) */}
          {type === 'breast' && (
            <div className="space-y-2">
              <Label>Side</Label>
              <div className="flex gap-2">
                {['left', 'right', 'both'].map((s) => (
                  <Button
                    key={s}
                    type="button"
                    variant={side === s ? 'default' : 'outline'}
                    className="flex-1"
                    onClick={() => setSide(s)}
                  >
                    {s.charAt(0).toUpperCase() + s.slice(1)}
                  </Button>
                ))}
              </div>
            </div>
          )}

          {/* Amount/Unit (for bottle/formula) */}
          {(type === 'bottle' || type === 'formula') && (
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="amount">Amount</Label>
                <Input
                  id="amount"
                  type="number"
                  placeholder="0"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  step="5"
                  min="0"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="unit">Unit</Label>
                <Select value={unit} onValueChange={setUnit}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="ml">ml</SelectItem>
                    <SelectItem value="oz">oz</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}

          {/* Notes */}
          <div className="space-y-2">
            <Label htmlFor="notes">Notes</Label>
            <Textarea
              id="notes"
              placeholder="Optional notes..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={3}
            />
          </div>

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
      </SheetContent>
    </Sheet>
  )
}
