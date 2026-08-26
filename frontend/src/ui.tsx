import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import { t } from './i18n'
export type Status = 'active'|'attention'|'expired'|'used'|'discarded'|'research_required'
const labels: Record<Status,string>={active:'Активен',attention:'Требует внимания',expired:'Срок истёк',used:'Использован',discarded:'Выброшен',research_required:'Правило уточняется'}
export function StatusBadge({status}:{status:Status}){const tone=status==='expired'?'danger':status==='attention'?'warning':status==='research_required'?'info':'';return <span className={`badge ${tone}`}><span aria-hidden="true">{status==='expired'?'!':status==='attention'?'!':'●'}</span>{labels[status]}</span>}
export function Alert({tone='info',children}:{tone?:'info'|'warning'|'danger';children:ReactNode}){return <div className={`alert ${tone}`} role={tone==='danger'?'alert':'status'}>{children}</div>}
export function Skeleton(){return <div className="skeleton" aria-busy="true" aria-label={t.loading}/>} export function EmptyState({children,action}:{children:ReactNode;action?:ReactNode}){return <section className="card stack"><p>{children}</p>{action}</section>}
export function AppShell({children}:{children:ReactNode}){return <><a className="skip" href="#main-content">Skip to content</a><div className="shell"><nav className="nav" aria-label="Navigation"><NavLink to="/products">{t.products}</NavLink><NavLink to="/recipes">{t.recipes}</NavLink><NavLink to="/settings">{t.settings}</NavLink></nav><main id="main-content" className="page">{children}</main></div></>}
