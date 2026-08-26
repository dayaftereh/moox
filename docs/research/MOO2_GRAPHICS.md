# MOO2 graphics research

Baseline: 2026-08-26

## Format

MOO2 graphic payloads stored inside normal LBX blocks use a compact indexed-color format. Public reverse-engineering notes describe:

- little-endian width/height,
- frame count and frame delay,
- bit flags,
- `(frame_count + 1)` frame offsets,
- optional internal/mixed palettes,
- 8-bit palette-index pixel runs with relative X/Y indents,
- transparent areas represented by omitted pixels,
- a special Y indent of 1000 marking end-of-frame.

Reference: https://masteroforion2.blogspot.com/2008/04/moo2-graphics.html

The internal palette uses DAC RGB values (0..63) that are scaled to 8-bit RGB by multiplying by four. Internal palettes can replace only part of the 256-entry palette, using a shift + color-count header.

## Pure-Go decoder

MOOX now contains `internal/moo2gfx`, a pure-Go decoder for this format. The first supported path deliberately focuses on blocks with an embedded/internal palette so colors can be reconstructed without guessing which external palette the original executable selected.

CLI example:

```powershell
out\moox-analyze-windows-amd64.exe image `
  -out reference\original\images\racesel\block_015.png `
  C:\ASH\Temp\mastori2\RACESEL.LBX 15
```

The command writes only outside the source installation. Canonical original-reference PNGs remain below the ignored `reference/original/` tree and are not committed. Temporary decoder output may still use `reference/extracted/`.

## Local 1.31 observations

### RACESEL.LBX

Blocks 15 through 28 are 14 single-frame 290x322 graphics with internal palettes (`flags=0x1000`). All 14 currently decode to PNG with zero missing palette pixels.

These are excellent private references for race-selection presentation. Their exact semantic mapping to preset races/custom selection will be established before we give the exported files semantic names; for now filenames retain their source block index.

### PLANETS.LBX

Blocks 0 through 29 are 30 single-frame 640x480 graphics with internal palettes (`flags=0x1000`). All 30 currently decode to PNG with zero missing palette pixels.

### Current private export set

The analyzer generated:

- 14 `RACESEL` PNGs under `reference/original/images/racesel/`,
- 30 `PLANETS` PNGs under `reference/original/images/planets/`.

This initial 44-image subset was the first visually verified seed. It has since been superseded by the complete internal-palette batch extraction described below. All original-derived PNGs remain ignored by Git.

## Broader local scan

A structural scan of all standard LBX containers in the local 1.31 installation found 6,532 plausible MOO2 graphic blocks. 713 advertise an internal palette and are candidates for color-correct decoding without first resolving an external palette context.

The first 44 promoted PNGs were only a seed set; the current private reference library now covers every internally paletted frame that the decoder can reconstruct.


## Batch extraction

`moox-analyze graphics` scans all LBX archives, records every structurally recognized graphic block, and exports every frame for blocks that provide an internal palette. The manifest keeps external-palette graphics visible as pending work rather than dropping them from the reference inventory.

## External palettes

Many useful graphics, including `RACEICON.LBX`, do not embed a full palette. Public format notes state that MOO2 also uses external palettes, including palette blocks in `FONTS.LBX` / `IFONTS.LBX`, and that the correct external palette choice can be context-dependent/hard-coded by the original engine.

The CLI therefore currently refuses normal PNG export for graphics without an internal palette rather than silently producing incorrect colors. External/mixed palette context resolution is the next graphics-decoder task.

## Future asset strategy

The extracted original artwork is a reference for fidelity, not distributable MOOX content. The eventual product should either use independently created artwork or have an explicit licensed compatibility/content strategy.

## Complete internal-palette decoder checkpoint - 2026-08-26

The full local official 1.31 scan now completes with zero frame decode failures.

Observed inventory:

- 363 real LBX archives scanned,
- 6,532 structurally recognized graphic blocks,
- 30,929 frames represented in the manifest,
- 713 graphic blocks advertise an internal palette,
- 5,819 graphic blocks depend on an external palette/context,
- 10,655 PNG frames exported from internal-palette blocks,
- 9,797 exported frames have complete palette coverage,
- 858 exported frames have partial/mixed palette coverage,
- 0 frame decode failures,
- current private PNG footprint: approximately 414.35 MiB after Junction frames are materialized as complete display frames.

### NoCompression

Local 1.31 data resolves the previously unknown `0x0100` flag. `FlagNoCompression` frames are raw row-major 8-bit palette-index buffers with no compressed-frame start indicator or run headers. The pixel payload is `width * height` bytes and the stored frame length is padded with zero bytes to a 4-byte boundary: `align4(width * height)`.

A complete audit of all 43 local NoCompression blocks / 213 frames confirms this alignment rule: every stored frame length equals `align4(width * height)` and every alignment byte observed is zero. This includes 31x31 `BUFFER0` frames stored as 961 pixel bytes plus 3 zero padding bytes, as well as the former failures in `CMBTSFX.LBX` and `RACEOPT.LBX`. The decoder validates both the aligned length and zero padding.

### Junction

`FlagJunction` (`0x2000`) marks delta animations. Frame 0 establishes the image and later frames contain only changed pixels. The display frame is produced by compositing each delta over the previous result.

Local inventory contains:

- 250 Junction graphic blocks,
- 3,099 Junction frames total,
- 39 internally paletted Junction blocks,
- 1,339 internally paletted Junction frames now exported as cumulative display frames.

This especially affects the `SR_R*_SC/SP/TR.LBX` animation archives. Their later frame payloads are small deltas, but their PNG references are now full 640x480 display states rather than sparse patches.

### Remaining graphics work

The raw frame codec is no longer the limiting factor. The evidence-gated palette resolver now resolves 5,646 contexts; 844 remain pending because their exact runtime/base palette is not yet known or they are partial source/carrier records. Functional-color behavior used for effects such as dynamic shadows/transparency also remains a separate fidelity task.

See `MOO2_PALETTES.md` for the evidence registry and palette-resolution plan.
## Canonical graphics extraction checkpoint - 2026-08-26

The batch exporter now has an evidence-gated palette-context resolver. It records palette context status, source, evidence URL and confidence in manifest schema v2, and exports an externally paletted frame only when an enabled rule supplies a known context.

The first confirmed rules add:

- 360 `BLDG0.LBX` blocks / 360 frames using `FONTS.LBX#2`,
- 196 `COUNCIL.LBX` blocks / 1,258 frames using the complete internal palette carried by `COUNCIL.LBX#0`.

A complete local 1.31 export now reports:

- 30,395 PNG frames exported,`n- 30,225 exported frames with complete palette coverage,`n- 170 exported frames with partial palette coverage, all belonging to `pending` contexts,`n- 534 external-palette frames still unexported/pending,`n- 5,646 palette contexts resolved automatically,`n- 844 palette contexts still pending,`n- 0 frame decode failures,`n- approximately 581.03 MiB of private PNG references.

An independent end-to-end check of `BLDG0.LBX` block 0 through the explicit `image -palette-file FONTS.LBX -palette-block 2` path produced the same SHA-256 PNG as the automatic batch resolver (`895A3C64102716349E0FA388C381280A12C5C2F691C7D8EC5CE4E8C85A852629`).

Community-derived palette mappings are documented in `MOO2_PALETTES.md`, but they remain disabled by default until promoted with stronger evidence.
### Expanded verified palette contexts

MoO2 Workshop dependency descriptions provide a second independent per-block source for the official 1.31 data set. After cross-checking its block numbering against the already independently confirmed `BLDG0 -> FONTS#2` and `COUNCIL -> COUNCIL#0` cases, MOOX now also resolves `BLDG1..4`, `RACEICON`, `DESIGN`, `MAINMENU`, `CMBTMISL`, and the mixed `SHIPS` palette groups.

`SHIPS` is especially important: Workshop identifies `FONTS#1` as the base and the local blocks 49/99/149/199/249/299/349/399 as the color-specific high-palette carriers for the preceding 49-image groups. Local carrier RGB ramps and rendered spot checks agree with the eight player colors. This supersedes the older OpenMOO2 `IFONTS#3` assumption.

See `MOO2_PALETTES.md` for exact ranges, provenance and unresolved cases.