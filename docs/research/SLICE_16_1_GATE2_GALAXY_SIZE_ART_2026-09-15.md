# Slice 16.1 Gate 2 - Galaxy Size art pass

Date: 2026-09-15
Status: **review candidate; Gate 2 remains open**

## Goal

Replace the generic CSS geometry in the New Game Galaxy Size selector with five original, astronomy-informed images that make relative size immediately legible without pretending the current server supports every option or that galaxy morphology maps one-to-one to game-generator size.

## Astronomy research used for the visual grammar

Primary references:

- NASA Hubble, **POX 186: A Tiny Galaxy is Born**: POX 186 is described as an extremely small dwarf galaxy, about 900 light-years across and containing about 10 million stars, versus roughly 100,000 light-years and over 100 billion stars for the Milky Way.
  - https://science.nasa.gov/asset/hubble/pox-186-a-tiny-galaxy-is-born/
- NASA Hubble, **Hubble Spies a Diminutive Galaxy IC 3430**: dwarf galaxies are galaxies with fewer stars, usually below a billion, and dwarf versions can occur as elliptical, irregular, spheroidal and even spiral galaxies.
  - https://science.nasa.gov/image-detail/hubble-ic3430-potw2431a/
- ESA/Hubble, **Nature's grand design**: grand-design spirals are characterised by prominent, well-defined arms, while spiral structure varies substantially between individual galaxies; some are patchy, barred, colossal, radiant, dim or diminutive.
  - https://www.esa.int/ESA_Multimedia/Images/2020/02/Nature_s_grand_design
- ESA/Hubble, **Hubble revisits a grand spiral**: NGC 5643 is a grand-design spiral with two large winding arms, reinforcing that arm count is a morphology feature rather than a numeric size scale.
  - https://www.esa.int/ESA_Multimedia/Images/2024/12/Hubble_revisits_a_grand_spiral
- NASA/JPL, **NASA's Galex Reveals the Largest-Known Spiral Galaxy**: NGC 6872 is discussed as a giant barred spiral; its scale is communicated by the extent of the system rather than by simply counting spiral arms.
  - https://www.jpl.nasa.gov/news/nasas-galex-reveals-the-largest-known-spiral-galaxy/

## Visual decision

The selector therefore does **not** encode `Tiny -> Huge` as `2 arms -> 4 arms -> 6 arms`.

Instead, relative game size is communicated through a controlled increase in:

- visible galaxy footprint;
- stellar count/density;
- disk extent;
- core luminosity;
- star-forming knots;
- arm continuity and outer branching;
- overall visual richness.

Morphology remains illustrative. The information dialog explicitly states that the depicted form and arm count are not direct physical size measurements and do not promise the exact generated map shape.

## Five generated assets

All assets are original MOX procedural artwork, generated deterministically from `web/scripts/generate-galaxy-size-art.mjs` at 1200 x 675 with a 16:9 viewBox.

| Option | Visual language | Approx. SVG object richness |
| --- | --- | ---: |
| Tiny | compact dwarf irregular; asymmetric, sparse, a few star-forming knots | 237 circles |
| Small | compact flocculent spiral; fragmented two-arm language | 387 circles |
| Medium | barred spiral; broader disk and denser star field | 597 circles |
| Large | luminous grand-design spiral; broad disk and branching structure | 920 circles |
| Huge | extended giant spiral; brightest core, richest field, broadest disk and outer branches | 1,488 circles |

Runtime paths:

- `/assets/new-game/galaxy-size/tiny.svg`
- `/assets/new-game/galaxy-size/small.svg`
- `/assets/new-game/galaxy-size/medium.svg`
- `/assets/new-game/galaxy-size/large.svg`
- `/assets/new-game/galaxy-size/huge.svg`

## Why SVG instead of PNG/WebP

For this setting selector the art is generated vector/procedural illustration rather than photographic texture. SVG is the better runtime format because:

- the same source remains sharp on ~342 px mobile cards and ~606 px desktop cards;
- no multiple density variants are required;
- the files remain deterministic and reproducible from source;
- star counts, glow structure and morphology can be changed without raster resampling;
- artwork contains no embedded localized UI text.

Painterly race portraits can still use WebP/PNG under the existing format policy.

## Contract and provenance

- The five paths are registered in `web/public/assets/new-game/manifest.json`.
- Provenance is `original-procedural-mox`.
- Telescope imagery was used only as morphology/scale research and is **not copied, traced or shipped**.
- Only `small` remains server-supported today; all other sizes remain browseable preview art with Create Game disabled.

## QA target

Before Gate-2 freeze:

- build and UTF-8 guard must pass;
- all five SVGs must load in the live selector;
- image/card must retain rounded clipping and the overlaid information control;
- 390 px mobile must have no horizontal overflow;
- desktop/tablet scaling must remain crisp;
- option navigation and supported/planned semantics must be unchanged.

## Live QA result

- `npm run build` passes, including the UTF-8/mojibake guard.
- Re-running the SVG generator produces byte-identical SHA-256 results for all five assets.
- All five selector options load their own SVG path and report native dimensions of `1200 x 675`.
- 390 px mobile: artwork `342 x 213.75 px`, 16 px rounded clipping, document scroll width 390 px.
- 1280 px desktop: artwork `606 x 340.875 px`, 20 px rounded clipping, no unexpected horizontal overflow.
- `Small` keeps Create Game enabled; Tiny/Medium/Large/Huge remain preview-only and keep Create Game disabled.
- The information dialog explains the selected illustrative morphology plus the common warning that arm count/morphology are not direct size measurements or generator promises.

This is the current user-review candidate; Gate 2 is not frozen yet.
