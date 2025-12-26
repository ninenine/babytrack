import { useState } from 'react'
import { format, isPast } from 'date-fns'
import { Syringe, Check, AlertTriangle, Pencil } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { RecordVaccinationDialog } from './RecordVaccinationDialog'
import type { LocalVaccination } from '@/db/dexie'

interface VaccinationCardProps {
  vaccination: LocalVaccination
  onEdit?: (vaccination: LocalVaccination) => void
}

export function VaccinationCard({ vaccination, onEdit }: VaccinationCardProps) {
  const [showRecordDialog, setShowRecordDialog] = useState(false)

  const isOverdue = !vaccination.completed && isPast(new Date(vaccination.scheduledAt))

  return (
    <>
      <Card className={`group ${isOverdue ? 'border-destructive' : ''}`}>
        <CardContent className="p-4">
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className={`p-2 rounded-full ${
                vaccination.completed
                  ? 'bg-green-100 dark:bg-green-900'
                  : isOverdue
                  ? 'bg-red-100 dark:bg-red-900'
                  : 'bg-blue-100 dark:bg-blue-900'
              }`}>
                {vaccination.completed ? (
                  <Check className="h-5 w-5 text-green-600 dark:text-green-400" />
                ) : isOverdue ? (
                  <AlertTriangle className="h-5 w-5 text-red-600 dark:text-red-400" />
                ) : (
                  <Syringe className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                )}
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{vaccination.name}</span>
                  <Badge variant="secondary" className="text-xs">
                    Dose {vaccination.dose}
                  </Badge>
                  {isOverdue && (
                    <Badge variant="destructive" className="text-xs">
                      Overdue
                    </Badge>
                  )}
                  {vaccination.pendingSync && (
                    <Badge variant="outline" className="text-xs text-yellow-600">
                      Pending
                    </Badge>
                  )}
                </div>
                <div className="text-sm text-muted-foreground">
                  {vaccination.completed ? (
                    <>
                      Given: {format(new Date(vaccination.administeredAt!), 'MMM d, yyyy')}
                      {vaccination.provider && ` by ${vaccination.provider}`}
                    </>
                  ) : (
                    <>Scheduled: {format(new Date(vaccination.scheduledAt), 'MMM d, yyyy')}</>
                  )}
                </div>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
                onClick={() => onEdit?.(vaccination)}
              >
                <Pencil className="h-4 w-4 text-muted-foreground" />
              </Button>
              {!vaccination.completed && (
                <Button
                  size="sm"
                  variant={isOverdue ? 'destructive' : 'outline'}
                  onClick={() => setShowRecordDialog(true)}
                >
                  Record
                </Button>
              )}
            </div>
          </div>

          {vaccination.notes && (
            <p className="mt-2 text-sm text-muted-foreground">{vaccination.notes}</p>
          )}

          {vaccination.completed && vaccination.location && (
            <div className="mt-2 text-xs text-muted-foreground">
              Location: {vaccination.location}
              {vaccination.lotNumber && ` · Lot: ${vaccination.lotNumber}`}
            </div>
          )}
        </CardContent>
      </Card>

      <RecordVaccinationDialog
        vaccination={vaccination}
        open={showRecordDialog}
        onOpenChange={setShowRecordDialog}
      />
    </>
  )
}
