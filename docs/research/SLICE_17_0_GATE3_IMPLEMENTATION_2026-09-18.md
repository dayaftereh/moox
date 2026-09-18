# Slice 17.0 Gate 3 - Reference Harness Implementation

Date: 2026-09-18
Slice: 17.0
Gate: 3 - implementation
Contract baseline: `0b9b154` (`docs(slice): freeze slice 17.0 gate 2 harness`)
Session: `ses-20260918T152048-00240c450344`

## Result

Gate 3 implements the Gate-2 contract without opening Slice 17.1 or adding new hull/component/weapon/special gameplay breadth.

The durable Reference Harness now consists of:

- baseline Triangle: `game-triangle-2pc`;
- Turn-1 Mid-Tech Triangle: `game-triangle-mid-tech`;
- Turn-1 All-Tech Triangle: `game-triangle-all-tech`;
- trusted per-game reference metadata;
- bounded normal-turn advancement;
- normal construction-completion advancement;
- reusable built-Ship fixture materialization through production design validation;
- a development-only HTTP control surface;
- a bilingual responsive Reference Lab HMI;
- regressions covering deterministic profiles, security isolation, persistence and browser behavior.

## 1. Deterministic Triangle technology profiles

Added `internal/game/reference_triangle_profiles.go`.

The implementation freezes the Gate-2 identities and field sets directly in production test-harness code.

### Mid-Tech

`ReferenceTriangleProfileMidTech` materializes the frozen 36-field set and derives all technology applications in those fields from the normalized rules.

Tests verify:

- 36 known fields;
- 94 concrete technologies;
- exact frozen field and technology fingerprints;
- predecessor-closed field graph;
- no Core ID allocation;
- no RNG-state consumption;
- all non-technology Turn-1 state remains identical to the baseline Triangle.

### All-Tech

`ReferenceTriangleProfileAllTech` materializes fields `0..74` and every normalized research-addressable concrete technology in those fields.

Tests verify:

- 75 known fields;
- 202 concrete technologies;
- exact frozen fingerprints;
- technology 125 `phase_shifter` remains excluded while it has normalized `tech_field_id=-1`;
- no Hyper-Advanced completion state is fabricated;
- no Core ID/RNG divergence from baseline.

The profile is applied after the ordinary Triangle generator has created the normal game state. After bootstrap, `GameState.Validate()` remains authoritative.

## 2. Durable server registration and trusted metadata

`cmd/moox-server` now registers four games under explicit `-reference-games`:

1. legacy `game-1`;
2. `game-triangle-2pc`;
3. `game-triangle-mid-tech`;
4. `game-triangle-all-tech`.

The three Triangle variants use the same `ReferenceTriangleSeed` / `0x800A`, topology and Human/Darlok/Psilon controller layout.

Added `app.ReferenceGameInfo` and trusted metadata on `app.Registration` / hosted games.

The metadata contains:

- scenario identity;
- profile identity;
- control seat;
- frozen Advance-N bound;
- frozen construction-runner bound.

Game IDs do not grant capability.

`GameSummary` and `PlayerSnapshot` project the trusted metadata for HMI discovery.

The legacy `game-1` remains a reference convenience game but does not receive the new Reference Lab mutation capability.

## 3. Normal-turn Reference runner

Added `internal/app/reference.go`.

`Host.AdvanceReference` implements both frozen modes:

- `turns`: 1..25 turns;
- `until_construction_complete`: up to 512 completed normal turns.

The runner does not call `GameSession.CompleteTurn()` as a shortcut.

Every automated turn:

1. starts at an eligible Planning boundary;
2. rejects an already-submitted control seat;
3. rejects a non-empty saved Planning Draft;
4. submits an ordinary zero-command `CommandBatch` for the Human reference control seat;
5. allows the existing built-in AI controllers to plan normally;
6. uses the existing hosted phase driver for Strategic Resolution, Encounters, Invasion/Post Resolution and normal turn completion;
7. publishes the ordinary snapshot invalidation after state changes.

The runner stops at:

- requested turn count;
- matching construction completion;
- Human interactive boundary;
- construction change;
- game completion;
- frozen hard limit.

It never gifts PP, BC, RP, population, fleet movement or construction progress.

## 4. Construction completion evidence

Construction mode captures the authoritative project signature at request start.

Completion is accepted only after the session event history contains the matching ordinary completion event for that project.

A disappearing or changed construction pointer without matching completion evidence stops as `construction_changed`.

Finite completion-event mappings are implemented for:

- Building;
- Colony Ship;
- Outpost Ship;
- Troop Transport;
- Military Ship;
- Freighter Fleet;
- Planetary Transformation.

Continuous Housing is not accepted as an "until construction complete" target.

A deterministic regression proves a Building completes through normal resolution and that the runner only returns `construction_completed` after the normal completion event.

## 5. Strict development-only HTTP boundary

Added the conditional route:

`POST /api/v1/games/{gameID}/reference/advance`

The route is registered only when `server.Config.ReferenceControlsEnabled` is enabled by the explicit `-reference-games` process option.

The application layer additionally requires trusted per-game Reference metadata and the correct control seat.

Regressions prove:

- capability disabled -> route absent / 404;
- capability enabled + ordinary/untrusted game -> 403 `reference_control_forbidden`;
- stale revision -> conflict path;
- malformed request/mode/range -> bad request;
- trusted reference -> normal advancement succeeds.

The demo-fixture switch does not enable Reference controls.

## 6. Persistence cannot mint development capability

Reference authorization remains host registration metadata and is not embedded as trusted Core simulation state.

Regression coverage proves:

- exporting a trusted reference game works through normal live-snapshot persistence;
- generic snapshot import registers the imported game without Reference capability;
- game ID/profile-looking state cannot mint development controls;
- restoring a snapshot into an already trusted registered Reference game retains that host's trusted Reference metadata.

This preserves the Gate-2 non-escalation contract.

## 7. Shared immutable built-Ship fixture path

Refactored ordinary military construction so `completeMilitaryShip` delegates concrete Ship/Fleet snapshot creation to a shared `materializeMilitaryShip` helper.

Added `EconomyResolver.MaterializeReferenceMilitaryShipFixture`.

A reference combat fixture now:

1. creates a real military design with the normal save command/compiler;
2. receives a stable design ID/revision;
3. uses the same ShipDesignSpec snapshot materializer as normal Colony military construction;
4. deep-copies weapon mounts and visual genome;
5. creates an ordinary Combat Fleet;
6. passes `GameState.Validate()`.

Regression coverage mutates the source design after materialization and proves the built Ship retains its original immutable weapon snapshot.

No fixture-only Ship schema was introduced.

## 8. Reference Lab HMI

Added `web/src/ReferenceLabPanel.tsx`.

The panel is rendered only when the authoritative PlayerSnapshot carries trusted Reference metadata and the current seat is the control seat.

It provides:

- visible Reference Lab identity;
- Baseline / Mid-Tech / All-Tech profile label;
- explicit All-Tech support caveat;
- Advance 1 Turn;
- bounded Advance N control using server-projected maximum 25;
- Advance until construction completes when an owned selected Colony has active construction;
- disabled state outside an eligible Planning boundary or while a non-empty Planning Draft exists;
- authoritative reload after every operation;
- explicit stop-reason feedback.

The UI is bilingual DE/EN and has a dedicated responsive 390px layout.

The ordinary game `game-1` does not render Reference controls.

## 9. Browser regression

Added:

`web/scripts/check-reference-lab-browser.mjs`

and npm entry:

`check:reference-lab:browser`.

The headless Chrome smoke test verifies a real server started with `-reference-games`:

- trusted Mid-Tech and All-Tech projection;
- no Reference capability on `game-1`;
- All-Tech Reference Lab visible at 390px;
- German profile/title copy visible;
- Advance-N max is 25;
- no horizontal overflow at 390px;
- clicking Advance 1 advances the authoritative server game by exactly one normal turn;
- result feedback is visible;
- desktop layout adapts without horizontal overflow;
- navigating to `game-1` removes the Reference Lab UI.

Result: PASS.

## 10. Verification

### Focused game/app/server tests

PASS:

- Reference technology profile/fingerprint regressions;
- predecessor closure;
- normal reference turn pipeline;
- ordinary-game rejection;
- stale-revision rejection;
- non-empty draft rejection;
- construction completion event proof;
- persistence import non-escalation / trusted restore;
- conditional route isolation;
- trusted HTTP advancement;
- immutable Reference military Ship fixture.

### Broad Go suite

`go test ./...`

PASS across all packages, including the long-running `internal/session` suite.

A duplicate test run triggered by the managed-tool timeout also completed PASS.

### Static Go analysis

`go vet ./...`

PASS.

### Web production build

`npm run build`

PASS, including:

- mojibake/UTF-8 check;
- New Game selector contract;
- race assets;
- technology-start contract;
- opponent-composition contract;
- TypeScript build;
- Vite production build.

### Browser

`npm run check:reference-lab:browser -- http://127.0.0.1:7183`

PASS.

## Gate 3 amendment

A later same-Gate amendment adds the normal MOO2 construction-buyout path plus Reference-only BC grants and moves the round controls under Development Tools. See `docs/research/SLICE_17_0_GATE3_BUYOUT_DEV_BC_AMENDMENT_2026-09-18.md`.

## Gate 3 result

The frozen Slice-17.0 Reference Harness is implemented end-to-end without introducing a parallel simulation path.

The next step is Gate 4 QA/close only.

Gate 4 must independently exercise the final acceptance matrix and close Slice 17.0. It must not implicitly open Slice 17.1.

## Gate 3 live amendment - same-turn drafted construction buyout

Live 7171 review exposed a UI/authority mismatch: a saved Planning Draft could already be rendered as the visible current construction while buyout remained hidden until the draft had first been submitted as a normal turn.

The amendment removes that mismatch:

- the visible drafted current project now receives the same original MOO2 buyout-price projection in the HMI;
- clicking Buy on a saved construction draft sends the normal construction-buyout command;
- the Host detects the trusted saved construction draft for that colony and atomically resolves an internal `colony.buy_planned_construction` command;
- the normal construction-queue validator applies the saved queue on a cloned authoritative state before the ordinary BC buyout resolver runs;
- if any validation or affordability check fails, no state is committed;
- on success, the matching construction Draft order is removed;
- the bought project remains at 100% PP and completes on the next normal turn exactly like an authoritative buyout.

Regression coverage proves the same-planning-phase path, exact treasury deduction, draft removal and next-turn completion. The headless Chrome smoke now exercises the drafted path directly.
