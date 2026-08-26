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

Total: 44 current reference PNGs. These are deliberately ignored by Git because they are original copyrighted game artwork used only for local development comparison.

## Broader local scan

A structural scan of all standard LBX containers in the local 1.31 installation found 6,532 plausible MOO2 graphic blocks. 713 advertise an internal palette and are candidates for color-correct decoding without first resolving an external palette context.

The 44 currently promoted PNGs are only the first verified subset, not the limit of available reference artwork.


## Batch extraction

`moox-analyze graphics` scans all LBX archives, records every structurally recognized graphic block, and exports every frame for blocks that provide an internal palette. The manifest keeps external-palette graphics visible as pending work rather than dropping them from the reference inventory.

## External palettes

Many useful graphics, including `RACEICON.LBX`, do not embed a full palette. Public format notes state that MOO2 also uses external palettes, including palette blocks in `FONTS.LBX` / `IFONTS.LBX`, and that the correct external palette choice can be context-dependent/hard-coded by the original engine.

The CLI therefore currently refuses normal PNG export for graphics without an internal palette rather than silently producing incorrect colors. External/mixed palette context resolution is the next graphics-decoder task.

## Future asset strategy

The extracted original artwork is a reference for fidelity, not distributable MOOX content. The eventual product should either use independently created artwork or have an explicit licensed compatibility/content strategy.

## Full private export checkpoint - 2026-08-26

The first complete batch scan/export of the local official 1.31 reference finished successfully.

Observed inventory:

- 363 real LBX archives scanned,
- 6,532 structurally recognized graphic blocks,
- 30,929 frames represented in the manifest,
- 713 graphic blocks advertise an internal palette,
- 5,819 graphic blocks depend on an external palette/context,
- 10,646 PNG frames exported from internal-palette blocks,
- 9,795 exported frames have complete palette coverage,
- 851 exported frames have partial palette coverage,
- 9 internal-palette frames are not yet decoded,
- current private PNG footprint: approximately 135.46 MiB.

The 9 decode failures are narrowly grouped:

- `CMBTSFX.LBX` block 1, frames 0..6: start indicator `0`,
- `RACEOPT.LBX` blocks 0 and 4, frame 0: start indicator `257`.

These are retained in the manifest as `decode_failed`; they are not silently skipped. The partial-palette group is also explicitly tagged. `BUFFER0.LBX` accounts for 640 of the 851 partial frames, with smaller groups in `SPHERSFX`, `BEAMS`, `CMBTSFX`, `RACESEL`, `PLANETS`, animation archives and a few UI archives.

This checkpoint means the private reference library is now broad enough to start systematic semantic classification while palette/codec edge cases continue to be improved.