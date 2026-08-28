import { Link, Navigate, Outlet, Route, Routes, useLocation, useNavigate, useParams } from 'react-router-dom'
import { useEffect, useId, useMemo, useRef, useState, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { Package } from 'lucide-react'
import { api, ApiError, apiMode, type NotificationSetting, type Profile } from './api'
import { Alert, AppShell, EmptyState, Skeleton, StatusBadge, statusLabel, trapDialogTabKey, type Status } from './ui'
import { locale, t } from './i18n'
import type { Product, ProductInput } from './mock-api'

const productKeys = { list: ['products'] as const, detail: (id: string) => ['products', id] as const }
const message = (error: unknown) => error instanceof ApiError && error.status === 401 ? t.sessionEnded : t.requestFailed
const Page = ({ title, kicker, subtitle, children }: { title: string; kicker?: string; subtitle?: string; children?: ReactNode }) => <div className="page-stack"><section className="hero">{kicker && <span className="kicker">{kicker}</span>}<h1>{title}</h1>{subtitle && <p className="hero-sub">{subtitle}</p>}</section>{children}</div>
const PublicPage = ({ title, kicker, subtitle, variant, children }: { title: string; kicker?: string; subtitle?: string; variant?: 'auth'; children?: ReactNode }) => variant === 'auth'
  ? <main id="main-content" className="auth-layout">
      <aside className="auth-story" data-testid="auth-story" aria-label={t.authStoryLabel}>
        <p className="auth-story__eyebrow">Pantry Ledger</p>
        <p className="auth-story__quote">{t.authStoryQuote}</p>
      </aside>
      <section className="auth-panel"><Page title={title} kicker={kicker} subtitle={subtitle}>{children}</Page></section>
    </main>
  : <main id="main-content" className="public-page"><Page title={title} kicker={kicker} subtitle={subtitle}>{children}</Page></main>
const dateTypeLabel = (type: Product['dateType']) => type === 'use_by' ? t.useBy : t.bestBefore

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
  return <button className="button-secondary" type="button" disabled={logout.isPending} onClick={() => logout.mutate()}>{logout.isPending ? t.loggingOut : t.logOut}</button>
}

function RequestError({ retry }: { retry: () => void }) { return <Alert tone="danger">{t.loadFailed} <button className="inline-retry" type="button" onClick={retry}>{t.retry}</button></Alert> }

function LifecycleDialog({ product, action, onClose, onConfirm, pending, error }: { product: Product; action: 'used' | 'discarded'; onClose: () => void; onConfirm: () => void; pending: boolean; error: unknown }) {
  const dialogRef = useRef<HTMLElement>(null); const confirmRef = useRef<HTMLButtonElement>(null); const titleId = useId()
  useEffect(() => { confirmRef.current?.focus(); const onKeyDown = (event: KeyboardEvent) => { if (event.key === 'Escape' && !pending) onClose(); trapDialogTabKey(event, dialogRef.current) }; document.addEventListener('keydown', onKeyDown); return () => document.removeEventListener('keydown', onKeyDown) }, [onClose, pending])
  const isUsed = action === 'used'
  return <div className="dialog-backdrop" onMouseDown={() => !pending && onClose()}><section ref={dialogRef} className="confirm-dialog" role="dialog" aria-modal="true" aria-labelledby={titleId} onMouseDown={(event) => event.stopPropagation()}><h2 id={titleId}>{isUsed ? t.confirmUsedTitle : t.confirmDiscardedTitle}</h2><p>{isUsed ? t.confirmUsedText : t.confirmDiscardedText}</p><p className="muted">{product.name}</p>{error ? <Alert tone="danger">{message(error)}</Alert> : null}<div className="form-actions"><button ref={confirmRef} type="button" disabled={pending} onClick={onConfirm}>{pending ? t.updating : t.update}</button><button className="button-secondary" type="button" disabled={pending} onClick={onClose}>{t.cancel}</button></div></section></div>
}

function LifecycleActions({ product, isPending, error, onComplete }: { product: Product; isPending: boolean; error: unknown; onComplete: (action: 'used' | 'discarded', onSuccess: () => void) => void }) {
  const [action, setAction] = useState<'used' | 'discarded' | null>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const close = () => { setAction(null); window.setTimeout(() => triggerRef.current?.focus(), 0) }
  const open = (next: 'used' | 'discarded', trigger: HTMLButtonElement) => { triggerRef.current = trigger; setAction(next) }
  const confirm = () => { if (!action) return; onComplete(action, close) }
  if (['used', 'discarded'].includes(product.status)) return null
  return <><div className="row-actions"><button ref={triggerRef} type="button" disabled={isPending} onClick={(event) => open('used', event.currentTarget)}>{t.used}</button><button className="button-secondary" type="button" disabled={isPending} onClick={(event) => open('discarded', event.currentTarget)}>{t.discarded}</button></div>{action && <LifecycleDialog product={product} action={action} onClose={close} onConfirm={confirm} pending={isPending} error={error} />}</>
}

function ProductList() {
  const [query, setQuery] = useState(''); const [status, setStatus] = useState<Status | 'all'>('all'); const queryClient = useQueryClient()
  const products = useQuery({ queryKey: productKeys.list, queryFn: api.products.list, retry: false })
  const complete = useMutation({ mutationFn: ({ id, action }: { id: string; action: 'used' | 'discarded' }) => api.products.complete(id, action), onSuccess: () => queryClient.invalidateQueries({ queryKey: productKeys.list }) })
  const shown = useMemo(() => (products.data ?? []).filter((item) => (status === 'all' || item.status === status) && item.name.toLowerCase().includes(query.toLowerCase())).sort((a, b) => a.expiryDate.localeCompare(b.expiryDate)), [products.data, query, status])
  const urgent = useMemo(() => (products.data ?? []).filter((item) => !['used', 'discarded'].includes(item.status)).sort((a, b) => a.expiryDate.localeCompare(b.expiryDate)).slice(0, 3), [products.data])
  const [heroLine1, heroLine2] = t.heroTitle.split('\n')
  const todayLabel = useMemo(() => new Date().toLocaleDateString(locale === 'ru' ? 'ru-RU' : 'en-US', { weekday: 'long', day: 'numeric', month: 'long' }), [])
  const statusFilters: { value: Status | 'all'; label: string }[] = [{ value: 'all', label: t.filterAll }, { value: 'active', label: statusLabel('active') }, { value: 'attention', label: statusLabel('attention') }, { value: 'expired', label: statusLabel('expired') }]
  return <div className="dashboard"><section className="hero hero--dashboard"><div className="hero-title"><span className="kicker">{todayLabel}</span><h1>{heroLine1}<br />{heroLine2}</h1></div><p className="hero-sub">{t.heroSubtitle}</p></section>
    {products.isPending ? <section className="priority" aria-labelledby="priority-title"><div className="priority-intro"><span className="priority-eyebrow">{t.priorityEyebrow}</span><h2 id="priority-title">{t.urgency}</h2></div><Skeleton /></section> : (products.data ?? []).length > 0 && <section className="priority" aria-labelledby="priority-title"><div className="priority-intro"><span className="priority-eyebrow">{t.priorityEyebrow}</span><h2 id="priority-title">{t.urgency}</h2><span className="priority-count">{urgent.length}</span>{urgent.length > 0 && <a className="priority-link" href="#inventory-table">{t.allProducts} →</a>}</div>{urgent.length === 0 ? <p className="priority-empty">{t.allCaughtUp}</p> : urgent.map((item, index) => <article className="priority-card" key={item.id}><span className="card-num">{String(index + 1).padStart(2, '0')}</span><h3>{item.name}</h3><p>{item.location || '—'} · <time dateTime={item.expiryDate}>{item.expiryDate}</time></p><StatusBadge status={item.status} /></article>)}</section>}
    <div className="rule-note"><StatusBadge status="research_required" /><Alert>{t.ruleText}</Alert></div><section className="controls"><h2>{t.products} <span className="count">{shown.length}</span></h2><div className="filter-group">{statusFilters.map((filter) => <button key={filter.value} type="button" className={`filter${status === filter.value ? ' active' : ''}`} aria-pressed={status === filter.value} onClick={() => setStatus(filter.value)}>{filter.label}</button>)}</div></section><label className="search-field">{t.search}<input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t.search} /></label>
    {products.isPending ? <section className="table" aria-busy="true"><Skeleton /><Skeleton /><Skeleton /></section> : products.isError ? <RequestError retry={() => products.refetch()} /> : shown.length === 0 ? <EmptyState action={<Link className="button-link" to="/products/new">{t.add}</Link>}>{t.empty}</EmptyState> : <div className="table" id="inventory-table" role="table" aria-label={t.products}><div className="head" role="row"><span role="columnheader">{t.tableProduct}</span><span role="columnheader">{t.tableStorage}</span><span role="columnheader">{t.tableDate}</span><span role="columnheader">{t.tableStatus}</span><span role="columnheader" aria-hidden="true" /></div>{shown.map((item) => <div className="tr" role="row" key={item.id}><span role="cell" className="name"><span className="row-icon" aria-hidden="true"><Package size={16} /></span><span className="name-copy"><Link to={`/products/${item.id}`}>{item.name}</Link><span className="row-meta">{item.location || '—'} · {item.expiryDate}</span></span></span><span role="cell" className="cell">{item.location || '—'}</span><span role="cell" className="cell"><time dateTime={item.expiryDate}>{item.expiryDate}</time><br />{dateTypeLabel(item.dateType)}</span><span role="cell"><StatusBadge status={item.status} /></span><span role="cell"><LifecycleActions product={item} isPending={complete.isPending} error={complete.error} onComplete={(action, onSuccess) => complete.mutate({ id: item.id, action }, { onSuccess })} />{complete.isError && <Alert tone="danger">{message(complete.error)}</Alert>}</span></div>)}</div>}
  </div>
}

const productSchema = z.object({ name: z.string().trim().min(1, t.enterEmail), dateType: z.enum(['use_by', 'best_before']), expiryDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, t.date), location: z.string(), group: z.string(), quantity: z.string().refine((value) => value === '' || (Number.isFinite(Number(value)) && Number(value) >= 0), t.quantity), unit: z.string() })
type FormData = z.infer<typeof productSchema>
const errorId = (field: keyof FormData) => `error-${field}`
const toProductInput = ({ quantity, unit, ...data }: FormData): ProductInput => ({ ...data, ...(quantity === '' ? {} : { quantity: Number(quantity) }), ...(unit.trim() ? { unit: unit.trim() } : {}) })

function ProductForm({ title, defaults, onSave }: { title: string; defaults?: Partial<ProductInput>; onSave?: (input: ProductInput) => Promise<unknown> }) {
  const navigate = useNavigate(); const queryClient = useQueryClient(); const [serverError, setServerError] = useState<string>()
  const profile = useQuery({ queryKey: ['profile'], queryFn: api.profile.get, retry: false })
  const form = useForm<FormData>({ resolver: zodResolver(productSchema), defaultValues: { name: defaults?.name ?? '', dateType: defaults?.dateType ?? 'use_by', expiryDate: defaults?.expiryDate ?? '', location: defaults?.location ?? '', group: defaults?.group ?? '', quantity: defaults?.quantity?.toString() ?? '', unit: defaults?.unit ?? '' } })
  const create = useMutation({ mutationFn: (input: ProductInput) => { const withCountry = profile.data?.country_code ? { ...input, countryCode: profile.data.country_code } : input; return onSave ? onSave(withCountry) : api.products.create(withCountry) }, onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: productKeys.list }); navigate('/', { state: { notice: onSave ? t.draftApproved : t.productAdded } }) }, onError: (error) => setServerError(message(error)) })
  const submit = (data: FormData) => { setServerError(undefined); create.mutate(toProductInput(data)) }
  const field = (name: keyof FormData) => ({ 'aria-invalid': Boolean(form.formState.errors[name]), 'aria-describedby': form.formState.errors[name] ? errorId(name) : undefined })
  return <Page title={title}><form className="form-card" onSubmit={form.handleSubmit(submit)} noValidate><div className="form-grid"><label>{t.productName}<input {...form.register('name')} {...field('name')} />{form.formState.errors.name && <span id={errorId('name')} role="alert">{form.formState.errors.name.message}</span>}</label><label>{t.dateType}<select {...form.register('dateType')}><option value="use_by">{t.useBy}</option><option value="best_before">{t.bestBefore}</option></select></label><label>{t.date}<input type="date" {...form.register('expiryDate')} {...field('expiryDate')} />{form.formState.errors.expiryDate && <span id={errorId('expiryDate')} role="alert">{form.formState.errors.expiryDate.message}</span>}</label></div><details className="storage-details"><summary>{t.storageDetails} <span>{t.optional}</span></summary><div className="form-grid"><label>{t.location}<input {...form.register('location')} /></label><label>{t.group}<input {...form.register('group')} /></label><label>{t.quantity}<input inputMode="decimal" {...form.register('quantity')} {...field('quantity')} />{form.formState.errors.quantity && <span id={errorId('quantity')} role="alert">{form.formState.errors.quantity.message}</span>}</label><label>{t.unit}<input {...form.register('unit')} /></label></div></details><Alert tone="info">{t.safetyHint}</Alert>{profile.isPending && <Alert>{t.loading}</Alert>}{!profile.isPending && !profile.data?.country_code && <Alert tone="warning">{t.profileBlocked} <Link to="/settings">{t.settings}</Link></Alert>}{profile.isError && <Alert tone="warning">{t.requestFailed}</Alert>}{serverError && <Alert tone="danger">{serverError}</Alert>}<div className="form-actions"><button type="submit" disabled={create.isPending || profile.isPending}>{create.isPending ? t.saving : onSave ? t.saveAfterReview : t.save}</button>{!onSave && <Link to="/products/new/photo">{t.photo}</Link>}</div></form></Page>
}

function PhotoUpload() { const navigate = useNavigate(); const recognize = useMutation({ mutationFn: api.drafts.recognize, onSuccess: (draft) => navigate(`/product-drafts/${draft.id}`) }); return <Page title={t.photo}><div className="card stack"><label>{t.packagePhoto}<input type="file" accept="image/*" disabled={recognize.isPending} onChange={(event) => { const file = event.target.files?.[0]; if (file) recognize.mutate(file) }} /></label>{recognize.isPending && <Alert>{t.uploading}</Alert>}{recognize.isError && <RequestError retry={() => recognize.reset()} />}<Link to="/products/new">{t.manual}</Link></div></Page> }

function DraftReview() { const { id = '' } = useParams(); const navigate = useNavigate(); const draft = useQuery({ queryKey: ['draft', id], queryFn: () => api.drafts.get(id), retry: false }); const reject = useMutation({ mutationFn: () => api.drafts.reject(id), onSuccess: () => navigate('/products/new', { state: { notice: t.draftRejected } }) }); if (draft.isPending) return <Page title={t.draft}><Skeleton /></Page>; if (draft.isError) return <Page title={t.openDraftFailed}><RequestError retry={() => draft.refetch()} /></Page>; return <><ProductForm title={t.draft} defaults={draft.data.fields} onSave={(input) => api.drafts.approve(id, input)} /><div className="form-actions"><button className="button-secondary" type="button" disabled={reject.isPending} onClick={() => reject.mutate()}>{reject.isPending ? t.rejecting : t.rejectDraft}</button>{reject.isError && <Alert tone="danger">{message(reject.error)}</Alert>}</div></> }

function ProductDetail() { const { id = '' } = useParams(); const queryClient = useQueryClient(); const product = useQuery({ queryKey: productKeys.detail(id), queryFn: () => api.products.get(id), retry: false }); const complete = useMutation({ mutationFn: (action: 'used' | 'discarded') => api.products.complete(id, action), onSuccess: () => { queryClient.invalidateQueries({ queryKey: productKeys.list }); product.refetch() } }); if (product.isPending) return <Page title={t.product}><Skeleton /></Page>; if (product.isError) return <Page title={t.openProductFailed}><RequestError retry={() => product.refetch()} /></Page>; const item = product.data; return <Page title={item.name}><section className="card stack"><StatusBadge status={item.status} /><p>{dateTypeLabel(item.dateType)}: <time dateTime={item.expiryDate}>{item.expiryDate}</time></p><p>{item.location || t.unspecifiedLocation}</p><LifecycleActions product={item} isPending={complete.isPending} error={complete.error} onComplete={(action, onSuccess) => complete.mutate(action, { onSuccess })} />{complete.isError && <Alert tone="danger">{message(complete.error)}</Alert>}</section></Page> }

function Recipes() { const recipes = useQuery({ queryKey: ['recipes'], queryFn: api.recipes, retry: false }); const products = useQuery({ queryKey: productKeys.list, queryFn: api.products.list, retry: false }); const productsById = new Map((products.data ?? []).map((product) => [product.id, product])); return <Page title={t.recipes} kicker={t.recipesKicker} subtitle={t.recipesSubtitle}>{recipes.isPending ? <Skeleton /> : recipes.isError ? <RequestError retry={() => recipes.refetch()} /> : recipes.data.length === 0 ? <EmptyState action={<Link to="/">{t.products}</Link>}>{t.noRecipes}</EmptyState> : <ul className="product-list">{recipes.data.map((recipe) => { const ingredient = productsById.get(recipe.product_ids[0]); return <li className="card" key={recipe.title}><h2>{recipe.title}</h2><p className="muted">{t.recipeUses} {ingredient ? `${ingredient.name} — ${ingredient.expiryDate}` : t.fromInventory}</p></li> })}</ul>}</Page> }
function NotificationSettingForm({ setting }: { setting: NotificationSetting }) {
  const [threshold, setThreshold] = useState(String(setting.alert_threshold_minutes)); const [notice, setNotice] = useState<string>()
  const update = useMutation({ mutationFn: () => api.notifications.save({ ...setting, alert_threshold_minutes: Number(threshold) }), onSuccess: () => setNotice(t.saved) })
  return <form className="form-actions" onSubmit={(event) => { event.preventDefault(); setNotice(undefined); update.mutate() }}><label>{setting.product_group}<input type="number" min="60" value={threshold} onChange={(event) => setThreshold(event.target.value)} /></label><button type="submit" disabled={update.isPending || Number(threshold) < 60}>{update.isPending ? t.savingSettings : t.saveChanges}</button>{update.isError && <Alert tone="danger">{message(update.error)}</Alert>}{notice && <Alert>{notice}</Alert>}</form>
}

function Settings() {
  const profile = useQuery({ queryKey: ['profile'], queryFn: api.profile.get, retry: false }); const settings = useQuery({ queryKey: ['notification-settings'], queryFn: api.notifications.list, retry: false })
  const form = useForm<Pick<Profile, 'country_code' | 'language'>>({ defaultValues: { country_code: '', language: locale } })
  useEffect(() => { if (profile.data) form.reset({ country_code: profile.data.country_code, language: profile.data.language || locale }) }, [form, profile.data])
  const saveProfile = useMutation({ mutationFn: (input: Pick<Profile, 'country_code' | 'language'>) => api.profile.save({ ...input, country_code: input.country_code.toUpperCase() }), onSuccess: () => { profile.refetch() } })
  return <Page title={t.settings} kicker={t.settingsKicker} subtitle={t.settingsSubtitle}><section className="card stack"><h2>{t.profile}</h2>{profile.isPending ? <Skeleton /> : profile.isError ? <RequestError retry={() => profile.refetch()} /> : <form className="form-grid" onSubmit={form.handleSubmit((input) => saveProfile.mutate(input))}><label>{t.country}<input maxLength={2} autoComplete="country" {...form.register('country_code', { required: true, pattern: /^[A-Za-z]{2}$/ })} /></label><label>{t.language}<select {...form.register('language')}><option value="ru">Русский</option><option value="en">English</option></select></label>{profile.data?.regulator_group && <p className="muted">{t.regulator}: {profile.data.regulator_group}</p>}{saveProfile.isError && <Alert tone="danger">{message(saveProfile.error)}</Alert>}<button type="submit" disabled={saveProfile.isPending}>{saveProfile.isPending ? t.savingSettings : t.saveChanges}</button></form>}</section><section className="card stack"><h2>{t.notifications}</h2>{settings.isPending ? <Skeleton /> : settings.isError ? <RequestError retry={() => settings.refetch()} /> : settings.data.settings.map((setting) => <NotificationSettingForm key={setting.product_group} setting={setting} />)}</section><p className="muted">{t.environmentNote}: {apiMode === 'fixture' ? 'fixture' : 'API'}.</p></Page>
}

function Credentials({ register: isRegister }: { register?: boolean }) {
  const navigate = useNavigate(); const location = useLocation(); const queryClient = useQueryClient(); const [error, setError] = useState<string>()
  const form = useForm<{ email: string; password: string }>({ defaultValues: { email: '', password: '' } })
  const errorId = (field: 'email' | 'password') => `credential-${field}-error`
  const action = useMutation({ mutationFn: ({ email, password }: { email: string; password: string }) => isRegister ? api.auth.register(email, password) : api.auth.login(email, password), onSuccess: async () => { if (isRegister) { navigate('/login'); return } await queryClient.invalidateQueries({ queryKey: ['session'] }); navigate((location.state as { from?: string } | null)?.from ?? '/') }, onError: (failure) => setError(message(failure)) })
  const field = (name: 'email' | 'password') => ({ 'aria-invalid': Boolean(form.formState.errors[name]), 'aria-describedby': form.formState.errors[name] ? errorId(name) : undefined })
  const showErrorSummary = Object.keys(form.formState.errors).length >= 2
  const errorSummaryRef = useRef<HTMLDivElement>(null)
  useEffect(() => { if (showErrorSummary) errorSummaryRef.current?.focus() }, [showErrorSummary])
  return <PublicPage variant="auth" title={isRegister ? t.register : t.login} kicker={isRegister ? t.registerKicker : t.loginKicker} subtitle={isRegister ? t.registerSubtitle : t.loginSubtitle}>
    {isRegister && <Alert tone="warning">{t.profileBlocked}</Alert>}
    <form className="form-card auth-form" onSubmit={form.handleSubmit((values) => { setError(undefined); action.mutate(values) })} noValidate>
      {showErrorSummary && <div ref={errorSummaryRef} className="auth-error-summary" role="alert" tabIndex={-1} aria-labelledby="auth-error-summary-title"><h2 id="auth-error-summary-title">{t.formErrorSummary}</h2><a href="#email">{t.invalidEmail}</a><a href="#password">{t.passwordMin}</a></div>}
      <label>{t.email}<input id="email" type="email" autoComplete="email" {...form.register('email', { required: t.enterEmail, pattern: { value: /^\S+@\S+\.\S+$/, message: t.invalidEmail } })} {...field('email')} />{form.formState.errors.email && <span id={errorId('email')} role="alert">{form.formState.errors.email.message}</span>}</label>
      <label>{t.password}<input id="password" type="password" autoComplete={isRegister ? 'new-password' : 'current-password'} {...form.register('password', { required: t.enterPassword, minLength: { value: 8, message: t.passwordMin } })} {...field('password')} />{form.formState.errors.password && <span id={errorId('password')} role="alert">{form.formState.errors.password.message}</span>}</label>
      {error && <Alert tone="danger">{error}</Alert>}
      <button disabled={action.isPending} type="submit">{action.isPending ? t.sending : isRegister ? t.register : t.login}</button>
      <Link to={isRegister ? '/login' : '/register'}>{isRegister ? t.login : t.register}</Link>
    </form>
  </PublicPage>
}

export function App() { return <Routes><Route path="/login" element={<Credentials />} /><Route path="/register" element={<Credentials register />} /><Route element={<ProtectedLayout />}><Route path="/" element={<ProductList />} /><Route path="/products" element={<Navigate to="/" replace />} /><Route path="/products/new" element={<ProductForm title={t.add} />} /><Route path="/products/new/photo" element={<PhotoUpload />} /><Route path="/product-drafts/:id" element={<DraftReview />} /><Route path="/products/:id" element={<ProductDetail />} /><Route path="/recipes" element={<Recipes />} /><Route path="/settings" element={<Settings />} /></Route><Route path="*" element={<PublicPage title={t.notFound}><Link to="/">{t.back}</Link></PublicPage>} /></Routes> }
