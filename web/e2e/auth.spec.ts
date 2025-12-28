import { test, expect } from './fixtures'

test.describe('Authentication', () => {
  test('should show login page for unauthenticated users', async ({ page }) => {
    await page.goto('/')

    // Should redirect to login
    await expect(page).toHaveURL('/login')
  })

  test('should display login page with branding', async ({ page }) => {
    await page.goto('/login')

    await expect(page.locator('text=BabyTrack')).toBeVisible()
    await expect(
      page.locator("text=Track your baby's feeding, sleep, and more")
    ).toBeVisible()
  })

  test('should display Google sign-in button', async ({ page }) => {
    await page.goto('/login')

    await expect(page.locator('text=Continue with Google')).toBeVisible()
  })

  test('should redirect protected routes to login', async ({ page }) => {
    const protectedRoutes = [
      '/feeding',
      '/sleep',
      '/medications',
      '/notes',
      '/settings',
    ]

    for (const route of protectedRoutes) {
      await page.goto(route)
      await expect(page).toHaveURL('/login')
    }
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

  test('should display onboarding page for new users', async ({ page }) => {
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

    await page.goto('/onboarding')

    // Should show onboarding content
    await expect(page).toHaveURL('/onboarding')
  })

  test('should preserve login page on direct access when not authenticated', async ({
    page,
  }) => {
    await page.goto('/login')

    // Should stay on login page
    await expect(page).toHaveURL('/login')
    await expect(page.locator('text=Continue with Google')).toBeVisible()
  })
})
