# Slice 17.6 Gate 3 - Implementation Evidence

Date: 2026-10-01
Status: **Checkpoint 1 complete; Checkpoint 2 ready**

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
