import {
  type MilitaryDesignSpecialRowPreview,
  type MilitaryDesignerSpecialChoice,
  type ShipSpecialMount,
} from '../api'
import { useI18n, type TranslationKey, type TranslationVars } from '../i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type Props = {
  choices: MilitaryDesignerSpecialChoice[]
  mounts: ShipSpecialMount[]
  previews?: MilitaryDesignSpecialRowPreview[]
  hullIndex: number
  disabled?: boolean
  error?: string
  onChange: (mounts: ShipSpecialMount[]) => void
  t: Translator
}

const labels: Record<string, { en: string; de: string }> = {
  achilles_targeting_unit: { en: 'Achilles Targeting Unit', de: 'Achilles-Zielsystem' },
  augmented_engines: { en: 'Augmented Engines', de: 'Verbesserte Triebwerke' },
  automated_repair_unit: { en: 'Automated Repair Unit', de: 'Automatische Reparatureinheit' },
  battle_pods: { en: 'Battle Pods', de: 'Kampfkapseln' },
  battle_scanner: { en: 'Battle Scanner', de: 'Gefechtsscanner' },
  cloaking_device: { en: 'Cloaking Device', de: 'Tarnvorrichtung' },
  damper_field: { en: 'Damper Field', de: 'Dämpfungsfeld' },
  displacement_device: { en: 'Displacement Device', de: 'Versatzvorrichtung' },
  ecm_jammer: { en: 'ECM Jammer', de: 'ECM-Störsender' },
  energy_absorber: { en: 'Energy Absorber', de: 'Energieabsorber' },
  extended_fuel_tanks: { en: 'Extended Fuel Tanks', de: 'Erweiterte Treibstofftanks' },
  fast_missile_racks: { en: 'Fast Missile Racks', de: 'Schnellfeuer-Raketenracks' },
  hard_shields: { en: 'Hard Shields', de: 'Gehärtete Schilde' },
  heavy_armor: { en: 'Heavy Armor', de: 'Schwere Panzerung' },
  high_energy_focus: { en: 'High Energy Focus', de: 'Hochenergie-Fokus' },
  hyper_x_capacitors: { en: 'Hyper-X Capacitors', de: 'Hyper-X-Kondensatoren' },
  inertial_nullifier: { en: 'Inertial Nullifier', de: 'Trägheitsnullifikator' },
  inertial_stabilizer: { en: 'Inertial Stabilizer', de: 'Trägheitsstabilisator' },
  lightning_field: { en: 'Lightning Field', de: 'Blitzfeld' },
  multi_phased_shields: { en: 'Multi-Phased Shields', de: 'Mehrphasen-Schilde' },
  multi_wave_ecm_jammer: { en: 'Multi-Wave ECM Jammer', de: 'Mehrwellen-ECM-Störsender' },
  phase_shifter: { en: 'Phase Shifter', de: 'Phasenschieber' },
  phasing_cloak: { en: 'Phasing Cloak', de: 'Phasen-Tarnung' },
  quantum_detonator: { en: 'Quantum Detonator', de: 'Quantendetonator' },
  rangemaster_unit: { en: 'Rangemaster Unit', de: 'Rangemaster-Einheit' },
  reflection_field: { en: 'Reflection Field', de: 'Reflexionsfeld' },
  reinforced_hull: { en: 'Reinforced Hull', de: 'Verstärkter Rumpf' },
  scout_lab: { en: 'Scout Lab', de: 'Aufklärungslabor' },
  security_stations: { en: 'Security Stations', de: 'Sicherheitsstationen' },
  shield_capacitors: { en: 'Shield Capacitors', de: 'Schildkondensatoren' },
  stealth_field: { en: 'Stealth Field', de: 'Tarnfeld' },
  structural_analyzer: { en: 'Structural Analyzer', de: 'Strukturanalysator' },
  sub_space_teleporter: { en: 'Sub-Space Teleporter', de: 'Subraum-Teleporter' },
  time_warp_facilitator: { en: 'Time Warp Facilitator', de: 'Zeitkrümmungs-Fazilitator' },
  transporters: { en: 'Transporters', de: 'Transporter' },
  troop_pods: { en: 'Troop Pods', de: 'Truppenkapseln' },
  warp_dissipater: { en: 'Warp Dissipater', de: 'Warp-Dissipator' },
  wide_area_jammer: { en: 'Wide Area Jammer', de: 'Weitbereichs-Störsender' },
}

function humanize(value: string): string {
  return value.replaceAll('_', ' ').replace(/w/g, (letter) => letter.toUpperCase())
}

function supportLabel(t: Translator, support: string): string {
  switch (support) {
    case 'strategic_only':
      return t('shipbuilder.specialSupportStrategic')
    case 'turning_mode':
      return t('shipbuilder.specialSupportTurning')
    default:
      return t('shipbuilder.specialSupportDeferred')
  }
}

export function SpecialSystemsPanel({
  choices,
  mounts,
  previews,
  hullIndex,
  disabled = false,
  error,
  onChange,
  t,
}: Props) {
  const { locale } = useI18n()
  const normalized = [...mounts].sort((a, b) => a.slot - b.slot)

  function changeSlot(slot: number, specialID: string) {
    const next = normalized.filter((mount) => mount.slot !== slot)
    if (specialID) next.push({ slot, special_id: specialID })
    next.sort((a, b) => a.slot - b.slot)
    onChange(next)
  }

  return (
    <section className="shipdesigner-specials-section" data-special-systems>
      <div className="shipdesigner-specials-heading">
        <div>
          <p className="eyebrow">{t('shipbuilder.specialSystems')}</p>
          <small className="muted">{t('shipbuilder.specialSystemsHint')}</small>
        </div>
        <span className="badge">{mounts.length}/8</span>
      </div>

      <div className="shipdesigner-specials-grid">
        {Array.from({ length: 8 }, (_, slot) => {
          const mount = normalized.find((candidate) => candidate.slot === slot)
          const choice = choices.find((candidate) => candidate.id === mount?.special_id)
          const preview = previews?.find((candidate) => candidate.slot === slot)
          const space = preview?.unit_space ?? choice?.effective_space_by_hull?.[hullIndex]
          const cost = preview?.unit_cost_pp ?? choice?.effective_cost_by_hull_pp?.[hullIndex]
          const support = preview?.tactical_support ?? choice?.tactical_support
          const labelSet = choice ? labels[choice.id] : undefined
          const label = choice ? (labelSet?.[locale === 'de' ? 'de' : 'en'] ?? humanize(choice.id)) : ''

          return (
            <article className={"shipdesigner-special-slot" + (mount ? " installed" : "")} key={slot} data-special-slot={slot}>
              <label>
                <span>{t('shipbuilder.specialSlot', { slot: slot + 1 })}</span>
                <select
                  value={mount?.special_id ?? ''}
                  disabled={disabled}
                  onChange={(event) => changeSlot(slot, event.target.value)}
                  aria-label={t('shipbuilder.specialSlot', { slot: slot + 1 })}
                >
                  <option value="">{t('shipbuilder.specialNone')}</option>
                  {choices.map((candidate) => {
                    const candidateLabels = labels[candidate.id]
                    const candidateLabel = candidateLabels?.[locale === 'de' ? 'de' : 'en'] ?? humanize(candidate.id)
                    const candidateSpace = candidate.effective_space_by_hull?.[hullIndex]
                    const candidateCost = candidate.effective_cost_by_hull_pp?.[hullIndex]
                    const suffix = candidateSpace == null || candidateCost == null ? '' : ` · ${candidateSpace} S · ${candidateCost} PP`
                    return (
                      <option
                        key={candidate.id}
                        value={candidate.id}
                        disabled={!candidate.available && candidate.id !== mount?.special_id}
                      >
                        {candidateLabel}{suffix}{!candidate.available ? ` · ${t('shipbuilder.specialLocked')}` : ''}
                      </option>
                    )
                  })}
                </select>
              </label>
              {mount && choice && (
                <div className="shipdesigner-special-meta">
                  <strong>{label}</strong>
                  <span>{space == null ? '—' : `${space} S`} · {cost == null ? '—' : `${cost} PP`}</span>
                  {support && <span className={"badge support " + support}>{supportLabel(t, support)}</span>}
                </div>
              )}
            </article>
          )
        })}
      </div>
      {error && <p className="shipbuilder-save-state error shipdesigner-specials-error">{error}</p>}
    </section>
  )
}
