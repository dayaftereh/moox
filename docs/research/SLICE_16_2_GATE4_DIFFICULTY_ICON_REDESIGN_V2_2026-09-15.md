# Slice 16.2 Gate 4 - Difficulty icon redesign V2

Date: 2026-09-15
Status: **visual review candidate; Gate 4 remains open**

## User feedback

The first production Difficulty command-crest family was technically valid and broadly acceptable, but the final/Impossible icon accumulated too many visual overlays. The user explicitly requested a substantially new icon concept rather than another small adjustment to the same shield/rank language.

## Previous visual problem

The V1 command-crest generator increased perceived difficulty by stacking additional visual layers:

- shield outline + inner shield outline;
- outer halo/ring;
- separate glow-shield pass;
- crown/spike stack;
- central circle/crown/core;
- multiple rank chevrons;
- side guides;
- Impossible added a broken multi-segment outer ring.

This made Impossible the busiest composition rather than the clearest/strongest one.

## V2 direction: minimalist threat sigils

The command-shield metaphor is intentionally removed.

V2 uses one dominant geometric threat symbol per difficulty, large negative space, a quiet shared space background and nearly flat element complexity across the whole family.

Common rules:

- 1200 x 675 / 16:9 SVG;
- 12 sparse background stars only;
- one subtle radial aura behind the symbol;
- no shield;
- no rank-chevron stack;
- no crown/spike stack;
- no broken multi-ring overlay;
- no duplicated blurred outline pass;
- no embedded localized text;
- deterministic vector generator remains authoritative;
- existing semantic asset IDs/paths remain unchanged.

## Five new symbols

### Easy

A calm open orbital arc around a compact core, with one small orbital endpoint.

Intent: open, approachable, low tension.

### Normal

A clean closed hexagonal contact around the same compact core.

Intent: balanced, stable and neutral.

### Hard

A large triangular threat marker with a single central diamond.

Intent: the first clearly aggressive/angular warning shape.

### Very Hard

A razor diamond with one smaller inner diamond.

Intent: more compressed, focused and dangerous without adding decorative layers.

### Impossible

A single eight-point singularity flare behind a dark central core with one small energy center.

Intent: strongest silhouette of the family while staying visually simpler than the previous Impossible crest. It communicates an extreme/terminal threat through shape and contrast rather than through stacked badges, spikes and rings.

## Complexity change

The new main symbols stay intentionally flat in complexity:

- Easy: 1 main path + core circles;
- Normal: 1 main path + core circle;
- Hard: 2 paths;
- Very Hard: 2 paths;
- Impossible: 1 main flare path + 2 circles.

The previous Impossible composition combined multiple ring segments, glow/shield layers, seven spikes and five rank marks. V2 removes that accumulation entirely.

## Colour progression

The family keeps a readable difficulty-temperature progression while using geometry as the primary signal:

- Easy: cyan;
- Normal: cool blue;
- Hard: amber;
- Very Hard: orange;
- Impossible: red/magenta.

Colour is supportive, not the only differentiator.

## QA

After regeneration:

- `npm run build` passes;
- UTF-8/mojibake guard passes;
- deterministic New Game contract check passes for all ten New Game SVG assets;
- `git diff --check` passes;
- real Chrome browser smoke passes for Galaxy Size + Difficulty on mobile and desktop;
- all five Difficulty SVGs load at native 1200 x 675;
- mouse/touch/keyboard selector behavior is unchanged;
- server-derived Difficulty facts remain unchanged;
- no gameplay/server code was modified.

## Gate-4 status

This is a **visual review candidate**, not Gate-4 closure evidence.

Gate 4 remains open until the user reviews and accepts the redesigned family. Further visual tuning may replace/regenerate these assets without reopening Gate 3 or changing the frozen Difficulty mechanics.
