import { test, expect, mockAuthenticatedUser } from './fixtures'

test.describe('Notes Page', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedUser(page)
  })

  test('should display notes page header', async ({ page }) => {
    await page.goto('/notes')

    await expect(page.locator('h1:has-text("Notes")')).toBeVisible()
    await expect(page.locator('button:has-text("Add")')).toBeVisible()
  })

  test('should display search input', async ({ page }) => {
    await page.goto('/notes')

    await expect(page.locator('input[placeholder="Search notes..."]')).toBeVisible()
  })

  test('should show empty state when no notes', async ({ page }) => {
    await page.goto('/notes')

    await expect(page.locator('text=No notes yet')).toBeVisible()
  })

  test('should open add note dialog', async ({ page }) => {
    await page.goto('/notes')

    await page.click('button:has-text("Add")')

    await expect(page.locator('[role="dialog"]')).toBeVisible()
  })

  test('should allow typing in search', async ({ page }) => {
    await page.goto('/notes')

    const searchInput = page.locator('input[placeholder="Search notes..."]')
    await searchInput.fill('test search')

    await expect(searchInput).toHaveValue('test search')
  })

  test('should be accessible from More dropdown', async ({ page }) => {
    await page.goto('/')

    await page.click('text=More')
    await page.getByRole('menuitem', { name: 'Notes' }).click()

    await expect(page).toHaveURL('/notes')
    await expect(page.locator('h1:has-text("Notes")')).toBeVisible()
  })
})
