# Slice 15.5 interactive Tactical Combat product direction

Date: **2026-09-04**

Status: **accepted downstream product direction; dedicated Slice 15.5 inserted before final browser integration**.

## Decision

Interactive Tactical Combat is too large for visual-identity Slice 15.3 or rich decisions/persistence Slice 15.4. Slice 15 therefore expands from five to **six** sub-slices: 15.1 UX/navigation, 15.2 strategic HMI, 15.3 visual identity/assets, 15.4 rich decisions/persistence, **15.5 interactive 2D Tactical Combat**, and **15.6 final browser vertical slice/QA**. The former planned final 15.5 becomes 15.6.

## Existing proven Tactical baseline

Slice 07 already provides deterministic BattleSession-local tactical state/RNG, initiative/activation flow, fixed X/Y post-deployment positions, one standard Laser path, Armor/Structure damage/destruction, `battle.fire_beam`, `battle.end_activation`, and terminal strategic survivor/loss reconciliation. It explicitly does not yet provide player-controlled movement/turning, deployment, broad facing/shields/weapons or a general multi-ship tactical model.

## Accepted target

15.5 must turn that narrow baseline into a genuine turn-based 2D battle surface: active ship selection/indication, server-projected legal movement positions, authoritative move submission, server-projected supported weapon actions/legal targets, firing, damage/destruction visualization, activation/round progression and deterministic return to the strategic game. Desktop is mouse-first; phone/tablet is touch-first with no required hover. Exact MOO2 coordinate/grid, movement cost/range and any required facing/turning semantics remain Gate-1 research questions.

## Slice boundaries

- **15.3** may create battle backgrounds, ships, selection/target markers and effects, but owns no Tactical mechanics.
- **15.4** owns Encounter entry/return, non-tactical decisions and persistence/error UX, stopping at the battlefield boundary.
- **15.5** owns Tactical mechanics audit, authoritative movement/legal-action additions, interactive battlefield and Tactical QA.
- **15.6** integrates the accepted Tactical path into the complete browser-only Human-vs-AI journey.

## Authority rule

React only visualizes player-safe state and submits server-projected choices. Go owns tactical legality, positions, movement costs, target legality, readiness, range, RNG, damage, destruction, initiative and battle-result reconciliation.

## Breadth guard

15.5 is the first interactive Tactical baseline, not all MOO2 tactical depth. Gate 1/2 must freeze the minimum credible movement/fire/multi-ship shape and explicitly defer broad missiles/fighters/boarding/retreat/specials/planetary combat and other unbounded families.
