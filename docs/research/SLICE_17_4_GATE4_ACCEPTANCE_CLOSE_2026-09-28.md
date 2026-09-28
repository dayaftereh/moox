# Slice 17.4 Gate 4 - Acceptance Close

Date: 2026-09-28
Status: **accepted and closed; Gates 1-4 complete**

Parent: Slice 17 - Military Ship Design + Tactical Combat
Child: Slice 17.4 - Ship Specials and Fighter Systems

## 1. Acceptance basis

The user manually reviewed the implemented Specials/Fighter Ship Designer on the canonical 7171 All-Tech reference path and explicitly accepted the final Gate-4 result.

The accepted HMI refinement aligns Specials with the established weapon-selection grammar:
- installed Specials use the same compact card language as installed Weapons;
- the Special area exposes remaining/free slots rather than eight empty permanent selector rows;
- + Special uses the same add-card grammar as + Weapon;
- the selector exposes the available Special list, then a focused detail view with original description and authoritative design values;
- installed Specials expose ? information and delete controls;
- the add control disappears at 8/8 installed Special slots;
- oversized weapon/Special choices are visibly marked against current authoritative free Space while remaining inspectable;
- desktop and 390px/mobile layouts remain non-overlapping.

This acceptance validates the frozen Gate-2 contract and completed Gate-3 implementation rather than widening scope.

## 2. Accepted Slice-17.4 product surface

### 2.1 Authoritative Special design breadth

- 38 player-visible original Special choices are represented authoritatively.
- Regeneration remains locked from normal Designer selection.
- Special technology ownership, per-hull Space/PP, miniaturization, exclusion and shield-prerequisite rules remain server-owned.
- Eight unique Special slots are persisted in immutable Ship design snapshots.
- Duplicate, incompatible and unsupported Special choices are rejected by the authoritative preview/save path.

### 2.2 Frozen strategic/design effects

- Megafluxers expands effective hull design Space according to the frozen contract.
- Battle Pods retain their original negative Space contribution and bypass ordinary miniaturization.
- Extended Fuel Tanks applies the frozen range effect.
- Inertial Stabilizer / Inertial Nullifier connect only to the already-existing Tactical turning consumer.
- No other combat-facing Special effect is silently simulated in Slice 17.4.

### 2.3 Fighter/Bomber carrier representation

- Fighter Bays, Bomber Bays and Heavy Fighter Bays are present as supported design rows.
- Count represents bays; each bay carries four craft.
- Fighter rows do not expose Beam firing-arc/mount controls or ordnance Ammo controls.
- Space, PP, technology and miniaturization use the same authoritative quote path as the other supported Designer weapon families.
- Tactical fighter resolution remains metadata-only and explicitly handed to Slice 17.8.

### 2.4 Immutable persistence and Tactical handoff

Regression coverage proves that supported Specials and Fighter bays survive:
- preview / save canonicalization;
- design reload/revision;
- construction queue snapshot;
- built-Ship immutable snapshot;
- Tactical metadata projection.

Deferred combat effects remain explicit metadata rather than fake runtime behavior.

### 2.5 Accepted desktop/mobile HMI

The user accepted the final Specials HMI after iterative review.

Final accepted presentation includes:
- weapon-consistent installed-Special cards;
- remaining/free slot count;
- matching blue dashed + Special card;
- selector list plus focused information/detail pane;
- original Special technology descriptions;
- authoritative current Space/PP preview before install;
- ? information action on installed Specials;
- delete action on installed Specials;
- no add control once all eight slots are occupied;
- free-Space fit markers for weapon/Special choices;
- stable desktop and 390px/mobile layout.

The browser remains a presentation/client layer: legality, Space, PP, exclusions, prerequisites and persistence are server-authoritative.

## 3. Implementation evidence

Key implementation/refinement commits:
- 79ab78f - feat(slice): implement 17.4 specials and fighters
- 910b9e8 - fix(ui): align 17.4 special selection flow
- 0884ee2 - fix(ui): match special slot layout to weapons
- 4b5f159 - fix(ui): improve special editing feedback
- c33cca6 - fix(ui): mark ship choices that exceed free space

Gate evidence:
- Gate 1: docs/research/SLICE_17_4_GATE1_SPECIALS_FIGHTERS_AUDIT_2026-09-26.md
- Gate 2: docs/research/SLICE_17_4_GATE2_SPECIALS_FIGHTERS_CONTRACT_FREEZE_2026-09-27.md
- Gate 3: docs/research/SLICE_17_4_GATE3_IMPLEMENTATION_2026-09-28.md

## 4. Final verification

Gate-4 close verification passed on the accepted final state:

- npm --prefix web run build - PASS
- go test ./... -count=1 - PASS
- go vet ./... - PASS
- npm --prefix web run check:reference-lab:browser - PASS
- git diff --check - PASS before close documentation

The canonical review server remains the 0.0.0.0:7171 Reference-Games path used for the manual acceptance. The browser smoke covers desktop and 390px/mobile layout.

## 5. Explicit downstream boundary

Slice 17.4 does **not** claim complete Tactical Fighter/Bomber/Combat-Special behavior.

Planned Slice 17.8 remains the owner of:
- Fighter/Bomber/Heavy Fighter launch and runtime behavior;
- fighter movement, targeting, weapons/bombs and return/recovery;
- tactical combat-special consumers without an existing 17.4 runtime owner;
- corresponding Tactical actions/events/UI;
- any additional combat-only interactions frozen as deferred metadata.

Slice 17.5 remains the next planned child in dependency order and is **not opened automatically** by this closeout.

## 6. Gate-4 checklist

- [x] Every visible supported Special has the promised effect/support classification.
- [x] No unsupported Tactical effect is silently simulated or ignored.
- [x] Compatibility/Space/Cost/technology rules pass deterministic matrices.
- [x] Late-game Specials/Fighter designs are practical to review through the reference lab.
- [x] Construction/persistence/revision remain exact.
- [x] Desktop/mobile Designer HMI is accepted and consistent with the Weapon workflow.
- [x] Full tests/vet/web/browser checks and git diff --check pass.

## 7. Close result

Slice 17.4 is accepted and closed with Gates 1-4 complete.

The exit criterion is satisfied:
- supported Specials/Fighter systems are authoritative design objects;
- accepted strategic and existing Tactical consumers are explicit and tested;
- immutable design/construction/Tactical metadata is preserved;
- deferred combat consumers remain owned by Slice 17.8;
- the final Specials/Fighter Designer HMI is manually accepted.

The Slice-17.4 _OPEN_ marker is removed by this closeout.

No later Slice-17 child is opened automatically.
