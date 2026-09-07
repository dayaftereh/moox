import { type FormEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  aggregatePopulation,
  createGame,
  getPlayerSnapshot,
  listGames,
  previewPlanning,
  removeDraftOrder,
  replaceDraftOrder,
  streamURL,
  submitDiplomacy,
  submitInvasion,
  submitPlanning,
  type DiplomacyCommandKind,
  type DiplomaticStance,
  type DraftOrder,
  type GameSummary,
  type Notification,
  type PlanningPreviewSnapshot,
  type PlayerSnapshot,
} from './api'
import { AppShell, LanguageSwitch, StandaloneHeader, type ResourceChip } from './components/AppShell'
import { Card, EmptyState, Metric, Notice, PageHeader } from './components/ui'
import { type TranslationKey, type TranslationVars, useI18n } from './i18n'
import { type AppRoute, type GameSection, navigate, parseRoute } from './navigation'
import { ShipBuilderView } from './ShipBuilderView'
import {
  StrategicColoniesView,
  StrategicConstructionView,
  StrategicDiplomacyView,
  StrategicEspionageView,
  StrategicFleetsView,
  StrategicGalaxyView,
  StrategicResearchOverlay,
  StrategicResearchView,
} from './StrategicViews'
import './styles.css'

type AssignmentDraft = {
  farmers: string
  workers: string
  scientists: string
}

type StatusMessage = {
  key: TranslationKey
  vars?: TranslationVars
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
  const [newGameID, setNewGameID] = useState('game-1')
  const [newGameSeed, setNewGameSeed] = useState('0x8009')
  const [humanName, setHumanName] = useState('Human')
  const [darlokName, setDarlokName] = useState('Darlok')
  const [creatingGame, setCreatingGame] = useState(false)
  const [gameID, setGameID] = useState('')
  const [seatID, setSeatID] = useState(1)
  const [snapshot, setSnapshot] = useState<PlayerSnapshot | null>(null)
  const [assignment, setAssignment] = useState<AssignmentDraft>({ farmers: '0', workers: '0', scientists: '0' })
  const [draftOrders, setDraftOrders] = useState<DraftOrder[]>([])
  const [planningPreview, setPlanningPreview] = useState<PlanningPreviewSnapshot | null>(null)
  const [previewBusy, setPreviewBusy] = useState(false)
  const [planningBusy, setPlanningBusy] = useState(false)
  const [status, setStatus] = useState<StatusMessage>({ key: 'status.connecting' })
  const [error, setError] = useState('')
  const [diplomacyBusy, setDiplomacyBusy] = useState(false)
  const [invasionBusy, setInvasionBusy] = useState(false)
  const [researchOverlayOpen, setResearchOverlayOpen] = useState(false)
  const [lastNotification, setLastNotification] = useState<Notification | null>(null)
  const reconnectTimer = useRef<number | null>(null)
  const activeGameRef = useRef(gameID)
  const activeSeatRef = useRef(seatID)
  activeGameRef.current = gameID
  activeSeatRef.current = seatID

  useEffect(() => {
    const syncRoute = () => setRoute(parseRoute())
    window.addEventListener('hashchange', syncRoute)
    if (!window.location.hash) navigate({ kind: 'home' })
    return () => window.removeEventListener('hashchange', syncRoute)
  }, [])

  useEffect(() => {
    if (route.kind === 'game' && route.gameID !== gameID) {
      setSnapshot(null)
      setError('')
      setGameID(route.gameID)
    }
  }, [route, gameID])

  const loadSnapshot = useCallback(async (selectedGameID = gameID, selectedSeatID = seatID, signal?: AbortSignal) => {
    if (!selectedGameID || selectedSeatID <= 0) return
    const next = await getPlayerSnapshot(selectedGameID, selectedSeatID, signal)
    if (selectedGameID !== activeGameRef.current || selectedSeatID !== activeSeatRef.current) return
    setSnapshot(next)
    setDraftOrders([])
    setPlanningPreview(null)
    const firstColony = next.view.colonies[0]
    if (firstColony) {
      const population = aggregatePopulation(firstColony)
      setAssignment({
        farmers: String(population.farmers),
        workers: String(population.workers),
        scientists: String(population.scientists),
      })
    }
    setStatus({ key: 'status.snapshotSynced', vars: { change: next.change_sequence } })
    setError('')
  }, [gameID, seatID])

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
          setError(reason instanceof Error ? reason.message : String(reason))
          setStatus({ key: 'status.discoverFailed' })
        }
      })
    return () => controller.abort()
  }, [])

  useEffect(() => {
    if (!gameID || seatID <= 0) return
    const controller = new AbortController()
    void loadSnapshot(gameID, seatID, controller.signal).catch((reason: unknown) => {
      if (!controller.signal.aborted) setError(reason instanceof Error ? reason.message : String(reason))
    })
    return () => controller.abort()
  }, [gameID, seatID, loadSnapshot])

  useEffect(() => {
    if (!gameID) return
    let disposed = false
    let socket: WebSocket | null = null

    const connect = () => {
      if (disposed) return
      void loadSnapshot(gameID, seatID).catch(() => undefined)
      socket = new WebSocket(streamURL(gameID))
      socket.onopen = () => setStatus({ key: 'status.liveConnected' })
      socket.onmessage = (event) => {
        try {
          const notification = JSON.parse(String(event.data)) as Notification
          setLastNotification(notification)
          setStatus({ key: 'status.invalidated', vars: { reason: notification.reason } })
          setSnapshot((current) => {
            if (!current || notification.change_sequence > current.change_sequence) {
              void loadSnapshot(gameID, seatID).catch((reason: unknown) => {
                setError(reason instanceof Error ? reason.message : String(reason))
              })
            }
            return current
          })
        } catch (reason) {
          setError(reason instanceof Error ? reason.message : String(reason))
        }
      }
      socket.onclose = () => {
        if (!disposed) {
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
      socket?.close(1000, 'component disposed')
    }
  }, [gameID, seatID, loadSnapshot])

  useEffect(() => {
    if (!snapshot || snapshot.view.phase !== 'planning' || snapshot.view.seat.submitted) {
      setPlanningPreview(null)
      setPreviewBusy(false)
      return
    }
    const controller = new AbortController()
    setPreviewBusy(true)
    void previewPlanning(snapshot, seatID, draftOrders, controller.signal)
      .then((preview) => {
        if (!controller.signal.aborted) {
          setPlanningPreview(preview)
          setError('')
        }
      })
      .catch((cause) => {
        if (!controller.signal.aborted) setError(cause instanceof Error ? cause.message : String(cause))
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
  const diplomacyClosed = snapshot?.view.phase !== 'planning' || Boolean(snapshot?.view.seats.some((seat) => seat.submitted))
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
  const statusText = t(status.key, status.vars)
  const statusTone: 'neutral' | 'success' | 'warning' | 'danger' = error
    ? 'danger'
    : status.key === 'status.liveDisconnected'
      ? 'warning'
      : ['status.liveConnected', 'status.snapshotSynced', 'status.commandAccepted', 'status.diplomacyAccepted', 'status.invasionAccepted'].includes(status.key)
        ? 'success'
        : 'neutral'

  async function submitNewGame(event: FormEvent) {
    event.preventDefault()
    setCreatingGame(true)
    setError('')
    try {
      const created = await createGame({
        schema_version: 1,
        game_id: newGameID,
        seed: newGameSeed,
        settings: {
          galaxy_size: 'small',
          galaxy_age: 'normal',
          technology_level: 'average',
          strategic_combat: false,
          players: [
            { seat_id: 1, empire_name: humanName, race_id: 'human' },
            { seat_id: 2, empire_name: darlokName, race_id: 'darlok' },
          ],
        },
      })
      setGames(await listGames())
      setSnapshot(null)
      setGameID(created.game.game_id)
      setSeatID(created.players[0]?.seat_id ?? 1)
      setStatus({ key: 'status.createdGame', vars: { game: created.game.game_id, seed: newGameSeed } })
      navigate({ kind: 'game', gameID: created.game.game_id, section: 'galaxy' })
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason))
    } finally {
      setCreatingGame(false)
    }
  }

  function planOrder(order: DraftOrder) {
    setDraftOrders((current) => replaceDraftOrder(current, order))
  }

  function removePlannedOrder(key: string) {
    setDraftOrders((current) => removeDraftOrder(current, key))
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

  async function endTurn() {
    if (!snapshot || planningBusy || snapshot.view.phase !== 'planning') return
    setPlanningBusy(true)
    setError('')
    try {
      const receipt = await submitPlanning(snapshot, seatID, draftOrders)
      setDraftOrders([])
      setPlanningPreview(null)
      setStatus({ key: 'status.commandAccepted', vars: { change: receipt.change_sequence, revision: receipt.game_revision } })
      await loadSnapshot(snapshot.view.game_id, seatID)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally {
      setPlanningBusy(false)
    }
  }
  async function runDiplomacy(kind: DiplomacyCommandKind, otherEmpireID: number) {
    if (!snapshot) return
    setDiplomacyBusy(true)
    setError('')
    try {
      const receipt = await submitDiplomacy(snapshot, seatID, kind, otherEmpireID)
      setStatus({ key: 'status.diplomacyAccepted', vars: { change: receipt.change_sequence, revision: receipt.game_revision } })
      await loadSnapshot(snapshot.view.game_id, seatID)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason))
    } finally {
      setDiplomacyBusy(false)
    }
  }

  async function runInvasion(action: 'invade' | 'decline') {
    if (!snapshot?.view.invasion) return
    setInvasionBusy(true)
    setError('')
    try {
      const receipt = await submitInvasion(snapshot, seatID, action)
      setStatus({ key: 'status.invasionAccepted', vars: { change: receipt.change_sequence, revision: receipt.game_revision } })
      await loadSnapshot()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally {
      setInvasionBusy(false)
    }
  }

  function enterGame(selectedGameID: string, section: GameSection = 'galaxy') {
    const reloadSelectedGame = selectedGameID === gameID
    setGameID(selectedGameID)
    setSnapshot(null)
    navigate({ kind: 'game', gameID: selectedGameID, section })
    if (reloadSelectedGame) {
      void loadSnapshot(selectedGameID, seatID).catch((cause: unknown) => {
        setError(cause instanceof Error ? cause.message : String(cause))
      })
    }
  }

  function selectHostedGame(selectedGameID: string) {
    enterGame(selectedGameID, route.kind === 'game' ? route.section : 'galaxy')
  }

  if (route.kind === 'home') {
    return (
      <div className="standalone-shell">
        <StandaloneHeader />
        <main className="standalone-content home-content">
          <section className="home-hero">
            <p className="eyebrow">{t('menu.eyebrow')}</p>
            <h1>{t('menu.title')}</h1>
            <p>{t('menu.subtitle')}</p>
            <div className="hero-actions">
              <button type="button" className="button-primary" onClick={() => navigate({ kind: 'new-game' })}>{t('menu.newGame')}</button>
              {gameID && <button type="button" className="button-secondary" onClick={() => enterGame(gameID)}>{t('menu.resume')}</button>}
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

          <Card>
            <form className="new-game-form" onSubmit={submitNewGame}>
              <div className="form-grid">
                <label>{t('newGame.gameId')}<input value={newGameID} onChange={(event) => setNewGameID(event.target.value)} required /></label>
                <label>{t('newGame.seed')}<input value={newGameSeed} onChange={(event) => setNewGameSeed(event.target.value)} required placeholder={t('newGame.seedPlaceholder')} /></label>
                <label>{t('newGame.galaxy')}<input value={t('newGame.galaxyValue')} disabled /></label>
                <label>{t('newGame.techCombat')}<input value={t('newGame.techCombatValue')} disabled /></label>
                <label>{t('newGame.humanEmpire')}<input value={humanName} onChange={(event) => setHumanName(event.target.value)} required /></label>
                <label>{t('newGame.darlokEmpire')}<input value={darlokName} onChange={(event) => setDarlokName(event.target.value)} required /></label>
              </div>
              <button type="submit" className="button-primary button-wide" disabled={creatingGame}>{creatingGame ? t('newGame.creating') : t('newGame.create')}</button>
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
      resources={resourceChips}
      onNavigate={(section) => navigate({ kind: 'game', gameID: route.gameID, section })}
      onResourceActivate={(resourceID) => {
        if (resourceID !== 'research') return false
        setResearchOverlayOpen(true)
        return true
      }}
      onHome={() => navigate({ kind: 'home' })}
      onEndTurn={snapshot ? () => void endTurn() : undefined}
      endTurnDisabled={!snapshot || snapshot.view.phase !== 'planning' || planningBusy || previewBusy}
      endTurnLabel={planningBusy ? t('planning.resolving') : t('planning.endTurn')}
    >
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

      {snapshot?.view.invasion && (
        <Card className="priority-card">
          <p className="eyebrow">{t('invasion.eyebrow')}</p>
          <h2>{t('invasion.title')}</h2>
          <p>{t('invasion.details', {
            colony: snapshot.view.invasion.colony_id,
            system: snapshot.view.invasion.system_id,
            defender: snapshot.view.invasion.defender_empire_id,
          })}</p>
          <p className="muted">{t('invasion.transports', { count: snapshot.view.invasion.eligible_transport_fleet_ids.length })}</p>
          <div className="action-row">
            <button type="button" className="button-primary" disabled={invasionBusy || snapshot.view.phase !== 'invasion_decisions'} onClick={() => void runInvasion('invade')}>{t('invasion.invade')}</button>
            <button type="button" className="button-secondary" disabled={invasionBusy || snapshot.view.phase !== 'invasion_decisions'} onClick={() => void runInvasion('decline')}>{t('invasion.decline')}</button>
          </div>
        </Card>
      )}

      {snapshot?.view.result && (
        <Card className="result-card" as="article">
          <p className="eyebrow">{t('result.eyebrow')}</p>
          <h2>{t('result.title')}</h2>
          <p><strong>{t('result.winner', { empire: snapshot.view.result.winner_empire_id, seat: snapshot.view.result.winner_seat_id })}</strong></p>
          <p className="muted">{t('result.completed', { turn: snapshot.view.result.completed_turn, revision: snapshot.view.result.completed_revision })}</p>
          <p className="muted">{t('result.eliminated', { empires: snapshot.view.result.eliminated_empire_ids.join(', ') || t('result.none') })}</p>
        </Card>
      )}

      {!snapshot ? (
        error ? null : <EmptyState title={t('state.loadingTitle')} body={t('state.loadingBody')} />
      ) : activeSection === 'galaxy' ? (
        <StrategicGalaxyView
          snapshot={snapshot}
          selectedSystemID={route.entityID}
          onSelectSystem={(systemID) => navigate({ kind: 'game', gameID: route.gameID, section: 'galaxy', entityID: systemID })}
          onCloseSystem={() => navigate({ kind: 'game', gameID: route.gameID, section: 'galaxy' })}
          onOpenColony={(colonyID) => navigate({ kind: 'game', gameID: route.gameID, section: 'colonies', entityID: colonyID })}
          onPlanOrder={planOrder}
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
        <StrategicResearchView snapshot={snapshot} preview={planningPreview} onPlanOrder={planOrder} t={t} />
      ) : activeSection === 'diplomacy' ? (
        <StrategicDiplomacyView snapshot={snapshot} busy={diplomacyBusy} closed={Boolean(diplomacyClosed)} onCommand={runDiplomacy} t={t} />
      ) : activeSection === 'espionage' ? (
        <StrategicEspionageView t={t} />
      ) : activeSection === 'shipbuilder' ? (
        <ShipBuilderView snapshot={snapshot} t={t} />
      ) : (
        <MoreView
          snapshot={snapshot}
          games={games}
          gameID={gameID}
          seatID={seatID}
          setSeatID={setSeatID}
          selectHostedGame={selectHostedGame}
          loadSnapshot={loadSnapshot}
          diplomacyBusy={diplomacyBusy}
          diplomacyClosed={Boolean(diplomacyClosed)}
          runDiplomacy={runDiplomacy}
          lastNotification={lastNotification}
          onHome={() => navigate({ kind: 'home' })}
          t={t}
        />
      )}
      {researchOverlayOpen && snapshot && (
        <StrategicResearchOverlay
          snapshot={snapshot}
          preview={planningPreview}
          draftOrders={draftOrders}
          onPlanOrder={(order) => {
            planOrder(order)
            setResearchOverlayOpen(false)
            navigate({ kind: 'game', gameID: route.gameID, section: 'galaxy' })
          }}
          onClose={() => setResearchOverlayOpen(false)}
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
                    <div className="form-footer"><span>{t('colonies.assignedTotal')}: <strong>{Number.isFinite(totalDraft) ? totalDraft : t('colonies.invalid')}</strong></span><button type="submit" className="button-primary">{t('colonies.submit')}</button></div>
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
