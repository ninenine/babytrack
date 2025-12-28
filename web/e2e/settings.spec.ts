import { test, expect, mockAuthenticatedUser, mockUser } from './fixtures'

test.describe('Settings Page', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedUser(page)
  })

  test('should display settings page header', async ({ page }) => {
    await page.goto('/settings')

    await expect(page.locator('h1:has-text("Settings")')).toBeVisible()
  })

  test('should display account card with user info', async ({ page }) => {
    await page.goto('/settings')

    // Scroll down to find account section
    await page.evaluate(() => window.scrollTo(0, 500))
    await expect(page.getByText('Account', { exact: true }).first()).toBeVisible()
    // User info from mock should be displayed
    await expect(page.getByText(mockUser.name)).toBeVisible()
  })

  test('should display preferences card', async ({ page }) => {
    await page.goto('/settings')

    await expect(page.locator('text=Preferences')).toBeVisible()
    await expect(page.locator('text=Dark Mode')).toBeVisible()
    await expect(page.locator('text=Push Notifications')).toBeVisible()
  })

  test('should display data card', async ({ page }) => {
    await page.goto('/settings')

    // Scroll to data section
    await page.locator('text=Clear Local Data').scrollIntoViewIfNeeded()
    await expect(page.locator('text=Clear Local Data')).toBeVisible()
    await expect(page.locator('button:has-text("Clear & Sync")')).toBeVisible()
  })

  test('should display about card', async ({ page }) => {
    await page.goto('/settings')

    // Scroll to about section - version is dynamic from API (mocked as 'test')
    await page.locator('text=BabyTrack test').scrollIntoViewIfNeeded()
    await expect(page.locator('text=BabyTrack test')).toBeVisible()
  })

  test('should have dark mode toggle', async ({ page }) => {
    await page.goto('/settings')

    const darkModeSwitch = page.locator('#dark-mode')
    await expect(darkModeSwitch).toBeVisible()
  })

  test('should display logout button', async ({ page }) => {
    await page.goto('/settings')

    await expect(page.locator('button:has-text("Log out")')).toBeVisible()
  })

  test('should open clear data confirmation dialog', async ({ page }) => {
    await page.goto('/settings')

    await page.click('button:has-text("Clear & Sync")')

    await expect(page.locator('text=Clear local data?')).toBeVisible()
    await expect(page.locator('button:has-text("Cancel")')).toBeVisible()
  })

  test('should close confirmation dialog on cancel', async ({ page }) => {
    await page.goto('/settings')

    await page.click('button:has-text("Clear & Sync")')
    await expect(page.locator('text=Clear local data?')).toBeVisible()

    await page.click('button:has-text("Cancel")')

    await expect(page.locator('text=Clear local data?')).not.toBeVisible()
  })
})
