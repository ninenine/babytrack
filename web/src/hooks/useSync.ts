import { useCallback, useState } from 'react'
import { syncPendingEvents, pullFromServer } from '@/db/sync'
import { getPendingEventCount } from '@/db/events'

interface SyncState {
  isSyncing: boolean
  lastSyncTime: Date | null
  pendingCount: number
  error: string | null
}

export function useSync() {
  const [state, setState] = useState<SyncState>({
    isSyncing: false,
    lastSyncTime: null,
    pendingCount: 0,
    error: null,
  })

  const syncPending = useCallback(async () => {
    setState((prev) => ({ ...prev, isSyncing: true, error: null }))

    try {
      const { synced, failed } = await syncPendingEvents()
      const pendingCount = await getPendingEventCount()

      setState((prev) => ({
        ...prev,
        isSyncing: false,
        lastSyncTime: new Date(),
        pendingCount,
        error: failed > 0 ? `${failed} events failed to sync` : null,
      }))

      return { synced, failed }
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isSyncing: false,
        error: error instanceof Error ? error.message : 'Sync failed',
      }))
      throw error
    }
  }, [])

  const pullUpdates = useCallback(async (lastSync?: string) => {
    setState((prev) => ({ ...prev, isSyncing: true, error: null }))

    try {
      await pullFromServer(lastSync)
      setState((prev) => ({
        ...prev,
        isSyncing: false,
        lastSyncTime: new Date(),
      }))
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isSyncing: false,
        error: error instanceof Error ? error.message : 'Pull failed',
      }))
      throw error
    }
  }, [])

  const fullSync = useCallback(async (lastSync?: string) => {
    await syncPending()
    await pullUpdates(lastSync)
  }, [syncPending, pullUpdates])

  const refreshPendingCount = useCallback(async () => {
    const count = await getPendingEventCount()
    setState((prev) => ({ ...prev, pendingCount: count }))
  }, [])

  return {
    ...state,
    syncPendingEvents: syncPending,
    pullUpdates,
    fullSync,
    refreshPendingCount,
  }
}
