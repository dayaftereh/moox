# Slice 15.4 Gate 3 - Research UX unification and technology information

Date: 2026-09-09
Status: **implemented**

## Goal

Remove the accidental duplicate Research selection surfaces and make every researchable technology inspectable before the player chooses it.

## Canonical Research surface

Research selection now has one canonical browser surface: `StrategicResearchOverlay`.

It is used by:

- the Research resource control in the top HUD;
- the stable `/research` game route;
- the Research Breakthrough transition action **Choose next research**.

The legacy standalone `StrategicResearchView` implementation has been removed. A Research Breakthrough therefore acknowledges only the presentation event and opens the same chooser the player already knows from the HUD; it does not invent a second selection workflow.

## Technology information

Every concrete Technology shown in the Research chooser now has a dedicated `?` control. The detail dialog shows:

- Technology name and Research category;
- authoritative TechField and RP cost;
- the active selection rule (`choose_one`, `all`, `fixed_one`, or repeat field) in localized form;
- concrete gameplay unlocks/effects that are already normalized in the current MOOX runtime.

The player-safe Research choice projection now carries `technology_info`, derived server-side from the same rules used by gameplay. Currently derivable effect families include:

- Colony Building unlocks, including PP cost and BC/turn maintenance;
- Planetary Project unlocks and PP cost;
- Ship Drive unlocks and FTL speed;
- Ship Computer, Armor and Shield unlocks;
- Ship Fuel Cell unlocks and range in parsecs;
- normalized Population Growth technology bonuses;
- Advanced City Planning population-capacity bonus.

Examples verified against the committed normalized rules:

- **Biospheres** -> Building unlock, 60 PP, 1 BC/turn maintenance;
- **Fusion Drive** -> Ship Drive unlock, FTL speed 3;
- **Research Laboratory** -> Building unlock, 60 PP, 1 BC/turn maintenance;
- **Deuterium Fuel Cells** -> Fuel Cell unlock, 6 parsecs range.

## Honest incomplete-description behavior

`technologies.json` currently contains stable Technology IDs/keys/names/TechField ownership and provenance, but not a general authoritative description string for every Technology application. Historical `TECHDESC.LBX` extraction research identified short-description arrays, but their semantic mapping is not yet normalized broadly enough to attach those strings blindly to all 203 normalized Technologies.

Therefore the UI deliberately does **not** fabricate prose. If a technology is researchable but its detailed gameplay effect is not currently normalized (for example some Weapon/Special applications), the `?` dialog explicitly says that the technology will be granted authoritatively but its detailed effect is not normalized yet. Future TECHDESC/application normalization can enrich this same dialog without creating another UX.

## Validation

- `go test ./...` passes.
- `go vet ./...` passes.
- `npm run build` passes (`tsc -b && vite build`).
- Browser QA on a fresh Human-vs-Built-in-AI game verified:
  - HUD Research opens the canonical overlay;
  - the `/research` route renders the exact same overlay;
  - every visible Technology exposes a keyboard/touch-capable `?` control;
  - Research Laboratory details display the normalized 60 PP / 1 BC effect;
  - an unnormalized Anti Missile Rockets effect is clearly marked as not normalized rather than guessed;
  - a real Research Breakthrough was driven to completion; **Choose next research** closes the Breakthrough card and opens the same canonical Research overlay;
  - choosing the next Technology closes the overlay and updates the planning preview.

## Scope boundary

This is a Slice 15.4 Gate-3 UX correction/enrichment before Block 6. It changes neither Research resolution semantics nor Slice 15.5 Tactical scope.
## Post-review layout hotfix

Manual review found that the first `?` control styling forced each Technology row to 30 px height. In the managed review viewport this made the Research grid 584 px tall while only 490 px was available, and child rows could extend below the fixed-height Research field panels.

The compact layout now uses 16 px desktop info controls (18 px in the narrow/mobile rule), tighter row/body spacing, and a reduced field-panel minimum height. Browser geometry verification at a 791 x 605 viewport now reports:

- all 8 Research fields inside the visible grid;
- grid `scrollHeight == clientHeight == 466 px`;
- four two-column field rows at 111 px each;
- Technology rows and `?` controls at 16 px height;
- no Technology choice row extends beyond its Research panel;
- the `?` detail dialog remains clickable and opens normally.