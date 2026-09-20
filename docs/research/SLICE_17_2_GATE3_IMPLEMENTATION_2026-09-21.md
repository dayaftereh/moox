# Slice 17.2 Gate 3 - Beam Weapons and Beam Mount Modifiers Implementation

Date: 2026-09-21
Result: **PASS / implementation complete**
Input contracts:

- `docs/research/SLICE_17_2_GATE1_BEAM_WEAPONS_AUDIT_2026-09-20.md`
- `docs/research/SLICE_17_2_GATE2_BEAM_CONTRACT_FREEZE_2026-09-20.md`
- `docs/research/SLICE_17_2_GATE2_HMI_ADD_FLOW_AMENDMENT_2026-09-20.md`

## 1. Gate-3 result

Slice 17.2 now has an authoritative 13-Beam design path from rules data through Designer selection, server preview, save, construction/built-Ship snapshot and BattleSession metadata projection.

Implemented Beam identities:

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

The canonical Laser tactical firing baseline remains unchanged. Broader Beam effects and mount/modifier combat behavior that Gate 2 marked deferred are transported as authoritative metadata but are not silently simulated.

## 2. Authoritative rules data

`data/rulesets/moo2-1.31/tactical_combat.json` now carries:

- the 13 frozen Beam rows;
- original index and technology/field identity;
- base Space and PP;
- min/max damage;
- executable modifier masks;
- innate original-effect IDs;
- tactical support/defer metadata;
- Beam Space miniaturization class-1 factors;
- cost miniaturization factors;
- the eight frozen mount/modifier definitions.

`internal/ruleset/tactical_combat.go` exposes the normalized Beam table through `BeamWeapons()` and `WeaponByID()` while preserving the legacy single-Laser field for compatibility with the canonical Tactical baseline.

The reviewed ruleset fingerprint is now:

`275fdd2344cafa97ff9e0460669a02db73ec6bf3c10d08352bc0a836cc7f0a9d`

## 3. Weapon snapshot schema

`core.ShipWeaponMount` now persists:

- `Slot`;
- `WeaponID`;
- `Count`;
- `MountType`;
- sorted canonical `ModifierIDs`.

Mount types:

- Standard;
- Heavy;
- Point Defense.

Special modifiers:

- Armor Piercing;
- Continuous;
- No Range Dissipation;
- Shield Piercing;
- Autofire;
- Enveloping.

Legacy missing mount type normalizes to Standard. Modifier lists are validated for known IDs, canonical order and duplicates. Deep-clone helpers copy the modifier slice rather than aliasing it.

## 4. Server-authoritative legality and arithmetic

`internal/game/military_beam_weapons.go` owns the shared Beam legality/quote path.

It validates:

- the 13 frozen Beam IDs;
- technology ownership;
- slot range and uniqueness;
- positive count;
- mount compatibility from the executable mask;
- special-modifier compatibility;
- refinement/miniaturization requirements;
- duplicate/unknown modifiers.

Stable rejection reasons from Gate 2 are preserved.

The frozen arithmetic order is implemented server-side:

1. mount factor;
2. additive special-modifier factor;
3. Beam miniaturization;
4. physical count.

TypeScript does not reproduce the authoritative rules.

## 5. Field-74 terminal fix

Gate-3 QA exposed an important original-data edge case.

Original technology Field 74 is the special Antaran field and is intentionally self-linked:

- `previous_id = 74`;
- `next_id = 74`.

The previous general miniaturization successor loop had no cycle guard. Once Particle Beam / Death Ray entered the expanded Designer catalog, All-Tech snapshot generation could loop forever on Field 74.

`shipComponentMiniaturizationLevel()` now seeds a visited-field set with the starting field and stops on a repeated successor.

Result:

- Particle Beam and Death Ray correctly remain M0 on special Field 74;
- malformed/cyclic successor data can no longer hang Player Snapshot / Designer projection;
- a permanent regression covers technologies 123 and 47 plus the Particle Beam quote.

## 6. Server-authoritative draft preview

A dedicated design-preview path now spans:

- game rules;
- session;
- host;
- HTTP server;
- web API.

Route:

`POST /api/v1/games/{gameID}/seats/{seatID}/military-design-preview`

The preview returns:

- compiled design totals;
- Hull Space / used Space / remaining Space;
- base and production PP;
- canonical weapon rows;
- per-row unit and total Space/PP;
- miniaturization level;
- stable rejection reason/message;
- save availability.

Preview and Save share the same Beam compiler/validation helpers.

## 7. Weapon Selector HMI

The user-approved Gate-2 HMI amendment is implemented.

Main flow:

1. press the compact weapon `+`;
2. open Weapon Selector;
3. search/filter/sort the 13-Beam catalog;
4. click a weapon;
5. configure Standard / Heavy / Point Defense;
6. configure compatible special modifiers;
7. choose quantity;
8. receive live server-authoritative Space/PP/design preview;
9. press `Install / Einbauen`;
10. return to the compact Installed Weapons list.

The selector includes:

- free-text search;
- availability filter;
- modifier-support filter;
- Technology / Name / Damage / Space / PP sorting;
- visible locked/refinement choices;
- German and English Beam/mount/modifier labels.

The main Designer remains stable in height. Installed rows keep compact inline `- count +` controls plus icon-only Remove.

The browser smoke verifies a Standard Laser and a Heavy Laser installed through this selector flow.

## 8. Construction and immutable built-Ship snapshots

The existing `cloneShipDesignSpec()` path now deep-clones the expanded weapon mount schema.

Regression coverage verifies that:

- configured Beam mount/modifier data survives Save;
- configured Beam data survives construction into the built Ship;
- later design mutation does not mutate the already-built Ship;
- modifier slices do not alias between snapshots.

## 9. BattleSession projection

The BattleSession weapon snapshot now carries:

- Slot / WeaponID / Count;
- MountType;
- ModifierIDs;
- min/max damage;
- innate effect IDs;
- tactical support/defer mode.

Game encounter projection now uses the actual frozen Beam row instead of substituting Laser damage for every weapon.

A regression uses:

- Standard Laser;
- Heavy Phasor;
- Shield Piercing.

It proves that the Phasor metadata reaches `battle.Session` losslessly and is deep-cloned.

### Tactical scope honesty

Only the unmodified Standard Laser remains the firing baseline.

Metadata-only Beam rows:

- are accepted into the BattleSession;
- are visible in Tactical snapshots;
- are not marked Ready;
- do not receive legal Fire actions.

This preserves the Gate-2 support/defer boundary without inventing broad Tactical behavior.

## 10. Reference-lab / HMI regression

The reference-lab browser smoke was updated from the former immediate-Laser-add interaction to the frozen selector-first workflow.

It now checks, among the existing broad regression set:

- 13 weapon selector rows in All-Tech;
- search/filter controls;
- Standard / Heavy / Point Defense choices;
- special-modifier section including Autofire;
- explicit Install action;
- Standard and Heavy Laser insertion;
- compact installed rows;
- Heavy configuration summary;
- icon-only accessible remove controls;
- mobile 390px and desktop layout;
- ordinary-game isolation.

## 11. Final Gate-3 QA

Final verification on the completed working tree:

- `go test ./... -count=1 -timeout 120s` - **PASS**
- `go vet ./...` - **PASS**
- `npm run build` from `web/` - **PASS**
- `git diff --check` - **PASS** (line-ending warnings only)
- All-Tech snapshot endpoint on fresh port 7201 - **HTTP 200**
- `npm run check:reference-lab:browser -- http://127.0.0.1:7201` - **PASS**

The final browser smoke ran against a freshly restarted review server built from the final Gate-3 tree.

## 12. Gate-3 acceptance

- [x] Expand Beam catalog/rules.
- [x] Expand save/design validation and weapon mounts.
- [x] Implement selected modifier metadata and bounded Tactical consumers.
- [x] Update Designer HMI for multiple Beam slots/counts/modifiers.
- [x] Preserve construction/persistence/snapshot semantics.
- [x] Add deterministic Tactical/Designer Lab regressions.

**Gate 3 result: PASS. Gate 4 is next and has not started.**
