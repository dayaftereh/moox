# LBX block classification - checkpoint 2026-08-26

The pure-Go block classifier analyzes every block from the official local MOO2 1.31 reference using structural evidence only. Archive names are not used to assign semantic meaning except for the already documented external palette block positions in `FONTS.LBX` / `IFONTS.LBX`.

Private catalog:

```text
reference/catalogs/block-classification.json
```

## Current classification

10,498 LBX blocks total:

- 6,532 `graphic`
- 3,746 `fixed_record_v1`
- 96 `riff_wave`
- 24 `string_table_candidate`
- 17 `external_palette`
- 38 `empty`
- 45 `unknown`

The private JSON catalog is approximately 5.38 MiB and records block hashes, first bytes, printable-ASCII metrics and classification evidence.

## `fixed_record_v1`

This class is structural, not semantic. A block qualifies when:

```text
uint16 little-endian type == 1
uint16 little-endian body_size == block_size - 4
```

The body is therefore a fixed-size record following a four-byte header. Printable text metrics are recorded independently.

The class is heavily concentrated in these archives:

- `JIMTEXT.LBX`: 498
- `KENTEXT.LBX`: 480
- `JIMTEXT2.LBX`: 480
- `BILLTEXT.LBX`: 468
- `DIPLOMSG/S/E/F.LBX`: 180 each
- `KENTEXT1.LBX`: 174
- `BILLTEX2.LBX`: 168
- `EVENTMSG/S/E/F.LBX`: 152 each
- `COUNCMSG.LBX`: 84
- `ANTARMSG.LBX`: 48

A representative record:

```text
JIMTEXT.LBX block 0
header: 01 00 64 00
body_size: 100
block_size: 104
body begins: "Type ORION2 to run the game!"
```

Another example:

```text
EVENTMSE.LBX block 0
header: 01 00 2C 01
body_size: 300
block_size: 304
body begins with a GNN message
```

The four large header families observed earlier are therefore not arbitrary headers: the second 16-bit value is exactly the fixed body size (100, 5202, 300, 200 bytes respectively).

## RIFF/WAVE

All 96 blocks starting with `RIFF` were independently verified to have `WAVE` as the RIFF form type. No non-WAVE RIFF block was found.

They occur in:

- `SOUND.LBX`: blocks 1..68 (68 WAV records)
- `STREAM.LBX`: 8 WAV records
- `STREAMHD.LBX`: blocks 1..20 (20 WAV records)

These can be copied losslessly to `.wav`; no audio transcoding is required.

## Remaining unknowns

Only 45 blocks remain `unknown`. The largest groups by archive currently include:

- `HELP.LBX`: 17
- `TECHDESC.LBX`: 4
- `STREAM.LBX`: 3
- `RACESTUF.LBX`: 2
- `SKILDESC.LBX`: 2
- `PLAYSPEC.LBX`: 2
- `STARNAME.LBX`: 2
- several single-block cases.

`unknown` is intentional: the classifier does not infer a type merely from a suggestive filename.

## Next passes

1. Copy the 96 verified RIFF/WAVE blocks into a private audio reference tree.
2. Decode/export `fixed_record_v1` bodies into a private language/message corpus while preserving block provenance and raw bytes.
3. Investigate the 24 string-table candidates separately.
4. Work the 45 unknown blocks down with explicit format evidence.
