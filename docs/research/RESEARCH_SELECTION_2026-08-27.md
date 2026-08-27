# Research selection and legal-action projection - 2026-08-27

Status: implemented first authoritative MOOX research-selection slice.

## Goal

Human UI and AI must choose research from the same server-derived legal-action surface. A client is never authoritative for which Technologies belong to a TechField.

## ResearchChoices

`EconomyRules.AvailableResearchChoices` produces a stable `ResearchChoice` list for one Empire.

Each choice contains:

- `tech_field_id`,
- predecessor and successor TechField IDs from the normalized field graph,
- `base_cost_rp` as `float64`,
- numeric Technology IDs,
- stable speaking Technology keys,
- language `name_key` values.

The first implementation defines the legal frontier as an unknown field whose normalized `previous_id` is either `0` or already present in `KnownTechnologyFieldIDs`.

Fields with no concrete normalized Technology IDs are not exposed yet. This prevents the current hyper-advanced placeholders from creating a ResearchState the runtime cannot materialize correctly.

When an Empire already has active research, `ResearchChoices` is empty. Switching an active project is intentionally not invented in this slice.

For the normalized Pre-Warp start the current tested frontier is:

```text
TechField IDs: 4, 7, 10, 18, 22, 28, 55, 57
```

For example TechField 4 has:

```text
base cost: 80 RP
previous field: 29
Technologies:
- anti_missile_rockets
- reinforced_hull
- fighter_bays
```

## Authoritative command

The gameplay command is:

```text
empire.select_research
```

Its client payload contains only:

```json
{
  "tech_field_id": 4
}
```

It does not contain Empire ID, Technology IDs, costs, progress or keys.

During strategic resolution the server:

1. derives the Empire from Seat authority,
2. recomputes `ResearchChoices`,
3. rejects a TechField outside that legal frontier,
4. copies the Technology IDs from the authoritative choice,
5. creates `ResearchState` with `ProgressRP = 0`,
6. emits `empire.research_selected` with numeric IDs plus speaking keys/name keys,
7. later in the same strategic resolution applies that turn's research output.

This makes the command safe for local UI, remote Human clients and AI controllers without separate gameplay logic.

## Session API

`GameSession.ResearchChoices(seatID, rules)` resolves Seat -> Empire under the same authoritative session boundary used by `BuildingChoices`.

Normal clients therefore follow:

```text
ResearchChoices
      |
      v
select one tech_field_id
      |
      v
SubmitTurn(empire.select_research)
      |
      v
server validates/materializes ResearchState
```

## Numeric model

Research selection uses the RP-native float architecture from `ADR-0002-research-float64.md`.

`ResearchState.ProgressRP` is `float64`; current Colony research milli-units are summed and then converted to RP without per-Colony integer truncation. The breakthrough percentage rounds only at its explicit percent/RNG boundary.

## Deferred

This slice does not yet define:

- changing/cancelling an active research project,
- Creative/Uncreative acquisition semantics,
- hyper-advanced repeated-field cost state,
- randomized/race-aware Advanced-start grants,
- frontend/Wails presentation,
- network/MCP-specific adapters.

Those adapters must consume the same `ResearchChoices` and `empire.select_research` semantics rather than reimplementing legality.
