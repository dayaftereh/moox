import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const outDir = process.env.MOOX_DIFFICULTY_ART_OUT_DIR
  ? path.resolve(process.env.MOOX_DIFFICULTY_ART_OUT_DIR)
  : path.resolve(here, '../public/assets/new-game/difficulty')

const configs = [
  { id: 'easy', ranks: 1, primary: '#78c7e8', secondary: '#bfe9ff', glow: '#2b7196', spikes: 0, broken: false },
  { id: 'normal', ranks: 2, primary: '#87d0ee', secondary: '#d9f2ff', glow: '#317ca0', spikes: 0, broken: false },
  { id: 'hard', ranks: 3, primary: '#e1a449', secondary: '#ffe4a1', glow: '#9a6328', spikes: 3, broken: false },
  { id: 'very_hard', assetId: 'very-hard', ranks: 4, primary: '#dc7650', secondary: '#ffd0a0', glow: '#8e3f31', spikes: 5, broken: false },
  { id: 'impossible', ranks: 5, primary: '#e86f68', secondary: '#ffe4ca', glow: '#8b355d', spikes: 7, broken: true },
]

const starField = [
  [88,92,2],[176,148,1.5],[278,74,2.2],[382,132,1.4],[492,72,1.8],[604,111,2.1],[721,76,1.4],[834,139,1.9],[944,84,1.6],[1092,132,2.3],
  [128,266,1.6],[244,331,2.1],[355,245,1.3],[472,306,2.4],[716,298,1.5],[832,245,2.2],[958,318,1.5],[1098,259,1.9],
  [96,520,2.2],[214,585,1.3],[342,505,1.7],[448,596,2.0],[744,572,1.4],[856,516,2.2],[982,589,1.3],[1110,506,2.1],
]

function rankMarks(count) {
  const lines = []
  const startY = 414 - (count - 1) * 16
  for (let i = 0; i < count; i++) {
    const y = startY + i * 32
    const inset = Math.min(22, i * 4)
    lines.push(`<path d="M${486 + inset} ${y} L600 ${y + 54} L${714 - inset} ${y}" fill="none" stroke="url(#accent)" stroke-width="14" stroke-linecap="round" stroke-linejoin="round"/>`)
  }
  return lines.join('\n      ')
}

function crownSpikes(count) {
  if (!count) return ''
  const cx = 600
  const span = 150
  const points = []
  for (let i = 0; i < count; i++) {
    const t = count === 1 ? 0.5 : i / (count - 1)
    const x = cx - span / 2 + span * t
    const height = 28 + (1 - Math.abs(t - 0.5) * 1.6) * 38
    points.push(`<path d="M${x - 13} 248 L${x} ${248 - height} L${x + 13} 248" fill="none" stroke="url(#accent)" stroke-width="10" stroke-linejoin="round"/>`)
  }
  return points.join('\n      ')
}

function outerRing(config) {
  if (!config.broken) {
    return '<ellipse cx="600" cy="335" rx="300" ry="250" fill="none" stroke="url(#ring)" stroke-width="3" opacity="0.6"/>'
  }
  return [
    '<path d="M390 154 A300 250 0 0 1 790 145" fill="none" stroke="url(#ring)" stroke-width="5" stroke-linecap="round"/>',
    '<path d="M866 208 A300 250 0 0 1 851 500" fill="none" stroke="url(#ring)" stroke-width="5" stroke-linecap="round"/>',
    '<path d="M780 548 A300 250 0 0 1 425 545" fill="none" stroke="url(#ring)" stroke-width="5" stroke-linecap="round"/>',
    '<path d="M345 482 A300 250 0 0 1 336 245" fill="none" stroke="url(#ring)" stroke-width="5" stroke-linecap="round"/>',
  ].join('\n      ')
}

function render(config) {
  const stars = starField.map(([x,y,r], i) => `<circle cx="${x}" cy="${y}" r="${r}" fill="${i % 4 === 0 ? config.secondary : '#7d9ab1'}" opacity="${i % 3 === 0 ? '0.72' : '0.42'}"/>`).join('\n    ')
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="675" viewBox="0 0 1200 675" role="img" aria-labelledby="title desc">
  <title id="title">MOX difficulty ${config.id} command crest</title>
  <desc id="desc">Original vector command crest for the ${config.id} difficulty option, using a shared shield family with ${config.ranks} rank marks.</desc>
  <defs>
    <radialGradient id="background" cx="50%" cy="45%" r="72%"><stop offset="0" stop-color="#102033"/><stop offset="0.56" stop-color="#07111d"/><stop offset="1" stop-color="#02060b"/></radialGradient>
    <linearGradient id="accent" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${config.secondary}"/><stop offset="0.52" stop-color="${config.primary}"/><stop offset="1" stop-color="${config.glow}"/></linearGradient>
    <radialGradient id="core" cx="50%" cy="38%" r="62%"><stop offset="0" stop-color="${config.secondary}" stop-opacity="0.24"/><stop offset="0.55" stop-color="${config.primary}" stop-opacity="0.08"/><stop offset="1" stop-color="#07111d" stop-opacity="0"/></radialGradient>
    <linearGradient id="ring" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${config.primary}" stop-opacity="0.15"/><stop offset="0.5" stop-color="${config.secondary}" stop-opacity="0.8"/><stop offset="1" stop-color="${config.glow}" stop-opacity="0.18"/></linearGradient>
    <filter id="glow" x="-80%" y="-80%" width="260%" height="260%"><feGaussianBlur stdDeviation="13" result="b"/><feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge></filter>
  </defs>
  <rect width="1200" height="675" fill="url(#background)"/>
  <g>${stars}</g>
  <ellipse cx="600" cy="335" rx="370" ry="290" fill="url(#core)"/>
  <g opacity="0.9">${outerRing(config)}</g>
  <g filter="url(#glow)" opacity="0.5"><path d="M600 164 L790 242 L758 435 Q718 522 600 575 Q482 522 442 435 L410 242 Z" fill="none" stroke="${config.glow}" stroke-width="22"/></g>
  <g>
    <path d="M600 164 L790 242 L758 435 Q718 522 600 575 Q482 522 442 435 L410 242 Z" fill="#08131f" fill-opacity="0.86" stroke="url(#accent)" stroke-width="14" stroke-linejoin="round"/>
    <path d="M600 194 L748 255 L724 413 Q693 480 600 525 Q507 480 476 413 L452 255 Z" fill="none" stroke="${config.primary}" stroke-opacity="0.35" stroke-width="4"/>
    ${crownSpikes(config.spikes)}
    <circle cx="600" cy="326" r="78" fill="#06111b" stroke="url(#accent)" stroke-width="12"/>
    <path d="M555 326 L584 297 L600 319 L616 297 L645 326 L624 363 L576 363 Z" fill="none" stroke="url(#accent)" stroke-width="10" stroke-linejoin="round"/>
    <circle cx="600" cy="326" r="18" fill="${config.secondary}" opacity="0.82"/>
    ${rankMarks(config.ranks)}
  </g>
  <g opacity="0.52"><path d="M265 336 H390 M810 336 H935" stroke="${config.primary}" stroke-width="3" stroke-dasharray="12 18"/><circle cx="247" cy="336" r="5" fill="${config.secondary}"/><circle cx="953" cy="336" r="5" fill="${config.secondary}"/></g>
</svg>
`
}

fs.mkdirSync(outDir, { recursive: true })
for (const config of configs) {
  const svg = render(config).replace(/^[ \t]+$/gm, '')
  fs.writeFileSync(path.join(outDir, `${config.assetId ?? config.id}.svg`), svg, 'utf8')
}
console.log(`Generated ${configs.length} deterministic Difficulty SVG assets in ${outDir}`)
