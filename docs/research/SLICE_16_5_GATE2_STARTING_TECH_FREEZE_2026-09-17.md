# Slice 16.5 Gate 2 - Starting technology selector / full-start contract freeze

Date: 2026-09-17
Status: **Gate 2 complete / Gate 3 next**

## Purpose

Freeze the smallest honest Slice-16.5 contract that can add real starting-technology breadth while preserving original-start semantics. This freeze separates three concerns that must not be conflated:

1. the canonical three-level technology-start catalog;
2. the full New Game start state that may actually be submitted today;
3. the visual selector/catalog experience, which may show planned choices without pretending they are server-enabled.

Gate 2 does not widen runtime behavior. Gate 3 may implement only the contract frozen here.

## Evidence basis

- `docs/research/SLICE_16_5_GATE1_STARTING_TECH_AUDIT_2026-09-16.md`;
- `docs/research/prototypes/SLICE_16_5_STARTING_TECH_2026-09-16/`;
- `data/rulesets/moo2-1.31/technologies.json`;
- `internal/game/research.go` and `internal/game/research_advanced.go`;
- `internal/game/new_game.go` and current New Game regression tests;
- `docs/research/SLICE_16_1_GATE2_FREEZE_2026-09-15.md` for the shared visual-selector grammar.

## Decision summary

Slice 16.5 freezes **three canonical catalog choices** but only **two Gate-3 runtime-selectable starts**.

| Player label | Canonical ID | Catalog-visible | Gate-3 selectable | Status |
| --- | --- | --- | --- | --- |
| Pre-Warp | `pre_warp` | yes | yes | supported in Gate 3 |
| Average | `average` | yes | yes | supported baseline |
| Advanced | `advanced` | yes | no | planned / visible only |

This is the same supported/planned distinction already frozen by Slice 16.1: a canonical option may be browsed and explained without being accepted by Create Game.

### Why Advanced is not enabled in Gate 3

Advanced technology ownership itself is normalized, but a truthful **full Advanced New Game start** is not yet normalized in production. Enabling the enum alone would be false parity because:

- production New Game does not yet materialize the required per-Empire `AdvancedPreferences`;
- Advanced begins with a larger empire rather than the current single-homeworld bootstrap;
- the non-Advanced Population-8 branch must not be reused blindly;
- the original Advanced starting fleet fills available Command Points rather than reusing the Average two-Scout + Colony Ship bootstrap;
- the extra 19 research-field grants are race-, RNG- and cross-Empire-order dependent.

Gate 3 therefore keeps `advanced` rejected by the authoritative New Game validator while exposing it as a clearly planned catalog card. A later bounded slice may enable it only after the complete empire/population/fleet/preference bootstrap is frozen and implemented.

## Frozen full-start contract

Technology level changes the complete start bootstrap, not only technology ownership.

### Common non-Advanced baseline

For `pre_warp` and `average`, preserve the current normalized non-Advanced base where technology level does not explicitly override it:

- one starting star / one home colony per Empire;
- Population 8, using the existing normalized 4 Farmers / 2 Workers / 2 Scientists start split;
- 50 BC starting treasury;
- 0 freighters;
- all existing deterministic galaxy/home-system/race/difficulty behavior remains unchanged.

### Pre-Warp - `pre_warp`

Freeze these technology-level effects:

- completed starting fields: `0, 29`;
- known normalized Technology IDs: `32, 40, 103, 145, 166, 168`;
- **no starting ships**;
- no Average Scout pair and no Colony Ship are created;
- no starting-fleet/design bootstrap may require Average-only interstellar components;
- the existing common non-Advanced home-colony/economy baseline remains otherwise unchanged.

The runtime must derive this state from the authoritative server contract; the frontend does not construct grants or ships.

### Average - `average`

Average remains the non-regression baseline:

- completed starting fields: `0, 22, 23, 28, 29, 55, 57`;
- 20 known Tactical-combat applications under the Slice-16 fixed Tactical setting;
- one starting star / current Population-8 home-colony baseline;
- 2 Scouts;
- 1 Colony Ship;
- 50 BC;
- 0 freighters.

Gate 3 must prove that selecting Average produces the same deterministic start as before Slice 16.5.

### Advanced - `advanced` (catalog/planned only)

Freeze the explanatory contract but **not** runtime acceptance:

- Average technology baseline first;
- exactly 19 additional research-field grants per Empire;
- larger starting empire;
- starting fleet fills available Command Points;
- exact technology applications vary with seed, race, preference profile and cross-Empire initialization context.

No fixed Advanced technology list or fixed final application count may appear in frontend copy.

## Frozen Advanced RNG / race semantics

When Advanced is implemented later, these semantics are binding:

- one caller-owned New Game RNG drives initialization;
- Empire initialization order is part of deterministic behavior;
- race research modifiers participate in chooser weights;
- personality/objective/theme preference profiles are server-owned inputs;
- earlier Empires' ownership may alter competition weights for later Empires;
- Creative gains every legal application in a chosen field;
- Uncreative follows its fixed New Game application plan;
- game/race eligibility filters remain authoritative.

Deterministic Advanced fixtures must therefore lock at least seed, Empire/race order, preference profiles and combat mode. The UI must never try to reproduce the chooser.

## Frozen server-owned catalog / fact contract

The server is the authority for option ID, label, support state and compact facts. Gate 3 should expose all three catalog entries and mark `advanced` unsupported/planned for submission.

Stable compact facts:

### Pre-Warp

- `Single star`
- `No starting ships`
- `Interstellar capability must be researched`

Technical detail may additionally state `2 completed starting fields / 6 known applications` when useful in the information surface.

### Average

- `Single star`
- `2 Scouts + 1 Colony Ship`
- `7 completed starting fields / 20 Tactical applications`

### Advanced

- `Larger starting empire`
- `Average baseline + 19 extra research fields`
- `Starting fleet fills Command Points`
- `Exact grants vary by seed, race and initialization context`

The compact card must not hard-code hidden grants. Detailed fact copy is rendered from the server-owned normalized catalog/contract.

## Frozen visual / asset contract

The accepted Gate-1 progression becomes the production direction:

1. **Pre-Warp / workshop-orbit dawn** - sparse instrumentation, one planetary industrial/workshop focal point, incomplete orbital structure, no outbound fleet silhouette.
2. **Average / interstellar launch** - mature command/research facility, readable two-scout + colony-vessel motif, established orbital infrastructure.
3. **Advanced / networked stellar command** - broader multi-node footprint, denser command/fleet presence and refined luminous research core.

Production convention:

- use original deterministic SVG artwork derived from the accepted Gate-1 prototypes;
- semantic asset IDs: `new-game:technology-level:<option-id>`;
- runtime paths: `/assets/new-game/technology-level/<option-id>.svg`;
- no localized/accessibility copy embedded in artwork;
- all three share one composition grammar and center-safe focal region;
- the distinct progression must remain legible in the validated 390 px single-card presentation;
- no copied MOO2 UI panels or artwork.

## Frozen selector / accessibility behavior

Inherit Slice 16.1 without a technology-specific fork:

- one compact setting card using the shared `VisualSelector` grammar;
- previous/next mouse and touch controls;
- ArrowLeft, ArrowRight, Home and End keyboard navigation;
- minimum 44 px mobile controls;
- artwork `?` opens detailed information;
- mobile bottom sheet / larger-screen centered modal;
- focus remains contained while the dialog is open and returns to the info control on close;
- unsupported/planned state is explicit in the information surface;
- Create Game remains disabled/rejected when the selected option is not accepted by the authoritative server contract.

## Gate-3 implementation boundary

Gate 3 may:

1. add a server-owned three-entry starting-technology catalog with supported/planned status and facts;
2. widen authoritative New Game validation from Average-only to `pre_warp | average`;
3. branch the full non-Advanced start bootstrap so Pre-Warp creates no starting ships while Average remains unchanged;
4. bind technology initialization to the selected supported level;
5. add deterministic Pre-Warp/Average fixtures across the accepted race matrix;
6. add the three production SVGs and the shared visual selector;
7. show Advanced as planned while preventing its submission.

Gate 3 must **not**:

- accept `advanced` in Create Game;
- fabricate `AdvancedPreferences`;
- reuse the Average homeworld/fleet bootstrap for Advanced;
- hard-code Advanced grant lists in the frontend;
- widen unrelated galaxy/opponent/race/game-mode contracts.

## Gate-2 exit criteria

Gate 2 is complete when:

- the three canonical IDs/labels are frozen;
- runtime support is explicitly `pre_warp | average`, with `advanced` planned;
- the complete Pre-Warp/Average technology-level start deltas are frozen against the existing non-Advanced baseline;
- Advanced RNG/race/order semantics and deferral boundary are explicit;
- server-owned facts are frozen;
- the original SVG progression, runtime asset convention and 390 px behavior are frozen;
- the inherited selector/accessibility behavior is explicit;
- Gate 3 has a bounded implementation/test sequence with no hidden Advanced parity claim.

All criteria above are satisfied by this document. Proceed to Gate 3 implementation.
