# OPEN - Slice 17.2 Beam Weapons and Beam Mount Modifiers

Opened: 2026-09-20
Parent: Slice 17 - Military Ship Design Component/Weapon Breadth
Status: **Gate 1 complete / Gate 2 next**

## Objective

Expand the accepted Slice-17.1 military Designer from the current standard Laser baseline into an original-evidenced Beam-weapon catalog with authoritative technology, damage, Space, Cost, miniaturization, mount/count and legal modifier rules.

## Gate 1 active scope

- inventory all normalized original Beam weapon identities required by the supported MOO2 design catalog;
- map each Beam to technology unlock, base damage, Space, Cost and range/dissipation metadata;
- re-audit original miniaturization behavior for Beam Space/Cost;
- audit weapon-slot/count semantics and current `ShipWeaponMount` representation;
- audit modifier legality/compatibility per Beam, including Point Defense, Heavy Mount, Autofire, Armor Piercing, Continuous and No Range Dissipation only where primary/original evidence supports them;
- distinguish design metadata from Tactical behavior already implemented versus explicitly deferred;
- define Baseline/Mid/All-Tech acceptance vectors using the durable Slice-17.0 reference labs;
- preserve construction/persistence/revision snapshot authority from closed Slice 17.1.

## HMI direction captured from user review

The main Ship Designer should remain compact.

Preferred interaction model:

- installed weapons remain a compact stable list;
- broad weapon selection opens a dedicated selection dialog/sheet rather than permanently expanding the main page;
- the selection dialog may support searchable/filterable weapon rows/cards with the relevant damage, Space, Cost, availability and modifier information;
- clicking an installed weapon should become the natural future entry point for weapon-specific configuration/modifiers;
- the existing `+` action in Available is the current hook for opening that selector once the Gate-2 contract is frozen;
- Special Systems use the same high-level interaction grammar later, but their authoritative catalog/effects remain owned by Slice 17.4 and must not be pulled into 17.2.

This records the UX direction only. Gate 1 must not invent Beam identities/modifier compatibility or implement broad weapon rules before the audit and Gate-2 freeze.

## Guardrail

Slice 17.1 is closed/accepted and remains the foundation.

Gate 1 is complete. Do not implement the complete Beam matrix or modifier behavior until Gate 2 is explicitly started and freezes the supported contract.

Do not open Slice 17.3 or later children while Slice 17.2 is open unless explicitly requested.

## Gate 1 completion target

Gate 1 may close only when the repository contains a durable evidence document that resolves:

- supported Beam identity matrix;
- technology ownership;
- damage/range metadata;
- base Space/Cost;
- miniaturization evidence;
- slot/count representation;
- modifier compatibility matrix;
- current Tactical consumers versus deferred behavior;
- reference-lab vectors needed for Gate 2/3.

## Gate 1 complete

Evidence: docs/research/SLICE_17_2_GATE1_BEAM_WEAPONS_AUDIT_2026-09-20.md.

Gate 1 froze the evidence inputs, not the implementation contract: 13 original Beam identities, exact base tech/damage/Space/Cost, Beam miniaturization class, per-Beam executable modifier masks, original eight-row weapon-array semantics, the current ShipWeaponMount modifier-snapshot gap, Slice-17.0 reference vectors and explicit Tactical support/defer boundaries.

Gate 2 is next but has not started.