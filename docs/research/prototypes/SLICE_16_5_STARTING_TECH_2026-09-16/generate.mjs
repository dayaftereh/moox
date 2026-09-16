import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))

const configs = [
  {
    id: 'pre_warp',
    title: 'Pre-Warp',
    accent: '#d98b46',
    secondary: '#ffd39b',
    network: 0,
    scouts: 0,
    colony: false,
    orbital: 0.42,
  },
  {
    id: 'average',
    title: 'Average',
    accent: '#5da9d6',
    secondary: '#d5f1ff',
    network: 1,
    scouts: 2,
    colony: true,
    orbital: 0.78,
  },
  {
    id: 'advanced',
    title: 'Advanced',
    accent: '#9a7be8',
    secondary: '#e9ddff',
    network: 3,
    scouts: 5,
    colony: true,
    orbital: 1,
  },
]

const stars = [
  [92,88,1.4],[162,142,1.0],[244,76,1.4],[328,130,1.1],[435,72,1.0],[548,112,1.4],
  [676,68,1.2],[786,132,1.5],[892,83,1.0],[1014,146,1.6],[1110,92,1.1],
  [122,356,1.2],[216,302,1.0],[994,332,1.1],[1096,380,1.5],[1042,546,1.0],[155,552,1.1],
]

function starField() {
  return stars.map(([x,y,r], i) => `<circle cx="${x}" cy="${y}" r="${r}" fill="${i % 4 === 0 ? '#dcecff' : '#7690aa'}" opacity="${i % 3 === 0 ? '0.68' : '0.38'}"/>`).join('\n    ')
}

function planet(config) {
  return `
    <defs>
      <radialGradient id="planetGlow" cx="45%" cy="36%" r="70%">
        <stop offset="0" stop-color="${config.secondary}" stop-opacity="0.28"/>
        <stop offset="0.48" stop-color="${config.accent}" stop-opacity="0.18"/>
        <stop offset="1" stop-color="#07111c" stop-opacity="0"/>
      </radialGradient>
    </defs>
    <circle cx="600" cy="520" r="220" fill="url(#planetGlow)"/>
    <path d="M390 548 C452 474 528 440 610 444 C698 448 777 489 816 550 C744 597 670 620 590 619 C510 618 443 594 390 548 Z" fill="#0a1622" stroke="${config.accent}" stroke-width="5" opacity="0.95"/>
    <path d="M429 545 C493 502 548 492 610 495 C678 497 735 514 784 548" fill="none" stroke="${config.secondary}" stroke-width="5" opacity="0.34"/>
  `
}

function orbital(config) {
  const dash = config.orbital < 0.6 ? '34 28' : config.orbital < 0.9 ? '84 20' : '0'
  return `
    <ellipse cx="600" cy="407" rx="286" ry="102" fill="none" stroke="#31485b" stroke-width="12" opacity="0.52"/>
    <ellipse cx="600" cy="407" rx="286" ry="102" fill="none" stroke="${config.accent}" stroke-width="6" stroke-dasharray="${dash}" opacity="${0.42 + config.orbital * 0.4}"/>
    <g opacity="${0.45 + config.orbital * 0.4}">
      <rect x="570" y="312" width="60" height="86" rx="12" fill="#0c1c2b" stroke="${config.secondary}" stroke-width="5"/>
      <path d="M600 312 L600 260 M574 280 L626 280" stroke="${config.secondary}" stroke-width="6" stroke-linecap="round"/>
      <circle cx="600" cy="257" r="12" fill="${config.secondary}" opacity="0.78"/>
    </g>
  `
}

function scout(x, y, scale, config) {
  return `<g transform="translate(${x} ${y}) scale(${scale})" opacity="0.9">
    <path d="M0 -20 L34 16 L10 12 L0 24 L-10 12 L-34 16 Z" fill="#0d2234" stroke="${config.secondary}" stroke-width="5" stroke-linejoin="round"/>
    <path d="M0 -12 L0 12" stroke="${config.accent}" stroke-width="5" stroke-linecap="round"/>
  </g>`
}

function colonyShip(config) {
  if (!config.colony) return ''
  return `<g transform="translate(850 336)" opacity="0.92">
    <path d="M-58 8 L-18 -28 L44 -24 L68 8 L42 35 L-20 34 Z" fill="#0d2234" stroke="${config.secondary}" stroke-width="5"/>
    <circle cx="8" cy="4" r="22" fill="${config.accent}" opacity="0.36" stroke="${config.secondary}" stroke-width="4"/>
    <path d="M-58 8 L-82 8 M68 8 L88 8" stroke="${config.accent}" stroke-width="6" stroke-linecap="round"/>
  </g>`
}

function fleet(config) {
  const positions = [[356,340,0.72],[438,296,0.62],[755,286,0.58],[938,392,0.52],[292,415,0.48]]
  return positions.slice(0, config.scouts).map(([x,y,s]) => scout(x,y,s,config)).join('\n    ')
}

function network(config) {
  if (config.network === 0) {
    return `<g opacity="0.78">
      <path d="M530 533 L530 477 L556 448 L582 477 L582 533" fill="#102130" stroke="${config.accent}" stroke-width="4"/>
      <path d="M620 534 L620 462 L650 425 L680 462 L680 534" fill="#102130" stroke="${config.accent}" stroke-width="4"/>
      <path d="M708 536 L708 490 L735 466 L760 490 L760 536" fill="#102130" stroke="${config.accent}" stroke-width="4"/>
    </g>`
  }
  const nodes = config.network === 1
    ? [[470,512],[600,477],[730,512]]
    : [[400,510],[505,463],[600,426],[695,463],[800,510]]
  const links = nodes.slice(0,-1).map((n,i) => `<path d="M${n[0]} ${n[1]} L${nodes[i+1][0]} ${nodes[i+1][1]}" stroke="${config.accent}" stroke-width="5" opacity="0.46"/>`).join('\n    ')
  const glyphs = nodes.map(([x,y],i) => `<g><circle cx="${x}" cy="${y}" r="${config.network === 3 ? 28 : 24}" fill="#0d1d2b" stroke="${config.secondary}" stroke-width="5"/><circle cx="${x}" cy="${y}" r="${9 + (i%2)*3}" fill="${config.accent}" opacity="0.54"/></g>`).join('\n    ')
  return `<g>${links}${glyphs}</g>`
}

function advancedHalo(config) {
  if (config.network < 3) return ''
  return `<g opacity="0.72">
    <ellipse cx="600" cy="400" rx="365" ry="142" fill="none" stroke="${config.accent}" stroke-width="3" stroke-dasharray="18 16"/>
    <ellipse cx="600" cy="400" rx="338" ry="125" fill="none" stroke="${config.secondary}" stroke-width="2" opacity="0.55"/>
    <circle cx="600" cy="398" r="58" fill="${config.accent}" opacity="0.12"/>
  </g>`
}

function render(config) {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="675" viewBox="0 0 1200 675" role="img" aria-labelledby="title desc">
  <title id="title">MOOX ${config.title} starting technology prototype</title>
  <desc id="desc">Research-only original MOOX illustration showing a coherent progression from planetary workshop to interstellar launch to networked stellar command. It is not a technology-tree screenshot and promises no fixed Advanced technology list.</desc>
  <defs>
    <radialGradient id="background" cx="50%" cy="44%" r="80%"><stop offset="0" stop-color="#102033"/><stop offset="0.62" stop-color="#07111c"/><stop offset="1" stop-color="#02060b"/></radialGradient>
    <radialGradient id="coreGlow" cx="50%" cy="50%" r="50%"><stop offset="0" stop-color="${config.secondary}" stop-opacity="0.26"/><stop offset="0.58" stop-color="${config.accent}" stop-opacity="0.08"/><stop offset="1" stop-color="#02060b" stop-opacity="0"/></radialGradient>
  </defs>
  <rect width="1200" height="675" rx="40" fill="url(#background)"/>
  <rect x="28" y="28" width="1144" height="619" rx="32" fill="none" stroke="#294054" stroke-width="3" opacity="0.7"/>
  <g>${starField()}</g>
  <ellipse cx="600" cy="370" rx="390" ry="225" fill="url(#coreGlow)"/>
  ${advancedHalo(config)}
  ${planet(config)}
  ${orbital(config)}
  ${network(config)}
  ${fleet(config)}
  ${colonyShip(config)}
  <circle cx="600" cy="394" r="8" fill="${config.secondary}" opacity="0.8"/>
</svg>\n`
}

for (const config of configs) {
  fs.writeFileSync(path.join(here, `${config.id}.svg`), render(config), 'utf8')
}

const mobileRows = configs.map((config, i) => {
  const y = 24 + i * 232
  return `<g transform="translate(15 ${y})"><rect width="360" height="203" rx="18" fill="#07111c" stroke="#294054" stroke-width="2"/><image href="${config.id}.svg" x="0" y="0" width="360" height="203" preserveAspectRatio="xMidYMid slice"/></g>`
}).join('\n  ')

fs.writeFileSync(path.join(here, 'mobile_390_preview.svg'), `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="390" height="720" viewBox="0 0 390 720" role="img" aria-labelledby="title desc">
  <title id="title">MOOX starting technology 390 pixel prototype preview</title>
  <desc id="desc">Research contact sheet of the three technology-start illustrations at a 390 pixel mobile viewport width.</desc>
  <rect width="390" height="720" fill="#02060b"/>
  ${mobileRows}
</svg>\n`, 'utf8')
