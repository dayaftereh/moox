# Original MOO2 1.31 Uncreative initial technology selection - 2026-08-28

Status: the timing and main RNG ownership of the original Uncreative fixed-application selection are directly verified from the original executable and implemented in MOOX.

This checkpoint answers the previously open question:

> When does MOO2 decide which single application an Uncreative empire is allowed to research in each ordinary TechField?

Answer:

```text
During new-game player technology initialization.
```

The original generates the per-field choices inside `Init_Player_Tech_`, called from `Init_Players_` while `Init_New_Game_` is running. It is therefore neither a first-open UI decision nor a research-selection-time roll.

## Reference binary

```text
C:\ASH\Temp\mastori2\Orion2.exe
SHA-256 7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5
```

The established bound-LE mapping is unchanged:

```text
code object:      object 1
object-1 base VA: 0x10000
```

Embedded Watcom debug symbols are used to anchor the functions below.

## Original new-game call chain

Relevant functions:

```text
Init_New_Game_     VA 0x12479
Init_Players_      VA 0x12983
Init_Player_Tech_  VA 0x5E55F
Random_            VA 0x1247A0
```

Direct calls:

```text
0x124EC -> Init_Players_     (0x12983)
0x12CF2 -> Init_Player_Tech_ (0x5E55F)
```

So the fixed Uncreative application plan is constructed while each player is initialized as part of a new game, before the first technology-selection UI, first turn or first research breakthrough.

## Shared new-game RNG ownership

`Init_Player_Tech_` calls the global original `Random_` function directly.

`Init_New_Game_` enters the game-generation RNG path before `Init_Players_`:

```text
0x12482 -> Get_Random_Seed_       (0x12484C)
0x12495 -> Set_Game_Random_Seed_  (0x107CA)
...
0x124EC -> Init_Players_
            -> Init_Player_Tech_
                 -> Random_
```

The supported fidelity conclusion is that Uncreative selection consumes the same new-game generation RNG stream used by surrounding game initialization. There is no original evidence for a dedicated Uncreative-only seed or PRNG stream.

MOOX still intentionally uses deterministic SplitMix64. This checkpoint copies RNG ownership/timing and gameplay-level draws, not bit-identical original LCG output for equal numeric seeds.

## Original application status model

The original per-player technology byte at:

```text
player + technology_id + 0x117
```

is used by the research code as application state. Relevant observed values are:

```text
0 = not currently available
1 = research/gettable application
3 = acquired/known application
```

`Give_Player_Field_` grants applications whose state is `1`; `Player_Gets_Tech_App_` changes an acquired application to state `3`.

## General fields are not reduced to one application

Inside `Init_Player_Tech_`, Uncreative suppression leaves normal/general handling in place when either the TechField research cost is 50 RP or the TechField ID is 23.

The normalized 50-RP fields are exactly:

```text
22, 28, 29, 55, 57
```

plus special General field `23`. Together with always-known field 0, this exactly matches MOOX `GeneralResearchFieldIDs` derived from:

```text
always_known_tech_field_id = 0
staged_known_tech_field_ids = 29,55,22,57,28,23
```

Therefore the existing rule remains correct:

```text
General field + any race -> selection_mode = all
```

Uncreative `fixed_one` applies only to ordinary fields.

## Original fixed-plan field range

`Init_Player_Tech_` walks TechFields with:

```text
cx = 1
...
cmp cx, 0x4A
jl  loop
```

Therefore the exact initial fixed-selection range is TechField `1..73` inclusive.

MOOX previously used `< 75`, which incorrectly included TechField 74. Normalized TechField 74 is the special Antaran technology field containing applications including `black_hole_generator`, `damper_field`, `death_ray`, `particle_beam`, `quantum_detonator`, `reflection_field`, `spatial_compressor` and `xentronium_armor`. It is not part of the ordinary Uncreative research plan.

MOOX now fixes the range to exactly `1..73`.

## Original per-field selection loop

For each TechField 1..73 the original scans the field's application slots. If any application in that field is already available/known, no Uncreative draw is needed; this naturally skips General fields.

Otherwise it:

1. counts the nonzero application slots;
2. calls `Random_(candidate_count)`;
3. indexes the selected field application;
4. checks whether that application is legal for the player/game mode;
5. if illegal, repeats the random candidate draw;
6. when legal, marks exactly that application state `1`.

Concrete window:

```text
0x5E853: candidate_count -> eax
0x5E857: call Random_ (0x1247A0)
0x5E85C: map 1..N result to field application slot
0x5E86D: call application eligibility helper (0x5E481)
0x5E874: if rejected, jump back and draw again
0x5E883: selected application state = 1
```

MOOX now mirrors this gameplay-level rejection loop through its caller-owned `NewGameRNG`.

## Original eligibility filters now normalized in MOOX

The helper at VA `0x5E481` contains several race/game-mode exclusions.

### Unification

For Unification it rejects:

```text
Technology 86  holo_simulator
Technology 141 pleasure_dome
Technology 195 virtual_reality_network
```

MOOX uses `GovernmentTraitID == government_unification`.

### Tolerant

For Tolerant it rejects:

```text
Technology 19  atmospheric_renewer
Technology 50  core_waste_dumps
Technology 113 nano_disassemblers
Technology 142 pollution_processor
```

MOOX uses the normalized `Tolerant` race flag.

### Lithovore

For Lithovore it rejects:

```text
Technology 6   android_farmers
Technology 29  biomorphic_fungi
Technology 68  food_replicators
Technology 87  hydroponic_farm
Technology 162 soil_enrichment
Technology 178 subterranean_farms
```

MOOX uses the normalized `Lithovore` flag.

### Strategic Combat

When Strategic Combat is active, the helper rejects applications whose original strategic-availability byte is false. MOOX already has normalized per-Technology strategic availability and now uses it in this rejection loop.

### Resolved Random Events gate: Dimensional Portal

The original helper rejects Technology 52 `dimensional_portal` when global byte `0x21CAF == 0`. Direct New Game UI, event-system and technology-path evidence now identifies `0x21CAF` as **Random Events enabled**. MOOX therefore excludes Technology 52 when `RandomEventsDisabled` is true. See `RANDOM_EVENTS_TECH_GATE_2026-08-28.md`.

## `Ensure_Uncreative_Field_OK_` is a replacement/repair path

Relevant functions:

```text
Ensure_Uncreative_Field_OK_ VA 0xE408F
Player_Gets_Tech_App_       VA 0xE4204
```

Direct xref analysis found exactly one direct caller of `Ensure_Uncreative_Field_OK_`:

```text
0xE43EE inside Player_Gets_Tech_App_
```

`Player_Gets_Tech_App_` first marks an acquired application state `3`. For an Uncreative player it then calls `Ensure_Uncreative_Field_OK_` for the corresponding field.

`Ensure_Uncreative_Field_OK_` leaves the field alone if an application is still state `1`; otherwise it examines eligible state-0 applications, uses reservoir-style `Random_` selection, and marks a replacement application state `1` if one exists.

This explains the future case where an Uncreative empire externally acquires its currently fixed application before researching that field. The original can select a replacement. MOOX currently persists one fixed choice and explicitly rejects this external-acquisition inconsistency; replacement remains a later separate slice.

## MOOX runtime change

Before this checkpoint MOOX used:

```go
UncreativeSelectionSeed uint64
```

and created a private SplitMix64 stream for the Uncreative plan. That was deterministic but not faithful to original RNG ownership.

`NewGameTechnologyOptions` now uses:

```go
NewGameRNG *core.RNG
```

For an Uncreative empire:

- `NewGameRNG` is required;
- the caller passes the shared new-game RNG;
- initialization mutates that RNG directly;
- when using `GameState`, the caller commits it back to `GameState.RNGState`;
- multiple Uncreative players consume one continuing stream in player-initialization order.

A non-Uncreative empire consumes no RNG through this Uncreative-selection path.

The persistent authoritative state remains `Empire.UncreativeResearchChoices[]`; no query-time randomness is introduced.

## ResearchChoice / protocol behavior is unchanged

The existing contract remains:

```text
ordinary     -> choose_one
Creative     -> all
General      -> all
Uncreative   -> fixed_one
```

For `fixed_one`, UI/AI receives exactly the persisted server-fixed application, the client does not submit its own Technology ID, `ResearchChoices()` is read-only, and active `ResearchState` stores the fixed application.

## Regression coverage

Tests now lock:

- Uncreative requires a caller-owned `NewGameRNG`;
- same initial New Game RNG state gives the same fixed plan and final RNG state;
- New Game RNG advances during Uncreative initialization;
- two Uncreative players consume one continuing shared stream deterministically;
- a normal player does not consume the Uncreative-selection RNG path;
- fixed-plan fields are only `1..73`;
- General fields are never included;
- TechField 74 is never included;
- Unification/Tolerant/Lithovore exclusions are enforced;
- Strategic Combat rejects a normalized non-strategic application;
- later `ResearchChoices()` queries remain state/RNG read-only.

## Schema impact

None.

```text
StateSchemaVersion   = 6
EconomySchemaVersion = 5
```

`UncreativeResearchChoices` was already persisted by State schema 6. `NewGameRNG` is a runtime initialization dependency, not persistent state.

## Fidelity boundary

This checkpoint does not claim original RNG bitstream identity. Original MOO2 uses its own `Random_` implementation/LCG; MOOX intentionally uses SplitMix64.

Copied semantics are:

- new-game initialization timing;
- shared RNG ownership;
- field order/range;
- General-field no-draw behavior;
- candidate redraw for modeled illegal applications;
- persisted result and query-time RNG silence.

Application-slot ordering is currently derived from normalized Technology membership rather than stored as a separate original field-slot vector. Because selection is uniform, gameplay distribution is correct; exact same-seed original/MOOX choices are not promised.

## Next Research work

The initial Uncreative selection timing/RNG question is closed.

Next narrow targets:

1. model original Hyper-Advanced repeated-field level/cost state;
2. then implement Advanced-start randomized/race-aware technology ownership;
3. wire the now-implemented Uncreative repair helper into future external Technology-grant transitions;
4. keep the resolved Random Events / Dimensional Portal gate centralized in those grant rules.