# Slice 15.3 Gate 1 - Ship Genome lock controls

Date: **2026-09-07**

Status: **implemented and browser-validated**.

## Goal

Directed evolution is much more useful when the player can preserve the parts of a generated ship that already look right and mutate only the remaining visual genes. Gate 1 therefore adds deterministic partial locks to Visual Genome v2 rather than relying on ad-hoc UI-only freezes.

These controls are presentation-only. They do not lock or alter authoritative gameplay equipment, costs, hull capacities, technologies or weapon legality.

## Lock contract

`web/src/shipVisualGenome.ts` now defines:

```text
ShipGenomeLocks
- core
- primitives
- engines
- cutouts
```

`mutateShipGenome(parent, mutationSeed, amount, locks)` applies the lock set to the exact selected parent genome.

### Core hull lock

Preserves exactly:

- `stationWidths`;
- `notchDepths`.

Result: the central hull envelope, waist and outer concave bays remain unchanged while unlocked primitives, engines or cutouts may continue to evolve.

Hull class, style DNA, length and beam are lineage identity and already remain stable across normal parent mutation; they are not separate Gate-1 lock toggles.

### Wings / parts lock

Preserves the complete primitive list exactly, including:

- Wedge / Spike / Pod kind;
- position;
- side;
- length;
- width;
- sweep;
- mirrored/asymmetric state;
- primitive count.

While this lock is active, primitive add/remove mutation is also disabled.

### Engines lock

Preserves engine count exactly. Because hull length/beam remain stable inside a lineage, this also preserves the deterministic engine placement rendered from that count.

### Cutouts lock

Preserves every negative-space gene exactly:

- longitudinal position;
- side;
- offset;
- radius/size;
- rotation.

The resulting SVG-mask openings therefore remain unchanged while other unlocked geometry can mutate.

## Shipbuilder UX

A new **Preserve parts / Bereiche behalten** panel appears underneath Shape Mutation.

Four toggle buttons are available:

1. Core hull / Kernhülle;
2. Wings / parts / Flügel / Teile;
3. Engines / Triebwerke;
4. Cutouts / Aussparungen.

Each control exposes pressed/locked state through `aria-pressed`. A `0/4` through `4/4` badge gives immediate lock-count feedback.

Locks can be selected before or after choosing a parent. They become relevant whenever descendants are generated from a selected parent. `New random family` resets lineage but intentionally does not silently clear the player's preferred lock configuration.

## Determinism and exact preservation

The lock implementation copies locked genome fields from the selected parent instead of trying to re-create them from an approximate seed. This is important: a locked section is not merely "similar"; it is the same visual gene data.

The renderer now also exposes `data-engine-count` for QA alongside existing genome/style/primitive metadata.

## Desktop browser validation

Validated on the still-running fresh `game-1` at **995x605**, using a Spear/Angular Doom Star parent.

### All four locks active

Parent characteristics in the tested lineage included:

- Doom Star core hull with 3 concave notches and 2 cutouts;
- 17 sharp primitives;
- 4 engines.

With Core + Wings/Parts + Engines + Cutouts all locked:

- all six children had body paths identical to the parent;
- all primitive SVG paths were identical to the parent;
- all engine positions/counts were identical;
- all cutout ellipse geometry was identical;
- all six full geometry signatures matched the parent exactly.

This proves the four locks compose correctly.

### Only Wings / parts unlocked

The same Generation-2 child population was recomputed with:

- Core locked;
- Wings / parts unlocked;
- Engines locked;
- Cutouts locked.

For all six descendants:

- core body path remained exactly equal to the parent;
- engines remained exactly equal;
- cutouts remained exactly equal;
- primitive paths changed in every child;
- one deterministic child reduced primitive count from 17 to 16, demonstrating add/remove mutation remains active only for the unlocked primitive gene group.

This proves that partial locking isolates mutation to the intended gene group rather than freezing or rerolling the whole design.

## Mobile validation

At **360x646**:

- four lock controls render as a 2x2 grid;
- six ship candidates remain a two-column grid;
- Style DNA remains a one-column selector;
- no horizontal page overflow.

## Persistence implication

The lock state itself is a Shipbuilder editing preference and does not need to become part of a completed ship design. The **kept/accepted genome**, however, must eventually persist its concrete Visual Genome version/data or an equivalently stable canonical representation so a chosen design survives save/resume and future generator changes.

This distinction should be frozen in the Gate-2 visual persistence contract:

- edit locks = transient editor state;
- accepted visual genome = persistent design identity;
- gameplay equipment = separate server-authoritative design state.

## Next Gate-1 work

The remaining high-value generator work is now less about raw randomization and more about persistence and semantic structure:

- define canonical accepted Visual Genome serialization/versioning;
- define migration/fallback behavior if generator versions change;
- decide whether dedicated sub-gene locks such as Nose / Wing family / Engine-pod shape are needed beyond the four coarse groups;
- map Style DNA / faction families to race visual direction without making visual style a gameplay rule;
- prepare the Gate-2 visual pipeline and persistence freeze.