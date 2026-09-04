# Planned slice 15.3 - MOOX visual identity / graphics / asset pipeline

Status: **planned / queued; not open**.

Queue position: **15.3 of Slice-15 family**.

## Objective

Create a distinctive, coherent Master of Orion X visual language and a reproducible asset pipeline that can feed the browser HMI without coupling art iteration to gameplay authority.

## Dependencies

- Slice 15.1 interaction/design-system contract.
- Slice 15.2 functional screens provide real integration targets.
- Existing graphic/palette/catalog extraction infrastructure may be reused where provenance permits.

## Scope guard

In scope:

- MOOX corporate/brand direction, logo treatment and visual identity rules;
- final color palette roles, typography hierarchy and icon language;
- starfield/background/world-map treatment;
- planet/system imagery and presentation rules;
- race/faction portrait direction for currently supported content;
- ship/fleet/building/technology imagery strategy;
- UI ornamentation, frames, hover/selection/disabled states and motion language;
- 2D illustration, generated imagery and optional 3D-model -> render/sprite workflows;
- asset naming, resolution/DPI/aspect-ratio variants, compression and runtime loading/fallback policy;
- source/provenance/license metadata and clear separation of original MOO2 reference evidence from shippable MOOX assets;
- integration contract between asset catalog and React components.

Defer:

- new gameplay rules or content breadth;
- broad Race Designer / Ship Designer mechanics;
- expensive cinematic/animated sequences unless explicitly accepted in Gate 2;
- reuse of copyrighted original-game assets without explicit provenance/license permission.

## Gate 1 - Art/asset audit and direction exploration

- [ ] Inventory current web visuals plus graphic/palette/catalog capabilities already present in the repo.
- [ ] Define visual goals: readable strategy UI, recognizable MOOX identity, sci-fi tone and acceptable nostalgia/reference distance.
- [ ] Produce multiple styleboard/art-direction candidates for shell, galaxy, planet, portrait and ship imagery.
- [ ] Define target resolutions/aspect ratios and responsive/cropping requirements.
- [ ] Define provenance/license rules for generated, commissioned, extracted/reference and third-party assets.
- [ ] Present Gate-2 visual-direction and asset-pipeline contract with representative mockups.

## Gate 2 - Visual identity / pipeline freeze

- [ ] Freeze MOOX visual direction, logo usage, palette, typography and icon grammar.
- [ ] Freeze core asset categories and minimum asset set for the canonical match.
- [ ] Freeze image/model/render pipeline, naming/versioning and fallback policy.
- [ ] Freeze provenance/license metadata requirements.
- [ ] Freeze representative acceptance mockups for strategic shell, galaxy/system, colony and battle presentation.

## Gate 3 - Implementation

- [ ] Implement design tokens/theme assets in the browser shell.
- [ ] Build/import the accepted minimum asset set and runtime catalog mapping.
- [ ] Integrate backgrounds, planets, race/ship/building/technology visuals and icons into functional screens.
- [ ] Add fallback/loading behavior for missing or delayed assets.
- [ ] Document repeatable generation/render/import workflow for future content expansion.

## Gate 4 - QA + close

- [ ] Verify representative screens at accepted resolutions and responsive viewports.
- [ ] Verify asset provenance metadata and no accidental unapproved original-game asset shipping.
- [ ] Verify runtime asset load/fallback paths and production web build.
- [ ] Run relevant tests plus `git diff --check`, update evidence/status/HISTORY and close marker.
