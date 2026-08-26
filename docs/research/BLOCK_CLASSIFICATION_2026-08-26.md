# LBX block classification - checkpoint 2026-08-26

The pure-Go block classifier analyzes all 10,498 blocks from the official local MOO2 1.31 reference using structural evidence. Semantic names are used only where the source position/format is independently established; heuristic classes carry a `_candidate` suffix.

Private catalog:

```text
reference/catalogs/block-classification.json
```

## Current classification - complete first pass

- 6,532 `graphic`
- 3,746 `fixed_record_v1`
- 96 `riff_wave`
- 38 `empty`
- 36 `fixed_array_v1`
- 28 `ascii_blob_candidate`
- 17 `external_palette`
- 2 `font_data`
- 2 `signed_byte_table_candidate`
- 1 `fixed_ascii_slots_candidate`
- 0 unclassified/unknown blocks

Every block now has a structural class, but candidate classes are explicitly not semantic guarantees.

## `fixed_record_v1`

A block qualifies when:

```text
uint16 little-endian type == 1
uint16 little-endian body_size == block_size - 4
```

There are 3,746 such records, concentrated in message/dialog archives. The body may contain text, control bytes, placeholders or be empty; the class itself asserts only the binary structure.

## `fixed_array_v1`

A second common structure is:

```text
uint16 little-endian count
uint16 little-endian record_size
record[count] where total size == 4 + count * record_size
```

There are 36 such blocks. Examples discovered directly from the data:

- `HELP.LBX#0`: 707 records x 1,403 bytes
- `HELP.LBX#1`: 9 x 84
- `HERODATA.LBX#0`: 67 x 59
- `RACENAME.LBX#0`: 104 x 20
- `SHIPNAME.LBX#0`: 672 x 16
- `STARNAME.LBX#0`: 13 x 15
- `STARNAME.LBX#1`: 829 x 15
- `SKILDESC.LBX`: fixed arrays for skill names/descriptions
- `TECHDESC.LBX`: fixed arrays for tech/special/weapon names and descriptions
- `ENGMSG.LBX` / `MSGENG.LBX`: fixed arrays with 1,063-byte records

This record-boundary information is particularly useful for later localization and semantic-data decoders.

## Known format classes

- `riff_wave`: verified `RIFF` plus `WAVE` form signature (96 blocks)
- `external_palette`: documented `FONTS.LBX` blocks 1..13 and `IFONTS.LBX` blocks 1..4
- `font_data`: documented block 0 of `FONTS.LBX` / `IFONTS.LBX`
- `graphic`: structurally valid MOO2 graphic header/frame table

## Candidate classes

### `ascii_blob_candidate`

At least 85% printable ASCII with at least one printable run. This includes strong text tables such as `TECHNAME` / `RACESTUF`, but the classifier deliberately does not assign their gameplay semantics.

### `signed_byte_table_candidate`

Small, non-text binary tables whose bytes overwhelmingly fit a small signed-int8 range. There are two:

- `RACESTUF.LBX#6` - separately proven by the Race Designer decoder to be the 53 Pick costs + four zero bytes
- `CUSTMSTR.LBX#3` - related variant, preserved but not yet assigned the same runtime role

### `fixed_ascii_slots_candidate`

`SOUND.LBX#0` is the sole current example. It is 30,720 bytes and fits 20-byte slots; 137 slots are non-empty. 137 of 138 printable string runs start on a 20-byte boundary, and the longest run is 17 bytes, making 20 the smallest compatible slot size among tested divisors. Its exact relationship to the WAV blocks will be established separately.

## Why candidates matter

The goal of this layer is completeness without false certainty. `candidate` means the binary organization is strongly evidenced, while semantic meaning still needs a decoder, cross-reference or runtime observation.
