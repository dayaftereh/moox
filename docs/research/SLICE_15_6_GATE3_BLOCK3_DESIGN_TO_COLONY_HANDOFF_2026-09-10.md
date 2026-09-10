# Slice 15.6 Gate 3 - Block 3 Ship Design to Colony handoff

Date: **2026-09-10**

Status: **implemented and focused browser verified**.

## Goal

Verify and finish the visible browser handoff from a newly saved named Ship Design into Colony Construction, while keeping the block limited to navigation, identity/revision preservation and queue planning UX.

## Isolated browser scenario

Temporary QA server: **127.0.0.1:7181**. Canonical 7171 remained untouched.

Fresh game:

- `designer-b3`;
- seed `0x8009`;
- Human local human vs Darlok Built-in AI;
- Turn 1 / Planning.

Through visible browser controls only, the test created and saved:

- name: `Falcon Mk I`;
- hull: Frigate;
- weapon: `1x Laser Cannon`;
- Production Cost: **30 PP**;
- Command Points: **1 CP**;
- design space: **10 / 25**.

## Colony construction catalog

Using normal bottom navigation:

1. Ship Designer -> Colonies;
2. open Colony 53;
3. open Construction Manager.

The build catalog displayed separate authoritative choices:

- `Scout` / Military Ship / 25 PP;
- `Falcon Mk I` / Military Ship / 30 PP.

Selecting Falcon displayed:

- selected project: `Falcon Mk I`;
- kind: Military Ship;
- Production Cost: 30 PP;
- Ship Design: `Falcon Mk I`;
- Revision: `r1`.

Clicking **Einplanen** created the normal Planning draft and authoritative planning preview. The visible queue then showed:

- current project: `Falcon Mk I`;
- progress: 0.0 / 30 PP;
- ETA: 5 turns;
- selected-project facts remained `Falcon Mk I`, r1, 30 PP.

This proves the visible design-name and exact-revision handoff through the normal Colony construction planning path.

## Direct Colony -> Ship Designer handoff

The existing `Schiff entwerfen` button was still a disabled placeholder with obsolete text saying the interactive Ship Designer would arrive in a later slice.

It is now active. For Military Ship construction choices it navigates to the Ship Designer using the selected `ship_design_id` as the route entity ID.

Verified URL for the saved Falcon:

`#/game/designer-b3/shipbuilder/80`

The Ship Designer accepts the optional route design ID and starts directly on the matching saved design. Browser verification showed:

- selected library item: `Falcon Mk I / Fregatte / r1`;
- name field: `Falcon Mk I`;
- Production: 30 PP;
- Command Points: 1 CP;
- design space: 10 / 25;
- no horizontal overflow at the actual 776 CSS-px managed-browser viewport.

The obsolete placeholder note was replaced by guidance that the button opens the exact named design and that saved revisions return as authoritative Colony build choices.

## Validation

- `npm run build`: PASS after the route/prop wiring fix;
- route/entity handoff verified in managed Chrome 152;
- Colony catalog -> selected project -> planned queue verified through visible browser controls;
- temporary Chrome session closed;
- temporary 7181 server closed;
- canonical 7171 service not restarted or mutated.

## Next

After checkpoint/recovery, the next small block can refresh the canonical 7171 review runtime to the latest committed code and create a fresh review game so the user can directly inspect the new Ship Designer. After review, continue with the canonical seed Human-vs-Built-in-AI vertical-slice dry run rather than broadening the Designer further without evidence.
