import { useState } from 'react'
import { Copy, Mail, UserPlus, Check, Trash2 } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
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

interface FamilyMember {
  id: string
  name: string
  email: string
  role: 'owner' | 'admin' | 'member'
  status: 'active' | 'pending'
}

// Mock data - in a real app this would come from the API
const mockMembers: FamilyMember[] = [
  { id: '1', name: 'You', email: 'you@example.com', role: 'owner', status: 'active' },
]

export function InviteMemberCard() {
  const { currentFamily } = useFamilyStore()
  const { user } = useSessionStore()
  const [inviteDialogOpen, setInviteDialogOpen] = useState(false)
  const [email, setEmail] = useState('')
  const [isInviting, setIsInviting] = useState(false)
  const [copied, setCopied] = useState(false)
  const [removingMember, setRemovingMember] = useState<FamilyMember | null>(null)

  // Generate an invite link (mock)
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

  const handleRemoveMember = () => {
    if (removingMember) {
      // TODO: Call API to remove member
      console.log('Removing member:', removingMember.id)
      setRemovingMember(null)
    }
  }

  // Build members list with current user
  const members: FamilyMember[] = user
    ? [
        {
          id: user.id,
          name: user.name || 'You',
          email: user.email || '',
          role: 'owner',
          status: 'active',
        },
        // Add mock pending invites for demo
      ]
    : mockMembers

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
            {members.map((member) => (
              <div
                key={member.id}
                className="flex items-center justify-between p-3 rounded-lg border"
              >
                <div className="flex items-center gap-3">
                  <Avatar className="h-10 w-10">
                    <AvatarFallback>
                      {member.name.charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="font-medium flex items-center gap-2">
                      {member.name}
                      {member.id === user?.id && (
                        <span className="text-xs text-muted-foreground">(you)</span>
                      )}
                      <Badge
                        variant={member.role === 'owner' ? 'default' : 'secondary'}
                        className="text-xs"
                      >
                        {member.role}
                      </Badge>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {member.email}
                      {member.status === 'pending' && (
                        <Badge variant="outline" className="ml-2 text-xs">
                          Pending
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>
                {member.id !== user?.id && (
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setRemovingMember(member)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </div>
            ))}
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
