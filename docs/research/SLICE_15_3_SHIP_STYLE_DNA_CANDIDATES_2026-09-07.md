# Slice 15.3 Gate 1 - Ship Style DNA candidates

Date: **2026-09-07**

Status: **three representative art-direction DNA families implemented and browser-validated**.

## Goal

After establishing Visual Genome v2 and directed evolution, Gate 1 needs representative art-direction families that can be compared inside the same Shipbuilder workflow. These families are visual-only biases; they do not change hull capacities, costs, weapons, legality or other gameplay authority.

The first three candidates are intentionally broad enough to expose whether the genome architecture can produce recognizably different visual languages before Gate 2 freezes the generator contract.

## Style DNA A - Spear / Angular

Intent:

- sharp;
- aggressive;
- military;
- visually closest to the user's preferred pointed/fancy Proceduro reference direction.

Genome bias:

- slightly longer hull envelope;
- slightly reduced beam;
- high probability of bilateral mirroring;
- primitive distribution strongly favors Wedge and Spike elements;
- primitive length is increased;
- large sweep range creates pointed fins, lances and hard angular projections;
- current concavity/cutout rules remain active.

Use case candidate:

- militaristic or predatory factions;
- aggressive player-made designs;
- strong default MOOX procedural visual direction if later review continues to favor this family.

## Style DNA B - Sleek / High-Tech

Intent:

- fast;
- precise;
- technological;
- controlled rather than busy.

Genome bias:

- longest relative hull envelope of the three styles;
- substantially reduced beam;
- ~96% primitive mirroring probability;
- Wedge primitives dominate;
- Spike probability is deliberately low;
- primitives are narrower and use a restrained sweep range;
- concave notch depth is reduced relative to Spear/Organic.

Use case candidate:

- advanced/high-technology factions;
- scout/recon/fast-attack visual language;
- ships that should read as engineered precision rather than brute mass.

## Style DNA C - Organic / Alien

Intent:

- alien;
- broad;
- less mechanically regular;
- visibly different from Human-like symmetric engineering.

Genome bias:

- slightly shorter relative length;
- substantially increased beam;
- lower mirroring probability (~68%), intentionally allowing more asymmetric primitive arrangements;
- Pod primitives dominate;
- primitives are broader and less lance-like;
- deeper waists/concavity than Sleek;
- existing negative-space cutouts remain active on larger hull classes.

Use case candidate:

- biological/alien faction language;
- less conventional ship silhouettes;
- a test bed for future controlled asymmetry rules.

## Implementation

`web/src/shipVisualGenome.ts` now defines:

`ShipStyleID = 'spear' | 'sleek' | 'organic'`

Every `ShipVisualGenome` stores its `styleId` as part of the persisted visual identity candidate.

`createShipGenome(seed, hullId, styleId)` applies style-specific biases to:

- hull length/beam;
- notch depth;
- primitive family probabilities;
- primitive length/width;
- primitive sweep range;
- mirror probability.

`mutateShipGenome(...)` preserves the selected parent's style DNA. Mutation operates inside that style family rather than silently drifting to a different art direction.

## Shipbuilder UX

A new **Style DNA** section sits between hull size and directed evolution.

The player can compare three live mini-previews for the currently selected hull:

1. Spear / Angular;
2. Sleek / High-Tech;
3. Organic / Alien.

Selecting a different style:

- resets the current evolution parent;
- returns the population to Generation 1;
- creates six fresh candidates from the selected style grammar;
- does **not** alter the active game or server-authoritative ship-design data;
- does not delete an independently kept design.

## Browser evidence

Validated on the still-running fresh `game-1`.

Desktop **995x605**:

- three Style-DNA options render with live SVG previews;
- default selected style is Spear / Angular;
- all six candidate genomes report `styleId=spear` initially;
- selecting Sleek resets to Generation 1 and all six candidates report `styleId=sleek`;
- selecting Organic resets to Generation 1 and all six candidates report `styleId=organic`;
- no horizontal overflow.

Representative Scout style-preview body bounds in the current deterministic preview seeds:

| style | body width | body height | representative primitives |
| --- | ---: | ---: | ---: |
| Spear / Angular | ~59.2 | ~15.6 | 3 |
| Sleek / High-Tech | ~60.4 | ~15.8 | 3 |
| Organic / Alien | ~54.2 | ~20.4 | 5 |

This confirms the intended directional difference: Organic is visibly broader, while Spear/Sleek remain long and narrow but differ primarily through primitive grammar and symmetry.

Mobile **360x646**:

- Style-DNA selector collapses to one column;
- six-candidate population remains a two-column grid;
- no horizontal overflow;
- all three style options remain accessible.

## Gate-1 interpretation

The architecture now demonstrates that one deterministic genome system can support:

- hull-class scaling;
- concavity and negative space;
- sharp primitive layering;
- parent/child directed evolution;
- distinct style/faction biases.

This is sufficient to begin discussing a Gate-2 generator contract without committing every race to one style yet.

## Next work

The next useful controls should focus on **player-directed partial locking**, for example:

- keep core hull;
- keep wings/primitives;
- keep engines;
- keep cutouts/negative-space pattern;
- regenerate details only.

These locks should operate on Visual Genome fields, making them deterministic and persistable rather than ad-hoc DOM effects.

After that, Gate 1 should map candidate style DNA to race/faction art-direction requirements and define the visual-seed/genome persistence/versioning contract for Gate 2.