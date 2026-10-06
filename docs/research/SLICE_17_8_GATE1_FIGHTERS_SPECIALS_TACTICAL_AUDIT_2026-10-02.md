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
- the normal-player race Ship Attack byte (`player+0x8A6`) for the relevant case;
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

## 13. Checkpoint B - Fighter return/recovery and Boarding resolution arithmetic

This checkpoint continues the direct Orion2.exe audit from Checkpoint A. It closes the previously open fighter return/recovery path and resolves the core Raid/Capture casualty machinery far enough to constrain the Gate-2 state machine.

### 13.1 Launched fighters remain authoritative carrier state

`Is_Ship_Out_Of_Ammo_` has a dedicated Fighter-class branch. When the carrier's Fighter weapon row is otherwise exhausted, it calls `Check_For_Launched_Fighters_`.

`Check_For_Launched_Fighters_` scans the common 300-entry moving-actor array, matches records whose source ship is the carrier and whose original weapon class is Fighter (`class = 4`).

Therefore a carrier is **not** considered out of ammunition while one of its launched Fighter-class actors still exists. Launched groups are persistent authoritative combat state, not instant visual effects.

### 13.2 Exhausted Fighter groups explicitly return to the carrier

The previously truncated tail of `Resolve_Missile_` contains the return state transition.

For a Fighter-class moving actor after its attack-stock byte at `+0x17` reaches zero:

1. recover the source/launching ship from actor field `+0x04`;
2. verify owner compatibility;
3. require the source ship to remain functional via `Combat_Ship_Is_Functional_`;
4. exclude Assault Shuttle type `0x1C` from the normal Fighter-return branch;
5. change the actor target mode back to a ship target;
6. set actor target ship `+0x06` to its source carrier;
7. set the actor active state and let the common moving-actor path fly it home.

If the launching carrier is no longer functional when that return transition is attempted, the actor follows the `Kill_Missile_` cleanup path instead.

`Destroy_Ship_` itself does not immediately delete Fighter actors launched by that ship. The original therefore allows an already-launched group to remain an independent moving actor until its own lifecycle reaches the return/cleanup decision.

### 13.3 Return converts surviving craft back into Bay availability

When a Fighter-class actor resolves against its own source carrier, `Resolve_Missile_` enters a distinct recovery branch before killing the moving actor:

- surviving actor multiplicity at `+0x0F` is integer-divided by **4**;
- the source ship's matching weapon rows are found by original weapon ID;
- the matching combat weapon-row state is re-enabled/updated;
- the row's returned count is capped by the original row count;
- any remaining full returned groups can spill into another matching row;
- after the carrier state is reconciled, the moving actor is removed through `Kill_Missile_`.

Because original launch multiplicity is `BayCount * 4`, return recovers only complete four-craft groups. A damaged remainder below four craft does not recreate a complete Bay group.

This directly freezes the lifecycle shape required by MOOX:

`Bay available -> launched group -> attacks/losses -> return flight -> full-group recovery or carrier-loss cleanup`.

The exact MOOX field names remain Gate-2 work, but this lifecycle must be server-authoritative and persistence/replay safe.

### 13.4 Raid bonus stack and RNG differential are now direct

`Resolve_Raid_` builds separate attacker and defender modifiers.

Attacker modifier:

- `Get_Fleet_Commando_Bonus_(attacker empire)`;
- plus `Crew_Boarding_Combat_Bonus_(attacking ship)`.

Defender modifier:

- `Get_Fleet_Commando_Bonus_(defender empire)`;
- plus `Get_Fleet_Security_Bonus_(defender empire)`;
- plus **20** when Special ID 29 `security_stations` is installed;
- plus `Crew_Boarding_Combat_Bonus_(defending ship)`.

For each boarding-party iteration the core comparison is:

`diff = attacker_modifier + Random(100) - defender_modifier - Random(100)`

Direct branches then show:

- `diff > 20`: select an internal raid target and apply **two** internal-damage selections through `Select_Raid_Internal_` / `Apply_Internal_Damage_`;
- `diff > 50` and defending Marines remain: decrement defending Marine count by one;
- `diff < -50`: decrement the attacking boarding-party count by one.

This distinguishes Raid from Capture: Raid can damage internal systems and may kill Marines without transferring ship ownership.

### 13.5 Crew, Commando and Security bonus tables

`Crew_Boarding_Combat_Bonus_` indexes the original crew-quality table. The already extracted original values are:

| Crew quality | Boarding bonus |
|---:|---:|
| 0 | 0 |
| 1 | 5 |
| 2 | 10 |
| 3 | 15 |

`Commando_Bonus_` and `Security_Bonus_` each expose two officer-skill strengths:

- one skill flag yields `2 * (OfficerLevel + 1)`;
- the stronger flag yields `3 * (OfficerLevel + 1)`.

`Get_Fleet_Commando_Bonus_` and `Get_Fleet_Security_Bonus_` scan eligible ships of the empire and return the **maximum** applicable attached-officer bonus rather than summing all officers.

Gate 2 should preserve these as separate fleet bonuses; it must not fold them into crew quality or Security Stations.

### 13.6 Capture uses the original Ground-Combat engine

`Resolve_Capture_` constructs the boarding sides through `Get_Boarding_Info_`.

The directly passed Capture modifiers are:

- attacker: Fleet Commando bonus;
- defender: Fleet Commando bonus + Fleet Security bonus + Security Stations `+20`.

Unlike `Resolve_Raid_`, `Resolve_Capture_` does **not** directly call `Crew_Boarding_Combat_Bonus_`. This corrects the broader wording from the earlier working checkpoint.

`Get_Boarding_Info_` additionally calls `Compute_Player_Ground_Combat_Bonuses_` and `Compute_Ground_Combat_Info_`, so Capture uses the original player-wide ground-combat technology/racial equipment stack plus the explicit boarding modifiers above.

`Resolve_Capture_` then loops `Ground_Combat_Round_` while both sides still have Marines.

For each ground-combat round:

1. each side computes `current_tier_combat_value + Random(100)`;
2. if attacker roll is less than or equal to defender roll, attacker casualty-progress advances;
3. if attacker roll is greater than or equal to defender roll, defender casualty-progress advances;
4. therefore a tie advances casualty-progress on **both** sides;
5. reaching the current tier's casualty threshold decrements one Marine from that tier, records a casualty marker, resets its progress and advances to the next non-empty tier when required.

`Resolve_Capture_` observes those casualty markers and decrements its local attacker/defender Marine totals until one side reaches zero.

The captured ship's Marine state is updated before `Capture_Ship_`; the final survivor-to-captured-ship local-variable mapping remains to be labeled precisely for the deterministic Gate-2 vector.

### 13.7 Original Tactical AI owns boarding decisions

Object-1 cross references place both:

- `Boarding_Action_Type_`;
- `Resolve_Ai_Boarding_`

inside `Do_Auto_Ship_Turn_`.

The original built-in Tactical AI therefore evaluates boarding legality/value and executes either Raid or Capture, including the temporary boarded/capturing flags.

MOOX Gate 2 must consequently include boarding in the server-projected Tactical AI action catalog; boarding cannot be a player-only UI command.

### 13.8 Carrier-loss and return semantics for Gate 2

Current direct evidence supports the following boundary:

- carrier death does not synchronously erase already-launched Fighter groups;
- a group can continue to exist as a moving actor;
- once its attack stock is exhausted, a functional source carrier is required for the normal return/recovery transition;
- if that carrier is no longer functional, the group is cleaned up instead of being recovered;
- Assault Shuttles are explicitly excluded from the normal Fighter-return branch and remain owned by the boarding/capture lifecycle.

This is sufficient to freeze the state-machine ownership in Gate 2 without inventing client-side recovery behavior.

## 14. Checkpoint C - deterministic Fighter/Bomber/Heavy-Fighter vectors

This checkpoint converts the direct Fighter actor formulas into explicit Gate-2-ready reference vectors. Executable formulas take precedence over HELP prose where the two differ.

### 14.1 Original actor identities and per-launch multiplicity

The common moving-actor type IDs are:

| Original type | Role | Launch multiplicity per Bay | Attack stock |
|---:|---|---:|---:|
| `0x1C` | Assault Shuttle | 4 | 1 |
| `0x1D` | Heavy Fighter | 4 | 2 |
| `0x1E` | Bomber | 4 | 1 |
| `0x1F` | Fighter / Interceptor | 4 | 4 |

`Fire_Missile_` initializes Fighter-class multiplicity from `BayCount * 4`. Losses operate against this authoritative craft multiplicity and return/recovery later converts surviving craft back to complete groups with integer `remaining / 4`.

### 14.2 Speed formula

`Missile_Speed_` uses these Fighter-class base speeds:

| Role | Base speed |
|---|---:|
| Assault Shuttle | 6 |
| Heavy Fighter | 8 |
| Bomber | 8 |
| Fighter / Interceptor | 10 |

For supported Fighter actors the executable formula is:

`speed = base_speed + 2 * (BestWarpDriveIndex - 1) + (TransDimensional ? 4 : 0)`

The generic moving-actor helper also adds the already-audited `Fast` actor modifier bonus `+4` when that modifier bit is present. Slice-17.4 Fighter rows expose no weapon modifiers, so normal Fighter/Bomber/Heavy-Fighter Gate-2 vectors must **not** silently add or expose Fast.

The MOOX drive component order already matches the original progression:

| Drive index | Drive | Fighter speed bonus |
|---:|---|---:|
| 1 | Nuclear Drive | 0 |
| 2 | Fusion Drive | +2 |
| 3 | Ion Drive | +4 |
| 4 | Anti-Matter Drive | +6 |
| 5 | Hyper Drive | +8 |
| 6 | Interphased Drive | +10 |

Direct prior race reverse engineering identifies original player byte `+0x8BC` as `trans_dimensional`; `Missile_Speed_` adds +4 for that player trait.

This dynamic formula explains why HELP examples such as Bomber speed 10 or Heavy Fighter speed 12 are not universal constants: actual combat speed also depends on the empire's best current drive.

### 14.3 Durability formula and executable-vs-HELP difference

`Get_Missile_Hits_` assigns Fighter-class base hits:

| Role | Base hits |
|---|---:|
| Assault Shuttle | 3 |
| Heavy Fighter | 5 |
| Bomber | 4 |
| Fighter / Interceptor | 2 |

The original Armor data field read by `Get_Missile_Hits_` is the same durability progression already represented by MOOX armor component indices:

| Armor index | Armor | Executable durability multiple |
|---:|---|---:|
| 1 | Titanium | 1x |
| 2 | Tritanium | 2x |
| 3 | Zortrium | 4x |
| 4 | Neutronium | 6x |
| 5 | Adamantium | 8x |
| 6 | Xentronium | 10x |

For the supported Fighter rows:

`fighter_hits = base_hits * armor_durability_multiple`

This is direct executable behavior. Some HELP descriptions claim different Fighter-only armor multipliers (for example Tritanium 1.5x and later armors 2x/4x/5x). Gate 2 must preserve the executable formula above rather than replacing it with those HELP values.

### 14.4 Deterministic baseline vector - Nuclear Drive + Titanium

For a non-Trans-Dimensional empire using Nuclear Drive and Titanium Armor:

| Role | Speed | Hits per craft | Attack stock | Per-stock action |
|---|---:|---:|---:|---|
| Assault Shuttle | 6 | 3 | 1 | boarding/capture attempt |
| Heavy Fighter | 8 | 5 | 2 | one Fighter Beam pass **and** one Fighter Bomb pass |
| Bomber | 8 | 4 | 1 | one Fighter Bomb pass |
| Fighter / Interceptor | 10 | 2 | 4 | one Fighter Beam pass |

Each Bay launches 4 craft. A two-Bay Fighter row therefore launches multiplicity 8, with each surviving craft participating in each Beam/Bomb pass.

### 14.5 Deterministic late-game vector - Interphased + Xentronium + Trans-Dimensional

For a Trans-Dimensional empire using Interphased Drive and Xentronium Armor:

| Role | Speed | Hits per craft | Attack stock |
|---|---:|---:|---:|
| Assault Shuttle | 20 | 30 | 1 |
| Heavy Fighter | 22 | 50 | 2 |
| Bomber | 22 | 40 | 1 |
| Fighter / Interceptor | 24 | 20 | 4 |

These values follow only from the direct binary formula: drive bonus +10, Trans-Dimensional +4, and Xentronium durability multiple 10x.

### 14.6 Fighter payload selection is player-authoritative

`Fire_Fighter_Beam_` reads the player's precomputed best Fighter Beam from the original table at `0x21254`; `Fire_Fighter_Bomb_` reads the player's precomputed best Bomb from `0x21242`.

The earlier 17.4 HELP audit remains semantically correct about payload roles:

- Fighter / Interceptor: best Beam that supports Point Defense;
- Bomber: best available Bomb;
- Heavy Fighter: both the best Fighter Beam and the best Bomb;
- Assault Shuttle: boarding Marine payload rather than Beam/Bomb fire.

17.8 should therefore snapshot or derive the authoritative payload identity at launch/server resolution; the browser must not choose a weapon dynamically.

### 14.7 Beam and Bomb passes resolve per surviving craft

`Fire_Fighter_Beam_` loops from zero to the actor's current multiplicity at `+0x0F`. Each surviving craft performs its own hit/damage resolution. Against moving actors it uses `Missile_Dcv_` / `Apply_Damage_To_Missile_`; against ships it uses the normal ship-damage path.

`Fire_Fighter_Bomb_` likewise loops over current actor multiplicity and performs a separate `Random(100)` / damage application for every surviving craft.

The Heavy Fighter branch in `Resolve_Missile_` directly proves its paired payload cadence:

1. fire Fighter Beam pass;
2. do **not** decrement stock yet for Heavy Fighter;
3. if stock remains, fire Fighter Bomb pass;
4. decrement stock once.

With initial stock 2, a surviving Heavy Fighter group therefore gets two Beam+Bomb attack cycles before return. A Bomber skips the Beam path and consumes its single stock on Bomb; a Fighter skips Bomb and consumes one stock after each Beam pass, for four Beam passes.

### 14.8 Return-loss vector

The return rule from Checkpoint B gives a deterministic loss/recovery vector:

- launch 2 Bays -> 8 craft;
- if 8 survive -> `8 / 4 = 2` complete Bay groups recovered;
- if 7 survive -> `7 / 4 = 1` complete Bay group recovered;
- if 4 survive -> `4 / 4 = 1` complete Bay group recovered;
- if 3 survive -> `3 / 4 = 0` Bay groups recovered.

The incomplete remainder is not restored as fractional Bay availability.

## 15. Checkpoint D - Capture survivor transfer and reconciliation ordering

This checkpoint closes the remaining ownership/Marine ambiguity around successful ship capture and freezes the directly evidenced post-capture reconciliation order.

### 15.1 Successful Capture transfers surviving attacking Marines

`Resolve_Capture_` keeps the attacking boarding-party total in one mutable local and the defending Marine total in a separate local. `Ground_Combat_Round_` casualty markers decrement those two totals until one side reaches zero.

At the outcome branch:

- if attacker survivors are greater than zero, `Resolve_Capture_` writes the surviving **attacker** count to target ship `MarineCurrent` (`ship+0xAF`) and then calls `Capture_Ship_`;
- if attacker survivors are zero, no ownership transfer occurs and the target ship receives the surviving **defender** Marine count.

Therefore a successful capture does not refill Marines from hull capacity and does not retain the defeated defending Marines. The captured ship begins under new ownership with exactly the surviving boarding Marines.

### 15.2 Capture reconciliation order

`Capture_Ship_` performs the following authoritative sequence:

1. if the target ship has an attached officer, call `Kill_Officer_`, clear the ship officer ID and clear the officer's ship assignment;
2. clear the captured ship's temporary boarding-related bit-field state at ship field `+0x48`, bit/index 4;
3. test Special ID `0x18` = `quantum_detonator`;
4. if Quantum Detonator is installed, call `Random(10)` and self-destruct the captured ship through `Destroy_Ship_` when the result is `< 6` - an exact **60%** capture-time detonation chance;
5. if the ship survives, change `ship+0x20` to the opposite combat-side owner;
6. call `Retarget_Missiles_Aimed_At_` for the captured ship;
7. call `Remove_Continuing_Effect_Specials_`;
8. apply the normal post-capture ship state unless the new owner is Telepathic;
9. clear additional ship turn/action fields (`+0x4B = 0`, `+0x42 = 0xFF`);
10. reconcile the original combat ship/list counters.

The Marine survivor write happens **before** this sequence. Quantum Detonator therefore tests after attacker Marines have won the boarding action but before durable ownership reconciliation completes.

### 15.3 Telepathic capture exception is direct

Prior race reverse engineering maps original player byte `+0x8B8` to `telepathic`.

`Capture_Ship_` sets ship byte `+0xB0 = 1` for a normal non-Telepathic new owner. When the new owner is Telepathic, that assignment is skipped.

`Start_Turn_For_Ship_` later explicitly checks `ship+0xB0 == 1`. In that branch it:

- clears/initializes ship movement/action fields;
- sets a dedicated turn-state value;
- clears or recomputes per-weapon turn-state bytes before continuing the normal start-turn path.

The audited `Start_Turn_For_Ship_` body reads this flag but does not itself clear it. Gate 2 should therefore model it conservatively as an authoritative `post_capture_state` with the directly evidenced start-turn consequences; its eventual clearing lifetime must not be guessed unless a later original consumer is recovered.

Telepathic ownership bypasses that normal post-capture state, matching the already-known original Telepathic captured-ship exception.

### 15.4 Deterministic Raid reference vectors

Given equal static attacker/defender bonuses:

| Attacker Random(100) | Defender Random(100) | `diff` | Result |
|---:|---:|---:|---|
| 80 | 20 | +60 | internal raid damage branch (`>20`) and defender Marine loss (`>50`) |
| 60 | 30 | +30 | internal raid damage branch only |
| 50 | 50 | 0 | no threshold effect |
| 20 | 80 | -60 | attacking boarding party loses one Marine/party unit (`<-50`) |

With non-zero static modifiers, add attacker static bonuses and subtract defender static bonuses before applying the same strict thresholds.

### 15.5 Capture result vector

A deterministic ownership vector now follows independently of the internal Ground-Combat tier values:

- initial attacker boarding party: 8 Marines;
- Ground-Combat resolution leaves attacker 3, defender 0;
- target `MarineCurrent` is written to **3**;
- `Capture_Ship_` removes the target officer if present;
- Quantum Detonator absent -> ownership flips to the attacker side;
- inbound missiles are retargeted and continuing effects removed;
- normal non-Telepathic owner receives `post_capture_state`;
- Telepathic owner skips that state.

Quantum Detonator variant:

- same 3 surviving attackers are written first;
- `Random(10)` in `0..5` -> captured ship self-destructs instead of completing normal ownership reconciliation;
- `Random(10)` in `6..9` -> ship survives and normal capture reconciliation proceeds.

### 15.6 Gate-2 implications

The server-authoritative capture result must preserve at least:

- attacker and defender Marine survivor counts;
- target `MarineCurrent` transfer;
- officer removal;
- Quantum Detonator RNG/result;
- owner change;
- missile retargeting;
- continuing-effect cleanup;
- Telepathic versus normal post-capture state;
- event/replay ordering matching the sequence above.

These are state transitions, not browser presentation decisions.

## 16. Remaining Gate-1 blockers after Checkpoint D

Still required before Gate 1 can close:

- freeze a representative Fighter Beam hit/damage/RNG vector using the already recovered `Fighter_Ocv_` and existing 17.6 Beam primitives;
- complete the per-row Class-3 Special support/effect matrix;
- recover exact formulas/order for repair, shield recharge, anti-missile rockets, Lightning Field, Energy Absorber, Reflection/Damper fields, cloak/phasing, teleporter/time effects and Beam-support Specials;
- map command/event/projection/UI requirements for Fighter, boarding and active-Special actions;
- map built-in-AI decisions beyond the now-proven boarding ownership;
- define persistence migration for Fighter actors, return state and temporary Special/boarding state;
- produce deterministic representative active/passive Special vectors.


## 17. Checkpoint E - Fighter Beam RNG/damage vector

Direct disassembly of `Fighter_Ocv_` (`0x02DF8D..0x02DFE0`) and `Fire_Fighter_Beam_` (`0x02AD57..0x02AFB3`) closes the representative Fighter-Beam RNG blocker.

### 17.1 Fighter OCV

For the normal player index path (`player < 8`), the original computes:

`fighter_ocv = Fleet_Offensive_Bonus + Best_Computer.offense + player_ship_attack_bonus(+0x8A6) + Get_Fleet_Pilot_Bonus + 50`

For the non-player/special index path, the direct player Ship Attack byte is skipped. All arithmetic reaches the firing path as signed integer state.

Target defense is then:

- ship target: ship Beam Defense;
- moving missile/fighter actor: `Missile_Dcv_`.

Checkpoint-3 re-disassembly additionally confirms the Fighter-class `Missile_Dcv_` branch as `speed * 5 + Get_Fleet_Pilot_Bonus + player Ship Defense (+0x8A5)` for normal players. The `+0x8A5/+0x8A6` semantics are independently mapped in `ADVANCED_START_RESEARCH_2026-08-28.md`.

Define `delta = fighter_ocv - target_defense`.

### 17.2 Effective hit roll

For every surviving member of the launched Fighter group, `Fire_Fighter_Beam_` calls `Random_(100)` once for the accuracy roll.

- if raw RNG is `96..99`, effective roll is forced to `100`;
- otherwise effective roll is `min(100, raw_rng + delta)`;
- the shot misses when effective roll is strictly below `40`;
- therefore `40` is the inclusive hit threshold.

This is the original Fighter-Beam accuracy gate and must use authoritative combat RNG in MOOX.

### 17.3 Damage interpolation

Let the selected Fighter Beam's original damage endpoints be `min_damage` and `max_damage`.

For a hit:

- `spread = max_damage - min_damage`;
- if `spread <= 0`, raw damage is `min_damage`;
- otherwise raw damage is
  `min_damage + ((effective_roll - 40) * (spread + 1)) / 60`
  using original integer division semantics.

The resulting damage then enters the existing shared target path:

- ship target -> `Apply_Ship_Damage_`;
- moving actor -> `Apply_Damage_To_Missile_`.

The already-audited retarget/kill path remains shared for moving actors.

A directly visible weapon-flag branch (`flag & 0x04`) adds `+25` to the accuracy delta and halves both damage endpoints with upward integer rounding before the hit/damage calculation. Gate 2 must preserve that ordering rather than folding it into post-hit damage.

### 17.4 Deterministic representative vectors

Normalized constant-damage vector, chosen to isolate the accuracy gate:

- `Fleet_Offensive_Bonus = 10`
- best-computer offense = `20`
- player Ship Attack bonus (`+0x8A6`) = `5`
- fleet pilot bonus = `15`
- therefore `fighter_ocv = 100`
- beam damage endpoints = `5..5`

Against target defense `120`, `delta = -20`:

| raw `Random_(100)` | effective roll | result | raw damage |
|---:|---:|---|---:|
| 59 | 39 | miss | 0 |
| 60 | 40 | hit | 5 |
| 95 | 75 | hit | 5 |
| 96 | 100 | hit via forced-100 branch | 5 |

Against target defense `40`, `delta = +60`:

- raw RNG `0` -> effective `60` -> hit;
- raw RNG `40` -> effective `100` after clamp -> hit.

Interpolation vector:

- `min_damage = 4`, `max_damage = 8`, effective roll `64`;
- `spread = 4`;
- raw damage = `4 + ((64 - 40) * 5) / 60 = 6`.

These vectors are resolver references, not claims about a particular named Beam technology.

### 17.5 Gate-2 consequence

Fighter and Heavy-Fighter Beam fire should not gain a parallel accuracy/damage model. The server-authoritative Fighter actor needs to feed the recovered Fighter OCV and per-member RNG into the existing Slice-17.6 Beam-defense/damage primitives, while moving-actor targets continue through the shared missile/fighter actor defense/damage path.

The Fighter-Beam RNG/vector blocker from Checkpoint D is therefore closed.


## 18. Checkpoint F - Class-3 per-row support/effect matrix

This checkpoint classifies every deferred 17.4 Class-3 row by original effect family and by current evidence strength. A row can be **consumer-proven** while its exact formula/order remains **OPEN**; that distinction is intentional.

Evidence levels used here:

- **FROZEN**: direct executable path is sufficient to freeze the listed effect/order now;
- **PROVEN / formula open**: an original Tactical consumer is directly identified, but further callsite/arithmetic recovery is required;
- **HELP-classified / consumer open**: original HELP and the 17.4 row audit define the effect family, but Gate 1 has not yet accepted an executable consumer/formula;
- **existing MOOX snapshot**: the state projection already exists, but this is not itself proof of original Tactical resolution.

| ID | Special | Classification | Direct Gate-1 status |
|---:|---|---|---|
| 1 | Achilles Targeting Unit | Beam-related / Tactical-common | **PROVEN / formula open**: shared ship-damage resolution receives the Achilles presence flag; exact subsystem/armor ordering still needs the 17.6 damage-call contract cross-check. |
| 2 | Augmented Engines | Tactical-common / mobility | **HELP-classified / consumer open**: +5 combat speed is source-backed; exact original speed derivation/ordering remains to be recovered. |
| 3 | Automated Repair Unit | Tactical-common / repair | **FROZEN core rate**: `Repair_All_Combat_Ships_` gives the Special a base repair amount of 20 and adds the side Fleet Engineer bonus before `Repair_Combat_Ship_`; stasis and post-capture state suppress this repair pass. Detailed allocation order inside `Repair_Combat_Ship_` is directly recovered but field naming still needs mapping. |
| 5 | Battle Scanner | Beam-related plus strategic scan | **HELP-classified / consumer open**: +50 Beam hit and +2 parsec scan are source-backed; Tactical hit-stack insertion point remains open. |
| 6 | Cloaking Device | Tactical-common / defense/state plus strategic stealth | **PROVEN / formula open**: original cloak state/presentation exists and boarding explicitly manipulates cloak/phasing state; exact +Beam-defense/missile-miss/recloak turn ordering remains open. |
| 7 | Damper Field | Tactical-common / damage plus boarding-related | **FROZEN boarding hook**: Transporter boarding party is integer-divided by 3 in the evidenced Damper branch. General 3/4 damage reduction ordering remains open. |
| 8 | Displacement Device | Tactical-common / defense | **PROVEN / partial formula**: `Energy_Absorber_Attack_` checks ID 8 and zeros that attack when `Random_(100) > 30`; generic all-weapon placement/order remains open and HELP must not override executable off-by-one semantics. |
| 9 | ECM Jammer | ordnance-related / Tactical-common | **HELP-classified / consumer open**: +70 missile evasion source-backed; exact `Missile_Dcv_` contribution remains to be recovered. |
| 10 | Energy Absorber | Tactical-common / stored-damage active effect | **PROVEN / formula open**: dedicated `Energy_Absorber_Attack_` and presentation path exist; storage fraction, expiry and activation ordering still require callsite audit. |
| 12 | Fast Missile Racks | ordnance-related / readiness | **HELP-classified / consumer open**: two-volley/reload behavior source-backed; exact weapon-row readiness state transition remains open. |
| 13 | Hard Shields | Tactical-common / shield plus boarding-related | **FROZEN boarding hook**: Transporter boarding is rejected when ID 13 is present. General -3 damage, nebula and shield-piercing ordering remains open. |
| 14 | Heavy Armor | Tactical-common / survivability | **PROVEN / partial formula**: `Repair_Combat_Ship_` checks ID 14 and triples the computed armor capacity used by repair; HELP independently states triple armor and armor-piercing negation. Damage-path ordering still needs cross-check. |
| 15 | High Energy Focus | Beam-related | **HELP-classified / consumer open**: +50% energy-weapon damage source-backed; exact pre/post-range placement remains open. |
| 16 | Hyper-X Capacitors | Beam-related / readiness | **HELP-classified / consumer open**: two Beam volleys then reload-turn behavior source-backed; exact readiness state machine remains open. |
| 19 | Lightning Field | ordnance/fighter-related area defense | **PROVEN / formula open**: dedicated `Lightning_Field_` presentation routine exists, but the recovered body is animation/geometry; the actual 50% destroy resolver and target-loop/RNG position still need the caller. |
| 20 | Multi-Phased Shields | Tactical-common / shield | **HELP-classified / consumer open**: +50% max shield strength source-backed; exact max-shield derivation hook remains open. |
| 21 | Multi-Wave ECM Jammer | ordnance-related / Tactical-common | **HELP-classified / consumer open**: +100 missile evasion source-backed; exact defense-stack insertion remains open. |
| 22 | Phase Shifter | Tactical-common / active mobility | **PROVEN / formula open**: teleport family has direct combat-grid consumers; Phase Shifter's self/other 18-square command legality and once-per-turn state still require the exact caller. |
| 23 | Phasing Cloak | Tactical-common / defense/state | **PROVEN / formula open**: cloak/phasing state is directly present; 10-turn phase lifetime and transition to ordinary cloak remain to be recovered from executable state consumers. |
| 24 | Quantum Detonator | Tactical-common / capture/destruction | **FROZEN capture hook**: `Capture_Ship_` calls `Random_(10)`; results 0..5 self-destruct, 6..9 continue capture. HELP's prose percentage is therefore not used as the executable contract. Explosion-power hook remains open. |
| 25 | Rangemaster Unit | Beam-related | **HELP-classified / consumer open**: range/3 for to-hit is source-backed; exact range-to-hit helper insertion remains open. |
| 26 | Reflection Field | Beam-related / Tactical-common defense | **PROVEN / formula open**: dedicated reflection-field hit/display subsystem exists; exact reflection chance/damage return resolver remains open. |
| 27 | Reinforced Hull | Tactical-common / survivability | **HELP-classified / consumer open**: triple structure and drive protection source-backed; exact snapshot/damage-component derivation hook remains to be proven in the original path. |
| 29 | Security Stations | boarding-related | **FROZEN**: +20 is included in defender boarding combat rolls through the recovered security-bonus path. |
| 30 | Shield Capacitors | Tactical-common / shield recharge | **FROZEN**: `Recharge_Shields_` uses factor 3 normally and factor 10 with ID 30, so recharge budget changes from 30% to 100% of the base per-facing shield value per invocation; budget is distributed one point at a time across deficient facings until exhausted or all four are full. |
| 32 | Structural Analyzer | Beam-related / Tactical-common damage | **PROVEN / formula open**: shared `Apply_Ship_Damage_` call receives ID 32 presence; HELP says penetrating damage doubles. Exact shield-boundary/order remains to be cross-checked against the 17.6 damage contract. |
| 33 | Sub-Space Teleporter | Tactical-common / active mobility | **PROVEN / formula open**: `Teleport_Enemy_Ship_` uses combat-grid/legal-move authority and updates coordinates; the own-ship 20-square command/use-state path still needs exact caller/order. |
| 34 | Time Warp Facilitator | Tactical-common / turn economy | **HELP-classified / consumer open**: additional end-of-round turn is source-backed; exact scheduler insertion and action/readiness reset semantics remain open. |
| 35 | Transporters | boarding-related / active | **FROZEN core legality**: direct boarding subsystem owns the action; 12-square HELP range, facing-shield condition, Hard-Shield rejection and Damper interaction are now all part of the Gate-1 evidence. Planet bomb-drop remains Slice-19 boundary. |
| 36 | Troop Pods | boarding-related / Tactical-common state | **FROZEN snapshot effect**: doubles `MarineMax` in the current Tactical snapshot and original HELP; boarding consumes the resulting Marine state normally. |
| 37 | Warp Dissipator | Tactical-common / retreat control | **HELP-classified / consumer open**: prevents enemy retreat while functioning; exact retreat legality hook remains open. The current generated `ship_specials.json` does not expose this named row even though the frozen 17.4 original table has selectable row 37, so Gate 2 must resolve that data-projection gap explicitly rather than silently omitting the row. |
| 38 | Wide Area Jammer | ordnance-related / fleet defense | **HELP-classified / consumer open**: +130 self / +70 fleet missile evasion and best-only fleet stacking are source-backed; exact fleet-defense helper and ordering remain open. |

### 18.1 Automated Repair direct arithmetic

`Repair_All_Combat_Ships_` obtains each side's `Get_Fleet_Engineer_Bonus_`. A directly checked player flag at player-record `+0x8B0` adds another +10 for that side; the identity of that player attribute is not named here until its race/leader mapping is independently confirmed.

For each eligible ship:

- destroyed/inactive state is skipped;
- `post_capture_state` is skipped;
- stasis is skipped;
- Automated Repair Unit (ID 3) contributes **20**;
- the side Engineer bonus is added;
- only a positive total calls `Repair_Combat_Ship_(ship, repair_percent)`.

`Repair_Combat_Ship_` then converts the supplied percentage into bounded repairs against original maximum values. The recovered body includes armor/structure/internal-system pools and minimum-one behavior in the early special branches. Field labels and the exact pool priority will be mapped before Gate 1 closure instead of guessed from offsets.

### 18.2 Shield Capacitor direct arithmetic

`Recharge_Shields_`:

- skips ships without shields;
- skips stasis;
- requires the original recharge-eligible state at ship `+0x134`;
- uses recharge factor **3** normally and **10** with Shield Capacitors (ID 30);
- computes budget as `base_shield_value * factor / 10`;
- repeatedly scans the four shield facings in index order, adding one point to a deficient facing and decrementing budget;
- stops when budget is exhausted or a complete pass changes nothing.

Thus Shield Capacitors do not merely add an abstract +70 modifier after recharge: they replace the per-invocation budget factor from 30% to 100%, preserving the original pointwise facing distribution order.

### 18.3 Anti-Missile Rockets supporting evidence

Although Anti-Missile Rockets is outside the user's deferred Class-3 list, its direct resolver constrains the shared ordnance-defense architecture:

- one `Random_(100)` is consumed per rocket attempt;
- an attempt consumes the ship's rocket count/ammunition state whether it succeeds or fails;
- success decrements the target moving-actor multiplicity and kills the actor when that count reaches zero.

The exact success threshold is executable but depends on an unresolved call argument. Gate 1 therefore does not yet name that argument or freeze the final threshold formula.

### 18.4 Matrix consequence

The **per-row classification** blocker is now closed. Gate 1 still cannot close because rows marked formula/order open need executable placement, and active-special command/event/persistence/AI contracts remain unfinished.


## 19. Checkpoint G - cross-slice Beam Special reconciliation

The accepted Slice-17.6 original Beam audit already recovered executable consumers that 17.8 must now consume rather than re-audit from HELP prose. These findings supersede the weaker evidence labels in the Checkpoint-F table where applicable.

### 19.1 Beam-support effects already directly observed in 17.6

The original Beam path directly established:

- **Rangemaster Unit (ID 25):** divide the range input by **3** before the Beam range-To-Hit lookup.
- **High Energy Focus (ID 15):** add **+50** to the Beam damage multiplier.
- **Structural Analyzer (ID 32):** double a downstream **post-Shield** damage value.
- **Hard Shields (ID 13):** neutralize the Shield-Piercing bypass behavior.
- **Heavy Armor (ID 14):** participates in Armor-Piercing resistance.
- **Damper Field (ID 7):** suppresses the Neutron/Death-Ray Marine-kill hook.
- **Achilles Targeting Unit (ID 1):** can force the deeper-internal damage-selection path; the exact subsystem weighting remains owned by 17.8 but the consumer itself is proven.

These are executable dependencies recorded by the frozen 17.6 contract, not new approximations.

### 19.2 Updated matrix status

Accordingly:

| Special | Checkpoint-F label | Superseding Gate-1 status |
|---|---|---|
| High Energy Focus | HELP-classified / consumer open | **FROZEN Beam hook:** +50 damage multiplier. |
| Rangemaster Unit | HELP-classified / consumer open | **FROZEN Beam hook:** range input /3 for To-Hit lookup. |
| Structural Analyzer | PROVEN / formula open | **FROZEN Beam hook:** x2 downstream post-Shield damage; broader non-Beam applicability remains open. |
| Hard Shields | FROZEN boarding hook; general Beam open | **FROZEN Beam + boarding hooks:** cancels Shield Piercing and rejects Transporter boarding. Other -3/nebula semantics remain open. |
| Heavy Armor | PROVEN / partial formula | **FROZEN Beam resistance + repair-capacity hook:** participates in Armor-Piercing defense and triples recovered armor capacity. |
| Damper Field | FROZEN boarding hook; damage open | **FROZEN Beam-marine + boarding hooks:** suppresses Beam Marine-kill hook and reduces Transporter boarding party through the recovered /3 branch. General incoming-damage reduction remains open. |
| Achilles Targeting Unit | PROVEN / formula open | **PROVEN executable Beam damage-selection hook:** exact internal subsystem weighting still open. |

### 19.3 Ownership consequence

Gate 2 must expose these as shared Tactical resolver hooks rather than duplicate Beam formulas inside the 17.8 Special subsystem. Slice 17.8 owns Special state/effect activation; Slice 17.6's already-frozen Beam resolver remains the consumer for Beam-specific range, damage, shield-piercing, armor-piercing and Marine-kill behavior.

This closes the Beam-support placement blocker for High Energy Focus and Rangemaster and narrows Structural Analyzer, Hard Shields, Heavy Armor, Damper Field and Achilles to their remaining non-Beam or detailed-selection questions.

## 20. Checkpoint H - non-Beam defensive/readiness Special executable order

This checkpoint continues the direct Orion2.exe audit and reconciles the Slice-17.7 neutral hook positions with their original Slice-17.8 Special consumers. Executable behavior takes precedence over HELP prose where the two differ.

### 20.1 Lightning Field - exact pre-impact grouped-actor loop

`Resolve_Missile_` checks Special ID 19 on the target ship before `Final_Fire_` / reactive Beam-PD and before MIRV expansion.

When Lightning Field is present, the original loop:

1. initializes an iteration index to zero;
2. calls `Random_(100)`;
3. when the roll is `< 50`, decrements the moving actor's live multiplicity at `+0x0F`;
4. increments the iteration index;
5. compares that index against the **mutated current multiplicity**, not the launch-time multiplicity;
6. repeats while the index is still lower than the current multiplicity.

Therefore the exact original contract is not 'roll once for every originally launched child.' Destruction during the loop shortens the same loop.

Representative grouped-actor vector, initial multiplicity 4:

- rolls `<50,<50` -> multiplicity 4 -> 3 -> 2, then stop because iteration index 2 is no longer below live multiplicity 2;
- rolls `<50,>=50,>=50` -> multiplicity 4 -> 3 and stop after three attempts;
- all rolls `>=50` -> four attempts, multiplicity remains 4.

The optional `Lightning_Field_` routine recovered earlier is the presentation/animation consumer; the authoritative destruction RNG and multiplicity mutation live in `Resolve_Missile_`.

This directly fills Slice 17.7's frozen `pre_impact_defense` hook and preserves its ordering before reactive PD and MIRV split.

### 20.2 Displacement Device - target-path executable branches

For ordinary ordnance impact in `Resolve_Missile_`:

- Special ID 8 triggers one additional `Random_(100)`;
- roll `< 30` forces the impact comparison into the miss branch;
- roll `>= 30` continues normal impact resolution.

For Beam fire in `Fire_Beam_Weapon_`:

- Special ID 8 again consumes a separate `Random_(100)`;
- the direct branch forces the Beam miss when the roll is **`<= 30`**;
- the normal path continues only when the roll is `> 30`.

That Beam off-by-one difference is executable evidence and must not be normalized to HELP's prose percentage.

The Energy-Absorber active attack also contains an ID-8 branch, but its surrounding special-attack semantics remain separately open; Gate 1 does not generalize the ordinary Beam/ordnance branch into that active-special path.

### 20.3 ECM / jammer composition and fleet stacking

`Ship_Specials_Defensive_Bonus_` directly applies:

- active Cloaking Device state: **+80 Beam Defense**;
- ECM Jammer ID 9: **+70 Missile Defense**;
- Multi-Wave ECM Jammer ID 21: **+100 Missile Defense**;
- Wide Area Jammer ID 38: **+130 Missile Defense** to the equipped ship.

For Wide Area Jammer, `Fleet_Defensive_Bonus_` scans functional friendly ships and contributes **+70 Missile Defense** to other friendly ships when an active friendly ID-38 carrier exists.

`Defensive_Combat_Bonus_` receives both the equipped ship's jammer magnitude and the fleet jammer magnitude and explicitly removes the weaker overlapping jammer contribution. The original therefore implements **best-only jammer stacking**, not additive ECM + Multi-Wave + Wide-Area accumulation.

These values now fill Slice 17.7's frozen `impact_defense` Special hook.
### 20.4 Cloaking Device / Phasing Cloak state machine - directly evidenced portion

`Init_Special_Devices_` initializes:

- Cloaking Device ID 6 -> ship cloak state `+0x40 = 1`;
- Phasing Cloak ID 23 -> ship cloak state `+0x40 = 4` and byte `+0x41 = 10`.

Direct consumers prove:

- state 1 -> +80 Beam Defense through `Ship_Specials_Defensive_Bonus_`;
- state 1 -> an additional `Random_(100)` in `Resolve_Missile_`, with roll `<50` forcing the missile impact to miss;
- state 4 -> untargetable/excluded by multiple target-selection, retarget, defensive-fire, movement/boarding eligibility paths;
- weapon fire breaks state 1 -> 2 and state 4 -> 6;
- initiating/performing boarding likewise breaks 1 -> 2 and 4 -> 6;
- next-turn initialization advances 2 -> 3 and 6 -> 7;
- firing while in 3/7 moves those states back to 2/6.

A whole-code direct-field scan found the `+0x41 = 10` initialization but no accepted Tactical decrement/expiry consumer yet, and no direct `3 -> 1` / `7 -> 4` writer. Gate 1 therefore **does not yet freeze** HELP's ten-turn Phasing-Cloak expiry or the exact ordinary recloak transition. Those lifecycle transitions remain OPEN rather than inferred.

### 20.5 Energy Absorber - incoming storage and pre-weapon discharge

`Apply_Ship_Damage_` checks Energy Absorber ID 10 before Damper Field and before the later shield/internal application:

1. compute signed incoming damage /4 using original integer semantics;
2. clamp the absorbed amount to at least 1;
3. subtract that amount from the current incoming-damage value;
4. add exactly that amount to ship word `+0x45`;
5. only then apply the Damper Field ID-7 reduction to the remaining damage.

Thus Energy Absorber storage precedes Damper Field.

`Pre_Weapon_Specials_` later sees stored energy `+0x45 > 0` and, for a ship target at range `< 8` while the Special is active, dispatches `Energy_Absorber_Attack_` **before** ordinary weapon resolution. The attack consumes the stored value and clears `+0x45 = 0` after resolution.

A direct writer scan finds initialization, damage-time accumulation and attack-time clear, but no accepted turn-expiry writer. HELP's 'fire next turn or it is lost' lifetime therefore remains OPEN until an indirect consumer is recovered; MOOX must not invent that expiry yet.

### 20.6 Damper Field general damage ordering is now direct

Immediately after Energy Absorber handling, `Apply_Ship_Damage_` checks Damper Field ID 7 and reduces the remaining incoming damage to one quarter with original signed shift/division semantics before subsequent shield/internal handling.

Together with earlier checkpoints, Damper Field now has three distinct direct consumers:

- general incoming-damage quartering;
- Beam Marine-kill suppression;
- Transporter-boarding party reduction.

### 20.7 Reflection Field - executable Beam reflection branch

`Fire_Beam_Weapon_` checks Reflection Field ID 26 on the target before ordinary `Apply_Ship_Damage_`.

For current Beam hit damage `D`:

- call `Random_(D + 15)`;
- compare the roll with `D`;
- if the roll is strictly greater than `D`, move the entire current hit damage `D` into the reflection output and set target damage to zero;
- otherwise continue normal target damage.

The caller then applies the reflected amount back to the firing ship through `Apply_Ship_Damage_`.

This exact `Random_(D+15) > D` executable branch supersedes HELP's differently worded ratio formula for original-parity purposes.
### 20.8 Hyper-X Capacitors / Fast Missile Racks - shared weapon-slot state

`Init_Special_Devices_` initializes weapon-slot byte `+0x5C = 1`:

- for Beam rows when Hyper-X Capacitors ID 16 is present;
- for Missile rows when Fast Missile Racks ID 12 is present.

At turn start, `Init_Ship_For_Start_Of_Turn_`:

- restores the live per-turn weapon count;
- for an eligible Hyper-X/FMR row whose `+0x5C != 1`, writes state `2`.

`Fire_Ship_` directly uses the state:

- when the current per-turn weapon count is empty and `+0x5C == 1`, it refills the live count from the installed/base count and clears `+0x5C` to 0, permitting the extra initial volley;
- after a firing action, state `2` is cleared to 0.

A whole Tactical-code writer scan found no other normal writer that restores `+0x5C = 1` after combat initialization. `Init_Special_Devices_` is called from combat/setup/load contexts, not as the ordinary per-turn reset.

Therefore the directly observed Orion2.exe state machine provides an **initial second-volley capability**, followed by the ordinary per-turn state transitions above. The HELP claim that one full non-firing turn re-arms another double volley is **not yet supported by the recovered executable writer graph**. Gate 1 records this as an executable-vs-HELP discrepancy rather than fabricating a re-arm writer.

### 20.9 Anti-Missile Rockets - 17.7/17.8 ownership reconciliation

The accepted Slice-17.7 original audit already froze the neutral projectile-interceptor hook and direct original facts needed by 17.8:

- active dispatch through `Fire_Special_Weapon_`;
- valid ordinary Missile actor targeting plus the separate original compatible actor class;
- max range 15;
- hit chance `85% - 5% per square`;
- projectile durability/multiplicity damage and `Kill_Missile_` at zero.

The current `Fire_Special_Weapon_` disassembly reconfirms the moving-actor legality/owner checks and dispatch to `Anti_Missile_Rockets_`. 17.8 should consume the frozen 17.7 `projectile_interceptor` hook rather than creating a second projectile model.

### 20.10 Warp Dissipator - player and AI retreat legality

`Enemy_Dissipator_Exists_` checks Special ID 37 across the combat ship list and returns true only for an opposing ship that:

- carries Warp Dissipator;
- is not in the normal post-capture state;
- is not destroyed/retreated/inactive;
- belongs to the opposing combat side.

Direct callers prove three consumers:

- `Do_Auto_Ship_Turn_` will not set the AI retreat state while an enemy Dissipator exists;
- `Retreat_Check_` aborts retreat evaluation when an enemy Dissipator exists;
- the manual Tactical command path in `Do_Combat_Turn_` calls `Enemy_Dissipator_Exists_` before normal retreat eligibility and branches directly to its rejection/message path when true.

Warp Dissipator is therefore an authoritative Tactical retreat-legality rule for both human commands and AI, not a browser/UI-only restriction.

The previously identified data gap remains: original selectable row 37 exists, but the current generated MOOX `ship_specials.json` omits the named row. Gate 2 must repair that projection explicitly before implementation.

### 20.11 Remaining blockers after Checkpoint H

The following remain open before Gate 1 can close:

- exact ordinary recloak transition and Phasing-Cloak ten-turn lifetime, if an executable indirect consumer exists;
- exact Energy-Absorber stored-energy lifetime/expiry, if any;
- Phase Shifter / Sub-Space Teleporter active command legality, range/use-state and persistence;
- Time Warp Facilitator scheduler/action-reset semantics;
- Achilles exact internal subsystem weighting;
- remaining active-special command/event/rejection/projection/UI contracts;
- built-in AI ownership for active Special commands beyond the already-proven retreat/boarding consumers;
- persistence/replay schema for cloak/phasing, stored energy, active-special use state and fighter return state;
- deterministic representative vectors for the remaining active Specials.

Gate 1 remains **IN PROGRESS**. No Slice-17.8 production implementation is authorized yet.

## 21. Checkpoint I - lifecycle correction, teleporter use and Time-Warp scheduler

This checkpoint extends the disassembly window through the full `Do_Combat_Turn_` body. It **supersedes** the provisional OPEN conclusions in Checkpoint H for Hyper-X/FMR re-arm, Cloak/Phasing lifecycle and Energy-Absorber expiry: the missing writers were located late in the large turn-cleanup body.

### 21.1 Hyper-X Capacitors / Fast Missile Racks re-arm is now closed

At the end-of-pass cleanup inside `Do_Combat_Turn_`, every functional weapon slot whose special-readiness byte is `+0x5C == 2` is changed to `1`.

Combined with Checkpoint H:

- combat initialization gives eligible Beam/Missile rows state 1;
- state 1 permits the immediate extra volley and is consumed to 0;
- a later turn start changes non-1 state to 2;
- firing during that turn consumes 2 -> 0;
- **not firing** leaves state 2 intact;
- end-of-pass cleanup changes surviving state 2 -> 1;
- the next turn therefore regains the double-volley opportunity.

This directly proves HELP's 'spend one full turn without firing to reload the double volley' behavior. The earlier Checkpoint-H statement that no re-arm writer existed is superseded by this full-function recovery.

### 21.2 Cloaking Device recloak and Phasing-Cloak lifetime are now closed

The same `Do_Combat_Turn_` cleanup resolves the previously missing state transitions:

- Cloaking state 3 -> draw recloak transition -> state 1;
- Phasing state 4 -> decrement `ship+0x41` by 1;
- while that counter remains non-zero, state 4 persists;
- when the counter reaches zero, the executable transitions from Phasing presentation to ordinary Cloaking state 1;
- state 7 -> state 4 and the resolver retargets relevant incoming moving actors/missiles away from the newly phased ship.

Together with the already frozen start/fire transitions, the ordinary Cloaking cycle is therefore:

`1 (cloaked) -> fire/board -> 2 -> next-turn start -> 3 -> no further fire during that turn -> end-pass -> 1`.

For Phasing Cloak:

- combat init: state 4, counter 10;
- active state 4 is untargetable;
- each recovered active-state cleanup decrements the counter;
- fire/board breaks 4 -> 6;
- next-turn start advances 6 -> 7;
- end-pass 7 -> 4 and retargets relevant incoming actors;
- when the active counter reaches zero, Phasing converts to ordinary Cloaking state 1.

The current Orion2.exe cleanup loop processes functional combat ships at each `Do_Combat_Turn_` pass; the owner-comparison bytes in this loop are patched to NOPs in the audited 1.31 executable. Gate 2 must therefore model the recovered pass ordering explicitly rather than assuming a generic once-per-round Phasing decrement.

### 21.3 Energy Absorber expiry is now direct

Late in `Do_Combat_Turn_`, before the remaining end-pass Special cleanup, the original clears stored Energy-Absorber word `ship+0x45 = 0` for the acting side's eligible functional ships in the ordinary local-combat path.

Combined with the damage-time store and pre-weapon discharge:

- damage received during the opponent's pass stores one quarter before Damper Field;
- the ship can discharge that stored energy through `Pre_Weapon_Specials_` during its next own pass;
- if not discharged, the value is cleared at the end of that own pass.

This closes the HELP lifetime semantics without inventing a timer field.

### 21.4 Teleporter used-state and self-teleport contract

`Used_Teleporter_` tests bit 0 of ship bitfield `+0x48`; `Set_Teleporter_Used_` sets the same bit. `Init_Ship_For_Start_Of_Turn_` clears `ship+0x48` before restoring any boarding/capture flag, so the teleporter-use bit is reset at the start of each eligible ship turn.

`Move_Ship_` directly checks for either:

- Sub-Space Teleporter ID 33, or
- Phase Shifter ID 22.

If one is present, the teleporter has not already been used, and `Range_To_XY_ <= 20`, the original:

1. sets the teleporter-used bit;
2. runs the teleport presentation;
3. jumps directly to the destination coordinate update;
4. does not traverse the normal movement path.

The caller compares teleporter-used state before and after `Move_Ship_`; when the bit changed, it skips subtraction of ordinary movement cost. The coordinate update changes the ship grid position without a facing-change step, matching the Sub-Space Teleporter direction-preservation rule.

`Draw_Legal_Boxes_` also allows ships with ID 33 or ID 22 and unused teleporter state to expose teleport movement even when ordinary movement points are exhausted. `Set_Legal_Moves_` expands the relevant legal-move radius to 20 for either special; its explicit Sub-Space-Teleporter special-cell marking also requires the Special to be turned on.

Thus the directly evidenced self-teleport range is **20 original Range units for both IDs 22 and 33**. The Phase-Shifter HELP value 18 is not used to override the executable self-move check.

### 21.5 Phase Shifter other-ship teleport remains unproven

A named `Teleport_Enemy_Ship_` routine exists and can choose a legal nearby destination and update another ship's coordinates, but a whole code-object scan found no direct CALL/JMP or immediate code reference to its entry point in the audited executable.

Gate 1 therefore does **not** yet claim that this routine is the reachable Phase-Shifter user action. The directly evidenced ID-22 self-teleport is frozen; other-ship teleport remains OPEN until a dispatch-table/data-reference or live command path is recovered.

### 21.6 Time Warp Facilitator is a second authoritative turn pass

`Tactical_Combat_` calls `Do_Combat_Turn_` in this round order:

1. normal pass for the first combat side with mode 0;
2. winner/effect checks;
3. normal pass for the second combat side with mode 0;
4. winner/effect checks;
5. extra pass for the first side with mode 1;
6. winner check;
7. extra pass for the second side with mode 1;
8. winner check;
9. increment the combat round and run global `End_Of_Turn_Bookeeping_`.

Inside `Do_Combat_Turn_(side, 1)`:

- only functional/non-stasis ships carrying Time Warp Facilitator ID 34 are initialized for turn start;
- only ID-34 ships are inserted into that pass's ship turn order;
- `Init_Ship_For_Start_Of_Turn_` is called, so movement/action/weapon and per-turn Special state receive the normal start-turn reset for the extra pass;
- end-pass cleanup still runs;
- the normal end-pass `Seeking_Missiles_` call is skipped for mode 1.

Global shield recharge and Automated Repair remain in `End_Of_Turn_Bookeeping_` after both normal and Time-Warp passes, so Time Warp does **not** independently double the once-per-round shield-recharge/repair bookkeeping.

A Time-Warp ship that also carries a teleporter receives a fresh teleporter-used reset through the extra pass's `Init_Ship_For_Start_Of_Turn_`, consistent with the extra pass being a real additional ship turn rather than a partial action token.

### 21.7 Corrected remaining blockers

Closed by this checkpoint:

- Hyper-X / Fast Missile Racks full-turn re-arm;
- ordinary Cloak recloak;
- Phasing-Cloak counter decrement and transition to ordinary cloak;
- Energy-Absorber stored-energy expiry;
- Time Warp Facilitator pass scheduling and major reset/bookkeeping boundaries;
- self-teleporter range, used-state reset and no-normal-movement-cost behavior.

Still open before Gate 1 closure:

- reachable Phase-Shifter other-ship teleport dispatch, if present in the 1.31 executable;
- Achilles Targeting Unit exact subsystem-selection weighting;
- authoritative MOOX command/rejection/event/projection/UI names and payloads for fighter/boarding/active-Special actions;
- persistence/replay migration for fighter actors, cloak/phasing counters/states, stored energy, teleporter-used state and temporary boarding/capture state;
- built-in AI choices for active Specials beyond the already recovered boarding, retreat and generic movement consumers;
- final deterministic active-Special reference vectors and Gate-1 closure checklist.

Gate 1 remains **IN PROGRESS**; no production implementation starts here.

## 22. Checkpoint J - Achilles Targeting Unit internal-selection formula

The remaining Achilles Targeting Unit blocker is now closed from the shared Beam -> `Apply_Ship_Damage_` -> `Select_Internal_` path.

### 22.1 Attacker Special is passed explicitly into ship damage

`Fire_Beam_Weapon_` checks the firing ship for:

- Special ID 1 = Achilles Targeting Unit;
- Special ID 32 = Structural Analyzer.

The resulting booleans are passed explicitly into `Apply_Ship_Damage_`; Achilles is therefore an attacker-side damage-selection modifier, not a target metadata flag.

### 22.2 Achilles forces the armor-bypass internal path

Before internal selection, `Apply_Ship_Damage_` derives an armor-bypass control value.

- when Achilles is present, that control is forced to 1 immediately;
- otherwise ordinary Armor-Piercing weapon bit `0x0008` can request the bypass;
- the ordinary Armor-Piercing request is cancelled by Heavy Armor ID 14 or Xentronium armor;
- the Achilles branch occurs before those cancellation checks and jumps directly to the enabled-bypass state.

Therefore Achilles completely bypasses the target Armor layer even when ordinary Armor Piercing would have been neutralized by Heavy Armor/Xentronium. This directly matches the qualitative HELP statement, but is now executable evidence.

### 22.3 Exact weighted-choice transform

`Select_Internal_` builds a 19-entry weighted choice from current surviving internal pools.

The recovered choice mapping includes:

- two fixed leading internal categories;
- eight Weapon-slot categories sourced from ship fields `+0x104 .. +0x112`;
- eight additional internal/Special-slot categories sourced from `+0x124 .. +0x132`;
- the Shield-generator/internal shield-system category sourced from `+0x134`.

When the Achilles boolean is non-zero, the executable modifies the weighted-choice input before `Get_Weighted_Choice_Short_`:

- each of the **eight Weapon-slot weights is multiplied by 3**;
- the **Shield-generator weight is multiplied by 3**;
- the other recovered internal-category weights are left at their normal values.

The selected weighted index is then mapped back to the existing internal type/index representation and damaged through `Apply_Internal_Damage_`.

Thus Achilles does not perform a separate second hit roll and does not multiply final damage. It changes internal-selection probability by a precise 3x weight transform for Weapon systems and the Shield generator after forcing Armor bypass.

### 22.4 Deterministic weight reference vector

For a normalized internal-choice state with:

- Weapon-slot A weight = 10;
- Weapon-slot B weight = 4;
- Shield-generator weight = 20;
- another non-targeted internal category weight = 12;

the Achilles pre-choice transform is:

- Weapon A: 10 -> 30;
- Weapon B: 4 -> 12;
- Shield generator: 20 -> 60;
- other category: 12 -> 12.

`Get_Weighted_Choice_Short_` then consumes the transformed array through the existing shared internal-choice RNG path. Gate 2 must preserve the transformed weights and existing RNG ordering rather than inventing a separate Achilles RNG event.

### 22.5 Gate consequence

The last open Beam-side Achilles formula is closed. Slice 17.8 should activate the already-frozen Slice-17.6 damage hook by supplying:

- `armor_bypass = true` for Achilles;
- `internal_weapon_weight_multiplier = 3`;
- `internal_shield_generator_weight_multiplier = 3`;
- neutral multiplier 1 for other internal categories.

No parallel Achilles damage resolver is required.
## 23. Checkpoint K - remaining original Special closure

This checkpoint closes the remaining original-executable Special questions that were still weaker or open after Checkpoints I/J. It also resolves the Phase-Shifter other-ship dispatch question conservatively from the audited 1.31 executable.

### 23.1 Phase Shifter other-ship teleport is not an accepted reachable 1.31 action

The named routine `Teleport_Enemy_Ship_` exists at code offset `0x09BFF3` and can choose a nearby legal destination, animate and overwrite another ship's coordinates. Presence of that routine alone is not sufficient to accept a gameplay command.

A direct reachability audit found:

- no relative `CALL` to `0x09BFF3` in the code object;
- no relative `JMP` to `0x09BFF3`;
- no decoded instruction operand that references `0x09BFF3`;
- no valid absolute code/data pointer to the function entry;
- no matching internal target in the LE fixup-record region.

One raw byte-pattern hit inside `Redraw_Main_Screen_` was disassembled and proved to be bytes spanning an unrelated instruction sequence, not a reference.

Gate 1 therefore freezes **only the directly reachable ID-22 self-teleport** already recovered in Checkpoint I: once per eligible ship turn, executable range `<= 20`, no ordinary movement-cost subtraction and facing preserved. Other-ship teleport is **unsupported/deferred for original-parity Slice 17.8** unless new executable evidence is later found. HELP prose or the orphaned routine must not create an unproven command.

### 23.2 Augmented Engines - exact speed insertion

`Calc_Current_Speed_` reads the ship's Special bitfield directly. When Special ID 2 is installed and its matching damage/disabled bit is clear, it adds exactly **+5** to the working combat-speed value before the later engine-damage scaling and other recovered speed modifiers.

Thus Augmented Engines is a passive Tactical-common speed derivation hook, not an active command or presentation-only effect.

### 23.3 Battle Scanner - exact Tactical offensive bonus

`Offensive_Combat_Bonus_` checks Battle Scanner ID 5. Presence contributes exactly **+50** to the shared offensive-combat bonus stack alongside computer, fleet and crew/player contributions.

This is the executable Tactical hit-stack insertion required by the earlier Class-3 matrix. The separate strategic scan-range effect remains strategic and does not belong in BattleSession combat arithmetic.

### 23.4 Multi-Phased Shields and Hard Shields - shield-max / nebula order

`Calc_Combat_Shield_Max_` establishes this order:

1. derive the normal per-facing combat shield maximum;
2. when the combat nebula/global shield-suppression flag is active, zero that shield maximum unless Hard Shields ID 13 is present;
3. then test Multi-Phased Shields ID 20;
4. with ID 20, replace the surviving maximum with signed integer `(3 * base) / 2` semantics.

For positive shield values this is exactly **+50% max shield strength**. Therefore Multi-Phased Shields multiplies the value only after the Hard-Shields/nebula survival decision.

`Apply_Ship_Damage_` also directly proves the general Hard-Shields absorption hook: when shield absorption is active, the normal shield-strength subtraction is followed by an additional **3 damage** subtraction for ID 13, then damage is clamped at zero. Together with the already frozen Shield-Piercing cancellation and Transporter rejection, Hard Shields now has a complete accepted Tactical hook set for Slice 17.8.

### 23.5 Reinforced Hull - exact structure derivation

`Get_Design_Structure_` tests Reinforced Hull ID 27 in the source design Special bitfield. It first derives the normal final structure value, including the ordinary hull/material percentage path, then multiplies that result by **3** when Reinforced Hull is present.

`Get_Ship_Structure_` delegates ordinary ship structure to that derivation and `Load_Combat_Ship_` stores the result as the combat ship's structure maximum. Reinforced Hull therefore triples the final normal structure pool before Tactical damage begins.

### 23.6 Quantum Detonator - explosion power in addition to capture self-destruction

`Damage_From_Ship_Exploding_` derives the normal ship-explosion power as:

`base_explosion_damage = 5 * hull_explosion_value`

and multiplies that value by **3** when Quantum Detonator ID 24 is present.

`Destroy_Ship_FX_` passes this returned value into the shared spherical ship-explosion damage/effect path. `Is_Combat_Ship_Dead_` also gives a dead ID-24 carrier the dedicated detonator death state. This complements the already frozen capture-time `Random_(10)` branch where 0..5 self-destruct and 6..9 continue capture.

Gate 2 therefore needs one shared ship-explosion consumer with a Quantum multiplier; it must not represent the Detonator as only a capture hook.

### 23.7 Automated Repair - exact outer and internal repair budgets/order

`Repair_All_Combat_Ships_` supplies the already frozen repair percentage: Automated Repair Unit ID 3 contributes **20**, then the side Fleet Engineer bonus is added. Destroyed/inactive, post-capture and stasis ships are skipped.

`Repair_Combat_Ship_` uses that percentage in two deterministic passes.

**Durability pass**

- derive the ship's maximum Structure plus maximum Armor pool; Heavy Armor ID 14 is already included by tripling the Armor maximum;
- compute `repair_percent * (structure_max + armor_max) / 100`, with minimum one;
- apply that repair first against accumulated Structure damage;
- if the repair overflows past zero Structure damage, carry the remainder into current Armor, capped at Armor maximum.

**Internal-system pass**

- derive `Max_Total_Internal_Space_`;
- compute `repair_percent * max_internal_space / 200`, with minimum one;
- consume that budget in fixed order:
  1. fixed internal pool `+0x100` (the engine/drive-related pool; damage to this pool immediately triggers `Calc_Current_Speed_`);
  2. fixed internal pool `+0x102`;
  3. Weapon-slot pools `+0x104 .. +0x112`, ascending slot 0..7;
  4. Shield-generator/internal shield pool `+0x134`;
  5. Special/internal-slot pools `+0x124 .. +0x132`, ascending slot 0..7.

Each pool is bounded by its original maximum and consumes only the amount needed before the remaining budget proceeds to the next pool.

This closes the earlier Repair allocation-order note without inventing a parallel percentage or random selection.

### 23.8 Checkpoint-K consequence

The remaining original-formula blockers from the deferred Class-3 matrix are now either:

- directly frozen with executable arithmetic/order; or
- explicitly unsupported/deferred where the audited 1.31 executable does not expose a reachable consumer (Phase-Shifter other-ship teleport).

No accepted Slice-17.8 Tactical behavior now depends solely on HELP prose for its core combat formula. The remaining Gate-1 work is MOOX runtime-contract mapping, persistence/projection/AI requirements and final representative closure vectors.

## 24. Checkpoint L - MOOX runtime contract mapping

Checkpoint K closes the remaining original-executable formula/reachability questions. This checkpoint maps those audited mechanics onto the current MOOX BattleSession architecture without implementing them.

### 24.1 Existing authority surfaces that 17.8 must extend

Current MOOX already has the correct authority boundaries:

- immutable ship combat identity is carried by `TacticalShipSpec`;
- installed Specials are already projected as ordered `TacticalSpecialSpec{Slot, SpecialID, TacticalSupport}`;
- mutable ship combat state is owned by `TacticalShipState`;
- command sequencing and RNG live in `TacticalState`;
- authoritative mutations emit ordered `TacticalEvent` records;
- `PrepareCommand` validates against the current command sequence and works on a cloned runtime before commit;
- `CommitPreparedCommand` rejects stale prepared revisions;
- player views redact RNG state and expose actions only to the active controlling seat;
- `LiveSnapshot` round-trips Tactical state, ordnance state, events and next-event sequence;
- built-in AI consumes the same projected legal-action catalog used by normal BattleSession authority.

17.8 must extend those surfaces. It must not create a second fighter/special simulation beside BattleSession.

### 24.2 Current gaps are explicit

At this checkpoint the current runtime still has these deliberate gaps:

- Fighter/Bomber/Heavy-Fighter weapon rows remain `metadata_only`; no fighter actor state is projected or persisted.
- `TacticalSpecialSpec` is immutable metadata only. Current Tactical support validation accepts only `metadata_only`, `strategic_only` and `turning_mode`.
- `TacticalShipState` has no mutable cloak/phasing counter, Energy-Absorber store, active-Special used flag, automated-repair temporary state, boarding-used flag or captured-owner/control field.
- `TacticalShipView` exposes weapons and ship durability/Marine state but not installed Special state/action availability.
- `TacticalView` has no fighter actor collection, boarding action catalog or active-Special action catalog.
- `PrepareCommand` currently dispatches Move, Beam, Ordnance, End, Wait and Retreat only; it has no Fighter-launch, Boarding or active-Special command.
- `TacticalLiveSnapshot` carries Tactical state, ordnance, events and event sequence only; there is no 17.8 fighter/special transient state to validate or restore.
- the built-in AI battle catalog currently consumes `WeaponActions`, approach Move and End Activation; it has no projected Boarding/Fighter/active-Special decision surface.
- the browser Tactical client consumes projected weapon actions/events and therefore likewise has no authoritative 17.8 action surface yet.

These are implementation gaps, not evidence gaps.

### 24.3 Gate-2 state ownership requirements

Gate 2 should freeze one server-owned 17.8 runtime extension with the following state categories.

**Fighter actor/group state**

Must persist enough information to continue deterministically after save/reconnect:

- stable actor/group ID;
- owner seat/empire;
- carrier ship ID and source weapon slot;
- fighter kind: Fighter / Bomber / Heavy Fighter;
- live craft multiplicity;
- Tactical position/target/return state;
- speed and durability inputs frozen from the launch snapshot;
- payload identity needed by the recovered Beam/Bomb passes;
- lifecycle phase sufficient to distinguish outbound/engaged/returning/recovered/destroyed;
- authoritative RNG/order remains in the shared BattleSession RNG, not actor-local RNG.

The existing Slice-17.7 moving-actor/ordnance substrate should be reused where representation is compatible, but Fighter groups must retain carrier-return/recovery semantics rather than masquerading as disposable missiles.

**Per-ship Special state**

Only transient state proven necessary by the audit should be mutable:

- cloak/phasing state and Phasing lifetime counter;
- Energy Absorber stored energy;
- once-per-turn active-Special used flags for reachable teleport use;
- Hyper-X / Fast-Missile-Rack readiness/rearm state on the relevant weapon row;
- any shield-recharge/repair state that cannot be derived from ordinary ship state at the authoritative phase boundary.

Pure passive Specials such as Augmented Engines, Battle Scanner, Reinforced Hull, Heavy Armor, Hard Shields, Multi-Phased Shields, ECM families, Rangemaster and High Energy Focus should be derived from immutable Special presence plus existing ship state, not duplicated as mutable booleans.

**Boarding/capture state**

Reuse existing `MarineCurrent` / `MarineMax`. Add only what the original lifecycle requires:

- per-combat boarding-used/eligibility state;
- authoritative current owner/control identity if capture changes command ownership in-place;
- post-capture state required by repair/retreat/Special eligibility;
- any capture reconciliation field that cannot be derived from the transferred owner and current Marine state.

### 24.4 Command/action projection requirements

Gate 2 should preserve the current rule that the browser and AI consume **projected legal actions** and never calculate combat legality themselves.

17.8 therefore requires server-projected action catalogs for three semantic families:

1. **Fighter launch/action**
   - only active ship / correct seat;
   - source Fighter weapon row must be executable and ready;
   - legal target must come from server projection;
   - rejection consumes no RNG/readiness;
   - accepted launch creates authoritative actor/group state and consumes the Bay row's launch/readiness state.

2. **Boarding**
   - only active ship / correct seat;
   - enough Marines and unused boarding capability;
   - original normal/Transporter range and facing-shield/Hard-Shield/Damper rules;
   - target must be enemy and capture-eligible;
   - rejection consumes no Marines/RNG;
   - accepted command owns all raid/capture RNG, casualties, ownership transfer and post-capture reconciliation.

3. **Active Special**
   - only Specials with a directly reachable active command are projected;
   - Phase-Shifter self-teleport is eligible; orphaned other-ship teleport is not;
   - legal destination/range/use-state comes from server projection;
   - passive/reactive Specials never appear as fake player commands.

Exact protocol command strings belong to Gate 2; Gate 1 freezes the semantic separation and authority rules.

### 24.5 Rejection contract

17.8 must keep the already accepted Tactical rejection invariant:

- malformed/stale/wrong-seat/wrong-active-ship/unsupported/not-ready/no-target/out-of-range/already-used actions reject before authoritative mutation;
- rejected commands consume zero combat RNG;
- rejected commands do not consume Marines, Bay craft, weapon readiness, Special use state or movement;
- rejected commands do not append success/resolution events;
- a failed prepared-command commit because of revision staleness cannot overwrite newer Tactical state.

Where an unavailable action is useful to the UI, it should be projected with a stable unavailable reason instead of requiring the browser to rediscover legality.

### 24.6 Event vocabulary requirements

17.8 events must describe authoritative mutations, not animation instructions. At minimum the contract needs semantically distinct events for:

**Fighters**
- group launched;
- group moved/retargeted when needed for deterministic replay/presentation;
- fighter attack resolved with enough aggregate/per-craft detail to explain losses/damage;
- fighter group damaged/destroyed;
- return/recovery completed or failed because the carrier no longer exists.

**Boarding**
- boarding action started/resolved;
- attacker/defender Marine casualties;
- raid effect when no capture occurs;
- capture success/repel result;
- ship owner/control transfer;
- Quantum-Detonator capture destruction when triggered.

**Specials**
- active teleport resolved;
- stored Energy-Absorber attack resolved;
- reactive projectile-defense mutation should reuse the existing Slice-17.7 projectile/interception event family where the affected object is the same shared moving actor;
- repair/shield-recharge/passive modifiers need events only when they cause an externally visible mutable change not already represented by an existing generic damage/ship-state event.

Event payloads must contain stable IDs and before/after values sufficient for observer/browser explanation; presentation timing remains client-side.

### 24.7 Player / observer projection

Projection should remain state-first and server-authoritative:

- active seat receives legal 17.8 action catalogs;
- inactive players and observers receive state/events but no executable action catalog for ships they do not control;
- RNG state remains redacted exactly as today;
- installed Special identities needed by the Tactical scan/UI must be projected from the immutable ship spec or an explicit Special view;
- mutable cloak/phasing/Absorber/use-state is projected only to the visibility level required by the original Tactical information model; Gate 2 must make this explicit rather than leaking internal RNG/counters by default;
- Fighter actors/groups require stable public identity, owner, position, multiplicity and lifecycle state sufficient for rendering and target selection.

### 24.8 Persistence / replay / reconnect contract

The existing Battle `LiveSnapshot` is the mandatory persistence boundary.

Gate 2 must extend it so that a save taken at any accepted command boundary round-trips:

- fighter actor/group collection and next actor ID;
- Bay launch/readiness state;
- Fighter target/return/carrier linkage;
- cloak/phasing state and counter;
- Energy Absorber stored energy;
- active-Special used state;
- boarding-used/post-capture/owner-control state;
- existing Marines, shield arcs, weapon readiness, ordnance, RNG state and events.

Restore validation must reject impossible state, including:

- unknown carrier/target/owner IDs;
- Fighter multiplicity outside the frozen launch/survivor bounds;
- recovered/destroyed actors still targetable;
- active Special state on a ship that does not carry the Special;
- Phasing/Absorber/use counters outside their frozen domains;
- Marine counts outside `0..MarineMax`;
- captured ownership/control inconsistent with the authoritative ship state;
- event/command sequence mismatches.

A restored session given the same next legal command must produce byte-equivalent authoritative state/events to the un-restored session.

### 24.9 Built-in AI contract

The built-in AI must not receive a private resolver.

Its Battle decision catalog must consume the same new server-projected actions as a human seat:

- legal Fighter launch/target choices;
- legal Boarding choices;
- legal active-Special choices;
- existing Beam/Ordnance/Move/End choices.

The baseline policy may remain simple, but it must be capable of selecting each supported action family and must never bypass server legality. Original AI evidence already proves Boarding ownership and Warp-Dissipator retreat interaction; Gate 2/3 implementation may choose deterministic baseline priorities without changing the recovered combat formula.

### 24.10 Browser/UI contract

The browser remains presentation-only for combat math.

It may:

- render projected Fighter groups and multiplicity;
- offer projected Fighter/Boarding/active-Special actions;
- display stable unavailable reasons;
- animate from Tactical events;
- show Marine/capture/Special/Fighter state exposed by the server.

It must not:

- calculate Fighter hit/damage/loss RNG;
- decide transporter/shield boarding legality;
- derive Special ranges/cooldowns;
- mutate owner/control locally;
- emulate fighter return/recovery;
- predict hidden RNG to decide availability.

This keeps Slice 17.8 consistent with the accepted 17.5-17.7 authority model.

## 25. Checkpoint M - deterministic closure vectors

These vectors are implementation references assembled only from formulas/order already frozen in Checkpoints A-K. They are not new mechanics.

### 25.1 Passive shield stack vector

Assume normal per-facing shield maximum = 40.

**Nebula/global shield suppression, no Hard Shields**
- suppression reduces max to 0;
- Multi-Phased Shields then applies to the surviving value;
- result remains 0.

**Same ship with Hard Shields**
- Hard Shields preserves the normal 40 through suppression;
- Multi-Phased Shields applies `(3 * 40) / 2 = 60`;
- final per-facing maximum = 60.

For an incoming shielded hit whose current damage has already reached the normal shield-absorption step with 10 damage and shield strength 5:

- normal shield subtraction -> 5;
- Hard Shields additional subtraction -> 2;
- final value continues from 2, clamped at zero if the extra subtraction would underflow.

### 25.2 Energy Absorber -> Damper ordering vector

Incoming current damage = 20.

Energy Absorber:
- absorbed store increment = `20 / 4 = 5`;
- remaining current damage = 15.

Damper Field then quarters the remaining damage:
- `15 / 4 = 3` with the recovered positive integer semantics.

Therefore the combined ordered result is:
- stored Energy Absorber energy +5;
- downstream current damage 3.

Applying Damper first would produce a different result and is therefore invalid.

### 25.3 Reflection Field vector

Current Beam hit damage `D = 20`.

Original reflection bound:
- `Random_(D + 15) = Random_(35)`.

Representative rolls:
- roll 20 -> not strictly greater than D -> no reflection, target continues with 20;
- roll 21 -> strictly greater than D -> target damage becomes 0 and reflected output becomes 20.

The reflected 20 then re-enters the shared ship-damage resolver against the firing ship.

### 25.4 Lightning Field grouped-actor vector

Initial moving-actor multiplicity = 4.

Roll sequence: success, success where success means `<50`.

- attempt 0 succeeds: multiplicity 4 -> 3;
- iteration index becomes 1; 1 < 3, continue;
- attempt 1 succeeds: multiplicity 3 -> 2;
- iteration index becomes 2; 2 < 2 is false, stop.

Result:
- exactly two RNG draws;
- final multiplicity = 2.

An implementation that rolls four times from the original launch count is not original parity.

### 25.5 Shield Capacitor distribution vector

Assume base per-facing shield value = 20 and all four facings are currently 4 points below maximum.

Normal recharge factor 3:
- budget = `20 * 3 / 10 = 6`;
- pointwise scan order yields six one-point repairs;
- with equal deficits this produces increments `[2,2,1,1]`.

Shield Capacitors factor 10:
- budget = `20 * 10 / 10 = 20`;
- only 16 total points are missing;
- all facings reach maximum and the pass stops with unused budget.
### 25.6 Quantum Detonator explosion vector

Assume `hull_explosion_value = 8`.

Normal:
- `5 * 8 = 40` explosion power.

Quantum Detonator:
- `40 * 3 = 120` explosion power.

This is separate from the capture-time `Random_(10)` self-destruction branch.

### 25.7 Automated Repair durability vector

Assume:
- Structure max = 100;
- Armor max = 200;
- effective repair percent = 20;
- current Structure damage = 50;
- current Armor is 180/200.

Durability budget:
- `20 * (100 + 200) / 100 = 60`.

Application:
- first 50 removes all Structure damage;
- 10 remains;
- remaining 10 restores Armor 180 -> 190.

The durability pass does not randomly choose a pool.

For internal repair, if `Max_Total_Internal_Space = 300`:
- internal budget = `20 * 300 / 200 = 30`;
- those 30 points are consumed in the recovered fixed pool order until exhausted.

### 25.8 Augmented Engines / Battle Scanner derivation vector

Assume the normal working combat speed before engine-damage scaling is 8 and both Specials are functional.

Augmented Engines:
- working speed becomes 13 before the later recovered scaling/modifier stages.

Battle Scanner:
- shared offensive-combat bonus receives +50.

Neither effect consumes a command, use-state or RNG.

### 25.9 Fighter / Boarding references

The dedicated earlier vectors remain authoritative and are not duplicated:

- Fighter/Bomber/Heavy-Fighter launch multiplicity, speed, durability, payload and return/loss vectors: Checkpoint C;
- Fighter Beam hit/damage/RNG vectors: Checkpoint E;
- Boarding raid/capture differential, casualty and survivor-transfer vectors: Checkpoints B-D.

Together with Sections 25.1-25.8 these satisfy the representative Fighter, Bomber, Boarding and passive/active/reactive Special vector requirement.

## 26. Gate-1 completion review

Re-evaluating the Gate-1 criteria from Section 6:

- **Fighter/Bomber/Heavy-Fighter identity and lifecycle:** COMPLETE from direct actor substrate, multiplicity, speed/durability, payload, target, loss, return/recovery and carrier-loss evidence.
- **Class-3 per-row classification:** COMPLETE; every deferred row is classified and its ownership/support state is explicit.
- **Accepted Tactical Special formula/order/RNG/target semantics:** COMPLETE for accepted 17.8 behavior. Unreachable Phase-Shifter other-ship teleport is explicitly unsupported/deferred rather than guessed.
- **Boarding/Marine semantics:** COMPLETE, including normal/Transporter legality, shields/Hard Shields/Damper, force selection, RNG differential, casualties, raid/capture, Security Stations, Quantum Detonator, survivor transfer and capture reconciliation.
- **Persistence/replay/Observer/UI/AI requirements:** COMPLETE as MOOX runtime-contract mapping in Checkpoint L.
- **Deterministic representative vectors:** COMPLETE across Fighter/Bomber, Fighter Beam, Boarding, Lightning, Reflection, shield stack/recharge, Absorber/Damper, Quantum, repair and passive derivation.
- **Unsupported behavior guessed:** NONE. HELP/executable discrepancies are retained explicitly and unreachable behavior is deferred.

### 26.1 Gate-1 result

**Slice 17.8 Gate 1 is COMPLETE.**

Gate 1 closes as an audit/evidence gate only. No production implementation has been started.

The next allowed step is **Gate 2 - contract freeze**, which must translate this evidence into exact MOOX schemas, protocol names, state domains, event payloads, ordering hooks, persistence validation and implementation checkpoints before any 17.8 production mechanics are added.

Gate 2 must in particular preserve:

- reuse of 17.6 Beam and 17.7 moving-actor/ordnance authority;
- one BattleSession RNG/order stream;
- no client-side combat authority;
- no second parallel Special representation;
- explicit Warp-Dissipator row-37 data projection repair;
- explicit unsupported/deferred Phase-Shifter other-ship teleport;
- save/restore determinism for every new mutable 17.8 field.
## 27. Gate-2 validation correction - Raid Marine persistence

During Gate-2 contract validation, the direct caller/tail path around `Resolve_Raid_` was traced one level farther than the Gate-1 summary.

This **refines persistence only**; it does not change the already-frozen Raid RNG thresholds or internal-damage behavior.

Direct sequence:

1. `Initiate_Boarding_Action_` subtracts the selected boarding party from the source Ship's `MarineCurrent` before calling the resolver;
2. `Resolve_Raid_` uses temporary attacker/defender Marine totals while evaluating the frozen Raid differential and internal-damage branches;
3. before returning, `Resolve_Raid_` restores the target Ship's `MarineCurrent` to its saved pre-Raid value and writes the surviving/returned party value through the caller-owned party pointer;
4. the caller then adds that returned party value back to the source Ship's `MarineCurrent`.

Therefore the original Raid contract for Slice 17.8 is:

- Raid Marine losses shown/used inside the resolver are **temporary resolution accounting**, not durable Tactical Marine casualties;
- the target Ship's persistent `MarineCurrent` is unchanged by a completed Raid;
- the returned boarding party is reconciled back to the source Ship;
- persistent Raid output is the selected internal-system damage/state mutation;
- Capture remains the Boarding mode that can persist Marine survivor transfer and ownership change.

This correction supersedes any earlier Gate-1 wording that could be read as Raid permanently killing source/target Marines. Gate-1 completion remains valid; Gate 2 freezes the corrected persistence semantics.