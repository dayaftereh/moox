# Slice 15.3 Gate 1 - Body, fleet-role, construction and diplomacy icon pass

Date: **2026-09-07**

Status: **implemented and live-reviewed on canonical port 7171**.

## Scope

This pass continues the shared SVG visual language beyond the Research/Galaxy marker block. It replaces remaining CSS-only symbolic placeholders in high-value strategic surfaces with typed `GameIcon` identities.

Covered surfaces:

1. System orbital bodies and settlement state;
2. Fleet role semantics;
3. Construction catalog / project / current queue semantics;
4. Colony built-building tiles;
5. Construction queue direction controls;
6. Diplomacy stance and action semantics.

## System orbital bodies

The previous CSS-only body dots are no longer the primary body representation. The orbit scene now uses typed SVG identities:

- planet -> `planet`;
- gas giant -> `gas-giant`;
- asteroid belt -> `asteroid-belt`.

Body-size presentation remains differentiated:

- planet icon: 18x18 inside 26x26 reticle;
- gas giant: 24x24 inside 32x32 reticle;
- asteroid belt: 30x18 inside a 38x22 wide reticle.

Planet climate classes now tint the SVG/reticle identity instead of relying only on the obsolete dot background. Current mappings cover Terran, Ocean, Desert/Arid, Tundra/Arctic, Barren, Toxic and Radiated.

Settlement badges are layered directly on the body reticle:

- own colony -> `colonies`, success tone;
- visible outpost -> `outpost`, warning tone.

The selected body gets the existing accent selection treatment around the SVG reticle.

### Live evidence

`game-1` System 04 contains all three supported body forms and was browser-verified:

- System 04 Asteroid Belt 1 -> `asteroid-belt`;
- System 04 II -> `planet`;
- System 04 Gas Giant 3 -> `gas-giant`;
- System 04 Asteroid Belt 5 -> `asteroid-belt`.

System 01 I shows a `planet` body plus a `colonies` settlement badge.

## Fleet role semantics

New role primitives:

- `fleet-combat`;
- `fleet-civilian`;
- `fleet-scout`.

Special strategic vessels deliberately reuse more specific existing identities:

- colony ship -> `flag`;
- outpost ship -> `outpost`;
- troop transport -> `transport`.

Fallback remains the shared `fleets` icon.

The Fleet page now shows the role icon:

- beside the role heading;
- in the ship-count badge.

The System Fleet dialog also puts the same role identity beside each fleet roster entry. Existing procedural ship visuals remain intact; the icon communicates fleet purpose rather than replacing ship appearance.

### Live evidence

Current `game-1` exposes:

- combat fleet -> `fleet-combat`;
- civilian colony-ship fleet -> `flag`.

Both desktop and 320px mobile were verified without overflow.

## Construction category family

The construction catalog no longer uses CSS-generated abstract shapes as its primary glyphs.

Project-kind mapping:

- building -> building-specific icon when known, otherwise `build`;
- housing -> `housing`;
- colony ship -> `flag`;
- outpost ship -> `outpost`;
- troop transport -> `transport`;
- military ship -> `ship`;
- freighter fleet -> `freighter`;
- planetary transformation -> `terraform`.

Known early-game buildings get distinct icons:

- Capitol -> `building-capitol`;
- Colony Base -> `building-colony-base`;
- Marine Barracks -> `building-barracks`;
- Star Base -> `building-star-base`.

Unknown/future buildings safely fall back to the generic `build` identity until their dedicated art is added.

The same mapping is reused in:

- catalog rows;
- large selected-project visual;
- Add to Queue button;
- current-build heading;
- built-building tiles on the Colony surface.

Construction queue Up/Down controls now use shared `arrow-up` / `arrow-down` SVGs instead of text-arrow glyphs.

### Live evidence

Current Colony 53 catalog was verified with these ten visible projects:

1. Capitol -> `building-capitol`;
2. Colony Base -> `building-colony-base`;
3. Marine Barracks -> `building-barracks`;
4. Star Base -> `building-star-base`;
5. Housing -> `housing`;
6. Colony Ship -> `flag`;
7. Outpost Ship -> `outpost`;
8. Troop Transport -> `transport`;
9. Scout / Military Ship -> `ship`;
10. Freighter Fleet -> `freighter`.

Catalog icons measure 19x19. The selected-project hero is 66px desktop and 48px on 320px mobile.

## Diplomacy

New semantic primitives:

- `neutral`;
- `peace`;
- `war`.

Relation badges now combine text + stance SVG and reuse the same strategic relation tone used by Galaxy contacts:

- peace -> friendly/success;
- neutral -> neutral slate;
- war -> hostile/danger.

Diplomacy action buttons also carry their intent icon:

- Declare War -> `war`;
- Offer/Accept Peace -> `peace`.

### Live evidence

The current contact with Darlok is neutral and renders `neutral`; the legal `Krieg erklaeren` action renders `war`.

## Placeholder sweep

After this pass, the strategic TSX sweep leaves only intentional visual structures such as:

- orbit rings;
- procedural ship glyphs;
- typed population SVG containers;
- typed construction/building SVG containers;
- the disabled Ship Designer placeholder control whose label is intentionally functional Slice-17 scaffolding.

No remaining obvious text-arrow / text-X / CSS-only construction glyph was found in the strategic views.

## Browser QA

Canonical review server: **7171**.

Desktop verified:

- System 01 settlement marker;
- System 04 planet / gas giant / asteroid belt differentiation;
- Fleet combat vs colony-special role identities;
- all ten Colony 53 construction catalog mappings;
- building-specific selected-project hero;
- neutral Darlok stance and War action;
- no danger banner;
- no horizontal overflow.

320x646 mobile verified:

- System 04 body icons: planet 18px, gas giant 24px, asteroid belt 30x18;
- Fleet role chips: 16px icon in compact chip;
- construction catalog: 19px icons;
- selected construction hero: 48px;
- document width remains exactly 320px;
- no danger banner.

## Gate-1 status

The visible strategic icon grammar is now substantially broader and closer to a Gate-2 freeze candidate.

Recommended next work:

1. one broad visual-consistency audit across Galaxy, Colony, Fleet, Research, Diplomacy and Espionage;
2. decide whether Espionage's intentionally unavailable state needs a dedicated empty-state visual before Slice 20;
3. address any remaining status/action icon inconsistencies found by the audit;
4. then prepare Slice 15.3 Gate-2 visual grammar freeze rather than inventing more icon variants without a concrete screen need.