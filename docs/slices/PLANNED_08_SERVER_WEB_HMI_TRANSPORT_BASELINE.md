# Planned slice 08 - Authoritative server, web HMI and transport baseline

Status: **active; Gates 1-3 complete / Gate 4 pending**.

Queue position: **8 of 12**.

This file is a prepared slice specification. It is not an `_OPEN_*.md` recovery marker and does not mean implementation has started.

## Objective

Implement the accepted `ADR-0004-authoritative-server-web-client.md` direction as the first real application boundary: one authoritative Go server, one browser-first HMI transport contract, HTTP request/command endpoints and WebSocket revision/lifecycle notifications.

Wails v3 is deliberately demoted from primary gameplay/UI architecture to an optional later native wrapper that starts/hosts the same server and opens the same web client.

This slice changes application/transport architecture, not MOO2 gameplay rules.

## Existing baseline

- `internal/protocol` already defines versioned `Command`, `CommandBatch`, Seat identity, sequence, turn and base revision.
- `internal/session` owns GameSession/BattleSession authority, revisions, player/observer views and deterministic events.
- all meaningful gameplay mutation already occurs below a transport-independent boundary.
- no `cmd/moox-server` or browser HMI currently exists.
- ADR-0004 is accepted as the target application topology.

## Dependencies

- completed deterministic/session foundations through Slice 07;
- no new gameplay slice dependency.

## Scope guard

In scope:

- a pure-Go application/server package and executable entry point;
- versioned HTTP API for session/game discovery, player/observer projection, strategic command submission and BattleSession command submission;
- WebSocket notification stream keyed by authoritative revision/session/battle lifecycle;
- reconnect/resync contract where HTTP projection remains source of truth;
- minimal browser HMI skeleton proving remote control of one existing GameSession flow;
- static web asset serving/development setup;
- explicit bind/authentication/TLS boundary suitable for local development and later remote deployment;
- adapter tests proving transport does not bypass Seat/revision/command validation;
- optional Wails wrapper contract/documentation, without requiring a second gameplay API.

Defer unless Gate 2 deliberately includes a tiny proof:

- polished game UI;
- full PWA/offline behavior;
- production identity provider/account system;
- matchmaking/lobby service;
- public Internet deployment automation;
- tactical 3D rendering;
- native Wails installer/package implementation;
- any new gameplay rule.

## Gate 1 - Checkup + architecture/transport analysis

- [x] Re-check Git/session state and confirm no open slice/foreign writer.
- [x] Re-read ADR-0004, `internal/protocol`, GameSession/BattleSession and Observer boundaries.
- [x] Inventory every currently public use case that a remote player/observer must invoke or read.
- [x] Prove which existing revisions/events are sufficient for reconnect/resync and identify any missing transport-neutral projection method.
- [x] Evaluate standard-library vs selected HTTP/router/WebSocket dependencies and record the dependency/security trade-off.
- [x] Define local-development bind defaults and the boundary for remote authentication/TLS without pretending production auth is solved.
- [x] Evaluate browser frontend stack/build embedding choices and choose the smallest durable baseline.
- [x] Verify that an optional Wails shell can consume the same HTTP/WebSocket client contract without direct gameplay calls.
- [x] Produce permanent architecture/transport evidence and present the exact Gate-2 contract before implementation.

## Gate 2 - Implementation decision

- [x] Accept API versioning and endpoint/resource model.
- [x] Accept HTTP command/result/error envelope and mapping to existing protocol/session errors.
- [x] Accept WebSocket message vocabulary, revision semantics and reconnect/resync behavior.
- [x] Accept player/observer authority/authentication boundary for local baseline.
- [x] Accept frontend framework/build/static-asset approach.
- [x] Accept package/executable boundaries and dependency policy.
- [x] Accept optional Wails-wrapper contract and explicit non-gameplay native surface.
- [x] Freeze compatibility requirements for existing headless tests/CLI/analyzer.

## Gate 3 - Implementation

- [x] Implement the agreed Go server/application adapter and `cmd/moox-server` or equivalent.
- [x] Implement versioned HTTP projection/command endpoints without duplicating gameplay rules.
- [x] Implement WebSocket revision/lifecycle notifications and reconnect/resync behavior.
- [x] Implement the minimal browser HMI proving remote session observation and at least one authoritative command flow.
- [x] Add transport authority/revision/error/reconnect tests.
- [x] Add deterministic integration coverage proving HTTP/WS adapter behavior does not change core/session replay results.
- [x] Document local run/development workflow and optional Wails packaging boundary.

## Gate 4 - Follow-up QA + commit + close

- [ ] Run formatter/linter/build checks for Go and the selected web toolchain.
- [ ] Run focused HTTP/WebSocket authority/revision/reconnect regressions.
- [ ] Run full Go tests and vet.
- [ ] Verify headless server starts without Wails/native GUI dependencies.
- [ ] Verify browser HMI can operate against a remote/loopback server using the same contract.
- [ ] Verify no direct frontend/Wails gameplay bypass was introduced.
- [ ] Run `git diff --check`, update architecture/status/HISTORY and close the marker.

## Exit criterion

A browser can connect to an authoritative MOOX Go server, obtain a revisioned player/observer projection, submit existing strategic/BattleSession commands through the same authority checks used headlessly, receive live revision/lifecycle notifications over WebSocket and recover from a missed connection by resynchronizing over HTTP. Wails is documented only as an optional native wrapper over this same web/server contract.
