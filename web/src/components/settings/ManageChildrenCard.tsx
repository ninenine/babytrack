import { useState } from 'react'
import { format, parseISO } from 'date-fns'
import { Pencil, Trash2, Plus } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useFamilyStore, type Child } from '@/stores/family.store'
import { useSessionStore } from '@/stores/session.store'
import { API_ENDPOINTS } from '@/lib/constants'
import { toast } from 'sonner'
import { ChildFormDialog } from './ChildFormDialog'
import {
  differenceInYears,
  differenceInMonths,
  differenceInWeeks,
  differenceInDays,
  differenceInHours,
} from 'date-fns'

function formatAge(dateOfBirth: string): string {
  const dob = parseISO(dateOfBirth)
  const now = new Date()

  const years = differenceInYears(now, dob)
  if (years >= 1) return `${years}y`

  const months = differenceInMonths(now, dob)
  if (months >= 1) return `${months}mo`

  const weeks = differenceInWeeks(now, dob)
  if (weeks >= 1) return `${weeks}w`

  const days = differenceInDays(now, dob)
  if (days >= 1) return `${days}d`

  const hours = differenceInHours(now, dob)
  return `${hours}h`
}

export function ManageChildrenCard() {
  const { currentFamily, removeChild } = useFamilyStore()
  const { token } = useSessionStore()
  const [childFormOpen, setChildFormOpen] = useState(false)
  const [editingChild, setEditingChild] = useState<Child | null>(null)
  const [deletingChild, setDeletingChild] = useState<Child | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const children = currentFamily?.children || []

  const handleEdit = (child: Child) => {
    setEditingChild(child)
    setChildFormOpen(true)
  }

  const handleAdd = () => {
    setEditingChild(null)
    setChildFormOpen(true)
  }

  const handleDelete = async () => {
    if (!deletingChild || !currentFamily) return

    setIsDeleting(true)
    try {
      const response = await fetch(
        API_ENDPOINTS.FAMILIES.CHILD_BY_ID(currentFamily.id, deletingChild.id),
        {
          method: 'DELETE',
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      )

      if (!response.ok) {
        throw new Error('Failed to delete child')
      }

      removeChild(deletingChild.id)
      setDeletingChild(null)
      toast.success('Child removed')
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to delete child')
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Children</CardTitle>
              <CardDescription>Manage children in your family</CardDescription>
            </div>
            <Button size="sm" onClick={handleAdd}>
              <Plus className="h-4 w-4 mr-1" />
              Add
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {children.length === 0 ? (
            <div className="text-center py-6 text-muted-foreground">
              <p>No children added yet</p>
              <Button variant="link" onClick={handleAdd}>
                Add your first child
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              {children.map((child) => (
                <div
                  key={child.id}
                  className="flex items-center justify-between p-3 rounded-lg border"
                >
                  <div className="flex items-center gap-3">
                    <Avatar className="h-10 w-10">
                      <AvatarImage src={child.avatarUrl} alt={child.name} />
                      <AvatarFallback>
                        {child.name.charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <div className="font-medium flex items-center gap-2">
                        {child.name}
                        <Badge variant="secondary" className="text-xs">
                          {formatAge(child.dateOfBirth)}
                        </Badge>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        Born {format(parseISO(child.dateOfBirth), 'MMM d, yyyy')}
                        {child.gender && ` · ${child.gender}`}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleEdit(child)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setDeletingChild(child)}
                      disabled={children.length === 1}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <ChildFormDialog
        open={childFormOpen}
        onOpenChange={setChildFormOpen}
        child={editingChild}
      />

      <AlertDialog open={!!deletingChild} onOpenChange={() => !isDeleting && setDeletingChild(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Child</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove {deletingChild?.name}? This will also
              remove all their associated data. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? 'Removing...' : 'Remove'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
