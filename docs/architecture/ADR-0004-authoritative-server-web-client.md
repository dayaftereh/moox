# ADR-0004: Authoritative game server with web-first client and optional Wails shell

- Status: Accepted
- Date: 2026-09-01
- Scope: application/runtime topology, player/observer transport, browser HMI, optional desktop packaging
- Supersedes: the Wails-first application-shell direction in `ADR-0001-go-wails-v3.md`; ADR-0001's pure-Go/Wails-independent core boundary remains valid

## Context

MOOX now has a substantial deterministic headless runtime: versioned commands, Seat authority, revisions, GameSession/BattleSession lifecycle, Observer views, deterministic replay and transactional strategic/tactical resolution. The project is also developed and exercised remotely across multiple machines.

Making Wails the primary UI/application boundary would create little value at this stage and risks introducing a second privileged path into gameplay if a desktop frontend can call Go game methods directly while remote clients use HTTP/WebSocket transport.

The desired product topology is instead a server-authoritative game that can be driven from a normal browser on desktop, tablet or phone. A native desktop package remains useful for one-click local startup, installer/window integration and platform conveniences, but it must not own a separate gameplay API.

## Decision

The primary MOOX application is a **Go authoritative game server plus a web client**.

```text
Browser / PWA / optional Wails WebView
                |
          HTTPS + WebSocket
                |
        MOOX application server
                |
      GameSession / BattleSession
                |
       deterministic Go core
```

The server is the only authority allowed to mutate gameplay state.

### HTTP responsibilities

HTTP is the request/response transport for operations such as:

- create/open/list game/session;
- obtain player or observer projections;
- submit strategic `protocol.CommandBatch` values;
- submit BattleSession commands;
- commit/advance turn lifecycle operations;
- save/load/import/export endpoints when those become available.

The existing command/revision/Seat model remains the gameplay contract. HTTP handlers are adapters; they must not duplicate gameplay validation or rules.

### WebSocket responsibilities

WebSocket is the server-to-client live notification channel for revision/event/lifecycle changes, for example:

- revision changed;
- turn/session phase changed;
- battle created/updated/completed;
- state or legal-action projection invalidated;
- reconnect/resync notification.

WebSocket delivery is not a second source of truth. A client must always be able to resynchronize from an authoritative HTTP projection/revision after reconnect or missed messages.

### Browser HMI

The browser frontend is the primary HMI and may render the same authoritative projections in different ways:

- normal strategic UI in HTML/CSS/SVG/Canvas/WebGL as appropriate;
- tactical combat initially in 2D;
- optional future tactical 3D rendering through Three.js/WebGL/WebGPU or equivalent without changing server gameplay semantics.

Rendering, animation and camera state are presentation concerns and never authoritative combat state.

### Wails role

Wails v3 is retained only as an **optional native distribution shell**. A Wails package may:

- start/stop a local MOOX server;
- open a native WebView/window pointed at the local web client;
- provide file dialogs, tray/menu integration, fullscreen/window state and OS notifications;
- bundle assets/runtime for a one-click local user experience.

Wails must not expose privileged gameplay calls such as MoveFleet, SaveMilitaryDesign or FireBeam that bypass the HTTP/WebSocket server contract.

The same web client must work against:

- a local server started manually;
- a remote/dedicated server;
- a server launched by an optional Wails shell.

## Consequences

- remote development/testing becomes the normal path rather than a special adapter;
- browser, PWA and optional desktop shell share one gameplay transport contract;
- multiplayer remains architecturally compatible with existing Seat/Command/Revision semantics;
- 2D/3D presentation can evolve independently from deterministic gameplay;
- the pure-Go domain/session packages remain reusable and headless;
- future authentication/TLS/deployment policy belongs to the transport/application layer, not Core/Game rules.

## Non-goals of this ADR

This decision does not yet choose:

- the final frontend framework;
- the final HTTP router/websocket library;
- production authentication/identity provider;
- public Internet deployment topology;
- tactical 3D art/rendering technology;
- whether the optional Wails wrapper is shipped for every platform.

Those implementation-level decisions are prepared by Slice 08.
