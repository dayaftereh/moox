# Slice 17.4 Gate 3 - Ship Specials and Fighter Systems Implementation

Date: 2026-09-28
Status: **Implemented; automated Gate 3 verification green; Gate 4 review remains open.**

## Implemented scope

Gate 3 implements the Gate-2-frozen Ship Specials and Fighter Systems contract without widening into deferred Slice 17.8 combat behavior.

### Authoritative rules and design snapshots

- Added ship_specials.json as an authoritative ruleset input and included it in the rules identity.
- Added the 38 player-visible original Special choices, their technology ownership, per-hull Space/PP economics, miniaturization class, support classification, exclusions and shield prerequisite metadata.
- Regeneration remains excluded from normal Designer choice.
- Added immutable ShipSpecialMount snapshots to saved military designs and preserved them through save, construction queue and built Ship materialization.
- Special slots are canonicalized as unique ascending slots in 0..7; duplicate Special IDs, conflicting exclusions and unsupported choices are rejected server-side.

### Frozen strategic effects

- Megafluxers expands effective hull design Space to 125% when its technology is known.
- Battle Pods retain their original negative Space contribution and are not clamped away early.
- Extended Fuel Tanks applies the frozen 1.5x range rule.
- Special Space/PP and miniaturization remain server-authoritative and flow through the normal preview/save quote path.

### Fighter systems

- Added Fighter Bays, Bomber Bays and Heavy Fighter Bays as fighter weapon rows.
- Quantity counts bays; each bay exposes craft_per_unit = 4.
- Fighters do not expose Beam mount modifiers, ordnance ammo, or firing-arc controls.
- Fighter rows use the same server-authoritative technology, miniaturization, Space and PP pipeline as the other supported Designer weapon families.
- Deferred fighter tactical combat behavior remains metadata-only for Slice 17.8.

### Tactical handoff

- Tactical Ship snapshots now carry immutable Special metadata.
- Only the already-supported turning modes are consumed in this slice:
  - Inertial Stabilizer -> TacticalTurningInertialStabilizer
  - Inertial Nullifier -> TacticalTurningInertialNullifier
- Other combat-facing Specials remain explicit metadata/strategic-only handoffs and are not silently simulated.

### Ship Designer HMI

- Added an 8-slot Special Systems panel.
- Catalog/preview data supplies availability, Space, PP, support classification and locks; the browser does not reimplement gameplay formulas.
- Added German/English Special presentation labels.
- Added localized Fighter/Bomber/Heavy Fighter rows and visible 4 craft per bay / 4 Jager pro Bucht metadata.
- Special state survives live preview, save/reload and weapon-editor preview flows.

## Regression coverage

Dedicated Slice-17.4 tests cover:

- complete Special catalog shape;
- Battle Pods / Megafluxers / Extended Fuel Tanks golden vectors;
- exclusion, shield-prerequisite and unsupported-Regeneration rejection;
- fighter quote semantics and absence of ordnance/Beam controls;
- Tactical Special metadata plus inertial turning-mode projection;
- save payload canonicalization;
- save -> construction queue -> built Ship preservation for Specials and Fighter bays.

Expected baselines were updated only after these tests were established:
- All-Tech effective Cruiser hull Space reflects Megafluxers;
- Designer weapon catalog grows from 26 to 29 rows;
- ruleset identity changes because ship_specials.json is authoritative.

## Final Gate 3 verification

Fresh finalization session: ses-20260928T070302-7965c8131e2a

- npm --prefix web run build - PASS
- go test ./... -count=1 - PASS (completed twice during the final run due tool transport replay)
- go vet ./... - PASS
- npm --prefix web run check:reference-lab:browser - PASS
- git diff --check - PASS

Direct All-Tech browser verification on the fresh -reference-games server also confirmed:

- 8 Special slots;
- Empty + 38 original Special choices per slot;
- Battle Pods visibly use the authoritative negative-Space quote;
- 29 weapon selector rows;
- Fighter Bays, Bomber Bays and Heavy Fighter Bays are all visible;
- all three Fighter rows show 4 craft per bay;
- no horizontal overflow in the direct review, while the automated browser smoke covers the 390 px mobile layout.

## Gate 4 boundary

Gate 3 implementation is ready for review. Slice 17.4 remains open.

Gate 4 must not invent deferred Special/Fighter tactical behavior. Slice 17.8 remains the owner of those combat consumers.
