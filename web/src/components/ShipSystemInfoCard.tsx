import { useEffect, useState } from 'react'
import {
  type MilitaryDesignerCatalog,
  type MilitaryDesignerHullChoice,
  type MilitaryDesignerWeaponChoice,
  type ShipDesignSpec,
  type ShipWeaponMount,
} from '../api'
import { GameIcon } from './GameIcon'
import type { TranslationKey, TranslationVars } from '../i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string
type ShipSystemKind = 'drive' | 'computer' | 'armor' | 'shield' | 'fuel'
type DialogState = { system: ShipSystemKind; mode: 'info' | 'select' } | null

type Props = {
  designer: MilitaryDesignerCatalog
  spec: ShipDesignSpec
  currentHull?: MilitaryDesignerHullChoice
  computerID: string
  shieldID: string
  weaponMounts: ShipWeaponMount[]
  laser?: MilitaryDesignerWeaponChoice
  remainingSpace: number
  beamDefense: number
  missileEvasion: number
  onComputerChange: (id: string) => void
  onShieldChange: (id: string) => void
  onWeaponCountChange: (slot: number, delta: number) => void
  onRemoveWeapon: (slot: number) => void
  t: Translator
}

const componentTranslationKeys: Record<string, TranslationKey> = {
  electronic_computer: 'shipbuilder.component.electronicComputer',
  optronic_computer: 'shipbuilder.component.optronicComputer',
  positronic_computer: 'shipbuilder.component.positronicComputer',
  cybertronic_computer: 'shipbuilder.component.cybertronicComputer',
  moleculartronic_computer: 'shipbuilder.component.moleculartronicComputer',
  class_i_shield: 'shipbuilder.component.classIShield',
  class_iii_shield: 'shipbuilder.component.classIIIShield',
  class_v_shield: 'shipbuilder.component.classVShield',
  class_vii_shield: 'shipbuilder.component.classVIIShield',
  class_x_shield: 'shipbuilder.component.classXShield',
  nuclear_drive: 'shipbuilder.component.nuclearDrive',
  fusion_drive: 'shipbuilder.component.fusionDrive',
  ion_drive: 'shipbuilder.component.ionDrive',
  anti_matter_drive: 'shipbuilder.component.antiMatterDrive',
  hyper_drive: 'shipbuilder.component.hyperDrive',
  interphased_drive: 'shipbuilder.component.interphasedDrive',
  titanium_armor: 'shipbuilder.component.titaniumArmor',
  tritanium_armor: 'shipbuilder.component.tritaniumArmor',
  zortrium_armor: 'shipbuilder.component.zortriumArmor',
  neutronium_armor: 'shipbuilder.component.neutroniumArmor',
  adamantium_armor: 'shipbuilder.component.adamantiumArmor',
  xentronium_armor: 'shipbuilder.component.xentroniumArmor',
  standard_fuel_cells: 'shipbuilder.component.standardFuelCells',
  deuterium_fuel_cells: 'shipbuilder.component.deuteriumFuelCells',
  iridium_fuel_cells: 'shipbuilder.component.iridiumFuelCells',
  urridium_fuel_cells: 'shipbuilder.component.urridiumFuelCells',
  thorium_fuel_cells: 'shipbuilder.component.thoriumFuelCells',
}

function componentLabel(t: Translator, id: string): string {
  const key = componentTranslationKeys[id]
  return key ? t(key) : id.replaceAll('_', ' ')
}

function componentCostText(t: Translator, cost: number, reduction: number): string {
  return reduction > 0
    ? t('shipbuilder.componentCostReduction', { cost, reduction })
    : t('shipbuilder.componentCostBase', { cost })
}

function shieldCostText(t: Translator, cost: number, space: number, costReduction: number, spaceReduction: number): string {
  return costReduction > 0 || spaceReduction > 0
    ? t('shipbuilder.shieldCostReduction', { cost, space, costReduction, spaceReduction })
    : t('shipbuilder.shieldCostBase', { cost, space })
}

export function ShipSystemInfoCard({
  designer,
  spec,
  currentHull,
  computerID,
  shieldID,
  weaponMounts,
  laser,
  remainingSpace,
  beamDefense,
  missileEvasion,
  onComputerChange,
  onShieldChange,
  onWeaponCountChange,
  onRemoveWeapon,
  t,
}: Props) {
  const [dialog, setDialog] = useState<DialogState>(null)
  const hullSizeIndex = currentHull?.size_index ?? 0
  const hullClass = hullSizeIndex + 1
  const computers = designer.computers ?? []
  const shields = designer.shields ?? []
  const selectedComputer = computers.find((choice) => choice.id === computerID)
  const selectedShield = shieldID ? shields.find((choice) => choice.id === shieldID) : undefined
  const selectedComputerCost = selectedComputer?.effective_cost_by_hull_pp?.[hullSizeIndex] ?? 0
  const selectedShieldCost = selectedShield?.effective_cost_by_hull_pp?.[hullSizeIndex] ?? 0
  const selectedShieldSpace = selectedShield?.effective_space_by_hull?.[hullSizeIndex] ?? 0
  const selectedShieldCapacity = (selectedShield?.shield_capacity_per_hull_class ?? 0) * hullClass

  useEffect(() => {
    if (!dialog) return
    const onKeyDown = (event: globalThis.KeyboardEvent) => {
      if (event.key === 'Escape') setDialog(null)
    }
    document.body.classList.add('modal-open')
    window.addEventListener('keydown', onKeyDown)
    return () => {
      window.removeEventListener('keydown', onKeyDown)
      document.body.classList.remove('modal-open')
    }
  }, [dialog])

  const closeDialog = () => setDialog(null)
  const chooseComputer = (id: string) => {
    onComputerChange(id)
    closeDialog()
  }
  const chooseShield = (id: string) => {
    onShieldChange(id)
    closeDialog()
  }

  return (
    <>
      <section className="shipdesigner-component-section shipdesigner-ship-info-card">
        <div className="card-heading shipdesigner-ship-info-heading">
          <div>
            <p className="eyebrow">{t('shipbuilder.installed')}</p>
            <h3>{t('shipbuilder.shipInfoTitle')}</h3>
          </div>
        </div>

        <div className="shipdesigner-system-list">
          <article className="shipdesigner-system-row" data-system="drive">
            <div className="shipdesigner-system-row-main">
              <span>{t('shipbuilder.drive')}</span>
              <strong>{componentLabel(t, spec.warp_drive_id)}</strong>
            </div>
            <button type="button" className="shipdesigner-system-help" onClick={() => setDialog({ system: 'drive', mode: 'info' })} aria-label={t('shipbuilder.systemDetails', { system: t('shipbuilder.drive') })}>?</button>
          </article>

          <article className="shipdesigner-system-row selectable" data-system="computer">
            <button type="button" className="shipdesigner-system-row-main shipdesigner-system-select-trigger" onClick={() => setDialog({ system: 'computer', mode: 'select' })}>
              <span>{t('shipbuilder.computer')}</span>
              <strong>{componentLabel(t, spec.computer_id)}</strong>
            </button>
            <button type="button" className="shipdesigner-system-help" onClick={() => setDialog({ system: 'computer', mode: 'info' })} aria-label={t('shipbuilder.systemDetails', { system: t('shipbuilder.computer') })}>?</button>
          </article>

          <article className="shipdesigner-system-row" data-system="armor">
            <div className="shipdesigner-system-row-main">
              <span>{t('shipbuilder.armor')}</span>
              <strong>{componentLabel(t, spec.armor_id)}</strong>
            </div>
            <button type="button" className="shipdesigner-system-help" onClick={() => setDialog({ system: 'armor', mode: 'info' })} aria-label={t('shipbuilder.systemDetails', { system: t('shipbuilder.armor') })}>?</button>
          </article>

          <article className="shipdesigner-system-row selectable" data-system="shield">
            <button type="button" className="shipdesigner-system-row-main shipdesigner-system-select-trigger" onClick={() => setDialog({ system: 'shield', mode: 'select' })}>
              <span>{t('shipbuilder.shield')}</span>
              <strong>{spec.shield_id ? componentLabel(t, spec.shield_id) : t('common.none')}</strong>
            </button>
            <button type="button" className="shipdesigner-system-help" onClick={() => setDialog({ system: 'shield', mode: 'info' })} aria-label={t('shipbuilder.systemDetails', { system: t('shipbuilder.shield') })}>?</button>
          </article>

          <article className="shipdesigner-system-row" data-system="fuel">
            <div className="shipdesigner-system-row-main">
              <span>{t('shipbuilder.fuel')}</span>
              <strong>{componentLabel(t, spec.fuel_cell_id)}</strong>
            </div>
            <button type="button" className="shipdesigner-system-help" onClick={() => setDialog({ system: 'fuel', mode: 'info' })} aria-label={t('shipbuilder.systemDetails', { system: t('shipbuilder.fuel') })}>?</button>
          </article>
        </div>

        <div className="shipdesigner-combat-ratings" aria-label={t('shipbuilder.combatRatings')}>
          <div><span>{t('shipbuilder.beamDefense')}</span><strong>+{beamDefense}</strong></div>
          <div><span>{t('shipbuilder.missileEvasion')}</span><strong>{missileEvasion}%</strong></div>
        </div>

        {weaponMounts.length > 0 && (
          <div className="shipdesigner-installed-weapons">
            <p className="eyebrow">{t('shipbuilder.installedComponents')}</p>
            <div className="shipdesigner-installed-info-grid">
              {weaponMounts.map((mount) => (
                <article className="shipdesigner-info-card weapon" key={mount.slot}>
                  <span>{t('shipbuilder.slot', { slot: mount.slot + 1 })}</span>
                  <strong>{mount.count}× {t('shipbuilder.laserCannon')}</strong>
                  {laser && <p>{t('shipbuilder.weaponEffect', {
                    min: mount.count * laser.min_damage,
                    max: mount.count * laser.max_damage,
                    space: mount.count * laser.base_space,
                    cost: mount.count * laser.base_cost_pp,
                  })}</p>}
                  <div className="shipdesigner-weapon-quantity">
                    <button type="button" className="button-ghost" aria-label={t('shipbuilder.decreaseQuantity')} onClick={() => onWeaponCountChange(mount.slot, -1)}>-</button>
                    <strong className="shipdesigner-weapon-count">{mount.count}</strong>
                    <button type="button" className="button-ghost" aria-label={t('shipbuilder.increaseQuantity')} disabled={!laser || remainingSpace < laser.base_space} onClick={() => onWeaponCountChange(mount.slot, 1)}>+</button>
                    <button type="button" className="button-ghost shipdesigner-weapon-remove" onClick={() => onRemoveWeapon(mount.slot)}>{t('shipbuilder.remove')}</button>
                  </div>
                </article>
              ))}
            </div>
          </div>
        )}

        <p className="muted shipdesigner-future-note">{t('shipbuilder.systemEffectScope')}</p>
      </section>

      {dialog && (
        <div className="shipdesigner-system-dialog-backdrop" role="presentation" onPointerDown={(event) => { if (event.target === event.currentTarget) closeDialog() }}>
          <section className="shipdesigner-system-dialog" role="dialog" aria-modal="true" aria-labelledby="shipdesigner-system-dialog-title">
            <header className="shipdesigner-system-dialog-header">
              <div>
                <p className="eyebrow">{dialog.mode === 'select' ? t('shipbuilder.chooseSystem') : t('shipbuilder.systemDetailsTitle')}</p>
                <h3 id="shipdesigner-system-dialog-title">
                  {dialog.system === 'drive' && t('shipbuilder.drive')}
                  {dialog.system === 'computer' && t('shipbuilder.computer')}
                  {dialog.system === 'armor' && t('shipbuilder.armor')}
                  {dialog.system === 'shield' && t('shipbuilder.shield')}
                  {dialog.system === 'fuel' && t('shipbuilder.fuel')}
                </h3>
              </div>
              <button type="button" className="button-ghost shipdesigner-system-dialog-close" onClick={closeDialog} aria-label={t('common.close')}><GameIcon name="close" /></button>
            </header>

            {dialog.mode === 'select' && dialog.system === 'computer' && (
              <>
                <p className="muted">{t('shipbuilder.systemSelectionHint')}</p>
                <div className="shipdesigner-system-dialog-choice-grid shipdesigner-computer-dialog">
                  {computers.map((choice) => {
                    const cost = choice.effective_cost_by_hull_pp?.[hullSizeIndex] ?? 0
                    const selected = computerID === choice.id
                    return (
                      <button
                        type="button"
                        key={choice.id}
                        data-component-id={choice.id}
                        className={"shipdesigner-choice-card" + (selected ? " selected" : "") + (!choice.available ? " locked" : "")}
                        aria-pressed={selected}
                        disabled={!choice.available}
                        onClick={() => chooseComputer(choice.id)}
                      >
                        <span className="shipdesigner-choice-heading"><strong>{componentLabel(t, choice.id)}</strong>{selected && <span className="badge">{t('shipbuilder.selected')}</span>}</span>
                        <span className="shipdesigner-choice-effect">{t('shipbuilder.computerEffect', { bonus: choice.beam_attack_bonus ?? 0 })}</span>
                        <small>{componentCostText(t, cost, choice.cost_reduction_percent ?? 0)}</small>
                      </button>
                    )
                  })}
                </div>
              </>
            )}

            {dialog.mode === 'select' && dialog.system === 'shield' && (
              <>
                <p className="muted">{t('shipbuilder.systemSelectionHint')}</p>
                <div className="shipdesigner-system-dialog-choice-grid shipdesigner-shield-dialog">
                  <button
                    type="button"
                    data-shield-id=""
                    className={"shipdesigner-choice-card" + (!shieldID ? " selected" : "")}
                    aria-pressed={!shieldID}
                    onClick={() => chooseShield('')}
                  >
                    <span className="shipdesigner-choice-heading"><strong>{t('common.none')}</strong>{!shieldID && <span className="badge">{t('shipbuilder.selected')}</span>}</span>
                    <span className="shipdesigner-choice-effect">{t('shipbuilder.noShieldEffect')}</span>
                    <small>{t('shipbuilder.noShieldCost')}</small>
                  </button>
                  {shields.map((choice) => {
                    const cost = choice.effective_cost_by_hull_pp?.[hullSizeIndex] ?? 0
                    const space = choice.effective_space_by_hull?.[hullSizeIndex] ?? 0
                    const selected = shieldID === choice.id
                    const capacity = (choice.shield_capacity_per_hull_class ?? 0) * hullClass
                    return (
                      <button
                        type="button"
                        key={choice.id}
                        data-shield-id={choice.id}
                        className={"shipdesigner-choice-card" + (selected ? " selected" : "") + (!choice.available ? " locked" : "")}
                        aria-pressed={selected}
                        disabled={!choice.available}
                        onClick={() => chooseShield(choice.id)}
                      >
                        <span className="shipdesigner-choice-heading"><strong>{componentLabel(t, choice.id)}</strong>{selected && <span className="badge">{t('shipbuilder.selected')}</span>}</span>
                        <span className="shipdesigner-choice-effect">{t('shipbuilder.shieldEffectCurrent', {
                          block: choice.shield_damage_block_per_hit ?? 0,
                          capacity,
                          perClass: choice.shield_capacity_per_hull_class ?? 0,
                          hullClass,
                        })}</span>
                        <small>{shieldCostText(t, cost, space, choice.cost_reduction_percent ?? 0, choice.space_reduction_percent ?? 0)}</small>
                      </button>
                    )
                  })}
                </div>
              </>
            )}

            {dialog.mode === 'info' && (
              <div className="shipdesigner-system-dialog-detail">
                {dialog.system === 'drive' && (
                  <>
                    <strong>{componentLabel(t, spec.warp_drive_id)}</strong>
                    <p>{t('shipbuilder.driveEffect', { ftl: spec.ftl_speed, tactical: designer.automatic_systems.tactical_speed_bonus ?? 0 })}</p>
                    <span className="badge">{t('shipbuilder.automaticSystem')}</span>
                  </>
                )}
                {dialog.system === 'computer' && (
                  <>
                    <strong>{componentLabel(t, spec.computer_id)}</strong>
                    <p>{t('shipbuilder.computerEffect', { bonus: selectedComputer?.beam_attack_bonus ?? 0 })}</p>
                    <small>{componentCostText(t, selectedComputerCost, selectedComputer?.cost_reduction_percent ?? 0)}</small>
                    <button type="button" className="button-secondary" onClick={() => setDialog({ system: 'computer', mode: 'select' })}>{t('shipbuilder.change')}</button>
                  </>
                )}
                {dialog.system === 'armor' && (
                  <>
                    <strong>{componentLabel(t, spec.armor_id)}</strong>
                    <p>{t('shipbuilder.armorEffect', { multiplier: designer.automatic_systems.armor_durability_multiple ?? 1 })}</p>
                    <span className="badge">{t('shipbuilder.automaticSystem')}</span>
                  </>
                )}
                {dialog.system === 'shield' && (
                  <>
                    <strong>{spec.shield_id ? componentLabel(t, spec.shield_id) : t('common.none')}</strong>
                    <p>{selectedShield
                      ? t('shipbuilder.shieldEffectCurrent', {
                          block: selectedShield.shield_damage_block_per_hit ?? 0,
                          capacity: selectedShieldCapacity,
                          perClass: selectedShield.shield_capacity_per_hull_class ?? 0,
                          hullClass,
                        })
                      : t('shipbuilder.noShieldEffect')}</p>
                    <small>{selectedShield
                      ? shieldCostText(t, selectedShieldCost, selectedShieldSpace, selectedShield.cost_reduction_percent ?? 0, selectedShield.space_reduction_percent ?? 0)
                      : t('shipbuilder.noShieldCost')}</small>
                    <button type="button" className="button-secondary" onClick={() => setDialog({ system: 'shield', mode: 'select' })}>{t('shipbuilder.change')}</button>
                  </>
                )}
                {dialog.system === 'fuel' && (
                  <>
                    <strong>{componentLabel(t, spec.fuel_cell_id)}</strong>
                    <p>{spec.fuel_range_parsecs >= 255 ? t('shipbuilder.fuelUnlimitedEffect') : t('shipbuilder.fuelEffect', { range: spec.fuel_range_parsecs })}</p>
                    <span className="badge">{t('shipbuilder.automaticSystem')}</span>
                  </>
                )}

                <div className="shipdesigner-system-dialog-scope">
                  <h4>{t('shipbuilder.whatItDoes')}</h4>
                  <p>{t('shipbuilder.systemEffectScope')}</p>
                </div>
              </div>
            )}

            <footer className="shipdesigner-system-dialog-actions">
              <button type="button" className="button-secondary" onClick={closeDialog}>{t('common.close')}</button>
            </footer>
          </section>
        </div>
      )}
    </>
  )
}
