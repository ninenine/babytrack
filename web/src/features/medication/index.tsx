import { useState, useEffect } from 'react'
import { useLiveQuery } from 'dexie-react-hooks'
import { db, LocalMedication, LocalMedicationLog } from '@/db/dexie'
import { useFamilyStore } from '@/stores/family.store'
import { useSessionStore } from '@/stores/session.store'
import { apiClient } from '@/api/client'

type MedicationFormData = {
  name: string
  dosage: string
  unit: string
  frequency: string
  instructions: string
  startDate: string
}

const frequencyOptions = [
  { value: 'once_daily', label: 'Once daily' },
  { value: 'twice_daily', label: 'Twice daily' },
  { value: 'three_times_daily', label: 'Three times daily' },
  { value: 'four_times_daily', label: 'Four times daily' },
  { value: 'every_4_hours', label: 'Every 4 hours' },
  { value: 'every_6_hours', label: 'Every 6 hours' },
  { value: 'every_8_hours', label: 'Every 8 hours' },
  { value: 'as_needed', label: 'As needed' },
]

const unitOptions = ['mg', 'ml', 'g', 'drops', 'tablets', 'capsules', 'tsp', 'tbsp']

export function MedicationList() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const user = useSessionStore((state) => state.user)
  const [showAddForm, setShowAddForm] = useState(false)
  const [showLogForm, setShowLogForm] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formData, setFormData] = useState<MedicationFormData>({
    name: '',
    dosage: '',
    unit: 'mg',
    frequency: 'once_daily',
    instructions: '',
    startDate: new Date().toISOString().split('T')[0],
  })
  const [lastLogs, setLastLogs] = useState<Record<string, LocalMedicationLog | null>>({})

  const medications = useLiveQuery(
    () =>
      currentChild
        ? db.medications
            .where('childId')
            .equals(currentChild.id)
            .filter((m) => m.active)
            .toArray()
        : [],
    [currentChild?.id]
  )

  // Sync from server on mount
  useEffect(() => {
    if (currentChild) {
      syncFromServer()
    }
  }, [currentChild?.id])

  // Load last logs for each medication
  useEffect(() => {
    if (medications) {
      medications.forEach(async (med) => {
        const lastLog = await db.medicationLogs
          .where('medicationId')
          .equals(med.id)
          .reverse()
          .sortBy('givenAt')
          .then((logs) => logs[0] || null)
        setLastLogs((prev) => ({ ...prev, [med.id]: lastLog }))
      })
    }
  }, [medications])

  async function syncFromServer() {
    if (!currentChild) return
    try {
      const response = await apiClient.get<Array<{
        id: string
        child_id: string
        name: string
        dosage: string
        unit: string
        frequency: string
        instructions?: string
        start_date: string
        end_date?: string
        active: boolean
      }>>(`/api/medications?child_id=${currentChild.id}&active_only=true`)

      for (const med of response.data) {
        await db.medications.put({
          id: med.id,
          childId: med.child_id,
          name: med.name,
          dosage: med.dosage,
          unit: med.unit,
          frequency: med.frequency,
          instructions: med.instructions,
          startDate: new Date(med.start_date),
          endDate: med.end_date ? new Date(med.end_date) : undefined,
          active: med.active,
          syncedAt: new Date(),
          pendingSync: false,
        })
      }
    } catch {
      // Offline - use local data
    }
  }

  async function handleAddMedication(e: React.FormEvent) {
    e.preventDefault()
    if (!currentChild) return

    setIsSubmitting(true)
    try {
      const response = await apiClient.post<{
        id: string
        child_id: string
        name: string
        dosage: string
        unit: string
        frequency: string
        instructions?: string
        start_date: string
        active: boolean
      }>('/api/medications', {
        child_id: currentChild.id,
        name: formData.name,
        dosage: formData.dosage,
        unit: formData.unit,
        frequency: formData.frequency,
        instructions: formData.instructions || undefined,
        start_date: new Date(formData.startDate).toISOString(),
      })

      await db.medications.add({
        id: response.data.id,
        childId: response.data.child_id,
        name: response.data.name,
        dosage: response.data.dosage,
        unit: response.data.unit,
        frequency: response.data.frequency,
        instructions: response.data.instructions,
        startDate: new Date(response.data.start_date),
        active: response.data.active,
        syncedAt: new Date(),
        pendingSync: false,
      })
    } catch {
      // Save locally for later sync
      await db.medications.add({
        id: crypto.randomUUID(),
        childId: currentChild.id,
        name: formData.name,
        dosage: formData.dosage,
        unit: formData.unit,
        frequency: formData.frequency,
        instructions: formData.instructions || undefined,
        startDate: new Date(formData.startDate),
        active: true,
        pendingSync: true,
      })
    }

    setShowAddForm(false)
    setFormData({
      name: '',
      dosage: '',
      unit: 'mg',
      frequency: 'once_daily',
      instructions: '',
      startDate: new Date().toISOString().split('T')[0],
    })
    setIsSubmitting(false)
  }

  async function handleLogDose(medication: LocalMedication) {
    if (!user) return

    setIsSubmitting(true)
    try {
      const response = await apiClient.post<{
        id: string
        medication_id: string
        child_id: string
        given_at: string
        given_by: string
        dosage: string
      }>('/api/medications/log', {
        medication_id: medication.id,
        given_at: new Date().toISOString(),
        dosage: medication.dosage,
      })

      const log: LocalMedicationLog = {
        id: response.data.id,
        medicationId: response.data.medication_id,
        childId: response.data.child_id,
        givenAt: new Date(response.data.given_at),
        givenBy: response.data.given_by,
        dosage: response.data.dosage,
        syncedAt: new Date(),
        pendingSync: false,
      }

      await db.medicationLogs.add(log)
      setLastLogs((prev) => ({ ...prev, [medication.id]: log }))
    } catch {
      // Save locally
      const log: LocalMedicationLog = {
        id: crypto.randomUUID(),
        medicationId: medication.id,
        childId: medication.childId,
        givenAt: new Date(),
        givenBy: user.id,
        dosage: medication.dosage,
        pendingSync: true,
      }
      await db.medicationLogs.add(log)
      setLastLogs((prev) => ({ ...prev, [medication.id]: log }))
    }
    setIsSubmitting(false)
    setShowLogForm(null)
  }

  async function handleDeactivate(medicationId: string) {
    try {
      await apiClient.post(`/api/medications/${medicationId}/deactivate`)
      await db.medications.update(medicationId, { active: false, endDate: new Date() })
    } catch {
      // Update locally
      await db.medications.update(medicationId, { active: false, endDate: new Date(), pendingSync: true })
    }
  }

  function formatTimeSince(date: Date): string {
    const diff = Date.now() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `${hours}h ago`
    const days = Math.floor(hours / 24)
    return `${days}d ago`
  }

  function getFrequencyLabel(value: string): string {
    return frequencyOptions.find((f) => f.value === value)?.label || value
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
      <header style={styles.header}>
        <h2 style={styles.title}>Medications</h2>
        <button style={styles.addButton} onClick={() => setShowAddForm(true)}>
          + Add
        </button>
      </header>

      {!medications || medications.length === 0 ? (
        <div style={styles.emptyState}>
          <p>No active medications.</p>
          <p style={styles.emptyHint}>Tap + Add to add a medication</p>
        </div>
      ) : (
        <div style={styles.list}>
          {medications.map((med) => {
            const lastLog = lastLogs[med.id]
            return (
              <div key={med.id} style={styles.card}>
                <div style={styles.cardHeader}>
                  <div style={styles.medName}>{med.name}</div>
                  {med.pendingSync && <span style={styles.syncBadge}>Pending sync</span>}
                </div>
                <div style={styles.medDetails}>
                  {med.dosage} {med.unit} • {getFrequencyLabel(med.frequency)}
                </div>
                {med.instructions && (
                  <div style={styles.instructions}>{med.instructions}</div>
                )}
                {lastLog && (
                  <div style={styles.lastDose}>
                    Last dose: {formatTimeSince(lastLog.givenAt)}
                  </div>
                )}
                <div style={styles.cardActions}>
                  <button
                    style={styles.logButton}
                    onClick={() => handleLogDose(med)}
                    disabled={isSubmitting}
                  >
                    Log Dose
                  </button>
                  <button
                    style={styles.menuButton}
                    onClick={() => setShowLogForm(showLogForm === med.id ? null : med.id)}
                  >
                    •••
                  </button>
                </div>
                {showLogForm === med.id && (
                  <div style={styles.menu}>
                    <button
                      style={styles.menuItem}
                      onClick={() => handleDeactivate(med.id)}
                    >
                      Deactivate Medication
                    </button>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      {/* Add Medication Modal */}
      {showAddForm && (
        <div style={styles.modalOverlay} onClick={() => setShowAddForm(false)}>
          <div style={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div style={styles.modalHeader}>
              <h3 style={styles.modalTitle}>Add Medication</h3>
              <button style={styles.closeButton} onClick={() => setShowAddForm(false)}>
                ×
              </button>
            </div>
            <form onSubmit={handleAddMedication} style={styles.form}>
              <div style={styles.formGroup}>
                <label style={styles.label}>Medication Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  style={styles.input}
                  placeholder="e.g., Amoxicillin"
                  required
                />
              </div>

              <div style={styles.formRow}>
                <div style={styles.formGroup}>
                  <label style={styles.label}>Dosage</label>
                  <input
                    type="text"
                    value={formData.dosage}
                    onChange={(e) => setFormData({ ...formData, dosage: e.target.value })}
                    style={styles.input}
                    placeholder="e.g., 250"
                    required
                  />
                </div>
                <div style={styles.formGroup}>
                  <label style={styles.label}>Unit</label>
                  <select
                    value={formData.unit}
                    onChange={(e) => setFormData({ ...formData, unit: e.target.value })}
                    style={styles.select}
                  >
                    {unitOptions.map((unit) => (
                      <option key={unit} value={unit}>
                        {unit}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div style={styles.formGroup}>
                <label style={styles.label}>Frequency</label>
                <select
                  value={formData.frequency}
                  onChange={(e) => setFormData({ ...formData, frequency: e.target.value })}
                  style={styles.select}
                >
                  {frequencyOptions.map((freq) => (
                    <option key={freq.value} value={freq.value}>
                      {freq.label}
                    </option>
                  ))}
                </select>
              </div>

              <div style={styles.formGroup}>
                <label style={styles.label}>Instructions (optional)</label>
                <input
                  type="text"
                  value={formData.instructions}
                  onChange={(e) => setFormData({ ...formData, instructions: e.target.value })}
                  style={styles.input}
                  placeholder="e.g., Take with food"
                />
              </div>

              <div style={styles.formGroup}>
                <label style={styles.label}>Start Date</label>
                <input
                  type="date"
                  value={formData.startDate}
                  onChange={(e) => setFormData({ ...formData, startDate: e.target.value })}
                  style={styles.input}
                  required
                />
              </div>

              <button type="submit" style={styles.submitButton} disabled={isSubmitting}>
                {isSubmitting ? 'Adding...' : 'Add Medication'}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    padding: '1rem',
  },
  empty: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: '50vh',
    color: 'var(--text-secondary)',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '1rem',
  },
  title: {
    fontSize: '1.25rem',
    fontWeight: '600',
    margin: 0,
  },
  addButton: {
    backgroundColor: 'var(--primary)',
    color: 'white',
    border: 'none',
    padding: '0.5rem 1rem',
    borderRadius: '0.5rem',
    cursor: 'pointer',
    fontWeight: '500',
  },
  emptyState: {
    textAlign: 'center',
    padding: '3rem 1rem',
    color: 'var(--text-secondary)',
  },
  emptyHint: {
    fontSize: '0.875rem',
    marginTop: '0.5rem',
  },
  list: {
    display: 'flex',
    flexDirection: 'column',
    gap: '0.75rem',
  },
  card: {
    backgroundColor: 'var(--surface)',
    border: '1px solid var(--border)',
    borderRadius: '0.75rem',
    padding: '1rem',
  },
  cardHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '0.25rem',
  },
  medName: {
    fontWeight: '600',
    fontSize: '1rem',
  },
  syncBadge: {
    fontSize: '0.75rem',
    color: 'var(--warning)',
    backgroundColor: 'rgba(255, 193, 7, 0.1)',
    padding: '0.125rem 0.5rem',
    borderRadius: '0.25rem',
  },
  medDetails: {
    color: 'var(--text-secondary)',
    fontSize: '0.875rem',
  },
  instructions: {
    fontSize: '0.8125rem',
    color: 'var(--text-secondary)',
    marginTop: '0.5rem',
    fontStyle: 'italic',
  },
  lastDose: {
    fontSize: '0.8125rem',
    color: 'var(--primary)',
    marginTop: '0.5rem',
  },
  cardActions: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: '0.75rem',
    paddingTop: '0.75rem',
    borderTop: '1px solid var(--border)',
  },
  logButton: {
    backgroundColor: 'var(--success)',
    color: 'white',
    border: 'none',
    padding: '0.5rem 1.5rem',
    borderRadius: '0.5rem',
    cursor: 'pointer',
    fontWeight: '500',
  },
  menuButton: {
    backgroundColor: 'transparent',
    border: 'none',
    padding: '0.5rem',
    cursor: 'pointer',
    color: 'var(--text-secondary)',
    fontSize: '1rem',
  },
  menu: {
    marginTop: '0.5rem',
    padding: '0.5rem',
    backgroundColor: 'var(--background)',
    borderRadius: '0.5rem',
  },
  menuItem: {
    width: '100%',
    textAlign: 'left',
    backgroundColor: 'transparent',
    border: 'none',
    padding: '0.5rem',
    cursor: 'pointer',
    color: 'var(--error)',
    fontSize: '0.875rem',
  },
  modalOverlay: {
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
    backgroundColor: 'var(--background)',
    borderRadius: '1rem 1rem 0 0',
    width: '100%',
    maxWidth: '480px',
    maxHeight: '90vh',
    overflow: 'auto',
  },
  modalHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '1rem',
    borderBottom: '1px solid var(--border)',
  },
  modalTitle: {
    margin: 0,
    fontSize: '1.125rem',
    fontWeight: '600',
  },
  closeButton: {
    backgroundColor: 'transparent',
    border: 'none',
    fontSize: '1.5rem',
    cursor: 'pointer',
    color: 'var(--text-secondary)',
    padding: '0.25rem',
  },
  form: {
    padding: '1rem',
  },
  formGroup: {
    marginBottom: '1rem',
    flex: 1,
  },
  formRow: {
    display: 'flex',
    gap: '1rem',
  },
  label: {
    display: 'block',
    fontSize: '0.875rem',
    fontWeight: '500',
    marginBottom: '0.25rem',
    color: 'var(--text-secondary)',
  },
  input: {
    width: '100%',
    padding: '0.75rem',
    border: '1px solid var(--border)',
    borderRadius: '0.5rem',
    fontSize: '1rem',
    backgroundColor: 'var(--surface)',
    boxSizing: 'border-box',
  },
  select: {
    width: '100%',
    padding: '0.75rem',
    border: '1px solid var(--border)',
    borderRadius: '0.5rem',
    fontSize: '1rem',
    backgroundColor: 'var(--surface)',
    boxSizing: 'border-box',
  },
  submitButton: {
    width: '100%',
    padding: '0.875rem',
    backgroundColor: 'var(--primary)',
    color: 'white',
    border: 'none',
    borderRadius: '0.5rem',
    fontSize: '1rem',
    fontWeight: '600',
    cursor: 'pointer',
    marginTop: '0.5rem',
  },
}
