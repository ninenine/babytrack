import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { DatePicker } from '@/components/shared/date-picker'
import { toAPIDate, fromAPIDate } from '@/lib/dates'
import { Input } from '@/components/ui/input'
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
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { useFamilyStore, type Child } from '@/stores/family.store'

const childFormSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  dateOfBirth: z.date({ error: 'Date of birth is required' }),
  gender: z.string().optional(),
})

type ChildFormValues = z.infer<typeof childFormSchema>

interface ChildFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  child?: Child | null
}

function getDefaultValues(): ChildFormValues {
  return {
    name: '',
    dateOfBirth: undefined as unknown as Date,
    gender: '',
  }
}

function childToFormValues(child: Child): ChildFormValues {
  return {
    name: child.name,
    dateOfBirth: fromAPIDate(child.dateOfBirth),
    gender: child.gender || '',
  }
}

export function ChildFormDialog({ open, onOpenChange, child }: ChildFormDialogProps) {
  const { addChild, updateChild } = useFamilyStore()
  const isEditing = !!child

  const form = useForm<ChildFormValues>({
    resolver: zodResolver(childFormSchema),
    defaultValues: getDefaultValues(),
  })

  useEffect(() => {
    if (open) {
      if (child) {
        form.reset(childToFormValues(child))
      } else {
        form.reset(getDefaultValues())
      }
    }
  }, [child, open, form])

  const onSubmit = async (values: ChildFormValues) => {
    try {
      const childData: Child = {
        id: child?.id || crypto.randomUUID(),
        name: values.name,
        dateOfBirth: toAPIDate(values.dateOfBirth),
        gender: values.gender || undefined,
        avatarUrl: child?.avatarUrl,
      }

      if (isEditing) {
        updateChild(childData)
        toast.success('Child updated')
      } else {
        addChild(childData)
        toast.success('Child added')
      }
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to save child')
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
              : 'Add a new child to your family.'}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Child's name" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="dateOfBirth"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Date of Birth</FormLabel>
                  <FormControl>
                    <DatePicker
                      date={field.value}
                      onDateChange={field.onChange}
                      placeholder="Select date of birth"
                      toDate={new Date()}
                      captionLayout="dropdown"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="gender"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Gender (optional)</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select gender" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="male">Male</SelectItem>
                      <SelectItem value="female">Female</SelectItem>
                      <SelectItem value="other">Other</SelectItem>
                    </SelectContent>
                  </Select>
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
              <Button type="submit" disabled={form.formState.isSubmitting}>
                {form.formState.isSubmitting
                  ? 'Saving...'
                  : isEditing
                    ? 'Save Changes'
                    : 'Add Child'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
