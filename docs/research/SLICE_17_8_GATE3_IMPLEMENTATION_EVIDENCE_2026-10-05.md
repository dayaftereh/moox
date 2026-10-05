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
