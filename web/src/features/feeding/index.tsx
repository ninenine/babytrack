import { useLiveQuery } from 'dexie-react-hooks'
import { db } from '@/db/dexie'
import { useFamilyStore } from '@/stores/family.store'

export function FeedingList() {
  const currentChild = useFamilyStore((state) => state.currentChild)

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

  return (
    <div style={{ padding: '1rem' }}>
      <header style={{ marginBottom: '1.5rem' }}>
        <h1>Feeding Tracker</h1>
        {currentChild && <p>Tracking for {currentChild.name}</p>}
      </header>

      <button
        style={{
          backgroundColor: 'var(--primary)',
          color: 'white',
          border: 'none',
          padding: '0.75rem 1.5rem',
          borderRadius: '0.5rem',
          cursor: 'pointer',
          marginBottom: '1rem',
        }}
      >
        + Log Feeding
      </button>

      <div>
        {!feedings || feedings.length === 0 ? (
          <p>No feedings recorded yet.</p>
        ) : (
          <ul style={{ listStyle: 'none', padding: 0 }}>
            {feedings.map((feeding) => (
              <li
                key={feeding.id}
                style={{
                  border: '1px solid var(--border)',
                  borderRadius: '0.5rem',
                  padding: '1rem',
                  marginBottom: '0.5rem',
                }}
              >
                <div style={{ fontWeight: 'bold' }}>{feeding.type}</div>
                <div style={{ color: 'var(--text-secondary)' }}>
                  {new Date(feeding.startTime).toLocaleString()}
                </div>
                {feeding.amount && (
                  <div>
                    {feeding.amount} {feeding.unit}
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
