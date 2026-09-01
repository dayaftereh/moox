# ADR-0001: Go core with Wails v3 application shell

Date: 2026-08-26
Status: Partially superseded by `ADR-0004-authoritative-server-web-client.md`: the pure-Go/Wails-independent core boundary remains accepted, while Wails is no longer the primary application/UI architecture.

## Decision

MOOX will be designed around a **Go-first architecture**.

The gameplay simulation, data model, persistence logic, tools and reverse-engineering/analyzer code must be implemented as **pure Go wherever possible**, with the explicit target that these packages compile with:

```text
CGO_ENABLED=0
```

The application UI is planned around **Wails v3** (referred to in the initial project discussion as "Wayless V3") so that the same Go-centric project can target multiple desktop platforms and, where supported by Wails v3, Android/mobile targets.

## Architectural boundary

The important portability rule applies to the **MOOX-owned core**, not necessarily to every native implementation detail used internally by Wails on every platform.

Therefore the repository should keep these layers separate:

```text
MOOX Core / Rules / Analyzer / Persistence
    pure Go
    no Wails imports
    no CGO dependencies
    deterministic and testable headlessly

Application Services / Adapters
    pure Go where possible
    expose use cases to the UI

Wails v3 Shell
    frontend + generated bindings
    platform integration
    may use platform-specific Wails build/runtime mechanisms
```

This prevents the simulation and analysis code from becoming dependent on WebView, desktop windowing libraries, JNI, NDK, Apple frameworks or other platform-specific details.

## Platform goal

Target platforms should include at least:

- Windows
- Linux
- macOS
- Android

A later iOS target remains possible through the same architecture, although iOS builds require Apple's toolchain/macOS.

## Current Wails v3 note

As of 2026-08-26, Wails v3 documents Android and iOS support using the same Go application/frontend model as desktop. Android uses an Android WebView plus native Android tooling/NDK and a Go shared library/JNI bridge. Consequently, we should **not** assume that `CGO_ENABLED=0` is a valid requirement for every final Wails platform wrapper/build step.

The stronger and more useful invariant is:

> All MOOX domain logic and developer tools, including the MOO2 analyzer, remain pure-Go and Wails-independent, so they can be built/tested with `CGO_ENABLED=0` and reused from any frontend/platform wrapper.

## Consequences for the MOO2 analyzer

The analyzer should not initially be built as a Wails feature. It should be a reusable Go library plus a thin CLI.

Suggested separation:

```text
internal/lbx/        LBX container/parser library
internal/moo2data/   typed MOO2-specific decoders
internal/catalog/    normalized inventory/provenance model
cmd/moox-analyze/    CLI entry point
```

Later, the Wails application can call the same analyzer packages through an application service without duplicating parsing logic.

## Why

- one language for simulation, tools and backend services;
- deterministic headless tests are easy to run in CI;
- cross-compilation remains practical;
- no UI dependency leaks into game rules;
- Android/mobile support does not force a second game implementation;
- the reverse-engineering tool can run independently of the game UI;
- future server/multiplayer/simulation tooling can reuse the same core.
