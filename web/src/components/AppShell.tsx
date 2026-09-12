import { ReactNode, useEffect, useRef, useState } from 'react'
import { type GameSection } from '../navigation'
import { type TranslationKey, useI18n } from '../i18n'
import { ProceduralShipGlyph } from './ProceduralShipGlyph'
import { GameIcon, type GameIconName } from './GameIcon'

const primaryNavItems: Array<{ section: GameSection; label: TranslationKey; icon: GameIconName }> = [
  { section: 'galaxy', label: 'nav.galaxy', icon: 'galaxy' },
  { section: 'colonies', label: 'nav.colonies', icon: 'colonies' },
  { section: 'fleets', label: 'nav.fleets', icon: 'fleets' },
  { section: 'diplomacy', label: 'nav.diplomacy', icon: 'diplomacy' },
  { section: 'espionage', label: 'nav.espionage', icon: 'espionage' },
]
type ResourceTone = 'neutral' | 'positive' | 'warning' | 'danger'
export type ResourceChip = {
  id: string
  label: string
  shortLabel?: string
  icon: GameIconName
  value?: string
  delta?: string
  deltaTone?: ResourceTone
  tone?: ResourceTone
  progressPercent?: number
  detailTitle: string
  details: Array<{ label: string; value: string; tone?: ResourceTone }>
}

type AppShellProps = {
  activeSection: GameSection
  gameID: string
  turn?: number
  phaseLabel?: string
  status: string
  statusTone: 'neutral' | 'success' | 'warning' | 'danger'
  lifecycle?: string
  resources?: ResourceChip[]
  onNavigate: (section: GameSection) => void
  onResourceActivate?: (resourceID: string) => boolean
  onHome: () => void
  onSaveGame?: () => void
  onLoadGame?: () => void
  persistenceBusy?: boolean
  onEndTurn?: () => void
  endTurnDisabled?: boolean
  endTurnLabel?: string
  navigationLocked?: boolean
  immersive?: boolean
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

function NavItems({ items, activeSection, onNavigate, disabled = false }: {
  items: typeof primaryNavItems
  activeSection: GameSection
  onNavigate: (section: GameSection) => void
  disabled?: boolean
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
          disabled={disabled}
          onClick={() => onNavigate(item.section)}
        >
          <span className="nav-glyph" aria-hidden="true"><GameIcon name={item.icon} /></span>
          <span>{t(item.label)}</span>
        </button>
      )
    })}
  </>
}

export function AppShell({ activeSection, gameID, turn, phaseLabel, status, statusTone, lifecycle, resources = [], onNavigate, onResourceActivate, onHome, onSaveGame, onLoadGame, persistenceBusy = false, onEndTurn, endTurnDisabled = false, endTurnLabel, navigationLocked = false, immersive = false, children }: AppShellProps) {
  const { t } = useI18n()
  const [menuOpen, setMenuOpen] = useState(false)
  const [activeResourceID, setActiveResourceID] = useState<string | null>(null)
  const menuRef = useRef<HTMLDivElement | null>(null)
  const resourcePopupRef = useRef<HTMLDivElement | null>(null)
  const secondaryActive = activeSection === 'espionage' || activeSection === 'shipbuilder' || activeSection === 'more'

  useEffect(() => {
    if (!activeResourceID) return
    const closeOnPointer = (event: PointerEvent) => {
      if (!resourcePopupRef.current?.contains(event.target as Node) && !(event.target as Element | null)?.closest?.('[data-resource-id]')) setActiveResourceID(null)
    }
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setActiveResourceID(null)
    }
    window.addEventListener('pointerdown', closeOnPointer)
    window.addEventListener('keydown', closeOnEscape)
    return () => {
      window.removeEventListener('pointerdown', closeOnPointer)
      window.removeEventListener('keydown', closeOnEscape)
    }
  }, [activeResourceID])
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
    if (navigationLocked) return
    setMenuOpen(false)
    onNavigate(section)
  }

  return (
    <div className={'game-shell section-' + activeSection + (immersive ? ' shell-immersive' : '')} data-lifecycle={lifecycle}>
      <header className="topbar">
        <div className="main-menu-anchor" ref={menuRef}>
          <button
            type="button"
            className={'main-menu-trigger connection-frame-' + statusTone + (menuOpen || secondaryActive ? ' active' : '')}
            aria-label={t('gameMenu.open') + ' · ' + status}
            data-connection-status={statusTone}
            aria-haspopup="dialog"
            aria-expanded={menuOpen}
            title={t('gameMenu.open') + ' · ' + status}
            onClick={() => { setActiveResourceID(null); setMenuOpen((open) => !open) }}
          >
            <GameIcon name="menu" aria-hidden="true" />
          </button>
          {menuOpen && (
            <section className="main-menu-popover" role="dialog" aria-label={t('gameMenu.title')} data-game-id={gameID}>
              <header className="main-menu-header">
                <div><p className="eyebrow">{t('gameMenu.title')}</p></div>
                <button type="button" className="button-ghost main-menu-close" onClick={() => setMenuOpen(false)} aria-label={t('common.close')}><GameIcon name="close" /></button>
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
                <button type="button" className="main-menu-item" disabled={!onSaveGame || persistenceBusy} onClick={() => { setMenuOpen(false); onSaveGame?.() }}>
                  <span className="main-menu-item-glyph" aria-hidden="true"><GameIcon name="check" /></span>
                  <span><strong>{t('gameMenu.saveGame')}</strong><small>{t('gameMenu.saveGameHint')}</small></span>
                </button>
                <button type="button" className="main-menu-item" disabled={!onLoadGame || persistenceBusy} onClick={() => { setMenuOpen(false); onLoadGame?.() }}>
                  <span className="main-menu-item-glyph" aria-hidden="true"><GameIcon name="open" /></span>
                  <span><strong>{t('gameMenu.loadGame')}</strong><small>{t('gameMenu.loadGameHint')}</small></span>
                </button>
                <button type="button" className="main-menu-item" disabled={navigationLocked} onClick={() => navigateFromMenu('shipbuilder')}>
                  <ProceduralShipGlyph seed="shipbuilder-menu" hullId="frigate" className="main-menu-vector-glyph" />
                  <span><strong>{t('gameMenu.shipbuilder')}</strong><small>{t('gameMenu.shipbuilderHint')}</small></span>
                </button>
                <button type="button" className="main-menu-item" disabled={navigationLocked} onClick={() => navigateFromMenu('more')}>
                  <span className="main-menu-item-glyph" aria-hidden="true"><GameIcon name="more" /></span>
                  <span><strong>{t('gameMenu.advanced')}</strong><small>{t('gameMenu.advancedHint')}</small></span>
                </button>
              </div>

              <div className="main-menu-section main-menu-section-last">
                <button type="button" className="main-menu-item main-menu-home" onClick={() => { setMenuOpen(false); onHome() }}>
                  <span className="main-menu-item-glyph" aria-hidden="true"><GameIcon name="home" /></span>
                  <span><strong>{t('gameMenu.mainMenu')}</strong><small>{t('gameMenu.mainMenuHint')}</small></span>
                </button>
              </div>
            </section>
          )}
        </div>

        <div className="topbar-resources" aria-label={t('a11y.resources')}>
          {resources.map((resource) => (
            <button
              type="button"
              className={`topbar-resource resource-${resource.tone ?? 'neutral'}${activeResourceID === resource.id ? ' active' : ''}`}
              key={resource.id}
              data-resource-id={resource.id}
              aria-expanded={activeResourceID === resource.id}
              title={`${resource.label}: ${[resource.value, resource.delta].filter(Boolean).join(' ')}`}
              onClick={() => {
                setMenuOpen(false)
                if (onResourceActivate?.(resource.id)) { setActiveResourceID(null); return }
                setActiveResourceID((current) => current === resource.id ? null : resource.id)
              }}
            >
              <span className="resource-icon" aria-hidden="true"><GameIcon name={resource.icon} /></span>
              <span className="resource-label resource-label-long">{resource.label}</span>
              <span className="resource-label resource-label-short">{resource.shortLabel ?? resource.label}</span>
              {resource.value && <strong>{resource.value}</strong>}
              {resource.delta && <span className={`resource-delta resource-${resource.deltaTone ?? 'neutral'}`}>{resource.delta}</span>}
            </button>
          ))}
        </div>

        {activeResourceID && (() => {
          const resource = resources.find((item) => item.id === activeResourceID)
          if (!resource) return null
          return (
            <section ref={resourcePopupRef} className="resource-detail-popover" role="dialog" aria-label={resource.detailTitle}>
              <header className="resource-detail-header">
                <div className="resource-detail-heading">
                  <span className="resource-detail-icon" aria-hidden="true"><GameIcon name={resource.icon} /></span>
                  <div><p className="eyebrow">{resource.label}</p><strong>{resource.detailTitle}</strong></div>
                </div>
                <button type="button" className="button-ghost resource-detail-close" onClick={() => setActiveResourceID(null)} aria-label={t('common.close')}><GameIcon name="close" /></button>
              </header>
              {resource.progressPercent !== undefined && (
                <div className="resource-progress" aria-label={`${Math.round(resource.progressPercent)}%`}>
                  <span style={{ width: `${Math.max(0, Math.min(100, resource.progressPercent))}%` }} />
                </div>
              )}
              <dl className="resource-detail-list">
                {resource.details.map((detail) => (
                  <div key={detail.label}><dt>{detail.label}</dt><dd className={`resource-text-${detail.tone ?? 'neutral'}`}>{detail.value}</dd></div>
                ))}
              </dl>
            </section>
          )
        })()}

        <div className="topbar-context" aria-label={t('a11y.gameStatus')}>
          {turn !== undefined && <span className="status-chip">{t('top.turn', { turn })}</span>}
          {phaseLabel && <span className="status-chip status-chip-phase">{phaseLabel}</span>}
        </div>

      </header>

      <div className="shell-layout">
        <aside className="side-nav" aria-label={t('a11y.primaryNavigation')}>
          <div className="side-nav-items"><NavItems items={primaryNavItems} activeSection={activeSection} onNavigate={onNavigate} disabled={navigationLocked} /></div>
        </aside>
        <main className="game-content">{children}</main>
      </div>

      <div className="bottom-command-bar">
        <nav className="bottom-nav" aria-label={t('a11y.primaryNavigation')}>
          <NavItems items={primaryNavItems} activeSection={activeSection} onNavigate={onNavigate} disabled={navigationLocked} />
        </nav>
        <button
          type="button"
          className="button-primary bottom-end-turn"
          disabled={!onEndTurn || endTurnDisabled || navigationLocked}
          onClick={onEndTurn}
        >
          <GameIcon name="check" />{endTurnLabel ?? t('planning.endTurn')}
        </button>
      </div>
    </div>
  )
}export { LanguageSwitch }
