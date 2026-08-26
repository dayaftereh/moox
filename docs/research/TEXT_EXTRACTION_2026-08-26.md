# MOO2 text catalog - checkpoint 2026-08-26

The text pass intentionally separates raw bytes from readable previews. Raw LBX blocks remain authoritative under `reference/original/unpacked/`; the text catalog adds structural metadata and ASCII-only previews without guessing a character encoding for bytes >= 0x80.

Private output:

```text
reference/original/text/
```

## Result

- 3,746 `fixed_record_v1` records
- 24 `string_table_candidate` blocks
- 3,770 text records total
- 16,227 printable ASCII runs
- 1,222 records also contain bytes >= `0x80`
- 3,513 ASCII preview files plus the manifest (3,514 files total)
- approximately 4.46 MiB private text catalog/previews

## Safety for localization

The preview files are explicitly ASCII-only. They look like:

```text
# TECHNAME.LBX block 0 (string_table_candidate)
# This is an ASCII-only research preview. Raw bytes remain authoritative.
# High bytes are not decoded here.

0x0000  Starting Tech
0x000E  Advanced Biology
...
```

This prevents an early charset assumption from corrupting German/French/other localized data or special game control bytes.

## Strong string-table candidates

The 24 high-confidence candidates include:

- `TECHNAME.LBX` blocks 0..5
- `RACESTUF.LBX` multiple language/string blocks
- `CUSTMSTR.LBX` blocks 0..2
- `PLAYSPEC.LBX` blocks 1..2
- `FILEDATA.LBX` block 1

`TECHNAME.LBX` is particularly useful for the future key-based localization layer because each block contains hundreds of technology-name strings.

## Fixed records

The `fixed_record_v1` body begins at byte 4. The manifest stores body size/hash, non-zero count, high-byte count, printable ratio and safe ASCII runs. Preview offsets are converted back to original block offsets.

No fixed-record body is normalized into MOOX runtime language data yet. The next step is to establish which source blocks/locales correspond to each other and then map them to stable keys; the ruleset must continue referencing only language keys, never source text positions.
