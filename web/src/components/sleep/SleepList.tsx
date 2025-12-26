import { format, isToday, isYesterday } from 'date-fns'
import { SleepCard } from './SleepCard'
import { Skeleton } from '@/components/ui/skeleton'
import type { LocalSleep } from '@/db/dexie'

interface SleepListProps {
  sleepRecords: LocalSleep[]
  isLoading: boolean
  onEdit?: (sleep: LocalSleep) => void
}

function formatDateHeader(date: Date): string {
  if (isToday(date)) return 'Today'
  if (isYesterday(date)) return 'Yesterday'
  return format(date, 'EEEE, MMMM d')
}

function groupSleepByDate(records: LocalSleep[]): Map<string, LocalSleep[]> {
  const groups = new Map<string, LocalSleep[]>()

  for (const record of records) {
    const dateKey = format(new Date(record.startTime), 'yyyy-MM-dd')
    const existing = groups.get(dateKey) || []
    groups.set(dateKey, [...existing, record])
  }

  return groups
}

export function SleepList({ sleepRecords, isLoading, onEdit }: SleepListProps) {
  // Filter out active sleep (no endTime) for the list
  const completedRecords = sleepRecords.filter((s) => s.endTime)

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="space-y-2">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-20 w-full" />
          </div>
        ))}
      </div>
    )
  }

  if (completedRecords.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p className="text-lg">No sleep records yet</p>
        <p className="text-sm mt-1">Start tracking sleep sessions above</p>
      </div>
    )
  }

  const groupedRecords = groupSleepByDate(completedRecords)

  return (
    <div className="space-y-6">
      {Array.from(groupedRecords.entries()).map(([dateKey, dayRecords]) => {
        const date = new Date(dateKey)
        return (
          <div key={dateKey}>
            <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-3">
              {formatDateHeader(date)}
            </h3>
            <div className="space-y-2">
              {dayRecords.map((record) => (
                <SleepCard key={record.id} sleep={record} onEdit={onEdit} />
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}
