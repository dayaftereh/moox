# Slice 16.7 Gate 2 - Advanced deferral and downstream handoff

Date: 2026-09-18

Status: **Gate 2 complete by explicit deferral/handoff decision; Slice 16.7 ready to close without Advanced runtime enablement**.

## Decision

Full Advanced Starting Civilization implementation is **not** forced into Slice 16.

The current New Game/runtime architecture is structurally capable of carrying Advanced, but complete original-style parity would pull substantial unfinished strategic mechanics into a New Game milestone. Those mechanics now have explicit downstream owners:

- Slice 17 - ship/component/weapon breadth relevant to any final Advanced fleet contract;
- Slice 20 - Espionage/intelligence and spying-dependent race breadth;
- Slice 21 - Colony / Buildings / Planet Effects / Pollution fidelity;
- Slice 22 - Preset Race Completion + Custom Race Designer;
- Slice 23 - Full Advanced Starting Civilization parity.

The downstream intended order is: **17 -> 20 -> 21 -> 22 -> 23**. Slices 18/19 remain intentionally unassigned.

## Gate-2 feasibility result

### Architecture is not the blocker

Gate-2 research confirmed that current authoritative state/runtime can already carry the structural shape needed by Advanced: arbitrary multi-Colony Empires, variable Population, per-Colony Economy/Food logistics, Treasury, Research, Command Points, built-in AI, save/export/import/restore and deterministic New Game RNG state.

Advanced preference profiles can also be generated deterministically as transient New Game bootstrap inputs before technology initialization. They do not require a new persistence schema merely to run the start chooser, provided their materialized technology result and final RNG state are authoritative.

### Advanced technology foundation is already strong

The existing Advanced chooser implements and tests Average + exactly 19 additional grants, exactly 26 completed starting fields, stable Empire ordering, shared New Game RNG, race weighting, Creative / Uncreative behavior, government/application eligibility, Tactical/Strategic filtering and cross-Empire technology competition. That work is preserved for Slice 23.

### Original territory breadth is representable

Gate 1 proved original total starting Colonies per Empire: **ceil(floor(NUM_STARS / 2) / NUM_PLAYERS)**.

Current Core/Persistence can materialize those Colony counts, and Gate-2 probes showed generated galaxies contain enough planets overall for the required counts.

### Naive territory assignment is not gameplay-safe

A temporary research-only feasibility matrix across galaxy size/age, 2/3 players and multiple seeds demonstrated that simply taking current generated planets at original Advanced Colony counts is not sufficient.

With original Advanced Colony counts, zero starting Freighters, no original Advanced planet balancing and no original starting Building effects, many starts do not contain enough directly self-sufficient farmable planets. This is direct evidence that the original Advanced planet-selection/twiddle path is mechanically important rather than cosmetic.

### Planet twiddle is partly representable now, but full fidelity belongs downstream

Fresh disassembly confirmed the original basic Advanced planet twiddle uses fields mapping cleanly onto existing MOOX concepts: size, climate and mineral class. Current climates can be moved monotonically into food-producing states without a Core schema redesign.

However, complete original Advanced planet-worth/balancing also reaches richer planet/star-special behavior not yet completely represented by the current authoritative planet model. That belongs to Slice 21 rather than an Advanced-only approximation.

### Starting Buildings are the major current fidelity blocker

All 48 normalized Buildings exist in rules/data and can broadly be owned, costed, technology-gated and maintained.

But Gate-2 capability mapping found the running game explicitly models the promised gameplay effect for only a subset of the original Advanced starting-Building candidate set. The original Advanced candidate list contains 42 product IDs; roughly 9 currently have sufficiently explicit gameplay consumers, while roughly 33 would risk becoming ownership/buildability/maintenance records without their promised mechanical effect.

Granting those Buildings during Advanced today would therefore create partially inert or misleading starting infrastructure.

### Pollution is part of the dependency

Repository evidence already normalizes Pollution Processor, Atmospheric Renewer, Core Waste Dumps, Recyclotron and Tolerant/Pollution-research interactions, while permanent Economy research records Pollution as a contextual production layer not yet fully active in runtime.

Because original Advanced starting infrastructure can include production/Pollution-related Buildings, Advanced should not grant them before their actual effects exist. Slice 21 therefore owns gross production, Pollution-relevant production, Pollution control, Pollution loss, final effective production, Pollution-exempt production and Tolerant interaction.

### Fleet contract remains intentionally unresolved

Original HELP says an Advanced start uses all available Command Points. The owned Orion2.exe 1.31 runtime path calls Init_Advanced_Civilization_Fleet_, but that verified entry point is a single RET. The ordinary two-Scout + Colony Ship start is gated to Average, not Advanced.

Gate 2 does not guess which interpretation should become the final MOOX product contract. Slice 23 must re-decide explicitly between executable 1.31 parity and documented HELP/product-intent parity.

### Difficulty secondary layer remains evidence-limited

Original global 0x21CB0 participates in an additional Advanced chooser filter, but no proven mapping to the current five MOOX Difficulty IDs exists yet. Slice 16 does not invent that mapping. Slice 23 must re-audit it against the then-current AI/Difficulty system.

## Frozen Slice-16 contract for Advanced

For Slice-16 closure:

- stable technology ID remains advanced;
- server catalog continues to expose Advanced as **planned**;
- browser continues to show the existing Advanced visual card in planned/disabled form;
- public New Game submission with Advanced remains rejected;
- no silent fallback to Average is allowed;
- no partial Advanced Colony/Building/Fleet bootstrap is introduced;
- Pre-Warp and Average remain the supported starting-technology modes;
- all permanent Gate-1/Gate-2 research is preserved as mandatory Slice-23 input.

## Gate-2 checklist resolution

- [x] Freeze exact current server-owned Advanced identity/label behavior: stable ID, catalog-visible, planned/locked.
- [x] Freeze bootstrap behavior for Slice 16: **no partial Advanced bootstrap**; complete semantics move to Slice 23.
- [x] Freeze race/cross-setting behavior for Slice 16: Advanced remains rejected for every current race/difficulty/galaxy/player combination.
- [x] Freeze Advanced visual-card facts: show only server-derived planned/locked facts; do not imply implementation.
- [x] Freeze acceptance required before future enablement: Slice 23 must satisfy its complete deterministic cross-setting matrix before catalog/API/UI promotion.

## Gates 3 and 4 disposition

The original Slice-16.7 Gate-3 implementation and Advanced-specific Gate-4 acceptance work are **not executed and are not marked complete**.

They are superseded by the explicit downstream product decision and moved into docs/slices/PLANNED_23_FULL_ADVANCED_STARTING_CIVILIZATION_PARITY.md. This prevents documentation from falsely claiming Advanced runtime parity.

Slice-16 closure instead verifies the current supported New Game contract from Slices 16.1-16.6 remains intact, Advanced remains visibly planned and backend-rejected, no Advanced fallback exists, all Advanced research/decisions have a durable downstream owner and no open Slice-16 marker remains.

## Permanent evidence handoff to Slice 23

Slice 23 inherits at minimum:

- docs/research/SLICE_16_7_GATE1_ADVANCED_PARITY_AUDIT_2026-09-18.md;
- this Gate-2 deferral/handoff evidence;
- docs/research/TECHNOLOGY_START_RESEARCH_2026-08-27.md;
- docs/research/ADVANCED_START_RESEARCH_2026-08-28.md;
- docs/research/ADVANCED_START_CHOOSER_2026-08-28.md;
- docs/research/NEW_GAME_GALAXY_GENERATION_BASELINE_2026-09-02.md;
- docs/research/COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md;
- Slice-16.5/16.6 New Game evidence.

## Gate-2 conclusion

Advanced is feasible in the existing architecture, but implementing it now would either widen Slice 16 into Building/Planet/Pollution/Race/Ship mechanics or knowingly ship approximated/inert mechanics. Neither is acceptable.

The correct closure is therefore to keep Advanced honestly planned/locked, preserve the research, finish Slice 16 at its supported Pre-Warp/Average breadth, and implement full Advanced parity later in Slice 23 on top of the normal gameplay systems it actually requires.
