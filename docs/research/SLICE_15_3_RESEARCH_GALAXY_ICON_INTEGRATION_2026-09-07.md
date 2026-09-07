# Slice 15.3 Gate 1 - Research category and Galaxy marker integration

Date: **2026-09-07**

Status: **implemented and live-reviewed on canonical port 7171**.

## Scope

This is the first screen-specific integration pass after the shared `GameIcon` foundation.

It covers two visible systems:

1. the eight authoritative Research categories;
2. first strategic Galaxy/System/Fleet/contact markers.

## Research category family

The global Research identity remains the microscope (`research`). The eight server-provided Research categories now get a distinct category identity based on their authoritative order:

| Order | Category | Icon |
| ---: | --- | --- |
| 0 | Construction | `research-construction` |
| 1 | Chemistry | `test-tube` |
| 2 | Computer | `research-computer` |
| 3 | Physics | `research-physics` |
| 4 | Energy | `research-energy` |
| 5 | Sociology | `research-sociology` |
| 6 | Biology | `research-biology` |
| 7 | Force Fields | `research-force-field` |

The geometry uses the existing 24x24 / currentColor / 1.55-stroke SVG grammar.

The category icons are visible in both Research surfaces:

- full strategic Research page: 25x25 desktop / 22x22 mobile in a compact category tile;
- classic Research overlay: 15x15 inside each field bar;
- non-choose-one field action buttons reuse the category icon;
- overlay close now uses the common `close` SVG instead of a text glyph.

The Chemistry category deliberately uses the previously reserved `test-tube` primitive. Global Research still uses the microscope.

## Galaxy / system markers

The previous Galaxy system center was an 8px CSS dot. It is now the shared `star-system` SVG at 18x18.

Each system can additionally show compact 14-15px marker pills to its upper-right:

- own colony -> `colonies`, success tone;
- own outpost -> `outpost`, warning tone;
- own fleet -> `fleets`, player/accent tone, with unit count when greater than one;
- known foreign colony/outpost/fleet contact -> icon by contact kind plus diplomacy tone.

Diplomacy tones are derived from the authoritative player decision view:

- own -> player/accent;
- peace -> friendly/success;
- neutral/unknown -> neutral slate;
- war -> hostile/danger.

Duplicate contacts with the same empire + contact kind are collapsed before rendering, so repeated fleet contacts do not create redundant icons.

## System dialog

The selected system dialog now:

- uses `star-system` beside the system title;
- uses the same icon vocabulary in the contact/traffic badges;
- uses the same friendly/neutral/hostile relation tone as the Galaxy marker.

This keeps map and drill-down semantics aligned.

## Live QA

Review target: `http://127.0.0.1:7171/`, existing persistence-enabled `game-1`.

### Research desktop

Observed all eight categories and exact registry identities:

- Construction -> `research-construction`
- Chemistry -> `test-tube`
- Computer -> `research-computer`
- Physics -> `research-physics`
- Energy -> `research-energy`
- Sociology -> `research-sociology`
- Biology -> `research-biology`
- Force Fields -> `research-force-field`

All full-page category icons measured 25x25. No danger banner and no horizontal overflow.

### Research overlay

All eight field bars expose the same category identities at 15x15. The common close icon is present. On 320x646 mobile the field grid resolves to one 281px column and document width remains exactly 320px.

### Galaxy desktop

The current test game exposes useful real data:

- one own-colony marker;
- one own-fleet marker;
- System 20 exposes one neutral foreign-colony marker and one neutral foreign-fleet marker after deduplication;
- every system center uses `star-system` at 18x18.

Opening System 20 shows the same `star-system` title icon and two deduplicated neutral traffic badges (`colonies` and `fleets`).

### Galaxy mobile 320x646

- system star icons remain 18x18;
- status/contact markers resolve to 14x14;
- document width remains exactly 320px;
- no danger/error banner.

## Gate-1 status

This is still Gate-1 visual integration, not the final Gate-2 icon grammar freeze.

Recommended next pass:

1. richer planet/body markers and colony/outpost state inside the system scene;
2. fleet-role differentiation (combat, scout, colony, outpost, transport);
3. construction/building category family;
4. final broad icon consistency review before Gate-2 freeze.

## Body/Fleet/Construction/Diplomacy follow-up (2026-09-07)

The next Gate-1 pass replaced System body dots with planet/gas-giant/asteroid SVGs plus settlement badges, differentiated Fleet roles, replaced Construction CSS glyphs with typed project/building identities, moved queue arrows to SVG, and added Diplomacy stance/action symbols. See `docs/research/SLICE_15_3_BODY_FLEET_CONSTRUCTION_DIPLOMACY_ICON_PASS_2026-09-07.md`.
