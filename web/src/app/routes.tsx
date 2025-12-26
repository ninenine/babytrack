import { Routes, Route, Navigate } from 'react-router-dom'
import { useSessionStore } from '@/stores/session.store'

// Feature components
import { LoginPage } from '@/features/auth'
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

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />

      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Dashboard />
          </ProtectedRoute>
        }
      />

      <Route
        path="/feeding"
        element={
          <ProtectedRoute>
            <FeedingList />
          </ProtectedRoute>
        }
      />

      <Route
        path="/sleep"
        element={
          <ProtectedRoute>
            <SleepList />
          </ProtectedRoute>
        }
      />

      <Route
        path="/medications"
        element={
          <ProtectedRoute>
            <MedicationList />
          </ProtectedRoute>
        }
      />

      <Route
        path="/notes"
        element={
          <ProtectedRoute>
            <NotesList />
          </ProtectedRoute>
        }
      />

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
