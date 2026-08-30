# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: **none**.
- Latest completed gameplay slice: **Treasury Maintenance categories and deficit/scrap policy**.
- Implementation commit: `a1ad15b` (`game: align treasury freighter and food rounding`).
- Core `StateSchemaVersion`: **15**.
- Economy ruleset schema: **7**.
- Permanent evidence: `docs/research/TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; resume any marker before opening another slice.

## Closed Treasury checkpoint

Direct MOO2 1.31 executable evidence resolves the six original Maintenance buckets as:

1. Colony/Building Maintenance;
2. active Freighter operating cost;
3. Ship command-point overage Maintenance;
4. Spy Maintenance;
5. outgoing Tribute payments;
6. Officer/Leader Maintenance.

The original forced-deficit routine begins only when `Treasury + current net BC < 0` and uses deterministic staged liquidation across Buildings, Spies, Ships and Leaders. Freighters are not a direct forced-liquidation class.

The dependency-safe runtime corrections are complete without schema changes:

- Population-transfer reservations and Food-transfer use share one active-Freighter Maintenance bucket;
- active-Freighter cost uses the original aggregate whole-BC boundary (`floor(active * 0.5)` in the MOO2 1.31 ruleset);
- `SurplusFoodSold` remains continuous while `SurplusFoodIncomeBC` uses the original whole-BC conversion boundary;
- negative `BalanceBC` remains an explicit modernization until canonical Ship/Spy/Leader/treaty state can support the full original liquidation policy.

Gate 4 passed `gofmt`, `go test ./...`, `go vet ./...` and `git diff --check`.

## Next queued runtime slice

Continue the first headless vertical game loop toward **colony-ship production, strategic Fleet movement/colonization and a second colony**.

That slice should begin with a fresh Gate 1 and should resolve only the minimum original semantics needed for:

- producing a Colony Ship through the existing semantic Construction project model;
- representing a mobile strategic Fleet without collapsing movement/path/ETA into the current `AtSystemID` blockade-only state;
- arriving at a target system and colonizing a legal unowned planet;
- creating the second Colony through authoritative Core/Game/Session transitions;
- preserving deterministic save/load and Observer/replay behavior.

Do not absorb tactical ship design/combat, broad diplomacy, Outpost behavior or AI unless direct dependency evidence requires them.

## Later deferred dependencies

- Ship command-point capacity/usage and overage Maintenance;
- Spy state/Maintenance;
- trade/research/Tribute treaty economics;
- Officer/Leader state, Maintenance and economic bonuses;
- complete original staged Treasury-deficit liquidation;
- active conquest/occupation/automatic-assimilation progression;
- Android/Native population and persisted custom race designs;
- special/non-player Fleet owners and tactical Fleet composition;
- original Hyper-Advanced selection-screen preview quirks as optional UI fidelity.

Original copyrighted assets remain private reference material and are not distributable MOOX content. Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.