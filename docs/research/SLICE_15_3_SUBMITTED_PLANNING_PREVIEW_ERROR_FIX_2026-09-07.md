# Slice 15.3 - Submitted-seat planning preview internal-error fix

Date: **2026-09-07**

Status: **diagnosed, fixed, regression-covered**.

## Symptom

The live `game-1` displayed a persistent UI notice:

`internal_error: internal server error`

The Shipbuilder itself continued rendering and client-side visual evolution remained functional.

## Root cause

The player snapshot showed `view.seat.submitted = true` while the game remained in Planning because another seat had not submitted yet.

`web/src/App.tsx` previously started the automatic `planning-preview` effect whenever the phase was Planning, without checking whether the local seat had already submitted.

That caused repeated calls to:

`POST /api/v1/games/game-1/seats/1/planning-preview`

with an otherwise valid empty command batch. `GameSession.PlanningPreview` correctly rejected that boundary state with `seat 1 already submitted turn 1`, but the error was not typed as a normal session rejection. `writeHostError` therefore fell through to HTTP 500 / `internal_error`.

## Fix

1. The web client no longer requests a Planning preview when `snapshot.view.seat.submitted` is true.
2. Session-level Planning-preview boundary rejections now wrap `session.ErrPlanningPreviewRejected`.
3. `Host.PlanningPreview` maps that typed rejection to `app.ErrSessionRejected`, which the HTTP layer already exposes as HTTP 409 / `session_rejected` rather than 500.
4. Genuine resolver/internal Planning-preview failures remain unmapped and therefore still surface as internal server errors.

## Regression coverage

Added app tests proving:

- a freshly generated game supports an empty Planning preview before submission;
- after one local seat submits while another local seat remains pending, another Planning preview for the submitted seat returns `ErrSessionRejected`.

Existing session Planning-preview recalculation test remains green.

## Live browser verification

The existing preview server was **not restarted**, preserving the in-memory `game-1`.

After rebuilding/reloading only the web client:

- `game-1` still reported `seat.submitted = true`;
- no danger/error notice rendered;
- browser Resource Timing contained no `planning-preview` request;
- only normal game-list and player-snapshot API reads occurred.

The backend status-code hardening will apply on the next intentional server restart; the client-side fix already removes the live recurring 500 without destroying the current game state.