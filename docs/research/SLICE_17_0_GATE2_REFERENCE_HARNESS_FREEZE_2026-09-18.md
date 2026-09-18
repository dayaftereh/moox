# Slice 17.0 Gate 2 - Triangle Reference Harness Contract Freeze

Date: 2026-09-18
Slice: 17.0
Gate: 2 - contract freeze
Baseline: `b00c798` (`docs(slice): make triangle labs durable`)
Session: `ses-20260918T152048-00240c450344`

## Purpose

Freeze the durable development-only Triangle reference harness before implementation.

This Gate does not implement the harness. It freezes the scenario family, deterministic technology profiles, normal-turn advancement controls, built-Ship fixture rules, authorization boundary, HTTP/HMI contract and persistence behavior that Gate 3 must implement.

## 1. Durable Triangle scenario family

The public reference product is one durable Triangle family, not a set of disposable Designer worlds.

| Profile | Public game ID | Scenario identity | Seed | Turn-1 controller layout |
| --- | --- | --- | --- | --- |
| Baseline | `game-triangle-2pc` | `triangle-2pc-v1` | `0x800A` | Seat 1 Human/local human; Seat 2 Darlok/builtin AI; Seat 3 Psilon/builtin AI |
| Mid-Tech | `game-triangle-mid-tech` | `triangle-2pc-mid-tech-v1` | `0x800A` | identical |
| All-Tech | `game-triangle-all-tech` | `triangle-2pc-all-tech-v1` | `0x800A` | identical |

All three variants keep the existing deterministic three-system 2-pc Triangle topology.

### Binding comparability invariant

Mid-Tech and All-Tech are derived from the ordinary baseline Triangle initial state.

The profile is applied after the normal Triangle game has created its races, technologies, systems, starting assets and economy. Applying a profile:

- allocates no new Core IDs;
- consumes no RNG;
- does not replace the starting ships/fleets;
- does not alter systems, planets, colonies, population, treasury or production;
- does not alter controllers;
- changes only the Empire technology-ownership fields defined below;
- is followed by ordinary `GameState.Validate()`.

After stripping the two technology-ownership lists, the Turn-1 Core state of all three variants must be structurally identical.

There are no durable `game-designer-mid`, `game-designer-high` or generic `game-construction-lab` worlds.

Designer and Construction Lab controls attach to the Triangle family.

## 2. Technology-profile contract

Source of truth:

- `data/rulesets/moo2-1.31/technologies.json`
- current file SHA-256: `d6f5b76c10e4248ed8ccc7515a78112e821ba694d00f33e9f9cd92b2d9d0e0fc`

The profiles are QA bootstrap states, not original MOO2 New Game technology levels.

They deliberately grant every concrete application in an included field so the reference lab can expose breadth deterministically even for races that would normally choose one application from a field.

This exception exists only during reference bootstrap. All later design/build/research/gameplay authority remains normal production code.

### 2.1 Baseline

Baseline remains the existing `NewReferenceTriangleGame` Average-start technology state with no extra grant.

### 2.2 Mid-Tech

Stable profile ID: `mid_tech`.

Definition:

1. include the always-known field `0`;
2. include every normalized research field in `1..74` whose `research_cost_rp <= 1150`;
3. grant every concrete technology application whose `tech_field_id` is one of those fields;
4. set `KnownTechnologyFieldIDs` and `KnownTechnologyIDs` to the exact sorted profile sets;
5. active `Research` remains nil at bootstrap;
6. do not create Hyper-Advanced levels.

Frozen Mid-Tech field IDs:

`0,1,2,3,4,5,7,9,10,15,16,18,19,20,21,22,23,28,29,31,34,35,36,41,43,45,47,54,55,56,57,60,62,63,66,73`

Count: **36 fields**.

Sorted-field fingerprint, SHA-256 of the comma-separated decimal list:

`5933dd42d93850acea1493767d23a988be40234c4f35097ccbd4d0d75dbbfc6b`

Current Mid-Tech technology count: **94**.

Sorted-technology-ID fingerprint, SHA-256 of the comma-separated decimal list:

`90a821a5c4890c5ab3098990acbe86ad49dee74040f7ed167985540955f82d7c`

The threshold is predecessor-closed in the committed 1.31 research graph; no included field depends on a predecessor above the 1150-RP boundary.

Representative Ship Designer coverage at this boundary includes:

- Titan Construction;
- Fusion and Ion Drive;
- Electronic, Optronic and Positronic Computer;
- Titanium and Tritanium Armor;
- Class I and Class III Shield;
- Standard, Deuterium and Iridium Fuel Cells;
- Laser Cannon.

It intentionally does not include Doom Star Construction or the late-game mandatory-system tiers above the threshold.

### 2.3 All-Tech

Stable profile ID: `all_tech`.

Definition:

1. `KnownTechnologyFieldIDs` is exactly `0..74`;
2. `KnownTechnologyIDs` contains every normalized concrete technology with `tech_field_id` in `0..74`;
3. active `Research` remains nil at bootstrap;
4. Hyper-Advanced fields `75..82` are not marked permanently known;
5. no Hyper-Advanced completed levels are fabricated.

Field count: **75**.

Sorted-field fingerprint:

`0354bfea6fb292379fc7f5ffcf8754effad7496722f0c0152a435f694ccabff6`

Technology count: **202**.

The exact current set is technology IDs **1..203 except 125**.

Sorted-technology-ID fingerprint:

`ce320d510044b036a9e2920f817bcacdea5bc5d587e0787237454a7787a5efca`

Technology 125, `phase_shifter`, is excluded because the normalized 1.31 dataset currently has `tech_field_id=-1`. A later slice may only include it after explicitly freezing a legitimate acquisition/unlock contract.

"All-Tech" therefore means **all normalized research-addressable concrete technologies**, not "every gameplay mechanic is implemented". Unsupported Designer/Tactical consumers remain locked or explicitly unsupported.

This definition is intentionally broader and more durable than "all technologies currently exposed by the Designer": when Slices 17.1-17.4 add normal consumers for already-normalized technologies, the All-Tech Triangle should not require profile rewrites.

## 3. Reference ownership and immutable metadata

Every registered reference Triangle carries host-level immutable metadata:

- `scenario_id`;
- `profile_id`: `baseline`, `mid_tech` or `all_tech`;
- `control_seat_id = 1`;
- reference-control capability values.

Presence of this metadata is the per-game authorization token.

Game-ID strings are descriptive only and never authorize a development action.

Seat 1 remains the only reference-control owner. Seats 2 and 3 remain normal built-in AI controllers.

## 4. Strict development-only authorization

Reference controls require both:

1. the server process was explicitly started with the existing `-reference-games` development capability; and
2. the target hosted game carries immutable reference metadata from trusted reference registration.

The `demo-fixture` switch does not enable reference controls.

There is no public command/API that converts an ordinary game into a reference game.

When `-reference-games` is disabled, reference-control HTTP routes are not registered.

When the global capability is enabled but the target game is not a registered reference scenario, the server rejects the action.

A game named `game-triangle-all-tech` that arrived through normal game creation or snapshot import is still not authorized.

## 5. Projected reference capability

Reference-aware application projections add an optional top-level `reference` object to Game Summary and Player Snapshot.

Absence means ordinary game.

Frozen logical shape:

```text
reference:
  scenario_id
  profile_id
  control_seat_id
  max_turns_per_request = 25
  max_construction_turns = 512
```

Only the control seat receives actionable controls in the HMI. Observer/other-seat projections may show the scenario/profile identity but must not receive an actionable control state.

## 6. Normal-turn advancement contract

Reference advancement is an orchestration layer around the existing production Host/GameSession path.

It must never call `GameSession.CompleteTurn()` directly to skip phases.

For each automated turn it:

1. requires the reference control seat to be at Planning and not already submitted;
2. requires no non-empty saved Planning Draft for the control seat;
3. submits an ordinary zero-command `CommandBatch` for Seat 1 through normal `Host.SubmitTurn`;
4. allows Seats 2/3 to use their ordinary built-in AI planning;
5. allows the ordinary Host phase driver to run Strategic Resolution, Encounters, Invasion handling, Post Resolution and normal turn completion;
6. stops if the Host reaches a human interactive boundary rather than inventing a decision for the human.

No PP, BC, RP, population, fleet movement, construction progress or other gameplay state is edited by the runner.

### 6.1 Advance 1 Turn

This is `mode=turns`, `turns=1`.

### 6.2 Advance N Turns

`mode=turns`.

Client range: **1..25** turns per request.

The runner performs one ordinary Host submission/resolution cycle at a time and stops early on any frozen stop condition.

The multi-turn operation is sequential, not one giant atomic simulation transaction. Each completed turn is ordinary committed game state and emits the normal invalidation/stream behavior.

### 6.3 Advance until current construction completes

`mode=until_construction_complete`.

The request names one `colony_id`.

At start, the server captures the authoritative current construction signature for that colony, including the project kind/ID and, for military Ships, design ID + revision.

The runner advances ordinary turns until the captured project emits its normal matching completion result.

Hard ceiling: **512 completed normal turns**.

The runner does not treat "construction pointer changed/disappeared" alone as successful completion. If the original project changes without its matching normal completion evidence, it stops as `construction_changed`.

This mode does not grant Production Points and does not bypass construction formulas.

## 7. Frozen advancement stop reasons

Expected successful response stop reasons:

- `requested_turns_reached`;
- `construction_completed`;
- `interactive_boundary`;
- `construction_changed`;
- `game_completed`;
- `hard_limit_reached`.

`interactive_boundary` includes a Tactical Battle, invasion decision or any later human-owned decision phase the normal Host cannot progress without player input.

A control request never auto-resolves Tactical combat or an invasion choice for the local human.

## 8. HTTP contract

Development-only route:

`POST /api/v1/games/{gameID}/reference/advance`

Logical request:

```json
{
  "schema_version": 1,
  "seat_id": 1,
  "base_revision": 123,
  "mode": "turns",
  "turns": 5
}
```

or:

```json
{
  "schema_version": 1,
  "seat_id": 1,
  "base_revision": 123,
  "mode": "until_construction_complete",
  "colony_id": 456
}
```

Frozen validation:

- `seat_id` must equal the scenario's `control_seat_id`;
- `base_revision` must equal the current session revision at request start;
- `turns` is required only for `mode=turns` and must be 1..25;
- `colony_id` is required only for construction mode;
- the control seat must be at an eligible Planning boundary;
- a non-empty saved Planning Draft rejects automatic advancement rather than being silently executed or discarded.

Logical success response includes:

- game ID;
- start turn;
- end turn;
- turns advanced;
- stop reason;
- final phase;
- final game revision;
- final change sequence.

Expected HTTP behavior:

- route absent / 404 when reference capability is globally disabled;
- 403 `reference_control_forbidden` for enabled server + non-reference game or wrong seat;
- 400 for malformed mode/range/payload;
- 409 stale-revision behavior for stale `base_revision`;
- 409 `reference_not_ready` when phase/submission/draft state is incompatible;
- 409 `reference_no_active_construction` when construction mode has no current project;
- expected stop reasons return 200, including hard-limit and interactive-boundary stops.

## 9. HMI contract

Reference HMI is development-only and server-capability driven.

For a reference game it shows:

- clear "Reference Lab" identity;
- profile label: Baseline, Mid-Tech or All-Tech;
- `Advance 1 Turn`;
- bounded `Advance N Turns` control, 1..25;
- `Advance until construction complete` when a Colony with active construction is selected.

The controls are hidden for ordinary games.

They are disabled outside an eligible Planning boundary and while a request is running.

The HMI always reloads authoritative state after the operation and displays the returned stop reason instead of pretending the requested number of turns necessarily completed.

The 390px mobile layout must remain usable.

All-Tech explanatory copy must state that all research-addressable technologies are granted while Tactical/gameplay support still depends on implemented consumers.

## 10. Built-Ship / combat fixture contract

17.0 may prepare built Ships only during reference bootstrap.

It must not introduce a second fixture-only Ship schema.

A reusable reference Ship fixture must:

1. create/save a real `ShipDesign` through the same production design compiler/validation path used by normal military design save;
2. carry a real stable design ID and revision;
3. materialize the concrete `core.Ship` through a shared production/reference materialization helper factored from normal military-ship completion;
4. deep-copy `ShipDesignSpec`, weapon mounts and visual genome exactly as normal construction does;
5. set `SourceDesignID` and `SourceDesignRevision`;
6. create an ordinary Combat Fleet placement;
7. pass normal `GameState.Validate()`.

Reference bootstrap may skip PP accumulation for these pre-built combat fixtures because they are initial-state fixtures. It may not skip design legality/snapshot validation.

The ordinary construction path continues to require PP and normal turns.

The durable Triangle games do not need extra pre-built combat ships by default. Reusable fixture helpers may inject them for tests or a later explicitly justified `game-combat-<fixture-slug>` scenario.

A public specialized combat scenario is not automatically authorized merely because it uses that name; it still requires trusted reference registration metadata.

## 11. Tactical support honesty

The All-Tech profile may know technologies whose Tactical consumers do not yet exist.

That does not make those systems tactically supported.

A reference fixture can only start interactive Tactical combat when the normal Tactical handoff accepts the built Ship snapshot.

Until a downstream child extends Tactical support, the existing narrow Frigate/Laser contract remains the immediately runnable Tactical baseline.

## 12. Persistence / import / restore contract

Reference authorization is host registration metadata, not Core simulation state.

It is deliberately not trusted from a save file.

### Export

Normal live-snapshot export remains the ordinary game/session snapshot. No save payload can mint reference capability.

### Restore into an existing reference game

Restoring a valid snapshot into an already registered reference game through the existing target-game restore path retains the target hosted game's immutable reference metadata.

This is the supported reference save/resume workflow across server restarts:

1. start server with `-reference-games`;
2. the trusted Triangle variants are registered;
3. restore the matching snapshot into the matching registered reference game.

### Import as a new game

Generic snapshot import always registers the imported game as an ordinary non-reference game, regardless of its game ID.

Import never restores or derives reference capability from `game_id`, profile-looking state or user-supplied JSON.

This prevents a save file from escalating into development controls.

## 13. Determinism acceptance contract

Gate 3/4 tests must prove:

- equal scenario/profile produces byte-stable relevant Core state;
- baseline/Mid/All share the same Turn-1 topology, IDs and RNG state after ignoring the frozen technology-list differences;
- Mid-Tech profile has 36 known fields and 94 technologies with the frozen fingerprints;
- All-Tech has fields 0..74 and technology IDs 1..203 except 125;
- profile application allocates no IDs and consumes no RNG;
- normal server mode exposes no reference route/control;
- imported snapshots cannot acquire reference control capability;
- normal construction advancement never receives free PP;
- multi-turn advancement stops at human interaction;
- built fixture Ships use normal immutable design snapshots.

## Gate 2 result

The Slice-17.0 harness contract is frozen.

Gate 3 may implement the frozen infrastructure without revisiting scenario identity, profile meaning, turn bounds, capability security, persistence authority or fixture schema.

Any later change to these contracts requires an explicit documented amendment rather than silent implementation drift.

## Gate 3 boundary

Gate 3 may now implement:

- profile/bootstrap helpers;
- trusted reference metadata;
- Mid-Tech and All-Tech Triangle registration;
- reference capability projection;
- bounded normal-turn runner;
- construction-completion runner;
- reusable built-Ship fixture materialization;
- HTTP/HMI controls;
- isolation/determinism/persistence regressions.

Gate 3 must not open Slice 17.1 or implement new hull/component/weapon/special gameplay mechanics.
