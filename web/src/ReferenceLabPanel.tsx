import { useMemo, useState } from 'react'
import { advanceReference, isAPIError, type PlayerSnapshot, type ReferenceAdvanceMode, type ReferenceAdvanceResult } from './api'
import { type TranslationKey, type TranslationVars } from './i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type Props = {
  snapshot: PlayerSnapshot
  seatID: number
  selectedColonyID?: number
  reloadSnapshot: () => Promise<PlayerSnapshot | undefined>
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

export function ReferenceLabPanel({ snapshot, seatID, selectedColonyID, reloadSnapshot, t }: Props) {
  const reference = snapshot.reference
  const [turns, setTurns] = useState(5)
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState<ReferenceAdvanceResult | null>(null)
  const [error, setError] = useState('')

  const selectedColony = useMemo(
    () => snapshot.view.colonies.find((colony) => colony.id === selectedColonyID),
    [snapshot.view.colonies, selectedColonyID],
  )

  if (!reference || reference.control_seat_id !== seatID) return null

  const hasDraft = (snapshot.planning_draft?.orders.length ?? 0) > 0
  const eligible = snapshot.view.phase === 'planning' && !snapshot.view.seat.submitted && !hasDraft
  const construction = selectedColony?.construction

  const run = async (mode: ReferenceAdvanceMode, count?: number, colonyID?: number) => {
    if (!eligible || busy) return
    setBusy(true)
    setError('')
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

  const normalizedTurns = Math.max(1, Math.min(reference.max_turns_per_request, Math.trunc(turns || 1)))
  const stopLabel = result ? t(stopReasonKey[result.stop_reason] ?? 'reference.stop.other') : ''

  return (
    <section className="reference-lab" aria-label={t('reference.title')}>
      <div className="reference-lab__identity">
        <div>
          <p className="eyebrow">{t('reference.eyebrow')}</p>
          <strong>{t('reference.title')} · {profileLabel(t, reference.profile_id)}</strong>
          <small>{reference.scenario_id}</small>
        </div>
        {reference.profile_id === 'all_tech' && <p className="reference-lab__note">{t('reference.allTechNote')}</p>}
      </div>

      <div className="reference-lab__controls">
        <button type="button" disabled={!eligible || busy} onClick={() => void run('turns', 1)}>
          {t('reference.advanceOne')}
        </button>
        <label className="reference-lab__turn-count">
          <span>{t('reference.turnsLabel')}</span>
          <input
            type="number"
            min={1}
            max={reference.max_turns_per_request}
            value={turns}
            disabled={!eligible || busy}
            onChange={(event) => setTurns(Number(event.target.value))}
          />
        </label>
        <button type="button" disabled={!eligible || busy} onClick={() => void run('turns', normalizedTurns)}>
          {t('reference.advanceMany', { turns: normalizedTurns })}
        </button>
        {construction && selectedColony && (
          <button
            type="button"
            disabled={!eligible || busy}
            onClick={() => void run('until_construction_complete', undefined, selectedColony.id)}
          >
            {t('reference.untilConstruction')}
          </button>
        )}
      </div>

      {!eligible && hasDraft && <p className="reference-lab__state">{t('reference.draftBlocked')}</p>}
      {!eligible && !hasDraft && snapshot.view.phase !== 'planning' && <p className="reference-lab__state">{t('reference.phaseBlocked')}</p>}
      {busy && <p className="reference-lab__state">{t('reference.busy')}</p>}
      {result && (
        <p className="reference-lab__state">
          {t('reference.result', { turns: result.turns_advanced, reason: stopLabel, turn: result.end_turn })}
        </p>
      )}
      {error && <p className="reference-lab__error">{error}</p>}
    </section>
  )
}
