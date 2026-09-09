# Slice 15.4 Gate 3 - Block 2 Save / Load / Restore browser UX

Date: **2026-09-09**

Status: **implemented and end-to-end verified**.

## Scope

Implements the accepted local-file persistence contract on top of the existing server-authoritative live-snapshot API.

## API wrappers

The browser client now has structured wrappers for:

- `GET /api/v1/games/{gameID}/live-snapshot` -> export Blob;
- `POST /api/v1/games/import` -> import new hosted game;
- `PUT /api/v1/games/{gameID}/live-snapshot` -> atomically restore an existing hosted game.

All three use the structured `APIError` path from Gate-3 Block 1.

## Save Game

Available from the in-game main menu.

- Requires synchronized authoritative state.
- Uses the server export endpoint; no React-owned serialization.
- Downloads `moox-<gameID>-turn-<turn>.json`.
- Lifecycle becomes `exporting` during the request, then returns to the actual connection state.
- Direct gameplay mutation is temporarily gated while persistence work is active.

Live browser verification on canonical 7171:

- menu item present and enabled in synced Planning;
- export succeeded;
- status reported `game-1, Runde 1 gespeichert`;
- lifecycle returned to `synced`;
- no danger notice.

## Load Game reachability

Load Game is exposed in **two places**:

1. in-game main menu;
2. Main Menu/Home hero.

The Home entry is required so a fresh persistence-enabled server with zero hosted games can import a local save.

## Local file metadata

The browser parses only presentation metadata before upload:

- game ID;
- turn when present;
- revision;
- phase;
- schema version.

Server validation remains final authority. Missing/invalid game ID is rejected as file metadata before any destructive request.

## Existing hosted game -> Restore

If fresh `listGames` shows the save's game ID is already hosted, no overwrite occurs automatically. A blocking confirmation displays:

- save game ID;
- save turn;
- explicit warning that the hosted game will be replaced only after server validation;
- Restore / Cancel actions.

Restore uses the existing atomic PUT endpoint.

### Canonical 7171 E2E

A live export of `game-1` was injected into the Load flow:

- lifecycle -> `loading-file`;
- explicit Restore dialog appeared;
- Restore command completed;
- dialog closed;
- lifecycle returned to `synced`;
- End Turn became enabled again;
- host change sequence advanced from 1 to 2 as expected.

Most importantly, the authoritative exported snapshot after restore remained exactly unchanged:

- bytes: **17062**;
- SHA-256: **0819C728B3518F8AC265564CE7FDBA537A67568834E272C3E302E643D760500F**.

The Restore E2E therefore changed host invalidation sequence but not game state.

## New game ID -> Import

The Import branch was tested against an isolated temporary persistence server on `127.0.0.1:7191` so canonical 7171 remained clean.

Test setup:

- fresh server initially exposed zero games;
- Main Menu showed New Game + Load Game;
- test save was the same valid snapshot with its sole game ID changed to `game-import-test`;
- browser selected it through the real Load file change path.

Result:

- no Restore dialog because the ID was not hosted;
- client used POST Import;
- browser automatically navigated to `#/game/game-import-test/galaxy`;
- lifecycle reached `synced`;
- server game list contained exactly `game-import-test`, revision 1, turn 1, phase Planning.

The temporary 7191 server was then terminated and its static test file removed. Canonical 7171 remained healthy.

## Mobile QA

At **320x646**:

- in-game persistence menu fits without document overflow;
- Save/Load entries remain usable;
- Restore sheet is exactly 320px wide and about 187px high for the current content;
- no horizontal overflow;
- Cancel returns `loading-file -> synced`.

## Lifecycle hardening

During review, invalidation handling was tightened so an old/duplicate WebSocket notification does **not** set `refreshing` unless its `change_sequence` is actually newer than the current snapshot. This prevents a theoretical stuck Refreshing state.

Home status also keeps discovery messages such as "no hosted games" unless a persistence lifecycle operation is actually active.

## Regression

- `go test ./internal/app ./internal/server -count=1` PASS;
- TypeScript project build PASS;
- Vite production build PASS;
- `git diff --check` PASS;
- canonical 7171 health remains HTTP 200.
