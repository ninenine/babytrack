import { useMemo } from 'react'
import { format, isToday, isYesterday, differenceInMinutes } from 'date-fns'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Utensils, Moon, FileText, Pill, User } from 'lucide-react'
import { useFamilyStore } from '@/stores/family.store'
import { useFeedings, useSleep, useNotes, useAllMedicationLogs } from '@/hooks'
import type { LocalFeeding, LocalSleep, LocalNote, LocalMedicationLog } from '@/db/dexie'

type TimelineEventType = 'feeding' | 'sleep' | 'note' | 'medication'

interface EnrichedMedicationLog extends LocalMedicationLog {
  medicationName: string
}

interface TimelineEvent {
  id: string
  type: TimelineEventType
  time: Date
  title: string
  subtitle?: string
  badge?: string
  badgeVariant?: 'default' | 'secondary' | 'outline'
  loggedBy?: string
  data: LocalFeeding | LocalSleep | LocalNote | EnrichedMedicationLog
}

const feedingTypeLabels: Record<string, string> = {
  breast: 'Breastfeeding',
  bottle: 'Bottle',
  formula: 'Formula',
  solid: 'Solid food',
}

function formatDuration(start: Date, end: Date): string {
  const minutes = differenceInMinutes(end, start)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  const remainingMins = minutes % 60
  return remainingMins > 0 ? `${hours}h ${remainingMins}m` : `${hours}h`
}

function formatDateHeader(date: Date): string {
  if (isToday(date)) return 'Today'
  if (isYesterday(date)) return 'Yesterday'
  return format(date, 'EEEE, MMMM d')
}

function TimelineEventCard({ event }: { event: TimelineEvent }) {
  const getIcon = () => {
    switch (event.type) {
      case 'feeding':
        return <Utensils className="h-4 w-4" />
      case 'sleep':
        return <Moon className="h-4 w-4" />
      case 'note':
        return <FileText className="h-4 w-4" />
      case 'medication':
        return <Pill className="h-4 w-4" />
      default:
        return null
    }
  }

  const getIconBgColor = () => {
    switch (event.type) {
      case 'feeding':
        return 'bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-400'
      case 'sleep':
        return 'bg-indigo-100 text-indigo-600 dark:bg-indigo-900 dark:text-indigo-400'
      case 'note':
        return 'bg-amber-100 text-amber-600 dark:bg-amber-900 dark:text-amber-400'
      case 'medication':
        return 'bg-green-100 text-green-600 dark:bg-green-900 dark:text-green-400'
      default:
        return 'bg-muted'
    }
  }

  return (
    <Card>
      <CardContent className="flex items-center gap-3 py-3 px-4">
        <div className={`p-2 rounded-full ${getIconBgColor()}`}>
          {getIcon()}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium truncate">{event.title}</span>
            {event.badge && (
              <Badge variant={event.badgeVariant || 'secondary'} className="text-xs shrink-0">
                {event.badge}
              </Badge>
            )}
          </div>
          {event.subtitle && (
            <div className="text-xs text-muted-foreground truncate">{event.subtitle}</div>
          )}
          {event.loggedBy && (
            <div className="flex items-center gap-1 text-xs text-muted-foreground mt-0.5">
              <User className="h-3 w-3" />
              <span>{event.loggedBy}</span>
            </div>
          )}
        </div>
        <div className="text-xs text-muted-foreground shrink-0">
          {format(event.time, 'h:mm a')}
        </div>
      </CardContent>
    </Card>
  )
}

export function TimelineTab() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const { feedings, isLoading: feedingsLoading } = useFeedings()
  const { sleepRecords, isLoading: sleepLoading } = useSleep()
  const { notes, isLoading: notesLoading } = useNotes()
  const { logs: medicationLogs, isLoading: medsLoading } = useAllMedicationLogs()

  const isLoading = feedingsLoading || sleepLoading || notesLoading || medsLoading

  // Combine all events into a unified timeline
  const events = useMemo(() => {
    const allEvents: TimelineEvent[] = []

    // Add feedings
    feedings.forEach((feeding) => {
      const duration = feeding.endTime
        ? formatDuration(new Date(feeding.startTime), new Date(feeding.endTime))
        : undefined

      allEvents.push({
        id: `feeding-${feeding.id}`,
        type: 'feeding',
        time: new Date(feeding.startTime),
        title: feedingTypeLabels[feeding.type] || feeding.type,
        subtitle: [
          feeding.side && `${feeding.side} side`,
          feeding.amount && `${feeding.amount}${feeding.unit}`,
          duration,
        ]
          .filter(Boolean)
          .join(' · '),
        badge: feeding.type === 'breast' && feeding.side ? feeding.side : undefined,
        data: feeding,
      })
    })

    // Add completed sleep records
    sleepRecords.forEach((sleep) => {
      if (!sleep.endTime) return // Skip active sleep

      const duration = formatDuration(new Date(sleep.startTime), new Date(sleep.endTime))

      allEvents.push({
        id: `sleep-${sleep.id}`,
        type: 'sleep',
        time: new Date(sleep.startTime),
        title: sleep.type === 'nap' ? 'Nap' : 'Night Sleep',
        subtitle: `${format(new Date(sleep.startTime), 'h:mm a')} - ${format(new Date(sleep.endTime), 'h:mm a')}`,
        badge: duration,
        data: sleep,
      })
    })

    // Add notes
    notes.forEach((note) => {
      allEvents.push({
        id: `note-${note.id}`,
        type: 'note',
        time: new Date(note.syncedAt || Date.now()),
        title: note.title || 'Note',
        subtitle: note.content.length > 50 ? `${note.content.substring(0, 50)}...` : note.content,
        badge: note.pinned ? 'Pinned' : undefined,
        badgeVariant: 'outline' as const,
        data: note,
      })
    })

    // Add medication logs
    medicationLogs.forEach((log) => {
      allEvents.push({
        id: `medication-${log.id}`,
        type: 'medication',
        time: new Date(log.givenAt),
        title: log.medicationName,
        subtitle: log.dosage,
        badge: 'Dose',
        badgeVariant: 'secondary' as const,
        loggedBy: log.givenBy,
        data: log,
      })
    })

    // Sort by time descending
    return allEvents.sort((a, b) => b.time.getTime() - a.time.getTime())
  }, [feedings, sleepRecords, notes, medicationLogs])

  // Group events by date
  const groupedEvents = useMemo(() => {
    const groups = new Map<string, TimelineEvent[]>()

    events.forEach((event) => {
      const dateKey = format(event.time, 'yyyy-MM-dd')
      const existing = groups.get(dateKey) || []
      groups.set(dateKey, [...existing, event])
    })

    return groups
  }, [events])

  if (!currentChild) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No child selected
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-16 w-full" />
      </div>
    )
  }

  if (events.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-muted-foreground mb-2">No activities logged yet</div>
        <p className="text-sm text-muted-foreground">
          Start logging {currentChild.name}'s activities to see them here
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {Array.from(groupedEvents.entries()).map(([dateKey, dayEvents]) => {
        const date = new Date(dateKey)
        return (
          <div key={dateKey}>
            <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-3">
              {formatDateHeader(date)}
            </h3>
            <div className="space-y-2">
              {dayEvents.map((event) => (
                <TimelineEventCard key={event.id} event={event} />
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}
