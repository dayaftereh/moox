import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const W = 1200, H = 1500

const configs = [
  { id: 'alkari', accent: '#76b9e8', rim: '#ffe2a3', secondary: '#41587a' },
  { id: 'meklar', accent: '#69d5e6', rim: '#d78b4b', secondary: '#593c31' },
  { id: 'silicoid', accent: '#84d9df', rim: '#d47cff', secondary: '#35314b' },
]

function defs(c) {
  return `<defs>
    <radialGradient id="bg" cx="50%" cy="36%" r="78%">
      <stop offset="0" stop-color="#162434"/><stop offset="0.48" stop-color="#08111c"/><stop offset="1" stop-color="#02050a"/>
    </radialGradient>
    <radialGradient id="halo" cx="50%" cy="43%" r="50%">
      <stop offset="0" stop-color="${c.accent}" stop-opacity="0.17"/><stop offset="1" stop-color="${c.accent}" stop-opacity="0"/>
    </radialGradient>
    <linearGradient id="rim" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="${c.rim}"/><stop offset="0.55" stop-color="${c.accent}"/><stop offset="1" stop-color="${c.secondary}"/>
    </linearGradient>
    <linearGradient id="metal" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="#647588"/><stop offset="0.45" stop-color="#1f2a35"/><stop offset="1" stop-color="#080d13"/>
    </linearGradient>
    <filter id="soft"><feGaussianBlur stdDeviation="18"/></filter>
    <filter id="glow"><feGaussianBlur stdDeviation="7"/></filter>
  </defs>`
}

function common(c) {
  const stars = [[98,139,2],[174,273,1.4],[1025,181,1.7],[1105,337,2],[127,655,1.5],[1063,691,1.5],[203,1041,1.2],[1043,1114,1.7],[316,1280,1.4],[921,1310,1.3]]
  return `<rect width="${W}" height="${H}" fill="url(#bg)"/>
    <ellipse cx="600" cy="625" rx="455" ry="525" fill="url(#halo)"/>
    <g opacity="0.48">${stars.map(([x,y,r])=>`<circle cx="${x}" cy="${y}" r="${r}" fill="#d9efff"/>`).join('')}</g>
    <path d="M150 1260 C330 1370 870 1370 1050 1260" fill="none" stroke="#31465b" stroke-width="3" opacity="0.38"/>
    <path d="M212 1320 C390 1405 810 1405 988 1320" fill="none" stroke="${c.accent}" stroke-width="2" opacity="0.16"/>`
}

function alkari(c) {
  return `
    <g>
      <ellipse cx="600" cy="650" rx="325" ry="420" fill="#07101a" stroke="${c.accent}" stroke-width="5" opacity="0.92"/>
      <!-- swept cockpit / gyroscopic flight reference -->
      <ellipse cx="600" cy="625" rx="400" ry="340" fill="none" stroke="#426c91" stroke-width="10" opacity="0.20" transform="rotate(-17 600 625)"/>
      <ellipse cx="600" cy="625" rx="400" ry="340" fill="none" stroke="#426c91" stroke-width="5" opacity="0.12" transform="rotate(23 600 625)"/>
      <!-- shoulders -->
      <path d="M286 1115 C344 978 440 926 520 910 L680 910 C766 928 858 986 914 1115 L1000 1330 H200 Z" fill="#111c28" stroke="url(#rim)" stroke-width="9"/>
      <path d="M318 1120 C378 1036 438 1005 508 985 L436 1188 L264 1270 Z" fill="#253645" stroke="#7ea9c3" stroke-width="6" opacity="0.86"/>
      <path d="M882 1120 C822 1036 762 1005 692 985 L764 1188 L936 1270 Z" fill="#253645" stroke="#7ea9c3" stroke-width="6" opacity="0.86"/>
      <!-- long avian-reptilian neck -->
      <path d="M512 940 C470 835 470 746 510 665 L690 665 C728 750 726 842 682 946 C640 982 554 982 512 940 Z" fill="#7f8f83" stroke="#d7d2aa" stroke-width="7"/>
      <path d="M530 918 C496 822 508 760 536 704 M670 918 C704 822 692 760 664 704" fill="none" stroke="#a9b6a0" stroke-width="10" opacity="0.46"/>
      <!-- head / beak -->
      <path d="M454 480 C478 365 542 300 618 298 C704 300 772 366 782 486 L736 610 C698 680 656 725 590 724 C520 719 478 669 451 592 Z" fill="#95a493" stroke="#e3d8ab" stroke-width="8"/>
      <path d="M514 352 C526 267 564 208 617 173 C642 241 694 274 748 286 C700 329 653 342 592 341 Z" fill="#b5b89b" stroke="#ffe2a3" stroke-width="7"/>
      <path d="M467 514 C530 480 588 472 654 492 C617 532 578 566 512 580 Z" fill="#463c32" stroke="#d8c595" stroke-width="7"/>
      <path d="M654 492 L778 515 L674 560 L606 552 Z" fill="#b99d69" stroke="#f0d39c" stroke-width="8"/>
      <!-- eyes -->
      <path d="M520 439 Q567 405 606 438 Q567 475 523 457 Z" fill="#101a22" stroke="#b9eaff" stroke-width="5"/>
      <circle cx="565" cy="444" r="12" fill="#70d8ff"/><circle cx="565" cy="444" r="26" fill="#70d8ff" opacity="0.13" filter="url(#glow)"/>
      <!-- keratin / feather-scale planes -->
      <path d="M480 600 L432 676 L503 654 M719 602 L765 682 L696 653" fill="none" stroke="#d8dfc7" stroke-width="14" stroke-linecap="round" opacity="0.72"/>
      <!-- abstract artifact geometry -->
      <g transform="translate(879 418)" opacity="0.48">
        <path d="M0 -62 L54 -30 L54 32 L0 64 L-54 32 L-54 -30 Z" fill="none" stroke="#f4be64" stroke-width="7"/>
        <circle r="22" fill="none" stroke="#ffe4a5" stroke-width="5"/><path d="M-78 0 H78 M0-84 V84" stroke="#d69a45" stroke-width="3"/>
      </g>
    </g>`
}

function meklar(c) {
  return `
    <g>
      <!-- industrial exoskeleton silhouette -->
      <path d="M215 1320 L252 892 L363 760 L470 706 L730 706 L840 760 L948 892 L985 1320 Z" fill="#0c1117" stroke="url(#rim)" stroke-width="10"/>
      <path d="M265 1120 L198 995 L229 782 L340 730 L408 828 L352 1008 Z" fill="url(#metal)" stroke="#7894a8" stroke-width="8"/>
      <path d="M935 1120 L1002 995 L971 782 L860 730 L792 828 L848 1008 Z" fill="url(#metal)" stroke="#7894a8" stroke-width="8"/>
      <!-- machinery shoulder pistons -->
      <path d="M246 940 L128 1100 M954 940 L1072 1100" stroke="#a76740" stroke-width="28" stroke-linecap="round"/>
      <circle cx="238" cy="944" r="42" fill="#17222c" stroke="#7ed7e0" stroke-width="7"/><circle cx="962" cy="944" r="42" fill="#17222c" stroke="#7ed7e0" stroke-width="7"/>
      <!-- life support cradle -->
      <ellipse cx="600" cy="610" rx="238" ry="286" fill="#101a22" stroke="#83bac7" stroke-width="10"/>
      <ellipse cx="600" cy="610" rx="184" ry="230" fill="#061017" stroke="#4fd5e6" stroke-width="5" opacity="0.92"/>
      <ellipse cx="600" cy="610" rx="169" ry="213" fill="#42b5c7" opacity="0.06"/>
      <!-- small atrophied biological core -->
      <path d="M534 510 C540 428 573 390 612 391 C653 392 681 433 680 515 L655 628 C646 675 626 706 596 707 C564 704 544 674 537 628 Z" fill="#8b6d5f" stroke="#d9a88f" stroke-width="7"/>
      <path d="M550 491 Q590 460 624 487 Q588 516 553 507 Z" fill="#10212a" stroke="#8ff4ff" stroke-width="5"/>
      <circle cx="589" cy="492" r="10" fill="#92f4ff"/>
      <path d="M661 491 Q632 462 607 488" fill="none" stroke="#e1b7a2" stroke-width="5" opacity="0.55"/>
      <!-- neural support crown -->
      <path d="M502 471 C458 397 482 318 540 284 L561 373 M698 471 C742 397 718 318 660 284 L639 373" fill="none" stroke="#80909c" stroke-width="17" stroke-linecap="round"/>
      <path d="M511 388 L443 310 M689 388 L757 310 M545 353 L531 242 M655 353 L669 242" stroke="#52d5e6" stroke-width="7" opacity="0.65"/>
      <!-- industrial service conduits -->
      <path d="M416 727 C370 609 389 494 466 420 M784 727 C830 609 811 494 734 420" fill="none" stroke="#c97545" stroke-width="18" opacity="0.78"/>
      <path d="M447 1280 V980 M753 1280 V980" stroke="#4f6170" stroke-width="38"/>
      <path d="M447 1280 V980 M753 1280 V980" stroke="#74dfe9" stroke-width="5" opacity="0.55"/>
      <!-- core status light -->
      <circle cx="600" cy="833" r="25" fill="#0c2028" stroke="#7cf5ff" stroke-width="7"/><circle cx="600" cy="833" r="55" fill="#54e0eb" opacity="0.13" filter="url(#glow)"/>
    </g>`
}

function silicoid(c) {
  return `
    <g>
      <!-- lattice background -->
      <g opacity="0.17" stroke="#ba8cff" stroke-width="4" fill="none">
        <path d="M298 1110 L418 930 L558 1040 L703 902 L874 1075"/>
        <path d="M355 1195 L485 1110 L629 1210 L768 1090 L930 1215"/>
      </g>
      <!-- non-humanoid crystalline body -->
      <path d="M244 1304 L334 1002 L472 930 L510 711 L636 620 L726 748 L889 844 L964 1304 Z" fill="#181925" stroke="url(#rim)" stroke-width="9"/>
      <path d="M334 1002 L470 930 L510 711 L416 768 Z" fill="#453651" stroke="#a873d5" stroke-width="6"/>
      <path d="M510 711 L636 620 L608 892 L470 930 Z" fill="#2d4051" stroke="#83d7df" stroke-width="6"/>
      <path d="M636 620 L726 748 L815 644 L766 915 L608 892 Z" fill="#4a2d58" stroke="#dc7dff" stroke-width="6"/>
      <path d="M726 748 L889 844 L766 915 Z" fill="#293b4e" stroke="#91e6e9" stroke-width="6"/>
      <path d="M470 930 L608 892 L558 1188 L398 1125 Z" fill="#344252" stroke="#77cad8" stroke-width="5"/>
      <path d="M608 892 L766 915 L827 1198 L558 1188 Z" fill="#40294e" stroke="#bf70e7" stroke-width="5"/>
      <!-- asymmetric crown/sensory crystals -->
      <path d="M476 704 L420 409 L550 553 Z" fill="#345c6b" stroke="#8be9ed" stroke-width="7"/>
      <path d="M585 626 L600 250 L680 566 Z" fill="#58356c" stroke="#e28cff" stroke-width="8"/>
      <path d="M690 663 L824 381 L759 721 Z" fill="#314c60" stroke="#85dde5" stroke-width="7"/>
      <path d="M755 772 L963 606 L859 865 Z" fill="#4e315e" stroke="#d77bfc" stroke-width="6"/>
      <!-- internal refractive energy, not face -->
      <ellipse cx="617" cy="782" rx="128" ry="88" fill="#6ee8ed" opacity="0.08" filter="url(#soft)"/>
      <path d="M538 791 L602 719 L683 771 L650 858 L560 850 Z" fill="#79e7e9" opacity="0.13" stroke="#d9ffff" stroke-width="5"/>
      <circle cx="605" cy="788" r="18" fill="#ecb0ff"/><circle cx="605" cy="788" r="67" fill="#d577ff" opacity="0.14" filter="url(#glow)"/>
      <!-- mineral feeding / grown command lattice -->
      <path d="M318 1262 L234 1125 L353 1081 M877 1268 L1010 1135 L863 1068" fill="none" stroke="#7997a6" stroke-width="18" stroke-linecap="round"/>
      <circle cx="236" cy="1126" r="28" fill="#29333c" stroke="#82e3e8" stroke-width="6"/>
      <circle cx="1005" cy="1135" r="28" fill="#35283f" stroke="#d988ff" stroke-width="6"/>
    </g>`
}

function render(c) {
  const subject = c.id === 'alkari' ? alkari(c) : c.id === 'meklar' ? meklar(c) : silicoid(c)
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}" role="img" aria-labelledby="title desc">
  <title id="title">MOX ${c.id} Gate 1 race portrait composition prototype</title>
  <desc id="desc">Original research-only MOOX portrait composition prototype. It tests a shared 4:5 cinematic portrait system and mobile-safe silhouette without copying original Master of Orion II portrait artwork.</desc>
  ${defs(c)}
  ${common(c)}
  ${subject}
</svg>`
}

for (const c of configs) fs.writeFileSync(path.join(here, `${c.id}.svg`), render(c).replace(/^[ \t]+$/gm, ''), 'utf8')
console.log(`Generated ${configs.length} research portrait composition prototypes`)
