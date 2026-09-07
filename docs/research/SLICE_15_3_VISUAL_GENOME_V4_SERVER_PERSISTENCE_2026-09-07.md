# Slice 15.3 Gate 1 - Visual Genome v4 server persistence

Date: **2026-09-07**

Status: **implemented and regression-covered**.

## Goal

A ship visual selected with **Use this design** must become authoritative persistent game data rather than browser-local state. The same accepted shape must survive:

- page/snapshot reload;
- live save/export and restore/import;
- later generator implementation changes;
- later gameplay edits to the same ship design;
- construction of individual ships.

Visual identity remains presentation-only in Slice 15.3. It does not alter combat values, production cost, hull legality or equipment authority.

## Persist the resolved genome, not only the seed

MOOX now persists the complete resolved Visual Genome v4:

- version;
- visual hull ID;
- Style DNA;
- morphology ID;
- source seed;
- length/beam;
- station count and half-widths;
- concave notch depths;
- engine/detail counts;
- canonical half-cutouts;
- canonical half-primitives.

This is deliberate. A seed alone would depend on the exact implementation of a future generator. Persisting resolved geometry means an accepted v4 ship remains the same even if v5/v6 changes the random grammar.

## Core data contract

New authoritative core types live in:

`internal/core/ship_visual.go`

`ShipVisualGenomeVersion = 4`

The server validates:

- exact supported genome version;
- the seven symmetric v4 morphology families;
- supported Style-DNA IDs;
- supported visual hull IDs;
- finite/bounded dimensions;
- exact station-array lengths;
- valid half-hull nose termination;
- bounded engine/detail counts;
- bounded cutout/primitive counts and geometry;
- v4 primitive kinds (`wedge`, `spike`, `pod`).

The removed default `asymmetric` morphology is rejected by the v4 persistence contract.

## ShipDesign persistence

`core.ShipDesign` now has optional fields:

- `visual_revision`;
- `visual_genome`.

The existing `revision` remains the **gameplay/equipment revision**.

This separation is important. Clicking a different ship picture must not invalidate a colony construction order that references gameplay design revision N.

Therefore:

- equipment/spec save -> increments `ShipDesign.Revision`;
- visual save -> increments `ShipDesign.VisualRevision` only.

A later gameplay design edit preserves the stored visual revision/genome.

## Immediate server command

New command:

`empire.set_military_design_visual`

Payload:

- `design_id`;
- full `visual_genome`.

It is handled as an immediate strategic presentation mutation rather than a turn-planning order. The player does not need to end the turn to save the chosen appearance.

The resolver verifies:

- design exists;
- design belongs to the requesting empire;
- genome passes v4 validation;
- visual revision does not overflow.

The stored genome is deep-copied so caller-owned arrays cannot mutate authoritative state after submission.

## Submitted-seat boundary

A player is allowed to change a ship picture after already pressing **End Turn** while the game remains in Planning waiting on another seat.

An initial regression test exposed an important revision-boundary issue: an immediate visual mutation increments the global session revision, while submitted gameplay batches originally retained the previous `base_revision`. That made a live snapshot invalid even though gameplay intent had not changed.

The final contract is:

- the visual command changes presentation only;
- existing submitted gameplay batches remain semantically unchanged;
- after the visual-only revision increment, submitted batches for the same turn have only their `base_revision` advanced to the new authoritative revision;
- command payloads/order remain unchanged.

This keeps the live-save planning boundary consistent without reopening or changing the player's gameplay submission.

## Built ships freeze visual identity

`core.Ship` now has optional:

- `source_visual_revision`;
- `visual_genome`.

When a military ship completes, MOOX deep-copies the design's current accepted Visual Genome and VisualRevision into the concrete ship.

Result:

- newly built ship A keeps the appearance that design had when A was built;
- changing the design picture later affects future ships/design presentation, not ship A;
- fleets can render each concrete ship from its own persisted visual snapshot.

This mirrors the existing `source_design_revision` gameplay pattern without coupling the two revision streams.

## Snapshot / save compatibility

The fields are optional JSON additions to existing core state, so `StateSchemaVersion` remains **23**.

Older state/saves without visual fields continue to validate and load. Their web renderer uses the existing deterministic seed fallback.

New state with Visual Genome v4 round-trips through:

- `core.MarshalState` / `core.UnmarshalState`;
- GameSession live snapshot serialization;
- app/Host `ExportLiveSnapshot` / fresh-host `ImportLiveSnapshot`.

Unsupported or malformed new genomes are rejected rather than silently rewritten.

## Web wire adapter

The browser generator uses camelCase TypeScript fields while Go/JSON uses snake_case.

`web/src/api.ts` now provides explicit:

- `ShipVisualGenomeWire`;
- `encodeShipVisualGenome(...)`;
- `decodeShipVisualGenome(...)`;
- `submitMilitaryDesignVisual(...)`.

This avoids leaking frontend naming conventions into the authoritative protocol and makes version boundaries explicit.

## Shipbuilder behavior

**Use this design** now submits the current candidate to the immediate-command endpoint for the existing baseline military design.

On success:

- the kept-design panel shows the accepted genome;
- the UI reports server-side VisualRevision;
- a later snapshot/page reload restores the kept design from `ship_designs[].visual_genome`.

The primary random-generation flow remains unchanged.

Until Slice 17 broadens authoritative hull mechanics, the selected visual hull is presentation-only and is intentionally **not required to equal** `ShipDesign.Spec.HullID` (currently narrow Frigate gameplay support).

## Strategic rendering

Real ships in system fleet rosters and the Fleets page now prefer `ship.visual_genome` from the authoritative snapshot.

Old ships without a persisted genome retain the deterministic design/revision/picture seed fallback.

## Regression evidence

### Core

- valid symmetric v4 genome accepted;
- removed `asymmetric` morphology rejected;
- Visual Genome round-trips through State Schema 23 exactly.

### Game

- visual save increments only VisualRevision;
- gameplay Revision stays unchanged;
- authoritative genome deep-copy isolation verified;
- second visual save increments VisualRevision again;
- later gameplay design revision preserves VisualRevision/genome;
- completed ship receives exact genome + source visual revision;
- later design visual mutation does not alter already built ship.

### Session

- visual immediate command works while own seat is already submitted and another seat is pending;
- session revision increments;
- gameplay submission remains submitted;
- live snapshot export/import restores exact visual genome.

### Host

- fresh generated game accepts visual immediate command;
- PlayerSnapshot exposes VisualRevision/genome;
- gameplay Revision is unchanged;
- Host ExportLiveSnapshot -> fresh Host ImportLiveSnapshot -> PlayerSnapshot returns the exact same visual genome.

## Migration / future contract

For future Visual Genome versions:

1. Never regenerate an accepted v4 design merely because generator code changed.
2. Continue rendering persisted resolved v4 geometry as long as v4 renderer support exists.
3. If a future migration is required, make it explicit/versioned and preserve visual intent; do not silently substitute a random seed.
4. Legacy designs with no genome remain on deterministic fallback until the player chooses/saves a visual or a deliberate migration is introduced.

## Next work

With persistence established, the next Slice 15.3 work can move away from ship-generator internals:

- optionally tune the seven symmetric morphology families from player review;
- define race/faction probability weighting without restricting full-random choice;
- continue the SVG icon/visual language for Galaxy, Colony, Fleet, Research, Diplomacy and Espionage;
- prepare Gate-2 visual pipeline freeze.
## Isolated browser end-to-end QA

To avoid destroying the existing in-memory review `game-1` on port 7171, the current uncommitted backend was started separately on loopback port **7172** and a fresh `game-1` was created there through the normal New Game UI.

The normal Shipbuilder/HTTP flow was then exercised without direct server mutation:

1. Initial baseline design reported gameplay `revision = 1`, `visual_revision = 0`, no visual genome.
2. The initial Scout/Manta/Spear candidate was accepted with **Use this design**.
3. UI reported `Auf dem Server gespeichert · Visual-Revision 1`.
4. A fresh PlayerSnapshot reported:
   - gameplay revision still `1`;
   - visual revision `1`;
   - genome v4;
   - exact source seed `shipbuilder:game-1:scout:full-random:1`;
   - exact Manta/Spear identity.
5. Reloading the Shipbuilder reconstructed the kept-card SVG with the exact same body path as the accepted candidate.
6. Clicking the ship generated a new Chevron candidate while the kept Manta remained unchanged.
7. Accepting the Chevron produced server `visual_revision = 2`, retained gameplay `revision = 1`, and stored seed `shipbuilder:game-1:scout:full-random:2`.
8. The kept-card SVG matched the newly accepted candidate exactly and no save error was shown.

This proves the actual browser -> HTTP immediate command -> authoritative snapshot -> browser decode/render loop, in addition to the Core/Game/Session/Host unit/integration tests.