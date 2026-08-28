import { forwardRef, type ReactNode, useEffect, useRef, useState } from 'react'
import { Link, NavLink, useLocation } from 'react-router-dom'
import { ChefHat, ChevronRight, FileText, Package, Camera, Plus, Settings as SettingsIcon, X } from 'lucide-react'
import { t } from './i18n'

export type Status = 'active' | 'attention' | 'expired' | 'used' | 'discarded' | 'research_required'
const labels: Record<Status, string> = {
    active: 'Активен',
    attention: 'Требует внимания',
    expired: 'Срок истёк',
    used: 'Использован',
    discarded: 'Выброшен',
    research_required: 'Правило уточняется',
}
const badgeTone: Record<Status, 'active' | 'attention' | 'expired' | 'neutral' | 'info'> = {
    active: 'active',
    attention: 'attention',
    expired: 'expired',
    used: 'neutral',
    discarded: 'neutral',
    research_required: 'info',
}

export const statusLabel = (status: Status) => labels[status]

export function StatusBadge({ status }: { status: Status }) {
    return <span className={`badge badge--${badgeTone[status]}`}><i className="badge__dot" aria-hidden="true" />{labels[status]}</span>
}

export function Alert({ tone = 'info', children }: { tone?: 'info' | 'warning' | 'danger'; children: ReactNode }) {
    return <div className={`alert alert--${tone}`} role={tone === 'danger' ? 'alert' : 'status'}>{children}</div>
}

export function Skeleton() {
    return <div className="skeleton" aria-busy="true" aria-label={t.loading} />
}

export function EmptyState({ children, action }: { children: ReactNode; action?: ReactNode }) {
    return <section className="empty-state"><p>{children}</p>{action}</section>
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

export function AddProductSheet({ open, onClose }: { open: boolean; onClose: () => void }) {
    useEffect(() => {
        if (!open) return
        const close = (event: KeyboardEvent) => {
            if (event.key === 'Escape') onClose()
        }
        document.addEventListener('keydown', close)
        return () => document.removeEventListener('keydown', close)
    }, [open, onClose])
    if (!open) return null
    return (
        <div className="sheet-backdrop" onMouseDown={onClose}>
            <section className="add-sheet" role="dialog" aria-modal="true" aria-label={t.add}
                      onMouseDown={(event) => event.stopPropagation()}>
                <div className="sheet-handle" aria-hidden="true" />
                <div className="sheet-head">
                    <div><h2>{t.add}</h2><p>{t.chooseAddMethod}</p></div>
                    <IconButton label={t.close} className="sheet-close" onClick={onClose}><X size={18} /></IconButton>
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
    const crumb = location.pathname === '/recipes' ? t.recipes : location.pathname === '/settings' ? t.settings : t.products
    useEffect(() => {
        if (!addOpen) trigger.current?.focus()
    }, [addOpen])
    return (
        <>
            <a className="skip" href="#main-content">Skip to content</a>
            <div className="shell">
                <div className="rail">
                    <Link className="brand" to="/"><span>pantry<br />ledger</span></Link>
                    <nav className="nav" aria-label={t.navigation}>
                        <NavLink to="/" end><Package aria-hidden="true" /><span>{t.products}</span></NavLink>
                        <NavLink to="/recipes"><ChefHat aria-hidden="true" /><span>{t.recipes}</span></NavLink>
                        {isMobile && <IconButton ref={trigger} label={t.add} className="nav-fab" expanded={addOpen} onClick={onOpenAdd}><Plus aria-hidden="true" /></IconButton>}
                        <NavLink to="/settings"><SettingsIcon aria-hidden="true" /><span>{t.settings}</span></NavLink>
                    </nav>
                    {!isMobile && railFooter && <div className="rail-footer">{railFooter}</div>}
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
            <AddProductSheet open={addOpen} onClose={onCloseAdd} />
        </>
    )
}
