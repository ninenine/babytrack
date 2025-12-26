import { Moon, Sun } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useStartSleep } from '@/hooks'

export function StartSleepButtons() {
  const startSleep = useStartSleep()

  const handleStartNap = async () => {
    await startSleep.mutateAsync({ type: 'nap' })
  }

  const handleStartNight = async () => {
    await startSleep.mutateAsync({ type: 'night' })
  }

  return (
    <div className="grid grid-cols-2 gap-3">
      <Button
        size="lg"
        className="h-20 flex-col gap-2"
        onClick={handleStartNap}
        disabled={startSleep.isPending}
      >
        <Sun className="h-6 w-6" />
        <span>Start Nap</span>
      </Button>
      <Button
        size="lg"
        variant="secondary"
        className="h-20 flex-col gap-2"
        onClick={handleStartNight}
        disabled={startSleep.isPending}
      >
        <Moon className="h-6 w-6" />
        <span>Start Night Sleep</span>
      </Button>
    </div>
  )
}
