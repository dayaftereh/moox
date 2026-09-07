export type VisualHullID = 'scout' | 'frigate' | 'destroyer' | 'cruiser' | 'battleship' | 'titan' | 'doom_star' | string
export type ShipStyleID = 'spear' | 'sleek' | 'organic'
export type ShipMorphologyID = 'needle' | 'barge' | 'manta' | 'fork' | 'chevron' | 'hammer' | 'bulb'
export type ShipPrimitiveKind = 'wedge' | 'spike' | 'pod'

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

// Visual Genome v4 stores only one canonical half of lateral details.
// Rendering mirrors that half across the longitudinal X axis.
export type ShipCutoutGene = {
  t: number
  offset: number
  rx: number
  ry: number
  angle: number
}

export type ShipPrimitiveGene = {
  kind: ShipPrimitiveKind
  t: number
  length: number
  width: number
  sweep: number
}

export type ShipVisualGenome = {
  version: 4
  hullId: VisualHullID
  styleId: ShipStyleID
  morphologyId: ShipMorphologyID
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

type MorphologyProfile = {
  lengthScale: number
  beamScale: number
  coreDensity: number
  frontBias: number
  rearBias: number
  notchScale: number
  primitiveScale: number
  cutoutBonus: number
}

const styles: ShipStyleID[] = ['spear', 'sleek', 'organic']
const morphologies: ShipMorphologyID[] = ['needle', 'barge', 'manta', 'fork', 'chevron', 'hammer', 'bulb']

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

export function shipHullSpace(hullId: VisualHullID): number {
  const token = String(hullId).toLowerCase().replace(/[\s-]+/g, '_')
  if (token.includes('doom')) return 1200
  if (token.includes('titan')) return 500
  if (token.includes('battleship')) return 250
  if (token.includes('cruiser')) return 120
  if (token.includes('destroyer')) return 60
  // Scout is currently a Frigate-rules visual role.
  return 25
}

export function shipHullFootprint(hullId: VisualHullID): number {
  const token = String(hullId).toLowerCase().replace(/[\s-]+/g, '_')
  if (token.includes('doom')) return 1
  if (token.includes('titan')) return .9
  if (token.includes('battleship')) return .78
  if (token.includes('cruiser')) return .68
  if (token.includes('destroyer')) return .58
  if (token.includes('frigate')) return .48
  if (token.includes('scout')) return .4
  return .5
}

function profileFor(hullId: VisualHullID): HullProfile {
  const token = String(hullId).toLowerCase().replace(/[\s-]+/g, '_')
  if (token.includes('doom')) return { length: 110, beam: 29, stationCount: 13, notchCount: 3, cutoutCount: 2, primitiveMin: 12, primitiveMax: 20, engineMin: 4, engineMax: 6, detailCount: 7 }
  if (token.includes('titan')) return { length: 104, beam: 27, stationCount: 12, notchCount: 3, cutoutCount: 2, primitiveMin: 10, primitiveMax: 16, engineMin: 3, engineMax: 5, detailCount: 6 }
  if (token.includes('battleship')) return { length: 96, beam: 23, stationCount: 11, notchCount: 2, cutoutCount: 1, primitiveMin: 8, primitiveMax: 14, engineMin: 3, engineMax: 4, detailCount: 5 }
  if (token.includes('cruiser')) return { length: 88, beam: 19, stationCount: 10, notchCount: 2, cutoutCount: 1, primitiveMin: 7, primitiveMax: 11, engineMin: 2, engineMax: 3, detailCount: 5 }
  if (token.includes('destroyer')) return { length: 78, beam: 15, stationCount: 9, notchCount: 1, cutoutCount: 0, primitiveMin: 5, primitiveMax: 9, engineMin: 2, engineMax: 3, detailCount: 4 }
  if (token.includes('scout')) return { length: 54, beam: 8.5, stationCount: 7, notchCount: 0, cutoutCount: 0, primitiveMin: 3, primitiveMax: 5, engineMin: 1, engineMax: 1, detailCount: 2 }
  if (token.includes('frigate')) return { length: 66, beam: 11.5, stationCount: 8, notchCount: 1, cutoutCount: 0, primitiveMin: 4, primitiveMax: 7, engineMin: 1, engineMax: 2, detailCount: 3 }
  return { length: 68, beam: 12, stationCount: 8, notchCount: 1, cutoutCount: 0, primitiveMin: 4, primitiveMax: 7, engineMin: 1, engineMax: 2, detailCount: 3 }
}

function styleScale(styleId: ShipStyleID): { length: number; beam: number } {
  if (styleId === 'sleek') return { length: 1.06, beam: .9 }
  if (styleId === 'organic') return { length: .96, beam: 1.12 }
  return { length: 1.02, beam: 1 }
}

function morphologyProfile(morphologyId: ShipMorphologyID): MorphologyProfile {
  switch (morphologyId) {
    case 'needle': return { lengthScale: 1.28, beamScale: .58, coreDensity: .72, frontBias: .88, rearBias: .72, notchScale: .55, primitiveScale: .78, cutoutBonus: 0 }
    case 'barge': return { lengthScale: .76, beamScale: 1.62, coreDensity: 1.18, frontBias: .78, rearBias: .94, notchScale: .72, primitiveScale: 1.08, cutoutBonus: 0 }
    case 'manta': return { lengthScale: .84, beamScale: 1.82, coreDensity: .86, frontBias: .58, rearBias: .52, notchScale: .9, primitiveScale: 1.34, cutoutBonus: 1 }
    case 'fork': return { lengthScale: .98, beamScale: 1.26, coreDensity: .64, frontBias: .52, rearBias: .82, notchScale: 1.18, primitiveScale: 1.24, cutoutBonus: 1 }
    case 'chevron': return { lengthScale: .82, beamScale: 1.7, coreDensity: .6, frontBias: .44, rearBias: .7, notchScale: 1.08, primitiveScale: 1.5, cutoutBonus: 0 }
    case 'hammer': return { lengthScale: .78, beamScale: 1.42, coreDensity: .92, frontBias: 1.38, rearBias: .62, notchScale: .8, primitiveScale: 1.2, cutoutBonus: 0 }
    case 'bulb': return { lengthScale: .68, beamScale: 1.9, coreDensity: 1.34, frontBias: .92, rearBias: .92, notchScale: .38, primitiveScale: 1.06, cutoutBonus: 1 }
  }
}

function shuffledStations(random: () => number, count: number, stationCount: number): Set<number> {
  const candidates = Array.from({ length: Math.max(0, stationCount - 4) }, (_, index) => index + 2)
  const selected = new Set<number>()
  while (selected.size < Math.min(count, candidates.length)) selected.add(candidates[Math.floor(random() * candidates.length)])
  return selected
}

function randomPrimitive(random: () => number, beam: number, length: number, styleId: ShipStyleID, morphologyId: ShipMorphologyID): ShipPrimitiveGene {
  const roll = random()
  let kind: ShipPrimitiveKind
  if (morphologyId === 'bulb' || styleId === 'organic') kind = roll < .18 ? 'wedge' : roll < .32 ? 'spike' : 'pod'
  else if (morphologyId === 'needle') kind = roll < .34 ? 'wedge' : roll < .88 ? 'spike' : 'pod'
  else if (morphologyId === 'manta' || morphologyId === 'chevron') kind = roll < .72 ? 'wedge' : roll < .84 ? 'spike' : 'pod'
  else if (styleId === 'sleek') kind = roll < .62 ? 'wedge' : roll < .72 ? 'spike' : 'pod'
  else kind = roll < .48 ? 'wedge' : roll < .82 ? 'spike' : 'pod'

  const morph = morphologyProfile(morphologyId)
  const styleLength = styleId === 'spear' ? 1.15 : styleId === 'organic' ? .88 : 1.02
  const styleWidth = styleId === 'organic' ? 1.18 : styleId === 'sleek' ? .8 : 1
  const sweepRange = morphologyId === 'chevron' ? 1.45 : morphologyId === 'manta' ? 1.1 : morphologyId === 'needle' ? .55 : 1

  return {
    kind,
    t: .12 + random() * .76,
    length: Math.max(4.5, length * (.045 + random() * .11) * styleLength * morph.primitiveScale),
    width: Math.max(2.5, beam * (.1 + random() * .28) * styleWidth),
    sweep: -sweepRange + random() * sweepRange * 2,
  }
}

function signaturePrimitives(morphologyId: ShipMorphologyID, beam: number, length: number): ShipPrimitiveGene[] {
  switch (morphologyId) {
    case 'manta':
      return [
        { kind: 'wedge', t: .46, length: length * .28, width: beam * .8, sweep: -.9 },
        { kind: 'wedge', t: .58, length: length * .22, width: beam * .66, sweep: .55 },
      ]
    case 'fork':
      return [
        { kind: 'spike', t: .7, length: length * .34, width: beam * .42, sweep: .12 },
        { kind: 'pod', t: .36, length: length * .14, width: beam * .46, sweep: -.25 },
      ]
    case 'chevron':
      return [
        { kind: 'wedge', t: .5, length: length * .4, width: beam * .86, sweep: -1.2 },
        { kind: 'spike', t: .34, length: length * .26, width: beam * .46, sweep: -.5 },
      ]
    case 'hammer':
      return [
        { kind: 'pod', t: .72, length: length * .22, width: beam * .7, sweep: .08 },
        { kind: 'wedge', t: .68, length: length * .18, width: beam * .58, sweep: .48 },
      ]
    case 'bulb':
      return [
        { kind: 'pod', t: .46, length: length * .28, width: beam * .72, sweep: 0 },
        { kind: 'pod', t: .62, length: length * .22, width: beam * .62, sweep: .18 },
      ]
    case 'barge':
      return [{ kind: 'pod', t: .48, length: length * .2, width: beam * .42, sweep: 0 }]
    case 'needle':
      return [{ kind: 'spike', t: .74, length: length * .18, width: beam * .24, sweep: .2 }]
  }
}

function envelopeFor(morphologyId: ShipMorphologyID, t: number): number {
  const sin = Math.max(0, Math.sin(Math.PI * Math.min(.995, t + .02)))
  switch (morphologyId) {
    case 'needle': return Math.pow(sin, .9)
    case 'barge': return .58 + .42 * Math.pow(sin, .32)
    case 'manta': return Math.pow(sin, .24)
    case 'fork': return .4 + .6 * Math.pow(sin, .62)
    case 'chevron': return .28 + .72 * Math.pow(sin, .72)
    case 'hammer': return .38 + .62 * Math.pow(t, .82)
    case 'bulb': return .72 + .28 * Math.pow(sin, .16)
  }
}

export function createShipGenome(
  seed: string,
  hullId: VisualHullID,
  styleId: ShipStyleID = 'spear',
  morphologyId: ShipMorphologyID = morphologies[hashSeed(`${seed}|morphology`) % morphologies.length],
): ShipVisualGenome {
  const profile = profileFor(hullId)
  const style = styleScale(styleId)
  const morph = morphologyProfile(morphologyId)
  const length = profile.length * style.length * morph.lengthScale
  const beam = profile.beam * style.beam * morph.beamScale
  const random = createRandom(hashSeed(`${seed}|${hullId}|${styleId}|${morphologyId}|v4`))
  const notchCount = Math.min(profile.stationCount - 4, Math.max(0, Math.round(profile.notchCount * morph.notchScale)))
  const notchStations = shuffledStations(random, notchCount, profile.stationCount)

  // Canonical +Y half-width samples. Rendering mirrors them across the X axis.
  const stationWidths = Array.from({ length: profile.stationCount }, (_, station) => {
    const t = station / Math.max(1, profile.stationCount - 1)
    if (station === profile.stationCount - 1) return 0
    const envelope = envelopeFor(morphologyId, t)
    const endBias = (1 - t) * morph.rearBias + t * morph.frontBias
    const jitter = .8 + random() * .4
    const startScale = station === 0 ? .72 : 1
    return Math.max(2.2, beam * (.18 + .82 * envelope) * morph.coreDensity * endBias * jitter * startScale)
  })

  const notchBase = morphologyId === 'fork' || morphologyId === 'chevron' ? .34 : styleId === 'sleek' ? .65 : styleId === 'organic' ? .4 : .5
  const notchDepths = stationWidths.map((width, station) => notchStations.has(station) ? width * (notchBase + random() * .18) : 0)
  const engineCount = profile.engineMin + Math.floor(random() * (profile.engineMax - profile.engineMin + 1))
  const primitiveCount = profile.primitiveMin + Math.floor(random() * (profile.primitiveMax - profile.primitiveMin + 1))
  const primitives = [
    ...signaturePrimitives(morphologyId, beam, length),
    ...Array.from({ length: primitiveCount }, () => randomPrimitive(random, beam, length, styleId, morphologyId)),
  ]

  const cutoutCount = Math.min(3, profile.cutoutCount + morph.cutoutBonus)
  const cutouts = Array.from({ length: cutoutCount }, (_, index): ShipCutoutGene => {
    const isForkCenter = morphologyId === 'fork' && index === 0
    return {
      t: isForkCenter ? .66 : .34 + ((index + 1) / (cutoutCount + 1)) * .4 + (random() - .5) * .05,
      offset: isForkCenter ? 0 : beam * (.08 + random() * .16),
      rx: isForkCenter ? Math.max(5, length * .09) : Math.max(3.2, length * (.03 + random() * .024)),
      ry: isForkCenter ? Math.max(3, beam * .22) : Math.max(2, beam * (.1 + random() * .08)),
      angle: isForkCenter ? 0 : -28 + random() * 56,
    }
  })

  return {
    version: 4,
    hullId,
    styleId,
    morphologyId,
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

export function createRandomShipGenome(seed: string, hullId: VisualHullID): ShipVisualGenome {
  const random = createRandom(hashSeed(`${seed}|full-random-v4`))
  const morphologyId = morphologies[Math.floor(random() * morphologies.length)]
  const styleId = styles[Math.floor(random() * styles.length)]
  return createShipGenome(seed, hullId, styleId, morphologyId)
}

function mutateNumber(value: number, random: () => number, scale: number, amount: number, min: number, max: number): number {
  return clamp(value + (random() * 2 - 1) * scale * amount, min, max)
}

// Retained as advanced generator research; the primary player UX uses full rerolls.
export function mutateShipGenome(parent: ShipVisualGenome, mutationSeed: string, amount: number, locks: Partial<ShipGenomeLocks> = {}): ShipVisualGenome {
  const profile = profileFor(parent.hullId)
  const random = createRandom(hashSeed(`${mutationSeed}|${parent.styleId}|${parent.morphologyId}|mutate|v4`))
  const mutation = clamp(amount, .03, 1)
  const stationWidths = locks.core
    ? [...parent.stationWidths]
    : parent.stationWidths.map((value, index) => index === parent.stationWidths.length - 1 ? 0 : mutateNumber(value, random, parent.beam * .18, mutation, 2.2, parent.beam * 1.55))
  const notchDepths = locks.core
    ? [...parent.notchDepths]
    : parent.notchDepths.map((value, index) => value <= 0 ? 0 : mutateNumber(value, random, parent.beam * .14, mutation, stationWidths[index] * .2, stationWidths[index] * .9))

  let primitives = locks.primitives
    ? parent.primitives.map((primitive) => ({ ...primitive }))
    : parent.primitives.map((primitive): ShipPrimitiveGene => ({
        ...primitive,
        t: mutateNumber(primitive.t, random, .12, mutation, .06, .94),
        length: mutateNumber(primitive.length, random, parent.length * .08, mutation, 4, parent.length * .34),
        width: mutateNumber(primitive.width, random, parent.beam * .18, mutation, 2, parent.beam * .8),
        sweep: mutateNumber(primitive.sweep, random, .7, mutation, -1.7, 1.7),
      }))

  if (!locks.primitives && random() < mutation * .72 && primitives.length < profile.primitiveMax + 8) {
    primitives = [...primitives, randomPrimitive(random, parent.beam, parent.length, parent.styleId, parent.morphologyId)]
  }
  if (!locks.primitives && random() < mutation * .42 && primitives.length > Math.max(2, profile.primitiveMin - 1)) {
    primitives = primitives.filter((_, index) => index !== Math.floor(random() * primitives.length))
  }

  const cutouts = locks.cutouts
    ? parent.cutouts.map((cutout) => ({ ...cutout }))
    : parent.cutouts.map((cutout): ShipCutoutGene => ({
        ...cutout,
        t: mutateNumber(cutout.t, random, .08, mutation, .18, .86),
        offset: mutateNumber(cutout.offset, random, parent.beam * .1, mutation, 0, parent.beam * .45),
        rx: mutateNumber(cutout.rx, random, parent.length * .03, mutation, 2.8, parent.length * .12),
        ry: mutateNumber(cutout.ry, random, parent.beam * .08, mutation, 1.8, parent.beam * .34),
        angle: mutateNumber(cutout.angle, random, 28, mutation, -48, 48),
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