import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const outDir = path.resolve(here, '../public/assets/new-game/galaxy-size')
fs.mkdirSync(outDir, { recursive: true })

const W = 1200
const H = 675
const CX = W / 2
const CY = H / 2

function rng(seed) {
  let state = seed >>> 0
  return () => {
    state = (state * 1664525 + 1013904223) >>> 0
    return state / 0x100000000
  }
}

function gaussian(random) {
  let u = 0
  let v = 0
  while (u === 0) u = random()
  while (v === 0) v = random()
  return Math.sqrt(-2 * Math.log(u)) * Math.cos(2 * Math.PI * v)
}

function backgroundStars(random, count) {
  const stars = []
  for (let i = 0; i < count; i++) {
    const x = random() * W
    const y = random() * H
    const r = 0.35 + random() * 1.2
    const o = 0.16 + random() * 0.5
    const hue = random() < 0.14 ? '#9fd8ff' : random() < 0.08 ? '#ffd3a1' : '#eef7ff'
    stars.push(`<circle cx="${x.toFixed(1)}" cy="${y.toFixed(1)}" r="${r.toFixed(2)}" fill="${hue}" opacity="${o.toFixed(2)}"/>`)
  }
  return stars.join('\n')
}

function galaxyStar(x, y, radius, color, opacity = 1) {
  return `<circle cx="${x.toFixed(1)}" cy="${y.toFixed(1)}" r="${radius.toFixed(2)}" fill="${color}" opacity="${opacity.toFixed(2)}"/>`
}

function irregularGalaxy(random, cfg) {
  const stars = []
  const knots = []
  for (let i = 0; i < cfg.stars; i++) {
    const angle = random() * Math.PI * 2
    const radial = Math.min(1, Math.abs(gaussian(random)) * 0.42)
    const x = CX + Math.cos(angle) * radial * cfg.rx + gaussian(random) * 16
    const y = CY + Math.sin(angle) * radial * cfg.ry + gaussian(random) * 12
    const p = random()
    const color = p < 0.18 ? '#76cfff' : p < 0.28 ? '#ffd19a' : '#f2f8ff'
    stars.push(galaxyStar(x, y, 0.7 + random() * 2.1, color, 0.5 + random() * 0.5))
  }
  for (let i = 0; i < cfg.knots; i++) {
    const x = CX + gaussian(random) * cfg.rx * 0.42
    const y = CY + gaussian(random) * cfg.ry * 0.38
    const r = 6 + random() * 11
    const color = random() < 0.5 ? '#5bbdff' : '#ff8ac5'
    knots.push(`<circle cx="${x.toFixed(1)}" cy="${y.toFixed(1)}" r="${r.toFixed(1)}" fill="${color}" opacity="0.28" filter="url(#softGlow)"/>`)
  }
  return `
    <ellipse cx="${CX}" cy="${CY}" rx="${cfg.rx * 1.08}" ry="${cfg.ry * 1.08}" fill="url(#dwarfGlow)" opacity="0.62" filter="url(#softBlur)"/>
    <path d="M ${CX - cfg.rx * 0.9} ${CY + 20} C ${CX - cfg.rx * 0.4} ${CY - cfg.ry * 0.75}, ${CX + cfg.rx * 0.15} ${CY + cfg.ry * 0.35}, ${CX + cfg.rx * 0.88} ${CY - 12}" fill="none" stroke="#7bc8ff" stroke-width="18" stroke-linecap="round" opacity="0.08" filter="url(#softBlur)"/>
    ${knots.join('\n')}
    ${stars.join('\n')}
  `
}

function spiralPoints(random, cfg, armIndex, starCount, armStrength = 1, phaseOffset = 0) {
  const stars = []
  const armPhase = (armIndex / cfg.arms) * Math.PI * 2 + phaseOffset
  for (let i = 0; i < starCount; i++) {
    const t = Math.pow(random(), 0.78)
    const theta = armPhase + cfg.turns * Math.PI * 2 * t
    const r = cfg.innerR + (cfg.outerR - cfg.innerR) * t
    const armJitter = gaussian(random) * (cfg.armWidth * (0.55 + 0.75 * t))
    const tangential = gaussian(random) * cfg.tangentJitter
    const baseX = Math.cos(theta) * (r + armJitter)
    const baseY = Math.sin(theta) * (r + armJitter) * cfg.yScale
    const x = CX + baseX - Math.sin(theta) * tangential
    const y = CY + baseY + Math.cos(theta) * tangential * cfg.yScale
    const p = random()
    const color = p < 0.22 * armStrength ? '#75cfff' : p < 0.31 ? '#ffcc96' : '#f3f8ff'
    const size = 0.55 + random() * (1.7 + 1.15 * armStrength)
    stars.push(galaxyStar(x, y, size, color, (0.38 + random() * 0.62) * Math.min(1, 0.68 + armStrength * 0.35)))
  }
  return stars.join('\n')
}

function spiralPath(cfg, armIndex, phaseOffset = 0, radiusScale = 1) {
  const points = []
  const armPhase = (armIndex / cfg.arms) * Math.PI * 2 + phaseOffset
  for (let step = 0; step <= 110; step++) {
    const t = step / 110
    const theta = armPhase + cfg.turns * Math.PI * 2 * t
    const r = (cfg.innerR + (cfg.outerR - cfg.innerR) * t) * radiusScale
    const x = CX + Math.cos(theta) * r
    const y = CY + Math.sin(theta) * r * cfg.yScale
    points.push(`${step === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`)
  }
  return points.join(' ')
}

function spiralGalaxy(random, cfg) {
  const layers = []
  layers.push(`<ellipse cx="${CX}" cy="${CY}" rx="${(cfg.outerR * 1.08).toFixed(1)}" ry="${(cfg.outerR * cfg.yScale * 1.08).toFixed(1)}" fill="url(#diskGlow)" opacity="${cfg.diskOpacity}" filter="url(#softBlur)"/>`)
  if (cfg.bar) {
    layers.push(`<rect x="${CX - cfg.bar * 0.5}" y="${CY - 18}" width="${cfg.bar}" height="36" rx="18" fill="url(#barGlow)" opacity="0.62" transform="rotate(-22 ${CX} ${CY})" filter="url(#softGlow)"/>`)
  }
  for (let arm = 0; arm < cfg.arms; arm++) {
    layers.push(`<path d="${spiralPath(cfg, arm)}" fill="none" stroke="#7fcfff" stroke-width="${cfg.armGlow}" stroke-linecap="round" opacity="${cfg.armOpacity}" filter="url(#armBlur)"/>`)
    layers.push(`<path d="${spiralPath(cfg, arm, 0.03, 0.97)}" fill="none" stroke="#07111d" stroke-width="${Math.max(5, cfg.armGlow * 0.22)}" stroke-linecap="round" opacity="0.48"/>`)
    layers.push(spiralPoints(random, cfg, arm, Math.floor(cfg.stars / cfg.arms), 1))
  }
  for (let branch = 0; branch < (cfg.branches ?? 0); branch++) {
    const phase = 0.28 + branch * 0.63
    const arm = branch % cfg.arms
    layers.push(`<path d="${spiralPath(cfg, arm, phase, 0.88 + (branch % 2) * 0.08)}" fill="none" stroke="#5ab6f4" stroke-width="${Math.max(8, cfg.armGlow * 0.45)}" stroke-linecap="round" opacity="0.10" filter="url(#armBlur)"/>`)
    layers.push(spiralPoints(random, cfg, arm, Math.floor(cfg.stars * 0.07), 0.55, phase))
  }
  for (let i = 0; i < cfg.nebulae; i++) {
    const a = random() * Math.PI * 2
    const r = cfg.innerR + random() * (cfg.outerR - cfg.innerR) * 0.95
    const x = CX + Math.cos(a) * r
    const y = CY + Math.sin(a) * r * cfg.yScale
    const color = random() < 0.62 ? '#ff6eac' : '#55c5ff'
    layers.push(`<circle cx="${x.toFixed(1)}" cy="${y.toFixed(1)}" r="${(4 + random() * 8).toFixed(1)}" fill="${color}" opacity="0.20" filter="url(#softGlow)"/>`)
  }
  layers.push(`<ellipse cx="${CX}" cy="${CY}" rx="${cfg.bulge}" ry="${cfg.bulge * cfg.yScale * 0.82}" fill="url(#coreGlow)" opacity="${cfg.coreOpacity}" filter="url(#softGlow)"/>`)
  layers.push(`<circle cx="${CX}" cy="${CY}" r="${Math.max(7, cfg.bulge * 0.16)}" fill="#fff5d9" opacity="0.88"/>`)
  return layers.join('\n')
}

const configs = {
  tiny: {
    kind: 'irregular', seed: 17, rx: 150, ry: 92, stars: 115, knots: 7, background: 115,
    title: 'Tiny galaxy — compact dwarf irregular', description: 'Sparse, asymmetric dwarf-galaxy illustration with a small footprint and few luminous star-forming knots.'
  },
  small: {
    kind: 'spiral', seed: 31, arms: 2, innerR: 35, outerR: 175, turns: 1.18, yScale: 0.74, armWidth: 17, tangentJitter: 5, stars: 210, nebulae: 9, bar: 0, bulge: 48, armGlow: 26, armOpacity: 0.10, diskOpacity: 0.38, coreOpacity: 0.68, branches: 3, background: 125,
    title: 'Small galaxy — compact flocculent spiral', description: 'Compact spiral with fragmented arms, modest disk extent and a relatively sparse stellar population.'
  },
  medium: {
    kind: 'spiral', seed: 47, arms: 2, innerR: 42, outerR: 245, turns: 1.42, yScale: 0.76, armWidth: 20, tangentJitter: 6, stars: 350, nebulae: 15, bar: 135, bulge: 62, armGlow: 34, armOpacity: 0.12, diskOpacity: 0.46, coreOpacity: 0.76, branches: 4, background: 135,
    title: 'Medium galaxy — barred spiral', description: 'Milky-Way-like barred spiral language with a broader disk, denser stellar population and more visible star-forming structure.'
  },
  large: {
    kind: 'spiral', seed: 67, arms: 2, innerR: 48, outerR: 315, turns: 1.58, yScale: 0.78, armWidth: 23, tangentJitter: 7, stars: 530, nebulae: 22, bar: 150, bulge: 72, armGlow: 43, armOpacity: 0.14, diskOpacity: 0.54, coreOpacity: 0.82, branches: 6, background: 145,
    title: 'Large galaxy — grand-design spiral', description: 'Large luminous disk with two prominent grand-design arms, branching outer structure and a denser field of stars.'
  },
  huge: {
    kind: 'spiral', seed: 101, arms: 2, innerR: 52, outerR: 385, turns: 1.72, yScale: 0.76, armWidth: 26, tangentJitter: 8, stars: 760, nebulae: 32, bar: 175, bulge: 84, armGlow: 52, armOpacity: 0.16, diskOpacity: 0.62, coreOpacity: 0.90, branches: 10, background: 165,
    title: 'Huge galaxy — extended giant spiral', description: 'Extended giant spiral with the broadest disk, brightest core, richest stellar field and extensive outer branches.'
  },
}

function svgFor(id, cfg) {
  const random = rng(cfg.seed)
  const galaxy = cfg.kind === 'irregular' ? irregularGalaxy(random, cfg) : spiralGalaxy(random, cfg)
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}" role="img" aria-labelledby="title desc">
  <title id="title">${cfg.title}</title>
  <desc id="desc">${cfg.description}</desc>
  <defs>
    <radialGradient id="bg" cx="50%" cy="46%" r="72%"><stop offset="0" stop-color="#0b1a2b"/><stop offset="0.46" stop-color="#07111e"/><stop offset="1" stop-color="#02050a"/></radialGradient>
    <radialGradient id="diskGlow"><stop offset="0" stop-color="#8fd9ff" stop-opacity="0.30"/><stop offset="0.42" stop-color="#4a8fc0" stop-opacity="0.22"/><stop offset="1" stop-color="#102a44" stop-opacity="0"/></radialGradient>
    <radialGradient id="dwarfGlow"><stop offset="0" stop-color="#a7dfff" stop-opacity="0.42"/><stop offset="0.5" stop-color="#5387ac" stop-opacity="0.24"/><stop offset="1" stop-color="#12273b" stop-opacity="0"/></radialGradient>
    <radialGradient id="coreGlow"><stop offset="0" stop-color="#fff7dc"/><stop offset="0.19" stop-color="#ffd69b" stop-opacity="0.92"/><stop offset="0.48" stop-color="#8ecfff" stop-opacity="0.54"/><stop offset="1" stop-color="#286087" stop-opacity="0"/></radialGradient>
    <linearGradient id="barGlow" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="#5eaadd" stop-opacity="0"/><stop offset="0.5" stop-color="#fff0cb" stop-opacity="0.92"/><stop offset="1" stop-color="#5eaadd" stop-opacity="0"/></linearGradient>
    <filter id="softBlur" x="-30%" y="-30%" width="160%" height="160%"><feGaussianBlur stdDeviation="18"/></filter>
    <filter id="armBlur" x="-30%" y="-30%" width="160%" height="160%"><feGaussianBlur stdDeviation="9"/></filter>
    <filter id="softGlow" x="-60%" y="-60%" width="220%" height="220%"><feGaussianBlur stdDeviation="4"/></filter>
    <radialGradient id="vignette"><stop offset="58%" stop-color="#000" stop-opacity="0"/><stop offset="100%" stop-color="#000" stop-opacity="0.62"/></radialGradient>
  </defs>
  <rect width="${W}" height="${H}" fill="url(#bg)"/>
  <g>${backgroundStars(random, cfg.background)}</g>
  <g>${galaxy}</g>
  <rect width="${W}" height="${H}" fill="url(#vignette)" pointer-events="none"/>
</svg>
`
}

for (const [id, cfg] of Object.entries(configs)) {
  const target = path.join(outDir, `${id}.svg`)
  fs.writeFileSync(target, svgFor(id, cfg), 'utf8')
  console.log(`generated ${path.relative(process.cwd(), target)}`)
}
