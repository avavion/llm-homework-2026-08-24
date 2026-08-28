import { expect, test } from '@playwright/test'

test('sign-in keeps labels and the story panel on desktop', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 900 })
  await page.goto('/login')
  await expect(page.getByTestId('auth-story')).toBeVisible()
  await expect(page.getByLabel(/e-mail|email/i)).toBeVisible()
  await expect(page.getByLabel(/пароль|password/i)).toBeVisible()
  expect(await page.locator('html').evaluate((node) => node.scrollWidth <= node.clientWidth)).toBe(true)
})
