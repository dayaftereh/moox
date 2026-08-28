# Multi-Technology research selection - 2026-08-27

Status: implemented server-authoritative Ordinary / Creative / Uncreative application selection for concrete Technology Fields.

This checkpoint fixes the previous temporary behavior where completing any researched TechField granted every Technology attached to that field. The runtime now models the distinction between general fields, ordinary one-application research, Creative all-applications research and Uncreative fixed-one research.

## Evidence

### Original manual / manual transcription

The original manual's Research chapter states that:

- a field contains one to four possible applications;
- a few basic/general fields research the entire field;
- later fields allow only one technological application to be researched;
- Creative races are the exception and can research all applications;
- the player selects an application (or a field) when directing research.

The Race Design section separately states that Uncreative researchers recognize only one technology associated with a field, while Creative races can discover all applications appropriate to the field.

These points are strong evidence for the application policy and for ordinary application selection being part of starting the research project rather than an arbitrary client-side ownership decision at completion.

### StrategyWiki cross-check

StrategyWiki independently describes:

```text
ordinary / non-Creative -> choose 1 technology from a level
Creative                -> receive all technologies in that level
Uncreative              -> no choice; game software randomly selects 1
General levels          -> every empire receives all technologies
```

StrategyWiki also notes a small set of general levels where the Creative distinction does not apply.

## MOOX selection modes

The authoritative protocol/state now uses:

```text
all
choose_one
fixed_one
```

as `core.ResearchSelectionMode`.

Semantics:

```text
General/basic field                 -> all
Creative race, non-General field    -> all
ordinary race, non-General field    -> choose_one
Uncreative race, non-General field  -> fixed_one
```

`ResearchChoice.SelectionMode` exposes this directly to Human UI, built-in AI and future remote/MCP agents.

## General-field mapping

The normalized technology start data contains the six deterministic staged starting fields:

```text
29, 55, 22, 57, 28, 23
```

plus always-known field `0`.

These staged fields contain the baseline applications needed for the deterministic Pre-Warp/Average transition, including sets such as Freighters/Nuclear Drive and Colony/Outpost/Transport. Their shape aligns with the documented general-level behavior and the secondary estimate that roughly six levels are exceptions to normal Creative/non-Creative choice.

MOOX therefore currently uses the normalized staged-start field set plus field `0` as `GeneralResearchFieldIDs`.

This mapping is a strong normalization inference from the already-observed start-field topology plus independent general-level documentation. If later original-executable evidence yields an explicit general-field flag/table, that evidence should replace the derivation without changing the public ResearchChoice API.

## Ordinary research

For a normal race, a non-General TechField is projected as:

```text
selection_mode = choose_one
technology_ids = all legal applications visible in this field
```

The command is now:

```text
empire.select_research {
    tech_field_id,
    technology_id
}
```

For `choose_one`, `technology_id` is mandatory and must be one of the server-projected applications.

Example TechField 4:

```text
13 anti_missile_rockets
56 reinforced_hull
66 fighter_bays
```

A Human selection of `reinforced_hull` stores only:

```text
ResearchState.TechnologyIDs = [56]
```

and completion grants only Technology 56. The two unselected applications remain available only through future external acquisition systems such as trade, espionage or conquest.

## Creative research

For a Creative race, a non-General field is:

```text
selection_mode = all
```

The client does not select a Technology ID. The authoritative server stores all currently researchable applications of that field in the active `ResearchState`, and completion grants the complete set for the same field RP cost.

A client attempting to send an individual `technology_id` for `all` mode is rejected.

## Uncreative research

For an Uncreative race, a non-General field is:

```text
selection_mode = fixed_one
```

The client is shown exactly one application and is not permitted to send a Technology ID. The active `ResearchState` contains exactly that server-fixed application.

### Persisted deterministic plan

Research-choice projection must be read-only. Calling `ResearchChoices()` from UI, AI or networking must never consume authoritative RNG or change which technology an Uncreative empire receives.

MOOX persists:

```text
Empire.UncreativeResearchChoices[]
    TechFieldID
    TechnologyID
```

Direct original-executable analysis now proves that this full fixed plan is generated during `Init_Player_Tech_` inside New Game player initialization. MOOX therefore uses the caller-owned shared `NewGameRNG` rather than a separate Uncreative-only seed.

The verified initial range is exactly TechField `1..73`. General fields remain `all`; special Antaran TechField 74 and Hyper-Advanced fields are not part of the initial fixed plan. The generator retries applications rejected by normalized Unification, Tolerant, Lithovore or Strategic-Combat rules. Legal-action queries only read the persisted result.

### Fidelity boundary

Timing, shared RNG ownership, field range and the modeled rejection loop are now original-derived. MOOX intentionally keeps SplitMix64 rather than claiming bit-identical original LCG output, and original field-slot ordering is not yet stored as a separate normalized vector.

One original eligibility gate remains deliberately unresolved: Technology 52 `dimensional_portal` depends on global byte `0x21CAF`, whose semantic setting identity is not yet normalized.

See `UNCREATIVE_INITIAL_SELECTION_2026-08-28.md` for addresses and control/data-flow evidence.
## Server authority and command validation

The client never submits an ownership set.

Rules are:

```text
all:
    technology_id must be omitted
    server stores all legal applications

choose_one:
    technology_id is required
    it must exist in the projected ResearchChoice
    server stores exactly that one application

fixed_one:
    technology_id must be omitted
    server stores exactly its persisted Uncreative application
```

`CompleteResearchField` revalidates the active `ResearchState` against race + field policy before changing ownership. A corrupt/load-manipulated state cannot make a Human act Creative or replace an Uncreative fixed application.

## Observer / replay

`selection_mode` is carried by:

```text
empire.research_selected
empire.research_progressed
empire.research_completed
```

and is persisted in `ResearchState`.

This gives Observer/replay enough information to explain whether a project represented:

- a Human/AI application choice;
- a Creative/General all-applications project;
- an Uncreative server-fixed project.

No hidden model chain-of-thought is involved; this is authoritative game semantics only.

## State schema

This checkpoint bumps:

```text
StateSchemaVersion = 6
```

New persistent fields include:

```text
ResearchState.selection_mode
Empire.uncreative_research_choices
```

`EconomySchemaVersion` remains `5`; no new economy-rule scalar data was required.

## Tested behavior

Regression coverage includes:

- Human TechField 4 projects as `choose_one` with all three legal applications;
- missing application choice is rejected for ordinary research;
- an application from another TechField is rejected;
- selecting Reinforced Hull stores/grants only Technology 56;
- General TechField 55 projects as `all` for an ordinary Human;
- Psilon/Creative TechField 4 projects as `all` and grants all three applications;
- Klackon/Uncreative TechField 4 projects as `fixed_one` with the persisted application;
- an Uncreative client cannot override the fixed application;
- an Uncreative General field still uses `all`;
- same shared New Game RNG state yields the same persisted Uncreative plan and final RNG state;
- Uncreative initialization requires a caller-owned New Game RNG;
- querying `ResearchChoices()` leaves the serialized authoritative state/RNG byte-identical;
- existing GameSession authority, breakthrough and replay paths remain green.

## Open research / next slices

Initial Uncreative fixed-application timing/RNG ownership is now resolved in `UNCREATIVE_INITIAL_SELECTION_2026-08-28.md`. Active Research work should proceed as:

1. model Hyper-Advanced repeated-field level/cost state;
2. implement Advanced-start randomized/race-aware technology ownership;
3. later implement Uncreative external-acquisition replacement using the verified `Ensure_Uncreative_Field_OK_` path;
4. separately identify the `0x21CAF` Dimensional Portal eligibility gate when needed.

## Superseding Hyper-Advanced checkpoint

Hyper-Advanced fields 75..82 are now modeled separately as server-owned `repeat_field` research with no Technology IDs and dynamic per-Empire level costs. See `HYPER_ADVANCED_RESEARCH_2026-08-28.md`. This does not change the ordinary/Creative/Uncreative application policy defined in this checkpoint.

## Superseding switching checkpoint

Active-project switching and accumulated-RP transfer are now implemented and documented in `RESEARCH_SWITCHING_2026-08-27.md`. The `all` / `choose_one` / `fixed_one` policy defined here remains unchanged during switches.
