# MOOX language data

Runtime rulesets do not contain user-visible names or descriptions. They reference stable translation keys such as:

```text
race_traits.group.population_growth.name
race_traits.option.creative.name
```

Language files live here and map those keys to Unicode text. The UI/game layer selects a locale and resolves keys at runtime; changing language never changes rules or save-game identities.

## Current extracted Race Designer locales

The analyzer currently generates five MOO2 1.31 Race Designer language files:

- `en.json` - English, `RACESTUF.LBX` block 0
- `de.json` - German, block 1, fallback `en`
- `fr.json` - French, block 2, fallback `en`
- `es.json` - Spanish, block 3, fallback `en`
- `it.json` - Italian, block 4, fallback `en`

Block 5 is a byte-identical duplicate of English block 0 and is not emitted as a separate locale.

Each file currently contains the same 64 Race Designer keys. Source archive/block/hash provenance is embedded in the language file.

## Original glyph encoding

MOO2's localized source bytes use language/font-specific glyph substitutions in ASCII punctuation positions. The analyzer converts only mappings that have been established from unambiguous parallel source words. See `docs/research/LOCALIZATION_LAYOUT_2026-08-26.md`.

Descriptions will use their own keys (`...description`) when their source structures are decoded. Original long-form source text remains private reference material; MOOX runtime localization should be built from stable semantic keys rather than source offsets.
