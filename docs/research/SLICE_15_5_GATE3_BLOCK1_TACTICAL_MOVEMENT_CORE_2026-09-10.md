# Slice 15.5 Gate 3 Block 1 - Tactical movement / facing / 2v2 authority core

Date: 2026-09-10

## Goal

Implement the first frozen Gate-2 server-authority block before any battlefield UI: Tactical runtime positions, 16-way facing, original movement cost, movement-point budgets, deterministic legal-move projection, `battle.move_ship`, 1v1-2v2 BattleSession breadth and strategic Encounter metadata capable of producing those supported Tactical specs.

## Runtime state

`TacticalShipState` now owns mutable Tactical position and activation state:

- X / Y integer world coordinate;
- Facing `0..15`;
- MovementCurrent / MovementMax;
- ActivationComplete;
- existing Armor / Structure / weapon readiness / destroyed state.

The immutable `TacticalShipSpec` now includes initial Facing and TurningMode. Initial movement current/max is derived from the existing authoritative CurrentCombatSpeed.

## Original movement economics

Movement uses the Gate-1 original-evidence formula frozen at Gate 2:

`translation = ceil(sqrt(dx^2 + dy^2))`

`facingSteps = shortestCircularDifference(currentFacing, destinationFacing, 16)`

`moveCost = translation + facingSteps * turnCostPerStep`

Facing is quantized into sixteen sectors. Turning cost is:

- 2 normal;
- 1 Inertial Stabilizer;
- 0 Inertial Nullifier.

The current normalized strategic ShipDesignSpec does not yet carry ship specials, so generated strategic Tactical metadata currently uses normal turning mode. The Battle core is already structured for Stabilizer/Nullifier once those specials are normalized.

## Effectively unbounded plane

No gameplay board width/height was added. Coordinates are ordinary Tactical world coordinates. To guard integer/pathological input only, the server rejects coordinates beyond a large technical safety envelope of +/-1,000,000. This is not projected as an arena edge and is not normal movement gameplay.

Legal moves remain finite because MovementCurrent is finite. Before any distance work, obviously over-budget axis deltas are rejected. Occupied live-ship coordinates are not legal destinations.

## Authoritative command

New command:

`battle.move_ship`

Trusted client payload:

- ship_id;
- destination x;
- destination y.

The server validates:

- active Battle / correct battle sequence;
- controlling seat;
- active ship identity;
- ship alive and activation unfinished;
- destination differs from current location (no hidden rotate-only command);
- safety envelope;
- destination unoccupied;
- exact authoritative movement cost within remaining budget.

Success atomically changes X/Y/Facing/MovementCurrent in the prepared Tactical runtime and appends deterministic `ship_moved` event data. Movement consumes no RNG.

Rejected moves remain zero-mutation because all changes occur on the existing prepared-runtime clone and commit only after successful validation.

## Legal-move projection

`TacticalView` now projects deterministic `legal_moves[]` for the active ship. Each move includes:

- x / y;
- move_cost;
- resulting_facing;
- movement_remaining_after.

The set is generated server-side from the remaining movement budget, filtered through occupancy/cost legality and sorted deterministically by cost, then Y, then X. `can_end_activation` is also projected while the Battle is active.

Completed Battles deliberately project no new legal moves/actions.

## Fire / activation integration

Laser range now uses mutable runtime X/Y rather than the immutable deployment coordinates, so movement immediately changes authoritative range calculations.

`battle.end_activation` now records ActivationComplete and skips already-completed/destroyed ships when choosing the next active combatant. At a new round, all live ships reset:

- ActivationComplete = false;
- MovementCurrent = MovementMax;
- supported weapon readiness = true.

Terminal Tactical victory now reports every ship already marked destroyed, not only the final killing target, which is required for multi-ship strategic reconciliation.

## 1v1-2v2 strategic metadata

`game.tacticalMetadataForEncounter` now supports one or two combat Ships per strategic side while preserving explicit unsupported boundaries:

- no civilian Fleet context;
- no Colony/planet-defense context;
- Frigate / Electronic Computer / Titanium Armor / no Shield / standard fuel baseline;
- Nuclear or Fusion drive (combat speed derived from the actual drive rule);
- zero or one slot-0 standard Laser per combatant;
- at least one supported Laser somewhere in the Battle, preventing an all-unarmed Tactical deadlock;
- more than two ships per side remains explicit unsupported.

Single-ship deployment keeps the historical Slice-07 fixture coordinates for regression stability. Multi-ship deployment is relative to an encounter origin:

- attacker X=10, Facing=0;
- defender X=14, Facing=8;
- two-ship wings use Y=9 and Y=11.

These coordinates are initial placement only; they are not board edges.

## Regression coverage

New Battle-core regressions cover:

- straight movement;
- normal quarter-turn cost;
- Inertial Stabilizer and Nullifier turning costs;
- diagonal facing quantization;
- legal-move projection;
- occupied-square exclusion;
- successful move updates to X/Y/Facing/remaining movement;
- movement does not consume RNG;
- event / command sequence advancement;
- zero-mutation rejection for wrong seat, stale/wrong sequence, wrong active ship, same coordinate, occupied destination, over-budget destination and technical-envelope violation;
- 2v2 initiative progression and round reset;
- explicit rejection above two ships per side.

New Game metadata regressions cover:

- supported 2v2 Tactical generation with unique open-field deployment and correct facings/speeds;
- explicit all-unarmed unsupported result;
- explicit 3v1 unsupported result.

The former Slice-07 unsupported-manual-lifecycle regression was updated to use a genuinely unsupported two-cannon mount; arming both sides with one supported Laser is now intentionally legal under Slice 15.5.

## Validation

Passed after implementation:

- `go test ./internal/battle`;
- `go test ./internal/game ./internal/session ./internal/battle`;
- `go test ./...`;
- `go vet ./...`;
- `git diff --check`.

## Boundary to next block

This block does not yet expose the full participant-safe Tactical scan/action projection through App/API and does not implement the browser battlefield. Those are subsequent Gate-3 blocks. React still owns no movement legality, movement cost, facing result, RNG or combat result.
