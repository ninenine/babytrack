import { useState } from 'react'
import { Plus, Calendar, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { VaccinationForm, VaccinationCard } from '@/components/vaccination'
import { useVaccinations, useGenerateVaccinationSchedule } from '@/hooks'
import { useFamilyStore } from '@/stores/family.store'
import type { LocalVaccination } from '@/db/dexie'

export function VaccinationPage() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const { upcoming, overdue, completed, vaccinations, isLoading, isSyncing } = useVaccinations()
  const generateSchedule = useGenerateVaccinationSchedule()
  const [showForm, setShowForm] = useState(false)
  const [editingVaccination, setEditingVaccination] = useState<LocalVaccination | null>(null)

  const handleEdit = (vaccination: LocalVaccination) => {
    setEditingVaccination(vaccination)
    setShowForm(true)
  }

  if (!currentChild) {
    return (
      <div className="flex items-center justify-center h-[50vh] text-muted-foreground">
        No child selected
      </div>
    )
  }

  const handleGenerateSchedule = async () => {
    await generateSchedule.mutateAsync()
  }

  const pendingVaccinations = [...overdue, ...upcoming]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h1 className="text-2xl font-bold">Vaccinations</h1>
          {isSyncing && (
            <RefreshCw className="h-4 w-4 animate-spin text-muted-foreground" />
          )}
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleGenerateSchedule}
            disabled={generateSchedule.isPending}
          >
            <Calendar className="h-4 w-4 mr-1" />
            {generateSchedule.isPending ? 'Generating...' : 'Generate'}
          </Button>
          <Button size="sm" onClick={() => setShowForm(true)}>
            <Plus className="h-4 w-4 mr-1" />
            Add
          </Button>
        </div>
      </div>

      <Tabs defaultValue="upcoming" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="upcoming">
            Upcoming {pendingVaccinations.length > 0 && `(${pendingVaccinations.length})`}
          </TabsTrigger>
          <TabsTrigger value="completed">
            Completed {completed.length > 0 && `(${completed.length})`}
          </TabsTrigger>
          <TabsTrigger value="all">All</TabsTrigger>
        </TabsList>

        <TabsContent value="upcoming" className="mt-4">
          {isLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
            </div>
          ) : pendingVaccinations.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No upcoming vaccinations</p>
              <p className="text-sm mt-1">Generate a schedule to get started</p>
            </div>
          ) : (
            <div className="space-y-3">
              {pendingVaccinations.map((vax) => (
                <VaccinationCard key={vax.id} vaccination={vax} onEdit={handleEdit} />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="completed" className="mt-4">
          {isLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-20 w-full" />
            </div>
          ) : completed.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No completed vaccinations</p>
            </div>
          ) : (
            <div className="space-y-3">
              {completed.map((vax) => (
                <VaccinationCard key={vax.id} vaccination={vax} onEdit={handleEdit} />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="all" className="mt-4">
          {isLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
            </div>
          ) : vaccinations.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No vaccinations</p>
            </div>
          ) : (
            <div className="space-y-3">
              {vaccinations.map((vax) => (
                <VaccinationCard key={vax.id} vaccination={vax} onEdit={handleEdit} />
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>

      <VaccinationForm
        open={showForm}
        onOpenChange={(open) => {
          setShowForm(open)
          if (!open) setEditingVaccination(null)
        }}
        vaccination={editingVaccination}
      />
    </div>
  )
}
