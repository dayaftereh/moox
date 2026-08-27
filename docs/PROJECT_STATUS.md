# Master of Orion X project status

Snapshot: **2026-08-27**

This document is the high-level status page for Master of Orion X. Detailed reverse-engineering evidence remains in `docs/research/`; the single active investigation is tracked in `docs/research/ACTIVE_RESEARCH.md`.

## Executive summary

Master of Orion X is now in **Phase 1 deterministic runtime development**, building on the advanced research/data-normalization baseline.

The project is no longer an empty bootstrap repository: it has a pure-Go Master of Orion II 1.31 analyzer, verified original-file parsers, normalized runtime datasets, localization keys, a private reference-extraction pipeline and substantial semantic graphics mapping derived from original 1.31 data/executable behavior.

The game is **not playable yet**, but the deterministic core and multiplayer/session boundary now exist. The repository has deterministic state/RNG/save-load plus versioned command batches, an atomic transport-independent strategic resolver contract, authoritative session phases, parallel seat submissions, player/observer projections, observer-only draft telemetry and minimal parallel tactical battle sessions. Gameplay command resolution, economy, fleets/combat rules, AI behavior and application UI remain to be implemented.

## Current repository / tool baseline

| Item | Current state |
| --- | --- |
| Branch | `main` |
| Go module | `moox`, Go `1.26` |
| Analyzer | `moox-analyze 0.22.0` |
| Core architecture | Go-first, deterministic/headless core; Wails v3 planned only as application shell |
| CGO policy | core/analyzer/tooling kept pure Go / `CGO_ENABLED=0` where possible |
| Original gameplay baseline | Master of Orion II 1.31 |
| Research workflow | `docs/WORKING_RULES.md` + `docs/research/ACTIVE_RESEARCH.md` + `tools/research-preflight.ps1` |

Important recent checkpoints:

- `980098c feat: normalize player ship hulls`
- `7e2125e feat: map strategic ship graphics from original executable`
- `53d3a6b feat: map tactical ship frames from original executable`
- `873f88f docs: define convergent research workflow`

## Original 1.31 reference inventory

Private local reference: `C:\ASH\Temp\mastori2`.

The tracked inventory metadata reports:

- 420 files,
- 338,042,423 bytes total,
- 363 SimTex LBX containers,
- 10 Smacker files stored under `.LBX` names,
- 47 other/support files.

Original binaries/archive bytes and extracted copyrighted artwork/audio remain private reference material and are not committed as distributable MOOX content.

Private extraction/reference tooling currently includes:

- recursive file/hash inventory,
- LBX parsing and raw extraction,
- block classification,
- graphics catalog/PNG extraction,
- external palette cataloging/resolution,
- RIFF/WAVE extraction,
- text cataloging,
- support-file snapshotting,
- bound MZ/LE reading for original DOS executable tables/code ranges.

The current canonical private graphics export contains 30,395 PNGs from 30,929 cataloged frames; unresolved palette contexts stay explicitly pending rather than being color-guessed.

## Normalized runtime data

Tracked normalized data lives under `data/rulesets/moo2-1.31/`.

| Dataset | Current normalized coverage | Status / important gap |
| --- | --- | --- |
| `race_traits.json` | 11 Race Designer groups, 53 selectable options | identities/options/pick model present; some exact behavior/cost parity work remains |
| `races.json` | 13 standard races | stable preset identities/traits/provenance present |
| `technologies.json` | 203 original technology identities/IDs | research fields, costs, unlock/effect formulas still need normalization |
| `buildings.json` | 48 original building IDs + original technology links | identities/links proven; costs, maintenance and gameplay effects remain |
| `ship_hulls.json` | 6 military hull identities | picture identities/mappings proven; full hull stats/components/design rules remain |
| `assets.json` | 156 semantic records | 143 confirmed, 13 deliberately pending generic race icons |
| `planet_classes.json` | 5 sizes, 5 mineral classes, 3 gravity classes, 10 climates | size thresholds, mineral extraction and base food/farmer proven; broader galaxy generation intentionally deferred |
| `economy.json` | base research/income/industry + Aquatic + 3x3 gravity matrix + four starting-government economy modifiers | base + Gravity/Government layers implemented; morale/buildings/pollution/logistics and advanced-government effects remain |

### Semantic asset coverage

`assets.json` currently contains:

- 14 race portraits,
- 52 race role icons,
- 13 deliberately pending generic race-icon keys,
- selected UI/background/production references,
- 48 building colony sets with **1,728** original-position variants,
- 6 military strategic hull sets + Colony/Outpost/Transport with **352** strategic ship variants,
- 6 military tactical hull sets with **6,560** tactical frame variants.

The tactical ship mapping is original-executable-derived: standard player hulls use the same picture IDs in strategic and tactical lookup, and each tactical picture contains five folded orientation groups x four neutral animation phases.

## Localization status

Runtime logic references stable translation keys; user-visible strings are not embedded in ruleset identities.

Current committed language files:

| Locale | Keys | Current coverage |
| --- | ---: | --- |
| English (`en`) | 357 | Race Designer + race names + buildings + technologies + ship hull names |
| German (`de`) | 64 | Race Designer; English fallback for other current keys |
| French (`fr`) | 64 | Race Designer; English fallback for other current keys |
| Spanish (`es`) | 64 | Race Designer; English fallback for other current keys |
| Italian (`it`) | 64 | Race Designer; English fallback for other current keys |

Original localized glyph substitutions are only decoded where the mapping is evidenced; no mojibake is intentionally promoted into runtime language files.

## Closed research areas

The following research slices have stable artifacts/tests/commits and should not be broadly rediscovered unless contradictory evidence appears:

- generic LBX structure and inventory pipeline,
- core graphics decoding and verified palette contexts already documented,
- Race Designer option identities and language-key architecture,
- 13 preset race identities,
- 203 technology identities,
- 48 building IDs and building -> technology table links,
- building `BLDG0..4` archive/group/6x6 position mapping,
- six player military hull identities,
- standard strategic player ship picture/color mapping,
- Colony Ship / Outpost Ship / Transport strategic picture identities,
- standard tactical `CMBTSHP` player-color mapping,
- tactical 20-frame structure (5 folded orientations x 4 animation phases).

See `docs/research/ACTIVE_RESEARCH.md` for the explicit closed-milestone list and commit references.

## Active work

The planet-class normalization checkpoint is complete. The temporary probe programs have been removed, the normalized artifact is committed-ready, and the full Go test/vet baseline is green.

**Phase 1 - deterministic simulation skeleton is active.** `colony.assign_population` now flows through trusted Seat -> Empire authorization and a transactional `EconomyResolver` that materializes three explicit views: base `Economy`, `EconomyContext`, and `AdjustedEconomy`. Gravity penalties (0/25/50%) and the starting-government food/industry/research/tax effects are normalized and tested, including Unification morale-ignore. The next fidelity slice is morale plus the minimum building/technology coefficients needed to test it.

## What is not implemented yet

The research/data tooling should not be confused with a playable engine. Major missing runtime systems include:

- original-faithful galaxy/star/planet generation (the current small galaxy is deterministic test scaffolding),
- full strategic turn processing beyond the first population/economy resolver,
- contextual/net colony economy beyond the implemented Gravity/Government layer (morale, buildings, pollution, logistics, maintenance),
- race-aware population cohorts for conquered/mixed-race colonies and persisted custom race designs,
- high-precision population growth/capacity and food consumption,
- research progression/effects,
- strategic fleet movement/colonization,
- save-format versioning/migrations beyond the current exact state round trip,
- race-government runtime modifiers,
- ship designer/components/weapons/specials,
- tactical combat rules,
- diplomacy/espionage/leaders,
- conquest/victory/Antaran/Orion systems,
- AI,
- Wails application shell/UI.

## Next milestones

### Immediate research checkpoint

Planet-class normalization is complete. Preserve `docs/research/PLANET_CLASSES_2026-08-27.md` as the checkpoint and do not expand back into planet graphics or broad galaxy-generation research without a concrete simulation requirement.

### Engineering transition

The initial Phase 1 foundation and first strategic command are complete. Preserve the current deterministic/session contract while extending effective colony output in independently proven layers:

1. gravity compatibility and production penalties,
2. government production/research/income effects,
3. morale interaction,
4. building/technology flat and per-population bonuses,
5. pollution and pollution-control processing,
6. food consumption, freighters and surplus-food handling.

Do not collapse these into one formula until ordering and rounding behavior are evidenced.

### First headless vertical slice

The first meaningful game loop remains:

1. generate a small deterministic galaxy,
2. create one empire/homeworld,
3. assign population roles,
4. process food/production/research/money,
5. research a technology,
6. build a colony ship,
7. move to a second system,
8. colonize,
9. save and reload exactly the same state.

UI work comes after this loop is stable; Wails v3 is already the planned application shell, so UI-framework selection is no longer an open architectural decision.

## Where to look

- `docs/research/ACTIVE_RESEARCH.md` - single current research ticket and exact next action.
- `docs/WORKING_RULES.md` - anti-drift/research/checkpoint rules.
- `docs/IMPLEMENTATION_PLAN.md` - runtime implementation phases.
- `docs/ANALYZER.md` - analyzer commands and normalized-data generation.
- `docs/architecture/ADR-0001-go-wails-v3.md` - architecture decision.
- `docs/architecture/ASSET_PIPELINE.md` - original-reference -> semantic-spec -> independent MOOX art pipeline.
- `data/rulesets/moo2-1.31/README.md` - normalized ruleset provenance/details.
- `docs/research/MOO2_SHIP_GRAPHICS.md` - current strategic/tactical ship evidence.
- `docs/research/MOO2_BUILDING_GRAPHICS.md` - original building graphic/index formula evidence.
