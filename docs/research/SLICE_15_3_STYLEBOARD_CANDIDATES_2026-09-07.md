# Slice 15.3 Gate 1 - visual styleboard candidates

Date: **2026-09-07**

Live review artifact: `/styleboard.html` on the canonical MOOX review server.

## Goal

Compare three representative art-direction candidates against the same strategic information architecture and icon concepts. The candidates vary surface treatment, palette, glow and density without inventing different gameplay.

The styleboard deliberately includes the Gate-2 representative surfaces requested by Slice 15.3:

- OX/MOOX brand treatment;
- strategic shell/navigation/resources;
- 2D Galaxy;
- star-system/orbit view;
- Colony metrics/population roles;
- Fleet/ship presentation;
- all eight Research categories.

## Candidate A - Deep Space Command

**Status: recommended; matches the current accepted implementation.**

Characteristics:

- deep navy/black strategic background;
- restrained electric-blue action/selection accent;
- semantic green/amber/red state colors;
- moderate rounded corners;
- low-to-medium glow reserved for selection, stars and important vectors;
- precise vector-first icons with thin consistent stroke;
- high information density without making the strategic screen feel like an engineering panel.

Why it is recommended:

- it preserves continuity with the already accepted Slice-15.1/15.2 HMI;
- the current implementation has already been exercised across 791px and 320px review widths;
- small semantic markers remain legible without requiring aggressive glow;
- it leaves visual headroom for richer future planet/portrait artwork;
- it does not visually compete with the procedural ship silhouettes.

## Candidate B - Orbital Glass

Characteristics:

- teal/cyan primary accent;
- brighter luminous edges and glass-like surfaces;
- softer/larger corner radii;
- more atmospheric glow;
- more open/airy presentation.

Strengths:

- visually attractive for showcase/marketing surfaces;
- gives stars/system objects a strong luminous identity;
- can feel more modern and premium.

Risks:

- glow competes with dense status values and tiny strategic markers;
- softer glass panels reduce contrast between hierarchy levels;
- requires more careful dark-mode contrast testing on low-quality mobile displays.

Recommendation: retain as a reference direction, not the strategic default.

## Candidate C - Industrial Tactical

Characteristics:

- charcoal/brown-black surfaces;
- amber action accent;
- smaller radii/harder borders;
- denser instrument-panel feel;
- lower decorative softness.

Strengths:

- excellent for tactical/instrumented information presentation;
- high differentiation of boundaries and controls;
- a useful vocabulary for Slice 15.5 tactical combat.

Risks:

- feels more like a command console than the broader Master of Orion X strategic layer;
- amber dominance reduces the clear separation between warning and primary-action color;
- less continuity with the already accepted blue MOOX shell.

Recommendation: consider parts of this language for Tactical 15.5, not as the strategic default.

## Gate-2 recommendation

Freeze **Candidate A / Deep Space Command** as the strategic visual baseline, while allowing screen-specific sublanguages that preserve semantic tokens:

- Tactical may borrow Candidate C's harder instrument treatment;
- showcase/hero art may borrow Candidate B's atmospheric glow;
- semantic action/status colors and icon identities remain shared.

This keeps MOOX coherent without requiring every screen to look mechanically identical.

## Responsive validation

The static styleboard itself is responsive and was browser-validated:

- 791px review width: all three candidate boards render without document overflow;
- 320x646: each board stacks to a single content column, width 296px inside the 320px viewport, no horizontal overflow;
- all three boards include eight Research SVG identities (24 category instances total).

The styleboard is a review artifact only and is intentionally outside the game's route/navigation structure.
