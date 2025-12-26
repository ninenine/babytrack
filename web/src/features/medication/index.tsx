import { useLiveQuery } from 'dexie-react-hooks'
import { db } from '@/db/dexie'
import { useFamilyStore } from '@/stores/family.store'

export function MedicationList() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  const medications = useLiveQuery(
    () =>
      currentChild
        ? db.medications
            .where('childId')
            .equals(currentChild.id)
            .and((m) => m.active)
            .toArray()
        : [],
    [currentChild?.id]
  )

  return (
    <div style={{ padding: '1rem' }}>
      <header style={{ marginBottom: '1.5rem' }}>
        <h1>Medication Tracker</h1>
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
        + Add Medication
      </button>

      <div>
        {!medications || medications.length === 0 ? (
          <p>No active medications.</p>
        ) : (
          <ul style={{ listStyle: 'none', padding: 0 }}>
            {medications.map((med) => (
              <li
                key={med.id}
                style={{
                  border: '1px solid var(--border)',
                  borderRadius: '0.5rem',
                  padding: '1rem',
                  marginBottom: '0.5rem',
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                  }}
                >
                  <div>
                    <div style={{ fontWeight: 'bold' }}>{med.name}</div>
                    <div style={{ color: 'var(--text-secondary)' }}>
                      {med.dosage} {med.unit} - {med.frequency}
                    </div>
                    {med.instructions && (
                      <div style={{ fontSize: '0.875rem', marginTop: '0.25rem' }}>
                        {med.instructions}
                      </div>
                    )}
                  </div>
                  <button
                    style={{
                      backgroundColor: 'var(--success)',
                      color: 'white',
                      border: 'none',
                      padding: '0.5rem 1rem',
                      borderRadius: '0.25rem',
                      cursor: 'pointer',
                    }}
                  >
                    Log Dose
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
