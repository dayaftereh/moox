import { useState } from 'react'
import type { PlayerSnapshot } from './api'
import { ProceduralShipGlyph } from './components/ProceduralShipGlyph'
import { Card, PageHeader } from './components/ui'
import type { TranslationKey, TranslationVars } from './i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type HullID = 'frigate' | 'destroyer' | 'cruiser' | 'battleship' | 'titan'

type Candidate = {
  hullID: HullID
  generation: number
  seed: string
}

const hulls: Array<{ id: HullID; label: TranslationKey; authoritative: boolean }> = [
  { id: 'frigate', label: 'shipbuilder.hull.frigate', authoritative: true },
  { id: 'destroyer', label: 'shipbuilder.hull.destroyer', authoritative: false },
  { id: 'cruiser', label: 'shipbuilder.hull.cruiser', authoritative: false },
  { id: 'battleship', label: 'shipbuilder.hull.battleship', authoritative: false },
  { id: 'titan', label: 'shipbuilder.hull.titan', authoritative: false },
]

function candidateFor(gameID: string, hullID: HullID, generation: number): Candidate {
  return { hullID, generation, seed: `shipbuilder:${gameID}:${hullID}:${generation}` }
}

export function ShipBuilderView({ snapshot, t }: { snapshot: PlayerSnapshot; t: Translator }) {
  const [hullID, setHullID] = useState<HullID>('frigate')
  const [generation, setGeneration] = useState(1)
  const [recent, setRecent] = useState<Candidate[]>([])
  const [kept, setKept] = useState<Candidate | null>(null)
  const current = candidateFor(snapshot.view.game_id, hullID, generation)
  const currentHull = hulls.find((hull) => hull.id === hullID) ?? hulls[0]
  const designs = snapshot.decision?.strategic.ship_designs ?? []
  const baseline = designs[0]
  const baselineWeapons = baseline?.spec.weapons ?? []

  function remember(candidate: Candidate) {
    setRecent((items) => [candidate, ...items.filter((item) => item.seed !== candidate.seed)].slice(0, 6))
  }

  function generate() {
    remember(current)
    setGeneration((value) => value + 1)
  }

  function selectHull(nextHullID: HullID) {
    if (nextHullID === hullID) return
    remember(current)
    setHullID(nextHullID)
    setGeneration(1)
  }

  function selectRecent(candidate: Candidate) {
    setHullID(candidate.hullID)
    setGeneration(candidate.generation)
  }

  return (
    <>
      <PageHeader eyebrow={t('shipbuilder.eyebrow')} title={t('shipbuilder.title')} subtitle={t('shipbuilder.subtitle')} />

      <Card className="shipbuilder-size-card">
        <div className="card-heading">
          <div><p className="eyebrow">{t('shipbuilder.stepSize')}</p><h2>{t('shipbuilder.size')}</h2></div>
          <span className="badge">{currentHull.authoritative ? t('shipbuilder.authoritative') : t('shipbuilder.visualPreview')}</span>
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
              <ProceduralShipGlyph seed={`shipbuilder:hull:${hull.id}`} hullId={hull.id} className="shipbuilder-hull-thumb" />
              <span><strong>{t(hull.label)}</strong><small>{hull.authoritative ? t('shipbuilder.authoritative') : t('shipbuilder.visualPreview')}</small></span>
            </button>
          ))}
        </div>
      </Card>

      <div className="shipbuilder-layout">
        <Card className="shipbuilder-generator-card">
          <div className="card-heading">
            <div><p className="eyebrow">{t('shipbuilder.stepGenerate')}</p><h2>{t('shipbuilder.candidate')}</h2></div>
            <span className="badge">{t('shipbuilder.generation', { count: generation })}</span>
          </div>
          <div className="shipbuilder-stage">
            <ProceduralShipGlyph seed={current.seed} hullId={current.hullID} className="shipbuilder-main-ship" label={`${t('shipbuilder.candidate')} ${generation}`} />
          </div>
          <div className="shipbuilder-actions">
            <button type="button" className="button-primary" onClick={generate}>{t('shipbuilder.generate')}</button>
            <button type="button" className="button-secondary" onClick={() => setKept(current)}>{t('shipbuilder.keep')}</button>
          </div>
          <p className="muted shipbuilder-seed">{t('shipbuilder.seed')}: <code>{current.seed}</code></p>

          {recent.length > 0 && (
            <div className="shipbuilder-recent">
              <p className="eyebrow">{t('shipbuilder.recent')}</p>
              <div className="shipbuilder-recent-grid">
                {recent.map((candidate) => (
                  <button type="button" key={candidate.seed} onClick={() => selectRecent(candidate)} className="shipbuilder-recent-item" title={candidate.seed}>
                    <ProceduralShipGlyph seed={candidate.seed} hullId={candidate.hullID} className="shipbuilder-recent-ship" />
                    <small>{t(hulls.find((hull) => hull.id === candidate.hullID)?.label ?? 'shipbuilder.hull.frigate')} · #{candidate.generation}</small>
                  </button>
                ))}
              </div>
            </div>
          )}
        </Card>

        <div className="shipbuilder-side-stack">
          <Card className="shipbuilder-kept-card">
            <p className="eyebrow">{t('shipbuilder.stepKeep')}</p>
            <h2>{t('shipbuilder.kept')}</h2>
            {kept ? (
              <>
                <div className="shipbuilder-kept-stage"><ProceduralShipGlyph seed={kept.seed} hullId={kept.hullID} className="shipbuilder-kept-ship" /></div>
                <strong>{t(hulls.find((hull) => hull.id === kept.hullID)?.label ?? 'shipbuilder.hull.frigate')}</strong>
                <small className="muted">{t('shipbuilder.generation', { count: kept.generation })}</small>
              </>
            ) : <p className="muted">{t('shipbuilder.keepHint')}</p>}
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
                  <div><dt>{t('shipbuilder.weapons')}</dt><dd>{baselineWeapons.length > 0 ? baselineWeapons.map((mount) => `${mount.count}× ${mount.weapon_id}`).join(', ') : t('common.none')}</dd></div>
                </dl>
              </>
            ) : <p className="muted">{t('shipbuilder.noBaseline')}</p>}
          </Card>
        </div>
      </div>
    </>
  )
}
