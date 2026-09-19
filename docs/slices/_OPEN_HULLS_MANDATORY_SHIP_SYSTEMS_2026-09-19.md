# OPEN - Slice 17.1 Hulls and Mandatory Ship Systems

Opened: 2026-09-19
Parent: Slice 17 - Military Ship Design Component/Weapon Breadth
Status: **Gates 1-3 complete / Gate 4 next**

## Objective

Re-audit and freeze evidence for the six original-valid military hulls and mandatory ship-system breadth before any Gate-2 contract or Gate-3 implementation begins.

## Gate 1 active scope

- all six hull size/space/cost/Command Point data;
- Colony infrastructure/buildability constraints for large hulls;
- Titan and Doom Star technology locks;
- Warp Drive, Computer, Armor, Shield and Fuel Cell catalogs with technology IDs;
- original mandatory-component cost/space behavior;
- direct miniaturization evidence and current implementation gaps;
- derived strategic/tactical stats attributable to these mandatory systems;
- explicit downstream classification for tactical-only effects that 17.1 must not fake.

## Guardrail

Gates 1-3 are complete. Do not start Gate 4 until the user explicitly releases it. Do not open Slice 17.2 or later children.

Slice 17.0 Reference/Test Harness is closed and remains the QA foundation for this work.

## Gate 1 complete

Evidence: `docs/research/SLICE_17_1_GATE1_HULLS_MANDATORY_SYSTEMS_AUDIT_2026-09-19.md`.

Gate 2 contract freeze is complete.

## Primary miniaturization addendum

The hash-identical MOO2 1.31 executable was re-located at `C:\ASH\Installation\mastori2\Orion2.exe`. Gate-1 evidence now directly verifies miniaturization level traversal, Hyper-Advanced repeat contribution, exact cost factors, exact Shield Space factors and integer rounding. The former secondary-evidence warning is closed.

## Gate 2 complete

Evidence: `docs/research/SLICE_17_1_GATE2_HULL_SYSTEM_CONTRACT_FREEZE_2026-09-19.md`.

The contract freezes six-hull availability/buildability, Computer/Shield selection, automatic systems, exact original miniaturization, dynamic effective saved-design recompilation, fixed queue-time construction snapshots, CP/rejection semantics and durable Baseline/Mid/All-Tech fixtures.

## Gate 3 complete

Evidence: `docs/research/SLICE_17_1_GATE3_IMPLEMENTATION_2026-09-19.md`.

The frozen contract is implemented end-to-end across server authority, construction/persistence snapshots, player projection, API/HMI and deterministic backend/browser checks. Gate 4 QA/close is next but not started. Wait for explicit user release.


## Gate 3 HMI information amendment

User review requested effect-first Designer information before Gate 4. Computer and Shield are now separate non-dropdown cards, choices show their verified original effects plus current miniaturized cost/Space, and the complete installed-system/effect summary is placed under the Ship preview.

Evidence remains: docs/research/SLICE_17_1_GATE3_IMPLEMENTATION_2026-09-19.md.

Gate 4 is still next and has not started.
