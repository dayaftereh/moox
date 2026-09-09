# Slice 15.4 Gate 3 Block 5 - Resolution summary, Research breakthrough and result UX

Date: 2026-09-09
Status: **complete**

## Objective

Close the Gate-2 transition-summary gap without moving gameplay authority into React. Automatic Research completion, external Technology grants, Battle/Invasion outcomes, Empire elimination and game completion must be presentable from a compact player-safe authoritative read surface with stable identities. Research acknowledgement remains presentation-only; completed games become read-only and receive a dedicated result surface.

## Server read contract

`PlayerView.recent_resolutions` is now derived from the persisted authoritative session event log. It does not mutate simulation state and does not add a new live-save schema field.

Each summary carries:

- stable ID `event-<event_sequence>` plus event sequence, turn and revision;
- one typed kind: `research_breakthrough`, `technology_granted`, `battle_completed`, `invasion_resolved`, `invasion_declined`, `empire_eliminated` or `game_completed`;
- only the compact data required to present that resolution.

Privacy/authority rules:

- Research completion and Technology grants are projected only to the owning Empire;
- Battle completion is projected only to participating seats and includes the already-visible Battle Spec/result context, destroyed ships and derived surviving visible ships;
- Invasion result/decline is projected only to the attacker/defender Empires;
- Empire elimination and game completion are public major-result summaries;
- routine historical event-log contents remain private; only the typed projection crosses the player boundary;
- the recent window is bounded to 12 summaries and current/previous strategic turns, while `game_completed` remains available for a completed game.

The stable summary ID is the persisted event sequence, so reload/reconnect does not require React to diff whole snapshots or invent transition identity.

## Research breakthrough UX

The browser now detects the newest unacknowledged authoritative `research_breakthrough` summary and presents a blocking transition card:

- Research field completion;
- unlocked Technology keys/IDs when available;
- Hyper-Advanced research level when present;
- **Choose next research** routes to the existing Research UI;
- **Continue** only acknowledges presentation.

Acknowledgement is stored per game/seat in browser `sessionStorage`, capped to the most recent 32 IDs. No acknowledgement endpoint or fake `Complete Research` gameplay command was introduced.

## Victory/result UX

A completed game now uses a dedicated result surface instead of the previous small inline card:

- public winner identity/name with numeric fallback;
- conquest title and completed turn;
- eliminated Empire names with numeric fallback;
- explicit authoritative/read-only message;
- Main Menu action.

Normal strategic section rendering is suppressed while `phase === completed`, so the finished game no longer exposes ordinary planning/mutation surfaces underneath the result card.

## Regression evidence

Server/session:

- added `resolution_summary_test.go` privacy/stability coverage;
- owning Research/Technology data is visible only to its owner;
- non-participant Battle summaries are excluded;
- involved players receive Battle/Invasion summaries;
- repeated snapshots retain identical summary IDs;
- stale old summaries fall outside the bounded recent window;
- `go test ./internal/session` passes;
- `go test ./...` passes;
- `go vet ./...` passes.

Web:

- TypeScript contracts mirror the new player read projection;
- `npm run build` (`tsc -b && vite build`) passes;
- `git diff --check` passes (only the repository's configured LF/CRLF conversion warnings are emitted).

## Files

- `internal/session/resolution_summary.go`
- `internal/session/resolution_summary_test.go`
- `internal/session/session.go`
- `web/src/api.ts`
- `web/src/App.tsx`
- `web/src/i18n.tsx`
- `web/src/styles.css`

## Scope guard / next block

Block 5 deliberately does not implement the interactive Tactical battlefield. Slice 15.4 Gate 3 Block 6 remains the Battle Entry/Return route and shell using the frozen handoff contract; interactive Tactical mechanics remain Slice 15.5.
