import { useEffect, useLayoutEffect, useMemo, useRef, useState, type MouseEvent as ReactMouseEvent, type PointerEvent as ReactPointerEvent, type WheelEvent as ReactWheelEvent } from 'react'
import {
  decodeShipVisualGenome,
  tacticalBoardShipCommand,
  tacticalEndActivationCommand,
  tacticalFireBeamCommand,
  tacticalFireBeamFighterCommand,
  tacticalFireBeamProjectileCommand,
  tacticalFireOrdnanceCommand,
  tacticalLaunchFightersCommand,
  tacticalMoveCommand,
  tacticalRetreatCommand,
  tacticalWaitActivationCommand,
  type BattleView,
  type ProtocolCommand,
  type TacticalEvent,
  type TacticalFighterGroupState,
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
type TacticalShipDrag = { pointerID: number; startX: number; startY: number; dragging: boolean }
type GestureSnapshot = { x: number; y: number; distance: number }
type TacticalDamageLayerKind = 'shield' | 'armor' | 'structure' | 'projectile'
type TacticalDamageLayer = { kind: TacticalDamageLayerKind; amount: number }
type TacticalBeamAnimation = { sequence: number; fromX: number; fromY: number; toX: number; toY: number; hit: boolean; damage: number; shotCount: number; shotHits: boolean[]; weaponID: string; physicalWeaponCount: number; burstShotCount: number; layers: TacticalDamageLayer[]; shieldBroken: boolean; shieldPiercing: boolean }
type TacticalMoveAnimation = { sequence: number; fromX: number; fromY: number; toX: number; toY: number; movementKind: string }
type TacticalProjectileAnimation = { sequence: number; projectileID: number; kind: string; fromX?: number; fromY?: number; toX: number; toY: number; multiplicity: number }
type TacticalFighterAttackAnimation = { sequence: number; groupID: number; payloadKind: string; weaponID: string; fromX: number; fromY: number; toX: number; toY: number; hit: boolean; damage: number; layers: TacticalDamageLayer[]; shieldBroken: boolean }
type TacticalImpactAnimation = { sequence: number; projectileID: number; targetShipID: number; x: number; y: number; layers: TacticalDamageLayer[]; shieldBroken: boolean }
type TacticalShipExplosionAnimation = { sequence: number; shipID: number; x: number; y: number; quantum: boolean; power: number; radiusFields: number }
type TacticalPresentationCue = {
  sequence: number
  followCamera: boolean
  fromX: number
  fromY: number
  toX: number
  toY: number
  holdMs: number
  beam?: TacticalBeamAnimation
  move?: TacticalMoveAnimation
  projectile?: TacticalProjectileAnimation
  fighterAttack?: TacticalFighterAttackAnimation
  impact?: TacticalImpactAnimation
  shipExplosion?: TacticalShipExplosionAnimation
}


type TacticalMiniGlyph = { x: number; y: number; scale: number }

function tacticalMiniFormation(multiplicity: number): TacticalMiniGlyph[] {
  const count = Math.max(1, Math.trunc(multiplicity) || 1)
  const columns = Math.max(1, Math.ceil(Math.sqrt(count)))
  const rows = Math.max(1, Math.ceil(count / columns))
  const spacingX = columns > 1 ? Math.min(.24, .72 / (columns - 1)) : 0
  const spacingY = rows > 1 ? Math.min(.24, .72 / (rows - 1)) : 0
  const scale = Math.max(.14, Math.min(.34, .4 - (Math.max(columns, rows) - 1) * .035))
  const startX = -((columns - 1) * spacingX) / 2
  const startY = -((rows - 1) * spacingY) / 2
  return Array.from({ length: count }, (_, index) => ({
    x: startX + (index % columns) * spacingX,
    y: startY + Math.floor(index / columns) * spacingY,
    scale,
  }))
}

function fighterMiniGlyphPath(kind: string): string {
  if (kind === 'bomber') return 'M .3 0 L -.08 -.23 L -.25 -.17 L -.12 0 L -.25 .17 L -.08 .23 Z'
  if (kind === 'heavy_fighter') return 'M .34 0 L -.03 -.2 L -.21 -.17 L -.1 0 L -.21 .17 L -.03 .2 Z'
  return 'M .32 0 L -.22 -.13 L -.08 0 L -.22 .13 Z'
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
const COMBAT_CAMERA_MIN_ZOOM = 1.05
const COMBAT_CAMERA_MAX_ZOOM = 2.8
const COMBAT_CAMERA_PAD_X = 12
const COMBAT_CAMERA_PAD_Y = 9
const COMBAT_CUE_REACTIVE_PD_MS = 624
const COMBAT_CUE_BEAM_MS = 1320
const COMBAT_CUE_FIGHTER_ATTACK_MS = 1500
const COMBAT_CUE_MOVE_MS = 840
const COMBAT_CUE_PROJECTILE_MS = 900
const COMBAT_CUE_EXPLOSION_MS = 2160
const DRAG_THRESHOLD_PX = 4
const PINCH_THRESHOLD = 0.01
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

function eventString(data: Record<string, unknown> | undefined, key: string): string | undefined {
  const value = data?.[key]
  return typeof value === 'string' ? value : undefined
}
function damageLayersFromApplications(damageEvents: TacticalEvent[]): TacticalDamageLayer[] {
  // The server reports actual shield charge, armor and structure before/after
  // values. Never infer a red hull number from raw incoming weapon damage.
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

function damageLayersForFighterAttack(events: TacticalEvent[], attack: TacticalEvent): TacticalDamageLayer[] {
  const groupID = eventNumber(attack.data, 'group_id')
  if (groupID == null) return []
  const previousAttack = events
    .filter((event) => event.kind === 'fighter_attack_resolved' && event.command_sequence === attack.command_sequence
      && eventNumber(event.data, 'group_id') === groupID && event.sequence < attack.sequence)
    .reduce((latest, event) => Math.max(latest, event.sequence), 0)
  const targetShipID = eventNumber(attack.data, 'target_ship_id')
  return damageLayersFromApplications(events.filter((event) => event.kind === 'battle_damage_applied'
    && event.command_sequence === attack.command_sequence
    && event.sequence < attack.sequence && event.sequence > previousAttack
    && eventNumber(event.data, 'fighter_group_id') === groupID
    && (targetShipID == null || eventNumber(event.data, 'target_ship_id') === targetShipID)))
}

function damageLayersForProjectileImpact(events: TacticalEvent[], impact: TacticalEvent): TacticalDamageLayer[] {
  const projectileID = eventNumber(impact.data, 'projectile_id')
  const impactIndex = eventNumber(impact.data, 'impact_index')
  if (projectileID == null || impactIndex == null) return []
  return damageLayersFromApplications(events.filter((event) => event.kind === 'battle_damage_applied'
    && event.command_sequence === impact.command_sequence
    && event.sequence > impact.sequence
    && eventNumber(event.data, 'projectile_id') === projectileID
    && eventNumber(event.data, 'impact_index') === impactIndex
    && eventNumber(event.data, 'target_ship_id') === eventNumber(impact.data, 'target_ship_id')))
}

function shieldBrokenAfterHit(target: TacticalShipView | undefined, layers: TacticalDamageLayer[]): boolean {
  // Only announce a shield collapse for a real shield-charge decrease ending
  // with all four authoritative arcs depleted.
  return !!target && layers.some((layer) => layer.kind === 'shield')
    && target.shield_arc_current.every((value) => value <= 0)
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
  return damageLayersFromApplications(damageEvents)
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
}

function fighterGroupEventPosition(
  events: TacticalEvent[],
  event: TacticalEvent,
  groupID: number,
  fighters: TacticalFighterGroupState[],
): { x: number; y: number } | null {
  const eventIndex = events.findIndex((candidate) => candidate.sequence === event.sequence)
  for (let i = (eventIndex >= 0 ? eventIndex : events.length) - 1; i >= 0; i -= 1) {
    const candidate = events[i]
    if (eventNumber(candidate.data, 'group_id') !== groupID) continue
    const x = eventNumber(candidate.data, 'to_x') ?? eventNumber(candidate.data, 'start_x')
    const y = eventNumber(candidate.data, 'to_y') ?? eventNumber(candidate.data, 'start_y')
    if (x != null && y != null) return { x, y }
  }
  const live = fighters.find((fighter) => fighter.group_id === groupID)
  return live ? { x: live.x, y: live.y } : null
}

function fighterAttackTargetPosition(
  events: TacticalEvent[],
  event: TacticalEvent,
  fighters: TacticalFighterGroupState[],
  projectiles: TacticalProjectileState[],
  ships: TacticalShipView[],
): { x: number; y: number } | null {
  const targetShipID = eventNumber(event.data, 'target_ship_id')
  if (targetShipID != null && targetShipID !== 0) {
    const target = ships.find((ship) => ship.ship_id === targetShipID)
    if (target) return { x: target.x, y: target.y }
  }
  const targetFighterGroupID = eventNumber(event.data, 'target_fighter_group_id')
  if (targetFighterGroupID != null && targetFighterGroupID !== 0) {
    return fighterGroupEventPosition(events, event, targetFighterGroupID, fighters)
  }
  const targetProjectileID = eventNumber(event.data, 'target_projectile_id')
  if (targetProjectileID != null && targetProjectileID !== 0) {
    const eventIndex = events.findIndex((candidate) => candidate.sequence === event.sequence)
    for (let i = (eventIndex >= 0 ? eventIndex : events.length) - 1; i >= 0; i -= 1) {
      const candidate = events[i]
      if (eventNumber(candidate.data, 'projectile_id') !== targetProjectileID) continue
      const x = eventNumber(candidate.data, 'to_x') ?? eventNumber(candidate.data, 'start_x') ?? eventNumber(candidate.data, 'target_x')
      const y = eventNumber(candidate.data, 'to_y') ?? eventNumber(candidate.data, 'start_y') ?? eventNumber(candidate.data, 'target_y')
      if (x != null && y != null) return { x, y }
    }
    const live = projectiles.find((projectile) => projectile.projectile_id === targetProjectileID)
    if (live) return { x: live.x, y: live.y }
  }
  return null
}

function fighterAttackOutcome(data: Record<string, unknown> | undefined): { hit: boolean; damage: number } {
  const attempts = Array.isArray(data?.attempts) ? data.attempts : []
  let hit = false
  let damage = 0
  for (const raw of attempts) {
    if (!raw || typeof raw !== 'object') continue
    const attempt = raw as Record<string, unknown>
    if (attempt.hit === true) hit = true
    if (typeof attempt.damage === 'number' && Number.isFinite(attempt.damage)) damage += Math.max(0, attempt.damage)
  }
  return { hit, damage }
}

function beamShotOffset(animation: TacticalBeamAnimation, index: number) {
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

/** Arc indices relative to facing (0=front, 3=right, 2=rear, 1=left).
 * Positive SVG Y is down, matching tacticalShieldImpactArc. */
const TACTICAL_SHIELD_SECTORS = [
  { index: 0, key: 'front', centerDegrees: 0, label: 'battlefield.shieldFront' },
  { index: 3, key: 'right', centerDegrees: 90, label: 'battlefield.shieldRight' },
  { index: 2, key: 'rear', centerDegrees: 180, label: 'battlefield.shieldRear' },
  { index: 1, key: 'left', centerDegrees: 270, label: 'battlefield.shieldLeft' },
] as const
const SHIELD_SECTOR_RADIUS = .77
const SHIELD_SECTOR_HALF_ANGLE = 36

function tacticalShieldSectorPath(centerDegrees: number, fraction: number): string {
  const amount = clamp(fraction, 0, 1)
  if (amount <= 0) return ''
  const start = (centerDegrees - SHIELD_SECTOR_HALF_ANGLE) * Math.PI / 180
  const end = (centerDegrees - SHIELD_SECTOR_HALF_ANGLE + SHIELD_SECTOR_HALF_ANGLE * 2 * amount) * Math.PI / 180
  const point = (angle: number) => [
    (Math.cos(angle) * SHIELD_SECTOR_RADIUS).toFixed(4),
    (Math.sin(angle) * SHIELD_SECTOR_RADIUS).toFixed(4),
  ]
  const [startX, startY] = point(start)
  const [endX, endY] = point(end)
  return 'M ' + startX + ' ' + startY + ' A ' + SHIELD_SECTOR_RADIUS + ' ' + SHIELD_SECTOR_RADIUS + ' 0 0 1 ' + endX + ' ' + endY
}

function tacticalWeaponCanPierceShields(weapon: { weapon_id: string; modifier_ids?: string[]; innate_effect_ids?: string[] }): boolean {
  return weapon.weapon_id === 'particle_beam'
    || weapon.modifier_ids?.includes('shield_piercing') === true
    || weapon.innate_effect_ids?.includes('natural_shield_piercing') === true
    || weapon.innate_effect_ids?.includes('shield_piercing') === true
}

function tacticalBeamShotCanPierceShields(event: TacticalEvent): boolean {
  const modifiers = event.data?.modifier_ids
  return event.data?.weapon_id === 'particle_beam'
    || (Array.isArray(modifiers) && modifiers.includes('shield_piercing'))
}

function TacticalDamageFeedback({ layers, x, y, shieldBroken }: {
  layers: TacticalDamageLayer[]; x: number; y: number; shieldBroken: boolean
}) {
  if (layers.length === 0) return null
  const shieldHit = layers.some((layer) => layer.kind === 'shield')
  return (
    <g className="tactical-layer-damage-feedback" pointerEvents="none" data-layers={layers.map((layer) => layer.kind).join(',')}>
      {shieldHit && (
        <circle cx={x} cy={y} r={0.73}
          className={'tactical-shield-hit-flash' + (shieldBroken ? ' is-shield-break' : '')}
          data-shield-break={shieldBroken ? 'true' : 'false'} />
      )}
      {layers.map((layer, index) => (
        <text key={layer.kind} x={x} y={y - .18 - index * .28}
          className={'tactical-damage-number is-' + layer.kind} data-layer={layer.kind}>
          -{layer.amount}
        </text>
      ))}
    </g>
  )
}

export function TacticalBattlefield({ battle, ownSeatID, shipName, empireName, commandsDisabled, onCommand, t }: TacticalBattlefieldProps) {
  const tactical = battle.tactical
  if (!tactical) return <Notice title={t('battle.tacticalUnavailableTitle')} tone="warning">{t('battle.noTacticalSpec')}</Notice>

  const ships = tactical.ships ?? []
  const projectiles = tactical.ordnance?.projectiles ?? []
  const fighterGroups = tactical.fighters?.groups ?? []
  const fighterBayStates = tactical.fighters?.bay_states ?? []
  const ordnanceWeaponStates = tactical.ordnance?.weapon_states ?? []
  const ownShips = ships.filter((ship) => ship.seat_id === ownSeatID)
  const remainingOwnShips = ownShips.filter((ship) => !ship.destroyed && !ship.activation_complete)
  const legalMoves = tactical.legal_moves ?? []
  const weaponActions = tactical.weapon_actions ?? []
  const boardingActions = tactical.boarding_actions ?? []
  const availableBoardingActions = boardingActions.filter((action) => action.available && action.max_marine_count > 0)
  const waitTargetShipIDs = tactical.wait_target_ship_ids ?? []
  const waitTargetShipIDSet = useMemo(() => new Set(waitTargetShipIDs), [waitTargetShipIDs])
  const activeShip = ships.find((ship) => ship.ship_id === tactical.state.active_ship_id)
  const ownActivation = activeShip?.seat_id === ownSeatID
  const [mode, setMode] = useState<TacticalMode>('combat')
  const [selectedWeaponSlot, setSelectedWeaponSlot] = useState<number | null>(() => weaponActions.find((action) => action.available)?.weapon_slot ?? null)
  const [beamAnimation, setBeamAnimation] = useState<TacticalBeamAnimation | null>(null)
  const [moveAnimation, setMoveAnimation] = useState<TacticalMoveAnimation | null>(null)
  const [projectileAnimation, setProjectileAnimation] = useState<TacticalProjectileAnimation | null>(null)
  const [fighterAttackAnimations, setFighterAttackAnimations] = useState<TacticalFighterAttackAnimation[]>([])
  const [impactAnimation, setImpactAnimation] = useState<TacticalImpactAnimation | null>(null)
  const [shipExplosionAnimation, setShipExplosionAnimation] = useState<TacticalShipExplosionAnimation | null>(null)
  const [scannedShipID, setScannedShipID] = useState<number | null>(null)
  const [selectedShipID, setSelectedShipID] = useState<number | null>(() => ownActivation ? activeShip?.ship_id ?? null : null)
  const [busy, setBusy] = useState(false)
  const [presentationLocked, setPresentationLocked] = useState(false)
  const presentationLockedRef = useRef(false)
  const commandInFlightRef = useRef(false)
  const currentActivationRef = useRef({ ship: activeShip, own: ownActivation })
  currentActivationRef.current = { ship: activeShip, own: ownActivation }


  const [retreatConfirmOpen, setRetreatConfirmOpen] = useState(false)
  const [commandError, setCommandError] = useState('')
  const [selectedBoardingTargetID, setSelectedBoardingTargetID] = useState<number | null>(() => availableBoardingActions[0]?.target_ship_id ?? null)
  const [boardingMarineCount, setBoardingMarineCount] = useState(() => availableBoardingActions[0]?.max_marine_count ?? 1)
  const [boardingMode, setBoardingMode] = useState(() => availableBoardingActions[0]?.allowed_modes?.[0] ?? 'raid')

  const points = useMemo(() => {
    const shipPoints = ships.map((ship) => ({ x: ship.x, y: ship.y }))
    const projectilePoints = projectiles.map((projectile) => ({ x: projectile.x, y: projectile.y }))
    const fighterPoints = fighterGroups.map((fighter) => ({ x: fighter.x, y: fighter.y }))
    const movePoints = legalMoves.map((move) => ({ x: move.x, y: move.y }))
    return shipPoints.concat(projectilePoints, fighterPoints, movePoints)
  }, [ships, projectiles, fighterGroups, legalMoves])

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

  // Start a human activation with a battlefield overview instead of an auto-zoomed ship.
  const [camera, setCamera] = useState<CameraState>(() => ownActivation
    ? { cx: bounds.cx, cy: bounds.cy, zoom: 1 }
    : activeShip
      ? { cx: activeShip.x, cy: activeShip.y, zoom: COMBAT_CAMERA_MAX_ZOOM }
      : { cx: bounds.cx, cy: bounds.cy, zoom: 1 })
  const manualCameraRef = useRef<CameraState>({ cx: bounds.cx, cy: bounds.cy, zoom: 1 })
  const previousOwnActivationRef = useRef(ownActivation)
  // Keep manual zoom fixed as legal moves, fighters and projectiles alter the battlefield bounds.
  const cameraExtentRef = useRef({ width: bounds.width, height: bounds.height })

  useLayoutEffect(() => {
    if (previousOwnActivationRef.current === ownActivation) return
    if (ownActivation) {
      // Returning control to the player must also restore their last manual viewport.
      setCamera(manualCameraRef.current)
    } else {
      manualCameraRef.current = camera
    }
    previousOwnActivationRef.current = ownActivation
  }, [ownActivation, camera])
  const pointersRef = useRef(new Map<number, PointerPoint>())
  const shipDragRef = useRef<TacticalShipDrag | null>(null)

  const gestureRef = useRef<GestureSnapshot | null>(null)
  const draggedRef = useRef(false)
  const dragDistanceRef = useRef(0)
  const ignoreClickRef = useRef(false)
  const observedEventSequenceRef = useRef<number | null>(null)
  const presentationQueueRef = useRef<TacticalPresentationCue[]>([])
  const presentationTimerRef = useRef<number | null>(null)
  const presentationPlayingRef = useRef(false)
  const presentationAutoFollowedRef = useRef(false)

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
  const fighterBayStateBySlot = useMemo(() => new Map(
    fighterBayStates
      .filter((state) => state.ship_id === activeShip?.ship_id)
      .map((state) => [state.weapon_slot, state]),
  ), [fighterBayStates, activeShip?.ship_id])
  const activeWeapons = ownActivation ? (activeShip?.weapons ?? []) : []
  const selectedTargets = selectedWeaponAction?.available ? selectedWeaponAction.targets ?? [] : []
  const legalShipTargetByID = useMemo(() => new Map(
    selectedTargets.filter((target) => target.target_ship_id != null).map((target) => [target.target_ship_id as number, target]),
  ), [selectedTargets])
  const legalProjectileTargetByID = useMemo(() => new Map(
    selectedTargets.filter((target) => target.target_projectile_id != null).map((target) => [target.target_projectile_id as number, target]),
  ), [selectedTargets])
  const legalFighterTargetByID = useMemo(() => new Map(
    selectedTargets.filter((target) => target.target_fighter_group_id != null).map((target) => [target.target_fighter_group_id as number, target]),
  ), [selectedTargets])
  const selectedBoardingAction = availableBoardingActions.find((action) => action.target_ship_id === selectedBoardingTargetID) ?? availableBoardingActions[0]
  const legalMoveByCell = useMemo(() => new Map(legalMoves.map((move) => [`${move.x}:${move.y}`, move])), [legalMoves])
  const reachableGridD = useMemo(() => reachableGridPath(legalMoves), [legalMoves])
  const scannedShip = ships.find((ship) => ship.ship_id === scannedShipID)
  const selectedShip = ships.find((ship) => ship.ship_id === selectedShipID && ship.seat_id === ownSeatID)
  // Show movement overlays for the selected active ship only while it has
  // server-authorized move cells and movement points. Presentation playback,
  // outstanding commands and completed activations must leave the arena clear.
  const movementSelectionActive = !presentationLocked && !commandsDisabled && !busy
    && mode !== 'scan' && ownActivation && !!activeShip && !activeShip.activation_complete
    && selectedShipID === activeShip.ship_id && activeShip.movement_current > 0
    && legalMoves.length > 0
  const boardingSelectionKey = availableBoardingActions.map((action) => `${action.target_ship_id}:${action.max_marine_count}:${(action.allowed_modes ?? []).join(',')}`).join('|')
  useEffect(() => {
    const action = availableBoardingActions.find((candidate) => candidate.target_ship_id === selectedBoardingTargetID) ?? availableBoardingActions[0]
    if (!action) {
      setSelectedBoardingTargetID(null)
      setBoardingMarineCount(1)
      setBoardingMode('raid')
      return
    }
    if (selectedBoardingTargetID !== action.target_ship_id) setSelectedBoardingTargetID(action.target_ship_id)
    setBoardingMarineCount((current) => clamp(current, 1, action.max_marine_count))
    if (!(action.allowed_modes ?? []).includes(boardingMode)) setBoardingMode(action.allowed_modes?.[0] ?? 'raid')
  }, [boardingSelectionKey, selectedBoardingTargetID, boardingMode])
  const viewWidth = cameraExtentRef.current.width / camera.zoom
  const viewHeight = cameraExtentRef.current.height / camera.zoom
  const viewX = camera.cx - viewWidth / 2
  const viewY = camera.cy - viewHeight / 2
  const controlsDisabled = commandsDisabled || busy || presentationLocked || !ownActivation

  const updatePresentationLock = (locked: boolean) => {
    if (presentationLockedRef.current === locked) return
    presentationLockedRef.current = locked
    setPresentationLocked(locked)
  }

  const zoomBy = (factor: number) => {
    if (presentationLockedRef.current) return
    setCamera((current) => {
      const next = { ...current, zoom: clamp(current.zoom * factor, MIN_ZOOM, MAX_ZOOM) }
      if (ownActivation) manualCameraRef.current = next
      return next
    })
  }
  const focusShip = (ship: TacticalShipView) => setCamera((current) => {
    const next = { cx: ship.x, cy: ship.y, zoom: Math.max(current.zoom, FOCUS_ZOOM) }
    if (ownActivation) manualCameraRef.current = next
    return next
  })

  const focusCombatSpan = (fromX: number, fromY: number, toX: number, toY: number) => {
    const spanX = Math.abs(toX - fromX)
    const spanY = Math.abs(toY - fromY)
    const zoomX = cameraExtentRef.current.width / Math.max(14, spanX + COMBAT_CAMERA_PAD_X)
    const zoomY = cameraExtentRef.current.height / Math.max(10, spanY + COMBAT_CAMERA_PAD_Y)
    setCamera({
      cx: (fromX + toX) / 2,
      cy: (fromY + toY) / 2,
      zoom: clamp(Math.min(zoomX, zoomY), COMBAT_CAMERA_MIN_ZOOM, COMBAT_CAMERA_MAX_ZOOM),
    })
  }

  const playNextPresentationCue = () => {
    const cue = presentationQueueRef.current.shift()
    if (!cue) {
      presentationPlayingRef.current = false
      presentationTimerRef.current = null
      // Replay may include the entire AI turn even when the server has already
      // handed control back to the Human in the final authoritative snapshot.
      if (presentationAutoFollowedRef.current && currentActivationRef.current.own) {
        setCamera(manualCameraRef.current)
      }
      presentationAutoFollowedRef.current = false
      updatePresentationLock(false)
      return
    }
    presentationPlayingRef.current = true
    updatePresentationLock(true)
    // An authoritative cue still plays/locks input during the player's turn,
    // but only an opponent activation may steer the camera to the action.
    if (cue.followCamera) {
      presentationAutoFollowedRef.current = true
      focusCombatSpan(cue.fromX, cue.fromY, cue.toX, cue.toY)
    }
    setBeamAnimation(cue.beam ?? null)
    setMoveAnimation(cue.move ?? null)
    setProjectileAnimation(cue.projectile ?? null)
    setFighterAttackAnimations(cue.fighterAttack ? [cue.fighterAttack] : [])
    setImpactAnimation(cue.impact ?? null)
    setShipExplosionAnimation(cue.shipExplosion ?? null)
    // Never re-enable commands before the principal effect for this cue finishes.
    // Floating damage labels may linger, but moving ships, shots and explosions must end.
    const visualMs = cue.shipExplosion ? 2280
      : cue.projectile ? 1176
        : cue.move ? 864
          : cue.fighterAttack ? 1260
            : cue.impact ? 1100
              : cue.beam ? 1620
                : 0
    presentationTimerRef.current = window.setTimeout(playNextPresentationCue, Math.max(cue.holdMs, visualMs))
  }

  const enqueuePresentationCue = (cue: TacticalPresentationCue) => {
    updatePresentationLock(true)
    presentationQueueRef.current.push(cue)
    if (!presentationPlayingRef.current) {
      presentationPlayingRef.current = true
      playNextPresentationCue()
    }
  }

  useEffect(() => () => {
    if (presentationTimerRef.current != null) window.clearTimeout(presentationTimerRef.current)
    presentationQueueRef.current = []
    presentationPlayingRef.current = false
    presentationAutoFollowedRef.current = false
  }, [])

  useEffect(() => {
    if (!activeShip || ownActivation || presentationPlayingRef.current) return
    setCamera((current) => ({
      cx: activeShip.x, cy: activeShip.y,
      zoom: clamp(current.zoom, COMBAT_CAMERA_MIN_ZOOM, COMBAT_CAMERA_MAX_ZOOM),
    }))
  }, [activeShip?.ship_id, ownActivation])

  // Layout effect acquires the lock before the just-refreshed server view is
  // painted: the previous command cannot become clickable between snapshots.
  useLayoutEffect(() => {
    const events = tactical.events ?? []
    const newestSequence = events.reduce((max, event) => Math.max(max, event.sequence), 0)
    if (observedEventSequenceRef.current == null) {
      observedEventSequenceRef.current = newestSequence
      return
    }
    const unseen = events.filter((event) => event.sequence > (observedEventSequenceRef.current ?? 0)).sort((a, b) => a.sequence - b.sequence)
    observedEventSequenceRef.current = newestSequence

    // Replay follows the ownership of the authoritative activation timeline,
    // not the current snapshot's active ship. The server may already have
    // returned to the Human while a whole AI round is still being presented.
    const seatByShip = new Map(ships.map((ship) => [ship.ship_id, ship.seat_id]))
    const followCameraBySequence = new Map<number, boolean>()
    let roundInitiative: number[] = []
    let initiativeIndex = 0
    let knownRound = false
    for (const event of [...events].sort((a, b) => a.sequence - b.sequence)) {
      if (event.kind === 'round_started') {
        const order = event.data?.initiative_order
        roundInitiative = Array.isArray(order) ? order.filter((shipID): shipID is number => typeof shipID === 'number') : []
        const activeID = eventNumber(event.data, 'active_ship_id')
        const startedIndex = roundInitiative.indexOf(activeID ?? -1)
        initiativeIndex = startedIndex >= 0 ? startedIndex : 0
        knownRound = roundInitiative.length > 0
      }
      const actingShipID = roundInitiative[initiativeIndex]
      const humanTurn = knownRound && actingShipID !== undefined && seatByShip.get(actingShipID) === ownSeatID
      followCameraBySequence.set(event.sequence, knownRound ? !humanTurn : !ownActivation)
      if (event.kind === 'activation_ended' || event.kind === 'activation_auto_ended') {
        const endedID = eventNumber(event.data, 'ship_id')
        const endedIndex = roundInitiative.indexOf(endedID ?? -1)
        if (endedIndex >= 0) initiativeIndex = endedIndex + 1
      }
    }
    const enqueueEventCue = (cue: Omit<TacticalPresentationCue, 'followCamera'>) =>
      enqueuePresentationCue({
        ...cue,
        followCamera: followCameraBySequence.get(cue.sequence) ?? !ownActivation,
      })
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
        // Functional Hard Shields counter natural and modified Shield Piercing.
        const targetHardShields = target?.specials?.some((special) =>
          special.special_id === 'hard_shields' && special.functional) === true
        const shieldPiercing = !!target && !targetHardShields
          && (shotEvents.some(tacticalBeamShotCanPierceShields)
            || source?.weapons?.some((weapon) => weapon.slot === weaponSlot && tacticalWeaponCanPierceShields(weapon)) === true)
        if (source && targetX != null && targetY != null) {
          const layers = damageLayersForBeam(events, event)
          const beam: TacticalBeamAnimation = {
            sequence: event.sequence, fromX: source.x, fromY: source.y, toX: targetX, toY: targetY,
            hit: shotHits.some(Boolean),
            damage: event.kind === 'beam_volley_resolved' ? eventNumber(event.data, 'total_damage') ?? 0 : eventNumber(event.data, 'damage') ?? 0,
            shotCount,
            shotHits,
            weaponID: typeof event.data?.weapon_id === 'string' ? event.data.weapon_id : '',
            physicalWeaponCount: eventNumber(event.data, 'physical_weapon_count') ?? 1,
            burstShotCount: eventNumber(event.data, 'burst_shot_count') ?? 1,
            layers,
            shieldBroken: shieldBrokenAfterHit(target, layers),
            shieldPiercing,
          }
          enqueueEventCue({
            sequence: event.sequence,
            fromX: source.x, fromY: source.y, toX: targetX, toY: targetY,
            holdMs: event.data?.reactive_point_defense === true ? COMBAT_CUE_REACTIVE_PD_MS : COMBAT_CUE_BEAM_MS,
            beam,
          })
        }
      } else if (event.kind === 'fighter_attack_resolved') {
        const groupID = eventNumber(event.data, 'group_id')
        const payloadKind = eventString(event.data, 'payload_kind') ?? 'beam'
        const source = groupID != null ? fighterGroupEventPosition(events, event, groupID, fighterGroups) : null
        const target = fighterAttackTargetPosition(events, event, fighterGroups, projectiles, ships)
        if (groupID != null && source && target) {
          const outcome = fighterAttackOutcome(event.data)
          const layers = damageLayersForFighterAttack(events, event)
          const fighterAttack: TacticalFighterAttackAnimation = {
            sequence: event.sequence,
            groupID,
            payloadKind,
            weaponID: eventString(event.data, 'weapon_id') ?? '',
            fromX: source.x,
            fromY: source.y,
            toX: target.x,
            toY: target.y,
            hit: outcome.hit,
            damage: outcome.damage,
            layers,
            shieldBroken: shieldBrokenAfterHit(ships.find((ship) => ship.ship_id === eventNumber(event.data, 'target_ship_id')), layers),
          }
          enqueueEventCue({
            sequence: event.sequence,
            fromX: source.x, fromY: source.y, toX: target.x, toY: target.y,
            holdMs: COMBAT_CUE_FIGHTER_ATTACK_MS,
            fighterAttack,
          })
        }
      } else if (event.kind === 'projectile_impact_resolved') {
        const projectileID = eventNumber(event.data, 'projectile_id')
        const targetShipID = eventNumber(event.data, 'target_ship_id')
        const target = ships.find((ship) => ship.ship_id === targetShipID)
        if (projectileID != null && targetShipID != null && target) {
          const layers = damageLayersForProjectileImpact(events, event)
          if (layers.length > 0) {
            const impact: TacticalImpactAnimation = {
              sequence: event.sequence, projectileID, targetShipID,
              x: target.x, y: target.y, layers,
              shieldBroken: shieldBrokenAfterHit(target, layers),
            }
            enqueueEventCue({
              sequence: event.sequence,
              fromX: target.x, fromY: target.y, toX: target.x, toY: target.y,
              holdMs: 1100, impact,
            })
          }
        }
      } else if (event.kind === 'ship_destroyed') {
        const shipID = eventNumber(event.data, 'ship_id')
        const ship = ships.find((candidate) => candidate.ship_id === shipID)
        if (shipID != null && ship) {
          const radiusPixels = eventNumber(event.data, 'explosion_radius_pixels') ?? 0
          const pixelsPerField = eventNumber(event.data, 'explosion_pixels_per_field') ?? 20
          const radiusFields = radiusPixels > 0 && pixelsPerField > 0 ? radiusPixels / pixelsPerField : 0
          const quantum = event.data?.quantum_detonator === true
          const shipExplosion: TacticalShipExplosionAnimation = {
            sequence: event.sequence,
            shipID,
            x: ship.x,
            y: ship.y,
            quantum,
            power: eventNumber(event.data, 'explosion_power') ?? 0,
            radiusFields,
          }
          enqueueEventCue({
            sequence: event.sequence,
            fromX: quantum && radiusFields > 0 ? ship.x - radiusFields : ship.x,
            fromY: ship.y,
            toX: quantum && radiusFields > 0 ? ship.x + radiusFields : ship.x,
            toY: ship.y,
            holdMs: COMBAT_CUE_EXPLOSION_MS,
            shipExplosion,
          })
        }
      } else if (event.kind === 'ship_moved') {
        const fromX = eventNumber(event.data, 'from_x')
        const fromY = eventNumber(event.data, 'from_y')
        const toX = eventNumber(event.data, 'to_x')
        const toY = eventNumber(event.data, 'to_y')
        if (fromX != null && fromY != null && toX != null && toY != null) {
          const move: TacticalMoveAnimation = {
            sequence: event.sequence,
            fromX, fromY, toX, toY,
            movementKind: eventString(event.data, 'movement_kind') ?? 'normal',
          }
          enqueueEventCue({ sequence: event.sequence, fromX, fromY, toX, toY, holdMs: COMBAT_CUE_MOVE_MS, move })
        }
      } else if (event.kind === 'fighter_group_moved') {
        const groupID = eventNumber(event.data, 'group_id')
        const fromX = eventNumber(event.data, 'from_x')
        const fromY = eventNumber(event.data, 'from_y')
        const toX = eventNumber(event.data, 'to_x')
        const toY = eventNumber(event.data, 'to_y')
        const phase = eventString(event.data, 'phase') ?? 'outbound'
        const attacksImmediately = phase !== 'returning' && groupID != null && unseen.some((candidate) =>
          candidate.sequence > event.sequence
          && candidate.sequence <= event.sequence + 24
          && candidate.kind === 'fighter_attack_resolved'
          && eventNumber(candidate.data, 'group_id') === groupID)
        if (!attacksImmediately && fromX != null && fromY != null && toX != null && toY != null) {
          const move: TacticalMoveAnimation = {
            sequence: event.sequence,
            fromX, fromY, toX, toY,
            movementKind: phase === 'returning' ? 'fighter-returning' : 'fighter',
          }
          enqueueEventCue({ sequence: event.sequence, fromX, fromY, toX, toY, holdMs: COMBAT_CUE_MOVE_MS, move })
        }
      } else if (event.kind === 'projectile_moved') {
        const projectileID = eventNumber(event.data, 'projectile_id')
        const position = projectileEventPosition(events, event, projectiles, ships)
        if (projectileID != null && position) {
          const liveProjectile = projectiles.find((projectile) => projectile.projectile_id === projectileID)
          const multiplicity = eventNumber(event.data, 'multiplicity_after')
            ?? eventNumber(event.data, 'multiplicity_before')
            ?? eventNumber(event.data, 'multiplicity')
            ?? liveProjectile?.multiplicity
            ?? 1
          const fromX = eventNumber(event.data, 'from_x') ?? eventNumber(event.data, 'start_x') ?? position.x
          const fromY = eventNumber(event.data, 'from_y') ?? eventNumber(event.data, 'start_y') ?? position.y
          const projectile: TacticalProjectileAnimation = {
            sequence: event.sequence,
            projectileID,
            kind: event.kind,
            fromX: eventNumber(event.data, 'from_x'),
            fromY: eventNumber(event.data, 'from_y'),
            toX: position.x,
            toY: position.y,
            multiplicity: Math.max(1, Math.trunc(multiplicity)),
          }
          enqueueEventCue({
            sequence: event.sequence,
            fromX, fromY, toX: position.x, toY: position.y,
            holdMs: COMBAT_CUE_PROJECTILE_MS,
            projectile,
          })
        }
      }
    }
  }, [tactical.events, ships, projectiles, fighterGroups, ownSeatID, ownActivation])

  useEffect(() => {
    if (!impactAnimation) return
    const timer = window.setTimeout(() => setImpactAnimation((current) =>
      current?.sequence === impactAnimation.sequence ? null : current), DAMAGE_FEEDBACK_MS)
    return () => window.clearTimeout(timer)
  }, [impactAnimation])

  useEffect(() => {
    if (!beamAnimation) return
    const timer = window.setTimeout(() => setBeamAnimation((current) => current?.sequence === beamAnimation.sequence ? null : current), DAMAGE_FEEDBACK_MS)
    return () => window.clearTimeout(timer)
  }, [beamAnimation])

  useEffect(() => {
    if (!moveAnimation) return
    const timer = window.setTimeout(() => setMoveAnimation((current) => current?.sequence === moveAnimation.sequence ? null : current), 864)
    return () => window.clearTimeout(timer)
  }, [moveAnimation])
  useEffect(() => {
    if (!projectileAnimation) return
    const timer = window.setTimeout(() => setProjectileAnimation((current) => current?.sequence === projectileAnimation.sequence ? null : current), 1176)
    return () => window.clearTimeout(timer)
  }, [projectileAnimation])

  useEffect(() => {
    if (fighterAttackAnimations.length === 0) return
    const timer = window.setTimeout(() => setFighterAttackAnimations([]), 1260)
    return () => window.clearTimeout(timer)
  }, [fighterAttackAnimations.map((animation) => animation.sequence).join(':')])

  useEffect(() => {
    if (!shipExplosionAnimation) return
    const timer = window.setTimeout(
      () => setShipExplosionAnimation((current) => current?.sequence === shipExplosionAnimation.sequence ? null : current),
      2280,
    )
    return () => window.clearTimeout(timer)
  }, [shipExplosionAnimation])

  const runCommand = async (command: ProtocolCommand) => {
    if (commandsDisabled || commandInFlightRef.current || presentationLockedRef.current || !ownActivation) return
    commandInFlightRef.current = true
    setBusy(true)
    setCommandError('')
    try {
      await onCommand(command)
    } catch (reason) {
      setCommandError(reason instanceof Error ? reason.message : String(reason))
    } finally {
      commandInFlightRef.current = false
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
    if (ignoreClickRef.current || presentationLockedRef.current) return
    if (mode === 'scan') {
      setScannedShipID(ship.ship_id)
      focusShip(ship)
      return
    }
    if (ship.seat_id === ownSeatID) {
      // Selecting another friendly ship must not pan or zoom the player's viewport.
      setSelectedShipID(ship.ship_id)
      if (!controlsDisabled && ownActivation && activeShip && ship.ship_id !== activeShip.ship_id && waitTargetShipIDSet.has(ship.ship_id)) {
        void runCommand(tacticalWaitActivationCommand(tactical, activeShip.ship_id, ship.ship_id))
      }
      return
    }
    if (controlsDisabled || !activeShip || !selectedWeaponAction?.available) return
    const target = legalShipTargetByID.get(ship.ship_id)
    if (!target) return
    if (selectedWeaponAction.command_kind === 'battle.fire_beam') {
      void runCommand(tacticalFireBeamCommand(tactical, activeShip.ship_id, ship.ship_id, selectedWeaponAction.weapon_slot))
    } else if (selectedWeaponAction.command_kind === 'battle.fire_ordnance') {
      void runCommand(tacticalFireOrdnanceCommand(tactical, activeShip.ship_id, ship.ship_id, selectedWeaponAction.weapon_slot))
    } else if (selectedWeaponAction.command_kind === 'battle.launch_fighters') {
      void runCommand(tacticalLaunchFightersCommand(tactical, activeShip.ship_id, selectedWeaponAction.weapon_slot, target))
    }
  }
  const selectProjectile = (projectile: TacticalProjectileState) => {
    if (ignoreClickRef.current || controlsDisabled || !activeShip || !selectedWeaponAction?.available) return
    const target = legalProjectileTargetByID.get(projectile.projectile_id)
    if (!target) return
    if (selectedWeaponAction.command_kind === 'battle.fire_beam') {
      void runCommand(tacticalFireBeamProjectileCommand(tactical, activeShip.ship_id, projectile.projectile_id, selectedWeaponAction.weapon_slot))
    } else if (selectedWeaponAction.command_kind === 'battle.launch_fighters') {
      void runCommand(tacticalLaunchFightersCommand(tactical, activeShip.ship_id, selectedWeaponAction.weapon_slot, target))
    }
  }
  const selectFighterGroup = (fighter: TacticalFighterGroupState) => {
    if (ignoreClickRef.current || controlsDisabled || !activeShip || !selectedWeaponAction?.available) return
    const target = legalFighterTargetByID.get(fighter.group_id)
    if (!target) return
    if (selectedWeaponAction.command_kind === 'battle.fire_beam') {
      void runCommand(tacticalFireBeamFighterCommand(tactical, activeShip.ship_id, fighter.group_id, selectedWeaponAction.weapon_slot))
    } else if (selectedWeaponAction.command_kind === 'battle.launch_fighters') {
      void runCommand(tacticalLaunchFightersCommand(tactical, activeShip.ship_id, selectedWeaponAction.weapon_slot, target))
    }
  }
  const boardSelectedTarget = () => {
    if (controlsDisabled || !activeShip || !selectedBoardingAction) return
    const max = selectedBoardingAction.max_marine_count
    const count = clamp(Math.trunc(boardingMarineCount), 1, max)
    if (!(selectedBoardingAction.allowed_modes ?? []).includes(boardingMode)) return
    void runCommand(tacticalBoardShipCommand(tactical, activeShip.ship_id, selectedBoardingAction.target_ship_id, count, boardingMode))
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
    if (presentationLockedRef.current) return
    if (event.pointerType === 'mouse' && event.button !== 0 && event.button !== 2) return
    if (event.pointerType === 'mouse' && event.button === 2) event.preventDefault()
    if (pointersRef.current.size === 0) {
      draggedRef.current = false
      dragDistanceRef.current = 0
      ignoreClickRef.current = false
    }
    const isActiveShip = (event.target as Element).closest?.('.tactical-ship.own.active')
    if (pointersRef.current.size === 0 && event.button === 0 && movementSelectionActive
      && !controlsDisabled && legalMoves.length > 0 && isActiveShip) {
      shipDragRef.current = { pointerID: event.pointerId, startX: event.clientX, startY: event.clientY, dragging: false }
    } else if (shipDragRef.current) {
      // A second pointer is a pan/pinch gesture, not a ship movement.
      shipDragRef.current = null
    }
    pointersRef.current.set(event.pointerId, { x: event.clientX, y: event.clientY })
    gestureRef.current = gestureSnapshot(pointersRef.current)
  }

  const pointerMove = (event: ReactPointerEvent<SVGSVGElement>) => {
    if (presentationLockedRef.current || !pointersRef.current.has(event.pointerId)) return
    const shipDrag = shipDragRef.current
    if (shipDrag?.pointerID === event.pointerId) {
      pointersRef.current.set(event.pointerId, { x: event.clientX, y: event.clientY })
      gestureRef.current = gestureSnapshot(pointersRef.current)
      if (!shipDrag.dragging &&
        Math.hypot(event.clientX - shipDrag.startX, event.clientY - shipDrag.startY) >= DRAG_THRESHOLD_PX) {
        shipDrag.dragging = true
        draggedRef.current = true
        ignoreClickRef.current = true

        if (!event.currentTarget.hasPointerCapture(event.pointerId)) {
          try { event.currentTarget.setPointerCapture(event.pointerId) } catch { /* synthetic or ended pointer */ }
        }
      }
      // Dragging a ship previews its legal moves; it must not pan the camera.
      return
    }
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
    dragDistanceRef.current += Math.hypot(dx, dy)
    const moved = dragDistanceRef.current >= DRAG_THRESHOLD_PX || Math.abs(pinchFactor - 1) > PINCH_THRESHOLD
    if (moved) {
      draggedRef.current = true
      ignoreClickRef.current = true
      if (!event.currentTarget.hasPointerCapture(event.pointerId)) {
        try { event.currentTarget.setPointerCapture(event.pointerId) } catch { /* synthetic/ended pointer: capture is optional */ }
      }
    }
    if (draggedRef.current) {
      setCamera((current) => {
        const next = {
          cx: current.cx - dx * (cameraExtentRef.current.width / current.zoom) / rect.width,
          cy: current.cy - dy * (cameraExtentRef.current.height / current.zoom) / rect.height,
          zoom: clamp(current.zoom * pinchFactor, MIN_ZOOM, MAX_ZOOM),
        }
        if (ownActivation) manualCameraRef.current = next
        return next
      })
    }
    gestureRef.current = currentGesture
  }

  const pointerEnd = (event: ReactPointerEvent<SVGSVGElement>) => {
    const shipDrag = shipDragRef.current
    const shipDrop = shipDrag?.pointerID === event.pointerId && shipDrag.dragging && event.type === 'pointerup'
    if (shipDrag?.pointerID === event.pointerId) {
      shipDragRef.current = null
    }
    pointersRef.current.delete(event.pointerId)
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      try { event.currentTarget.releasePointerCapture(event.pointerId) } catch { /* capture already ended */ }
    }
    gestureRef.current = gestureSnapshot(pointersRef.current)
    if (pointersRef.current.size === 0 && draggedRef.current) {
      window.setTimeout(() => { ignoreClickRef.current = false }, 0)
    }
    if (shipDrop && movementSelectionActive && !controlsDisabled && activeShip) {
      const rect = event.currentTarget.getBoundingClientRect()
      if (rect.width > 0 && rect.height > 0) {
        const x = Math.round(viewX + ((event.clientX - rect.left) / rect.width) * viewWidth)
        const y = Math.round(viewY + ((event.clientY - rect.top) / rect.height) * viewHeight)
        const move = legalMoveByCell.get((x) + ':' + (y))
        if (move) void runCommand(tacticalMoveCommand(tactical, activeShip.ship_id, move.x, move.y))
      }
    }
  }

  const wheel = (event: ReactWheelEvent<SVGSVGElement>) => {
    event.preventDefault()
    zoomBy(Math.exp(-event.deltaY * 0.0015))
  }

  return (
    <div
      className={'tactical-battlefield tactical-screen' + (presentationLocked ? ' is-presentation-locked' : '')}
      data-presentation-locked={presentationLocked ? 'true' : 'false'}
      aria-busy={presentationLocked}
    >
      <div className="tactical-screen-status" aria-live="polite">
        <span className={`tactical-status-dot ${ownActivation ? 'is-own' : 'is-waiting'}`} />
        <strong>{presentationLocked ? t('battlefield.presentationLocked') : activeShip ? (ownActivation ? t('battlefield.yourActivation') : t('battlefield.waitingActivation', { empire: empireName(activeShip.empire_id) })) : t('battlefield.noActiveShip')}</strong>
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

          {movementSelectionActive && selectedShip && (
            <rect x={selectedShip.x - .5} y={selectedShip.y - .5} width={1} height={1} className="tactical-selected-cell" pointerEvents="none" />
          )}

          {movementSelectionActive && selectedShip && ships.filter((ship) => !ship.destroyed && ship.ship_id !== selectedShip.ship_id).map((ship) => (
            <rect key={`occupied:${ship.ship_id}`} x={ship.x - .5} y={ship.y - .5} width={1} height={1} className="tactical-occupied-cell" pointerEvents="none" />
          ))}

          {movementSelectionActive && reachableGridD && (
            <path d={reachableGridD} className="tactical-reachable-grid" pointerEvents="none" />
          )}
          {movementSelectionActive && legalMoves.map((move) => (
            <rect
              key={`move-cursor:${move.movement_kind}:${move.x}:${move.y}`}
              x={move.x - .5}
              y={move.y - .5}
              width={1}
              height={1}
              className="tactical-move-cursor-cell"
              aria-hidden="true"
            />
          ))}
          {movementSelectionActive && legalMoves.filter((move) => move.movement_kind === 'teleport').map((move) => (
            <circle key={`teleport:${move.x}:${move.y}`} cx={move.x} cy={move.y} r={0.32} className="tactical-teleport-node" pointerEvents="none" />
          ))}
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
                pointerEvents={ship.destroyed && mode !== 'scan' ? 'none' : undefined}
                onClick={(event) => { event.stopPropagation(); selectShip(ship) }}
                role="button"
                tabIndex={ship.destroyed && mode !== 'scan' ? -1 : 0}
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
                {!ship.destroyed && ship.shield_arc_current.some((value) => value > 0) && (
                  <g className="tactical-shield-field"
                    transform={'rotate(' + angle + ')'}
                    pointerEvents="none" aria-hidden="true"
                    data-shield-current={ship.shield_arc_current.reduce((sum, value) => sum + value, 0)}
                    data-shield-max={ship.shield_arc_max * 4}>
                    {TACTICAL_SHIELD_SECTORS.map((sector) => {
                      const current = Math.max(0, ship.shield_arc_current[sector.index] ?? 0)
                      const fraction = clamp(current / Math.max(1, ship.shield_arc_max), 0, 1)
                      return (
                        <g key={sector.index} className="tactical-shield-sector"
                          data-sector={sector.key} data-arc-index={sector.index}
                          data-current={current} data-max={ship.shield_arc_max}
                          data-fraction={fraction.toFixed(2)}>
                          <path d={tacticalShieldSectorPath(sector.centerDegrees, 1)}
                            className="tactical-shield-sector-track" />
                          {fraction > 0 && (
                            <path d={tacticalShieldSectorPath(sector.centerDegrees, fraction)}
                              className="tactical-shield-sector-charge" />
                          )}
                        </g>
                      )
                    })}
                  </g>
                )}
              </g>
            )
          })}

          {projectiles.map((projectile) => {
            // The authoritative snapshot already contains the projectile at its
            // final position. Do not draw a second rocket (and second engine)
            // there while its flight cue is showing the same projectile in motion.
            if (projectileAnimation?.kind === 'projectile_moved' && projectileAnimation.projectileID === projectile.projectile_id) return null
            const isOwn = projectile.source_seat_id === ownSeatID
            const isLegalTarget = mode !== 'scan' && legalProjectileTargetByID.has(projectile.projectile_id)
            const angle = projectile.facing * 22.5
            return (
              <g
                key={`projectile:${projectile.projectile_id}`}
                className={`tactical-projectile ${isOwn ? 'own' : 'enemy'} ${isLegalTarget ? 'legal-target' : ''}`}
                transform={`translate(${projectile.x} ${projectile.y})`}
                pointerEvents={mode !== 'scan' && isLegalTarget ? 'auto' : 'none'}
                onClick={(event) => { event.stopPropagation(); selectProjectile(projectile) }}
                role="button"
                tabIndex={isLegalTarget ? 0 : -1}
                aria-label={t('battlefield.projectileMarker', { id: projectile.projectile_id, weapon: humanize(projectile.weapon_id), count: projectile.multiplicity, x: projectile.x, y: projectile.y })}
                onKeyDown={(event) => { if (isLegalTarget && (event.key === 'Enter' || event.key === ' ')) selectProjectile(projectile) }}
              >
                <circle r={0.48} className="tactical-projectile-hit-cell" />
                <g transform={'rotate(' + angle + ')'} pointerEvents="none" data-glyph-count={projectile.multiplicity}>
                  {tacticalMiniFormation(projectile.multiplicity).map((glyph, index) => (
                    <g key={'projectile-glyph:' + projectile.projectile_id + ':' + index} transform={'translate(' + glyph.x + ' ' + glyph.y + ') scale(' + glyph.scale + ')'}>
                      <path d="M -0.34 -0.085 L 0.18 -0.085 L 0.42 0 L 0.18 0.085 L -0.34 0.085 Z" className="tactical-projectile-vector tactical-projectile-body" />
                      <path d="M .18 -.085 L .42 0 L .18 .085 L .08 0 Z" className="tactical-projectile-nose" />
                      <line x1={-0.12} y1={0} x2={0.2} y2={0} className="tactical-projectile-energy-spine" />
                      <path d="M -0.24 -0.085 L -0.38 -0.19 L -0.08 -0.085 Z M -0.24 0.085 L -0.38 0.19 L -0.08 0.085 Z" className="tactical-projectile-fin" />
                      <circle cx={-0.34} cy={0} r={0.038} className="tactical-projectile-engine-core" />
                      <path d="M -.34 -.06 L -.72 0 L -.34 .06 Z" className="tactical-projectile-static-plume" />
                    </g>
                  ))}
                </g>
              </g>            )
          })}

          {fighterGroups.map((fighter) => {
            const isOwn = fighter.source_seat_id === ownSeatID
            const isLegalTarget = mode !== 'scan' && legalFighterTargetByID.has(fighter.group_id)
            const angle = fighter.facing * 22.5
            return (
              <g
                key={`fighter-group:${fighter.group_id}`}
                className={`tactical-fighter-group ${isOwn ? 'own' : 'enemy'} ${isLegalTarget ? 'legal-target' : ''} is-${fighter.phase}`}
                transform={`translate(${fighter.x} ${fighter.y})`}
                pointerEvents={mode !== 'scan' && isLegalTarget ? 'auto' : 'none'}
                onClick={(event) => { event.stopPropagation(); selectFighterGroup(fighter) }}
                role="button"
                tabIndex={isLegalTarget ? 0 : -1}
                aria-label={t('battlefield.fighterGroupMarker', { id: fighter.group_id, count: fighter.multiplicity, x: fighter.x, y: fighter.y })}
                onKeyDown={(event) => { if (isLegalTarget && (event.key === 'Enter' || event.key === ' ')) selectFighterGroup(fighter) }}
              >
                <circle r={0.46} className="tactical-fighter-hit-cell" />
                <g transform={'rotate(' + angle + ')'} pointerEvents="none" data-glyph-count={fighter.multiplicity} data-fighter-kind={fighter.kind}>
                  {tacticalMiniFormation(fighter.multiplicity).map((glyph, index) => (
                    <g key={'fighter-glyph:' + fighter.group_id + ':' + index} transform={'translate(' + glyph.x + ' ' + glyph.y + ') scale(' + glyph.scale + ')'}>
                      <path d={fighterMiniGlyphPath(fighter.kind)} className={'tactical-fighter-vector is-' + fighter.kind} />
                      <path d="M .17 0 L .025 -.06 L -.09 0 L .025 .06 Z" className="tactical-fighter-cockpit" />
                      <path d="M .05 -.045 L -.17 -.14 M .05 .045 L -.17 .14" className="tactical-fighter-panel" />
                      {fighter.kind === 'bomber' ? (
                        <>
                          <circle cx={-.15} cy={-.085} r={.035} className="tactical-fighter-engine-core" />
                          <circle cx={-.15} cy={.085} r={.035} className="tactical-fighter-engine-core" />
                          <path d="M -.16 -.12 L -.48 -.08 L -.16 -.04 Z" className="tactical-fighter-exhaust" />
                          <path d="M -.16 .12 L -.48 .08 L -.16 .04 Z" className="tactical-fighter-exhaust" />
                        </>
                      ) : (
                        <>
                          <circle cx={-.17} cy={0} r={.04} className="tactical-fighter-engine-core" />
                          <path d="M -.17 -.055 L -.48 0 L -.17 .055 Z" className="tactical-fighter-exhaust" />
                        </>
                      )}
                    </g>
                  ))}
                </g>
              </g>
            )
          })}


          {fighterAttackAnimations.map((animation) => (
            <g
              key={'fighter-attack:' + animation.sequence}
              className={'tactical-fighter-attack-animation is-' + animation.payloadKind + ' ' + (animation.hit ? 'is-hit' : 'is-miss')}
              data-sequence={animation.sequence}
              data-group-id={animation.groupID}
              data-payload-kind={animation.payloadKind}
              data-weapon-id={animation.weaponID}
              data-hit={animation.hit ? 'true' : 'false'}
              data-damage={animation.damage}
              pointerEvents="none"
            >
              {animation.payloadKind === 'bomb' ? (
                <>
                  <line x1={animation.fromX} y1={animation.fromY} x2={animation.toX} y2={animation.toY} className="tactical-fighter-bomb-trail" />
                  <circle cx={animation.toX} cy={animation.toY} r={0.34} className="tactical-fighter-bomb-impact" />
                </>
              ) : (
                <>
                  <line x1={animation.fromX} y1={animation.fromY} x2={animation.toX} y2={animation.toY} className="tactical-fighter-beam-glow" />
                  <line x1={animation.fromX} y1={animation.fromY} x2={animation.toX} y2={animation.toY} className="tactical-fighter-beam-core" />
                  <circle cx={animation.toX} cy={animation.toY} r={0.24} className="tactical-fighter-beam-impact" />
                </>
              )}
              <TacticalDamageFeedback layers={animation.layers} x={animation.toX}
                y={animation.toY} shieldBroken={animation.shieldBroken} />
            </g>
          ))}

          {impactAnimation && (
            <g className="tactical-projectile-impact-animation"
              data-sequence={impactAnimation.sequence}
              data-projectile-id={impactAnimation.projectileID}
              data-target-ship-id={impactAnimation.targetShipID}
              pointerEvents="none">
              <TacticalDamageFeedback layers={impactAnimation.layers} x={impactAnimation.x}
                y={impactAnimation.y} shieldBroken={impactAnimation.shieldBroken} />
            </g>
          )}

          {moveAnimation && (
            <g className={`tactical-move-animation is-${moveAnimation.movementKind}`} data-sequence={moveAnimation.sequence} data-movement-kind={moveAnimation.movementKind} pointerEvents="none">
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
                    <g transform={'translate(' + projectileAnimation.fromX + ' ' + projectileAnimation.fromY + ')'}>
                      <g className="tactical-projectile-flight" data-multiplicity={projectileAnimation.multiplicity}>
                        <animateTransform
                          attributeName="transform"
                          type="translate"
                          from="0 0"
                          to={(projectileAnimation.toX - projectileAnimation.fromX) + ' ' + (projectileAnimation.toY - projectileAnimation.fromY)}
                          dur="984ms"
                          fill="freeze"
                        />
                        <g transform={'rotate(' + (Math.atan2(projectileAnimation.toY - projectileAnimation.fromY, projectileAnimation.toX - projectileAnimation.fromX) * 180 / Math.PI) + ')'}>
                        {tacticalMiniFormation(projectileAnimation.multiplicity).map((glyph, index) => (
                          <g key={'flight-projectile:' + projectileAnimation.projectileID + ':' + index} transform={'translate(' + glyph.x + ' ' + glyph.y + ') scale(' + glyph.scale + ')'}>
                            <path d="M -0.46 -0.1 L 0.2 -0.1 L 0.48 0 L 0.2 0.1 L -0.46 0.1 Z" className="tactical-projectile-flight-body" />
                            <path d="M .2 -.1 L .48 0 L .2 .1 L .08 0 Z" className="tactical-projectile-flight-nose" />
                            <line x1={-0.14} y1={0} x2={0.23} y2={0} className="tactical-projectile-flight-spine" />
                            <path d="M -0.28 -0.1 L -0.46 -0.24 L -0.08 -0.1 Z M -0.28 0.1 L -0.46 0.24 L -0.08 0.1 Z" className="tactical-projectile-flight-fin" />
                            <circle cx={-0.46} cy={0} r={0.052} className="tactical-projectile-flight-engine-core" />
                            <path d="M -0.47 -0.07 L -1.06 0 L -0.47 0.07 Z" className="tactical-projectile-flight-exhaust" />
                            {/* Keep the engine flare at a fixed offset from the
                                tail. The pulsing CSS glow provides flicker without
                                moving the exhaust independently of the rocket. */}
                            <circle cx={-0.62} cy={0} r={0.06} className="tactical-projectile-flight-flame" />
                          </g>
                          ))}
                        </g>
                      </g>
                    </g>
                  )}
                </>
              )}
              <circle cx={projectileAnimation.toX} cy={projectileAnimation.toY} r={0.34} className="tactical-projectile-event-pulse" />
            </g>
          )}

          {shipExplosionAnimation && (
            <g
              key={'ship-explosion:' + shipExplosionAnimation.sequence}
              className={'tactical-ship-explosion-animation ' + (shipExplosionAnimation.quantum ? 'is-quantum' : 'is-normal')}
              data-sequence={shipExplosionAnimation.sequence}
              data-ship-id={shipExplosionAnimation.shipID}
              data-quantum={shipExplosionAnimation.quantum ? 'true' : 'false'}
              data-explosion-power={shipExplosionAnimation.power}
              transform={'translate(' + shipExplosionAnimation.x + ' ' + shipExplosionAnimation.y + ')'}
              pointerEvents="none"
            >
              <circle r={0.28} className="tactical-ship-explosion-core" />
              <circle r={0.64} className="tactical-ship-explosion-halo" />
              <circle r={0.42} className="tactical-ship-explosion-shockwave" />
              <circle r={0.8} className="tactical-ship-explosion-shockwave is-secondary" />
              <g className="tactical-ship-explosion-rays">
                {Array.from({ length: 8 }, (_, index) => (
                  <line key={'ray:' + index} x1={0.22} y1={0} x2={1.24 + (index % 3) * 0.18} y2={0} transform={'rotate(' + index * 45 + ')'} />
                ))}
              </g>
              <g className="tactical-ship-explosion-debris">
                {Array.from({ length: 6 }, (_, index) => (
                  <g key={'debris:' + index} transform={'rotate(' + (index * 60 + 18) + ')'}>
                    <line x1={0.32} y1={0} x2={0.72 + (index % 2) * 0.2} y2={0} />
                  </g>
                ))}
              </g>
              {shipExplosionAnimation.quantum && shipExplosionAnimation.radiusFields > 0 && (
                <>
                  <circle r={shipExplosionAnimation.radiusFields} className="tactical-ship-explosion-quantum-field" />
                  <circle r={1.05} className="tactical-ship-explosion-quantum-core" />
                </>
              )}
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
              <TacticalDamageFeedback layers={beamAnimation.layers} x={beamAnimation.toX}
                y={beamAnimation.toY} shieldBroken={beamAnimation.shieldBroken} />
              {beamAnimation.hit && beamAnimation.shieldPiercing && (
                <text x={beamAnimation.toX} y={beamAnimation.toY - .97}
                  className="tactical-beam-shield-piercing"
                  data-shield-piercing="true">{t('battlefield.shieldPiercing')}</text>
              )}

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
                  disabled={presentationLocked}
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
                  const fighterBayState = fighterBayStateBySlot.get(weapon.slot)
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
                        <strong>S{weapon.slot + 1} {humanize(weapon.weapon_id)} ×{weapon.count}{weapon.mount_type === 'point_defense' ? ' · PD' : weapon.mount_type === 'heavy' ? ' · Heavy' : ''}</strong>
                        {tacticalWeaponCanPierceShields(weapon) && (
                          <small className="tactical-weapon-shield-piercing" data-shield-piercing="capability">
                            {t('battlefield.shieldPiercing')}
                          </small>
                        )}
                        <small>{t('battlefield.baseDamageShort', { min: weapon.min_damage, max: weapon.max_damage })}{weapon.mount_type === 'point_defense' ? ` · ${weapon.firing_arc_id === '360' ? '360°' : humanize(weapon.firing_arc_id ?? '')}` : ''} · {action?.available ? t('battlefield.targets', { count: action.targets?.length ?? 0 }) : weaponUnavailableLabel(action?.unavailable_reason, t)}{ordnanceStatus ? ` · ${ordnanceStatus}` : ''}</small>
                      </span>
                    </button>
                  )
                })}
              </div>
            </div>
          )}
          {selectedBoardingAction && (
            <div className="tactical-boarding-strip" aria-label={t('battlefield.boarding')}>
              <label>
                <span>{t('battlefield.boardingTarget')}</span>
                <select value={selectedBoardingAction.target_ship_id} disabled={controlsDisabled} onChange={(event) => setSelectedBoardingTargetID(Number(event.target.value))}>
                  {availableBoardingActions.map((action) => <option key={action.target_ship_id} value={action.target_ship_id}>{shipName(action.target_ship_id)}</option>)}
                </select>
              </label>
              <label>
                <span>{t('battlefield.boardingMode')}</span>
                <select value={boardingMode} disabled={controlsDisabled} onChange={(event) => setBoardingMode(event.target.value)}>
                  {(selectedBoardingAction.allowed_modes ?? []).map((boardingModeID) => <option key={boardingModeID} value={boardingModeID}>{humanize(boardingModeID)}</option>)}
                </select>
              </label>
              <label>
                <span>{t('battlefield.boardingMarines')}</span>
                <input type="number" min={1} max={selectedBoardingAction.max_marine_count} value={boardingMarineCount} disabled={controlsDisabled} onChange={(event) => setBoardingMarineCount(clamp(Number(event.target.value) || 1, 1, selectedBoardingAction.max_marine_count))} />
              </label>
              <button type="button" disabled={controlsDisabled} onClick={boardSelectedTarget}><GameIcon name="command" />{t('battlefield.boardingAction')}</button>
            </div>
          )}          <div className="tactical-control-commit">
            <button
              type="button"
              className={`tactical-scan-toggle ${mode === 'scan' ? 'is-active' : ''}`}
              disabled={presentationLocked}
              onClick={() => {
                if (presentationLockedRef.current) return
                setMode((current) => current === 'scan' ? 'combat' : 'scan')
                setScannedShipID(null)
              }}
            >
              <GameIcon name="info" />{t('battlefield.scan')}
            </button>
            <button type="button" className="tactical-wait-ship" disabled={!ownActivation || !tactical.can_wait_activation || controlsDisabled} onClick={waitActivation}><GameIcon name="command" />{t('battlefield.waitActivation')}</button>
            <button type="button" className="tactical-next-ship" disabled={!tactical.can_end_activation || controlsDisabled} onClick={endActivation}><GameIcon name="check" />{busy ? t('battlefield.commandBusy') : t('battlefield.finishActivation')}</button>
            <button type="button" className="tactical-retreat" disabled={controlsDisabled || !tactical.can_retreat} onClick={requestRetreat}><GameIcon name="flag" />{t('battlefield.retreat')}</button>
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
              <button type="button" className="tactical-confirm-danger" disabled={controlsDisabled || !tactical.can_retreat} onClick={confirmRetreat}>{t('battlefield.retreatConfirmAction')}</button>
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
        <div><dt>{t('battlefield.beamDefense')}</dt><dd>{ship.beam_defense}</dd></div>        <div><dt>{t('battlefield.marines')}</dt><dd>{ship.marine_current}/{ship.marine_max}</dd></div>
        <div><dt>{t('battlefield.shields')}</dt><dd>{ship.shield_arc_current.reduce((sum, value) => sum + value, 0)}/{ship.shield_arc_max * 4}</dd></div>
        {ship.shield_arc_max > 0 && (
          <div className="tactical-scan-shield-sectors">
            {TACTICAL_SHIELD_SECTORS.map((sector) => (
              <span key={sector.key} data-sector={sector.key}>
                <strong>{t(sector.label)}</strong>
                <span>{ship.shield_arc_current[sector.index]}/{ship.shield_arc_max}</span>
              </span>
            ))}
          </div>
        )}
        <div><dt>{t('battlefield.cloakState')}</dt><dd>{humanize(ship.cloak_state)}</dd></div>
      </dl>
      <div className="tactical-health-row">
        <span>{t('battlefield.armor')}</span><strong>{ship.armor_current}/{ship.armor_max}</strong>
        <div className="tactical-health-meter"><span style={{ width: `${meterPercent(ship.armor_current, ship.armor_max)}%` }} /></div>
      </div>
      <div className="tactical-health-row">
        <span>{t('battlefield.structure')}</span><strong>{ship.structure_current}/{ship.structure_max}</strong>
        <div className="tactical-health-meter"><span style={{ width: `${meterPercent(ship.structure_current, ship.structure_max)}%` }} /></div>
      </div>
      <div className="tactical-scan-internal galaxy-fleet-info-weapons">
        <strong>{t('battlefield.internalSystems')}</strong>
        <div className="tactical-weapon-row"><span>Engine</span><span>{ship.internal.engine_current}/{ship.internal.engine_max}</span></div>
        <div className="tactical-weapon-row"><span>Secondary</span><span>{ship.internal.secondary_current}/{ship.internal.secondary_max}</span></div>
        <div className="tactical-weapon-row"><span>Shield Generator</span><span>{ship.internal.shield_generator_current}/{ship.internal.shield_generator_max}</span></div>
      </div>
      <div className="tactical-scan-specials galaxy-fleet-info-weapons">
        <strong>{t('battlefield.specials')}</strong>
        {ship.specials?.length ? ship.specials.map((special) => (
          <div className="tactical-weapon-row" key={special.slot}>
            <span>{humanize(special.special_id)}</span>
            <span>{humanize(special.tactical_support)}</span>
            <span className={`badge ${special.functional ? '' : 'danger'}`}>{special.functional ? t('battlefield.functional') : t('battlefield.disabled')}</span>
          </div>
        )) : <p className="muted">—</p>}
      </div>      <div className="tactical-scan-weapons galaxy-fleet-info-weapons">
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
