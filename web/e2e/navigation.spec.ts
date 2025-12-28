import { test, expect, mockAuthenticatedUser } from './fixtures'

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedUser(page)
  })

  test('should show home page for authenticated users', async ({ page }) => {
    await page.goto('/')

    // Should stay on home page
    await expect(page).toHaveURL('/')
  })

  test('should navigate to feeding page', async ({ page }) => {
    await page.goto('/')

    // Click on feeding link in navigation
    await page.click('a[href="/feeding"]')

    await expect(page).toHaveURL('/feeding')
  })

  test('should navigate to sleep page', async ({ page }) => {
    await page.goto('/')

    await page.click('a[href="/sleep"]')

    await expect(page).toHaveURL('/sleep')
  })

  test('should navigate to settings page', async ({ page }) => {
    await page.goto('/')

    // Settings is in the user avatar dropdown menu
    await page.click('button:has(.rounded-full)') // Avatar button
    await page.click('text=Settings')

    await expect(page).toHaveURL('/settings')
  })

  test('should show current child name', async ({ page }) => {
    await page.goto('/')

    // The child name should be visible somewhere
    await expect(page.locator('text=Emma')).toBeVisible()
  })
})
