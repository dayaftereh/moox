# Slice 17.6 Gate 3 - Implementation Evidence

Date: 2026-10-01
Status: **Checkpoints 1-4 complete; Checkpoint 5 ready**

Contract authority:

- `docs/research/SLICE_17_6_GATE2_TACTICAL_BEAM_CONTRACT_FREEZE_2026-09-30.md`

## Checkpoint 1 - shared Beam rules + Selector preview

### Shared server Beam derivation

Added `internal/beam/configured.go` as the shared server-side home for design-static Beam derivation. It is independent of both `internal/game` and `internal/battle`, so later Tactical execution can consume the same rules without creating a game/battle package cycle.

The shared preview freezes:

- family = `beam`;
- base min/max damage;
- configured min/max damage;
- shots per physical weapon;
- range profile (`standard`, `heavy`, `point_defense`);
- static per-shot To-Hit adjustment;
- canonical configured/innate effect IDs.

Configured damage uses the Gate-2 neutral context:

- Standard = 100%;
- Heavy = 150%;
- Point Defense = 50%;
- no range/attacker/target/Special runtime contribution.

### Original percentage rounding

`beam.RoundPercent` and `beam.ScaleDamageBounds` implement the frozen original order:

`scaled = (base * percent + 50) / 100`

then:

- configured minimum clamps to at least 1;
- configured maximum clamps to at least configured minimum.

The canonical Laser 1-4 @ 90% edge is regression-covered and returns **1-4**, not the old truncating **0-3** result.

This checkpoint does **not** change the current Tactical Laser execution helper yet; that runtime migration is intentionally deferred to later Gate-3 checkpoints against the frozen shared derivation.

### Mount / modifier preview facts

Configured derivation exposes:

- Heavy -> `range_profile=heavy`, 150% configured damage;
- Point Defense -> `range_profile=point_defense`, 50% configured damage and +25 static To-Hit;
- Autofire -> 3 shots per physical weapon and -20 static per-shot attack adjustment;
- Continuous -> +25 static attack adjustment;
- selected Beam modifier IDs as canonical effect IDs.

Frozen executable-native effects are normalized for truthful preview, including:

- natural Shield Piercing;
- Armor-first bypass;
- monster/Antaran quarter-damage fact;
- extra Structure damage;
- marine kill;
- Enveloping;
- doubled signed range To-Hit;
- Always Hit;
- Death Ray natural Continuous.

### Military Design / Selector transport

`quoteBeamWeapon` now attaches the shared Beam configured preview after existing legality/canonicalization succeeds.

`MilitaryDesignWeaponRowPreview` gains optional:

- `beam_tactical`

with the exact Gate-2 JSON semantics. Non-Beam families keep the field absent.

The TypeScript API mirrors the same shape in `MilitaryDesignBeamTacticalPreview`.

### Selector presentation

The Weapon Selector consumes only server-returned `beam_tactical` values. React performs no Beam formula.

The selected Beam configuration now displays:

- configured damage separately from the existing base-damage display;
- shots per weapon;
- range profile;
- static To-Hit adjustment;
- configured/innate effect labels.

English/German labels were added for the new preview facts and innate effect descriptions.

## Checkpoint 1 regression evidence

Passed on the Checkpoint-1 working tree:

- `go test ./internal/beam -count=1`
- focused `TestSlice172MilitaryDesignPreview`
- `go test ./internal/server -count=1`
- `go test ./internal/beam ./internal/game ./internal/server -count=1`
- `go vet ./...`
- `npm --prefix web run build`
- temporary reference server `127.0.0.1:7216` health check
- `npm --prefix web run check:reference-lab:browser -- http://127.0.0.1:7216`
- `git diff --check`

The temporary 7216 server was stopped after the browser smoke.

## Scope guard

Checkpoint 1 deliberately does **not**:

- widen Tactical execution beyond the existing Laser consumer;
- change Beam target legality/arcs;
- implement Count/Autofire shot loops in BattleSession;
- implement Beam damage modifiers against live targets;
- implement projectile lifecycle or interception outcomes.

Those remain owned by Checkpoints 2-4 and Slice 17.7 exactly as frozen.

## Next

Checkpoint 2 - generic Beam execution + arcs + Count/Autofire:

- make the 13 frozen Beam identities executable under `TacticalSupport="beam"`;
- remove Laser-only gameplay authority while preserving legacy compatibility;
- enforce frozen firing-arc target legality;
- resolve physical Count outer loop and Autofire inner three-shot ordering;
- preserve zero RNG/readiness/events on rejected actions;
- keep current Laser regressions green.
## Checkpoint 2 - generic Beam execution + arcs + Count/Autofire

Checkpoint 2 replaces the Laser-only Tactical fire predicate with the frozen executable Beam consumer while preserving the legacy Laser snapshot path.

Implemented:

- all 13 frozen Beam identities now project with TacticalSupport="beam";
- internal/battle consumes the shared internal/beam configured derivation rather than duplicating mount/modifier facts;
- server-side target projection and command preparation enforce the persisted Beam firing arc before RNG/readiness/event mutation;
- the five Slice-17.2A arcs use the frozen original sectors: Forward 120 degrees, Forward Extended 240 degrees, Back Extended 240 degrees, Rear 120 degrees, and unrestricted 360;
- pre-17.2A in-memory Laser fixtures with an empty arc remain compatible, while newly projected persisted Beam rows carry canonical non-empty arcs;
- Heavy / Point Defense effective range-index transformation and configured static To-Hit facts are consumed by the hit path;
- doubled signed range To-Hit and Always-Hit participate in the generic hit path where present;
- physical Count is the outer battery loop and Autofire is the inner three-shot loop;
- flattened shot order is deterministic (unit1 s1/s2/s3, then unit2 ...);
- one accepted battery command consumes readiness once after the entire battery;
- rejected out-of-arc commands preserve State/RNG/readiness/events.

Target-side Beam damage effects, exact configured damage scaling against live targets, marine state/loss, and broader Shield/Armor/Structure effect hooks remain Checkpoint 3.

### Checkpoint 2 regression evidence

Passed on the Checkpoint-2 working tree:

- go test ./internal/beam ./internal/battle ./internal/game -count=1
- focused Slice-17.6 arc-reject and Count=2 x Autofire=3 BattleSession regressions
- updated live All-Tech game/session regressions proving Beam slots execute while deferred Ordnance remains unavailable
- go test ./... -count=1
- go vet ./...
- npm --prefix web run build
- git diff --check

The authoritative ruleset fingerprint changed only because the 13 frozen Beam rows moved from baseline / metadata_only to beam; the reviewed test fingerprint was updated accordingly.

## Checkpoint 2 close result

Gate 3 checkpoints 1-2 are complete. Checkpoint 3 is released for Beam target effects, exact damage application and minimal Marine state. Projectile lifecycle remains Slice 17.7 and active combat-Special mechanics remain Slice 17.8.

## Checkpoint 3 - Beam effects + minimal Marine state

Checkpoint 3 moves the frozen Gate-2 Beam damage/effect contract into the authoritative Tactical path without pulling reactive projectile Point Defense or broader combat-Special behavior forward.

Implemented:

- shared `internal/beam` runtime damage derivation now applies the frozen percentage order, original nearest-integer scaling, min/max clamps and final hit-damage interpolation;
- Heavy and Point Defense mounts consume their transformed range profile and live ship-target damage percentages; Point Defense remains a normal ship-target Beam mount here;
- selected Armor Piercing, Continuous, No Range Dissipation, Shield Piercing, Autofire and Enveloping facts resolve through the shared configured Beam identity;
- frozen innate effects resolve through canonical runtime facts: natural Shield Piercing, doubled signed range To-Hit, Armor-first bypass, monster/Antaran quarter-damage dependency hook, extra Structure damage, marine kill, Enveloping, Always Hit and Death Ray natural Continuous;
- Shield, Armor and Structure damage application accepts explicit effect options while preserving the established aggregate damage model;
- Enveloping applies a fresh copy of the same hit damage to shield facings 0, 1, 2 and 3;
- `MarineMax` is persisted in Tactical ship specs and `MarineCurrent` is server-owned mutable Tactical state/view;
- original hull MarineMax values are frozen as Frigate 5, Destroyer 8, Cruiser 12, Battleship 20, Titan 30 and Doom Star 50;
- passive `troop_pods` doubles initial MarineMax only; boarding/transport actions remain Slice 17.8;
- Neutron Blaster and Death Ray apply `postShieldDamage / 5` marine loss with integer truncation and saturation at zero;
- deferred Slice-17.8 dependencies remain explicit neutral hooks when their later consumer is absent.

### Checkpoint 3 deterministic evidence

The frozen A-D calculator vectors are regression-covered:

- A: Laser Standard 1-4 at 90% remains 1-4 and resolves 4 at the frozen roll;
- B: Particle Beam Heavy uses geometric index 4 -> Heavy index 2, 140%, scaled 14-42 and damage 31 at the frozen roll;
- C: Mauler Heavy consumes a hit RNG draw, overwrites to Always Hit, doubles signed range To-Hit and resolves 140 damage;
- D: Laser Point Defense ship target uses geometric index 2 -> PD index 4, 20%, scaled 1-1 and damage 1.

Additional Tactical regressions cover:

- Heavy and Point Defense live range/damage profiles;
- selected No Range Dissipation and Shield Piercing;
- innate Ion Pulse Armor-first bypass with neutral monster/Antaran hook;
- Fusion doubled signed range To-Hit;
- Mauler low raw-roll consumption followed by Always Hit;
- Particle natural Shield Piercing;
- Graviton extra Structure damage;
- Plasma Enveloping four-facing applications;
- Neutron and Death Ray marine loss;
- Troop Pods passive MarineMax x2 projection.

### Checkpoint 3 regression evidence

Passed on the Checkpoint-3 working tree:

- `go test ./internal/beam ./internal/battle ./internal/game ./internal/ruleset ./internal/session -count=1`;
- `go test ./... -count=1`;
- `go vet ./...`;
- `npm --prefix web run build`;
- `git diff --check`.

## Checkpoint 3 close result

Gate 3 checkpoints 1-3 are complete. Checkpoint 4 is released for reactive Beam Point Defense handoff plus frozen event/VFX identity. Projectile lifecycle remains Slice 17.7, and broader active combat-Special mechanics remain Slice 17.8.

## Checkpoint 4 - reactive PD handoff + events/VFX identity

Checkpoint 4 exposes the frozen Beam Point Defense reaction boundary without pulling Slice 17.7 projectile lifecycle into 17.6.

Implemented:

- `BattleSession.ResolveReactiveBeamPointDefense(...)` is a server-owned internal reaction surface rather than a new player protocol command;
- the caller supplies the incoming actor coordinates plus the defense input that Slice 17.7 will derive from projectile semantics;
- only Point Defense Beam mounts are accepted; normal Standard/Heavy Beam mounts reject without state/event mutation;
- the reaction uses the configured Beam identity, transformed PD range, frozen To-Hit/RNG rules, Count and Autofire ordering;
- reactive shots consume authoritative Beam hit RNG and emit Beam shot/volley facts, but do not mutate projectile state or invent interception semantics;
- all normal weapon readiness bytes on the reacting ship are restored exactly after the reaction, including a pre-existing `ready=false` state;
- reactive execution does not advance the player Tactical command sequence;
- normal `beam_fired` events now include `weapon_id`, `mount_type`, canonical `modifier_ids`, physical weapon index/count, burst index/count, flattened shot index/count and the frozen hit/damage facts;
- `battle_damage_applied` carries the same shot/weapon identity plus Enveloping `application_index` / `application_count` and facing result identity;
- `beam_volley_resolved` carries weapon identity and physical/burst grouping facts without replacing per-shot authority;
- browser Beam presentation remains event-only and reconstructs grouped Autofire hit/miss state from the authoritative per-shot `beam_fired` events instead of collapsing a mixed volley into one invented hit state.

### Checkpoint 4 deterministic evidence

Regression coverage verifies:

- Count=2 + Autofire emits six authoritative shots in physical-weapon outer / burst inner order;
- each shot carries physical index/count, burst index/count and flattened index/count;
- Enveloping emits four correlated damage applications with application indexes 1..4 and shield facings 0..3;
- reactive PD against actor coordinates consumes Beam hit RNG while preserving normal readiness and command sequence;
- a non-PD Beam reaction request rejects atomically.

### Checkpoint 4 regression evidence

Passed on the Checkpoint-4 working tree:

- `go test ./internal/battle -run TestSlice176Checkpoint4 -count=1`;
- `go test ./internal/beam ./internal/battle ./internal/game ./internal/ruleset ./internal/session -count=1`;
- `npm --prefix web run build`;
- `git diff --check`.

## Checkpoint 4 close result

Gate 3 checkpoints 1-4 are complete. Checkpoint 5 is released for active Tactical persistence/replay, Observer/public projection and end-to-end Selector -> Save -> Build -> BattleSession -> Tactical evidence. Projectile creation, flight, interception timing/eligibility, durability/evasion, projectile mutation and impact remain Slice 17.7.
