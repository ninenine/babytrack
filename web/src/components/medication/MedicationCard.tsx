import { useState } from 'react'
import { format } from 'date-fns'
import { Pill, Clock, Trash2, Plus, ChevronDown, ChevronUp, Pencil } from 'lucide-react'
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
import { LogDoseDialog } from './LogDoseDialog'
import { MedicationLogFormDialog } from './MedicationLogFormDialog'
import { useDeleteMedication, useMedicationLogs, useDeleteMedicationLog } from '@/hooks'
import type { LocalMedication, LocalMedicationLog } from '@/db/dexie'

interface MedicationCardProps {
  medication: LocalMedication
}

const frequencyLabels: Record<string, string> = {
  once_daily: 'Once daily',
  twice_daily: 'Twice daily',
  three_times_daily: '3x daily',
  four_times_daily: '4x daily',
  every_4_hours: 'Every 4h',
  every_6_hours: 'Every 6h',
  every_8_hours: 'Every 8h',
  as_needed: 'As needed',
  weekly: 'Weekly',
}

interface EnrichedMedicationLog extends LocalMedicationLog {
  medicationName?: string
}

export function MedicationCard({ medication }: MedicationCardProps) {
  const deleteMedication = useDeleteMedication()
  const deleteLog = useDeleteMedicationLog()
  const { logs } = useMedicationLogs(medication.id)
  const [showLogDialog, setShowLogDialog] = useState(false)
  const [showHistory, setShowHistory] = useState(false)
  const [editingLog, setEditingLog] = useState<EnrichedMedicationLog | null>(null)

  const lastLog = logs[0]

  const handleDelete = () => {
    deleteMedication.mutate(medication.id)
  }

  const handleDeleteLog = (logId: string) => {
    deleteLog.mutate(logId)
  }

  return (
    <>
      <Card className="relative group">
        <CardContent className="p-4">
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-full bg-green-100 dark:bg-green-900">
                <Pill className="h-5 w-5 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{medication.name}</span>
                  {medication.pendingSync && (
                    <Badge variant="outline" className="text-xs text-yellow-600">
                      Pending
                    </Badge>
                  )}
                </div>
                <div className="text-sm text-muted-foreground">
                  {medication.dosage} {medication.unit} · {frequencyLabels[medication.frequency] ?? medication.frequency}
                </div>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Button
                size="sm"
                variant="outline"
                className="gap-1"
                onClick={() => setShowLogDialog(true)}
              >
                <Plus className="h-3 w-3" />
                Log
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
                    <AlertDialogTitle>Delete Medication</AlertDialogTitle>
                    <AlertDialogDescription>
                      Are you sure you want to delete {medication.name}? This will also delete all dose logs. This action cannot be undone.
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

          {medication.instructions && (
            <p className="mt-2 text-sm text-muted-foreground">{medication.instructions}</p>
          )}

          {lastLog && (
            <div className="mt-3 flex items-center justify-between">
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <Clock className="h-3 w-3" />
                <span>
                  Last dose: {format(new Date(lastLog.givenAt), 'MMM d, h:mm a')} by {lastLog.givenBy}
                </span>
              </div>
              {logs.length > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 px-2 text-xs"
                  onClick={() => setShowHistory(!showHistory)}
                >
                  {showHistory ? (
                    <>
                      <ChevronUp className="h-3 w-3 mr-1" />
                      Hide history
                    </>
                  ) : (
                    <>
                      <ChevronDown className="h-3 w-3 mr-1" />
                      View history ({logs.length})
                    </>
                  )}
                </Button>
              )}
            </div>
          )}

          {showHistory && logs.length > 0 && (
            <div className="mt-3 border-t pt-3 space-y-2">
              {logs.map((log) => (
                <div
                  key={log.id}
                  className="flex items-center justify-between py-1 px-2 rounded hover:bg-muted/50 group/log"
                >
                  <div className="text-sm">
                    <span className="font-medium">
                      {format(new Date(log.givenAt), 'MMM d, h:mm a')}
                    </span>
                    <span className="text-muted-foreground">
                      {' '}— {log.dosage} by {log.givenBy}
                    </span>
                    {log.notes && (
                      <span className="text-muted-foreground italic"> "{log.notes}"</span>
                    )}
                  </div>
                  <div className="flex items-center gap-1 opacity-0 group-hover/log:opacity-100 transition-opacity">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={() => setEditingLog({ ...log, medicationName: medication.name })}
                    >
                      <Pencil className="h-3 w-3" />
                    </Button>
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-6 w-6">
                          <Trash2 className="h-3 w-3 text-destructive" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete Log Entry</AlertDialogTitle>
                          <AlertDialogDescription>
                            Are you sure you want to delete this dose log?
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            onClick={() => handleDeleteLog(log.id)}
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                          >
                            Delete
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <LogDoseDialog
        medication={medication}
        open={showLogDialog}
        onOpenChange={setShowLogDialog}
      />

      <MedicationLogFormDialog
        open={!!editingLog}
        onOpenChange={(open) => !open && setEditingLog(null)}
        log={editingLog}
      />
    </>
  )
}
