# Completed work slices

This ledger is the compact audit trail for **closed** Master of Orion X implementation slices. It complements the permanent evidence documents under `docs/research/` and prevents completed work from remaining accidentally listed as pending in live roadmaps.

A row belongs here only after implementation/documentation/QA has been closed. An unfinished slice is represented instead by `docs/slices/_OPEN_*.md`.

## Research and Economy sequence

| Slice | Result | Closing commit | Permanent evidence / checkpoint |
| --- | --- | --- | --- |
| Race-aware multi-Technology research | General/ordinary/Creative/Uncreative application policies and shared legal actions implemented | `f36ab5a` | `docs/research/RESEARCH_MULTI_TECH_2026-08-27.md` |
| Research project switching | Active project can change with complete accumulated RP transfer and Observer/replay history | `5d2f983` | `docs/research/RESEARCH_SWITCHING_2026-08-27.md` |
| Strategic turn order | Direct executable evidence fixed Research -> Growth/Starvation -> Construction with pre-growth RP/PP snapshots | `30e2b6d` | `docs/research/TURN_ORDER_2026-08-28.md` |
| Uncreative initialization | Fixed applications generated during New Game from the shared New Game RNG | `945f777` | `docs/research/UNCREATIVE_INITIAL_SELECTION_2026-08-28.md` |
| Hyper-Advanced repeated research | Per-Empire repeat levels, dynamic strategic cost and repeat completion implemented; original UI preview off-by-one documented but not copied into the core | `212632a` | `docs/research/HYPER_ADVANCED_RESEARCH_2026-08-28.md` |
| Advanced-start technology ownership | Original-style randomized/race-aware Advanced start implemented with shared New Game RNG and exactly 19 weighted extra grants | `f80b9bc` | `docs/research/ADVANCED_START_RESEARCH_2026-08-28.md`, `docs/research/ADVANCED_START_CHOOSER_2026-08-28.md` |
| External Technology grants / Uncreative repair | Server-authoritative grant transition implemented; incomplete TechFields are preserved and Uncreative fixed choices are repaired through authoritative RNG | `6258ac5` | `docs/research/UNCREATIVE_EXTERNAL_REPAIR_2026-08-28.md` |
| Construction project generalization | Construction state generalized from Building-only identity to semantic project kind/id as prerequisite for non-building projects | `3724275` | architecture/runtime checkpoint in current status docs |
| Freighter Fleet construction | 50 PP project adds 5 Empire Freighters through authoritative Construction legal actions/events | `9ae1ef8` | `docs/research/FOOD_FREIGHTER_LOGISTICS_2026-08-27.md` |
| Treasury settlement | 50 BC New Game Treasury and modeled strategic settlement implemented before Research | `2861150` | `docs/research/TREASURY_SETTLEMENT_2026-08-28.md` |
| Insufficient-Freighter priority | Temporary proportional allocation replaced by original-style round-robin deficit allocation | `b972ee0` | `docs/research/INSUFFICIENT_FREIGHTER_PRIORITY_2026-08-28.md` |
| System blockade -> Food logistics | Authoritative blockaded-Empire system state excludes Colonies from Food import/export/sale pools | `0c699a8` | `docs/research/BLOCKADE_FOOD_LOGISTICS_2026-08-28.md` |
| Population relocation / shared Freighters | Same-system immediate moves and interstellar Settler reservations/ETA/arrival/loss handling implemented; Core schema advanced to 11 | `f816eaf` | `docs/research/POPULATION_TRANSPORT_FREIGHTER_2026-08-29.md` |
| Population growth building / medicine modifiers | Housing continuous Production mode, Microbiotics +25pp, Universal Antidote +50pp and flat Cloning Center +0.1 Population/turn implemented; Core schema advanced to 12 | `e7f0d72` | `docs/research/POPULATION_GROWTH_MODIFIERS_2026-08-29.md` |
| Population capacity transitions | Advanced City Planning +5, Biospheres +2, semantic Terraforming/Gaia climate transformations and immediate capacity-clamp invariant implemented; Core schema advanced to 13 | `25d05a9` | `docs/research/POPULATION_CAPACITY_TRANSITIONS_2026-08-29.md` |
| Race-aware Population cohorts | Core schema 14 organic cohorts, source-race Economy, exact four-pass Food priority, per-origin Growth/Starvation, heterogeneous capacity and cohort-aware Population transfer implemented | `7c49e90` | `docs/research/POPULATION_COHORTS_2026-08-29.md` |
| Strategic Fleet/Diplomacy blockade production | Core schema 15 minimal strategic Fleets/directed relations and deterministic Colony-presence + hostile-combat-Fleet system blockade production implemented | `d148502` | `docs/research/FLEET_DIPLOMACY_BLOCKADES_2026-08-29.md` |
| Treasury Maintenance categories / deficit boundary | Six original Maintenance buckets and staged deficit policy resolved; active-Freighter reservations/usage and surplus-Food whole-BC boundaries implemented without schema changes | `a1ad15b` | `docs/research/TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md` |
| Colony Ship production / strategic movement / colonization | Core schema 16 Tech-41 Colony Ship Construction, original Fuel ranges, deterministic pre-Construction transit and explicit second-Colony creation/ship consumption implemented | `db9d01e` | `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md` |

## Meta workflow

| Slice | Result | Closing commit | Permanent documentation |
| --- | --- | --- | --- |
| Recovery-safe slice protocol | Four-gate workflow and `_OPEN_*.md` recovery marker convention established | `204ced7` | `docs/slices/README.md` |

## Current audit state - 2026-08-30

- `docs/slices/` contains **no** `_OPEN_*.md` marker.
- The completed rows above have committed runtime/tests and/or explicit closing documentation in Git history.
- The Colony Ship build/move/colonize slice is closed in `db9d01e`; no gameplay slice is currently open. Six `PLANNED_01...06` specifications now prepare the next roadmap, but none counts as open work; the next started objective must still begin with a fresh Gate 1 and exactly one `_OPEN_` marker.
- Live current status is authoritative in `README.md`, `docs/PROJECT_STATUS.md`, `docs/architecture/README.md`, and `docs/research/ACTIVE_RESEARCH.md`.
- Historical research documents are evidence checkpoints and are not rewritten merely because a later slice resolved one of their then-deferred items.