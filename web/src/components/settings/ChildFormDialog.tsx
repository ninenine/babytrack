import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { DatePicker } from '@/components/ui/date-picker'
import { useFamilyStore, type Child } from '@/stores/family.store'
import { parseISO } from 'date-fns'

interface ChildFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  child?: Child | null
}

export function ChildFormDialog({ open, onOpenChange, child }: ChildFormDialogProps) {
  const { addChild, updateChild } = useFamilyStore()
  const isEditing = !!child

  const [name, setName] = useState('')
  const [dateOfBirth, setDateOfBirth] = useState<Date | undefined>()
  const [gender, setGender] = useState<string>('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (child) {
      setName(child.name)
      setDateOfBirth(parseISO(child.dateOfBirth))
      setGender(child.gender || '')
    } else {
      setName('')
      setDateOfBirth(undefined)
      setGender('')
    }
  }, [child, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name || !dateOfBirth) return

    setIsSubmitting(true)

    try {
      const childData: Child = {
        id: child?.id || crypto.randomUUID(),
        name,
        dateOfBirth: dateOfBirth.toISOString().split('T')[0],
        gender: gender || undefined,
        avatarUrl: child?.avatarUrl,
      }

      if (isEditing) {
        updateChild(childData)
      } else {
        addChild(childData)
      }

      onOpenChange(false)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{isEditing ? 'Edit Child' : 'Add Child'}</DialogTitle>
          <DialogDescription>
            {isEditing
              ? "Update your child's information."
              : "Add a new child to your family."}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              placeholder="Child's name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label>Date of Birth</Label>
            <DatePicker
              date={dateOfBirth}
              onDateChange={setDateOfBirth}
              placeholder="Select date of birth"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="gender">Gender (optional)</Label>
            <Select value={gender} onValueChange={setGender}>
              <SelectTrigger>
                <SelectValue placeholder="Select gender" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="male">Male</SelectItem>
                <SelectItem value="female">Female</SelectItem>
                <SelectItem value="other">Other</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting || !name || !dateOfBirth}>
              {isSubmitting ? 'Saving...' : isEditing ? 'Save Changes' : 'Add Child'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
