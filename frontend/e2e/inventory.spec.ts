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
    await expect(page).toHaveURL(/\/$/)
    await expect(page.getByRole('button', { name: /добавить продукт|add product/i })).toBeVisible()
    await expect(page.getByRole('navigation')).toBeVisible()
    expect(await page.locator('html').evaluate((element) => element.scrollWidth <= element.clientWidth)).toBe(true)
  })
}

test('mobile add-product sheet opens, is keyboard-dismissable, and returns focus', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 740 })
  await page.goto('/login')
  await page.getByLabel('E-mail').fill('playwright@example.com')
  await page.getByLabel('Пароль').fill('password123')
  await page.getByRole('button', { name: /войти|sign in/i }).click()
  await expect(page).toHaveURL(/\/$/)

  const trigger = page.getByRole('button', { name: /добавить продукт|add product/i })
  await trigger.click()
  const dialog = page.getByRole('dialog', { name: /добавить продукт|add product/i })
  await expect(dialog).toBeVisible()
  await expect(dialog.getByRole('link', { name: /ввести вручную|enter manually/i })).toBeVisible()
  await expect(dialog.getByRole('link', { name: /заполнить по фото|fill from photo/i })).toBeVisible()

  await page.keyboard.press('Escape')
  await expect(dialog).toBeHidden()
  await expect(trigger).toBeFocused()
})

test('reduced motion removes the sheet transition', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await page.setViewportSize({ width: 375, height: 740 })
  await page.goto('/login')
  await page.getByLabel('E-mail').fill('playwright@example.com')
  await page.getByLabel('Пароль').fill('password123')
  await page.getByRole('button', { name: /войти|sign in/i }).click()
  await expect(page).toHaveURL(/\/$/)

  await page.getByRole('button', { name: /добавить продукт|add product/i }).click()
  const sheet = page.locator('.add-sheet')
  await expect(sheet).toBeVisible()
  const animationName = await sheet.evaluate((element) => getComputedStyle(element).animationName)
  expect(animationName).toBe('none')
})

test('shows the server-provided status text for active, attention and expired fixtures', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('E-mail').fill('playwright@example.com')
  await page.getByLabel('Пароль').fill('password123')
  await page.getByRole('button', { name: /войти|sign in/i }).click()
  await expect(page).toHaveURL(/\/$/)

  const table = page.locator('#inventory-table')
  await expect(table.getByText('Активен', { exact: true })).toBeVisible()
  await expect(table.getByText('Требует внимания', { exact: true })).toBeVisible()
  await expect(table.getByText('Срок истёк', { exact: true })).toBeVisible()
})
