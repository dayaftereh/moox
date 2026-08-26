# MOO2 localization layout - checkpoint 2026-08-26

This checkpoint establishes the first directly usable multi-language mapping for MOOX while preserving the original MOO2 byte representation as provenance.

## Parallel language-block order

Two independent MOO2 1.31 archives show the same six-block language ordering:

| Block | Locale | Evidence |
| ---: | --- | --- |
| 0 | English | readable English content |
| 1 | German | readable German content |
| 2 | French | readable French content |
| 3 | Spanish | readable Spanish content |
| 4 | Italian | readable Italian content |
| 5 | English duplicate | byte-identical to block 0 |

### TECHNAME.LBX hashes

- block 0 / block 5: `3079adf60539bd200d0b55b750f52c1e5cd1d2b59bdf4e2243f1c17d6d3978e9`
- block 1: `127054683ccf32af95acbc29be0930d4d7c9b6e9f56b3f02312ff637d844f40e`
- block 2: `b3939fb72d6f8266b90e21910dd0256786e5ed936343b669c2acb17284dc0804`
- block 3: `567d68a2ea41b6cd3ab682dbba498a6dceb30d70da3e3c821019a69e1b570510`
- block 4: `512c6ecd34f2e20dcbc67661b81587c6c489af884be661903c4acc3f859d5e35`

### RACESTUF.LBX hashes

- block 0 / block 5: `5583d8d5744e3a075b2e925282ab7f98f0c40418e3c029e82718136ffd665970`
- block 1: `902ad33556b382e92ae4887a9ad57b0dbf8413fdb44d50f1aec4ffcdf9778cde`
- block 2: `cc5f365e7d34f7fb6a7dc06c1c6a999d019dc0a2911534d92028447323202841`
- block 3: `b32179cf539b67e62f188ea55560d3978c714785c64c6285060f5dedb82e4686`
- block 4: `910ccdb0da34f1b162d19fffb15b370ff06b797983ebd247e807cac3311e90e1`

The duplicated English block is therefore directly established in both archives, not inferred from display text alone.

## Language-specific glyph encoding

Localized source strings do **not** store all accented characters as a standard modern character set. Printable ASCII punctuation byte positions are repurposed by the language/font context.

The following mappings are used for the Race Designer because they are repeatedly confirmed by unambiguous words in `TECHNAME.LBX` and/or `RACESTUF.LBX`.

### German

| Source byte/character | Unicode | Example source -> decoded |
| --- | --- | --- |
| `$` | `Ü` | `$berlebens` -> `Überlebens` |
| `]` | `Ä` | `]tzschleim` -> `Ätzschleim` |
| `{` | `Ö` | `{konomie` -> `Ökonomie` |
| `[` | `ä` | `Milit[rische` -> `Militärische` |
| `}` | `ö` | `Bev}lkerung` -> `Bevölkerung` |
| `#` | `ü` | `K#nstliche` -> `Künstliche` |
| `|` | `ß` | `Au|enposten` -> `Außenposten` |

### French

| Source | Unicode | Example |
| --- | --- | --- |
| `#` | `é` | `D#mocratie` -> `Démocratie` |
| `>` | `è` | `Plan>te` -> `Planète` |
| `$` | `â` | `b$timent` -> `bâtiment` |
| `{` | `ô` | `Contr{le` -> `Contrôle` |
| `<` | `ç` | `commer<ants` -> `commerçants` |
| `[` | `ï` | `Andro[des` -> `Androïdes` |

### Spanish

| Source | Unicode | Example |
| --- | --- | --- |
| `{` | `í` | `Biolog{a` -> `Biología` |
| `}` | `ó` | `Poblaci}n` -> `Población` |
| `]` | `é` | `M]todos` -> `Métodos` |
| byte `0x60` | `ú` | `C` + `0x60` + `pula` -> `Cúpula` |
| `<` | `á` | `T<cticas` -> `Tácticas` |
| `|` | `ñ` | `Ense|anza` -> `Enseñanza` |

### Italian

| Source | Unicode | Example |
| --- | --- | --- |
| `&` | `à` | `Abilit&` -> `Abilità` |

These mappings are locale-specific. The same source byte can represent a different glyph in a different language, so there must never be a single global MOO2-to-Unicode table.

## Runtime language files now generated

The Race Designer normalization can now emit:

```text
data/languages/en.json
data/languages/de.json
data/languages/fr.json
data/languages/es.json
data/languages/it.json
```

Each file contains the same 64 stable keys. Non-English files declare English as fallback. Example key:

```text
race_traits.group.population_growth.name
```

Example values:

- EN: `Population`
- DE: `Bevölkerung`
- FR: `Population`
- ES: `Población`
- IT: `Popolazione`

The source block/hash is kept in each language file, while the ruleset itself continues to contain only language keys.

## Scope and caution

The glyph mappings above are currently validated for the Race Designer/technology-name evidence set. They should be expanded only when additional source strings prove more glyphs. Long-form message/help text may use further control bytes or language-specific glyph codes, so the raw private text catalog remains authoritative until those mappings are verified.
