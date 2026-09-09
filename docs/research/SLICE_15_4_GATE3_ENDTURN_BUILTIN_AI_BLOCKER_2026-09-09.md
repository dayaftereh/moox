# Slice 15.4 Gate 3 - End Turn / Built-in AI blocker remediation

Date: 2026-09-09
Status: **fixed**

## Symptom

In the canonical browser `game-1`, the Human could queue construction and select Research, then press End Turn. The submission was accepted, but the game remained on Turn 1 / `planning` instead of resolving and advancing.

## Root cause

The server already supports explicit `controllers` on New Game and its Human-vs-Built-in-AI automation path is covered by regressions. The browser `CreateGameRequest` type and `App.tsx` New Game request did not send controller assignments. The server therefore used its default `local_human` controller for both generated seats.

Observed live blocker state:

- Seat 1: `local_human`, submitted = true;
- Seat 2: `local_human`, submitted = false;
- Turn 1 / `planning` remained blocked waiting for a second Human submission.

## Fix

- `web/src/api.ts` now exposes the existing server `controllers` request field.
- Browser New Game explicitly creates Seat 1 as `local_human` and Seat 2 as `builtin_ai`.
- No new simulation or automation behavior was introduced; this wires the browser to the already-authoritative server feature.

## Canonical game repair

Before changing the running game, the exact blocked live snapshot was exported to:

`C:\ASH\Workspace\Projects\_moox_game1_endturn_blocker_20260909.json`

The submitted Seat-1 batch contained exactly:

1. `colony.set_construction_queue` -> Colony 53 -> `marine_barracks`;
2. `empire.select_research` -> TechField 4 -> Technology 13.

For recovery, Seat 2 was changed to `builtin_ai`, the already-submitted flag was removed from a temporary repaired base snapshot, and the exact saved Human batch was re-submitted through the normal HTTP turn-submission endpoint. This exercised the ordinary authoritative path and allowed the Built-in AI to respond.

Verified resulting live state:

- Turn **2**, phase **planning**, revision **3**;
- Seat 1 `local_human`, Seat 2 `builtin_ai`;
- Human Research active: TechField **4**, Technology **13**, progress **9 RP**;
- Colony 53 construction active: **marine_barracks**, progress **6 PP**.

## Regression evidence

- `npm run build` passes.
- `go test ./internal/app ./internal/server -run "TestHumanVsBuiltinAIAutomaticTurnsAreExact|TestHTTPCreateGameAcceptsBuiltinAIControllerAssignment"` passes.
- Existing app regression proves a Human submission automatically advances through the Built-in-AI response to the next Human planning boundary when Seat 2 is correctly configured.

## Scope

This is a browser-to-server configuration bug fix. Slice 15.4 Gate 3 Block 6 remains the next planned implementation block.