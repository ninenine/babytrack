import { useLiveQuery } from 'dexie-react-hooks'
import { db } from '@/db/dexie'
import { useFamilyStore } from '@/stores/family.store'
import { useTimer } from '@/hooks/useTimers'

export function SleepList() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  const sleepRecords = useLiveQuery(
    () =>
      currentChild
        ? db.sleep
            .where('childId')
            .equals(currentChild.id)
            .reverse()
            .sortBy('startTime')
        : [],
    [currentChild?.id]
  )

  const activeSleep = sleepRecords?.find((s) => !s.endTime)

  return (
    <div style={{ padding: '1rem' }}>
      <header style={{ marginBottom: '1.5rem' }}>
        <h1>Sleep Tracker</h1>
        {currentChild && <p>Tracking for {currentChild.name}</p>}
      </header>

      {activeSleep ? (
        <ActiveSleepCard sleep={activeSleep} />
      ) : (
        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
          <button
            style={{
              backgroundColor: 'var(--primary)',
              color: 'white',
              border: 'none',
              padding: '0.75rem 1.5rem',
              borderRadius: '0.5rem',
              cursor: 'pointer',
            }}
          >
            Start Nap
          </button>
          <button
            style={{
              backgroundColor: 'var(--primary)',
              color: 'white',
              border: 'none',
              padding: '0.75rem 1.5rem',
              borderRadius: '0.5rem',
              cursor: 'pointer',
            }}
          >
            Start Night Sleep
          </button>
        </div>
      )}

      <div>
        {!sleepRecords || sleepRecords.length === 0 ? (
          <p>No sleep records yet.</p>
        ) : (
          <ul style={{ listStyle: 'none', padding: 0 }}>
            {sleepRecords
              .filter((s) => s.endTime)
              .map((sleep) => (
                <li
                  key={sleep.id}
                  style={{
                    border: '1px solid var(--border)',
                    borderRadius: '0.5rem',
                    padding: '1rem',
                    marginBottom: '0.5rem',
                  }}
                >
                  <div style={{ fontWeight: 'bold' }}>{sleep.type}</div>
                  <div style={{ color: 'var(--text-secondary)' }}>
                    {new Date(sleep.startTime).toLocaleString()}
                  </div>
                  {sleep.endTime && (
                    <div>
                      Duration:{' '}
                      {Math.round(
                        (new Date(sleep.endTime).getTime() -
                          new Date(sleep.startTime).getTime()) /
                          60000
                      )}{' '}
                      minutes
                    </div>
                  )}
                </li>
              ))}
          </ul>
        )}
      </div>
    </div>
  )
}

function ActiveSleepCard({ sleep }: { sleep: { startTime: Date; type: string } }) {
  const { formattedElapsed } = useTimer(new Date(sleep.startTime))

  return (
    <div
      style={{
        backgroundColor: 'var(--surface)',
        border: '2px solid var(--primary)',
        borderRadius: '0.5rem',
        padding: '1.5rem',
        marginBottom: '1rem',
        textAlign: 'center',
      }}
    >
      <div style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
        {sleep.type.toUpperCase()} IN PROGRESS
      </div>
      <div style={{ fontSize: '2.5rem', fontWeight: 'bold', margin: '0.5rem 0' }}>
        {formattedElapsed}
      </div>
      <button
        style={{
          backgroundColor: 'var(--error)',
          color: 'white',
          border: 'none',
          padding: '0.75rem 2rem',
          borderRadius: '0.5rem',
          cursor: 'pointer',
        }}
      >
        End Sleep
      </button>
    </div>
  )
}
