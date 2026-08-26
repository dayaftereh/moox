# Non-LBX reference snapshot - checkpoint 2026-08-26

The MOO2 reference extraction now also snapshots every file that is **not** named `.LBX`, preserving relative paths and hashing every copied byte.

Private output:

```text
reference/original/support/
```

## Result

- 47 non-LBX files
- 4.97 MiB total
- 0 SHA-256 mismatches after copying

Extension distribution:

- 16 `.MDI`
- 9 `.DIG`
- 6 `.INI`
- 3 `.EXE`
- 2 `.BAT`
- one each of `.CAT`, `.BNK`, `.MT`, `.GAM`, `.OPL`, `.AD`, `.ba1`, `.SET`, `.LST`, `.COM`, `.TXT`

Subdirectory paths such as `MT32/`, `SB16/` and `SC55/` are retained beneath `files/`.

## Important preserved files

- `Orion2.exe` - 2,644,842 bytes - SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- `ORION95.EXE` - 1,564,160 bytes - SHA-256 `6e19afdc98f1aedcb8d2f974d5b658b0c855f54529bdabdde193f5266e275185`
- `SETSOUND.EXE` - 325,221 bytes - SHA-256 `687fd763c561301bea50d12f52ff83503a9c35deba62da2d1e067e8f9fa15ff1`
- `README.TXT` - 18,735 bytes - SHA-256 `bd6e43b6bf7453408db84178229ccb35779f8143884c71edeaf39ef1f8ca139f`
- `SAVE10.GAM` - 208,000 bytes - SHA-256 `ece2eb06d782078dd0a6f746020a05691355303ceb02bbfbbe2233e987272be1`

These files are original reference material and remain ignored by Git.

## Reference coverage after this checkpoint

The private reference tree now covers the installation through complementary views:

- `original/unpacked/` - every payload block of every normal LBX plus the ten Smacker files
- `original/images/` - decoded graphics catalog / internally-paletted PNG frames
- `original/palettes/` - known external palette data and swatches
- `original/audio/` - lossless verified WAVE blocks
- `original/text/` - structural text records and ASCII-safe previews
- `original/support/` - all 47 non-LBX files including executables and savegame
- `catalogs/` - inventories and structural classification

This is enough to continue reverse engineering from stable local references without modifying or repeatedly probing the original installation tree.
