# Slice 17.6 Gate 4 - Acceptance Close

Date: 2026-10-01
Status: **accepted and closed; Gates 1-4 complete**

Parent: Slice 17 - Military Ship Design + Tactical Combat
Child: Slice 17.6 - Tactical Beam Weapons & Beam Modifiers

## 1. Acceptance basis

Gate 4 closes Slice 17.6 against the frozen Gate-2 contract and completed Gate-3 implementation evidence.

This closeout is acceptance-only. No new Beam gameplay formula or ownership was introduced in Gate 4.

The final acceptance proves that:

- all 13 frozen Beam identities project as executable server-authoritative Tactical Beam actions;
- Standard, Heavy and Point Defense use the shared configured Beam derivation;
- frozen hit/range/damage/RNG/readiness order remains deterministic;
- supported selected modifiers and innate Beam effects resolve through BattleSession;
- exact percentage rounding and the frozen reference vectors remain regression-locked;
- immutable configured Beam snapshots survive Selector -> Save -> Build -> BattleSession -> Tactical;
- active Tactical persistence/replay preserves configured Beam identity, readiness, Marines, RNG and event sequence;
- reactive Beam Point Defense remains a bounded handoff for Slice 17.7 and does not consume normal activation readiness;
- browser presentation consumes authoritative Beam/damage events and does not calculate hit/damage gameplay locally;
- desktop/mobile Weapon Selector and multi-row Designer usage remain usable;
- deferred projectile lifecycle and broader combat-Special behavior remain owned by Slices 17.7 and 17.8.

## 2. Accepted Tactical Beam surface

### 2.1 Supported Beam identities

The accepted executable Beam family is:

1. Mass Driver
2. Gauss Cannon
3. Laser Cannon
4. Particle Beam
5. Fusion Beam
6. Ion Pulse Cannon
7. Graviton Beam
8. Neutron Blaster
9. Phasor
10. Disrupter Cannon
11. Death Ray
12. Plasma Cannon
13. Mauler Device

Gate-4 regression `TestSlice176Gate4AllSupportedBeamsProjectAndExecute` proves that every identity:

- survives the real ruleset -> strategic Ship snapshot -> Tactical projection path;
- is projected with `WeaponFamily="beam"` and `TacticalSupport="beam"`;
- receives an available server-projected `battle.fire_beam` action against a legal target;
- executes through `BattleSession`;
- emits authoritative `beam_fired` output.

### 2.2 Shared configured Beam derivation

The same server-side Beam derivation remains shared by Designer/Weapon Selector preview and Tactical resolution.

Accepted configured facts include:

- base versus configured min/max damage;
- Standard / Heavy / Point Defense range profile;
- selected modifier set;
- innate effect set;
- shots per physical weapon;
- static configured To-Hit adjustment.

No second React-side Beam formula is accepted.

### 2.3 Frozen hit / damage / RNG behavior

The Gate-2 ordering remains authoritative and regression-locked.

The frozen deterministic controls A-E pass:

- Vector A - Laser / Standard rounding;
- Vector B - Particle Beam / Heavy with natural Shield Piercing;
- Vector C - Mauler / Heavy with consumed hit RNG and Always Hit override;
- Vector D - Laser / Point Defense ship target;
- Vector E - Count 2 + Autofire six-shot ordering.

The accepted implementation preserves:

- transformed Heavy/Point Defense range;
- signed range To-Hit handling;
- exact nearest-integer damage percentage scaling;
- physical Count outer ordering;
- Autofire inner three-shot ordering;
- miss/hit RNG ordering;
- invalid/out-of-arc zero-mutation behavior;
- one accepted battery command consuming normal readiness once.

### 2.4 Supported effects

Accepted selected/innate effect consumers include the frozen Slice-17.6 matrix:

- Shield Piercing;
- Armor Piercing / first-layer bypass;
- No Range Dissipation;
- Continuous;
- Autofire;
- Enveloping;
- natural Beam-specific effect identities;
- Graviton extra Structure damage;
- Neutron Blaster / Death Ray marine loss;
- Ion Pulse armor/structure bypass with the deferred exception hook;
- Plasma four-facing application;
- Fusion doubled signed range To-Hit;
- Mauler Always Hit / doubled range dissipation behavior.

`MarineMax` / `MarineCurrent` and Troop Pods x2 Marine capacity remain part of the accepted Tactical snapshot/runtime contract.

### 2.5 Reactive Beam Point Defense boundary

Slice 17.6 accepts the Beam-side reactive Point Defense capability only.

The accepted handoff:

- fires against later projectile actor coordinates;
- uses the frozen Beam hit/damage rules;
- preserves normal activation readiness afterward;
- emits frozen Beam shot identity;
- does not create, move, mutate, destroy or impact a projectile actor.

Projectile lifecycle, interception eligibility/timing, projectile durability/evasion and impact remain Slice 17.7.

## 3. Persistence, replay and projection acceptance

Configured Beam Tactical state is accepted across the active live-snapshot boundary.

Regression coverage proves preservation of:

- immutable configured Beam spec;
- current weapon readiness;
- Shield / Armor / Structure state;
- `MarineCurrent`;
- Tactical RNG state;
- command sequence;
- event sequence and event payload identity.

Corrupt `MarineCurrent` outside `0..MarineMax` is rejected during restore.

Identical Tactical spec + RNG + continuation commands reproduce identical state/events after restore.

Observer/public projections are detached from authority. Participant views redact Tactical RNG and only project command authority to the active controlling seat.

## 4. End-to-end integration acceptance

The real session integration regression proves:

`Weapon Selector / design configuration -> Save -> Colony Build -> immutable built Ship -> Encounter -> BattleSession -> Tactical -> Beam fire`

The representative All-Tech configuration is:

- Phasor;
- Point Defense mount;
- Count 3;
- Shield Piercing;
- Autofire;
- 360-degree firing arc.

The configured identity survives unchanged from saved design through built Ship and Tactical spec and executes through `battle.fire_beam`.

The active configured-Beam GameSession also round-trips through the public live-snapshot boundary and continues deterministically.

## 5. Browser / presentation acceptance

`TacticalBattlefield` remains event-driven.

Beam animation and damage-layer presentation consume authoritative:

- `beam_fired`;
- `beam_volley_resolved`;
- `battle_damage_applied`;

including server-projected hit, damage, shot/group identity and layer results.

Gate-4 inspection found no `Math.random` path in the Tactical Beam presentation and no client-owned Beam hit/damage formula.

The Reference-Lab browser smoke passes both mobile and desktop Designer coverage, including:

- 390px/mobile non-overflow;
- desktop Designer layout;
- Weapon Selector opening/filter/configuration flow;
- compact stable installed-weapon rows;
- multiple weapon-row usability and icon controls;
- authoritative Space/PP/combat-rating display;
- ordinary-game/reference capability isolation.

## 6. Implementation evidence

Gate evidence:

- Gate 1: `docs/research/SLICE_17_6_GATE1_TACTICAL_BEAM_EFFECT_AUDIT_2026-09-30.md`
- Gate 2: `docs/research/SLICE_17_6_GATE2_TACTICAL_BEAM_CONTRACT_FREEZE_2026-09-30.md`
- Gate 3 checkpoint plan: `docs/research/SLICE_17_6_GATE3_IMPLEMENTATION_CHECKPOINT_PLAN_2026-09-30.md`
- Gate 3 evidence: `docs/research/SLICE_17_6_GATE3_IMPLEMENTATION_EVIDENCE_2026-10-01.md`

Gate-3 implementation commits:

- `0fa8b8f feat(slice): complete 17.6 gate 3 checkpoint 1`
- `28ee836 feat(slice): complete 17.6 gate 3 checkpoint 2`
- `7dc88e2 feat(slice): complete 17.6 gate 3 checkpoint 3`
- `3c0f58a feat(slice): complete 17.6 gate 3 checkpoint 4`
- `738cf23 feat(slice): complete 17.6 gate 3 checkpoint 5`

Gate 4 adds only the explicit all-13-Beam acceptance regression and close documentation.

## 7. Final verification

Gate-4 acceptance passed from clean Gate-3 HEAD `738cf23` plus the acceptance regression:

- `go test ./internal/game -run TestSlice176Gate4AllSupportedBeamsProjectAndExecute -count=1` - **PASS**
- focused Slice-17.6 Beam/Battle/Game/Session regressions - **PASS**
- shared configured Beam + Designer preview regressions - **PASS**
- frozen vectors A-E - **PASS**
- configured Beam persistence/replay/Observer regressions - **PASS**
- Selector -> Save -> Build -> BattleSession -> Tactical integration - **PASS**
- `go test ./... -count=1` - **PASS**
- `go vet ./...` - **PASS**
- `npm --prefix web run build` - **PASS**
- fresh Reference-Games/persistence server on `127.0.0.1:7217` - **healthy**
- `npm --prefix web run check:reference-lab:browser -- http://127.0.0.1:7217` - **PASS**
- browser smoke covers mobile/desktop Designer and ordinary-game isolation - **PASS**
- All-Tech HTTP live-snapshot GET -> PUT -> GET after browser mutation - **200 and byte-identical**
- `git diff --check` and staged diff check - required green before commit

The temporary `7217` acceptance service is stopped after final verification.

The canonical `7171` development server remains on the accepted code path with the backed-up live games restored; the existing All-Tech game remains preserved.

## 8. Gate 4 checklist

- [x] Every supported Beam can be selected in Tactical and resolves through server-projected legal actions.
- [x] Hit/miss, range and damage rolls match frozen reference vectors.
- [x] Heavy and Point Defense produce the correct configured preview and Tactical result.
- [x] Supported modifiers produce exact frozen Tactical effects and combinations.
- [x] Selector -> Save -> Build -> BattleSession -> Tactical preserves exact configured Beam semantics.
- [x] Base/configured damage and affected stats match between Selector and Combat Lab through the shared derivation.
- [x] Beam/impact/shield/damage animations correspond to authoritative events and remain presentation-only.
- [x] Desktop/mobile multi-weapon selection remains usable.
- [x] Full Go/vet/web/browser/persistence/replay/reference-game QA passes.

## 9. Close result

Slice 17.6 is accepted and closed with Gates 1-4 complete.

The exit criterion is satisfied:

- every Beam mount/modifier effect accepted into the frozen Slice-17.6 matrix has an authoritative Tactical consumer;
- Beam hit/range/damage/RNG/readiness resolution is server-owned;
- Weapon Selector and Tactical consume the same configured Beam rule source;
- immutable configured Beam semantics persist from Designer through built Ship into BattleSession;
- Tactical persistence/replay and projection boundaries remain deterministic and authority-safe;
- browser Beam presentation is driven by authoritative events rather than client gameplay formulas.

The Slice-17.6 `_OPEN_` marker is removed by this closeout.

Slice 17.7 - Tactical Missiles / Torpedoes / Bombs is the next prepared child and remains **planned/not open** until explicitly started.
