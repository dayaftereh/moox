# Slice 15.3 Gate 1 - visual consistency audit

Date: **2026-09-07**

Status: **pass; no blocking strategic visual inconsistency found**.

## Purpose

This audit follows the Research/Galaxy and Body/Fleet/Construction/Diplomacy integration passes. Its purpose is to stop adding icon variants speculatively and instead verify that the current strategic visual language is coherent enough to support a Gate-2 freeze decision.

## Audited surfaces

The persistence-enabled `game-1` on canonical port 7171 was used for live browser inspection.

| Surface | 791px review | 320px review | Result |
| --- | --- | --- | --- |
| Galaxy | no danger state; no horizontal overflow; shared star/system/contact icon language | previously + current Gate-1 passes verified at 320px | PASS |
| Colony list | no danger state; no horizontal overflow; role icons present | exact 320px width; no danger state | PASS |
| Colony Detail | no danger state; no horizontal overflow; population/building visual language coherent | prior 320px role/building QA remains valid | PASS |
| Construction manager | no danger state; no horizontal overflow; typed project/building icons | 19px catalog icons, 48px hero, exact 320px width | PASS |
| Fleets | combat and civilian/special roles distinct; procedural ship visuals retained | prior exact 320px QA; compact 16px role chips | PASS |
| Research | exactly eight category cards; shared category identities | prior exact 320px QA; one-column classic overlay | PASS |
| Diplomacy | neutral stance + War action use semantic SVGs; no overflow | exact 320px width; no danger state | PASS |
| Espionage | intentionally unavailable; generic empty state only | exact 320px width; no danger state | PASS / deferred gameplay |
| Shipbuilder | procedural ship + common action icons; no overflow | exact 320px width; procedural ship present | PASS |

## Semantic consistency findings

### Global visual grammar

- Navigation, resources, population roles, actions, Research categories, strategic markers, body types, fleet roles, construction categories and Diplomacy states all use typed SVGs rather than Unicode/icon-font dependencies.
- The common icon grammar remains 24x24 viewBox, `currentColor`, rounded line caps/joins and approximately 1.55 stroke weight.
- Small action icons are intentionally selective. There is no requirement to decorate every text button.
- Semantic state color remains consistent across surfaces:
  - accent/player/action: blue;
  - success/friendly/colony: green;
  - warning/outpost: amber;
  - danger/hostile/war: red;
  - neutral/unknown: slate.

### Galaxy and System

- Galaxy stars and the System dialog title share `star-system`.
- Colony, outpost and fleet presence use the same identities on Galaxy markers and System drill-down UI.
- Planet, gas giant and asteroid belt are visibly distinct without requiring text.
- Climate tint is secondary information layered onto the body identity rather than a replacement for shape/type.

### Colony and Construction

- Farmer/Worker/Scientist headings and individual population markers use the same role identity.
- Built-building tiles and construction choices resolve through the same project/building icon mapping.
- Known early-game buildings have distinct identities; unknown future buildings fall back safely to generic `build`.
- Queue direction controls no longer rely on text-arrow glyphs.

### Fleet and Ship

- Fleet-purpose iconography is separate from the procedural ship silhouette. This avoids conflating a design's visual hull with the fleet's strategic role.
- Special vessels reuse specific semantic identities (`flag`, `outpost`, `transport`) instead of multiplying near-duplicate fleet icons.

### Research

- The global Research identity remains the microscope.
- The eight authoritative categories have stable distinct identities.
- Chemistry uses the test-tube primitive as intended; it does not replace the global Research microscope.

### Diplomacy and Espionage

- Diplomacy stance and action semantics now share the same friendly/neutral/hostile language used by strategic contacts.
- Espionage is intentionally not expanded into fake mechanics during Slice 15.3. Its empty state remains acceptable until Slice 20 provides real data/actions.

## Placeholder sweep

A source sweep for obvious strategic text-symbol/CSS-glyph placeholders found no blocking candidates after the current passes.

Remaining matches are intentional:

- orbit rings: actual geometric visualization;
- `ProceduralShipGlyph`: accepted procedural ship artwork;
- typed population/building icon containers;
- disabled Ship Designer scaffold: functional Slice-17 boundary, not an art placeholder.

## Gate-1 conclusion

The currently implemented **Deep Space Command** direction is internally coherent across the functional strategic surfaces and survives the 320px floor without horizontal overflow.

No additional icon invention is justified before a Gate-2 direction decision.

The remaining Gate-1 work is therefore contractual/presentational rather than screen-by-screen implementation:

1. compare representative styleboard candidates;
2. define responsive/resolution rules;
3. define provenance/license rules;
4. present the Gate-2 asset/pipeline contract and acceptance mockups.
