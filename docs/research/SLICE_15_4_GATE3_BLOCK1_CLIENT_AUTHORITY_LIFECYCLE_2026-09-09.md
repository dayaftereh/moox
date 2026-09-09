# Slice 15.4 Gate 3 - Block 1 client authority/lifecycle foundation

Date: **2026-09-09**

Status: **implemented and browser-verified**.

## Scope

First implementation block under the accepted Gate-2 contract:

- client types for already-projected Colony Base and Battle decisions;
- fuller player-safe Battle Spec/View types matching existing Go JSON;
- structured API errors preserving HTTP status, server error code and message;
- explicit game lifecycle state separate from free-form event/status messages;
- mutation gate for direct End Turn, Diplomacy and Invasion operations while authoritative state is not synchronized;
- lifecycle exported into DOM through `data-lifecycle` for deterministic UI/QA behavior.

## Client contract additions

`web/src/api.ts` now types:

- `ColonyBaseResolution`;
- `BattleSide`;
- `BattleSpec`;
- `BattleResult`;
- `BattleView`;
- `BattleDecision`;
- `DecisionCatalog.colony_base`;
- typed `DecisionCatalog.battles`.

No battle legality is inferred in TypeScript. Battle commands remain the exact projected protocol-command array.

## Structured API error

`APIError` preserves:

- HTTP status;
- server `error.code`;
- authoritative message.

`requestJSON` now throws `APIError` instead of flattening everything into a generic `Error` string. Existing UI can display `code: message`, while later Save/Load conflict UX can branch directly on codes such as `game_exists`, `ruleset_mismatch` and `game_id_mismatch`.

## Lifecycle state

Frozen lifecycle enum implemented in App:

- `initial-loading`;
- `connecting`;
- `synced`;
- `refreshing`;
- `reconnecting`;
- `refresh-failed`;
- `restoring`;
- `importing`;
- `fatal`.

Current WebSocket/snapshot behavior now updates this explicitly. Last-known state may remain rendered during reconnect/refresh failures, but direct server mutation is gated unless lifecycle is `synced`.

## Browser verification

Canonical port: **7171**.

Normal `game-1` route:

- `data-lifecycle="synced"`;
- connection frame tone success;
- no danger notice;
- End Turn/Fertig enabled in Planning;
- no document overflow.

Intentional nonexistent-game test:

- route changed to `does-not-exist`;
- lifecycle became `fatal`;
- rendered authoritative structured error `not_found: not found: game "does-not-exist"`;
- connection label changed to authoritative state unavailable.

Recovery test:

- navigate back to `game-1`;
- lifecycle returned to `synced`;
- error cleared;
- End Turn/Fertig re-enabled.

## Regression

- TypeScript project build PASS;
- Vite production build PASS;
- `git diff --check` PASS.

This block intentionally does not yet expose Save/Load controls; it provides the authority/lifecycle foundation they require.
