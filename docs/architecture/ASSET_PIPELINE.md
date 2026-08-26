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

A future asset catalog can carry fields such as:

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

As of 2026-08-26:

- 6,532 LBX blocks look structurally like MOO2 graphics,
- 713 of those advertise an internal palette and are candidates for color-correct decoding without external palette context,
- 44 verified PNG references are currently stored under `reference/original/images/`:
  - 14 from `RACESEL.LBX`,
  - 30 from `PLANETS.LBX`.

A batch catalog/export command now inventories all recognized graphics and exports internal-palette frames. The next graphics tooling improvement is external-palette resolution, followed by semantic mapping of the resulting private reference library.
