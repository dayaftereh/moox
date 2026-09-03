# Post-Slice-12 fidelity / depth backlog audit - 2026-09-03

## Purpose

Slice 12 closed the first supported deterministic headless match lifecycle. This audit re-baselines the repository after that milestone and separates three different kinds of remaining work:

1. **playability blockers** - systems required to turn the headless lifecycle into a usable single-player game;
2. **runtime breadth gaps** - existing narrow baselines that deliberately accept only a small subset of MOO2 inputs/content;
3. **fidelity/depth gaps** - systems that are present structurally but remain materially shallower than the original game.

No implementation slice is opened by this audit.

## Repository state at audit start

- Branch: `main`.
- Worktree: clean.
- Slices 01-12: closed.
- `_OPEN_*.md` markers: zero.
- First full lifecycle baseline: implementation/evidence commit `90b83d8`, closure commit `fcd15ce`.
- Core state schema: 23.
- Economy ruleset schema: 8.
- Canonical end-to-end seed: `0x8009`.

## Important re-baseline finding

`docs/PROJECT_STATUS.md` still contained a pre-Slice-12 "What is not implemented yet" list. Several entries in that list are now demonstrably stale: Research progression, strategic Fleet movement/colonization, diplomacy war/peace, tactical combat baseline, invasion/conquest and elimination victory are implemented. This audit replaces that historical list with a current gap classification.

## What is already proven end-to-end

The current runtime can now execute, through authoritative public session/application surfaces:

- deterministic two-Empire New Game creation;
- Colony Economy, Population/Cohorts, Food/Freighters, Research and Construction;
- Colony Ship and Outpost expansion, supply range and strategic Fleet movement;
- military Ship/Fleet persistence plus merge/split, Command Points and maintenance;
- reciprocal war plus directional peace offer/accept baseline;
- hostile encounter -> BattleSession handoff;
- a narrow deterministic tactical Beam-combat vertical slice;
- Troop Transport invasion, Ground Combat and Colony conquest;
- zero-normal-Colony Empire elimination and immutable conquest winner;
- HTTP/WebSocket player/observer projections and minimal React HMI proof;
- exact completed-session save/reload and same-seed/same-command replay equality.

That means the project is no longer blocked on *having a complete match lifecycle*. The next blockers are about making that lifecycle independently playable, broader and more MOO2-complete.

## Audit criteria

Each backlog domain is ranked using five questions:

- **Player value** - does it immediately make a match more usable, varied or recognizable?
- **Dependency leverage** - does later work become easier/safer after this exists?
- **Runtime gap severity** - is the feature absent, or only narrow?
- **Evidence readiness** - how much local normalized/original evidence already exists?
- **Scope isolation** - can it be implemented behind current authoritative command/state boundaries without reopening completed cores?

Priority bands:

- **P0** - next milestone blocker;
- **P1** - high-value breadth/depth immediately after the next milestone;
- **P2** - important fidelity work with useful prerequisites still missing;
- **P3** - presentation, optional transports or late-game breadth that should not drive the next tranche.

## Gap matrix

| Domain | Current runtime boundary | Priority | Audit conclusion |
| --- | --- | --- | --- |
| Built-in AI | `ControllerBuiltinAI` exists as a valid seat type, but there is no non-doc runtime implementation of `builtin_ai`. | **P0** | Largest single-player blocker. Add a deterministic, rules-legal command producer using the same public surfaces as humans; no direct state mutation or hidden economy cheats. |
| Live save/resume | Core state round-trip exists and Slice 12 adds completed-session snapshots, but there is no in-progress GameSession/hosted-game save/load surface. | **P0** | Required before a real long-running browser match. Persist stable interactive boundaries, controller identity, RNG/revision/events, Battles/Invasion and result. |
| Playable browser HMI | Current React proof has New Game, snapshot/result, invasion, diplomacy, participant-battle and one concrete command panel, but no real galaxy/colony/research/construction/fleet workflow. | **P0** | After AI+persistence, connect the existing command surfaces into a human-vs-AI strategic UI without adding gameplay authority to React. |
| New Game breadth | Runtime explicitly supports only `small`, `normal` age, `average` tech, tactical combat, exactly two players and exactly Human+Darlok. | **P1** | Strong breadth gap. Expand only after AI exists so added seats/races/settings can be exercised autonomously and deterministically. |
| Preset races / governments | Ruleset contains 13 normalized preset races and existing Race/Government modifiers, but New Game exposes only Human/Darlok and custom-race construction is absent. | **P1** | Fold preset-race runtime breadth into New Game breadth; keep full custom race designer as a later dedicated fidelity slice. |
| Ship designer | Persisted design model exists, but current runtime explicitly allows only the Slice-07 minimal design: supported hull + slot 0 + one Laser Cannon; weapon modifiers remain future work. | **P1** | Excellent post-playability depth slice. Expand hull/components/weapon mounts/space/cost before deeper tactical weapons. |
| Tactical combat | Current tactical command surface is only `battle.fire_beam` and `battle.end_activation`; internal-system overflow/selection is explicitly unsupported. | **P1/P2** | Deepening should follow Ship Designer data/model breadth. Add movement/range/facing, multiple weapon families, missiles/fighters, retreat, boarding and planetary defenses in staged slices. |
| Diplomacy breadth | War, offer peace and accept peace are implemented. No treaty/alliance/non-aggression/tribute/tech-trade runtime exists. | **P2** | Important MOO2 identity, but conquest gameplay already works. Do after next playability tranche; it also enables a more meaningful Council later. |
| Espionage / Leaders | Spy/Officer Treasury categories and race spying modifiers exist, but there is no espionage, sabotage or leader runtime. | **P2** | Separate into evidence-driven slices. Avoid coupling with baseline AI until strategic AI can already finish games. |
| Colony economy/building fidelity | 48 buildings are normalized and broad Economy/Construction baselines exist, but many contextual building, pollution, morale, maintenance and empire-wide effects remain incomplete. | **P2** | Improve incrementally from concrete mismatches/regressions. Do not reopen the proven Economy core as one giant slice. |
| Population occupation/custom race depth | Conquest cohort handoff exists, but active occupation/assimilation progression plus Android/Native/custom-race Population behavior remains incomplete. | **P2** | Valuable after preset race breadth; keep as its own Population fidelity tranche. |
| Alternative victories | Conquest winner is implemented; Council, Orion/Guardian, Antarans, dimensional gate/final battle and score/time-limit families are absent. | **P2** | Do not prioritize immediately: Council benefits from multi-Empire AI+diplomacy; Orion/Antaran benefit from deeper tactical/event systems. |
| Random events | Some original event/tech-gate evidence exists but no full strategic event runtime. | **P2/P3** | Add after AI/new-game breadth so events can be exercised over long autonomous games. |
| Original-faithful galaxy breadth | Current deterministic generator is a narrow Small/Normal baseline rather than full original size/age/player distribution fidelity. | **P1/P2** | Address as part of New Game breadth, with exact evidence tests for each newly accepted setting. |
| Remote/external/MCP AI and multiplayer | Controller enum anticipates remote/external/MCP AI, and server transport exists, but no complete lobby/auth/network lifecycle is present. | **P3** | Preserve interfaces now; do not let optional transports delay local single-player. |
| Native packaging / presentation | Browser-first server contract is established; independent final art/audio/UI skin/cinematics are not complete. | **P3** | Later productization layer. Keep optional Wails/native shell thin and authority-free. |

## The next milestone

The recommended milestone is now:

> **A human can start, save/resume and finish a deterministic Human-vs-built-in-AI match from the browser using only authoritative server commands.**

That is a materially stronger product milestone than implementing another isolated original subsystem.

## Recommended numbered tranche

### Slice 13 - Built-in strategic AI baseline

Why first:

- the controller type already exists but has no runtime;
- current headless command surfaces are finally complete enough for an AI to consume rather than bypass;
- it creates a continuous autonomous regression harness for every later system;
- it unlocks single-player without requiring broader New Game settings first.

Narrow target: current two-player Small/Normal/Average/Tactical lifecycle, deterministic and cheat-free. Original-like AI fidelity remains later; first prove a legal AI can finish the supported game.

### Slice 14 - Live GameSession save / resume baseline

Why second:

- completed-session save/load is not enough for a real match;
- persistence should be stable before a browser workflow depends on it;
- a versioned live snapshot becomes a powerful AI/HMI/reconnect regression surface.

Narrow target: save/resume only at stable interactive boundaries (`planning`, encounter/battle decisions, `invasion_decisions`, `completed`), not during an in-flight resolver mutation.

### Slice 15 - Playable browser strategic HMI loop

Why third:

- the browser architecture is already accepted and transport-proven;
- AI and live persistence remove the two most expensive hidden dependencies;
- React can remain a pure client over existing authoritative APIs.

Narrow target: start a supported Human-vs-built-in-AI game; see a usable galaxy/empire state; manage research/construction/colonization/fleets/war/invasion/basic battles; end turn; save/load; see final result.

**Milestone after Slice 15:** first actually playable, saveable browser single-player match.

### Slice 16 - New Game + preset-race breadth

Why fourth:

- current runtime accepts only one very narrow setup despite 13 normalized races;
- AI can now populate extra seats and exercise newly accepted settings;
- HMI can expose the options without inventing new authority paths.

Narrow target should be frozen by Gate 1, but should expand settings/races in evidence-backed steps rather than enabling every MOO2 option at once.

### Slice 17 - Military ship designer component/weapon breadth

Why fifth:

- current design persistence/commands are architecturally ready but content validation is intentionally minimal;
- deeper Tactical Combat should consume a richer canonical design model rather than invent combat-only equipment.

Narrow target: multiple standard hulls/components and multiple legal weapon mounts with deterministic space/cost/technology validation. Defer missiles/fighters/boarding/refits if needed to keep the slice bounded.

## Backlog after Slice 17 - keep unnumbered until next audit

Suggested priority clusters, not yet slice numbers:

1. Tactical Combat depth I/II - movement/facing/range, internal systems, missiles/fighters, retreat/boarding, planetary defenses.
2. Diplomacy treaties/trade/tribute + relation history.
3. Espionage/sabotage, then Leaders/Officers.
4. Economy/building/pollution/morale/deficit fidelity driven by original-evidence fixtures.
5. Population occupation/assimilation + Android/Native/custom-race depth.
6. Full custom Race Designer and advanced government/trait interactions.
7. Galactic Council, then Orion/Guardian and Antaran/dimensional-gate victory families.
8. Random events and late-game world systems.
9. Remote/multiplayer/external-AI transports.
10. Independent presentation/content/native packaging.

## Explicit non-recommendations

### Do not make Council/Antaran Slice 13

Conquest already proves game completion. Council without multi-Empire AI and richer diplomacy would be mostly a special-case end condition; Orion/Antaran without deeper tactical/event systems would force placeholder combat/content. Both are more valuable after their dependencies are real.

### Do not reopen Economy as a monolith

The current economy is deep enough to drive a complete match and has substantial regression coverage. Remaining gaps should be evidence-driven vertical slices (specific building/pollution/morale/maintenance behavior), not an uncontrolled "finish Economy" rewrite.

### Do not put gameplay authority into the HMI

Slice 15 must only compose legal actions already owned by Go Session/App/Server surfaces. If the UI needs a rule, expose a canonical query/command; do not compute acceptance, costs, target legality or victory client-side.

### Do not attempt exact original AI before a legal baseline AI exists

The first AI objective is deterministic, observable and rules-legal. Original-like personalities/difficulty and stronger modern AI are separate fidelity/quality layers once the command-driving architecture is proven.

## Audit conclusion

Slice 12 changed the project's bottleneck. The engine no longer needs another subsystem merely to prove a match can end. The highest-leverage path is now to turn the proven lifecycle into a **saveable human-vs-AI browser game**, then widen setup/race and military-design fidelity.

Prepared next tranche: **Slices 13-17**. None are open. Slice 13 must still begin with a fresh Gate 1 repository/evidence check when explicitly started.