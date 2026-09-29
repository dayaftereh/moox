# Slice 17.5 Gate 3 - Tactical Combat Foundation Implementation Evidence

Date: 2026-09-29
Status: **Implemented; all five Gate 3 checkpoints complete; automated Gate 3 verification green; Gate 4 review remains open.**

## Frozen contract and implementation checkpoints

Gate 3 implements the frozen contract in:

- `docs/research/SLICE_17_5_GATE2_TACTICAL_FOUNDATION_CONTRACT_FREEZE_2026-09-29.md`
- `docs/research/SLICE_17_5_GATE3_IMPLEMENTATION_CHECKPOINT_PLAN_2026-09-29.md`

The recoverable checkpoint sequence completed as:

1. Tactical Ruleset and Ship Projection - `8715ae3 feat(slice): complete 17.5 gate 3 checkpoint 1`
2. Broad Combatant Admission - `5897755 feat(slice): complete 17.5 gate 3 checkpoint 2`
3. Shields and Shared Damage Foundation - `a55490c feat(slice): complete 17.5 gate 3 checkpoint 3`
4. Generic Weapon Action Availability and UI Consumption - `bd70189 feat(slice): complete 17.5 gate 3 checkpoint 4`
5. Integration, Regression and Gate 3 Evidence - this dedicated evidence/finalization change

No Slice 17.6 Beam breadth, Slice 17.7 Ordnance behavior, or Slice 17.8 Fighter/Combat-Special behavior is implemented by this Gate.

## Frozen foundation coverage

### Early / Mid / Late combatant matrix

The frozen six-hull Early/Mid/Late matrix is covered by the Slice-17.5 focused tests, including:

- `TestSlice175Checkpoint1TacticalFoundationMatrix`
- `TestSlice175Checkpoint1UnusedSpaceSpeedUsesCeiling`
- `TestSlice175Checkpoint1BaselineEncounterProjectsTruthfulFoundation`
- `TestSlice175Checkpoint2BroadCombatantAdmissionMatrix`
- broad BattleSession validation for derived combatants and malformed/unknown identities.

This covers Hull, Drive, Computer, Armor, Shield, Fuel, minimum/current combat speed, Beam Offense, Beam Defense and durability projection across the frozen tiers.

### Shield, damage and recharge vectors

The frozen shared-damage contract is covered by:

- `TestSlice175Checkpoint3ShieldClassesInitializeFourFacingPoolsAndView`
- `TestSlice175Checkpoint3ShieldImpactFacingBoundaries`
- `TestSlice175Checkpoint3FrozenSharedDamageVectors`
- `TestSlice175Checkpoint3DeterministicRechargeAndRoundTransition`
- `TestSlice175Checkpoint3BattleDamageEventCarriesShieldEnvelope`
- `TestSlice175Checkpoint3ShieldStateLiveSnapshotBounds`

These tests cover no-Shield Armor -> Structure flow, Class III block/absorption/overflow examples, deterministic four-arc recharge allocation, Shield overflow, Armor exhaustion, Structure destruction and event before/after envelopes.

### Generic weapon-action and unsupported-family vectors

Checkpoint 4 remains the UI-facing action authority and is covered by:

- `TestSlice175Checkpoint4ProjectsOneAuthoritativeActionPerInstalledSlot`
- `TestSlice175Checkpoint4StableUnavailableReasons`
- `TestSlice175Checkpoint4RejectedUnsupportedActionConsumesNoRNGOrReadiness`
- `TestSlice175Checkpoint4MixedWeaponProjectionCarriesAuthoritativeFamilies`

Supported Laser slots remain executable through the existing `battle.fire_beam` baseline. Later-family weapon slots remain admitted metadata but are explicitly unavailable with `tactical_consumer_unavailable`, zero targets and no RNG/readiness consumption on rejection. Legacy `legal_fire_actions` remains compatibility-only; `weapon_actions` is the UI-facing contract.

## Integration, observer/public projection and persistence evidence

Checkpoint 5 adds `TestSlice175Checkpoint5ObserverPlayerAndPersistenceBoundary`.

The test uses a mixed Laser/Missile active combatant plus a Class III Shield defender and verifies together that:

- Observer `View()` exposes the authoritative generic weapon descriptors and Shield foundation.
- The active participant receives its `weapon_actions` command authority.
- The non-active participant receives no weapon command authority.
- A real JSON marshal/unmarshal of `battle.LiveSnapshot` followed by `RestoreLiveSnapshot` preserves the immutable weapon-family foundation, Tactical runtime state and derived generic actions.
- Mutating a restored observer projection does not alias back into authoritative battle state.

The HTTP battle integration test `TestBattleCommandEndpointAcceptsParticipantTacticalCommand` now also verifies that the player snapshot publicly carries the generic Laser action descriptor and that the non-active seat receives no `weapon_actions`.

Existing deterministic replay and persistence boundaries were re-run explicitly:

- `TestTacticalReplayAndViewAreDeterministicDetached`
- `TestLiveSnapshotActiveTacticalRoundTripAndContinuation`
- `TestLiveSnapshotActiveTacticalSessionRoundTripAndContinuation`

## Final Gate 3 verification

Fresh Checkpoint-5 verification:

- `go test ./internal/game ./internal/battle ./internal/session -run TestSlice175 -count=1` - **PASS**
- `go test ./internal/battle -run TestTacticalReplayAndViewAreDeterministicDetached|TestLiveSnapshotActiveTacticalRoundTripAndContinuation -count=1` - **PASS**
- `go test ./internal/session -run TestLiveSnapshotActiveTacticalSessionRoundTripAndContinuation -count=1` - **PASS**
- `go test ./internal/server -run TestBattleCommandEndpointAcceptsParticipantTacticalCommand -count=1` - **PASS**
- `go test ./... -count=1` - **PASS**
- `go vet ./...` - **PASS**
- `npm --prefix web run build` - **PASS**, including UTF-8 checks, TypeScript compilation and Vite production build.
- Fresh `-reference-games` server on `127.0.0.1:7215` plus `npm --prefix web run check:reference-lab:browser -- http://127.0.0.1:7215` - **PASS**.
- `git diff --check` - **PASS** after documentation finalization.

The browser smoke additionally re-verifies the existing six-hull All-Tech Designer/reference path, mobile/desktop layout and ordinary-game isolation while the Tactical foundation changes remain server-authoritative.

## Ownership and Gate 4 boundary

Gate 3 is complete without broadening child-slice ownership:

- Slice 17.6 still owns full Tactical Beam/modifier hit, range, arc, damage and effect breadth.
- Slice 17.7 still owns projectile/rack/ammo/flight/interception behavior.
- Slice 17.8 still owns Fighter launch/return and combat-Special consumers.
- Slice 17.5 provides only the common combatant/action/RNG/damage/event foundations those children must consume.

Slice 17.5 remains open. The next action is **Gate 4 review**, not a Slice close and not downstream Slice 17.6 implementation.
