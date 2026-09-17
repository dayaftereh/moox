import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const webRoot = path.resolve(here, '..')
const outDir = process.env.MOOX_TECHNOLOGY_ART_OUT_DIR || path.join(webRoot, 'public', 'assets', 'new-game', 'technology-level')
fs.mkdirSync(outDir, { recursive: true })

const stars = Array.from({ length: 42 }, (_, i) => {
  const x = 28 + ((i * 149) % 1140)
  const y = 24 + ((i * 83) % 520)
  const r = 1 + (i % 3)
  const opacity = (0.25 + (i % 5) * 0.11).toFixed(2)
  return `<circle cx="${x}" cy="${y}" r="${r}" fill="#e8f8ff" opacity="${opacity}"/>`
}).join('')

function frame({ title, desc, id, accentA, accentB, scene }) {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="675" viewBox="0 0 1200 675" role="img" data-technology-level="${id}">
  <title>${title}</title>
  <desc>${desc}</desc>
  <defs>
    <radialGradient id="bg" cx="50%" cy="42%" r="78%"><stop offset="0" stop-color="#183d58"/><stop offset="0.58" stop-color="#0b2235"/><stop offset="1" stop-color="#050c14"/></radialGradient>
    <linearGradient id="accent" x1="0" y1="0" x2="1" y2="1"><stop stop-color="${accentA}"/><stop offset="1" stop-color="${accentB}"/></linearGradient>
    <radialGradient id="glow"><stop offset="0" stop-color="${accentA}" stop-opacity="0.95"/><stop offset="1" stop-color="${accentB}" stop-opacity="0"/></radialGradient>
    <linearGradient id="metal" x1="0" y1="0" x2="1" y2="1"><stop stop-color="#b9d5df"/><stop offset="0.45" stop-color="#476b7c"/><stop offset="1" stop-color="#162c3a"/></linearGradient>
    <filter id="softGlow" x="-40%" y="-40%" width="180%" height="180%"><feGaussianBlur stdDeviation="14"/></filter>
  </defs>
  <rect width="1200" height="675" fill="url(#bg)"/>
  <g aria-hidden="true">${stars}</g>
  <ellipse cx="600" cy="310" rx="420" ry="250" fill="url(#glow)" opacity="0.12" filter="url(#softGlow)"/>
  ${scene.trimStart()}
  <path d="M0 645 H1200" stroke="${accentA}" stroke-width="2" opacity="0.12"/>
</svg>`
}

function preWarp() {
  const scene = `
  <g id="scene-pre-warp" data-art-signature="surface-launchpad">
    <circle cx="1000" cy="150" r="46" fill="#89a6b0" opacity="0.62"/>
    <circle cx="1000" cy="150" r="59" fill="none" stroke="#c5dae1" stroke-width="2" opacity="0.18"/>
    <circle cx="600" cy="760" r="330" fill="#172c34"/>
    <circle cx="600" cy="760" r="326" fill="url(#accent)" opacity="0.26"/>
    <path d="M0 587 C230 542 370 550 520 579 C690 612 900 554 1200 585 L1200 675 H0 Z" fill="#0a1821"/>
    <path d="M0 587 C230 542 370 550 520 579 C690 612 900 554 1200 585" fill="none" stroke="#f0a14d" stroke-width="4" opacity="0.42"/>
    <g id="pre-warp-launchpad" transform="translate(220 300)">
      <rect x="0" y="195" width="405" height="30" rx="8" fill="#122a37" stroke="#f0a14d" stroke-width="3" opacity="0.95"/>
      <rect x="62" y="60" width="70" height="135" rx="8" fill="url(#metal)"/>
      <rect x="76" y="20" width="42" height="46" rx="5" fill="#284c5b" stroke="#f0a14d" stroke-width="3"/>
      <path d="M97 20 V-22 M79 -10 H115" stroke="#f0a14d" stroke-width="5" stroke-linecap="round"/>
      <path d="M190 190 V72 L228 20 L266 72 V190 Z" fill="#173341" stroke="#e9b46d" stroke-width="3"/>
      <path d="M228 36 L247 86 H209 Z" fill="#e9b46d" opacity="0.7"/>
      <rect x="310" y="110" width="54" height="80" rx="6" fill="#244652"/>
      <path d="M337 110 V72" stroke="#d8edf3" stroke-width="4"/>
      <path d="M307 72 Q337 40 367 72 Q337 90 307 72 Z" fill="#7f9eaa" stroke="#d8edf3" stroke-width="3"/>
    </g>
    <g id="pre-warp-orbital-satellite" transform="translate(840 280) rotate(-12)">
      <rect x="-20" y="-12" width="40" height="24" rx="4" fill="#a7c2cc"/>
      <rect x="-72" y="-18" width="44" height="36" fill="#356479" stroke="#f0a14d" stroke-width="2"/>
      <rect x="28" y="-18" width="44" height="36" fill="#356479" stroke="#f0a14d" stroke-width="2"/>
      <path d="M0 -12 V-42" stroke="#d7edf5" stroke-width="3"/>
      <circle cx="0" cy="-46" r="6" fill="#f0a14d"/>
    </g>
    <path d="M720 225 C810 168 900 164 982 188" fill="none" stroke="#f0a14d" stroke-width="2" stroke-dasharray="9 12" opacity="0.36"/>
  </g>`
  return frame({
    id: 'pre-warp',
    title: 'Pre-Warp starting technology',
    desc: 'Surface launch complex and primitive orbital infrastructure with no interstellar starting fleet.',
    accentA: '#f0a14d',
    accentB: '#7b3f25',
    scene,
  })
}

function average() {
  const scene = `
  <g id="scene-average" data-art-signature="orbital-fleet">
    <circle cx="255" cy="365" r="142" fill="#173b54" stroke="#5bc7e8" stroke-width="4"/>
    <path d="M135 345 Q255 280 375 338 Q330 410 255 462 Q170 432 135 345 Z" fill="#2c6d72" opacity="0.54"/>
    <ellipse cx="255" cy="365" rx="185" ry="64" fill="none" stroke="#9ee9ff" stroke-width="3" opacity="0.32" transform="rotate(-14 255 365)"/>
    <g id="average-orbital-station" transform="translate(640 334)">
      <circle r="92" fill="#0d2636" stroke="#67d5f4" stroke-width="5"/>
      <circle r="58" fill="none" stroke="#b8f2ff" stroke-width="9" opacity="0.78"/>
      <circle r="24" fill="url(#accent)"/>
      <path d="M0 -118 V-58 M0 58 V118 M-118 0 H-58 M58 0 H118" stroke="#a8e8f8" stroke-width="11" stroke-linecap="round"/>
      <path d="M-82 -82 L-42 -42 M82 -82 L42 -42 M-82 82 L-42 42 M82 82 L42 42" stroke="#4fb7d6" stroke-width="8"/>
    </g>
    <g id="average-scout-a" transform="translate(480 205) rotate(9)">
      <path d="M-50 0 L14 -24 L52 0 L14 24 Z" fill="#bfefff" stroke="#52c9ec" stroke-width="3"/>
      <path d="M-28 0 H32" stroke="#193746" stroke-width="6"/>
    </g>
    <g id="average-scout-b" transform="translate(845 230) rotate(-13)">
      <path d="M-50 0 L14 -24 L52 0 L14 24 Z" fill="#bfefff" stroke="#52c9ec" stroke-width="3"/>
      <path d="M-28 0 H32" stroke="#193746" stroke-width="6"/>
    </g>
    <g id="average-colony-ship" transform="translate(835 455) rotate(-7)">
      <path d="M-95 0 L-38 -42 H48 L98 0 L48 42 H-38 Z" fill="url(#metal)" stroke="#9ee9ff" stroke-width="4"/>
      <ellipse cx="7" cy="0" rx="34" ry="22" fill="#5bc7e8" opacity="0.68"/>
      <path d="M-76 -14 H-112 M-76 14 H-112" stroke="#bdefff" stroke-width="8" stroke-linecap="round"/>
    </g>
    <path d="M354 338 Q475 310 545 323 M735 315 Q790 283 840 251 M722 394 Q770 428 806 443" fill="none" stroke="#5bc7e8" stroke-width="3" stroke-dasharray="8 10" opacity="0.42"/>
  </g>`
  return frame({
    id: 'average',
    title: 'Average starting technology',
    desc: 'Established orbital station with two scouts and one colony ship representing the standard interstellar start.',
    accentA: '#5bc7e8',
    accentB: '#17658a',
    scene,
  })
}

function advanced() {
  const links = [
    [600, 330, 310, 210], [600, 330, 915, 195], [600, 330, 990, 430], [600, 330, 300, 470],
    [310, 210, 180, 350], [915, 195, 1030, 310], [300, 470, 520, 535], [990, 430, 760, 520],
  ].map(([x1, y1, x2, y2]) => `<path d="M${x1} ${y1} L${x2} ${y2}" stroke="#b89cff" stroke-width="4" opacity="0.46"/>`).join('')
  const nodes = [
    [310, 210, 42], [915, 195, 32], [990, 430, 38], [300, 470, 34], [180, 350, 24], [1030, 310, 24], [520, 535, 26], [760, 520, 28],
  ].map(([x, y, r], i) => `<g transform="translate(${x} ${y})"><circle r="${r + 16}" fill="url(#glow)" opacity="0.28"/><circle r="${r}" fill="#171f43" stroke="#c5b4ff" stroke-width="4"/><circle r="${Math.max(9, r - 18)}" fill="${i % 2 ? '#63e7ff' : '#d6b8ff'}" opacity="0.82"/></g>`).join('')
  const scene = `
  <g id="scene-advanced" data-art-signature="hyperlane-network">
    <ellipse cx="600" cy="330" rx="475" ry="255" fill="none" stroke="#7b65d5" stroke-width="2" opacity="0.28"/>
    <ellipse cx="600" cy="330" rx="390" ry="205" fill="none" stroke="#63e7ff" stroke-width="2" opacity="0.18" transform="rotate(-8 600 330)"/>
    ${links}
    ${nodes}
    <g id="advanced-hypernet-core" transform="translate(600 330)">
      <circle r="118" fill="url(#glow)" opacity="0.52" filter="url(#softGlow)"/>
      <path d="M0 -78 L68 -39 L68 39 L0 78 L-68 39 L-68 -39 Z" fill="#101a38" stroke="#d0c2ff" stroke-width="7"/>
      <path d="M0 -46 L40 -23 L40 23 L0 46 L-40 23 L-40 -23 Z" fill="url(#accent)" opacity="0.9"/>
      <circle r="16" fill="#f2ecff"/>
    </g>
    <g id="advanced-fleet" fill="#e5dcff" stroke="#8ceaff" stroke-width="2">
      <path d="M420 125 l42 -14 l34 22 l-38 14 z"/>
      <path d="M735 118 l45 -18 l30 26 l-44 12 z"/>
      <path d="M1080 215 l36 -12 l28 19 l-34 12 z"/>
      <path d="M1080 505 l38 -13 l30 20 l-36 13 z"/>
      <path d="M105 505 l40 -14 l31 21 l-38 14 z"/>
    </g>
    <circle cx="112" cy="138" r="54" fill="#263b6b" stroke="#9b8bf0" stroke-width="3" opacity="0.78"/>
    <circle cx="112" cy="138" r="30" fill="#4f6a8b" opacity="0.68"/>
  </g>`
  return frame({
    id: 'advanced',
    title: 'Advanced starting technology',
    desc: 'Dense hyperlane network, multiple developed nodes and an expanded fleet representing a broader advanced start.',
    accentA: '#c2a8ff',
    accentB: '#5d42c8',
    scene,
  })
}

const outputs = {
  'pre-warp': preWarp(),
  average: average(),
  advanced: advanced(),
}

for (const [id, svg] of Object.entries(outputs)) fs.writeFileSync(path.join(outDir, `${id}.svg`), svg)
console.log(`generated ${Object.keys(outputs).length} distinct technology-level SVG assets in ${outDir}`)