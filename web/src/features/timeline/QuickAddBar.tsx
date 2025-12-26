import { useRef } from 'react'

interface QuickAddBarProps {
  onFeedTap: () => void
  onFeedLongPress: () => void
  onSleepTap: () => void
  isSleeping: boolean
  lastFeedingType?: string
}

export function QuickAddBar({
  onFeedTap,
  onFeedLongPress,
  onSleepTap,
  isSleeping,
  lastFeedingType,
}: QuickAddBarProps) {
  const feedPressTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const feedPressed = useRef(false)

  function handleFeedTouchStart() {
    feedPressed.current = false
    feedPressTimer.current = setTimeout(() => {
      feedPressed.current = true
      onFeedLongPress()
    }, 500)
  }

  function handleFeedTouchEnd() {
    if (feedPressTimer.current) {
      clearTimeout(feedPressTimer.current)
    }
    if (!feedPressed.current) {
      onFeedTap()
    }
  }

  const feedingLabels: Record<string, string> = {
    breast: 'Breast',
    bottle: 'Bottle',
    formula: 'Formula',
    solid: 'Solids',
  }

  return (
    <div style={styles.container}>
      <div style={styles.hint}>
        Tap to log quickly, hold for options
      </div>
      <div style={styles.bar}>
        {/* Feed Button */}
        <button
          style={styles.button}
          onTouchStart={handleFeedTouchStart}
          onTouchEnd={handleFeedTouchEnd}
          onMouseDown={handleFeedTouchStart}
          onMouseUp={handleFeedTouchEnd}
          onMouseLeave={() => feedPressTimer.current && clearTimeout(feedPressTimer.current)}
        >
          <span style={styles.buttonIcon}>🍼</span>
          <span style={styles.buttonLabel}>
            {lastFeedingType ? feedingLabels[lastFeedingType] : 'Feed'}
          </span>
        </button>

        {/* Sleep Button */}
        <button
          style={{
            ...styles.button,
            ...(isSleeping ? styles.buttonActive : {}),
          }}
          onClick={onSleepTap}
        >
          <span style={styles.buttonIcon}>{isSleeping ? '⏰' : '😴'}</span>
          <span style={styles.buttonLabel}>
            {isSleeping ? 'Wake' : 'Sleep'}
          </span>
        </button>

        {/* Meds Button (placeholder) */}
        <button style={styles.button} disabled>
          <span style={styles.buttonIcon}>💊</span>
          <span style={styles.buttonLabel}>Meds</span>
        </button>
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    position: 'fixed',
    bottom: '4rem', // Above the nav bar
    left: '50%',
    transform: 'translateX(-50%)',
    width: '100%',
    maxWidth: '480px',
    padding: '0 1rem',
    paddingBottom: 'env(safe-area-inset-bottom)',
    zIndex: 50,
  },
  hint: {
    textAlign: 'center',
    fontSize: '0.6875rem',
    color: 'var(--text-secondary)',
    marginBottom: '0.5rem',
  },
  bar: {
    display: 'flex',
    gap: '0.5rem',
    padding: '0.5rem',
    backgroundColor: 'var(--surface)',
    borderRadius: '1rem',
    border: '1px solid var(--border)',
    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
  },
  button: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: '0.25rem',
    padding: '0.75rem 0.5rem',
    backgroundColor: 'var(--background)',
    border: '1px solid var(--border)',
    borderRadius: '0.75rem',
    cursor: 'pointer',
    transition: 'all 0.15s ease',
    userSelect: 'none',
    WebkitTapHighlightColor: 'transparent',
  },
  buttonActive: {
    backgroundColor: 'var(--primary)',
    borderColor: 'var(--primary)',
    color: 'white',
  },
  buttonIcon: {
    fontSize: '1.5rem',
  },
  buttonLabel: {
    fontSize: '0.75rem',
    fontWeight: '600',
  },
}
