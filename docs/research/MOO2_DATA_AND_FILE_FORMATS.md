# Master of Orion II - Data and File-Format Research

Research baseline: 2026-08-26

## 1. Why original data matters

MOO2 exposes a large amount of rules/content through SimTex data archives. A legally owned installation is valuable for three separate purposes:

1. **Behavioral research** - compare visible values and outcomes against MOOX.
2. **Data archaeology** - identify tables for technologies, races, buildings, weapons, leaders, strings, palettes and other content.
3. **Regression fixtures** - store our own normalized facts/hashes/test cases without redistributing the original copyrighted files.

Original files are *inputs to private research*, not assets that should automatically become part of the MOOX repository.

## 2. SimTex LBX container

Public reverse-engineering documentation describes LBX as a SimTex archive/container used by Master of Orion, Master of Magic, Master of Orion II and other SimTex titles.

Known container properties:

- little-endian structure,
- archive begins with a file-count field,
- commonly documented magic/signature bytes: `AD FE 00 00`,
- an offset table points to contained blocks,
- optional short names/descriptions can appear around the 0x200 region,
- blocks are generally not compressed,
- the container itself does not reliably identify each block's semantic type,
- some files with an `.LBX` extension can actually be video files rather than normal LBX archives.

A generic extractor can split the archive, but understanding MOO2 requires mapping `archive + block index -> semantic meaning`.

## 3. Content reported inside MOO2 LBX files

Community reverse-engineering sources report content such as:

- single-frame images,
- multi-frame/animated images,
- palettes,
- fonts,
- WAV audio,
- text-string arrays,
- structured binary data,
- video payloads in files using the LBX extension.

The original engine generally knows what decoder/schema to use based on the specific archive name and block index rather than a self-describing type tag.

## 4. Graphics/palette research

Community graphics documentation describes multiple palette modes and DAC-style RGB values used by MOO2. This is potentially useful if we later want an **analysis viewer** that renders original graphics from a user's local installation for side-by-side reference.

That viewer should be a development tool only. Extracted original images stay in ignored local storage and should not be packaged with MOOX unless redistribution rights are established.

## 5. Useful public code references

### OpenMOO2

Repository: `https://github.com/mimi1vx/openmoo2`

- Open-source MOO2 clone.
- GPL-2.0.
- Python-based.
- Explicitly requires an original MOO2 release and uses the original `.LBX` files.
- Repository structure includes areas for AI, battle, data loading, game objects, research, UI and networking.

Use: architecture/reverse-engineering reference. Do not copy code into a differently licensed MOOX codebase without deliberately accepting GPL obligations.

### LbxExtractor

Repository: `https://github.com/LouisIngenthron/LbxExtractor`

- .NET command-line LBX archive extractor.
- Public repository reports GPL-3.0 licensing.
- It extracts archive blocks but does not interpret every MOO2-specific payload.

Use: cross-check generic archive parsing.

### Master of Orion 2 save/galaxy tools

Repository: `https://github.com/danielrh/masteroforion2`

Publicly described as a Master of Orion 2 save-game/galaxy-related editor/toolset.

Use: investigate save structures and galaxy state once we begin exact save/game-state parity research.

## 6. 1.50 community patch/config system

The MOO2 1.50 community patch documents a text configuration mechanism (`ORION2.CFG` and included config files) used for gameplay modifications. This is useful because it exposes many values in a more human-readable form than raw binary archives and can help distinguish:

- original fixed behavior,
- patch behavior,
- mod-configurable behavior.

For MOOX we should treat **official 1.31 rules** as the first exact-fidelity target and track 1.50/ICE/community differences separately rather than silently mixing them.

## 7. Version baseline

Public historical sources identify:

- `1.31` as the last official MOO2 patch,
- later `1.50.x` community/fan patch lines as unofficial extensions.

The data research layer should record the source version for every extracted fact. A technology or formula from a modded 1.50 installation must not overwrite the 1.31 baseline without an explicit variant tag.

Suggested variant model:

```text
ruleset: moo2-1.31
ruleset: moo2-1.50.22
ruleset: moo2-ice-x
ruleset: moox-default
```

## 8. Local installation status

A local reference installation was subsequently found at `C:\ASH\Temp\mastori2`. Its included `README.TXT` identifies it as **Master of Orion II Version 1.31**, dated 11 April 1997. The set contains 420 files (about 322.38 MiB), including 373 `.LBX` archives, `Orion2.exe`, `ORION95.EXE`, `PATCH13.LBX` and one `.GAM` save. See `LOCAL_REFERENCE_2026-08-26.md` for the non-copyrighted inventory metadata.

No game data was downloaded from abandonware/piracy sites as part of this research.

A legitimate current copy is available commercially through GOG and Steam. GOG additionally lists manuals/reference extras for owners of the package.

## 9. Planned local-reference layout

When the user has installed or supplied a legitimate copy, use a local layout like:

```text
moox/
  reference/
    original/          # ignored - symlink/copy or configured path to owned game
    extracted/         # ignored - raw extracted blocks/images/audio
    catalogs/          # ignored if it contains copyrighted text/data dumps
  data/
    normalized/        # our independently structured factual game data
  tools/
    lbx/
```

Prefer pointing at the installed game rather than copying the entire installation into the project.

## 10. First extraction tool requirements

The first MOOX-owned LBX tool should be intentionally conservative and auditable.

### Phase A - inventory only

Given a path to an owned MOO2 installation:

- enumerate files,
- SHA-256 each file,
- identify likely LBX archives by signature rather than extension alone,
- read file count/info fields,
- enumerate block offsets and sizes,
- read optional names/descriptions if present,
- identify obvious RIFF/WAV, Smacker/video and other signatures,
- write a JSON catalog,
- never modify the source installation.

### Phase B - typed decoders

Implement decoders only when their formats are understood and verified:

- strings,
- palettes,
- still graphics,
- animations,
- audio,
- structured game-rule tables,
- save games.

### Phase C - normalization

Convert factual rule data into MOOX-owned schemas such as:

```text
races.json
race_traits.json
governments.json
planet_types.json
buildings.json
technologies.json
ship_hulls.json
ship_weapons.json
ship_specials.json
leaders.json
random_events.json
```

Every normalized record should retain provenance:

```json
{
  "source_ruleset": "moo2-1.31",
  "source_archive": "...",
  "source_block": 0,
  "verification": "observed-and-cross-checked"
}
```

Do not store original prose, art or audio merely because the extractor can read it.

## 11. Clean-room / IP boundary

For this project, distinguish four categories:

### A. Facts and mechanics

Examples: a weapon's damage value, technology cost, planet-size multiplier, victory threshold. These are the primary reverse-engineering target and can be expressed in independently authored data structures.

### B. Independently authored code

MOOX simulation, UI and tools should be written independently. Public GPL projects can be studied and/or used as separate tools, but copying their code changes licensing obligations.

### C. Original expressive assets

Original sprites, portraits, music, sound, cinematics, manual prose and branded UI art are copyrighted expressive material. Keep them out of Git unless separately licensed.

### D. Trademarks/names

`Master of Orion`, faction names and other branding belong to their respective rights holders. During research we can identify the reference game accurately, but a distributable MOOX product should get its own branding/content strategy before release.

## 12. What to do when a legal MOO2 copy is available

1. Record exact edition/platform/version.
2. Hash all game files.
3. Confirm whether the executable/data are official 1.31 or community-patched.
4. Run non-destructive LBX inventory.
5. Compare archive list against public format documentation.
6. Build a block-type map.
7. Extract only the rule/data structures needed for implementation research.
8. Create automated comparison cases against the running original game.
9. Keep all original/extracted assets ignored locally.
10. Commit only MOOX-authored schemas, observations, tests and documentation.

