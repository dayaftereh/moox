# Planned Slice 17.0 - Ship Designer / Triangle Reference Test Harness

Status: **active; Gates 1-3 complete / Gate 4 ready**.

## Objective

Create durable development-only test scenarios so Ship Designer, Construction and later Tactical/Triangle work can be validated without manually researching late-game technologies or clicking through dozens of turns.

## Existing foundation

- `-reference-games` is already development-only.
- `game-triangle-2pc` is already a deterministic 3-player Triangle reference game.
- `NewReferenceTriangleGame` uses normal rules, technology initialization, starting assets, economy finalization and GameState validation.
- normal designer, construction, turn and battle infrastructure already has deterministic tests.

## Required test modes

### A. Durable Triangle Designer Lab family

The existing `game-triangle-2pc` is the canonical durable reference-game foundation, not a disposable one-off fixture. Slice 17.0 extends that same Triangle concept with deterministic technology-profile variants available before the first Planning phase / on Turn 1:

- baseline Triangle: the existing movement/contact reference state;
- Triangle Mid-Tech: curated deterministic mid-game technology state;
- Triangle All-Tech: curated deterministic all-supported technology state for the current runtime / active Slice-17 scope.

Designer and Construction QA should reuse this Triangle family instead of creating parallel `game-designer-*` throwaway worlds. The variants keep the same three-player 2-pc Triangle topology and ordinary GameSession/Host authority; only reference bootstrap state differs.

The All-Tech profile must allow direct QA of things such as Doom Star Construction, Reinforced Hull, Heavy Armor, advanced computers, drives, shields, armor and the weapon/special set completed by the active 17.x child. `All-Tech` is a reference-lab label, not a claim that every granted technology already has complete Tactical gameplay support.

Technology is granted only during reference-game bootstrap. After bootstrap, normal production/design legality code is used.

### B. Construction Lab

Allow a tester to save a normal design, queue it through normal Colony Construction and avoid manual repeated turn clicks.

Preferred development-only controls:

- advance 1 normal turn;
- advance N normal turns within a safe bound;
- advance until current construction completes, still invoking the normal turn resolver each iteration.

No production is gifted and no construction formula is bypassed.

### C. Combat / Triangle Lab

Allow durable deterministic scenarios containing already-built concrete Ships created from ordinary ShipDesignSpec snapshots so Tactical/Triangle tests can begin immediately.

Examples:

- Frigate vs Frigate;
- Cruiser Beam vs Cruiser Beam;
- ordnance test pair;
- Heavy Armor / Reinforced Hull defensive target;
- late-game Battleship/Titan/Doom Star fixtures as those designs become supported.

These fixtures must not invent a second ship schema.

## Scope guard

17.0 is QA infrastructure only. It does not implement new hull/component/weapon/special mechanics itself.

Reference controls must not appear in ordinary games when the server is not started with the explicit development reference-game capability.

## Gate 1 - harness audit

- [x] Audit `-reference-games`, standard reference game and Triangle reference-game registration.
- [x] Audit technology bootstrap APIs/state and define safe deterministic technology grants.
- [x] Audit normal Construction/Turn APIs for bounded automated advancement.
- [x] Audit BattleSession/Triangle fixture entry and built-Ship snapshot requirements.
- [x] Define stable reference scenario IDs and ownership.
- [x] Define strict dev-only exposure/authorization boundary.

## Gate 1 evidence

- `docs/research/SLICE_17_0_GATE1_REFERENCE_HARNESS_AUDIT_2026-09-18.md`
- Gate 1 found no architecture blocker. The harness must wrap the normal technology/design/construction/turn/Ship/Battle paths and must remain double-gated by explicit server + per-game reference capability.

## Gate 2 - harness contract freeze

- [x] Freeze durable Triangle scenario matrix: baseline, Turn-1 Mid-Tech and Turn-1 All-Tech/all-supported, plus any explicitly justified combat-only fixtures.
- [x] Freeze exact Mid-Tech and All-Tech technology materialization per Triangle scenario.
- [x] Freeze bounded turn-runner semantics and stop conditions.
- [x] Freeze how built Ships are materialized without bypassing design snapshot validation.
- [x] Freeze HTTP/HMI exposure of reference-only controls.
- [x] Freeze deterministic seeds and persistence expectations.

## Gate 2 evidence

- `docs/research/SLICE_17_0_GATE2_REFERENCE_HARNESS_FREEZE_2026-09-18.md`
- Frozen: durable baseline/Mid-Tech/All-Tech Triangle family, exact profile semantics/fingerprints, same seed/topology, 1..25 bounded Advance N, 512-turn construction runner ceiling, normal Host/turn authority, built-Ship fixture snapshot rules, double dev authorization, HTTP/HMI contract and non-escalating persistence/import behavior.

## Gate 3 - implementation

- [x] Add reusable reference-game bootstrap helpers.
- [x] Add curated technology-profile reference scenarios.
- [x] Add normal-resolver turn advancement controls for reference games only.
- [x] Add construction-completion runner with hard iteration bound.
- [x] Add reusable combat-ready Ship snapshot fixtures.
- [x] Expose clear reference-game identity in HMI.
- [x] Add unit/server/browser regressions proving reference-only isolation.

## Gate 3 evidence

- `docs/research/SLICE_17_0_GATE3_IMPLEMENTATION_2026-09-18.md`
- Implemented durable Triangle profiles, trusted reference metadata, bounded normal-turn/construction controls, shared immutable Ship fixture materialization, dev-only HTTP security, persistence isolation, bilingual responsive HMI and unit/server/browser regressions.

## Gate 4 - QA + close

- [ ] Turn-1 Triangle All-Tech Designer Lab can open Doom Star/high-tech design QA without research grinding once downstream 17.x mechanics exist.
- [ ] Construction Lab builds through normal production/turn resolution.
- [ ] Combat Lab ships use the same immutable ShipDesignSpec as ordinary constructed ships.
- [ ] Reference controls are unavailable in normal server mode.
- [ ] Equal scenario seed/profile produces identical state.
- [ ] Save/restore works for reference scenarios where persistence is enabled.
- [ ] Full tests/vet/web/browser checks and `git diff --check` pass.

## Exit criterion

A developer/tester can load deterministic technology-rich designer, construction and combat scenarios in seconds while all post-bootstrap gameplay continues through the same authoritative code paths used by normal games.
