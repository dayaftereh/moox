# Slice 15.3 - Playability visual baseline freeze

Date: **2026-09-09**

Decision: **accepted as the baseline for moving toward a playable vertical slice; richer art iteration intentionally deferred**.

## Frozen baseline

- Strategic direction: **Deep Space Command**.
- Canonical development/review port: **7171**.
- Responsive floor: **320 CSS px**, no document-level horizontal scroll.
- Galaxy stars: class-colored luminous **StarArt** (B/F/G/K/M/BD/BH), with System-dialog reuse.
- Orbital bodies: spherical climate-driven **OrbitalBodyArt** with deterministic noise, atmosphere and clouds where appropriate.
- Buildings/stations: reusable **BuildingArt**; dedicated Capitol/Colony Base/Marine Barracks/Star Base/Battlestation/Star Fortress plus deterministic family coverage for current ruleset IDs. Housing remains a repeatable project, not a persistent building.
- Ships: persisted **Visual Genome v4** procedural visual grammar; gameplay design/equipment breadth remains Slice 17.
- UI semantics: typed 24x24 SVG icon grammar and semantic action/status colors.
- Provenance: no unlicensed original-game asset shipping; reference material is evidence only.

## Why this is frozen now

The visual layer is now materially above placeholder quality and sufficient to support real gameplay iteration. Continuing to polish art before exercising Encounter, persistence, Tactical Combat and the end-to-end browser game would optimize the wrong bottleneck. Full-game playthroughs will provide better evidence for the next art pass.

## Deferred art backlog

- Colony/planet-surface building composition.
- Individual bespoke artwork for all later buildings.
- richer race portraits and colony/environment scenes.
- additional star/planet/material tuning.
- motion/effects/audio polish.

These are **not cancelled**. They are explicitly parked until the playable vertical slice exposes where visual investment produces the most value.

## Carry-forward contract for 15.4-15.6

15.4-15.6 may rely on the current visual components and semantic IDs without reopening 15.3. New gameplay work should prefer existing visual primitives/fallbacks and must not make decorative art authoritative for game rules.

## Evidence

- Gate-1 direction package and styleboards.
- `docs/research/SLICE_15_3_ART_FIDELITY_PLANETS_BUILDINGS_2026-09-08.md`.
- `docs/research/SLICE_15_3_SPECTRAL_STARS_BUILDING_COVERAGE_LAYOUT_FIX_2026-09-08.md`.
- live `/styleboard.html`.
- live `/art-fidelity.html`.
- full regressions at commit `e251c32`: Go tests, `go vet`, production web build and `git diff --check` all passed.
