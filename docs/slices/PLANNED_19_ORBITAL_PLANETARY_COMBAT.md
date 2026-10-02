# Planned slice 19 - Orbital / Planetary Combat

Status: **reserved / planned; not open**.

Queue position: **reserved future Slice 19 by product decision on 2026-10-02**. Slice 19 follows Slice 18 and precedes Slice 20. Reserving the number does not open the slice or create an `_OPEN_` marker.

## Objective

Add the authoritative orbital/planetary combat layer that connects the mature ship-Tactical runtime to defended colonies and the existing invasion/conquest baseline.

Target end-to-end assault path:

`hostile fleet arrives -> orbital/station battle -> planetary defensive fire/fighters -> optional bombardment -> troop invasion/conquest`

Slice 19 must reuse the combat systems already proven in Slice 17 rather than create a second incompatible combat engine.

## Inputs / dependencies

Gate 1 must re-verify the then-current state of:

- Slice 11 Troop Transport / Invasion / Conquest baseline;
- Slice 17.6 Beam combat;
- Slice 17.7 missile/torpedo projectile combat and the explicit `bombardment_only` Bomb/Bio handoff;
- Slice 17.8 Fighter/Bomber/Boarding/Combat-Special consumers;
- Slice 17.9 final Ship Designer/Tactical integration;
- Slice 18 galaxy/navigation encounter positioning where it affects assault entry;
- the colony/building identities later deepened by Slice 21.

Slice 19 owns combat consumers. It does **not** own the full strategic Building/Economy fidelity work reserved for Slice 21.

## Ownership boundary

In scope, subject to direct original-game evidence and Gate-2 freeze:

- Star Base, Battlestation and Star Fortress as orbital Tactical combatants;
- station deployment, targeting, shields, armor/structure, weapon readiness and destruction;
- reuse of Slice-17 Beam, ordnance, fighter and supported combat-special mechanics by orbital defenses;
- Planetary Missile Base / missile-defense combat consumer;
- Fighter Garrison launch/readiness and reuse of Slice-17.8 fighter actors;
- Ground Batteries and planetary defensive fire where original evidence supports a combat consumer;
- planetary shields and defense modifiers relevant to space assault or bombardment;
- authoritative Bomb/Bio planetary bombardment consumer for the `bombardment_only` weapon snapshots frozen by Slice 17.7;
- transition from defeated orbital defense to bombardment / invasion choices;
- handoff to the existing Slice-11 troop-invasion/conquest resolver;
- persistence, replay, Observer/player projection and browser presentation for active orbital/planetary combat;
- deterministic AI/reference fixtures needed to exercise the supported assault path.

Explicitly out of scope unless Gate 1/2 reassigns ownership:

- redesigning generic ship-Tactical Beam/ordnance/fighter engines already owned by Slice 17;
- ship-vs-ship boarding, Transporters, Troop Pods or Security Stations, which remain Slice 17.8;
- complete Building/Economy/Pollution implementation owned by Slice 21;
- broad tactical ground-combat UI beyond the Slice-11 invasion baseline;
- galaxy-generation/navigation fidelity owned by Slice 18;
- inventing bombardment or planetary-defense formulas without sufficient original evidence.

## Architecture direction

1. **Reuse combat actors and rules.** Orbital stations consume the same authoritative Beam/ordnance contracts as ships where original behavior is equivalent.
2. **Reuse fighters.** Fighter Garrison launches the shared fighter/group actor model frozen in Slice 17.8.
3. **Reuse projectiles.** Planetary Missile Base fire uses the shared projectile lifecycle where applicable.
4. **Keep Colony state authoritative.** Colony/building state owns persistent defenses; Slice 19 owns their combat-time projection and reconciliation.
5. **Do not fake Bomb/Bio damage.** Slice 19 must recover/freeze the original bombardment contract before making `bombardment_only` weapons executable.
6. **Preserve invasion authority.** Actual troop invasion/conquest continues through the existing Slice-11 strategic/session authority.

## Provisional child decomposition

The numbering below is reserved planning structure, not opened work.

### 19.0 - Orbital / Planetary Combat audit and reference harness

- original encounter sequence and participant audit;
- station/base/planet defense identities and stats;
- Bomb/Bio bombardment formulas, ordering and targets;
- transition rules between space battle, bombardment and invasion;
- deterministic reference fixtures.

### 19.1 - Orbital Station Tactical Combatants

- Star Base;
- Battlestation;
- Star Fortress;
- stationary combatant lifecycle and reconciliation;
- Beam/ordnance/special reuse.

### 19.2 - Planetary Defensive Fire and Shields

- Planetary Missile Base;
- Ground Batteries;
- planetary shields / defense modifiers;
- exact launch/readiness/targeting rules.

### 19.3 - Planetary Fighter Operations

- Fighter Garrison;
- launch/readiness/capacity;
- Slice-17.8 fighter actor reuse;
- loss/recovery/reconciliation.

### 19.4 - Planetary Bombardment

- Nuclear / Fusion / Anti-Matter / Neutronium Bombs;
- Death Spores / Bio Terminator where supported;
- population/building/defense/planet damage only according to frozen original evidence;
- bombardment events, persistence and UI.

### 19.5 - Assault -> Bombardment -> Invasion Integration

- post-space-battle decision boundary;
- bombard / invade / disengage legality;
- Troop Transport availability and Slice-11 handoff;
- colony capture and post-combat reconciliation.

### 19.6 - Full Planet Assault Acceptance

- normal-game browser flow;
- save/load/reconnect/replay;
- representative station + missile base + fighter garrison + bombardment + invasion scenarios;
- deterministic AI/reference coverage;
- no duplicate combat formulas across ship, station and planetary consumers.

## Gate 1 - original audit / dependency check

- [ ] verify original station/base/fighter-garrison combat representation;
- [ ] verify encounter ordering and combat contexts;
- [ ] verify Bomb/Bio bombardment formulas, valid targets, defenses, casualties and ordering;
- [ ] verify planetary shields and defensive-building interactions;
- [ ] verify post-battle bombardment/invasion choice ordering;
- [ ] map Slice-21 orbital/defense buildings to Slice-19 combat consumers or explicit deferrals;
- [ ] verify which Slice-17 mechanics are directly reusable;
- [ ] freeze a bounded Gate-2 candidate scope from direct evidence.

## Gate 2 - authority / combat contract freeze

- [ ] freeze combatant/state schemas for stations and planetary defenses;
- [ ] freeze reuse boundaries for Beam, ordnance, fighters and specials;
- [ ] freeze bombardment contract and RNG/order;
- [ ] freeze strategic <-> Tactical <-> invasion handoff;
- [ ] freeze persistence/replay/event/UI contracts;
- [ ] explicitly defer unsupported defense families rather than approximating them.

## Gate 3 - implementation

- [ ] implement the frozen orbital/planetary combat consumers;
- [ ] integrate them through normal encounter/session authority;
- [ ] implement supported bombardment and invasion handoff;
- [ ] add event-driven browser presentation;
- [ ] add deterministic regressions, persistence and replay.

## Gate 4 - QA + close

- [ ] representative defended-colony assaults resolve end to end;
- [ ] station/planet weapons reuse shared combat rules without forked formulas;
- [ ] Bomb/Bio execute only in verified bombardment context;
- [ ] Fighter Garrison uses the shared fighter runtime;
- [ ] bombardment/invasion reconciliation is deterministic;
- [ ] save/load/replay/Observer/browser acceptance passes;
- [ ] Slice-21 combat-consumer dependencies are explicitly documented.

## Exit criterion

Slice 19 closes when an enemy colony with supported orbital/planetary defenses can enter a normal authoritative assault, resolve its station/base/fighter defenses using shared Slice-17 combat mechanics, optionally resolve verified Bomb/Bio bombardment, hand off to the established troop-invasion/conquest path, and reconcile the result through normal session/persistence/browser flows without special fixtures or duplicated combat engines.
