export type VisualHullID = 'scout' | 'frigate' | 'destroyer' | 'cruiser' | 'battleship' | 'titan' | 'doom_star' | string
export type ShipStyleID = 'spear' | 'sleek' | 'organic'
export type ShipGenomeLocks = {
  core: boolean
  primitives: boolean
  engines: boolean
  cutouts: boolean
}

export const emptyShipGenomeLocks: ShipGenomeLocks = {
  core: false,
  primitives: false,
  engines: false,
  cutouts: false,
}
export type ShipPrimitiveKind = 'wedge' | 'spike' | 'pod'

export type ShipCutoutGene = {
  t: number
  side: -1 | 1
  offset: number
  rx: number
  ry: number
  angle: number
}

export type ShipPrimitiveGene = {
  kind: ShipPrimitiveKind
  t: number
  side: -1 | 1
  length: number
  width: number
  sweep: number
  mirrored: boolean
}

export type ShipVisualGenome = {
  version: 2
  hullId: VisualHullID
  styleId: ShipStyleID
  seed: string
  length: number
  beam: number
  stationCount: number
  stationWidths: number[]
  notchDepths: number[]
  engineCount: number
  detailCount: number
  cutouts: ShipCutoutGene[]
  primitives: ShipPrimitiveGene[]
}

type HullProfile = {
  length: number
  beam: number
  stationCount: number
  notchCount: number
  cutoutCount: number
  primitiveMin: number
  primitiveMax: number
  engineMin: number
  engineMax: number
  detailCount: number
}

export function hashSeed(value: string): number {
  let hash = 0x811c9dc5
  for (let index = 0; index < value.length; index += 1) {
    hash ^= value.charCodeAt(index)
    hash = Math.imul(hash, 0x01000193)
  }
  return hash >>> 0
}

export function createRandom(seed: number): () => number {
  let state = seed || 0x6d2b79f5
  return () => {
    state += 0x6d2b79f5
    let value = state
    value = Math.imul(value ^ (value >>> 15), value | 1)
    value ^= value + Math.imul(value ^ (value >>> 7), value | 61)
    return ((value ^ (value >>> 14)) >>> 0) / 4294967296
  }
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value))
}

function profileFor(hullId: VisualHullID): HullProfile {
  const token = String(hullId).toLowerCase().replace(/[\s-]+/g, '_')
  if (token.includes('doom')) return { length: 114, beam: 31.5, stationCount: 13, notchCount: 3, cutoutCount: 2, primitiveMin: 12, primitiveMax: 20, engineMin: 4, engineMax: 6, detailCount: 7 }
  if (token.includes('titan')) return { length: 108, beam: 29, stationCount: 12, notchCount: 3, cutoutCount: 2, primitiveMin: 10, primitiveMax: 16, engineMin: 3, engineMax: 5, detailCount: 6 }
  if (token.includes('battleship')) return { length: 101, beam: 25, stationCount: 11, notchCount: 2, cutoutCount: 1, primitiveMin: 8, primitiveMax: 14, engineMin: 3, engineMax: 4, detailCount: 5 }
  if (token.includes('cruiser')) return { length: 92, beam: 20, stationCount: 10, notchCount: 2, cutoutCount: 1, primitiveMin: 7, primitiveMax: 11, engineMin: 2, engineMax: 3, detailCount: 5 }
  if (token.includes('destroyer')) return { length: 82, beam: 16, stationCount: 9, notchCount: 1, cutoutCount: 0, primitiveMin: 5, primitiveMax: 9, engineMin: 2, engineMax: 3, detailCount: 4 }
  if (token.includes('scout')) return { length: 57, beam: 8.5, stationCount: 7, notchCount: 0, cutoutCount: 0, primitiveMin: 3, primitiveMax: 5, engineMin: 1, engineMax: 1, detailCount: 2 }
  if (token.includes('frigate')) return { length: 70, beam: 12, stationCount: 8, notchCount: 1, cutoutCount: 0, primitiveMin: 4, primitiveMax: 7, engineMin: 1, engineMax: 2, detailCount: 3 }
  return { length: 72, beam: 13, stationCount: 8, notchCount: 1, cutoutCount: 0, primitiveMin: 4, primitiveMax: 7, engineMin: 1, engineMax: 2, detailCount: 3 }
}

function styleScale(styleId: ShipStyleID): { length: number; beam: number; mirrorChance: number } {
  if (styleId === 'sleek') return { length: 1.06, beam: .82, mirrorChance: .96 }
  if (styleId === 'organic') return { length: .95, beam: 1.12, mirrorChance: .68 }
  return { length: 1.04, beam: .92, mirrorChance: .92 }
}

function shuffledStations(random: () => number, count: number, stationCount: number): Set<number> {
  const candidates = Array.from({ length: Math.max(0, stationCount - 4) }, (_, index) => index + 2)
  const selected = new Set<number>()
  while (selected.size < Math.min(count, candidates.length)) {
    selected.add(candidates[Math.floor(random() * candidates.length)])
  }
  return selected
}

function randomPrimitive(random: () => number, beam: number, length: number, styleId: ShipStyleID): ShipPrimitiveGene {
  const roll = random()
  const kind: ShipPrimitiveKind = styleId === 'organic'
    ? (roll < .22 ? 'wedge' : roll < .36 ? 'spike' : 'pod')
    : styleId === 'sleek'
      ? (roll < .62 ? 'wedge' : roll < .72 ? 'spike' : 'pod')
      : (roll < .5 ? 'wedge' : roll < .9 ? 'spike' : 'pod')
  const side: -1 | 1 = random() < .5 ? -1 : 1
  const lengthFactor = styleId === 'spear' ? 1.2 : styleId === 'sleek' ? 1.05 : .82
  const widthFactor = styleId === 'organic' ? 1.22 : styleId === 'sleek' ? .72 : .9
  const sweepRange = styleId === 'spear' ? 1.18 : styleId === 'sleek' ? .55 : .9
  return {
    kind,
    t: .18 + random() * .68,
    side,
    length: Math.max(5, length * (.055 + random() * .105) * lengthFactor),
    width: Math.max(2.5, beam * (.12 + random() * .26) * widthFactor),
    sweep: -sweepRange + random() * sweepRange * 2,
    mirrored: random() < styleScale(styleId).mirrorChance,
  }
}

export function createShipGenome(seed: string, hullId: VisualHullID, styleId: ShipStyleID = 'spear'): ShipVisualGenome {
  const profile = profileFor(hullId)
  const style = styleScale(styleId)
  const length = profile.length * style.length
  const beam = profile.beam * style.beam
  const random = createRandom(hashSeed(`${seed}|${hullId}|${styleId}|v2`))
  const notchStations = shuffledStations(random, profile.notchCount, profile.stationCount)
  const stationWidths = Array.from({ length: profile.stationCount }, (_, station) => {
    const t = station / Math.max(1, profile.stationCount - 1)
    if (station === profile.stationCount - 1) return 0
    const envelope = Math.pow(Math.sin(Math.PI * Math.min(.985, t + .035)), .58)
    const taper = .84 + random() * .3
    return Math.max(2.4, beam * (.25 + .75 * envelope) * taper * (station === 0 ? .7 : 1))
  })
  const notchBase = styleId === 'sleek' ? .7 : styleId === 'organic' ? .42 : .5
  const notchDepths = stationWidths.map((width, station) => notchStations.has(station) ? width * (notchBase + random() * .2) : 0)
  const engineCount = profile.engineMin + Math.floor(random() * (profile.engineMax - profile.engineMin + 1))
  const primitiveCount = profile.primitiveMin + Math.floor(random() * (profile.primitiveMax - profile.primitiveMin + 1))
  const primitives = Array.from({ length: primitiveCount }, () => randomPrimitive(random, beam, length, styleId))
  const cutouts = Array.from({ length: profile.cutoutCount }, (_, index): ShipCutoutGene => ({
    t: .42 + ((index + 1) / (profile.cutoutCount + 1)) * .25 + (random() - .5) * .06,
    side: index % 2 === 0 ? -1 : 1,
    offset: beam * (.12 + random() * .14),
    rx: Math.max(3.3, length * (.035 + random() * .018)),
    ry: Math.max(2.1, beam * (.11 + random() * .06)),
    angle: -18 + random() * 36,
  }))

  return {
    version: 2,
    hullId,
    styleId,
    seed,
    length,
    beam,
    stationCount: profile.stationCount,
    stationWidths,
    notchDepths,
    engineCount,
    detailCount: profile.detailCount,
    cutouts,
    primitives,
  }
}

function mutateNumber(value: number, random: () => number, scale: number, amount: number, min: number, max: number): number {
  return clamp(value + (random() * 2 - 1) * scale * amount, min, max)
}

export function mutateShipGenome(parent: ShipVisualGenome, mutationSeed: string, amount: number, locks: Partial<ShipGenomeLocks> = {}): ShipVisualGenome {
  const profile = profileFor(parent.hullId)
  const random = createRandom(hashSeed(`${mutationSeed}|${parent.styleId}|mutate|v2`))
  const mutation = clamp(amount, .03, 1)

  const stationWidths = locks.core
    ? [...parent.stationWidths]
    : parent.stationWidths.map((value, index) => {
        if (index === parent.stationWidths.length - 1) return 0
        return mutateNumber(value, random, parent.beam * .18, mutation, 2.2, parent.beam * 1.35)
      })
  const notchDepths = locks.core
    ? [...parent.notchDepths]
    : parent.notchDepths.map((value, index) => value <= 0 ? 0 : mutateNumber(value, random, parent.beam * .14, mutation, stationWidths[index] * .24, stationWidths[index] * .9))

  let primitives = locks.primitives
    ? parent.primitives.map((primitive) => ({ ...primitive }))
    : parent.primitives.map((primitive): ShipPrimitiveGene => ({
        ...primitive,
        t: mutateNumber(primitive.t, random, .12, mutation, .08, .92),
        length: mutateNumber(primitive.length, random, parent.length * .08, mutation, 4, parent.length * .26),
        width: mutateNumber(primitive.width, random, parent.beam * .18, mutation, 2, parent.beam * .62),
        sweep: mutateNumber(primitive.sweep, random, .7, mutation, -1.5, 1.5),
      }))

  if (!locks.primitives && random() < mutation * .72 && primitives.length < profile.primitiveMax + 4) {
    primitives = [...primitives, randomPrimitive(random, parent.beam, parent.length, parent.styleId)]
  }
  if (!locks.primitives && random() < mutation * .42 && primitives.length > Math.max(2, profile.primitiveMin - 1)) {
    primitives = primitives.filter((_, index) => index !== Math.floor(random() * primitives.length))
  }

  const cutouts = locks.cutouts
    ? parent.cutouts.map((cutout) => ({ ...cutout }))
    : parent.cutouts.map((cutout): ShipCutoutGene => ({
        ...cutout,
        t: mutateNumber(cutout.t, random, .08, mutation, .2, .82),
        offset: mutateNumber(cutout.offset, random, parent.beam * .1, mutation, 0, parent.beam * .4),
        rx: mutateNumber(cutout.rx, random, parent.length * .03, mutation, 2.8, parent.length * .09),
        ry: mutateNumber(cutout.ry, random, parent.beam * .08, mutation, 1.8, parent.beam * .28),
        angle: mutateNumber(cutout.angle, random, 28, mutation, -45, 45),
      }))

  const engineDelta = locks.engines ? 0 : (random() < mutation * .35 ? (random() < .5 ? -1 : 1) : 0)
  return {
    ...parent,
    seed: mutationSeed,
    stationWidths,
    notchDepths,
    engineCount: Math.round(clamp(parent.engineCount + engineDelta, profile.engineMin, profile.engineMax)),
    primitives,
    cutouts,
  }
}