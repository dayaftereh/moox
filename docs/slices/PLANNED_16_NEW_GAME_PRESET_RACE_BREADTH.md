# Planned slice 16 - New Game and preset-race breadth

Status: **program in progress; Slices 16.1-16.3 closed, Slice 16.4 next prepared / not open**.

Position: **Slices 16.1, 16.2 and 16.3 are closed; Slice 16.4 is the next prepared New Game breadth slice but remains unopened until explicitly started.** Slice 17 remains the alternative prepared Military Ship Designer breadth milestone.

## Objective

Widen the deliberately narrow Slice-09 New Game contract using existing original/normalized evidence, with particular focus on preset races and settings that can now be exercised by built-in AI and the browser HMI.

Slice 16 is also the product/UI pass that changes New Game from a mostly form-like settings screen into a **visual, image-led setup experience**. The user should browse major options using large original MOOX artwork plus previous/next arrows, in the same interaction family as the Ship Designer rather than selecting everything from dense dropdowns.

## Binding Slice-16 product direction

The following direction is now part of the prepared Slice-16 contract and must not be lost during Gate-1 implementation planning:

- New Game uses a reusable **carousel / previous-next visual selector** on desktop and mobile.
- Difficulty is image-led and should visually communicate the progression Easy -> Normal -> Hard -> Very Hard -> Impossible; Gate 1 still audits the authoritative gameplay mapping.
- Galaxy size/age use original visual galaxy illustrations so scale/character are understandable at a glance.
- Every preset race accepted into the runtime contract gets its own **original MOOX race portrait**, browsable with left/right arrows.
- Race portraits may be inspired by classic space-4X archetypes, but original MOO2 portrait/UI artwork must not be copied, traced, repainted or reused.
- Complex race portraits default to optimized **PNG/WebP raster art**; SVG is preferred for frames, emblems, arrows, number tiles, abstract icons and simple symbolic art. Full-vector race portraits are optional, not mandatory.
- Starting technology/technology level is presented with **three strong visual cards/illustrations** if Gate 1 confirms the expected three-level authoritative contract.
- Opponent count uses a visual number selector, targeting original MOOX-styled numerals **1-9**, with the enabled range constrained by the authoritative galaxy/player-count contract.
- Opponent composition reuses the same canonical race portraits instead of separate New Game-only art.
- A final launch summary shows the selected galaxy, difficulty, Human race, technology start, opponent count/composition and seed visually before game creation.
- Images are never the sole source of information: all settings remain accessible by text/ARIA and server-authoritative validation.

## Existing narrow boundary

Current runtime explicitly accepts only Small galaxy, Normal age, Average technology, Tactical combat, exactly two players, exactly one Human and one Darlok. The normalized ruleset already contains 13 preset races.

## Dependencies

- Slice 13 built-in AI.
- Slice 15.6 completed browser vertical slice for the established user-facing shell; this slice widens its New Game/race breadth.
- Existing Slice-15 visual-system work supplies the MOOX visual language; Slice 16 extends it with race portraits and New Game setting art.

## Slice-16 decomposition

Slice 16 is intentionally decomposed into six recovery-safe sub-slices. These files are the binding prepared work breakdown:

1. **16.1 - New Game visual selection grammar**  
   `PLANNED_16_1_NEW_GAME_VISUAL_SELECTION_GRAMMAR.md`  
   Shared carousel/arrows, responsive layout, accessibility and asset-manifest conventions.

2. **16.2 - Difficulty visual selector**  
   `PLANNED_16_2_DIFFICULTY_VISUAL_SELECTOR.md`  
   Difficulty contract plus original visual progression, including the preferred Easy/Normal/Hard/Very Hard/Impossible presentation vocabulary.

3. **16.3 - Galaxy visual selector**  
   `PLANNED_16_3_GALAXY_VISUAL_SELECTOR.md`  
   Galaxy size and age support plus image-led scale/stellar-character selection.

4. **16.4 - Preset race portraits and carousel**  
   `PLANNED_16_4_PRESET_RACE_PORTRAITS_CAROUSEL.md`  
   Compatibility audit for the 13 normalized preset races, original portrait art bible/manifest and reusable race carousel identity.

5. **16.5 - Starting technology visual selector**  
   `PLANNED_16_5_STARTING_TECH_VISUAL_SELECTOR.md`  
   Evidence-backed technology-start breadth with three visual cards when confirmed by Gate 1.

6. **16.6 - Player composition and final New Game integration**  
   `PLANNED_16_6_PLAYER_COMPOSITION_FINAL_NEW_GAME_INTEGRATION.md`  
   Visual opponent count, opponent races/controllers, cross-setting validation, final launch summary and full Slice-16 acceptance.

The normal four-gate protocol still applies. The six sub-slices are not permission to implement unsupported settings optimistically: each sub-slice performs its own evidence/current-runtime check before freezing gameplay semantics.

## Scope guard

In scope candidates to freeze during Gate 1 / the relevant 16.x audit:

- additional preset races with correct starting modifiers/technologies/government;
- one original reusable race portrait per accepted preset race;
- more than two seats where evidence/runtime invariants are ready;
- additional galaxy sizes/ages/technology levels;
- difficulty/controller assignment if it can be separated cleanly from AI quality;
- deterministic seed equality per supported settings tuple;
- image-led HMI exposure for accepted options;
- original MOOX setting art/portraits/number tiles required by the accepted visual selectors.

Defer:

- full custom Race Designer;
- unsupported trait interactions not required by preset races;
- Council/Antaran/Orion victory;
- broad AI personality fidelity;
- animated/cinematic race leaders as a requirement for Slice 16;
- copying/reusing original MOO2 artwork.

## Parent Gate 1 - Original/settings audit

- [ ] Re-check original New Game option tables and generator evidence.
- [ ] Inventory which of 13 normalized races are runtime-compatible without new subsystem work.
- [ ] Map size/age/tech/player-count options to generator rules/evidence.
- [ ] Identify preset races blocked by missing trait/government/population mechanics.
- [ ] Choose a bounded expansion set rather than enabling unsupported options optimistically.
- [ ] Define seed/hash fixtures for each newly accepted setup.
- [ ] Confirm the 16.1-16.6 work breakdown still matches current runtime dependencies.
- [ ] Present Gate-2 breadth contract.

## Parent Gate 2 - Implementation decision

- [ ] Freeze newly supported settings/races/player counts.
- [ ] Freeze exact unsupported/deferred list.
- [ ] Freeze deterministic fixture matrix and HMI exposure.
- [ ] Freeze the common visual selector/art-manifest contract and representative art direction.

## Parent Gate 3 - Implementation

- [ ] Complete 16.1-16.6 in dependency order or an explicitly justified equivalent order.
- [ ] Expand validated New Game settings/race support.
- [ ] Add required preset-race runtime modifiers only where evidence-backed.
- [ ] Produce original runtime artwork for every visual option that is accepted into the final Slice-16 contract.
- [ ] Extend server/HMI creation surfaces.
- [ ] Add deterministic multi-seed/settings regressions.

## Parent Gate 4 - Follow-up QA + commit + close

- [ ] Repeat each supported settings fixture exactly.
- [ ] Verify AI and HMI can start/play newly accepted setups.
- [ ] Verify all accepted races/options have non-placeholder imagery and usable text/accessibility fallbacks.
- [ ] Desktop and real/mobile-width New Game walkthrough.
- [ ] Run full tests/vet/web checks and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close Slice 16.

## Parent exit criterion

Slice 16 is complete when New Game supports the frozen broader settings/race/player matrix deterministically and presents those choices through one polished, image-led MOOX setup flow with original race portraits and visual setting cards, without enabling mechanics the authoritative runtime cannot yet support.