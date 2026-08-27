import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import { t } from './i18n'
export type Status = 'active'|'attention'|'expired'|'used'|'discarded'|'research_required'
const labels: Record<Status,string>={active:'Активен',attention:'Требует внимания',expired:'Срок истёк',used:'Использован',discarded:'Выброшен',research_required:'Правило уточняется'}
const statusMarks: Record<Status,string>={active:'✓',attention:'!',expired:'×',used:'✓',discarded:'–',research_required:'i'}
export function StatusBadge({status}:{status:Status}){return <span className={`badge badge--${status}`}><span className="badge__mark" aria-hidden="true">{statusMarks[status]}</span>{labels[status]}</span>}
export function Alert({tone='info',children}:{tone?:'info'|'warning'|'danger';children:ReactNode}){return <div className={`alert alert--${tone}`} role={tone==='danger'?'alert':'status'}>{children}</div>}
export function Skeleton(){return <div className="skeleton" aria-busy="true" aria-label={t.loading}/>} export function EmptyState({children,action}:{children:ReactNode;action?:ReactNode}){return <section className="empty-state"><p>{children}</p>{action}</section>}
export function AppShell({children}:{children:ReactNode}){return <><a className="skip" href="#main-content">Skip to content</a><div className="shell"><aside className="brand-rail"><a className="brand" href="/products"><span className="brand__mark" aria-hidden="true">◒</span><span>pantry<br/>ledger</span></a><p>Домашний запас<br/>без лишнего шума</p></aside><nav className="nav" aria-label={t.navigation}><NavLink to="/products"><span aria-hidden="true">⌂</span>{t.products}</NavLink><NavLink to="/recipes"><span aria-hidden="true">♧</span>{t.recipes}</NavLink><NavLink to="/settings"><span aria-hidden="true">◌</span>{t.settings}</NavLink></nav><main id="main-content" className="page">{children}</main></div></>}
