import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Loader2, ArrowRight, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { DateTimePicker } from '@/components/shared/datetime-picker'
import { useFamilyStore } from '@/stores/family.store'
import { apiClient } from '@/lib/api-client'
import { API_ENDPOINTS } from '@/lib/constants'
import { toAPIDateTime } from '@/lib/dates'

type Step = 'family' | 'child'

export function OnboardingPage() {
  const navigate = useNavigate()
  const { setFamilies } = useFamilyStore()

  const [step, setStep] = useState<Step>('family')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Form state
  const [familyName, setFamilyName] = useState('')
  const [familyId, setFamilyId] = useState<string | null>(null)
  const [childName, setChildName] = useState('')
  const [childDob, setChildDob] = useState<Date | undefined>(undefined)

  async function handleCreateFamily(e: React.FormEvent) {
    e.preventDefault()
    if (!familyName.trim()) return

    setIsLoading(true)
    setError(null)

    try {
      const { data } = await apiClient.post<{
        id: string
        name: string
      }>(API_ENDPOINTS.FAMILIES.BASE, {
        name: familyName,
      })

      setFamilyId(data.id)
      setStep('child')
    } catch (err) {
      console.error('Create family error:', err)
      setError('Failed to create family. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  async function handleAddChild(e: React.FormEvent) {
    e.preventDefault()
    if (!childName.trim() || !childDob || !familyId) return

    setIsLoading(true)
    setError(null)

    try {
      const { data: child } = await apiClient.post<{
        id: string
        name: string
        date_of_birth: string
      }>(API_ENDPOINTS.FAMILIES.CHILDREN(familyId), {
        name: childName,
        date_of_birth: toAPIDateTime(childDob),
      })

      // Set the family with the new child
      setFamilies([
        {
          id: familyId,
          name: familyName,
          children: [
            {
              id: child.id,
              name: child.name,
              dateOfBirth: child.date_of_birth,
            },
          ],
        },
      ])

      navigate('/')
    } catch (err) {
      console.error('Add child error:', err)
      setError('Failed to add child. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-linear-to-b from-background to-muted p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <img src="/logo-64.png" alt="BabyTrack" className="h-16 w-16" />
          </div>
          <CardTitle className="text-2xl">Welcome to BabyTrack</CardTitle>
          <CardDescription>
            {step === 'family'
              ? "Let's set up your family"
              : 'Now add your first child'}
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* Progress indicator */}
          <div className="flex items-center justify-center gap-2">
            <div
              className={`flex items-center justify-center w-8 h-8 rounded-full ${
                step === 'family'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-green-500 text-white'
              }`}
            >
              {step === 'child' ? <Check className="h-4 w-4" /> : '1'}
            </div>
            <div className="w-12 h-0.5 bg-muted" />
            <div
              className={`flex items-center justify-center w-8 h-8 rounded-full ${
                step === 'child'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted text-muted-foreground'
              }`}
            >
              2
            </div>
          </div>

          {error && (
            <div className="p-3 rounded-md bg-destructive/10 text-destructive text-sm text-center">
              {error}
            </div>
          )}

          {step === 'family' ? (
            <form onSubmit={handleCreateFamily} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="familyName">Family Name</Label>
                <Input
                  id="familyName"
                  placeholder="e.g., The Smiths"
                  value={familyName}
                  onChange={(e) => setFamilyName(e.target.value)}
                  required
                  disabled={isLoading}
                />
              </div>

              <Button
                type="submit"
                className="w-full"
                disabled={isLoading || !familyName.trim()}
              >
                {isLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                ) : (
                  <ArrowRight className="h-4 w-4 mr-2" />
                )}
                Continue
              </Button>
            </form>
          ) : (
            <form onSubmit={handleAddChild} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="childName">Child's Name</Label>
                <Input
                  id="childName"
                  placeholder="e.g., Emma"
                  value={childName}
                  onChange={(e) => setChildName(e.target.value)}
                  required
                  disabled={isLoading}
                />
              </div>

              <div className="space-y-2">
                <Label>Date & Time of Birth</Label>
                <DateTimePicker
                  date={childDob}
                  onDateChange={setChildDob}
                  placeholder="Select date & time of birth"
                  disabled={isLoading}
                  toDate={new Date()}
                />
              </div>

              <Button
                type="submit"
                className="w-full"
                disabled={isLoading || !childName.trim() || !childDob}
              >
                {isLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                ) : (
                  <Check className="h-4 w-4 mr-2" />
                )}
                Get Started
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
