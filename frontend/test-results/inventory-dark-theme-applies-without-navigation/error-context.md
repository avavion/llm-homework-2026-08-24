# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: inventory.spec.ts >> dark theme applies without navigation
- Location: e2e/inventory.spec.ts:21:1

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: locator.click: Test timeout of 30000ms exceeded.
Call log:
  - waiting for getByRole('button', { name: /тёмная|dark/i })

```

# Page snapshot

```yaml
- main [ref=f1e3]:
  - generic [ref=f1e4]:
    - heading "Sign in" [level=1] [ref=f1e5]
    - generic [ref=f1e6]:
      - generic [ref=f1e7]:
        - text: E-mail
        - textbox "E-mail" [ref=f1e8]
      - generic [ref=f1e9]:
        - text: Пароль
        - textbox "Пароль" [ref=f1e10]
      - button "Sign in" [ref=f1e11] [cursor=pointer]
      - link "Create account" [ref=f1e12] [cursor=pointer]:
        - /url: /register
```

# Test source

```ts
  1  | import { expect, test } from '@playwright/test'
  2  | 
  3  | for (const viewport of [
  4  |   { name: 'mobile', width: 320, height: 740 },
  5  |   { name: 'tablet', width: 768, height: 900 },
  6  |   { name: 'desktop', width: 1440, height: 960 },
  7  | ]) {
  8  |   test(`inventory remains usable on ${viewport.name}`, async ({ page }) => {
  9  |     await page.setViewportSize(viewport)
  10 |     await page.goto('/login')
  11 |     await page.getByLabel('E-mail').fill('playwright@example.com')
  12 |     await page.getByLabel('Пароль').fill('password123')
  13 |     await page.getByRole('button', { name: /войти|sign in/i }).click()
  14 |     await expect(page).toHaveURL(/\/products$/)
  15 |     await expect(page.getByRole('link', { name: /добавить продукт|add product/i })).toBeVisible()
  16 |     await expect(page.getByRole('navigation')).toBeVisible()
  17 |     expect(await page.locator('html').evaluate((element) => element.scrollWidth <= element.clientWidth)).toBe(true)
  18 |   })
  19 | }
  20 | 
  21 | test('dark theme applies without navigation', async ({ page }) => {
  22 |   await page.goto('/login')
  23 |   await page.getByLabel('E-mail').fill('playwright@example.com')
  24 |   await page.getByLabel('Пароль').fill('password123')
  25 |   await page.getByRole('button', { name: /войти|sign in/i }).click()
  26 |   await expect(page).toHaveURL(/\/products$/)
  27 |   await page.goto('/settings')
> 28 |   await page.getByRole('button', { name: /тёмная|dark/i }).click()
     |                                                            ^ Error: locator.click: Test timeout of 30000ms exceeded.
  29 |   await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')
  30 | })
  31 | 
```