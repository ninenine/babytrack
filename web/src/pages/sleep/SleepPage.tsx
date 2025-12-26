import { useState } from 'react'
import { RefreshCw } from 'lucide-react'
import { ActiveSleepCard, SleepList, StartSleepButtons, SleepFormDialog } from '@/components/sleep'
import { useSleep } from '@/hooks'
import { useFamilyStore } from '@/stores/family.store'
import type { LocalSleep } from '@/db/dexie'

export function SleepPage() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const { sleepRecords, activeSleep, isLoading, isSyncing } = useSleep()
  const [editingSleep, setEditingSleep] = useState<LocalSleep | null>(null)

  if (!currentChild) {
    return (
      <div className="flex items-center justify-center h-[50vh] text-muted-foreground">
        No child selected
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h1 className="text-2xl font-bold">Sleep</h1>
          {isSyncing && (
            <RefreshCw className="h-4 w-4 animate-spin text-muted-foreground" />
          )}
        </div>
      </div>

      {activeSleep ? (
        <ActiveSleepCard sleep={activeSleep} />
      ) : (
        <StartSleepButtons />
      )}

      <div>
        <h2 className="text-lg font-semibold mb-3">History</h2>
        <SleepList
          sleepRecords={sleepRecords}
          isLoading={isLoading}
          onEdit={setEditingSleep}
        />
      </div>

      <SleepFormDialog
        open={!!editingSleep}
        onOpenChange={(open) => !open && setEditingSleep(null)}
        sleep={editingSleep}
      />
    </div>
  )
}
