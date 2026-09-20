import { useEffect, useMemo, useState } from 'react'
import {
  previewMilitaryDesign,
  type MilitaryDesignDraftProjection,
  type MilitaryDesignerWeaponChoice,
  type PlayerSnapshot,
  type ShipWeaponModifierID,
  type ShipWeaponMount,
  type ShipWeaponMountType,
} from '../api'
import type { TranslationKey, TranslationVars } from '../i18n'
import { GameIcon } from './GameIcon'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type AvailabilityFilter = 'all' | 'available' | 'locked'
type SortKey = 'technology' | 'name' | 'damage' | 'space' | 'pp'
type SortDirection = 'asc' | 'desc'
type ModifierFilter = 'all' | 'heavy' | 'point_defense' | ShipWeaponModifierID

type Props = {
  snapshot: PlayerSnapshot
  weapons: MilitaryDesignerWeaponChoice[]
  existingMounts: ShipWeaponMount[]
  hullID: string
  strategicPictureID: number
  computerID: string
  shieldID: string
  nextFreeSlot?: number
  onInstall: (mount: ShipWeaponMount, preview: MilitaryDesignDraftProjection) => void
  onClose: () => void
  t: Translator
}

const weaponTranslationKeys: Record<string, TranslationKey> = {
  mass_driver: 'shipbuilder.weapon.massDriver',
  gauss_cannon: 'shipbuilder.weapon.gaussCannon',
  laser_cannon: 'shipbuilder.weapon.laserCannon',
  particle_beam: 'shipbuilder.weapon.particleBeam',
  fusion_beam: 'shipbuilder.weapon.fusionBeam',
  ion_pulse_cannon: 'shipbuilder.weapon.ionPulseCannon',
  graviton_beam: 'shipbuilder.weapon.gravitonBeam',
  neutron_blaster: 'shipbuilder.weapon.neutronBlaster',
  phasor: 'shipbuilder.weapon.phasor',
  disrupter_cannon: 'shipbuilder.weapon.disrupterCannon',
  death_ray: 'shipbuilder.weapon.deathRay',
  plasma_cannon: 'shipbuilder.weapon.plasmaCannon',
  mauler_device: 'shipbuilder.weapon.maulerDevice',
}

const mountTranslationKeys: Record<ShipWeaponMountType, TranslationKey> = {
  standard: 'shipbuilder.mount.standard',
  heavy: 'shipbuilder.mount.heavy',
  point_defense: 'shipbuilder.mount.pointDefense',
}

const modifierTranslationKeys: Record<ShipWeaponModifierID, TranslationKey> = {
  armor_piercing: 'shipbuilder.modifier.armorPiercing',
  continuous: 'shipbuilder.modifier.continuous',
  no_range_dissipation: 'shipbuilder.modifier.noRangeDissipation',
  shield_piercing: 'shipbuilder.modifier.shieldPiercing',
  autofire: 'shipbuilder.modifier.autofire',
  enveloping: 'shipbuilder.modifier.enveloping',
}

function weaponLabel(t: Translator, weapon: MilitaryDesignerWeaponChoice): string {
  const key = weaponTranslationKeys[weapon.id]
  return key ? t(key) : weapon.id.replaceAll('_', ' ')
}

function modifierFilterLabel(t: Translator, filter: ModifierFilter): string {
  if (filter === 'all') return t('shipbuilder.filterAll')
  if (filter === 'heavy' || filter === 'point_defense') return t(mountTranslationKeys[filter])
  return t(modifierTranslationKeys[filter])
}

function damageLabel(weapon: MilitaryDesignerWeaponChoice): string {
  return weapon.min_damage === weapon.max_damage
    ? String(weapon.max_damage)
    : `${weapon.min_damage}–${weapon.max_damage}`
}

function cloneMounts(items: ShipWeaponMount[]): ShipWeaponMount[] {
  return items.map((item) => ({ ...item, modifier_ids: item.modifier_ids ? [...item.modifier_ids] : undefined }))
}

export function WeaponSelectorDialog({
  snapshot,
  weapons,
  existingMounts,
  hullID,
  strategicPictureID,
  computerID,
  shieldID,
  nextFreeSlot,
  onInstall,
  onClose,
  t,
}: Props) {
  const [search, setSearch] = useState('')
  const [availability, setAvailability] = useState<AvailabilityFilter>('all')
  const [modifierFilter, setModifierFilter] = useState<ModifierFilter>('all')
  const [sortKey, setSortKey] = useState<SortKey>('technology')
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
  const [selectedWeaponID, setSelectedWeaponID] = useState<string>('')
  const [mountType, setMountType] = useState<ShipWeaponMountType>('standard')
  const [modifierIDs, setModifierIDs] = useState<ShipWeaponModifierID[]>([])
  const [quantity, setQuantity] = useState(1)
  const [preview, setPreview] = useState<MilitaryDesignDraftProjection | null>(null)
  const [previewing, setPreviewing] = useState(false)
  const [previewError, setPreviewError] = useState('')

  const selectedWeapon = weapons.find((weapon) => weapon.id === selectedWeaponID)
  const normalizedSearch = search.trim().toLocaleLowerCase()

  const filteredWeapons = useMemo(() => {
    const items = weapons.filter((weapon) => {
      const label = weaponLabel(t, weapon).toLocaleLowerCase()
      if (normalizedSearch && !label.includes(normalizedSearch) && !weapon.id.includes(normalizedSearch)) return false
      if (availability === 'available' && !weapon.available) return false
      if (availability === 'locked' && weapon.available) return false
      if (modifierFilter !== 'all') {
        if (modifierFilter === 'heavy' || modifierFilter === 'point_defense') {
          if (!weapon.mount_choices.some((choice) => choice.id === modifierFilter)) return false
        } else if (!weapon.modifier_choices.some((choice) => choice.id === modifierFilter)) {
          return false
        }
      }
      return true
    })
    const direction = sortDirection === 'asc' ? 1 : -1
    return [...items].sort((a, b) => {
      let delta = 0
      switch (sortKey) {
      case 'name':
        delta = weaponLabel(t, a).localeCompare(weaponLabel(t, b))
        break
      case 'damage':
        delta = a.max_damage - b.max_damage || a.min_damage - b.min_damage
        break
      case 'space':
        delta = a.effective_standard_space - b.effective_standard_space
        break
      case 'pp':
        delta = a.effective_standard_cost_pp - b.effective_standard_cost_pp
        break
      default:
        delta = a.technology_field_id - b.technology_field_id || a.original_index - b.original_index
      }
      return delta * direction
    })
  }, [weapons, normalizedSearch, availability, modifierFilter, sortKey, sortDirection, t])

  const candidateMount = useMemo<ShipWeaponMount | null>(() => {
    if (!selectedWeapon || nextFreeSlot == null) return null
    return {
      slot: nextFreeSlot,
      weapon_id: selectedWeapon.id,
      count: quantity,
      mount_type: mountType,
      ...(modifierIDs.length > 0 ? { modifier_ids: [...modifierIDs] } : {}),
    }
  }, [selectedWeapon, nextFreeSlot, quantity, mountType, modifierIDs])

  useEffect(() => {
    if (!candidateMount || !selectedWeapon?.available) {
      setPreview(null)
      setPreviewing(false)
      setPreviewError('')
      return
    }
    let cancelled = false
    setPreviewing(true)
    setPreviewError('')
    const draftWeapons = [...cloneMounts(existingMounts), candidateMount].sort((a, b) => a.slot - b.slot)
    void previewMilitaryDesign(snapshot, snapshot.view.seat.seat.id, {
      hull_id: hullID,
      strategic_picture_id: strategicPictureID,
      computer_id: computerID,
      shield_id: shieldID,
      weapons: draftWeapons,
    }).then((result) => {
      if (cancelled) return
      setPreview(result.preview.projection)
    }).catch((error: unknown) => {
      if (cancelled) return
      setPreview(null)
      setPreviewError(error instanceof Error ? error.message : String(error))
    }).finally(() => {
      if (!cancelled) setPreviewing(false)
    })
    return () => { cancelled = true }
  }, [
    candidateMount?.slot,
    candidateMount?.weapon_id,
    candidateMount?.count,
    candidateMount?.mount_type,
    candidateMount?.modifier_ids?.join('|'),
    selectedWeapon?.available,
    existingMounts,
    hullID,
    strategicPictureID,
    computerID,
    shieldID,
    snapshot,
  ])

  function selectWeapon(id: string) {
    setSelectedWeaponID(id)
    setMountType('standard')
    setModifierIDs([])
    setQuantity(1)
    setPreview(null)
    setPreviewError('')
  }

  function toggleModifier(id: ShipWeaponModifierID) {
    if (!selectedWeapon) return
    const choice = selectedWeapon.modifier_choices.find((candidate) => candidate.id === id)
    if (!choice?.available) return
    setModifierIDs((current) => {
      const enabled = current.includes(id)
      const next = enabled ? current.filter((candidate) => candidate !== id) : [...current, id]
      const order = new Map(selectedWeapon.modifier_choices.map((candidate, index) => [candidate.id, index]))
      return next.sort((a, b) => (order.get(a) ?? 99) - (order.get(b) ?? 99))
    })
  }

  const candidatePreview = candidateMount
    ? preview?.weapons?.find((row) => row.slot === candidateMount.slot)
    : undefined
  const canInstall = Boolean(
    selectedWeapon?.available
    && candidateMount
    && nextFreeSlot != null
    && preview?.save_available
    && candidatePreview
    && !previewing,
  )

  function install() {
    if (!canInstall || !candidatePreview || !preview) return
    onInstall({
      slot: candidatePreview.slot,
      weapon_id: candidatePreview.weapon_id,
      count: candidatePreview.count,
      mount_type: candidatePreview.mount_type,
      ...(candidatePreview.modifier_ids?.length ? { modifier_ids: [...candidatePreview.modifier_ids] } : {}),
    }, preview)
  }

  const modifierFilterOptions: ModifierFilter[] = [
    'all', 'heavy', 'point_defense',
    'armor_piercing', 'continuous', 'no_range_dissipation', 'shield_piercing', 'autofire', 'enveloping',
  ]

  return (
    <div className="shipdesigner-system-dialog-backdrop shipdesigner-weapon-selector-backdrop" role="presentation" onPointerDown={(event) => { if (event.target === event.currentTarget) onClose() }}>
      <section className="shipdesigner-system-dialog shipdesigner-weapon-selector" role="dialog" aria-modal="true" aria-labelledby="shipdesigner-weapon-selector-title">
        <header className="shipdesigner-system-dialog-header">
          <div>
            <p className="eyebrow">{t('shipbuilder.weapons')}</p>
            <h3 id="shipdesigner-weapon-selector-title">{t('shipbuilder.weaponSelectorTitle')}</h3>
            <p className="muted">{t('shipbuilder.weaponSelectorHint')}</p>
          </div>
          <button type="button" className="button-ghost shipdesigner-system-dialog-close" onClick={onClose} aria-label={t('common.close')}><GameIcon name="close" /></button>
        </header>

        <div className="shipdesigner-weapon-selector-toolbar">
          <label className="shipdesigner-weapon-search">
            <span>{t('shipbuilder.searchWeapons')}</span>
            <input value={search} onChange={(event) => setSearch(event.target.value)} />
          </label>
          <label>
            <span>{t('shipbuilder.filterAvailability')}</span>
            <select value={availability} onChange={(event) => setAvailability(event.target.value as AvailabilityFilter)}>
              <option value="all">{t('shipbuilder.filterAll')}</option>
              <option value="available">{t('shipbuilder.filterAvailable')}</option>
              <option value="locked">{t('shipbuilder.filterLocked')}</option>
            </select>
          </label>
          <label>
            <span>{t('shipbuilder.filterModifier')}</span>
            <select value={modifierFilter} onChange={(event) => setModifierFilter(event.target.value as ModifierFilter)}>
              {modifierFilterOptions.map((option) => <option value={option} key={option}>{modifierFilterLabel(t, option)}</option>)}
            </select>
          </label>
          <label>
            <span>{t('shipbuilder.sortBy')}</span>
            <select value={sortKey} onChange={(event) => setSortKey(event.target.value as SortKey)}>
              <option value="technology">{t('shipbuilder.sortTechnology')}</option>
              <option value="name">{t('shipbuilder.sortName')}</option>
              <option value="damage">{t('shipbuilder.sortDamage')}</option>
              <option value="space">{t('shipbuilder.sortSpace')}</option>
              <option value="pp">{t('shipbuilder.sortPP')}</option>
            </select>
          </label>
          <button type="button" className="button-ghost shipdesigner-sort-direction" onClick={() => setSortDirection((current) => current === 'asc' ? 'desc' : 'asc')} aria-label={sortDirection === 'asc' ? 'Ascending' : 'Descending'}>
            {sortDirection === 'asc' ? '↑' : '↓'}
          </button>
        </div>

        <div className="shipdesigner-weapon-selector-body">
          <div className="shipdesigner-weapon-selector-list" role="listbox" aria-label={t('shipbuilder.weaponSelectorTitle')}>
            {filteredWeapons.map((weapon) => {
              const selected = weapon.id === selectedWeaponID
              return (
                <button
                  type="button"
                  role="option"
                  aria-selected={selected}
                  className={`shipdesigner-weapon-selector-row${selected ? ' selected' : ''}${weapon.available ? '' : ' locked'}`}
                  key={weapon.id}
                  onClick={() => selectWeapon(weapon.id)}
                >
                  <span className="shipdesigner-weapon-selector-name">
                    <strong>{weaponLabel(t, weapon)}</strong>
                    {!weapon.available && <small>{t('shipbuilder.filterLocked')}</small>}
                  </span>
                  <span>{damageLabel(weapon)}</span>
                  <span>{weapon.effective_standard_space} {t('shipbuilder.space')}</span>
                  <span>{weapon.effective_standard_cost_pp} PP</span>
                </button>
              )
            })}
          </div>

          <aside className="shipdesigner-weapon-config">
            {selectedWeapon ? (
              <>
                <div className="shipdesigner-weapon-config-title">
                  <div>
                    <p className="eyebrow">{t('shipbuilder.weaponSelectorTitle')}</p>
                    <h4>{weaponLabel(t, selectedWeapon)}</h4>
                  </div>
                  <span className="badge">{damageLabel(selectedWeapon)}</span>
                </div>

                <fieldset className="shipdesigner-weapon-config-group" disabled={!selectedWeapon.available}>
                  <legend>{t('shipbuilder.mountType')}</legend>
                  <div className="shipdesigner-weapon-option-grid">
                    {selectedWeapon.mount_choices.map((choice) => (
                      <label className={`shipdesigner-weapon-option${choice.available ? '' : ' locked'}`} key={choice.id}>
                        <input
                          type="radio"
                          name="beam-mount"
                          value={choice.id}
                          checked={mountType === choice.id}
                          disabled={!choice.available}
                          onChange={() => setMountType(choice.id)}
                        />
                        <span>{t(mountTranslationKeys[choice.id])}</span>
                      </label>
                    ))}
                  </div>
                </fieldset>

                <fieldset className="shipdesigner-weapon-config-group" disabled={!selectedWeapon.available}>
                  <legend>{t('shipbuilder.specialModifiers')}</legend>
                  <div className="shipdesigner-weapon-option-grid">
                    {selectedWeapon.modifier_choices.length > 0 ? selectedWeapon.modifier_choices.map((choice) => (
                      <label className={`shipdesigner-weapon-option${choice.available ? '' : ' locked'}`} key={choice.id}>
                        <input
                          type="checkbox"
                          checked={modifierIDs.includes(choice.id)}
                          disabled={!choice.available}
                          onChange={() => toggleModifier(choice.id)}
                        />
                        <span>
                          {t(modifierTranslationKeys[choice.id])}
                          {!choice.available && <small>{t('shipbuilder.refinementRequired', { level: choice.required_mini_level })}</small>}
                        </span>
                      </label>
                    )) : <p className="muted">{t('common.none')}</p>}
                  </div>
                </fieldset>

                <div className="shipdesigner-weapon-config-bottom">
                  <label className="shipdesigner-weapon-config-quantity">
                    <span>{t('shipbuilder.quantity')}</span>
                    <div className="shipdesigner-weapon-quantity">
                      <button type="button" className="button-ghost" disabled={quantity <= 1} onClick={() => setQuantity((value) => Math.max(1, value - 1))}>-</button>
                      <strong className="shipdesigner-weapon-count">{quantity}</strong>
                      <button type="button" className="button-ghost" onClick={() => setQuantity((value) => value + 1)}>+</button>
                    </div>
                  </label>
                  <div className="shipdesigner-weapon-preview">
                    {previewing && <small>{t('shipbuilder.previewing')}</small>}
                    {!previewing && candidatePreview && (
                      <strong>{t('shipbuilder.currentWeaponTotals', { space: candidatePreview.row_space, cost: candidatePreview.row_cost_pp })}</strong>
                    )}
                    {!previewing && preview && <small>{preview.remaining_space} {t('shipbuilder.spaceAvailable')}</small>}
                    {!previewing && preview && !preview.save_available && <small className="error">{t('shipbuilder.previewRejected')}</small>}
                    {previewError && <small className="error">{previewError}</small>}
                  </div>
                </div>

                <footer className="shipdesigner-system-dialog-actions shipdesigner-weapon-selector-actions">
                  <button type="button" className="button-ghost" onClick={onClose}>{t('common.close')}</button>
                  <button type="button" className="button-primary" disabled={!canInstall} onClick={install}><span aria-hidden="true">+</span>{t('shipbuilder.install')}</button>
                </footer>
              </>
            ) : (
              <div className="shipdesigner-weapon-config-empty">{t('shipbuilder.weaponSelectorHint')}</div>
            )}
          </aside>
        </div>
      </section>
    </div>
  )
}
