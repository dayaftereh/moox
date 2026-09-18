# Slice 16.7 Gate 1 - Advanced starting technology parity audit

Date: 2026-09-18

Status: **Gate 1 complete; Gate 2 next**.

## Scope and guardrail

Slice 16.7 is the final Slice-16 sub-slice. Its product objective is to turn the already visible but planned **advanced** starting-technology choice into a complete, deterministic, server-authoritative New Game mode.

Gate 1 is audit/evidence only. It does **not** enable Advanced, alter catalog availability, or freeze a Gate-2 implementation contract.

The audit compares Advanced against the frozen Slice-16.5 Pre-Warp/Average baseline and against the Slice-16.6 two-/three-player New Game composition matrix.

## Evidence basis

Gate 1 used only evidence already owned or normalized by the repository plus a fresh private re-read of the owned 1.31 executable:

- original HELP records 548-551 in reference/original/text/help/block_0000.ascii.txt;
- reference/original/support/files/Orion2.exe;
- the repository LE reader internal/moo2exe;
- the existing ignored build/research-disasm helper using x86asm;
- normalized technologies.json, buildings.json, race data and command-point rules;
- prior permanent technology, Advanced-start, New Game, difficulty, command-point, Slice-16.5 and Slice-16.6 research.

No private executable extract or disassembly output is committed by this Gate.

## Original public contract

Original HELP exposes exactly three technology-level choices:

1. Prewarp;
2. Average;
3. Advanced.

HELP describes Advanced as materially different in three ways:

- the player begins with a larger Empire filling the player's share of the galaxy according to galaxy size and player count;
- the player begins several technology levels ahead;
- the starting fleet uses all available Command Points.

The first two statements are strongly corroborated by executable behavior. The fleet statement conflicts with the verified 1.31 executable path and therefore remains an explicit Gate-2 parity decision rather than an assumption.

## Existing normalized Advanced technology layer

The technology portion of Advanced is already substantially implemented below the product New Game boundary.

### Stable identifiers and Average baseline

MOOX uses the stable IDs pre_warp, average and advanced.

The frozen Average start completes the seven fields:

0, 22, 23, 28, 29, 55, 57

and, in Tactical Combat mode, owns the already frozen set of 20 concrete starting technologies.

### Advanced grants

Prior executable research proves original technology-level global 0x21CB5 performs:

- Pre-Warp: 1 staged grant;
- Average: 6 staged grants after the always-known field;
- Advanced: 25 staged grants after the same base.

Therefore Advanced is **Average plus exactly 19 additional field grants**, ending with exactly **26 completed starting fields** per Empire.

The concrete applications are not one global fixed set. They depend on:

- shared New Game RNG state;
- the current open technology frontier;
- race research modifiers;
- personality/objective/theme preference weights;
- Creative / Uncreative semantics;
- government and game-mode eligibility;
- competition caused by technologies already owned by Empires initialized earlier.

The UI and Gate-2 contract must therefore not promise a fixed Advanced application list.

### Normalized source identity

data/rulesets/moo2-1.31/technologies.json already records:

- always-known field 0;
- staged fields 29, 55, 22, 57, 28, 23;
- source id moo2-1.31-orion2-new-game-techfields;
- source offset 2095024.

The same ruleset carries original-observed technology/technology-field tables and AI-research metadata used by the Advanced chooser.

### Current implementation state

internal/game/research_advanced.go already implements and tests:

- advancedNewGameExtraGrantCount = 19;
- validated Advanced preference profiles;
- stable Empire-ID ordering;
- one shared caller-owned New Game RNG;
- Average initialization followed by 19 Advanced chooser/grant cycles;
- race weighting;
- Creative and Uncreative behavior;
- government/application eligibility;
- Strategic/Tactical availability filtering;
- cross-Empire competition;
- original integer weighting and one-draw weighted-choice behavior.

The focused Advanced test suite passed during this Gate.

Production EconomyRules.NewGame, however, passes no AdvancedPreferences, and the authoritative settings validator still rejects advanced. No non-test production caller materializes those preference profiles.

## Original New Game call order

Fresh 1.31 disassembly reconfirmed the ordering that matters for deterministic integration.

Init_New_Game_ at VA 0x12479 performs:

1. ship/colony/global setup;
2. Init_Players_;
3. universe generation;
4. homeworld generation;
5. for Advanced, Advanced_Civilization_Colonies_;
6. for Advanced, Assign_Advanced_Civilization_Starting_Ships_;
7. later normal ownership/body/report initialization.

Init_Players_ at VA 0x12983 calls Init_NPC_Personalities_Objectives_Themes_ before per-player Init_Player_Tech_.

Consequences:

- Advanced preference materialization precedes Advanced technology selection;
- technology selection consumes the shared New Game RNG before galaxy generation;
- later Empire grants may depend on earlier Empire grants;
- extra Advanced colonies are selected only after galaxy/homeworld generation;
- a final fleet contract must run after inputs that determine Command-Point capacity.

## Dedicated original Advanced territory pipeline

Fresh debug-symbol resolution confirms the dedicated 1.31 Advanced territory path:

| Function | VA |
| --- | ---: |
| Num_Adv_Civ_Planets_ | 0x62BB7 |
| Advanced_Civilization_Colonies_ | 0x62C70 |
| Build_Adv_Civ_Star_List_ | 0x62E98 |
| Build_Adv_Civ_Planet_List_ | 0x63035 |
| Get_Worth_For_All_Planets_ | 0x63156 |
| Get_Adv_Civ_Planet_Worthiness_ | 0x63259 |
| Get_Next_Adv_Civ_Planet_ | 0x6341C |
| Choose_Adv_Civ_Planets_ | 0x63577 |
| Assign_Advanced_Civilization_Starting_Ships_ | 0x63848 |
| Twiddle_Selected_Adv_Civ_Planets_ | 0x638A9 |
| Init_Adv_Civ_Homeworlds_ | 0x63B5A |
| Twiddle_Planet_ | 0x63B8F |
| Write_Adv_Civ_Planet_Info_ | 0x63D0A |

Advanced_Civilization_Colonies_ builds parsec-distance state, chooses/twiddles Advanced planets, then initializes selected planets through the non-homeworld colony initializer. This is not equivalent to the current one-home-per-Empire MOOX bootstrap.

## Exact Advanced colony-count formula

Direct Num_Adv_Civ_Planets_ disassembly computes:

**additional colonies per Empire = ceil(floor(NUM_STARS / 2) / NUM_PLAYERS) - 1**

The existing homeworld supplies the +1, so total starting colonies per Empire are:

**ceil(floor(NUM_STARS / 2) / NUM_PLAYERS)**

For the exact Slice-16.6 supported galaxy/player matrix:

| Galaxy | Stars | 2 players: extra / total | 3 players: extra / total |
| --- | ---: | ---: | ---: |
| Small | 20 | 4 / 5 | 3 / 4 |
| Medium | 36 | 8 / 9 | 5 / 6 |
| Large | 54 | 13 / 14 | 8 / 9 |
| Huge | 71 | 17 / 18 | 11 / 12 |

The **count** contract is directly proven. The complete original planet-worth/selection/twiddle algorithm is not yet normalized into production MOOX and remains Gate-2 work.

## Advanced population is not the frozen Population-8 baseline

Fresh Init_Homeworld_Colony2_ disassembly proves non-Advanced and Advanced population paths diverge.

Let:

- L be the race-adjusted colony population limit returned by the original colony population-limit calculation;
- G be the signed race population-growth delta at original player+0x8A0.

For Advanced, the original integer arithmetic is equivalent to:

**P = trunc(L * (G + 300) / 500)**

and, when L <= 8, applies:

**P = trunc((P + L) / 2)**

It then clamps population to at least 1. For the actual homeworld only, it additionally clamps to at least 8.

For non-Advanced, the path assigns fixed Population 8.

Init_Non_Homeworld_Colony_ reaches the same common initializer without the homeworld-minimum flag, so extra Advanced colonies receive the variable Advanced population calculation without the minimum-8 clamp.

For the currently supported Slice-16.6 races Human, Klackon and Darlok, no normalized population-growth/aquatic/subterranean/tolerant trait changes these race-population inputs. Selected planet capacity still matters.

MOOX's current 4 Farmers / 2 Workers / 2 Scientists start split is a deliberate normalized baseline, not an original Advanced job-allocation proof. Gate 2 must explicitly freeze how variable Advanced population is divided into jobs.

## Advanced starting buildings

The original common colony initializer clears the starting-building array and scans a 42-product candidate list at LE object 2 offset 0x58AC.

Fresh disassembly proves the technology-level building cap table at object-1 offset 0x3A3A is:

- Pre-Warp: 3;
- Average: 5;
- Advanced: 9.

Candidates remain subject to original buildability, population and race/product conditions. Capitol handling is separate.

The exact original candidate production-ID order is:

41, 8, 40, 21, 22, 15, 7, 20, 37, 4, 31, 33, 34, 39, 2, 12, 13, 32, 35, 10, 43, 16, 28, 25, 18, 19, 24, 26, 47, 27, 6, 5, 23, 36, 29, 30, 46, 14, 45, 3, 42, 38

These IDs map cleanly to normalized original-observed buildings.json entries, including Star Fortress, Battlestation, Star Base, Hydroponic Farms, Marine Barracks, Biospheres, Automated Factories, research/production/economy facilities and late defenses.

Slice 09 / Slice 16.5 intentionally froze Pre-Warp/Average MOOX starts with no starting Buildings as a documented simplification. Full Slice-16.7 Advanced parity cannot accidentally inherit that simplification without an explicit Gate-2 decision.

## Treasury and Freighters

Existing direct Init_Players_ evidence writes the common player values before the Advanced territory stage:

- Treasury: **50 BC**;
- Freighters: **0**.

Gate 1 found no Advanced override. They remain the evidence-backed common baseline unless Gate 2 discovers contrary direct evidence.

## Advanced fleet: HELP intent versus verified 1.31 runtime

This Gate found a material evidence conflict that must remain visible.

Original HELP says the Advanced starting fleet uses all available Command Points.

The owned 1.31 executable performs:

Assign_Advanced_Civilization_Starting_Ships_ -> loop each player -> call Init_Advanced_Civilization_Fleet_.

But the exact Init_Advanced_Civilization_Fleet_ entry at VA 0x13FB7 is a single RET instruction.

The ordinary homeworld initializer independently proves the two Scouts plus Colony Ship creation block is gated specifically to technology level 1 (Average). Technology level 2 (Advanced) skips it.

Therefore, for the verified owned Orion2.exe 1.31 runtime path:

- the documented “fill Command Points” fleet is **not executed**;
- Advanced does **not** inherit the Average two-Scout + Colony-Ship block through this path.

Gate 2 must explicitly choose the parity target:

1. **Executable parity:** preserve the verified 1.31 no-op Advanced fleet path; or
2. **Documented product-intent parity:** implement a deterministic Advanced fleet filling final Command-Point capacity.

If option 2 is selected, exact ship composition/hulls/equipment is not yet normalized and must be frozen deliberately rather than invented.

## Command-Point ordering dependency

MOOX already derives Command-Point capacity from normalized rules:

- base capacity 5;
- command-station buildings;
- known Communications technology;
- Warlord +2 per owned Colony;
- Dictatorship plus known Imperium: +50% with integer truncation.

Command-Point usage is derived from actual ships/fleets.

Therefore any HELP-intent “fill all Command Points” implementation must run only **after**:

- Advanced technology ownership is final;
- all Advanced colonies are final;
- starting buildings/stations are final.

Otherwise the fleet target capacity can be wrong.

## Race-specific interactions in the supported composition matrix

Slice 16.6 currently supports local Human or Klackon with Darlok and, for three players, the alternate supported race.

Relevant normalized traits:

- Human: Charismatic + Democracy;
- Klackon: Unification + farming/industry bonuses + **Uncreative**;
- Darlok: spying bonus + Stealthy Ships + Dictatorship.

Advanced implications:

- Klackon exercises the deterministic Uncreative application-plan path;
- Human government/race profile affects Advanced weighting/eligibility;
- Darlok race/government profile affects the chooser;
- cross-Empire competition means initialization order changes later Empire candidate weights.

Gate-2 fixtures must cover both frozen Slice-16.6 composition orders, not only one Human/Darlok game.

## Difficulty interaction remains unresolved at the original secondary-filter layer

MOOX has five stable difficulty IDs:

easy, normal, hard, very_hard, impossible.

Prior original research identified global byte 0x21CB0:

- default game settings set it to 0;
- non-default values activate an additional Advanced chooser filter;
- the same original byte participates in NPC/personality/maintenance behavior.

The repository has deliberately **not** claimed a public mapping between 0x21CB0 and the modern five MOOX difficulty IDs.

Current InitializeNewGameTechnologies does not consume DifficultyID.

Gate 2 must explicitly choose one evidence-safe direction:

- perform enough additional reverse engineering to freeze the original setting mapping; or
- define the first supported MOOX Advanced chooser as the directly proven default 0x21CB0 == 0 path across all five current MOOX difficulties and document the unresolved original secondary layer.

Gate 1 does not guess the mapping.

## Galaxy and player-composition interaction

Advanced colony count directly depends on both galaxy star count and total player count.

Therefore:

- 2-player and 3-player games have different colony targets;
- Small/Medium/Large/Huge all have different targets;
- current home-placement alone is insufficient to create the Advanced Empire footprint.

The current 2-/3-player controller model remains compatible with Advanced in principle, but the Advanced territory algorithm must be integrated before the mode can be enabled.

## Persistence and Save/Resume

Core serialization already persists materialized authoritative state, including:

- Seed and RNG state;
- Difficulty;
- Galaxy and planets;
- Empires and known technology;
- Colonies;
- ship designs, ships and strategic fleets;
- Command Points and Treasury;
- Uncreative/hyper-advanced research state where applicable.

Existing Slice-16.6 three-player persistence tests already prove byte-identical export/import/re-export for current modes.

Advanced preference profiles are **not** currently Core/App/Server state. They are initialization inputs only.

This creates a Gate-2 design point:

- if personality/objective/theme is bootstrap-only for first MOOX Advanced support, deterministic generation must reproduce those profiles from seed/settings/order before technology initialization;
- if those profiles are intended to affect later AI behavior, they must become explicit authoritative persisted state rather than ephemeral bootstrap data.

Save/Resume after bootstrap can use the existing persistence model if all resulting Advanced state is materialized in existing Core fields.

## Current production gap table

| Area | Current MOOX state | Advanced Gate-1 finding |
| --- | --- | --- |
| Catalog/UI | Advanced visible, planned/locked | Correct until backend parity exists |
| Settings validation | Advanced rejected | Must remain rejected through Gate 1/2 |
| Tech chooser | Advanced algorithm implemented/tested | Needs production preference materialization |
| Empire order / competition | Stable algorithm exists | Preserve Slice-16.6 seat/order semantics |
| Territory count | One home colony currently | Exact Advanced colony-count formula proven |
| Territory selection | Current home selection only | Original Advanced planet selection/twiddle not normalized |
| Population | Frozen Population 8 | Advanced variable population formula proven |
| Jobs | Fixed 4/2/2 MOOX baseline | Variable-population split not frozen |
| Buildings | Empty start baseline | Original Advanced cap/order proven; exact resulting fixture not frozen |
| Treasury/Freighters | 50 BC / 0 | Evidence supports retaining both |
| Fleet | Pre-Warp none; Average 2 Scouts + Colony Ship | HELP/runtime conflict requires Gate-2 decision |
| Command Points | Fully normalized derivation | Fleet, if used, must follow colonies/buildings/tech |
| Difficulty | Five MOOX IDs | Original 0x21CB0 mapping unresolved |
| Persistence | Materialized state round-trips | Preference profile lifecycle needs explicit decision |

## Deterministic fixture evidence required for Gate 2/3

The later implementation contract should freeze fixed-seed exact fixtures for all supported Advanced composition classes, at minimum:

1. Human + Darlok, 2 players;
2. Klackon + Darlok, 2 players;
3. Human + Darlok + Klackon, 3 players;
4. Klackon + Darlok + Human, 3 players.

Across acceptance, cover all four galaxy sizes because colony counts change with star count, and all five MOOX difficulty IDs because the public product supports them even if Gate 2 freezes identical default-path chooser semantics.

Each exact fixture should be capable of asserting:

- exactly 26 completed starting fields;
- exact fixed-seed selected field IDs;
- exact known concrete technology IDs for that seed/race/profile/order;
- exact Uncreative choice state where applicable;
- deterministic Empire initialization order;
- exact post-bootstrap RNG state;
- exact Advanced colony count and selected system/planet IDs;
- capital/homeworld identity;
- exact population totals per starting Colony;
- frozen Advanced job allocation;
- exact starting Buildings if Gate 2 chooses original building parity;
- Treasury 50 BC and 0 Freighters;
- exact ship/design/fleet state under the explicit Gate-2 fleet decision;
- exact Command-Point Capacity/Used;
- byte-equal authoritative state for repeated equal seed/settings.

Product acceptance must later add:

- public HTTP create/play for 2- and 3-player Advanced games;
- built-in-AI turn advancement;
- catalog/UI supported state only after backend parity exists;
- launch briefing accuracy;
- byte-identical export/import/re-export;
- restore plus continued turn resolution.

## Gate-2 decisions/blockers

Gate 1 deliberately leaves these questions open:

1. **Advanced fleet target:** verified executable no-op versus HELP “fill Command Points” product intent.
2. **Advanced territory selection:** freeze/implement enough original planet-worth/selection/twiddle behavior to produce deterministic exact planets.
3. **Variable population jobs:** define deterministic job split for non-8 Advanced populations.
4. **Starting Buildings:** decide whether Slice 16.7 requires the proven original Advanced building bootstrap rather than the earlier empty-building simplification; if yes, freeze exact candidate/buildability behavior.
5. **Preference profiles:** define deterministic profile generation and whether profiles are bootstrap-only or persistent AI state.
6. **Difficulty secondary layer:** either normalize 0x21CB0 or explicitly freeze default-path semantics across the five MOOX difficulties.

These decisions are substantial enough that Gate 2 must be explicitly released before any semantic freeze or runtime implementation.

## Gate-1 conclusion

Gate 1 confirms that Advanced is not “Average plus 19 tech rolls.” Its complete original bootstrap also changes territorial breadth, population and starting-building behavior, and it contains a documented-versus-executable fleet discrepancy.

The existing MOOX Advanced technology chooser is a strong foundation and remains green, but product enablement is correctly still blocked.

**Next step after explicit release: Slice 16.7 Gate 2 - freeze the complete Advanced bootstrap contract, resolving the six decisions above.**
