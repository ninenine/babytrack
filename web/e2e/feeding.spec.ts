import { test, expect, mockAuthenticatedUser } from './fixtures'

test.describe('Feeding Page', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedUser(page)
  })

  test('should display feeding page header', async ({ page }) => {
    await page.goto('/feeding')

    await expect(page.locator('h1:has-text("Feeding")')).toBeVisible()
    await expect(page.locator('button:has-text("Add")')).toBeVisible()
  })

  test('should display filter tabs', async ({ page }) => {
    await page.goto('/feeding')

    await expect(page.locator('button:has-text("All")')).toBeVisible()
    await expect(page.locator('button:has-text("Breast")')).toBeVisible()
    await expect(page.locator('button:has-text("Bottle")')).toBeVisible()
    await expect(page.locator('button:has-text("Formula")')).toBeVisible()
    await expect(page.locator('button:has-text("Solid")')).toBeVisible()
  })

  test('should have All tab selected by default', async ({ page }) => {
    await page.goto('/feeding')

    await expect(
      page.locator('button[data-state="active"]:has-text("All")')
    ).toBeVisible()
  })

  test('should switch filter tabs', async ({ page }) => {
    await page.goto('/feeding')

    await page.click('button:has-text("Breast")')
    await expect(
      page.locator('button[data-state="active"]:has-text("Breast")')
    ).toBeVisible()

    await page.click('button:has-text("Bottle")')
    await expect(
      page.locator('button[data-state="active"]:has-text("Bottle")')
    ).toBeVisible()
  })

  test('should open add feeding dialog', async ({ page }) => {
    await page.goto('/feeding')

    await page.click('button:has-text("Add")')

    // Dialog should appear
    await expect(page.locator('[role="dialog"]')).toBeVisible()
  })

  test('should close dialog on cancel', async ({ page }) => {
    await page.goto('/feeding')

    await page.click('button:has-text("Add")')
    await expect(page.locator('[role="dialog"]')).toBeVisible()

    // Close the dialog (click outside or press escape)
    await page.keyboard.press('Escape')

    await expect(page.locator('[role="dialog"]')).not.toBeVisible()
  })
})
