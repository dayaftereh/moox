# Planned slice 07 - Tactical ship combat baseline

Status: **closed; Gates 1-4 complete**.

Queue position: **7 of 7**.

## Objective

Implement the first narrow, directly evidenced deterministic tactical ship-combat vertical slice on top of the BattleSession handoff.

This is intentionally a **baseline**, not an attempt to implement every MOO2 weapon/special/boarding/planet-defense rule at once. Gate 1 must choose the smallest canonical combat fixture that can run end-to-end and produce an authoritative battle result.

## Existing baseline

- `internal/battle` has deterministic session lifecycle, battle IDs/seeds and copy-safe views.
- GameSession can host parallel BattleSessions and commit their completion in stable order.
- Tactical ship artwork/hull identities are normalized, but tactical ship state, movement, firing and damage rules are not yet implemented.
- Slices 02/03/05 are intended to provide concrete military ships, Fleet composition and automatic strategic encounter handoff.

## Dependencies

- slice 03 military design/ship state;
- slice 04 Fleet composition/movement;
- slice 06 BattleSession strategic handoff.

## Scope guard

The final Gate-2 scope should be deliberately small. Candidate first fixture:

- two opposing basic military ships;
- deterministic tactical deployment/initiative;
- minimum movement/facing only if required by the selected weapon rule;
- one directly evidenced basic direct-fire weapon path;
- hit/damage to the minimum hull/armor/structure state;
- destruction/victory result returned to BattleSession.

Defer unless Gate 1 proves essential:

- missiles, bombs, fighters;
- boarding/marines;
- cloaking/stealth;
- specials;
- complex shield arcs/critical systems;
- retreat;
- planetary defenses;
- tactical AI;
- UI/animation/audio.

## Gate 1 - Checkup + original analysis / reverse engineering

- [x] Re-check `internal/battle`, military design state and encounter handoff.
- [x] Create permanent tactical-baseline evidence document.
- [x] Choose a canonical minimal original MOO2 1.31 combat fixture whose rules can be proven independently.
- [x] Verify tactical turn/initiative order for that fixture.
- [x] Verify deployment/position coordinates and movement/facing only as required by the fixture.
- [x] Verify the selected weapon's range, firing eligibility, hit-chance RNG ownership and damage calculation.
- [x] Verify the minimum shield/armor/structure interaction needed for the fixture.
- [x] Verify ship destruction and battle-victory termination semantics.
- [x] Verify Battle RNG seed/stream ownership and whether the current infrastructure seed must be replaced/augmented by original-derived RNG semantics.
- [x] Verify how tactical survivors/losses map back to strategic ship/Fleet state.
- [x] Explicitly list every combat family deferred from the first baseline.
- [x] Present Gate-1 evidence and a very narrow deterministic implementation contract; stop before implementation.

## Gate 2 - Implementation decision

- [x] Accept exact first tactical fixture and supported weapon/system subset.
- [x] Decide tactical Core state model: combatants, positions, facing, HP layers and turn state.
- [x] Decide battle command/event API and authority model.
- [x] Decide RNG stream ownership/replay strategy.
- [x] Decide how tactical snapshots/reference identities relate to strategic Ship instances.
- [x] Decide result/loss application back into strategic state.
- [x] Decide what remains hard-rejected rather than silently approximated.

## Gate 3 - Implementation

- [x] Implement deterministic tactical state for the accepted fixture.
- [x] Implement turn/initiative and legal action validation.
- [x] Implement selected movement/firing path only.
- [x] Implement hit/damage/destruction layers proven by Gate 1.
- [x] Emit deterministic battle events suitable for replay/Observer views.
- [x] Produce an authoritative BattleSession result from the tactical engine.
- [x] Apply survivor/loss result back to strategic ship/Fleet state through the slice-06 handoff.
- [x] Add exact fixed-seed tactical fixtures and save/replay tests.
- [x] Add rejection tests for explicitly unsupported combat systems.

## Gate 4 - Follow-up QA + commit + close

- [x] `gofmt` changed Go files.
- [x] `go test ./... -count=1`.
- [x] `go vet ./...`.
- [x] `git diff --check`.
- [x] Run focused battle/session/strategic handoff and fixed-seed combat regressions.
- [x] Verify tactical result order/replay is independent of wall-clock execution.
- [x] Verify strategic ship losses/survivors round-trip exactly.
- [x] Update evidence/status/HISTORY and close marker.

Gameplay/data/evidence commit: `889f977` (`game: add tactical ship combat baseline`).

## Exit criterion

One deliberately narrow original-derived ship-vs-ship tactical fixture can be created from a strategic encounter, executed deterministically to victory/destruction, recorded/replayed, and applied back to strategic Fleet state. Unsupported combat families remain explicit rather than approximated.
