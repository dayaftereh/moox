# OPEN - Slice 17.2 Beam Weapons and Beam Mount Modifiers

Opened: 2026-09-20
Parent: Slice 17 - Military Ship Design Component/Weapon Breadth
Status: **Gate 4 review active - automated QA passed / manual visual review pending**

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

Gate 2 is complete. Gate 3 is complete. Gate 4 is active; automated QA passed and manual canonical-7171 visual review is pending.
## Gate 2 complete

Contract: docs/research/SLICE_17_2_GATE2_BEAM_CONTRACT_FREEZE_2026-09-20.md.

Gate 2 freezes the 13-Beam catalog, canonical standard/heavy/point-defense mount type, explicit sorted special-modifier IDs, exact legality/refinement/rejection reasons, original mount->special->miniaturization->count Space/PP arithmetic, server-authoritative draft preview boundary, compact Weapon Selector/configuration dialog contract, reference-lab matrices and Tactical support/defer boundary.

Gate 3 is complete. Gate 4 is active; automated QA passed and manual canonical-7171 visual review is pending.

## Gate 3 HMI add-flow amendment

Initial weapon selection and Standard/Heavy/Point Defense + special-modifier configuration now happen inside the Weapon Selector before pressing Install. Installed rows retain compact +/- quantity controls. Evidence: docs/research/SLICE_17_2_GATE2_HMI_ADD_FLOW_AMENDMENT_2026-09-20.md.


## Gate 3 complete

Implementation: docs/research/SLICE_17_2_GATE3_IMPLEMENTATION_2026-09-21.md.

All 13 frozen Beams now flow through authoritative Designer selection/preview/save, immutable construction/built-Ship snapshots and lossless BattleSession metadata. The amended selector-first HMI is implemented. The Field-74 Antaran self-loop is handled terminally at M0. Non-baseline Tactical effects remain explicitly deferred; only the unmodified Standard Laser receives firing behavior. Gate 4 review is active; automated QA passed and manual canonical-7171 visual review is pending.


## Gate 3 mobile HMI amendment

Follow-up evidence: docs/research/SLICE_17_2_GATE3_MOBILE_WEAPON_SELECTOR_AMENDMENT_2026-09-21.md.

The redundant Available Weapons card is removed from the main Designer, Installed Weapons now owns the Add entry point and Space Available display, filters default collapsed, and the mobile selector keeps the weapon list/configuration primary with Install visible at 390x844. Gate 4 review is active; automated QA passed and manual canonical-7171 visual review is pending.


## Gate 3 weapon-slot/edit HMI amendment

Follow-up evidence: docs/research/SLICE_17_2_GATE3_WEAPON_SLOT_EDIT_AMENDMENT_2026-09-22.md.

Installed weapons now use fixed-height slot cards with inline quantity controls and an Edit pencil. Edit reuses the authoritative Weapon Selector, preloads the existing slot and replaces that slot rather than duplicating it. Mount/modifier choices expose compact `?` help with frozen Space/PP/refinement information and explicit Tactical defer scope. The blank installed-weapons area remains an Add target when no weapon is installed. Full Web/browser/Go QA passes.

Gate 4 review is active; automated QA passed and manual canonical-7171 visual review is pending.

## Gate 4 Weapon Selector review

Evidence: docs/research/SLICE_17_2_GATE4_WEAPON_SELECTOR_REVIEW_2026-09-22.md.

Automated Gate-4 QA passes after the mobile selector polish: selected mount/modifier cards no longer duplicate their state with visible radio/checkbox glyphs, help controls have reserved non-overlapping space, damage is explicitly labelled Basisschaden/Base damage, and Heavy plus Point Defense both prove live authoritative Space/PP preview changes. Quantity +/- is now owned only by installed weapon cards: new installs enter as count 1, Edit preserves the existing count, and the browser regression proves 1 -> 2 on-card then mount/modifier editing keeps 2. Keep this marker until the user completes the canonical-7171 visual review.
