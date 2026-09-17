import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const webRoot = path.resolve(here, '..')
const outDir = process.env.MOOX_OPPONENT_COUNT_ART_OUT_DIR || path.join(webRoot, 'public', 'assets', 'new-game', 'opponent-count')
fs.mkdirSync(outDir, { recursive: true })

const segmentSets = {
  1: ['b', 'c'],
  2: ['a', 'b', 'g', 'e', 'd'],
  3: ['a', 'b', 'g', 'c', 'd'],
  4: ['f', 'g', 'b', 'c'],
  5: ['a', 'f', 'g', 'c', 'd'],
  6: ['a', 'f', 'g', 'e', 'c', 'd'],
  7: ['a', 'b', 'c'],
}
const segmentGeometry = {
  a: [470, 145, 260, 34], g: [470, 320, 260, 34], d: [470, 495, 260, 34],
  f: [438, 175, 34, 145], b: [728, 175, 34, 145], e: [438, 350, 34, 145], c: [728, 350, 34, 145],
}
const stars = Array.from({ length: 52 }, (_, i) => {
  const x = 30 + ((i * 151) % 1140)
  const y = 26 + ((i * 79) % 590)
  const r = 1 + (i % 3)
  return `<circle cx="${x}" cy="${y}" r="${r}" fill="#e8f8ff" opacity="${(0.18 + (i % 5) * 0.1).toFixed(2)}"/>`
}).join('')

function svgFor(count) {
  const segments = segmentSets[count].map((id) => {
    const [x, y, w, h] = segmentGeometry[id]
    return `<rect x="${x}" y="${y}" width="${w}" height="${h}" rx="17" fill="url(#digit)" filter="url(#glow)"/>`
  }).join('')
  const nodes = Array.from({ length: count }, (_, i) => {
    const angle = (-90 + (360 / count) * i) * Math.PI / 180
    const x = (600 + Math.cos(angle) * 430).toFixed(1)
    const y = (338 + Math.sin(angle) * 245).toFixed(1)
    const r = 12 + (i % 3) * 3
    return `<g><circle cx="${x}" cy="${y}" r="${r + 8}" fill="#4edcff" opacity="0.10"/><circle cx="${x}" cy="${y}" r="${r}" fill="#112d43" stroke="#66ddff" stroke-width="3"/><circle cx="${x}" cy="${y}" r="4" fill="#f4fbff"/></g>`
  }).join('')
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="675" viewBox="0 0 1200 675" role="img" data-opponent-count="${count}" data-art-signature="opponent-count-${count}">
  <title>${count} opponent${count === 1 ? '' : 's'}</title>
  <desc>Geometric MOOX opponent-count emblem for ${count} opponent${count === 1 ? '' : 's'}.</desc>
  <defs>
    <radialGradient id="bg" cx="50%" cy="48%" r="75%"><stop stop-color="#183d58"/><stop offset="0.62" stop-color="#0b2235"/><stop offset="1" stop-color="#050c14"/></radialGradient>
    <linearGradient id="digit" x1="0" y1="0" x2="1" y2="1"><stop stop-color="#d9f8ff"/><stop offset="0.48" stop-color="#62dfff"/><stop offset="1" stop-color="#5876ff"/></linearGradient>
    <filter id="glow" x="-35%" y="-35%" width="170%" height="170%"><feGaussianBlur stdDeviation="5" result="b"/><feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge></filter>
  </defs>
  <rect width="1200" height="675" fill="url(#bg)"/>
  <g aria-hidden="true">${stars}</g>
  <ellipse cx="600" cy="338" rx="485" ry="270" fill="none" stroke="#55cfee" stroke-width="2" opacity="0.16"/>
  <ellipse cx="600" cy="338" rx="385" ry="215" fill="none" stroke="#728cff" stroke-width="2" opacity="0.14" transform="rotate(-8 600 338)"/>
  <g id="opponent-nodes-${count}">${nodes}</g>
  <g id="opponent-number-${count}">${segments}</g>
  <path d="M160 592 H1040" stroke="#69dfff" stroke-width="2" opacity="0.12"/>
</svg>`
}

for (let count = 1; count <= 7; count++) fs.writeFileSync(path.join(outDir, `${count}.svg`), svgFor(count))
console.log(`generated 7 opponent-count SVG assets in ${outDir}`)