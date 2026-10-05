# OPEN Slice 17.8 - Tactical Fighters / Bombers / Combat Specials

Opened: 2026-10-02

Planned specification:
`docs/slices/PLANNED_17_8_TACTICAL_FIGHTERS_BOMBERS_COMBAT_SPECIALS.md`

Gate-1 evidence:
`docs/research/SLICE_17_8_GATE1_FIGHTERS_SPECIALS_TACTICAL_AUDIT_2026-10-02.md`

## Opening baseline

- branch: `main`
- opening HEAD: `0a3c325 feat(slice): close 17.7 tactical ordnance`
- opening Working Tree: clean
- opening divergence: `0 ahead / 0 behind` versus `origin/main`
- other `_OPEN_*.md` markers at open time: none
- Slice 17.4 Ship Specials / Fighter Systems: closed/accepted
- Slice 17.5 Tactical Combat Foundation: closed/accepted
- Slice 17.6 Tactical Beam Weapons / Modifiers: closed/accepted
- Slice 17.7 Tactical Missiles / Torpedoes / Bombs: closed/accepted

## Current gate

**Gate 1 - fighter/special Tactical audit: IN PROGRESS**

No Slice-17.8 production implementation is allowed before Gate 1 closes and Gate 2 freezes the supported fighter/special combat contract.

## Gate-1 audit threads

- Fighter Bays / Bomber Bays / Heavy Fighter Bays identity, craft-per-bay, launch, readiness and loss/recovery semantics.
- Fighter/Bomber/Interceptor actor/group representation, movement, targeting, attack and return.
- Bomber target eligibility and damage context.
- Beam / Point Defense / ordnance / defensive-system interactions with fighter actors.
- Full reclassification of every 17.4 Class-3 Special into strategic-only, Tactical-common, Beam-related, ordnance-related, fighter-related, boarding-related or unsupported/deferred.
- Explicit boarding audit for Transporters, Troop Pods, Security Stations, Marine state, repel/capture flow and ownership change.
- Repair/regeneration, cloak/phasing/phase-shift, ECM/jamming/anti-missile, shield/damage-field, teleport/time/area and remaining combat-special effects.
- BattleSession persistence/replay/Observer requirements for launched craft and temporary special state.
- Tactical UI/action/event/presentation requirements.
- Deterministic original/reference vectors for representative carrier, boarding and special-effect cases.

## Scope guard

- Do not reopen 17.6 Beam or 17.7 ordnance formulas unless Gate-1 evidence demonstrates a required shared hook.
- Do not make metadata-only 17.4 rows executable merely because they are Designer-legal.
- Do not infer browser-side fighter physics or special effects.
- Keep strategic-only effects strategic.
- Unsupported/unproven behavior stays explicit rather than approximated.
- Planetary Fighter Garrison / orbital-planetary combat remains reserved for Slice 19; Slice 17.8 builds the reusable ship-Tactical fighter/combat-special runtime.

## Current audit checkpoint

Checkpoint D now also freezes Capture reconciliation:

- successful Capture transfers surviving attacking Marines into target `MarineCurrent` before owner change;
- failed Capture leaves surviving defenders and does not change owner;
- target officer is removed before Capture reconciliation;
- Quantum Detonator has an exact 60% capture-time self-destruction chance;
- surviving captured ships change owner, retarget inbound missiles and remove continuing effects;
- non-Telepathic owners receive the original `post_capture_state`; Telepathic owners bypass it;
- original Tactical AI owns boarding decisions.

## Next

Continue Gate 1 with a representative Fighter Beam RNG vector, then the full Class-3 Special effect/order matrix and authoritative command/event/UI/persistence contracts before any Gate-2 freeze.


### Checkpoint E update

The representative Fighter-Beam RNG/damage vector is now frozen from direct `Fighter_Ocv_` / `Fire_Fighter_Beam_` disassembly:

- Fighter OCV composition is explicit;
- ship Beam Defense vs moving-actor `Missile_Dcv_` target-defense split is explicit;
- `Random_(100)` accuracy order, forced-`100` branch for 96..99, clamp and inclusive hit threshold 40 are explicit;
- integer damage interpolation and the visible `flag & 0x04` pre-hit modifier ordering are explicit;
- representative deterministic hit/miss/damage vectors are recorded in Gate-1 evidence.

The Fighter-Beam vector blocker is closed. Gate 1 remains open for the full Class-3 Special effect/order matrix plus command/event/UI/persistence/AI contracts.

## Next

Continue Gate 1 with the full Class-3 Special support/effect/order matrix and authoritative command/event/UI/persistence/AI contracts before any Gate-2 freeze.


### Checkpoint F update

The full deferred Class-3 row list is now classified per row with an explicit evidence level:

- **FROZEN** where direct executable arithmetic/order is sufficient;
- **PROVEN / formula open** where a Tactical consumer is directly evidenced but final arithmetic/order is not yet closed;
- **HELP-classified / consumer open** where effect family is source-backed but an executable consumer still needs recovery.

New directly frozen/advanced items include:

- Automated Repair Unit: base combat repair contribution 20 plus Fleet Engineer bonus; stasis/post-capture suppression is explicit.
- Shield Capacitors: shield-recharge budget changes from 30% to 100% per invocation and is distributed pointwise across deficient facings.
- Heavy Armor: original repair capacity path directly checks ID 14 and triples the computed armor capacity.
- Displacement Device: a direct special-attack path consumes ID 8 and authoritative RNG; generic all-weapon ordering remains open.
- Achilles / Structural Analyzer: shared ship-damage resolver receives their presence flags; exact 17.6 damage-order semantics remain to be cross-checked.
- Warp Dissipator: frozen original row 37 exists, but the current generated `ship_specials.json` lacks the named row; Gate 2 must resolve this projection gap explicitly.

Per-row classification is no longer a Gate-1 blocker. Exact formulas/order for the remaining open rows plus active command/event/persistence/AI contracts are still required.

## Next

Recover the remaining executable placement/order for Beam-support, ECM/jamming, cloak/phasing, active teleport/time, Lightning/Reflection/Energy-Absorber and retreat-control effects; then freeze the server command/event/projection/persistence/AI contract before Gate 1 closure.


### Checkpoint G update

Cross-slice reconciliation with the accepted 17.6 original Beam audit has upgraded several Class-3 rows from HELP-only/partial evidence to executable Beam hooks:

- Rangemaster Unit: range input /3 before Beam To-Hit lookup.
- High Energy Focus: +50 to Beam damage multiplier.
- Structural Analyzer: x2 downstream post-Shield damage.
- Hard Shields: cancels Shield Piercing in addition to its frozen Transporter rejection.
- Heavy Armor: participates in Armor-Piercing resistance in addition to its recovered armor-capacity hook.
- Damper Field: suppresses the Beam Marine-kill hook in addition to its boarding interaction.
- Achilles Targeting Unit: executable deeper-internal damage-selection hook confirmed; exact subsystem weighting remains open.

17.8 must reuse the frozen 17.6 Beam resolver hooks instead of duplicating Beam formulas.

## Next

Continue non-Beam Special placement/order: ECM/jammers and Lightning against moving actors; cloak/phasing state machine; Reflection/Energy Absorber; teleport/time active commands; Warp Dissipator retreat legality; then command/event/projection/persistence/AI contract.

### Checkpoint H update

The non-Beam defensive/readiness Special audit now has direct executable ordering for:

- Lightning Field pre-impact grouped-actor RNG/multiplicity mutation before reactive PD/MIRV;
- Displacement Device's distinct ordnance and Beam forced-miss branches;
- ECM / Multi-Wave ECM / Wide Area Jammer self and fleet values with best-only stacking;
- Cloaking Device active +80 Beam Defense and 50% missile forced-miss hook;
- directly evidenced Cloak/Phasing state transitions, with unresolved recloak/10-turn lifetime kept OPEN;
- Energy Absorber 1/4 incoming-damage storage before Damper Field plus automatic pre-weapon discharge;
- Damper Field quartering of remaining incoming damage;
- Reflection Field's direct `Random_(D+15) > D` full-reflection branch;
- Hyper-X Capacitors / Fast Missile Racks shared weapon-slot readiness state and executable-vs-HELP re-arm discrepancy;
- Anti-Missile Rockets ownership through the frozen 17.7 projectile-interceptor hook;
- Warp Dissipator ID37 blocking both AI and manual Tactical retreat while a functional enemy carrier exists.

Gate 1 remains open for active teleport/time mechanics, unresolved cloak/absorber lifetimes, Achilles internal weighting, command/event/UI/persistence/AI contracts and remaining deterministic Special vectors.

## Next

Audit Phase Shifter / Sub-Space Teleporter command legality and use state, then Time Warp Facilitator scheduler semantics. Afterwards close the remaining lifecycle/persistence/event/AI blockers before any Gate-2 freeze.

### Checkpoint I correction/update

Full `Do_Combat_Turn_` recovery supersedes several provisional Checkpoint-H OPEN statements:

- Hyper-X Capacitors / Fast Missile Racks **do** re-arm after one complete non-firing turn: end-pass changes surviving readiness state 2 -> 1.
- ordinary Cloaking state 3 -> 1 at end-pass; firing during the waiting turn sends 3 -> 2 and prevents recloak.
- Phasing Cloak state 4 decrements its initialized counter 10; zero converts to ordinary Cloaking state 1; broken state 7 returns to state 4 and retargets relevant incoming actors.
- Energy Absorber stored energy is cleared at the end of the acting side's pass if not discharged.
- IDs 22/33 share once-per-turn self-teleport state; direct executable range is <=20, teleporter use avoids ordinary movement-cost subtraction and preserves facing.
- Phase-Shifter other-ship teleport remains unproven despite the presence of an unreferenced named `Teleport_Enemy_Ship_` routine.
- Time Warp Facilitator ID34 is implemented as a second `Do_Combat_Turn_(side,1)` pass for each side after both normal passes; only ID34 ships participate, they receive normal start-turn initialization, Seeking Missiles is skipped in the extra pass, and shield recharge/Automated Repair remain once-per-round global bookkeeping.

## Next

Recover Achilles exact internal-selection weighting if possible, then freeze the authoritative command/rejection/event/projection/persistence/AI contract and final deterministic vectors required to close Gate 1.

### Checkpoint J update

Achilles Targeting Unit is now fully recovered from the executable Beam/internal-damage path:

- attacker Achilles forces Armor bypass before Heavy-Armor/Xentronium cancellation logic;
- `Select_Internal_` multiplies all eight Weapon-slot selection weights by 3;
- the Shield-generator selection weight is multiplied by 3;
- other recovered internal-category weights remain unchanged;
- the existing shared weighted-choice RNG is reused; Achilles adds no separate hit/damage RNG.

The Achilles formula blocker is closed.

## Next

Resolve the remaining Phase-Shifter other-ship dispatch question, then freeze Gate-1 command/rejection/event/projection/persistence/AI requirements and final closure vectors.
### Checkpoint K update

The remaining original-Special formula/reachability audit is now closed:

- Phase Shifter ID22 has a proven once-per-turn self-teleport at executable range <=20; the named `Teleport_Enemy_Ship_` routine has no accepted CALL/JMP/instruction-pointer/LE-fixup reachability in the audited 1.31 executable, so other-ship teleport is unsupported/deferred rather than inferred.
- Augmented Engines adds +5 combat speed when installed and functional.
- Battle Scanner contributes +50 to the Tactical offensive-combat bonus stack.
- Multi-Phased Shields applies a 3/2 max-shield transform after Hard-Shields/nebula survival.
- Hard Shields preserves shields through the nebula-suppression branch and subtracts an additional 3 damage during active shield absorption, in addition to the already frozen Shield-Piercing and boarding hooks.
- Reinforced Hull triples the final normal Structure pool.
- Quantum Detonator triples normal ship-explosion power in addition to its already frozen 60% capture-time self-destruction branch.
- Automated Repair now has exact durability/internal repair budgets and deterministic pool order.

The original-formula/reachability blocker is closed. Gate 1 remains open only for the MOOX command/event/projection/persistence/UI/AI contract mapping and final closure vectors/checklist.

## Next

Map the current BattleSession command/view/live-snapshot/AI gaps into the Gate-2 requirements, add final representative vectors, then evaluate Gate-1 completion without starting production implementation.
### Checkpoint L/M update

The current MOOX BattleSession/runtime gap is now mapped into explicit Gate-2 requirements without implementing production mechanics:

- extend the existing TacticalShipSpec/TacticalShipState/TacticalEvent/LiveSnapshot architecture rather than creating a parallel fighter/special engine;
- persist authoritative Fighter group lifecycle, carrier linkage, multiplicity/target/return state and shared BattleSession RNG ordering;
- keep only proven transient Special state mutable (cloak/phasing, Absorber store, teleport use-state, Hyper-X/FMR readiness/rearm); derive passive Specials from immutable installed identity;
- reuse MarineCurrent/MarineMax and add only boarding-used/post-capture/owner-control state required by the original lifecycle;
- add server-projected legal action families for Fighter launch, Boarding and directly reachable active Specials; rejection consumes zero RNG/state;
- add authoritative Fighter/Boarding/Special events with stable IDs and before/after mutation detail;
- extend LiveSnapshot restore validation so save/reconnect/replay is deterministic for every new mutable 17.8 field;
- expose the same projected actions to built-in AI; browser remains presentation-only.

Final deterministic closure vectors are now recorded for shield stack/Hard Shields, Energy Absorber -> Damper order, Reflection, Lightning grouped multiplicity, Shield Capacitor distribution, Quantum explosion, Automated Repair, Augmented Engines/Battle Scanner, with the earlier Fighter/Bomber/Boarding vectors retained as authoritative references.

## Gate 1 result

**Slice 17.8 Gate 1 is COMPLETE.**

All original formula/reachability questions accepted for 17.8 are frozen or explicitly unsupported/deferred, runtime/persistence/UI/AI requirements are mapped, and representative deterministic vectors exist.

No Slice-17.8 production implementation has started.

## Next

Start **Gate 2 - contract freeze** only: define exact MOOX state schemas, protocol/action names, event payloads, validation domains, persistence migration and implementation checkpoints from the Gate-1 evidence. Do not begin production mechanics until that Gate-2 contract is frozen.
### Gate 2 contract-freeze update

Gate 2 is now frozen in:

`docs/research/SLICE_17_8_GATE2_TACTICAL_FIGHTERS_SPECIALS_CONTRACT_FREEZE_2026-10-05.md`

Frozen contract highlights:

- Fighter/Bomber/Heavy-Fighter rows become `TacticalSupport="fighter"` with server-owned Bay/group lifecycle and shared moving-actor ordering;
- Fighter doctrine snapshots freeze best-drive, armor-durability, Trans-Dimensional and Fighter Beam/Bomb payload identity per participant;
- real internal-system state is introduced for Achilles, Raid damage and Automated Repair rather than faking subsystem effects;
- Boarding freezes `battle.board_ship` with immediate physical/Transporter resolution, temporary Raid Marine accounting, persistent Capture survivor/owner transfer, and server-projected legality;
- self-teleport reuses `battle.move_ship` with `movement_kind=teleport`; no generic client Special command is invented;
- mutable owner/controller, cloak/phasing, Absorber, teleporter, Boarding and weapon Special-readiness state are server-owned;
- LiveSnapshot v2 freezes all new mutable state; schema-1 active battles restore compatibly with 17.8 execution disabled rather than guessed hidden state;
- built-in AI and browser consume the same server-projected action/state/event surfaces;
- Warp Dissipator original row 37 must be restored to rules data before its Tactical consumer is enabled;
- Phase-Shifter other-ship teleport remains unsupported/deferred;
- Anti-Missile Rockets remain deferred until their original consumable live-count source can be represented without invention.

Gate-2 validation also refined Raid persistence: `Resolve_Raid_` restores target Marines and returns the temporary boarding party to the source; persistent Raid output is internal-system damage, while Capture owns durable Marine/ownership transfer.

## Gate 2 result

**Slice 17.8 Gate 2 is FROZEN / COMPLETE. Gate 3 implementation is ready but has not started.**

## Next

Begin Gate 3 only from the frozen six-checkpoint implementation plan: (1) rules/data/schema, (2) internal systems/passive derivation, (3) Fighter lifecycle, (4) Boarding/Capture, (5) Special lifecycle/scheduler, (6) persistence/AI/UI/integration.

### Gate 3 Checkpoint 1 - complete

Implementation evidence:

`docs/research/SLICE_17_8_GATE3_IMPLEMENTATION_EVIDENCE_2026-10-05.md`

Checkpoint-1 result: **PASS / COMPLETE**.

Implemented foundation:

- Fighter/Bomber/Heavy-Fighter support identity is now `fighter` with four craft per Bay unit;
- frozen Combat-Special support classes are projected and strictly validated;
- original Special row 37 is canonicalized to `warp_dissipator`;
- frozen Fighter/Boarding/Special rule blocks are in rules data and immutable Tactical snapshots;
- Fighter doctrine snapshots derive best-known Empire drive/armor plus Trans-Dimensional state;
- immutable Internal-system DTO/validation/clone shape is reserved for Checkpoint 2 consumers;
- LiveSnapshot v2 migration scaffolding is active with `legacy_17_8_disabled`, shared moving-actor ordering and deterministic schema-v1 projectile migration;
- reviewed ruleset SHA is `800c8813437503390d2420f8ee50eeb6e528b9bf960fce725f88448b966fda3a`.

Validation passed:

- `go test ./internal/ruleset`
- `go test ./internal/battle`
- `go test ./internal/game`
- `go test ./...`
- `git diff --check`

No Fighter actor lifecycle, Boarding resolver or Combat-Special effect consumer has been implemented yet.

### Next

Gate 3 Checkpoint 2 - **Internal systems / passive derivation**:

- immutable/current internal pools;
- mutable current combat speed;
- shared Special-functionality helper;
- Reinforced Hull / Heavy Armor / Hard Shields / Multi-Phased Shields;
- Augmented Engines / Battle Scanner / ECM derivations;
- internal-damage event foundation.

Do not advance to Checkpoint 3 before Checkpoint 2 is independently reviewed and green.
