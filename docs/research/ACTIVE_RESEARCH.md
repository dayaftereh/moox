# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: `docs/slices/_OPEN_FLEET_DIPLOMACY_BLOCKADES_2026-08-29.md`.
- Active slice: **Canonical strategic Fleet/Diplomacy blockade production**.
- Latest completed gameplay slice: Race-aware Population cohorts (`7c49e90`).
- Core `StateSchemaVersion`: **15**.
- Economy ruleset schema: **7**.
- Current slice evidence: `docs/research/FLEET_DIPLOMACY_BLOCKADES_2026-08-29.md`.
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
- cohort-aware Population transfer preserved through GameSession, Observer and replay.

## Current objective

### Canonical strategic Fleet/Diplomacy state for blockade production

Current gate: **Gate 4 - QA + commit + close pending**. Gates 1-3 are complete. Gate 2 was accepted on 2026-08-29 without scope changes; Gate 3 implemented Core schema 15 strategic Fleet/directed relation state plus deterministic blockade production and original-order integration before Population-transfer arrivals/final Food. `docs/research/FLEET_DIPLOMACY_BLOCKADES_2026-08-29.md` is the permanent evidence and implementation record.

Gate 1 now establishes:

- target system presence is derived from **Colony ownership**, not defending Fleet presence;
- only stationary (`fleet +0x64 == 0`) ordinary/combat (`fleet +0x11 == 0`) strategic Fleets participate in normal blockade production;
- Fleet owner and system are the direct identities used by the original producer;
- normal blockade hostility is a **directed fleet-owner -> target-owner** relationship predicate, with original relation values `4..6` accepted;
- several hostile Fleet owners combine idempotently into the target blockade mask;
- the original also materializes per-target blockader masks, but current MOOX consumers require only the already-persisted `StarSystem.BlockadedEmpireIDs`;
- original timing is first Colony/Food pass -> `Compute_Blockades_` -> Settler movement -> final Colony/Food pass, mapping cleanly to a MOOX recomputation immediately before `advancePopulationTransfers`;
- special/non-player fleet owners, full movement-state enums, full diplomacy state names and tactical fleet composition remain deliberately deferred.

### Proposed Gate 2 shape

Advance Core schema **14 -> 15** with minimal semantic strategic state:

- `StrategicFleet { ID, EmpireID, Role, AtSystemID }`, where only `role=Combat` and a non-zero `AtSystemID` can blockade;
- directed `DiplomaticRelation { FromEmpireID, ToEmpireID, Stance }`, initially semantic `neutral|hostile` rather than guessed names for legacy relation values;
- no duplicate System-presence state: Colony owners are derived from existing Planet/Colony state;
- pure deterministic `recomputeSystemBlockades` clears/rebuilds sorted/unique `StarSystem.BlockadedEmpireIDs` from Fleet + relation + Colony state;
- run that phase after Construction and before Population-transfer resolution/final Food recalculation;
- no new player command and no tactical/broad-diplomacy scope.

### Scope guard

Keep this slice focused on canonical strategic Fleet/Diplomacy state **only as far as required to produce blockade state**. Do not absorb tactical combat, ship-design fidelity, fleet strength, full movement/ETA, diplomacy negotiations/treaties, special/non-player fleets, AI strategy or UI work.

## Planned queue after blockade production

1. Missing original Treasury income/Maintenance categories and deficit/scrap policy once their dependent systems exist.
2. Continue the first headless vertical game loop toward colony-ship production, strategic movement/colonization and a second colony.

## Explicitly parked / UI-only fidelity

- Active conquest/occupation/automatic-assimilation progression, Android/Native population and persisted custom race designs remain later Population extensions; the cohort slice intentionally did not absorb them.
- The original Hyper-Advanced research-selection screen temporarily previews counters at `completed + 1` and has a 20-level list boundary. MOOX deliberately keeps authoritative strategic cost/progression separate from that original UI quirk.
- Original copyrighted assets remain private reference material and are not distributable MOOX content.
- Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.
