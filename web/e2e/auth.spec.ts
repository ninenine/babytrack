import { test, expect } from './fixtures'

test.describe('Authentication', () => {
  test('should show login page for unauthenticated users', async ({ page }) => {
    await page.goto('/')

    // Should redirect to login
    await expect(page).toHaveURL('/login')
  })

  test('should display login page with Google sign-in', async ({ page }) => {
    await page.goto('/login')

    // Check for login page elements
    await expect(page.locator('text=BabyTrack')).toBeVisible()
    await expect(page.locator('text=Continue with Google')).toBeVisible()
  })

  test('should redirect protected routes to login', async ({ page }) => {
    // Try to access protected route
    await page.goto('/feeding')

    // Should redirect to login
    await expect(page).toHaveURL('/login')
  })

  test('should redirect to onboarding if authenticated but no family', async ({
    page,
  }) => {
    // Mock authenticated state without family
    await page.addInitScript(() => {
      window.localStorage.setItem(
        'session-storage',
        JSON.stringify({
          state: {
            user: { id: 'user-1', email: 'test@example.com', name: 'Test' },
            token: 'test-token',
            isAuthenticated: true,
          },
          version: 0,
        })
      )
    })

    await page.goto('/')

    // Should redirect to onboarding
    await expect(page).toHaveURL('/onboarding')
  })
})
