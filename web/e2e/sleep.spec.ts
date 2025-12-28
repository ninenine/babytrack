import { test, expect, mockAuthenticatedUser } from './fixtures'

test.describe('Sleep Page', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedUser(page)
  })

  test('should display sleep page header', async ({ page }) => {
    await page.goto('/sleep')

    await expect(page.locator('h1:has-text("Sleep")')).toBeVisible()
  })

  test('should display start sleep buttons when no active sleep', async ({
    page,
  }) => {
    await page.goto('/sleep')

    // Should show buttons to start different types of sleep
    await expect(page.locator('text=Nap')).toBeVisible()
    await expect(page.locator('text=Night')).toBeVisible()
  })

  test('should display history section', async ({ page }) => {
    await page.goto('/sleep')

    await expect(page.locator('h2:has-text("History")')).toBeVisible()
  })

  test('should be accessible from navigation', async ({ page }) => {
    await page.goto('/')

    await page.click('a[href="/sleep"]')

    await expect(page).toHaveURL('/sleep')
    await expect(page.locator('h1:has-text("Sleep")')).toBeVisible()
  })
})
