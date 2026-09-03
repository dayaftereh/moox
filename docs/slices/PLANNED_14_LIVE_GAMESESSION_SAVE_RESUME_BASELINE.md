# Planned slice 14 - Live GameSession save / resume baseline

Status: **open; Gate 1 complete; Gate 2 approval pending**.

Queue position: **14 of 17**.

## Objective

Add versioned persistence for an in-progress authoritative GameSession/hosted game so a real match can be stopped and resumed without changing deterministic outcome, revision, RNG, events, pending decisions or controller identity.

## Dependencies

- Slice 12 completed-session snapshot baseline.
- Slice 13 built-in AI controller contract.

## Scope guard

In scope:

- a versioned live-session snapshot format;
- stable save boundaries: Planning and other interactive decision boundaries, plus Completed;
- GameState, RNG/revision/turn/phase, seats/controllers, submissions where valid, event/telemetry history, Battles, Invasion/handled keys and Result;
- strict load validation and migration-version rejection baseline;
- Host/App/HTTP save/load or export/import surfaces without gameplay bypasses;
- exact save -> load -> continue deterministic equality;
- built-in AI resumes without hidden non-persisted decision state.

Defer:

- save during an in-flight resolver mutation;
- cloud sync/save browser;
- backward migrations from hypothetical future formats;
- original MOO2 SAV binary compatibility.

## Gate 1 - Checkup + persistence audit

- [x] Inventory all GameSession/hosted-game mutable state that affects future determinism.
- [x] Define stable saveable phases and explicit unsaveable boundaries.
- [x] Reconcile Core23 serialization and completed-session snapshot formats.
- [x] Define schema/version/migration policy and canonical byte ordering.
- [x] Define server save/load authority and reconnect semantics.
- [x] Define mid-match AI save/resume exact-continuation fixture.
- [x] Present Gate-2 persistence contract.

## Gate 2 - Implementation decision

- [ ] Freeze live snapshot schema and save boundaries.
- [ ] Freeze load validation/atomic replacement semantics.
- [ ] Freeze HTTP/App surfaces and filename/storage-neutral contract.
- [ ] Freeze exact-continuation regression.

## Gate 3 - Implementation

- [ ] Implement live snapshot marshal/unmarshal.
- [ ] Implement Host/App/server save/load surfaces.
- [ ] Add exact mid-game roundtrip + continuation regressions.
- [ ] Verify AI/human controller identity and pending decisions survive.

## Gate 4 - Follow-up QA + commit + close

- [ ] Repeat roundtrip at every supported phase.
- [ ] Verify malformed/old/partial saves fail atomically.
- [ ] Run full tests/vet/web checks and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.