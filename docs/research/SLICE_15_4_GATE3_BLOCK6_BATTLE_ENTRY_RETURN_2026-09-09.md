# Slice 15.4 Gate 3 Block 6 - Battle Entry / Return Route + Shell

Date: 2026-09-09

## Goal

Implement the frozen Slice-15.4 Encounter hand-off contract without absorbing Slice 15.5 Tactical mechanics. A human participant must be routed into a stable Battle shell, must not be able to pretend strategic play continued while the encounter is unresolved, and must receive a one-time participant-safe authoritative result before returning to strategy.

## Stable route

The browser now accepts and emits:

`#/game/<gameID>/battle/<battleID>`

`battle` is a game section for routing purposes but is not exposed as a normal primary-navigation destination.

## Battle Entry

For a participant-visible active `BattleView`, the Battle route shows only projected/public information already present in the player snapshot:

- system and spectral star context;
- attacker/defender public empire identity;
- visible combat-fleet and ship IDs/names;
- defender-colony context where projected;
- strategic turn and Battle ID;
- Tactical availability from `BattleSpec.tactical` / `tactical_unsupported_reason`.

When a supported Tactical spec exists, `Enter Tactical Combat` opens the Slice-15.4 Tactical shell. The shell explicitly stops at the hand-off boundary: movement, firing and activation controls remain Slice 15.5. When Tactical is unsupported, the server reason is shown and no Tactical/autoresolve CTA is invented.

## Strategic-continuation lock

An unresolved participant battle is blocking. The browser automatically redirects a strategic route to its Battle route and disables primary side/bottom navigation plus End Turn while the Battle remains open. Research overlay activation is also suppressed. Main-menu/home lifecycle actions remain available, but the browser cannot navigate to a strategic section and imply the unresolved encounter was skipped.

## Battle Return

Block 5 already established participant-safe event-derived `battle_completed` summaries. Block 6 consumes that authority directly and presents:

- winner and outcome;
- destroyed visible ship IDs/names;
- surviving visible ship IDs/names;
- defender-colony context where present;
- a presentation-only `Continue to strategy` acknowledgement.

The acknowledgement is stored through the existing per-game/per-seat resolution-ack mechanism and then returns to the Galaxy route. Re-visiting an already acknowledged Battle result redirects to Galaxy.

### Final-Battle lifecycle hardening

The server appends the authoritative `battle_completed` event before strategic continuation, but after the final encounter it may clear `s.battles` as the turn resumes. Therefore the Return surface does **not** depend on the old live `BattleView` being retained. An unacknowledged `battle_completed` resolution itself becomes a blocking Battle route and can render a reduced attacker/defender context from its authoritative empire IDs even when `snapshot.battles` is already empty.

This matches the real session lifecycle and prevents a race where the completed Battle shell would disappear before the player saw the result.

If the final Battle also completes the whole game, the unacknowledged Battle return is rendered before the dedicated Victory surface. After Continue to strategy acknowledges the Battle summary, the normal completed-game result surface becomes eligible, preventing a locked result screen with no Battle acknowledgement path.

## Browser QA

A temporary local mock was built from a real current player snapshot and removed after QA. It exercised three cases through the production React build:

1. **Supported active Battle**
   - starting on `/galaxy` auto-routed to `/battle/9001`;
   - system/sides/fleets/ships rendered from participant-safe snapshot data;
   - side nav, bottom nav and End Turn were disabled;
   - Tactical CTA opened only the Slice-15.4 placeholder shell.
2. **Unsupported active Battle**
   - auto-routed to Battle;
   - displayed the server unsupported reason;
   - rendered no Tactical/autoresolve primary CTA.
3. **Completed final Battle with `battles=[]`**
   - unacknowledged `battle_completed` forced `/battle/9001` even though the live BattleView had already been released;
   - winner/outcome/destroyed/surviving context rendered from the ResolutionSummary;
   - strategic navigation stayed locked;
   - Continue stored `event-999` in the existing resolution acknowledgement and returned to `/galaxy`;
   - a later direct revisit of the acknowledged Battle route redirected back to Galaxy.

## Boundary to Slice 15.5

No Tactical movement, beam firing, weapon selection, activation handling, autoresolve or React-owned combat rule was added. Slice 15.5 remains the owner of interactive battlefield mechanics and will consume the already authoritative Tactical spec/decision catalog inside this shell.
