# Slice 15.3 Gate 1 - full-random ship morphology generator v3

Date: **2026-09-07**

Status: **implemented, browser-validated, primary Shipbuilder UX**.

## Product decision

The directed-evolution prototype proved that MOOX can mutate exact visual genomes, but it is **not the intended primary in-game Shipbuilder flow**.

The accepted player interaction is deliberately simpler:

`choose hull size -> see one large random ship -> click the ship or Generate -> completely new ship -> repeat until one looks right -> Use this design -> equip it`

The user should not need to understand parents, generations, mutation rates or genome locks merely to get a ship silhouette they like.

A click is therefore a **full reroll**, not a child mutation.

## Why v2 still looked too similar

Visual Genome v2 varied:

- station widths;
- concavity;
- Wedge / Spike / Pod details;
- Style DNA;
- symmetry;
- cutouts.

However, almost every candidate still shared one strong assumption: a long central fore/aft hull envelope. Detail mutation could make one design sharper, broader or more organic, but the overall visual category remained recognizable as a stretched rocket.

That was too little topological variation for the desired experience.

## Visual Genome v3

Visual Genome v3 adds an explicit **morphology / body-topology gene** above the detail grammar.

A full random roll independently chooses:

1. hull class - fixed by the player;
2. one coarse morphology family;
3. one Style-DNA family;
4. deterministic core-envelope variation;
5. signature morphology primitives;
6. additional random Wedge / Spike / Pod structures;
7. concavity / cutouts;
8. engine count and detail density.

The result remains deterministic for the same seed, but successive Shipbuilder rolls deliberately use unrelated seeds rather than evolving from the previous candidate.

## Eight morphology families

### 1. Needle / Cigarette

Purpose: preserve the possibility of a very long, thin, fast-looking ship.

Bias:

- longest envelope;
- strongly reduced beam;
- light core density;
- very high symmetry;
- Spike-heavy detail grammar.

Observed Scout sample: roughly **8.4:1 to 9.15:1** rendered structural aspect ratio.

### 2. Barge / Heavy

Purpose: short, thick, industrial mass.

Bias:

- much shorter envelope;
- strongly enlarged beam;
- dense core;
- broad central Pod structures.

Observed Scout sample: about **1.1:1 to 1.5:1**.

### 3. Manta / Widewing

Purpose: a strongly lateral, wing-dominated silhouette rather than a tube.

Bias:

- short/wide core;
- very large beam;
- two signature swept Wedge structures;
- high symmetry;
- optional extra negative space.

Observed Scout sample: about **1.1:1 to 1.5:1**.

### 4. Fork / Twin boom

Purpose: visibly split/twin-forward structure.

Bias:

- moderate length;
- thin central body;
- paired forward Spike structures;
- centerline negative-space opening where hull class allows it.

Observed Scout sample: about **2.6:1** in the representative browser run.

### 5. Chevron / A-shape

Purpose: produce A/V-like abstract spacecraft silhouettes.

Bias:

- short central hull;
- high lateral span;
- oversized swept Wedge primitives;
- strong bilateral symmetry.

Observed Scout sample: about **1.5:1**.

### 6. Hammer / Front-heavy

Purpose: move visual mass toward the bow and break the uniform cigar distribution.

Bias:

- short body;
- front-weighted station envelope;
- broad forward Pod/Wedge signature.

Observed Scout sample: about **1.4:1 to 1.8:1**.

### 7. Bulb / Blob

Purpose: deliberately compact, almost round / bobble-like spacecraft.

Bias:

- shortest length family;
- largest relative beam;
- dense body envelope;
- Pod-dominated side masses;
- additional negative-space possibility.

Observed Scout samples ranged from roughly **0.7:1 to 0.98:1**, meaning some generated bodies are literally taller/wider than they are long.

### 8. Asymmetric / Oddform

Purpose: allow deliberately strange, non-mirrored ships.

Bias:

- medium-short body;
- broad envelope;
- low primitive mirror probability;
- explicit non-mirrored Wedge / Pod / Spike signature parts.

Observed Scout sample: roughly **1.8:1 to 2.5:1**.

## Randomness axes

A single Generate click can now change all of these at once:

- morphology family;
- Style DNA (`spear`, `sleek`, `organic`);
- total hull length;
- beam;
- fore/aft mass distribution;
- station-width jitter;
- notch placement/depth;
- primitive count;
- primitive family;
- primitive placement;
- primitive sweep;
- primitive width/length;
- mirrored versus non-mirrored detail;
- cutout count/placement;
- engine count.

Hull class still controls the broad **scale and complexity budget**, so a Scout remains a small-class visual design and a Doom Star remains a large-class design, but class no longer forces one aspect ratio or one silhouette topology.

## Primary Shipbuilder UX

The previous six-candidate population, Parent, Generation, Shape Mutation and lock panel are removed from the primary Shipbuilder view.

The current view presents:

- hull-size selector;
- hull-size thumbnails use one consistent Manta reference morphology so Scout -> Doom Star scale remains visually comparable; full randomness starts in the large candidate stage;
- one large generated ship;
- click-on-ship reroll interaction;
- explicit Generate button;
- Use this design button;
- small diagnostic labels for current morphology / Style DNA / generated part count;
- kept-design preview;
- existing server-authoritative equipment baseline.

Clicking the large ship and pressing Generate are equivalent.

`Use this design` stores the exact in-memory Visual Genome v3 candidate in the kept-design state. Subsequent rerolls do not mutate or replace the kept design.

## Rendering envelope

`ProceduralShipGlyph` now uses a **genome-dependent SVG viewBox**: width grows with generated hull length and height grows with generated beam, with a 180 x 120 minimum. This keeps Needle designs readable while giving broad Manta/Bulb/Chevron ships - including Titan/Doom Star variants - enough vertical and lateral room without clipping.

The generator remains fully SVG-native and resolution-independent.

## Browser QA evidence

Validated against the still-running in-memory `game-1`; the preview server was not restarted.

### Desktop 995 x 605

- no persistent server-error banner;
- primary Shipbuilder contains **zero** old six-candidate cards;
- one large clickable generated ship is present;
- all generated SVGs report `genome version = 3`;
- no horizontal page overflow;
- 24 rolls were sufficient to encounter all eight morphology families;
- observed structural ratios spanned from a compact Organic Bulb at about **0.70:1** to a Sleek Needle at about **9.15:1**;
- Use this design preserved the exact Bulb genome while a following reroll changed the current candidate to a Hammer design.

Representative first-seen samples in the validation run:

| morphology | style | approximate structural ratio | parts |
| --- | --- | ---: | ---: |
| Fork | Organic | 2.65 | 7 |
| Chevron | Sleek | 1.51 | 5 |
| Needle | Spear | 8.40 | 6 |
| Barge | Sleek | 1.50 | 4 |
| Asymmetric | Spear | 2.15 | 7 |
| Manta | Organic | 1.10 | 6 |
| Hammer | Organic | 1.43 | 6 |
| Bulb | Organic | 0.70 | 7 |

### Mobile 360 x 646

- one large ship stage remains usable;
- metadata collapses to one column;
- Generate / Use-this-design actions collapse to one column;
- ten consecutive rerolls covering Needle, Bulb, Manta, Barge, Hammer and Asymmetric showed no horizontal overflow;
- aspect-ratio extremes remain visible on the same responsive SVG stage.

## Relationship to the earlier directed-evolution prototype

The earlier v2 work is retained as useful research and code capability:

- exact parent mutation;
- genome locks;
- mutation semantics;
- Style DNA.

However, it is now **historical/advanced generator capability**, not the primary in-game interaction.

The player-facing flow is full-random reroll because that better matches the intended rapid visual search behavior.

## Next work

The next generator work should focus on **quality and persistence**, not reintroducing evolution complexity:

- add more topology families only when they create clearly new silhouettes;
- tune per-hull morphology envelopes so large ships become even more massive without losing topology diversity;
- optionally add race/faction bias as a probability weighting, while still allowing broad random variation;
- persist the accepted Visual Genome v3 in the authoritative/save design contract;
- define migration/fallback behavior for future genome versions;
- later connect the kept visual design to Slice 17 ship equipment editing.