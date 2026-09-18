# Planned slice 23 - Full Advanced Starting Civilization parity

Status: **planned / queued; not open**.

Queue position: **reserved future Slice 23 by product decision**. Slice 23 is the downstream home for the complete Advanced-start work discovered during Slice 16.7. It should open only after the strategic mechanics required for an honest Advanced civilization exist.

## Objective

Implement the original-style **Advanced starting civilization** as a complete, deterministic, server-authoritative New Game mode after the game has matured enough to support its consequences honestly.

Slice 23 inherits the permanent evidence produced by Slice 16.7 rather than restarting the research from zero.

The target is broader than “Average plus 19 technology grants”. It includes:

- Advanced technology selection;
- deterministic research-profile/RNG behavior;
- larger starting territory;
- variable starting population;
- starting Buildings;
- the accepted starting fleet/Command-Point contract;
- complete interaction with all then-supported preset/custom races;
- Galaxy Size / Age / Difficulty / player-composition interaction;
- persistence/replay;
- full New Game HMI/API enablement only after backend parity exists.

## Why Advanced is deferred to Slice 23

Slice 16.7 Gate-1/Gate-2 research proved that the current architecture can structurally carry:

- arbitrary multi-Colony starting Empires;
- variable Population;
- deterministic Advanced technology state;
- Command Points;
- multi-AI games;
- full save/resume state.

However, full Advanced parity also pulls in gameplay breadth that is intentionally not complete during Slice 16:

- original Advanced starting Buildings can include many normalized buildings whose effects are still partial today;
- Advanced territory balancing can require richer planet-effect/twiddle semantics;
- HELP and the owned 1.31 executable disagree on Advanced starting-fleet behavior;
- full race breadth changes Advanced chooser/population/homeworld interactions;
- original secondary Advanced chooser filtering at 0x21CB0 is not yet mapped cleanly to the modern five Difficulty IDs.

Forcing those systems into Slice 16 would turn a New Game/UI program into a large Economy/Planet/Building implementation milestone. Slice 23 makes that dependency explicit instead.

## Required upstream milestones

Before Slice 23 moves beyond Gate 1, re-audit at least:

- **Slice 17** - ship/component/weapon breadth relevant to any chosen Advanced fleet contract;
- **Slice 18** - corrected galaxy scale/density, homeworld reachability and authoritative black-hole strategic navigation required by Advanced territory/navigation fixtures;
- **Slice 20** - Espionage/intelligence and spying-dependent race behavior used by downstream race completion;
- **Slice 21** - Colony / Buildings / Planet Effects / Pollution fidelity;
- **Slice 22** - Preset Race Completion + Custom Race Designer, including final authoritative race-modifier breadth.

The intended product order is:

**17 -> 18 -> 20 -> 21 -> 22 -> 23**

Slice 19 remains intentionally unassigned unless a separate product decision fills it.

Slice 23 must still perform a fresh Gate-1 dependency check; completion of an earlier slice is not permission to assume every Advanced dependency is solved.

## Permanent inherited Slice-16.7 evidence

Slice 23 begins from and preserves provenance to:

- docs/research/SLICE_16_7_GATE1_ADVANCED_PARITY_AUDIT_2026-09-18.md;
- the Slice-16.7 Gate-2 decision/deferral evidence once closed;
- TECHNOLOGY_START_RESEARCH_2026-08-27.md;
- ADVANCED_START_RESEARCH_2026-08-28.md;
- ADVANCED_START_CHOOSER_2026-08-28.md;
- NEW_GAME_GALAXY_GENERATION_BASELINE_2026-09-02.md;
- COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md;
- Slice-16.5 and Slice-16.6 New Game evidence.

Do not repeat reverse engineering already proven unless later code/data changes invalidate the assumption.

## Known Advanced contract already proven

### Technology breadth

Advanced starts from Average and performs exactly **19 additional grants**, ending with exactly **26 completed starting fields** per Empire.

Concrete applications are deterministic but seed/race/profile/order dependent.

The existing Advanced chooser already models:

- shared New Game RNG;
- stable Empire ordering;
- race weighting;
- Creative / Uncreative;
- government/application eligibility;
- Tactical/Strategic availability;
- cross-Empire technology competition.

### Territory count

Original executable evidence proves total starting Colonies per Empire:

**ceil(floor(NUM_STARS / 2) / NUM_PLAYERS)**

For the currently normalized galaxy sizes:

| Galaxy | Stars | 2 players | 3 players |
| --- | ---: | ---: | ---: |
| Small | 20 | 5 | 4 |
| Medium | 36 | 9 | 6 |
| Large | 54 | 14 | 9 |
| Huge | 71 | 18 | 12 |

Slice 23 must re-evaluate the matrix if player-count breadth has widened by then.

### Population

Advanced does not use the fixed Population-8 non-Advanced baseline.

Original evidence derives Advanced population from colony population limit and race population-growth delta, with a minimum-1 clamp and a minimum-8 clamp for the actual homeworld.

Slice 23 should use the full then-current race/planet/population mechanics from Slices 21/22 rather than a special Advanced approximation.

### Starting Buildings

Original evidence confirms a dedicated starting-building bootstrap with:

- Advanced candidate cap **9**;
- exact original 42-product candidate ordering already recorded in Slice-16.7 evidence;
- buildability/race/population gating;
- separate Capitol behavior.

Slice 23 must reuse Slice-21 building effects. It must not grant buildings whose promised gameplay behavior is inert.

### Treasury / Freighters

Current evidence supports the common start baseline:

- Treasury 50 BC;
- 0 Freighters.

Re-audit at Gate 1 in case later work proves an Advanced-specific override.

### Fleet discrepancy

This is a mandatory Gate-1/Gate-2 decision.

Original HELP says the Advanced starting fleet consumes all available Command Points.

The owned Orion2.exe 1.31 path calls Init_Advanced_Civilization_Fleet_, whose verified entry point is a single RET; the normal 2 Scouts + Colony Ship block is gated to Average, not Advanced.

Slice 23 must choose explicitly between:

1. executable 1.31 parity; or
2. documented product-intent parity.

By Slice 23, Slice-17 ship-design breadth and Slice-21 Command/building effects should make an evidence-backed product-intent fleet practical if that direction is chosen.

## Territory / planet balancing

Slice-16.7 Gate-2 feasibility research established:

- current Core/Persistence can hold the original Advanced colony counts;
- generated galaxies contain enough planets overall;
- a naive “pick current worlds without balancing” approach can create severe immediate food deficits;
- original Advanced has a dedicated planet selection/twiddle pipeline;
- its basic size/climate/mineral twiddle fields map onto current MOOX planet semantics;
- richer planet/special behavior belongs in Slice 21 rather than being faked inside Advanced.

Slice 23 must reuse the completed Slice-21 planet/economy model for territory selection and normalization.

## Race breadth

Slice 23 is intentionally after Slice 22 so Advanced can be validated against the full race system rather than only the early Human/Klackon local-player subset.

Gate 1 must audit:

- all preset races marked supported after Slice 22;
- representative custom races if Slice 22 permits custom New Game;
- Creative / Uncreative;
- government;
- population growth/capacity;
- environmental/homeworld traits;
- Tolerant/Pollution behavior;
- Artifacts/Rich/Large Home World;
- race research weights;
- any trait that changes Command Points or starting infrastructure.

Advanced must not introduce a second race-modifier path.

## Difficulty / original secondary filter

Slice-16.7 research identified original global byte 0x21CB0 as an additional Advanced chooser filter with no proven mapping to the current five MOOX Difficulty IDs.

By Slice 23 Gate 1:

- re-audit whether later AI/Difficulty work has resolved that mapping;
- if proven, integrate it;
- if still unproven, Gate 2 must explicitly freeze an evidence-safe supported behavior instead of guessing.

## Binding architecture direction

- Advanced remains a server-owned technology_level = advanced New Game mode.
- No frontend-only Advanced state.
- No Advanced-only duplicate Economy/Building/Planet/Race implementation.
- Slice-21 mechanics own Colony/Building/Planet/Pollution effects.
- Slice-22 mechanics own race/custom-race effects.
- Slice-17/normal ship-design rules own ship/fleet legality.
- Equal complete settings + equal seed + equal race definitions produce deterministic authoritative state.
- Advanced preferences may be transient bootstrap inputs only if resulting materialized state/RNG is sufficient for later gameplay; otherwise persist them explicitly.
- Catalog/UI may switch Advanced from planned to supported only after authoritative runtime and acceptance matrix are green.

## Gate 1 - re-audit inherited evidence against mature runtime

- [ ] Re-read Slice-16.7 permanent Advanced audit/deferral evidence.
- [ ] Re-audit Advanced chooser against the then-current Research/AI/Difficulty runtime.
- [ ] Re-audit exact territory-count/selection/twiddle behavior against Slice-21 planet mechanics.
- [ ] Re-audit Advanced population against final race/planet/population pipeline.
- [ ] Re-audit original starting-building candidate behavior against all Slice-21-complete building effects.
- [ ] Re-audit HELP-versus-executable fleet discrepancy against Slice-17 ship breadth and final Command-Point rules.
- [ ] Re-audit all Slice-22 supported preset/custom race interactions.
- [ ] Re-audit original 0x21CB0 mapping evidence.
- [ ] Re-audit persistence/replay shape and RNG ordering.
- [ ] Produce an exact Gate-2 parity matrix with no hidden fallback semantics.

## Gate 2 - final Advanced contract freeze

- [ ] Freeze exact supported advanced API/catalog identity.
- [ ] Freeze Advanced preference/profile generation and lifecycle.
- [ ] Freeze exact technology-grant ordering/RNG/competition contract.
- [ ] Freeze Advanced colony-count and deterministic territory-selection/twiddle contract.
- [ ] Freeze population and starting job-allocation contract.
- [ ] Freeze starting-building ownership/effects.
- [ ] Freeze Treasury/Freighters.
- [ ] Freeze explicit fleet parity target and exact resulting design/ship/fleet state.
- [ ] Freeze Command-Point ordering and expected Capacity/Used fixtures.
- [ ] Freeze all supported preset/custom race interactions.
- [ ] Freeze Difficulty behavior including resolved or explicitly bounded 0x21CB0 semantics.
- [ ] Freeze player-composition breadth supported by the then-current New Game runtime.
- [ ] Freeze persistence/replay and exact RNG-state fixtures.
- [ ] Freeze server-derived Advanced visual-card facts and launch-briefing content.
- [ ] Freeze complete cross-setting deterministic acceptance matrix.

## Gate 3 - implementation

- [ ] Wire server-authoritative Advanced settings through catalog/API validation.
- [ ] Generate deterministic preference/profile inputs.
- [ ] Run Average + 19 Advanced technology grants with exact shared RNG/order semantics.
- [ ] Build deterministic Advanced territory using the frozen planet-selection/twiddle contract.
- [ ] Initialize variable population/jobs through normal Population/Economy rules.
- [ ] Grant starting Buildings through Slice-21 mechanics.
- [ ] Create the frozen Advanced fleet through normal ship/fleet rules.
- [ ] Recalculate Command Points only after final technology/colonies/buildings/fleet prerequisites are materialized.
- [ ] Support all Gate-2-frozen preset/custom race combinations.
- [ ] Persist/reload/replay Advanced state exactly.
- [ ] Promote Advanced catalog/UI to supported only when implementation is complete.
- [ ] Add exact fixtures and crafted-request rejection tests.

## Gate 4 - QA + close

- [ ] Advanced create/play works through public HTTP for every frozen player/race/difficulty/galaxy class.
- [ ] Equal seed/settings/races produce deterministic authoritative state and exact post-bootstrap RNG.
- [ ] Exact technology fields/applications match frozen fixtures.
- [ ] Exact Colony count/identity/population/jobs/buildings match fixtures.
- [ ] Exact fleet/design/Command Points match the chosen parity contract.
- [ ] Built-in AI advances turns from Advanced games without stalling.
- [ ] Pollution/food/economy remain viable immediately after bootstrap.
- [ ] Save/export/import/restore/replay remains byte/deterministically stable.
- [ ] Advanced UI card and launch briefing report only server-owned facts.
- [ ] Desktop and 390 px New Game flow passes.
- [ ] Full Go tests, vet, Web build/browser smoke and git diff --check pass.
- [ ] Update New Game docs/roadmap/HISTORY and close Slice 23.

## Exit criterion

Advanced is no longer a planned placeholder. It is a complete server-authoritative starting-civilization mode built on mature normal gameplay systems for Research, Colonies, Buildings, Planet effects, Pollution, Race traits, Ships and Command Points. It works deterministically across the frozen New Game matrix, survives save/resume/replay exactly, and neither HMI nor backend relies on Advanced-only approximations for mechanics implemented elsewhere.
