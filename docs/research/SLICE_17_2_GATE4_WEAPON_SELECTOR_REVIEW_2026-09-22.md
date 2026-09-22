# Slice 17.2 Gate 4 - Weapon Selector review block

Date: 2026-09-22
Status: **automated QA passed; manual visual review pending before Slice 17.2 close**

Parent slice: `docs/slices/PLANNED_17_2_BEAM_WEAPONS_MOUNT_MODIFIERS.md`
Open marker: `docs/slices/_OPEN_BEAM_WEAPONS_MOUNT_MODIFIERS_2026-09-20.md`

## Review trigger

Mobile review of the Weapon Selector showed three usability problems:

- mount/modifier cards were visually noisy because the card selection state was duplicated by native radio/checkbox glyphs;
- compact `?` help controls could collide with long localized labels;
- the damage badge looked like a live derived stat even though Slice 17.2 freezes only base Beam min/max damage while Heavy/Point Defense Tactical damage/range behavior remains deferred;
- quantity +/- inside the selector duplicated a control that is more naturally owned by the installed weapon card.

## Implemented review amendment

The selector now uses the whole option card as the interactive selection surface:

- selected mount/modifier cards carry the blue selected border/background;
- native radio/checkbox glyphs are removed from the visible UI;
- each choice exposes `aria-pressed` and a stable data selector for deterministic browser tests;
- help buttons occupy a dedicated top-right area with reserved text padding;
- the weapon damage value is explicitly labelled **Base damage / Basisschaden**;
- the existing server-authoritative design preview remains the source of Space/PP totals;
- new installs use count 1; selector editing preserves the installed row count without exposing quantity controls;
- all quantity changes happen on the installed weapon card with its existing - / count / + controls.

This amendment does **not** invent Tactical damage multipliers. The frozen Gate-2 contract still defers Heavy Mount Tactical range/damage and Point Defense Tactical hit/range behavior. Heavy and Point Defense therefore keep the Beam's base min/max damage metadata while their already-frozen Space/PP factors remain authoritative. The quantity-placement change supersedes only the Gate-2 HMI amendment's earlier selector-local quantity control; the canonical weapon-row count and all gameplay arithmetic remain unchanged.

## Regression coverage

The reference-lab browser smoke now proves at 390 px:

- no visible option-card radio/checkbox glyphs;
- help-button geometry does not overlap the option label;
- selected mount/modifier state is carried by the card and `aria-pressed`;
- Heavy changes the authoritative Space/PP preview relative to Standard;
- Point Defense changes the authoritative Space/PP preview relative to Standard;
- the unchanged damage value is labelled Basisschaden rather than presented as a derived mount-adjusted value;
- the selector exposes no quantity +/- control on add or edit;
- installed weapon cards retain their quantity controls; the browser regression raises a row from 1 to 2 before Edit and proves Point Defense + Autofire editing preserves count 2;
- edit/reopen persists Point Defense + Autofire selection state.

## QA

Fresh temporary review server: `127.0.0.1:7201`.

Passed:

- `node --check web/scripts/check-reference-lab-browser.mjs`;
- `node web/scripts/check-reference-lab-browser.mjs http://127.0.0.1:7201`;
- `npm --prefix web run build`;
- `go test ./... -count=1 -timeout 120s`;
- `go vet ./...`;
- `git diff --check` (line-ending warning only, no whitespace error).

## Close boundary

Keep Slice 17.2 open for user visual review on the canonical port-7171 server. Do not open Slice 17.3 automatically. After visual acceptance, Gate 4 may perform the final closeout commit/evidence update and remove the Slice-17.2 `_OPEN_` marker.
