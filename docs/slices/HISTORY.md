# Completed work slices

This ledger is the compact audit trail for **closed** Master of Orion X implementation slices. It complements the permanent evidence documents under `docs/research/` and prevents completed work from remaining accidentally listed as pending in live roadmaps.

A row belongs here only after implementation/documentation/QA has been closed. An unfinished slice is represented instead by `docs/slices/_OPEN_*.md`.

## Research and Economy sequence

| Slice | Result | Closing commit | Permanent evidence / checkpoint |
| --- | --- | --- | --- |
| Race-aware multi-Technology research | General/ordinary/Creative/Uncreative application policies and shared legal actions implemented | `f36ab5a` | `docs/research/RESEARCH_MULTI_TECH_2026-08-27.md` |
| Research project switching | Active project can change with complete accumulated RP transfer and Observer/replay history | `5d2f983` | `docs/research/RESEARCH_SWITCHING_2026-08-27.md` |
| Strategic turn order | Direct executable evidence fixed Research -> Growth/Starvation -> Construction with pre-growth RP/PP snapshots | `30e2b6d` | `docs/research/TURN_ORDER_2026-08-28.md` |
| Uncreative initialization | Fixed applications generated during New Game from the shared New Game RNG | `945f777` | `docs/research/UNCREATIVE_INITIAL_SELECTION_2026-08-28.md` |
| Hyper-Advanced repeated research | Per-Empire repeat levels, dynamic strategic cost and repeat completion implemented; original UI preview off-by-one documented but not copied into the core | `212632a` | `docs/research/HYPER_ADVANCED_RESEARCH_2026-08-28.md` |
| Advanced-start technology ownership | Original-style randomized/race-aware Advanced start implemented with shared New Game RNG and exactly 19 weighted extra grants | `f80b9bc` | `docs/research/ADVANCED_START_RESEARCH_2026-08-28.md`, `docs/research/ADVANCED_START_CHOOSER_2026-08-28.md` |
| External Technology grants / Uncreative repair | Server-authoritative grant transition implemented; incomplete TechFields are preserved and Uncreative fixed choices are repaired through authoritative RNG | `6258ac5` | `docs/research/UNCREATIVE_EXTERNAL_REPAIR_2026-08-28.md` |
| Construction project generalization | Construction state generalized from Building-only identity to semantic project kind/id as prerequisite for non-building projects | `3724275` | architecture/runtime checkpoint in current status docs |
| Freighter Fleet construction | 50 PP project adds 5 Empire Freighters through authoritative Construction legal actions/events | `9ae1ef8` | `docs/research/FOOD_FREIGHTER_LOGISTICS_2026-08-27.md` |
| Treasury settlement | 50 BC New Game Treasury and modeled strategic settlement implemented before Research | `2861150` | `docs/research/TREASURY_SETTLEMENT_2026-08-28.md` |
| Insufficient-Freighter priority | Temporary proportional allocation replaced by original-style round-robin deficit allocation | `b972ee0` | `docs/research/INSUFFICIENT_FREIGHTER_PRIORITY_2026-08-28.md` |
| System blockade -> Food logistics | Authoritative blockaded-Empire system state excludes Colonies from Food import/export/sale pools | `0c699a8` | `docs/research/BLOCKADE_FOOD_LOGISTICS_2026-08-28.md` |
| Population relocation / shared Freighters | Same-system immediate moves and interstellar Settler reservations/ETA/arrival/loss handling implemented; Core schema advanced to 11 | `f816eaf` | `docs/research/POPULATION_TRANSPORT_FREIGHTER_2026-08-29.md` |
| Population growth building / medicine modifiers | Housing continuous Production mode, Microbiotics +25pp, Universal Antidote +50pp and flat Cloning Center +0.1 Population/turn implemented; Core schema advanced to 12 | `e7f0d72` | `docs/research/POPULATION_GROWTH_MODIFIERS_2026-08-29.md` |
| Population capacity transitions | Advanced City Planning +5, Biospheres +2, semantic Terraforming/Gaia climate transformations and immediate capacity-clamp invariant implemented; Core schema advanced to 13 | `25d05a9` | `docs/research/POPULATION_CAPACITY_TRANSITIONS_2026-08-29.md` |
| Race-aware Population cohorts | Core schema 14 organic cohorts, source-race Economy, exact four-pass Food priority, per-origin Growth/Starvation, heterogeneous capacity and cohort-aware Population transfer implemented | `7c49e90` | `docs/research/POPULATION_COHORTS_2026-08-29.md` |
| Strategic Fleet/Diplomacy blockade production | Core schema 15 minimal strategic Fleets/directed relations and deterministic Colony-presence + hostile-combat-Fleet system blockade production implemented | `d148502` | `docs/research/FLEET_DIPLOMACY_BLOCKADES_2026-08-29.md` |
| Treasury Maintenance categories / deficit boundary | Six original Maintenance buckets and staged deficit policy resolved; active-Freighter reservations/usage and surplus-Food whole-BC boundaries implemented without schema changes | `a1ad15b` | `docs/research/TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md` |
| Colony Ship production / strategic movement / colonization | Core schema 16 Tech-41 Colony Ship Construction, original Fuel ranges, deterministic pre-Construction transit and explicit second-Colony creation/ship consumption implemented | `db9d01e` | `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md` |
| Colony Base / same-system colonization | Building 11 / Tech 40 same-system buildability, mandatory post-Production colonize-or-trash resolution, 100 BC refund and shared normal-Colony founding implemented without schema change | `261e418` | `docs/research/COLONY_BASE_SAME_SYSTEM_COLONIZATION_2026-08-31.md` |
| Outpost Ship / Outpost state / supply range | Core schema 17 dedicated Outposts, Tech-109 Outpost Ship construction/transit/deployment, Colony+Outpost supply origins and same-owner Outpost-to-Colony replacement implemented | `7c7284e` | `docs/research/OUTPOST_SHIP_SUPPLY_RANGE_2026-08-31.md` |

## Meta workflow

| Slice | Result | Closing commit | Permanent documentation |
| --- | --- | --- | --- |
| Recovery-safe slice protocol | Four-gate workflow and `_OPEN_*.md` recovery marker convention established | `204ced7` | `docs/slices/README.md` |

## Current audit state - 2026-09-01

- `docs/slices/` contains **no** `_OPEN_*.md` marker.
- The completed rows above have committed runtime/tests and/or explicit closing documentation in Git history.
- Slice 09 Deterministic New Game / galaxy generation baseline is closed with implementation/data/evidence commit `a5f3c13`; no implementation slice is currently open.
- Slices 07 and 08 are closed. Prepared specifications 09-12 remain queued; Slice 09 New Game / galaxy generation baseline is the recommended next objective and must begin with a fresh Gate 1 plus exactly one dated `_OPEN_` marker.
- Live current status is authoritative in `README.md`, `docs/PROJECT_STATUS.md`, `docs/architecture/README.md`, and `docs/research/ACTIVE_RESEARCH.md`.
- Historical research documents are evidence checkpoints and are not rewritten merely because a later slice resolved one of their then-deferred items.
- 2026-08-31 - Slice 03 Military Ship core / design baseline closed. Core schema 18 adds unbounded ID-based current ShipDesign catalogs, immutable concrete Ship snapshots, combat Fleet ShipIDs, normalized original hull/mandatory-system rules, authoritative military_ship construction and one-Ship Fleet completion. Gate 4 passed gofmt, go test ./... -count=1, go vet ./..., focused Military/ShipDesign/Construction/StrategicFleet/ColonyShip/Outpost/Blockade/Observer/Save/Load regressions, and working/staged diff checks. Gameplay commit: 2670d36. Next prepared objective: Slice 04 Combat Fleet movement / merge / split.

- 2026-08-31 - Slice 04 Combat Fleet movement / merge / split closed. Core schema 19 adds concrete combat-Fleet semantic transit, current-Empire Warp Drive + Trans-Dimensional speed, current Fuel Cell/supply range legality, deterministic whole/subset movement, stationary split/merge, composition-safe arrival and blockade timing, plus save/load, Observer isolation and deterministic replay coverage. Gate 4 passed gofmt, go test ./... -count=1, go vet ./..., focused CombatFleet/Strategic/Military/ColonyShip/Outpost/Blockade/Observer/Save/Load/Replay regressions and diff checks. Gameplay commit: f297480. Next prepared objective: Slice 05 Command Points / ship maintenance.

- 2026-08-31 - Slice 05 Command Points / ship Maintenance closed. Core schema 20 materializes the last-settlement Empire Command-Point `{Capacity, Used}` snapshot and Treasury ship-command Maintenance; economy ruleset schema 8 adds the original-derived base/station/Communications/Warlord/Imperium and 10 BC/excess-CP accounting constants while concrete military Ship usage remains hull `SizeIndex + 1`. Colony/Outpost special Fleets cost 1 CP while extant, Population transfers/Freighters remain excluded, split/merge/movement preserve usage, and legal special-Ship consumption removes usage. Star Base/Battlestation/Star Fortress replacement/buildability is normalized. Treasury now precomputes and validates every Empire before the first mutation, preserving all-Empire transaction safety while retaining pre-Construction timing. Gate 4 passed gofmt, full `go test ./... -count=1`, `go vet ./...`, focused original-accounting/CommandPoint/Treasury/Military/CombatFleet/ColonyShip/Outpost/Construction/Observer/Save/Load/Replay regressions, schema-20 compatibility review and diff checks. Gameplay/data/evidence commit: `16afe0b`. Next prepared objective: Slice 06 Strategic hostile encounters / BattleSession handoff.

- 2026-09-01 - Slice 06 Strategic hostile encounters / BattleSession handoff closed. Strategic resolution now pauses at the evidence-backed boundary after Fleet transit + Construction and before Blockades/Population transfers, aggregates same-Empire stationary combat Fleets into directed-hostile attacker/defender sides, creates deterministic parallel-per-system/sequential-per-system BattleSession waves, and atomically reconciles singular validated battle results. Concrete casualties remove Ships/empty Fleets; surviving losing combat and Colony/Outpost civilian Fleets retreat to the closest other owned Colony System or are destroyed when retreat cannot be established; civilian-only overrun remains limited to systems without defender combat Fleets or Colonies. Core schema remains 20 and economy ruleset schema remains 8. Gate 4 passed gofmt, focused CombatFleet/hostility/Blockade/Encounter/BattleSession ordering and replay regressions, full `go test ./... -count=1`, `go vet ./...`, and diff/whitespace checks. Gameplay/evidence commit: `096fd0a`. Next prepared objective: Slice 07 Tactical ship combat baseline.

- 2026-09-01 - Slice 07 Tactical ship combat baseline closed. Core schema 21 adds persistent structural weapon-mount identity to ShipDesign and built-Ship snapshots; tactical ruleset schema 1 adds only the evidenced initiative/original uint32 RNG/Beam/Laser/Frigate/Nuclear/Fusion/Electronic-Computer baseline while economy schema 8 and ship-hulls schema 3 remain unchanged. The exact supported Fusion-Laser-Frigate vs unarmed Nuclear-Frigate fixture runs through BattleSession-local round/initiative/command-sequence/Armor/Structure/readiness/destruction state, `battle.fire_beam` and `battle.end_activation`, golden RNG `100,19,100,100` ending at `0xFD95EBB9`, transactional rejection semantics and atomic Slice-06 strategic casualty reconciliation. Unsupported tactical families remain explicit lifecycle/manual-result or rejected paths. Gate 4 passed changed-file gofmt, focused fixed-seed Battle/Session/Strategic/Weapon/Replay regressions, wall-clock-independent Battle-ID ordering, exact strategic loss/survivor reconciliation, full `go test ./... -count=1`, `go vet ./...` and diff checks. Gameplay/data/evidence commit: `889f977`. No later slice is open; select and prepare the next objective before creating a new recovery marker.

- 2026-09-01 - Slice 08 Authoritative server / web HMI / transport baseline closed. `internal/session` now exposes lightweight transport-neutral Status and participant-only Battle projections; `internal/app` hosts GameSessions, automatic server-owned phase driving, detached player/observer snapshot envelopes, non-authoritative `change_sequence` and invalidation subscribers; `internal/server` provides the versioned `net/http` API, strict/bounded JSON and same-origin mutation checks, optional Observer projection, `github.com/coder/websocket v1.8.15` notification-only streaming and SPA serving; `cmd/moox-server` adds a loopback-first standalone host; and `web/` adds the React/TypeScript/Vite browser proof using native HTTP/WebSocket. Integration coverage proves real `colony.assign_population`, WebSocket invalidation + HTTP resync, accepted real `battle.fire_beam` transport, origin/error/Observer boundaries and byte-identical direct-vs-HTTP authoritative Observer state. Gate 4 repeated changed/new gofmt, focused transport/session regressions, full `go test ./... -count=1`, `go vet ./...`, `npm run build`, standalone server + headless Chrome runtime proof, dependency/bypass checks and `git diff --check`. Persisted gameplay schemas remain Core21 / Command1 / Event1 / Economy8 / Ship-Hulls3 / Tactical1. Implementation/data/evidence commit: `9315111`. Next prepared objective: Slice 09 deterministic New Game / galaxy generation baseline.

- 2026-09-02 - Slice 09 Deterministic New Game / galaxy generation baseline closed. New Game now owns one shared seeded RNG and stable generation/ID order, creates the frozen Small/Normal/Average/Tactical 20-star Human+Darlok start from normalized original-derived galaxy tables, validates legal Homeworld/Population/Research/Treasury/Fleet state, and feeds the generated result directly into the real two-seat GameSession. `internal/app` and `POST /api/v1/games` provide atomic server-authoritative creation with string uint64 seeds and stable 400/409 errors; normal `moox-server` startup is empty while the legacy fixture is explicit `-demo-fixture`; the React/Vite HMI creates games through the server and reuses snapshot/CommandBatch authority. Gate 4 repeated the golden `0x8009` fingerprint (`1d89bf9e8a5ce47c81a481ad669916d727357587dfc40e46904b9503ba314a89`), semantic plus byte-identical equal-seed state, different-seed divergence, generated Turn 1 -> Turn 2 session execution, full Go tests/vet/web build, frontend authority scan, standalone server + managed Chrome runtime proof and diff checks. Implementation/data/evidence commit: `a5f3c13`. Next prepared objective: Slice 10 diplomacy / war / peace baseline.
- 2026-09-03 - Slice 10 Diplomacy / war / peace baseline closed. Core schema 22 now persists canonical reciprocal `peace|war` relations plus directional pending peace offers; missing relation remains neutral. Server-authoritative `diplomacy.declare_war`, `diplomacy.offer_peace` and `diplomacy.accept_peace` commands are strict, revision-bound and only accepted during Planning before the first current-turn submission; deterministic events and player-safe projections drive the React HMI. `MayAttackEmpire` is the shared war-only attack authorization consumed by blockade and strategic encounter derivation. Timing regressions prove neutral same-system fleets do not fight, pre-boundary war enables the encounter, accepted peace preserves fleets while suppressing the next encounter, and diplomacy cannot cancel an already materialized tactical Battle. Gate 4 repeated focused diplomacy/blockade/encounter/tactical/New-Game coverage five times, exact diplomacy replay ten times, Core22 save/load and noncanonical-state rejection, server/browser authority scans, full `go test ./... -count=1`, `go vet ./...`, `npm --prefix web run build` and `git diff --check`. Implementation/data/evidence commit: `b91f65e`. Next prepared objective: Slice 11 Troop Transport / invasion / Colony conquest baseline.
- 2026-09-03 - Slice 11 Troop Transport / invasion / Colony conquest baseline closed. Core schema 23 now persists Colony standing Infantry and the fixed civilian `troop_transport`; Transports are baseline-buildable at 100 PP (Feudal 67 PP), use existing strategic movement/fuel range, count 1 CP at ordinary settlement and remain civilian Encounter assets. Server-authoritative `invasion_decisions` plus strict revision-bound `invasion.invade|decline` drive the original-evidenced narrow Infantry/Militia d100 Ground Combat, selected-Transport loss/survivor packing, Colony ownership transfer, conquered/assimilated cohort handoff, Construction clearing and Capital reassignment while war remains active. Post-conquest continuation rematerializes Blockades, Population transfers, Colony Economy/Population dynamics and Food without a second Treasury/CP settlement. Gate 4 independently repeated focused Transport/Invasion/Encounter/Population/Economy/authority regressions, added exact captured ownership/cohort/Capital Core23 save-load proof plus handled-key reset/no-double-settlement proof, and passed full `go test ./... -count=1`, `go vet ./...`, `npm --prefix web run build` and `git diff --check`. Implementation/evidence commit: `5b363ce`. Next prepared objective: Slice 12 Empire elimination and first deterministic headless victory loop.
- 2026-09-03 - Slice 12 Empire elimination / first deterministic headless victory loop closed. Direct original 1.31 evidence proves zero normal Colonies eliminates an Empire while Outposts/Fleets do not delay elimination. MOOX now exposes terminal `session.PhaseCompleted` plus immutable conquest result, deterministic eliminated-asset cleanup, read-only post-game behavior, completed HTTP/WebSocket/React projection and exact completed-session snapshot roundtrip. The canonical real `NewGame(0x8009)` fixture uses only public application/session commands after New Game to research Urridium range, colonize/extend supply with Outposts, declare war, move a real combat Fleet/Troop Transport, invade the final Darlok Colony and end with a byte-identical replayed Human winner; surviving off-system Darlok Fleets are removed by elimination. The first long-run attempt also exposed and fixed a Population Cohort projected-growth capacity clamp bug. Full Go tests, vet, Web build and diff checks passed. Implementation/evidence commit: `90b83d8`. Remaining Council/Orion/Antaran/score victory, AI and broad fidelity/depth work are post-milestone.
- 2026-09-03 - Post-Slice-12 fidelity/depth backlog audit completed. The audit re-baselines stale status documentation after the first complete deterministic headless lifecycle and separates P0 playability blockers from P1 breadth and P2/P3 fidelity/productization work. It identifies built-in AI, live GameSession save/resume and a playable browser strategic HMI as the next milestone blockers, followed by New Game/preset-race breadth and military ship-design component/weapon breadth. Prepared Slices 13-17 were created; none is open. Permanent audit: `docs/research/POST_MILESTONE_FIDELITY_DEPTH_BACKLOG_AUDIT_2026-09-03.md`.
- 2026-09-03 - Slice 13 Built-in strategic AI baseline closed. MOOX now has deterministic no-cheat `baseline_v1` built-in controllers driven only through player-safe DecisionView/DecisionCatalog and normal Session/App authority, including controller-neutral Research settlement, strategic expansion/construction/movement/diplomacy, supported Tactical Beam actions and invasion. Optional per-seat `builtin_ai` assignment is exposed through New Game/HTTP, while `Host.AdvanceAutomation` provides bounded autonomous stepping. The canonical real `NewGame(0x8009)` AI-vs-AI fixture completes by conquest without direct post-NewGame state mutation (diagnostic baseline Turn 430 / Revision 871), and committed regressions require byte-identical completed snapshots including result/event history plus deterministic Human-vs-AI continuation. Gate 4 independently repeated the exact AI and no-cheat/planner/Tactical regressions, then passed full `go test ./... -count=1`, `go vet ./...`, `npm run build` in `web`, and `git diff --check`. Implementation/evidence commit: `371aa0c`. Next prepared objective: Slice 14 Live GameSession save/resume baseline.

- 2026-09-04 - Slice 14 Live GameSession save/resume baseline closed. LiveSnapshot v1 now persists the deterministic in-progress Session/Battle continuation state over Core23, including controller/submission/event/telemetry/counter state, active Tactical RNG/revision/sequence state, Invasion/Colony-Base interactive boundaries, explicit simulation compatibility and compacted loaded-rules fingerprinting. Host/HTTP export/import/atomic restore is storage-neutral and privileged/default-off. Gate 4 independently repeated the canonical real `NewGame(0x8009)` AI-vs-AI save at Planning Turn 250 -> fresh Host import -> byte-identical re-export -> identical final existing CompletedSnapshot, all supported phase/Tactical/HTTP roundtrips with repeated runs, and malformed/rules/simulation/counter/controller/resolver/GameID rejection. Gate 4 added explicit unknown-schema and partial-save atomic rejection coverage in `e6a5e69`, then passed final `go test ./... -count=1` (`internal/session` 44.693 s), `go vet ./...`, `npm run build` and `git diff --check`. Runtime implementation: `a2c4ccd`. Next prepared objective: Slice 15 Playable browser strategic HMI loop Gate 1.

- 2026-09-04 - Prepared Slice 15 browser-playability work was decomposed from one oversized HMI slice into a parent epic plus five independently gated sub-slices: 15.1 UX/navigation/design system, 15.2 functional strategic gameplay HMI, 15.3 MOOX visual identity/graphics/asset pipeline, 15.4 rich gameplay/battle/persistence UX, and 15.5 complete browser vertical slice/polish/QA. Slice 16 New Game/preset-race breadth and Slice 17 military-design breadth remain separate. No implementation slice was opened; 15.1 Gate 1 is the next prepared objective.

- 2026-09-04 - Slice 15.1 opened. Gate 1 audited the current single-page React HMI and froze the next decision around a mobile-first browser shell: phone-primary navigation/status/context surfaces, adaptive desktop layout, explicit reconnect/invalidation/error states, reusable structural design-system primitives and a required post-implementation NetBird phone-preview server. Permanent evidence: `docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE1_2026-09-04.md`. Gate 2 user approval is pending.

- 2026-09-04 - Slice 15.1 scope clarified before Gate 2: MOOX browser HMI is multi-language from the structural baseline, initially German + English with runtime switching and persisted client preference. Locale is presentation-only and must not affect deterministic game/session/save state. The shared translation-key layer is inherited by all later Slice 15.x UI work. The user also confirmed successful phone access to the live `0.0.0.0:7171` preview through NetBird.

- 2026-09-04 - Slice 15.1 Gate 2 approved/frozen: mobile-first top status + five-item bottom navigation, desktop-adaptive left rail, lightweight recoverable route model, explicit lifecycle/error/pending-decision states, structural semantic design tokens/components, German/English runtime localization and the `0.0.0.0:7171` + managed-Chrome review contract. Gate 3 implementation authorized. Permanent evidence: `docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE2_2026-09-04.md`.

- 2026-09-04 - Slice 15.1 Gate 3 implemented the mobile-first/desktop-adaptive React shell, lightweight recoverable five-domain navigation, structural design-system primitives/tokens, explicit lifecycle/error/invalidation surfaces and persistent German/English runtime localization while retaining real New Game, population, Diplomacy and Invasion authority. Managed Chrome proved desktop rail behavior and a 390x844 no-overflow mobile bottom-navigation layout; permanent evidence: `docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE3_2026-09-04.md`. Gate 4 independent QA remains open.

- 2026-09-04 - Accepted downstream Slice 15.2/15.3 strategic HMI product direction: Galaxy becomes the interactive 2D primary map; known stars open an orbital star-system dialog with planets/gas giants/asteroid belts and present Fleets/Ships; legal Colony-Ship colonization is confirmed from the selected planet and owned planets/Colony-table rows converge on a full Colony Detail screen with top population jobs, planet/buildings and right-side construction queue. Colonies use a management table, Fleets use location-grouped tiles, Research uses eight classic categories, Diplomacy and Espionage are first-class product areas, and the OX lettermark remains an accepted brand element with a future spacecraft/star/orbit icon. Espionage currently has no backend command family, so 15.2 Gate 1 must freeze a minimal authoritative baseline or split a dedicated mechanics sub-slice. Permanent direction: `docs/research/SLICE_15_2_STRATEGIC_HMI_PRODUCT_DIRECTION_2026-09-04.md`.

- 2026-09-04 - Slice-15 roadmap amended after Tactical scope review: the family expands from five to six sub-slices. 15.4 is narrowed to rich non-tactical decisions/Encounter transitions/persistence UX; new **15.5 Interactive 2D Tactical Combat** owns the original-mechanics audit, authoritative movement/legal-action additions and mouse/touch battlefield on top of the narrow Slice-07 fixed-position Laser baseline; the former final browser-integration Slice 15.5 is renumbered to **15.6** and must integrate the accepted Tactical path into the complete browser-only Human-vs-AI journey. Permanent direction: `docs/research/SLICE_15_5_INTERACTIVE_TACTICAL_COMBAT_PRODUCT_DIRECTION_2026-09-04.md`.

- 2026-09-04 - **Slice 15.1 closed, Gates 1-4 complete.** Independent browser QA re-verified Main Menu/New Game/Galaxy/Colonies/Fleets/Research/More routing, direct deep-link recovery, DE/EN runtime switching and persistence, desktop adaptive navigation, 390/360/320px mobile behavior, visible focus, explicit loading and error states, live `0.0.0.0:7171` serving through NetBird `100.120.252.216`, and full repository QA. Gate 4 found and fixed one stale-snapshot race where an old game snapshot could render below a missing-game error after route changes; post-fix missing-game UI shows only the error shell. `npm run build`, `go test ./...`, `go vet ./...` and `git diff --check` passed. Permanent closure evidence: `docs/research/SLICE_15_1_MOBILE_FIRST_UX_GATE4_2026-09-04.md`. Next prepared objective: Slice 15.2 Gate 1; no next `_OPEN_` marker created yet.

- 2026-09-04 - **Slice 15.2 Gate 1 complete.** Functional HMI authority/projection audit found the existing deterministic player-safe `session.PlayerDecisionView` is explicitly designed for built-in AI and future strategic HMI clients and already projects Galaxy, own Ships/Fleets/Outposts/Colonies, public strategic contacts and legal Research/Construction/Population/FleetMove/Colonization/Outpost/Diplomacy choices. The current HTTP PlayerSnapshot exposes only the narrower PlayerView, so Gate 2 should freeze an atomic DecisionView-derived browser projection instead of screen-owned legality. Gate 1 also identified real authoritative gaps: non-colonizable orbital bodies/star spectral metadata are not persisted; Construction is one active project rather than a true ordered queue; Population transfer has mechanics but no legal target catalog; no discovery/fog state exists in the current no-fog baseline; eight Research categories need normalized category metadata; and actionable Espionage has modifiers but no state/commands/rules and therefore requires a dedicated later mechanics slice. Colonize remains a Planning CommandBatch action after confirm, with actual Colony creation after strategic resolution. Permanent evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE1_2026-09-04.md`. Next: Gate 2 contract freeze.

- 2026-09-04 - **Slice 15.2 Gate 2 complete.** User accepted the DecisionView/Planning-draft authority architecture and froze the strategic HMI contract with additional gameplay/UI requirements: persistent star/orbital-body state; original-like Outpost legality on normal planets, gas giants and asteroid belts; a real persistent construction queue with no MOOX user-visible item-count cap (deliberate divergence from original MOO2 seven slots); pure non-mutating Planning previews that recalculate Food/PP/RP, build ETA, empire Research ETA, Freighter usage and transfer ETA as job/queue drafts change; Colony-table same-row job reassignment vs cross-row Population transfer with target-job semantics and confirmation; sticky BC/net-income, Freighters and Command Points; normalized eight-category Research; War/Peace Diplomacy; and actionable Espionage deferred to newly prepared Slice 20. Permanent contract: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE2_2026-09-04.md`. Next: Gate 3 implementation.

- 2026-09-04 - **Slice 15.2 Gate-2 Colony-overview refinement accepted.** Before Gate-3 implementation the player added an explicit Colony Detail round-up requirement: show current Population/capacity/free capacity, projected Population growth per turn or starvation, next-Pop ETA when finite, planet traits, Food/PP/RP/economic summary, built Buildings/infrastructure and current construction/queue. Existing `Colony.Buildings`, `ColonyPopulationDynamics`, Colony economy and Planet traits remain authoritative; Planning preview supplies live derived growth/timing values and React must not fabricate unsupported Morale/Pollution. Final art remains Slice 15.3.

## 2026-09-07 - Slice 15.2 closed / Slice 15.3 opened

- Slice 15.2 Gate 4-C independent QA passed: full Go tests, go vet, web production build and git diff --check are green; live demo Galaxy workflow smoke-tested at desktop and mobile form factors.
- Removed the Slice 15.2 _OPEN marker and synchronized project/research status to closed.
- Opened Slice 15.3 Gate 1: MOOX visual identity / graphics / asset-pipeline audit and direction exploration.
- No push performed.

## 2026-09-07 - Slice 15.3 Gate 1 vector/procedural direction started

- Accepted a vector-first SVG direction for scalable MOOX UI/iconography while retaining modern raster delivery for rich textured artwork where appropriate.
- Captured a deterministic procedural 2D ship-generation direction with stable design seeds and a future 2D-to-3D path.
- Added the first React/SVG procedural ship proof of concept and integrated it into Fleet/Ship presentation paths without changing gameplay authority.
- Web production build and `git diff --check` pass.

## 2026-09-07 - Slice 15.3 Gate 1 real-game Shipbuilder prototype

- Restarted the public development server without demo fixture or persistence, removing all previous in-memory games, and created a fresh real `game-1` through the normal New Game browser flow with seed `0x8009`.
- Fresh game evidence: 20 systems, one authoritative `Scout`/Frigate design, two ships and two strategic fleets.
- Hardened the Galaxy HMI for real generated systems whose optional planet/body collections serialize as JSON `null`; the real Galaxy now loads cleanly.
- Verified the procedural SVG generator against real authoritative Fleet ships rather than the old demo fixture.
- Added an in-game Shipbuilder Gate-1 prototype: hull-size choice, repeated deterministic Generate, recent-history selection, Keep design and read-only authoritative equipment handoff.
- Frigate is labelled current ruleset; Destroyer/Cruiser/Battleship/Titan are visual previews only until Slice 17 implements authoritative breadth.
- Desktop 995x605 and mobile 360x646 browser QA passed; mobile has no horizontal overflow.

## 2026-09-07 - Slice 15.3 Gate 1 hull scale + concavity grammar

- Extended the visual Shipbuilder range from a small Scout role through Frigate, Destroyer, Cruiser, Battleship, Titan and Doom Star; Scout remains explicitly mapped to Frigate gameplay rules rather than becoming a seventh authoritative hull.
- Replaced the narrow scalar ship silhouette model with class-specific visual profiles for length, beam, station count, engines and detail density.
- Added deterministic external concave shoulder/notch/shoulder bays and transparent SVG-mask negative-space cutouts; complexity increases with hull scale.
- Browser-measured representative body bounds increase monotonically from Scout 57x18 to Doom Star 114x82 SVG units.
- Desktop 995x605 and mobile 360x646 Shipbuilder QA pass without horizontal overflow; Doom Star generations change the actual polygon path while retaining class constraints.
- Fresh in-memory `game-1` was preserved; no server restart and no push performed.

## 2026-09-07 - Slice 15.3 Gate 1 directed evolution + primitive genome

- Studied proceduro/spaceship-2d as an algorithmic/UI reference and documented its directed-evolution workflow, tetrahedron-based mirrored geometry, WebGL sprite outputs and Unlicense provenance.
- Implemented an original MOOX Visual Genome v2 in TypeScript/SVG rather than vendoring the reference implementation.
- Added deterministic Wedge/Spike/Pod primitive genes, parent-based mutation, primitive add/remove mutation and retained class-scale/concavity/cutout rules.
- Replaced the one-at-a-time Shipbuilder roll with six candidates, parent selection, six mutated descendants, Shape Mutation control, New random family and independent Keep design.
- Desktop 995x605 and mobile 360x646 browser QA pass with six candidates and no horizontal overflow.

## 2026-09-07 - Slice 15.3 Gate 1 Style-DNA candidates

- Added Spear/Angular, Sleek/High-Tech and Organic/Alien visual style genomes on top of Visual Genome v2.
- Style DNA biases hull length/beam, concavity depth, Wedge/Spike/Pod distributions, sweep and mirror probability while remaining presentation-only.
- Added a responsive live Style-DNA selector; switching style resets the evolution lineage to Generation 1 without touching the game or kept design.
- Desktop 995x605 and mobile 360x646 browser QA pass without horizontal overflow.

## 2026-09-07 - Slice 15.3 Gate 1 genome locks

- Added deterministic Core hull, Wings/Parts, Engines and Cutouts locks to Visual Genome v2 mutation.
- Locked groups copy exact parent gene data; primitive add/remove is disabled only while Wings/Parts is locked.
- Desktop QA proved all-four-lock descendants match the Doom Star parent geometry exactly, and partial-lock QA proved only the unlocked primitives mutate.
- Mobile 360x646 renders four locks as 2x2 while retaining the six-candidate 2-column population without horizontal overflow.

## 2026-09-07 - Slice 15.3 submitted-seat Planning-preview 500 fix

- Diagnosed persistent internal_error in live game-1 as an automatic Planning-preview request after Seat 1 had already submitted while the phase remained Planning.
- Client now suppresses Planning-preview for submitted local seats.
- Session/App layers type expected Planning-preview boundary rejection as session_rejected instead of allowing HTTP 500 fallback.
- Added generated-game and submitted-seat regression coverage; live reload of existing game-1 shows no error banner and no Planning-preview request.

## 2026-09-07 - Slice 15.3 full-random morphology v3

- Replaced the primary six-candidate directed-evolution flow with one large clickable ship and full-random rerolls.
- Added Visual Genome v3 morphology identity with Needle, Barge, Manta, Fork, Chevron/A-shape, Hammer, Bulb/Blob and Asymmetric topology families.
- Each roll independently randomizes morphology and Style DNA, then generates topology-specific envelope, signature primitives, details, concavity/cutouts and engines.
- Desktop browser QA reached all eight topology families within 24 rolls, with structural aspect ratios spanning roughly 0.70:1 to 9.15:1.
- Mobile 360x646 remains overflow-free across repeated extreme rerolls; Use this design preserves the exact chosen genome while later rolls continue independently.

## 2026-09-07 - Slice 15.3 symmetric half-hull v4

- Advanced procedural ship visuals to Visual Genome v4: lateral details are generated once and mirrored exactly across the longitudinal X axis.
- Removed Asymmetric/Oddform from the standard random morphology pool; seven symmetric topology families remain.
- Mirrored primitives, cutouts and hardpoints now have explicit QA metadata; engines remain axis-centered.
- Added hull Space display: Scout/Frigate 25, Destroyer 60, Cruiser 120, Battleship 250, Titan 500, Doom Star 1200.
- Added explicit visible class footprints 0.40/0.48/0.58/0.68/0.78/0.90/1.00 so Doom Star fills the shared preview while smaller hulls remain visibly smaller.
- Desktop measured Scout ~184x136, Cruiser ~313x231, Doom Star ~461x340 in the same 512x390 stage; mobile Scout ~120x78 versus Doom Star ~299x195 in the same 311x220 stage.

## 2026-09-07 - Slice 15.3 Visual Genome v4 server persistence

- Added authoritative optional Visual Genome v4 payloads to ShipDesign/Ship and a separate visual revision stream so picture changes never invalidate gameplay design revisions.
- Added immediate empire.set_military_design_visual; full resolved geometry is deep-copied and persisted rather than storing only a generator seed.
- Submitted gameplay batches remain valid after a visual-only immediate revision by advancing only their base-revision boundary.
- Completed ships freeze the design's Visual Genome/source visual revision; later design image changes do not mutate existing ships.
- Web Shipbuilder now saves Use this design server-side and restores the kept design from snapshots; strategic fleet glyphs prefer concrete persisted ship genomes with legacy seed fallback.
- Core state, GameSession live snapshot and Host export/import tests all round-trip the exact genome; gameplay revision independence is regression-covered.

## 2026-09-07 - Slice 15.3 core SVG icon language

- Added typed GameIcon React/SVG registry with a 24x24 vector-first technical grammar and 26 initial icon identities.
- Replaced primary Galaxy/Colonies/Fleets/Diplomacy/Espionage Unicode nav glyphs with shared SVGs.
- Replaced BC/Food/Freighter/Command/Research text/Unicode resource artwork and reused the same SVG in compact chips and detail popovers.
- Replaced menu/more/home/close shell glyphs with common vector icons.
- Added dedicated Farmer/Worker/Scientist role figures plus Colony info/close vectors.
- Desktop and 320px-mobile browser QA confirm responsive icon sizes and no actual horizontal document overflow; production web build passes.

## 2026-09-07 - Slice 15.3 Research/population/action icon refinement

- Changed global Research identity from an atom-like glyph to a microscope and added a separate test-tube primitive for future Chemistry/Biology/Pharma categories.
- Replaced CSS-only generic population people with role-specific Farmer/Worker/Scientist SVG markers while preserving drag/fractional behavior.
- Added typed action icons for generate/check/open/play/flag/outpost/build and applied them selectively to high-value Shipbuilder, planning, colony, system, construction and Research actions.
- Standardized icon-bearing buttons on inline-flex with 7px label gap, 15px normal icons and 13px compact icons.
- Desktop and 320px mobile browser QA pass without danger banners or actual horizontal overflow.

## 2026-09-07 - Canonical MOOX development port 7171

- Changed the standalone `moox-server` default bind from `127.0.0.1:8080` to `127.0.0.1:7171`.
- Updated the Vite API/health proxy and architecture/research references to use 7171 consistently.
- Runtime review server was migrated from temporary 7172 back to 7171 after exporting/importing `game-1`; the restored snapshot matched the source SHA-256 exactly.

## 2026-09-07 - Slice 15.3 Research categories and Galaxy markers

- Added distinct SVG identities for Construction, Chemistry, Computer, Physics, Energy, Sociology, Biology and Force Fields while keeping the global Research microscope.
- Replaced the Galaxy system CSS dot with a shared `star-system` SVG.
- Added compact own Colony/Outpost/Fleet markers and diplomacy-toned foreign Colony/Outpost/Fleet contact markers; duplicate empire+kind contacts are collapsed.
- Reused the same star/contact language in the System dialog.
- Desktop and 320px-mobile browser QA on canonical port 7171 passes without danger banners or horizontal overflow.

## 2026-09-07 - Slice 15.3 body, fleet, construction and diplomacy icon pass

- Replaced System body dots with typed planet/gas-giant/asteroid-belt SVGs and added colony/outpost badges plus climate tinting.
- Added fleet-role semantics for combat/civilian/scout and special colony/outpost/transport vessels while retaining procedural ship visuals.
- Replaced Construction catalog/project CSS glyphs with typed project icons; added dedicated Capitol, Colony Base, Marine Barracks and Star Base identities plus Housing/Terraform support.
- Reused building icons on Colony built-building tiles and replaced queue text arrows with shared arrow SVGs.
- Added neutral/peace/war Diplomacy stance icons and matching action-button icons.
- Desktop and 320px mobile QA pass without danger banners or horizontal overflow.

## 2026-09-07 - Slice 15.3 Gate-1 direction package completed

- Audited all functional strategic surfaces at the review width plus the 320px floor; no blocking visual inconsistency or document-level overflow remains.
- Added a live responsive `/styleboard.html` with three representative directions: Deep Space Command (recommended/current), Orbital Glass and Industrial Tactical.
- Defined target responsive tiers, rich-asset resolutions/crops, vector icon grammar, semantic asset IDs, fallback behavior and procedural-ship presentation boundaries.
- Defined provenance/license classes and required metadata; private/original-game reference material remains non-distributable without explicit rights.
- Prepared the Gate-2 visual identity / asset-pipeline contract draft and marked Gate 1 direction work complete; Gate 2 remains unfrozen pending explicit review.

## 2026-09-08 - Slice 15.3 art-fidelity planets/buildings review revision

- Replaced diagrammatic System body icons with spherical procedural world art using radial lighting, deterministic fractal-noise surfaces, climate palettes and cloud layers where appropriate.
- Added richer gas-giant band/storm rendering and deterministic asteroid-belt bodies.
- Added reusable `BuildingArt` illustrations for Capitol, Colony Base, Marine Barracks, Star Base, Battlestation and Star Fortress; command-station progression now has three genuinely different images.
- Construction catalog/hero and Colony built-building tiles reuse the same building art source; non-building actions remain concise semantic icons.
- Added a multi-entry runtime `/art-fidelity.html` review lab that renders the actual React components, including a future planet-surface composition concept.
- This is a Gate-2 review revision; Gate 2 remains unfrozen pending user acceptance of the richer fidelity direction.

## 2026-09-08 - Slice 15.3 spectral stars / building coverage / Construction layout revision

- Replaced the Galaxy semantic star icon with a luminous runtime `StarArt` using class-specific core/corona colors, radial fade, deterministic texture and a separate black-hole composition.
- Reused StarArt in the System dialog and changed raw normalized spectral integers to recognizable B/F/G/K/M/BD/BH labels.
- Audited the authoritative 48-entry building ruleset; Housing remains correctly a repeatable non-persistent project but now receives rich habitat art.
- Added deterministic family art for all building IDs not yet hand-authored, while retaining dedicated Capitol/Colony Base/Marine Barracks/Star Base/Battlestation/Star Fortress illustrations.
- Fixed rich Construction layout: full-width 16:7 detail art, aligned PP/facts/description/actions, and catalog columns sized to 68px/60px rich thumbnails.
- Expanded `/art-fidelity.html` to show all seven spectral identities and all 48 building art routes plus Housing; 320px review has no horizontal overflow.

## 2026-09-09 - Slice 15.3 playability visual baseline frozen

- Accepted the current Deep Space Command visual direction as the stable baseline for pursuing playability rather than continuing art polish.
- Froze Gate-2 baseline contracts for semantic palette/icon grammar, luminous spectral stars, spherical procedural planets, Visual Genome v4 ships, reusable building/station art, responsive floor and provenance policy.
- Explicitly parked Colony planet-surface composition, bespoke late-building art, richer portraits/environments and further effects until full-game playthrough evidence justifies another visual pass.
- Left Gate 3/4 visually parked rather than falsely marking final production art complete; Slice 15.4 may proceed against this baseline.

## 2026-09-09 - Slice 15.4 opened; Gate 1 rich-state UX audit complete

- Opened Slice 15.4 after freezing the 15.3 playability visual baseline.
- Audited all strategic session phases and confirmed human interaction boundaries: Encounters, Invasion Decisions and Post-Resolution Colony Base decisions; Research completion is automatic and Completed is terminal.
- Confirmed server DecisionView already projects Colony Base and Battle decisions, while current TypeScript omits Colony Base and leaves Battles untyped.
- Confirmed the browser has functional but numeric/minimal Invasion UI, no Colony Base UI, no Battle Entry route/command wrapper, and only a minimal result card.
- Audited persistence export/import/restore and WebSocket invalidation/reconnect; restore is atomic and publishes `game_loaded`, but the browser has no persistence UI/API wrappers and currently keeps a generic stale snapshot during reconnect.
- Defined structured lifecycle/error states, player-safe public identity/resolution-summary additions and the Gate-2 blocking-decision/transition-summary contract.
- Gate 2 remains unfrozen pending explicit review.

## 2026-09-09 - Slice 15.4 Gate 2 rich interaction contract frozen

- User approved the Gate-2 proposal without changes.
- Froze blocking decisions for Encounter/Invasion/Colony Base and transition summaries for Research/Battle return/Game Loaded/Victory.
- Froze stable Battle handoff route to Slice 15.5, local-file Save/Load/Restore semantics, read-only stale reconnect state, structured API errors/lifecycle and minimum player-safe identity/resolution reads.
- Gate 3 implementation is now active.

## 2026-09-09 - Slice 15.4 Gate 3 block 1

- Added typed client contracts for already-authoritative Colony Base and Battle projections.
- Added structured APIError preserving HTTP status/code/message.
- Added explicit authoritative lifecycle state and synchronized-only direct mutation gating.
- Browser-verified synced -> fatal/not_found -> synced recovery on canonical 7171 without changing game state.

## 2026-09-09 - Slice 15.4 Gate 3 block 2

- Added local Save Game export through the authoritative live-snapshot endpoint.
- Added Load Game to both the in-game menu and Main Menu so fresh servers can import saves.
- Added explicit atomic Restore confirmation for already-hosted game IDs and automatic Import for new IDs.
- Browser-E2E verified canonical game-1 Restore without state drift and isolated fresh-server Import on temporary 7191; desktop and 320px mobile remain overflow-free.
- Hardened WebSocket lifecycle so duplicate/old invalidations cannot leave the client stuck in refreshing.

## 2026-09-09 - Slice 15.4 Gate 3 block 3

- Closed the human Colony Base playability blocker with a server-projected Post-Resolution decision sheet.
- Added legal target planet cards using existing OrbitalBodyArt and exact server-projected 100 BC Scrap refund; no React-owned legality or refund math.
- Isolated E2E verified both Colonize -> new Colony / Turn 2 Planning and Scrap -> +100 BC / Turn 2 Planning.
- Verified 320px sheet with 44px minimum action targets and no horizontal overflow; canonical 7171 remained untouched/healthy during fixture QA.

## 2026-09-09 - Slice 15.4 Gate 3 block 4

- Replaced ID-heavy Invasion card with a blocking contextual decision showing target planet/system, defender public identity and exact eligible transport fleets.
- Added a minimal player-safe public empire identity projection (ID/name/race only); no foreign economy/research state exposed.
- Isolated E2E verified Decline -> Turn 2 Planning and Invade -> target capture plus conquest completion for the fixture.
- Verified 320px mobile, 44px actions and no horizontal overflow.


## 2026-09-09 - Slice 15.4 Gate 3 block 5

- Added a bounded player-safe `recent_resolutions` projection derived from the existing authoritative event log, with stable `event-<sequence>` IDs and no new save-schema state.
- Research completion/Technology grants are owner-only; Battle summaries are participant-only; Invasion summaries are limited to involved Empires; elimination/game completion remain public major results.
- Added automatic Research Breakthrough transition presentation with presentation-only per-game/per-seat acknowledgement and navigation into the existing Research UI.
- Replaced the small completed-game inline card with a dedicated read-only conquest result surface using public Empire names and a Main Menu action; ordinary strategic sections no longer render below a completed result.
- Regression evidence: `go test ./...`, `go vet ./...`, `npm run build`, focused resolution privacy/stable-ID tests and `git diff --check` pass. Evidence: `docs/research/SLICE_15_4_GATE3_BLOCK5_RESOLUTION_SUMMARY_RESEARCH_RESULT_2026-09-09.md`.
- Next Gate-3 block: Battle Entry/Return route and shell; interactive Tactical mechanics remain Slice 15.5.

## 2026-09-09 - Slice 15.4 Gate 3 - End Turn / Built-in AI blocker remediation

- Diagnosed live `game-1`: Seat 1 submission was accepted, but Seat 2 was incorrectly `local_human`, so Turn 1 Planning waited forever for a second Human submission.
- Fixed browser New Game to send the already-supported controller map explicitly: Seat 1 `local_human`, Seat 2 `builtin_ai`.
- Preserved the user's exact submitted `marine_barracks` construction + TechField 4 / Technology 13 Research batch, repaired Seat 2, and re-submitted the exact batch through the normal authoritative endpoint.
- Verified canonical live game at Turn 2 Planning with Darlok as Built-in AI, Research at 9 RP and Marine Barracks at 6 PP.
- Focused Human-vs-AI/server controller tests and web production build pass. Evidence: `docs/research/SLICE_15_4_GATE3_ENDTURN_BUILTIN_AI_BLOCKER_2026-09-09.md`.

## 2026-09-09 - Slice 15.4 Gate 3 - Research UX unification and technology information

- Removed the duplicate standalone Research selection surface; HUD, `/research` and Research Breakthrough now use the same `StrategicResearchOverlay`.
- Added a touch/keyboard-capable `?` action to every concrete Research Technology.
- Extended the authoritative Research choice projection with normalized Technology effect metadata for Buildings, planetary projects, ship mandatory components and modeled population bonuses.
- Detail UI reports only effects supported by current normalized rules and explicitly labels still-unmodeled Technology details rather than inventing descriptions.
- Browser QA verified both known-effect and unnormalized-effect dialogs and the full Breakthrough -> same chooser -> next Research flow.
- Full Go tests/vet and web production build pass. Evidence: `docs/research/SLICE_15_4_GATE3_RESEARCH_UX_TECH_INFO_2026-09-09.md`.
- Post-review layout hotfix compacted the `?` controls and Research field rows so all eight fields fit the visible two-column chooser again; browser geometry confirms no grid scrolling or clipped Technology rows at the managed desktop review viewport.
- Research field cards now size to `max-content`, preventing later 4-8 Technology fields from being clipped; TechField 4 was verified as exactly three authoritative choices and an 8-row browser probe grows the card instead of hiding entries.
- Research Technology help now includes source-backed original MOO2 HELP.LBX descriptions for all 203 Technologies, shown separately from normalized MOOX runtime effects so unimplemented original mechanics are informative without being misrepresented as implemented.
- Colony build-detail UX now explains selected buildings inline: dynamic cost/maintenance, a list of current normalized MOOX effects (supporting multiple effects), and the source-backed original MOO2 description; unimplemented original mechanics remain clearly separated from runtime truth.
- Slice 15.4 Gate 3 Block 6 added the stable `#/game/<gameID>/battle/<battleID>` encounter shell, strategic-continuation lock, Tactical hand-off boundary and one-time participant-safe Battle return summary; completed final Battles remain reviewable from `battle_completed` even after the live BattleView is cleared.
- Slice 15.4 Gate 3 Block 7 closed the stale-client/regression responsive gap: all critical mutations share 409 refetch-without-resubmit recovery, refresh-failed is explicitly retryable/read-only, Planning Preview errors are isolated, mobile primary actions honor 44px, and existing encounter/persistence/session regressions are mapped as Gate evidence.
- Slice 15.4 closed at Gate 4: rich non-tactical decisions, save/load/reconnect, stale-client recovery, Battle Entry/Return and participant-safe transition summaries are frozen; `#/game/<gameID>/battle/<battleID>` is the authoritative UI handoff point for Slice 15.5 Interactive Tactical Combat.
- Slice 15.5 opened at Gate 1. Original MOO2 executable/HELP audit closed the interactive movement model: 16 facings, Euclidean-ceiling translation plus per-facing turn cost, combat-speed movement budget, and distinct weapon-range metric; Gate-2 draft proposes 2v2, server-projected legal moves/targets, `battle.move_ship`, desktop/touch battlefield UX and unchanged Go authority.
- Slice 15.5 Gate-2 draft was amended before freeze to include original-style Tactical Scan/ship inspection and a MOOX-modernized effectively unbounded battlefield plane; the original finite combat-area evidence remains documented rather than being silently reinterpreted.
- Slice 15.5 Gate 2 frozen after user acceptance: Tactical movement/facing, 2v2 breadth, Scan, open-field camera model, server-owned legality and Laser/activation baseline are now implementation constraints for Gate 3.
- Slice 15.5 Gate 3 Block 1 implemented server-authoritative Tactical movement/facing and 2v2 breadth: finite projected legal moves on an effectively unbounded coordinate plane, original distance+turn cost, atomic move rejection, round movement reset and moved-position Laser range.
- Slice 15.5 Gate 3 Block 2 hardened Tactical player projection: RNG is hidden from participants, only the active controlling seat receives legal command catalogs, Scan/status remains visible, and the browser API now has typed move/fire/end-activation command plumbing.
- Slice 15.5 Gate 3 Block 3 replaced the hand-off placeholder with the first browser-playable Tactical battlefield: open SVG plane, server-projected movement/targets, Laser fire, Scan details, End Activation, Built-in-AI round progression and responsive pointer/touch camera controls.
- Slice 15.5 Gate 3 Block 4a removed the retained Slice-07 Armor-overflow blocker: excess Laser damage now reaches aggregate Structure and can destroy ships while internal subsystem damage remains deferred.
- Slice 15.5 Gate 3 Block 4b integrated the 15.3 procedural ship visuals and closed the full Tactical-to-strategic browser loop with exact destroyed/surviving ship identities.
- Slice 15.5 Gate 3 Block 5 closed Tactical rejection UX: stale/illegal move and fire remain non-destructive, refetch authority on 409, never auto-resubmit, and are visible directly inside the battlefield.
- Slice 15.5 closed at Gate 4: the first browser-playable server-authoritative Tactical baseline supports 1v1-2v2 movement/facing, Laser combat through ship destruction, Tactical Scan, responsive effectively-unbounded battlefield controls, rejection/refetch safety and exact Battle Return reconciliation; handoff is now Slice 15.6 full browser vertical-slice integration.
- Slice 15.6 opened at Gate 1: cross-slice audit found the minimal Military Design browser bridge (Frigate with optional one Laser) as the only concrete New-Game-to-Tactical gap; full Built-in-AI browser conquest, Save/Resume, responsive/runtime thresholds and no-dev-tool evidence rules are captured in the Gate-2 draft.
- Slice 15.6 Gate 2 frozen: complete-browser acceptance plus durable Ship Designer shell contract accepted; Gate3 implementation may begin in small recovered blocks.


## 2026-09-14 - Slice 15.6 final browser vertical slice closed

- Closed Slice 15.6 with Gates 1-4 complete.
- Canonical visible-browser seed `0x8009` acceptance passed New Game, meaningful research, real Save/Load/Resume, expansion/supply, ordinary interactive Tactical + Battle Return, Troop Transport/Invasion and authoritative Human Conquest Victory over Darlok.
- Final authoritative acceptance result: conquest, Human Empire 2 / Seat 1, Darlok Empire 3 eliminated, completed Round 562 / Revision 1137.
- Full `go test ./... -count=1` and `npm run build` passed; responsive evidence includes desktop/<=980 closeout plus existing real 390x844 Ship Designer/Tactical QA.
- Explicitly deferred non-required scanner intelligence, special-ship visual genomes, deeper Tactical batching/fire-all/split-fire, richer art and broad Military Ship Designer breadth (Slice 17) rather than carrying them as hidden Slice-15.6 work.
- Closure evidence: `docs/research/SLICE_15_6_GATE3_FINAL_ACCEPTANCE_RUN_2026-09-14.md` and `docs/research/SLICE_15_6_GATE4_CLOSE_2026-09-14.md`.

## 2026-09-15 - Slice 16.1 New Game visual selection grammar closed

- Closed Slice 16.1 with Gates 1-4 complete after explicit user acceptance of the compact Galaxy Size selector/art direction.
- Frozen and implemented the reusable typed `VisualSelector<TId>`, responsive setting-card grid, rounded image-led artwork, info-on-demand modal/bottom-sheet behavior, keyboard/mouse/touch interaction baseline, semantic New Game asset conventions and server-authoritative supported/planned presentation.
- Accepted five original deterministic Galaxy Size SVG reference assets (Tiny/Small/Medium/Large/Huge) plus reproducible generator/manifest integration; arm count is treated as illustrative morphology, not a direct size metric.
- Added build-time UTF-8 and deterministic selector/asset contract checks plus reusable real-Chrome browser smoke at 390 px mobile and desktop.
- Final Gate-4 acceptance passed full web build, no-horizontal-overflow checks, mouse interaction, real CDP touch tap, ArrowLeft/ArrowRight/Home/End keyboard navigation, info-dialog focus/Escape behavior and Small-only Create Game support semantics.
- Detailed explanatory copy remains refinable downstream without reopening the frozen interaction/asset contract.
- Slice 16.2-16.6 remain separate prepared work; no next `_OPEN_` marker was created during closure.
- Closure evidence: `docs/research/SLICE_16_1_GATE4_CLOSE_2026-09-15.md`.
- Closing commit subject: `docs: close slice 16.1 visual selection grammar`.
