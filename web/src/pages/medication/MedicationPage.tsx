import { useState } from 'react'
import { Plus, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { MedicationForm, MedicationCard } from '@/components/medication'
import { useMedications } from '@/hooks'
import { useFamilyStore } from '@/stores/family.store'

export function MedicationPage() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const { activeMedications, inactiveMedications, isLoading, isSyncing } = useMedications()
  const [showForm, setShowForm] = useState(false)

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
          <h1 className="text-2xl font-bold">Medications</h1>
          {isSyncing && (
            <RefreshCw className="h-4 w-4 animate-spin text-muted-foreground" />
          )}
        </div>
        <Button size="sm" onClick={() => setShowForm(true)}>
          <Plus className="h-4 w-4 mr-1" />
          Add
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
      ) : activeMedications.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <p className="text-lg">No active medications</p>
          <p className="text-sm mt-1">Tap + Add to add a medication</p>
        </div>
      ) : (
        <div className="space-y-3">
          {activeMedications.map((medication) => (
            <MedicationCard key={medication.id} medication={medication} />
          ))}
        </div>
      )}

      {inactiveMedications.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold mb-3 text-muted-foreground">Past Medications</h2>
          <div className="space-y-3 opacity-60">
            {inactiveMedications.map((medication) => (
              <MedicationCard key={medication.id} medication={medication} />
            ))}
          </div>
        </div>
      )}

      <MedicationForm open={showForm} onOpenChange={setShowForm} />
    </div>
  )
}
