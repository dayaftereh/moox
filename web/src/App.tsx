import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  aggregatePopulation,
  assignPopulation,
  getPlayerSnapshot,
  listGames,
  streamURL,
  type GameSummary,
  type Notification,
  type PlayerSnapshot,
} from './api'
import './styles.css'

type AssignmentDraft = {
  farmers: string
  workers: string
  scientists: string
}

function App() {
  const [games, setGames] = useState<GameSummary[]>([])
  const [gameID, setGameID] = useState('')
  const [seatID, setSeatID] = useState(1)
  const [snapshot, setSnapshot] = useState<PlayerSnapshot | null>(null)
  const [assignment, setAssignment] = useState<AssignmentDraft>({ farmers: '0', workers: '0', scientists: '0' })
  const [status, setStatus] = useState('Connecting to MOOX server...')
  const [error, setError] = useState('')
  const [lastNotification, setLastNotification] = useState<Notification | null>(null)
  const reconnectTimer = useRef<number | null>(null)

  const loadSnapshot = useCallback(async (selectedGameID = gameID, selectedSeatID = seatID, signal?: AbortSignal) => {
    if (!selectedGameID || selectedSeatID <= 0) return
    const next = await getPlayerSnapshot(selectedGameID, selectedSeatID, signal)
    setSnapshot(next)
    const firstColony = next.view.colonies[0]
    if (firstColony) {
      const population = aggregatePopulation(firstColony)
      setAssignment({
        farmers: String(population.farmers),
        workers: String(population.workers),
        scientists: String(population.scientists),
      })
    }
    setStatus(`Snapshot synchronized at change ${next.change_sequence}`)
    setError('')
  }, [gameID, seatID])

  useEffect(() => {
    const controller = new AbortController()
    void listGames(controller.signal)
      .then((available) => {
        setGames(available)
        if (available.length === 0) {
          setStatus('No hosted games are available.')
          return
        }
        const selected = available[0].game_id
        setGameID(selected)
        setStatus(`Connected to hosted game ${selected}`)
      })
      .catch((reason: unknown) => {
        if (!controller.signal.aborted) {
          setError(reason instanceof Error ? reason.message : String(reason))
          setStatus('Unable to discover hosted games.')
        }
      })
    return () => controller.abort()
  }, [])

  useEffect(() => {
    if (!gameID || seatID <= 0) return
    const controller = new AbortController()
    void loadSnapshot(gameID, seatID, controller.signal).catch((reason: unknown) => {
      if (!controller.signal.aborted) {
        setError(reason instanceof Error ? reason.message : String(reason))
      }
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
      socket.onopen = () => setStatus('Live invalidation stream connected.')
      socket.onmessage = (event) => {
        try {
          const notification = JSON.parse(String(event.data)) as Notification
          setLastNotification(notification)
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
          setStatus('Live stream disconnected; resynchronizing and retrying...')
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

  const firstColony = snapshot?.view.colonies[0]
  const totalDraft = useMemo(
    () => Number(assignment.farmers || 0) + Number(assignment.workers || 0) + Number(assignment.scientists || 0),
    [assignment],
  )

  async function submitAssignment(event: FormEvent) {
    event.preventDefault()
    if (!snapshot || !firstColony) return
    setError('')
    try {
      const receipt = await assignPopulation(
        snapshot,
        seatID,
        firstColony.id,
        Number(assignment.farmers),
        Number(assignment.workers),
        Number(assignment.scientists),
      )
      setStatus(`Command accepted: change ${receipt.change_sequence}, game revision ${receipt.game_revision}`)
      await loadSnapshot(snapshot.view.game_id, seatID)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason))
    }
  }

  return (
    <main className="shell">
      <header className="hero">
        <div>
          <p className="eyebrow">Master of Orion X</p>
          <h1>Authoritative Web HMI</h1>
          <p className="muted">Slice 08 transport baseline. HTTP is authoritative; WebSocket only invalidates snapshots.</p>
        </div>
        <div className="connection" aria-live="polite">{status}</div>
      </header>

      <section className="panel controls">
        <label>
          Hosted game
          <select value={gameID} onChange={(event) => setGameID(event.target.value)}>
            {games.map((game) => <option key={game.game_id} value={game.game_id}>{game.game_id}</option>)}
          </select>
        </label>
        <label>
          Development Seat
          <input type="number" min={1} value={seatID} onChange={(event) => setSeatID(Number(event.target.value))} />
        </label>
        <button type="button" onClick={() => void loadSnapshot()} disabled={!gameID}>Refresh snapshot</button>
      </section>

      {error && <section className="panel error" role="alert">{error}</section>}

      {snapshot && (
        <>
          <section className="metrics">
            <article className="metric"><span>Game</span><strong>{snapshot.view.game_id}</strong></article>
            <article className="metric"><span>Turn</span><strong>{snapshot.view.turn}</strong></article>
            <article className="metric"><span>Phase</span><strong>{snapshot.view.phase}</strong></article>
            <article className="metric"><span>Game revision</span><strong>{snapshot.view.revision}</strong></article>
            <article className="metric"><span>Change sequence</span><strong>{snapshot.change_sequence}</strong></article>
          </section>

          <section className="grid">
            <article className="panel">
              <h2>{snapshot.view.empire.name}</h2>
              <dl>
                <div><dt>Empire ID</dt><dd>{snapshot.view.empire.id}</dd></div>
                <div><dt>Race</dt><dd>{snapshot.view.empire.race_id}</dd></div>
                <div><dt>Seat</dt><dd>{snapshot.view.seat.seat.id} / {snapshot.view.seat.seat.controller}</dd></div>
              </dl>
            </article>

            <article className="panel">
              <h2>Participant Battles</h2>
              {snapshot.battles.length === 0 ? <p className="muted">No active participant battles.</p> : (
                <ul>{snapshot.battles.map((battle) => <li key={battle.spec.id}>Battle {battle.spec.id}: {battle.phase}</li>)}</ul>
              )}
            </article>
          </section>

          {firstColony && (
            <section className="panel">
              <h2>Concrete gameplay command</h2>
              <p className="muted">This form submits the real <code>colony.assign_population</code> command in a revision-bound CommandBatch.</p>
              <form className="assignment" onSubmit={submitAssignment}>
                <label>Farmers<input type="number" min="0" step="0.1" value={assignment.farmers} onChange={(event) => setAssignment((draft) => ({ ...draft, farmers: event.target.value }))} /></label>
                <label>Workers<input type="number" min="0" step="0.1" value={assignment.workers} onChange={(event) => setAssignment((draft) => ({ ...draft, workers: event.target.value }))} /></label>
                <label>Scientists<input type="number" min="0" step="0.1" value={assignment.scientists} onChange={(event) => setAssignment((draft) => ({ ...draft, scientists: event.target.value }))} /></label>
                <div className="total">Assigned total: <strong>{Number.isFinite(totalDraft) ? totalDraft : 'invalid'}</strong></div>
                <button type="submit">Submit turn</button>
              </form>
            </section>
          )}

          <section className="panel">
            <h2>Latest invalidation</h2>
            {lastNotification ? (
              <pre>{JSON.stringify(lastNotification, null, 2)}</pre>
            ) : <p className="muted">No WebSocket invalidation received in this browser session yet.</p>}
          </section>
        </>
      )}
    </main>
  )
}

export default App
