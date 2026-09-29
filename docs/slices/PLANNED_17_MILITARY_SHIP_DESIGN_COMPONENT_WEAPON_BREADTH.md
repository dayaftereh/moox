# Planned Slice 17 - Military Ship Design + Tactical Combat Program

Status: **active parent program; children 17.0-17.4 and follow-up 17.2A closed/accepted; Slice 17.5 Gates 1-2 complete / Gate 3 Checkpoint 1 complete / Checkpoint 2 ready; children 17.6-17.9 planned/not open**.

Queue position: **current gameplay program after closed Slice 16**. Slice 18 is reserved downstream for Galaxy/Strategic Navigation fidelity and is not a prerequisite for Slice 17.

## Objective

Expand the intentionally narrow military-design and Tactical baseline into one server-authoritative military-ship foundation: broad legal Ship Designer output, immutable construction snapshots, broad Tactical combatant admission, complete supported Beam/ordnance/fighter/special combat consumers, and end-to-end integration without hard-coded fixture ships.

Slice 17 is a **parent program**, not one monolithic implementation slice. Design breadth, Tactical foundation, weapon-family combat and final integration are deliberately separated into recovery-safe children.

## Binding decomposition

Slice 17 is executed in dependency order as:

1. **17.0 - Ship Designer / Triangle Reference Test Harness**
   - durable development-only reference scenarios;
   - high-tech/all-tech designer laboratories;
   - deterministic construction acceleration using normal turn resolution;
   - direct built-Ship / battle fixture entry for Tactical QA.
2. **17.1 - Hulls and Mandatory Ship Systems**
   - Frigate through Doom Star;
   - drives, computers, armor, shields, fuel;
   - hull/buildability technology and infrastructure restrictions;
   - cost/space/Command-Point and miniaturization foundation.
3. **17.2 - Beam Weapons and Beam Mount Modifiers**
   - Beam design breadth;
   - mount count/slots and applicable Beam modifiers;
   - cost/space/unlock/miniaturization metadata;
   - immutable Beam snapshots and explicit Tactical support/defer metadata;
   - downstream Tactical owner: 17.6.
4. **17.3 - Missiles, Torpedoes and Bomb Ordnance**
   - design-side ordnance representation, ammunition/counts and legal mounting;
   - authoritative snapshots and costs;
   - downstream Tactical projectile/interception/impact owner: 17.7.
5. **17.2A - Beam Weapon Firing Arcs**
   - bounded follow-up to the accepted 17.2 Beam design breadth;
   - original Beam firing-arc identities/defaults/Space/PP and Selector support;
   - immutable arc snapshots handed to downstream Tactical Beam owner 17.6;
   - opens only after 17.3 closes, without reopening 17.2.
6. **17.4 - Ship Specials and Fighter Systems**
   - design-side ship specials, Fighter/Bomber/Interceptor systems and derived ship stats;
   - explicit support classification for every accepted special;
   - downstream fighter/combat-special Tactical owner: 17.8.
7. **17.5 - Tactical Combat Foundation / Ship Combatant Breadth**
   - replace the old Slice-15.5 Frigate/Electronic/Titanium/no-Shield/Standard-Fuel fixture boundary;
   - verified Frigate-through-Doom-Star hull/mandatory-system Tactical values;
   - generic server-owned weapon/action/RNG/damage/event foundation;
   - broad Tactical UI/action/event plumbing without weapon-family-specific formulas.
8. **17.6 - Tactical Beam Weapons & Beam Modifiers**
   - complete supported Beam hit/range/damage/readiness combat;
   - Heavy / Point Defense and supported Beam modifiers;
   - Weapon Selector base-versus-configured/effective preview from the same server rules;
   - authoritative Beam/impact/shield/damage presentation.
9. **17.7 - Tactical Missiles / Torpedoes / Bombs**
   - authoritative ordnance launch/ammo/readiness;
   - projectile lifecycle, interception, impact and damage;
   - Beam Point Defense integration;
   - projectile/interception/explosion presentation and verified bomb combat context.
10. **17.8 - Tactical Fighters / Bombers / Combat Specials**
   - Fighter/Bomber/Interceptor combat actors and carrier readiness;
   - verified combat-special consumers;
   - cross-family defensive/offensive interactions;
   - event-driven fighter/special presentation.
11. **17.9 - Full Ship Designer + Tactical Integration and Acceptance**
   - unified Designer/Construction/Fleet/Encounter/Tactical flow;
   - immutable design/built-Ship identity through combat;
   - persistence/reconnect/replay;
   - mixed weapon-family acceptance on desktop/mobile;
   - final Slice-17 program QA and closeout.

Only one 17.x child may be active unless a checkpoint explicitly documents a justified parallel exception.

## Existing foundation

The program starts from substantial already-implemented infrastructure:

- persisted Empire ShipDesigns;
- built Ships carry immutable ShipDesignSpec snapshots;
- design IDs and revisions survive persistence;
- Colony Construction can build a selected military design;
- Command Points and Combat Fleets already consume built ships;
- normalized six-hull catalog exists;
- mandatory drive/computer/armor/shield/fuel fields already exist in ShipDesignSpec;
- weapon mount snapshots already exist;
- Slice 15.6 already provides the browser Ship Designer shell and Colony Build -> Designer -> Construction handoff;
- Slice 15.5/15.6 provide a playable server-authoritative Tactical Laser baseline with movement, selection, weapon-slot fire, damage events and event-driven beam presentation;
- development-only `-reference-games` registers normal reference games plus the durable Triangle scenario.

## Program architecture rules

- A reference/test harness may alter **initial state**, not gameplay formulas.
- Test scenarios must use the same designer, validation, construction, turn, fleet and battle code paths as ordinary games after bootstrap.
- No debug-only technology or free-build rule may leak into normal New Game/API behavior.
- Built Ship snapshots remain immutable when a design slot is later revised.
- Design legality is server-owned; HMI consumes catalogs and rejection reasons.
- Design-side children through 17.4, including the post-17.3 follow-up 17.2A, define authoritative installed-ship data; Slices 17.5-17.8 consume those exact snapshots in Tactical without translation hacks.
- Tactical hit, RNG, damage, projectile and fighter rules are server-owned. React submits intent and presents authoritative state/events.
- Unsupported weapon-family actions must be explicit and local to that action; an otherwise legal encounter must not be rejected merely because another family is not implemented yet.
- Animation is presentation-only and never controls command legality, RNG, damage or BattleSession progression.
- A design must never be advertised as tactically complete merely because it can be persisted.

## Parent opening guard

Before opening 17.0 Gate 1:

- Git clean;
- zero unrelated `_OPEN_` markers;
- Slice 16 closed;
- Slice 18 remains planned only;
- current reference-game/Triangle scenario green;
- current narrow military design/construction/tactical Laser regressions green.

## Parent completion contract

- [x] 17.0 closed: durable Ship/Combat QA harness exists.
- [x] 17.1 closed: hulls + mandatory systems are fully supported to the frozen design breadth.
- [x] 17.2 closed: Beam design breadth is complete to the frozen contract.
- [x] 17.3 closed: ordnance design breadth is complete to the frozen contract.
- [x] 17.2A closed: supported Beam firing arcs are original-evidenced, server-authoritative and persisted before 17.4 opens.
- [x] 17.4 closed: supported specials/fighter design breadth is complete.
- [ ] 17.5 closed: broad Tactical combatant admission/shared combat foundation replaces the old fixture-only path.
- [ ] 17.6 closed: supported Beam weapons/mounts/modifiers resolve authoritatively and Selector preview matches Tactical rules.
- [ ] 17.7 closed: supported missiles/torpedoes/bombs resolve authoritatively, including projectile/interception/impact behavior.
- [ ] 17.8 closed: supported Fighter/Bomber/Interceptor systems and accepted combat specials resolve authoritatively.
- [ ] 17.9 closed: unified Designer -> Construction -> Fleet -> Encounter -> Tactical -> strategic reconciliation acceptance passes.
- [ ] Every remaining intentionally unsupported mechanic has an explicit owner/reason and is not represented as already working.
- [ ] Final full Go/vet/web/browser/persistence/replay/reference-game QA passes.

## Exit criterion

Slice 17 is complete when a player can construct and persist a broad technology-gated military ship across the frozen hull/component/weapon/special matrix, build that exact immutable revision through normal Colony Construction, deploy it through normal Fleet/Encounter flow, enter Tactical without fixture-only restrictions, use every Beam/ordnance/fighter/special mechanic accepted by the 17.6-17.8 contracts, see the server-authoritative hit/damage/projectile/fighter outcomes presented correctly, complete the battle, and continue strategic play with exact reconciliation and stable persistence.
