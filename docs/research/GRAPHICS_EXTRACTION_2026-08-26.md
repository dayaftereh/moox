# Full MOO2 1.31 graphics extraction - 2026-08-26

Source: local official Master of Orion II 1.31 reference at `C:\ASH\Temp\mastori2`.

## Batch result

The pure-Go `moox-analyze graphics` command scanned every normal LBX archive and rebuilt the canonical private reference directory at:

```text
reference/original/images/
```

Result:

- 363 normal LBX archives scanned,
- 6,532 structurally recognized graphic blocks,
- 30,929 total frames described by those blocks,
- 713 blocks advertise an internal palette,
- 5,819 blocks require an external palette/context,
- 10,646 PNG frames exported from the internal-palette set,
- 9,795 exported frames have complete palette coverage,
- 851 exported frames are marked `partial_palette`,
- 9 frames could not be decoded by the normal sparse-frame decoder,
- exported PNG payload: approximately 135.46 MiB,
- private manifest: approximately 7.11 MiB.

All PNGs and the manifest are ignored by Git.

## Partial palettes

A block can advertise an internal palette while still using pixel indices not defined by that embedded palette. Those frames are retained as reference PNGs with unresolved pixels transparent and are marked `partial_palette`; they are not treated as color-complete originals.

Largest partial-palette groups observed:

- `BUFFER0.LBX`: 640 frames,
- `SPHERSFX.LBX`: 35,
- `BEAMS.LBX`: 33,
- `TANM_110.LBX`: 30,
- `CMBTSFX.LBX`: 29,
- `TANM_001.LBX`: 19,
- `TANM_044.LBX`: 18,
- `RACESEL.LBX`: 11,
- `PLANETS.LBX`: 10.

These are candidates for mixed/external palette context resolution.

## Nine normal-decoder rejects

The normal frame decoder currently rejects these frame-start indicators:

- `CMBTSFX.LBX`, block 1, frames 0..6: indicator `0`,
- `RACEOPT.LBX`, block 0, frame 0: indicator `257`,
- `RACEOPT.LBX`, block 4, frame 0: indicator `257`.

The public MOO2 graphics description states that the two-byte frame beginning indicator is expected to be `1`; when it is not, the frame should be skipped. These nine records therefore remain explicitly marked `decode_failed` rather than interpreting unknown data as normal pixel sequences.

Reference: https://masteroforion2.blogspot.com/2008/04/moo2-graphics.html

## Manifest provenance

`reference/original/images/manifest.json` records for every recognized graphic:

- source archive SHA-256,
- source block index and block SHA-256,
- width/height,
- frame count and delay,
- flags,
- palette mode/embedded palette coverage,
- per-frame decode status,
- PNG path and PNG SHA-256 when exported.

External-palette frames remain cataloged as `external_palette_pending`, so the reference inventory does not lose them merely because their palette context is not solved yet.

## Next extraction work

1. Catalog/export known external palettes from `FONTS.LBX` / `IFONTS.LBX`.
2. Establish archive/block-to-palette context mappings.
3. Re-run partial/external graphics with resolved palette contexts.
4. Semantically map portraits, planets, ships, buildings, technology/UI imagery.
5. Keep semantic metadata in Git while original decoded pixels remain private/ignored.
