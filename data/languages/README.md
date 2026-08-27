# MOOX language data

Runtime rulesets do not contain user-visible names/descriptions as authoritative logic. They reference stable translation keys such as:

```text
race_traits.group.population_growth.name
race_traits.option.creative.name
building.automated_factories.name
technology.anti_matter_drive.name
ship_hull.frigate.name
```

Language files live here and map those keys to Unicode text. The UI/game layer selects a locale and resolves keys at runtime; changing language never changes rules, IDs or save-game identities.

## Current committed coverage

Current key counts:

- `en.json` - **334** keys,
- `de.json` - **64** keys, fallback `en`,
- `fr.json` - **64** keys, fallback `en`,
- `es.json` - **64** keys, fallback `en`,
- `it.json` - **64** keys, fallback `en`.

English currently contains the complete normalized name layer for data that has been committed so far:

- 64 Race Designer group/option keys,
- 13 canonical preset-race names,
- 48 building names,
- 203 technology names,
- 6 military ship-hull names.

The four non-English catalogs currently contain only the 64 verified Race Designer keys and resolve all other current keys through their English fallback. This is intentional: translations are not invented merely to equalize key counts.

## Extracted Race Designer locales

The analyzer generates five MOO2 1.31 Race Designer language files:

- `en.json` - English, `RACESTUF.LBX` block 0,
- `de.json` - German, block 1, fallback `en`,
- `fr.json` - French, block 2, fallback `en`,
- `es.json` - Spanish, block 3, fallback `en`,
- `it.json` - Italian, block 4, fallback `en`.

Block 5 is a byte-identical duplicate of English block 0 and is not emitted as a separate locale.

## Original glyph encoding

MOO2's localized source bytes use language/font-specific glyph substitutions in ASCII punctuation positions. The analyzer converts only mappings that have been established from unambiguous parallel source words. See `docs/research/LOCALIZATION_LAYOUT_2026-08-26.md`.

Descriptions will use their own keys (`...description`) when their source structures are decoded. Original long-form source text remains private reference material; MOOX runtime localization should be built from stable semantic keys rather than source offsets.

## Expansion policy

When a new normalized dataset is committed:

1. its stable semantic name keys are added to English from verified original source text,
2. other locales get source-backed translations only when their encoding/source mapping is established,
3. otherwise their fallback remains English,
4. gameplay code must never parse translated text to derive behavior.

This keeps localization independent from game logic while allowing language coverage to expand incrementally.
