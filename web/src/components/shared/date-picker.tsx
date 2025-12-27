import * as React from 'react'
import { format } from 'date-fns'
import { CalendarIcon } from 'lucide-react'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

interface DatePickerProps {
  date?: Date
  onDateChange: (date: Date | undefined) => void
  placeholder?: string
  disabled?: boolean
  className?: string
  fromDate?: Date
  toDate?: Date
  /** Use dropdown for month/year selection (useful for date of birth) */
  captionLayout?: 'dropdown' | 'dropdown-months' | 'dropdown-years' | 'label'
}

export function DatePicker({
  date,
  onDateChange,
  placeholder = 'Pick a date',
  disabled,
  className,
  fromDate,
  toDate,
  captionLayout,
}: DatePickerProps) {
  const [open, setOpen] = React.useState(false)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          disabled={disabled}
          className={cn(
            'w-full justify-start text-left font-normal',
            !date && 'text-muted-foreground',
            className
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4 shrink-0" />
          <span className="truncate">
            {date ? format(date, 'MMM d, yyyy') : placeholder}
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="single"
          selected={date}
          onSelect={(newDate) => {
            onDateChange(newDate)
            setOpen(false)
          }}
          disabled={(day) => {
            if (fromDate && day < fromDate) return true
            if (toDate && day > toDate) return true
            return false
          }}
          captionLayout={captionLayout}
          fromYear={fromDate?.getFullYear() ?? 1900}
          toYear={toDate?.getFullYear() ?? new Date().getFullYear() + 10}
          defaultMonth={date}
          initialFocus
        />
      </PopoverContent>
    </Popover>
  )
}
