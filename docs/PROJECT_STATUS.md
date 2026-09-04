# Master of Orion X project status

Snapshot: **2026-09-02**

This document is the high-level status page for Master of Orion X. Detailed reverse-engineering evidence remains in `docs/research/`; the live handoff and next prepared objective are tracked in `docs/research/ACTIVE_RESEARCH.md`.

## Executive summary

Master of Orion X is now in **Phase 1 deterministic runtime development**, building on the advanced research/data-normalization baseline.

The project is no longer an empty bootstrap repository: it has a pure-Go Master of Orion II 1.31 analyzer, verified original-file parsers, normalized runtime datasets, localization keys, a private reference-extraction pipeline and substantial semantic graphics mapping derived from original 1.31 data/executable behavior.

The deterministic runtime now has its **first complete supported headless match lifecycle**. Slice 12 connects the existing New Game, Economy/Research/Construction, expansion, diplomacy/war, Fleet movement, tactical/encounter authority and invasion/Colony conquest paths to authoritative zero-Colony Empire elimination and an immutable conquest winner in `session.PhaseCompleted`. The canonical seed `0x8009` is played end-to-end using public application/session commands and exact replay/save-load verification. This does not mean broad MOO2 fidelity is complete: alternative victory families, deeper tactical/economy systems, AI, leaders/espionage and endgame presentation remain post-milestone work.

## Current repository / tool baseline

| Item | Current state |
| --- | --- |
| Branch | `main` |
| Go module | `moox`, Go `1.26` |
| Analyzer | `moox-analyze 0.22.0` |
| Core architecture | authoritative Go game server + browser-first HTTP/WebSocket HMI; Wails v3 optional one-click native wrapper only (`ADR-0004`) |
| CGO policy | core/analyzer/tooling kept pure Go / `CGO_ENABLED=0` where possible |
| Original gameplay baseline | Master of Orion II 1.31 |
| Research workflow | `docs/WORKING_RULES.md` + `docs/research/ACTIVE_RESEARCH.md` + `tools/research-preflight.ps1` |

## Post-milestone prepared roadmap

Slices 01-14 are closed. Slice 14 Live GameSession save/resume passed independent Gate-4 QA; implementation commit `a2c4ccd` plus Gate-4 rejection-coverage commit `e6a5e69` preserve Core23 plus Session/Battle continuation state and bind saves to deterministic simulation compatibility plus the exact compacted loaded-rules fingerprint. No `_OPEN_` slice marker remains. The roadmap state is:

1. Slice 08 - **closed** - authoritative server, web HMI and HTTP/WebSocket transport baseline (`9315111`).
2. Slice 09 - **closed** - deterministic New Game and galaxy generation baseline (`a5f3c13`).
3. Slice 10 - **closed** - diplomacy / war / peace baseline.
4. Slice 11 - **closed, Gates 1-4 complete** - Troop Transport, invasion and Colony conquest baseline (`5b363ce`).
5. Slice 12 - **closed** - Empire elimination and first deterministic headless victory loop (`90b83d8`).
6. Slice 13 - **closed** - deterministic built-in strategic AI baseline; implementation/evidence commit `371aa0c`.
7. Slice 14 - **closed; Gates 1-4 complete** - live GameSession save/resume baseline.
8. Slice 15 - **prepared parent epic** - playable browser strategic HMI, decomposed into:
   - 15.1 **open; Gates 1-3 complete, Gate 4 pending** - mobile-first UX architecture / navigation / bilingual design-system shell.
   - 15.2 **prepared** - functional strategic gameplay HMI.
   - 15.3 **prepared** - MOOX visual identity / graphics / asset pipeline.
   - 15.4 **prepared** - rich gameplay decisions / Encounter transitions / persistence UX.
   - 15.5 **prepared** - interactive 2D Tactical Combat: authoritative movement/legal targets plus mouse/touch battlefield on the Slice-07 combat baseline.
   - 15.6 **prepared** - full browser vertical slice / polish / complete-game QA.
9. Slice 16 - **prepared** - New Game + preset-race breadth.
10. Slice 17 - **prepared** - military ship design component/weapon breadth.

Application direction is accepted in `docs/architecture/ADR-0004-authoritative-server-web-client.md`. Slice 15.1 is open with Gates 1-3 complete and Gate 4 independent QA pending. Desktop remains a required adaptive layout. After 15.1 implementation/QA, the built UI must be exposed over the LLO Desktop NetBird interface for direct phone review. Slice 16/17 remain independent breadth work after the browser-playability family.

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
| `technologies.json` | 203 original technology identities/IDs + 82 TechFields/cost/topology metadata | starting ownership, breakthrough resolution, race-aware General/ordinary/Creative/Uncreative application selection and legal research actions are implemented; original Uncreative initial-selection timing/shared-RNG ownership and Hyper-Advanced repeated-field progression are verified; Advanced-start generator structure, original AI weighting metadata and the deterministic 19-grant runtime are implemented; broader technology effects remain |
| `buildings.json` | 48 original building IDs + original technology links | identities/links proven; costs, maintenance and gameplay effects remain |
| `ship_hulls.json` | 6 military hull identities | picture identities/mappings proven; full hull stats/components/design rules remain |
| `assets.json` | 156 semantic records | 143 confirmed, 13 deliberately pending generic race icons |
| `planet_classes.json` | 5 sizes, 5 mineral classes, 3 gravity classes, 10 climates | size thresholds, mineral extraction and base food/farmer proven; broader galaxy generation intentionally deferred |
| `economy.json` | base research/income/industry + Population/Sustenance + Food/Freighter/Starvation + Population Growth modifiers + Aquatic + gravity + starting-government + local Morale rules | base/contextual Economy, original-style round-robin insufficient-Freighter import priority, authoritative system-blockade exclusion from Food import/export/sale pools, exact four-pass mixed-cohort Food priority, Freighter balancing/cost snapshot, surplus Food valuation, Starvation, 50-PP Freighter Fleet production (+5 Freighters), authoritative modeled Treasury settlement, Population relocation through the shared Freighter pool, Housing, Microbiotics, Universal Antidote and Cloning Center growth effects, Advanced City Planning +5, Biospheres +2 and semantic Terraforming/Gaia climate transformations implemented; runtime Ship/Spy/Tribute/Officer/Leader Treasury categories, full deficit liquidation, broader buildings/pollution and advanced-government effects remain |

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

The Treasury Maintenance/deficit checkpoint is complete. Direct MOO2 1.31 evidence resolves all six Maintenance categories and the staged negative-Treasury liquidation boundary. Without changing Core schema 15 or ruleset schema 7, runtime now charges the complete currently-authoritative active-Freighter bucket (Food use plus Population-transfer reservations) with original aggregate whole-BC truncation and materializes surplus-Food income at the original whole-BC boundary. Full asset liquidation stays deferred until canonical Ship/Spy/Leader/treaty systems exist, and the full Go test/vet baseline is green.

The Colony Ship vertical-loop checkpoint is complete. That checkpoint introduced Core schema 16 with Technology-41 Colony Ship construction, original Fuel Cell ranges, persisted installed FTL speed, deterministic strategic transit before current-turn Construction and explicit same-System colonization. Successful colonization creates/links a second Colony with one assimilated founding Population unit and consumes the Colony Ship; save/load, Observer isolation/history and the full build -> move -> arrive -> colonize loop are covered by deterministic regressions in `db9d01e`.

The Colony Base same-system checkpoint is complete. Building 11 / Technology 40 now has original-style same-system buildability, generic 200-PP Construction completion, mandatory post-resolution colonize-or-trash handling, exact 100-BC trash refund and no source-Population subtraction. Pending decisions are derived from existing Building ownership, so Core stayed at schema 16 for that checkpoint; Colony Ship and Colony Base share the same normal-Colony initializer, and save/load, Observer and deterministic replay coverage are green in `261e418`.
Slice 02 **Outpost Ship / Outpost state / supply range** is closed in implementation commit `7c7284e`. Core schema 17 persists dedicated Planet-linked Outposts, Tech-109 fixed Outpost Ship construction and installed FTL state, shared Colony/Outpost special-ship Fuel-range transit, Colony+Outpost supply origins, explicit Planet-level Outpost deployment/ship consumption and same-owner Outpost-to-Colony replacement for Colony Ship and Colony Base founding. Blockade alone does not suppress an Outpost supply origin; hostile-control/conquest semantics remain deferred. Gate 4 passed fresh `gofmt`, full `go test ./... -count=1`, `go vet ./...`, focused Outpost/Colony Ship/Colony Base/Supply-range regressions and working/staged diff checks.
Slice 03 **Military Ship core / design baseline** is closed after Gate 4 final QA and commit closure. Core schema 18 now implements the accepted arbitrary-size ID-based current design catalog, immutable built-Ship snapshots, normalized original hull/mandatory-component tables, authoritative cleared-Frigate design creation, `military_ship` Construction and concrete one-Ship combat-Fleet completion at the producing System. The first exact baseline costs 25 PP (20 hull + 5 Electronic Computer), or 17 PP under the current Feudal ship-cost rule. Save/load, Observer isolation, deterministic replay and >6-design coverage are green under the full repository test suite; generic combat-Fleet movement, Command Points and tactical combat remain deferred to Slices 04-07.


Slice 04 **Combat Fleet movement / merge / split** is closed after Gate 4 final QA and two-commit closure. Core schema 19 permits ordinary combat-Fleet semantic transit with concrete non-empty `ShipIDs` while keeping combat `FTLSpeed` derived/non-authoritative. `empire.move_fleet` supports whole-Fleet or selected-subset movement, with subset movement atomically emitting `fleet_split` before `fleet_movement_started`; explicit stationary `split_fleet` / `merge_fleets` preserve exact composition and stable IDs. Current Empire best Warp Drive + Trans-Dimensional and current Fuel Cell capability drive movement legality, Colony/Outpost supply origins are shared, arrivals precede blockade recomputation, and hostile co-location remains the temporary boundary until Slice 06. Save/load, Observer isolation and identical-session replay coverage are green. Gameplay commit: `f297480`; next prepared objective is Slice 05 **Command Points / ship maintenance**.

Slice 05 **Command Points / ship Maintenance** is closed. Core schema 20 persists the last-settlement `CommandPoints {Capacity, Used}` snapshot and Treasury `ShipCommandMaintenanceBC`; economy ruleset schema 8 normalizes the original base/station/Communications/Warlord/Imperium/10-BC-overage rules while military usage remains derived from hull `SizeIndex + 1`. Colony/Outpost special Fleets cost 1 CP, Population transfers/freighters cost none, split/merge/movement preserve usage, and legal special-Ship consumption removes usage. Star Base -> Battlestation -> Star Fortress replacement/buildability is normalized. Treasury precomputes and validates every Empire before any mutation and remains pre-Construction, so new Ships/station upgrades first affect the following settlement. Gate 4 passed gofmt, focused original-accounting/compatibility regressions, `go test ./... -count=1`, `go vet ./...` and diff checks. Gameplay/evidence commit: `16afe0b`. At Slice 05 closure, Slice 06 Strategic hostile encounters / BattleSession handoff was the next prepared objective; Slice 06 is now closed; Slice 07 Tactical ship combat baseline is the next prepared objective. Leader/Transport/NPC-rate/deficit/tactical-station boundaries remain deferred.
**Phase 1 - deterministic simulation skeleton is active.** `colony.assign_population` now materializes base `Economy`, explicit `EconomyContext`, and `AdjustedEconomy` through a transactional resolver. Gravity, starting-government and local Morale are implemented: Feudal/Dictatorship -20% without Barracks, Holo +20%, Pleasure +30% cumulative, with Unification recording but suppressing Morale. The first single-project colony construction path is implemented: authoritative state now uses generic `project_kind` / `project_id` construction identity while the existing `building` project kind preserves normalized building PP costs/BC maintenance, ownership-validated `colony.queue_building`, domain-native fractional PP progress after Cybernetic sustenance, and observer/replay completion. A completed building cannot retroactively change consumed PP, but it is visible when the post-growth next-state snapshot is recalculated. Technology-gated buildability is now also enforced and `GameSession` can project authority-filtered building choices for the same legal-action surface used by future UI and AI controllers. Original technology-field tables and RP costs are now normalized as well; deterministic Pre-Warp/Average starts materialize known fields/technologies, and confirmed research completion is committed through `GameSession` into the strategic event stream. The standard-field breakthrough curve/RNG/overflow behavior is implemented. The numeric architecture is now generalized beyond Research: Population allocations, Food, Production, Research, BC and Construction progress use domain-native `float64`, preserving fractional values until explicit gameplay-rule rounding. `StateSchemaVersion` is 20; Population persists ordered organic cohorts with origin, loyalty, assimilation state and per-job quantities rather than a second mutable aggregate truth, Construction uses semantic project kind/id, Empire Treasury uses domain-native `float64` balance plus an auditable modeled-income/Maintenance snapshot, interstellar Population transfers are persisted as semantic Settler state, and strategic Fleets plus directed diplomatic stances now provide the canonical inputs for derived system blockades; schema 17 persists Colony Ship and Outpost Ship fixed-special identity, installed FTL speed and semantic destination/ETA transit state, plus dedicated Planet-linked Outpost occupancy and supply state. `ColonyPopulationDynamics` now materializes race-aware capacity, local/post-logistics Food, Production sustenance, Growth and Starvation. Empire state owns a discrete Freighter pool plus materialized transport/cost/surplus-Food accounting. Direct MOO2 1.31 executable analysis now fixes the apply order as Research -> Population Growth/Starvation -> Construction, while both RP and PP are consumed from the already-materialized pre-growth snapshot. `GameSession.ResearchChoices` now exposes race-aware `all`, `choose_one`, `fixed_one`, and Hyper-Advanced `repeat_field` policies. Ordinary races choose one server-validated application, Creative/General fields acquire all, and Uncreative uses a persisted server-fixed application plan. Active-project switching with exact accumulated-RP transfer is implemented. Original Uncreative initial fixed selections now consume the shared New Game RNG during player initialization and exclude TechField 74; Hyper-Advanced repeated-field costs/progression are implemented with `repeat_field`; Advanced start is implemented with a shared New Game RNG, cross-Empire competition weighting and exactly 19 extra grants. Authoritative external Technology grants now use the verified Uncreative post-acquisition repair, preserve incomplete TechFields, consume the authoritative State RNG, and emit strategic Observer/replay events.

## Current runtime gaps after Slice 12

The project now has a complete supported headless conquest lifecycle, so the remaining gaps are no longer accurately described as one undifferentiated "not implemented" list. The post-milestone audit in `docs/research/POST_MILESTONE_FIDELITY_DEPTH_BACKLOG_AUDIT_2026-09-03.md` classifies them as:

### P0 - independent playability blockers

- **Built-in AI:** `ControllerBuiltinAI` exists, but no runtime AI drives commands yet.
- **Live save/resume:** exact Core and completed-session snapshots exist, but no in-progress GameSession/hosted-game persistence surface exists.
- **Playable browser HMI:** the React client proves transport and a few commands, but it is not yet a complete galaxy/colony/research/construction/fleet workflow.

### P1 - high-value runtime breadth

- New Game currently supports only Small galaxy, Normal age, Average technology, Tactical combat, exactly two players and exactly Human+Darlok.
- The ruleset contains 13 normalized preset races, but New Game exposes only Human/Darlok; full custom Race Designer is not implemented.
- Military design persistence exists, but current validation deliberately permits only the minimal supported hull with slot 0 / one Laser Cannon.
- Tactical combat supports the Beam-fire/end-activation vertical slice; movement/facing, additional weapon families, internal-system depth, retreat/boarding and planetary defenses remain later work.

### P2 - fidelity/depth backlog

- treaty/alliance/non-aggression/tribute/technology-trade diplomacy;
- espionage/sabotage and Leaders/Officers;
- contextual Building/Pollution/Morale/Maintenance and staged deficit fidelity;
- occupation/assimilation plus Android/Native/custom-race Population behavior;
- full custom Race Designer/government/trait interactions;
- Galactic Council, Orion/Guardian, Antaran/dimensional-gate and other victory families;
- random events and broader late-game systems.

### P3 - productization / optional transports

- remote/external/MCP AI and full multiplayer lifecycle;
- independent final art/audio/UI skin/cinematics;
- optional native/Wails packaging over the same server authority.

## Next milestones

### Prepared tranche - Slices 13-17

1. **Slice 13 - Built-in strategic AI baseline:** deterministic rules-legal/no-cheat AI over the current complete lifecycle.
2. **Slice 14 - Live GameSession save/resume baseline:** versioned in-progress persistence and exact continuation at stable interactive boundaries.
3. **Slice 15 - Playable browser strategic HMI parent epic:** 15.1 UX/navigation/design system -> 15.2 functional strategic HMI -> 15.3 MOOX visual identity/assets -> 15.4 rich decisions/persistence UX -> 15.5 interactive 2D Tactical Combat -> 15.6 complete browser vertical slice.
4. **Slice 16 - New Game + preset-race breadth:** widen the current Small/Normal/Average/two-player/Human+Darlok boundary using evidence-backed settings and preset races.
5. **Slice 17 - Military ship design component/weapon breadth:** expand the current one-Laser minimal design for broader combat/design fidelity beyond the first interactive Slice-15.5 Tactical baseline.

### Next product milestone

After Slice 15.6, a human should be able to start, save/resume, play the accepted interactive Tactical battle path and finish the supported deterministic single-player match from the browser without any gameplay authority living in React, through a coherent MOOX visual and interaction language.

### Unnumbered depth after Slice 17

Re-audit before numbering **additional Tactical breadth beyond Slice 15.5**, treaty/trade diplomacy, deeper Espionage/Leaders, Economy/Population fidelity, full custom Race Designer, Council/Orion/Antaran victories, random events, multiplayer/external-AI transports and presentation/native packaging.

No prepared slice is an active implementation slice until a fresh Gate 1 creates the corresponding `_OPEN_*.md` marker.

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


Turn-order fidelity checkpoint: docs/research/TURN_ORDER_2026-08-28.md documents direct MOO2 1.31 control/data-flow proof that Research applies before Population and Construction while both RP and PP are consumed from the pre-growth materialized snapshot.

Uncreative fidelity checkpoint: docs/research/UNCREATIVE_INITIAL_SELECTION_2026-08-28.md records direct original evidence for New Game selection timing, shared RNG ownership, TechField 1..73 range and race/mode rejection rules.

Hyper-Advanced fidelity checkpoint: `docs/research/HYPER_ADVANCED_RESEARCH_2026-08-28.md` records the original per-player counters, `+10000 RP` dynamic strategic cost, repeat completion semantics and the deliberately-unreproduced UI preview off-by-one.

Slice 06 **Strategic hostile encounters / BattleSession handoff** is closed. Gameplay/evidence commit `096fd0a` implements the staged encounter continuation after Fleet transit + Construction and before Blockades/transfers, deterministic directed-hostile Empire-aggregate BattleSession waves, explicit attacker/defender Fleet/Ship/civilian/Colony identity, singular validated battle results, atomic final-wave commit/retry semantics, concrete casualty cleanup, civilian-only overrun and closest-other-owned-Colony retreat/destruction. Core schema 20/economy schema 8 remain unchanged. Gate 4 passed gofmt, focused CombatFleet/Blockade/Encounter/BattleSession ordering regressions, full `go test ./... -count=1`, `go vet ./...` and diff checks. Slice 07 **Tactical ship combat baseline** is now closed; gameplay/data/evidence commit `889f977` completed the narrow tactical baseline, and no later slice is open.

Slice 07 **Tactical ship combat baseline** is closed. Gameplay/data/evidence commit `889f977` advances Core to schema21 with persistent weapon snapshots and adds tactical ruleset schema1 plus the exact Fusion-Laser-Frigate vs unarmed Nuclear-Frigate deterministic BattleSession path. Gate 4 re-ran focused battle/session/strategic handoff and fixed-seed regressions, verified wall-clock-independent Battle-ID ordering and exact strategic casualty reconciliation, then passed full `go test ./... -count=1`, `go vet ./...` and diff checks. No tactical scope beyond the accepted Slice-07 fixture was added.
