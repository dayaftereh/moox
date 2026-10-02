import { useEffect, useMemo, useRef, useState, type MouseEvent as ReactMouseEvent, type PointerEvent as ReactPointerEvent, type WheelEvent as ReactWheelEvent } from 'react'
import {
  decodeShipVisualGenome,
  tacticalEndActivationCommand,
  tacticalFireBeamCommand,
  tacticalFireBeamProjectileCommand,
  tacticalFireOrdnanceCommand,
  tacticalMoveCommand,
  tacticalRetreatCommand,
  tacticalWaitActivationCommand,
  type BattleView,
  type ProtocolCommand,
  type TacticalEvent,
  type TacticalProjectileState,
  type TacticalShipView,
} from './api'
import { GameIcon } from './components/GameIcon'
import { ProceduralShipGlyph } from './components/ProceduralShipGlyph'
import { Notice } from './components/ui'
import { shipHullFootprint } from './shipVisualGenome'
import { type TranslationKey, type TranslationVars } from './i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type TacticalBattlefieldProps = {
  battle: BattleView
  ownSeatID: number
  shipName: (shipID: number) => string
  empireName: (empireID: number) => string
  commandsDisabled: boolean
  onCommand: (command: ProtocolCommand) => Promise<void>
  t: Translator
}

type TacticalMode = 'combat' | 'scan'
type CameraState = { cx: number; cy: number; zoom: number }
type PointerPoint = { x: number; y: number }
type GestureSnapshot = { x: number; y: number; distance: number }
type TacticalDamageLayerKind = 'shield' | 'armor' | 'structure' | 'projectile'
type TacticalDamageLayer = { kind: TacticalDamageLayerKind; amount: number }
type TacticalBeamAnimation = { sequence: number; fromX: number; fromY: number; toX: number; toY: number; hit: boolean; damage: number; shotCount: number; shotHits: boolean[]; weaponID: string; physicalWeaponCount: number; burstShotCount: number; layers: TacticalDamageLayer[] }
type TacticalMoveAnimation = { sequence: number; fromX: number; fromY: number; toX: number; toY: number }
type TacticalProjectileAnimation = { sequence: number; projectileID: number; kind: string; fromX?: number; fromY?: number; toX: number; toY: number; multiplicity: number }

function projectileSalvoOffsets(multiplicity: number): number[] {
  const visible = Math.max(1, Math.min(5, Math.trunc(multiplicity) || 1))
  const spacing = 0.22
  const start = -((visible - 1) * spacing) / 2
  return Array.from({ length: visible }, (_, index) => start + index * spacing)
}
function weaponUnavailableLabel(reason: string | undefined, t: Translator): string {
  switch (reason) {
    case 'tactical_consumer_unavailable':
      return t('battlefield.tacticalConsumerUnavailable')
    case 'weapon_not_ready':
      return t('battlefield.weaponSpent')
    case 'no_legal_targets':
      return t('battlefield.noWeaponTargets')
    case 'ammo_exhausted':
      return t('battlefield.ammoExhausted')
    case 'weapon_cooldown':
      return t('battlefield.weaponCooldown')
    case 'ordnance_effect_unavailable':
      return t('battlefield.ordnanceEffectUnavailable')
    default:
      return t('battlefield.weaponUnavailable')
  }
}

const MIN_ZOOM = 0.45
const MAX_ZOOM = 7
const FOCUS_ZOOM = 3.2
const DAMAGE_FEEDBACK_MS = 4200

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

function eventNumber(data: Record<string, unknown> | undefined, key: string): number | undefined {
  const value = data?.[key]
  return typeof value === 'number' && Number.isFinite(value) ? value : undefined
}

function damageLayersForBeam(events: TacticalEvent[], beam: TacticalEvent): TacticalDamageLayer[] {
  const isVolley = beam.kind === 'beam_volley_resolved'
  const hit = isVolley ? (eventNumber(beam.data, 'hit_count') ?? 0) > 0 : beam.data?.hit === true
  if (!hit) return []
  const targetProjectileID = eventNumber(beam.data, 'target_projectile_id')
  if (targetProjectileID != null && targetProjectileID !== 0) {
    const damage = isVolley ? eventNumber(beam.data, 'total_damage') ?? 0 : eventNumber(beam.data, 'damage') ?? 0
    return damage > 0 ? [{ kind: 'projectile', amount: damage }] : []
  }
  const targetShipID = eventNumber(beam.data, 'target_ship_id')
  const weaponSlot = eventNumber(beam.data, 'weapon_slot')
  const shotIndex = eventNumber(beam.data, 'shot_index')
  const damageEvents = events.filter((event) => event.kind === 'battle_damage_applied'
    && event.command_sequence === beam.command_sequence
    && (targetShipID == null || eventNumber(event.data, 'target_ship_id') === targetShipID)
    && (weaponSlot == null || eventNumber(event.data, 'weapon_slot') === weaponSlot)
    && (shotIndex == null || eventNumber(event.data, 'shot_index') === shotIndex))
  const amounts = new Map<TacticalDamageLayerKind, number>()
  for (const damageEvent of damageEvents) {
    const shieldBefore = eventNumber(damageEvent.data, 'shield_before')
    const shieldAfter = eventNumber(damageEvent.data, 'shield_after')
    if (shieldBefore != null && shieldAfter != null && shieldBefore > shieldAfter) amounts.set('shield', (amounts.get('shield') ?? 0) + shieldBefore - shieldAfter)
    const armorBefore = eventNumber(damageEvent.data, 'armor_before')
    const armorAfter = eventNumber(damageEvent.data, 'armor_after')
    if (armorBefore != null && armorAfter != null && armorBefore > armorAfter) amounts.set('armor', (amounts.get('armor') ?? 0) + armorBefore - armorAfter)
    const structureBefore = eventNumber(damageEvent.data, 'structure_damage_before')
    const structureAfter = eventNumber(damageEvent.data, 'structure_damage_after')
    if (structureBefore != null && structureAfter != null && structureAfter > structureBefore) amounts.set('structure', (amounts.get('structure') ?? 0) + structureAfter - structureBefore)
  }
  return (['shield', 'armor', 'structure'] as TacticalDamageLayerKind[])
    .map((kind) => ({ kind, amount: amounts.get(kind) ?? 0 }))
    .filter((layer) => layer.amount > 0)
}

function projectileEventPosition(
  events: TacticalEvent[],
  event: TacticalEvent,
  projectiles: TacticalProjectileState[],
  ships: TacticalShipView[],
): { x: number; y: number } | null {
  const projectileID = eventNumber(event.data, 'projectile_id')
  if (projectileID == null) return null
  const live = projectiles.find((projectile) => projectile.projectile_id === projectileID)
  const directX = eventNumber(event.data, 'to_x') ?? eventNumber(event.data, 'start_x') ?? eventNumber(event.data, 'target_x')
  const directY = eventNumber(event.data, 'to_y') ?? eventNumber(event.data, 'start_y') ?? eventNumber(event.data, 'target_y')
  if (directX != null && directY != null) return { x: directX, y: directY }
  if (live) return { x: live.x, y: live.y }
  const targetShipID = eventNumber(event.data, 'target_ship_id')
  const target = ships.find((ship) => ship.ship_id === targetShipID)
  if (target && (event.kind === 'projectile_impact_resolved' || event.kind === 'projectile_destroyed')) return { x: target.x, y: target.y }
  const eventIndex = events.findIndex((candidate) => candidate.sequence === event.sequence)
  for (let i = (eventIndex >= 0 ? eventIndex : events.length) - 1; i >= 0; i -= 1) {
    const candidate = events[i]
    if (eventNumber(candidate.data, 'projectile_id') !== projectileID) continue
    const x = eventNumber(candidate.data, 'to_x') ?? eventNumber(candidate.data, 'start_x') ?? eventNumber(candidate.data, 'target_x')
    const y = eventNumber(candidate.data, 'to_y') ?? eventNumber(candidate.data, 'start_y') ?? eventNumber(candidate.data, 'target_y')
    if (x != null && y != null) return { x, y }
  }
  return null
}function beamShotOffset(animation: TacticalBeamAnimation, index: number) {
  if (animation.shotCount <= 1) return { x: 0, y: 0 }
  const dx = animation.toX - animation.fromX
  const dy = animation.toY - animation.fromY
  const length = Math.max(.001, Math.hypot(dx, dy))
  const spread = (index - (animation.shotCount - 1) / 2) * .09
  return { x: (-dy / length) * spread, y: (dx / length) * spread }
}

function reachableGridPath(moves: Array<{ x: number; y: number }>): string {
  const edges = new Set<string>()
  const add = (x1: number, y1: number, x2: number, y2: number) => {
    const a = x1.toFixed(2) + ',' + y1.toFixed(2)
    const b = x2.toFixed(2) + ',' + y2.toFixed(2)
    edges.add(a < b ? a + '|' + b : b + '|' + a)
  }
  for (const move of moves) {
    const left = move.x - .5
    const right = move.x + .5
    const top = move.y - .5
    const bottom = move.y + .5
    add(left, top, right, top)
    add(right, top, right, bottom)
    add(left, bottom, right, bottom)
    add(left, top, left, bottom)
  }
  return [...edges].map((edge) => {
    const parts = edge.split('|')
    const a = parts[0].split(',')
    const b = parts[1].split(',')
    return 'M ' + a[0] + ' ' + a[1] + ' L ' + b[0] + ' ' + b[1]
  }).join(' ')
}

export function TacticalBattlefield({ battle, ownSeatID, shipName, empireName, commandsDisabled, onCommand, t }: TacticalBattlefieldProps) {
  const tactical = battle.tactical
  if (!tactical) return <Notice title={t('battle.tacticalUnavailableTitle')} tone="warning">{t('battle.noTacticalSpec')}</Notice>

  const ships = tactical.ships ?? []
  const projectiles = tactical.ordnance?.projectiles ?? []
  const ordnanceWeaponStates = tactical.ordnance?.weapon_states ?? []
  const ownShips = ships.filter((ship) => ship.seat_id === ownSeatID)
  const remainingOwnShips = ownShips.filter((ship) => !ship.destroyed && !ship.activation_complete)
  const legalMoves = tactical.legal_moves ?? []
  const weaponActions = tactical.weapon_actions ?? []
  const waitTargetShipIDs = tactical.wait_target_ship_ids ?? []
  const waitTargetShipIDSet = useMemo(() => new Set(waitTargetShipIDs), [waitTargetShipIDs])
  const activeShip = ships.find((ship) => ship.ship_id === tactical.state.active_ship_id)
  const ownActivation = activeShip?.seat_id === ownSeatID
  const [mode, setMode] = useState<TacticalMode>('combat')
  const [selectedWeaponSlot, setSelectedWeaponSlot] = useState<number | null>(() => weaponActions.find((action) => action.available)?.weapon_slot ?? null)
  const [beamAnimation, setBeamAnimation] = useState<TacticalBeamAnimation | null>(null)
  const [moveAnimation, setMoveAnimation] = useState<TacticalMoveAnimation | null>(null)
  const [projectileAnimation, setProjectileAnimation] = useState<TacticalProjectileAnimation | null>(null)
  const [scannedShipID, setScannedShipID] = useState<number | null>(null)
  const [selectedShipID, setSelectedShipID] = useState<number | null>(() => ownActivation ? activeShip?.ship_id ?? null : null)
  const [busy, setBusy] = useState(false)
  const [retreatConfirmOpen, setRetreatConfirmOpen] = useState(false)
  const [commandError, setCommandError] = useState('')

  const points = useMemo(() => {
    const shipPoints = ships.map((ship) => ({ x: ship.x, y: ship.y }))
    const projectilePoints = projectiles.map((projectile) => ({ x: projectile.x, y: projectile.y }))
    const movePoints = legalMoves.map((move) => ({ x: move.x, y: move.y }))
    return shipPoints.concat(projectilePoints, movePoints)
  }, [ships, projectiles, legalMoves])

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

  const [camera, setCamera] = useState<CameraState>(() => activeShip
    ? { cx: activeShip.x, cy: activeShip.y, zoom: FOCUS_ZOOM }
    : { cx: bounds.cx, cy: bounds.cy, zoom: 1 })
  const pointersRef = useRef(new Map<number, PointerPoint>())
  const gestureRef = useRef<GestureSnapshot | null>(null)
  const draggedRef = useRef(false)
  const ignoreClickRef = useRef(false)
  const observedEventSequenceRef = useRef<number | null>(null)

  useEffect(() => {
    const selected = weaponActions.find((action) => action.weapon_slot === selectedWeaponSlot)
    if (selectedWeaponSlot == null || !selected?.available) {
      setSelectedWeaponSlot(weaponActions.find((action) => action.available)?.weapon_slot ?? null)
    }
  }, [weaponActions, selectedWeaponSlot])

  useEffect(() => {
    if (scannedShipID != null && !ships.some((ship) => ship.ship_id === scannedShipID)) setScannedShipID(null)
  }, [ships, scannedShipID])

  useEffect(() => {
    if (!retreatConfirmOpen) return undefined
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setRetreatConfirmOpen(false)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [retreatConfirmOpen])

  useEffect(() => {
    if (selectedShipID != null && !ships.some((ship) => ship.ship_id === selectedShipID && ship.seat_id === ownSeatID && !ship.destroyed)) setSelectedShipID(null)
  }, [ships, selectedShipID, ownSeatID])

  useEffect(() => {
    if (ownActivation && activeShip) setSelectedShipID(activeShip.ship_id)
  }, [activeShip?.ship_id, ownActivation])

  const selectedWeaponAction = weaponActions.find((action) => action.weapon_slot === selectedWeaponSlot) ?? weaponActions.find((action) => action.available)
  const weaponActionBySlot = useMemo(() => new Map(weaponActions.map((action) => [action.weapon_slot, action])), [weaponActions])
  const ordnanceWeaponStateBySlot = useMemo(() => new Map(
    ordnanceWeaponStates
      .filter((state) => state.ship_id === activeShip?.ship_id)
      .map((state) => [state.weapon_slot, state]),
  ), [ordnanceWeaponStates, activeShip?.ship_id])
  const activeWeapons = ownActivation ? (activeShip?.weapons ?? []) : []
  const selectedTargets = selectedWeaponAction?.available ? selectedWeaponAction.targets ?? [] : []
  const legalShipTargetByID = useMemo(() => new Map(
    selectedTargets.filter((target) => target.target_ship_id != null).map((target) => [target.target_ship_id as number, target]),
  ), [selectedTargets])
  const legalProjectileTargetByID = useMemo(() => new Map(
    selectedTargets.filter((target) => target.target_projectile_id != null).map((target) => [target.target_projectile_id as number, target]),
  ), [selectedTargets])
  const legalMoveByCell = useMemo(() => new Map(legalMoves.map((move) => [`${move.x}:${move.y}`, move])), [legalMoves])
  const reachableGridD = useMemo(() => reachableGridPath(legalMoves), [legalMoves])
  const scannedShip = ships.find((ship) => ship.ship_id === scannedShipID)
  const selectedShip = ships.find((ship) => ship.ship_id === selectedShipID && ship.seat_id === ownSeatID)
  const movementSelectionActive = mode !== 'scan' && ownActivation && !!activeShip && selectedShipID === activeShip.ship_id
  const viewWidth = bounds.width / camera.zoom
  const viewHeight = bounds.height / camera.zoom
  const viewX = camera.cx - viewWidth / 2
  const viewY = camera.cy - viewHeight / 2
  const controlsDisabled = commandsDisabled || busy || !ownActivation

  const zoomBy = (factor: number) => setCamera((current) => ({ ...current, zoom: clamp(current.zoom * factor, MIN_ZOOM, MAX_ZOOM) }))
  const focusShip = (ship: TacticalShipView) => setCamera((current) => ({ cx: ship.x, cy: ship.y, zoom: Math.max(current.zoom, FOCUS_ZOOM) }))

  useEffect(() => {
    if (!activeShip) return
    setCamera((current) => ({ cx: activeShip.x, cy: activeShip.y, zoom: Math.max(current.zoom, FOCUS_ZOOM) }))
  }, [activeShip?.ship_id])

  useEffect(() => {
    const events = tactical.events ?? []
    const newestSequence = events.reduce((max, event) => Math.max(max, event.sequence), 0)
    if (observedEventSequenceRef.current == null) {
      observedEventSequenceRef.current = newestSequence
      return
    }
    const unseen = events.filter((event) => event.sequence > (observedEventSequenceRef.current ?? 0)).sort((a, b) => a.sequence - b.sequence)
    observedEventSequenceRef.current = newestSequence
    for (const event of unseen) {
      if (event.kind === 'beam_fired' || event.kind === 'beam_volley_resolved') {
        const shotCount = eventNumber(event.data, 'shot_count') ?? 1
        if (event.kind === 'beam_fired' && shotCount > 1) continue
        const sourceID = eventNumber(event.data, 'ship_id')
        const targetID = eventNumber(event.data, 'target_ship_id')
        const targetProjectileID = eventNumber(event.data, 'target_projectile_id')
        const weaponSlot = eventNumber(event.data, 'weapon_slot')
        const source = ships.find((ship) => ship.ship_id === sourceID)
        const target = ships.find((ship) => ship.ship_id === targetID)
        const targetX = eventNumber(event.data, 'target_x') ?? target?.x
        const targetY = eventNumber(event.data, 'target_y') ?? target?.y
        const shotEvents = event.kind === 'beam_volley_resolved'
          ? events
            .filter((candidate) => candidate.kind === 'beam_fired'
              && candidate.command_sequence === event.command_sequence
              && eventNumber(candidate.data, 'ship_id') === sourceID
              && eventNumber(candidate.data, 'target_ship_id') === targetID
              && eventNumber(candidate.data, 'target_projectile_id') === targetProjectileID
              && eventNumber(candidate.data, 'weapon_slot') === weaponSlot)
            .sort((a, b) => (eventNumber(a.data, 'shot_index') ?? 0) - (eventNumber(b.data, 'shot_index') ?? 0))
          : [event]
        const shotHits = shotEvents.map((shot) => shot.data?.hit === true)
        if (source && targetX != null && targetY != null) {
          setBeamAnimation({
            sequence: event.sequence, fromX: source.x, fromY: source.y, toX: targetX, toY: targetY,
            hit: shotHits.some(Boolean),
            damage: event.kind === 'beam_volley_resolved' ? eventNumber(event.data, 'total_damage') ?? 0 : eventNumber(event.data, 'damage') ?? 0,
            shotCount,
            shotHits,
            weaponID: typeof event.data?.weapon_id === 'string' ? event.data.weapon_id : '',
            physicalWeaponCount: eventNumber(event.data, 'physical_weapon_count') ?? 1,
            burstShotCount: eventNumber(event.data, 'burst_shot_count') ?? 1,
            layers: damageLayersForBeam(events, event),
          })
        }
      } else if (event.kind === 'ship_moved') {
        const fromX = eventNumber(event.data, 'from_x')
        const fromY = eventNumber(event.data, 'from_y')
        const toX = eventNumber(event.data, 'to_x')
        const toY = eventNumber(event.data, 'to_y')
        if (fromX != null && fromY != null && toX != null && toY != null) {
          setMoveAnimation({ sequence: event.sequence, fromX, fromY, toX, toY })
        }
      } else if (event.kind === 'ordnance_launched'
        || event.kind === 'projectile_moved'
        || event.kind === 'projectile_retargeted'
        || event.kind === 'projectile_interception_resolved'
        || event.kind === 'projectile_mirv_split'
        || event.kind === 'projectile_impact_resolved'
        || event.kind === 'projectile_destroyed') {
        const projectileID = eventNumber(event.data, 'projectile_id')
        const position = projectileEventPosition(events, event, projectiles, ships)
        if (projectileID != null && position) {
          const liveProjectile = projectiles.find((projectile) => projectile.projectile_id === projectileID)
          const multiplicity = eventNumber(event.data, 'multiplicity_after')
            ?? eventNumber(event.data, 'multiplicity_before')
            ?? eventNumber(event.data, 'multiplicity')
            ?? liveProjectile?.multiplicity
            ?? 1
          setProjectileAnimation({
            sequence: event.sequence,
            projectileID,
            kind: event.kind,
            fromX: eventNumber(event.data, 'from_x'),
            fromY: eventNumber(event.data, 'from_y'),
            toX: position.x,
            toY: position.y,
            multiplicity: Math.max(1, Math.trunc(multiplicity)),
          })
        }
      }
    }
  }, [tactical.events, ships, projectiles])

  useEffect(() => {
    if (!beamAnimation) return
    const timer = window.setTimeout(() => setBeamAnimation((current) => current?.sequence === beamAnimation.sequence ? null : current), DAMAGE_FEEDBACK_MS)
    return () => window.clearTimeout(timer)
  }, [beamAnimation])

  useEffect(() => {
    if (!moveAnimation) return
    const timer = window.setTimeout(() => setMoveAnimation((current) => current?.sequence === moveAnimation.sequence ? null : current), 720)
    return () => window.clearTimeout(timer)
  }, [moveAnimation])
  useEffect(() => {
    if (!projectileAnimation) return
    const timer = window.setTimeout(() => setProjectileAnimation((current) => current?.sequence === projectileAnimation.sequence ? null : current), 980)
    return () => window.clearTimeout(timer)
  }, [projectileAnimation])

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

  const battlefieldClick = (event: ReactMouseEvent<SVGSVGElement>) => {
    if (!movementSelectionActive || ignoreClickRef.current || controlsDisabled || !activeShip) return
    if ((event.target as Element).closest?.('.tactical-ship')) return
    const rect = event.currentTarget.getBoundingClientRect()
    if (rect.width <= 0 || rect.height <= 0) return
    const x = Math.round(viewX + ((event.clientX - rect.left) / rect.width) * viewWidth)
    const y = Math.round(viewY + ((event.clientY - rect.top) / rect.height) * viewHeight)
    const move = legalMoveByCell.get(`${x}:${y}`)
    if (move) moveTo(move.x, move.y)
  }

  const selectShip = (ship: TacticalShipView) => {
    if (ignoreClickRef.current) return
    if (mode === 'scan') {
      setScannedShipID(ship.ship_id)
      focusShip(ship)
      return
    }
    if (ship.seat_id === ownSeatID) {
      setSelectedShipID(ship.ship_id)
      focusShip(ship)
      if (!controlsDisabled && ownActivation && activeShip && ship.ship_id !== activeShip.ship_id && waitTargetShipIDSet.has(ship.ship_id)) {
        void runCommand(tacticalWaitActivationCommand(tactical, activeShip.ship_id, ship.ship_id))
      }
      return
    }
    if (controlsDisabled || !activeShip || !selectedWeaponAction?.available) return
    if (!legalShipTargetByID.has(ship.ship_id)) return
    if (selectedWeaponAction.command_kind === 'battle.fire_beam') {
      void runCommand(tacticalFireBeamCommand(tactical, activeShip.ship_id, ship.ship_id, selectedWeaponAction.weapon_slot))
    } else if (selectedWeaponAction.command_kind === 'battle.fire_ordnance') {
      void runCommand(tacticalFireOrdnanceCommand(tactical, activeShip.ship_id, ship.ship_id, selectedWeaponAction.weapon_slot))
    }
  }
  const selectProjectile = (projectile: TacticalProjectileState) => {
    if (ignoreClickRef.current || controlsDisabled || !activeShip || !selectedWeaponAction?.available) return
    if (selectedWeaponAction.command_kind !== 'battle.fire_beam' || !legalProjectileTargetByID.has(projectile.projectile_id)) return
    void runCommand(tacticalFireBeamProjectileCommand(tactical, activeShip.ship_id, projectile.projectile_id, selectedWeaponAction.weapon_slot))
  }

  const waitActivation = () => {
    if (controlsDisabled || !ownActivation || !tactical.can_wait_activation || !activeShip) return
    void runCommand(tacticalWaitActivationCommand(tactical, activeShip.ship_id))
  }

  const endActivation = () => {
    if (controlsDisabled || !tactical.can_end_activation || !activeShip) return
    void runCommand(tacticalEndActivationCommand(tactical, activeShip.ship_id))
  }

  const requestRetreat = () => {
    if (controlsDisabled || !activeShip) return
    setRetreatConfirmOpen(true)
  }

  const cancelRetreat = () => setRetreatConfirmOpen(false)

  const confirmRetreat = () => {
    if (controlsDisabled || !activeShip) return
    setRetreatConfirmOpen(false)
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
        <strong>{activeShip ? (ownActivation ? t('battlefield.yourActivation') : t('battlefield.waitingActivation', { empire: empireName(activeShip.empire_id) })) : t('battlefield.noActiveShip')}</strong>
        {(selectedShip ?? activeShip) && <span>{selectedShip ? `${shipName(selectedShip.ship_id)} · ` : ''}{(selectedShip ?? activeShip)?.movement_current}/{(selectedShip ?? activeShip)?.movement_max}</span>}
      </div>


      {commandError && <div className="tactical-command-error"><Notice title={t('battlefield.commandRejectedTitle')} tone="warning">{commandError} · {t('battlefield.commandRejectedBody')}</Notice></div>}

      <div className="tactical-viewport-shell">
        <svg
          className="tactical-viewport"
          viewBox={`${viewX} ${viewY} ${viewWidth} ${viewHeight}`}
          onPointerDown={pointerDown}
          onPointerMove={pointerMove}
          onClick={battlefieldClick}
          onPointerUp={pointerEnd}
          onPointerCancel={pointerEnd}
          onWheel={wheel}
          onContextMenu={(event) => event.preventDefault()}
          aria-label={t('battlefield.viewportLabel')}
        >
          <rect x={viewX} y={viewY} width={viewWidth} height={viewHeight} className="tactical-space" />

          {selectedShip && (
            <rect x={selectedShip.x - .5} y={selectedShip.y - .5} width={1} height={1} className="tactical-selected-cell" pointerEvents="none" />
          )}

          {selectedShip && ships.filter((ship) => !ship.destroyed && ship.ship_id !== selectedShip.ship_id).map((ship) => (
            <rect key={`occupied:${ship.ship_id}`} x={ship.x - .5} y={ship.y - .5} width={1} height={1} className="tactical-occupied-cell" pointerEvents="none" />
          ))}

          {movementSelectionActive && reachableGridD && (
            <path d={reachableGridD} className="tactical-reachable-grid" pointerEvents="none" />
          )}
          {projectiles.map((projectile) => {
            const isOwn = projectile.source_seat_id === ownSeatID
            const isLegalTarget = mode !== 'scan' && legalProjectileTargetByID.has(projectile.projectile_id)
            const angle = projectile.facing * 22.5
            return (
              <g
                key={`projectile:${projectile.projectile_id}`}
                className={`tactical-projectile ${isOwn ? 'own' : 'enemy'} ${isLegalTarget ? 'legal-target' : ''}`}
                transform={`translate(${projectile.x} ${projectile.y})`}
                onClick={(event) => { event.stopPropagation(); selectProjectile(projectile) }}
                role="button"
                tabIndex={isLegalTarget ? 0 : -1}
                aria-label={t('battlefield.projectileMarker', { id: projectile.projectile_id, weapon: humanize(projectile.weapon_id), count: projectile.multiplicity, x: projectile.x, y: projectile.y })}
                onKeyDown={(event) => { if (isLegalTarget && (event.key === 'Enter' || event.key === ' ')) selectProjectile(projectile) }}
              >
                <circle r={0.48 + Math.min(4, projectileSalvoOffsets(projectile.multiplicity).length - 1) * 0.07} className="tactical-projectile-hit-cell" />
                <g transform={`rotate(${angle})`} pointerEvents="none">
                  {projectileSalvoOffsets(projectile.multiplicity).map((offset, index) => (
                    <g key={`projectile-glyph:${projectile.projectile_id}:${index}`} transform={`translate(${-Math.abs(offset) * 0.08} ${offset})`}>
                      <path d="M -0.34 -0.085 L 0.18 -0.085 L 0.42 0 L 0.18 0.085 L -0.34 0.085 Z" className="tactical-projectile-vector tactical-projectile-body" />
                      <path d="M -0.24 -0.085 L -0.38 -0.19 L -0.08 -0.085 Z M -0.24 0.085 L -0.38 0.19 L -0.08 0.085 Z" className="tactical-projectile-fin" />
                       <line x1={-0.36} y1={0} x2={-0.58} y2={0} className="tactical-projectile-static-exhaust" />
                    </g>
                  ))}
                </g>
                {projectile.multiplicity > 5 && <text x={0.38} y={-0.34} className="tactical-projectile-count" pointerEvents="none">×{projectile.multiplicity}</text>}
              </g>            )
          })}

          {ships.map((ship) => {
            const isActive = ship.ship_id === tactical.state.active_ship_id
            const isOwn = ship.seat_id === ownSeatID
            const isLegalTarget = mode !== 'scan' && legalShipTargetByID.has(ship.ship_id)
            const isScanned = mode === 'scan' && scannedShipID === ship.ship_id
            const genome = decodeShipVisualGenome(ship.visual_genome)
            const visualHullID = genome?.hullId ?? ship.hull_id
            const footprint = shipHullFootprint(visualHullID)
            const visualSize = Math.min(1, Math.max(.26, footprint))
            const angle = ship.facing * 22.5
            return (
              <g
                key={ship.ship_id}
                className={`tactical-ship ${isOwn ? 'own' : 'enemy'} ${isActive ? 'active' : ''} ${selectedShipID === ship.ship_id ? 'selected' : ''} ${isLegalTarget ? 'legal-target' : ''} ${isScanned ? 'scanned' : ''} ${ship.destroyed ? 'destroyed' : ''}`}
                transform={`translate(${ship.x} ${ship.y})`}
                onClick={(event) => { event.stopPropagation(); selectShip(ship) }}
                role="button"
                tabIndex={0}
                aria-label={t('battlefield.shipMarker', { name: shipName(ship.ship_id), x: ship.x, y: ship.y })}
                onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') selectShip(ship) }}
              >
                <rect x={-0.5} y={-0.5} width={1} height={1} className="tactical-ship-hit-cell" />
                <g className="tactical-ship-facing" transform={`rotate(${angle})`} pointerEvents="none">
                  <ProceduralShipGlyph
                    className="tactical-ship-vector"
                    x={-visualSize / 2}
                    y={-visualSize / 2}
                    width={visualSize}
                    height={visualSize}
                    seed={genome?.seed ?? `${ship.empire_id}:${ship.source_design_id ?? ship.ship_id}:${ship.source_design_revision ?? 0}:${ship.strategic_picture_id ?? 0}`}
                    hullId={visualHullID}
                    genome={genome}
                    weaponCount={(ship.weapons ?? []).reduce((sum, weapon) => sum + weapon.count, 0)}
                    footprint={footprint}
                    tightViewBox
                  />
                </g>
              </g>
            )
          })}

          {moveAnimation && (
            <g className="tactical-move-animation" data-sequence={moveAnimation.sequence} pointerEvents="none">
              <line x1={moveAnimation.fromX} y1={moveAnimation.fromY} x2={moveAnimation.toX} y2={moveAnimation.toY} className="tactical-move-trail" />
              <circle cx={moveAnimation.toX} cy={moveAnimation.toY} r={0.22} className="tactical-move-arrival" />
            </g>
          )}

          {projectileAnimation && (
            <g className={`tactical-projectile-animation is-${projectileAnimation.kind}`} data-sequence={projectileAnimation.sequence} data-projectile-id={projectileAnimation.projectileID} pointerEvents="none">
              {projectileAnimation.fromX != null && projectileAnimation.fromY != null && (
                <>
                  <line x1={projectileAnimation.fromX} y1={projectileAnimation.fromY} x2={projectileAnimation.toX} y2={projectileAnimation.toY} className="tactical-projectile-event-trail" />
                  {projectileAnimation.kind === 'projectile_moved' && (
                    <g className="tactical-projectile-flight" data-multiplicity={projectileAnimation.multiplicity}>
                      <animateTransform
                        attributeName="transform"
                        type="translate"
                        from={projectileAnimation.fromX + ' ' + projectileAnimation.fromY}
                        to={projectileAnimation.toX + ' ' + projectileAnimation.toY}
                        dur="820ms"
                        fill="freeze"
                      />
                      <g transform={'rotate(' + (Math.atan2(projectileAnimation.toY - projectileAnimation.fromY, projectileAnimation.toX - projectileAnimation.fromX) * 180 / Math.PI) + ')'}>
                        {projectileSalvoOffsets(projectileAnimation.multiplicity).map((offset, index) => (
                          <g key={`flight-projectile:${projectileAnimation.projectileID}:${index}`} transform={`translate(${-Math.abs(offset) * 0.1} ${offset})`}>
                            <path d="M -0.46 -0.1 L 0.2 -0.1 L 0.48 0 L 0.2 0.1 L -0.46 0.1 Z" className="tactical-projectile-flight-body" />
                            <path d="M -0.28 -0.1 L -0.46 -0.24 L -0.08 -0.1 Z M -0.28 0.1 L -0.46 0.24 L -0.08 0.1 Z" className="tactical-projectile-flight-fin" />
                            <path d="M -0.47 -0.07 L -0.96 0 L -0.47 0.07 Z" className="tactical-projectile-flight-exhaust" />
                            <circle cx={-0.56} cy={0} r={0.05} className="tactical-projectile-flight-flame">
                              <animate attributeName="cx" values="-0.56;-0.92" dur="180ms" repeatCount="indefinite" />
                              <animate attributeName="r" values="0.055;0.018" dur="180ms" repeatCount="indefinite" />
                              <animate attributeName="opacity" values="1;0" dur="180ms" repeatCount="indefinite" />
                            </circle>
                          </g>
                        ))}
                      </g>
                      {projectileAnimation.multiplicity > 5 && (
                        <text x={0.54} y={-0.4} className="tactical-projectile-flight-count">×{projectileAnimation.multiplicity}</text>
                      )}
                    </g>
                  )}
                </>
              )}
              <circle cx={projectileAnimation.toX} cy={projectileAnimation.toY} r={0.34} className="tactical-projectile-event-pulse" />
            </g>
          )}

          {beamAnimation && (
            <g key={beamAnimation.sequence} className={`tactical-beam-animation ${beamAnimation.hit ? 'is-hit' : 'is-miss'}`} data-sequence={beamAnimation.sequence} data-hit={beamAnimation.hit ? 'true' : 'false'} data-hit-count={beamAnimation.shotHits.filter(Boolean).length} data-damage={beamAnimation.damage} data-shot-count={beamAnimation.shotCount} data-physical-weapon-count={beamAnimation.physicalWeaponCount} data-burst-shot-count={beamAnimation.burstShotCount} data-weapon-id={beamAnimation.weaponID} pointerEvents="none">
              {Array.from({ length: Math.min(beamAnimation.shotCount, 8) }, (_, index) => {
                const offset = beamShotOffset(beamAnimation, index)
                const shotHit = beamAnimation.shotHits[index] === true
                return (
                  <g key={index} className={`tactical-beam-shot ${shotHit ? 'is-hit' : 'is-miss'}`} data-shot-index={index + 1} data-hit={shotHit ? 'true' : 'false'} transform={`translate(${offset.x} ${offset.y})`}>
                    <line x1={beamAnimation.fromX} y1={beamAnimation.fromY} x2={beamAnimation.toX} y2={beamAnimation.toY} className="tactical-beam-glow" />
                    <line x1={beamAnimation.fromX} y1={beamAnimation.fromY} x2={beamAnimation.toX} y2={beamAnimation.toY} className="tactical-beam-core" />
                    <circle cx={beamAnimation.toX} cy={beamAnimation.toY} r={0.36} className="tactical-beam-impact" />
                  </g>
                )
              })}
              {(beamAnimation.layers.length > 0 ? beamAnimation.layers : (beamAnimation.hit && beamAnimation.damage > 0 ? [{ kind: 'structure' as const, amount: beamAnimation.damage }] : [])).map((layer, index) => (
                <text
                  key={`${beamAnimation.sequence}:${layer.kind}:${index}`}
                  x={beamAnimation.toX}
                  y={beamAnimation.toY - .18 - index * .28}
                  className={`tactical-damage-number is-${layer.kind}`}
                  data-layer={layer.kind}
                >
                  -{layer.amount}
                </text>
              ))}
            </g>
          )}
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
        <section className="tactical-hud-roster" aria-label={t('battlefield.remainingShips')}>
          <span className="tactical-hud-label">{t('battlefield.remainingShips')}</span>
          <div className="tactical-ship-strip">
            {remainingOwnShips.map((ship) => {
              const active = ship.ship_id === tactical.state.active_ship_id
              const genome = decodeShipVisualGenome(ship.visual_genome)
              const visualHullID = genome?.hullId ?? ship.hull_id
              const weaponTotal = (ship.weapons ?? []).reduce((sum, weapon) => sum + weapon.count, 0)
              const weaponReady = (ship.weapons ?? []).reduce((sum, weapon) => sum + (weapon.ready ? weapon.count : 0), 0)
              const statusLabel = `${t('battlefield.movement')}: ${ship.movement_current}/${ship.movement_max} · ${t('battlefield.weapons')}: ${weaponReady}/${weaponTotal} ${t('battlefield.ready')}`
              return (
                <button
                  type="button"
                  key={ship.ship_id}
                  className={`tactical-roster-ship ${active ? 'is-active' : ''} ${selectedShipID === ship.ship_id ? 'is-selected' : ''}`}
                  aria-current={active ? 'true' : undefined}
                  aria-label={`${shipName(ship.ship_id)} · ${statusLabel}`}
                  title={`${shipName(ship.ship_id)} · ${statusLabel}`}
                  onClick={() => selectShip(ship)}
                >
                  <span className="tactical-roster-glyph" aria-hidden="true"><ProceduralShipGlyph seed={genome?.seed ?? `${ship.empire_id}:${ship.source_design_id ?? ship.ship_id}:${ship.source_design_revision ?? 0}:${ship.strategic_picture_id ?? 0}`} hullId={visualHullID} genome={genome} weaponCount={weaponTotal} footprint={shipHullFootprint(visualHullID)} /></span>
                  <span className="tactical-roster-metrics" aria-hidden="true">
                    <span><GameIcon name="command" /><strong>{ship.movement_current}</strong><small>/{ship.movement_max}</small></span>
                    <span><GameIcon name="fleet-combat" /><strong>{weaponReady}</strong><small>/{weaponTotal}</small></span>
                  </span>
                  <span className="tactical-roster-state" aria-hidden="true" />
                </button>
              )
            })}
          </div>
        </section>

        <section className="tactical-hud-controls" aria-label={t('battlefield.modeControls')}>
          {activeWeapons.length > 0 && (
            <div className="tactical-control-tools">
              <div className="tactical-weapon-strip" aria-label={t('battlefield.weaponAction')}>
                {activeWeapons.map((weapon) => {
                  const action = weaponActionBySlot.get(weapon.slot)
                  const ordnanceState = ordnanceWeaponStateBySlot.get(weapon.slot)
                  const isSelected = selectedWeaponSlot === weapon.slot
                  const canSelect = !!action?.available && !controlsDisabled
                  const ordnanceStatus = action?.weapon_family === 'missile' && ordnanceState
                    ? t('battlefield.ammoRemaining', { count: ordnanceState.ammo_remaining })
                    : action?.weapon_family === 'torpedo' && ordnanceState
                      ? t('battlefield.cooldownRemaining', { rounds: ordnanceState.cooldown_rounds })
                      : ''
                  return (
                    <button
                      key={weapon.slot}
                      type="button"
                      className={(isSelected ? 'is-active' : '') + (action?.unavailable_reason === 'weapon_not_ready' ? ' is-spent' : '')}
                      disabled={!canSelect}
                      aria-pressed={isSelected}
                      onClick={() => setSelectedWeaponSlot(weapon.slot)}
                    >
                      <GameIcon name="fleet-combat" />
                      <span>
                        <strong>S{weapon.slot + 1} {humanize(weapon.weapon_id)} ×{weapon.count}</strong>
                        <small>{t('battlefield.baseDamageShort', { min: weapon.min_damage, max: weapon.max_damage })} · {action?.available ? t('battlefield.targets', { count: action.targets?.length ?? 0 }) : weaponUnavailableLabel(action?.unavailable_reason, t)}{ordnanceStatus ? ` · ${ordnanceStatus}` : ''}</small>
                      </span>
                    </button>
                  )
                })}
              </div>
            </div>
          )}
          <div className="tactical-control-commit">
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
            <button type="button" className="tactical-wait-ship" disabled={!ownActivation || !tactical.can_wait_activation || controlsDisabled} onClick={waitActivation}><GameIcon name="command" />{t('battlefield.waitActivation')}</button>
            <button type="button" className="tactical-next-ship" disabled={!tactical.can_end_activation || controlsDisabled} onClick={endActivation}><GameIcon name="check" />{busy ? t('battlefield.commandBusy') : t('battlefield.finishActivation')}</button>
            <button type="button" className="tactical-retreat" disabled={controlsDisabled} onClick={requestRetreat}><GameIcon name="flag" />{t('battlefield.retreat')}</button>
          </div>
        </section>
      </footer>

      {retreatConfirmOpen && (
        <div className="tactical-confirm-backdrop" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) cancelRetreat() }}>
          <section className="tactical-confirm-dialog" role="dialog" aria-modal="true" aria-labelledby="tactical-retreat-confirm-title">
            <header>
              <span className="tactical-confirm-eyebrow">{t('battlefield.retreat')}</span>
              <h2 id="tactical-retreat-confirm-title">{t('battlefield.retreatConfirmTitle')}</h2>
            </header>
            <p>{t('battlefield.retreatConfirmWarning')}</p>
            <div className="tactical-confirm-actions">
              <button type="button" className="tactical-confirm-cancel" autoFocus onClick={cancelRetreat}>{t('common.cancel')}</button>
              <button type="button" className="tactical-confirm-danger" disabled={controlsDisabled} onClick={confirmRetreat}>{t('battlefield.retreatConfirmAction')}</button>
            </div>
          </section>
        </div>
      )}
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
