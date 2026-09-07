# Slice 15.3 Gate 1 - Shipbuilder generation UX

Date: **2026-09-07**

Status: **product flow captured; browser prototype implemented against a fresh real game**.

## Accepted product intent

The MOOX Shipbuilder should not begin with a fixed gallery of ship pictures. The desired player flow is:

1. choose a ship/hull size;
2. press **Generate** repeatedly to explore new procedural 2D ship concepts;
3. keep/select a visually appealing design;
4. equip that selected design with weapons and other ship components;
5. preserve the chosen visual identity so the same design remains recognizable across the game.

This makes visual ship creation part of the player's design process while avoiding a large fixed sprite library.

## Gate-1 browser prototype

The first playable visual prototype is available as the in-game `shipbuilder` route and from the game menu.

It currently provides:

- hull-size selectors for Frigate, Destroyer, Cruiser, Battleship and Titan;
- a large procedural SVG candidate preview;
- a **Generate** action that advances through a deterministic candidate sequence;
- a recent-generation strip so prior candidates can be revisited;
- a **Keep design** action that pins the current visual favorite;
- an equipment panel showing the current authoritative ship-design baseline from the running game.

### Current rules boundary

The current runtime ruleset only exposes the narrow authoritative **Frigate** military baseline. In the fresh New Game used for Gate-1 validation this is the `Scout` design.

Therefore:

- Frigate is labelled as the **current ruleset**;
- Destroyer, Cruiser, Battleship and Titan are explicitly labelled **visual preview** in Slice 15.3;
- visual preview hulls do not create gameplay designs, costs, capacities, weapons or legality that the server does not yet support;
- authoritative multi-hull, component and weapon breadth remains owned by **Slice 17**.

Slice 15.3 is allowed to establish and test the visual generation grammar now so Slice 17 can later attach real design rules to an already-accepted visual workflow.

## Deterministic generation contract

Pressing Generate does not use uncontrolled frame-time randomness.

Current candidate seed form:

`shipbuilder:<game-id>:<hull-id>:<generation>`

Examples:

- `shipbuilder:game-1:frigate:1`
- `shipbuilder:game-1:frigate:2`
- `shipbuilder:game-1:destroyer:1`

For a given generator version and seed, the same candidate must render the same silhouette. This makes the generation sequence reproducible and suitable for later save/resume/replay persistence.

The current Gate-1 candidate seed is presentation state only. A later authoritative design contract should persist the accepted visual seed/version with the ship design rather than letting the client silently regenerate a different appearance.

## Generator grammar direction

The current SVG proof of concept already varies the major hull envelope, wing/body station, engines, panel lines and visible weapon hardpoints. Gate 2 should define the richer grammar around independent layers:

1. hull class / scale;
2. faction or race visual family;
3. major silhouette seed;
4. engine arrangement;
5. functional hardpoint/module zones where authoritative components exist;
6. cosmetic armor/panel/fin/antenna details;
7. empire/faction palette;
8. optional small serial variation that never destroys recognition of the parent design.

Useful future player-facing generation controls can include selective locks such as **keep nose**, **keep wings**, **keep engines**, **symmetry/asymmetry**, or **regenerate details only**. These are visual controls and should not be confused with gameplay component legality.

## Equipment handoff

The intended long-term transition is:

`choose size -> generate visuals -> keep design -> equip authoritative components -> save design`

The Gate-1 equipment panel is deliberately read-only. It shows the server-projected baseline design so the UX can be evaluated without inventing unimplemented ship-design mechanics.

When Slice 17 arrives, the equipment step can become authoritative using the server-supported hull capacities, technologies, components, weapons, costs and legality while retaining the selected procedural visual seed.

## 2D now, future 3D path

Runtime presentation remains 2D/SVG for Slice 15.3. The generator geometry should remain semantic enough that a future renderer could use the same design seed and zones to:

- extrude the hull;
- generate layered height profiles;
- place engines/module sockets consistently;
- render richer offline sprites or a future native 3D tactical representation.

Runtime 3D is not required to achieve the current product goal.

## Fresh-game validation evidence

A clean server was restarted on 2026-09-07 without `-demo-fixture` and without persistence, so all previous in-memory hosted games were removed. A fresh real `game-1` was then created through the normal browser New Game flow with seed `0x8009`.

Fresh authoritative snapshot:

- 20 generated star systems;
- one Human Scout ship design;
- authoritative hull: Frigate;
- two ships;
- two strategic fleets.

Real generated galaxies exposed a Slice-15.2 HMI robustness gap: some systems legitimately serialize absent `planets`/`bodies` slices as JSON `null`. The Galaxy HMI was hardened to treat those optional collections as empty instead of throwing during `.some()`, `.map()` or `.find()`. The fresh Galaxy then rendered normally.

The fresh Fleet view renders the procedural SVG ship visuals from the actual authoritative ships, proving the generator is not limited to the previous demo fixture.

## Browser QA

Desktop acceptance at **995x605**:

- Shipbuilder route renders from the real game;
- five hull-size choices render;
- Frigate is clearly distinguished from visual-preview hulls;
- Generate advanced `frigate:1 -> frigate:2 -> frigate:3`;
- prior generations remained selectable;
- Keep design preserved Frigate generation 3 while the active generator changed to Destroyer generation 1;
- current authoritative Scout equipment remained visible separately.

Mobile acceptance at **360x646**:

- no horizontal page overflow;
- five hull choices remain usable;
- Generate advanced the deterministic candidate;
- recent history appeared;
- Keep design pinned the favorite;
- game menu exposes Shipbuilder as a first-class item;
- current game status/resources and bottom strategic navigation remain available.

## Gate-2 decisions still open

- visual style families / faction grammars;
- how much symmetry versus intentional asymmetry is allowed;
- player-facing lock/regenerate controls;
- empire palette application and accessibility constraints;
- visual-seed/version persistence schema;
- mapping authoritative modules to visible geometry without forcing every component to be literal;
- minimum uniqueness/readability requirements by hull class and viewport size;
- whether a small set of authored silhouette primitives should complement the procedural grammar.

## Hull-scale / concavity update (2026-09-07)

The visual size selector now extends from a small **Scout role** through Frigate, Destroyer, Cruiser, Battleship, Titan and Doom Star. Scout is explicitly a visual role using Frigate rules, not a new authoritative hull. Larger classes use progressively larger perceptual envelopes, more forced concave side bays and, from Cruiser upward, transparent negative-space cutouts. Full measured profile/QA evidence is in docs/research/SLICE_15_3_HULL_SCALE_CONCAVITY_GRAMMAR_2026-09-07.md.

## Directed-evolution v2 update (2026-09-07)

The Shipbuilder now shows six candidates at once. Selecting a candidate makes its resolved Visual Genome v2 the parent for the next six deterministic mutations. A Shape Mutation slider controls mutation strength, New random family resets the root population, and Keep design pins a favorite independently. The geometry adds sharp Wedge/Spike/Pod primitives on top of the accepted class-scale/concavity grammar. Reference/provenance and implementation differences are documented in docs/research/SLICE_15_3_PROCEDURO_DIRECTED_EVOLUTION_REFERENCE_2026-09-07.md.

## Style-DNA update (2026-09-07)

A new Style-DNA stage now sits between hull size and directed evolution. Spear/Angular, Sleek/High-Tech and Organic/Alien each bias hull proportions, primitive families, sweep and symmetry while keeping gameplay rules unchanged. Switching style resets only the active evolution parent/family. Details and browser evidence: docs/research/SLICE_15_3_SHIP_STYLE_DNA_CANDIDATES_2026-09-07.md.
