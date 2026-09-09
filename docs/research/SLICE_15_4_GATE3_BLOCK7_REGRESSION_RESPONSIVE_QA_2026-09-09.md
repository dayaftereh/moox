# Slice 15.4 Gate 3 Block 7 - Regression, stale-client and responsive QA

Date: 2026-09-09

## Goal

Close the remaining Gate-3 browser/server regression and responsive-contract work after the rich decision, persistence and Battle Entry/Return blocks. The frozen Gate-2 contract requires stale authoritative rejections to remain visible, trigger a refetch without silent resubmission, preserve only compatible local planning intent, keep stale snapshots read-only while refreshing/reconnecting, expose an explicit Retry after refresh failure, and keep primary mobile workflow actions touch-safe at the 320 CSS-px floor.

## Stale/conflict recovery hardening

A Gate-3 audit found that Colony Base and Invasion already refetched after HTTP 409 stale-revision rejection, while End Turn and Diplomacy only displayed the error. `App.tsx` now uses one `refreshAfterConflict` path for Planning, Diplomacy, Colony Base and Invasion.

The common behavior is:

1. accept only structured `APIError` status 409 as a stale/conflict recovery case;
2. display the authoritative rejection;
3. fetch a fresh authoritative snapshot;
4. never resubmit the rejected mutation automatically;
5. for a rejected End Turn, retain the local planning draft only if the fresh phase is still Planning;
6. otherwise clear the incompatible planning draft and report the fresh phase.

`loadSnapshot` now returns the fresh snapshot so recovery can make that compatibility decision from the authoritative phase/change sequence.

## Planning-preview error separation

Browser QA exposed a second authority-presentation bug: after a 409 conflict was correctly refetched, a successful background planning-preview request called the global `setError('')` and erased the still-important command rejection. Planning-preview failures now use their own `planningPreviewError`; successful previews clear only that preview-specific state and can no longer hide an unrelated server/action error.

## Refresh-failed and Retry

The existing lifecycle already retained the last known snapshot while `lifecycle != synced`, and `mutationLocked` made that stale state read-only. The missing Gate-2 requirement was a reliable visible Retry action.

The game surface now renders a lifecycle-driven `Refresh failed` / `Aktualisierung fehlgeschlagen` notice whenever a previous snapshot exists and lifecycle is `refresh-failed`, independent of which UI path initiated the failed refresh. It explains that the shown state is stale/read-only and offers an explicit Retry button.

## Managed-browser conflict QA

A temporary port-7181 production-build server was created from an actual current player snapshot. Its first End-Turn POST deliberately returned HTTP 409 with `stale_revision`; the next snapshot was revision/change 2. A debug endpoint counted writes and reads.

Observed after clicking End Turn:

- stale rejection remained visible in the global error notice;
- header status reported that the client refreshed to change 2 and did not resubmit;
- lifecycle returned to `synced` after the fresh snapshot;
- mock counters showed `submitCount = 1` and `conflictTriggered = true`;
- therefore the rejected command was not silently reissued.

## Managed-browser refresh-failure QA

The same temporary production-build harness could fail exactly one snapshot GET with HTTP 503.

Observed flow:

1. trigger one snapshot failure from the real `Snapshot aktualisieren` control;
2. lifecycle becomes `refresh-failed`;
3. the old game state remains visible;
4. End Turn is disabled because mutation is locked;
5. a visible `Aktualisierung erneut versuchen` button appears together with the stale/read-only explanation;
6. click Retry;
7. fresh snapshot succeeds, lifecycle returns to `synced`, errors/Retry disappear and End Turn becomes active again.

## 320 CSS-px responsive QA

Managed Chrome cannot physically shrink below its desktop minimum in this environment, so the established project QA technique was used: evaluate the production CSS at the 320-px media-rule set and constrain the tested rich component to exactly 320 CSS px, then inspect component scroll width and element geometry.

### Battle Entry

- root client width: 320 px;
- root scroll width: 320 px;
- no non-SVG horizontal overflow;
- attacker/defender comparison stacks to one column;
- initial Tactical CTA exposed a real Gate-2 touch-target regression: the compact `.game-shell` rule reduced the primary button to 32 px.

The fix is mobile-only: `.game-shell .button-primary` and `.button-secondary` regain a minimum 44-px height below 700 px; compact ghost/icon controls remain unchanged.

After the fix:

- `Taktischen Kampf betreten`: 44 px high;
- Battle root remains 320/320 with no horizontal overflow.

### Research

- canonical Research dialog constrained to 320 px has matching client/scroll width and no horizontal overflow;
- technology fields remain intentionally compact as previously approved; the 44-px rule is applied to primary workflow actions, not every compact list/info affordance.

### Colony construction

- construction workspace: 320 px client/scroll width;
- no horizontal overflow;
- project catalog items remain readable;
- primary `Einplanen` action is 44 px high.

### Persistence/session controls

- More/session content remains width-safe under the mobile rule set;
- `Snapshot aktualisieren` is 44 px high;
- in-game menu `Spiel speichern` and `Spiel laden` items are each 44 px high and require no drag/drop.

## Existing server-regression matrix

The Gate-3 requirement asks for browser/server regressions across critical decisions, encounter transitions and persistence. The repository already contains focused server/session regression suites, so Block 7 records and re-runs them instead of duplicating equivalent tests:

- `internal/server/persistence_test.go`
  - HTTP persistence disabled-by-default guard;
  - Export -> Import -> Restore byte-preserving roundtrip;
  - invalid-save bad-request mapping.
- `internal/session/live_snapshot_test.go`
  - partial Planning submission roundtrip;
  - strict compatibility/boundary rejection;
  - Invasion-decision roundtrip/continuation;
  - interactive post-resolution roundtrip/continuation;
  - active Tactical session roundtrip/continuation.
- `internal/session/strategic_encounter_test.go`
  - staged/parallel result ordering;
  - atomic and retryable final-result failure;
  - sequential encounter waves and no-encounter compatibility;
  - retreat/blockade/observer/replay coverage.
- `internal/session/tactical_battle_test.go`
  - Tactical Battle -> strategic handoff;
  - atomic/retryable terminal-command continuation failure;
  - returned pre-encounter state;
  - stable parallel completion ordering;
  - explicit unsupported/manual Battle lifecycle.
- `internal/session/invasion_test.go` and `colony_base_test.go`
  - stale revision/non-mutation and mandatory blocking-decision outcomes.
- `internal/session/resolution_summary_test.go`
  - participant-safe recent resolution projection and participant/age filtering.
- stale server/session commands also have explicit conflict coverage in Planning, Diplomacy, Invasion and Battle command tests.

Browser-managed QA from Blocks 3-7 supplies the client-side transition and presentation coverage that the Go tests intentionally do not own.

## 15.3 visual / motion-audio carry-forward

The frozen 15.3 visual baseline explicitly carries `StarArt`, `OrbitalBodyArt`, `BuildingArt`, procedural ship visuals and typed `GameIcon` semantics into 15.4-15.6. Current rich surfaces already reuse those primitives extensively (`App.tsx`, `StrategicViews.tsx`, `BattleRouteView.tsx`). The same 15.3 freeze explicitly defers motion/effects/audio polish, so Gate 3 does not invent new audio or animation timing. No decorative asset or presentation timing became authoritative for simulation or command legality.

## Result

Gate 3 now has implementation and QA evidence for the complete frozen 15.4 contract. Gate 4 can concentrate on whole-slice closure and the handoff into Slice 15.5 rather than reopen individual feature blocks.
