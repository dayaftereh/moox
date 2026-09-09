# Slice 15.4 Gate 1 - Rich-state UX / persistence audit

Date: **2026-09-09**

Status: **complete; Gate-2 contract ready for review**.

## Goal

Turn the already-authoritative strategic game into a coherent browser experience at every human interaction boundary outside the actual Tactical battlefield. This audit describes the **existing server truth**, the **current browser exposure**, and the minimum additions needed for 15.4 without moving gameplay authority into React.

## 1. Authoritative phase inventory

The strategic session has exactly six phases:

| Phase | Server behavior | Human boundary today | 15.4 responsibility |
| --- | --- | --- | --- |
| `planning` | accepts strategic batches and allowed planning immediates | already functional in 15.2 | retain; receive transition summaries from prior turn |
| `strategic_resolution` | automatic deterministic resolution | none | loading/progress presentation only; never expose fake commands |
| `encounters` | one or more BattleSession children; human participants stop host automation | **yes** | rich battle-entry/context/return shell; battlefield belongs to 15.5 |
| `invasion_decisions` | attacker must invade or decline | **yes** | replace numeric priority card with contextual blocking decision |
| `post_resolution` | auto-completes due research, then waits for human Colony Base resolutions; completes turn when clear | **yes for Colony Base** | breakthrough feedback plus blocking Colony Base workflow |
| `completed` | immutable completed game/result | terminal | dedicated result/victory presentation |

The host's `driveToInteractiveBoundary` already enforces this contract: it advances automatic phases and built-in AI, then returns when a non-AI human decision is pending. 15.4 should visualize those boundaries, not duplicate the phase driver.

## 2. Current DecisionView projection

Go `PlayerDecisionView.Decisions` already projects:

- Research categories and choices;
- Construction and Population choices;
- Population transfer choices;
- Fleet moves;
- Colonization;
- Outpost deployment;
- Diplomacy;
- **Colony Base resolutions**;
- **Invasion opportunity**;
- **Battle decisions/actions**.

### Browser type gap

The current TypeScript `DecisionCatalog` exposes invasion but does **not** type `colony_base`; `battles` is only `unknown[]`. This is a client-contract gap, not a missing server rule.

The current `BattleView` TypeScript type also truncates the existing server Battle Spec. The Go Battle Spec already contains player-useful context:

- battle ID;
- strategic turn;
- system ID;
- attacker/defender side;
- empire/seat IDs;
- combat fleet IDs;
- ship IDs;
- civilian fleet IDs;
- defender colony IDs;
- participants;
- Tactical spec or explicit unsupported reason;
- battle result when completed.

## 3. Encounter / Tactical handoff

### Existing truth

When human combat is present, the host stops in `encounters`. The browser currently only shows a small battle list under Fleets. It has **no battle route, no encounter brief, no typed BattleDecision and no battle-command API wrapper**.

### 15.4 boundary

15.4 owns:

1. detect the human-participant BattleView;
2. present the system, sides, fleets/ships and colony stakes;
3. route into a stable Battle screen shell;
4. expose explicit unsupported/Tactical-not-yet-available state when applicable;
5. preserve a stable route/context contract for 15.5;
6. after battle completion, show a strategic return summary before normal navigation resumes.

15.4 does **not** own movement, weapon legality, target selection, initiative or battlefield commands. Those remain 15.5.

Recommended route boundary:

`#/game/<gameID>/battle/<battleID>`

15.4 can implement the route and Battle Entry/Result shell. 15.5 mounts the interactive battlefield into that contract.

## 4. Invasion workflow

### Existing truth

`InvasionOpportunity` contains:

- system ID;
- colony ID;
- attacker/defender empire IDs;
- attacker seat ID;
- eligible transport fleet IDs.

The browser already has a functional priority card and `submitInvasion`, which submits all eligible transports for Invade or the colony ID for Decline.

### UX gap

The current card mostly displays raw numeric IDs. 15.4 should resolve safe public names/visual context and make this a blocking decision with:

- target system/planet/colony;
- defender public identity;
- available transport count/fleets;
- clear Invade / Decline actions;
- busy/error/retry state;
- no hidden combat formula invented in React.

## 5. Colony Base workflow

### Existing truth

Post-resolution can block on `ColonyBaseResolution`:

- empire ID;
- source colony ID;
- system ID;
- legal target planet IDs;
- trash/refund BC.

Server commands already exist through the generic immediate-command endpoint:

- colonize target planet with the Colony Base;
- trash the Colony Base for the projected refund.

Built-in AI already resolves these through the same DecisionView.

### Browser gap

There is currently **no Colony Base UX at all** and no TypeScript decision type/wrapper despite the server projection. This is a direct 15.4 implementation target.

Required blocking UI:

- source colony context;
- target-planet cards using existing System/planet visuals;
- one legal target selection;
- explicit Scrap/Trash option showing authoritative refund;
- command busy/rejection state;
- no ability to end/advance while pending.

## 6. Research breakthrough workflow

### Existing truth

During `post_resolution`, the host automatically calls `CompleteResearchField` for due empires before checking Colony Base decisions. The domain event is `empire.research_completed`. The next turn then returns to Planning, where the existing Research UI can select the next field/technology according to projected legal choices.

### UX gap

Because completion is automatic and the browser snapshot currently has no player-safe recent-resolution event summary, a breakthrough can disappear between invalidation/refetch and the next Planning snapshot.

### Required read addition

Add a small player-safe **resolution summary / recent transition events** projection keyed by authoritative event/change sequence. It should expose only presentation-safe facts such as:

- own research field completed;
- technologies granted;
- completed construction/buildings where useful;
- battle outcome visible to the participant;
- invasion/Colony Base result visible to the player;
- victory/result transition.

This is preferable to letting React diff arbitrary full snapshots and infer rules.

## 7. Victory/result workflow

The browser already renders a minimal result card with winner IDs, completed turn/revision and eliminated IDs. 15.4 should make the completed phase a dedicated terminal presentation.

Likely small player-safe enrichment:

- public empire names/race identity for winner/participants instead of ID-only presentation.

No further gameplay command is allowed after completion. Read-only inspection may remain available if desired.

## 8. Public identity requirement

Several rich decisions currently expose legal IDs but not convenient player-facing names for other empires. 15.4 should add/reuse a player-safe public identity directory containing only already-public information:

- empire ID;
- empire name;
- race ID / display identity.

This prevents Invasion/Encounter/Result screens from saying only "Empire 2" while preserving hidden economy/research state.

## 9. Persistence server truth

Persistence is already implemented and atomic behind the persistence-enabled server:

- `GET /api/v1/games/{gameID}/live-snapshot` - export;
- `POST /api/v1/games/import` - import as hosted game;
- `PUT /api/v1/games/{gameID}/live-snapshot` - restore/replace matching hosted game.

Restore:

- validates save/ruleset/game ID before mutation;
- atomically replaces the hosted session only after validation;
- increments change sequence;
- publishes `snapshot_invalidated` with reason `game_loaded`.

The current browser has **no persistence API wrappers and no save/load UI**.

## 10. Current WebSocket / reconnect behavior

The browser currently:

1. loads a snapshot;
2. opens the game WebSocket;
3. on every notification, refetches if `change_sequence` is newer;
4. on close, marks disconnected and reconnects after 1 second;
5. calls a snapshot load before each reconnect;
6. on a new snapshot, clears draft orders and planning preview;
7. guards against an old request overwriting a newly selected game/seat.

This is a solid technical baseline, but the user-facing lifecycle is too generic.

## 11. Explicit persistence/reconnect UX states

Freeze these conceptual states for Gate 2:

- **initial-loading** - no usable snapshot yet;
- **synced/live** - current snapshot + WebSocket connected;
- **refreshing** - invalidation received, current snapshot still visible but mutation temporarily gated;
- **offline/reconnecting** - last snapshot visible read-only, automatic retry;
- **refresh-failed** - stale read-only state plus Retry;
- **exporting/saving** - snapshot download in progress;
- **loading-file** - local save selected/validated;
- **restoring** - PUT existing game; mutating UI locked;
- **importing** - POST new hosted game;
- **loaded** - `game_loaded` received/refetched; local drafts/previews discarded explicitly;
- **fatal** - no safe snapshot or unrecoverable API state.

Mutation while disconnected/refreshing/restoring should be disabled rather than relying on predictable 409 rejection.

## 12. Error/conflict mapping

Server error codes already distinguish important cases:

| Code | HTTP | UX behavior |
| --- | ---: | --- |
| `session_rejected` | 409 | show authoritative reason, refetch; keep only still-compatible unsent local intent |
| `invalid_save` | 400 | reject file; no state change |
| `save_too_large` | 413 | reject file with size message |
| `ruleset_mismatch` | 409 | block restore/import; explain incompatible ruleset |
| `game_exists` | 409 | never overwrite implicitly; offer restore/cancel when appropriate |
| `game_id_mismatch` | 409 | block restore to wrong game |
| `persistence_unsupported` | 409 | explain server/session cannot safely restore |
| `not_found` | 404 | return to game selection / refresh game list |
| `internal_error` | 500 | non-destructive failure, Retry/manual recovery |

The current client collapses error code + message into a generic `Error` string. Gate 3 should introduce a typed client `APIError` preserving status/code/message so UX can branch safely without parsing strings.

## 13. Animation-safe authority boundary

Any rich transition/motion must be presentation-only:

- state transitions key off server `change_sequence`, game `revision`, battle revision/event sequence or explicit resolution-summary IDs;
- animation completion must never submit/advance gameplay implicitly;
- mutating controls always bind to the newest accepted snapshot revision;
- stale transition overlays may finish visually, but commands must validate against current state;
- `game_loaded` cancels queued visual transitions/drafts that belong to the replaced state.

## 14. Gate-1 conclusion

15.4 is mostly an **integration/presentation slice**, not a new-rules slice. The highest-value missing pieces are:

1. Battle Entry/Return shell and stable route for 15.5;
2. full Colony Base blocking decision;
3. richer Invasion presentation;
4. player-safe resolution summary for Research/result feedback;
5. Save/Load/Restore browser flow;
6. explicit reconnect/stale/conflict UX;
7. typed client contracts for currently projected Colony Base/Battle state and API errors.
