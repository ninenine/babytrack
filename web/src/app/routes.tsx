import { Routes, Route, Navigate } from 'react-router-dom'
import { useSessionStore } from '@/stores/session.store'
import { useFamilyStore } from '@/stores/family.store'

// Components
import { Layout } from '@/components/Layout'
import { LoginPage } from '@/features/auth'
import { OnboardingPage } from '@/features/onboarding'
import { Timeline } from '@/features/timeline'
import { FeedingList } from '@/features/feeding'
import { SleepList } from '@/features/sleep'
import { MedicationList } from '@/features/medication'
import { NotesList } from '@/features/notes'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useSessionStore((state) => state.isAuthenticated)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function RequireFamilyRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useSessionStore((state) => state.isAuthenticated)
  const currentFamily = useFamilyStore((state) => state.currentFamily)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  if (!currentFamily) {
    return <Navigate to="/onboarding" replace />
  }

  return <>{children}</>
}

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />

      <Route
        path="/onboarding"
        element={
          <ProtectedRoute>
            <OnboardingPage />
          </ProtectedRoute>
        }
      />

      {/* App routes with bottom navigation */}
      <Route
        element={
          <RequireFamilyRoute>
            <Layout />
          </RequireFamilyRoute>
        }
      >
        <Route index element={<Timeline />} />
        <Route path="feeding" element={<FeedingList />} />
        <Route path="sleep" element={<SleepList />} />
        <Route path="history" element={<HistoryPlaceholder />} />
        <Route path="health" element={<HealthPlaceholder />} />
        <Route path="medications" element={<MedicationList />} />
        <Route path="notes" element={<NotesList />} />
        <Route path="settings" element={<SettingsPlaceholder />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

// Placeholder components
function HistoryPlaceholder() {
  return (
    <div style={{ padding: '1rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
      <h2>History</h2>
      <p>Coming soon...</p>
    </div>
  )
}

function HealthPlaceholder() {
  return (
    <div style={{ padding: '1rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
      <h2>Health</h2>
      <p>Medications, vaccinations, and growth tracking coming soon...</p>
    </div>
  )
}

function SettingsPlaceholder() {
  return (
    <div style={{ padding: '1rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
      <h2>Settings</h2>
      <p>Account and preferences coming soon...</p>
    </div>
  )
}
