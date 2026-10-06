# Slice 17.8 Gate 3 - Implementation Evidence

Date: 2026-10-05
Status: **Gate 3 in progress - Checkpoints 1-5 complete; Checkpoint 6 next**

Contract authority:

- `docs/research/SLICE_17_8_GATE2_TACTICAL_FIGHTERS_SPECIALS_CONTRACT_FREEZE_2026-10-05.md`

## Checkpoint 1 - rules / data / schema foundation

### Fighter support identity

The three frozen Fighter-family Weapon rows are now executable-support identities rather than deferred metadata:

- `fighter_bays` -> `tactical_support="fighter"`, `craft_per_unit=4`;
- `bomber_bays` -> `tactical_support="fighter"`, `craft_per_unit=4`;
- `heavy_fighter_bays` -> `tactical_support="fighter"`, `craft_per_unit=4`.

Battle snapshot validation accepts `fighter` only for the Fighter family with the frozen four-craft unit and zero ordinary Weapon min/max damage. This checkpoint does not yet add the launch command or Fighter actor lifecycle.

### Combat-Special support migration

The frozen Slice-17.8 Special matrix is represented directly in `ship_specials.json`:

- passive/reactive consumers -> `tactical_passive`;
- Phase Shifter and Sub-Space Teleporter -> `tactical_active`;
- Inertial Stabilizer / Inertial Nullifier remain `turning_mode`;
- strategic/deferred metadata rows remain their existing truthful support class.

Ruleset validation now verifies the canonical Special-ID -> Tactical-support classification instead of merely accepting any non-empty string.

Original row 37 is repaired from the pre-Gate-3 typo `warp_dissipater` to canonical `warp_dissipator`, retaining original ID 37 and the frozen Tactical passive support.

### Frozen 17.8 rules snapshot blocks

`tactical_combat.json` and the ruleset DTOs now carry the Gate-2 scalar blocks for:

- Fighter movement/count/speed/hits/attack stock/drive/armor/recovery;
- Boarding ranges, Damper divisor, Security bonus, Raid thresholds, Quantum bound and crew-quality bonuses;
- active/passive Combat-Special scalar constants.

The same blocks are copied into immutable `TacticalRulesSnapshot` records and strictly validated by BattleSession admission.

### Fighter doctrine snapshot

`TacticalSpec` now carries ordered `FighterDoctrines`.

For each participating seat that actually owns an executable Fighter Bay, the production Encounter path snapshots:

- seat / empire identity;
- best known Empire Warp-Drive component index;
- best known Empire Armor durability multiple;
- Trans-Dimensional race state;
- optional immutable Fighter Beam/Bomb payload snapshot slots for the later Fighter-runtime checkpoint.

Doctrine validation requires exactly one ordered doctrine for every Fighter-capable seat and rejects mismatched authority or unsupported drive/armor values.

The production projection deliberately derives doctrine from Empire-wide best-known equipment rather than the drive/armor installed on the carrier.

### Immutable internal-system DTO boundary

`TacticalShipSpec` now has the frozen `TacticalInternalSpec` shape:

- engine max;
- secondary max;
- ordered Weapon-slot max/unit-size entries;
- ordered Special-slot max entries;
- Shield-generator max.

Checkpoint 1 adds schema/clone/validation support only. The actual internal-pool derivation/current-state consumer remains Checkpoint 2 per the Gate-2 boundary.

### LiveSnapshot v2 migration scaffolding

New Battle snapshots emit schema version **2**.

Schema-v2 Tactical snapshots add:

- `next_moving_actor_order`;
- persisted `legacy_17_8_disabled`;
- projectile `spawn_order`.

New projectile launches currently allocate from the shared monotone moving-actor counter while preserving all existing ordnance processing semantics.

Schema version 1 remains accepted. Migration:

1. deep-clones caller-owned Tactical state before mutation;
2. sets `legacy_17_8_disabled=true`;
3. assigns each surviving legacy projectile `spawn_order=projectile_id`;
4. initializes `next_moving_actor_order` above both active legacy spawn orders and the historical `NextProjectileID` counter;
5. re-exports the restored battle as schema v2.

The Gate-2 compatibility vector with legacy projectile IDs 2 and 5 is regression-covered and yields spawn orders 2/5 plus next order 6.

No Fighter/internal/Special hidden state is synthesized for a schema-v1 active battle.

### Ruleset identity

The reviewed `moo2-1.31` ruleset fingerprint changes only after the final Checkpoint-1 data shape is frozen.

New reviewed SHA-256:

`800c8813437503390d2420f8ee50eeb6e528b9bf960fce725f88448b966fda3a`

## Checkpoint 1 regression evidence

Passed on the final Checkpoint-1 working tree:

- `go test ./internal/ruleset`
- `go test ./internal/battle`
- `go test ./internal/game`
- focused Fighter-doctrine production projection test
- focused schema-v1 -> v2 compatibility/moving-actor migration tests
- `go test ./...`
- `git diff --check`

The first direct `go test ./...` transport call failed inside the remote-agent transport rather than in Go; a single bounded rerun with output redirected to `C:\ASH\Temp\moox-s178-g3-cp1-go-test-all.log` completed with exit code 0.

## Scope guard

Checkpoint 1 deliberately does **not** implement:

- Fighter launch/movement/targeting/attack/return/recovery;
- mutable internal damage pools or Automated Repair;
- Boarding/Raid/Capture execution;
- Combat-Special runtime effects;
- Fighter UI or AI actions;
- Phase-Shifter other-ship teleport;
- Anti-Missile-Rockets consumable behavior.

Those remain owned by Gate-3 Checkpoints 2-6 exactly as frozen by Gate 2.

## Result

**Checkpoint 1 COMPLETE / PASS.**

Next authorized work:

**Gate 3 Checkpoint 2 - Internal systems / passive derivation.**

## Checkpoint 2 - internal systems / passive derivation

Status: **COMPLETE / PASS**

### Immutable and mutable internal-system authority

Production Tactical projection now derives immutable Internal maxima from the recovered Original executable contract:

- fixed Engine / Secondary / Shield-generator maxima per hull;
- Reinforced Hull triples the Engine/drive Internal maximum exactly where the Original `Init_Internal_Space_` path does;
- Weapon Internal maxima derive from physical mount count multiplied by the frozen per-weapon unit size;
- damageable installed Specials receive ordered Internal pools while Original non-damageable Special IDs remain absent from the mutable pool list.

`TacticalShipState` now persists:

- mutable current combat speed;
- live physical Weapon counts;
- Engine / Secondary / Weapon / Special / Shield-generator current Internal state.

Views expose the same authoritative current/max Internal state without recomputing combat results client-side.

### Shared Special functionality gate

One server-side `TacticalSpecialFunctional(...)` helper now owns the common installed/functional decision for damageable Specials.

Existing and new passive consumers use the same live state instead of reading immutable installation alone.

### Passive combat derivations

Checkpoint 2 wires the frozen passive effects into authoritative Battle consumers:

- Reinforced Hull: x3 normal Structure maximum and recovered Engine Internal maximum interaction;
- Heavy Armor: x3 Armor maximum and Armor-Piercing / armor-first-bypass neutralization while functional;
- Hard Shields: Shield-Piercing cancellation plus the frozen extra 3 shield absorption in both Beam and Ordnance damage paths;
- Multi-Phased Shields: x3/2 shield-facing maximum while functional, with live facings clamped when the Special is lost;
- Augmented Engines: +5 current combat speed before Engine-damage scaling, with movement/initiative derived from live speed;
- Battle Scanner: +50 live Beam offense while functional;
- ECM Jammer / Multi-Wave ECM Jammer / Wide Area Jammer: best-only self contributions 70 / 100 / 130, plus Wide Area +70 fleet contribution.

The ECM constants are frozen in `tactical_combat.json` and copied into the immutable Battle rules snapshot; runtime logic no longer owns private numeric literals for those bonuses.

### Internal-damage event foundation

`internal_damage_applied` now records the stable affected ship/pool/slot/source plus before/after/max values and consequential externally useful state, including:

- live Weapon count after Weapon-pool damage;
- Special identity and functional state after Special-pool damage;
- shield-generator destruction and clamped facing state;
- combat-speed before/after Engine or Augmented-Engine effects.

Destroyed live Weapon counts are also consumed by Beam fire, reactive Beam point defense and Ordnance launch multiplicity.

### LiveSnapshot compatibility

Tactical LiveSnapshot state now carries `slice_17_8_state_version=1`.

Restore migration for pre-Checkpoint-2 schema-v2 snapshots:

1. deep-clones caller-owned Tactical state;
2. initializes new current-speed / live-Weapon-count fields from immutable snapshot identity;
3. initializes Internal current pools from immutable maxima;
4. preserves existing movement consumption while clamping to the migrated current-speed maximum;
5. validates live Weapon counts against Internal Weapon pools.

Legacy schema-v1 battles remain compatibility-disabled for 17.8 runtime mechanics and do not synthesize active Internal mechanics.

### Ruleset identity

Final reviewed Checkpoint-2 `moo2-1.31` SHA-256:

`68c06a57bb9fdc18c0ff8045b54241e51cc9cba41298898fbfb535b6a338bd46`

### Regression evidence

Passed on the final Checkpoint-2 working tree:

- dedicated `TestSlice178Gate3Checkpoint2*` Battle regressions;
- dedicated `TestSlice178Gate3Checkpoint2*` Game projection regressions;
- `go test ./internal/ruleset`;
- `go test ./internal/battle`;
- `go test ./internal/game`;
- `go test ./...`;
- `git diff --check`.

Full-suite log:

`C:\ASH\Temp\moox-s178-g3-cp2-go-test-all.log`

## Checkpoint 2 result

**Checkpoint 2 COMPLETE / PASS.**

Next authorized work:

**Gate 3 Checkpoint 3 - Fighter actor lifecycle / Fighter combat resolution.**


## Checkpoint 3 - Fighter actor lifecycle / Fighter combat resolution

Status: **COMPLETE / PASS**

### Authoritative Fighter runtime

Checkpoint 3 adds BattleSession-owned Fighter/Bomber/Heavy-Fighter actors rather than browser-side craft simulation:

- per-Bay `ready_groups` state;
- battle-local monotone Fighter `group_id`;
- shared non-zero `spawn_order` allocated from the same `next_moving_actor_order` counter as ordnance;
- immutable source seat/carrier/slot identity and Fighter kind;
- server-owned target identity, position/facing, speed/movement, multiplicity/durability/remainder, attack stock and outbound/returning phase;
- immutable embedded Beam/Bomb payload snapshots.

Projectiles and Fighter groups are processed together in ascending shared `spawn_order`.

### Launch command and action projection

Fighter-family rows now project through the existing Tactical WeaponAction surface with:

- `weapon_family = "fighter"`;
- `command_kind = "battle.launch_fighters"`;
- server-projected legal ship / projectile / Fighter-group targets according to Fighter kind;
- strict exactly-one-target validation;
- explicit `fighter_bays_empty`, ordinary readiness and no-target rejection paths.

Rejected launch does not allocate a group, consume readiness or advance Fighter/moving-actor counters.

Accepted launch consumes the complete ready Bay groups, freezes doctrine/payload state, emits `fighter_group_launched` and executes the frozen immediate movement quantum.

### Movement, retarget, attack and recovery

The Fighter lifecycle now covers:

- deterministic movement quanta using the existing 20-units-per-field Tactical geometry;
- target validation and deterministic retargeting across supported ship/projectile/Fighter targets;
- `fighter_group_retargeted`, returning/destroyed/recovered lifecycle events;
- carrier-availability and ownership-compatible recovery checks;
- integer recovery by the frozen divisor with incomplete craft discarded;
- Fighter Beam and Fighter Bomb passes per surviving craft;
- Heavy-Fighter paired Beam-then-Bomb cadence with one stock decrement per paired cycle;
- shared ship-damage and moving-actor interception application.

### Fighter Beam OCV/DCV contract correction

Implementation review found a real Gate-2 snapshot omission, now recorded as binding clarification 25.7 in the Gate-2 freeze.

Re-disassembly proves:

- `Fighter_Ocv_`: `Fleet_Offensive_Bonus + Best_Computer.offense + player Ship Attack (+0x8A6) + Get_Fleet_Pilot_Bonus + 50`;
- Fighter-class `Missile_Dcv_`: `speed * 5 + Get_Fleet_Pilot_Bonus + player Ship Defense (+0x8A5)` for normal players.

The immutable Fighter doctrine therefore freezes:

- `fighter_beam_offense`;
- `moving_actor_defense_bonus`.

Current MOOX can authoritatively project best-known Computer plus race Ship Attack/Ship Defense. It does not yet have authoritative fleet-officer / pilot runtime, so those unavailable original terms are explicitly frozen as zero rather than inferred. This limitation is visible and replay-stable; a later leader/officer slice can populate the same frozen scalars without changing BattleSession replay semantics.

The earlier Gate-1 wording that called `+0x8A6` a crew-dependent term was corrected to its verified Ship Attack meaning.

### Lightning Field and reactive Point Defense

Fighter arrival uses the frozen ordering:

1. validate/retarget;
2. Lightning Field grouped-actor loop;
3. reactive Point Defense;
4. surviving payload resolution.

The Lightning loop uses the live mutated multiplicity as its loop bound. Vector G therefore performs exactly two sub-50 destruction draws for the frozen 4 -> 3 -> 2 case before reactive PD.

### LiveSnapshot / restore validation

Tactical live state now persists the Fighter runtime and advances `slice_17_8_state_version` to 2 for the Checkpoint-3 state shape.

Restore/migration and validation cover:

- Fighter next-group ID, Bay state and active group state;
- no synthesized active Fighter state for legacy compatibility-disabled battles;
- shared projectile/Fighter `spawn_order` uniqueness and bounds;
- source carrier/slot/Fighter-row identity;
- ready-group, multiplicity, durability/remainder, stock, movement and phase bounds;
- immutable payload equality with the source doctrine;
- target existence and hostility for ship/projectile/Fighter targets;
- valid returning-carrier linkage.

### Frozen vectors and focused regressions

Dedicated Checkpoint-3 coverage includes:

- Vector A - Fighter launch / loss / return / recovery;
- Vector B - Heavy-Fighter paired Beam/Bomb cadence;
- Vector G - Lightning mutable-multiplicity loop before PD;
- Fighter moving-actor DCV = `speed * 5 + frozen moving_actor_defense_bonus`;
- production Fighter-doctrine projection from best-known equipment/race state.

### Regression evidence

Passed on the final reviewed Checkpoint-3 working tree before commit:

- dedicated `TestSlice178Gate3Checkpoint3*` Battle regressions;
- focused Fighter-doctrine Game projection regression;
- `go test ./internal/ruleset`;
- `go test ./internal/battle`;
- `go test ./internal/game`;
- `go test ./...`;
- `git diff --check`.

No rules-data file changes are required by Checkpoint 3, so the reviewed `moo2-1.31` ruleset fingerprint remains the Checkpoint-2 value:

`68c06a57bb9fdc18c0ff8045b54241e51cc9cba41298898fbfb535b6a338bd46`

## Checkpoint 3 result

**Checkpoint 3 COMPLETE / PASS.**

Next authorized work:

**Gate 3 Checkpoint 4 - Boarding / Capture.**


## Checkpoint 4 - Boarding / Capture

Status: **COMPLETE / PASS**

### Boarding / Raid / Capture

Checkpoint 4 implements the frozen `battle.board_ship` contract as BattleSession-owned Tactical state:

- server-projected physical/Transporter Boarding legality and bounded Marine selection;
- persisted Boarding/Transporter use flags;
- Raid temporary Marine accounting with persistent internal damage (Vector C);
- Capture Ground Combat, survivor transfer, Quantum-before-owner-transfer, mutable owner/controller and Telepathic post-capture handling (Vector D);
- immediate reconciliation of inbound projectiles/Fighter groups after ownership transfer.

Current MOOX has no authoritative fleet-officer or per-ship crew-quality runtime, so those original terms remain explicit frozen zero inputs. Ground-Combat and Telepathic race/research state are projected authoritatively.

### Mutable authority / moving actors

Post-Capture command, targeting, action-projection and winner authority now use mutable `OwnerEmpireID` / `ControllerSeatID`.

Already-launched Fighter/projectile actors retain their frozen launch `SourceSeatID`; new launches from a captured ship use its current controller. Fighter recovery requires the carrier's current controller to remain compatible with the frozen launch seat.

Capture immediately retargets inbound actors that become friendly after ownership transfer. Projectiles without a remaining hostile target use the existing `no_target` destruction path; Fighters use the existing retarget / return / carrier-unavailable lifecycle.

### Warp Dissipator

A functional enemy `warp_dissipator` now suppresses server `battle.retreat` and projected `CanRetreat`.

Destroyed, post-capture-invalid, damaged/non-functional, or newly friendly carriers do not block retreat. Built-in AI common action-catalog work remains reserved for Checkpoint 6; no separate current retreat mutation bypasses the server command path.

### LiveSnapshot v3

Restore validation now checks:

- valid mutable ship owner/controller pairs;
- frozen actor source-side identity from `SourceSeatID`;
- current ship-target hostility against mutable ownership;
- actor-vs-actor hostility using frozen launch-side identity.

### Regression evidence

Passed on the final reviewed Checkpoint-4 working tree:

- dedicated `TestSlice178Gate3Checkpoint4*` regressions, including Vectors C/D, mutable authority, actor reconciliation and Warp Dissipator;
- `go test ./internal/battle`;
- `go test ./internal/game`;
- `go test ./internal/ruleset`;
- `go test ./...`;
- `git diff --check`.

No rules-data files changed in Checkpoint 4. The reviewed `moo2-1.31` ruleset fingerprint remains:

`68c06a57bb9fdc18c0ff8045b54241e51cc9cba41298898fbfb535b6a338bd46`

## Checkpoint 4 result

**Checkpoint 4 COMPLETE / PASS.**

Next authorized work:

**Gate 3 Checkpoint 5 - Special lifecycle / scheduler.**

## Checkpoint 5 - Special lifecycle / scheduler

Status: **COMPLETE / PASS**

### Authoritative Special lifecycle state

Checkpoint 5 advances the Slice-17.8 live state to version 4 and adds replay-stable server authority for:

- Cloaking / Phasing lifecycle state plus remaining Phasing lifetime;
- Energy-Absorber stored damage;
- per-weapon Hyper-X / Fast-Missile-Racks `special_readiness`;
- self-teleporter per-turn use;
- Tactical pass kind, Time-Warp side and deterministic Time-Warp queue.

Restore/migration rejects invalid or unowned Special state, including Phasing counters outside the frozen 0..10 domain, Absorber store on a ship without the device, and Special readiness on non-eligible weapon rows. Pre-Checkpoint-5 snapshots migrate deterministically without inventing hidden legacy state.

### Cloaking / Phasing

The frozen original lifecycle is implemented as authoritative ship state:

- Cloaking Device starts cloaked; fire/Boarding breaks cloak; next turn enters recovering; a non-firing pass recloaks at own-pass cleanup;
- Phasing Cloak starts phased with lifetime 10 and is untargetable while actively phased;
- fire/Boarding breaks Phasing; next turn enters phase-recovery; own-pass cleanup restores Phasing and immediately retargets inbound projectile/Fighter actors away from the newly untargetable ship;
- each active phased cleanup decrements the lifetime and zero converts to ordinary Cloaking.

Beam defense and ordnance forced-miss consumers use the frozen Cloak state without browser authority.

### Damage / defensive Special ordering

Checkpoint 5 wires the frozen Special hook ordering into the shared damage and attack paths:

- Energy Absorber stores one quarter before Damper Field reduction (Vector E);
- stored Absorber energy auto-discharges before a normal weapon attack against a ship at Tactical range < 8 and expires at the owning side pass boundary;
- the direct `Energy_Absorber_Attack_` Displacement branch preserves the audited `Random_(100) > 30` miss behavior;
- Reflection Field uses the recovered `Random_(D+15) > D` boundary (Vector F);
- generic Beam Displacement preserves `<= 30`, ordnance Displacement preserves `< 30`, and Cloak missile forced-miss preserves `< 50` boundaries;
- Fighter / projectile impacts share the same Special-aware damage authority.

Vector G remains the already-green Checkpoint-3 Lightning grouped-actor regression and is intentionally not duplicated.

### Weapon readiness

Hyper-X Capacitors and Fast Missile Racks use the recovered shared weapon-row state machine:

- combat initialization: `extra_volley_ready`;
- first exhausted volley consumes the extra-volley state and immediately reloads one extra volley;
- later turn start converts non-ready state to `rearm_pending`;
- firing while pending consumes it and prevents re-arm;
- a full non-firing turn leaves `rearm_pending` intact so own-pass cleanup restores `extra_volley_ready`.

The same completion helper is used by ship Beam fire and reactive Beam fire against projectile/Fighter actors.

### Shield recharge / repair / teleport

- Shield recharge keeps the original pointwise distribution; normal factor 3 and Shield-Capacitor factor 10 are regression-covered by Vector H.
- Automated Repair applies the frozen base 20 percent plus any authoritative frozen Fleet-Engineer term, repairs Structure then Armor, then internal Engine -> Secondary -> Weapons -> Shield -> Specials, and skips destroyed/post-capture ships.
- Phase Shifter / Sub-Space Teleporter self-movement reuses `battle.move_ship`, projects `movement_kind=teleport`, preserves facing/movement, consumes the once-per-turn teleporter use flag and accepts the frozen max range 20 / rejects 21 (Vector I).
- the ordinary movement rejection order is preserved so legacy command/error semantics remain stable.

### Time-Warp scheduler

Time-Warp is implemented as an additive BattleSession scheduler pass rather than a second combat authority:

1. ordinary initiative/Wait activation flow remains unchanged;
2. each side finishes its ordinary pass and own-pass Special cleanup;
3. eligible Time-Warp ships receive deterministic side-ordered extra-pass queues;
4. Time-Warp pass completion runs the same own-pass cleanup for the participating ships;
5. after all Time-Warp side passes, shield recharge and Automated Repair run exactly once and the next ordinary round begins.

Pass kind, side and queue are persisted and restore-validated. Vector J covers deterministic pass ordering and one-shot end-of-round maintenance.

### Frozen vectors / regression evidence

Dedicated Checkpoint-5 coverage includes:

- Vector E - Energy Absorber before Damper;
- Vector F - Reflection boundary;
- Vector G - Lightning grouped multiplicity via the existing Checkpoint-3 regression;
- Vector H - Shield-Capacitor pointwise recharge;
- Vector I - self-teleport range/use semantics;
- Vector J - Time-Warp scheduling / maintenance ordering;
- Cloak / Phasing lifecycle and forced-miss boundaries;
- Hyper-X / Fast-Missile-Racks extra-volley and full-turn re-arm;
- Automated Repair pool order;
- Energy-Absorber pre-weapon discharge / expiry;
- LiveSnapshot v4 migration and restore hardening.

Passed on the final reviewed Checkpoint-5 working tree:

- focused `TestSlice178Gate3Checkpoint5*` regressions;
- `go test ./internal/ruleset`;
- `go test ./internal/battle`;
- `go test ./internal/game`;
- `go test ./...`;
- `git diff --check`.

No rules-data file changes are required by Checkpoint 5, so the reviewed `moo2-1.31` ruleset fingerprint remains:

`68c06a57bb9fdc18c0ff8045b54241e51cc9cba41298898fbfb535b6a338bd46`

## Checkpoint 5 result

**Checkpoint 5 COMPLETE / PASS.**

Next authorized work:

**Gate 3 Checkpoint 6 - persistence / AI / UI / integration.**
