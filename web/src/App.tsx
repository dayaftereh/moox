import { type ChangeEvent, type FormEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  aggregatePopulation,
  createGame,
  exportLiveSnapshot,
  getCompositionCatalog,
  getDifficultyCatalog,
  getGalaxyCatalog,
  getRaceCatalog,
  getTechnologyCatalog,
  getPlayerSnapshot,
  importLiveSnapshot,
  listGames,
  previewPlanning,
  removeDraftOrder,
  replaceDraftOrder,
  restoreLiveSnapshot,
  savePlanningDraft,
  streamURL,
  submitBattleCommand,
  submitColonyBase,
  submitDiplomacy,
  submitInvasion,
  submitPlanning,
  type DifficultyCatalog,
  type DifficultyID,
  type DifficultyProfile,
  type GalaxyAgeID,
  type GalaxyAgeProfile,
  type GalaxyBias,
  type GalaxyCatalog,
  type GalaxySizeID,
  type GalaxySizeProfile,
  type PresetRaceCatalog,
  type PresetRaceID,
  type PresetRaceProfile,
  type NewGameCompositionCatalog,
  type NewGameOpponentAssignment,
  type NewGameTechnologyCatalog,
  type NewGameTechnologyLevel,
  type NewGameTechnologyProfile,
  type DiplomacyCommandKind,
  type DiplomaticStance,
  type DraftOrder,
  type GameSummary,
  isAPIError,
  type Notification,
  type PlanningPreviewSnapshot,
  type PlayerSnapshot,
  type ProtocolCommand,
  type ResolutionSummary,
} from './api'
import { AppShell, LanguageSwitch, StandaloneHeader, type ResourceChip } from './components/AppShell'
import { BattleRouteView } from './BattleRouteView'
import { GameIcon } from './components/GameIcon'
import { VisualSelector } from './components/VisualSelector'
import { OrbitalBodyArt } from './components/OrbitalBodyArt'
import { Card, EmptyState, Metric, Notice, PageHeader } from './components/ui'
import { type TranslationKey, type TranslationVars, useI18n } from './i18n'
import { type AppRoute, type GameSection, navigate, parseRoute } from './navigation'
import { newGameAssetPath } from './newGameAssets'
import { ShipBuilderView } from './ShipBuilderView'
import {
  StrategicColoniesView,
  StrategicConstructionView,
  StrategicDiplomacyView,
  StrategicEspionageView,
  StrategicFleetsView,
  StrategicGalaxyView,
  StrategicResearchOverlay,
} from './StrategicViews'
import './styles.css'

const galaxySizeIDs: readonly GalaxySizeID[] = ['small', 'medium', 'large', 'huge']
const galaxyAgeIDs: readonly GalaxyAgeID[] = ['mineral_rich', 'normal', 'organic_rich']
const difficultyIDs: readonly DifficultyID[] = ['easy', 'normal', 'hard', 'very_hard', 'impossible']
const runtimeRacePortraitIDs = new Set<PresetRaceID>(['human', 'klackon', 'darlok'])
const technologyLevelIDs: readonly NewGameTechnologyLevel[] = ['pre_warp', 'average', 'advanced']
const technologyLevelAssetVersion = 'slice16-5-g3-art2'
const opponentCountIDs = ['1', '2', '3', '4', '5', '6', '7'] as const
type OpponentCountID = (typeof opponentCountIDs)[number]

function GalaxySizeArt({ size }: { size: GalaxySizeID }) {
  return (
    <div className="new-game-galaxy-art" aria-hidden="true">
      <img src={newGameAssetPath('galaxy-size', size, 'svg')} alt="" draggable={false} />
    </div>
  )
}

function galaxyAgeAssetOptionID(age: GalaxyAgeID): string {
  if (age === 'mineral_rich') return 'mineral-rich'
  if (age === 'organic_rich') return 'organic-rich'
  return age
}

function GalaxyAgeArt({ age }: { age: GalaxyAgeID }) {
  return (
    <div className="new-game-galaxy-art" aria-hidden="true">
      <img src={newGameAssetPath('galaxy-age', galaxyAgeAssetOptionID(age), 'svg')} alt="" draggable={false} />
    </div>
  )
}

function galaxySizeTitleKey(id: GalaxySizeID): TranslationKey {
  switch (id) {
    case 'small': return 'newGame.sizeSmall'
    case 'medium': return 'newGame.sizeMedium'
    case 'large': return 'newGame.sizeLarge'
    case 'huge': return 'newGame.sizeHuge'
  }
}

function galaxyAgeTitleKey(id: GalaxyAgeID): TranslationKey {
  switch (id) {
    case 'mineral_rich': return 'newGame.ageMineralRich'
    case 'normal': return 'newGame.ageNormal'
    case 'organic_rich': return 'newGame.ageOrganicRich'
  }
}

function galaxyBiasDetailKey(kind: 'mineral' | 'food', bias: GalaxyBias): TranslationKey {
  if (kind === 'mineral') {
    if (bias === 'higher') return 'newGame.mineralBiasHigher'
    if (bias === 'lower') return 'newGame.mineralBiasLower'
    return 'newGame.mineralBiasBaseline'
  }
  if (bias === 'higher') return 'newGame.foodBiasHigher'
  if (bias === 'lower') return 'newGame.foodBiasLower'
  return 'newGame.foodBiasBaseline'
}

function galaxySizeDetails(t: Translator, profile: GalaxySizeProfile): string[] {
  return [t('newGame.galaxyStarSystems', { count: profile.star_count })]
}

function galaxyAgeDetails(t: Translator, profile: GalaxyAgeProfile): string[] {
  return [
    t(galaxyBiasDetailKey('mineral', profile.mineral_resource_bias)),
    t(galaxyBiasDetailKey('food', profile.food_world_bias)),
  ]
}

function RaceArt({ race }: { race: PresetRaceID }) {
  if (runtimeRacePortraitIDs.has(race)) {
    return (
      <div className="new-game-race-art" aria-hidden="true">
        <img src={`/assets/races/${race}/portrait.webp`} alt="" draggable={false} />
      </div>
    )
  }
  return (
    <div className="new-game-race-art new-game-race-art-planned" aria-hidden="true">
      <div className="new-game-race-placeholder">
        <span className="new-game-race-placeholder-orbit" />
        <span className="new-game-race-placeholder-core" />
        <span className="new-game-race-placeholder-arc" />
      </div>
    </div>
  )
}

function raceTitle(t: Translator, profile: PresetRaceProfile): string {
  return t(profile.name_key as TranslationKey)
}

function raceDetails(t: Translator, profile: PresetRaceProfile): string[] {
  const facts = profile.card_fact_trait_ids.map((traitID) => t(`raceTrait.${traitID}` as TranslationKey))
  return profile.player_availability === 'planned' ? [t('newGame.racePlannedDetail'), ...facts] : facts
}

function technologyLevelAssetOptionID(level: NewGameTechnologyLevel): string {
  return level === 'pre_warp' ? 'pre-warp' : level
}

function TechnologyLevelArt({ level }: { level: NewGameTechnologyLevel }) {
  return (
    <div className="new-game-technology-art" aria-hidden="true">
      <img src={`${newGameAssetPath('technology-level', technologyLevelAssetOptionID(level), 'svg')}?v=${technologyLevelAssetVersion}`} alt="" draggable={false} />
    </div>
  )
}

function OpponentCountArt({ count }: { count: OpponentCountID }) {
  return (
    <div className="new-game-opponent-count-art" aria-hidden="true">
      <img src={newGameAssetPath('opponent-count', count, 'svg')} alt="" draggable={false} />
    </div>
  )
}

function technologyTitle(t: Translator, profile: NewGameTechnologyProfile): string {
  return t(profile.name_key as TranslationKey)
}
function difficultyAssetOptionID(difficulty: DifficultyID): string {
  return difficulty === 'very_hard' ? 'very-hard' : difficulty
}

function DifficultyArt({ difficulty }: { difficulty: DifficultyID }) {
  return (
    <div className="new-game-difficulty-art" aria-hidden="true">
      <img src={newGameAssetPath('difficulty', difficultyAssetOptionID(difficulty), 'svg')} alt="" draggable={false} />
    </div>
  )
}

function difficultyTitleKey(id: DifficultyID): TranslationKey {
  switch (id) {
    case 'easy': return 'newGame.difficultyEasy'
    case 'normal': return 'newGame.difficultyNormal'
    case 'hard': return 'newGame.difficultyHard'
    case 'very_hard': return 'newGame.difficultyVeryHard'
    case 'impossible': return 'newGame.difficultyImpossible'
  }
}

function difficultyNumber(eighths: number, signed = true): string {
  const value = eighths / 8
  const formatted = new Intl.NumberFormat(document.documentElement.lang || 'en', { maximumFractionDigits: 3 }).format(Math.abs(value))
  if (!signed || value === 0) return formatted
  return `${value > 0 ? '+' : '-'}${formatted}`
}

function difficultyDetails(t: Translator, profile: DifficultyProfile): string[] {
  return [
    t('newGame.difficultyAiOnly'),
    t('newGame.difficultyFood', { value: difficultyNumber(profile.ai_food_per_farmer_eighths) }),
    t('newGame.difficultyProduction', { value: difficultyNumber(profile.ai_production_per_worker_eighths) }),
    t('newGame.difficultyResearch', { value: difficultyNumber(profile.ai_research_per_scientist_eighths) }),
    t('newGame.difficultyTax', { value: difficultyNumber(profile.ai_tax_bc_per_population_eighths) }),
    t('newGame.difficultyCommandDeficit', { value: difficultyNumber(profile.ai_command_deficit_bc_per_point_eighths, false) }),
  ]
}
type AssignmentDraft = {
  farmers: string
  workers: string
  scientists: string
}

type StatusMessage = {
  key: TranslationKey
  vars?: TranslationVars
}

type GameLifecycle = 'initial-loading' | 'connecting' | 'synced' | 'refreshing' | 'reconnecting' | 'refresh-failed' | 'exporting' | 'loading-file' | 'restoring' | 'importing' | 'loaded' | 'fatal'

function lifecycleKey(lifecycle: GameLifecycle): TranslationKey {
  switch (lifecycle) {
    case 'initial-loading': return 'lifecycle.initialLoading'
    case 'connecting': return 'lifecycle.connecting'
    case 'synced': return 'lifecycle.synced'
    case 'refreshing': return 'lifecycle.refreshing'
    case 'reconnecting': return 'lifecycle.reconnecting'
    case 'refresh-failed': return 'lifecycle.refreshFailed'
    case 'exporting': return 'lifecycle.exporting'
    case 'loading-file': return 'lifecycle.loadingFile'
    case 'restoring': return 'lifecycle.restoring'
    case 'importing': return 'lifecycle.importing'
    case 'loaded': return 'lifecycle.loaded'
    case 'fatal': return 'lifecycle.fatal'
  }
}

function generateNewGameSeed(): string {
  const words = new Uint32Array(2)
  crypto.getRandomValues(words)
  if (words[0] === 0 && words[1] === 0) words[1] = 1
  return `0x${words[0].toString(16).padStart(8, '0')}${words[1].toString(16).padStart(8, '0')}`
}

function errorText(reason: unknown): string {
  if (isAPIError(reason)) return `${reason.code}: ${reason.message}`
  if (reason instanceof Error) return reason.message
  return String(reason)
}

function isObsoletePlanningPreviewConflict(reason: unknown): boolean {
  if (!isAPIError(reason) || reason.status !== 409 || reason.code !== 'session_rejected') return false
  return reason.message.includes('command batch base revision')
    || reason.message.includes('command batch targets turn')
    || reason.message.includes('already submitted turn')
    || reason.message.includes('cannot preview turn in phase')
}

type SaveFileMetadata = {
  gameID: string
  turn?: number
  revision?: number
  phase?: string
  schemaVersion?: number
}

type PendingRestore = { file: File; metadata: SaveFileMetadata }
function resolutionAckStorageKey(gameID: string, seatID: number): string {
  return `moox:resolution-acks:${gameID}:seat-${seatID}`
}

function readResolutionAcks(gameID: string, seatID: number): Set<string> {
  if (!gameID || seatID <= 0) return new Set()
  try {
    const raw = window.sessionStorage.getItem(resolutionAckStorageKey(gameID, seatID))
    if (!raw) return new Set()
    const values = JSON.parse(raw)
    return new Set(Array.isArray(values) ? values.filter((value): value is string => typeof value === 'string') : [])
  } catch {
    return new Set()
  }
}

function humanizeResolutionToken(value: string): string {
  const normalized = value.replace(/^technology\./, '').replace(/[._-]+/g, ' ').trim()
  return normalized ? normalized.charAt(0).toUpperCase() + normalized.slice(1) : value
}

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

function App() {
  const { t } = useI18n()
  const [route, setRoute] = useState<AppRoute>(() => parseRoute())
  const [games, setGames] = useState<GameSummary[]>([])
  const [newGameSeed, setNewGameSeed] = useState('0x8009')
  const [galaxySizeID, setGalaxySizeID] = useState<GalaxySizeID>('small')
  const [galaxyAgeID, setGalaxyAgeID] = useState<GalaxyAgeID>('normal')
  const [galaxyCatalog, setGalaxyCatalog] = useState<GalaxyCatalog | null>(null)
  const [galaxyCatalogError, setGalaxyCatalogError] = useState('')
  const [difficultyID, setDifficultyID] = useState<DifficultyID>('normal')
  const [difficultyCatalog, setDifficultyCatalog] = useState<DifficultyCatalog | null>(null)
  const [difficultyCatalogError, setDifficultyCatalogError] = useState('')
  const [playerRaceID, setPlayerRaceID] = useState<PresetRaceID>('human')
  const [raceCatalog, setRaceCatalog] = useState<PresetRaceCatalog | null>(null)
  const [raceCatalogError, setRaceCatalogError] = useState('')
  const [technologyLevelID, setTechnologyLevelID] = useState<NewGameTechnologyLevel>('average')
  const [technologyCatalog, setTechnologyCatalog] = useState<NewGameTechnologyCatalog | null>(null)
  const [technologyCatalogError, setTechnologyCatalogError] = useState('')
  const [opponentCountID, setOpponentCountID] = useState<OpponentCountID>('1')
  const [compositionCatalog, setCompositionCatalog] = useState<NewGameCompositionCatalog | null>(null)
  const [compositionCatalogError, setCompositionCatalogError] = useState('')
  const [playerName, setPlayerName] = useState('Human')
  const [creatingGame, setCreatingGame] = useState(false)
  const [gameID, setGameID] = useState('')
  const [seatID, setSeatID] = useState(1)
  const [snapshot, setSnapshot] = useState<PlayerSnapshot | null>(null)
  const [assignment, setAssignment] = useState<AssignmentDraft>({ farmers: '0', workers: '0', scientists: '0' })
  const [draftOrders, setDraftOrders] = useState<DraftOrder[]>([])
  const [planningPreview, setPlanningPreview] = useState<PlanningPreviewSnapshot | null>(null)
  const [planningPreviewError, setPlanningPreviewError] = useState('')
  const [previewBusy, setPreviewBusy] = useState(false)
  const [planningBusy, setPlanningBusy] = useState(false)
  const [status, setStatus] = useState<StatusMessage>({ key: 'status.connecting' })
  const [lifecycle, setLifecycle] = useState<GameLifecycle>('initial-loading')
  const [error, setError] = useState('')
  const [diplomacyBusy, setDiplomacyBusy] = useState(false)
  const [invasionBusy, setInvasionBusy] = useState(false)
  const [colonyBaseBusy, setColonyBaseBusy] = useState(false)
  const [researchOverlayOpen, setResearchOverlayOpen] = useState(false)
  const [acknowledgedResolutionIDs, setAcknowledgedResolutionIDs] = useState<Set<string>>(() => new Set())
  const [persistenceBusy, setPersistenceBusy] = useState(false)
  const [pendingRestore, setPendingRestore] = useState<PendingRestore | null>(null)
  const [lastNotification, setLastNotification] = useState<Notification | null>(null)
  const reconnectTimer = useRef<number | null>(null)
  const loadFileInputRef = useRef<HTMLInputElement | null>(null)
  const socketConnectedRef = useRef(false)
  const snapshotRef = useRef<PlayerSnapshot | null>(null)
  const draftOrdersRef = useRef<DraftOrder[]>([])
  const planningDraftRevisionRef = useRef(0)
  const activeGameRef = useRef(gameID)
  const activeSeatRef = useRef(seatID)
  activeGameRef.current = gameID
  activeSeatRef.current = seatID
  snapshotRef.current = snapshot
  draftOrdersRef.current = draftOrders

  const difficultyProfilesByID = useMemo(() => new Map((difficultyCatalog?.profiles ?? []).map((profile) => [profile.id, profile] as const)), [difficultyCatalog])
  const galaxySizeProfilesByID = useMemo(() => new Map((galaxyCatalog?.sizes ?? []).map((profile) => [profile.id, profile] as const)), [galaxyCatalog])
  const galaxyAgeProfilesByID = useMemo(() => new Map((galaxyCatalog?.ages ?? []).map((profile) => [profile.id, profile] as const)), [galaxyCatalog])
  const raceProfilesByID = useMemo(() => new Map((raceCatalog?.profiles ?? []).map((profile) => [profile.id, profile] as const)), [raceCatalog])
  const selectedPlayerRaceProfile = raceProfilesByID.get(playerRaceID)
  const technologyProfilesByID = useMemo(() => new Map((technologyCatalog?.profiles ?? []).map((profile) => [profile.id, profile] as const)), [technologyCatalog])
  const opponentCountProfilesByID = useMemo(() => new Map((compositionCatalog?.counts ?? []).map((profile) => [String(profile.opponent_count), profile] as const)), [compositionCatalog])
  const opponentGalaxyLimitsByID = useMemo(() => new Map((compositionCatalog?.galaxy_limits ?? []).map((limit) => [limit.galaxy_size, limit] as const)), [compositionCatalog])
  const selectedOpponentCount = Number(opponentCountID)
  const selectedOpponentProfile = opponentCountProfilesByID.get(opponentCountID)
  const selectedOpponentAssignment = useMemo<NewGameOpponentAssignment | undefined>(() => compositionCatalog?.assignments.find((assignment) => assignment.player_race_id === playerRaceID && assignment.opponent_count === selectedOpponentCount), [compositionCatalog, playerRaceID, selectedOpponentCount])
  const selectedOpponentLimit = opponentGalaxyLimitsByID.get(galaxySizeID)?.max_supported_opponents ?? 0
  const compositionSupported = selectedOpponentProfile?.availability === 'supported' && selectedOpponentCount <= selectedOpponentLimit && Boolean(selectedOpponentAssignment)

  useEffect(() => {
    const controller = new AbortController()
    setDifficultyCatalogError('')
    void getDifficultyCatalog(controller.signal)
      .then((catalog) => {
        setDifficultyCatalog(catalog)
        setDifficultyCatalogError('')
        setDifficultyID((current) => catalog.profiles.some((profile) => profile.id === current) ? current : catalog.default_id)
      })
      .catch((reason) => {
        if (controller.signal.aborted) return
        setDifficultyCatalog(null)
        setDifficultyCatalogError(errorText(reason))
      })
    return () => controller.abort()
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    setGalaxyCatalogError('')
    void getGalaxyCatalog(controller.signal)
      .then((catalog) => {
        setGalaxyCatalog(catalog)
        setGalaxyCatalogError('')
        setGalaxySizeID((current) => catalog.sizes.some((profile) => profile.id === current) ? current : catalog.default_size_id)
        setGalaxyAgeID((current) => catalog.ages.some((profile) => profile.id === current) ? current : catalog.default_age_id)
      })
      .catch((reason) => {
        if (controller.signal.aborted) return
        setGalaxyCatalog(null)
        setGalaxyCatalogError(errorText(reason))
      })
    return () => controller.abort()
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    setRaceCatalogError('')
    void getRaceCatalog(controller.signal)
      .then((catalog) => {
        setRaceCatalog(catalog)
        setRaceCatalogError('')
        setPlayerRaceID((current) => catalog.profiles.some((profile) => profile.id === current) ? current : catalog.default_player_race_id)
      })
      .catch((reason) => {
        if (controller.signal.aborted) return
        setRaceCatalog(null)
        setRaceCatalogError(errorText(reason))
      })
    return () => controller.abort()
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    setTechnologyCatalogError('')
    void getTechnologyCatalog(controller.signal)
      .then((catalog) => {
        setTechnologyCatalog(catalog)
        setTechnologyCatalogError('')
        setTechnologyLevelID((current) => catalog.profiles.some((profile) => profile.id === current) ? current : catalog.default_id)
      })
      .catch((reason) => {
        if (controller.signal.aborted) return
        setTechnologyCatalog(null)
        setTechnologyCatalogError(errorText(reason))
      })
    return () => controller.abort()
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    setCompositionCatalogError('')
    void getCompositionCatalog(controller.signal)
      .then((catalog) => {
        setCompositionCatalog(catalog)
        setCompositionCatalogError('')
        setOpponentCountID((current) => catalog.counts.some((profile) => String(profile.opponent_count) === current) ? current : String(catalog.default_opponent_count) as OpponentCountID)
      })
      .catch((reason) => {
        if (controller.signal.aborted) return
        setCompositionCatalog(null)
        setCompositionCatalogError(errorText(reason))
      })
    return () => controller.abort()
  }, [])
  useEffect(() => {
    const syncRoute = () => setRoute(parseRoute())
    window.addEventListener('hashchange', syncRoute)
    if (!window.location.hash) navigate({ kind: 'home' })
    return () => window.removeEventListener('hashchange', syncRoute)
  }, [])
  useEffect(() => {
    setAcknowledgedResolutionIDs(readResolutionAcks(gameID, seatID))
  }, [gameID, seatID])


  useEffect(() => {
    if (route.kind === 'game' && route.gameID !== gameID) {
      setSnapshot(null)
      draftOrdersRef.current = []
      planningDraftRevisionRef.current = 0
      setDraftOrders([])
      setLifecycle('initial-loading')
      setError('')
      setGameID(route.gameID)
    }
  }, [route, gameID])

  const loadSnapshot = useCallback(async (selectedGameID = gameID, selectedSeatID = seatID, signal?: AbortSignal) => {
    if (!selectedGameID || selectedSeatID <= 0) return
    setLifecycle((current) => current === 'synced' || current === 'refresh-failed' ? 'refreshing' : current)
    try {
      const next = await getPlayerSnapshot(selectedGameID, selectedSeatID, signal)
      if (selectedGameID !== activeGameRef.current || selectedSeatID !== activeSeatRef.current) return
      const previous = snapshotRef.current
      const samePlanningAuthority = Boolean(previous
        && previous.view.game_id === next.view.game_id
        && previous.view.turn === next.view.turn
        && previous.view.revision === next.view.revision)
      const serverDraft = next.planning_draft
      let hydratedDraft: DraftOrder[] = []
      if (next.view.phase === 'planning' && !next.view.seat.submitted) {
        const serverDraftMatches = Boolean(serverDraft
          && serverDraft.game_id === next.view.game_id
          && serverDraft.seat_id === selectedSeatID
          && serverDraft.turn === next.view.turn
          && serverDraft.base_revision === next.view.revision)
        if (serverDraftMatches && serverDraft) {
          if (!samePlanningAuthority || serverDraft.draft_revision >= planningDraftRevisionRef.current) {
            planningDraftRevisionRef.current = serverDraft.draft_revision
            hydratedDraft = serverDraft.orders.map((order) => ({ key: order.key, kind: order.kind, payload: order.payload }))
          } else {
            hydratedDraft = [...draftOrdersRef.current]
          }
        } else if (samePlanningAuthority && planningDraftRevisionRef.current > 0) {
          hydratedDraft = [...draftOrdersRef.current]
        } else {
          planningDraftRevisionRef.current = 0
        }
      } else {
        planningDraftRevisionRef.current = 0
      }
      draftOrdersRef.current = hydratedDraft
      setSnapshot(next)
      setDraftOrders(hydratedDraft)
      setPlanningPreview(null)
      setPlanningPreviewError('')
      const firstColony = next.view.colonies[0]
      if (firstColony) {
        const population = aggregatePopulation(firstColony)
        setAssignment({ farmers: String(population.farmers), workers: String(population.workers), scientists: String(population.scientists) })
      }
      setLifecycle(socketConnectedRef.current ? 'synced' : 'connecting')
      setStatus({ key: 'status.snapshotSynced', vars: { change: next.change_sequence } })
      setError('')
      return next
    } catch (reason) {
      if (!signal?.aborted && selectedGameID === activeGameRef.current && selectedSeatID === activeSeatRef.current) {
        setLifecycle(snapshotRef.current ? 'refresh-failed' : 'fatal')
        setError(errorText(reason))
      }
      throw reason
    }
  }, [gameID, seatID])

  const refreshAfterConflict = useCallback(async (reason: unknown, preservePlanningDraft = false) => {
    if (!isAPIError(reason) || reason.status !== 409) return false
    const current = snapshotRef.current
    if (!current) return false
    const preservedDraft = preservePlanningDraft ? [...draftOrdersRef.current] : []
    try {
      const next = await loadSnapshot(current.view.game_id, activeSeatRef.current)
      if (!next) return true
      if (preservedDraft.length > 0 && next.view.phase === 'planning') {
        setDraftOrders(preservedDraft)
        setStatus({ key: 'status.conflictRefreshedDraftPreserved', vars: { change: next.change_sequence, count: preservedDraft.length } })
      } else if (preservedDraft.length > 0) {
        setStatus({ key: 'status.conflictRefreshedDraftDropped', vars: { change: next.change_sequence, phase: next.view.phase } })
      } else {
        setStatus({ key: 'status.conflictRefreshed', vars: { change: next.change_sequence } })
      }
      // Keep the authoritative rejection visible after the successful refetch.
      // The rejected mutation is never automatically resubmitted.
      setError(errorText(reason))
    } catch (refreshReason) {
      setError(errorText(reason) + ' · ' + errorText(refreshReason))
    }
    return true
  }, [loadSnapshot])

  useEffect(() => {
    const controller = new AbortController()
    void listGames(controller.signal)
      .then((available) => {
        setGames(available)
        if (available.length === 0) {
          setStatus({ key: 'status.noGames' })
          return
        }
        const routeGame = parseRoute()
        const selected = routeGame.kind === 'game' && available.some((game) => game.game_id === routeGame.gameID)
          ? routeGame.gameID
          : available[0].game_id
        setGameID(selected)
        setStatus({ key: 'status.connectedGame', vars: { game: selected } })
      })
      .catch((reason: unknown) => {
        if (!controller.signal.aborted) {
          setError(errorText(reason))
          setStatus({ key: 'status.discoverFailed' })
        }
      })
    return () => controller.abort()
  }, [])

  useEffect(() => {
    if (!gameID || seatID <= 0) return
    const controller = new AbortController()
    void loadSnapshot(gameID, seatID, controller.signal).catch((reason: unknown) => {
      if (!controller.signal.aborted) setError(errorText(reason))
    })
    return () => controller.abort()
  }, [gameID, seatID, loadSnapshot])

  useEffect(() => {
    if (!gameID) return
    let disposed = false
    let socket: WebSocket | null = null

    const connect = () => {
      if (disposed) return
      socketConnectedRef.current = false
      setLifecycle(snapshotRef.current ? 'reconnecting' : 'connecting')
      void loadSnapshot(gameID, seatID).catch(() => undefined)
      socket = new WebSocket(streamURL(gameID))
      socket.onopen = () => {
        socketConnectedRef.current = true
        if (snapshotRef.current) setLifecycle('synced')
        setStatus({ key: 'status.liveConnected' })
      }
      socket.onmessage = (event) => {
        try {
          const notification = JSON.parse(String(event.data)) as Notification
          setLastNotification(notification)
          setStatus({ key: 'status.invalidated', vars: { reason: notification.reason } })
          setSnapshot((current) => {
            if (!current || notification.change_sequence > current.change_sequence) {
              setLifecycle('refreshing')
              void loadSnapshot(gameID, seatID).catch((reason: unknown) => {
                setError(errorText(reason))
              })
            }
            return current
          })
        } catch (reason) {
          setError(errorText(reason))
        }
      }
      socket.onclose = () => {
        socketConnectedRef.current = false
        if (!disposed) {
          setLifecycle(snapshotRef.current ? 'reconnecting' : 'connecting')
          setStatus({ key: 'status.liveDisconnected' })
          reconnectTimer.current = window.setTimeout(connect, 1000)
        }
      }
      socket.onerror = () => socket?.close()
    }

    connect()
    return () => {
      disposed = true
      if (reconnectTimer.current !== null) {
        window.clearTimeout(reconnectTimer.current)
        reconnectTimer.current = null
      }
      socketConnectedRef.current = false
      socket?.close(1000, 'component disposed')
    }
  }, [gameID, seatID, loadSnapshot])

  useEffect(() => {
    if (!snapshot || snapshot.view.phase !== 'planning' || snapshot.view.seat.submitted) {
      setPlanningPreview(null)
      setPlanningPreviewError('')
      setPreviewBusy(false)
      return
    }
    const controller = new AbortController()
    const previewGameID = snapshot.view.game_id
    const previewTurn = snapshot.view.turn
    const previewRevision = snapshot.view.revision
    setPreviewBusy(true)
    void previewPlanning(snapshot, seatID, draftOrders, controller.signal)
      .then((preview) => {
        if (!controller.signal.aborted) {
          setPlanningPreview(preview)
          setPlanningPreviewError('')
        }
      })
      .catch((cause) => {
        if (controller.signal.aborted) return
        const current = snapshotRef.current
        const snapshotMoved = !current
          || current.view.game_id !== previewGameID
          || current.view.turn !== previewTurn
          || current.view.revision !== previewRevision
          || current.view.phase !== 'planning'
          || current.view.seat.submitted
        if (snapshotMoved || isObsoletePlanningPreviewConflict(cause)) {
          setPlanningPreviewError('')
          return
        }
        setPlanningPreviewError(errorText(cause))
      })
      .finally(() => {
        if (!controller.signal.aborted) setPreviewBusy(false)
      })
    return () => controller.abort()
  }, [snapshot, seatID, draftOrders])

  const firstColony = snapshot?.view.colonies[0]
  const totalDraft = useMemo(
    () => Number(assignment.farmers || 0) + Number(assignment.workers || 0) + Number(assignment.scientists || 0),
    [assignment],
  )
  const mutationLocked = lifecycle !== 'synced'
  const invasionDecision = snapshot?.view.invasion
  const invasionSystem = invasionDecision ? snapshot?.decision?.strategic.galaxy.systems.find((system) => system.id === invasionDecision.system_id) : undefined
  const invasionColonyContact = invasionDecision ? snapshot?.decision?.strategic.contacts?.find((contact) => contact.kind === 'colony' && contact.colony_id === invasionDecision.colony_id) : undefined
  const invasionPlanet = invasionSystem && invasionColonyContact?.planet_id ? invasionSystem.planets.find((planet) => planet.id === invasionColonyContact.planet_id) : undefined
  const invasionDefender = invasionDecision ? snapshot?.decision?.public_empires?.find((empire) => empire.id === invasionDecision.defender_empire_id) : undefined
  const invasionTransportFleets = invasionDecision ? snapshot?.decision?.strategic.fleets?.filter((fleet) => invasionDecision.eligible_transport_fleet_ids.includes(fleet.id)) ?? [] : []
  const colonyBaseDecision = snapshot?.decision?.decisions.colony_base?.[0]
  const colonyBaseSourceColony = colonyBaseDecision ? snapshot?.decision?.colonies.find((colony) => colony.id === colonyBaseDecision.source_colony_id) : undefined
  const colonyBaseSystem = colonyBaseDecision ? snapshot?.decision?.strategic.galaxy.systems.find((system) => system.id === colonyBaseDecision.system_id) : undefined
  const colonyBaseSourcePlanet = colonyBaseSourceColony ? snapshot?.decision?.strategic.galaxy.systems.flatMap((system) => system.planets).find((planet) => planet.id === colonyBaseSourceColony.planet_id) : undefined
  const colonyBaseTargets = colonyBaseDecision && colonyBaseSystem ? colonyBaseSystem.planets.filter((planet) => colonyBaseDecision.target_planet_ids.includes(planet.id)) : []
  const diplomacyClosed = mutationLocked || snapshot?.view.phase !== 'planning' || Boolean(snapshot?.view.seats.some((seat) => seat.submitted))
  const projectedResources = planningPreview?.preview.projection
  const projectedEmpire = snapshot?.decision?.empire
  const projectedColonies = projectedResources?.colonies?.map((item) => item.colony) ?? snapshot?.decision?.colonies ?? snapshot?.view.colonies ?? []
  const foodProduced = projectedColonies.reduce((sum, colony) => sum + colony.adjusted_economy.food, 0)
  const foodRequired = projectedColonies.reduce((sum, colony) => sum + colony.population_dynamics.food_required, 0)
  const foodSurplus = projectedColonies.reduce((sum, colony) => sum + colony.population_dynamics.food_surplus, 0)
  const foodShortage = projectedColonies.reduce((sum, colony) => sum + colony.population_dynamics.food_shortage, 0)
  const foodImported = projectedColonies.reduce((sum, colony) => sum + colony.population_dynamics.food_imported, 0)
  const foodExported = projectedColonies.reduce((sum, colony) => sum + colony.population_dynamics.food_exported, 0)
  const foodNet = foodSurplus - foodShortage
  const projectedTaxIncome = projectedColonies.reduce((sum, colony) => sum + colony.adjusted_economy.tax_bc, 0)

  const treasuryBalance = projectedResources?.treasury_balance_bc ?? projectedEmpire?.treasury.balance_bc ?? 0
  const treasuryNet = projectedResources?.net_modeled_income_bc ?? projectedEmpire?.treasury.net_modeled_income_bc ?? 0
  const commandCapacity = projectedResources?.command_points.capacity ?? projectedEmpire?.command_points.capacity ?? 0
  const commandUsed = projectedResources?.command_points.used ?? projectedEmpire?.command_points.used ?? 0
  const commandAvailable = commandCapacity - commandUsed
  const commandOverage = projectedResources?.command_point_overage ?? Math.max(0, commandUsed - commandCapacity)
  const freighterTotal = projectedResources?.freighters.total ?? projectedEmpire?.freighters ?? 0
  const freighterFoodUsed = projectedResources?.freighters.food_used ?? projectedEmpire?.food_logistics.freighters_used ?? 0
  const freighterTransferReserved = projectedResources?.freighters.transfer_reserved ?? projectedEmpire?.food_logistics.population_transport_freighters_reserved ?? 0
  const freighterAvailable = projectedResources?.freighters.available ?? Math.max(0, freighterTotal - freighterFoodUsed - freighterTransferReserved)

  const researchProjection = projectedResources?.research
  const researchState = projectedEmpire?.research
  const researchFieldID = researchProjection?.tech_field_id ?? researchState?.tech_field_id
  const researchChoice = snapshot?.decision?.decisions.research?.find((choice) => choice.tech_field_id === researchFieldID)
  const researchProgress = researchProjection?.progress_rp ?? researchState?.progress_rp ?? 0
  const researchCost = researchProjection?.cost_rp ?? researchChoice?.base_cost_rp ?? 0
  const researchRemaining = researchProjection?.remaining_rp ?? Math.max(0, researchCost - researchProgress)
  const researchRate = researchProjection?.rp_per_turn ?? projectedColonies.reduce((sum, colony) => sum + colony.adjusted_economy.research, 0)
  const researchETA = researchProjection?.eta_turns
  const researchPercent = researchCost > 0 ? Math.max(0, Math.min(100, (researchProgress / researchCost) * 100)) : 0
  const hasActiveResearch = researchFieldID !== undefined && researchCost > 0
  const researchNearBreakthrough = hasActiveResearch && (researchETA !== undefined && researchETA <= 1
    || Boolean(researchProjection && researchProjection.rp_per_turn > 0 && researchProjection.remaining_rp <= researchProjection.rp_per_turn))

  function signed(value: number, digits = 0) {
    const rounded = value.toFixed(digits)
    return `${value > 0 ? '+' : ''}${rounded}`
  }

  function signedTone(value: number): 'positive' | 'danger' | 'neutral' {
    if (value > 0) return 'positive'
    if (value < 0) return 'danger'
    return 'neutral'
  }

  const resourceChips: ResourceChip[] = projectedEmpire ? [
    {
      id: 'bc',
      label: t('resource.bc'),
      shortLabel: 'BC',
      icon: 'credits',
      value: Math.round(treasuryBalance).toString(),
      delta: `[${signed(treasuryNet)}]`,
      deltaTone: signedTone(treasuryNet),
      detailTitle: t('resourceDetail.bcTitle'),
      details: [
        { label: t('resourceDetail.balance'), value: Math.round(treasuryBalance).toString() },
        { label: t('resourceDetail.netPerTurn'), value: signed(treasuryNet), tone: signedTone(treasuryNet) },
        { label: t('resourceDetail.taxIncome'), value: projectedTaxIncome.toFixed(1) },
        { label: t('resourceDetail.foodIncome'), value: projectedEmpire.treasury.surplus_food_income_bc.toFixed(1) },
        { label: t('resourceDetail.buildingMaintenance'), value: projectedEmpire.treasury.building_maintenance_bc.toFixed(1) },
        { label: t('resourceDetail.freighterCost'), value: projectedEmpire.treasury.freighter_operating_cost_bc.toFixed(1) },
        { label: t('resourceDetail.commandCost'), value: projectedEmpire.treasury.ship_command_maintenance_bc.toFixed(1) },
      ],
    },
    {
      id: 'food',
      label: t('resource.food'),
      shortLabel: 'Food',
      icon: 'food',
      delta: `[${signed(foodNet, 1)}]`,
      deltaTone: signedTone(foodNet),
      tone: foodNet < 0 ? 'danger' as const : foodNet > 0 ? 'positive' as const : 'neutral' as const,
      detailTitle: t('resourceDetail.foodTitle'),
      details: [
        { label: t('resourceDetail.foodNet'), value: signed(foodNet, 1), tone: signedTone(foodNet) },
        { label: t('resourceDetail.foodProduced'), value: foodProduced.toFixed(1) },
        { label: t('resourceDetail.foodRequired'), value: foodRequired.toFixed(1) },
        { label: t('resourceDetail.foodSurplus'), value: foodSurplus.toFixed(1), tone: foodSurplus > 0 ? 'positive' as const : 'neutral' as const },
        { label: t('resourceDetail.foodShortage'), value: foodShortage.toFixed(1), tone: foodShortage > 0 ? 'danger' as const : 'neutral' as const },
        { label: t('resourceDetail.foodImported'), value: foodImported.toFixed(1) },
        { label: t('resourceDetail.foodExported'), value: foodExported.toFixed(1) },
      ],
    },
    {
      id: 'freighters',
      label: t('resource.freighters'),
      shortLabel: 'Tr',
      icon: 'freighter',
      value: `${freighterAvailable}/${freighterTotal}`,
      tone: freighterAvailable <= 0 && freighterTotal > 0 ? 'warning' as const : 'neutral' as const,
      detailTitle: t('resourceDetail.freighterTitle'),
      details: [
        { label: t('resourceDetail.freighterTotal'), value: freighterTotal.toString() },
        { label: t('resourceDetail.freighterAvailable'), value: freighterAvailable.toString(), tone: freighterAvailable > 0 ? 'positive' as const : 'neutral' as const },
        { label: t('resourceDetail.freighterFood'), value: freighterFoodUsed.toString() },
        { label: t('resourceDetail.freighterTransfer'), value: freighterTransferReserved.toString() },
        { label: t('resourceDetail.freighterCost'), value: projectedEmpire.treasury.freighter_operating_cost_bc.toFixed(1) },
      ],
    },
    {
      id: 'command',
      label: t('resource.command'),
      shortLabel: 'CP',
      icon: 'command',
      value: commandCapacity.toString(),
      delta: `[${signed(commandAvailable)}]`,
      deltaTone: signedTone(commandAvailable),
      tone: commandOverage > 0 ? 'danger' as const : commandAvailable === 0 ? 'warning' as const : 'neutral' as const,
      detailTitle: t('resourceDetail.commandTitle'),
      details: [
        { label: t('resourceDetail.commandCapacity'), value: commandCapacity.toString() },
        { label: t('resourceDetail.commandUsed'), value: commandUsed.toString() },
        { label: t('resourceDetail.commandAvailable'), value: signed(commandAvailable), tone: signedTone(commandAvailable) },
        { label: t('resourceDetail.commandOverage'), value: commandOverage.toString(), tone: commandOverage > 0 ? 'danger' as const : 'neutral' as const },
        { label: t('resourceDetail.commandCost'), value: projectedEmpire.treasury.ship_command_maintenance_bc.toFixed(1) },
      ],
    },
    {
      id: 'research',
      label: t('resource.research'),
      shortLabel: 'RP',
      icon: 'research',
      value: `${researchRate.toFixed(1)} RP`,
      delta: hasActiveResearch ? `[${researchPercent.toFixed(0)}%${researchETA !== undefined ? ` · ${researchETA}T` : ''}]` : '[—]',
      deltaTone: researchNearBreakthrough ? 'warning' as const : 'neutral' as const,
      tone: researchNearBreakthrough ? 'warning' as const : 'neutral' as const,
      progressPercent: researchPercent,
      detailTitle: t('resourceDetail.researchTitle'),
      details: [
        { label: t('resourceDetail.researchRate'), value: `${researchRate.toFixed(1)} RP` },
        { label: t('resourceDetail.researchProgress'), value: researchCost > 0 ? `${researchProgress.toFixed(1)} / ${researchCost.toFixed(1)} RP` : researchProgress.toFixed(1) },
        { label: t('resourceDetail.researchPercent'), value: `${researchPercent.toFixed(1)}%` },
        { label: t('resourceDetail.researchRemaining'), value: `${researchRemaining.toFixed(1)} RP` },
        { label: t('resourceDetail.researchEta'), value: researchETA !== undefined ? `${researchETA}` : '—', tone: researchNearBreakthrough ? 'warning' as const : 'neutral' as const },
        { label: t('resourceDetail.researchField'), value: researchFieldID !== undefined ? `#${researchFieldID}${researchChoice?.category_id ? ` · ${researchChoice.category_id}` : ''}` : '—' },
        { label: t('resourceDetail.researchMode'), value: researchProjection?.selection_mode ?? researchState?.selection_mode ?? '—' },
        { label: t('resourceDetail.breakthrough'), value: !hasActiveResearch ? t('resourceDetail.researchNone') : researchNearBreakthrough ? t('resourceDetail.breakthroughNear') : t('resourceDetail.breakthroughNormal'), tone: researchNearBreakthrough ? 'warning' as const : 'neutral' as const },
      ],
    },
  ] : []
  const eventStatusText = t(status.key, status.vars)
  const lifecycleOwnsStatus = route.kind === 'game' || persistenceBusy || pendingRestore !== null
  const statusText = lifecycleOwnsStatus && lifecycle !== 'synced' ? t(lifecycleKey(lifecycle)) : eventStatusText
  const statusTone: 'neutral' | 'success' | 'warning' | 'danger' = error && (lifecycle === 'fatal' || lifecycle === 'refresh-failed')
    ? 'danger'
    : lifecycle === 'reconnecting' || lifecycle === 'refresh-failed'
      ? 'warning'
      : lifecycle === 'synced'
        ? 'success'
        : 'neutral'

  async function submitNewGame(event: FormEvent) {
    event.preventDefault()
    setCreatingGame(true)
    setError('')
    try {
      if (!compositionSupported || !selectedOpponentAssignment) throw new Error(t('newGame.compositionUnavailable'))
      const composedPlayers = [
        { seat_id: 1, empire_name: playerName, race_id: playerRaceID },
        ...selectedOpponentAssignment.opponents.map((opponent, index) => ({ seat_id: index + 2, empire_name: opponent.default_empire_name, race_id: opponent.race_id })),
      ]
      const composedControllers = [
        { seat_id: 1, controller: 'local_human' as const },
        ...selectedOpponentAssignment.opponents.map((_, index) => ({ seat_id: index + 2, controller: 'builtin_ai' as const })),
      ]
      const created = await createGame({
        schema_version: 1,
        game_id: '',
        seed: newGameSeed,
        controllers: composedControllers,
        settings: {
          difficulty_id: difficultyID,
          galaxy_size: galaxySizeID,
          galaxy_age: galaxyAgeID,
          technology_level: technologyLevelID,
          strategic_combat: false,
          players: composedPlayers,
        },
      })
      setGames(await listGames())
      setSnapshot(null)
      setGameID(created.game.game_id)
      setSeatID(created.players[0]?.seat_id ?? 1)
      setStatus({ key: 'status.createdGame', vars: { game: created.game.game_id, seed: newGameSeed } })
      navigate({ kind: 'game', gameID: created.game.game_id, section: 'galaxy' })
    } catch (reason) {
      setError(errorText(reason))
    } finally {
      setCreatingGame(false)
    }
  }

  function persistPlanningDraft(nextOrders: DraftOrder[]) {
    const current = snapshotRef.current
    const currentSeat = activeSeatRef.current
    if (!current || current.view.phase !== 'planning' || current.view.seat.submitted) return
    const draftRevision = planningDraftRevisionRef.current + 1
    planningDraftRevisionRef.current = draftRevision
    void savePlanningDraft(current, currentSeat, nextOrders, draftRevision)
      .then((saved) => {
        if (saved.draft.draft_revision > planningDraftRevisionRef.current) planningDraftRevisionRef.current = saved.draft.draft_revision
      })
      .catch((cause) => {
        const latest = snapshotRef.current
        const obsolete = !latest
          || latest.view.game_id !== current.view.game_id
          || latest.view.turn !== current.view.turn
          || latest.view.revision !== current.view.revision
          || latest.view.phase !== 'planning'
        if (!obsolete) setPlanningPreviewError(errorText(cause))
      })
  }

  function replacePlanningDraft(nextOrders: DraftOrder[]) {
    draftOrdersRef.current = nextOrders
    setDraftOrders(nextOrders)
    persistPlanningDraft(nextOrders)
  }

  function planOrder(order: DraftOrder) {
    replacePlanningDraft(replaceDraftOrder(draftOrdersRef.current, order))
  }

  function removePlannedOrder(key: string) {
    replacePlanningDraft(removeDraftOrder(draftOrdersRef.current, key))
  }

  function planPopulation(colonyID: number, farmers: number, workers: number, scientists: number) {
    planOrder({
      key: `population:${colonyID}`,
      kind: 'colony.assign_population',
      payload: { colony_id: colonyID, farmers, workers, scientists },
    })
  }

  function submitAssignment(event: FormEvent) {
    event.preventDefault()
    if (!firstColony) return
    planPopulation(firstColony.id, Number(assignment.farmers), Number(assignment.workers), Number(assignment.scientists))
  }

  function restoreConnectionLifecycle() {
    setLifecycle(snapshotRef.current ? (socketConnectedRef.current ? 'synced' : 'reconnecting') : 'connecting')
  }

  function parseSaveMetadata(fileText: string): SaveFileMetadata {
    const raw = JSON.parse(fileText) as { schema_version?: unknown; game_id?: unknown; revision?: unknown; phase?: unknown; state?: { turn?: unknown } }
    if (typeof raw.game_id !== 'string' || raw.game_id.trim() === '') throw new Error(t('persistence.invalidMetadata'))
    return {
      gameID: raw.game_id,
      turn: typeof raw.state?.turn === 'number' ? raw.state.turn : undefined,
      revision: typeof raw.revision === 'number' ? raw.revision : undefined,
      phase: typeof raw.phase === 'string' ? raw.phase : undefined,
      schemaVersion: typeof raw.schema_version === 'number' ? raw.schema_version : undefined,
    }
  }

  async function saveGame() {
    if (!gameID || !snapshot || persistenceBusy || lifecycle !== 'synced') return
    setPersistenceBusy(true)
    setLifecycle('exporting')
    setError('')
    try {
      const save = await exportLiveSnapshot(gameID)
      const objectURL = URL.createObjectURL(save)
      const link = document.createElement('a')
      const safeGameID = gameID.replace(/[^a-zA-Z0-9._-]+/g, '-')
      link.href = objectURL
      link.download = `moox-${safeGameID}-turn-${snapshot.view.turn}.json`
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.setTimeout(() => URL.revokeObjectURL(objectURL), 0)
      setStatus({ key: 'status.saveExported', vars: { game: gameID, turn: snapshot.view.turn } })
    } catch (reason) {
      setError(errorText(reason))
    } finally {
      setPersistenceBusy(false)
      restoreConnectionLifecycle()
    }
  }

  function chooseLoadGame() {
    if (persistenceBusy) return
    loadFileInputRef.current?.click()
  }

  async function loadSelectedFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.currentTarget.files?.[0]
    event.currentTarget.value = ''
    if (!file || persistenceBusy) return
    setPersistenceBusy(true)
    setPendingRestore(null)
    setLifecycle('loading-file')
    setError('')
    try {
      const metadata = parseSaveMetadata(await file.text())
      const hosted = await listGames()
      setGames(hosted)
      if (hosted.some((item) => item.game_id === metadata.gameID)) {
        setPendingRestore({ file, metadata })
        return
      }
      setLifecycle('importing')
      const imported = await importLiveSnapshot(file)
      setStatus({ key: 'status.gameImported', vars: { game: imported.game_id } })
      setPersistenceBusy(false)
      enterGame(imported.game_id)
    } catch (reason) {
      setError(errorText(reason))
      restoreConnectionLifecycle()
    } finally {
      setPersistenceBusy(false)
    }
  }

  function cancelRestore() {
    setPendingRestore(null)
    restoreConnectionLifecycle()
  }

  async function confirmRestore() {
    if (!pendingRestore || persistenceBusy) return
    const { file, metadata } = pendingRestore
    setPersistenceBusy(true)
    setLifecycle('restoring')
    setError('')
    try {
      const receipt = await restoreLiveSnapshot(metadata.gameID, file)
      setPendingRestore(null)
      setStatus({ key: 'status.gameRestored', vars: { game: metadata.gameID, revision: receipt.game_revision } })
      if (metadata.gameID === gameID) {
        await loadSnapshot(metadata.gameID, seatID)
        setLifecycle('loaded')
        window.setTimeout(() => restoreConnectionLifecycle(), 650)
      } else {
        enterGame(metadata.gameID)
      }
    } catch (reason) {
      setError(errorText(reason))
      restoreConnectionLifecycle()
    } finally {
      setPersistenceBusy(false)
    }
  }

  async function endTurn() {
    if (!snapshot || mutationLocked || planningBusy || snapshot.view.phase !== 'planning') return
    setPlanningBusy(true)
    setError('')
    try {
      const receipt = await submitPlanning(snapshot, seatID, draftOrders)
      draftOrdersRef.current = []
      planningDraftRevisionRef.current = 0
      setDraftOrders([])
      setPlanningPreview(null)
      setStatus({ key: 'status.commandAccepted', vars: { change: receipt.change_sequence, revision: receipt.game_revision } })
      await loadSnapshot(snapshot.view.game_id, seatID)
    } catch (cause) {
      setError(errorText(cause))
      await refreshAfterConflict(cause, true)
    } finally {
      setPlanningBusy(false)
    }
  }
  async function runDiplomacy(kind: DiplomacyCommandKind, otherEmpireID: number) {
    if (!snapshot || mutationLocked) return
    setDiplomacyBusy(true)
    setError('')
    try {
      const receipt = await submitDiplomacy(snapshot, seatID, kind, otherEmpireID)
      setStatus({ key: 'status.diplomacyAccepted', vars: { change: receipt.change_sequence, revision: receipt.game_revision } })
      await loadSnapshot(snapshot.view.game_id, seatID)
    } catch (reason) {
      setError(errorText(reason))
      await refreshAfterConflict(reason)
    } finally {
      setDiplomacyBusy(false)
    }
  }

  async function runColonyBase(action: 'colonize' | 'trash', planetID?: number) {
    if (!snapshot || !colonyBaseDecision || mutationLocked || colonyBaseBusy) return
    setColonyBaseBusy(true)
    setError('')
    try {
      const receipt = await submitColonyBase(snapshot, seatID, colonyBaseDecision, action, planetID)
      setStatus({ key: action === 'colonize' ? 'status.colonyBaseColonized' : 'status.colonyBaseTrashed', vars: { change: receipt.change_sequence, revision: receipt.game_revision } })
      await loadSnapshot(snapshot.view.game_id, seatID)
    } catch (reason) {
      setError(errorText(reason))
      await refreshAfterConflict(reason)
    } finally {
      setColonyBaseBusy(false)
    }
  }

  async function runInvasion(action: 'invade' | 'decline') {
    if (!snapshot?.view.invasion || mutationLocked) return
    setInvasionBusy(true)
    setError('')
    try {
      const receipt = await submitInvasion(snapshot, seatID, action)
      setStatus({ key: 'status.invasionAccepted', vars: { change: receipt.change_sequence, revision: receipt.game_revision } })
      await loadSnapshot()
    } catch (cause) {
      setError(errorText(cause))
      await refreshAfterConflict(cause)
    } finally {
      setInvasionBusy(false)
    }
  }

  async function runBattleCommand(battleID: number, command: ProtocolCommand) {
    if (!snapshot || mutationLocked) return
    setError('')
    try {
      await submitBattleCommand(snapshot.view.game_id, battleID, seatID, command)
      await loadSnapshot(snapshot.view.game_id, seatID)
    } catch (cause) {
      setError(errorText(cause))
      await refreshAfterConflict(cause)
      throw cause
    }
  }

  function enterGame(selectedGameID: string, section: GameSection = 'galaxy') {
    const reloadSelectedGame = selectedGameID === gameID
    setGameID(selectedGameID)
    setSnapshot(null)
    setLifecycle('initial-loading')
    navigate({ kind: 'game', gameID: selectedGameID, section })
    if (reloadSelectedGame) {
      void loadSnapshot(selectedGameID, seatID).catch((cause: unknown) => {
        setError(errorText(cause))
      })
    }
  }

  function selectHostedGame(selectedGameID: string) {
    enterGame(selectedGameID, route.kind === 'game' ? route.section : 'galaxy')
  }

  const persistenceControls = (
    <>
      <input ref={loadFileInputRef} className="persistence-file-input" type="file" accept="application/json,.json" onChange={(event) => void loadSelectedFile(event)} aria-hidden="true" tabIndex={-1} />
      {pendingRestore && (
        <div className="persistence-dialog-backdrop" role="presentation">
          <Card className="persistence-dialog" as="section">
            <p className="eyebrow">{t('persistence.loadEyebrow')}</p>
            <h2>{t('persistence.restoreTitle')}</h2>
            <p>{t('persistence.restoreDetails', { game: pendingRestore.metadata.gameID, turn: pendingRestore.metadata.turn ?? '?' })}</p>
            <p className="muted">{t('persistence.restoreWarning')}</p>
            <div className="action-row">
              <button type="button" className="button-primary" disabled={persistenceBusy} onClick={() => void confirmRestore()}>{t('persistence.restoreConfirm')}</button>
              <button type="button" className="button-secondary" disabled={persistenceBusy} onClick={cancelRestore}>{t('common.cancel')}</button>
            </div>
          </Card>
        </div>
      )}
    </>
  )

  const researchBreakthrough = useMemo<ResolutionSummary | undefined>(() => {
    if (!snapshot || snapshot.view.phase === 'completed') return undefined
    const summaries = snapshot.view.recent_resolutions ?? []
    for (let index = summaries.length - 1; index >= 0; index -= 1) {
      const summary = summaries[index]
      if (summary.kind === 'research_breakthrough' && summary.research && !acknowledgedResolutionIDs.has(summary.id)) return summary
    }
    return undefined
  }, [snapshot, acknowledgedResolutionIDs])

  const publicEmpires = snapshot?.decision?.public_empires ?? []
  const unresolvedBattle = useMemo(() => snapshot?.battles.find((battle) => battle.phase !== 'completed'), [snapshot])
  const pendingBattleResolution = useMemo<ResolutionSummary | undefined>(() => {
    if (!snapshot) return undefined
    const summaries = snapshot.view.recent_resolutions ?? []
    for (let index = summaries.length - 1; index >= 0; index -= 1) {
      const summary = summaries[index]
      if (summary.kind === 'battle_completed' && summary.battle && !acknowledgedResolutionIDs.has(summary.id)) return summary
    }
    return undefined
  }, [snapshot, acknowledgedResolutionIDs])
  const blockingBattleID = unresolvedBattle?.spec.id ?? pendingBattleResolution?.battle?.battle_id
  const routeBattleID = route.kind === 'game' && route.section === 'battle' ? route.entityID : undefined
  const battleResolution = useMemo<ResolutionSummary | undefined>(() => {
    if (!snapshot || !routeBattleID) return undefined
    const summaries = snapshot.view.recent_resolutions ?? []
    for (let index = summaries.length - 1; index >= 0; index -= 1) {
      const summary = summaries[index]
      if (summary.kind === 'battle_completed' && summary.battle?.battle_id === routeBattleID) return summary
    }
    return undefined
  }, [snapshot, routeBattleID])

  useEffect(() => {
    if (!snapshot || route.kind !== 'game') return
    const active = snapshot.battles.find((battle) => battle.phase !== 'completed')
    if (active) {
      setResearchOverlayOpen(false)
      if (route.section !== 'battle' || route.entityID !== active.spec.id) {
        navigate({ kind: 'game', gameID: route.gameID, section: 'battle', entityID: active.spec.id })
      }
      return
    }

    const summaries = snapshot.view.recent_resolutions ?? []
    let pendingSummary: ResolutionSummary | undefined
    for (let index = summaries.length - 1; index >= 0; index -= 1) {
      const summary = summaries[index]
      if (summary.kind === 'battle_completed' && summary.battle && !acknowledgedResolutionIDs.has(summary.id)) {
        pendingSummary = summary
        break
      }
    }
    if (pendingSummary?.battle) {
      setResearchOverlayOpen(false)
      if (route.section !== 'battle' || route.entityID !== pendingSummary.battle.battle_id) {
        navigate({ kind: 'game', gameID: route.gameID, section: 'battle', entityID: pendingSummary.battle.battle_id })
      }
      return
    }

    if (route.section !== 'battle') return
    const routedBattle = route.entityID ? snapshot.battles.find((battle) => battle.spec.id === route.entityID) : undefined
    const routedSummary = route.entityID ? summaries.find((summary) => summary.kind === 'battle_completed' && summary.battle?.battle_id === route.entityID) : undefined
    if (!routedBattle || (routedSummary && acknowledgedResolutionIDs.has(routedSummary.id))) {
      navigate({ kind: 'game', gameID: route.gameID, section: 'galaxy' })
    }
  }, [snapshot, route, acknowledgedResolutionIDs])

  const resultWinner = snapshot?.view.result
    ? publicEmpires.find((empire) => empire.id === snapshot.view.result?.winner_empire_id)
    : undefined
  const eliminatedEmpireNames = (snapshot?.view.result?.eliminated_empire_ids ?? []).map((empireID) =>
    publicEmpires.find((empire) => empire.id === empireID)?.name ?? t('result.empireFallback', { id: empireID }),
  )

  function acknowledgeResolution(summaryID: string) {
    if (!gameID || seatID <= 0) return
    setAcknowledgedResolutionIDs((current) => {
      if (current.has(summaryID)) return current
      const next = new Set(current)
      next.add(summaryID)
      try {
        window.sessionStorage.setItem(resolutionAckStorageKey(gameID, seatID), JSON.stringify([...next].slice(-32)))
      } catch {
        // Presentation acknowledgement may remain in-memory when sessionStorage is unavailable.
      }
      return next
    })
  }
  if (route.kind === 'home') {
    return (
      <div className="standalone-shell">
        {persistenceControls}
        <StandaloneHeader />
        <main className="standalone-content home-content">
          <section className="home-hero">
            <p className="eyebrow">{t('menu.eyebrow')}</p>
            <h1>{t('menu.title')}</h1>
            <p>{t('menu.subtitle')}</p>
            <div className="hero-actions">
              <button type="button" className="button-primary" onClick={() => navigate({ kind: 'new-game' })}><GameIcon name="star" />{t('menu.newGame')}</button>
              <button type="button" className="button-secondary" disabled={persistenceBusy} onClick={chooseLoadGame}><GameIcon name="open" />{t('gameMenu.loadGame')}</button>
              {gameID && <button type="button" className="button-secondary" onClick={() => enterGame(gameID)}><GameIcon name="play" />{t('menu.resume')}</button>}
            </div>
          </section>

          {error && <Notice title={t('state.errorTitle')} tone="danger"><p>{error}</p></Notice>}


          <Card>
            <div className="card-heading">
              <div><p className="eyebrow">{t('menu.resume')}</p><h2>{t('menu.hostedGames')}</h2></div>
              <span className={`connection-dot connection-${statusTone}`} aria-hidden="true" />
            </div>
            <p className="muted status-line">{statusText}</p>
            {games.length === 0 ? (
              <p className="muted">{t('menu.noGames')}</p>
            ) : (
              <div className="session-list">
                {games.map((game) => (
                  <button type="button" className="session-row" key={game.game_id} onClick={() => enterGame(game.game_id)}>
                    <span><strong>{game.game_id}</strong><small>{t('top.turn', { turn: game.turn })} · {localizedPhase(t, game.phase)}</small></span>
                    <span aria-hidden="true">›</span>
                  </button>
                ))}
              </div>
            )}
          </Card>
        </main>
      </div>
    )
  }

  if (route.kind === 'new-game') {
    return (
      <div className="standalone-shell">
        <StandaloneHeader onHome={() => navigate({ kind: 'home' })} />
        <main className="standalone-content">
          <PageHeader eyebrow={t('newGame.eyebrow')} title={t('newGame.title')} subtitle={t('newGame.subtitle')} actions={<button type="button" className="button-ghost" onClick={() => navigate({ kind: 'home' })}>{t('common.back')}</button>} />
          {error && <Notice title={t('state.errorTitle')} tone="danger"><p>{error}</p></Notice>}

          <div className="new-game-settings-grid">
            <Card className="new-game-visual-card">
              <VisualSelector<DifficultyID>
                settingId="difficulty"
                label={t('newGame.difficulty')}
                selectedId={difficultyID}
                onChange={setDifficultyID}
                previousLabel={t('newGame.previousOption')}
                nextLabel={t('newGame.nextOption')}
                positionLabel={(current, total) => t('newGame.optionPosition', { current, total })}
                infoLabel={(optionTitle) => t('newGame.moreInfo', { option: optionTitle })}
                closeInfoLabel={t('newGame.closeInfo')}
                options={difficultyIDs.map((id) => {
                  const profile = difficultyProfilesByID.get(id)
                  return {
                    id,
                    title: t(difficultyTitleKey(id)),
                    details: profile
                      ? difficultyDetails(t, profile)
                      : [difficultyCatalogError ? t('newGame.difficultyCatalogUnavailable') : t('newGame.difficultyCatalogLoading')],
                    visual: <DifficultyArt difficulty={id} />,
                    availability: profile ? 'supported' as const : 'planned' as const,
                    availabilityLabel: profile
                      ? t('newGame.supportedNow')
                      : difficultyCatalogError ? t('newGame.difficultyCatalogUnavailable') : t('newGame.difficultyCatalogLoading'),
                  }
                })}
              />
            </Card>
            <Card className="new-game-visual-card">
              <VisualSelector<GalaxySizeID>
                settingId="galaxy-size"
                label={t('newGame.galaxySize')}
                selectedId={galaxySizeID}
                onChange={setGalaxySizeID}
                previousLabel={t('newGame.previousOption')}
                nextLabel={t('newGame.nextOption')}
                positionLabel={(current, total) => t('newGame.optionPosition', { current, total })}
                infoLabel={(optionTitle) => t('newGame.moreInfo', { option: optionTitle })}
                closeInfoLabel={t('newGame.closeInfo')}
                options={galaxySizeIDs.map((id) => {
                  const profile = galaxySizeProfilesByID.get(id)
                  return {
                    id,
                    title: t(galaxySizeTitleKey(id)),
                    details: profile
                      ? galaxySizeDetails(t, profile)
                      : [galaxyCatalogError ? t('newGame.galaxyCatalogUnavailable') : t('newGame.galaxyCatalogLoading')],
                    visual: <GalaxySizeArt size={id} />,
                    availability: profile ? 'supported' as const : 'planned' as const,
                    availabilityLabel: profile
                      ? t('newGame.supportedNow')
                      : galaxyCatalogError ? t('newGame.galaxyCatalogUnavailable') : t('newGame.galaxyCatalogLoading'),
                  }
                })}
              />
            </Card>
            <Card className="new-game-visual-card">
              <VisualSelector<GalaxyAgeID>
                settingId="galaxy-age"
                label={t('newGame.galaxyAge')}
                selectedId={galaxyAgeID}
                onChange={setGalaxyAgeID}
                previousLabel={t('newGame.previousOption')}
                nextLabel={t('newGame.nextOption')}
                positionLabel={(current, total) => t('newGame.optionPosition', { current, total })}
                infoLabel={(optionTitle) => t('newGame.moreInfo', { option: optionTitle })}
                closeInfoLabel={t('newGame.closeInfo')}
                options={galaxyAgeIDs.map((id) => {
                  const profile = galaxyAgeProfilesByID.get(id)
                  return {
                    id,
                    title: t(galaxyAgeTitleKey(id)),
                    details: profile
                      ? galaxyAgeDetails(t, profile)
                      : [galaxyCatalogError ? t('newGame.galaxyCatalogUnavailable') : t('newGame.galaxyCatalogLoading')],
                    visual: <GalaxyAgeArt age={id} />,
                    availability: profile ? 'supported' as const : 'planned' as const,
                    availabilityLabel: profile
                      ? t('newGame.supportedNow')
                      : galaxyCatalogError ? t('newGame.galaxyCatalogUnavailable') : t('newGame.galaxyCatalogLoading'),
                  }
                })}
              />
            </Card>
            <Card className="new-game-visual-card">
              {technologyCatalog ? (
                <VisualSelector<NewGameTechnologyLevel>
                  settingId="technology-level"
                  label={t('newGame.technologyLevel')}
                  selectedId={technologyLevelID}
                  onChange={setTechnologyLevelID}
                  previousLabel={t('newGame.previousOption')}
                  nextLabel={t('newGame.nextOption')}
                  positionLabel={(current, total) => t('newGame.optionPosition', { current, total })}
                  infoLabel={(optionTitle) => t('newGame.moreInfo', { option: optionTitle })}
                  closeInfoLabel={t('newGame.closeInfo')}
                  options={technologyLevelIDs.map((id) => {
                    const profile = technologyProfilesByID.get(id)
                    return {
                      id,
                      title: profile ? technologyTitle(t, profile) : id,
                      details: profile ? profile.facts : [technologyCatalogError ? t('newGame.technologyCatalogUnavailable') : t('newGame.technologyCatalogLoading')],
                      visual: <TechnologyLevelArt level={id} />,
                      availability: profile?.availability ?? 'planned' as const,
                      availabilityLabel: profile?.availability === 'supported' ? t('newGame.supportedNow') : t('newGame.plannedOption'),
                    }
                  })}
                />
              ) : (
                <div className="new-game-race-loading">
                  <h2>{t('newGame.technologyLevel')}</h2>
                  <p>{technologyCatalogError ? t('newGame.technologyCatalogUnavailable') : t('newGame.technologyCatalogLoading')}</p>
                </div>
              )}
            </Card>            <Card className="new-game-visual-card new-game-race-card">
              {raceCatalog ? (
                <VisualSelector<PresetRaceID>
                  settingId="player-race"
                  label={t('newGame.playerRace')}
                  selectedId={playerRaceID}
                  onChange={(id) => {
                    setPlayerRaceID(id)
                    setPlayerName((current) => current === 'Human' || current === 'Klackon' ? (id === 'klackon' ? 'Klackon' : id === 'human' ? 'Human' : current) : current)
                  }}
                  previousLabel={t('newGame.previousOption')}
                  nextLabel={t('newGame.nextOption')}
                  positionLabel={(current, total) => t('newGame.optionPosition', { current, total })}
                  infoLabel={(optionTitle) => t('newGame.moreInfo', { option: optionTitle })}
                  closeInfoLabel={t('newGame.closeInfo')}
                  options={raceCatalog.profiles.map((profile) => ({
                    id: profile.id,
                    title: raceTitle(t, profile),
                    details: raceDetails(t, profile),
                    visual: <RaceArt race={profile.id} />,
                    availability: profile.player_availability,
                    availabilityLabel: profile.player_availability === 'supported' ? t('newGame.supportedNow') : t('newGame.plannedOption'),
                  }))}
                />
              ) : (
                <div className="new-game-race-loading">
                  <h2>{t('newGame.playerRace')}</h2>
                  <p>{raceCatalogError ? t('newGame.raceCatalogUnavailable') : t('newGame.raceCatalogLoading')}</p>
                </div>
              )}
            </Card>
            <Card className="new-game-visual-card new-game-opponent-card">
              {compositionCatalog ? (
                <>
                  <VisualSelector<OpponentCountID>
                    settingId="opponent-count"
                    label={t('newGame.opponents')}
                    selectedId={opponentCountID}
                    onChange={setOpponentCountID}
                    previousLabel={t('newGame.previousOption')}
                    nextLabel={t('newGame.nextOption')}
                    positionLabel={(current, total) => t('newGame.optionPosition', { current, total })}
                    infoLabel={(optionTitle) => t('newGame.moreInfo', { option: optionTitle })}
                    closeInfoLabel={t('newGame.closeInfo')}
                    options={opponentCountIDs.map((id) => {
                      const profile = opponentCountProfilesByID.get(id)
                      const count = Number(id)
                      const assignmentAvailable = compositionCatalog.assignments.some((assignment) => assignment.player_race_id === playerRaceID && assignment.opponent_count === count)
                      const withinGalaxyLimit = count <= (opponentGalaxyLimitsByID.get(galaxySizeID)?.max_supported_opponents ?? 0)
                      const combinationSupported = profile?.availability === 'supported' && assignmentAvailable && withinGalaxyLimit
                      const details = profile
                        ? [...profile.facts, ...(profile.reason_id ? [t('newGame.additionalAIRaceBreadthRequired')] : []), ...(!assignmentAvailable && selectedPlayerRaceProfile?.player_availability === 'planned' ? [t('newGame.playerRaceCompositionPlanned')] : [])]
                        : [compositionCatalogError ? t('newGame.compositionCatalogUnavailable') : t('newGame.compositionCatalogLoading')]
                      return {
                        id,
                        title: t('newGame.opponentCountValue', { count }),
                        details,
                        visual: <OpponentCountArt count={id} />,
                        availability: combinationSupported ? 'supported' as const : 'planned' as const,
                        availabilityLabel: combinationSupported ? t('newGame.supportedNow') : t('newGame.plannedOption'),
                      }
                    })}
                  />
                  {selectedPlayerRaceProfile && (
                    <div className="new-game-opponent-list" aria-label={t('newGame.opponents')}>
                      <div className="new-game-opponent-chip new-game-local-player-chip" data-role="local-player">
                        <img src={`/assets/races/${playerRaceID}/portrait.webp`} alt="" draggable={false} />
                        <span><strong>{raceTitle(t, selectedPlayerRaceProfile)}</strong><small>{t('newGame.localPlayer')}</small></span>
                      </div>
                      {selectedOpponentAssignment && selectedOpponentProfile?.availability === 'supported' && selectedOpponentAssignment.opponents.map((opponent) => {
                        const profile = raceProfilesByID.get(opponent.race_id)
                        return (
                          <div className="new-game-opponent-chip" data-role="opponent" key={opponent.race_id}>
                            <img src={`/assets/races/${opponent.race_id}/portrait.webp`} alt="" draggable={false} />
                            <span><strong>{profile ? raceTitle(t, profile) : opponent.race_id}</strong><small>{t('newGame.aiController')}</small></span>
                          </div>
                        )
                      })}
                    </div>
                  )}
                  {!selectedOpponentAssignment && selectedPlayerRaceProfile?.player_availability === 'planned' && <p className="new-game-composition-note">{t('newGame.playerRaceCompositionPlanned')}</p>}
                  {selectedOpponentProfile?.reason_id === 'additional_ai_race_breadth_required' && <p className="new-game-composition-reason">{t('newGame.additionalAIRaceBreadthRequired')}</p>}
                </>
              ) : (
                <div className="new-game-race-loading">
                  <h2>{t('newGame.opponents')}</h2>
                  <p>{compositionCatalogError ? t('newGame.compositionCatalogUnavailable') : t('newGame.compositionCatalogLoading')}</p>
                </div>
              )}
            </Card>
          </div>

          <Card className="new-game-finalize-card">
            <form className="new-game-form" onSubmit={submitNewGame}>
              <div className="form-grid">
                <label data-field="seed">
                  {t('newGame.seed')}
                  <span className="new-game-seed-control">
                    <input value={newGameSeed} onChange={(event) => setNewGameSeed(event.target.value)} required placeholder={t('newGame.seedPlaceholder')} />
                    <button type="button" className="new-game-seed-reroll" data-action="regenerate-seed" aria-label={t('newGame.regenerateSeed')} title={t('newGame.regenerateSeed')} onClick={() => setNewGameSeed(generateNewGameSeed())}>
                      <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                        <rect x="4" y="4" width="16" height="16" rx="3" />
                        <circle cx="8.5" cy="8.5" r="1.25" />
                        <circle cx="15.5" cy="8.5" r="1.25" />
                        <circle cx="12" cy="12" r="1.25" />
                        <circle cx="8.5" cy="15.5" r="1.25" />
                        <circle cx="15.5" cy="15.5" r="1.25" />
                      </svg>
                    </button>
                  </span>
                </label>
                <label data-field="player-empire">{t('newGame.playerEmpire')}<input value={playerName} onChange={(event) => setPlayerName(event.target.value)} required /></label>
              </div>
              <section className="new-game-launch-briefing" aria-label={t('newGame.launchBriefing')}>
                <h2>{t('newGame.launchBriefing')}</h2>
                <div className="new-game-launch-grid">
                  <article className="new-game-launch-item">
                    <img src={newGameAssetPath('galaxy-size', galaxySizeID, 'svg')} alt="" />
                    <span><small>{t('newGame.launchGalaxy')}</small><strong>{t(galaxySizeTitleKey(galaxySizeID))}</strong><em>{t(galaxyAgeTitleKey(galaxyAgeID))}</em></span>
                  </article>
                  <article className="new-game-launch-item">
                    <img src={newGameAssetPath('difficulty', difficultyAssetOptionID(difficultyID), 'svg')} alt="" />
                    <span><small>{t('newGame.launchDifficulty')}</small><strong>{t(difficultyTitleKey(difficultyID))}</strong></span>
                  </article>
                  <article className="new-game-launch-item">
                    <img src={`/assets/races/${playerRaceID}/portrait.webp`} alt="" />
                    <span><small>{t('newGame.launchPlayer')}</small><strong>{playerName}</strong><em>{raceProfilesByID.get(playerRaceID) ? raceTitle(t, raceProfilesByID.get(playerRaceID)!) : playerRaceID}</em></span>
                  </article>
                  <article className="new-game-launch-item">
                    <img src={`${newGameAssetPath('technology-level', technologyLevelAssetOptionID(technologyLevelID), 'svg')}?v=${technologyLevelAssetVersion}`} alt="" />
                    <span><small>{t('newGame.launchTechnology')}</small><strong>{technologyProfilesByID.get(technologyLevelID) ? technologyTitle(t, technologyProfilesByID.get(technologyLevelID)!) : technologyLevelID}</strong></span>
                  </article>
                  <article className="new-game-launch-item">
                    <img src={newGameAssetPath('opponent-count', opponentCountID, 'svg')} alt="" />
                    <span><small>{t('newGame.launchOpponents')}</small><strong>{t('newGame.opponentCountValue', { count: selectedOpponentCount })}</strong><em>{selectedOpponentAssignment?.opponents.map((opponent) => raceProfilesByID.get(opponent.race_id) ? raceTitle(t, raceProfilesByID.get(opponent.race_id)!) : opponent.race_id).join(' · ') ?? t('newGame.compositionUnavailable')}</em></span>
                  </article>
                  <article className="new-game-launch-item new-game-launch-seed">
                    <div className="new-game-launch-seed-mark" aria-hidden="true">#</div>
                    <span><small>{t('newGame.launchSeed')}</small><strong>{newGameSeed}</strong></span>
                  </article>
                </div>
              </section>
              <button type="submit" className="button-primary button-wide" disabled={creatingGame || !difficultyCatalog || !difficultyProfilesByID.has(difficultyID) || !galaxyCatalog || !galaxySizeProfilesByID.has(galaxySizeID) || !galaxyAgeProfilesByID.has(galaxyAgeID) || !technologyCatalog || technologyProfilesByID.get(technologyLevelID)?.availability !== 'supported' || !raceCatalog || raceProfilesByID.get(playerRaceID)?.player_availability !== 'supported' || !compositionCatalog || !compositionSupported}><GameIcon name="star" />{creatingGame ? t('newGame.creating') : t('newGame.create')}</button>
            </form>
          </Card>
        </main>
      </div>
    )
  }

  const activeSection = route.section
  const phaseText = snapshot ? localizedPhase(t, snapshot.view.phase) : undefined

  return (
    <AppShell
      activeSection={activeSection}
      gameID={route.gameID}
      turn={snapshot?.view.turn}
      phaseLabel={phaseText}
      status={statusText}
      statusTone={statusTone}
      lifecycle={lifecycle}
      resources={resourceChips}
      onNavigate={(section) => blockingBattleID
        ? navigate({ kind: 'game', gameID: route.gameID, section: 'battle', entityID: blockingBattleID })
        : navigate({ kind: 'game', gameID: route.gameID, section })}
      onResourceActivate={(resourceID) => {
        if (blockingBattleID) return false
        if (resourceID !== 'research') return false
        setResearchOverlayOpen(true)
        return true
      }}
      onHome={() => navigate({ kind: 'home' })}
      onSaveGame={() => void saveGame()}
      onLoadGame={chooseLoadGame}
      persistenceBusy={persistenceBusy}
      onEndTurn={snapshot ? () => void endTurn() : undefined}
      endTurnDisabled={!snapshot || mutationLocked || snapshot.view.phase !== 'planning' || planningBusy || previewBusy}
      endTurnLabel={planningBusy ? t('planning.resolving') : t('planning.endTurn')}
      navigationLocked={Boolean(blockingBattleID)}
      immersive={activeSection === 'battle'}
    >
      {persistenceControls}

      {colonyBaseDecision && (
        <div className="decision-dialog-backdrop colony-base-decision-backdrop" role="presentation">
          <Card className="decision-dialog colony-base-decision" as="section">
            <p className="eyebrow">{t('colonyBase.eyebrow')}</p>
            <h2>{t('colonyBase.title')}</h2>
            <p>{t('colonyBase.details', { source: colonyBaseSourcePlanet?.name ?? `#${colonyBaseDecision.source_colony_id}`, system: colonyBaseSystem?.name ?? `#${colonyBaseDecision.system_id}` })}</p>
            <div className="colony-base-target-grid">
              {colonyBaseTargets.map((planet) => (
                <button type="button" className="colony-base-target" key={planet.id} disabled={mutationLocked || colonyBaseBusy} onClick={() => void runColonyBase('colonize', planet.id)}>
                  <span className="colony-base-target-art" aria-hidden="true"><OrbitalBodyArt kind="planet" id={planet.id} climateId={planet.climate_id} /></span>
                  <span><strong>{planet.name}</strong><small>{planet.climate_id} · {planet.size_id} · {planet.mineral_id}</small></span>
                  <GameIcon name="flag" aria-hidden="true" />
                </button>
              ))}
            </div>
            {colonyBaseTargets.length === 0 && <p className="muted">{t('colonyBase.noTargets')}</p>}
            <div className="decision-divider" />
            <div className="colony-base-scrap-row">
              <div><strong>{t('colonyBase.scrapTitle')}</strong><small>{t('colonyBase.scrapHint', { refund: colonyBaseDecision.trash_refund_bc.toFixed(0) })}</small></div>
              <button type="button" className="button-secondary" disabled={mutationLocked || colonyBaseBusy} onClick={() => void runColonyBase('trash')}><GameIcon name="credits" />{t('colonyBase.scrapAction', { refund: colonyBaseDecision.trash_refund_bc.toFixed(0) })}</button>
            </div>
          </Card>
        </div>
      )}

      {planningPreviewError && <Notice title={t('planning.previewErrorTitle')} tone="warning"><p>{planningPreviewError}</p></Notice>}

      {lifecycle === 'refresh-failed' && snapshot && (
        <Notice title={t('state.refreshFailedTitle')} tone="warning">
          <p>{t('state.refreshFailedBody')}</p>
          <button type="button" className="button-secondary" onClick={() => {
            setError('')
            void loadSnapshot(snapshot.view.game_id, seatID).catch(() => undefined)
          }}>{t('state.retryRefresh')}</button>
        </Notice>
      )}

      {error && <Notice title={t('state.errorTitle')} tone="danger"><p>{error}</p></Notice>}

      {snapshot?.view.phase === 'planning' && (draftOrders.length > 0 || previewBusy) && (
        <Card className="planning-bar">
          <div>
            <p className="eyebrow">{t('planning.pending')}</p>
            <strong>{t('planning.draftCount', { count: draftOrders.length })}</strong>
            {previewBusy && <small>{t('planning.previewing')}</small>}
          </div>
        </Card>
      )}

      {invasionDecision && (
        <div className="decision-dialog-backdrop invasion-decision-backdrop" role="presentation">
          <Card className="decision-dialog invasion-decision" as="section">
            <p className="eyebrow">{t('invasion.eyebrow')}</p>
            <h2>{t('invasion.title')}</h2>
            <div className="invasion-brief-grid">
              <div className="invasion-target-art" aria-hidden="true">
                {invasionPlanet ? <OrbitalBodyArt kind="planet" id={invasionPlanet.id} climateId={invasionPlanet.climate_id} /> : <GameIcon name="planet" />}
              </div>
              <div className="invasion-target-copy">
                <span className="badge">{invasionSystem?.name ?? t('invasion.systemFallback', { id: invasionDecision.system_id })}</span>
                <strong>{invasionPlanet?.name ?? t('invasion.colonyFallback', { id: invasionDecision.colony_id })}</strong>
                <small>{invasionPlanet ? `${invasionPlanet.climate_id} · ${invasionPlanet.size_id} · ${invasionPlanet.mineral_id}` : t('invasion.colonyId', { id: invasionDecision.colony_id })}</small>
              </div>
              <div className="invasion-defender-card">
                <span>{t('invasion.defender')}</span>
                <strong>{invasionDefender?.name ?? t('invasion.empireFallback', { id: invasionDecision.defender_empire_id })}</strong>
                <small>{invasionDefender?.race_id ?? t('invasion.unknown')}</small>
              </div>
            </div>
            <div className="invasion-transport-summary">
              <div><GameIcon name="transport" /><span><strong>{t('invasion.transportTitle')}</strong><small>{t('invasion.transports', { count: invasionDecision.eligible_transport_fleet_ids.length })}</small></span></div>
              <div className="invasion-transport-chips">
                {invasionDecision.eligible_transport_fleet_ids.map((fleetID) => { const fleet = invasionTransportFleets.find((candidate) => candidate.id === fleetID); return <span className="badge" key={fleetID}>{t('invasion.fleet', { id: fleetID })}{fleet?.special_kind ? ` · ${fleet.special_kind}` : ''}</span> })}
              </div>
            </div>
            <p className="muted invasion-authority-note">{t('invasion.authorityNote')}</p>
            <div className="action-row invasion-actions">
              <button type="button" className="button-primary" disabled={mutationLocked || invasionBusy || snapshot.view.phase !== 'invasion_decisions'} onClick={() => void runInvasion('invade')}><GameIcon name="flag" />{t('invasion.invade')}</button>
              <button type="button" className="button-secondary" disabled={mutationLocked || invasionBusy || snapshot.view.phase !== 'invasion_decisions'} onClick={() => void runInvasion('decline')}><GameIcon name="close" />{t('invasion.decline')}</button>
            </div>
          </Card>
        </div>
      )}
      {!blockingBattleID && researchBreakthrough?.research && (
        <div className="decision-dialog-backdrop research-breakthrough-backdrop" role="presentation">
          <Card className="decision-dialog research-breakthrough-card" as="section">
            <p className="eyebrow">{t('researchBreakthrough.eyebrow')}</p>
            <h2>{t('researchBreakthrough.title')}</h2>
            <p><strong>{t('researchBreakthrough.field', { field: researchBreakthrough.research.tech_field_id })}</strong></p>
            {researchBreakthrough.research.research_level ? <p className="muted">{t('researchBreakthrough.level', { level: researchBreakthrough.research.research_level })}</p> : null}
            {(researchBreakthrough.research.technology_keys?.length ?? 0) > 0 ? (
              <div className="research-breakthrough-unlocks">
                <span>{t('researchBreakthrough.unlocked')}</span>
                <div className="research-breakthrough-chips">
                  {researchBreakthrough.research.technology_keys?.map((key) => <span className="badge" key={key}>{humanizeResolutionToken(key)}</span>)}
                </div>
              </div>
            ) : (researchBreakthrough.research.technology_ids?.length ?? 0) > 0 ? (
              <div className="research-breakthrough-unlocks">
                <span>{t('researchBreakthrough.unlocked')}</span>
                <div className="research-breakthrough-chips">
                  {researchBreakthrough.research.technology_ids?.map((technologyID) => <span className="badge" key={technologyID}>{t('researchBreakthrough.technologyFallback', { id: technologyID })}</span>)}
                </div>
              </div>
            ) : <p className="muted">{t('researchBreakthrough.noApplications')}</p>}
            <p className="muted research-breakthrough-authority">{t('researchBreakthrough.authority')}</p>
            <div className="action-row research-breakthrough-actions">
              <button type="button" className="button-primary" onClick={() => { acknowledgeResolution(researchBreakthrough.id); setResearchOverlayOpen(true) }}><GameIcon name="research" />{t('researchBreakthrough.chooseNext')}</button>
              <button type="button" className="button-secondary" onClick={() => acknowledgeResolution(researchBreakthrough.id)}>{t('researchBreakthrough.continue')}</button>
            </div>
          </Card>
        </div>
      )}

      {snapshot?.view.result && !blockingBattleID && (
        <section className="result-surface" aria-labelledby="game-result-title">
          <Card className="result-card result-card-dedicated" as="article">
            <p className="eyebrow">{t('result.eyebrow')}</p>
            <h2 id="game-result-title">{t('result.title')}</h2>
            <p className="result-winner"><strong>{t('result.winnerNamed', { empire: resultWinner?.name ?? t('result.empireFallback', { id: snapshot.view.result.winner_empire_id }) })}</strong></p>
            {resultWinner && <span className="badge result-race">{resultWinner.race_id}</span>}
            <p className="muted">{t('result.completedTurn', { turn: snapshot.view.result.completed_turn })}</p>
            <p className="muted">{t('result.eliminatedNamed', { empires: eliminatedEmpireNames.join(', ') || t('result.none') })}</p>
            <p className="muted result-authority">{t('result.authority')}</p>
            <div className="action-row result-actions">
              <button type="button" className="button-primary" onClick={() => navigate({ kind: 'home' })}><GameIcon name="home" />{t('result.mainMenu')}</button>
            </div>
          </Card>
        </section>
      )}

      {!snapshot ? (
        error ? null : <EmptyState title={t('state.loadingTitle')} body={t('state.loadingBody')} />
      ) : activeSection === 'battle' && route.entityID ? (
        <BattleRouteView
          snapshot={snapshot}
          battleID={route.entityID}
          resolution={battleResolution}
          onContinue={(summaryID) => {
            if (summaryID) acknowledgeResolution(summaryID)
            navigate({ kind: 'game', gameID: route.gameID, section: 'galaxy' })
          }}
          onBattleCommand={(command) => runBattleCommand(route.entityID!, command)}
          commandsDisabled={mutationLocked}
          t={t}
        />
      ) : snapshot.view.phase === 'completed' ? null
      : activeSection === 'galaxy' ? (
        <StrategicGalaxyView
          snapshot={snapshot}
          selectedSystemID={route.entityID}
          onSelectSystem={(systemID) => navigate({ kind: 'game', gameID: route.gameID, section: 'galaxy', entityID: systemID })}
          onCloseSystem={() => navigate({ kind: 'game', gameID: route.gameID, section: 'galaxy' })}
          onOpenColony={(colonyID) => navigate({ kind: 'game', gameID: route.gameID, section: 'colonies', entityID: colonyID })}
          draftOrders={draftOrders}
          onPlanOrder={planOrder}
          onRemoveOrder={removePlannedOrder}
          onDeclareWar={(empireID) => runDiplomacy('diplomacy.declare_war', empireID)}
          diplomacyDisabled={diplomacyBusy || diplomacyClosed || mutationLocked}
          t={t}
        />
      ) : activeSection === 'colonies' ? (
        route.subview === 'build' && route.entityID ? (
          <StrategicConstructionView
            snapshot={snapshot}
            preview={planningPreview}
            draftOrders={draftOrders}
            colonyID={route.entityID}
            onBack={() => navigate({ kind: 'game', gameID: route.gameID, section: 'colonies', entityID: route.entityID })}
            onOpenShipDesigner={(designID) => navigate({ kind: 'game', gameID: route.gameID, section: 'shipbuilder', entityID: designID })}
            onPlanOrder={planOrder}
            onRemoveOrder={removePlannedOrder}
            t={t}
          />
        ) : (
          <StrategicColoniesView
            snapshot={snapshot}
            preview={planningPreview}
            draftOrders={draftOrders}
            selectedColonyID={route.entityID}
            onOpenColony={(colonyID) => navigate({ kind: 'game', gameID: route.gameID, section: 'colonies', entityID: colonyID })}
            onOpenConstruction={(colonyID) => navigate({ kind: 'game', gameID: route.gameID, section: 'colonies', entityID: colonyID, subview: 'build' })}
            onBack={() => navigate({ kind: 'game', gameID: route.gameID, section: 'colonies' })}
            onPlanPopulation={planPopulation}
            onPlanOrder={planOrder}
            onRemoveOrder={removePlannedOrder}
            t={t}
          />
        )
      ) : activeSection === 'fleets' ? (
        <StrategicFleetsView snapshot={snapshot} onPlanOrder={planOrder} t={t} />
      ) : activeSection === 'research' ? (
        <div className="research-route-host" aria-hidden="true" />
      ) : activeSection === 'diplomacy' ? (
        <StrategicDiplomacyView snapshot={snapshot} busy={diplomacyBusy} closed={Boolean(diplomacyClosed)} onCommand={runDiplomacy} t={t} />
      ) : activeSection === 'espionage' ? (
        <StrategicEspionageView t={t} />
      ) : activeSection === 'shipbuilder' ? (
        <ShipBuilderView snapshot={snapshot} initialDesignID={route.entityID} reloadSnapshot={() => loadSnapshot()} t={t} />
      ) : (
        <MoreView
          snapshot={snapshot}
          games={games}
          gameID={gameID}
          seatID={seatID}
          setSeatID={setSeatID}
          selectHostedGame={selectHostedGame}
          loadSnapshot={async () => { await loadSnapshot() }}
          diplomacyBusy={diplomacyBusy}
          diplomacyClosed={Boolean(diplomacyClosed)}
          runDiplomacy={runDiplomacy}
          lastNotification={lastNotification}
          onHome={() => navigate({ kind: 'home' })}
          t={t}
        />
      )}
      {(researchOverlayOpen || activeSection === 'research') && snapshot && (
        <StrategicResearchOverlay
          snapshot={snapshot}
          preview={planningPreview}
          draftOrders={draftOrders}
          onPlanOrder={(order) => {
            planOrder(order)
            setResearchOverlayOpen(false)
            if (activeSection === 'research') navigate({ kind: 'game', gameID: route.gameID, section: 'galaxy' })
          }}
          onClose={() => {
            setResearchOverlayOpen(false)
            if (activeSection === 'research') navigate({ kind: 'game', gameID: route.gameID, section: 'galaxy' })
          }}
          t={t}
        />
      )}    </AppShell>
  )
}

function GalaxyView({ snapshot, t }: { snapshot: PlayerSnapshot; t: Translator }) {
  return (
    <>
      <PageHeader eyebrow={t('galaxy.eyebrow')} title={t('galaxy.title')} subtitle={t('galaxy.subtitle')} />
      <section className="metric-grid">
        <Metric label={t('metric.game')} value={snapshot.view.game_id} />
        <Metric label={t('metric.turn')} value={snapshot.view.turn} />
        <Metric label={t('metric.phase')} value={localizedPhase(t, snapshot.view.phase)} />
        <Metric label={t('metric.revision')} value={snapshot.view.revision} />
        <Metric label={t('metric.change')} value={snapshot.change_sequence} />
      </section>
      <div className="content-grid content-grid-2">
        <Card>
          <p className="eyebrow">{t('galaxy.empire')}</p>
          <h2>{snapshot.view.empire.name}</h2>
          <dl className="detail-list">
            <div><dt>{t('galaxy.empireId')}</dt><dd>{snapshot.view.empire.id}</dd></div>
            <div><dt>{t('galaxy.race')}</dt><dd>{snapshot.view.empire.race_id}</dd></div>
            <div><dt>{t('galaxy.seat')}</dt><dd>{snapshot.view.seat.seat.id} · {snapshot.view.seat.seat.controller}</dd></div>
          </dl>
        </Card>
        <Card>
          <div className="summary-stats">
            <div><span>{t('galaxy.colonies')}</span><strong>{snapshot.view.colonies.length}</strong></div>
            <div><span>{t('galaxy.battles')}</span><strong>{snapshot.battles.length}</strong></div>
            <div><span>{t('galaxy.contacts')}</span><strong>{snapshot.view.diplomacy?.length ?? 0}</strong></div>
          </div>
          <p className="muted card-note">{t('galaxy.mapPending')}</p>
        </Card>
      </div>
    </>
  )
}

function ColoniesView({ snapshot, assignment, setAssignment, totalDraft, onSubmit, t }: {
  snapshot: PlayerSnapshot
  assignment: AssignmentDraft
  setAssignment: (update: AssignmentDraft | ((current: AssignmentDraft) => AssignmentDraft)) => void
  totalDraft: number
  onSubmit: (event: FormEvent) => void
  t: Translator
}) {
  const first = snapshot.view.colonies[0]
  return (
    <>
      <PageHeader eyebrow={t('colonies.eyebrow')} title={t('colonies.title')} subtitle={t('colonies.subtitle')} />
      {snapshot.view.colonies.length === 0 ? (
        <EmptyState title={t('colonies.noColonies')} />
      ) : (
        <div className="content-grid content-grid-2">
          {snapshot.view.colonies.map((colony, index) => {
            const population = aggregatePopulation(colony)
            return (
              <Card key={colony.id} className={index === 0 ? 'primary-colony' : ''}>
                <div className="card-heading">
                  <div><p className="eyebrow">{t('colonies.planet', { id: colony.planet_id })}</p><h2>{t('colonies.colony', { id: colony.id })}</h2></div>
                  <span className="badge">{population.farmers + population.workers + population.scientists}</span>
                </div>
                <dl className="detail-list compact">
                  <div><dt>{t('colonies.farmers')}</dt><dd>{population.farmers}</dd></div>
                  <div><dt>{t('colonies.workers')}</dt><dd>{population.workers}</dd></div>
                  <div><dt>{t('colonies.scientists')}</dt><dd>{population.scientists}</dd></div>
                  <div><dt>{t('colonies.infantry')}</dt><dd>{colony.ground_forces?.infantry ?? 0}</dd></div>
                </dl>
                {first?.id === colony.id && (
                  <form className="population-form" onSubmit={onSubmit}>
                    <h3>{t('colonies.population')}</h3>
                    <div className="three-fields">
                      <label>{t('colonies.farmers')}<input type="number" min="0" step="0.1" value={assignment.farmers} onChange={(event) => setAssignment((draft) => ({ ...draft, farmers: event.target.value }))} /></label>
                      <label>{t('colonies.workers')}<input type="number" min="0" step="0.1" value={assignment.workers} onChange={(event) => setAssignment((draft) => ({ ...draft, workers: event.target.value }))} /></label>
                      <label>{t('colonies.scientists')}<input type="number" min="0" step="0.1" value={assignment.scientists} onChange={(event) => setAssignment((draft) => ({ ...draft, scientists: event.target.value }))} /></label>
                    </div>
                    <div className="form-footer"><span>{t('colonies.assignedTotal')}: <strong>{Number.isFinite(totalDraft) ? totalDraft : t('colonies.invalid')}</strong></span><button type="submit" className="button-primary"><GameIcon name="check" />{t('colonies.submit')}</button></div>
                  </form>
                )}
              </Card>
            )
          })}
        </div>
      )}
    </>
  )
}

function FleetsView({ snapshot, t }: { snapshot: PlayerSnapshot; t: Translator }) {
  return (
    <>
      <PageHeader eyebrow={t('fleets.eyebrow')} title={t('fleets.title')} subtitle={t('fleets.subtitle')} />
      <div className="content-grid content-grid-2">
        <Card>
          <h2>{t('fleets.battles')}</h2>
          {snapshot.battles.length === 0 ? <p className="muted">{t('fleets.noBattles')}</p> : (
            <div className="list-stack">{snapshot.battles.map((battle) => <div className="list-row" key={battle.spec.id}><span><strong>{t('fleets.battle', { id: battle.spec.id })}</strong><small>{localizedPhase(t, battle.phase)}</small></span><span className="badge">{battle.spec.participants.length}</span></div>)}</div>
          )}
        </Card>
        <EmptyState title={t('fleets.noProjection')} body={t('common.notAvailableYet')} />
      </div>
    </>
  )
}

function ResearchView({ t }: { t: Translator }) {
  return (
    <>
      <PageHeader eyebrow={t('research.eyebrow')} title={t('research.title')} subtitle={t('research.subtitle')} />
      <EmptyState title={t('research.pending')} body={t('common.notAvailableYet')} />
    </>
  )
}

function MoreView({ snapshot, games, gameID, seatID, setSeatID, selectHostedGame, loadSnapshot, diplomacyBusy, diplomacyClosed, runDiplomacy, lastNotification, onHome, t }: {
  snapshot: PlayerSnapshot
  games: GameSummary[]
  gameID: string
  seatID: number
  setSeatID: (seatID: number) => void
  selectHostedGame: (gameID: string) => void
  loadSnapshot: () => Promise<void>
  diplomacyBusy: boolean
  diplomacyClosed: boolean
  runDiplomacy: (kind: DiplomacyCommandKind, otherEmpireID: number) => Promise<void>
  lastNotification: Notification | null
  onHome: () => void
  t: Translator
}) {
  return (
    <>
      <PageHeader eyebrow={t('more.eyebrow')} title={t('more.title')} subtitle={t('more.subtitle')} />
      <div className="content-grid content-grid-2">
        <Card>
          <h2>{t('more.language')}</h2>
          <p className="muted">{t('more.languageHint')}</p>
          <LanguageSwitch />
        </Card>
        <Card>
          <h2>{t('more.gameSession')}</h2>
          <div className="settings-stack">
            <label>{t('more.hostedGame')}<select value={gameID} onChange={(event) => selectHostedGame(event.target.value)}>{games.map((game) => <option key={game.game_id} value={game.game_id}>{game.game_id}</option>)}</select></label>
            <label>{t('more.developmentSeat')}<input type="number" min="1" value={seatID} onChange={(event) => setSeatID(Number(event.target.value))} /></label>
            <div className="action-row"><button type="button" className="button-secondary" onClick={() => void loadSnapshot()} disabled={!gameID}>{t('more.refreshSnapshot')}</button><button type="button" className="button-ghost" onClick={onHome}>{t('more.mainMenu')}</button></div>
          </div>
        </Card>
        <Card>
          <h2>{t('more.diplomacy')}</h2>
          {(snapshot.view.diplomacy ?? []).length === 0 ? <p className="muted">{t('more.noEmpires')}</p> : (
            <div className="list-stack">
              {(snapshot.view.diplomacy ?? []).map((relation) => {
                const otherSeat = snapshot.view.seats.find((candidate) => candidate.seat.empire_id === relation.other_empire_id)
                const disabled = diplomacyBusy || diplomacyClosed
                return (
                  <div className="diplomacy-row" key={relation.other_empire_id}>
                    <div><strong>{otherSeat?.seat.name ?? t('common.empireFallback', { id: relation.other_empire_id })}</strong><small>{localizedStance(t, relation.stance)}</small></div>
                    <div className="action-row compact-actions">
                      {relation.incoming_peace_offer && <button type="button" className="button-secondary" disabled={disabled} onClick={() => void runDiplomacy('diplomacy.accept_peace', relation.other_empire_id)}>{t('more.acceptPeace')}</button>}
                      {relation.stance === 'war' && !relation.incoming_peace_offer && !relation.outgoing_peace_offer && <button type="button" className="button-secondary" disabled={disabled} onClick={() => void runDiplomacy('diplomacy.offer_peace', relation.other_empire_id)}>{t('more.offerPeace')}</button>}
                      {(relation.stance === 'neutral' || relation.stance === 'peace') && <button type="button" className="button-danger" disabled={disabled} onClick={() => void runDiplomacy('diplomacy.declare_war', relation.other_empire_id)}>{t('more.declareWar')}</button>}
                    </div>
                    {relation.outgoing_peace_offer && <small className="muted">{t('more.peacePending')}</small>}
                  </div>
                )
              })}
              {diplomacyClosed && <p className="muted">{t('more.diplomacyTiming')}</p>}
            </div>
          )}
        </Card>
        <Card>
          <h2>{t('more.diagnostics')}</h2>
          <p className="muted">{t('more.diagnosticsHint')}</p>
          <details className="diagnostics">
            <summary>{t('more.latestInvalidation')}</summary>
            {lastNotification ? <pre>{JSON.stringify(lastNotification, null, 2)}</pre> : <p className="muted">{t('more.noInvalidation')}</p>}
          </details>
        </Card>
      </div>
    </>
  )
}

export default App
