export type VisualHullID = 'scout' | 'frigate' | 'destroyer' | 'cruiser' | 'battleship' | 'titan' | 'doom_star' | string

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

function shuffledStations(random: () => number, count: number, stationCount: number): Set<number> {
  const candidates = Array.from({ length: Math.max(0, stationCount - 4) }, (_, index) => index + 2)
  const selected = new Set<number>()
  while (selected.size < Math.min(count, candidates.length)) {
    selected.add(candidates[Math.floor(random() * candidates.length)])
  }
  return selected
}

function randomPrimitive(random: () => number, beam: number, length: number): ShipPrimitiveGene {
  const roll = random()
  const kind: ShipPrimitiveKind = roll < .48 ? 'wedge' : roll < .8 ? 'spike' : 'pod'
  const side: -1 | 1 = random() < .5 ? -1 : 1
  return {
    kind,
    t: .18 + random() * .68,
    side,
    length: Math.max(5, length * (.055 + random() * .105)),
    width: Math.max(2.5, beam * (.12 + random() * .26)),
    sweep: -.85 + random() * 1.7,
    mirrored: random() < .9,
  }
}

export function createShipGenome(seed: string, hullId: VisualHullID): ShipVisualGenome {
  const profile = profileFor(hullId)
  const random = createRandom(hashSeed(`${seed}|${hullId}|v2`))
  const notchStations = shuffledStations(random, profile.notchCount, profile.stationCount)
  const stationWidths = Array.from({ length: profile.stationCount }, (_, station) => {
    const t = station / Math.max(1, profile.stationCount - 1)
    if (station === profile.stationCount - 1) return 0
    const envelope = Math.pow(Math.sin(Math.PI * Math.min(.985, t + .035)), .58)
    const taper = .84 + random() * .3
    return Math.max(2.4, profile.beam * (.25 + .75 * envelope) * taper * (station === 0 ? .7 : 1))
  })
  const notchDepths = stationWidths.map((width, station) => notchStations.has(station) ? width * (.5 + random() * .24) : 0)
  const engineCount = profile.engineMin + Math.floor(random() * (profile.engineMax - profile.engineMin + 1))
  const primitiveCount = profile.primitiveMin + Math.floor(random() * (profile.primitiveMax - profile.primitiveMin + 1))
  const primitives = Array.from({ length: primitiveCount }, () => randomPrimitive(random, profile.beam, profile.length))
  const cutouts = Array.from({ length: profile.cutoutCount }, (_, index): ShipCutoutGene => ({
    t: .42 + ((index + 1) / (profile.cutoutCount + 1)) * .25 + (random() - .5) * .06,
    side: index % 2 === 0 ? -1 : 1,
    offset: profile.beam * (.12 + random() * .14),
    rx: Math.max(3.3, profile.length * (.035 + random() * .018)),
    ry: Math.max(2.1, profile.beam * (.11 + random() * .06)),
    angle: -18 + random() * 36,
  }))

  return {
    version: 2,
    hullId,
    seed,
    length: profile.length,
    beam: profile.beam,
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

export function mutateShipGenome(parent: ShipVisualGenome, mutationSeed: string, amount: number): ShipVisualGenome {
  const profile = profileFor(parent.hullId)
  const random = createRandom(hashSeed(`${mutationSeed}|mutate|v2`))
  const mutation = clamp(amount, .03, 1)

  const stationWidths = parent.stationWidths.map((value, index) => {
    if (index === parent.stationWidths.length - 1) return 0
    return mutateNumber(value, random, profile.beam * .18, mutation, 2.2, profile.beam * 1.35)
  })
  const notchDepths = parent.notchDepths.map((value, index) => value <= 0 ? 0 : mutateNumber(value, random, profile.beam * .14, mutation, stationWidths[index] * .24, stationWidths[index] * .86))

  let primitives = parent.primitives.map((primitive): ShipPrimitiveGene => ({
    ...primitive,
    t: mutateNumber(primitive.t, random, .12, mutation, .08, .92),
    length: mutateNumber(primitive.length, random, profile.length * .08, mutation, 4, profile.length * .23),
    width: mutateNumber(primitive.width, random, profile.beam * .18, mutation, 2, profile.beam * .55),
    sweep: mutateNumber(primitive.sweep, random, .7, mutation, -1.4, 1.4),
  }))

  if (random() < mutation * .72 && primitives.length < profile.primitiveMax + 4) {
    primitives = [...primitives, randomPrimitive(random, profile.beam, profile.length)]
  }
  if (random() < mutation * .42 && primitives.length > Math.max(2, profile.primitiveMin - 1)) {
    primitives = primitives.filter((_, index) => index !== Math.floor(random() * primitives.length))
  }

  const cutouts = parent.cutouts.map((cutout): ShipCutoutGene => ({
    ...cutout,
    t: mutateNumber(cutout.t, random, .08, mutation, .2, .82),
    offset: mutateNumber(cutout.offset, random, profile.beam * .1, mutation, 0, profile.beam * .4),
    rx: mutateNumber(cutout.rx, random, profile.length * .03, mutation, 2.8, profile.length * .09),
    ry: mutateNumber(cutout.ry, random, profile.beam * .08, mutation, 1.8, profile.beam * .28),
    angle: mutateNumber(cutout.angle, random, 28, mutation, -45, 45),
  }))

  const engineDelta = random() < mutation * .35 ? (random() < .5 ? -1 : 1) : 0
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