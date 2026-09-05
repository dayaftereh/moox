import { ReactNode, useEffect, useRef, useState } from 'react'
import { type GameSection } from '../navigation'
import { type TranslationKey, useI18n } from '../i18n'

const primaryNavItems: Array<{ section: GameSection; label: TranslationKey; glyph: string }> = [
  { section: 'galaxy', label: 'nav.galaxy', glyph: '\u25C8' },
  { section: 'colonies', label: 'nav.colonies', glyph: '\u2302' },
  { section: 'fleets', label: 'nav.fleets', glyph: '\u2197' },
  { section: 'research', label: 'nav.research', glyph: '\u269B' },
  { section: 'diplomacy', label: 'nav.diplomacy', glyph: '\u2696' },
]
type ResourceChip = { label: string; shortLabel?: string; value: string; tone?: 'neutral' | 'warning' | 'danger' }

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
  onEndTurn?: () => void
  endTurnDisabled?: boolean
  endTurnLabel?: string
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

function NavItems({ items, activeSection, onNavigate }: {
  items: typeof primaryNavItems
  activeSection: GameSection
  onNavigate: (section: GameSection) => void
}) {
  const { t } = useI18n()
  return <>
    {items.map((item) => {
      const active = activeSection === item.section
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

export function AppShell({ activeSection, gameID, turn, phaseLabel, status, statusTone, resources = [], onNavigate, onHome, onEndTurn, endTurnDisabled = false, endTurnLabel, children }: AppShellProps) {
  const { t } = useI18n()
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement | null>(null)
  const secondaryActive = activeSection === 'espionage' || activeSection === 'more'

  useEffect(() => {
    if (!menuOpen) return
    const closeOnPointer = (event: PointerEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) setMenuOpen(false)
    }
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setMenuOpen(false)
    }
    window.addEventListener('pointerdown', closeOnPointer)
    window.addEventListener('keydown', closeOnEscape)
    return () => {
      window.removeEventListener('pointerdown', closeOnPointer)
      window.removeEventListener('keydown', closeOnEscape)
    }
  }, [menuOpen])

  function navigateFromMenu(section: GameSection) {
    setMenuOpen(false)
    onNavigate(section)
  }

  return (
    <div className={'game-shell section-' + activeSection}>
      <header className="topbar">
        <div className="main-menu-anchor" ref={menuRef}>
          <button
            type="button"
            className={'main-menu-trigger' + (menuOpen || secondaryActive ? ' active' : '')}
            aria-label={t('gameMenu.open')}
            aria-haspopup="dialog"
            aria-expanded={menuOpen}
            title={t('gameMenu.open')}
            onClick={() => setMenuOpen((open) => !open)}
          >
            <span aria-hidden="true">⋯</span>
          </button>
          {menuOpen && (
            <section className="main-menu-popover" role="dialog" aria-label={t('gameMenu.title')} data-game-id={gameID}>
              <header className="main-menu-header">
                <div><p className="eyebrow">{t('gameMenu.title')}</p></div>
                <button type="button" className="button-ghost main-menu-close" onClick={() => setMenuOpen(false)} aria-label={t('common.close')}>×</button>
              </header>

              <div className="main-menu-section">
                <span className="main-menu-section-title">{t('gameMenu.settings')}</span>
                <div className="main-menu-setting-row">
                  <span>{t('gameMenu.language')}</span>
                  <LanguageSwitch compact />
                </div>
              </div>

              <div className="main-menu-section">
                <span className="main-menu-section-title">{t('gameMenu.game')}</span>
                <button type="button" className="main-menu-item" onClick={() => navigateFromMenu('espionage')}>
                  <span className="main-menu-item-glyph" aria-hidden="true">◉</span>
                  <span><strong>{t('nav.espionage')}</strong><small>{t('gameMenu.espionageHint')}</small></span>
                </button>
                <button type="button" className="main-menu-item" onClick={() => navigateFromMenu('more')}>
                  <span className="main-menu-item-glyph" aria-hidden="true">⚙</span>
                  <span><strong>{t('gameMenu.advanced')}</strong><small>{t('gameMenu.advancedHint')}</small></span>
                </button>
              </div>

              <div className="main-menu-section main-menu-section-last">
                <button type="button" className="main-menu-item main-menu-home" onClick={() => { setMenuOpen(false); onHome() }}>
                  <span className="main-menu-item-glyph" aria-hidden="true">↩</span>
                  <span><strong>{t('gameMenu.mainMenu')}</strong><small>{t('gameMenu.mainMenuHint')}</small></span>
                </button>
              </div>
            </section>
          )}
        </div>

        <div className="topbar-resources" aria-label={t('a11y.resources')}>
          {resources.map((resource) => (
            <div className={`topbar-resource resource-${resource.tone ?? 'neutral'}`} key={resource.label} title={`${resource.label}: ${resource.value}`}>
              <span className="resource-label resource-label-long">{resource.label}</span>
              <span className="resource-label resource-label-short">{resource.shortLabel ?? resource.label}</span>
              <strong>{resource.value}</strong>
            </div>
          ))}
        </div>

        <div className="topbar-context" aria-label={t('a11y.gameStatus')}>
          {turn !== undefined && <span className="status-chip">{t('top.turn', { turn })}</span>}
          {phaseLabel && <span className="status-chip status-chip-phase">{phaseLabel}</span>}
        </div>
        <div className="topbar-live" aria-live="polite" title={status} aria-label={status}>
          <span className={`connection-dot connection-${statusTone}`} aria-hidden="true" />
        </div>
      </header>

      <div className="shell-layout">
        <aside className="side-nav" aria-label={t('a11y.primaryNavigation')}>
          <div className="side-nav-items"><NavItems items={primaryNavItems} activeSection={activeSection} onNavigate={onNavigate} /></div>
        </aside>
        <main className="game-content">{children}</main>
      </div>

      <div className="bottom-command-bar">
        <nav className="bottom-nav" aria-label={t('a11y.primaryNavigation')}>
          <NavItems items={primaryNavItems} activeSection={activeSection} onNavigate={onNavigate} />
        </nav>
        <button
          type="button"
          className="button-primary bottom-end-turn"
          disabled={!onEndTurn || endTurnDisabled}
          onClick={onEndTurn}
        >
          {endTurnLabel ?? t('planning.endTurn')}
        </button>
      </div>
    </div>
  )
}export { LanguageSwitch }
