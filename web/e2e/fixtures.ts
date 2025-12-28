import { test as base, expect } from '@playwright/test'

export const test = base.extend({
  // Auto-clear localStorage before each test
  page: async ({ page }, use) => {
    await page.addInitScript(() => {
      window.localStorage.clear()
    })
    await use(page)
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
