import {forwardRef, type ReactNode, useEffect, useRef, useState} from 'react'
import {Link, NavLink} from 'react-router-dom'
import {t} from './i18n'

export type Status = 'active' | 'attention' | 'expired' | 'used' | 'discarded' | 'research_required'
const labels: Record<Status, string> = {
    active: 'Активен',
    attention: 'Требует внимания',
    expired: 'Срок истёк',
    used: 'Использован',
    discarded: 'Выброшен',
    research_required: 'Правило уточняется'
}
const statusMarks: Record<Status, string> = {
    active: '✓',
    attention: '!',
    expired: '×',
    used: '✓',
    discarded: '–',
    research_required: 'i'
}

export function StatusBadge({status}: { status: Status }) {
    return <span className={`badge badge--${status}`}><span className="badge__mark"
                                                            aria-hidden="true">{statusMarks[status]}</span>{labels[status]}</span>
}

export function Alert({tone = 'info', children}: { tone?: 'info' | 'warning' | 'danger'; children: ReactNode }) {
    return <div className={`alert alert--${tone}`} role={tone === 'danger' ? 'alert' : 'status'}>{children}</div>
}

export function Skeleton() {
    return <div className="skeleton" aria-busy="true" aria-label={t.loading}/>
}

export function EmptyState({children, action}: { children: ReactNode; action?: ReactNode }) {
    return <section className="empty-state"><p>{children}</p>{action}</section>
}

export function Icon({children}: { children: ReactNode }) {
    return <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.8}
                strokeLinecap="round" strokeLinejoin="round">{children}</svg>
}

export const IconButton = forwardRef<HTMLButtonElement, {
    label: string;
    onClick?: () => void;
    className?: string;
    expanded?: boolean;
    children: ReactNode
}>(
    ({label, onClick, className, expanded, children}, ref) => (
        <button ref={ref} type="button" className={className} aria-label={label} aria-expanded={expanded}
                onClick={onClick}>{children}</button>
    ),
)
IconButton.displayName = 'IconButton'

export function AddProductSheet({open, onClose}: { open: boolean; onClose: () => void }) {
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
                <IconButton label={t.close} className="sheet-close" onClick={onClose}><Icon>
                    <path d="M6 6l12 12M18 6L6 18"/>
                </Icon></IconButton>
                <h2>{t.add}</h2>
                <p>{t.chooseAddMethod}</p>
                <Link className="sheet-action" to="/products/new" onClick={onClose}>{t.manual}</Link>
                <Link className="sheet-action" to="/products/new/photo" onClick={onClose}>{t.photo}</Link>
            </section>
        </div>
    )
}

export type Toast = { id: string; message: string }

export function ToastRegion({toasts, onDismiss}: { toasts: Toast[]; onDismiss: (id: string) => void }) {
    if (toasts.length === 0) return null
    return (
        <div className="toast-region" role="status" aria-live="polite">
            {toasts.map((toast) => (
                <div className="toast" key={toast.id}>
                    <span>{toast.message}</span>
                    <IconButton label={t.close} onClick={() => onDismiss(toast.id)}><Icon>
                        <path d="M6 6l12 12M18 6L6 18"/>
                    </Icon></IconButton>
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

export function AppShell({children, railFooter, addOpen, onOpenAdd, onCloseAdd}: {
    children: ReactNode;
    railFooter?: ReactNode;
    addOpen: boolean;
    onOpenAdd: () => void;
    onCloseAdd: () => void
}) {
    const isMobile = useIsMobile()
    const trigger = useRef<HTMLButtonElement>(null)
    useEffect(() => {
        if (!addOpen) trigger.current?.focus()
    }, [addOpen])
    const addIcon = <Icon><path d="M12 5v14M5 12h14"/></Icon>
    return (
        <>
            <a className="skip" href="#main-content">Skip to content</a>
            <div className="shell">
                <div className="rail">
                    <Link className="brand" to="/products"><span>pantry<br/>ledger</span></Link>
                    <nav className="nav" aria-label={t.navigation}>
                        <NavLink to="/products"><Icon>
                            <path d="M4 5h16v14H4zM4 10h16"/>
                        </Icon><span>{t.products}</span></NavLink>
                        <NavLink to="/recipes"><Icon>
                            <path d="M7 3v7a4 4 0 0 0 4 4h2a4 4 0 0 0 4-4V3M9 3v5m6-5v5M12 14v7"/>
                        </Icon><span>{t.recipes}</span></NavLink>
                        {isMobile && <IconButton ref={trigger} label={t.add} className="nav-fab" expanded={addOpen} onClick={onOpenAdd}>{addIcon}</IconButton>}
                        <NavLink to="/settings"><Icon>
                            <circle cx="12" cy="12" r="3"/>
                            <path d="M12 2v3m0 14v3M2 12h3m14 0h3"/>
                        </Icon><span>{t.settings}</span></NavLink>
                    </nav>
                    {!isMobile && railFooter && <div className="rail-footer">{railFooter}</div>}
                </div>
                <main id="main-content" className="page">
                    {isMobile && railFooter && <div className="mobile-account">{railFooter}</div>}
                    {!isMobile && (
                        <div className="workspace-toolbar">
                            <IconButton ref={trigger} label={t.add} className="add" expanded={addOpen} onClick={onOpenAdd}>{addIcon}<span>{t.add}</span></IconButton>
                        </div>
                    )}
                    {children}
                </main>
            </div>
            <AddProductSheet open={addOpen} onClose={onCloseAdd}/>
        </>
    )
}
