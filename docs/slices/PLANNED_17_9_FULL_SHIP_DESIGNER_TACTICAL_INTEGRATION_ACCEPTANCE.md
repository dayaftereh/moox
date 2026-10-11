# Planned Slice 17.9 - Full Ship Designer + Tactical Integration and Acceptance

Status: **planned; child of Slice 17; not open**.

## Objective

Integrate the completed Slice-17 design and Tactical breadth into one polished server-authoritative military-ship product flow and close the entire Slice-17 parent program with end-to-end Ship Designer, Construction, Fleet, Encounter, Tactical, persistence/replay and browser acceptance.

17.9 is an integration/acceptance slice, not a place to invent another weapon family or hide unresolved combat rules.

## Inputs

17.9 opens only after:

- 17.0 Reference/Test Harness is closed;
- 17.1 Hulls/Mandatory Systems is closed;
- 17.2 Beam design/mount metadata is closed;
- 17.3 Ordnance design breadth is closed;
- 17.2A Beam Weapon Firing Arcs is closed;
- 17.4 Specials/Fighter design breadth is closed;
- 17.5 Tactical Combat Foundation / Ship Combatant Breadth is closed;
- 17.6 Tactical Beam Weapons & Beam Modifiers is closed;
- 17.7 Tactical Missiles / Torpedoes / Bombs is closed;
- 17.8 Tactical Fighters / Bombers / Combat Specials is closed;
- every intentionally unsupported mechanic has an explicit downstream owner/reason rather than being represented as working.

## Integration scope

The final integrated product must preserve:

- hull + mandatory component + weapon + special selection;
- server-owned legality, technology, miniaturization, Space, Production and Command Point calculations;
- Weapon Selector effective-stat previews for supported Tactical effects;
- named design revisions and immutable built-Ship snapshots;
- exact Colony Construction handoff;
- persisted visual identity;
- Combat Fleet integration;
- broad Tactical admission;
- Beam, ordnance and fighter/special action selection;
- authoritative hit/RNG/damage/projectile/fighter/special results;
- event-driven Tactical presentation/animation;
- battle completion and strategic reconciliation;
- save/load/reconnect/replay/Observer compatibility;
- durable reference games and test labs.

## Gate 1 - integration audit

- [ ] Reconcile the design catalogs through 17.4, including follow-up 17.2A Beam firing arcs, with the 17.5-17.8 Tactical contracts into one coherent legal-design/combat-support model.
- [ ] Audit every supported design component for its exact Tactical support status and owner.
- [ ] Audit Selector effective-stat previews against the corresponding Tactical rules.
- [ ] Audit design revision -> construction -> built Ship -> Fleet -> BattleSession snapshot identity.
- [ ] Audit reference-lab coverage for early/mid/high-tech representative designs across Beam, ordnance, fighters and combat specials.
- [ ] Audit persistence/reconnect/replay with active Tactical state.
- [ ] Audit mobile/desktop navigation and action hierarchy across Ship Designer and Tactical.
- [ ] Inventory any remaining unsupported combinations; no silent fallback is permitted.

## Gate 2 - final integration contract freeze

- [ ] Freeze one final legal-design/support-status projection.
- [ ] Freeze end-to-end immutable snapshot identity and revision behavior.
- [ ] Freeze final Ship Designer and Tactical UI information hierarchy.
- [ ] Freeze reference-game/lab acceptance matrix.
- [ ] Freeze complete-game save/load/reconnect/replay acceptance.
- [ ] Freeze final unsupported-state/rejection behavior.
- [ ] Freeze Tactical presentation as event-driven/non-authoritative.
- [ ] Freeze final browser/mobile accessibility and touch-target criteria.

## Gate 3 - integration implementation

- [ ] Reconcile any remaining schema/API naming mismatch without duplicating combat formulas.
- [ ] Complete cross-screen navigation from Designer/Construction/Fleet/Encounter to Tactical and back.
- [ ] Ensure every supported built-Ship snapshot is consumed unchanged by Tactical.
- [ ] Ensure Selector previews and Tactical results share the same authoritative rules.
- [ ] Complete support-status/UI messaging for intentionally unsupported mechanics.
- [ ] Harden save/load/reconnect/replay during representative Tactical states.
- [ ] Finish non-authoritative presentation polish needed for readable multi-family combat.
- [ ] Keep 17.9 free of new un-audited gameplay rules.

## Gate 4 - full program QA + close

- [ ] Early-, mid- and all-tech representative designs build through normal UI/Construction.
- [ ] Frigate through Doom Star representative ships enter Tactical with correct mandatory-system values.
- [ ] Representative Beam combinations resolve hit/damage/modifier behavior correctly.
- [ ] Representative missile/torpedo scenarios resolve projectile/interception/impact correctly.
- [ ] Representative fighter/bomber/combat-special scenarios resolve correctly.
- [ ] Mixed supported fleets can select appropriate weapon/action families through Tactical UI.
- [ ] All authoritative results have readable event-driven presentation on desktop/mobile.
- [ ] Battle completion reconciles destroyed/surviving ships exactly back to strategic state.
- [ ] Save/load/reconnect/replay/reference-game workflows remain deterministic.
- [ ] Full Go tests, vet, web build, browser E2E and `git diff --check` pass.
- [ ] Exactly zero temporary Slice-17 `_OPEN_` markers remain after accepted closure.

## Exit criterion

Slice 17 closes only when a player can design a broad legal military ship, build that exact immutable revision, deploy it in a normal Combat Fleet, enter Tactical without fixture-only restrictions, use every weapon/fighter/special mechanic accepted by the 17.6-17.8 contracts, see authoritative outcomes presented correctly, finish the battle, and continue the strategic game with exact reconciliation and stable persistence.

Any intentionally unsupported mechanic must remain explicit and owned; no legal design is advertised as tactically complete merely because its metadata can be persisted.

## Conditional Tactical camera-performance follow-up (prepared 2026-10-11)

- User-reported Android Tactical 2D pan/pinch responsiveness remains a known **mobile acceptance risk**, despite SVG nodes falling from 904 to 524 in commit `5819db1`.
- `PLANNED_17_8A_TACTICAL_GPU_COMPOSITED_CAMERA.md` is a **conditional, not-open** performance follow-up. It is not automatically required for 17.9 and must not add a second `_OPEN_` marker while 17.8 is in User Review.
- At 17.9 Gate 1, explicitly check the real Android review evidence and decide: (a) performance accepted -> keep compositor work parked; (b) measurable mobile interaction blocker -> agree sequencing/ownership for a separate focused performance task before final browser/mobile acceptance.
- Gate 4 may not claim a smooth Android Tactical experience solely from desktop Chrome or SVG DOM counts; require the user's real-device acceptance or document the remaining performance limitation transparently.
- A full Canvas/WebGL renderer is **out of scope** unless a separate architecture decision and evidence justify it.
