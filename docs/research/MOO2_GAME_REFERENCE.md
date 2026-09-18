# Master of Orion II - Game/System Reference

Research baseline: 2026-08-26

This document is a working specification for reproducing the **behavior** of Master of Orion II: Battle at Antares. It deliberately summarizes factual game rules instead of copying original manual text or assets.

## 1. Identity and design center

- Original title: Master of Orion II: Battle at Antares.
- Developer: SimTex.
- Original publisher: MicroProse.
- Original PC release: 1996 (DOS/Windows; later Macintosh).
- Genre: turn-based 4X / grand-strategy / economic strategy with tactical ship combat.
- Core fantasy: create or select a species, expand from a homeworld, develop colonies and technology, negotiate or fight with other empires, design ships and win control of the galaxy.

The important design lesson for MOOX is that MOO2 is not one subsystem; it is the coupling between population management, research choices, colony specialization, diplomacy and highly configurable ships.

## 2. Victory conditions

There are three major victory paths to reproduce:

1. **Elimination** - remove all rival empires.
2. **Galactic Council** - win the galactic leadership vote with the required supermajority; voting power is driven by population.
3. **Antaran victory** - obtain the means to reach the Antaran dimension and defeat the Antaran homeworld.

Conquering Orion and defeating the Guardian provides extraordinary rewards but is not itself an automatic victory.

## 3. New-game configuration

The original new-game flow exposes the following dimensions:

- Difficulty from Tutorial through Impossible.
- Galaxy size.
- Galaxy age.
- Number of players/empires.
- Starting technology level:
  - Pre-Warp,
  - Average,
  - Advanced.
- Tactical or strategic combat.
- Random events on/off.
- Antaran attacks on/off.

Behavioral notes:

- Pre-Warp starts with one colony and essentially no interstellar capability.
- Average/Advanced starts seed progressively more technology and ships.
- Choosing strategic combat means the player does not manually design/control tactical warships in the same way as tactical mode.
- Random events can materially alter colonies, ships, diplomacy and research.

## 4. Galaxy, systems and planets

### 4.1 Galaxy model

The strategic map contains stars plus special objects such as black holes. A normal star can have multiple orbital bodies; the original game supports systems with up to five colonizable planets.

Strategic movement is range based, not constrained to fixed hyperlanes. This is a major gameplay property: there are no natural lane choke points like in many modern 4X games.

### 4.2 Planet properties

A colonizable planet is described by several independent dimensions:

- climate/environment,
- size,
- gravity,
- mineral richness,
- maximum population,
- specials/resources,
- ownership/colony state.

Notable specials include:

- artifacts,
- natives,
- gold/gem deposits,
- splinter colonies,
- guarded high-value worlds.

Gas giants and asteroid belts are not initially colonizable but can become relevant through late technology.

### 4.3 Climate and fertility

Planet environment affects farming and population potential. The environment ladder includes hostile types such as Toxic/Radiated/Barren and increasingly habitable types through Desert/Tundra/Ocean/Swamp/Arid/Terran/Gaia-style endpoints depending on transformation.

For fidelity, planet type must not be treated as a cosmetic tag; it drives food output, population capacity and terraforming paths.

### 4.4 Mineral richness

Mineral richness drives industrial output per worker. The game distinguishes a scale from very poor worlds through rich and ultra-rich worlds.

### 4.5 Gravity

Races have gravity affinity. Working on a planet outside that affinity causes economy penalties unless technology/buildings mitigate them. High-G and Low-G race traits therefore affect both colony selection and ground combat.

## 5. Population and colony economy

### 5.1 Population roles

Population units can be assigned to three principal jobs:

- Farmers -> food.
- Workers -> production.
- Scientists -> research.

This simple three-role assignment is central to the MOO2 feel because every colony is continuously rebalanced against empire-wide needs.

### 5.2 Food and freighters

Food can move between colonies using the empire's freighter pool. Agricultural worlds can feed industrial/research worlds. A colony that cannot obtain enough food suffers population consequences.

Freighters are therefore an empire-wide logistics resource, not just visual transport units.

### 5.3 Production

Workers and industrial buildings create production points. Production is spent on buildings, ships and other projects.

The effective production model is influenced by:

- planet mineral richness,
- race production modifiers,
- government,
- morale,
- gravity mismatch,
- blockades,
- colony leaders,
- industrial technologies/buildings,
- pollution.

Flat production from buildings and worker-generated production should be tracked separately because pollution calculations do not treat every source identically.

### 5.4 Pollution

Industrial activity generates pollution, reducing effective output unless controlled. Pollution processors, atmosphere renewers and late-game technologies progressively improve or remove this penalty.

This is a key MOO2 balancing loop: high industry requires parallel investment in pollution control.

### 5.5 Research

Scientists plus research buildings produce research points (RP). Only one active research project is selected at a time, and accumulated research completes technologies.

### 5.6 Money / BC

The empire has a treasury measured in BC. Income comes from population and economic bonuses/specials; expenses include maintenance and other empire costs. Cash can accelerate colony production via buyout.

For original MOO2 1.31 buyout calculations, let `X` be total PP cost and `Y` already-produced PP. The frozen evidence gives `4*X - 10*Y` through 10% completion, `3.5*X - 5*Y` from 10% through 50%, and `2*X - 2*Y` from 50% through completion. Equivalent marginal costs are 10 BC/PP for the first 10%, 5 BC/PP through 50%, then 2 BC/PP. A bought project completes when normal Turn resolution runs rather than immediately on the Buy click. Sources: StrategyWiki `Master_of_Orion_II:_Battle_at_Antares/Calculations` and `Speeding_up_production`, corroborated by the GameFAQs MOO2 strategy guide. Runtime implementation/evidence: `docs/research/SLICE_17_0_GATE3_BUYOUT_DEV_BC_AMENDMENT_2026-09-18.md`.

### 5.7 Morale and government

Morale affects colony productivity for most governments. Some government forms change or bypass morale behavior and also alter research/economic performance.

### 5.8 Population growth

Population grows per race on each colony. Growth depends on current population vs. capacity and is modified by racial growth traits, medicine technology, housing and cloning. Growth is strongest away from the population-cap extremes rather than being purely linear.

A high-fidelity implementation should maintain population at a finer internal precision than the visible worker icons.

## 6. Race design

### 6.1 Preset races

MOO2 contains 13 predefined playable races:

- Alkari
- Bulrathi
- Darlok
- Elerian
- Gnolam
- Human
- Klackon
- Meklar
- Mrrshan
- Psilon
- Sakkra
- Silicoid
- Trilarian

### 6.2 Custom race budget

Custom race design begins with 10 positive "picks". Negative traits can add points, with a maximum negative budget of 10 picks in the standard game. The result is a constrained build system rather than arbitrary sliders.

### 6.3 Trait families

Race options include modifiers for:

- population growth,
- farming,
- industry,
- research,
- money,
- ship attack,
- ship defense,
- ground combat,
- spying,
- homeworld size/richness,
- gravity affinity,
- government,
- special abilities.

Important named special traits include concepts such as:

- Creative / Uncreative,
- Lithovore,
- Cybernetic,
- Aquatic,
- Subterranean,
- Tolerant,
- Telepathic,
- Trans-Dimensional,
- Warlord,
- Charismatic / Repulsive,
- Lucky,
- Omniscient.

Trait interactions and incompatibilities are part of the game design and need explicit rule data, not UI-only validation.

### 6.4 Government

Government is one of the most consequential race choices. Base governments include forms such as Dictatorship, Democracy, Feudal and Unification, with advanced forms available through research. Government affects research, morale, assimilation and/or productivity.

## 7. Technology system

### 7.1 Eight research fields

MOO2 uses eight independent research fields:

1. Construction
2. Chemistry
3. Computers
4. Physics
5. Power
6. Sociology
7. Biology
8. Force Fields

Each field is a sequential ladder of research levels. A level normally offers one to three technologies.

### 7.2 Normal / Creative / Uncreative behavior

- Normal races generally select one technology from a completed level.
- Creative races obtain all technologies in the level.
- Uncreative races do not choose; the game assigns the result.

Other acquisition paths include diplomacy/trade, espionage, conquest/capture, artifacts and leaders.

### 7.3 Miniaturization

Researching beyond a component's introduction level makes many ship components smaller and cheaper. This means old technology does not simply become obsolete: a mature weapon can become extremely space-efficient.

Miniaturization is one of the most important systems to reproduce exactly because it heavily shapes ship-design meta.

### 7.4 Technology data model requirements

Every technology record should eventually support at least:

- stable id,
- research field,
- level/order,
- RP cost,
- choice group,
- prerequisites,
- unlocks,
- empire modifiers,
- colony/building effects,
- ship component effects,
- miniaturization category,
- AI valuation tags.

## 8. Colony construction

Colonies have a build queue/current project. Buildable categories include:

- colony buildings,
- orbital defenses/bases,
- military ships,
- colony/outpost/transports,
- housing,
- trade goods or equivalent conversion projects.

Key behaviors to preserve:

- production carries a concrete per-turn cost,
- remaining production can be bought with BC,
- buildings create flat and/or worker multipliers,
- building maintenance matters,
- obsolete/redundant buildings can be sold,
- colony specialization is strong because logistics allow food to move between worlds.

## 9. Fleet model and command points

The empire has a command-point capacity supplied by orbital infrastructure and technologies. Ships consume command points based on hull class; exceeding capacity creates a financial penalty.

This provides a strategic fleet-size constraint and makes larger hulls relatively command-efficient.

Support ships include fixed-purpose designs such as:

- Colony Ship
- Outpost Ship
- Troop Transport
- Freighter (logistical pool rather than a normal combat fleet unit)

Warships are player-designed when tactical combat is enabled.

## 10. Warship design

### 10.1 Hull classes

Military hull progression includes:

- Frigate
- Destroyer
- Cruiser
- Battleship
- Titan
- Doom Star

Larger hulls offer more internal space and durability but differ in cost, command usage and component scaling.

### 10.2 Design slots and refitting

The original game limits the number of concurrently editable/current ship classes while still allowing already-built older designs to remain in service. Ships can be refitted to newer designs/technology rather than automatically receiving every component upgrade.

### 10.3 Component families

A design combines categories such as:

- drive,
- battle computer,
- armor,
- shields,
- beam weapons,
- missiles/torpedoes,
- bombs,
- fighters,
- special systems.

Weapon systems support modifications unlocked through miniaturization/research (for example accuracy/range/damage or firing characteristics depending on weapon type).

## 11. Tactical space combat

Tactical combat is a turn-based 2D battlefield where individual ships move and attack.

Important combat state layers include:

1. Shields
2. Armor
3. Structure
4. Internal systems
5. Crew/marines

Damage that penetrates defenses can degrade specific ship systems. Drive destruction can be catastrophic. Ships can retreat, wait, board under the right conditions, launch ordnance and use special systems.

Combat needs deterministic random-number handling for testability while still reproducing MOO2-style chance behavior.

## 12. Planetary combat and conquest

After orbital defenses are removed, an attacker can bombard or invade.

Invasion uses troop transports and resolves ground combat automatically based on:

- troop numbers,
- racial ground-combat modifiers,
- ground-combat technology,
- leader bonuses,
- defensive conditions.

Conquered populations may remain as a different race. Assimilation/discontent matters, while Telepathic behavior is a special exception in the original rules.

## 13. Diplomacy and intelligence

Diplomacy supports a broad negotiation set:

- gifts/demands,
- technology exchange,
- trade treaties,
- non-aggression treaties,
- alliances,
- war/peace,
- colony/system concessions in applicable interactions.

Empire relations are not a single binary state; they reflect history, treaties, personality and actions.

Spies can be assigned to defense or offensive missions. Offensive intelligence includes espionage and sabotage. Race spying traits and computer-related technology matter strongly.

Repulsive and Charismatic are special diplomacy modifiers and should be modeled as rule changes, not simple numeric bonuses.

## 14. Leaders

The game contains hireable leaders in two broad roles:

- Colony leaders
- Ship leaders

Leaders have levels/skills and can modify economy, research, combat, navigation, finance, environment and other systems. Some can provide access to technologies.

A future MOOX data model should keep leaders fully data-driven so that traits can feed the same modifier pipeline as race, government, building and technology effects.

## 15. Random events and monsters

Random events include beneficial and harmful effects such as:

- salvage/technology finds,
- population booms,
- donations,
- plague,
- environmental/mineral shifts,
- pirates,
- special-feature gains/losses,
- space-monster attacks,
- hyperspace/time-space fluxes,
- computer virus,
- research breakthrough,
- diplomatic incidents,
- supernova/comet-type crises.

Space monsters guard some valuable systems and can also appear through events.

## 16. Orion and Antarans

The Orion system is protected by the Guardian. Defeating it yields unique strategic rewards and access to the exceptional Orion world.

The Antarans periodically attack when enabled. Late technology makes a counter-invasion possible, creating the third major victory path.

These should be implemented as normal game systems plus scripted content, not hardwired UI sequences.

## 17. AI fidelity

MOO2 difficulty is not solely smarter decision-making; higher difficulty also changes the AI's effective advantages. A fidelity mode should separate:

- decision policy/personality,
- difficulty bonuses,
- information available to AI,
- production/research/economic modifiers,
- ship-refit behavior.

That separation will let MOOX offer an "original rules" AI mode and later a stronger modern AI without changing the economy simulation.

## 18. Fidelity test strategy

The simulation should expose headless tests for at least:

- galaxy generation with a fixed RNG seed,
- colony food/production/research output,
- pollution,
- population growth,
- morale/government modifiers,
- research completion and Creative/Uncreative choices,
- ship-space calculations and miniaturization,
- command-point accounting,
- beam/missile combat resolution,
- invasion resolution,
- diplomacy relation changes,
- council vote calculation,
- save/load round-trip equality.

The long-term goal is to create fixtures from observations of a legally owned original installation so the same input state can be compared against MOOX output.

## 19. Open questions for exact parity

These areas need data extraction and/or controlled experiments against the original game before implementation can claim exact parity:

- exact galaxy-generation distributions by size/age/difficulty,
- complete race-pick costs and incompatibility matrix,
- full technology tree and every effect value,
- exact building costs/maintenance/effects,
- complete ship component statistics and modification rules,
- exact combat hit/damage/initiative formulas,
- complete leader roster and skill progression,
- diplomacy thresholds/personality tables,
- AI difficulty bonuses and hidden rules,
- random-event probability tables,
- exact save-game and original structured-data layouts.

Those become the next research/data-normalization phase once an original installation is available locally.
