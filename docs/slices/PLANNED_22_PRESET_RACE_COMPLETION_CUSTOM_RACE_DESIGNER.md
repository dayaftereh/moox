# Planned slice 22 - Preset race completion and Custom Race Designer

Status: **planned / queued; not open**.

Queue position: **reserved future Slice 22 by product decision**. Slice 22 is deliberately downstream of Slice 20 Espionage and Slice 21 Colony/Buildings/Planet/Pollution fidelity, and it should close before the full Advanced-start work in Slice 23 opens. Reserving the number does not itself satisfy those dependencies.

## Objective

Complete the race system after the game has gained more of the mechanics that the original preset traits depend on, then build one authoritative Custom Race Designer on top of the **same** race-trait rules.

Slice 22 has two linked product goals:

1. **Preset race completion** - widen the player/opponent supported set from the bounded Slice-16.4 baseline until all 13 normalized preset races can be selected honestly with their defining supported mechanics.
2. **Custom Race Designer** - allow a player to construct a legal custom race from the authoritative normalized race-trait catalog, pick budget, group/exclusivity rules and implemented gameplay effects rather than from frontend-only checkboxes.

The Custom Race Designer must reuse the same authoritative trait semantics used by preset races. There must not be a second "custom-race-only" implementation of farming, industry, government, spying, ship attack, population growth, special abilities, homeworld modifiers or other race effects.

## Why Slice 22 is intentionally later

Slice 16.4 proves the portrait/catalog/UI pipeline and introduces only a narrow honest player-supported set. Gate-1/2 of Slice 16.4 found that many normalized preset traits still depend on gameplay families that are incomplete or absent today.

Examples include:

- Espionage / Spying and Stealthy Ships;
- Omniscient / Telepathic information and capture breadth;
- Lucky / Random Events;
- Charismatic / Repulsive diplomacy consequences;
- ship attack / ship defense Tactical race bonuses;
- ground-combat race bonuses;
- Cybernetic combat repair;
- Trans-Dimensional Tactical movement breadth;
- Warlord troop/crew/leader breadth;
- Artifacts / Rich / Large Home World start rules;
- Fantastic Traders treaty/economy breadth.

It is therefore a product requirement that Slice 22 performs a **fresh dependency audit at Gate 1**. Reserving the number does not imply that these prerequisites are already complete.

## Known normalized foundation

Current normalized rules already provide a strong data base for Slice 22:

- `data/rulesets/moo2-1.31/races.json` - 13 canonical preset race definitions and trait selections;
- `data/rulesets/moo2-1.31/race_traits.json` - authoritative normalized race-trait groups/options, pick costs, scopes/values/abilities and provenance;
- normalized Race Designer pick budget currently records:
  - `starting_picks = 10`;
  - `max_negative_picks = 10`;
- existing rules loading already resolves preset selections into semantic race modifiers;
- Slice 16.4 establishes a server-owned race catalog, reusable portraits/manifest and carousel identity contract;
- Slice 16.6 established general player/opponent composition and final New Game integration.

Gate 1 must re-check original evidence and current runtime before treating any normalized field as a fully implemented gameplay effect.

## Dependencies / opening guard

Before Slice 22 may move beyond Gate 1, verify the state of at least:

- Slice 16.4 preset race catalog/portrait baseline;
- Slice 16.6 player/opponent composition and final New Game integration;
- Slice 20 Espionage/intelligence for spying-dependent races such as Darlok;
- Slice 21 Colony/Buildings/Planet Effects/Pollution fidelity for Tolerant, homeworld/environment, economy/building and other race-dependent strategic effects;
- Tactical race-effect breadth needed by ship attack/defense and Trans-Dimensional/Cybernetic effects;
- Invasion/ground-combat race modifiers;
- Diplomacy breadth needed by Charismatic/Repulsive;
- Random Events breadth needed by Lucky;
- homeworld-generation breadth needed by Artifacts/Rich/Large Home World;
- any Leader/crew systems required by the accepted Warlord/Charismatic contract.

If a defining trait still depends on an absent subsystem, Gate 2 must either keep that trait/preset unavailable or explicitly queue the missing prerequisite. Do not silently make the trait cosmetic.

Slice 23 must not consume a partially complete Slice-22 race contract as if it were full race parity; Slice 22 closure must state exactly which preset/custom race mechanics are authoritative.

## Binding architecture direction

- Preset and custom races use one authoritative `RaceDefinition` / race-modifier pipeline.
- Trait legality, pick costs, group selection, mutual exclusivity and total budget are server-owned.
- The browser may preview costs/effects, but the server re-validates every custom race submission.
- A custom race is persisted as normalized race-definition data, not merely as a UI label or opaque frontend JSON blob.
- Equal seed + equal complete custom race definition must produce deterministic New Game state.
- Save/resume/replay must preserve the exact custom race identity and trait configuration.
- Built-in AI must receive the same mechanical race effects as a Human-controlled Empire with the same race definition.
- No Custom Race option may promise an effect that the authoritative gameplay layer does not implement.
- Preset races remain immutable canonical templates; a custom race may start from a preset in the UI, but editing creates a separate custom definition rather than mutating the preset.

## Custom Race Designer product direction

The first complete designer should support, subject to Gate-2 evidence/freeze:

- custom race/civilization name;
- optional preset template as a starting point;
- normalized trait groups/options from `race_traits.json`;
- positive and negative pick costs;
- authoritative remaining-picks display;
- maximum-negative-picks enforcement;
- single-select / mutually exclusive trait-group behavior from normalized rules;
- government choice through the same trait model;
- environmental/gravity/homeworld traits where their mechanics are ready;
- special abilities only when their gameplay behavior is actually implemented;
- a compact live summary of resulting authoritative modifiers;
- portrait/emblem identity selection from a curated MOOX asset library.

Arbitrary AI-generated/user-uploaded custom portraits are **not required** for the first Slice-22 exit criterion. They may be a later art-workflow extension. The first designer should be mechanically complete before adding unlimited portrait generation/upload complexity.

## Preset completion target

The long-term target is that all canonical preset IDs can become `supported` in the server race catalog when their defining effects are implemented and tested:

```text
alkari
bulrathi
darlok
elerian
gnolam
human
klackon
meklar
mrrshan
psilon
sakkra
silicoid
trilarian
```

Gate 2 must freeze which of these are ready at Slice-22 implementation time. Do not mark all 13 supported merely because Slice 22 exists.

## Gate 1 - dependency / original Race Designer audit

- [ ] Re-audit all 13 preset races against the then-current gameplay runtime and classify every selected trait as complete, partial or blocked.
- [ ] Re-check original Race Designer behavior from local owned evidence, including starting pick budget, negative-pick limit, group selection and pick costs.
- [ ] Audit every normalized `race_traits.json` option against a real authoritative gameplay consumer.
- [ ] Identify traits whose primary effect is still missing and map each to a prerequisite slice/system.
- [ ] Re-check preset derived-pick totals against the normalized trait catalog and original evidence.
- [ ] Define the canonical custom-race state shape needed for New Game, GameState, save/resume and replay.
- [ ] Define how custom race definitions map into the existing RaceModifiers / RaceResearchModifiers pipeline without duplicating rules.
- [ ] Define built-in-AI handling for arbitrary legal custom-race definitions.
- [ ] Define player-safe race identity projection for diplomacy/intelligence/opponent surfaces.
- [ ] Audit portrait/emblem reuse needs for custom races and define a bounded first visual-identity library.
- [ ] Present the exact Gate-2 preset-support + custom-trait matrix.

## Gate 2 - authority freeze

- [ ] Freeze the preset-race set that becomes fully player/opponent supported in Slice 22.
- [ ] Freeze the complete supported custom-trait set and explicit deferred/disabled traits.
- [ ] Freeze authoritative pick budget / negative-pick rules / group exclusivity and cost validation.
- [ ] Freeze canonical custom race definition/state schema and stable identity/ID rules.
- [ ] Freeze how custom definitions resolve into gameplay modifiers and research preferences.
- [ ] Freeze New Game API/catalog extensions for presets vs custom races.
- [ ] Freeze persistence/replay/versioning/migration behavior for custom race definitions.
- [ ] Freeze opponent/built-in-AI support policy for custom races.
- [ ] Freeze Race Designer HMI, mobile layout, accessibility and portrait/emblem selection contract.
- [ ] Freeze deterministic fixtures for preset and representative custom-race combinations.

## Gate 3 - implementation

- [ ] Implement/complete the accepted missing race-trait mechanics required by the frozen supported set, using the normal authoritative subsystem owners.
- [ ] Promote newly complete preset races to server-catalog `supported` status with coverage.
- [ ] Add canonical custom-race definition/state and strict server validation.
- [ ] Route custom trait selections through the same authoritative race-modifier pipeline as presets.
- [ ] Add New Game request/response support for a custom race definition or stable custom-race reference.
- [ ] Persist custom race definitions through save/resume/replay and migrations.
- [ ] Support built-in AI with representative legal custom-race definitions where Gate 2 requires it.
- [ ] Build the image-led Custom Race Designer HMI with pick accounting and server-derived trait facts.
- [ ] Reuse canonical portrait/emblem assets and provide a bounded custom visual-identity chooser.
- [ ] Add deterministic fixtures for every newly supported preset and representative custom-race archetypes.
- [ ] Add explicit crafted-request rejection tests for illegal combinations, overspent budgets, too many negative picks and unsupported traits.

## Gate 4 - QA + close

- [ ] Every preset marked supported has all Gate-2-required defining mechanics active and tested.
- [ ] Preset definitions remain canonical/immutable while custom races persist as separate definitions.
- [ ] Server rejects every illegal custom trait/budget/group combination even when the browser is bypassed.
- [ ] Equal seed + equal settings + equal custom race definition produces byte/deterministically equal authoritative state.
- [ ] Save/reload/replay preserves custom identity, traits and effects exactly.
- [ ] Human and built-in-AI Empires receive identical race effects for the same definition where controller-independent mechanics apply.
- [ ] Browser designer passes desktop and 390 px mobile QA with mouse/touch/keyboard and text/ARIA fallback.
- [ ] Portrait/emblem identity remains usable outside New Game in opponent/diplomacy/intelligence surfaces.
- [ ] Full Go tests, vet, web build/browser smoke and `git diff --check` pass.
- [ ] Update race catalog documentation, roadmap, HISTORY and close marker.

## Exit criterion

Race choice is no longer bounded by the early Slice-16 compatibility subset: every preset race that Gate 2 classifies as mechanically complete is genuinely selectable through the server-owned race catalog, and a player can create a legal custom race from the same authoritative trait/pick system. Preset and custom races share one deterministic mechanics pipeline, survive save/resume/replay exactly, work with supported Human/AI controller paths, and the browser cannot create combinations that the server would reject.
