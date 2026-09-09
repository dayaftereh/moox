# Slice 15.4 Gate 3 - Block 3 Colony Base blocking decision UX

Date: **2026-09-09**

Status: **implemented and end-to-end verified**.

## Problem closed

The server has always treated a completed Colony Base as a mandatory Post-Resolution human decision. `GameSession.CompleteTurn()` refuses to advance while any `ColonyBaseResolution` remains pending. Before this block the browser had no surface for that decision, so a human game could stall even though the server state and legal commands were already authoritative.

## Existing authoritative contract reused

`ColonyBaseResolution` supplies:

- empire ID;
- source Colony ID;
- system ID;
- legal target planet IDs;
- authoritative scrap refund BC.

Existing commands are reused unchanged:

- `colony.colonize_with_base` with source Colony ID + selected projected target planet ID;
- `colony.trash_colony_base` with source Colony ID.

No colonization legality, refund math or target inference was added to React.

## Client implementation

`web/src/api.ts` adds `submitColonyBase(...)` through the existing immediate-command endpoint and current snapshot revision.

`web/src/App.tsx` now:

- reads the first pending `decision.decisions.colony_base` entry;
- resolves the source Colony, source planet and system from the already player-safe strategic projection;
- filters displayed target cards strictly by projected `target_planet_ids`;
- renders the existing `OrbitalBodyArt` for each legal target;
- displays planet climate/size/mineral context;
- shows the exact projected `trash_refund_bc`;
- blocks on a full decision overlay during the pending state;
- submits Colonize or Scrap only while lifecycle is `synced`;
- refetches after success;
- on structured HTTP 409 rejection, surfaces the authoritative error and requests a fresh snapshot.

Multiple pending Colony Bases are naturally serialized: after resolving the first, the fresh snapshot exposes the next one.

## Blocking presentation

The dialog contains:

- Post-Resolution decision eyebrow;
- Colony Base ready title;
- source planet + system context;
- legal planet cards;
- explicit no-target state when the only legal outcome is Scrap;
- separate Scrap section with authoritative BC refund.

The strategic view remains behind the modal but cannot progress the turn. End Turn is disabled outside Planning as before.

## Real server fixture

For E2E QA, a temporary Go test fixture used the existing authoritative session test helpers rather than manually editing JSON:

1. create the standard two-seat session fixture;
2. grant the known Colony Base technology;
3. add one same-system Terran target planet;
4. queue and complete a real Colony Base through strategic resolution;
5. assert Post Resolution and exactly one pending legal resolution;
6. marshal the resulting live snapshot with the production rules identity.

Fixture facts:

- game ID: `browser-colony-base`;
- source Colony: 10 / Alpha I;
- target Planet: 12 / Alpha II;
- phase: `post_resolution`;
- revision: 2;
- projected refund: 100 BC.

The temporary fixture source file was removed immediately after generating the snapshot and was never committed.

## Isolated browser E2E

The valid snapshot was imported into a temporary persistence server on **127.0.0.1:7192**. Canonical 7171 was not modified.

### Initial blocking state

Browser showed:

- lifecycle: `synced`;
- phase: Nachbereitung / Post Resolution;
- End Turn disabled;
- source: Alpha I;
- system: Alpha;
- exactly one legal target: Alpha II;
- target facts: Terran / Medium / Abundant;
- Scrap refund: 100 BC;
- no danger notice;
- no document horizontal overflow.

### Colonize path

Clicking Alpha II through the real browser control:

- submitted `colony.colonize_with_base`;
- blocking dialog disappeared;
- session advanced to **Turn 2 / Planning**;
- new Colony materialized on Planet 12;
- human Colony count became 2;
- End Turn was enabled again;
- lifecycle returned/remained `synced`.

Server verification after the click:

- phase: planning;
- turn: 2;
- revision: 4;
- Colonies include source Colony 10 and new Colony 13 on Planet 12.

### Scrap path

The same original valid fixture was atomically restored on the isolated server and the browser received the normal invalidation/refetch.

Clicking Scrap:

- submitted `colony.trash_colony_base`;
- blocking dialog disappeared;
- session advanced to **Turn 2 / Planning**;
- no new Colony was created;
- Colony Base was removed;
- Treasury became **106 BC**, exactly +100 BC from the projected refund;
- End Turn was enabled again;
- lifecycle stayed `synced`.

## 320px QA

Viewport: **320x646**.

- blocking sheet width: exactly 320px;
- document scroll width: exactly 320px;
- no horizontal overflow;
- legal target card: 298x68px;
- Scrap button after QA correction: 298x44px;
- dialog height: about 315px for this single-target fixture;
- no danger notice.

The same 44px minimum touch-height correction was also applied to persistence confirmation buttons so Gate-2 touch rules remain consistent.

## Regression

- `go test ./internal/session -run ColonyBase -count=1` PASS;
- `go test ./internal/app ./internal/server -count=1` PASS;
- TypeScript project build PASS;
- Vite production build PASS;
- `git diff --check` PASS;
- temporary 7192 server closed;
- canonical 7171 health remained HTTP 200.
