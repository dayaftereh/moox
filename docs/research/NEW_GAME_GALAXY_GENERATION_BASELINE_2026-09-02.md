# Slice 09 research - New Game / galaxy generation baseline

Status: **closed; Gates 1-4 complete**
Opened: 2026-09-02

## Objective

Replace the fixture-only startup path with the first authoritative deterministic `NewGame(seed, settings)` path that can create a playable minimal galaxy, Empires, Homeworlds and starting state and feed that state into the Slice-08 server/web application boundary.

The selected first target remains the prepared Slice-09 scope: a directly evidenced two-Empire, small-galaxy baseline rather than a broad full-MOO2 New Game implementation.

## Repository state at Gate-1 open

- Starting HEAD: `c41d425` (`docs: close server web hmi transport slice`).
- Branch: `main`, ahead of `origin/main` by 37 commits.
- Starting working tree: clean.
- Existing `_OPEN_*.md` markers before Slice 09: 0.
- Slice 08 is closed and provides the authoritative application/server transport boundary.
- Core `StateSchemaVersion`: 21 at open.
- At Gate-1 open, no Slice-09 gameplay implementation had started.

## Gate-1 questions

1. Which current fixture/state constructors and normalized galaxy/star/planet/race/start data can be reused without smuggling test-only assumptions into production New Game?
2. What are the original MOO2 1.31 New Game entry points and which object owns the RNG stream from settings selection through galaxy and starting-state construction?
3. Which small-galaxy star-count and coordinate constraints can be proven for a reproducible first fixture?
4. In what order are stars/systems/planets generated, and which random draws must be reproduced to keep deterministic results stable?
5. How are player Homeworlds/start systems selected, including spacing, ties and collision handling?
6. What exact Empire, Homeworld Colony, Population, Treasury, Research and starting Fleet/Ship inventory exists for the selected technology/start level?
7. Which settings/state rules are directly proven original behavior, and which narrow choices are deliberate MOOX modernization/simplification?
8. What exact production `NewGame(seed, settings)` contract should Gate 2 accept before implementation?

## Gate-1 rule

This document is evidence/research only until Gate 2 is accepted. Do not implement New Game/galaxy generation behavior during Gate 1.


## Original reference and methodology

Primary executable reference for this Gate:

- `C:\ASH\Temp\mastori2\Orion2.exe`
- Master of Orion II 1.31
- SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- bound LE executable with embedded Watcom debug symbols

The Gate reused the repository's verified `internal/moo2exe` LE reader and an ignored research-only helper under `build/research-disasm/` to resolve Watcom symbols, read exact LE object ranges and disassemble only the relevant New Game/Galaxy functions. The helper is under the already ignored `build/` tree and is not production code or a Gate-3 implementation.

Existing repository evidence was also reused where it already directly established original behavior, especially:

- `docs/research/UNCREATIVE_INITIAL_SELECTION_2026-08-28.md` for New Game RNG ownership/order;
- `docs/research/TECHNOLOGY_START_RESEARCH_2026-08-27.md` for Pre-Warp/Average technology ownership;
- `docs/research/TREASURY_SETTLEMENT_2026-08-28.md` for the original 50-BC start;
- `docs/research/PRESET_RACES_2026-08-26.md` and normalized race files for preset traits;
- `docs/research/POPULATION_COHORTS_2026-08-29.md` for the deliberate continuous MOOX Population/job representation;
- `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md` for the original colonizable body-state predicate and current special Colony Ship model.

## Current repository start-state inventory

There is currently **no production New Game/Galaxy generator**.

`internal/core/fixture.go` exposes `core.NewSmallFixture(seed)`, but its source comment explicitly states that the layout is test scaffolding and is not a claim about original MOO2 galaxy-generation probabilities. It hard-codes three named systems and one Human Empire/Colony. Its 2 Farmer / 1 Worker / 1 Scientist Population layout therefore must not leak into production New Game semantics.

Useful production-ready foundations already exist:

- `core.NewGameState(seed)` creates schema-21 state, turn 1, `RNGState=seed`, stable ID allocation and empty strategic collections;
- `GameState.RNG()` / `CommitRNG()` provide one caller-owned deterministic MOOX RNG stream;
- normalized planet size/mineral/gravity/climate identifiers already exist in `planet_classes.json`;
- preset races and race traits are normalized;
- Average/Pre-Warp research initialization already exists and accepts caller-owned New Game RNG;
- original starting Treasury 50 BC is already normalized;
- Colonies, Population cohorts, Ships, ShipDesigns, StrategicFleets and special Colony Ship fleets already have authoritative Core representations;
- Slice 08 provides a transport-neutral application host plus HTTP/WebSocket browser boundary into which real New Game creation can be inserted.

## Original New Game entry points and RNG ownership

Important Watcom symbols / object-1 offsets found directly in the 1.31 executable:

| Function | Object-1 offset | VA |
| --- | ---: | ---: |
| `Init_New_Game_` | `0x2479` | `0x12479` |
| `Init_Players_` | `0x2983` | `0x12983` |
| `Init_Homeworld_Colony2_` | `0x3A3D` | `0x13A3D` |
| `Init_Homeworld_Colony_` | `0x3FA7` | `0x13FA7` |
| `Init_Ship_Designs_` | `0x44FBF` | `0x54FBF` |
| `Load_Colony_Ship_Design_` | `0x464CD` | `0x564CD` |
| `Build_Home_Star_List_` | `0x6AD13` | `0x7AD13` |
| `Build_Min_Star_Distances_` | `0x6B020` | `0x7B020` |
| `Generate_Home_Worlds_` | `0x6B8CD` | `0x7B8CD` |
| `Modify_Home_Worlds_` | `0x6C4AF` | `0x7C4AF` |
| `Randomize_Home_Worlds_` | `0x6C6EF` | `0x7C6EF` |
| `Generate_Climate_` | `0x7BEAB` | `0x8BEAB` |
| `Generate_Gravity_Class_` | `0x7BFE0` | `0x8BFE0` |
| `Generate_Mineral_Class_` | `0x7C05B` | `0x8C05B` |
| `Generate_Number_Of_Satellites_` | `0x7C527` | `0x8C527` |
| `Generate_Orbit_` | `0x7C567` | `0x8C567` |
| `Generate_Satellite_` | `0x7C5D7` | `0x8C5D7` |
| `Generate_Satellite_Type_` | `0x7C6FE` | `0x8C6FE` |
| `Generate_Spectral_Class_` | `0x7C807` | `0x8C807` |
| `Set_Star_XYs_` | `0x7CFFF` | `0x8CFFF` |
| `Star_XY_Invalid_` | `0x7D43D` | `0x8D43D` |
| `Universe_Generation_` | `0x7DAE8` | `0x8DAE8` |
| `Planet_Generation_` | `0x7E280` | `0x8E280` |
| `Twiddle_Initial_Homeworlds_` | `0xD5832` | `0xE5832` |

Existing direct disassembly already established the New Game seed sequence:

1. `Init_New_Game_` calls `Get_Random_Seed_`;
2. it calls `Set_Game_Random_Seed_`;
3. it calls `Init_Players_`;
4. player/research initialization consumes the same global `Random_` stream;
5. New Game later performs its Universe/Homeworld generation from that same game RNG lifecycle.

The repository's current `core.RNG` is intentionally not claimed to be numerically identical to the original 1996 LCG for the same numeric seed. Slice 09 should preserve **one shared New Game RNG owner and a frozen MOOX draw order**, not invent a false equal-seed bit-parity guarantee against the original executable.

## Original New Game generation order

Direct `Init_New_Game_` disassembly establishes this high-level ordering:

1. seed/global initialization;
2. player initialization (`Init_Players_`);
3. ship/colony/global setup;
4. `Universe_Generation_`;
5. `Generate_Home_Worlds_`;
6. planet/link/index normalization;
7. Advanced-only extra colonies/fleet when applicable;
8. later original-only content such as monsters, wormholes and additional specials;
9. star-owner/fleet-slot updates;
10. `Twiddle_Initial_Homeworlds_` and subsequent final start setup.

Slice 09 should preserve the proven semantic ordering through players -> universe -> homeworld normalization -> starting strategic assets. Orion, monsters, nebulae, wormholes, marooned heroes, splinter colonies and broad system specials remain explicit later work.

## Galaxy size and star-count constraints

`Universe_Generation_` directly switches on original galaxy-size index and writes both `_NUM_STARS` and the grid dimensions passed toward star placement:

| Original size index | Stars | Grid columns | Grid rows |
| ---: | ---: | ---: | ---: |
| 0 | **20** | **5** | **4** |
| 1 | 36 | 6 | 6 |
| 2 | 54 | 9 | 6 |
| 3 | 71 | 9 | 8 |

The selected Slice-09 baseline is therefore original size index 0: **Small = exactly 20 stars on a 5 x 4 placement grid**.

Other sizes may be normalized into data for future use, but the production `NewGame` validator should reject them in this first slice.

## Original star-coordinate behavior and the MOOX normalization boundary

`Set_Star_XYs_` is not a simple fixed grid. Direct disassembly shows it:

- uses repeated global `Random_` calls for coordinates inside a grid-derived map region;
- stores X/Y in the original star record;
- calls `Star_XY_Invalid_` and retries invalid positions;
- contains an individual-placement retry counter around 150 before fallback behavior;
- validates pairwise proximity in scaled map coordinates;
- applies an additional special rule for original spectral class 6: those objects must remain at least 5 parsecs apart.

`Star_XY_Invalid_` includes a squared-distance comparison against `0x320` (800) in one proximity branch plus axis-based proximity checks. The exact map-scale/UI-coordinate pipeline is tightly coupled to the original rendering scale and is not represented by current Core coordinates.

### Accepted Gate-2 proposal for coordinates

Do **not** pretend current MOOX coordinates are original map pixels. Freeze a deliberate algorithm named conceptually `moox_grid_jitter_v1`:

- logical map width 1000, height 800;
- Small grid 5 x 4;
- cell centers at `(100 + 200*column, 100 + 200*row)`;
- in row-major cell order consume exactly two MOOX RNG draws per star;
- each axis jitter is `Intn(81)-40`, yielding X range 60..940 and Y range 60..740;
- no coordinate rejection loop in this baseline because the cell geometry itself guarantees useful separation;
- star IDs follow row-major generation order.

This preserves the proven Small 5x4 topology and seeded variation while explicitly versioning the coordinate choice as **MOOX normalization**, not original numeric parity. A later fidelity slice may port the full original scaled rejection function without changing the public New Game settings contract.

## Galaxy age

Original HELP record 544 states verbatim in semantic terms that Galaxy Age cycles through:

1. **mineral rich**;
2. **normal**;
3. **organic rich**.

It explains that mineral-rich age increases mineral-resource planets and reduces food-world likelihood, while organic-rich does the opposite.

The executable global is `_g_galaxy_age` (`0x21F2F`). `Generate_Climate_` branches over values 0,1,2 and `Generate_Spectral_Class_` selects one of three weight columns. The HELP cycle order plus the executable three-way branch gives the Gate-2 mapping:

- 0 = `mineral_rich`
- 1 = **`normal`**
- 2 = `organic_rich`

Slice 09 supports only **Normal galaxy age (index 1)**.

## Original spectral generation

`Generate_Spectral_Class_` builds seven weights and calls `Get_Weighted_Choice_Char_`. The exact 7 x 3 table starts at original object-2 offset `0x583E`.

Raw rows are spectral class 0..6; columns are Mineral Rich / Normal / Organic Rich:

```text
20 10  5
25 15  5
10 16 30
10 16 21
32 37 30
 1  2  3
 2  4  6
```

Each column sums to 100. For the selected **Normal** age, the spectral weights are therefore:

```text
10, 15, 16, 16, 37, 2, 4
```

The current Core does not persist spectral class. Gate 2 should keep spectral class in a transient generation record so it can drive planet probabilities without forcing a Core schema bump or prematurely exposing star presentation state.

## Original planet-generation order

`Planet_Generation_(starID)` performs:

1. `Generate_Orbit_(starID)`;
2. `Generate_Satellite_Type_(starID, orbit)`;
3. if the body is an accepted generated body type, create its record and call `Generate_Satellite_`;
4. increment planet/body counters.

`Generate_Satellite_` has the directly observed semantic order:

1. `Generate_Size_` -> `Random_(10)` and the already-normalized size thresholds 1/3/7/9/10;
2. `Generate_Mineral_Class_` -> `Random_(10)` plus spectral-class table;
3. `Get_Climate_Modifier_` (no RNG);
4. `Generate_Gravity_Class_` (table lookup; no RNG inside this function);
5. `Get_Planet_Group_`;
6. `Generate_Climate_` -> weighted choice for the selected Galaxy Age, with food-world reroll behavior when the original `_make_food_planet` flag is active;
7. `Random_(3)-1` for a secondary planet property;
8. secondary-data setup.

This exact semantic ordering is important even though MOOX uses its own RNG implementation.

### Satellite/body-count table

`Generate_Number_Of_Satellites_` performs exactly one `Random_(10)`, decrements to roll index 0..9 and indexes the 10 x 6 table at object-2 offset `0x5680` by `roll*6 + spectralClass`:

```text
roll1:   0 0 1 2 0 0
roll2:   1 1 2 2 1 0
roll3:   1 1 2 2 1 0
roll4:   2 1 2 3 1 0
roll5:   3 2 3 3 2 0
roll6:   3 2 3 4 2 0
roll7:   4 3 4 4 2 0
roll8:   4 3 4 5 3 1
roll9:   5 4 5 5 3 1
roll10:  5 4 5 5 4 1
```

The function accepts normal spectral classes below 6; original spectral class 6 is handled outside the normal planet-count path.

### Satellite/body type table

`Generate_Satellite_Type_` consumes `Random_(10)` and indexes a 10-roll x 5-orbit table at object-2 offset `0x56BC`:

```text
1 1 1 1 1
4 1 1 1 2
3 2 1 2 2
3 3 2 2 2
3 3 2 2 2
3 3 3 3 2
3 3 3 3 3
3 3 3 3 3
3 3 3 3 3
3 3 3 3 3
```

Special body type 4 has additional `Random_(100)` handling, and type 2 at orbit zero can reroll.

Prior direct colonization research established original `planet+0x04 == 3` as the real normal colonizable-planet body state. Current Core `Planet` semantically represents colonizable planets and has no gas-giant/asteroid body type.

**Gate-2 boundary:** keep original body type as transient generation state; materialize only body type **3** as `core.Planet`. Types 1/2/4 remain unmaterialized presentation/special bodies in Slice 09. This avoids a Core21 schema bump while preserving the original probability pipeline for colonizable worlds.

### Mineral generation

`Generate_Mineral_Class_` consumes `Random_(10)` and indexes object-2 table `0x56EE` by `roll*6 + spectralClass`:

```text
2 1 1 0 0 1
2 1 1 1 0 2
2 2 1 1 1 2
2 2 2 1 1 2
3 2 2 1 1 2
3 2 2 2 1 2
3 3 2 2 2 3
3 3 3 2 2 3
4 3 3 2 2 3
4 4 4 3 2 4
```

Values are the normalized mineral-class indexes 0..4.

### Gravity generation

`Generate_Gravity_Class_` performs no RNG. It reads the generated size and mineral class and indexes the first 25 bytes of object-2 table `0x572A` as a 5 x 5 mineral-by-size matrix:

```text
0 0 0 1 1
0 0 1 1 1
0 1 1 1 2
1 1 1 2 2
1 1 2 2 2
```

Values are gravity indexes 0=Low-G, 1=Normal-G, 2=Heavy-G.

### Climate generation

`Generate_Climate_` uses Galaxy Age plus planet group to construct 10 climate weights and calls `Get_Weighted_Choice_Char_`. The Normal-age baseline uses the original normal branch/table; climate IDs remain the normalized 0..9 sequence:

```text
toxic, radiated, barren, desert, tundra,
ocean, swamp, arid, terran, gaia
```

Gate 3 normalized the exact age/group weight rows into `data/rulesets/moo2-1.31/new_game_galaxy.json`; gameplay code no longer embeds those executable offsets. The raw source ranges remain object-2 `0x57A7` / `0x57CF`.

### Orbit selection boundary

The original `Generate_Orbit_` builds five orbit weights, calls `Get_Weighted_Choice_Char_` and rerolls occupied orbit slots. Its weight source is embedded through an original runtime pointer whose complete semantic settings mapping is not yet normalized into repository data.

Gate 2 should therefore make this one narrow and explicit MOOX choice rather than claiming exact original orbit parity:

- for each requested body count, build the list of currently unused orbit indexes 0..4;
- choose one unused orbit using `rng.Intn(len(unused))`;
- keep orbit indexes unique and sorted when materializing Core planets.

The body-type/size/mineral/gravity/climate tables remain original-derived; **orbit choice is an explicitly versioned MOOX simplification** in this first slice.

## Original Homeworld star selection

`Generate_Home_Worlds_` directly shows:

- every star begins without a home-owner;
- it computes candidate-distance information through `Build_Min_Star_Distances_`;
- it calls `Build_Home_Star_List_`;
- it requires a candidate set containing at least `_NUM_PLAYERS` stars;
- it retries up to **five** increasing/relaxing separation scales;
- `Build_Home_Star_List_` computes pairwise `Dist_Between_Stars_` and applies eligibility bitfields;
- candidate/tie selection consumes global `Random_`;
- selected candidates are shuffled repeatedly;
- `Randomize_Home_Worlds_` further randomizes which selected home star belongs to each player.

For two players, the original intent is therefore unequivocally **well-separated candidate stars plus randomized assignment**, not ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â‚¬Å¡Ã‚Â¬Ãƒâ€¦Ã¢â‚¬Å“first two starsÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â‚¬Å¡Ã‚Â¬Ãƒâ€šÃ‚Â.

The exact distance metric is tied to the original scaled map coordinate system that Gate 2 deliberately does not claim to reproduce numerically.

### Gate-2 Homeworld-selection simplification

For the first MOOX two-player baseline, choose the **farthest unordered pair of generated star coordinates**. If several pairs have identical squared distance, choose lexicographically by `(lowerStarID,higherStarID)`. Assign the first player to the lower StarID and the second to the higher StarID.

This consumes no extra RNG and is intentionally named/documented as a MOOX deterministic normalization. It preserves the original strong-separation purpose without fabricating an unverified conversion from original screen-scale thresholds into Core coordinates. Future full homeworld-selection fidelity can replace this internal algorithm behind the same public settings version.

## Original Homeworld physical normalization

`Modify_Home_Worlds_` directly proves that every chosen home system is normalized after general galaxy generation.

For each home system it:

- calls `Enforce_3_Planet_Minimum_`;
- determines race-driven home climate;
- normal non-Aquatic races use climate index **8 = Terran**;
- Aquatic uses climate index 5 = Ocean;
- normal races call `Enforce_Medium_Planet_Size_`;
- direct disassembly of that helper shows it writes size index **2 = Medium** exactly;
- Large Home World forces size index 3 = Large;
- after planet-type enforcement, default mineral class is explicitly set to **2 = Abundant**;
- Rich/Poor homeworld traits override to 3/1;
- Low-G / High-G / default force gravity 0 / 2 / **1 Normal-G**;
- Artifacts trait writes an original special flag;
- it initializes the Homeworld Colony;
- it calls `Guarantee_A_Planet_At_Min_Dist_(homeStar,4)`;
- it assigns home-star names;
- it ensures at least one food-capable planet exists.

Normalized preset evidence confirms **Human** and **Darlok** have no Large/Rich/Poor/Artifacts/Low-G/High-G/Aquatic homeworld modifier traits. They are therefore the safest exact physical Homeworld fixture for Slice 09.

For both selected races, Gate 2 can directly require the Homeworld itself to be:

```text
Size:     Medium
Mineral:  Abundant
Gravity:  Normal-G
Climate:  Terran
```

and the home system must contain at least **3 materialized colonizable planets** after normalization.

If ordinary generation produced fewer than three materialized type-3 planets, the MOOX baseline should fill the lowest unused orbit indexes deterministically with newly generated normal planet records using the same shared RNG and then force only the selected Homeworld physical class. This fill operation is a compatibility implementation of the original `Enforce_3_Planet_Minimum_` purpose; the exact original internal body-conversion sequence remains deferred.

## Original starting player/colony state

### Treasury and Freighters

Direct `Init_Players_` disassembly writes:

```text
player+0x32 = 0x32 = 50 BC
player+0x36 = 0      = 0 Freighters
```

The 50-BC value is independently cross-checked by the original `SAVE10.GAM` evidence already documented in the repository.

So Slice 09 should start each selected Empire with exactly:

- Treasury **50 BC**;
- Freighters **0**.

### Technology level

Original HELP record for Average Civilization states that it begins at one star, on the cusp of interstellar travel, with two Scouts and a Colony Ship.

Existing executable/research normalization proves the unfiltered Average technology ownership set is exactly:

```text
32, 40, 41, 58, 63, 69, 100, 101, 103, 109,
119, 120, 121, 145, 157, 166, 167, 168, 187, 189
```

The original Strategic-Combat save lacks only technology 63 because its Strategic availability flag is false. Slice 09's selected baseline is **Tactical Combat**, so it uses the full 20-ID Average set.

### Homeworld Population

`Init_Homeworld_Colony2_` directly sets non-Advanced starting Population to **8**.

The original stores whole Population entries with packed jobs. Current MOOX intentionally modernizes this into continuous Farmer/Worker/Scientist quantities. Gate 1 did not find a sufficiently clean direct mapping of the original initial packed job assignment into that continuous representation, so it must not be presented as original fact.

Gate-2 proposed deterministic MOOX bootstrap:

```text
Total       8
Farmers     4
Workers     2
Scientists  2
```

This is explicitly a **MOOX continuous-job bootstrap**, chosen so a normal Terran homeworld can sustain its starting Population while still exercising Production and Research. It is not a claim that the 1996 UI initially displayed this exact job split.

The sole initial cohort should have:

- `OriginEmpireID = owning Empire`;
- `LoyaltyEmpireID = owning Empire`;
- existing fully loyal/assimilated native-owner state;
- the 4/2/2 job allocation above.

### Starting Buildings

`Init_Homeworld_Colony2_` clears the original building array and then scans an original candidate product list, current technology level, Population and `Colony_Can_Build_Product_` before conditionally calling `Add_Building_`.

The original candidate list at object-2 `0x58AC` was directly dumped, but the complete technology-/race-/product-specific free-start condition matrix is outside the narrow first Galaxy/New Game objective.

Gate 2 should therefore **not guess** Marine Barracks, Star Base or any other free building. The baseline Homeworld starts with an empty `Buildings` set and uses current normal buildability immediately afterward. This is an explicit Slice-09 MOOX simplification and a documented future fidelity gap.

### Starting fleet

Direct `Init_Homeworld_Colony2_` disassembly proves that the Average path (`technology_level == 1`) creates:

- two ships using the same baseline starting design/type;
- then calls `Load_Colony_Ship_Design_`;
- then creates one Colony Ship.

This independently matches original HELP: **two Scouts + one Colony Ship**.

The exact auto-designed Scout component/special loadout is not yet normalized; pulling that in would require broad generic ship-design/special behavior, including currently deferred design specials. Gate 2 should preserve the exact unit count/roles while making the Scout snapshot an explicit MOOX baseline:

```text
Name: Scout
Hull: Frigate
Drive: Nuclear Drive
Computer: Electronic Computer
Armor: Titanium Armor
Fuel: Standard Fuel Cells
Shield: none
Weapons: none
Specials: none
```

Per Empire:

- one saved Scout `ShipDesign`;
- two concrete `Ship` snapshots from that design;
- one `StrategicFleetRoleCombat` fleet at the Home system containing those two Ships;
- one current supported special Colony Ship strategic fleet at the Home system.

This is sufficient for real post-New-Game expansion while leaving exact original Scout auto-design fidelity to a later ship-design slice.

## Selected first canonical New Game fixture

Gate 1 recommends freezing Slice 09 to exactly this first supported production configuration:

```text
Galaxy size:       Small (original index 0; 20 stars, 5x4)
Galaxy age:        Normal (original index 1)
Technology:        Average
Combat mode:       Tactical
Players:           exactly 2
Player 1 race:     Human
Player 2 race:     Darlok
Random events:     disabled for this slice
Antaran attacks:   disabled for this slice
Difficulty:        not yet gameplay-active; fixed baseline token only
Custom races:      unsupported
AI setup:          unsupported
```

Why Human + Darlok:

- both are normalized preset races;
- neither has a homeworld environmental modifier requiring Artifacts, Rich/Poor, Low/High-G, Large or Aquatic special state;
- their Homeworld physical normalization therefore fits current Core21 exactly;
- they still exercise two different governments/economy modifiers without expanding the New Game schema.

Random events are intentionally disabled even though the original exposes that setting: Slice 09 is about reproducible game creation and current event placement/execution is not complete. The setting can be added later without changing the core New Game architecture.

## Proven original behavior vs explicit Slice-09 MOOX normalization

### Directly proven / original-derived

- one shared New Game RNG ownership chain;
- player initialization precedes Universe/Homeworld generation;
- Small = 20 stars and 5x4 placement grid;
- Galaxy Age choices/order: Mineral Rich / Normal / Organic Rich;
- spectral-class weighted generation and Normal-age table;
- satellite-count, body-type, size, mineral, gravity and climate generation ordering;
- exact normalized size/mineral/gravity class indexes;
- Homeworld candidate logic seeks sufficiently separated stars and requires enough candidates for all players;
- Home system minimum 3 planets;
- Human/Darlok Homeworld physical state Medium/Abundant/Normal-G/Terran;
- starting Population total 8;
- starting Treasury 50 BC;
- starting Freighters 0;
- Average technology set;
- Average has two baseline Scouts plus one Colony Ship.

### Deliberate MOOX baseline choices

- current MOOX RNG algorithm instead of numeric original-LCG parity;
- `moox_grid_jitter_v1` coordinate system instead of original screen-scale rejection placement;
- farthest-pair Homeworld selection instead of original randomized five-scale candidate/shuffle implementation;
- uniform choice among currently unused orbit indexes because the original orbit-weight source is not yet fully normalized;
- only original body type 3 is materialized as Core `Planet`; non-colonizable bodies are transient/deferred;
- Homeworld underflow fills deterministic unused orbits rather than reproducing every internal `Enforce_3_Planet_Minimum_` conversion detail;
- 4 Farmer / 2 Worker / 2 Scientist continuous Population bootstrap;
- no free starting Buildings;
- simplified current-compatible unarmed Scout design rather than exact original auto-design loadout;
- no monsters, Orion, wormholes, nebulae, system specials, random events, Antarans or AI start logic.

These choices must appear in tests/documentation as named compatibility boundaries so later fidelity work can replace internals without rewriting the public New Game API.

## Gate-2 accepted implementation contract

Accepted by Gate 2 on **2026-09-02** after a current-code compatibility review and implemented by Gate 3 on the same date.

### 1. Rules/data

Add strict ruleset file:

`data/rulesets/moo2-1.31/new_game_galaxy.json`

Schema version **1**, containing provenance for:

- observed galaxy sizes/star counts/grid dimensions;
- three Galaxy Age IDs/order;
- 7x3 spectral weights;
- 10x6 satellite-count table;
- 10x5 body-type table;
- 10x6 mineral table;
- 5x5 gravity mineral/size table;
- Normal-age climate/group weights required by the selected fixture;
- homeworld normal defaults: Medium/Abundant/Normal-G/Terran, minimum 3 planets;
- original starting Population total 8, Treasury 50, Freighters 0, two Scouts + Colony Ship.

The file should also identify the deliberate MOOX algorithms `moox_grid_jitter_v1`, `farthest_pair_v1`, `uniform_unused_orbit_v1`, continuous-job `4/2/2`, no-start-building baseline and current-compatible Scout snapshot as implementation decisions, not original table values.

Load/validate it through the existing Economy/game rules aggregate. Economy schema remains **8**; Core remains **21**; existing planet-classes schema remains unchanged.

### 2. Game-layer settings/API

Add transport-independent types in `internal/game` without importing `internal/session`:

```text
GalaxySizeSmall = "small"
GalaxyAgeNormal = "normal"

NewGamePlayerSpec {
    SeatID      protocol.SeatID
    EmpireName  string
    RaceID      string
}

NewGameSettings {
    GalaxySize       GalaxySize
    GalaxyAge        GalaxyAge
    TechnologyLevel  NewGameTechnologyLevel
    StrategicCombat bool
    Players          []NewGamePlayerSpec
}

NewGamePlayerResult {
    SeatID    protocol.SeatID
    EmpireID  core.ID
    RaceID    string
    Name      string
}

NewGameResult {
    State    *core.GameState
    Players  []NewGamePlayerResult
}
```

Primary production entry point should be owned by the loaded rules/game layer, e.g.:

```text
func (r *EconomyRules) NewGame(seed uint64, settings NewGameSettings) (NewGameResult, error)
```

Exact validation for Slice 09:

- only `small`;
- only `normal` Galaxy Age;
- only Average technology;
- `StrategicCombat == false` (Tactical baseline);
- exactly two Players;
- Seat IDs non-zero, unique, strictly ascending;
- Race IDs exactly Human and Darlok, once each;
- Empire names non-empty and unique;
- no unsupported random-events/Antaran/custom-race/AI settings are silently accepted.

### 3. Deterministic ID allocation

Freeze ID creation order:

1. Galaxy ID;
2. Empire IDs in Player/Seat order;
3. 20 StarSystem IDs in row-major 5x4 order;
4. materialized Planet IDs in star ID then orbit order;
5. Homeworld Colony IDs in Player/Seat order;
6. Scout ShipDesign IDs in Player/Seat order;
7. two Scout Ship IDs + combat Fleet ID per Player in Player/Seat order;
8. Colony Ship special Fleet ID per Player.

No external/global state is mutated while constructing the candidate result.

### 4. Shared RNG sequence

Use one `rng := state.RNG()` for the complete generation and call `state.CommitRNG(rng)` exactly once after the candidate state is fully built and validated.

Freeze semantic phases:

1. player/race/default initialization using the same RNG owner;
2. Average research initialization;
3. star spectral generation and `moox_grid_jitter_v1` coordinates;
4. per-star body count/body/orbit/planet property generation in deterministic star order;
5. deterministic farthest-pair Homeworld selection;
6. Homeworld physical normalization/fill;
7. Colonies and starting strategic assets;
8. derived state/economy initialization needed for a valid turn-1 Session;
9. `game_started` Core event;
10. final `GameState.Validate` and one RNG commit.

Rejected settings or generation failure return an error with no externally visible partial game.

### 5. Galaxy/planet materialization

- always create exactly 20 StarSystems;
- transient star generation tracks spectral class but Core21 StarSystem does not persist it;
- coordinates use frozen `moox_grid_jitter_v1`;
- planet generation consumes the original-derived tables in the proven semantic ordering;
- only body type 3 becomes a `core.Planet`;
- Core Planet orbit indexes are unique 0..4 and stored sorted by orbit;
- all generated Planet size/mineral/gravity/climate IDs must cross-reference current `planet_classes.json`;
- every selected Home system ends with at least 3 materialized planets;
- selected Homeworld is the lowest-orbit materialized planet after fill and is forced Medium/Abundant/Normal-G/Terran;
- Homeworld itself satisfies the minimum food-world requirement for Human/Darlok.

### 6. Empires/Colonies/start assets

Per Player:

- Empire RaceID/name from settings;
- Treasury 50 BC;
- Freighters 0;
- full Tactical Average 20-tech set via existing New Game technology initializer;
- Homeworld Colony owned by Empire and set as Capital;
- one own-origin/own-loyalty Population cohort totaling 8 with MOOX jobs 4/2/2;
- no initial Buildings;
- current EconomyContext/derived state initialized through existing production helpers, not copied from fixture constants;
- one `Scout` design and two concrete Scout Ships using the accepted Slice-09 baseline snapshot;
- one Combat Fleet at the Home system containing both Scout Ships;
- one current special Colony Ship Fleet at the Home system;
- no Outpost, Population transfer or initial diplomatic hostility.

Initial Human/Darlok relation stays the current neutral/default relation representation until Slice 10 owns diplomacy/war lifecycle.

### 7. App/server integration

Replace fixture-only default startup with real game creation.

`internal/app` gains a New Game registration operation that:

1. invokes the game-layer generator;
2. builds `session.Seat` values from `NewGamePlayerResult`;
3. constructs/registers a `GameSession` with the existing resolver;
4. starts hosted `change_sequence` at 1;
5. rejects duplicate Game IDs atomically.

Add HTTP endpoint:

```text
POST /api/v1/games
```

Because JavaScript cannot safely represent arbitrary uint64 integers, the network request encodes `seed` as a **string** (decimal or `0x` hexadecimal). The internal game API still receives `uint64`.

Schema-1 request concept:

```text
{
  schema_version: 1,
  game_id: "game-1",
  seed: "0x8009",
  settings: {
    galaxy_size: "small",
    galaxy_age: "normal",
    technology_level: "average",
    strategic_combat: false,
    players: [
      {seat_id: 1, empire_name: "Human",  race_id: "human"},
      {seat_id: 2, empire_name: "Darlok", race_id: "darlok"}
    ]
  }
}
```

Response returns schema version, the normal hosted `GameSummary` and generated Seat/Empire identities. Existing HTTP error families remain; duplicate game or generation-settings conflict is a stable client-visible rejection without partial registration.

### 8. Standalone/web behavior

`cmd/moox-server` should no longer auto-create `core.NewSmallFixture` on normal startup. It starts with an empty Host and production New Game capability. An explicit `-demo-fixture` development option may retain the old fixture only for troubleshooting/tests; it must not be default production startup.

The React/Vite architecture-proof HMI gains a simple New Game form:

- Game ID;
- seed string;
- fixed/disabled Small, Normal, Average, Tactical selectors for this first baseline;
- editable Empire names;
- Human/Darlok preset assignment;
- create button -> `POST /api/v1/games`;
- on success refresh hosted games, choose Seat and enter the existing player snapshot/command flow.

### 9. Deterministic tests required by Gate 3

At minimum:

- same seed/settings -> exact DeepEqual State and generated Player result;
- same seed/settings after JSON save/load -> exact validated round trip;
- different seed -> different generated galaxy/RNG state while all invariants hold;
- exactly 20 stars, unique IDs/coordinates, valid ranges and 5x4 coordinate envelope;
- Homeworld pair is the deterministic farthest pair;
- each Home system >=3 planets;
- Human/Darlok Homeworlds exactly Medium/Abundant/Normal-G/Terran;
- each starting Population total8 and jobs4/2/2;
- Treasury50, Freighters0, exact 20 Average tech IDs;
- each Empire has 2 concrete Scouts in one Combat Fleet + one Colony Ship special fleet;
- no starting Buildings;
- malformed/unsupported settings fail before external registration;
- generator does not depend on `core.NewSmallFixture`;
- New Game-created State can construct `GameSession`, produce valid Player/Observer projections and execute at least one real turn;
- HTTP `POST /api/v1/games` followed by player snapshot and existing command path succeeds;
- duplicate Game ID is atomic;
- browser build remains green and can create a game from an initially empty server.

### 10. Explicitly deferred

- Medium/Large/Huge production support;
- Mineral Rich / Organic Rich production support;
- Pre-Warp / Advanced production New Game;
- Strategic Combat New Game;
- custom races and homeworld environmental trait permutations;
- exact original star screen-coordinate rejection algorithm;
- exact randomized Homeworld candidate/shuffle algorithm;
- exact original orbit-weight table mapping;
- gas giants/asteroids as persistent Core bodies;
- exact initial packed Population job layout;
- exact free starting-building matrix;
- exact original auto-designed Scout loadout/specials;
- Orion, monsters, wormholes, nebulae, system specials, splinter colonies, heroes;
- Random Events and Antaran start setup;
- AI start behavior;
- broad fog-of-war/knowledge expansion beyond the minimum player projection needed for the existing HMI.

## Gate-3 implementation result - 2026-09-02

Gate 3 implements the frozen New Game v1 contract in production paths:

- `data/rulesets/moo2-1.31/new_game_galaxy.json` schema 1 carries the observed galaxy sizes/ages, size-roll thresholds, spectral/satellite/body/mineral/gravity/planet-group tables, Normal-age climate weights, Homeworld/start defaults and named MOOX compatibility algorithms.
- Additional direct executable recheck fixed the Normal climate path to object-2 `0x57A7`; `Generate_Climate_` uses the 4 group columns from this table for galaxy ages 0/1, while age 2 uses `0x57CF`. `Get_Planet_Group_` indexes the 6x5 spectral/orbit table at object-2 `0x5761`, and `Generate_Size_` uses thresholds `1,3,7,9,10` at `0x57F7`.
- `EconomyRules.NewGame` owns `core.NewGameState(seed)` plus one shared RNG, freezes ID allocation/generation order, creates exactly 20 Small systems, materializes original-derived colonizable planets, selects/fills/normalizes two deterministic Home systems, creates Human+Darlok start state/assets and commits RNG once after final validation.
- Seed `0x8009` with the canonical settings is pinned to full-state JSON SHA-256 `1d89bf9e8a5ce47c81a481ad669916d727357587dfc40e46904b9503ba314a89`; same seed/settings DeepEqual, JSON save/load round-trip and different-seed divergence are regression-tested.
- Generated state constructs a real two-seat `GameSession`, projects player/observer views and completes a real Turn 1 -> Turn 2 economy/session cycle.
- `internal/app` atomically registers generated sessions with two `ControllerLocalHuman` seats and stable `ErrGameExists`; `POST /api/v1/games` accepts decimal/`0x` seed strings, returns 201 and maps invalid input to 400 plus duplicate IDs to 409 `game_exists`.
- normal `cmd/moox-server` startup is empty and New-Game-capable; legacy `core.NewSmallFixture` bootstrap is development-only behind `-demo-fixture`.
- the React/Vite HMI now exposes a server-authoritative New Game form for the frozen Small/Normal/Average/Tactical Human+Darlok fixture and enters the existing snapshot/command flow after creation.
- final Gate-3 regression: `go test ./... -count=1`, `go vet ./...`, `npm run build` and `git diff --check` all pass. The only diff-check diagnostics are existing LF/CRLF conversion warnings.
## Recovery re-verification - 2026-09-02

After the prior chat/session stalled, the remaining Galaxy-Age uncertainty was rechecked directly against the extracted original HELP text and the already captured executable branch evidence. HELP offset `0xBA564` labels **Galaxy Age** and the following record states the cycle order **mineral rich, normal, organic rich**. Combined with the executable `_g_galaxy_age` three-way `0/1/2` handling used by spectral/climate generation, the production-baseline mapping is therefore confirmed as `0=mineral_rich`, `1=normal`, `2=organic_rich`; Slice 09 selects `normal` / index `1`.

This recovery pass found no additional original-fact blocker for Gate 1. The remaining coordinate, Homeworld selection, orbit, population-job, free-building and Scout-loadout differences are already explicitly versioned as MOOX simplifications/deferred fidelity work rather than unresolved evidence claims.
## Gate-1 conclusion

Gate 1 now has enough original evidence to implement a narrow real New Game path without presenting MOOX compatibility choices as reverse-engineered facts. The strongest original guarantees for the selected fixture are Small=20/5x4, one shared New Game RNG lifecycle, original-derived planet property ordering/tables, separated Homeworld intent, Human/Darlok Medium-Abundant-NormalG-Terran Homeworlds with at least three planets, Population8, Treasury50, Freighters0, Average technologies and two Scouts plus Colony Ship.

The Gate-2 contract above deliberately versions the remaining coordinate/Homeworld/orbit/job/building/Scout choices as MOOX baselines. **Gate 3 implemented that frozen contract without widening Slice 09; Gate 4 owns final QA, commit and closure.**

## Gate-4 QA result - 2026-09-02

Gate 4 repeated the Slice-09 proof independently before closure:

- canonical New Game/golden tests were repeated multiple times; seed `0x8009` retains full-state SHA-256 `1d89bf9e8a5ce47c81a481ad669916d727357587dfc40e46904b9503ba314a89`;
- equal seed/settings now assert both semantic `DeepEqual` and byte-identical JSON state; different-seed divergence remains green;
- the generated two-seat state passes Core validation, Player/Observer projection and a real strategic Turn 1 -> Turn 2 session cycle;
- frontend authority scan found no browser-side RNG or `GameState` implementation; New Game creation calls the server and gameplay remains revision-bound CommandBatches;
- standalone runtime proof started the production server with `games=0`, served the built HMI, and managed Chrome created default `game-1` from `0x8009`; the browser then displayed the server snapshot at Turn 1 / planning, Human Empire ID 2, Seat 1 `local_human`, Population total 8;
- final changed/new Go files were gofmt'd; `go test ./... -count=1`, `go vet ./...`, `npm run build` and `git diff --check` all pass. Diff-check diagnostics are LF/CRLF conversion notices only.
## Slice closure - 2026-09-02

Slice 09 is closed. Implementation/data/evidence commit: `a5f3c13` (`game: add deterministic new game galaxy baseline`). Gate 4 completed the repeated golden/determinism/session/server/browser authority QA and the permanent status/HISTORY handoff. The next prepared objective is Slice 10 diplomacy / war / peace; no later implementation slice is opened by this closure.
### Core22 serialization addendum - Slice 10 Gate 3, 2026-09-03

Slice 09 closed under Core schema 21 with the historical seed `0x8009` full-state SHA-256 `1d89bf9e8a5ce47c81a481ad669916d727357587dfc40e46904b9503ba314a89`. Slice 10 Gate 3 advances the authoritative state schema to 22 for diplomacy. The same canonical seed/settings now serialize to SHA-256 `8effff679109cd10a427f1c80b83047dc0d2fa8e5e4871a67235a486dbc2fc68`; this fingerprint change is schema/state-shape driven, while the deterministic New Game generation contract remains unchanged and continues to be regression-tested.

### Core23 serialization addendum - Slice 11 Gate 3, 2026-09-03

Slice 09 closed under Core schema 21 with historical seed `0x8009` SHA-256 `1d89bf9e8a5ce47c81a481ad669916d727357587dfc40e46904b9503ba314a89`. Slice 10 then advanced to Core22 for diplomacy, producing the historical Core22 fingerprint `8effff679109cd10a427f1c80b83047dc0d2fa8e5e4871a67235a486dbc2fc68` for the same canonical seed/settings.

Slice 11 Gate 3 advances the authoritative state schema to **23** for persisted Colony standing Infantry plus the Troop Transport/conquest state vocabulary. The same canonical New Game now serializes to SHA-256 `83614740b216409877b03c536a16328c7fa7031867f58dbeb70857a8953f5762`. New Game still starts with zero standing Infantry and no Troop Transports; the fingerprint change is therefore a schema-version/state-shape serialization change, not a change to the Slice-09 galaxy/start-generation rules. The Core23 fingerprint is pinned in `internal/game/new_game_test.go` and the full repository regression suite is green.
