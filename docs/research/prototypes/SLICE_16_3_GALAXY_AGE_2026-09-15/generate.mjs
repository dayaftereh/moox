import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const configs = [
  { id: 'mineral_rich', marker: 0.18, mineral: 1.0, organic: 0.28, accent: '#e3aa55', secondary: '#ffe7aa' },
  { id: 'normal', marker: 0.50, mineral: 0.66, organic: 0.66, accent: '#79b8e8', secondary: '#dff4ff' },
  { id: 'organic_rich', marker: 0.82, mineral: 0.28, organic: 1.0, accent: '#6fc8b4', secondary: '#d9fff5' },
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
    <path d="M254 493 L294 440 L350 460 L370 520 L318 558 L265 535 Z" fill="#241b12" stroke="#e4aa55" stroke-width="9" stroke-linejoin="round"/>
    <path d="M294 440 L318 558 M254 493 L370 520 M294 440 L370 520" fill="none" stroke="#ffe2a0" stroke-width="4" opacity="0.72"/>
  </g>`
}

function organicGlyph(opacity) {
  return `<g opacity="${opacity}">
    <circle cx="904" cy="500" r="58" fill="#07191a" stroke="#69cbb5" stroke-width="9"/>
    <path d="M862 510 C884 478 914 466 947 474 C929 493 917 522 918 548 C895 544 876 531 862 510 Z" fill="#75d0a8" opacity="0.82"/>
    <path d="M879 469 C895 458 915 453 934 457" fill="none" stroke="#c9fff1" stroke-width="6" stroke-linecap="round" opacity="0.76"/>
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
  <title id="title">MOX galaxy age ${config.id} composition prototype</title>
  <desc id="desc">Research-only illustration using the same galaxy silhouette for every age and an abstract mineral-to-biosphere composition cue. It is not a generated map preview or an exact planet ratio.</desc>
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
  fs.writeFileSync(path.join(here, `${config.id}.svg`), render(config).replace(/^[ \t]+$/gm, ''), 'utf8')
}
console.log(`Generated ${configs.length} Slice-16.3 Galaxy Age research prototypes in ${here}`)
