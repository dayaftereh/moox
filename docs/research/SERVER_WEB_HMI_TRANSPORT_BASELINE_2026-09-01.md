# Slice 08 research - authoritative server, web HMI and transport baseline

Status: **Gates 1-3 complete; Gate 4 pending**
Date: 2026-09-01

## Accepted architecture baseline

`ADR-0004-authoritative-server-web-client.md` is accepted before this slice starts:

- one authoritative Go game server is the only gameplay mutation boundary;
- normal browser/web HMI is primary for local and remote use;
- HTTP carries authoritative request/command/projection operations;
- WebSocket carries revision/session/battle lifecycle notifications;
- missed WebSocket delivery is recoverable by fetching a fresh authoritative HTTP projection/revision;
- Wails v3 is optional native packaging/window integration only and must not expose a separate gameplay authority path;
- tactical rendering may later be 2D or 3D without changing BattleSession authority.

Slice 08 must turn that architecture into the smallest durable application/transport baseline without adding gameplay rules.

## Gate 1 repository/session check

Starting HEAD: `41f5fb7` (`docs: plan server-web roadmap through first victory`).
Branch: `main`, ahead of `origin/main` by 35 commits at Gate-1 start.
Starting working tree: clean.
Starting `_OPEN_*.md` count: 0.
Foreign active writer: none; Slice-08 Gate-1 write session is the sole active project writer.
Core state schema: 21.

A single recovery marker was then created at `docs/slices/_OPEN_SERVER_WEB_HMI_TRANSPORT_BASELINE_2026-09-01.md`.

## Investigation questions

1. Which currently public GameSession/BattleSession/player/observer use cases need remote transport exposure?
2. Which revision/event boundaries already support reconnect/resync, and where is a transport-neutral method missing?
3. What is the smallest sensible HTTP routing/WebSocket dependency set in Go today?
4. What local bind/auth/TLS default is safe without pretending public-Internet deployment/authentication is solved?
5. What browser frontend/build/embed approach gives the smallest durable baseline without creating a second game architecture?
6. How can a future Wails wrapper consume exactly the same web/server contract?
7. What exact API/message/package contract should Gate 2 accept before implementation?

Findings are appended below during Gate 1. No application implementation belongs in this document phase.


## Existing transport-neutral runtime inventory

Gate-1 inspection confirms there is currently no `internal/server`, no browser/frontend tree and no `cmd/moox-server`; the only executable under `cmd/` is `moox-analyze`. `go.mod` has no third-party requirements. The local toolchain is Go 1.26.3, Node 24.15.0 and npm 11.12.1.

The existing transport-neutral protocol is already a strong base:

- `protocol.CommandSchemaVersion = 1`;
- `Command {schema_version, sequence, kind, payload}`;
- `CommandBatch {schema_version, game_id, seat_id, turn, base_revision, commands}`;
- `DomainEvent {schema_version, sequence, turn, revision, scope, kind, seat_id, command_sequence, data}`;
- Seat -> Empire authority is held by `GameSession`, not by gameplay command payloads;
- `GameSession` and `BattleSession` clone/validate authoritative state and views rather than returning mutable pointers.

Current strategic command kinds already include Population assignment, Research selection, Construction queues, Colony/Outpost/Freighter/Military Ship production, military design save, Population transfer, Fleet move/split/merge, colonization and Outpost deployment. Tactical commands are currently exactly `battle.fire_beam` and `battle.end_activation`.

### Public use-case classification for a remote application

The exported `GameSession` surface should not be mapped blindly 1:1 to HTTP. Gate 1 classifies it as follows.

| Existing use case | Remote role | Slice-08 transport conclusion |
| --- | --- | --- |
| `PlayerView(seatID)` | player read | primary player projection; expose through player snapshot |
| `ObserverView()` | privileged observer/debug read | expose only through explicit observer endpoint/config boundary |
| `ResearchChoices(seatID, rules)` | player legal-action read | transport-worthy, but not required for the minimal first browser command proof |
| `ConstructionChoices(seatID, colonyID, rules)` | player legal-action read | transport-worthy, but not required for the minimal first browser command proof |
| `BuildingChoices(...)` | narrower/legacy legal-action read | do not make a separate baseline endpoint unless Gate 2 proves it is still needed beyond `ConstructionChoices` |
| `ColonyBaseResolutions(seatID)` | player immediate-choice read | already embedded in `PlayerView`; separate endpoint unnecessary for baseline |
| `SubmitTurn(CommandBatch)` | player mutation | first-class HTTP command endpoint |
| `SubmitBattleCommand(battleID, seatID, Command)` | player mutation | first-class HTTP battle-command endpoint |
| `ResolveColonyBaseCommand(...)` | player immediate post-resolution mutation | expose through a generic immediate-command endpoint because it can block turn completion |
| `PublishDraftTelemetry(...)` | AI/dev diagnostics | not part of baseline player HMI; observer/AI adapter can be added later |
| `ResolveStrategic(resolver)` | server-owned lifecycle | **never** direct player HTTP API |
| `CompleteTurn()` | server-owned lifecycle | **never** direct player HTTP API |
| `CompleteResearchField(...)` / `GrantTechnology(...)` | server/system transitions | not player API |
| `BeginEncounters(...)` | internal/test/legacy lifecycle | not player API |
| `CompleteBattle(...)` | legacy/external-result adapter | not a normal tactical-player API; tactical-enabled Battles use `SubmitBattleCommand` |
| `SubmittedBatches()` | internal/observer/testing | no separate player endpoint; observer state already contains submissions |

### Missing application-host layer

No package currently owns a registry such as `gameID -> GameSession`, shared rules/resolver dependencies, automatic server-side phase progression, live subscribers or static web serving. That is the main application gap for Slice 08.

Gate 1 recommends introducing a transport-neutral application host between Session and HTTP. It should own registered sessions and their resolver/rules dependencies while leaving all gameplay validation in existing Core/Game/Session packages.

## Reconnect/resync analysis

The existing strategic `revision` is necessary but **not sufficient as the sole browser invalidation sequence**.

Direct code evidence:

1. `GameSession` starts at revision 1.
2. A Seat submission is stored immediately during `SubmitTurn`, but revision does not increment for a partial submission. `PlayerView`/`ObserverView` can therefore change while the strategic revision remains unchanged.
3. Draft telemetry can change `ObserverView` without changing game revision.
4. Tactical commands increment a Battle-private `tacticalRevision`, but `battle.View` does not expose that revision. `TacticalState.NextCommandSequence` is exposed, but there is no unified game-level revision change for each nonterminal tactical action.
5. `PlayerView` currently contains no Battle views at all; only `ObserverView` exposes `Battles`. A remote participating player therefore has no transport-neutral way to reconstruct an active Tactical Battle after reconnect without using the privileged Observer projection.

Consequences:

- WebSocket must not be modeled as `revision_changed` only.
- A separate **application-level change sequence** is needed for transport invalidation. It is non-authoritative, is not saved/replayed and increments whenever a hosted public projection changes, including partial submission and Battle-local changes.
- HTTP remains the source of truth. WebSocket messages only say that a snapshot became stale; they do not carry authoritative patches in Slice 08.
- A participating-player Battle projection must be added without exposing `ObserverView`. Gate 2 should choose either a new `GameSession.PlayerBattleViews(seatID)` / `BattleView(seatID,battleID)` accessor or an equivalent player-scoped Session snapshot. Gate 1 recommends a player-scoped accessor owned by `internal/session`, so transport code never filters a privileged full Observer view.
- A small `GameSession.Status()` projection (`game_id`, revision, turn, phase) is also recommended so the application host can drive phases/list games without repeatedly cloning the full Observer state.

### Proposed reconnect protocol

Use a monotonic hosted-game `change_sequence` distinct from gameplay revision:

- HTTP player/observer snapshot envelope contains `change_sequence` and current game `revision`.
- WebSocket notifications contain `change_sequence`, current game `revision`, a reason/scope and optional Battle ID.
- WebSocket carries no authoritative state patch in Slice 08.
- On initial load/reconnect, client connects/reconnects and fetches a fresh HTTP snapshot. Any later notification with a higher `change_sequence` invalidates that snapshot and triggers a refetch.
- A lost WebSocket interval is harmless: reconnect + fresh HTTP snapshot recovers current truth. No notification replay buffer is required in the baseline.

This explicitly handles partial Seat submission and Battle-local changes that do not alter strategic revision.

## Server-side lifecycle coordination

The browser should not call `ResolveStrategic` or `CompleteTurn`. A transport-neutral application host should synchronously drive server-owned automatic phases after successful player mutations:

1. `SubmitTurn` is accepted by `GameSession`.
2. If the Session reaches `strategic_resolution`, host calls the configured resolver.
3. If resolution enters `encounters`, automatic progression stops and participants act through Battle commands.
4. If resolution reaches `post_resolution`, host checks pending interactive Colony Base resolutions.
5. If none remain, host calls `CompleteTurn` and returns to `planning`; if an interactive resolution is pending, it stops and waits for the player immediate command.
6. A successful terminal Battle command or Colony Base immediate command runs the same automatic-phase driver again.

This is application scheduling only; existing deterministic GameSession transitions remain authoritative and testable independently.

## Go HTTP/router/WebSocket dependency analysis

### HTTP routing

Use the Go standard library `net/http` and `http.ServeMux` rather than Gin/Echo/Chi for the Slice-08 baseline.

Reason:

- MOOX is already Go 1.26.
- Standard `ServeMux` has method-aware patterns and path wildcards since Go 1.22, including patterns such as `GET /resource/{id}` and `Request.PathValue`.
- The planned API is small enough that a framework adds dependency/API surface without a current requirement for advanced middleware/routing.

### WebSocket

The standard library still does not provide the browser WebSocket upgrade/API required here. Do **not** use deprecated `golang.org/x/net/websocket`.

Recommended dependency for Gate 2: `github.com/coder/websocket`, currently v1.8.15 on pkg.go.dev (published 2026-06-15, ISC license).

Reasons relevant to this baseline:

- minimal idiomatic API and first-class `context.Context` support;
- current active maintenance;
- JSON helper package;
- same-origin WebSocket requests are accepted by default while cross-origin requests are rejected unless explicitly configured;
- write-only/server-notification connections are directly supported through `CloseRead`;
- no web framework is required.

Gorilla WebSocket remains a valid mature alternative, but Slice 08 does not need its prepared-write/buffer tuning advantages. The smaller context-oriented Coder API fits the notification-only channel better.

### WebSocket semantics

Keep commands on HTTP. WebSocket is server -> browser invalidation/lifecycle notification only in Slice 08. This sharply reduces concurrency and authority complexity and preserves one normal request/result path for commands.

## Browser frontend/build analysis

There is currently no Node/package configuration in the repository. The installed Node 24.15.0 satisfies current Vite requirements.

Options considered:

- **Vanilla TypeScript + Vite:** smallest dependency surface, but the intended MOOX HMI is a long-lived multi-panel, highly stateful application; using no component/state library would quickly recreate framework responsibilities.
- **React + TypeScript + Vite:** modest dependency cost, official Vite `react-ts` template, clear component/state model and straightforward future integration of Canvas/Three.js/WebGL tactical rendering. This is the recommended baseline.
- full-stack/SSR frameworks: rejected for this slice because Go is the authoritative server and MOOX has no SEO/SSR requirement; a second Node server/runtime would duplicate application responsibility.

Gate-1 recommendation:

- `web/` as a Vite + React + TypeScript SPA;
- strict TypeScript;
- native browser `fetch` and `WebSocket` initially;
- no Redux/TanStack Query/router/UI component framework in Slice 08 unless Gate 2 identifies a concrete need;
- simple CSS and a minimal diagnostic/game shell only;
- Vite dev server proxies `/api` (including WebSocket path) to the Go server so browser traffic remains same-origin and no permissive CORS policy is required;
- do **not** enable WebSocket origin rewriting in the Vite proxy.

### Static production assets

Do not couple normal `go test ./...` to an already-generated Node `dist/` tree.

Recommended baseline:

- Vite builds to `web/dist`;
- `internal/server` accepts an `fs.FS`/static-handler source rather than hard-coding `go:embed`;
- standalone `cmd/moox-server` can serve `web/dist` through an `os.DirFS`/filesystem adapter;
- future Wails/release packaging can embed/copy the exact same built assets and pass them to the same server static-handler contract;
- a single-binary embed can be revisited as packaging work without changing HTTP/game semantics.

Go's standard `embed.FS` implements `io/fs.FS` and is compatible with `net/http`, so this interface keeps future embedding available without forcing generated assets into the baseline Go build.

## Bind, same-origin, authentication and TLS boundary

Gate 1 recommends a deliberately safe local default rather than pretending production multiplayer identity is solved.

Baseline security contract proposed for Gate 2:

- default standalone bind: `127.0.0.1:8080` (configurable);
- future native wrapper starts the same server on `127.0.0.1:0`/an ephemeral loopback port and points its WebView at that URL;
- production browser/API is same-origin; no wildcard CORS;
- mutating browser HTTP requests require JSON and should reject a mismatched `Origin` when an Origin header is present;
- WebSocket uses default same-origin verification; no `InsecureSkipVerify`/allow-all setting;
- request bodies have explicit size limits; server has read-header/idle timeouts and graceful shutdown;
- Observer API is an explicit privileged/debug surface, not something player endpoints reuse internally;
- Slice 08 does **not** implement a production account/identity provider.

Non-loopback/public exposure must be explicit and documented as requiring a trusted VPN/reverse proxy/auth/TLS boundary. The standalone server should refuse accidental non-loopback unauthenticated binding by default; if Gate 2 wants a development escape hatch it must be loudly opt-in and named insecure. Production-grade Seat authentication is a later application/security slice.

This separates two concepts that are currently easy to confuse:

- **gameplay authority:** already server-owned Seat -> Empire mapping and command legality;
- **transport identity:** proof that a network caller is allowed to act as a Seat/observer, not yet implemented.

## Optional Wails wrapper verification

Current Wails v3 documentation confirms both that:

- Wails can run a GUI-free server build with HTTP assets and WebSocket-backed events; and
- a native Wails `WebviewWindow` can load an arbitrary URL through its URL/`SetURL` capability.

MOOX deliberately does **not** use Wails server-mode bindings/events as the primary game protocol, because doing so would make the supposedly optional shell part of the authoritative transport contract. The useful verified property is simpler: a future Wails shell can start the same MOOX HTTP server and navigate its native WebView to `http://127.0.0.1:<port>` with no direct MoveFleet/FireBeam/etc. Go bindings.

Native Wails-only APIs may later cover file dialogs, tray/menu, fullscreen/window state and OS notifications. Gameplay remains HTTP/WebSocket.

## Proposed Gate-2 implementation contract

This is the Gate-1 recommendation to discuss/accept in Gate 2; it is **not implemented yet**.

### Packages / ownership

- `internal/app`: transport-neutral hosted-game registry, resolver/rules dependencies, automatic phase driver, player/observer snapshot envelopes, monotonic non-authoritative `change_sequence`, subscriber fan-out.
- `internal/server`: `net/http` API + `github.com/coder/websocket` adapter + static SPA serving. No gameplay rules.
- `cmd/moox-server`: standalone configuration/startup. Until Slice 09, an explicit development/demo mode registers one deterministic `core.NewSmallFixture` GameSession rather than inventing a second New Game system.
- `web/`: React + TypeScript + Vite browser HMI.
- `internal/session`: add only the minimal transport-neutral projections proven missing in Gate 1, preferably `Status()` and player-scoped Battle views; do not make `ObserverView` a player-data source.

### API version 1 draft

Use path versioning under `/api/v1` and JSON DTOs with their own `schema_version: 1` where they are not already `protocol` types.

Minimum endpoints:

- `GET /healthz`
- `GET /api/v1/games` -> hosted-game summaries/status
- `GET /api/v1/games/{gameID}/seats/{seatID}/snapshot` -> player snapshot envelope (`change_sequence`, Game revision + current `PlayerView` + player-scoped Battle views)
- `GET /api/v1/games/{gameID}/observer/snapshot` -> privileged observer snapshot envelope
- `POST /api/v1/games/{gameID}/turn-submissions` -> existing `protocol.CommandBatch`; path Game ID and body Game ID must agree
- `POST /api/v1/games/{gameID}/immediate-commands` -> `{schema_version, seat_id, command}` for currently supported post-resolution immediate commands such as Colony Base resolution
- `POST /api/v1/games/{gameID}/battles/{battleID}/commands` -> `{schema_version, seat_id, command}`
- `GET /api/v1/games/{gameID}/stream` upgraded to WebSocket -> invalidation notifications only

Legal-action endpoints (`ResearchChoices`, `ConstructionChoices`) should be added in the same application/server shape when the first real HMI screen requires them, but Gate 1 does not require them for the minimal remote-control proof. Do not create a separate endpoint for every exported Session helper merely because it exists.

### HTTP rejection envelope

Do not parse current `fmt.Errorf` strings into dozens of pseudo-stable error codes in Slice 08.

Baseline mapping:

- malformed path/JSON/schema/content type -> HTTP 400, code `bad_request`;
- unknown hosted game/resource -> 404, code `not_found`;
- disabled/unauthorized observer or transport principal once applicable -> 403, code `forbidden`;
- valid transport request rejected by Session/gameplay authority/phase/revision -> 409, code `session_rejected`, preserve human-readable message;
- unexpected adapter/internal failure -> 500, code `internal_error` without leaking sensitive internals.

Typed domain error refinement can be a later contract once concrete clients need machine-specific rejection handling.

### WebSocket notification draft

```text
{
  schema_version: 1,
  kind: "snapshot_invalidated",
  game_id: "...",
  change_sequence: N,
  game_revision: R,
  scope: "session" | "battle" | "observer",
  battle_id?: B,
  reason: "submission" | "strategic_resolution" | "battle_command" | "turn_advanced" | ...
}
```

The reason is transport/UI metadata only, not an authoritative replay event. Clients refetch after invalidation rather than applying patches.

### Minimal browser proof

The Slice-08 HMI only needs to prove the architecture, not become the final game UI:

- connect to/list the development hosted game;
- select/use the development Seat on loopback;
- show game ID, turn, phase, strategic revision and `change_sequence`;
- show current Empire/Colonies from `PlayerView`;
- submit at least one existing `CommandBatch` flow through HTTP (a simple Population assignment or empty-turn baseline is acceptable only if Gate 2 chooses it explicitly);
- receive WebSocket invalidation and refetch the resulting snapshot;
- display an active participant Battle when one is present and prove the battle-command HTTP endpoint in integration tests even if the minimal visual demo does not synthesize a battle every run;
- reconnect/refetch without relying on retained WebSocket messages.

### Explicit Slice-08 non-goals

No New Game generator, diplomacy, invasion, new combat mechanics, AI strategy, save-service redesign, public account system, matchmaking, full PWA/offline mode, polished game UI, Three.js tactical renderer or native Wails package is part of this slice.

## External technical references checked on 2026-09-01

- Go routing enhancements / standard `ServeMux`: https://go.dev/blog/routing-enhancements
- Go `embed` package: https://pkg.go.dev/embed
- `github.com/coder/websocket` package/docs: https://pkg.go.dev/github.com/coder/websocket
- Vite getting started/templates: https://vite.dev/guide/
- Vite dev proxy/WebSocket options: https://vite.dev/config/server-options
- React component/state guidance: https://react.dev/learn/managing-state
- Wails v3 server build: https://v3.wails.io/guides/server-build/
- Wails v3 Window URL API: https://v3.wails.io/reference/window/

## Gate-1 conclusion

The existing deterministic Session/Protocol architecture is suitable for a web-first application without reworking gameplay. Slice 08 should add a thin application host and HTTP/WebSocket/browser adapter, plus two small missing Session projections (`Status` and player-scoped Battles). The critical design point is to keep strategic revision, Battle command sequence and application `change_sequence` as distinct concepts rather than forcing every live UI change into the authoritative replay revision.

Gate 1 should stop before implementation and ask for Gate-2 acceptance of the proposed contract above.


## Gate-1 verification

Before the Gate-1 handoff, `go test ./... -count=1` passed all packages on the unchanged gameplay/runtime code. Gate 1 changes documentation/recovery state only; no Go, ruleset, protocol schema or frontend implementation has been added yet.


## Gate 2 accepted implementation contract

Status: **accepted on 2026-09-01; Gate 3 implementation pending**.

The user accepted the Gate-1 proposal without scope expansion. The following details are now binding for Gate 3 unless a concrete implementation contradiction is discovered and documented before changing them.

### 1. Application/package ownership

Gate 3 will add exactly these new primary boundaries:

- `internal/app`: transport-neutral hosted-game/application layer. It owns the in-process `gameID -> hosted GameSession` registry, loaded rules/resolver dependencies, automatic server-owned phase progression, player/observer snapshot envelopes, monotonic application `change_sequence`, and subscriber invalidation fan-out.
- `internal/server`: HTTP/WebSocket/static-file adapter only. It owns request decoding/encoding, API routing, transport error mapping, origin/content-type/body-limit policy and WebSocket connections. It must not duplicate gameplay legality or directly mutate `core.GameState`.
- `cmd/moox-server`: standalone server executable/configuration. Until Slice 09 introduces real New Game generation, an explicit development/demo bootstrap may host one deterministic `core.NewSmallFixture`-based one-seat `GameSession` and current EconomyResolver/rules. This is development scaffolding, not a second New Game implementation.
- `web/`: React + TypeScript + Vite SPA.
- `internal/session`: only the minimal missing transport-neutral read projections proven by Gate 1: lightweight Session status and participant-scoped Battle views. Do not expose/filter privileged `ObserverView` to synthesize a player Battle view.

No Core state schema, command schema, event schema, tactical ruleset schema or MOO2 gameplay rule is changed by Slice 08 unless Gate-3 implementation reveals an unavoidable contradiction and stops for review.

### 2. Session/application status projections

Add a detached lightweight Session status value containing at minimum:

```text
GameID
Revision
Turn
Phase
```

Add a participant-scoped Battle projection such that a Seat can read only Battles in which its SeatID is a participant. The result uses existing detached `battle.View` values, remains sorted by Battle ID and must not reveal nonparticipant Battle state.

The application player snapshot combines current `session.PlayerView` plus that Seat's participant Battle views. The observer snapshot may wrap the existing privileged `session.ObserverView` only through the explicit observer surface.

### 3. `change_sequence` versus authoritative revisions

The accepted web synchronization model has three deliberately distinct concepts:

- **Game revision**: existing authoritative `GameSession` strategic revision, preserved exactly for command/replay semantics.
- **Battle command/tactical state**: existing BattleSession command sequence/runtime revision semantics, preserved exactly for tactical command legality.
- **Application `change_sequence`**: new non-authoritative hosted-game invalidation sequence used only by web/application clients.

`change_sequence` begins at 1 when a GameSession is registered with the application host and increments exactly once after each successful externally visible hosted-game mutation operation after the application phase driver has reached the next stable interactive boundary. Examples include a partial Seat submission, final Seat submission plus automatic strategic resolution/turn progression, an accepted immediate command, and an accepted Battle command plus any resulting automatic continuation.

Rejected operations do not advance `change_sequence`.

It is not persisted into Core state, does not enter deterministic DomainEvent replay and is not used as gameplay legality input.

A single hosted operation may internally cause more than one authoritative Game revision transition; the resulting HTTP receipt/snapshot reports the final revision together with the one new `change_sequence`.

### 4. Automatic server-owned phase driver

The browser never receives endpoints for `ResolveStrategic`, `CompleteTurn`, `CompleteResearchField`, `GrantTechnology`, `BeginEncounters` or arbitrary legacy `CompleteBattle` injection.

After an accepted player mutation, `internal/app` drives only server-owned deterministic lifecycle steps until the next player-interactive boundary:

1. accepted `SubmitTurn`;
2. if Session phase becomes `strategic_resolution`, call the configured Resolver;
3. if phase becomes `encounters`, stop while any Battle requires participant action;
4. if phase becomes `post_resolution`, inspect pending Colony Base resolutions;
5. if an immediate player decision is pending, stop;
6. otherwise call `CompleteTurn` and stop at the next `planning` phase;
7. after accepted terminal Battle or immediate post-resolution command, run the same phase driver again.

The driver orchestrates existing Session APIs; it does not reimplement strategic/tactical rules.

### 5. HTTP API v1

Use Go standard `net/http`/`http.ServeMux` and path versioning under `/api/v1`.

Baseline endpoints are fixed as:

```text
GET  /healthz
GET  /api/v1/games
GET  /api/v1/games/{gameID}/seats/{seatID}/snapshot
GET  /api/v1/games/{gameID}/observer/snapshot
POST /api/v1/games/{gameID}/turn-submissions
POST /api/v1/games/{gameID}/immediate-commands
POST /api/v1/games/{gameID}/battles/{battleID}/commands
GET  /api/v1/games/{gameID}/stream    # WebSocket upgrade
```

`GET /api/v1/games` returns lightweight hosted-game summaries and no privileged Core state.

Player snapshot envelope concept:

```text
{
  schema_version: 1,
  change_sequence: N,
  view: <session.PlayerView>,
  battles: [<participant battle.View>]
}
```

Observer snapshot envelope concept:

```text
{
  schema_version: 1,
  change_sequence: N,
  view: <session.ObserverView>
}
```

Turn submission body is the existing `protocol.CommandBatch`; path `gameID` and body `game_id` must agree.

Immediate and Battle command request envelope:

```text
{
  schema_version: 1,
  seat_id: S,
  command: <protocol.Command>
}
```

Battle path ID must agree with the target Battle selected in the application operation.

Successful mutating endpoints return HTTP 200 with a receipt rather than returning an authoritative patch:

```text
{
  schema_version: 1,
  game_id: "...",
  change_sequence: N,
  game_revision: R
}
```

The client may immediately refetch; other clients learn about the invalidation over WebSocket.

Legal-action endpoints for Research/Construction are deliberately not required in Slice 08. Add them later through the same application boundary when the first real HMI screen needs them; do not create endpoints for every exported Session helper by default.

### 6. HTTP validation/error contract

All JSON endpoints use `application/json`. Mutating requests reject unsupported/missing JSON content type and enforce explicit bounded request-body sizes.

API error envelope schema 1:

```text
{
  schema_version: 1,
  error: {
    code: "bad_request" | "not_found" | "forbidden" | "session_rejected" | "internal_error",
    message: "human-readable detail"
  }
}
```

Mapping:

- malformed route/path/JSON/schema/content type -> 400 `bad_request`;
- unknown hosted game/resource -> 404 `not_found`;
- disabled observer or disallowed transport principal -> 403 `forbidden`;
- syntactically valid request rejected by Session/gameplay authority, phase, sequence or revision -> 409 `session_rejected`;
- unexpected adapter/application failure -> 500 `internal_error`, without leaking sensitive internals.

Gate 3 must not parse current Go error strings into a large unstable machine-code taxonomy. More specific typed domain errors are deferred until a client has a concrete need.

### 7. WebSocket notification/reconnect contract

Commands stay on HTTP. The WebSocket is server -> client invalidation/lifecycle notification only.

Schema 1 notification:

```text
{
  schema_version: 1,
  kind: "snapshot_invalidated",
  game_id: "...",
  change_sequence: N,
  game_revision: R,
  scope: "session" | "battle" | "observer",
  battle_id?: B,
  reason: "submission" | "strategic_resolution" | "battle_command" | "immediate_command" | "turn_advanced" | "other"
}
```

`reason`/`scope` are application/UI metadata only and are not authoritative DomainEvents.

There is no WebSocket state-patch protocol and no required notification replay buffer in Slice 08. On initial connect/reconnect the client fetches a fresh HTTP snapshot. A notification with a higher `change_sequence` invalidates the current snapshot and causes a refetch.

Subscriber implementation may coalesce/drop stale intermediate invalidations for a slow client as long as the newest sequence remains observable; correctness comes from refetching the HTTP snapshot, not from receiving every notification.

### 8. Local transport identity and security boundary

Slice 08 intentionally solves gameplay authority, not Internet account identity.

Accepted baseline:

- standalone default bind: `127.0.0.1:8080`;
- future Wails wrapper: loopback ephemeral port (`127.0.0.1:0` or equivalent);
- same-origin API/SPA/WebSocket operation; no wildcard CORS;
- mutating browser requests validate Origin when present;
- `github.com/coder/websocket` default same-origin policy is retained; no allow-all origin mode;
- Observer endpoint is disabled by default and requires an explicit local-development/debug server option to enable;
- Seat ID in Slice-08 loopback requests is a **development principal selector**, not cryptographic authentication;
- accidental unauthenticated non-loopback exposure is not a supported secure mode. A direct non-loopback development escape hatch, if implemented, must be explicit and visibly named unsafe/insecure;
- real remote/public deployment must remain behind a trusted same-host reverse proxy/VPN/auth/TLS boundary until MOOX gains an authenticated transport-identity layer.

Gameplay authority remains the existing server-owned Seat -> Empire mapping and Session validation; transport identity is a separate future security concern.

### 9. Go dependency and server policy

- HTTP/router: standard library `net/http` + method/path-aware `http.ServeMux`.
- WebSocket: `github.com/coder/websocket`, pinned by Go modules during Gate 3; do not use deprecated `golang.org/x/net/websocket`.
- No Gin/Echo/Chi dependency in the baseline.
- HTTP server config includes finite request-body limits, `ReadHeaderTimeout`, `IdleTimeout`, graceful shutdown and sane WebSocket close handling.
- Server/adapters must remain pure Go and headless; normal Go test/build paths must not require Wails or a GUI runtime.

### 10. Web frontend/build/static assets

Accepted browser stack:

- React;
- TypeScript in strict mode;
- Vite;
- native browser `fetch` and `WebSocket` for Slice 08.

Do not add Redux, TanStack Query, a client router, CSS/UI component framework, SSR/full-stack Node server or Three.js in Slice 08 unless a concrete Gate-3 blocker is demonstrated.

Development:

- Vite proxies `/api` and the WebSocket path to the Go server;
- browser traffic remains same-origin;
- do not rewrite WebSocket Origin to conceal a cross-origin design.

Production/static boundary:

- Vite outputs `web/dist`;
- `internal/server` accepts an `fs.FS`/static-handler-compatible asset source;
- normal `go test ./...` must not depend on an already-generated `web/dist`;
- standalone development/release server may use filesystem assets;
- future packaging may use `embed.FS` or Wails asset bundling without changing the API/gameplay contract.

### 11. Minimal Gate-3 browser proof

The initial HMI is an architecture proof, not the final MOO2 interface.

It must be able to:

- discover/connect to the development hosted game;
- choose/use the one development Seat on loopback;
- display game ID, turn, phase, authoritative game revision and application `change_sequence`;
- display the player's Empire and Colonies from `PlayerView`;
- expose simple farmer/worker/scientist inputs for one owned Colony and build a real `colony.assign_population` command inside an existing `protocol.CommandBatch`;
- submit the real turn batch over HTTP;
- receive `snapshot_invalidated` and refetch the resulting player snapshot;
- reconnect/refetch cleanly without relying on retained WebSocket messages;
- display participant Battle data when present. The Battle HTTP command endpoint must be proven by integration tests; Slice 08 does not need to synthesize/polish a tactical battle UI on every demo startup.

Using only an empty command batch is not sufficient as the primary browser proof; the browser must submit at least one existing concrete gameplay command.

### 12. Optional Wails wrapper contract

No Wails package is required in Gate 3.

If/when added later, Wails may only:

- start/stop the same local MOOX server;
- load the same `http://127.0.0.1:<port>` web client in its WebView;
- provide native-only conveniences such as file dialogs, tray/menu, fullscreen/window state and OS notifications.

It must not expose direct gameplay bindings such as MoveFleet, SaveMilitaryDesign, SubmitTurn bypasses or FireBeam. Browser/PWA/remote deployment/Wails all consume the same HTTP/WebSocket game contract.

### 13. Compatibility and non-goals frozen for Slice 08

Gate 3 must preserve:

- all current deterministic Core/Game/Protocol/Session/Battle tests and semantics;
- existing `cmd/moox-analyze` behavior/build;
- headless pure-Go operation;
- current schema versions unless a separately reviewed contradiction requires change;
- transport-arrival-order independence of authoritative replay/events.

Explicitly deferred from Slice 08:

- real New Game/galaxy generation (Slice 09);
- diplomacy, invasion or any new MOO2 gameplay rule;
- public account/authentication system and matchmaking;
- full PWA/offline behavior;
- production deployment automation;
- polished strategic UI;
- tactical 3D/Three.js;
- native Wails installer/wrapper implementation;
- broad legal-action endpoint catalog;
- active-game durable process-restart persistence.

**Gate 2 is accepted. Gate 3 must implement this contract without widening the application/gameplay surface.**


## Gate-2 verification

After accepting the implementation contract, `go test ./... -count=1` passed all packages on the unchanged gameplay/runtime code. Gate 2 remains documentation/decision-only: there are still no Go/server/frontend/ruleset implementation changes and no schema bumps.


## Gate 3 implementation evidence

Gate 3 implements the accepted Gate-2 contract without adding MOO2 gameplay rules or changing Core/Protocol/Event/Tactical schema versions.

### Session projection boundary

`internal/session` now exposes two small transport-neutral read surfaces:

- `GameSession.Status()` returns detached `game_id`, strategic `revision`, `turn` and `phase` without cloning the privileged full Observer state;
- `GameSession.PlayerBattleViews(seatID)` returns only detached Battles whose participant list contains that Seat, sorted by Battle ID.

Tests prove a partial Seat submission can change visible readiness while `Status().Revision` remains 1, and prove three different Seats receive only their participant Battles. The HTTP/application layer therefore never filters `ObserverView` to manufacture player Battle state.

### Hosted-game application layer

New package `internal/app` owns the transport-neutral hosted runtime:

- in-process GameSession registration and stable game listing;
- resolver/immediate-resolver dependencies;
- schema-1 player/observer snapshot envelopes and mutation receipts;
- non-authoritative `change_sequence` beginning at 1;
- subscriber fan-out with one-element newest-invalidation coalescing;
- automatic server-owned phase progression.

The accepted sequence distinction is preserved in code:

- strategic Game revision remains Session authority/replay state;
- Battle command/runtime sequencing remains BattleSession authority;
- application `change_sequence` is only a live projection invalidation sequence.

App tests prove the key counterexample: the first of two Seat submissions changes `change_sequence` from 1 -> 2 while Game revision remains 1. A rejected duplicate advances neither. The second Seat submission is one hosted operation; after automatic strategic resolution plus turn completion it produces `change_sequence` 3 and final Game revision 3 / turn 2 / planning. A slow subscriber may miss the intermediate notification but receives the newest sequence and can recover by HTTP refetch.

The phase driver exposes no player endpoint for `ResolveStrategic`, `CompleteTurn`, research grants/completion or arbitrary Battle completion. It continues existing Session transitions only until `planning`, an interactive `encounters` Battle, or a pending Colony Base decision.

### HTTP / WebSocket server

New package `internal/server` implements the accepted `net/http` API under `/api/v1` plus static SPA serving. The only new Go dependency is pinned as:

```text
github.com/coder/websocket v1.8.15
```

Implemented routes:

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

Transport controls implemented from the contract:

- JSON mutation bodies are bounded to 1 MiB, require `application/json`, reject unknown fields/trailing JSON and validate schema/Seat/command structure;
- mutation requests reject a present Origin whose host differs from the request Host;
- WebSocket uses Coder's default same-origin handshake verification; it calls `CloseRead` because browser data messages are not part of the protocol;
- Observer projection is disabled unless the server config explicitly enables it;
- adapter errors map to schema-1 `bad_request`, `not_found`, `forbidden`, `session_rejected` and generic `internal_error` envelopes;
- successful mutations return only the accepted schema-1 receipt; state is recovered from snapshots;
- WebSocket emits only `snapshot_invalidated` application notifications and carries no state patches or replay events;
- unknown `/api/...` paths cannot fall through to the SPA index fallback.

Server integration coverage proves:

1. a real `colony.assign_population` command submitted as a revision-bound `protocol.CommandBatch` through HTTP reaches the authoritative EconomyResolver event history and automatically returns the Session to planning on turn 2;
2. the corresponding WebSocket invalidation reports the final `change_sequence`/Game revision, and a fresh HTTP request after connection close reconstructs the same truth;
3. disabled Observer, unknown game, missing content type and cross-origin mutation boundaries return the intended HTTP/error families;
4. the Battle endpoint reaches Session validation for rejected targets;
5. an exact original-derived Fusion-Laser-Frigate vs unarmed Nuclear-Frigate strategic fixture is created by two HTTP turn submissions, appears only through the participant player Battle projection, and accepts a real `battle.fire_beam` HTTP command. That operation advances application sequence while leaving the strategic Game revision unchanged and advances Tactical `NextCommandSequence` from 1 to 2;
6. resolving the same deterministic Population command directly through GameSession/EconomyResolver and through the hosted HTTP path yields byte-identical final `session.ObserverView` JSON. The transport/application layers therefore do not change the authoritative replay/state result;
7. concrete static assets are served while browser routes fall back to `index.html` and unknown API routes remain 404.

### Standalone server

New executable `cmd/moox-server` provides the Slice-08 standalone application host.

Default/runtime policy:

- bind `127.0.0.1:8080`;
- refuse non-loopback binds unless the operator explicitly supplies `-insecure-allow-nonloopback`;
- privileged Observer endpoint is opt-in with `-enable-observer`;
- finite `ReadHeaderTimeout` and `IdleTimeout` plus graceful interrupt shutdown;
- normalized rules default to `data/rulesets/moo2-1.31`;
- built web assets default to `web/dist` but are optional;
- until Slice 09, development bootstrap hosts exactly one deterministic `demo` GameSession from `core.NewSmallFixture(0x8008)` with Seat 1 and the normal EconomyResolver.

This development fixture is explicitly not a second New Game implementation. A unit test verifies loopback acceptance/refusal policy.

### Browser HMI

New `web/` is a strict TypeScript React/Vite SPA using only native `fetch` and browser `WebSocket` for application communication.

Installed baseline at Gate-3 implementation time:

- React / React DOM 19.2.x;
- Vite 8.2.2;
- TypeScript 7.0.2;
- React Vite plugin and type packages.

No Redux, TanStack Query, router, UI framework, SSR/full-stack Node server or Three.js was introduced.

Vite development config binds loopback and proxies `/api` and `/healthz` to the Go server with `changeOrigin: false`; WebSocket upgrade is enabled through the `/api` proxy.

The HMI proves the contract rather than attempting final game UX. It:

- discovers hosted games;
- selects the development Seat;
- fetches a player snapshot;
- displays game, turn, phase, Game revision, application change sequence, Empire/Seat and owned Colonies;
- displays participant Battle summaries when present;
- aggregates current Population cohorts into farmer/worker/scientist form values;
- constructs the real schema-1 `colony.assign_population` command and `protocol.CommandBatch` with the snapshot's turn/revision;
- submits it through HTTP;
- connects to the notification WebSocket, refetches when a newer invalidation arrives and refetches again before/after reconnect attempts;
- displays the latest invalidation for diagnostic visibility.

`npm run build` passes and emits normal `web/dist` production assets. `web/dist`, `web/node_modules` and TypeScript build-info are local/generated and ignored by Git; normal Go builds/tests do not depend on them.

### Real standalone/browser proof

After the production frontend build, the actual standalone server was launched on loopback port 18080 with `web/dist` and verified live:

```text
/healthz                              -> ok
/api/v1/games                        -> one `demo`, change_sequence 1
/api/v1/games/demo/seats/1/snapshot -> turn 1, revision 1, change_sequence 1, one Colony
/                                     -> production Vite index/assets
```

Headless Chrome 151 with a virtual-time budget rendered the real SPA after asynchronous API/WS startup and showed:

- `Live invalidation stream connected`;
- hosted game `demo`;
- turn 1 / planning / Game revision 1 / change sequence 1;
- Fixture Empire / human / Seat 1 local_human;
- participant Battles section;
- real Population assignment form initialized to Farmers 2 / Workers 1 / Scientists 1 / total 4.

The temporary port-18080 server process was then explicitly stopped; no listener remains.

### Application documentation / Wails boundary

`docs/architecture/WEB_APPLICATION.md` now documents the implemented topology, production-style and Vite-development startup, API, revision model, phase driver, HMI proof, local security boundary and optional future Wails packaging.

Wails remains absent from the Slice-08 runtime dependency graph. A future native shell may start this same server and load this same URL, plus native file/window/tray conveniences. No direct gameplay binding was introduced.

### Schema and scope boundary

Gate 3 intentionally leaves these authoritative versions unchanged:

- Core StateSchemaVersion: **21**;
- command schema: **1**;
- event schema: **1**;
- economy ruleset schema: **8**;
- ship-hulls ruleset schema: **3**;
- tactical-combat ruleset schema: **1**.

No New Game generator, diplomacy, invasion/conquest, new tactical mechanic, Internet identity/account layer, matchmaking, PWA/offline mode, tactical 3D renderer or Wails package was added.

Gate 4 remains responsible for a fresh independent final QA, implementation/evidence commit, closing documentation/HISTORY commit and OPEN-marker removal.


## Gate-3 final verification

Fresh verification after all Gate-3 code, web assets/source and documentation changes:

- changed/new Go files inspected by `gofmt -l`: **8 files, 0 dirty**;
- focused `go test ./internal/session ./internal/app ./internal/server ./cmd/moox-server -count=1 -v`: **PASS**;
- `go test ./... -count=1`: **PASS all packages**, including existing `cmd/moox-analyze` and all deterministic Core/Game/Battle/Session regressions;
- `go vet ./...`: **PASS**;
- `npm run build` in `web/`: **PASS** with strict TypeScript and Vite production build;
- generated `web/node_modules`, `web/dist` and `web/tsconfig.tsbuildinfo`: **ignored and absent from Git status**;
- temporary standalone port 18080 listener count after browser proof: **0**;
- exactly one `_OPEN_*.md`: **Slice 08**;
- Gate checklist counts: **G1 unchecked 0, G2 unchecked 0, G3 unchecked 0, G4 unchecked 7**;
- schemas unchanged: **Core21 / Command1 / Event1 / Economy8 / Ship-Hulls3 / Tactical1**;
- forbidden alternate application dependencies (`Wails`, Gin, Echo, Chi, deprecated `x/net/websocket`): **0**;
- newly introduced gameplay command constants in the Gate-3 diff: **0**;
- live stale Gate-3-pending status references: **0**;
- trailing whitespace on changed/new text files: **0**;
- `git diff --check`: **PASS**.

Gate 3 is complete. Gate 4 is the only remaining Slice-08 gate and must perform a fresh independent QA before committing and closing the slice.


## Gate 4 fresh pre-commit verification

Gate 4 repeated the Slice-08 verification independently before any commit:

- changed/new Go files: **8**, `gofmt -l` dirty count **0**;
- focused `go test ./internal/session ./internal/app ./internal/server ./cmd/moox-server -count=1 -v`: **PASS**;
- full `go test ./... -count=1`: **PASS all packages**;
- `go vet ./...`: **PASS**;
- `npm run build`: **PASS** with Vite 8.2.2 / strict TypeScript;
- `go build ./cmd/moox-server`: **PASS**;
- dependency closure for `cmd/moox-server` contains **0 Wails packages**;
- frontend bypass scan found **0** direct Observer/server-owned lifecycle/Wails gameplay bypass references;
- generated `node_modules`, `dist` and `*.tsbuildinfo` remain ignored and absent from Git status;
- fresh standalone runtime on `127.0.0.1:18082`: `/healthz=ok`, hosted `demo` game/change 1, player snapshot turn 1/revision 1/change 1 and built SPA asset index all verified;
- headless Chrome rendered the production SPA with `Live invalidation stream connected`, `demo`, turn 1/planning/revision 1/change 1, Fixture Empire and the real Population command form;
- temporary port-18082 process was stopped, listener count returned to **0** and the temporary `moox-server.exe` produced by the build check was removed;
- `git diff --check`: **PASS**.

The implementation is therefore ready for the Gate-4 implementation/evidence commit, followed by closing documentation/HISTORY and OPEN-marker removal.
