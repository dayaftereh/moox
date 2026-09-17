# Planned slice 16.5 - Starting technology visual selector

Status: **open; Gate 3 complete, Gate 4 next (2026-09-17)**.

Parent: `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`.
Depends on: Slice 16.1 shared visual selection grammar.

## Objective

Present the New Game starting-technology / technology-level choice as three strong visual cards instead of a text-only setting. The player should understand the progression from a more primitive start to a more developed start before reading the exact rule text.

## Binding product direction

- The Slice-16 HMI targets **three visual technology-start choices** if Gate 1 confirms the expected three-level authoritative contract.
- Exact canonical/internal labels and mechanics are audited before implementation; presentation copy must not invent technology grants.
- Use the common previous/next carousel or a three-card variant built from the same 16.1 selection grammar.
- Each option has one primary original image plus a compact server-derived summary of what changes at game start.

## Artwork concept

The three images should clearly communicate technological progression while remaining in one coherent MOOX art style.

Suggested visual arc:

### Lower / early technology start

- practical workshop / early orbital industry;
- simpler hulls, exposed machinery, fewer luminous systems;
- visual feeling: resourceful civilization at the beginning of interstellar development.

### Middle / standard technology start

- mature starship command / research facility;
- recognizable interstellar infrastructure;
- visual feeling: balanced standard MOOX start.

### Higher / advanced technology start

- sophisticated command chamber / advanced starship core / luminous research complex;
- denser instrumentation and more refined materials;
- visual feeling: civilization beginning with a broader technological foundation.

The images are illustrative setting art, not screenshots of exact technology trees and not promises of specific granted technologies unless the server catalog explicitly lists them.

## Asset strategy

- Simple symbolic technology-level illustrations may be SVG if they remain visually rich enough.
- Painterly setting illustrations may use PNG/WebP.
- Reuse the same framing/safe-area rules as the other New Game selectors.
- No copied original-game UI panels or artwork.

## Information strip

Once Gate 1 freezes the exact authoritative rules, show concise facts such as:

- starting technology level name;
- number/category of granted technologies if meaningful and stable;
- starting research position or cost modifiers if part of the contract;
- compatibility notes with selected preset race only when server-authoritative.

The frontend must not hard-code hidden technology grants.

## Gate 1 - audit

- [x] Re-check original and normalized New Game technology-level options.
- [x] Confirm whether the supported target is exactly three levels and their authoritative IDs.
- [x] Inventory exact start-state effects/granted technologies for each level.
- [x] Identify race-specific interactions.
- [x] Prototype the three-image visual progression.

Permanent Gate-1 evidence: `docs/research/SLICE_16_5_GATE1_STARTING_TECH_AUDIT_2026-09-16.md`.
Prototype evidence: `docs/research/prototypes/SLICE_16_5_STARTING_TECH_2026-09-16/`.

## Gate 2 - freeze

- [x] Freeze supported technology-start set and labels.
- [x] Freeze exact server contract/effects.
- [x] Freeze art direction and asset format.
- [x] Freeze compact explanatory facts.

Permanent Gate-2 evidence: `docs/research/SLICE_16_5_GATE2_STARTING_TECH_FREEZE_2026-09-17.md`.

Advanced parity handoff: dvanced remains catalog-visible/planned during Slice 16.5 runtime implementation. Full authoritative Advanced empire/population/fleet/research bootstrap and enablement are explicitly owned by PLANNED_16_7_ADVANCED_STARTING_TECH_PARITY.md.

## Gate 3 - implementation

- [x] Extend New Game validation/state generation for accepted technology levels.
- [x] Add visual selector/cards.
- [x] Bind displayed facts to authoritative normalized/server data.
- [x] Add deterministic fixtures for race + technology-level combinations in the accepted matrix.

## Gate 4 - close

- [ ] Each accepted technology start creates the exact frozen initial state.
- [ ] Equal settings remain deterministic.
- [ ] Three images clearly differentiate progression on desktop and 390 px mobile.
- [ ] Full tests/build/diff checks.

## Exit criterion

Starting technology is a visually meaningful New Game choice with three coherent MOOX illustrations and authoritative, evidence-backed start-state semantics.
