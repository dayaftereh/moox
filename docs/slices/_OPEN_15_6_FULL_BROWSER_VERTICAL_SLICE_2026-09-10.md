# Planned slice 15.6 - Full browser vertical slice / polish / complete-game QA
Status: **open; Gate 2 frozen; Gate 3 implementation ready**.
Queue position: **15.6 of Slice-15 family**.
## Objective
Integrate and polish Slices 15.1-15.5 into the first complete browser-playable Master of Orion X vertical slice: Main Menu -> New Game -> strategic play -> save/reload/resume -> supported decisions and interactive Tactical battle -> victory.
## Dependencies
- Slices 15.1-15.5 closed.
- Slice 13 built-in AI.
- Slice 14 live persistence.
## Scope guard
In scope:
- coherent Main Menu / New Game / Resume / active-game / result journey;
- complete supported Human-vs-built-in-AI playthrough using browser surfaces only;
- final navigation consistency and visual polish across all implemented screens;
- integration of 2D Galaxy/system/Colony/Fleet/Research/Diplomacy surfaces with rich decisions/persistence and Slice-15.5 interactive Tactical Combat;
- loading/performance/runtime-error cleanup that blocks normal play;
- save/reload/reconnect smoke and recovery paths;
- representative responsive support for agreed viewports;
- E2E/browser smoke coverage for the canonical match journey;
- final authority audit proving React contains presentation/orchestration only.
Defer:
- Slice 16 New Game/preset-race breadth;
- Slice 17 broad military Ship Designer breadth;
- Tactical breadth beyond the explicit Slice-15.5 accepted baseline;
- advanced Diplomacy/Espionage/leader/alternative-victory systems beyond authoritative baselines explicitly accepted by preceding slices;
- native packaging and large-scale content production beyond the canonical vertical slice.
## Gate 1 - Integration readiness audit
- [x] Audit all 15.1-15.5 acceptance evidence and unresolved cross-slice UX seams.
- [x] Define the exact canonical browser playthrough and checkpoints from launch to victory, including at least one supported interactive Tactical battle path where the canonical scenario provides one.
- [x] Identify blockers requiring integration fixes rather than new gameplay scope.
- [x] Define performance/responsive/error thresholds for a playable vertical slice.
- [x] Define automated/manual browser E2E strategy including save/reload/reconnect and strategic <-> Tactical transitions.
- [x] Present Gate-2 final vertical-slice acceptance contract.
## Gate 2 - Vertical-slice freeze
- [x] Freeze canonical start-to-victory browser scenario and allowed setup.
- [x] Freeze final navigation/polish/performance acceptance thresholds.
- [x] Freeze save/reload/reconnect and failure-recovery acceptance cases.
- [x] Freeze browser/E2E evidence requirements and no-dev-tool rule.
- [x] Freeze expected strategic -> Tactical -> strategic integration checkpoints.
## Gate 3 - Integration and polish
- [ ] Close cross-screen/navigation/data-refresh seams.
- [ ] Complete Main Menu/New Game/Resume/result integration.
- [ ] Integrate and polish the accepted interactive Tactical path inside the complete strategic journey.
- [ ] Fix blocking performance/responsive/runtime-error issues.
- [ ] Add E2E/smoke automation where practical and deterministic.
- [ ] Perform complete Human-vs-built-in-AI browser playthroughs without direct API/dev-tool commands.
## Gate 4 - Final milestone QA + close
- [ ] Complete at least one full supported Human-vs-built-in-AI browser game from Main Menu to victory.
- [ ] Exercise at least one supported interactive Tactical battle entirely through the browser when present in the frozen canonical scenario.
- [ ] Save mid-game, reload/reconnect and finish with authoritative deterministic continuation.
- [ ] Verify representative responsive/browser-runtime behavior and production asset loading.
- [ ] Run full Go tests/vet/web build/E2E plus `git diff --check`.
- [ ] Perform final React authority scan and document milestone evidence.
- [ ] Update parent Slice-15 milestone, status/HISTORY and close marker.
Milestone on closure: **first saveable, resumable and finishable Human-vs-built-in-AI MOOX browser vertical slice with an integrated interactive Tactical Combat path**.

## Gate-1 evidence (2026-09-10)

Gate 1 is complete. Evidence: docs/research/SLICE_15_6_GATE1_INTEGRATION_READINESS_AUDIT_2026-09-10.md. Gate-2 draft: docs/research/SLICE_15_6_GATE2_VERTICAL_SLICE_CONTRACT_DRAFT_2026-09-10.md. Gate 3 remains blocked until explicit user acceptance/freeze.

## Gate-2 freeze (2026-09-10)

Explicitly accepted by user after Ship Designer UX refinements. Frozen evidence: docs/research/SLICE_15_6_GATE2_FREEZE_2026-09-10.md. Gate 3 may now implement in small checkpointed/recovered blocks.

## Gate-3 Block 5 plan - Galaxy fleet direct manipulation (2026-09-10)

User feedback identified fleet movement from the Galaxy map as the next integration blocker. The current home-system markers are colony presence plus aggregate own-fleet ship count, but the fleet marker is not independently interactive. The existing Fleets view can stage `empire.move_fleet` only when server-projected legal targets exist.

Planned Block 5 is now documented in `docs/research/SLICE_15_6_GATE3_BLOCK5_GALAXY_FLEET_MOVEMENT_PLAN_2026-09-10.md`:

Block 5A COMPLETE (`0f6c262`, `0588a9f`, `850350e`): persistent visited-system and known-empire authority, anonymous unvisited stars, hidden planets/special identity, pre-contact diplomacy/seat/contact filtering, server-side first-contact persistence, player-safe built-in AI exploration, full Go regression green, web build green, and visible canonical browser QA proving 1 named home + 19 anonymous stars + no pre-contact Diplomacy contact.

Block 5B COMPLETE (`09a5b03`): authoritative legal + visible-illegal fleet-target catalog with direct distance, supply distance, range, ETA and stable rejection reason; existing legal `fleet_moves` now derive from this same authority. Focused 2 pc legal / 5 pc out-of-range fixture plus command-validation parity and anonymous-target session coverage are green; full Go regression and web build are green. Next: 5C Galaxy fleet marker + floating picker.
Block 5C COMPLETE (`3338ead`): independent Galaxy fleet marker + compact draggable fleet/ship picker, exact persisted ship SVGs, whole/single-ship selection profiles backed only by `fleet_move_targets`, fixed Colony Ship support handling, and visible browser QA for marker/system-dialog separation, selection, dragging and canonical 0/19 reachable feedback. Full Go regression and web build are green. Next: 5D drag/drop + tap target mode, authoritative route/range feedback and staged movement.
5C UX refinement (`3503a6b`): replaced fleet tabs with one compact four-column system-ship pool, vertical scroll, stable Colony/Outpost/Transport vector silhouettes, green own-colony system names, and small per-fleet clockwise orbit markers. Canonical QA: Scout 1 + Scout 2 + Colony Ship together, 292 px picker, 2/3 cross-fleet selection, 0/19 authoritative targets. Next remains 5D.
- persistent per-empire visited-system knowledge and player-safe Galaxy projection: unvisited stars expose position + stellar appearance only (no true name, planets or special identity such as Orion); arrival reveals and names the system; unknown race/diplomacy identity remains hidden until authoritative first contact;
- authoritative legal + visible-illegal fleet target projection with range/ETA/reason feedback;
- independently clickable Galaxy fleet marker;
- compact draggable fleet picker with selectable exact ship SVGs;
- drag/drop-to-star plus tap/click `Ziel wÃƒÂ¤hlen` fallback;
- deterministic QA-only three-star micro fixture (`2 pc` reachable target, `5 pc` out-of-range target) beside, not instead of, canonical seed `0x8009`;
- visible-browser movement -> encounter -> Tactical regression, including Laser fire and strategic return.

The normal Small-Galaxy generator and canonical `0x8009` seed are explicitly not to be altered merely to make this focused QA case reachable.
Special-ship visual persistence remains OPEN: Colony/Outpost/Troop strategic units need authoritative persisted per-instance visual genomes with deterministic creation/backfill and save/load/reconnect stability; the current `SpecialShipGlyph` SVGs are interim presentation only.
