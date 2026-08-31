# Open slice 05 - Command Points and ship Maintenance

Status: **Gates 1-3 complete; Gate 4 pending**.

Planned specification: `PLANNED_05_COMMAND_POINTS_SHIP_MAINTENANCE.md`.

Permanent evidence: `docs/research/COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md`.

## Gate checklist

- [x] Gate 1: repository check + original MOO2 Command-Point / ship Maintenance evidence.
- [x] Gate 2: implementation contract decision.
- [x] Gate 3: implementation + deterministic tests.
- [ ] Gate 4: final QA + commits + close.

## Gate-3 implementation checkpoint

- Core `StateSchemaVersion` is **20** with materialized `Empire.CommandPoints {Capacity, Used}` and Treasury `ShipCommandMaintenanceBC`.
- Economy ruleset schema is **8** with original-derived base/station/Communications/Warlord/Imperium/overage rules; military Ship CP remains derived from hull `SizeIndex + 1`.
- Colony/Outpost special Fleets consume 1 CP each; Population transfers and freighters do not consume CP.
- Star Base -> Battlestation -> Star Fortress replacement/buildability is normalized; impossible multiple-tier station state rejects settlement.
- Treasury precomputes and validates every Empire before the first mutation, then applies CP/Treasury snapshots deterministically; overage costs 10 BC per excess CP.
- Settlement remains pre-Construction. Newly completed Ships/station upgrades affect CP and Maintenance only on the following settlement.
- Split, merge and strategic movement preserve Ship CP usage; legal Outpost Ship deployment removes its 1 CP on the next derivation/settlement.
- Schema-20 save/load, Observer isolation and identical-session replay/event equality are covered.
- Full repository tests and Gate-3 focused tests have passed; Gate 4 owns a fresh final `go vet`, full QA, commits and closure.

## Explicit deferrals preserved

Leader/Officer CP and Maintenance, original troop Transport, NPC/difficulty overage rate, deficit liquidation, tactical station behavior and hostile battle-destruction timing remain outside Slice 05.

## Current resume point

Gates 1-3 are complete. Resume at **Gate 4** for fresh final QA, `go vet`, diff review, commits, HISTORY/status closure and OPEN-marker removal. Do not start Slice 06 before this slice is closed.
