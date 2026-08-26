# Preset races - normalized MOO2 1.31 checkpoint

Checkpoint: 2026-08-26

## Output

The 13 standard Master of Orion II races are now normalized into:

```text
data/rulesets/moo2-1.31/races.json
```

The file contains no copied HELP prose. Each race contains:

- stable race ID,
- stable `name_key`,
- original source order,
- trait-selection IDs,
- verification level per trait mapping,
- derived Pick total using the directly extracted `RACESTUF.LBX` block-6 costs,
- semantic future asset keys (`race.<id>.portrait`, `race.<id>.icon`),
- exact HELP record index/offset/size/SHA-256 provenance.

## Source hierarchy

Three independent original sources now contribute different facts:

1. `ESTRINGS.LBX` block 0 - canonical singular race names.
2. `HELP.LBX` block 0 records 630..642 - one fixed 1,403-byte help record per standard race; these records state the preset racial characteristics used for the trait mapping.
3. `RACESTUF.LBX` block 6 - direct signed-byte Pick costs for the normalized trait IDs.

The generator checks every HELP race record against its known 1.31 record SHA-256. A changed/reordered reference set therefore fails normalization instead of silently attaching a race's traits to the wrong record.

## Preset mapping

| Order | Race | HELP record | Normalized traits | Derived Picks |
| ---: | --- | ---: | --- | ---: |
| 0 | Alkari | 630 | Ship Defense +50; Artifacts World; Dictatorship | 10 |
| 1 | Bulrathi | 631 | High-G World; Ship Attack +20; Ground Combat +10; Dictatorship | 10 |
| 2 | Darlok | 632 | Spying +20; Stealthy Ships; Dictatorship | 10 |
| 3 | Elerian | 633 | Omniscient; Ship Attack +20; Ship Defense +25; Telepathic; Feudal | 10 |
| 4 | Gnolam | 634 | Fantastic Traders; +1.0 BC; Lucky; Low-G World; Ground Combat -10; Dictatorship | 8 |
| 5 | Human | 635 | Charismatic; Democracy | 10 |
| 6 | Klackon | 636 | Unification; +1 Food; +1 Production; Uncreative | 9 |
| 7 | Meklar | 637 | Cybernetic; +2 Production; Dictatorship | 10 |
| 8 | Mrrshan | 638 | Ship Attack +50; Warlord; Rich Home World; Dictatorship | 10 |
| 9 | Psilon | 639 | +2 Research; Creative; Low-G World; Large Home World; Dictatorship | 10 |
| 10 | Sakkra | 640 | Growth +100%; Subterranean; +1 Food; Spying -10; Large Home World; Feudal | 10 |
| 11 | Silicoid | 641 | Lithovore; Tolerant; Growth -50%; Repulsive; Dictatorship | 10 |
| 12 | Trilarian | 642 | Aquatic; Trans Dimensional; Dictatorship | 10 |

The mapping is marked `help-description-derived`: the source record is directly verified, but the individual trait selections are a semantic interpretation of the behavior stated by that record rather than a decoded preset-race binary table.

## Gnolam and Klackon Pick totals

No missing trait is invented merely to force ten Picks:

- Gnolam currently totals 8 from the traits directly stated by its verified HELP record.
- Klackon currently totals 9.

The other eleven presets total 10 exactly. Gnolam/Klackon remain as observed mappings until a direct preset-race table or runtime evidence proves an additional modifier.

## Localization

`data/languages/en.json` now also contains 13 keys such as:

```text
race.alkari.name
race.darlok.name
race.trilarian.name
```

These names are sourced from `ESTRINGS.LBX`. DE/FR/ES/IT currently fall back to English for the race proper names rather than duplicating an unsupported localized source.

## Artwork status

The normalized races already expose stable semantic asset keys, but no `RACESEL`, `RACERPRT` or `RACEICON` block has been assigned to a specific race yet. Original portrait/icon mapping remains a separate evidence task.
