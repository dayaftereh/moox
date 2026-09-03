# Open slice 11 - Troop Transport, invasion and conquest baseline

Status: **Gate 4 independent QA complete; implementation/evidence commit + final closure pending**.

Opened: **2026-09-03**

Planned specification: `docs/slices/PLANNED_11_TROOP_TRANSPORT_INVASION_CONQUEST_BASELINE.md`

Permanent evidence: `docs/research/TROOP_TRANSPORT_INVASION_CONQUEST_BASELINE_2026-09-03.md`

Starting HEAD: `f353a05` (`docs: close diplomacy war peace slice`)

## Gate state

- [x] Gate 1: current-code checkup + original 1.31 Transport/invasion/conquest research.
- [x] Gate 2: accept/freeze exact state/session/ground-combat/conquest contract.
- [x] Gate 3: implementation + deterministic integration regressions.
- [ ] Gate 4: independent QA + commits + close.

## Scope guard

Gate 1 is research only. Do not implement Transport, invasion or conquest gameplay before Gate 2 freezes the contract.

In scope for the eventual implementation: concrete Troop Transport identity/payload/movement; enemy-Colony invasion eligibility after space-combat requirements; one directly evidenced ground-combat fixture; deterministic casualties/capture; Colony ownership transfer; conquered Population loyalty/assimilation initialization; transport cleanup; post-conquest Economy/Food/blockade/action recalculation.

Deferred: bombardment/bombs, bio-weapons, full marine/ground-tech/race modifier matrix, Colony destruction, simultaneous invasions, tactical ground-combat UI, advanced rebellion/occupation and AI invasion planning.

## Recovery note

Gates 1-3 are complete. Resume only at Gate 4 for independent final QA, HISTORY/status closure, commits and removal of this single OPEN marker. Do not expand into the deferred ground-war families.
## Gate-1 result

- Original Transport is fixed special Ship type 2, base cost 100 PP, 1 CP, normal strategic movement and noncombat civilian behavior.
- One extant Transport implicitly carries exactly 4 Infantry; no persistent partial cargo field is needed.
- Invasion is a post-space-combat Colony-resolution action before later Blockade/Settler/Colony settlement.
- Zero-modifier baseline Ground Combat is opposed `Random_(100)` rolls; Infantry score 0, Militia score -10, one-hit supported units, ties damage both sides.
- Basic Militia is derived as Population / 5.
- Failed invasion destroys every selected Transport. Successful invasion reserves one Infantry for the Colony, preserves only complete remaining groups of four as surviving Transports and moves the remainder into standing Colony Infantry.
- Capture maps to existing `conquered` Population semantics, clears Construction, reassigns a lost Capital to the defender's largest remaining Colony and leaves Empire elimination for Slice 12.
- Proposed Gate 2 adds Core23 standing Infantry + `troop_transport`, construction/CP/movement support, a new `invasion_decisions` Session phase, immediate `invasion.invade|decline`, narrow deterministic Infantry/Militia combat and pre-Blockade ownership transfer.
- Full findings and deferrals are in `docs/research/TROOP_TRANSPORT_INVASION_CONQUEST_BASELINE_2026-09-03.md`.

No gameplay implementation occurred in Gate 1.
## Gate-2 result

- Core23 frozen: `troop_transport` fixed civilian special, persisted Colony standing Infantry, Capital-owner validation; no variable Transport payload.
- Transport frozen at 100 PP base, existing Feudal reduction, no invented tech prerequisite, 1 CP at normal settlement, existing Fleet movement and civilian Encounter participation.
- New `invasion_decisions` Session phase accepted. One canonical active opportunity plus transient per-turn handled `(ColonyID, AttackerEmpireID)` keys prevents same-turn re-offer after decline/failure/success; independent opportunities may resolve sequentially.
- Orbital-clear baseline requires war, authorized attacker Seat, stationary combat Fleet + Transport, no remaining hostile combat Fleet in the System and no defender-owned modeled Star Base/Battlestation/Star Fortress to bypass.
- Ground Combat frozen to zero-modifier Infantry/Militia fixture: Transport=4 Infantry, standing Infantry persisted, Militia=`floor(non-conquered organic Population/5)`, Infantry score 0, Militia -10, attacker then defender d100, tie damages both, one hit per supported unit.
- Failure consumes selected Transports and preserves ownership; capture uses original-style full-four survivor packing, clears Construction, initializes conquered/assimilated cohorts, reassigns Capital and leaves war active.
- Blockade/Population-transfer/Colony-Economy/Food continuation runs after invasion; Treasury/CP is not double-settled in the same turn.
- Immediate commands remain revision-bound on the existing endpoint. Event/API/Command schema remain 1; Economy ruleset schema remains 8.
- All advanced ground/bombardment/Telepathic/AI/simultaneous-invasion families remain explicitly deferred.

Gate 3 may now implement; Gate 4 remains untouched.

## Gate-3 result

- Core23, standing Infantry, fixed Troop Transport construction/movement/1-CP/Encounter identity implemented.
- `invasion_decisions`, canonical opportunity + handled keys and revision-bound `invasion.invade|decline` implemented server-authoritatively.
- Narrow original-evidenced Infantry/Militia d100 Ground Combat, failure cleanup, successful Transport survivor packing, Colony ownership/cohort/Construction/Capital handoff and post-conquest strategic/economic continuation implemented.
- Attacker/player/observer projection, existing HTTP immediate-command path and minimal React Invasion panel implemented.
- Deterministic success/failure, stale/invalid atomicity, Core23 save/load, deterministic replay, declared-war + transit-arrival + capture integration and HTTP capture tests are green.
- Core23 New-Game seed `0x8009` fingerprint: `83614740b216409877b03c536a16328c7fa7031867f58dbeb70857a8953f5762`.
- Final Gate-3 QA: full Go tests, vet, web build and diff check green.

Gate 4 remains open. No commit or push was performed in Gate 3.


Gate-4 independent QA is complete. Exact captured ownership/cohort/Capital save-load, no double Treasury/CP settlement and handled-key reset now have explicit closure regressions. Full Go tests, vet, web build and diff checks are green. Final closure must commit implementation/evidence, synchronize HISTORY/live status, remove this marker and verify a clean repository.
