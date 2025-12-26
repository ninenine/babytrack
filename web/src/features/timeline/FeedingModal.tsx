import { useState } from 'react'
import { db } from '@/db/dexie'
import { apiClient } from '@/api/client'

type FeedingType = 'breast' | 'bottle' | 'formula' | 'solid'

interface FeedingModalProps {
  childId: string
  onClose: () => void
}

export function FeedingModal({ childId, onClose }: FeedingModalProps) {
  const [type, setType] = useState<FeedingType>('bottle')
  const [side, setSide] = useState<string>('')
  const [amount, setAmount] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit() {
    setIsSubmitting(true)

    const feedingData = {
      child_id: childId,
      type,
      start_time: new Date().toISOString(),
      end_time: new Date().toISOString(),
      side: side || undefined,
      amount: amount ? parseFloat(amount) : undefined,
      unit: amount ? 'ml' : undefined,
    }

    try {
      const response = await apiClient.post<{
        id: string
        child_id: string
        type: FeedingType
        start_time: string
        end_time?: string
        side?: string
        amount?: number
        unit?: string
      }>('/api/feeding', feedingData)

      await db.feedings.add({
        id: response.data.id,
        childId: response.data.child_id,
        type: response.data.type,
        startTime: new Date(response.data.start_time),
        endTime: response.data.end_time ? new Date(response.data.end_time) : undefined,
        side: response.data.side,
        amount: response.data.amount,
        unit: response.data.unit,
        syncedAt: new Date(),
        pendingSync: false,
      })
    } catch {
      await db.feedings.add({
        id: crypto.randomUUID(),
        childId,
        type,
        startTime: new Date(),
        endTime: new Date(),
        side: side || undefined,
        amount: amount ? parseFloat(amount) : undefined,
        unit: amount ? 'ml' : undefined,
        pendingSync: true,
      })
    }

    setIsSubmitting(false)
    onClose()
  }

  const typeOptions: { value: FeedingType; label: string; icon: string }[] = [
    { value: 'breast', label: 'Breast', icon: '🤱' },
    { value: 'bottle', label: 'Bottle', icon: '🍼' },
    { value: 'formula', label: 'Formula', icon: '🥛' },
    { value: 'solid', label: 'Solids', icon: '🥣' },
  ]

  return (
    <div style={styles.overlay} onClick={onClose}>
      <div style={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div style={styles.header}>
          <h2 style={styles.title}>Log Feeding</h2>
          <button style={styles.closeButton} onClick={onClose}>
            ✕
          </button>
        </div>

        {/* Type Selection */}
        <div style={styles.typeGrid}>
          {typeOptions.map((opt) => (
            <button
              key={opt.value}
              onClick={() => setType(opt.value)}
              style={{
                ...styles.typeButton,
                ...(type === opt.value ? styles.typeButtonActive : {}),
              }}
            >
              <span style={styles.typeIcon}>{opt.icon}</span>
              <span>{opt.label}</span>
            </button>
          ))}
        </div>

        {/* Side Selection (for breast) */}
        {type === 'breast' && (
          <div style={styles.section}>
            <label style={styles.label}>Side</label>
            <div style={styles.sideGrid}>
              {['left', 'right', 'both'].map((s) => (
                <button
                  key={s}
                  onClick={() => setSide(s)}
                  style={{
                    ...styles.sideButton,
                    ...(side === s ? styles.sideButtonActive : {}),
                  }}
                >
                  {s === 'left' ? '← Left' : s === 'right' ? 'Right →' : 'Both'}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Amount (for bottle/formula) */}
        {(type === 'bottle' || type === 'formula') && (
          <div style={styles.section}>
            <label style={styles.label}>Amount (ml)</label>
            <div style={styles.amountGrid}>
              {[30, 60, 90, 120, 150, 180].map((ml) => (
                <button
                  key={ml}
                  onClick={() => setAmount(ml.toString())}
                  style={{
                    ...styles.amountButton,
                    ...(amount === ml.toString() ? styles.amountButtonActive : {}),
                  }}
                >
                  {ml}
                </button>
              ))}
            </div>
            <input
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="Or enter amount..."
              style={styles.input}
            />
          </div>
        )}

        {/* Submit */}
        <button
          onClick={handleSubmit}
          disabled={isSubmitting}
          style={styles.submitButton}
        >
          {isSubmitting ? 'Saving...' : 'Log Feeding'}
        </button>
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  overlay: {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    display: 'flex',
    alignItems: 'flex-end',
    justifyContent: 'center',
    zIndex: 1000,
  },
  modal: {
    width: '100%',
    maxWidth: '480px',
    backgroundColor: 'var(--background)',
    borderRadius: '1.5rem 1.5rem 0 0',
    padding: '1.5rem',
    paddingBottom: 'calc(1.5rem + env(safe-area-inset-bottom))',
    maxHeight: '85vh',
    overflow: 'auto',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '1.5rem',
  },
  title: {
    fontSize: '1.25rem',
    fontWeight: '700',
    margin: 0,
  },
  closeButton: {
    width: '2rem',
    height: '2rem',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: 'var(--surface)',
    border: 'none',
    borderRadius: '50%',
    fontSize: '1rem',
    cursor: 'pointer',
  },
  typeGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(4, 1fr)',
    gap: '0.5rem',
    marginBottom: '1.5rem',
  },
  typeButton: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: '0.25rem',
    padding: '0.75rem 0.25rem',
    border: '2px solid var(--border)',
    borderRadius: '0.75rem',
    backgroundColor: 'var(--background)',
    cursor: 'pointer',
    fontSize: '0.75rem',
    fontWeight: '600',
  },
  typeButtonActive: {
    borderColor: 'var(--primary)',
    backgroundColor: 'var(--primary)',
    color: 'white',
  },
  typeIcon: {
    fontSize: '1.5rem',
  },
  section: {
    marginBottom: '1.5rem',
  },
  label: {
    display: 'block',
    fontSize: '0.875rem',
    fontWeight: '600',
    color: 'var(--text-secondary)',
    marginBottom: '0.75rem',
  },
  sideGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, 1fr)',
    gap: '0.5rem',
  },
  sideButton: {
    padding: '0.75rem',
    border: '2px solid var(--border)',
    borderRadius: '0.75rem',
    backgroundColor: 'var(--background)',
    cursor: 'pointer',
    fontSize: '0.875rem',
    fontWeight: '600',
  },
  sideButtonActive: {
    borderColor: 'var(--primary)',
    backgroundColor: 'var(--primary)',
    color: 'white',
  },
  amountGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(6, 1fr)',
    gap: '0.5rem',
    marginBottom: '0.75rem',
  },
  amountButton: {
    padding: '0.5rem',
    border: '2px solid var(--border)',
    borderRadius: '0.5rem',
    backgroundColor: 'var(--background)',
    cursor: 'pointer',
    fontSize: '0.875rem',
    fontWeight: '600',
  },
  amountButtonActive: {
    borderColor: 'var(--primary)',
    backgroundColor: 'var(--primary)',
    color: 'white',
  },
  input: {
    width: '100%',
    padding: '0.75rem',
    border: '1px solid var(--border)',
    borderRadius: '0.5rem',
    fontSize: '1rem',
    backgroundColor: 'var(--background)',
    boxSizing: 'border-box',
  },
  submitButton: {
    width: '100%',
    padding: '1rem',
    backgroundColor: 'var(--primary)',
    color: 'white',
    border: 'none',
    borderRadius: '0.75rem',
    fontSize: '1rem',
    fontWeight: '700',
    cursor: 'pointer',
  },
}
