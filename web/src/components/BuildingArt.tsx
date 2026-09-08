import type { ReactNode, ReactElement } from 'react'

type BuildingArtProps = {
  buildingId: string
  className?: string
  variant?: 'compact' | 'hero' | 'surface'
}

function safeID(value: string) {
  return value.replace(/[^a-zA-Z0-9_-]/g, '-')
}

function SceneFrame({ buildingId, variant, children }: BuildingArtProps & { children: ReactNode }) {
  const id = safeID(buildingId)
  return (
    <svg className={`building-art building-art-${variant ?? 'hero'}`} viewBox="0 0 160 112" role="img" aria-label={buildingId.replace(/_/g, ' ')}>
      <defs>
        <linearGradient id={`${id}-space`} x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#071222" /><stop offset="1" stopColor="#02070e" /></linearGradient>
        <linearGradient id={`${id}-metal`} x1="0" y1="0" x2="1" y2="1"><stop offset="0" stopColor="#c5d7e8" /><stop offset=".38" stopColor="#627d99" /><stop offset=".7" stopColor="#263a50" /><stop offset="1" stopColor="#101b28" /></linearGradient>
        <linearGradient id={`${id}-dark-metal`} x1="0" y1="0" x2="1" y2="1"><stop offset="0" stopColor="#61778c" /><stop offset=".55" stopColor="#243547" /><stop offset="1" stopColor="#0b141f" /></linearGradient>
        <radialGradient id={`${id}-glow`} cx="50%" cy="50%" r="50%"><stop offset="0" stopColor="#8fd8ff" stopOpacity=".85" /><stop offset=".35" stopColor="#4d9dff" stopOpacity=".45" /><stop offset="1" stopColor="#4d9dff" stopOpacity="0" /></radialGradient>
        <linearGradient id={`${id}-city`} x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#b6d7ed" /><stop offset=".5" stopColor="#54738f" /><stop offset="1" stopColor="#18293b" /></linearGradient>
        <linearGradient id={`${id}-ground`} x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#27475b" /><stop offset="1" stopColor="#09131e" /></linearGradient>
        <filter id={`${id}-soft-glow`} x="-100%" y="-100%" width="300%" height="300%"><feGaussianBlur stdDeviation="3.2" /></filter>
      </defs>
      <rect width="160" height="112" rx="8" fill={`url(#${id}-space)`} />
      <g opacity=".62">
        <circle cx="17" cy="18" r=".7" fill="#b8dcff" /><circle cx="37" cy="11" r=".45" fill="#fff" /><circle cx="142" cy="18" r=".55" fill="#c7e4ff" /><circle cx="132" cy="41" r=".4" fill="#fff" /><circle cx="21" cy="77" r=".5" fill="#b9d4ff" /><circle cx="151" cy="83" r=".7" fill="#fff" />
      </g>
      {children}
    </svg>
  )
}

function StarBase({ buildingId, variant }: BuildingArtProps) {
  const id = safeID(buildingId)
  return <SceneFrame buildingId={buildingId} variant={variant}>
    <ellipse cx="80" cy="58" rx="48" ry="18" fill={`url(#${id}-glow)`} opacity=".32" filter={`url(#${id}-soft-glow)`} />
    <g transform="translate(80 56)">
      <ellipse rx="41" ry="22" fill="none" stroke="#79a7d0" strokeWidth="5.5" opacity=".78" />
      <ellipse rx="32" ry="16" fill="none" stroke="#213d58" strokeWidth="7" />
      <path d="M-55 0h27M28 0h27M0-35v20M0 15v35" stroke="#7898b4" strokeWidth="5" />
      <rect x="-9" y="-20" width="18" height="40" rx="5" fill={`url(#${id}-metal)`} stroke="#b8d8ef" strokeOpacity=".45" />
      <circle r="10" fill="#18324b" stroke="#91c9ef" strokeWidth="2.3" />
      <circle r="3.4" fill="#8ce3ff" />
      <g fill="#9cc9e8"><rect x="-58" y="-5" width="15" height="10" rx="2" /><rect x="43" y="-5" width="15" height="10" rx="2" /><rect x="-6" y="-45" width="12" height="14" rx="2" /><rect x="-6" y="34" width="12" height="15" rx="2" /></g>
      <g fill="#2a5e8d"><rect x="-64" y="-2" width="8" height="4" /><rect x="56" y="-2" width="8" height="4" /><rect x="-2" y="-50" width="4" height="7" /><rect x="-2" y="47" width="4" height="7" /></g>
      <g fill="#86d9ff"><circle cx="-35" cy="-13" r="1.5" /><circle cx="35" cy="13" r="1.5" /><circle cx="29" cy="-16" r="1.2" /><circle cx="-26" cy="16" r="1.2" /></g>
    </g>
  </SceneFrame>
}

function BattleStation({ buildingId, variant }: BuildingArtProps) {
  const id = safeID(buildingId)
  return <SceneFrame buildingId={buildingId} variant={variant}>
    <ellipse cx="80" cy="57" rx="54" ry="27" fill={`url(#${id}-glow)`} opacity=".26" filter={`url(#${id}-soft-glow)`} />
    <g transform="translate(80 57)">
      <path d="M0-38 31-25 47 0 31 25 0 38-31 25-47 0-31-25Z" fill={`url(#${id}-dark-metal)`} stroke="#829fb7" strokeWidth="2.4" />
      <path d="M0-28 23-19 34 0 23 19 0 28-23 19-34 0-23-19Z" fill="#12263a" stroke="#567995" strokeWidth="5" />
      <circle r="16" fill={`url(#${id}-metal)`} stroke="#b6d8ef" strokeWidth="2" />
      <circle r="7" fill="#173c5d" stroke="#80d6ff" strokeWidth="2.2" />
      <circle r="2.8" fill="#a4ecff" />
      <g fill="#728da5" stroke="#a9bed0" strokeOpacity=".38">
        <rect x="-62" y="-8" width="22" height="16" rx="3" /><rect x="40" y="-8" width="22" height="16" rx="3" /><rect x="-9" y="-53" width="18" height="20" rx="3" /><rect x="-9" y="33" width="18" height="20" rx="3" />
      </g>
      <g fill="#ffb66c"><circle cx="-53" cy="0" r="2" /><circle cx="53" cy="0" r="2" /><circle cx="0" cy="-43" r="2" /><circle cx="0" cy="43" r="2" /></g>
      <g stroke="#dcefff" strokeWidth="1.4"><path d="M-56-12v-8M-50-12v-11M56-12v-8M50-12v-11M-14-43h-9M14-43h9M-14 43h-9M14 43h9" /></g>
    </g>
  </SceneFrame>
}

function StarFortress({ buildingId, variant }: BuildingArtProps) {
  const id = safeID(buildingId)
  return <SceneFrame buildingId={buildingId} variant={variant}>
    <ellipse cx="80" cy="56" rx="62" ry="31" fill={`url(#${id}-glow)`} opacity=".34" filter={`url(#${id}-soft-glow)`} />
    <g transform="translate(80 56)">
      <circle r="42" fill="#0c1825" stroke="#597b98" strokeWidth="4" />
      <circle r="32" fill="none" stroke="#2d5879" strokeWidth="8" />
      <circle r="20" fill={`url(#${id}-metal)`} stroke="#a7c8de" strokeWidth="2.4" />
      <path d="M0-53 8-31-8-31ZM53 0 31 8 31-8ZM0 53 8 31-8 31ZM-53 0-31 8-31-8Z" fill="#91b5d1" stroke="#d7ebf7" strokeOpacity=".35" />
      <path d="M-38-38-22-28-28-22ZM38-38 22-28 28-22ZM38 38 22 28 28 22ZM-38 38-22 28-28 22Z" fill="#4b6680" />
      <circle r="10" fill="#0d3656" stroke="#86ddff" strokeWidth="2.6" /><circle r="4" fill="#b0f0ff" />
      <g fill="#e5a75d"><circle cx="0" cy="-41" r="2.2" /><circle cx="41" cy="0" r="2.2" /><circle cx="0" cy="41" r="2.2" /><circle cx="-41" cy="0" r="2.2" /></g>
      <g stroke="#9cc9e8" strokeWidth="2"><path d="M-58-18-42-13M58-18 42-13M-58 18-42 13M58 18 42 13" /></g>
    </g>
  </SceneFrame>
}

function Capitol({ buildingId, variant }: BuildingArtProps) {
  const id = safeID(buildingId)
  return <SceneFrame buildingId={buildingId} variant={variant}>
    <ellipse cx="80" cy="91" rx="64" ry="13" fill="#18344a" opacity=".62" />
    <path d="M31 88 43 62h23l9-35h11l9 35h22l12 26Z" fill={`url(#${id}-city)`} stroke="#739ab8" strokeWidth="1.5" />
    <path d="M70 62h20l-4-35H75Z" fill="#294b66" /><path d="M80 17v12" stroke="#a8dfff" strokeWidth="2" />
    <ellipse cx="54" cy="67" rx="13" ry="8" fill="#27475e" stroke="#79a9c8" /><ellipse cx="106" cy="67" rx="13" ry="8" fill="#27475e" stroke="#79a9c8" />
    <path d="M37 88h86M46 79h68" stroke="#9dd7ef" strokeOpacity=".42" />
    <g fill="#89dfff"><rect x="78" y="38" width="4" height="7" rx="1" /><rect x="51" y="65" width="4" height="3" rx="1" /><rect x="105" y="65" width="4" height="3" rx="1" /></g>
  </SceneFrame>
}

function ColonyBase({ buildingId, variant }: BuildingArtProps) {
  const id = safeID(buildingId)
  return <SceneFrame buildingId={buildingId} variant={variant}>
    <ellipse cx="80" cy="90" rx="59" ry="12" fill={`url(#${id}-ground)`} opacity=".8" />
    <g fill={`url(#${id}-metal)`} stroke="#83abc5" strokeWidth="1.4"><path d="M35 83c2-20 10-30 24-30s22 10 24 30Z" /><path d="M78 83c2-27 13-41 31-41 17 0 26 14 28 41Z" /></g>
    <path d="M52 53c1-8 4-13 8-16M111 42V28M105 31h12" stroke="#a7d9ee" strokeWidth="2" />
    <rect x="55" y="72" width="14" height="11" rx="3" fill="#17334a" /><rect x="104" y="69" width="16" height="14" rx="3" fill="#17334a" />
    <g fill="#76d9ff"><circle cx="53" cy="63" r="1.4" /><circle cx="115" cy="57" r="1.4" /><circle cx="126" cy="69" r="1.2" /></g>
    <path d="M22 87h116" stroke="#50748d" strokeWidth="2" />
  </SceneFrame>
}

function Barracks({ buildingId, variant }: BuildingArtProps) {
  const id = safeID(buildingId)
  return <SceneFrame buildingId={buildingId} variant={variant}>
    <ellipse cx="80" cy="91" rx="61" ry="12" fill={`url(#${id}-ground)`} opacity=".7" />
    <path d="M31 83 38 48l42-14 42 14 7 35Z" fill={`url(#${id}-dark-metal)`} stroke="#8198ad" strokeWidth="1.8" />
    <path d="M48 56h64v27H48Z" fill="#203449" stroke="#6e879c" /><path d="M67 83V61h26v22" fill="#111f2d" stroke="#92a9bb" />
    <path d="m45 48 10-13M115 48l-10-13M53 38l-8-5M107 38l8-5" stroke="#9fb7c8" strokeWidth="2" />
    <g fill="#e37e6f"><rect x="55" y="61" width="5" height="3" /><rect x="100" y="61" width="5" height="3" /><circle cx="80" cy="46" r="2" /></g>
    <path d="M20 88h120" stroke="#526b7e" strokeWidth="2" />
  </SceneFrame>
}

function GenericBuilding({ buildingId, variant }: BuildingArtProps) {
  const id = safeID(buildingId)
  return <SceneFrame buildingId={buildingId} variant={variant}>
    <ellipse cx="80" cy="92" rx="58" ry="12" fill={`url(#${id}-ground)`} opacity=".7" />
    <path d="M40 84V52l22 12V45l22 12V31h19v36l20 11v6Z" fill={`url(#${id}-city)`} stroke="#7596af" strokeWidth="1.6" />
    <g fill="#88d8ff"><rect x="51" y="68" width="6" height="4" /><rect x="72" y="65" width="6" height="4" /><rect x="91" y="48" width="5" height="5" /><rect x="107" y="72" width="6" height="4" /></g>
  </SceneFrame>
}

export function BuildingArt({ buildingId, className, variant = 'hero' }: BuildingArtProps) {
  let art: ReactElement
  switch (buildingId) {
    case 'star_base': art = <StarBase buildingId={buildingId} variant={variant} />; break
    case 'battlestation': art = <BattleStation buildingId={buildingId} variant={variant} />; break
    case 'star_fortress': art = <StarFortress buildingId={buildingId} variant={variant} />; break
    case 'capitol': art = <Capitol buildingId={buildingId} variant={variant} />; break
    case 'colony_base': art = <ColonyBase buildingId={buildingId} variant={variant} />; break
    case 'marine_barracks': art = <Barracks buildingId={buildingId} variant={variant} />; break
    default: art = <GenericBuilding buildingId={buildingId} variant={variant} />
  }
  if (!className) return art
  return <span className={className}>{art}</span>
}
