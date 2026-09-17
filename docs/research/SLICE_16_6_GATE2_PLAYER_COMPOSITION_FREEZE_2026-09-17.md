# Slice 16.6 Gate 2 - Player composition and final New Game integration freeze

Date: 2026-09-17

Status: **Gate 2 complete; Gate 3 implementation next**.

Gate-1 basis: `docs/research/SLICE_16_6_GATE1_PLAYER_COMPOSITION_AUDIT_2026-09-17.md`.

## Objective

Freeze the first production player-composition contract for Slice 16.6 before any runtime breadth is enabled. This document is binding for Gate 3 unless a newly discovered hard contradiction is recorded explicitly.

The freeze distinguishes three different boundaries:

1. **original-facing range** - the documented Master of Orion II range of 1-7 opponents;
2. **Slice-16.6 V1 supported range** - the subset MOOX can honestly create/play now;
3. **planned range** - original-facing counts that remain visible but disabled until race/runtime breadth exists.

## 1. Opponent-count freeze

### Original-facing selector range

The visual selector exposes exactly:

```text
1, 2, 3, 4, 5, 6, 7 opponents
```

This is the original documented 2-8 total-player range expressed from the local player's point of view.

Counts 8-9 are **not** part of Slice-16 original parity. A future MOOX extension may add them only as an explicitly versioned product extension.

### Slice-16.6 V1 support

| Opponents | Total players | V1 availability | Reason |
| ---: | ---: | --- | --- |
| 1 | 2 | `supported` | existing production class |
| 2 | 3 | `supported` | existing three-seat/two-AI controller evidence plus bounded race set |
| 3 | 4 | `planned` | requires wider honest AI race/runtime support |
| 4 | 5 | `planned` | requires wider honest AI race/runtime support |
| 5 | 6 | `planned` | requires wider honest AI race/runtime support |
| 6 | 7 | `planned` | requires wider honest AI race/runtime support |
| 7 | 8 | `planned` | requires wider honest AI race/runtime support |

Binding disabled reason ID for counts 3-7:

```text
additional_ai_race_breadth_required
```

The UI localizes that server-owned reason ID. It must not invent a different support explanation from frontend-only logic.

## 2. Galaxy-size x opponent-count matrix

All four currently supported galaxy sizes accept both V1 opponent counts.

| Galaxy size | Stars | 1 opponent | 2 opponents | 3-7 opponents |
| --- | ---: | --- | --- | --- |
| `small` | 20 | supported | supported | planned |
| `medium` | 36 | supported | supported | planned |
| `large` | 54 | supported | supported | planned |
| `huge` | 71 | supported | supported | planned |

Why there is no smaller-size V1 cap:

- the original setup evidence does not establish a separate size-dependent hard player maximum;
- even Small materializes 20 star systems, far above the V1 maximum of three home systems;
- current homeworld normalization can force the required minimum colonizable planets at a chosen home system;
- the actual current blockers for 4-8 total players are wider race honesty and wider lifecycle evidence, not raw star count.

Gate 3 must expose this matrix from the server. The frontend must not derive capacity from `star_count`.

## 3. Deterministic N-player home-star normalization

Production `NewGame` currently calls `farthestNewGameSystemPair`, so simply changing the player-count validator would be invalid.

Gate 3 must replace the pair-only call site with a deterministic selector conceptually equivalent to:

```go
selectNewGameHomeSystems(systems, playerCount) []int
```

### Required behavior

#### 3.1 Input boundary

For Slice-16.6 V1:

```text
playerCount = 2 or 3
```

The helper may be implemented generically for larger N, but production validation remains capped at three total players until a later contract enables more.

#### 3.2 First two homes - exact backward compatibility

The first two selected systems must use the **exact current `farthestNewGameSystemPair` semantics**, including:

- squared Cartesian distance;
- maximum distance wins;
- existing System-ID tie break;
- existing returned pair order.

Consequences:

- every accepted two-player seed/settings tuple keeps the exact same home-system pair as before Slice 16.6;
- seat 1 and seat 2 keep their current home assignment order;
- two-player topology/golden fingerprints must not move merely because the helper becomes N-player-capable.

#### 3.3 Additional homes - deterministic maximin

For each additional home after the first pair:

1. consider every unselected system;
2. compute its squared distance to every already selected home;
3. take that candidate's **minimum** squared distance to the selected set;
4. choose the candidate with the **largest minimum distance**;
5. on equal minimum distance, choose the lower `System.ID`;
6. if IDs are somehow also equal, choose the lower source-slice index as a final defensive tie break;
7. append the chosen index to the selected-home list.

For Slice-16.6 V1 this loop runs at most once, selecting seat 3's home.

#### 3.4 Seat assignment

Selected home indices map to `settings.Players` order:

```text
selected[0] -> first player spec / seat 1
selected[1] -> second player spec / seat 2
selected[2] -> third player spec / seat 3
```

Seat IDs remain non-zero, unique and strictly ascending.

#### 3.5 RNG boundary

The home-system **selection** algorithm consumes **zero additional RNG draws**.

This is deliberate:

- two-player RNG draw order remains unchanged;
- three-player placement is deterministic from the already-generated galaxy geometry;
- later per-home normalization may consume its existing RNG in seat/home order, including the new third home as real additional content.

Gate 3 must add a regression proving two-player equal-seed state remains unchanged relative to the pre-16.6 baseline.

## 4. Local-player race and opponent assignment freeze

### Local player

Supported local preset races remain exactly:

```text
human
klackon
```

Seat 1 remains the single `local_human` controller for Slice-16.6 V1.

Remote-human, hotseat and multiple-human composition are outside this slice.

### Automatic opponent composition

Slice-16.6 V1 freezes **automatic deterministic composition**, not per-slot manual race editing.

| Local race | Opponent count | Seat 2 | Seat 3 |
| --- | ---: | --- | --- |
| `human` | 1 | `darlok` | - |
| `human` | 2 | `darlok` | `klackon` |
| `klackon` | 1 | `darlok` | - |
| `klackon` | 2 | `darlok` | `human` |

For every opponent seat:

```text
controller = builtin_ai
```

Binding reasons:

- preserves the current Darlok baseline for the one-opponent class;
- uses only the Human/Klackon mechanics already accepted by Slice 16.4 for the added opponent;
- maps to the three-seat/two-AI controller class already exercised in repository regressions;
- avoids pretending the other planned preset races have complete current runtime parity.

The opponent cards are informative in V1. They are not race pickers.

## 5. Duplicate-race freeze

Preset race IDs remain unique across all players in a New Game.

Gate 3 must preserve the existing duplicate-race rejection.

The automatic V1 composition above always produces unique race IDs.

No frontend option may create duplicate predefined races and rely on the server to silently repair them.

## 6. Random-opponent freeze

There is **no `random` opponent race option in Slice-16.6 V1**.

Prohibited in Gate 3:

- `Math.random()` or browser-side shuffling;
- client-selected random race IDs;
- hidden server RNG draws whose candidate order is not versioned;
- silently replacing an invalid requested race with another race.

A future seeded random mode requires its own server-owned candidate set/order, duplicate policy and exact New Game RNG/draw-order contract.

## 7. Opponent empire-name boundary

Race/controller composition is authoritative; opponent display names are not gameplay selection controls in V1.

The production HMI should remove the dedicated editable `Darlok empire` field and show compact opponent cards instead.

For generated frontend requests, each server composition assignment supplies a canonical `default_empire_name`:

```text
darlok  -> Darlok
human   -> Human
klackon -> Klackon
```

The server keeps the generic existing invariant that all submitted Empire names are non-empty and unique. It need not reject an otherwise valid API request solely because an opponent name differs from the catalog default; race/controller/seat composition is the authoritative part.

The local player's Empire name remains editable and race-neutral.
## 8. Server-owned composition catalog freeze

Gate 3 adds one authoritative catalog endpoint alongside the existing difficulty/galaxy/race/technology catalogs:

```text
GET /api/v1/new-game/compositions
```

The app layer exposes the same contract from the loaded rules/runtime support boundary. The Web client fetches it; it does not duplicate the matrix in constants.

### Required response shape

Semantically equivalent Go/JSON shape:

```text
NewGameCompositionCatalog {
    default_opponent_count: 1
    original_opponent_min: 1
    original_opponent_max: 7
    counts: []OpponentCountProfile
    galaxy_limits: []OpponentGalaxyLimit
    assignments: []OpponentAssignment
}

OpponentCountProfile {
    opponent_count: int
    availability: "supported" | "planned"
    reason_id?: string
    facts: []string
}

OpponentGalaxyLimit {
    galaxy_size: "small" | "medium" | "large" | "huge"
    max_supported_opponents: 2
}

OpponentAssignment {
    player_race_id: "human" | "klackon"
    opponent_count: 1 | 2
    opponents: []OpponentSpec
}

OpponentSpec {
    race_id: "darlok" | "human" | "klackon"
    controller: "builtin_ai"
    default_empire_name: string
}
```

The HTTP wrapper follows existing New Game catalog conventions and includes the current API `schema_version`.

### Frozen count facts

Stable server-owned facts should remain concise:

- count 1: `1 built-in AI opponent`, `Darlok baseline`;
- count 2: `2 built-in AI opponents`, `Darlok + supported alternate race`;
- counts 3-7: `Original-range count`, `Additional AI race breadth required`.

The frontend may localize labels but must not create gameplay claims not present in the catalog.

### Frozen assignment entries

The catalog must resolve exactly these four supported assignments:

```text
human  + 1 -> [darlok]
human  + 2 -> [darlok, klackon]
klackon + 1 -> [darlok]
klackon + 2 -> [darlok, human]
```

All assignment entries use `builtin_ai` opponents.

`galaxy_limits` is deliberately explicit even though every V1 galaxy currently has the same maximum. That keeps future capacity changes server-owned instead of teaching the frontend to infer them from star counts.

## 9. Authoritative create validation freeze

The composition catalog is informative; the create path remains authoritative.

Gate 3 must reject a request unless the complete effective tuple satisfies all of the following.

### Base setting validation

- difficulty ID is one of all five supported Slice-16.2 profiles;
- galaxy size is `small|medium|large|huge`;
- galaxy age is one of all three supported ages;
- technology level is `pre_warp|average`;
- `advanced` remains rejected until Slice 16.7;
- `strategic_combat == false` remains mandatory;
- seed/settings remain under the existing deterministic New Game contract.

### Player/composition validation

- player count is exactly 2 or 3;
- player specs remain ordered by strictly ascending, non-zero, unique seat IDs;
- first player spec is the local-player role and uses `human|klackon`;
- first player is not `BuiltinAIControlled` after controller materialization;
- every remaining player is `BuiltinAIControlled`;
- second player race is always `darlok`;
- if there is a third player, its race is exactly the other supported Human/Klackon race relative to the first player;
- race IDs are unique;
- Empire names remain non-empty and case-insensitively unique;
- every race must resolve through loaded rules/race modifiers;
- no planned preset race is accepted simply because it exists in the browse catalog.

### Controller validation at app/server boundary

For the canonical HMI/API tuple:

```text
player 0 -> local_human
player 1 -> builtin_ai
player 2 -> builtin_ai (only when opponent_count == 2)
```

The server must reject attempts to turn an opponent into `local_human` or `remote_human` inside this Slice-16.6 V1 composition flow.

Existing generic session/controller types remain valid elsewhere; this is specifically the New Game product contract for Slice 16.6.

### No silent repair

The server and frontend must never silently:

- reduce an unsupported opponent count;
- change galaxy size;
- substitute a race;
- drop a duplicate;
- convert an opponent controller;
- fall back from `advanced` to `average`;
- randomize a replacement.

The request is rejected instead.

## 10. Cross-setting freeze

For Slice-16.6 V1 there are no additional hidden compatibility exclusions among currently supported choices.

The complete accepted dimensions are:

```text
difficulty:       easy | normal | hard | very_hard | impossible
galaxy size:      small | medium | large | huge
galaxy age:       mineral_rich | normal | organic_rich
player race:      human | klackon
technology:       pre_warp | average
opponents:        1 | 2
combat:           tactical only (strategic_combat=false)
opponent control: builtin_ai only
```

Therefore every Cartesian tuple across those accepted values must either create successfully or expose a real newly discovered blocker before Gate 3 is considered complete.

No difficulty-specific opponent-count restriction is frozen: the difficulty system applies its AI effects per `BuiltinAIControlled` Empire.

No galaxy-age-specific composition restriction is frozen.

## 11. Count-selector HMI freeze

### Visual family

Gate 3 produces seven production numeral assets:

```text
web/public/assets/new-game/opponent-count/1.svg
...
web/public/assets/new-game/opponent-count/7.svg
```

Binding asset direction:

- original MOOX visual language, not copied original-game artwork;
- same image-led selector grammar as Slices 16.1-16.5;
- starfield/empire motifs may increase subtly with count;
- numerals are geometric/path-based rather than font-dependent runtime `<text>`;
- scalable SVG, deterministic/reproducible generator preferred;
- registered in the New Game asset manifest;
- the selected count remains understandable without relying only on color.

### Selector behavior

- default = `1` opponent;
- arrows/keyboard browse all 1-7 profiles;
- count 1/2 display `supported now`;
- count 3-7 display `planned` plus localized `additional_ai_race_breadth_required` reason;
- selecting a planned count may show its card/details but disables Create Game;
- returning to a supported count immediately resolves the appropriate opponent assignment from the server catalog;
- no numeric free-text input is exposed.

### Opponent cards

Under the main count selector, render one compact card/chip per resolved opponent:

- Slice-16.4 race portrait;
- race name;
- `AI` controller label;
- no per-slot race selector in V1;
- no opponent empire-name editor in V1.

Desktop may render chips/cards in one row when space permits. At 390 px they wrap naturally below the count card without shrinking the portrait below readable recognition size.

## 12. Disabled reason / mobile freeze

The generic selector availability state is extended so composition options can display a server reason ID.

For a planned count at 390 px:

1. the count remains visible/browseable;
2. availability is clearly `planned`/disabled;
3. one concise localized reason is immediately associated with the selected card;
4. Create Game is disabled;
5. opponent cards are not fabricated for an unsupported count;
6. no horizontal page overflow is introduced;
7. previous/next tap targets retain at least the existing shared selector touch target size.

The same semantic reason must be represented by the server catalog and create-path rejection. Exact localized wording may differ by locale, but the reason ID is authoritative.

## 13. Final New Game launch-summary freeze

Before the primary Create Game button, Gate 3 adds a compact **launch briefing** assembled entirely from the currently selected authoritative catalog entries.

### Desktop order

```text
[ Galaxy ] [ Difficulty ] [ Player Race ]
[ Start Tech ] [ Opponents ] [ Seed ]
                 [ Create Game ]
```

### Required summary content

1. **Galaxy**
   - selected galaxy-size artwork;
   - localized size title;
   - localized galaxy-age title.
2. **Difficulty**
   - selected difficulty emblem;
   - localized difficulty title.
3. **Player Empire**
   - selected Human/Klackon portrait;
   - local editable Empire name;
   - localized race title.
4. **Starting Technology**
   - selected Pre-Warp/Average art;
   - localized title.
5. **Opponents**
   - selected numeral art/count;
   - compact resolved opponent race portraits/names;
   - Built-in AI fact, not an editable controller.
6. **Seed**
   - exact seed text/value that will be submitted.

Tactical Combat remains a fixed product fact, not a selector. It may appear as a small secondary summary fact but must not look editable.

### 390 px behavior

At 390 px the briefing becomes a single-column/compact stacked layout in the same semantic order. Opponent portrait chips wrap within the Opponents summary card. The seed remains copy/readable and must not force horizontal overflow.

The summary is not a second source of state. It renders the same selected IDs/values that the request builder consumes.

## 14. Gate-3 implementation boundary

Gate 3 owns exactly:

1. add the server-owned composition catalog and API route;
2. implement deterministic N-player home selection with exact two-player backward compatibility;
3. widen New Game validation from exactly two players to the frozen 2/3-player classes only;
4. enforce the frozen automatic race/controller composition;
5. add seven production opponent-count SVGs + manifest/generator checks;
6. add the shared count selector, resolved opponent cards and disabled reasons;
7. remove the dedicated editable Darlok-opponent name control from the production HMI;
8. add the launch briefing;
9. submit the exact server-resolved supported composition;
10. add deterministic game/app/server/browser fixtures.

Gate 3 does **not** own:

- enabling opponents 3-7;
- additional AI race parity work merely to fill those slots;
- duplicate predefined races;
- random race assignment;
- remote/hotseat/multiple-human setup;
- Advanced Starting Technology;
- Random Events or Antaran Attacks;
- strategic-combat selection.

## 15. Determinism and acceptance fixture freeze

### Game-layer exhaustive creation matrix

Gate 3 must exercise every accepted settings class at the game layer across:

```text
4 galaxy sizes
x 3 galaxy ages
x 5 difficulties
x 2 player races
x 2 supported technology levels
x 2 opponent counts
= 480 accepted tuples
```

Use one fixed canonical seed for the exhaustive matrix. For each tuple:

- create twice from fresh rules/state;
- both creations must succeed;
- both resulting state/player results must be deeply deterministic/equivalent;
- `State.Validate()` must pass;
- empire/player/home counts must equal the frozen composition;
- controller-sensitive difficulty flags must identify every AI Empire correctly.

This matrix is intentionally game-layer/lightweight; it prevents hidden unsupported cross-setting combinations from slipping into V1.

### Two-player non-regression fixture

Retain the existing canonical two-player golden/fingerprint regression and add a specific assertion that introducing the N-player selector did not change the pre-16.6 two-player result for the frozen canonical seed/settings.

### Three-player topology fixture

For at least one Small and one Huge seed:

- first two home systems equal the old farthest-pair result;
- third home is the frozen maximin choice;
- repeated selection is byte/value deterministic;
- all three home indexes are unique;
- each Empire receives exactly its selected home in player-spec order.

### App/server composition matrix

At minimum cover all four automatic assignments through public create handling:

```text
human  + 1
human  + 2
klackon + 1
klackon + 2
```

and representative Small/Huge, Pre-Warp/Average and Easy/Impossible combinations.

Verify:

- controller projection;
- race IDs;
- player counts;
- snapshot/player views;
- Create Game response;
- rejection of count 3+ / wrong race / duplicate / wrong controller / Advanced.

### Save/Resume smoke

Before Gate 4, at least one 2-opponent/3-player game for each supported technology start must:

- create;
- enter normal browser/session play;
- persist/save;
- restore/resume;
- preserve all seats, races, controllers, Empire ownership and deterministic state.

## 16. Gate-4 acceptance handoff

Gate 4 must verify, not redesign, this freeze:

- desktop and real 390 px New Game walkthrough;
- supported counts 1/2 and planned count reason presentation;
- opponent card wrapping/readability;
- launch briefing correctness;
- representative create/play for every difficulty/galaxy/tech/player-count class;
- representative Human/Klackon and both two-opponent assignments;
- equal seed + complete settings tuple determinism;
- no unsupported composition through UI or API;
- full Go tests + vet;
- Web production build and selector/browser checks;
- `git diff --check`;
- parent/status/history closeout and handoff to Slice 16.7 Advanced parity.

## 17. Gate-2 decision summary

Slice-16.6 V1 is now frozen as:

```text
Original-facing opponent selector: 1-7
Supported now:                    1-2
Local humans:                     exactly 1
Builtin-AI opponents:             1-2
Local race:                       Human or Klackon
Opponent 1:                       Darlok
Opponent 2:                       other Human/Klackon race
Duplicate preset races:           no
Random opponents:                 no
Galaxy sizes:                     all four support counts 1-2
Galaxy ages:                      all three
Difficulty:                       all five
Technology:                       Pre-Warp or Average
Advanced:                         Slice 16.7
Combat selector:                  none; Tactical fixed
Home placement:                   farthest pair + deterministic maximin extension
Composition authority:            server catalog + server validation
```

Gate 3 must implement exactly this contract without broadening it opportunistically.