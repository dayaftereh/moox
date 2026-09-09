# Slice 15.5 Gate 2 - Interactive Tactical contract DRAFT

Date: 2026-09-09
Status: **DRAFT - awaiting user acceptance; NOT FROZEN**

This contract translates the completed Gate-1 research into an implementation boundary. Nothing in this document is frozen until user acceptance.

## Decision 1 - battlefield authority

Go/BattleSession remains the sole Tactical authority. React renders player-safe state, presents server-projected legal actions and submits intent.

React must not independently decide:

- legal movement;
- movement cost;
- resulting facing;
- target legality;
- beam range legality;
- hit chance/outcome;
- damage/destruction;
- initiative/round progression;
- battle result.

## Decision 2 - first supported battle breadth

Proposed Gate-2 freeze:

- support **1-vs-1 through 2-vs-2 combat ships** for the first interactive Tactical slice;
- make **2-vs-2** the primary acceptance fixture because it matches normal starting scout fleets;
- keep state/protocol list-shaped rather than encoding a hard two-ship ceiling into the wire format;
- reject larger/unsupported battle shapes explicitly until a later breadth slice.

This is deliberately broader than the Slice-07 exact 1-vs-1 fixture but still bounded.

## Decision 3 - Tactical coordinates and facing

- Tactical positions remain integer X/Y.
- Facing is authoritative integer `0..15`.
- Sixteen sectors correspond to nearest ~22.5-degree heading.
- A move to a different coordinate turns the ship to the server-derived destination facing as part of the committed move.
- Same-coordinate move is not used as an implicit rotate-only command in this first slice.

Broad firing arcs and shield-facing remain deferred.

## Decision 4 - movement budget and exact movement cost

Each active ship has current and maximum movement points. Round initialization derives the movement allowance from authoritative current combat speed using the accepted original/rules baseline.

Destination cost is frozen as:

`translation = ceil(sqrt(dx^2 + dy^2))`

`facingSteps = shortestCircularDifference(currentFacing, destinationFacing, 16)`

`moveCost = translation + facingSteps * turnCostPerStep`

Turn cost per facing step:

- normal: 2;
- Inertial Stabilizer: 1;
- Inertial Nullifier: 0.

A destination is legal only if all authoritative spatial/collision/bounds rules pass and `moveCost <= movementRemaining`.

The browser never recomputes this to decide legality.

## Decision 5 - legal movement projection

For the currently active player-controlled ship, the player-safe Battle view projects:

`legal_moves[]`

Each entry contains at minimum:

- `x`;
- `y`;
- `move_cost`;
- `resulting_facing`;
- `movement_remaining_after`.

Only these destinations are interactive in the browser.

Gate-3 implementation may optimize representation if the legal set is large, but any optimization must preserve server authority and deterministic equivalence.

## Decision 6 - move command

Add command type:

`battle.move_ship`

Payload:

- `ship_id`;
- destination `x`;
- destination `y`.

Preconditions include:

- Battle active;
- correct active seat/ship;
- expected battle/global sequence;
- destination currently legal;
- sufficient movement budget;
- ship alive/eligible.

On success the BattleSession atomically updates X/Y, facing and remaining movement points, emits deterministic battle-local event(s), increments sequence and reprojects legal actions.

On failure there is **zero Tactical mutation** and no RNG consumption.

## Decision 7 - initiative and activation

Preserve the accepted Initiative option/order baseline.

For the first 2-vs-2 breadth:

- one ship is active at a time;
- active ship may perform zero or more legal moves and the supported fire action while its state permits;
- `battle.end_activation` removes/finishes that ship for the current round, analogous to the original Done semantics;
- next eligible ship becomes active according to authoritative initiative/order;
- after all eligible ships complete, a new round initializes movement/readiness and rebuilds initiative/active state;
- Wait/reorder behavior is not required in the first implementation unless needed during Gate-3 usability review.

## Decision 8 - supported fire path

Retain the existing standard Laser `battle.fire_beam` path as the first weapon interaction.

Player-safe projection supplies:

- supported ready weapon/action;
- legal target ship IDs;
- authoritative range information needed for explanation/UI.

The browser selects a weapon/action and a projected target, then submits the existing authoritative fire command. Hit roll, range modifier, damage, RNG, destruction and readiness remain server-owned.

Movement distance and beam range remain separate formulas.

## Decision 9 - Tactical player-safe state

Proposed minimum Battle projection:

- round number;
- battle-local/global sequence;
- active seat ID;
- active ship ID;
- player-safe initiative/order state;
- Tactical ships with:
  - ship/empire/seat IDs;
  - X/Y;
  - facing;
  - movement current/max;
  - Armor current/max;
  - Structure current/max;
  - destroyed/eligible state;
  - supported weapon readiness;
- active-player `legal_moves`;
- legal fire/action catalog and target IDs;
- `can_end_activation`.

Hidden opponent information must not be added merely for convenience; use only participant-safe information justified by the current Tactical model.

## Decision 10 - battlefield interaction / desktop

- Battle route remains `#/game/<gameID>/battle/<battleID>`.
- `Taktischen Kampf betreten` opens the battlefield inside the existing Battle shell.
- Battlefield is the primary area.
- Active ship is unmistakably highlighted.
- Legal moves are visible overlays; click submits movement intent.
- Ships are rendered from accepted MOOX procedural visual primitives where available and rotated to authoritative facing.
- Action/status panel shows movement, facing, Armor/Structure and Laser readiness.
- Select Laser/action -> show projected target affordances -> click target -> submit fire.
- End Activation is a persistent primary action.
- Hover can add information but is never required.

## Decision 11 - battlefield interaction / mobile

- battlefield gets most of the viewport;
- one-finger pan and pinch zoom where needed;
- tap legal destination/ship/target;
- compact bottom sheet/overlay for active-ship status and actions;
- primary controls maintain the 44-px touch target baseline;
- no required hover;
- phone interaction submits exactly the same commands as desktop.

## Decision 12 - camera/pan/zoom

Camera state is client presentation state only and must never affect Tactical authority.

- desktop: wheel/pinch-equivalent zoom plus drag/pan;
- touch: pinch zoom and one-finger pan when not choosing an explicit legal destination;
- fit-to-battle/active-ship recenter is allowed;
- no camera coordinate is sent as a game-rule input.

## Decision 13 - stale/reconnect semantics

Reuse the frozen Slice-15.4 lifecycle contract:

- invalidation/refetch never auto-resubmits a Tactical command;
- stale/illegal Tactical rejection remains visible;
- during reconnect/refresh failure, last known Tactical state stays inspectable but read-only;
- explicit Retry refetches authority;
- reconnect/import/restore mid-Battle must restore exact BattleSession state and sequence before controls unlock.

No local optimistic mutation of authoritative ship position/damage is permitted.

## Decision 14 - Battle completion and return

Terminal Tactical result continues through the existing server/session reconciliation and frozen Slice-15.4 Battle Return surface.

Exact strategic ship identities are authoritative:

- destroyed Tactical ship IDs are removed/reconciled by Go;
- visible survivors remain concrete strategic ship IDs;
- losing survivors/retreat/blockade handling stays in strategic continuation;
- React never edits strategic fleets directly.

## Decision 15 - explicit deferrals

Not part of this Gate-2 draft unless separately pulled in before freeze:

- deployment editor;
- more than 2 combat ships per side as an accepted browser shape;
- rotate-only command;
- shield facing/arcs;
- broad beam/missile/torpedo/bomb/fighter/PD systems;
- boarding/capture;
- retreat command;
- stasis/self-destruct;
- repair/regeneration/special-system breadth;
- planet/station combat;
- monsters/Antarans;
- timing-sensitive cinematics/animation as authority.

## Decision 16 - required Gate-3/4 acceptance scenarios

At minimum:

1. normal strategic browser play enters a supported 2-vs-2 Tactical encounter;
2. battlefield shows authoritative active ship and projected legal moves;
3. desktop click and mobile tap can submit a legal move;
4. move updates X/Y, facing and remaining points authoritatively;
5. illegal/over-budget/stale move causes zero mutation and clear refetch/error state;
6. supported Laser exposes only legal projected targets and can be fired from the battlefield;
7. authoritative damage/destruction/readiness refreshes visibly;
8. End Activation advances through multiple ships and at least one full round boundary;
9. round reset restores authoritative movement/readiness correctly;
10. reconnect mid-Tactical restores exact state/sequence without duplicate command;
11. Battle resolves and destroyed/surviving strategic identities match the Slice-15.4 return summary;
12. representative desktop and 320px phone/touch QA has no catastrophic overflow and no required hover;
13. full Go tests/vet/web build/diff checks pass.

## Gate-2 review points requiring user acceptance

The key product choices to approve before freeze are:

1. **2-vs-2 is the minimum credible first supported Tactical breadth** rather than staying 1-vs-1;
2. **authentic 16-way facing and original turn-cost movement** are included now;
3. the first UI uses a modern pan/zoom battlefield with server-projected movement overlays instead of a literal original-pixel combat screen;
4. broad shields/arcs/missiles/etc. remain deferred;
5. mobile uses the same game semantics with touch/bottom-sheet ergonomics.

Until these are accepted, this file remains DRAFT and Gate 3 does not begin.
