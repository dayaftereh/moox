# Slice 17.4 Gate 2 - Ship Specials and Fighter Systems Contract Freeze
Date: 2026-09-27
Status: **Gate 2 complete; contract frozen; Gate 3 implementation ready / not started**
Parent: Slice 17 - Military Ship Design + Tactical Combat
Child: Slice 17.4 - Ship Specials and Fighter Systems
Gate-1 evidence: docs/research/SLICE_17_4_GATE1_SPECIALS_FIGHTERS_AUDIT_2026-09-26.md
## 1. Contract goal
Freeze one server-authoritative representation for original Ship Specials and Fighter/Bomber carrier systems before product implementation.
This contract owns:
- which original rows are visible/installable in 17.4;
- exact slot, technology, exclusion and quantity legality;
- exact Space/PP/miniaturization order;
- strategic effects implemented in 17.4;
- immutable ShipDesignSpec representation;
- Designer catalog/preview boundaries;
- exact Tactical metadata/consumer handoff to Slice 17.8.
Gate 3 may implement this contract. It may not silently broaden it.
## 2. Frozen supported matrix
### 2.1 Selectable Ship Specials
Original _specials IDs **1..38** are accepted as normal Designer Special choices.
Original record **39 (regeneration) is not a normal Designer choice** and stays locked in 17.4.
Accepted support classes:
- **Class 1 - design + strategic effect complete in 17.4**
  - battle_pods (id 4)
  - extended_fuel_tanks (id 11)
- **Class 2 - design + already-supported Tactical effect complete in 17.4**
  - inertial_nullifier (id 17)
  - inertial_stabilizer (id 18)
- **Class 3 - design representation complete; Tactical effect explicitly deferred to Slice 17.8**
  - every other normal selectable _specials row id 1..38.
- **Class 4 - locked**
  - regeneration original record id 39.
No Class-3 row may be presented as tactically implemented.
### 2.2 Fighter / bomber rows
The following original Weapon rows are accepted in 17.4:
| Original row | Canonical ID | Tech | Base Space | Base PP | Craft per Bay | 17.4 support |
| ---: | --- | ---: | ---: | ---: | ---: | --- |
| 29 | heavy_fighter_bays | 83 | 80 | 50 | 4 | design/count/payload metadata; Tactical fighter behavior deferred to 17.8 |
| 30 | bomber_bays | 31 | 60 | 30 | 4 | design/count/payload metadata; Tactical bomber behavior deferred to 17.8 |
| 31 | fighter_bays | 66 | 30 | 10 | 4 | design/count/payload metadata; Tactical fighter behavior deferred to 17.8 |
They remain Weapon rows. 17.4 does **not** convert them into Special slots.
### 2.3 Automatic/non-slot effects
megafluxers technology 89 is accepted as an automatic design-capacity effect:
- not selectable;
- not stored in the 8 Special slots;
- applied by authoritative compile/preview when the Empire knows the technology.
advanced_damage_control is not a selectable _specials row and has no 17.4 design-slot representation. Its post-battle lifecycle effect is outside this contract.
## 3. Frozen original Special base table
Hull order in the six-element arrays is:
Frigate / Destroyer / Cruiser / Battleship / Titan / Doom Star.
The local 1.31 Orion2.exe table is authoritative. The values below are frozen inputs, before normal miniaturization unless stated otherwise.| ID | Canonical key | Tech | Base Space F/D/C/B/T/DS | Base PP F/D/C/B/T/DS | Exclusion IDs |
| ---: | --- | ---: | --- | --- | --- |
| 1 | achilles_targeting_unit | 1 | [10 15 25 50 100 250] | [15 22 37 75 150 375] | [1] |
| 2 | augmented_engines | 20 | [10 15 25 50 100 250] | [5 7 12 25 50 125] | [2] |
| 3 | automated_repair_unit | 23 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [3] |
| 4 | battle_pods | 25 | [-12 -30 -60 -125 -250 -600] | [20 60 175 400 800 2500] | [4] |
| 5 | battle_scanner | 26 | [12 18 30 60 125 325] | [5 7 12 25 50 125] | [5] |
| 6 | cloaking_device | 38 | [12 18 30 60 125 325] | [12 18 30 60 125 325] | [23 6] |
| 7 | damper_field | 45 | [5 7 12 25 50 125] | [7 10 18 40 75 190] | [7 13 20 30] |
| 8 | displacement_device | 53 | [10 15 25 50 100 250] | [15 22 40 75 150 375] | [8] |
| 9 | ecm_jammer | 57 | [10 15 25 50 100 250] | [5 7 12 25 50 125] | [21 9 38] |
| 10 | energy_absorber | 60 | [12 18 30 60 125 325] | [12 18 30 60 125 325] | [10] |
| 11 | extended_fuel_tanks | 63 | [12 18 30 60 125 325] | [5 10 15 30 60 100] | [11] |
| 12 | fast_missile_racks | 64 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [12] |
| 13 | hard_shields | 81 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [13 7] |
| 14 | heavy_armor | 82 | [10 15 25 50 100 250] | [5 7 12 25 50 125] | [14] |
| 15 | high_energy_focus | 85 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [15] |
| 16 | hyper_x_capacitors | 90 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [16] |
| 17 | inertial_nullifier | 93 | [15 25 40 75 150 375] | [22 40 60 100 225 500] | [18 17] |
| 18 | inertial_stabilizer | 94 | [15 25 40 75 150 375] | [5 7 12 25 50 125] | [17 18] |
| 19 | lightning_field | 102 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [19] |
| 20 | multi_phased_shields | 112 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [20 7] |
| 21 | multi_wave_ecm_jammer | 111 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [9 21 38] |
| 22 | phase_shifter | 125 | [10 15 25 50 100 250] | [15 22 37 75 150 375] | [33 22] |
| 23 | phasing_cloak | 126 | [15 25 40 75 150 375] | [22 40 60 100 225 500] | [6 23] |
| 24 | quantum_detonator | 150 | [2 3 7 15 30 75] | [3 5 10 22 45 100] | [24] |
| 25 | rangemaster_unit | 151 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [25] |
| 26 | reflection_field | 153 | [10 15 25 50 100 250] | [15 22 37 75 150 375] | [26] |
| 27 | reinforced_hull | 56 | [10 15 25 50 100 250] | [5 7 12 25 50 125] | [27] |
| 28 | scout_lab | 158 | [12 20 40 75 150 375] | [7 12 20 37 75 180] | [28] |
| 29 | security_stations | 159 | [5 7 12 25 50 125] | [2 3 6 12 25 70] | [29] |
| 30 | shield_capacitors | 161 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [30 7] |
| 31 | stealth_field | 172 | [10 15 25 50 100 250] | [10 15 25 50 100 250] | [31] |
| 32 | structural_analyzer | 175 | [7 10 18 40 75 190] | [10 15 25 50 100 250] | [32] |
| 33 | sub_space_teleporter | 177 | [10 15 25 50 100 250] | [15 22 37 75 150 375] | [22 33] |
| 34 | time_warp_facilitator | 185 | [15 25 40 75 150 375] | [22 40 60 100 225 500] | [34] |
| 35 | transporters | 190 | [10 15 25 50 100 250] | [5 7 12 25 50 125] | [35] |
| 36 | troop_pods | 192 | [12 18 30 60 125 325] | [5 7 12 25 50 125] | [36] |
| 37 | warp_dissipater | 196 | [15 25 40 75 150 375] | [15 25 40 75 150 375] | [37] |
| 38 | wide_area_jammer | 199 | [20 30 50 100 200 500] | [20 30 50 100 200 500] | [9 38 21] |
| 39 | regeneration | 23 | [4 10 20 40 80 200] | [5 10 20 50 125 500] | [3] |
Important:
- row 39 is evidence-only/locked;
- local 1.31 exclusion arrays are authoritative, including the known differences from later 1.50 defaults;
- Gate 3 must normalize these values into repository-owned rules data/tests. Runtime product code must not parse Orion2.exe.
## 4. Canonical persisted Special snapshot
Gate 3 adds a dedicated Special identity conceptually equivalent to:
~~~go
type ShipSpecialMount struct {
    Slot      int    // json: slot
    SpecialID string // json: special_id
}
~~~
ShipDesignSpec gains an optional Specials slice.
Frozen invariants:
- maximum 8 entries;
- slots are 0..7;
- entries are strictly ascending by slot;
- SpecialID is a normalized canonical key;
- normal saved designs may use only accepted IDs 1..38;
- duplicate Special IDs are illegal;
- there is no Count field for a normal Special;
- no Space/PP values are persisted per Special; authoritative derived totals remain on ShipDesignSpec;
- clone/copy paths must deep-copy the slice.
Fighter/Bomber/Heavy Fighter stay in existing ShipDesignSpec.Weapons as normal ShipWeaponMount rows.
## 5. Technology / visibility contract
A selectable Special or Fighter row is installable only when its technology is known by the Empire under the existing server-owned research ownership rules.
Catalog behavior:
- known + supported: selectable unless a draft-specific legality rule blocks it;
- unknown: visible may be retained for progression UX, but Available=false and technology lock reason is server-owned;
- id39 Regeneration: never emitted as an installable normal Special;
- automatic Megafluxers is projected through automatic-system/final-capacity metadata, not through the Special choice list.
The browser must not infer technology legality.
## 6. Slot, duplicate, exclusion and prerequisite contract
### 6.1 Eight Special slots
A design has exactly 8 Special slots, independent of the existing 8 Weapon slots.
Normal Specials:
- use one slot each;
- cannot be stacked by count;
- cannot appear twice.
### 6.2 Exclusions
Each candidate uses its local-1.31 six-ID exclusion array.
The server must reject a candidate when any nonzero exclusion ID is already installed in another Special slot.
Important named mutual-exclusion groups include:
- Cloaking Device <-> Phasing Cloak;
- Damper Field <-> Hard Shields / Multi-Phased Shields / Shield Capacitors;
- ECM Jammer <-> Multi-Wave ECM Jammer <-> Wide Area Jammer;
- Inertial Stabilizer <-> Inertial Nullifier;
- Phase Shifter <-> Sub-Space Teleporter.
Self-exclusion on IDs 1..38 is the original duplicate rule.
### 6.3 Shield prerequisite
These Specials require a non-empty primary ShieldID:
- id 13 hard_shields;
- id 20 multi_phased_shields;
- id 30 shield_capacitors.
The requirement is validated server-side for preview and save.
## 7. Special Space / PP / miniaturization contract
### 7.1 Normal Special rows
For a normal accepted Special other than Battle Pods:
1. select base Space and base PP from the frozen six-hull row;
2. resolve the Special technology miniaturization level through the existing authoritative technology path;
3. use **Space class 1**;
4. reduce positive Space with the already-frozen class-1 Space reducer;
5. reduce positive PP with the already-frozen universal Cost reducer;
6. add the resulting Special Space/PP once to the draft totals.
Class-1 Space factors:
- mini 0: 100%;
- mini 1: 80%;
- mini 2: 70%;
- mini 3: 60%;
- mini 4: 50%;
- mini 5+: 40%.
Universal Cost factors:
- mini 0: 100%;
- mini 1: 75%;
- mini 2: 55%;
- mini 3: 40%;
- mini 4: 30%;
- mini 5+: 25%.
Rounding/minimum-one behavior reuses the already-audited 17.1/17.2 helpers. No 17.4 parallel approximation is allowed.
### 7.2 Battle Pods exception
battle_pods id 4 bypasses normal Special miniaturization.
Frozen Space by hull:
-12 / -30 / -60 / -125 / -250 / -600.
Frozen PP by hull:
20 / 60 / 175 / 400 / 800 / 2500.
Negative Space is intentionally added to SpaceUsed; it is not clamped to zero at the row stage.
### 7.3 Base design / production cost
Special and Fighter row PP contributes to BaseDesignCostPP before the existing final government production-cost transform.
There is one authoritative quote pipeline for:
- catalog preview;
- live draft preview;
- save validation;
- persisted ShipDesignSpec;
- immutable built-Ship snapshot.
## 8. Capacity and strategic-effect contract
### 8.1 Megafluxers
Let baseHullSpace be the original hull capacity.
When technology 89 megafluxers is known:
    finalHullSpace = floor(baseHullSpace * 125 / 100)
Otherwise:
    finalHullSpace = baseHullSpace
ShipDesignSpec.HullSpace stores this final compiled capacity.
Megafluxers is compile/preview automatic and consumes no Special slot.
Existing saved revisions and already-built Ship snapshots are not mutated retroactively by research. A newly compiled/revised design uses current Empire technology ownership.
### 8.2 Battle Pods
Battle Pods change used Space through their frozen negative row value.
They do **not** multiply HullSpace.
Therefore final validation order is:
1. compile HullSpace including automatic Megafluxers;
2. compile mandatory-system, Weapon/Fighter and Special Space;
3. include Battle Pods' negative Space in the Special total;
4. compute SpaceUsed;
5. reject only after the full draft if SpaceUsed > HullSpace.
No client-side capacity formula is allowed.
### 8.3 Extended Fuel Tanks
Let r be the already-compiled primary Fuel Cell range.
Without Extended Fuel Tanks:
    FuelRangeParsecs = r
With id 11 extended_fuel_tanks:
    FuelRangeParsecs = floor((3*r + 1) / 2)
For positive integer ranges this is ceil(1.5 * r).
This value is persisted in the existing ShipDesignSpec.FuelRangeParsecs field and is used by the existing strategic movement/range path.
## 9. Fighter/Bomber quote and configuration contract
Fighter rows use existing Weapon slots and ShipWeaponMount.Count.
Frozen meaning:
- Count = number of Bays in that Weapon row;
- effective craft payload = Count * 4;
- count must be positive;
- practical maximum is governed by Weapon-slot and final Space legality; no new arbitrary fighter count cap is introduced.
Configuration:
- canonical MountType = standard;
- no Heavy/Point-Defense mount choice;
- no Weapon modifier choices;
- no ammo/rack choice;
- no firing-arc choice;
- FiringArcID remains empty;
- browser must not overload Ammo to represent craft count.
All three rows use the normal Weapon technology-level Cost reducer and **Space class 1** reducer.
## 10. Server-authoritative Designer API contract
### 10.1 Catalog
Gate 3 adds a Specials catalog projection parallel to the current Component/Weapon projections.
Frozen conceptual fields:
~~~text
MilitaryDesignerSpecialChoice:
- ID
- NameKey
- OriginalID
- TechnologyID
- TechnologyFieldID
- TechnologyKnown
- Available
- LockReason
- SupportClass
- TacticalSupport
- MiniaturizationLevel
- CostReductionPercent
- SpaceReductionPercent
- EffectiveCostByHullPP[6]
- EffectiveSpaceByHull[6]
- ExcludedSpecialIDs
- RequiresShield
~~~
Exact Go field layout may follow repository conventions, but the JSON contract must expose equivalent authoritative facts.
The existing MilitaryDesignerAutomaticSystems projection must expose enough final capacity information for the client to show Megafluxers' effect without recalculating it locally.
Fighter/Bomber/Heavy Fighter are appended to the existing Weapon catalog with:
- Kind = fighter;
- the canonical IDs from section 2.2;
- no modifier/ammo/arc choices;
- TacticalSupport = metadata_only.
### 10.2 Draft preview
MilitaryDesignDraft gains Specials.
Preview gains Special-row quote metadata equivalent to:
- Slot;
- SpecialID;
- MiniaturizationLevel;
- UnitSpace;
- UnitCostPP;
- SupportClass;
- TacticalSupport.
The preview remains the source of:
- final HullSpace;
- final SpaceUsed;
- remaining Space;
- base/final PP;
- draft rejection reason/message;
- canonical saved Spec.
### 10.3 Client boundary
The web client may:
- select/remove/reorder by explicit slot;
- display server-projected costs, Space and support state;
- locally stage a draft before requesting preview.
The web client may **not**:
- calculate Special Space/PP;
- calculate Megafluxers/Battle Pods capacity;
- calculate Extended Fuel Tanks range;
- decide technology, duplicate, exclusion or Shield legality;
- invent Tactical support.
## 11. Stable rejection contract
Existing reasons remain:
- technology_required where the general component path applies;
- space_exceeded;
- existing weapon_* reasons for Fighter-row slot/count/configuration errors.
Gate 3 adds these stable Special reasons:
- special_unknown
- special_technology_required
- special_slot_out_of_range
- special_slot_duplicate
- special_duplicate
- special_exclusion_conflict
- special_requires_shield
- special_not_supported
special_not_supported covers normal save attempts for id39 or any non-frozen Special identity.
Draft preview must return these reasons through the repository's existing reason-prefix convention.
## 12. HMI contract
Add a dedicated **Special Systems** area to the Ship Designer.
Frozen behavior:
- up to 8 Special slots;
- Add / Replace / Remove interactions;
- choices come from server catalog;
- draft-specific conflicts come from authoritative preview;
- row/card shows localized name, effective Space and PP for the selected hull, and support badge;
- Class-3 choices must visibly state that the Tactical effect is deferred to Slice 17.8;
- Class-2 Inertial choices may state current Tactical turning support;
- Class-1 Battle Pods / Extended Fuel Tanks may show their active strategic/design effect;
- locked/unknown technology remains visibly locked with server reason.
Fighter/Bomber/Heavy Fighter remain in the existing Weapon selector:
- Kind = fighter;
- count edits number of Bays;
- UI may display derived 4 craft / Bay;
- no mount/modifier/ammo/firing-arc controls are shown for those rows.
## 13. Immutable persistence / construction contract
The exact path is:
    Designer draft
    -> authoritative preview/compile
    -> ShipDesign.Spec.Specials + existing Weapons
    -> save/reload/revision
    -> Construction item
    -> built Ship.Spec
    -> BattleSession projection
Requirements:
- Special slots/IDs survive exactly;
- Fighter row IDs/counts survive exactly;
- later design revisions do not mutate already-built Ships;
- category-3 Special identities still survive into built Ships even though their Tactical effects are deferred;
- no browser-only metadata is required to reconstruct the design.
## 14. Tactical handoff contract
Gate 3 adds Special metadata to the Tactical immutable spec, conceptually:
~~~text
TacticalSpecialSpec:
- Slot
- SpecialID
- TacticalSupport
~~~
TacticalShipSpec carries the canonical Special rows.
Frozen TacticalSupport values:
- battle_pods / extended_fuel_tanks -> strategic_only;
- inertial_stabilizer / inertial_nullifier -> turning_mode;
- every Class-3 Special -> metadata_only.

Frozen support mapping:
- inertial_stabilizer -> existing TacticalTurningInertialStabilizer;
- inertial_nullifier -> existing TacticalTurningInertialNullifier;
- no Inertial special -> existing TacticalTurningNormal;
- Class-3 Specials -> metadata_only;
- Battle Pods / Extended Fuel Tanks -> metadata only in Tactical because their owned 17.4 effects are design/strategic;
- Fighter/Bomber/Heavy Fighter Weapon rows -> existing Tactical Weapon metadata with TacticalSupport=metadata_only until Slice 17.8.
Slice 17.8 owns:
- fighter launch/movement/targeting/interception/fire/return/reload/damage;
- repair/regeneration combat behavior;
- cloak/phasing/phase-shift;
- boarding/transport/security actions;
- ECM/jamming/anti-missile Special effects;
- shield/damage-field Special effects;
- active/time/teleport/area/combat-special actions;
- Beam-support Special resolution not already implemented elsewhere.
Gate 3 must not fake any of those effects.
## 15. Reference-lab freeze
Use existing game-triangle-all-tech / Doom-Star design support.
Required Gate-3 vectors:
### A - Doom Star capacity/survivability snapshot
All-Tech Doom Star:
- automatic Megafluxers;
- Battle Pods;
- Reinforced Hull;
- Heavy Armor.
Expected capacity:
- base Doom Star Space = 1200;
- Megafluxers final HullSpace = 1500;
- Battle Pods Special Space = -600.
Verify quote, save/reload/revision/build snapshot exactly.
### B - Doom Star carrier
All-Tech Doom Star with:
- Fighter Bays;
- Bomber Bays;
- Heavy Fighter Bays.
Verify:
- exact base/miniaturized Space/PP;
- Count = Bays;
- 4 craft/Bay metadata;
- no false Tactical launch behavior.
### C - Inertial consumer
Separate legal revisions/designs:
- Inertial Stabilizer -> Tactical turning mode inertial_stabilizer;
- Inertial Nullifier -> Tactical turning mode inertial_nullifier;
- installing both is rejected by exclusion.
### D - strategic range
Design with Extended Fuel Tanks:
- authoritative FuelRangeParsecs = ceil(1.5 * base fuel-cell range);
- strategic movement consumes persisted range;
- save/reload/build keeps the exact compiled value.
### E - exclusion / prerequisite
Deterministically reject:
- Cloaking Device + Phasing Cloak;
- ECM Jammer + Multi-Wave ECM Jammer;
- Phase Shifter + Sub-Space Teleporter;
- Hard Shields without a primary Shield;
- duplicate Special IDs;
- ninth Special slot.
## 16. Gate-3 implementation order
1. normalize original Special/Fighter rules data;
2. add core ShipSpecialMount + clone/validation;
3. extend authoritative compile/quote/preview;
4. add automatic Megafluxers and Extended Fuel Tanks effects;
5. add Fighter rows to Weapon catalog/quote;
6. extend persistence/construction immutable snapshot tests;
7. project Special metadata + Inertial turning mode to Tactical;
8. add Designer Special HMI and Fighter UX;
9. add reference-lab/browser matrices;
10. run full Go/vet/web/browser regression.
## 17. Gate result
Gate 2 is **complete**.
Frozen:
- supported Special/Fighter matrix;
- per-system support classification;
- original compatibility, slot, duplicate, prerequisite and quantity rules;
- Space/PP/miniaturization pipeline;
- automatic Megafluxers and selectable Battle Pods capacity ordering;
- Extended Fuel Tanks strategic range formula;
- persisted snapshot representation;
- server catalog/preview and rejection boundary;
- HMI honesty boundary;
- exact Slice-17.8 Tactical ownership;
- Gate-3 reference vectors and implementation order.
No product implementation was performed in Gate 2.