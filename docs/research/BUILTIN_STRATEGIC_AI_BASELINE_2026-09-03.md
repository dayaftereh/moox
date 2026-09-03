# Built-in strategic AI baseline - Gate 1 evidence / architecture audit

Date: **2026-09-03**

Slice: **13 - Built-in strategic AI baseline**

Status: **Gates 1-3 complete; Gate 2 frozen by user approval; Gate 4 independent QA pending**.

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

## Gate 2 design review - 2026-09-03

The Gate-1 contract was re-checked against the actual Host locking/change-sequence model before freezing it. One important refinement is required: a built-in AI action must **not recursively call public `Host.Submit*` methods from inside `hostedGame.mutate()`**, because those methods reacquire the same hosted-game mutex and would make receipt/change-sequence semantics ambiguous. The correct authority boundary is therefore:

- `internal/ai` owns pure deterministic policy only;
- `internal/app` / `hostedGame` owns controller orchestration;
- orchestration invokes the same **public `GameSession` mutators and validators** that back human Host calls (`SubmitTurn`, `ResolveDiplomacyCommand`, `ResolveColonyBaseCommand`, `SubmitBattleCommand`, `ResolveInvasionCommand`);
- orchestration never edits `GameState`, Battle runtime, Treasury, Research or revisions directly;
- existing Host `mutate()` semantics remain the outer transaction: one initiating human/automation operation may contain automatic resolver/AI substeps and returns/publishes the final interactive-boundary revision in one Host change-sequence update.

This preserves human-equivalent game authority without Host re-entrancy.

### Player-safe projection and DecisionCatalog ownership

The safe projection belongs with Session authority, not inside the AI package. Gate 3 should add a player-facing decision snapshot that can later be reused unchanged by Slice 15 HMI:

- strategic geometry/IDs required to target commands;
- own Colonies, Outposts, Ships, ShipDesigns, Strategic Fleets and transfers;
- symmetric public strategic contacts for the current no-fog baseline;
- Diplomacy, pending Colony Base, Invasion and participant Battle state;
- **no** enemy Treasury, Research state/progress, population assignment, construction queue, unpublished submissions or observer telemetry.

A Session/rules-backed `DecisionCatalog` enumerates authoritative legal options. It must include at least:

- available Research choices;
- available Construction choices per owned Colony;
- server-projected Population-management choices/economic previews sufficient for the baseline AI to avoid duplicating food/output formulas;
- legal Fleet destinations;
- legal Colony Ship colonization targets;
- legal Outpost deployment targets;
- legal immediate Diplomacy actions from the currently supported war/peace model;
- legal Colony Base resolution actions;
- legal Battle actions for the active participant/Ship using BattleSession validation;
- legal Invasion/decline options from the projected opportunity.

The catalog may inspect authoritative state internally. The planner may not.

### Automatic-controller lifecycle

The existing `hostedGame.driveToInteractiveBoundary()` is the correct integration concept and should be extended, rather than creating an independent parallel game loop.

At `planning`:

- built-in seats without a submission are planned/submitted in stable Seat-ID order;
- if any active human seat is still unsubmitted, stop at the human boundary;
- if all active seats are AI or have submitted, continue through normal strategic resolution.

At `post_resolution`:

- settle due already-selected Research in stable Empire-ID order using existing `CompleteResearchField`; each completion keeps its normal Session revision/event;
- resolve pending Colony Base choices if their owner is built-in AI;
- if a human decision remains, stop;
- otherwise `CompleteTurn()` normally.

At `encounters` / tactical Battles:

- if the currently required participant action belongs to built-in AI, choose one legal catalog action and submit it through `GameSession.SubmitBattleCommand`;
- if a human action is required, stop.

At `invasion_decisions`:

- act only when the attacker is built-in AI; otherwise stop for the human.

At `completed`: stop permanently.

For Human-vs-AI, these automatic substeps remain grouped inside the initiating Host mutation and the existing one-change-sequence/final-revision receipt model. Session revisions/events still record the individual authoritative state transitions.

For all-AI regression and future automation, add an explicit Host automation entrypoint that advances **at most one strategic turn (or until completion / a non-AI boundary)** per call. The canonical AI-vs-AI test calls that entrypoint repeatedly, which avoids a single unbounded Register/HTTP operation.

## Gate-2 contract proposed for approval

1. **AI identity:** Slice 13 implements `baseline_v1`, a MOOX deterministic legal AI baseline, not exact original MOO2 AI parity.
2. **No difficulty cheats:** Average/no-bonus economy remains unchanged. Original difficulty bonuses, personalities and hidden AI rules are deferred.
3. **No-cheat input:** Built-in AI receives only the shared player-safe strategic/decision projection plus participant Battle/Invasion data. It never receives `ObserverView` or mutable `*core.GameState`.
4. **Visibility baseline:** Until exploration/sensor/fog state exists, strategic contacts are symmetrically visible to Human and AI; enemy economic/Research/queue/submission internals remain hidden.
5. **Rules stay authoritative:** Session/rules code generates the `DecisionCatalog` for Research, Construction, Population previews, movement, colonization, Outposts, Diplomacy, Colony Base, Battle and Invasion. AI ranks legal options; it does not duplicate rule formulas.
6. **Pure deterministic planner:** `internal/ai` is versioned `baseline_v1`, uses stable tie-breaking, mutates no input and has no private RNG in Slice 13.
7. **Host orchestration without re-entrancy:** `internal/app` drives built-in controllers under the hosted-game lifecycle and invokes public `GameSession` command mutators directly; it does not recursively call public Host endpoints and never directly mutates state/revision/event internals.
8. **Controller-neutral Research settlement:** due already-selected Research is deterministic Host progression for every controller in PostResolution, stable by Empire ID, before turn completion; Humans and AI only choose the next Research project through normal planning commands.
9. **Transaction semantics:** one initiating Host mutation may contain automatic AI/resolver substeps and produces one final Host change-sequence notification/receipt, preserving current Host behavior; individual Session revisions/events remain authoritative. An explicit bounded automation entrypoint advances at most one all-AI strategic turn per call.
10. **Baseline policy:** deterministic Research/range-first expansion, food-safe population management, legal construction/Colonization/Outpost supply extension, military+Transport readiness, bounded war/peace policy, first-legal tactical Beam action, and invade-with-all-eligible-transports behavior as documented in Gate 1.
11. **Canonical liveness fixture:** real `NewGame(0x8009)`, current Small/Normal/Average/Tactical Human+Darlok content, both seats `ControllerBuiltinAI`, no post-NewGame direct state mutation, hard cap **1000 strategic turns**, authoritative conquest completion required.
12. **Exact determinism:** repeat the canonical AI-vs-AI run from the same seed and require exact completed snapshot/result/event-history bytes; add pure-planner determinism, no-cheat projection, Human-vs-AI authority, Research settlement, Battle and Invasion boundary regressions.
13. **Explicitly deferred:** exact MOO2 personalities/difficulty advantages, modern stronger AI/search, broader races/settings, treaty/trade/espionage/leaders, sensor/fog fidelity and deeper tactical strategy.

### Gate 2 status

The design review contract above was **explicitly accepted by the user on 2026-09-03** and is frozen for Slice 13. Gate 3 implementation follows that contract; broader AI/tactical/race work remains deferred.

## Gate 3 implementation results - 2026-09-03

Gate 3 now implements the first legal autonomous strategic controller over the already-supported MOOX lifecycle. The implementation preserves the Gate-2 no-cheat boundary: the planner never receives ObserverView or mutable GameState and never writes authoritative state directly.

### Player-safe decision surface

- Added a deep-cloned PlayerDecisionView / StrategicView in Session authority.
- Own Empire/Colonies/Ships/ShipDesigns/Fleets/Outposts/transfers are projected directly.
- Foreign strategic presence is reduced to symmetric Colony/Outpost/Fleet contact identity/location; enemy Treasury, Research state/progress, population jobs, construction queues, Ship composition, unpublished submissions and observer telemetry are absent.
- Galaxy geometry is shared, while occupancy IDs and blockade internals are stripped from the raw Galaxy projection and represented only through safe contacts.
- Added deterministic DecisionCatalog sections for Research, Construction, authoritative Population/economy previews, Fleet movement (including legal combat subset moves), Colonization, Outpost deployment, current war/peace actions, Colony Base, participant Battle actions and Invasion.

### Deterministic baseline_v1 planner

- Added internal/ai as a pure policy package with PolicyVersion baseline_v1 and no private RNG.
- Research prioritizes the expansion-range chain 51 -> 106 -> 98 -> 108 -> 194 when available.
- Population policy chooses only server-projected assignments and keeps Food safe before optimizing Research/Production.
- Expansion consumes legal Colony/Outpost targets and server-authoritative movement choices; it does not reproduce fuel/supply formulas.
- Military policy creates/uses supported combat and Troop Transport assets, declares war only through normal Diplomacy commands, and invades using the authoritative InvasionOpportunity.
- Tactical policy selects the first server-validated Beam action for supported Slice-07 1-vs-1 tactical encounters.

### Host/controller lifecycle

- hostedGame records controller identity and drives builtin_ai seats at Planning, PostResolution, Encounters and Invasion boundaries.
- AI actions invoke the same public GameSession mutators used underneath human Host calls; there is no recursive Host.Submit* call and no direct GameState/revision/event write.
- Due selected Research is settled controller-neutrally in stable Empire-ID order during PostResolution using the existing CompleteResearchField path.
- Added Host.AdvanceAutomation: an all-builtin game advances by at most one strategic turn per call, making long autonomous tests bounded and observable.
- CreateGame and HTTP New Game accept optional per-seat controller assignments; omitted assignments remain local_human for backward compatibility.

### Tactical scope findings preserved rather than widened

The first live AI-vs-AI run exposed three existing Slice-07 Tactical scope guards: multiple combat Ships, civilian Fleet context and Colony/planet-defense context. Slice 13 deliberately did **not** broaden Tactical Combat. Instead baseline_v1 uses existing legal strategic movement/splitting to keep the canonical liveness path inside the already-supported strategic/invasion contract:

- combat Fleet moves may use an authoritative one-Ship subset, relying on existing deterministic split semantics;
- starting Colony Ships are staged away from home before the conquest phase;
- starting combat Fleets patrol away from the capital so the final Colony can be taken without pretending that unsupported Colony-defense Tactical exists;
- combat movement precedes Troop Transport movement, preventing unsupported civilian Battle context;
- the actual Tactical AI action path remains separately regression-tested with a supported 1-vs-1 fixture.

This is a Slice-13 baseline policy accommodation, not a claim that those Tactical contexts are complete.

### Autonomous completion proof

A real NewGame with seed 0x8009, current Small/Normal/Average/Tactical Human+Darlok content and both seats ControllerBuiltinAI now completes without direct post-NewGame state mutation. The diagnostic run completed at **Turn 430 / Revision 871** with:

- Result kind: conquest;
- Winner Empire: 2 (Human);
- Winner Seat: 1;
- Eliminated Empire IDs: [3].

The committed integration regression runs the complete autonomous match twice and requires exact completed-session snapshot bytes. That includes authoritative state, result and event history. The exact replay test is green.

### Gate-3 regressions added

- complete canonical AI-vs-AI match twice with exact completed bytes and <=1000-turn liveness cap;
- Human-vs-builtin-AI automatic progression back to each Human Planning boundary, repeated exactly;
- controller-neutral Research completion event presence;
- DecisionView no-cheat, deterministic-byte and deep-copy invariants;
- authoritative DecisionCatalog non-mutation/determinism and food-safe Population preview;
- pure planner deterministic/non-mutating behavior;
- supported Tactical Battle action catalog + planner selection;
- HTTP New Game controller assignment projection;
- Invasion behavior is exercised inside the complete autonomous conquest match.

Gate 3 is complete. Gate 4 must independently rerun the broad QA matrix before final closure.
