import { http, HttpResponse } from 'msw'
import { API_ENDPOINTS } from '@/lib/constants'

// Mock data
export const mockUser = {
  id: 'user-1',
  email: 'test@example.com',
  name: 'Test User',
  avatar_url: 'https://example.com/avatar.jpg',
}

export const mockFamily = {
  id: 'family-1',
  name: 'Test Family',
  children: [
    {
      id: 'child-1',
      name: 'Emma',
      date_of_birth: '2023-01-15',
      gender: 'female',
    },
  ],
}

export const mockFeeding = {
  id: 'feeding-1',
  child_id: 'child-1',
  type: 'bottle',
  amount: 120,
  unit: 'ml',
  start_time: '2024-03-15T10:30:00Z',
  end_time: '2024-03-15T10:45:00Z',
  notes: 'Fed well',
}

export const mockSleep = {
  id: 'sleep-1',
  child_id: 'child-1',
  type: 'nap',
  start_time: '2024-03-15T13:00:00Z',
  end_time: '2024-03-15T14:30:00Z',
}

// API Handlers
export const handlers = [
  // Auth
  http.get(API_ENDPOINTS.AUTH.ME, () => {
    return HttpResponse.json(mockUser)
  }),

  http.post(API_ENDPOINTS.AUTH.REFRESH, () => {
    return HttpResponse.json({
      token: 'new-test-token',
      user: mockUser,
    })
  }),

  // Families
  http.get(API_ENDPOINTS.FAMILIES.BASE, () => {
    return HttpResponse.json([mockFamily])
  }),

  http.get(API_ENDPOINTS.FAMILIES.BY_ID(':id'), ({ params }) => {
    if (params.id === mockFamily.id) {
      return HttpResponse.json(mockFamily)
    }
    return new HttpResponse(null, { status: 404 })
  }),

  http.get(API_ENDPOINTS.FAMILIES.CHILDREN(':familyId'), () => {
    return HttpResponse.json(mockFamily.children)
  }),

  // Feedings
  http.get(API_ENDPOINTS.FEEDINGS.BASE, ({ request }) => {
    const url = new URL(request.url)
    const childId = url.searchParams.get('child_id')
    if (childId === 'child-1') {
      return HttpResponse.json([mockFeeding])
    }
    return HttpResponse.json([])
  }),

  http.post(API_ENDPOINTS.FEEDINGS.BASE, async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json(
      { id: 'feeding-new', ...body },
      { status: 201 }
    )
  }),

  http.get(API_ENDPOINTS.FEEDINGS.BY_ID(':id'), ({ params }) => {
    if (params.id === mockFeeding.id) {
      return HttpResponse.json(mockFeeding)
    }
    return new HttpResponse(null, { status: 404 })
  }),

  http.put(API_ENDPOINTS.FEEDINGS.BY_ID(':id'), async ({ params, request }) => {
    const body = await request.json()
    return HttpResponse.json({ id: params.id, ...body })
  }),

  http.delete(API_ENDPOINTS.FEEDINGS.BY_ID(':id'), () => {
    return new HttpResponse(null, { status: 204 })
  }),

  // Sleep
  http.get(API_ENDPOINTS.SLEEP.BASE, ({ request }) => {
    const url = new URL(request.url)
    const childId = url.searchParams.get('child_id')
    if (childId === 'child-1') {
      return HttpResponse.json([mockSleep])
    }
    return HttpResponse.json([])
  }),

  http.post(API_ENDPOINTS.SLEEP.START, async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json(
      { id: 'sleep-new', ...body },
      { status: 201 }
    )
  }),

  http.post(API_ENDPOINTS.SLEEP.END(':id'), async ({ params, request }) => {
    const body = await request.json()
    return HttpResponse.json({ id: params.id, ...mockSleep, ...body })
  }),

  // Sync
  http.post(API_ENDPOINTS.SYNC.PUSH, async ({ request }) => {
    const body = await request.json() as { events: unknown[] }
    return HttpResponse.json({
      synced: body.events?.length || 0,
      failed: 0,
    })
  }),

  http.get(API_ENDPOINTS.SYNC.PULL, () => {
    return HttpResponse.json({ events: [] })
  }),

  http.get(API_ENDPOINTS.SYNC.STATUS, () => {
    return HttpResponse.json({
      last_sync: new Date().toISOString(),
      pending_count: 0,
    })
  }),
]

// Error handlers for testing error scenarios
export const errorHandlers = {
  unauthorized: http.get('*', () => {
    return new HttpResponse(null, { status: 401 })
  }),

  serverError: http.get('*', () => {
    return HttpResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }),

  networkError: http.get('*', () => {
    return HttpResponse.error()
  }),
}
