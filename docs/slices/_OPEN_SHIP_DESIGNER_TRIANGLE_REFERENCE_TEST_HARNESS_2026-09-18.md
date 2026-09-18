# Open Slice 17.0 - Ship Designer / Triangle Reference Test Harness

Status: **Gate 1 complete / Gate 2 next**

Opened: 2026-09-18
Session: ses-20260918T152048-00240c450344
Baseline commit: d69f9b3

Parent: docs/slices/PLANNED_17_MILITARY_SHIP_DESIGN_COMPONENT_WEAPON_BREADTH.md
Plan: docs/slices/PLANNED_17_0_SHIP_DESIGNER_TRIANGLE_REFERENCE_TEST_HARNESS.md

## Gate 1 checklist

- [x] Audit `-reference-games`, standard reference game and Triangle reference-game registration.
- [x] Audit technology bootstrap APIs/state and define safe deterministic technology grants.
- [x] Audit normal Construction/Turn APIs for bounded automated advancement.
- [x] Audit BattleSession/Triangle fixture entry and built-Ship snapshot requirements.
- [x] Define stable reference scenario IDs and ownership.
- [x] Define strict dev-only exposure/authorization boundary.

## Permanent Gate-1 evidence

docs/research/SLICE_17_0_GATE1_REFERENCE_HARNESS_AUDIT_2026-09-18.md

## Gate-1 result

The existing architecture is suitable without a parallel rules engine. Reference scenarios may alter deterministic initial state only; after bootstrap they must reuse ordinary Designer validation, Colony Construction, hosted turn resolution, persistent built-Ship snapshots and Tactical/BattleSession handoff.

High-tech Designer state must not reuse or unlock the deferred Advanced New Game path. Tactical support remains narrower than strategic ShipDesignSpec breadth and must only be advertised for equipment the Tactical consumer actually supports.

Future development controls require a double authorization boundary: explicit server reference capability plus immutable per-game reference-scenario metadata. Game-ID naming alone is not authorization.

The Triangle reference game is a durable family, not a disposable fixture: keep the existing baseline and add Turn-1 Mid-Tech plus Turn-1 All-Tech/all-supported variants. Designer and Construction lab controls belong on that Triangle family rather than on separate game-designer-* throwaway games.

## Gate 2 next

Freeze the exact Triangle baseline/Mid-Tech/All-Tech matrix and seeds, exact technology profiles, turn-runner bounds/stop semantics, built-Ship materialization contract, reference metadata, HTTP/HMI capability exposure and persistence behavior.

## Guardrail

Do not begin Gate 2 implementation before explicit user release. Do not open Slice 17.1 or later children. No push.
