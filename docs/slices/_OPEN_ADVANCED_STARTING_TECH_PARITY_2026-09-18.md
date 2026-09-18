# Open slice 16.7 - Advanced starting technology parity and Slice-16 closure

Status: **Gate 2 active - Advanced contract freeze**

Opened: 2026-09-18
Session: ses-20260918T115147-24f3d057e055
Baseline commit: 6682f05

Parent: docs/slices/PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md
Plan: docs/slices/PLANNED_16_7_ADVANCED_STARTING_TECH_PARITY.md

## Gate 1 checklist

- [x] Re-audit authoritative/original Advanced start behavior and normalized source IDs.
- [x] Diff Advanced against the frozen pre_warp and average initial-state contracts field by field.
- [x] Enumerate all granted technologies/research state and every non-technology bootstrap delta.
- [x] Inventory race-specific interactions and unsupported-mechanic dependencies.
- [x] Audit interactions with difficulty, galaxy settings and Slice 16.6 player composition.
- [x] Define deterministic fixture evidence needed to prove Advanced parity.

## Permanent Gate-1 evidence

docs/research/SLICE_16_7_GATE1_ADVANCED_PARITY_AUDIT_2026-09-18.md

## Gate-1 result

Advanced is confirmed to require a full bootstrap path, not only the already-implemented 19 additional technology grants. Gate 1 proved the original territory-count formula, variable Advanced population path and Advanced building cap/candidate ordering, and exposed a verified HELP-versus-1.31-runtime conflict for the Advanced starting fleet.

Gate 2 must explicitly freeze territory selection, variable-population jobs, starting Buildings, preference-profile state/generation, the unresolved original 0x21CB0 secondary-filter treatment, and the Advanced fleet parity target.

## Gate 1 guardrail

Gate 1 is evidence/audit only. Advanced remains planned/locked. Do not enable advanced, silently map it to Average, or freeze Gate-2 semantics before Gate 2 is explicitly released.

## Gate 2 checklist

- [ ] Freeze the exact server-owned advanced contract and accepted ID/label.
- [ ] Freeze full empire/population/colony/fleet/research bootstrap semantics.
- [ ] Freeze race-specific exception handling and cross-setting validation.
- [ ] Freeze the server-derived facts shown on the existing Advanced visual card.
- [ ] Freeze the acceptance matrix required before advanced can be enabled for submission.
