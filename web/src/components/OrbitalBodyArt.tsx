import type { OrbitalBodyKind } from '../api'

type Palette = {
  base: string
  light: string
  dark: string
  feature: string
  feature2: string
  atmosphere: string
  cloud?: string
}

const climatePalettes: Record<string, Palette> = {
  terran: { base: '#2469a6', light: '#58a9d8', dark: '#081d38', feature: '#4f8d48', feature2: '#9b7544', atmosphere: '#77c7ff', cloud: '#f1f7ff' },
  ocean: { base: '#146ba9', light: '#47b7db', dark: '#061d3b', feature: '#2388ad', feature2: '#73cfca', atmosphere: '#66d6ff', cloud: '#edfaff' },
  desert: { base: '#a7642b', light: '#e0b15f', dark: '#3b1b0e', feature: '#7b4623', feature2: '#d98b38', atmosphere: '#e4b579' },
  arid: { base: '#9a5c31', light: '#d59d55', dark: '#32170d', feature: '#6d3f29', feature2: '#c77b3e', atmosphere: '#d6a56f' },
  barren: { base: '#65594e', light: '#9b8973', dark: '#211c19', feature: '#443d37', feature2: '#827260', atmosphere: '#8e8276' },
  tundra: { base: '#78929e', light: '#d6edf3', dark: '#26343d', feature: '#526f7b', feature2: '#aebfc5', atmosphere: '#c4e9f4', cloud: '#f4fbff' },
  arctic: { base: '#8ba8b5', light: '#ecf8fc', dark: '#263944', feature: '#5f8494', feature2: '#c8dce3', atmosphere: '#d9f5ff', cloud: '#ffffff' },
  toxic: { base: '#78862e', light: '#c8d55c', dark: '#252b0a', feature: '#4f5d1b', feature2: '#a9b842', atmosphere: '#cbe36e' },
  radiated: { base: '#6d3e73', light: '#bf70b1', dark: '#211126', feature: '#42214d', feature2: '#985aa0', atmosphere: '#d179d0' },
  swamp: { base: '#456f56', light: '#7ca56a', dark: '#18271d', feature: '#314c34', feature2: '#6e7737', atmosphere: '#7fa787', cloud: '#dbe6d6' },
}

const fallbackPalette: Palette = { base: '#52677d', light: '#91afc7', dark: '#17212d', feature: '#394b5d', feature2: '#74899b', atmosphere: '#85a9c8' }

function hash(seed: number, salt: number) {
  let value = Math.imul((seed ^ salt) >>> 0, 2654435761) >>> 0
  value ^= value >>> 16
  value = Math.imul(value, 2246822519) >>> 0
  value ^= value >>> 13
  return (value >>> 0) / 4294967295
}

function climatePalette(climate?: string) {
  return climatePalettes[(climate ?? '').toLowerCase()] ?? fallbackPalette
}

function planetFeature(seed: number, index: number) {
  const x = 12 + hash(seed, 71 + index * 11) * 40
  const y = 12 + hash(seed, 193 + index * 17) * 40
  const rx = 3 + hash(seed, 317 + index * 23) * 9
  const ry = 1.8 + hash(seed, 431 + index * 29) * 5.5
  const angle = -55 + hash(seed, 557 + index * 31) * 110
  return { x, y, rx, ry, angle }
}

function PlanetArt({ id, climateId }: { id: number; climateId?: string }) {
  const p = climatePalette(climateId)
  const prefix = `planet-${id}`
  const clouded = Boolean(p.cloud)
  const features = Array.from({ length: climateId === 'ocean' ? 5 : 9 }, (_, index) => planetFeature(id, index))
  return (
    <svg className="orbital-body-art orbital-body-art-planet" viewBox="0 0 64 64" aria-hidden="true">
      <defs>
        <radialGradient id={`${prefix}-sphere`} cx="32%" cy="24%" r="76%">
          <stop offset="0" stopColor={p.light} />
          <stop offset=".47" stopColor={p.base} />
          <stop offset=".82" stopColor={p.base} />
          <stop offset="1" stopColor={p.dark} />
        </radialGradient>
        <radialGradient id={`${prefix}-shade`} cx="30%" cy="26%" r="72%">
          <stop offset="0" stopColor="#ffffff" stopOpacity=".18" />
          <stop offset=".58" stopColor="#000000" stopOpacity="0" />
          <stop offset="1" stopColor="#000000" stopOpacity=".68" />
        </radialGradient>
        <linearGradient id={`${prefix}-rim`} x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stopColor={p.atmosphere} stopOpacity=".86" />
          <stop offset=".5" stopColor={p.atmosphere} stopOpacity=".18" />
          <stop offset="1" stopColor="#000" stopOpacity="0" />
        </linearGradient>
        <clipPath id={`${prefix}-clip`}><circle cx="32" cy="32" r="26" /></clipPath>
        <filter id={`${prefix}-noise`} x="-10%" y="-10%" width="120%" height="120%">
          <feTurbulence type="fractalNoise" baseFrequency=".055 .085" numOctaves="4" seed={(id % 97) + 3} result="noise" />
          <feColorMatrix in="noise" type="saturate" values="0" result="mono" />
          <feComponentTransfer in="mono"><feFuncA type="table" tableValues="0 .52" /></feComponentTransfer>
          <feComposite in2="SourceGraphic" operator="in" />
        </filter>
        {clouded && (
          <filter id={`${prefix}-cloud-noise`} x="-15%" y="-15%" width="130%" height="130%">
            <feTurbulence type="fractalNoise" baseFrequency=".025 .07" numOctaves="3" seed={(id % 83) + 11} result="cloudNoise" />
            <feComponentTransfer in="cloudNoise"><feFuncA type="gamma" amplitude="1.8" exponent="2.2" offset="-.45" /></feComponentTransfer>
            <feGaussianBlur stdDeviation=".45" />
          </filter>
        )}
      </defs>
      <circle cx="32" cy="32" r="27.6" fill={p.atmosphere} opacity=".2" />
      <circle cx="32" cy="32" r="26" fill={`url(#${prefix}-sphere)`} />
      <g clipPath={`url(#${prefix}-clip)`}>
        <rect x="5" y="5" width="54" height="54" fill={p.feature} opacity=".48" filter={`url(#${prefix}-noise)`} />
        {features.map((feature, index) => (
          <ellipse key={index} cx={feature.x} cy={feature.y} rx={feature.rx} ry={feature.ry} transform={`rotate(${feature.angle} ${feature.x} ${feature.y})`} fill={index % 3 === 0 ? p.feature2 : p.feature} opacity={climateId === 'ocean' ? .34 : .3 + (index % 2) * .12} />
        ))}
        {climateId === 'terran' && <path d="M8 37c8-4 11-2 17 1 5 3 10 1 15-3 6-5 10-4 16-1v13H8Z" fill="#286d3d" opacity=".42" />}
        {(climateId === 'tundra' || climateId === 'arctic') && <path d="M9 15c9 3 15 3 23 0 7-3 13-2 22 2v8c-9-4-17-4-25-1-7 3-13 2-20-1Z" fill="#f0fbff" opacity=".44" />}
        {clouded && <rect x="4" y="4" width="56" height="56" fill={p.cloud} opacity={climateId === 'ocean' ? .32 : .25} filter={`url(#${prefix}-cloud-noise)`} />}
        <circle cx="32" cy="32" r="26" fill={`url(#${prefix}-shade)`} />
      </g>
      <circle cx="32" cy="32" r="25.4" fill="none" stroke={`url(#${prefix}-rim)`} strokeWidth="1.3" />
      <ellipse cx="23" cy="18" rx="9" ry="5" fill="#ffffff" opacity=".08" transform="rotate(-24 23 18)" />
    </svg>
  )
}

function GasGiantArt({ id }: { id: number }) {
  const prefix = `gas-${id}`
  const palettes = [
    ['#e7c28d', '#b66f47', '#613b36'],
    ['#d8b48a', '#9a7657', '#4c3c37'],
    ['#d2c1a5', '#8b8f8a', '#45515b'],
  ]
  const colors = palettes[id % palettes.length]
  return (
    <svg className="orbital-body-art orbital-body-art-gas" viewBox="0 0 64 64" aria-hidden="true">
      <defs>
        <radialGradient id={`${prefix}-sphere`} cx="30%" cy="24%" r="76%"><stop offset="0" stopColor={colors[0]} /><stop offset=".62" stopColor={colors[1]} /><stop offset="1" stopColor={colors[2]} /></radialGradient>
        <clipPath id={`${prefix}-clip`}><circle cx="32" cy="32" r="27" /></clipPath>
        <filter id={`${prefix}-warp`}><feTurbulence type="fractalNoise" baseFrequency=".018 .09" numOctaves="2" seed={(id % 67) + 5} result="n" /><feDisplacementMap in="SourceGraphic" in2="n" scale="4" xChannelSelector="R" yChannelSelector="B" /></filter>
        <radialGradient id={`${prefix}-shade`} cx="30%" cy="25%" r="75%"><stop offset=".05" stopColor="#fff" stopOpacity=".17" /><stop offset=".6" stopColor="#000" stopOpacity="0" /><stop offset="1" stopColor="#000" stopOpacity=".62" /></radialGradient>
      </defs>
      <circle cx="32" cy="32" r="28.4" fill={colors[0]} opacity=".16" />
      <circle cx="32" cy="32" r="27" fill={`url(#${prefix}-sphere)`} />
      <g clipPath={`url(#${prefix}-clip)`} filter={`url(#${prefix}-warp)`}>
        {[14, 20, 26, 33, 40, 47].map((y, index) => <rect key={y} x="4" y={y} width="56" height={index % 2 ? 4.2 : 3} rx="2" fill={index % 2 ? colors[0] : colors[2]} opacity={index % 2 ? .22 : .2} />)}
        <ellipse cx={40 + hash(id, 92) * 5} cy={34 + hash(id, 121) * 5} rx="6.5" ry="2.7" fill="#7b473b" opacity=".52" />
        <circle cx="32" cy="32" r="27" fill={`url(#${prefix}-shade)`} />
      </g>
      <circle cx="32" cy="32" r="26.5" fill="none" stroke={colors[0]} strokeOpacity=".45" strokeWidth="1" />
    </svg>
  )
}

function AsteroidBeltArt({ id }: { id: number }) {
  const rocks = Array.from({ length: 13 }, (_, index) => {
    const angle = (index / 13) * Math.PI * 2 + hash(id, index + 3) * .18
    const radius = 18 + hash(id, index + 40) * 6
    return { x: 32 + Math.cos(angle) * radius, y: 32 + Math.sin(angle) * radius * .48, r: 1.3 + hash(id, index + 70) * 2.1 }
  })
  return (
    <svg className="orbital-body-art orbital-body-art-asteroids" viewBox="0 0 64 64" aria-hidden="true">
      <ellipse cx="32" cy="32" rx="27" ry="13" fill="none" stroke="#7f8790" strokeOpacity=".28" />
      {rocks.map((rock, index) => <circle key={index} cx={rock.x} cy={rock.y} r={rock.r} fill={index % 3 === 0 ? '#a99d8e' : '#6f7072'} stroke="#d0c4b5" strokeOpacity=".22" strokeWidth=".6" />)}
      <ellipse cx="25" cy="22" rx="9" ry="3" fill="#ffffff" opacity=".04" />
    </svg>
  )
}

export function OrbitalBodyArt({ kind, id, climateId }: { kind: OrbitalBodyKind; id: number; climateId?: string }) {
  if (kind === 'gas_giant') return <GasGiantArt id={id} />
  if (kind === 'asteroid_belt') return <AsteroidBeltArt id={id} />
  return <PlanetArt id={id} climateId={climateId} />
}
