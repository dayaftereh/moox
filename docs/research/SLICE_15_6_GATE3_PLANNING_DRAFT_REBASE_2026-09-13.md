# Slice 15.6 Gate3 - Planning draft rebase across immediate revisions

Status: **IMPLEMENTED / exact Triangle browser regression green**
Date: 2026-09-13

## Trigger
During Triangle round-1 planning the user reassigned population, opened colony construction, edited/saved the Scout design, queued Scout + Colony Base, then used browser Back. Returning to the colony showed the original population split again.

Read-only canonical diagnosis proved Browser Back was only exposing the loss. The planning draft had already lost `population:17` when the Ship Builder saved the design. `game-triangle-2pc` was at revision 3 with Scout design revision 2 / visual revision 2 and a planning draft containing only `construction:17`.

## Root cause
Server planning drafts were bound to an exact `(turn, base_revision)`. `PlayerSnapshot` treated a draft whose `base_revision` differed from the current game revision as stale and removed it from the host.

Ship Builder design save performs immediate authoritative mutations:

1. `empire.save_military_design` advances the game revision;
2. snapshot reload;
3. `empire.set_military_design_visual` advances the game revision again;
4. snapshot reload.

A valid population/research/construction planning draft created before those immediate mutations therefore became revision-stale and was deleted on reload. Browser Back, normal reload, and reconnect then hydrated the colony from the authoritative base state and made the missing order visible.

## Fix
`hostedGame.mutate` now attempts a server-authoritative rebase of stored planning drafts whenever a mutation changes revision while remaining in the same strategic Planning turn.

For each stored draft:

- clone the draft;
- bind `base_revision` to the new authoritative game revision;
- rebuild its normal Planning `CommandBatch`;
- run the normal server `PlanningPreview` validation against the new state;
- only store the rebased draft when validation succeeds.

The server does not blindly carry an invalid draft across a revision change.

`Host.SavePlanningDraft` also accepts a draft whose base revision is older than the current revision when the game is still in the same Planning turn. Before storage it rebases the request to the current revision and performs the normal PlanningPreview validation. This covers the mobile/network race where an async draft save reaches the server after an immediate mutation already advanced the revision.

A draft targeting a future revision, a different turn, a non-Planning phase, or an invalid order is still rejected. An invalid stale save does not overwrite an already valid rebased draft.

## Automated coverage
`internal/app/planning_draft_rebase_test.go` proves:

- Human population draft `5 Farmer / 3 Worker / 0 Scientist` survives immediate Scout gameplay design save;
- it is rebased again across the subsequent visual-design save;
- the resulting draft base revision matches the current game revision;
- the rebased draft still produces a valid PlanningPreview with the intended population;
- a late stale-but-valid draft save is accepted, rebound to the current revision and validated;
- an invalid stale draft is rejected and cannot overwrite the valid rebased draft.

## Exact browser regression
Fresh isolated Triangle on port 7185, using the production React flow:

1. Round 1 / Human Home starts 4 Farmer / 2 Worker / 2 Scientist.
2. Move one Scientist to Farmer, then one Scientist to Worker -> draft projects **5 / 3 / 0**.
3. Open colony -> `Bau verwalten` -> select Scout -> `Schiff entwerfen`.
4. Install Laser Cannon and press `Design speichern`.
5. Ship Builder executes gameplay + visual saves; game reaches revision 3, Scout design r2 / visual r2.
6. Server snapshot still contains `population:17`, now rebased to base revision 3.
7. Browser History Back to construction.
8. Queue Scout revision 2 and Colony Base.
9. Server planning draft contains both `population:17` and `construction:17`.
10. Browser History Back to colony detail: visible population remains **5 Farmer / 3 Worker / 0 Scientist**, queue shows Military Ship + Colony Base.
11. Full page reload: same population and queue remain.
12. Close browser and reconnect from a fresh browser profile: same population and queue remain.

This fixes the reported data-loss path without introducing browser-history-specific state handling; navigation is now safe because the planning state remains authoritative on the server.
