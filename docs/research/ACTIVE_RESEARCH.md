# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: **none**.
- Active slice: **none**; the Population capacity transition slice is complete.
- Latest completed gameplay slice: Population capacity transitions (`25d05a9`).
- Core `StateSchemaVersion`: **13**.
- Economy ruleset schema: **7**.
- Permanent evidence: `docs/research/POPULATION_CAPACITY_TRANSITIONS_2026-08-29.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; resume any marker before opening another slice.

## Recently closed runtime sequence

The current Research/Economy chain is implemented and recorded in `docs/slices/HISTORY.md`, including:

- race-aware multi-Technology research, switching, Uncreative initialization and repair;
- Hyper-Advanced repeated fields and Advanced-start randomized/race-aware ownership;
- original strategic Research -> Population -> Construction ordering with pre-growth RP/PP snapshots;
- Freighter Fleet production, Treasury settlement, round-robin constrained Food imports and system-blockade Food exclusion;
- Population relocation through the shared Freighter pool;
- Housing, Microbiotics, Universal Antidote and Cloning Center Population-growth modifiers;
- Advanced City Planning +5 and Biospheres +2 Population-capacity layers;
- Terraforming and Gaia Transformation as one-shot planetary climate projects with original legality/mappings and deterministic Barren middle-orbit RNG;
- schema-13 persisted planetary-transformation construction plus authoritative GameSession/Observer lifecycle.

## Next queued objective

### Race-aware Population cohorts / conquered and mixed-race Colonies

Do **not** create an `_OPEN_*.md` marker until work on this next slice actually begins. When it begins, use the four-gate protocol in `docs/slices/README.md`.

Gate 1 should establish from original MOO2 1.31 executable/data/save evidence where possible:

1. authoritative Population cohort identity/storage semantics for native, conquered and mixed-race Colonies;
2. how Farmers/Workers/Scientists are associated with race/status cohorts and how assignments move between them;
3. exact race-specific capacity handling inside one mixed Colony, including removal priority when capacity falls;
4. the remaining original four-pass Food import priority thresholds that depend on cohort/status populations;
5. assimilation/conquest ownership transitions relevant to cohort state;
6. minimal Core/Game/Session schema and deterministic test shape without expanding into unrelated tactical or diplomacy systems.

### Scope guard

Keep the next slice focused on Population cohorts and the Food-priority behavior that directly depends on them. Do not absorb Fleet/Diplomacy blockade production, unrelated Building effects, UI fidelity or Treasury categories unless direct evidence makes one inseparable.

## Planned queue after cohorts

1. Full Fleet/Diplomacy-derived blockade production once canonical strategic Fleet/Diplomacy state exists.
2. Missing original Treasury income/Maintenance categories and deficit/scrap policy once dependent systems exist.
3. Continue the first headless vertical game loop toward colony-ship production, strategic movement/colonization and a second colony.

## Explicitly parked / UI-only fidelity

- The original Hyper-Advanced research-selection screen temporarily previews counters at `completed + 1` and has a 20-level list boundary. MOOX deliberately keeps authoritative strategic cost/progression separate from that original UI quirk.
- Original copyrighted assets remain private reference material and are not distributable MOOX content.
- Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.
