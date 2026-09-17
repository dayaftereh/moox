import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const webRoot = path.resolve(here, '..')
const outDir = process.env.MOOX_TECHNOLOGY_ART_OUT_DIR || path.join(webRoot, 'public', 'assets', 'new-game', 'technology-level')
fs.mkdirSync(outDir, { recursive: true })

const specs = {
  'pre-warp': { title: 'Pre-Warp starting technology', desc: 'Early orbital workshop with a single homeworld and no interstellar fleet.', rings: 1, nodes: 2, ships: 0 },
  average: { title: 'Average starting technology', desc: 'Established interstellar command with scouts and a colony vessel around one homeworld.', rings: 2, nodes: 4, ships: 3 },
  advanced: { title: 'Advanced starting technology', desc: 'Planned networked stellar command with a broader multi-node footprint and denser fleet presence.', rings: 3, nodes: 7, ships: 5 },
}

function svg(id, spec) {
  const nodes = Array.from({ length: spec.nodes }, (_, i) => {
    const x = 260 + ((i * 173 + spec.nodes * 41) % 680)
    const y = 170 + ((i * 113 + spec.rings * 59) % 300)
    const r = 10 + (i % 3) * 4
    return `<circle cx="${x}" cy="${y}" r="${r}" fill="url(#node)" opacity="${0.55 + (i % 3) * 0.15}"/>`
  }).join('')
  const ships = Array.from({ length: spec.ships }, (_, i) => {
    const x = 340 + i * 115
    const y = 470 - (i % 2) * 55
    return `<path d="M ${x} ${y} l 32 -12 l -10 12 l 10 12 z" fill="#d9f6ff" opacity="0.86"/>`
  }).join('') || '<!-- no starting ships -->'
  const rings = Array.from({ length: spec.rings }, (_, i) => `<ellipse cx="600" cy="338" rx="${170 + i * 105}" ry="${76 + i * 42}" fill="none" stroke="#74d4ff" stroke-width="${3 - Math.min(i, 1)}" opacity="${0.33 - i * 0.06}"/>`).join('')
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="675" viewBox="0 0 1200 675" role="img">
  <title>${spec.title}</title>
  <desc>${spec.desc}</desc>
  <defs>
    <radialGradient id="bg"><stop offset="0" stop-color="#153a55"/><stop offset="1" stop-color="#07111d"/></radialGradient>
    <radialGradient id="core"><stop offset="0" stop-color="#f3fbff"/><stop offset="0.32" stop-color="#70d9ff"/><stop offset="1" stop-color="#0b4f76"/></radialGradient>
    <radialGradient id="node"><stop offset="0" stop-color="#fff4bd"/><stop offset="1" stop-color="#45bde9"/></radialGradient>
    <linearGradient id="deck" x1="0" x2="1"><stop stop-color="#18364b"/><stop offset="1" stop-color="#0b1f30"/></linearGradient>
  </defs>
  <rect width="1200" height="675" fill="url(#bg)"/>
  <g opacity="0.45">${Array.from({length:34},(_,i)=>`<circle cx="${35 + (i*97)%1130}" cy="${28 + (i*61)%590}" r="${1 + i%3}" fill="#d9f6ff"/>`).join('')}</g>
  <path d="M80 570 C260 500 940 500 1120 570 L1120 675 L80 675 Z" fill="url(#deck)"/>
  ${rings}
  <circle cx="600" cy="338" r="88" fill="url(#core)"/>
  <circle cx="600" cy="338" r="118" fill="none" stroke="#e0f7ff" stroke-width="4" opacity="0.52"/>
  ${nodes}
  ${ships}
  <path d="M238 568 H962" stroke="#62c9ef" stroke-width="4" opacity="0.35"/>
  <path d="M300 600 H900" stroke="#9be7ff" stroke-width="2" opacity="0.22"/>
</svg>`
}

for (const [id, spec] of Object.entries(specs)) fs.writeFileSync(path.join(outDir, `${id}.svg`), svg(id, spec))
console.log(`generated ${Object.keys(specs).length} technology-level SVG assets in ${outDir}`)