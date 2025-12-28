import { test, expect, mockAuthenticatedUser, mockAuthenticatedUserWithTwoChildren, mockUser, mockFamily } from './fixtures'

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

test.describe('Child Management', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedUser(page)
  })

  test('should display children card with existing child', async ({ page }) => {
    await page.goto('/settings')

    // Wait for page to load
    await page.waitForLoadState('networkidle')

    // Check that the children section is visible
    await expect(page.getByText('Manage children in your family')).toBeVisible()
    // Check that the mock child is displayed
    await expect(page.getByText('Emma').first()).toBeVisible()
  })

  test('should display add child button', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // Find the Add button in the Children card
    const addButton = page.locator('button').filter({ hasText: 'Add' }).first()
    await expect(addButton).toBeVisible()
  })

  test('should open add child dialog', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // Click the Add button in the Children card
    await page.locator('button').filter({ hasText: 'Add' }).first().click()

    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(page.getByText('Add a new child to your family')).toBeVisible()
  })

  test('should close add child dialog on cancel', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.locator('button').filter({ hasText: 'Add' }).first().click()
    await expect(page.getByRole('dialog')).toBeVisible()

    await page.getByRole('button', { name: 'Cancel' }).click()

    await expect(page.getByRole('dialog')).not.toBeVisible()
  })

  test('should add a new child', async ({ page }) => {
    // Mock the POST endpoint for adding a child
    await page.route('**/api/families/*/children', (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({
          status: 201,
          json: {
            id: 'child-2',
            name: 'Oliver',
            date_of_birth: '2024-06-01T00:00:00Z',
            gender: 'male',
          },
        })
      }
      return route.fulfill({ status: 200, json: [] })
    })

    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.locator('button').filter({ hasText: 'Add' }).first().click()

    // Fill in the form
    await page.locator('input[name="name"]').fill('Oliver')
    // Click on date picker to open calendar
    await page.getByText('Select date & time of birth').click()
    // Wait for calendar popover and click on the 15th day of the current month (guaranteed to be enabled)
    await page.waitForSelector('[role="grid"]')
    await page.getByRole('button', { name: /15th,/ }).click()
    // Close the popover by pressing Escape
    await page.keyboard.press('Escape')
    await expect(page.locator('[role="grid"]')).not.toBeVisible()

    await page.getByRole('button', { name: 'Add Child' }).click()

    // Should show success toast
    await expect(page.getByText('Child added')).toBeVisible()
  })

  test('should open edit child dialog', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // Click the edit button (pencil icon) for Emma
    const editButton = page.locator('button').filter({ has: page.locator('svg.lucide-pencil') }).first()
    await editButton.click()

    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(page.getByText("Update your child's information")).toBeVisible()
  })

  test('should update an existing child', async ({ page }) => {
    // Mock the PUT endpoint for updating a child
    await page.route('**/api/families/*/children/*', (route) => {
      if (route.request().method() === 'PUT') {
        return route.fulfill({
          status: 200,
          json: {
            id: 'child-1',
            name: 'Emma Rose',
            date_of_birth: '2023-01-15T00:00:00Z',
            gender: 'female',
          },
        })
      }
      return route.fulfill({ status: 200, json: {} })
    })

    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // Click the edit button
    const editButton = page.locator('button').filter({ has: page.locator('svg.lucide-pencil') }).first()
    await editButton.click()

    // Update the name
    await page.locator('input[name="name"]').fill('Emma Rose')
    await page.getByRole('button', { name: 'Save Changes' }).click()

    // Should show success toast
    await expect(page.getByText('Child updated')).toBeVisible()
  })

  test('should show delete confirmation dialog', async ({ page }) => {
    // Need two children so delete button is enabled
    await mockAuthenticatedUserWithTwoChildren(page)
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // Click the delete button (trash icon) - should be enabled with two children
    const deleteButton = page.locator('button').filter({ has: page.locator('svg.lucide-trash-2') }).first()
    await deleteButton.click()

    await expect(page.getByRole('alertdialog')).toBeVisible()
    await expect(page.getByText('Are you sure you want to remove Emma?')).toBeVisible()
  })

  test('should close delete dialog on cancel', async ({ page }) => {
    // Need two children so delete button is enabled
    await mockAuthenticatedUserWithTwoChildren(page)
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    const deleteButton = page.locator('button').filter({ has: page.locator('svg.lucide-trash-2') }).first()
    await deleteButton.click()
    await expect(page.getByRole('alertdialog')).toBeVisible()

    await page.getByRole('button', { name: 'Cancel' }).click()

    await expect(page.getByRole('alertdialog')).not.toBeVisible()
  })

  test('should delete a child', async ({ page }) => {
    // Need two children so delete button is enabled
    await mockAuthenticatedUserWithTwoChildren(page)

    // Mock the DELETE endpoint
    await page.route('**/api/families/*/children/*', (route) => {
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.fulfill({ status: 200, json: {} })
    })

    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // Click the delete button for the second child (Oliver)
    const deleteButton = page.locator('button').filter({ has: page.locator('svg.lucide-trash-2') }).nth(1)
    await deleteButton.click()
    await expect(page.getByRole('alertdialog')).toBeVisible()

    await page.getByRole('button', { name: 'Remove' }).click()

    // Should show success toast
    await expect(page.getByText('Child removed')).toBeVisible()
  })

  test('should disable delete button when only one child', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // The delete button should be disabled when there's only one child
    const deleteButton = page.locator('button').filter({ has: page.locator('svg.lucide-trash-2') }).first()
    await expect(deleteButton).toBeDisabled()
  })
})

test.describe('Family Management', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticatedUser(page)
  })

  test('should display family card with current family', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // Check that the family section is visible
    await expect(page.getByText('Manage your family settings')).toBeVisible()
    // Check that the mock family name is displayed
    await expect(page.getByText(mockFamily.name)).toBeVisible()
  })

  test('should display rename button', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await expect(page.getByRole('button', { name: 'Rename' })).toBeVisible()
  })

  test('should display leave and delete family buttons', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await expect(page.getByRole('button', { name: 'Leave Family' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Delete Family' })).toBeVisible()
  })

  test('should open rename family dialog', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Rename' }).click()

    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(page.getByText('Rename Family')).toBeVisible()
    await expect(page.getByText('Enter a new name for your family')).toBeVisible()
  })

  test('should close rename dialog on cancel', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Rename' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()

    await page.getByRole('button', { name: 'Cancel' }).click()

    await expect(page.getByRole('dialog')).not.toBeVisible()
  })

  test('should rename family', async ({ page }) => {
    // Mock the PUT endpoint for renaming family
    await page.route('**/api/families/*', (route) => {
      if (route.request().method() === 'PUT') {
        return route.fulfill({
          status: 200,
          json: {
            id: 'family-1',
            name: 'Updated Family Name',
          },
        })
      }
      return route.fallback()
    })

    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Rename' }).click()

    // Update the family name
    await page.locator('input#familyName').fill('Updated Family Name')
    await page.getByRole('button', { name: 'Save' }).click()

    // Should show success toast
    await expect(page.getByText('Family name updated')).toBeVisible()
  })

  test('should show leave family confirmation dialog', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Leave Family' }).click()

    await expect(page.getByRole('alertdialog')).toBeVisible()
    await expect(page.getByText('Are you sure you want to leave Test Family?')).toBeVisible()
  })

  test('should close leave family dialog on cancel', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Leave Family' }).click()
    await expect(page.getByRole('alertdialog')).toBeVisible()

    await page.getByRole('button', { name: 'Cancel' }).click()

    await expect(page.getByRole('alertdialog')).not.toBeVisible()
  })

  test('should leave family', async ({ page }) => {
    // Mock the POST endpoint for leaving family
    await page.route('**/api/families/*/leave', (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ status: 204 })
      }
      return route.fallback()
    })

    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Leave Family' }).click()
    await expect(page.getByRole('alertdialog')).toBeVisible()

    await page.getByRole('button', { name: 'Leave' }).click()

    // After leaving the only family, should redirect to onboarding
    await page.waitForURL('**/onboarding')
  })

  test('should show delete family confirmation dialog', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Delete Family' }).click()

    await expect(page.getByRole('alertdialog')).toBeVisible()
    await expect(page.getByText('Are you sure you want to delete Test Family?')).toBeVisible()
    await expect(page.getByText('This action cannot be undone')).toBeVisible()
  })

  test('should close delete family dialog on cancel', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Delete Family' }).click()
    await expect(page.getByRole('alertdialog')).toBeVisible()

    await page.getByRole('button', { name: 'Cancel' }).click()

    await expect(page.getByRole('alertdialog')).not.toBeVisible()
  })

  test('should delete family', async ({ page }) => {
    // Mock the DELETE endpoint for deleting family
    await page.route('**/api/families/*', (route) => {
      if (route.request().method() === 'DELETE') {
        return route.fulfill({ status: 204 })
      }
      return route.fallback()
    })

    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Delete Family' }).click()
    await expect(page.getByRole('alertdialog')).toBeVisible()

    // Click Delete button in the confirmation dialog
    await page.getByRole('button', { name: 'Delete', exact: true }).click()

    // After deleting the only family, should redirect to onboarding
    await page.waitForURL('**/onboarding')
  })

  test('should show error when leaving as only admin', async ({ page }) => {
    // Mock the POST endpoint to return error
    await page.route('**/api/families/*/leave', (route) => {
      if (route.request().method() === 'POST') {
        return route.fulfill({
          status: 400,
          json: { error: 'cannot leave: you are the only admin' },
        })
      }
      return route.fallback()
    })

    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Leave Family' }).click()
    await page.getByRole('button', { name: 'Leave' }).click()

    // Should show error toast
    await expect(page.getByText('cannot leave: you are the only admin')).toBeVisible()
  })

  test('should show error when deleting as non-admin', async ({ page }) => {
    // Mock the DELETE endpoint to return forbidden error
    await page.route('**/api/families/*', (route) => {
      if (route.request().method() === 'DELETE') {
        return route.fulfill({
          status: 403,
          json: { error: 'only admins can delete a family' },
        })
      }
      return route.fallback()
    })

    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    await page.getByRole('button', { name: 'Delete Family' }).click()
    await page.getByRole('button', { name: 'Delete', exact: true }).click()

    // Should show error toast
    await expect(page.getByText('only admins can delete a family')).toBeVisible()
  })
})
