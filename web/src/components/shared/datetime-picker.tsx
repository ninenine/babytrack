import * as React from 'react'
import { format } from 'date-fns'
import { CalendarIcon, Clock } from 'lucide-react'

import { cn } from '@/lib/utils'
import { getTimezoneAbbr } from '@/lib/dates'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Input } from '@/components/ui/input'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

interface DateTimePickerProps {
  date?: Date
  onDateChange: (date: Date | undefined) => void
  placeholder?: string
  disabled?: boolean
  className?: string
  /** Show timezone abbreviation in the display */
  showTimezone?: boolean
  /** Maximum date that can be selected */
  toDate?: Date
  /** Minimum date that can be selected */
  fromDate?: Date
}

export function DateTimePicker({
  date,
  onDateChange,
  placeholder = 'Pick date & time',
  disabled,
  className,
  showTimezone = true,
  toDate,
  fromDate,
}: DateTimePickerProps) {
  const [open, setOpen] = React.useState(false)

  const timeValue = date ? format(date, 'HH:mm') : ''
  const tzAbbr = getTimezoneAbbr()

  // Build disabled matcher for calendar
  const disabledMatcher = React.useMemo(() => {
    if (toDate && fromDate) {
      return { after: toDate, before: fromDate }
    }
    if (toDate) {
      return { after: toDate }
    }
    if (fromDate) {
      return { before: fromDate }
    }
    return undefined
  }, [toDate, fromDate])

  const handleDateSelect = (selectedDate: Date | undefined) => {
    if (!selectedDate) {
      onDateChange(undefined)
      return
    }

    // Preserve the existing time if we have one, otherwise use current time
    if (date) {
      selectedDate.setHours(date.getHours(), date.getMinutes(), 0, 0)
    } else {
      const now = new Date()
      selectedDate.setHours(now.getHours(), now.getMinutes(), 0, 0)
    }
    onDateChange(selectedDate)
  }

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const [hours, minutes] = e.target.value.split(':').map(Number)
    if (isNaN(hours) || isNaN(minutes)) return

    const newDate = date ? new Date(date) : new Date()
    newDate.setHours(hours, minutes, 0, 0)
    onDateChange(newDate)
  }

  // Format: "Dec 27, 10:30 AM" or with timezone "Dec 27, 10:30 AM EAT"
  const displayValue = date
    ? `${format(date, 'MMM d, h:mm a')}${showTimezone ? ` ${tzAbbr}` : ''}`
    : null

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
            {displayValue || placeholder}
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="single"
          selected={date}
          onSelect={handleDateSelect}
          defaultMonth={date}
          disabled={disabledMatcher}
          initialFocus
        />
        <div className="border-t p-3">
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-muted-foreground" />
            <Input
              type="time"
              value={timeValue}
              onChange={handleTimeChange}
              className="w-30"
            />
            <span className="text-xs text-muted-foreground">{tzAbbr}</span>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}
