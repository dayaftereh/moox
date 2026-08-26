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

Status: **community implementation evidence only**. Several entries agree with later independent evidence, but the `SHIPS -> IFONTS#3` entry does not: MoO2 Workshop explicitly describes `FONTS#1` plus per-group `SHIPS` palette carriers. MOOX therefore does not treat the OpenMOO2 table as authoritative.

## MoO2 Workshop independent evidence

The archived MoO2 Workshop package contains per-block `.lbx.dsc` description files for the official 1.31 data set. The package is used only as an external factual reverse-engineering reference; its executable and description files are not copied into MOOX.

Source/history:

- https://moo2mod.com/doc/history/moo2_workshop.html
- https://moo2mod.com/

Workshop uses the same zero-based LBX block numbering as MOOX. This can be cross-checked against the already independently confirmed `BLDG0.LBX -> FONTS.LBX#2` relationship: Workshop records exactly the same dependency for `BLDG0` block 0 and every other building frame in that archive.

Workshop independently confirms the following archive-wide relationships now enabled in the resolver:

| Graphic archive | Confirmed palette context |
| --- | --- |
| BLDG0 | FONTS block 2 |
| BLDG1 | FONTS block 2 |
| BLDG2 | FONTS block 2 |
| BLDG3 | FONTS block 2 |
| BLDG4 | FONTS block 2 |
| RACEICON | FONTS block 2 |
| DESIGN | FONTS block 5 |
| MAINMENU external-palette blocks | FONTS block 6 |
| CMBTMISL | FONTS block 4 |

Mixed archives such as `CMBTSHP`, `CMBTSFX`, `BEAMS`, `OFFICER`, `BUFFER0`, `GAME`, and `FLEET` contain block-level dependency changes. Their Workshop descriptions are being reduced to independently authored rules/ranges rather than copied wholesale.
## SHIPS.LBX mixed palettes

Local 1.31 structure and MoO2 Workshop's per-block dependency descriptions now resolve the ship palette layout exactly for the standard color groups.

The base palette is `FONTS.LBX` block 1. Each 49-image empire-color group then overrides the high palette range with a tiny internal-palette carrier stored at the end of that group:

| Ship blocks | Palette carrier | Observed color ramp |
| --- | --- | --- |
| 0..48 | SHIPS block 49 | red |
| 50..98 | SHIPS block 99 | yellow/gold |
| 100..148 | SHIPS block 149 | green |
| 150..198 | SHIPS block 199 | gray/white |
| 200..248 | SHIPS block 249 | blue |
| 250..298 | SHIPS block 299 | brown |
| 300..348 | SHIPS block 349 | violet |
| 350..398 | SHIPS block 399 | orange |

The carriers contain 48 colors starting at palette index 192. Their locally observed RGB ramps match the eight expected player banner colors. This also fixes an important directionality detail: holder 49 belongs to blocks 0..48, holder 99 to blocks 50..98, and so on. The carriers do **not** apply to the following group.

Workshop also explicitly describes these special dependencies:

- blocks 400..404 -> `FONTS#1 + SHIPS#413`,
- block 407 -> `FONTS#1 + SHIPS#419`,
- blocks 408, 412, 420, 424 -> `FONTS#1 + SHIPS#414`,
- blocks 409, 421 -> `FONTS#1 + SHIPS#416`,
- blocks 410, 422 -> `FONTS#1 + SHIPS#418`,
- blocks 411, 423 -> `FONTS#1 + SHIPS#415`.

Blocks 405, 406 and blocks without an explicit Workshop dependency remain unresolved. Palette-carrier blocks themselves remain source/reference records rather than being assigned an invented display context.

An isolated end-to-end MOOX scan of only `FONTS.LBX` + `SHIPS.LBX` resolves 408 of 434 external ship graphics and exports 423 frames total when the 15 internal carrier frames are included, with zero decode failures. Visual spot checks confirm that block 50 is rendered with the yellow/gold holder 99 rather than the red holder 49.

The earlier OpenMOO2 `SHIPS -> IFONTS#3` mapping is therefore superseded for MOOX by the more specific Workshop dependency evidence plus local structural/color validation.


## CMBTSHP.LBX mixed combat-ship palettes

MoO2 Workshop describes `CMBTSHP.LBX` as eight regular 45-block groups. In each group the first 44 blocks are combat-ship graphics and the final block is a 32-color internal palette carrier. Local 1.31 headers confirm every carrier uses palette shift 32 with 32 entries.

The exact ranges are:

| Combat-ship blocks | Palette context |
| --- | --- |
| 0..43 | FONTS#4 + CMBTSHP#44 |
| 45..88 | FONTS#4 + CMBTSHP#89 |
| 90..133 | FONTS#4 + CMBTSHP#134 |
| 135..178 | FONTS#4 + CMBTSHP#179 |
| 180..223 | FONTS#4 + CMBTSHP#224 |
| 225..268 | FONTS#4 + CMBTSHP#269 |
| 270..313 | FONTS#4 + CMBTSHP#314 |
| 315..358 | FONTS#4 + CMBTSHP#359 |

This resolves all 352 externally paletted combat-ship blocks, covering 7,040 display frames. The eight carrier blocks contain 20 frames each and remain identifiable as palette-source records.

An isolated `FONTS.LBX + CMBTSHP.LBX` MOOX export produced all 7,200 frames with complete palette coverage and zero decode failures. Visual spot checks of blocks 0, 45 and 90 also show the expected successive player-color groups.

## CMBTSFX.LBX palette ranges

MoO2 Workshop describes explicit base-palette dependencies for 74 of the 79 `CMBTSFX` graphic blocks. MOOX applies these bases while preserving any embedded partial palette entries already carried by the target graphic:

- blocks 2..7 -> `FONTS#4`,
- block 8 -> `FONTS#1`,
- blocks 9..13 -> `FONTS#2`,
- blocks 14..15 -> `FONTS#4`,
- blocks 16..39 -> `FONTS#2`,
- blocks 41..42 -> `FONTS#2`,
- blocks 43..46 -> `FONTS#4`,
- blocks 48..51 -> `FONTS#4`,
- blocks 52..67 -> `FONTS#2`,
- blocks 69..78 -> `FONTS#1`.

Blocks 0, 1, 40, 47 and 68 carry internal partial palettes but have no Workshop base dependency, so they remain conservatively pending. The verified rules cover 1,189 frames.

## BEAMS.LBX palette ranges

`BEAMS.LBX` uses several base palettes plus a local high-range carrier in block 67. Local 1.31 inspection confirms block 67 carries 15 internal colors beginning at palette index 241.

Confirmed rules are:

- 1..16 -> `FONTS#3`,
- 17..32 -> `FONTS#1`,
- 33..48 -> `FONTS#3`,
- 49..64 -> `FONTS#1`,
- 65..66 -> `FONTS#2`,
- 67 -> `FONTS#4` while preserving its internal high colors,
- 68 -> `FONTS#4 + BEAMS#67`,
- 69 -> `FONTS#4`,
- 70..87 -> `FONTS#4 + BEAMS#67`,
- 88..108 -> `FONTS#4`,
- 109..129 -> `FONTS#4 + BEAMS#67`,
- 131..152 -> `FONTS#4 + BEAMS#67`.

Block 0 has its own complete 256-color internal palette. Block 130 is intentionally left pending: Workshop records only `BEAMS#67` there and does not identify the base palette, so MOOX does not infer one from neighboring blocks. The confirmed rules cover 151 blocks / 945 frames.

An isolated `FONTS + CMBTSFX + BEAMS` export produced 2,178 PNG frames with zero decode failures; 2,154 had complete palette coverage and 24 remained partial due only to the explicitly unresolved internal-palette special blocks.

## BUFFER0.LBX verified ranges

MoO2 Workshop assigns `FONTS#1` to five explicit ranges: blocks 1..12, 15..91, 112..121, 132..136 and 142..287. MOOX applies the base palette while preserving any embedded target entries.

This resolves 250 blocks / 2,107 frames. An isolated `FONTS + BUFFER0` export after the aligned NoCompression fix produced 3,473 PNGs with complete palette coverage and zero decode failures. The remaining 38 context records are kept pending because Workshop gives no palette dependency for them.

## OFFICER.LBX verified ranges

Workshop divides the complete officer archive into three palette regions:

- blocks 0..209 -> `FONTS#1`,
- blocks 210..276 -> `FONTS#2`,
- blocks 277..343 -> `FONTS#4`.

All 344 graphic blocks / 498 frames are therefore resolved.

## DIPLOMAT.LBX local full-palette carriers

`DIPLOMAT.LBX` blocks 0..12 each contain a complete 256-entry internal palette. Workshop pairs every following two diplomat graphics with one of those local palette sources: blocks 13..14 use block 0, 15..16 use block 1, continuing through blocks 37..38 using block 12.

MOOX derives that pairing as `(block - 13) / 2`. All 26 externally paletted diplomat blocks / 910 frames resolve from local full palettes without any FONTS guess. End-to-end checks of the first and last pairs produced complete palette coverage.

## FLEET.LBX conservative mixed-palette ranges

Workshop describes:

- blocks 0..44 -> `FONTS#1`,
- blocks 45..81 -> `FONTS#1 + FLEET#111`,
- block 82 -> `FLEET#111` only,
- blocks 83..110 -> `FONTS#1 + FLEET#111`.

Local block 111 is an internal carrier with 176 colors beginning at palette index 0. The combined FONTS+carrier ranges render with complete palette coverage, so MOOX resolves blocks 0..81 and 83..110 (110 blocks / 291 frames).

Block 82 is deliberately left pending: using only `FLEET#111` leaves 409 actually used pixels without a defined palette entry. Workshop identifies the local dependency, but that carrier alone is not a complete rendering context.

## MONSTER.LBX conservative mixed-palette ranges

Workshop contains both FONTS-backed and carrier-only dependencies. Local carrier blocks 13, 14, 15 and 18 each hold 32 colors at palette shift 32.

MOOX enables only rules whose rendered target has complete coverage:

- block 7 -> `FONTS#1 + MONSTER#14`,
- block 8 -> `FONTS#1`,
- block 9 -> `FONTS#1 + MONSTER#14`,
- block 12 -> `FONTS#1`,
- blocks 20..21 -> `FONTS#1`,
- block 24 -> `FONTS#1`,
- block 25 -> `FONTS#1 + MONSTER#13`.

Carrier-only Workshop dependencies remain pending. Empirical decoding showed missing used palette indices for blocks 0, 10 and 11, confirming that the named 32-color carrier by itself is insufficient. This conservative rule keeps those contexts explicit instead of fabricating a base palette.

## MULTIGM.LBX local palette carriers

`MULTIGM.LBX` contains two archive-local palette sources described by Workshop:

- block 0 carries a complete 256-entry internal palette,
- block 42 carries 192 internal colors beginning at index 0.

Workshop maps the external graphics as follows:

- blocks 1..39, 41, 46..149 and 254..260 -> `MULTIGM#0`,
- block 40, blocks 43..45 and 150..253 -> `MULTIGM#42`.

An isolated export confirms that both dependency groups provide complete palette coverage for every target: 259 resolved target blocks / 326 target frames, with zero partial target frames and zero decode failures. Block 42 itself remains a palette-source/context record rather than being assigned a fabricated base palette.

## Additional Workshop-confirmed archive ranges

The same validation pass enables several simple explicit dependencies:

- `GSTAR` blocks 0..22 -> `FONTS#1`, blocks 23..32 -> `FONTS#2` (33 blocks / 262 frames),
- all 80 `CMBTFGTR` blocks -> `FONTS#4` (256 frames),
- all 22 `COLGCBT` blocks -> `FONTS#2` (219 frames),
- all 156 `COLROADS` blocks -> `FONTS#2` (156 frames),
- `COLONY` blocks 5..18 -> `FONTS#2` (14 blocks / 194 frames); blocks 0..4 keep their own/no separately documented context.

An isolated export of `FONTS`, `MULTIGM`, `GSTAR`, `CMBTFGTR`, `COLGCBT`, `COLROADS`, and `COLONY` produced all 1,424 available frames with complete palette coverage and zero decode failures.

## UI and full-palette carrier batch

A further Workshop cross-check identifies several complete local palettes and deterministic UI ranges:

- `PLNTSUM#0` is a full 256-color source for blocks 1..83,
- `RACES#0` is a full 256-color source for blocks 1..63,
- `INFO#1` is a full 256-color source for `INFO#0`, `INFO#2..25`, and `APP_PICS#0`,
- `COLSUM#21` is a full 256-color source for blocks 1..20,
- `COMBAT#11` is a full 256-color source for blocks 0..10, 12..44 and 50..89; blocks 45..49 use `FONTS#4`,
- `DIPSTARS#0..11` use the already verified full palette from `DIPLOMAT#0`.

Workshop also gives direct FONTS ranges that now resolve:

- all documented `MAINPUPS` blocks, using `FONTS#1` except blocks 56, 61 and 73 which use `FONTS#2`,
- all `COLVEGGI` blocks -> `FONTS#2`,
- `COLSYSDI` blocks 0..5 -> `FONTS#1`, 6..65 -> `FONTS#2`,
- `SYSDISP` blocks 0..5 -> `FONTS#1`, 6..61 -> `FONTS#2`,
- all `STARBG` blocks -> `FONTS#1`,
- `GAME` blocks 0..19 and 27..31 -> `FONTS#1`, blocks 20..26 -> `FONTS#6`,
- `COLONY2` blocks 0..49 and 52 -> `FONTS#2`; blocks 50..51 remain undescribed/pending.

Mixed partial-palette contexts are reconstructed only where Workshop explicitly supplies the base:

- `RACESEL` blocks 0..14 and 29..33 use `FONTS#10`; block 32 carries the 128-entry high range used, together with its resolved FONTS base, by blocks 34..137,
- `TECHSEL` blocks 14..22 use `FONTS#1`; blocks 23..27 use `FONTS#1 + TECHSEL#14`,
- `NEWGAME` block 0 and blocks 4..22 use local carrier `NEWGAME#1`; blocks 23..29 use `FONTS#10 + NEWGAME#1`. The carrier-only targets were retained only after end-to-end coverage validation.

An isolated export across these 19 archives produced 2,436 PNGs with zero decode failures. All resolved-context frames had complete palette coverage; the 11 partial frames belonged exclusively to blocks still marked `pending` (`COLONY2#50` and undocumented `RACESEL` special blocks).
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

The graphics manifest records palette-context status, source, evidence and confidence per graphic block. Automatic extraction enables only rules for which MOOX has sufficiently specific evidence and complete target coverage.

The current full local 1.31 manifest-only validation reports:

- 5,582 palette contexts resolved automatically,
- 20,327 frames covered by those resolved contexts,
- 908 palette contexts still explicitly pending,
- 0 structural/frame decode failures.

The resolved set now includes verified archive-wide palettes, mixed ship/combat palettes, full archive-local palettes, tested partial-palette carriers, and deterministic UI/colony/diplomacy range rules.

No frame belonging to a `resolved` context is currently known to have incomplete palette coverage in the isolated validation sets. Carrier-only or undocumented contexts that leave actually used palette indices undefined remain `pending` even when a neighboring/source palette looks visually plausible.

The canonical private PNG tree under `reference/original/images/` has not yet been regenerated for every rule in this checkpoint; manifest-only and focused export validation are used while palette research is still advancing, and a later full export will materialize the accumulated verified contexts in one pass.