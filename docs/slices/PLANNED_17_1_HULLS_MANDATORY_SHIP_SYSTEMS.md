# Planned Slice 17.1 - Hulls and Mandatory Ship Systems

Status: **active; Gate 1 complete / Gate 2 next**.

## Objective

Expand military design from the narrow supported Frigate baseline to the frozen original-valid hull and mandatory-system breadth before broader weapons/specials are layered on top.

## Primary breadth

- Frigate;
- Destroyer;
- Cruiser;
- Battleship;
- Titan;
- Doom Star;
- Warp Drives;
- Computers / targeting computers;
- Armor materials;
- Shields;
- Fuel Cells;
- hull buildability / infrastructure / technology restrictions;
- authoritative Command Point cost;
- mandatory-system cost/space;
- miniaturization foundation where it affects these components.

## Existing foundation

The six hulls are already normalized with base PP, base Space, strategic pictures and CP size bands. ShipDesignSpec already snapshots hull, drive, computer, armor, shield and fuel. Current Designer Catalog already projects these categories but deliberately locks most choices under the current narrow scope.

## Scope guard

17.1 does not attempt broad Beam/ordnance/special mechanics. It may continue to use the minimal supported weapon baseline for construction/ship-snapshot QA.

## Gate 1 - original/system audit

- [x] Re-audit all six hull original size/space/cost/CP data.
- [x] Re-audit large-hull Colony buildability/infrastructure rules.
- [x] Re-audit Titan/Doom Star technology requirements.
- [x] Re-audit drive, computer, armor, shield and fuel catalogs with technology IDs.
- [x] Re-audit original mandatory-component cost/space and miniaturization behavior.
- [x] Re-audit derived strategic/tactical stats exposed by computers, drives, shields and armor.
- [x] Classify any tactical-only effects as downstream instead of faking them.

## Gate 1 evidence

- `docs/research/SLICE_17_1_GATE1_HULLS_MANDATORY_SYSTEMS_AUDIT_2026-09-19.md`
- Frozen evidence boundary: six hull base values/CP, Cruiser+ Star-Base-or-better construction requirement, Titan/Doom Star tech locks, full mandatory-system identity/tech/base-table audit, original Automatics semantics, current strategic Drive/Fuel behavior, direct executable miniaturization level/factor/rounding evidence, and explicit Tactical downstream classification.

## Gate 2 - hull/system contract freeze

- [ ] Freeze supported hull matrix and buildability.
- [ ] Freeze selectable mandatory component matrix.
- [ ] Freeze technology locks and stable rejection reasons.
- [ ] Freeze cost/space/miniaturization formulas.
- [ ] Freeze CP and derived ShipDesignSpec values.
- [ ] Freeze Designer Lab fixtures covering early/mid/high-tech.

## Gate 3 - implementation

- [ ] Expand server catalog and save validation for frozen hulls.
- [ ] Expand mandatory-component selection/validation.
- [ ] Implement/finalize cost/space/miniaturization foundation.
- [ ] Preserve immutable snapshots in built Ships.
- [ ] Wire Designer HMI controls for supported selections.
- [ ] Wire Colony Construction legality for newly supported hulls.
- [ ] Add deterministic reference-lab and construction fixtures.

## Gate 4 - QA + close

- [ ] Every frozen hull can be designed only when buildable/unlocked.
- [ ] Every frozen mandatory component obeys technology/space/cost rules.
- [ ] Doom Star and other late-game configurations are testable through 17.0 without research grinding.
- [ ] Revision never mutates already-built Ships.
- [ ] Construction and persistence round trips pass.
- [ ] Full tests/vet/web/browser checks and `git diff --check` pass.

## Exit criterion

The normal designer can author and build the frozen hull + mandatory-system combinations through one authoritative rules pipeline, including late-game hulls in reference laboratories, without yet claiming complete weapon/special breadth.
