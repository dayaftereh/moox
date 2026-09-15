# Open Slice 16.3 - Galaxy visual selector

Opened: 2026-09-15
Status: **Gates 1-2 complete / Gate 3 next**

Planned specification: `docs/slices/PLANNED_16_3_GALAXY_VISUAL_SELECTOR.md`
Parent: `docs/slices/PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`
Permanent Gate-1 evidence: `docs/research/SLICE_16_3_GATE1_GALAXY_VISUAL_SELECTOR_AUDIT_2026-09-15.md`
Permanent Gate-2 freeze evidence: `docs/research/SLICE_16_3_GATE2_GALAXY_VISUAL_SELECTOR_FREEZE_2026-09-15.md`
Depends on: closed Slice 16.1 shared visual selection grammar; closed Slice 16.2 Difficulty selector/server-catalog pattern is a reusable reference, not a dependency for galaxy mechanics.

## Recovery state

- Slice 16.2 is closed at `713fcc7` and has no open marker.
- Exactly one `_OPEN_` marker should exist while this slice is active: this file.
- Gate 1 audit is complete. Gate 2 freeze is next; no new galaxy setting has been enabled yet.
- Preview convention: run MOX on `0.0.0.0:7171` with `-insecure-allow-nonloopback` so external/NetBird review stays available.

## Gate 1 audit

- [x] Re-check original galaxy size and age tables.
- [x] Map current generator support and star/system counts.
- [x] Identify settings that are runtime-ready without new generator mechanics.
- [x] Prototype one coherent image series for size and one for age.
- [x] Verify no illustration implies data the generator does not guarantee.

## Research questions

1. What exact public/original MOO2 galaxy-size and galaxy-age labels/semantics are supported by owned/local reference evidence?
2. Which size/age IDs already exist in MOOX New Game validation and generator code, and what system-count/planet-table effects are actually implemented?
3. Which existing Slice-16.1 Galaxy Size assets are presentation-only versus already bindable to server settings?
4. What honest image language can communicate size and age without implying deterministic star positions or unsupported astronomy/gameplay effects?

## Key Gate-1 findings

- Authoritative original size table has four rows: small/medium/large/huge = 20/36/54/71 stars on 5x4/6x6/9x6/9x8 grids.
- Slice-16.1 Tiny is presentation-only and has no authoritative fifth size row.
- Original Galaxy Age is Mineral Rich / Normal / Organic Rich, not young/normal/old.
- Ruleset already owns all four size rows, all three age IDs and all three spectral columns; runtime still accepts only Small + Normal.
- Gate 1 directly rechecked Organic-Rich climate weights from owned Orion2.exe object2:0x57CF; runtime normalization/selection remains Gate-2/3 work.
- Current New Game remains exactly two Human/Darlok players; do not invent size-dependent player-cap facts.
- Existing Small/Medium/Large/Huge art is the size prototype; new research-only age SVGs use a mineral-to-biosphere composition cue without changing star count/map shape.

## Gate-2 freeze summary

- Gameplay Size IDs: `small / medium / large / huge`; Tiny is not authoritative and will not appear in the bound selector/catalog.
- Galaxy Age IDs: `mineral_rich / normal / organic_rich`; default Normal.
- Exact Size facts: 20/36/54/71 star systems.
- Grid-jitter normalization generalized data-driven with cell size 200 and unchanged Small output/RNG order.
- Every Age profile owns its 7 spectral weights and 4x10 climate weights; Organic uses the rechecked 0x57CF table.
- Server catalog route is frozen as `GET /api/v1/new-game/galaxy`; frontend receives exact star counts plus qualitative mineral/food bias fields.
- Existing Small/Medium/Large/Huge art is production direction; Age research prototypes promote to kebab-case runtime SVGs.
- Current two-player Human/Darlok composition remains unchanged by Slice 16.3.

## Next

Gate 3 implementation is next. Do not reopen product semantics unless implementation uncovers contradictory evidence; implement the frozen ruleset/profile/catalog/generator/UI/test contract and checkpoint in small blocks.
