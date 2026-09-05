# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: `docs/slices/_OPEN_15_2_FUNCTIONAL_STRATEGIC_GAMEPLAY_HMI_2026-09-04.md`.
- Active implementation slice: **Slice 15.2 Gates 1-3 complete; Gate 4 remediation G4-A/G4-B complete; G4-B2 interaction-density refinement accepted; G4-C deferred until B2 is complete**.
- Slices 01-12 are closed.
- Slice 10 implementation/data/evidence commit: `b91f65e` (`game: add diplomacy war peace baseline`).
- Core `StateSchemaVersion`: **23**.
- Economy ruleset schema: **8**.
- Slice-10 permanent evidence: `docs/research/DIPLOMACY_WAR_PEACE_BASELINE_2026-09-02.md`.
- Post-Slice-12 audit: docs/research/POST_MILESTONE_FIDELITY_DEPTH_BACKLOG_AUDIT_2026-09-03.md.
- Slices **13-14 and 15.1 are closed**; **15.2 is open with Gates 1-3 complete and Gate 4 pending**; 15.3-15.6, 16-17 and reserved Slice 20 remain prepared.
- Current objective: **Slice 15.2 Gate 4-B2-B - add the dedicated Colony build-management route/page after B2-A restored the table-first Colony overview and multi-unit desktop/touch Population drag/drop.**
- Slice-15.1 closure evidence: `docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE4_2026-09-04.md`.
- Slice-15.2 Gate-1 evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE1_2026-09-04.md`.
- Slice-15.2 Gate-2 evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE2_2026-09-04.md`.
- Slice-15.2 Gate-3 evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE3_2026-09-05.md`.
- Slice-15.2 Gate-4 blocker remediation: `docs/research/SLICE_15_2_GATE4_BLOCKER_REMEDIATION_2026-09-05.md`.
- Prepared future Espionage mechanics: `docs/slices/PLANNED_20_ESPIONAGE_INTELLIGENCE_BASELINE.md`.
- Accepted downstream 15.2/15.3 strategic product direction: `docs/research/SLICE_15_2_STRATEGIC_HMI_PRODUCT_DIRECTION_2026-09-04.md` - 2D Galaxy primary view, orbital star-system dialog, legal Colonize confirmation, Colony table + full Colony Detail/build queue, location-grouped Fleets, eight-category Research, first-class Diplomacy and planned Espionage, plus retained OX lettermark / future spacecraft-space icon.
- Accepted Tactical roadmap amendment: `docs/research/SLICE_15_5_INTERACTIVE_TACTICAL_COMBAT_PRODUCT_DIRECTION_2026-09-04.md` - interactive Tactical is a dedicated Slice 15.5; the former final integration slice moves to 15.6.


## Post-Slice-12 fidelity/depth audit - complete

The milestone re-audit identified independent playability as the main post-lifecycle bottleneck. Slices 13 Built-in AI and 14 live GameSession persistence are closed. Slice 15 is now a six-part browser-playability family: 15.1 UX/navigation/design system, 15.2 functional strategic HMI, 15.3 MOOX visual identity/assets, 15.4 rich gameplay decisions/persistence UX, **15.5 interactive 2D Tactical Combat**, 15.6 final browser vertical slice. Slice 16 New Game/preset-race breadth and Slice 17 military design breadth remain later independent work. Slice 15.1 Gates 1-4 are complete and the slice is closed.

Permanent audit: `docs/research/POST_MILESTONE_FIDELITY_DEPTH_BACKLOG_AUDIT_2026-09-03.md`.
## Closed Slice 12 - Empire elimination / first headless victory loop

Slice 12 is closed with Gates 1-4 complete. A real canonical `NewGame(0x8009)` now plays through public Research, Colony/Outpost expansion, war, Fleet movement, Troop Transport invasion and final Colony conquest to deterministic zero-Colony Empire elimination and an immutable `conquest` winner in `session.PhaseCompleted`. Remaining off-system loser Fleets/Outposts do not delay elimination and are cleaned deterministically. Completed sessions are read-only for gameplay while preserving Player/Observer/HTTP/WebSocket/reconnect and versioned exact snapshot access. The React HMI projects the same result. The complete match runs twice byte-identically in the canonical regression, and the implementation also fixed the long-run Population Cohort capacity clamp exposed by the first full match attempt. Implementation/evidence commit: `90b83d8`.

Permanent evidence: `docs/research/EMPIRE_ELIMINATION_FIRST_HEADLESS_VICTORY_LOOP_2026-09-03.md`.
## Closed Slice 11 - Troop Transport / invasion / conquest baseline

Slice 11 is closed with Gates 1-4 complete. Core23 standing Infantry, fixed Troop Transport, server-authoritative invasion decisions, deterministic Ground Combat, Colony conquest/cohort/Capital handoff, post-conquest strategic/economic continuation, HTTP authority and minimal React proof are implemented and independently Gate-4 verified. Implementation/evidence commit: `5b363ce`. Current canonical New-Game seed `0x8009` Core23 state hash remains `83614740b216409877b03c536a16328c7fa7031867f58dbeb70857a8953f5762`. Slice 12 subsequently closed the first supported deterministic headless match lifecycle.

Permanent evidence: `docs/research/TROOP_TRANSPORT_INVASION_CONQUEST_BASELINE_2026-09-03.md`.
## Closed Slice 10 - Diplomacy / war / peace baseline

Slice 10 is closed. Core22 implements reciprocal `peace|war`, directional pending peace offers, revision-bound pre-first-submission `declare_war|offer_peace|accept_peace`, deterministic diplomacy events and player-safe projections. `MayAttackEmpire` is the shared war-only blockade/encounter authority. HTTP and React proof cover Human war declaration -> peace offer -> Darlok acceptance, while timing tests prove neutral same-system suppression, war activation, accepted-peace suppression with Fleets retained and rejection after Battle materialization. Gate 4 independently repeated focused regressions, exact replay/save-load, authority scans and full Go/Web QA. Implementation/evidence commit: `b91f65e`. Automatic peace expiry, sneak attacks, treaties, AI and espionage remain deferred.

Permanent evidence: `docs/research/DIPLOMACY_WAR_PEACE_BASELINE_2026-09-02.md`.
## Closed Slice 09 - Deterministic New Game / galaxy generation baseline

**Gates 1-4 are complete.** Slice 09 replaces production fixture bootstrap with an authoritative deterministic `NewGame(seed, settings)` path for the frozen Small / Normal / Average / Tactical two-player Human+Darlok baseline.

Delivered baseline:

- original-derived galaxy generation tables normalized into `data/rulesets/moo2-1.31/new_game_galaxy.json`;
- one shared New Game RNG owner and stable generation/ID order;
- deterministic 20-star Small galaxy plus planet/Homeworld/start-state generation;
- legal Human+Darlok Colonies, Population, Research, Treasury and starting Fleets;
- generated state validated through a real two-seat `GameSession` and Turn 1 -> Turn 2 strategic cycle;
- authoritative `POST /api/v1/games` creation with string uint64 seed, stable error mapping and atomic duplicate rejection;
- production server starts empty; legacy fixture startup is explicit development-only `-demo-fixture`;
- browser New Game form uses the server-authoritative create path and existing snapshot/CommandBatch flow;
- golden seed `0x8009` current Core22 state SHA-256: `8effff679109cd10a427f1c80b83047dc0d2fa8e5e4871a67235a486dbc2fc68`;
- Gate 4 passed repeated golden/determinism tests, byte-identical equal-seed JSON state, generated strategic-turn integration, frontend authority scan, standalone server + managed Chrome runtime proof, full Go tests/vet/web build and diff checks.

Permanent evidence: `docs/research/NEW_GAME_GALAXY_GENERATION_BASELINE_2026-09-02.md`.

## Prepared next-slice queue - 2026-09-02

Accepted application direction: `docs/architecture/ADR-0004-authoritative-server-web-client.md`. The primary product topology is an authoritative Go game server plus browser-first web HMI over HTTP/WebSocket; Wails v3 is optional one-click native packaging only and must not bypass the server gameplay contract.

1. Slice 10 - **Diplomacy, war and peace baseline** (`PLANNED_10_DIPLOMACY_WAR_PEACE_BASELINE.md`).
2. Slice 11 - **closed, Gates 1-4 complete** (`PLANNED_11_TROOP_TRANSPORT_INVASION_CONQUEST_BASELINE.md`, implementation/evidence `5b363ce`).
3. Slice 12 - **Empire elimination and first headless victory loop** (`PLANNED_12_EMPIRE_ELIMINATION_FIRST_HEADLESS_VICTORY_LOOP.md`).

Roadmap milestone achieved by Slice 12. The post-milestone audit now prepares Slices 13-17 toward the next milestone: a saveable Human-vs-built-in-AI browser match, followed by New Game/preset-race and Ship Designer breadth.

## Closed Slice 07 - Tactical ship combat baseline

**Gates 1-4 are complete.** Gameplay/data/evidence commit `889f977` implements the deliberately narrow original-derived tactical ship-combat baseline on top of the Slice-06 encounter handoff.

Delivered baseline:

- Core schema **21** persists structural weapon-mount identity on ShipDesign/built Ship snapshots while preserving immutable built-Ship equipment across later design revision.
- `empire.save_military_design` remains unarmed-compatible and adds only the accepted one-slot standard Laser surface with technology, space and cost enforcement.
- tactical ruleset schema **1** contains the evidenced initiative, exact original uint32 RNG, Beam/Laser, Frigate, Nuclear/Fusion and Electronic-Computer facts; economy schema8 and ship-hulls schema3 remain unchanged.
- the exact supported 1-vs-1 strategic fixture freezes a self-contained TacticalSpec and executes `battle.fire_beam` / `battle.end_activation` with BattleSession-local round, initiative, command sequence, Armor/Structure, readiness, destruction and deterministic local events.
- the golden fixture consumes RNG values `100,19,100,100`, ends at `0xFD95EBB9`, applies Armor4->0 then Structure0->4 and yields `tactical_victory`.
- rejected or deferred tactical actions are transactional; unsupported strategic encounters remain explicit lifecycle/manual-result Battles rather than approximated.
- terminal tactical results reuse the Slice-06 cloned strategic continuation, including retry-safe failure semantics, concrete defender Ship/empty-Fleet cleanup and stable Battle-ID completion order across independent parallel systems.
- Gate 4 re-ran focused battle/session/strategic handoff and fixed-seed regressions, verified wall-clock-independent ordering and strategic loss reconciliation, and passed full `go test ./... -count=1`, `go vet ./...`, changed-file gofmt and `git diff --check`.

Permanent evidence: `docs/research/TACTICAL_SHIP_COMBAT_BASELINE_2026-09-01.md`.

There is no active slice after this closure. Choose and prepare the next objective before opening a new recovery marker.
## Closed Slice 06 - Strategic hostile encounters / BattleSession handoff

**Gates 1-4 are complete.** Gameplay/evidence commit `096fd0a` implements strategic hostile encounter production, staged BattleSession waves, casualty/retreat reconciliation and post-battle continuation; Core schema 20/economy schema 8 remain unchanged. Gate 4 passed focused strategic regressions, full tests, vet and diff checks.

Permanent evidence: `docs/research/STRATEGIC_HOSTILE_ENCOUNTERS_BATTLE_HANDOFF_2026-08-31.md`.

Gate-1 conclusions:

- corrected Watcom symbol parsing places `Next_Turn_Calc_` at VA `0x136B3`, `Search_For_Battles_` at `0xE9D62`, `Move_All_Ships_Toward_Stars_` at `0xFFEEA`; future RE must read each current symbol address from the four bytes before its record header, not the trailing next-symbol value;
- original `Next_Turn_Calc_` does movement -> player/colony changes/Production -> `Search_For_Battles_` -> later `Compute_Blockades_` -> `Move_Settlers_`, so the MOOX encounter boundary belongs after Fleet transit + Construction and before blockade/population transfers;
- combat search scans all stationary Ships at real stars, so already-stationary hostile co-location and same-turn arrivals are both eligible;
- original sides aggregate all same-owner Ships at a star; MOOX must aggregate all same-Empire combat Fleet containers/ShipIDs into one side rather than create Fleet-pair cross products;
- 3+ Empire systems resolve as sequential attacker/target pairs, while independent systems are safe candidates for MOOX parallel BattleSession waves;
- original battle-star/attacker order is RNG-driven and local humans choose targets; Gate 1 recommends canonical stable ordering for current already-hostile MOOX rather than copying only the RNG fragment without attack-choice/sneak-war semantics;
- current `hostile` directed relations are the safe automatic-attack subset; neutral sneak attacks/war declarations remain deferred;
- Colony/Transport/Outpost Ships are noncombat assets. Unescorted civilian-only targets are destroyed without a fake tactical battle; civilians accompanying a losing combat side retreat with it;
- original losing survivors are marked retreating, then moved toward the closest other owned Colony star before Blockades are computed; if retreat cannot be established they are destroyed;
- current `battle.Result{WinnerSeats, Outcome}` is therefore too weak for a correct strategic resume. Gate 2 must accept a minimum winner/post-battle-disposition contract; tactical damage remains Slice 07;
- original Colony/orbital defense is a valid combat target, but MOOX lacks a tactical station model. Carry Colony context for Fleet-vs-Fleet handoff; defer colony/station-only battles, bombardment and invasion;
- newly arrived Colony/Outpost special Ships pass through combat search before later human colonization/outpost opportunity discovery;
- current `GameSession` commits a fully resolved State before entering `PhaseEncounters`; Slice 06 requires a pre-encounter / encounter-wave / post-encounter continuation instead so Blockades/transfers wait for battle results;
- no durable GameSession/BattleSession persistence exists today; Gate 1 recommends deterministic replay/Observer recovery only, not pretending active-battle disk saves exist.

Proposed Gate-2 shape: keep Core schema 20, split strategic resolution after Construction, generate one pair per System per wave using enhanced System/attacker/defender/Fleet/Ship/civilian IDs, apply results atomically in Battle-ID order, re-evaluate multi-party Systems, then run Blockades/transfers. Normal battle results require one winner and may carry destroyed Ship IDs; surviving loser assets perform the directly proven minimum retreat. Civilian-only overrun auto-resolves; station-only combat remains deferred.

Accepted Gate-2 contract:

- keep Core schema 20 and economy schema 8; no persisted Core encounter/retreat object;
- retain the existing one-shot `Resolver` for compatibility and add a staged `EncounterResolver` continuation for real encounter waves; `EconomyResolver` implements it and GameSession may retain that continuation in memory only while encounters are active;
- split the turn after Fleet transit + Construction and before Blockades/transfers; if no BattleSession is needed, run post-encounter work immediately, otherwise freeze the pre-post-resolution State in `PhaseEncounters`;
- encounter sides aggregate all same-Empire stationary combat Fleet IDs/Ship IDs and separate Colony/Outpost civilian Fleet IDs; explicit attacker/defender Empire+Seat identity, System ID and sorted defender Colony IDs are part of the Battle Spec;
- only directed `hostile` relations auto-attack; canonical order is System ID, attacker Empire ID, defender Empire ID; all participants must have real Seats;
- at most one Battle per System per wave; independent Systems may run in parallel; completed waves apply in Battle-ID order and contested Systems are re-derived before the next wave;
- current-wave IDs are frozen references into the Core State; no duplicate tactical Ship snapshot/HP model is added in Slice 06;
- supported Battle result has exactly one winner, non-empty Outcome metadata and optional sorted unique destroyed concrete Ship IDs restricted to that Battle Spec;
- the final outstanding result of a wave is validated without mutating the child, then continuation/next-wave preparation runs on cloned State; only successful preparation commits the final result, strategic changes and next wave atomically;
- destroyed concrete Ships are removed from `GameState.Ships` and Fleet ShipIDs; empty combat Fleets are removed;
- surviving loser combat/civilian Fleets retreat to the closest other owned Colony System by squared coordinate distance, tie lowest System ID; if no destination or movement cannot be established for an asset, that asset is removed rather than trying a second destination;
- unescorted civilian-only target Fleets can be immediately overrun only when no defender combat Fleet and no defender Colony are present; defender Colony suppresses fake civilian/station auto-combat;
- Fleet-vs-Fleet battles may carry defender Colony context but no synthetic station tactical unit, bombardment, invasion or ownership change;
- no-encounter turn commits normally; encounter turn commits the frozen pre-encounter State then one revision per completed wave, with final wave + post-encounter continuation allowed as one atomic revision;
- Observer/replay must preserve Battle Specs/IDs/seeds and stable event order; durable active-battle process-restart persistence remains deferred.

Gate-3 minimum tests include no-encounter compatibility, stationary and same-turn-arrival encounters, newly constructed Ship participation, same-Empire Fleet aggregation, directed/reciprocal hostility ordering, independent-system parallel waves, three-Empire sequential waves, wall-clock completion independence, atomic invalid-final-result retry, destroyed Ship/Fleet cleanup, deterministic retreat/destruction, civilian overrun/Colony suppression, Colony context, post-battle Blockade/transfer timing, Observer isolation and identical replay.

Gate-3 implementation now in tree:

- `EconomyResolver.Resolve` freezes the turn after Fleet transit + Construction and before Blockades/transfers when a BattleSession wave is required; `ResumeAfterEncounters` applies outcomes, re-derives follow-up waves and only then executes post-encounter Blockades/transfers/economy materialization;
- stationary same-Empire Fleet containers aggregate into directed attacker/defender sides with sorted Fleet/Ship/civilian IDs and defender Colony context; directed `hostile` only, canonical System/attacker/defender order, one Battle/System/wave;
- civilian-only Colony/Outpost Fleets may be overrun immediately only without defender combat Fleet/Colony; Colony-only/station-only combat remains deferred;
- `battle.Spec` now preserves strategic side context and `battle.Result` normalizes to exactly one winner with optional validated `DestroyedShipIDs`; `ValidateResult` is pure;
- GameSession retains an in-memory staged continuation, commits parallel results only in Battle-ID order and prepares the final continuation/next wave on cloned State before atomically completing the final child;
- casualties remove concrete Ships/empty Fleets; surviving loser combat and fixed-special civilian Fleets retreat to the closest other owned Colony System by squared distance/tie-lowest-SystemID or are destroyed if retreat cannot be established;
- same-turn arrival and military Construction participate before combat; Blockade and Population-transfer work waits until the final wave; Observer BattleSpecs are detached and identical real sessions replay identical State/events/battles;
- focused strategic encounter tests, `go test ./... -count=1` and `git diff --check` pass. Gate 4 still owns fresh `go vet ./...`, final diff review, commits and closure.
Slice 06 is closed. Start Slice 07 only with a fresh repository/session check and a new dated `_OPEN_` marker.
## Closed Slice 05 - Command Points and ship Maintenance

**Gates 1-4 are complete.** Gate 4 passed fresh gofmt, focused original-accounting/compatibility regressions, full tests, vet and diff checks; implementation commit `16afe0b` is recorded and the recovery marker is removed.

Permanent evidence: `docs/research/COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md`.

Gate-1 conclusions:

- military Ship CP usage is authoritative hull `SizeIndex + 1` (Frigate 1 through Doom Star 6);
- fixed Colony/Outpost special ships consume 1 CP each, including while in normal transit; Population-transfer Settlers/Freighters are not original troop Transports and do not acquire fake CP usage;
- original generated capacity starts at 5, adds Star Base/Battlestation/Star Fortress +1/+2/+3, strongest known Communications +1/+2/+3 per station, Warlord +2 per owned Colony, then Imperium adds `floor(capacity/2)`;
- an original officer-derived capacity source exists but remains zero/deferred because MOOX has no Leader subsystem;
- player-facing overage Maintenance is exactly 10 BC per excess CP; the separate original NPC/difficulty `12 - global setting` branch remains deferred;
- original scrap/dead cleanup precedes Maintenance and current-turn ship Production completes after Maintenance, so a newly completed Ship first affects CP Maintenance on the following settlement;
- Gate 1 found that pre-Slice MOOX permitted multiple orbital station tiers on one Colony; Gate 3 now enforces the original Star Base -> Battlestation -> Star Fortress replacement/buildability relationship;
- implementation uses Core schema 20 + materialized Empire `{capacity, used}`, Treasury ship-command Maintenance, economy ruleset schema 8 Command-Point constants/producers, no ship-hulls schema bump, and no new player command.

Gate 4 re-ran the complete QA surface successfully, including `go vet ./...`, reviewed accounting/schema/station/Treasury diffs, and committed gameplay/evidence as `16afe0b`. Slice 05 is closed; Slice 06 is next.
## Closed Slice 04 - Combat Fleet movement, merge and split

**Gates 1-4 are complete.** Final Gate 4 gofmt, full tests, vet, focused compatibility regressions and diff checks passed; the implementation is committed and the recovery marker is removed.

Permanent evidence: `docs/research/COMBAT_FLEET_MOVEMENT_MERGE_SPLIT_2026-08-31.md`.

Gate-1 conclusions:

- original strategic movement operates on an arbitrary selected subset of co-located ships; MOOX must map that to explicit stable Fleet containers;
- normal in-hyperspace destination changes are technology-gated by Hyperspace Communications, so this slice can reject in-transit split/merge/reroute;
- current ordinary military movement should derive effective FTL speed from the Empire's current best Warp Drive plus Trans-Dimensional, rather than from stale build-snapshot speeds;
- current Fleet range similarly uses the Empire's best Fuel Cell and the already-proven Colony/Outpost supply origins; Extended Fuel Tanks remains a future per-Ship modifier boundary;
- existing `AtSystemID` / `DestinationSystemID` / `RemainingTurns` state can be generalized directly; no second transit model is needed;
- original selected-subset movement maps cleanly to `move_fleet` with optional `ship_ids` for atomic split+move, plus explicit stationary split/merge commands;
- Fleet movement/arrival precedes blockade recomputation, so departure removes and arrival can establish blockade in the same resolution;
- hostile arrival may temporarily materialize stationary co-location/blockade; Slice 06 owns encounter/BattleSession handoff.

Gate 2 accepted the schema-19/API/event contract and Gate 3 now implements it.

Gate-3 runtime result:

- Core schema 19 accepts ordinary concrete combat-Fleet transit and rejects empty/locationless/cached-FTL combat Fleet state;
- combat movement derives current Empire Warp Drive + Trans-Dimensional speed and current Fuel Cell range while validating every concrete member Ship;
- `empire.move_fleet` supports deterministic whole-Fleet or selected-subset movement; proper subsets allocate one new Fleet ID and emit `empire.fleet_split` before `empire.fleet_movement_started`;
- `empire.split_fleet` and `empire.merge_fleets` operate only on owned stationary ordinary combat Fleets and preserve sorted unique non-empty Ship composition;
- rejected subset movement validates before mutation and does not consume `NextID`;
- combat transit uses the existing destination/remaining-turn countdown, arrival materializes before blockade recomputation, and no in-transit reroute/split/merge is introduced;
- schema-19 save/load, Observer `ShipIDs` isolation, identical-session state/event replay, Colony/Outpost compatibility and full repository tests are green.

Gate 4 closure: fresh repository/session check found no foreign writer and exactly one expected OPEN marker. `gofmt` on changed Go files, `go test ./... -count=1`, `go vet ./...`, focused CombatFleet/Strategic/Military/ColonyShip/Outpost/Blockade/Observer/Save/Load/Replay regressions and `git diff --check` all passed. Gameplay/evidence commit: `f297480`. The OPEN marker is removed in the closing documentation commit. No push was performed. Next prepared objective: Slice 05 **Command Points / ship maintenance**.
## Closed Slice 03 - Military Ship core / design baseline

**Gates 1-4 are complete.** Final Gate 4 gofmt, full tests, vet, focused compatibility regressions and diff checks passed; the implementation is committed and the recovery marker is removed.

Current runtime boundaries:

- Core `StateSchemaVersion` is **18** with persisted `ShipDesigns`, concrete `Ships` and combat-Fleet `ShipIDs`;
- MOOX intentionally permits an arbitrary number of active designs per Empire; the original six `0x63` slots remain evidence only, with dedicated >6-design validation/round-trip coverage;
- current designs use stable IDs and monotonic revisions, while built Ships persist an independent source-design snapshot so later current-design edits cannot mutate existing Ships;
- `ship_hulls.json` schema 3 now normalizes the six direct original hull cost/space pairs plus the supported mandatory Drive/Computer/Armor/Shield/Fuel tables from `Orion2.exe`;
- the first supported authoritative design surface is a cleared/unarmed Frigate with a legal original Frigate picture and the best known supported mandatory strategic systems;
- the exact first baseline is 20 PP Frigate hull + 5 PP Electronic Computer = 25 PP, with current Feudal production cost 17 PP;
- every owned active design is surfaced as a distinct authoritative `military_ship` Construction choice; `colony.queue_military_ship` snapshots exact design ID/revision and design revision is blocked while that design is under Construction;
- normal Production completion creates an independent concrete Ship and a new stationary one-Ship combat Fleet at the producing Colony's System;
- generic combat-Fleet movement/merge/split remains Slice 04; Command Points/ship-overage Maintenance remains Slice 05; hostile encounter handoff and tactical weapons/combat remain Slices 06-07;
- detailed component miniaturization beyond the first exact baseline remains deferred rather than being falsely approximated as fully original-equivalent;
- Core save/load, GameSession legal actions, Observer clone isolation and identical-session deterministic replay are covered.

Gate-3 QA already passed:

```text
go test ./internal/game -run Military -count=1
go test ./internal/session -run Military -count=1
go test ./internal/core ./internal/ruleset ./internal/moo2data ./internal/game ./internal/session -count=1
go test ./... -count=1
git diff --check
```

All temporary Gate-3 reverse-engineering/editor artifacts were removed. Gate 4 then passed the fresh session/repository conflict check, final gofmt/test/vet/focused compatibility regressions and working/staged diff checks; implementation commit `2670d36` is recorded and the closing documentation removes the recovery marker.
## Closed Outpost Ship / supply-range checkpoint

Slice 02 **Outpost Ship / Outpost state / supply range** is closed. Gate 4 passed and the runtime/tests/evidence are committed in `7c7284e`.

Closed runtime boundaries:

- Core schema 17 owns dedicated Planet-linked `Outpost` state rather than fake zero-Population Colonies;
- Outpost Ship is Technology 109, 100 PP base / 67 PP current Feudal, with installed best FTL speed persisted at completion;
- Colony Ship and Outpost Ship share the fixed-special-ship Fuel-range/transit path while generic military Fleet movement remains deferred;
- strategic supply origins are owned Colonies plus owned Outposts, and blockade alone does not suppress the range origin;
- `fleet.deploy_outpost` deploys on an unoccupied same-System Planet/body, consumes the ship and exposes the new supply origin immediately;
- Colony Ship and Colony Base founding atomically replace a same-owner Outpost; foreign Outpost conquest/destruction stays deferred;
- the unresolved original `star+0x37` control/UI gate remains deferred to the strategic hostile-encounter/control work rather than being invented here;
- schema-17 save/load, invalid-link validation, GameSession Observer isolation and deterministic repeated-session replay are covered.

Gate 4 passed fresh conflict/status inspection, `gofmt` verification, `go test ./... -count=1`, `go vet ./...`, focused Outpost/Colony Ship/Colony Base/Supply-range regressions and working/staged `git diff --check`.
## Closed Colony Base checkpoint

Slice 01 **Colony Base / same-system colonization** is closed. Gate 4 passed and the runtime/tests/evidence are committed in `261e418`.

Closed runtime boundaries:

- Colony Base remains normal Building 11 / `colony_base`, Tech 40, 200 PP and 0 BC Maintenance;
- buildability and direct queue validation require at least one empty Planet in the source Colony's StarSystem;
- Construction persists no target Planet; completed Bases are derived from Building ownership;
- `colony.colonize_with_base` creates a normal same-system Colony, consumes the Base and leaves source Population unchanged;
- `colony.trash_colony_base` consumes the Base and refunds exactly 100 BC, including targetless completed Bases;
- `CompleteTurn` is blocked until all completed Colony Bases are resolved;
- Colony Ship and Colony Base share one deterministic normal-Colony founding helper;
- PlayerView/ObserverView expose authority-correct pending resolution state;
- Core schema remains **16**, with exact save/load reconstruction and deterministic replay coverage.

Gate 4 passed `gofmt`, `go test ./... -count=1`, `go vet ./...`, focused Colony Base/Colony Ship/Building-choice regressions and both working-tree/staged `git diff --check`.
## Closed Colony Ship checkpoint

The first authoritative headless expansion loop is complete:

```text
home Colony
-> construct Tech-41 Colony Ship
-> create stationary civilian Colony Ship at producing System
-> issue Fuel-range-checked strategic move
-> advance deterministic ETA before current-turn Construction
-> arrive at target System
-> explicitly colonize an empty same-System Planet
-> consume Colony Ship
-> create/link second Colony with one assimilated founding Population unit
```

Direct MOO2 1.31 evidence and the accepted Gate-2 contract are preserved in `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md`.

Runtime boundaries now include:

- base Colony Ship cost 500 PP, with the current Feudal observed-rule cost of 334 PP;
- original Fuel Cell strategic ranges: Standard 4 pc, Deuterium 6, Iridium 9, Urridium 12 and Thorium 255;
- persisted Colony Ship special identity, installed FTL speed, destination and ETA in Core schema 16;
- at that Colony-Ship checkpoint, Colony-owned Systems were the only authoritative supply origins; active Slice 02 now extends this to owned Outposts;
- Colony Ship transit before current-turn Construction, so a newly produced ship cannot also move that turn;
- explicit colonization only after arrival; no automatic same-step colonization;
- second-Colony Planet link, founding cohort/economy recalculation and Colony Ship consumption;
- exact Core save/load plus GameSession Observer/replay coverage.

Gate 4 passed `gofmt`, `go test ./... -count=1`, `go vet ./...`, focused Colony Ship/transit/blockade/Population-transfer regressions and `git diff --check`.

## Historical prepared-queue checkpoint (before Slice 04)

The following block is retained as historical context from the pre-Slice04 handoff; all listed slices 04-07 are now closed.

Recommended later order:

- **Slice 04 - Generic combat Fleet movement / merge / split** - `docs/slices/PLANNED_04_COMBAT_FLEET_MOVEMENT_MERGE_SPLIT.md`
- **Slice 05 - Command Points / ship overage Maintenance** - `docs/slices/PLANNED_05_COMMAND_POINTS_SHIP_MAINTENANCE.md`
- **Slice 06 - Strategic hostile encounters -> BattleSession handoff** - `docs/slices/PLANNED_06_STRATEGIC_HOSTILE_ENCOUNTERS_BATTLE_HANDOFF.md`
- **Slice 07 - Tactical ship combat baseline** - `docs/slices/PLANNED_07_TACTICAL_SHIP_COMBAT_BASELINE.md`

Slice 06 is closed. Slice 07 has since closed via gameplay/data/evidence commit `889f977`; no later slice is currently open.
## Later deferred dependencies

- tactical ship combat state/commands, weapon resolution and tactical positioning;
- Transport/troop movement and invasion;
- Spy state/Maintenance;
- trade/research/Tribute treaty economics;
- Officer/Leader state, Maintenance and economic bonuses;
- complete original staged Treasury-deficit liquidation;
- active conquest/occupation/automatic-assimilation progression;
- Android/Native population and persisted custom race designs;
- AI colony targeting/auto-colonize policy and broader diplomacy;
- special/non-player Fleet owners.

Original copyrighted assets remain private reference material and are not distributable MOOX content. Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.
