import { ReactNode } from 'react'

export function Card({ children, className = '', as: Element = 'section' }: { children: ReactNode; className?: string; as?: 'section' | 'article' | 'div' }) {
  return <Element className={`card ${className}`.trim()}>{children}</Element>
}

export function PageHeader({ eyebrow, title, subtitle, actions }: { eyebrow: string; title: string; subtitle?: string; actions?: ReactNode }) {
  return (
    <header className="page-header">
      <div>
        <p className="eyebrow">{eyebrow}</p>
        <h1>{title}</h1>
        {subtitle && <p className="page-subtitle">{subtitle}</p>}
      </div>
      {actions && <div className="page-actions">{actions}</div>}
    </header>
  )
}

export function Metric({ label, value }: { label: string; value: ReactNode }) {
  return (
    <article className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  )
}

export function Notice({ title, children, tone = 'info' }: { title: string; children?: ReactNode; tone?: 'info' | 'warning' | 'danger' | 'success' }) {
  return (
    <section className={`notice notice-${tone}`} role={tone === 'danger' ? 'alert' : 'status'}>
      <strong>{title}</strong>
      {children && <div className="notice-body">{children}</div>}
    </section>
  )
}

export function EmptyState({ title, body }: { title: string; body?: string }) {
  return (
    <Card className="empty-state">
      <div className="empty-orbit" aria-hidden="true"><span /></div>
      <h2>{title}</h2>
      {body && <p className="muted">{body}</p>}
    </Card>
  )
}
