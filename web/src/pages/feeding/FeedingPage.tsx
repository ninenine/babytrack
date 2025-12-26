import { useState } from 'react'
import { Plus, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { FeedingForm, FeedingList } from '@/components/feeding'
import { useFeedings, type FeedingType } from '@/hooks'
import { useFamilyStore } from '@/stores/family.store'
import type { LocalFeeding } from '@/db/dexie'

export function FeedingPage() {
  const currentChild = useFamilyStore((state) => state.currentChild)
  const { feedings, isLoading, isSyncing } = useFeedings()
  const [showForm, setShowForm] = useState(false)
  const [editingFeeding, setEditingFeeding] = useState<LocalFeeding | null>(null)
  const [filter, setFilter] = useState<FeedingType | 'all'>('all')

  if (!currentChild) {
    return (
      <div className="flex items-center justify-center h-[50vh] text-muted-foreground">
        No child selected
      </div>
    )
  }

  const filteredFeedings = filter === 'all'
    ? feedings
    : feedings.filter((f) => f.type === filter)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h1 className="text-2xl font-bold">Feeding</h1>
          {isSyncing && (
            <RefreshCw className="h-4 w-4 animate-spin text-muted-foreground" />
          )}
        </div>
        <Button size="sm" onClick={() => setShowForm(true)}>
          <Plus className="h-4 w-4 mr-1" />
          Add
        </Button>
      </div>

      <Tabs value={filter} onValueChange={(v) => setFilter(v as FeedingType | 'all')}>
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="all">All</TabsTrigger>
          <TabsTrigger value="breast">Breast</TabsTrigger>
          <TabsTrigger value="bottle">Bottle</TabsTrigger>
          <TabsTrigger value="formula">Formula</TabsTrigger>
          <TabsTrigger value="solid">Solid</TabsTrigger>
        </TabsList>
      </Tabs>

      <FeedingList
        feedings={filteredFeedings}
        isLoading={isLoading}
        onEdit={(feeding) => {
          setEditingFeeding(feeding)
          setShowForm(true)
        }}
      />

      <FeedingForm
        open={showForm}
        onOpenChange={(open) => {
          setShowForm(open)
          if (!open) setEditingFeeding(null)
        }}
        feeding={editingFeeding}
      />
    </div>
  )
}
