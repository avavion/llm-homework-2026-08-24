import { fireEvent, render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { expect, test } from 'vitest'
import { App } from './app'
import { api } from './api'

async function renderApp(path: string, authenticated = false) {
  if (authenticated) await api.auth.login('test@example.com', 'password123')
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(<QueryClientProvider client={client}><MemoryRouter initialEntries={[path]}><App /></MemoryRouter></QueryClientProvider>)
}

test('protects the home route without a session', async () => {
  await api.auth.logout()
  await renderApp('/')
  expect(await screen.findByRole('heading', { name: /Войти|Sign in/ })).toBeInTheDocument()
})

test('renders the sign-in screen as an accessible auth layout', async () => {
  await api.auth.logout()
  await renderApp('/login')
  expect(await screen.findByTestId('auth-story')).toBeInTheDocument()
  expect(screen.getByRole('main')).toHaveClass('auth-layout')
  expect(screen.getByLabelText(/пароль|password/i)).toHaveAttribute('autocomplete', 'current-password')
})

test('links both auth validation errors to their fields', async () => {
  await api.auth.logout()
  await renderApp('/register')
  fireEvent.click(screen.getByRole('button', { name: /создать аккаунт|create account/i }))
  const summary = await screen.findByRole('alert', { name: /исправьте ошибки|fix the errors/i })
  expect(summary).toHaveFocus()
  expect(summary).toContainElement(screen.getByRole('link', { name: /корректный e-mail|valid email/i }))
  expect(screen.getByRole('link', { name: /минимум 8 символов|at least 8 characters/i })).toHaveAttribute('href', '#password')
  expect(screen.getByLabelText(/пароль|password/i)).toHaveAttribute('autocomplete', 'new-password')
})

test('renders navigation and inventory on the home route for an authenticated session', async () => {
  await renderApp('/', true)
  expect(await screen.findByRole('navigation', { name: /основная навигация|primary navigation/i })).toBeInTheDocument()
  expect(screen.getByRole('link', { name: /мои продукты|my products/i })).toBeInTheDocument()
  expect(await screen.findByRole('heading', { name: /использовать первым|use first/i })).toBeInTheDocument()
})

test('redirects the legacy /products path to the home route', async () => {
  await renderApp('/products', true)
  expect(await screen.findByRole('heading', { name: /использовать первым|use first/i })).toBeInTheDocument()
})

test('links from the priority ribbon to the full inventory table', async () => {
  await renderApp('/', true)
  const link = await screen.findByRole('link', { name: /все продукты|all products/i })
  expect(link).toHaveAttribute('href', '#inventory-table')
})

test('keeps a photo product as a draft until explicit approve', async () => {
  await api.auth.login('test@example.com', 'password123')
  const before = await api.products.list()
  const draft = await api.drafts.recognize(new File(['image'], 'label.png', { type: 'image/png' }))
  expect((await api.products.list())).toHaveLength(before.length)
  await api.drafts.approve(draft.id, { name: 'Йогурт', dateType: 'best_before', expiryDate: '2026-09-01', location: '', group: '' })
  expect((await api.products.list())).toHaveLength(before.length + 1)
})

test('renders required non-product routes instead of a 404', async () => {
  await renderApp('/recipes', true)
  expect(await screen.findByRole('heading', { name: /Рецепты|Recipes/ })).toBeInTheDocument()
})

test('does not offer a theme picker in settings', async () => {
  await renderApp('/settings', true)
  expect(await screen.findByRole('heading', { level: 1, name: /настройки|settings/i })).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /тёмная|dark/i })).not.toBeInTheDocument()
})

test('opens and closes the mobile add-product sheet', async () => {
  await renderApp('/', true)
  fireEvent.click(await screen.findByRole('button', { name: /добавить продукт|add product/i }))
  expect(screen.getByRole('dialog', { name: /добавить продукт|add product/i })).toBeInTheDocument()
  fireEvent.keyDown(document, { key: 'Escape' })
  expect(screen.queryByRole('dialog', { name: /добавить продукт|add product/i })).not.toBeInTheDocument()
})

test('keeps keyboard focus inside the lifecycle dialog', async () => {
  await renderApp('/', true)
  const triggers = await screen.findAllByRole('button', { name: /использован|used/i })
  fireEvent.click(triggers[0])
  const dialog = screen.getByRole('dialog')
  const buttons = screen.getAllByRole('button', { hidden: false }).filter((button) => dialog.contains(button))
  buttons.at(-1)?.focus()
  fireEvent.keyDown(document, { key: 'Tab' })
  expect(buttons[0]).toHaveFocus()
})
