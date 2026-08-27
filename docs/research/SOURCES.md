# MOO2 Research Source Register

Baseline created 2026-08-26.

This file records useful public sources and what we use them for. It is intentionally link-oriented: copyrighted manuals, screenshots, game binaries and music are not mirrored into this repository merely because they can be found online.

## Primary gameplay references

### Master of Orion II manual (public mirror)

https://manuals.plus/m/5622eabc23dba2bd8c537b7dee95fe7a11c82d6d6607ee13b1b091f07e284617.pdf

Use: primary/manual cross-check for colony-economy terminology and mineral base industry per worker (`1, 2, 3, 5, 8` for Ultra Poor through Ultra Rich). The manual is referenced, not redistributed.
### StrategyWiki - MOO2 calculations

https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Calculations

Use: secondary cross-check for colony production/research/population-income formulas, additive government/morale/leader composition, gravity colonist penalties, and money rounding. Original HELP/executable evidence remains preferred where available.

### StrategyWiki - MOO2 race design options

https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Race_design_options

Use: secondary cross-check for the effective Low-G/Normal-G/High-G 25/50% economy penalty matrix, including the effective High-G-on-Low-G 25% penalty despite the misleading planet UI display.
### StrategyWiki - Master of Orion II: Battle at Antares

- Main gameplay: https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Gameplay
- Starting a game: https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Starting_a_game
- Race design options: https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Race_design_options
- Technologies: https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Technologies
- Warship technologies: https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Warship_technologies
- Warship design: https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Warship_design
- Calculations: https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Calculations

Use: broad mechanics, formulas, cross-checking and terminology.

License note: StrategyWiki states CC BY-SA 4.0 for page content unless otherwise noted. Game screenshots/assets embedded there may have separate copyright status. MOOX docs summarize facts instead of copying page text/images.

### GameFAQs strategy guide

https://gamefaqs.gamespot.com/pc/197873-master-of-orion-ii-battle-at-antares/faqs/72470

Use: detailed tables and cross-checking of technologies, buildings, environments, leaders, ship equipment and combat-related values.

Do not bulk-copy prose/tables; use it to verify factual values against other sources/original-game observations.

## Official/current commercial distribution

### GOG - Master of Orion 1+2

https://www.gog.com/en/game/master_of_orion_1_2

Use: legitimate DRM-free game source for an owned reference installation. GOG also lists owner extras including manuals/reference material and the MOO2 soundtrack.

### Steam - Master of Orion 2

https://store.steampowered.com/app/410980/

Use: legitimate current source for an owned reference installation and current product metadata.

## Historical/release metadata

### MobyGames releases

https://www.mobygames.com/game/182/master-of-orion-ii-battle-at-antares/releases/

Use: release/platform/publisher/developer timeline cross-check.

## Binary formats and reverse engineering

### ModdingWiki - LBX Format

https://moddingwiki.shikadi.net/wiki/LBX_Format

Use: generic SimTex LBX container header, offset table, optional names/descriptions, extractor notes and known game usage.

### MOO2 graphics-format notes

https://masteroforion2.blogspot.com/2008/04/moo2-graphics.html

Use: community documentation of MOO2 palette/graphics structures and history of graphics tooling.

### Orion Nebula - LBX extraction discussion

https://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?p=1329

Use: historical notes about the variety of content stored in MOO2 LBX archives and MoO2 Workshop.

## Open-source code references


### MoO2 Workshop archive/history

https://moo2mod.com/doc/history/moo2_workshop.html

https://moo2mod.com/

Use: independent historical per-block LBX dependency descriptions for the official 1.31 data set, especially exact external/mixed palette relationships. The archived package is inspected only as an external research source; its executable and description files are not copied into MOOX.
### OpenMOO2

https://github.com/mimi1vx/openmoo2

License reported by repository: GPL-2.0.

Use: open-source reimplementation reference. The project explicitly requires original MOO2 `.LBX` data. Its internal layout separates areas such as AI, battle, objects, research and UI.

Important: do not copy source code into MOOX without an explicit licensing decision.

### LbxExtractor

https://github.com/LouisIngenthron/LbxExtractor

License reported by repository: GPL-3.0.

Use: generic .NET LBX extraction reference.

### Master of Orion 2 save/galaxy tooling

https://github.com/danielrh/masteroforion2

Use: possible reference when researching save-game and galaxy state formats.

### moo2_python

https://github.com/agftw/moo2_python

Use: secondary historical implementation/reference to inspect if a specific subsystem is useful.

## Community patch/modding

### MOO2 1.50 manual - config system

https://moo2mod.com/manual/MANUAL_150.html

Use: understand community patch configuration and identify values that became mod-configurable.

Do not mix 1.50 behavior with the official 1.31 target without an explicit ruleset tag.

## Screenshot/UI references

Useful public galleries exist at MobyGames, StrategyWiki and Old-Games-style preservation sites. We currently do **not** mirror their screenshots into Git because game screenshots contain copyrighted art/UI assets.

For high-fidelity UI study, prefer screenshots captured locally from a legitimately owned installation. Store those under an ignored `reference/screenshots/` directory and use them as private development references.

## Manual/reference-card policy

Public archive sites host scans of the original manual. GOG also provides manuals/extras to owners. The manual is valuable for understanding intended rules and UI, but the repository should store a link plus our independently written specification rather than a copied 153-page copyrighted manual.

If a legitimate GOG/physical manual is made available locally, it can be indexed for research without being committed.

## Source confidence labels for future normalized data

When we begin building exact data tables, mark every field with one of:

- `original-observed` - measured/read directly from a legally owned original installation.
- `multi-source-verified` - agrees across independent public sources.
- `single-source` - only one public source found.
- `inferred` - inferred from behavior; needs experiment.
- `modded` - known to come from 1.50/ICE/community rules, not baseline 1.31.

The exact-fidelity database should eventually prefer `original-observed` over secondary documentation.

### Open Watcom V2 LE executable format sources

https://github.com/open-watcom/open-watcom-v2/blob/master/bld/watcom/h/exeflat.h

https://github.com/open-watcom/open-watcom-v2/blob/master/bld/wl/c/loadflat.c

Use: factual interpretation of the bound MZ/LE object/page-map format while reading the original DOS 1.31 `Orion2.exe`. No Open Watcom source code is copied into MOOX.