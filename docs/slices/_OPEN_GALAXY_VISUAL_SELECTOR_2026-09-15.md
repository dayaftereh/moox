# Open Slice 16.3 - Galaxy visual selector

Opened: 2026-09-15
Status: **Gate 1 active**

Planned specification: `docs/slices/PLANNED_16_3_GALAXY_VISUAL_SELECTOR.md`
Parent: `docs/slices/PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`
Depends on: closed Slice 16.1 shared visual selection grammar; closed Slice 16.2 Difficulty selector/server-catalog pattern is a reusable reference, not a dependency for galaxy mechanics.

## Recovery state

- Slice 16.2 is closed at `713fcc7` and has no open marker.
- Exactly one `_OPEN_` marker should exist while this slice is active: this file.
- Gate 1 is audit/research only. Do not perform Gate-2 freeze or Gate-3 generator/runtime implementation yet.
- Preview convention: run MOX on `0.0.0.0:7171` with `-insecure-allow-nonloopback` so external/NetBird review stays available.

## Gate 1 audit

- [ ] Re-check original galaxy size and age tables.
- [ ] Map current generator support and star/system counts.
- [ ] Identify settings that are runtime-ready without new generator mechanics.
- [ ] Prototype one coherent image series for size and one for age.
- [ ] Verify no illustration implies data the generator does not guarantee.

## Research questions

1. What exact public/original MOO2 galaxy-size and galaxy-age labels/semantics are supported by owned/local reference evidence?
2. Which size/age IDs already exist in MOOX New Game validation and generator code, and what system-count/planet-table effects are actually implemented?
3. Which existing Slice-16.1 Galaxy Size assets are presentation-only versus already bindable to server settings?
4. What honest image language can communicate size and age without implying deterministic star positions or unsupported astronomy/gameplay effects?

## Next

Complete Gate-1 audit/evidence and prototypes, update the Gate-1 checklist, commit/checkpoint the research result, and stop before Gate 2 unless explicitly continued.
