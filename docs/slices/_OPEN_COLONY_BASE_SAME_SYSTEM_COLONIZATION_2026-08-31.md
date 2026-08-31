# OPEN - Colony Base and same-system colonization - 2026-08-31

Status: **Gate 3 complete; Gate 4 pending**.

Planned specification: `docs/slices/PLANNED_01_COLONY_BASE_SAME_SYSTEM_COLONIZATION.md`

Permanent evidence: `docs/research/COLONY_BASE_SAME_SYSTEM_COLONIZATION_2026-08-31.md`

This is the single recovery marker for the active MOOX gameplay slice.

## Gate 1 - Checkup + original analysis / reverse engineering

- [x] Re-check Git status, `_OPEN_*.md`, Core schema and current Building/Construction handling.
- [x] Verify normalized Technology 40 / Building 11 / 200 PP / 0 BC provenance.
- [x] Verify Colony Base has a dedicated original buildability rule requiring an empty colonizable Planet in the source Colony's StarSystem.
- [x] Verify completed Colony Base is represented through Building-ownership slot 11 (`colony+0x136+11 == colony+0x141`).
- [x] Verify original `Colonization_` treats Colony Base as a colonizer source alongside Colony Ship and Outpost Ship.
- [x] Verify same-system target selection happens during the Colonization/report flow, not as part of the Construction queue/buildability check.
- [x] Verify successful Colony Base colonization creates a normal Colony through `Make_New_Colony_` and consumes the source Colony Base through `Remove_Building_(11)`.
- [x] Close source-Population and exact completion/report timing questions at the fidelity level needed for Gate 2.
- [x] Remove all temporary/private executable extracts and helper scripts.
- [x] Update permanent evidence and `ACTIVE_RESEARCH.md` with Gate-1-complete findings.
- [x] Run `git diff --check` and verify only intended Gate-1 docs/marker remain.
- [x] Present Gate-1 findings plus proposed Gate-2 contract; stop before gameplay implementation.

## Gate 2 - Accepted implementation contract

- [x] Gate-2 contract in `docs/research/COLONY_BASE_SAME_SYSTEM_COLONIZATION_2026-08-31.md` accepted on 2026-08-31.
- [x] Keep Colony Base as normal Building 11 / `colony_base` Construction.
- [x] Require a same-system empty Planet for queue legality; do not persist a queue-time target.
- [x] Derive completed Colony Base decisions from Building ownership without a Core schema bump.
- [x] Resolve completed Bases after strategic Production by colonizing or trashing for 100 BC.
- [x] Preserve source Population during Colony Base founding.

## Gate 3 - Implementation

- [x] Specialize `AvailableBuildingChoices` and direct queue validation for same-system Colony Base legality.
- [x] Add deterministic `PendingColonyBaseResolutions` projection with legal same-system target Planet IDs and 100-BC trash value.
- [x] Add authoritative `colony.colonize_with_base` and `colony.trash_colony_base` post-resolution commands.
- [x] Reuse one shared normal-Colony founding helper for Colony Ship and Colony Base.
- [x] Consume Building `colony_base` on successful founding; leave source Population unchanged.
- [x] Remove Building `colony_base` and credit exactly 100 BC on trash resolution, including targetless completed Bases.
- [x] Block `CompleteTurn` while any completed Colony Base remains unresolved.
- [x] Project pending resolutions into PlayerView, ObserverView and seat-scoped legal resolution surface.
- [x] Add dedicated `colony.colony_base_colonized` / `colony.colony_base_trashed` events while retaining generic `empire.planet_colonized` on founding.
- [x] Preserve Core `StateSchemaVersion` 16; verify pending resolution survives exact save/load round-trip.
- [x] Add deterministic replay, Observer clone-isolation, authority, event-order and source-Population regression tests.
- [x] Focused Colony Base, Colony Ship, Construction, Treasury and session regressions pass.
- [x] `go test ./... -count=1` passes.
- [x] `go vet ./...` passes.

## Gate 4 - QA / commit / close

Pending. Gate 4 must perform final format/diff QA, update HISTORY/closed handoff, create the implementation and closing-doc commits, and delete this `_OPEN_` marker.
