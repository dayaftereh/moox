import { createShipGenome, hashSeed, type ShipPrimitiveGene, type ShipVisualGenome } from '../shipVisualGenome'

type Point = readonly [number, number]

type ProceduralShipGlyphProps = {
  seed: string
  hullId?: string
  weaponCount?: number
  className?: string
  label?: string
  genome?: ShipVisualGenome
  footprint?: number
  x?: number
  y?: number
  width?: number
  height?: number
  tightViewBox?: boolean
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
  const rootY = centerY + side * beam * .34
  const wingLength = primitive.length
  const wingWidth = primitive.width
  const sweep = primitive.sweep * wingLength * .18

  if (primitive.kind === 'spike') {
    return pointPath([
      [x - wingLength * .42, rootY],
      [x + wingLength * .54 + sweep, rootY + side * wingWidth],
      [x + wingLength * .12, rootY + side * wingWidth * .12],
    ])
  }

  if (primitive.kind === 'pod') {
    return pointPath([
      [x - wingLength * .48, rootY + side * wingWidth * .22],
      [x, rootY + side * wingWidth],
      [x + wingLength * .48, rootY + side * wingWidth * .22],
      [x, rootY - side * wingWidth * .08],
    ])
  }

  return pointPath([
    [x - wingLength * .48, rootY],
    [x + sweep, rootY + side * wingWidth * .44],
    [x + wingLength * .56 + sweep, rootY + side * wingWidth],
    [x + wingLength * .18, rootY + side * wingWidth * .08],
  ])
}

function makeGeometry(genome: ShipVisualGenome, weaponCount: number, canvas: { width: number; height: number }): ShipGeometry {
  const centerX = canvas.width / 2
  const centerY = canvas.height / 2
  const rearX = centerX - genome.length / 2
  const noseX = centerX + genome.length / 2
  const segment = genome.length / Math.max(1, genome.stationCount - 1)

  // Generate one canonical half only. stationWidths/notchDepths are positive
  // distances from the longitudinal X axis; the lower half is a strict mirror.
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

  const lower: Point[] = [...upper].reverse().map(([x, y]) => [x, 2 * centerY - y] as Point)
  const hullPath = pointPath([...upper, ...lower])

  // Every primitive exists only once in the genome and is always rendered as
  // an exact upper/lower pair. This removes accidental whole-ship asymmetry.
  const primitivePaths = genome.primitives.flatMap((primitive) => [
    primitivePath(primitive, -1, centerX, centerY, genome.length, genome.beam),
    primitivePath(primitive, 1, centerX, centerY, genome.length, genome.beam),
  ])

  const panelLines: string[] = []
  for (let index = 0; index < genome.detailCount; index += 1) {
    const t = .2 + ((index + 1) / (genome.detailCount + 1)) * .62
    const x = rearX + t * genome.length
    const envelope = Math.pow(Math.sin(Math.PI * t), .62)
    const half = Math.max(3.5, envelope * genome.beam * (.42 + (index % 2) * .08))
    panelLines.push(`M ${x.toFixed(1)} ${(centerY - half).toFixed(1)} L ${x.toFixed(1)} ${(centerY + half).toFixed(1)}`)
  }

  // Engine placement is centered around the axis, so even and odd engine
  // counts remain bilaterally symmetric.
  const engineSpacing = Math.min(9, Math.max(4.5, genome.beam * .34))
  const engineYs = Array.from({ length: genome.engineCount }, (_, index) => centerY + (index - (genome.engineCount - 1) / 2) * engineSpacing)
  const engineX = rearX + Math.max(1.8, genome.length * .018)
  const engineTrailX = Math.max(2, rearX - Math.max(7, genome.length * .085))

  // Weapon dots also preserve bilateral symmetry. An odd final marker sits on
  // the centerline rather than breaking the silhouette.
  const hardpointCount = Math.min(8, Math.max(0, weaponCount))
  const pairCount = Math.floor(hardpointCount / 2)
  const hardpoints: Point[] = []
  for (let index = 0; index < pairCount; index += 1) {
    const t = .42 + (index / Math.max(1, pairCount - 1)) * .34
    const x = rearX + t * genome.length
    const envelope = Math.pow(Math.sin(Math.PI * t), .62)
    const offset = Math.max(4, envelope * genome.beam * .54)
    hardpoints.push([x, centerY - offset], [x, centerY + offset])
  }
  if (hardpointCount % 2 === 1) hardpoints.push([rearX + genome.length * .64, centerY])

  // Cutouts are also stored as one half-gene and mirrored. Centerline cutouts
  // (offset 0, e.g. Fork) are emitted once.
  const cutouts: Cutout[] = genome.cutouts.flatMap((cutout) => {
    const cx = rearX + cutout.t * genome.length
    if (Math.abs(cutout.offset) < .001) {
      return [{ cx, cy: centerY, rx: cutout.rx, ry: cutout.ry, angle: 0 }]
    }
    return [
      { cx, cy: centerY - cutout.offset, rx: cutout.rx, ry: cutout.ry, angle: cutout.angle },
      { cx, cy: centerY + cutout.offset, rx: cutout.rx, ry: cutout.ry, angle: -cutout.angle },
    ]
  })

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

export function ProceduralShipGlyph({ seed, hullId, weaponCount = 0, className = '', label, genome, footprint = 1, x, y, width, height, tightViewBox = false }: ProceduralShipGlyphProps) {
  const resolvedGenome = genome ?? createShipGenome(seed, hullId ?? 'generic')
  const canvas = {
    width: Math.ceil(Math.max(180, resolvedGenome.length * 1.65)),
    height: Math.ceil(Math.max(120, resolvedGenome.beam * 4.1)),
  }
  const geometry = makeGeometry(resolvedGenome, weaponCount, canvas)
  const tightWidth = Math.min(canvas.width, Math.max(72, resolvedGenome.length * 1.45))
  const tightHeight = Math.min(canvas.height, Math.max(56, resolvedGenome.beam * 3))
  const viewBox = tightViewBox
    ? `${((canvas.width - tightWidth) / 2).toFixed(2)} ${((canvas.height - tightHeight) / 2).toFixed(2)} ${tightWidth.toFixed(2)} ${tightHeight.toFixed(2)}`
    : `0 0 ${canvas.width} ${canvas.height}`
  const classes = `procedural-ship-glyph${className ? ` ${className}` : ''}`
  const maskID = `ship-mask-${hashSeed(`${resolvedGenome.seed}|${geometry.profileKey}|v4`).toString(16)}`

  return (
    <svg
      className={classes}
      x={x}
      y={y}
      width={width}
      height={height}
      preserveAspectRatio="xMidYMid meet"
      viewBox={viewBox}
      role={label ? 'img' : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
      focusable="false"
      style={tightViewBox ? undefined : { transform: `scale(${footprint})`, transformOrigin: '50% 50%' }}
      data-footprint={footprint.toFixed(2)}
      data-tight-view-box={tightViewBox ? 'true' : undefined}
      data-hull-id={geometry.profileKey}
      data-symmetry="x-axis"
      data-notch-count={geometry.notchCount}
      data-half-cutout-count={resolvedGenome.cutouts.length}
      data-cutout-count={geometry.cutouts.length}
      data-half-primitive-count={resolvedGenome.primitives.length}
      data-primitive-count={geometry.primitivePaths.length}
      data-engine-count={resolvedGenome.engineCount}
      data-genome-version={resolvedGenome.version}
      data-style-id={resolvedGenome.styleId}
      data-morphology-id={resolvedGenome.morphologyId}
    >
      <defs>
        <mask id={maskID} maskUnits="userSpaceOnUse" x="0" y="0" width={canvas.width} height={canvas.height}>
          <rect x="0" y="0" width={canvas.width} height={canvas.height} fill="white" />
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