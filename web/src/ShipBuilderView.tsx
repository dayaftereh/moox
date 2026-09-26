import { useEffect, useMemo, useState } from 'react'
import {
  decodeShipVisualGenome,
  isAPIError,
  previewMilitaryDesign,
  submitMilitaryDesign,
  submitMilitaryDesignVisual,
  type MilitaryDesignDraftProjection,
  type MilitaryDesignerHullChoice,
  type PlayerSnapshot,
  type ShipDesign,
  type ShipWeaponMount,
} from './api'
import { ProceduralShipGlyph } from './components/ProceduralShipGlyph'
import { GameIcon } from './components/GameIcon'
import { ShipSystemInfoCard } from './components/ShipSystemInfoCard'
import { WeaponSelectorDialog } from './components/WeaponSelectorDialog'
import { Card, PageHeader } from './components/ui'
import type { TranslationKey, TranslationVars } from './i18n'
import { createRandomShipGenome, shipHullDesignerFootprint, shipHullFootprint, type ShipVisualGenome } from './shipVisualGenome'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type Props = {
  snapshot: PlayerSnapshot
  initialDesignID?: number
  reloadSnapshot: () => Promise<PlayerSnapshot | undefined>
  t: Translator
}

const hullTranslationKeys: Record<string, TranslationKey> = {
  frigate: 'shipbuilder.hull.frigate',
  destroyer: 'shipbuilder.hull.destroyer',
  cruiser: 'shipbuilder.hull.cruiser',
  battleship: 'shipbuilder.hull.battleship',
  titan: 'shipbuilder.hull.titan',
  doom_star: 'shipbuilder.hull.doomStar',
}

function hullLabel(t: Translator, hullID: string): string {
  return t(hullTranslationKeys[hullID] ?? 'shipbuilder.hull.frigate')
}

function designVisual(design: ShipDesign): ShipVisualGenome {
  return decodeShipVisualGenome(design.visual_genome)
    ?? createRandomShipGenome(`shipbuilder:catalog:${design.id}:${design.revision}`, design.spec.hull_id)

}
function cloneWeaponMounts(weapons?: ShipWeaponMount[]): ShipWeaponMount[] {
  return (weapons ?? [])
    .map((mount) => ({ ...mount, modifier_ids: mount.modifier_ids ? [...mount.modifier_ids] : undefined }))
    .sort((a, b) => a.slot - b.slot)
}

function lockLabel(t: Translator, hull?: MilitaryDesignerHullChoice): string {
  if (!hull?.lock_reason) return ''
  if (hull.lock_reason === 'technology_required') {
    return t('shipbuilder.lockTechnology', { technology: hull.required_technology_key ?? hull.required_technology_id ?? '?' })
  }
  return t('shipbuilder.lockScope')
}

export function ShipBuilderView({ snapshot, initialDesignID, reloadSnapshot, t }: Props) {
  const designs = snapshot.decision?.strategic.ship_designs ?? []
  const designer = snapshot.decision?.decisions.ship_designer
  const hulls = designer?.hulls ?? []
  const computers = designer?.computers ?? []
  const shields = designer?.shields ?? []
  const firstAvailableHull = hulls.find((hull) => hull.save_available) ?? hulls[0]
  const bestAvailableComputer = [...computers].reverse().find((choice) => choice.available)
  const bestAvailableShield = [...shields].reverse().find((choice) => choice.available)

  const [selectedDesignID, setSelectedDesignID] = useState<number | null>(() => designs.find((design) => design.id === initialDesignID)?.id ?? designs[0]?.id ?? null)
  const selectedDesign = designs.find((design) => design.id === selectedDesignID)
  const [name, setName] = useState(() => selectedDesign?.name ?? '')
  const [hullID, setHullID] = useState(() => selectedDesign?.spec.hull_id ?? firstAvailableHull?.id ?? 'frigate')
  const [computerID, setComputerID] = useState(() => selectedDesign?.spec.computer_id ?? bestAvailableComputer?.id ?? '')
  const [shieldID, setShieldID] = useState(() => selectedDesign ? (selectedDesign.spec.shield_id ?? '') : (bestAvailableShield?.id ?? ''))
  const [weaponMounts, setWeaponMounts] = useState<ShipWeaponMount[]>(() => cloneWeaponMounts(selectedDesign?.spec.weapons))
  const [visualOverride, setVisualOverride] = useState<ShipVisualGenome | null>(() => selectedDesign ? designVisual(selectedDesign) : null)
  const [roll, setRoll] = useState(1)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')
  const [saveNotice, setSaveNotice] = useState('')
  const [designPreview, setDesignPreview] = useState<MilitaryDesignDraftProjection | null>(null)
  const [designPreviewError, setDesignPreviewError] = useState('')
  const [weaponSelectorOpen, setWeaponSelectorOpen] = useState(false)
  const [weaponEditorSlot, setWeaponEditorSlot] = useState<number | null>(null)
  const planningWritable = snapshot.view.phase === 'planning' && !snapshot.view.seat.submitted

  useEffect(() => {
    if (!selectedDesign) return
    setName(selectedDesign.name)
    setHullID(selectedDesign.spec.hull_id)
    setComputerID(selectedDesign.spec.computer_id)
    setShieldID(selectedDesign.spec.shield_id ?? '')
    setWeaponMounts(cloneWeaponMounts(selectedDesign.spec.weapons))
    setVisualOverride(designVisual(selectedDesign))
    setSaveError('')
  }, [selectedDesign?.id, selectedDesign?.revision, selectedDesign?.visual_revision])

  const currentHullIndex = Math.max(0, hulls.findIndex((hull) => hull.id === hullID))
  const currentHull = hulls[currentHullIndex]
  const selectedComputer = computers.find((choice) => choice.id === computerID && choice.available)
  const selectedShield = shieldID ? shields.find((choice) => choice.id === shieldID && choice.available) : undefined
  const strategicPictureID = selectedDesign?.spec.hull_id === currentHull?.id
    ? selectedDesign.spec.strategic_picture_id
    : (currentHull?.strategic_picture_ids[0] ?? 0)
  const nextFreeWeaponSlot = Array.from({ length: 8 }, (_, slot) => slot).find((slot) => !weaponMounts.some((mount) => mount.slot === slot))
  const editingWeaponMount = weaponEditorSlot == null ? undefined : weaponMounts.find((mount) => mount.slot === weaponEditorSlot)
  const spec = planningWritable ? designPreview?.spec : selectedDesign?.spec
  const remainingSpace = designPreview?.remaining_space ?? (spec ? spec.hull_space - spec.space_used : 0)
  const unusedSpaceSpeedBonus = spec && currentHull
    ? Math.ceil(((designer?.max_unused_space_combat_speed_bonus ?? 0) * Math.max(0, remainingSpace)) / Math.max(1, spec.hull_space))
    : 0
  const combatSpeed = currentHull ? currentHull.base_combat_speed + (designer?.automatic_systems.tactical_speed_bonus ?? 0) + unusedSpaceSpeedBonus : 0
  const beamDefense = combatSpeed * (designer?.beam_defense_per_combat_speed ?? 0)
  const missileEvasion = designer?.base_missile_evasion ?? 0
  const canOpenWeaponSelector = Boolean(planningWritable && currentHull?.save_available && nextFreeWeaponSlot != null && designer?.weapons.length)
  useEffect(() => {
    if (!planningWritable || !currentHull?.save_available || !selectedComputer || (shieldID && !selectedShield)) {
      setDesignPreview(null)
      setDesignPreviewError('')
      return
    }
    let cancelled = false
    const timer = window.setTimeout(() => {
      void previewMilitaryDesign(snapshot, snapshot.view.seat.seat.id, {
        hull_id: currentHull.id,
        strategic_picture_id: strategicPictureID,
        computer_id: selectedComputer.id,
        shield_id: shieldID,
        ...(weaponMounts.length > 0 ? { weapons: cloneWeaponMounts(weaponMounts) } : {}),
      }).then((result) => {
        if (cancelled) return
        setDesignPreview(result.preview.projection)
        setDesignPreviewError('')
      }).catch((error: unknown) => {
        if (cancelled) return
        setDesignPreview(null)
        setDesignPreviewError(error instanceof Error ? error.message : String(error))
      })
    }, 60)
    return () => {
      cancelled = true
      window.clearTimeout(timer)
    }
  }, [
    planningWritable,
    snapshot.view.revision,
    currentHull?.id,
    currentHull?.save_available,
    strategicPictureID,
    selectedComputer?.id,
    shieldID,
    selectedShield?.id,
    weaponMounts,
  ])
  const candidate = useMemo(() => {
    if (visualOverride && String(visualOverride.hullId) === hullID) return visualOverride
    return createRandomShipGenome(`shipbuilder:${snapshot.view.game_id}:${selectedDesignID ?? 'new'}:${hullID}:${roll}`, hullID)
  }, [hullID, roll, selectedDesignID, snapshot.view.game_id, visualOverride])

  function selectDesign(design: ShipDesign) {
    setSelectedDesignID(design.id)
    setName(design.name)
    setHullID(design.spec.hull_id)
    setComputerID(design.spec.computer_id)
    setShieldID(design.spec.shield_id ?? '')
    setWeaponMounts(cloneWeaponMounts(design.spec.weapons))
    setVisualOverride(designVisual(design))
    setRoll((value) => value + 1)
    setSaveError('')
    setSaveNotice('')
  }

  function startNewDesign() {
    setSelectedDesignID(null)
    setName('')
    setHullID(firstAvailableHull?.id ?? 'frigate')
    setComputerID(bestAvailableComputer?.id ?? '')
    setShieldID(bestAvailableShield?.id ?? '')
    setWeaponMounts([])
    setVisualOverride(null)
    setRoll((value) => value + 1)
    setSaveError('')
    setSaveNotice('')
  }

  function stepHull(delta: number) {
    if (hulls.length === 0) return
    const nextIndex = Math.min(hulls.length - 1, Math.max(0, currentHullIndex + delta))
    const next = hulls[nextIndex]
    if (!next || next.id === hullID) return
    setHullID(next.id)
    setWeaponMounts([])
    setVisualOverride(null)
    setRoll((value) => value + 1)
    setSaveError('')
    setSaveNotice('')
  }

  function rerollVisual() {
    setVisualOverride(null)
    setRoll((value) => value + 1)
    setSaveNotice('')
  }

  async function previewWeaponMounts(mounts: ShipWeaponMount[]): Promise<MilitaryDesignDraftProjection | undefined> {
    if (!currentHull?.save_available || !selectedComputer) return undefined
    try {
      const result = await previewMilitaryDesign(snapshot, snapshot.view.seat.seat.id, {
        hull_id: currentHull.id,
        strategic_picture_id: strategicPictureID,
        computer_id: selectedComputer.id,
        shield_id: shieldID,
        ...(mounts.length > 0 ? { weapons: cloneWeaponMounts(mounts) } : {}),
      })
      return result.preview.projection
    } catch (error: unknown) {
      setDesignPreviewError(error instanceof Error ? error.message : String(error))
      return undefined
    }
  }
  function openAddWeapon() {
    setWeaponEditorSlot(null)
    setWeaponSelectorOpen(true)
  }

  function openEditWeapon(slot: number) {
    setWeaponEditorSlot(slot)
    setWeaponSelectorOpen(true)
  }

  function closeWeaponSelector() {
    setWeaponSelectorOpen(false)
    setWeaponEditorSlot(null)
  }

  function installWeaponMount(mount: ShipWeaponMount, preview: MilitaryDesignDraftProjection) {
    setWeaponMounts((current) => [
      ...cloneWeaponMounts(current).filter((candidate) => candidate.slot !== mount.slot),
      { ...mount, modifier_ids: mount.modifier_ids ? [...mount.modifier_ids] : undefined },
    ].sort((a, b) => a.slot - b.slot))
    setDesignPreview(preview)
    setDesignPreviewError('')
    closeWeaponSelector()
    setSaveNotice('')
  }

  function changeWeaponMountCount(slot: number, delta: number) {
    if (delta === 0) return
    const mount = weaponMounts.find((candidate) => candidate.slot === slot)
    if (!mount) return
    const nextCount = mount.count + delta
    const next = nextCount <= 0
      ? weaponMounts.filter((candidate) => candidate.slot !== slot)
      : weaponMounts.map((candidate) => candidate.slot === slot ? { ...candidate, count: nextCount } : candidate)
    void previewWeaponMounts(next).then((projection) => {
      if (!projection) return
      if (!projection.save_available) {
        setDesignPreviewError(projection.rejection_message ?? t('shipbuilder.previewRejected'))
        return
      }
      setWeaponMounts(cloneWeaponMounts(next))
      setDesignPreview(projection)
      setDesignPreviewError('')
      setSaveNotice('')
    })
  }
  const canSave = Boolean(currentHull?.save_available && selectedComputer?.available && (!shieldID || selectedShield?.available) && spec && designPreview?.save_available && name.trim() && planningWritable && !saving)

  async function saveDesign() {
    if (!canSave || !currentHull || !spec) return
    setSaving(true)
    setSaveError('')
    setSaveNotice('')
    const beforeIDs = new Set(designs.map((design) => design.id))
    let gameplaySaved = false
    try {
      await submitMilitaryDesign(snapshot, snapshot.view.seat.seat.id, {
        ...(selectedDesignID ? { design_id: selectedDesignID } : {}),
        name: name.trim(),
        hull_id: currentHull.id,
        strategic_picture_id: strategicPictureID,
        computer_id: selectedComputer?.id ?? '',
        shield_id: shieldID,
        ...(weaponMounts.length > 0 ? { weapons: cloneWeaponMounts(weaponMounts) } : {}),
      })
      gameplaySaved = true

      const afterDesign = await reloadSnapshot()
      const refreshedDesigns = afterDesign?.decision?.strategic.ship_designs ?? []
      const savedID = selectedDesignID
        ?? refreshedDesigns.find((design) => !beforeIDs.has(design.id))?.id
        ?? [...refreshedDesigns].sort((a, b) => b.id - a.id)[0]?.id
      if (!afterDesign || !savedID) throw new Error(t('shipbuilder.savedDesignMissing'))

      setSelectedDesignID(savedID)
      await submitMilitaryDesignVisual(afterDesign, afterDesign.view.seat.seat.id, savedID, candidate)
      const finalSnapshot = await reloadSnapshot()
      const finalDesign = finalSnapshot?.decision?.strategic.ship_designs?.find((design) => design.id === savedID)
      if (finalDesign) {
        setVisualOverride(designVisual(finalDesign))
        setName(finalDesign.name)
        setHullID(finalDesign.spec.hull_id)
        setComputerID(finalDesign.spec.computer_id)
        setShieldID(finalDesign.spec.shield_id ?? '')
        setWeaponMounts(cloneWeaponMounts(finalDesign.spec.weapons))
        setSaveNotice(t('shipbuilder.designSaved', { revision: finalDesign.revision, visualRevision: finalDesign.visual_revision ?? 0 }))
      } else {
        setSaveNotice(t('shipbuilder.designSavedSimple'))
      }
    } catch (cause) {
      if (isAPIError(cause) && cause.status === 409) {
        try { await reloadSnapshot() } catch { /* keep the original rejection visible */ }
      }
      const message = cause instanceof Error ? cause.message : String(cause)
      setSaveError(gameplaySaved ? `${t('shipbuilder.visualSaveFailed')} ${message}` : message)
    } finally {
      setSaving(false)
    }
  }

  if (!designer || hulls.length === 0) {
    return (
      <>
        <PageHeader eyebrow={t('shipbuilder.eyebrow')} title={t('shipbuilder.title')} subtitle={t('shipbuilder.designerSubtitle')} />
        <Card><p className="muted">{t('shipbuilder.designerUnavailable')}</p></Card>
      </>
    )
  }

  const hullLock = lockLabel(t, currentHull)

  return (
    <>
      <PageHeader eyebrow={t('shipbuilder.eyebrow')} title={t('shipbuilder.title')} subtitle={t('shipbuilder.designerSubtitle')} />

      <div className="shipdesigner-layout">
        <Card className="shipdesigner-catalog-card">
          <div className="card-heading">
            <div><p className="eyebrow">{t('shipbuilder.designLibrary')}</p><h2>{t('shipbuilder.designs')}</h2></div>
            <button type="button" className="button-secondary shipdesigner-new" onClick={startNewDesign}>{t('shipbuilder.newDesign')}</button>
          </div>
          <p className="muted">{t('shipbuilder.designLibraryHint')}</p>
          <div className="shipdesigner-design-list" role="list">
            {designs.map((design) => {
              const genome = designVisual(design)
              return (
                <button
                  type="button"
                  key={design.id}
                  className={`shipdesigner-design-item${selectedDesignID === design.id ? ' selected' : ''}`}
                  onClick={() => selectDesign(design)}
                  aria-pressed={selectedDesignID === design.id}
                >
                  <ProceduralShipGlyph seed={genome.seed} hullId={design.spec.hull_id} genome={genome} footprint={shipHullFootprint(design.spec.hull_id)} className="shipdesigner-design-thumb" />
                  <span><strong>{design.name}</strong><small>{hullLabel(t, design.spec.hull_id)} · r{design.revision}</small></span>
                </button>
              )
            })}
            {designs.length === 0 && <p className="muted">{t('shipbuilder.noDesigns')}</p>}
          </div>
        </Card>

        <div className="shipdesigner-editor-stack">
          <Card className="shipdesigner-editor-card">
            <div className="shipdesigner-name-row">
              <label htmlFor="ship-design-name"><span className="eyebrow">{t('shipbuilder.designName')}</span></label>
              <input id="ship-design-name" value={name} onChange={(event) => { setName(event.target.value); setSaveNotice('') }} maxLength={80} placeholder={t('shipbuilder.designNamePlaceholder')} />
            </div>

            <div className="shipdesigner-hull-section">
              <p className="eyebrow">{t('shipbuilder.hullSize')}</p>
              <div className="shipdesigner-hull-stepper">
                <button type="button" className="shipdesigner-arrow" onClick={() => stepHull(-1)} disabled={currentHullIndex <= 0} aria-label={t('shipbuilder.previousHull')}>‹</button>
                <div className={`shipdesigner-hull-current${currentHull?.save_available ? '' : ' locked'}`}>
                  <strong>{hullLabel(t, currentHull?.id ?? hullID)}</strong>
                  <small>{currentHull?.command_point_cost ?? '—'} CP · {currentHull?.base_space ?? '—'} {t('shipbuilder.space')}</small>
                  {hullLock && <span className="badge warning">{hullLock}</span>}
                </div>
                <button type="button" className="shipdesigner-arrow" onClick={() => stepHull(1)} disabled={currentHullIndex >= hulls.length - 1} aria-label={t('shipbuilder.nextHull')}>›</button>
              </div>
            </div>

            <button type="button" className="shipdesigner-preview" onClick={rerollVisual} title={t('shipbuilder.clickToReroll')}>
              <ProceduralShipGlyph seed={candidate.seed} hullId={hullID} genome={candidate} footprint={shipHullDesignerFootprint(hullID)} className="shipdesigner-preview-ship" label={name.trim() || hullLabel(t, hullID)} normalizeVisualFootprint />
              <span className="shipdesigner-preview-overlay"><GameIcon name="generate" />{t('shipbuilder.clickToReroll')}</span>
            </button>

            {spec ? (
              <ShipSystemInfoCard
                designer={designer}
                spec={spec}
                currentHull={currentHull}
                computerID={computerID}
                shieldID={shieldID}
                weaponMounts={weaponMounts}
                weapons={designer.weapons}
                weaponPreviews={designPreview?.weapons}
                remainingSpace={remainingSpace}
                combatSpeed={combatSpeed}
                unusedSpaceSpeedBonus={unusedSpaceSpeedBonus}
                beamDefense={beamDefense}
                missileEvasion={missileEvasion}
                onComputerChange={(id) => { setComputerID(id); setSaveNotice('') }}
                onShieldChange={(id) => { setShieldID(id); setSaveNotice('') }}
                canAddWeapon={canOpenWeaponSelector}
                weaponError={designPreviewError}
                onAddWeapon={openAddWeapon}
                onEditWeapon={openEditWeapon}
                onWeaponCountChange={changeWeaponMountCount}
                t={t}
              />
            ) : (
              <section className="shipdesigner-component-section shipdesigner-ship-info-card">
                <p className="muted">{hullLock || t('shipbuilder.noAuthoritativePreview')}</p>
              </section>
            )}
          </Card>

          <Card className="shipdesigner-summary-card">
            <div className="shipdesigner-save-row">
              <div className="shipdesigner-save-details">
                <div className="shipdesigner-design-cost">
                  <span>{t('shipbuilder.productionCost')}</span>
                  <strong>{spec ? `${spec.production_cost_pp} PP` : '—'}</strong>
                </div>
                <div className="shipdesigner-save-messages">
                {saveError && <p className="shipbuilder-save-state error">{saveError}</p>}
                {!saveError && saveNotice && <p className="shipbuilder-save-state">{saveNotice}</p>}
                {!planningWritable && <p className="muted">{t('shipbuilder.savePlanningOnly')}</p>}
                </div>
              </div>
              <button type="button" className="button-primary" disabled={!canSave} onClick={() => void saveDesign()}><GameIcon name="check" />{saving ? t('shipbuilder.saving') : t('shipbuilder.saveDesign')}</button>
            </div>
          </Card>
        </div>
      </div>
      {weaponSelectorOpen && designer && currentHull && selectedComputer && (
        <WeaponSelectorDialog
          snapshot={snapshot}
          weapons={designer.weapons}
          existingMounts={weaponMounts}
          hullID={currentHull.id}
          strategicPictureID={strategicPictureID}
          shipGenome={candidate}
          computerID={selectedComputer.id}
          shieldID={shieldID}
          nextFreeSlot={editingWeaponMount?.slot ?? nextFreeWeaponSlot}
          editingMount={editingWeaponMount}
          onInstall={installWeaponMount}
          onClose={closeWeaponSelector}
          t={t}
        />
      )}
    </>
  )
}
