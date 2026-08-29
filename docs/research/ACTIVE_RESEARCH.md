# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: `docs/slices/_OPEN_POPULATION_CAPACITY_TRANSITIONS_2026-08-29.md`.
- Active slice: **Population capacity transitions**.
- Latest completed gameplay slice: Population growth building / medicine modifiers (`e7f0d72`).
- Core `StateSchemaVersion`: **12**.
- Economy ruleset schema: **6**.
- Permanent evidence: `docs/research/POPULATION_GROWTH_MODIFIERS_2026-08-29.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; resume any marker before opening another slice.

## Recently closed runtime sequence

The current Research/Economy chain is implemented and recorded in `docs/slices/HISTORY.md`, including:

- race-aware multi-Technology research, switching, Uncreative initialization and repair;
- Hyper-Advanced repeated fields and Advanced-start randomized/race-aware ownership;
- original strategic Research -> Population -> Construction ordering with pre-growth RP/PP snapshots;
- Freighter Fleet production, Treasury settlement, round-robin constrained Food imports and system-blockade Food exclusion;
- Population relocation through the shared Freighter pool;
- Housing as a semantic continuous Production project;
- Microbiotics +25 percentage points and Universal Antidote +50 percentage points on natural Population Growth;
- Cloning Center flat +0.1 Population/turn, separate from the natural-growth multiplier;
- original Housing Production scaling using Production available after Cybernetic sustenance;
- schema-12 persisted Housing state plus authoritative GameSession/Observer lifecycle.

Historical research documents may still describe later items as deferred at the time they were written. Treat `README.md`, `docs/PROJECT_STATUS.md`, this file and `docs/slices/HISTORY.md` as the current status surfaces.

## Current objective

### Biospheres / Advanced City Planning / terraforming Population-capacity transitions

Current gate: **Gate 4 - QA / commit / close in progress**. Gate 3 implementation is complete; run full repository QA, commit the implementation, reconcile status/history, remove the open marker and close the slice.

Gate 1 should establish from original MOO2 1.31 executable/data/save evidence where possible:

1. exact capacity effects and ordering for Biospheres;
2. exact Advanced City Planning capacity effect and ownership/application semantics;
3. terraforming/climate-transition effects on current and maximum Population;
4. interaction with Aquatic, Tolerant and Subterranean capacity rules already normalized;
5. whether capacity decreases can force Population loss immediately or only clamp future growth;
6. minimal normalized Building/Technology/runtime identities plus deterministic Core/Game/Session test shape.

### Scope guard

Keep the next slice focused on Population **capacity transitions**. Do not expand it into mixed/conquered Population cohorts, unrelated Building economy effects, Fleet/Diplomacy blockade production, UI fidelity or missing Treasury categories unless direct evidence makes one inseparable.

## Planned queue after capacity transitions

1. Race-aware Population cohorts, including conquered/mixed-race Colonies and later original Food-priority passes.
2. Full Fleet/Diplomacy-derived blockade production once canonical strategic Fleet/Diplomacy state exists.
3. Missing original Treasury income/Maintenance categories and deficit/scrap policy once dependent systems exist.
4. Continue the first headless vertical game loop toward colony-ship production, strategic movement/colonization and a second colony.

## Explicitly parked / UI-only fidelity

- The original Hyper-Advanced research-selection screen temporarily previews counters at `completed + 1` and has a 20-level list boundary. MOOX deliberately keeps authoritative strategic cost/progression separate from that original UI quirk.
- Original copyrighted assets remain private reference material and are not distributable MOOX content.
- Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.

## Slice close checklist

Before closing any future slice:

- materialize original evidence / provenance in a permanent `docs/research/...` document;
- synchronize runtime/data/API decisions and deterministic tests;
- run `gofmt` on changed Go files, `go test ./...`, `go vet ./...`, and `git diff --check`;
- update the current live status surfaces;
- add the slice to `docs/slices/HISTORY.md` with implementation commit and evidence document;
- delete its `_OPEN_*.md` marker in the closing documentation commit;
- leave the next slice unmarked until it is actually started.
