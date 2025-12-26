import { Outlet } from 'react-router-dom'
import { TopNavbar } from './top-navbar'
import { BottomNav } from './bottom-nav'
import { useMobile } from '@/hooks'

export function AppShell() {
  const isMobile = useMobile()

  return (
    <div className="min-h-screen bg-background">
      {/* Top Navbar */}
      <TopNavbar />

      {/* Main Content */}
      <main className={`container max-w-4xl mx-auto px-4 py-4 ${isMobile ? 'pb-20' : ''}`}>
        <Outlet />
      </main>

      {/* Bottom Navigation (mobile only) */}
      {isMobile && <BottomNav />}
    </div>
  )
}
