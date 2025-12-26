import { Skeleton } from '@/components/ui/skeleton'

export function PageLoader() {
  return (
    <div className="space-y-6 p-4">
      {/* Header skeleton */}
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-9 w-20" />
      </div>

      {/* Tabs skeleton */}
      <Skeleton className="h-10 w-full" />

      {/* Content skeletons */}
      <div className="space-y-3">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    </div>
  )
}

export function HomePageLoader() {
  return (
    <div className="space-y-6 p-4">
      {/* Tabs skeleton */}
      <Skeleton className="h-10 w-full max-w-xs" />

      {/* Dashboard cards */}
      <div className="grid grid-cols-2 gap-4">
        <Skeleton className="h-28 w-full" />
        <Skeleton className="h-28 w-full" />
        <Skeleton className="h-28 w-full" />
        <Skeleton className="h-28 w-full" />
      </div>

      {/* Quick actions */}
      <Skeleton className="h-12 w-full" />
    </div>
  )
}

export function SettingsPageLoader() {
  return (
    <div className="space-y-6 p-4">
      <Skeleton className="h-8 w-24" />

      {/* Settings sections */}
      <div className="space-y-4">
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    </div>
  )
}

export function AuthPageLoader() {
  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <div className="w-full max-w-md space-y-6">
        <div className="flex flex-col items-center gap-4">
          <Skeleton className="h-16 w-16 rounded-full" />
          <Skeleton className="h-8 w-48" />
        </div>
        <Skeleton className="h-12 w-full" />
      </div>
    </div>
  )
}
