import { useMemo, useState } from 'react'
import type { PlayerSnapshot } from './api'
import { ProceduralShipGlyph } from './components/ProceduralShipGlyph'
import { Card, PageHeader } from './components/ui'
import type { TranslationKey, TranslationVars } from './i18n'
import { createShipGenome, mutateShipGenome, type ShipVisualGenome } from './shipVisualGenome'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type HullID = 'scout' | 'frigate' | 'destroyer' | 'cruiser' | 'battleship' | 'titan' | 'doom_star'

type Candidate = {
  hullID: HullID
  generation: number
  index: number
  seed: string
  genome: ShipVisualGenome
}

const populationSize = 6

const hulls: Array<{ id: HullID; label: TranslationKey; status: TranslationKey }> = [
  { id: 'scout', label: 'shipbuilder.hull.scout', status: 'shipbuilder.scoutRole' },
  { id: 'frigate', label: 'shipbuilder.hull.frigate', status: 'shipbuilder.authoritative' },
  { id: 'destroyer', label: 'shipbuilder.hull.destroyer', status: 'shipbuilder.visualPreview' },
  { id: 'cruiser', label: 'shipbuilder.hull.cruiser', status: 'shipbuilder.visualPreview' },
  { id: 'battleship', label: 'shipbuilder.hull.battleship', status: 'shipbuilder.visualPreview' },
  { id: 'titan', label: 'shipbuilder.hull.titan', status: 'shipbuilder.visualPreview' },
  { id: 'doom_star', label: 'shipbuilder.hull.doomStar', status: 'shipbuilder.visualPreview' },
]

function hullLabel(t: Translator, hullID: HullID): string {
  return t(hulls.find((hull) => hull.id === hullID)?.label ?? 'shipbuilder.hull.frigate')
}

export function ShipBuilderView({ snapshot, t }: { snapshot: PlayerSnapshot; t: Translator }) {
  const [hullID, setHullID] = useState<HullID>('scout')
  const [generation, setGeneration] = useState(1)
  const [familyRound, setFamilyRound] = useState(1)
  const [mutation, setMutation] = useState(.32)
  const [parent, setParent] = useState<Candidate | null>(null)
  const [kept, setKept] = useState<Candidate | null>(null)
  const currentHull = hulls.find((hull) => hull.id === hullID) ?? hulls[0]
  const designs = snapshot.decision?.strategic.ship_designs ?? []
  const baseline = designs[0]
  const baselineWeapons = baseline?.spec.weapons ?? []

  const candidates = useMemo<Candidate[]>(() => {
    return Array.from({ length: populationSize }, (_, index) => {
      const seed = parent && parent.hullID === hullID
        ? `shipbuilder:${snapshot.view.game_id}:${hullID}:evolve:${parent.seed}:g${generation}:c${index + 1}`
        : `shipbuilder:${snapshot.view.game_id}:${hullID}:family:${familyRound}:c${index + 1}`
      const genome = parent && parent.hullID === hullID
        ? mutateShipGenome(parent.genome, seed, mutation)
        : createShipGenome(seed, hullID)
      return { hullID, generation, index, seed, genome }
    })
  }, [familyRound, generation, hullID, mutation, parent, snapshot.view.game_id])

  function selectHull(nextHullID: HullID) {
    if (nextHullID === hullID) return
    setHullID(nextHullID)
    setParent(null)
    setGeneration(1)
    setFamilyRound((value) => value + 1)
  }

  function evolveFrom(candidate: Candidate) {
    setParent(candidate)
    setGeneration((value) => value + 1)
  }

  function freshFamily() {
    setParent(null)
    setGeneration(1)
    setFamilyRound((value) => value + 1)
  }

  return (
    <>
      <PageHeader eyebrow={t('shipbuilder.eyebrow')} title={t('shipbuilder.title')} subtitle={t('shipbuilder.subtitle')} />

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
              <ProceduralShipGlyph seed={`shipbuilder:hull:${hull.id}`} hullId={hull.id} className="shipbuilder-hull-thumb" />
              <span><strong>{t(hull.label)}</strong><small>{t(hull.status)}</small></span>
            </button>
          ))}
        </div>
      </Card>

      <div className="shipbuilder-evolution-layout">
        <Card className="shipbuilder-evolution-card">
          <div className="card-heading shipbuilder-evolution-heading">
            <div>
              <p className="eyebrow">{t('shipbuilder.stepGenerate')}</p>
              <h2>{parent ? t('shipbuilder.evolutionChildren') : t('shipbuilder.evolutionPopulation')}</h2>
              <small className="muted">
                {parent
                  ? t('shipbuilder.parentHint', { hull: hullLabel(t, parent.hullID), generation: parent.generation, candidate: parent.index + 1 })
                  : t('shipbuilder.populationHint')}
              </small>
            </div>
            <span className="badge">{t('shipbuilder.generation', { count: generation })}</span>
          </div>

          <div className="shipbuilder-evolution-controls">
            <label className="shipbuilder-mutation-control">
              <span><strong>{t('shipbuilder.shapeMutation')}</strong><small>{Math.round(mutation * 100)}%</small></span>
              <input
                type="range"
                min="0.05"
                max="0.9"
                step="0.05"
                value={mutation}
                onChange={(event) => setMutation(Number(event.target.value))}
              />
            </label>
            <button type="button" className="button-secondary" onClick={freshFamily}>{t('shipbuilder.freshFamily')}</button>
          </div>

          <div className="shipbuilder-candidate-grid">
            {candidates.map((candidate) => (
              <article className="shipbuilder-candidate" key={candidate.seed} data-candidate-index={candidate.index + 1}>
                <button type="button" className="shipbuilder-candidate-visual" onClick={() => evolveFrom(candidate)} title={t('shipbuilder.evolveTitle')}>
                  <ProceduralShipGlyph
                    seed={candidate.seed}
                    hullId={candidate.hullID}
                    genome={candidate.genome}
                    className="shipbuilder-candidate-ship"
                    label={`${hullLabel(t, candidate.hullID)} ${candidate.index + 1}`}
                  />
                  <span className="shipbuilder-candidate-number">{candidate.index + 1}</span>
                </button>
                <div className="shipbuilder-candidate-meta">
                  <span><strong>{hullLabel(t, candidate.hullID)}</strong><small>{candidate.genome.primitives.length} {t('shipbuilder.primitives')}</small></span>
                  <div className="shipbuilder-candidate-actions">
                    <button type="button" className="button-primary" onClick={() => evolveFrom(candidate)}>{t('shipbuilder.evolve')}</button>
                    <button type="button" className="button-secondary" onClick={() => setKept(candidate)}>{t('shipbuilder.keep')}</button>
                  </div>
                </div>
              </article>
            ))}
          </div>

          <p className="muted shipbuilder-evolution-note">{t('shipbuilder.evolutionNote')}</p>
        </Card>

        <div className="shipbuilder-side-stack evolution">
          <Card className="shipbuilder-parent-card">
            <p className="eyebrow">{t('shipbuilder.evolutionParent')}</p>
            <h2>{parent ? t('shipbuilder.selectedParent') : t('shipbuilder.noParent')}</h2>
            {parent ? (
              <>
                <div className="shipbuilder-kept-stage">
                  <ProceduralShipGlyph seed={parent.seed} hullId={parent.hullID} genome={parent.genome} className="shipbuilder-kept-ship" />
                </div>
                <strong>{hullLabel(t, parent.hullID)} · #{parent.index + 1}</strong>
                <small className="muted">{t('shipbuilder.generation', { count: parent.generation })} · {parent.genome.primitives.length} {t('shipbuilder.primitives')}</small>
              </>
            ) : <p className="muted">{t('shipbuilder.noParentHint')}</p>}
          </Card>

          <Card className="shipbuilder-kept-card">
            <p className="eyebrow">{t('shipbuilder.stepKeep')}</p>
            <h2>{t('shipbuilder.kept')}</h2>
            {kept ? (
              <>
                <div className="shipbuilder-kept-stage"><ProceduralShipGlyph seed={kept.seed} hullId={kept.hullID} genome={kept.genome} className="shipbuilder-kept-ship" /></div>
                <strong>{hullLabel(t, kept.hullID)} · #{kept.index + 1}</strong>
                <small className="muted">{t('shipbuilder.generation', { count: kept.generation })} · genome v{kept.genome.version}</small>
              </>
            ) : <p className="muted">{t('shipbuilder.keepHintEvolution')}</p>}
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