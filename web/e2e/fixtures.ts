import { test as base, expect } from '@playwright/test'

export const test = base.extend({
  // Auto-clear localStorage and mock API routes before each test
  page: async ({ page }, use) => {
    await page.addInitScript(() => {
      window.localStorage.clear()
    })

    // Mock all API routes to prevent ECONNREFUSED errors
    await page.route('**/api/**', (route) => {
      const url = route.request().url()

      // Return appropriate empty responses based on endpoint
      if (url.includes('/auth/')) {
        return route.fulfill({ status: 200, json: { user: null, token: null } })
      }
      if (url.includes('/sync/')) {
        return route.fulfill({ status: 200, json: { events: [], synced: 0 } })
      }
      if (url.includes('/version')) {
        return route.fulfill({ status: 200, json: { version: 'test' } })
      }
      // Default: return empty array for list endpoints
      return route.fulfill({ status: 200, json: [] })
    })

    await use(page) // eslint-disable-line react-hooks/rules-of-hooks
  },
})

export { expect }

// Mock data for E2E tests
export const mockUser = {
  id: 'user-1',
  email: 'test@example.com',
  name: 'Test User',
}

export const mockFamily = {
  id: 'family-1',
  name: 'Test Family',
  children: [
    {
      id: 'child-1',
      name: 'Emma',
      dateOfBirth: '2023-01-15',
      gender: 'female',
    },
  ],
}

// Helper to mock authenticated state
export async function mockAuthenticatedUser(page: typeof base.prototype.page) {
  await page.addInitScript(
    ({ user, family }) => {
      // Set up session store
      window.localStorage.setItem(
        'session-storage',
        JSON.stringify({
          state: {
            user,
            token: 'test-token',
            isAuthenticated: true,
          },
          version: 0,
        })
      )
      // Set up family store
      window.localStorage.setItem(
        'family-storage',
        JSON.stringify({
          state: {
            currentFamily: family,
            currentChild: family.children[0],
            families: [family],
          },
          version: 0,
        })
      )
    },
    { user: mockUser, family: mockFamily }
  )
}
