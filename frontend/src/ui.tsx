import { forwardRef, type ReactNode, type RefObject, useEffect, useId, useMemo, useRef, useState } from 'react'
import { Link, NavLink, useLocation, useNavigate } from 'react-router-dom'
import { AlertCircle, AlertTriangle, Calendar as CalendarIcon, ChefHat, ChevronLeft, ChevronRight, FileText, Info, Package, Camera, Plus, Settings as SettingsIcon, X } from 'lucide-react'
import { locale, statusLabel as localizedStatusLabel, t } from './i18n'

export type Status = 'active' | 'attention' | 'expired' | 'used' | 'discarded' | 'research_required'
const badgeTone: Record<Status, 'active' | 'attention' | 'expired' | 'neutral' | 'info'> = {
    active: 'active',
    attention: 'attention',
    expired: 'expired',
    used: 'neutral',
    discarded: 'neutral',
    research_required: 'info',
}

export const statusLabel = (status: Status) => localizedStatusLabel(status)

export function StatusBadge({ status }: { status: Status }) {
    return <span className={`badge badge--${badgeTone[status]}`}><i className="badge__dot" aria-hidden="true" />{statusLabel(status)}</span>
}

const alertIcon = { info: Info, warning: AlertTriangle, danger: AlertCircle }

export function Alert({ tone = 'info', children }: { tone?: 'info' | 'warning' | 'danger'; children: ReactNode }) {
    const Icon = alertIcon[tone]
    return <div className={`alert alert--${tone}`} role={tone === 'danger' ? 'alert' : 'status'}><Icon className="alert__icon" aria-hidden="true" size={18} /><span>{children}</span></div>
}

export function Skeleton() {
    return <div className="skeleton" aria-busy="true" aria-label={t.loading} />
}

export function EmptyState({ children, action }: { children: ReactNode; action?: ReactNode }) {
    return <section className="empty-state"><span className="empty-icon" aria-hidden="true"><Package size={22} /></span><p>{children}</p>{action}</section>
}

// Decorative section transition echoing the wavy dashed dividers from the
// design reference. Purely presentational (aria-hidden); `to` names the CSS
// custom property the wave reveals underneath the band it sits on.
export function WaveDivider({ to = '--canvas' }: { to?: string }) {
    return (
        <svg className="wave-divider" viewBox="0 0 1440 60" preserveAspectRatio="none" aria-hidden="true" focusable="false">
            <path d="M0,24 C120,50 240,-4 360,24 C480,52 600,-4 720,24 C840,52 960,-4 1080,24 C1200,52 1320,-4 1440,24 L1440,60 L0,60 Z" fill={`var(${to})`} />
            <path d="M0,24 C120,50 240,-4 360,24 C480,52 600,-4 720,24 C840,52 960,-4 1080,24 C1200,52 1320,-4 1440,24" fill="none" stroke="var(--brand-strong)" strokeOpacity="0.55" strokeWidth="2.5" strokeDasharray="8 7" strokeLinecap="round" />
        </svg>
    )
}

const localeTag = locale === 'ru' ? 'ru-RU' : 'en-US'
const pad2 = (value: number) => String(value).padStart(2, '0')
const toISODate = (date: Date) => `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())}`
const parseISODate = (value: string): Date | null => {
    const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
    if (!match) return null
    const date = new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]))
    return Number.isNaN(date.getTime()) ? null : date
}
const isSameDay = (a: Date, b: Date) => a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
// Monday-first grid regardless of locale, matching the rest of the app's fixed layout choices.
const weekdayLabels = Array.from({ length: 7 }, (_, index) => new Intl.DateTimeFormat(localeTag, { weekday: 'short' }).format(new Date(2024, 0, index + 1)))

/** A self-contained popover calendar replacing the native date input, styled to match the app's editorial theme. */
export function DatePicker({ id, label, value, onChange, invalid, describedBy }: {
    id?: string;
    label: string;
    value: string;
    onChange: (value: string) => void;
    invalid?: boolean;
    describedBy?: string
}) {
    const [open, setOpen] = useState(false)
    const selected = useMemo(() => parseISODate(value), [value])
    const [viewDate, setViewDate] = useState(() => selected ?? new Date())
    const containerRef = useRef<HTMLDivElement>(null)
    const triggerRef = useRef<HTMLButtonElement>(null)

    useEffect(() => { if (selected) setViewDate(selected) }, [selected])

    useEffect(() => {
        if (!open) return
        const onOutsideClick = (event: MouseEvent) => {
            if (containerRef.current && !containerRef.current.contains(event.target as Node)) setOpen(false)
        }
        const onKeyDown = (event: KeyboardEvent) => {
            if (event.key === 'Escape') { setOpen(false); triggerRef.current?.focus() }
        }
        document.addEventListener('mousedown', onOutsideClick)
        document.addEventListener('keydown', onKeyDown)
        return () => {
            document.removeEventListener('mousedown', onOutsideClick)
            document.removeEventListener('keydown', onKeyDown)
        }
    }, [open])

    const year = viewDate.getFullYear()
    const month = viewDate.getMonth()
    const startOffset = (new Date(year, month, 1).getDay() + 6) % 7
    const daysInMonth = new Date(year, month + 1, 0).getDate()
    const cells: (Date | null)[] = [...Array(startOffset).fill(null), ...Array.from({ length: daysInMonth }, (_, index) => new Date(year, month, index + 1))]
    const today = new Date()

    const pick = (date: Date) => { onChange(toISODate(date)); setOpen(false); triggerRef.current?.focus() }
    const display = selected ? new Intl.DateTimeFormat(localeTag, { day: '2-digit', month: 'long', year: 'numeric' }).format(selected) : t.pickDate
    // aria-label (not the wrapping <label>) pins the accessible name so it can't pick up
    // the popover's own text/aria-label once it renders as a sibling inside that label.
    const accessibleLabel = selected ? `${label}: ${display}` : label

    return (
        <div className="date-picker" ref={containerRef}>
            <button ref={triggerRef} type="button" id={id} className="date-picker-trigger" aria-haspopup="dialog"
                    aria-label={accessibleLabel} aria-expanded={open} aria-invalid={invalid} aria-describedby={describedBy}
                    onClick={() => setOpen((current) => !current)}>
                <CalendarIcon size={16} aria-hidden="true" />
                <span className={selected ? undefined : 'muted'}>{display}</span>
            </button>
            {open && (
                <div className="date-picker-panel" role="dialog" aria-modal="true" aria-label={t.calendar}>
                    <div className="date-picker-head">
                        <IconButton label={t.prevMonth} className="date-picker-nav" onClick={() => setViewDate(new Date(year, month - 1, 1))}><ChevronLeft size={16} /></IconButton>
                        <span className="date-picker-month">{new Intl.DateTimeFormat(localeTag, { month: 'long', year: 'numeric' }).format(viewDate)}</span>
                        <IconButton label={t.nextMonth} className="date-picker-nav" onClick={() => setViewDate(new Date(year, month + 1, 1))}><ChevronRight size={16} /></IconButton>
                    </div>
                    <div className="date-picker-weekdays" aria-hidden="true">{weekdayLabels.map((label, index) => <span key={index}>{label}</span>)}</div>
                    <div className="date-picker-grid">
                        {cells.map((date, index) => date
                            ? <button type="button" key={date.getTime()}
                                      className={`date-picker-day${selected && isSameDay(date, selected) ? ' is-selected' : ''}${isSameDay(date, today) ? ' is-today' : ''}`}
                                      onClick={() => pick(date)}>{date.getDate()}</button>
                            : <span key={`empty-${index}`} aria-hidden="true" />)}
                    </div>
                </div>
            )}
        </div>
    )
}

export const IconButton = forwardRef<HTMLButtonElement, {
    label: string;
    onClick?: () => void;
    className?: string;
    expanded?: boolean;
    children: ReactNode
}>(
    ({ label, onClick, className, expanded, children }, ref) => (
        <button ref={ref} type="button" className={className} aria-label={label} aria-expanded={expanded}
                onClick={onClick}>{children}</button>
    ),
)
IconButton.displayName = 'IconButton'

const focusableSelector = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

export function trapDialogTabKey(event: KeyboardEvent, root: HTMLElement | null) {
    if (event.key !== 'Tab') return
    const focusable = Array.from(root?.querySelectorAll<HTMLElement>(focusableSelector) ?? [])
    if (focusable.length === 0) { event.preventDefault(); return }
    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus() }
    if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus() }
}

export function AddProductSheet({ open, onClose, returnFocusRef }: { open: boolean; onClose: () => void; returnFocusRef: RefObject<HTMLButtonElement | null> }) {
    const dialogRef = useRef<HTMLElement>(null)
    const closeRef = useRef<HTMLButtonElement>(null)
    const titleId = useId()
    useEffect(() => {
        if (!open) return
        closeRef.current?.focus()
        const close = (event: KeyboardEvent) => {
            if (event.key === 'Escape') onClose()
            trapDialogTabKey(event, dialogRef.current)
        }
        document.addEventListener('keydown', close)
        return () => { document.removeEventListener('keydown', close); returnFocusRef.current?.focus() }
    }, [open, onClose, returnFocusRef])
    if (!open) return null
    return (
        <div className="sheet-backdrop" onMouseDown={onClose}>
            <section ref={dialogRef} className="add-sheet" role="dialog" aria-modal="true" aria-labelledby={titleId}
                      onMouseDown={(event) => event.stopPropagation()}>
                <div className="sheet-handle" aria-hidden="true" />
                <div className="sheet-head">
                    <div><h2 id={titleId}>{t.add}</h2><p>{t.chooseAddMethod}</p></div>
                    <IconButton ref={closeRef} label={t.close} className="sheet-close" onClick={onClose}><X size={18} /></IconButton>
                </div>
                <Link className="add-option" to="/products/new" onClick={onClose}>
                    <span className="option-icon"><FileText size={20} /></span>
                    <span className="option-copy"><b>{t.manual}</b><span>{t.manualHint}</span></span>
                    <ChevronRight className="arrow" aria-hidden="true" size={20} />
                </Link>
                <Link className="add-option" to="/products/new/photo" onClick={onClose}>
                    <span className="option-icon"><Camera size={20} /></span>
                    <span className="option-copy"><b>{t.photo}</b><span>{t.photoHint}</span></span>
                    <ChevronRight className="arrow" aria-hidden="true" size={20} />
                </Link>
            </section>
        </div>
    )
}

export type Toast = { id: string; message: string }

export function ToastRegion({ toasts, onDismiss }: { toasts: Toast[]; onDismiss: (id: string) => void }) {
    if (toasts.length === 0) return null
    return (
        <div className="toast-region" role="status" aria-live="polite">
            {toasts.map((toast) => (
                <div className="toast" key={toast.id}>
                    <span>{toast.message}</span>
                    <IconButton label={t.close} onClick={() => onDismiss(toast.id)}><X size={14} /></IconButton>
                </div>
            ))}
        </div>
    )
}

function useIsMobile() {
    const [isMobile, setIsMobile] = useState(() => typeof window !== 'undefined' ? window.innerWidth < 640 : false)
    useEffect(() => {
        const onResize = () => setIsMobile(window.innerWidth < 640)
        window.addEventListener('resize', onResize)
        return () => window.removeEventListener('resize', onResize)
    }, [])
    return isMobile
}

export function AppShell({ children, railFooter, addOpen, onOpenAdd, onCloseAdd }: {
    children: ReactNode;
    railFooter?: ReactNode;
    addOpen: boolean;
    onOpenAdd: () => void;
    onCloseAdd: () => void
}) {
    const isMobile = useIsMobile()
    const trigger = useRef<HTMLButtonElement>(null)
    const location = useLocation()
    const navigate = useNavigate()
    const [toasts, setToasts] = useState<Toast[]>([])
    const crumb = location.pathname === '/recipes' ? t.recipes : location.pathname === '/settings' ? t.settings : t.products
    useEffect(() => {
        const notice = (location.state as { notice?: string } | null)?.notice
        if (!notice) return
        const id = globalThis.crypto?.randomUUID?.() ?? `toast-${Date.now()}`
        setToasts((current) => [...current, { id, message: notice }])
        navigate(location.pathname, { replace: true, state: null })
        const timer = window.setTimeout(() => setToasts((current) => current.filter((toast) => toast.id !== id)), 5_000)
        return () => window.clearTimeout(timer)
    }, [location.pathname, location.state, navigate])
    return (
        <>
            <a className="skip" href="#main-content">Skip to content</a>
            <div className="shell">
                <div className="rail">
                    <Link className="brand" to="/"><span>pantry<br />ledger</span></Link>
                    <span className="tagline">{t.sidebarTagline}</span>
                    <nav className="nav" aria-label={t.navigation}>
                        <NavLink to="/" end><Package aria-hidden="true" /><span>{t.products}</span></NavLink>
                        <NavLink to="/recipes"><ChefHat aria-hidden="true" /><span>{t.recipes}</span></NavLink>
                        {isMobile && <IconButton ref={trigger} label={t.add} className="nav-fab" expanded={addOpen} onClick={onOpenAdd}><Plus aria-hidden="true" /></IconButton>}
                        <NavLink to="/settings"><SettingsIcon aria-hidden="true" /><span>{t.settings}</span></NavLink>
                    </nav>
                    {!isMobile && (
                        <div className="rail-footer">
                            <div className="help-tip"><b>{t.helpTitle}</b>{t.helpText}</div>
                            {railFooter}
                        </div>
                    )}
                </div>
                <main id="main-content" className="page">
                    {isMobile && railFooter && <div className="mobile-account">{railFooter}</div>}
                    {!isMobile && (
                        <div className="workspace-toolbar">
                            <span className="crumb">{crumb}</span>
                            <IconButton ref={trigger} label={t.add} className="add" expanded={addOpen} onClick={onOpenAdd}><Plus aria-hidden="true" size={18} /><span>{t.add}</span></IconButton>
                        </div>
                    )}
                    {children}
                </main>
            </div>
            <AddProductSheet open={addOpen} onClose={onCloseAdd} returnFocusRef={trigger} />
            <ToastRegion toasts={toasts} onDismiss={(id) => setToasts((current) => current.filter((toast) => toast.id !== id))} />
        </>
    )
}
