# Advanced Start Research - 2026-08-28

## Scope

This checkpoint records direct Master of Orion II 1.31 executable evidence for the randomized Advanced-start technology generator. It deliberately separates proven generator structure from the still-unimplemented MOOX runtime port.

## Original initialization order

`Init_Players_` at VA `0x12983` calls `Init_NPC_Personalities_Objectives_Themes_` (`0x589D6`) before it iterates players and calls `Init_Player_Tech_` (`0x5E55F`). Therefore the technology chooser can consume each player's already-materialized personality/objective/theme profile and can observe technology ownership granted to players initialized earlier in the array.

The technology-start mode byte at `0x21CB5` maps to the following `Init_Player_Tech_` iteration counts:

```text
Pre-Warp -> 1
Average  -> 6
Advanced -> 25
```

The first six iterations are the already-normalized staged TechFields:

```text
29, 55, 22, 57, 28, 23
```

Advanced therefore performs exactly **19 additional randomized grant iterations** after the Average baseline.

## Advanced extra-grant policy

For each extra iteration `Init_Player_Tech_` calls `Choose_Tech_Application_` at VA `0xFD335`. The selected concrete Technology identifies its TechField. The field is completed and its successor is opened.

Application acquisition is race-aware:

- ordinary empire: acquire the selected Technology application;
- Creative (`player+0x8B5`): acquire all legal applications in the selected TechField;
- Uncreative (`player+0x8B4`): the candidate statuses are already constrained by the fixed application plan generated earlier during the same New Game initialization, so the weighted chooser observes the server-fixed application rather than making a new client-visible choice.

Strategic Combat filtering remains part of Technology eligibility.

## Candidate frontier and cost tiers

`Choose_Tech_Application_` scans concrete Technology IDs `1..211`, but normal Advanced candidates are concrete normalized Technologies and require:

- application status `1` (researchable),
- owning TechField status `2` (open frontier),
- TechField `< 75`, excluding Hyper-Advanced repeats,
- Strategic-Combat availability when that game option is enabled.

After the six Average fields are complete the chooser adds the directly observed bonuses:

- TechField `4`: weight x2;
- Technology `114`: weight x2;
- Technology `51`: weight x5;
- TechField `73`: weight x2;
- a resulting zero weight is promoted to `1`.

The candidate-cost threshold starts at `15`. A candidate with field cost `cost` and current divisor `d` is admitted when `max(1, cost/d) <= threshold`, then receives the scaled chooser weight:

```text
scaled = base_weight * threshold / max(1, cost/divisor)
```

If no candidate is admitted the threshold advances by integer `floor(3*threshold/2)`, producing `15, 22, 33, 49, 73, ...` until at least one candidate exists.

At New Game the research divisor is zero and is normalized by the original code to `1`.

## AI/research metadata normalized in MOOX

`Calc_Tech_Value_` at VA `0xFC845` uses original static metadata that is now normalized instead of remaining as magic executable offsets:

- Technology record byte `+3`: AI class (`0..40`);
- 41 class records at file offset `0x1FB82A`, each containing a base weight plus the original competition-sensitivity flag;
- TechField AI group, already present in the 23-byte TechField table;
- 23 field-group progression values at file offset `0x201CA0`.

Technology ruleset schema **4** stores these as `ai_class`, `ai_research.technology_classes` and `ai_research.field_group_values` with original provenance.

## Race block mapping

Direct `Convert_Flags_To_Player_` / `Setup_Evolutionary_Upgrade_` analysis proves the original 30-byte Race block at `player+0x8A0..+0x8BD`:

```text
+8A0 population growth
+8A1 farming
+8A2 industry
+8A3 science
+8A4 money
+8A5 ship defense
+8A6 ship attack
+8A7 ground combat
+8A8 spying
+8A9 low-g world
+8AA high-g world
+8AB aquatic
+8AC subterranean
+8AD large home world
+8AE home-world mineral signed value (+1 rich / -1 poor)
+8AF artifacts world
+8B0 cybernetic
+8B1 lithovore
+8B2 repulsive
+8B3 charismatic
+8B4 uncreative
+8B5 creative
+8B6 tolerant
+8B7 fantastic traders
+8B8 telepathic
+8B9 lucky
+8BA omniscient
+8BB stealthy ships
+8BC trans-dimensional
+8BD warlord
```

MOOX must continue using semantic Race Trait IDs rather than persisting this legacy byte layout.

## Player preference / competition dependency

`Init_NPC_Personalities_Objectives_Themes_` writes the two additional preference bytes used by `Calc_Tech_Value_`:

- `player+0x205` - objective/profile selection;
- `player+0x206` - theme/profile selection.

The human personality sentinel at `player+0x28 == 100` is retained by that initializer, but `Init_Player_Tech_` temporarily substitutes profile `1` while starting Technologies are generated and restores `100` afterwards.

`Set_Competition_Tech_Values_` (`0xFD219`) is also called before weighting. It computes, for selected Technology AI classes, the highest TechField AI group already owned by **other players**. Because player technologies are initialized sequentially, later players can observe Advanced grants already assigned to earlier players.

This is why exact Advanced initialization belongs at the full New Game state boundary rather than inside a stateless per-Empire helper.

## Default-game secondary filter

A second chooser filter is gated by global `0x21CB0`. `Set_Default_Game_Settings_` sets this value to `0`, while other values also alter NPC personality weights. MOOX has not yet normalized that broader difficulty/AI setting, so the first Advanced implementation should reproduce the directly verified default path and keep the `0x21CB0 > 0` layer explicit/deferred rather than guessing its public option meaning.

## Next implementation slice

1. expose the normalized AI class/field-group metadata through `EconomyRules`;
2. introduce an explicit Advanced preference profile instead of embedding `+0x28/+0x205/+0x206` legacy bytes;
3. add a full-state New Game technology initializer so shared RNG and cross-player competition are deterministic and ordered;
4. port the default-path `Calc_Tech_Value_` and `Choose_Tech_Application_` weighting;
5. grant exactly 19 extra Advanced iterations with ordinary / Creative / Uncreative semantics;
6. test repeatability, strategic filtering and cross-player ordering.

The existing per-Empire initializer remains appropriate for deterministic Pre-Warp/Average ownership and lower-level tests.