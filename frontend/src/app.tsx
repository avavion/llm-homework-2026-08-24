import { Link, Navigate, Outlet, Route, Routes, useLocation, useNavigate, useParams } from 'react-router-dom'
import { useMemo, useState, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { api, ApiError, apiMode } from './api'
import { Alert, AppShell, EmptyState, Skeleton, StatusBadge, statusLabel, type Status } from './ui'
import { locale, t } from './i18n'
import type { Product, ProductInput } from './mock-api'

const productKeys = { list: ['products'] as const, detail: (id: string) => ['products', id] as const }
const message = (error: unknown) => error instanceof ApiError && error.status === 401 ? 'Сессия завершена. Войдите снова.' : 'Не удалось выполнить действие. Проверьте соединение и повторите.'
const Page = ({ title, children }: { title: string; children?: ReactNode }) => <div className="page-stack"><h1>{title}</h1>{children}</div>
const PublicPage = ({ title, children }: { title: string; children?: ReactNode }) => <main id="main-content" className="public-page"><Page title={title}>{children}</Page></main>
const dateTypeLabel = (type: Product['dateType']) => type === 'use_by' ? 'Годен до' : 'Лучше употребить до'

function ProtectedLayout() {
  const location = useLocation()
  const session = useQuery({ queryKey: ['session'], queryFn: api.auth.session, retry: false, staleTime: 60_000 })
  const [addOpen, setAddOpen] = useState(false)
  if (session.isPending) return <PublicPage title={t.loading}><Skeleton /></PublicPage>
  if (session.isError) return <Navigate to="/login" replace state={{ from: location.pathname }} />
  return <AppShell addOpen={addOpen} onOpenAdd={() => setAddOpen(true)} onCloseAdd={() => setAddOpen(false)} railFooter={<LogoutButton />}><Outlet /></AppShell>
}

function LogoutButton() {
  const navigate = useNavigate(); const queryClient = useQueryClient()
  const logout = useMutation({ mutationFn: api.auth.logout, onSuccess: () => { queryClient.removeQueries({ queryKey: ['session'] }); navigate('/login') } })
  return <button className="button-secondary" type="button" disabled={logout.isPending} onClick={() => logout.mutate()}>{logout.isPending ? 'Выходим…' : 'Выйти'}</button>
}

function RequestError({ retry }: { retry: () => void }) { return <Alert tone="danger">Не удалось загрузить данные. <button className="inline-retry" type="button" onClick={retry}>Повторить</button></Alert> }

function ProductList() {
  const [query, setQuery] = useState(''); const [status, setStatus] = useState<Status | 'all'>('all'); const queryClient = useQueryClient()
  const products = useQuery({ queryKey: productKeys.list, queryFn: api.products.list, retry: false })
  const complete = useMutation({ mutationFn: ({ id, action }: { id: string; action: 'used' | 'discarded' }) => api.products.complete(id, action), onSuccess: () => queryClient.invalidateQueries({ queryKey: productKeys.list }) })
  const shown = useMemo(() => (products.data ?? []).filter((item) => (status === 'all' || item.status === status) && item.name.toLowerCase().includes(query.toLowerCase())).sort((a, b) => a.expiryDate.localeCompare(b.expiryDate)), [products.data, query, status])
  const urgent = useMemo(() => (products.data ?? []).filter((item) => !['used', 'discarded'].includes(item.status)).sort((a, b) => a.expiryDate.localeCompare(b.expiryDate)).slice(0, 3), [products.data])
  const [heroLine1, heroLine2] = t.heroTitle.split('\n')
  const statusFilters: { value: Status | 'all'; label: string }[] = [{ value: 'all', label: t.filterAll }, { value: 'active', label: statusLabel('active') }, { value: 'attention', label: statusLabel('attention') }, { value: 'expired', label: statusLabel('expired') }]
  return <div className="page-stack pantry-home">
    <section className="hero"><span className="kicker">{t.todayKicker}</span><h1>{heroLine1}<br />{heroLine2}</h1><p className="hero-sub">{t.heroSubtitle}</p></section>
    <section className="priority" aria-labelledby="priority-title">
      <div className="priority-intro"><span className="priority-eyebrow">{t.priorityEyebrow}</span><h2 id="priority-title">{t.urgency}</h2><span className="priority-count">{urgent.length}</span></div>
      {products.isPending ? <Skeleton /> : urgent.map((item, index) => <article className="priority-card" key={item.id}><span className="card-num">{String(index + 1).padStart(2, '0')}</span><h3>{item.name}</h3><p>{item.location || '—'} · <time dateTime={item.expiryDate}>{item.expiryDate}</time></p><StatusBadge status={item.status} /></article>)}
    </section>
    <div className="rule-note"><StatusBadge status="research_required" /><Alert>{t.ruleText}</Alert></div>
    <section className="controls">
      <h2>{t.products} <span className="count">{shown.length}</span></h2>
      <div className="filter-group">{statusFilters.map((filter) => <button key={filter.value} type="button" className={`filter${status === filter.value ? ' active' : ''}`} aria-pressed={status === filter.value} onClick={() => setStatus(filter.value)}>{filter.label}</button>)}</div>
    </section>
    <label className="search-field">{t.search}<input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t.search} /></label>
    {products.isPending ? <section className="table" aria-busy="true"><Skeleton /><Skeleton /><Skeleton /></section> : products.isError ? <RequestError retry={() => products.refetch()} /> : shown.length === 0 ? <EmptyState action={<Link className="button-link" to="/products/new">{t.add}</Link>}>{t.empty}</EmptyState> : (
      <div className="table" role="table" aria-label={t.products}>
        <div className="head" role="row"><span role="columnheader">{t.tableProduct}</span><span role="columnheader">{t.tableStorage}</span><span role="columnheader">{t.tableDate}</span><span role="columnheader">{t.tableStatus}</span><span role="columnheader" aria-hidden="true" /></div>
        {shown.map((item) => <div className="tr" role="row" key={item.id}>
          <span role="cell" className="name"><Link to={`/products/${item.id}`}>{item.name}</Link><span className="row-meta">{item.location || '—'} · {item.expiryDate}</span></span>
          <span role="cell" className="cell">{item.location || '—'}</span>
          <span role="cell" className="cell"><time dateTime={item.expiryDate}>{item.expiryDate}</time><br />{dateTypeLabel(item.dateType)}</span>
          <span role="cell"><StatusBadge status={item.status} /></span>
          <span role="cell" className="row-actions">{!['used', 'discarded'].includes(item.status) && <><button disabled={complete.isPending} onClick={() => complete.mutate({ id: item.id, action: 'used' })}>{complete.isPending ? '…' : t.used}</button><button className="button-secondary" disabled={complete.isPending} onClick={() => complete.mutate({ id: item.id, action: 'discarded' })}>{t.discarded}</button></>}{complete.isError && <Alert tone="danger">{message(complete.error)}</Alert>}</span>
        </div>)}
      </div>
    )}
  </div>
}

const productSchema = z.object({ name: z.string().trim().min(1, 'Введите название'), dateType: z.enum(['use_by', 'best_before']), expiryDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Укажите дату'), location: z.string(), group: z.string() })
type FormData = z.infer<typeof productSchema>
const errorId = (field: keyof FormData) => `error-${field}`

function ProductForm({ title, defaults, onSave }: { title: string; defaults?: Partial<ProductInput>; onSave?: (input: ProductInput) => Promise<unknown> }) {
  const navigate = useNavigate(); const queryClient = useQueryClient(); const [serverError, setServerError] = useState<string>()
  const form = useForm<FormData>({ resolver: zodResolver(productSchema), defaultValues: { name: '', dateType: 'use_by', expiryDate: '', location: '', group: '', ...defaults } })
  const create = useMutation({ mutationFn: (input: ProductInput) => onSave ? onSave(input) : api.products.create(input), onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: productKeys.list }); navigate('/', { state: { notice: 'Продукт добавлен.' } }) }, onError: (error) => setServerError(message(error)) })
  const submit = (data: FormData) => { setServerError(undefined); create.mutate(data) }
  const field = (name: keyof FormData) => ({ 'aria-invalid': Boolean(form.formState.errors[name]), 'aria-describedby': form.formState.errors[name] ? errorId(name) : undefined })
  return <Page title={title}><form className="form-card" onSubmit={form.handleSubmit(submit)} noValidate><div className="form-grid"><label>Название<input {...form.register('name')} {...field('name')} />{form.formState.errors.name && <span id={errorId('name')} role="alert">{form.formState.errors.name.message}</span>}</label><label>Тип даты<select {...form.register('dateType')}><option value="use_by">Годен до</option><option value="best_before">Лучше употребить до</option></select></label><label>Дата<input type="date" {...form.register('expiryDate')} {...field('expiryDate')} />{form.formState.errors.expiryDate && <span id={errorId('expiryDate')} role="alert">{form.formState.errors.expiryDate.message}</span>}</label><label>Место хранения<input {...form.register('location')} /></label><label>Группа<input {...form.register('group')} /></label></div><Alert tone="info">«Годен до» — дата безопасности. «Лучше употребить до» — дата качества.</Alert>{serverError && <Alert tone="danger">{serverError}</Alert>}<div className="form-actions"><button type="submit" disabled={create.isPending}>{create.isPending ? 'Сохраняем…' : onSave ? 'Сохранить после проверки' : t.save}</button>{!onSave && <Link to="/products/new/photo">{t.photo}</Link>}</div></form></Page>
}

function PhotoUpload() { const navigate = useNavigate(); const recognize = useMutation({ mutationFn: api.drafts.recognize, onSuccess: (draft) => navigate(`/product-drafts/${draft.id}`) }); return <Page title={t.photo}><div className="card stack"><label>Фото упаковки<input type="file" accept="image/*" disabled={recognize.isPending} onChange={(event) => { const file = event.target.files?.[0]; if (file) recognize.mutate(file) }} /></label>{recognize.isPending && <Alert>Распознаём данные…</Alert>}{recognize.isError && <RequestError retry={() => recognize.reset()} />}<Link to="/products/new">{t.manual}</Link></div></Page> }

function DraftReview() { const { id = '' } = useParams(); const navigate = useNavigate(); const draft = useQuery({ queryKey: ['draft', id], queryFn: () => api.drafts.get(id), retry: false }); const reject = useMutation({ mutationFn: () => api.drafts.reject(id), onSuccess: () => navigate('/products/new') }); if (draft.isPending) return <Page title={t.draft}><Skeleton /></Page>; if (draft.isError) return <Page title="Не удалось открыть черновик"><RequestError retry={() => draft.refetch()} /></Page>; return <><ProductForm title={t.draft} defaults={draft.data.fields} onSave={(input) => api.drafts.approve(id, input)} /><div className="form-actions"><button className="button-secondary" disabled={reject.isPending} onClick={() => reject.mutate()}>{reject.isPending ? 'Отклоняем…' : 'Отклонить черновик'}</button>{reject.isError && <Alert tone="danger">{message(reject.error)}</Alert>}</div></> }

function ProductDetail() { const { id = '' } = useParams(); const queryClient = useQueryClient(); const product = useQuery({ queryKey: productKeys.detail(id), queryFn: () => api.products.get(id), retry: false }); const complete = useMutation({ mutationFn: (action: 'used' | 'discarded') => api.products.complete(id, action), onSuccess: () => { queryClient.invalidateQueries({ queryKey: productKeys.list }); product.refetch() } }); if (product.isPending) return <Page title="Продукт"><Skeleton /></Page>; if (product.isError) return <Page title="Не удалось открыть продукт"><RequestError retry={() => product.refetch()} /></Page>; const item = product.data; return <Page title={item.name}><section className="card stack"><StatusBadge status={item.status} /><p>{dateTypeLabel(item.dateType)}: <time dateTime={item.expiryDate}>{item.expiryDate}</time></p><p>{item.location || 'Место хранения не указано'}</p>{!['used', 'discarded'].includes(item.status) && <div className="form-actions"><button disabled={complete.isPending} onClick={() => complete.mutate('used')}>Использован</button><button className="button-secondary" disabled={complete.isPending} onClick={() => complete.mutate('discarded')}>Выброшен</button></div>}{complete.isError && <Alert tone="danger">{message(complete.error)}</Alert>}</section></Page> }

function Recipes() { const recipes = useQuery({ queryKey: ['recipes'], queryFn: api.recipes, retry: false }); return <Page title={t.recipes}>{recipes.isPending ? <Skeleton /> : recipes.isError ? <RequestError retry={() => recipes.refetch()} /> : recipes.data.length === 0 ? <EmptyState action={<Link to="/">{t.products}</Link>}>{t.noRecipes}</EmptyState> : <ul className="product-list">{recipes.data.map((recipe) => <li className="card" key={recipe.title}><h2>{recipe.title}</h2><p>Рекомендация основана на доступных продуктах.</p></li>)}</ul>}</Page> }
function Settings() {
  return <Page title={t.settings}><section className="card stack"><div><h2>{locale === 'ru' ? 'Настройки приложения' : 'Application settings'}</h2><p className="muted">{locale === 'ru' ? 'Интерфейс использует единую светлую тему.' : 'The interface uses one light theme.'}</p></div><Alert>{t.settingsBlocked}</Alert><p>{t.profileBlocked}</p><p className="muted">{locale === 'ru' ? 'В этом окружении: ' : 'In this environment: '}{apiMode === 'fixture' ? (locale === 'ru' ? 'тестовые данные разработки' : 'development fixture data') : (locale === 'ru' ? 'подключён API-сервер' : 'API server connected')}.</p></section></Page> }

function Credentials({ register }: { register?: boolean }) { const navigate = useNavigate(); const location = useLocation(); const queryClient = useQueryClient(); const [error, setError] = useState<string>(); const form = useForm<{ email: string; password: string }>({ defaultValues: { email: '', password: '' } }); const action = useMutation({ mutationFn: ({ email, password }: { email: string; password: string }) => register ? api.auth.register(email, password) : api.auth.login(email, password), onSuccess: async () => { if (register) { navigate('/login'); return } await queryClient.invalidateQueries({ queryKey: ['session'] }); navigate((location.state as { from?: string } | null)?.from ?? '/') }, onError: (failure) => setError(message(failure)) }); return <PublicPage title={register ? t.register : t.login}>{register && <Alert tone="warning">{t.profileBlocked}</Alert>}<form className="form-card" onSubmit={form.handleSubmit((values) => { setError(undefined); action.mutate(values) })} noValidate><label>E-mail<input type="email" autoComplete="email" {...form.register('email', { required: 'Введите e-mail' })} /></label><label>Пароль<input type="password" autoComplete={register ? 'new-password' : 'current-password'} {...form.register('password', { required: 'Введите пароль', minLength: { value: 8, message: 'Минимум 8 символов' } })} /></label>{error && <Alert tone="danger">{error}</Alert>}<button disabled={action.isPending} type="submit">{action.isPending ? 'Отправляем…' : register ? t.register : t.login}</button><Link to={register ? '/login' : '/register'}>{register ? t.login : t.register}</Link></form></PublicPage> }

export function App() { return <Routes><Route path="/login" element={<Credentials />} /><Route path="/register" element={<Credentials register />} /><Route element={<ProtectedLayout />}><Route path="/" element={<ProductList />} /><Route path="/products" element={<Navigate to="/" replace />} /><Route path="/products/new" element={<ProductForm title={t.add} />} /><Route path="/products/new/photo" element={<PhotoUpload />} /><Route path="/product-drafts/:id" element={<DraftReview />} /><Route path="/products/:id" element={<ProductDetail />} /><Route path="/recipes" element={<Recipes />} /><Route path="/settings" element={<Settings />} /></Route><Route path="*" element={<PublicPage title={t.notFound}><Link to="/">{t.back}</Link></PublicPage>} /></Routes> }
