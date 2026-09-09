import { useEffect, useMemo, useRef, useState, type PointerEvent as ReactPointerEvent, type WheelEvent as ReactWheelEvent } from 'react'
import {
  tacticalEndActivationCommand,
  tacticalFireBeamCommand,
  tacticalMoveCommand,
  type BattleView,
  type ProtocolCommand,
  type TacticalFireAction,
  type TacticalShipView,
} from './api'
import { GameIcon } from './components/GameIcon'
import { Card, Notice } from './components/ui'
import { type TranslationKey, type TranslationVars } from './i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type TacticalBattlefieldProps = {
  battle: BattleView
  ownSeatID: number
  shipName: (shipID: number) => string
  empireName: (empireID: number) => string
  commandsDisabled: boolean
  onCommand: (command: ProtocolCommand) => Promise<void>
  onBack: () => void
  t: Translator
}

type TacticalMode = 'move' | 'fire' | 'scan'
type CameraState = { cx: number; cy: number; zoom: number }
type PointerPoint = { x: number; y: number }

type GestureSnapshot = {
  x: number
  y: number
  distance: number
}

const MIN_ZOOM = 0.45
const MAX_ZOOM = 7

function humanize(value: string | undefined) {
  if (!value) return '—'
  return value.replace(/[_-]+/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase())
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function gestureSnapshot(points: Map<number, PointerPoint>): GestureSnapshot | null {
  if (points.size === 0) return null
  const values = [...points.values()]
  const x = values.reduce((sum, point) => sum + point.x, 0) / values.length
  const y = values.reduce((sum, point) => sum + point.y, 0) / values.length
  if (values.length < 2) return { x, y, distance: 0 }
  const a = values[0]
  const b = values[1]
  return { x, y, distance: Math.hypot(a.x - b.x, a.y - b.y) }
}

function meterPercent(current: number, max: number) {
  if (max <= 0) return 0
  return clamp((current / max) * 100, 0, 100)
}

export function TacticalBattlefield({ battle, ownSeatID, shipName, empireName, commandsDisabled, onCommand, onBack, t }: TacticalBattlefieldProps) {
  const tactical = battle.tactical
  if (!tactical) return <Notice title={t('battle.tacticalUnavailableTitle')} tone="warning">{t('battle.noTacticalSpec')}</Notice>

  const ships = tactical.ships ?? []
  const legalMoves = tactical.legal_moves ?? []
  const legalFireActions = tactical.legal_fire_actions ?? []
  const activeShip = ships.find((ship) => ship.ship_id === tactical.state.active_ship_id)
  const ownActivation = activeShip?.seat_id === ownSeatID
  const [mode, setMode] = useState<TacticalMode>(() => legalMoves.length > 0 ? 'move' : 'scan')
  const [selectedWeaponSlot, setSelectedWeaponSlot] = useState<number | null>(() => legalFireActions[0]?.weapon_slot ?? null)
  const [scannedShipID, setScannedShipID] = useState<number | null>(() => activeShip?.ship_id ?? ships[0]?.ship_id ?? null)
  const [busy, setBusy] = useState(false)

  const points = useMemo(() => {
    const shipPoints = ships.map((ship) => ({ x: ship.x, y: ship.y }))
    const movePoints = legalMoves.map((move) => ({ x: move.x, y: move.y }))
    return shipPoints.concat(movePoints)
  }, [ships, legalMoves])

  const bounds = useMemo(() => {
    if (points.length === 0) return { minX: -12, maxX: 12, minY: -8, maxY: 8, cx: 0, cy: 0, width: 36, height: 22 }
    const xs = points.map((point) => point.x)
    const ys = points.map((point) => point.y)
    const minX = Math.min(...xs)
    const maxX = Math.max(...xs)
    const minY = Math.min(...ys)
    const maxY = Math.max(...ys)
    return {
      minX,
      maxX,
      minY,
      maxY,
      cx: (minX + maxX) / 2,
      cy: (minY + maxY) / 2,
      width: Math.max(36, maxX - minX + 12),
      height: Math.max(22, maxY - minY + 12),
    }
  }, [points])

  const [camera, setCamera] = useState<CameraState>(() => ({ cx: bounds.cx, cy: bounds.cy, zoom: 1 }))
  const pointersRef = useRef(new Map<number, PointerPoint>())
  const gestureRef = useRef<GestureSnapshot | null>(null)

  useEffect(() => {
    if (!ownActivation && mode !== 'scan') setMode('scan')
  }, [ownActivation, mode])

  useEffect(() => {
    if (selectedWeaponSlot == null || !legalFireActions.some((action) => action.weapon_slot === selectedWeaponSlot)) {
      setSelectedWeaponSlot(legalFireActions[0]?.weapon_slot ?? null)
    }
  }, [legalFireActions, selectedWeaponSlot])

  useEffect(() => {
    if (scannedShipID != null && !ships.some((ship) => ship.ship_id === scannedShipID)) setScannedShipID(activeShip?.ship_id ?? ships[0]?.ship_id ?? null)
  }, [ships, scannedShipID, activeShip?.ship_id])

  const selectedFireAction = legalFireActions.find((action) => action.weapon_slot === selectedWeaponSlot) ?? legalFireActions[0]
  const legalTargetByID = useMemo(() => new Map((selectedFireAction?.targets ?? []).map((target) => [target.target_ship_id, target])), [selectedFireAction])
  const scannedShip = ships.find((ship) => ship.ship_id === scannedShipID)
  const viewWidth = bounds.width / camera.zoom
  const viewHeight = bounds.height / camera.zoom
  const viewX = camera.cx - viewWidth / 2
  const viewY = camera.cy - viewHeight / 2
  const controlsDisabled = commandsDisabled || busy || !ownActivation

  const recenter = () => setCamera({ cx: bounds.cx, cy: bounds.cy, zoom: 1 })
  const zoomBy = (factor: number) => setCamera((current) => ({ ...current, zoom: clamp(current.zoom * factor, MIN_ZOOM, MAX_ZOOM) }))

  const runCommand = async (command: ProtocolCommand) => {
    if (commandsDisabled || busy) return
    setBusy(true)
    try {
      await onCommand(command)
    } finally {
      setBusy(false)
    }
  }

  const moveTo = (x: number, y: number) => {
    if (mode !== 'move' || controlsDisabled || !activeShip) return
    void runCommand(tacticalMoveCommand(tactical, activeShip.ship_id, x, y))
  }

  const selectShip = (ship: TacticalShipView) => {
    if (mode === 'scan') {
      setScannedShipID(ship.ship_id)
      return
    }
    if (mode !== 'fire' || controlsDisabled || !activeShip || !selectedFireAction) return
    if (!legalTargetByID.has(ship.ship_id)) return
    void runCommand(tacticalFireBeamCommand(tactical, activeShip.ship_id, ship.ship_id, selectedFireAction.weapon_slot))
  }

  const endActivation = () => {
    if (controlsDisabled || !tactical.can_end_activation || !activeShip) return
    void runCommand(tacticalEndActivationCommand(tactical, activeShip.ship_id))
  }

  const pointerDown = (event: ReactPointerEvent<SVGSVGElement>) => {
    event.currentTarget.setPointerCapture(event.pointerId)
    pointersRef.current.set(event.pointerId, { x: event.clientX, y: event.clientY })
    gestureRef.current = gestureSnapshot(pointersRef.current)
  }

  const pointerMove = (event: ReactPointerEvent<SVGSVGElement>) => {
    if (!pointersRef.current.has(event.pointerId)) return
    const previous = gestureRef.current
    pointersRef.current.set(event.pointerId, { x: event.clientX, y: event.clientY })
    const currentGesture = gestureSnapshot(pointersRef.current)
    if (!previous || !currentGesture) {
      gestureRef.current = currentGesture
      return
    }
    const rect = event.currentTarget.getBoundingClientRect()
    if (rect.width <= 0 || rect.height <= 0) return
    const dx = currentGesture.x - previous.x
    const dy = currentGesture.y - previous.y
    const pinchFactor = previous.distance > 0 && currentGesture.distance > 0 ? currentGesture.distance / previous.distance : 1
    setCamera((current) => ({
      cx: current.cx - dx * (bounds.width / current.zoom) / rect.width,
      cy: current.cy - dy * (bounds.height / current.zoom) / rect.height,
      zoom: clamp(current.zoom * pinchFactor, MIN_ZOOM, MAX_ZOOM),
    }))
    gestureRef.current = currentGesture
  }

  const pointerEnd = (event: ReactPointerEvent<SVGSVGElement>) => {
    pointersRef.current.delete(event.pointerId)
    gestureRef.current = gestureSnapshot(pointersRef.current)
  }

  const wheel = (event: ReactWheelEvent<SVGSVGElement>) => {
    event.preventDefault()
    zoomBy(Math.exp(-event.deltaY * 0.0015))
  }

  return (
    <div className="tactical-battlefield">
      <div className="tactical-battlefield-topbar">
        <div>
          <p className="eyebrow">{t('battlefield.eyebrow')}</p>
          <h3>{t('battlefield.title')}</h3>
          <p className="muted">{t('battlefield.roundContext', { round: tactical.state.round, sequence: tactical.state.next_command_sequence })}</p>
        </div>
        <div className="tactical-camera-controls" aria-label={t('battlefield.cameraControls')}>
          <button type="button" className="button-secondary tactical-icon-button" onClick={() => zoomBy(1.3)} aria-label={t('battlefield.zoomIn')}>+</button>
          <button type="button" className="button-secondary tactical-icon-button" onClick={() => zoomBy(1 / 1.3)} aria-label={t('battlefield.zoomOut')}>−</button>
          <button type="button" className="button-secondary" onClick={recenter}>{t('battlefield.recenter')}</button>
        </div>
      </div>

      <div className="tactical-modebar" role="group" aria-label={t('battlefield.modeControls')}>
        <button type="button" className={mode === 'move' ? 'button-primary' : 'button-secondary'} disabled={!ownActivation || legalMoves.length === 0 || commandsDisabled || busy} onClick={() => setMode('move')}><GameIcon name="arrow-up" />{t('battlefield.move')}</button>
        <button type="button" className={mode === 'fire' ? 'button-primary' : 'button-secondary'} disabled={!ownActivation || legalFireActions.length === 0 || commandsDisabled || busy} onClick={() => setMode('fire')}><GameIcon name="fleet-combat" />{t('battlefield.fire')}</button>
        <button type="button" className={mode === 'scan' ? 'button-primary' : 'button-secondary'} onClick={() => { setMode('scan'); setScannedShipID(activeShip?.ship_id ?? scannedShipID) }}><GameIcon name="info" />{t('battlefield.scan')}</button>
      </div>

      <div className="tactical-layout">
        <div className="tactical-viewport-shell">
          <svg
            className="tactical-viewport"
            viewBox={`${viewX} ${viewY} ${viewWidth} ${viewHeight}`}
            onPointerDown={pointerDown}
            onPointerMove={pointerMove}
            onPointerUp={pointerEnd}
            onPointerCancel={pointerEnd}
            onWheel={wheel}
            aria-label={t('battlefield.viewportLabel')}
          >
            <defs>
              <pattern id={`tactical-grid-${battle.spec.id}`} width="2" height="2" patternUnits="userSpaceOnUse">
                <path d="M 2 0 L 0 0 0 2" className="tactical-grid-line" />
              </pattern>
            </defs>
            <rect x={viewX} y={viewY} width={viewWidth} height={viewHeight} fill={`url(#tactical-grid-${battle.spec.id})`} className="tactical-space" />

            {mode === 'move' && ownActivation && legalMoves.map((move) => (
              <g key={`${move.x}:${move.y}`} className="tactical-move-node" onPointerDown={(event) => event.stopPropagation()} onClick={() => moveTo(move.x, move.y)} role="button" tabIndex={0} aria-label={t('battlefield.moveTarget', { x: move.x, y: move.y, cost: move.move_cost })} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') moveTo(move.x, move.y) }}>
                <circle cx={move.x} cy={move.y} r={0.34} />
                <circle cx={move.x} cy={move.y} r={0.72} className="tactical-move-hit" />
              </g>
            ))}

            {ships.map((ship) => {
              const isActive = ship.ship_id === tactical.state.active_ship_id
              const isOwn = ship.seat_id === ownSeatID
              const isLegalTarget = mode === 'fire' && legalTargetByID.has(ship.ship_id)
              const isScanned = mode === 'scan' && scannedShipID === ship.ship_id
              const angle = ship.facing * 22.5
              return (
                <g
                  key={ship.ship_id}
                  className={`tactical-ship ${isOwn ? 'own' : 'enemy'} ${isActive ? 'active' : ''} ${isLegalTarget ? 'legal-target' : ''} ${isScanned ? 'scanned' : ''} ${ship.destroyed ? 'destroyed' : ''}`}
                  transform={`translate(${ship.x} ${ship.y}) rotate(${angle})`}
                  onPointerDown={(event) => event.stopPropagation()}
                  onClick={() => selectShip(ship)}
                  role="button"
                  tabIndex={0}
                  aria-label={t('battlefield.shipMarker', { name: shipName(ship.ship_id), x: ship.x, y: ship.y })}
                  onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') selectShip(ship) }}
                >
                  {isActive && <circle r={1.25} className="tactical-active-ring" />}
                  {isLegalTarget && <circle r={1.05} className="tactical-target-ring" />}
                  {isScanned && <circle r={0.95} className="tactical-scan-ring" />}
                  <path d="M 0.82 0 L -0.58 -0.5 L -0.32 0 L -0.58 0.5 Z" className="tactical-ship-hull" />
                  <circle cx={-0.48} cy={0} r={0.15} className="tactical-ship-engine" />
                  <text x={0} y={1.6} textAnchor="middle" className="tactical-ship-label" transform={`rotate(${-angle})`}>{shipName(ship.ship_id)}</text>
                </g>
              )
            })}
          </svg>
          <div className="tactical-viewport-hint">{t('battlefield.panZoomHint')}</div>
        </div>

        <aside className="tactical-command-panel">
          <Card className="tactical-active-card" as="section">
            <p className="eyebrow">{t('battlefield.activeShip')}</p>
            {activeShip ? (
              <>
                <h4>{shipName(activeShip.ship_id)}</h4>
                <p className="muted">{ownActivation ? t('battlefield.yourActivation') : t('battlefield.waitingActivation', { empire: empireName(activeShip.empire_id) })}</p>
                <dl className="tactical-quick-stats">
                  <div><dt>{t('battlefield.movement')}</dt><dd>{activeShip.movement_current}/{activeShip.movement_max}</dd></div>
                  <div><dt>{t('battlefield.facing')}</dt><dd>{activeShip.facing}/15</dd></div>
                  <div><dt>{t('battlefield.position')}</dt><dd>{activeShip.x}, {activeShip.y}</dd></div>
                </dl>
              </>
            ) : <p className="muted">{t('battlefield.noActiveShip')}</p>}
          </Card>

          {mode === 'fire' && (
            <Card className="tactical-action-card" as="section">
              <p className="eyebrow">{t('battlefield.weaponAction')}</p>
              {legalFireActions.length > 0 ? legalFireActions.map((action) => (
                <button key={action.weapon_slot} type="button" className={selectedWeaponSlot === action.weapon_slot ? 'button-primary' : 'button-secondary'} disabled={commandsDisabled || busy} onClick={() => setSelectedWeaponSlot(action.weapon_slot)}>
                  {humanize(action.weapon_id)} · {t('battlefield.targets', { count: action.targets.length })}
                </button>
              )) : <p className="muted">{t('battlefield.noWeapon')}</p>}
              {selectedFireAction && <div className="tactical-target-list">{selectedFireAction.targets.map((target) => <button key={target.target_ship_id} type="button" className="button-secondary" disabled={controlsDisabled} onClick={() => { const ship = ships.find((candidate) => candidate.ship_id === target.target_ship_id); if (ship) selectShip(ship) }}>{shipName(target.target_ship_id)} · {t('battlefield.rangeIndex', { range: target.range_index })}</button>)}</div>}
            </Card>
          )}

          {mode === 'scan' && (
            <Card className="tactical-scan-card" as="section">
              <p className="eyebrow">{t('battlefield.scan')}</p>
              {scannedShip ? <TacticalScanDetails ship={scannedShip} shipName={shipName} empireName={empireName} t={t} /> : <p className="muted">{t('battlefield.scanHint')}</p>}
            </Card>
          )}

          <div className="tactical-primary-actions">
            <button type="button" className="button-primary" disabled={!tactical.can_end_activation || controlsDisabled} onClick={endActivation}>{busy ? t('battlefield.commandBusy') : t('battlefield.endActivation')}</button>
            <button type="button" className="button-secondary" onClick={onBack}>{t('battle.backToEncounter')}</button>
          </div>
        </aside>
      </div>
    </div>
  )
}

function TacticalScanDetails({ ship, shipName, empireName, t }: { ship: TacticalShipView; shipName: (shipID: number) => string; empireName: (empireID: number) => string; t: Translator }) {
  return (
    <div className="tactical-scan-details">
      <div>
        <h4>{shipName(ship.ship_id)}</h4>
        <p className="muted">{empireName(ship.empire_id)} · {humanize(ship.hull_id)} · {humanize(ship.warp_drive_id)}</p>
      </div>
      <dl className="tactical-scan-facts">
        <div><dt>{t('battlefield.position')}</dt><dd>{ship.x}, {ship.y}</dd></div>
        <div><dt>{t('battlefield.facing')}</dt><dd>{ship.facing}/15</dd></div>
        <div><dt>{t('battlefield.movement')}</dt><dd>{ship.movement_current}/{ship.movement_max}</dd></div>
        <div><dt>{t('battlefield.beamOffense')}</dt><dd>{ship.beam_offense}</dd></div>
        <div><dt>{t('battlefield.beamDefense')}</dt><dd>{ship.beam_defense}</dd></div>
      </dl>
      <div className="tactical-health-row">
        <span>{t('battlefield.armor')}</span><strong>{ship.armor_current}/{ship.armor_max}</strong>
        <div className="tactical-health-meter"><span style={{ width: `${meterPercent(ship.armor_current, ship.armor_max)}%` }} /></div>
      </div>
      <div className="tactical-health-row">
        <span>{t('battlefield.structure')}</span><strong>{ship.structure_current}/{ship.structure_max}</strong>
        <div className="tactical-health-meter"><span style={{ width: `${meterPercent(ship.structure_current, ship.structure_max)}%` }} /></div>
      </div>
      <div className="tactical-scan-weapons">
        <strong>{t('battlefield.weapons')}</strong>
        {ship.weapons?.length ? ship.weapons.map((weapon) => (
          <div className="tactical-weapon-row" key={weapon.slot}>
            <span>{humanize(weapon.weapon_id)} × {weapon.count}</span>
            <span>{weapon.min_damage}–{weapon.max_damage}</span>
            <span className={`badge ${weapon.ready ? '' : 'danger'}`}>{weapon.ready ? t('battlefield.ready') : t('battlefield.spent')}</span>
          </div>
        )) : <p className="muted">{t('battlefield.unarmed')}</p>}
      </div>
    </div>
  )
}
