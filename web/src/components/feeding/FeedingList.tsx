import { format, isToday, isYesterday } from 'date-fns'
import { FeedingCard } from './FeedingCard'
import { Skeleton } from '@/components/ui/skeleton'
import type { LocalFeeding } from '@/db/dexie'

interface FeedingListProps {
  feedings: LocalFeeding[]
  isLoading: boolean
  onEdit?: (feeding: LocalFeeding) => void
}

function formatDateHeader(date: Date): string {
  if (isToday(date)) return 'Today'
  if (isYesterday(date)) return 'Yesterday'
  return format(date, 'EEEE, MMMM d')
}

function groupFeedingsByDate(feedings: LocalFeeding[]): Map<string, LocalFeeding[]> {
  const groups = new Map<string, LocalFeeding[]>()

  for (const feeding of feedings) {
    const dateKey = format(new Date(feeding.startTime), 'yyyy-MM-dd')
    const existing = groups.get(dateKey) || []
    groups.set(dateKey, [...existing, feeding])
  }

  return groups
}

export function FeedingList({ feedings, isLoading, onEdit }: FeedingListProps) {
  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="space-y-2">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
          </div>
        ))}
      </div>
    )
  }

  if (feedings.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p className="text-lg">No feedings recorded yet</p>
        <p className="text-sm mt-1">Tap + Add to log a feeding</p>
      </div>
    )
  }

  const groupedFeedings = groupFeedingsByDate(feedings)

  return (
    <div className="space-y-6">
      {Array.from(groupedFeedings.entries()).map(([dateKey, dayFeedings]) => {
        const date = new Date(dateKey)
        return (
          <div key={dateKey}>
            <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-3">
              {formatDateHeader(date)}
            </h3>
            <div className="space-y-2">
              {dayFeedings.map((feeding) => (
                <FeedingCard key={feeding.id} feeding={feeding} onEdit={onEdit} />
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}
