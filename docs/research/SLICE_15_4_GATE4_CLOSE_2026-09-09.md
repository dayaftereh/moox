# Slice 15.4 Gate 4 - QA close and Slice 15.5 handoff

Date: 2026-09-09

## Close decision

Slice 15.4 - Rich gameplay decisions / persistence UX is **closed**. Gates 1-4 are complete. The interactive Tactical battlefield remains intentionally outside this slice and passes to Slice 15.5.

## Gate-4 closure matrix

### Supported non-tactical browser phases

The supported interactive phases have direct managed-browser evidence across the Slice-15.4 implementation blocks:

- Planning / End Turn and Built-in-AI continuation: `SLICE_15_4_GATE3_ENDTURN_BUILTIN_AI_BLOCKER_2026-09-09.md`;
- Save / Load / Import / Restore: `SLICE_15_4_GATE3_BLOCK2_SAVE_LOAD_RESTORE_UX_2026-09-09.md`;
- Colony Base mandatory post-resolution choice: `SLICE_15_4_GATE3_BLOCK3_COLONY_BASE_UX_2026-09-09.md`;
- Invasion blocking decision: `SLICE_15_4_GATE3_BLOCK4_INVASION_UX_2026-09-09.md`;
- Resolution summaries, Research Breakthrough and Victory: `SLICE_15_4_GATE3_BLOCK5_RESOLUTION_SUMMARY_RESEARCH_RESULT_2026-09-09.md`;
- canonical Research chooser and source-backed Technology information: `SLICE_15_4_GATE3_RESEARCH_UX_TECH_INFO_2026-09-09.md` plus the later source-backed-description update;
- Battle Entry / Return and strategic-continuation lock: `SLICE_15_4_GATE3_BLOCK6_BATTLE_ENTRY_RETURN_2026-09-09.md`;
- stale/conflict/refetch/retry and responsive contract: `SLICE_15_4_GATE3_BLOCK7_REGRESSION_RESPONSIVE_QA_2026-09-09.md`.

Diplomacy remains an authoritative immediate-command surface during legal Planning state and participates in the same stale-conflict recovery path; no new diplomacy simulation was added by 15.4.

### Encounter -> Tactical handoff authority

The Battle route consumes participant-safe authoritative `BattleView` / `BattleSpec` data and never owns battle legality in React. It preserves:

- stable Battle ID and system context;
- public participant identity;
- visible fleet/ship context;
- defender-colony context when projected;
- supported Tactical spec or explicit unsupported reason;
- strategic blocking while a human-participant encounter is unresolved;
- participant-safe `battle_completed` return summary even after the live BattleView is released.

This is sufficient for Slice 15.5 to replace the current Tactical-shell placeholder with interactive battlefield controls without changing the strategic entry/return contract.

### Persistence / reconnect / stale-client behavior

Gate-2 lifecycle semantics were exercised end to end:

- last known state remains inspectable during reconnect/refresh failure;
- mutation is disabled unless lifecycle is synchronized;
- snapshot invalidation causes refetch;
- HTTP 409 rejection remains visible, refetches fresh authority and is never auto-resubmitted;
- compatible local Planning drafts may survive for review, while incompatible phase changes drop them explicitly;
- refresh failure exposes a visible Retry and successful Retry returns the client to synchronized writable state;
- save/export/import/restore roundtrips and compatibility/boundary rejection are covered by server/session regressions.

### Rejected commands and saves

- stale command QA proved exactly one submit was sent despite refetch;
- stale Invasion/Planning/Diplomacy server regressions prove rejected revisions do not mutate authoritative state;
- invalid persistence input maps to a clear bad request and does not replace live authority;
- Restore remains an explicit confirmation path rather than a silent overwrite.

### Regression / build gate

Immediately before Gate-4 close, the Slice-15.4 Block-7 implementation passed:

- `go test ./...`;
- `go vet ./...`;
- `npm run build`;
- `git diff --check`;
- staged diff check before commit `7d61683`.

The full run includes session encounter, Tactical handoff, persistence, Invasion, Colony Base, resolution-summary and stale-revision regression suites.

## Responsive close

The frozen 320 CSS-px floor remains intact for the rich gameplay surfaces. Final Block-7 QA found and fixed one mobile touch-target regression caused by the compact game-shell button rule. Primary and secondary mobile workflow actions now have a 44-px minimum height while compact ghost/info affordances remain visually dense.

## Handoff to Slice 15.5

Slice 15.5 can start from the stable route:

`#/game/<gameID>/battle/<battleID>`

The current `Taktischen Kampf betreten` action proves that a supported Tactical spec has arrived. Slice 15.5 owns what happens inside that shell:

- battlefield rendering and camera/viewport behavior;
- ship selection/activation;
- movement controls and legal destination presentation;
- target selection;
- weapon/beam firing controls;
- command submission and authoritative invalidation/refetch during Tactical play;
- terminal battle transition back through the already-frozen Slice-15.4 return surface.

No strategic rule, Battle result or Tactical damage/movement computation should migrate into React during that work.

## Final status

**Slice 15.4: CLOSED - Gates 1, 2, 3 and 4 complete.**

Next queued slice: **15.5 - Interactive Tactical Combat**.
