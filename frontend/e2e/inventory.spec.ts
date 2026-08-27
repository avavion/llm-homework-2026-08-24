import { expect, test } from '@playwright/test'

for (const viewport of [
  { name: 'mobile', width: 320, height: 740 },
  { name: 'tablet', width: 768, height: 900 },
  { name: 'desktop', width: 1440, height: 960 },
]) {
  test(`inventory remains usable on ${viewport.name}`, async ({ page }) => {
    await page.setViewportSize(viewport)
    await page.goto('/login')
    await page.getByLabel('E-mail').fill('playwright@example.com')
    await page.getByLabel('Пароль').fill('password123')
    await page.getByRole('button', { name: /войти|sign in/i }).click()
    await expect(page).toHaveURL(/\/products$/)
    await expect(page.getByRole('link', { name: /добавить продукт|add product/i })).toBeVisible()
    await expect(page.getByRole('navigation')).toBeVisible()
    expect(await page.locator('html').evaluate((element) => element.scrollWidth <= element.clientWidth)).toBe(true)
  })
}

test('dark theme applies without navigation', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('E-mail').fill('playwright@example.com')
  await page.getByLabel('Пароль').fill('password123')
  await page.getByRole('button', { name: /войти|sign in/i }).click()
  await expect(page).toHaveURL(/\/products$/)
  await page.goto('/settings')
  await page.getByRole('button', { name: /тёмная|dark/i }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')
})
