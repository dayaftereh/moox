type StarPalette = {
  key: string
  label: string
  spectral: string
  core: string
  hot: string
  mid: string
  glow: string
  ray: string
}

const starPalettes: Record<number, StarPalette> = {
  // MOO2-compatible normalized spectral order used by the current galaxy ruleset.
  0: { key: 'blue-white', label: 'Blue-White', spectral: 'B', core: '#ffffff', hot: '#dff3ff', mid: '#88c8ff', glow: '#3d83ff', ray: '#b8ddff' },
  1: { key: 'white', label: 'White', spectral: 'F', core: '#ffffff', hot: '#fffef4', mid: '#d9e8ff', glow: '#92b7e8', ray: '#f0f7ff' },
  2: { key: 'yellow', label: 'Yellow', spectral: 'G', core: '#fffef0', hot: '#fff6a9', mid: '#ffd94f', glow: '#eaa81c', ray: '#fff1a2' },
  3: { key: 'orange', label: 'Orange', spectral: 'K', core: '#fff4df', hot: '#ffc576', mid: '#f18b32', glow: '#c95518', ray: '#ffd0a0' },
  4: { key: 'red', label: 'Red', spectral: 'M', core: '#fff0e9', hot: '#ff9d82', mid: '#ef4b45', glow: '#a5162c', ray: '#ff9b92' },
  5: { key: 'brown', label: 'Brown', spectral: 'BD', core: '#ffe8bd', hot: '#d99a54', mid: '#9a5a31', glow: '#63321f', ray: '#e5a56b' },
  6: { key: 'black-hole', label: 'Black Hole', spectral: 'BH', core: '#010205', hot: '#5f78a9', mid: '#4e4b96', glow: '#2d285c', ray: '#7e91d4' },
}

const fallbackPalette = starPalettes[2]

function hash(seed: number, salt: number) {
  let value = Math.imul((seed ^ salt) >>> 0, 2654435761) >>> 0
  value ^= value >>> 16
  value = Math.imul(value, 2246822519) >>> 0
  value ^= value >>> 13
  return (value >>> 0) / 4294967295
}

export function starPalette(spectralClass: number): StarPalette {
  return starPalettes[spectralClass] ?? fallbackPalette
}

export function StarArt({ spectralClass, seed, className = '' }: { spectralClass: number; seed: number; className?: string }) {
  const p = starPalette(spectralClass)
  const prefix = `star-${seed}-${spectralClass}`
  const rayRotation = Math.round(hash(seed, 11) * 35 - 17)
  const flare = .82 + hash(seed, 29) * .35
  const haloScale = .92 + hash(seed, 53) * .18

  if (spectralClass === 6) {
    return (
      <svg className={`star-art star-art-black-hole ${className}`.trim()} viewBox="0 0 96 96" aria-hidden="true" data-spectral-class={spectralClass} data-star-kind={p.key}>
        <defs>
          <radialGradient id={`${prefix}-halo`} cx="50%" cy="50%" r="50%">
            <stop offset="0" stopColor="#8ca2ff" stopOpacity=".12" />
            <stop offset=".32" stopColor={p.mid} stopOpacity=".3" />
            <stop offset=".58" stopColor={p.glow} stopOpacity=".2" />
            <stop offset="1" stopColor={p.glow} stopOpacity="0" />
          </radialGradient>
          <linearGradient id={`${prefix}-disk`} x1="0" y1="0" x2="1" y2="0">
            <stop offset="0" stopColor="#45579c" stopOpacity="0" />
            <stop offset=".2" stopColor="#7791e2" stopOpacity=".72" />
            <stop offset=".48" stopColor="#d4dcff" stopOpacity=".95" />
            <stop offset=".52" stopColor="#f3e4ff" stopOpacity=".96" />
            <stop offset=".8" stopColor="#6b5bb3" stopOpacity=".64" />
            <stop offset="1" stopColor="#433468" stopOpacity="0" />
          </linearGradient>
          <filter id={`${prefix}-blur`} x="-80%" y="-80%" width="260%" height="260%"><feGaussianBlur stdDeviation="4" /></filter>
          <filter id={`${prefix}-disk-blur`} x="-20%" y="-70%" width="140%" height="240%"><feGaussianBlur stdDeviation="1.3" /></filter>
        </defs>
        <circle cx="48" cy="48" r="43" fill={`url(#${prefix}-halo)`} transform={`scale(${haloScale}) translate(${48 / haloScale - 48} ${48 / haloScale - 48})`} />
        <ellipse cx="48" cy="48" rx="37" ry="9.2" fill={`url(#${prefix}-disk)`} filter={`url(#${prefix}-disk-blur)`} transform={`rotate(${rayRotation * .35} 48 48)`} />
        <ellipse cx="48" cy="48" rx="30" ry="5.3" fill="#c5d0ff" opacity=".18" filter={`url(#${prefix}-blur)`} />
        <circle cx="48" cy="48" r="13.5" fill="#000105" />
        <circle cx="48" cy="48" r="14.2" fill="none" stroke="#7275bd" strokeOpacity=".36" strokeWidth="1.4" />
        <ellipse cx="48" cy="44" rx="9" ry="2.2" fill="#ffffff" opacity=".04" />
      </svg>
    )
  }

  return (
    <svg className={`star-art ${className}`.trim()} viewBox="0 0 96 96" aria-hidden="true" data-spectral-class={spectralClass} data-star-kind={p.key}>
      <defs>
        <radialGradient id={`${prefix}-halo`} cx="50%" cy="50%" r="50%">
          <stop offset="0" stopColor={p.hot} stopOpacity=".74" />
          <stop offset=".18" stopColor={p.mid} stopOpacity=".58" />
          <stop offset=".48" stopColor={p.glow} stopOpacity=".26" />
          <stop offset="1" stopColor={p.glow} stopOpacity="0" />
        </radialGradient>
        <radialGradient id={`${prefix}-body`} cx="38%" cy="34%" r="68%">
          <stop offset="0" stopColor={p.core} />
          <stop offset=".18" stopColor={p.hot} />
          <stop offset=".52" stopColor={p.mid} />
          <stop offset=".83" stopColor={p.mid} />
          <stop offset="1" stopColor={p.glow} />
        </radialGradient>
        <linearGradient id={`${prefix}-ray`} x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stopColor={p.ray} stopOpacity="0" />
          <stop offset=".43" stopColor={p.ray} stopOpacity=".08" />
          <stop offset=".5" stopColor={p.core} stopOpacity=".82" />
          <stop offset=".57" stopColor={p.ray} stopOpacity=".08" />
          <stop offset="1" stopColor={p.ray} stopOpacity="0" />
        </linearGradient>
        <filter id={`${prefix}-blur`} x="-100%" y="-100%" width="300%" height="300%"><feGaussianBlur stdDeviation="5.5" /></filter>
        <filter id={`${prefix}-soft`} x="-80%" y="-80%" width="260%" height="260%"><feGaussianBlur stdDeviation="1.8" /></filter>
        <filter id={`${prefix}-surface`} x="-20%" y="-20%" width="140%" height="140%">
          <feTurbulence type="fractalNoise" baseFrequency=".11" numOctaves="2" seed={(seed % 89) + 7} result="noise" />
          <feColorMatrix in="noise" type="matrix" values="1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 .16 0" />
        </filter>
        <clipPath id={`${prefix}-clip`}><circle cx="48" cy="48" r="12.4" /></clipPath>
      </defs>
      <circle cx="48" cy="48" r="44" fill={`url(#${prefix}-halo)`} opacity={flare} />
      <circle cx="48" cy="48" r="30" fill={p.glow} opacity=".2" filter={`url(#${prefix}-blur)`} />
      <g transform={`rotate(${rayRotation} 48 48)`} opacity=".9">
        <rect x="6" y="47.25" width="84" height="1.5" rx=".75" fill={`url(#${prefix}-ray)`} filter={`url(#${prefix}-soft)`} />
        <rect x="10" y="47.45" width="76" height="1.1" rx=".55" fill={`url(#${prefix}-ray)`} transform="rotate(90 48 48)" filter={`url(#${prefix}-soft)`} />
        <rect x="19" y="47.6" width="58" height=".8" rx=".4" fill={`url(#${prefix}-ray)`} transform="rotate(45 48 48)" opacity=".52" />
        <rect x="19" y="47.6" width="58" height=".8" rx=".4" fill={`url(#${prefix}-ray)`} transform="rotate(-45 48 48)" opacity=".52" />
      </g>
      <circle cx="48" cy="48" r="13.2" fill={p.mid} opacity=".38" filter={`url(#${prefix}-soft)`} />
      <circle cx="48" cy="48" r="12.4" fill={`url(#${prefix}-body)`} />
      <g clipPath={`url(#${prefix}-clip)`} opacity=".7"><rect x="34" y="34" width="28" height="28" fill={p.hot} filter={`url(#${prefix}-surface)`} /></g>
      <circle cx="44" cy="43" r="4.1" fill={p.core} opacity=".58" filter={`url(#${prefix}-soft)`} />
      <circle cx="48" cy="48" r="12.2" fill="none" stroke={p.hot} strokeOpacity=".42" strokeWidth="1" />
    </svg>
  )
}
