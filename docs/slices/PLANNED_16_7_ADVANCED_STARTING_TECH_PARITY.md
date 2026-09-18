# Planned slice 16.7 - Advanced starting technology parity and Slice-16 closure

Status: **closed as audited/deferred handoff; Gate 1 audit complete, Gate 2 deferral contract complete, Advanced implementation moved to Slice 23 (2026-09-18)**.

Parent: `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`.
Depends on: Slices 16.1-16.6, especially the Slice 16.5 starting-technology catalog/visual grammar and the Slice 16.6 player-composition integration.

## Objective

Finish the Slice-16 New Game program by promoting the catalog-visible `advanced` starting-technology option from planned/disabled to a fully normalized, server-authoritative and deterministic game-start mode.

This slice exists explicitly so Advanced is not forgotten or implemented as a partial UI-only toggle. It owns the remaining parity work that Slice 16.5 intentionally defers.

**16.7 is the final Slice-16 sub-slice. Do not add further 16.x slices after this one; later breadth moves to the next numbered milestone.**

## Binding Advanced scope

Advanced must be implemented from authoritative evidence rather than inferred from the label. Gate 1 must enumerate every initial-state difference relative to `pre_warp` and `average`, including where applicable:

- authoritative starting technology grants and research state;
- empire-level starting values/resources/preferences;
- homeworld, colony and population state;
- starting buildings/facilities;
- starting fleet, hulls, equipment and ship state;
- race-specific modifiers or exceptions;
- difficulty, galaxy and player-composition interactions;
- deterministic seed behavior;
- Save/Resume parity after an Advanced start.

The existing Slice 16.5 visual card/art direction is reused. Slice 16.7 is primarily the semantic/runtime parity slice; it must not introduce a second, conflicting selector grammar.

## Gate 1 - Advanced audit

- [x] Re-audit authoritative/original Advanced start behavior and normalized source IDs.
- [x] Diff Advanced against the frozen `pre_warp` and `average` initial-state contracts field by field.
- [x] Enumerate all granted technologies/research state and every non-technology bootstrap delta.
- [x] Inventory race-specific interactions and unsupported-mechanic dependencies.
- [x] Audit interactions with difficulty, galaxy settings and Slice 16.6 player composition.
- [x] Define deterministic fixture evidence needed to prove Advanced parity.

Permanent Gate-1 evidence: docs/research/SLICE_16_7_GATE1_ADVANCED_PARITY_AUDIT_2026-09-18.md.

## Gate 2 - freeze

- [x] Freeze the exact server-owned `advanced` contract and accepted ID/label.
- [x] Freeze full empire/population/colony/fleet/research bootstrap semantics.
- [x] Freeze race-specific exception handling and cross-setting validation.
- [x] Freeze the server-derived facts shown on the existing Advanced visual card.
- [x] Freeze the acceptance matrix required before `advanced` can be enabled for submission.

## Downstream roadmap handoff

Gate-2 feasibility work found that the Core/Persistence architecture can carry Advanced, but complete original-style Advanced parity would pull unfinished Building effects, Pollution/Planet breadth and later race mechanics into Slice 16.

By product decision, the complete implementation target is therefore reserved as `PLANNED_23_FULL_ADVANCED_STARTING_CIVILIZATION_PARITY.md`, downstream of:

- Slice 17 ship/component breadth;
- Slice 20 Espionage/intelligence;
- Slice 21 Colony/Buildings/Planet Effects/Pollution fidelity;
- Slice 22 Preset Race Completion + Custom Race Designer.

Slice 16.7 remains open until its Gate-2 deferral/handoff evidence and Slice-16 closure are completed. This note does **not** open Slice 23 and does not enable `advanced` in the current runtime.

Permanent Gate-2 evidence: docs/research/SLICE_16_7_GATE2_ADVANCED_DEFERRAL_HANDOFF_2026-09-18.md.

## Gate 3 - implementation (deferred to Slice 23; not executed in Slice 16.7)

These implementation tasks are preserved as Slice-23 work and remain intentionally unchecked here. Slice 16.7 does not claim Advanced runtime parity.

- [ ] Enable `advanced` in New Game validation only after the frozen contract is implemented.
- [ ] Implement the complete authoritative Advanced initial-state bootstrap.
- [ ] Reuse the Slice 16.5 Advanced card and switch it from planned/disabled to enabled based on server data.
- [ ] Add deterministic fixtures covering representative races and accepted cross-setting combinations.
- [ ] Verify Save/Resume and subsequent browser play from Advanced starts.
- [ ] Keep API and UI rejection in place for any Advanced combination not in the frozen acceptance matrix.

## Gate 4 - Advanced implementation QA (deferred to Slice 23; not executed in Slice 16.7)

These Advanced-specific QA items move with implementation to Slice 23. Slice-16 closure instead verifies the supported Pre-Warp/Average contract remains intact and Advanced remains planned/locked.

- [ ] Advanced creates the exact frozen authoritative initial state with no silent fallback to Average.
- [ ] Equal seed + complete settings tuple remains deterministic for Advanced.
- [ ] Representative preset races and Slice 16.6 player compositions pass Advanced start/play smoke tests.
- [ ] Desktop and real 390 px mobile New Game flow exposes all accepted starting-tech choices coherently.
- [ ] Full Go tests/vet, Web production build and `git diff --check` pass.
- [ ] Update parent Slice-16 plan, ACTIVE_RESEARCH, PROJECT_STATUS and HISTORY.
- [ ] Close Slice 16 as complete and continue future breadth outside the 16.x numbering series.

## Exit criterion

Slice 16.7 is closed when the Advanced evidence is durable, the current server/UI contract remains honestly planned/locked, and full implementation has an explicit downstream owner rather than a partial fallback. That criterion is satisfied by the Gate-1 audit, Gate-2 deferral contract and reserved Slice 23.

This closure does not claim Advanced runtime parity.
