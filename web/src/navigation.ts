export type GameSection = 'galaxy' | 'colonies' | 'fleets' | 'research' | 'more'

export type AppRoute =
  | { kind: 'home' }
  | { kind: 'new-game' }
  | { kind: 'game'; gameID: string; section: GameSection }

const sections = new Set<GameSection>(['galaxy', 'colonies', 'fleets', 'research', 'more'])

export function parseRoute(hash = window.location.hash): AppRoute {
  const value = hash.replace(/^#/, '') || '/'
  const parts = value.split('/').filter(Boolean)
  if (parts.length === 1 && parts[0] === 'new-game') return { kind: 'new-game' }
  if (parts.length >= 3 && parts[0] === 'game') {
    const section = parts[2] as GameSection
    if (sections.has(section)) {
      return { kind: 'game', gameID: decodeURIComponent(parts[1]), section }
    }
  }
  return { kind: 'home' }
}

export function routeHash(route: AppRoute): string {
  if (route.kind === 'home') return '#/'
  if (route.kind === 'new-game') return '#/new-game'
  return `#/game/${encodeURIComponent(route.gameID)}/${route.section}`
}

export function navigate(route: AppRoute) {
  const next = routeHash(route)
  if (window.location.hash === next) {
    window.dispatchEvent(new HashChangeEvent('hashchange'))
    return
  }
  window.location.hash = next
}
