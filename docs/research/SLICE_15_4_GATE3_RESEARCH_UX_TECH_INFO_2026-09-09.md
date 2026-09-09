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

## Variable Technology-count layout hardening

A follow-up review checked the actual Technology cardinality instead of only the initial screen geometry. TechField 4 is authoritatively a three-Technology `choose_one` field: Anti Missile Rockets, Reinforced Hull and Fighter Bays. Ruleset, server projection and rendered DOM all agree on exactly three entries.

The full normalized ruleset also contains later TechFields with four and up to eight Technology applications. Therefore Research field cards must not rely on a fixed content height. The field panel now uses `height: max-content` with the compact minimum height retained. This preserves the compact initial eight-field desktop layout while allowing a card to grow whenever its Technology list needs more vertical space; the outer Research grid handles scrolling.

Validation:

- current desktop review: all 8 frontier fields visible, no clipped rows, grid `scrollHeight == clientHeight == 466 px`;
- TechField 4: 3/3 technologies visible in DOM, no clipping;
- narrow/mobile-style probe: 3-entry cards grow to 117 px automatically and no row clips;
- synthetic 8-entry field probe: card grows to 196 px, `clipped=false`, outer grid becomes scrollable instead of hiding entries.
## Source-backed Technology descriptions

Follow-up manual review showed that technologies without a currently normalized MOOX gameplay effect still fell back to a generic explanation. The original MOO2 data already contains a stronger authoritative source: `HELP.LBX` block 0 is a fixed array of 707 records x 1403 bytes, and records 1..203 align directly with the 203 concrete Technology IDs normalized from `TECHNAME.LBX`.

The normalized Technology ruleset now carries the original explanatory text and field provenance for every Technology:

- `technologies.json` schema version 5;
- `Technology.description` plus `description_source`;
- source `moo2-1.31-help-block0-technology-descriptions` from `HELP.LBX` block 0;
- 203/203 Technologies have a non-empty description with provenance;
- the technology normalizer reads and validates `HELP.LBX` together with `TECHNAME.LBX` and `Orion2.exe`;
- punctuation-only name differences are normalized for validation, while five known source naming/spelling variants remain explicitly accepted.

The Research `?` dialog deliberately separates two authority levels:

1. **What you get / Was du bekommst**: concrete effects that the current MOOX runtime has normalized, such as building unlock cost/maintenance, drive speed, fuel range, armor/shield/computer unlocks, or modeled population bonuses.
2. **Original MOO2 description**: the source-backed English HELP.LBX explanation of the intended technology behavior. This may describe mechanics that MOOX has not implemented yet, so it is never presented as current runtime truth.

Browser QA verified both sides of that contract. A previously generic-only example such as Augmented Engines now exposes the original +5 combat-speed behavior after the runtime-normalization notice. A normalized example such as Research Laboratory keeps its current MOOX building-unlock metadata and additionally shows the source explanation of the laboratory's research behavior. The dialog remains within the managed viewport and its content area retains scrolling for longer descriptions.
