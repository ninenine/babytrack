import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { DashboardTab } from '@/components/home/DashboardTab'
import { TimelineTab } from '@/components/home/TimelineTab'

export function HomePage() {
  const [activeTab, setActiveTab] = useState('dashboard')

  return (
    <div className="space-y-4">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="dashboard">Dashboard</TabsTrigger>
          <TabsTrigger value="timeline">Timeline</TabsTrigger>
        </TabsList>
        <TabsContent value="dashboard" className="mt-4">
          <DashboardTab />
        </TabsContent>
        <TabsContent value="timeline" className="mt-4">
          <TimelineTab />
        </TabsContent>
      </Tabs>
    </div>
  )
}
