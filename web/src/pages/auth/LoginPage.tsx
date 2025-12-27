import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useSessionStore } from '@/stores/session.store'
import { useFamilyStore } from '@/stores/family.store'
import { apiClient } from '@/lib/api-client'
import { API_ENDPOINTS } from '@/lib/constants'

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { setSession, isAuthenticated } = useSessionStore()
  const { setFamilies } = useFamilyStore()

  // Handle OAuth callback
  useEffect(() => {
    const token = searchParams.get('token')

    if (token) {
      handleAuthCallback(token)
    } else if (isAuthenticated) {
      navigate('/')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams, isAuthenticated])

  async function handleAuthCallback(token: string) {
    setIsLoading(true)
    setError(null)

    try {
      // Store token temporarily to make API call
      useSessionStore.setState({ token })

      // Fetch user info
      const { data: user } = await apiClient.get<{
        id: string
        email: string
        name: string
        avatar_url?: string
      }>(API_ENDPOINTS.AUTH.ME)

      // Set session
      setSession(
        {
          id: user.id,
          email: user.email,
          name: user.name,
          avatarUrl: user.avatar_url,
        },
        token
      )

      // Fetch families
      const { data: families } = await apiClient.get<Array<{
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

      if (families.length > 0) {
        setFamilies(
          families.map((f) => ({
            id: f.id,
            name: f.name,
            children: (f.children || []).map((c) => ({
              id: c.id,
              name: c.name,
              dateOfBirth: c.date_of_birth,
              gender: c.gender,
              avatarUrl: c.avatar_url,
            })),
          }))
        )
      }

      // Check for invite redirect
      const inviteRedirect = sessionStorage.getItem('invite_redirect')
      if (inviteRedirect) {
        navigate(inviteRedirect)
      } else if (families.length > 0) {
        navigate('/')
      } else {
        navigate('/onboarding')
      }
    } catch (err) {
      console.error('Auth callback error:', err)
      setError('Failed to complete login. Please try again.')
      useSessionStore.getState().clearSession()
    } finally {
      setIsLoading(false)
    }
  }

  function handleGoogleLogin() {
    window.location.href = API_ENDPOINTS.AUTH.GOOGLE
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-linear-to-b from-background to-muted p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <img src="/logo-64.png" alt="BabyTrack" className="h-16 w-16" />
          </div>
          <CardTitle className="text-2xl">BabyTrack</CardTitle>
          <CardDescription>
            Track your baby's feeding, sleep, and more
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <div className="p-3 rounded-md bg-destructive/10 text-destructive text-sm text-center">
              {error}
            </div>
          )}

          {isLoading ? (
            <div className="flex flex-col items-center gap-2 py-4">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <p className="text-sm text-muted-foreground">Signing you in...</p>
            </div>
          ) : (
            <Button
              variant="outline"
              className="w-full h-12 gap-2"
              onClick={handleGoogleLogin}
            >
              <svg className="h-5 w-5" viewBox="0 0 24 24">
                <path
                  fill="currentColor"
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                />
                <path
                  fill="currentColor"
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                />
                <path
                  fill="currentColor"
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                />
                <path
                  fill="currentColor"
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                />
              </svg>
              Continue with Google
            </Button>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
