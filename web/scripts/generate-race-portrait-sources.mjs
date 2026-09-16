import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(here, '..', '..')
const defaultOut = path.join(repoRoot, 'docs', 'research', 'prototypes', 'SLICE_16_4_RUNTIME_PORTRAITS_2026-09-16')
const outDir = process.argv[2] ? path.resolve(process.argv[2]) : defaultOut
fs.mkdirSync(outDir, { recursive: true })

const W = 1200
const H = 1500

const configs = {
  human: { accent: '#79c9ef', rim: '#f1c6a2', secondary: '#304c66' },
  klackon: { accent: '#a4d36f', rim: '#e2b35d', secondary: '#35452c' },
  darlok: { accent: '#8ddbe0', rim: '#c68cff', secondary: '#2c2642' },
}

function defs(c) {
  return `<defs>
    <radialGradient id="bg" cx="50%" cy="34%" r="78%"><stop offset="0" stop-color="#152536"/><stop offset="0.52" stop-color="#07111c"/><stop offset="1" stop-color="#02050a"/></radialGradient>
    <radialGradient id="halo" cx="50%" cy="40%" r="52%"><stop offset="0" stop-color="${c.accent}" stop-opacity="0.18"/><stop offset="1" stop-color="${c.accent}" stop-opacity="0"/></radialGradient>
    <linearGradient id="rim" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${c.rim}"/><stop offset="0.55" stop-color="${c.accent}"/><stop offset="1" stop-color="${c.secondary}"/></linearGradient>
    <linearGradient id="metal" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#61778a"/><stop offset="0.45" stop-color="#1d2b37"/><stop offset="1" stop-color="#080e15"/></linearGradient>
    <filter id="soft"><feGaussianBlur stdDeviation="18"/></filter>
    <filter id="glow"><feGaussianBlur stdDeviation="7"/></filter>
  </defs>`
}

function common(c) {
  const stars = [[91,154,2],[185,276,1.5],[1027,185,1.6],[1112,344,2],[138,665,1.4],[1068,700,1.5],[211,1054,1.2],[1038,1118,1.7],[321,1287,1.3],[918,1317,1.2]]
  return `<rect width="${W}" height="${H}" fill="url(#bg)"/>
    <ellipse cx="600" cy="625" rx="455" ry="525" fill="url(#halo)"/>
    <g opacity="0.46">${stars.map(([x,y,r])=>`<circle cx="${x}" cy="${y}" r="${r}" fill="#d9efff"/>`).join('')}</g>
    <path d="M150 1260 C330 1370 870 1370 1050 1260" fill="none" stroke="#31465b" stroke-width="3" opacity="0.38"/>
    <path d="M212 1320 C390 1405 810 1405 988 1320" fill="none" stroke="${c.accent}" stroke-width="2" opacity="0.16"/>`
}

function human(c) {
  return `<g>
    <path d="M258 1324 C300 1115 390 975 500 925 L700 925 C810 975 900 1115 942 1324 Z" fill="#111c29" stroke="url(#rim)" stroke-width="9"/>
    <path d="M420 1034 L500 925 H700 L782 1034 L715 1245 H485 Z" fill="#1c3041" stroke="#557a94" stroke-width="6"/>
    <path d="M526 948 C494 855 492 792 520 720 H680 C708 792 706 855 674 948 C642 982 558 982 526 948 Z" fill="#b98268" stroke="#efc0a4" stroke-width="7"/>
    <path d="M438 505 C450 380 512 301 600 294 C688 301 750 380 762 505 L731 647 C709 732 659 790 600 793 C541 790 491 732 469 647 Z" fill="#b98268" stroke="#f0c8ae" stroke-width="8"/>
    <path d="M462 498 C484 352 532 274 600 274 C668 274 716 352 738 498 C698 454 655 434 600 430 C545 434 502 454 462 498 Z" fill="#2c2526" stroke="#83665d" stroke-width="5"/>
    <path d="M500 526 Q548 493 583 526 Q548 559 505 546 Z" fill="#16232b" stroke="#c9ecff" stroke-width="4"/>
    <path d="M700 526 Q652 493 617 526 Q652 559 695 546 Z" fill="#16232b" stroke="#c9ecff" stroke-width="4"/>
    <circle cx="548" cy="532" r="10" fill="#72d9f8"/><circle cx="652" cy="532" r="10" fill="#72d9f8"/>
    <path d="M566 642 Q600 661 634 642" fill="none" stroke="#6f443b" stroke-width="8" stroke-linecap="round"/>
    <path d="M497 930 L600 1024 L703 930" fill="#d9e5ec" opacity="0.88"/>
    <path d="M600 1024 V1244" stroke="#79c9ef" stroke-width="6" opacity="0.5"/>
    <g transform="translate(865 440)" opacity="0.55"><circle r="70" fill="none" stroke="#79c9ef" stroke-width="5"/><path d="M-50 20 C-20 -22 20 -22 50 20 M0-70 V70" fill="none" stroke="#d7f3ff" stroke-width="5"/><circle cy="-22" r="13" fill="#79c9ef"/></g>
  </g>`
}

function klackon(c) {
  return `<g>
    <path d="M220 1320 C262 1110 340 974 452 902 L748 902 C860 974 938 1110 980 1320 Z" fill="#101812" stroke="url(#rim)" stroke-width="10"/>
    <path d="M282 1140 L352 932 L465 887 L416 1198 Z" fill="#263524" stroke="#87a862" stroke-width="7"/>
    <path d="M918 1140 L848 932 L735 887 L784 1198 Z" fill="#263524" stroke="#87a862" stroke-width="7"/>
    <path d="M506 930 C466 815 471 743 516 664 H684 C729 743 734 815 694 930 C648 963 552 963 506 930 Z" fill="#554a2c" stroke="#e0ba63" stroke-width="8"/>
    <path d="M432 496 C456 365 514 294 600 286 C686 294 744 365 768 496 L736 672 C704 754 660 798 600 803 C540 798 496 754 464 672 Z" fill="#655732" stroke="#e3bd64" stroke-width="9"/>
    <path d="M456 470 L358 350 M744 470 L842 350 M490 394 L432 238 M710 394 L768 238" fill="none" stroke="#c9e38b" stroke-width="12" stroke-linecap="round"/>
    <path d="M466 505 L530 468 L582 502 L528 546 Z M734 505 L670 468 L618 502 L672 546 Z" fill="#132017" stroke="#b7eb84" stroke-width="5"/>
    <circle cx="525" cy="505" r="12" fill="#a8e86f"/><circle cx="675" cy="505" r="12" fill="#a8e86f"/>
    <path d="M537 611 L600 650 L663 611 L640 699 L600 734 L560 699 Z" fill="#3b3422" stroke="#d4a955" stroke-width="7"/>
    <path d="M548 650 L508 708 M652 650 L692 708" stroke="#e3c87e" stroke-width="9" stroke-linecap="round"/>
    <path d="M420 840 L350 785 M780 840 L850 785" stroke="#81955f" stroke-width="24" stroke-linecap="round"/>
    <g opacity="0.34" stroke="#90c86a" fill="none"><path d="M210 360 H370 L430 420 M990 360 H830 L770 420" stroke-width="5"/><circle cx="200" cy="360" r="12"/><circle cx="1000" cy="360" r="12"/></g>
  </g>`
}

function darlok(c) {
  return `<g>
    <path d="M248 1322 C288 1118 380 975 500 910 L700 910 C820 975 912 1118 952 1322 Z" fill="#0e1420" stroke="url(#rim)" stroke-width="9"/>
    <path d="M318 1222 C390 1038 456 958 526 912 L674 912 C744 958 810 1038 882 1222" fill="#29213a" stroke="#7a6b9c" stroke-width="8" opacity="0.8"/>
    <path d="M500 940 C470 845 474 770 512 684 H688 C726 770 730 845 700 940 C660 976 540 976 500 940 Z" fill="#505b65" stroke="#9ed8dc" stroke-width="7" opacity="0.9"/>
    <path d="M430 510 C447 378 514 300 600 296 C686 300 753 378 770 510 L736 676 C703 756 657 798 600 800 C543 798 497 756 464 676 Z" fill="#59636b" stroke="#a7e0df" stroke-width="8" opacity="0.9"/>
    <path d="M452 468 C492 352 542 304 600 302 C665 310 716 365 746 484 C704 445 654 427 600 430 C542 424 494 442 452 468 Z" fill="#1e2230" stroke="#815e9d" stroke-width="6"/>
    <path d="M494 525 Q548 487 586 523 Q547 562 500 548 Z" fill="#101923" stroke="#a9f3ef" stroke-width="4"/><path d="M706 525 Q652 487 614 523 Q653 562 700 548 Z" fill="#101923" stroke="#d4aaff" stroke-width="4"/>
    <circle cx="548" cy="531" r="11" fill="#83e6e2"/><circle cx="652" cy="531" r="11" fill="#c58aff"/>
    <path d="M555 651 Q600 670 645 651" fill="none" stroke="#a2bdc0" stroke-width="7" stroke-linecap="round" opacity="0.68"/>
    <!-- layered unresolved alternate facial planes -->
    <path d="M476 420 C520 350 592 330 662 360 C716 383 747 437 754 515 L724 663 C694 721 653 750 611 756" fill="none" stroke="#84e8e1" stroke-width="5" opacity="0.28" transform="translate(-22 10)"/>
    <path d="M450 540 C500 450 565 423 640 444 C702 462 746 518 754 598" fill="none" stroke="#ca87ff" stroke-width="8" opacity="0.25" transform="translate(25 -5)"/>
    <path d="M475 792 C518 828 682 828 725 792" fill="none" stroke="#b0f4f1" stroke-width="5" opacity="0.25"/>
    <g transform="translate(870 444)" opacity="0.42"><path d="M0-72 L58-34 L48 40 L0 72 L-58 34 L-48-40 Z" fill="none" stroke="#b77cea" stroke-width="6"/><path d="M-54 0 H54 M0-66 V66" stroke="#80ddd9" stroke-width="4"/><circle r="18" fill="#111928" stroke="#d5b3ff" stroke-width="4"/></g>
  </g>`
}

function render(id) {
  const c = configs[id]
  const subject = id === 'human' ? human(c) : id === 'klackon' ? klackon(c) : darlok(c)
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}" role="img" aria-labelledby="title desc">
  <title id="title">MOX ${id} race portrait source</title>
  <desc id="desc">Original MOOX 4:5 race portrait source generated for Slice 16.4. No Master of Orion II portrait pixels are copied.</desc>
  ${defs(c)}
  ${common(c)}
  ${subject}
</svg>`.replace(/^[ \t]+$/gm, '')
}

for (const id of Object.keys(configs)) fs.writeFileSync(path.join(outDir, `${id}.svg`), render(id), 'utf8')
console.log(`Generated ${Object.keys(configs).length} race portrait SVG sources in ${outDir}`)
