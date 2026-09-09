import type { ShipVisualGenome } from './shipVisualGenome'
export type NewGamePlayer = {
  seat_id: number
  empire_id: number
  race_id: string
  name: string
}

export type NewGameSettings = {
  galaxy_size: 'small'
  galaxy_age: 'normal'
  technology_level: 'average'
  strategic_combat: false
  players: Array<{
    seat_id: number
    empire_name: string
    race_id: 'human' | 'darlok'
  }>
}

export type CreateGameRequest = {
  schema_version: 1
  game_id: string
  seed: string
  settings: NewGameSettings
  controllers?: Array<{ seat_id: number; controller: 'local_human' | 'builtin_ai' | 'remote_human' }>
}

export type SessionResult = {
  kind: 'conquest'
  winner_empire_id: number
  winner_seat_id: number
  eliminated_empire_ids: number[]
  completed_turn: number
  completed_revision: number
}

export type GameSummary = {
  schema_version: number
  game_id: string
  change_sequence: number
  revision: number
  turn: number
  phase: string
  result?: SessionResult
}

export type CreateGameResponse = {
  schema_version: number
  game: GameSummary
  players: NewGamePlayer[]
}

export type PopulationJob = 'farmer' | 'worker' | 'scientist'
export type PopulationAssimilationState = string

export type PopulationCohortKey = {
  origin_empire_id: number
  loyalty_empire_id: number
  assimilation_state: PopulationAssimilationState
}

export type PopulationCohort = PopulationCohortKey & {
  farmers: number
  workers: number
  scientists: number
}

export type ColonyEconomy = {
  food: number
  production: number
  research: number
  tax_bc: number
}

export type PopulationDynamics = {
  capacity: number
  food_required: number
  local_food_surplus: number
  local_food_shortage: number
  food_imported: number
  food_exported: number
  food_surplus: number
  food_shortage: number
  production_required: number
  production_shortage: number
  production_available: number
  base_growth: number
  growth_multiplier: number
  projected_growth: number
  projected_starvation: number
}

export type ConstructionProjectKind =
  | 'building'
  | 'colony_ship'
  | 'outpost_ship'
  | 'troop_transport'
  | 'military_ship'
  | 'freighter_fleet'
  | 'housing'
  | 'planetary_transformation'

export type ConstructionState = {
  project_kind: ConstructionProjectKind
  project_id: string
  progress_pp: number
  ship_design_id?: number
  ship_design_revision?: number
}

export type Colony = {
  id: number
  empire_id: number
  planet_id: number
  population: { cohorts?: PopulationCohort[] }
  ground_forces?: { infantry?: number }
  buildings?: string[]
  economy: ColonyEconomy
  economy_context?: Record<string, unknown>
  adjusted_economy: ColonyEconomy
  population_dynamics: PopulationDynamics
  construction?: ConstructionState
  construction_queue?: ConstructionState[]
  construction_reserve_pp?: number
}

export type EmpireCommandPoints = { capacity: number; used: number }
export type EmpireTreasury = {
  balance_bc: number
  tax_income_bc: number
  surplus_food_income_bc: number
  gross_income_bc: number
  building_maintenance_bc: number
  freighter_operating_cost_bc: number
  ship_command_maintenance_bc: number
  total_modeled_maintenance_bc: number
  net_modeled_income_bc: number
}
export type EmpireFoodLogistics = {
  freighters_required: number
  freighters_used: number
  population_transport_freighters_reserved: number
  freighters_available_for_food: number
  food_transferred: number
  food_unmet: number
}
export type ResearchState = {
  tech_field_id: number
  selection_mode: ResearchSelectionMode
  technology_ids: number[]
  progress_rp: number
}
export type Empire = {
  id: number
  name: string
  race_id: string
  capital_colony_id?: number
  freighters: number
  command_points: EmpireCommandPoints
  treasury: EmpireTreasury
  food_logistics: EmpireFoodLogistics
  known_technology_ids?: number[]
  known_technology_field_ids?: number[]
  research?: ResearchState
}

export type Planet = {
  id: number
  name: string
  orbit: number
  size_id: string
  mineral_id: string
  gravity_id: string
  climate_id: string
}
export type OrbitalBodyKind = 'planet' | 'gas_giant' | 'asteroid_belt'
export type OrbitalBody = {
  id: number
  name: string
  orbit: number
  kind: OrbitalBodyKind
  planet_id?: number
  outpost_id?: number
}
export type StarSystem = {
  id: number
  name: string
  x: number
  y: number
  spectral_class: number
  planets: Planet[]
  bodies?: OrbitalBody[]
}
export type Galaxy = { id: number; systems: StarSystem[] }
export type PlanetPotential = {
  planet_id: number
  food_per_farmer: number
  production_per_worker: number
  research_per_scientist: number
  gravity_penalty_percent: number
  climate_habitability_percent: number
  size_base_capacity: number
  population_capacity: number
}
export type Outpost = { id: number; empire_id: number; body_id?: number; planet_id?: number }

export type ShipWeaponMount = { slot: number; weapon_id: string; count: number }
export type ShipVisualGenomeWire = {
  version: 4
  hull_id: string
  style_id: 'spear' | 'sleek' | 'organic'
  morphology_id: 'needle' | 'barge' | 'manta' | 'fork' | 'chevron' | 'hammer' | 'bulb'
  seed: string
  length: number
  beam: number
  station_count: number
  station_widths: number[]
  notch_depths: number[]
  engine_count: number
  detail_count: number
  cutouts?: Array<{ t: number; offset: number; rx: number; ry: number; angle: number }>
  primitives?: Array<{ kind: 'wedge' | 'spike' | 'pod'; t: number; length: number; width: number; sweep: number }>
}
export type ShipDesignSpec = {
  hull_id: string
  strategic_picture_id: number
  warp_drive_id: string
  ftl_speed: number
  computer_id: string
  armor_id: string
  shield_id?: string
  fuel_cell_id: string
  fuel_range_parsecs: number
  production_cost_pp: number
  weapons?: ShipWeaponMount[]
}
export type ShipDesign = { id: number; empire_id: number; revision: number; visual_revision?: number; name: string; spec: ShipDesignSpec; visual_genome?: ShipVisualGenomeWire }
export type Ship = { id: number; empire_id: number; source_design_id: number; source_design_revision: number; source_visual_revision?: number; name: string; spec: ShipDesignSpec; visual_genome?: ShipVisualGenomeWire }
export type StrategicFleet = {
  id: number
  empire_id: number
  role: string
  special_kind?: string
  at_system_id?: number
  destination_system_id?: number
  remaining_turns?: number
  ftl_speed?: number
  ship_ids?: number[]
}
export type PopulationTransfer = {
  id: number
  empire_id: number
  source_colony_id: number
  destination_colony_id: number
  origin_empire_id: number
  loyalty_empire_id: number
  assimilation_state: string
  job?: PopulationJob
  source_job?: PopulationJob
  destination_job?: PopulationJob
  remaining_turns: number
}
export type StrategicContact = {
  kind: 'colony' | 'outpost' | 'fleet'
  empire_id: number
  colony_id?: number
  outpost_id?: number
  fleet_id?: number
  system_id?: number
  planet_id?: number
  destination_system_id?: number
  remaining_turns?: number
  role?: string
  special_kind?: string
}
export type StrategicView = {
  galaxy: Galaxy
  planet_potentials?: PlanetPotential[]
  outposts?: Outpost[]
  ship_designs?: ShipDesign[]
  ships?: Ship[]
  fleets?: StrategicFleet[]
  population_transfers?: PopulationTransfer[]
  contacts?: StrategicContact[]
}

export type DiplomaticStance = 'neutral' | 'peace' | 'war'
export type DiplomacyView = {
  other_empire_id: number
  stance: DiplomaticStance
  incoming_peace_offer?: boolean
  outgoing_peace_offer?: boolean
}
export type DiplomacyCommandKind = 'diplomacy.declare_war' | 'diplomacy.offer_peace' | 'diplomacy.accept_peace'
export type DiplomacyDecision = { kind: DiplomacyCommandKind; target_empire_id: number }

export type InvasionOpportunity = {
  system_id: number
  colony_id: number
  attacker_empire_id: number
  defender_empire_id: number
  attacker_seat_id: number
  eligible_transport_fleet_ids: number[]
}
export type InvasionAction = 'invade' | 'decline'

export type ConstructionChoice = {
  project_kind: ConstructionProjectKind
  project_id: string
  production_cost_pp: number
  technology_id?: number
  production_id?: number
  maintenance_bc?: number
  freighters_added?: number
  ship_design_id?: number
  ship_design_revision?: number
  ship_design_name?: string
}
export type ColonyConstructionDecision = { colony_id: number; choices: ConstructionChoice[] }
export type PopulationChoice = {
  colony_id: number
  farmers: number
  workers: number
  scientists: number
  adjusted_economy: ColonyEconomy
  population_dynamics: PopulationDynamics
  food_safe: boolean
}
export type ColonyPopulationDecision = { colony_id: number; choices: PopulationChoice[] }
export type FleetMoveChoice = {
  fleet_id: number
  source_system_id: number
  destination_system_id: number
  eta: number
  fuel_range_parsecs: number
  supply_distance_parsecs: number
  ship_ids?: number[]
}
export type ColonizationChoice = { fleet_id: number; system_id: number; planet_id: number }
export type OutpostDeploymentChoice = { fleet_id: number; system_id: number; body_id: number; planet_id?: number }
export type PopulationTransferChoice = {
  source_colony_id: number
  destination_colony_id: number
  cohort: PopulationCohortKey
  source_job: PopulationJob
  destination_job: PopulationJob
  amount: number
  same_system: boolean
  eta: number
  freighters_required: number
}
export type ResearchSelectionMode = 'all' | 'choose_one' | 'fixed_one' | 'repeat_field'
export type ResearchCategory = { id: string; order: number; name_key: string; root_tech_field_id: number }
export type ResearchTechnologyEffect = {
  kind: 'building_unlock' | 'planetary_project_unlock' | 'ship_drive_unlock' | 'ship_computer_unlock' | 'ship_armor_unlock' | 'ship_shield_unlock' | 'ship_fuel_cell_unlock' | 'population_growth_bonus' | 'population_capacity_bonus' | string
  id?: string
  production_cost_pp?: number
  maintenance_bc?: number
  ftl_speed?: number
  range_parsecs?: number
  population_bonus?: number
  growth_bonus?: number
}
export type ResearchTechnologyInfo = {
  technology_id: number
  technology_key: string
  technology_name_key: string
  description?: string
  effects?: ResearchTechnologyEffect[]
}
export type ResearchChoice = {
  category_id: string
  category_order: number
  category_name_key: string
  tech_field_id: number
  previous_tech_field_id: number
  next_tech_field_id: number
  base_cost_rp: number
  selection_mode: ResearchSelectionMode
  technology_ids: number[]
  technology_keys: string[]
  technology_name_keys: string[]
  technology_info?: ResearchTechnologyInfo[]
  completed_levels?: number
  research_level?: number
}
export type ColonyBaseResolution = {
  empire_id: number
  source_colony_id: number
  system_id: number
  target_planet_ids: number[]
  trash_refund_bc: number
}
export type BattleSide = {
  empire_id: number
  seat_id: number
  combat_fleet_ids: number[]
  ship_ids: number[]
  civilian_fleet_ids?: number[]
}
export type BattleSpec = {
  id: number
  game_id: string
  strategic_turn: number
  system_id?: number
  attacker?: BattleSide
  defender?: BattleSide
  defender_colony_ids?: number[]
  participants: number[]
  seed: number
  tactical?: unknown
  tactical_unsupported_reason?: string
}
export type BattleResult = {
  winner_seat: number
  winner_seats?: number[]
  outcome: string
  destroyed_ship_ids?: number[]
}
export type BattleView = {
  spec: BattleSpec
  phase: string
  result?: BattleResult
  tactical?: unknown
}
export type ProtocolCommand = { schema_version: 1; sequence: number; kind: string; payload: Record<string, unknown> }
export type BattleDecision = { battle_id: number; actions: ProtocolCommand[] }
export type DecisionCatalog = {
  research_categories?: ResearchCategory[]
  research?: ResearchChoice[]
  construction?: ColonyConstructionDecision[]
  population?: ColonyPopulationDecision[]
  population_transfers?: PopulationTransferChoice[]
  fleet_moves?: FleetMoveChoice[]
  colonization?: ColonizationChoice[]
  outpost_deployment?: OutpostDeploymentChoice[]
  diplomacy?: DiplomacyDecision[]
  colony_base?: ColonyBaseResolution[]
  invasion?: InvasionOpportunity
  battles?: BattleDecision[]
}
export type PublicEmpireIdentity = { id: number; name: string; race_id: string }
export type ResolutionSummaryKind =
  | 'research_breakthrough'
  | 'technology_granted'
  | 'battle_completed'
  | 'invasion_resolved'
  | 'invasion_declined'
  | 'empire_eliminated'
  | 'game_completed'
export type ResearchResolutionSummary = {
  empire_id: number
  tech_field_id: number
  selection_mode: ResearchSelectionMode
  technology_ids?: number[]
  technology_keys?: string[]
  completed_levels?: number
  research_level?: number
}
export type TechnologyGrantResolutionSummary = {
  empire_id: number
  technology_id: number
  technology_key: string
  technology_name_key: string
  tech_field_id: number
  source_kind: string
}
export type BattleResolutionSummary = {
  battle_id: number
  system_id?: number
  attacker_empire_id?: number
  defender_empire_id?: number
  winner_seat_id: number
  winner_empire_id?: number
  outcome: string
  destroyed_ship_ids?: number[]
  surviving_ship_ids?: number[]
  defender_colony_ids?: number[]
}
export type InvasionResolutionSummary = {
  system_id: number
  colony_id: number
  attacker_empire_id: number
  defender_empire_id: number
  outcome: string
  selected_transport_fleet_ids?: number[]
  initial_attacker_infantry?: number
  initial_defender_infantry?: number
  initial_defender_militia?: number
  surviving_attacker_infantry?: number
  surviving_defender_infantry?: number
  surviving_defender_militia?: number
  consumed_transport_fleet_ids?: number[]
  surviving_transport_fleet_ids?: number[]
}
export type ResolutionSummary = {
  id: string
  event_sequence: number
  turn: number
  revision: number
  kind: ResolutionSummaryKind
  research?: ResearchResolutionSummary
  technology_grant?: TechnologyGrantResolutionSummary
  battle?: BattleResolutionSummary
  invasion?: InvasionResolutionSummary
  eliminated_empire_id?: number
  result?: SessionResult
}
export type PlayerDecisionView = {
  game_id: string
  revision: number
  turn: number
  phase: string
  seat: PlayerView['seat']
  empire: Empire
  public_empires?: PublicEmpireIdentity[]
  colonies: Colony[]
  diplomacy?: DiplomacyView[]
  strategic: StrategicView
  decisions: DecisionCatalog
}

export type PlayerView = {
  game_id: string
  revision: number
  turn: number
  phase: string
  eliminated_empire_ids?: number[]
  result?: SessionResult
  seat: { seat: { id: number; empire_id: number; name: string; controller: string }; submitted: boolean }
  seats: Array<{ seat: { id: number; empire_id: number; name: string; controller: string }; submitted: boolean }>
  empire: { id: number; name: string; race_id: string }
  colonies: Colony[]
  diplomacy?: DiplomacyView[]
  invasion?: InvasionOpportunity
  recent_resolutions?: ResolutionSummary[]
}

export type PlayerSnapshot = {
  schema_version: number
  change_sequence: number
  view: PlayerView
  decision?: PlayerDecisionView
  battles: BattleView[]
}

export type PlanningCommandPayload = Record<string, unknown>
export type DraftOrder = { key: string; kind: string; payload: PlanningCommandPayload }
export type CommandBatch = {
  schema_version: 1
  game_id: string
  seat_id: number
  turn: number
  base_revision: number
  commands: ProtocolCommand[]
}

export type PlanningMetricComponent = { id: string; value: number }
export type PlanningMetricBreakdown = { total: number; components?: PlanningMetricComponent[] }
export type PlanningColonyBreakdowns = { growth: PlanningMetricBreakdown; tax_bc: PlanningMetricBreakdown }

export type PlanningConstructionPreview = { project: ConstructionState; cost_pp: number; remaining_pp: number; eta_turns?: number }
export type PlanningColonyPreview = {
  colony: Colony
  breakdowns: PlanningColonyBreakdowns
  free_population_capacity: number
  population_growth_per_turn: number
  population_loss_per_turn: number
  next_population_eta_turns?: number
  construction: PlanningConstructionPreview[]
}
export type PlanningPreviewProjection = {
  empire_id: number
  treasury_balance_bc: number
  net_modeled_income_bc: number
  freighters: { total: number; food_used: number; transfer_reserved: number; available: number }
  command_points: EmpireCommandPoints
  command_point_overage: number
  research: {
    tech_field_id?: number
    selection_mode?: ResearchSelectionMode
    technology_ids?: number[]
    progress_rp: number
    cost_rp: number
    remaining_rp: number
    rp_per_turn: number
    eta_turns?: number
  }
  colonies: PlanningColonyPreview[]
  population_transfers?: PopulationTransferChoice[]
}
export type PlanningPreviewSnapshot = {
  schema_version: number
  change_sequence: number
  preview: { game_id: string; turn: number; base_revision: number; projection: PlanningPreviewProjection }
}

export type Receipt = { schema_version: number; game_id: string; change_sequence: number; game_revision: number }
export type Notification = {
  schema_version: number
  kind: 'snapshot_invalidated'
  game_id: string
  change_sequence: number
  game_revision: number
  scope: 'session' | 'battle' | 'observer'
  battle_id?: number
  reason: string
}

type APIErrorEnvelope = { schema_version: number; error: { code: string; message: string } }

export class APIError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
  }
}

export function isAPIError(reason: unknown): reason is APIError {
  return reason instanceof APIError
}

export async function createGame(request: CreateGameRequest): Promise<CreateGameResponse> {
  return requestJSON<CreateGameResponse>('/api/v1/games', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(request) })
}
export async function listGames(signal?: AbortSignal): Promise<GameSummary[]> {
  return requestJSON<GameSummary[]>('/api/v1/games', { signal })
}
export async function getPlayerSnapshot(gameID: string, seatID: number, signal?: AbortSignal): Promise<PlayerSnapshot> {
  return requestJSON<PlayerSnapshot>(`/api/v1/games/${encodeURIComponent(gameID)}/seats/${seatID}/snapshot`, { signal })
}

export async function exportLiveSnapshot(gameID: string): Promise<Blob> {
  return requestBlob(`/api/v1/games/${encodeURIComponent(gameID)}/live-snapshot`)
}

export async function importLiveSnapshot(save: Blob): Promise<GameSummary> {
  return requestJSON<GameSummary>('/api/v1/games/import', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: save,
  })
}

export async function restoreLiveSnapshot(gameID: string, save: Blob): Promise<Receipt> {
  return requestJSON<Receipt>(`/api/v1/games/${encodeURIComponent(gameID)}/live-snapshot`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: save,
  })
}

export function buildCommandBatch(snapshot: PlayerSnapshot, seatID: number, orders: DraftOrder[]): CommandBatch {
  return {
    schema_version: 1,
    game_id: snapshot.view.game_id,
    seat_id: seatID,
    turn: snapshot.view.turn,
    base_revision: snapshot.view.revision,
    commands: orders.map((order, index) => ({ schema_version: 1, sequence: index + 1, kind: order.kind, payload: order.payload })),
  }
}

export async function previewPlanning(snapshot: PlayerSnapshot, seatID: number, orders: DraftOrder[], signal?: AbortSignal): Promise<PlanningPreviewSnapshot> {
  const body = buildCommandBatch(snapshot, seatID, orders)
  return requestJSON<PlanningPreviewSnapshot>(`/api/v1/games/${encodeURIComponent(snapshot.view.game_id)}/seats/${seatID}/planning-preview`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body), signal,
  })
}

export async function submitPlanning(snapshot: PlayerSnapshot, seatID: number, orders: DraftOrder[]): Promise<Receipt> {
  const body = buildCommandBatch(snapshot, seatID, orders)
  return requestJSON<Receipt>(`/api/v1/games/${encodeURIComponent(snapshot.view.game_id)}/turn-submissions`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body),
  })
}

export function encodeShipVisualGenome(genome: ShipVisualGenome): ShipVisualGenomeWire {
  return {
    version: 4,
    hull_id: String(genome.hullId),
    style_id: genome.styleId,
    morphology_id: genome.morphologyId,
    seed: genome.seed,
    length: genome.length,
    beam: genome.beam,
    station_count: genome.stationCount,
    station_widths: [...genome.stationWidths],
    notch_depths: [...genome.notchDepths],
    engine_count: genome.engineCount,
    detail_count: genome.detailCount,
    ...(genome.cutouts.length > 0 ? { cutouts: genome.cutouts.map((cutout) => ({ ...cutout })) } : {}),
    ...(genome.primitives.length > 0 ? { primitives: genome.primitives.map((primitive) => ({ ...primitive })) } : {}),
  }
}

export function decodeShipVisualGenome(genome?: ShipVisualGenomeWire): ShipVisualGenome | undefined {
  if (!genome) return undefined
  return {
    version: 4,
    hullId: genome.hull_id,
    styleId: genome.style_id,
    morphologyId: genome.morphology_id,
    seed: genome.seed,
    length: genome.length,
    beam: genome.beam,
    stationCount: genome.station_count,
    stationWidths: [...genome.station_widths],
    notchDepths: [...genome.notch_depths],
    engineCount: genome.engine_count,
    detailCount: genome.detail_count,
    cutouts: (genome.cutouts ?? []).map((cutout) => ({ ...cutout })),
    primitives: (genome.primitives ?? []).map((primitive) => ({ ...primitive })),
  }
}

export async function submitMilitaryDesignVisual(snapshot: PlayerSnapshot, seatID: number, designID: number, genome: ShipVisualGenome): Promise<Receipt> {
  return requestJSON<Receipt>(`/api/v1/games/${encodeURIComponent(snapshot.view.game_id)}/immediate-commands`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      schema_version: 1,
      seat_id: seatID,
      base_revision: snapshot.view.revision,
      command: {
        schema_version: 1,
        sequence: 1,
        kind: 'empire.set_military_design_visual',
        payload: { design_id: designID, visual_genome: encodeShipVisualGenome(genome) },
      },
    }),
  })
}
export async function submitDiplomacy(snapshot: PlayerSnapshot, seatID: number, kind: DiplomacyCommandKind, otherEmpireID: number): Promise<Receipt> {
  const payload = kind === 'diplomacy.accept_peace' ? { from_empire_id: otherEmpireID } : { target_empire_id: otherEmpireID }
  return requestJSON<Receipt>(`/api/v1/games/${encodeURIComponent(snapshot.view.game_id)}/immediate-commands`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ schema_version: 1, seat_id: seatID, base_revision: snapshot.view.revision, command: { schema_version: 1, sequence: 1, kind, payload } }),
  })
}

export async function submitInvasion(snapshot: PlayerSnapshot, seatID: number, action: InvasionAction): Promise<Receipt> {
  const invasion = snapshot.view.invasion
  if (!invasion) throw new Error('No invasion opportunity is available')
  const kind = action === 'invade' ? 'invasion.invade' : 'invasion.decline'
  const payload = action === 'invade' ? { colony_id: invasion.colony_id, transport_fleet_ids: invasion.eligible_transport_fleet_ids } : { colony_id: invasion.colony_id }
  return requestJSON<Receipt>(`/api/v1/games/${encodeURIComponent(snapshot.view.game_id)}/immediate-commands`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ schema_version: 1, seat_id: seatID, base_revision: snapshot.view.revision, command: { schema_version: 1, sequence: 1, kind, payload } }),
  })
}

export async function submitColonyBase(snapshot: PlayerSnapshot, seatID: number, resolution: ColonyBaseResolution, action: 'colonize' | 'trash', planetID?: number): Promise<Receipt> {
  const kind = action === 'colonize' ? 'colony.colonize_with_base' : 'colony.trash_colony_base'
  const payload = action === 'colonize'
    ? { source_colony_id: resolution.source_colony_id, planet_id: planetID }
    : { source_colony_id: resolution.source_colony_id }
  return requestJSON<Receipt>(`/api/v1/games/${encodeURIComponent(snapshot.view.game_id)}/immediate-commands`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      schema_version: 1,
      seat_id: seatID,
      base_revision: snapshot.view.revision,
      command: { schema_version: 1, sequence: 1, kind, payload },
    }),
  })
}

export function streamURL(gameID: string): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}/api/v1/games/${encodeURIComponent(gameID)}/stream`
}

async function readAPIError(response: Response): Promise<APIError> {
  let code = 'http_error'
  let message = `${response.status} ${response.statusText}`
  try {
    const payload = (await response.json()) as APIErrorEnvelope
    if (payload?.error?.code) code = payload.error.code
    if (payload?.error?.message) message = payload.error.message
  } catch {
    // Retain HTTP status and fallback code.
  }
  return new APIError(response.status, code, message)
}

async function requestBlob(input: RequestInfo | URL, init?: RequestInit): Promise<Blob> {
  const response = await fetch(input, init)
  if (!response.ok) throw await readAPIError(response)
  return response.blob()
}

async function requestJSON<T>(input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const response = await fetch(input, init)
  if (!response.ok) throw await readAPIError(response)
  return (await response.json()) as T
}

export function aggregatePopulation(colony: Colony): { farmers: number; workers: number; scientists: number; total: number } {
  const result = (colony.population.cohorts ?? []).reduce(
    (sum, cohort) => ({ farmers: sum.farmers + cohort.farmers, workers: sum.workers + cohort.workers, scientists: sum.scientists + cohort.scientists }),
    { farmers: 0, workers: 0, scientists: 0 },
  )
  return { ...result, total: result.farmers + result.workers + result.scientists }
}

export function replaceDraftOrder(orders: DraftOrder[], order: DraftOrder): DraftOrder[] {
  return [...orders.filter((item) => item.key !== order.key), order]
}

export function removeDraftOrder(orders: DraftOrder[], key: string): DraftOrder[] {
  return orders.filter((item) => item.key !== key)
}
