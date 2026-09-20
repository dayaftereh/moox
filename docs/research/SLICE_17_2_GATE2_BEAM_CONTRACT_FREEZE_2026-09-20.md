# Slice 17.2 Gate 2 - Beam Weapons and Beam Mount Modifiers Contract Freeze

Date: 2026-09-20  
Result: **FROZEN**  
Input evidence: `docs/research/SLICE_17_2_GATE1_BEAM_WEAPONS_AUDIT_2026-09-20.md`

## 1. Contract goal

Gate 2 freezes the authoritative data and HMI contract that Gate 3 must implement.

The contract deliberately separates four concerns:

1. original Beam identity/rules data;
2. immutable design/built-Ship weapon snapshots;
3. server-authoritative Designer preview/validation;
4. compact selection/configuration HMI.

No Beam identity, modifier compatibility or cost/Space formula may be invented in TypeScript or inferred from translated text.

## 2. Supported Beam matrix

Slice 17.2 supports exactly the 13 Gate-1 Beam identities:

1. `mass_driver`
2. `gauss_cannon`
3. `laser_cannon`
4. `particle_beam`
5. `fusion_beam`
6. `ion_pulse_cannon`
7. `graviton_beam`
8. `neutron_blaster`
9. `phasor`
10. `disrupter_cannon`
11. `death_ray`
12. `plasma_cannon`
13. `mauler_device`

Technology IDs/fields, base Space, base PP, min/max damage and exact executable modifier masks are frozen to the Gate-1 matrix.

The existing spelling `disrupter_cannon` remains canonical for compatibility.

No Missile/Torpedo/Bomb/Fighter weapon is added in 17.2.

## 3. Canonical weapon configuration model

The original design format distinguishes regular mount selection from special Beam modifiers. MOOX keeps that distinction explicitly instead of exposing one opaque integer mask as the public design model.

### 3.1 Mount type

Canonical mount type enum:

- `standard`
- `heavy`
- `point_defense`

Only one mount type can exist on one weapon row.

Original mask mapping:

- standard -> no regular-mount bit;
- heavy -> `0x0002`;
- point_defense -> `0x0004`.

### 3.2 Special modifier IDs

Canonical special modifier IDs:

- `armor_piercing` -> `0x0008`
- `continuous` -> `0x0010`
- `no_range_dissipation` -> `0x0020`
- `shield_piercing` -> `0x0040`
- `autofire` -> `0x0080`
- `enveloping` -> `0x0100`

A weapon row may have zero or more legal special modifiers.

Canonical ordering in snapshots and API payloads is exactly the order above. Duplicate modifier IDs are invalid.

### 3.3 ShipWeaponMount snapshot shape

Gate 3 must extend the authoritative snapshot conceptually to:

```text
ShipWeaponMount {
    Slot
    WeaponID
    Count
    MountType
    ModifierIDs
}
```

Required semantics:

- `Slot`: integer 0..7;
- `WeaponID`: one frozen Beam ID in 17.2;
- `Count`: positive physical weapon count;
- `MountType`: one canonical mount type;
- `ModifierIDs`: sorted, unique canonical special-modifier IDs.

Backward compatibility:

- legacy/missing/empty mount type decodes as `standard`;
- legacy/missing/null modifier list decodes as no modifiers;
- canonical newly saved snapshots normalize these values before validation/persistence.

The immutable built-Ship and construction snapshots store the configuration directly. They do not re-read current research or current design rows later.

### 3.4 No packed-ID hacks

Forbidden representations:

- `laser_cannon_heavy`;
- `phasor_autofire`;
- comma-delimited modifiers inside `WeaponID`;
- modifier state held only in browser/UI state;
- reconstruction from translated labels.

An original-compatible bit mask may be computed internally for validation/golden tests, but it is derived from the canonical mount/modifier fields and is not a second mutable source of truth.

## 4. Slot/count contract

The accepted Block-A contract remains:

- exactly eight logical weapon rows, slots `0..7`;
- at most eight mounted rows;
- slots unique and canonical ascending in persisted snapshots;
- count must be >0;
- one row is one independently configured/firing group;
- two rows may use the same WeaponID with different mount/modifier configurations;
- changing quantity on one row never merges it implicitly with another row.

Gate 3 may offer a UI convenience to add to an existing identical configuration, but any such merge is explicit deterministic UI behavior before save, not a rules-layer assumption.

## 5. Technology and modifier legality

A weapon row is legal only when all of the following pass:

1. weapon ID exists in the frozen Beam matrix;
2. the empire knows the Beam technology;
3. slot/count contract passes;
4. selected mount type is allowed by that weapon's executable mask;
5. every special modifier is allowed by that weapon's executable mask;
6. every special modifier meets its required Beam miniaturization level;
7. no duplicate modifier ID exists;
8. resulting total design Space fits the selected hull.

### 5.1 Modifier refinement requirements

Frozen requirements:

| Choice | Required mini level |
| --- | ---: |
| Standard | 0 |
| Heavy | 0 |
| Point Defense | 0 |
| Armor Piercing | 1 |
| Continuous | 1 |
| No Range Dissipation | 1 |
| Shield Piercing | 1 |
| Autofire | 2 |
| Enveloping | 2 |

A modifier can be executable-mask-compatible yet temporarily locked because the empire has not reached the required miniaturization/refinement level.

### 5.2 Stable rejection reasons

Gate 3 must expose deterministic machine-readable reasons. Exact constants are frozen conceptually as:

- `weapon_unknown`
- `weapon_technology_required`
- `weapon_slot_out_of_range`
- `weapon_slot_duplicate`
- `weapon_count_invalid`
- `weapon_mount_not_allowed`
- `weapon_modifier_unknown`
- `weapon_modifier_not_allowed`
- `weapon_modifier_refinement_required`
- `weapon_modifier_duplicate`
- existing global `space_exceeded`

UI text is translated from these reasons; rules code never parses translated messages.

## 6. Exact Space/Cost pipeline

All frozen Beam base Space/PP values are positive.

For one physical weapon, let:

- `B` = original base Space or PP;
- `Fmount` = mount factor;
- `Dspecial` = sum of selected special-modifier percentage deltas;
- `M` = current Beam miniaturization level.

### 6.1 Mount stage

Mount factors:

- standard: 100%;
- heavy: 200%;
- point defense: 50%.

Original executable behavior freezes:

`after_mount = floor(B * Fmount / 100)`

Heavy and Point Defense are mutually exclusive by the mount enum.

### 6.2 Special-modifier stage

Frozen special deltas for both Space and PP:

- Armor Piercing: +50%;
- Continuous: +50%;
- No Range Dissipation: +25%;
- Shield Piercing: +50%;
- Autofire: +50%;
- Enveloping: +100%.

The original helper initializes a zero bonus and **adds** each selected special modifier's table value. Therefore special modifiers are additive within the special group, then applied as one factor:

`after_special = floor(after_mount * (100 + Dspecial) / 100)`

This is intentionally not repeated per-modifier compounding.

### 6.3 Miniaturization stage

Cost factors:

- M0 100%;
- M1 75%;
- M2 55%;
- M3 40%;
- M4 30%;
- M5+ 25%.

`unit_cost = max(1, floor(after_special_cost * cost_factor / 100))`

Beam Space is original miniaturization class 1:

- M0 100%;
- M1 80%;
- M2 70%;
- M3 60%;
- M4 50%;
- M5+ 40%.

`unit_space = max(1, floor((after_special_space * space_factor_per_thousand + 500) / 1000))`

The Space expression is positive nearest-integer rounding with half boundaries upward.

### 6.4 Count stage

For mount count `N`:

- `row_space = unit_space * N`;
- `row_cost = unit_cost * N`.

Design Space/PP sums row totals after mandatory-system values.

Miniaturization is evaluated from the empire's current research when editing/saving a design. Queue-time and built-Ship snapshots then freeze the resulting authoritative design values according to the already accepted Slice-17.1 rules.

## 7. Server-authoritative Designer contract

The current 17.1 browser calculates Laser Space/Cost locally. Gate 3 must remove that Beam-specific client authority.

### 7.1 Catalog projection

The server Designer catalog must expose each Beam with at least:

- canonical ID and translation key;
- technology ID and field ID;
- technology-known / available state;
- lock reason;
- base Space;
- base PP;
- min/max damage;
- current miniaturization level;
- current standard-mount effective unit Space/PP;
- legal mount choices;
- legal special modifier choices;
- per-choice required mini level and available/locked state;
- modifier Space/PP deltas;
- innate original-effect tags;
- Tactical support/defer metadata.

Original executable bit values may also be exposed for diagnostics/reference tests, but HMI logic uses canonical IDs.

### 7.2 Draft preview/quote

Gate 3 must provide one backend-authoritative draft preview path using the same compiler/validator as Save.

Conceptual input:

```text
MilitaryDesignDraft {
    hull_id
    strategic_picture_id
    computer_id
    shield_id
    weapons[]
}
```

Conceptual response includes:

- compiled preview `ShipDesignSpec`;
- remaining Space;
- production/base PP;
- per-row unit Space/PP and row totals;
- save-available boolean;
- global rejection reason if invalid;
- per-weapon-row legality/lock reason where useful.

The final Save command and preview must share the same rules helpers. A combination that previews legal may not later fail because TypeScript and Go used different formulas.

### 7.3 Client rule boundary

TypeScript may:

- maintain draft form state;
- search/filter/sort the server-projected catalog;
- display server-projected modifier availability;
- request previews;
- perform presentational comparisons.

TypeScript must not independently implement:

- Beam miniaturization;
- modifier compatibility;
- refinement unlocks;
- mount/special percentage arithmetic;
- authoritative Space/PP legality.

## 8. Weapon-selector HMI contract

The main Ship Designer remains compact.

### 8.1 Entry point

The existing `+` action in the weapon area opens a dedicated **Weapon Selector** dialog/sheet.

Behavior:

- Desktop: wide modal/dialog optimized for scanning rows/cards;
- Mobile: near-fullscreen dialog/bottom sheet with sticky search/filter controls;
- selector does not increase the height of the main Designer;
- closing/canceling selector leaves the draft unchanged.

Slice 17.2 selector category is Beam only. The component may be architected for later categories, but Missile/Torpedo/Bomb tabs are not shown as fake/empty 17.2 content.

### 8.2 Frozen search/filter controls

Gate 3 implements:

- free-text search by localized weapon name;
- availability filter: `All | Available | Locked`;
- modifier-support filter using the eight frozen choices;
- sort: `Technology | Name | Damage | Space | PP`;
- ascending/descending where meaningful.

A numeric Range filter is **not** part of the 17.2 freeze because Gate 1 did not freeze a generic numeric range field. Innate no-range/doubled-range behavior appears as original-effect tags/information instead.

### 8.3 Selector row/card information

Primary information:

- weapon name;
- damage, including fixed versus min-max;
- current standard-mount unit Space;
- current standard-mount unit PP;
- availability/technology lock;
- concise innate-effect tags where relevant.

Secondary/details:

- technology field/tier;
- base Space/PP;
- current mini level/reduction;
- supported modifier tags;
- Tactical support/defer note.

Locked weapons remain visible when `All` or `Locked` is selected and cannot be added.

### 8.4 Add behavior

Selecting **Add** on a legal weapon:

- creates `count=1`;
- uses `standard` mount;
- uses no special modifiers;
- chooses the first free slot in `0..7`;
- requests a server draft preview;
- commits to the local draft only if the preview is legal.

If all eight rows are occupied, Add is disabled with a stable reason.

Insufficient Space is determined by server preview rather than a TypeScript multiplication.

## 9. Installed-weapon configuration dialog

The existing compact installed-weapon list remains fixed-height/scrollable.

Clicking a weapon row opens a **Weapon Configuration** dialog.

Frozen controls:

- quantity;
- mount type: Standard / Heavy / Point Defense, showing only executable-compatible choices;
- special modifiers, each as an independent toggle/checkbox;
- locked compatible modifiers remain visible with their required mini level;
- incompatible modifiers are not selectable and may be omitted or shown as unavailable in details;
- live server-authoritative unit/row Space and PP preview;
- Remove action;
- Cancel;
- Apply.

The main compact row continues to show only essential information. Gate 3 must not turn it back into a large always-expanded modifier card.

The inline +/- quantity controls may remain as a convenience; they use the same server preview path.

## 10. Future Special Systems boundary

Special Systems are not weapon modifiers.

Slice 17.4 may reuse the same high-level interaction grammar:

- compact installed list;
- `+` opens selector;
- click installed item opens configuration/details.

But 17.2 must not add Battle Pods, Heavy Armor, Reinforced Hull, Fighter Bays or other Special Systems to the Beam selector or Beam snapshot structure.

## 11. Tactical support/defer contract

Slice 17.2 freezes **design/snapshot support for all 13 Beams and all legal frozen modifier combinations**.

Tactical consumption is feature-scoped.

### 11.1 Supported base data path

Battle projection may consume for every frozen Beam:

- WeaponID;
- slot/count;
- min/max damage;
- mount type;
- special modifier IDs;
- innate effect IDs/support metadata.

This prevents later translation hacks.

The existing generic Beam damage path can consume min/max damage. That does not by itself claim every original Beam effect is simulated.

### 11.2 Canonical full baseline

The existing unmodified standard Laser remains the canonical fully regression-covered Tactical baseline.

### 11.3 Explicitly deferred original effects

Until a bounded Tactical consumer is separately implemented and regression-covered, these remain explicit deferred metadata:

- Heavy Mount Tactical range/damage behavior;
- Point Defense Tactical hit/range behavior;
- Armor Piercing;
- Continuous;
- No Range Dissipation modifier effect;
- Shield Piercing modifier effect;
- Autofire;
- Enveloping;
- innate Mass Driver/Gauss/Disrupter no-range-dissipation behavior;
- Particle Beam natural shield piercing;
- Ion Pulse armor/structure bypass and immunity exceptions;
- Graviton extra structural damage;
- Neutron/Death Ray marine-kill effects;
- Plasma all-four-shield and doubled-range-penalty behavior;
- Mauler always-hit and doubled-range-dissipation behavior.

Gate 3 may implement a bounded subset only when it adds direct tests. Everything else must remain labeled deferred rather than silently approximated as fully original.

## 12. Reference-lab freeze

The Slice-17.0 fixtures remain the acceptance laboratories.

### Baseline / Average

Expected Beam catalog:

- Laser Cannon only.

Acceptance uses:

- standard Laser;
- legal Heavy/Point Defense choice visibility where executable-compatible;
- refinement-locked special modifiers.

### Mid-Tech

Expected Beam catalog:

- Mass Driver;
- Laser Cannon;
- Fusion Beam;
- Ion Pulse Cannon;
- Graviton Beam;
- Neutron Blaster.

Primary integration vectors include:

- fixed-damage versus range-damage weapons;
- differing executable modifier masks;
- at least one M1 special-modifier unlock;
- Space/PP miniaturization;
- two rows using the same Beam with different configurations.

### All-Tech

Expected Beam catalog:

- all 13 frozen Beams.

Primary acceptance includes:

- all eight modifier identities somewhere in the available matrix;
- late Beam weapons and field-74 choices;
- M2 Autofire/Enveloping eligibility where executable-compatible;
- 8-row capacity/rejection;
- mixed weapon configurations;
- save -> construction -> built Ship -> BattleSession snapshot preservation.

## 13. Gate-3 implementation order

Gate 3 should proceed in bounded blocks:

1. ruleset Beam data + golden table tests;
2. core snapshot schema + persistence/backward compatibility;
3. shared modifier/miniaturization legality and quote helpers;
4. Designer catalog + draft preview;
5. save/compiler/construction snapshot path;
6. Weapon Selector HMI;
7. installed-weapon configuration HMI;
8. BattleSession lossless projection + any explicitly selected bounded Tactical consumers;
9. reference-lab/browser/persistence regressions.

Short checkpoints should be placed after blocks 2, 4, 7 and final QA.

## 14. Gate-2 acceptance

- [x] Supported Beam matrix frozen.
- [x] Slot/count model frozen.
- [x] Modifier compatibility/rejection reasons frozen.
- [x] Cost/Space/miniaturization rules frozen.
- [x] ShipDesignSpec weapon snapshot shape frozen.
- [x] Slice-17.0 Designer/Combat Lab fixtures frozen.
- [x] Weapon-selector/configuration HMI and filter contract frozen.
- [x] Tactical support/defer boundary frozen.

**Gate 2 result: PASS / contract frozen. Gate 3 is next but not started.**
