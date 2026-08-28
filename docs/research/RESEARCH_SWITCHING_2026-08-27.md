# Research project switching and RP transfer - 2026-08-27

Status: implemented deterministic research switching with complete accumulated-RP transfer.

This checkpoint extends the race-aware ResearchChoices/`empire.select_research` model so a running research project can be redirected without discarding the already accumulated Research Points.

## Evidence

The original manual's research-screen description explicitly states that the Research indicator can be used to change the current direction of research and that, when the goal is changed, all points already spent toward the former goal are applied to the new goal.

Independent secondary walkthrough/community examples describe the same behavior: a partially researched project can be switched to another field while keeping the accumulated RP pool.

That is sufficient evidence to implement RP transfer as a gameplay rule rather than a MOOX convenience.

## Command model

No second transport command is introduced. The existing authoritative command remains:

```text
empire.select_research
```

Its meaning is now:

```text
no active project:
    start research

active project + different legal target/application:
    switch research and transfer the complete RP pool

active project + identical target/application:
    reject as a no-op
```

All General/ordinary/Creative/Uncreative selection rules remain unchanged:

- `all`: client omits `technology_id`;
- `choose_one`: client supplies one server-projected `technology_id`;
- `fixed_one`: client omits `technology_id`; server uses the persisted Uncreative choice.

## RP transfer

Switching copies the previous active project's `ProgressRP` exactly into the new `ResearchState`:

```text
new.ProgressRP = old.ProgressRP
```

There is deliberately no scaling by old/new cost and no clamping to the new TechField's base cost.

Example:

```text
old field cost: 80 RP
accumulated:     70 RP
new field cost: 50 RP

new ProgressRP: 70 RP
```

The subsequent normal research-progress/breakthrough phase handles that over-funded project. Switching itself does not complete research and does not manufacture an additional carry-over pool.

Existing completion semantics still clear the active project. Therefore RP remaining above the completed target is not automatically carried into a later unrelated project.

## Field and application switching

MOOX supports both forms evidenced by the research UI model:

### Switch TechField

```text
TechField 4 / Reinforced Hull / 70 RP
        ->
TechField 55 / General field / 70 RP
```

### Switch application within the same TechField

For an ordinary `choose_one` field:

```text
TechField 4 / Reinforced Hull / 40 RP
        ->
TechField 4 / Anti-Missile Rockets / 40 RP
```

The Technology application changes; the accumulated field research pool is preserved.

Creative and Uncreative switches continue to use their `all` / `fixed_one` server-owned policies.

## Legal-action projection

`ResearchChoices()` no longer returns an empty list merely because a project is active.

The authoritative frontier remains available while researching so Human UI, built-in AI and remote agents can choose another legal target. This query remains read-only and does not change ProgressRP, RNG or the Uncreative fixed plan.

The currently active exact selection can appear in the projected choices, but submitting that identical selection is rejected as a no-op.

## Event / Observer / replay

Starting a project still emits:

```text
empire.research_selected
```

Changing an active project emits:

```text
empire.research_switched
```

with:

```text
empire_id
previous:
    tech_field_id
    selection_mode
    technology_ids
    technology_keys
current:
    tech_field_id
    selection_mode
    technology_ids
    technology_keys
transferred_rp
```

The event is command-attributed like the original selection command and is visible in the authoritative Observer/replay stream.

A two-turn `GameSession` regression test proves:

1. a project is selected and receives normal turn RP;
2. the strategic turn completes;
3. the next turn submits a different `empire.select_research` command;
4. `empire.research_switched` records the exact prior accumulated RP;
5. the new project begins with that pool;
6. current-turn RP is then added by the normal Research phase.

## No cancel-to-none command

This checkpoint does **not** add a command to cancel research entirely and leave the empire researching nothing.

The evidence found clearly supports changing the research direction. It does not establish a meaningful original strategic state where an empire intentionally abandons all research while preserving/discarding a pool.

MOOX therefore implements the evidenced switch operation only.

## State / schema

No new persistent state is required. Switching reuses `ResearchState` and the already-persisted selection mode/application set.

Therefore:

```text
StateSchemaVersion   = 6
EconomySchemaVersion = 5
```

remain unchanged.

## Tests

Coverage includes:

- ResearchChoices remain available with an active project;
- switch between two TechFields preserves ProgressRP exactly;
- switch between two ordinary applications in the same TechField preserves ProgressRP;
- switching to a cheaper target does not clamp the transferred pool;
- identical selection is rejected without changing serialized state;
- event previous/current snapshots and `transferred_rp` are exact;
- a two-turn authoritative `GameSession` switch is visible through Observer and receives the new turn's RP only after the transfer event.

## Next research question

The next active fidelity question is the strategic 1.31 phase order:

- Population growth/starvation,
- Food/Freighter balancing,
- Food/PP/RP/BC production,
- Construction,
- Research progress/breakthrough.

A secondary source described as checked against MOO2 1.31 conflicts with the current MOOX ordering. This remains unresolved and must be strengthened with independent/original-observed evidence before the resolver phases are reordered.

After that:

- exact Uncreative RNG/initialization timing if practical;
- hyper-advanced repeated-field state/cost;
- Advanced-start randomized/race-aware technology ownership.
