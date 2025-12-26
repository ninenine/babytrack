import { Moon, Sun, Square } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useTimer, useEndSleep } from '@/hooks'
import type { LocalSleep } from '@/db/dexie'

interface ActiveSleepCardProps {
  sleep: LocalSleep
}

export function ActiveSleepCard({ sleep }: ActiveSleepCardProps) {
  const { formattedElapsed } = useTimer(new Date(sleep.startTime))
  const endSleep = useEndSleep()

  const handleEndSleep = async () => {
    await endSleep.mutateAsync({ id: sleep.id })
  }

  const isNight = sleep.type === 'night'

  return (
    <Card className="border-2 border-primary bg-primary/5">
      <CardContent className="p-6 text-center">
        <div className="flex items-center justify-center gap-2 text-muted-foreground mb-2">
          {isNight ? (
            <Moon className="h-5 w-5" />
          ) : (
            <Sun className="h-5 w-5" />
          )}
          <span className="text-sm font-medium uppercase tracking-wide">
            {sleep.type} in progress
          </span>
        </div>

        <div className="text-5xl font-bold font-mono mb-4">
          {formattedElapsed}
        </div>

        <Button
          variant="destructive"
          size="lg"
          className="gap-2"
          onClick={handleEndSleep}
          disabled={endSleep.isPending}
        >
          <Square className="h-4 w-4 fill-current" />
          {endSleep.isPending ? 'Ending...' : 'End Sleep'}
        </Button>
      </CardContent>
    </Card>
  )
}
