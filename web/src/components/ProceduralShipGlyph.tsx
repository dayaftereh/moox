import { createShipGenome, hashSeed, type ShipPrimitiveGene, type ShipVisualGenome } from '../shipVisualGenome'

type Point = readonly [number, number]

type ProceduralShipGlyphProps = {
  seed: string
  hullId?: string
  weaponCount?: number
  className?: string
  label?: string
  genome?: ShipVisualGenome
}

type Cutout = {
  cx: number
  cy: number
  rx: number
  ry: number
  angle: number
}

type ShipGeometry = {
  hullPath: string
  primitivePaths: string[]
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

function pointPath(points: Point[]): string {
  return points.map(([x, y], index) => `${index === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`).join(' ') + ' Z'
}

function primitivePath(primitive: ShipPrimitiveGene, side: -1 | 1, centerX: number, centerY: number, length: number, beam: number): string {
  const x = centerX - length / 2 + primitive.t * length
  const outward = side
  const rootY = centerY + outward * beam * .34
  const wingLength = primitive.length
  const wingWidth = primitive.width
  const sweep = primitive.sweep * wingLength * .18

  if (primitive.kind === 'spike') {
    return pointPath([
      [x - wingLength * .42, rootY],
      [x + wingLength * .54 + sweep, rootY + outward * wingWidth],
      [x + wingLength * .12, rootY + outward * wingWidth * .12],
    ])
  }

  if (primitive.kind === 'pod') {
    return pointPath([
      [x - wingLength * .48, rootY + outward * wingWidth * .22],
      [x, rootY + outward * wingWidth],
      [x + wingLength * .48, rootY + outward * wingWidth * .22],
      [x, rootY - outward * wingWidth * .08],
    ])
  }

  return pointPath([
    [x - wingLength * .48, rootY],
    [x + sweep, rootY + outward * wingWidth * .44],
    [x + wingLength * .56 + sweep, rootY + outward * wingWidth],
    [x + wingLength * .18, rootY + outward * wingWidth * .08],
  ])
}

function makeGeometry(genome: ShipVisualGenome, weaponCount: number): ShipGeometry {
  const centerX = 68
  const centerY = 42
  const rearX = centerX - genome.length / 2
  const noseX = centerX + genome.length / 2
  const segment = genome.length / Math.max(1, genome.stationCount - 1)
  const upper: Point[] = []
  let notchCount = 0

  for (let station = 0; station < genome.stationCount; station += 1) {
    const t = station / Math.max(1, genome.stationCount - 1)
    const x = rearX + t * genome.length
    if (station === genome.stationCount - 1) {
      upper.push([noseX, centerY])
      continue
    }

    const halfWidth = genome.stationWidths[station] ?? Math.max(2.4, genome.beam * .3)
    const notch = genome.notchDepths[station] ?? 0
    if (notch > 0) {
      notchCount += 1
      const shoulderWidth = Math.min(genome.beam * 1.3, halfWidth * 1.12 + genome.beam * .08)
      const shoulderDX = segment * .22
      upper.push([x - shoulderDX, centerY - shoulderWidth])
      upper.push([x, centerY - notch])
      upper.push([x + shoulderDX, centerY - shoulderWidth * .96])
    } else {
      upper.push([x, centerY - halfWidth])
    }
  }

  const lower: Point[] = [...upper].reverse().map(([x, y]) => [x, centerY + (centerY - y)] as Point)
  const hullPath = pointPath([...upper, ...lower])

  const primitivePaths: string[] = []
  genome.primitives.forEach((primitive) => {
    primitivePaths.push(primitivePath(primitive, primitive.side, centerX, centerY, genome.length, genome.beam))
    if (primitive.mirrored) primitivePaths.push(primitivePath(primitive, primitive.side === 1 ? -1 : 1, centerX, centerY, genome.length, genome.beam))
  })

  const panelLines: string[] = []
  for (let index = 0; index < genome.detailCount; index += 1) {
    const t = .2 + ((index + 1) / (genome.detailCount + 1)) * .62
    const x = rearX + t * genome.length
    const envelope = Math.pow(Math.sin(Math.PI * t), .62)
    const half = Math.max(3.5, envelope * genome.beam * (.42 + (index % 2) * .08))
    panelLines.push(`M ${x.toFixed(1)} ${(centerY - half).toFixed(1)} L ${x.toFixed(1)} ${(centerY + half).toFixed(1)}`)
  }

  const engineSpacing = Math.min(9, Math.max(4.5, genome.beam * .34))
  const engineYs = Array.from({ length: genome.engineCount }, (_, index) => centerY + (index - (genome.engineCount - 1) / 2) * engineSpacing)
  const engineX = rearX + Math.max(1.8, genome.length * .018)
  const engineTrailX = Math.max(2, rearX - Math.max(7, genome.length * .085))

  const hardpointCount = Math.min(8, Math.max(0, weaponCount))
  const hardpoints: Point[] = Array.from({ length: hardpointCount }, (_, index) => {
    const t = .38 + (index / Math.max(1, hardpointCount - 1)) * .44
    const x = rearX + t * genome.length
    const side = index % 2 === 0 ? -1 : 1
    const envelope = Math.pow(Math.sin(Math.PI * t), .62)
    const y = centerY + side * Math.max(4, envelope * genome.beam * .54)
    return [x, y] as Point
  })

  const cutouts: Cutout[] = genome.cutouts.map((cutout) => ({
    cx: rearX + cutout.t * genome.length,
    cy: centerY + cutout.side * cutout.offset,
    rx: cutout.rx,
    ry: cutout.ry,
    angle: cutout.angle,
  }))

  return {
    hullPath,
    primitivePaths,
    panelLines,
    engineYs,
    engineX,
    engineTrailX,
    hardpoints,
    cutouts,
    spinePath: `M ${(rearX + genome.length * .08).toFixed(1)} ${centerY} L ${(noseX - genome.length * .08).toFixed(1)} ${centerY}`,
    notchCount,
    profileKey: String(genome.hullId).toLowerCase().replace(/[\s-]+/g, '_'),
  }
}

export function ProceduralShipGlyph({ seed, hullId, weaponCount = 0, className = '', label, genome }: ProceduralShipGlyphProps) {
  const resolvedGenome = genome ?? createShipGenome(seed, hullId ?? 'generic')
  const geometry = makeGeometry(resolvedGenome, weaponCount)
  const classes = `procedural-ship-glyph${className ? ` ${className}` : ''}`
  const maskID = `ship-mask-${hashSeed(`${resolvedGenome.seed}|${geometry.profileKey}`).toString(16)}`

  return (
    <svg
      className={classes}
      viewBox="0 0 136 84"
      role={label ? 'img' : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
      focusable="false"
      data-hull-id={geometry.profileKey}
      data-notch-count={geometry.notchCount}
      data-cutout-count={geometry.cutouts.length}
      data-primitive-count={resolvedGenome.primitives.length}
      data-genome-version={resolvedGenome.version}
    >
      <defs>
        <mask id={maskID} maskUnits="userSpaceOnUse" x="0" y="0" width="136" height="84">
          <rect x="0" y="0" width="136" height="84" fill="white" />
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
        <g className="procedural-ship-primitives">
          {geometry.primitivePaths.map((path, index) => <path d={path} key={`primitive-${index}`} />)}
        </g>
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