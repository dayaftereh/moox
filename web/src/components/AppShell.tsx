import { ReactNode } from 'react'
import { type GameSection } from '../navigation'
import { type TranslationKey, useI18n } from '../i18n'

const desktopNavItems: Array<{ section: GameSection; label: TranslationKey; glyph: string }> = [
  { section: 'galaxy', label: 'nav.galaxy', glyph: '\u25C8' },
  { section: 'colonies', label: 'nav.colonies', glyph: '\u2302' },
  { section: 'fleets', label: 'nav.fleets', glyph: '\u2197' },
  { section: 'research', label: 'nav.research', glyph: '\u269B' },
  { section: 'diplomacy', label: 'nav.diplomacy', glyph: '\u2696' },
  { section: 'espionage', label: 'nav.espionage', glyph: '\u25C9' },
  { section: 'more', label: 'nav.more', glyph: '\u2026' },
]

const mobileNavItems = desktopNavItems.filter((item) => ['galaxy', 'colonies', 'fleets', 'research', 'more'].includes(item.section))

type ResourceChip = { label: string; value: string; tone?: 'neutral' | 'warning' | 'danger' }

type AppShellProps = {
  activeSection: GameSection
  gameID: string
  turn?: number
  phaseLabel?: string
  status: string
  statusTone: 'neutral' | 'success' | 'warning' | 'danger'
  resources?: ResourceChip[]
  onNavigate: (section: GameSection) => void
  onHome: () => void
  children: ReactNode
}

function LanguageSwitch({ compact = false }: { compact?: boolean }) {
  const { locale, setLocale, t } = useI18n()
  return (
    <div className={`language-switch ${compact ? 'language-switch-compact' : ''}`} role="group" aria-label={t('a11y.language')}>
      <button type="button" className={locale === 'de' ? 'active' : ''} aria-pressed={locale === 'de'} onClick={() => setLocale('de')}>DE</button>
      <button type="button" className={locale === 'en' ? 'active' : ''} aria-pressed={locale === 'en'} onClick={() => setLocale('en')}>EN</button>
    </div>
  )
}

export function StandaloneHeader({ onHome }: { onHome?: () => void }) {
  const { t } = useI18n()
  return (
    <header className="standalone-header">
      <button type="button" className="brand-button" onClick={onHome} disabled={!onHome}>
        <span className="brand-mark" aria-hidden="true">OX</span>
        <span><strong>{t('app.name')}</strong><small>{t('app.tagline')}</small></span>
      </button>
      <LanguageSwitch compact />
    </header>
  )
}

function NavItems({ items, activeSection, onNavigate, mobile = false }: {
  items: typeof desktopNavItems
  activeSection: GameSection
  onNavigate: (section: GameSection) => void
  mobile?: boolean
}) {
  const { t } = useI18n()
  return <>
    {items.map((item) => {
      const active = mobile && item.section === 'more'
        ? ['more', 'diplomacy', 'espionage'].includes(activeSection)
        : activeSection === item.section
      return (
        <button
          type="button"
          key={item.section}
          className={`nav-item ${active ? 'active' : ''}`}
          aria-current={active ? 'page' : undefined}
          onClick={() => onNavigate(item.section)}
        >
          <span className="nav-glyph" aria-hidden="true">{item.glyph}</span>
          <span>{t(item.label)}</span>
        </button>
      )
    })}
  </>
}

export function AppShell({ activeSection, gameID, turn, phaseLabel, status, statusTone, resources = [], onNavigate, onHome, children }: AppShellProps) {
  const { t } = useI18n()
  return (
    <div className="game-shell">
      <header className="topbar">
        <button type="button" className="brand-button topbar-brand" onClick={onHome}>
          <span className="brand-mark" aria-hidden="true">OX</span>
          <span className="brand-copy"><strong>{t('app.name')}</strong><small>{gameID}</small></span>
        </button>
        <div className="topbar-context" aria-label={t('a11y.gameStatus')}>
          {turn !== undefined && <span className="status-chip">{t('top.turn', { turn })}</span>}
          {phaseLabel && <span className="status-chip status-chip-phase">{phaseLabel}</span>}
        </div>
        <div className="topbar-live" aria-live="polite">
          <span className={`connection-dot connection-${statusTone}`} aria-hidden="true" />
          <span className="connection-copy">{status}</span>
        </div>
        <LanguageSwitch compact />
      </header>

      {resources.length > 0 && (
        <div className="resource-strip" aria-label={t('a11y.resources')}>
          {resources.map((resource) => (
            <div className={`resource-chip resource-${resource.tone ?? 'neutral'}`} key={resource.label}>
              <span>{resource.label}</span><strong>{resource.value}</strong>
            </div>
          ))}
        </div>
      )}

      <div className="shell-layout">
        <aside className="side-nav" aria-label={t('a11y.primaryNavigation')}>
          <div className="side-nav-items"><NavItems items={desktopNavItems} activeSection={activeSection} onNavigate={onNavigate} /></div>
          <div className="side-nav-footer"><LanguageSwitch /></div>
        </aside>
        <main className="game-content">{children}</main>
      </div>

      <nav className="bottom-nav" aria-label={t('a11y.primaryNavigation')}>
        <NavItems items={mobileNavItems} activeSection={activeSection} onNavigate={onNavigate} mobile />
      </nav>
    </div>
  )
}

export { LanguageSwitch }
