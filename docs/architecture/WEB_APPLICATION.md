# Web application and authoritative server baseline

Status: implemented by Slice 08 Gate 3 on 2026-09-01.

This document describes how to run and extend the browser-first MOOX application introduced by `ADR-0004-authoritative-server-web-client.md`.

## Runtime topology

```text
Browser / future PWA / optional future Wails WebView
                    |
              HTTP + WebSocket
                    |
             internal/server
                    |
              internal/app
                    |
      GameSession / BattleSession
                    |
       deterministic Core/Game
```

`internal/server` is an adapter. It does not own gameplay rules. `internal/app` hosts sessions, drives server-owned phases and emits non-authoritative snapshot invalidations. `internal/session`, `internal/game`, `internal/battle` and `internal/core` remain the authoritative gameplay boundary.

## Local production-style run

Build the browser assets:

```powershell
cd web
npm install
npm run build
cd ..
```

Start the standalone server from the repository root:

```powershell
go run ./cmd/moox-server
```

Default URL:

```text
http://127.0.0.1:8080/
```

Until Slice 09 adds real New Game generation, `moox-server` hosts one deterministic development game named `demo` based on `core.NewSmallFixture` with Seat 1 as a local developer human.

The standalone server only serves `web/dist` when a built `index.html` exists. The Go test/build path does not require generated frontend assets.

Useful options:

```text
-addr 127.0.0.1:8080
-rules data/rulesets/moo2-1.31
-web web/dist
-enable-observer
-insecure-allow-nonloopback
```

`-enable-observer` exposes the privileged Observer snapshot endpoint and is disabled by default.

`-insecure-allow-nonloopback` is intentionally named unsafe. It only removes the loopback bind guard; it does not add authentication or TLS.

## Frontend development run

Terminal 1:

```powershell
go run ./cmd/moox-server -web ""
```

Terminal 2:

```powershell
cd web
npm install
npm run dev
```

Open:

```text
http://127.0.0.1:5173/
```

Vite proxies `/api` and `/healthz` to `http://127.0.0.1:8080`. The WebSocket path is below `/api`, so the same proxy preserves a same-origin browser contract. `changeOrigin` is deliberately false.

## API v1

```text
GET  /healthz
GET  /api/v1/games
GET  /api/v1/games/{gameID}/seats/{seatID}/snapshot
GET  /api/v1/games/{gameID}/observer/snapshot
POST /api/v1/games/{gameID}/turn-submissions
POST /api/v1/games/{gameID}/immediate-commands
POST /api/v1/games/{gameID}/battles/{battleID}/commands
GET  /api/v1/games/{gameID}/stream
```

The stream endpoint upgrades to WebSocket and emits schema-1 `snapshot_invalidated` messages. It is not a state patch or gameplay event stream. Clients resynchronize from the HTTP snapshot.

## Revision model

Three concepts remain separate:

- `GameSession.Revision`: authoritative strategic revision and command legality input;
- Battle command/tactical sequence: authoritative BattleSession command legality;
- hosted-game `change_sequence`: non-authoritative application invalidation sequence for browser clients.

A partial Seat submission can change a player/observer projection without changing the strategic Game revision. A tactical command can change Battle state without changing the strategic Game revision. This is why the application sequence must not be replaced by or written into the deterministic Game revision.

The hosted `change_sequence` starts at 1 and increments once after each successful externally visible hosted mutation after automatic phase progression reaches the next stable interactive boundary. Rejected operations do not advance it.

## Automatic phase driving

The browser cannot invoke server-owned transitions such as `ResolveStrategic` or `CompleteTurn`.

After a successful player mutation, `internal/app` continues deterministic Session lifecycle work until the next interaction is required:

```text
SubmitTurn
  -> strategic_resolution when all Seats submitted
  -> ResolveStrategic
  -> encounters: wait for participant Battle commands
  -> or post_resolution
       -> pending Colony Base decision: wait
       -> otherwise CompleteTurn
  -> planning
```

Terminal Battle commands and immediate post-resolution commands enter the same driver afterward.

## Current HMI proof

The Slice-08 browser HMI deliberately remains small. It currently proves:

- hosted-game discovery;
- development Seat selection;
- player snapshot retrieval;
- display of game ID, turn, phase, Game revision and application change sequence;
- player Empire and Colony projection;
- participant-only Battle projection;
- creation and submission of a real `colony.assign_population` command in a revision-bound `protocol.CommandBatch`;
- WebSocket invalidation and HTTP refetch;
- reconnect/resync without depending on retained WebSocket messages.

This is an architecture proof, not the final strategic UI.

## Security boundary

Slice 08 does not implement Internet account identity.

Default assumptions:

- standalone bind is loopback only;
- API, SPA and WebSocket are same-origin;
- wildcard CORS is not enabled;
- mutating requests reject mismatched Origin when Origin is present;
- Coder WebSocket default origin verification remains enabled;
- Observer API is opt-in;
- Seat ID is only a local development principal selector, not cryptographic authentication.

Remote/public use must stay behind a trusted VPN or reverse proxy that supplies authentication and TLS until MOOX receives a dedicated transport-identity/security slice.

## Optional Wails packaging

Wails is not part of the Gate-3 runtime dependency graph.

A future Wails package may:

1. start this same local MOOX server on a loopback/ephemeral port;
2. load the same web client URL in its WebView;
3. add native conveniences such as file dialogs, tray/menu, fullscreen/window state and OS notifications.

It must not introduce direct gameplay bindings that bypass HTTP/WebSocket. `MoveFleet`, `SubmitTurn`, `FireBeam`, military design mutations and similar actions remain server-authoritative commands.

## Verification commands

From repository root:

```powershell
go test ./...
go vet ./...
```

From `web/`:

```powershell
npm run build
```

Slice 08 integration tests additionally cover real HTTP Population assignment, WebSocket invalidation/refetch, participant tactical Battle command acceptance, error/origin/Observer boundaries, static SPA fallback and equality between direct Session resolution and the HTTP-hosted authoritative Observer result.
