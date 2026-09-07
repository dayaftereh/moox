type Point = readonly [number, number]

type ProceduralShipGlyphProps = {
  seed: string
  hullId?: string
  weaponCount?: number
  className?: string
  label?: string
}

type Cutout = {
  cx: number
  cy: number
  rx: number
  ry: number
  angle: number
}

type HullVisualProfile = {
  key: string
  length: number
  beam: number
  stations: number
  notches: number
  cutouts: number
  engineMin: number
  engineMax: number
  detail: number
}

type ShipGeometry = {
  hullPath: string
  panelLines: string[]
  engineYs: number[]
  engineX: number
  engineTrailX: number
  hardpoints: Point[]
  cutouts: Cutout[]
  spinePath: string
  notchCount: number
  profileKey: string
}

function hashSeed(value: string): number {
  let hash = 0x811c9dc5
  for (let index = 0; index < value.length; index += 1) {
    hash ^= value.charCodeAt(index)
    hash = Math.imul(hash, 0x01000193)
  }
  return hash >>> 0
}

function createRandom(seed: number): () => number {
  let state = seed || 0x6d2b79f5
  return () => {
    state += 0x6d2b79f5
    let value = state
    value = Math.imul(value ^ (value >>> 15), value | 1)
    value ^= value + Math.imul(value ^ (value >>> 7), value | 61)
    return ((value ^ (value >>> 14)) >>> 0) / 4294967296
  }
}

function visualProfile(hullId: string | undefined): HullVisualProfile {
  const token = (hullId ?? '').toLowerCase().replace(/[\s-]+/g, '_')
  if (token.includes('doom')) return { key: 'doom_star', length: 114, beam: 31.5, stations: 13, notches: 3, cutouts: 2, engineMin: 4, engineMax: 6, detail: 1 }
  if (token.includes('titan')) return { key: 'titan', length: 108, beam: 30, stations: 11, notches: 3, cutouts: 2, engineMin: 3, engineMax: 5, detail: .9 }
  if (token.includes('battleship')) return { key: 'battleship', length: 101, beam: 25, stations: 10, notches: 2, cutouts: 1, engineMin: 3, engineMax: 4, detail: .78 }
  if (token.includes('cruiser')) return { key: 'cruiser', length: 92, beam: 20, stations: 9, notches: 2, cutouts: 1, engineMin: 2, engineMax: 3, detail: .66 }
  if (token.includes('destroyer')) return { key: 'destroyer', length: 82, beam: 16, stations: 8, notches: 1, cutouts: 0, engineMin: 2, engineMax: 3, detail: .54 }
  if (token.includes('scout')) return { key: 'scout', length: 57, beam: 8.5, stations: 6, notches: 0, cutouts: 0, engineMin: 1, engineMax: 1, detail: .24 }
  if (token.includes('frigate')) return { key: 'frigate', length: 70, beam: 12, stations: 7, notches: 1, cutouts: 0, engineMin: 1, engineMax: 2, detail: .4 }
  if (token.includes('colony')) return { key: 'colony', length: 78, beam: 17, stations: 8, notches: 1, cutouts: 1, engineMin: 2, engineMax: 3, detail: .52 }
  if (token.includes('outpost')) return { key: 'outpost', length: 74, beam: 15, stations: 8, notches: 1, cutouts: 1, engineMin: 2, engineMax: 3, detail: .5 }
  return { key: 'generic', length: 72, beam: 13, stations: 7, notches: 1, cutouts: 0, engineMin: 1, engineMax: 2, detail: .42 }
}

function pointPath(points: Point[]): string {
  return points.map(([x, y], index) => `${index === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`).join(' ') + ' Z'
}

function chooseNotchStations(random: () => number, stationCount: number, count: number): Set<number> {
  const candidates = Array.from({ length: Math.max(0, stationCount - 4) }, (_, index) => index + 2)
  const selected = new Set<number>()
  while (selected.size < Math.min(count, candidates.length)) {
    selected.add(candidates[Math.floor(random() * candidates.length)])
  }
  return selected
}

function makeGeometry(seed: string, hullId: string | undefined, weaponCount: number): ShipGeometry {
  const profile = visualProfile(hullId)
  const random = createRandom(hashSeed(`${seed}|${profile.key}`))
  const centerX = 68
  const centerY = 42
  const rearX = centerX - profile.length / 2
  const noseX = centerX + profile.length / 2
  const segment = profile.length / Math.max(1, profile.stations - 1)
  const notchStations = chooseNotchStations(random, profile.stations, profile.notches)
  const wingStation = profile.stations > 6 ? 2 + Math.floor(random() * Math.max(1, profile.stations - 5)) : -1
  const upper: Point[] = []

  for (let station = 0; station < profile.stations; station += 1) {
    const t = station / (profile.stations - 1)
    const x = rearX + t * profile.length
    if (station === profile.stations - 1) {
      upper.push([noseX, centerY])
      continue
    }

    const envelope = Math.pow(Math.sin(Math.PI * Math.min(.985, t + .035)), .58)
    const taper = .86 + random() * .26
    const rearMass = station === 0 ? .7 : 1
    const wingBoost = station === wingStation ? profile.beam * (.12 + random() * .24) : 0
    const halfWidth = Math.max(2.8, profile.beam * (.28 + .72 * envelope) * taper * rearMass + wingBoost)

    if (notchStations.has(station)) {
      const shoulderWidth = Math.min(profile.beam * 1.28, halfWidth * (1.08 + random() * .16) + profile.beam * .12)
      const notchWidth = Math.max(profile.beam * .18, halfWidth * (.28 + random() * .2))
      const shoulderDX = segment * (.2 + random() * .08)
      upper.push([x - shoulderDX, centerY - shoulderWidth])
      upper.push([x, centerY - notchWidth])
      upper.push([x + shoulderDX, centerY - shoulderWidth * (.93 + random() * .1)])
      continue
    }

    upper.push([x, centerY - halfWidth])
  }

  const lower: Point[] = [...upper].reverse().map(([x, y]) => [x, centerY + (centerY - y)] as Point)
  const hullPath = pointPath([...upper, ...lower])

  const panelLines: string[] = []
  const panelCount = 2 + Math.round(profile.detail * 4) + Math.floor(random() * 2)
  for (let index = 0; index < panelCount; index += 1) {
    const t = .22 + ((index + 1) / (panelCount + 1)) * .58
    const x = rearX + t * profile.length
    const envelope = Math.pow(Math.sin(Math.PI * t), .62)
    const half = Math.max(3.5, envelope * profile.beam * (.38 + random() * .17))
    panelLines.push(`M ${x.toFixed(1)} ${(centerY - half).toFixed(1)} L ${x.toFixed(1)} ${(centerY + half).toFixed(1)}`)
  }

  const engineCount = profile.engineMin + Math.floor(random() * (profile.engineMax - profile.engineMin + 1))
  const engineSpacing = Math.min(9, Math.max(4.5, profile.beam * .34))
  const engineYs = Array.from({ length: engineCount }, (_, index) => centerY + (index - (engineCount - 1) / 2) * engineSpacing)
  const engineX = rearX + Math.max(1.8, profile.length * .018)
  const engineTrailX = Math.max(2, rearX - Math.max(7, profile.length * .085))

  const hardpointCount = Math.min(8, Math.max(0, weaponCount))
  const hardpoints: Point[] = Array.from({ length: hardpointCount }, (_, index) => {
    const t = .38 + (index / Math.max(1, hardpointCount - 1)) * .44
    const x = rearX + t * profile.length
    const side = index % 2 === 0 ? -1 : 1
    const envelope = Math.pow(Math.sin(Math.PI * t), .62)
    const y = centerY + side * Math.max(4, envelope * profile.beam * .54)
    return [x, y] as Point
  })

  const cutouts: Cutout[] = Array.from({ length: profile.cutouts }, (_, index) => {
    const t = .43 + ((index + 1) / (profile.cutouts + 1)) * .25
    const cx = rearX + t * profile.length + (random() - .5) * profile.length * .06
    const side = index % 2 === 0 ? -1 : 1
    const cy = centerY + side * profile.beam * (.17 + random() * .1)
    return {
      cx,
      cy,
      rx: Math.max(3.3, profile.length * (.035 + random() * .018)),
      ry: Math.max(2.1, profile.beam * (.11 + random() * .06)),
      angle: -18 + random() * 36,
    }
  })

  const spinePath = `M ${(rearX + profile.length * .08).toFixed(1)} ${centerY} L ${(noseX - profile.length * .08).toFixed(1)} ${centerY}`

  return {
    hullPath,
    panelLines,
    engineYs,
    engineX,
    engineTrailX,
    hardpoints,
    cutouts,
    spinePath,
    notchCount: notchStations.size,
    profileKey: profile.key,
  }
}

export function ProceduralShipGlyph({ seed, hullId, weaponCount = 0, className = '', label }: ProceduralShipGlyphProps) {
  const geometry = makeGeometry(seed, hullId, weaponCount)
  const classes = `procedural-ship-glyph${className ? ` ${className}` : ''}`
  const maskID = `ship-mask-${hashSeed(`${seed}|${geometry.profileKey}`).toString(16)}`

  return (
    <svg
      className={classes}
      viewBox="0 0 128 84"
      role={label ? 'img' : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
      focusable="false"
      data-hull-id={geometry.profileKey}
      data-notch-count={geometry.notchCount}
      data-cutout-count={geometry.cutouts.length}
    >
      <defs>
        <mask id={maskID} maskUnits="userSpaceOnUse" x="0" y="0" width="128" height="84">
          <rect x="0" y="0" width="128" height="84" fill="white" />
          {geometry.cutouts.map((cutout, index) => (
            <ellipse
              key={`mask-cutout-${index}`}
              cx={cutout.cx}
              cy={cutout.cy}
              rx={cutout.rx}
              ry={cutout.ry}
              transform={`rotate(${cutout.angle.toFixed(1)} ${cutout.cx.toFixed(1)} ${cutout.cy.toFixed(1)})`}
              fill="black"
            />
          ))}
        </mask>
      </defs>

      <g className="procedural-ship-engines">
        {geometry.engineYs.map((y, index) => (
          <g key={`engine-${index}`}>
            <path d={`M ${geometry.engineX.toFixed(1)} ${y.toFixed(1)} L ${geometry.engineTrailX.toFixed(1)} ${y.toFixed(1)}`} />
            <circle cx={geometry.engineX} cy={y} r="2.1" />
          </g>
        ))}
      </g>

      <g mask={`url(#${maskID})`}>
        <path className="procedural-ship-body" d={geometry.hullPath} />
        <g className="procedural-ship-panels">
          {geometry.panelLines.map((path, index) => <path d={path} key={`panel-${index}`} />)}
        </g>
        <g className="procedural-ship-hardpoints">
          {geometry.hardpoints.map(([x, y], index) => <circle cx={x} cy={y} r="1.6" key={`hardpoint-${index}`} />)}
        </g>
        <path className="procedural-ship-spine" d={geometry.spinePath} />
      </g>

      {geometry.cutouts.length > 0 && (
        <g className="procedural-ship-cutouts">
          {geometry.cutouts.map((cutout, index) => (
            <ellipse
              key={`cutout-${index}`}
              cx={cutout.cx}
              cy={cutout.cy}
              rx={cutout.rx}
              ry={cutout.ry}
              transform={`rotate(${cutout.angle.toFixed(1)} ${cutout.cx.toFixed(1)} ${cutout.cy.toFixed(1)})`}
            />
          ))}
        </g>
      )}
    </svg>
  )
}
