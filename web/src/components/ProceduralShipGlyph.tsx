type Point = readonly [number, number]

type ProceduralShipGlyphProps = {
  seed: string
  hullId?: string
  weaponCount?: number
  className?: string
  label?: string
}

type ShipGeometry = {
  hullPath: string
  panelLines: string[]
  engineYs: number[]
  hardpoints: Point[]
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

function hullScale(hullId: string | undefined): number {
  const token = (hullId ?? '').toLowerCase()
  if (token.includes('titan')) return 1.38
  if (token.includes('battleship')) return 1.28
  if (token.includes('cruiser')) return 1.18
  if (token.includes('destroyer')) return 1.08
  if (token.includes('frigate')) return 0.95
  if (token.includes('colony')) return 1.12
  if (token.includes('outpost')) return 1.04
  return 1
}

function pointPath(points: Point[]): string {
  return points.map(([x, y], index) => `${index === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`).join(' ') + ' Z'
}

function makeGeometry(seed: string, hullId: string | undefined, weaponCount: number): ShipGeometry {
  const random = createRandom(hashSeed(`${seed}|${hullId ?? 'unknown'}`))
  const centerY = 36
  const scale = hullScale(hullId)
  const stationCount = 7 + Math.floor(random() * 3)
  const baseWidth = (10 + random() * 8) * scale
  const wingStation = 2 + Math.floor(random() * Math.max(2, stationCount - 4))
  const wingBoost = (5 + random() * 10) * scale
  const upper: Point[] = []

  for (let station = 0; station < stationCount; station += 1) {
    const t = station / (stationCount - 1)
    const x = 12 + t * 92
    if (station === stationCount - 1) {
      upper.push([x, centerY])
      continue
    }
    const envelope = Math.pow(Math.sin(Math.PI * Math.min(0.98, t + 0.04)), 0.62)
    const taper = 0.8 + random() * 0.35
    const wing = station === wingStation ? wingBoost : 0
    const halfWidth = 4.5 + envelope * baseWidth * taper + wing
    upper.push([x, centerY - halfWidth])
  }

  const lower: Point[] = [...upper].reverse().map(([x, y]) => [x, centerY + (centerY - y)] as Point)
  const hullPath = pointPath([...upper, ...lower])

  const panelLines: string[] = []
  const panelCount = 2 + Math.floor(random() * 3)
  for (let index = 0; index < panelCount; index += 1) {
    const t = 0.28 + ((index + 1) / (panelCount + 1)) * 0.5
    const x = 12 + t * 92
    const envelope = Math.pow(Math.sin(Math.PI * t), 0.62)
    const half = Math.max(5, envelope * baseWidth * 0.58)
    panelLines.push(`M ${x.toFixed(1)} ${(centerY - half).toFixed(1)} L ${x.toFixed(1)} ${(centerY + half).toFixed(1)}`)
  }

  const engineCount = scale > 1.2 ? 3 : random() > 0.55 ? 2 : 1
  const engineSpacing = 7 * scale
  const engineYs = Array.from({ length: engineCount }, (_, index) => centerY + (index - (engineCount - 1) / 2) * engineSpacing)

  const hardpointCount = Math.min(6, Math.max(0, weaponCount))
  const hardpoints: Point[] = Array.from({ length: hardpointCount }, (_, index) => {
    const t = 0.42 + (index / Math.max(1, hardpointCount - 1)) * 0.42
    const x = 12 + t * 92
    const side = index % 2 === 0 ? -1 : 1
    const envelope = Math.pow(Math.sin(Math.PI * t), 0.62)
    const y = centerY + side * Math.max(5, envelope * baseWidth * 0.62)
    return [x, y] as Point
  })

  return { hullPath, panelLines, engineYs, hardpoints }
}

export function ProceduralShipGlyph({ seed, hullId, weaponCount = 0, className = '', label }: ProceduralShipGlyphProps) {
  const geometry = makeGeometry(seed, hullId, weaponCount)
  const classes = `procedural-ship-glyph${className ? ` ${className}` : ''}`

  return (
    <svg
      className={classes}
      viewBox="0 0 116 72"
      role={label ? 'img' : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
      focusable="false"
    >
      <g className="procedural-ship-engines">
        {geometry.engineYs.map((y, index) => (
          <g key={`engine-${index}`}>
            <path d={`M 13 ${y.toFixed(1)} L 3 ${y.toFixed(1)}`} />
            <circle cx="13" cy={y} r="2.1" />
          </g>
        ))}
      </g>
      <path className="procedural-ship-body" d={geometry.hullPath} />
      <g className="procedural-ship-panels">
        {geometry.panelLines.map((path, index) => <path d={path} key={`panel-${index}`} />)}
      </g>
      <g className="procedural-ship-hardpoints">
        {geometry.hardpoints.map(([x, y], index) => <circle cx={x} cy={y} r="1.6" key={`hardpoint-${index}`} />)}
      </g>
      <path className="procedural-ship-spine" d="M 18 36 L 99 36" />
    </svg>
  )
}
