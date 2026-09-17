# Planned slice 16.6 - Player composition and New Game integration

Status: **open; Gate 1 complete, Gate 2 next (2026-09-17)**.

Parent: `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`.
Depends on: Slices 16.1-16.5.

## Objective

Complete the player-composition and cross-setting integration stage of Slice 16 by turning opponent/player composition into the same image-led selection experience and integrating the currently accepted New Game settings into one coherent, deterministic creation flow. Full Advanced starting-technology parity and final Slice-16 closure are explicitly deferred to Slice 16.7.

## Binding player-count direction

The opponent-count selector should be visually immediate rather than a numeric input box.

Preferred MOOX interaction:

- large central illustrated number tile;
- previous/next arrows from the shared 16.1 selector;
- visually distinct number art for **1 through 7 opponents** as the original-parity presentation range; any future 8-9 opponent mode is a deliberate MOOX extension, not original parity;
- exact enabled maximum is server-authoritative and depends on the supported player-count/galaxy-size contract frozen in Gate 1/2;
- unavailable counts remain visible only if useful, but are clearly disabled with a short reason.

The number graphics should look intentionally designed, not like plain HTML digits: original MOOX-styled numerals with subtle starfield/empire motifs are preferred. These are good SVG candidates because they are symbolic, scalable and lightweight.

## Opponent composition

The composition UI should distinguish:

- Human player's selected preset race;
- opponent count;
- opponent race assignment where supported;
- built-in AI controller assignment;
- duplicate-race policy if allowed;
- any random-race option only if deterministic under the New Game seed/settings contract.

Reuse Slice-16.4 race portraits for opponent slots/cards rather than introducing separate opponent art.

Possible presentation once the contract is frozen:

```text
OPPONENTS
<        4        >

[ Race portrait ] Darlok
[ Race portrait ] ...
[ Race portrait ] ...
[ Race portrait ] ...
```

For larger counts, use compact portrait chips/cards under the main number selector rather than shrinking the race artwork into unreadable thumbnails.

## Cross-setting validation

This slice owns the cross-setting interaction between all currently accepted settings:

- **Combat mode is not selectable:** Tactical Combat remains fixed (`strategic_combat=false`) by product decision.
- **Random Events and Antaran Attacks remain outside Slice 16:** their future toggles belong to dedicated later gameplay slices, not this integration.

- galaxy size / age;
- difficulty;
- Human preset race;
- starting technology level;
- opponent count and race/controller composition;
- deterministic seed.

The UI must surface server validation rather than silently repairing unsupported combinations.

Examples:

- selected galaxy size cannot host selected player count;
- a preset race is blocked because a required mechanic is not supported yet;
- an opponent count is outside the accepted runtime matrix;
- a selected difficulty is not supported for the chosen controller contract.

## Final New Game summary

Before `Spiel erstellen`, present one compact visual summary using the selected artwork:

- galaxy illustration;
- difficulty emblem;
- Human race portrait;
- starting-tech artwork;
- opponent count and compact opponent portraits;
- seed.

This should feel like a launch briefing, not a debug settings dump.

## Gate 1 - composition audit

- [x] Audit original/runtime player-count limits and galaxy-size dependencies.
- [x] Audit built-in AI support for more than one opponent.
- [x] Freeze deterministic race-assignment/randomization behavior candidates.
- [x] Prototype number tiles 1-7 and compact opponent portrait layout.
- [x] Verify how disabled counts/combinations are explained on mobile.

Permanent Gate-1 evidence: `docs/research/SLICE_16_6_GATE1_PLAYER_COMPOSITION_AUDIT_2026-09-17.md`.
Research prototype: `docs/research/prototypes/SLICE_16_6_OPPONENT_COMPOSITION_2026-09-17.svg`.

Gate-1 finding: the original documented range is 2-8 total players (1-7 opponents), while current production MOOX is exactly two players and its generic home-star allocator is pair-only. The strongest evidence-backed Gate-2 V1 candidate is one local Human/Klackon plus one or two Builtin-AI opponents using unique races from Darlok and the other supported Human/Klackon race; 3-7 opponents remain visible original-range targets but require an explicit wider race/runtime policy before support.

## Gate 2 - freeze

- [ ] Freeze supported opponent/player-count matrix.
- [ ] Freeze opponent race/controller assignment rules.
- [ ] Freeze duplicate/random policy.
- [ ] Freeze final cross-setting validation and New Game summary layout.

## Gate 3 - implementation

- [ ] Extend server New Game request/validation for accepted multi-seat settings.
- [ ] Extend built-in AI seat bootstrap as required by the frozen matrix.
- [ ] Add opponent-count visual selector and opponent race composition UI.
- [ ] Integrate all Slice-16 selectors into one coherent New Game flow.
- [ ] Add deterministic seed/settings/race/player-count fixture matrix.
- [ ] Ensure Save/Resume and subsequent browser play work for every accepted setup class.

## Gate 4 - player-composition integration acceptance

- [ ] Desktop and real 390 px mobile New Game walkthrough.
- [ ] Create/play smoke for every supported difficulty/galaxy/tech/player-count class.
- [ ] Validate representative preset races and multi-AI sessions.
- [ ] Equal seed + complete settings tuple remains deterministic.
- [ ] No unsupported combination can be created through UI or API.
- [ ] Full Go tests/vet, Web production build and `git diff --check`.
- [ ] Update parent Slice-16 plan, ACTIVE_RESEARCH, PROJECT_STATUS and HISTORY; hand off final Advanced parity and Slice-16 closure to Slice 16.7.

## Exit criterion

The player can configure a substantially broader game from one polished image-led New Game flow, including visual race choice and opponent composition, and every accepted option maps exactly to a deterministic server-authoritative game setup. Tactical Combat remains the fixed MOOX combat mode; Random Events and Antaran Attacks are intentionally absent until dedicated later slices make those systems real.
