# Slice 17.8 Gate 2 - Tactical Fighters / Boarding / Combat Specials Contract Freeze

Date: 2026-10-05
Status: **FROZEN / COMPLETE - Gate 3 implementation ready / not started**

Parent: Slice 17 - Military Ship Design + Tactical Combat
Child: Slice 17.8 - Tactical Fighters / Bombers / Combat Specials

Evidence authority:

- `docs/research/SLICE_17_8_GATE1_FIGHTERS_SPECIALS_TACTICAL_AUDIT_2026-10-02.md`
- frozen Slice 17.4 Fighter/Special design contract
- frozen Slice 17.5 Tactical foundation contract
- frozen Slice 17.6 Tactical Beam contract
- frozen Slice 17.7 Tactical ordnance / moving-actor contract
- current BattleSession command/event/live-snapshot conventions

This document freezes the implementation/API/event/persistence contract for Gate 3. Gate 3 may implement production Slice-17.8 behavior only to match this contract; it must not invent new gameplay semantics.

## 1. Authority and slice boundary

### 1.1 Sole mutable gameplay authority

`BattleSession` remains the sole mutable Tactical gameplay authority.

The browser:

- never advances Fighter/craft actors;
- never computes Fighter Beam/Bomb hit, damage or loss RNG;
- never decides Boarding legality or Marine casualties;
- never decides Capture ownership transfer;
- never computes Special range, cooldown/readiness, reflection, absorption, cloak/phasing, repair or shield recharge;
- never advances Tactical RNG;
- never performs fighter recovery or carrier-loss cleanup.

React renders server-projected state/actions and animates authoritative Tactical events only.

### 1.2 Slice ownership

Slice 17.8 owns:

- Fighter / Bomber / Heavy-Fighter execution and lifecycle;
- Fighter payload resolution and return/recovery;
- Boarding Raid/Capture legality, casualties and ownership reconciliation;
- active/passive/reactive Combat-Special consumers frozen by Gate 1;
- mutable Special state, internal-system state required by accepted 17.8 effects, persistence/replay and UI/AI projection.

Slice 17.6 continues to own Beam hit/damage/RNG and the shared ship-damage pipeline.
Slice 17.7 continues to own ordinary Missile/Torpedo actor semantics and reactive Beam-PD primitives.
Slice 19 remains the owner of planetary Fighter Garrison and planetary/orbital combat.

## 2. Tactical support vocabulary

### 2.1 Fighter weapon support

The three 17.4 Fighter weapon rows become executable with:

- `TacticalSupport = "fighter"`

for:

- `fighter_bays`;
- `bomber_bays`;
- `heavy_fighter_bays`.

`TacticalWeaponSpec` gains `craft_per_unit`; for all three accepted rows it is 4.

### 2.2 Special support modes

`TacticalSpecialSpec.TacticalSupport` accepts these additive execution identities:

- `tactical_passive` - automatic/passive/reactive server consumer, no direct player command;
- `tactical_active` - server-projected active Special command;
- existing `turning_mode`, `strategic_only`, `metadata_only` remain valid.

Accepted 17.8 rows freeze as:

| Special | Support | Contract role |
|---|---|---|
| Achilles Targeting Unit | tactical_passive | Beam/internal damage selection |
| Augmented Engines | tactical_passive | combat speed derivation |
| Automated Repair Unit | tactical_passive | round bookkeeping repair |
| Battle Scanner | tactical_passive | offensive combat bonus |
| Cloaking Device | tactical_passive | cloak state / defense |
| Damper Field | tactical_passive | damage + Boarding hooks |
| Displacement Device | tactical_passive | forced-miss defense |
| ECM Jammer | tactical_passive | missile defense |
| Energy Absorber | tactical_passive | stored damage / automatic discharge |
| Fast Missile Racks | tactical_passive | ordnance readiness state |
| Hard Shields | tactical_passive | shield / Boarding hooks |
| Heavy Armor | tactical_passive | armor/internal durability |
| High Energy Focus | tactical_passive | Beam damage hook |
| Hyper-X Capacitors | tactical_passive | Beam readiness state |
| Lightning Field | tactical_passive | pre-impact moving-actor defense |
| Multi-Phased Shields | tactical_passive | shield maximum |
| Multi-Wave ECM Jammer | tactical_passive | missile defense |
| Phase Shifter | tactical_active | self-teleport only; other-ship routine deferred |
| Phasing Cloak | tactical_passive | phased/cloak lifecycle |
| Quantum Detonator | tactical_passive | capture/death explosion hooks |
| Rangemaster Unit | tactical_passive | Beam range hook |
| Reflection Field | tactical_passive | Beam reflection |
| Reinforced Hull | tactical_passive | Structure/internal durability |
| Security Stations | tactical_passive | Boarding defense |
| Shield Capacitors | tactical_passive | shield recharge budget |
| Structural Analyzer | tactical_passive | post-Shield damage hook |
| Sub-Space Teleporter | tactical_active | self-teleport |
| Time Warp Facilitator | tactical_passive | extra-turn scheduler |
| Transporters | tactical_passive | Boarding legality/range modifier |
| Troop Pods | tactical_passive | MarineMax x2 |
| Warp Dissipator | tactical_passive | Retreat legality |
| Wide Area Jammer | tactical_passive | self/fleet missile defense |

Inertial Stabilizer / Inertial Nullifier remain `turning_mode` and keep the existing consumer.

### 2.3 Explicit deferred boundary

Phase-Shifter other-ship teleport is not projected as an action. The audited `Teleport_Enemy_Ship_` routine has no accepted reachable dispatch in the 1.31 executable.

Anti-Missile Rockets retain the frozen 17.7 `projectile_interceptor` hook, range and hit formula, but current MOOX has no faithful source representation for the original special-weapon live count consumed by `Anti_Missile_Rockets_`. Gate 3 must not invent a charge count. The row remains metadata/deferred until that source-state mapping is represented explicitly.

## 3. Ruleset / snapshot rule blocks

The tactical-combat rules schema remains version 1 for additive ruleset fields. Battle live-snapshot schema changes separately in Section 15.

`TacticalRulesSnapshot` gains optional server-owned 17.8 blocks copied into every new Tactical battle:

### 3.1 Fighter rules

`TacticalFighterRulesSnapshot` freezes:

- `movement_units_per_field = 20`;
- `craft_per_bay = 4`;
- base speed: Fighter 10, Bomber 8, Heavy Fighter 8;
- base hits: Fighter 2, Bomber 4, Heavy Fighter 5;
- attack stock: Fighter 4, Bomber 1, Heavy Fighter 2;
- warp-drive speed step `+2` per best-drive index above Nuclear;
- Trans-Dimensional speed bonus `+4`;
- armor durability multiples: Titanium 1, Tritanium 2, Zortrium 4, Neutronium 6, Adamantium 8, Xentronium 10;
- recovery divisor `4` complete craft per Bay group.

### 3.2 Boarding rules

`TacticalBoardingRulesSnapshot` freezes:

- physical Boarding range `1`;
- Transporter Boarding range `4`;
- selectable party maximum `(MarineCurrent + 1) / 2`;
- Damper Transporter divisor `3`;
- Security Stations defender bonus `+20`;
- Raid strict thresholds `>20`, `>50`, `<-50`;
- Capture Quantum bound `Random(10)` with self-destruction for results `0..5`;
- crew-quality bonuses `0,5,10,15`;
- Fleet Commando/Security officer formula families exactly as frozen by Gate 1.

### 3.3 Special rules

`TacticalSpecialRulesSnapshot` carries the scalar constants required for replay without consulting mutable external rules, including at least:

- Augmented Engines +5 speed;
- Battle Scanner +50 offense;
- Cloak +80 Beam Defense and `<50` Missile forced-miss threshold;
- Phasing initial lifetime 10;
- Energy Absorber divisor 4;
- Damper damage divisor 4;
- Displacement ordnance `<30` / Beam `<=30` forced-miss branch identities;
- Shield Capacitor normal/special recharge factors 3/10;
- Hard Shields additional shield absorption 3;
- Multi-Phased Shield ratio 3/2;
- Lightning Field `<50` destruction threshold;
- Quantum explosion multiplier 3;
- Automated Repair base percent 20;
- self-teleport maximum Range units 20.

Code may use typed helpers, but these replay-relevant constants must come from the Tactical rules snapshot or an equivalent immutable snapshot block, not live rules files.

## 4. Fighter doctrine snapshot

Fighter payload/speed/durability depend on empire-wide best equipment and race state, not merely the carrier's installed drive/armor.

`TacticalSpec` therefore gains ordered `FighterDoctrines []TacticalFighterDoctrineSpec`, one per participating seat/empire that can launch craft.

Each record freezes:

- `seat_id`;
- `empire_id`;
- `best_warp_drive_index` (1..6);
- `armor_durability_multiple` (1,2,4,6,8,10);
- `trans_dimensional`;
- `fighter_beam_offense` - immutable fully projected Fighter OCV scalar consumed by Fighter Beam replay;
- `moving_actor_defense_bonus` - immutable additive Fighter moving-actor DCV term applied after `speed * 5`;
- optional Fighter Beam payload identity plus base min/max damage;
- optional Fighter Bomb payload identity plus base min/max damage.

Rows are sorted by seat ID. Fighter payload identity is server-owned and never selected by the browser at launch time.

## 5. Internal-system snapshot and live state

Accepted 17.8 behavior requires internal-system state for Achilles, Raid damage, Automated Repair and damage-driven speed/weapon/Special/shield consequences.

### 5.1 Immutable internal maxima

`TacticalShipSpec` gains `Internal TacticalInternalSpec` with:

- `engine_max` - the proven speed-affecting fixed pool;
- `secondary_max` - the second fixed original pool; no unsupported semantic label is invented;
- eight ordered Weapon-slot records with `slot`, `max`, `unit_size`;
- eight ordered Special-slot records with `slot`, `max`;
- `shield_generator_max`.

Projection uses the original combat-load derivation. Direct validation during Gate 2 confirms:

- Weapon `unit_size = max(1, One_Weapon_Space_Without_M_Level / 10)`;
- Weapon pool max = live installed weapon count * unit size;
- damageable installed Special pool max = `max(1, original Special space / 10)`;
- original non-damageable Special IDs 27, 11, 4, 14 and 24 do not receive ordinary Special-slot pool weight;
- Shield-generator/fixed pools come from the original hull/system combat-load derivation;
- Reinforced Hull participates in the proven fixed/Structure durability derivation.

### 5.2 Mutable internal state

`TacticalShipState` gains `Internal TacticalInternalState` containing current values for the same pools.

All current values initialize from the immutable maxima for a fresh BattleSession unless an accepted strategic-damage source explicitly projects a reduced initial value.

Damage semantics freeze:

- weighted internal selection uses current surviving pool values;
- Achilles multiplies all eight Weapon-slot weights and the Shield-generator weight by 3 before the shared weighted-choice RNG;
- Engine damage immediately recalculates current combat speed;
- Weapon damage recomputes live physical weapon count as ceiling(current pool / unit_size);
- a Special pool reaching zero marks that installed Special non-functional for consumers that require functionality;
- Shield-generator pool reaching zero clears active shield capacity/facings through the shared shield state.

Internal current state is persisted and projected; no browser-side subsystem simulation is allowed.
## 6. Fighter Bay and craft-group live state

### 6.1 Fighter Bay state

Gate 3 adds ordered `TacticalFighterBayState` records for executable Fighter weapon rows:

- `ship_id`;
- `weapon_slot`;
- `ready_groups`.

`ready_groups` initializes to the row's immutable Bay Count.

One accepted launch from a row:

1. requires `ready_groups > 0` and ordinary activation readiness;
2. launches one grouped actor with `multiplicity = ready_groups * craft_per_unit`;
3. sets that row's `ready_groups = 0`;
4. consumes the row's ordinary activation readiness.

`ready_groups` is not refilled by a generic new-round reset. Only Fighter recovery restores complete groups.

### 6.2 Fighter group state

Gate 3 adds `TacticalFighterGroupState` with frozen JSON semantics:

- `group_id` - non-zero monotone battle-local Fighter-group ID;
- `spawn_order` - non-zero monotone moving-actor order key;
- `source_seat_id`;
- `source_ship_id`;
- `source_weapon_slot`;
- `kind` - `fighter`, `bomber`, or `heavy_fighter`;
- exactly one target identity: `target_ship_id`, `target_projectile_id`, or `target_fighter_group_id`;
- `x`, `y`, `facing`;
- `speed`;
- `movement_remaining`;
- `distance_traveled`;
- `multiplicity`;
- `unit_durability`;
- `intercept_damage_remainder`;
- `attack_stock`;
- `phase` - `outbound` or `returning`.

Only active unresolved groups remain in the list.

### 6.3 Shared moving-actor order

Gate 3 must preserve one deterministic order across existing ordnance actors and new craft actors.

Add `spawn_order` to projectile state and a BattleSession-level `next_moving_actor_order` counter. New Missile/Torpedo projectiles and Fighter groups allocate from the same counter.

For legacy ordnance snapshots, migration assigns each existing projectile `spawn_order = projectile_id` in ascending projectile order and initializes the next order above the maximum.

At a moving-actor phase, projectiles and Fighter groups are merged and processed by ascending `spawn_order`. This is the MOOX representation of the original shared moving-actor substrate and avoids independent per-family RNG scheduling.

### 6.4 Movement schedule

Reuse the accepted Slice-17.7 movement quantum:

- one field = 20 movement units;
- a newly launched craft actor receives one immediate movement quantum;
- every still-active moving actor receives one quantum at deterministic new-round processing;
- actor processing never waits for browser animation.

Fighter movement uses its frozen `speed * 20` budget and the same server-owned geometry/retarget helpers as the shared moving-actor foundation.

## 7. Fighter target / attack / recovery contract

### 7.1 Target legality

`battle.launch_fighters` receives server-projected legal targets only.

Target classes:

- Fighter / Heavy Fighter may target a legal enemy Ship or accepted hostile moving actor;
- Bomber targets legal enemy Ships only;
- target identities are stable server IDs; the browser does not derive actor eligibility.

Lost/invalid hostile targets use the shared deterministic retarget path. If no legal hostile target remains, the group transitions toward return when its lifecycle permits; it is never client-retargeted.

### 7.2 Pre-attack defense

When a Fighter group reaches its current target, moving-actor defenses execute before its payload:

1. validate/retarget target;
2. accepted Slice-17.8 pre-impact defenses such as Lightning Field;
3. reactive Point-Defense Beam handoff through the existing Slice-17.6/17.7 capability where the group is a legal moving-actor target;
4. if the group survives, execute its payload cycle.

Beam damage against Fighter groups uses the same multiplicity/durability/remainder arithmetic already frozen for moving actors.

### 7.3 Payload cycles

For each surviving craft in group order:

- Fighter: one Fighter-Beam pass, then decrement attack stock once;
- Bomber: one Fighter-Bomb pass, then decrement attack stock once;
- Heavy Fighter: Fighter-Beam pass followed by Fighter-Bomb pass, then decrement attack stock once.

Each surviving craft consumes its own frozen hit/damage RNG where required.

Fighter Beam uses the Gate-1 Fighter OCV / target defense / `Random(100)` / damage interpolation contract and the existing shared ship/moving-actor damage applications.

Fighter Bomb uses the frozen Gate-1 per-craft Bomb resolver and payload snapshot.

### 7.4 Return and carrier loss

When `attack_stock` reaches zero:

- if the source carrier remains functional and ownership-compatible, phase becomes `returning` and target becomes the source carrier;
- otherwise remove the group with carrier-loss cleanup reason.

Carrier destruction does not synchronously erase a still-outbound Fighter group; cleanup occurs when its own lifecycle reaches the return requirement.

### 7.5 Recovery

When a returning group reaches its functional source carrier:

1. `full_groups = multiplicity / 4`;
2. discard any craft remainder below a complete group;
3. refill the source row first up to its immutable configured Bay Count;
4. spill remaining complete groups into other matching Fighter weapon rows of the same weapon ID in ascending slot order, each capped by its immutable Count;
5. reconcile row readiness without exceeding activation legality;
6. remove the recovered moving actor.

Recovery is an authoritative state mutation and event; no browser-side Bay refill exists.

## 8. Boarding command and immediate-resolution contract

Boarding is an explicit Tactical command surface but does **not** create a persistent Assault-Shuttle actor in the accepted 17.8 contract.

The audited `Initiate_Boarding_Action_` path validates the target, subtracts the selected party, then dispatches Raid/Capture directly. The original moving-actor type `0x1C` remains reference evidence only and is not promoted into a MOOX executable row without a separate supported design/source contract.

### 8.1 Command

Freeze:

`CommandBoardShip = "battle.board_ship"`

with `BoardShipPayload`:

- `ship_id`;
- `target_ship_id`;
- `marine_count`;
- `mode` - `raid` or `capture`.

### 8.2 Common legality

An accepted Boarding command requires:

- acting Ship is current, functional and controlled by the submitting seat;
- target is a live enemy Ship;
- `marine_count > 0` and `marine_count <= (MarineCurrent + 1) / 2`;
- the source has not exhausted the applicable Boarding-use state;
- legality is server-projected before submission.

Selected `marine_count` is subtracted from source `MarineCurrent` atomically before Raid/Capture resolution.

### 8.3 Physical Boarding

Without functional Transporters:

- range must be `<= 1`;
- target effective current combat speed must be zero;
- resolution occurs immediately at the command boundary; no persistent shuttle actor is created.

### 8.4 Transporter Boarding

With functional Transporters:

- range must be `<= 4`;
- target immobility is not required;
- determine the facing Shield arc from relative bearing;
- active facing Shield blocks the action;
- Hard Shields blocks Transporter Boarding regardless;
- Damper Field integer-divides the selected effective boarding party by 3 before resolution;
- Transporter/use state is consumed through the persisted Boarding/Special flags.

### 8.5 Raid persistence correction

Gate-2 call-chain validation closes a subtle state question that the Gate-1 prose described too broadly.

`Resolve_Raid_` uses temporary attacker/defender Marine counts for its RNG/internal-damage loop, then restores the target Ship's original `MarineCurrent` and restores the boarding-party count passed by pointer. The caller subsequently adds that returned party count back to the source Ship.

Therefore accepted Raid persistence is:

- selected Marines leave the source only for the duration of the resolution;
- the selected party is returned to the source after Raid;
- target `MarineCurrent` is restored to its pre-Raid value;
- persistent Raid outcome is the resolved internal-system damage/state changes, not durable Marine casualties;
- Raid never transfers ownership.

### 8.6 Capture persistence

Capture does not return the selected party to the source.

- if attacker survivors remain, target `MarineCurrent` becomes exactly the surviving attacking count before `Capture_Ship_`;
- if attacker survivors reach zero, ownership does not transfer and target `MarineCurrent` becomes the surviving defender count;
- successful capture then follows the frozen officer removal -> Quantum Detonator -> owner/control transfer -> inbound retarget -> continuing-effect cleanup -> post-capture-state ordering.

### 8.7 Boarding usage state

`already_boarded`, `being_captured` and the relevant teleporter/use flag are authoritative persisted state. Gate 3 must follow the audited helper transitions and may not infer these flags from MarineCurrent.

## 9. Mutable per-ship 17.8 state

`TacticalShipState` gains additive server-owned fields:

- `owner_empire_id`;
- `controller_seat_id`;
- `cloak_state`;
- `phasing_turns_remaining`;
- `energy_absorber_stored`;
- `teleporter_used`;
- `already_boarded`;
- `being_captured`;
- `post_capture_state`;
- `internal` state from Section 5.

Owner/controller initialize from immutable `TacticalShipSpec.EmpireID/SeatID`. Tactical views and action authority use the mutable owner/controller after capture.

### 9.1 Cloak state vocabulary

Freeze semantic states:

- `none`;
- `cloaked`;
- `cloak_broken`;
- `cloak_recovering`;
- `phased`;
- `phase_broken`;
- `phase_recovering`.

Transitions map exactly to the audited original 0/1/2/3/4/6/7 lifecycle. `phasing_turns_remaining` is valid only for the Phasing-Cloak lifecycle, initializes to 10 and decrements only at the frozen phase transition.

### 9.2 Energy Absorber

`energy_absorber_stored` initializes zero.

On incoming accepted ship damage:

1. store `max(1, incoming_damage / 4)`;
2. subtract the same amount from current incoming damage;
3. then apply Damper Field quartering if active.

Stored energy is automatically eligible for the frozen pre-weapon discharge against a Ship at range `<8`. It is cleared after discharge or at the end of the owning side pass if unused.

### 9.3 Teleporter / Boarding flags

`teleporter_used` resets at each eligible ship turn start and is shared by the directly evidenced self-teleport use path.

`already_boarded` and `being_captured` are persisted temporary Boarding state and are never inferred from MarineCurrent alone.

`post_capture_state` is set for normal non-Telepathic captured ships and bypassed for Telepathic new owners; Gate 3 must preserve the audited start-turn consequences without inventing an unsupported clear lifetime.

## 10. Hyper-X / Fast Missile Racks weapon-state extension

`TacticalWeaponState` gains optional `special_readiness` with enum:

- `none`;
- `extra_volley_ready`;
- `rearm_pending`.

Mapping of the original 0/1/2 state:

- original 0 -> `none`;
- original 1 -> `extra_volley_ready`;
- original 2 -> `rearm_pending`.

Hyper-X applies this state only to executable Beam rows. Fast Missile Racks applies it only to executable Missile rows.

Combat initialization arms eligible rows. Firing consumes the extra volley according to Gate-1 order. A complete non-firing eligible turn advances surviving `rearm_pending` back to `extra_volley_ready` through the accepted end-pass transition.

Ordinary weapon readiness remains separate; `special_readiness` must not be overloaded into the existing `Ready` boolean.
## 11. Command and action-projection contract

### 11.1 Fighter launch command

Freeze:

`CommandLaunchFighters = "battle.launch_fighters"`

with `LaunchFightersPayload`:

- `ship_id`;
- `weapon_slot`;
- optional `target_ship_id`;
- optional `target_projectile_id`;
- optional `target_fighter_group_id`.

Exactly one target identity must be non-zero.

Existing `TacticalWeaponActionAvailability` remains the projection surface for Fighter rows:

- `command_kind = "battle.launch_fighters"`;
- `weapon_family = "fighter"`;
- targets are server-projected according to Section 7;
- a row with no complete Bay groups is unavailable.

Add stable unavailable reasons:

- `fighter_bays_empty`;
- `fighter_target_unavailable`;
- existing `weapon_not_ready`, `no_legal_targets`, `tactical_consumer_unavailable` remain valid.

Rejected launch commands consume zero RNG, mutate no Bay/readiness/actor state and append no gameplay event.

### 11.2 Boarding action projection

`TacticalView` gains ordered `BoardingActions []TacticalBoardingActionAvailability`.

Each record freezes:

- `ship_id`;
- `target_ship_id`;
- `delivery_mode` - `physical` or `transporter`;
- `max_marine_count`;
- `allowed_modes` containing `raid`, `capture` as legal;
- `available`;
- optional stable `unavailable_reason`.

Add stable reasons:

- `boarding_already_used`;
- `boarding_no_marines`;
- `boarding_target_unavailable`;
- `boarding_target_moving`;
- `boarding_out_of_range`;
- `transporter_shield_blocked`;
- `hard_shields_block_transporters`;
- `boarding_special_disabled`.

The browser chooses only mode and a Marine count within the server-projected maximum.

### 11.3 Teleporter movement uses the existing Move command

No generic `battle.use_special` command is introduced for ID 22/33.

Gate-1 evidence shows self-teleport is a `Move_Ship_` path. Gate 3 therefore extends `TacticalMoveOption` with:

- `movement_kind` - `normal` or `teleport`.

For functional Phase Shifter or Sub-Space Teleporter with unused teleporter state:

- legal teleport destinations use maximum Range units 20;
- ordinary movement points are not required or subtracted;
- facing is preserved;
- accepting the teleport sets `teleporter_used`;
- `ship_moved` remains the authoritative event and gains `movement_kind`.

Other-ship Phase-Shifter teleport is never projected.

### 11.4 Command rejection invariant

For Fighter launch, Boarding and teleport Move:

- wrong phase/seat/active ship;
- stale sequence/revision;
- malformed payload;
- unsupported/dead/disabled source;
- unavailable target/destination;
- exhausted Bay/use/readiness state;
- illegal Marine count/range/shield state

must reject before authoritative mutation.

Rejected commands consume zero RNG, append zero gameplay events and consume zero Marines/Bay groups/readiness/movement/Special state.

## 12. Beam target transport extension for Fighter groups

Fighter groups are accepted moving-actor Beam targets.

`TacticalFireTarget` and `FireBeamPayload` gain optional `target_fighter_group_id`.

For normal player Beam fire exactly one of:

- `target_ship_id`;
- `target_projectile_id`;
- `target_fighter_group_id`

must be non-zero.

Fighter-group Beam targeting reuses the accepted moving-actor contract:

- Standard and Point Defense mounts may target the group;
- Heavy mount is unavailable against moving actors;
- target Beam Defense input is the group's derived moving-actor DCV path;
- Slice 17.6 owns Beam RNG/damage;
- 17.8 applies returned damage to group durability/multiplicity/remainder.

Reactive Point Defense against a group remains an internal BattleSession capability, not a protocol command.

## 13. Boarding Raid / Capture resolver contract

### 13.1 Raid

Static Raid modifiers:

- attacker = Fleet Commando bonus + attacking ship Crew Boarding bonus;
- defender = Fleet Commando bonus + Fleet Security bonus + Security Stations 20 + defending ship Crew Boarding bonus.

For each temporary boarding-party iteration:

`diff = attacker_modifier + Random(100) - defender_modifier - Random(100)`

Strict branches:

- `diff > 20` -> choose one internal target and apply **two** internal-damage selections;
- `diff > 50` -> temporary defender Marine decrement for Raid resolution accounting;
- `diff < -50` -> temporary attacker-party decrement.

At Raid completion, temporary Marine counts are restored as frozen in Section 8.5. Internal-system damage persists.

### 13.2 Capture

Capture modifiers:

- attacker: Fleet Commando bonus plus the original player-wide Ground-Combat stack;
- defender: Fleet Commando + Fleet Security + Security Stations 20 plus the original player-wide Ground-Combat stack.

Capture does not directly add ship Crew Boarding bonus outside the original Ground-Combat input path.

Each Ground-Combat round consumes one `Random(100)` per side. A tie advances casualty progress on both sides exactly as frozen by Gate 1.

Capture result ordering:

1. finish Ground-Combat casualty loop;
2. write target `MarineCurrent` from attacker survivors on success or defender survivors on repel;
3. on success remove attached officer / assignment;
4. clear captured-ship Boarding transient state required by the original capture path;
5. if Quantum Detonator exists, consume `Random(10)`; `0..5` destroys the ship and ends capture reconciliation;
6. if it survives, update mutable `owner_empire_id` / `controller_seat_id`;
7. retarget inbound projectiles/moving actors according to the accepted shared retarget rules;
8. remove continuing-effect Specials as frozen;
9. set normal `post_capture_state` unless the new owner is Telepathic;
10. reconcile action/initiative/ownership state.

Ownership mutation never edits the immutable source design/spec identity.

### 13.3 Retreat interaction

Warp Dissipator is checked by the server for both manual and built-in AI retreat.

A functional enemy carrier of Warp Dissipator makes `battle.retreat` unavailable/rejected. Destroyed/inactive/post-capture-invalid carriers do not block retreat.

## 14. Frozen Special hook order

Gate 3 must activate Special effects only at the hook positions already owned by 17.6/17.7 or the round/ship lifecycle.

| Special family | Frozen hook/order |
|---|---|
| Augmented Engines | +5 working combat speed before later engine-damage scaling |
| Battle Scanner | +50 shared offensive-combat bonus |
| Rangemaster Unit | divide Beam range input by 3 before Beam range To-Hit lookup |
| High Energy Focus | +50 Beam damage multiplier at frozen 17.6 hook |
| Achilles | force Armor bypass; x3 Weapon-slot + Shield-generator internal weights; no extra RNG |
| Structural Analyzer | x2 downstream post-Shield damage |
| Hard Shields | preserve shield through nebula suppression; cancel Shield Piercing; subtract extra 3 during active shield absorption; block Transporter Boarding |
| Heavy Armor | triple Armor capacity and resist ordinary Armor Piercing as frozen |
| Multi-Phased Shields | apply `(3 * surviving_shield_max) / 2` after Hard-Shields/nebula decision |
| Shield Capacitors | recharge factor 10 instead of normal 3; pointwise facing scan order |
| Automated Repair | once-per-round bookkeeping; 20 + Fleet Engineer percent; durability then fixed internal-pool order |
| Energy Absorber | store 1/4 incoming before Damper; automatic pre-weapon discharge; clear end own pass if unused |
| Damper Field | quarter remaining incoming after Absorber; suppress Beam Marine-kill; Transporter party /3 |
| Reflection Field | Beam branch `Random(D+15) > D` reflects full current D to attacker |
| Displacement Device | ordnance forced miss on `<30`; Beam forced miss on `<=30`; separate RNG at frozen target-side branch |
| ECM / Multi-Wave / Wide Area | +70 / +100 / +130 self; Wide Area +70 fleet; best-only stacking |
| Lightning Field | pre-impact grouped-actor loop before reactive PD/MIRV; `<50` decrements live multiplicity and loop bound uses mutated count |
| Cloaking Device | cloak state gives +80 Beam Defense and `<50` Missile forced miss; fire/boarding breaks cloak; lifecycle transitions as frozen |
| Phasing Cloak | phased untargetability state; initial 10-turn counter; broken/recovery transitions as frozen |
| Hyper-X / Fast Missile Racks | weapon `special_readiness` state machine from Section 10 |
| Phase Shifter / Sub-Space Teleporter | existing Move command teleport path, once per eligible ship turn, range 20, no move-cost subtraction, facing preserved |
| Time Warp Facilitator | second authoritative side pass after both normal passes; only functional ID34 ships; normal start-turn reset; Seeking Missiles skipped in extra pass |
| Quantum Detonator | 60% capture-time destruction before owner transfer; x3 ordinary ship-explosion power |
| Security Stations | +20 defender Boarding modifier |
| Transporters | Boarding range/shield/Hard-Shield/Damper contract from Section 8 |
| Troop Pods | initial MarineMax x2 only; current Marines follow normal state |
| Reinforced Hull | triple final normal Structure pool plus frozen internal durability contribution |
| Warp Dissipator | authoritative retreat legality |

Global shield recharge and Automated Repair remain once-per-round after normal + Time-Warp passes; Time Warp does not duplicate them.
## 15. Event contract

All new events use the existing `TacticalEvent` envelope and monotonically increasing sequence. Events describe authoritative mutations, not animation timing.

### 15.1 Fighter events

Freeze:

#### `fighter_group_launched`

Required data:

- group/spawn order IDs;
- source seat/ship/weapon slot;
- fighter kind;
- target identity;
- start x/y/facing;
- speed, multiplicity, unit durability, attack stock;
- Bay `ready_groups` after launch;
- frozen Beam/Bomb payload identities where applicable.

#### `fighter_group_moved`

- group ID;
- from/to x/y;
- movement used/remaining;
- distance traveled;
- target identity;
- phase.

#### `fighter_group_retargeted`

- group ID;
- previous/new target identity;
- reason.

#### `fighter_attack_resolved`

- group ID and fighter kind;
- target identity;
- attack stock before/after;
- multiplicity participating;
- payload kind(s);
- per-cycle hit/damage/loss aggregate sufficient for UI/replay correlation.

Per-craft ship damage continues to use `battle_damage_applied` with additive `fighter_group_id`, `craft_index`, `payload_kind` identity. Moving-actor damage uses the appropriate actor-side result event.

#### `fighter_group_returning`

- group ID;
- carrier Ship ID;
- surviving multiplicity.

#### `fighter_group_recovered`

- group ID;
- surviving multiplicity;
- complete groups recovered;
- discarded incomplete craft remainder;
- affected Bay slots and ready-group before/after values.

#### `fighter_group_destroyed`

- group ID;
- reason: `intercepted`, `carrier_unavailable`, `no_target`, or other explicitly frozen lifecycle reason;
- multiplicity before removal.

### 15.2 Moving-actor interception

Existing `projectile_interception_resolved` remains projectile-specific.

Add `fighter_interception_resolved` for damage against Fighter groups with the same durability facts:

- group ID;
- interceptor source kind/ship/slot where applicable;
- damage;
- destroyed craft;
- multiplicity before/after;
- remainder before/after.

Lightning Field against an ordnance projectile continues to use the existing projectile event with `interceptor_source_kind = "lightning_field"`. Against a Fighter group it uses the Fighter interception event.

### 15.3 Internal damage

Add `internal_damage_applied` with:

- target Ship ID;
- source kind and optional source Ship/weapon/fighter/boarding identity;
- pool kind: `engine`, `secondary`, `weapon`, `special`, `shield_generator`;
- optional slot;
- current before/after;
- max;
- damage consumed;
- resulting weapon live Count / Special functional / shield-generator consequence where applicable.

Achilles weighted selection changes the chosen pool probability but does not create a separate RNG event.

### 15.4 Boarding events

Freeze:

#### `boarding_started`

- source/target Ship IDs;
- mode `raid`/`capture`;
- delivery mode `physical`/`transporter`;
- selected Marine count;
- effective party after Damper where applicable;
- source Marines before/after temporary dispatch.

#### `boarding_resolved`

- source/target Ship IDs;
- mode/delivery mode;
- result: `raid_complete`, `captured`, `repelled`, `quantum_detonated`;
- attacker/defender temporary or final survivors as appropriate;
- target Marine before/after persistent state;
- source Marine after final reconciliation;
- static Boarding modifiers used;
- RNG/casualty summary sufficient for deterministic debug.

#### `ship_captured`

- Ship ID;
- previous/new Empire and Seat/controller;
- MarineCurrent after capture;
- officer removal flag/identity where projected;
- post-capture-state flag;
- Telepathic bypass flag.

Quantum destruction also emits the existing `ship_destroyed` event after the Boarding result identifies the detonator outcome.

### 15.5 Special/lifecycle events

Use existing events where they already describe the authoritative mutation:

- teleport extends `ship_moved` with `movement_kind = "teleport"`;
- reflected/Absorber/Structural/Damper damage continues through `battle_damage_applied` plus identity fields.

Add narrowly scoped events:

- `ship_repaired` - durability/internal before/after and effective repair percent;
- `shields_recharged` - facing before/after and budget consumed;
- `special_state_changed` - ship/special/effect plus before/after for cloak/phasing, Absorber store/discharge, readiness/lifecycle states that need presentation;
- `time_warp_pass_started` - side/round and ordered eligible Ship IDs.

Passive static derivations such as Augmented Engines or ECM do not emit noise events merely for existing.

## 16. Tactical view / projection contract

### 16.1 Ship view

`TacticalShipView` gains/changes:

- mutable `owner_empire_id` / `controller_seat_id` sourced from `TacticalShipState`, not immutable original ownership;
- `current_combat_speed`;
- effective Beam/Missile defense values after runtime engine/Special state;
- detached `internal` current/max projection;
- detached ordered `specials []TacticalSpecialView` with slot, ID, support, functional state;
- cloak/phasing/post-capture state required for truthful Tactical rendering;
- existing Marine state remains projected.

`TacticalShipState` therefore also gains mutable `current_combat_speed`. Movement max/reset and speed-dependent Beam Defense derive from this server-owned value.

### 16.2 Fighter projection

`TacticalView` gains detached `Fighters TacticalFighterView` containing:

- Bay states;
- active Fighter groups;
- no mutable aliases;
- no RNG state.

Groups expose stable ID, owner/source, kind, target, position/facing, multiplicity, durability/remainder, attack stock and lifecycle phase sufficient for rendering/target selection.

### 16.3 Action projection

`TacticalView` gains `boarding_actions` and extends existing `weapon_actions` / `legal_moves` as frozen in Section 11.

Only the active controlling seat receives executable action catalogs. Inactive participants and observers receive authoritative state/events but no executable action authority.

PlayerView continues to zero Tactical RNG state.

## 17. Special functionality and damage-state projection

A Special is functional when:

- it is installed in immutable `TacticalSpecialSpec`;
- the Ship is otherwise eligible/functioning;
- and, for an original damageable Special slot, its internal Special pool remains above zero.

Original non-damageable Special IDs frozen in Section 5 do not become disabled merely because no ordinary internal pool exists.

Every Special consumer calls one shared server helper equivalent to `TacticalSpecialFunctional(ship, specialID)`. Parallel ad-hoc ID/health checks are prohibited.

Similarly, Beam/Fighter/Boarding/AI/UI action projection all consume the same helper so a damaged Special cannot remain active in one subsystem while disabled in another.

## 18. Round / turn scheduling

Normal round ordering extends the existing BattleSession scheduler deterministically:

1. normal start-of-round resets/bookkeeping already owned by 17.5-17.7;
2. decrement/reconcile ordnance cooldown/readiness;
3. apply 17.8 weapon special-readiness transitions required at round/turn boundaries;
4. process all active moving actors in merged ascending `spawn_order`;
5. remove/resolve dead actors/Ships and retarget as required;
6. establish normal Ship initiative/activation flow;
7. after both normal combat-side passes, execute Time-Warp extra passes in side order;
8. after normal + Time-Warp passes, run once-per-round shield recharge and Automated Repair;
9. advance round and emit authoritative round/pass events.

Time-Warp extra pass:

- contains only functional Time Warp Facilitator ships;
- calls ordinary ship turn initialization, including fresh teleporter use-state and weapon/action reset;
- skips the ordinary Seeking-Missile/moving-projectile end-pass call frozen by Gate 1;
- does not duplicate global shield recharge or Automated Repair.

## 19. Live snapshot v2 and migration

### 19.1 Version

`LiveSnapshotSchemaVersion` becomes **2** for newly emitted Battle snapshots.

`RestoreLiveSnapshot` accepts schema 1 and 2.

Version 2 is required because 17.8 introduces mutable owner/control, internal pools, Special lifecycle state and Fighter actor state that cannot be truthfully reconstructed from an arbitrary old Battle snapshot without additional source data.

### 19.2 Version-2 content

`TacticalLiveSnapshot` v2 preserves existing State/Ordnance/Events/NextEventSequence and adds:

- `fighters` runtime container with next group ID, Bay states and active groups;
- `next_moving_actor_order`;
- projectile `spawn_order` values;
- all new per-ship state from Sections 5, 9, 10 and 16;
- Boarding transient state;
- mutable owner/controller.

Every field that affects future RNG, legality, scheduling, ownership or recovery is persisted.

### 19.3 Schema-1 migration

Pre-17.8 production could not execute Fighter rows, Boarding or these Combat-Special consumers.

Therefore v1 restore uses a compatibility flag equivalent to `legacy_17_8_disabled = true` for that active Battle:

- existing Beam/ordnance behavior and ordnance actors restore exactly as before;
- existing projectile IDs receive deterministic `spawn_order = projectile_id` and `next_moving_actor_order` is initialized above the maximum;
- mutable owner/controller initializes from the old immutable Ship spec;
- no Fighter groups are synthesized;
- no hidden Internal/Special lifecycle values are guessed;
- Fighter/Boarding/new Special action projection remains unavailable for that migrated active Battle.

Once that legacy battle completes, future newly created battles use full v2 state.

This preserves old saves without pretending that metadata-only v1 battles had unrecorded 17.8 state.

### 19.4 Version-2 restore validation

Restore rejects at least:

- unknown/duplicate Fighter group IDs;
- invalid/duplicate spawn order;
- Fighter source Ship/slot mismatch or non-Fighter source row;
- `ready_groups` outside `0..configured Count`;
- multiplicity <=0 or above launch/recovery-consistent bounds;
- invalid target IDs/lifecycle combinations;
- returning group with wrong/missing source carrier identity;
- Internal current values outside `0..max`;
- Weapon internal pool inconsistent with projected live Count/unit size;
- Special lifecycle state for an uninstalled Special;
- Phasing counter outside 0..10;
- negative Absorber store;
- owner/controller not matching a participating battle side after allowed capture transitions;
- Marines outside `0..MarineMax`;
- invalid `special_readiness` for a non-eligible row;
- command/event sequence disagreement.

### 19.5 Determinism

Same v2 snapshot + same accepted next command stream must reproduce canonically:

- moving-actor order and IDs;
- Fighter attacks/losses/return/recovery;
- Boarding RNG/outcomes/ownership;
- Special RNG/order/effects;
- internal damage and repairs;
- events;
- final Battle result.

Observer/player projections remain detached and never expose mutable aliases.
## 20. Built-in AI and browser contract

### 20.1 Built-in AI uses the same legal-action surface

The built-in Tactical AI receives no private Fighter, Boarding or Special resolver.

`battleDecisionCatalogLocked` must consume the same detached BattleSession projections used by a human seat:

- executable Beam/Ordnance/Fighter `WeaponActions`;
- `BoardingActions`;
- normal/teleport `LegalMoves`;
- End/Wait/Retreat availability.

Deterministic catalog ordering freezes as:

1. executable weapon actions in existing weapon-slot order, with targets in the server-projected stable target order;
2. Boarding actions sorted by target Ship ID, then mode `raid` before `capture`;
3. movement fallback through the existing best-approach selector, which may consume `movement_kind=teleport` options;
4. End/Wait/Retreat according to their existing decision surfaces.

The baseline AI may remain strategically simple. Gate 3 only requires that each supported 17.8 action family is represented in the common catalog and can be selected without bypassing BattleSession legality.

Original AI evidence already constrains two behaviors:

- Boarding is an AI-owned Tactical action family, not player-only UI;
- Warp Dissipator must suppress AI retreat through the same server legality used for player retreat.

Time Warp, reactive defenses, repair/recharge and other passive Specials are automatic BattleSession behavior and never become AI commands.

### 20.2 Browser/API projection

The browser adds transport/types only for server-projected facts:

- Fighter Bay/group state and target identity;
- Fighter launch targets through `weapon_actions`;
- Boarding actions and Marine-count selection bounded by `max_marine_count`;
- `movement_kind` on legal moves and `ship_moved` events;
- internal-system state suitable for Tactical scan/status display;
- installed Special identity, functional state and visible lifecycle state;
- mutable owner/controller after Capture;
- the event vocabulary from Section 15.

The browser must not:

- generate Fighter/Boarding targets absent from the server list;
- compute Boarding range/shield/Damper legality;
- calculate Special functionality from design metadata alone;
- predict internal selection or damage;
- synthesize Capture owner changes;
- extrapolate moving-actor state after the last authoritative event.

Animation completion never gates BattleSession mutation or command acceptance.

## 21. Rules/data migration contract

### 21.1 Fighter rows

In `tactical_combat.json`, the three accepted Fighter rows migrate from `metadata_only` to `fighter` while retaining `craft_per_unit = 4` and their frozen design identity.

No Fighter modifier/ammo/firing-arc choices are added by Slice 17.8.

### 21.2 Combat-Special support projection

`ship_specials.json` migrates each row in Section 2.2 to its frozen `tactical_passive` or `tactical_active` support identity.

`strategic_only`, `turning_mode` and explicitly deferred rows remain truthful and are not upgraded merely because Gate 3 contains generic Special infrastructure.

### 21.3 Warp Dissipator row 37 repair

The frozen original 17.4 table contains selectable Warp Dissipator original row 37, but the current generated `ship_specials.json` omits that named row.

Gate 3 checkpoint 1 must repair the data projection from the already-frozen original row evidence before the Tactical consumer is enabled:

- canonical ID `warp_dissipator`;
- original ID 37;
- design space/cost/technology identity sourced from the frozen original dataset, not guessed;
- `tactical_support = "tactical_passive"`.

Ruleset validation and Designer tests must prove that this repair neither changes unrelated row order nor duplicates another Special.

### 21.4 Deferred Anti-Missile Rockets

Anti-Missile Rockets remain research/technology metadata and do not gain a fabricated `ship_specials.json` row or hidden charge count in Gate 3.

The frozen original projectile-interceptor hook remains reserved. A later evidence-backed source-state mapping may activate it without changing ordinary 17.8 moving-actor architecture.

### 21.5 Support validation

Ruleset/Tactical snapshot validation accepts the new support modes only for rows with a matching consumer:

- Fighter weapon support requires `kind=fighter`, positive `craft_per_unit`, zero ordinary min/max damage, and accepted Fighter identity;
- `tactical_active` / `tactical_passive` require a canonical known Special ID;
- unsupported support/identity combinations reject during rules/snapshot validation, not later inside combat resolution.

## 22. Frozen deterministic Gate-3 reference vectors

Gate 3 tests must encode the Gate-1 vectors plus these contract-level state transitions.

### Vector A - Fighter launch / loss / recovery

Baseline doctrine: Nuclear Drive, Titanium, non-Trans-Dimensional.

Carrier row: `fighter_bays`, configured Count 2.

- initial `ready_groups = 2`;
- accepted launch creates one Fighter group, multiplicity 8, speed 10, unit durability 2, attack stock 4;
- row `ready_groups -> 0` and ordinary row readiness is consumed;
- after attack cycles, assume multiplicity 7 and stock 0;
- functional carrier causes phase `outbound -> returning`;
- recovery computes `7 / 4 = 1` complete group and remainder 3;
- source Bay row recovers `ready_groups = 1`; three incomplete craft are not restored;
- group is removed after recovery.

### Vector B - Heavy Fighter paired payload

One Heavy-Fighter Bay under the same baseline doctrine:

- multiplicity 4;
- speed 8;
- unit durability 5;
- attack stock 2;
- each cycle executes Fighter Beam then Fighter Bomb for each surviving craft;
- stock decrements once after the paired pass;
- exactly two paired cycles are available before return if the group survives.

### Vector C - Raid temporary Marines / persistent internal damage

Source has `MarineCurrent = 10`, selected party 4, no Transporter/Damper.

- command temporarily subtracts source Marines 10 -> 6;
- Raid loop may decrement temporary attacker/defender counters and may apply internal damage according to strict diff thresholds;
- on Raid completion the returned party is added back to source, restoring source Marines to 10;
- target persistent `MarineCurrent` equals its pre-Raid value;
- any resolved `internal_damage_applied` mutations remain.

This vector explicitly guards against incorrectly persisting Raid Marine casualties.

### Vector D - Capture / Quantum Detonator

Source selected party 8; Ground-Combat leaves attacker 3, defender 0.

- source selected party is not returned;
- target `MarineCurrent = 3` before `Capture_Ship` reconciliation;
- Quantum absent -> owner/controller transfers and non-Telepathic new owner gets post-capture state;
- Quantum present with `Random(10)=5` -> ship destroys before owner transfer;
- Quantum present with `Random(10)=6` -> transfer proceeds.

### Vector E - Energy Absorber + Damper

Incoming current damage 20 with both functional:

- Absorber stores 5, remaining damage 15;
- Damper quarters 15 -> 3;
- downstream damage begins from 3;
- reversing this order is invalid.

### Vector F - Reflection

Current Beam hit damage D=20:

- `Random(35)=20` -> no reflection;
- `Random(35)=21` -> target damage 0, reflected output 20 enters shared ship-damage resolver against attacker.

### Vector G - Lightning grouped actor

Moving actor multiplicity 4; first two Lightning rolls are `<50`:

- draw 1: multiplicity 4 -> 3;
- draw 2: multiplicity 3 -> 2;
- loop ends because iteration index 2 is not below mutated multiplicity 2;
- exactly two Lightning RNG draws occur before reactive PD.

### Vector H - Shield stack / recharge

Normal per-facing max 40 in nebula/global suppression:

- no Hard Shields -> max 0; Multi-Phased remains 0;
- Hard Shields -> survive at 40; Multi-Phased -> 60.

For base shield value 20 with four facings each four points low:

- normal recharge budget = 6 -> increments `[2,2,1,1]` in pointwise facing order;
- Shield Capacitors budget = 20 -> all 16 missing points restore, then stop.

### Vector I - teleport Move

Functional ID22/33 ship, unused teleporter, ordinary movement points 0, legal destination at Range 20:

- projected move is available with `movement_kind=teleport`;
- accepted `battle.move_ship` updates coordinate, preserves facing, consumes no ordinary movement points and sets `teleporter_used=true`;
- a second teleport in the same eligible ship turn is unavailable;
- destination at Range 21 is unavailable.

### Vector J - Time Warp scheduling

One functional ID34 ship on side A, none on side B:

- normal side-A pass;
- normal side-B pass;
- extra side-A pass includes only the ID34 ship with normal start-turn reset;
- extra side-B pass contains no eligible ship action;
- global shield recharge and Automated Repair run once after the pass sequence, not once per Time-Warp activation.

### Vector K - Internal Achilles selection weights

Current pool weights: Weapon A 10, Weapon B 4, Shield generator 20, other fixed category 12.

With attacker Achilles:

- Weapon A -> 30;
- Weapon B -> 12;
- Shield generator -> 60;
- other -> 12;
- shared weighted-choice RNG consumes the transformed array; no extra Achilles roll occurs.

### Vector L - v1 snapshot compatibility

Restore a valid pre-17.8 schema-1 battle with existing projectile IDs 2 and 5:

- Beam/ordnance state restores unchanged;
- projectile spawn orders become 2 and 5 in ascending projectile order;
- next moving-actor order becomes 6;
- no Fighter/internal/Special hidden state is synthesized;
- `legacy_17_8_disabled` suppresses Fighter/Boarding/new-Special actions for that active battle;
- saving the migrated battle emits schema 2 while preserving the compatibility-disabled state until battle completion.

## 23. Gate-3 implementation checkpoints

Gate 3 must be executed as bounded, separately reviewable checkpoints:

1. **Rules/data/schema foundation** - Fighter support identity, Special support migration, Warp-Dissipator row repair, rules snapshot blocks, Fighter doctrine snapshot, LiveSnapshot v2 migration scaffolding and strict validation.
2. **Internal systems / passive derivation** - immutable/current internal pools, mutable current speed, shared Special-functionality helper, Reinforced/Heavy/Hard/Multi-Phased/Augmented/Battle-Scanner/ECM hooks and internal damage event.
3. **Fighter actor lifecycle** - Bay state, group IDs/shared spawn order, launch command/action projection, movement/retarget, durability/interception, payload resolution, return/recovery and vectors A/B/G.
4. **Boarding / Capture** - Boarding action catalog and command, physical/Transporter legality, Raid temporary reconciliation, Capture/Ground-Combat/owner-control/Quantum/post-capture state, Warp-Dissipator retreat and vectors C/D.
5. **Special lifecycle / scheduler** - cloak/phasing, Energy Absorber/Damper/Reflection/Displacement/Lightning, Hyper-X/FMR, Shield Capacitors/Automated Repair, teleport Move and Time-Warp pass ordering with vectors E-J.
6. **Persistence / AI / UI / integration** - complete v2 save/load/replay/reconnect, observer/player detachment, built-in AI common action catalog, event-driven Tactical UI, Selector/Build -> BattleSession integration and vectors K/L plus the full Gate-1 reference suite.

Every checkpoint ends with focused tests, the relevant wider package suite and `git diff --check`. Production behavior outside the checkpoint contract must not be added opportunistically.

Gate 3 does not start until this Gate-2 freeze is committed and the working tree is clean.

## 24. Gate-2 result

**FROZEN / PASS - Gate 2 complete; Gate 3 implementation ready / not started.**

No Slice-17.8 production Tactical behavior is changed by this contract document.

Implementation authority is now fully constrained by:

- the Gate-1 executable audit;
- this exact state/command/event/persistence contract;
- inherited 17.5-17.7 authority and deterministic scheduling.

Unsupported behavior remains explicit:

- Phase-Shifter other-ship teleport is deferred;
- Anti-Missile Rockets are not activated until their original consumable live-count source can be represented without invention;
- planetary Fighter Garrison / planetary combat remains Slice 19.
## 25. Binding implementation clarifications

These clarifications remove remaining transport/state ambiguity and are part of the Gate-2 freeze.

### 25.1 Fighter payload snapshot is self-contained

The Fighter doctrine payload entries are immutable embedded payload specs, not live ruleset lookups.

`TacticalFighterPayloadSpec` carries at least:

- `weapon_id`;
- base `min_damage` / `max_damage`;
- canonical innate/effect IDs used by the Fighter resolver;
- any frozen immutable mode/flag required by the Gate-1 Fighter Beam/Bomb path.

New BattleSession replay must therefore not depend on a mutable external weapon table after the Tactical spec is created.

### 25.2 Mutable ownership is the only runtime control authority after Capture

`TacticalShipSpec.EmpireID/SeatID` remain immutable **initial/source ownership** for provenance only.

After Tactical start, all authority-sensitive runtime logic uses `TacticalShipState.owner_empire_id` / `controller_seat_id`, including:

- command ownership;
- ally/enemy target legality;
- Boarding legality;
- initiative/active-seat action projection;
- built-in AI decision ownership;
- remaining-winner determination;
- fleet-side Special effects such as Wide Area Jammer/Warp Dissipator;
- Fighter return ownership compatibility.

On Capture, the mutable owner/controller changes atomically at the frozen reconciliation point. A launched Fighter group retains its launch/source seat identity; if its original carrier changes to an incompatible owner before recovery, the normal ownership-compatible recovery check fails.

### 25.3 Legacy compatibility flag is persisted in v2

`TacticalLiveSnapshot` v2 includes explicit `legacy_17_8_disabled`.

When a schema-1 active battle is migrated, this flag becomes true and remains true for that battle through subsequent v2 saves/restores. While true:

- no 17.8 Fighter/Boarding/new-Special action is projected;
- no hidden 17.8 mutable state is synthesized;
- existing 17.5-17.7 behavior continues normally.

Newly created post-Gate-3 battles emit the flag false.

### 25.4 Fighter attack event attempt records

`fighter_attack_resolved` contains an ordered `attempts` list, stable by payload pass then craft index.

Each attempt carries the payload-appropriate authoritative facts needed to explain the RNG result, including:

- `craft_index`;
- `payload_kind` (`beam` or `bomb`);
- target identity;
- raw roll(s) consumed by that payload;
- effective hit/defense values and threshold where applicable;
- hit flag;
- resolved raw/current damage before shared target-layer application.

Shared `battle_damage_applied` / actor-interception events continue to describe the actual target-state mutation. The attempt list does not become gameplay input.

### 25.5 Boarding event round/iteration records

`boarding_resolved` carries ordered resolver evidence rather than only a textual summary:

- Raid: ordered iterations with attacker raw roll, defender raw roll, computed diff and triggered threshold branches;
- Capture: ordered Ground-Combat rounds with both raw rolls, tier/progress before/after and casualty markers;
- Quantum capture test: raw `Random(10)` result when applicable.

This event detail is replay/debug evidence only; authoritative state is still the BattleSession mutation.

### 25.6 Post-capture state lifetime

No unsupported clear transition is invented.

A normal non-Telepathic Capture sets `post_capture_state`; Gate 3 applies every directly frozen start-turn/eligibility consequence while the flag is set. In the absence of a frozen original clear transition, the implementation does not spontaneously clear it. Telepathic Capture does not set it.

Any future change to that lifetime requires new original evidence and a later contract revision.

### 25.7 Fighter OCV / moving-actor DCV snapshot correction (2026-10-06)

Checkpoint-3 implementation review exposed one Gate-2 snapshot omission. Direct re-disassembly of original `Fighter_Ocv_` and `Missile_Dcv_` proves that future Fighter Beam RNG depends on values beyond payload identity, drive and armor, so those values must be frozen in the Tactical doctrine rather than re-derived from mutable strategic state during replay.

Binding correction:

- `fighter_beam_offense` is the immutable Fighter OCV scalar used by Fighter Beam resolution;
- for normal players the evidenced original stack is `Fleet_Offensive_Bonus + Best_Computer.offense + player Ship Attack (+0x8A6) + Get_Fleet_Pilot_Bonus + 50`;
- `moving_actor_defense_bonus` is the immutable additive term paired with Fighter `speed * 5` for Fighter-class `Missile_Dcv_`; the evidenced normal-player additions are `Get_Fleet_Pilot_Bonus + player Ship Defense (+0x8A5)`;
- the existing Fighter payload snapshots remain immutable and self-contained as already frozen.

Current MOOX has authoritative best-known Computer and race Ship Attack/Ship Defense state, but no authoritative fleet-officer / pilot runtime. Therefore Battle creation freezes the representable terms now (`+50 + best Computer offense + race Ship Attack` and race Ship Defense respectively); unavailable fleet-offensive/pilot terms are explicitly `0`, not inferred. A later leader/officer slice may populate those terms at Battle creation without changing Fighter replay semantics, because the BattleSession continues consuming only the frozen scalars.

This is a contract correction required by already accepted Gate-1 executable evidence, not a new mechanic.

## 2026-10-07 user-review refinement - ship-target attack stations

Manual Slice-17.8 review exposed a usability and spatial-consistency defect in the frozen arrival wording: Fighter-family moving actors were persisted on the same tactical cell as their target Ship while executing payload cycles.

The accepted refinement is server-authoritative:

- a Ship-targeting Fighter/Bomber/Heavy-Fighter group must never persist on the target Ship cell;
- the server resolves a deterministic free adjacent attack station around the live target Ship and treats that station as the group's outbound arrival coordinate;
- free neighboring stations are distributed across groups when possible; if every neighboring station is already occupied by moving actors, station sharing is permitted, but the Ship cell itself remains forbidden;
- pre-impact defenses and payload execution occur when the group reaches that adjacent attack station;
- returning groups still recover at their source carrier through the existing return/recovery lifecycle.

This refines the spatial meaning of "reaches its current target" in Section 7.2 without changing Fighter/Bomber/Heavy-Fighter payload composition, attack stock, damage RNG, retargeting or recovery arithmetic.

### 2026-10-07 user-review refinement - shared diagonal trajectory

Fighter-family actors inherit the refined shared moving-actor partial trajectory: a quantum uses its full field budget and selects the straightest in-budget grid endpoint toward its current authoritative target. Fighter attack-station, payload, retarget and recovery semantics remain unchanged.
