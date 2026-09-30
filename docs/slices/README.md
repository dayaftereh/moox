# Work slice protocol

This directory defines the recovery-safe workflow for focused Master of Orion X implementation slices.

Closed slices are audited in docs/slices/HISTORY.md; the current/next objective remains authoritative in docs/research/ACTIVE_RESEARCH.md.

## Open-slice marker

An unfinished slice is represented by a file named:

```text
docs/slices/_OPEN_<AREA>_<SLUG>_<YYYY-MM-DD>.md
```

The leading `_OPEN_` is intentional: it is the repository-visible signal that work was started but the slice is not closed yet.

By default there should be exactly one `_OPEN_*.md` marker. Do not start a later slice while an open marker exists unless parallel work was explicitly intended.

The marker is a compact recovery/checklist file. Long-lived evidence, reverse-engineering findings, architectural decisions and final results belong in a permanent domain document, for example `docs/research/<SLICE>_<DATE>.md`.

## Planned-slice specifications

Future work may be prepared without opening it. Prepared queue entries use:

```text
docs/slices/PLANNED_<NN>_<SLUG>.md
```

`PLANNED_*.md` files are **not** recovery markers. Multiple planned specifications may coexist and none of them counts as active work. They may contain the complete intended Gate 1-4 checklist, dependencies, scope guard and exit criterion so a later chat can start from an explicit contract rather than reconstructing the roadmap.

When a planned slice actually starts:

1. verify there is no existing `_OPEN_*.md` marker;
2. read the planned specification and current `docs/research/ACTIVE_RESEARCH.md`;
3. create exactly one dated `_OPEN_<AREA>_<SLUG>_<YYYY-MM-DD>.md` recovery marker referencing the planned specification;
4. create/update the permanent `docs/research/<SLICE>_<DATE>.md` evidence document during Gate 1;
5. follow the normal four-gate protocol below.

Do not mark Gate 1 complete merely because a `PLANNED_` checklist exists. Gate 1 always requires a fresh repository checkup and original-evidence review at the time work begins.

### Current active / prepared queue - 2026-09-28

Active:

- **No Slice-17 child is currently open. Slice 17.5 - Tactical Combat Foundation / Ship Combatant Breadth is closed/accepted with Gates 1-4 complete.** Slices 17.0-17.5 and follow-up 17.2A are closed/accepted; Slice 17.6 is the next prepared objective and remains planned/not open; Slices 17.7-17.9 remain planned/not open.

Closed/parked Slice-15 state:

- Slice 15.1 and 15.2: closed.
- Slice 15.3: playability visual baseline frozen/parked; richer-art follow-up remains backlog rather than active slice work.
- Slice 15.4: closed rich gameplay/persistence UX baseline.
- Slice 15.5: closed accepted interactive Tactical baseline.
- Slice 15.6: closed final browser vertical slice; canonical New Game -> Save/Resume -> Tactical -> Invasion -> Conquest acceptance passed.

Prepared queue:

1. PLANNED_17_MILITARY_SHIP_DESIGN_COMPONENT_WEAPON_BREADTH.md - active parent program; child order 17.0 -> 17.1 -> 17.2 -> 17.3 -> 17.2A -> 17.4 -> 17.5 -> 17.6 -> 17.7 -> 17.8 -> 17.9; 17.0-17.5 and follow-up 17.2A are closed/accepted at Gates 1-4; 17.6 is the next prepared objective and remains planned/not open; 17.7-17.9 remain planned/not open.
2. PLANNED_18_GALAXY_GENERATION_STRATEGIC_NAVIGATION_FIDELITY.md - reserved Galaxy Generation / Strategic Navigation fidelity pass; planned, not open.
3. PLANNED_20_ESPIONAGE_INTELLIGENCE_BASELINE.md - dedicated authoritative Spy/intelligence mechanics; Slice 19 remains intentionally unassigned.
4. PLANNED_21_COLONY_BUILDINGS_PLANET_EFFECTS_POLLUTION_FIDELITY.md - authoritative Colony/Building/Planet effects including Pollution.
5. PLANNED_22_PRESET_RACE_COMPLETION_CUSTOM_RACE_DESIGNER.md - preset-race completion plus Custom Race Designer.
6. PLANNED_23_FULL_ADVANCED_STARTING_CIVILIZATION_PARITY.md - full Advanced parity inheriting Slice-16.7 evidence.

Planned downstream order: **17 -> 18 -> 20 -> 21 -> 22 -> 23**. Slice 19 remains intentionally unassigned. No downstream slice is opened merely by being reserved.

### Binding Slice-17 decomposition

Slice 17 is a parent program. These child specifications are recovery-safe milestones and must be completed in dependency order unless a checkpoint explicitly justifies otherwise:

1. **Slice 17.0 - Ship Designer / Triangle Reference Test Harness**
   `PLANNED_17_0_SHIP_DESIGNER_TRIANGLE_REFERENCE_TEST_HARNESS.md`
   Dev-only Designer/Construction/Combat labs, technology-rich reference states and bounded normal-resolver turn advancement.
2. **Slice 17.1 - Hulls and Mandatory Ship Systems**
   `PLANNED_17_1_HULLS_MANDATORY_SHIP_SYSTEMS.md`
   Frigate through Doom Star plus drives, computers, armor, shields, fuel, hull buildability and miniaturization foundation.
3. **Slice 17.2 - Beam Weapons and Mount Modifiers**
   `PLANNED_17_2_BEAM_WEAPONS_MOUNT_MODIFIERS.md`
   **Closed/accepted 2026-09-23.** Beam breadth, mount slots/counts, applicable modifiers, cost/space/unlocks and explicit Tactical handoff to 17.6.
4. **Slice 17.3 - Missiles, Torpedoes and Bomb Ordnance**
   `PLANNED_17_3_MISSILES_TORPEDOES_BOMBS_ORDNANCE.md`
   **Closed/accepted; Gates 1-4 complete (2026-09-24).** Authoritative missile/torpedo/bomb design breadth is complete to the frozen contract; complete projectile/interception/impact combat remains explicitly owned by 17.7.
5. **Slice 17.2A - Beam Weapon Firing Arcs**
   `PLANNED_17_2A_BEAM_WEAPON_FIRING_ARCS.md`
   **Closed/accepted; Gates 1-4 complete (2026-09-26).** Original Beam arc evidence, frozen contract, server-authoritative implementation and accepted radial Selector HMI are durable; Tactical arc enforcement remains owned by 17.6.
6. **Slice 17.4 - Ship Specials and Fighter Systems**
   `PLANNED_17_4_SHIP_SPECIALS_FIGHTER_SYSTEMS.md`
   **Closed/accepted; Gates 1-4 complete (2026-09-28).** Authoritative Specials/Fighter design breadth, immutable snapshots and final weapon-consistent Designer HMI are accepted; deferred combat-special/fighter Tactical consumers remain owned by 17.8.
7. **Slice 17.5 - Tactical Combat Foundation / Ship Combatant Breadth**
   `PLANNED_17_5_TACTICAL_COMBAT_FOUNDATION_SHIP_COMBATANT_BREADTH.md`
   Replaces the old baseline-only Tactical fixture with verified broad hull/mandatory-system admission plus generic server-owned action/RNG/damage/event/UI plumbing.
8. **Slice 17.6 - Tactical Beam Weapons & Beam Modifiers**
   `PLANNED_17_6_TACTICAL_BEAM_WEAPONS_MODIFIERS.md`
   Full supported Beam hit/range/damage, Heavy/Point Defense/modifiers, Selector effective-stat parity and Beam presentation.
9. **Slice 17.7 - Tactical Missiles / Torpedoes / Bombs**
   `PLANNED_17_7_TACTICAL_ORDNANCE_MISSILES_TORPEDOES_BOMBS.md`
   Ordnance launch/ammo/projectiles, interception, impact/damage, Point Defense integration and ordnance presentation.
10. **Slice 17.8 - Tactical Fighters / Bombers / Combat Specials**
   `PLANNED_17_8_TACTICAL_FIGHTERS_BOMBERS_COMBAT_SPECIALS.md`
   Fighter/Bomber/Interceptor combat plus supported Tactical special consumers and event-driven presentation.
11. **Slice 17.9 - Full Ship Designer + Tactical Integration and Acceptance**
   `PLANNED_17_9_FULL_SHIP_DESIGNER_TACTICAL_INTEGRATION_ACCEPTANCE.md`
   Unified Designer -> Construction -> Fleet -> Encounter -> Tactical -> strategic reconciliation, persistence/replay and final browser/program QA.

The parent program file remains `PLANNED_17_MILITARY_SHIP_DESIGN_COMPONENT_WEAPON_BREADTH.md`. Slice 18 remains downstream and independent of Slice-17 completion.

### Binding Slice-16 decomposition

The following six specifications are the prepared Slice-16 work breakdown. They are intentionally separate recovery-safe slices and must be read together with the parent `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`.

1. **Slice 16.1 - New Game visual selection grammar**  
   `PLANNED_16_1_NEW_GAME_VISUAL_SELECTION_GRAMMAR.md`  
   Shared image-led carousel/previous-next interaction, desktop/mobile layout, accessibility and asset-manifest conventions.

2. **Slice 16.2 - Difficulty visual selector**  
   `PLANNED_16_2_DIFFICULTY_VISUAL_SELECTOR.md`  
   Preferred Easy/Normal/Hard/Very Hard/Impossible presentation with an original visual strength/progression motif and evidence-backed gameplay mapping.

3. **Slice 16.3 - Galaxy visual selector**  
   `PLANNED_16_3_GALAXY_VISUAL_SELECTOR.md`  
   Galaxy size/age breadth shown through original galaxy illustrations and the shared left/right selection grammar.

4. **Slice 16.4 - Preset race portraits and carousel**  
   `PLANNED_16_4_PRESET_RACE_PORTRAITS_CAROUSEL.md`  
   Runtime compatibility audit for the 13 normalized preset races plus one original reusable MOOX portrait identity per accepted race. Complex portraits default to PNG/WebP; SVG is preferred for frames, emblems and symbolic art. Original MOO2 portrait/UI art must not be copied/traced/reused.

5. **Slice 16.5 - Starting technology visual selector**  
   `PLANNED_16_5_STARTING_TECH_VISUAL_SELECTOR.md`  
   Three visual starting-technology/technology-level cards when the expected authoritative three-level contract is confirmed by Gate 1.

6. **Slice 16.6 - Player composition and New Game integration**
   `PLANNED_16_6_PLAYER_COMPOSITION_FINAL_NEW_GAME_INTEGRATION.md`  
   Visual opponent-count selector targeting MOOX-styled numerals 1-7, reusable opponent race portraits, cross-setting validation and launch-summary integration.

7. **Slice 16.7 - closed as audited/deferred Advanced handoff** - Gate 1 audited complete Advanced parity; Gate 2 froze the planned/locked contract and handed implementation to Slice 23. Evidence: docs/research/SLICE_16_7_GATE2_ADVANCED_DEFERRAL_HANDOFF_2026-09-18.md.
   `PLANNED_16_7_ADVANCED_STARTING_TECH_PARITY.md`
   Full authoritative Advanced start parity across technology, empire/population/fleet/research bootstrap and accepted cross-setting interactions. This is the final 16.x slice.

This decomposition is **binding product direction**, not permission to skip the four-gate protocol. Exact original/runtime settings, modifiers, race compatibility and player-count limits still require Gate-1 evidence before gameplay support is enabled.

Milestone after Slice 15.6: **achieved 2026-09-14** - a human can start, save/resume, play the accepted interactive Tactical battle path and finish the supported Human-vs-built-in-AI match from the browser through one coherent MOOX visual/interaction language. The 2026-09-23 Slice-17 product decision supersedes the earlier plan to postpone additional Tactical weapon/special breadth until after Slice 20: Tactical foundation/Beam/ordnance/fighter-special breadth is now explicitly decomposed into Slices 17.5-17.8, with 17.9 final integration. Deeper Diplomacy/Leaders/post-baseline Espionage and alternative-victory depth remain later roadmap work.

The queue order is the current recommendation, not a hard dependency lock. Reprioritization is allowed before a later slice is opened if the dependency check still passes.

## Four gates

Every slice uses the same four gates.

### 1. Checkup + original analysis / reverse engineering

- inspect Git status, current runtime/data/docs and the previous checkpoint before changing behavior;
- inspect original MOO2 executable/data/save evidence where the slice concerns original behavior;
- distinguish directly proven behavior from inference and deliberate MOOX modernization;
- record findings, addresses/offsets/provenance, unknowns and blockers in the permanent slice document;
- update the `_OPEN_` marker;
- **stop before implementation and present the findings plus the proposed implementation shape for discussion.**

This gate combines the repository checkup and the original/reverse-engineering investigation.

### 2. Implementation decision

- discuss the implementation shape with the user;
- record the accepted state/data/API/rule design and any deliberate divergence from original MOO2;
- mark this gate complete only after that decision is clear.

No gameplay implementation should begin before this gate is complete.

### 3. Implementation

- implement only the agreed slice;
- add/update deterministic tests and normalized data where required;
- keep authoritative gameplay logic in core/game/session layers rather than UI/network adapters;
- keep the permanent slice document synchronized with material changes.

### 4. Follow-up QA + commit + close

Run the relevant final gates, normally including:

```text
gofmt on changed Go files
go test ./...
go vet ./...
git diff --check
```

Also run slice-specific generators/analyzers/regressions when applicable.

Then:

- record final behavior, QA and the commit subject/result in the permanent slice document;
- update `docs/research/ACTIVE_RESEARCH.md` to the next objective;
- delete the completed slice's `_OPEN_*.md` marker in the closing commit;
- create the next `_OPEN_*.md` marker only when the next slice actually starts.

A completed slice therefore leaves a permanent evidence/result document but no `_OPEN_` marker.

## Recovery rule

At the start of work:

1. inspect `git status` without discarding unknown changes;
2. check `docs/slices/_OPEN_*.md`;
3. if an open marker exists, resume that slice first and read its permanent document;
4. only if no open marker exists, use `docs/research/ACTIVE_RESEARCH.md` to choose the next slice.

This convention is specifically intended to make interrupted chats or agent handoffs obvious and recoverable.
