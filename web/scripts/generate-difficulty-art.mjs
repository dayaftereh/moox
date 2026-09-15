import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const outDir = process.env.MOOX_DIFFICULTY_ART_OUT_DIR
  ? path.resolve(process.env.MOOX_DIFFICULTY_ART_OUT_DIR)
  : path.resolve(here, '../public/assets/new-game/difficulty')

const configs = [
  { id: 'easy', primary: '#79cce9', secondary: '#d4f4ff', deep: '#214d67' },
  { id: 'normal', primary: '#7db9e2', secondary: '#e0f2ff', deep: '#315f82' },
  { id: 'hard', primary: '#e2a34b', secondary: '#ffe4a4', deep: '#7e5521' },
  { id: 'very_hard', assetId: 'very-hard', primary: '#e17a50', secondary: '#ffd3a9', deep: '#783d2c' },
  { id: 'impossible', primary: '#e66168', secondary: '#ffd4d7', deep: '#652338' },
]

const sparseStars = [
  [112,112,1.8],[235,191,1.2],[355,101,1.5],[845,108,1.3],[978,205,1.9],[1081,122,1.2],
  [151,493,1.4],[304,566,1.8],[908,548,1.4],[1056,476,1.8],[425,603,1.1],[785,585,1.2],
]

function starField(config) {
  return sparseStars.map(([x, y, r], index) =>
    `<circle cx="${x}" cy="${y}" r="${r}" fill="${index % 4 === 0 ? config.secondary : '#7590a5'}" opacity="${index % 3 === 0 ? '0.58' : '0.32'}"/>`,
  ).join('\n    ')
}

function easyGlyph() {
  return `
    <path d="M470 380 C485 276 573 220 663 244 C711 257 746 291 760 337" fill="none" stroke="url(#accent)" stroke-width="18" stroke-linecap="round"/>
    <circle cx="600" cy="337" r="52" fill="#06121c" stroke="url(#accent)" stroke-width="12"/>
    <circle cx="760" cy="337" r="10" fill="var(--secondary)"/>`
}

function normalGlyph() {
  return `
    <path d="M600 202 L720 270 L720 404 L600 472 L480 404 L480 270 Z" fill="none" stroke="url(#accent)" stroke-width="18" stroke-linejoin="round"/>
    <circle cx="600" cy="337" r="54" fill="#06121c" stroke="url(#accent)" stroke-width="12"/>`
}

function hardGlyph() {
  return `
    <path d="M600 190 L770 450 L430 450 Z" fill="none" stroke="url(#accent)" stroke-width="20" stroke-linejoin="round"/>
    <path d="M600 270 L653 367 L600 420 L547 367 Z" fill="#08131d" stroke="url(#accent)" stroke-width="12" stroke-linejoin="round"/>`
}

function veryHardGlyph() {
  return `
    <path d="M600 176 L770 337 L600 498 L430 337 Z" fill="none" stroke="url(#accent)" stroke-width="20" stroke-linejoin="round"/>
    <path d="M600 256 L680 337 L600 418 L520 337 Z" fill="#08111a" stroke="url(#accent)" stroke-width="12" stroke-linejoin="round"/>`
}

function impossibleGlyph() {
  return `
    <path d="M600 166 L642 285 L765 337 L642 389 L600 508 L558 389 L435 337 L558 285 Z" fill="url(#accent)" opacity="0.92"/>
    <circle cx="600" cy="337" r="94" fill="#02050a" stroke="var(--secondary)" stroke-width="9"/>
    <circle cx="600" cy="337" r="26" fill="var(--primary)" opacity="0.82"/>`
}

function glyph(config) {
  switch (config.id) {
    case 'easy': return easyGlyph()
    case 'normal': return normalGlyph()
    case 'hard': return hardGlyph()
    case 'very_hard': return veryHardGlyph()
    case 'impossible': return impossibleGlyph()
    default: throw new Error(`unknown Difficulty art id: ${config.id}`)
  }
}

function render(config) {
  const stars = starField(config)
  const symbol = glyph(config)
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="675" viewBox="0 0 1200 675" role="img" aria-labelledby="title desc" style="--primary:${config.primary};--secondary:${config.secondary}">
  <title id="title">MOX difficulty ${config.id} threat sigil</title>
  <desc id="desc">Original minimalist vector threat sigil for the ${config.id} difficulty option.</desc>
  <defs>
    <radialGradient id="background" cx="50%" cy="47%" r="76%">
      <stop offset="0" stop-color="#102033"/>
      <stop offset="0.55" stop-color="#07111d"/>
      <stop offset="1" stop-color="#02060b"/>
    </radialGradient>
    <radialGradient id="aura" cx="50%" cy="50%" r="50%">
      <stop offset="0" stop-color="${config.primary}" stop-opacity="0.16"/>
      <stop offset="0.55" stop-color="${config.deep}" stop-opacity="0.06"/>
      <stop offset="1" stop-color="#02060b" stop-opacity="0"/>
    </radialGradient>
    <linearGradient id="accent" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="${config.secondary}"/>
      <stop offset="0.52" stop-color="${config.primary}"/>
      <stop offset="1" stop-color="${config.deep}"/>
    </linearGradient>
  </defs>
  <rect width="1200" height="675" fill="url(#background)"/>
  <g>${stars}</g>
  <ellipse cx="600" cy="337" rx="330" ry="260" fill="url(#aura)"/>
  <g>${symbol}</g>
</svg>
`
}

fs.mkdirSync(outDir, { recursive: true })
for (const config of configs) {
  const svg = render(config).replace(/^[ \t]+$/gm, '')
  fs.writeFileSync(path.join(outDir, `${config.assetId ?? config.id}.svg`), svg, 'utf8')
}
console.log(`Generated ${configs.length} deterministic Difficulty threat-sigil SVG assets in ${outDir}`)
