import type { MilitaryDesignerWeaponFiringArcChoice, ShipWeaponFiringArcID } from '../api'
import type { ShipVisualGenome } from '../shipVisualGenome'
import { ProceduralShipGlyph } from './ProceduralShipGlyph'

type Props = {
  choices?: MilitaryDesignerWeaponFiringArcChoice[]
  selectedArcID?: ShipWeaponFiringArcID
  defaultArcID?: ShipWeaponFiringArcID
  fixedArcID?: ShipWeaponFiringArcID
  hullID: string
  shipGenome: ShipVisualGenome
  weaponCount: number
  locale: string
  disabled?: boolean
  fixedLabel?: string
  onChange: (arcID: ShipWeaponFiringArcID) => void
}

const arcShortLabels: Record<ShipWeaponFiringArcID, string> = {
  forward: 'F',
  forward_extended: 'Fx',
  back_extended: 'Bx',
  rear: 'B',
  '360': '360',
}

const arcPositionClasses: Partial<Record<ShipWeaponFiringArcID, string>> = {
  forward: 'north',
  forward_extended: 'east',
  rear: 'south',
  back_extended: 'west',
}

function choiceLabel(choice: MilitaryDesignerWeaponFiringArcChoice | undefined, locale: string, fallback: string): string {
  if (!choice) return fallback
  return locale === 'de' ? choice.label_de : choice.label_en
}

function costLabel(choice: MilitaryDesignerWeaponFiringArcChoice | undefined): string {
  if (!choice) return ''
  const percent = Math.max(choice.space_cost_percent, choice.production_cost_percent)
  return percent > 0 ? `+${percent}%` : '±0%'
}

export function FiringArcCompass({
  choices = [],
  selectedArcID,
  defaultArcID,
  fixedArcID,
  hullID,
  shipGenome,
  weaponCount,
  locale,
  disabled = false,
  fixedLabel,
  onChange,
}: Props) {
  const resolvedArc = fixedArcID ?? selectedArcID ?? defaultArcID ?? choices[0]?.id ?? '360'
  const choiceByID = new Map(choices.map((choice) => [choice.id, choice]))
  const fixed = Boolean(fixedArcID)
  const selectable = !fixed && !disabled
  const centerChoice = choiceByID.get('360')

  return (
    <div
      className={`shipdesigner-firing-compass arc-${resolvedArc}${fixed ? ' fixed' : ''}`}
      data-firing-arc-compass
      data-active-firing-arc={resolvedArc}
    >
      <div className="shipdesigner-firing-compass-face" aria-hidden="true">
        <div className="shipdesigner-firing-compass-grid" />
        <div className="shipdesigner-firing-compass-coverage" />
        <div className="shipdesigner-firing-compass-radar-ring ring-one" />
        <div className="shipdesigner-firing-compass-radar-ring ring-two" />
        <span className="shipdesigner-firing-compass-heading heading-n">▲</span>
        <span className="shipdesigner-firing-compass-heading heading-e">›</span>
        <span className="shipdesigner-firing-compass-heading heading-s">▼</span>
        <span className="shipdesigner-firing-compass-heading heading-w">‹</span>
      </div>

      {fixed && ([
        ['forward', 'north'],
        ['forward_extended', 'east'],
        ['rear', 'south'],
        ['back_extended', 'west'],
      ] as const).map(([arcID, position]) => (
        <span
          key={arcID}
          className={`shipdesigner-firing-compass-button decorative ${position}`}
          aria-hidden="true"
        >
          <strong>{arcShortLabels[arcID]}</strong>
        </span>
      ))}

      {choices.filter((choice) => choice.id !== '360').map((choice) => {
        const active = resolvedArc === choice.id
        const position = arcPositionClasses[choice.id] ?? ''
        const label = choiceLabel(choice, locale, arcShortLabels[choice.id])
        return (
          <button
            type="button"
            key={choice.id}
            className={`shipdesigner-firing-compass-button ${position}${active ? ' active' : ''}`}
            data-weapon-firing-arc={choice.id}
            aria-pressed={active}
            aria-label={`${label}, ${costLabel(choice)}`}
            title={label}
            disabled={!selectable}
            onClick={() => onChange(choice.id)}
          >
            <strong>{arcShortLabels[choice.id]}</strong>
            <small>{costLabel(choice)}</small>
          </button>
        )
      })}

      <button
        type="button"
        className={`shipdesigner-firing-compass-center${resolvedArc === '360' ? ' active' : ''}`}
        data-weapon-firing-arc="360"
        aria-pressed={resolvedArc === '360'}
        aria-label={`${choiceLabel(centerChoice, locale, '360')}, ${costLabel(centerChoice)}`}
        title={choiceLabel(centerChoice, locale, '360')}
        disabled={!selectable}
        onClick={() => onChange('360')}
      >
        <ProceduralShipGlyph
          seed={shipGenome.seed}
          hullId={hullID}
          genome={shipGenome}
          weaponCount={weaponCount}
          className="shipdesigner-firing-compass-ship"
          tightViewBox
          normalizeVisualFootprint
        />
        <span className="shipdesigner-firing-compass-center-label">
          <strong>360°</strong>
          {centerChoice && <small>{costLabel(centerChoice)}</small>}
        </span>
      </button>

      {fixed && fixedLabel && <div className="shipdesigner-firing-compass-fixed">{fixedLabel}</div>}
    </div>
  )
}
