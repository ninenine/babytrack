import { ReactNode, useEffect } from 'react'
import { useOnline } from '@/hooks/useOnline'
import { useSync } from '@/hooks/useSync'

interface ProvidersProps {
  children: ReactNode
}

export function Providers({ children }: ProvidersProps) {
  const isOnline = useOnline()
  const { syncPendingEvents } = useSync()

  // Sync when coming back online
  useEffect(() => {
    if (isOnline) {
      syncPendingEvents()
    }
  }, [isOnline, syncPendingEvents])

  return <>{children}</>
}
