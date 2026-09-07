export type GameSection = 'galaxy' | 'colonies' | 'fleets' | 'research' | 'diplomacy' | 'espionage' | 'shipbuilder' | 'more'
export type GameSubview = 'build'

export type AppRoute =
  | { kind: 'home' }
  | { kind: 'new-game' }
  | { kind: 'game'; gameID: string; section: GameSection; entityID?: number; subview?: GameSubview }

const sections = new Set<GameSection>(['galaxy', 'colonies', 'fleets', 'research', 'diplomacy', 'espionage', 'shipbuilder', 'more'])
const subviews = new Set<GameSubview>(['build'])

export function parseRoute(hash = window.location.hash): AppRoute {
  const value = hash.replace(/^#/, '') || '/'
  const parts = value.split('/').filter(Boolean)
  if (parts.length === 1 && parts[0] === 'new-game') return { kind: 'new-game' }
  if (parts.length >= 3 && parts[0] === 'game') {
    const section = parts[2] as GameSection
    if (sections.has(section)) {
      const parsed = parts.length >= 4 ? Number(parts[3]) : undefined
      const entityID = parsed !== undefined && Number.isFinite(parsed) && parsed > 0 ? parsed : undefined
      const parsedSubview = entityID && parts.length >= 5 && subviews.has(parts[4] as GameSubview) ? parts[4] as GameSubview : undefined
      return {
        kind: 'game',
        gameID: decodeURIComponent(parts[1]),
        section,
        entityID,
        subview: parsedSubview,
      }
    }
  }
  return { kind: 'home' }
}

export function routeHash(route: AppRoute): string {
  if (route.kind === 'home') return '#/'
  if (route.kind === 'new-game') return '#/new-game'
  const entitySuffix = route.entityID ? `/${route.entityID}` : ''
  const subviewSuffix = route.entityID && route.subview ? `/${route.subview}` : ''
  return `#/game/${encodeURIComponent(route.gameID)}/${route.section}${entitySuffix}${subviewSuffix}`
}

export function navigate(route: AppRoute) {
  const next = routeHash(route)
  if (window.location.hash === next) {
    window.dispatchEvent(new HashChangeEvent('hashchange'))
    return
  }
  window.location.hash = next
}
