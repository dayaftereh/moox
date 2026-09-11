import { useEffect, useMemo, useRef, useState, type DragEvent as ReactDragEvent, type PointerEvent as ReactPointerEvent, type WheelEvent as ReactWheelEvent } from 'react'
import {
  aggregatePopulation,
  decodeShipVisualGenome,
  type Colony,
  type ConstructionChoice,
  type ConstructionEffect,
  type ConstructionProjectKind,
  type ConstructionState,
  type DraftOrder,
  type FleetMoveTarget,
  type FleetMoveTargetReason,
  type DiplomacyCommandKind,
  type DiplomaticStance,
  type DiplomacyView,
  type PlanningMetricBreakdown,
  type OrbitalBodyKind,
  type PlanningPreviewSnapshot,
  type PlayerSnapshot,
  type PopulationJob,
  type PopulationTransferChoice,
  type ResearchChoice,
  type ResearchTechnologyEffect,
  type ResearchTechnologyInfo,
  type Ship,
  type ShipDesign,
  type StrategicContact,
  type StrategicFleet,
  type StarSystem,
} from './api'
import { ProceduralShipGlyph } from './components/ProceduralShipGlyph'
import { SpecialShipGlyph, type SpecialShipKind } from './components/SpecialShipGlyph'
import { GameIcon, type GameIconName } from './components/GameIcon'
import { OrbitalBodyArt } from './components/OrbitalBodyArt'
import { BuildingArt } from './components/BuildingArt'
import { StarArt, starPalette } from './components/StarArt'
import { Card, EmptyState, PageHeader } from './components/ui'
import { type TranslationKey, type TranslationVars } from './i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string
type GalaxyFleetPickerUnit = {
  key: string
  fleet: StrategicFleet
  ship?: Ship
}

type GalaxyFleetTargetFeedback = {
  systemID: number
  outcome: 'planned' | 'blocked' | 'unavailable' | 'cancelled'
  target?: FleetMoveTarget
  orderCount?: number
}

type GalaxyRoute = {
  key: string
  sourceSystemID: number
  destinationSystemID: number
  tone: 'planned' | 'blocked' | 'transit'
  orderCount: number
  progress?: number
  remainingTurns?: number
  remainingDistanceParsecs?: number
  fleetIDs?: number[]
  unitCount?: number
}

const populationJobIcons: Record<PopulationJob, GameIconName> = {
  farmer: 'farmer',
  worker: 'worker',
  scientist: 'scientist',
}

const researchCategoryIcons: readonly GameIconName[] = [
  'research-construction',
  'test-tube',
  'research-computer',
  'research-physics',
  'research-energy',
  'research-sociology',
  'research-biology',
  'research-force-field',
]

function researchCategoryIcon(order: number): GameIconName {
  return researchCategoryIcons[order] ?? 'research'
}

type StrategicRelationTone = 'own' | 'friendly' | 'neutral' | 'hostile'

function strategicRelationTone(ownEmpireID: number, otherEmpireID: number, diplomacy: DiplomacyView[] | undefined): StrategicRelationTone {
  if (otherEmpireID === ownEmpireID) return 'own'
  const stance = diplomacy?.find((item) => item.other_empire_id === otherEmpireID)?.stance
  if (stance === 'war') return 'hostile'
  if (stance === 'peace') return 'friendly'
  return 'neutral'
}

function strategicContactIcon(kind: StrategicContact['kind']): GameIconName {
  if (kind === 'colony') return 'colonies'
  if (kind === 'outpost') return 'outpost'
  return 'fleets'
}

function orbitalBodyIcon(kind: OrbitalBodyKind): GameIconName {
  if (kind === 'gas_giant') return 'gas-giant'
  if (kind === 'asteroid_belt') return 'asteroid-belt'
  return 'planet'
}

function specialFleetShipLabel(t: Translator, kind: SpecialShipKind): string {
  if (kind === 'colony_ship') return t('system.colonyShip')
  if (kind === 'outpost_ship') return t('system.outpostShip')
  return t('system.troopTransport')
}

function fleetTargetReasonLabel(t: Translator, reason: FleetMoveTargetReason | undefined): string {
  if (reason === 'no_supply') return t('galaxy.fleetTargetNoSupply')
  if (reason === 'out_of_fuel_range') return t('galaxy.fleetTargetOutOfRange')
  if (reason === 'invalid_eta') return t('galaxy.fleetTargetInvalidEta')
  return t('galaxy.fleetTargetUnavailable')
}

function playerColorSlotClass(slot: number | undefined): string {
  const normalized = slot && slot >= 1 && slot <= 8 ? slot : 1
  return `player-color-slot-${normalized}`
}


function galaxyFleetMarkerPosition(index: number) {
  const slot = index % 4
  const ring = Math.floor(index / 4)
  const extra = ring * 10
  if (slot === 0) return { left: 31 + extra, top: -7 - extra }
  if (slot === 1) return { left: 31 + extra, top: 32 + extra }
  if (slot === 2) return { left: -7 - extra, top: 32 + extra }
  return { left: -7 - extra, top: -7 - extra }
}
function fleetRoleIcon(fleet: StrategicFleet): GameIconName {
  if (fleet.special_kind === 'colony_ship') return 'flag'
  if (fleet.special_kind === 'outpost_ship') return 'outpost'
  if (fleet.special_kind === 'troop_transport') return 'transport'
  const role = fleet.role.toLowerCase()
  if (role.includes('scout')) return 'fleet-scout'
  if (role.includes('combat') || role.includes('military')) return 'fleet-combat'
  if (role.includes('civilian')) return 'fleet-civilian'
  return 'fleets'
}

function constructionProjectIcon(kind: ConstructionProjectKind, projectID?: string): GameIconName {
  if (kind === 'building') {
    switch (projectID) {
      case 'capitol': return 'building-capitol'
      case 'colony_base': return 'building-colony-base'
      case 'marine_barracks': return 'building-barracks'
      case 'star_base': return 'building-star-base'
      default: return 'build'
    }
  }
  switch (kind) {
    case 'housing': return 'housing'
    case 'colony_ship': return 'flag'
    case 'outpost_ship': return 'outpost'
    case 'troop_transport': return 'transport'
    case 'military_ship': return 'ship'
    case 'freighter_fleet': return 'freighter'
    case 'planetary_transformation': return 'terraform'
    default: return 'build'
  }
}

function constructionProjectUsesRichArt(kind: ConstructionProjectKind): boolean {
  return kind === 'building' || kind === 'housing' || kind === 'planetary_transformation'
}

function constructionProjectArtID(kind: ConstructionProjectKind, projectID: string): string {
  return kind === 'housing' ? 'housing' : projectID
}

function constructionProjectUsesShipArt(kind: ConstructionProjectKind): boolean {
  return kind === 'military_ship'
}

function constructionProjectUsesLargeArt(kind: ConstructionProjectKind): boolean {
  return constructionProjectUsesRichArt(kind) || constructionProjectUsesShipArt(kind)
}

function constructionShipDesign(choice: Pick<ConstructionChoice, 'project_kind' | 'ship_design_id' | 'ship_design_revision'>, shipDesigns: ShipDesign[]): ShipDesign | undefined {
  if (choice.project_kind !== 'military_ship' || !choice.ship_design_id) return undefined
  return shipDesigns.find((design) => design.id === choice.ship_design_id
    && (choice.ship_design_revision === undefined || design.revision === choice.ship_design_revision))
}

function ConstructionChoiceArt({ choice, shipDesigns, variant }: { choice: ConstructionChoice; shipDesigns: ShipDesign[]; variant: 'compact' | 'hero' }) {
  const design = constructionShipDesign(choice, shipDesigns)
  if (design) {
    return (
      <ProceduralShipGlyph
        className={variant === 'compact' ? 'construction-catalog-ship-glyph' : 'construction-project-ship-glyph'}
        seed={`${design.empire_id}:${design.id}:${design.revision}:${design.spec.strategic_picture_id}`}
        genome={decodeShipVisualGenome(design.visual_genome)}
        hullId={design.spec.hull_id}
        weaponCount={design.spec.weapons?.reduce((sum, mount) => sum + mount.count, 0) ?? 0}
      />
    )
  }
  if (constructionProjectUsesRichArt(choice.project_kind)) {
    return <BuildingArt buildingId={constructionProjectArtID(choice.project_kind, choice.project_id)} variant={variant} />
  }
  return <GameIcon name={constructionProjectIcon(choice.project_kind, choice.project_id)} />
}

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

function diplomaticStanceIcon(stance: DiplomaticStance): GameIconName {
  if (stance === 'war') return 'war'
  if (stance === 'peace') return 'peace'
  return 'neutral'
}

function diplomacyActionIcon(kind: DiplomacyCommandKind): GameIconName {
  if (kind === 'diplomacy.declare_war') return 'war'
  return 'peace'
}

function serverLabel(t: Translator, key: string | undefined, fallback: string): string {
  if (!key) return fallback
  const value = t(key as TranslationKey)
  return value === key ? fallback : value
}

function humanizeToken(value: string): string {
  return value.replace(/_/g, ' ').replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function researchSelectionModeLabel(t: Translator, mode: ResearchChoice['selection_mode']): string {
  return mode === 'choose_one'
    ? t('research.modeChooseOne')
    : mode === 'all'
      ? t('research.modeAll')
      : mode === 'fixed_one'
        ? t('research.modeFixedOne')
        : t('research.modeRepeatField')
}

function constructionEffectText(t: Translator, effect: ConstructionEffect) {
  switch (effect.kind) {
    case 'population_growth_flat': return t('construction.effectPopulationGrowth', { value: effect.value.toFixed(1) })
    case 'population_capacity_flat': return t('construction.effectPopulationCapacity', { value: effect.value.toFixed(0) })
    case 'morale_percent': return t('construction.effectMorale', { value: effect.value.toFixed(0) })
    case 'morale_barracks_relief_percent': return t('construction.effectBarracksRelief', { value: effect.value.toFixed(0) })
    case 'command_points': return t('construction.effectCommandPoints', { value: effect.value.toFixed(0) })
    default: return t('construction.effectUnknown', { kind: humanizeToken(effect.kind), value: effect.value.toFixed(1) })
  }
}

function researchEffectObjectName(t: Translator, effect: ResearchTechnologyEffect): string {
  if (!effect.id) return t('research.info.unknownEffect')
  const fallback = humanizeToken(effect.id)
  if (effect.kind === 'building_unlock') return serverLabel(t, `building.${effect.id}.name`, fallback)
  return serverLabel(t, `technology.${effect.id}.name`, fallback)
}

function researchEffectText(t: Translator, effect: ResearchTechnologyEffect): string {
  const name = researchEffectObjectName(t, effect)
  switch (effect.kind) {
    case 'building_unlock':
      return t('research.info.effectBuilding', { name, cost: effect.production_cost_pp ?? 0, maintenance: effect.maintenance_bc ?? 0 })
    case 'planetary_project_unlock':
      return t('research.info.effectPlanetaryProject', { name, cost: effect.production_cost_pp ?? 0 })
    case 'ship_drive_unlock':
      return t('research.info.effectShipDrive', { name, speed: effect.ftl_speed ?? 0 })
    case 'ship_computer_unlock':
      return t('research.info.effectShipComputer', { name })
    case 'ship_armor_unlock':
      return t('research.info.effectShipArmor', { name })
    case 'ship_shield_unlock':
      return t('research.info.effectShipShield', { name })
    case 'ship_fuel_cell_unlock':
      return t('research.info.effectFuelCell', { name, range: effect.range_parsecs ?? 0 })
    case 'population_growth_bonus':
      return t('research.info.effectGrowth', { bonus: effect.growth_bonus ?? 0 })
    case 'population_capacity_bonus':
      return t('research.info.effectCapacity', { bonus: effect.population_bonus ?? 0 })
    default:
      return t('research.info.effectGeneric', { name })
  }
}

function planetTraitLabel(t: Translator, group: 'climate' | 'size' | 'mineral' | 'gravity', id: string): string {
  const key = `planetTrait.${group}.${id}` as TranslationKey
  const translated = t(key)
  return translated === key ? humanizeToken(id) : translated
}

function formatEta(t: Translator, value: number | undefined): string {
  return value && value > 0 ? t('common.turns', { turns: value }) : t('common.noEta')
}

function metricTone(value: number): 'positive' | 'danger' | 'neutral' {
  return value > 0.0000001 ? 'positive' : value < -0.0000001 ? 'danger' : 'neutral'
}

function metricComponentLabel(t: Translator, id: string): string {
  switch (id) {
    case 'natural': return t('colony.breakdownNatural')
    case 'technology': return t('colony.breakdownTechnology')
    case 'housing': return t('colony.breakdownHousing')
    case 'cloning_center': return t('colony.breakdownCloningCenter')
    case 'starvation': return t('colony.breakdownStarvation')
    case 'population_tax': return t('colony.breakdownPopulationTax')
    case 'government': return t('colony.breakdownGovernment')
    case 'morale': return t('colony.breakdownMorale')
    case 'other': return t('colony.breakdownOther')
    default: return humanizeToken(id)
  }
}

function MetricBreakdown({ title, breakdown, unit, digits, t }: {
  title: string
  breakdown: PlanningMetricBreakdown
  unit: string
  digits: number
  t: Translator
}) {
  const components = breakdown.components ?? []
  return (
    <section className="colony-metric-breakdown">
      <header>
        <strong>{title}</strong>
        <span className={`resource-text-${metricTone(breakdown.total)}`}>{breakdown.total > 0 ? '+' : ''}{breakdown.total.toFixed(digits)} {unit}</span>
      </header>
      <div className="colony-metric-components">
        {components.length === 0 ? <span className="muted">{t('common.none')}</span> : components.map((component) => (
          <div key={component.id}>
            <span>{metricComponentLabel(t, component.id)}</span>
            <strong className={`resource-text-${metricTone(component.value)}`}>{component.value > 0 ? '+' : ''}{component.value.toFixed(digits)} {unit}</strong>
          </div>
        ))}
      </div>
    </section>
  )
}

function bodyLabel(t: Translator, kind: string): string {
  if (kind === 'gas_giant') return t('system.gasGiant')
  if (kind === 'asteroid_belt') return t('system.asteroidBelt')
  return t('system.planet')
}

export function StrategicGalaxyView({ snapshot, selectedSystemID, onSelectSystem, onCloseSystem, onOpenColony, draftOrders, onPlanOrder, onRemoveOrder, t }: {
  snapshot: PlayerSnapshot
  selectedSystemID?: number
  onSelectSystem: (systemID: number) => void
  onCloseSystem: () => void
  onOpenColony: (colonyID: number) => void
  draftOrders: DraftOrder[]
  onPlanOrder: (order: DraftOrder) => void
  onRemoveOrder: (key: string) => void
  t: Translator
}) {
  const decision = snapshot.decision
  const systems = decision?.strategic.galaxy.systems ?? []
  const visitedSystemIDs = useMemo(() => {
    const projected = decision?.strategic.visited_system_ids
    if (projected) return new Set(projected)
    // Compatibility while a pre-visited-authority server is still running.
    return new Set(systems.filter((system) => system.name !== '' && (system.planets?.length ?? 0) > 0).map((system) => system.id))
  }, [decision?.strategic.visited_system_ids, systems])
  const [fleetPicker, setFleetPicker] = useState<{
    systemID: number
    selectedUnitKeys: string[]
  } | null>(null)
  const [fleetPickerPosition, setFleetPickerPosition] = useState<{ x: number; y: number } | null>(null)
  const fleetPickerRef = useRef<HTMLElement | null>(null)
  const fleetPickerDragRef = useRef<{
    pointerID: number
    offsetX: number
    offsetY: number
    startX: number
    startY: number
    moved: boolean
  } | null>(null)
  const [fleetTargetFeedback, setFleetTargetFeedback] = useState<GalaxyFleetTargetFeedback | null>(null)
  const [transitFleetInfoKey, setTransitFleetInfoKey] = useState<string | null>(null)
  const [fleetInfoUnitKey, setFleetInfoUnitKey] = useState<string | null>(null)
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
  const mapPointForSystem = (systemID: number) => {
    const system = systems.find((candidate) => candidate.id === systemID)
    if (!system) return undefined
    return {
      x: 8 + ((system.x - bounds.minX) / spanX) * 84,
      y: 8 + ((system.y - bounds.minY) / spanY) * 84,
    }
  }
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

  useEffect(() => {
    if (!fleetPicker || !decision) return
    const stillAvailable = (decision.strategic.fleets ?? []).some((fleet) => (
      fleet.empire_id === decision.empire.id
      && fleet.at_system_id === fleetPicker.systemID
    ))
    if (!stillAvailable) {
      setFleetPicker(null)
      setFleetPickerPosition(null)
    }
  }, [decision, fleetPicker])
  if (!decision) {
    return (
      <>
        <PageHeader eyebrow={t('galaxy.eyebrow')} title={t('galaxy.title')} subtitle={t('galaxy.subtitle')} />
        <EmptyState title={t('galaxy.mapPending')} body={t('common.notAvailableYet')} />
      </>
    )
  }
  const selected = systems.find((system) => system.id === selectedSystemID)
  const selectedVisited = selected ? visitedSystemIDs.has(selected.id) : false

  const changeZoom = (factor: number) => setZoom((current) => Math.round(clampZoom(current * factor) * 100) / 100)

  const pickerSystem = fleetPicker ? systems.find((system) => system.id === fleetPicker.systemID) : undefined
  const pickerFleets = fleetPicker
    ? (decision.strategic.fleets ?? [])
      .filter((fleet) => fleet.empire_id === decision.empire.id && fleet.at_system_id === fleetPicker.systemID)
      .slice()
      .sort((a, b) => a.id - b.id)
    : []
  const shipsByID = new Map((decision.strategic.ships ?? []).map((ship) => [ship.id, ship]))
  const pickerUnits: GalaxyFleetPickerUnit[] = pickerFleets.flatMap((fleet): GalaxyFleetPickerUnit[] => {
    if (fleet.special_kind) return [{ key: `fleet:${fleet.id}`, fleet, ship: undefined }]
    return (fleet.ship_ids ?? []).flatMap((shipID) => {
      const ship = shipsByID.get(shipID)
      return ship ? [{ key: `ship:${ship.id}`, fleet, ship }] : []
    })
  })
  const pickerSelectedUnitKeys = new Set(fleetPicker?.selectedUnitKeys ?? [])
  const pickerSelectedUnits = pickerUnits.filter((unit) => pickerSelectedUnitKeys.has(unit.key))
  const fleetInfoUnit = pickerUnits.find((unit) => unit.key === fleetInfoUnitKey)
  const fleetInfoLabel = fleetInfoUnit?.ship
    ? fleetInfoUnit.ship.name
    : fleetInfoUnit
      ? specialFleetShipLabel(t, fleetInfoUnit.fleet.special_kind as SpecialShipKind)
      : ''
  const pickerAllUnitsSelected = pickerUnits.length > 0 && pickerSelectedUnits.length === pickerUnits.length
  const pickerProfiles = pickerFleets.flatMap((fleet) => {
    const selected = pickerSelectedUnits.filter((unit) => unit.fleet.id === fleet.id)
    if (selected.length === 0) return []
    if (fleet.special_kind) return [{ fleetID: fleet.id, shipIDs: [] as number[] | null }]
    const allShipIDs = [...(fleet.ship_ids ?? [])].sort((a, b) => a - b)
    const selectedShipIDs = selected.flatMap((unit) => unit.ship ? [unit.ship.id] : []).sort((a, b) => a - b)
    if (selectedShipIDs.length === allShipIDs.length && selectedShipIDs.every((shipID, index) => shipID === allShipIDs[index])) {
      return [{ fleetID: fleet.id, shipIDs: [] as number[] | null }]
    }
    if (selectedShipIDs.length === 1) return [{ fleetID: fleet.id, shipIDs: selectedShipIDs as number[] | null }]
    return [{ fleetID: fleet.id, shipIDs: null as number[] | null }]
  })
  const pickerProfilesSupported = pickerProfiles.length > 0 && pickerProfiles.every((profile) => profile.shipIDs !== null)
  const targetsForPickerProfile = (profile: { fleetID: number; shipIDs: number[] | null }) => {
    const profileShipIDs = profile.shipIDs
    if (profileShipIDs === null) return []
    return (decision.decisions.fleet_move_targets ?? []).filter((target) => {
      if (target.fleet_id !== profile.fleetID) return false
      const targetShipIDs = target.ship_ids ?? []
      return targetShipIDs.length === profileShipIDs.length
        && targetShipIDs.every((shipID, index) => shipID === profileShipIDs[index])
    })
  }
  const pickerTargetGroups = pickerProfilesSupported ? pickerProfiles.map(targetsForPickerProfile) : []
  const pickerCommonDestinationIDs = pickerTargetGroups.length > 0
    ? pickerTargetGroups[0]
      .map((target) => target.destination_system_id)
      .filter((destinationID) => pickerTargetGroups.every((group) => group.some((target) => target.destination_system_id === destinationID)))
    : []
  const pickerReachableTargets = pickerCommonDestinationIDs.filter((destinationID) => (
    pickerTargetGroups.every((group) => group.find((target) => target.destination_system_id === destinationID)?.legal === true)
  )).length
  const pickerTotalTargets = pickerCommonDestinationIDs.length
  const pickerTargetsByDestination = new Map<number, FleetMoveTarget[]>()
  if (pickerProfilesSupported) {
    for (const destinationSystemID of pickerCommonDestinationIDs) {
      const targets = pickerTargetGroups
        .map((group) => group.find((target) => target.destination_system_id === destinationSystemID))
        .filter((target): target is FleetMoveTarget => Boolean(target))
      if (targets.length === pickerProfiles.length) pickerTargetsByDestination.set(destinationSystemID, targets)
    }
  }
  const fleetTargetVisualActive = Boolean(fleetPicker)
    && pickerProfilesSupported
    && pickerSelectedUnits.length > 0
  const fleetTargetFeedbackSystem = fleetTargetFeedback
    ? systems.find((system) => system.id === fleetTargetFeedback.systemID)
    : undefined
  const fleetTargetFeedbackLabel = fleetTargetFeedbackSystem && visitedSystemIDs.has(fleetTargetFeedbackSystem.id)
    ? fleetTargetFeedbackSystem.name
    : t('galaxy.unknownStar')

  const plannedRouteMap = new Map<string, GalaxyRoute>()
  for (const order of draftOrders) {
    if (order.kind !== 'empire.move_fleet') continue
    const fleetID = Number(order.payload.fleet_id)
    const destinationSystemID = Number(order.payload.destination_system_id)
    if (!Number.isInteger(fleetID) || fleetID <= 0 || !Number.isInteger(destinationSystemID) || destinationSystemID <= 0) continue
    const fleet = (decision.strategic.fleets ?? []).find((candidate) => candidate.id === fleetID && candidate.empire_id === decision.empire.id)
    const sourceSystemID = fleet?.at_system_id
    if (!sourceSystemID) continue
    const key = `${sourceSystemID}:${destinationSystemID}`
    const existing = plannedRouteMap.get(key)
    if (existing) existing.orderCount += 1
    else plannedRouteMap.set(key, { key: `planned:${key}`, sourceSystemID, destinationSystemID, tone: 'planned', orderCount: 1 })
  }
  const plannedRoutes = Array.from(plannedRouteMap.values())

  const blockedRoute: GalaxyRoute | undefined = fleetPicker && fleetTargetFeedback?.outcome === 'blocked' && fleetTargetFeedback.target
    ? {
      key: `blocked:${fleetTargetFeedback.target.source_system_id}:${fleetTargetFeedback.target.destination_system_id}`,
      sourceSystemID: fleetTargetFeedback.target.source_system_id,
      destinationSystemID: fleetTargetFeedback.target.destination_system_id,
      tone: 'blocked',
      orderCount: pickerProfiles.length,
    }
    : undefined

  const transitRouteMap = new Map<string, GalaxyRoute>()
  for (const fleet of decision.strategic.fleets ?? []) {
    if (fleet.empire_id !== decision.empire.id || fleet.at_system_id || !fleet.source_system_id || !fleet.destination_system_id || !fleet.remaining_turns) continue
    const total = fleet.transit_turns_total ?? 0
    const progress = total > 0 ? Math.max(0, Math.min(1, (total - fleet.remaining_turns) / total)) : 0.5
    const key = `${fleet.source_system_id}:${fleet.destination_system_id}:${fleet.remaining_turns}:${total}`
    const unitCount = fleet.special_kind ? 1 : Math.max(1, fleet.ship_ids?.length ?? 0)
    const existing = transitRouteMap.get(key)
    if (existing) {
      existing.orderCount += 1
      existing.unitCount = (existing.unitCount ?? 0) + unitCount
      existing.fleetIDs = [...(existing.fleetIDs ?? []), fleet.id]
      existing.remainingDistanceParsecs = Math.max(existing.remainingDistanceParsecs ?? 0, fleet.remaining_distance_parsecs ?? 0)
    } else {
      transitRouteMap.set(key, {
        key: `transit:${key}`,
        sourceSystemID: fleet.source_system_id,
        destinationSystemID: fleet.destination_system_id,
        tone: 'transit',
        orderCount: 1,
        progress,
        remainingTurns: fleet.remaining_turns,
        remainingDistanceParsecs: fleet.remaining_distance_parsecs,
        fleetIDs: [fleet.id],
        unitCount,
      })
    }
  }
  const transitRoutes = Array.from(transitRouteMap.values())
  const selectedTransitRoute = transitRoutes.find((route) => route.key === transitFleetInfoKey)
  const selectedTransitDestination = selectedTransitRoute
    ? systems.find((system) => system.id === selectedTransitRoute.destinationSystemID)
    : undefined
  const selectedTransitDestinationLabel = selectedTransitDestination && visitedSystemIDs.has(selectedTransitDestination.id)
    ? selectedTransitDestination.name
    : t('galaxy.unknownStar')
  const routeGeometry = (route: GalaxyRoute) => {
    const source = mapPointForSystem(route.sourceSystemID)
    const destination = mapPointForSystem(route.destinationSystemID)
    if (!source || !destination) return undefined
    const progress = route.tone === 'transit' ? Math.max(0, Math.min(1, route.progress ?? 0)) : 0
    const current = {
      x: source.x + (destination.x - source.x) * progress,
      y: source.y + (destination.y - source.y) * progress,
    }
    return { source, current, destination }
  }

  const planFleetDestination = (destinationSystemID: number): boolean => {
    if (!fleetPicker || !pickerProfilesSupported || pickerSelectedUnits.length === 0) return false
    if (destinationSystemID === fleetPicker.systemID) {
      let removed = 0
      for (const profile of pickerProfiles) {
        const key = `fleet:${profile.fleetID}`
        if (draftOrders.some((order) => order.key === key && order.kind === 'empire.move_fleet')) removed += 1
        onRemoveOrder(key)
      }
      setFleetTargetFeedback({ systemID: destinationSystemID, outcome: 'cancelled', orderCount: removed })
      return true
    }
    const targets = pickerTargetsByDestination.get(destinationSystemID)
    if (!targets || targets.length !== pickerProfiles.length) {
      setFleetTargetFeedback({ systemID: destinationSystemID, outcome: 'unavailable' })
      return true
    }
    const blocked = targets.find((target) => !target.legal)
    if (blocked) {
      setFleetTargetFeedback({ systemID: destinationSystemID, outcome: 'blocked', target: blocked })
      return true
    }
    for (const target of targets) {
      onPlanOrder({
        key: `fleet:${target.fleet_id}`,
        kind: 'empire.move_fleet',
        payload: {
          fleet_id: target.fleet_id,
          destination_system_id: target.destination_system_id,
          ...(target.ship_ids?.length ? { ship_ids: [...target.ship_ids] } : {}),
        },
      })
    }
    setFleetTargetFeedback({
      systemID: destinationSystemID,
      outcome: 'planned',
      target: targets[0],
      orderCount: targets.length,
    })
    return true
  }

  const clampFleetPickerPosition = (x: number, y: number) => {
    const rect = fleetPickerRef.current?.getBoundingClientRect()
    const width = rect?.width ?? Math.min(292, Math.max(220, window.innerWidth * 0.76))
    const height = rect?.height ?? Math.min(360, Math.max(220, window.innerHeight * 0.52))
    const margin = 8
    return {
      x: Math.max(margin, Math.min(Math.max(margin, window.innerWidth - width - margin), x)),
      y: Math.max(margin, Math.min(Math.max(margin, window.innerHeight - height - margin), y)),
    }
  }

  const openFleetPicker = (systemID: number, fleets: StrategicFleet[], clientX: number, clientY: number) => {
    const ordered = fleets.slice().sort((a, b) => a.id - b.id)
    const unitKeys = ordered.flatMap((fleet) => (
      fleet.special_kind
        ? [`fleet:${fleet.id}`]
        : (fleet.ship_ids ?? []).map((shipID) => `ship:${shipID}`)
    ))
    if (unitKeys.length === 0) return
    onCloseSystem()
    setFleetPicker({ systemID, selectedUnitKeys: unitKeys })
    setFleetTargetFeedback(null)
    setFleetInfoUnitKey(null)
    const width = Math.min(292, Math.max(220, window.innerWidth * 0.76))
    const height = Math.min(340, Math.max(220, window.innerHeight * 0.5))
    const x = clientX + 14 + width <= window.innerWidth ? clientX + 14 : clientX - width - 14
    const y = Math.min(Math.max(8, clientY - 34), Math.max(8, window.innerHeight - height - 8))
    setFleetPickerPosition({ x: Math.max(8, x), y })
  }

  const togglePickerUnit = (unitKey: string) => {
    setFleetTargetFeedback(null)
    setFleetInfoUnitKey(null)
    setFleetPicker((current) => {
      if (!current) return current
      const selected = new Set(current.selectedUnitKeys)
      if (selected.has(unitKey)) selected.delete(unitKey)
      else selected.add(unitKey)
      return { ...current, selectedUnitKeys: Array.from(selected) }
    })
  }

  const startFleetPickerDrag = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (event.button !== 0) return
    const panel = fleetPickerRef.current
    if (!panel) return
    const rect = panel.getBoundingClientRect()
    fleetPickerDragRef.current = {
      pointerID: event.pointerId,
      offsetX: event.clientX - rect.left,
      offsetY: event.clientY - rect.top,
      startX: event.clientX,
      startY: event.clientY,
      moved: false,
    }
    setFleetPickerPosition({ x: rect.left, y: rect.top })
    event.currentTarget.setPointerCapture(event.pointerId)
    event.preventDefault()
  }

  const moveFleetPickerDrag = (event: ReactPointerEvent<HTMLDivElement>) => {
    const drag = fleetPickerDragRef.current
    if (!drag || drag.pointerID !== event.pointerId) return
    if (!drag.moved && Math.hypot(event.clientX - drag.startX, event.clientY - drag.startY) >= 6) {
      drag.moved = true
      setFleetTargetFeedback(null)
    }
    if (!drag.moved) return
    setFleetPickerPosition(clampFleetPickerPosition(event.clientX - drag.offsetX, event.clientY - drag.offsetY))
  }

  const finishFleetPickerDrag = (event: ReactPointerEvent<HTMLDivElement>) => {
    const drag = fleetPickerDragRef.current
    if (!drag || drag.pointerID !== event.pointerId) return
    fleetPickerDragRef.current = null
    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId)
    if (!drag.moved) return
    const targetElement = document.elementsFromPoint(event.clientX, event.clientY)
      .map((element) => element.closest<HTMLElement>('[data-galaxy-system-id]'))
      .find((element): element is HTMLElement => Boolean(element))
    const destinationSystemID = Number(targetElement?.dataset.galaxySystemId)
    if (Number.isInteger(destinationSystemID) && destinationSystemID > 0) {
      planFleetDestination(destinationSystemID)
    }
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
          className={'galaxy-map' + (dragging ? ' galaxy-map-dragging' : '') + (fleetTargetVisualActive ? ' galaxy-map-targeting' : '')}
          role="list"
          aria-label={t('galaxy.mapTitle')}
          onWheel={handleWheel}
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={finishPointer}
          onPointerCancel={finishPointer}
        >
          <div className="galaxy-map-layer" style={{ transform: 'translate3d(' + pan.x + 'px, ' + pan.y + 'px, 0) scale(' + zoom + ')' }}>
            <svg className="galaxy-route-layer" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
              {plannedRoutes.map((route) => {
                const geometry = routeGeometry(route)
                if (!geometry) return null
                return (
                  <line
                    key={route.key}
                    className="galaxy-route-line is-planned"
                    data-route-tone="planned"
                    data-route-source-system-id={route.sourceSystemID}
                    data-route-destination-system-id={route.destinationSystemID}
                    x1={geometry.source.x}
                    y1={geometry.source.y}
                    x2={geometry.destination.x}
                    y2={geometry.destination.y}
                  />
                )
              })}
              {blockedRoute && (() => {
                const geometry = routeGeometry(blockedRoute)
                if (!geometry) return null
                return (
                  <line
                    key={blockedRoute.key}
                    className="galaxy-route-line is-blocked"
                    data-route-tone="blocked"
                    data-route-source-system-id={blockedRoute.sourceSystemID}
                    data-route-destination-system-id={blockedRoute.destinationSystemID}
                    x1={geometry.source.x}
                    y1={geometry.source.y}
                    x2={geometry.destination.x}
                    y2={geometry.destination.y}
                  />
                )
              })()}
              {transitRoutes.map((route) => {
                const geometry = routeGeometry(route)
                if (!geometry) return null
                return (
                  <line
                    key={route.key}
                    className="galaxy-route-line is-transit"
                    data-route-tone="transit"
                    data-route-source-system-id={route.sourceSystemID}
                    data-route-destination-system-id={route.destinationSystemID}
                    x1={geometry.current.x}
                    y1={geometry.current.y}
                    x2={geometry.destination.x}
                    y2={geometry.destination.y}
                  />
                )
              })}
            </svg>
            {transitRoutes.map((route) => {
              const geometry = routeGeometry(route)
              if (!geometry) return null
              return (
                <button
                  key={'marker-' + route.key}
                  type="button"
                  className={`galaxy-transit-fleet-marker ${playerColorSlotClass(decision.empire.player_color_slot)}${transitFleetInfoKey === route.key ? ' is-open' : ''}`}
                  data-galaxy-transit-route={route.key}
                  style={{ left: geometry.current.x + '%', top: geometry.current.y + '%' }}
                  aria-label={t('galaxy.transitFleetOpen', { count: route.unitCount ?? route.orderCount, eta: route.remainingTurns ?? 0 })}
                  title={t('galaxy.transitFleetOpen', { count: route.unitCount ?? route.orderCount, eta: route.remainingTurns ?? 0 })}
                  onPointerDown={(event) => event.stopPropagation()}
                  onClick={(event) => {
                    event.stopPropagation()
                    setTransitFleetInfoKey((current) => current === route.key ? null : route.key)
                  }}
                >
                  <GameIcon name="fleets" />
                  <span aria-hidden="true">{route.remainingTurns}</span>
                </button>
              )
            })}
            {systems.map((system) => {
              const isVisited = visitedSystemIDs.has(system.id)
              const ownsColony = (system.planets ?? []).some((planet) => decision.colonies.some((colony) => colony.planet_id === planet.id))
              const ownOutpost = (system.bodies ?? []).some((body) => Boolean(body.outpost_id) && decision.strategic.outposts?.some((outpost) => outpost.id === body.outpost_id && outpost.empire_id === decision.empire.id))
              const ownFleets = (decision.strategic.fleets ?? []).filter((fleet) => fleet.empire_id === decision.empire.id && fleet.at_system_id === system.id)
              const ownFleetUnits = ownFleets.reduce((sum, fleet) => sum + (fleet.special_kind ? 1 : Math.max(1, fleet.ship_ids?.length ?? 0)), 0)
              const foreignFleetEmpireIDs = Array.from(new Set(
                (decision.strategic.contacts ?? [])
                  .filter((contact) => contact.system_id === system.id && contact.empire_id !== decision.empire.id && contact.kind === 'fleet')
                  .map((contact) => contact.empire_id),
              )).sort((a, b) => a - b)
              const fleetEmpireMarkers = [
                ...(ownFleets.length > 0 ? [{ empireID: decision.empire.id, own: true }] : []),
                ...foreignFleetEmpireIDs.map((empireID) => ({ empireID, own: false })),
              ]
              const fleetTargetBundle = fleetTargetVisualActive ? pickerTargetsByDestination.get(system.id) : undefined
              const fleetTargetOrigin = fleetTargetVisualActive && fleetPicker?.systemID === system.id
              const fleetTargetCandidate = Boolean(fleetTargetBundle?.length)
              const fleetTargetLegal = fleetTargetCandidate && fleetTargetBundle?.every((target) => target.legal) === true
              const fleetTargetReference = fleetTargetBundle?.find((target) => !target.legal) ?? fleetTargetBundle?.[0]
              const normalSystemTitle = isVisited
                ? system.name + ' (' + system.x + ', ' + system.y + ')'
                : t('galaxy.unvisitedTitle')
              const fleetTargetTitle = fleetTargetOrigin
                ? t('galaxy.fleetTargetStay')
                : fleetTargetCandidate && fleetTargetReference
                  ? `${isVisited ? system.name : t('galaxy.unknownStar')} · ${fleetTargetReference.distance_parsecs} pc · ${t('galaxy.fleetTargetRange', { range: fleetTargetReference.fuel_range_parsecs })} · ${t('galaxy.fleetTargetEta', { eta: fleetTargetReference.eta })}`
                  : normalSystemTitle
              return (
                <div
                  key={system.id}
                  role="listitem"
                  data-galaxy-system-id={system.id}
                  className="galaxy-node-cluster"
                  style={{
                    left: (8 + ((system.x - bounds.minX) / spanX) * 84) + '%',
                    top: (8 + ((system.y - bounds.minY) / spanY) * 84) + '%',
                  }}
                >
                  <button
                    type="button"
                    className={'galaxy-node' + (selectedSystemID === system.id ? ' galaxy-node-selected' : '') + (ownsColony ? ' galaxy-node-colony' : ownOutpost ? ' galaxy-node-outpost' : '') + (ownFleetUnits > 0 ? ' galaxy-node-fleet' : '') + (fleetTargetOrigin ? ' galaxy-node-target-origin' : fleetTargetCandidate ? (fleetTargetLegal ? ' galaxy-node-target-legal' : ' galaxy-node-target-blocked') : '')}
                    aria-label={isVisited ? system.name : t('galaxy.unknownStar')}
                    onClick={() => {
                      if (ignoreClickRef.current) return
                      if (fleetTargetVisualActive && planFleetDestination(system.id)) return
                      onSelectSystem(system.id)
                    }}
                    title={fleetTargetTitle}
                  >
                    <span className="galaxy-star" aria-hidden="true"><StarArt spectralClass={system.spectral_class} seed={system.id} /></span>
                    {isVisited && (
                      <span className={'galaxy-node-label' + (ownsColony ? ' galaxy-node-label-owned-colony' : '')} aria-hidden="true">
                        {system.name}
                      </span>
                    )}
                  </button>
                  {fleetEmpireMarkers.length > 0 && (
                    <span className="galaxy-node-fleet-ring">
                      {fleetEmpireMarkers.map((marker, index) => {
                        const identity = marker.own
                          ? decision.empire
                          : decision.public_empires?.find((empire) => empire.id === marker.empireID)
                        const colorSlot = marker.own ? decision.empire.player_color_slot : identity?.player_color_slot
                        const markerClass = `galaxy-node-marker galaxy-node-marker-player-fleet ${playerColorSlotClass(colorSlot)}`
                        if (marker.own) {
                          return (
                            <button
                              type="button"
                              key={`fleet-empire-${marker.empireID}`}
                              className={markerClass + ' galaxy-node-marker-button' + (fleetPicker?.systemID === system.id ? ' is-open' : '')}
                              style={galaxyFleetMarkerPosition(index)}
                              aria-label={t('galaxy.fleetPickerOpen', { count: ownFleetUnits })}
                              aria-expanded={fleetPicker?.systemID === system.id}
                              aria-controls={fleetPicker?.systemID === system.id ? 'galaxy-fleet-picker' : undefined}
                              title={t('galaxy.ownFleetMarker')}
                              onPointerDown={(event) => event.stopPropagation()}
                              onClick={(event) => {
                                event.stopPropagation()
                                if (ignoreClickRef.current) return
                                openFleetPicker(system.id, ownFleets, event.clientX, event.clientY)
                              }}
                            >
                              <GameIcon name="fleets" />
                            </button>
                          )
                        }
                        const empireName = identity?.name ?? t('common.empireFallback', { id: marker.empireID })
                        return (
                          <span
                            key={`fleet-empire-${marker.empireID}`}
                            className={markerClass}
                            style={galaxyFleetMarkerPosition(index)}
                            role="img"
                            aria-label={t('galaxy.foreignFleetMarker', { empire: empireName })}
                            title={t('galaxy.foreignFleetMarker', { empire: empireName })}
                          >
                            <GameIcon name="fleets" />
                          </span>
                        )
                      })}
                    </span>
                  )}
                </div>
              )
            })}
          </div>
          {selectedTransitRoute && (
            <aside className="galaxy-transit-fleet-popover" role="status">
              <div>
                <small>{t('galaxy.transitFleetTitle')}</small>
                <strong>{selectedTransitDestinationLabel}</strong>
                <span>{t('galaxy.transitFleetSummary', { count: selectedTransitRoute.unitCount ?? selectedTransitRoute.orderCount, distance: selectedTransitRoute.remainingDistanceParsecs ?? '?', eta: selectedTransitRoute.remainingTurns ?? 0 })}</span>
                <small>{t('galaxy.transitFleetLocked')}</small>
              </div>
              <button type="button" aria-label={t('common.close')} title={t('common.close')} onClick={() => setTransitFleetInfoKey(null)}>
                <GameIcon name="close" />
              </button>
            </aside>
          )}
        </div>
      </Card>
      {fleetPicker && pickerSystem && pickerUnits.length > 0 && (
        <aside
          id="galaxy-fleet-picker"
          ref={fleetPickerRef}
          className="galaxy-fleet-picker galaxy-fleet-picker-compact"
          role="dialog"
          aria-label={t('galaxy.fleetPickerTitle')}
          style={fleetPickerPosition ? { left: fleetPickerPosition.x, top: fleetPickerPosition.y, right: 'auto', bottom: 'auto' } : undefined}
        >
          <div className="galaxy-fleet-picker-header">
            <div
              className="galaxy-fleet-picker-drag-handle"
              title={t('galaxy.fleetPickerDrag')}
              onPointerDown={startFleetPickerDrag}
              onPointerMove={moveFleetPickerDrag}
              onPointerUp={finishFleetPickerDrag}
              onPointerCancel={finishFleetPickerDrag}
            >
              <span>
                <small>{pickerSystem.name || t('galaxy.unknownStar')}</small>
                <strong>{t('galaxy.fleetPickerTitle')} · {pickerUnits.length}</strong>
              </span>
            </div>
            <button
              type="button"
              className="galaxy-fleet-picker-close"
              aria-label={t('common.close')}
              title={t('common.close')}
              onPointerDown={(event) => event.stopPropagation()}
              onClick={() => {
                setFleetPicker(null)
                setFleetPickerPosition(null)
                            setFleetTargetFeedback(null)
                setFleetInfoUnitKey(null)
              }}
            >
              <GameIcon name="close" />
            </button>
          </div>

          <div className="galaxy-fleet-picker-body">
            <div className="galaxy-fleet-picker-selection-head">
              <small>{t('galaxy.fleetPickerSelected', { selected: pickerSelectedUnits.length, total: pickerUnits.length })}</small>
              <button
                type="button"
                className="ghost-button"
                disabled={pickerAllUnitsSelected}
                onClick={() => {
                  setFleetTargetFeedback(null)
                  setFleetInfoUnitKey(null)
                  setFleetPicker((current) => current ? {
                    ...current,
                    selectedUnitKeys: pickerUnits.map((unit) => unit.key),
                  } : current)
                }}
              >
                {t('galaxy.fleetPickerSelectAll')}
              </button>
            </div>
            <div className="galaxy-fleet-picker-ships galaxy-fleet-picker-grid">
              {pickerUnits.map((unit) => {
                const isSelected = pickerSelectedUnitKeys.has(unit.key)
                const label = unit.ship
                  ? unit.ship.name
                  : specialFleetShipLabel(t, unit.fleet.special_kind as SpecialShipKind)
                const title = unit.ship
                  ? `${label} · ${humanizeToken(unit.ship.spec.hull_id)} · R${unit.ship.source_design_revision}`
                  : `${label} · ${t('galaxy.fleet')} #${unit.fleet.id}`
                return (
                  <div className="galaxy-fleet-picker-tile-wrap" key={unit.key}>
                    <button
                      type="button"
                      className={'galaxy-fleet-picker-ship galaxy-fleet-picker-tile' + (isSelected ? ' is-selected' : '')}
                      aria-pressed={isSelected}
                      aria-label={label}
                      title={title}
                      onClick={() => togglePickerUnit(unit.key)}
                    >
                      {unit.ship ? (
                        <ProceduralShipGlyph
                          className="galaxy-fleet-picker-ship-glyph"
                          seed={`${unit.ship.empire_id}:${unit.ship.source_design_id}:${unit.ship.source_design_revision}:${unit.ship.spec.strategic_picture_id}`}
                          genome={decodeShipVisualGenome(unit.ship.visual_genome)}
                          hullId={unit.ship.spec.hull_id}
                          weaponCount={unit.ship.spec.weapons?.reduce((sum, mount) => sum + mount.count, 0) ?? 0}
                        />
                      ) : (
                        <SpecialShipGlyph
                          className="galaxy-fleet-picker-ship-glyph galaxy-fleet-picker-special-ship-glyph"
                          kind={unit.fleet.special_kind as SpecialShipKind}
                        />
                      )}
                      <span className="galaxy-fleet-picker-ship-copy"><strong>{label}</strong></span>
                      <span className="galaxy-fleet-picker-ship-check" aria-hidden="true">{isSelected && <GameIcon name="check" />}</span>
                    </button>
                    <button
                      type="button"
                      className={'galaxy-fleet-picker-ship-info' + (fleetInfoUnitKey === unit.key ? ' is-open' : '')}
                      aria-label={t('galaxy.fleetInfoOpen', { ship: label })}
                      aria-pressed={fleetInfoUnitKey === unit.key}
                      title={t('galaxy.fleetInfoOpen', { ship: label })}
                      onPointerDown={(event) => event.stopPropagation()}
                      onClick={(event) => {
                        event.stopPropagation()
                        setFleetTargetFeedback(null)
                        setFleetInfoUnitKey((current) => current === unit.key ? null : unit.key)
                      }}
                    >
                      ?
                    </button>
                  </div>
                )
              })}
            </div>
            {fleetInfoUnit && (
              <section className="galaxy-fleet-info-popover" role="dialog" aria-label={t('galaxy.fleetInfoTitle', { ship: fleetInfoLabel })}>
                <header className="galaxy-fleet-info-header">
                  <div>
                    <small>{t('galaxy.fleetInfoHeading')}</small>
                    <strong>{fleetInfoLabel}</strong>
                  </div>
                  <button
                    type="button"
                    className="galaxy-fleet-info-close"
                    aria-label={t('common.close')}
                    title={t('common.close')}
                    onClick={() => setFleetInfoUnitKey(null)}
                  >
                    <GameIcon name="close" />
                  </button>
                </header>
                {fleetInfoUnit.ship ? (
                  <div className="galaxy-fleet-info-content">
                    <dl className="galaxy-fleet-info-grid">
                      <div><dt>{t('system.hull')}</dt><dd>{humanizeToken(fleetInfoUnit.ship.spec.hull_id)}</dd></div>
                      <div><dt>{t('system.warpDrive')}</dt><dd>{humanizeToken(fleetInfoUnit.ship.spec.warp_drive_id)}</dd></div>
                      <div><dt>{t('system.ftlSpeed')}</dt><dd>{fleetInfoUnit.ship.spec.ftl_speed}</dd></div>
                      <div><dt>{t('system.computer')}</dt><dd>{humanizeToken(fleetInfoUnit.ship.spec.computer_id)}</dd></div>
                      <div><dt>{t('system.armor')}</dt><dd>{humanizeToken(fleetInfoUnit.ship.spec.armor_id)}</dd></div>
                      <div><dt>{t('system.shield')}</dt><dd>{fleetInfoUnit.ship.spec.shield_id ? humanizeToken(fleetInfoUnit.ship.spec.shield_id) : t('common.none')}</dd></div>
                      <div><dt>{t('system.fuelCell')}</dt><dd>{humanizeToken(fleetInfoUnit.ship.spec.fuel_cell_id)}</dd></div>
                      <div><dt>{t('system.range')}</dt><dd>{fleetInfoUnit.ship.spec.fuel_range_parsecs} pc</dd></div>
                      <div><dt>{t('system.productionCost')}</dt><dd>{fleetInfoUnit.ship.spec.production_cost_pp} PP</dd></div>
                      <div><dt>{t('system.design')}</dt><dd>#{fleetInfoUnit.ship.source_design_id} · R{fleetInfoUnit.ship.source_design_revision}</dd></div>
                    </dl>
                    <section className="galaxy-fleet-info-weapons">
                      <strong>{t('system.weapons')}</strong>
                      {(fleetInfoUnit.ship.spec.weapons?.length ?? 0) > 0 ? (
                        <div>
                          {fleetInfoUnit.ship.spec.weapons?.map((weapon) => (
                            <span key={'fleet-info-weapon-' + weapon.slot}>{weapon.count}× {humanizeToken(weapon.weapon_id)}</span>
                          ))}
                        </div>
                      ) : <small>{t('system.noWeapons')}</small>}
                    </section>
                    <small className="galaxy-fleet-info-damage">{t('system.damageNotStrategic')}</small>
                  </div>
                ) : (
                  <div className="galaxy-fleet-info-content">
                    <dl className="galaxy-fleet-info-grid">
                      <div><dt>{t('system.fleetType')}</dt><dd>{humanizeToken(fleetInfoUnit.fleet.special_kind ?? fleetInfoUnit.fleet.role)}</dd></div>
                      {fleetInfoUnit.fleet.warp_drive_id && <div><dt>{t('system.warpDrive')}</dt><dd>{humanizeToken(fleetInfoUnit.fleet.warp_drive_id)}</dd></div>}
                      {fleetInfoUnit.fleet.ftl_speed !== undefined && <div><dt>{t('system.ftlSpeed')}</dt><dd>{fleetInfoUnit.fleet.ftl_speed}</dd></div>}
                      {fleetInfoUnit.fleet.fuel_cell_id && <div><dt>{t('system.fuelCell')}</dt><dd>{humanizeToken(fleetInfoUnit.fleet.fuel_cell_id)}</dd></div>}
                      {fleetInfoUnit.fleet.fuel_range_parsecs !== undefined && <div><dt>{t('system.range')}</dt><dd>{fleetInfoUnit.fleet.fuel_range_parsecs} pc</dd></div>}
                      <div><dt>{t('system.location')}</dt><dd>{pickerSystem?.name ?? t('galaxy.unknownStar')}</dd></div>
                    </dl>
                    <small className="galaxy-fleet-info-damage">{t('system.specialVesselLoadoutUnavailable')}</small>
                  </div>
                )}
              </section>
            )}
          </div>

          <div className="galaxy-fleet-picker-footer galaxy-fleet-picker-footer-compact">
            {fleetTargetFeedback && (
              <div className={'galaxy-fleet-target-feedback is-' + fleetTargetFeedback.outcome} role="status">
                <strong>
                  {fleetTargetFeedbackLabel} · {fleetTargetFeedback.outcome === 'planned'
                    ? t('galaxy.fleetTargetPlanned', { count: fleetTargetFeedback.orderCount ?? 1 })
                    : fleetTargetFeedback.outcome === 'blocked'
                      ? fleetTargetReasonLabel(t, fleetTargetFeedback.target?.reason)
                      : fleetTargetFeedback.outcome === 'cancelled'
                        ? t('galaxy.fleetTargetCancelled')
                        : t('galaxy.fleetTargetUnavailable')}
                </strong>
                {fleetTargetFeedback.target && (
                  <small>
                    {fleetTargetFeedback.target.distance_parsecs} pc · {t('galaxy.fleetTargetRange', { range: fleetTargetFeedback.target.fuel_range_parsecs })} · {t('galaxy.fleetTargetEta', { eta: fleetTargetFeedback.target.eta })}
                  </small>
                )}
              </div>
            )}
            <span className="galaxy-fleet-picker-selection-count">{pickerSelectedUnits.length}/{pickerUnits.length}</span>
            {pickerProfilesSupported ? (
              <span className="badge galaxy-fleet-picker-target-summary">{t('galaxy.fleetPickerTargetsCompact', { reachable: pickerReachableTargets, total: pickerTotalTargets })}</span>
            ) : (
              <small>{t('galaxy.fleetPickerSubsetProfile')}</small>
            )}
          </div>
        </aside>
      )}
      {selected && !selectedVisited && (
        <aside className="galaxy-unvisited-dialog" role="dialog" aria-label={t('galaxy.unvisitedTitle')}>
          <div>
            <small>{t('galaxy.unvisitedTitle')}</small>
            <strong>{t('galaxy.unknownStar')}</strong>
            <p>{t('galaxy.unvisitedBody')}</p>
          </div>
          <button type="button" className="ghost-button" onClick={onCloseSystem}>{t('common.close')}</button>
        </aside>
      )}
      {selected && selectedVisited && (
        <SystemDialog
          snapshot={snapshot}
          system={selected}
          onClose={onCloseSystem}
          onOpenColony={onOpenColony}
          onPlanOrder={onPlanOrder}
          onOpenFleetPicker={openFleetPicker}
          t={t}
        />
      )}
    </>
  )
}

function SystemDialog({ snapshot, system, onClose, onOpenColony, onPlanOrder, onOpenFleetPicker, t }: {
  snapshot: PlayerSnapshot
  system: StarSystem
  onClose: () => void
  onOpenColony: (colonyID: number) => void
  onPlanOrder: (order: DraftOrder) => void
  onOpenFleetPicker: (systemID: number, fleets: StrategicFleet[], clientX: number, clientY: number) => void
  t: Translator
}) {
  const bodies = system.bodies ?? (system.planets ?? []).map((planet) => ({
    id: planet.id,
    name: planet.name,
    orbit: planet.orbit,
    kind: 'planet' as const,
    planet_id: planet.id,
    outpost_id: undefined as number | undefined,
  }))
  const orderedBodies = [...bodies].sort((a, b) => a.orbit - b.orbit || a.id - b.id)
  const [selectedBodyID, setSelectedBodyID] = useState<number | null>(orderedBodies[0]?.id ?? null)

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
  }, [system.id])

  const decision = snapshot.decision
  if (!decision) return null

  const fleets = decision.strategic.fleets?.filter((fleet) => fleet.empire_id === decision.empire.id && fleet.at_system_id === system.id) ?? []
  const contacts = Array.from(new Map(
    (decision.strategic.contacts ?? [])
      .filter((contact) => contact.system_id === system.id)
      .map((contact) => [contact.empire_id + ':' + contact.kind, contact]),
  ).values())
  const fleetUnitCount = fleets.reduce((sum, fleet) => sum + (fleet.special_kind ? 1 : Math.max(1, fleet.ship_ids?.length ?? 0)), 0)


  const selectedBody = orderedBodies.find((body) => body.id === selectedBodyID) ?? orderedBodies[0]
  const selectedPlanet = selectedBody?.planet_id ? (system.planets ?? []).find((planet) => planet.id === selectedBody.planet_id) : undefined
  const selectedColony = selectedPlanet ? decision.colonies.find((colony) => colony.planet_id === selectedPlanet.id) : undefined
  const selectedPotential = selectedPlanet ? decision.strategic.planet_potentials?.find((item) => item.planet_id === selectedPlanet.id) : undefined
  const colonizeChoices = selectedPlanet
    ? (decision.decisions.colonization ?? []).filter((choice) => choice.system_id === system.id && choice.planet_id === selectedPlanet.id)
    : []
  const outpostChoices = selectedBody
    ? (decision.decisions.outpost_deployment ?? []).filter((choice) => choice.system_id === system.id && choice.body_id === selectedBody.id)
    : []

  const selectedStatus = selectedColony
    ? t('system.colonyStatus')
    : selectedBody?.outpost_id
      ? t('system.outpostStatus')
      : selectedBody?.kind === 'planet'
        ? t('system.uncolonized')
        : t('system.noSettlement')

  return (
    <div className="system-dialog-backdrop" role="presentation" onPointerDown={(event) => {
      if (event.target === event.currentTarget) onClose()
    }}>
      <section className="system-dialog system-dialog-classic" role="dialog" aria-modal="true" aria-labelledby={'system-dialog-title-' + system.id}>
        <header className="system-dialog-header system-dialog-classic-header">
          <div>
            <p className="eyebrow">{t('system.title', { system: system.name })}</p>
            <h2 id={'system-dialog-title-' + system.id}><GameIcon name="star-system" />{system.name}</h2>
          </div>
          <span className="badge system-star-class" title={starPalette(system.spectral_class).label}>{t('system.starClass', { class: starPalette(system.spectral_class).spectral })}</span>
        </header>

        <div className="system-dialog-scene">
          <div className="system-orbit-stage system-orbit-stage-classic" aria-label={t('system.bodies')}>
            <div className="system-orbit-canvas">
              <div className={'system-star-core system-star-class-' + system.spectral_class} aria-hidden="true"><StarArt spectralClass={system.spectral_class} seed={system.id} /></div>
              {orderedBodies.map((body, index) => {
                const radius = 17 + ((index + 1) / (orderedBodies.length + 1)) * 31
                const angle = (((system.id * 31) + (body.id * 67) + (index * 103)) % 360) * Math.PI / 180
                const x = 50 + Math.cos(angle) * radius
                const y = 50 + Math.sin(angle) * radius
                const planet = body.planet_id ? (system.planets ?? []).find((item) => item.id === body.planet_id) : undefined
                const colony = planet ? decision.colonies.find((item) => item.planet_id === planet.id) : undefined
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
                    <span className="system-orbit-ring" style={{ width: (radius * 2) + '%', height: (radius * 2) + '%' }} aria-hidden="true" />
                    <button
                      type="button"
                      className={bodyClasses}
                      style={{ left: x + '%', top: y + '%' }}
                      title={colony ? body.name + ' - ' + t('system.openColony') : body.name + ' - ' + bodyLabel(t, body.kind)}
                      aria-pressed={selected}
                      onClick={() => {
                        if (colony) {
                          onOpenColony(colony.id)
                          return
                        }
                        setSelectedBodyID(body.id)
                      }}
                    >
                      <span className="system-orbit-body-reticle" aria-hidden="true">
                        <span className="system-orbit-body-icon"><OrbitalBodyArt kind={body.kind} id={body.id} climateId={planet?.climate_id} /></span>
                        {colony && <span className="system-orbit-settlement system-orbit-settlement-colony"><GameIcon name="colonies" /></span>}
                        {!colony && body.outpost_id && <span className="system-orbit-settlement system-orbit-settlement-outpost"><GameIcon name="outpost" /></span>}
                      </span>
                      <strong>{body.name}</strong>
                    </button>
                  </div>
                )
              })}
            </div>
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
                {selectedPlanet && <div><dt>{t('system.climate')}</dt><dd>{planetTraitLabel(t, 'climate', selectedPlanet.climate_id)}</dd></div>}
                {selectedPlanet && <div><dt>{t('system.size')}</dt><dd>{planetTraitLabel(t, 'size', selectedPlanet.size_id)}</dd></div>}
                {selectedPlanet && <div><dt>{t('system.minerals')}</dt><dd>{planetTraitLabel(t, 'mineral', selectedPlanet.mineral_id)}</dd></div>}
                {selectedPlanet && <div><dt>{t('system.gravity')}</dt><dd>{planetTraitLabel(t, 'gravity', selectedPlanet.gravity_id)}</dd></div>}
                <div><dt>{t('system.status')}</dt><dd>{selectedStatus}</dd></div>
              </dl>
              {selectedPlanet && !selectedColony && selectedPotential && (
                <section className="system-planet-potential">
                  <p className="eyebrow">{t('system.potentialTitle')}</p>
                  <div className="system-potential-grid">
                    <div><span>{t('system.potentialFarmer')}</span><strong>{selectedPotential.food_per_farmer.toFixed(1)} F</strong></div>
                    <div><span>{t('system.potentialWorker')}</span><strong>{selectedPotential.production_per_worker.toFixed(1)} PP</strong></div>
                    <div><span>{t('system.potentialScientist')}</span><strong>{selectedPotential.research_per_scientist.toFixed(1)} RP</strong></div>
                    <div><span>{t('system.habitability')}</span><strong>{selectedPotential.climate_habitability_percent}%</strong></div>
                    <div><span>{t('system.gravityEffect')}</span><strong>{selectedPotential.gravity_penalty_percent > 0 ? `-${selectedPotential.gravity_penalty_percent}%` : '0%'}</strong></div>
                    <div><span>{t('system.populationPotential')}</span><strong>{selectedPotential.population_capacity.toFixed(0)}</strong></div>
                  </div>
                </section>
              )}
              <div className="system-body-status-row">
                {selectedColony && <span className="badge">{t('system.colonyStatus')}</span>}
                {selectedBody.outpost_id && <span className="badge">{t('system.outpostStatus')}</span>}
              </div>
            </aside>
          )}
        </div>

        <footer className="system-dialog-footer">
          <div className="system-dialog-footer-left">
            {contacts.length > 0 && (
              <div className="system-traffic-strip" aria-label={t('system.fleets')}>
                {contacts.map((contact, index) => {
                  const tone = strategicRelationTone(decision.empire.id, contact.empire_id, decision.diplomacy)
                  return (
                    <span className={'badge system-contact-badge relation-' + tone} key={'contact-' + index + '-' + contact.empire_id}>
                      <GameIcon name={strategicContactIcon(contact.kind)} />
                      {t('common.empireFallback', { id: contact.empire_id })} Â· {contact.kind}
                    </span>
                  )
                })}
              </div>
            )}
          </div>
          <div className="system-dialog-footer-actions">
            {fleets.length > 0 && (
              <button type="button" className="button-secondary system-fleets-button" onClick={(event) => onOpenFleetPicker(system.id, fleets, event.clientX, event.clientY)}>
                <GameIcon name="fleets" />{t('system.fleetsShips')} <span className="badge">{fleetUnitCount}</span>
              </button>
            )}
            {selectedBody && colonizeChoices.map((choice) => (
              <button
                type="button"
                className="button-primary"
                key={'colonize-' + choice.fleet_id}
                onClick={() => {
                  if (!window.confirm(t('system.colonizeConfirm', { body: selectedBody.name, fleet: choice.fleet_id }))) return
                  onPlanOrder({ key: 'fleet:' + choice.fleet_id, kind: 'empire.colonize_planet', payload: { fleet_id: choice.fleet_id, planet_id: choice.planet_id } })
                }}
              >
                <GameIcon name="flag" />{t('system.colonize')}
              </button>
            ))}
            {selectedBody && outpostChoices.map((choice) => (
              <button
                type="button"
                className="button-secondary"
                key={'outpost-' + choice.fleet_id}
                onClick={() => {
                  if (!window.confirm(t('system.outpostConfirm', { body: selectedBody.name, fleet: choice.fleet_id }))) return
                  onPlanOrder({ key: 'fleet:' + choice.fleet_id, kind: 'fleet.deploy_outpost', payload: { fleet_id: choice.fleet_id, body_id: choice.body_id } })
                }}
              >
                <GameIcon name="outpost" />{t('system.buildOutpost')}
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
                    <td><button type="button" className="button-secondary button-compact" onClick={() => onOpenColony(colony.id)}><GameIcon name="open" />{t('colonies.open')}</button></td>
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

function PopulationMoveControls({ colony, onPlan, t, compact = false, outputs }: {
  colony: Colony
  onPlan: (colonyID: number, farmers: number, workers: number, scientists: number) => void
  t: Translator
  compact?: boolean
  outputs?: Partial<Record<PopulationJob, { total: number; unit: string }>>
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
                <span className="population-job-label"><span className="population-job-name"><GameIcon name={populationJobIcons[job]} /><span>{labels[job]}</span></span><strong>{values[job].toFixed(1)}</strong></span>
                {outputs?.[job] && (
                  <span className="population-job-output">
                    <strong>{outputs[job]?.total.toFixed(1)} {outputs[job]?.unit}</strong>
                  </span>
                )}
              </header>
              <div className="population-people" aria-label={labels[job] + ': ' + values[job].toFixed(1)}>
                {people.length === 0 ? <span className="population-empty" aria-hidden="true"><GameIcon name={populationJobIcons[job]} /></span> : people.map((person, index) => {
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
                      <span className="population-person-glyph" aria-hidden="true"><GameIcon name={populationJobIcons[job]} /></span>
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
  void onRemoveOrder
  const planetInfoRef = useRef<HTMLDetailsElement | null>(null)
  useEffect(() => {
    const closePlanetInfoOnOutsideClick = (event: MouseEvent) => {
      const details = planetInfoRef.current
      if (!details?.open) return
      if (event.target instanceof Node && !details.contains(event.target)) details.open = false
    }
    document.addEventListener('click', closePlanetInfoOnOutsideClick)
    return () => document.removeEventListener('click', closePlanetInfoOnOutsideClick)
  }, [])
  const displayColony = preview?.colony ?? colony
  const decision = snapshot.decision
  const galaxy = decision?.strategic.galaxy
  const planetContext = galaxy?.systems
    .flatMap((system) => (system.planets ?? []).map((item) => ({ system, planet: item })))
    .find((item) => item.planet.id === colony.planet_id)
  const planetPotential = planetContext
    ? decision?.strategic.planet_potentials?.find((item) => item.planet_id === planetContext.planet.id)
    : undefined
  const population = aggregatePopulation(displayColony)
  const growth = preview?.population_growth_per_turn ?? Math.max(0, displayColony.population_dynamics.projected_growth)
  const loss = preview?.population_loss_per_turn ?? Math.max(0, displayColony.population_dynamics.projected_starvation)
  const signedGrowth = loss > 0 ? -loss : growth
  const growthTone = metricTone(signedGrowth)
  const nextPopulationETA = preview?.next_population_eta_turns
  const metricBreakdowns = preview?.breakdowns
  const growthEta = signedGrowth > 0 && nextPopulationETA && nextPopulationETA > 0 ? ` (${formatEta(t, nextPopulationETA)})` : ''
  const transferChoices = decision?.decisions.population_transfers?.filter((item) => item.source_colony_id === colony.id) ?? []
  const jobOutputs: Partial<Record<PopulationJob, { total: number; unit: string }>> = {
    farmer: {
      total: displayColony.adjusted_economy.food,
      unit: 'F',
    },
    worker: {
      total: displayColony.adjusted_economy.production,
      unit: 'PP',
    },
    scientist: {
      total: displayColony.adjusted_economy.research,
      unit: 'RP',
    },
  }

  return (
    <>
      <header className="colony-detail-toolbar">
        <div className="colony-detail-toolbar-title">
          <strong>{t('colonies.colony', { id: colony.id })}</strong>
          <span>{planetContext ? `${planetContext.system.name} Â· ${planetContext.planet.name}` : t('colonies.planet', { id: colony.planet_id })}</span>
        </div>
        <button type="button" className="button-ghost colony-detail-back" onClick={onBack}>{t('common.back')}</button>
      </header>

      <div className="colony-command-layout">
        <Card className="colony-profile-card colony-information-card">
          <section className="colony-information-heading">
            <p className="eyebrow">{t('colony.information')}</p>
            {planetContext && (
              <details className="colony-planet-info" ref={planetInfoRef}>
                <summary aria-label={t('colony.planetInfo')} title={t('colony.planetInfo')}><GameIcon name="info" /></summary>
                <div className="colony-planet-info-popover">
                  <button type="button" className="colony-planet-info-close" aria-label={t('common.close')} title={t('common.close')} onClick={() => { if (planetInfoRef.current) planetInfoRef.current.open = false }}><GameIcon name="close" /></button>
                  <div className="colony-planet-info-item">
                    <span>{t('system.climate')}</span>
                    <strong>{planetTraitLabel(t, 'climate', planetContext.planet.climate_id)}</strong>
                    {planetPotential && <small>{t('colony.climateEffect', { food: planetPotential.food_per_farmer.toFixed(1), habitability: planetPotential.climate_habitability_percent })}</small>}
                  </div>
                  <div className="colony-planet-info-item">
                    <span>{t('system.size')}</span>
                    <strong>{planetTraitLabel(t, 'size', planetContext.planet.size_id)}</strong>
                    {planetPotential && <small>{t('colony.sizeEffect', { capacity: planetPotential.size_base_capacity.toFixed(0) })}</small>}
                  </div>
                  <div className="colony-planet-info-item">
                    <span>{t('system.minerals')}</span>
                    <strong>{planetTraitLabel(t, 'mineral', planetContext.planet.mineral_id)}</strong>
                    {planetPotential && <small>{t('colony.mineralEffect', { production: planetPotential.production_per_worker.toFixed(1) })}</small>}
                  </div>
                  <div className="colony-planet-info-item">
                    <span>{t('system.gravity')}</span>
                    <strong>{planetTraitLabel(t, 'gravity', planetContext.planet.gravity_id)}</strong>
                    {planetPotential && <small>{planetPotential.gravity_penalty_percent > 0 ? t('colony.gravityEffect', { penalty: planetPotential.gravity_penalty_percent }) : t('colony.gravityNoPenalty')}</small>}
                  </div>
                </div>
              </details>
            )}
          </section>

          <dl className="colony-profile-stats colony-information-stats">
            <div><dt>{t('colony.populationStatus')}</dt><dd>{population.total.toFixed(2)} / {displayColony.population_dynamics.capacity.toFixed(2)}</dd></div>
            <div>
              <dt>{t('colonies.growth')}</dt>
              <dd>
                {metricBreakdowns ? (
                  <details className="colony-metric-value-info" name={`colony-metric-${colony.id}`}>
                    <summary className={`resource-text-${growthTone}`} aria-label={t('colonies.growth')}>{signedGrowth > 0 ? '+' : ''}{signedGrowth.toFixed(2)}{growthEta}</summary>
                    <div className="colony-metric-value-popover">
                      <MetricBreakdown title={t('colonies.growth')} breakdown={metricBreakdowns.growth} unit={t('colony.popUnit')} digits={2} t={t} />
                    </div>
                  </details>
                ) : <span className={`resource-text-${growthTone}`}>{signedGrowth > 0 ? '+' : ''}{signedGrowth.toFixed(2)}{growthEta}</span>}
              </dd>
            </div>
            <div><dt>{t('colony.groundForces')}</dt><dd>{displayColony.ground_forces?.infantry ?? 0}</dd></div>
            <div>
              <dt>{t('colony.taxContribution')}</dt>
              <dd>
                {metricBreakdowns ? (
                  <details className="colony-metric-value-info" name={`colony-metric-${colony.id}`}>
                    <summary aria-label={t('colony.taxContribution')}>{displayColony.adjusted_economy.tax_bc.toFixed(1)} BC</summary>
                    <div className="colony-metric-value-popover">
                      <MetricBreakdown title={t('colony.taxContribution')} breakdown={metricBreakdowns.tax_bc} unit="BC" digits={1} t={t} />
                    </div>
                  </details>
                ) : <span>{displayColony.adjusted_economy.tax_bc.toFixed(1)} BC</span>}
              </dd>
            </div>
          </dl>
        </Card>
        <Card className="colony-jobs-card">
          <div className="card-heading colony-panel-heading colony-panel-heading-compact">
            <p className="eyebrow">{t('colony.jobOutputs')}</p>
          </div>
          <PopulationMoveControls colony={displayColony} onPlan={onPlanPopulation} outputs={jobOutputs} t={t} />
        </Card>

        <ConstructionSummary
          colony={displayColony}
          preview={preview}
          draftOrders={draftOrders}
          onOpen={() => onOpenConstruction(colony.id)}
          t={t}
        />
      </div>

      <Card className="colony-surface-card">
        <div className="card-heading colony-panel-heading">
          <div><p className="eyebrow">{t('colony.surface')}</p><h2>{t('colony.buildings')}</h2></div>
          <span className="badge">{(displayColony.buildings ?? []).length}</span>
        </div>
        <p className="muted colony-surface-hint">{t('colony.surfaceHint')}</p>
        {(displayColony.buildings ?? []).length === 0 ? <p className="muted">{t('colony.noBuildings')}</p> : (
          <div className="building-grid colony-building-grid">{(displayColony.buildings ?? []).map((building) => <div className="building-tile building-tile-rich" key={building}><BuildingArt buildingId={building} variant="compact" className="building-tile-art" /><strong>{humanizeToken(building)}</strong></div>)}</div>
        )}
      </Card>

      <Card className="colony-transfer-card">
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
  const currentProgressPP = projected?.project.progress_pp ?? colony.construction?.progress_pp ?? 0
  const currentCostPP = projected?.cost_pp
  const currentProgressPercent = currentCostPP && currentCostPP > 0 ? Math.max(0, Math.min(100, (currentProgressPP / currentCostPP) * 100)) : 0
  const futureQueueCount = Math.max(0, items.length - (current ? 1 : 0))
  return (
    <Card className="construction-summary-card">
      <div className="card-heading construction-summary-heading">
        <div>
          <p className="eyebrow">{t('colony.construction')}</p>
          <h2>{current ? humanizeToken(current.project_id) : t('construction.empty')}</h2>
        </div>
        <span className="badge">{t('construction.queueCount', { count: futureQueueCount })}</span>
      </div>
      {current && currentCostPP !== undefined && currentCostPP > 0 && (
        <div className="construction-summary-progress">
          <div className="construction-progress-track" aria-label={t('construction.progress')}>
            <span style={{ width: `${currentProgressPercent}%` }} />
          </div>
          <div className="construction-summary-progress-meta">
            <span>{currentProgressPP.toFixed(1)} / {currentCostPP.toFixed(0)} PP Â· {currentProgressPercent.toFixed(0)}%</span>
            <strong>{projected ? formatEta(t, projected.eta_turns) : t('common.noEta')}</strong>
          </div>
        </div>
      )}
      <div className="construction-summary-meta">
        {(!current || currentCostPP === undefined || currentCostPP <= 0) && <span>{projected ? formatEta(t, projected.eta_turns) : t('common.noEta')}</span>}
        {items.length > 1 && <small>{items.slice(1, 4).map((item) => humanizeToken(item.project_id)).join(' Â· ')}{items.length > 4 ? ' â€¦' : ''}</small>}
      </div>
      <button type="button" className="button-secondary button-wide" onClick={onOpen}><GameIcon name="build" />{t('construction.openManager')}</button>
    </Card>
  )
}

export function StrategicConstructionView({ snapshot, preview, draftOrders, colonyID, onBack, onOpenShipDesigner, onPlanOrder, onRemoveOrder, t }: {
  snapshot: PlayerSnapshot
  preview: PlanningPreviewSnapshot | null
  draftOrders: DraftOrder[]
  colonyID: number
  onBack: () => void
  onOpenShipDesigner: (designID?: number) => void
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
      <header className="colony-detail-toolbar construction-detail-toolbar">
        <div className="colony-detail-toolbar-title">
          <strong>{t('construction.manageTitle', { colony: colony.id })}</strong>
        </div>
        <button type="button" className="button-ghost colony-detail-back" onClick={onBack}>{t('common.back')}</button>
      </header>
      <ConstructionEditor
        colony={displayColony}
        preview={projected}
        choices={constructionDecision?.choices ?? []}
        shipDesigns={snapshot.decision?.strategic.ship_designs ?? []}
        draftOrders={draftOrders}
        onOpenShipDesigner={onOpenShipDesigner}
        onPlanOrder={onPlanOrder}
        onRemoveOrder={onRemoveOrder}
        t={t}
      />
    </>
  )
}

function ConstructionEditor({ colony, preview, choices, shipDesigns, draftOrders, onOpenShipDesigner, onPlanOrder, onRemoveOrder, t }: {
  colony: Colony
  preview?: PlanningPreviewSnapshot['preview']['projection']['colonies'][number]
  choices: ConstructionChoice[]
  shipDesigns: ShipDesign[]
  draftOrders: DraftOrder[]
  onOpenShipDesigner: (designID?: number) => void
  onPlanOrder: (order: DraftOrder) => void
  onRemoveOrder: (key: string) => void
  t: Translator
}) {
  const draftKey = `construction:${colony.id}`
  const drafted = draftOrders.find((order) => order.key === draftKey)
  const draftedItems = drafted?.payload.items
  const items = Array.isArray(draftedItems) ? draftedItems as DraftQueueItem[] : queueItemsFromColony(colony)
  const [selectedChoiceIndex, setSelectedChoiceIndex] = useState(() => {
    const current = items[0]
    if (!current) return 0
    const index = choices.findIndex((choice) => current.project_kind === choice.project_kind
      && current.project_id === choice.project_id
      && (current.ship_design_id ?? 0) === (choice.ship_design_id ?? 0)
      && (current.ship_design_revision ?? 0) === (choice.ship_design_revision ?? 0))
    return index >= 0 ? index : 0
  })
  const [abortConfirmOpen, setAbortConfirmOpen] = useState(false)

  useEffect(() => {
    if (choices.length === 0 && selectedChoiceIndex !== 0) setSelectedChoiceIndex(0)
    if (choices.length > 0 && selectedChoiceIndex >= choices.length) setSelectedChoiceIndex(choices.length - 1)
  }, [choices.length, selectedChoiceIndex])

  useEffect(() => {
    if (!abortConfirmOpen) return
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setAbortConfirmOpen(false)
    }
    window.addEventListener('keydown', closeOnEscape)
    return () => window.removeEventListener('keydown', closeOnEscape)
  }, [abortConfirmOpen])

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

  function matchesChoice(item: DraftQueueItem, choice: ConstructionChoice) {
    return item.project_kind === choice.project_kind
      && item.project_id === choice.project_id
      && (item.ship_design_id ?? 0) === (choice.ship_design_id ?? 0)
      && (item.ship_design_revision ?? 0) === (choice.ship_design_revision ?? 0)
  }

  function displayChoice(choice: ConstructionChoice) {
    return humanizeToken(choice.ship_design_name ?? choice.project_id)
  }

  function displayItem(item: DraftQueueItem) {
    const choice = choices.find((candidate) => matchesChoice(item, candidate))
    return choice ? displayChoice(choice) : humanizeToken(item.project_id)
  }

  function queueItemArt(item: DraftQueueItem, className: string) {
    const design = constructionShipDesign(item, shipDesigns)
    if (design) {
      return (
        <ProceduralShipGlyph
          className={className}
          seed={`${design.empire_id}:${design.id}:${design.revision}:${design.spec.strategic_picture_id}`}
          genome={decodeShipVisualGenome(design.visual_genome)}
          hullId={design.spec.hull_id}
          weaponCount={design.spec.weapons?.reduce((sum, mount) => sum + mount.count, 0) ?? 0}
        />
      )
    }
    return <GameIcon name={constructionProjectIcon(item.project_kind, item.project_id)} />
  }

  function selectQueueItem(item: DraftQueueItem) {
    const index = choices.findIndex((choice) => matchesChoice(item, choice))
    if (index >= 0) setSelectedChoiceIndex(index)
  }

  function abortCurrentBuild() {
    if (items.length === 0) return
    save(items.slice(1))
    setAbortConfirmOpen(false)
  }

  const selectedChoice = choices[selectedChoiceIndex]
  const selectedQueueIndex = selectedChoice ? items.findIndex((item) => matchesChoice(item, selectedChoice)) : -1
  const selectedProjected = selectedQueueIndex >= 0 ? preview?.construction?.[selectedQueueIndex] : undefined
  const currentItem = items[0]
  const queuedItems = items.slice(1)
  const currentProjected = preview?.construction?.[0]
  const currentProgressPP = currentProjected?.project.progress_pp ?? colony.construction?.progress_pp ?? 0
  const currentCostPP = currentProjected?.cost_pp
  const currentProgressPercent = currentCostPP && currentCostPP > 0 ? Math.max(0, Math.min(100, (currentProgressPP / currentCostPP) * 100)) : 0

  return (
    <div className="construction-workspace construction-workspace-classic">
      <Card className="construction-catalog-panel">
        <div className="card-heading construction-panel-heading">
          <div><p className="eyebrow">{t('construction.catalog')}</p><h2>{t('construction.available')}</h2></div>
          <span className="badge">{choices.length}</span>
        </div>
        {choices.length === 0 ? <p className="muted">{t('construction.noAvailable')}</p> : (
          <div className="construction-catalog-list">
            {choices.map((choice, index) => {
              const nonRepeatable = choice.project_kind === 'building' || choice.project_kind === 'planetary_transformation'
              const alreadyQueued = nonRepeatable && items.some((item) => matchesChoice(item, choice))
              const selected = index === selectedChoiceIndex
              return (
                <button
                  type="button"
                  className={'construction-catalog-item' + (selected ? ' selected' : '') + (constructionProjectUsesLargeArt(choice.project_kind) ? ' construction-catalog-item-rich' : '')}
                  key={`${choice.project_kind}-${choice.project_id}-${choice.ship_design_id ?? 0}-${choice.ship_design_revision ?? 0}-${index}`}
                  aria-pressed={selected}
                  onClick={() => setSelectedChoiceIndex(index)}
                >
                  <span className={constructionProjectUsesLargeArt(choice.project_kind) ? 'construction-catalog-glyph construction-catalog-glyph-rich' : 'construction-catalog-glyph'} data-kind={choice.project_kind} aria-hidden="true"><ConstructionChoiceArt choice={choice} shipDesigns={shipDesigns} variant="compact" /></span>
                  <span className="construction-catalog-copy">
                    <strong>{displayChoice(choice)}</strong>
                    <small>{humanizeToken(choice.project_kind)} Â· {t('construction.cost', { pp: choice.production_cost_pp.toFixed(0) })}</small>
                  </span>
                  {alreadyQueued && <span className="badge">{t('construction.queued')}</span>}
                </button>
              )
            })}
          </div>
        )}
      </Card>

      <Card className="construction-project-panel">
        {selectedChoice ? (
          <>
            <div className={constructionProjectUsesLargeArt(selectedChoice.project_kind) ? 'construction-project-hero construction-project-hero-rich' : 'construction-project-hero'}>
              <div className={constructionProjectUsesLargeArt(selectedChoice.project_kind) ? 'construction-project-visual construction-project-visual-rich' : 'construction-project-visual'} data-kind={selectedChoice.project_kind} aria-hidden="true">
                <ConstructionChoiceArt choice={selectedChoice} shipDesigns={shipDesigns} variant="hero" />
              </div>
              <div className="construction-project-title">
                <p className="eyebrow">{t('construction.selectedProject')}</p>
                <h2>{displayChoice(selectedChoice)}</h2>
                <span className="badge">{humanizeToken(selectedChoice.project_kind)}</span>
              </div>
            </div>

            <dl className="construction-project-facts">
              <div><dt>{t('construction.productionCost')}</dt><dd>{selectedChoice.production_cost_pp.toFixed(0)} PP</dd></div>
              {selectedChoice.maintenance_bc !== undefined && <div><dt>{t('construction.maintenance')}</dt><dd>{selectedChoice.maintenance_bc.toFixed(1)} BC</dd></div>}
              {selectedChoice.freighters_added !== undefined && <div><dt>{t('construction.freightersAdded')}</dt><dd>+{selectedChoice.freighters_added}</dd></div>}
              {selectedChoice.ship_design_name && <div><dt>{t('construction.shipDesign')}</dt><dd>{humanizeToken(selectedChoice.ship_design_name)}</dd></div>}
              {selectedChoice.ship_design_revision !== undefined && <div><dt>{t('construction.revision')}</dt><dd>r{selectedChoice.ship_design_revision}</dd></div>}
              {selectedProjected && <div><dt>{t('construction.remaining')}</dt><dd>{selectedProjected.remaining_pp.toFixed(1)} PP</dd></div>}
              {selectedProjected && <div><dt>{t('construction.eta')}</dt><dd>{formatEta(t, selectedProjected.eta_turns)}</dd></div>}
            </dl>

            <div className="construction-project-description">
              {selectedChoice.project_kind === 'building' && (
                <div className="construction-building-info">
                  <div className="construction-building-effects">
                    <h3>{t('construction.runtimeEffects')}</h3>
                    {(selectedChoice.effects?.length ?? 0) > 0 ? (
                      <ul>
                        {selectedChoice.effects?.map((effect, index) => <li key={`${effect.kind}:${index}`}>{constructionEffectText(t, effect)}</li>)}
                      </ul>
                    ) : (
                      <p className="muted">{t('construction.noRuntimeEffect')}</p>
                    )}
                  </div>
                  {selectedChoice.original_description && (
                    <div className="construction-building-original">
                      <h3>{t('construction.originalDescription')}</h3>
                      <p>{selectedChoice.original_description}</p>
                    </div>
                  )}
                </div>
              )}
              <p className="construction-queue-hint">{selectedQueueIndex >= 0 ? t('construction.projectQueuedHint') : t('construction.projectAvailableHint')}</p>
            </div>

            <div className="construction-project-actions">
              {(() => {
                const nonRepeatable = selectedChoice.project_kind === 'building' || selectedChoice.project_kind === 'planetary_transformation'
                const alreadyQueued = nonRepeatable && items.some((item) => matchesChoice(item, selectedChoice))
                return (
                  <button
                    type="button"
                    className="button-primary"
                    disabled={alreadyQueued}
                    onClick={() => save([...items, queueItemFromChoice(selectedChoice)])}
                  >
                    <GameIcon name={constructionProjectIcon(selectedChoice.project_kind, selectedChoice.project_id)} />{alreadyQueued ? t('construction.alreadyQueued') : t('construction.addToQueue')}
                  </button>
                )
              })()}
              {(selectedChoice.project_kind === 'military_ship' || selectedChoice.ship_design_id !== undefined) && (
                <button type="button" className="button-secondary construction-designer-placeholder" onClick={() => onOpenShipDesigner(selectedChoice.ship_design_id)} title={t('construction.shipDesignerHint')}>
                  <GameIcon name="ship-designer" />{t('construction.openShipDesigner')}
                </button>
              )}
            </div>

            {(selectedChoice.project_kind === 'military_ship' || selectedChoice.ship_design_id !== undefined) && (
              <p className="muted construction-designer-note">{t('construction.shipDesignerHint')}</p>
            )}
          </>
        ) : (
          <div className="empty-state construction-project-empty">
            <h2>{t('construction.noSelection')}</h2>
            <p>{t('construction.noSelectionHint')}</p>
          </div>
        )}
      </Card>

      <Card className="construction-queue-panel">

        {currentItem ? (
          <section className="construction-current-build">
            <div className="construction-current-head">
              <div className="construction-current-title">
                <span className="construction-current-kind-icon">
                  {queueItemArt(currentItem, 'construction-current-ship-glyph')}
                </span>
                <div><p className="eyebrow">{t('construction.current')}</p><strong>{displayItem(currentItem)}</strong></div>
              </div>
              <button type="button" className="button-danger construction-abort-button" onClick={() => setAbortConfirmOpen(true)}>{t('construction.abort')}</button>
            </div>
            <div className="construction-progress-track" aria-label={t('construction.progress')}>
              <span style={{ width: `${currentProgressPercent}%` }} />
            </div>
            <div className="construction-current-meta">
              <span>{currentCostPP ? `${currentProgressPP.toFixed(1)} / ${currentCostPP.toFixed(0)} PP Â· ${currentProgressPercent.toFixed(0)}%` : `${currentProgressPP.toFixed(1)} PP`}</span>
              <span>{currentProjected ? formatEta(t, currentProjected.eta_turns) : t('common.noEta')}</span>
            </div>
          </section>
        ) : <p className="muted construction-current-empty">{t('construction.empty')}</p>}
        <div className="card-heading construction-panel-heading construction-queue-heading-compact">
          <h2>{t('construction.queue')}</h2>
          <span className="badge">{queuedItems.length}</span>
        </div>

        {queuedItems.length > 0 && (
          <div className="construction-queue-list construction-queue-list-classic">
            {queuedItems.map((item, queueIndex) => {
              const index = queueIndex + 1
              const projectedItem = preview?.construction?.[index]
              return (
                <div className="queue-row construction-queue-row" key={`${item.project_kind}-${item.project_id}-${item.ship_design_id ?? 0}-${item.ship_design_revision ?? 0}-${index}`}>
                  <span className="queue-position">{queueIndex + 1}</span>
                  <button type="button" className="construction-queue-copy" onClick={() => selectQueueItem(item)}>
                    <span className="construction-queue-art">
                      {queueItemArt(item, 'construction-queue-ship-glyph')}
                    </span>
                    <span className="construction-queue-copy-text">
                      <strong>{displayItem(item)}</strong>
                      <small>{humanizeToken(item.project_kind)} Â· {projectedItem ? formatEta(t, projectedItem.eta_turns) : t('common.noEta')}</small>
                    </span>
                  </button>
                  <div className="action-row compact-actions construction-queue-actions">
                    <button type="button" className="button-ghost" disabled={queueIndex === 0} onClick={() => move(index, -1)} aria-label={t('construction.moveUp')}><GameIcon name="arrow-up" /></button>
                    <button type="button" className="button-ghost" disabled={queueIndex === queuedItems.length - 1} onClick={() => move(index, 1)} aria-label={t('construction.moveDown')}><GameIcon name="arrow-down" /></button>
                    <button type="button" className="button-ghost" onClick={() => save(items.filter((_, itemIndex) => itemIndex !== index))}>{t('common.remove')}</button>
                  </div>
                </div>
              )
            })}
          </div>
        )}

        <div className="construction-queue-footer">
          {drafted && <button type="button" className="button-secondary construction-cancel-draft" onClick={() => onRemoveOrder(draftKey)}>{t('construction.resetDraft')}</button>}
          <span className="construction-reserve">{t('construction.reserve', { pp: (colony.construction_reserve_pp ?? 0).toFixed(1) })}</span>
        </div>
      </Card>

      {abortConfirmOpen && currentItem && (
        <div className="construction-confirm-backdrop" role="presentation" onPointerDown={(event) => {
          if (event.target === event.currentTarget) setAbortConfirmOpen(false)
        }}>
          <section className="construction-confirm-dialog" role="alertdialog" aria-modal="true" aria-labelledby={`construction-abort-title-${colony.id}`} aria-describedby={`construction-abort-body-${colony.id}`}>
            <p className="eyebrow">{t('construction.abort')}</p>
            <h2 id={`construction-abort-title-${colony.id}`}>{displayItem(currentItem)}</h2>
            <p id={`construction-abort-body-${colony.id}`}>{t('construction.abortConfirmBody', { pp: currentProgressPP.toFixed(1) })}</p>
            <div className="action-row construction-confirm-actions">
              <button type="button" className="button-secondary" onClick={() => setAbortConfirmOpen(false)}>{t('common.cancel')}</button>
              <button type="button" className="button-danger" onClick={abortCurrentBuild}>{t('construction.abortConfirm')}</button>
            </div>
          </section>
        </div>
      )}
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
      `${choice.source_colony_id} â†’ ${choice.destination_colony_id}`,
      `${choice.source_job} â†’ ${choice.destination_job}`,
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
        <strong>{choice.source_colony_id} â†’ {choice.destination_colony_id}</strong>
        <small>{choice.source_job} â†’ {choice.destination_job} Â· {t('transfer.freighters')}: {choice.freighters_required} Â· {t('transfer.eta')}: {choice.eta}{choice.same_system ? ` Â· ${t('transfer.sameSystem')}` : ''}</small>
      </span>
      <button type="button" className="button-secondary" onClick={confirmTransfer}><GameIcon name="check" />{t('transfer.confirm')}</button>
    </div>
  )
}

export function StrategicFleetsView({ snapshot, onPlanOrder, t }: { snapshot: PlayerSnapshot; onPlanOrder: (order: DraftOrder) => void; t: Translator }) {
  const decision = snapshot.decision
  const fleets = decision?.strategic.fleets ?? []
  const shipsByID = new Map((decision?.strategic.ships ?? []).map((ship) => [ship.id, ship]))
  return (
    <>
      <PageHeader eyebrow={t('fleets.eyebrow')} title={t('fleets.title')} subtitle={t('fleets.subtitle')} />
      <div className="content-grid content-grid-2">
        {fleets.length === 0 ? <EmptyState title={t('fleets.noProjection')} /> : fleets.map((fleet) => {
          const moves = decision?.decisions.fleet_moves?.filter((choice) => choice.fleet_id === fleet.id) ?? []
          const fleetShips = fleet.ship_ids?.map((shipID) => shipsByID.get(shipID)).filter((ship): ship is NonNullable<typeof ship> => Boolean(ship)) ?? []
          const leadShip = fleetShips[0]
          return (
            <Card key={fleet.id}>
              <div className="card-heading">
                <div className="fleet-card-title">
                  <ProceduralShipGlyph
                    className="fleet-card-ship-visual"
                    seed={leadShip
                      ? `${leadShip.empire_id}:${leadShip.source_design_id}:${leadShip.source_design_revision}:${leadShip.spec.strategic_picture_id}`
                      : `fleet:${fleet.empire_id}:${fleet.id}`}
                    genome={decodeShipVisualGenome(leadShip?.visual_genome)}
                    hullId={leadShip?.spec.hull_id}
                    weaponCount={leadShip?.spec.weapons?.reduce((sum, mount) => sum + mount.count, 0) ?? 0}
                  />
                  <div><p className="eyebrow">{t('galaxy.fleet', { id: fleet.id })}</p><h2 className="fleet-role-heading"><span className="fleet-role-chip"><GameIcon name={fleetRoleIcon(fleet)} /></span>{fleet.role}</h2></div>
                </div>
                <span className="badge fleet-role-badge"><GameIcon name={fleetRoleIcon(fleet)} />{t('fleets.ships')}: {fleet.ship_ids?.length ?? 0}</span>
              </div>
              <dl className="detail-list compact">
                <div><dt>{t('fleets.atSystem')}</dt><dd>{fleet.at_system_id ?? 'â€”'}</dd></div>
                <div><dt>{t('fleets.destination')}</dt><dd>{fleet.destination_system_id ?? 'â€”'}</dd></div>
                <div><dt>{t('fleets.eta')}</dt><dd>{fleet.remaining_turns ?? 'â€”'}</dd></div>
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
                      <strong>{t('fleets.move')} â†’ {choice.destination_system_id}</strong>
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
  const [inspectedTechnology, setInspectedTechnology] = useState<{
    choice: ResearchChoice
    info: ResearchTechnologyInfo
    name: string
    categoryName: string
  } | null>(null)
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
      if (event.key !== 'Escape') return
      if (inspectedTechnology) setInspectedTechnology(null)
      else onClose()
    }
    document.body.classList.add('modal-open')
    window.addEventListener('keydown', onKey)
    return () => {
      document.body.classList.remove('modal-open')
      window.removeEventListener('keydown', onKey)
    }
  }, [inspectedTechnology, onClose])

  function inspectTechnology(choice: ResearchChoice, technologyID: number, index: number) {
    const technologyKey = choice.technology_keys[index] ?? String(technologyID)
    const technologyNameKey = choice.technology_name_keys[index] ?? `technology.${technologyKey}.name`
    const name = serverLabel(t, technologyNameKey, humanizeToken(technologyKey))
    const info = choice.technology_info?.find((candidate) => candidate.technology_id === technologyID) ?? {
      technology_id: technologyID,
      technology_key: technologyKey,
      technology_name_key: technologyNameKey,
      effects: [],
    }
    setInspectedTechnology({
      choice,
      info,
      name,
      categoryName: serverLabel(t, choice.category_name_key, choice.category_id),
    })
  }

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
            <span>{t('research.rpRate')}</span><strong>{active?.rp_per_turn.toFixed(1) ?? 'â€”'} RP</strong>
            <span>{t('research.eta')}</span><strong>{formatEta(t, active?.eta_turns)}</strong>
          </div>
          <button type="button" className="button-ghost research-overlay-close" onClick={onClose} aria-label={t('common.close')}><GameIcon name="close" /></button>
        </header>

        <div className="research-grid-classic">
          {categories.map((category) => {
            const categoryIcon = researchCategoryIcon(category.order)
            const choice = choices.find((item) => item.category_id === category.id)
            if (!choice) {
              return <section className="research-field-panel research-field-disabled" key={category.id}><div className="research-field-bar"><span className="research-field-title"><GameIcon name={categoryIcon} /><strong>{serverLabel(t, category.name_key, category.id)}</strong></span></div><p>{t('common.none')}</p></section>
            }
            const isActive = activeFieldID === choice.tech_field_id
            const fieldName = t('research.fieldNumber', { id: choice.tech_field_id })
            return (
              <section className={`research-field-panel${isActive ? ' active' : ''}`} key={category.id}>
                <div className="research-field-bar">
                  <span className="research-field-title"><GameIcon name={categoryIcon} /><strong>{serverLabel(t, category.name_key, category.id)}</strong></span>
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
                        <div className="research-tech-choice-row" key={technologyID}>
                          <button
                            type="button"
                            className={`research-tech-choice${selected ? ' selected' : ''}`}
                            aria-pressed={selected}
                            onClick={() => selectResearch(choice, technologyID)}
                          >
                            {selected
                              ? <strong className="research-tech-name selected-name">{technologyName}</strong>
                              : <span className="research-tech-name">{technologyName}</span>}
                            {selected && <strong className="research-current-badge">{'\u2713'} {t('research.currentChoice')}</strong>}
                          </button>
                          <button
                            type="button"
                            className="research-tech-info-button"
                            aria-label={t('research.info.open', { technology: technologyName })}
                            title={t('research.info.open', { technology: technologyName })}
                            onClick={() => inspectTechnology(choice, technologyID, index)}
                          >?</button>
                        </div>
                      )
                    }) : (
                      <>
                        {choice.technology_ids.map((technologyID, index) => {
                          const selected = isActive && activeTechnologyIDs.has(technologyID)
                          const technologyName = serverLabel(t, choice.technology_name_keys[index], humanizeToken(choice.technology_keys[index] ?? String(technologyID)))
                          return (
                            <div className="research-tech-choice-row" key={technologyID}>
                              <span className={`research-tech-passive${selected ? ' selected' : ''}`}>
                                {technologyName}
                                {selected && <strong>{t('research.currentChoice')}</strong>}
                              </span>
                              <button
                                type="button"
                                className="research-tech-info-button"
                                aria-label={t('research.info.open', { technology: technologyName })}
                                title={t('research.info.open', { technology: technologyName })}
                                onClick={() => inspectTechnology(choice, technologyID, index)}
                              >?</button>
                            </div>
                          )
                        })}
                        <button type="button" className="research-field-select" onClick={() => selectResearch(choice)}><GameIcon name={categoryIcon} />{isActive ? t('research.reselect') : t('research.select')}</button>
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

        {inspectedTechnology && (
          <div className="research-tech-info-layer" role="presentation" onPointerDown={(event) => { if (event.target === event.currentTarget) setInspectedTechnology(null) }}>
            <section className="research-tech-info-dialog" role="dialog" aria-modal="true" aria-labelledby="research-tech-info-title">
              <header className="research-tech-info-header">
                <div>
                  <p className="eyebrow">{inspectedTechnology.categoryName}</p>
                  <h3 id="research-tech-info-title">{inspectedTechnology.name}</h3>
                </div>
                <button type="button" className="button-ghost research-tech-info-close" onClick={() => setInspectedTechnology(null)} aria-label={t('common.close')}><GameIcon name="close" /></button>
              </header>
              <dl className="research-tech-info-meta">
                <div><dt>{t('research.cost')}</dt><dd>{inspectedTechnology.choice.base_cost_rp.toFixed(0)} RP</dd></div>
                <div><dt>{t('research.info.field')}</dt><dd>#{inspectedTechnology.choice.tech_field_id}</dd></div>
                <div><dt>{t('research.info.mode')}</dt><dd>{researchSelectionModeLabel(t, inspectedTechnology.choice.selection_mode)}</dd></div>
              </dl>
              <div className="research-tech-info-content">
                <h4>{t('research.info.whatYouGet')}</h4>
                {(inspectedTechnology.info.effects?.length ?? 0) > 0 ? (
                  <ul>
                    {inspectedTechnology.info.effects?.map((effect, index) => <li key={`${effect.kind}:${effect.id ?? index}`}>{researchEffectText(t, effect)}</li>)}
                  </ul>
                ) : (
                  <p className="muted">{t('research.info.notNormalized')}</p>
                )}
                {inspectedTechnology.info.description && (
                  <div className="research-tech-info-original">
                    <h4>{t('research.info.originalDescription')}</h4>
                    <p>{inspectedTechnology.info.description}</p>
                  </div>
                )}
                <p className="research-tech-info-authority">{t('research.info.authority')}</p>
              </div>
              <footer className="research-tech-info-actions">
                <button type="button" className="button-secondary" onClick={() => setInspectedTechnology(null)}>{t('common.close')}</button>
              </footer>
            </section>
          </div>
        )}
      </section>
    </div>
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
                  <span className={'badge diplomacy-stance-badge relation-' + strategicRelationTone(snapshot.view.empire.id, relation.other_empire_id, relations)}><GameIcon name={diplomaticStanceIcon(relation.stance)} />{localizedStance(t, relation.stance)}</span>
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
                      <GameIcon name={diplomacyActionIcon(action.kind)} />{action.kind === 'diplomacy.declare_war'
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
