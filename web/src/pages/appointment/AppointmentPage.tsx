import { useState } from 'react'
import { Plus, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { AppointmentForm, AppointmentCard } from '@/components/appointment'
import { useAppointments } from '@/hooks'
import { useFamilyStore } from '@/stores/family.store'
import type { LocalAppointment } from '@/db/dexie'

export function AppointmentPage() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const { upcoming, past, appointments, isLoading, isSyncing } = useAppointments()
  const [showForm, setShowForm] = useState(false)
  const [editingAppointment, setEditingAppointment] = useState<LocalAppointment | null>(null)

  const handleEdit = (appointment: LocalAppointment) => {
    setEditingAppointment(appointment)
    setShowForm(true)
  }

  if (!currentChild) {
    return (
      <div className="flex items-center justify-center h-[50vh] text-muted-foreground">
        No child selected
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h1 className="text-2xl font-bold">Appointments</h1>
          {isSyncing && (
            <RefreshCw className="h-4 w-4 animate-spin text-muted-foreground" />
          )}
        </div>
        <Button size="sm" onClick={() => setShowForm(true)}>
          <Plus className="h-4 w-4 mr-1" />
          Add
        </Button>
      </div>

      <Tabs defaultValue="upcoming" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="upcoming">
            Upcoming {upcoming.length > 0 && `(${upcoming.length})`}
          </TabsTrigger>
          <TabsTrigger value="past">
            Past {past.length > 0 && `(${past.length})`}
          </TabsTrigger>
          <TabsTrigger value="all">All</TabsTrigger>
        </TabsList>

        <TabsContent value="upcoming" className="mt-4">
          {isLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-32 w-full" />
              <Skeleton className="h-32 w-full" />
            </div>
          ) : upcoming.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No upcoming appointments</p>
              <p className="text-sm mt-1">Tap + Add to schedule an appointment</p>
            </div>
          ) : (
            <div className="space-y-3">
              {upcoming.map((apt) => (
                <AppointmentCard key={apt.id} appointment={apt} onEdit={handleEdit} />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="past" className="mt-4">
          {isLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-32 w-full" />
            </div>
          ) : past.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No past appointments</p>
            </div>
          ) : (
            <div className="space-y-3">
              {past.map((apt) => (
                <AppointmentCard key={apt.id} appointment={apt} onEdit={handleEdit} />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="all" className="mt-4">
          {isLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-32 w-full" />
              <Skeleton className="h-32 w-full" />
            </div>
          ) : appointments.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No appointments</p>
            </div>
          ) : (
            <div className="space-y-3">
              {appointments.map((apt) => (
                <AppointmentCard key={apt.id} appointment={apt} onEdit={handleEdit} />
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>

      <AppointmentForm
        open={showForm}
        onOpenChange={(open) => {
          setShowForm(open)
          if (!open) setEditingAppointment(null)
        }}
        appointment={editingAppointment}
      />
    </div>
  )
}
