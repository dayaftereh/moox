# Slice 15.6 Gate 2 - Full browser vertical-slice acceptance contract DRAFT

Date: 2026-09-10
Status: **DRAFT - awaiting explicit user acceptance; NOT FROZEN**

Source audit: `docs/research/SLICE_15_6_GATE1_INTEGRATION_READINESS_AUDIT_2026-09-10.md`.

## Product milestone

Close Slice 15.6 only when Master of Orion X has one saveable, resumable and finishable Human-vs-Built-in-AI browser vertical slice from Main Menu to Conquest Victory, including at least one genuinely interactive supported Tactical battle reached through ordinary strategic gameplay.

## Canonical scenario identity

- start surface: Main Menu;
- game ID: `vertical-0x8009`;
- seed: `0x8009`;
- galaxy: Small;
- age: Normal;
- technology: Average;
- Seat 1: Human / local_human;
- Seat 2: Darlok / builtin_ai;
- Strategic Combat: false;
- victory: authoritative Conquest after final Darlok Colony loss.

No imported state or direct API mutation is permitted in the milestone acceptance playthrough.

## Frozen integration bridge proposed

15.6 may expose the already-authoritative narrow Military Design baseline in the existing Ship Builder because it is required to reach the closed 15.5 Tactical feature from a normal New Game.

Allowed UI scope:

- edit/save a Frigate design;
- name;
- weapon `none` or exactly `1x Laser Cannon`;
- authoritative server-derived mandatory component/cost/space display after save/refetch;
- current procedural visual genome controls remain available;
- construction continues to use immutable design revision.

Not allowed in 15.6:

- broader hull gameplay selection;
- multiple weapons/mount layout;
- arbitrary computers/drives/armor/shields/fuel selection;
- specials;
- missiles/bombs/fighters;
- broad Ship Designer mechanics reserved for Slice 17.

## Canonical journey checkpoints

1. Main Menu -> New Game -> create `vertical-0x8009`.
2. Turn 1 Galaxy is synchronized and writable.
3. Research progresses through the pinned seed expansion chain required for 12-pc supply.
4. Human saves/revises one Laser Frigate design through the browser.
5. Human builds two armed Frigates as separate one-ship fleets.
6. Human establishes the normal seed expansion/supply chain beginning with system 10 / planet 30 and outpost hops 11/32, 12/33, 13/34, 17/43 as authoritative choices permit.
7. Human builds a Troop Transport.
8. Mid-game Save is created with visible browser control; visible Load/Resume restores the same authoritative checkpoint and play continues.
9. Contact/War is established through normal Diplomacy.
10. Normal strategic movement creates at least one Battle Entry with server-supported Tactical spec and 1-2 visible combat ships per side.
11. Tactical run includes Move, Scan, Laser Fire, activation progression and terminal Battle Return; no strategic surface is writable until return is acknowledged.
12. Strategic play continues after exact destroyed/surviving ship reconciliation.
13. Human invades/conquers all remaining Darlok Colonies using normal troop transport/Invasion UX.
14. Game reaches completed/Conquest with Human winner and Victory result surface.
15. Completed game remains identifiable from Main Menu/hosted games without re-enabling mutable game actions.

The exact turn/system of the Tactical encounter is determined by the actual deterministic Built-in-AI match. Absence of any supported encounter is a Gate-3 blocker, not permission to inject a fixture.

## No-dev-tool acceptance rule

The one run counted as Slice-15.6 milestone evidence must use visible browser interactions only.

Forbidden in that run:

- direct REST/API state mutation;
- live-snapshot import;
- editing JSON/state through console/devtools;
- test-only fixture server state;
- controlling Darlok as Local Human;
- direct route/state injection that bypasses required preceding workflow;
- hidden Tactical autoresolve in place of the required interactive supported battle.

Focused regression tests may use server/API fixtures, but are supplemental evidence only.

## Persistence/reconnect/failure acceptance

- Save from visible game UI at stable mid-game Planning state.
- Leave/reload/resume through visible browser flow.
- State identity must preserve game ID, authoritative turn/revision and meaningful strategic progress.
- Restore remains explicit confirmation.
- A stale 409 command is visible, non-destructive, refetched once and never auto-resubmitted.
- Refresh/reconnect failure preserves last known readable state and exposes Retry.
- Successful Retry/resume returns to one synchronized writable authority state.
- Tactical reconnect/reload must return to the same unresolved Battle rather than silently resuming Strategy.

## Responsive acceptance

- 1440x900 primary desktop target.
- 390px representative mobile target.
- 360px minimum functional target.
- 320px catastrophic-overflow guard.
- No required horizontal page scrolling.
- No required hover.
- Primary/secondary mobile workflow actions >=44px high.
- Galaxy and Tactical pan/zoom gestures remain usable without mutating game state.

## Runtime/performance acceptance

On the LLO Desktop same-host canonical QA baseline:

- normal route + authoritative snapshot ready <=2 seconds;
- initial canonical-route transfer <=1 MiB;
- individual JS chunk <=350 KiB unless a larger chunk is explicitly reviewed and justified;
- no permanent spinner after successful response;
- no uncaught runtime exception/unhandled Promise rejection during canonical journey;
- no missing production asset required by active surface;
- no route/navigation loop that prevents progress.

Performance limits are guardrails, not simulation authority.

## Evidence requirements

### Automated/server

- full `go test ./...`;
- full `go vet ./...`;
- `npm run build`;
- `git diff --check`;
- deterministic authority regression for canonical critical state boundaries;
- persistence/stale/Tactical/result/conquest regression coverage.

### Real browser

Capture evidence for:

- Main Menu/New Game;
- synchronized Turn-1 Galaxy;
- Research;
- narrow Laser military-design save;
- Military Ship construction;
- expansion/supply/outpost path;
- visible Save/Load/Resume;
- Diplomacy/War;
- strategic -> Tactical entry;
- Move/Scan/Fire/activation;
- Tactical -> strategic Battle Return;
- Invasion;
- Victory;
- desktop/mobile responsive checkpoints;
- no fatal runtime/asset errors.

No particular E2E framework is frozen. Add a browser-test dependency only if it materially improves repeatability; managed Chrome/manual acceptance remains valid milestone evidence when recorded precisely.

## Gate-3 implementation order after acceptance

1. narrow Military Design browser bridge + tests;
2. first canonical seed dry run against Built-in AI to locate real integration dead-ends;
3. repair only blocking cross-screen/data/navigation lifecycle seams;
4. add repeatable regressions where the dry run exposes fragile boundaries;
5. full browser Save/Resume/Tactical/Victory run;
6. responsive/performance/error polish;
7. independent Gate-4 close pass.

## Review points requiring explicit user acceptance

- seed `0x8009` / Human-vs-Darlok Built-in-AI canonical scenario;
- narrow Frigate `none | 1x Laser` browser bridge is allowed in 15.6 but broad Ship Designer stays Slice 17;
- exact Tactical encounter is produced by real AI gameplay, never a fixture, and must be 1v1-2v2 supported;
- visible browser Save/Load/Resume is mandatory in the canonical acceptance run;
- responsive/runtime/performance guardrails above;
- no-dev-tool milestone rule and framework-neutral evidence strategy.

Until those points are accepted, this document remains **DRAFT** and Gate 3 implementation does not begin.

## Gate-2 user refinement: durable Ship Designer shell

Before freeze, the Ship Designer target is refined so Gate 3 does not build a disposable one-off Laser toggle.

- **Name first:** every saved Ship Design has a player-editable name.
- **Hull size via left/right arrows:** use a stepper/carousel, not image tiles. Ruleset order is Frigate -> Destroyer -> Cruiser -> Battleship -> Titan -> Doom Star. The current `Scout` is a design name on a Frigate, not a separate hull size.
- **Technology/availability locks:** unavailable hulls/components may remain visible for progression context but are clearly locked/disabled/struck and cannot be saved or built. Availability must be projected by server authority, never inferred in React. The ruleset already contains `titan_construction` and `doom_star_construction`; complete unlock mapping still needs an authoritative projection.
- **Available list then installed list:** player-visible weapons/components appear in an available list; installed components appear separately below with stable mount/slot identity.
- **Future-ready mount rows:** the UI structure must allow later separate weapon slots, per-mount quantity +/- and weapon modifiers/upgrades without redesign. Current 15.6 authority must not falsely enable those mechanics: the present save contract still supports only none or one slot-0 Laser Cannon count 1.
- **Authoritative totals:** show Production Cost, Command Point cost/impact and used/available design space from server authority. Current engine already derives Production Cost and normal-ship Command Point usage from hull size index + 1 (1..6 CP for the six hull sizes), but React must not duplicate that gameplay formula as authority.
- **Persistent design catalog:** saving/revising creates a named design/revision visible again in the player's Ship Design list together with its visual identity/loadout.
- **Colony build handoff:** that exact named design must appear as a Military Ship construction choice on Colonies carrying `ship_design_id`, `ship_design_revision` and `ship_design_name`. Completed ships retain their source design identity/revision.

This refinement changes the **15.6 UI foundation**, not the already-frozen Tactical-15.5 combat breadth. Broader hull save support, multiple live weapon mounts/count >1, weapon modifiers and per-slot Tactical destruction remain disabled until their server/Tactical contracts are explicitly expanded.