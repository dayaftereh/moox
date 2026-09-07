# Slice 15.3 Gate 1 - Proceduro directed-evolution reference and MOOX adaptation

Date: **2026-09-07**

Status: **reference studied, provenance recorded, MOOX-native adaptation implemented**.

Reference repository: `https://github.com/proceduro/spaceship-2d`

## Why this reference matters

The user identified `proceduro/spaceship-2d` as a strong visual reference because its generated ships are sharp, angular, highly varied and more abstract than a single smooth/convex hull outline. The interaction model also maps unusually well onto the desired MOOX Shipbuilder: generate several options, choose the one that looks best, then generate descendants from that choice.

This document records what the reference actually does and which ideas MOOX adopts or deliberately does not adopt.

## Source-derived findings

### Directed evolution UX

The repository README describes an explicit directed-evolution workflow:

1. a set of generated ships is shown;
2. the user clicks the most attractive ship;
3. a new set is generated from that selected ship with its attributes changed slightly;
4. mutation rates can be adjusted for base color, detail color and shape;
5. a chosen ship can be saved at a selected resolution.

The renderer can export diffuse, normal, depth and position sprites.

Source: `README.md` in `proceduro/spaceship-2d`.

### Geometry mechanism

The core generator is not a conventional 2D polygon-outline generator. `spaceship2d/index.js` creates a ship code containing multiple randomly transformed tetrahedron primitives. Each primitive receives independently randomized/mutated:

- color;
- X/Y/Z rotation;
- X/Y/Z scale;
- translation along the ship axis.

The generated primitives are merged on one side, cloned/reflected for bilateral symmetry, merged again, flattened in Z, centered/normalized and rendered through WebGL.

Mutation changes the existing primitive transforms and colors and can also add or remove primitive elements. That add/remove primitive behavior is an important contributor to the sharp, layered silhouette diversity.

Source: `spaceship2d/index.js` in `proceduro/spaceship-2d`.

### Rendering mechanism

The reference renderer uses WebGL and several GLSL programs / framebuffers to derive diffuse, normal, position and depth information plus SSAO/final color output. This explains the rich pseudo-3D sprite presentation, but it is materially heavier than MOOX needs for the current strategic HMI.

Source: `spaceship2d/index.js` and `spaceship2d/glsl/`.

### License / provenance

The repository `LICENSE` states that the software is released into the public domain under the **Unlicense**, permitting copy, modification, publication, commercial/non-commercial use and distribution.

MOOX nevertheless does **not** vendor or copy the reference implementation in this Gate-1 work. The implementation below is an original TypeScript/SVG adaptation of the general techniques so it fits MOOX determinism, responsive rendering and future design persistence.

## MOOX adaptation decision

MOOX adopts four principles from the reference:

1. **directed evolution** rather than blind one-at-a-time rerolling;
2. **many independent sharp primitives** layered onto a core hull grammar;
3. **mutation of an accepted parent** rather than replacing the entire design identity;
4. **add/remove primitive mutation** so descendants can structurally diverge over multiple generations.

MOOX deliberately changes the technical implementation:

- SVG/React instead of runtime WebGL sprite generation;
- deterministic seeded PRNG instead of `Math.random()`;
- explicit `ShipVisualGenome` data rather than an implicit render-only mesh code;
- hull-class size/concavity constraints retained from existing Slice 15.3 work;
- gameplay equipment remains server-authoritative and separate from visual evolution;
- currentColor/CSS token theming remains available;
- the generator stays lightweight enough for mobile HMI use.

## Visual Genome v2

New file: `web/src/shipVisualGenome.ts`.

The genome now stores independent genes for:

- hull class and perceptual length/beam;
- station widths forming the core hull envelope;
- deterministic concave notch depths;
- engine count;
- surface detail count;
- transparent negative-space cutouts;
- a variable list of sharp visual primitives.

### Primitive families

Gate-1 v2 introduces three SVG-native primitive families:

- **wedge** - swept angular wing/body plate;
- **spike** - aggressive triangular projection;
- **pod** - pointed diamond-like secondary mass.

Primitives are usually mirrored across the centerline but the genome preserves the per-primitive side/mirroring fields so later art-direction work can introduce deliberate asymmetry.

Hull size controls primitive-count envelopes. Representative ranges currently grow from roughly 3-5 primitives for Scout to 12-20 for Doom Star before mutations add/remove parts.

## Deterministic mutation

`mutateShipGenome(parent, mutationSeed, amount)` mutates a concrete resolved parent genome.

The mutation amount affects:

- core station width drift;
- notch depth;
- primitive position/length/width/sweep;
- cutout position/size/rotation;
- engine-count variation;
- probability of adding/removing one primitive.

The result is deterministic for the same parent genome + mutation seed + mutation value.

This solves an important limitation of the earlier seed-only generator: an evolutionary child can remain genuinely similar to the exact selected parent because the parent genome itself is the mutation source.

## Shipbuilder UX v2

The one-at-a-time Generate stage is replaced by a six-candidate population.

Current flow:

`choose visual size -> inspect six candidates -> click/Evolve preferred candidate -> inspect six mutated descendants -> repeat -> Keep any favorite -> later equipment handoff`

Controls:

- **Shape mutation** slider, currently 5%-90%;
- **New random family** resets to a fresh six-candidate root population;
- clicking the candidate artwork or **Evolve** chooses that genome as parent;
- **Keep design** pins any candidate independently of the active evolution line.

The current generation shows the selected parent in a separate card, making lineage visible and understandable.

## Browser validation

Validated against the still-running fresh `game-1`; no server restart was required.

Desktop **995x605**:

- exactly six candidate ships render;
- all candidate SVGs report genome version 2;
- root Scout population showed varied primitive counts;
- selecting candidate #2 advanced Generation 1 -> Generation 2;
- the selected parent card preserved candidate #2 / Generation 1;
- Generation-2 body paths remained close to the selected parent at the default 32% mutation while still differing;
- most children retained the parent primitive count, with at least one deterministic add/remove variation;
- Keep design preserved a Generation-2 candidate independently;
- no horizontal overflow.

Mobile **360x646**:

- six candidates render as a two-column grid;
- mutation controls collapse to one column;
- no horizontal overflow;
- SVG genome v2 candidate rendering remains active.

## Originality / provenance rule

The Proceduro project is treated as a **documented algorithmic/UI reference**. MOOX source must keep this provenance note but should not imply that generated MOOX artwork is authored by Proceduro or that MOOX ships reproduce reference assets.

Our generated designs arise from MOOX-specific hull profiles, seeds, primitives, concavity rules, responsive SVG rendering and future faction grammars.

## Next Gate-1 work

With directed evolution now established, the next useful layer is **art-direction/faction style DNA**. The same genome should be biased into clearly different families without changing gameplay rules, for example:

- spear / angular military;
- sleek / high-technology;
- organic / alien;
- later crystalline, industrial, stealth or intentionally asymmetric families.

Gate 1 should compare 2-3 representative families in the same six-candidate evolution UI before Gate 2 freezes the generator/pipeline contract.
## Primary-UX follow-up (2026-09-07)

Directed evolution remains useful algorithmic research, but subsequent user review rejected it as the primary Shipbuilder interaction. MOOX now uses full-random Visual Genome v3 rerolls for the player-facing flow while retaining the reference-derived primitive/mutation ideas as internal generator capability. See docs/research/SLICE_15_3_FULL_RANDOM_SHIP_MORPHOLOGY_V3_2026-09-07.md.
