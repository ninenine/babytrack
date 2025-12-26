import { format } from 'date-fns'
import { Calendar, Clock, MapPin, User, Trash2, Check, X, Pencil } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { useUpdateAppointment, useDeleteAppointment, type AppointmentType } from '@/hooks'
import type { LocalAppointment } from '@/db/dexie'

interface AppointmentCardProps {
  appointment: LocalAppointment
  onEdit?: (appointment: LocalAppointment) => void
}

const typeConfig: Record<AppointmentType, { label: string; color: string }> = {
  well_visit: { label: 'Well Visit', color: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' },
  sick_visit: { label: 'Sick Visit', color: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' },
  specialist: { label: 'Specialist', color: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200' },
  dental: { label: 'Dental', color: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' },
  other: { label: 'Other', color: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200' },
}

export function AppointmentCard({ appointment, onEdit }: AppointmentCardProps) {
  const updateAppointment = useUpdateAppointment()
  const deleteAppointment = useDeleteAppointment()

  const config = typeConfig[appointment.type]
  const isPast = new Date(appointment.scheduledAt) < new Date()

  const handleComplete = async () => {
    await updateAppointment.mutateAsync({ id: appointment.id, completed: true })
  }

  const handleCancel = async () => {
    await updateAppointment.mutateAsync({ id: appointment.id, cancelled: true })
  }

  const handleDelete = () => {
    deleteAppointment.mutate(appointment.id)
  }

  return (
    <Card className={appointment.cancelled ? 'opacity-50' : ''}>
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="font-medium">{appointment.title}</span>
              <Badge className={config.color}>{config.label}</Badge>
              {appointment.completed && (
                <Badge variant="secondary" className="bg-green-100 text-green-800">
                  Completed
                </Badge>
              )}
              {appointment.cancelled && (
                <Badge variant="secondary" className="bg-gray-100 text-gray-800">
                  Cancelled
                </Badge>
              )}
              {appointment.pendingSync && (
                <Badge variant="outline" className="text-xs text-yellow-600">
                  Pending
                </Badge>
              )}
            </div>

            <div className="mt-2 space-y-1 text-sm text-muted-foreground">
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                <span>{format(new Date(appointment.scheduledAt), 'EEEE, MMMM d, yyyy')}</span>
              </div>
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4" />
                <span>
                  {format(new Date(appointment.scheduledAt), 'h:mm a')} ({appointment.duration} min)
                </span>
              </div>
              {appointment.provider && (
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4" />
                  <span>{appointment.provider}</span>
                </div>
              )}
              {appointment.location && (
                <div className="flex items-center gap-2">
                  <MapPin className="h-4 w-4" />
                  <span>{appointment.location}</span>
                </div>
              )}
            </div>

            {appointment.notes && (
              <p className="mt-2 text-sm text-muted-foreground italic">{appointment.notes}</p>
            )}
          </div>

          <div className="flex flex-col gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onEdit?.(appointment)}
            >
              <Pencil className="h-4 w-4 text-muted-foreground" />
            </Button>

            {!appointment.completed && !appointment.cancelled && (
              <>
                {isPast && (
                  <Button
                    size="sm"
                    variant="outline"
                    className="gap-1"
                    onClick={handleComplete}
                    disabled={updateAppointment.isPending}
                  >
                    <Check className="h-3 w-3" />
                    Done
                  </Button>
                )}
                <Button
                  size="sm"
                  variant="ghost"
                  className="gap-1 text-muted-foreground"
                  onClick={handleCancel}
                  disabled={updateAppointment.isPending}
                >
                  <X className="h-3 w-3" />
                  Cancel
                </Button>
              </>
            )}

            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8">
                  <Trash2 className="h-4 w-4 text-muted-foreground" />
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete Appointment</AlertDialogTitle>
                  <AlertDialogDescription>
                    Are you sure you want to delete this appointment? This action cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                    Delete
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
