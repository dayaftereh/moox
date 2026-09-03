# Open slice 13 - Built-in strategic AI baseline

Status: **Gate 1 complete; Gate 2 approval pending; no AI implementation yet**.

Opened: **2026-09-03**

Planned specification: `docs/slices/PLANNED_13_BUILTIN_STRATEGIC_AI_BASELINE.md`

Permanent evidence: `docs/research/BUILTIN_STRATEGIC_AI_BASELINE_2026-09-03.md`

Starting HEAD: `3dc5455` (`docs: audit post-milestone fidelity backlog`)

## Gate state

- [x] Gate 1: repository/controller audit + original-AI evidence + deterministic no-cheat AI contract.
- [ ] Gate 2: accept/freeze exact AI visibility/decision/lifecycle contract.
- [ ] Gate 3: implementation + deterministic AI-driven match regressions.
- [ ] Gate 4: independent QA + commits + close.

## Scope guard

Gate 1 is research and architecture analysis only. Do not implement the built-in AI driver before Gate 2 freezes the contract.

In scope for eventual implementation: deterministic built-in AI using the same authoritative legal Session/App command surfaces as human seats; Research, Colony/Construction, expansion/Fleet movement, war/peace baseline, Battle/Encounter/Invasion decisions needed for the currently supported two-player lifecycle; AI-vs-AI and Human-vs-AI regressions; explicit no-cheat/no-direct-state-mutation boundary.

Deferred: exact original MOO2 AI personality/difficulty fidelity, difficulty cheats, stronger modern search/planning AI, broader New Game settings/races, treaty/trade/espionage/leader depth.

## Gate 1 checklist

- [x] Re-check repo/session/controller/command surfaces and canonical Slice-12 match.
- [x] Inventory every stable boundary that can require a seat decision.
- [x] Inspect available original AI evidence and separate proven original behavior from MOOX baseline policy.
- [x] Define deterministic AI input visibility and no-cheat contract.
- [x] Define minimal legal-action/query gaps that must be exposed by Go authority rather than recomputed by AI.
- [x] Define AI-vs-AI completion fixture/turn cap and failure diagnostics.
- [x] Present exact Gate-2 AI contract before implementation.