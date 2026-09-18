import { useState } from 'react'
import {
  advanceReference,
  grantReferenceBC,
  isAPIError,
  type PlayerSnapshot,
  type ReferenceAdvanceMode,
  type ReferenceAdvanceResult,
} from './api'
import { type TranslationKey, type TranslationVars } from './i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type Props = {
  snapshot: PlayerSnapshot
  seatID: number
  reloadSnapshot: () => Promise<unknown>
  t: Translator
}

const stopReasonKey: Record<string, TranslationKey> = {
  requested_turns_reached: 'reference.stop.requested',
  construction_completed: 'reference.stop.construction',
  interactive_boundary: 'reference.stop.interactive',
  construction_changed: 'reference.stop.changed',
  game_completed: 'reference.stop.completed',
  hard_limit_reached: 'reference.stop.limit',
}

function profileLabel(t: Translator, profileID: string): string {
  if (profileID === 'baseline') return t('reference.profile.baseline')
  if (profileID === 'mid_tech') return t('reference.profile.midTech')
  if (profileID === 'all_tech') return t('reference.profile.allTech')
  return profileID
}

export function ReferenceLabPanel({ snapshot, seatID, reloadSnapshot, t }: Props) {
  const reference = snapshot.reference
  const [turns, setTurns] = useState(5)
  const [constructionColonyID, setConstructionColonyID] = useState(0)
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState<ReferenceAdvanceResult | null>(null)
  const [grantMessage, setGrantMessage] = useState('')
  const [error, setError] = useState('')

  if (!reference || reference.control_seat_id !== seatID) return null

  const constructionColonies = snapshot.view.colonies.filter(
    (colony) => colony.construction && colony.construction.project_kind !== 'housing',
  )
  const effectiveConstructionColonyID = constructionColonies.some((colony) => colony.id === constructionColonyID)
    ? constructionColonyID
    : (constructionColonies[0]?.id ?? 0)

  const hasDraft = (snapshot.planning_draft?.orders.length ?? 0) > 0
  const planningEligible = snapshot.view.phase === 'planning' && !snapshot.view.seat.submitted
  const advanceEligible = planningEligible && !hasDraft

  const run = async (mode: ReferenceAdvanceMode, count?: number, colonyID?: number) => {
    if (!advanceEligible || busy) return
    setBusy(true)
    setError('')
    setGrantMessage('')
    setResult(null)
    try {
      const next = await advanceReference(snapshot, seatID, {
        mode,
        ...(count !== undefined ? { turns: count } : {}),
        ...(colonyID !== undefined ? { colony_id: colonyID } : {}),
      })
      setResult(next)
      await reloadSnapshot()
    } catch (reason) {
      setError(isAPIError(reason) ? `${reason.code}: ${reason.message}` : reason instanceof Error ? reason.message : String(reason))
      await reloadSnapshot().catch(() => undefined)
    } finally {
      setBusy(false)
    }
  }

  const grantBC = async (amount: 100 | 1000 | 10000) => {
    if (!planningEligible || busy) return
    setBusy(true)
    setError('')
    setResult(null)
    setGrantMessage('')
    try {
      const next = await grantReferenceBC(snapshot, seatID, amount)
      setGrantMessage(t('reference.grantResult', { amount, balance: next.balance_bc.toFixed(1) }))
      await reloadSnapshot()
    } catch (reason) {
      setError(isAPIError(reason) ? `${reason.code}: ${reason.message}` : reason instanceof Error ? reason.message : String(reason))
      await reloadSnapshot().catch(() => undefined)
    } finally {
      setBusy(false)
    }
  }

  const normalizedTurns = Math.max(1, Math.min(reference.max_turns_per_request, Math.trunc(turns || 1)))
  const stopLabel = result ? t(stopReasonKey[result.stop_reason] ?? 'reference.stop.other') : ''

  return (
    <section className="reference-lab" aria-label={t('reference.title')}>
      <div className="reference-lab__identity">
        <div>
          <p className="eyebrow">{t('reference.eyebrow')}</p>
          <strong>{t('reference.title')}</strong>
          <small>{profileLabel(t, reference.profile_id)} · {reference.scenario_id}</small>
        </div>
        {reference.profile_id === 'all_tech' && <p className="reference-lab__note">{t('reference.allTechNote')}</p>}
      </div>

      <div className="reference-lab__group">
        <h3>{t('reference.turnGroup')}</h3>
        <div className="reference-lab__controls">
          <button type="button" disabled={!advanceEligible || busy} onClick={() => void run('turns', 1)}>
            {t('reference.advanceOne')}
          </button>
          <button type="button" disabled={!advanceEligible || busy} onClick={() => void run('turns', 5)}>
            {t('reference.advanceFive')}
          </button>
          <label className="reference-lab__turn-count">
            <span>{t('reference.turnsLabel')}</span>
            <input
              type="number"
              min={1}
              max={reference.max_turns_per_request}
              value={turns}
              disabled={!advanceEligible || busy}
              onChange={(event) => setTurns(Number(event.target.value))}
            />
          </label>
          <button type="button" disabled={!advanceEligible || busy} onClick={() => void run('turns', normalizedTurns)}>
            {t('reference.advanceMany', { turns: normalizedTurns })}
          </button>
        </div>

        {constructionColonies.length > 0 && (
          <div className="reference-lab__construction">
            <label className="reference-lab__construction-select">
              <span>{t('reference.constructionTarget')}</span>
              <select
                value={effectiveConstructionColonyID}
                disabled={!advanceEligible || busy}
                onChange={(event) => setConstructionColonyID(Number(event.target.value))}
              >
                {constructionColonies.map((colony) => (
                  <option key={colony.id} value={colony.id}>
                    {t('colonies.colony', { id: colony.id })} · {colony.construction?.project_id}
                  </option>
                ))}
              </select>
            </label>
            <button
              type="button"
              disabled={!advanceEligible || busy || effectiveConstructionColonyID === 0}
              onClick={() => void run('until_construction_complete', undefined, effectiveConstructionColonyID)}
            >
              {t('reference.untilConstruction')}
            </button>
          </div>
        )}
      </div>

      <div className="reference-lab__group">
        <div className="reference-lab__group-heading">
          <h3>{t('reference.economyGroup')}</h3>
          <span className="badge">{t('reference.balance', { balance: snapshot.view.empire.treasury.balance_bc.toFixed(1) })}</span>
        </div>
        <div className="reference-lab__controls">
          <button type="button" disabled={!planningEligible || busy} onClick={() => void grantBC(100)}>+100 BC</button>
          <button type="button" disabled={!planningEligible || busy} onClick={() => void grantBC(1000)}>+1.000 BC</button>
          <button type="button" disabled={!planningEligible || busy} onClick={() => void grantBC(10000)}>+10.000 BC</button>
        </div>
      </div>

      {!advanceEligible && hasDraft && <p className="reference-lab__state">{t('reference.draftBlocked')}</p>}
      {!planningEligible && snapshot.view.phase !== 'planning' && <p className="reference-lab__state">{t('reference.phaseBlocked')}</p>}
      {busy && <p className="reference-lab__state">{t('reference.busy')}</p>}
      {result && (
        <p className="reference-lab__state">
          {t('reference.result', { turns: result.turns_advanced, reason: stopLabel, turn: result.end_turn })}
        </p>
      )}
      {grantMessage && <p className="reference-lab__state">{grantMessage}</p>}
      {error && <p className="reference-lab__error">{error}</p>}
    </section>
  )
}
