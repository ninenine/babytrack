import { lazy, Suspense } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useSessionStore } from '@/stores/session.store'
import { useFamilyStore } from '@/stores/family.store'
import {
  PageLoader,
  HomePageLoader,
  SettingsPageLoader,
  AuthPageLoader,
} from '@/components/shared/PageLoader'

// Layout - loaded eagerly as it's needed for all routes
import { AppShell } from '@/components/layout/app-shell'

// Lazy loaded pages
const LoginPage = lazy(() =>
  import('@/pages/auth/LoginPage').then((m) => ({ default: m.LoginPage }))
)
const OnboardingPage = lazy(() =>
  import('@/pages/onboarding/OnboardingPage').then((m) => ({ default: m.OnboardingPage }))
)
const HomePage = lazy(() =>
  import('@/pages/home/HomePage').then((m) => ({ default: m.HomePage }))
)
const FeedingPage = lazy(() =>
  import('@/pages/feeding/FeedingPage').then((m) => ({ default: m.FeedingPage }))
)
const SleepPage = lazy(() =>
  import('@/pages/sleep/SleepPage').then((m) => ({ default: m.SleepPage }))
)
const MedicationPage = lazy(() =>
  import('@/pages/medication/MedicationPage').then((m) => ({ default: m.MedicationPage }))
)
const VaccinationPage = lazy(() =>
  import('@/pages/vaccination/VaccinationPage').then((m) => ({ default: m.VaccinationPage }))
)
const AppointmentPage = lazy(() =>
  import('@/pages/appointment/AppointmentPage').then((m) => ({ default: m.AppointmentPage }))
)
const NotesPage = lazy(() =>
  import('@/pages/notes/NotesPage').then((m) => ({ default: m.NotesPage }))
)
const SettingsPage = lazy(() =>
  import('@/pages/settings/SettingsPage').then((m) => ({ default: m.SettingsPage }))
)
const InvitePage = lazy(() =>
  import('@/pages/invite/InvitePage').then((m) => ({ default: m.InvitePage }))
)

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
      {/* Public routes */}
      <Route
        path="/login"
        element={
          <Suspense fallback={<AuthPageLoader />}>
            <LoginPage />
          </Suspense>
        }
      />

      {/* Protected route without family */}
      <Route
        path="/onboarding"
        element={
          <ProtectedRoute>
            <Suspense fallback={<AuthPageLoader />}>
              <OnboardingPage />
            </Suspense>
          </ProtectedRoute>
        }
      />

      {/* Invite route - protected but doesn't require family */}
      <Route
        path="/invite/:familyId"
        element={
          <Suspense fallback={<AuthPageLoader />}>
            <InvitePage />
          </Suspense>
        }
      />

      {/* App routes with layout */}
      <Route
        element={
          <RequireFamilyRoute>
            <AppShell />
          </RequireFamilyRoute>
        }
      >
        <Route
          index
          element={
            <Suspense fallback={<HomePageLoader />}>
              <HomePage />
            </Suspense>
          }
        />
        <Route
          path="feeding"
          element={
            <Suspense fallback={<PageLoader />}>
              <FeedingPage />
            </Suspense>
          }
        />
        <Route
          path="sleep"
          element={
            <Suspense fallback={<PageLoader />}>
              <SleepPage />
            </Suspense>
          }
        />
        <Route
          path="medications"
          element={
            <Suspense fallback={<PageLoader />}>
              <MedicationPage />
            </Suspense>
          }
        />
        <Route
          path="vaccinations"
          element={
            <Suspense fallback={<PageLoader />}>
              <VaccinationPage />
            </Suspense>
          }
        />
        <Route
          path="appointments"
          element={
            <Suspense fallback={<PageLoader />}>
              <AppointmentPage />
            </Suspense>
          }
        />
        <Route
          path="notes"
          element={
            <Suspense fallback={<PageLoader />}>
              <NotesPage />
            </Suspense>
          }
        />
        <Route
          path="settings"
          element={
            <Suspense fallback={<SettingsPageLoader />}>
              <SettingsPage />
            </Suspense>
          }
        />
      </Route>

      {/* Fallback */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
