import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { expect, test } from 'vitest'
import { App } from './app'
test('renders the product empty state and a skip link', () => { render(<MemoryRouter initialEntries={['/products']}><App /></MemoryRouter>); expect(screen.getByRole('heading', {name:/Мои продукты|My products/})).toBeInTheDocument(); expect(screen.getByRole('link',{name:'Skip to content'})).toHaveAttribute('href','#main-content') })
test('renders a not-found page', () => { render(<MemoryRouter initialEntries={['/missing']}><App /></MemoryRouter>); expect(screen.getByRole('heading',{name:/Страница не найдена|Page not found/})).toBeInTheDocument() })
