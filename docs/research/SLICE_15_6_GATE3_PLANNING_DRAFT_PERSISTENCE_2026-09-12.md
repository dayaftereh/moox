# Slice 15.6 Gate3 - Planning Draft Persistence

Status: **IMPLEMENTED / browser-QA green**
Date: 2026-09-12

## Problem
Strategic planning changes were held only in React `draftOrders`. Construction queues, colony population assignments, research and fleet movement looked correct through the non-mutating planning preview, but a browser reload called `loadSnapshot`, cleared `draftOrders`, and lost every unsubmitted change. Only strategic `Fertig` sent the batch to `/turn-submissions`.

Ship-design save is intentionally different and remains immediate/authoritative. Diplomacy war declaration also remains immediate and is not silently moved into the strategic planning draft contract.

## Frozen authority contract
- A planning draft is server-hosted per game + seat + strategic turn/base revision.
- Draft entries retain the React/UI replacement key plus command kind/payload, so reload can reconstruct exact `draftOrders` semantics.
- Each client mutation carries a monotonic `draft_revision`. The host ignores stale/out-of-order saves, including an older save arriving after a newer clear/tombstone.
- Saving a draft validates the derived command batch through the existing authoritative `PlanningPreview` path before it is accepted.
- Draft save does **not** mutate game state, game revision or global change sequence.
- Player snapshot projects only that seat's current valid planning draft. A draft tied to an obsolete game turn/revision is discarded instead of being replayed against new authority.
- Successful `SubmitTurn` atomically clears that seat's hosted planning draft.
- Empty drafts are retained as revisioned tombstones until turn submission/revision invalidation so late older requests cannot resurrect removed orders.

## HTTP/UI integration
New route:

`PUT /api/v1/games/{gameID}/seats/{seatID}/planning-draft`

The existing player snapshot now optionally includes `planning_draft`. React immediately saves each draft replacement/removal using `keepalive`, increments `draft_revision`, and rehydrates from `planning_draft` on browser load/reconnect. Local newer in-flight state is preserved against an older same-authority snapshot, while a turn/revision change resets stale local draft state.

Because this persists the generic `DraftOrder[]`, the mechanism covers every existing planning action using that path without category-specific storage code. Current browser-visible coverage includes:
- `colony.assign_population` - Farmer/Worker/Scientist redistribution;
- `colony.set_construction_queue` - colony build queues / military ship queue;
- `empire.select_research` - research selection;
- `empire.move_fleet` - fleet/ship movement drafts.

Ship designs remain immediate authoritative commands. Diplomacy/war declaration remains outside this contract pending an explicit diplomacy planning decision.

## Regression tests
Server HTTP regression `TestHTTPPlanningDraftSurvivesReloadOrderingAndSubmission` proves:
1. initial snapshot has no draft;
2. a valid population draft can be PUT and is returned on a later GET player snapshot;
3. a newer empty draft/tombstone survives;
4. an older delayed revision cannot overwrite that tombstone;
5. a later valid revision can be stored;
6. successful turn submission clears the hosted draft.

## Browser QA - Triangle Round 1 on isolated 7176
Starting from clean `game-triangle-2pc` Round1/Planning:

1. Moved one Farmer to Worker: 4/2/2 -> **3/3/2**. Server snapshot immediately contained draft revision 1 with `population:17`.
2. Added Scout design r1 to colony #17 construction: server draft revision 2 contained population + `construction:17`. Planning preview showed Scout, 25 PP, ETA 3.
3. Hard browser navigation/reload to the same build route: UI rehydrated **2 planned actions**, Food preview remained -2.0, colony detail remained 3/3/2 and Scout remained current planned construction.
4. Selected Anti Missile Rockets research: server draft revision 3 added `research`; top Research preview changed to 0% / 9T.
5. Opened the Human Home fleet picker with all three ships selected and targeted the legal unknown system. Existing fleet-splitting logic produced two `empire.move_fleet` draft entries, bringing the generic draft to **5 orders** total.
6. Hard browser reload on Galaxy: snapshot still returned draft revision 5 with population, construction, research and both fleet move orders. UI displayed `5 geplante Aktion(en)`, Food -2.0 and research 0% / 9T.
7. Pressed strategic `Fertig`: authoritative resolution advanced to Round2/Planning revision 3 and player snapshot returned `planning_draft: null`.

This directly reproduces and closes the reported browser-reload bug for the requested planning categories.

## Persistence boundary
This block makes drafts authoritative within the running hosted game and reload/reconnect safe. It does not yet embed planning drafts inside the separate live-snapshot/save-file serialization format; a full server process restart or explicit live-snapshot restore therefore starts from the serialized game state, not an unsubmitted hosted draft. If we want crash/restart durability for unsubmitted drafts, that should be a separate persistence-format versioning step rather than silently changing the established save schema here.
