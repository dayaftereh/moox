export type SpecialShipKind = 'colony_ship' | 'outpost_ship' | 'troop_transport'

type SpecialShipGlyphProps = {
  kind: SpecialShipKind
  className?: string
  label?: string
}

export function SpecialShipGlyph({ kind, className = '', label }: SpecialShipGlyphProps) {
  return (
    <svg
      className={className}
      viewBox="0 0 120 72"
      role={label ? 'img' : 'presentation'}
      aria-label={label}
      aria-hidden={label ? undefined : true}
    >
      {kind === 'colony_ship' && (
        <g fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round">
          <path d="M8 36 25 25 58 18 92 25 112 36 92 47 58 54 25 47Z" fill="currentColor" fillOpacity=".08" strokeWidth="2.2" />
          <path d="M18 36 43 31 82 31 103 36 82 41 43 41Z" fill="currentColor" fillOpacity=".1" strokeWidth="1.8" />
          <path d="M45 28 56 13h18l11 15M45 44l11 15h18l11-15" strokeWidth="2" />
          <ellipse cx="65" cy="18" rx="10" ry="7" fill="currentColor" fillOpacity=".13" strokeWidth="1.8" />
          <ellipse cx="65" cy="54" rx="10" ry="7" fill="currentColor" fillOpacity=".13" strokeWidth="1.8" />
          <path d="M27 31V21M27 41v10M94 31v-8M94 41v8" strokeWidth="1.7" />
          <circle cx="27" cy="18" r="3" fill="currentColor" fillOpacity=".22" strokeWidth="1.4" />
          <circle cx="27" cy="54" r="3" fill="currentColor" fillOpacity=".22" strokeWidth="1.4" />
        </g>
      )}
      {kind === 'outpost_ship' && (
        <g fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round">
          <path d="M10 36 31 25h48l31 11-31 11H31Z" fill="currentColor" fillOpacity=".08" strokeWidth="2.2" />
          <rect x="40" y="20" width="37" height="32" rx="5" fill="currentColor" fillOpacity=".08" strokeWidth="1.8" />
          <path d="M48 20V9h21v11M48 52v11h21V52M77 27l20-11M77 45l20 11" strokeWidth="1.8" />
          <circle cx="101" cy="14" r="4" fill="currentColor" fillOpacity=".18" strokeWidth="1.5" />
          <circle cx="101" cy="58" r="4" fill="currentColor" fillOpacity=".18" strokeWidth="1.5" />
          <path d="M18 32v8M25 29v14" strokeWidth="1.7" />
        </g>
      )}
      {kind === 'troop_transport' && (
        <g fill="none" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round">
          <path d="M9 36 31 23h54l27 13-27 13H31Z" fill="currentColor" fillOpacity=".08" strokeWidth="2.2" />
          <path d="M32 28 49 11h25l12 17M32 44l17 17h25l12-17" strokeWidth="1.9" />
          <rect x="49" y="27" width="33" height="18" rx="4" fill="currentColor" fillOpacity=".11" strokeWidth="1.7" />
          <path d="M56 31h19M56 36h19M56 41h19M23 31v10" strokeWidth="1.5" />
        </g>
      )}
    </svg>
  )
}
