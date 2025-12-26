import { useState } from 'react'
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
import { useCreateMedication, type CreateMedicationInput } from '@/hooks'

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

export function MedicationForm({ open, onOpenChange }: MedicationFormProps) {
  const createMedication = useCreateMedication()

  const [name, setName] = useState('')
  const [dosage, setDosage] = useState('')
  const [unit, setUnit] = useState('ml')
  const [frequency, setFrequency] = useState('twice_daily')
  const [instructions, setInstructions] = useState('')
  const [startDate, setStartDate] = useState(format(new Date(), 'yyyy-MM-dd'))
  const [endDate, setEndDate] = useState('')

  const resetForm = () => {
    setName('')
    setDosage('')
    setUnit('ml')
    setFrequency('twice_daily')
    setInstructions('')
    setStartDate(format(new Date(), 'yyyy-MM-dd'))
    setEndDate('')
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const input: CreateMedicationInput = {
      name,
      dosage,
      unit,
      frequency,
      instructions: instructions || undefined,
      startDate: new Date(startDate),
      endDate: endDate ? new Date(endDate) : undefined,
    }

    await createMedication.mutateAsync(input)
    resetForm()
    onOpenChange(false)
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="bottom" className="h-[85vh] overflow-y-auto px-4 sm:px-6">
        <SheetHeader>
          <SheetTitle>Add Medication</SheetTitle>
        </SheetHeader>

        <form onSubmit={handleSubmit} className="space-y-6 py-4">
          <div className="space-y-2">
            <Label htmlFor="name">Medication Name</Label>
            <Input
              id="name"
              placeholder="e.g., Amoxicillin"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="dosage">Dosage</Label>
              <Input
                id="dosage"
                placeholder="e.g., 5"
                value={dosage}
                onChange={(e) => setDosage(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="unit">Unit</Label>
              <Select value={unit} onValueChange={setUnit}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {units.map((u) => (
                    <SelectItem key={u} value={u}>
                      {u}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="frequency">Frequency</Label>
            <Select value={frequency} onValueChange={setFrequency}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {frequencies.map((f) => (
                  <SelectItem key={f.value} value={f.value}>
                    {f.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="startDate">Start Date</Label>
              <Input
                id="startDate"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="endDate">End Date (optional)</Label>
              <Input
                id="endDate"
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="instructions">Instructions</Label>
            <Textarea
              id="instructions"
              placeholder="e.g., Take with food"
              value={instructions}
              onChange={(e) => setInstructions(e.target.value)}
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
              disabled={createMedication.isPending}
            >
              {createMedication.isPending ? 'Saving...' : 'Save'}
            </Button>
          </div>
        </form>
      </SheetContent>
    </Sheet>
  )
}
