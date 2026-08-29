# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: **none**.
- Active slice: **none**; the Canonical strategic Fleet/Diplomacy blockade-production slice is complete.
- Latest completed gameplay slice: Canonical strategic Fleet/Diplomacy blockade production (`d148502`).
- Core `StateSchemaVersion`: **15**.
- Economy ruleset schema: **7**.
- Latest permanent evidence: `docs/research/FLEET_DIPLOMACY_BLOCKADES_2026-08-29.md`.
- Previous closed-slice evidence: `docs/research/POPULATION_COHORTS_2026-08-29.md`.
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
- cohort-aware Population transfer preserved through GameSession, Observer and replay;
- Core schema 15 strategic Fleet ownership/location plus directed Empire relations;
- deterministic system-blockade production from Colony presence + stationary combat Fleets + directed hostility, materialized before Settler arrival/final Food.

## Next queued objective

### Missing original Treasury income/Maintenance categories and deficit/scrap policy

Do **not** create an `_OPEN_*.md` marker until work on this next slice actually begins. When it begins, use the four-gate protocol in `docs/slices/README.md`.

The current Treasury runtime already has the directly verified 50 BC New Game balance and strategic settlement of the currently modeled Tax + surplus-Food income against Building + active-Freighter Maintenance. The next queued slice should extend that model only with original categories and deficit behavior that can be evidenced without inventing dependencies that do not yet exist.

Gate 1 should establish from original MOO2 1.31 executable/data/save evidence where possible:

1. which additional income and Maintenance categories are part of the strategic Treasury settlement path;
2. which categories can be represented with systems MOOX already owns and which must remain deferred until dependent systems exist;
3. the exact order and rounding boundaries of Treasury aggregation relative to the already-implemented settlement timing;
4. what happens when available BC cannot satisfy required Maintenance, including the original deficit/scrap selection policy and any protected/excluded assets;
5. the minimum Core/Game/Session state and deterministic tests needed to add those categories without duplicating future Fleet/Ship/Leader state.

### Scope guard

Keep the next slice focused on Treasury categories and deficit/scrap behavior whose dependencies are already authoritative. Do not fabricate Ship/Leader/Spy/Fleet Maintenance records or broad economic systems solely to fill missing Treasury lines; evidence-dependent categories may stay deferred.

## Planned queue after Treasury completion

1. Continue the first headless vertical game loop toward colony-ship production, strategic movement/colonization and a second colony.
2. Extend strategic Fleet movement/location beyond the current at-system blockade presence when the movement/colonization slice begins.

## Explicitly parked / UI-only fidelity

- Full Fleet movement/path/ETA, special/non-player Fleet owners, tactical Fleet composition and broad diplomacy/treaty negotiation remain later strategic slices; the blockade slice intentionally implemented only the minimum producer state.
- Active conquest/occupation/automatic-assimilation progression, Android/Native population and persisted custom race designs remain later Population extensions.
- The original Hyper-Advanced research-selection screen temporarily previews counters at `completed + 1` and has a 20-level list boundary. MOOX deliberately keeps authoritative strategic cost/progression separate from that original UI quirk.
- Original copyrighted assets remain private reference material and are not distributable MOOX content.
- Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.
