# Active research / implementation handoff

This file is the authoritative **live** handoff for the next Master of Orion X implementation slice.
Permanent reverse-engineering evidence and closed results live in `docs/research/`. The completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: **none**. `docs/slices/_OPEN_*.md` must therefore be empty.
- Latest completed gameplay slice: Population relocation through the shared Freighter pool (`f816eaf`).
- Core `StateSchemaVersion`: **11**.
- Do not start a later slice while an `_OPEN_*.md` marker exists; resume that marker first according to `docs/slices/README.md`.

## Closed sequence - do not rediscover without contradictory evidence

The recent Research/Economy sequence is closed and recorded in `docs/slices/HISTORY.md`. In particular, the following are implemented rather than pending:

- General / ordinary / Creative / Uncreative multi-Technology research policies;
- exact research-project switching with full RP transfer;
- verified MOO2 strategic Research -> Population Growth/Starvation -> Construction ordering with pre-growth RP/PP snapshots;
- Uncreative initial fixed-choice generation from the shared New Game RNG;
- Hyper-Advanced repeat-field state, strategic cost progression and repeat completion;
- Advanced-start randomized/race-aware technology ownership with the verified 19 extra grants;
- authoritative external Technology grants plus verified Uncreative fixed-choice repair;
- Freighter Fleet construction (50 PP -> 5 Empire Freighters);
- authoritative modeled Treasury settlement;
- original insufficient-Freighter round-robin import priority;
- authoritative system-blockade exclusion from Food import/export/sale pools;
- Population relocation / Settlers using the shared Freighter pool, including same-system moves, interstellar reservation, ETA and arrival/loss handling.

Historical research documents may still say that a later item was "deferred" at the time that document was written. Treat `README.md`, `docs/PROJECT_STATUS.md`, this file and `docs/slices/HISTORY.md` as the current status surfaces.

## Next slice to start

### Population growth building / medicine modifiers

Investigate the original MOO2 1.31 Housing / Cloning Center / medicine Population-growth modifiers before extending the current classic base growth curve.

Gate 1 should establish, from executable/data/save evidence where possible:

1. the exact original functions/data paths for Housing, Cloning Center and medicine/Technology growth bonuses;
2. whether each modifier is additive Population/turn, multiplicative, capacity-dependent, Production-dependent or applied in another explicit form;
3. stacking/order against race growth multipliers and the already implemented capacity-limited classic curve;
4. the minimal normalized Building/Technology identities/effects required by the runtime;
5. the deterministic Core/Game/Session test shape while preserving the verified rule that freshly grown Population cannot contribute RP/PP retroactively in the same turn.

At the moment this slice actually starts, create a recovery marker such as:

```text
docs/slices/_OPEN_POPULATION_GROWTH_MODIFIERS_2026-08-29.md
```

Do not implement gameplay before the original-evidence check and implementation shape have been recorded according to the four-gate slice protocol.

## Planned queue after growth modifiers

1. Biospheres / Advanced City Planning / terraforming Population-capacity transitions.
2. Race-aware Population cohorts, including conquered/mixed-race colonies and the later original Food-priority passes.
3. Full Fleet/Diplomacy-derived `Compute_Blockades_` materialization once canonical strategic Fleet/Diplomacy state exists.
4. Missing original Treasury income/Maintenance categories and deficit/scrap policy once their dependent systems exist.
5. Continue the first headless vertical game loop toward colony-ship production, strategic movement/colonization and a second colony.

## Explicitly parked / UI-only fidelity

- The original Hyper-Advanced research-selection screen temporarily previews counters at `completed + 1` and has a 20-level list boundary. MOOX deliberately keeps authoritative strategic cost/progression separate from that original UI quirk; revisit only when implementing UI-fidelity behavior.
- Original copyrighted assets remain private reference material and are not distributable MOOX content.
- Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.

## Slice close checklist

Before closing any future slice:

- materialize original evidence / provenance in a permanent `docs/research/...` document;
- synchronize runtime/data/API decisions and deterministic tests;
- run `gofmt` on changed Go files, `go test ./...`, `go vet ./...`, and `git diff --check`;
- update the current live status surfaces;
- add the slice to `docs/slices/HISTORY.md` with closing commit and evidence document;
- delete its `_OPEN_*.md` marker in the closing commit;
- leave the next slice unmarked until it is actually started.