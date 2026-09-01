# MOOX Architecture

Status: architecture checkpoint updated on 2026-09-01; ADR-0004 makes the authoritative Go server + web client the primary application topology.

This document is the central runtime-architecture overview for MOOX. More focused decisions and contracts remain in the other files in this directory.

## Goals

MOOX is designed as a portable, deterministic, server-authoritative game engine with multiple interchangeable clients and controllers.

The main goals are:

- keep all game rules and simulation logic in Go and independent from UI/transport code;
- make one authoritative Go game server the only gameplay mutation boundary;
- make a normal browser/web client the primary HMI for local and remote use;
- use HTTP for authoritative requests/commands/projections and WebSocket for revision/lifecycle notifications;
- keep HTTP projection/revision state as the resynchronization source of truth after reconnect or missed WebSocket delivery;
- retain Wails v3 only as an optional native one-click wrapper around the same local server/web client contract;
- support local single-player and multiplayer from the same simulation/session model;
- allow multiple players to plan strategic turns in parallel;
- support human players, built-in AI, remote AI and future MCP-controlled AI through the same command/session authority model;
- make tactical battles isolated child sessions of the strategic game;
- allow 2D or future 3D rendering without moving tactical authority into the frontend;
- preserve deterministic replay, save/load, observer/debug visibility and headless testability.

## Layered architecture

```text
 Browser / PWA             Optional Wails native shell
       |                           |
       +-----------+---------------+
                   |
             HTTP + WebSocket
                   |
        +----------v-----------+
        |  MOOX Game Server    |
        | authoritative host   |
        +----------+-----------+
                   |
        Application / adapters
                   |
        +----------v-----------+
        |     GameSession      |
        |   BattleSession      |
        +----------+-----------+
                   |
            Commands / Events
                   |
        +----------v-----------+
        | Deterministic Core   |
        | rules + GameState    |
        +----------+-----------+
                   |
        +----------v-----------+
        | Normalized rulesets  |
        +----------------------+
```

The browser frontend never owns authoritative game rules. HTTP/WebSocket handlers adapt existing Session/Protocol use cases; they do not reimplement validation. A missed WebSocket notification must be recoverable by fetching a fresh authoritative projection/revision over HTTP.

Wails is not a separate gameplay frontend contract. If shipped, it starts/hosts the same local server and opens the same web HMI, while native-only helpers are limited to platform conveniences such as window/file/tray integration.

Other controllers/adapters reuse the same authoritative application/session boundary:

- headless/dedicated server deployments;
- built-in AI;
- external AI agent API;
- MCP adapter for LLM-based players;
- observer/debug tooling.

Accepted decision: `ADR-0004-authoritative-server-web-client.md`.

## Runtime package direction

The implemented runtime split after Slice 08 is:

```text
internal/core
    deterministic state, IDs, RNG, validation, save/load and pure rules

internal/game
    strategic commands, validation, resolution and domain events

internal/session
    GameSession, seats, turn phases, submissions and player/observer projections

internal/protocol
    versioned commands, events and transport-neutral envelopes

internal/battle
    tactical BattleSession state and combat commands/resolution

internal/app
    hosted-game registry, automatic server-owned phase driver, player/observer snapshot envelopes,
    non-authoritative change_sequence and invalidation subscribers

internal/server
    net/http API, WebSocket invalidation stream, transport validation/error mapping and SPA assets

cmd/moox-server
    loopback-first standalone server and deterministic development fixture bootstrap

web
    React + TypeScript + Vite browser presentation and user interaction only

internal/agent
    future/common controller boundary for built-in and external AI

internal/mcp
    optional MCP adapter over the normal application/game API

optional future Wails shell
    starts the same local server and loads the same web URL; native conveniences only
```

The authoritative application topology and local workflow are documented in `WEB_APPLICATION.md`. Wails is no longer a privileged application-service layer and must never bypass the HTTP/WebSocket gameplay contract.

## Authoritative state

A `GameSession` owns exactly one authoritative `GameState`.

Clients never submit arbitrary replacement state. They submit player intent as commands. This prevents a client from directly inventing resources, ships, technology or other state.

```text
Player intent
    |
    v
Command
    |
    v
server validation + deterministic resolution
    |
    +--> DomainEvent(s)
    |
    v
new authoritative GameState
```

The same rule applies to human players and AI controllers.

## Parallel strategic planning

Strategic planning is parallel by design.

At the beginning of a strategic turn every player receives a view derived from the same authoritative turn state. Each client maintains a local draft while the player works.

```text
                    Authoritative Turn N
                           /      \
                          /        \
                 Player A Draft  Player B Draft
                    |                 |
              local commands     local commands
                    |                 |
                    +-------+---------+
                            |
                       SubmitTurn
                            |
                            v
                    server resolution
```

Draft changes are immediate locally so the UI remains responsive. A player may experiment, undo changes and change decisions without mutating the server state.

When the player chooses `End Turn`, the client submits an ordered command batch rather than a complete state snapshot.

Conceptually:

```text
SubmitTurn
    game_id
    player_id / seat_id
    turn
    base_revision
    commands[]
```

The server validates that the submission belongs to the expected game, seat, turn and revision before accepting it.

## Commands, events and state

Commands describe intent. Events describe what actually happened. State is the current materialized result.

Example:

```text
MoveFleetCommand(Fleet 7 -> Orion)
              |
              v
        strategic resolver
              |
              +--> FleetMovementStarted
              +--> FleetArrived
              +--> EnemyFleetDetected
              +--> EncounterCreated
              |
              v
          updated GameState
```

This distinction is central to the architecture because it supports:

- deterministic replay;
- multiplayer synchronization;
- validation and anti-cheat boundaries;
- observer/debug timelines;
- AI diagnostics;
- regression tests;
- future savegame migration and audit tooling.

The engine does not need to be a pure event-sourced system internally, but commands and resulting domain events are first-class runtime data.

## Strategic turn lifecycle

A strategic turn is modeled as a session state machine rather than a single `Turn++` call.

```text
TURN N
  |
  +-- PLANNING
  |     - players work in parallel
  |     - submitted players wait
  |     - AI seats may plan concurrently
  |
  +-- STRATEGIC RESOLUTION
  |     - submitted command batches are validated/resolved
  |     - movement/economy/research/etc. produce events
  |
  +-- ENCOUNTERS
  |     - tactical battles are created when required
  |     - unrelated battles may run in parallel
  |
  +-- POST RESOLUTION
  |     - battle results and remaining effects are applied
  |
  +-- TURN N+1
```

The exact original MOO2 resolution ordering remains an implementation/research question. The networking/session architecture must not force an unverified rule ordering.

## Encounters and tactical battles

A tactical battle is a child session of the strategic `GameSession`.

```text
GameSession
    |
    +-- StrategicState
    |
    +-- BattleSession #501
    |      Human vs Psilon
    |
    +-- BattleSession #502
           Sakkra vs Meklar
```

Each `BattleSession` owns its own deterministic battle state, participants, battle turn/initiative state, RNG state, commands, events and final result.

Known original-game evidence already confirms that tactical space combat is a turn-based 2D battlefield with individual ships and concepts including movement, attacks, shields, armor, structure, internal systems, boarding and retreat. Exact initiative, command ordering and combat formulas remain separate fidelity work and are not assumed by this architecture.

### Parallel battles

Independent battles may execute in parallel so unrelated players do not block each other unnecessarily.

```text
Strategic resolution
       |
       +--> Battle A: Player 1 vs Player 2
       |
       +--> Battle B: Player 3 vs AI 4

Battle A and Battle B may be played concurrently.
```

Each encounter receives a stable identity and deterministic RNG context. Completed battle results are applied to the strategic state in a deterministic ordering independent of wall-clock completion time.

Players not involved in an unresolved encounter enter a waiting state for that dependency while the observer can inspect active encounters.

## Seats and controllers

A game seat represents participation in a game. The gameplay model should not depend on whether that seat is controlled by a human or an AI.

Possible controllers include:

```text
Seat 1 -> local human
Seat 2 -> remote human
Seat 3 -> built-in Go AI
Seat 4 -> external AI agent
Seat 5 -> MCP-controlled LLM agent
```

Every normal controller receives an allowed player projection and submits normal game commands.

This keeps AI from bypassing game rules and makes human/AI multiplayer combinations natural.

## PlayerView and information boundaries

Clients do not receive the complete authoritative `GameState`.

The server projects state according to the requesting role:

```text
GameState
    |
    +--> PlayerView(player A)
    +--> PlayerView(player B)
    +--> ObserverView(admin/observer)
```

`PlayerView` contains only information legally visible to that player. Hidden systems, enemy information, intelligence data and other fog-of-war state must remain server-side.

AI controllers use the same player-facing information boundary unless they are explicitly configured as privileged/debug agents.

## Observer and replay model

Observer support is a first-class architectural feature, not a later debug hack.

A privileged observer can inspect:

- complete authoritative strategic state;
- all seats and readiness/submission status;
- strategic command batches;
- resulting domain events;
- active and completed encounters;
- tactical battle sessions;
- AI actions and optional decision metadata;
- turn/session phase and waiting dependencies.

A replay-oriented view presents the command/event history chronologically.

Example:

```text
Turn 42
  Psilon      AssignPopulation(...)
  Human       MoveFleet(Fleet 17 -> Sol)
  Sakkra AI   QueueProduction(Colony Ship)
  System      EncounterCreated(#193)
  Human       Battle #193 won
  System      FleetDestroyed(...)
```

### Optional draft telemetry

Unsubmitted player drafts are not authoritative and normally remain private/local.

For development, AI debugging or explicitly privileged observer sessions, an optional telemetry channel may expose non-authoritative planning activity such as draft commands, reversals and AI decision metadata.

Draft telemetry must remain separate from the authoritative command/event log so it cannot affect replay or deterministic game resolution.

AI diagnostics may include explicit short metadata such as a goal or reason supplied by the AI. The architecture does not require or depend on hidden model reasoning.

## AI and external agent architecture

The canonical AI integration point is a normal game/agent API, not MCP itself.

```text
                    Game / Agent API
                         /   |   \
                        /    |    \
                Built-in AI HTTP/WS  MCP adapter
```

The common contract should expose player-visible state/legal actions and accept the same commands used by human clients.

A future external agent can therefore be implemented using:

- native Go;
- remote HTTP/WebSocket;
- another local process;
- an LLM through MCP;
- another automation system.

### MCP role

MCP is an optional semantic adapter for LLM-controlled players and tooling. It must not become the authoritative multiplayer protocol or contain separate game logic.

Potential MCP resources/tools can map onto the normal API, for example:

```text
Resources
    game://<id>/player/me
    game://<id>/colonies
    game://<id>/fleets
    game://<id>/research

Tools
    list_legal_actions
    assign_population
    queue_production
    select_research
    move_fleet
    colonize_planet
    end_turn
```

Tactical combat can expose the same pattern through `BattleView` plus normal battle commands such as move, rotate, fire, retreat and end-action/turn once the exact combat model has been implemented.

## Transport independence

The application/game boundary is transport-independent.

The same operations must be usable through:

- direct in-process Wails services for local desktop play;
- HTTP/WebSocket for remote multiplayer;
- a headless/dedicated server;
- AI agents;
- MCP tooling.

No game rule may depend on a particular transport or frontend lifecycle.

## Determinism contract

Multiplayer and parallel battles must not weaken the existing deterministic-core guarantees.

Runtime resolution must therefore avoid depending on:

- wall-clock completion order;
- Go map iteration order;
- frontend/UI state;
- transport packet timing;
- process scheduling;
- uncontrolled random-number sources.

Stable game/turn/encounter IDs and deterministic RNG streams/seeds are used so that identical accepted commands against identical initial state produce identical authoritative results.

## Current implementation relationship

The architecture now has two implemented runtime foundations.

`internal/core` provides:

- stable IDs;
- simulation-owned SplitMix64 RNG with serializable state;
- minimal galaxy/empire/colony state;
- turn and legacy core event primitives;
- validation;
- deterministic JSON serialization;
- atomic save/load;
- a fixed deterministic regression fixture.

The command/session boundary is now implemented across `internal/protocol`, `internal/game`, `internal/session` and `internal/battle`:

- versioned `Command` and `CommandBatch` envelopes with ordered command sequences;
- transport-independent `game.Resolver` input/output contract for resolved state, strategic domain events and encounter requests;
- transactional `ResolveStrategic`: resolver output is fully validated/prepared before any authoritative state/revision/event commit;
- authoritative `GameSession` with `planning`, `strategic_resolution`, `encounters` and `post_resolution` phases;
- stable `Seat` ordering with local-human, remote-human, built-in-AI, external-AI and MCP-AI controller identities;
- parallel turn submission against `turn` + `base_revision` guards;
- deterministic authoritative submission logging in seat order, independent of network arrival order;
- minimal `PlayerView` that exposes only the requesting empire/colonies plus public seat readiness;
- privileged `ObserverView` with full state, submissions, events, battles and optional draft telemetry;
- observer-only non-authoritative draft/AI telemetry kept outside the replay event stream;
- deterministic encounter IDs/seeds and minimal tactical `BattleSession` lifecycle;
- parallel battle completion with authoritative result logging in battle-ID order rather than wall-clock completion order;
- detached/copying views so callers cannot mutate authoritative session state through returned DTOs;
- server-owned Seat -> Empire resolver authority, preventing a command from claiming another empire;
- first real `colony.assign_population` command with ownership/assignment validation;
- domain-native `float64` Population, Food, Production, Research, BC and Construction progress driven by normalized ruleset data per ADR-0003;
- materialized Population capacity, Food/Production sustenance, available Construction PP and direct-float turn-end Growth;
- explicit Gravity/Government/local-Morale economy context and adjusted output, kept separate from base output for later additive leader/technology layers;
- minimal colony building inventory with ruleset validation for Barracks/Holo/Pleasure morale behavior.

The strategic runtime follows the original MOO2 1.31 materialize/apply pipeline. Current-turn Economy/Dynamics and Food logistics are materialized first; Research consumes the pre-growth RP snapshot, Population Growth/Starvation is applied, then Construction consumes the already-materialized pre-growth PP snapshot. The post-turn Economy/Dynamics snapshot is recalculated afterwards. Direct original-executable control/data flow and regression tests lock this ordering.

Starvation and Food/Freighter logistics are implemented as an empire-wide deterministic phase. Direct original-executable analysis now also fixes the insufficient-Freighter import policy as round-robin deficit allocation for MOOX's current single-cohort Population model. Research now has server-authoritative General/ordinary/Creative/Uncreative multi-application semantics. Research project switching is implemented with exact RP transfer and Observer/replay history. Direct original-executable analysis has resolved both the MOO2 1.31 turn-order question and Uncreative new-game fixed-application timing/shared-RNG ownership. Hyper-Advanced repeated-field level/cost state, Advanced-start randomized/race-aware technology ownership and authoritative external Technology grants are implemented from direct original evidence. System-level blockade state and its Food/Freighter import/export/sale effects are now implemented from direct original evidence. Population relocation now uses authoritative semantic Settler state: each interstellar transfer reserves five Freighters ahead of Food allocation, same-system moves are immediate, and original ETA/FTL/arrival rules are modeled in Core schema 11. The active Economy follow-up is Housing / Cloning Center / medicine Population-growth modifiers; Fleet/Diplomacy-derived blockade production remains attached to the future strategic Fleet model.

Network transports, Wails services and MCP remain adapters to this boundary and should not be introduced into the deterministic core.
## Related architecture documents

- `ADR-0001-go-wails-v3.md` - Go/Wails v3 portability and layer-boundary decision.
- `ADR-0002-research-float64.md` - historical first RP-native `float64` decision; superseded for general numeric policy by ADR-0003.
- `ADR-0003-domain-native-float64.md` - canonical continuous-quantity and explicit rounding-boundary policy.
- `ADR-0004-authoritative-server-web-client.md` - authoritative Go server, browser-first HTTP/WebSocket client and optional Wails wrapper decision.
- `WEB_APPLICATION.md` - implemented Slice-08 server/web runtime, API/revision/security model and local development workflow.
- `CORE.md` - deterministic core-state, RNG and save/load contract.
- `SESSION_PROTOCOL.md` - concrete implemented session/command/observer contract.
- `ASSET_PIPELINE.md` - normalized asset pipeline.
- `../IMPLEMENTATION_PLAN.md` - staged clean-room implementation plan.
