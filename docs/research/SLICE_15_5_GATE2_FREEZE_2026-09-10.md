# Slice 15.5 Gate 2 - Interactive Tactical contract freeze

Date: 2026-09-10
Status: **FROZEN / explicitly accepted by user**

The user explicitly accepted the Gate-2 Tactical contract after the final amendments for Tactical Scan/ship inspection and the effectively unbounded MOOX battlefield presentation.

Frozen implementation boundary:

- browser Tactical combat supports 1v1 through 2v2 in the first accepted breadth, with 2v2 as the primary normal-game acceptance fixture;
- Tactical positions remain authoritative integer X/Y;
- ships have authoritative 16-way facing;
- movement cost follows the original evidence: `ceil(sqrt(dx^2 + dy^2)) + facingSteps * turnCostPerStep`;
- turn cost is 2 normally, 1 with Inertial Stabilizer and 0 with Inertial Nullifier;
- Go/BattleSession projects legal movement and fire/target actions; React never computes hidden legality or outcomes;
- add `battle.move_ship`, trusting only ship ID and destination X/Y from the client;
- retain the existing server-owned Laser `battle.fire_beam` and `battle.end_activation` paths;
- Tactical Scan is a free read-only presentation mode for any participant-visible combat ship, including authoritative weapons/readiness, Armor/Structure damage, movement/facing and normalized combat ratings; no invented client-side composite strength score;
- the MOOX battlefield is an effectively unbounded integer-coordinate plane with no normal fixed public width/height and no visible arena wall; finite movement budgets keep each legal-move projection locally finite;
- desktop uses click/drag/wheel interaction; mobile uses tap, one-finger pan, pinch zoom and touch-safe action surfaces with no required hover;
- stale/reconnect/error semantics reuse the frozen Slice-15.4 lifecycle contract and never auto-resubmit rejected Tactical commands;
- Battle completion returns through the already-frozen Slice-15.4 Battle Return surface;
- broader shields/arcs/missiles/fighters/boarding/retreat/etc. remain explicitly deferred unless separately accepted later.

Gate 3 may now implement within this boundary.
