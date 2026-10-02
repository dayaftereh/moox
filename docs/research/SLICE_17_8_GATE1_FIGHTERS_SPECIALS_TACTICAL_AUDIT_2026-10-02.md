# Slice 17.8 Gate 1 - Fighter / Bomber / Combat-Special Tactical Audit

Date: 2026-10-02

Status: **IN PROGRESS**

## 1. Purpose

Re-audit the original Tactical behavior required to turn the authoritative Slice-17.4 Fighter/Bomber/Heavy-Fighter weapon rows and deferred Class-3 Special snapshots into server-authoritative Tactical mechanics.

This Gate does not implement product behavior. It establishes the exact supported matrix, ownership boundaries, interaction order, persistence/event requirements and deterministic reference vectors required for Gate 2.

## 2. Verified dependency baseline at open

Repository/opening state:

- branch `main`;
- opening HEAD `0a3c325 feat(slice): close 17.7 tactical ordnance`;
- Working Tree clean;
- `HEAD...origin/main = 0/0`;
- no pre-existing `_OPEN_*.md` marker.

Required predecessor state is satisfied:

- Slice 17.4 Ship Specials / Fighter Systems is closed/accepted and owns the Designer/persistence snapshot contract;
- Slice 17.5 Tactical Combat Foundation is closed/accepted;
- Slice 17.6 Tactical Beam Weapons / Modifiers is closed/accepted;
- Slice 17.7 Tactical Missiles / Torpedoes / Bombs is closed/accepted;
- immutable fighter/special snapshots already reach built Ship/BattleSession without pretending deferred Tactical consumers are implemented.

## 3. Frozen 17.4 handoff that Gate 1 must consume, not reinterpret

### 3.1 Fighter weapon rows

Slice 17.4 froze three normal Weapon rows, each at **4 craft per Bay**, with Tactical behavior deliberately `metadata_only` until 17.8:

- `heavy_fighter_bays`;
- `bomber_bays`;
- `fighter_bays`.

They remain Weapon rows in `ShipDesignSpec.Weapons`; 17.8 must not convert them into Special slots.

### 3.2 Special support classes

Slice 17.4 froze original selectable Special rows 1..38:

- Class 1 already complete as design/strategic effects: `battle_pods`, `extended_fuel_tanks`;
- Class 2 already complete for their existing Tactical consumer: `inertial_nullifier`, `inertial_stabilizer`;
- Class 3: every other selectable Special row 1..38; design representation complete, Tactical effect deferred to 17.8;
- original regeneration row 39 remains locked/non-normal-Designer evidence and must not be silently promoted into a selectable Special.

Gate 1 must preserve those ownership decisions while determining which Class-3 rows actually have Tactical, strategic-only or unsupported consumers.

### 3.3 Explicit Slice-17.8 ownership handed off by the frozen 17.4 contract

The 17.4 Gate-2 contract assigns 17.8 responsibility for auditing/freezing:

- fighter launch / movement / targeting / interception / fire / return / reload / damage;
- repair/regeneration combat behavior;
- cloak / phasing / phase-shift behavior;
- boarding / transport / security actions;
- ECM / jamming / anti-missile Special effects;
- shield / damage-field Special effects;
- active/time/teleport/area/combat-special actions;
- Beam-support Special resolution not already owned by Slice 17.6.

## 4. Gate-1 work packages

### A - Fighter/Bomber/Interceptor actor audit

Determine from direct original evidence:

- whether craft are represented individually or in groups/squadrons;
- exact craft count per launched Bay and how multiple Bays aggregate;
- launch timing/action cost/readiness;
- initial position and movement model;
- target eligibility and retargeting;
- attack cadence, hit/damage semantics and RNG;
- Fighter vs Bomber vs Heavy Fighter role differences;
- loss accounting, return/recovery/reload and carrier-destruction consequences;
- whether/how fighter actors can be attacked or intercepted by Beams, Point Defense, ordnance and defensive Specials.

### B - Bomber context audit

Freeze:

- ship/planet/installation target eligibility;
- damage model and defenses;
- relation to future planetary Fighter Garrison / bombardment work;
- explicit boundary with Slice 19 Orbital / Planetary Combat.

### C - Combat-Special classification matrix

Re-audit every Class-3 selectable Special and classify it as one or more of:

- strategic-only;
- Tactical-common;
- Beam-related;
- ordnance-related;
- fighter-related;
- boarding-related;
- unsupported/deferred.

For every accepted Tactical consumer, recover exact effect, timing/order, target scope, stacking/exclusion and RNG/rounding behavior.

### D - Boarding / Marine audit

Explicitly cover:

- `transporters`;
- `troop_pods`;
- `security_stations`;
- existing `MarineMax` / `MarineCurrent` state and any already-frozen Troop-Pod capacity effect;
- boarding eligibility/range/action timing;
- attacker/defender Marine casualty resolution;
- Security Stations / defenses;
- capture versus repel outcomes;
- ship owner/control transfer and post-capture readiness/state;
- persistence/replay/session reconciliation and AI/UI requirements.

Do not reuse the strategic planet-invasion resolver unless direct evidence establishes a shared formula; ship boarding and planetary invasion remain distinct domains.

### E - Defensive/active Special interactions

Audit at least the frozen 17.4 families that may alter existing 17.6/17.7/fighter mechanics:

- Automated Repair Unit / regeneration-related combat lifecycle;
- Battle Scanner;
- Cloaking Device / Phasing Cloak / Stealth Field;
- Damper Field / Hard Shields / Multi-Phased Shields / Shield Capacitors;
- Displacement Device / Reflection Field / Energy Absorber;
- ECM Jammer / Multi-Wave ECM Jammer / Wide Area Jammer;
- Fast Missile Racks / Lightning Field;
- High Energy Focus / Hyper-X Capacitors / Rangemaster Unit / Structural Analyzer / Achilles Targeting Unit;
- Phase Shifter / Sub-Space Teleporter / Time Warp Facilitator;
- Quantum Detonator / Warp Dissipater;
- Reinforced Hull and other passive survivability rows where current Tactical snapshots already expose their owned effect.

Each row must be assigned to existing shared authority, a new 17.8 consumer, strategic-only behavior or explicit deferral.

### F - Runtime/persistence/UI audit

Determine:

- authoritative temporary fighter/special state schema;
- command/action projection;
- event vocabulary;
- Observer/player-safe projection;
- save/load/reconnect/replay requirements;
- built-in AI decision hooks;
- event-driven browser presentation with no client-side combat authority.

## 5. Current implementation boundary at Gate-1 open

At open time:

- fighter/bomber/heavy-fighter rows are design-valid but Tactical metadata-only;
- no new 17.8 fighter actor lifecycle is accepted yet;
- no Class-3 Special is accepted as tactically working solely by Designer legality;
- 17.6 Beam and 17.7 ordnance remain authoritative shared foundations;
- Slice 19 owns future planetary/orbital combat consumers such as Fighter Garrison, but should reuse the fighter runtime frozen here rather than invent a second fighter engine.

## 6. Gate-1 completion criteria

Gate 1 may close only when:

- every supported fighter/bomber/heavy-fighter identity and lifecycle decision is backed by direct evidence;
- all Class-3 Specials have an explicit support/ownership classification;
- every accepted Tactical Special has exact effect/order/RNG/target semantics;
- boarding/Marine semantics are either frozen from evidence or explicitly deferred;
- persistence/replay/Observer/UI/AI requirements are mapped;
- deterministic reference vectors exist for representative fighter, bomber, boarding and special-effect cases;
- no unsupported behavior is being guessed.

## 7. Immediate next audit actions

1. inventory all current fighter/special snapshots and dormant Tactical fields/hooks;
2. re-open the direct original evidence/disassembly sources used during Slice 17.4 and trace fighter/combat-special consumers;
3. map existing 17.6 Beam and 17.7 ordnance hooks that 17.8 can reuse;
4. recover exact fighter craft stats/lifecycle and boarding/Marine formulas;
5. build the full per-row support/effect matrix;
6. freeze deterministic vectors only after the above evidence converges.
