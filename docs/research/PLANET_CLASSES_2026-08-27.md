# MOO2 planet-class normalization - 2026-08-27

## Scope

This checkpoint normalizes the original-observed planet classification primitives needed by the first MOOX simulation layer. It deliberately does not cover planet graphics, full galaxy generation, population-capacity formulas, specials, artifacts, natives or terraforming.

## Verified 1.31 evidence

The decoder verifies the relevant `Orion2.exe` code ranges by SHA-256 before accepting table data and resolves canonical English names from `ESTRINGS.LBX` block 0 using the original runtime string indices.

Normalized classes:

- sizes: Tiny, Small, Medium, Large, Huge;
- minerals: Ultra Poor, Poor, Abundant, Rich, Ultra Rich;
- gravity: Low G, Normal G, Heavy G;
- climates: Toxic, Radiated, Barren, Desert, Tundra, Ocean, Swamp, Arid, Terran, Gaia.

Verified primitive values:

| Primitive | Original-observed values |
| --- | --- |
| size roll upper thresholds | `1, 3, 7, 9, 10` |
| mineral base extraction | `1, 2, 3, 4, 5` |
| climate base food/farmer | `0, 0, 0, 1, 1, 2, 2, 1, 2, 3` |

The earlier temporary probes also exposed additional nearby tables (`class_to_mineral`, gravity/climate generation tables and climate modifiers). Those values are not promoted into this checkpoint because their full runtime semantics are outside this ticket.

## Runtime artifacts

- `data/rulesets/moo2-1.31/planet_classes.json`
- 23 canonical English keys merged into `data/languages/en.json`
- decoder: `internal/moo2data/planet_classes.go`
- schema/validation: `internal/ruleset/planet_classes.go`

Regeneration command:

```text
moox-analyze normalize planet-classes -out data/rulesets/moo2-1.31/planet_classes.json -languages-dir data/languages <installation-directory>
```

## Validation

The checkpoint is covered by synthetic decoder/schema tests and was regenerated successfully against the local clean MOO2 1.31 reference. Repository-wide tests are expected to be green after removal of the temporary root probe programs.