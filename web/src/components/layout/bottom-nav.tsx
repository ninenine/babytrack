import { NavLink } from 'react-router-dom'
import { Home, Utensils, Moon, Heart, MoreHorizontal } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useState } from 'react'

const mainTabs = [
  { to: '/', icon: Home, label: 'Home' },
  { to: '/feeding', icon: Utensils, label: 'Feeding' },
  { to: '/sleep', icon: Moon, label: 'Sleep' },
  { to: '/health', icon: Heart, label: 'Health' },
]

const moreLinks = [
  { to: '/medications', label: 'Medications' },
  { to: '/vaccinations', label: 'Vaccinations' },
  { to: '/appointments', label: 'Appointments' },
  { to: '/notes', label: 'Notes' },
  { to: '/settings', label: 'Settings' },
]

export function BottomNav() {
  const [sheetOpen, setSheetOpen] = useState(false)

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t bg-background safe-area-pb">
      <div className="flex items-center justify-around h-16">
        {mainTabs.map((tab) => (
          <NavLink
            key={tab.to}
            to={tab.to}
            end={tab.to === '/'}
            className={({ isActive }) =>
              cn(
                'flex flex-col items-center justify-center flex-1 h-full gap-1 text-xs transition-colors',
                isActive
                  ? 'text-primary'
                  : 'text-muted-foreground hover:text-foreground'
              )
            }
          >
            {({ isActive }) => (
              <>
                <tab.icon className={cn('h-5 w-5', isActive && 'fill-primary/20')} />
                <span>{tab.label}</span>
              </>
            )}
          </NavLink>
        ))}

        {/* More Sheet */}
        <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
          <SheetTrigger asChild>
            <button className="flex flex-col items-center justify-center flex-1 h-full gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors">
              <MoreHorizontal className="h-5 w-5" />
              <span>More</span>
            </button>
          </SheetTrigger>
          <SheetContent side="bottom" className="h-auto px-4 sm:px-6">
            <SheetHeader>
              <SheetTitle>More Options</SheetTitle>
            </SheetHeader>
            <div className="grid grid-cols-2 gap-2 py-4">
              {moreLinks.map((link) => (
                <Button
                  key={link.to}
                  variant="outline"
                  className="h-14 justify-start"
                  asChild
                  onClick={() => setSheetOpen(false)}
                >
                  <NavLink to={link.to}>{link.label}</NavLink>
                </Button>
              ))}
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </nav>
  )
}

