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
}

export type CreateGameResponse = {
  schema_version: number
  game: GameSummary
  players: NewGamePlayer[]
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

export type PopulationCohort = {
  origin_empire_id: number
  loyalty_empire_id: number
  assimilation_state: string
  farmers: number
  workers: number
  scientists: number
}

export type Colony = {
  id: number
  empire_id: number
  planet_id: number
  population: {
    cohorts?: PopulationCohort[]
  }
  ground_forces?: {
    infantry?: number
  }
}

export type DiplomaticStance = 'neutral' | 'peace' | 'war'

export type DiplomacyView = {
  other_empire_id: number
  stance: DiplomaticStance
  incoming_peace_offer?: boolean
  outgoing_peace_offer?: boolean
}

export type DiplomacyCommandKind = 'diplomacy.declare_war' | 'diplomacy.offer_peace' | 'diplomacy.accept_peace'

export type InvasionOpportunity = {
  system_id: number
  colony_id: number
  attacker_empire_id: number
  defender_empire_id: number
  attacker_seat_id: number
  eligible_transport_fleet_ids: number[]
}

export type InvasionAction = 'invade' | 'decline'

export type PlayerView = {
  game_id: string
  revision: number
  turn: number
  phase: string
  eliminated_empire_ids?: number[]
  result?: SessionResult
  seat: {
    seat: {
      id: number
      empire_id: number
      name: string
      controller: string
    }
    submitted: boolean
  }
  seats: Array<{
    seat: {
      id: number
      empire_id: number
      name: string
      controller: string
    }
    submitted: boolean
  }>
  empire: {
    id: number
    name: string
    race_id: string
  }
  colonies: Colony[]
  diplomacy?: DiplomacyView[]
  invasion?: InvasionOpportunity
}

export type BattleView = {
  spec: {
    id: number
    participants: number[]
    tactical_unsupported_reason?: string
  }
  phase: string
  tactical?: unknown
}

export type PlayerSnapshot = {
  schema_version: number
  change_sequence: number
  view: PlayerView
  battles: BattleView[]
}

export type Receipt = {
  schema_version: number
  game_id: string
  change_sequence: number
  game_revision: number
}

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

type APIError = {
  schema_version: number
  error: {
    code: string
    message: string
  }
}

export async function createGame(request: CreateGameRequest): Promise<CreateGameResponse> {
  return requestJSON<CreateGameResponse>('/api/v1/games', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  })
}
export async function listGames(signal?: AbortSignal): Promise<GameSummary[]> {
  return requestJSON<GameSummary[]>('/api/v1/games', { signal })
}

export async function getPlayerSnapshot(gameID: string, seatID: number, signal?: AbortSignal): Promise<PlayerSnapshot> {
  return requestJSON<PlayerSnapshot>(`/api/v1/games/${encodeURIComponent(gameID)}/seats/${seatID}/snapshot`, { signal })
}

export async function assignPopulation(
  snapshot: PlayerSnapshot,
  seatID: number,
  colonyID: number,
  farmers: number,
  workers: number,
  scientists: number,
): Promise<Receipt> {
  const body = {
    schema_version: 1,
    game_id: snapshot.view.game_id,
    seat_id: seatID,
    turn: snapshot.view.turn,
    base_revision: snapshot.view.revision,
    commands: [
      {
        schema_version: 1,
        sequence: 1,
        kind: 'colony.assign_population',
        payload: {
          colony_id: colonyID,
          farmers,
          workers,
          scientists,
        },
      },
    ],
  }
  return requestJSON<Receipt>(`/api/v1/games/${encodeURIComponent(snapshot.view.game_id)}/turn-submissions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

export async function submitDiplomacy(snapshot: PlayerSnapshot, seatID: number, kind: DiplomacyCommandKind, otherEmpireID: number): Promise<Receipt> {
  const payload = kind === 'diplomacy.accept_peace'
    ? { from_empire_id: otherEmpireID }
    : { target_empire_id: otherEmpireID }
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


export async function submitInvasion(snapshot: PlayerSnapshot, seatID: number, action: InvasionAction): Promise<Receipt> {
  const invasion = snapshot.view.invasion
  if (!invasion) throw new Error('No invasion opportunity is available')
  const kind = action === 'invade' ? 'invasion.invade' : 'invasion.decline'
  const payload = action === 'invade'
    ? { colony_id: invasion.colony_id, transport_fleet_ids: invasion.eligible_transport_fleet_ids }
    : { colony_id: invasion.colony_id }
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

async function requestJSON<T>(input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const response = await fetch(input, init)
  if (!response.ok) {
    let message = `${response.status} ${response.statusText}`
    try {
      const payload = (await response.json()) as APIError
      if (payload?.error?.message) {
        message = `${payload.error.code}: ${payload.error.message}`
      }
    } catch {
      // Keep the HTTP status if the response is not a JSON API error.
    }
    throw new Error(message)
  }
  return (await response.json()) as T
}

export function aggregatePopulation(colony: Colony): { farmers: number; workers: number; scientists: number } {
  return (colony.population.cohorts ?? []).reduce(
    (sum, cohort) => ({
      farmers: sum.farmers + cohort.farmers,
      workers: sum.workers + cohort.workers,
      scientists: sum.scientists + cohort.scientists,
    }),
    { farmers: 0, workers: 0, scientists: 0 },
  )
}
