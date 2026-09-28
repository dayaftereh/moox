import { useEffect, useRef, useState } from 'react'
import {
  type MilitaryDesignDraftProjection,
  type MilitaryDesignSpecialRowPreview,
  type MilitaryDesignerSpecialChoice,
  type ShipSpecialMount,
} from '../api'
import { useI18n, type TranslationKey, type TranslationVars } from '../i18n'
import { GameIcon } from './GameIcon'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string
type DialogMode = 'add' | 'info'

type Props = {
  choices: MilitaryDesignerSpecialChoice[]
  mounts: ShipSpecialMount[]
  previews?: MilitaryDesignSpecialRowPreview[]
  hullIndex: number
  disabled?: boolean
  error?: string
  onPreview: (mounts: ShipSpecialMount[]) => Promise<MilitaryDesignDraftProjection | undefined>
  onChange: (mounts: ShipSpecialMount[], preview?: MilitaryDesignDraftProjection) => void
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
  return value.replaceAll('_', ' ').replace(/\b\w/g, (letter) => letter.toUpperCase())
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

function nextFreeSlot(mounts: ShipSpecialMount[]): number | undefined {
  return Array.from({ length: 8 }, (_, slot) => slot).find((slot) => !mounts.some((mount) => mount.slot === slot))
}

export function SpecialSystemsPanel({
  choices,
  mounts,
  previews,
  hullIndex,
  disabled = false,
  error,
  onPreview,
  onChange,
  t,
}: Props) {
  const { locale } = useI18n()
  const [dialogMode, setDialogMode] = useState<DialogMode | null>(null)
  const [selectedSpecialID, setSelectedSpecialID] = useState('')
  const [candidatePreview, setCandidatePreview] = useState<MilitaryDesignDraftProjection>()
  const [previewing, setPreviewing] = useState(false)
  const [previewError, setPreviewError] = useState('')
  const [operationError, setOperationError] = useState('')
  const previewRequest = useRef(0)

  const normalized = [...mounts].sort((a, b) => a.slot - b.slot)
  const installedIDs = new Set(normalized.map((mount) => mount.special_id))
  const freeSlot = nextFreeSlot(normalized)
  const addChoices = choices.filter((choice) => !installedIDs.has(choice.id))
  const canAdd = !disabled && normalized.length < 8 && freeSlot != null && addChoices.length > 0
  const listChoices = dialogMode === 'add' ? addChoices : choices
  const selectedChoice = choices.find((choice) => choice.id === selectedSpecialID)
  const selectedInstalledMount = normalized.find((mount) => mount.special_id === selectedSpecialID)
  const selectedExistingPreview = selectedInstalledMount
    ? previews?.find((row) => row.slot === selectedInstalledMount.slot)
    : undefined
  const selectedCandidatePreview = dialogMode === 'add' && freeSlot != null
    ? candidatePreview?.specials?.find((row) => row.slot === freeSlot)
    : undefined
  const selectedRow = selectedCandidatePreview ?? selectedExistingPreview
  const selectedSpace = selectedRow?.unit_space ?? selectedChoice?.effective_space_by_hull?.[hullIndex]
  const selectedCost = selectedRow?.unit_cost_pp ?? selectedChoice?.effective_cost_by_hull_pp?.[hullIndex]
  const selectedSupport = selectedRow?.tactical_support ?? selectedChoice?.tactical_support
  const canInstall = Boolean(
    dialogMode === 'add'
      && selectedChoice?.available
      && freeSlot != null
      && !previewing
      && candidatePreview?.save_available
      && selectedCandidatePreview,
  )

  function specialLabel(choice: MilitaryDesignerSpecialChoice | undefined, fallback = ''): string {
    if (!choice) return fallback
    const translated = labels[choice.id]
    return translated?.[locale === 'de' ? 'de' : 'en'] ?? humanize(choice.id)
  }

  function labelForID(id: string): string {
    return specialLabel(choices.find((choice) => choice.id === id), humanize(id))
  }

  useEffect(() => {
    if (!dialogMode) return
    const onKeyDown = (event: globalThis.KeyboardEvent) => {
      if (event.key === 'Escape') closeDialog()
    }
    document.body.classList.add('modal-open')
    window.addEventListener('keydown', onKeyDown)
    return () => {
      window.removeEventListener('keydown', onKeyDown)
      document.body.classList.remove('modal-open')
    }
  }, [dialogMode])

  function closeDialog() {
    previewRequest.current += 1
    setDialogMode(null)
    setSelectedSpecialID('')
    setCandidatePreview(undefined)
    setPreviewing(false)
    setPreviewError('')
  }

  function openAdd() {
    setOperationError('')
    setSelectedSpecialID('')
    setCandidatePreview(undefined)
    setPreviewError('')
    setDialogMode('add')
  }

  function openInfo(specialID: string) {
    setOperationError('')
    setSelectedSpecialID(specialID)
    setCandidatePreview(undefined)
    setPreviewError('')
    setDialogMode('info')
  }

  async function selectSpecial(specialID: string) {
    setSelectedSpecialID(specialID)
    setPreviewError('')
    setCandidatePreview(undefined)
    if (dialogMode !== 'add' || freeSlot == null) return

    const request = ++previewRequest.current
    const next = [
      ...normalized,
      { slot: freeSlot, special_id: specialID },
    ].sort((a, b) => a.slot - b.slot)

    setPreviewing(true)
    const projection = await onPreview(next)
    if (request !== previewRequest.current) return
    setPreviewing(false)
    if (!projection) {
      setPreviewError(t('shipbuilder.specialPreviewUnavailable'))
      return
    }
    setCandidatePreview(projection)
    if (!projection.save_available) {
      setPreviewError(projection.rejection_message ?? t('shipbuilder.previewRejected'))
    }
  }

  async function removeSpecial(slot: number) {
    if (disabled) return
    setOperationError('')
    const next = normalized.filter((mount) => mount.slot !== slot)
    const projection = await onPreview(next)
    if (!projection) {
      setOperationError(t('shipbuilder.specialPreviewUnavailable'))
      return
    }
    if (!projection.save_available) {
      setOperationError(projection.rejection_message ?? t('shipbuilder.previewRejected'))
      return
    }
    onChange(projection.spec?.specials ?? next, projection)
  }

  function installSelected() {
    if (!canInstall || !selectedChoice || freeSlot == null || !candidatePreview) return
    const next = [
      ...normalized,
      { slot: freeSlot, special_id: selectedChoice.id },
    ].sort((a, b) => a.slot - b.slot)
    onChange(candidatePreview.spec?.specials ?? next, candidatePreview)
    closeDialog()
  }

  const excludedNames = selectedChoice?.excluded_special_ids?.map(labelForID) ?? []

  return (
    <>
      <section className="shipdesigner-specials-section" data-special-systems>
        <div className="shipdesigner-specials-heading">
          <div>
            <p className="eyebrow">{t('shipbuilder.installedSpecials')}</p>
            <small className="muted">{t('shipbuilder.specialSystemsHint')}</small>
          </div>
          <span className="badge">{normalized.length}/8</span>
        </div>

        <div className={"shipdesigner-specials-scroll" + (normalized.length === 0 && canAdd ? " add-target" : "")} onClick={normalized.length === 0 && canAdd ? openAdd : undefined}>
          <div className="shipdesigner-installed-info-grid">
            {normalized.map((mount) => {
              const choice = choices.find((candidate) => candidate.id === mount.special_id)
              return (
                <article
                  className="shipdesigner-info-card weapon shipdesigner-special-slot-card"
                  key={mount.slot}
                  data-special-slot={mount.slot}
                  data-special-id={mount.special_id}
                >
                  <strong>{specialLabel(choice, mount.special_id)}</strong>
                  <div className="shipdesigner-weapon-slot-controls">
                    <button
                      type="button"
                      className="button-ghost shipdesigner-weapon-edit shipdesigner-special-info"
                      aria-label={t('shipbuilder.specialDetails', { special: specialLabel(choice, mount.special_id) })}
                      title={t('shipbuilder.specialDetails', { special: specialLabel(choice, mount.special_id) })}
                      onClick={() => openInfo(mount.special_id)}
                    >?</button>
                    <button
                      type="button"
                      className="button-ghost shipdesigner-weapon-edit shipdesigner-special-remove"
                      aria-label={t('shipbuilder.removeSpecial', { special: specialLabel(choice, mount.special_id) })}
                      title={t('shipbuilder.removeSpecial', { special: specialLabel(choice, mount.special_id) })}
                      disabled={disabled}
                      onClick={() => { void removeSpecial(mount.slot) }}
                    ><GameIcon name="close" /></button>
                  </div>
                </article>
              )
            })}

            {canAdd && (
              <button
                type="button"
                className="shipdesigner-info-card weapon shipdesigner-special-slot-add"
                data-add-special
                aria-label={t('shipbuilder.addSpecial')}
                onClick={(event) => { event.stopPropagation(); openAdd() }}
              >
                <span className="shipdesigner-special-slot-add-icon" aria-hidden="true">+</span>
                <strong>{t('shipbuilder.addSpecial')}</strong>
              </button>
            )}

            {normalized.length === 0 && !canAdd && (
              <div className="shipdesigner-specials-empty">{t('shipbuilder.noInstalledSpecials')}</div>
            )}
          </div>
        </div>

        {(operationError || error) && (
          <p className="shipbuilder-save-state error shipdesigner-specials-error">
            {operationError || error}
          </p>
        )}
      </section>

      {dialogMode && (
        <div
          className="shipdesigner-system-dialog-backdrop shipdesigner-weapon-selector-backdrop"
          role="presentation"
          onPointerDown={(event) => { if (event.target === event.currentTarget) closeDialog() }}
        >
          <section
            className="shipdesigner-system-dialog shipdesigner-weapon-selector shipdesigner-special-selector"
            role="dialog"
            aria-modal="true"
            aria-labelledby="shipdesigner-special-selector-title"
          >
            <header className="shipdesigner-system-dialog-header">
              <div>
                <p className="eyebrow">{t('shipbuilder.specialSystems')}</p>
                <h3 id="shipdesigner-special-selector-title">
                  {dialogMode === 'add' ? t('shipbuilder.specialSelectorTitle') : t('shipbuilder.specialInfoTitle')}
                </h3>
                <p className="muted">{t('shipbuilder.specialSelectorHint')}</p>
              </div>
              <button type="button" className="button-ghost shipdesigner-system-dialog-close" onClick={closeDialog} aria-label={t('common.close')}>
                <GameIcon name="close" />
              </button>
            </header>

            <div className={"shipdesigner-weapon-selector-body" + (selectedChoice ? " has-selection" : "")}>
              <div className="shipdesigner-weapon-selector-list" role="listbox" aria-label={t('shipbuilder.specialSelectorTitle')}>
                {listChoices.map((choice) => {
                  const selected = choice.id === selectedSpecialID
                  const space = choice.effective_space_by_hull?.[hullIndex]
                  const cost = choice.effective_cost_by_hull_pp?.[hullIndex]
                  return (
                    <button
                      type="button"
                      role="option"
                      aria-selected={selected}
                      className={"shipdesigner-weapon-selector-row shipdesigner-special-selector-row" + (selected ? " selected" : "") + (choice.available ? "" : " locked")}
                      key={choice.id}
                      data-special-choice={choice.id}
                      onClick={() => { void selectSpecial(choice.id) }}
                    >
                      <span className="shipdesigner-weapon-selector-name">
                        <strong>{specialLabel(choice)}</strong>
                        {!choice.available && <small>{t('shipbuilder.specialLocked')}</small>}
                        <small>{supportLabel(t, choice.tactical_support)}</small>
                      </span>
                      <span>{t('shipbuilder.specialMiniLevel', { level: choice.miniaturization_level })}</span>
                      <span>{space == null ? '—' : `${space} ${t('shipbuilder.space')}`}</span>
                      <span>{cost == null ? '—' : `${cost} PP`}</span>
                    </button>
                  )
                })}
              </div>

              <aside className="shipdesigner-weapon-config">
                {selectedChoice ? (
                  <>
                    <div className="shipdesigner-weapon-config-scroll">
                      <div className="shipdesigner-weapon-config-title">
                        <div>
                          <p className="eyebrow">{t('shipbuilder.specialDetailsTitle')}</p>
                          <h4>{specialLabel(selectedChoice)}</h4>
                        </div>
                        {selectedSupport && <span className={"badge support " + selectedSupport}>{supportLabel(t, selectedSupport)}</span>}
                      </div>

                      <div className="shipdesigner-system-dialog-detail shipdesigner-special-description">
                        <strong>{t('shipbuilder.specialOriginalDescription')}</strong>
                        <p>{selectedChoice.description || t('shipbuilder.specialDescriptionUnavailable')}</p>
                      </div>

                      <fieldset className="shipdesigner-weapon-config-group">
                        <legend>{t('shipbuilder.specialCosts')}</legend>
                        <div className="shipdesigner-special-detail-stats">
                          <div><small>{t('shipbuilder.space')}</small><strong>{selectedSpace == null ? '—' : selectedSpace}</strong></div>
                          <div><small>PP</small><strong>{selectedCost == null ? '—' : selectedCost}</strong></div>
                          <div><small>{t('shipbuilder.specialMiniaturization')}</small><strong>{selectedChoice.miniaturization_level}</strong></div>
                        </div>
                      </fieldset>

                      {(selectedChoice.requires_shield || excludedNames.length > 0) && (
                        <fieldset className="shipdesigner-weapon-config-group">
                          <legend>{t('shipbuilder.specialRequirements')}</legend>
                          {selectedChoice.requires_shield && <p className="muted">{t('shipbuilder.specialRequiresShield')}</p>}
                          {excludedNames.length > 0 && <p className="muted">{t('shipbuilder.specialConflictsWith', { specials: excludedNames.join(', ') })}</p>}
                        </fieldset>
                      )}

                      {dialogMode === 'add' && (
                        <div className="shipdesigner-weapon-config-bottom">
                          <div className="shipdesigner-weapon-preview">
                            {previewing && <small>{t('shipbuilder.previewing')}</small>}
                            {!previewing && selectedCandidatePreview && (
                              <strong>{t('shipbuilder.specialCurrentTotals', { space: selectedCandidatePreview.unit_space, cost: selectedCandidatePreview.unit_cost_pp })}</strong>
                            )}
                            {!previewing && candidatePreview && <small>{candidatePreview.remaining_space} {t('shipbuilder.spaceAvailable')}</small>}
                            {previewError && <small className="error">{previewError}</small>}
                          </div>
                        </div>
                      )}
                    </div>

                    <footer className="shipdesigner-system-dialog-actions shipdesigner-weapon-selector-actions">
                      <button type="button" className="button-ghost" onClick={closeDialog}>{t('common.close')}</button>
                      {dialogMode === 'add' && (
                        <button type="button" className="button-primary" disabled={!canInstall} onClick={installSelected}>
                          <span aria-hidden="true">+</span>{t('shipbuilder.installSpecial')}
                        </button>
                      )}
                    </footer>
                  </>
                ) : (
                  <div className="shipdesigner-weapon-config-empty">{t('shipbuilder.specialSelectorHint')}</div>
                )}
              </aside>
            </div>
          </section>
        </div>
      )}
    </>
  )
}
