import { NavLink, Outlet } from 'react-router-dom'
import { useFamilyStore } from '@/stores/family.store'

export function Layout() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  return (
    <div style={styles.container}>
      {/* Header */}
      <header style={styles.header}>
        <div style={styles.headerContent}>
          <h1 style={styles.logo}>BabyTrack</h1>
          {currentChild && (
            <div style={styles.childBadge}>
              {currentChild.name}
            </div>
          )}
        </div>
      </header>

      {/* Main content */}
      <main style={styles.main}>
        <Outlet />
      </main>

      {/* Bottom Navigation */}
      <nav style={styles.bottomNav}>
        <NavLink to="/" style={({ isActive }) => ({ ...styles.navItem, ...(isActive ? styles.navItemActive : {}) })} end>
          <span style={styles.navIcon}>📅</span>
          <span style={styles.navLabel}>Today</span>
        </NavLink>
        <NavLink to="/history" style={({ isActive }) => ({ ...styles.navItem, ...(isActive ? styles.navItemActive : {}) })}>
          <span style={styles.navIcon}>📊</span>
          <span style={styles.navLabel}>History</span>
        </NavLink>
        <NavLink to="/health" style={({ isActive }) => ({ ...styles.navItem, ...(isActive ? styles.navItemActive : {}) })}>
          <span style={styles.navIcon}>💊</span>
          <span style={styles.navLabel}>Health</span>
        </NavLink>
        <NavLink to="/settings" style={({ isActive }) => ({ ...styles.navItem, ...(isActive ? styles.navItemActive : {}) })}>
          <span style={styles.navIcon}>⚙️</span>
          <span style={styles.navLabel}>Settings</span>
        </NavLink>
      </nav>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    minHeight: '100vh',
    maxWidth: '480px',
    margin: '0 auto',
    backgroundColor: 'var(--background)',
  },
  header: {
    position: 'sticky',
    top: 0,
    backgroundColor: 'var(--background)',
    borderBottom: '1px solid var(--border)',
    zIndex: 100,
  },
  headerContent: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '0.75rem 1rem',
  },
  logo: {
    fontSize: '1.25rem',
    fontWeight: '700',
    color: 'var(--primary)',
    margin: 0,
  },
  childBadge: {
    fontSize: '0.875rem',
    fontWeight: '500',
    color: 'var(--text-secondary)',
    backgroundColor: 'var(--surface)',
    padding: '0.25rem 0.75rem',
    borderRadius: '1rem',
  },
  main: {
    flex: 1,
    paddingBottom: '5rem', // Space for bottom nav + quick add
    overflowY: 'auto',
  },
  bottomNav: {
    position: 'fixed',
    bottom: 0,
    left: '50%',
    transform: 'translateX(-50%)',
    width: '100%',
    maxWidth: '480px',
    display: 'flex',
    justifyContent: 'space-around',
    backgroundColor: 'var(--background)',
    borderTop: '1px solid var(--border)',
    padding: '0.5rem 0',
    paddingBottom: 'calc(0.5rem + env(safe-area-inset-bottom))',
  },
  navItem: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    padding: '0.5rem 1rem',
    textDecoration: 'none',
    color: 'var(--text-secondary)',
    borderRadius: '0.5rem',
    transition: 'color 0.2s',
  },
  navItemActive: {
    color: 'var(--primary)',
  },
  navIcon: {
    fontSize: '1.25rem',
    marginBottom: '0.125rem',
  },
  navLabel: {
    fontSize: '0.75rem',
    fontWeight: '500',
  },
}
