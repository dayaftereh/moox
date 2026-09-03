# Built-in strategic AI baseline - Gate 1 evidence / architecture audit

Date: **2026-09-03**

Slice: **13 - Built-in strategic AI baseline**

Status: **Gate 1 complete; Gate 2 contract proposed; no AI implementation yet**.

Starting HEAD: `3dc5455` (`docs: audit post-milestone fidelity backlog`)

## Purpose

Slice 12 proves that the supported game can run from a real deterministic `NewGame(0x8009)` through Research, expansion, war, Troop Transport invasion, Empire elimination and an immutable conquest winner. Slice 13 must replace the scripted second player with a deterministic built-in controller without granting that controller hidden state, direct mutation access or private economy bonuses.

This slice is deliberately a **legal baseline AI architecture**, not a claim of exact original MOO2 AI parity.

## Original / historical AI evidence checked

The original Master of Orion II manual separates game difficulty from pure decision quality:

- Average states that all races develop at normal rates.
- Easy states that other races' production/research proceed more slowly and that they are more friendly.
- Impossible states that other races have significantly accelerated production/research and begin strongly hostile.
- The diplomacy section states that first-contact flavor depends on the opposing emperor's personality.

Local `docs/research/MOO2_GAME_REFERENCE.md` already turns that into an explicit architectural warning: original-like AI should separate decision policy/personality, difficulty bonuses, information available to AI, economic modifiers and ship-refit behavior. The same document lists exact diplomacy/personality tables plus exact AI difficulty bonuses/hidden rules as unresolved parity questions.

Gate-1 conclusion: **do not invent those unknown original tables inside Slice 13**. The baseline AI is a MOOX deterministic policy over Average/no-bonus rules. Original personalities and difficulty advantages remain later evidence-driven AI fidelity work.

Reference material checked:

- original MOO2 instruction-manual text mirrored by Manualzz / Manualshelf (Difficulty and Diplomacy sections);
- `docs/research/MOO2_GAME_REFERENCE.md`, especially section 17 AI fidelity and section 19 open parity questions;
- existing normalized race/economy/research data and all current Slice-12 authoritative command/session contracts.

## Current controller/runtime finding

`session.ControllerBuiltinAI` already exists alongside local/remote human and external/MCP AI controller identifiers, but non-test runtime code does **nothing** with it. The controller value is presently metadata only.

This is useful: Slice 13 can add a controller driver without changing the meaning of existing human seats.

## Stable decision boundaries that require a controller

### 1. Planning

The seat may submit one revision-bound `protocol.CommandBatch`. Current supported planning commands include:

- population assignment;
- Research selection;
- Building/Housing/Planetary Transformation/Freighter construction;
- Colony Ship, Outpost Ship, military Ship and Troop Transport construction;
- Population transfer;
- military design save;
- Fleet move/split/merge;
- Colony Ship colonization and Outpost deployment.

An empty batch is also legal when the AI intentionally has no action.

### 2. Immediate diplomacy

At the appropriate session revision, the seat may:

- declare war;
- offer peace;
- accept an incoming peace offer.

The baseline AI only needs a deterministic war/peace policy. Treaty/trade/tribute/espionage diplomacy is not part of the current game contract and stays deferred.

### 3. Post-resolution Colony Base decisions

`PhasePostResolution` can remain interactive when a Colony Base resolution is pending. The controller must choose the already-authoritative colonize/trash path when such a project exists.

The baseline policy can avoid Colony Base construction unless needed, but the driver must still be able to resolve the boundary so an AI game cannot stall if a pending decision occurs.

### 4. Encounters / BattleSession

The Host stops at `PhaseEncounters`. Participant seats receive player-filtered `battle.View` values and submit the same `battle.fire_beam` / `battle.end_activation` commands as a human.

The narrow tactical baseline is sufficient for Slice 13: choose the first legal Beam action deterministically by stable Ship/weapon/target identity, otherwise end activation. Tactical strategy fidelity is deferred.

### 5. Invasion decisions

At `PhaseInvasionDecisions`, only the attacker receives `InvasionOpportunity`. The controller must choose invade or decline. Baseline policy: invade with all eligible transports when an enemy Colony is the target.

### 6. Completed

No controller action. Completed games remain immutable/readable as established by Slice 12.

## Deterministic non-choice lifecycle gap found

Research breakthrough settlement is currently a `GameSession.CompleteResearchField(empireID, resolver)` call and is exercised directly by the Slice-12 headless test. It is **not exposed through Host/HTTP and is not a genuine strategic choice once the technology has already been selected**.

Gate-2 proposal: treat due Research completion as deterministic engine progression in the Host boundary driver, not as an AI-only mutation. The Host should settle due selected Research for every active Empire before advancing from PostResolution. Humans and AI then share identical behavior and only choose the next Research project through normal planning commands.

This is a lifecycle/transport gap discovered by the AI audit, not AI policy.

## No-cheat visibility audit

### Current problem

`session.PlayerView` currently provides:

- own Empire;
- own Colonies;
- seat/submission state;
- diplomatic stance/peace-offer summary;
- pending Colony Base resolutions;
- attacker-only Invasion opportunity.

`app.PlayerSnapshot` adds participant-filtered Battle views.

It does **not** currently provide the strategic Galaxy, own Ships/ShipDesigns/StrategicFleets/Outposts, or enemy strategic contacts. A useful AI cannot plan expansion or conquest from this view.

Giving the AI `ObserverView` or `*core.GameState` would expose enemy Treasury, Research, Colony internals, command submissions and other information unavailable to a human client. That is rejected for Slice 13.

### Gate-2 proposal: shared player-safe strategic projection

Add a canonical server/player projection used by **both built-in AI and the future Slice-15 HMI**. It should provide only what the current baseline player is allowed to use:

- stable Galaxy/System/Planet geometry and IDs needed to target strategic commands;
- own Ships, ShipDesigns, StrategicFleets, Outposts and Population transfers;
- public strategic contacts needed by the currently supported no-fog baseline: enemy Colony/Outpost owner+location and Fleet owner+strategic location/contact identity;
- no enemy Treasury, Research progress/choices, population jobs, construction queues, internal Colony economy, submissions or hidden telemetry.

MOOX does not yet implement exploration/sensor/fog-of-war state. Slice 13 therefore uses a **symmetric full strategic-contact baseline**: whatever strategic contact data the AI receives must be exposed identically to the human player projection. This is explicitly a temporary visibility baseline, not a claim of original MOO2 sensor fidelity.

## Authoritative legal-action/query audit

Already reusable server-authoritative query logic exists for:

- `AvailableResearchChoices`;
- `AvailableConstructionChoices` / `AvailableBuildingChoices`;
- pending Colony Base resolutions;
- attacker Invasion opportunity;
- participant Battle views.

Important missing query/projection coverage for a non-cheating planner:

- legal Fleet destinations under installed drive/fuel/supply rules;
- legal Colony Ship colonization targets;
- legal Outpost deployment targets;
- player-safe own strategic assets + public strategic contacts.

Gate-2 proposal: add a compact **player DecisionCatalog** generated from authoritative state/rules by Go. The AI consumes the catalog; it does not reimplement fuel range, supply, ownership, buildability or Research eligibility formulas.

Recommended catalog sections:

- Research choices;
- construction choices keyed by owned Colony;
- legal move destinations keyed by owned at-system Fleet;
- legal colonization/deployment target tuples keyed by civilian Fleet;
- diplomacy summary already in PlayerView;
- Colony Base resolution choices;
- Invasion opportunity;
- participant Battle views.

The server may inspect authoritative state to build the catalog. The AI itself may not receive that state.

## Proposed package / authority boundary

### `internal/ai`

Pure deterministic planner. It receives only versioned player-safe decision input and returns one next controller action. It must not import or receive `ObserverView` or mutable `*core.GameState`.

The planner owns **policy**, not rules:

- rank legal Research choices;
- rank legal construction choices;
- choose among already-legal movement/colonization targets;
- decide when to declare war/offer or accept peace;
- choose tactical target among player-visible Battle participants;
- choose invade/decline.

### App/Host AI driver

An orchestration layer identifies seats whose controller is `builtin_ai`, asks `internal/ai` for the next action and dispatches that action through the same public Host mutation surfaces used by humans:

- `SubmitTurn`;
- `SubmitImmediateCommand`;
- `SubmitBattleCommand`.

Deterministic non-choice progression (resolution, due Research completion, CompleteTurn) remains engine/Host responsibility, not AI authority.

The AI driver must not mutate `GameSession`, `GameState`, BattleSession, Treasury, Research progress or revisions directly.

## Baseline policy v1 - bounded strategy, not original parity

Policy decisions are deterministic with stable ascending-ID tie breaks and no AI-private RNG.

### Research

- if expansion/conquest is range-blocked, prioritize an available Fuel Cell / strategic-range technology;
- otherwise prefer the lowest-cost available frontier, then stable field/technology ID;
- never inspect opponent Research.

### Population / economy

- keep Colonies food-sustaining under the authoritative projected economy;
- emphasize Scientists while range/required strategic technology is missing;
- emphasize Workers when a required Ship/project is queued;
- do not inject BC/PP/RP or bypass maintenance.

### Construction / expansion

Priority when legal/needed:

1. Colony Ship if an immediately legal useful colonization target exists;
2. Outpost Ship when supply extension is needed to reach a useful frontier;
3. enough Troop Transports for a conquest attempt;
4. at least one supported military Fleet/Ship for attack authorization/encounters;
5. otherwise choose a deterministic legal economic project or idle/continue current queue.

### Fleet strategy

- colonize nearest legal unowned useful planet by deterministic distance/ID order;
- deploy Outposts along the shortest legal supply-extension chain when no Colony target is directly reachable;
- once invasion-ready, target the nearest visible enemy Colony;
- use public legal destination sets rather than recomputing fuel/supply legality.

### Diplomacy

- accept an incoming peace offer only when not invasion-ready and materially unable to threaten an enemy Colony;
- declare war once a military Fleet + sufficient Troop Transport capacity + legal route/contact to an enemy Colony exist;
- no hidden relation modifiers or fabricated hostility bonuses in baseline v1.

### Tactical

- if the active AI ship has a legal Beam attack, choose the first legal weapon/target by stable IDs;
- otherwise end activation.

### Invasion

- invade with all eligible Troop Transports against an enemy Colony;
- decline only if no eligible transport exists or the projected opportunity is no longer valid.

## Determinism / replay contract

Baseline AI v1 has no independent RNG. Given identical player-safe input, it must emit byte-equivalent controller actions.

Every accepted AI mutation therefore advances the same normal Session/App revisions, events and Host change sequence as the equivalent human command. Planner evaluation itself advances nothing and emits no gameplay event.

If future AI randomness is added, it must use explicit persisted RNG state and cannot consume the simulation RNG implicitly.

## Gate-3 regression fixtures proposed for Gate 2

### A. Pure planner determinism

For frozen decision snapshots, planning the same input repeatedly produces exactly the same action bytes and does not mutate the input.

### B. No-cheat projection

AI input contains no opponent Treasury, Research progress, population assignment, construction queue or unpublished command batch. The planner API has no ObserverView/GameState input path.

### C. Canonical AI-vs-AI completion

- real `NewGame(0x8009)`;
- current Small / Normal / Average / Tactical two-player Human+Darlok content baseline;
- both seats `ControllerBuiltinAI`;
- no direct post-NewGame state mutation;
- hard liveness cap: **1000 strategic turns**;
- completion must be authoritative `ResultConquest`;
- repeat from the same seed and compare exact completed snapshot/result/event history bytes.

On failure at the cap, dump deterministic diagnostics: turn/phase/revision, Colony ownership, Fleets and locations, active Research/progress, construction queues, diplomacy stance, current legal-action counts and the last controller actions.

### D. Human-vs-AI authority regression

Run a scripted human seat against the built-in AI through Host surfaces and verify that replaying the same human commands produces identical AI commands and final state. The purpose is transport/controller determinism, not a guaranteed human or AI winner.

### E. Tactical/invasion boundary regressions

Construct participant Battle and Invasion fixtures where an AI seat is required to act; verify the AI advances each boundary using only player-visible inputs and normal Host mutations.

## Gate-2 contract proposed

Approve all of the following before Gate 3 implementation:

1. Slice 13 implements **MOOX baseline AI v1**, not exact original MOO2 AI parity.
2. Average/no-bonus economy rules remain unchanged; difficulty bonuses/personality fidelity are deferred.
3. Built-in AI receives only a shared player-safe StrategicView + DecisionCatalog + participant Battle/Invasion projections; never ObserverView or mutable GameState.
4. Until a sensor/fog system exists, strategic contacts are symmetrically visible to human and AI, while enemy economic/research/queue internals remain hidden.
5. Authoritative Go queries enumerate Research/construction/movement/colonization/outpost legality; AI ranks legal options instead of duplicating rule formulas.
6. The AI planner is deterministic, versioned (`baseline_v1`) and has no private RNG in this slice.
7. App/Host orchestration dispatches AI choices through the same `SubmitTurn`, `SubmitImmediateCommand` and `SubmitBattleCommand` mutation paths as humans.
8. Due already-selected Research completion becomes deterministic Host progression for all controllers; it is not an AI-only mutation.
9. Baseline strategic policy is the bounded Research/Expansion/War/Tactical/Invasion policy documented above.
10. Canonical AI-vs-AI `0x8009` must finish within 1000 turns and reproduce exact completed bytes across repeated runs; Human-vs-AI plus boundary/no-cheat tests are required.
11. Exact MOO2 personalities, difficulty cheats/bonuses, stronger modern AI, broader races/settings, treaty/trade/espionage/leaders and deep tactical strategy remain deferred.

Gate 1 is complete when this contract is presented for approval. No Gate-3 AI implementation is authorized by this document alone.