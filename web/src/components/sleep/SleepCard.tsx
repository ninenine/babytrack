import { format } from 'date-fns'
import { Moon, Sun, Trash2, Pencil } from 'lucide-react'
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
import { useDeleteSleep } from '@/hooks'
import type { LocalSleep } from '@/db/dexie'

interface SleepCardProps {
  sleep: LocalSleep
  onEdit?: (sleep: LocalSleep) => void
}

function formatDuration(start: Date, end: Date): string {
  const diff = end.getTime() - start.getTime()
  const hours = Math.floor(diff / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)

  if (hours === 0) return `${minutes}m`
  if (minutes === 0) return `${hours}h`
  return `${hours}h ${minutes}m`
}

export function SleepCard({ sleep, onEdit }: SleepCardProps) {
  const deleteSleep = useDeleteSleep()
  const isNight = sleep.type === 'night'

  const handleDelete = () => {
    deleteSleep.mutate(sleep.id)
  }

  return (
    <Card className="relative group">
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-3">
          <div
            className="flex items-center gap-3 flex-1 cursor-pointer"
            onClick={() => onEdit?.(sleep)}
          >
            <div className={`p-2 rounded-full ${isNight ? 'bg-indigo-100 dark:bg-indigo-900' : 'bg-amber-100 dark:bg-amber-900'}`}>
              {isNight ? (
                <Moon className="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
              ) : (
                <Sun className="h-5 w-5 text-amber-600 dark:text-amber-400" />
              )}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-medium capitalize">{sleep.type}</span>
                {sleep.quality && (
                  <Badge variant="secondary" className="text-xs">
                    Quality: {sleep.quality}/5
                  </Badge>
                )}
                {sleep.pendingSync && (
                  <Badge variant="outline" className="text-xs text-yellow-600">
                    Pending
                  </Badge>
                )}
              </div>
              <div className="text-sm text-muted-foreground">
                {format(new Date(sleep.startTime), 'h:mm a')}
                {sleep.endTime && ` - ${format(new Date(sleep.endTime), 'h:mm a')}`}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {sleep.endTime && (
              <div className="text-right">
                <div className="font-medium">
                  {formatDuration(new Date(sleep.startTime), new Date(sleep.endTime))}
                </div>
              </div>
            )}

            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
              onClick={() => onEdit?.(sleep)}
            >
              <Pencil className="h-4 w-4 text-muted-foreground" />
            </Button>

            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <Trash2 className="h-4 w-4 text-muted-foreground" />
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete Sleep Record</AlertDialogTitle>
                  <AlertDialogDescription>
                    Are you sure you want to delete this sleep record? This action cannot be undone.
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

        {sleep.notes && (
          <p className="mt-2 text-sm text-muted-foreground italic">{sleep.notes}</p>
        )}
      </CardContent>
    </Card>
  )
}
