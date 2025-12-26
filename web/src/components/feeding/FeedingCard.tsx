import { format } from 'date-fns'
import { Trash2, Pencil } from 'lucide-react'
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
import { useDeleteFeeding, type FeedingType } from '@/hooks'
import type { LocalFeeding } from '@/db/dexie'

interface FeedingCardProps {
  feeding: LocalFeeding
  onEdit?: (feeding: LocalFeeding) => void
}

const feedingTypeConfig: Record<FeedingType, { label: string; icon: string; color: string }> = {
  breast: { label: 'Breast', icon: '🤱', color: 'bg-pink-100 text-pink-800 dark:bg-pink-900 dark:text-pink-200' },
  bottle: { label: 'Bottle', icon: '🍼', color: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' },
  formula: { label: 'Formula', icon: '🥛', color: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200' },
  solid: { label: 'Solid', icon: '🥣', color: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200' },
}

function formatDuration(start: Date, end?: Date): string {
  if (!end) return 'In progress'
  const diff = end.getTime() - start.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 60) return `${minutes} min`
  const hours = Math.floor(minutes / 60)
  const remainingMins = minutes % 60
  return remainingMins > 0 ? `${hours}h ${remainingMins}m` : `${hours}h`
}

export function FeedingCard({ feeding, onEdit }: FeedingCardProps) {
  const deleteFeeding = useDeleteFeeding()
  const config = feedingTypeConfig[feeding.type]

  const handleDelete = () => {
    deleteFeeding.mutate(feeding.id)
  }

  return (
    <Card className="relative group cursor-pointer hover:bg-accent/50 transition-colors" onClick={() => onEdit?.(feeding)}>
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-3">
            <span className="text-2xl">{config.icon}</span>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-medium">{config.label}</span>
                {feeding.side && (
                  <Badge variant="secondary" className="text-xs">
                    {feeding.side}
                  </Badge>
                )}
                {feeding.pendingSync && (
                  <Badge variant="outline" className="text-xs text-yellow-600">
                    Pending
                  </Badge>
                )}
              </div>
              <div className="text-sm text-muted-foreground">
                {format(new Date(feeding.startTime), 'h:mm a')}
                {feeding.endTime && ` - ${format(new Date(feeding.endTime), 'h:mm a')}`}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <div className="text-right">
              {feeding.amount && (
                <div className="font-medium">
                  {feeding.amount} {feeding.unit}
                </div>
              )}
              {feeding.endTime && (
                <div className="text-sm text-muted-foreground">
                  {formatDuration(new Date(feeding.startTime), new Date(feeding.endTime))}
                </div>
              )}
            </div>

            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
              onClick={(e) => { e.stopPropagation(); onEdit?.(feeding) }}
            >
              <Pencil className="h-4 w-4 text-muted-foreground" />
            </Button>

            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
                  onClick={(e) => e.stopPropagation()}
                >
                  <Trash2 className="h-4 w-4 text-muted-foreground" />
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete Feeding</AlertDialogTitle>
                  <AlertDialogDescription>
                    Are you sure you want to delete this feeding record? This action cannot be undone.
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

        {feeding.notes && (
          <p className="mt-2 text-sm text-muted-foreground italic">{feeding.notes}</p>
        )}
      </CardContent>
    </Card>
  )
}
