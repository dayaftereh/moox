# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: `docs/slices/_OPEN_15_4_RICH_GAMEPLAY_PERSISTENCE_UX_2026-09-09.md`.
- Active implementation slice: **Slice 15.4 Gate 3**. Blocks 1-5 are complete; Block 6 Battle Entry/Return route + shell is next.
- Slices 01-12 are closed.
- Slice 10 implementation/data/evidence commit: `b91f65e` (`game: add diplomacy war peace baseline`).
- Core `StateSchemaVersion`: **23**.
- Economy ruleset schema: **8**.
- Slice-10 permanent evidence: `docs/research/DIPLOMACY_WAR_PEACE_BASELINE_2026-09-02.md`.
- Post-Slice-12 audit: docs/research/POST_MILESTONE_FIDELITY_DEPTH_BACKLOG_AUDIT_2026-09-03.md.
- Slices **13-14 and 15.1-15.2 are closed**; the **15.3 playability visual baseline is frozen/parked**; **15.4 Gate 3 is active with Blocks 1-5 complete**; 15.5-15.6, 16-17 and reserved Slice 20 remain prepared.
- Current objective: **Slice 15.4 Gate 3 Block 6 - Battle Entry/Return route and shell, preserving the 15.5 Tactical boundary.**
- Slice-15.1 closure evidence: `docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE4_2026-09-04.md`.
- Slice-15.2 Gate-1 evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE1_2026-09-04.md`.
- Slice-15.2 Gate-2 evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE2_2026-09-04.md`.
- Slice-15.2 Gate-3 evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE3_2026-09-05.md`.
- Slice-15.2 Gate-4 blocker remediation: `docs/research/SLICE_15_2_GATE4_BLOCKER_REMEDIATION_2026-09-05.md`.
- Prepared future Espionage mechanics: `docs/slices/PLANNED_20_ESPIONAGE_INTELLIGENCE_BASELINE.md`.
- Accepted downstream 15.2/15.3 strategic product direction: `docs/research/SLICE_15_2_STRATEGIC_HMI_PRODUCT_DIRECTION_2026-09-04.md` - 2D Galaxy primary view, orbital star-system dialog, legal Colonize confirmation, Colony table + full Colony Detail/build queue, location-grouped Fleets, eight-category Research, first-class Diplomacy and planned Espionage, plus retained OX lettermark / future spacecraft-space icon.

- Active Slice 15.3 Gate-1 audit: `docs/research/SLICE_15_3_GATE1_ART_ASSET_AUDIT_2026-09-07.md` - initial web asset inventory, catalog starting point and visual/pipeline questions.
- Vector/procedural graphics direction: `docs/research/SLICE_15_3_VECTOR_PROCEDURAL_GRAPHICS_DIRECTION_2026-09-07.md` - SVG-first icons/overlays plus deterministic procedural 2D ship geometry with a future 3D path.
- Shipbuilder generation UX: docs/research/SLICE_15_3_SHIPBUILDER_GENERATION_UX_2026-09-07.md - choose hull size -> repeatedly Generate deterministic 2D concepts -> Keep favorite -> later authoritative equipment handoff in Slice 17.
- Hull-scale/concavity grammar: docs/research/SLICE_15_3_HULL_SCALE_CONCAVITY_GRAMMAR_2026-09-07.md - Scout-to-Doom-Star perceptual scale progression, concave external bays and transparent negative-space cutouts.
- Directed-evolution / Proceduro reference: docs/research/SLICE_15_3_PROCEDURO_DIRECTED_EVOLUTION_REFERENCE_2026-09-07.md - six-candidate deterministic visual-genome evolution with sharp Wedge/Spike/Pod primitives and documented Unlicense provenance.
- Ship Style-DNA candidates: docs/research/SLICE_15_3_SHIP_STYLE_DNA_CANDIDATES_2026-09-07.md - Spear/Angular, Sleek/High-Tech and Organic/Alien geometry biases integrated into the same deterministic evolution genome.
- Ship Genome lock controls: docs/research/SLICE_15_3_SHIP_GENOME_LOCK_CONTROLS_2026-09-07.md - exact partial preservation of core hull, sharp primitives, engines and cutouts during directed evolution.
- Accepted Tactical roadmap amendment: `docs/research/SLICE_15_5_INTERACTIVE_TACTICAL_COMBAT_PRODUCT_DIRECTION_2026-09-04.md` - interactive Tactical is a dedicated Slice 15.5; the former final integration slice moves to 15.6.


## Post-Slice-12 fidelity/depth audit - complete

The milestone re-audit identified independent playability as the main post-lifecycle bottleneck. Slices 13 Built-in AI and 14 live GameSession persistence are closed. Slice 15 is now a six-part browser-playability family: 15.1 UX/navigation/design system, 15.2 functional strategic HMI, 15.3 MOOX visual identity/assets, 15.4 rich gameplay decisions/persistence UX, **15.5 interactive 2D Tactical Combat**, 15.6 final browser vertical slice. Slice 16 New Game/preset-race breadth and Slice 17 military design breadth remain later independent work. Slices 15.1 and 15.2 are closed with Gates 1-4 complete. Slice 15.3 reached its frozen playability visual baseline and is parked for richer-art follow-up; Slice 15.4 Gate 3 is active with Blocks 1-5 complete.

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

- Submitted-seat Planning-preview fix: docs/research/SLICE_15_3_SUBMITTED_PLANNING_PREVIEW_ERROR_FIX_2026-09-07.md - prevents post-submit preview calls and maps expected preview rejection away from HTTP 500.

- Full-random Ship Morphology v3: docs/research/SLICE_15_3_FULL_RANDOM_SHIP_MORPHOLOGY_V3_2026-09-07.md - primary single-ship reroll UX with eight topology families spanning Needle through Bulb/Asymmetric.

- Symmetric half-hull / class footprint v4: docs/research/SLICE_15_3_SYMMETRIC_HALF_HULL_SCALE_V4_2026-09-07.md - strict X-axis mirroring, asymmetric morphology removed from default pool, Space values surfaced, and Scout-to-Doom-Star visible footprint scaling.

- Visual Genome v4 server persistence: docs/research/SLICE_15_3_VISUAL_GENOME_V4_SERVER_PERSISTENCE_2026-09-07.md - authoritative resolved genome, separate VisualRevision, immediate save, built-ship freeze and save/export/import round-trip contract.

- Core SVG icon language implementation: docs/research/SLICE_15_3_CORE_SVG_ICON_LANGUAGE_IMPLEMENTATION_2026-09-07.md - typed 24x24 React/SVG registry, shell/resources/navigation/population-role integration and desktop/mobile QA.

- Research/population/action icon refinement: docs/research/SLICE_15_3_ICON_REFINEMENT_RESEARCH_ACTIONS_2026-09-07.md - microscope Research identity, test-tube category primitive, role-specific population markers and selective leading action icons.

- Research-category and Galaxy marker integration: `docs/research/SLICE_15_3_RESEARCH_GALAXY_ICON_INTEGRATION_2026-09-07.md` - eight category SVG identities plus strategic star/colony/outpost/fleet/contact markers and desktop/mobile QA.

- Body/Fleet/Construction/Diplomacy icon pass: `docs/research/SLICE_15_3_BODY_FLEET_CONSTRUCTION_DIPLOMACY_ICON_PASS_2026-09-07.md` - orbital body/settlement SVGs, fleet-role semantics, construction/building family, queue arrows and diplomacy stance/actions.
- Gate-1 visual consistency audit: `docs/research/SLICE_15_3_GATE1_VISUAL_CONSISTENCY_AUDIT_2026-09-07.md` - cross-screen 791px/320px audit and placeholder sweep.
- Gate-1 styleboard candidates: `docs/research/SLICE_15_3_STYLEBOARD_CANDIDATES_2026-09-07.md` - Deep Space Command vs Orbital Glass vs Industrial Tactical; Candidate A recommended.
- Gate-2 visual/asset contract draft: `docs/research/SLICE_15_3_GATE2_VISUAL_ASSET_PIPELINE_CONTRACT_DRAFT_2026-09-07.md` - palette/icon grammar, responsive/raster targets, semantic IDs, provenance/license, fallback and procedural-ship contract.

- Art-fidelity planets/buildings follow-up: `docs/research/SLICE_15_3_ART_FIDELITY_PLANETS_BUILDINGS_2026-09-08.md` - spherical climate/noise/cloud planets, rich runtime building/station illustrations, Art Fidelity Lab and future surface reuse boundary.

- Spectral stars / building coverage / Construction layout review revision: `docs/research/SLICE_15_3_SPECTRAL_STARS_BUILDING_COVERAGE_LAYOUT_FIX_2026-09-08.md` - luminous class-colored StarArt, all-48 building visual routing, rich Housing project art and corrected rich-project grid sizing.
- Slice 15.3 playability visual baseline freeze: `docs/research/SLICE_15_3_PLAYABILITY_VISUAL_BASELINE_FREEZE_2026-09-09.md` - current Deep Space Command/StarArt/OrbitalBodyArt/BuildingArt/Visual Genome v4 baseline frozen so 15.4-15.6 can pursue playability; richer art is parked, not cancelled.
- Slice 15.4 Gate-1 rich-state UX audit: `docs/research/SLICE_15_4_GATE1_RICH_STATE_UX_AUDIT_2026-09-09.md` - phase/DecisionView gaps, Encounter/Invasion/Colony Base/Research boundaries, persistence/reconnect/error audit.
- Slice 15.4 Gate-2 rich interaction draft: `docs/research/SLICE_15_4_GATE2_RICH_INTERACTION_CONTRACT_DRAFT_2026-09-09.md` - blocking decisions, Battle handoff, local Save/Load, lifecycle/error contract and minimum player-safe reads.
- Slice 15.4 Gate-3 block 1: `docs/research/SLICE_15_4_GATE3_BLOCK1_CLIENT_AUTHORITY_LIFECYCLE_2026-09-09.md` - typed Colony Base/Battle contracts, structured APIError and explicit authoritative connection lifecycle.
- Slice 15.4 Gate-3 block 2: `docs/research/SLICE_15_4_GATE3_BLOCK2_SAVE_LOAD_RESTORE_UX_2026-09-09.md` - local Save/Load, explicit atomic Restore, fresh-server Import and desktop/320px E2E evidence.
- Slice 15.4 Gate-3 block 3: `docs/research/SLICE_15_4_GATE3_BLOCK3_COLONY_BASE_UX_2026-09-09.md` - mandatory Post-Resolution Colony Base target/Scrap UX with isolated real-session E2E for both outcomes.
- Slice 15.4 Gate-3 block 4: `docs/research/SLICE_15_4_GATE3_BLOCK4_INVASION_UX_2026-09-09.md` - public empire identity + rich blocking Invasion context with real Decline/Invade E2E and 320px QA.
- Slice 15.4 Gate-3 block 5: `docs/research/SLICE_15_4_GATE3_BLOCK5_RESOLUTION_SUMMARY_RESEARCH_RESULT_2026-09-09.md` - stable player-safe event-derived resolution summaries, automatic Research Breakthrough presentation and dedicated read-only Victory result surface.
- Slice 15.4 Gate-3 blocker remediation: `docs/research/SLICE_15_4_GATE3_ENDTURN_BUILTIN_AI_BLOCKER_2026-09-09.md` - browser New Game now assigns Seat 2 to Built-in AI; canonical game-1 was repaired without losing the submitted construction/research turn and verified at Turn 2 Planning.
- Slice 15.4 Gate-3 Research UX correction: `docs/research/SLICE_15_4_GATE3_RESEARCH_UX_TECH_INFO_2026-09-09.md` - one canonical Research chooser now serves HUD, route and Breakthrough flow; each Technology has a `?` detail dialog backed by normalized runtime effects, with unnormalized details explicitly marked instead of guessed.
- Slice 15.4 Research help: all 203 normalized Technologies now carry source-backed HELP.LBX descriptions in the `?` dialog, explicitly separated from current MOOX runtime effects.
- Slice 15.4 building detail UX: construction choices now show authoritative cost/maintenance, zero-or-more normalized MOOX building effects, and the linked source-backed MOO2 description directly in the existing build panel.
- Slice 15.4 Gate 3 Block 6: stable participant-safe Battle entry/return route is implemented; unresolved human Battles block strategic navigation, supported Tactical specs hand off to a 15.4 shell, and final results render from authoritative `battle_completed` summaries even after the live BattleView is released.
- Slice 15.4 Gate 3 Block 7: stale 409 recovery now refetches consistently without auto-resubmit, refresh-failed has an explicit read-only Retry flow, planning-preview errors no longer erase command rejections, and Battle/Research/Construction/Persistence were checked against the 320px responsive/touch contract; Gate 3 is complete pending Gate-4 closure.
- Slice 15.4 is CLOSED (Gates 1-4 complete). Gate-4 closure confirms browser phase coverage, authoritative Battle handoff/return, persistence/reconnect/stale-client behavior, non-destructive rejection handling, responsive floor and full regression/build gates. Next queued slice: 15.5 Interactive Tactical Combat.
- Slice 15.5 is OPEN. Gate 1 Tactical mechanics/UX audit is complete: original movement is integer X/Y with 16 facings and `ceil(Euclidean distance) + facing steps * turn cost` (2 normal / 1 Inertial Stabilizer / 0 Inertial Nullifier); first credible breadth is recommended as 1-2 ships per side with 2v2 acceptance. Gate-2 Tactical contract is DRAFT and awaiting user acceptance before implementation.
- Slice 15.5 Gate-2 draft amendment: Tactical Scan is now proposed as a free read-only inspect mode for any player-visible ship (weapons/readiness, Armor/Structure damage, movement/facing and normalized combat ratings); Battle Scanner technology is not required for the generic Scan UI. Battlefield presentation is proposed as effectively unbounded: no ordinary fixed width/height or visible arena edge, while finite movement budget keeps each authoritative legal-move projection bounded.
- Slice 15.5 Gate 2 is FROZEN by explicit user acceptance (2026-09-10): 1v1-2v2 breadth, 16-way facing/original movement cost, server-projected legal moves/targets, `battle.move_ship`, Laser baseline, Tactical Scan, effectively unbounded battlefield, desktop/mobile semantics, and 15.4 stale/reconnect authority rules. Gate 3 implementation may proceed.
- Slice 15.5 Gate 3 Block 1 complete: BattleSession now owns mutable X/Y/Facing/Movement Points, exact original movement cost, `battle.move_ship`, deterministic legal moves, 2v2 activation/round reset and moved-position Laser range; strategic metadata supports 1v1-2v2 while all-unarmed and >2-per-side Battles remain explicit unsupported.
- Slice 15.5 Gate 3 Block 2 complete: participant Battle snapshots redact Tactical RNG and opponent command catalogs while projecting read-only Scan ship details, own legal moves/Lasers/targets and typed web Battle command helpers; HTTP regression covers projected move -> submit -> refreshed position -> projected fire -> submit.
- Slice 15.5 Gate 3 Block 3 complete: the Tactical placeholder is now a real responsive SVG battlefield driven only by participant-safe projections. Real 2v2 browser QA covered strategic entry, Move (20->19 MP), enemy Scan, Laser damage (Armor 4->0), activation progression, Built-in-AI turns, Round 2 reset, pan/zoom/pinch and a clean 320px/44px touch layout.
- Slice 15.5 Gate 3 Block 4a fixed the full-battle damage blocker found by browser QA: beam damage now exhausts Armor then continues into aggregate Structure; deferred internal-selection RNG is still consumed but no longer aborts or fabricates subsystem damage. Battle/session/server regressions pass.
- Slice 15.5 Gate 3 Block 4b complete: Tactical uses accepted 15.3 procedural ship glyphs with SVG fallback, and an isolated real 2v2 browser battle now runs from normal End Turn through multiple rounds, destruction of both enemies, Slice-15.4 Battle Return, authoritative destroyed [60,61]/survivor [57,58] summary, and Turn 2 Planning.
- Slice 15.5 Gate 3 Block 5 complete: Tactical stale/illegal commands now surface an in-Battle rejection while retaining shared conflict refetch, never auto-resubmit, and preserve authority. Browser QA proved 1 POST/1 refetch/zero mutation for stale move, occupied move and illegal own-ship fire; server regression reproduces all three.
- Slice 15.5 is CLOSED (Gates 1-4 complete). Independent Gate-4 validation confirms the server-authoritative 1v1-2v2 Tactical baseline, movement/facing, Laser/Armor->Structure destruction, Scan, procedural ship visuals, stale/illegal recovery, responsive open-field UI and exact Tactical->strategic result reconciliation. Next queued slice: 15.6 Full browser vertical slice / polish / complete-game QA.
- Slice 15.6 Gate 1 integration-readiness audit complete: 15.1-15.5 compose into the planned Main Menu -> New Game -> strategic/persistence -> Tactical -> conquest journey with one concrete browser blocker: the already-authoritative narrow Frigate none|1x-Laser military-design command is not exposed by ShipBuilderView. Seed0x8009 starts with Laser tech100, Military Ship/Troop Transport construction is already available, and the existing canonical headless conquest pins the 12-pc research/supply route; complete Human-vs-Built-in-AI browser conquest remains the 15.6 evidence target. Gate2 contract is DRAFT awaiting user acceptance.
- Slice 15.6 Gate 2 FROZEN by explicit user acceptance after durable Ship Designer refinement: name-first design, left/right hull stepper Frigate->Doom Star, server-authoritative locks, Available/Installed lists, future-ready slot/count rows, authoritative Production/CP/space, persistent named design catalog -> exact Colony construction revision, and click/tap SVG visual reroll that changes no gameplay state. Gate3 may implement in small recovered blocks.
- Slice 15.6 Gate 3 Block 2 implemented: durable server-authoritative Ship Designer shell now supports named design library, left/right hull stepping with locks, click/tap SVG reroll, Available/Installed component presentation, authoritative PP/CP/space, and real Frigate none|1xLaser save. Focused browser QA created `Falcon Mk I` and verified exact ID/revision/name handoff into Colony construction choices; full tests/vet/web build pass. Next: visible Colony queue handoff in a fresh recovered block.
- Slice 15.6 Gate 3 Block 3 verified the full visible Ship Design -> Colony handoff: saved `Falcon Mk I` appears separately from Scout in Colony Construction at 30 PP with exact r1 identity, can be added to the normal planning queue/preview, and the formerly disabled `Schiff entwerfen` placeholder now opens that exact design by route ID (`shipbuilder/80`). Web build and managed-Chrome QA pass; next small block is canonical 7171 review refresh.
- Slice 15.6 Gate3 Block4B user-review fix: Colony Construction now renders exact saved Military Ship procedural SVG artwork from design ID/revision in catalog + selected-project hero; obsolete Planning Preview lifecycle/revision conflicts are suppressed while genuine draft errors remain visible. Production web build passed; canonical 7171 browser QA follows after commit/service refresh.

- Slice 15.6 Gate3 Block5D COMPLETE (`ee9f60e`): the compact Galaxy picker now supports `Ziel wÃ¤hlen` plus direct drag/drop onto star systems. Legal/blocked highlighting, distance, range, ETA and rejection reasons come exclusively from authoritative `fleet_move_targets`; clicking a blocked anonymous target does not open SystemDialog or leak its name and stages no order. A legal common destination stages one `empire.move_fleet` draft per selected source profile using the exact projected `fleet_id`/optional `ship_ids`. Canonical `0x8009` visible QA proves 19 blocked targets and `8 pc Â· Reichweite 4 pc Â· ETA 4`, plus equivalent drag rejection and empty-map reposition behavior. Web build, focused movement tests and full Go regression are green. Next: Block5E deterministic 2 pc/5 pc micro fixture to visibly prove the positive green-target -> staged move -> arrival -> first-contact -> Tactical path.
- Fleet-marker correction (`7131046`): authoritative `Empire.PlayerColorSlot` is persisted for new games and compatibly projected for legacy saves; Galaxy presence is grouped by empire/player. Canonical Human home with Combat fleet #59 plus Colony fleet #63 therefore renders exactly one blue Human fleet icon while the common picker still contains Scout 1, Scout 2, and Colony Ship.
- OPEN special-ship visual persistence: Colony Ship, Outpost Ship, and Troop Transport still lack authoritative persisted per-instance `visual_genome` data because they are modeled as fixed `StrategicFleet.SpecialKind` units rather than concrete `Ship` instances. `SpecialShipGlyph.tsx` is interim only. Future work must persist a stable genome at creation, round-trip it through save/load/reconnect, project it to clients, deterministically backfill older saves, and prove the same visual identity from construction completion through Galaxy/System views. Do not close this item merely because static SVG silhouettes exist.
- Slice 15.6 Gate3 Block5D UX refinement (`def7af0`): fleet-picker open is now automatic target mode; the explicit `Ziel wÃ¤hlen` control is removed. Green/red targets recompute immediately from authoritative `fleet_move_targets` as selection changes. Each unit tile has a sibling `?` info control with a floating technical-detail overlay. `SystemDialog -> Flotten / Schiffe` routes to the same compact picker and closes the SystemDialog; the old large System fleet modal is no longer player-accessible. Canonical visible QA confirms 19 immediate blocked targets, unchanged selection when opening Scout/Colony info, 292x202 base picker geometry, and the same anonymous 8pc/range4/ETA4 rejection. Next remains Block5E deterministic positive 2pc/5pc movement -> arrival -> first contact -> Tactical acceptance.
- Slice 15.6 Gate3 Block5D follow-up (`43e8035`, `f6fc6a5`): physically removed the obsolete large System Fleet dialog. Special strategic unit `?` details now expose server-projected effective drive, FTL speed, Fuel Cell and parsec range. Per-instance special-fleet FTL remains authoritative; Fuel Cell/range intentionally follow current empire fuel technology, preserving existing upgrade semantics. Projection-only metadata is rejected in authoritative GameState. Full Go regression + web build green.
- Slice 15.6 Gate3 Block5E reference environment (`819880a`, `52523b2`): replaced the disposable micro-fixture idea with two durable dev reference games. `game-1` stays canonical `0x8009`/20-star Human-vs-Darlok and preserves the initial 0/19 blocked movement case. New `game-triangle-2pc` / `triangle-2pc-v1` uses Human + Darlok AI + Psilon AI on three symmetric home systems whose every authoritative edge is exactly 2 pc. Initial Human picker visibly shows 2/2 green anonymous destinations at range4/ETA1; clicking one stages the expected two source-fleet movement orders without submitting the turn. `-reference-games` now bootstraps both on development server start. Next Block5E work can use triangle for actual submit -> arrival -> visit/contact -> blue/red/third-color markers -> Tactical while retaining game-1 for normal-generator regression.
- Slice 15.6 Gate3 Block5D Fleet route lifecycle: pre-`Fertig` move drafts are editable and visualized as grouped green dashed routes; blocked clicks show a red dashed diagnostic route and never alter a valid order; source click cancels only pre-submit drafts. Once `Fertig` starts transit, the flight is immutable until arrival. Authoritative Fleet transit now persists source/destination/remaining/total turns, player projection supplies server-derived total/remaining pc, and the clickable transit marker is status-only (remaining pc + ETA, no cancel/retarget). `game-1` red-route visual QA and combat/special in-transit immutability tests are green.
- Slice 15.6 Gate3 Block5D Fleet-marker route anchor/refinement: green planned and red blocked lines now start at the own Fleet marker rather than the star center. A started Fleet opens the compact Fleet dialog in read-only transit mode: normal unit glyphs and `?` ship details remain inspectable, but selection/Select-All/target/retarget/cancel controls are absent. Transit status continues to use server-projected remaining pc + turns/ETA and hidden destination identity remains anonymous.
- Slice 15.6 Gate3 **Block5F queued after 5E**: foreign Fleet detection / scanner intelligence. Extend the new transit marker/route/read-only Fleet dialog to scanner-authorized foreign traffic: foreign player-colored marker, red dashed incoming route to permitted targets, server-projected remaining pc/ETA, and intelligence-redacted ship details. Current ruleset evidence includes Space Scanner 1pc, Tachyon 3pc, Neutron 5pc base detection plus transit size-class language, and Battle Scanner +2pc galactic range. Exact stacking, scanner sources, size-class contribution, last-known contacts and deep-space grouping require research before implementation. See `SLICE_15_6_GATE3_BLOCK5F_FOREIGN_FLEET_SCANNERS_PLAN_2026-09-12.md`.
- Slice 15.6 Gate3 triangle regression (`5c2c569`): fixed a built-in AI research-selection bug exposed by the 3-player reference game. Psilon (`ResearchSelectionAll`) incorrectly submitted a concrete `technology_id`, causing strategic resolution to reject the turn after Human clicked `Fertig`. Symptom: phase stuck in `strategic_resolution`, planning preview disappeared so the queued Laser Scout looked empty, and `Fertig` became unavailable. The Human Scout Rev.2 + Laser + construction order itself was valid and remained in the submitted batch. Baseline AI now only supplies `technology_id` for `choose_one`; `all`, `fixed_one`, and `repeat_field` leave it omitted as required by server authority. End-to-end host regression reproduces Scout r1 -> Laser Rev.2 -> queue -> submit with Darlok/Psilon built-in AI and proves the turn leaves strategic_resolution with the Rev.2 construction committed. Visible browser QA reached Round 2 Planning with Scout 6/30 PP, 24 PP remaining, ETA 4 and `Fertig` available again.
- Slice 15.6 Gate3 **Block5E COMPLETE** - co-located foreign Fleet inspection/Tactical expansion (`3b2d734`): Triangle Scouts remain normal and unarmed. Known foreign Fleets at a visited shared system now project a read-only Fleet composition and open from their Galaxy marker with ship/special-vessel `?` inspection. `Angreifen` is available directly there and invokes authoritative declare-war; encounter creation remains on the next `Fertig` resolution boundary. Tactical accepts 3v2+ combat Ship counts, permits unarmed combat Ships, ignores CivilianFleetIDs/colony context when materializing the ship battlefield, and adds an explicit Retreat action. Civilians stay strategic and never become Tactical ships. Clean Round1 browser acceptance reached Round2 shared Human Home, inspected Darlok Scout1/Scout2/Colony Ship, declared attack in-place, then created a supported 2v2 all-unarmed Tactical battle with exactly four Scout tactical ships and a visible RÃ¼ckzug control. Remote scanner/redacted Fleet intelligence remains Block5F. **User browser acceptance 2026-09-12:** the Triangle scenario was played end-to-end through the previously blocked larger-Fleet encounter path and successfully entered the Battle/Tactical UI. This closes Block5E as the Tactical-entry enabler. The next active review focus is 5A Tactical UI; Block5F remains queued behind that review.
- Slice 15.6 Gate3 Tactical UI refinement review OPEN (user shorthand 5A Tactical UI): Block5E now supplies the real Triangle strategic-to-Tactical entry path, including the larger-Fleet regression. Preserve the closed Slice15.5 authority baseline while reviewing battlefield hierarchy, ship/selection readability, move/fire/initiative/damage affordances, action panels, camera/framing, desktop/mobile ergonomics and battle transitions. Final acceptance should use the real Triangle encounter; Block5F remains queued. Plan: SLICE_15_6_GATE3_TACTICAL_UI_REFINEMENT_REVIEW_2026-09-12.md. **Tactical UI Block1 immersive shell IMPLEMENTED / browser-QA green:** full-screen Battle/Tactical shell, floating game menu only, strategic chrome hidden, 1x1 movement cells with 4x4 major grid, green server-projected reachable cells, bottom Ship/action HUD and Fleet-style Scan popover. Current Tactical commands still submit immediately; staged activation commit and fire-all remain explicit authority follow-ups, not UI-only behavior. **Tactical UI Block2 implicit interaction IMPLEMENTED / browser-QA green:** Move/Fire mode buttons are removed; legal grid click moves, legal enemy Ship click fires the selected authoritative weapon, Scan is the only explicit temporary mode, the HUD is reduced to ~94px in the QA viewport, the 1x1/4x4 grid is materially stronger, and camera input is wheel + left/right drag + one-finger pan + two-finger pinch with drag-safe click suppression. Exact 3v2 destructive QA ran only on an isolated 7173 clone of the preserved 7171 snapshot. **Tactical UI Block3 persisted ship/cell fidelity IMPLEMENTED / browser-QA green:** Tactical now transports persisted Ship visual genomes and source-design identity; persisted ships render the exact saved SVG genome, legacy ships use the same Fleet/System design seed, all ship art is constrained to one 1x1 cell and scaled by shared hull footprint. Permanent grid rendering is removed. Selecting the active own Ship toggles one blue current cell, only server-authoritative legal destinations in green, and all other occupied ship cells in red; occupied cells never appear green and diagonal legal movement remains available around blockers. Browser QA used isolated 7174; canonical 7171 remained untouched. **Tactical UI Block4 automatic activation/motion IMPLEMENTED / browser-QA green:** the duplicate local movement-selection toggle is removed; server `active_ship_id` now immediately drives the blue active cell, all green server-legal destinations and red occupied cells, with automatic camera recenter on active-Ship changes. Field Ship names are removed and the own-Ship strip is icon-only SVG. New authoritative `beam_fired` events drive visible beam/impact animation and `ship_moved` events drive a movement trail/arrival pulse. Isolated 7174 QA proved immediate 1 blue + 514 green + 4 red on entry, event-driven 3-damage laser animation, movement 10,13 -> 11,13 with trail, and automatic Round3 activation transfer to Ship23 at 14,12 without an extra click.
- Slice 15.6 Gate3 planning-draft persistence IMPLEMENTED / browser-QA green: strategic draftOrders are now server-hosted per game/seat/turn/base revision with monotonic draft_revision ordering, player-snapshot rehydration, stale/out-of-order protection and automatic clear on successful Fertig. Generic coverage includes colony population, construction queues, research and fleet moves; Ship Designer stays immediate-authoritative and diplomacy/war stays outside until explicitly redesigned. Browser QA on isolated 7176 reproduced the original reload bug and proved 5 mixed draft orders survive hard reload then clear after turn submission. Plan: SLICE_15_6_GATE3_PLANNING_DRAFT_PERSISTENCE_2026-09-12.md.
- Slice 15.6 Gate3 Tactical mobile performance / ship readability / damage feedback IMPLEMENTED / browser-QA green: per-cell reachable rectangles removed in favor of one crisp green SVG grid path over server-projected legal_moves; click-to-cell remains authority-gated. Original MOO2 drive evidence was rechecked after user feedback: pristine Frigate Nuclear/Fusion combat speed is 20/22, while 10/12 is the engine-damage minimum, so the temporary MinSpeed baseline has been reverted. Tactical ship art remains centered native nested SVG. New-game starting Scout designs now receive a fully resolved persistent ShipVisualGenome v4 at creation (VisualRevision 1), and starting ships clone that exact genome rather than relying on the browser fallback. Isolated fresh reference-game QA on 7178 confirms Triangle Human sleek/bulb, Darlok spear/fork, Psilon organic/manta and Game One Human spear/hammer Scout genomes.

- Slice 15.6 Gate3 Tactical direct-selection / hull-scale / Finish / damage-dwell follow-up IMPLEMENTED / isolated real-battle browser-QA green: own ship/cell click selects and focuses; selected 1x1 cell is filled blue; other occupied live ship cells are red outlines; authoritative legal_moves remain one green grid path only for the server-active selected ship; all selection/target circles removed. Tactical art resolves persisted visual hull before gameplay hull and uses shared one-cell footprints Scout .40, Frigate .48, Destroyer .58, Cruiser .68, Battleship .78, Titan .90, Doom Star 1.00 plus geometry-derived tight SVG viewBox. Remaining-ships HUD contains only unfinished live own ships. Fertig is HMI text for existing battle.end_activation, not a client-side turn switch. Damage-number dwell is 4s with feedback state retained 4.2s. Real Triangle Battle #1 QA on 7179: Scout 20/20, 515 legal moves, 1 blue cell, 3 red occupied outlines, 1 green path, 0 rings; selecting inactive Scout hides legal grid without changing authority; Fertig advances authoritatively and auto-selects next Scout. See SLICE_15_6_GATE3_TACTICAL_SELECTION_SCALE_FINISH_DAMAGE_DWELL_2026-09-12.md.

- Slice 15.6 Gate3 Tactical retreat confirmation guard IMPLEMENTED / isolated real-battle browser-QA green: canonical Triangle loss diagnosed as an actual Human battle.retreat command followed by no_retreat_destination destruction, not a Finish-button mapping bug. Rückzug now opens an in-game confirmation modal first; no Battle command is sent until Rückzug bestätigen. Warning explains that retreat ends the battle immediately and a fleet is lost when no valid retreat destination exists. Abbrechen/Escape/backdrop are zero-mutation exits; Abbrechen receives initial focus. Isolated Battle #2 on 7180 proved first-click event/revision stability and confirm-only side_retreated. See SLICE_15_6_GATE3_TACTICAL_RETREAT_CONFIRM_GUARD_2026-09-13.md.

- Slice 15.6 Gate3 Tactical Wait/Done authoritative reordering IMPLEMENTED / isolated current-battle browser-QA green: added server command battle.wait_activation with optional explicit friendly target; Wait changes only active_ship_id and preserves movement, weapon readiness and unfinished state. Tactical view projects can_wait_activation + wait_target_ship_ids. Direct click on an unfinished own ship uses the same authority command; Warten auto-selects the next unfinished friendly ship. Fertig remains the only activation_complete operation. End-activation now searches circular initiative order so out-of-order Wait cannot start a new round while any live ship remains unfinished. Browser QA from a copy of canonical Battle#1: round2 reset all Human ships to 20/20 and Laser ready; Scout23 moved to 18/20, Wait -> Scout24, direct click back -> Scout23 still 18/20, Finish completed only Scout23, direct click Laser35 switched authority and exposed Laser Cannon with two legal targets. Native Tactical tap-highlight disabled. See SLICE_15_6_GATE3_TACTICAL_WAIT_DONE_REORDER_2026-09-13.md.

- Slice 15.6 Gate3 Tactical automatic activation completion IMPLEMENTED / isolated current-battle browser-QA green: after a successful Tactical move or Beam fire, the server now marks the active ship complete automatically iff movement_current <= 0 AND no runtime weapon remains ready. Unarmed ships therefore auto-finish at 0 movement; armed ships at 0 movement stay active while any weapon is still ready and auto-finish only after the final ready weapon is fired. Emits activation_auto_ended tied to the triggering command sequence and uses the same next-ship/new-round progression as explicit Done. Isolated Triangle browser QA: round2 Scout23 used a legal exact-cost 20 move to 0/20, emitted ship_moved + activation_auto_ended sequence10, became complete, and Scout24 automatically became active at 20/20 without clicking Fertig. See SLICE_15_6_GATE3_TACTICAL_AUTO_FINISH_2026-09-13.md.

- Slice 15.6 Gate3 Planning-draft revision rebase IMPLEMENTED / exact Triangle browser regression green: planning drafts no longer disappear when immediate mutations such as Ship Builder gameplay/visual saves advance the game revision during the same Planning turn. Host mutation rebases + revalidates stored drafts; stale-but-valid async draft saves are rebound to current revision and validated, while invalid stale drafts cannot overwrite a valid draft. Exact browser flow 5/3/0 population -> Laser Scout r2 -> Scout + Colony Base -> browser Back -> reload -> fresh reconnect preserved both population and construction orders. See SLICE_15_6_GATE3_PLANNING_DRAFT_REBASE_2026-09-13.md.
