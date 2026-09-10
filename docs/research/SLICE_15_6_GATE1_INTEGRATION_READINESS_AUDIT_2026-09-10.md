# Slice 15.6 Gate 1 - Full browser vertical-slice integration readiness audit

Date: 2026-09-10
Status: **Gate 1 complete; Gate 2 contract DRAFT awaiting user acceptance**

## Objective

Audit the closed Slice-15.1 through Slice-15.5 browser foundations as one continuous Human-vs-Built-in-AI journey and identify only the integration gaps that prevent a saveable, resumable and finishable browser game with at least one supported interactive Tactical battle.

This audit does not reopen the accepted simulation breadth of prior slices and does not pull the broad Slice-17 Ship Designer forward.

## Closed dependency baseline

### Slice 15.1 - shell / navigation / responsive foundation

Accepted evidence already provides the bilingual game shell, route ownership, desktop side rail, mobile bottom navigation, no-horizontal-overflow contract, 390px representative mobile target, 360px minimum functional target, 320px catastrophic-overflow guard, and 44px primary/secondary mobile action floor.

### Slice 15.2 - functional strategic browser play

Accepted evidence provides authoritative browser surfaces for Galaxy/System/Colony/Fleet/Research/Diplomacy and legal strategic planning/movement/colonization/outpost/construction flows.

### Slice 15.3 - visual identity

Accepted evidence provides procedural/vector ship, star, planet/body, building and icon language. Slice 15.5 already consumes the procedural ship glyph in Tactical combat.

### Slice 15.4 - rich decisions / persistence / result UX

Accepted evidence provides blocking Colony Base/Invasion/Research result surfaces, Battle entry/return, local Save/Load/Restore, hosted-game resume, stale 409 refetch without auto-resubmit, refresh-failed read-only Retry, responsive rich-state UX and Victory result presentation.

### Slice 15.5 - interactive Tactical combat

Accepted evidence provides browser-playable server-authoritative 1v1-2v2 Tactical movement/facing, Laser combat through destruction, Scan, open-field camera, Built-in-AI Tactical activation, stale/illegal command recovery and exact destroyed/survivor strategic reconciliation.

## Canonical New Game readiness

Current browser New Game already exposes:

- Game ID;
- seed, default `0x8009`;
- fixed Small galaxy;
- Normal galaxy age;
- Average technology level;
- Strategic Combat disabled;
- Seat 1 Human/local_human;
- Seat 2 Darlok/builtin_ai.

The browser immediately enters Galaxy after successful creation.

Seed `0x8009` remains the proposed canonical 15.6 seed because the existing deterministic headless conquest regression already pins the expansion/supply route and exact replay for that generated galaxy.

## Canonical seed 0x8009 authority facts

The current real player snapshot confirms:

- Human begins with Laser Cannon technology ID 100;
- initial Scout design ID 55 revision 1 is a Frigate with Nuclear Drive and no weapon;
- the existing Scout design is already a normal Colony Military Ship construction choice (25 PP in the current baseline);
- Troop Transport is already a normal 100-PP construction choice;
- construction of a Military Ship creates a separate one-ship Combat Fleet, so two newly built armed Frigates can participate as exactly two ships without merging into the initial two-Scout fleet.

The existing canonical headless victory regression further proves the ordinary seed route to 12-pc supply without state mutation:

1. research Deuterium Fuel Cells - field 9 / technology 51;
2. research required intermediate field 2 / technology 106;
3. research Iridium Fuel Cells - field 47 / technology 98;
4. research field 53 / technology 108;
5. research Urridium Fuel Cells - field 50 / technology 194;
6. move the starting Colony Ship to system 10 and colonize planet 30;
7. extend supply with Outpost Ships to system/planet 11/32, 12/33, 13/34 and 17/43;
8. build a Troop Transport;
9. declare War when contact/legality permits;
10. reach and invade Darlok holdings until the final Darlok Colony is captured;
11. authoritative conquest completion produces the Victory result.

All mechanics in that chain already have browser surfaces. The existing headless regression controls Darlok as a second local human to pin one exact reference result, so it is authority evidence rather than proof of the final required Human-vs-Built-in-AI browser journey.

## Cross-slice blocker: narrow Military Design bridge

This is the one concrete missing browser action found by Gate 1.

The server already implements authoritative command `empire.save_military_design` and its current accepted military baseline supports:

- hull exactly `frigate`;
- zero weapons, or exactly one slot-0 `laser_cannon` with count 1;
- mandatory drive/computer/armor/fuel/shield derived from known Empire technology;
- validation of Laser technology ownership;
- immutable construction revision behavior.

`ShipBuilderView` currently only saves the procedural visual genome of an existing design. There is no web submit helper or interactive control for `empire.save_military_design`.

Therefore a normal browser player cannot currently turn the known Laser technology into a buildable armed Frigate without API/dev-fixture assistance.

### Required 15.6 integration fix

Expose only the already-authoritative narrow baseline in the existing Ship Builder:

- existing/current Frigate design identity and name;
- weapon choice `none` or `1x Laser Cannon`;
- server-derived equipment/cost/space shown after authoritative refresh;
- save/revise through `empire.save_military_design`;
- preserve the existing visual-genome workflow.

Explicitly do **not** add Destroyer/Cruiser/etc. gameplay design breadth, multi-weapon layout, specials, broad component picking or other Slice-17 designer scope.

Once that one bridge exists, Colony Construction already lists the resulting revision and can build it through the normal queue.

## Built-in-AI vertical-journey evidence gap

A complete Human-vs-Built-in-AI browser conquest has not yet been demonstrated end to end.

This is an integration/evidence gap, not evidence that a new AI system is required. Baseline AI already has authoritative policies for:

- research/expansion;
- outpost construction/deployment;
- Military Ship and Troop Transport construction;
- fleet movement toward enemy Colonies;
- declaring War when an invasion departure is ready;
- Battle actions;
- Invasion;
- Colony Base decisions.

15.6 must run the canonical seed against the actual Built-in AI and treat any journey dead-end as an integration blocker. The accepted answer may be a small integration fix, but not hidden state mutation or a test-only opponent controller in the canonical acceptance run.

## Proposed exact canonical browser journey

The Gate-2 draft should freeze the following scenario identity and checkpoint rules:

1. Start at **Main Menu**, not a deep route.
2. Open **New Game**.
3. Create game `vertical-0x8009` with seed `0x8009`, Human vs Darlok Built-in AI and the fixed current Small/Normal/Average settings.
4. Enter Galaxy and verify Turn 1 / Planning / synchronized lifecycle.
5. Use the existing Research UX to follow the pinned 12-pc fuel/supply chain as choices become authoritative.
6. Use the narrow Military Design bridge to save/revise a Frigate with exactly one Laser Cannon.
7. Build **two** armed Frigates as their normal separate one-ship Combat Fleets. Keep the initial unarmed two-Scout fleet out of the chosen Tactical encounter so the Human side remains within the accepted 1-2-ship Tactical breadth.
8. Use the starting Colony Ship to establish the system-10/planet-30 supply Colony and normal Outpost construction/deployment for the pinned 11/32 -> 12/33 -> 13/34 -> 17/43 supply chain as required by authoritative movement choices.
9. Build at least one Troop Transport through normal Colony construction.
10. Exercise **Save** from the visible game menu at a stable mid-game Planning checkpoint after meaningful progress, then leave/resume/reload using visible browser persistence controls and verify the same authoritative Turn/revision/state before continuing.
11. Establish contact and declare War through the visible Diplomacy surface when legal.
12. Produce at least one **supported Tactical encounter** entirely from normal strategic fleet movement. Acceptance target is the first authoritative enemy encounter whose Battle Entry reports Tactical available and whose visible side counts are within the frozen 1v1-2v2 baseline. No fixture may be imported to make that encounter legal.
13. Enter Tactical, perform at least one Move, Scan and Laser Fire, progress activations/rounds, resolve the Battle, inspect Battle Return and continue to Strategy.
14. Continue normal strategic play and Invasions against Built-in AI holdings until the final Darlok Colony is conquered.
15. Verify authoritative `completed` phase, Conquest winner Human and the dedicated Victory result surface.
16. Return to Main Menu and verify the completed hosted game remains identifiable without reviving mutable strategic controls.

The exact Tactical turn/system is allowed to be an outcome of deterministic Built-in-AI play rather than a dev-authored fixture. If seed `0x8009` produces no supported 1-2-ship enemy encounter in the real browser journey, Gate 3 must report that as a blocker; it may not silently import or mutate one.

## Performance / responsive / runtime thresholds proposed for Gate 2

### Responsive

Carry forward accepted contracts:

- 1440x900 primary desktop target;
- 390px representative mobile target;
- 360px minimum functional target;
- 320px catastrophic-overflow guard, not preferred ergonomics;
- no required horizontal page scrolling at any accepted viewport;
- primary/secondary mobile workflow actions at least 44px high;
- no workflow may require hover.

### Runtime/error

Canonical journey must have:

- no uncaught browser runtime exception or unhandled Promise rejection;
- no missing production asset required by the active screen;
- rejected commands visible and non-destructive;
- 409 stale rejection sent once, followed by authoritative refetch, never auto-resubmitted;
- refresh failure leaves last known state inspectable and offers Retry;
- Restore remains explicit confirmation, never silent overwrite;
- Tactical/strategic transition never leaves both surfaces simultaneously writable.

### Performance

Measured `ca54326` local baseline on Chrome 152 / LLO Desktop:

- Galaxy navigation DOMContentLoaded ~399 ms;
- load event ~401 ms;
- slowest resource ~115 ms;
- initial resource transfer ~718,728 bytes;
- production `dist` ~652,664 bytes;
- main JS 247,276 bytes;
- BuildingArt JS 219,472 bytes;
- main CSS 145,449 bytes.

Recommended acceptance ceilings with substantial regression headroom:

- normal same-host route + authoritative snapshot ready within 2 seconds;
- initial canonical-route transfer <=1 MiB;
- no individual JS chunk >350 KiB without explicit evidence/review;
- no permanent loading/spinner state after a successful response;
- camera pan/zoom remains presentation-only and does not trigger game mutations.

These are draft thresholds until Gate 2 is accepted.

## Browser/E2E evidence strategy

The repository currently has no Playwright/Cypress/browser E2E dependency. Gate 1 does not add one merely to create a checkbox.

Freeze requirements, not a framework:

### Tier A - deterministic authority regression

Go/session/server tests must cover the same core state boundaries used by the browser journey, including exact replay, persistence, rejection, Tactical result reconciliation and conquest completion.

### Tier B - real browser canonical acceptance

A real managed Chrome or human browser run must execute the canonical journey using only visible user actions. For the run counted as milestone acceptance:

- no direct API mutation;
- no imported live-snapshot fixture;
- no console/state editing;
- no test-only opponent controller;
- no direct Battle construction;
- no hidden autoresolve substituted for the required supported Tactical battle.

Direct API/fixture helpers remain allowed for focused regression tests, but those runs do not count as the canonical acceptance proof.

### Tier C - responsive/runtime checkpoints

At minimum sample Main Menu/New Game, Galaxy, Colony/Construction, Research, Save/Restore, Tactical and Victory at desktop and representative phone sizes, with the 320px catastrophic-overflow guard checked on the dense high-risk surfaces.

### Save/reload/reconnect

The canonical acceptance run must exercise visible Save/Load/Resume behavior. Focused automated persistence tests may continue to use HTTP/server harnesses. Reconnect/failure QA should use an isolated QA server so the canonical 7171 review service remains available.

## Gate-1 conclusion

Slices 15.1-15.5 are integration-ready with one concrete browser blocker before the full vertical slice can be honestly claimed: **the narrow existing Frigate/optional-Laser military-design command is not exposed in the browser**.

The complete Human-vs-Built-in-AI browser victory remains the principal 15.6 evidence target. No additional simulation breadth is justified by this audit unless the real canonical run exposes a specific integration dead-end.

Gate 2 may be reviewed/frozen after user acceptance of the accompanying draft contract.
