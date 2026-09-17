# Slice 16.6 Gate 1 - Player composition and final New Game integration audit

Date: 2026-09-17

Status: **Gate 1 complete; Gate 2 freeze next**.

## Objective

Audit the real player-count, multi-AI, race-assignment, deterministic-randomization and HMI boundaries before Slice 16.6 widens production New Game. Gate 1 is evidence/prototype work only: it does not widen the server validator or silently enable unsupported combinations.

## Sources inspected

Repository/runtime:

- `internal/game/new_game.go`
- `internal/game/reference_triangle.go`
- `internal/game/race_catalog.go`
- `internal/game/difficulty.go`
- `internal/app/new_game.go`
- `internal/app/host.go`
- `internal/app/builtin_ai_test.go`
- `web/src/App.tsx`
- `web/src/api.ts`
- `web/src/components/VisualSelector.tsx`
- Slice 16.3/16.4/16.5 freeze and closure evidence.

Original/reference evidence:

- `reference/original/support/files/Orion2.exe` 1.31, including embedded Watcom symbols such as `Init_New_Game_`, `Init_Players_`, `Players_Home_Star_`, `Galaxy_Size_From_N_Stars_`, `_NUM_PLAYERS`, `_g_galaxy_size`, `N_Living_Players_`, and Hotseat/multiplayer helpers;
- `docs/research/NEW_GAME_GALAXY_GENERATION_BASELINE_2026-09-02.md`, especially the direct `Generate_Home_Worlds_` analysis;
- original Master of Orion II manual, *Beginning a New Game / Number of Players*, externally cross-checked on 2026-09-17;
- StrategyWiki race-selection description as a secondary cross-check only for duplicate-race behavior.

Research-only visual prototype:

- `docs/research/prototypes/SLICE_16_6_OPPONENT_COMPOSITION_2026-09-17.svg`

## 1. Original player-count boundary

The original manual explicitly allows **2 through 8 total players**, described as the human player plus **1 through 7 opponents**.

Therefore the prior Slice-16.6 planning target of **1 through 9 opponents** is not original-parity evidence. Gate 1 corrects the parity target to **1 through 7 opponents**. Any future 8-9 opponent mode would be a deliberate MOOX extension and must not be presented as original behavior.

The manual frames player count as an independent Game Setup setting. Gate 1 did not find a manual rule that directly assigns a different hard maximum player count to Small/Medium/Large/Huge. The manual instead states qualitatively that more players make resources relatively scarcer and contact earlier.

This is distinct from implementation feasibility: MOOX may still need a conservative size/count support matrix because its normalized home-star allocator and race/runtime breadth are not yet at original full range.

## 2. Original home-star dependency

Existing direct executable research already establishes that original `Generate_Home_Worlds_`:

- builds minimum-star-distance information;
- builds a home-star candidate list;
- requires that candidate list to contain at least `_NUM_PLAYERS` stars;
- retries through up to five progressively relaxed separation scales;
- consumes the shared New Game RNG for candidate/tie selection and shuffling;
- calls `Randomize_Home_Worlds_` so selected home stars are randomized across players.

The important Gate-1 conclusion is that the original algorithm is **N-player aware**. It is not a two-star special case.

The exact original distance thresholds remain outside the current normalized MOOX contract, so Gate 2 must not invent a numerical galaxy-size/player-count cap and label it original fact.

## 3. Current production MOOX boundary

`internal/game/new_game.go` currently requires:

- exactly **2** players;
- non-zero, unique, strictly ascending seat IDs;
- unique race IDs;
- exactly one fixed `darlok` opponent;
- exactly one supported local-player race from `{human, klackon}`;
- every submitted race must exist in loaded `RaceModifiers`.

The normal generator is also structurally two-player-only:

- it calls `farthestNewGameSystemPair(...)`;
- that helper returns exactly `[2]int`;
- `initializeNewGameStartingAssets` requires one home index per Empire.

Therefore widening only the validation count would be incorrect. A production 3+ player start needs an N-player home-star allocator first.

## 4. Galaxy sizes and current capacity metadata

The current server-owned galaxy catalog exposes:

| Galaxy size | Stars |
| --- | ---: |
| Small | 20 |
| Medium | 36 |
| Large | 54 |
| Huge | 71 |

There is currently **no** authoritative `max_players` / `max_opponents` field in the galaxy catalog.

Slice 16.3 deliberately deferred player-capacity rules to Slice 16.6. Gate 2 therefore owns the decision whether the first supported composition matrix is identical across all four galaxy sizes or is deliberately narrower on smaller sizes.

No size-dependent cap should be placed in frontend constants before that server contract exists.

## 5. Core/session multi-player capability

The two-player production validator is narrower than the underlying state/session model.

`NewReferenceTriangleGame` already creates three Empires/players: Human, Darlok and Psilon. This proves that Core state, seats, diplomacy/contact and strategic systems are not intrinsically limited to two Empires.

The reference triangle is a QA scenario, not proof that generic production New Game can create arbitrary player counts. Its three home systems are explicitly supplied rather than generated by the normal pair allocator.

## 6. Built-in AI support for more than one opponent

The host automation drivers iterate the complete `g.seats` collection for:

- planning;
- post-resolution decisions;
- encounter/battle control;
- invasion decisions.

Controller behavior is seat-based rather than hard-coded to seat 2.

More importantly, existing regression `TestTriangleLaserScoutQueueDoesNotStallStrategicResolution` constructs the three-player reference triangle and assigns:

- seat 1 -> `local_human`;
- seats 2 and 3 -> `builtin_ai`.

The targeted test passes on the current HEAD.

This is direct evidence for the **three-total-player / two-AI-opponent controller class**. It is not yet evidence for 4-8 total seats through the complete production lifecycle. Gate 2 must distinguish those claims.

Difficulty effects are also applied per Empire through `Empire.BuiltinAIControlled`, so the difficulty model is structurally compatible with multiple AI Empires. Wider multi-AI regressions are still required before claiming full 8-player production support.

## 7. Race-support boundary

The server catalog contains all 13 preset races for browsing, but local-player runtime availability is currently only:

- `human`;
- `klackon`.

Darlok remains the fixed/grandfathered AI baseline even though its defining Espionage/Stealth breadth is incomplete. Slice 16.4 explicitly did **not** claim full Darlok player parity.

The Slice-16.4 readiness audit classifies the remaining races as blocked or requiring bounded fixes. Consequently, player count and race breadth cannot be treated as independent switches.

### Conservative Gate-2 candidate

Without importing new race-mechanics work, the strongest evidence-backed first composition candidate is:

| Local player race | 1 opponent | 2 opponents |
| --- | --- | --- |
| Human | Darlok | Darlok + Klackon |
| Klackon | Darlok | Darlok + Human |

All opponents are `builtin_ai`.

Why this candidate is strong:

- preserves the accepted Darlok baseline;
- uses Human/Klackon mechanics already accepted by Slice 16.4;
- keeps race IDs unique;
- maps exactly to the existing three-player/two-AI session evidence;
- needs no fake support claim for the other 10 planned preset races.

This is a **Gate-1 candidate**, not yet the Gate-2 freeze.

### Why 3+ opponents are not ready to freeze automatically

With unique preset races and the current honest support boundary, after the local Human/Klackon seat only two acceptable opponent races remain: Darlok plus the other supported race.

Supporting 3-7 opponents therefore requires at least one explicit Gate-2 policy expansion:

1. enable additional preset races for AI despite known mechanics gaps;
2. close enough bounded race-mechanics gaps first;
3. allow duplicate preset races;
4. defer those counts until later.

Gate 1 recommends **not** solving this by frontend-only duplication or by silently treating `planned` races as supported AI races.

## 8. Duplicate-race policy

Current MOOX New Game rejects duplicate race IDs.

A secondary original-game cross-check (StrategyWiki) states that choosing a predefined race prevents an enemy Empire from using the same predefined race. That is directionally consistent with the current unique-race validator, but Gate 1 does not promote that secondary source to direct executable proof for every multiplayer/custom-race case.

Candidate for Gate 2: retain **unique preset race IDs** for the first Slice-16.6 production matrix. If MOOX ever allows duplicates as a product extension, it should be a deliberate versioned contract change.

## 9. Opponent race assignment and randomization candidates

The original single-player setup exposes a total player count before the player selects their own race; it does not present the Slice-16.6-style per-opponent race editor in that setup flow. Therefore per-slot opponent selection is a MOOX product feature, not something Gate 1 should describe as original UI parity.

Gate 2 should choose explicitly among these candidates:

### Candidate A - deterministic automatic composition (recommended first freeze)

- 1 opponent -> Darlok;
- 2 opponents -> Darlok plus the other supported Human/Klackon race;
- no extra RNG draw for race assignment;
- compact opponent cards are informative rather than editable.

This has the smallest deterministic surface and preserves all current accepted behavior.

### Candidate B - explicit supported per-slot selection

- server returns allowed race IDs per opponent slot;
- HMI uses Slice-16.4 portraits;
- duplicates rejected;
- no random option initially.

This offers more agency but needs a server composition catalog and cross-slot validation.

### Candidate C - seeded server-owned random opponents

A future `random` choice is acceptable only if:

- the candidate race set is server-owned and versioned;
- duplicate policy is explicit;
- stable canonical candidate ordering is frozen;
- the shared New Game RNG owner/draw order is frozen;
- equal complete settings + seed reproduce identical race assignments.

Frontend `Math.random()` or client-side shuffling is prohibited.

Gate 1 recommends deferring Candidate C until the accepted AI race set is wider than two races.

## 10. Current API/HMI gap

The API model is already array-shaped:

- `settings.players[]`;
- optional `controllers[]`.

So the wire shape itself is not intrinsically two-seat-only.

The React New Game screen is concrete/hard-coded, however:

- exactly seat 1 `local_human`;
- exactly seat 2 `builtin_ai`;
- seat 2 race hard-coded to `darlok`;
- a dedicated `Darlok empire` name field;
- no opponent-count selector;
- no opponent composition cards;
- no final launch-summary card;
- Create Game has no composition-catalog validation yet.

Slice 16.6 must remove those concrete assumptions only after Gate 2 freezes the authoritative composition contract.

## 11. Visual prototype

Research prototype:

`docs/research/prototypes/SLICE_16_6_OPPONENT_COMPOSITION_2026-09-17.svg`

It deliberately shows:

- symbolic number tiles for **1-7 opponents**, matching the original documented range;
- `2` as an example supported candidate;
- compact Darlok + Human/Klackon opponent cards;
- a 390-px mobile concept;
- a visible disabled `3 opponents` state with a short reason.

The prototype is **not** registered in the runtime asset manifest and does not imply Gate-2 support for 1-7 opponents.

## 12. Mobile disabled-state contract candidate

The shared `VisualSelector` already supports visible availability/status semantics and keyboard/arrow navigation. Slice 16.6 should extend the server contract so an opponent-count option can also carry a short **reason** when unavailable.

Candidate 390-px behavior:

- keep all original-range count tiles 1-7 browseable if Gate 2 finds that useful;
- unsupported count remains visibly disabled/planned rather than disappearing;
- title/count remains one line;
- previous/next targets remain at least 44 px;
- compact opponent portrait chips wrap below the number selector instead of shrinking portraits to unreadable size;
- one concise reason sits immediately below the disabled count/card;
- Create Game is disabled whenever the selected composition is not server-supported;
- API submission of the same unsupported tuple is rejected with the same semantic reason;
- the UI must never silently reduce opponent count, substitute a race, or change galaxy size.

The current generic selector has availability presentation but no server-owned per-option composition reason yet; that belongs in the Gate-2 API freeze.

## 13. Gate-2 decisions now ready to freeze

Gate 1 leaves the following explicit decisions for Gate 2:

1. **Supported V1 opponent counts:** whether to freeze the conservative `1-2` opponent matrix now, or perform additional runtime/race work before Gate 3.
2. **Galaxy-size matrix:** same accepted count range for all four sizes vs a deliberate server-owned per-size capacity table.
3. **Home-star allocator:** exact deterministic N-player MOOX normalization replacing `farthestNewGameSystemPair` while preserving well-separated-home intent.
4. **Opponent race policy:** automatic candidate A vs explicit per-slot candidate B.
5. **Duplicate policy:** recommended unique preset IDs.
6. **Random policy:** recommended none in V1; if added, server-owned seeded assignment only.
7. **Controller policy:** first production breadth should remain exactly one local human plus N built-in AI opponents; remote/hotseat seats are outside this slice unless explicitly frozen.
8. **Composition catalog/API:** server-owned count availability, race assignment/facts and disabled reasons; frontend must not derive hidden support rules.
9. **Final launch summary:** exact compact contents/order for galaxy, difficulty, local race, technology start, opponent count/races, and seed.
10. **Acceptance matrix:** deterministic fixtures and create/play/save/resume coverage per accepted count class.

## Gate-1 conclusion

Slice 16.6 should not leap directly from the current 1-opponent baseline to the original seven-opponent maximum.

The repository has unusually strong evidence for a safe intermediate class: **one local player plus up to two built-in AI opponents**. Production New Game still requires an N-player home-star allocator and server composition contract before even that class is enabled. Counts 3-7 remain an important original-facing target, but need broader race support and wider multi-AI lifecycle evidence first.

Gate 2 can now freeze that boundary without inventing frontend-only behavior or confusing original capability with current MOOX capability.