import { useEffect, useMemo, useState } from 'react'
import { type BattleSide, type PlayerSnapshot, type ProtocolCommand, type ResolutionSummary } from './api'
import { TacticalBattlefield } from './TacticalBattlefield'
import { GameIcon } from './components/GameIcon'
import { StarArt } from './components/StarArt'
import { Card, EmptyState, Notice } from './components/ui'
import { type TranslationKey, type TranslationVars } from './i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

type BattleRouteViewProps = {
  snapshot: PlayerSnapshot
  battleID: number
  resolution?: ResolutionSummary
  onContinue: (summaryID?: string) => void
  onBattleCommand: (command: ProtocolCommand) => Promise<void>
  commandsDisabled: boolean
  t: Translator
}

function humanize(value: string | undefined) {
  if (!value) return '—'
  return value.replace(/[_-]+/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase())
}

export function BattleRouteView({ snapshot, battleID, resolution, onContinue, onBattleCommand, commandsDisabled, t }: BattleRouteViewProps) {
  const [tacticalShellOpen, setTacticalShellOpen] = useState(false)
  const battle = snapshot.battles.find((candidate) => candidate.spec.id === battleID)
  const battleSummary = resolution?.kind === 'battle_completed' && resolution.battle?.battle_id === battleID ? resolution.battle : undefined
  const strategic = snapshot.decision?.strategic
  const publicEmpires = snapshot.decision?.public_empires ?? []
  const ships = strategic?.ships ?? []
  const fleets = strategic?.fleets ?? []
  const systemID = battle?.spec.system_id ?? battleSummary?.system_id
  const system = useMemo(() => systemID ? strategic?.galaxy.systems.find((candidate) => candidate.id === systemID) : undefined, [systemID, strategic])

  useEffect(() => setTacticalShellOpen(false), [battleID])

  if (!battle && !battleSummary) {
    return <EmptyState title={t('battle.notFoundTitle')} body={t('battle.notFoundBody', { id: battleID })} />
  }

  const empireName = (empireID: number | undefined) => {
    if (!empireID) return t('battle.unknownEmpire')
    if (snapshot.decision?.empire.id === empireID) return snapshot.decision.empire.name
    return publicEmpires.find((empire) => empire.id === empireID)?.name ?? t('battle.empireFallback', { id: empireID })
  }
  const shipName = (shipID: number) => ships.find((ship) => ship.id === shipID)?.name ?? t('battle.shipFallback', { id: shipID })
  const fleetName = (fleetID: number) => {
    const fleet = fleets.find((candidate) => candidate.id === fleetID)
    return fleet ? t('battle.fleetVisible', { id: fleet.id, role: humanize(fleet.role) }) : t('battle.fleetFallback', { id: fleetID })
  }
  const colonyName = (colonyID: number) => {
    const own = snapshot.decision?.colonies.find((colony) => colony.id === colonyID)
    if (own) return t('battle.colonyOwned', { id: own.id })
    const contact = strategic?.contacts?.find((candidate) => candidate.kind === 'colony' && candidate.colony_id === colonyID)
    return contact?.colony_id ? t('battle.colonyVisible', { id: contact.colony_id }) : t('battle.colonyFallback', { id: colonyID })
  }

  const renderSide = (side: BattleSide | undefined, empireID: number | undefined, label: string) => (
    <Card className="battle-side-card" as="section">
      <p className="eyebrow">{label}</p>
      <h2>{empireName(side?.empire_id ?? empireID)}</h2>
      {side ? (
        <>
          <div className="battle-side-metrics">
            <span><GameIcon name="fleets" /><strong>{side.combat_fleet_ids.length}</strong>{t('battle.combatFleets')}</span>
            <span><GameIcon name="ship" /><strong>{side.ship_ids.length}</strong>{t('battle.combatShips')}</span>
          </div>
          {side.combat_fleet_ids.length > 0 && <div className="battle-chip-list">{side.combat_fleet_ids.map((id) => <span className="badge" key={id}>{fleetName(id)}</span>)}</div>}
          {side.ship_ids.length > 0 && <ul className="battle-ship-list">{side.ship_ids.map((id) => <li key={id}><GameIcon name="ship" />{shipName(id)}</li>)}</ul>}
        </>
      ) : <p className="muted">{t('battle.summaryOnlySide')}</p>}
    </Card>
  )

  const completed = Boolean(battleSummary) || battle?.phase === 'completed'
  const tacticalSupported = Boolean(battle?.spec.tactical) && !battle?.spec.tactical_unsupported_reason
  const winner = battleSummary?.winner_empire_id
    ? empireName(battleSummary.winner_empire_id)
    : battle?.result?.winner_seat
      ? t('battle.seatFallback', { id: battle.result.winner_seat })
      : t('battle.unknownWinner')
  const destroyed = battleSummary?.destroyed_ship_ids ?? battle?.result?.destroyed_ship_ids ?? []
  const survivors = battleSummary?.surviving_ship_ids ?? []
  const colonyIDs = battle?.spec.defender_colony_ids ?? battleSummary?.defender_colony_ids ?? []
  const strategicTurn = battle?.spec.strategic_turn ?? resolution?.turn ?? snapshot.view.turn
  const starSeed = battle?.spec.seed ?? battleID

  return (
    <main className="battle-route-view">
      <header className="battle-route-header">
        <div>
          <p className="eyebrow">{completed ? t('battle.completedEyebrow') : t('battle.encounterEyebrow')}</p>
          <h1>{system?.name ?? t('battle.systemFallback', { id: systemID ?? 0 })}</h1>
          <p className="muted">{t('battle.turnContext', { turn: strategicTurn, id: battleID })}</p>
        </div>
        {system && <div className="battle-star-art" aria-hidden="true"><StarArt spectralClass={system.spectral_class} seed={starSeed} /></div>}
      </header>

      <div className="battle-sides-grid">
        {renderSide(battle?.spec.attacker, battleSummary?.attacker_empire_id, t('battle.attacker'))}
        <div className="battle-versus" aria-hidden="true">VS</div>
        {renderSide(battle?.spec.defender, battleSummary?.defender_empire_id, t('battle.defender'))}
      </div>

      {colonyIDs.length > 0 && (
        <Card className="battle-colony-context" as="section">
          <p className="eyebrow">{t('battle.colonyContext')}</p>
          <div className="battle-chip-list">{colonyIDs.map((id) => <span className="badge" key={id}>{colonyName(id)}</span>)}</div>
        </Card>
      )}

      {battle && !completed && (
        <Card className="battle-entry-card" as="section">
          <h2>{t('battle.entryTitle')}</h2>
          {tacticalSupported ? (
            <>
              <p>{t('battle.entryBody')}</p>
              {!tacticalShellOpen ? (
                <button type="button" className="button-primary" onClick={() => setTacticalShellOpen(true)}><GameIcon name="ship" />{t('battle.enterTactical')}</button>
              ) : battle.tactical ? (
                <TacticalBattlefield
                  battle={battle}
                  ownSeatID={snapshot.view.seat.seat.id}
                  shipName={shipName}
                  empireName={empireName}
                  commandsDisabled={commandsDisabled}
                  onCommand={onBattleCommand}
                  onBack={() => setTacticalShellOpen(false)}
                  t={t}
                />
              ) : (
                <Notice title={t('battle.tacticalUnavailableTitle')} tone="warning">{t('battle.noTacticalSpec')}</Notice>
              )}
            </>
          ) : (
            <Notice title={t('battle.tacticalUnavailableTitle')} tone="warning">{t('battle.tacticalUnavailable', { reason: battle.spec.tactical_unsupported_reason || t('battle.noTacticalSpec') })}</Notice>
          )}
          <p className="battle-blocking-note"><GameIcon name="info" />{t('battle.blocking')}</p>
        </Card>
      )}

      {completed && (
        <Card className="battle-return-card" as="section">
          <p className="eyebrow">{t('battle.returnEyebrow')}</p>
          <h2>{t('battle.returnTitle')}</h2>
          {battleSummary ? (
            <>
              <dl className="battle-result-facts">
                <div><dt>{t('battle.winner')}</dt><dd>{winner}</dd></div>
                <div><dt>{t('battle.outcome')}</dt><dd>{humanize(battleSummary.outcome)}</dd></div>
                <div><dt>{t('battle.destroyed')}</dt><dd>{destroyed.length}</dd></div>
                <div><dt>{t('battle.surviving')}</dt><dd>{survivors.length}</dd></div>
              </dl>
              {destroyed.length > 0 && <div className="battle-result-list"><strong>{t('battle.destroyedShips')}</strong><div className="battle-chip-list">{destroyed.map((id) => <span className="badge danger" key={id}>{shipName(id)}</span>)}</div></div>}
              {survivors.length > 0 && <div className="battle-result-list"><strong>{t('battle.survivingShips')}</strong><div className="battle-chip-list">{survivors.map((id) => <span className="badge" key={id}>{shipName(id)}</span>)}</div></div>}
              {colonyIDs.length > 0 && <div className="battle-result-list"><strong>{t('battle.colonyConsequence')}</strong><div className="battle-chip-list">{colonyIDs.map((id) => <span className="badge" key={id}>{colonyName(id)}</span>)}</div></div>}
              <button type="button" className="button-primary" onClick={() => onContinue(resolution?.id)}>{t('battle.continue')}</button>
            </>
          ) : <Notice title={t('battle.waitingTitle')}>{t('battle.waitingForSummary')}</Notice>}
        </Card>
      )}
    </main>
  )
}
