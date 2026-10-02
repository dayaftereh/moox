# OPEN Slice 17.8 - Tactical Fighters / Bombers / Combat Specials

Opened: 2026-10-02

Planned specification:
`docs/slices/PLANNED_17_8_TACTICAL_FIGHTERS_BOMBERS_COMBAT_SPECIALS.md`

Gate-1 evidence:
`docs/research/SLICE_17_8_GATE1_FIGHTERS_SPECIALS_TACTICAL_AUDIT_2026-10-02.md`

## Opening baseline

- branch: `main`
- opening HEAD: `0a3c325 feat(slice): close 17.7 tactical ordnance`
- opening Working Tree: clean
- opening divergence: `0 ahead / 0 behind` versus `origin/main`
- other `_OPEN_*.md` markers at open time: none
- Slice 17.4 Ship Specials / Fighter Systems: closed/accepted
- Slice 17.5 Tactical Combat Foundation: closed/accepted
- Slice 17.6 Tactical Beam Weapons / Modifiers: closed/accepted
- Slice 17.7 Tactical Missiles / Torpedoes / Bombs: closed/accepted

## Current gate

**Gate 1 - fighter/special Tactical audit: IN PROGRESS**

No Slice-17.8 production implementation is allowed before Gate 1 closes and Gate 2 freezes the supported fighter/special combat contract.

## Gate-1 audit threads

- Fighter Bays / Bomber Bays / Heavy Fighter Bays identity, craft-per-bay, launch, readiness and loss/recovery semantics.
- Fighter/Bomber/Interceptor actor/group representation, movement, targeting, attack and return.
- Bomber target eligibility and damage context.
- Beam / Point Defense / ordnance / defensive-system interactions with fighter actors.
- Full reclassification of every 17.4 Class-3 Special into strategic-only, Tactical-common, Beam-related, ordnance-related, fighter-related, boarding-related or unsupported/deferred.
- Explicit boarding audit for Transporters, Troop Pods, Security Stations, Marine state, repel/capture flow and ownership change.
- Repair/regeneration, cloak/phasing/phase-shift, ECM/jamming/anti-missile, shield/damage-field, teleport/time/area and remaining combat-special effects.
- BattleSession persistence/replay/Observer requirements for launched craft and temporary special state.
- Tactical UI/action/event/presentation requirements.
- Deterministic original/reference vectors for representative carrier, boarding and special-effect cases.

## Scope guard

- Do not reopen 17.6 Beam or 17.7 ordnance formulas unless Gate-1 evidence demonstrates a required shared hook.
- Do not make metadata-only 17.4 rows executable merely because they are Designer-legal.
- Do not infer browser-side fighter physics or special effects.
- Keep strategic-only effects strategic.
- Unsupported/unproven behavior stays explicit rather than approximated.
- Planetary Fighter Garrison / orbital-planetary combat remains reserved for Slice 19; Slice 17.8 builds the reusable ship-Tactical fighter/combat-special runtime.

## Next

Continue Gate 1 with direct original-evidence review and current-runtime inventory, then produce the complete support/effect/ordering matrix and deterministic reference vectors before any Gate-2 freeze.


## Gate-1 progress - Checkpoint A

Direct original/runtime audit now establishes:

- original fighters share substantial missile/moving-actor infrastructure; Fighter Beam can target ships and moving missile/fighter actors through shared damage/retarget/kill paths;
- MOOX already has MarineMax/MarineCurrent and immutable Special metadata, while active fighter lifecycle state is still absent;
- normal boarding range is 1; Transporters extend it to 4 and add shield/Hard-Shield restrictions;
- the original boarding path has explicit Marine-party selection/mutation, Already-Boarded/Teleporter-used temporary flags, crew/security modifiers, Raid/Capture branches and Quantum-Detonator capture handling;
- direct original Tactical consumers are located for repair, Shield Capacitors/recharge, Anti-Missile Rockets, Lightning Field, Energy Absorber, teleporter effects, Reflection/Damper fields and boarding/security systems;
- architectural direction is to reuse/factor the authoritative Slice-17.7 moving-actor foundation for fighter groups while keeping fighter payload/lifecycle state distinct.

Gate 1 remains **IN PROGRESS**. No production implementation is authorized yet.

Next audit focus:

1. fighter launch/move/return/reload state transitions and exact payload/durability/speed vectors;
2. exact Raid/Capture casualty arithmetic and ownership reconciliation;
3. full per-Class-3-Special effect/support matrix and exact formulas/order;
4. persistence/event/UI/AI contracts and deterministic reference vectors.
