# Slice 17.2 Gate 3 - Weapon Slot / Edit HMI Amendment

Date: 2026-09-22
Result: **PASS**

## 1. Purpose

This follow-up closes the remaining Ship Designer weapon-slot HMI work after the compact mobile selector amendment.

The accepted interaction now separates the two jobs cleanly:

- the main Ship Designer shows compact, stable installed-weapon slot cards;
- the Weapon Selector owns weapon identity, mount type, special modifiers and quantity configuration.

Gate 4 is **not** started by this amendment.

## 2. Installed-weapon slot cards

Installed weapons now render as fixed-height cards rather than expanding with configuration prose.

Each occupied slot shows:

- weapon name;
- inline `- quantity +` controls;
- an icon-only Edit action.

The cards deliberately do **not** repeat Standard / Heavy / Point Defense or special-modifier prose. Those details remain available through Edit, which prevents the main Designer from reflowing as weapon configuration grows.

Mobile card height is deterministically **48 px**; desktop card height is **52 px**.

No separate Remove button is rendered. Reducing quantity below one keeps the existing authoritative remove behavior without adding a second destructive control.

A matching fixed-size **Add weapon** slot remains available while another weapon slot is free.

When no weapon is installed, the blank installed-weapons scroll area also remains an Add target, preserving the earlier mobile interaction requirement.

## 3. Edit-in-place flow

Each installed slot has an Edit pencil.

Opening Edit:

- reuses the existing Weapon Selector;
- preselects the installed weapon;
- restores its mount type;
- restores its special modifier IDs;
- restores its quantity;
- keeps the original slot ID.

Applying the edit replaces that slot in the draft instead of adding a new slot.

The authoritative design-preview request excludes the old copy of the edited slot before adding the candidate replacement, so Space/PP validation is evaluated against the edited result rather than a duplicate.

## 4. Mount and modifier help

Standard / Heavy / Point Defense choices and each supported special modifier now have compact `?` affordances.

The inline help explains:

- what the option means;
- Space/PP delta from the frozen Gate-2 contract;
- refinement requirement where applicable;
- the current Tactical support/defer boundary.

The help text does not claim deferred Tactical effects are already simulated.

## 5. Mobile behavior

The previous mobile-selector constraints remain intact:

- no separate Available Weapons card;
- filters start collapsed;
- weapon list remains the primary pre-selection surface;
- configuration stays compact after selection;
- Install / Apply remains reachable without page-level layout growth.

The installed-weapons panel keeps a stable height and scroll behavior while slot cards stay fixed-size.

## 6. Regression coverage

The reference-lab browser smoke now verifies at 390 px:

- blank installed-weapons area opens Add;
- Add slot is present and compact;
- two installed weapon cards remain fixed at 48 px;
- no slot-number labels are exposed;
- no duplicate dedicated Remove control exists;
- two accessible icon-only Edit controls exist;
- Edit opens the existing installed Laser prefilled with Standard and quantity 1;
- Point Defense help exposes its meaning/cost delta;
- Autofire help exposes the deferred Tactical scope;
- Point Defense + Autofire can be applied to the edited slot;
- applying Edit does not duplicate or remove the slot;
- reopening Edit restores Point Defense + Autofire;
- existing selector, mobile, desktop, ordinary-game and reference-lab regressions remain green.

## 7. Final QA

- `npm run build` - **PASS**
- `node --check scripts/check-reference-lab-browser.mjs` - **PASS**
- `npm run check:reference-lab:browser -- http://127.0.0.1:7201` - **PASS**
- `go test ./... -count=1 -timeout 120s` - **PASS**
- `go vet ./...` - **PASS**
- `git diff --check` - required again after documentation before commit

## 8. Status

This amendment completes the recovered Gate-3 weapon-slot/edit unfinished work.

**Gate 3 remains complete. Gate 4 is next and has not started. Slice 17.3+ remain closed.**
