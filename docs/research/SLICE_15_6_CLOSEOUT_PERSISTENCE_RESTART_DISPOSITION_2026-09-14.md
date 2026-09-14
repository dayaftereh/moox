# Slice 15.6 Closeout - persistence / server-restart disposition

Date: 2026-09-14
Status: **DISPOSITIONED - not a Gate-2 persistence defect**

## Observed symptom

During repeated development-server refreshes on port 7171, `game-triangle-2pc` appeared to fall back from the current playtest state to the original Round-1 reference state. We protected the user state by exporting the current live snapshot before restart and restoring it afterward.

## Root cause

The observation was real, but the original interpretation was wrong.

`moox-server -enable-persistence` does **not** configure automatic disk autosave/reload. It enables the privileged explicit live-snapshot persistence API:

- export the current hosted game (`GET .../live-snapshot`);
- import a snapshot as a hosted game;
- explicitly restore a snapshot into an existing hosted game (`PUT .../live-snapshot`).

`newServerHost()` always creates a new in-memory Host. `-reference-games` then deliberately bootstraps fresh development reference games (`game-1` and `game-triangle-2pc`) on each process start. Therefore a process restart without an explicit restore is expected to produce a fresh reference-game state.

The browser persistence UX is a different and already-authoritative path: `saveGame()` exports the current live snapshot to a downloaded save file, and Load/Restore explicitly reads that file, validates metadata and restores/imports it through the persistence API.

## Slice-15.6 contract impact

The frozen Slice-15.6 Gate-2 contract requires **visible user Save/Load/Resume**, preserving authoritative game identity, turn/revision and meaningful progress. It does not promise hidden server-process crash autosave or automatic reference-game recovery.

Therefore automatic server-restart continuation is not a Slice-15.6 closure blocker and no new disk-persistence architecture is introduced during closeout.

The actual user-facing Save/Load/Resume workflow remains a mandatory item in the final canonical acceptance run and must be proven there.

## Operational hardening

The server flag/help and startup logging now state explicitly that persistence is snapshot-endpoint based and does not auto-save/reload disk state. Development maintenance that must preserve a live reference playtest must continue to:

1. export the current live snapshot;
2. restart/rebuild the server;
3. explicitly restore that exact snapshot;
4. verify turn/phase/revision/game identity.

This is an operational workflow, not a substitute for the browser Save/Load acceptance test.
