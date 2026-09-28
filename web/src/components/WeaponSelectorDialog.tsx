import { useEffect, useMemo, useState } from 'react'
import {
  previewMilitaryDesign,
  type MilitaryDesignDraftProjection,
  type MilitaryDesignerWeaponChoice,
  type PlayerSnapshot,
  type ShipSpecialMount,
  type ShipWeaponFiringArcID,
  type ShipWeaponModifierID,
  type ShipWeaponMount,
  type ShipWeaponMountType,
} from '../api'
import { useI18n, type TranslationKey, type TranslationVars } from '../i18n'
import type { ShipVisualGenome } from '../shipVisualGenome'
import { FiringArcCompass } from './FiringArcCompass'
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
  specials: ShipSpecialMount[]
  hullID: string
  strategicPictureID: number
  shipGenome: ShipVisualGenome
  computerID: string
  shieldID: string
  nextFreeSlot?: number
  availableSpace: number
  editingMount?: ShipWeaponMount
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
  nuclear_missile: 'shipbuilder.weapon.nuclearMissile',
  merculite_missile: 'shipbuilder.weapon.merculiteMissile',
  pulson_missile: 'shipbuilder.weapon.pulsonMissile',
  zeon_missile: 'shipbuilder.weapon.zeonMissile',
  anti_matter_torpedoes: 'shipbuilder.weapon.antiMatterTorpedoes',
  proton_torpedoes: 'shipbuilder.weapon.protonTorpedoes',
  plasma_torpedoes: 'shipbuilder.weapon.plasmaTorpedoes',
  nuclear_bomb: 'shipbuilder.weapon.nuclearBomb',
  fusion_bomb: 'shipbuilder.weapon.fusionBomb',
  anti_matter_bomb: 'shipbuilder.weapon.antiMatterBomb',
  neutronium_bomb: 'shipbuilder.weapon.neutroniumBomb',
  death_spores: 'shipbuilder.weapon.deathSpores',
  bio_terminator: 'shipbuilder.weapon.bioTerminator',
  fighter_bays: 'shipbuilder.weapon.fighterBays',
  bomber_bays: 'shipbuilder.weapon.bomberBays',
  heavy_fighter_bays: 'shipbuilder.weapon.heavyFighterBays',
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
  mirv: 'shipbuilder.modifier.mirv',
  eccm: 'shipbuilder.modifier.eccm',
  armored: 'shipbuilder.modifier.armored',
  fast: 'shipbuilder.modifier.fast',
  emissions_guidance: 'shipbuilder.modifier.emissionsGuidance',
  overloaded: 'shipbuilder.modifier.overloaded',
}

const mountHelpTranslationKeys: Record<ShipWeaponMountType, TranslationKey> = {
  standard: 'shipbuilder.mount.standardHelp',
  heavy: 'shipbuilder.mount.heavyHelp',
  point_defense: 'shipbuilder.mount.pointDefenseHelp',
}

const modifierHelpTranslationKeys: Record<ShipWeaponModifierID, TranslationKey> = {
  armor_piercing: 'shipbuilder.modifier.armorPiercingHelp',
  continuous: 'shipbuilder.modifier.continuousHelp',
  no_range_dissipation: 'shipbuilder.modifier.noRangeDissipationHelp',
  shield_piercing: 'shipbuilder.modifier.shieldPiercingHelp',
  autofire: 'shipbuilder.modifier.autofireHelp',
  enveloping: 'shipbuilder.modifier.envelopingHelp',
  mirv: 'shipbuilder.modifier.mirvHelp',
  eccm: 'shipbuilder.modifier.eccmHelp',
  armored: 'shipbuilder.modifier.armoredHelp',
  fast: 'shipbuilder.modifier.fastHelp',
  emissions_guidance: 'shipbuilder.modifier.emissionsGuidanceHelp',
  overloaded: 'shipbuilder.modifier.overloadedHelp',
}

function percentDeltaLabel(value: number): string {
  if (value === 0) return '0%'
  return (value > 0 ? '+' : '') + value + '%'
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
  specials,
  hullID,
  strategicPictureID,
  shipGenome,
  computerID,
  shieldID,
  nextFreeSlot,
  availableSpace,
  editingMount,
  onInstall,
  onClose,
  t,
}: Props) {
  const { locale } = useI18n()
  const [search, setSearch] = useState('')
  const [availability, setAvailability] = useState<AvailabilityFilter>('all')
  const [modifierFilter, setModifierFilter] = useState<ModifierFilter>('all')
  const [sortKey, setSortKey] = useState<SortKey>('technology')
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
  const [selectedWeaponID, setSelectedWeaponID] = useState<string>(() => editingMount?.weapon_id ?? '')
  const [mountType, setMountType] = useState<ShipWeaponMountType>(() => editingMount?.mount_type ?? 'standard')
  const [modifierIDs, setModifierIDs] = useState<ShipWeaponModifierID[]>(() => editingMount?.modifier_ids ? [...editingMount.modifier_ids] : [])
  const [ammo, setAmmo] = useState<number | undefined>(() => editingMount?.ammo)
  const [firingArcID, setFiringArcID] = useState<ShipWeaponFiringArcID | undefined>(() => editingMount?.firing_arc_id)
  const quantity = editingMount?.count ?? 1
  const [preview, setPreview] = useState<MilitaryDesignDraftProjection | null>(null)
  const [previewing, setPreviewing] = useState(false)
  const [previewError, setPreviewError] = useState('')
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [helpTopic, setHelpTopic] = useState<{ kind: 'mount' | 'modifier'; id: string } | null>(null)

  const selectedWeapon = weapons.find((weapon) => weapon.id === selectedWeaponID)
  const targetSlot = editingMount?.slot ?? nextFreeSlot
  const activeMountHelp = helpTopic?.kind === 'mount' ? selectedWeapon?.mount_choices.find((choice) => choice.id === helpTopic.id) : undefined
  const activeModifierHelp = helpTopic?.kind === 'modifier' ? selectedWeapon?.modifier_choices.find((choice) => choice.id === helpTopic.id) : undefined
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
    if (!selectedWeapon || targetSlot == null) return null
    const arcID = selectedWeapon.fixed_firing_arc_id ?? firingArcID ?? selectedWeapon.default_firing_arc_id
    return {
      slot: targetSlot,
      weapon_id: selectedWeapon.id,
      count: quantity,
      mount_type: mountType,
      ...(modifierIDs.length > 0 ? { modifier_ids: [...modifierIDs] } : {}),
      ...(ammo != null && ammo > 0 ? { ammo } : {}),
      ...(arcID ? { firing_arc_id: arcID } : {}),
    }
  }, [selectedWeapon, targetSlot, quantity, mountType, modifierIDs, ammo, firingArcID])

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
    const draftWeapons = [...cloneMounts(existingMounts.filter((mount) => mount.slot !== targetSlot)), candidateMount].sort((a, b) => a.slot - b.slot)
    void previewMilitaryDesign(snapshot, snapshot.view.seat.seat.id, {
      hull_id: hullID,
      strategic_picture_id: strategicPictureID,
      computer_id: computerID,
      shield_id: shieldID,
      ...(specials.length > 0 ? { specials: specials.map((mount) => ({ ...mount })) } : {}),
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
    candidateMount?.ammo,
    candidateMount?.firing_arc_id,
    selectedWeapon?.available,
    existingMounts,
    hullID,
    strategicPictureID,
  shipGenome,
    computerID,
    shieldID,
    snapshot,
  ])

  function selectWeapon(id: string) {
    if (id === selectedWeaponID) return
    const nextWeapon = weapons.find((weapon) => weapon.id === id)
    setSelectedWeaponID(id)
    setMountType('standard')
    setModifierIDs([])
    setAmmo(nextWeapon?.default_ammo)
    setHelpTopic(null)
    setFiringArcID(nextWeapon?.fixed_firing_arc_id ?? nextWeapon?.default_firing_arc_id)
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
    && targetSlot != null
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
      ...(candidatePreview.ammo ? { ammo: candidatePreview.ammo } : {}),
      ...(candidatePreview.firing_arc_id ? { firing_arc_id: candidatePreview.firing_arc_id } : {}),
    }, preview)
  }

  const modifierFilterOptions: ModifierFilter[] = [
    'all', 'heavy', 'point_defense',
    'armor_piercing', 'continuous', 'no_range_dissipation', 'shield_piercing', 'autofire', 'enveloping',
    'mirv', 'eccm', 'armored', 'fast', 'emissions_guidance', 'overloaded',
  ]

  return (
    <div className="shipdesigner-system-dialog-backdrop shipdesigner-weapon-selector-backdrop" role="presentation" onPointerDown={(event) => { if (event.target === event.currentTarget) onClose() }}>
      <section className={"shipdesigner-system-dialog shipdesigner-weapon-selector" + (filtersOpen ? " filters-open" : "")} role="dialog" aria-modal="true" aria-labelledby="shipdesigner-weapon-selector-title">
        <header className="shipdesigner-system-dialog-header">
          <div>
            <p className="eyebrow">{t('shipbuilder.weapons')}</p>
            <h3 id="shipdesigner-weapon-selector-title">{editingMount ? t('shipbuilder.editWeapon') : t('shipbuilder.weaponSelectorTitle')}</h3>
            <p className="muted">{t('shipbuilder.weaponSelectorHint')}</p>
          </div>
          <div className="shipdesigner-weapon-selector-header-actions">
            <button
              type="button"
              className="button-ghost shipdesigner-filter-toggle"
              aria-expanded={filtersOpen}
              onClick={() => setFiltersOpen((current) => !current)}
            >{t('shipbuilder.filters')}</button>
            <button type="button" className="button-ghost shipdesigner-system-dialog-close" onClick={onClose} aria-label={t('common.close')}><GameIcon name="close" /></button>
          </div>
        </header>

        {filtersOpen && (
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
            {sortDirection === 'asc' ? '\u2191' : '\u2193'}
          </button>
        </div>

        )}
        <div className={"shipdesigner-weapon-selector-body" + (selectedWeapon ? " has-selection" : "")}>
          <div className="shipdesigner-weapon-selector-list" role="listbox" aria-label={t('shipbuilder.weaponSelectorTitle')}>
            {filteredWeapons.map((weapon) => {
              const selected = weapon.id === selectedWeaponID
              const insufficientSpace = weapon.effective_standard_space > availableSpace
              return (
                <button
                  type="button"
                  role="option"
                  aria-selected={selected}
                  data-insufficient-space={insufficientSpace ? 'true' : undefined}
                  className={`shipdesigner-weapon-selector-row${selected ? ' selected' : ''}${weapon.available ? '' : ' locked'}`}
                  key={weapon.id}
                  onClick={() => selectWeapon(weapon.id)}
                >
                  <span className="shipdesigner-weapon-selector-name">
                    <strong className={insufficientSpace ? 'shipdesigner-selector-name-no-space' : undefined}>{weaponLabel(t, weapon)}</strong>
                    {!weapon.available && <small>{t('shipbuilder.filterLocked')}</small>}
                    {weapon.craft_per_unit ? <small>{t('shipbuilder.craftPerBay', { count: weapon.craft_per_unit })}</small> : null}
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
                <div className="shipdesigner-weapon-config-scroll">
                <div className="shipdesigner-weapon-config-title">
                  <div>
                    <p className="eyebrow">{t('shipbuilder.weaponSelectorTitle')}</p>
                    <h4>{weaponLabel(t, selectedWeapon)}</h4>
                    {selectedWeapon.craft_per_unit ? <small className="muted">{t('shipbuilder.craftPerBay', { count: selectedWeapon.craft_per_unit })}</small> : null}
                  </div>
                  <span className="shipdesigner-weapon-damage">
                    <small>{t('shipbuilder.baseDamage')}</small>
                    <strong>{damageLabel(selectedWeapon)}</strong>
                  </span>
                </div>

                <fieldset className="shipdesigner-weapon-config-group">
                  <legend>{t('shipbuilder.mountType')}</legend>
                  <div className="shipdesigner-weapon-option-grid">
                    {selectedWeapon.mount_choices.map((choice) => {
                      const selected = mountType === choice.id
                      return (
                        <div className={"shipdesigner-weapon-option" + (selected ? " selected" : "") + (choice.available ? "" : " locked")} key={choice.id}>
                          <button
                            type="button"
                            className="shipdesigner-weapon-option-choice"
                            data-weapon-mount={choice.id}
                            aria-pressed={selected}
                            disabled={!choice.available}
                            onClick={() => setMountType(choice.id)}
                          >
                            <span>{t(mountTranslationKeys[choice.id])}</span>
                          </button>
                          <button
                            type="button"
                            className="shipdesigner-weapon-option-help"
                            aria-label={t('shipbuilder.weaponOptionDetails', { option: t(mountTranslationKeys[choice.id]) })}
                            onClick={() => setHelpTopic((current) => current?.kind === 'mount' && current.id === choice.id ? null : { kind: 'mount', id: choice.id })}
                          >?</button>
                        </div>
                      )
                    })}
                  </div>
                  {activeMountHelp && (
                    <div className="shipdesigner-weapon-inline-help">
                      <div className="shipdesigner-weapon-inline-help-heading">
                        <strong>{t(mountTranslationKeys[activeMountHelp.id])}</strong>
                        <button type="button" className="button-ghost" onClick={() => setHelpTopic(null)} aria-label={t('common.close')}><GameIcon name="close" /></button>
                      </div>
                      <p>{t(mountHelpTranslationKeys[activeMountHelp.id])}</p>
                      <small>{t('shipbuilder.weaponOptionCosts', { space: percentDeltaLabel(activeMountHelp.space_percent_delta), cost: percentDeltaLabel(activeMountHelp.cost_percent_delta) })}</small>
                    </div>
                  )}
                </fieldset>
                {(selectedWeapon.firing_arc_choices?.length || selectedWeapon.fixed_firing_arc_id) ? (
                  <fieldset className="shipdesigner-weapon-config-group shipdesigner-firing-arc-group">
                    <legend>{t('shipbuilder.firingArc')}</legend>
                    <FiringArcCompass
                      choices={selectedWeapon.firing_arc_choices}
                      selectedArcID={firingArcID}
                      defaultArcID={selectedWeapon.default_firing_arc_id}
                      fixedArcID={selectedWeapon.fixed_firing_arc_id}
                      hullID={hullID}
                      shipGenome={shipGenome}
                      weaponCount={existingMounts.reduce((sum, mount) => sum + mount.count, 0) + quantity}
                      locale={locale}
                      disabled={!selectedWeapon.available}
                      fixedLabel={t('shipbuilder.fixedArc360')}
                      onChange={setFiringArcID}
                    />
                  </fieldset>
                ) : null}

                {selectedWeapon.default_ammo ? (
                  <fieldset className="shipdesigner-weapon-config-group">
                    <legend>{t('shipbuilder.ammoRack')}</legend>
                    {selectedWeapon.ammo_choices?.length ? (
                      <div className="shipdesigner-weapon-option-grid">
                        {selectedWeapon.ammo_choices.map((choice) => {
                          const selected = ammo === choice
                          return (
                            <div className={"shipdesigner-weapon-option" + (selected ? " selected" : "") + (selectedWeapon.available ? "" : " locked")} key={choice}>
                              <button
                                type="button"
                                className="shipdesigner-weapon-option-choice"
                                data-weapon-ammo={choice}
                                aria-pressed={selected}
                                disabled={!selectedWeapon.available}
                                onClick={() => setAmmo(choice)}
                              >
                                <span>{choice}</span>
                              </button>
                            </div>
                          )
                        })}
                      </div>
                    ) : (
                      <p className="muted">{t('shipbuilder.fixedAmmo')}: <strong>{selectedWeapon.default_ammo}</strong></p>
                    )}
                  </fieldset>
                ) : null}

                <fieldset className="shipdesigner-weapon-config-group">
                  <legend>{t('shipbuilder.specialModifiers')}</legend>
                  <div className="shipdesigner-weapon-option-grid">
                    {selectedWeapon.modifier_choices.length > 0 ? selectedWeapon.modifier_choices.map((choice) => {
                      const selected = modifierIDs.includes(choice.id)
                      return (
                        <div className={"shipdesigner-weapon-option" + (selected ? " selected" : "") + (choice.available ? "" : " locked")} key={choice.id}>
                          <button
                            type="button"
                            className="shipdesigner-weapon-option-choice"
                            data-weapon-modifier={choice.id}
                            aria-pressed={selected}
                            disabled={!choice.available}
                            onClick={() => toggleModifier(choice.id)}
                          >
                            <span>
                              {t(modifierTranslationKeys[choice.id])}
                              {!choice.available && <small>{t('shipbuilder.refinementRequired', { level: choice.required_mini_level })}</small>}
                            </span>
                          </button>
                          <button
                            type="button"
                            className="shipdesigner-weapon-option-help"
                            aria-label={t('shipbuilder.weaponOptionDetails', { option: t(modifierTranslationKeys[choice.id]) })}
                            onClick={() => setHelpTopic((current) => current?.kind === 'modifier' && current.id === choice.id ? null : { kind: 'modifier', id: choice.id })}
                          >?</button>
                        </div>
                      )
                    }) : <p className="muted">{t('common.none')}</p>}
                  </div>
                  {activeModifierHelp && (
                    <div className="shipdesigner-weapon-inline-help">
                      <div className="shipdesigner-weapon-inline-help-heading">
                        <strong>{t(modifierTranslationKeys[activeModifierHelp.id])}</strong>
                        <button type="button" className="button-ghost" onClick={() => setHelpTopic(null)} aria-label={t('common.close')}><GameIcon name="close" /></button>
                      </div>
                      <p>{t(modifierHelpTranslationKeys[activeModifierHelp.id])}</p>
                      <small>{t('shipbuilder.weaponOptionCosts', { space: percentDeltaLabel(activeModifierHelp.space_percent_delta), cost: percentDeltaLabel(activeModifierHelp.cost_percent_delta) })}{activeModifierHelp.required_mini_level > 0 ? ' / ' + t('shipbuilder.refinementRequired', { level: activeModifierHelp.required_mini_level }) : ''}</small>
                    </div>
                  )}
                </fieldset>

                <div className="shipdesigner-weapon-config-bottom">
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

                </div>
                <footer className="shipdesigner-system-dialog-actions shipdesigner-weapon-selector-actions">
                  <button type="button" className="button-ghost" onClick={onClose}>{t('common.close')}</button>
                  <button type="button" className="button-primary" disabled={!canInstall} onClick={install}>{editingMount ? <GameIcon name="edit" /> : <span aria-hidden="true">+</span>}{editingMount ? t('shipbuilder.applyChanges') : t('shipbuilder.install')}</button>
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
