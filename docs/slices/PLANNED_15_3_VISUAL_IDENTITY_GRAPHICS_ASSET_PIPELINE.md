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

- MOOX corporate/brand direction, retaining/exploring the accepted **OX** lettermark plus a compact original spacecraft/star/orbit app icon, logo treatment and visual identity rules;
- final color palette roles, typography hierarchy and icon language;
- starfield/background and **2D Galaxy-map** treatment;
- star/system imagery plus the accepted orbital **star-system dialog** visual language (central sun, rocky/gas planets, asteroid belts, colony/fleet presence);
- planet/Colony Detail imagery, buildings/infrastructure presentation rules and construction-queue visual language;
- race/faction portrait direction for currently supported content;
- ship/fleet imagery for location-grouped Fleet tiles and star-system presence;
- building/infrastructure imagery strategy;
- visual identity for the eight Research categories (Construction, Power, Chemistry, Sociology, Computers, Biology, Physics, Force Fields);
- Diplomacy and Espionage icon/presentation language that can scale with future mechanics;
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
- [ ] Produce multiple styleboard/art-direction candidates for the OX+icon brand, shell, 2D Galaxy, star-system dialog, Colony Detail, planet, portrait, Fleet/ship and eight-category Research imagery.
- [ ] Define target resolutions/aspect ratios and responsive/cropping requirements.
- [ ] Define provenance/license rules for generated, commissioned, extracted/reference and third-party assets.
- [ ] Present Gate-2 visual-direction and asset-pipeline contract with representative mockups.

## Gate 2 - Visual identity / pipeline freeze

- [ ] Freeze MOOX visual direction, logo usage, palette, typography and icon grammar.
- [ ] Freeze core asset categories and minimum asset set for the canonical match.
- [ ] Freeze image/model/render pipeline, naming/versioning and fallback policy.
- [ ] Freeze provenance/license metadata requirements.
- [ ] Freeze representative acceptance mockups for strategic shell, 2D Galaxy + star-system dialog, Colony table + Colony Detail/build queue, location-grouped Fleets, eight-category Research and battle presentation.

## Gate 3 - Implementation

- [ ] Implement design tokens/theme assets in the browser shell.
- [ ] Build/import the accepted minimum asset set and runtime catalog mapping.
- [ ] Integrate the accepted OX/icon treatment, Galaxy/star-system visuals, planets, Colony/building/queue visuals, race/ship/Fleet imagery, Research-category visuals and icons into functional screens.
- [ ] Add fallback/loading behavior for missing or delayed assets.
- [ ] Document repeatable generation/render/import workflow for future content expansion.

## Gate 4 - QA + close

- [ ] Verify representative screens at accepted resolutions and responsive viewports.
- [ ] Verify asset provenance metadata and no accidental unapproved original-game asset shipping.
- [ ] Verify runtime asset load/fallback paths and production web build.
- [ ] Run relevant tests plus `git diff --check`, update evidence/status/HISTORY and close marker.
