# Planned slice 21 - Colony / Buildings / Planet Effects / Pollution Fidelity

Status: **planned / queued; not open**.

Queue position: **reserved future Slice 21 by product decision**. Slice 21 is the dedicated strategic Colony/Economy fidelity milestone between Slice 20 Espionage and Slice 22 Race Completion. It is also a deliberate prerequisite for the deferred full Advanced-start work in Slice 23.

## Objective

Complete the Colony / Building / Planet-effect foundation far enough that normalized buildings and planet mechanics have real authoritative gameplay effects rather than existing mainly as catalog, buildability or maintenance data.

This is an evidence-driven fidelity pass, not an uncontrolled Economy rewrite. The target includes:

- all normalized building effects required by supported gameplay;
- colony production, research, food and economy modifiers;
- population growth/capacity/support effects;
- morale and colony-status effects where evidence/runtime support them;
- Command Point and orbital-station effects;
- planet size, climate, mineral, gravity and special effects;
- planet transformation and habitability;
- **Pollution as a first-class authoritative Economy effect**;
- server projection of these values to Strategic HMI only after the server owns them.

## Why Slice 21 exists

Gate-2 research for Slice 16.7 Advanced exposed a concrete downstream dependency:

- Core/Persistence can already carry many Colonies and variable Population;
- all **48 normalized buildings** exist in data/rulesets/moo2-1.31/buildings.json;
- buildability, costs, maintenance and technology links are broadly normalized;
- many building-specific gameplay effects are still partial or absent;
- full original Advanced starts can grant buildings that would therefore be mechanically inert today;
- original Advanced territory balancing also needs richer planet-effect/twiddle semantics than the narrow current New Game homeworld path.

Slice 21 turns those known gaps into a dedicated gameplay milestone before Race Completion and later Advanced-start parity.

## Known normalized foundation

Current repository data/research already provides:

- 48 normalized building definitions in buildings.json;
- normalized building technologies and original descriptions in technologies.json;
- Colony economy coefficients in economy.json;
- planet climate/class data in planet_classes.json;
- persisted Colony.Buildings, race-aware Population cohorts, planet climate/size/mineral/gravity, Treasury and Command Points;
- existing runtime consumers for selected building families such as command stations, Biospheres, Cloning Center, barracks/morale and maintenance/buildability;
- permanent Economy, Construction, Morale, Population, Planet and Command-Point research.

Gate 1 must re-audit the then-current runtime rather than assume every normalized building still lacks its effect.

## Pollution is explicitly in Slice 21

Pollution is a required Slice-21 gameplay effect, not a visual-only status field.

Repository evidence already records Pollution as a missing contextual Economy layer. Strategic HMI must not fabricate it before the server owns it.

Normalized original descriptions already establish important mechanics that Gate 1 must verify and freeze precisely:

- **Pollution Processor**: only half of actual planetary production is used when calculating Pollution.
- **Atmospheric Renewer**: production is quartered before Pollution is calculated and is cumulative with Pollution Processor.
- **Core Waste Dumps**: eliminates planetary Pollution.
- **Recyclotron**: +1 production per Population; this bonus does not count toward Pollution.
- **Tolerant**: current research logic treats Pollution-management applications as redundant; Gate 1 must re-confirm the exact production/Pollution behavior rather than infer it only from chooser exclusions.

The authoritative model must distinguish at least:

1. gross worker/building/flat production;
2. Pollution-relevant production;
3. Pollution-control reductions;
4. Pollution loss;
5. final effective production;
6. explicitly Pollution-exempt production.

Pollution must never be a frontend-only percentage disconnected from production.

## Building-effect families to audit

Gate 1 must classify all 48 normalized buildings as complete / partial / missing against real runtime consumers.

### Food / population / biosphere

Hydroponic Farms, Subterranean Farms, Food Replicators, Soil Enrichment, Biospheres, Cloning Center, Terraforming, Gaia Transformation and Weather Control System.

### Industry / Pollution

Automated Factories, Robotic Factory, Robo Miners, Deep Core Mining, Recyclotron, Pollution Processor, Atmospheric Renewer and Core Waste Dumps.

### Research

Research Laboratory, Autolab, Astro University, Planetary Supercomputer and Galactic Cybernet.

### Economy / trade

Spaceport, Galactic Currency Exchange, Planetary Stock Exchange and connected morale/economy buildings.

### Morale / military population

Holo Simulator, Marine Barracks, Armor Barracks, Capitol, Alien Management Center and conquered-population/assimilation dependencies.

### Command / orbital / defense

Star Base, Battlestation, Star Fortress, Space Academy, Planetary Missile Base, Fighter Garrison, Ground Batteries, planetary shields, Warp Interdictor, Artemis System Net, Dimensional Portal and Stellar Converter.

### Planet transformation / special infrastructure

Planetary Gravity Generator, Artificial Planet, Terraforming/Gaia facilities and any accepted effect that mutates climate, size, gravity, mineral, habitability or strategic planet identity.

If an effect depends on Tactical, Invasion, Diplomacy, Leaders or another incomplete subsystem, Gate 2 must explicitly name that dependency. Do not mark a building complete merely because it can be constructed.

## Planet-effect breadth

Slice 21 must audit authoritative effects and mutation rules for:

- climate;
- planet size;
- mineral richness;
- gravity;
- population capacity;
- farming capability;
- terraforming and Gaia transition;
- gravity normalization;
- artificial planets;
- normalized planet specials / artifacts where represented and required downstream;
- every planet value used by the original Advanced territory worth/twiddle logic that Slice 23 will need.

Gate 1 must distinguish data already represented in core.Planet, normalized data not yet represented, new persisted fields, and presentation-only identity.

## Dependencies / opening guard

Before Slice 21 moves beyond Gate 1, verify:

- current Economy/Population/Morale/Construction baselines;
- Slice 17 where orbital/defensive buildings need ship/Tactical consumers;
- current planet-generation and persistence schema;
- exact original Pollution ordering and rounding;
- exact building-effect provenance.

Slice 21 does not require Slice 22. Slice 22 is downstream and should consume Slice-21 mechanics.

## Binding architecture direction

- Building effects live in server-owned gameplay rules, not UI conditionals.
- Colony.Buildings remains authoritative ownership; derived Colony/Empire values are recalculated from rules/state.
- Equal state + equal rules derive equal production, Pollution, morale, Command Points and planet effects.
- Production-source categories must be explicit enough to support Pollution-exempt production correctly.
- Empire-wide effects must not accidentally be applied per Colony.
- Planet mutations must be persisted/replayable and deterministic.
- HMI may display Pollution/building/planet effects only from server-owned projections.
- Race traits such as Tolerant reuse the same Economy/Pollution pipeline.
- Slice 23 must reuse these mechanics; no Advanced-only duplicate building/planet implementation.

## Gate 1 - building / planet / Pollution audit

- [ ] Re-audit all 48 normalized buildings against actual runtime consumers and classify every effect complete / partial / missing.
- [ ] Re-audit original/normalized building effects, scope, ordering, rounding, buildability and technology requirements.
- [ ] Audit the complete Pollution formula and rounding order from owned original evidence.
- [ ] Audit production-source categories and identify which sources count toward Pollution and which are exempt.
- [ ] Re-confirm Pollution Processor, Atmospheric Renewer, Core Waste Dumps and Recyclotron semantics.
- [ ] Re-confirm Tolerant Pollution behavior and technology redundancy.
- [ ] Audit climate/size/mineral/gravity/special effects against current core.Planet.
- [ ] Audit planet transformation and persistent mutation semantics.
- [ ] Audit orbital/defensive building consumers against the then-current Slice-17/Tactical runtime.
- [ ] Audit morale/conquered-population building dependencies.
- [ ] Produce a building-effect matrix with subsystem owner and downstream dependency for every partial/missing effect.
- [ ] Present exact Gate-2 implementation scope and explicit deferrals.

## Gate 2 - authority freeze

- [ ] Freeze authoritative production -> Pollution -> effective-production ordering and rounding.
- [ ] Freeze all Pollution-control, exemption and Tolerant behavior.
- [ ] Freeze the building-effect set Slice 21 guarantees complete.
- [ ] Freeze explicit deferred building effects and owning subsystem.
- [ ] Freeze per-Colony versus per-Empire effect scope.
- [ ] Freeze planet mutation/effect schema required by the accepted set.
- [ ] Freeze persistence/replay behavior for changed planet state.
- [ ] Freeze server projection for Colony/planet/Pollution HMI.
- [ ] Freeze deterministic fixtures spanning representative building/effect combinations.
- [ ] Freeze cross-race fixtures needed by Slice 22 and cross-start fixtures needed by Slice 23.

## Gate 3 - implementation

- [ ] Implement the frozen production/Pollution pipeline.
- [ ] Implement all accepted Pollution-control and Pollution-exempt production sources.
- [ ] Implement/finalize frozen building effects through normal Economy/Population/Research/Morale/Command/Tactical owners.
- [ ] Implement accepted planet mutations/effects and persisted state.
- [ ] Wire Tolerant and other race modifiers through the same pipeline.
- [ ] Extend server snapshots/projections with Pollution and accepted derived values.
- [ ] Extend Strategic HMI only after server projections exist.
- [ ] Add deterministic unit/integration fixtures for every newly completed effect family.
- [ ] Add crafted-state regressions for stacking, ordering and rounding.
- [ ] Add save/resume/replay coverage for planet mutations and building effects.

## Gate 4 - QA + close

- [ ] Every building marked Slice-21-complete has its promised gameplay effect active.
- [ ] No supported building is mechanically inert where Gate 2 promised an effect.
- [ ] Pollution changes effective production exactly according to the frozen contract.
- [ ] Pollution-control buildings/technologies and Tolerant stack or short-circuit correctly.
- [ ] Pollution-exempt production remains exempt.
- [ ] Colony economy remains deterministic across save/resume/replay.
- [ ] Planet mutations persist and reproduce exactly.
- [ ] Command Point/orbital effects agree with fleet/runtime state.
- [ ] Strategic HMI shows only authoritative Pollution/building/planet facts.
- [ ] Representative desktop + 390 px Colony views remain usable.
- [ ] Full Go tests, vet, Web build/browser smoke and git diff --check pass.
- [ ] Re-audit Slice-22 and Slice-23 opening dependencies, update roadmap/HISTORY and close marker.

## Exit criterion

The Colony/Building/Planet foundation is broad enough that supported normalized buildings have real authoritative effects, Pollution is a deterministic server-owned part of production, important planet mutations/effects are persisted and replay-safe, and downstream Race Completion and Advanced-start work can reuse those mechanics without special-case substitutes.
