import { test, expect, mockAuthenticatedUser } from './fixtures'

test.describe('Home Page', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedUser(page)
  })

  test('should display dashboard tab by default', async ({ page }) => {
    await page.goto('/')

    await expect(page.locator('button[data-state="active"]')).toHaveText(
      'Dashboard'
    )
    await expect(page.locator('text=Quick Actions')).toBeVisible()
    await expect(page.locator("text=Today's Summary")).toBeVisible()
  })

  test('should switch to timeline tab', async ({ page }) => {
    await page.goto('/')

    await page.click('button:has-text("Timeline")')

    await expect(
      page.locator('button[data-state="active"]:has-text("Timeline")')
    ).toBeVisible()
  })

  test('should display quick action buttons', async ({ page }) => {
    await page.goto('/')

    await expect(page.locator('text=Log Feeding')).toBeVisible()
    await expect(page.locator('text=Log Sleep')).toBeVisible()
    await expect(page.locator('text=Log Dose')).toBeVisible()
    await expect(page.locator('text=Add Note')).toBeVisible()
  })

  test('should display today summary section', async ({ page }) => {
    await page.goto('/')

    await expect(page.locator('text=Feedings')).toBeVisible()
    await expect(page.locator('text=Sleep')).toBeVisible()
    await expect(page.locator('text=Active Meds')).toBeVisible()
  })

  test('should display upcoming section', async ({ page }) => {
    await page.goto('/')

    // Scroll down and wait for upcoming section
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
    await expect(page.getByText('Upcoming', { exact: true })).toBeVisible()
  })

  test('should navigate to feeding page from quick action', async ({
    page,
  }) => {
    await page.goto('/')

    await page.click('text=Log Feeding')

    await expect(page).toHaveURL('/feeding')
  })

  test('should navigate to sleep page from quick action', async ({ page }) => {
    await page.goto('/')

    await page.click('text=Log Sleep')

    await expect(page).toHaveURL('/sleep')
  })
})
