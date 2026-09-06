import { useEffect, useMemo, useRef, useState, type DragEvent as ReactDragEvent, type PointerEvent as ReactPointerEvent, type WheelEvent as ReactWheelEvent } from 'react'
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

export function StrategicGalaxyView({ snapshot, selectedSystemID, onSelectSystem, onCloseSystem, onOpenColony, onPlanOrder, t }: {
  snapshot: PlayerSnapshot
  selectedSystemID?: number
  onSelectSystem: (systemID: number) => void
  onCloseSystem: () => void
  onOpenColony: (colonyID: number) => void
  onPlanOrder: (order: DraftOrder) => void
  t: Translator
}) {
  const decision = snapshot.decision
  const systems = decision?.strategic.galaxy.systems ?? []
  const [zoom, setZoom] = useState(1)
  const [pan, setPan] = useState({ x: 0, y: 0 })
  const [dragging, setDragging] = useState(false)
  const mapRef = useRef<HTMLDivElement | null>(null)
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
  const clampPan = (next: { x: number; y: number }, zoomValue = zoom) => {
    const map = mapRef.current
    if (!map) return next
    const width = Math.max(1, map.clientWidth)
    const height = Math.max(1, map.clientHeight)
    const edgeOffset = Math.max(18, Math.min(42, Math.min(width, height) * 0.06))
    const maxX = Math.max(edgeOffset, ((zoomValue - 1) * width) / 2 + edgeOffset)
    const maxY = Math.max(edgeOffset, ((zoomValue - 1) * height) / 2 + edgeOffset)
    const x = Math.max(-maxX, Math.min(maxX, next.x))
    const y = Math.max(-maxY, Math.min(maxY, next.y))
    return x === next.x && y === next.y ? next : { x, y }
  }

  useEffect(() => {
    const map = mapRef.current
    if (!map) return
    const enforceBounds = () => setPan((current) => clampPan(current, zoom))
    enforceBounds()
    const observer = new ResizeObserver(enforceBounds)
    observer.observe(map)
    return () => observer.disconnect()
  }, [zoom])

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
          setPan((current) => clampPan({ x: current.x + dx, y: current.y + dy }, zoom))
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
          setPan((current) => clampPan({ x: current.x + dx, y: current.y + dy }, zoom))
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
      <Card className="galaxy-card galaxy-card-full">

        <div
          ref={mapRef}
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
        <SystemDialog
          snapshot={snapshot}
          system={selected}
          onClose={onCloseSystem}
          onOpenColony={onOpenColony}
          onPlanOrder={onPlanOrder}
          t={t}
        />
      )}
    </>
  )
}

function SystemDialog({ snapshot, system, onClose, onOpenColony, onPlanOrder, t }: {
  snapshot: PlayerSnapshot
  system: StarSystem
  onClose: () => void
  onOpenColony: (colonyID: number) => void
  onPlanOrder: (order: DraftOrder) => void
  t: Translator
}) {
  const bodies = system.bodies ?? system.planets.map((planet) => ({
    id: planet.id,
    name: planet.name,
    orbit: planet.orbit,
    kind: 'planet' as const,
    planet_id: planet.id,
    outpost_id: undefined as number | undefined,
  }))
  const orderedBodies = [...bodies].sort((a, b) => a.orbit - b.orbit || a.id - b.id)
  const [selectedBodyID, setSelectedBodyID] = useState<number | null>(orderedBodies[0]?.id ?? null)
  const [selectedFleetID, setSelectedFleetID] = useState<number | null>(null)
  const [selectedShipID, setSelectedShipID] = useState<number | null>(null)

  useEffect(() => {
    const handleKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKey)
    document.body.classList.add('modal-open')
    return () => {
      window.removeEventListener('keydown', handleKey)
      document.body.classList.remove('modal-open')
    }
  }, [onClose])

  useEffect(() => {
    setSelectedBodyID(orderedBodies[0]?.id ?? null)
    setSelectedFleetID(null)
    setSelectedShipID(null)
  }, [system.id])

  const decision = snapshot.decision
  if (!decision) return null

  const fleets = decision.strategic.fleets?.filter((fleet) => fleet.at_system_id === system.id) ?? []
  const contacts = decision.strategic.contacts?.filter((contact) => contact.system_id === system.id) ?? []
  const selectedFleet = fleets.find((fleet) => fleet.id === selectedFleetID)
  const selectedBody = selectedFleet ? undefined : (orderedBodies.find((body) => body.id === selectedBodyID) ?? orderedBodies[0])
  const selectedPlanet = selectedBody?.planet_id ? system.planets.find((planet) => planet.id === selectedBody.planet_id) : undefined
  const selectedColony = selectedPlanet ? decision.colonies.find((colony) => colony.planet_id === selectedPlanet.id) : undefined
  const colonizeChoices = selectedPlanet
    ? (decision.decisions.colonization ?? []).filter((choice) => choice.system_id === system.id && choice.planet_id === selectedPlanet.id)
    : []
  const outpostChoices = selectedBody
    ? (decision.decisions.outpost_deployment ?? []).filter((choice) => choice.system_id === system.id && choice.body_id === selectedBody.id)
    : []
  const shipsByID = new Map((decision.strategic.ships ?? []).map((ship) => [ship.id, ship]))
  const fleetShips = selectedFleet?.ship_ids?.map((shipID) => shipsByID.get(shipID)).filter((ship): ship is NonNullable<typeof ship> => Boolean(ship)) ?? []
  const selectedShip = fleetShips.find((ship) => ship.id === selectedShipID) ?? fleetShips[0]

  const selectedStatus = selectedColony
    ? t('system.colonyStatus')
    : selectedBody?.outpost_id
      ? t('system.outpostStatus')
      : selectedBody?.kind === 'planet'
        ? t('system.uncolonized')
        : t('system.noSettlement')

  const selectBody = (bodyID: number) => {
    setSelectedFleetID(null)
    setSelectedShipID(null)
    setSelectedBodyID(bodyID)
  }
  const selectFleet = (fleetID: number) => {
    const fleet = fleets.find((item) => item.id === fleetID)
    setSelectedFleetID(fleetID)
    setSelectedBodyID(null)
    setSelectedShipID(fleet?.ship_ids?.[0] ?? null)
  }

  return (
    <div className="system-dialog-backdrop" role="presentation" onPointerDown={(event) => {
      if (event.target === event.currentTarget) onClose()
    }}>
      <section className="system-dialog system-dialog-classic" role="dialog" aria-modal="true" aria-labelledby={'system-dialog-title-' + system.id}>
        <header className="system-dialog-header system-dialog-classic-header">
          <div>
            <p className="eyebrow">{t('system.title', { system: system.name })}</p>
            <h2 id={'system-dialog-title-' + system.id}>{system.name}</h2>
          </div>
          <span className="badge system-star-class">{t('system.starClass', { class: system.spectral_class })}</span>
        </header>

        <div className="system-dialog-scene">
          <div className="system-orbit-stage system-orbit-stage-classic" aria-label={t('system.bodies')}>
            <div className="system-orbit-canvas">
              <div className={'system-star-core system-star-class-' + system.spectral_class} aria-hidden="true"><span /></div>
              {orderedBodies.map((body, index) => {
                const radius = 17 + ((index + 1) / (orderedBodies.length + 1)) * 31
                const angle = (((system.id * 31) + (body.id * 67) + (index * 103)) % 360) * Math.PI / 180
                const x = 50 + Math.cos(angle) * radius
                const y = 50 + Math.sin(angle) * radius
                const planet = body.planet_id ? system.planets.find((item) => item.id === body.planet_id) : undefined
                const selected = selectedBody?.id === body.id
                const bodyClasses = [
                  'system-orbit-body',
                  'system-orbit-body-' + body.kind.replace(/_/g, '-'),
                  planet?.climate_id ? 'system-orbit-body-climate-' + planet.climate_id.replace(/_/g, '-') : '',
                  planet?.size_id ? 'system-orbit-body-size-' + planet.size_id.replace(/_/g, '-') : '',
                  selected ? 'selected' : '',
                ].filter(Boolean).join(' ')
                return (
                  <div className="system-orbit-body-layer" key={'orbit-' + body.id}>
                    <span
                      className="system-orbit-ring"
                      style={{ width: (radius * 2) + '%', height: (radius * 2) + '%' }}
                      aria-hidden="true"
                    />
                    <button
                      type="button"
                      className={bodyClasses}
                      style={{ left: x + '%', top: y + '%' }}
                      title={body.name + ' - ' + bodyLabel(t, body.kind)}
                      aria-pressed={selected}
                      onClick={() => selectBody(body.id)}
                    >
                      <span className="system-orbit-body-reticle" aria-hidden="true">
                        <span className="system-orbit-body-dot" />
                      </span>
                      <strong>{body.name}</strong>
                    </button>
                  </div>
                )
              })}
            </div>

            {fleets.length > 0 && (
              <div className="system-fleet-dock" aria-label={t('system.systemOrbit')}>
                <span className="system-fleet-dock-label">{t('system.systemOrbit')}</span>
                {fleets.map((fleet) => {
                  const selected = selectedFleet?.id === fleet.id
                  return (
                    <button
                      type="button"
                      className={'system-fleet-marker' + (selected ? ' selected' : '')}
                      key={'fleet-marker-' + fleet.id}
                      aria-pressed={selected}
                      title={t('system.fleetMarkerTitle', { id: fleet.id })}
                      onClick={() => selectFleet(fleet.id)}
                    >
                      <span className="system-fleet-glyph" aria-hidden="true">▲</span>
                      <span><strong>{t('galaxy.fleet', { id: fleet.id })}</strong><small>{t('fleets.ships', { count: fleet.ship_ids?.length ?? 0 })}</small></span>
                    </button>
                  )
                })}
              </div>
            )}
          </div>

          {selectedBody && (
            <aside className="system-body-inspector" aria-live="polite">
              <header>
                <div>
                  <p className="eyebrow">{t('system.bodyDetails')}</p>
                  <h3>{selectedBody.name}</h3>
                </div>
                <span className="badge">{selectedBody.orbit}</span>
              </header>
              <dl className="system-body-facts">
                <div><dt>{t('system.kind')}</dt><dd>{bodyLabel(t, selectedBody.kind)}</dd></div>
                <div><dt>{t('system.orbit')}</dt><dd>{selectedBody.orbit}</dd></div>
                {selectedPlanet && <div><dt>{t('system.climate')}</dt><dd>{humanizeToken(selectedPlanet.climate_id)}</dd></div>}
                {selectedPlanet && <div><dt>{t('system.size')}</dt><dd>{humanizeToken(selectedPlanet.size_id)}</dd></div>}
                {selectedPlanet && <div><dt>{t('system.minerals')}</dt><dd>{humanizeToken(selectedPlanet.mineral_id)}</dd></div>}
                {selectedPlanet && <div><dt>{t('system.gravity')}</dt><dd>{humanizeToken(selectedPlanet.gravity_id)}</dd></div>}
                <div><dt>{t('system.status')}</dt><dd>{selectedStatus}</dd></div>
              </dl>
              <div className="system-body-status-row">
                {selectedColony && <span className="badge">{t('system.colonyStatus')}</span>}
                {selectedBody.outpost_id && <span className="badge">{t('system.outpostStatus')}</span>}
              </div>
            </aside>
          )}

          {selectedFleet && (
            <aside className="system-body-inspector system-fleet-inspector" aria-live="polite">
              <header>
                <div>
                  <p className="eyebrow">{t('system.fleetDetails')}</p>
                  <h3>{t('galaxy.fleet', { id: selectedFleet.id })}</h3>
                </div>
                <span className="badge">{t('system.systemOrbit')}</span>
              </header>
              <dl className="system-body-facts">
                <div><dt>{t('system.fleetRole')}</dt><dd>{humanizeToken(selectedFleet.role)}</dd></div>
                {selectedFleet.special_kind && <div><dt>{t('system.fleetType')}</dt><dd>{humanizeToken(selectedFleet.special_kind)}</dd></div>}
                <div><dt>{t('system.shipCount')}</dt><dd>{selectedFleet.ship_ids?.length ?? 0}</dd></div>
                <div><dt>{t('system.location')}</dt><dd>{t('system.systemLevelLocation')}</dd></div>
              </dl>
              {fleetShips.length > 0 && (
                <div className="system-fleet-ships">
                  {fleetShips.map((ship) => (
                    <button
                      type="button"
                      className={'system-ship-choice' + (selectedShip?.id === ship.id ? ' selected' : '')}
                      key={'system-ship-' + ship.id}
                      aria-pressed={selectedShip?.id === ship.id}
                      onClick={() => setSelectedShipID(ship.id)}
                    >
                      <span className="system-ship-glyph" aria-hidden="true">◆</span>
                      <span><strong>{ship.name}</strong><small>{humanizeToken(ship.spec.hull_id)}</small></span>
                    </button>
                  ))}
                </div>
              )}
              {selectedShip && (
                <div className="system-ship-details">
                  <strong>{selectedShip.name}</strong>
                  <span>{t('system.hull')}: {humanizeToken(selectedShip.spec.hull_id)}</span>
                  <span>{t('system.design')}: #{selectedShip.source_design_id} · r{selectedShip.source_design_revision}</span>
                </div>
              )}
              {(selectedFleet.ship_ids?.length ?? 0) > 0 && fleetShips.length === 0 && <p className="muted system-fleet-projection-note">{t('system.shipDetailsUnavailable')}</p>}
            </aside>
          )}
        </div>

        <footer className="system-dialog-footer">
          <div className="system-dialog-footer-left">
            {contacts.length > 0 && (
              <div className="system-traffic-strip" aria-label={t('system.fleets')}>
                {contacts.map((contact, index) => <span className="badge" key={'contact-' + index + '-' + contact.empire_id}>{t('common.empireFallback', { id: contact.empire_id })} · {contact.kind}</span>)}
              </div>
            )}
          </div>
          <div className="system-dialog-footer-actions">
            {selectedColony && <button type="button" className="button-primary" onClick={() => onOpenColony(selectedColony.id)}>{t('system.openColony')}</button>}
            {selectedBody && colonizeChoices.map((choice) => (
              <button
                type="button"
                className="button-primary"
                key={'colonize-' + choice.fleet_id}
                onClick={() => {
                  if (!window.confirm(t('system.colonizeConfirm', { body: selectedBody.name, fleet: choice.fleet_id }))) return
                  onPlanOrder({
                    key: 'fleet:' + choice.fleet_id,
                    kind: 'empire.colonize_planet',
                    payload: { fleet_id: choice.fleet_id, planet_id: choice.planet_id },
                  })
                }}
              >
                {t('system.colonize')}
              </button>
            ))}
            {selectedBody && outpostChoices.map((choice) => (
              <button
                type="button"
                className="button-secondary"
                key={'outpost-' + choice.fleet_id}
                onClick={() => {
                  if (!window.confirm(t('system.outpostConfirm', { body: selectedBody.name, fleet: choice.fleet_id }))) return
                  onPlanOrder({
                    key: 'fleet:' + choice.fleet_id,
                    kind: 'fleet.deploy_outpost',
                    payload: { fleet_id: choice.fleet_id, body_id: choice.body_id },
                  })
                }}
              >
                {t('system.buildOutpost')}
              </button>
            ))}
            <button type="button" className="button-secondary system-dialog-close" onClick={onClose}>{t('common.close')}</button>
          </div>
        </footer>
      </section>
    </div>
  )
}
export function StrategicColoniesView({ snapshot, preview, draftOrders, selectedColonyID, onOpenColony, onOpenConstruction, onBack, onPlanPopulation, onPlanOrder, onRemoveOrder, t }: {
  snapshot: PlayerSnapshot
  preview: PlanningPreviewSnapshot | null
  draftOrders: DraftOrder[]
  selectedColonyID?: number
  onOpenColony: (colonyID: number) => void
  onOpenConstruction: (colonyID: number) => void
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
        onOpenConstruction={onOpenConstruction}
        onPlanPopulation={onPlanPopulation}
        onPlanOrder={onPlanOrder}
        onRemoveOrder={onRemoveOrder}
        t={t}
      />
    )
  }

  return (
    <>
      <PageHeader eyebrow={t('colonies.eyebrow')} title={t('colonies.title')} />
      {colonies.length === 0 ? <EmptyState title={t('colonies.noColonies')} /> : (
        <div className="table-scroll colony-table-wrap">
          <table className="colony-table colony-table-dense">
            <thead>
              <tr>
                <th className="colony-sticky-column">{t('colonies.tableColony')}</th>
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
                    <td className="colony-sticky-column">
                      <button type="button" className="colony-name-button" onClick={() => onOpenColony(colony.id)}>
                        <strong>{t('colonies.colony', { id: colony.id })}</strong>
                        <small>{t('colonies.planet', { id: colony.planet_id })}</small>
                      </button>
                    </td>
                    <td className="colony-population-cell">
                      <div className="colony-population-total">{population.total.toFixed(1)} / {colony.population_dynamics.capacity.toFixed(1)}</div>
                      <PopulationMoveControls colony={colony} onPlan={onPlanPopulation} t={t} compact />
                    </td>
                    <td className="numeric-cell">{colony.adjusted_economy.food.toFixed(1)}</td>
                    <td className="numeric-cell">{colony.adjusted_economy.production.toFixed(1)}</td>
                    <td className="numeric-cell">{colony.adjusted_economy.research.toFixed(1)}</td>
                    <td className="colony-build-cell"><strong>{build?.project.project_id ?? colony.construction?.project_id ?? t('construction.empty')}</strong><small>{formatEta(t, build?.eta_turns)}</small></td>
                    <td><button type="button" className="button-secondary button-compact" onClick={() => onOpenColony(colony.id)}>{t('colonies.open')}</button></td>
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
  type Selection = { job: PopulationJob; indexes: number[] }
  type DragPayload = Selection & { amount: number }
  const [selected, setSelected] = useState<Selection | null>(null)
  const dragPayloadRef = useRef<DragPayload | null>(null)
  const pointerDragRef = useRef<(DragPayload & { pointerId: number; startX: number; startY: number; dragging: boolean }) | null>(null)
  const suppressClickRef = useRef(false)
  const population = aggregatePopulation(colony)
  const jobs: PopulationJob[] = ['farmer', 'worker', 'scientist']
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

  function markers(value: number): Array<{ amount: number; fractional: boolean }> {
    const whole = Math.floor(value + 0.0001)
    const fraction = Math.max(0, value - whole)
    const result = Array.from({ length: whole }, () => ({ amount: 1, fractional: false }))
    if (fraction > 0.01) result.push({ amount: fraction, fractional: true })
    return result
  }

  const markerMap: Record<PopulationJob, Array<{ amount: number; fractional: boolean }>> = {
    farmer: markers(values.farmer),
    worker: markers(values.worker),
    scientist: markers(values.scientist),
  }

  function amountFor(selection: Selection): number {
    return selection.indexes.reduce((sum, index) => sum + (markerMap[selection.job][index]?.amount ?? 0), 0)
  }

  function selectionFor(job: PopulationJob, index: number): Selection {
    if (selected?.job === job && selected.indexes.includes(index)) return selected
    return { job, indexes: [index] }
  }

  function toggleMarker(job: PopulationJob, index: number) {
    if (suppressClickRef.current) {
      suppressClickRef.current = false
      return
    }
    setSelected((current) => {
      if (!current || current.job !== job) return { job, indexes: [index] }
      const exists = current.indexes.includes(index)
      const indexes = exists ? current.indexes.filter((item) => item !== index) : [...current.indexes, index].sort((a, b) => a - b)
      return indexes.length > 0 ? { job, indexes } : null
    })
  }

  function moveAmount(source: PopulationJob, destination: PopulationJob, amount: number) {
    if (source === destination || amount <= 0 || values[source] + 0.0001 < amount) return
    const next = {
      ...values,
      [source]: values[source] - amount,
      [destination]: values[destination] + amount,
    }
    onPlan(colony.id, next.farmer, next.worker, next.scientist)
    setSelected(null)
    dragPayloadRef.current = null
    pointerDragRef.current = null
  }

  function moveSelection(destination: PopulationJob) {
    if (!selected) return
    moveAmount(selected.job, destination, amountFor(selected))
  }

  function handleDragStart(event: ReactDragEvent<HTMLButtonElement>, job: PopulationJob, index: number) {
    const selection = selectionFor(job, index)
    const payload = { ...selection, amount: amountFor(selection) }
    dragPayloadRef.current = payload
    if (selected !== selection) setSelected(selection)
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', job)
  }

  function handleDrop(event: ReactDragEvent<HTMLElement>, destination: PopulationJob) {
    event.preventDefault()
    const payload = dragPayloadRef.current
    if (!payload) return
    moveAmount(payload.job, destination, payload.amount)
  }

  function handlePointerDown(event: ReactPointerEvent<HTMLButtonElement>, job: PopulationJob, index: number) {
    if (event.pointerType === 'mouse') return
    const selection = selectionFor(job, index)
    const payload: DragPayload = { ...selection, amount: amountFor(selection) }
    pointerDragRef.current = { ...payload, pointerId: event.pointerId, startX: event.clientX, startY: event.clientY, dragging: false }
    if (selected !== selection) setSelected(selection)
    event.currentTarget.setPointerCapture(event.pointerId)
  }

  function handlePointerMove(event: ReactPointerEvent<HTMLButtonElement>) {
    const drag = pointerDragRef.current
    if (!drag || drag.pointerId !== event.pointerId) return
    if (Math.hypot(event.clientX - drag.startX, event.clientY - drag.startY) >= 8) {
      drag.dragging = true
      event.preventDefault()
    }
  }

  function finishPointerDrag(event: ReactPointerEvent<HTMLButtonElement>) {
    const drag = pointerDragRef.current
    if (!drag || drag.pointerId !== event.pointerId) return
    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId)
    if (drag.dragging) {
      const target = document.elementFromPoint(event.clientX, event.clientY)?.closest<HTMLElement>('[data-population-job]')
      const destination = target?.dataset.populationJob as PopulationJob | undefined
      if (destination) moveAmount(drag.job, destination, drag.amount)
      suppressClickRef.current = true
      window.setTimeout(() => { suppressClickRef.current = false }, 0)
    }
    pointerDragRef.current = null
  }

  const selectionAmount = selected ? amountFor(selected) : 0

  return (
    <div className={'population-board' + (compact ? ' population-board-compact' : '')}>
      {!compact && <p className="muted population-board-hint">{t('jobs.dragHint')}</p>}
      <div className="population-job-grid">
        {jobs.map((job) => {
          const sourceSelected = selected?.job === job
          const people = markerMap[job]
          return (
            <section
              className={'population-job population-job-' + job + (sourceSelected ? ' population-job-source' : '') + (selected && selected.job !== job ? ' population-job-target' : '')}
              key={job}
              data-population-job={job}
              onDragOver={(event) => event.preventDefault()}
              onDrop={(event) => handleDrop(event, job)}
            >
              <header>
                <span>{labels[job]}</span>
                <strong>{values[job].toFixed(1)}</strong>
              </header>
              <div className="population-people" aria-label={labels[job] + ': ' + values[job].toFixed(1)}>
                {people.length === 0 ? <span className="population-empty" aria-hidden="true">—</span> : people.map((person, index) => {
                  const isSelected = selected?.job === job && selected.indexes.includes(index)
                  return (
                    <button
                      type="button"
                      draggable
                      className={'population-person' + (person.fractional ? ' population-person-fractional' : '') + (isSelected ? ' population-person-selected' : '')}
                      key={job + '-' + index}
                      aria-pressed={isSelected}
                      aria-label={t('jobs.selectSource', { job: labels[job] })}
                      title={person.fractional ? person.amount.toFixed(1) : labels[job]}
                      onClick={() => toggleMarker(job, index)}
                      onDragStart={(event) => handleDragStart(event, job, index)}
                      onDragEnd={() => { dragPayloadRef.current = null }}
                      onPointerDown={(event) => handlePointerDown(event, job, index)}
                      onPointerMove={handlePointerMove}
                      onPointerUp={finishPointerDrag}
                      onPointerCancel={finishPointerDrag}
                    >
                      <span className="population-person-glyph" aria-hidden="true" />
                      {person.fractional && <small>{person.amount.toFixed(1)}</small>}
                    </button>
                  )
                })}
              </div>
              {selected && selected.job !== job && (
                <button type="button" className="population-destination" onClick={() => moveSelection(job)}>
                  {t('jobs.dropTo', { job: labels[job] })}
                </button>
              )}
            </section>
          )
        })}
      </div>
      {selected && (
        <div className="population-selection">
          <span>{t('jobs.selectedCount', { count: selected.indexes.length, amount: selectionAmount.toFixed(1), job: labels[selected.job] })}</span>
          <button type="button" className="button-ghost" onClick={() => setSelected(null)}>{t('common.cancel')}</button>
        </div>
      )}
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

function ColonyDetail({ snapshot, colony, preview, draftOrders, onBack, onOpenConstruction, onPlanPopulation, onPlanOrder, onRemoveOrder, t }: {
  snapshot: PlayerSnapshot
  colony: Colony
  preview?: PlanningPreviewSnapshot['preview']['projection']['colonies'][number]
  draftOrders: DraftOrder[]
  onBack: () => void
  onOpenConstruction: (colonyID: number) => void
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
        <ConstructionSummary
          colony={displayColony}
          preview={preview}
          draftOrders={draftOrders}
          onOpen={() => onOpenConstruction(colony.id)}
          t={t}
        />
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
            <div className="building-grid">{(displayColony.buildings ?? []).map((building) => <div className="building-tile" key={building}><span className="building-placeholder" aria-hidden="true" /><strong>{building}</strong></div>)}</div>
          )}
        </Card>

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

function ConstructionSummary({ colony, preview, draftOrders, onOpen, t }: {
  colony: Colony
  preview?: PlanningPreviewSnapshot['preview']['projection']['colonies'][number]
  draftOrders: DraftOrder[]
  onOpen: () => void
  t: Translator
}) {
  const draftKey = `construction:${colony.id}`
  const drafted = draftOrders.find((order) => order.key === draftKey)
  const draftedItems = drafted?.payload.items
  const items = Array.isArray(draftedItems) ? draftedItems as DraftQueueItem[] : queueItemsFromColony(colony)
  const current = items[0]
  const projected = preview?.construction?.[0]
  return (
    <Card className="construction-summary-card">
      <div className="card-heading construction-summary-heading">
        <div>
          <p className="eyebrow">{t('colony.construction')}</p>
          <h2>{current ? humanizeToken(current.project_id) : t('construction.empty')}</h2>
        </div>
        <span className="badge">{t('construction.queueCount', { count: items.length })}</span>
      </div>
      <div className="construction-summary-meta">
        <span>{projected ? formatEta(t, projected.eta_turns) : t('common.noEta')}</span>
        {items.length > 1 && <small>{items.slice(1, 4).map((item) => humanizeToken(item.project_id)).join(' · ')}{items.length > 4 ? ' …' : ''}</small>}
      </div>
      <button type="button" className="button-secondary button-wide" onClick={onOpen}>{t('construction.openManager')}</button>
    </Card>
  )
}

export function StrategicConstructionView({ snapshot, preview, draftOrders, colonyID, onBack, onPlanOrder, onRemoveOrder, t }: {
  snapshot: PlayerSnapshot
  preview: PlanningPreviewSnapshot | null
  draftOrders: DraftOrder[]
  colonyID: number
  onBack: () => void
  onPlanOrder: (order: DraftOrder) => void
  onRemoveOrder: (key: string) => void
  t: Translator
}) {
  const colonies = snapshot.decision?.colonies ?? snapshot.view.colonies
  const colony = colonies.find((item) => item.id === colonyID)
  if (!colony) return <EmptyState title={t('colonies.noColonies')} />
  const projected = preview?.preview.projection.colonies.find((item) => item.colony.id === colony.id)
  const displayColony = projected?.colony ?? colony
  const constructionDecision = snapshot.decision?.decisions.construction?.find((item) => item.colony_id === colony.id)
  return (
    <>
      <PageHeader
        eyebrow={t('colony.construction')}
        title={t('construction.manageTitle', { colony: colony.id })}
        subtitle={t('colonies.planet', { id: colony.planet_id })}
        actions={<button type="button" className="button-secondary" onClick={onBack}>{t('construction.backToColony')}</button>}
      />
      <ConstructionEditor
        colony={displayColony}
        preview={projected}
        choices={constructionDecision?.choices ?? []}
        draftOrders={draftOrders}
        onPlanOrder={onPlanOrder}
        onRemoveOrder={onRemoveOrder}
        t={t}
      />
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
    <div className="construction-workspace">
      <Card className="construction-available-card">
        <div className="card-heading">
          <div><p className="eyebrow">{t('construction.catalog')}</p><h2>{t('construction.available')}</h2></div>
          <span className="badge">{choices.length}</span>
        </div>
        <div className="choice-grid construction-choice-grid">
          {choices.map((choice, index) => {
            const nonRepeatable = choice.project_kind === 'building' || choice.project_kind === 'planetary_transformation'
            const alreadyQueued = nonRepeatable && items.some((item) => item.project_kind === choice.project_kind && item.project_id === choice.project_id)
            return (
              <button
                type="button"
                className="choice-button construction-choice"
                key={`${choice.project_kind}-${choice.project_id}-${index}`}
                disabled={alreadyQueued}
                onClick={() => save([...items, queueItemFromChoice(choice)])}
              >
                <strong>{humanizeToken(choice.ship_design_name ?? choice.project_id)}</strong>
                <small>{humanizeToken(choice.project_kind)} · {t('construction.cost', { pp: choice.production_cost_pp.toFixed(0) })}</small>
              </button>
            )
          })}
        </div>
      </Card>

      <Card className="construction-queue-card">
        <div className="card-heading">
          <div><p className="eyebrow">{t('colony.construction')}</p><h2>{t('construction.queue')}</h2></div>
          <span className="badge">{items.length}</span>
        </div>
        {items.length === 0 ? <p className="muted">{t('construction.empty')}</p> : (
          <div className="construction-queue-list">
            {items.map((item, index) => {
              const projectedItem = preview?.construction?.[index]
              return (
                <div className="queue-row construction-queue-row" key={`${item.project_kind}-${item.project_id}-${index}`}>
                  <span className="queue-position">{index + 1}</span>
                  <span className="queue-copy">
                    <strong>{humanizeToken(item.project_id)}</strong>
                    <small>{humanizeToken(item.project_kind)} · {projectedItem ? formatEta(t, projectedItem.eta_turns) : t('common.noEta')}</small>
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
        {drafted && <button type="button" className="button-secondary construction-cancel-draft" onClick={() => onRemoveOrder(draftKey)}>{t('construction.resetDraft')}</button>}
      </Card>
    </div>
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

export function StrategicResearchOverlay({ snapshot, preview, draftOrders, onPlanOrder, onClose, t }: {
  snapshot: PlayerSnapshot
  preview: PlanningPreviewSnapshot | null
  draftOrders: DraftOrder[]
  onPlanOrder: (order: DraftOrder) => void
  onClose: () => void
  t: Translator
}) {
  const decision = snapshot.decision
  const categories = [...(decision?.decisions.research_categories ?? [])].sort((a, b) => a.order - b.order)
  const choices = decision?.decisions.research ?? []
  const active = preview?.preview.projection.research
  const researchDraft = draftOrders.find((order) => order.key === 'research' && order.kind === 'empire.select_research')
  const draftFieldID = typeof researchDraft?.payload.tech_field_id === 'number' ? researchDraft.payload.tech_field_id : undefined
  const draftTechnologyID = typeof researchDraft?.payload.technology_id === 'number' ? researchDraft.payload.technology_id : undefined
  const activeFieldID = draftFieldID ?? active?.tech_field_id ?? decision?.empire.research?.tech_field_id
  const activeTechnologyIDs = new Set(
    draftTechnologyID !== undefined
      ? [draftTechnologyID]
      : activeFieldID === active?.tech_field_id
        ? (active?.technology_ids ?? [])
        : activeFieldID === decision?.empire.research?.tech_field_id
          ? (decision?.empire.research?.technology_ids ?? [])
          : [],
  )

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    document.body.classList.add('modal-open')
    window.addEventListener('keydown', onKey)
    return () => {
      document.body.classList.remove('modal-open')
      window.removeEventListener('keydown', onKey)
    }
  }, [onClose])

  function selectResearch(choice: ResearchChoice, technologyID = 0) {
    onPlanOrder({
      key: 'research',
      kind: 'empire.select_research',
      payload: {
        tech_field_id: choice.tech_field_id,
        ...(technologyID ? { technology_id: technologyID } : {}),
      },
    })
    onClose()
  }

  return (
    <div className="research-overlay-backdrop" role="presentation" onPointerDown={(event) => { if (event.target === event.currentTarget) onClose() }}>
      <section className="research-overlay" role="dialog" aria-modal="true" aria-label={t('research.changeTitle')}>
        <header className="research-overlay-header">
          <div>
            <p className="eyebrow">{t('research.eyebrow')}</p>
            <h2>{t('research.changeTitle')}</h2>
          </div>
          <div className="research-overlay-current">
            <span>{t('research.rpRate')}</span><strong>{active?.rp_per_turn.toFixed(1) ?? '—'} RP</strong>
            <span>{t('research.eta')}</span><strong>{formatEta(t, active?.eta_turns)}</strong>
          </div>
          <button type="button" className="button-ghost research-overlay-close" onClick={onClose} aria-label={t('common.close')}>×</button>
        </header>

        <div className="research-grid-classic">
          {categories.map((category) => {
            const choice = choices.find((item) => item.category_id === category.id)
            if (!choice) {
              return <section className="research-field-panel research-field-disabled" key={category.id}><div className="research-field-bar"><strong>{serverLabel(t, category.name_key, category.id)}</strong></div><p>{t('common.none')}</p></section>
            }
            const isActive = activeFieldID === choice.tech_field_id
            const fieldName = t('research.fieldNumber', { id: choice.tech_field_id })
            return (
              <section className={`research-field-panel${isActive ? ' active' : ''}`} key={category.id}>
                <div className="research-field-bar">
                  <strong>{serverLabel(t, category.name_key, category.id)}</strong>
                  <span>{choice.base_cost_rp.toFixed(0)} RP</span>
                </div>
                <div className="research-field-body">
                  <h3>{fieldName}</h3>
                  <div className="research-selection-mode">
                    {choice.selection_mode === 'choose_one'
                      ? t('research.modeChooseOne')
                      : choice.selection_mode === 'all'
                        ? t('research.modeAll')
                        : choice.selection_mode === 'fixed_one'
                          ? t('research.modeFixedOne')
                          : t('research.modeRepeatField')}
                  </div>
                  <div className="research-tech-list">
                    {choice.selection_mode === 'choose_one' ? choice.technology_ids.map((technologyID, index) => {
                      const selected = isActive && activeTechnologyIDs.has(technologyID)
                      const technologyName = serverLabel(t, choice.technology_name_keys[index], humanizeToken(choice.technology_keys[index] ?? String(technologyID)))
                      return (
                        <button
                          type="button"
                          className={`research-tech-choice${selected ? ' selected' : ''}`}
                          key={technologyID}
                          aria-pressed={selected}
                          onClick={() => selectResearch(choice, technologyID)}
                        >
                          {selected
                            ? <strong className="research-tech-name selected-name">{technologyName}</strong>
                            : <span className="research-tech-name">{technologyName}</span>}
                          {selected && <strong className="research-current-badge">✓ {t('research.currentChoice')}</strong>}
                        </button>
                      )
                    }) : (
                      <>
                        {choice.technology_ids.map((technologyID, index) => {
                          const selected = isActive && activeTechnologyIDs.has(technologyID)
                          return (
                            <span className={selected ? 'selected' : ''} key={technologyID}>
                              {serverLabel(t, choice.technology_name_keys[index], humanizeToken(choice.technology_keys[index] ?? String(technologyID)))}
                              {selected && <strong>{t('research.currentChoice')}</strong>}
                            </span>
                          )
                        })}
                        <button type="button" className="research-field-select" onClick={() => selectResearch(choice)}>{isActive ? t('research.reselect') : t('research.select')}</button>
                      </>
                    )}
                  </div>
                </div>
              </section>
            )
          })}
        </div>

        <footer className="research-overlay-footer">
          <button type="button" className="button-secondary" onClick={onClose}>{t('common.cancel')}</button>
        </footer>
      </section>
    </div>
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
