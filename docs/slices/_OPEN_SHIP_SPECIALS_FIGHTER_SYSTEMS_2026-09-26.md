# OPEN Slice 17.4 - Ship Specials and Fighter Systems
Date opened: 2026-09-26
Status: **Gates 1-3 complete; Gate 4 review pending**
Plan: docs/slices/PLANNED_17_4_SHIP_SPECIALS_FIGHTER_SYSTEMS.md
## Scope
This child adds authoritative Ship Designer representation for supported specials and fighter/carrier systems after the accepted Beam and Ordnance design breadth.
Gate 1 must establish durable original/normalized evidence for:
- the complete candidate special/fighter catalog;
- technology, Space, Cost and miniaturization;
- stacking, exclusivity and quantity limits;
- strategic derived-stat effects that can be owned in 17.4;
- existing Tactical consumers versus missing combat subsystems;
- Fighter/Bomber/Interceptor carrier representation and exact Slice-17.8 dependencies;
- high-tech/Doom-Star reference fixtures and current MOOX gaps.
## Guardrails
- Slices 17.0-17.3 and follow-up 17.2A remain closed/accepted and are not reopened.
- Gate 1 is audit/evidence only; do not implement specials/fighter breadth before Gate 2 freezes the contract.
- Do not open Slice 17.5 or later children while 17.4 is open.
- Every candidate special must receive an explicit support classification before acceptance.
- Deferred fighter/combat-special Tactical consumers remain owned by Slice 17.8.
- Do not silently simulate or ignore unsupported Tactical effects.

## Gate 1 progress

Durable audit: docs/research/SLICE_17_4_GATE1_SPECIALS_FIGHTERS_AUDIT_2026-09-26.md

Completed: candidate inventory, HELP effect audit, MOOX gap audit, existing Tactical consumer classification, fighter/17.8 handoff, and Doom-Star reference vectors.

Gate 1 closed: original 1.31 economics/miniaturization, 8-slot/duplicate/exclusion legality, fighter quantity/economics, capacity ordering and explicit support classification are durable.

Next: Gate 2 freezes the supported matrix, compatibility/quote rules, snapshot representation and 17.8 handoffs.

## Gate 2 close

Contract: docs/research/SLICE_17_4_GATE2_SPECIALS_FIGHTERS_CONTRACT_FREEZE_2026-09-27.md

Frozen: supported matrix, support classes, 8-slot/exclusion/prerequisite rules, Space/PP/miniaturization, Special snapshot shape, Megafluxers/Battle Pods/Extended Fuel Tanks effects, Fighter quantity semantics and exact 17.8 handoffs.

Next: Gate 3 implementation in the frozen order; no scope broadening.

## Gate 3 implementation

Implementation evidence: docs/research/SLICE_17_4_GATE3_IMPLEMENTATION_2026-09-28.md

Completed: authoritative Special rules/catalog/quote/save snapshots, Megafluxers/Battle Pods/Extended Fuel Tanks frozen strategic effects, Fighter/Bomber/Heavy Fighter bay rows, immutable construction persistence, explicit Tactical metadata handoff with inertial turning consumers, and responsive 8-slot Ship Designer HMI.

Verification: full Go tests, go vet, production web build, Reference-Lab browser smoke including 390 px layout, direct All-Tech Special/Fighter review, and diff checks all pass.

Next: Gate 4 review/acceptance. Deferred Special/Fighter tactical consumers remain Slice 17.8 scope.
