# Slice 15.4 Gate 3 - Building detail descriptions and runtime effects

Date: 2026-09-09

## Goal

Make the existing colony construction detail panel explain what a selected building costs, what it currently does in MOOX, and what the original MOO2 building/technology description says, without adding another `?` dialog.

## Projection contract

`ConstructionChoice` now exposes optional building-only metadata:

- `original_description`: source-backed MOO2 HELP text already normalized through the linked Technology;
- `effects[]`: zero or more structured effects that the current MOOX runtime actually models.

Production cost and maintenance remain the existing authoritative dynamic fields on the same `ConstructionChoice`.

The effect array intentionally supports more than one entry. UI rendering is list-based, so later normalized building mechanics can add another effect without redesigning the panel.

## Current normalized building effects

The server projects only mechanics already represented in MOOX rules:

- Cloning Center: flat population-growth contribution;
- Biospheres: flat population-capacity contribution;
- Holo Simulator / Pleasure Dome: morale percentage;
- Marine Barracks / Armor Barracks: government-dependent barracks morale-penalty relief;
- Star Base / Battlestation / Star Fortress: command points.

Buildings with no separate normalized runtime effect are explicitly labelled as such. Their MOO2 description is still shown below, but is not presented as implemented MOOX behavior.

## Source descriptions

All 48 normalized colony buildings link to a concrete Technology. The Technology normalization added in commit `2da3440` provides a non-empty source-backed HELP.LBX description for every one of those linked Technologies, so no second hard-coded building-description table is required.

The UI label is `Originalbeschreibung aus MOO2 (Englisch)` to make the provenance and language clear.

## Browser QA

Isolated QA game `qa-building-info` on port 7181 verified:

- Capitol: 200 PP, no separate normalized runtime effect, full original description shown;
- Marine Barracks: 60 PP, 1 BC maintenance, normalized morale-penalty relief shown, original description shown;
- Star Base: 400 PP, 2 BC maintenance, `+1 Kommandopunkte`, original description shown;
- long descriptions remain readable in the normal page-scroll flow;
- existing `Einplanen` queue action remains unchanged.

## Verification

- `go test ./...` passed after the feature implementation;
- `go vet ./...` passed;
- `npm run build` passed;
- `git diff --check` passed;
- targeted `go test ./internal/game` passed after the explicit multi-effect-array regression test was added.
