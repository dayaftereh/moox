import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const outDir = process.env.MOOX_GALAXY_AGE_ART_OUT_DIR
  ? path.resolve(process.env.MOOX_GALAXY_AGE_ART_OUT_DIR)
  : path.resolve(here, '../public/assets/new-game/galaxy-age')
fs.mkdirSync(outDir, { recursive: true })

const configs = [
  { id: 'mineral_rich', assetId: 'mineral-rich', marker: 0.18, mineral: 1.0, organic: 0.14, accent: '#e3aa55', secondary: '#ffe7aa' },
  { id: 'normal', marker: 0.50, mineral: 0.46, organic: 0.46, accent: '#79b8e8', secondary: '#dff4ff' },
  { id: 'organic_rich', assetId: 'organic-rich', marker: 0.82, mineral: 0.14, organic: 1.0, accent: '#6fc8b4', secondary: '#d9fff5' },
]

const stars = [
  [118,104,2],[184,188,1.2],[272,94,1.5],[347,161,1.1],[424,83,1.4],[775,92,1.2],[852,161,1.7],[934,91,1.2],[1038,174,1.8],[1095,106,1.1],
  [104,462,1.4],[196,548,1.1],[302,505,1.7],[406,585,1.2],[792,572,1.3],[892,505,1.7],[1008,562,1.2],[1094,472,1.5],
]

function starField() {
  return stars.map(([x,y,r], i) => `<circle cx="${x}" cy="${y}" r="${r}" fill="${i % 4 === 0 ? '#dfefff' : '#7e98ad'}" opacity="${i % 3 === 0 ? '0.66' : '0.38'}"/>`).join('\n    ')
}

function galaxy(config) {
  return `
    <ellipse cx="600" cy="306" rx="300" ry="170" fill="url(#galaxyGlow)" opacity="0.72"/>
    <path d="M374 317 C448 215 583 208 704 263 C785 301 823 365 786 410" fill="none" stroke="url(#accent)" stroke-width="22" stroke-linecap="round" opacity="0.54"/>
    <path d="M826 295 C750 388 622 410 505 365 C429 336 385 283 414 235" fill="none" stroke="url(#accent)" stroke-width="16" stroke-linecap="round" opacity="0.38"/>
    <path d="M466 324 C523 276 594 263 660 286 C705 301 730 329 722 357" fill="none" stroke="${config.secondary}" stroke-width="8" stroke-linecap="round" opacity="0.52"/>
    <ellipse cx="600" cy="310" rx="82" ry="54" fill="${config.secondary}" opacity="0.10"/>
    <circle cx="600" cy="310" r="18" fill="${config.secondary}" opacity="0.82"/>
  `
}

function mineralGlyph(opacity) {
  return `<g opacity="${opacity}">
    <!-- ore / resource mass -->
    <path d="M224 506 L263 449 L310 460 L333 506 L294 548 L244 538 Z" fill="#241b12" stroke="#e4aa55" stroke-width="8" stroke-linejoin="round"/>
    <path d="M263 449 L294 548 M224 506 L333 506 M263 449 L333 506" fill="none" stroke="#ffe2a0" stroke-width="4" opacity="0.74"/>

    <!-- production gear -->
    <g transform="translate(348 480)">
      <circle r="43" fill="#0b1118" stroke="#d58e38" stroke-width="8"/>
      <circle r="16" fill="#e3aa55" opacity="0.92"/>
      <path d="M0-62 V-42 M0 42 V62 M-62 0 H-42 M42 0 H62 M-44-44 L-30-30 M30 30 L44 44 M44-44 L30-30 M-30 30 L-44 44" fill="none" stroke="#ffe7aa" stroke-width="10" stroke-linecap="square"/>
      <path d="M0-35 V35 M-35 0 H35 M-25-25 L25 25 M25-25 L-25 25" fill="none" stroke="#ffe7aa" stroke-width="6" opacity="0.82"/>
    </g>

    <!-- compact conveyor / work line -->
    <path d="M220 568 H395" fill="none" stroke="#8ba0af" stroke-width="10" stroke-linecap="round" opacity="0.86"/>
    <circle cx="246" cy="568" r="12" fill="#101820" stroke="#d9edf8" stroke-width="4"/>
    <circle cx="369" cy="568" r="12" fill="#101820" stroke="#d9edf8" stroke-width="4"/>
    <path d="M272 548 L293 528 L316 548 Z M323 548 L340 531 L359 548 Z" fill="#e3aa55" stroke="#ffe7aa" stroke-width="3" opacity="0.92"/>
    <path d="M211 425 V468 M228 409 V455" stroke="#7f95a6" stroke-width="9" stroke-linecap="round" opacity="0.70"/>
  </g>`
}

function organicGlyph(opacity) {
  return `<g opacity="${opacity}">
    <!-- living / biosphere cell -->
    <circle cx="912" cy="500" r="68" fill="#07191a" stroke="#69cbb5" stroke-width="9"/>
    <path d="M864 512 C886 480 916 468 949 476 C931 495 919 524 920 550 C897 546 878 533 864 512 Z" fill="#75d0a8" opacity="0.86"/>
    <path d="M881 471 C897 460 917 455 936 459" fill="none" stroke="#c9fff1" stroke-width="6" stroke-linecap="round" opacity="0.78"/>

    <!-- biotech molecule network -->
    <path d="M951 451 L980 467 L980 500 L951 517 L922 500 L922 467 Z" fill="#0b2022" stroke="#8be6d4" stroke-width="5" opacity="0.96"/>
    <path d="M951 451 L951 430 M980 467 L1001 454 M980 500 L1002 513" fill="none" stroke="#c9fff1" stroke-width="5" stroke-linecap="round"/>
    <circle cx="951" cy="428" r="8" fill="#c9fff1"/>
    <circle cx="1005" cy="451" r="8" fill="#8be6d4"/>
    <circle cx="1006" cy="516" r="8" fill="#8d82e8"/>

    <!-- laboratory vial -->
    <path d="M842 448 H866 M848 448 V477 L828 529 C823 542 832 553 846 553 H874 C888 553 897 542 892 529 L872 477 V448" fill="#0a2324" stroke="#bdf8eb" stroke-width="6" stroke-linejoin="round"/>
    <path d="M837 523 H884" stroke="#69cbb5" stroke-width="9" opacity="0.82"/>
    <circle cx="852" cy="508" r="6" fill="#d9fff5" opacity="0.9"/>
    <circle cx="869" cy="535" r="5" fill="#8d82e8" opacity="0.85"/>

    <!-- pharma capsule cue -->
    <g transform="translate(963 555) rotate(-18)">
      <rect x="-36" y="-14" width="72" height="28" rx="14" fill="#d9fff5" stroke="#69cbb5" stroke-width="5"/>
      <path d="M0 -14 V14" stroke="#6a70c9" stroke-width="5"/>
      <path d="M-31 0 H-5" stroke="#8d82e8" stroke-width="8" stroke-linecap="round" opacity="0.72"/>
    </g>
  </g>`
}

function compositionDial(config) {
  const x1 = 382, x2 = 818, y = 520
  const markerX = x1 + (x2 - x1) * config.marker
  return `
    <path d="M${x1} ${y} C500 558 700 558 ${x2} ${y}" fill="none" stroke="#4f6a7f" stroke-width="5" stroke-linecap="round" opacity="0.62"/>
    <circle cx="${markerX}" cy="${y + 22}" r="15" fill="${config.secondary}" stroke="${config.accent}" stroke-width="6"/>
  `
}

function render(config) {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="675" viewBox="0 0 1200 675" role="img" aria-labelledby="title desc">
  <title id="title">MOX galaxy age ${config.id} composition illustration</title>
  <desc id="desc">Deterministic illustration using the same galaxy silhouette for every age and abstract industrial-to-biotech composition cues. It is visual language only, not a generated map preview, exact planet ratio, or pharmaceutical gameplay promise.</desc>
  <defs>
    <radialGradient id="background" cx="50%" cy="46%" r="76%"><stop offset="0" stop-color="#102033"/><stop offset="0.58" stop-color="#07111d"/><stop offset="1" stop-color="#02060b"/></radialGradient>
    <radialGradient id="galaxyGlow" cx="50%" cy="50%" r="50%"><stop offset="0" stop-color="${config.secondary}" stop-opacity="0.22"/><stop offset="0.65" stop-color="${config.accent}" stop-opacity="0.07"/><stop offset="1" stop-color="#02060b" stop-opacity="0"/></radialGradient>
    <linearGradient id="accent" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${config.secondary}"/><stop offset="0.55" stop-color="${config.accent}"/><stop offset="1" stop-color="#35546f"/></linearGradient>
  </defs>
  <rect width="1200" height="675" fill="url(#background)"/>
  <g>${starField()}</g>
  <g>${galaxy(config)}</g>
  ${mineralGlyph(config.mineral)}
  ${organicGlyph(config.organic)}
  ${compositionDial(config)}
</svg>
`
}

for (const config of configs) {
  fs.writeFileSync(path.join(outDir, `${config.assetId ?? config.id}.svg`), render(config).replace(/^[ \t]+$/gm, ''), 'utf8')
}
console.log(`Generated ${configs.length} deterministic Galaxy Age SVG assets in ${outDir}`)
