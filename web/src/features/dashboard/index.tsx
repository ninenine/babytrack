import { useFamilyStore } from '@/stores/family.store'
import { useOnline } from '@/hooks/useOnline'

export function Dashboard() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const isOnline = useOnline()

  return (
    <div style={{ padding: '1rem' }}>
      <header style={{ marginBottom: '1.5rem' }}>
        <h1>Family Tracker</h1>
        {!isOnline && (
          <div style={{ color: 'var(--warning)', marginTop: '0.5rem' }}>
            You are offline. Changes will sync when you reconnect.
          </div>
        )}
      </header>

      {currentChild ? (
        <div>
          <h2>Dashboard for {currentChild.name}</h2>

          <div style={{ display: 'grid', gap: '1rem', marginTop: '1rem' }}>
            <DashboardCard title="Last Feeding" href="/feeding">
              <p>No recent feedings</p>
            </DashboardCard>

            <DashboardCard title="Sleep Status" href="/sleep">
              <p>No active sleep session</p>
            </DashboardCard>

            <DashboardCard title="Medications Due" href="/medications">
              <p>No medications due</p>
            </DashboardCard>

            <DashboardCard title="Recent Notes" href="/notes">
              <p>No recent notes</p>
            </DashboardCard>
          </div>
        </div>
      ) : (
        <div>
          <p>No child selected. Please add a child to get started.</p>
        </div>
      )}
    </div>
  )
}

function DashboardCard({
  title,
  href,
  children,
}: {
  title: string
  href: string
  children: React.ReactNode
}) {
  return (
    <div
      style={{
        border: '1px solid var(--border)',
        borderRadius: '0.5rem',
        padding: '1rem',
      }}
    >
      <h3 style={{ marginBottom: '0.5rem' }}>
        <a href={href} style={{ color: 'var(--primary)', textDecoration: 'none' }}>
          {title}
        </a>
      </h3>
      {children}
    </div>
  )
}
