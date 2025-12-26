import { format, isToday, differenceInMinutes } from 'date-fns'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Utensils, Moon, Pill, Syringe, Calendar, Plus, Clock, AlertTriangle } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useFamilyStore } from '@/stores/family.store'
import {
  useFeedings,
  useSleep,
  useMedications,
  useVaccinations,
  useAppointments,
  useTimer,
} from '@/hooks'

function ActiveSleepBanner() {
  const { activeSleep } = useSleep()
  const { formattedElapsed } = useTimer(activeSleep ? new Date(activeSleep.startTime) : null)

  if (!activeSleep) return null

  return (
    <Card className="border-primary bg-primary/5">
      <CardContent className="py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-full bg-primary/10">
              <Moon className="h-5 w-5 text-primary" />
            </div>
            <div>
              <div className="font-medium capitalize">{activeSleep.type} in progress</div>
              <div className="text-sm text-muted-foreground">
                Started {format(new Date(activeSleep.startTime), 'h:mm a')}
              </div>
            </div>
          </div>
          <div className="text-right">
            <div className="text-2xl font-bold font-mono">{formattedElapsed}</div>
            <Button size="sm" variant="outline" asChild className="mt-1">
              <Link to="/sleep">View</Link>
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function TodaysSummary() {
  const { feedings, isLoading: feedingsLoading } = useFeedings()
  const { sleepRecords, isLoading: sleepLoading } = useSleep()
  const { activeMedications } = useMedications()

  // Count today's feedings
  const todaysFeedings = feedings.filter((f) => isToday(new Date(f.startTime)))

  // Calculate today's sleep hours
  const todaysSleep = sleepRecords.filter((s) => {
    const start = new Date(s.startTime)
    return isToday(start) && s.endTime
  })
  const totalSleepMinutes = todaysSleep.reduce((acc, s) => {
    if (!s.endTime) return acc
    return acc + differenceInMinutes(new Date(s.endTime), new Date(s.startTime))
  }, 0)
  const sleepHours = Math.floor(totalSleepMinutes / 60)
  const sleepMins = totalSleepMinutes % 60

  const isLoading = feedingsLoading || sleepLoading

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-lg">Today's Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4 text-center">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-lg">Today's Summary</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-3 gap-4 text-center">
          <Link to="/feeding" className="block hover:bg-muted/50 rounded-lg p-2 -m-2 transition-colors">
            <div className="text-2xl font-bold text-primary">{todaysFeedings.length}</div>
            <div className="text-xs text-muted-foreground">Feedings</div>
          </Link>
          <Link to="/sleep" className="block hover:bg-muted/50 rounded-lg p-2 -m-2 transition-colors">
            <div className="text-2xl font-bold text-primary">
              {sleepHours > 0 ? `${sleepHours}h` : '0h'}
              {sleepMins > 0 && <span className="text-lg">{sleepMins}m</span>}
            </div>
            <div className="text-xs text-muted-foreground">Sleep</div>
          </Link>
          <Link to="/medications" className="block hover:bg-muted/50 rounded-lg p-2 -m-2 transition-colors">
            <div className="text-2xl font-bold text-primary">{activeMedications.length}</div>
            <div className="text-xs text-muted-foreground">Active Meds</div>
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}

function LastFeeding() {
  const { feedings, isLoading } = useFeedings()

  if (isLoading) return null

  const lastFeeding = feedings[0]
  if (!lastFeeding) return null

  const feedingTime = new Date(lastFeeding.startTime)
  const minutesAgo = differenceInMinutes(new Date(), feedingTime)
  const hoursAgo = Math.floor(minutesAgo / 60)

  let timeAgoText = ''
  if (minutesAgo < 60) {
    timeAgoText = `${minutesAgo}m ago`
  } else if (hoursAgo < 24) {
    timeAgoText = `${hoursAgo}h ago`
  } else {
    timeAgoText = format(feedingTime, 'MMM d')
  }

  const feedingTypeLabels: Record<string, string> = {
    breast: 'Breastfeeding',
    bottle: 'Bottle',
    formula: 'Formula',
    solid: 'Solid food',
  }

  return (
    <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/50">
      <Utensils className="h-4 w-4 text-muted-foreground" />
      <div className="flex-1">
        <div className="text-sm font-medium">Last feeding</div>
        <div className="text-xs text-muted-foreground">
          {feedingTypeLabels[lastFeeding.type] || lastFeeding.type}
          {lastFeeding.amount && ` · ${lastFeeding.amount}${lastFeeding.unit}`}
        </div>
      </div>
      <div className="flex items-center gap-1 text-xs text-muted-foreground">
        <Clock className="h-3 w-3" />
        {timeAgoText}
      </div>
    </div>
  )
}

function UpcomingSection() {
  const { upcoming: upcomingVax, overdue: overdueVax, isLoading: vaxLoading } = useVaccinations()
  const { upcoming: upcomingApt, isLoading: aptLoading } = useAppointments()

  const isLoading = vaxLoading || aptLoading

  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-lg">Upcoming</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </CardContent>
      </Card>
    )
  }

  const nextVax = upcomingVax[0]
  const nextOverdueVax = overdueVax[0]
  const nextApt = upcomingApt[0]

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-lg">Upcoming</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {/* Overdue vaccination warning */}
        {nextOverdueVax && (
          <Link
            to="/vaccinations"
            className="flex items-center gap-3 p-3 rounded-lg bg-destructive/10 hover:bg-destructive/20 transition-colors"
          >
            <AlertTriangle className="h-4 w-4 text-destructive" />
            <div className="flex-1">
              <div className="text-sm font-medium text-destructive">Overdue Vaccination</div>
              <div className="text-xs text-muted-foreground">
                {nextOverdueVax.name} (Dose {nextOverdueVax.dose})
              </div>
            </div>
            <Badge variant="destructive" className="text-xs">
              {format(new Date(nextOverdueVax.scheduledAt), 'MMM d')}
            </Badge>
          </Link>
        )}

        {/* Next vaccination */}
        {nextVax ? (
          <Link
            to="/vaccinations"
            className="flex items-center gap-3 p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors"
          >
            <Syringe className="h-4 w-4 text-muted-foreground" />
            <div className="flex-1">
              <div className="text-sm font-medium">{nextVax.name}</div>
              <div className="text-xs text-muted-foreground">Dose {nextVax.dose}</div>
            </div>
            <Badge variant="secondary" className="text-xs">
              {format(new Date(nextVax.scheduledAt), 'MMM d')}
            </Badge>
          </Link>
        ) : !nextOverdueVax ? (
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <Syringe className="h-4 w-4" />
            <span>No upcoming vaccinations</span>
          </div>
        ) : null}

        {/* Next appointment */}
        {nextApt ? (
          <Link
            to="/appointments"
            className="flex items-center gap-3 p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors"
          >
            <Calendar className="h-4 w-4 text-muted-foreground" />
            <div className="flex-1">
              <div className="text-sm font-medium">{nextApt.title}</div>
              <div className="text-xs text-muted-foreground">
                {nextApt.provider && `${nextApt.provider} · `}
                {format(new Date(nextApt.scheduledAt), 'h:mm a')}
              </div>
            </div>
            <Badge variant="secondary" className="text-xs">
              {format(new Date(nextApt.scheduledAt), 'MMM d')}
            </Badge>
          </Link>
        ) : (
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <Calendar className="h-4 w-4" />
            <span>No upcoming appointments</span>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export function DashboardTab() {
  const currentChild = useFamilyStore((state) => state.currentChild)

  if (!currentChild) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No child selected
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Active Sleep Banner */}
      <ActiveSleepBanner />

      {/* Quick Actions */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-lg">Quick Actions</CardTitle>
          <CardDescription>Log activities for {currentChild.name}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-2">
            <Button variant="outline" className="h-16 flex-col gap-1" asChild>
              <Link to="/feeding">
                <Utensils className="h-5 w-5" />
                <span className="text-xs">Log Feeding</span>
              </Link>
            </Button>
            <Button variant="outline" className="h-16 flex-col gap-1" asChild>
              <Link to="/sleep">
                <Moon className="h-5 w-5" />
                <span className="text-xs">Log Sleep</span>
              </Link>
            </Button>
            <Button variant="outline" className="h-16 flex-col gap-1" asChild>
              <Link to="/medications">
                <Pill className="h-5 w-5" />
                <span className="text-xs">Log Dose</span>
              </Link>
            </Button>
            <Button variant="outline" className="h-16 flex-col gap-1" asChild>
              <Link to="/notes">
                <Plus className="h-5 w-5" />
                <span className="text-xs">Add Note</span>
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Today's Summary */}
      <TodaysSummary />

      {/* Last Feeding */}
      <LastFeeding />

      {/* Upcoming */}
      <UpcomingSection />
    </div>
  )
}
