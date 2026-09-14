# Slice 15.6 Gate 3 - Weapon Slots Block A Contract Amendment

Date: 2026-09-14
Status: implementation block opened and approved by user
Parent plan: `SLICE_15_6_GATE3_WEAPON_SLOTS_TACTICAL_SELECTION_PLAN_2026-09-14.md`

## Amendment

Slice 15.6 Gate 3 deliberately expands the previously frozen military-design weapon baseline from `none | slot 0 Laser Cannon x1` to the following narrow authoritative surface:

- zero to eight stable weapon mounts using slots 0..7;
- currently supported live weapon type remains `laser_cannon` only;
- every mount carries a positive `count`;
- mount rows must be strictly ascending by slot and may not duplicate a slot;
- repeated physical Laser Cannons in one mount use `count > 1` and form one Tactical firing group;
- repeated Laser mounts use distinct slots and remain independently selectable firing groups;
- hull space and design/production cost scale with the total physical weapon count, not with mount-row count;
- technology gating remains authoritative and unchanged: the empire must know Laser Cannon technology 100;
- unsupported weapon types remain rejected.

This amendment intentionally does not broaden hull support, introduce missiles/ammunition, weapon modifiers, firing arcs, component damage, or the general Slice 17 Ship Designer surface.

## Rules acceptance for Block A/B

The server-side military-design contract must prove:

1. existing unarmed and `slot 0 Laser x1` designs remain valid;
2. a single Laser mount with `count > 1` is accepted when hull space permits;
3. multiple Laser mounts with independent slots are accepted when hull space permits;
4. grouped and split mounts consume identical total space/cost when their physical weapon count is equal;
5. invalid slot ids, duplicate/descending slots, non-positive counts, more than eight mounts and unsupported weapon ids reject before state mutation;
6. hull-space overflow rejects without mutating the game state;
7. saved designs preserve exact slot/count identity for later construction/Tactical propagation.

## Next block after this contract/rules foundation

Block C replaces the Ship Builder's current boolean Laser toggle with installed mount rows and per-slot quantity controls. No Tactical UI is reintroduced until those authoritative design semantics are available end to end.
