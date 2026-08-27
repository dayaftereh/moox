# Building costs, maintenance and first construction slice - 2026-08-27

Status: implemented and verified against the private Master of Orion II 1.31 reference executable.

## Original building table

The existing clean-room building decoder reads the original `_buildings` table from `Orion2.exe` 1.31:

- LE object: 2
- object-relative offset: `0x6B3D`
- record count: 49 (dummy record 0 plus building IDs 1..48)
- record size: `0x13` bytes
- original executable SHA-256: `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`

The fields used by this checkpoint are:

```text
+0x04 uint16  building ID
+0x06 uint16  technology ID
+0x08 uint16  production cost in PP
+0x0C uint16  maintenance in BC per turn
```

The decoder validates the original building IDs and technology links and now normalizes the two additional economy fields with per-field provenance offsets.

Representative observed values include:

| Building | PP cost | BC/turn |
| --- | ---: | ---: |
| Alien Management Center | 60 | 1 |
| Automated Factory | 60 | 1 |
| Holo Simulator | 120 | 1 |
| Hydroponic Farms | 60 | 2 |
| Marine Barracks | 60 | 1 |
| Supercomputer | 150 | 2 |
| Pleasure Dome | 250 | 3 |
| Research Lab | 60 | 1 |
| Robo Miner Plant | 150 | 2 |
| Spaceport | 80 | 1 |
| Star Base | 400 | 2 |
| Star Fortress | 2500 | 4 |
| Subterranean Farms | 150 | 4 |
| Weather Control System | 200 | 3 |

`data/rulesets/moo2-1.31/buildings.json` is regenerated from the private reference and contains these fields for all 48 normalized colony buildings.

## Runtime construction semantics in this checkpoint

The first deterministic construction slice intentionally remains narrow:

- a colony may have one active `ConstructionState`;
- `colony.queue_building` is an authoritative strategic command;
- the server maps `SeatID -> EmpireID` and rejects construction on foreign colonies;
- unknown buildings, already-owned buildings and a second active construction are rejected;
- the active project receives that resolution's domain-native `PopulationDynamics.ProductionAvailable` in PP, after any Cybernetic Production sustenance;
- applied production is capped at the remaining building cost;
- zero production does not emit a progress event;
- completion appends the building to the colony and clears the active construction;
- a newly completed building does not retroactively change the PP already consumed for that turn; after turn-end Population Growth, the post-turn snapshot is recalculated and may already show the completed building's next-state effect.

Construction now uses domain-native `float64` PP. Building definitions expose `production_cost_pp` directly, `ConstructionState.ProgressPP` preserves fractional production, and completion uses a small deterministic tolerance for floating-point residue.

## Events

The resolver emits:

- `colony.construction_queued` - attributed to the submitting seat and command;
- `colony.construction_progressed` - system event containing applied/progress/remaining PP as numeric `float64` values;
- `colony.building_completed` - system event when the exact cost is reached.

These events flow through the existing deterministic `GameSession` observer/replay stream.

## Technology buildability boundary

The normalized building table already carries the original `technology_id` for every building. The runtime now uses that proven link as the minimum buildability prerequisite:

- `Empire.KnownTechnologyIDs` is persistent deterministic state;
- the list is validated as unique, strictly ascending original technology IDs in the normalized 1..203 range;
- `colony.queue_building` rejects a building whose technology is not known by the owning empire;
- `AvailableBuildingChoices` projects only buildings whose technology is known, which are not already present on the colony, and only while the colony has no active construction project;
- choices are returned in stable original production-ID order;
- `GameSession.BuildingChoices` applies the authoritative seat-to-empire mapping before returning this projection, so human clients and AI agents can consume the same legal-action list.

This checkpoint deliberately does **not** assign starting technologies to the deterministic fixture and does not claim the original new-game technology grant. Tests inject explicit known technology IDs where a buildability scenario requires them. Original starting/research acquisition behavior remains separate evidence work.
## Deliberately not claimed yet

This checkpoint does not establish or implement:

- production overflow/carry to a following queue item;
- a multi-item construction queue;
- buyout/rush-production rules;
- maintenance deductions from empire treasury;
- original starting technology grants and research acquisition timing;
- replacement/exclusion relationships between buildings;
- ship construction;
- pollution interaction with construction output.

Although maintenance is now normalized from the original table, it remains data-only until the empire finance/maintenance layer is implemented.
