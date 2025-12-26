import { useState, useEffect } from 'react'
import { useLiveQuery } from 'dexie-react-hooks'
import { db, LocalFeeding } from '@/db/dexie'
import { useFamilyStore } from '@/stores/family.store'
import { apiClient } from '@/api/client'

type FeedingType = 'breast' | 'bottle' | 'formula' | 'solid'

interface FeedingFormData {
  type: FeedingType
  startTime: string
  endTime: string
  amount: string
  unit: string
  side: string
  notes: string
}

const initialFormData: FeedingFormData = {
  type: 'bottle',
  startTime: new Date().toISOString().slice(0, 16),
  endTime: '',
  amount: '',
  unit: 'ml',
  side: '',
  notes: '',
}

export function FeedingList() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState<FeedingFormData>(initialFormData)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isSyncing, setIsSyncing] = useState(false)

  const feedings = useLiveQuery(
    () =>
      currentChild
        ? db.feedings
            .where('childId')
            .equals(currentChild.id)
            .reverse()
            .sortBy('startTime')
        : [],
    [currentChild?.id]
  )

  // Sync feedings from server on mount
  useEffect(() => {
    if (currentChild) {
      syncFromServer()
    }
  }, [currentChild?.id])

  async function syncFromServer() {
    if (!currentChild) return
    setIsSyncing(true)

    try {
      const response = await apiClient.get<Array<{
        id: string
        child_id: string
        type: FeedingType
        start_time: string
        end_time?: string
        amount?: number
        unit?: string
        side?: string
        notes?: string
      }>>('/api/feeding', { params: { child_id: currentChild.id } })

      // Upsert feedings into local DB
      for (const feeding of response.data) {
        await db.feedings.put({
          id: feeding.id,
          childId: feeding.child_id,
          type: feeding.type,
          startTime: new Date(feeding.start_time),
          endTime: feeding.end_time ? new Date(feeding.end_time) : undefined,
          amount: feeding.amount,
          unit: feeding.unit,
          side: feeding.side,
          notes: feeding.notes,
          syncedAt: new Date(),
          pendingSync: false,
        })
      }
    } catch (error) {
      console.error('Failed to sync feedings:', error)
    } finally {
      setIsSyncing(false)
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!currentChild) return

    setIsSubmitting(true)

    const feedingData = {
      child_id: currentChild.id,
      type: formData.type,
      start_time: new Date(formData.startTime).toISOString(),
      end_time: formData.endTime ? new Date(formData.endTime).toISOString() : undefined,
      amount: formData.amount ? parseFloat(formData.amount) : undefined,
      unit: formData.unit || undefined,
      side: formData.side || undefined,
      notes: formData.notes || undefined,
    }

    try {
      const response = await apiClient.post<{
        id: string
        child_id: string
        type: FeedingType
        start_time: string
        end_time?: string
        amount?: number
        unit?: string
        side?: string
        notes?: string
      }>('/api/feeding', feedingData)

      // Add to local DB
      await db.feedings.add({
        id: response.data.id,
        childId: response.data.child_id,
        type: response.data.type,
        startTime: new Date(response.data.start_time),
        endTime: response.data.end_time ? new Date(response.data.end_time) : undefined,
        amount: response.data.amount,
        unit: response.data.unit,
        side: response.data.side,
        notes: response.data.notes,
        syncedAt: new Date(),
        pendingSync: false,
      })

      setShowForm(false)
      setFormData({ ...initialFormData, startTime: new Date().toISOString().slice(0, 16) })
    } catch (error) {
      console.error('Failed to save feeding:', error)
      // Save locally for later sync
      const localId = crypto.randomUUID()
      await db.feedings.add({
        id: localId,
        childId: currentChild.id,
        type: formData.type,
        startTime: new Date(formData.startTime),
        endTime: formData.endTime ? new Date(formData.endTime) : undefined,
        amount: formData.amount ? parseFloat(formData.amount) : undefined,
        unit: formData.unit || undefined,
        side: formData.side || undefined,
        notes: formData.notes || undefined,
        pendingSync: true,
      })
      setShowForm(false)
      setFormData({ ...initialFormData, startTime: new Date().toISOString().slice(0, 16) })
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this feeding record?')) return

    try {
      await apiClient.delete(`/api/feeding/${id}`)
      await db.feedings.delete(id)
    } catch (error) {
      console.error('Failed to delete feeding:', error)
    }
  }

  function formatDuration(start: Date, end?: Date): string {
    if (!end) return 'In progress'
    const diff = end.getTime() - start.getTime()
    const minutes = Math.floor(diff / 60000)
    if (minutes < 60) return `${minutes} min`
    const hours = Math.floor(minutes / 60)
    const remainingMins = minutes % 60
    return `${hours}h ${remainingMins}m`
  }

  function formatTime(date: Date): string {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  const feedingTypeLabels: Record<FeedingType, string> = {
    breast: 'Breastfeeding',
    bottle: 'Bottle',
    formula: 'Formula',
    solid: 'Solid Food',
  }

  const feedingTypeEmojis: Record<FeedingType, string> = {
    breast: '',
    bottle: '',
    formula: '',
    solid: '',
  }

  return (
    <div style={styles.container}>
      <header style={styles.header}>
        <div>
          <h1 style={styles.title}>Feeding</h1>
          {currentChild && <p style={styles.subtitle}>Tracking for {currentChild.name}</p>}
        </div>
        {isSyncing && <span style={styles.syncBadge}>Syncing...</span>}
      </header>

      <button onClick={() => setShowForm(true)} style={styles.addButton}>
        + Log Feeding
      </button>

      {showForm && (
        <div style={styles.modalOverlay} onClick={() => setShowForm(false)}>
          <div style={styles.modal} onClick={(e) => e.stopPropagation()}>
            <h2 style={styles.modalTitle}>Log Feeding</h2>
            <form onSubmit={handleSubmit}>
              <div style={styles.formGroup}>
                <label style={styles.label}>Type</label>
                <div style={styles.typeGrid}>
                  {(['breast', 'bottle', 'formula', 'solid'] as FeedingType[]).map((type) => (
                    <button
                      key={type}
                      type="button"
                      onClick={() => setFormData({ ...formData, type })}
                      style={{
                        ...styles.typeButton,
                        ...(formData.type === type ? styles.typeButtonActive : {}),
                      }}
                    >
                      <span style={styles.typeEmoji}>{feedingTypeEmojis[type]}</span>
                      <span>{feedingTypeLabels[type]}</span>
                    </button>
                  ))}
                </div>
              </div>

              <div style={styles.formRow}>
                <div style={styles.formGroup}>
                  <label style={styles.label}>Start Time</label>
                  <input
                    type="datetime-local"
                    value={formData.startTime}
                    onChange={(e) => setFormData({ ...formData, startTime: e.target.value })}
                    style={styles.input}
                    required
                  />
                </div>
                <div style={styles.formGroup}>
                  <label style={styles.label}>End Time</label>
                  <input
                    type="datetime-local"
                    value={formData.endTime}
                    onChange={(e) => setFormData({ ...formData, endTime: e.target.value })}
                    style={styles.input}
                  />
                </div>
              </div>

              {formData.type === 'breast' && (
                <div style={styles.formGroup}>
                  <label style={styles.label}>Side</label>
                  <div style={styles.sideButtons}>
                    {['left', 'right', 'both'].map((side) => (
                      <button
                        key={side}
                        type="button"
                        onClick={() => setFormData({ ...formData, side })}
                        style={{
                          ...styles.sideButton,
                          ...(formData.side === side ? styles.sideButtonActive : {}),
                        }}
                      >
                        {side.charAt(0).toUpperCase() + side.slice(1)}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {(formData.type === 'bottle' || formData.type === 'formula') && (
                <div style={styles.formRow}>
                  <div style={styles.formGroup}>
                    <label style={styles.label}>Amount</label>
                    <input
                      type="number"
                      value={formData.amount}
                      onChange={(e) => setFormData({ ...formData, amount: e.target.value })}
                      style={styles.input}
                      placeholder="0"
                      step="5"
                    />
                  </div>
                  <div style={styles.formGroup}>
                    <label style={styles.label}>Unit</label>
                    <select
                      value={formData.unit}
                      onChange={(e) => setFormData({ ...formData, unit: e.target.value })}
                      style={styles.input}
                    >
                      <option value="ml">ml</option>
                      <option value="oz">oz</option>
                    </select>
                  </div>
                </div>
              )}

              <div style={styles.formGroup}>
                <label style={styles.label}>Notes</label>
                <textarea
                  value={formData.notes}
                  onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  style={{ ...styles.input, minHeight: '80px', resize: 'vertical' }}
                  placeholder="Optional notes..."
                />
              </div>

              <div style={styles.formActions}>
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  style={styles.cancelButton}
                >
                  Cancel
                </button>
                <button type="submit" style={styles.submitButton} disabled={isSubmitting}>
                  {isSubmitting ? 'Saving...' : 'Save'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      <div style={styles.feedingList}>
        {!feedings || feedings.length === 0 ? (
          <div style={styles.emptyState}>
            <p>No feedings recorded yet.</p>
            <p style={styles.emptySubtext}>Tap the button above to log a feeding.</p>
          </div>
        ) : (
          groupFeedingsByDate(feedings).map(([date, dayFeedings]) => (
            <div key={date} style={styles.dateGroup}>
              <h3 style={styles.dateHeader}>{date}</h3>
              {dayFeedings.map((feeding) => (
                <div key={feeding.id} style={styles.feedingCard}>
                  <div style={styles.feedingMain}>
                    <div style={styles.feedingType}>
                      <span style={styles.feedingEmoji}>{feedingTypeEmojis[feeding.type]}</span>
                      <span style={styles.feedingTypeLabel}>{feedingTypeLabels[feeding.type]}</span>
                      {feeding.side && (
                        <span style={styles.feedingSide}>({feeding.side})</span>
                      )}
                    </div>
                    <div style={styles.feedingTime}>
                      {formatTime(feeding.startTime)}
                      {feeding.endTime && ` - ${formatTime(feeding.endTime)}`}
                    </div>
                  </div>
                  <div style={styles.feedingDetails}>
                    {feeding.amount && (
                      <span style={styles.feedingAmount}>
                        {feeding.amount} {feeding.unit}
                      </span>
                    )}
                    {feeding.endTime && (
                      <span style={styles.feedingDuration}>
                        {formatDuration(feeding.startTime, feeding.endTime)}
                      </span>
                    )}
                    {feeding.pendingSync && (
                      <span style={styles.pendingBadge}>Pending sync</span>
                    )}
                  </div>
                  {feeding.notes && <p style={styles.feedingNotes}>{feeding.notes}</p>}
                  <button
                    onClick={() => handleDelete(feeding.id)}
                    style={styles.deleteButton}
                    title="Delete"
                  >
                    Delete
                  </button>
                </div>
              ))}
            </div>
          ))
        )}
      </div>
    </div>
  )
}

function groupFeedingsByDate(feedings: LocalFeeding[]): [string, LocalFeeding[]][] {
  const groups: Map<string, LocalFeeding[]> = new Map()

  for (const feeding of feedings) {
    const date = formatDateForGroup(feeding.startTime)
    const existing = groups.get(date) || []
    groups.set(date, [...existing, feeding])
  }

  return Array.from(groups.entries())
}

function formatDateForGroup(date: Date): string {
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  if (date.toDateString() === today.toDateString()) {
    return 'Today'
  } else if (date.toDateString() === yesterday.toDateString()) {
    return 'Yesterday'
  }
  return date.toLocaleDateString([], { weekday: 'long', month: 'long', day: 'numeric' })
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    padding: '1rem',
    maxWidth: '600px',
    margin: '0 auto',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: '1.5rem',
  },
  title: {
    fontSize: '1.5rem',
    fontWeight: 'bold',
    margin: 0,
  },
  subtitle: {
    color: 'var(--text-secondary)',
    margin: '0.25rem 0 0 0',
  },
  syncBadge: {
    fontSize: '0.75rem',
    color: 'var(--text-secondary)',
    backgroundColor: 'var(--surface)',
    padding: '0.25rem 0.5rem',
    borderRadius: '0.25rem',
  },
  addButton: {
    width: '100%',
    backgroundColor: 'var(--primary)',
    color: 'white',
    border: 'none',
    padding: '0.875rem 1.5rem',
    borderRadius: '0.5rem',
    fontSize: '1rem',
    fontWeight: '600',
    cursor: 'pointer',
    marginBottom: '1.5rem',
  },
  modalOverlay: {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    padding: '1rem',
    zIndex: 1000,
  },
  modal: {
    backgroundColor: 'var(--background)',
    borderRadius: '1rem',
    padding: '1.5rem',
    width: '100%',
    maxWidth: '400px',
    maxHeight: '90vh',
    overflow: 'auto',
  },
  modalTitle: {
    fontSize: '1.25rem',
    fontWeight: 'bold',
    marginBottom: '1rem',
  },
  formGroup: {
    marginBottom: '1rem',
  },
  formRow: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '1rem',
  },
  label: {
    display: 'block',
    marginBottom: '0.5rem',
    fontWeight: '500',
    fontSize: '0.875rem',
  },
  input: {
    width: '100%',
    padding: '0.75rem',
    borderRadius: '0.5rem',
    border: '1px solid var(--border)',
    fontSize: '1rem',
    backgroundColor: 'var(--background)',
    color: 'var(--text)',
    boxSizing: 'border-box',
  },
  typeGrid: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '0.5rem',
  },
  typeButton: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    padding: '0.75rem',
    border: '1px solid var(--border)',
    borderRadius: '0.5rem',
    backgroundColor: 'var(--background)',
    cursor: 'pointer',
    fontSize: '0.875rem',
  },
  typeButtonActive: {
    borderColor: 'var(--primary)',
    backgroundColor: 'var(--primary)',
    color: 'white',
  },
  typeEmoji: {
    fontSize: '1.5rem',
    marginBottom: '0.25rem',
  },
  sideButtons: {
    display: 'flex',
    gap: '0.5rem',
  },
  sideButton: {
    flex: 1,
    padding: '0.5rem',
    border: '1px solid var(--border)',
    borderRadius: '0.5rem',
    backgroundColor: 'var(--background)',
    cursor: 'pointer',
  },
  sideButtonActive: {
    borderColor: 'var(--primary)',
    backgroundColor: 'var(--primary)',
    color: 'white',
  },
  formActions: {
    display: 'flex',
    gap: '0.5rem',
    marginTop: '1.5rem',
  },
  cancelButton: {
    flex: 1,
    padding: '0.75rem',
    border: '1px solid var(--border)',
    borderRadius: '0.5rem',
    backgroundColor: 'var(--background)',
    cursor: 'pointer',
    fontSize: '1rem',
  },
  submitButton: {
    flex: 1,
    padding: '0.75rem',
    border: 'none',
    borderRadius: '0.5rem',
    backgroundColor: 'var(--primary)',
    color: 'white',
    cursor: 'pointer',
    fontSize: '1rem',
    fontWeight: '600',
  },
  feedingList: {
    display: 'flex',
    flexDirection: 'column',
    gap: '1rem',
  },
  emptyState: {
    textAlign: 'center',
    padding: '3rem 1rem',
    color: 'var(--text-secondary)',
  },
  emptySubtext: {
    fontSize: '0.875rem',
    marginTop: '0.5rem',
  },
  dateGroup: {
    marginBottom: '1rem',
  },
  dateHeader: {
    fontSize: '0.875rem',
    fontWeight: '600',
    color: 'var(--text-secondary)',
    marginBottom: '0.5rem',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
  },
  feedingCard: {
    border: '1px solid var(--border)',
    borderRadius: '0.75rem',
    padding: '1rem',
    marginBottom: '0.5rem',
    position: 'relative',
  },
  feedingMain: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: '0.5rem',
  },
  feedingType: {
    display: 'flex',
    alignItems: 'center',
    gap: '0.5rem',
  },
  feedingEmoji: {
    fontSize: '1.25rem',
  },
  feedingTypeLabel: {
    fontWeight: '600',
  },
  feedingSide: {
    color: 'var(--text-secondary)',
    fontSize: '0.875rem',
  },
  feedingTime: {
    color: 'var(--text-secondary)',
    fontSize: '0.875rem',
  },
  feedingDetails: {
    display: 'flex',
    gap: '1rem',
    fontSize: '0.875rem',
  },
  feedingAmount: {
    color: 'var(--text)',
    fontWeight: '500',
  },
  feedingDuration: {
    color: 'var(--text-secondary)',
  },
  pendingBadge: {
    fontSize: '0.75rem',
    color: 'var(--warning)',
    backgroundColor: 'var(--warning-bg)',
    padding: '0.125rem 0.5rem',
    borderRadius: '0.25rem',
  },
  feedingNotes: {
    marginTop: '0.5rem',
    fontSize: '0.875rem',
    color: 'var(--text-secondary)',
    fontStyle: 'italic',
  },
  deleteButton: {
    position: 'absolute',
    top: '0.5rem',
    right: '0.5rem',
    padding: '0.25rem 0.5rem',
    fontSize: '0.75rem',
    color: 'var(--error)',
    backgroundColor: 'transparent',
    border: 'none',
    cursor: 'pointer',
    opacity: 0.7,
  },
}
