import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { apiClient } from '@/api/client'
import { useFamilyStore } from '@/stores/family.store'

type Step = 'family' | 'child'

interface FamilyResponse {
  id: string
  name: string
  created_at: string
  updated_at: string
}

interface ChildResponse {
  id: string
  family_id: string
  name: string
  date_of_birth: string
  gender?: string
  avatar_url?: string
  created_at: string
  updated_at: string
}

export function OnboardingPage() {
  const navigate = useNavigate()
  const { setFamilies } = useFamilyStore()

  const [step, setStep] = useState<Step>('family')
  const [familyId, setFamilyId] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Family form
  const [familyName, setFamilyName] = useState('')

  // Child form
  const [childName, setChildName] = useState('')
  const [dateOfBirth, setDateOfBirth] = useState('')
  const [gender, setGender] = useState('')

  async function handleCreateFamily(e: React.FormEvent) {
    e.preventDefault()
    if (!familyName.trim()) return

    setIsSubmitting(true)
    setError(null)

    try {
      const response = await apiClient.post<FamilyResponse>('/api/families', {
        name: familyName.trim(),
      })

      setFamilyId(response.data.id)
      setStep('child')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create family')
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleAddChild(e: React.FormEvent) {
    e.preventDefault()
    if (!childName.trim() || !dateOfBirth || !familyId) return

    setIsSubmitting(true)
    setError(null)

    try {
      const childResponse = await apiClient.post<ChildResponse>(
        `/api/families/${familyId}/children`,
        {
          name: childName.trim(),
          date_of_birth: new Date(dateOfBirth).toISOString(),
          gender: gender || undefined,
        }
      )

      // Set the newly created family with child in store
      const newChild = {
        id: childResponse.data.id,
        name: childResponse.data.name,
        dateOfBirth: childResponse.data.date_of_birth,
        gender: childResponse.data.gender,
        avatarUrl: childResponse.data.avatar_url,
      }

      // Fetch the family we just created
      const familyResponse = await apiClient.get<FamilyResponse>(`/api/families/${familyId}`)

      setFamilies([
        {
          id: familyResponse.data.id,
          name: familyResponse.data.name,
          children: [newChild],
        },
      ])
      navigate('/', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add child')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <div style={styles.stepIndicator}>
          <div style={{ ...styles.stepDot, ...(step === 'family' ? styles.stepDotActive : styles.stepDotComplete) }}>
            1
          </div>
          <div style={styles.stepLine} />
          <div style={{ ...styles.stepDot, ...(step === 'child' ? styles.stepDotActive : {}) }}>2</div>
        </div>

        {step === 'family' && (
          <>
            <h1 style={styles.title}>Create Your Family</h1>
            <p style={styles.subtitle}>
              Let's start by giving your family a name. This helps organize everything in one place.
            </p>

            <form onSubmit={handleCreateFamily}>
              <div style={styles.formGroup}>
                <label htmlFor="familyName" style={styles.label}>
                  Family Name
                </label>
                <input
                  id="familyName"
                  type="text"
                  value={familyName}
                  onChange={(e) => setFamilyName(e.target.value)}
                  placeholder="e.g., The Smiths"
                  style={styles.input}
                  required
                  autoFocus
                />
              </div>

              {error && <p style={styles.error}>{error}</p>}

              <button type="submit" style={styles.button} disabled={isSubmitting || !familyName.trim()}>
                {isSubmitting ? 'Creating...' : 'Continue'}
              </button>
            </form>
          </>
        )}

        {step === 'child' && (
          <>
            <h1 style={styles.title}>Add Your First Child</h1>
            <p style={styles.subtitle}>Add your child's information to start tracking their activities.</p>

            <form onSubmit={handleAddChild}>
              <div style={styles.formGroup}>
                <label htmlFor="childName" style={styles.label}>
                  Child's Name
                </label>
                <input
                  id="childName"
                  type="text"
                  value={childName}
                  onChange={(e) => setChildName(e.target.value)}
                  placeholder="e.g., Emma"
                  style={styles.input}
                  required
                  autoFocus
                />
              </div>

              <div style={styles.formGroup}>
                <label htmlFor="dateOfBirth" style={styles.label}>
                  Date of Birth
                </label>
                <input
                  id="dateOfBirth"
                  type="date"
                  value={dateOfBirth}
                  onChange={(e) => setDateOfBirth(e.target.value)}
                  style={styles.input}
                  required
                  max={new Date().toISOString().split('T')[0]}
                />
              </div>

              <div style={styles.formGroup}>
                <label htmlFor="gender" style={styles.label}>
                  Gender (optional)
                </label>
                <select
                  id="gender"
                  value={gender}
                  onChange={(e) => setGender(e.target.value)}
                  style={styles.input}
                >
                  <option value="">Prefer not to say</option>
                  <option value="male">Male</option>
                  <option value="female">Female</option>
                  <option value="other">Other</option>
                </select>
              </div>

              {error && <p style={styles.error}>{error}</p>}

              <button
                type="submit"
                style={styles.button}
                disabled={isSubmitting || !childName.trim() || !dateOfBirth}
              >
                {isSubmitting ? 'Saving...' : 'Get Started'}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    minHeight: '100vh',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: 'var(--surface)',
    padding: '1rem',
  },
  card: {
    backgroundColor: 'var(--background)',
    borderRadius: '1rem',
    padding: '2.5rem',
    maxWidth: '400px',
    width: '100%',
    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
  },
  stepIndicator: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: '2rem',
  },
  stepDot: {
    width: '32px',
    height: '32px',
    borderRadius: '50%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '0.875rem',
    fontWeight: '600',
    backgroundColor: 'var(--surface)',
    color: 'var(--text-secondary)',
    border: '2px solid var(--border)',
  },
  stepDotActive: {
    backgroundColor: 'var(--primary)',
    color: 'white',
    borderColor: 'var(--primary)',
  },
  stepDotComplete: {
    backgroundColor: 'var(--success)',
    color: 'white',
    borderColor: 'var(--success)',
  },
  stepLine: {
    width: '60px',
    height: '2px',
    backgroundColor: 'var(--border)',
    margin: '0 0.5rem',
  },
  title: {
    fontSize: '1.5rem',
    fontWeight: 'bold',
    marginBottom: '0.5rem',
    color: 'var(--text)',
    textAlign: 'center',
  },
  subtitle: {
    color: 'var(--text-secondary)',
    marginBottom: '2rem',
    lineHeight: '1.5',
    textAlign: 'center',
  },
  formGroup: {
    marginBottom: '1.25rem',
  },
  label: {
    display: 'block',
    marginBottom: '0.5rem',
    fontWeight: '500',
    color: 'var(--text)',
    fontSize: '0.875rem',
  },
  input: {
    width: '100%',
    padding: '0.75rem 1rem',
    borderRadius: '0.5rem',
    border: '1px solid var(--border)',
    fontSize: '1rem',
    backgroundColor: 'var(--background)',
    color: 'var(--text)',
    boxSizing: 'border-box',
  },
  button: {
    width: '100%',
    padding: '0.875rem 1.5rem',
    backgroundColor: 'var(--primary)',
    color: 'white',
    border: 'none',
    borderRadius: '0.5rem',
    fontSize: '1rem',
    fontWeight: '600',
    cursor: 'pointer',
    marginTop: '0.5rem',
  },
  error: {
    color: 'var(--error)',
    fontSize: '0.875rem',
    marginBottom: '1rem',
    textAlign: 'center',
  },
}
