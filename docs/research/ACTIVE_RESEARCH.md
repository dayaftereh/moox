# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: `docs/slices/_OPEN_TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md`.
- Active slice: **Treasury Maintenance categories and deficit/scrap policy**.
- Current gate: **Gate 4 - QA and close in progress**.
- Latest completed gameplay slice: Canonical strategic Fleet/Diplomacy blockade production (`d148502`).
- Core `StateSchemaVersion`: **15**.
- Economy ruleset schema: **7**.
- Current permanent evidence: `docs/research/TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md`.
- Previous Treasury evidence: `docs/research/TREASURY_SETTLEMENT_2026-08-28.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; resume any marker before opening another slice.

## Recently closed runtime sequence

The current Research/Economy chain is implemented and recorded in `docs/slices/HISTORY.md`, including race-aware research, original strategic turn ordering, Freighter/Food logistics, modeled Treasury settlement, Population relocation/growth/capacity/cohorts, and Core schema 15 strategic Fleet/Diplomacy blockade production.

## Current objective

### Treasury Maintenance categories and deficit/scrap policy

Gate 1 is complete and the proposed narrow Gate 2 contract was accepted on 2026-08-30. Gate 3 is implementing only the two dependency-safe fidelity corrections.

Gate 1 resolves the original six Maintenance component words at player `+0xB8..+0xC2` as:

1. Colony/Building Maintenance;
2. active Freighter operating cost;
3. Ship command-point overage Maintenance;
4. Spy Maintenance;
5. outgoing Tribute payments;
6. Officer/Leader Maintenance.

The original gross/current BC path additionally includes Colony BC production, Leader economic bonuses, treaty effects, received Tribute and positive surplus-Food sale income.

The original forced deficit routine `Player_Maintenance_` begins only when `Treasury + current net BC < 0` and then uses staged, deterministic, asset-aware liquidation across Buildings, Spies, Ships and Leaders. It does not directly auto-scrap Freighters. Because MOOX lacks canonical Ship/Spy/Leader and treaty-economic state, implementing only the currently available asset classes would be observably non-original and remains deferred.

### Proposed Gate 2 shape

Keep Core schema **15** and Economy ruleset schema **7**. Make only the two dependency-safe fidelity corrections now:

- active Freighter Maintenance uses existing `PopulationTransportFreightersReserved + FreightersUsed` and materializes whole-BC cost as `floor(activeFreighters / 2)`; this closes the prior Settler-reservation Maintenance uncertainty;
- surplus-Food sale retains continuous `SurplusFoodSold` telemetry but materializes original whole-BC income as `floor(SurplusFoodSold * rate)`, with normal rate 0.5 and Fantastic Traders rate 1.0.

Keep current Building Maintenance and Treasury settlement order unchanged. Do not add zero-valued fake Ship/Spy/Tribute/Officer categories and do not implement partial automatic deficit liquidation. Negative `BalanceBC` remains an explicit modernization until the source asset systems required by original `Player_Maintenance_` are authoritative.

### Scope guard

Do not absorb Ship command-point state, Spies, Leaders, treaty economics, ship design/combat or broad diplomacy merely to make Treasury totals look complete. These dependencies should be implemented in their own evidence-driven slices and then wired into the already-resolved original Treasury buckets.

## Planned queue after this Treasury slice

1. Continue the first headless vertical game loop toward colony-ship production, strategic movement/colonization and a second colony.
2. Extend strategic Fleet movement/location beyond current at-system blockade presence when the movement/colonization slice begins.
3. Revisit complete deficit/scrap liquidation only after canonical Ship/Spy/Leader/treaty state exists.

## Explicitly parked / UI-only fidelity

- Full Fleet movement/path/ETA, special/non-player Fleet owners, tactical Fleet composition and broad diplomacy/treaty negotiation remain later strategic slices.
- Active conquest/occupation/automatic-assimilation progression, Android/Native population and persisted custom race designs remain later Population extensions.
- The original Hyper-Advanced research-selection screen preview/list-boundary quirks remain optional UI fidelity, not strategic-core rules.
- Original copyrighted assets remain private reference material and are not distributable MOOX content.
- Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.