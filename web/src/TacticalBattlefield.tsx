import { useEffect, useMemo, useRef, useState, type PointerEvent as ReactPointerEvent, type WheelEvent as ReactWheelEvent } from 'react'
import {
  tacticalEndActivationCommand,
  tacticalFireBeamCommand,
  tacticalMoveCommand,
  tacticalRetreatCommand,
  type BattleView,
  type ProtocolCommand,
  type TacticalShipView,
} from './api'
import { GameIcon } from './components/GameIcon'
import { ProceduralShipGlyph } from './components/ProceduralShipGlyph'
import { Notice } from './components/ui'
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

type TacticalMode = 'combat' | 'scan'
type CameraState = { cx: number; cy: number; zoom: number }
type PointerPoint = { x: number; y: number }
type GestureSnapshot = { x: number; y: number; distance: number }

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
  const ownShips = ships.filter((ship) => ship.seat_id === ownSeatID)
  const legalMoves = tactical.legal_moves ?? []
  const legalFireActions = tactical.legal_fire_actions ?? []
  const activeShip = ships.find((ship) => ship.ship_id === tactical.state.active_ship_id)
  const ownActivation = activeShip?.seat_id === ownSeatID
  const [mode, setMode] = useState<TacticalMode>('combat')
  const [selectedWeaponSlot, setSelectedWeaponSlot] = useState<number | null>(() => legalFireActions[0]?.weapon_slot ?? null)
  const [scannedShipID, setScannedShipID] = useState<number | null>(null)
  const [busy, setBusy] = useState(false)
  const [commandError, setCommandError] = useState('')

  const points = useMemo(() => {
    const shipPoints = ships.map((ship) => ({ x: ship.x, y: ship.y }))
    const movePoints = legalMoves.map((move) => ({ x: move.x, y: move.y }))
    return shipPoints.concat(movePoints)
  }, [ships, legalMoves])

  const bounds = useMemo(() => {
    if (points.length === 0) return { minX: -16, maxX: 16, minY: -10, maxY: 10, cx: 0, cy: 0, width: 44, height: 28 }
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
      width: Math.max(44, maxX - minX + 16),
      height: Math.max(28, maxY - minY + 16),
    }
  }, [points])

  const [camera, setCamera] = useState<CameraState>(() => ({ cx: bounds.cx, cy: bounds.cy, zoom: 1 }))
  const pointersRef = useRef(new Map<number, PointerPoint>())
  const gestureRef = useRef<GestureSnapshot | null>(null)
  const draggedRef = useRef(false)
  const ignoreClickRef = useRef(false)


  useEffect(() => {
    if (selectedWeaponSlot == null || !legalFireActions.some((action) => action.weapon_slot === selectedWeaponSlot)) {
      setSelectedWeaponSlot(legalFireActions[0]?.weapon_slot ?? null)
    }
  }, [legalFireActions, selectedWeaponSlot])

  useEffect(() => {
    if (scannedShipID != null && !ships.some((ship) => ship.ship_id === scannedShipID)) setScannedShipID(null)
  }, [ships, scannedShipID])

  const selectedFireAction = legalFireActions.find((action) => action.weapon_slot === selectedWeaponSlot) ?? legalFireActions[0]
  const legalTargetByID = useMemo(() => new Map((selectedFireAction?.targets ?? []).map((target) => [target.target_ship_id, target])), [selectedFireAction])
  const scannedShip = ships.find((ship) => ship.ship_id === scannedShipID)
  const viewWidth = bounds.width / camera.zoom
  const viewHeight = bounds.height / camera.zoom
  const viewX = camera.cx - viewWidth / 2
  const viewY = camera.cy - viewHeight / 2
  const controlsDisabled = commandsDisabled || busy || !ownActivation

  const zoomBy = (factor: number) => setCamera((current) => ({ ...current, zoom: clamp(current.zoom * factor, MIN_ZOOM, MAX_ZOOM) }))
  const focusShip = (ship: TacticalShipView) => setCamera((current) => ({ cx: ship.x, cy: ship.y, zoom: Math.max(current.zoom, 1.45) }))

  const runCommand = async (command: ProtocolCommand) => {
    if (commandsDisabled || busy) return
    setBusy(true)
    setCommandError('')
    try {
      await onCommand(command)
    } catch (reason) {
      setCommandError(reason instanceof Error ? reason.message : String(reason))
    } finally {
      setBusy(false)
    }
  }

  const moveTo = (x: number, y: number) => {
    if (ignoreClickRef.current || mode === 'scan' || controlsDisabled || !activeShip) return
    void runCommand(tacticalMoveCommand(tactical, activeShip.ship_id, x, y))
  }

  const selectShip = (ship: TacticalShipView) => {
    if (ignoreClickRef.current) return
    if (mode === 'scan') {
      setScannedShipID(ship.ship_id)
      focusShip(ship)
      return
    }
    if (ship.seat_id === ownSeatID) {
      focusShip(ship)
      return
    }
    if (controlsDisabled || !activeShip || !selectedFireAction) return
    if (!legalTargetByID.has(ship.ship_id)) return
    void runCommand(tacticalFireBeamCommand(tactical, activeShip.ship_id, ship.ship_id, selectedFireAction.weapon_slot))
  }

  const endActivation = () => {
    if (controlsDisabled || !tactical.can_end_activation || !activeShip) return
    void runCommand(tacticalEndActivationCommand(tactical, activeShip.ship_id))
  }

  const retreat = () => {
    if (controlsDisabled || !activeShip) return
    void runCommand(tacticalRetreatCommand(tactical, activeShip.ship_id))
  }

  const pointerDown = (event: ReactPointerEvent<SVGSVGElement>) => {
    if (event.pointerType === 'mouse' && event.button !== 0 && event.button !== 2) return
    if (event.pointerType === 'mouse' && event.button === 2) event.preventDefault()
    if (pointersRef.current.size === 0) {
      draggedRef.current = false
      ignoreClickRef.current = false
    }
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
    const moved = Math.abs(dx) + Math.abs(dy) > 0.5 || Math.abs(pinchFactor - 1) > 0.002
    if (moved) {
      draggedRef.current = true
      ignoreClickRef.current = true
      if (!event.currentTarget.hasPointerCapture(event.pointerId)) {
        try { event.currentTarget.setPointerCapture(event.pointerId) } catch { /* synthetic/ended pointer: capture is optional */ }
      }
    }
    setCamera((current) => ({
      cx: current.cx - dx * (bounds.width / current.zoom) / rect.width,
      cy: current.cy - dy * (bounds.height / current.zoom) / rect.height,
      zoom: clamp(current.zoom * pinchFactor, MIN_ZOOM, MAX_ZOOM),
    }))
    gestureRef.current = currentGesture
  }

  const pointerEnd = (event: ReactPointerEvent<SVGSVGElement>) => {
    pointersRef.current.delete(event.pointerId)
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      try { event.currentTarget.releasePointerCapture(event.pointerId) } catch { /* capture already ended */ }
    }
    gestureRef.current = gestureSnapshot(pointersRef.current)
    if (pointersRef.current.size === 0 && draggedRef.current) {
      window.setTimeout(() => { ignoreClickRef.current = false }, 0)
    }
  }

  const wheel = (event: ReactWheelEvent<SVGSVGElement>) => {
    event.preventDefault()
    zoomBy(Math.exp(-event.deltaY * 0.0015))
  }

  return (
    <div className="tactical-battlefield tactical-screen">
      <div className="tactical-screen-status" aria-live="polite">
        <span className={`tactical-status-dot ${ownActivation ? 'is-own' : 'is-waiting'}`} />
        <strong>{activeShip ? shipName(activeShip.ship_id) : t('battlefield.noActiveShip')}</strong>
        {activeShip && <span>{ownActivation ? t('battlefield.yourActivation') : t('battlefield.waitingActivation', { empire: empireName(activeShip.empire_id) })}</span>}
      </div>


      {commandError && <div className="tactical-command-error"><Notice title={t('battlefield.commandRejectedTitle')} tone="warning">{commandError} · {t('battlefield.commandRejectedBody')}</Notice></div>}

      <div className="tactical-viewport-shell">
        <svg
          className="tactical-viewport"
          viewBox={`${viewX} ${viewY} ${viewWidth} ${viewHeight}`}
          onPointerDown={pointerDown}
          onPointerMove={pointerMove}
          onPointerUp={pointerEnd}
          onPointerCancel={pointerEnd}
          onWheel={wheel}
          onContextMenu={(event) => event.preventDefault()}
          aria-label={t('battlefield.viewportLabel')}
        >
          <defs>
            <pattern id={`tactical-grid-minor-${battle.spec.id}`} width="1" height="1" patternUnits="userSpaceOnUse">
              <path d="M 1 0 L 0 0 0 1" className="tactical-grid-line tactical-grid-line-minor" />
            </pattern>
            <pattern id={`tactical-grid-major-${battle.spec.id}`} width="4" height="4" patternUnits="userSpaceOnUse">
              <rect width="4" height="4" fill={`url(#tactical-grid-minor-${battle.spec.id})`} />
              <path d="M 4 0 L 0 0 0 4" className="tactical-grid-line tactical-grid-line-major" />
            </pattern>
          </defs>
          <rect x={viewX} y={viewY} width={viewWidth} height={viewHeight} fill={`url(#tactical-grid-major-${battle.spec.id})`} className="tactical-space" />

          {mode !== 'scan' && ownActivation && legalMoves.map((move) => (
            <g key={`${move.x}:${move.y}`} className="tactical-move-cell" onClick={() => moveTo(move.x, move.y)} role="button" tabIndex={0} aria-label={t('battlefield.moveTarget', { x: move.x, y: move.y, cost: move.move_cost })} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') moveTo(move.x, move.y) }}>
              <rect x={move.x - 0.48} y={move.y - 0.48} width={0.96} height={0.96} rx={0.08} />
            </g>
          ))}

          {ships.map((ship) => {
            const isActive = ship.ship_id === tactical.state.active_ship_id
            const isOwn = ship.seat_id === ownSeatID
            const isLegalTarget = mode !== 'scan' && legalTargetByID.has(ship.ship_id)
            const isScanned = mode === 'scan' && scannedShipID === ship.ship_id
            const angle = ship.facing * 22.5
            return (
              <g
                key={ship.ship_id}
                className={`tactical-ship ${isOwn ? 'own' : 'enemy'} ${isActive ? 'active' : ''} ${isLegalTarget ? 'legal-target' : ''} ${isScanned ? 'scanned' : ''} ${ship.destroyed ? 'destroyed' : ''}`}
                transform={`translate(${ship.x} ${ship.y}) rotate(${angle})`}
                onClick={() => selectShip(ship)}
                role="button"
                tabIndex={0}
                aria-label={t('battlefield.shipMarker', { name: shipName(ship.ship_id), x: ship.x, y: ship.y })}
                onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') selectShip(ship) }}
              >
                {isActive && <circle r={1.25} className="tactical-active-ring" />}
                {isLegalTarget && <circle r={1.08} className="tactical-target-ring" />}
                {isScanned && <circle r={0.98} className="tactical-scan-ring" />}
                <path d="M 0.82 0 L -0.58 -0.5 L -0.32 0 L -0.58 0.5 Z" className="tactical-ship-fallback" />
                <foreignObject x={-0.92} y={-0.64} width={1.84} height={1.28} className="tactical-ship-art" pointerEvents="none">
                  <div className="tactical-procedural-ship">
                    <ProceduralShipGlyph
                      seed={`tactical|empire:${ship.empire_id}|ship:${ship.ship_id}`}
                      hullId={ship.hull_id}
                      weaponCount={(ship.weapons ?? []).reduce((sum, weapon) => sum + weapon.count, 0)}
                      footprint={1}
                    />
                  </div>
                </foreignObject>
                <text x={0} y={1.6} textAnchor="middle" className="tactical-ship-label" transform={`rotate(${-angle})`}>{shipName(ship.ship_id)}</text>
              </g>
            )
          })}
        </svg>
        <div className="tactical-viewport-hint">{t('battlefield.panZoomHint')}</div>
      </div>

      {mode === 'scan' && scannedShip && (
        <div className="tactical-scan-layer" role="presentation" onPointerDown={(event) => { if (event.target === event.currentTarget) setScannedShipID(null) }}>
          <section className="galaxy-fleet-info-popover tactical-scan-popover" role="dialog" aria-label={`${t('battlefield.scan')}: ${shipName(scannedShip.ship_id)}`}>
            <header className="galaxy-fleet-info-header">
              <div><p className="eyebrow">{t('battlefield.scan')}</p><strong>{shipName(scannedShip.ship_id)}</strong></div>
              <button type="button" className="galaxy-fleet-info-close" aria-label={t('common.close')} title={t('common.close')} onClick={() => setScannedShipID(null)}><GameIcon name="close" /></button>
            </header>
            <TacticalScanDetails ship={scannedShip} shipName={shipName} empireName={empireName} t={t} />
          </section>
        </div>
      )}

      <footer className="tactical-hud">
        <section className="tactical-hud-roster" aria-label={t('battlefield.ownShips')}>
          <span className="tactical-hud-label">{t('battlefield.ownShips')}</span>
          <div className="tactical-ship-strip">
            {ownShips.map((ship) => {
              const active = ship.ship_id === tactical.state.active_ship_id
              return (
                <button
                  type="button"
                  key={ship.ship_id}
                  className={`tactical-roster-ship ${active ? 'is-active' : ''} ${ship.activation_complete ? 'is-complete' : ''} ${ship.destroyed ? 'is-destroyed' : ''}`}
                  aria-current={active ? 'true' : undefined}
                  onClick={() => { focusShip(ship); if (mode === 'scan') setScannedShipID(ship.ship_id) }}
                >
                  <span className="tactical-roster-glyph" aria-hidden="true"><ProceduralShipGlyph seed={`tactical-roster|empire:${ship.empire_id}|ship:${ship.ship_id}`} hullId={ship.hull_id} weaponCount={(ship.weapons ?? []).reduce((sum, weapon) => sum + weapon.count, 0)} footprint={1} /></span>
                  <span className="tactical-roster-copy"><strong>{shipName(ship.ship_id)}</strong><small>{ship.movement_current}/{ship.movement_max} · {ship.activation_complete ? t('battlefield.shipDone') : active ? t('battlefield.shipActive') : t('battlefield.shipReady')}</small></span>
                </button>
              )
            })}
          </div>
        </section>

        <section className="tactical-hud-actions" aria-label={t('battlefield.modeControls')}>
          <button
            type="button"
            className={`tactical-scan-toggle ${mode === 'scan' ? 'is-active' : ''}`}
            onClick={() => {
              setMode((current) => current === 'scan' ? 'combat' : 'scan')
              setScannedShipID(null)
            }}
          >
            <GameIcon name="info" />{t('battlefield.scan')}
          </button>
          {legalFireActions.length > 0 && (
            <div className="tactical-weapon-strip" aria-label={t('battlefield.weaponAction')}>
              {legalFireActions.map((action) => (
                <button key={action.weapon_slot} type="button" className={selectedWeaponSlot === action.weapon_slot ? 'is-active' : ''} disabled={commandsDisabled || busy} onClick={() => setSelectedWeaponSlot(action.weapon_slot)}>
                  <strong>{humanize(action.weapon_id)}</strong><small>{t('battlefield.targets', { count: action.targets.length })}</small>
                </button>
              ))}
            </div>
          )}
        </section>

        <section className="tactical-hud-commit">
          <button type="button" className="tactical-next-ship" disabled={!tactical.can_end_activation || controlsDisabled} onClick={endActivation}><GameIcon name="check" />{busy ? t('battlefield.commandBusy') : t('battlefield.nextShip')}</button>
          <button type="button" className="tactical-retreat" disabled={controlsDisabled} onClick={retreat}>{t('battlefield.retreat')}</button>
          <button type="button" className="tactical-overview" onClick={onBack}>{t('battle.backToEncounter')}</button>
        </section>
      </footer>
    </div>
  )
}

function TacticalScanDetails({ ship, shipName, empireName, t }: { ship: TacticalShipView; shipName: (shipID: number) => string; empireName: (empireID: number) => string; t: Translator }) {
  return (
    <div className="tactical-scan-details galaxy-fleet-info-content">
      <div>
        <h4>{shipName(ship.ship_id)}</h4>
        <p className="muted">{empireName(ship.empire_id)} · {humanize(ship.hull_id)} · {humanize(ship.warp_drive_id)}</p>
      </div>
      <dl className="tactical-scan-facts galaxy-fleet-info-grid">
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
      <div className="tactical-scan-weapons galaxy-fleet-info-weapons">
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
