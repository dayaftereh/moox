import type { ReactNode, SVGProps } from 'react'

export type GameIconName =
  | 'galaxy'
  | 'colonies'
  | 'fleets'
  | 'research'
  | 'diplomacy'
  | 'espionage'
  | 'ship-designer'
  | 'menu'
  | 'more'
  | 'home'
  | 'close'
  | 'credits'
  | 'food'
  | 'freighter'
  | 'command'
  | 'population'
  | 'production'
  | 'science'
  | 'planet'
  | 'star'
  | 'ship'
  | 'transport'
  | 'farmer'
  | 'worker'
  | 'scientist'
  | 'info'

export type GameIconProps = Omit<SVGProps<SVGSVGElement>, 'name'> & {
  name: GameIconName
  label?: string
}

const icons: Record<GameIconName, ReactNode> = {
  galaxy: <>
    <circle cx="12" cy="12" r="1.7" fill="currentColor" stroke="none" />
    <ellipse cx="12" cy="12" rx="8.1" ry="4.2" transform="rotate(-25 12 12)" />
    <ellipse cx="12" cy="12" rx="8.1" ry="4.2" transform="rotate(35 12 12)" />
    <circle cx="18.4" cy="8" r="1" fill="currentColor" stroke="none" />
    <circle cx="5.7" cy="15.8" r=".8" fill="currentColor" stroke="none" />
  </>,
  colonies: <>
    <path d="M4 17.5c1.8-3.9 4.4-5.9 8-5.9s6.2 2 8 5.9" />
    <path d="M7.4 16v-5.2l4.6-3.4 4.6 3.4V16" />
    <path d="M10.2 16v-3h3.6v3" />
    <path d="M4 19.5h16" />
  </>,
  fleets: <>
    <path d="M4 16.8 10.2 5l2.1 6.7L20 8.9l-5.7 10.2-2.1-5.5L4 16.8Z" />
    <path d="m6.1 18.3-2 1.4M9.1 19.4l-1 1.8" />
  </>,
  research: <>
    <circle cx="12" cy="12" r="1.5" fill="currentColor" stroke="none" />
    <ellipse cx="12" cy="12" rx="8.3" ry="3.3" />
    <ellipse cx="12" cy="12" rx="8.3" ry="3.3" transform="rotate(60 12 12)" />
    <ellipse cx="12" cy="12" rx="8.3" ry="3.3" transform="rotate(120 12 12)" />
  </>,
  diplomacy: <>
    <path d="m4.5 9.1 4.1-3.4 3 2.2 2.7-1.8 5.2 4.2" />
    <path d="m6.2 10.6 4.3 4.5c.8.8 2 .8 2.8 0l3.9-3.8" />
    <path d="m8.5 13 2.1-2.1c.8-.8 2.1-.8 2.9 0l1.9 1.9" />
    <path d="m4.7 8.8-1.4 5.4 3 2.7 2-2.1M19.1 10.1l1.6 4.5-3 2.7-2-2" />
  </>,
  espionage: <>
    <path d="M3.4 12s3.3-5.3 8.6-5.3 8.6 5.3 8.6 5.3-3.3 5.3-8.6 5.3S3.4 12 3.4 12Z" />
    <path d="m12 8.9 3.1 3.1-3.1 3.1L8.9 12 12 8.9Z" />
    <circle cx="12" cy="12" r="1" fill="currentColor" stroke="none" />
  </>,
  'ship-designer': <>
    <path d="M12 3.2 15 9l4.5 2.2-3.4 2.2.9 4.8-5-2.5-5 2.5.9-4.8-3.4-2.2L9 9l3-5.8Z" />
    <path d="M12 6v9.6M8.1 13.4h7.8" />
  </>,
  menu: <>
    <path d="M4 6.5h12M8 12h12M4 17.5h12" />
    <path d="M18.5 5v3M5.5 10.5v3M18.5 16v3" />
  </>,
  more: <>
    <path d="m12 3.5 7.3 4.2v8.6L12 20.5l-7.3-4.2V7.7L12 3.5Z" />
    <circle cx="8.2" cy="12" r="1" fill="currentColor" stroke="none" />
    <circle cx="12" cy="12" r="1" fill="currentColor" stroke="none" />
    <circle cx="15.8" cy="12" r="1" fill="currentColor" stroke="none" />
  </>,
  home: <>
    <path d="M4.2 11.2 12 4.8l7.8 6.4" />
    <path d="M6.3 10.2v8.5h11.4v-8.5M10 18.7v-5h4v5" />
  </>,
  close: <>
    <path d="m6 6 12 12M18 6 6 18" />
  </>,
  credits: <>
    <path d="m12 3.4 7.2 4.2v8.8L12 20.6l-7.2-4.2V7.6L12 3.4Z" />
    <path d="M15.2 8.4c-.8-.7-1.8-1.1-3-1.1-2.5 0-4.4 2-4.4 4.7s1.9 4.7 4.4 4.7c1.2 0 2.2-.4 3-1.1" />
    <path d="M6.6 10.4h6.2M6.6 13.6h6.2" />
  </>,
  food: <>
    <path d="M12 20V8.5" />
    <path d="M11.9 10.3C8.7 10.1 6.5 8.6 5.3 5c3.5-.3 5.7 1.3 6.6 5.3ZM12.1 13.9c3.3-.2 5.4-1.8 6.6-5.3-3.5-.3-5.7 1.3-6.6 5.3Z" />
    <path d="M8.3 18.3c1.3-1.2 2.5-1.8 3.7-1.8s2.4.6 3.7 1.8" />
  </>,
  freighter: <>
    <path d="M5.5 8.1h9.2l4.1 3.9-4.1 3.9H5.5V8.1Z" />
    <path d="M8.2 8.1v7.8M11.6 8.1v7.8M4 10H2.6M4 14H2.6" />
    <path d="m15.3 10.1 2 1.9-2 1.9" />
  </>,
  command: <>
    <circle cx="12" cy="12" r="7.6" />
    <circle cx="12" cy="12" r="3.8" />
    <path d="M12 4.4V8M12 16v3.6M4.4 12H8M16 12h3.6" />
    <path d="m10.3 13.2 1.7-5 1.7 5-1.7 2.2-1.7-2.2Z" fill="currentColor" stroke="none" />
  </>,
  population: <>
    <circle cx="12" cy="7.2" r="2.4" />
    <circle cx="6.5" cy="9.6" r="1.7" />
    <circle cx="17.5" cy="9.6" r="1.7" />
    <path d="M7.5 18.4c.4-3.4 1.9-5.1 4.5-5.1s4.1 1.7 4.5 5.1M3.8 17c.3-2.6 1.3-3.9 3.1-3.9.7 0 1.3.2 1.8.6M20.2 17c-.3-2.6-1.3-3.9-3.1-3.9-.7 0-1.3.2-1.8.6" />
  </>,
  production: <>
    <path d="M4.5 19V9.8l4.2 2.3V8.9l4.1 2.3V5h3.5v8.1l3.2 1.8V19h-15Z" />
    <path d="M8 15.2h2M13 15.2h2M17.1 15.2h1.1" />
  </>,
  science: <>
    <path d="M9 4h6M10 4v5.1l-4.4 7.5c-.8 1.4.2 3.2 1.8 3.2h9.2c1.6 0 2.6-1.8 1.8-3.2L14 9.1V4" />
    <path d="M8.2 14h7.6M9.5 11.8h5" />
    <circle cx="10" cy="16.7" r=".8" fill="currentColor" stroke="none" />
    <circle cx="14.5" cy="17" r=".7" fill="currentColor" stroke="none" />
  </>,
  planet: <>
    <circle cx="12" cy="12" r="5.5" />
    <path d="M3.4 14.1c3.1 1 7.3.8 11.5-.7 3.2-1.1 5.5-2.7 6.4-4" />
  </>,
  star: <>
    <path d="m12 3.3 1.8 5.5 5.8.1-4.6 3.4 1.7 5.6-4.7-3.3-4.7 3.3 1.7-5.6-4.6-3.4 5.8-.1L12 3.3Z" />
  </>,
  ship: <>
    <path d="M12 3.2 15 9l4.5 2.2-4.5 2.5.7 5.1-3.7-2.5-3.7 2.5.7-5.1-4.5-2.5L9 9l3-5.8Z" />
  </>,
  transport: <>
    <path d="M5 8h10.2l3.8 4-3.8 4H5V8Z" />
    <path d="M8 8v8M12 8v8M4 10H2.8M4 14H2.8" />
  </>,
  farmer: <>
    <circle cx="10.5" cy="8" r="2.5" />
    <path d="M5.5 19c.5-4.2 2.1-6.3 5-6.3 2.2 0 3.7 1.2 4.5 3.6" />
    <path d="M16.3 18.6v-5.4M16.3 14.2c-2.1-.2-3.5-1.2-4.1-3.2 2.3-.2 3.7.9 4.1 3.2ZM16.4 16.6c2.1-.2 3.4-1.3 4.1-3.2-2.3-.2-3.7.9-4.1 3.2Z" />
  </>,
  worker: <>
    <circle cx="10.5" cy="8.3" r="2.4" />
    <path d="M6 7.7c.4-2.5 2-4 4.5-4s4.1 1.5 4.5 4H6Z" />
    <path d="M5.2 19c.5-4.2 2.3-6.3 5.3-6.3 2 0 3.5.9 4.4 2.7" />
    <path d="m18 13.8 2 1.2v2.4l-2 1.2-2-1.2V15l2-1.2ZM18 11.8v2M18 18.6v2M14.8 13.6l1.5 1M19.7 18l1.5 1" />
  </>,
  scientist: <>
    <circle cx="10.2" cy="7.7" r="2.4" />
    <path d="M5.1 19c.5-4.2 2.2-6.3 5.1-6.3 1.6 0 2.9.6 3.8 1.7" />
    <path d="M7.8 7.4h4.8M8.1 6.4h1.6v2H8.1zM10.8 6.4h1.6v2h-1.6z" />
    <path d="M17 12.7v2.7l-2.4 3.9h5.7l-2.4-3.9v-2.7M15.9 16.7h3" />
  </>,
  info: <>
    <circle cx="12" cy="12" r="8" />
    <circle cx="12" cy="8" r=".9" fill="currentColor" stroke="none" />
    <path d="M12 11v5" />
  </>,
}

export function GameIcon({ name, label, className = '', ...props }: GameIconProps) {
  return (
    <svg
      {...props}
      className={`game-icon${className ? ` ${className}` : ''}`}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.55"
      strokeLinecap="round"
      strokeLinejoin="round"
      role={label ? 'img' : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
      focusable="false"
      data-icon={name}
    >
      {label && <title>{label}</title>}
      <g vectorEffect="non-scaling-stroke">{icons[name]}</g>
    </svg>
  )
}
