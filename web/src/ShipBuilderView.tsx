import { useEffect, useMemo, useState } from 'react'
import {
  decodeShipVisualGenome,
  isAPIError,
  submitMilitaryDesign,
  submitMilitaryDesignVisual,
  type MilitaryDesignerHullChoice,
  type PlayerSnapshot,
  type ShipDesign,
} from './api'
import { ProceduralShipGlyph } from './components/ProceduralShipGlyph'
import { GameIcon } from './components/GameIcon'
import { Card, PageHeader } from './components/ui'
import type { TranslationKey, TranslationVars } from './i18n'
import { createRandomShipGenome, shipHullFootprint, type ShipVisualGenome } from './shipVisualGenome'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type Props = {
  snapshot: PlayerSnapshot
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

function lockLabel(t: Translator, hull?: MilitaryDesignerHullChoice): string {
  if (!hull?.lock_reason) return ''
  if (hull.lock_reason === 'technology_required') {
    return t('shipbuilder.lockTechnology', { technology: hull.required_technology_key ?? hull.required_technology_id ?? '?' })
  }
  return t('shipbuilder.lockScope')
}

export function ShipBuilderView({ snapshot, reloadSnapshot, t }: Props) {
  const designs = snapshot.decision?.strategic.ship_designs ?? []
  const designer = snapshot.decision?.decisions.ship_designer
  const hulls = designer?.hulls ?? []
  const firstAvailableHull = hulls.find((hull) => hull.save_available) ?? hulls[0]

  const [selectedDesignID, setSelectedDesignID] = useState<number | null>(() => designs[0]?.id ?? null)
  const selectedDesign = designs.find((design) => design.id === selectedDesignID)
  const [name, setName] = useState(() => selectedDesign?.name ?? '')
  const [hullID, setHullID] = useState(() => selectedDesign?.spec.hull_id ?? firstAvailableHull?.id ?? 'frigate')
  const [laserInstalled, setLaserInstalled] = useState(() => selectedDesign?.spec.weapons?.some((mount) => mount.weapon_id === 'laser_cannon') ?? false)
  const [visualOverride, setVisualOverride] = useState<ShipVisualGenome | null>(() => selectedDesign ? designVisual(selectedDesign) : null)
  const [roll, setRoll] = useState(1)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')
  const [saveNotice, setSaveNotice] = useState('')

  useEffect(() => {
    if (!selectedDesign) return
    setName(selectedDesign.name)
    setHullID(selectedDesign.spec.hull_id)
    setLaserInstalled(selectedDesign.spec.weapons?.some((mount) => mount.weapon_id === 'laser_cannon') ?? false)
    setVisualOverride(designVisual(selectedDesign))
    setSaveError('')
  }, [selectedDesign?.id, selectedDesign?.revision, selectedDesign?.visual_revision])

  const currentHullIndex = Math.max(0, hulls.findIndex((hull) => hull.id === hullID))
  const currentHull = hulls[currentHullIndex]
  const laser = designer?.weapons.find((weapon) => weapon.id === 'laser_cannon')
  const variantKey = hullID === 'frigate' ? `frigate:${laserInstalled ? 'laser_cannon' : 'none'}` : ''
  const currentVariant = designer?.variants.find((variant) => variant.key === variantKey)
  const candidate = useMemo(() => {
    if (visualOverride && String(visualOverride.hullId) === hullID) return visualOverride
    return createRandomShipGenome(`shipbuilder:${snapshot.view.game_id}:${selectedDesignID ?? 'new'}:${hullID}:${roll}`, hullID)
  }, [hullID, roll, selectedDesignID, snapshot.view.game_id, visualOverride])

  function selectDesign(design: ShipDesign) {
    setSelectedDesignID(design.id)
    setName(design.name)
    setHullID(design.spec.hull_id)
    setLaserInstalled(design.spec.weapons?.some((mount) => mount.weapon_id === 'laser_cannon') ?? false)
    setVisualOverride(designVisual(design))
    setRoll((value) => value + 1)
    setSaveError('')
    setSaveNotice('')
  }

  function startNewDesign() {
    setSelectedDesignID(null)
    setName('')
    setHullID(firstAvailableHull?.id ?? 'frigate')
    setLaserInstalled(false)
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
    setLaserInstalled(false)
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

  const planningWritable = snapshot.view.phase === 'planning' && !snapshot.view.seat.submitted
  const canSave = Boolean(currentHull?.save_available && currentVariant && name.trim() && planningWritable && !saving)

  async function saveDesign() {
    if (!canSave || !currentHull || !currentVariant) return
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
        strategic_picture_id: selectedDesign?.spec.strategic_picture_id ?? currentHull.strategic_picture_ids[0] ?? 0,
        ...(laserInstalled ? { weapons: [{ slot: 0, weapon_id: 'laser_cannon', count: 1 }] } : {}),
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
        setLaserInstalled(finalDesign.spec.weapons?.some((mount) => mount.weapon_id === 'laser_cannon') ?? false)
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

  const spec = currentVariant?.spec
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
              <ProceduralShipGlyph seed={candidate.seed} hullId={hullID} genome={candidate} footprint={shipHullFootprint(hullID)} className="shipdesigner-preview-ship" label={name.trim() || hullLabel(t, hullID)} />
              <span className="shipdesigner-preview-overlay"><GameIcon name="generate" />{t('shipbuilder.clickToReroll')}</span>
            </button>

            <div className="shipdesigner-component-columns">
              <section className="shipdesigner-component-section">
                <div className="card-heading"><div><p className="eyebrow">{t('shipbuilder.available')}</p><h3>{t('shipbuilder.availableComponents')}</h3></div></div>
                {laser ? (
                  <div className={`shipdesigner-component-row${laser.available ? '' : ' locked'}`}>
                    <div><strong>{t('shipbuilder.laserCannon')}</strong><small>{laser.base_space} {t('shipbuilder.space')} · {laser.base_cost_pp} PP</small></div>
                    <button type="button" className="button-secondary" disabled={!laser.available || laserInstalled || !currentHull?.save_available} onClick={() => { setLaserInstalled(true); setSaveNotice('') }}>{laser.available ? t('shipbuilder.add') : t('shipbuilder.locked')}</button>
                  </div>
                ) : <p className="muted">{t('shipbuilder.noAvailableComponents')}</p>}
              </section>

              <section className="shipdesigner-component-section">
                <div className="card-heading"><div><p className="eyebrow">{t('shipbuilder.installed')}</p><h3>{t('shipbuilder.installedComponents')}</h3></div></div>
                {spec ? (
                  <div className="shipdesigner-installed-list">
                    <div className="shipdesigner-installed-row"><span>{t('shipbuilder.drive')}</span><strong>{spec.warp_drive_id}</strong></div>
                    <div className="shipdesigner-installed-row"><span>{t('shipbuilder.computer')}</span><strong>{spec.computer_id}</strong></div>
                    <div className="shipdesigner-installed-row"><span>{t('shipbuilder.armor')}</span><strong>{spec.armor_id}</strong></div>
                    {spec.shield_id && <div className="shipdesigner-installed-row"><span>{t('shipbuilder.shield')}</span><strong>{spec.shield_id}</strong></div>}
                    <div className="shipdesigner-installed-row"><span>{t('shipbuilder.fuel')}</span><strong>{spec.fuel_cell_id}</strong></div>
                    {laserInstalled && (
                      <div className="shipdesigner-installed-row weapon">
                        <span>{t('shipbuilder.slot', { slot: 1 })}</span>
                        <strong>1× {t('shipbuilder.laserCannon')}</strong>
                        <button type="button" className="button-ghost" onClick={() => { setLaserInstalled(false); setSaveNotice('') }}>{t('shipbuilder.remove')}</button>
                      </div>
                    )}
                  </div>
                ) : <p className="muted">{hullLock || t('shipbuilder.noAuthoritativePreview')}</p>}
                <p className="muted shipdesigner-future-note">{t('shipbuilder.futureMounts')}</p>
              </section>
            </div>
          </Card>

          <Card className="shipdesigner-summary-card">
            <div className="shipdesigner-summary-grid">
              <div><span className="eyebrow">{t('shipbuilder.productionCost')}</span><strong>{spec ? `${spec.production_cost_pp} PP` : '—'}</strong></div>
              <div><span className="eyebrow">{t('shipbuilder.commandPoints')}</span><strong>{currentVariant ? `${currentVariant.command_point_cost} CP` : currentHull ? `${currentHull.command_point_cost} CP` : '—'}</strong></div>
              <div><span className="eyebrow">{t('shipbuilder.designSpace')}</span><strong>{spec ? `${spec.space_used} / ${spec.hull_space}` : `— / ${currentHull?.base_space ?? '—'}`}</strong></div>
            </div>
            <div className="shipdesigner-save-row">
              <div>
                {saveError && <p className="shipbuilder-save-state error">{saveError}</p>}
                {!saveError && saveNotice && <p className="shipbuilder-save-state">{saveNotice}</p>}
                {!planningWritable && <p className="muted">{t('shipbuilder.savePlanningOnly')}</p>}
              </div>
              <button type="button" className="button-primary" disabled={!canSave} onClick={() => void saveDesign()}><GameIcon name="check" />{saving ? t('shipbuilder.saving') : t('shipbuilder.saveDesign')}</button>
            </div>
          </Card>
        </div>
      </div>
    </>
  )
}
