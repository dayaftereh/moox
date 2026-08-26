# MOOX ruleset data

This directory contains normalized, MOOX-owned rule data that the game core can load directly.

Each ruleset lives below a stable ID such as:

```text
data/rulesets/moo2-1.31/
```

The files are generated and/or verified from research sources, but they do not contain copied game binaries, graphics, audio, manual prose or bulk original text.

## Design rules

- Stable machine IDs are authoritative.
- User-visible text is referenced only through `name_key` / `description_key` and resolved from `data/languages/`.
- Gameplay code must never parse translated text to derive rules.
- Every ruleset file has a `schema_version` and `ruleset`.
- Source/provenance information remains in normalized files so parity work can distinguish directly observed original data from secondary-reference values.
- Exact behavior that has not yet been experimentally verified is marked as such instead of being presented as certain.
- Runtime loaders validate schema and referential integrity before accepting a file.

## Current files

### `moo2-1.31/race_traits.json`

Contains the Master of Orion II 1.31 custom-race design model:

- 10 starting Picks and maximum 10 negative Picks,
- 11 selection groups,
- 53 selectable race-design options,
- stable IDs,
- language-independent `name_key` references for every group/option,
- standard-game Pick costs,
- simple numeric modifier values for directly readable stat traits,
- government/ability IDs,
- known incompatibilities,
- per-field verification/provenance metadata.

The game core should consume this file through the Go `internal/ruleset` loader. UI code resolves names/descriptions through `internal/i18n` and the selected file in `data/languages/`.
