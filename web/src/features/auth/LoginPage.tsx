import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useSessionStore } from '@/stores/session.store'
import { useFamilyStore } from '@/stores/family.store'
import { apiClient } from '@/api/client'

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
}

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setSession, isAuthenticated } = useSessionStore()
  const { setFamilies } = useFamilyStore()

  // Handle OAuth callback
  useEffect(() => {
    const token = searchParams.get('token')
    const error = searchParams.get('error')

    if (error) {
      console.error('OAuth error:', error)
      return
    }

    if (token) {
      // Fetch user info and set session
      handleTokenCallback(token)
    }
  }, [searchParams])

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/')
    }
  }, [isAuthenticated, navigate])

  async function handleTokenCallback(token: string) {
    try {
      const response = await apiClient.get<{
        id: string
        email: string
        name: string
        avatar_url?: string
      }>('/api/auth/me', {
        params: { token },
      })

      setSession(
        {
          id: response.data.id,
          email: response.data.email,
          name: response.data.name,
          avatarUrl: response.data.avatar_url,
        },
        token
      )

      // Fetch user's families
      try {
        const familiesResponse = await apiClient.get<FamilyResponse[]>('/api/families')
        const families = await Promise.all(
          familiesResponse.data.map(async (f) => {
            // Fetch children for each family
            const childrenResponse = await apiClient.get<ChildResponse[]>(
              `/api/families/${f.id}/children`
            )
            return {
              id: f.id,
              name: f.name,
              children: childrenResponse.data.map((c) => ({
                id: c.id,
                name: c.name,
                dateOfBirth: c.date_of_birth,
                gender: c.gender,
                avatarUrl: c.avatar_url,
              })),
            }
          })
        )
        setFamilies(families)
      } catch {
        // No families yet, that's ok
        setFamilies([])
      }

      // Clear URL params and redirect
      navigate('/', { replace: true })
    } catch (error) {
      console.error('Failed to fetch user info:', error)
    }
  }

  function handleGoogleLogin() {
    // Redirect to backend OAuth endpoint
    window.location.href = '/api/auth/google'
  }

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <h1 style={styles.title}>Family Tracker</h1>
        <p style={styles.subtitle}>
          Track feeding, sleep, medications, and more for your little ones.
        </p>

        <button onClick={handleGoogleLogin} style={styles.googleButton}>
          <GoogleIcon />
          <span>Continue with Google</span>
        </button>

        <p style={styles.disclaimer}>
          By continuing, you agree to our Terms of Service and Privacy Policy.
        </p>
      </div>
    </div>
  )
}

function GoogleIcon() {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 18 18"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      style={{ marginRight: '8px' }}
    >
      <path
        d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844c-.209 1.125-.843 2.078-1.796 2.717v2.258h2.908c1.702-1.567 2.684-3.874 2.684-6.615z"
        fill="#4285F4"
      />
      <path
        d="M9.003 18c2.43 0 4.467-.806 5.956-2.18l-2.909-2.26c-.806.54-1.836.86-3.047.86-2.344 0-4.328-1.584-5.036-3.711H.96v2.332C2.44 15.983 5.485 18 9.003 18z"
        fill="#34A853"
      />
      <path
        d="M3.964 10.712c-.18-.54-.282-1.117-.282-1.71 0-.593.102-1.17.282-1.71V4.96H.957C.347 6.175 0 7.55 0 9.002c0 1.452.348 2.827.957 4.042l3.007-2.332z"
        fill="#FBBC05"
      />
      <path
        d="M9.003 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.464.891 11.428 0 9.002 0 5.485 0 2.44 2.017.96 4.958L3.967 7.29c.708-2.127 2.692-3.71 5.036-3.71z"
        fill="#EA4335"
      />
    </svg>
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
    textAlign: 'center',
    boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
  },
  title: {
    fontSize: '1.875rem',
    fontWeight: 'bold',
    marginBottom: '0.5rem',
    color: 'var(--text)',
  },
  subtitle: {
    color: 'var(--text-secondary)',
    marginBottom: '2rem',
    lineHeight: '1.5',
  },
  googleButton: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: '100%',
    padding: '0.75rem 1.5rem',
    backgroundColor: 'var(--background)',
    border: '1px solid var(--border)',
    borderRadius: '0.5rem',
    fontSize: '1rem',
    fontWeight: '500',
    cursor: 'pointer',
    transition: 'background-color 0.2s',
  },
  disclaimer: {
    marginTop: '1.5rem',
    fontSize: '0.75rem',
    color: 'var(--text-secondary)',
  },
}
