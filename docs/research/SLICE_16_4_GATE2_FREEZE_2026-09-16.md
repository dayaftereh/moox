# Slice 16.4 Gate 2 - Preset race catalog / support / portrait contract freeze

Date: 2026-09-16
Status: **Gate 2 complete / Gate 3 next**

## Purpose

Freeze the smallest honest Slice-16.4 contract that can turn preset race choice into an image-led carousel without pretending all 13 normalized MOO2 presets already have complete MOOX gameplay parity.

This freeze is based on:

- `docs/research/SLICE_16_4_GATE1_PRESET_RACE_AUDIT_2026-09-16.md`;
- the 13 normalized races in `data/rulesets/moo2-1.31/races.json`;
- normalized trait definitions in `data/rulesets/moo2-1.31/race_traits.json`;
- the current Slice-09-derived two-player New Game validation;
- the shared Slice-16.1 `VisualSelector` interaction grammar;
- the Gate-1 Alkari/Meklar/Silicoid 4:5, WebP and 390 px prototype evidence.

Gate 2 does not implement race mechanics. It defines exactly what Gate 3 may expose and change.

## Decision summary

Slice 16.4 uses **two distinct support layers**:

1. **Preset catalog visibility** - all 13 normalized preset races are server-owned catalog entries and may be browsed in the carousel.
2. **Player runtime availability** - only races explicitly marked supported by the server may be submitted as the local player's race.

The two layers must never be conflated. A race may be visible because its identity is canonical while still being marked `planned` for player selection.

### Gate-3 player-selectable set

Freeze the initial player-selectable set to:

```text
human
klackon
```

Rationale:

- `human` is the existing local-player baseline and must remain a non-regression path. Its currently unimplemented Charismatic diplomacy breadth is **not** advertised as an active compact card effect.
- `klackon` is the only new preset classified `runtime-ready` by the Gate-1 audit: Unification, +1 food, +1 industry and Uncreative all have current authoritative consumers.
- adding Klackon materially widens race choice while keeping Slice 16.4 focused on catalog/portrait/HMI integration rather than importing unrelated gameplay-system work.

### Fixed opponent for Slice 16.4

Freeze the built-in-AI opponent to:

```text
darlok
```

This preserves the current deterministic two-player baseline. Darlok's continued role as the fixed AI opponent is **grandfathered baseline behavior**, not a claim that Darlok has complete player-facing race parity. Darlok remains `planned` for local-player selection until its defining Espionage/Stealth breadth is implemented or explicitly re-scoped later.

Slice 16.6 owns general opponent composition and race assignment breadth.

### Catalog-visible but player-planned races

The following remain browsable catalog entries but are not player-selectable in Slice 16.4 Gate 3:

```text
alkari
bulrathi
darlok
elerian
gnolam
meklar
mrrshan
psilon
sakkra
silicoid
trilarian
```

This list is not an ordering or quality judgment. It only records current implementation readiness from Gate 1.

## Bounded-fix scope freeze

Slice 16.4 Gate 3 may make **only** the runtime changes needed to support the frozen Human/Klackon player selector and server-owned catalog.

Allowed runtime changes:

- add a server-owned preset-race catalog endpoint;
- allow the non-Darlok player in the current two-player New Game tuple to use `human` or `klackon`;
- preserve one fixed Darlok opponent;
- preserve unique race IDs in the two-player tuple;
- add deterministic Human/Klackon New Game fixtures;
- send the selected supported race ID in the existing New Game payload;
- make the player empire-name UI race-neutral instead of hard-coded to Human.

Explicitly deferred from Slice 16.4:

- Artifacts / Rich / Large Home World rules;
- ship attack/defense race bonuses in Tactical Combat;
- ground-combat race bonuses in Invasion;
- Espionage and race spying bonuses;
- Stealthy Ships detection rules;
- Omniscient map/fleet knowledge;
- Telepathic capture / captured-ship breadth;
- Lucky / Random Events;
- Charismatic / Repulsive diplomacy parity;
- Cybernetic combat-repair breadth;
- Warlord troop/armor/crew/leader breadth;
- Trans-Dimensional Tactical speed parity;
- Fantastic Traders treaty breadth;
- general opponent race selection;
- custom race designer.

The existing effects that already work remain authoritative; this freeze merely prevents 16.4 from expanding into unrelated parity projects.

## Server-owned preset race catalog

Gate 3 adds:

```text
GET /api/v1/new-game/races
```

The response contract is frozen conceptually as:

```text
PresetRaceCatalog
  schema_version
  default_player_race_id
  fixed_opponent_race_id
  profiles[]

PresetRaceProfile
  id
  order
  name_key
  player_availability     # supported | planned
  trait_ids[]             # full normalized preset trait list
  card_fact_trait_ids[]   # compact server-curated identity/fact subset
```

Binding rules:

- `profiles` are emitted in normalized `races.json` order.
- IDs and full `trait_ids` come from the validated normalized ruleset, never duplicated in frontend constants.
- `default_player_race_id` is `human`.
- `fixed_opponent_race_id` is `darlok` for Slice 16.4.
- `player_availability=supported` only for `human` and `klackon`.
- every other profile is `planned`.
- the frontend may localize `name_key` and trait labels, but it must not invent availability or numeric gameplay facts.
- unknown/duplicate race IDs remain server errors.

A later slice may broaden `supported` without breaking the catalog shape.

## Compact race-card facts

The primary carousel is not a full Race Designer sheet.

Freeze these semantics:

- race name;
- availability badge (`supported now` / `planned`);
- 1-4 compact trait facts selected by the server through `card_fact_trait_ids`;
- portrait as the dominant visual;
- details modal may expose the complete normalized `trait_ids` later, but the primary card remains compact.

### Frozen Gate-3 fact selection

These fact IDs are **preset identity facts**. For `planned` races their presence does not promise the effect is currently implemented; the planned availability badge must remain visible.

| Race | `card_fact_trait_ids` |
| --- | --- |
| Alkari | `ship_defense_plus_50`, `artifacts_world`, `government_dictatorship` |
| Bulrathi | `high_g_world`, `ship_attack_plus_20`, `ground_combat_plus_10` |
| Darlok | `spying_plus_20`, `stealthy_ships`, `government_dictatorship` |
| Elerian | `omniscient`, `telepathic`, `government_feudal` |
| Gnolam | `fantastic_traders`, `money_plus_1_0`, `lucky` |
| Human | `government_democracy` |
| Klackon | `government_unification`, `farming_plus_1`, `industry_plus_1`, `uncreative` |
| Meklar | `cybernetic`, `industry_plus_2`, `government_dictatorship` |
| Mrrshan | `ship_attack_plus_50`, `warlord`, `rich_home_world` |
| Psilon | `science_plus_2`, `creative`, `low_g_world` |
| Sakkra | `population_growth_plus_100`, `subterranean`, `farming_plus_1` |
| Silicoid | `lithovore`, `tolerant`, `population_growth_minus_50` |
| Trilarian | `aquatic`, `trans_dimensional`, `government_dictatorship` |

Human deliberately shows only Democracy in the compact supported card so Slice 16.4 does not present Charismatic as a currently active diplomacy bonus.

## New Game validation freeze

The current validation requiring exactly one Human and one Darlok is replaced in Gate 3 with the narrow frozen support contract:

- still exactly two players;
- seat IDs remain non-zero, unique and strictly ascending;
- race IDs remain unique;
- exactly one player uses fixed opponent race `darlok`;
- the other player's race must be in the server-supported player set `{human, klackon}`;
- all race IDs must resolve through loaded `RaceModifiers`;
- no other preset is accepted merely because it exists in `races.json`.

This keeps deterministic topology unchanged while allowing the first real race choice.

## HMI behavior freeze

### Race carousel

Reuse the Slice-16.1 `VisualSelector` interaction family.

Binding behavior:

- setting ID: `player-race`;
- default selection: `human`;
- all 13 catalog races are browseable in canonical order;
- `human` and `klackon` display `supported now`;
- all other entries display `planned`;
- a planned race may be inspected but must not enable Create Game;
- selecting a supported race updates seat 1 `race_id` in the New Game request;
- seat 2 remains the server-frozen Darlok baseline;
- the existing `Human empire` field becomes a race-neutral `Player empire` field;
- the opponent empire-name field remains Darlok-specific until Slice 16.6.

### Accessibility / input

Inherit the shared selector contract:

- previous/next buttons at least 44 px; target remains 52 px in the visual card;
- keyboard Left/Right and Home/End navigation;
- touch/swipe behavior from the shared selector;
- visible race name and availability independent of the image;
- ARIA/text remains sufficient with images unavailable;
- details/info affordance remains keyboard reachable;
- planned status is conveyed by text, not color alone.

### Mobile

At 390 px viewport:

- portrait display target: `min(82vw, 320px)`;
- portrait remains 4:5 in the primary selector;
- name/facts live outside the artwork;
- no horizontal page overflow;
- controls remain >=44 px.

## Portrait art / runtime format freeze

The Gate-1 representative direction is **provisionally accepted for implementation**. The user explicitly accepted the current prototypes as good enough to proceed while reserving the right to change the images later.

This means Gate 3 may integrate the framing/material language, but the technical Alkari/Meklar/Silicoid research SVGs are not automatically final production portraits.

### Canonical portrait contract

- authoring/master target: **1200 x 1500 px, 4:5**;
- runtime portrait format: **WebP**;
- PNG is optional for lossless review/source export only and is not duplicated into runtime by default;
- no baked localized text;
- strong head/primary-sensory/upper-body equivalent silhouette;
- shared dark cinematic presentation with race-specific material/environment cues;
- family-consistent key/rim lighting;
- source/generation provenance recorded for every runtime portrait.

### Safe area

Default normalized safe area:

```text
x = 0.14
y = 0.08
width = 0.72
height = 0.84
```

Default focal point target:

```text
x = 0.50
y = 0.36
```

Per-race manifest overrides are allowed only when the non-humanoid anatomy requires them.

Identity must survive a centered 1:1 presentation crop. The 1:1 crop is CSS/presentation behavior, not a second mandatory source asset.

## Runtime race asset naming freeze

Canonical runtime root:

```text
web/public/assets/races/
```

Portrait path:

```text
web/public/assets/races/<race-id>/portrait.webp
/assets/races/<race-id>/portrait.webp
```

Optional emblem path:

```text
web/public/assets/races/<race-id>/emblem.svg
/assets/races/<race-id>/emblem.svg
```

Race asset manifest:

```text
web/public/assets/races/manifest.json
/assets/races/manifest.json
```

Semantic IDs:

```text
race:<race-id>:portrait
race:<race-id>:emblem
```

The existing lowercase kebab-case semantic segment validator remains binding. Current race IDs are already compatible.

### Manifest portrait entry minimum

```text
id
race_id
kind = portrait
format = webp
path
alt_name_key
focal_point { x, y }
safe_area { x, y, width, height }
provenance
source_note / generator note
```

Emblems are optional in Slice 16.4 and must not block the carousel.

## Gate-3 required runtime portrait set

Unique runtime portraits are required in Gate 3 for:

```text
human
klackon
darlok
```

Why Darlok is included even though local-player availability is planned:

- Darlok is the fixed current opponent;
- the same canonical portrait identity must be reusable later by Slice 16.6 / diplomacy / intelligence surfaces;
- this avoids leaving the current opponent as an anonymous database label while the player gets visual identity.

The Gate-1 Alkari/Meklar/Silicoid research prototypes remain committed research evidence. They may be promoted or replaced later, but they do not force those races into the Gate-3 supported player set.

Other planned races may use a deliberate non-race-specific `planned` visual state until their unique portraits are curated. Do not reuse another race's portrait as a placeholder.

## Explicit Gate-2 non-goals

Gate 2 / Gate 3 do not:

- declare all 13 presets gameplay-complete;
- implement missing race parity systems listed in the Gate-1 audit;
- implement opponent composition;
- implement custom race design;
- generate/finalize all 13 painterly portraits before the first working carousel;
- change Galaxy, Difficulty, Galaxy Age or Tactical Combat contracts;
- change seed/RNG behavior except for the selected supported player race's already-authoritative mechanics.

## Acceptance frozen for Gate 3

Gate 3 must prove:

1. `GET /api/v1/new-game/races` is server-owned, deterministic and returns all 13 profiles in canonical order.
2. only Human and Klackon are player-supported; all other catalog entries remain planned.
3. current Human-vs-Darlok start remains deterministic and non-regressed.
4. Klackon-vs-Darlok is accepted and deterministic for equal seed/settings.
5. unsupported/planned player race IDs are rejected server-side even if a client crafts the payload.
6. Web UI uses the catalog, not duplicated race availability constants.
7. Human/Klackon/Darlok have canonical unique runtime portraits and manifest entries.
8. carousel is usable on desktop and at 390 px mobile width with mouse/touch/keyboard.
9. planned races are browseable but cannot enable Create Game.
10. build/tests/diff checks pass.

## Gate-2 result

This freeze gives Slice 16.4 a real, testable breadth increase without hiding known parity gaps:

- **13 canonical identities visible**;
- **2 local-player races supported** (`human`, `klackon`);
- **1 fixed visualized AI opponent** (`darlok`);
- **one stable server catalog contract** ready for later expansion;
- **one canonical portrait asset family** reusable beyond New Game.
