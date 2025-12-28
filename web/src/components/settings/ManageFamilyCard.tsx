import { useState } from 'react'
import { Pencil, LogOut, Trash2 } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import { useFamilyStore } from '@/stores/family.store'
import { useSessionStore } from '@/stores/session.store'
import { API_ENDPOINTS } from '@/lib/constants'
import { toast } from 'sonner'

export function ManageFamilyCard() {
  const { currentFamily, updateFamily, removeFamily } = useFamilyStore()
  const { token, clearSession } = useSessionStore()
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [leaveDialogOpen, setLeaveDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [familyName, setFamilyName] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  if (!currentFamily) return null

  const handleEditOpen = () => {
    setFamilyName(currentFamily.name)
    setEditDialogOpen(true)
  }

  const handleEditSave = async () => {
    if (!familyName.trim()) return

    setIsSubmitting(true)
    try {
      const response = await fetch(API_ENDPOINTS.FAMILIES.BY_ID(currentFamily.id), {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name: familyName.trim() }),
      })

      if (!response.ok) {
        throw new Error('Failed to update family')
      }

      updateFamily(familyName.trim())
      setEditDialogOpen(false)
      toast.success('Family name updated')
    } catch {
      toast.error('Failed to update family name')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleLeave = async () => {
    setIsSubmitting(true)
    try {
      const response = await fetch(API_ENDPOINTS.FAMILIES.LEAVE(currentFamily.id), {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Failed to leave family')
      }

      removeFamily(currentFamily.id)
      setLeaveDialogOpen(false)
      toast.success('You have left the family')

      // If no families left, redirect to onboarding
      const remainingFamilies = useFamilyStore.getState().families
      if (remainingFamilies.length === 0) {
        clearSession()
        window.location.href = '/onboarding'
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to leave family')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async () => {
    setIsSubmitting(true)
    try {
      const response = await fetch(API_ENDPOINTS.FAMILIES.BY_ID(currentFamily.id), {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Failed to delete family')
      }

      removeFamily(currentFamily.id)
      setDeleteDialogOpen(false)
      toast.success('Family deleted')

      // If no families left, redirect to onboarding
      const remainingFamilies = useFamilyStore.getState().families
      if (remainingFamilies.length === 0) {
        window.location.href = '/onboarding'
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to delete family')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Family</CardTitle>
          <CardDescription>Manage your family settings</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium">{currentFamily.name}</div>
              <div className="text-sm text-muted-foreground">
                {currentFamily.children.length} {currentFamily.children.length === 1 ? 'child' : 'children'}
              </div>
            </div>
            <Button variant="outline" size="sm" onClick={handleEditOpen}>
              <Pencil className="h-4 w-4 mr-1" />
              Rename
            </Button>
          </div>

          <div className="flex gap-2 pt-2">
            <Button
              variant="outline"
              size="sm"
              className="text-muted-foreground"
              onClick={() => setLeaveDialogOpen(true)}
            >
              <LogOut className="h-4 w-4 mr-1" />
              Leave Family
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-destructive hover:text-destructive"
              onClick={() => setDeleteDialogOpen(true)}
            >
              <Trash2 className="h-4 w-4 mr-1" />
              Delete Family
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Edit Family Name Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Rename Family</DialogTitle>
            <DialogDescription>
              Enter a new name for your family.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Label htmlFor="familyName">Family Name</Label>
            <Input
              id="familyName"
              value={familyName}
              onChange={(e) => setFamilyName(e.target.value)}
              placeholder="Enter family name"
              className="mt-2"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleEditSave} disabled={isSubmitting || !familyName.trim()}>
              {isSubmitting ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Leave Family Confirmation */}
      <AlertDialog open={leaveDialogOpen} onOpenChange={setLeaveDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Leave Family</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to leave {currentFamily.name}? You will lose access to all
              family data. You can rejoin later with an invite link.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isSubmitting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleLeave}
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Leaving...' : 'Leave'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Delete Family Confirmation */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Family</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete {currentFamily.name}? This will permanently delete
              all family data including children, feedings, sleep records, and more.
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isSubmitting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isSubmitting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isSubmitting ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
