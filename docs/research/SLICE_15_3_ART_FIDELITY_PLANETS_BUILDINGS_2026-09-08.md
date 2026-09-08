# Slice 15.3 - Art fidelity follow-up: planets and buildings

Date: **2026-09-08**

Status: **first fidelity block implemented during Gate-2 visual review; Gate 2 remains unfrozen**.

## Feedback incorporated

The semantic SVG pass was readable but still too diagrammatic for major game-world art. The review specifically calls for:

- planets that read as round worlds, not flat icons;
- climate-dependent palettes rather than a single recolored outline;
- deterministic surface variation/noise so worlds are not uniformly flat;
- clouds where appropriate, especially Terran/Ocean;
- buildings/stations that look like actual game objects rather than one generic symbol;
- distinct images for the command-station progression;
- a reusable building-art layer that can later be composed onto a Colony/planet surface.

## Runtime components

### `web/src/components/OrbitalBodyArt.tsx`

Replaces the System-scene body icon as the primary world rendering.

Planet rendering now uses:

- radial sphere lighting;
- atmosphere/rim treatment;
- dark terminator/shadow gradient;
- deterministic `feTurbulence` fractal noise;
- deterministic secondary surface patches;
- climate-specific base/light/dark/feature palettes;
- optional cloud-noise layer;
- stable visual seed from body ID.

Current climate palette coverage:

- Terran;
- Ocean;
- Desert;
- Arid;
- Barren;
- Tundra;
- Arctic;
- Toxic;
- Radiated;
- Swamp;
- neutral fallback.

Cloud layer currently applies to Terran, Ocean, Tundra, Arctic and Swamp. Terran/Ocean explicitly satisfy the review requirement for cloud-capable worlds.

Gas giants now use a separate spherical treatment with:

- horizontal atmospheric bands;
- deterministic turbulence/displacement;
- a storm oval;
- sphere lighting/terminator.

Asteroid belts now render deterministic individual rocks around an elliptical belt instead of a single asteroid icon.

### `web/src/components/BuildingArt.tsx`

Introduces richer vector illustrations on a 160x112 scene canvas. The component is deliberately reusable between:

- Construction catalog;
- selected Construction hero;
- built-building tiles;
- future Colony/planet-surface placement.

Current distinct building/station artwork:

- `capitol`;
- `colony_base`;
- `marine_barracks`;
- `star_base`;
- `battlestation`;
- `star_fortress`;
- generic building fallback.

The command-station progression is intentionally visually different:

1. **Star Base** - lighter orbital ring, docking arms, central hub;
2. **Battlestation** - armored octagonal combat body with heavy pods/gun hardpoints;
3. **Star Fortress** - massive layered fortress with multiple defensive rings and towers.

The repository already has gameplay/rule support for `star_base`, `battlestation` and `star_fortress`; the art mapping follows those exact authoritative IDs.

## Product integration

### System scene

Planet/gas/asteroid `GameIcon` bodies are replaced by `OrbitalBodyArt` while settlement badges remain a compact semantic overlay.

Current larger runtime sizes:

- planet: 38px desktop / 34px compact;
- gas giant: 46px desktop / 41px compact;
- asteroid belt: 50x32 desktop / 44x29 compact.

This is intentionally larger than the previous semantic icon pass so texture, cloud and lighting detail remains visible.

### Construction

Building projects now use `BuildingArt` rather than the 19px generic building glyph:

- catalog building preview: about 68x48 desktop, 60x42 compact;
- selected building hero: full-width scene, minimum 190px desktop / 150px compact;
- non-building projects continue to use the concise semantic icon family.

This creates a useful distinction:

- semantic icons for actions/project categories;
- richer illustrations for game-world objects.

### Colony built-building list

Built buildings now render the same `BuildingArt` source used by Construction. This avoids duplicated artwork and establishes the reuse boundary needed for future planet-surface composition.

## Art Fidelity Lab

A second Vite HTML entry provides a review-only runtime gallery:

`/art-fidelity.html`

It uses the exact production React components rather than duplicate static mockup SVGs.

The page shows:

- nine climate planet examples;
- gas giant;
- asteroid belt;
- Star Base;
- Battlestation;
- Star Fortress;
- Capitol;
- Colony Base;
- Marine Barracks;
- an initial future planet-surface composition concept using the same components.

The review page is not part of gameplay navigation and does not change game rules/state.

## Browser QA

### Current game

On `game-1` System 01:

- all visible planets render `OrbitalBodyArt`;
- homeworld includes two turbulence layers (surface + clouds);
- the colony badge remains present;
- body art is 38x38 desktop;
- no danger banner;
- no horizontal document overflow.

On Colony 53 Construction:

- Capitol, Colony Base, Marine Barracks and Star Base render rich catalog illustrations;
- selected Capitol renders a 254.5x190 rich hero scene in the tested desktop layout;
- non-building choices retain their semantic icons;
- no danger banner;
- no horizontal overflow.

### Art Fidelity Lab

Desktop:

- all 11 orbital-body examples render;
- Terran/Ocean/Tundra/Arctic/Swamp report cloud-noise layers;
- all three command stations render through actual `BuildingArt` runtime components;
- no horizontal overflow.

Compact browser:

- planet gallery resolves to two columns;
- command stations resolve to one column;
- no horizontal overflow.

## Surface-composition boundary

This block does **not** yet make Colony buildings authoritative positions on a planet surface.

The intended next architecture is:

1. keep building identity/art keyed by stable building ID;
2. create a non-authoritative visual layout of built buildings on a Colony surface;
3. derive placement deterministically from colony/planet/building IDs unless gameplay later needs manual placement;
4. do not make decorative pixel coordinates part of economy/combat authority;
5. reuse `BuildingArt` rather than create separate surface-only illustrations.

## Gate status

Gate 1 remains formally complete as a direction package. This work is a **Gate-2 review revision** prompted by fidelity feedback before the visual contract is frozen.

Gate 2 should not be frozen until the richer planet/building direction is reviewed and accepted.
