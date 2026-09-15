export type NewGameAssetDomain =
  | 'difficulty'
  | 'galaxy-size'
  | 'galaxy-age'
  | 'technology-level'
  | 'player-count'

export type NewGameAssetFormat = 'svg' | 'webp' | 'png'

export const newGameAssetManifestPath = '/assets/new-game/manifest.json'

const semanticSegment = /^[a-z0-9]+(?:-[a-z0-9]+)*$/

function assertSemanticSegment(value: string, label: string) {
  if (!semanticSegment.test(value)) throw new Error(`invalid ${label} ${JSON.stringify(value)}`)
}

export function newGameAssetID(domain: NewGameAssetDomain, optionID: string): string {
  assertSemanticSegment(optionID, 'New Game option id')
  return `new-game:${domain}:${optionID}`
}

export function newGameAssetPath(domain: NewGameAssetDomain, optionID: string, format: NewGameAssetFormat): string {
  assertSemanticSegment(optionID, 'New Game option id')
  return `/assets/new-game/${domain}/${optionID}.${format}`
}

export function racePortraitAssetPath(raceID: string, format: 'webp' | 'png' = 'webp'): string {
  assertSemanticSegment(raceID, 'race id')
  return `/assets/races/${raceID}/portrait.${format}`
}

export function raceEmblemAssetPath(raceID: string): string {
  assertSemanticSegment(raceID, 'race id')
  return `/assets/races/${raceID}/emblem.svg`
}
