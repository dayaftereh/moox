# Slice 16.4 Gate 1 - Preset race runtime and visual-DNA audit

Date: 2026-09-16
Status: **Gate 1 complete / Gate 2 next**

## Objective

Audit all 13 normalized preset races before binding a race carousel or producing final runtime portraits. This document separates four different questions that must not be conflated:

1. Is the race identity/trait set normalized and evidence-backed?
2. Does the current game layer actually apply those traits in authoritative gameplay?
3. Is the race currently accepted by New Game validation?
4. What factual anatomy/culture/lifestyle cues may inform **new original MOOX art** without copying original MOO2 portraits?

Gate 1 is research/prototype work only. Gate 2 will freeze the accepted race set, fact surface, art contract and runtime-fix scope.

## Primary local evidence

- `data/rulesets/moo2-1.31/races.json` - normalized 13 preset-race IDs/order/trait selections and portrait/icon identity keys.
- `data/rulesets/moo2-1.31/race_traits.json` - normalized trait IDs, scopes, values/abilities, pick costs and verification notes.
- `reference/original/text/help/block_0000.ascii.txt` HELP records 630-642 - original textual preset-race descriptions.
- `internal/game/economy_rules.go` - semantic projection of normalized race traits into active MOOX economy/research modifiers.
- `internal/game/research_advanced.go` - research preference projection for trait semantics.
- `internal/game/new_game.go` - current narrow Human + Darlok New Game validation.
- direct consumers in `internal/game/*` for economy, population, food logistics, command points, research, fleet movement, buildability, etc.

No original MOO2 portrait bitmap is a runtime-art source. Original graphic assets may remain provenance/reference data elsewhere in the repository, but Slice 16.4 runtime portraits must be independently created MOOX artwork.

## Canonical race inventory

The normalized preset list is exactly 13 races, in this order:

| Order | ID | Normalized trait selections |
| ---: | --- | --- |
| 0 | `alkari` | ship defense +50, Artifacts World, Dictatorship |
| 1 | `bulrathi` | High-G World, ship attack +20, ground combat +10, Dictatorship |
| 2 | `darlok` | spying +20, Stealthy Ships, Dictatorship |
| 3 | `elerian` | Omniscient, ship attack +20, ship defense +25, Telepathic, Feudal |
| 4 | `gnolam` | Fantastic Traders, +1 BC/pop, Lucky, Low-G World, ground combat -10, Dictatorship |
| 5 | `human` | Charismatic, Democracy |
| 6 | `klackon` | Unification, +1 food, +1 industry, Uncreative |
| 7 | `meklar` | Cybernetic, +2 industry, Dictatorship |
| 8 | `mrrshan` | ship attack +50, Warlord, Rich Home World, Dictatorship |
| 9 | `psilon` | +2 science, Creative, Low-G World, Large Home World, Dictatorship |
| 10 | `sakkra` | population growth +100%, Subterranean, +1 food, spying -10, Large Home World, Feudal |
| 11 | `silicoid` | Lithovore, Tolerant, population growth -50%, Repulsive, Dictatorship |
| 12 | `trilarian` | Aquatic, Trans-Dimensional, Dictatorship |

`races.json` and `race_traits.json` validate these selections and every race resolves into `RaceModifiers` / `RaceResearchModifiers` during rules loading.

## Common New Game blocker

Current `validateNewGameSettings` still imposes the old Slice-09 player restriction:

- exactly two players;
- only `human` and `darlok` race IDs;
- exactly one Human and one Darlok.

Therefore **none of the other 11 races can yet be selected in a live New Game**, even when their underlying race mechanics are otherwise usable. Slice 16.4/16.6 must remove that narrow whitelist through a server-owned supported-race contract rather than by frontend bypass.

This common validation restriction is separate from the per-race mechanics classification below.

## Runtime coverage definitions

### `runtime-ready`

All identity-defining traits that matter in the currently implemented game scope have meaningful authoritative consumers. A later subsystem may still add an original secondary effect that cannot matter yet (for example a future heroes subsystem), but the race is not presently missing its core playable identity.

### `bounded-fix`

The race has a meaningful implemented core, but one or more trait effects are missing from an **already existing** subsystem or from New Game homeworld initialization. The gap can plausibly be closed inside the current Slice-16 breadth without first building an entirely new gameplay family.

### `blocked`

A major positive/identity-defining trait fundamentally depends on a gameplay family that does not yet exist, so exposing the race as fully supported now would materially misrepresent it. Examples are Espionage, Random Events, or broad omniscience/telepathic intelligence behavior.

## Active trait coverage already present

The audit confirmed real authoritative consumers, not merely normalized fields, for substantial race breadth:

- food/industry/science/tax scalar modifiers;
- population growth multipliers;
- Aquatic climate treatment;
- Subterranean population-capacity behavior;
- Tolerant environmental/pollution/economy behavior;
- Lithovore food behavior;
- Cybernetic food + production sustenance and starvation/economy behavior;
- Creative / Uncreative research-choice behavior;
- Fantastic Traders surplus-food sale rate;
- Warlord command-point capacity;
- Trans-Dimensional strategic fleet/population-transport speed bonus;
- Low-G / High-G semantic race state used by existing economic/research planning surfaces;
- Dictatorship / Feudal / Democracy / Unification government rules used across economy/buildability/research/command-point paths.

The loader also records ship attack/defense, ground combat, spying, Telepathic and Stealthy Ships semantics for research valuation, but this **does not mean their intended combat/espionage effects are implemented**.

## Confirmed missing/partial trait families

### No direct game-layer hook currently found

- `artifacts_world`
- `rich_home_world`
- `large_home_world`
- `omniscient`
- `lucky`
- `charismatic`
- `repulsive`

Current homeworld generation applies one generic Small/Normal homeworld profile and does not branch on race for Artifacts/Rich/Large Home World.

### Recorded but not applied to the intended primary subsystem

- ship attack bonuses: represented in `RaceResearchModifiers`, not applied as race combat accuracy in current Tactical authority;
- ship defense bonuses: same limitation;
- ground combat bonuses: represented for research valuation, not applied to invasion combat;
- spying bonuses: no Espionage gameplay system yet;
- Telepathic: research preference effect exists, but not automatic undefended-world capture / immediate captured-ship behavior;
- Stealthy Ships: research preference effect exists, but no race stealth/detection implementation;
- Trans-Dimensional: +2 strategic speed exists; the original tactical combat-speed component is not yet implemented;
- Cybernetic: economic sustenance is implemented; original post-combat/per-round ship repair breadth is not complete;
- Warlord: command-point effect exists; full original troop/armor/crew/leader breadth is not complete;
- Fantastic Traders: current MOOX surplus-food sale modifier exists; original trade-treaty breadth is not complete.

## Per-race runtime-readiness matrix

| Race | Class | Implemented core | Material gaps before honest full support |
| --- | --- | --- | --- |
| Alkari | **bounded-fix** | government; normalized defense semantics | Tactical ship-defense bonus; Artifacts Home World initialization |
| Bulrathi | **bounded-fix** | High-G identity/government | Tactical attack bonus; invasion ground-combat bonus |
| Darlok | **blocked** | government; normalized spy/stealth semantics | Espionage system is absent; Stealthy Ships detection effect absent |
| Elerian | **blocked** | Feudal government; normalized combat/telepathic semantics | Omniscience, Telepathic capture/ship behavior, actual Tactical attack/defense hooks |
| Gnolam | **blocked** | +1 BC/pop; Low-G; partial Fantastic Traders | Lucky requires Random Events; original trade-treaty breadth absent; ground-combat penalty absent |
| Human | **bounded-fix** | Democracy | Charismatic diplomacy modifier missing; future leader-hiring discount has no leader system yet |
| Klackon | **runtime-ready** | Unification, +1 food, +1 industry, Uncreative all have active authoritative consumers | no known current-scope blocker from preset traits |
| Meklar | **bounded-fix** | Cybernetic economy/sustenance; +2 industry; Dictatorship | Cybernetic combat repair breadth not complete |
| Mrrshan | **bounded-fix** | Warlord command points; Dictatorship | actual ship attack +50; Rich Home World; remaining Warlord troop/crew/leader breadth |
| Psilon | **bounded-fix** | +2 science, Creative, Low-G, Dictatorship | Large Home World initialization |
| Sakkra | **bounded-fix** | growth x2, Subterranean, +1 food, Feudal | Large Home World; spying -10 waits for Espionage but is non-operative until that subsystem exists |
| Silicoid | **bounded-fix** | Lithovore, Tolerant, half growth, Dictatorship | Repulsive diplomacy effect; future leader cost consequence unavailable until leader system |
| Trilarian | **bounded-fix** | Aquatic, +2 strategic Trans-Dimensional speed, Dictatorship | original Tactical Trans-Dimensional speed component |

### Gate-1 interpretation

The matrix does **not** mean Slice 16.4 must implement every original deferred secondary effect. Gate 2 must decide the supported-race policy explicitly:

- accept only fully/current-scope ready races;
- or close bounded fixes needed for a wider supported set;
- but do not advertise Darlok/Gnolam/Elerian as fully faithful while their defining missing systems are absent.

The existing Human+Darlok whitelist is not a sensible permanent supported-set policy: Darlok is currently one of the most mechanically incomplete presets because its primary positive identity is Espionage/stealth.

## Original HELP-derived anatomy / culture facts

The following are **textual factual cues** from HELP records 630-642, paraphrased for art planning. They are not instructions to reproduce the original portraits.

- **Alkari:** avian race descended from large flying reptiles; pilots raised around three-dimensional motion; homeworld retains ancient Orion artifacts.
- **Bulrathi:** large bear-like people; physically strong/hardy; heavy-gravity homeworld; aggressive martial reputation.
- **Darlok:** shape-shifters able to assume almost any humanoid form; spy/stealth identity.
- **Elerian:** humanoid with strong angular, almost elf-like facial features; female mystic order; psychic/telepathic culture.
- **Gnolam:** dwarf-like/small-stature people; society strongly focused on commerce/monetary gain; Low-G origin.
- **Human:** charismatic negotiators/deal makers; unique democratic government among the original presets.
- **Klackon:** insectoid people; hive-mind/unified society; highly industrious workers.
- **Meklar:** cybernetic race whose biological bodies have atrophied into small fragile frames inside/alongside heavy cybernetic exoskeleton dependence.
- **Mrrshan:** cat-like race with keen senses; strict warrior code and warlord culture; mineral-rich homeworld.
- **Psilon:** delicate low-gravity people of brilliant researchers; science/creative identity.
- **Sakkra:** reptilian people with strong regeneration; subterranean life; agriculture emphasis.
- **Silicoid:** crystalline beings; lithovores consuming minerals; environmentally tolerant; socially alien/repulsive to other species.
- **Trilarian:** aquatic life forms from an ocean world; trans-dimensional ability described as mentally folding space.

These facts are sufficient to establish strongly different silhouettes without consulting/copied portrait pixels.

## Visual-DNA records

In every record below, **anatomy / culture evidence** is sourced from the normalized traits + HELP text. Palette, materials and composition are **new MOOX art-direction choices**, not claims about original colors or clothing.

### Alkari

- **Evidence anatomy:** avian, descended from flying reptiles.
- **Evidence culture/ability:** three-dimensional pilot training; artifact-bearing homeworld.
- **MOOX visual direction:** streamlined avian-reptilian cranial silhouette, swept aerodynamic neck/shoulder armor, subtle flight-control or gyroscopic cockpit cues, one abstract ancient-artifact geometry in background.
- **Palette inference:** cool deep-space blue with pale bone/keratin and restrained amber artifact light.
- **Leadership tone:** alert, spatially aware, controlled rather than aggressive.
- **Avoid:** copying the original MOO2 bird portrait, Earth-eagle heraldry, generic angel wings, Star Trek-like uniform mimicry.

### Bulrathi

- **Evidence anatomy:** large bear-like, physically powerful/hardy, Heavy-G origin.
- **MOOX visual direction:** broad low-center-of-mass torso, dense musculature/fur analogue, compact reinforced Heavy-G pressure harness, massive structural shoulder silhouette.
- **Palette inference:** iron, basalt, muted rust, small high-energy accents.
- **Tone:** durable, direct, intimidating without becoming a fantasy bear warrior.
- **Avoid:** literal Earth brown-bear costume, Viking/barbarian clichés.

### Darlok

- **Evidence anatomy:** shape-shifters; can assume almost any humanoid form.
- **MOOX visual direction:** deliberately unstable identity - asymmetric translucent facial planes, partially unresolved silhouette, layered adaptive membrane/stealth material rather than one canonical humanoid species face.
- **Palette inference:** near-black violet, spectral teal highlights, controlled iridescence.
- **Tone:** unreadable, observational, covert.
- **Avoid:** generic hooded spy, copying known Darlok face, a single fixed species anatomy presented as canonical.

### Elerian

- **Evidence anatomy:** humanoid, strongly angular facial features, almost elf-like; female mystic order.
- **MOOX visual direction:** angular humanoid leader/mystic with precise geometric ceremonial layers, psychic-sensor architecture and restrained non-magical neural/field motifs.
- **Palette inference:** midnight indigo, pale metallic ceramic, controlled violet-white field light.
- **Tone:** composed, foresighted, aristocratic/mystic.
- **Avoid:** fantasy elf ears as the whole identity, medieval robes, direct original-portrait likeness.

### Gnolam

- **Evidence anatomy:** dwarf-like/small stature; Low-G origin.
- **Evidence culture:** commerce/monetary focus.
- **MOOX visual direction:** compact low-G physiology with elongated support frame or seated command architecture; trade-network holography and precision mercantile instruments rather than coins/cash.
- **Palette inference:** dark graphite, warm electrum, cyan market/network data.
- **Tone:** confident, calculating, socially polished.
- **Avoid:** fantasy dwarf beard/armor, money bags/coins, caricatured banker stereotypes.

### Human

- **Evidence anatomy:** human.
- **Evidence culture:** charismatic diplomacy; democratic government.
- **MOOX visual direction:** civilian-diplomatic command portrait rather than default military officer; subtle coalition/civic interface background.
- **Palette inference:** neutral navy/graphite with warm skin-light and restrained civic cyan.
- **Tone:** persuasive, accessible, politically competent.
- **Avoid:** one ethnicity as the species identity, contemporary national flags, real-world military insignia.

### Klackon

- **Evidence anatomy:** insectoid.
- **Evidence culture:** hive mind / Unification; industrious workers.
- **MOOX visual direction:** segmented exoskeleton with multiple sensory surfaces, worker/administrator caste rather than queen cliché, distributed network/colony geometry echoed in garment/tool forms.
- **Palette inference:** dark chitin, muted amber/green biochemical highlights, hard industrial blue-white work light.
- **Tone:** focused, collective, tireless.
- **Avoid:** direct ant/bee costume, horror-only giant insect, copying original face morphology.

### Meklar

- **Evidence anatomy:** small fragile/atrophied biological frame; heavily dependent on cybernetic exoskeletons.
- **Evidence culture/ability:** cybernetic sustenance/repair identity, extreme industry.
- **MOOX visual direction:** visible contrast between a compact protected biological core and a much larger modular machine support frame; industrial manipulators and service conduits should read as life-support, not generic robot body.
- **Palette inference:** blackened alloy, oxidized copper, surgical cyan, small organic warm core.
- **Tone:** relentless, analytical, post-biological dependency rather than emotionless robot.
- **Avoid:** Terminator/Cylon/Warhammer-like robot silhouette; making the race wholly mechanical.

### Mrrshan

- **Evidence anatomy:** cat-like with keen senses.
- **Evidence culture:** strict warrior code / Warlord; mineral-rich homeworld.
- **MOOX visual direction:** feline sensory anatomy with prominent directional ears/whisker-like sensor organs, disciplined high-tech duelist/commander clothing and mineral-derived armor materials.
- **Palette inference:** dark mineral blue/black, copper-gold ore inclusions, warm amber eyes/sensors.
- **Tone:** disciplined, intensely attentive, proud.
- **Avoid:** house-cat cuteness, anime cat-person shorthand, medieval samurai cosplay.

### Psilon

- **Evidence anatomy:** delicate Low-G people.
- **Evidence culture:** brilliant researchers; Creative.
- **MOOX visual direction:** lightweight elongated low-G physiology, minimal load-bearing clothing, dense scientific instrument field and multi-layer optical/data interfaces.
- **Palette inference:** pale ceramic, cool turquoise, ultraviolet data glow.
- **Tone:** curious, analytical, calm.
- **Avoid:** mandatory oversized-brain alien cliché unsupported by text, generic grey alien.

### Sakkra

- **Evidence anatomy:** reptilian; strong regeneration; subterranean.
- **Evidence culture:** agriculture emphasis.
- **MOOX visual direction:** reptilian skin/scale morphology with subtle regenerative tissue patterning, robust subterranean light adaptation, bio-agricultural engineering cues rather than primitive jungle aesthetics.
- **Palette inference:** deep olive/obsidian, bioluminescent lime/amber agricultural accents.
- **Tone:** vigorous, adaptive, fecund.
- **Avoid:** dinosaur-man cliché, tribal/jungle costume shorthand.

### Silicoid

- **Evidence anatomy:** crystalline beings; lithovores.
- **Evidence culture/biology:** mineral consumption, environmental tolerance, slow growth, difficulty relating to organics.
- **MOOX visual direction:** genuinely non-organic crystalline body made from asymmetric interlocking mineral masses; internal light transport/refraction instead of eyes/mouth unless needed as abstract sensory structures; no clothing required, but optional grown lattice command interface.
- **Palette inference:** smoky quartz/obsidian mass, internal magenta-cyan refraction, sparse white edge light.
- **Tone:** ancient, deliberate, physically alien.
- **Avoid:** humanoid rock golem, fantasy crystal elemental, giving it a human face merely for readability.

### Trilarian

- **Evidence anatomy:** aquatic life forms from an ocean world.
- **Evidence ability:** mentally folds space / Trans-Dimensional.
- **MOOX visual direction:** aquatic physiology with membrane/fins/gill-like structures optimized for fluid motion, supported by a moisture-field command environment; subtle local lensing/phase distortion around neural/crest area.
- **Palette inference:** deep ocean cyan, blue-black, pearl/bioluminescent highlights.
- **Tone:** fluid, spatially intuitive, distant but not mystical-fantasy.
- **Avoid:** simple fish-person, diving suit cliché, copying original aquatic portrait.

## Representative Gate-1 portrait prototype set

The first three prototype subjects are intentionally selected for **visual-pipeline stress**, not because they are a gameplay ranking:

1. **Alkari** - organic avian/reptilian morphology + armor/cockpit materials;
2. **Meklar** - biological/cybernetic contrast and hard-surface machinery;
3. **Silicoid** - non-humanoid transparent/crystalline material rendering.

A single art style must make all three feel like the same game's leadership portrait system while preserving radically different materials and silhouettes.

## Prototype framing contract to test

Research prototype master:

```text
1200 x 1500 px (4:5)
```

Safe framing:

- head/primary sensory mass center around 50% width, 34-40% height;
- critical silhouette and primary eyes/sensors remain within central 72% width;
- upper torso/equivalent body mass remains visible through ~82% height;
- top/bottom 8% may be cropped without losing identity;
- no baked text or insignia required for recognition.

Carousel research crops:

- desktop hero: 4:5 full portrait;
- compact/card: centered 1:1 crop;
- mobile 390px: portrait displayed at ~min(82vw, 320px), with title/facts outside artwork;
- crop must never remove the primary sensory focal point or make Meklar's biological core / Silicoid's non-humanoid nature unreadable.

## Format experiment plan

For the three prototypes, Gate 1 compares:

- lossless PNG reference;
- WebP at a visually conservative quality setting;
- file size ratio;
- 4:5 full view and 1:1/mobile crop;
- edge behavior on feathers/chitin/machinery/crystal-like high-frequency detail.

SVG remains appropriate for future race emblems/icons, not for the final painterly portrait masters unless Gate 1 unexpectedly proves a deliberately vector-native portrait style superior.

## Gate-1 provisional findings

- Canonical race IDs/order are already normalized and stable.
- Current New Game validation is the immediate common server blocker for 11 races beyond Human+Darlok.
- Race mechanics coverage is much broader than the current New Game whitelist, but not complete enough to honestly mark all 13 equally ready.
- `klackon` is the cleanest current-scope mechanically complete preset.
- `darlok`, `elerian`, and `gnolam` have major defining abilities tied to absent gameplay families and should be treated as blocked until Gate 2 explicitly resolves policy.
- Most other presets are bounded-fix candidates, with missing homeworld/tactical/diplomacy hooks rather than completely absent core loops.
- Original HELP text provides sufficient factual anatomy/lifestyle distinctions to create strongly differentiated original MOOX portraits without copying original MOO2 portrait composition or pixels.

Gate-1 portrait/format/mobile prototype evidence is recorded in `docs/research/prototypes/SLICE_16_4_RACE_PORTRAITS_2026-09-16/FORMAT_COMPARISON.md`. Three deterministic Alkari/Meklar/Silicoid composition prototypes and compact WebP review assets are retained under that research directory.

Gate 1 is complete. Gate 2 is next and must explicitly freeze the supported-race policy, bounded-fix scope, final portrait style/format contract and server-owned race catalog before full implementation.
