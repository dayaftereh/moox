# Open Slice 08 - Authoritative server, web HMI and transport baseline

Status: **Gates 1-3 complete; Gate 4 pending**.

Planned specification: `docs/slices/PLANNED_08_SERVER_WEB_HMI_TRANSPORT_BASELINE.md`.
Permanent evidence: `docs/research/SERVER_WEB_HMI_TRANSPORT_BASELINE_2026-09-01.md`.
Architecture decision: `docs/architecture/ADR-0004-authoritative-server-web-client.md`.

## Recovery checklist

- [x] Gate 1: repository/session + architecture/transport analysis complete.
- [x] Gate 2: exact implementation decision accepted.
- [x] Gate 3: implementation + transport/browser integration tests.
- [ ] Gate 4: final QA, commits and closure.

## Gate-1 scope

Gate 1 is analysis/documentation only. Do not implement the server, HTTP/WebSocket adapters, browser HMI or Wails wrapper yet.

Required analysis:

- current Protocol/GameSession/BattleSession/player/observer use-case inventory;
- revision/event/reconnect/resync sufficiency and missing transport-neutral boundaries;
- HTTP/router/WebSocket dependency trade-off;
- local bind/security/auth/TLS boundary;
- browser frontend/build/embed baseline;
- optional Wails wrapper consuming exactly the same server/web contract;
- exact proposed Gate-2 implementation contract.

## Current resume point

Gates 1-3 are complete. The accepted server/web contract is implemented and integration-tested. Resume only at Gate 4 for fresh final QA, commits and closure; do not widen gameplay scope.
