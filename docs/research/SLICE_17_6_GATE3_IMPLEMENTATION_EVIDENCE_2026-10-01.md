# Slice 17.6 Gate 3 - Implementation Evidence

Date: 2026-10-01
Status: **Checkpoints 1-2 complete; Checkpoint 3 ready**

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
