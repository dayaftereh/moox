# Slice 15.3 Gate 1 - Research, population and action icon refinement

Date: **2026-09-07**

Status: **implemented and browser-validated**.

## Product feedback incorporated

The first shared SVG language was accepted as a useful direction with three requested refinements:

1. the generic Research symbol should read as a **microscope**;
2. population people should carry their role identity directly rather than relying only on generic silhouettes;
3. important buttons should gain small explanatory icons without turning every control into decorated UI.

## Research icon

The registry key `research` now renders a microscope rather than the previous atom-like symbol.

This affects every existing use of the shared Research icon automatically, including the global RP resource chip and Research selection actions.

A separate `test-tube` icon is now also part of the registry. It is intentionally not used as the global Research symbol. It is reserved for future Chemistry / Biology / Pharma-like research categories where a laboratory-vessel metaphor is more specific than the microscope.

The older generic `science` flask remains available for compatibility while category-specific Research icon work is still pending.

## Population role people

The Farmer / Worker / Scientist headings already used dedicated role SVGs. This refinement extends the same identity into every draggable population marker:

- Farmer marker -> person + plant/leaf identity;
- Worker marker -> person + industrial/gear identity;
- Scientist marker -> person + laboratory/flask identity.

The old marker was a CSS-only generic head/body silhouette. The marker is now a real `GameIcon` instance and therefore uses the same vector grammar, scaling and role color as the heading.

Marker size is currently **18 x 18 px**. Fractional-population badges remain unchanged.

## Selective action-button icons

Button icons follow a deliberately selective rule: add them to high-value actions where the symbol improves scan speed or explains intent; do not add an icon to every text button.

Current integrations:

### Global / planning

- End Turn / Ready -> `check`

### Ship visual designer

- Generate -> `generate`
- Use this design -> `check`

### New game shell

- New Game -> `star`
- Resume -> `play`
- Create Game -> `star`

### Colony / strategic actions

- submit population assignment -> `check`
- open colony -> `open`
- fleet list/action -> `fleets`
- colonize -> `flag`
- build outpost -> `outpost`
- open construction manager -> `build`
- add construction item -> `build`
- open Ship Designer placeholder -> `ship-designer`
- confirm population transfer -> `check`
- select Research project -> `research` microscope

## New action icon vocabulary

Added to the typed registry for this pass:

- `test-tube`
- `generate`
- `check`
- `open`
- `play`
- `flag`
- `outpost`
- `build`

These remain part of the same 24x24/currentColor/1.55-stroke icon grammar.

## Button layout rule

The common primary / secondary / ghost / danger button classes now use inline-flex alignment with a **7 px** content gap. Leading button icons render at **15 x 15 px** by default and **13 x 13 px** in compact buttons.

This keeps label baselines stable and gives icons a consistent visual footprint across desktop and mobile.

## Browser QA

Current review server: port 7172, existing game-1 preserved.

### Desktop

- global RP icon renders the new microscope geometry;
- DOM confirms the Research SVG starts with microscope geometry and no longer contains atom ellipses;
- End Turn shows `check` at 15 px;
- Shipbuilder Generate shows `generate` at 15 px;
- Shipbuilder Use this design shows `check` at 15 px;
- no danger/error banner;
- no horizontal overflow.

### Mobile 320 x 646 Colony Detail

- Farmer heading and each Farmer population marker use `farmer` SVG;
- Worker heading and each Worker marker use `worker` SVG;
- Scientist heading and each Scientist marker use `scientist` SVG;
- individual markers render at 18x18;
- End Turn retains the common `check` icon;
- document width remains exactly 320 px in the tested view;
- no danger/error banner.

## Gate-1 status

This improves the same accepted icon foundation; it does **not** freeze final Gate-2 art direction.

Next relevant visual work remains:

- dedicated Research category icon family using microscope/test-tube/engineering/computer/etc. appropriately;
- Galaxy/star-system object markers;
- Fleet/entity markers;
- construction/building categories;
- broader action-button review after more screens mature.