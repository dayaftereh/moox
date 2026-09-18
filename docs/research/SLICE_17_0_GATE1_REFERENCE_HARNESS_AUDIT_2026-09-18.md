# Slice 17.0 Gate 1 - Ship Designer / Triangle Reference Test Harness Audit

Date: 2026-09-18
Slice: 17.0
Gate: 1 - audit / research
Baseline: `d69f9b3` (`docs(slice): decompose slice 17 ship design`)
Session: `ses-20260918T152048-00240c450344`

## Recovery / repository check

The handoff was independently reconciled before opening the child slice:

- session state: active / healthy / write;
- checkpoint: `slice17-decomposition-committed`;
- branch: `main`;
- HEAD: `d69f9b3`;
- working tree before opening 17.0: clean;
- `main` ahead of `origin/main` by 75 commits;
- open slice markers before opening 17.0: 0;
- no push performed.

A recovery checkpoint `slice17-0-g1-recovery-start` was persisted before the audit changes.

## 1. Existing reference-game foundation

The development entry point is real and already isolated at startup:

- `cmd/moox-server/main.go` exposes `-reference-games` with an explicit development-only description.
- Production startup with `referenceGames=false` registers no reference games.
- With the flag enabled, `registerReferenceGames` registers:
  - `game-1`: a normal two-player Average New Game through `Host.CreateGame`;
  - `game-triangle-2pc`: the durable three-player Triangle game.
- `NewReferenceTriangleGame` uses the normal rules object, normal technology initialization, normal starting-asset materialization, normal economy finalization and final `GameState.Validate()`.
- Triangle registration then creates an ordinary `session.GameSession`, ordinary `EconomyResolver` and ordinary `app.Host` registration.
- Existing tests already prove production/reference separation and Triangle determinism.

The existing Triangle scenario identity `triangle-2pc-v1` and public game ID `game-triangle-2pc` must remain stable.

## 2. Technology bootstrap audit

The supported new-game technology boundary is `EconomyRules.InitializeNewGameTechnologies`.

Current contract:

- Pre-Warp and Average are supported.
- Advanced remains intentionally locked/deferred and must not be used as a shortcut for the Designer Lab.
- Technology ownership is persisted on `core.Empire` through sorted `KnownTechnologyIDs` plus `KnownTechnologyFieldIDs`; active research is separate.
- Existing focused tests already use deterministic direct technology materialization for ship-design QA by adding normalized technology IDs before exercising the normal designer.

Gate-1 safe bootstrap rule:

1. create the reference game through an ordinary supported start, normally Average unless Gate 2 freezes another supported baseline;
2. only during reference bootstrap, add the curated extra normalized technology IDs required by the selected lab profile;
3. every granted technology ID must exist in the loaded normalized ruleset / `TechnologyFieldByID`;
4. grants must be sorted and de-duplicated deterministically;
5. do not call or emulate the locked Advanced-start generator;
6. do not fabricate production, design or tactical effects with the technology grant;
7. leave ordinary research-field progression state intact unless Gate 2 explicitly freezes a profile-specific field policy;
8. require a valid post-bootstrap `GameState.Validate()` before registration.

This keeps the high-tech lab a deterministic initial-state fixture rather than a second technology rules engine.

## 3. Ship Designer / Construction / turn path audit

The existing military-design path is suitable for reuse.

### Design authority

- `core.ShipDesignSpec` is the persisted authoritative design snapshot.
- `ShipDesign` has stable ID + revision.
- Weapon mounts are persisted as ordered `ShipWeaponMount` entries.
- `MilitaryDesignerCatalog` already projects six normalized hull identities while intentionally allowing only the current supported save scope.
- Designer save commands flow through the ordinary immediate-command authority.

### Construction authority

Military construction stores the selected design ID and revision. Completion resolves the ordinary design again and calls `completeMilitaryShip`.

`completeMilitaryShip`:

- copies the current `ShipDesignSpec`;
- deep-copies weapon mounts;
- clones the visual genome;
- writes `SourceDesignID` + `SourceDesignRevision`;
- creates an ordinary persistent `core.Ship`;
- creates an ordinary strategic Combat Fleet at the Colony system.

Core validation requires every built Ship to reference a real design and a valid source revision. Therefore 17.0 combat fixtures must materialize ships through the same snapshot shape; no fixture-only ship schema is permitted.

### Normal turn advancement

`GameSession.SubmitTurn` accepts a valid zero-command `CommandBatch`. `Host.SubmitTurn` then goes through the ordinary hosted mutation path and the existing phase driver.

The host already owns the normal progression logic for:

- planning submissions;
- strategic resolution;
- encounters / Battles;
- invasion decisions;
- post-resolution;
- `GameSession.CompleteTurn`;
- return to the next Planning phase.

`Host.AdvanceAutomation` proves the same host can intentionally advance at most one turn when all active empires are built-in AI, but it deliberately rejects the normal local-human reference-game shape.

Therefore the 17.0 dev control must not call `CompleteTurn` directly. The safe shape for a human reference lab is a reference-only helper that submits the current local-human seat's ordinary empty/current plan through `Host.SubmitTurn` and lets the existing host driver advance AI/resolution normally.

The bounded runner must stop rather than bypass a newly exposed human interactive boundary such as Tactical combat or an invasion decision. Gate 2 owns the exact stop/result contract and numeric bound.

No PP may be added, no construction progress may be edited directly, and "advance until construction complete" is only a bounded loop around normal turn submission/resolution.

## 4. BattleSession / Triangle / built-Ship handoff audit

The Tactical handoff already consumes concrete strategic Ships:

- `tacticalMetadataForEncounter` resolves each encounter Ship ID back to `core.GameState.Ships`;
- tactical metadata is derived from the built Ship's persisted design snapshot;
- `GameSession.prepareEncountersLocked` validates the prepared Tactical snapshot against the strategic Ship before creating `battle.Session`;
- `battle.NewSession` is therefore downstream of the ordinary strategic Ship identity.

Current Tactical support is intentionally narrow. `baselineCombatantUnsupportedReason` currently allows only the Slice-15.5 baseline: Frigate, Electronic Computer, Titanium Armor, no shield, supported drive, Standard Fuel and positive-count standard Laser mounts. The session-side Tactical snapshot validation is similarly narrow.

Consequences for 17.0:

- the first immediately runnable combat fixture can be the already-supported Frigate/Laser vertical slice;
- richer built Ships may be materialized as valid strategic reference fixtures as their 17.x design support lands;
- a fixture must not claim Tactical support until the Tactical consumer accepts that equipment;
- High-Tech Designer Lab availability and High-Tech Tactical support are separate capabilities.

This preserves the Slice-17.4 rule that unsupported tactical effects are never presented as complete.

## 5. Stable reference scenario IDs and ownership

Existing IDs remain unchanged:

| Purpose | Public game ID | Internal scenario identity |
| --- | --- | --- |
| Standard reference game | `game-1` | legacy standard reference |
| Triangle movement/contact lab | `game-triangle-2pc` | `triangle-2pc-v1` |

Reserved 17.0 public IDs:

| Purpose | Public game ID |
| --- | --- |
| Designer early-tech lab | `game-designer-early` |
| Designer mid-tech lab | `game-designer-mid` |
| Designer high/all-supported-tech lab | `game-designer-high` |
| Construction lab | `game-construction-lab` |
| First runnable combat lab | `game-combat-frigate-laser` |

Future combat fixtures use the stable namespace `game-combat-<fixture-slug>` and are only registered when the corresponding production design/tactical support exists. Gate 2 freezes the exact first matrix and seeds; Gate 1 only reserves the durable IDs/namespace above.

Code ownership:

- `internal/game`: scenario bootstrap/profile definitions, normalized technology grants and built-Ship fixture materialization;
- `cmd/moox-server`: explicit development flag and reference-game registration only;
- `internal/app`: per-game reference metadata and normal turn-driver invocation;
- `internal/server`: reference-control HTTP authorization/exposure;
- `web`: controls rendered only from server-advertised reference capability.

## 6. Strict dev-only exposure boundary

Current protection is sufficient for fixture registration but not yet sufficient for future mutation controls:

- `-reference-games` controls whether reference games are registered;
- `app.Registration` currently has no reference/capability metadata;
- `app.GameSummary` currently has no reference scenario identity;
- `server.Config` currently has no reference-control capability flag.

Gate 2 must freeze a double guard before implementation:

1. **global server capability**: reference controls exist only when the process was explicitly started with the reference-games development capability;
2. **per-game capability**: the target game must have been registered as a reference scenario with an explicit immutable scenario identity.

Guessing a reference-looking game ID must never authorize a development action. Ordinary created/imported games must remain ineligible even on a server that happens to host reference scenarios.

The HMI must render development controls only when the server projects the explicit capability. Client-side hiding is convenience, not authorization; the server remains authoritative.

## 7. Focused verification

Existing production tests were run without product-code changes:

- `go test ./cmd/moox-server -run Reference -count=1` - PASS
- `go test ./internal/game -run "ReferenceTriangle|MilitaryDesign|MilitaryDesignerCatalog|TacticalMetadata" -count=1` - PASS
- `go test ./internal/session -run "Military|Tactical" -count=1` - PASS

## Gate 1 result

Gate 1 finds no architecture blocker.

The existing code already supplies the important production paths: deterministic reference registration, supported technology state, normal Designer authority, immutable built-Ship snapshots, normal Colony Construction, normal hosted turn resolution and real Tactical/BattleSession handoff.

The missing Slice-17.0 work is infrastructure around those paths, not a replacement rules engine.

## Gate 2 decisions to freeze

Gate 2 must explicitly freeze:

- exact scenario matrix and deterministic seeds;
- exact early/mid/high technology ID profiles;
- whether any profile also adjusts known-field state;
- maximum `Advance N Turns` bound;
- exact stop/result semantics for Battles, invasions, victory and construction completion;
- reusable built-Ship materialization contract;
- per-game reference metadata shape;
- HTTP endpoints and response codes for unauthorized/non-reference games;
- HMI capability projection and control placement;
- persistence/import behavior for reference scenario identity.

No Gate-2 contract has been implemented or implicitly opened by this audit.
