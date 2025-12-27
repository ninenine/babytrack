import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Users, Loader2, Check, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useSessionStore } from '@/stores/session.store'
import { useFamilyStore } from '@/stores/family.store'
import { apiClient } from '@/lib/api-client'
import { API_ENDPOINTS } from '@/lib/constants'

interface FamilyInfo {
  id: string
  name: string
}

export function InvitePage() {
  const navigate = useNavigate()
  const { familyId } = useParams<{ familyId: string }>()
  const { isAuthenticated } = useSessionStore()
  const { setFamilies, families } = useFamilyStore()

  const [isLoading, setIsLoading] = useState(true)
  const [isJoining, setIsJoining] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [family, setFamily] = useState<FamilyInfo | null>(null)
  const [joined, setJoined] = useState(false)

  useEffect(() => {
    if (!familyId) {
      setError('Invalid invite link')
      setIsLoading(false)
      return
    }

    // If not authenticated, redirect to login with return URL
    if (!isAuthenticated) {
      // Store the invite URL to return to after login
      sessionStorage.setItem('invite_redirect', `/invite/${familyId}`)
      navigate('/login')
      return
    }

    // Fetch family info
    fetchFamilyInfo()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [familyId, isAuthenticated])

  async function fetchFamilyInfo() {
    if (!familyId) return

    setIsLoading(true)
    setError(null)

    try {
      const { data } = await apiClient.get<FamilyInfo>(API_ENDPOINTS.FAMILIES.BY_ID(familyId))
      setFamily(data)
    } catch (err) {
      console.error('Fetch family error:', err)
      setError('Family not found or invite link is invalid')
    } finally {
      setIsLoading(false)
    }
  }

  async function handleJoinFamily() {
    if (!familyId || !family) return

    setIsJoining(true)
    setError(null)

    try {
      await apiClient.post(API_ENDPOINTS.FAMILIES.JOIN(familyId))

      // Fetch updated families list
      const { data: updatedFamilies } = await apiClient.get<Array<{
        id: string
        name: string
        children: Array<{
          id: string
          name: string
          date_of_birth: string
          gender?: string
          avatar_url?: string
        }>
      }>>(API_ENDPOINTS.FAMILIES.BASE)

      setFamilies(
        updatedFamilies.map((f) => ({
          id: f.id,
          name: f.name,
          children: f.children.map((c) => ({
            id: c.id,
            name: c.name,
            dateOfBirth: c.date_of_birth,
            gender: c.gender,
            avatarUrl: c.avatar_url,
          })),
        }))
      )

      setJoined(true)

      // Clear any stored redirect
      sessionStorage.removeItem('invite_redirect')

      // Redirect to home after a short delay
      setTimeout(() => {
        navigate('/')
      }, 2000)
    } catch (err) {
      console.error('Join family error:', err)
      setError('Failed to join family. Please try again.')
    } finally {
      setIsJoining(false)
    }
  }

  // Check if already a member
  const isAlreadyMember = families.some((f) => f.id === familyId)

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-b from-background to-muted p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <div className="p-3 rounded-full bg-primary/10">
              <Users className="h-10 w-10 text-primary" />
            </div>
          </div>
          <CardTitle className="text-2xl">Join Family</CardTitle>
          <CardDescription>
            You've been invited to join a family on BabyTrack
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          {isLoading ? (
            <div className="flex flex-col items-center gap-2 py-8">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <p className="text-sm text-muted-foreground">Loading...</p>
            </div>
          ) : error ? (
            <div className="flex flex-col items-center gap-4 py-4">
              <div className="p-3 rounded-full bg-destructive/10">
                <AlertCircle className="h-8 w-8 text-destructive" />
              </div>
              <p className="text-center text-destructive">{error}</p>
              <Button variant="outline" onClick={() => navigate('/login')}>
                Go to Login
              </Button>
            </div>
          ) : joined ? (
            <div className="flex flex-col items-center gap-4 py-4">
              <div className="p-3 rounded-full bg-green-100 dark:bg-green-900/30">
                <Check className="h-8 w-8 text-green-600 dark:text-green-400" />
              </div>
              <div className="text-center">
                <p className="font-medium">Welcome to {family?.name}!</p>
                <p className="text-sm text-muted-foreground mt-1">
                  Redirecting you to the app...
                </p>
              </div>
            </div>
          ) : isAlreadyMember ? (
            <div className="flex flex-col items-center gap-4 py-4">
              <div className="p-3 rounded-full bg-green-100 dark:bg-green-900/30">
                <Check className="h-8 w-8 text-green-600 dark:text-green-400" />
              </div>
              <div className="text-center">
                <p className="font-medium">You're already a member of {family?.name}</p>
              </div>
              <Button onClick={() => navigate('/')}>
                Go to Home
              </Button>
            </div>
          ) : (
            <>
              <div className="p-4 rounded-lg bg-muted text-center">
                <p className="text-sm text-muted-foreground">Family</p>
                <p className="text-xl font-semibold mt-1">{family?.name}</p>
              </div>

              <Button
                className="w-full"
                onClick={handleJoinFamily}
                disabled={isJoining}
              >
                {isJoining ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    Joining...
                  </>
                ) : (
                  <>
                    <Users className="h-4 w-4 mr-2" />
                    Join Family
                  </>
                )}
              </Button>

              <p className="text-xs text-center text-muted-foreground">
                By joining, you'll be able to view and track activities for this family's children.
              </p>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
