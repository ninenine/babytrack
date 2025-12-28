import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
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
import { useSessionStore } from '@/stores/session.store'
import { useTheme } from '@/hooks/use-theme'
import { useQueryClient } from '@tanstack/react-query'
import { db } from '@/db/dexie'
import { toast } from 'sonner'
import { ManageChildrenCard, InviteMemberCard, ManageFamilyCard } from '@/components/settings'
import { API_ENDPOINTS } from '@/lib/constants'

export function SettingsPage() {
  const { user, clearSession } = useSessionStore()
  const { isDark, toggleTheme } = useTheme()
  const queryClient = useQueryClient()
  const [isClearing, setIsClearing] = useState(false)
  const [version, setVersion] = useState<string>('...')
  const [notificationsEnabled, setNotificationsEnabled] = useState(() => {
    return localStorage.getItem('notifications') === 'true'
  })

  useEffect(() => {
    fetch(API_ENDPOINTS.VERSION)
      .then((res) => res.json())
      .then((data) => setVersion(data.version))
      .catch(() => setVersion('unknown'))
  }, [])

  const handleLogout = () => {
    clearSession()
    window.location.href = '/login'
  }

  const handleNotificationsToggle = async (enabled: boolean) => {
    if (enabled && 'Notification' in window) {
      const permission = await Notification.requestPermission()
      if (permission === 'granted') {
        setNotificationsEnabled(true)
        localStorage.setItem('notifications', 'true')
      }
    } else {
      setNotificationsEnabled(false)
      localStorage.setItem('notifications', 'false')
    }
  }

  const handleClearAndSync = async () => {
    setIsClearing(true)
    try {
      // Clear all tables in the local database
      await Promise.all([
        db.feedings.clear(),
        db.sleep.clear(),
        db.medications.clear(),
        db.medicationLogs.clear(),
        db.notes.clear(),
        db.vaccinations.clear(),
        db.appointments.clear(),
        db.pendingEvents.clear(),
      ])

      // Invalidate all queries to trigger a fresh sync from server
      await queryClient.invalidateQueries()

      toast.success('Local data cleared and syncing from server')
    } catch (error) {
      toast.error('Failed to clear local data')
      console.error('Clear and sync error:', error)
    } finally {
      setIsClearing(false)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Settings</h1>

      {/* Manage Children */}
      <ManageChildrenCard />

      {/* Family Members */}
      <InviteMemberCard />

      {/* Manage Family */}
      <ManageFamilyCard />

      {/* Account */}
      <Card>
        <CardHeader>
          <CardTitle>Account</CardTitle>
          <CardDescription>Manage your account settings</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium">{user?.name || 'User'}</div>
              <div className="text-sm text-muted-foreground">{user?.email || 'No email'}</div>
            </div>
          </div>
          <Separator />
          <Button variant="destructive" onClick={handleLogout}>
            Log out
          </Button>
        </CardContent>
      </Card>

      {/* Preferences */}
      <Card>
        <CardHeader>
          <CardTitle>Preferences</CardTitle>
          <CardDescription>Customize your experience</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <Label htmlFor="dark-mode">Dark Mode</Label>
              <p className="text-sm text-muted-foreground">
                Toggle dark theme for the app
              </p>
            </div>
            <Switch
              id="dark-mode"
              checked={isDark}
              onCheckedChange={toggleTheme}
            />
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <div>
              <Label htmlFor="notifications">Push Notifications</Label>
              <p className="text-sm text-muted-foreground">
                Get reminders for medications and appointments
              </p>
            </div>
            <Switch
              id="notifications"
              checked={notificationsEnabled}
              onCheckedChange={handleNotificationsToggle}
            />
          </div>
          {notificationsEnabled && (
            <div className="rounded-md bg-muted p-3 text-sm text-muted-foreground">
              <p className="font-medium text-foreground mb-2">Notification Schedule</p>
              <ul className="space-y-1">
                <li>Medications: checked every 15 minutes</li>
                <li>Appointments: checked every 30 minutes</li>
                <li>Vaccinations: checked every 6 hours</li>
                <li>Sleep insights: checked hourly, summary at 7 AM</li>
              </ul>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Data */}
      <Card>
        <CardHeader>
          <CardTitle>Data</CardTitle>
          <CardDescription>Manage your local data</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <Label>Clear Local Data</Label>
              <p className="text-sm text-muted-foreground">
                Clear cached data and re-sync from server
              </p>
            </div>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="outline" disabled={isClearing}>
                  {isClearing ? 'Clearing...' : 'Clear & Sync'}
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Clear local data?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will clear all locally cached data and re-download everything from the server.
                    Any unsynced changes will be lost.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction onClick={handleClearAndSync}>
                    Clear & Sync
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </CardContent>
      </Card>

      {/* About */}
      <Card>
        <CardHeader>
          <CardTitle>About</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          <p>BabyTrack {version}</p>
          <p className="mt-1">Track your baby's activities with ease</p>
        </CardContent>
      </Card>
    </div>
  )
}
