import { Check, ChevronsUpDown, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { useFamilyStore } from '@/stores/family.store'
import { cn } from '@/lib/utils'
import {
  differenceInYears,
  differenceInMonths,
  differenceInWeeks,
  differenceInDays,
  differenceInHours,
  parseISO,
} from 'date-fns'

function formatAge(dateOfBirth: string): string {
  const dob = parseISO(dateOfBirth)
  const now = new Date()

  const years = differenceInYears(now, dob)
  if (years >= 1) {
    return `${years}y`
  }

  const months = differenceInMonths(now, dob)
  if (months >= 1) {
    return `${months}mo`
  }

  const weeks = differenceInWeeks(now, dob)
  if (weeks >= 1) {
    return `${weeks}w`
  }

  const days = differenceInDays(now, dob)
  if (days >= 1) {
    return `${days}d`
  }

  const hours = differenceInHours(now, dob)
  return `${hours}h`
}

export function ChildSwitcher() {
  const { currentFamily, currentChild, setCurrentChild } = useFamilyStore()

  if (!currentFamily || !currentChild) {
    return null
  }

  const children = currentFamily.children

  // If only one child, show simple badge
  if (children.length === 1) {
    return (
      <div className="flex items-center gap-2 px-2 py-1 rounded-md bg-muted">
        <Avatar className="h-6 w-6">
          <AvatarImage src={currentChild.avatarUrl} alt={currentChild.name} />
          <AvatarFallback className="text-xs">
            {currentChild.name.charAt(0).toUpperCase()}
          </AvatarFallback>
        </Avatar>
        <span className="text-sm font-medium">{currentChild.name}</span>
        <Badge variant="secondary" className="text-xs">
          {formatAge(currentChild.dateOfBirth)}
        </Badge>
      </div>
    )
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <Avatar className="h-5 w-5">
            <AvatarImage src={currentChild.avatarUrl} alt={currentChild.name} />
            <AvatarFallback className="text-xs">
              {currentChild.name.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <span className="max-w-25 truncate">{currentChild.name}</span>
          <Badge variant="secondary" className="text-xs">
            {formatAge(currentChild.dateOfBirth)}
          </Badge>
          <ChevronsUpDown className="h-4 w-4 opacity-50" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel>Switch Child</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {children.map((child) => (
          <DropdownMenuItem
            key={child.id}
            onClick={() => setCurrentChild(child)}
            className={cn(
              'gap-2',
              currentChild.id === child.id && 'bg-accent'
            )}
          >
            <Avatar className="h-6 w-6">
              <AvatarImage src={child.avatarUrl} alt={child.name} />
              <AvatarFallback className="text-xs">
                {child.name.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <span className="flex-1">{child.name}</span>
            <Badge variant="secondary" className="text-xs">
              {formatAge(child.dateOfBirth)}
            </Badge>
            {currentChild.id === child.id && (
              <Check className="h-4 w-4 text-primary" />
            )}
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <Plus className="mr-2 h-4 w-4" />
          Add child
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
