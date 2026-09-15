# Planned slice 16.3 - Galaxy visual selector

Status: **open; Gate 1 complete, Gate 2 next (2026-09-15)**.

Parent: `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`.
Depends on: Slice 16.1 shared visual selection grammar.

## Objective

Make galaxy setup visual and spatial. The player should see and feel the difference between galaxy sizes/ages rather than selecting abstract values from a dropdown.

## Binding visual direction

Use the shared previous/next carousel for galaxy configuration. The central image should be an original MOOX galaxy diagram/illustration whose star density/extent visibly changes with the setting.

### Galaxy size

The supported size set is audited by Gate 1 and frozen only in Gate 2. Gate-1 evidence identifies the fidelity-oriented four-row candidate `small / medium / large / huge`; the older Slice-16.1 `tiny` artwork is not an authoritative fifth generator row. The presentation should communicate an obvious progression such as:

- small / compact;
- medium / broader;
- large;
- very large / huge where supported.

Do not bind unsupported options just because artwork exists.

Suggested art behavior:

- same framing language for every size;
- increasingly large/dense star distribution;
- optional subtle scale grid/range ring;
- no misleading exact star positions: this is an option illustration, not the generated map preview.

### Galaxy age

Gate-1 audit corrected the original semantics: Galaxy Age cycles **Mineral Rich / Normal / Organic Rich**. It is not a young/normal/old astronomical-age selector.

Candidate visual cues therefore communicate composition bias rather than temporal age:

- Mineral Rich: stronger faceted/mineral-resource cue;
- Normal: balanced mineral/biosphere cue;
- Organic Rich: stronger biosphere/food-world cue.

Keep the same galaxy silhouette, decorative star count and overall map footprint across the three Age illustrations so the artwork does not imply a size change or generated-map preview. Do not invent exact planet ratios or frontend-owned spectral-colour semantics. Gameplay effects and facts must come from the authoritative generator/server contract.

## Information strip

Show short facts derived from server/normalized settings data, such as:

- expected/system count when authoritative;
- generator size ID;
- age effect summary if evidence-backed;
- player-capacity constraints if the selected size limits opponent count.

## Gate 1 - audit

- [x] Re-check original galaxy size and age tables.
- [x] Map current generator support and star/system counts.
- [x] Identify settings that are runtime-ready without new generator mechanics.
- [x] Prototype one coherent image series for size and one for age.
- [x] Verify no illustration implies data the generator does not guarantee.

Permanent Gate-1 evidence: `docs/research/SLICE_16_3_GATE1_GALAXY_VISUAL_SELECTOR_AUDIT_2026-09-15.md`.

## Gate 2 - freeze

- [ ] Freeze supported galaxy sizes and ages.
- [ ] Freeze exact runtime IDs and displayed facts.
- [ ] Freeze asset variants and responsive presentation.

## Gate 3 - implementation

- [ ] Extend authoritative New Game validation/generation for the accepted settings.
- [ ] Add visual carousel selectors for size and age.
- [ ] Add deterministic seed fixtures per supported tuple.
- [ ] Keep invalid size/player-count combinations server-rejected and visibly disabled.

## Gate 4 - close

- [ ] Repeat equal-seed/settings generation and hash equality.
- [ ] Verify visual selection maps exactly to authoritative generator settings.
- [ ] Desktop/mobile browser QA.
- [ ] Full tests/build/diff checks.

## Exit criterion

Galaxy size and age can be chosen visually with left/right navigation, are understandable from the artwork plus compact facts, and generate the exact deterministic authoritative galaxy contract selected by the player.