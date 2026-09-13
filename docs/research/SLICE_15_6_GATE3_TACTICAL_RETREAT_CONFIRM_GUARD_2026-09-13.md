# Slice 15.6 Gate3 - Tactical retreat confirmation guard

Status: **IMPLEMENTED / isolated real-battle browser QA green**
Date: 2026-09-13

## Trigger
In the canonical Triangle game, Human unintentionally issued `battle.retreat` during Tactical round 2. The battle completed as `tactical_retreat` with Darlok as winner. Because Human owned no other colony system, strategic retreat resolution found no valid retreat destination and destroyed Human combat fleets 25 and 36 plus civilian fleet 32 (`no_retreat_destination`), removing ships 23, 24 and 35.

The existing `Fertig` button was not miswired: it submits `battle.end_activation`. `Rückzug` was a separate one-click destructive action with no confirmation.

## Guard
Tactical `Rückzug` is now two-stage:

1. First click opens an in-game modal and sends no Battle command.
2. The dialog warns: `Der Rückzug beendet diesen Kampf sofort. Wenn kein gültiges Rückzugsziel verfügbar ist, geht deine Flotte verloren.`
3. `Abbrechen`, Escape, or clicking the backdrop closes the dialog without mutation.
4. `Rückzug bestätigen` alone submits the existing server-authoritative `battle.retreat` command.

The safe action `Abbrechen` receives initial focus so an accidental Enter does not confirm retreat.

The dialog does not attempt to invent retreat-destination legality in React. The warning is deliberately accurate for both cases: retreat ends the Tactical battle immediately, and lack of a valid strategic retreat destination can destroy the fleet under current server rules.

## Browser acceptance
An isolated 7180 server restored a known Triangle save, then created a real Human-vs-Psilon Tactical Battle #2 through the normal Galaxy UI.

Before opening confirmation:
- Battle phase: active
- Tactical events: 1 (`round_started`)
- next command sequence: 1

After clicking `Rückzug`:
- confirmation dialog visible with warning and `Abbrechen` / `Rückzug bestätigen`
- Battle remained active
- Tactical event count stayed 1
- next command sequence stayed 1
- no result created

After `Abbrechen`:
- dialog closed
- no Battle mutation

After reopening and pressing `Rückzug bestätigen`:
- authoritative `side_retreated` event created at command sequence 1
- battle completed as `tactical_retreat`
- opposing seat became winner

This proves the destructive action cannot occur from the first button press anymore.
