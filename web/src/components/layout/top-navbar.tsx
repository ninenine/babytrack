import { Link, NavLink } from 'react-router-dom'
import { Baby, LogOut, Settings, User } from 'lucide-react'
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
import { ChildSwitcher } from './child-switcher'
import { SyncIndicator } from './sync-indicator'
import { useMobile } from '@/hooks'
import { useSessionStore } from '@/stores/session.store'
import { cn } from '@/lib/utils'

const navLinks = [
  { to: '/', label: 'Home' },
  { to: '/feeding', label: 'Feeding' },
  { to: '/sleep', label: 'Sleep' },
  { to: '/medications', label: 'Medications' },
  { to: '/vaccinations', label: 'Vaccinations' },
  { to: '/appointments', label: 'Appointments' },
  { to: '/notes', label: 'Notes' },
]

export function TopNavbar() {
  const isMobile = useMobile()
  const { user, clearSession } = useSessionStore()

  const handleLogout = () => {
    clearSession()
    window.location.href = '/login'
  }

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container max-w-4xl mx-auto flex h-14 items-center px-4">
        {/* Logo */}
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <Baby className="h-6 w-6 text-primary" />
          <span className="hidden sm:inline">BabyTrack</span>
        </Link>

        {/* Desktop Navigation */}
        {!isMobile && (
          <nav className="ml-6 flex items-center gap-1">
            {navLinks.slice(0, 4).map((link) => (
              <NavLink
                key={link.to}
                to={link.to}
                className={({ isActive }) =>
                  cn(
                    'px-3 py-2 text-sm font-medium rounded-md transition-colors',
                    isActive
                      ? 'bg-primary/10 text-primary'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                  )
                }
              >
                {link.label}
              </NavLink>
            ))}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="text-muted-foreground">
                  More
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start">
                {navLinks.slice(4).map((link) => (
                  <DropdownMenuItem key={link.to} asChild>
                    <Link to={link.to}>{link.label}</Link>
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </nav>
        )}

        {/* Right Side */}
        <div className="ml-auto flex items-center gap-2">
          {/* Child Switcher */}
          <ChildSwitcher />

          {/* Sync Indicator */}
          <SyncIndicator />

          {/* User Menu */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="rounded-full">
                <Avatar className="h-8 w-8">
                  <AvatarImage src={user?.avatarUrl} alt={user?.name} />
                  <AvatarFallback>
                    {user?.name?.charAt(0)?.toUpperCase() || <User className="h-4 w-4" />}
                  </AvatarFallback>
                </Avatar>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel>
                <div className="flex flex-col">
                  <span>{user?.name}</span>
                  <span className="text-xs font-normal text-muted-foreground">
                    {user?.email}
                  </span>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild>
                <Link to="/settings">
                  <Settings className="mr-2 h-4 w-4" />
                  Settings
                </Link>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleLogout} className="text-destructive">
                <LogOut className="mr-2 h-4 w-4" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
