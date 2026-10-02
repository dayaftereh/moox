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


## 8. Checkpoint A - current MOOX runtime inventory

### 8.1 Fighter rows are still metadata-only

The accepted ruleset currently carries:

- `heavy_fighter_bays` - original weapon row 29, kind `fighter`, 4 craft per unit, `tactical_support=metadata_only`;
- `bomber_bays` - original weapon row 30, kind `fighter`, 4 craft per unit, `tactical_support=metadata_only`;
- `fighter_bays` - original weapon row 31, kind `fighter`, 4 craft per unit, `tactical_support=metadata_only`.

This is consistent with the Slice-17.4 freeze: Bay count is stored in the normal Weapon mount count and no launch/runtime claim exists yet.

### 8.2 Tactical ship state already has the boarding foundation

Current `battle.TacticalShipSpec` already projects:

- `Specials []TacticalSpecialSpec` with canonical `SpecialID` and `TacticalSupport`;
- `MarineMax`.

Current `battle.TacticalShipState` already persists:

- `MarineCurrent`.

This means Slice 17.8 does not need to invent Marine capacity/state transport. It does need commands, legality, casualty/capture resolution, temporary per-battle flags and reconciliation.

### 8.3 Slice-17.7 actor infrastructure is a candidate shared foundation

Current Slice-17.7 runtime owns a persistent `TacticalOrdnanceState` containing `Projectiles []TacticalProjectileState`. Projectile state already has:

- stable actor identity;
- source seat/ship/weapon slot;
- target ship;
- position/facing/speed/movement remainder;
- multiplicity;
- per-unit durability/intercept damage remainder;
- deterministic persistence/replay/projection.

Gate 1 does **not** yet freeze fighters as `TacticalProjectileState`; fighter state needs additional lifecycle fields. However, the original-code audit below demonstrates that original fighters reuse substantial missile-actor infrastructure, so Gate 2 should prefer a shared moving-combat-actor layer or a fighter extension over a second unrelated flight simulation.

## 9. Checkpoint A - direct original 1.31 Fighter evidence

Audit source: local `C:\ASH\Installation\mastori2\Orion2.exe`, using the same LE-object/symbol/disassembly tooling used by prior Slice-17 audits. Audit helpers are kept outside the repository under `C:\ASH\Temp\moox-s178-g1-*`.

### 9.1 Original fighter functions recovered

Direct symbols include:

- `Check_For_Launched_Fighters_` at `0x019480`;
- `Fire_Fighter_Bomb_` at `0x02AC20`;
- `Fire_Fighter_Beam_` at `0x02AD57`;
- `Fighter_Ocv_` at `0x02DF8D`;
- `Draw_Fighter_` at `0x02DBD1`;
- `Fighter_Combat_SFX_` at `0x02E1A4`;
- `Draw_Fighter_Beam_` at `0x099033`.

The earlier 17.4 original audit remains valid for carrier payload identity/economics: 4 craft per Bay; Fighter, Bomber and Heavy Fighter are distinct original Weapon rows.

### 9.2 Fighters share the original moving-missile actor substrate

`Check_For_Launched_Fighters_` iterates the original combat moving-actor array in 0x1A-byte records and identifies records whose underlying weapon/type table row has class byte `4`.

More importantly, `Fire_Fighter_Beam_` directly calls original missile-actor helpers:

- `Missile_Dcv_`;
- `Apply_Damage_To_Missile_`;
- `Is_Missile_Target_Alive_`;
- `Retarget_Missile_`;
- `Kill_Missile_`.

Therefore original Fighter Beam fire can engage both ship targets and compatible moving missile/fighter actors through shared combat-actor code. This is direct architectural evidence against implementing MOOX fighters as an unrelated client-side or isolated simulation.

### 9.3 Fighter groups carry multiplicity

`Fire_Fighter_Bomb_` loops against the launched actor record's count field at offset `+0x0F`. The routine performs per-member RNG/damage work and calls `Apply_Ship_Damage_`.

This aligns with the previously audited 4-craft-per-Bay semantics and supports an authoritative group actor with explicit remaining multiplicity rather than four browser-only sprites.

### 9.4 Fighter Beam accuracy uses the normal combat offensive stack

`Fighter_Ocv_` directly composes:

- `Fleet_Offensive_Bonus_`;
- the best-computer offensive value;
- a player/crew-dependent bonus for the relevant case;
- `Get_Fleet_Pilot_Bonus_`;
- a fixed +50 baseline.

`Fire_Fighter_Beam_` then compares that result against ship Beam Defense for ship targets, or against `Missile_Dcv_` for moving-actor targets, and consumes the normal original RNG/damage paths.

Exact signed/rounding semantics for every Fighter/Heavy-Fighter Beam shot remain to be converted into deterministic Gate-2 vectors before Gate 1 closes.

### 9.5 Inherited payload facts still requiring lifecycle confirmation

The 17.4 direct HELP/table audit established:

- normal Fighters: best point-defense-capable Beam, four firings before return;
- Bombers: one best available bomb, speed 10;
- Heavy Fighters: two best point-defense Beams plus two best bombs, 5 base hits modified by best armor, speed 12.

Gate 1 now needs to trace the launch/move/return/reload state transitions that consume those payloads. No separate named `Launch_Fighter_` symbol exists; this further suggests launch/movement is embedded in the common missile/moving-actor path.

## 10. Checkpoint A - direct original 1.31 Boarding evidence

### 10.1 Original boarding/capture function graph recovered

Direct symbols include:

- `Initiate_Boarding_Action_` at `0x0359E7`;
- `Init_Boarding_Action_` at `0x0272BD`;
- `Resolve_Raid_` at `0x027308`;
- `Resolve_Capture_` at `0x027DA8`;
- `Capture_Ship_` at `0x028312`;
- `Select_Raid_Internal_` at `0x0284A8`;
- `Resolve_Ai_Boarding_` at `0x01BF73`;
- `Crew_Boarding_Combat_Bonus_` at `0x025CAD`;
- `Get_Fleet_Security_Bonus_` / `Security_Bonus_`;
- `Used_Teleporter_` / `Set_Teleporter_Used_`;
- `Already_Boarded_` / `Set_Boarded_Flag_`;
- `Being_Captured_`.

Boarding is therefore an explicit Tactical subsystem, not a reuse of the strategic planet-invasion resolver.

### 10.2 Normal boarding versus Transporters

`Initiate_Boarding_Action_` checks Special ID 35 (`transporters`):

- without Transporters, permitted boarding range is 1;
- with Transporters, permitted range is 4.

For the non-Transporter path, the original code additionally calls `Calc_Current_Speed_` and rejects a target that is still moving. This supports the classic disabled/immobile-target requirement for physical boarding.

The Transporter path bypasses that immobility requirement but has additional shield/special restrictions.

### 10.3 Marine availability and selected boarding force

The combat ship record stores current Marines in the byte used by `Init_Boarding_Action_`. The original popup maximum is calculated as:

`(MarineCurrent + 1) / 2`

before the player selects the boarding party. `Initiate_Boarding_Action_` subtracts the selected Marine count from the attacking ship before resolution.

Gate 2 therefore needs an explicit boarding-party count in the command/event contract and must treat MarineCurrent mutation atomically with raid/capture resolution.

### 10.4 Transporter shield restrictions

For Transporter boarding, the original action path:

- resolves relative bearing and the target's facing Shield arc;
- rejects when the relevant Shield arc remains active;
- explicitly checks Special ID 13 (`hard_shields`) and rejects Transporter boarding when present.

This matches the original technology text previously normalized into repository rules and is now backed by a direct Tactical call path.

### 10.5 Damper Field interaction exists in the boarding path

The same Transporter branch explicitly checks Special ID 7 (`damper_field`). In the evidenced branch, the selected boarding Marine count is integer-divided by 3 before resolution.

The surrounding branch conditions still need final semantic labeling in the Gate-2 vector (distance/shield-state context) before freezing the exact rule, but the divisor and direct Damper-Field consumer are no longer speculative.

### 10.6 Boarding has per-combat usage flags

The original combat-ship bit field exposes explicit helpers:

- `Used_Teleporter_` / `Set_Teleporter_Used_`;
- `Already_Boarded_` / `Set_Boarded_Flag_`;
- `Being_Captured_`.

MOOX therefore needs persisted/replay-safe temporary boarding flags; `MarineCurrent` alone is insufficient.

### 10.7 Raid/capture consumes crew/security modifiers and RNG

`Resolve_Raid_` directly calls:

- `Crew_Boarding_Combat_Bonus_`;
- `Get_Fleet_Security_Bonus_`;
- `Random_`.

The capture path also consumes fleet security state before `Capture_Ship_`.

Exact casualty arithmetic remains open, but Gate 1 has now established the authoritative dependency set: Marine counts, crew boarding combat bonus, fleet/security bonus and original deterministic RNG order.

### 10.8 Quantum Detonator is a capture hook

`Capture_Ship_` checks Special ID 24 (`quantum_detonator`), calls `Random_(10)`, and destroys the ship when the result is below 6.

That is a direct 6/10 capture-time self-destruction branch. Gate 2 must freeze its RNG position relative to capture ownership/reconciliation and event emission.

## 11. Checkpoint A - direct combat-Special consumer anchors

The symbol/disassembly audit has recovered direct original Tactical consumers for several previously metadata-only families:

- Automated Repair Unit: `Repair_All_Combat_Ships_` checks Special ID 3 and feeds `Repair_Combat_Ship_`; fleet Engineer bonus participates in the repair amount.
- Shield Capacitors: `Recharge_Shields_` checks Special ID 30.
- Anti-Missile Rockets: dedicated `Anti_Missile_Rockets_` combat function.
- Lightning Field: dedicated `Lightning_Field_` combat function.
- Energy Absorber: dedicated `Energy_Absorber_Attack_` plus display path.
- Sub-Space Teleporter/teleport effects: `Teleport_Enemy_Ship_` selects a legal destination through combat-grid/legal-move authority and then updates ship coordinates.
- Reflection Field: dedicated reflection-field hit/display subsystem exists.
- Damper Field: direct `Has_Damper_Field_` plus the evidenced boarding interaction above.
- Cloak/phasing state is directly manipulated during boarding and has a dedicated `Draw_Cloak_` presentation path.
- Security Stations/Transporters/Troop Pods are part of the explicit boarding subsystem rather than passive metadata.

This does not yet mean every formula is frozen. It establishes that these systems have original Tactical consumers that Slice 17.8 must either implement exactly or explicitly defer after the remaining disassembly audit.

## 12. Preliminary architecture consequences

The evidence currently supports these Gate-2 directions, subject to completion of Gate 1:

1. **Shared moving actor substrate:** factor/reuse the 17.7 authoritative moving-actor/projectile lifecycle for fighter groups rather than creating independent browser physics.
2. **Distinct fighter state:** fighters need payload/shot/return/role fields beyond the existing missile `TacticalProjectileState`; do not overload ordnance semantics blindly.
3. **Group multiplicity is authoritative:** Bay payloads should result in server-owned group multiplicity/losses.
4. **Boarding is a first-class Tactical command family:** command legality must cover range/mobility/shields/specials/usage flags/Marine count and resolve atomically.
5. **Temporary Special state must persist:** teleporter/boarded/capture and other per-round/per-battle states belong in save/replay-authoritative state.
6. **Existing 17.6/17.7 resolvers stay shared:** Fighter Beam versus ships and moving actors should use frozen Beam/defense/damage semantics where the original path demonstrably shares them.
7. **Slice 19 reuse boundary remains intact:** planetary Fighter Garrison should consume the 17.8 fighter runtime later; 17.8 does not implement planetary combat.

## 13. Remaining Gate-1 blockers after Checkpoint A

Still required before Gate 1 can close:

- trace fighter launch/movement/return/reload state transitions in the common moving-actor path;
- freeze Fighter/Bomber/Heavy-Fighter speed, durability, payload and firing-count fields from direct binary/table evidence into deterministic vectors;
- finish `Resolve_Raid_` / `Resolve_Capture_` casualty arithmetic and capture ownership/reconciliation ordering;
- map Security Stations, crew/fleet bonus exact numeric tables and Troop-Pod interaction into boarding vectors;
- complete the per-row Class-3 Special support/effect matrix;
- recover exact formulas/order for repair, shield recharge, anti-missile rockets, Lightning Field, Energy Absorber, Reflection/Damper fields, cloak/phasing, teleporter/time effects and Beam-support Specials;
- map command/event/projection/UI and built-in-AI requirements;
- define persistence migration for new fighter/special temporary state;
- produce deterministic reference vectors for fighter, bomber, boarding and representative active/passive Special cases.
