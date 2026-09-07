import { useMemo, useState } from 'react'
import type { PlayerSnapshot } from './api'
import { ProceduralShipGlyph } from './components/ProceduralShipGlyph'
import { Card, PageHeader } from './components/ui'
import type { TranslationKey, TranslationVars } from './i18n'
import { createRandomShipGenome, createShipGenome, shipHullFootprint, shipHullSpace, type ShipMorphologyID, type ShipStyleID, type ShipVisualGenome } from './shipVisualGenome'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type HullID = 'scout' | 'frigate' | 'destroyer' | 'cruiser' | 'battleship' | 'titan' | 'doom_star'

const hulls: Array<{ id: HullID; label: TranslationKey; status: TranslationKey }> = [
  { id: 'scout', label: 'shipbuilder.hull.scout', status: 'shipbuilder.scoutRole' },
  { id: 'frigate', label: 'shipbuilder.hull.frigate', status: 'shipbuilder.authoritative' },
  { id: 'destroyer', label: 'shipbuilder.hull.destroyer', status: 'shipbuilder.visualPreview' },
  { id: 'cruiser', label: 'shipbuilder.hull.cruiser', status: 'shipbuilder.visualPreview' },
  { id: 'battleship', label: 'shipbuilder.hull.battleship', status: 'shipbuilder.visualPreview' },
  { id: 'titan', label: 'shipbuilder.hull.titan', status: 'shipbuilder.visualPreview' },
  { id: 'doom_star', label: 'shipbuilder.hull.doomStar', status: 'shipbuilder.visualPreview' },
]

const morphologyLabels: Record<ShipMorphologyID, TranslationKey> = {
  needle: 'shipbuilder.morphology.needle',
  barge: 'shipbuilder.morphology.barge',
  manta: 'shipbuilder.morphology.manta',
  fork: 'shipbuilder.morphology.fork',
  chevron: 'shipbuilder.morphology.chevron',
  hammer: 'shipbuilder.morphology.hammer',
  bulb: 'shipbuilder.morphology.bulb',
}

const styleLabels: Record<ShipStyleID, TranslationKey> = {
  spear: 'shipbuilder.style.spear',
  sleek: 'shipbuilder.style.sleek',
  organic: 'shipbuilder.style.organic',
}

function hullLabel(t: Translator, hullID: HullID): string {
  return t(hulls.find((hull) => hull.id === hullID)?.label ?? 'shipbuilder.hull.frigate')
}

export function ShipBuilderView({ snapshot, t }: { snapshot: PlayerSnapshot; t: Translator }) {
  const [hullID, setHullID] = useState<HullID>('scout')
  const [roll, setRoll] = useState(1)
  const [kept, setKept] = useState<ShipVisualGenome | null>(null)
  const currentHull = hulls.find((hull) => hull.id === hullID) ?? hulls[0]
  const designs = snapshot.decision?.strategic.ship_designs ?? []
  const baseline = designs[0]
  const baselineWeapons = baseline?.spec.weapons ?? []

  const candidate = useMemo(() => {
    const seed = `shipbuilder:${snapshot.view.game_id}:${hullID}:full-random:${roll}`
    return createRandomShipGenome(seed, hullID)
  }, [hullID, roll, snapshot.view.game_id])

  function reroll() {
    setRoll((value) => value + 1)
  }

  function selectHull(nextHullID: HullID) {
    if (nextHullID === hullID) return
    setHullID(nextHullID)
    setRoll((value) => value + 1)
    setKept(null)
  }

  return (
    <>
      <PageHeader eyebrow={t('shipbuilder.eyebrow')} title={t('shipbuilder.title')} subtitle={t('shipbuilder.randomSubtitle')} />

      <Card className="shipbuilder-size-card">
        <div className="card-heading">
          <div><p className="eyebrow">{t('shipbuilder.stepSize')}</p><h2>{t('shipbuilder.size')}</h2></div>
          <span className="badge">{t(currentHull.status)}</span>
        </div>
        <div className="shipbuilder-hull-grid" role="list" aria-label={t('shipbuilder.size')}>
          {hulls.map((hull) => (
            <button
              type="button"
              className={`shipbuilder-hull-option${hull.id === hullID ? ' selected' : ''}`}
              key={hull.id}
              aria-pressed={hull.id === hullID}
              onClick={() => selectHull(hull.id)}
            >
              <ProceduralShipGlyph seed={`shipbuilder:hull:${hull.id}`} hullId={hull.id} genome={createShipGenome(`shipbuilder:hull:${hull.id}`, hull.id, 'spear', 'manta')} footprint={shipHullFootprint(hull.id)} className="shipbuilder-hull-thumb" />
              <span><strong>{t(hull.label)}</strong><small>{t(hull.status)}</small><small>{t('shipbuilder.spaceValue', { space: shipHullSpace(hull.id) })}</small></span>
            </button>
          ))}
        </div>
      </Card>

      <div className="shipbuilder-random-layout">
        <Card className="shipbuilder-random-card">
          <div className="card-heading">
            <div>
              <p className="eyebrow">{t('shipbuilder.stepGenerate')}</p>
              <h2>{t('shipbuilder.randomTitle')}</h2>
              <small className="muted">{t('shipbuilder.randomHint')}</small>
            </div>
            <span className="badge">#{roll}</span>
          </div>

          <button type="button" className="shipbuilder-random-stage" onClick={reroll} title={t('shipbuilder.clickToReroll')}>
            <ProceduralShipGlyph
              seed={candidate.seed}
              hullId={hullID}
              genome={candidate}
              footprint={shipHullFootprint(hullID)}
              className="shipbuilder-random-ship"
              label={`${hullLabel(t, hullID)} ${roll}`}
            />
            <span className="shipbuilder-random-overlay">{t('shipbuilder.clickToReroll')}</span>
          </button>

          <div className="shipbuilder-random-meta">
            <div>
              <span className="eyebrow">{t('shipbuilder.morphology')}</span>
              <strong>{t(morphologyLabels[candidate.morphologyId])}</strong>
            </div>
            <div>
              <span className="eyebrow">{t('shipbuilder.styleDNA')}</span>
              <strong>{t(styleLabels[candidate.styleId])}</strong>
            </div>
            <div>
              <span className="eyebrow">{t('shipbuilder.randomness')}</span>
              <strong>{candidate.primitives.length} {t('shipbuilder.primitives')}</strong>
            </div>
            <div>
              <span className="eyebrow">{t('shipbuilder.space')}</span>
              <strong>{shipHullSpace(hullID)}</strong>
            </div>
          </div>

          <div className="shipbuilder-random-actions">
            <button type="button" className="button-primary" onClick={reroll}>{t('shipbuilder.generate')}</button>
            <button type="button" className="button-secondary" onClick={() => setKept(candidate)}>{t('shipbuilder.takeDesign')}</button>
          </div>
        </Card>

        <div className="shipbuilder-side-stack random">
          <Card className="shipbuilder-kept-card">
            <p className="eyebrow">{t('shipbuilder.stepKeep')}</p>
            <h2>{t('shipbuilder.kept')}</h2>
            {kept ? (
              <>
                <div className="shipbuilder-kept-stage">
                  <ProceduralShipGlyph seed={kept.seed} hullId={String(kept.hullId)} genome={kept} footprint={shipHullFootprint(kept.hullId)} className="shipbuilder-kept-ship" />
                </div>
                <strong>{hullLabel(t, kept.hullId as HullID)}</strong>
                <small className="muted">{t(morphologyLabels[kept.morphologyId])} · {t(styleLabels[kept.styleId])} · genome v{kept.version} · {t('shipbuilder.spaceValue', { space: shipHullSpace(kept.hullId) })}</small>
              </>
            ) : <p className="muted">{t('shipbuilder.randomKeepHint')}</p>}
          </Card>

          <Card className="shipbuilder-equipment-card">
            <p className="eyebrow">{t('shipbuilder.stepEquipment')}</p>
            <h2>{t('shipbuilder.equipment')}</h2>
            <p className="muted">{t('shipbuilder.equipmentHint')}</p>
            {baseline ? (
              <>
                <div className="shipbuilder-baseline-title"><strong>{baseline.name}</strong><span className="badge">{baseline.spec.hull_id}</span></div>
                <dl className="detail-list compact">
                  <div><dt>{t('shipbuilder.drive')}</dt><dd>{baseline.spec.warp_drive_id}</dd></div>
                  <div><dt>{t('shipbuilder.computer')}</dt><dd>{baseline.spec.computer_id}</dd></div>
                  <div><dt>{t('shipbuilder.armor')}</dt><dd>{baseline.spec.armor_id}</dd></div>
                  <div><dt>{t('shipbuilder.shield')}</dt><dd>{baseline.spec.shield_id ?? t('common.none')}</dd></div>
                  <div><dt>{t('shipbuilder.fuel')}</dt><dd>{baseline.spec.fuel_cell_id}</dd></div>
                  <div><dt>{t('shipbuilder.weapons')}</dt><dd>{baselineWeapons.length > 0 ? baselineWeapons.map((mount) => `${mount.count}x ${mount.weapon_id}`).join(', ') : t('common.none')}</dd></div>
                </dl>
              </>
            ) : <p className="muted">{t('shipbuilder.noBaseline')}</p>}
          </Card>
        </div>
      </div>
    </>
  )
}