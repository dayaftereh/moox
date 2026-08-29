# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: **none**.
- Active slice: **none**; the Race-aware Population cohorts slice is complete.
- Latest completed gameplay slice: Race-aware Population cohorts (`7c49e90`).
- Core `StateSchemaVersion`: **14**.
- Economy ruleset schema: **7**.
- Latest permanent evidence: `docs/research/POPULATION_COHORTS_2026-08-29.md`.
- Previous closed-slice evidence: `docs/research/POPULATION_CAPACITY_TRANSITIONS_2026-08-29.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; resume any marker before opening another slice.

## Recently closed runtime sequence

The current Research/Economy chain is implemented and recorded in `docs/slices/HISTORY.md`, including:

- race-aware multi-Technology research, switching, Uncreative initialization and repair;
- Hyper-Advanced repeated fields and Advanced-start randomized/race-aware ownership;
- original strategic Research -> Population -> Construction ordering with pre-growth RP/PP snapshots;
- Freighter Fleet production, Treasury settlement, constrained Food imports and system-blockade Food exclusion;
- Population relocation through the shared Freighter pool;
- Housing, Microbiotics, Universal Antidote and Cloning Center Population-growth modifiers;
- Advanced City Planning +5 and Biospheres +2 Population-capacity layers;
- Terraforming and Gaia Transformation as one-shot planetary climate projects with original legality/mappings and deterministic Barren middle-orbit RNG;
- Core schema 14 organic Population cohorts with origin/loyalty/assimilation identity;
- race-aware mixed-colony Economy, exact four-pass Food priority, per-origin Growth/Starvation and heterogeneous capacity;
- cohort-aware Population transfer preserved through GameSession, Observer and replay.

## Next queued objective

### Canonical strategic Fleet/Diplomacy state for blockade production

Do **not** create an `_OPEN_*.md` marker until work on this next slice actually begins. When it begins, use the four-gate protocol in `docs/slices/README.md`.

The current Food-logistics consumer already understands authoritative per-system blockade state, but blockade production is still deliberately not invented from temporary data. The next queued slice should establish the minimum canonical strategic Fleet/Diplomacy state needed to derive that blockade state deterministically.

Gate 1 should establish from original MOO2 1.31 executable/data/save evidence where possible:

1. the minimum authoritative strategic Fleet identity/location/owner state required for blockade production;
2. how hostile/friendly/diplomatic relationships participate in system blockade eligibility;
3. when blockade state is derived or refreshed in the strategic turn lifecycle;
4. whether multiple fleets/empires combine, cancel or otherwise affect blockade presence;
5. the minimal Core/Game/Session/Observer schema that lets existing Food logistics consume derived blockades without importing tactical-combat scope.

### Scope guard

Keep the next slice focused on canonical strategic Fleet/Diplomacy state **only as far as required to produce blockade state**. Do not absorb tactical combat, ship-design fidelity, broad diplomacy negotiations, AI strategy or UI work unless direct evidence makes one inseparable from blockade legality.

## Planned queue after blockade production

1. Missing original Treasury income/Maintenance categories and deficit/scrap policy once their dependent systems exist.
2. Continue the first headless vertical game loop toward colony-ship production, strategic movement/colonization and a second colony.

## Explicitly parked / UI-only fidelity

- Active conquest/occupation/automatic-assimilation progression, Android/Native population and persisted custom race designs remain later Population extensions; Gate 3 intentionally did not absorb them.
- The original Hyper-Advanced research-selection screen temporarily previews counters at `completed + 1` and has a 20-level list boundary. MOOX deliberately keeps authoritative strategic cost/progression separate from that original UI quirk.
- Original copyrighted assets remain private reference material and are not distributable MOOX content.
- Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.
