# Slice 16.5 Gate 1 - Starting technology contract and visual-progression audit

Date: 2026-09-16
Status: **Gate 1 active - authoritative contract audited; visual prototype pending**

## Objective

Audit the original and normalized New Game technology-level contract before widening the live New Game validator or binding the visual selector. Gate 1 must keep three layers separate:

1. original MOO2 1.31 evidence;
2. normalized/runtime technology initialization already present in MOOX;
3. the still-narrow full New Game start-state generator.

The frontend must not infer or hard-code technology grants, especially for Advanced starts whose grants are deliberately race-, RNG- and cross-Empire-order dependent.

## Primary evidence

- `reference/original/text/help/block_0000.ascii.txt`, HELP records 548-551 - original New Game technology-level labels and player-facing start descriptions.
- `data/rulesets/moo2-1.31/technologies.json` - original-derived field/application data plus `new_game_start` field list and Strategic Combat availability.
- `docs/research/TECHNOLOGY_START_RESEARCH_2026-08-27.md` - direct executable evidence for Pre-Warp/Average staged fields and technology ownership.
- `docs/research/ADVANCED_START_RESEARCH_2026-08-28.md` and `ADVANCED_START_CHOOSER_2026-08-28.md` - direct executable evidence for the 19 extra Advanced grants, weighting, race semantics and ordering.
- `internal/game/research.go` - normalized `pre_warp` / `average` per-Empire initialization.
- `internal/game/research_advanced.go` - full-state Advanced initialization.
- `internal/game/new_game.go` - current production New Game validation/generation boundary.
- `internal/game/research_test.go`, `research_advanced_test.go`, `new_game_test.go` - current regression contract.

Targeted Gate-1 regression recheck passed on 2026-09-16:

```text
go test ./internal/game -run "TestInitializeEmpireTechnologiesFromOriginalStartFields|TestInitializeNewGameTechnologiesAdvancedEndToEnd|TestNewGameRejectsUnsupportedSettings" -count=1
ok moox/internal/game
```

## Canonical three-level contract

Original HELP explicitly lists exactly three starting-technology levels: Prewarp, Average and Advanced. MOOX already has stable normalized IDs for the same three states.

| Presentation label | MOOX ID | Original staged initialization | Gate-1 result |
| --- | --- | --- | --- |
| Pre-Warp | `pre_warp` | first staged field after always-known field 0 | canonical |
| Average | `average` | all six staged fields after field 0 | canonical |
| Advanced | `advanced` | Average baseline + exactly 19 randomized extra field grants | canonical |

The product target is therefore genuinely three levels. The blocker is not naming or missing technology research data; it is full New Game start-state parity/integration for Pre-Warp and especially Advanced.

## Normalized staged fields

`technologies.json:new_game_start` records:

- always-known TechField: `0`;
- staged fields in original order: `29, 55, 22, 57, 28, 23`.

### Pre-Warp technology ownership

Known fields:

```text
0, 29
```

Known normalized Technology IDs:

```text
32, 40, 103, 145, 166, 168
```

Current names:

| ID | Normalized key |
| ---: | --- |
| 32 | `capitol` |
| 40 | `colony_base` |
| 103 | `marine_barracks` |
| 145 | `pulse_rifle` |
| 166 | `spy_network` |
| 168 | `star_base` |

This is deterministic. It intentionally does **not** include Nuclear Drive, Colony Ship or the other Average interstellar baseline applications.

### Average technology ownership

Known fields:

```text
0, 22, 23, 28, 29, 55, 57
```

Tactical-combat Technology IDs:

```text
32, 40, 41, 58, 63, 69, 100, 101, 103, 109,
119, 120, 121, 145, 157, 166, 167, 168, 187, 189
```

This is 20 deterministic applications. Technology 63 (`extended_fuel_tanks`) is unavailable under Strategic Combat, yielding the independently observed 19-application original save set. Slice 16 keeps Tactical Combat fixed, so the product-facing Slice-16 Average facts should describe the Tactical 20-application baseline rather than exposing a combat-mode branch the player cannot choose.

### Advanced technology ownership

Advanced uses the deterministic Average baseline first, then exactly **19 additional TechField grants** per Empire. Therefore each Empire ends with 26 completed starting fields (`7 + 19`).

The concrete application set is intentionally **not** a static catalog:

- one shared caller-owned New Game RNG drives weighted selection;
- candidates come from the current open research frontier;
- race research modifiers affect weights;
- personality/objective/theme preference values affect weights;
- selected applications/fields owned by earlier Empires can affect competition weights for later Empires;
- ordinary races acquire the selected legal application;
- Creative acquires every legal application in the chosen field;
- Uncreative follows its fixed New Game application plan;
- Strategic Combat availability remains an eligibility filter.

A UI card may safely state `Average baseline + 19 extra research fields`, but it must **not** promise a fixed Advanced technology list or a fixed technology-count total.

## Race-specific interactions

### Pre-Warp / Average

The deterministic known-field/application baseline is race-independent for the normalized start fields. None of the currently normalized race-specific ineligible technology IDs intersects the Pre-Warp/Average ownership sets.

Uncreative is still significant: initialization builds its fixed future research-application plan using the caller-owned New Game RNG. This affects deterministic RNG consumption/state even though the initially known Pre-Warp/Average Technology IDs remain the same.

### Advanced

Advanced is explicitly race-aware. `RaceResearchModifiers` projects original race traits into chooser weights, including food/industry/science/money deltas, combat/spying modifiers and special traits such as Cybernetic, Lithovore, Tolerant, Telepathic, Stealthy Ships and government family.

Technology eligibility also excludes applications made invalid/redundant by race/game-mode rules. Examples already normalized in the shared eligibility helper include:

- Unification exclusions for morale-building applications;
- Tolerant exclusions for pollution-management applications;
- Lithovore exclusions for food/farming applications;
- Strategic Combat-unavailable applications.

Creative/Uncreative change ownership semantics after a field is selected, so race identity can change both **which** Advanced fields/applications are chosen and **how many applications** are actually granted.

## Advanced preference integration blocker

`InitializeNewGameTechnologies` already supports Advanced, but `NewGameTechnologyStateOptions` requires `AdvancedPreferences` per Empire. The current production `EconomyRules.NewGame` call supplies only:

- technology level;
- Tactical/Strategic flag;
- shared New Game RNG.

It does not yet materialize/pass the personality/objective/theme preference map. There is no non-test caller of `AdvancedPreferences` outside the research implementation/tests. Widening validation to `advanced` today would therefore fail during initialization even before the broader Advanced empire/fleet start-state differences are addressed.

Gate 2/3 must choose and freeze the correct server-owned New Game preference-generation seam rather than fabricating a frontend-visible profile.

## Full New Game start-state effects

Original HELP makes clear that technology level is not merely a research-ownership dropdown.

### Pre-Warp - original player-facing contract

- starts at a single star;
- starts with **no ships**;
- even basic interstellar capability must first be researched.

Direct executable research already proves non-Advanced homeworld Population is set to 8, so Pre-Warp belongs to that non-Advanced population branch.

### Average - original player-facing contract

- starts at a single star;
- two Scouts;
- one Colony Ship;
- basic interstellar technology baseline.

The current Slice-09 New Game generator implements this Average baseline: Population 8, 2 Scouts, 1 Colony Ship, 50 BC, 0 freighters and the deterministic Average technology set.

### Advanced - original player-facing contract

- starts with a **larger empire** that fills the player's share of the galaxy according to galaxy size/player count;
- starts with several levels of technological advancement;
- starts with a fleet that consumes the available Command Points.

Existing executable research fully normalizes the Advanced technology ownership algorithm, but the exact broader Advanced colony/territory/fleet bootstrap is not yet normalized into the production New Game generator. `Init_Homeworld_Colony2_` evidence also shows Population 8 is specifically the **non-Advanced** branch, confirming that simply reusing the Average colony bootstrap for Advanced would be false parity.

## Current MOOX production boundary

`validateNewGameSettings` currently accepts only `technology_level == average`. This is correct defensive behavior for the present full New Game generator.

If that guard were simply removed today:

- `pre_warp` technology ownership could initialize, but the generator would still incorrectly create the Average two-Scout + Colony Ship start;
- `advanced` would first fail because production New Game does not provide `AdvancedPreferences`;
- even after that preference seam were wired, the generator would still incorrectly reuse the single-homeworld/Average fleet/population bootstrap instead of the original larger-empire/CP-filled Advanced start.

Therefore Gate 3 must widen the **full authoritative start-state contract**, not only the validator enum.

## UI/fact implications for Gate 2

Safe evidence-backed fact candidates:

### Pre-Warp

- `Single star`
- `No starting ships`
- `Interstellar capability must be researched`
- optional technical detail: `2 completed starting fields / 6 known applications`

### Average

- `Single star`
- `2 Scouts + 1 Colony Ship`
- `7 completed starting fields / 20 Tactical applications`

### Advanced

- `Larger starting empire`
- `Average baseline + 19 extra research fields`
- `Starting fleet fills Command Points`
- `Exact grants vary by seed, race and initialization context`

Do not display a fixed Advanced application count/list.

## Visual progression direction to prototype

The three illustrations should communicate **capability/state progression**, not a literal technology-tree screenshot:

1. **Pre-Warp / workshop-orbit dawn** - one planetary industrial/workshop focal point, incomplete orbital structure, no outbound fleet silhouette; restrained instrumentation and sparse energy accents.
2. **Average / interstellar launch** - mature command/research facility with a small readable scout pair plus colony-vessel motif; balanced instrumentation and established orbital infrastructure.
3. **Advanced / networked stellar command** - multi-node stellar infrastructure, denser fleet/command presence and a refined luminous research/command core; broader spatial footprint rather than merely making the same image brighter.

All three must share one composition grammar, center-safe focal region and Slice-16.1 selector framing. Artwork must remain original MOOX material and must not copy original MOO2 UI/art.

## Gate-1 provisional conclusion

- Exactly three canonical levels are confirmed: `pre_warp`, `average`, `advanced`.
- Technology ownership semantics are already strongly normalized for all three.
- Average is the only full production New Game start currently accepted.
- Pre-Warp needs its no-ship/full-start bootstrap before it is honestly selectable.
- Advanced additionally needs preference generation plus the larger-empire/fleet/population start contract; enabling only its technology initializer would be incomplete.
- The visual selector can safely prototype all three now, but Gate 2 must freeze whether Slice 16.5 implements the complete three-level full-state breadth or intentionally splits unresolved Advanced empire-bootstrap fidelity into a bounded follow-up.

Gate 1 remains active until the three-image progression prototype and 390 px readability review are recorded.
