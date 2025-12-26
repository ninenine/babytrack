import { Cloud, CloudOff, RefreshCw, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Badge } from '@/components/ui/badge'
import { useOnline, useSync } from '@/hooks'

export function SyncIndicator() {
  const isOnline = useOnline()
  const { isSyncing, pendingCount, error, syncPendingEvents } = useSync()

  const handleSync = async () => {
    try {
      await syncPendingEvents()
    } catch (e) {
      console.error('Sync failed:', e)
    }
  }

  // Offline state
  if (!isOnline) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" className="relative" disabled>
              <CloudOff className="h-5 w-5 text-muted-foreground" />
              {pendingCount > 0 && (
                <Badge
                  variant="secondary"
                  className="absolute -top-1 -right-1 h-5 w-5 p-0 flex items-center justify-center text-xs"
                >
                  {pendingCount}
                </Badge>
              )}
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Offline</p>
            {pendingCount > 0 && (
              <p className="text-xs text-muted-foreground">
                {pendingCount} changes pending
              </p>
            )}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  // Error state
  if (error) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="relative"
              onClick={handleSync}
            >
              <AlertCircle className="h-5 w-5 text-destructive" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Sync error: {error}</p>
            <p className="text-xs text-muted-foreground">Tap to retry</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  // Syncing state
  if (isSyncing) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" disabled>
              <RefreshCw className="h-5 w-5 animate-spin text-primary" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Syncing...</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  // Pending changes
  if (pendingCount > 0) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="relative"
              onClick={handleSync}
            >
              <Cloud className="h-5 w-5 text-warning" />
              <Badge
                variant="secondary"
                className="absolute -top-1 -right-1 h-5 w-5 p-0 flex items-center justify-center text-xs bg-warning text-warning-foreground"
              >
                {pendingCount}
              </Badge>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{pendingCount} changes pending</p>
            <p className="text-xs text-muted-foreground">Tap to sync now</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  // Synced state
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="ghost" size="icon" onClick={handleSync}>
            <Cloud className="h-5 w-5 text-green-500" />
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>All synced</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
