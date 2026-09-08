# Open slice 15.3 - MOOX visual identity / graphics / asset pipeline

Status: **open; Gate 1 art/asset audit and direction package complete; awaiting Gate-2 visual-direction review**.

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
- final color palette roles, typography hierarchy and **vector-first SVG icon language**;
- starfield/background and **2D Galaxy-map** treatment;
- star/system imagery plus the accepted orbital **star-system dialog** visual language (central sun, rocky/gas planets, asteroid belts, colony/fleet presence);
- planet/Colony Detail imagery, buildings/infrastructure presentation rules and construction-queue visual language;
- race/faction portrait direction for currently supported content;
- ship/fleet imagery for location-grouped Fleet tiles and star-system presence, with a preferred **deterministic procedural 2D SVG ship generator** rather than a fixed ship-picture catalog;
- Tactical-battle visual assets for Slice 15.5, such as battlefield backgrounds, ship markers/sprites, selection/movement/target overlays and non-authoritative effects;
- building/infrastructure imagery strategy;
- visual identity for the eight Research categories (Construction, Power, Chemistry, Sociology, Computers, Biology, Physics, Force Fields);
- Diplomacy and Espionage icon/presentation language that can scale with future mechanics;
- UI ornamentation, frames, hover/selection/disabled states and motion language;
- 2D illustration, generated imagery and optional 3D-model -> render/sprite workflows; preserve a future 2D-geometry -> 3D/extrusion path without requiring runtime 3D in this slice;
- asset naming, resolution/DPI/aspect-ratio variants, compression and runtime loading/fallback policy;
- source/provenance/license metadata and clear separation of original MOO2 reference evidence from shippable MOOX assets;
- integration contract between asset catalog and React components.

Defer:

- new gameplay rules or content breadth;
- interactive Tactical mechanics, movement/target legality and battlefield implementation to Slice 15.5;
- broad Race Designer / Ship Designer mechanics;
- expensive cinematic/animated sequences unless explicitly accepted in Gate 2;
- reuse of copyrighted original-game assets without explicit provenance/license permission.

## Gate 1 - Art/asset audit and direction exploration

- [x] Inventory current web visuals plus graphic/palette/catalog capabilities already present in the repo.
- [x] Define visual goals: strategy-first readability, recognizable original MOOX identity, modern sci-fi tone, semantic consistency and nostalgia through functional reference rather than copied pixels; codified in the Gate-2 contract draft.
- [x] Capture vector-first SVG direction and deterministic procedural-ship proof-of-concept contract.
- [x] Capture and browser-prototype the Shipbuilder generate/keep/equip product flow against a fresh real New Game; broader authoritative hull/component/weapon breadth remains Slice 17.
- [x] Add Scout-to-Doom-Star perceptual size scaling plus deterministic concave/negative-space silhouette grammar; Scout remains a Frigate-rules visual role.
- [x] Add a directed-evolution six-candidate Shipbuilder population, deterministic Visual Genome v2 mutation, and sharp primitive geometry inspired by the documented Proceduro reference without vendoring its implementation.
- [x] Add three representative Style-DNA families (Spear/Angular, Sleek/High-Tech, Organic/Alien) with live Shipbuilder previews and evolution-family preservation.
- [x] Add partial Visual-Genome locks for core hull, Wings/Parts primitives, engines and cutouts; browser-verify exact locked-gene preservation and isolated unlocked mutation.
- [x] Replace the primary directed-evolution Shipbuilder UX with a single-click full-random Visual Genome v3 flow and eight coarse morphology families; retain evolution/locks as historical advanced generator research.
- [x] Advance to Visual Genome v4 with strict symmetric half-hull generation/mirroring, remove default asymmetry, expose hull Space, and add explicit Scout-to-Doom-Star preview footprint scaling.
- [x] Persist accepted Visual Genome v4 as authoritative optional ShipDesign/Ship state with separate VisualRevision, immediate save, exact save/restore round-trip and built-ship visual freeze.
- [x] Implement a typed core SVG icon registry and replace shell/navigation/resource plus Colony Farmer/Worker/Scientist placeholder glyphs; browser-verify desktop/mobile scaling.
- [x] Refine Research to a microscope, reserve a test-tube primitive for Chemistry/Pharma categories, put role SVGs on individual population markers, and add selective leading icons to high-value action buttons.
- [x] Produce multiple styleboard/art-direction candidates for the OX+icon brand, shell, 2D Galaxy, star-system dialog, Colony Detail, planet, portrait, Fleet/ship and eight-category Research imagery; live review board at `/styleboard.html`, with Deep Space Command recommended.
- [x] Define target resolutions/aspect ratios and responsive/cropping requirements in the Gate-2 contract draft, including the canonical 320px floor and rich-asset target matrix.
- [x] Define provenance/license rules for generated, commissioned, extracted/reference and third-party assets; original-game/extracted reference material remains non-distributable absent explicit rights.
- [x] Present Gate-2 visual-direction and asset-pipeline contract with representative mockups: live A/B/C styleboard plus functional 7171 acceptance routes.

## Gate-1 completion evidence (2026-09-07)

- Broad functional visual-consistency audit: `docs/research/SLICE_15_3_GATE1_VISUAL_CONSISTENCY_AUDIT_2026-09-07.md`.
- Styleboard candidates and recommendation: `docs/research/SLICE_15_3_STYLEBOARD_CANDIDATES_2026-09-07.md`.
- Live review artifact: `/styleboard.html` on canonical port 7171.
- Gate-2 contract draft: `docs/research/SLICE_15_3_GATE2_VISUAL_ASSET_PIPELINE_CONTRACT_DRAFT_2026-09-07.md`.
- Gate 2 remains **unfrozen** until explicit review/acceptance.

## Gate 2 - Visual identity / pipeline freeze

- [ ] Freeze MOOX visual direction, logo usage, palette, typography and icon grammar.
- [ ] Freeze core asset categories and minimum asset set for the canonical match.
- [ ] Freeze image/model/render pipeline, naming/versioning and fallback policy, including SVG source/runtime rules and modern raster fallback formats.
- [ ] Freeze procedural ship generator grammar: design seed, hull/faction style, module zones, palette and optional per-ship serial variation.
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

- [x] Integrate the eight authoritative Research categories with distinct SVG identities and add first Galaxy/System/Fleet/contact markers with diplomacy tones; browser-verify desktop and 320px mobile on canonical port 7171.

- [x] Integrate system-body, settlement, fleet-role, construction/building and diplomacy icons; replace remaining obvious strategic CSS/text glyphs; browser-verify desktop and 320px mobile.

## Art-fidelity review revision (2026-09-08)

User review accepted the broad Deep Space Command direction but requested materially richer game-world art before Gate-2 freeze. Implemented the first revision block: spherical procedural planets with climate/noise/cloud treatment and rich reusable BuildingArt illustrations. Star Base, Battlestation and Star Fortress are distinct runtime visuals keyed to their authoritative IDs. Live review artifact: `/art-fidelity.html`. Gate 1 remains formally complete; Gate 2 remains **unfrozen** pending fidelity review. See `docs/research/SLICE_15_3_ART_FIDELITY_PLANETS_BUILDINGS_2026-09-08.md`.
