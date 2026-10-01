# Slice 17.6 Gate 3 - Implementation Checkpoint Plan

Date: 2026-09-30
Status: **ACTIVE - checkpoints 1-3 complete; checkpoint 4 ready**

Contract authority:

- `docs/research/SLICE_17_6_GATE2_TACTICAL_BEAM_CONTRACT_FREEZE_2026-09-30.md`

Gate 3 is split into five bounded recovery checkpoints. Every checkpoint must leave the repository committed and the affected test surface green.

## Checkpoint 1 - shared Beam rules + Selector preview

Status: **COMPLETE - 2026-10-01**

Deliver:

- pure shared server-side Beam configured-effect derivation;
- original percentage rounding helper;
- configured base-vs-configured damage;
- shots-per-weapon / range-profile / static To-Hit / effect IDs;
- `MilitaryDesignWeaponRowPreview` Tactical Beam preview;
- deterministic tests including the Laser 1-4 @ 90% rounding edge.

Do not widen Tactical fire execution yet.

## Checkpoint 2 - generic Beam execution + arcs + Count/Autofire

Status: **COMPLETE - 2026-10-01**

Deliver:

- executable `TacticalSupport="beam"` for the 13 frozen Beam identities;
- remove Laser-only fire predicate from authoritative path;
- firing-arc target legality;
- physical Count outer loop;
- Autofire inner three-shot ordering;
- frozen hit/range/RNG/readiness behavior;
- all rejected actions preserve zero RNG/readiness/events.

Preserve all existing Laser regressions.

## Checkpoint 3 - Beam effects + minimal Marine state

Status: **COMPLETE - 2026-10-01**

Deliver:

- Heavy / Point Defense normal ship-target effects;
- selected modifiers;
- innate Beam effects;
- exact damage percentage/rounding;
- Shield/Armor/Structure application hooks;
- `MarineMax` / `MarineCurrent`;
- original hull Marine values and passive Troop Pods x2 MarineMax input;
- Neutron/Death marine loss;
- deterministic reference vectors A-D.

Deferred Slice-17.8 hooks remain explicit and neutral when the later consumer is absent.

## Checkpoint 4 - reactive PD handoff + events/VFX identity

Deliver:

- server Beam-PD reactive capability for later 17.7 projectile actors;
- normal activation readiness restoration after reactive PD;
- frozen Beam event identity fields;
- Enveloping per-facing application identity;
- grouped/Autofire presentation facts;
- browser presentation uses events only.

No projectile lifecycle is implemented here.

## Checkpoint 5 - persistence / replay / Observer / integration evidence

Deliver:

- active Tactical save/load/replay with configured Beams and MarineCurrent;
- deterministic replay/event/RNG controls;
- Observer/player/public projection checks;
- Selector -> Save -> Build -> BattleSession -> Tactical integration;
- all five Gate-1 reference vectors as golden regressions;
- full Go/vet/web/browser/reference-game QA;
- Gate-3 evidence document.

Gate 4 begins only after all five checkpoints are committed and green.
