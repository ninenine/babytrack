import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useLiveQuery } from 'dexie-react-hooks'
import { db, LocalFeeding, LocalSleep } from '@/db/dexie'
import { useFamilyStore } from '@/stores/family.store'
import { apiClient } from '@/api/client'
import { QuickAddBar } from './QuickAddBar'
import { FeedingModal } from './FeedingModal'

type TimelineEvent = {
  id: string
  type: 'feeding' | 'sleep' | 'medication'
  time: Date
  endTime?: Date
  data: LocalFeeding | LocalSleep | unknown
}

export function Timeline() {
  const navigate = useNavigate()
  const currentChild = useFamilyStore((state) => state.currentChild)
  const [showFeedingModal, setShowFeedingModal] = useState(false)
  const [activeSleep, setActiveSleep] = useState<LocalSleep | null>(null)

  // Get today's date range
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const tomorrow = new Date(today)
  tomorrow.setDate(tomorrow.getDate() + 1)

  // Query today's feedings
  const feedings = useLiveQuery(
    () =>
      currentChild
        ? db.feedings
            .where('childId')
            .equals(currentChild.id)
            .filter((f) => f.startTime >= today && f.startTime < tomorrow)
            .toArray()
        : [],
    [currentChild?.id, today.toDateString()]
  )

  // Query today's sleep
  const sleepSessions = useLiveQuery(
    () =>
      currentChild
        ? db.sleep
            .where('childId')
            .equals(currentChild.id)
            .filter((s) => s.startTime >= today && s.startTime < tomorrow)
            .toArray()
        : [],
    [currentChild?.id, today.toDateString()]
  )

  // Check for active sleep session
  useEffect(() => {
    if (sleepSessions) {
      const active = sleepSessions.find((s) => !s.endTime)
      setActiveSleep(active || null)
    }
  }, [sleepSessions])

  // Combine and sort events
  const events: TimelineEvent[] = [
    ...(feedings || []).map((f) => ({
      id: f.id,
      type: 'feeding' as const,
      time: f.startTime,
      endTime: f.endTime,
      data: f,
    })),
    ...(sleepSessions || []).map((s) => ({
      id: s.id,
      type: 'sleep' as const,
      time: s.startTime,
      endTime: s.endTime,
      data: s,
    })),
  ].sort((a, b) => b.time.getTime() - a.time.getTime())

  // Get last feeding info
  const lastFeeding = feedings?.length ? feedings[feedings.length - 1] : null
  const timeSinceLastFeed = lastFeeding
    ? Math.floor((Date.now() - lastFeeding.startTime.getTime()) / 60000)
    : null

  async function handleQuickFeed() {
    if (!currentChild) return

    // Get last feeding to use as default
    const lastFeed = await db.feedings
      .where('childId')
      .equals(currentChild.id)
      .last()

    const feedingData = {
      child_id: currentChild.id,
      type: lastFeed?.type || 'bottle',
      start_time: new Date().toISOString(),
      side: lastFeed?.side === 'left' ? 'right' : lastFeed?.side === 'right' ? 'left' : undefined,
    }

    try {
      const response = await apiClient.post<{
        id: string
        child_id: string
        type: 'breast' | 'bottle' | 'formula' | 'solid'
        start_time: string
        side?: string
      }>('/api/feeding', feedingData)

      await db.feedings.add({
        id: response.data.id,
        childId: response.data.child_id,
        type: response.data.type,
        startTime: new Date(response.data.start_time),
        side: response.data.side,
        syncedAt: new Date(),
        pendingSync: false,
      })
    } catch {
      // Save locally
      await db.feedings.add({
        id: crypto.randomUUID(),
        childId: currentChild.id,
        type: lastFeed?.type || 'bottle',
        startTime: new Date(),
        side: lastFeed?.side === 'left' ? 'right' : lastFeed?.side === 'right' ? 'left' : undefined,
        pendingSync: true,
      })
    }
  }

  async function handleSleepToggle() {
    if (!currentChild) return

    if (activeSleep) {
      // End sleep
      const endTime = new Date()
      await db.sleep.update(activeSleep.id, { endTime, pendingSync: true })

      try {
        await apiClient.put(`/api/sleep/${activeSleep.id}`, {
          end_time: endTime.toISOString(),
        })
        await db.sleep.update(activeSleep.id, { syncedAt: new Date(), pendingSync: false })
      } catch {
        // Will sync later
      }
    } else {
      // Start sleep
      const sleepData = {
        child_id: currentChild.id,
        type: 'nap',
        start_time: new Date().toISOString(),
      }

      try {
        const response = await apiClient.post<{
          id: string
          child_id: string
          type: string
          start_time: string
        }>('/api/sleep', sleepData)

        await db.sleep.add({
          id: response.data.id,
          childId: response.data.child_id,
          type: response.data.type as 'nap' | 'night',
          startTime: new Date(response.data.start_time),
          syncedAt: new Date(),
          pendingSync: false,
        })
      } catch {
        await db.sleep.add({
          id: crypto.randomUUID(),
          childId: currentChild.id,
          type: 'nap',
          startTime: new Date(),
          pendingSync: true,
        })
      }
    }
  }

  function formatTime(date: Date): string {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  function formatDuration(start: Date, end?: Date): string {
    const endTime = end || new Date()
    const diff = endTime.getTime() - start.getTime()
    const minutes = Math.floor(diff / 60000)
    if (minutes < 60) return `${minutes}m`
    const hours = Math.floor(minutes / 60)
    const mins = minutes % 60
    return `${hours}h ${mins}m`
  }

  function formatTimeSince(minutes: number): string {
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    const mins = minutes % 60
    return mins > 0 ? `${hours}h ${mins}m ago` : `${hours}h ago`
  }

  const feedingLabels: Record<string, string> = {
    breast: 'Breastfed',
    bottle: 'Bottle',
    formula: 'Formula',
    solid: 'Solids',
  }

  if (!currentChild) {
    return (
      <div style={styles.empty}>
        <p>No child selected</p>
      </div>
    )
  }

  return (
    <div style={styles.container}>
      {/* Status Cards */}
      <div style={styles.statusCards}>
        {/* Last Feeding */}
        <div style={styles.statusCard}>
          <div style={styles.statusIcon}>🍼</div>
          <div style={styles.statusInfo}>
            <div style={styles.statusLabel}>Last Feed</div>
            <div style={styles.statusValue}>
              {timeSinceLastFeed !== null
                ? formatTimeSince(timeSinceLastFeed)
                : 'No feeds today'}
            </div>
          </div>
        </div>

        {/* Sleep Status */}
        <div style={{ ...styles.statusCard, ...(activeSleep ? styles.statusCardActive : {}) }}>
          <div style={styles.statusIcon}>😴</div>
          <div style={styles.statusInfo}>
            <div style={styles.statusLabel}>Sleep</div>
            <div style={styles.statusValue}>
              {activeSleep
                ? `Sleeping ${formatDuration(activeSleep.startTime)}`
                : 'Awake'}
            </div>
          </div>
        </div>
      </div>

      {/* Timeline */}
      <div style={styles.timeline}>
        <h2 style={styles.sectionTitle}>Today's Activity</h2>

        {events.length === 0 ? (
          <div style={styles.emptyTimeline}>
            <p>No activity logged yet today.</p>
            <p style={styles.emptyHint}>Use the buttons below to get started!</p>
          </div>
        ) : (
          <div style={styles.eventList}>
            {events.map((event) => (
              <div key={event.id} style={styles.eventCard}>
                <div style={styles.eventTime}>{formatTime(event.time)}</div>
                <div style={styles.eventContent}>
                  {event.type === 'feeding' && (
                    <>
                      <span style={styles.eventIcon}>🍼</span>
                      <div style={styles.eventDetails}>
                        <div style={styles.eventTitle}>
                          {feedingLabels[(event.data as LocalFeeding).type]}
                          {(event.data as LocalFeeding).side && (
                            <span style={styles.eventMeta}>
                              {' '}({(event.data as LocalFeeding).side})
                            </span>
                          )}
                        </div>
                        {(event.data as LocalFeeding).amount && (
                          <div style={styles.eventSubtitle}>
                            {(event.data as LocalFeeding).amount} {(event.data as LocalFeeding).unit}
                          </div>
                        )}
                        {event.endTime && (
                          <div style={styles.eventSubtitle}>
                            {formatDuration(event.time, event.endTime)}
                          </div>
                        )}
                      </div>
                    </>
                  )}
                  {event.type === 'sleep' && (
                    <>
                      <span style={styles.eventIcon}>😴</span>
                      <div style={styles.eventDetails}>
                        <div style={styles.eventTitle}>
                          {(event.data as LocalSleep).type === 'nap' ? 'Nap' : 'Night Sleep'}
                        </div>
                        <div style={styles.eventSubtitle}>
                          {event.endTime
                            ? formatDuration(event.time, event.endTime)
                            : 'In progress...'}
                        </div>
                      </div>
                    </>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Quick Add Bar */}
      <QuickAddBar
        onFeedTap={handleQuickFeed}
        onFeedLongPress={() => setShowFeedingModal(true)}
        onSleepTap={handleSleepToggle}
        onMedsTap={() => navigate('/medications')}
        isSleeping={!!activeSleep}
        lastFeedingType={(feedings?.[0] as LocalFeeding | undefined)?.type}
      />

      {/* Feeding Modal */}
      {showFeedingModal && (
        <FeedingModal
          childId={currentChild.id}
          onClose={() => setShowFeedingModal(false)}
        />
      )}
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    padding: '1rem',
    paddingBottom: '6rem',
  },
  empty: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: '50vh',
    color: 'var(--text-secondary)',
  },
  statusCards: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '0.75rem',
    marginBottom: '1.5rem',
  },
  statusCard: {
    display: 'flex',
    alignItems: 'center',
    gap: '0.75rem',
    padding: '1rem',
    backgroundColor: 'var(--surface)',
    borderRadius: '0.75rem',
    border: '1px solid var(--border)',
  },
  statusCardActive: {
    backgroundColor: 'var(--primary)',
    borderColor: 'var(--primary)',
    color: 'white',
  },
  statusIcon: {
    fontSize: '1.5rem',
  },
  statusInfo: {
    flex: 1,
  },
  statusLabel: {
    fontSize: '0.75rem',
    fontWeight: '500',
    opacity: 0.8,
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
  },
  statusValue: {
    fontSize: '0.875rem',
    fontWeight: '600',
  },
  timeline: {
    marginBottom: '1rem',
  },
  sectionTitle: {
    fontSize: '0.875rem',
    fontWeight: '600',
    color: 'var(--text-secondary)',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
    marginBottom: '1rem',
  },
  emptyTimeline: {
    textAlign: 'center',
    padding: '3rem 1rem',
    color: 'var(--text-secondary)',
  },
  emptyHint: {
    fontSize: '0.875rem',
    marginTop: '0.5rem',
  },
  eventList: {
    display: 'flex',
    flexDirection: 'column',
    gap: '0.5rem',
  },
  eventCard: {
    display: 'flex',
    gap: '1rem',
    padding: '0.875rem',
    backgroundColor: 'var(--surface)',
    borderRadius: '0.75rem',
    border: '1px solid var(--border)',
  },
  eventTime: {
    fontSize: '0.875rem',
    fontWeight: '500',
    color: 'var(--text-secondary)',
    minWidth: '50px',
  },
  eventContent: {
    display: 'flex',
    alignItems: 'flex-start',
    gap: '0.75rem',
    flex: 1,
  },
  eventIcon: {
    fontSize: '1.25rem',
  },
  eventDetails: {
    flex: 1,
  },
  eventTitle: {
    fontWeight: '600',
    fontSize: '0.9375rem',
  },
  eventMeta: {
    fontWeight: '400',
    color: 'var(--text-secondary)',
  },
  eventSubtitle: {
    fontSize: '0.8125rem',
    color: 'var(--text-secondary)',
    marginTop: '0.125rem',
  },
}
