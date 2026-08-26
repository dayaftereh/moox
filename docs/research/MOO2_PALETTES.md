# MOO2 palette-context research

Baseline: 2026-08-26

MOO2 graphics can use external, internal, or mixed palettes. A structurally valid graphic is therefore not necessarily color-correct until the palette context selected by the original game is known.

This document records palette relationships separately from the raw graphics decoder so uncertain mappings are never silently treated as facts.

## Palette sources

External 256-color palettes are stored in:

- `FONTS.LBX`, blocks 1..13,
- `IFONTS.LBX`, blocks 1..4.

Each palette occupies the first 1024 bytes of its block as 256 four-byte DAC RGB entries. The local analyzer already extracts these palettes with hashes and swatches under `reference/original/palettes/`.

Primary reverse-engineering reference:

- https://masteroforion2.blogspot.com/2008/04/moo2-graphics.html

## Confirmed context relationships

### BLDG0.LBX

The published MOO2 graphics-format notes explicitly identify `FONTS.LBX` block 2 as the external palette used to render `BLDG0.LBX` building graphics.

Status: **confirmed from public reverse-engineering documentation**.

### COUNCIL.LBX

`COUNCIL.LBX` block 0 is a 24x24 graphic with a complete 256-entry internal palette. Local 1.31 inspection confirms that all following graphic blocks in the archive lack an internal palette.

Historical tool-development notes report that the normal `FONTS`/`IFONTS` external palettes do not render these graphics correctly, while the internal palette from the first image does. This provides a strong archive-local palette-context rule: use block 0's full internal palette for subsequent `COUNCIL.LBX` graphics.

Reference:

- https://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?t=377

Status: **confirmed by local structure and specific historical reverse-engineering observation**.

## Community implementation evidence

The archived GPL-2.0 OpenMOO2 implementation contains an explicit palette table. We use this only as factual research evidence; no GPL source code is copied into MOOX.

Observed mappings in that implementation:

| Graphic context | Base palette |
| --- | --- |
| APP_PICS | FONTS block 2 |
| BUFFER0 | FONTS block 2 |
| COLBLDG | FONTS block 2 |
| COLONY2 | FONTS block 2 |
| COLPUPS | FONTS block 2 |
| COLSUM | FONTS block 2 |
| COLSYSDI | FONTS block 2 |
| GAME | FONTS block 1 |
| INFO | FONTS block 1 |
| MAINMENU | FONTS block 6 |
| OFFICER | FONTS block 2 |
| PLANETS | FONTS block 2 |
| RACEICON | FONTS block 2 |
| SR_R9_SC | FONTS block 2 |
| SHIPS | IFONTS block 3 |

Reference implementation:

- https://github.com/mimi1vx/openmoo2/blob/2cd3c344aed24380390caaaa819bf7a010b8f4a2/oldmess/gui/gui_client.py

Status: **community implementation evidence**. These mappings should be validated against the local 1.31 data and visual/original-game behavior before being promoted to `confirmed` in the extractor.

## SHIPS.LBX mixed palettes

Local 1.31 structure strongly matches historical descriptions of ship-color palette carriers:

- ship graphics are grouped in blocks of 50,
- blocks 49, 99, 149, 199, 249, 299, 349 and 399 are tiny 2x1 internal-palette graphics,
- each of those carries 48 colors starting at palette index 192,
- blocks 413..419 are also tiny 2x1 carriers with 64 colors starting at index 192,
- normal ship graphics around them have no internal palette.

Historical notes describe these tiny graphics as partial palette holders that modify a base palette for the following ship group. The OpenMOO2 implementation used `IFONTS.LBX` block 3 as the SHIPS base palette.

This gives a strong candidate algorithm:

1. load the verified base palette,
2. process `SHIPS.LBX` in block order,
3. when a palette-carrier block is encountered, merge its internal palette range into the current base palette,
4. apply that current mixed palette to following ship graphics until the next carrier.

The first ship group (blocks 0..48) has no preceding local carrier, so its exact initial mixed context must still be validated before automatic bulk export is enabled.

## Junction and functional colors

Palette context is independent from frame reconstruction:

- `Junction` graphics require cumulative frame compositing,
- `Fill background` tells the original game to clear the destination area before drawing,
- `Functional color` is used for effects such as shadows/transparency and is not yet fully reproduced by the reference PNG exporter.

The PNG catalog therefore distinguishes palette coverage from complete rendering fidelity. A frame can have complete RGB palette coverage while still using a functional-color effect that the original engine applied dynamically.

## Resolution policy

MOOX should use an explicit palette-resolution registry with provenance. Every automatic context rule should record:

- target archive/block scope,
- source palette archive/block or source graphic,
- whether a mixed internal range is applied,
- evidence/source,
- confidence/status.

Unknown contexts stay `external_palette_pending`; they must never be rendered with an arbitrary palette merely to increase the PNG count.

## Implemented resolver checkpoint - 2026-08-26

The graphics manifest schema now records palette-context status, source, evidence and confidence per graphic block. Automatic extraction currently enables only the two `confirmed` rules above.

A full local 1.31 manifest-only validation reports:

- 556 palette contexts resolved automatically,
- 1,618 frames covered by those resolved contexts,
- 360 `BLDG0.LBX` blocks / 360 frames from `FONTS.LBX#2`,
- 196 `COUNCIL.LBX` blocks / 1,258 frames from `COUNCIL.LBX#0`,
- 5,934 palette contexts still explicitly pending.

The pending count includes both completely external-palette graphics and internally mixed/partial-palette graphics whose base context is not yet confirmed. Community-evidence mappings listed above are intentionally not enabled by default yet.