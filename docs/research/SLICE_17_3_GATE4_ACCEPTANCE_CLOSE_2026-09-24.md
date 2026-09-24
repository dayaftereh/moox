# Slice 17.3 Gate 4 - Acceptance Close

Date: 2026-09-24

Status: **accepted; Slice 17.3 closed with Gates 1-4 complete**

Parent slice: `docs/slices/PLANNED_17_3_MISSILES_TORPEDOES_BOMBS_ORDNANCE.md`

Gate-1 evidence: `docs/research/SLICE_17_3_GATE1_ORDNANCE_AUDIT_2026-09-23.md`

Gate-2 contract: `docs/research/SLICE_17_3_GATE2_ORDNANCE_CONTRACT_FREEZE_2026-09-23.md`

Implementation commits:
- `283a7d0 feat(slice): implement 17.3 ordnance designer`
- `8073090 test(slice): cover 17.3 ordnance selector`

## Acceptance

The user manually reviewed the implemented ordnance Ship Designer on the canonical 7171 All-Tech reference path, judged the implementation orderly/good, and explicitly requested Gate 4.

This acceptance closes the **design-side Missile/Torpedo/Bomb breadth** delivered by Slice 17.3. The user also noted that Tactical UI/behavior may require later refinement; that remains intentionally downstream rather than a Slice-17.3 defect.

## Accepted Slice-17.3 product surface

- The authoritative Designer catalog contains the frozen 13 ordnance rows in addition to the accepted 13 Beam rows.
- Missile rack/ammo choices are persisted from `{2,5,10,15,20}` with default 5.
- Torpedo Ammo is fixed at 2; Bomb/Bio Ammo is fixed at 10.
- Every supported 17.3 ordnance snapshot persists the frozen `360` firing arc.
- Technology gating, damage, Space/PP, class-1 miniaturization and legal Missile/Torpedo modifiers follow the Gate-2 contract.
- Ordnance save/preview validation uses stable rejection reasons and the frozen validation order.
- Construction and built-Ship snapshots preserve exact Count/Ammo/FiringArcID metadata independently of later design revisions.
- BattleSession receives the same immutable Ammo/FiringArcID/modifier metadata.
- Deferred ordnance is projected as `metadata_only`, is not Ready and receives no fake Tactical fire action.
- DE/EN Weapon Selector presentation exposes authoritative ordnance choices without introducing free Beam-style firing-arc controls.

## Gate-4 regression evidence

Gate 4 added direct regression proof for previously generic boundaries:

1. `TestSlice173AllTechCatalogExposesHighTechOrdnance`
   - proves high-tech Zeon Missile and Plasma Torpedo are immediately available through the 17.0 All-Tech profile;
   - proves server-owned Ammo/default/fixed-360 catalog metadata.

2. `TestSlice173BuiltMilitaryShipKeepsOrdnanceSnapshotAcrossDesignRevision`
   - saves and builds a Missile design;
   - proves canonical Ammo 5 + FiringArcID 360 persist on the built Ship;
   - performs a real design revision and proves the built revision remains unchanged.

3. `TestSlice173TacticalMetadataPreservesOrdnanceSnapshotWithoutEnablingDeferredFire`
   - proves the Tactical handoff preserves Ammo 5 + FiringArcID 360;
   - proves BattleSession preserves the metadata;
   - proves `metadata_only` ordnance is not Ready and receives no fire action.

## Final QA evidence

The final Gate-4 baseline passed:

- focused Slice-17.3 Gate-4 regression tests;
- `go test ./... -count=1`;
- `go vet ./...`;
- `npm run build` from `web`;
- `git diff --check`;
- fresh canonical `go run ./cmd/moox-server -addr 0.0.0.0:7171 -insecure-allow-nonloopback -enable-persistence -reference-games`;
- `/healthz` = 200;
- `npm run check:reference-lab:browser -- http://127.0.0.1:7171`.

The browser smoke runs against the All-Tech reference game and verifies the 26-row weapon selector plus Slice-17.3 ordnance presence.

## Explicit downstream boundary

Slice 17.3 does **not** claim complete Tactical ordnance behavior.

Planned **Slice 17.7 - Tactical Ordnance** remains the owner of:
- launch/readiness/salvo runtime behavior;
- projectile state, movement, ETA and tracking;
- Point Defense/interception;
- ECM/evasion interactions;
- impact, damage and explosion resolution;
- verified Bomb/Bio Tactical or bombardment behavior;
- Tactical ordnance UI/actions/presentation.

Any later Tactical UI refinement observed after this accepted Designer baseline belongs to that downstream consumer scope unless it reveals loss/corruption of the frozen 17.3 snapshot contract.

## Close result

Slice 17.3 is closed with Gates 1-4 complete. Its `_OPEN_` marker is removed.

No later child is opened by this closeout.

The next planned child in dependency order is **Slice 17.2A - Beam Weapon Firing Arcs**, followed by Slice 17.4.
