# Open Slice 17.0 - Ship Designer / Triangle Reference Test Harness

Status: **Gates 1-3 complete / Gate 4 ready**

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

## Gate 2 complete

Frozen in `docs/research/SLICE_17_0_GATE2_REFERENCE_HARNESS_FREEZE_2026-09-18.md`: durable baseline/Mid-Tech/All-Tech Triangle family, exact technology profiles, same deterministic seed/topology, bounded normal-turn runner, built-Ship materialization, reference metadata/security, HTTP/HMI and persistence behavior.

## Gate 3 complete

Implemented and verified in `docs/research/SLICE_17_0_GATE3_IMPLEMENTATION_2026-09-18.md`. The durable Reference Harness now exists end-to-end through the ordinary authority paths.

Gate-3 amendment `docs/research/SLICE_17_0_GATE3_BUYOUT_DEV_BC_AMENDMENT_2026-09-18.md` adds normal server-authoritative MOO2 BC construction buyout, moves round controls under Development Tools and adds trusted Reference-only +100/+1,000/+10,000 BC grants. Gate 4 remains not started.

## Guardrail

Gate 4 QA/close is ready but not started. Wait for explicit user release. Do not open Slice 17.1 or later children. No push.
