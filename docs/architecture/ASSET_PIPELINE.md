# MOOX asset pipeline

## Goal

MOOX should be able to reproduce the *function and visual information hierarchy* of Master of Orion II while keeping original game artwork separate from the distributable game.

## Three asset layers

### 1. Original reference

Local only:

```text
reference/original/
```

This contains material copied or decoded 1:1 from the legally owned MOO2 1.31 installation. It is ignored by Git and used for fidelity analysis.

Decoded graphics retain mechanical source names first (`archive + block + frame`). A manifest records provenance and hashes. Semantic names are added only after the content is confidently identified.

### 2. Semantic asset specification

Tracked, MOOX-owned metadata describes what an asset *is* and what it needs to communicate. Examples:

- `race.alkari.portrait`
- `planet.terran.background`
- `building.automated_factory.icon`
- `technology.advanced_damage_control.icon`

The tracked semantic asset catalog now carries fields in this form:

```json
{
  "id": "race.alkari.portrait",
  "type": "race_portrait",
  "reference": {
    "archive": "RACESEL.LBX",
    "block": 15,
    "frame": 0
  },
  "requirements": {
    "aspect_ratio": "290:322",
    "subject": "...",
    "mood": "...",
    "ui_readability": "portrait must remain readable at race-selection size"
  }
}
```

The reference coordinates are factual metadata; original image bytes remain local.

### 3. Distributable MOOX artwork

Tracked under:

```text
assets/
```

These assets are independently created or explicitly licensed. Game data refers to semantic asset IDs, not original filenames or generated filenames.

## What can be generated later

The available MOO2 reference images plus normalized game data are enough to prepare consistent new visual sets for, among other things:

- race portraits and small race icons,
- planet/system backgrounds and planet thumbnails,
- ship silhouettes, hull portraits and tactical sprites,
- technologies and research icons,
- colony buildings and production icons,
- diplomacy/event/leader illustrations,
- UI panels, backgrounds and decorative frames.

The data is valuable beyond style reference: race traits, planet class, technology purpose, ship role and building function can be fed into an art specification so new assets reflect actual gameplay semantics.

## Style direction

Do not require pixel-for-pixel reproduction. Preserve useful MOO2 qualities such as immediate readability, strong silhouettes, clear faction identity, information density and a coherent retro-futuristic 4X tone, while allowing higher resolution and a cleaner modern presentation.

A single MOOX art bible should eventually define palette families, framing, lighting, perspective, icon sizes and UI-safe crop zones so generated assets look like one game rather than isolated images.

## Current local reference state

As of 2026-08-26, the canonical private graphics extraction contains 30,395 PNG frames from 30,929 cataloged frames. Palette resolution and decode provenance are recorded in the private graphics manifest; unresolved contexts remain explicit rather than being color-guessed.

### Reference extraction checkpoint

The first complete extraction pass now provides 30,395 local PNG frames backed by a manifest covering all 6,532 recognized graphic blocks; internally paletted raw and Junction frames now decode with zero frame failures. Semantic classification can proceed against this library without moving original-derived bytes into the distributable `assets/` tree.
The graphics catalog now also records explicit palette-context provenance. The first confirmed resolver rules cover 5,646 palette-context blocks / 20,475 frames across verified UI, colony, diplomacy, ship, combat, planet and event contexts; all unverified contexts remain pending rather than being guessed.
The 2026-08-26 canonical clean export materializes 30,395 PNGs (~581.03 MiB) from 30,929 cataloged frames. 30,225 exported frames have complete palette coverage; all 170 partial exported frames are still explicitly `pending`, and no `resolved` context produces a partial frame. The 534 remaining unexported frames are external-palette `pending` contexts. Frame decode failures are zero.
### Semantic race-asset checkpoint

`data/rulesets/moo2-1.31/assets.json` is now the first tracked semantic asset catalog. It contains factual source coordinates and hashes only; no original artwork bytes are committed.

The first slice contains 79 race-related asset records:

- 13 confirmed preset-race portraits from `RACESEL.LBX` blocks 15..27,
- the confirmed custom-race portrait at `RACESEL.LBX` block 28,
- 52 confirmed role icons from the 13x13 `RACEICON.LBX` matrix: farmer, worker, scientist and marine for each preset race,
- 13 deliberately pending generic `race.<id>.icon` keys because the original data exposes 13 role/context variants and no verified source yet identifies one as a universal generic icon.

The portrait sequence is tied to the same canonical race order already normalized in `races.json`. The role-icon block formula is `race_order * 13 + variant`, with variants 1/3/5/7 verified as farmer/worker/scientist/marine by observed community implementation usage and cross-checked against the local matrix. Every confirmed record stores archive, block, frame, dimensions and original block SHA-256.

Generate/validate this layer with:

```powershell
out\moox-analyze-windows-amd64.exe normalize assets `
  -out data\rulesets\moo2-1.31\assets.json `
  -races data\rulesets\moo2-1.31\races.json `
  C:\ASH\Temp\mastori2
```

The next semantic mapping work should extend this catalog with colony buildings, ship/hull graphics, planet classes, technologies and high-value UI assets, always preserving a verification/confidence boundary.