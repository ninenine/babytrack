import { useState, useEffect } from 'react'
import { Copy, Mail, UserPlus, Check, Trash2, Loader2 } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
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
import { apiClient } from '@/lib/api-client'
import { API_ENDPOINTS } from '@/lib/constants'

interface FamilyMember {
  id: string
  user_id: string
  name: string
  email: string
  avatar_url?: string
  role: 'admin' | 'member'
  joined_at: string
}

export function InviteMemberCard() {
  const { currentFamily } = useFamilyStore()
  const { user } = useSessionStore()
  const [members, setMembers] = useState<FamilyMember[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [inviteDialogOpen, setInviteDialogOpen] = useState(false)
  const [email, setEmail] = useState('')
  const [isInviting, setIsInviting] = useState(false)
  const [copied, setCopied] = useState(false)
  const [removingMember, setRemovingMember] = useState<FamilyMember | null>(null)

  // Fetch family members
  useEffect(() => {
    if (!currentFamily?.id) return

    const fetchMembers = async () => {
      setIsLoading(true)
      try {
        const { data } = await apiClient.get<FamilyMember[]>(
          API_ENDPOINTS.FAMILIES.MEMBERS(currentFamily.id)
        )
        setMembers(data || [])
      } catch (err) {
        console.error('Failed to fetch family members:', err)
      } finally {
        setIsLoading(false)
      }
    }

    fetchMembers()
  }, [currentFamily?.id])

  // Generate an invite link
  const inviteLink = `${window.location.origin}/invite/${currentFamily?.id || 'family'}`

  const handleCopyLink = async () => {
    await navigator.clipboard.writeText(inviteLink)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleInviteByEmail = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email) return

    setIsInviting(true)
    try {
      // TODO: Call API to send invite
      console.log('Inviting:', email)
      await new Promise((resolve) => setTimeout(resolve, 1000))
      setEmail('')
      setInviteDialogOpen(false)
    } finally {
      setIsInviting(false)
    }
  }

  const handleRemoveMember = async () => {
    if (!removingMember || !currentFamily) return

    try {
      await apiClient.delete(
        `/api/families/${currentFamily.id}/members/${removingMember.user_id}`
      )
      setMembers((prev) => prev.filter((m) => m.id !== removingMember.id))
    } catch (err) {
      console.error('Failed to remove member:', err)
    } finally {
      setRemovingMember(null)
    }
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Family Members</CardTitle>
              <CardDescription>
                Invite others to help track your children
              </CardDescription>
            </div>
            <Button size="sm" onClick={() => setInviteDialogOpen(true)}>
              <UserPlus className="h-4 w-4 mr-1" />
              Invite
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {isLoading ? (
              <div className="flex items-center justify-center py-6">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : members.length === 0 ? (
              <div className="text-center py-6 text-muted-foreground">
                No family members yet
              </div>
            ) : (
              members.map((member) => (
                <div
                  key={member.id}
                  className="flex items-center justify-between p-3 rounded-lg border"
                >
                  <div className="flex items-center gap-3">
                    <Avatar className="h-10 w-10">
                      <AvatarImage src={member.avatar_url} alt={member.name} />
                      <AvatarFallback>
                        {member.name.charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <div className="font-medium flex items-center gap-2">
                        {member.name}
                        {member.user_id === user?.id && (
                          <span className="text-xs text-muted-foreground">(you)</span>
                        )}
                        <Badge
                          variant={member.role === 'admin' ? 'default' : 'secondary'}
                          className="text-xs"
                        >
                          {member.role}
                        </Badge>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {member.email}
                      </div>
                    </div>
                  </div>
                  {member.user_id !== user?.id && member.role !== 'admin' && (
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setRemovingMember(member)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              ))
            )}
          </div>

          {/* Invite Link Section */}
          <div className="mt-6 pt-4 border-t">
            <Label className="text-sm font-medium">Share Invite Link</Label>
            <p className="text-xs text-muted-foreground mb-2">
              Anyone with this link can request to join your family
            </p>
            <div className="flex gap-2">
              <Input
                value={inviteLink}
                readOnly
                className="text-sm text-muted-foreground"
              />
              <Button variant="outline" size="icon" onClick={handleCopyLink}>
                {copied ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Invite by Email Dialog */}
      <Dialog open={inviteDialogOpen} onOpenChange={setInviteDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Invite Family Member</DialogTitle>
            <DialogDescription>
              Send an invitation to join your family. They'll be able to view and
              log activities for your children.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleInviteByEmail} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email Address</Label>
              <div className="flex gap-2">
                <Mail className="h-4 w-4 mt-3 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  placeholder="partner@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setInviteDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isInviting || !email}>
                {isInviting ? 'Sending...' : 'Send Invite'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Remove Member Confirmation */}
      <AlertDialog open={!!removingMember} onOpenChange={() => setRemovingMember(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Family Member</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove {removingMember?.name} from your
              family? They will no longer be able to view or log activities.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRemoveMember}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
