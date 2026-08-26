# MOO2 text catalog - checkpoint 2026-08-26

The private text pass separates raw bytes from readable previews. Raw LBX blocks remain authoritative under `reference/original/unpacked/`; the text catalog adds structural metadata and ASCII-only previews without guessing a character encoding for bytes >= 0x80.

Private output:

```text
reference/original/text/
```

## Current result (schema v2)

- 3,746 `fixed_record_v1` blocks
- 32 text-bearing `fixed_array_v1` blocks
- 23 free-form `string_table_candidate` blocks
- 3,801 text-bearing blocks total
- 3,449 individually indexed fixed-array records
- 20,923 printable ASCII runs
- 3,545 private files including the manifest
- approximately 6.43 MiB private text catalog/previews

Some structurally valid fixed arrays contain no safe ASCII runs and are intentionally absent from the text-specific catalog even though they remain present in the full block classifier.

## Fixed-array record boundaries

Schema v2 preserves the exact row layout for text-bearing fixed arrays. Every row records:

- record index,
- original block offset,
- record size,
- record SHA-256,
- nonzero/high-byte metrics,
- safe ASCII runs local to that record.

The preview files show record boundaries explicitly.

### HELP.LBX block 0

```text
707 records x 1403 bytes
```

Record 0 contains `No Tech` plus `No description`; subsequent records directly pair item names with their long help text. Examples observed include `Achilles Targeting Unit`, `Adamantium Armor`, `Advanced City Planning`, `Advanced Damage Control`, and their corresponding descriptions in the same record.

This is a much stronger basis for future `name_key` / `description_key` generation than treating the archive as an unordered string bag.

### TECHDESC.LBX

- block 0: 39 records x 38 bytes - names/specials
- block 1: 39 records x 100 bytes - matching short descriptions
- block 2: 40 records x 38 bytes - another name family
- block 3: 40 records x 100 bytes - matching short descriptions

The shared record counts strongly support index-based pairing within each 0/1 and 2/3 pair; semantic naming still belongs in a dedicated decoder rather than this generic layer.

### Other useful fixed arrays

- `HERODATA.LBX#0`: 67 records x 59 bytes, with leader names/roles visible recordwise
- `RACENAME.LBX#0`: 104 x 20
- `SHIPNAME.LBX#0`: 672 x 16
- `STARNAME.LBX#0`: 13 x 15
- `STARNAME.LBX#1`: 829 x 15
- `SKILDESC.LBX#0/#1`: 27 records each, suitable for index pairing of skill name/description data
- `ENGMSG.LBX#0`: 20 x 1063
- `MSGENG.LBX#0`: 384 x 1063
- `FILEDATA.LBX#1`: 85 x 17

## Localization safety

Preview files remain explicitly ASCII-only:

```text
# Raw bytes remain authoritative.
# High bytes are not decoded here.
```

This is intentional. Many source records contain bytes >= `0x80`; converting them prematurely as Windows-1252, CP437, UTF-8, or any other guessed character set would destroy evidence needed to reconstruct the game's localized text mapping correctly.

## Next localization work

1. Identify the locale/order of multi-block tables such as `TECHNAME.LBX` and `RACESTUF.LBX` using only unambiguous source evidence.
2. Establish the original high-byte character mapping, likely with help from the font data and parallel localized strings.
3. Build stable MOOX language keys from semantic record IDs, not archive offsets.
4. Keep original source strings private; commit only MOOX-owned normalized localization data needed by the implementation.
