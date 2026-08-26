# MOO2 building graphics mapping

Baseline: official Master of Orion II 1.31, local reference `C:\ASH\Temp\mastori2`.

This note records the evidence used to map the 48 standard colony-building IDs to the `BLDG0.LBX` .. `BLDG4.LBX` colony-screen graphics. The mapping is derived from the original 1.31 executable and original LBX data, not inferred from image order.

## Bound MZ/LE executable

`Orion2.exe` is a DOS/4GW-bound executable containing an inner MZ module whose `e_lfanew` points to an LE executable. The pure-Go `internal/moo2exe` reader resolves the bound module, LE object table and LE page map so original data/code can be read by object number and virtual offset.

For the local official 1.31 reference:

- `Orion2.exe` SHA-256: `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`,
- inner MZ module begins at file offset `0x26654`,
- LE header is at file offset `0x292E4`,
- LE page size is 4096 bytes,
- object 1 contains executable code,
- object 2 contains the building table used below.

Open Watcom's LE definitions/loader source were used to verify the container/page-map interpretation. No Open Watcom or original game code is copied into MOOX.

## Original building table

The original debug symbols identify `_buildings` in LE object 2 at offset `0x6B3D`.

Observed table structure:

- 49 records,
- 19 bytes (`0x13`) per record,
- record 0 is the dummy/no-building record,
- records 1..48 are the standard building IDs,
- each record carries its own building ID at offset `+4`,
- each record carries its technology ID at offset `+6`.

`N_Bldgs_` independently iterates building IDs 1 through 48. The normalized `buildings.json` therefore reads its production/building ID and technology ID directly from this original table and cross-checks the technology ID against the 203-entry original `TECHNAME.LBX` technology catalog.

## Original BLDG archive/group formula

The original debug symbols identify `Cache_Load_Bldg_` in LE object 1 at offset `0x9F6DC`.

Its 1.31 instruction bytes are hash-validated by the asset normalizer before the building graphic mapping is generated:

- code range: object 1 `0x9F6DC..0x9F741`,
- SHA-256: `c3f6d4f657563a83c6f9b2643dbd5baa711f1019eaef66576de1fcc8296e7162`.

The function implements the following mapping for a standard `building_id`:

```text
zero = building_id - 1
archive_index = zero / 10
group_within_archive = zero % 10
base_block = group_within_archive * 36
```

The archive name is produced from the original `ESTRINGS.LBX` block 0 string `BLDG%D.LBX` at offset `0xE0F`.

Therefore IDs map to archives as:

- 1..10 -> `BLDG0.LBX`,
- 11..20 -> `BLDG1.LBX`,
- 21..30 -> `BLDG2.LBX`,
- 31..40 -> `BLDG3.LBX`,
- 41..48 -> `BLDG4.LBX` groups 0..7.

## Original 6x6 position/frame formula

`Cache_Load_Bldg_` calls the original `Bldg_Coords_To_Effective_Frame_`, found in LE object 1 at offset `0xAC8A6`.

The asset normalizer also hash-validates this function:

- code range: object 1 `0xAC8A6..0xAC8CE`,
- SHA-256: `4a3023201766b2797af90d45149739dc912498cee72f7ba929e05ce7b724c678`.

For colony-grid coordinates `x,y` in `[0,5]`, the original function maps the 6x6 grid into a serpentine effective frame:

```text
if y is even:
    effective_frame = y * 6 + x
else:
    effective_frame = y * 6 + 5 - x
```

The final LBX block is:

```text
block = group_within_archive * 36 + effective_frame
```

Each standard building therefore has exactly 36 colony-screen graphic variants.

## Semantic asset representation

`data/rulesets/moo2-1.31/assets.json` schema 2 stores one semantic asset per building:

```text
building.<building-id>.colony
```

Each building asset contains 36 confirmed variants. Every variant records:

- `grid_x`,
- `grid_y`,
- `effective_frame`,
- source BLDG archive,
- source LBX block,
- frame 0,
- original block SHA-256,
- expected 640x480 dimensions.

Across the 48 standard buildings this produces 1,728 independently hash-addressed references.

## Deliberately unresolved BLDG content

`BLDG4.LBX` contains one additional 36-block group at blocks 288..323. The original standard-building formula for IDs 1..48 never reaches this group. It is therefore **not** assigned to a building or given a semantic key until its purpose is independently identified.

`BLDG5.LBX` has no normal graphic payload in the local reference.

This boundary is intentional: the analyzer does not infer a 49th standard building simply because another 36-image group exists.
