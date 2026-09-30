# Slice 17.5 Gate 4 - Acceptance Close

Date: 2026-09-30
Status: **accepted and closed; Gates 1-4 complete**

Parent: Slice 17 - Military Ship Design + Tactical Combat
Child: Slice 17.5 - Tactical Combat Foundation / Ship Combatant Breadth

## 1. Acceptance basis

Gate 4 closes Slice 17.5 after automated QA plus live user review on the canonical 7171 All-Tech path.

The live review proved the intended product boundary end to end:

- a late-game custom warship with later Hull/Drive/Computer/Armor/Shield/Fuel equipment can enter Tactical instead of being rejected by the old Slice-15.5 baseline gate;
- metadata-only later-family weapons remain visible as authoritative per-slot actions with `tactical_consumer_unavailable` instead of blocking the whole encounter;
- the user successfully reached the Tactical UI with a broad All-Tech combatant;
- the live review exposed one real admission regression: ordnance modifier `fast` was valid in Core/Designer but rejected by Battle tactical snapshot validation;
- commit `56dee99` fixed that regression by aligning Battle modifier canonical ordering with the persisted Core modifier set, without implementing ordnance behavior early;
- the user then identified that built-Ship Specials were missing from Galaxy fleet ship details; commit `f29f0c7` added display-only Specials to own-local, own-transit and foreign concrete-Ship detail popovers.

These live fixes tighten the accepted Slice-17.5 foundation without widening into Beam, ordnance or combat-Special mechanics owned by Slices 17.6-17.8.

## 2. Accepted Tactical combatant foundation

### 2.1 Broad combatant admission

The accepted foundation admits the frozen Slice-17.1 matrix without hidden baseline substitution:

- Frigate, Destroyer, Cruiser, Battleship, Titan and Doom Star;
- Titanium through Xentronium Armor;
- supported Computers and their frozen Beam-Offense rows;
- no Shield plus Class I/III/V/VII/X four-facing Shield state;
- Nuclear through Interphased Drive combat-speed rows;
- Fuel Cell identity as Tactical-admission-neutral metadata.

The old Frigate/Electronic/Titanium/no-Shield/Standard-Fuel guard no longer blocks otherwise valid richer designs.

### 2.2 Shared server authority

BattleSession remains the sole Tactical gameplay authority.

Accepted common infrastructure includes:

- immutable Tactical ship foundation snapshots;
- server-owned combat speed, Beam Offense and Beam Defense foundation values;
- four-facing Shield pools, block, shared Shield -> Armor -> Structure damage and deterministic recharge;
- deterministic server RNG sequencing;
- shared damage/event envelopes;
- slot-based weapon readiness/action projection;
- authoritative target lists for executable actions;
- explicit per-slot unsupported-family projection.

React consumes projected state/actions/events and does not invent hit, damage, readiness or target legality.

### 2.3 Generic weapon-family handoff

The existing Laser vertical slice remains executable and regression-green.

Later-family weapons are admitted as metadata and visible actions while their owner slices remain pending:

- complete Beam behavior -> Slice 17.6;
- missiles/torpedoes/bombs -> Slice 17.7;
- Fighters/Bombers/combat Specials -> Slice 17.8.

Unsupported slots use `tactical_consumer_unavailable` and do not consume RNG or readiness.

## 3. Gate 4 live-review fixes

### 3.1 Metadata-only modifier admission

The live All-Tech Battleship used Mauler Device, a missile with modifier `fast`, Fusion Beam and Battle Pods.

Before the fix, strategic resolution stopped before encounter commit because `internal/battle` still used a Beam-only modifier rank and rejected valid metadata-only ordnance modifier `fast`.

Commit:

- `56dee99 fix(tactical): admit metadata-only weapon modifiers`

Regression coverage now proves both direct BattleSession admission and the full strategic `ResolveStrategic -> Encounter -> Tactical` path for the representative live loadout.

### 3.2 Galaxy built-Ship Specials visibility

Built-Ship `ShipSpecialMount` snapshots were already persisted, but the Galaxy Fleet/Ship `?` detail popovers rendered Weapons only.

Commit:

- `f29f0c7 ui(galaxy): show ship specials in fleet details`

The three concrete-Ship detail surfaces now show persisted Specials directly below Weapons:

- own local fleet ship;
- own transit ship;
- foreign concrete ship.

This is presentation completeness only; combat-Special behavior remains owned by Slice 17.8.

## 4. Implementation evidence

Gate evidence:

- Gate 1: `docs/research/SLICE_17_5_GATE1_TACTICAL_FOUNDATION_AUDIT_2026-09-28.md`
- Gate 2: `docs/research/SLICE_17_5_GATE2_TACTICAL_FOUNDATION_CONTRACT_FREEZE_2026-09-29.md`
- Gate 3 plan: `docs/research/SLICE_17_5_GATE3_IMPLEMENTATION_CHECKPOINT_PLAN_2026-09-29.md`
- Gate 3 evidence: `docs/research/SLICE_17_5_GATE3_IMPLEMENTATION_EVIDENCE_2026-09-29.md`

Gate-3 implementation checkpoints:

- `8715ae3 feat(slice): complete 17.5 gate 3 checkpoint 1`
- `5897755 feat(slice): complete 17.5 gate 3 checkpoint 2`
- `a55490c feat(slice): complete 17.5 gate 3 checkpoint 3`
- `bd70189 feat(slice): complete 17.5 gate 3 checkpoint 4`
- `80fbdd7 test(slice): complete 17.5 gate 3 checkpoint 5`

Gate-4 live-review fixes:

- `56dee99 fix(tactical): admit metadata-only weapon modifiers`
- `f29f0c7 ui(galaxy): show ship specials in fleet details`

## 5. Final verification

Gate-4 close verification passed on final HEAD `f29f0c7` before close documentation:

- focused Laser Tactical regression - **PASS**
- focused `TestSlice175` game/battle/session coverage - **PASS**
- Tactical replay + Battle/Session LiveSnapshot roundtrip/continuation - **PASS**
- `go test ./... -count=1` - **PASS**
- `go vet ./...` - **PASS**
- `npm --prefix web run build` - **PASS**, including UTF-8, TypeScript and Vite production build
- fresh `-reference-games -enable-persistence` server on `127.0.0.1:7216` - healthy
- `npm --prefix web run check:reference-lab:browser -- http://127.0.0.1:7216` - **PASS**
- `git diff --check` - required green again after close documentation

The temporary 7216 verification service was stopped after the browser smoke.

## 6. Gate 4 checklist

- [x] Every accepted 17.1 hull/mandatory-system representative enters Tactical with frozen durability/offense/defense/speed values.
- [x] The old Slice-15.5 Frigate/Electronic/Titanium/no-shield/standard-fuel rejection no longer blocks otherwise legal richer designs.
- [x] No hidden baseline substitution remains for Titan/Doom Star, later Armor, later Computers, Shields or Drives.
- [x] A ship carrying a not-yet-implemented weapon family can still enter Tactical; only that specific action is explicitly unavailable.
- [x] Existing Laser combat remains authoritative and regression-green.
- [x] Tactical UI consumes generic action/event data without adding client-side hit/damage formulas.
- [x] Generic event-driven presentation remains presentation-only.
- [x] Full Go/vet/web/browser/persistence/replay/reference-game QA passes.

## 7. Close result

Slice 17.5 is accepted and closed with Gates 1-4 complete.

The exit criterion is satisfied:

- Tactical is a truthful destination for the accepted broad Ship Designer combatant matrix;
- shared authority, RNG, damage, Shield, action and event plumbing is no longer tied to the old baseline fixture;
- later weapon-family slices can add their rules without reopening combatant admission;
- deferred family behavior remains explicit rather than silently simulated.

The Slice-17.5 `_OPEN_` marker is removed by this closeout.

Slice 17.6 - Tactical Beam Weapons & Beam Modifiers is the next prepared objective and remains **planned/not open** until explicitly started.
