# Slice 15.2/15.3 accepted strategic HMI product direction

Date: **2026-09-04**

Status: **accepted downstream product direction; implementation belongs to Slice 15.2/15.3, not Slice 15.1 Gate 4**.

## Navigation / information architecture

The Slice-15.1 five-destination mobile shell remains the compact phone baseline. The product-level strategic areas are now:

1. **Galaxy**
2. **Colonies**
3. **Fleets**
4. **Research**
5. **Diplomacy**
6. **Espionage**

Desktop:

- the left navigation rail should expose Galaxy, Colonies, Fleets, Research, Diplomacy and Espionage directly;
- Settings/session/diagnostics remain separate low-priority controls rather than gameplay destinations.

Mobile:

- keep the five-slot bottom bar practical rather than crowding it;
- Galaxy / Colonies / Fleets / Research remain direct bottom destinations;
- the fifth `More` destination contains direct first-class entries for Diplomacy and Espionage plus settings/session controls;
- contextual deep links may take the player directly to Diplomacy/Espionage without changing gameplay semantics.

## Galaxy - primary strategic view

Galaxy is not a dashboard. It is the main **interactive 2D galaxy map**, inspired by the interaction model of classic MOO2 while using original MOOX visuals.

The map must support:

- stars/systems positioned in the authoritative generated galaxy;
- discovery/known-state presentation;
- ownership/faction information where player knowledge permits it;
- own and known foreign fleets;
- fleet locations and travel routes/targets;
- pan/zoom on desktop and touch pan/pinch-zoom on phone;
- tap/click a known star/system to open the star-system dialog;
- map selection and overlays must remain presentation-only; server projections/legal actions remain authoritative.

## Star-system dialog / system inspection

Clicking/tapping a discovered/known star opens an in-map **star-system dialog** rather than navigating away from the galaxy by default.

The dialog should visually show the system as a miniature orbital scene:

- star/sun in the center;
- all known orbital bodies arranged clearly around it;
- habitable/colonizable rocky planets;
- gas giants;
- hostile/uninhabitable planets where applicable;
- asteroid belts where applicable;
- colony/ownership markers and names where known;
- player fleets currently present in the system;
- fleet names and the ships belonging to each visible player fleet;
- system name prominently visible.

Fleet/ship inspection authority note:
- the current strategic model locates fleets authoritatively at star-system level (`AtSystemID`) or in transit to a destination system;
- keep the orbital picture itself free of inferred fleet placement. When own fleets/special vessels are present, expose a contextual `Fleets / Ships` action from the system dialog and open a dedicated nested fleet roster/detail view;
- concrete projected `Ship` objects may expose their authoritative design/spec loadout (hull, drive, computer, armor, shield, fuel cell/range and weapon mounts), while Colony Ship / Outpost Ship / Troop Transport remain selectable special strategic vessels even when the current model represents them only as `StrategicFleet` objects;
- do not visually bind a fleet to a particular planet/body unless the core exposes an explicit body-level location. If local planet-orbit positioning becomes gameplay-significant, introduce an authoritative body anchor (for example `AtBodyID`) and its movement/stacking semantics first;
- do not fabricate strategic damage. Tactical battle state tracks armor/structure damage while a battle is active, but `core.Ship` currently has no persistent partial-damage state and strategic resolution only permanently removes destroyed ships. Persistent ship damage needs its own authoritative state extension before the strategic HMI can show health/damage percentages.

Desktop:

- modal/large overlay on top of the 2D galaxy;
- enough room to show orbital bodies and a fleet/system information column.

Mobile:

- large dialog/bottom-sheet/full-height overlay is acceptable;
- orbital scene remains the primary visual focus;
- fleet/planet details may stack beneath or open as nested sheets;
- returning closes the system view back to the same galaxy-map position/zoom.

## Planet selection inside the system dialog

A planet/orbital body is selectable when the player's authoritative knowledge allows it.

### Colonizable planet

If the server exposes a legal colonization choice for the selected planet (for example because an eligible Colony Ship is present in the system):

- show a prominent **Colonize / Kolonisieren** action on that planet;
- the browser must not infer colonization legality from fleet graphics alone; it consumes server-projected legal choices;
- activating Colonize opens a confirmation dialog naming the target planet/system and the Colony Ship/fleet that will be consumed/used where that information is available;
- only after explicit confirmation is the authoritative colonization command submitted;
- on success, refresh authoritative state and offer/open the newly created Colony detail view.

If no legal colonization choice exists, the planet remains read-only and no fake disabled gameplay path should imply that the client owns the rule.

### Existing owned colony

Clicking/tapping a planet that contains one of the player's colonies opens the dedicated **Colony Detail** screen.

The same Colony Detail route must also be reachable from the Colonies table, so Galaxy and Colony management converge on one canonical colony screen.

## Colonies - management table

The primary Colonies screen is a **dense, sortable management table**, not a card gallery.

Each row represents one colony and should eventually expose the high-value at-a-glance information needed to manage many colonies quickly, such as:

- colony / planet / system name;
- population and capacity;
- Farmers / Workers / Scientists;
- food / production / research outputs where authoritative projection supports them;
- morale/pollution/status where supported;
- current construction project and progress;
- warnings/blockers;
- quick job controls.

Population/job interaction:

- desktop: click controls and drag/drop where practical;
- phone/tablet: touch drag/drop where reliable;
- always retain explicit accessible +/- or equivalent tap controls so drag/drop is never the only interaction path;
- all submitted population assignment remains revision-bound and server-authoritative.

Selecting a colony row opens the canonical Colony Detail screen.

## Colony Detail - full planet management screen

Owned colonies get a complete dedicated screen rather than being edited only inside the table.

The screen should evoke the classic planet-management flow while using the MOOX design language.

Required regions:

### Population/jobs header

At the top, prominently show colony population and the three job groups:

- Farmers;
- Workers;
- Scientists.

The player can redistribute population there using the same authoritative assignment mechanics as the table. Desktop drag/click and mobile touch/tap patterns may differ visually but must submit the same server command semantics.

### Planet / colony presentation

The main area should show:

- the planet prominently;
- planet/system/colony name;
- relevant environment/size/biome/status information from the player-safe projection;
- colony ownership/faction identity;
- contextual economic/status information without duplicating hidden formulas in React.

### Buildings / infrastructure

Show the colony's existing buildings/infrastructure clearly, preferably as an inspectable list/grid with localized names and icons/art assets once Slice 15.3 provides them.

### Construction / build queue

Desktop target:

- construction/current project and build queue on the **right side**, mirroring the fast-management intent of the classic screen;
- available legal construction choices should be searchable/inspectable and enqueue through authoritative commands.

Mobile target:

- same semantics in a stacked panel, drawer or tab/sheet without losing the planet/job context;
- current project/progress must remain easy to reach.

The current backend already exposes construction choices/commands for building and strategic ship projects; Slice 15.2 must project the complete player-safe data needed by this screen rather than reimplementing cost/legality in React.

## Fleets

The Fleets area should use **fleet/ship tiles grouped by location**.

Primary grouping:

- one section per star system for fleets currently located there;
- a separate in-transit grouping for travelling fleets where applicable.

Each group/card should expose:

- fleet name/id;
- constituent ships with names/classes/role information available to the player;
- current system or destination;
- contextual legal movement/colonize/outpost actions;
- selecting a fleet from Galaxy and selecting the same fleet in Fleets should cross-link to the same underlying fleet identity.

## Research

Research is a visual strategic screen organized around the **eight classic MOO2 research categories**:

1. Construction
2. Power
3. Chemistry
4. Sociology
5. Computers
6. Biology
7. Physics
8. Force Fields

Each category should be represented visually rather than as a generic form, showing current/progress/available field information and technology imagery once 15.3 assets exist.

The browser must not hard-code Creative/Uncreative research legality. Slice 15.2 should project the legal research selection granularity/options from the authoritative rules/state so the UI can correctly represent whether the player selects a field/category/application or receives all/limited outcomes according to the race traits and current rules.

## Diplomacy

Diplomacy is a first-class strategic area.

Initial implementation can directly expose the already-supported authoritative baseline:

- declare war;
- offer peace;
- accept peace;
- relationship/stance state and pending peace offers.

The presentation should be designed so later diplomacy breadth can grow without changing the primary navigation model.

## Espionage

Espionage is now a first-class planned strategic area, but it must not be faked in React.

Current repository audit on 2026-09-04 found no Spy/Espionage command family in the backend.

Therefore Slice 15.2 Gate 1 must:

1. audit original/research evidence and desired minimum espionage gameplay;
2. identify required authoritative state, legal actions and command semantics;
3. decide whether a minimal espionage baseline can safely fit inside 15.2 or needs a dedicated sub-slice/mechanics slice;
4. keep the UI/navigation slot planned even if the mechanics are implemented one slice later.

Until authoritative mechanics exist, the user-facing production UI must not expose fake actionable espionage controls.

## Visual identity / icon direction

The current **OX** lettermark is accepted as a useful brand element and should remain visible in the shell unless the 15.3 visual-direction gate explicitly replaces it.

Slice 15.3 should explore an original compact MOOX app/product icon that remains legible at phone/favicon sizes, with the preferred concept space combining:

- spacecraft/ship silhouette;
- star/orbit/space motif;
- optionally a subtle X/orbit crossing;
- no dependence on copied original MOO2 art.

The 2D galaxy, star-system orbital dialog, planets, ships/fleet tiles, colony screen/buildings and eight research-category visuals are explicit 15.3 integration targets.

## Authority boundary

All of these views are presentation/orchestration only.

The server remains authoritative for:

- discovered/known information;
- fleet/system/planet state;
- colonization eligibility and target choices;
- Colony population/job legality;
- construction costs/choices/queue commands;
- fleet movement/legal targets;
- research options and Creative/Uncreative behavior;
- diplomacy;
- future espionage mechanics.

The UI may optimize interaction and visualization but must not independently derive hidden/legal game outcomes.

## Gate-2 accepted functional refinements - 2026-09-04

The player refined and accepted the functional contract before 15.2 implementation:

- gas giants and asteroid belts become persistent authoritative orbital bodies rather than presentation-only decoration; Outpost Ships must be able to establish legal Outposts on them as in original MOO2;
- star spectral/class identity is retained for the functional system view, while final artwork remains 15.3;
- Colony construction becomes a real ordered persistent queue with **no user-visible item-count cap** in MOOX; the original seven-slot limit is deliberately not carried forward;
- the dedicated build workspace follows a responsive catalog -> selected-project detail -> current build/queue information architecture inspired by the classic screen: three functional columns on wide desktop, a two-column collapse at intermediate widths, and vertical stacking on mobile;
- selecting a buildable is inspect-first; enqueue remains an explicit authoritative `colony.set_construction_queue` planning action;
- aborting the current build is destructive enough to require an explicit confirmation dialog, but it must describe the actual core semantics: accumulated construction PP are preserved and transferred to the new queue head or construction reserve rather than silently discarded;
- military Ship Design choices may reserve a disabled future `Ship Designer` affordance in Slice 15.2, but no working-looking designer route may be fabricated before the real Ship Designer workflow exists;
- Colony Detail's construction editor uses available-buildables on one side and the ordered queue on the other; normal Colony Detail and Colony table always expose the current build and ETA;
- Colony-table same-row population moves reassign Farmer/Worker/Scientist jobs, while cross-row moves initiate Population transfer with a server-derived Freighter/ETA confirmation; the destination column is the desired destination job;
- job changes update Food/PP/RP plus build/research timing immediately through a pure non-mutating authoritative Planning preview; React never owns those formulas;
- Research ETA uses projected empire-wide RP/turn and changes when Scientist assignments change on any Colony; build ETA changes with projected Workers/PP;
- the running-game sticky status area permanently surfaces BC balance/net income, Freighters available/total and Command Points used/capacity;
- Espionage remains first-class in navigation but all actionable mechanics move to prepared **Slice 20 Espionage / Intelligence baseline**; 15.2 must not fake Spy controls.

Permanent Gate-2 authority contract: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE2_2026-09-04.md`.

### Colony overview refinement

The full Colony Detail screen must also function as a compact planet/Colony round-up:

- Population / capacity and free capacity;
- visible Population growth per turn (or starvation/loss) and next-Pop ETA where finite;
- climate / size / minerals / gravity;
- Food / Production / Research and available economic contribution;
- existing Buildings/infrastructure as an inspectable list/grid;
- current build/progress/ETA plus upcoming queue.

The functional data already present in Core (`Colony.Buildings`, `ColonyPopulationDynamics`, Colony economy and Planet traits) is authoritative. MOOX should not invent unsupported Morale/Pollution values solely for visual imitation. Final planet/building art remains 15.3.
### Colony and uncolonized-planet presentation authority note

- Colony Detail should use the classic information hierarchy (planet profile + job/output assignment + current build + colony/building surface) while preserving the existing authoritative drag/drop population workflow.
- Planet traits must be explanatory, not raw IDs: climate should expose base food/habitability, size its base capacity, mineral class its worker-production potential, and gravity the viewing population's penalty.
- Uncolonized planet inspection must consume a player-safe server projection derived from EconomyRules. The browser must not hardcode MOO2 food/industry/gravity/capacity tables or infer race effects itself.
- `PlanetPotential` is explicitly pre-Colony/base context: race job modifiers and player gravity/capacity are valid, while Government/morale and Colony-local building effects remain outside that base projection.
- The live Colony job bands display actual authoritative `adjusted_economy` totals; per-Pop base values are explanatory context only.
- Colony BC contribution may surface authoritative `adjusted_economy.tax_bc` directly.
- Do not add Morale/Pollution fields unless the server owns and projects them. Rich terrain/building artwork remains Slice 15.3 presentation work.
### Morale presentation boundary

- Local Morale is not an invented future mechanic: Phase-1 authoritative economy context already supports government barracks penalties, morale buildings and Unification immunity, and those effects participate in adjusted Colony outputs.
- Slice 15.2 must nevertheless not display a guessed Morale number because `MoralePercent` / its source breakdown are not currently part of the player-safe browser DecisionView.
- When Morale is projected player-safely, display it in the compact Colony Profile as a signed percentage (positive green, negative red, zero neutral) with a click/tap breakdown of authoritative sources rather than duplicating the full rule explanation in the main screen.
- Full morale breadth (empire-wide morale technology, Capitol loss, conquered-population/assimilation morale) remains part of the evidence-driven economy/building/pollution/morale fidelity backlog after Slice 17; it should not reopen Economy as one monolithic rewrite.
