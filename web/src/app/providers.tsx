import { useEffect } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { Toaster } from '@/components/ui/sonner'
import { queryClient } from '@/lib/query-client'
import { useOnline, useSync } from '@/hooks'

function SyncProvider({ children }: { children: React.ReactNode }) {
  const isOnline = useOnline()
  const { syncPendingEvents, refreshPendingCount } = useSync()

  // Sync when coming back online
  useEffect(() => {
    if (isOnline) {
      syncPendingEvents().catch(console.error)
    }
  }, [isOnline, syncPendingEvents])

  // Refresh pending count on mount
  useEffect(() => {
    refreshPendingCount()
  }, [refreshPendingCount])

  return <>{children}</>
}

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <SyncProvider>
          {children}
          <Toaster position="top-center" richColors closeButton />
        </SyncProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
