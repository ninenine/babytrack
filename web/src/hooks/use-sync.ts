import { useCallback, useState } from 'react'
import { toast } from 'sonner'
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

  const syncPending = useCallback(async (showToast = false) => {
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

      if (showToast) {
        if (synced > 0 && failed === 0) {
          toast.success(`Synced ${synced} ${synced === 1 ? 'item' : 'items'}`)
        } else if (failed > 0) {
          toast.warning(`Synced ${synced}, ${failed} failed`)
        }
      }

      return { synced, failed }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Sync failed'
      setState((prev) => ({
        ...prev,
        isSyncing: false,
        error: message,
      }))
      if (showToast) {
        toast.error(`Sync failed: ${message}`)
      }
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
