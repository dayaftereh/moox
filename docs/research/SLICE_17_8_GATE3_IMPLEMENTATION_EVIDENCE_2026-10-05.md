# Slice 17.8 Gate 3 - Implementation Evidence

Date: 2026-10-05
Status: **Gate 3 in progress - Checkpoint 1 complete; Checkpoint 2 ready**

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
