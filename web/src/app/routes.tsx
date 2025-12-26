import { Routes, Route, Navigate } from 'react-router-dom'
import { useSessionStore } from '@/stores/session.store'
import { useFamilyStore } from '@/stores/family.store'

// Feature components
import { LoginPage } from '@/features/auth'
import { OnboardingPage } from '@/features/onboarding'
import { Dashboard } from '@/features/dashboard'
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

      <Route
        path="/"
        element={
          <RequireFamilyRoute>
            <Dashboard />
          </RequireFamilyRoute>
        }
      />

      <Route
        path="/feeding"
        element={
          <RequireFamilyRoute>
            <FeedingList />
          </RequireFamilyRoute>
        }
      />

      <Route
        path="/sleep"
        element={
          <RequireFamilyRoute>
            <SleepList />
          </RequireFamilyRoute>
        }
      />

      <Route
        path="/medications"
        element={
          <RequireFamilyRoute>
            <MedicationList />
          </RequireFamilyRoute>
        }
      />

      <Route
        path="/notes"
        element={
          <RequireFamilyRoute>
            <NotesList />
          </RequireFamilyRoute>
        }
      />

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
