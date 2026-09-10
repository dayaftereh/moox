# Slice 15.6 Gate 3 - Block 2 Ship Designer shell

Date: **2026-09-10**

Status: **implemented and focused browser verified**.

## Goal

Implement the frozen durable Ship Designer interaction shell without broadening the currently authoritative gameplay-design scope beyond Frigate with no weapon or exactly one Laser Cannon.

## Browser shell

The previous visual-only hull tile prototype is replaced by a server-driven designer:

- persistent named Ship Design library with procedural thumbnails and design revision;
- `+ New design` flow;
- player-editable design name;
- left/right hull stepper in server-projected order Frigate -> Destroyer -> Cruiser -> Battleship -> Titan -> Doom Star;
- server-projected lock state and Command Point / base-space context for each hull;
- large clickable/tappable procedural ship preview; reroll changes only the local visual candidate until Save;
- Available weapons/components list;
- separate Installed components list with stable weapon-slot presentation;
- authoritative mandatory drive/computer/armor/shield/fuel values from the selected server preview;
- authoritative Production Cost, Command Points and used/available design space;
- Save flow for a new or revised named design;
- visible note that additional weapon mounts/counts/modifiers remain future server-authoritative breadth.

The existing `Scout` is presented correctly as a saved Frigate design rather than as a separate hull class.

## Save lifecycle

Gameplay and visual identity remain separate authoritative mutations.

The browser now performs:

1. `empire.save_military_design` against the current authoritative revision;
2. authoritative snapshot refetch;
3. identify the created/updated design ID;
4. `empire.set_military_design_visual` against that fresh revision;
5. final authoritative refetch.

A 409 causes authoritative refetch and is never auto-resubmitted. Design saving is disabled outside open Planning or after the player's turn submission.

## Immediate-command routing bug found during browser QA

The first real browser save exposed an existing integration hole: `Host.SubmitImmediateCommand` did not classify `empire.save_military_design`, so it fell through to the Colony Base resolver and rejected Planning with `cannot resolve Colony Base in phase "planning"`.

The fix adds an explicit military-design immediate-command classifier and session resolver. The gameplay design mutation now:

- requires Planning;
- requires exact base revision;
- rejects after the first turn submission;
- checks seat/elimination authority;
- clones state;
- resolves through the existing authoritative military-design economy rule;
- validates and commits a normal strategic domain event/revision.

A dedicated HTTP regression test protects the exact route and stale-revision contract.

## Projection hardening

A legacy visual-only fixture intentionally has no complete mandatory military-system technologies. The new Designer projection initially caused that fixture's entire DecisionView to fail while trying to build a preview.

The catalog now degrades safely: the relevant hull becomes `mandatory_systems_required` / unavailable and no illegal preview is projected, while the rest of the player snapshot remains usable.

## Focused real-browser QA

Isolated server: **127.0.0.1:7181**. Canonical 7171 was not modified.

Fresh test game:

- game `designer-b2`;
- seed `0x8009`;
- Human local human vs Darlok Built-in AI;
- Turn 1 / Planning / revision 1.

Verified in managed Chrome 152:

- existing `Scout` appears as `Fregatte · r1` in the design library;
- Frigate shows `1 CP · 25 Space`;
- stepping right shows Destroyer with `2 CP · 60 Space`, visibly locked as not yet supported;
- Save is disabled on the locked Destroyer;
- clicking the ship preview changed the generated SVG path while preserving the hull/gameplay values;
- no horizontal page overflow at the actual 776 CSS-px managed-browser viewport;
- New Design -> name `Falcon Mk I` -> Add Laser produced:
  - Production: **30 PP**;
  - Command Points: **1 CP**;
  - design space: **10 / 25**;
  - installed row: `Slot 1 / 1x Laser Cannon`;
- Save completed through the real browser and returned `Design saved · revision 1 · visual 1`;
- the design library immediately contained both `Scout` and `Falcon Mk I`.

Authoritative post-save snapshot:

- game revision: **3**;
- `Falcon Mk I` design ID: **80**;
- gameplay revision: **1**;
- visual revision: **1**;
- hull: `frigate`;
- weapon: `laser_cannon`;
- Production Cost: **30 PP**;
- Space Used: **10**;
- matching Colony Military Ship construction choice count: **1**;
- construction choice name: `Falcon Mk I`;
- construction choice revision: **1**.

This proves the persisted-design-to-Colony-choice data handoff. The next block will exercise that handoff through the Colony browser UI and queue the exact design.

## Regression

- focused `internal/game`, `internal/session`, `internal/app`, `internal/server` military designer/design tests: PASS;
- dedicated HTTP immediate Military Design create + stale revision regression: PASS;
- full `go test ./...`: PASS;
- full `go vet ./...`: PASS;
- `npm run build`: PASS;
- `git diff --check`: PASS;
- production main JS after this block: about **255.23 kB** (well below frozen 350 KiB review guardrail);
- temporary 7181 server closed after QA.

## Next

Gate 3 Block 3: verify the named saved design through the real Colony Construction UI, queue the exact design/revision through visible browser interaction, and confirm the construction state preserves the design identity. Keep this as a separate recovered session block.
