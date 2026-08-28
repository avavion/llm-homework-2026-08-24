# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: inventory.spec.ts >> HTML-like product names render as text instead of executable markup
- Location: e2e/inventory.spec.ts:78:1

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: locator.fill: Test timeout of 30000ms exceeded.
Call log:
  - waiting for getByLabel('E-mail')

```

# Page snapshot

```yaml
- main [ref=e3]:
  - generic [ref=e4]:
    - generic [ref=e5]:
      - generic [ref=e6]: Welcome back
      - heading "Sign in" [level=1] [ref=e7]
      - paragraph [ref=e8]: Sign in to see your products and upcoming dates.
    - generic [ref=e9]:
      - generic [ref=e10]:
        - text: Email
        - textbox "Email" [ref=e11]
      - generic [ref=e12]:
        - text: Password
        - textbox "Password" [ref=e13]
      - button "Sign in" [ref=e14] [cursor=pointer]
      - link "Create account" [ref=e15] [cursor=pointer]:
        - /url: /register
```

# Test source

```ts
  1  | import { expect, test } from '@playwright/test'
  2  | 
  3  | async function login(page: import('@playwright/test').Page) {
  4  |   await page.addInitScript(() => {
  5  |     Date = class extends Date {
  6  |       constructor(...args: ConstructorParameters<typeof Date>) {
  7  |         super(args.length === 0 ? '2026-08-28T12:00:00Z' : args[0])
  8  |       }
  9  |       static now = () => new Date('2026-08-28T12:00:00Z').valueOf()
  10 |     } as DateConstructor
  11 |   })
  12 |   await page.goto('/login')
> 13 |   await page.getByLabel('E-mail').fill('playwright@example.com')
     |                                   ^ Error: locator.fill: Test timeout of 30000ms exceeded.
  14 |   await page.getByLabel('Пароль').fill('password123')
  15 |   await page.getByRole('button', { name: /войти|sign in/i }).click()
  16 |   await expect(page).toHaveURL(/\/$/)
  17 | }
  18 | 
  19 | for (const viewport of [
  20 |   { name: 'mobile', width: 320, height: 740 },
  21 |   { name: 'tablet', width: 768, height: 900 },
  22 |   { name: 'desktop', width: 1440, height: 960 },
  23 | ]) {
  24 |   test(`inventory remains usable on ${viewport.name}`, async ({ page }) => {
  25 |     await page.setViewportSize(viewport)
  26 |     await login(page)
  27 |     await expect(page.getByRole('button', { name: /добавить продукт|add product/i })).toBeVisible()
  28 |     await expect(page.getByRole('navigation')).toBeVisible()
  29 |     expect(await page.locator('html').evaluate((element) => element.scrollWidth <= element.clientWidth)).toBe(true)
  30 |     await expect(page).toHaveScreenshot(`inventory-${viewport.name}.png`, { fullPage: true })
  31 |   })
  32 | }
  33 | 
  34 | test('mobile add-product sheet opens, is keyboard-dismissable, and returns focus', async ({ page }) => {
  35 |   await page.setViewportSize({ width: 375, height: 740 })
  36 |   await login(page)
  37 | 
  38 |   const trigger = page.getByRole('button', { name: /добавить продукт|add product/i })
  39 |   await trigger.click()
  40 |   const dialog = page.getByRole('dialog', { name: /добавить продукт|add product/i })
  41 |   await expect(dialog).toBeVisible()
  42 |   await expect(dialog.getByRole('link', { name: /ввести вручную|enter manually/i })).toBeVisible()
  43 |   await expect(dialog.getByRole('link', { name: /заполнить по фото|fill from photo/i })).toBeVisible()
  44 | 
  45 |   await page.keyboard.press('Escape')
  46 |   await expect(dialog).toBeHidden()
  47 |   await expect(trigger).toBeFocused()
  48 | })
  49 | 
  50 | test('reduced motion removes the sheet transition', async ({ page }) => {
  51 |   await page.emulateMedia({ reducedMotion: 'reduce' })
  52 |   await page.setViewportSize({ width: 375, height: 740 })
  53 |   await login(page)
  54 | 
  55 |   await page.getByRole('button', { name: /добавить продукт|add product/i }).click()
  56 |   const sheet = page.locator('.add-sheet')
  57 |   await expect(sheet).toBeVisible()
  58 |   const animationName = await sheet.evaluate((element) => getComputedStyle(element).animationName)
  59 |   expect(animationName).toBe('none')
  60 | })
  61 | 
  62 | test('shows the server-provided status text for active, attention and expired fixtures', async ({ page }) => {
  63 |   await login(page)
  64 | 
  65 |   const table = page.locator('#inventory-table')
  66 |   await expect(table.getByText('Активен', { exact: true })).toBeVisible()
  67 |   await expect(table.getByText('Требует внимания', { exact: true })).toBeVisible()
  68 |   await expect(table.getByText('Срок истёк', { exact: true })).toBeVisible()
  69 | })
  70 | 
  71 | test('route guard does not expose inventory without a session', async ({ page }) => {
  72 |   await page.goto('/')
  73 |   await expect(page).toHaveURL(/\/login$/)
  74 |   await expect(page.getByRole('heading', { name: /войти|sign in/i })).toBeVisible()
  75 |   await expect(page.locator('#inventory-table')).toHaveCount(0)
  76 | })
  77 | 
  78 | test('HTML-like product names render as text instead of executable markup', async ({ page }) => {
  79 |   await login(page)
  80 |   await page.getByRole('button', { name: /добавить продукт|add product/i }).click()
  81 |   await page.getByRole('link', { name: /ввести вручную|enter manually/i }).click()
  82 |   await page.getByLabel('Название').fill('<img src=x onerror=alert(1)>')
  83 |   await page.getByLabel('Дата').fill('2026-09-01')
  84 |   await page.getByRole('button', { name: /сохранить продукт|save product/i }).click()
  85 |   await expect(page).toHaveURL(/\/$/)
  86 |   await expect(page.getByText('<img src=x onerror=alert(1)>', { exact: true })).toBeVisible()
  87 |   await expect(page.locator('#inventory-table img')).toHaveCount(0)
  88 | })
  89 | 
```