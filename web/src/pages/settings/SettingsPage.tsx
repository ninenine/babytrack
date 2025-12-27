import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { useSessionStore } from '@/stores/session.store'
import { useTheme } from '@/hooks/use-theme'
import { ManageChildrenCard, InviteMemberCard } from '@/components/settings'

export function SettingsPage() {
  const { user, clearSession } = useSessionStore()
  const { isDark, toggleTheme } = useTheme()
  const [notificationsEnabled, setNotificationsEnabled] = useState(() => {
    return localStorage.getItem('notifications') === 'true'
  })

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

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Settings</h1>

      {/* Manage Children */}
      <ManageChildrenCard />

      {/* Family Members */}
      <InviteMemberCard />

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

      {/* About */}
      <Card>
        <CardHeader>
          <CardTitle>About</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          <p>BabyTrack v1.0.0</p>
          <p className="mt-1">Track your baby's activities with ease</p>
        </CardContent>
      </Card>
    </div>
  )
}
