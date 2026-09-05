import { useMemo, useRef, useState, type PointerEvent as ReactPointerEvent, type WheelEvent as ReactWheelEvent } from 'react'
import {
  aggregatePopulation,
  type Colony,
  type ConstructionChoice,
  type ConstructionState,
  type DraftOrder,
  type DiplomacyCommandKind,
  type DiplomaticStance,
  type PlanningPreviewSnapshot,
  type PlayerSnapshot,
  type PopulationJob,
  type PopulationTransferChoice,
  type ResearchChoice,
  type StarSystem,
} from './api'
import { Card, EmptyState, PageHeader } from './components/ui'
import { type TranslationKey, type TranslationVars } from './i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

function localizedPhase(t: Translator, phase: string): string {
  switch (phase) {
    case 'planning': return t('phase.planning')
    case 'encounters': return t('phase.encounters')
    case 'invasion_decisions': return t('phase.invasion_decisions')
    case 'strategic_resolution': return t('phase.strategic_resolution')
    case 'post_resolution': return t('phase.post_resolution')
    case 'completed': return t('phase.completed')
    default: return t('phase.unknown', { phase })
  }
}

function localizedStance(t: Translator, stance: DiplomaticStance): string {
  if (stance === 'war') return t('stance.war')
  if (stance === 'peace') return t('stance.peace')
  return t('stance.neutral')
}

function serverLabel(t: Translator, key: string | undefined, fallback: string): string {
  if (!key) return fallback
  const value = t(key as TranslationKey)
  return value === key ? fallback : value
}

function humanizeToken(value: string): string {
  return value.replace(/_/g, ' ').replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function formatEta(t: Translator, value: number | undefined): string {
  return value && value > 0 ? t('common.turns', { turns: value }) : t('common.noEta')
}

function bodyLabel(t: Translator, kind: string): string {
  if (kind === 'gas_giant') return t('system.gasGiant')
  if (kind === 'asteroid_belt') return t('system.asteroidBelt')
  return t('system.planet')
}

export function StrategicGalaxyView({ snapshot, selectedSystemID, onSelectSystem, onOpenColony, onPlanOrder, t }: {
  snapshot: PlayerSnapshot
  selectedSystemID?: number
  onSelectSystem: (systemID: number) => void
  onOpenColony: (colonyID: number) => void
  onPlanOrder: (order: DraftOrder) => void
  t: Translator
}) {
  const decision = snapshot.decision
  const systems = decision?.strategic.galaxy.systems ?? []
  const [zoom, setZoom] = useState(1)
  const [pan, setPan] = useState({ x: 0, y: 0 })
  const [dragging, setDragging] = useState(false)
  const ignoreClickRef = useRef(false)
  const gestureRef = useRef<{
    pointers: Map<number, { x: number; y: number }>
    lastCenter?: { x: number; y: number }
    lastDistance?: number
    dragged: boolean
  }>({ pointers: new Map(), dragged: false })
  const bounds = useMemo(() => {
    if (systems.length === 0) return { minX: 0, maxX: 1, minY: 0, maxY: 1 }
    return {
      minX: Math.min(...systems.map((system) => system.x)),
      maxX: Math.max(...systems.map((system) => system.x)),
      minY: Math.min(...systems.map((system) => system.y)),
      maxY: Math.max(...systems.map((system) => system.y)),
    }
  }, [systems])
  const spanX = Math.max(1, bounds.maxX - bounds.minX)
  const spanY = Math.max(1, bounds.maxY - bounds.minY)
  const minZoom = 0.7
  const maxZoom = 4
  const clampZoom = (value: number) => Math.max(minZoom, Math.min(maxZoom, value))

  if (!decision) {
    return (
      <>
        <PageHeader eyebrow={t('galaxy.eyebrow')} title={t('galaxy.title')} subtitle={t('galaxy.subtitle')} />
        <EmptyState title={t('galaxy.mapPending')} body={t('common.notAvailableYet')} />
      </>
    )
  }
  const selected = systems.find((system) => system.id === selectedSystemID)

  const changeZoom = (factor: number) => setZoom((current) => Math.round(clampZoom(current * factor) * 100) / 100)
  const resetView = () => {
    setZoom(1)
    setPan({ x: 0, y: 0 })
  }
  const pointerPair = () => Array.from(gestureRef.current.pointers.values()).slice(0, 2)
  const pairCenter = (points: Array<{ x: number; y: number }>) => ({
    x: (points[0].x + points[1].x) / 2,
    y: (points[0].y + points[1].y) / 2,
  })
  const pairDistance = (points: Array<{ x: number; y: number }>) => Math.hypot(points[1].x - points[0].x, points[1].y - points[0].y)

  const handlePointerDown = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (event.pointerType === 'mouse' && event.button !== 0) return
    event.currentTarget.setPointerCapture(event.pointerId)
    const point = { x: event.clientX, y: event.clientY }
    const gesture = gestureRef.current
    gesture.pointers.set(event.pointerId, point)
    gesture.dragged = false
    ignoreClickRef.current = false
    if (gesture.pointers.size >= 2) {
      const points = pointerPair()
      gesture.lastCenter = pairCenter(points)
      gesture.lastDistance = pairDistance(points)
    } else {
      gesture.lastCenter = point
      gesture.lastDistance = undefined
    }
    setDragging(true)
  }

  const handlePointerMove = (event: ReactPointerEvent<HTMLDivElement>) => {
    const gesture = gestureRef.current
    if (!gesture.pointers.has(event.pointerId)) return
    gesture.pointers.set(event.pointerId, { x: event.clientX, y: event.clientY })

    if (gesture.pointers.size >= 2) {
      const points = pointerPair()
      const center = pairCenter(points)
      const distance = Math.max(1, pairDistance(points))
      if (gesture.lastCenter) {
        const dx = center.x - gesture.lastCenter.x
        const dy = center.y - gesture.lastCenter.y
        if (Math.abs(dx) + Math.abs(dy) > 0.5) {
          setPan((current) => ({ x: current.x + dx, y: current.y + dy }))
          gesture.dragged = true
        }
      }
      if (gesture.lastDistance) {
        const ratio = distance / gesture.lastDistance
        if (Math.abs(ratio - 1) > 0.002) {
          setZoom((current) => clampZoom(current * ratio))
          gesture.dragged = true
        }
      }
      gesture.lastCenter = center
      gesture.lastDistance = distance
    } else {
      const point = gesture.pointers.get(event.pointerId)!
      if (gesture.lastCenter) {
        const dx = point.x - gesture.lastCenter.x
        const dy = point.y - gesture.lastCenter.y
        if (Math.abs(dx) + Math.abs(dy) > 0.5) {
          setPan((current) => ({ x: current.x + dx, y: current.y + dy }))
          gesture.dragged = true
        }
      }
      gesture.lastCenter = point
    }
    if (gesture.dragged) ignoreClickRef.current = true
  }

  const finishPointer = (event: ReactPointerEvent<HTMLDivElement>) => {
    const gesture = gestureRef.current
    gesture.pointers.delete(event.pointerId)
    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId)
    if (gesture.pointers.size === 1) {
      gesture.lastCenter = Array.from(gesture.pointers.values())[0]
      gesture.lastDistance = undefined
    } else if (gesture.pointers.size >= 2) {
      const points = pointerPair()
      gesture.lastCenter = pairCenter(points)
      gesture.lastDistance = pairDistance(points)
    } else {
      gesture.lastCenter = undefined
      gesture.lastDistance = undefined
      setDragging(false)
      if (gesture.dragged) window.setTimeout(() => { ignoreClickRef.current = false }, 0)
    }
  }

  const handleWheel = (event: ReactWheelEvent<HTMLDivElement>) => {
    event.preventDefault()
    changeZoom(event.deltaY < 0 ? 1.12 : 1 / 1.12)
  }

  return (
    <>
      <PageHeader eyebrow={t('galaxy.eyebrow')} title={t('galaxy.title')} />
      <Card className="galaxy-card">
        <div className="card-heading galaxy-map-heading galaxy-map-heading-compact">
          <small className="muted galaxy-map-hint">{t('galaxy.mapHint')}</small>
          <div className="galaxy-map-meta">
            <div className="galaxy-map-tools" role="group" aria-label={t('galaxy.mapControls')}>
              <button type="button" className="button-ghost" aria-label={t('galaxy.zoomOut')} onClick={() => changeZoom(1 / 1.2)}>−</button>
              <span className="badge" aria-live="polite">{Math.round(zoom * 100)}%</span>
              <button type="button" className="button-ghost" aria-label={t('galaxy.zoomIn')} onClick={() => changeZoom(1.2)}>+</button>
              <button type="button" className="button-ghost" onClick={resetView}>{t('galaxy.resetView')}</button>
            </div>
          </div>
        </div>
        <div
          className={'galaxy-map' + (dragging ? ' galaxy-map-dragging' : '')}
          role="list"
          aria-label={t('galaxy.mapTitle')}
          onWheel={handleWheel}
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={finishPointer}
          onPointerCancel={finishPointer}
        >
          <div className="galaxy-map-layer" style={{ transform: 'translate3d(' + pan.x + 'px, ' + pan.y + 'px, 0) scale(' + zoom + ')' }}>
            {systems.map((system) => {
              const ownsColony = system.planets.some((planet) => decision.colonies.some((colony) => colony.planet_id === planet.id))
              const ownOutpost = (system.bodies ?? []).some((body) => Boolean(body.outpost_id) && decision.strategic.outposts?.some((outpost) => outpost.id === body.outpost_id && outpost.empire_id === decision.empire.id))
              return (
                <button
                  key={system.id}
                  type="button"
                  className={'galaxy-node' + (selectedSystemID === system.id ? ' galaxy-node-selected' : '') + (ownsColony ? ' galaxy-node-colony' : ownOutpost ? ' galaxy-node-outpost' : '')}
                  style={{
                    left: (8 + ((system.x - bounds.minX) / spanX) * 84) + '%',
                    top: (8 + ((system.y - bounds.minY) / spanY) * 84) + '%',
                  }}
                  aria-label={system.name}
                  onClick={() => {
                    if (ignoreClickRef.current) return
                    onSelectSystem(system.id)
                  }}
                  title={system.name + ' (' + system.x + ', ' + system.y + ')'}
                >
                  <span className="galaxy-star" aria-hidden="true" />
                  <span className="galaxy-node-label" aria-hidden="true">{system.name}</span>
                </button>
              )
            })}
          </div>
        </div>
      </Card>
      {selected && (
        <SystemDetail
          snapshot={snapshot}
          system={selected}
          onOpenColony={onOpenColony}
          onPlanOrder={onPlanOrder}
          t={t}
        />
      )}
    </>
  )
}

function SystemDetail({ snapshot, system, onOpenColony, onPlanOrder, t }: {
  snapshot: PlayerSnapshot
  system: StarSystem
  onOpenColony: (colonyID: number) => void
  onPlanOrder: (order: DraftOrder) => void
  t: Translator
}) {
  const decision = snapshot.decision
  if (!decision) return null
  const bodies = system.bodies ?? system.planets.map((planet) => ({
    id: planet.id,
    name: planet.name,
    orbit: planet.orbit,
    kind: 'planet' as const,
    planet_id: planet.id,
  }))
  const fleets = decision.strategic.fleets?.filter((fleet) => fleet.at_system_id === system.id) ?? []
  const contacts = decision.strategic.contacts?.filter((contact) => contact.system_id === system.id) ?? []

  return (
    <section className="content-grid content-grid-2 system-detail">
      <Card>
        <div className="card-heading">
          <div><p className="eyebrow">{t('system.title', { system: system.name })}</p><h2>{system.name}</h2></div>
          <span className="badge">{t('system.starClass', { class: system.spectral_class })}</span>
        </div>
        <div className="orbit-list">
          {bodies.map((body) => {
            const colony = body.planet_id ? decision.colonies.find((item) => item.planet_id === body.planet_id) : undefined
            const colonizeChoices = (decision.decisions.colonization ?? []).filter((choice) => choice.system_id === system.id && choice.planet_id === body.planet_id)
            const outpostChoices = (decision.decisions.outpost_deployment ?? []).filter((choice) => choice.system_id === system.id && choice.body_id === body.id)
            return (
              <div className="orbit-row" key={body.id}>
                <div>
                  <strong>{body.orbit}. {body.name}</strong>
                  <small>{bodyLabel(t, body.kind)}</small>
                </div>
                <div className="action-row compact-actions">
                  {colony && <button type="button" className="button-secondary" onClick={() => onOpenColony(colony.id)}>{t('system.openColony')}</button>}
                  {colonizeChoices.map((choice) => (
                    <button
                      type="button"
                      className="button-primary"
                      key={`colonize-${choice.fleet_id}`}
                      onClick={() => {
                        if (!window.confirm(t('system.colonizeConfirm', { body: body.name, fleet: choice.fleet_id }))) return
                        onPlanOrder({
                          key: `fleet:${choice.fleet_id}`,
                          kind: 'empire.colonize_planet',
                          payload: { fleet_id: choice.fleet_id, planet_id: choice.planet_id },
                        })
                      }}
                    >
                      {t('system.colonize')}
                    </button>
                  ))}
                  {outpostChoices.map((choice) => (
                    <button
                      type="button"
                      className="button-secondary"
                      key={`outpost-${choice.fleet_id}`}
                      onClick={() => {
                        if (!window.confirm(t('system.outpostConfirm', { body: body.name, fleet: choice.fleet_id }))) return
                        onPlanOrder({
                          key: `fleet:${choice.fleet_id}`,
                          kind: 'fleet.deploy_outpost',
                          payload: { fleet_id: choice.fleet_id, body_id: choice.body_id },
                        })
                      }}
                    >
                      {t('system.buildOutpost')}
                    </button>
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      </Card>
      <Card>
        <h2>{t('system.fleets')}</h2>
        {fleets.length === 0 && contacts.length === 0 ? <p className="muted">{t('common.none')}</p> : (
          <div className="list-stack">
            {fleets.map((fleet) => (
              <div className="list-row" key={`own-${fleet.id}`}>
                <span><strong>{t('galaxy.fleet', { id: fleet.id })}</strong><small>{fleet.role}{fleet.special_kind ? ` · ${fleet.special_kind}` : ''}</small></span>
                <span className="badge">{fleet.ship_ids?.length ?? 0}</span>
              </div>
            ))}
            {contacts.map((contact, index) => (
              <div className="list-row" key={`contact-${index}-${contact.empire_id}`}>
                <span><strong>{t('common.empireFallback', { id: contact.empire_id })}</strong><small>{contact.kind}</small></span>
                <span className="badge">{contact.role ?? contact.special_kind ?? '—'}</span>
              </div>
            ))}
          </div>
        )}
      </Card>
    </section>
  )
}

export function StrategicColoniesView({ snapshot, preview, draftOrders, selectedColonyID, onOpenColony, onBack, onPlanPopulation, onPlanOrder, onRemoveOrder, t }: {
  snapshot: PlayerSnapshot
  preview: PlanningPreviewSnapshot | null
  draftOrders: DraftOrder[]
  selectedColonyID?: number
  onOpenColony: (colonyID: number) => void
  onBack: () => void
  onPlanPopulation: (colonyID: number, farmers: number, workers: number, scientists: number) => void
  onPlanOrder: (order: DraftOrder) => void
  onRemoveOrder: (key: string) => void
  t: Translator
}) {
  const colonies = snapshot.decision?.colonies ?? snapshot.view.colonies
  const projected = preview?.preview.projection.colonies ?? []
  if (selectedColonyID) {
    const colony = colonies.find((item) => item.id === selectedColonyID)
    if (!colony) return <EmptyState title={t('colonies.noColonies')} />
    return (
      <ColonyDetail
        snapshot={snapshot}
        colony={colony}
        preview={projected.find((item) => item.colony.id === colony.id)}
        draftOrders={draftOrders}
        onBack={onBack}
        onPlanPopulation={onPlanPopulation}
        onPlanOrder={onPlanOrder}
        onRemoveOrder={onRemoveOrder}
        t={t}
      />
    )
  }

  return (
    <>
      <PageHeader eyebrow={t('colonies.eyebrow')} title={t('colonies.title')} subtitle={t('colonies.subtitle')} />
      {colonies.length === 0 ? <EmptyState title={t('colonies.noColonies')} /> : (
        <div className="table-scroll">
          <table className="colony-table">
            <thead>
              <tr>
                <th>{t('colonies.tableColony')}</th>
                <th>{t('colonies.tablePopulation')}</th>
                <th>{t('colonies.tableFood')}</th>
                <th>{t('colonies.tableProduction')}</th>
                <th>{t('colonies.tableResearch')}</th>
                <th>{t('colonies.tableBuild')}</th>
                <th>{t('common.details')}</th>
              </tr>
            </thead>
            <tbody>
              {colonies.map((baseColony) => {
                const previewColony = projected.find((item) => item.colony.id === baseColony.id)
                const colony = previewColony?.colony ?? baseColony
                const population = aggregatePopulation(colony)
                const build = previewColony?.construction?.[0]
                return (
                  <tr key={colony.id}>
                    <td><strong>{t('colonies.colony', { id: colony.id })}</strong><small>{t('colonies.planet', { id: colony.planet_id })}</small></td>
                    <td>
                      <strong>{population.total.toFixed(1)} / {colony.population_dynamics.capacity.toFixed(1)}</strong>
                      <small>{population.farmers}/{population.workers}/{population.scientists}</small>
                    </td>
                    <td>{colony.adjusted_economy.food.toFixed(1)}</td>
                    <td>{colony.adjusted_economy.production.toFixed(1)}</td>
                    <td>{colony.adjusted_economy.research.toFixed(1)}</td>
                    <td><strong>{build?.project.project_id ?? colony.construction?.project_id ?? t('construction.empty')}</strong><small>{formatEta(t, build?.eta_turns)}</small></td>
                    <td>
                      <div className="action-row compact-actions">
                        <button type="button" className="button-secondary" onClick={() => onOpenColony(colony.id)}>{t('colonies.open')}</button>
                        <PopulationMoveControls colony={colony} onPlan={onPlanPopulation} t={t} compact />
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </>
  )
}

function PopulationMoveControls({ colony, onPlan, t, compact = false }: {
  colony: Colony
  onPlan: (colonyID: number, farmers: number, workers: number, scientists: number) => void
  t: Translator
  compact?: boolean
}) {
  const [source, setSource] = useState<PopulationJob>('farmer')
  const [destination, setDestination] = useState<PopulationJob>('worker')
  const population = aggregatePopulation(colony)
  const values: Record<PopulationJob, number> = {
    farmer: population.farmers,
    worker: population.workers,
    scientist: population.scientists,
  }
  const labels: Record<PopulationJob, string> = {
    farmer: t('jobs.farmer'),
    worker: t('jobs.worker'),
    scientist: t('jobs.scientist'),
  }
  const canMove = source !== destination && values[source] > 0

  function moveOne() {
    if (!canMove) return
    const amount = Math.min(1, values[source])
    const next = { ...values, [source]: values[source] - amount, [destination]: values[destination] + amount }
    onPlan(colony.id, next.farmer, next.worker, next.scientist)
  }

  return (
    <div className={compact ? 'job-move job-move-compact' : 'job-move'}>
      {!compact && (
        <div className="summary-stats">
          <div><span>{labels.farmer}</span><strong>{values.farmer.toFixed(1)}</strong></div>
          <div><span>{labels.worker}</span><strong>{values.worker.toFixed(1)}</strong></div>
          <div><span>{labels.scientist}</span><strong>{values.scientist.toFixed(1)}</strong></div>
        </div>
      )}
      <select aria-label={labels[source]} value={source} onChange={(event) => setSource(event.target.value as PopulationJob)}>
        {(['farmer', 'worker', 'scientist'] as PopulationJob[]).map((job) => <option key={job} value={job}>{labels[job]}</option>)}
      </select>
      <span aria-hidden="true">→</span>
      <select aria-label={labels[destination]} value={destination} onChange={(event) => setDestination(event.target.value as PopulationJob)}>
        {(['farmer', 'worker', 'scientist'] as PopulationJob[]).map((job) => <option key={job} value={job}>{labels[job]}</option>)}
      </select>
      <button type="button" className="button-secondary" disabled={!canMove} onClick={moveOne}>{t('jobs.moveOne')}</button>
    </div>
  )
}

type DraftQueueItem = {
  project_kind: ConstructionState['project_kind']
  project_id: string
  ship_design_id?: number
  ship_design_revision?: number
}

function queueItemsFromColony(colony: Colony): DraftQueueItem[] {
  const states = [colony.construction, ...(colony.construction_queue ?? [])].filter((item): item is ConstructionState => Boolean(item))
  return states.map((item) => ({
    project_kind: item.project_kind,
    project_id: item.project_id,
    ...(item.ship_design_id ? { ship_design_id: item.ship_design_id } : {}),
    ...(item.ship_design_revision ? { ship_design_revision: item.ship_design_revision } : {}),
  }))
}

function queueItemFromChoice(choice: ConstructionChoice): DraftQueueItem {
  return {
    project_kind: choice.project_kind,
    project_id: choice.project_id,
    ...(choice.ship_design_id ? { ship_design_id: choice.ship_design_id } : {}),
    ...(choice.ship_design_revision ? { ship_design_revision: choice.ship_design_revision } : {}),
  }
}

function ColonyDetail({ snapshot, colony, preview, draftOrders, onBack, onPlanPopulation, onPlanOrder, onRemoveOrder, t }: {
  snapshot: PlayerSnapshot
  colony: Colony
  preview?: PlanningPreviewSnapshot['preview']['projection']['colonies'][number]
  draftOrders: DraftOrder[]
  onBack: () => void
  onPlanPopulation: (colonyID: number, farmers: number, workers: number, scientists: number) => void
  onPlanOrder: (order: DraftOrder) => void
  onRemoveOrder: (key: string) => void
  t: Translator
}) {
  const displayColony = preview?.colony ?? colony
  const decision = snapshot.decision
  const galaxy = decision?.strategic.galaxy
  const planet = galaxy?.systems.flatMap((system) => system.planets.map((item) => ({ system, planet: item }))).find((item) => item.planet.id === colony.planet_id)
  const population = aggregatePopulation(displayColony)
  const growth = preview?.population_growth_per_turn ?? Math.max(0, displayColony.population_dynamics.projected_growth)
  const loss = preview?.population_loss_per_turn ?? Math.max(0, displayColony.population_dynamics.projected_starvation)
  const freeCapacity = preview?.free_population_capacity ?? Math.max(0, displayColony.population_dynamics.capacity - population.total)
  const constructionDecision = decision?.decisions.construction?.find((item) => item.colony_id === colony.id)
  const transferChoices = decision?.decisions.population_transfers?.filter((item) => item.source_colony_id === colony.id) ?? []

  return (
    <>
      <PageHeader
        eyebrow={t('colonies.eyebrow')}
        title={t('colonies.colony', { id: colony.id })}
        subtitle={planet ? `${planet.system.name} · ${planet.planet.name}` : t('colonies.planet', { id: colony.planet_id })}
        actions={<button type="button" className="button-ghost" onClick={onBack}>{t('common.back')}</button>}
      />
      <div className="content-grid content-grid-2">
        <Card>
          <p className="eyebrow">{t('colony.overview')}</p>
          <h2>{t('colony.population')}</h2>
          <dl className="detail-list">
            <div><dt>{t('colonies.population')}</dt><dd>{population.total.toFixed(2)} / {displayColony.population_dynamics.capacity.toFixed(2)}</dd></div>
            <div><dt>{t('colonies.freeCapacity')}</dt><dd>{freeCapacity.toFixed(2)}</dd></div>
            <div><dt>{loss > 0 ? t('colonies.starvation') : t('colonies.growth')}</dt><dd>{loss > 0 ? `-${loss.toFixed(2)}` : `+${growth.toFixed(2)}`}</dd></div>
            <div><dt>{t('colonies.nextPop')}</dt><dd>{formatEta(t, preview?.next_population_eta_turns)}</dd></div>
            <div><dt>{t('colony.groundForces')}</dt><dd>{displayColony.ground_forces?.infantry ?? 0}</dd></div>
          </dl>
          <PopulationMoveControls colony={displayColony} onPlan={onPlanPopulation} t={t} />
        </Card>
        <Card>
          <p className="eyebrow">{t('colony.outputs')}</p>
          <div className="summary-stats">
            <div><span>{t('colonies.tableFood')}</span><strong>{displayColony.adjusted_economy.food.toFixed(1)}</strong></div>
            <div><span>{t('colonies.tableProduction')}</span><strong>{displayColony.adjusted_economy.production.toFixed(1)}</strong></div>
            <div><span>{t('colonies.tableResearch')}</span><strong>{displayColony.adjusted_economy.research.toFixed(1)}</strong></div>
          </div>
          {planet && (
            <>
              <h3>{t('colony.planetTraits')}</h3>
              <p className="muted">{planet.planet.climate_id} · {planet.planet.size_id} · {planet.planet.mineral_id} · {planet.planet.gravity_id}</p>
            </>
          )}
        </Card>
        <Card>
          <p className="eyebrow">{t('colony.buildings')}</p>
          {(displayColony.buildings ?? []).length === 0 ? <p className="muted">{t('colony.noBuildings')}</p> : (
            <div className="tag-list">{(displayColony.buildings ?? []).map((building) => <span className="badge" key={building}>{building}</span>)}</div>
          )}
        </Card>
        <ConstructionEditor
          colony={displayColony}
          preview={preview}
          choices={constructionDecision?.choices ?? []}
          draftOrders={draftOrders}
          onPlanOrder={onPlanOrder}
          onRemoveOrder={onRemoveOrder}
          t={t}
        />
        <Card className="span-two">
          <p className="eyebrow">{t('transfer.title')}</p>
          {transferChoices.length === 0 ? <p className="muted">{t('common.none')}</p> : (
            <div className="list-stack">
              {transferChoices.slice(0, 24).map((choice, index) => (
                <PopulationTransferRow
                  key={`${choice.destination_colony_id}-${choice.source_job}-${choice.destination_job}-${index}`}
                  choice={choice}
                  onPlanOrder={onPlanOrder}
                  t={t}
                />
              ))}
            </div>
          )}
        </Card>
      </div>
    </>
  )
}

function ConstructionEditor({ colony, preview, choices, draftOrders, onPlanOrder, onRemoveOrder, t }: {
  colony: Colony
  preview?: PlanningPreviewSnapshot['preview']['projection']['colonies'][number]
  choices: ConstructionChoice[]
  draftOrders: DraftOrder[]
  onPlanOrder: (order: DraftOrder) => void
  onRemoveOrder: (key: string) => void
  t: Translator
}) {
  const draftKey = `construction:${colony.id}`
  const drafted = draftOrders.find((order) => order.key === draftKey)
  const draftedItems = drafted?.payload.items
  const items = Array.isArray(draftedItems) ? draftedItems as DraftQueueItem[] : queueItemsFromColony(colony)

  function save(next: DraftQueueItem[]) {
    onPlanOrder({ key: draftKey, kind: 'colony.set_construction_queue', payload: { colony_id: colony.id, items: next } })
  }

  function move(index: number, direction: -1 | 1) {
    const target = index + direction
    if (target < 0 || target >= items.length) return
    const next = [...items]
    ;[next[index], next[target]] = [next[target], next[index]]
    save(next)
  }

  return (
    <Card>
      <p className="eyebrow">{t('colony.construction')}</p>
      <h2>{t('construction.queue')}</h2>
      {items.length === 0 ? <p className="muted">{t('construction.empty')}</p> : (
        <div className="list-stack">
          {items.map((item, index) => {
            const projected = preview?.construction?.[index]
            return (
              <div className="queue-row" key={`${item.project_kind}-${item.project_id}-${index}`}>
                <span>
                  <strong>{humanizeToken(item.project_id)}</strong>
                  <small>{item.project_kind !== item.project_id ? `${humanizeToken(item.project_kind)} · ` : ''}{projected ? formatEta(t, projected.eta_turns) : t('common.noEta')}</small>
                </span>
                <div className="action-row compact-actions">
                  <button type="button" className="button-ghost" disabled={index === 0} onClick={() => move(index, -1)}>{t('construction.moveUp')}</button>
                  <button type="button" className="button-ghost" disabled={index === items.length - 1} onClick={() => move(index, 1)}>{t('construction.moveDown')}</button>
                  <button type="button" className="button-ghost" onClick={() => save(items.filter((_, itemIndex) => itemIndex !== index))}>{t('common.remove')}</button>
                </div>
              </div>
            )
          })}
        </div>
      )}
      <h3>{t('construction.available')}</h3>
      <div className="choice-grid">
        {choices.map((choice, index) => {
          const nonRepeatable = choice.project_kind === 'building' || choice.project_kind === 'planetary_transformation'
          const alreadyQueued = nonRepeatable && items.some((item) => item.project_kind === choice.project_kind && item.project_id === choice.project_id)
          return (
            <button
              type="button"
              className="choice-button"
              key={`${choice.project_kind}-${choice.project_id}-${index}`}
              disabled={alreadyQueued}
              onClick={() => save([...items, queueItemFromChoice(choice)])}
            >
              <strong>{humanizeToken(choice.ship_design_name ?? choice.project_id)}</strong>
              <small>{choice.project_kind !== choice.project_id ? `${choice.project_kind} · ` : ''}{t('construction.cost', { pp: choice.production_cost_pp.toFixed(0) })}</small>
            </button>
          )
        })}
      </div>
      {drafted && <button type="button" className="button-ghost" onClick={() => onRemoveOrder(draftKey)}>{t('common.cancel')}</button>}
    </Card>
  )
}

function PopulationTransferRow({ choice, onPlanOrder, t }: {
  choice: PopulationTransferChoice
  onPlanOrder: (order: DraftOrder) => void
  t: Translator
}) {
  function confirmTransfer() {
    const message = [
      t('transfer.title'),
      `${choice.source_colony_id} → ${choice.destination_colony_id}`,
      `${choice.source_job} → ${choice.destination_job}`,
      `${t('transfer.freighters')}: ${choice.freighters_required}`,
      `${t('transfer.eta')}: ${choice.eta}`,
    ].join('\n')
    if (!window.confirm(message)) return
    const cohort = choice.cohort
    onPlanOrder({
      key: `transfer:${choice.source_colony_id}:${choice.destination_colony_id}:${cohort.origin_empire_id}:${cohort.loyalty_empire_id}:${cohort.assimilation_state}:${choice.source_job}:${choice.destination_job}`,
      kind: 'colony.transfer_population',
      payload: {
        source_colony_id: choice.source_colony_id,
        destination_colony_id: choice.destination_colony_id,
        cohort: choice.cohort,
        source_job: choice.source_job,
        destination_job: choice.destination_job,
      },
    })
  }

  return (
    <div className="list-row">
      <span>
        <strong>{choice.source_colony_id} → {choice.destination_colony_id}</strong>
        <small>{choice.source_job} → {choice.destination_job} · {t('transfer.freighters')}: {choice.freighters_required} · {t('transfer.eta')}: {choice.eta}{choice.same_system ? ` · ${t('transfer.sameSystem')}` : ''}</small>
      </span>
      <button type="button" className="button-secondary" onClick={confirmTransfer}>{t('transfer.confirm')}</button>
    </div>
  )
}

export function StrategicFleetsView({ snapshot, onPlanOrder, t }: { snapshot: PlayerSnapshot; onPlanOrder: (order: DraftOrder) => void; t: Translator }) {
  const decision = snapshot.decision
  const fleets = decision?.strategic.fleets ?? []
  return (
    <>
      <PageHeader eyebrow={t('fleets.eyebrow')} title={t('fleets.title')} subtitle={t('fleets.subtitle')} />
      <div className="content-grid content-grid-2">
        {fleets.length === 0 ? <EmptyState title={t('fleets.noProjection')} /> : fleets.map((fleet) => {
          const moves = decision?.decisions.fleet_moves?.filter((choice) => choice.fleet_id === fleet.id) ?? []
          return (
            <Card key={fleet.id}>
              <div className="card-heading">
                <div><p className="eyebrow">{t('galaxy.fleet', { id: fleet.id })}</p><h2>{fleet.role}</h2></div>
                <span className="badge">{t('fleets.ships')}: {fleet.ship_ids?.length ?? 0}</span>
              </div>
              <dl className="detail-list compact">
                <div><dt>{t('fleets.atSystem')}</dt><dd>{fleet.at_system_id ?? '—'}</dd></div>
                <div><dt>{t('fleets.destination')}</dt><dd>{fleet.destination_system_id ?? '—'}</dd></div>
                <div><dt>{t('fleets.eta')}</dt><dd>{fleet.remaining_turns ?? '—'}</dd></div>
              </dl>
              {moves.length === 0 ? <p className="muted">{t('fleets.noMoves')}</p> : (
                <div className="choice-grid">
                  {moves.map((choice) => (
                    <button
                      type="button"
                      className="choice-button"
                      key={`${choice.fleet_id}-${choice.destination_system_id}`}
                      onClick={() => onPlanOrder({
                        key: `fleet:${choice.fleet_id}`,
                        kind: 'empire.move_fleet',
                        payload: {
                          fleet_id: choice.fleet_id,
                          destination_system_id: choice.destination_system_id,
                          ...(choice.ship_ids?.length ? { ship_ids: choice.ship_ids } : {}),
                        },
                      })}
                    >
                      <strong>{t('fleets.move')} → {choice.destination_system_id}</strong>
                      <small>{t('fleets.eta')}: {choice.eta}</small>
                    </button>
                  ))}
                </div>
              )}
            </Card>
          )
        })}
        <Card>
          <h2>{t('fleets.battles')}</h2>
          {snapshot.battles.length === 0 ? <p className="muted">{t('fleets.noBattles')}</p> : (
            <div className="list-stack">{snapshot.battles.map((battle) => <div className="list-row" key={battle.spec.id}><span><strong>{t('fleets.battle', { id: battle.spec.id })}</strong><small>{localizedPhase(t, battle.phase)}</small></span><span className="badge">{battle.spec.participants.length}</span></div>)}</div>
          )}
        </Card>
      </div>
    </>
  )
}

export function StrategicResearchView({ snapshot, preview, onPlanOrder, t }: {
  snapshot: PlayerSnapshot
  preview: PlanningPreviewSnapshot | null
  onPlanOrder: (order: DraftOrder) => void
  t: Translator
}) {
  const decision = snapshot.decision
  const categories = [...(decision?.decisions.research_categories ?? [])].sort((a, b) => a.order - b.order)
  const choices = decision?.decisions.research ?? []
  const active = preview?.preview.projection.research

  function planResearch(choice: ResearchChoice, technologyID = 0) {
    onPlanOrder({
      key: 'research',
      kind: 'empire.select_research',
      payload: {
        tech_field_id: choice.tech_field_id,
        ...(technologyID ? { technology_id: technologyID } : {}),
      },
    })
  }

  return (
    <>
      <PageHeader eyebrow={t('research.eyebrow')} title={t('research.title')} subtitle={t('research.subtitle')} />
      <Card className="research-active">
        <p className="eyebrow">{t('research.active')}</p>
        <div className="summary-stats">
          <div><span>{t('research.rpRate')}</span><strong>{active?.rp_per_turn.toFixed(1) ?? '—'}</strong></div>
          <div><span>{t('research.eta')}</span><strong>{formatEta(t, active?.eta_turns)}</strong></div>
          <div><span>RP</span><strong>{active ? `${active.progress_rp.toFixed(0)} / ${active.cost_rp.toFixed(0)}` : '—'}</strong></div>
        </div>
      </Card>
      <div className="content-grid content-grid-2">
        {categories.length === 0 ? <EmptyState title={t('research.pending')} /> : categories.map((category) => {
          const choice = choices.find((item) => item.category_id === category.id)
          return (
            <Card key={category.id}>
              <p className="eyebrow">{category.order}. {serverLabel(t, category.name_key, category.id)}</p>
              <h2>{serverLabel(t, category.name_key, category.id)}</h2>
              {!choice ? <p className="muted">{t('common.none')}</p> : (
                <>
                  <dl className="detail-list compact">
                    <div><dt>Field</dt><dd>{choice.tech_field_id}</dd></div>
                    <div><dt>{t('research.cost')}</dt><dd>{choice.base_cost_rp.toFixed(0)} RP</dd></div>
                    <div><dt>Mode</dt><dd>{choice.selection_mode}</dd></div>
                  </dl>
                  {choice.selection_mode === 'choose_one' ? (
                    <div className="choice-grid">
                      {choice.technology_ids.map((technologyID, index) => (
                        <button type="button" className="choice-button" key={technologyID} onClick={() => planResearch(choice, technologyID)}>
                          <strong>{serverLabel(t, choice.technology_name_keys[index], humanizeToken(choice.technology_keys[index] ?? String(technologyID)))}</strong>
                          <small>{t('research.chooseApplication')}</small>
                        </button>
                      ))}
                    </div>
                  ) : (
                    <button type="button" className="button-primary" onClick={() => planResearch(choice)}>
                      {t('research.select')}
                    </button>
                  )}
                </>
              )}
            </Card>
          )
        })}
      </div>
    </>
  )
}

export function StrategicDiplomacyView({ snapshot, busy, closed, onCommand, t }: {
  snapshot: PlayerSnapshot
  busy: boolean
  closed: boolean
  onCommand: (kind: DiplomacyCommandKind, otherEmpireID: number) => Promise<void>
  t: Translator
}) {
  const relations = snapshot.decision?.diplomacy ?? snapshot.view.diplomacy ?? []
  const legal = snapshot.decision?.decisions.diplomacy ?? []
  return (
    <>
      <PageHeader eyebrow={t('diplomacy.eyebrow')} title={t('diplomacy.title')} subtitle={t('diplomacy.subtitle')} />
      {relations.length === 0 ? <EmptyState title={t('diplomacy.noContacts')} /> : (
        <div className="content-grid content-grid-2">
          {relations.map((relation) => {
            const otherSeat = snapshot.view.seats.find((candidate) => candidate.seat.empire_id === relation.other_empire_id)
            const actions = legal.filter((decision) => decision.target_empire_id === relation.other_empire_id)
            return (
              <Card key={relation.other_empire_id}>
                <div className="card-heading">
                  <div><p className="eyebrow">{t('common.empireFallback', { id: relation.other_empire_id })}</p><h2>{otherSeat?.seat.name ?? t('common.empireFallback', { id: relation.other_empire_id })}</h2></div>
                  <span className="badge">{localizedStance(t, relation.stance)}</span>
                </div>
                <div className="action-row">
                  {actions.map((action) => (
                    <button
                      type="button"
                      className={action.kind === 'diplomacy.declare_war' ? 'button-danger' : 'button-secondary'}
                      disabled={busy || closed}
                      key={action.kind}
                      onClick={() => void onCommand(action.kind, relation.other_empire_id)}
                    >
                      {action.kind === 'diplomacy.declare_war'
                        ? t('more.declareWar')
                        : action.kind === 'diplomacy.offer_peace'
                          ? t('more.offerPeace')
                          : t('more.acceptPeace')}
                    </button>
                  ))}
                </div>
              </Card>
            )
          })}
        </div>
      )}
    </>
  )
}

export function StrategicEspionageView({ t }: { t: Translator }) {
  return (
    <>
      <PageHeader eyebrow={t('espionage.eyebrow')} title={t('espionage.title')} subtitle={t('espionage.subtitle')} />
      <EmptyState title={t('espionage.unavailable')} />
    </>
  )
}
