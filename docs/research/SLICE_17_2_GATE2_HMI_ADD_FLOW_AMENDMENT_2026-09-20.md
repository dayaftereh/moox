# Slice 17.2 Gate 2 - HMI Add-Flow Amendment

Date: 2026-09-20
Status: **accepted amendment to frozen Gate-2 HMI contract**

## User-approved interaction change

The initial Beam configuration is completed **inside the Weapon Selector before the weapon is inserted**.

The main Ship Designer therefore keeps the compact installed-weapon list as the post-insertion working view.

### Weapon Selector flow

1. Player presses the weapon `+` action.
2. Weapon Selector opens with the frozen search/filter/sort controls.
3. Player selects a weapon row/card.
4. A configuration area for that selected weapon becomes visible in the same selector.
5. The configuration area shows:
   - mount type: Standard / Heavy / Point Defense, limited by the executable compatibility mask;
   - compatible special modifiers as selectable toggles/checkboxes;
   - compatible-but-refinement-locked modifiers visible but disabled with the required refinement level;
   - quantity, default `1`;
   - server-authoritative preview of unit/row Space and PP and resulting design Space;
   - one explicit `Install / Einbauen` action.
6. Only pressing `Install / Einbauen` adds the configured row to the local design draft and returns to the main Designer.
7. Cancel/back closes or deselects without mutating the installed weapon list.

This replaces the earlier Gate-2 wording where Add immediately inserted a standard count-1 row and the first configuration step happened only after clicking the installed row.

## Installed Weapons after insertion

The fixed-height compact Installed Weapons panel remains.

Each row keeps:

- weapon name;
- concise current configuration summary;
- inline `-` quantity `+` controls;
- icon-only Remove control.

Changing quantity with `-` / `+` uses the same server-authoritative preview/validation pipeline.

A later Edit affordance may reopen the same configuration surface for an already-installed row, but **editing the row is not required to perform the initial configuration**.

## Configuration semantics

Defaults after selecting a weapon:

- quantity = 1;
- mount = Standard;
- special modifiers = none.

Changing weapon selection resets these defaults unless Gate 3 intentionally preserves only values that are legal for the newly selected weapon.

The `Install` action is enabled only when:

- a free slot exists;
- the weapon technology is available;
- the selected mount/modifier combination is legal;
- refinement requirements are met;
- server preview confirms the resulting design fits.

## Scope

This is an HMI workflow amendment only.

The frozen Gate-2 rules remain unchanged:

- 13 Beam identities;
- Standard/Heavy/Point Defense mount model;
- canonical special modifier IDs;
- executable compatibility masks;
- refinement requirements;
- exact Space/PP calculation order;
- server-authoritative preview;
- snapshot representation;
- Tactical support/defer boundary.

**Gate 3 implements this amended add flow.**
