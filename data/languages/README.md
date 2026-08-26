# MOOX language data

Runtime rulesets do not contain user-visible names or descriptions. They reference stable translation keys such as:

```text
race_traits.group.population_growth.name
race_traits.option.creative.name
```

Language files live here and map those keys to text. The UI/game layer chooses a locale and resolves the key at runtime.

Current baseline:

- `en.json` - English names currently extracted from the official MOO2 1.31 `RACESTUF.LBX` English block.

Future files can include `de.json`, `fr.json`, etc. A translation does not need to change any ruleset or saved-game data.

Descriptions will use their own keys (`...description`) when we begin extracting/authoring them; they are deliberately not embedded into ruleset JSON.

The original MOO2 data already contains several localized text blocks, but some languages use a game-specific character mapping. Those will be decoded carefully before committing locale files rather than preserving mojibake such as custom glyph placeholders.
